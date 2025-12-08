#!/usr/bin/env bash
# Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0
#
# Bundlephobia API Client
# Queries bundlephobia.com for package bundle sizes
#
# Usage:
#   source bundlephobia-client.sh
#   get_bundle_size "react" "18.2.0"
#   get_bundle_size "lodash"  # Uses latest version

set -euo pipefail

# Configuration
BUNDLEPHOBIA_API="https://bundlephobia.com/api/size"
BUNDLEPHOBIA_RATE_LIMIT_MS=100  # Delay between requests
BUNDLEPHOBIA_RETRY_COUNT=3
BUNDLEPHOBIA_RETRY_DELAY_MS=1000

# Cache directory (optional, set externally)
BUNDLEPHOBIA_CACHE_DIR="${BUNDLEPHOBIA_CACHE_DIR:-}"

# Last request timestamp for rate limiting
_bundlephobia_last_request=0

# Rate limit helper - ensures minimum delay between requests
_bundlephobia_rate_limit() {
    local now_ms
    now_ms=$(date +%s%3N 2>/dev/null || echo "0")

    if [[ "$_bundlephobia_last_request" -gt 0 ]]; then
        local elapsed=$((now_ms - _bundlephobia_last_request))
        if [[ "$elapsed" -lt "$BUNDLEPHOBIA_RATE_LIMIT_MS" ]]; then
            local sleep_ms=$((BUNDLEPHOBIA_RATE_LIMIT_MS - elapsed))
            sleep "0.$(printf '%03d' $sleep_ms)" 2>/dev/null || sleep 0.1
        fi
    fi

    _bundlephobia_last_request=$(date +%s%3N 2>/dev/null || echo "0")
}

# Get cache key for a package
_bundlephobia_cache_key() {
    local package="$1"
    local version="${2:-latest}"
    echo "${package}@${version}" | tr '/' '_' | tr '@' '_'
}

# Check cache for package data
_bundlephobia_cache_get() {
    local package="$1"
    local version="${2:-}"

    if [[ -z "$BUNDLEPHOBIA_CACHE_DIR" ]] || [[ ! -d "$BUNDLEPHOBIA_CACHE_DIR" ]]; then
        return 1
    fi

    local cache_key=$(_bundlephobia_cache_key "$package" "$version")
    local cache_file="$BUNDLEPHOBIA_CACHE_DIR/$cache_key.json"

    if [[ -f "$cache_file" ]]; then
        # Check if cache is less than 24 hours old
        local now=$(date +%s)
        local file_time=$(stat -f %m "$cache_file" 2>/dev/null || stat -c %Y "$cache_file" 2>/dev/null || echo 0)
        local age=$((now - file_time))

        if [[ "$age" -lt 86400 ]]; then  # 24 hours
            cat "$cache_file"
            return 0
        fi
    fi

    return 1
}

# Save to cache
_bundlephobia_cache_set() {
    local package="$1"
    local version="${2:-}"
    local data="$3"

    if [[ -z "$BUNDLEPHOBIA_CACHE_DIR" ]]; then
        return 0
    fi

    mkdir -p "$BUNDLEPHOBIA_CACHE_DIR" 2>/dev/null || true

    local cache_key=$(_bundlephobia_cache_key "$package" "$version")
    local cache_file="$BUNDLEPHOBIA_CACHE_DIR/$cache_key.json"

    echo "$data" > "$cache_file" 2>/dev/null || true
}

# Query Bundlephobia API for a single package
# Usage: get_bundle_size "package-name" ["version"]
# Returns JSON: { name, version, size, gzip, dependencyCount, ... }
get_bundle_size() {
    local package="$1"
    local version="${2:-}"

    # Check cache first
    local cached
    if cached=$(_bundlephobia_cache_get "$package" "$version"); then
        echo "$cached"
        return 0
    fi

    # Build URL
    local url="$BUNDLEPHOBIA_API?package=$package"
    if [[ -n "$version" ]]; then
        url="$BUNDLEPHOBIA_API?package=${package}@${version}"
    fi

    # Rate limit
    _bundlephobia_rate_limit

    # Query with retry
    local attempt=0
    local response=""
    local http_code=""

    while [[ "$attempt" -lt "$BUNDLEPHOBIA_RETRY_COUNT" ]]; do
        attempt=$((attempt + 1))

        # Make request
        response=$(curl -s -w "\n%{http_code}" "$url" 2>/dev/null || echo -e "\n000")
        http_code=$(echo "$response" | tail -n1)
        response=$(echo "$response" | sed '$d')

        case "$http_code" in
            200)
                # Success - cache and return
                _bundlephobia_cache_set "$package" "$version" "$response"
                echo "$response"
                return 0
                ;;
            429)
                # Rate limited - wait and retry
                local delay_s=$((BUNDLEPHOBIA_RETRY_DELAY_MS * attempt / 1000))
                sleep "$delay_s"
                ;;
            404)
                # Package not found
                echo '{"error": "not_found", "package": "'"$package"'", "message": "Package not found on bundlephobia"}'
                return 1
                ;;
            *)
                # Other error - retry
                if [[ "$attempt" -lt "$BUNDLEPHOBIA_RETRY_COUNT" ]]; then
                    sleep 1
                fi
                ;;
        esac
    done

    # All retries failed
    echo '{"error": "api_error", "package": "'"$package"'", "http_code": '"$http_code"', "message": "Failed to fetch from bundlephobia after '"$BUNDLEPHOBIA_RETRY_COUNT"' attempts"}'
    return 1
}

# Get bundle sizes for multiple packages
# Usage: get_bundle_sizes_batch "package1@version" "package2" "package3@version"
# Returns JSON array of results
get_bundle_sizes_batch() {
    local packages=("$@")
    local results=()

    for pkg_spec in "${packages[@]}"; do
        local package="${pkg_spec%@*}"
        local version=""

        if [[ "$pkg_spec" == *"@"* ]]; then
            version="${pkg_spec##*@}"
        fi

        local result
        result=$(get_bundle_size "$package" "$version")
        results+=("$result")
    done

    # Combine into JSON array
    printf '%s\n' "${results[@]}" | jq -s '.' 2>/dev/null || echo '[]'
}

# Classify package size
# Usage: classify_bundle_size 67000
# Returns: minimal, small, medium, large, very_large
classify_bundle_size() {
    local gzip_size="$1"

    if [[ "$gzip_size" -lt 5000 ]]; then
        echo "minimal"
    elif [[ "$gzip_size" -lt 25000 ]]; then
        echo "small"
    elif [[ "$gzip_size" -lt 50000 ]]; then
        echo "medium"
    elif [[ "$gzip_size" -lt 100000 ]]; then
        echo "large"
    else
        echo "very_large"
    fi
}

# Check if package is considered "heavy"
# Usage: is_heavy_package 67000
# Returns: 0 (true) if heavy, 1 (false) if not
is_heavy_package() {
    local gzip_size="$1"
    local threshold="${2:-50000}"  # Default 50KB

    [[ "$gzip_size" -ge "$threshold" ]]
}

# Export functions for use in other scripts
export -f get_bundle_size
export -f get_bundle_sizes_batch
export -f classify_bundle_size
export -f is_heavy_package
export -f _bundlephobia_rate_limit
export -f _bundlephobia_cache_get
export -f _bundlephobia_cache_set
export -f _bundlephobia_cache_key
