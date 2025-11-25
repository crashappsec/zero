#!/bin/bash
# Technical Debt Scorer
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0
#
# Scores technical debt in dependencies based on multiple factors.
# Part of the Developer Productivity module.

set -eo pipefail

# Get script directory
LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUNDLE_DIR="$(dirname "$LIB_DIR")"
SUPPLY_CHAIN_DIR="$(dirname "$BUNDLE_DIR")"

# Load deps.dev client from global lib
if [[ -f "$SUPPLY_CHAIN_DIR/lib/deps-dev-client.sh" ]]; then
    source "$SUPPLY_CHAIN_DIR/lib/deps-dev-client.sh"
fi

# Load related modules if available
if [[ -f "$SUPPLY_CHAIN_DIR/package-health-analysis/lib/abandonment-detector.sh" ]]; then
    source "$SUPPLY_CHAIN_DIR/package-health-analysis/lib/abandonment-detector.sh"
fi
if [[ -f "$SUPPLY_CHAIN_DIR/library-recommendations/lib/recommender.sh" ]]; then
    source "$SUPPLY_CHAIN_DIR/library-recommendations/lib/recommender.sh"
fi

#############################################################################
# Debt Scoring Weights
#############################################################################

# Weight factors for debt calculation (0-100)
WEIGHT_OUTDATED=25          # Outdated major versions
WEIGHT_DEPRECATED=30        # Deprecated packages
WEIGHT_ABANDONED=35         # Abandoned packages
WEIGHT_SECURITY=40          # Security vulnerabilities
WEIGHT_REPLACEMENTS=20      # Has better alternatives
WEIGHT_SIZE=15              # Oversized packages
WEIGHT_LICENSE=25           # License concerns
WEIGHT_MAINTENANCE=20       # Poor maintenance score

#############################################################################
# Debt Factor Scoring
#############################################################################

# Score outdated versions
# Usage: score_outdated <current_version> <latest_version>
score_outdated() {
    local current="$1"
    local latest="$2"

    # Extract major versions
    local current_major=$(echo "$current" | cut -d'.' -f1 | tr -dc '0-9')
    local latest_major=$(echo "$latest" | cut -d'.' -f1 | tr -dc '0-9')

    [[ -z "$current_major" ]] && current_major=0
    [[ -z "$latest_major" ]] && latest_major=0

    local diff=$((latest_major - current_major))

    if [[ $diff -le 0 ]]; then
        echo "0"  # Up to date
    elif [[ $diff -eq 1 ]]; then
        echo "30"  # One major version behind
    elif [[ $diff -eq 2 ]]; then
        echo "60"  # Two major versions behind
    else
        echo "100"  # Three or more major versions behind
    fi
}

# Score abandonment status
# Usage: score_abandonment <status>
score_abandonment() {
    local status="$1"

    case "$status" in
        healthy)
            echo "0"
            ;;
        stale)
            echo "40"
            ;;
        deprecated)
            echo "70"
            ;;
        abandoned)
            echo "90"
            ;;
        archived)
            echo "100"
            ;;
        *)
            echo "20"  # Unknown
            ;;
    esac
}

# Score based on replacement availability
# Usage: score_replacement <has_replacement> <migration_effort>
score_replacement() {
    local has_replacement="$1"
    local effort="${2:-moderate}"

    if [[ "$has_replacement" != "true" ]]; then
        echo "0"
        return
    fi

    case "$effort" in
        trivial)
            echo "80"  # Easy to replace, should do it
            ;;
        easy)
            echo "60"
            ;;
        moderate)
            echo "40"
            ;;
        significant)
            echo "20"
            ;;
        major)
            echo "10"  # Hard to replace
            ;;
        *)
            echo "30"
            ;;
    esac
}

# Score based on OpenSSF scorecard
# Usage: score_maintenance <openssf_score>
score_maintenance() {
    local score="$1"

    if [[ "$score" == "null" || -z "$score" ]]; then
        echo "30"  # Unknown = some concern
        return
    fi

    # OpenSSF scores are 0-10, higher is better
    # Convert to debt score (0-100, higher is more debt)
    if [[ $(echo "$score >= 8" | bc -l 2>/dev/null || echo "0") == "1" ]]; then
        echo "0"
    elif [[ $(echo "$score >= 6" | bc -l 2>/dev/null || echo "0") == "1" ]]; then
        echo "20"
    elif [[ $(echo "$score >= 4" | bc -l 2>/dev/null || echo "0") == "1" ]]; then
        echo "50"
    elif [[ $(echo "$score >= 2" | bc -l 2>/dev/null || echo "0") == "1" ]]; then
        echo "80"
    else
        echo "100"
    fi
}

# Score based on package size
# Usage: score_size <size_bytes> <category>
score_size() {
    local size="$1"
    local category="${2:-general}"

    # Different thresholds for different categories
    local threshold=50000  # 50KB default

    case "$category" in
        utility)
            threshold=10000  # Utilities should be small
            ;;
        framework)
            threshold=200000  # Frameworks can be larger
            ;;
        ui-component)
            threshold=30000
            ;;
    esac

    if [[ $size -lt $threshold ]]; then
        echo "0"
    elif [[ $size -lt $((threshold * 2)) ]]; then
        echo "30"
    elif [[ $size -lt $((threshold * 5)) ]]; then
        echo "60"
    else
        echo "100"
    fi
}

#############################################################################
# Comprehensive Debt Calculation
#############################################################################

# Calculate overall debt score for a package
# Usage: calculate_debt_score <package> <ecosystem> [version]
calculate_debt_score() {
    local pkg="$1"
    local ecosystem="${2:-npm}"
    local version="${3:-}"

    local factors="[]"
    local total_weighted_score=0
    local total_weight=0

    # Factor 1: Abandonment status
    if type check_abandonment_status &>/dev/null; then
        local abandonment=$(check_abandonment_status "$pkg" "$ecosystem" 2>/dev/null || echo '{"status": "unknown"}')
        local status=$(echo "$abandonment" | jq -r '.status // "unknown"')
        local abandonment_score=$(score_abandonment "$status")

        factors=$(echo "$factors" | jq --arg name "abandonment" --arg status "$status" --argjson score "$abandonment_score" --argjson weight "$WEIGHT_ABANDONED" \
            '. + [{"factor": $name, "status": $status, "score": $score, "weight": $weight}]')

        total_weighted_score=$((total_weighted_score + (abandonment_score * WEIGHT_ABANDONED)))
        total_weight=$((total_weight + WEIGHT_ABANDONED))
    fi

    # Factor 2: Replacement availability
    if type has_replacement &>/dev/null; then
        local has_rep=$(has_replacement "$pkg" "$ecosystem" 2>/dev/null || echo "false")
        local rep_score=0

        if [[ "$has_rep" == "true" ]]; then
            local replacements=$(get_replacements "$pkg" "$ecosystem" 2>/dev/null || echo "[]")
            local effort=$(echo "$replacements" | jq -r '.[0].migration_effort // "moderate"')
            rep_score=$(score_replacement "$has_rep" "$effort")

            factors=$(echo "$factors" | jq --arg name "replacement_available" --arg effort "$effort" --argjson score "$rep_score" --argjson weight "$WEIGHT_REPLACEMENTS" \
                '. + [{"factor": $name, "migration_effort": $effort, "score": $score, "weight": $weight}]')

            total_weighted_score=$((total_weighted_score + (rep_score * WEIGHT_REPLACEMENTS)))
            total_weight=$((total_weight + WEIGHT_REPLACEMENTS))
        fi
    fi

    # Factor 3: Maintenance (OpenSSF score)
    if type get_maintenance_metrics &>/dev/null; then
        local metrics=$(get_maintenance_metrics "$pkg" "$ecosystem" 2>/dev/null || echo '{}')
        local openssf=$(echo "$metrics" | jq -r '.openssf_score // null')
        local maintenance_score=$(score_maintenance "$openssf")

        factors=$(echo "$factors" | jq --arg name "maintenance" --arg openssf "$openssf" --argjson score "$maintenance_score" --argjson weight "$WEIGHT_MAINTENANCE" \
            '. + [{"factor": $name, "openssf_score": $openssf, "score": $score, "weight": $weight}]')

        total_weighted_score=$((total_weighted_score + (maintenance_score * WEIGHT_MAINTENANCE)))
        total_weight=$((total_weight + WEIGHT_MAINTENANCE))
    fi

    # Calculate final score
    local final_score=0
    if [[ $total_weight -gt 0 ]]; then
        final_score=$((total_weighted_score / total_weight))
    fi

    # Determine debt level
    local debt_level="low"
    if [[ $final_score -gt 70 ]]; then
        debt_level="critical"
    elif [[ $final_score -gt 50 ]]; then
        debt_level="high"
    elif [[ $final_score -gt 30 ]]; then
        debt_level="medium"
    fi

    # Generate recommendations
    local recommendations=()
    if [[ $final_score -gt 70 ]]; then
        recommendations+=("Critical technical debt - prioritize migration")
    fi
    if [[ $final_score -gt 50 ]]; then
        recommendations+=("Plan migration to alternative")
    fi
    if [[ $final_score -gt 30 ]]; then
        recommendations+=("Monitor for issues and plan future migration")
    fi

    local recommendations_json=$(printf '%s\n' "${recommendations[@]}" 2>/dev/null | jq -R . | jq -s '.' || echo "[]")

    echo "{
        \"package\": \"$pkg\",
        \"ecosystem\": \"$ecosystem\",
        \"debt_score\": $final_score,
        \"debt_level\": \"$debt_level\",
        \"factors\": $factors,
        \"recommendations\": $recommendations_json
    }" | jq '.'
}

# Calculate debt for all dependencies in a project
# Usage: calculate_project_debt <project_dir>
calculate_project_debt() {
    local project_dir="$1"

    if [[ ! -d "$project_dir" ]]; then
        echo '{"error": "directory_not_found"}'
        return 1
    fi

    local results="[]"
    local ecosystem=""
    local packages=""

    # Detect ecosystem and get packages
    if [[ -f "$project_dir/package.json" ]]; then
        ecosystem="npm"
        packages=$(jq -r '.dependencies // {} | keys[]' "$project_dir/package.json" 2>/dev/null)
    elif [[ -f "$project_dir/requirements.txt" ]]; then
        ecosystem="python"
        packages=$(cut -d'=' -f1 "$project_dir/requirements.txt" 2>/dev/null | cut -d'>' -f1 | cut -d'<' -f1 | grep -v '^#' | grep -v '^$')
    elif [[ -f "$project_dir/go.mod" ]]; then
        ecosystem="go"
        packages=$(grep -E '^\s+[a-z]' "$project_dir/go.mod" 2>/dev/null | awk '{print $1}')
    else
        echo '{"error": "no_supported_manifest"}'
        return 1
    fi

    local total_score=0
    local count=0
    local critical=0
    local high=0
    local medium=0

    while IFS= read -r pkg; do
        [[ -z "$pkg" ]] && continue

        local debt=$(calculate_debt_score "$pkg" "$ecosystem")
        results=$(echo "$results" | jq --argjson d "$debt" '. + [$d]')

        local score=$(echo "$debt" | jq -r '.debt_score // 0')
        local level=$(echo "$debt" | jq -r '.debt_level')

        total_score=$((total_score + score))
        count=$((count + 1))

        case "$level" in
            critical) critical=$((critical + 1)) ;;
            high) high=$((high + 1)) ;;
            medium) medium=$((medium + 1)) ;;
        esac
    done <<< "$packages"

    # Calculate average debt score
    local avg_score=0
    if [[ $count -gt 0 ]]; then
        avg_score=$((total_score / count))
    fi

    # Overall project debt level
    local project_level="low"
    if [[ $critical -gt 0 || $avg_score -gt 60 ]]; then
        project_level="critical"
    elif [[ $high -gt 2 || $avg_score -gt 40 ]]; then
        project_level="high"
    elif [[ $medium -gt 5 || $avg_score -gt 25 ]]; then
        project_level="medium"
    fi

    # Sort by debt score descending
    results=$(echo "$results" | jq 'sort_by(-.debt_score)')

    # Get priority items
    local priority_items=$(echo "$results" | jq '[.[] | select(.debt_level == "critical" or .debt_level == "high")]')

    echo "{
        \"project_dir\": \"$project_dir\",
        \"ecosystem\": \"$ecosystem\",
        \"summary\": {
            \"total_packages\": $count,
            \"average_debt_score\": $avg_score,
            \"project_debt_level\": \"$project_level\",
            \"debt_breakdown\": {
                \"critical\": $critical,
                \"high\": $high,
                \"medium\": $medium,
                \"low\": $((count - critical - high - medium))
            }
        },
        \"priority_items\": $priority_items,
        \"all_packages\": $results
    }" | jq '.'
}

# Generate debt reduction roadmap
# Usage: generate_debt_roadmap <project_dir>
generate_debt_roadmap() {
    local project_dir="$1"

    local debt=$(calculate_project_debt "$project_dir")

    if echo "$debt" | jq -e '.error' >/dev/null 2>&1; then
        echo "$debt"
        return 1
    fi

    local priority_items=$(echo "$debt" | jq -r '.priority_items')
    local total_packages=$(echo "$debt" | jq -r '.summary.total_packages')
    local avg_score=$(echo "$debt" | jq -r '.summary.average_debt_score')

    # Generate phased roadmap
    local phase1="[]"  # Critical - immediate action
    local phase2="[]"  # High - short term
    local phase3="[]"  # Medium - medium term

    while IFS= read -r item; do
        [[ -z "$item" || "$item" == "null" ]] && continue
        local level=$(echo "$item" | jq -r '.debt_level')

        case "$level" in
            critical)
                phase1=$(echo "$phase1" | jq --argjson i "$item" '. + [$i]')
                ;;
            high)
                phase2=$(echo "$phase2" | jq --argjson i "$item" '. + [$i]')
                ;;
            medium)
                phase3=$(echo "$phase3" | jq --argjson i "$item" '. + [$i]')
                ;;
        esac
    done < <(echo "$debt" | jq -c '.all_packages[]')

    # Estimate effort
    local p1_count=$(echo "$phase1" | jq 'length')
    local p2_count=$(echo "$phase2" | jq 'length')
    local p3_count=$(echo "$phase3" | jq 'length')

    echo "{
        \"project_dir\": \"$project_dir\",
        \"current_state\": {
            \"total_packages\": $total_packages,
            \"average_debt_score\": $avg_score
        },
        \"roadmap\": {
            \"phase1_immediate\": {
                \"description\": \"Critical debt - address immediately\",
                \"package_count\": $p1_count,
                \"packages\": $phase1
            },
            \"phase2_short_term\": {
                \"description\": \"High debt - address within next sprint\",
                \"package_count\": $p2_count,
                \"packages\": $phase2
            },
            \"phase3_medium_term\": {
                \"description\": \"Medium debt - plan for future sprints\",
                \"package_count\": $p3_count,
                \"packages\": $phase3
            }
        },
        \"target_state\": {
            \"target_debt_score\": 20,
            \"packages_to_address\": $((p1_count + p2_count + p3_count))
        }
    }" | jq '.'
}

#############################################################################
# Export Functions
#############################################################################

export -f score_outdated
export -f score_abandonment
export -f score_replacement
export -f score_maintenance
export -f score_size
export -f calculate_debt_score
export -f calculate_project_debt
export -f generate_debt_roadmap
