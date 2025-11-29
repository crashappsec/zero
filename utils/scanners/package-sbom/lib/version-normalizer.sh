#!/bin/bash
# Version Normalizer Library
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0
#
# Normalizes version strings across ecosystems for reliable dependency resolution
# and consistent vulnerability matching. Part of the Reliability & Standardization module.

set -euo pipefail

#############################################################################
# Version Normalization Functions
# Ensures consistent versioning across ecosystems for reproducible builds
#############################################################################

# Normalize version based on ecosystem
# Usage: normalize_version <version> <ecosystem>
# Ecosystems: npm, pypi, maven, nuget, go, cargo, rubygems
normalize_version() {
    local version="$1"
    local ecosystem="${2:-npm}"

    case "$ecosystem" in
        npm|node)
            normalize_npm_version "$version"
            ;;
        pypi|python)
            normalize_pypi_version "$version"
            ;;
        maven|java)
            normalize_maven_version "$version"
            ;;
        nuget|dotnet)
            normalize_nuget_version "$version"
            ;;
        go|golang)
            normalize_go_version "$version"
            ;;
        cargo|rust)
            normalize_cargo_version "$version"
            ;;
        rubygems|ruby)
            normalize_rubygems_version "$version"
            ;;
        *)
            # Unknown ecosystem - return as-is
            echo "$version"
            ;;
    esac
}

# npm/Node.js - SemVer normalization
# Removes 'v' prefix, pads to 3 segments, handles pre-release tags
normalize_npm_version() {
    local version="$1"

    # Remove leading 'v' or 'V'
    version="${version#[vV]}"

    # Remove build metadata (everything after +)
    version="${version%%+*}"

    # Split on hyphen to handle pre-release separately
    local main_version="${version%%-*}"
    local prerelease=""
    if [[ "$version" == *-* ]]; then
        prerelease="-${version#*-}"
    fi

    # Pad to 3 segments (major.minor.patch) - compatible with both bash and zsh
    local major minor patch
    local seg_count=$(echo "$main_version" | tr '.' '\n' | wc -l | tr -d ' ')

    major=$(echo "$main_version" | cut -d. -f1)
    if [[ $seg_count -ge 2 ]]; then
        minor=$(echo "$main_version" | cut -d. -f2)
    else
        minor=""
    fi
    if [[ $seg_count -ge 3 ]]; then
        patch=$(echo "$main_version" | cut -d. -f3)
    else
        patch=""
    fi

    # Default to 0 if empty
    major="${major:-0}"
    minor="${minor:-0}"
    patch="${patch:-0}"

    # Remove leading zeros from numeric segments
    major=$((10#$major))
    minor=$((10#$minor))
    patch=$((10#$patch))

    echo "${major}.${minor}.${patch}${prerelease}"
}

# PyPI/Python - PEP 440 normalization
# Handles epoch, pre-release, post-release, dev, and local versions
normalize_pypi_version() {
    local version="$1"

    # Lowercase everything
    version=$(echo "$version" | tr '[:upper:]' '[:lower:]')

    # Replace _ and - with . for consistency
    version=$(echo "$version" | sed 's/[_-]/./g')

    # Normalize pre-release tags
    version=$(echo "$version" | sed -E 's/\.?alpha\.?/a/g')
    version=$(echo "$version" | sed -E 's/\.?beta\.?/b/g')
    version=$(echo "$version" | sed -E 's/\.?preview\.?/rc/g')
    version=$(echo "$version" | sed -E 's/\.?rc\.?/rc/g')
    version=$(echo "$version" | sed -E 's/\.?c\.?/rc/g')

    # Normalize post-release
    version=$(echo "$version" | sed -E 's/\.?(post|rev|r)\.?/\.post/g')

    # Normalize dev release
    version=$(echo "$version" | sed -E 's/\.?dev\.?/\.dev/g')

    # Remove leading zeros from numeric segments
    # Process each segment using simpler approach
    local result=""
    local remaining="$version"
    while [[ -n "$remaining" ]]; do
        local seg="${remaining%%.*}"
        if [[ "$remaining" == *"."* ]]; then
            remaining="${remaining#*.}"
        else
            remaining=""
        fi

        # If segment is purely numeric, strip leading zeros
        if [[ "$seg" =~ ^[0-9]+$ ]]; then
            seg=$((10#$seg))
        fi

        if [[ -z "$result" ]]; then
            result="$seg"
        else
            result="${result}.${seg}"
        fi
    done

    echo "$result"
}

# Maven/Java version normalization
# Handles qualifiers like -SNAPSHOT, -RELEASE, -M1, -RC1
normalize_maven_version() {
    local version="$1"

    # Normalize common qualifiers (case-insensitive replacement)
    # SNAPSHOT stays as-is (preferred form)
    version=$(echo "$version" | sed -E 's/-[sS][nN][aA][pP][sS][hH][oO][tT]$/-SNAPSHOT/')

    # Remove -release and -final suffixes
    version=$(echo "$version" | sed -E 's/-[rR][eE][lL][eE][aA][sS][eE]$//')
    version=$(echo "$version" | sed -E 's/-[fF][iI][nN][aA][lL]$//')
    version=$(echo "$version" | sed -E 's/\.[rR][eE][lL][eE][aA][sS][eE]$//')
    version=$(echo "$version" | sed -E 's/\.[fF][iI][nN][aA][lL]$//')

    # Normalize milestone versions -m1 -> -M1
    version=$(echo "$version" | sed -E 's/-[mM]([0-9]+)$/-M\1/')

    # Normalize RC versions -rc1 -> -RC1
    version=$(echo "$version" | sed -E 's/-[rR][cC]([0-9]+)$/-RC\1/')

    echo "$version"
}

# NuGet/.NET version normalization
# Follows SemVer 2.0 with some extensions
normalize_nuget_version() {
    local version="$1"

    # Remove leading 'v'
    version="${version#[vV]}"

    # Split main version and suffix
    local main_version="${version%%-*}"
    local suffix=""
    if [[ "$version" == *-* ]]; then
        suffix="-${version#*-}"
    fi

    # Count segments and check for 4th segment being 0
    local seg_count=$(echo "$main_version" | tr '.' '\n' | wc -l | tr -d ' ')
    if [[ $seg_count -eq 4 ]]; then
        local fourth=$(echo "$main_version" | cut -d. -f4)
        if [[ "$fourth" == "0" ]]; then
            main_version=$(echo "$main_version" | cut -d. -f1-3)
        fi
    fi

    echo "${main_version}${suffix}"
}

# Go module version normalization
# Handles v prefix requirement and pseudo-versions
normalize_go_version() {
    local version="$1"

    # Go requires v prefix for module versions
    if [[ ! "$version" =~ ^v ]]; then
        version="v${version}"
    fi

    # Handle pseudo-versions (v0.0.0-timestamp-commithash)
    if [[ "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+-[0-9]+-[a-f0-9]+$ ]]; then
        # This is a pseudo-version, keep as-is
        echo "$version"
        return
    fi

    # Handle +incompatible suffix
    local incompatible=""
    if [[ "$version" == *+incompatible ]]; then
        incompatible="+incompatible"
        version="${version%+incompatible}"
    fi

    # Remove v prefix temporarily for normalization
    version="${version#v}"

    # Pad to 3 segments using cut
    local main_version="${version%%-*}"
    local prerelease=""
    if [[ "$version" == *-* ]]; then
        prerelease="-${version#*-}"
    fi

    local major=$(echo "$main_version" | cut -d. -f1)
    local minor=$(echo "$main_version" | cut -d. -f2)
    local patch=$(echo "$main_version" | cut -d. -f3)

    major="${major:-0}"
    minor="${minor:-0}"
    patch="${patch:-0}"

    echo "v${major}.${minor}.${patch}${prerelease}${incompatible}"
}

# Cargo/Rust version normalization
# Follows SemVer strictly
normalize_cargo_version() {
    local version="$1"

    # Remove leading 'v' if present (Cargo doesn't use it)
    version="${version#[vV]}"

    # Cargo follows SemVer strictly - use npm normalization
    normalize_npm_version "$version"
}

# RubyGems version normalization
# Ruby uses a relaxed versioning scheme
normalize_rubygems_version() {
    local version="$1"

    # Remove leading 'v'
    version="${version#[vV]}"

    # Normalize pre-release markers
    version=$(echo "$version" | sed -E 's/\.alpha\./\.a\./g')
    version=$(echo "$version" | sed -E 's/\.beta\./\.b\./g')
    version=$(echo "$version" | sed -E 's/\.pre\./\.pre\./g')

    # Ensure at least 3 segments
    local segment_count=$(echo "$version" | tr '.' '\n' | wc -l | tr -d ' ')
    if [[ $segment_count -lt 3 ]]; then
        while [[ $segment_count -lt 3 ]]; do
            version="${version}.0"
            ((segment_count++))
        done
    fi

    echo "$version"
}

#############################################################################
# Version Comparison Functions
#############################################################################

# Compare two versions
# Usage: compare_versions <v1> <v2> <ecosystem>
# Returns: -1 (v1 < v2), 0 (v1 == v2), 1 (v1 > v2)
compare_versions() {
    local v1="$1"
    local v2="$2"
    local ecosystem="${3:-npm}"

    # Normalize both versions first
    local nv1=$(normalize_version "$v1" "$ecosystem")
    local nv2=$(normalize_version "$v2" "$ecosystem")

    # Use sort -V for natural version comparison
    if [[ "$nv1" == "$nv2" ]]; then
        echo "0"
        return
    fi

    # Compare using sort -V
    local sorted=$(printf '%s\n%s' "$nv1" "$nv2" | sort -V | head -1)
    if [[ "$sorted" == "$nv1" ]]; then
        echo "-1"
    else
        echo "1"
    fi
}

# Check if version satisfies a version range
# Usage: version_satisfies <version> <range> <ecosystem>
# Range formats: ">=1.0.0", "^1.0.0", "~1.0.0", "1.0.0 - 2.0.0", etc.
version_satisfies() {
    local version="$1"
    local range="$2"
    local ecosystem="${3:-npm}"

    local normalized=$(normalize_version "$version" "$ecosystem")

    # Handle exact version (no operators)
    if [[ ! "$range" =~ [\<\>=~^] && ! "$range" =~ " - " ]]; then
        local normalized_range=$(normalize_version "$range" "$ecosystem")
        [[ "$normalized" == "$normalized_range" ]] && echo "true" || echo "false"
        return
    fi

    # Handle >= constraint
    if [[ "$range" == ">="* ]]; then
        local min="${range#>=}"
        local cmp=$(compare_versions "$normalized" "$min" "$ecosystem")
        [[ "$cmp" -ge 0 ]] && echo "true" || echo "false"
        return
    fi

    # Handle > constraint (but not >=)
    if [[ "$range" == ">"* && "$range" != ">="* ]]; then
        local min="${range#>}"
        local cmp=$(compare_versions "$normalized" "$min" "$ecosystem")
        [[ "$cmp" -gt 0 ]] && echo "true" || echo "false"
        return
    fi

    # Handle <= constraint
    if [[ "$range" == "<="* ]]; then
        local max="${range#<=}"
        local cmp=$(compare_versions "$normalized" "$max" "$ecosystem")
        [[ "$cmp" -le 0 ]] && echo "true" || echo "false"
        return
    fi

    # Handle < constraint (but not <=)
    if [[ "$range" == "<"* && "$range" != "<="* ]]; then
        local max="${range#<}"
        local cmp=$(compare_versions "$normalized" "$max" "$ecosystem")
        [[ "$cmp" -lt 0 ]] && echo "true" || echo "false"
        return
    fi

    # Default: unable to determine
    echo "unknown"
}

# Parse version range into components
# Usage: parse_version_range <range> <ecosystem>
# Returns JSON: {"min": "x.y.z", "max": "x.y.z", "min_inclusive": true, "max_inclusive": true}
parse_version_range() {
    local range="$1"
    local ecosystem="${2:-npm}"

    local min=""
    local max=""
    local min_inclusive="true"
    local max_inclusive="true"

    # Handle hyphen range: "1.0.0 - 2.0.0"
    if [[ "$range" =~ ^([^ ]+)\ -\ ([^ ]+)$ ]]; then
        min="${BASH_REMATCH[1]}"
        max="${BASH_REMATCH[2]}"
        min_inclusive="true"
        max_inclusive="true"
    # Handle >= constraint
    elif [[ "$range" == ">="* ]]; then
        min="${range#>=}"
        min_inclusive="true"
    # Handle > constraint (but not >=)
    elif [[ "$range" == ">"* && "$range" != ">="* ]]; then
        min="${range#>}"
        min_inclusive="false"
    # Handle <= constraint
    elif [[ "$range" == "<="* ]]; then
        max="${range#<=}"
        max_inclusive="true"
    # Handle < constraint (but not <=)
    elif [[ "$range" == "<"* && "$range" != "<="* ]]; then
        max="${range#<}"
        max_inclusive="false"
    # Exact version
    else
        min="$range"
        max="$range"
    fi

    # Normalize versions
    [[ -n "$min" ]] && min=$(normalize_version "$min" "$ecosystem")
    [[ -n "$max" ]] && max=$(normalize_version "$max" "$ecosystem")

    echo "{\"min\": \"$min\", \"max\": \"$max\", \"min_inclusive\": $min_inclusive, \"max_inclusive\": $max_inclusive}"
}

#############################################################################
# Utility Functions
#############################################################################

# Extract major version
# Usage: get_major_version <version> <ecosystem>
get_major_version() {
    local version="$1"
    local ecosystem="${2:-npm}"

    local normalized=$(normalize_version "$version" "$ecosystem")

    # Remove v prefix if present
    normalized="${normalized#[vV]}"

    # Get first segment
    echo "${normalized%%.*}"
}

# Check if version is pre-release
# Usage: is_prerelease <version> <ecosystem>
is_prerelease() {
    local version="$1"
    local ecosystem="${2:-npm}"

    local normalized=$(normalize_version "$version" "$ecosystem")

    # Check for common pre-release indicators
    if [[ "$normalized" =~ (-alpha|-beta|-rc|-pre|-dev|\.a[0-9]|\.b[0-9]|\.rc[0-9]) ]]; then
        echo "true"
    elif [[ "$normalized" =~ (SNAPSHOT|snapshot) ]]; then
        echo "true"
    else
        echo "false"
    fi
}

# Get versions between two versions (for counting versions behind)
# Usage: count_versions_behind <current> <latest> <all_versions_json> <ecosystem>
# all_versions_json should be a JSON array: ["1.0.0", "1.1.0", "2.0.0"]
count_versions_behind() {
    local current="$1"
    local latest="$2"
    local all_versions_json="$3"
    local ecosystem="${4:-npm}"

    local current_normalized=$(normalize_version "$current" "$ecosystem")
    local latest_normalized=$(normalize_version "$latest" "$ecosystem")

    # Parse JSON array and count versions between current and latest
    local count=0
    while IFS= read -r version; do
        local normalized=$(normalize_version "$version" "$ecosystem")
        local cmp_current=$(compare_versions "$normalized" "$current_normalized" "$ecosystem")
        local cmp_latest=$(compare_versions "$normalized" "$latest_normalized" "$ecosystem")

        # Count if version is greater than current and less than or equal to latest
        if [[ "$cmp_current" -gt 0 && "$cmp_latest" -le 0 ]]; then
            ((count++))
        fi
    done < <(echo "$all_versions_json" | jq -r '.[]' 2>/dev/null)

    echo "$count"
}

# Batch normalize versions
# Usage: normalize_versions_batch <json_array>
# Input format: [{"name": "pkg", "version": "1.0", "ecosystem": "npm"}, ...]
# Output format: [{"name": "pkg", "original": "1.0", "normalized": "1.0.0", "ecosystem": "npm"}, ...]
normalize_versions_batch() {
    local packages_json="$1"
    local results="[]"

    while IFS= read -r pkg; do
        local name=$(echo "$pkg" | jq -r '.name')
        local version=$(echo "$pkg" | jq -r '.version')
        local ecosystem=$(echo "$pkg" | jq -r '.ecosystem // "npm"')

        local normalized=$(normalize_version "$version" "$ecosystem")

        results=$(echo "$results" | jq --arg name "$name" \
            --arg original "$version" \
            --arg normalized "$normalized" \
            --arg ecosystem "$ecosystem" \
            '. + [{"name": $name, "original": $original, "normalized": $normalized, "ecosystem": $ecosystem}]')
    done < <(echo "$packages_json" | jq -c '.[]')

    echo "$results"
}

# Export all functions
export -f normalize_version
export -f normalize_npm_version
export -f normalize_pypi_version
export -f normalize_maven_version
export -f normalize_nuget_version
export -f normalize_go_version
export -f normalize_cargo_version
export -f normalize_rubygems_version
export -f compare_versions
export -f version_satisfies
export -f parse_version_range
export -f get_major_version
export -f is_prerelease
export -f count_versions_behind
export -f normalize_versions_batch
