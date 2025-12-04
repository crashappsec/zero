#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Supply Chain Report Type
# Dependencies, provenance, health assessment
#############################################################################

# Generate supply-chain report data as JSON
generate_report_data() {
    local project_id="$1"
    local analysis_path="$2"

    # Load manifest
    local manifest=$(load_manifest "$analysis_path")
    local scan_id=$(echo "$manifest" | jq -r '.scan_id // "unknown"')
    local profile=$(echo "$manifest" | jq -r '.scan.profile // "standard"')
    local completed_at=$(echo "$manifest" | jq -r '.scan.completed_at // ""')

    # SBOM data
    local sbom_data='{}'
    local packages='[]'
    if has_scanner_data "$analysis_path" "package-sbom"; then
        sbom_data=$(load_scanner_data "$analysis_path" "package-sbom")
        packages=$(echo "$sbom_data" | jq '.packages // .dependencies // []')
    fi
    local total_deps=$(echo "$sbom_data" | jq -r '.total_dependencies // .summary.total // 0')
    local direct_deps=$(echo "$sbom_data" | jq -r '.direct_dependencies // .summary.direct // 0')
    local ecosystems=$(echo "$sbom_data" | jq '.summary.ecosystems // {}')

    # Package health
    local health_data='{}'
    local health_issues='[]'
    if has_scanner_data "$analysis_path" "package-health"; then
        health_data=$(load_scanner_data "$analysis_path" "package-health")
        health_issues=$(echo "$health_data" | jq '.issues // []')
    fi
    local abandoned=$(echo "$health_data" | jq -r '.summary.abandoned // 0')
    local deprecated=$(echo "$health_data" | jq -r '.summary.deprecated // 0')
    local outdated=$(echo "$health_data" | jq -r '.summary.outdated // 0')
    local typosquatting=$(echo "$health_data" | jq -r '.summary.typosquatting_suspects // 0')

    # Vulnerability data
    local vulns=$(aggregate_vulns "$analysis_path")
    local vuln_critical=$(echo "$vulns" | jq -r '.critical')
    local vuln_high=$(echo "$vulns" | jq -r '.high')
    local vuln_total=$(echo "$vulns" | jq -r '.total')

    # Provenance data
    local prov_data='{}'
    if has_scanner_data "$analysis_path" "package-provenance"; then
        prov_data=$(load_scanner_data "$analysis_path" "package-provenance")
    fi
    local signed_packages=$(echo "$prov_data" | jq -r '.summary.signed_packages // 0')
    local slsa_level=$(echo "$prov_data" | jq -r '.summary.slsa_level // "none"')
    local attestations=$(echo "$prov_data" | jq -r '.summary.attestations // 0')

    # Calculate supply chain health score
    local supply_chain_score=100
    supply_chain_score=$((supply_chain_score - abandoned * 5))
    supply_chain_score=$((supply_chain_score - deprecated * 3))
    supply_chain_score=$((supply_chain_score - typosquatting * 20))
    supply_chain_score=$((supply_chain_score - vuln_critical * 15))
    supply_chain_score=$((supply_chain_score - vuln_high * 8))
    [[ $supply_chain_score -lt 0 ]] && supply_chain_score=0

    # Risk level
    local risk_level="low"
    [[ $typosquatting -gt 0 ]] && risk_level="critical"
    [[ $vuln_critical -gt 0 ]] && risk_level="critical"
    [[ $vuln_high -gt 0 ]] && risk_level="high"
    [[ $abandoned -gt 3 ]] && risk_level="medium"

    # Build JSON
    jq -n \
        --arg project_id "$project_id" \
        --arg scan_id "$scan_id" \
        --arg profile "$profile" \
        --arg completed_at "$completed_at" \
        --arg risk_level "$risk_level" \
        --argjson supply_chain_score "$supply_chain_score" \
        --argjson total_deps "$total_deps" \
        --argjson direct_deps "$direct_deps" \
        --argjson ecosystems "$ecosystems" \
        --argjson packages "$packages" \
        --argjson abandoned "$abandoned" \
        --argjson deprecated "$deprecated" \
        --argjson outdated "$outdated" \
        --argjson typosquatting "$typosquatting" \
        --argjson health_issues "$health_issues" \
        --argjson vuln_critical "$vuln_critical" \
        --argjson vuln_high "$vuln_high" \
        --argjson vuln_total "$vuln_total" \
        --argjson signed_packages "$signed_packages" \
        --arg slsa_level "$slsa_level" \
        --argjson attestations "$attestations" \
        '{
            report_type: "supply-chain",
            report_version: "1.0.0",
            generated_at: (now | todate),
            project: {
                id: $project_id,
                scan_id: $scan_id,
                profile: $profile,
                completed_at: $completed_at
            },
            supply_chain_score: $supply_chain_score,
            risk: {
                level: $risk_level
            },
            dependencies: {
                total: $total_deps,
                direct: $direct_deps,
                transitive: ($total_deps - $direct_deps),
                ecosystems: $ecosystems,
                packages: ($packages | if length > 100 then .[0:100] else . end)
            },
            health: {
                abandoned: $abandoned,
                deprecated: $deprecated,
                outdated: $outdated,
                typosquatting_suspects: $typosquatting,
                issues: ($health_issues | if length > 20 then .[0:20] else . end)
            },
            vulnerabilities: {
                critical: $vuln_critical,
                high: $vuln_high,
                total: $vuln_total
            },
            provenance: {
                signed_packages: $signed_packages,
                slsa_level: $slsa_level,
                attestations: $attestations,
                signed_percentage: (if $total_deps > 0 then (($signed_packages / $total_deps) * 100 | floor) else 0 end)
            }
        }'
}

# Generate org aggregate supply-chain data
generate_org_report_data() {
    local org="$1"
    local projects="$2"

    local project_count=$(echo "$projects" | wc -w | tr -d ' ')
    local total_deps=0
    local total_abandoned=0
    local total_typosquatting=0
    local repos_at_risk=()

    for repo in $projects; do
        local project_id="$org/$repo"
        local analysis_path="$GIBSON_PROJECTS_DIR/$project_id/analysis"

        if [[ -d "$analysis_path" ]]; then
            local deps=$(aggregate_deps "$analysis_path")
            local d=$(echo "$deps" | jq -r '.total')
            total_deps=$((total_deps + d))

            if has_scanner_data "$analysis_path" "package-health"; then
                local health=$(load_scanner_data "$analysis_path" "package-health")
                local a=$(echo "$health" | jq -r '.summary.abandoned // 0')
                local t=$(echo "$health" | jq -r '.summary.typosquatting_suspects // 0')
                total_abandoned=$((total_abandoned + a))
                total_typosquatting=$((total_typosquatting + t))
                [[ $t -gt 0 ]] && repos_at_risk+=("$repo")
            fi
        fi
    done

    local risk_level="low"
    [[ $total_typosquatting -gt 0 ]] && risk_level="critical"
    [[ $total_abandoned -gt 10 ]] && risk_level="medium"

    jq -n \
        --arg org "$org" \
        --argjson project_count "$project_count" \
        --arg risk_level "$risk_level" \
        --argjson total_deps "$total_deps" \
        --argjson total_abandoned "$total_abandoned" \
        --argjson total_typosquatting "$total_typosquatting" \
        --arg repos_at_risk "$(printf '%s\n' "${repos_at_risk[@]}" | paste -sd, -)" \
        '{
            report_type: "supply-chain",
            report_version: "1.0.0",
            generated_at: (now | todate),
            organization: $org,
            projects: {
                count: $project_count,
                at_risk: ($repos_at_risk | split(",") | map(select(length > 0)))
            },
            risk: {
                level: $risk_level
            },
            dependencies: {
                total: $total_deps
            },
            health: {
                abandoned: $total_abandoned,
                typosquatting_suspects: $total_typosquatting
            }
        }'
}

export -f generate_report_data
export -f generate_org_report_data
