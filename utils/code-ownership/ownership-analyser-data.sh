#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Code Ownership - Data Collector
# Analyzes git history for code ownership patterns
# Outputs JSON for agent analysis - NO AI calls
#
# Usage: ./ownership-analyser-data.sh [options] <target>
# Output: JSON with ownership metrics, contributors, and CODEOWNERS validation
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
UTILS_ROOT="$(dirname "$SCRIPT_DIR")"
REPO_ROOT="$(dirname "$UTILS_ROOT")"

# Default options
OUTPUT_FILE=""
LOCAL_PATH=""
TEMP_DIR=""
CLEANUP=true
TARGET=""
DAYS=90
CODEOWNERS_PATH=".github/CODEOWNERS"

usage() {
    cat << EOF
Code Ownership - Data Collector (JSON output for agent analysis)

Usage: $0 [OPTIONS] <target>

TARGET:
    Git repository URL      Clone and analyze repository
    Local directory path    Analyze local git repository

OPTIONS:
    --local-path PATH       Use pre-cloned repository (skips cloning)
    --days N                Analysis period in days (default: 90)
    --codeowners PATH       Path to CODEOWNERS file (default: .github/CODEOWNERS)
    -o, --output FILE       Write JSON to file (default: stdout)
    -k, --keep-clone        Keep cloned repository
    -h, --help              Show this help

OUTPUT:
    JSON object with:
    - metadata: scan timestamp, target, period
    - contributors: list with commit counts, files touched
    - ownership: file ownership distribution
    - codeowners: CODEOWNERS file analysis if present

EXAMPLES:
    $0 https://github.com/expressjs/express
    $0 --local-path ~/.gibson/projects/foo/repo --days 180
    $0 -o ownership.json /path/to/project

EOF
    exit 0
}

# Clone repository (full history needed for ownership)
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

# Main analysis
analyze_ownership() {
    local repo_path="$1"
    local days="$2"

    cd "$repo_path" || { echo '{"error": "Cannot access repository"}'; exit 1; }

    # Check if it's a git repository
    if [[ ! -d ".git" ]]; then
        echo '{"error": "Not a git repository"}'
        exit 1
    fi

    echo -e "${BLUE}Analyzing code ownership (last $days days)...${NC}" >&2

    # Calculate date range
    local since_date=$(date -v-${days}d +%Y-%m-%d 2>/dev/null || date -d "$days days ago" +%Y-%m-%d 2>/dev/null)

    # Total files
    local total_files=$(git ls-files 2>/dev/null | wc -l | tr -d ' ')

    # Total commits in period
    local total_commits=$(git log --since="$since_date" --oneline 2>/dev/null | wc -l | tr -d ' ')

    # If no commits in period, use all-time
    if [[ "$total_commits" -eq 0 ]]; then
        echo -e "${YELLOW}⚠ No commits in last $days days, using all-time data${NC}" >&2
        since_date=""
        total_commits=$(git log --oneline | wc -l | tr -d ' ')
    fi

    # Get contributor data
    echo -e "${BLUE}Collecting contributor data...${NC}" >&2
    local contributors_json="[]"

    local contributor_data=""
    if [[ -n "$since_date" ]]; then
        contributor_data=$(git log --since="$since_date" --format="%an|%ae" | sort | uniq -c | sort -rn | head -20)
    else
        contributor_data=$(git log --format="%an|%ae" | sort | uniq -c | sort -rn | head -20)
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

        # Get lines added/deleted
        local stats=""
        if [[ -n "$since_date" ]]; then
            stats=$(git log --since="$since_date" --author="$email" --numstat --format="" 2>/dev/null | awk '{added+=$1; deleted+=$2} END {print added"|"deleted}')
        else
            stats=$(git log --author="$email" --numstat --format="" 2>/dev/null | awk '{added+=$1; deleted+=$2} END {print added"|"deleted}')
        fi
        local added=$(echo "$stats" | cut -d'|' -f1)
        local deleted=$(echo "$stats" | cut -d'|' -f2)
        [[ -z "$added" ]] && added=0
        [[ -z "$deleted" ]] && deleted=0

        contributors_json=$(echo "$contributors_json" | jq \
            --arg name "$author" \
            --arg email "$email" \
            --argjson commits "$count" \
            --argjson files "$files_touched" \
            --argjson added "$added" \
            --argjson deleted "$deleted" \
            '. + [{"name": $name, "email": $email, "commits": $commits, "files_touched": $files, "lines_added": $added, "lines_deleted": $deleted}]')
    done <<< "$contributor_data"

    local contributor_count=$(echo "$contributors_json" | jq 'length')
    echo -e "${GREEN}✓ Found $contributor_count contributors${NC}" >&2

    # CODEOWNERS analysis
    echo -e "${BLUE}Checking CODEOWNERS...${NC}" >&2
    local codeowners_json='{"exists": false}'

    if [[ -f "$repo_path/$CODEOWNERS_PATH" ]]; then
        local patterns=$(grep -v "^#" "$repo_path/$CODEOWNERS_PATH" | grep -v "^$" | wc -l | tr -d ' ')
        local owners_list=$(grep -v "^#" "$repo_path/$CODEOWNERS_PATH" | grep -v "^$" | awk '{for(i=2;i<=NF;i++) print $i}' | sort -u)
        local unique_owners=$(echo "$owners_list" | wc -l | tr -d ' ')

        # Parse patterns
        local patterns_json="[]"
        while IFS= read -r line; do
            [[ "$line" =~ ^#.*$ ]] && continue
            [[ -z "$line" ]] && continue

            local pattern=$(echo "$line" | awk '{print $1}')
            local owners=$(echo "$line" | cut -d' ' -f2- | tr ' ' ',')

            patterns_json=$(echo "$patterns_json" | jq \
                --arg pattern "$pattern" \
                --arg owners "$owners" \
                '. + [{"pattern": $pattern, "owners": $owners}]')
        done < "$repo_path/$CODEOWNERS_PATH"

        codeowners_json=$(jq -n \
            --arg path "$CODEOWNERS_PATH" \
            --argjson patterns "$patterns" \
            --argjson unique_owners "$unique_owners" \
            --argjson pattern_list "$patterns_json" \
            '{
                exists: true,
                path: $path,
                total_patterns: $patterns,
                unique_owners: $unique_owners,
                patterns: $pattern_list
            }')

        echo -e "${GREEN}✓ CODEOWNERS found: $patterns patterns, $unique_owners owners${NC}" >&2
    else
        echo -e "${YELLOW}⚠ No CODEOWNERS file found${NC}" >&2
    fi

    # Calculate bus factor (rough estimate)
    # Bus factor = minimum contributors who own 50% of code
    local top_contributor_commits=0
    local total_period_commits=$total_commits
    local bus_factor=1

    if [[ "$contributor_count" -gt 0 ]] && [[ "$total_period_commits" -gt 0 ]]; then
        local cumulative=0
        local threshold=$(( total_period_commits / 2 ))

        while read contributor; do
            local commits=$(echo "$contributor" | jq -r '.commits')
            cumulative=$((cumulative + commits))
            if [[ $cumulative -ge $threshold ]]; then
                break
            fi
            ((bus_factor++))
        done <<< "$(echo "$contributors_json" | jq -c '.[]')"
    fi

    # Build output
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    jq -n \
        --arg ts "$timestamp" \
        --arg tgt "$TARGET" \
        --arg ver "1.0.0" \
        --argjson days "$days" \
        --argjson total_files "$total_files" \
        --argjson total_commits "$total_commits" \
        --argjson contributor_count "$contributor_count" \
        --argjson bus_factor "$bus_factor" \
        --argjson contributors "$contributors_json" \
        --argjson codeowners "$codeowners_json" \
        '{
            analyzer: "code-ownership",
            version: $ver,
            timestamp: $ts,
            target: $tgt,
            period_days: $days,
            summary: {
                total_files: $total_files,
                total_commits: $total_commits,
                active_contributors: $contributor_count,
                estimated_bus_factor: $bus_factor
            },
            contributors: $contributors,
            codeowners: $codeowners,
            risk_assessment: {
                bus_factor_risk: (if $bus_factor <= 1 then "critical" elif $bus_factor <= 2 then "high" elif $bus_factor <= 3 then "medium" else "low" end),
                bus_factor_description: "Minimum contributors owning 50% of commits"
            }
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
        --codeowners)
            CODEOWNERS_PATH="$2"
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

final_json=$(analyze_ownership "$scan_path" "$DAYS")

# Output
if [[ -n "$OUTPUT_FILE" ]]; then
    echo "$final_json" > "$OUTPUT_FILE"
    echo -e "${GREEN}✓ Results written to $OUTPUT_FILE${NC}" >&2
else
    echo "$final_json"
fi
