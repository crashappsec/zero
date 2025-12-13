#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Package Download Library
#
# Shared library for downloading package artifacts from registries.
# Supports npm, PyPI, and other package ecosystems.
#
# Usage: source this file from any scanner that needs to download packages
#   source "$UTILS_ROOT/lib/package-download.sh"
#############################################################################

set -eo pipefail

# Get script directory
PACKAGE_DL_LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
UTILS_ROOT="$(dirname "$PACKAGE_DL_LIB_DIR")"

# Package cache configuration
PACKAGE_CACHE_DIR="${PACKAGE_CACHE_DIR:-$HOME/.cache/gibson/packages}"
PACKAGE_CACHE_TTL_HOURS="${PACKAGE_CACHE_TTL_HOURS:-168}"  # 7 days default

# Ensure cache directory exists
package_dl_init_cache() {
    mkdir -p "$PACKAGE_CACHE_DIR"/{npm,pypi,cargo,gem,go,maven}
}

# Get cache file path for a package
_package_dl_cache_path() {
    local ecosystem="$1"
    local name="$2"
    local version="$3"

    # Sanitize name for filesystem (replace @ and / with _)
    local safe_name="${name//@/_}"
    safe_name="${safe_name//\//_}"

    echo "$PACKAGE_CACHE_DIR/$ecosystem/${safe_name}-${version}"
}

# Check if cached package exists and is valid
_package_dl_check_cache() {
    local cache_path="$1"

    # Find any file matching the cache path prefix
    local cached_file=$(find "$cache_path"* -type f 2>/dev/null | head -1)

    if [[ -z "$cached_file" ]] || [[ ! -f "$cached_file" ]]; then
        return 1
    fi

    # Check age
    local cache_age=$(( $(date +%s) - $(stat -f %m "$cached_file" 2>/dev/null || stat -c %Y "$cached_file" 2>/dev/null || echo 0) ))
    local cache_max_age=$(( PACKAGE_CACHE_TTL_HOURS * 3600 ))

    if [[ "$cache_age" -lt "$cache_max_age" ]]; then
        echo "$cached_file"
        return 0
    fi

    return 1
}

# Download npm package tarball
# Usage: package_dl_npm <name> <version> [output_dir]
# Returns: path to downloaded file
package_dl_npm() {
    local name="$1"
    local version="$2"
    local output_dir="${3:-$PACKAGE_CACHE_DIR/npm}"

    mkdir -p "$output_dir"

    # Check cache first
    local cache_path=$(_package_dl_cache_path "npm" "$name" "$version")
    local cached_file
    if cached_file=$(_package_dl_check_cache "$cache_path"); then
        echo "$cached_file"
        return 0
    fi

    # Handle scoped packages (e.g., @babel/core)
    local encoded_name="${name//@/%40}"
    encoded_name="${encoded_name//\//%2F}"

    # Get package metadata from npm registry
    local package_info
    package_info=$(curl -sL "https://registry.npmjs.org/${encoded_name}/${version}" 2>/dev/null)

    if [[ -z "$package_info" ]] || echo "$package_info" | jq -e '.error' &>/dev/null; then
        echo "Error: Failed to get npm package info for ${name}@${version}" >&2
        return 1
    fi

    local tarball_url
    tarball_url=$(echo "$package_info" | jq -r '.dist.tarball // empty')

    if [[ -z "$tarball_url" ]]; then
        echo "Error: No tarball URL for ${name}@${version}" >&2
        return 1
    fi

    # Sanitize filename
    local safe_name="${name//@/_}"
    safe_name="${safe_name//\//_}"
    local output_file="$output_dir/${safe_name}-${version}.tgz"

    if curl -sL "$tarball_url" -o "$output_file" 2>/dev/null; then
        echo "$output_file"
        return 0
    else
        echo "Error: Failed to download ${name}@${version}" >&2
        return 1
    fi
}

# Download PyPI package
# Usage: package_dl_pypi <name> <version> [output_dir]
# Returns: path to downloaded file
package_dl_pypi() {
    local name="$1"
    local version="$2"
    local output_dir="${3:-$PACKAGE_CACHE_DIR/pypi}"

    mkdir -p "$output_dir"

    # Check cache first
    local cache_path=$(_package_dl_cache_path "pypi" "$name" "$version")
    local cached_file
    if cached_file=$(_package_dl_check_cache "$cache_path"); then
        echo "$cached_file"
        return 0
    fi

    # Get package metadata from PyPI
    local package_info
    package_info=$(curl -sL "https://pypi.org/pypi/${name}/${version}/json" 2>/dev/null)

    if [[ -z "$package_info" ]] || echo "$package_info" | jq -e '.message' &>/dev/null; then
        echo "Error: Failed to get PyPI package info for ${name}==${version}" >&2
        return 1
    fi

    # Prefer source distribution (.tar.gz), fallback to wheel
    local download_url
    download_url=$(echo "$package_info" | jq -r '
        .urls |
        (map(select(.packagetype == "sdist")) | first //
         map(select(.packagetype == "bdist_wheel")) | first) |
        .url // empty
    ')

    if [[ -z "$download_url" ]]; then
        echo "Error: No download URL for ${name}==${version}" >&2
        return 1
    fi

    local filename
    filename=$(basename "$download_url")
    local output_file="$output_dir/$filename"

    if curl -sL "$download_url" -o "$output_file" 2>/dev/null; then
        echo "$output_file"
        return 0
    else
        echo "Error: Failed to download ${name}==${version}" >&2
        return 1
    fi
}

# Download package based on ecosystem
# Usage: package_dl_download <ecosystem> <name> <version> [output_dir]
# Returns: path to downloaded file
package_dl_download() {
    local ecosystem="$1"
    local name="$2"
    local version="$3"
    local output_dir="${4:-$PACKAGE_CACHE_DIR/$ecosystem}"

    package_dl_init_cache

    case "$ecosystem" in
        npm)
            package_dl_npm "$name" "$version" "$output_dir"
            ;;
        pypi)
            package_dl_pypi "$name" "$version" "$output_dir"
            ;;
        cargo|gem|go|maven)
            echo "Warning: Package download not yet implemented for ecosystem: $ecosystem" >&2
            return 1
            ;;
        *)
            echo "Error: Unknown ecosystem: $ecosystem" >&2
            return 1
            ;;
    esac
}

# Parse ecosystem from purl
# Usage: package_dl_parse_ecosystem <purl>
# Returns: ecosystem name (npm, pypi, cargo, etc.)
package_dl_parse_ecosystem() {
    local purl="$1"

    if [[ "$purl" =~ ^pkg:npm/ ]]; then
        echo "npm"
    elif [[ "$purl" =~ ^pkg:pypi/ ]]; then
        echo "pypi"
    elif [[ "$purl" =~ ^pkg:cargo/ ]]; then
        echo "cargo"
    elif [[ "$purl" =~ ^pkg:gem/ ]]; then
        echo "gem"
    elif [[ "$purl" =~ ^pkg:golang/ ]]; then
        echo "go"
    elif [[ "$purl" =~ ^pkg:maven/ ]]; then
        echo "maven"
    else
        echo "unknown"
    fi
}

# Extract packages from CycloneDX SBOM
# Usage: package_dl_extract_from_sbom <sbom_file>
# Returns: newline-separated list of "name|version|ecosystem|purl"
package_dl_extract_from_sbom() {
    local sbom_file="$1"

    if [[ ! -f "$sbom_file" ]]; then
        echo "Error: SBOM file not found: $sbom_file" >&2
        return 1
    fi

    jq -r '
        .components[]? |
        select(.type == "library") |
        .purl as $purl |
        (
            if ($purl | test("^pkg:npm/")) then "npm"
            elif ($purl | test("^pkg:pypi/")) then "pypi"
            elif ($purl | test("^pkg:cargo/")) then "cargo"
            elif ($purl | test("^pkg:gem/")) then "gem"
            elif ($purl | test("^pkg:golang/")) then "go"
            elif ($purl | test("^pkg:maven/")) then "maven"
            else "unknown"
            end
        ) as $ecosystem |
        "\(.name)|\(.version // "latest")|\($ecosystem)|\($purl // "")"
    ' "$sbom_file" 2>/dev/null
}

# Clean up old cached packages
# Usage: package_dl_clean_cache [max_age_days]
package_dl_clean_cache() {
    local max_age_days="${1:-7}"

    if [[ -d "$PACKAGE_CACHE_DIR" ]]; then
        find "$PACKAGE_CACHE_DIR" -type f -mtime "+$max_age_days" -delete 2>/dev/null
        echo "Cleaned packages older than $max_age_days days from $PACKAGE_CACHE_DIR" >&2
    fi
}

# Get cache statistics
# Usage: package_dl_cache_stats
package_dl_cache_stats() {
    if [[ ! -d "$PACKAGE_CACHE_DIR" ]]; then
        echo '{"total_files": 0, "total_size": "0"}'
        return
    fi

    local total_files=$(find "$PACKAGE_CACHE_DIR" -type f 2>/dev/null | wc -l | tr -d ' ')
    local total_size=$(du -sh "$PACKAGE_CACHE_DIR" 2>/dev/null | cut -f1)

    jq -n \
        --argjson files "$total_files" \
        --arg size "$total_size" \
        '{total_files: $files, total_size: $size}'
}

# Initialize cache on load
package_dl_init_cache

# Export functions
export -f package_dl_init_cache
export -f package_dl_npm
export -f package_dl_pypi
export -f package_dl_download
export -f package_dl_parse_ecosystem
export -f package_dl_extract_from_sbom
export -f package_dl_clean_cache
export -f package_dl_cache_stats
