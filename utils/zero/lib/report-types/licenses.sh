#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# License Report Type
# Deep dive into license compliance - repository license, dependency licenses
#############################################################################

# Generate license report data as JSON
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

    # License data from scanner
    local license_data='{}'
    if has_scanner_data "$analysis_path" "licenses"; then
        license_data=$(load_scanner_data "$analysis_path" "licenses")
    fi

    # Extract repo license
    local repo_license=$(echo "$license_data" | jq -r '.repository_license.license // "Not Found"')
    local repo_license_file=$(echo "$license_data" | jq -r '.repository_license.file // null')

    # Overall status
    local overall_status=$(echo "$license_data" | jq -r '.summary.overall_status // "unknown"')
    local license_violations=$(echo "$license_data" | jq -r '.summary.license_violations // 0')
    local dep_violations=$(echo "$license_data" | jq -r '.summary.dependency_license_violations // 0')
    local total_deps_with_licenses=$(echo "$license_data" | jq -r '.summary.total_dependencies_with_licenses // 0')

    # Project license files
    local project_licenses=$(echo "$license_data" | jq '.licenses // []')

    # Dependency licenses
    local dep_by_license=$(echo "$license_data" | jq '.dependency_licenses.by_license // {}')
    local dep_denied=$(echo "$license_data" | jq '.dependency_licenses.denied // []')
    local dep_review=$(echo "$license_data" | jq '.dependency_licenses.review_required // []')

    # Content policy
    local profanity_count=$(echo "$license_data" | jq -r '.summary.profanity_issues // 0')
    local inclusive_count=$(echo "$license_data" | jq -r '.summary.inclusive_language_issues // 0')
    local content_profanity=$(echo "$license_data" | jq '.content_policy.profanity // []')
    local content_inclusive=$(echo "$license_data" | jq '.content_policy.inclusive_language // []')

    # Policy reference
    local policy=$(echo "$license_data" | jq '.policy // {}')

    # Build JSON
    jq -n \
        --arg project_id "$project_id" \
        --arg scan_id "$scan_id" \
        --arg profile "$profile" \
        --arg completed_at "$completed_at" \
        --arg overall_status "$overall_status" \
        --arg repo_license "$repo_license" \
        --arg repo_license_file "$repo_license_file" \
        --argjson license_violations "$license_violations" \
        --argjson dep_violations "$dep_violations" \
        --argjson total_deps "$total_deps_with_licenses" \
        --argjson project_licenses "$project_licenses" \
        --argjson dep_by_license "$dep_by_license" \
        --argjson dep_denied "$dep_denied" \
        --argjson dep_review "$dep_review" \
        --argjson profanity_count "$profanity_count" \
        --argjson inclusive_count "$inclusive_count" \
        --argjson content_profanity "$content_profanity" \
        --argjson content_inclusive "$content_inclusive" \
        --argjson policy "$policy" \
        '{
            report_type: "licenses",
            report_version: "1.0.0",
            generated_at: (now | todate),
            project: {
                id: $project_id,
                scan_id: $scan_id,
                profile: $profile,
                completed_at: $completed_at
            },
            overall_status: $overall_status,
            repository_license: {
                license: $repo_license,
                file: (if $repo_license_file == "null" or $repo_license_file == "" then null else $repo_license_file end)
            },
            summary: {
                project_license_violations: $license_violations,
                dependency_license_violations: $dep_violations,
                total_dependencies_scanned: $total_deps,
                denied_license_packages: ([$dep_denied[].count] | add // 0),
                review_required_packages: ([$dep_review[].count] | add // 0)
            },
            project_licenses: $project_licenses,
            dependency_licenses: {
                by_license: $dep_by_license,
                denied: $dep_denied,
                review_required: $dep_review
            },
            content_policy: {
                profanity_issues: $profanity_count,
                inclusive_language_issues: $inclusive_count,
                profanity_findings: $content_profanity,
                inclusive_language_findings: $content_inclusive
            },
            policy: $policy
        }'
}

# Generate org aggregate license data
generate_org_report_data() {
    local org="$1"
    local projects="$2"

    local project_count=$(echo "$projects" | wc -w | tr -d ' ')
    local total_violations=0
    local total_dep_violations=0
    local repos_with_violations=()
    local repos_with_gpl=()

    for repo in $projects; do
        local project_id="$org/$repo"
        local analysis_path="$ZERO_PROJECTS_DIR/$project_id/analysis"

        if [[ -d "$analysis_path" ]] && has_scanner_data "$analysis_path" "licenses"; then
            local lic=$(load_scanner_data "$analysis_path" "licenses")
            local v=$(echo "$lic" | jq -r '.summary.license_violations // 0')
            local dv=$(echo "$lic" | jq -r '.summary.dependency_license_violations // 0')

            total_violations=$((total_violations + v))
            total_dep_violations=$((total_dep_violations + dv))

            [[ $v -gt 0 ]] && repos_with_violations+=("$repo")
            [[ $dv -gt 0 ]] && repos_with_gpl+=("$repo")
        fi
    done

    local overall_status="pass"
    [[ $total_violations -gt 0 ]] && overall_status="fail"
    [[ $total_dep_violations -gt 0 ]] && overall_status="warning"

    jq -n \
        --arg org "$org" \
        --argjson project_count "$project_count" \
        --arg overall_status "$overall_status" \
        --argjson total_violations "$total_violations" \
        --argjson total_dep_violations "$total_dep_violations" \
        --arg repos_violations "$(printf '%s\n' "${repos_with_violations[@]}" | paste -sd, -)" \
        --arg repos_gpl "$(printf '%s\n' "${repos_with_gpl[@]}" | paste -sd, -)" \
        '{
            report_type: "licenses",
            report_version: "1.0.0",
            generated_at: (now | todate),
            organization: $org,
            projects: {
                count: $project_count,
                with_violations: ($repos_violations | split(",") | map(select(length > 0))),
                with_gpl_deps: ($repos_gpl | split(",") | map(select(length > 0)))
            },
            overall_status: $overall_status,
            summary: {
                project_license_violations: $total_violations,
                dependency_license_violations: $total_dep_violations
            }
        }'
}

export -f generate_report_data
export -f generate_org_report_data
