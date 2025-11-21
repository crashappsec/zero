#!/bin/bash
# deps.dev API Client Library
# Copyright (c) 2024 Crash Override Inc
# SPDX-License-Identifier: GPL-3.0

set -euo pipefail

# Get script directory for loading shared libraries
LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
UTILS_ROOT="$(cd "$LIB_DIR/../.." && pwd)"

# Load configuration
if [ -f "$UTILS_ROOT/lib/config-loader.sh" ]; then
    source "$UTILS_ROOT/lib/config-loader.sh"
    CONFIG=$(load_config "package-health-analysis")
else
    CONFIG="{}"
fi

# Configuration
DEPS_DEV_BASE_URL=$(echo "$CONFIG" | jq -r '.package_health.api.deps_dev_base_url // "https://api.deps.dev/v3alpha"')
API_TIMEOUT=$(echo "$CONFIG" | jq -r '.package_health.api.timeout // 30')
RETRY_ATTEMPTS=$(echo "$CONFIG" | jq -r '.package_health.api.retry_attempts // 3')
RATE_LIMIT_DELAY=$(echo "$CONFIG" | jq -r '.package_health.api.rate_limit_delay // 1')

# Cache configuration
CACHE_ENABLED=$(echo "$CONFIG" | jq -r '.package_health.cache.enabled // true')
CACHE_TTL_HOURS=$(echo "$CONFIG" | jq -r '.package_health.cache.ttl_hours // 24')
CACHE_DIR=$(echo "$CONFIG" | jq -r '.package_health.cache.cache_dir // "/tmp/package-health-cache"')

# Initialize cache directory
init_cache() {
    if [ "$CACHE_ENABLED" = "true" ]; then
        mkdir -p "$CACHE_DIR"
    fi
}

# Generate cache key
cache_key() {
    local endpoint=$1
    echo -n "$endpoint" | md5sum | cut -d' ' -f1
}

# Get from cache
get_from_cache() {
    local endpoint=$1
    local key=$(cache_key "$endpoint")
    local cache_file="$CACHE_DIR/$key.json"

    if [ "$CACHE_ENABLED" = "false" ]; then
        return 1
    fi

    if [ -f "$cache_file" ]; then
        # Check if cache is still valid
        local cache_age=$(( $(date +%s) - $(stat -f %m "$cache_file" 2>/dev/null || stat -c %Y "$cache_file" 2>/dev/null || echo 0) ))
        local cache_max_age=$(( CACHE_TTL_HOURS * 3600 ))

        if [ "$cache_age" -lt "$cache_max_age" ]; then
            cat "$cache_file"
            return 0
        fi
    fi

    return 1
}

# Save to cache
save_to_cache() {
    local endpoint=$1
    local data=$2
    local key=$(cache_key "$endpoint")
    local cache_file="$CACHE_DIR/$key.json"

    if [ "$CACHE_ENABLED" = "true" ]; then
        echo "$data" > "$cache_file"
    fi
}

# Make API request with retry logic
api_request() {
    local url=$1
    local attempt=1

    # Check cache first
    local cached_data
    if cached_data=$(get_from_cache "$url"); then
        echo "$cached_data"
        return 0
    fi

    while [ $attempt -le $RETRY_ATTEMPTS ]; do
        local response
        local http_code

        # Make request
        response=$(curl -s -w "\n%{http_code}" --max-time "$API_TIMEOUT" "$url" 2>/dev/null || echo -e "\n000")
        http_code=$(echo "$response" | tail -n1)
        local body=$(echo "$response" | sed '$d')

        case $http_code in
            200)
                # Success - validate JSON before caching
                if echo "$body" | jq empty 2>/dev/null; then
                    save_to_cache "$url" "$body"
                    echo "$body"
                    return 0
                else
                    # Invalid JSON response
                    echo '{"error": "invalid_json_response"}'
                    return 1
                fi
                ;;
            429)
                # Rate limited - wait and retry
                if [ $attempt -lt $RETRY_ATTEMPTS ]; then
                    sleep $((RATE_LIMIT_DELAY * attempt))
                    ((attempt++))
                    continue
                fi
                ;;
            404)
                # Not found - return empty but valid JSON
                echo '{"error": "not_found"}'
                return 0
                ;;
            000)
                # Network error - retry
                if [ $attempt -lt $RETRY_ATTEMPTS ]; then
                    sleep 2
                    ((attempt++))
                    continue
                fi
                ;;
        esac

        ((attempt++))
    done

    # All retries failed - return valid JSON error to stdout
    echo '{"error": "api_request_failed"}'
    return 1
}

# Get package information
# Usage: get_package_info <system> <package>
get_package_info() {
    local system=$1
    local package=$2

    # URL encode package name
    local encoded_package=$(echo "$package" | jq -sRr @uri)
    local url="${DEPS_DEV_BASE_URL}/systems/${system}/packages/${encoded_package}"

    api_request "$url"
}

# Get package version information
# Usage: get_package_version <system> <package> <version>
get_package_version() {
    local system=$1
    local package=$2
    local version=$3

    # URL encode package name and version
    local encoded_package=$(echo "$package" | jq -sRr @uri)
    local encoded_version=$(echo "$version" | jq -sRr @uri)
    local url="${DEPS_DEV_BASE_URL}/systems/${system}/packages/${encoded_package}/versions/${encoded_version}"

    api_request "$url"
}

# Get project information (for OpenSSF Scorecard)
# Usage: get_project_info <project_key>
get_project_info() {
    local project_key=$1

    # URL encode project key
    local encoded_key=$(echo "$project_key" | jq -sRr @uri)
    local url="${DEPS_DEV_BASE_URL}/projects/${encoded_key}"

    api_request "$url"
}

# Extract OpenSSF Scorecard from package info
# Usage: extract_openssf_score <package_json>
extract_openssf_score() {
    local package_json=$1

    echo "$package_json" | jq -r '
        .scorecard.score // null
    '
}

# Extract OpenSSF Scorecard checks
# Usage: extract_openssf_checks <package_json>
extract_openssf_checks() {
    local package_json=$1

    echo "$package_json" | jq -r '
        .scorecard.checks // []
    '
}

# Check if package is deprecated
# Usage: check_deprecation <package_json>
check_deprecation() {
    local package_json=$1

    echo "$package_json" | jq -r '
        .deprecated // false
    '
}

# Get deprecation message
# Usage: get_deprecation_message <package_json>
get_deprecation_message() {
    local package_json=$1

    echo "$package_json" | jq -r '
        .deprecationMessage // ""
    '
}

# Get package licenses
# Usage: get_licenses <version_json>
get_licenses() {
    local version_json=$1

    echo "$version_json" | jq -r '
        .licenses // []
    '
}

# Get package dependencies
# Usage: get_dependencies <version_json>
get_dependencies() {
    local version_json=$1

    echo "$version_json" | jq -r '
        .dependencies // []
    '
}

# Get known vulnerabilities
# Usage: get_vulnerabilities <version_json>
get_vulnerabilities() {
    local version_json=$1

    echo "$version_json" | jq -r '
        .advisories // []
    '
}

# Get package popularity metrics
# Usage: get_popularity_metrics <package_json>
get_popularity_metrics() {
    local package_json=$1

    echo "$package_json" | jq -r '{
        dependent_count: (.dependentCount // 0),
        dependent_repos_count: (.dependentReposCount // 0)
    }'
}

# Get latest version
# Usage: get_latest_version <package_json>
get_latest_version() {
    local package_json=$1

    echo "$package_json" | jq -r '
        .versions[-1].versionKey.version // "unknown"
    '
}

# Get all versions
# Usage: get_all_versions <package_json>
get_all_versions() {
    local package_json=$1

    echo "$package_json" | jq -r '
        [.versions[].versionKey.version] // []
    '
}

# Get package metadata summary
# Usage: get_package_summary <system> <package>
get_package_summary() {
    local system=$1
    local package=$2

    local package_info=$(get_package_info "$system" "$package")

    if [ $? -ne 0 ]; then
        echo '{"error": "failed_to_fetch_package"}'
        return 1
    fi

    # Validate package_info is valid JSON
    if ! echo "$package_info" | jq empty 2>/dev/null; then
        echo '{"error": "invalid_response"}'
        return 1
    fi

    # Check for error in response
    if echo "$package_info" | jq -e '.error' > /dev/null 2>&1; then
        echo "$package_info"
        return 0
    fi

    # Extract summary information safely
    local summary=$(echo "$package_info" | jq -r '{
        name: (.packageKey.name // "unknown"),
        system: (.packageKey.system // "unknown"),
        deprecated: (.deprecated // false),
        deprecation_message: (.deprecationMessage // ""),
        latest_version: (.versions[-1].versionKey.version // "unknown"),
        openssf_score: (.scorecard.score // null),
        openssf_date: (.scorecard.date // null),
        dependent_count: (.dependentCount // 0),
        dependent_repos_count: (.dependentReposCount // 0),
        project_url: (.projectKey // null)
    }' 2>/dev/null)

    # Validate output
    if echo "$summary" | jq empty 2>/dev/null; then
        echo "$summary"
    else
        echo '{"error": "failed_to_parse_response"}'
        return 1
    fi
}

# Initialize cache on load
init_cache

# Export functions
export -f get_package_info
export -f get_package_version
export -f get_project_info
export -f extract_openssf_score
export -f extract_openssf_checks
export -f check_deprecation
export -f get_deprecation_message
export -f get_licenses
export -f get_dependencies
export -f get_vulnerabilities
export -f get_popularity_metrics
export -f get_latest_version
export -f get_all_versions
export -f get_package_summary
