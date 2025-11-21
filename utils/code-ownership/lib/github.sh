#!/bin/bash
# Copyright (c) 2024 Crash Override Inc
# 101 Fulton St, 416, New York 10038
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# GitHub Integration Library
# Enhanced GitHub profile mapping and API integration
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

# List repositories in organization
list_org_repos() {
    local org="$1"
    local github_token="${GITHUB_TOKEN:-}"

    local api_url="https://api.github.com/orgs/$org/repos?per_page=100"
    local auth_header=""

    if [[ -n "$github_token" ]]; then
        auth_header="-H \"Authorization: token $github_token\""
    fi

    # Note: This only gets first 100 repos. For more, implement pagination
    curl -s \
        -H "Accept: application/vnd.github.v3+json" \
        $auth_header \
        "$api_url" 2>/dev/null | jq -r '.[].full_name' 2>/dev/null
}

# Get pull request reviews for a repo
# This would require PR numbers, so it's a placeholder for future enhancement
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
export -f has_github_token
export -f get_rate_limit
export -f format_contributor_with_github
