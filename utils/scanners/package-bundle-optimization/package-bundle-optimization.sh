#!/bin/bash
# Bundle Size Analyzer
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0
#
# Analyzes bundle sizes and identifies optimization opportunities.
# Part of the Developer Productivity module.

set -eo pipefail

# Get script directory
LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

#############################################################################
# Bundle Size Thresholds (in bytes)
#############################################################################

# Size thresholds for npm packages
NPM_SIZE_THRESHOLDS="
lodash:70000:Consider lodash-es or individual imports
moment:300000:Consider dayjs or date-fns
rxjs:200000:Ensure proper tree-shaking
jquery:90000:Consider vanilla JS or smaller alternatives
bootstrap:200000:Consider CSS-only or utility-first alternatives
@angular/core:500000:Angular is large but tree-shakeable
react-dom:130000:Expected size for React
@mui/material:1000000:Large but modular, ensure selective imports
antd:2000000:Very large, use babel-plugin-import
lodash.debounce:2000:Prefer this over full lodash for single functions
"

# Recommended size budgets
BUNDLE_BUDGET_INITIAL_JS=${BUNDLE_BUDGET_INITIAL_JS:-200000}      # 200KB initial JS
BUNDLE_BUDGET_INITIAL_CSS=${BUNDLE_BUDGET_INITIAL_CSS:-100000}    # 100KB initial CSS
BUNDLE_BUDGET_TOTAL_JS=${BUNDLE_BUDGET_TOTAL_JS:-500000}          # 500KB total JS
BUNDLE_BUDGET_PER_ROUTE=${BUNDLE_BUDGET_PER_ROUTE:-100000}        # 100KB per route

#############################################################################
# Size Analysis Functions
#############################################################################

# Get package size from bundlephobia API (cached)
# Usage: get_package_size <package> [version]
get_package_size() {
    local pkg="$1"
    local version="${2:-latest}"

    # URL encode the package name
    local encoded_pkg=$(echo "$pkg" | sed 's/@/%40/g; s/\//%2F/g')

    # Try bundlephobia API
    local response=$(curl -s --max-time 10 "https://bundlephobia.com/api/size?package=${encoded_pkg}@${version}" 2>/dev/null || echo "")

    if [[ -z "$response" || "$response" == *"error"* ]]; then
        echo "{\"error\": \"api_unavailable\", \"package\": \"$pkg\"}"
        return
    fi

    # Extract relevant fields
    local size=$(echo "$response" | jq -r '.size // null')
    local gzip=$(echo "$response" | jq -r '.gzip // null')
    local dep_count=$(echo "$response" | jq -r '.dependencyCount // 0')

    if [[ "$size" == "null" ]]; then
        echo "{\"error\": \"package_not_found\", \"package\": \"$pkg\"}"
        return
    fi

    echo "{
        \"package\": \"$pkg\",
        \"version\": \"$version\",
        \"size\": $size,
        \"gzip\": $gzip,
        \"dependency_count\": $dep_count
    }" | jq '.'
}

# Get size rating for a package
# Usage: get_size_rating <size_bytes>
get_size_rating() {
    local size="$1"

    if [[ $size -lt 5000 ]]; then
        echo "excellent"
    elif [[ $size -lt 20000 ]]; then
        echo "good"
    elif [[ $size -lt 50000 ]]; then
        echo "moderate"
    elif [[ $size -lt 100000 ]]; then
        echo "large"
    else
        echo "very_large"
    fi
}

# Check if package has known size concerns
# Usage: check_size_concerns <package> <size>
check_size_concerns() {
    local pkg="$1"
    local size="$2"
    local concerns=()

    while IFS= read -r line; do
        [[ -z "$line" ]] && continue
        local pattern=$(echo "$line" | cut -d':' -f1)
        local threshold=$(echo "$line" | cut -d':' -f2)
        local message=$(echo "$line" | cut -d':' -f3-)

        if [[ "$pkg" == "$pattern" && $size -gt $threshold ]]; then
            concerns+=("{\"type\": \"known_large_package\", \"message\": \"$message\", \"threshold\": $threshold}")
        fi
    done <<< "$NPM_SIZE_THRESHOLDS"

    if [[ ${#concerns[@]} -gt 0 ]]; then
        printf '%s\n' "${concerns[@]}" | jq -s '.'
    else
        echo "[]"
    fi
}

# Analyze a single package
# Usage: analyze_package_size <package> [version]
analyze_package_size() {
    local pkg="$1"
    local version="${2:-latest}"

    local size_info=$(get_package_size "$pkg" "$version")

    if echo "$size_info" | jq -e '.error' >/dev/null 2>&1; then
        echo "$size_info"
        return
    fi

    local size=$(echo "$size_info" | jq -r '.size')
    local gzip=$(echo "$size_info" | jq -r '.gzip')
    local dep_count=$(echo "$size_info" | jq -r '.dependency_count')

    local rating=$(get_size_rating "$size")
    local concerns=$(check_size_concerns "$pkg" "$size")
    local concern_count=$(echo "$concerns" | jq 'length')

    local recommendations=()

    # Generate recommendations based on size
    if [[ $size -gt 100000 ]]; then
        recommendations+=("Consider code splitting or lazy loading")
    fi
    if [[ $dep_count -gt 20 ]]; then
        recommendations+=("High dependency count - review transitive deps")
    fi
    if [[ "$rating" == "very_large" ]]; then
        recommendations+=("Look for smaller alternatives")
    fi

    local recommendations_json=$(printf '%s\n' "${recommendations[@]}" 2>/dev/null | jq -R . | jq -s '.' || echo "[]")

    echo "{
        \"package\": \"$pkg\",
        \"version\": \"$version\",
        \"size\": $size,
        \"size_kb\": $(echo "scale=2; $size / 1024" | bc),
        \"gzip\": $gzip,
        \"gzip_kb\": $(echo "scale=2; $gzip / 1024" | bc),
        \"dependency_count\": $dep_count,
        \"rating\": \"$rating\",
        \"concerns\": $concerns,
        \"has_concerns\": $([ $concern_count -gt 0 ] && echo "true" || echo "false"),
        \"recommendations\": $recommendations_json
    }" | jq '.'
}

#############################################################################
# Project Analysis Functions
#############################################################################

# Analyze package.json dependencies
# Usage: analyze_npm_bundle <project_dir>
analyze_npm_bundle() {
    local project_dir="$1"
    local package_json="$project_dir/package.json"

    if [[ ! -f "$package_json" ]]; then
        echo '{"error": "no_package_json"}'
        return 1
    fi

    local deps=$(jq -r '.dependencies // {} | keys[]' "$package_json" 2>/dev/null)
    local results="[]"
    local total_size=0
    local total_gzip=0
    local analyzed=0
    local failed=0

    while IFS= read -r pkg; do
        [[ -z "$pkg" ]] && continue

        # Get version from package.json
        local version=$(jq -r ".dependencies[\"$pkg\"] // \"latest\"" "$package_json" | sed 's/[^0-9.]//g')
        [[ -z "$version" ]] && version="latest"

        local analysis=$(analyze_package_size "$pkg" "$version")

        if echo "$analysis" | jq -e '.error' >/dev/null 2>&1; then
            failed=$((failed + 1))
        else
            analyzed=$((analyzed + 1))
            local size=$(echo "$analysis" | jq -r '.size // 0')
            local gzip=$(echo "$analysis" | jq -r '.gzip // 0')
            total_size=$((total_size + size))
            total_gzip=$((total_gzip + gzip))
        fi

        results=$(echo "$results" | jq --argjson a "$analysis" '. + [$a]')
    done <<< "$deps"

    # Sort by size descending
    results=$(echo "$results" | jq 'sort_by(-.size // 0)')

    # Get top 10 largest
    local top_10=$(echo "$results" | jq '.[0:10]')

    # Budget analysis
    local over_budget="false"
    local budget_status="under"
    if [[ $total_gzip -gt $BUNDLE_BUDGET_TOTAL_JS ]]; then
        over_budget="true"
        budget_status="over"
    fi

    echo "{
        \"project_dir\": \"$project_dir\",
        \"summary\": {
            \"total_dependencies\": $(echo \"$deps\" | grep -c . || echo 0),
            \"analyzed\": $analyzed,
            \"failed\": $failed,
            \"total_size\": $total_size,
            \"total_size_kb\": $(echo "scale=2; $total_size / 1024" | bc),
            \"total_gzip\": $total_gzip,
            \"total_gzip_kb\": $(echo "scale=2; $total_gzip / 1024" | bc),
            \"budget\": {
                \"limit_kb\": $(echo "scale=2; $BUNDLE_BUDGET_TOTAL_JS / 1024" | bc),
                \"status\": \"$budget_status\",
                \"over_budget\": $over_budget
            }
        },
        \"top_largest\": $top_10,
        \"all_packages\": $results
    }" | jq '.'
}

# Estimate bundle impact of adding a new package
# Usage: estimate_impact <package> [version]
estimate_impact() {
    local pkg="$1"
    local version="${2:-latest}"

    local analysis=$(analyze_package_size "$pkg" "$version")

    if echo "$analysis" | jq -e '.error' >/dev/null 2>&1; then
        echo "$analysis"
        return
    fi

    local size=$(echo "$analysis" | jq -r '.size')
    local gzip=$(echo "$analysis" | jq -r '.gzip')
    local dep_count=$(echo "$analysis" | jq -r '.dependency_count')

    # Calculate impact percentages
    local budget_impact=$(echo "scale=2; ($gzip / $BUNDLE_BUDGET_TOTAL_JS) * 100" | bc)
    local initial_impact=$(echo "scale=2; ($gzip / $BUNDLE_BUDGET_INITIAL_JS) * 100" | bc)

    local impact_rating="low"
    if [[ $(echo "$budget_impact > 20" | bc -l) == "1" ]]; then
        impact_rating="very_high"
    elif [[ $(echo "$budget_impact > 10" | bc -l) == "1" ]]; then
        impact_rating="high"
    elif [[ $(echo "$budget_impact > 5" | bc -l) == "1" ]]; then
        impact_rating="medium"
    fi

    local recommendations=()
    if [[ "$impact_rating" == "very_high" || "$impact_rating" == "high" ]]; then
        recommendations+=("Consider lazy loading this package")
        recommendations+=("Look for smaller alternatives")
    fi
    if [[ $dep_count -gt 10 ]]; then
        recommendations+=("High dependency count may increase total bundle size")
    fi

    local recommendations_json=$(printf '%s\n' "${recommendations[@]}" 2>/dev/null | jq -R . | jq -s '.' || echo "[]")

    echo "{
        \"package\": \"$pkg\",
        \"version\": \"$version\",
        \"size_kb\": $(echo "scale=2; $size / 1024" | bc),
        \"gzip_kb\": $(echo "scale=2; $gzip / 1024" | bc),
        \"dependency_count\": $dep_count,
        \"impact\": {
            \"budget_percentage\": $budget_impact,
            \"initial_load_percentage\": $initial_impact,
            \"rating\": \"$impact_rating\"
        },
        \"recommendations\": $recommendations_json
    }" | jq '.'
}

#############################################################################
# Export Functions
#############################################################################

export -f get_package_size
export -f get_size_rating
export -f check_size_concerns
export -f analyze_package_size
export -f analyze_npm_bundle
export -f estimate_impact
