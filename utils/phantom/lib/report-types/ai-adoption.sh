#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# AI Adoption Report Type
# Combines technology detection with code ownership to show AI adoption
#############################################################################

# Get script directory for loading AI categories
_AI_ADOPTION_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
_REPO_ROOT="$(dirname "$(dirname "$(dirname "$(dirname "$_AI_ADOPTION_DIR")")")")"
_AI_CATEGORIES_FILE="$_REPO_ROOT/rag/ai-adoption/ai-categories.json"

# Load AI category patterns
_load_ai_categories() {
    if [[ -f "$_AI_CATEGORIES_FILE" ]]; then
        cat "$_AI_CATEGORIES_FILE"
    else
        # Fallback defaults
        echo '{"ai_category_patterns": ["ai-ml/*", "genai-tools/*"]}'
    fi
}

# Check if a category matches AI patterns
_is_ai_category() {
    local category="$1"
    local patterns=$(echo "$(_load_ai_categories)" | jq -r '.ai_category_patterns[]')

    while IFS= read -r pattern; do
        # Convert glob pattern to regex (simple * wildcard)
        local regex="^${pattern//\*/.*}$"
        if [[ "$category" =~ $regex ]]; then
            return 0
        fi
    done <<< "$patterns"

    return 1
}

# Get friendly label for category
_get_category_label() {
    local category="$1"
    local label=$(echo "$(_load_ai_categories)" | jq -r --arg cat "$category" '.categories[$cat].label // empty')

    if [[ -n "$label" ]]; then
        echo "$label"
    else
        # Generate label from category path
        echo "$category" | sed 's|.*/||' | sed 's/-/ /g' | awk '{for(i=1;i<=NF;i++) $i=toupper(substr($i,1,1)) tolower(substr($i,2))}1'
    fi
}

# Generate ai-adoption report data as JSON
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
    # AI TECHNOLOGIES - Filter tech-discovery for AI categories
    #########################################################################
    local ai_technologies="[]"
    local ai_tech_count=0
    local ai_categories_found="{}"

    if has_scanner_data "$analysis_path" "tech-discovery"; then
        local tech_data=$(load_scanner_data "$analysis_path" "tech-discovery")

        # Filter technologies for AI categories
        ai_technologies=$(echo "$tech_data" | jq '
            [.technologies[]? | select(
                .category |
                (startswith("ai-ml/") or startswith("genai-tools"))
            )]
        ')

        ai_tech_count=$(echo "$ai_technologies" | jq 'length')

        # Group by category
        ai_categories_found=$(echo "$ai_technologies" | jq '
            group_by(.category) |
            map({
                key: .[0].category,
                value: {
                    count: length,
                    technologies: [.[].name]
                }
            }) |
            from_entries
        ')
    fi

    # Also check for AI in technology.json (older format)
    if [[ "$ai_tech_count" -eq 0 ]] && has_scanner_data "$analysis_path" "technology"; then
        local tech_data=$(load_scanner_data "$analysis_path" "technology")

        ai_technologies=$(echo "$tech_data" | jq '
            [.technologies[]? | select(
                .category |
                (startswith("ai-ml/") or startswith("genai-tools"))
            )]
        ')

        ai_tech_count=$(echo "$ai_technologies" | jq 'length')

        ai_categories_found=$(echo "$ai_technologies" | jq '
            group_by(.category) |
            map({
                key: .[0].category,
                value: {
                    count: length,
                    technologies: [.[].name]
                }
            }) |
            from_entries
        ')
    fi

    #########################################################################
    # CODE OWNERSHIP - Get contributor data
    #########################################################################
    local contributors="[]"
    local total_contributors=0
    local bus_factor=0
    local bus_factor_risk="unknown"

    if has_scanner_data "$analysis_path" "code-ownership"; then
        local ownership_data=$(load_scanner_data "$analysis_path" "code-ownership")

        contributors=$(echo "$ownership_data" | jq '.contributors // []')
        total_contributors=$(echo "$ownership_data" | jq '.summary.active_contributors // 0')
        bus_factor=$(echo "$ownership_data" | jq '.summary.estimated_bus_factor // 0')
    fi

    # Check bus-factor scanner for more detailed data
    if has_scanner_data "$analysis_path" "bus-factor"; then
        local bus_data=$(load_scanner_data "$analysis_path" "bus-factor")
        bus_factor=$(echo "$bus_data" | jq '.summary.bus_factor // 0')
        bus_factor_risk=$(echo "$bus_data" | jq -r '.summary.risk_level // "unknown"')

        # Prefer bus-factor contributors if available (more detailed)
        local bf_contributors=$(echo "$bus_data" | jq '.contributors // []')
        if [[ $(echo "$bf_contributors" | jq 'length') -gt 0 ]]; then
            contributors="$bf_contributors"
        fi
    fi

    # Get top 10 contributors
    local top_contributors=$(echo "$contributors" | jq '
        sort_by(-.commits) |
        .[0:10] |
        map({
            name: .name,
            email: .email,
            commits: .commits,
            lines_added: .lines_added,
            percentage: (if .percentage then .percentage else 0 end)
        })
    ')

    #########################################################################
    # BUILD CATEGORY DETAILS with labels
    #########################################################################
    local categories_with_labels="[]"

    # Process each category found
    local cat_keys=$(echo "$ai_categories_found" | jq -r 'keys[]')
    while IFS= read -r cat_key; do
        [[ -z "$cat_key" ]] && continue

        local label=$(_get_category_label "$cat_key")
        local cat_info=$(echo "$ai_categories_found" | jq --arg key "$cat_key" '.[$key]')
        local count=$(echo "$cat_info" | jq '.count')
        local techs=$(echo "$cat_info" | jq '.technologies')

        categories_with_labels=$(echo "$categories_with_labels" | jq \
            --arg cat "$cat_key" \
            --arg label "$label" \
            --argjson count "$count" \
            --argjson techs "$techs" \
            '. + [{category: $cat, label: $label, count: $count, technologies: $techs}]')
    done <<< "$cat_keys"

    # Sort categories by count descending
    categories_with_labels=$(echo "$categories_with_labels" | jq 'sort_by(-.count)')

    #########################################################################
    # ADOPTION SUMMARY
    #########################################################################
    local has_ai_adoption=false
    [[ "$ai_tech_count" -gt 0 ]] && has_ai_adoption=true

    local category_count=$(echo "$categories_with_labels" | jq 'length')

    #########################################################################
    # BUILD OUTPUT JSON
    #########################################################################
    jq -n \
        --arg report_type "ai-adoption" \
        --arg generated_at "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        --arg project_id "$project_id" \
        --arg scan_id "$scan_id" \
        --arg profile "$profile" \
        --arg completed_at "$completed_at" \
        --arg commit_short "$commit_short" \
        --arg branch "$branch" \
        --argjson has_ai_adoption "$has_ai_adoption" \
        --argjson ai_tech_count "$ai_tech_count" \
        --argjson category_count "$category_count" \
        --argjson total_contributors "$total_contributors" \
        --argjson bus_factor "$bus_factor" \
        --arg bus_factor_risk "$bus_factor_risk" \
        --argjson ai_technologies "$ai_technologies" \
        --argjson categories "$categories_with_labels" \
        --argjson top_contributors "$top_contributors" \
        '{
            report_type: $report_type,
            report_version: "1.0.0",
            generated_at: $generated_at,
            project: {
                id: $project_id,
                scan_id: $scan_id,
                profile: $profile,
                completed_at: $completed_at,
                git: {
                    commit: $commit_short,
                    branch: $branch
                }
            },
            summary: {
                has_ai_adoption: $has_ai_adoption,
                ai_technologies_count: $ai_tech_count,
                ai_categories_count: $category_count,
                total_contributors: $total_contributors,
                bus_factor: $bus_factor,
                bus_factor_risk: $bus_factor_risk
            },
            ai_technologies: $ai_technologies,
            categories: $categories,
            contributors: $top_contributors,
            governance: {
                note: "Phase 1 MVP - shows AI technologies and contributors separately. Phase 2 will add file-level correlation."
            }
        }'
}

export -f generate_report_data
