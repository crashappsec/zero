#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# DORA Report Type
# DevOps performance metrics
#############################################################################

# Generate DORA report data as JSON
generate_report_data() {
    local project_id="$1"
    local analysis_path="$2"

    # Load manifest
    local manifest=$(load_manifest "$analysis_path")
    local scan_id=$(echo "$manifest" | jq -r '.scan_id // "unknown"')
    local profile=$(echo "$manifest" | jq -r '.scan.profile // "standard"')
    local completed_at=$(echo "$manifest" | jq -r '.scan.completed_at // ""')

    # DORA metrics
    local dora_data='{}'
    if has_scanner_data "$analysis_path" "dora"; then
        dora_data=$(load_scanner_data "$analysis_path" "dora")
    fi

    local overall_perf=$(echo "$dora_data" | jq -r '.summary.overall_performance // "N/A"')
    local deploy_freq=$(echo "$dora_data" | jq -r '.summary.deployment_frequency // "N/A"')
    local deploy_freq_value=$(echo "$dora_data" | jq -r '.metrics.deployment_frequency.value // 0')
    local deploy_freq_unit=$(echo "$dora_data" | jq -r '.metrics.deployment_frequency.unit // "per_day"')

    local lead_time=$(echo "$dora_data" | jq -r '.summary.lead_time // "N/A"')
    local lead_time_hours=$(echo "$dora_data" | jq -r '.metrics.lead_time.hours // 0')

    local change_failure=$(echo "$dora_data" | jq -r '.summary.change_failure_rate // "N/A"')
    local change_failure_pct=$(echo "$dora_data" | jq -r '.metrics.change_failure_rate.percentage // 0')

    local mttr=$(echo "$dora_data" | jq -r '.summary.mttr // "N/A"')
    local mttr_hours=$(echo "$dora_data" | jq -r '.metrics.mttr.hours // 0')

    # Git insights for additional context
    local git_data='{}'
    if has_scanner_data "$analysis_path" "git"; then
        git_data=$(load_scanner_data "$analysis_path" "git")
    fi

    local total_commits=$(echo "$git_data" | jq -r '.summary.total_commits // 0')
    local contributors=$(echo "$git_data" | jq -r '.summary.contributors // 0')
    local avg_commits_per_day=$(echo "$git_data" | jq -r '.summary.avg_commits_per_day // 0')
    local merge_frequency=$(echo "$git_data" | jq -r '.summary.merge_frequency // "N/A"')

    # Test coverage for quality context
    local test_data='{}'
    if has_scanner_data "$analysis_path" "test-coverage"; then
        test_data=$(load_scanner_data "$analysis_path" "test-coverage")
    fi
    local test_coverage=$(echo "$test_data" | jq -r '.summary.coverage_percentage // 0')
    local has_tests=$(echo "$test_data" | jq -r '.summary.has_tests // false')

    # Performance level interpretation
    local perf_description=""
    case "$overall_perf" in
        ELITE) perf_description="Elite performers deploy multiple times per day with minimal failures" ;;
        HIGH) perf_description="High performers deploy frequently with good recovery times" ;;
        MEDIUM) perf_description="Medium performers have room for improvement in deployment velocity" ;;
        LOW) perf_description="Low performers should focus on automation and process improvements" ;;
        *) perf_description="DORA metrics not available - run with a profile that includes DORA analysis" ;;
    esac

    # Recommendations
    local recommendations=()
    if [[ "$overall_perf" == "LOW" ]] || [[ "$overall_perf" == "MEDIUM" ]]; then
        [[ "$deploy_freq" == "weekly" ]] || [[ "$deploy_freq" == "monthly" ]] && \
            recommendations+=("Increase deployment frequency through better CI/CD automation")
        [[ $lead_time_hours -gt 168 ]] && \
            recommendations+=("Reduce lead time by streamlining code review process")
        [[ $change_failure_pct -gt 15 ]] && \
            recommendations+=("Lower change failure rate through better testing and staged rollouts")
    fi
    [[ "$test_coverage" -lt 60 ]] && [[ "$has_tests" == "true" ]] && \
        recommendations+=("Improve test coverage (currently ${test_coverage}%)")

    # Build JSON
    jq -n \
        --arg project_id "$project_id" \
        --arg scan_id "$scan_id" \
        --arg profile "$profile" \
        --arg completed_at "$completed_at" \
        --arg overall_perf "$overall_perf" \
        --arg perf_description "$perf_description" \
        --arg deploy_freq "$deploy_freq" \
        --argjson deploy_freq_value "$deploy_freq_value" \
        --arg deploy_freq_unit "$deploy_freq_unit" \
        --arg lead_time "$lead_time" \
        --argjson lead_time_hours "$lead_time_hours" \
        --arg change_failure "$change_failure" \
        --argjson change_failure_pct "$change_failure_pct" \
        --arg mttr "$mttr" \
        --argjson mttr_hours "$mttr_hours" \
        --argjson total_commits "$total_commits" \
        --argjson contributors "$contributors" \
        --argjson avg_commits_per_day "$avg_commits_per_day" \
        --argjson test_coverage "$test_coverage" \
        --arg recommendations "$(printf '%s\n' "${recommendations[@]}")" \
        '{
            report_type: "dora",
            report_version: "1.0.0",
            generated_at: (now | todate),
            project: {
                id: $project_id,
                scan_id: $scan_id,
                profile: $profile,
                completed_at: $completed_at
            },
            dora: {
                overall_performance: $overall_perf,
                description: $perf_description,
                metrics: {
                    deployment_frequency: {
                        level: $deploy_freq,
                        value: $deploy_freq_value,
                        unit: $deploy_freq_unit
                    },
                    lead_time: {
                        level: $lead_time,
                        hours: $lead_time_hours
                    },
                    change_failure_rate: {
                        level: $change_failure,
                        percentage: $change_failure_pct
                    },
                    mttr: {
                        level: $mttr,
                        hours: $mttr_hours
                    }
                }
            },
            git_activity: {
                total_commits: $total_commits,
                contributors: $contributors,
                avg_commits_per_day: $avg_commits_per_day
            },
            quality: {
                test_coverage_percentage: $test_coverage
            },
            recommendations: ($recommendations | split("\n") | map(select(length > 0)))
        }'
}

# Generate org aggregate DORA data
generate_org_report_data() {
    local org="$1"
    local projects="$2"

    local project_count=$(echo "$projects" | wc -w | tr -d ' ')
    local elite_count=0
    local high_count=0
    local medium_count=0
    local low_count=0
    local na_count=0

    for repo in $projects; do
        local project_id="$org/$repo"
        local analysis_path="$GIBSON_PROJECTS_DIR/$project_id/analysis"

        if [[ -d "$analysis_path" ]] && has_scanner_data "$analysis_path" "dora"; then
            local perf=$(load_scanner_data "$analysis_path" "dora" | jq -r '.summary.overall_performance // "N/A"')
            case "$perf" in
                ELITE) elite_count=$((elite_count + 1)) ;;
                HIGH) high_count=$((high_count + 1)) ;;
                MEDIUM) medium_count=$((medium_count + 1)) ;;
                LOW) low_count=$((low_count + 1)) ;;
                *) na_count=$((na_count + 1)) ;;
            esac
        else
            na_count=$((na_count + 1))
        fi
    done

    # Determine overall org performance
    local org_perf="N/A"
    if [[ $elite_count -gt $((project_count / 2)) ]]; then
        org_perf="ELITE"
    elif [[ $((elite_count + high_count)) -gt $((project_count / 2)) ]]; then
        org_perf="HIGH"
    elif [[ $low_count -gt $((project_count / 2)) ]]; then
        org_perf="LOW"
    elif [[ $na_count -lt $project_count ]]; then
        org_perf="MEDIUM"
    fi

    jq -n \
        --arg org "$org" \
        --argjson project_count "$project_count" \
        --arg org_perf "$org_perf" \
        --argjson elite_count "$elite_count" \
        --argjson high_count "$high_count" \
        --argjson medium_count "$medium_count" \
        --argjson low_count "$low_count" \
        --argjson na_count "$na_count" \
        '{
            report_type: "dora",
            report_version: "1.0.0",
            generated_at: (now | todate),
            organization: $org,
            projects: {
                count: $project_count
            },
            dora: {
                overall_performance: $org_perf,
                distribution: {
                    elite: $elite_count,
                    high: $high_count,
                    medium: $medium_count,
                    low: $low_count,
                    not_available: $na_count
                }
            }
        }'
}

export -f generate_report_data
export -f generate_org_report_data
