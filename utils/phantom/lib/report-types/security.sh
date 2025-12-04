#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Security Report Type
# Deep dive into security posture - vulnerabilities, secrets, code security
#############################################################################

# Generate security report data as JSON
generate_report_data() {
    local project_id="$1"
    local analysis_path="$2"

    # Load manifest
    local manifest=$(load_manifest "$analysis_path")
    local scan_id=$(echo "$manifest" | jq -r '.scan_id // "unknown"')
    local profile=$(echo "$manifest" | jq -r '.scan.profile // "standard"')

    # Get completed_at - prefer scan.completed_at, fallback to most recent analyzer timestamp
    local completed_at=$(echo "$manifest" | jq -r '
        if .scan.completed_at != null then .scan.completed_at
        else [.analyses[].completed_at | select(. != null)] | sort | last // ""
        end
    ')

    # Vulnerability details
    local vulns_data='{}'
    local vuln_list='[]'
    if has_scanner_data "$analysis_path" "package-vulns"; then
        vulns_data=$(load_scanner_data "$analysis_path" "package-vulns")
        vuln_list=$(echo "$vulns_data" | jq '.vulnerabilities // []')
    fi
    local critical=$(echo "$vulns_data" | jq -r '.summary.critical // 0')
    local high=$(echo "$vulns_data" | jq -r '.summary.high // 0')
    local medium=$(echo "$vulns_data" | jq -r '.summary.medium // 0')
    local low=$(echo "$vulns_data" | jq -r '.summary.low // 0')
    local total_vulns=$(echo "$vulns_data" | jq -r '.summary.total // 0')

    # Secrets details
    local secrets_data='{}'
    local secrets_list='[]'
    if has_scanner_data "$analysis_path" "code-secrets"; then
        secrets_data=$(load_scanner_data "$analysis_path" "code-secrets")
        secrets_list=$(echo "$secrets_data" | jq '.findings // []')
    fi
    local secrets_count=$(echo "$secrets_data" | jq -r '.summary.total_findings // 0')
    local secrets_by_type=$(echo "$secrets_data" | jq '.summary.by_type // {}')

    # Code security findings
    local code_security_data='{}'
    local code_findings='[]'
    if has_scanner_data "$analysis_path" "code-security"; then
        code_security_data=$(load_scanner_data "$analysis_path" "code-security")
        code_findings=$(echo "$code_security_data" | jq '.findings // []')
    fi
    local code_security_count=$(echo "$code_security_data" | jq -r '.summary.total // 0')
    local code_security_critical=$(echo "$code_security_data" | jq -r '.summary.critical // 0')
    local code_security_high=$(echo "$code_security_data" | jq -r '.summary.high // 0')

    # IaC security findings
    local iac_data='{}'
    local iac_findings='[]'
    if has_scanner_data "$analysis_path" "iac-security"; then
        iac_data=$(load_scanner_data "$analysis_path" "iac-security")
        iac_findings=$(echo "$iac_data" | jq '.findings // []')
    fi
    local iac_count=$(echo "$iac_data" | jq -r '.summary.total // 0')
    local iac_critical=$(echo "$iac_data" | jq -r '.summary.critical // 0')

    # Calculate overall security score (0-100)
    local security_score=100
    security_score=$((security_score - critical * 20))
    security_score=$((security_score - high * 10))
    security_score=$((security_score - medium * 3))
    security_score=$((security_score - secrets_count * 15))
    security_score=$((security_score - code_security_critical * 15))
    security_score=$((security_score - code_security_high * 8))
    [[ $security_score -lt 0 ]] && security_score=0

    # Risk level
    local risk_level=$(calculate_risk_level "$critical" "$high" "$medium")

    # Build JSON
    jq -n \
        --arg project_id "$project_id" \
        --arg scan_id "$scan_id" \
        --arg profile "$profile" \
        --arg completed_at "$completed_at" \
        --arg risk_level "$risk_level" \
        --argjson security_score "$security_score" \
        --argjson critical "$critical" \
        --argjson high "$high" \
        --argjson medium "$medium" \
        --argjson low "$low" \
        --argjson total_vulns "$total_vulns" \
        --argjson vuln_list "$vuln_list" \
        --argjson secrets_count "$secrets_count" \
        --argjson secrets_by_type "$secrets_by_type" \
        --argjson secrets_list "$secrets_list" \
        --argjson code_security_count "$code_security_count" \
        --argjson code_security_critical "$code_security_critical" \
        --argjson code_security_high "$code_security_high" \
        --argjson code_findings "$code_findings" \
        --argjson iac_count "$iac_count" \
        --argjson iac_critical "$iac_critical" \
        --argjson iac_findings "$iac_findings" \
        '{
            report_type: "security",
            report_version: "1.0.0",
            generated_at: (now | todate),
            project: {
                id: $project_id,
                scan_id: $scan_id,
                profile: $profile,
                completed_at: $completed_at
            },
            security_score: $security_score,
            risk: {
                level: $risk_level
            },
            vulnerabilities: {
                summary: {
                    critical: $critical,
                    high: $high,
                    medium: $medium,
                    low: $low,
                    total: $total_vulns
                },
                details: ($vuln_list | if length > 20 then .[0:20] else . end)
            },
            secrets: {
                summary: {
                    total: $secrets_count,
                    by_type: $secrets_by_type
                },
                details: ($secrets_list | if length > 10 then .[0:10] else . end)
            },
            code_security: {
                summary: {
                    total: $code_security_count,
                    critical: $code_security_critical,
                    high: $code_security_high
                },
                details: ($code_findings | if length > 10 then .[0:10] else . end)
            },
            iac_security: {
                summary: {
                    total: $iac_count,
                    critical: $iac_critical
                },
                details: ($iac_findings | if length > 10 then .[0:10] else . end)
            }
        }'
}

# Generate org aggregate security data
generate_org_report_data() {
    local org="$1"
    local projects="$2"

    local project_count=$(echo "$projects" | wc -w | tr -d ' ')
    local total_vulns=0
    local total_critical=0
    local total_high=0
    local total_secrets=0
    local repos_with_critical=()

    for repo in $projects; do
        local project_id="$org/$repo"
        local analysis_path="$GIBSON_PROJECTS_DIR/$project_id/analysis"

        if [[ -d "$analysis_path" ]]; then
            local vulns=$(aggregate_vulns "$analysis_path")
            local c=$(echo "$vulns" | jq -r '.critical')
            local h=$(echo "$vulns" | jq -r '.high')
            local t=$(echo "$vulns" | jq -r '.total')

            total_critical=$((total_critical + c))
            total_high=$((total_high + h))
            total_vulns=$((total_vulns + t))

            if has_scanner_data "$analysis_path" "code-secrets"; then
                local s=$(load_scanner_data "$analysis_path" "code-secrets" | jq -r '.summary.total_findings // 0')
                total_secrets=$((total_secrets + s))
            fi

            if [[ $c -gt 0 ]]; then
                repos_with_critical+=("$repo")
            fi
        fi
    done

    local risk_level=$(calculate_risk_level "$total_critical" "$total_high" "0")

    jq -n \
        --arg org "$org" \
        --argjson project_count "$project_count" \
        --arg risk_level "$risk_level" \
        --argjson total_vulns "$total_vulns" \
        --argjson total_critical "$total_critical" \
        --argjson total_high "$total_high" \
        --argjson total_secrets "$total_secrets" \
        --arg repos_critical "$(printf '%s\n' "${repos_with_critical[@]}" | paste -sd, -)" \
        '{
            report_type: "security",
            report_version: "1.0.0",
            generated_at: (now | todate),
            organization: $org,
            projects: {
                count: $project_count,
                with_critical: ($repos_critical | split(",") | map(select(length > 0)))
            },
            risk: {
                level: $risk_level
            },
            vulnerabilities: {
                total: $total_vulns,
                critical: $total_critical,
                high: $total_high
            },
            secrets: {
                total: $total_secrets
            }
        }'
}

export -f generate_report_data
export -f generate_org_report_data
