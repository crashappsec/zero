#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Malcontent Report Type
# Supply chain compromise detection analysis aligned with Scout agent
#
# This report presents findings from Chainguard's malcontent tool,
# focusing on suspicious behaviors detected in repository code.
#############################################################################

# Generate malcontent report data as JSON
generate_report_data() {
    local project_id="$1"
    local analysis_path="$2"

    # Load manifest
    local manifest=$(load_manifest "$analysis_path")
    local scan_id=$(echo "$manifest" | jq -r '.scan_id // "unknown"')
    local profile=$(echo "$manifest" | jq -r '.scan.profile // "standard"')
    local completed_at=$(echo "$manifest" | jq -r '.scan.completed_at // ""')

    # Load malcontent data
    local malcontent_data='{}'
    local findings='[]'
    local summary='{}'

    if has_scanner_data "$analysis_path" "package-malcontent"; then
        malcontent_data=$(load_scanner_data "$analysis_path" "package-malcontent")
        findings=$(echo "$malcontent_data" | jq '.findings // []')
        summary=$(echo "$malcontent_data" | jq '.summary // {}')
    fi

    # Extract summary metrics
    local total_files=$(echo "$summary" | jq -r '.total_files // 0')
    local total_rules=$(echo "$summary" | jq -r '.total_rules_matched // 0')
    local critical=$(echo "$summary" | jq -r '.by_risk.critical // 0')
    local high=$(echo "$summary" | jq -r '.by_risk.high // 0')
    local medium=$(echo "$summary" | jq -r '.by_risk.medium // 0')
    local low=$(echo "$summary" | jq -r '.by_risk.low // 0')

    # SBOM data for package context
    local sbom_data='{}'
    local total_packages=0
    if has_scanner_data "$analysis_path" "package-sbom"; then
        sbom_data=$(load_scanner_data "$analysis_path" "package-sbom")
        total_packages=$(echo "$sbom_data" | jq -r '.total_dependencies // .summary.total // 0')
    fi

    # Calculate risk level
    local risk_level="low"
    [[ $critical -gt 0 ]] && risk_level="critical"
    [[ $high -gt 0 ]] && [[ "$risk_level" != "critical" ]] && risk_level="high"
    [[ $medium -gt 5 ]] && [[ "$risk_level" == "low" ]] && risk_level="medium"

    # Group findings by severity for easier display
    local critical_findings=$(echo "$findings" | jq '[.[] | select(.risk == "critical")]')
    local high_findings=$(echo "$findings" | jq '[.[] | select(.risk == "high")]')
    local medium_findings=$(echo "$findings" | jq '[.[] | select(.risk == "medium")] | .[0:20]')

    # Build JSON
    jq -n \
        --arg project_id "$project_id" \
        --arg scan_id "$scan_id" \
        --arg profile "$profile" \
        --arg completed_at "$completed_at" \
        --arg risk_level "$risk_level" \
        --argjson total_files "$total_files" \
        --argjson total_rules "$total_rules" \
        --argjson total_packages "$total_packages" \
        --argjson critical "$critical" \
        --argjson high "$high" \
        --argjson medium "$medium" \
        --argjson low "$low" \
        --argjson critical_findings "$critical_findings" \
        --argjson high_findings "$high_findings" \
        --argjson medium_findings "$medium_findings" \
        --argjson all_findings "$findings" \
        '{
            report_type: "malcontent",
            report_version: "1.0.0",
            generated_at: (now | todate),
            project: {
                id: $project_id,
                scan_id: $scan_id,
                profile: $profile,
                completed_at: $completed_at
            },
            risk: {
                level: $risk_level
            },
            summary: {
                total_files_scanned: $total_files,
                total_rules_matched: $total_rules,
                total_packages: $total_packages,
                by_severity: {
                    critical: $critical,
                    high: $high,
                    medium: $medium,
                    low: $low
                }
            },
            findings: {
                critical: $critical_findings,
                high: $high_findings,
                medium: $medium_findings
            },
            all_findings: ($all_findings | if length > 50 then .[0:50] else . end)
        }'
}

# Generate org aggregate malcontent data
generate_org_report_data() {
    local org="$1"
    local projects="$2"

    local project_count=$(echo "$projects" | wc -w | tr -d ' ')
    local total_critical=0
    local total_high=0
    local total_medium=0
    local total_files=0
    local repos_with_issues=()

    for repo in $projects; do
        local project_id="$org/$repo"
        local analysis_path="$GIBSON_PROJECTS_DIR/$project_id/analysis"

        if [[ -d "$analysis_path" ]] && has_scanner_data "$analysis_path" "package-malcontent"; then
            local data=$(load_scanner_data "$analysis_path" "package-malcontent")
            local c=$(echo "$data" | jq -r '.summary.by_risk.critical // 0')
            local h=$(echo "$data" | jq -r '.summary.by_risk.high // 0')
            local m=$(echo "$data" | jq -r '.summary.by_risk.medium // 0')
            local f=$(echo "$data" | jq -r '.summary.total_files // 0')

            total_critical=$((total_critical + c))
            total_high=$((total_high + h))
            total_medium=$((total_medium + m))
            total_files=$((total_files + f))

            [[ $c -gt 0 || $h -gt 0 ]] && repos_with_issues+=("$repo")
        fi
    done

    local risk_level="low"
    [[ $total_critical -gt 0 ]] && risk_level="critical"
    [[ $total_high -gt 0 ]] && [[ "$risk_level" != "critical" ]] && risk_level="high"

    jq -n \
        --arg org "$org" \
        --argjson project_count "$project_count" \
        --arg risk_level "$risk_level" \
        --argjson total_files "$total_files" \
        --argjson total_critical "$total_critical" \
        --argjson total_high "$total_high" \
        --argjson total_medium "$total_medium" \
        --arg repos_with_issues "$(printf '%s\n' "${repos_with_issues[@]}" | paste -sd, -)" \
        '{
            report_type: "malcontent",
            report_version: "1.0.0",
            generated_at: (now | todate),
            organization: $org,
            projects: {
                count: $project_count,
                with_issues: ($repos_with_issues | split(",") | map(select(length > 0)))
            },
            risk: {
                level: $risk_level
            },
            summary: {
                total_files_scanned: $total_files,
                by_severity: {
                    critical: $total_critical,
                    high: $total_high,
                    medium: $total_medium
                }
            }
        }'
}

export -f generate_report_data
export -f generate_org_report_data
