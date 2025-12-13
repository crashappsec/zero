#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Code Ownership Report Type
# 3-tier ownership analysis: Basic → Analysis → AI Insights
#############################################################################

# Generate code-ownership report data as JSON
# Usage: generate_report_data <project_id> <analysis_path>
generate_report_data() {
    local project_id="$1"
    local analysis_path="$2"

    # Load manifest
    local manifest=$(load_manifest "$analysis_path")

    # Extract metadata
    local scan_id=$(echo "$manifest" | jq -r '.scan_id // "unknown"')
    local profile=$(echo "$manifest" | jq -r '.scan.profile // .mode // "standard"')
    local completed_at=$(echo "$manifest" | jq -r '
        if .scan.completed_at != null then .scan.completed_at
        else [.analyses[].completed_at | select(. != null)] | sort | last // ""
        end
    ')
    local commit_short=$(echo "$manifest" | jq -r '.git.commit_short // ""')
    local branch=$(echo "$manifest" | jq -r '.git.branch // ""')

    #########################################################################
    # TIER 1: BASIC - CODEOWNERS file detection
    #########################################################################
    local tier1_available=false
    local codeowners_exists=false
    local codeowners_path=""
    local codeowners_valid="unknown"
    local codeowners_patterns=0
    local codeowners_owners=0
    local codeowners_issues="[]"

    if has_scanner_data "$analysis_path" "code-ownership"; then
        tier1_available=true
        local ownership_data=$(load_scanner_data "$analysis_path" "code-ownership")

        codeowners_exists=$(echo "$ownership_data" | jq -r '.codeowners.exists // false')
        if [[ "$codeowners_exists" == "true" ]]; then
            codeowners_path=$(echo "$ownership_data" | jq -r '.codeowners.path // ""')
            codeowners_valid=$(echo "$ownership_data" | jq -r '.codeowners.validation.status // "unknown"')
            codeowners_patterns=$(echo "$ownership_data" | jq -r '.codeowners.total_patterns // 0')
            codeowners_owners=$(echo "$ownership_data" | jq -r '.codeowners.unique_owners // 0')
            codeowners_issues=$(echo "$ownership_data" | jq '.codeowners.validation.issues // []')
        fi
    fi

    #########################################################################
    # TIER 2: ANALYSIS - Bus factor, concentration metrics
    #########################################################################
    local tier2_available=false
    local bus_factor=0
    local risk_level="unknown"
    local risk_description=""
    local gini_coefficient=0
    local top_contributor_pct=0
    local top3_pct=0
    local total_commits=0
    local active_contributors=0
    local contributors_json="[]"
    local bus_factor_contributors="[]"
    local period_days=90

    if has_scanner_data "$analysis_path" "bus-factor"; then
        tier2_available=true
        local bus_data=$(load_scanner_data "$analysis_path" "bus-factor")

        bus_factor=$(echo "$bus_data" | jq -r '.summary.bus_factor // 0')
        risk_level=$(echo "$bus_data" | jq -r '.summary.risk_level // "unknown"')
        risk_description=$(echo "$bus_data" | jq -r '.bus_factor_analysis.risk_description // ""')
        gini_coefficient=$(echo "$bus_data" | jq -r '.concentration_metrics.gini_coefficient // 0')
        top_contributor_pct=$(echo "$bus_data" | jq -r '.concentration_metrics.top_contributor_percentage // 0')
        top3_pct=$(echo "$bus_data" | jq -r '.concentration_metrics.top_3_contributors_percentage // 0')
        total_commits=$(echo "$bus_data" | jq -r '.summary.total_commits // 0')
        active_contributors=$(echo "$bus_data" | jq -r '.summary.active_contributors // 0')
        contributors_json=$(echo "$bus_data" | jq '.contributors // []')
        bus_factor_contributors=$(echo "$bus_data" | jq '.bus_factor_analysis.contributors_for_threshold // []')
        period_days=$(echo "$bus_data" | jq -r '.period_days // 90')
    elif has_scanner_data "$analysis_path" "code-ownership"; then
        # Fallback to code-ownership data if bus-factor not available
        local ownership_data=$(load_scanner_data "$analysis_path" "code-ownership")
        total_commits=$(echo "$ownership_data" | jq -r '.summary.total_commits // 0')
        active_contributors=$(echo "$ownership_data" | jq -r '.summary.active_contributors // 0')
        contributors_json=$(echo "$ownership_data" | jq '.contributors // []')
        period_days=$(echo "$ownership_data" | jq -r '.period_days // 90')
    fi

    #########################################################################
    # TIER 3: AI INSIGHTS - Claude analysis
    #########################################################################
    local tier3_available=false
    local ai_insights="[]"
    local ai_recommendations="[]"
    local succession_plan="null"
    local knowledge_silos="[]"
    local suggested_codeowners="[]"
    local risk_areas="[]"
    local action_items="[]"
    local ai_model=""

    # Check if we have Claude-generated analysis
    if has_scanner_data "$analysis_path" "code-ownership-ai"; then
        tier3_available=true
        local ai_data=$(load_scanner_data "$analysis_path" "code-ownership-ai")
        ai_insights=$(echo "$ai_data" | jq '.insights // []')
        ai_recommendations=$(echo "$ai_data" | jq '.recommendations // []')
        succession_plan=$(echo "$ai_data" | jq '.succession_plan // null')
        knowledge_silos=$(echo "$ai_data" | jq '.knowledge_silos // []')
        suggested_codeowners=$(echo "$ai_data" | jq '.suggested_codeowners // []')
        risk_areas=$(echo "$ai_data" | jq '.risk_areas // []')
        action_items=$(echo "$ai_data" | jq '.action_items // []')
        ai_model=$(echo "$ai_data" | jq -r '.model // ""')
    fi

    # Calculate derived metrics
    local ownership_health="unknown"
    if [[ "$tier2_available" == "true" ]]; then
        if [[ "$bus_factor" -le 1 ]]; then
            ownership_health="critical"
        elif [[ "$bus_factor" -le 2 ]]; then
            ownership_health="poor"
        elif [[ "$bus_factor" -le 3 ]]; then
            ownership_health="fair"
        else
            ownership_health="good"
        fi
    fi

    # Build output JSON
    jq -n \
        --arg project_id "$project_id" \
        --arg scan_id "$scan_id" \
        --arg profile "$profile" \
        --arg completed_at "$completed_at" \
        --arg commit_short "$commit_short" \
        --arg branch "$branch" \
        --arg report_type "code-ownership" \
        --argjson tier1_available "$tier1_available" \
        --argjson tier2_available "$tier2_available" \
        --argjson tier3_available "$tier3_available" \
        --argjson codeowners_exists "$codeowners_exists" \
        --arg codeowners_path "$codeowners_path" \
        --arg codeowners_valid "$codeowners_valid" \
        --argjson codeowners_patterns "$codeowners_patterns" \
        --argjson codeowners_owners "$codeowners_owners" \
        --argjson codeowners_issues "$codeowners_issues" \
        --argjson bus_factor "$bus_factor" \
        --arg risk_level "$risk_level" \
        --arg risk_description "$risk_description" \
        --argjson gini_coefficient "$gini_coefficient" \
        --argjson top_contributor_pct "$top_contributor_pct" \
        --argjson top3_pct "$top3_pct" \
        --argjson total_commits "$total_commits" \
        --argjson active_contributors "$active_contributors" \
        --argjson period_days "$period_days" \
        --argjson contributors "$contributors_json" \
        --argjson bus_factor_contributors "$bus_factor_contributors" \
        --arg ownership_health "$ownership_health" \
        --argjson ai_insights "$ai_insights" \
        --argjson ai_recommendations "$ai_recommendations" \
        --argjson succession_plan "$succession_plan" \
        --argjson knowledge_silos "$knowledge_silos" \
        --argjson suggested_codeowners "$suggested_codeowners" \
        --argjson risk_areas "$risk_areas" \
        --argjson action_items "$action_items" \
        --arg ai_model "$ai_model" \
        --arg generated_at "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        '{
            report_type: $report_type,
            generated_at: $generated_at,
            project: {
                id: $project_id,
                scan_id: $scan_id,
                profile: $profile,
                completed_at: $completed_at,
                commit_short: $commit_short,
                branch: $branch
            },
            tiers: {
                basic: $tier1_available,
                analysis: $tier2_available,
                ai_insights: $tier3_available
            },
            tier1_basic: {
                available: $tier1_available,
                codeowners: {
                    exists: $codeowners_exists,
                    path: $codeowners_path,
                    valid: $codeowners_valid,
                    total_patterns: $codeowners_patterns,
                    unique_owners: $codeowners_owners,
                    issues: $codeowners_issues
                }
            },
            tier2_analysis: {
                available: $tier2_available,
                period_days: $period_days,
                summary: {
                    total_commits: $total_commits,
                    active_contributors: $active_contributors,
                    bus_factor: $bus_factor,
                    risk_level: $risk_level,
                    ownership_health: $ownership_health
                },
                bus_factor: {
                    value: $bus_factor,
                    risk_level: $risk_level,
                    risk_description: $risk_description,
                    contributors_for_threshold: $bus_factor_contributors
                },
                concentration: {
                    gini_coefficient: $gini_coefficient,
                    top_contributor_percentage: $top_contributor_pct,
                    top_3_contributors_percentage: $top3_pct
                },
                contributors: $contributors
            },
            tier3_ai_insights: {
                available: $tier3_available,
                model: $ai_model,
                insights: $ai_insights,
                recommendations: $ai_recommendations,
                succession_plan: $succession_plan,
                knowledge_silos: $knowledge_silos,
                suggested_codeowners: $suggested_codeowners,
                risk_areas: $risk_areas,
                action_items: $action_items
            }
        }'
}

export -f generate_report_data
