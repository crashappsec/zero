#!/bin/bash
# deps.dev API Client Library
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0
#
# Shared library for interacting with the deps.dev API.
# Provides package information, vulnerability data, and OpenSSF Scorecard metrics.
#
# Usage: source this file from any scanner that needs deps.dev API access
#   source "$SCANNERS_ROOT/shared/lib/deps-dev-client.sh"

set -eo pipefail

# Get script directory for loading shared libraries
DEPS_DEV_LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SHARED_DIR="$(dirname "$DEPS_DEV_LIB_DIR")"
SCANNERS_ROOT="$(dirname "$SHARED_DIR")"
UTILS_ROOT="$(dirname "$SCANNERS_ROOT")"

# Load configuration if available
DEPS_DEV_CONFIG="{}"
if [ -f "$UTILS_ROOT/lib/config-loader.sh" ]; then
    source "$UTILS_ROOT/lib/config-loader.sh"
    # Try multiple config names for backward compatibility
    DEPS_DEV_CONFIG=$(load_config "deps-dev" 2>/dev/null || load_config "package-health-analysis" 2>/dev/null || load_config "supply-chain" 2>/dev/null || echo "{}")
fi

# Configuration with defaults - support multiple config path formats
_get_config_value() {
    local key=$1
    local default=$2
    local value

    # Try different config path patterns
    value=$(echo "$DEPS_DEV_CONFIG" | jq -r ".api.${key} // .package_health.api.${key} // null" 2>/dev/null)

    if [[ -z "$value" || "$value" == "null" ]]; then
        echo "$default"
    else
        echo "$value"
    fi
}

_get_cache_config_value() {
    local key=$1
    local default=$2
    local value

    # Try different config path patterns
    value=$(echo "$DEPS_DEV_CONFIG" | jq -r ".cache.${key} // .package_health.cache.${key} // null" 2>/dev/null)

    if [[ -z "$value" || "$value" == "null" ]]; then
        echo "$default"
    else
        echo "$value"
    fi
}

# Configuration with environment variable overrides
DEPS_DEV_BASE_URL="${DEPS_DEV_BASE_URL:-$(_get_config_value "deps_dev_base_url" "https://api.deps.dev/v3alpha")}"
DEPS_DEV_API_TIMEOUT="${DEPS_DEV_API_TIMEOUT:-$(_get_config_value "timeout" "30")}"
DEPS_DEV_RETRY_ATTEMPTS="${DEPS_DEV_RETRY_ATTEMPTS:-$(_get_config_value "retry_attempts" "3")}"
DEPS_DEV_RATE_LIMIT_DELAY="${DEPS_DEV_RATE_LIMIT_DELAY:-$(_get_config_value "rate_limit_delay" "1")}"

# Cache configuration
DEPS_DEV_CACHE_ENABLED="${DEPS_DEV_CACHE_ENABLED:-$(_get_cache_config_value "enabled" "true")}"
DEPS_DEV_CACHE_TTL_HOURS="${DEPS_DEV_CACHE_TTL_HOURS:-$(_get_cache_config_value "ttl_hours" "24")}"
DEPS_DEV_CACHE_DIR="${DEPS_DEV_CACHE_DIR:-$(_get_cache_config_value "cache_dir" "/tmp/deps-dev-cache")}"

# Initialize cache directory
deps_dev_init_cache() {
    if [ "$DEPS_DEV_CACHE_ENABLED" = "true" ]; then
        mkdir -p "$DEPS_DEV_CACHE_DIR"
    fi
}

# Generate cache key
_deps_dev_cache_key() {
    local endpoint=$1
    echo -n "$endpoint" | md5sum | cut -d' ' -f1
}

# Get from cache
_deps_dev_get_from_cache() {
    local endpoint=$1
    local key=$(_deps_dev_cache_key "$endpoint")
    local cache_file="$DEPS_DEV_CACHE_DIR/$key.json"

    if [ "$DEPS_DEV_CACHE_ENABLED" = "false" ]; then
        return 1
    fi

    if [ -f "$cache_file" ]; then
        # Check if cache is still valid (cross-platform stat)
        local cache_age=$(( $(date +%s) - $(stat -f %m "$cache_file" 2>/dev/null || stat -c %Y "$cache_file" 2>/dev/null || echo 0) ))
        local cache_max_age=$(( DEPS_DEV_CACHE_TTL_HOURS * 3600 ))

        if [ "$cache_age" -lt "$cache_max_age" ]; then
            cat "$cache_file"
            return 0
        fi
    fi

    return 1
}

# Save to cache
_deps_dev_save_to_cache() {
    local endpoint=$1
    local data=$2
    local key=$(_deps_dev_cache_key "$endpoint")
    local cache_file="$DEPS_DEV_CACHE_DIR/$key.json"

    if [ "$DEPS_DEV_CACHE_ENABLED" = "true" ]; then
        echo "$data" > "$cache_file"
    fi
}

# Make API request with retry logic
deps_dev_api_request() {
    local url=$1
    local attempt=1

    # Check cache first
    local cached_data
    if cached_data=$(_deps_dev_get_from_cache "$url"); then
        echo "$cached_data"
        return 0
    fi

    while [ "$attempt" -le "${DEPS_DEV_RETRY_ATTEMPTS:-3}" ]; do
        local response
        local http_code

        # Make request
        response=$(curl -s -w "\n%{http_code}" --max-time "$DEPS_DEV_API_TIMEOUT" "$url" 2>/dev/null || echo -e "\n000")
        http_code=$(echo "$response" | tail -n1)
        local body=$(echo "$response" | sed '$d')

        case $http_code in
            200)
                # Success - validate JSON before caching
                if echo "$body" | jq empty 2>/dev/null; then
                    _deps_dev_save_to_cache "$url" "$body"
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
                if [ $attempt -lt $DEPS_DEV_RETRY_ATTEMPTS ]; then
                    sleep $((DEPS_DEV_RATE_LIMIT_DELAY * attempt))
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
                if [ $attempt -lt $DEPS_DEV_RETRY_ATTEMPTS ]; then
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
# Usage: deps_dev_get_package_info <system> <package>
deps_dev_get_package_info() {
    local system=$1
    local package=$2

    # URL encode package name (use printf to avoid newline)
    local encoded_package=$(printf '%s' "$package" | jq -sRr @uri)
    local url="${DEPS_DEV_BASE_URL}/systems/${system}/packages/${encoded_package}"

    deps_dev_api_request "$url"
}

# Get package version information
# Usage: deps_dev_get_package_version <system> <package> <version>
deps_dev_get_package_version() {
    local system=$1
    local package=$2
    local version=$3

    # URL encode package name and version (use printf to avoid newline)
    local encoded_package=$(printf '%s' "$package" | jq -sRr @uri)
    local encoded_version=$(printf '%s' "$version" | jq -sRr @uri)
    local url="${DEPS_DEV_BASE_URL}/systems/${system}/packages/${encoded_package}/versions/${encoded_version}"

    deps_dev_api_request "$url"
}

# Batch get package versions (up to 5000 packages)
# Usage: deps_dev_get_versions_batch <json_array_of_packages>
# Input format: [{"system": "npm", "name": "react", "version": "18.0.0"}, ...]
# Output format: {"responses": [{version data}, ...]}
deps_dev_get_versions_batch() {
    local packages_json=$1
    local batch_url="${DEPS_DEV_BASE_URL}/versionbatch"

    # Build batch request with proper format
    local batch_request=$(echo "$packages_json" | jq '{
        requests: [.[] | {
            versionKey: {
                system: (.system | ascii_upcase),
                name: .name,
                version: .version
            }
        }]
    }')

    # Make POST request
    local response=$(curl -s --max-time "$DEPS_DEV_API_TIMEOUT" \
        -H "Content-Type: application/json" \
        -d "$batch_request" \
        "$batch_url" 2>/dev/null || echo '{"error": "batch_request_failed"}')

    # Validate JSON
    if echo "$response" | jq empty 2>/dev/null; then
        echo "$response"
    else
        echo '{"error": "invalid_batch_response"}'
    fi
}

# Batch get package info (using individual requests in parallel batches)
# Usage: deps_dev_get_packages_batch <json_array_of_packages>
# Input format: [{"system": "npm", "name": "react"}, ...]
deps_dev_get_packages_batch() {
    local packages_json=$1
    local results="[]"

    # Process each package
    while IFS= read -r pkg; do
        local system=$(echo "$pkg" | jq -r '.system')
        local name=$(echo "$pkg" | jq -r '.name')

        local pkg_info=$(deps_dev_get_package_info "$system" "$name")
        results=$(echo "$results" | jq --argjson item "$pkg_info" '. + [$item]')
    done < <(echo "$packages_json" | jq -c '.[]')

    echo "$results"
}

# Get project information (for OpenSSF Scorecard)
# Usage: deps_dev_get_project_info <project_key>
deps_dev_get_project_info() {
    local project_key=$1

    # URL encode project key (use printf to avoid newline)
    local encoded_key=$(printf '%s' "$project_key" | jq -sRr @uri)
    local url="${DEPS_DEV_BASE_URL}/projects/${encoded_key}"

    deps_dev_api_request "$url"
}

# Extract OpenSSF Scorecard from package info
# Usage: deps_dev_extract_openssf_score <package_json>
deps_dev_extract_openssf_score() {
    local package_json=$1

    echo "$package_json" | jq -r '
        .scorecard.score // null
    '
}

# Extract OpenSSF Scorecard checks
# Usage: deps_dev_extract_openssf_checks <package_json>
deps_dev_extract_openssf_checks() {
    local package_json=$1

    echo "$package_json" | jq -r '
        .scorecard.checks // []
    '
}

# Check if package is deprecated
# Usage: deps_dev_check_deprecation <package_json>
deps_dev_check_deprecation() {
    local package_json=$1

    echo "$package_json" | jq -r '
        .deprecated // false
    '
}

# Get deprecation message
# Usage: deps_dev_get_deprecation_message <package_json>
deps_dev_get_deprecation_message() {
    local package_json=$1

    echo "$package_json" | jq -r '
        .deprecationMessage // ""
    '
}

# Get package licenses
# Usage: deps_dev_get_licenses <version_json>
deps_dev_get_licenses() {
    local version_json=$1

    echo "$version_json" | jq -r '
        .licenses // []
    '
}

# Get package dependencies
# Usage: deps_dev_get_dependencies <version_json>
deps_dev_get_dependencies() {
    local version_json=$1

    echo "$version_json" | jq -r '
        .dependencies // []
    '
}

# Get known vulnerabilities
# Usage: deps_dev_get_vulnerabilities <version_json>
deps_dev_get_vulnerabilities() {
    local version_json=$1

    echo "$version_json" | jq -r '
        .advisories // []
    '
}

# Get package popularity metrics
# Usage: deps_dev_get_popularity_metrics <package_json>
deps_dev_get_popularity_metrics() {
    local package_json=$1

    echo "$package_json" | jq -r '{
        dependent_count: (.dependentCount // 0),
        dependent_repos_count: (.dependentReposCount // 0)
    }'
}

# Get latest version
# Usage: deps_dev_get_latest_version <package_json>
deps_dev_get_latest_version() {
    local package_json=$1

    echo "$package_json" | jq -r '
        .versions[-1].versionKey.version // "unknown"
    '
}

# Get all versions
# Usage: deps_dev_get_all_versions <package_json>
deps_dev_get_all_versions() {
    local package_json=$1

    echo "$package_json" | jq -r '
        [.versions[].versionKey.version] // []
    '
}

# Get package metadata summary
# Usage: deps_dev_get_package_summary <system> <package>
deps_dev_get_package_summary() {
    local system=$1
    local package=$2

    local package_info=$(deps_dev_get_package_info "$system" "$package")

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

# Backward compatibility aliases (old function names)
# These allow existing code to work without changes
get_package_info() { deps_dev_get_package_info "$@"; }
get_package_version() { deps_dev_get_package_version "$@"; }
get_versions_batch() { deps_dev_get_versions_batch "$@"; }
get_packages_batch() { deps_dev_get_packages_batch "$@"; }
get_project_info() { deps_dev_get_project_info "$@"; }
extract_openssf_score() { deps_dev_extract_openssf_score "$@"; }
extract_openssf_checks() { deps_dev_extract_openssf_checks "$@"; }
check_deprecation() { deps_dev_check_deprecation "$@"; }
get_deprecation_message() { deps_dev_get_deprecation_message "$@"; }
get_licenses() { deps_dev_get_licenses "$@"; }
get_dependencies() { deps_dev_get_dependencies "$@"; }
get_vulnerabilities() { deps_dev_get_vulnerabilities "$@"; }
get_popularity_metrics() { deps_dev_get_popularity_metrics "$@"; }
get_latest_version() { deps_dev_get_latest_version "$@"; }
get_all_versions() { deps_dev_get_all_versions "$@"; }
get_package_summary() { deps_dev_get_package_summary "$@"; }
api_request() { deps_dev_api_request "$@"; }
init_cache() { deps_dev_init_cache "$@"; }

# Initialize cache on load
deps_dev_init_cache

# Export functions for subshell use
export -f deps_dev_get_package_info
export -f deps_dev_get_package_version
export -f deps_dev_get_versions_batch
export -f deps_dev_get_packages_batch
export -f deps_dev_get_project_info
export -f deps_dev_extract_openssf_score
export -f deps_dev_extract_openssf_checks
export -f deps_dev_check_deprecation
export -f deps_dev_get_deprecation_message
export -f deps_dev_get_licenses
export -f deps_dev_get_dependencies
export -f deps_dev_get_vulnerabilities
export -f deps_dev_get_popularity_metrics
export -f deps_dev_get_latest_version
export -f deps_dev_get_all_versions
export -f deps_dev_get_package_summary
export -f deps_dev_api_request
export -f deps_dev_init_cache

# Export backward compatibility aliases
export -f get_package_info
export -f get_package_version
export -f get_versions_batch
export -f get_packages_batch
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
export -f api_request
export -f init_cache
