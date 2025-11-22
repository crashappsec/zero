#!/bin/bash
# Copyright (c) 2024 Gibson Powers Contributors
# 
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Code Ownership Analyzer - Basic Analysis
# Analyzes git repository to identify code owners and generate reports
# Usage: ./ownership-analyzer.sh [options] <repository-path>
#############################################################################

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Default options
DAYS=90
FORMAT="text"
OUTPUT_FILE=""
VALIDATE_CODEOWNERS=false
CODEOWNERS_PATH=".github/CODEOWNERS"
TEMP_DIR=""
CLEANUP=true
USE_CLAUDE=false
ANTHROPIC_API_KEY="${ANTHROPIC_API_KEY:-}"

usage() {
    cat << EOF
Code Ownership Analyzer - Analyze repository ownership patterns

Analyzes git history to identify code owners, calculate metrics, and
optionally validate CODEOWNERS files.

Usage: $0 [OPTIONS] <target>

TARGET:
    Local directory path    Analyze local repository
    Git repository URL      Clone and analyze repository

OPTIONS:
    -d, --days N            Analyze last N days of history (default: 90)
    -f, --format FORMAT     Output format: text, json, csv (default: text)
    -o, --output FILE       Write output to file instead of stdout
    -v, --validate          Validate CODEOWNERS file if present
    -c, --codeowners PATH   Path to CODEOWNERS file (default: .github/CODEOWNERS)
    --claude                Use Claude AI for advanced analysis (requires ANTHROPIC_API_KEY)
    -k, --api-key KEY       Anthropic API key (or set ANTHROPIC_API_KEY env var)
    --keep-clone            Keep cloned repository (don't cleanup)
    -h, --help              Show this help message

EXAMPLES:
    # Analyze current directory (last 90 days)
    $0 .

    # Analyze GitHub repository
    $0 https://github.com/org/repo

    # Analyze specific repository (last 180 days)
    $0 --days 180 /path/to/repo

    # Generate JSON output
    $0 --format json --output ownership.json .

    # Validate CODEOWNERS file
    $0 --validate --codeowners .github/CODEOWNERS .

    # Full analysis with validation
    $0 --days 90 --validate --format json --output report.json .

    # Use Claude AI for advanced insights
    $0 --claude .

EOF
    exit 1
}

check_prerequisites() {
    if ! command -v git &> /dev/null; then
        echo -e "${RED}Error: git is not installed${NC}"
        exit 1
    fi

    if ! command -v jq &> /dev/null; then
        echo -e "${RED}Error: jq is not installed${NC}"
        echo "Install: brew install jq  (or apt-get install jq)"
        exit 1
    fi

    if ! command -v bc &> /dev/null; then
        echo -e "${RED}Error: bc is not installed${NC}"
        echo "Install: brew install bc  (or apt-get install bc)"
        exit 1
    fi
}

# Function to detect if target is a Git URL
is_git_url() {
    [[ "$1" =~ ^(https?|git)://.*\.git$ ]] || [[ "$1" =~ ^git@.*:.*\.git$ ]] || [[ "$1" =~ github\.com|gitlab\.com|bitbucket\.org ]]
}

# Function to clone repository
clone_repository() {
    local repo_url="$1"
    TEMP_DIR=$(mktemp -d)

    echo -e "${BLUE}Cloning repository (full history for ownership analysis)...${NC}"
    if git clone "$repo_url" "$TEMP_DIR"; then
        echo -e "${GREEN}✓ Repository cloned with full history${NC}"
        return 0
    else
        echo -e "${RED}✗ Failed to clone repository${NC}"
        echo -e "${YELLOW}Note: For private repositories, ensure you have proper SSH keys or use HTTPS with credentials${NC}"
        return 1
    fi
}

# Function to cleanup
cleanup() {
    if [[ "$CLEANUP" == true ]] && [[ -n "$TEMP_DIR" ]] && [[ -d "$TEMP_DIR" ]]; then
        rm -rf "$TEMP_DIR"
    fi
}

is_git_repository() {
    local repo_path="$1"
    if [[ ! -d "$repo_path/.git" ]]; then
        echo -e "${RED}Error: $repo_path is not a git repository${NC}"
        exit 1
    fi
}

# Function to look up GitHub profile by email
lookup_github_profile() {
    local email="$1"
    local name="$2"

    # Try to find GitHub username from git remote URL first
    local remote_url=$(git config --get remote.origin.url 2>/dev/null || echo "")
    local github_org=""

    if [[ "$remote_url" =~ github\.com[:/]([^/]+)/ ]]; then
        github_org="${BASH_REMATCH[1]}"
    fi

    # Use GitHub API to search for user by email (requires authenticated API call for best results)
    # For now, we'll extract username from email if it's a github.com email or noreply email
    local github_user=""

    if [[ "$email" =~ ^([^@]+)@users\.noreply\.github\.com$ ]]; then
        # Extract from noreply email: username@users.noreply.github.com or 12345+username@...
        github_user=$(echo "$email" | sed -E 's/^([0-9]+\+)?([^@]+)@.*/\2/')
    elif [[ "$email" =~ ^([^@]+)@github\.com$ ]]; then
        github_user="${BASH_REMATCH[1]}"
    else
        # Try to query GitHub API search (public API, no auth needed but rate limited)
        local api_response=$(curl -s -H "Accept: application/vnd.github.v3+json" \
            "https://api.github.com/search/users?q=$email+in:email" 2>/dev/null)

        if [[ $(echo "$api_response" | jq -r '.total_count // 0' 2>/dev/null) -gt 0 ]]; then
            github_user=$(echo "$api_response" | jq -r '.items[0].login // ""' 2>/dev/null)
        fi
    fi

    if [[ -n "$github_user" ]]; then
        echo "$github_user|https://github.com/$github_user"
    else
        echo "|"
    fi
}

# Cache file for GitHub profile lookups
GITHUB_CACHE_FILE="/tmp/github_profile_cache_$$.tmp"

# Function to get GitHub profile with caching
get_github_profile() {
    local email="$1"
    local name="$2"

    # Check cache first (file-based for bash 3.x compatibility)
    if [[ -f "$GITHUB_CACHE_FILE" ]]; then
        local cached=$(grep "^$email|" "$GITHUB_CACHE_FILE" 2>/dev/null | cut -d'|' -f2-)
        if [[ -n "$cached" ]]; then
            echo "$cached"
            return
        fi
    fi

    # Look up profile
    local result=$(lookup_github_profile "$email" "$name")
    echo "$email|$result" >> "$GITHUB_CACHE_FILE"
    echo "$result"
}

analyze_ownership() {
    local repo_path="$1"
    local days="$2"

    cd "$repo_path" || exit 1

    echo -e "${BLUE}Analyzing repository: $(basename "$repo_path")${NC}"
    echo -e "${BLUE}Time period: Last $days days${NC}"
    echo ""

    # Get total files
    local total_files=$(git ls-files | wc -l | tr -d ' ')

    # Get contributors in the time period
    local since_date=$(date -v-${days}d +%Y-%m-%d 2>/dev/null || date -d "$days days ago" +%Y-%m-%d)

    # Collect ownership data
    local ownership_data=$(mktemp)
    git log --since="$since_date" --format="%an|%ae" --name-only | \
        grep -v "^$" | \
        awk -F'|' '
        NF==2 {author=$1; email=$2; next}
        {files[author"|"email"|"$0]++}
        END {for (key in files) print key"|"files[key]}
        ' > "$ownership_data"

    # Calculate per-author statistics
    local author_stats=$(mktemp)
    git log --since="$since_date" --format="%an|%ae" --numstat | \
        awk -F'|' '
        NF==2 {author=$1; next}
        NF==3 {
            commits[author]++
            added[author]+=$1
            deleted[author]+=$2
        }
        END {
            for (author in commits) {
                print author"|"commits[author]"|"added[author]"|"deleted[author]
            }
        }
        ' > "$author_stats"

    # Get commit counts per author
    local commit_counts=$(git log --since="$since_date" --format="%an" | sort | uniq -c | sort -rn)

    # Calculate metrics
    local total_commits=$(git log --since="$since_date" --oneline | wc -l | tr -d ' ')
    local active_authors=$(git log --since="$since_date" --format="%an" | sort -u | wc -l | tr -d ' ')

    # Export results based on format
    if [[ "$FORMAT" == "json" ]]; then
        generate_json_output "$total_files" "$total_commits" "$active_authors" "$ownership_data" "$author_stats"
    elif [[ "$FORMAT" == "csv" ]]; then
        generate_csv_output "$ownership_data" "$author_stats"
    else
        generate_text_output "$total_files" "$total_commits" "$active_authors" "$ownership_data" "$author_stats" "$commit_counts"
    fi

    rm -f "$ownership_data" "$author_stats"
}

generate_text_output() {
    local total_files="$1"
    local total_commits="$2"
    local active_authors="$3"
    local ownership_data="$4"
    local author_stats="$5"
    local commit_counts="$6"

    cat << EOF
========================================
  Code Ownership Analysis Report
========================================

Repository Metrics:
-------------------
Total Files: $total_files
Total Commits (period): $total_commits
Active Contributors: $active_authors

Top Contributors (by commit count):
-----------------------------------
EOF

    # Get email addresses for top contributors
    git log --since="$since_date" --format="%an|%ae" | sort -u > /tmp/author_emails.tmp

    echo "$commit_counts" | head -10 | while read count author; do
        # Look up email for this author
        local email=$(grep "^$author|" /tmp/author_emails.tmp | head -1 | cut -d'|' -f2)
        local github_info=$(get_github_profile "$email" "$author")
        local github_user=$(echo "$github_info" | cut -d'|' -f1)
        local github_url=$(echo "$github_info" | cut -d'|' -f2)

        if [[ -n "$github_user" ]]; then
            printf "%-40s %5d commits (@%s - %s)\n" "$author" "$count" "$github_user" "$github_url"
        else
            printf "%-40s %5d commits\n" "$author" "$count"
        fi
    done

    rm -f /tmp/author_emails.tmp

    echo ""
    echo "Ownership Distribution:"
    echo "----------------------"

    # Calculate top file owners
    awk -F'|' '{
        author_email=$1"|"$2
        authors[author_email]+=$4
    }
    END {
        for (ae in authors) {
            split(ae, parts, "|")
            print authors[ae]"|"parts[1]
        }
    }' "$ownership_data" | sort -t'|' -k1 -rn | head -10 | while IFS='|' read files author; do
        printf "%-40s %5d files\n" "$author" "$files"
    done

    echo ""
    echo "Activity Summary:"
    echo "----------------"

    while IFS='|' read author commits added deleted; do
        local net=$((added - deleted))
        printf "%-30s %4d commits, +%d -%d lines (net: %+d)\n" \
            "$author" "$commits" "$added" "$deleted" "$net"
    done < "$author_stats" | head -10

    echo ""
    echo "========================================="
}

generate_json_output() {
    local total_files="$1"
    local total_commits="$2"
    local active_authors="$3"
    local ownership_data="$4"
    local author_stats="$5"

    jq -n \
        --arg total_files "$total_files" \
        --arg total_commits "$total_commits" \
        --arg active_authors "$active_authors" \
        --arg analysis_date "$(date +%Y-%m-%d)" \
        --arg days "$DAYS" \
        '{
            metadata: {
                repository: $ENV.PWD,
                analysis_date: $analysis_date,
                time_period_days: ($days | tonumber),
                total_files: ($total_files | tonumber),
                total_commits: ($total_commits | tonumber),
                active_contributors: ($active_authors | tonumber)
            },
            contributors: []
        }'
}

generate_csv_output() {
    local ownership_data="$1"
    local author_stats="$2"

    echo "Author,Email,Files Owned,Commits,Lines Added,Lines Deleted,Net Lines"

    # Combine data
    join -t'|' \
        <(awk -F'|' '{ae=$1"|"$2; files[ae]+=$4} END {for(k in files) print k"|"files[k]}' "$ownership_data" | sort) \
        <(sort "$author_stats") | \
        while IFS='|' read author email files commits added deleted; do
            local net=$((added - deleted))
            echo "$author,$email,$files,$commits,$added,$deleted,$net"
        done
}

validate_codeowners_file() {
    local repo_path="$1"
    local codeowners_file="$repo_path/$CODEOWNERS_PATH"

    if [[ ! -f "$codeowners_file" ]]; then
        echo -e "${YELLOW}⚠ CODEOWNERS file not found at: $codeowners_file${NC}"
        return 1
    fi

    echo -e "${BLUE}Validating CODEOWNERS file: $codeowners_file${NC}"
    echo ""

    local total_patterns=0
    local valid_patterns=0
    local issues=0

    while IFS= read -r line; do
        # Skip comments and empty lines
        [[ "$line" =~ ^#.*$ ]] && continue
        [[ -z "$line" ]] && continue

        ((total_patterns++))

        # Extract pattern and owners
        local pattern=$(echo "$line" | awk '{print $1}')
        local owners=$(echo "$line" | cut -d' ' -f2-)

        # Basic validation
        if [[ -n "$pattern" && -n "$owners" ]]; then
            ((valid_patterns++))
        else
            echo -e "${YELLOW}⚠ Invalid pattern: $line${NC}"
            ((issues++))
        fi
    done < "$codeowners_file"

    echo "CODEOWNERS Validation Results:"
    echo "----------------------------"
    echo "Total patterns: $total_patterns"
    echo "Valid patterns: $valid_patterns"
    echo "Issues found: $issues"

    if [[ $issues -eq 0 ]]; then
        echo -e "${GREEN}✓ CODEOWNERS file syntax is valid${NC}"
    else
        echo -e "${RED}✗ CODEOWNERS file has $issues issues${NC}"
    fi
    echo ""
}

#############################################################################
# Claude AI Functions (only used when --claude flag is set)
#############################################################################

collect_repository_data_for_claude() {
    local repo_path="$1"
    local days="$2"

    cd "$repo_path" || exit 1

    local repo_name=$(basename "$(git rev-parse --show-toplevel)")
    local since_date=$(date -v-${days}d +%Y-%m-%d 2>/dev/null || date -d "${days} days ago" +%Y-%m-%d)

    # Basic repository stats
    local total_files=$(git ls-files | wc -l | tr -d ' ')
    local total_commits=$(git log --since="$since_date" --oneline | wc -l | tr -d ' ')
    local top_contributors=$(git log --since="$since_date" --format="%an" | sort | uniq -c | sort -rn | head -10)

    # Contributor details
    local contributor_files=$(mktemp)
    git log --since="$since_date" --name-only --format="%an" | \
        awk 'NF==1 && $1!="" {author=$1; next} {files[author"|"$0]++} END {
            for(key in files) {
                split(key, parts, "|")
                author_files[parts[1]]++
            }
            for(author in author_files) {
                print author"|"author_files[author]
            }
        }' | sort -t'|' -k2 -rn > "$contributor_files"

    # Build JSON
    local contributors_json="["
    local first=true
    while IFS='|' read author files; do
        if [[ "$first" == "true" ]]; then
            first=false
        else
            contributors_json+=","
        fi
        contributors_json+=$(jq -n \
            --arg author "$author" \
            --arg files "$files" \
            '{author: $author, files: ($files | tonumber)}')
    done < "$contributor_files"
    contributors_json+="]"

    rm -f "$contributor_files"

    # Build final JSON
    jq -n \
        --arg repo_name "$repo_name" \
        --arg total_files "$total_files" \
        --arg total_commits "$total_commits" \
        --arg days "$days" \
        --argjson contributors "$contributors_json" \
        '{
            repository: $repo_name,
            analysis_period_days: ($days | tonumber),
            total_files: ($total_files | tonumber),
            total_commits: ($total_commits | tonumber),
            contributors: $contributors
        }'
}

analyze_with_claude() {
    local data="$1"
    local model="claude-sonnet-4-20250514"

    if [[ -z "$ANTHROPIC_API_KEY" ]]; then
        echo -e "${RED}Error: ANTHROPIC_API_KEY is required for Claude analysis${NC}"
        echo "Set it with: export ANTHROPIC_API_KEY=your-key"
        echo "Or use: --api-key your-key"
        exit 1
    fi

    echo -e "${BLUE}Analyzing with Claude AI...${NC}"

    local prompt="Analyze this code ownership data and provide insights on:
1. Ownership concentration and bus factor risks
2. Knowledge silos and recommendations
3. CODEOWNERS file recommendations
4. Succession planning suggestions

Repository Data:
$data

Provide a concise, actionable analysis."

    local response=$(curl -s https://api.anthropic.com/v1/messages \
        -H "content-type: application/json" \
        -H "x-api-key: $ANTHROPIC_API_KEY" \
        -H "anthropic-version: 2023-06-01" \
        -d "$(jq -n \
            --arg model "$model" \
            --arg prompt "$prompt" \
            '{
                model: $model,
                max_tokens: 4096,
                messages: [{
                    role: "user",
                    content: $prompt
                }]
            }')")

    # Record API usage for cost tracking
    if command -v record_api_usage &> /dev/null; then
        record_api_usage "$response" "$model" > /dev/null
    fi

    local analysis=$(echo "$response" | jq -r '.content[0].text // empty')

    if [[ -z "$analysis" ]]; then
        echo -e "${RED}✗ Claude API error${NC}"
        echo "$response" | jq .
        return 1
    fi

    echo -e "${GREEN}✓ Analysis complete${NC}"
    echo ""
    echo "========================================="
    echo "  Claude AI Code Ownership Analysis"
    echo "========================================="
    echo ""
    echo "$analysis"
    echo ""
}

# Parse arguments
REPO_PATH=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -d|--days)
            DAYS="$2"
            shift 2
            ;;
        -f|--format)
            FORMAT="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_FILE="$2"
            shift 2
            ;;
        -v|--validate)
            VALIDATE_CODEOWNERS=true
            shift
            ;;
        -c|--codeowners)
            CODEOWNERS_PATH="$2"
            shift 2
            ;;
        --claude)
            USE_CLAUDE=true
            shift
            ;;
        -k|--api-key)
            ANTHROPIC_API_KEY="$2"
            shift 2
            ;;
        --keep-clone)
            CLEANUP=false
            shift
            ;;
        -h|--help)
            usage
            ;;
        *)
            REPO_PATH="$1"
            shift
            ;;
    esac
done

if [[ -z "$REPO_PATH" ]]; then
    echo -e "${RED}Error: No repository path specified${NC}"
    usage
fi

# Load cost tracking library if using Claude
if [[ "$USE_CLAUDE" == "true" ]]; then
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
    if [ -f "$REPO_ROOT/utils/lib/claude-cost.sh" ]; then
        source "$REPO_ROOT/utils/lib/claude-cost.sh"
        init_cost_tracking
    fi
fi

# Main execution
echo ""
echo "========================================="
if [[ "$USE_CLAUDE" == "true" ]]; then
    echo "  Code Ownership Analyzer (Claude AI Mode)"
else
    echo "  Code Ownership Analyzer"
fi
echo "========================================="
echo ""

check_prerequisites

# Determine target type and set REPO_PATH
if is_git_url "$REPO_PATH"; then
    echo -e "${GREEN}Target: Git repository${NC}"
    if clone_repository "$REPO_PATH"; then
        ACTUAL_REPO_PATH="$TEMP_DIR"
    else
        exit 1
    fi
else
    echo -e "${GREEN}Target: Local directory${NC}"
    ACTUAL_REPO_PATH="$REPO_PATH"
fi

is_git_repository "$ACTUAL_REPO_PATH"

if [[ "$VALIDATE_CODEOWNERS" == "true" ]]; then
    validate_codeowners_file "$ACTUAL_REPO_PATH"
fi

# Run analysis based on mode
if [[ "$USE_CLAUDE" == "true" ]]; then
    # Claude AI analysis mode
    data=$(collect_repository_data_for_claude "$ACTUAL_REPO_PATH" "$DAYS")
    if [[ -n "$OUTPUT_FILE" ]]; then
        analyze_with_claude "$data" > "$OUTPUT_FILE"
        echo -e "${GREEN}✓ Analysis complete${NC}"
        echo -e "Output written to: $OUTPUT_FILE"
    else
        analyze_with_claude "$data"
    fi

    # Display cost summary
    if command -v display_api_cost_summary &> /dev/null; then
        display_api_cost_summary
    fi
else
    # Standard analysis mode
    if [[ -n "$OUTPUT_FILE" ]]; then
        analyze_ownership "$ACTUAL_REPO_PATH" "$DAYS" > "$OUTPUT_FILE"
        echo -e "${GREEN}✓ Analysis complete${NC}"
        echo -e "Output written to: $OUTPUT_FILE"
    else
        analyze_ownership "$ACTUAL_REPO_PATH" "$DAYS"
    fi
fi

# Cleanup if we cloned a repo
cleanup

echo ""
echo "========================================="
echo -e "${GREEN}  Analysis Complete${NC}"
echo "========================================="
echo ""
