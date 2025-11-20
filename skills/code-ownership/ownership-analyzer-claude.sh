#!/bin/bash
# Copyright (c) 2024 Crash Override Inc
# 101 Fulton St, 416, New York 10038
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Code Ownership Analyzer with Claude AI Integration
# Analyzes ownership and provides AI-enhanced insights and recommendations
# Usage: ./ownership-analyzer-claude.sh [options] <repository-path>
#############################################################################

set -e

# Load environment variables from .env file if it exists
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
if [ -f "$REPO_ROOT/.env" ]; then
    source "$REPO_ROOT/.env"
fi

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

ANTHROPIC_API_KEY="${ANTHROPIC_API_KEY:-}"
DAYS=90
CODEOWNERS_PATH=".github/CODEOWNERS"
TEMP_DIR=""
CLEANUP=true

usage() {
    cat << EOF
Code Ownership Analyzer with Claude AI - Enhanced analysis with insights

Analyzes git repository ownership patterns and provides AI-powered:
- Risk assessment and prioritized recommendations
- CODEOWNERS validation with specific fixes
- Knowledge transfer planning
- Succession recommendations

Usage: $0 [OPTIONS] <target>

TARGET:
    Local directory path    Analyze local repository
    Git repository URL      Clone and analyze repository

OPTIONS:
    -d, --days N            Analyze last N days of history (default: 90)
    -k, --api-key KEY       Anthropic API key (or set ANTHROPIC_API_KEY env var)
    -c, --codeowners PATH   Path to CODEOWNERS file (default: .github/CODEOWNERS)
    --keep-clone            Keep cloned repository (don't cleanup)
    -h, --help              Show this help message

ENVIRONMENT:
    ANTHROPIC_API_KEY       Your Anthropic API key

EXAMPLES:
    # Analyze local repository
    $0 .

    # Analyze GitHub repository
    $0 https://github.com/org/repo

    # Analyze specific time period
    $0 --days 180 /path/to/repo

    # Specify API key directly
    $0 --api-key sk-ant-xxx https://github.com/org/repo

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
        echo "Install: brew install jq"
        exit 1
    fi

    if [[ -z "$ANTHROPIC_API_KEY" ]]; then
        echo -e "${RED}Error: ANTHROPIC_API_KEY not set${NC}"
        echo ""
        echo "Set your API key:"
        echo "  export ANTHROPIC_API_KEY=sk-ant-xxx"
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

# Function to look up GitHub profile by email
lookup_github_profile() {
    local email="$1"
    local name="$2"

    # Extract username from email if it's a github.com email or noreply email
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

# Cache for GitHub profile lookups
declare -A GITHUB_PROFILE_CACHE

# Function to get GitHub profile with caching
get_github_profile() {
    local email="$1"
    local name="$2"

    # Check cache first
    if [[ -n "${GITHUB_PROFILE_CACHE[$email]}" ]]; then
        echo "${GITHUB_PROFILE_CACHE[$email]}"
        return
    fi

    # Look up profile
    local result=$(lookup_github_profile "$email" "$name")
    GITHUB_PROFILE_CACHE[$email]="$result"
    echo "$result"
}

collect_repository_data() {
    local repo_path="$1"
    local days="$2"

    cd "$repo_path" || exit 1

    echo -e "${BLUE}Collecting repository data...${NC}"

    local repo_name=$(basename "$(git rev-parse --show-toplevel)")
    local total_files=$(git ls-files | wc -l | tr -d ' ')
    local since_date=$(date -v-${days}d +%Y-%m-%d 2>/dev/null || date -d "$days days ago" +%Y-%m-%d)
    local total_commits=$(git log --since="$since_date" --oneline | wc -l | tr -d ' ')

    # Get top contributors
    local top_contributors=$(git log --since="$since_date" --format="%an" | \
        sort | uniq -c | sort -rn | head -10 | \
        awk '{printf "  - %s: %d commits\n", $2, $1}')

    # Get file count per contributor
    local contributor_files=$(mktemp)
    git log --since="$since_date" --format="%an" --name-only | \
        awk 'NF==1 && $1!="" {author=$1; next} {files[author"|"$0]++} END {
            for(key in files) {
                split(key, parts, "|")
                author_files[parts[1]]++
            }
            for(author in author_files) {
                print author"|"author_files[author]
            }
        }' | sort -t'|' -k2 -rn > "$contributor_files"

    # Get recent activity per contributor
    local contributor_activity=$(mktemp)
    git log --since="$since_date" --format="%an|%ad" --date=short | \
        awk -F'|' '{last[$1]=$2} END {for(author in last) print author"|"last[author]}' \
        > "$contributor_activity"

    # Check for CODEOWNERS file
    local has_codeowners="false"
    local codeowners_patterns=0
    if [[ -f "$CODEOWNERS_PATH" ]]; then
        has_codeowners="true"
        codeowners_patterns=$(grep -v "^#" "$CODEOWNERS_PATH" | grep -v "^$" | wc -l | tr -d ' ')
    fi

    # Get file ownership distribution (files changed by single author)
    local single_author_files=$(git ls-files | while read file; do
        local author_count=$(git log --since="$since_date" --format="%an" -- "$file" | sort -u | wc -l | tr -d ' ')
        if [[ $author_count -eq 1 ]]; then
            echo "$file"
        fi
    done | wc -l | tr -d ' ')

    # Build JSON data
    local data=$(jq -n \
        --arg repo_name "$repo_name" \
        --arg total_files "$total_files" \
        --arg total_commits "$total_commits" \
        --arg days "$days" \
        --arg has_codeowners "$has_codeowners" \
        --arg codeowners_patterns "$codeowners_patterns" \
        --arg single_author_files "$single_author_files" \
        --arg top_contributors "$top_contributors" \
        '{
            repository: $repo_name,
            analysis_period_days: ($days | tonumber),
            total_files: ($total_files | tonumber),
            total_commits: ($total_commits | tonumber),
            has_codeowners: ($has_codeowners == "true"),
            codeowners_patterns: ($codeowners_patterns | tonumber),
            single_author_files: ($single_author_files | tonumber),
            top_contributors: $top_contributors
        }')

    # Get author emails for GitHub profile lookup
    git log --since="$since_date" --format="%an|%ae" | sort -u > /tmp/author_emails_$$.tmp

    # Add contributor details
    local contributors_json="["
    local first=true
    while IFS='|' read author files; do
        local last_activity=$(grep "^$author|" "$contributor_activity" | cut -d'|' -f2)
        local email=$(grep "^$author|" /tmp/author_emails_$$.tmp | head -1 | cut -d'|' -f2)
        local github_info=$(get_github_profile "$email" "$author")
        local github_user=$(echo "$github_info" | cut -d'|' -f1)
        local github_url=$(echo "$github_info" | cut -d'|' -f2)

        if [[ "$first" == "true" ]]; then
            first=false
        else
            contributors_json+=","
        fi
        contributors_json+=$(jq -n \
            --arg author "$author" \
            --arg files "$files" \
            --arg last_activity "${last_activity:-unknown}" \
            --arg github_user "$github_user" \
            --arg github_url "$github_url" \
            '{author: $author, files: ($files | tonumber), last_activity: $last_activity, github_username: $github_user, github_profile: $github_url}')
    done < "$contributor_files"
    contributors_json+="]"

    rm -f /tmp/author_emails_$$.tmp

    data=$(echo "$data" | jq --argjson contributors "$contributors_json" '. + {contributors: $contributors}')

    rm -f "$contributor_files" "$contributor_activity"

    echo "$data"
}

analyze_with_claude() {
    local data="$1"

    echo -e "${BLUE}Analyzing with Claude AI...${NC}"

    local prompt="I need you to analyze this code ownership data for a git repository.

Data:
\`\`\`json
$data
\`\`\`

Please provide a comprehensive code ownership analysis with:

1. **Executive Summary**
   - Overall ownership health assessment
   - Key findings and concerns
   - Critical actions needed

2. **Ownership Analysis**
   - Coverage assessment (files with clear owners)
   - Distribution analysis (concentration vs balance)
   - Top contributors and their impact
   - Single points of failure (SPOFs)

3. **Risk Assessment**
   - Identify critical risks (high concentration, inactive owners, knowledge gaps)
   - Calculate estimated bus factor
   - Prioritize risks by impact

4. **CODEOWNERS File Assessment**
   - If file exists: validate accuracy and completeness
   - If missing: assess whether one should be created
   - Note any discrepancies between file and actual contributions

Be specific and data-driven. Focus on objective analysis of ownership patterns and risks."

    local response=$(curl -s https://api.anthropic.com/v1/messages \
        -H "content-type: application/json" \
        -H "x-api-key: $ANTHROPIC_API_KEY" \
        -H "anthropic-version: 2023-06-01" \
        -d "{
            \"model\": \"claude-sonnet-4-20250514\",
            \"max_tokens\": 4096,
            \"messages\": [{
                \"role\": \"user\",
                \"content\": $(echo "$prompt" | jq -Rs .)
            }]
        }")

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
        -k|--api-key)
            ANTHROPIC_API_KEY="$2"
            shift 2
            ;;
        -c|--codeowners)
            CODEOWNERS_PATH="$2"
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

# Main
echo ""
echo "========================================="
echo "  Ownership Analyzer with Claude AI"
echo "========================================="
echo ""

check_prerequisites

# Determine target type
if is_git_url "$REPO_PATH"; then
    echo -e "${GREEN}Target: Git repository${NC}"
    if clone_repository "$REPO_PATH"; then
        data=$(collect_repository_data "$TEMP_DIR" "$DAYS")
        analyze_with_claude "$data"
        cleanup
    else
        exit 1
    fi
elif [[ -d "$REPO_PATH" ]]; then
    echo -e "${GREEN}Target: Local directory${NC}"
    if [[ ! -d "$REPO_PATH/.git" ]]; then
        echo -e "${RED}Error: $REPO_PATH is not a git repository${NC}"
        exit 1
    fi
    data=$(collect_repository_data "$REPO_PATH" "$DAYS")
    analyze_with_claude "$data"
else
    echo -e "${RED}Error: Invalid target${NC}"
    echo "Target must be a local directory or Git repository URL"
    exit 1
fi

echo "========================================="
echo -e "${GREEN}  Analysis Complete${NC}"
echo "========================================="
echo ""
