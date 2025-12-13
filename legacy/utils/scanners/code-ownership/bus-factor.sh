#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Bus Factor - Data Collector
# Analyzes git history to calculate bus factor risk metrics
# Outputs JSON for agent analysis - NO AI calls
#
# Usage: ./bus-factor.sh [options] <target>
# Output: JSON with bus factor metrics and contributor concentration
#############################################################################

set -e

# Colors for terminal output (stderr only)
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SCANNERS_ROOT="$(dirname "$SCRIPT_DIR")"
UTILS_ROOT="$(dirname "$SCANNERS_ROOT")"
REPO_ROOT="$(dirname "$UTILS_ROOT")"

# Default options
OUTPUT_FILE=""
LOCAL_PATH=""
TEMP_DIR=""
CLEANUP=true
TARGET=""
DAYS=90
THRESHOLD=50  # Percentage threshold for bus factor calculation

usage() {
    cat << EOF
Bus Factor - Data Collector (JSON output for agent analysis)

Usage: $0 [OPTIONS] <target>

TARGET:
    Git repository URL      Clone and analyze repository
    Local directory path    Analyze local git repository

OPTIONS:
    --local-path PATH       Use pre-cloned repository (skips cloning)
    --days N                Analysis period in days (default: 90)
    --threshold N           Percentage threshold for bus factor (default: 50)
    -o, --output FILE       Write JSON to file (default: stdout)
    -k, --keep-clone        Keep cloned repository
    -h, --help              Show this help

OUTPUT:
    JSON object with:
    - metadata: scan timestamp, target, period
    - bus_factor: minimum contributors owning threshold% of commits
    - contributors: list with commit counts and ownership percentages
    - concentration: code ownership concentration metrics

EXAMPLES:
    $0 https://github.com/expressjs/express
    $0 --local-path ~/.gibson/projects/foo/repo --days 180
    $0 -o bus-factor.json /path/to/project

EOF
    exit 0
}

# Clone repository (full history needed for bus factor analysis)
clone_repository() {
    local repo_url="$1"
    TEMP_DIR=$(mktemp -d)

    echo -e "${BLUE}Cloning repository (full history)...${NC}" >&2
    if git clone "$repo_url" "$TEMP_DIR" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Cloned${NC}" >&2
        return 0
    else
        echo '{"error": "Failed to clone repository"}'
        exit 1
    fi
}

# Cleanup
cleanup() {
    if [[ "$CLEANUP" == true ]] && [[ -n "$TEMP_DIR" ]] && [[ -d "$TEMP_DIR" ]]; then
        rm -rf "$TEMP_DIR"
    fi
}
trap cleanup EXIT

# Detect if target is a Git URL
is_git_url() {
    [[ "$1" =~ ^(https?|git)://.*\.git$ ]] || [[ "$1" =~ ^git@.*:.*\.git$ ]] || [[ "$1" =~ github\.com|gitlab\.com|bitbucket\.org ]]
}

# Calculate Gini coefficient for ownership concentration
calculate_gini() {
    local commits_array="$1"

    # Parse commits into sorted array
    local sorted_commits=$(echo "$commits_array" | jq -r '.[] | .commits' | sort -n)
    local n=$(echo "$sorted_commits" | wc -l | tr -d ' ')

    if [[ $n -eq 0 ]]; then
        echo "0"
        return
    fi

    local sum=0
    local weighted_sum=0
    local i=1

    while read commits; do
        sum=$((sum + commits))
        weighted_sum=$((weighted_sum + i * commits))
        ((i++))
    done <<< "$sorted_commits"

    if [[ $sum -eq 0 ]]; then
        echo "0"
        return
    fi

    # Gini = (2 * sum(i * x_i)) / (n * sum(x_i)) - (n + 1) / n
    # Using awk for floating point
    echo "$weighted_sum $n $sum" | awk '{
        gini = (2 * $1) / ($2 * $3) - ($2 + 1) / $2
        printf "%.3f", gini
    }'
}

# Main analysis
analyze_bus_factor() {
    local repo_path="$1"
    local days="$2"
    local threshold="$3"

    cd "$repo_path" || { echo '{"error": "Cannot access repository"}'; exit 1; }

    # Check if it's a git repository
    if [[ ! -d ".git" ]]; then
        echo '{"error": "Not a git repository"}'
        exit 1
    fi

    echo -e "${BLUE}Analyzing bus factor (last $days days)...${NC}" >&2

    # Calculate date range
    local since_date=$(date -v-${days}d +%Y-%m-%d 2>/dev/null || date -d "$days days ago" +%Y-%m-%d 2>/dev/null)

    # Total commits in period
    local total_commits=$(git log --since="$since_date" --oneline 2>/dev/null | wc -l | tr -d ' ')

    # If no commits in period, use all-time
    local period_description="last $days days"
    if [[ "$total_commits" -eq 0 ]]; then
        echo -e "${YELLOW}⚠ No commits in last $days days, using all-time data${NC}" >&2
        since_date=""
        total_commits=$(git log --oneline | wc -l | tr -d ' ')
        period_description="all time"
    fi

    # Get contributor data
    echo -e "${BLUE}Collecting contributor data...${NC}" >&2
    local contributors_json="[]"

    local contributor_data=""
    if [[ -n "$since_date" ]]; then
        contributor_data=$(git log --since="$since_date" --format="%an|%ae" | sort | uniq -c | sort -rn)
    else
        contributor_data=$(git log --format="%an|%ae" | sort | uniq -c | sort -rn)
    fi

    while read count author_email; do
        [[ -z "$count" ]] && continue
        local author=$(echo "$author_email" | cut -d'|' -f1)
        local email=$(echo "$author_email" | cut -d'|' -f2)

        # Get files touched by this author
        local files_touched=0
        if [[ -n "$since_date" ]]; then
            files_touched=$(git log --since="$since_date" --author="$email" --name-only --format="" 2>/dev/null | sort -u | wc -l | tr -d ' ')
        else
            files_touched=$(git log --author="$email" --name-only --format="" 2>/dev/null | sort -u | wc -l | tr -d ' ')
        fi

        # Calculate ownership percentage
        local ownership_pct=0
        if [[ $total_commits -gt 0 ]]; then
            ownership_pct=$(echo "$count $total_commits" | awk '{printf "%.2f", ($1 / $2) * 100}')
        fi

        contributors_json=$(echo "$contributors_json" | jq \
            --arg name "$author" \
            --arg email "$email" \
            --argjson commits "$count" \
            --argjson files "$files_touched" \
            --arg ownership "$ownership_pct" \
            '. + [{"name": $name, "email": $email, "commits": $commits, "files_touched": $files, "ownership_percentage": ($ownership | tonumber)}]')
    done <<< "$contributor_data"

    local contributor_count=$(echo "$contributors_json" | jq 'length')
    echo -e "${GREEN}✓ Found $contributor_count contributors${NC}" >&2

    # Calculate bus factor (minimum contributors who own threshold% of commits)
    local bus_factor=0
    local cumulative=0
    local target_commits=$(echo "$total_commits $threshold" | awk '{printf "%.0f", $1 * $2 / 100}')
    local bus_factor_contributors="[]"

    if [[ "$contributor_count" -gt 0 ]] && [[ "$total_commits" -gt 0 ]]; then
        while read contributor; do
            local commits=$(echo "$contributor" | jq -r '.commits')
            local name=$(echo "$contributor" | jq -r '.name')
            cumulative=$((cumulative + commits))
            ((bus_factor++))

            bus_factor_contributors=$(echo "$bus_factor_contributors" | jq \
                --arg name "$name" \
                --argjson commits "$commits" \
                '. + [{"name": $name, "commits": $commits}]')

            if [[ $cumulative -ge $target_commits ]]; then
                break
            fi
        done <<< "$(echo "$contributors_json" | jq -c '.[]')"
    fi

    echo -e "${BLUE}Bus factor: $bus_factor (${threshold}% threshold)${NC}" >&2

    # Calculate concentration metrics
    local gini_coefficient=$(calculate_gini "$contributors_json")

    # Top contributor percentage
    local top_contributor_pct=0
    if [[ "$contributor_count" -gt 0 ]]; then
        top_contributor_pct=$(echo "$contributors_json" | jq '.[0].ownership_percentage')
    fi

    # Top 3 contributors percentage
    local top3_pct=0
    if [[ "$contributor_count" -gt 0 ]]; then
        top3_pct=$(echo "$contributors_json" | jq '[.[:3][].ownership_percentage] | add')
    fi

    # Determine risk level
    local risk_level="low"
    local risk_description=""
    if [[ $bus_factor -le 1 ]]; then
        risk_level="critical"
        risk_description="Single point of failure - project depends on one contributor"
    elif [[ $bus_factor -le 2 ]]; then
        risk_level="high"
        risk_description="Very few contributors own majority of code"
    elif [[ $bus_factor -le 3 ]]; then
        risk_level="medium"
        risk_description="Limited contributor diversity for code ownership"
    else
        risk_level="low"
        risk_description="Healthy distribution of code ownership"
    fi

    # Build output
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    jq -n \
        --arg ts "$timestamp" \
        --arg tgt "$TARGET" \
        --arg ver "1.0.0" \
        --arg period "$period_description" \
        --argjson days "$days" \
        --argjson threshold "$threshold" \
        --argjson total_commits "$total_commits" \
        --argjson contributor_count "$contributor_count" \
        --argjson bus_factor "$bus_factor" \
        --argjson bus_factor_contributors "$bus_factor_contributors" \
        --argjson contributors "$contributors_json" \
        --arg gini "$gini_coefficient" \
        --argjson top_contributor_pct "$top_contributor_pct" \
        --argjson top3_pct "$top3_pct" \
        --arg risk_level "$risk_level" \
        --arg risk_description "$risk_description" \
        '{
            analyzer: "bus-factor",
            version: $ver,
            timestamp: $ts,
            target: $tgt,
            period_days: $days,
            period_description: $period,
            threshold_percentage: $threshold,
            summary: {
                total_commits: $total_commits,
                active_contributors: $contributor_count,
                bus_factor: $bus_factor,
                risk_level: $risk_level
            },
            bus_factor_analysis: {
                bus_factor: $bus_factor,
                threshold_percentage: $threshold,
                contributors_for_threshold: $bus_factor_contributors,
                risk_level: $risk_level,
                risk_description: $risk_description
            },
            concentration_metrics: {
                gini_coefficient: ($gini | tonumber),
                top_contributor_percentage: $top_contributor_pct,
                top_3_contributors_percentage: $top3_pct,
                interpretation: {
                    gini: (if ($gini | tonumber) > 0.6 then "High concentration - few contributors dominate" elif ($gini | tonumber) > 0.4 then "Moderate concentration" else "Well distributed" end)
                }
            },
            contributors: $contributors
        }'
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help) usage ;;
        --local-path)
            LOCAL_PATH="$2"
            shift 2
            ;;
        --days)
            DAYS="$2"
            shift 2
            ;;
        --threshold)
            THRESHOLD="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_FILE="$2"
            shift 2
            ;;
        -k|--keep-clone)
            CLEANUP=false
            shift
            ;;
        -*)
            echo "Unknown option: $1" >&2
            exit 1
            ;;
        *)
            TARGET="$1"
            shift
            ;;
    esac
done

# Main execution
scan_path=""

if [[ -n "$LOCAL_PATH" ]]; then
    [[ ! -d "$LOCAL_PATH" ]] && { echo '{"error": "Local path does not exist"}'; exit 1; }
    scan_path="$LOCAL_PATH"
    TARGET="$LOCAL_PATH"
elif [[ -n "$TARGET" ]]; then
    if is_git_url "$TARGET"; then
        clone_repository "$TARGET"
        scan_path="$TEMP_DIR"
    elif [[ -d "$TARGET" ]]; then
        scan_path="$TARGET"
    else
        echo '{"error": "Invalid target - must be URL or directory"}'
        exit 1
    fi
else
    echo '{"error": "No target specified"}'
    exit 1
fi

echo -e "${BLUE}Analyzing: $TARGET${NC}" >&2

final_json=$(analyze_bus_factor "$scan_path" "$DAYS" "$THRESHOLD")

# Output
if [[ -n "$OUTPUT_FILE" ]]; then
    echo "$final_json" > "$OUTPUT_FILE"
    echo -e "${GREEN}✓ Results written to $OUTPUT_FILE${NC}" >&2
else
    echo "$final_json"
fi
