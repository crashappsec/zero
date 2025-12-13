#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# GitHub Integration Library
# Comprehensive GitHub API integration and profile management
#############################################################################

# Cache file for GitHub profile lookups (session-based)
GITHUB_CACHE_FILE="/tmp/github_profile_cache_$$.tmp"

# Initialize cache
init_github_cache() {
    : > "$GITHUB_CACHE_FILE"
}

# Clean up cache
cleanup_github_cache() {
    rm -f "$GITHUB_CACHE_FILE"
}

# Extract GitHub username from noreply email
# Handles both username@users.noreply.github.com and 12345+username@users.noreply.github.com
extract_username_from_noreply() {
    local email="$1"

    if [[ "$email" =~ ^([0-9]+\+)?([^@]+)@users\.noreply\.github\.com$ ]]; then
        echo "${BASH_REMATCH[2]}"
        return 0
    fi

    return 1
}

# Extract GitHub username from github.com email
extract_username_from_github_email() {
    local email="$1"

    if [[ "$email" =~ ^([^@]+)@github\.com$ ]]; then
        echo "${BASH_REMATCH[1]}"
        return 0
    fi

    return 1
}

# Query GitHub API to search for user by email
# Note: Rate limited, use sparingly
query_github_api_for_user() {
    local email="$1"
    local github_token="${GITHUB_TOKEN:-}"

    # Build API request
    local api_url="https://api.github.com/search/users?q=${email}+in:email"
    local auth_header=""

    if [[ -n "$github_token" ]]; then
        auth_header="-H \"Authorization: token $github_token\""
    fi

    # Query API
    local response=$(curl -s \
        -H "Accept: application/vnd.github.v3+json" \
        $auth_header \
        "$api_url" 2>/dev/null)

    # Check if we got results
    local total_count=$(echo "$response" | jq -r '.total_count // 0' 2>/dev/null)

    if [[ "$total_count" -gt 0 ]]; then
        local username=$(echo "$response" | jq -r '.items[0].login // ""' 2>/dev/null)
        if [[ -n "$username" ]]; then
            echo "$username"
            return 0
        fi
    fi

    return 1
}

# Comprehensive GitHub username lookup
# Tries multiple methods in order of reliability
lookup_github_username() {
    local email="$1"

    # Method 1: Extract from noreply email (most reliable)
    local username
    if username=$(extract_username_from_noreply "$email"); then
        echo "$username"
        return 0
    fi

    # Method 2: Extract from github.com email
    if username=$(extract_username_from_github_email "$email"); then
        echo "$username"
        return 0
    fi

    # Method 3: Query GitHub API (rate limited, use sparingly)
    if username=$(query_github_api_for_user "$email"); then
        echo "$username"
        return 0
    fi

    # No username found
    return 1
}

# Get GitHub profile with caching
# Returns: username|profile_url or empty if not found
get_github_profile() {
    local email="$1"

    # Check cache first
    if [[ -f "$GITHUB_CACHE_FILE" ]]; then
        local cached=$(grep "^${email}|" "$GITHUB_CACHE_FILE" 2>/dev/null | cut -d'|' -f2-)
        if [[ -n "$cached" ]]; then
            echo "$cached"
            return 0
        fi
    fi

    # Look up username
    local username
    if username=$(lookup_github_username "$email"); then
        local profile_url="https://github.com/$username"
        local result="$username|$profile_url"

        # Cache result
        echo "$email|$result" >> "$GITHUB_CACHE_FILE"
        echo "$result"
        return 0
    fi

    # No profile found - cache empty result to avoid re-lookup
    echo "$email|" >> "$GITHUB_CACHE_FILE"
    return 1
}

# Get repository info from GitHub API
get_repo_info() {
    local repo="$1"  # Format: owner/repo
    local github_token="${GITHUB_TOKEN:-}"

    local api_url="https://api.github.com/repos/$repo"
    local auth_header=""

    if [[ -n "$github_token" ]]; then
        auth_header="-H \"Authorization: token $github_token\""
    fi

    curl -s \
        -H "Accept: application/vnd.github.v3+json" \
        $auth_header \
        "$api_url" 2>/dev/null
}

# List repositories in organization with pagination support
# Usage: list_org_repos "organization-name"
list_org_repos() {
    local org="$1"
    local github_token="${GITHUB_TOKEN:-}"
    local page=1
    local per_page=100
    local all_repos=""

    while true; do
        local url="https://api.github.com/orgs/$org/repos?type=all&per_page=$per_page&page=$page"

        if [[ -n "$github_token" ]]; then
            response=$(curl -s -H "Authorization: token $github_token" "$url")
        else
            response=$(curl -s "$url")
        fi

        # Check for errors
        if echo "$response" | jq -e '.message' > /dev/null 2>&1; then
            local error_msg=$(echo "$response" | jq -r '.message')
            echo "Error: $error_msg" >&2
            return 1
        fi

        # Extract repository names
        local repos=$(echo "$response" | jq -r '.[].full_name' 2>/dev/null)

        # If no repos returned, we're done
        if [[ -z "$repos" ]]; then
            break
        fi

        # Append to results
        if [[ -n "$all_repos" ]]; then
            all_repos="$all_repos"$'\n'"$repos"
        else
            all_repos="$repos"
        fi

        # If we got less than per_page results, we're done
        local repo_count=$(echo "$response" | jq '. | length' 2>/dev/null)
        if [[ $repo_count -lt $per_page ]]; then
            break
        fi

        page=$((page + 1))
    done

    echo "$all_repos"
}

# Parse repository slug from various URL formats
# Usage: parse_repo_slug "https://github.com/owner/repo"
parse_repo_slug() {
    local url="$1"

    # Remove trailing .git
    url="${url%.git}"

    # Extract owner/repo from various formats
    if [[ "$url" =~ github\.com[:/]([^/]+/[^/]+) ]]; then
        echo "${BASH_REMATCH[1]}"
    else
        echo "$url"
    fi
}

# Check if string is a GitHub URL
# Usage: is_github_url "https://github.com/owner/repo"
is_github_url() {
    [[ "$1" =~ github\.com ]]
}

# Get pull request reviews for a repo
get_pr_reviews() {
    local repo="$1"
    local pr_number="$2"
    local github_token="${GITHUB_TOKEN:-}"

    local api_url="https://api.github.com/repos/$repo/pulls/$pr_number/reviews"
    local auth_header=""

    if [[ -n "$github_token" ]]; then
        auth_header="-H \"Authorization: token $github_token\""
    fi

    curl -s \
        -H "Accept: application/vnd.github.v3+json" \
        $auth_header \
        "$api_url" 2>/dev/null
}

# Get all pull requests for a repository (last N days)
get_repo_pull_requests() {
    local repo="$1"
    local since_date="$2"
    local state="${3:-all}"  # all, open, closed
    local github_token="${GITHUB_TOKEN:-}"

    local api_url="https://api.github.com/repos/$repo/pulls?state=$state&sort=updated&direction=desc&per_page=100"
    local auth_header=""

    if [[ -n "$github_token" ]]; then
        auth_header="-H \"Authorization: token $github_token\""
    fi

    # Note: This gets the first 100 PRs. For more, implement pagination
    curl -s \
        -H "Accept: application/vnd.github.v3+json" \
        $auth_header \
        "$api_url" 2>/dev/null
}

# Get review participation metrics for a contributor
# Returns: reviews_given|reviews_received|prs_authored|review_response_time
get_contributor_review_metrics() {
    local repo="$1"
    local contributor_email="$2"
    local since_date="$3"
    local github_token="${GITHUB_TOKEN:-}"

    if [[ -z "$github_token" ]]; then
        # Cannot query API without token
        echo "0|0|0|0"
        return 1
    fi

    # Get GitHub username from email
    local username
    if ! username=$(lookup_github_username "$contributor_email"); then
        echo "0|0|0|0"
        return 1
    fi

    # Get all PRs in repo
    local prs=$(get_repo_pull_requests "$repo" "$since_date")

    if [[ -z "$prs" ]]; then
        echo "0|0|0|0"
        return 0
    fi

    # Count PRs authored by this user
    local prs_authored=$(echo "$prs" | jq -r --arg user "$username" '.[] | select(.user.login == $user) | .number' 2>/dev/null | wc -l | tr -d ' ')

    # Count reviews given by this user
    local reviews_given=0
    local pr_numbers=$(echo "$prs" | jq -r '.[].number' 2>/dev/null)

    while IFS= read -r pr_num; do
        if [[ -n "$pr_num" ]]; then
            local reviews=$(get_pr_reviews "$repo" "$pr_num")
            local user_reviews=$(echo "$reviews" | jq -r --arg user "$username" '.[] | select(.user.login == $user) | .id' 2>/dev/null | wc -l | tr -d ' ')
            reviews_given=$((reviews_given + user_reviews))
        fi
    done <<< "$pr_numbers"

    # Count reviews received (on their PRs)
    local reviews_received=0
    local user_pr_numbers=$(echo "$prs" | jq -r --arg user "$username" '.[] | select(.user.login == $user) | .number' 2>/dev/null)

    while IFS= read -r pr_num; do
        if [[ -n "$pr_num" ]]; then
            local reviews=$(get_pr_reviews "$repo" "$pr_num")
            local review_count=$(echo "$reviews" | jq -r 'length' 2>/dev/null || echo "0")
            reviews_received=$((reviews_received + review_count))
        fi
    done <<< "$user_pr_numbers"

    # Calculate average review response time (placeholder - would need detailed PR data)
    local avg_response_time=0

    echo "$reviews_given|$reviews_received|$prs_authored|$avg_response_time"
}

# Get review metrics for all contributors in a repository
# Creates a cache file with contributor metrics
cache_repository_review_metrics() {
    local repo="$1"
    local since_date="$2"
    local output_file="$3"
    local github_token="${GITHUB_TOKEN:-}"

    if [[ -z "$github_token" ]]; then
        echo "Error: GITHUB_TOKEN required for review metrics" >&2
        return 1
    fi

    # Extract owner/repo from URL or path
    local repo_slug
    if [[ "$repo" =~ github\.com[/:]([^/]+/[^/]+) ]]; then
        repo_slug="${BASH_REMATCH[1]}"
        repo_slug="${repo_slug%.git}"
    else
        repo_slug="$repo"
    fi

    # Get all PRs
    local prs=$(get_repo_pull_requests "$repo_slug" "$since_date")

    if [[ -z "$prs" ]]; then
        return 0
    fi

    # Extract unique contributors
    local contributors=$(echo "$prs" | jq -r '.[] | .user.login' 2>/dev/null | sort -u)

    # Get metrics for each contributor
    while IFS= read -r username; do
        if [[ -n "$username" ]]; then
            # Get user's email (would need separate API call)
            # For now, store by username
            local metrics=$(get_contributor_review_metrics "$repo_slug" "$username@users.noreply.github.com" "$since_date")
            echo "$username|$metrics" >> "$output_file"
        fi
    done <<< "$contributors"
}

# Calculate review participation score from cached metrics
calculate_review_score_from_cache() {
    local contributor_email="$1"
    local metrics_cache_file="$2"

    if [[ ! -f "$metrics_cache_file" ]]; then
        echo "50"  # Default score if no cache
        return
    fi

    # Get username from email
    local username
    if ! username=$(lookup_github_username "$contributor_email"); then
        echo "50"
        return
    fi

    # Find metrics in cache
    local metrics=$(grep "^$username|" "$metrics_cache_file" 2>/dev/null | head -1)

    if [[ -z "$metrics" ]]; then
        echo "50"
        return
    fi

    # Parse metrics: username|reviews_given|reviews_received|prs_authored|response_time
    IFS='|' read -r _ reviews_given reviews_received prs_authored _ <<< "$metrics"

    # Get max values from cache for normalization
    local max_given=$(awk -F'|' '{print $2}' "$metrics_cache_file" | sort -rn | head -1)
    local max_received=$(awk -F'|' '{print $3}' "$metrics_cache_file" | sort -rn | head -1)

    max_given="${max_given:-1}"
    max_received="${max_received:-1}"

    # Calculate score (0-100)
    local given_score=$(echo "scale=2; ($reviews_given / $max_given) * 50" | bc -l)
    local received_score=$(echo "scale=2; ($reviews_received / $max_received) * 50" | bc -l)

    echo "scale=0; $given_score + $received_score" | bc -l
}

# Check if GitHub token is configured
has_github_token() {
    [[ -n "${GITHUB_TOKEN:-}" ]]
}

# Get GitHub API rate limit status
get_rate_limit() {
    local github_token="${GITHUB_TOKEN:-}"
    local auth_header=""

    if [[ -n "$github_token" ]]; then
        auth_header="-H \"Authorization: token $github_token\""
    fi

    curl -s \
        -H "Accept: application/vnd.github.v3+json" \
        $auth_header \
        "https://api.github.com/rate_limit" 2>/dev/null
}

# Format contributor with GitHub profile info
format_contributor_with_github() {
    local name="$1"
    local email="$2"
    local include_profile="${3:-false}"

    local profile_info
    if profile_info=$(get_github_profile "$email"); then
        local username=$(echo "$profile_info" | cut -d'|' -f1)
        local profile_url=$(echo "$profile_info" | cut -d'|' -f2)

        if [[ "$include_profile" == "true" ]]; then
            echo "$name (@$username) - $profile_url"
        else
            echo "$name (@$username)"
        fi
    else
        echo "$name"
    fi
}

# Clone a GitHub repository with proper error handling and progress display
# Usage: github_clone_repository "url" "destination_dir" [--depth N] [--quiet]
# Returns: 0 on success, 1 on failure
github_clone_repository() {
    local repo_url="$1"
    local dest_dir="$2"
    local dest_provided=true
    local depth=""
    local quiet=""
    local show_progress=true

    # If no destination provided, create temp directory
    if [[ -z "$dest_dir" ]]; then
        dest_dir="$(mktemp -d)"
        dest_provided=false
    fi

    # Parse optional arguments
    shift 2 2>/dev/null || shift $#
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --depth)
                depth="--depth $2"
                shift 2
                ;;
            --quiet)
                quiet="--quiet"
                show_progress=false
                shift
                ;;
            *)
                shift
                ;;
        esac
    done

    # Normalize URL format
    if [[ ! "$repo_url" =~ ^https?:// ]] && [[ ! "$repo_url" =~ ^git@ ]]; then
        # Assume it's in owner/repo format, convert to HTTPS
        repo_url="https://github.com/$repo_url"
    fi

    # Show progress
    if [[ "$show_progress" == true ]]; then
        echo "Cloning repository from $repo_url..."
    fi

    # Clone repository
    if git clone $depth $quiet "$repo_url" "$dest_dir" 2>&1; then
        if [[ "$show_progress" == true ]]; then
            echo "✓ Repository cloned successfully to: $dest_dir"
        fi
        # Only echo path if it was auto-generated (for programmatic use)
        if [[ "$dest_provided" == false ]]; then
            echo "$dest_dir"
        fi
        return 0
    else
        echo "Error: Failed to clone repository from $repo_url" >&2
        echo "Note: For private repositories, ensure you have proper SSH keys or authentication" >&2
        echo "      Large repositories may take longer - consider using --depth 1 for shallow clone" >&2
        return 1
    fi
}

#############################################################################
# Organization Scanning Functions (merged from org-scanner.sh)
#############################################################################

# Colors (may be overridden by sourcing script)
: ${RED:='\033[0;31m'}
: ${GREEN:='\033[0;32m'}
: ${YELLOW:='\033[1;33m'}
: ${BLUE:='\033[0;34m'}
: ${CYAN:='\033[0;36m'}
: ${NC:='\033[0m'}

# Collect repos from org, separating cloned from uncloned
# Sets: REPOS_TO_SCAN (array), REPOS_NOT_CLONED (array)
# Args: $1 = org path
collect_org_repos() {
    local org_path="$1"
    REPOS_TO_SCAN=()
    REPOS_NOT_CLONED=()

    for repo_dir in "$org_path"/*/; do
        local repo_name=$(basename "$repo_dir")
        if [[ -d "$repo_dir/repo" ]]; then
            REPOS_TO_SCAN+=("$repo_name")
        else
            REPOS_NOT_CLONED+=("$repo_name")
        fi
    done
}

# Prompt user to hydrate uncloned repos
# Args: $1 = org name, $2 = org path, $3 = repo root path (for hydrate.sh)
# Returns: 0 if user wants to continue, 1 if cancelled
prompt_hydrate_repos() {
    local org="$1"
    local org_path="$2"
    local repo_root="$3"

    if [[ ${#REPOS_NOT_CLONED[@]} -eq 0 ]]; then
        return 0
    fi

    echo -e "${YELLOW}Found ${#REPOS_NOT_CLONED[@]} repositories without cloned code:${NC}" >&2
    for repo in "${REPOS_NOT_CLONED[@]}"; do
        echo -e "  - $repo" >&2
    done
    echo "" >&2

    # Prompt user
    read -p "Would you like to hydrate these repos for analysis? [y/N] " -n 1 -r >&2
    echo "" >&2

    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${BLUE}Hydrating ${#REPOS_NOT_CLONED[@]} repositories...${NC}" >&2

        # Run hydration for each uncloned repo
        for repo in "${REPOS_NOT_CLONED[@]}"; do
            echo -e "${CYAN}Cloning $org/$repo...${NC}" >&2
            "$repo_root/utils/phantom/hydrate.sh" --repo "$org/$repo" --quick >&2 2>&1 || true

            # Check if clone succeeded
            if [[ -d "$org_path/$repo/repo" ]]; then
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

    return 0
}

# Export functions
export -f init_github_cache
export -f cleanup_github_cache
export -f extract_username_from_noreply
export -f extract_username_from_github_email
export -f query_github_api_for_user
export -f lookup_github_username
export -f get_github_profile
export -f get_repo_info
export -f list_org_repos
export -f parse_repo_slug
export -f is_github_url
export -f get_pr_reviews
export -f get_repo_pull_requests
export -f get_contributor_review_metrics
export -f cache_repository_review_metrics
export -f calculate_review_score_from_cache
export -f has_github_token
export -f get_rate_limit
export -f format_contributor_with_github
export -f github_clone_repository
export -f collect_org_repos
export -f prompt_hydrate_repos
