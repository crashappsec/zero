#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Compliance Report Type
# License and policy compliance status
#############################################################################

# Generate compliance report data as JSON
generate_report_data() {
    local project_id="$1"
    local analysis_path="$2"

    # Load manifest
    local manifest=$(load_manifest "$analysis_path")
    local scan_id=$(echo "$manifest" | jq -r '.scan_id // "unknown"')
    local profile=$(echo "$manifest" | jq -r '.scan.profile // "standard"')
    local completed_at=$(echo "$manifest" | jq -r '.scan.completed_at // ""')

    # License data
    local license_data='{}'
    local license_list='[]'
    if has_scanner_data "$analysis_path" "licenses"; then
        license_data=$(load_scanner_data "$analysis_path" "licenses")
        license_list=$(echo "$license_data" | jq '.licenses // []')
    fi
    local license_status=$(echo "$license_data" | jq -r '.summary.overall_status // "unknown"')
    local license_violations=$(echo "$license_data" | jq -r '.summary.license_violations // 0')
    local copyleft_count=$(echo "$license_data" | jq -r '.summary.copyleft_count // 0')
    local unknown_licenses=$(echo "$license_data" | jq -r '.summary.unknown_count // 0')
    local license_types=$(echo "$license_data" | jq '.summary.license_types // {}')

    # SBOM completeness
    local sbom_data='{}'
    if has_scanner_data "$analysis_path" "package-sbom"; then
        sbom_data=$(load_scanner_data "$analysis_path" "package-sbom")
    fi
    local total_deps=$(echo "$sbom_data" | jq -r '.total_dependencies // .summary.total // 0')
    local direct_deps=$(echo "$sbom_data" | jq -r '.direct_dependencies // .summary.direct // 0')

    # Documentation coverage
    local doc_data='{}'
    if has_scanner_data "$analysis_path" "documentation"; then
        doc_data=$(load_scanner_data "$analysis_path" "documentation")
    fi
    local has_readme=$(echo "$doc_data" | jq -r '.summary.has_readme // false')
    local has_license_file=$(echo "$doc_data" | jq -r '.summary.has_license // false')
    local has_contributing=$(echo "$doc_data" | jq -r '.summary.has_contributing // false')
    local doc_score=$(echo "$doc_data" | jq -r '.summary.score // 0')

    # Code ownership
    local ownership_data='{}'
    if has_scanner_data "$analysis_path" "code-ownership"; then
        ownership_data=$(load_scanner_data "$analysis_path" "code-ownership")
    fi
    local has_codeowners=$(echo "$ownership_data" | jq -r '.summary.has_codeowners // false')
    local coverage_pct=$(echo "$ownership_data" | jq -r '.summary.coverage_percentage // 0')

    # Calculate compliance score
    local compliance_score=100
    [[ "$license_status" != "pass" ]] && compliance_score=$((compliance_score - 20))
    [[ $license_violations -gt 0 ]] && compliance_score=$((compliance_score - license_violations * 10))
    [[ $copyleft_count -gt 0 ]] && compliance_score=$((compliance_score - 5))
    [[ $unknown_licenses -gt 0 ]] && compliance_score=$((compliance_score - unknown_licenses * 3))
    [[ "$has_readme" != "true" ]] && compliance_score=$((compliance_score - 10))
    [[ "$has_license_file" != "true" ]] && compliance_score=$((compliance_score - 15))
    [[ $compliance_score -lt 0 ]] && compliance_score=0

    # Overall status
    local overall_status="PASS"
    [[ $license_violations -gt 0 ]] && overall_status="FAIL"
    [[ $compliance_score -lt 70 ]] && overall_status="WARN"

    # Build JSON
    jq -n \
        --arg project_id "$project_id" \
        --arg scan_id "$scan_id" \
        --arg profile "$profile" \
        --arg completed_at "$completed_at" \
        --arg overall_status "$overall_status" \
        --argjson compliance_score "$compliance_score" \
        --arg license_status "$license_status" \
        --argjson license_violations "$license_violations" \
        --argjson copyleft_count "$copyleft_count" \
        --argjson unknown_licenses "$unknown_licenses" \
        --argjson license_types "$license_types" \
        --argjson license_list "$license_list" \
        --argjson total_deps "$total_deps" \
        --argjson direct_deps "$direct_deps" \
        --arg has_readme "$has_readme" \
        --arg has_license_file "$has_license_file" \
        --arg has_contributing "$has_contributing" \
        --argjson doc_score "$doc_score" \
        --arg has_codeowners "$has_codeowners" \
        --argjson coverage_pct "$coverage_pct" \
        '{
            report_type: "compliance",
            report_version: "1.0.0",
            generated_at: (now | todate),
            project: {
                id: $project_id,
                scan_id: $scan_id,
                profile: $profile,
                completed_at: $completed_at
            },
            overall_status: $overall_status,
            compliance_score: $compliance_score,
            licenses: {
                status: $license_status,
                violations: $license_violations,
                copyleft_packages: $copyleft_count,
                unknown_licenses: $unknown_licenses,
                license_types: $license_types,
                details: ($license_list | if length > 50 then .[0:50] else . end)
            },
            sbom: {
                total_dependencies: $total_deps,
                direct_dependencies: $direct_deps,
                completeness: (if $total_deps > 0 then "complete" else "incomplete" end)
            },
            documentation: {
                has_readme: ($has_readme == "true"),
                has_license_file: ($has_license_file == "true"),
                has_contributing: ($has_contributing == "true"),
                score: $doc_score
            },
            ownership: {
                has_codeowners: ($has_codeowners == "true"),
                coverage_percentage: $coverage_pct
            }
        }'
}

# Generate org aggregate compliance data
generate_org_report_data() {
    local org="$1"
    local projects="$2"

    local project_count=$(echo "$projects" | wc -w | tr -d ' ')
    local total_violations=0
    local repos_failing=()
    local repos_missing_license=()

    for repo in $projects; do
        local project_id="$org/$repo"
        local analysis_path="$GIBSON_PROJECTS_DIR/$project_id/analysis"

        if [[ -d "$analysis_path" ]] && has_scanner_data "$analysis_path" "licenses"; then
            local lic=$(load_scanner_data "$analysis_path" "licenses")
            local v=$(echo "$lic" | jq -r '.summary.license_violations // 0')
            total_violations=$((total_violations + v))
            [[ $v -gt 0 ]] && repos_failing+=("$repo")
        fi

        if [[ -d "$analysis_path" ]] && has_scanner_data "$analysis_path" "documentation"; then
            local doc=$(load_scanner_data "$analysis_path" "documentation")
            local has_lic=$(echo "$doc" | jq -r '.summary.has_license // false')
            [[ "$has_lic" != "true" ]] && repos_missing_license+=("$repo")
        fi
    done

    local overall_status="PASS"
    [[ $total_violations -gt 0 ]] && overall_status="FAIL"

    jq -n \
        --arg org "$org" \
        --argjson project_count "$project_count" \
        --arg overall_status "$overall_status" \
        --argjson total_violations "$total_violations" \
        --arg repos_failing "$(printf '%s\n' "${repos_failing[@]}" | paste -sd, -)" \
        --arg repos_missing_license "$(printf '%s\n' "${repos_missing_license[@]}" | paste -sd, -)" \
        '{
            report_type: "compliance",
            report_version: "1.0.0",
            generated_at: (now | todate),
            organization: $org,
            projects: {
                count: $project_count
            },
            overall_status: $overall_status,
            licenses: {
                total_violations: $total_violations,
                repos_failing: ($repos_failing | split(",") | map(select(length > 0)))
            },
            documentation: {
                repos_missing_license: ($repos_missing_license | split(",") | map(select(length > 0)))
            }
        }'
}

export -f generate_report_data
export -f generate_org_report_data
