#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Git Insights - Data Collector
# Analyzes git history for contributor patterns, churn, and activity metrics
# Outputs JSON for agent analysis - NO AI calls
#
# Usage: ./git-insights-data.sh [options] <target>
# Output: JSON with contributor stats, churn analysis, and patterns
#############################################################################

set -e

# Colors for terminal output (stderr only)
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SCANNERS_ROOT="$(dirname "$SCRIPT_DIR")"
UTILS_ROOT="$(dirname "$SCANNERS_ROOT")"
REPO_ROOT="$(dirname "$UTILS_ROOT")"

# Default options
OUTPUT_FILE=""
LOCAL_PATH=""
REPO=""
ORG=""
TEMP_DIR=""
CLEANUP=true
TARGET=""
DAYS_30=$(date -v-30d +%Y-%m-%d 2>/dev/null || date -d "30 days ago" +%Y-%m-%d 2>/dev/null || echo "")
DAYS_90=$(date -v-90d +%Y-%m-%d 2>/dev/null || date -d "90 days ago" +%Y-%m-%d 2>/dev/null || echo "")
DAYS_365=$(date -v-365d +%Y-%m-%d 2>/dev/null || date -d "365 days ago" +%Y-%m-%d 2>/dev/null || echo "")

usage() {
    cat << EOF
Git Insights - Data Collector (JSON output for agent analysis)

Usage: $0 [OPTIONS] <target>

TARGET:
    Git repository URL      Clone and analyze repository
    Local directory path    Analyze local repository

OPTIONS:
    --local-path PATH       Use pre-cloned repository (skips cloning)
    --repo OWNER/REPO       GitHub repository (looks in zero cache)
    --org ORG               GitHub org (uses first repo found in zero cache)
    -o, --output FILE       Write JSON to file (default: stdout)
    -k, --keep-clone        Keep cloned repository
    -h, --help              Show this help

OUTPUT:
    JSON object with:
    - summary: total commits, active contributors, bus factor
    - contributors: activity by contributor
    - high_churn_files: frequently modified files
    - code_age: distribution of code age
    - patterns: commit timing and size patterns

EXAMPLES:
    $0 https://github.com/expressjs/express
    $0 --local-path ~/.zero/projects/foo/repo
    $0 -o git-insights.json /path/to/project

EOF
    exit 0
}

# Clone repository
clone_repository() {
    local repo_url="$1"
    TEMP_DIR=$(mktemp -d)

    echo -e "${BLUE}Cloning repository...${NC}" >&2
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

# Get total commit count
get_total_commits() {
    local repo_dir="$1"
    cd "$repo_dir" && git rev-list --count HEAD 2>/dev/null || echo "0"
}

# Get commits in date range
get_commits_since() {
    local repo_dir="$1"
    local since_date="$2"
    cd "$repo_dir" && git rev-list --count HEAD --since="$since_date" 2>/dev/null || echo "0"
}

# Get contributor stats
get_contributors() {
    local repo_dir="$1"
    local contributors="[]"

    # Get all contributors with commit counts
    local author_stats=$(cd "$repo_dir" && git shortlog -sne HEAD 2>/dev/null)

    while IFS=$'\t' read -r count author_info; do
        [[ -z "$count" ]] && continue

        # Parse name and email
        local name=$(echo "$author_info" | sed 's/ <.*$//')
        local email=$(echo "$author_info" | sed 's/.*<\(.*\)>/\1/')

        # Get stats for different time windows
        local commits_30d=0
        local commits_90d=0
        local commits_365d=0

        if [[ -n "$DAYS_30" ]]; then
            commits_30d=$(cd "$repo_dir" && git rev-list --count HEAD --author="$email" --since="$DAYS_30" 2>/dev/null || echo "0")
        fi
        if [[ -n "$DAYS_90" ]]; then
            commits_90d=$(cd "$repo_dir" && git rev-list --count HEAD --author="$email" --since="$DAYS_90" 2>/dev/null || echo "0")
        fi
        if [[ -n "$DAYS_365" ]]; then
            commits_365d=$(cd "$repo_dir" && git rev-list --count HEAD --author="$email" --since="$DAYS_365" 2>/dev/null || echo "0")
        fi

        # Get lines added/removed in last 90 days (simplified)
        local lines_added=0
        local lines_removed=0
        if [[ -n "$DAYS_90" ]] && [[ "$commits_90d" -gt 0 ]]; then
            local stats=$(cd "$repo_dir" && git log --author="$email" --since="$DAYS_90" --pretty=tformat: --numstat 2>/dev/null | awk '{add+=$1; del+=$2} END {print add, del}')
            lines_added=$(echo "$stats" | awk '{print $1}')
            lines_removed=$(echo "$stats" | awk '{print $2}')
            [[ -z "$lines_added" ]] && lines_added=0
            [[ -z "$lines_removed" ]] && lines_removed=0
        fi

        contributors=$(echo "$contributors" | jq \
            --arg name "$name" \
            --arg email "$email" \
            --argjson total "$count" \
            --argjson commits_30d "$commits_30d" \
            --argjson commits_90d "$commits_90d" \
            --argjson commits_365d "$commits_365d" \
            --argjson lines_added "$lines_added" \
            --argjson lines_removed "$lines_removed" \
            '. + [{
                "name": $name,
                "email": $email,
                "total_commits": $total,
                "commits_30d": $commits_30d,
                "commits_90d": $commits_90d,
                "commits_365d": $commits_365d,
                "lines_added_90d": $lines_added,
                "lines_removed_90d": $lines_removed
            }]')
    done <<< "$author_stats"

    # Sort by total commits descending
    echo "$contributors" | jq 'sort_by(-.total_commits)'
}

# Calculate bus factor
calculate_bus_factor() {
    local contributors="$1"

    # Bus factor: minimum contributors needed for 50% of commits
    local total_commits=$(echo "$contributors" | jq '[.[].total_commits] | add // 0')
    local threshold=$((total_commits / 2))

    local cumulative=0
    local bus_factor=0

    while read -r commits; do
        [[ -z "$commits" ]] && continue
        cumulative=$((cumulative + commits))
        ((bus_factor++))
        if [[ "$cumulative" -ge "$threshold" ]]; then
            break
        fi
    done < <(echo "$contributors" | jq -r '.[].total_commits')

    echo "$bus_factor"
}

# Get high churn files
get_high_churn_files() {
    local repo_dir="$1"
    local since_date="$2"
    local churn_files="[]"

    if [[ -z "$since_date" ]]; then
        echo "$churn_files"
        return
    fi

    # Get files with most changes
    local file_changes=$(cd "$repo_dir" && git log --since="$since_date" --pretty=format: --name-only 2>/dev/null | \
        grep -v '^$' | \
        grep -v 'node_modules\|vendor\|\.git\|dist\|build' | \
        sort | uniq -c | sort -rn | head -30)

    while read -r count file; do
        [[ -z "$count" ]] && continue
        [[ "$count" -lt 5 ]] && continue  # Skip files with fewer than 5 changes

        # Get unique contributors to this file
        local file_contributors=$(cd "$repo_dir" && git log --since="$since_date" --format='%ae' -- "$file" 2>/dev/null | sort -u | wc -l | tr -d ' ')

        churn_files=$(echo "$churn_files" | jq \
            --arg file "$file" \
            --argjson changes "$count" \
            --argjson contributors "$file_contributors" \
            '. + [{
                "file": $file,
                "changes_90d": $changes,
                "contributors": $contributors
            }]')
    done <<< "$file_changes"

    echo "$churn_files"
}

# Calculate code age distribution
calculate_code_age() {
    local repo_dir="$1"
    local code_age='{}'

    local total_files=0
    local age_0_30=0
    local age_31_90=0
    local age_91_365=0
    local age_365_plus=0

    # Sample files for age calculation (limit for performance)
    local files=$(cd "$repo_dir" && find . -type f \( \
        -name "*.py" -o -name "*.js" -o -name "*.ts" -o -name "*.java" -o -name "*.go" \
        -o -name "*.rb" -o -name "*.php" -o -name "*.c" -o -name "*.cpp" \
    \) ! -path "*node_modules*" ! -path "*vendor*" ! -path "*.git*" 2>/dev/null | head -200)

    local now=$(date +%s)
    local days_30_ago=$((now - 30*86400))
    local days_90_ago=$((now - 90*86400))
    local days_365_ago=$((now - 365*86400))

    while IFS= read -r file; do
        [[ -z "$file" ]] && continue
        ((total_files++))

        # Get last modification date from git
        local last_mod=$(cd "$repo_dir" && git log -1 --format="%at" -- "$file" 2>/dev/null || echo "0")
        [[ "$last_mod" == "0" ]] && continue

        if [[ "$last_mod" -ge "$days_30_ago" ]]; then
            ((age_0_30++))
        elif [[ "$last_mod" -ge "$days_90_ago" ]]; then
            ((age_31_90++))
        elif [[ "$last_mod" -ge "$days_365_ago" ]]; then
            ((age_91_365++))
        else
            ((age_365_plus++))
        fi
    done <<< "$files"

    # Calculate percentages
    if [[ "$total_files" -gt 0 ]]; then
        local pct_0_30=$(echo "scale=2; $age_0_30 * 100 / $total_files" | bc)
        local pct_31_90=$(echo "scale=2; $age_31_90 * 100 / $total_files" | bc)
        local pct_91_365=$(echo "scale=2; $age_91_365 * 100 / $total_files" | bc)
        local pct_365_plus=$(echo "scale=2; $age_365_plus * 100 / $total_files" | bc)

        jq -n \
            --argjson total "$total_files" \
            --argjson age_0_30 "$age_0_30" \
            --argjson age_31_90 "$age_31_90" \
            --argjson age_91_365 "$age_91_365" \
            --argjson age_365_plus "$age_365_plus" \
            --arg pct_0_30 "$pct_0_30" \
            --arg pct_31_90 "$pct_31_90" \
            --arg pct_91_365 "$pct_91_365" \
            --arg pct_365_plus "$pct_365_plus" \
            '{
                "sampled_files": $total,
                "0_30_days": {"count": $age_0_30, "percentage": ($pct_0_30 | tonumber)},
                "31_90_days": {"count": $age_31_90, "percentage": ($pct_31_90 | tonumber)},
                "91_365_days": {"count": $age_91_365, "percentage": ($pct_91_365 | tonumber)},
                "365_plus_days": {"count": $age_365_plus, "percentage": ($pct_365_plus | tonumber)}
            }'
    else
        echo '{"sampled_files": 0}'
    fi
}

# Analyze commit patterns
analyze_commit_patterns() {
    local repo_dir="$1"
    local patterns='{}'

    # Get day of week distribution
    local dow_stats=$(cd "$repo_dir" && git log --format='%ad' --date=format:'%A' 2>/dev/null | sort | uniq -c | sort -rn)

    local most_active_day=$(echo "$dow_stats" | head -1 | awk '{print $2}')
    patterns=$(echo "$patterns" | jq --arg day "$most_active_day" '.most_active_day = $day')

    # Get hour distribution
    local hour_stats=$(cd "$repo_dir" && git log --format='%ad' --date=format:'%H' 2>/dev/null | sort | uniq -c | sort -rn)
    local most_active_hour=$(echo "$hour_stats" | head -1 | awk '{print $2}' | sed 's/^0//')
    patterns=$(echo "$patterns" | jq --argjson hour "$most_active_hour" '.most_active_hour = $hour')

    # Calculate average commit size (lines changed)
    local avg_size=0
    if [[ -n "$DAYS_90" ]]; then
        local total_changes=$(cd "$repo_dir" && git log --since="$DAYS_90" --pretty=tformat: --shortstat 2>/dev/null | \
            awk '/files changed/ {total += $4 + $6} END {print total}')
        local commit_count=$(cd "$repo_dir" && git rev-list --count HEAD --since="$DAYS_90" 2>/dev/null || echo "1")
        if [[ "$commit_count" -gt 0 ]] && [[ -n "$total_changes" ]]; then
            avg_size=$((total_changes / commit_count))
        fi
    fi
    patterns=$(echo "$patterns" | jq --argjson size "$avg_size" '.avg_commit_size_lines = $size')

    # Get first and last commit dates
    local first_commit=$(cd "$repo_dir" && git log --reverse --format='%aI' 2>/dev/null | head -1)
    local last_commit=$(cd "$repo_dir" && git log -1 --format='%aI' 2>/dev/null)
    patterns=$(echo "$patterns" | jq --arg first "$first_commit" --arg last "$last_commit" \
        '.first_commit = $first | .last_commit = $last')

    # Calculate commits per week (last 90 days)
    local commits_90d=$(cd "$repo_dir" && git rev-list --count HEAD --since="$DAYS_90" 2>/dev/null || echo "0")
    local weeks=13
    local commits_per_week=$((commits_90d / weeks))
    patterns=$(echo "$patterns" | jq --argjson cpw "$commits_per_week" '.avg_commits_per_week = $cpw')

    echo "$patterns"
}

# Get branch info
get_branch_info() {
    local repo_dir="$1"
    local branch_info='{}'

    # Current branch
    local current_branch=$(cd "$repo_dir" && git branch --show-current 2>/dev/null || echo "unknown")
    branch_info=$(echo "$branch_info" | jq --arg branch "$current_branch" '.current = $branch')

    # Total branches
    local total_branches=$(cd "$repo_dir" && git branch -a 2>/dev/null | wc -l | tr -d ' ')
    branch_info=$(echo "$branch_info" | jq --argjson total "$total_branches" '.total_count = $total')

    # Remote branches
    local remote_branches=$(cd "$repo_dir" && git branch -r 2>/dev/null | wc -l | tr -d ' ')
    branch_info=$(echo "$branch_info" | jq --argjson remote "$remote_branches" '.remote_count = $remote')

    # Default branch
    local default_branch=$(cd "$repo_dir" && git symbolic-ref refs/remotes/origin/HEAD 2>/dev/null | sed 's@^refs/remotes/origin/@@' || echo "main")
    branch_info=$(echo "$branch_info" | jq --arg default "$default_branch" '.default = $default')

    echo "$branch_info"
}

# Main analysis
analyze_target() {
    local repo_dir="$1"

    # Check if it's a git repo
    if [[ ! -d "$repo_dir/.git" ]]; then
        echo '{"error": "Not a git repository"}'
        exit 1
    fi

    echo -e "${BLUE}Counting total commits...${NC}" >&2
    local total_commits=$(get_total_commits "$repo_dir")
    echo -e "${GREEN}✓ $total_commits total commits${NC}" >&2

    echo -e "${BLUE}Analyzing contributors...${NC}" >&2
    local contributors=$(get_contributors "$repo_dir")
    local total_contributors=$(echo "$contributors" | jq 'length')
    local active_30d=$(echo "$contributors" | jq '[.[] | select(.commits_30d > 0)] | length')
    local active_90d=$(echo "$contributors" | jq '[.[] | select(.commits_90d > 0)] | length')
    echo -e "${GREEN}✓ $total_contributors contributors ($active_90d active in 90d)${NC}" >&2

    echo -e "${BLUE}Calculating bus factor...${NC}" >&2
    local bus_factor=$(calculate_bus_factor "$contributors")
    echo -e "${GREEN}✓ Bus factor: $bus_factor${NC}" >&2

    echo -e "${BLUE}Finding high churn files...${NC}" >&2
    local churn_files=$(get_high_churn_files "$repo_dir" "$DAYS_90")
    local churn_count=$(echo "$churn_files" | jq 'length')
    echo -e "${GREEN}✓ $churn_count high-churn files${NC}" >&2

    echo -e "${BLUE}Calculating code age distribution...${NC}" >&2
    local code_age=$(calculate_code_age "$repo_dir")
    echo -e "${GREEN}✓ Code age calculated${NC}" >&2

    echo -e "${BLUE}Analyzing commit patterns...${NC}" >&2
    local patterns=$(analyze_commit_patterns "$repo_dir")
    echo -e "${GREEN}✓ Patterns analyzed${NC}" >&2

    echo -e "${BLUE}Getting branch info...${NC}" >&2
    local branch_info=$(get_branch_info "$repo_dir")
    echo -e "${GREEN}✓ Branch info collected${NC}" >&2

    # Calculate activity score
    local activity_score="low"
    local commits_90d=$(cd "$repo_dir" && git rev-list --count HEAD --since="$DAYS_90" 2>/dev/null || echo "0")
    if [[ "$commits_90d" -gt 500 ]]; then
        activity_score="very_high"
    elif [[ "$commits_90d" -gt 200 ]]; then
        activity_score="high"
    elif [[ "$commits_90d" -gt 50 ]]; then
        activity_score="medium"
    fi

    echo -e "${CYAN}Activity: $activity_score ($commits_90d commits in 90d)${NC}" >&2

    # Build final output
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    jq -n \
        --arg ts "$timestamp" \
        --arg tgt "$TARGET" \
        --arg ver "1.0.0" \
        --argjson total_commits "$total_commits" \
        --argjson total_contributors "$total_contributors" \
        --argjson active_30d "$active_30d" \
        --argjson active_90d "$active_90d" \
        --argjson bus_factor "$bus_factor" \
        --argjson commits_90d "$commits_90d" \
        --arg activity "$activity_score" \
        --argjson contributors "$contributors" \
        --argjson churn_files "$churn_files" \
        --argjson code_age "$code_age" \
        --argjson patterns "$patterns" \
        --argjson branches "$branch_info" \
        '{
            analyzer: "git-insights",
            version: $ver,
            timestamp: $ts,
            target: $tgt,
            summary: {
                total_commits: $total_commits,
                total_contributors: $total_contributors,
                active_contributors_30d: $active_30d,
                active_contributors_90d: $active_90d,
                commits_90d: $commits_90d,
                bus_factor: $bus_factor,
                activity_level: $activity
            },
            contributors: $contributors,
            high_churn_files: $churn_files,
            code_age: $code_age,
            patterns: $patterns,
            branches: $branches
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
        --repo)
            REPO="$2"
            shift 2
            ;;
        --org)
            ORG="$2"
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
elif [[ -n "$REPO" ]]; then
    # Look in zero cache
    REPO_ORG=$(echo "$REPO" | cut -d'/' -f1)
    REPO_NAME=$(echo "$REPO" | cut -d'/' -f2)
    ZERO_CACHE_PATH="$HOME/.zero/projects/$REPO_ORG/$REPO_NAME/repo"
    LEGACY_PATH="$HOME/.zero/projects/${REPO_ORG}-${REPO_NAME}/repo"

    if [[ -d "$ZERO_CACHE_PATH" ]]; then
        scan_path="$ZERO_CACHE_PATH"
        TARGET="$REPO"
    elif [[ -d "$LEGACY_PATH" ]]; then
        scan_path="$LEGACY_PATH"
        TARGET="$REPO"
    else
        echo '{"error": "Repository not found in cache. Clone it first or use --local-path"}'
        exit 1
    fi
elif [[ -n "$ORG" ]]; then
    # Scan ALL repos in the org
    ORG_PATH="$HOME/.zero/projects/$ORG"
    if [[ -d "$ORG_PATH" ]]; then
        # Collect repos with and without cloned code
        REPOS_TO_SCAN=()
        REPOS_NOT_CLONED=()
        for repo_dir in "$ORG_PATH"/*/; do
            repo_name=$(basename "$repo_dir")
            if [[ -d "$repo_dir/repo" ]]; then
                REPOS_TO_SCAN+=("$repo_name")
            else
                REPOS_NOT_CLONED+=("$repo_name")
            fi
        done

        # Check if there are uncloned repos and prompt user
        if [[ ${#REPOS_NOT_CLONED[@]} -gt 0 ]]; then
            echo -e "${YELLOW}Found ${#REPOS_NOT_CLONED[@]} repositories without cloned code:${NC}" >&2
            for repo in "${REPOS_NOT_CLONED[@]}"; do
                echo -e "  - $repo" >&2
            done
            echo "" >&2

            # Only prompt if interactive terminal
            if [[ -t 0 ]]; then
                read -p "Would you like to hydrate these repos for analysis? [y/N] " -n 1 -r >&2
                echo "" >&2
            else
                echo -e "${CYAN}Non-interactive mode: skipping uncloned repos${NC}" >&2
                REPLY="n"
            fi

            if [[ $REPLY =~ ^[Yy]$ ]]; then
                echo -e "${BLUE}Hydrating ${#REPOS_NOT_CLONED[@]} repositories...${NC}" >&2
                for repo in "${REPOS_NOT_CLONED[@]}"; do
                    echo -e "${CYAN}Cloning $ORG/$repo...${NC}" >&2
                    "$REPO_ROOT/utils/zero/hydrate.sh" --repo "$ORG/$repo" --quick >&2 2>&1 || true
                    if [[ -d "$ORG_PATH/$repo/repo" ]]; then
                        REPOS_TO_SCAN+=("$repo")
                        echo -e "${GREEN}✓ $repo ready${NC}" >&2
                    else
                        echo -e "${RED}✗ Failed to clone $repo${NC}" >&2
                    fi
                done
                echo "" >&2
            else
                echo -e "${CYAN}Continuing with ${#REPOS_TO_SCAN[@]} already-cloned repositories...${NC}" >&2
            fi
        fi

        if [[ ${#REPOS_TO_SCAN[@]} -eq 0 ]]; then
            echo '{"error": "No repositories with cloned code found in org cache. Hydrate repos first."}'
            exit 1
        fi

        # Analyze each repo and aggregate results
        echo -e "${BLUE}Scanning ${#REPOS_TO_SCAN[@]} repositories in $ORG...${NC}" >&2

        all_results="[]"
        repo_count=0
        total_repos=${#REPOS_TO_SCAN[@]}

        for repo_name in "${REPOS_TO_SCAN[@]}"; do
            ((repo_count++))
            scan_path="$ORG_PATH/$repo_name/repo"
            TARGET="$ORG/$repo_name"

            echo -e "\n${CYAN}[$repo_count/$total_repos] Analyzing: $TARGET${NC}" >&2

            repo_json=$(analyze_target "$scan_path")
            repo_json=$(echo "$repo_json" | jq --arg repo "$TARGET" '. + {repository: $repo}')

            all_results=$(echo "$all_results" | jq --argjson repo "$repo_json" '. + [$repo]')
        done

        # Build aggregated output
        timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

        final_json=$(jq -n \
            --arg ts "$timestamp" \
            --arg org "$ORG" \
            --arg ver "1.0.0" \
            --argjson repo_count "$total_repos" \
            --argjson repositories "$all_results" \
            '{
                analyzer: "git-insights",
                version: $ver,
                timestamp: $ts,
                organization: $org,
                summary: {
                    repositories_scanned: $repo_count
                },
                repositories: $repositories
            }')

        echo -e "\n${CYAN}=== Organization Summary ===${NC}" >&2
        echo -e "${CYAN}Repos analyzed: $total_repos${NC}" >&2

        # Output
        if [[ -n "$OUTPUT_FILE" ]]; then
            echo "$final_json" > "$OUTPUT_FILE"
            echo -e "${GREEN}✓ Results written to $OUTPUT_FILE${NC}" >&2
        else
            echo "$final_json"
        fi
        exit 0
    else
        echo '{"error": "Org not found in cache. Hydrate repos first."}'
        exit 1
    fi
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

echo -e "${BLUE}Analyzing git history: $TARGET${NC}" >&2

final_json=$(analyze_target "$scan_path")

# Output
if [[ -n "$OUTPUT_FILE" ]]; then
    echo "$final_json" > "$OUTPUT_FILE"
    echo -e "${GREEN}✓ Results written to $OUTPUT_FILE${NC}" >&2
else
    echo "$final_json"
fi
