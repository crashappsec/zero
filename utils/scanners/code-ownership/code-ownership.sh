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

    # Get contributor data efficiently using git shortlog
    # This avoids running multiple expensive git log commands per contributor
    echo -e "${BLUE}Collecting contributor data...${NC}" >&2
    local contributors_json="[]"

    # Use git shortlog for efficient contributor stats (single pass through git history)
    local shortlog_args="-sne"
    if [[ -n "$since_date" ]]; then
        shortlog_args="$shortlog_args --since=$since_date"
    fi

    # Get top 20 contributors with commit counts
    local contributor_data=$(git shortlog $shortlog_args 2>/dev/null | head -20)

    while IFS= read -r line; do
        [[ -z "$line" ]] && continue

        # Parse shortlog output: "   123\tAuthor Name <email@example.com>"
        local count=$(echo "$line" | awk '{print $1}')
        local name_email=$(echo "$line" | sed 's/^[[:space:]]*[0-9]*[[:space:]]*//')
        local author=$(echo "$name_email" | sed 's/<.*>//' | sed 's/[[:space:]]*$//')
        local email=$(echo "$name_email" | grep -o '<[^>]*>' | tr -d '<>')

        [[ -z "$count" ]] && continue
        [[ -z "$email" ]] && continue

        # Skip expensive per-author git log operations for large repos
        # Just use commit count as primary metric
        contributors_json=$(echo "$contributors_json" | jq \
            --arg name "$author" \
            --arg email "$email" \
            --argjson commits "$count" \
            '. + [{"name": $name, "email": $email, "commits": $commits}]')
    done <<< "$contributor_data"

    local contributor_count=$(echo "$contributors_json" | jq 'length')
    echo -e "${GREEN}✓ Found $contributor_count contributors${NC}" >&2

    # CODEOWNERS analysis (with validation from codeowners-validator)
    echo -e "${BLUE}Checking CODEOWNERS...${NC}" >&2
    local codeowners_json='{"exists": false}'
    local codeowners_file="$repo_path/$CODEOWNERS_PATH"

    # Also check alternate locations
    if [[ ! -f "$codeowners_file" ]]; then
        for alt_path in "CODEOWNERS" "docs/CODEOWNERS" ".gitlab/CODEOWNERS"; do
            if [[ -f "$repo_path/$alt_path" ]]; then
                codeowners_file="$repo_path/$alt_path"
                CODEOWNERS_PATH="$alt_path"
                break
            fi
        done
    fi

    if [[ -f "$codeowners_file" ]]; then
        local patterns=$(grep -v "^#" "$codeowners_file" | grep -v "^$" | wc -l | tr -d ' ')
        local owners_list=$(grep -v "^#" "$codeowners_file" | grep -v "^$" | awk '{for(i=2;i<=NF;i++) print $i}' | sort -u)
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
        done < "$codeowners_file"

        # === VALIDATION (from codeowners-validator) ===
        local validation_issues="[]"
        local syntax_errors=0
        local antipattern_count=0

        # Syntax validation
        local line_num=0
        while IFS= read -r line; do
            ((line_num++))
            [[ "$line" =~ ^#.*$ ]] && continue
            [[ -z "$line" ]] && continue

            local pattern=$(echo "$line" | awk '{print $1}')
            local owners=$(echo "$line" | cut -d' ' -f2-)

            # Check for missing owners
            if [[ -z "$owners" ]]; then
                validation_issues=$(echo "$validation_issues" | jq \
                    --arg msg "Line $line_num: Pattern '$pattern' has no owners" \
                    --arg type "syntax" \
                    '. + [{"type": $type, "message": $msg, "severity": "error"}]')
                ((syntax_errors++))
                continue
            fi

            # Validate owner format (@username or @org/team)
            for owner in $owners; do
                if [[ ! "$owner" =~ ^@[a-zA-Z0-9_-]+(/[a-zA-Z0-9_-]+)?$ ]]; then
                    validation_issues=$(echo "$validation_issues" | jq \
                        --arg msg "Line $line_num: Invalid owner format '$owner'" \
                        --arg type "syntax" \
                        '. + [{"type": $type, "message": $msg, "severity": "error"}]')
                    ((syntax_errors++))
                fi
            done
        done < "$codeowners_file"

        # Anti-pattern detection
        # 1. Single owner for all files
        if grep -q "^\* @[a-zA-Z0-9_-]*$" "$codeowners_file" 2>/dev/null; then
            validation_issues=$(echo "$validation_issues" | jq \
                '. + [{"type": "antipattern", "message": "Single owner for all files (*) - consider distributing ownership", "severity": "warning"}]')
            ((antipattern_count++))
        fi

        # 2. Too many owners (>5) on single pattern
        grep -v "^#" "$codeowners_file" | grep -v "^$" | while IFS= read -r line; do
            local owner_count=$(echo "$line" | awk '{print NF-1}')
            if [[ $owner_count -gt 5 ]]; then
                local pat=$(echo "$line" | awk '{print $1}')
                validation_issues=$(echo "$validation_issues" | jq \
                    --arg msg "Pattern '$pat' has $owner_count owners - consider reducing" \
                    '. + [{"type": "antipattern", "message": $msg, "severity": "warning"}]')
                ((antipattern_count++))
            fi
        done

        # 3. No default owner pattern
        if ! grep -q "^\*" "$codeowners_file" 2>/dev/null; then
            validation_issues=$(echo "$validation_issues" | jq \
                '. + [{"type": "antipattern", "message": "No default owner pattern (*) - some files may lack ownership", "severity": "info"}]')
            ((antipattern_count++))
        fi

        # 4. Duplicate patterns
        local dup_patterns=$(grep -v "^#" "$codeowners_file" | grep -v "^$" | awk '{print $1}' | sort | uniq -d)
        if [[ -n "$dup_patterns" ]]; then
            validation_issues=$(echo "$validation_issues" | jq \
                --arg dups "$dup_patterns" \
                '. + [{"type": "antipattern", "message": ("Duplicate patterns: " + $dups), "severity": "warning"}]')
            ((antipattern_count++))
        fi

        local validation_status="valid"
        if [[ $syntax_errors -gt 0 ]]; then
            validation_status="invalid"
        elif [[ $antipattern_count -gt 0 ]]; then
            validation_status="warnings"
        fi

        codeowners_json=$(jq -n \
            --arg path "$CODEOWNERS_PATH" \
            --argjson patterns "$patterns" \
            --argjson unique_owners "$unique_owners" \
            --argjson pattern_list "$patterns_json" \
            --argjson validation_issues "$validation_issues" \
            --arg validation_status "$validation_status" \
            --argjson syntax_errors "$syntax_errors" \
            --argjson antipattern_count "$antipattern_count" \
            '{
                exists: true,
                path: $path,
                total_patterns: $patterns,
                unique_owners: $unique_owners,
                patterns: $pattern_list,
                validation: {
                    status: $validation_status,
                    syntax_errors: $syntax_errors,
                    antipatterns: $antipattern_count,
                    issues: $validation_issues
                }
            }')

        echo -e "${GREEN}✓ CODEOWNERS found: $patterns patterns, $unique_owners owners${NC}" >&2
        if [[ $syntax_errors -gt 0 ]]; then
            echo -e "${RED}  ✗ $syntax_errors syntax errors${NC}" >&2
        fi
        if [[ $antipattern_count -gt 0 ]]; then
            echo -e "${YELLOW}  ⚠ $antipattern_count anti-patterns detected${NC}" >&2
        fi
    else
        echo -e "${YELLOW}⚠ No CODEOWNERS file found${NC}" >&2
    fi

    # Build output
    # Note: Bus factor analysis has been moved to the separate bus-factor scanner
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    jq -n \
        --arg ts "$timestamp" \
        --arg tgt "$TARGET" \
        --arg ver "1.2.0" \
        --argjson days "$days" \
        --argjson total_files "$total_files" \
        --argjson total_commits "$total_commits" \
        --argjson contributor_count "$contributor_count" \
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
                active_contributors: $contributor_count
            },
            contributors: $contributors,
            codeowners: $codeowners,
            note: "For bus factor analysis, use the dedicated bus-factor scanner"
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
