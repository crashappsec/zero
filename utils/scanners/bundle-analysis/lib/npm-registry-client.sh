#!/usr/bin/env bash
# Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0
#
# NPM Registry API Client
# Queries registry.npmjs.org for package metadata
#
# Usage:
#   source npm-registry-client.sh
#   get_npm_metadata "react"
#   get_npm_metadata "lodash" "4.17.21"

set -euo pipefail

# Configuration
NPM_REGISTRY_API="https://registry.npmjs.org"
NPM_RATE_LIMIT_MS=50  # NPM is more tolerant
NPM_RETRY_COUNT=3

# Cache directory (optional, set externally)
NPM_CACHE_DIR="${NPM_CACHE_DIR:-}"

# Last request timestamp for rate limiting
_npm_last_request=0

# Rate limit helper
_npm_rate_limit() {
    local now_ms
    now_ms=$(date +%s%3N 2>/dev/null || echo "0")

    if [[ "$_npm_last_request" -gt 0 ]]; then
        local elapsed=$((now_ms - _npm_last_request))
        if [[ "$elapsed" -lt "$NPM_RATE_LIMIT_MS" ]]; then
            local sleep_ms=$((NPM_RATE_LIMIT_MS - elapsed))
            sleep "0.$(printf '%03d' $sleep_ms)" 2>/dev/null || sleep 0.05
        fi
    fi

    _npm_last_request=$(date +%s%3N 2>/dev/null || echo "0")
}

# Get cache key for a package
_npm_cache_key() {
    local package="$1"
    local version="${2:-latest}"
    echo "npm_${package}_${version}" | tr '/' '_' | tr '@' '_'
}

# Check cache for package data
_npm_cache_get() {
    local package="$1"
    local version="${2:-}"

    if [[ -z "$NPM_CACHE_DIR" ]] || [[ ! -d "$NPM_CACHE_DIR" ]]; then
        return 1
    fi

    local cache_key=$(_npm_cache_key "$package" "$version")
    local cache_file="$NPM_CACHE_DIR/$cache_key.json"

    if [[ -f "$cache_file" ]]; then
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
_npm_cache_set() {
    local package="$1"
    local version="${2:-}"
    local data="$3"

    if [[ -z "$NPM_CACHE_DIR" ]]; then
        return 0
    fi

    mkdir -p "$NPM_CACHE_DIR" 2>/dev/null || true

    local cache_key=$(_npm_cache_key "$package" "$version")
    local cache_file="$NPM_CACHE_DIR/$cache_key.json"

    echo "$data" > "$cache_file" 2>/dev/null || true
}

# Get full package metadata from npm registry
# Usage: get_npm_metadata "package-name" ["version"]
# Returns JSON with package.json fields for the specified version
get_npm_metadata() {
    local package="$1"
    local version="${2:-}"

    # Check cache first
    local cached
    if cached=$(_npm_cache_get "$package" "$version"); then
        echo "$cached"
        return 0
    fi

    # Rate limit
    _npm_rate_limit

    # Query registry
    local url="$NPM_REGISTRY_API/$package"
    local response
    local http_code

    local attempt=0
    while [[ "$attempt" -lt "$NPM_RETRY_COUNT" ]]; do
        attempt=$((attempt + 1))

        response=$(curl -s -w "\n%{http_code}" "$url" 2>/dev/null || echo -e "\n000")
        http_code=$(echo "$response" | tail -n1)
        response=$(echo "$response" | sed '$d')

        case "$http_code" in
            200)
                # Extract version-specific data or latest
                local version_data
                if [[ -n "$version" ]]; then
                    version_data=$(echo "$response" | jq --arg v "$version" '.versions[$v] // empty' 2>/dev/null)
                else
                    # Get latest version
                    local latest_version
                    latest_version=$(echo "$response" | jq -r '."dist-tags".latest // empty' 2>/dev/null)
                    if [[ -n "$latest_version" ]]; then
                        version_data=$(echo "$response" | jq --arg v "$latest_version" '.versions[$v] // empty' 2>/dev/null)
                    fi
                fi

                if [[ -n "$version_data" ]] && [[ "$version_data" != "null" ]]; then
                    _npm_cache_set "$package" "$version" "$version_data"
                    echo "$version_data"
                    return 0
                else
                    echo '{"error": "version_not_found", "package": "'"$package"'", "version": "'"$version"'"}'
                    return 1
                fi
                ;;
            404)
                echo '{"error": "not_found", "package": "'"$package"'"}'
                return 1
                ;;
            *)
                if [[ "$attempt" -lt "$NPM_RETRY_COUNT" ]]; then
                    sleep 1
                fi
                ;;
        esac
    done

    echo '{"error": "api_error", "package": "'"$package"'", "http_code": '"$http_code"'}'
    return 1
}

# Extract ESM-related fields from package metadata
# Usage: get_esm_info "package-name" ["version"]
# Returns JSON with module, exports, type, sideEffects fields
get_esm_info() {
    local package="$1"
    local version="${2:-}"

    local metadata
    metadata=$(get_npm_metadata "$package" "$version")

    if echo "$metadata" | jq -e '.error' &>/dev/null; then
        echo "$metadata"
        return 1
    fi

    # Extract relevant fields
    echo "$metadata" | jq '{
        name: .name,
        version: .version,
        module: .module,
        main: .main,
        type: .type,
        exports: .exports,
        sideEffects: .sideEffects,
        browser: .browser,
        unpkg: .unpkg,
        jsdelivr: .jsdelivr
    }' 2>/dev/null || echo '{"error": "parse_error"}'
}

# Check if package has ESM support
# Usage: has_esm_support "package-name" ["version"]
# Returns: 0 (true) if ESM, 1 (false) if not
has_esm_support() {
    local package="$1"
    local version="${2:-}"

    local esm_info
    esm_info=$(get_esm_info "$package" "$version")

    # Check for ESM indicators
    local has_module has_exports has_type_module

    has_module=$(echo "$esm_info" | jq -r '.module // empty' 2>/dev/null)
    has_type_module=$(echo "$esm_info" | jq -r 'select(.type == "module") | .type' 2>/dev/null)
    has_exports=$(echo "$esm_info" | jq -r '.exports // empty' 2>/dev/null)

    # ESM if any of these are present
    if [[ -n "$has_module" ]] || [[ -n "$has_type_module" ]] || [[ -n "$has_exports" && "$has_exports" != "null" ]]; then
        return 0
    fi

    return 1
}

# Get sideEffects value
# Usage: get_side_effects "package-name" ["version"]
# Returns: "true", "false", or array of files, or "unknown"
get_side_effects() {
    local package="$1"
    local version="${2:-}"

    local metadata
    metadata=$(get_npm_metadata "$package" "$version")

    local side_effects
    side_effects=$(echo "$metadata" | jq -r '.sideEffects // "unknown"' 2>/dev/null)

    echo "$side_effects"
}

# Check if package is tree-shakeable based on npm metadata
# Usage: is_tree_shakeable "package-name" ["version"]
# Returns: 0 (true) if tree-shakeable, 1 (false) if not
is_tree_shakeable() {
    local package="$1"
    local version="${2:-}"

    # Must have ESM support
    if ! has_esm_support "$package" "$version"; then
        return 1
    fi

    # Check sideEffects - if explicitly false, definitely tree-shakeable
    local side_effects
    side_effects=$(get_side_effects "$package" "$version")

    case "$side_effects" in
        "false")
            return 0
            ;;
        "true")
            return 1
            ;;
        "unknown")
            # Has ESM but no sideEffects declaration - partially tree-shakeable
            return 0
            ;;
        *)
            # Array of files with side effects - partially tree-shakeable
            return 0
            ;;
    esac
}

# Export functions
export -f get_npm_metadata
export -f get_esm_info
export -f has_esm_support
export -f get_side_effects
export -f is_tree_shakeable
export -f _npm_rate_limit
export -f _npm_cache_get
export -f _npm_cache_set
export -f _npm_cache_key
