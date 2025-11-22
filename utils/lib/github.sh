#!/bin/bash
# Copyright (c) 2024 Gibson Powers Contributors
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# GitHub Integration Library
# Provides functions for GitHub API operations
#############################################################################

# List repositories for an organization
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

# Get GitHub profile from email
# Usage: get_github_profile "email@example.com" "Full Name"
get_github_profile() {
    local email="$1"
    local name="$2"
    local github_token="${GITHUB_TOKEN:-}"

    # Try to find user by email via search API
    local url="https://api.github.com/search/users?q=$email+in:email"

    if [[ -n "$github_token" ]]; then
        response=$(curl -s -H "Authorization: token $github_token" "$url")
    else
        response=$(curl -s "$url")
    fi

    local username=$(echo "$response" | jq -r '.items[0].login // empty' 2>/dev/null)

    if [[ -n "$username" ]]; then
        echo "$username|https://github.com/$username"
    else
        echo "|"
    fi
}

# Export functions
export -f list_org_repos
export -f parse_repo_slug
export -f is_github_url
export -f get_github_profile
