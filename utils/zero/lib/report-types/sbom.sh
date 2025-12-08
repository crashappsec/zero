#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# SBOM Report Type
# Software Bill of Materials
#############################################################################

# Generate SBOM report data as JSON
generate_report_data() {
    local project_id="$1"
    local analysis_path="$2"

    # Load manifest
    local manifest=$(load_manifest "$analysis_path")
    local scan_id=$(echo "$manifest" | jq -r '.scan_id // "unknown"')
    local profile=$(echo "$manifest" | jq -r '.scan.profile // "standard"')
    local completed_at=$(echo "$manifest" | jq -r '.scan.completed_at // ""')
    local commit=$(echo "$manifest" | jq -r '.git.commit_short // ""')
    local branch=$(echo "$manifest" | jq -r '.git.branch // ""')

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

    # License data for each package
    local license_data='{}'
    local license_list='[]'
    if has_scanner_data "$analysis_path" "licenses"; then
        license_data=$(load_scanner_data "$analysis_path" "licenses")
        license_list=$(echo "$license_data" | jq '.licenses // []')
    fi

    # Provenance data
    local prov_data='{}'
    if has_scanner_data "$analysis_path" "package-provenance"; then
        prov_data=$(load_scanner_data "$analysis_path" "package-provenance")
    fi
    local signed_packages=$(echo "$prov_data" | jq -r '.summary.signed_packages // 0')

    # Check for CycloneDX SBOM file
    local has_cyclonedx=false
    local cyclonedx_path="$analysis_path/sbom.cdx.json"
    [[ -f "$cyclonedx_path" ]] && has_cyclonedx=true

    # Build component list with license info
    local components=$(echo "$packages" | jq --argjson licenses "$license_list" '
        [.[] | . as $pkg | {
            name: (.name // .package // "unknown"),
            version: (.version // "unknown"),
            ecosystem: (.ecosystem // .type // "unknown"),
            direct: (.direct // false),
            license: (($licenses[] | select(.package == $pkg.name) | .license) // "unknown")
        }]
    ' 2>/dev/null || echo '[]')

    # Ecosystem breakdown
    local ecosystem_counts=$(echo "$packages" | jq '
        group_by(.ecosystem // .type // "unknown") |
        map({key: .[0].ecosystem // .[0].type // "unknown", value: length}) |
        from_entries
    ' 2>/dev/null || echo '{}')

    # Build JSON
    jq -n \
        --arg project_id "$project_id" \
        --arg scan_id "$scan_id" \
        --arg profile "$profile" \
        --arg completed_at "$completed_at" \
        --arg commit "$commit" \
        --arg branch "$branch" \
        --argjson total_deps "$total_deps" \
        --argjson direct_deps "$direct_deps" \
        --argjson ecosystem_counts "$ecosystem_counts" \
        --argjson components "$components" \
        --argjson signed_packages "$signed_packages" \
        --arg has_cyclonedx "$has_cyclonedx" \
        '{
            report_type: "sbom",
            report_version: "1.0.0",
            generated_at: (now | todate),
            sbom_format: "zero-sbom",
            sbom_version: "1.0",
            project: {
                id: $project_id,
                scan_id: $scan_id,
                profile: $profile,
                completed_at: $completed_at,
                git: {
                    commit: $commit,
                    branch: $branch
                }
            },
            summary: {
                total_components: $total_deps,
                direct_dependencies: $direct_deps,
                transitive_dependencies: ($total_deps - $direct_deps),
                signed_components: $signed_packages,
                ecosystems: $ecosystem_counts
            },
            formats_available: {
                cyclonedx: ($has_cyclonedx == "true"),
                zero_json: true
            },
            components: $components
        }'
}

# Generate org aggregate SBOM data
generate_org_report_data() {
    local org="$1"
    local projects="$2"

    local project_count=$(echo "$projects" | wc -w | tr -d ' ')
    local total_deps=0
    local all_ecosystems=()

    for repo in $projects; do
        local project_id="$org/$repo"
        local analysis_path="$ZERO_PROJECTS_DIR/$project_id/analysis"

        if [[ -d "$analysis_path" ]]; then
            local deps=$(aggregate_deps "$analysis_path")
            local d=$(echo "$deps" | jq -r '.total')
            total_deps=$((total_deps + d))
        fi
    done

    jq -n \
        --arg org "$org" \
        --argjson project_count "$project_count" \
        --argjson total_deps "$total_deps" \
        '{
            report_type: "sbom",
            report_version: "1.0.0",
            generated_at: (now | todate),
            organization: $org,
            projects: {
                count: $project_count
            },
            summary: {
                total_components: $total_deps
            }
        }'
}

export -f generate_report_data
export -f generate_org_report_data
