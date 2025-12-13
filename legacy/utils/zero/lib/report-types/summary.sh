#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Summary Report Type
# High-level overview of all findings
#############################################################################

# Generate summary report data as JSON
# Usage: generate_report_data <project_id> <analysis_path>
generate_report_data() {
    local project_id="$1"
    local analysis_path="$2"

    # Load manifest
    local manifest=$(load_manifest "$analysis_path")

    # Extract metadata
    local scan_id=$(echo "$manifest" | jq -r '.scan_id // "unknown"')
    local profile=$(echo "$manifest" | jq -r '.scan.profile // .mode // "standard"')

    # Get completed_at - prefer scan.completed_at, fallback to most recent analyzer timestamp
    local completed_at=$(echo "$manifest" | jq -r '
        if .scan.completed_at != null then .scan.completed_at
        else [.analyses[].completed_at | select(. != null)] | sort | last // ""
        end
    ')

    local duration=$(echo "$manifest" | jq -r '.scan.duration_seconds // 0')
    local commit_short=$(echo "$manifest" | jq -r '.git.commit_short // ""')
    local branch=$(echo "$manifest" | jq -r '.git.branch // ""')

    # Aggregate vulnerability data
    local vulns=$(aggregate_vulns "$analysis_path")
    local critical=$(echo "$vulns" | jq -r '.critical')
    local high=$(echo "$vulns" | jq -r '.high')
    local medium=$(echo "$vulns" | jq -r '.medium')
    local low=$(echo "$vulns" | jq -r '.low')
    local total_vulns=$(echo "$vulns" | jq -r '.total')

    # Calculate risk level
    local risk_level=$(calculate_risk_level "$critical" "$high" "$medium")

    # Aggregate dependency data
    local deps=$(aggregate_deps "$analysis_path")
    local total_deps=$(echo "$deps" | jq -r '.total')
    local direct_deps=$(echo "$deps" | jq -r '.direct')

    # Secrets count
    local secrets_count=0
    if has_scanner_data "$analysis_path" "code-secrets"; then
        secrets_count=$(load_scanner_data "$analysis_path" "code-secrets" | jq -r '.summary.total_findings // 0')
    fi

    # License status
    local license_status="unknown"
    local license_violations=0
    if has_scanner_data "$analysis_path" "licenses"; then
        local licenses=$(load_scanner_data "$analysis_path" "licenses")
        license_status=$(echo "$licenses" | jq -r '.summary.overall_status // "unknown"')
        license_violations=$(echo "$licenses" | jq -r '.summary.license_violations // 0')
    fi

    # Package health
    local abandoned=0
    local deprecated=0
    if has_scanner_data "$analysis_path" "package-health"; then
        local health=$(load_scanner_data "$analysis_path" "package-health")
        abandoned=$(echo "$health" | jq -r '.summary.abandoned // 0')
        deprecated=$(echo "$health" | jq -r '.summary.deprecated // 0')
    fi

    # DORA performance
    local dora_performance="N/A"
    if has_scanner_data "$analysis_path" "dora"; then
        dora_performance=$(load_scanner_data "$analysis_path" "dora" | jq -r '.summary.overall_performance // "N/A"')
    fi

    # Get top issues
    local top_issues=$(get_top_issues "$analysis_path" 5)

    # Get repo metadata (size, files)
    local repo_path="${analysis_path%/analysis}/repo"
    local repo_size="unknown"
    local repo_files=0
    if [[ -d "$repo_path" ]]; then
        repo_size=$(du -sh "$repo_path" 2>/dev/null | cut -f1 | tr -d '[:space:]')
        repo_files=$(find "$repo_path" -type f 2>/dev/null | wc -l | tr -d '[:space:]')
    fi

    # Get languages from tech-discovery
    local languages_json="[]"
    if has_scanner_data "$analysis_path" "tech-discovery"; then
        languages_json=$(load_scanner_data "$analysis_path" "tech-discovery" | jq '
            [.technologies[]? | select(.category | startswith("languages/")) | .name] | unique
        ')
    fi

    # Get package breakdown from SBOM (ecosystem and source files)
    local packages_json="[]"
    local sbom_file="$analysis_path/sbom.cdx.json"
    if [[ -f "$sbom_file" ]]; then
        packages_json=$(jq '
            [.components[]? | {
                ecosystem: (.properties[]? | select(.name == "syft:package:type") | .value),
                source: (.properties[]? | select(.name | startswith("syft:location")) | .value)
            } | select(.ecosystem != null)]
            | group_by(.ecosystem)
            | map({
                ecosystem: .[0].ecosystem,
                count: length,
                sources: ([.[].source] | unique | map(select(. != null)))
            })
        ' "$sbom_file" 2>/dev/null || echo "[]")
    fi

    # Get scan sources - which analyzers contributed data
    local scan_sources_json=$(echo "$manifest" | jq '
        [.analyses | to_entries[] | select(.value.status == "complete") | {
            scanner: .key,
            completed_at: .value.completed_at,
            duration_ms: .value.duration_ms
        }] | sort_by(.completed_at)
    ')

    # Build JSON output
    jq -n \
        --arg project_id "$project_id" \
        --arg scan_id "$scan_id" \
        --arg profile "$profile" \
        --arg completed_at "$completed_at" \
        --arg duration "$duration" \
        --arg commit_short "$commit_short" \
        --arg branch "$branch" \
        --argjson critical "$critical" \
        --argjson high "$high" \
        --argjson medium "$medium" \
        --argjson low "$low" \
        --argjson total_vulns "$total_vulns" \
        --argjson total_deps "$total_deps" \
        --argjson direct_deps "$direct_deps" \
        --argjson secrets_count "$secrets_count" \
        --arg license_status "$license_status" \
        --argjson license_violations "$license_violations" \
        --argjson abandoned "$abandoned" \
        --argjson deprecated "$deprecated" \
        --arg dora_performance "$dora_performance" \
        --arg top_issues "$top_issues" \
        --arg repo_size "$repo_size" \
        --argjson repo_files "$repo_files" \
        --argjson languages "$languages_json" \
        --argjson packages "$packages_json" \
        --argjson scan_sources "$scan_sources_json" \
        '{
            report_type: "summary",
            report_version: "1.0.0",
            generated_at: (now | todate),
            project: {
                id: $project_id,
                scan_id: $scan_id,
                profile: $profile,
                completed_at: $completed_at,
                duration_seconds: ($duration | tonumber),
                git: {
                    commit: $commit_short,
                    branch: $branch
                }
            },
            repository: {
                size: $repo_size,
                files: $repo_files,
                languages: $languages
            },
            vulnerabilities: {
                critical: $critical,
                high: $high,
                medium: $medium,
                low: $low,
                total: $total_vulns
            },
            dependencies: {
                total: $total_deps,
                direct: $direct_deps,
                abandoned: $abandoned,
                deprecated: $deprecated,
                packages: $packages
            },
            secrets: {
                exposed: $secrets_count
            },
            licenses: {
                status: $license_status,
                violations: $license_violations
            },
            dora: {
                performance: $dora_performance
            },
            scan_sources: $scan_sources,
            top_issues: ($top_issues | split("\n") | map(select(length > 0)))
        }'
}

# Generate org aggregate summary data
# Usage: generate_org_report_data <org> <projects>
generate_org_report_data() {
    local org="$1"
    local projects="$2"

    local project_count=$(echo "$projects" | wc -w | tr -d ' ')
    local total_vulns=0
    local total_critical=0
    local total_high=0
    local total_deps=0
    local at_risk_repos=()

    # Aggregate across projects
    for repo in $projects; do
        local project_id="$org/$repo"
        local analysis_path="$ZERO_PROJECTS_DIR/$project_id/analysis"

        if [[ -d "$analysis_path" ]]; then
            local vulns=$(aggregate_vulns "$analysis_path")
            local c=$(echo "$vulns" | jq -r '.critical')
            local h=$(echo "$vulns" | jq -r '.high')
            local t=$(echo "$vulns" | jq -r '.total')

            total_critical=$((total_critical + c))
            total_high=$((total_high + h))
            total_vulns=$((total_vulns + t))

            local deps=$(aggregate_deps "$analysis_path")
            local d=$(echo "$deps" | jq -r '.total')
            total_deps=$((total_deps + d))

            # Track at-risk repos
            if [[ $c -gt 0 ]] || [[ $h -gt 0 ]]; then
                at_risk_repos+=("$repo")
            fi
        fi
    done

    # Calculate overall risk
    local risk_level=$(calculate_risk_level "$total_critical" "$total_high" "0")

    # Build JSON
    jq -n \
        --arg org "$org" \
        --argjson project_count "$project_count" \
        --arg risk_level "$risk_level" \
        --argjson total_vulns "$total_vulns" \
        --argjson total_critical "$total_critical" \
        --argjson total_high "$total_high" \
        --argjson total_deps "$total_deps" \
        --arg at_risk "$(printf '%s\n' "${at_risk_repos[@]}" | paste -sd, -)" \
        '{
            report_type: "summary",
            report_version: "1.0.0",
            generated_at: (now | todate),
            organization: $org,
            projects: {
                count: $project_count,
                at_risk: ($at_risk | split(",") | map(select(length > 0)))
            },
            risk: {
                level: $risk_level,
                vulnerabilities: {
                    total: $total_vulns,
                    critical: $total_critical,
                    high: $total_high
                }
            },
            dependencies: {
                total: $total_deps
            }
        }'
}

export -f generate_report_data
export -f generate_org_report_data
