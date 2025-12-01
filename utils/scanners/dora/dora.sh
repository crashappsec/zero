#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# DORA Metrics - Data Collector
# Calculates DORA metrics from git history
# Outputs JSON for agent analysis - NO AI calls
#
# Usage: ./dora-analyser-data.sh [options] <target>
# Output: JSON with DORA metrics and classifications
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

usage() {
    cat << EOF
DORA Metrics - Data Collector (JSON output for agent analysis)

Usage: $0 [OPTIONS] <target>

TARGET:
    Git repository URL      Clone and analyze repository
    Local directory path    Analyze local git repository

OPTIONS:
    --local-path PATH       Use pre-cloned repository (skips cloning)
    --days N                Analysis period in days (default: 90)
    -o, --output FILE       Write JSON to file (default: stdout)
    -k, --keep-clone        Keep cloned repository
    -h, --help              Show this help

OUTPUT:
    JSON object with:
    - metadata: scan timestamp, target, period
    - metrics: deployment_frequency, lead_time, change_failure_rate, mttr
    - classifications: ELITE, HIGH, MEDIUM, LOW for each metric

EXAMPLES:
    $0 https://github.com/expressjs/express
    $0 --local-path ~/.gibson/projects/foo/repo --days 180
    $0 -o dora.json /path/to/project

EOF
    exit 0
}

# Clone repository (full history needed for DORA)
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

# Classify deployment frequency
classify_df() {
    local df="$1"
    if (( $(echo "$df >= 1" | bc -l) )); then
        echo "ELITE"
    elif (( $(echo "$df >= 0.14" | bc -l) )); then
        echo "HIGH"
    elif (( $(echo "$df >= 0.03" | bc -l) )); then
        echo "MEDIUM"
    else
        echo "LOW"
    fi
}

# Classify lead time (hours)
classify_lt() {
    local lt="$1"
    if (( $(echo "$lt < 24" | bc -l) )); then
        echo "ELITE"
    elif (( $(echo "$lt < 168" | bc -l) )); then
        echo "HIGH"
    elif (( $(echo "$lt < 730" | bc -l) )); then
        echo "MEDIUM"
    else
        echo "LOW"
    fi
}

# Classify change failure rate (percent)
classify_cfr() {
    local cfr="$1"
    if (( $(echo "$cfr <= 15" | bc -l) )); then
        echo "ELITE"
    elif (( $(echo "$cfr <= 30" | bc -l) )); then
        echo "HIGH"
    elif (( $(echo "$cfr <= 45" | bc -l) )); then
        echo "MEDIUM"
    else
        echo "LOW"
    fi
}

# Calculate overall performance
calculate_overall() {
    local df_class="$1"
    local lt_class="$2"
    local cfr_class="$3"

    local elite=0 high=0 medium=0 low=0

    for class in "$df_class" "$lt_class" "$cfr_class"; do
        case $class in
            ELITE) ((elite++)) ;;
            HIGH) ((high++)) ;;
            MEDIUM) ((medium++)) ;;
            LOW) ((low++)) ;;
        esac
    done

    if (( elite >= 2 )); then
        echo "ELITE"
    elif (( elite + high >= 2 )); then
        echo "HIGH"
    elif (( low <= 1 )); then
        echo "MEDIUM"
    else
        echo "LOW"
    fi
}

# Main analysis
analyze_dora() {
    local repo_path="$1"
    local days="$2"

    cd "$repo_path" || { echo '{"error": "Cannot access repository"}'; exit 1; }

    # Check if it's a git repository
    if [[ ! -d ".git" ]]; then
        echo '{"error": "Not a git repository"}'
        exit 1
    fi

    echo -e "${BLUE}Analyzing DORA metrics (last $days days)...${NC}" >&2

    # Calculate date range
    local since_date=$(date -v-${days}d +%Y-%m-%d 2>/dev/null || date -d "$days days ago" +%Y-%m-%d 2>/dev/null)

    # Total commits in period
    local total_commits=$(git log --since="$since_date" --oneline 2>/dev/null | wc -l | tr -d ' ')

    if [[ "$total_commits" -eq 0 ]]; then
        echo -e "${YELLOW}⚠ No commits in last $days days, using all-time data${NC}" >&2
        since_date=""
        days=$(( ($(date +%s) - $(git log --reverse --format=%ct | head -1)) / 86400 ))
        [[ "$days" -lt 1 ]] && days=1
        total_commits=$(git log --oneline | wc -l | tr -d ' ')
    fi

    # Deployment frequency (commits per day as proxy)
    local df=$(echo "scale=3; $total_commits / $days" | bc)
    local df_class=$(classify_df "$df")

    # Lead time estimation (average time between commits in hours)
    local commit_times=""
    if [[ -n "$since_date" ]]; then
        commit_times=$(git log --since="$since_date" --format=%ct --reverse 2>/dev/null)
    else
        commit_times=$(git log --format=%ct --reverse 2>/dev/null)
    fi

    local lead_time_hours=0
    local prev_time=""
    local time_diffs=()
    while IFS= read -r ts; do
        [[ -z "$ts" ]] && continue
        if [[ -n "$prev_time" ]]; then
            local diff=$(( (ts - prev_time) / 3600 ))
            [[ $diff -gt 0 ]] && time_diffs+=("$diff")
        fi
        prev_time="$ts"
    done <<< "$commit_times"

    if [[ ${#time_diffs[@]} -gt 0 ]]; then
        local sum=0
        for d in "${time_diffs[@]}"; do
            sum=$((sum + d))
        done
        lead_time_hours=$((sum / ${#time_diffs[@]}))
    fi

    local lt_class=$(classify_lt "$lead_time_hours")

    # Change failure rate (estimate from revert commits)
    local revert_commits=0
    if [[ -n "$since_date" ]]; then
        revert_commits=$(git log --since="$since_date" --grep="revert\|fix\|hotfix" -i --oneline 2>/dev/null | wc -l | tr -d ' ')
    else
        revert_commits=$(git log --grep="revert\|fix\|hotfix" -i --oneline 2>/dev/null | wc -l | tr -d ' ')
    fi

    local cfr=0
    if [[ "$total_commits" -gt 0 ]]; then
        cfr=$(echo "scale=1; ($revert_commits / $total_commits) * 100" | bc)
    fi
    local cfr_class=$(classify_cfr "$cfr")

    # Active contributors
    local contributors=0
    if [[ -n "$since_date" ]]; then
        contributors=$(git log --since="$since_date" --format="%an" 2>/dev/null | sort -u | wc -l | tr -d ' ')
    else
        contributors=$(git log --format="%an" 2>/dev/null | sort -u | wc -l | tr -d ' ')
    fi

    # Overall classification
    local overall=$(calculate_overall "$df_class" "$lt_class" "$cfr_class")

    echo -e "${GREEN}✓ Analysis complete${NC}" >&2

    # Build output
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    jq -n \
        --arg ts "$timestamp" \
        --arg tgt "$TARGET" \
        --arg ver "1.0.0" \
        --argjson days "$days" \
        --argjson total_commits "$total_commits" \
        --argjson contributors "$contributors" \
        --arg df "$df" \
        --arg df_class "$df_class" \
        --argjson lt "$lead_time_hours" \
        --arg lt_class "$lt_class" \
        --arg cfr "$cfr" \
        --arg cfr_class "$cfr_class" \
        --argjson reverts "$revert_commits" \
        --arg overall "$overall" \
        '{
            analyzer: "dora-metrics",
            version: $ver,
            timestamp: $ts,
            target: $tgt,
            period_days: $days,
            summary: {
                overall_performance: $overall,
                total_commits: $total_commits,
                active_contributors: $contributors
            },
            metrics: {
                deployment_frequency: {
                    value: ($df | tonumber),
                    unit: "commits_per_day",
                    classification: $df_class,
                    description: "Proxy: commit frequency (higher = more frequent deployments)"
                },
                lead_time_for_changes: {
                    value: $lt,
                    unit: "hours",
                    classification: $lt_class,
                    description: "Average time between commits"
                },
                change_failure_rate: {
                    value: ($cfr | tonumber),
                    unit: "percent",
                    classification: $cfr_class,
                    revert_fix_commits: $reverts,
                    description: "Estimated from revert/fix/hotfix commits"
                }
            },
            benchmarks: {
                ELITE: "Multiple deploys/day, <1d lead time, <15% CFR",
                HIGH: "Daily-weekly deploys, 1d-1w lead time, 16-30% CFR",
                MEDIUM: "Weekly-monthly deploys, 1w-1m lead time, 31-45% CFR",
                LOW: "Monthly+ deploys, >1m lead time, >45% CFR"
            },
            note: "Metrics estimated from git history. Actual deployment data would provide more accurate results."
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

# Check prerequisites
if ! command -v bc &> /dev/null; then
    echo '{"error": "bc is required but not installed"}'
    exit 1
fi

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

final_json=$(analyze_dora "$scan_path" "$DAYS")

# Output
if [[ -n "$OUTPUT_FILE" ]]; then
    echo "$final_json" > "$OUTPUT_FILE"
    echo -e "${GREEN}✓ Results written to $OUTPUT_FILE${NC}" >&2
else
    echo "$final_json"
fi
