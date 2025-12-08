#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Bundle Analysis Scanner
# Analyzes npm/JavaScript bundle sizes, tree-shaking compatibility,
# and identifies heavy packages using Bundlephobia and npm registry APIs
#
# Usage: ./bundle-analysis.sh [options] <target>
# Output: JSON with bundle metrics, heavy packages, and recommendations
#############################################################################

set -e

# Colors for terminal output (stderr only)
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SCANNERS_ROOT="$(dirname "$SCRIPT_DIR")"
UTILS_ROOT="$(dirname "$SCANNERS_ROOT")"
REPO_ROOT="$(dirname "$UTILS_ROOT")"

# Source library files
source "$SCRIPT_DIR/lib/bundlephobia-client.sh"
source "$SCRIPT_DIR/lib/npm-registry-client.sh"
source "$SCRIPT_DIR/lib/tree-shake-detector.sh"

# Default options
OUTPUT_FILE=""
LOCAL_PATH=""
REPO=""
ORG=""
TEMP_DIR=""
CLEANUP=true
TARGET=""
HEAVY_THRESHOLD=50000  # 50KB gzipped
CACHE_DIR=""

# Version
VERSION="1.0.0"

usage() {
    cat << EOF
Bundle Analysis Scanner - Analyze npm bundle sizes and tree-shaking

Usage: $0 [OPTIONS] <target>

TARGET:
    Git repository URL      Clone and analyze repository
    Local directory path    Analyze local repository

OPTIONS:
    --local-path PATH       Use pre-cloned repository (skips cloning)
    --repo OWNER/REPO       GitHub repository (looks in zero cache)
    --org ORG               GitHub org (uses first repo found in zero cache)
    -o, --output FILE       Write JSON to file (default: stdout)
    --threshold BYTES       Heavy package threshold in gzipped bytes (default: 50000)
    --cache-dir DIR         Directory for API response caching
    -k, --keep-clone        Keep cloned repository
    -h, --help              Show this help

OUTPUT:
    JSON object with:
    - summary: total dependencies, sizes, heavy packages count
    - packages: array of package analysis results
    - top_largest: top 10 largest packages
    - tree_shaking_issues: packages with tree-shaking problems
    - recommendations: actionable suggestions

EXAMPLES:
    $0 https://github.com/expressjs/express
    $0 --local-path ~/.zero/projects/foo/repo
    $0 -o bundle-analysis.json /path/to/project
    $0 --threshold 100000 /path/to/project  # Flag packages > 100KB

EOF
    exit 0
}

# Clone repository
clone_repository() {
    local repo_url="$1"
    TEMP_DIR=$(mktemp -d)

    echo -e "${BLUE}Cloning repository...${NC}" >&2
    if git clone --depth 1 "$repo_url" "$TEMP_DIR" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Cloned${NC}" >&2
        return 0
    else
        echo '{"error": "Failed to clone repository"}'
        exit 1
    fi
}

# Cleanup
cleanup() {
    if [[ "$CLEANUP" == true ]] && [[ -n "$TEMP_DIR" ]] && [[ -d "$TEMP_DIR" ]]; then
        rm -rf "$TEMP_DIR"
    fi
}
trap cleanup EXIT

# Detect if target is a Git URL
is_git_url() {
    local url="$1"
    [[ "$url" =~ ^(https?://|git@|git://) ]]
}

# Find package.json files in project
find_package_json_files() {
    local target="$1"
    find "$target" -name "package.json" -not -path "*/node_modules/*" -type f 2>/dev/null
}

# Parse dependencies from package.json
parse_dependencies() {
    local package_json="$1"

    if [[ ! -f "$package_json" ]]; then
        echo '[]'
        return
    fi

    # Extract dependencies and devDependencies
    jq -r '
        ((.dependencies // {}) | to_entries | map({name: .key, version: .value, type: "production"})) +
        ((.devDependencies // {}) | to_entries | map({name: .key, version: .value, type: "development"}))
    ' "$package_json" 2>/dev/null || echo '[]'
}

# Clean version string (remove ^, ~, etc.)
clean_version() {
    local version="$1"
    # Remove semver prefixes and extract version number
    echo "$version" | sed -E 's/^[\^~>=<]*//' | sed -E 's/\s.*//'
}

# Analyze a single package
analyze_package() {
    local name="$1"
    local version="$2"
    local dep_type="$3"

    local clean_ver
    clean_ver=$(clean_version "$version")

    # Get bundle size from Bundlephobia
    local bundle_data
    bundle_data=$(get_bundle_size "$name" "$clean_ver" 2>/dev/null || echo '{"error": "fetch_failed"}')

    # Check for error
    if echo "$bundle_data" | jq -e '.error' &>/dev/null; then
        jq -n \
            --arg name "$name" \
            --arg version "$version" \
            --arg type "$dep_type" \
            --arg error "$(echo "$bundle_data" | jq -r '.error // "unknown"')" \
            '{
                name: $name,
                version: $version,
                type: $type,
                error: $error,
                analyzed: false
            }'
        return 1
    fi

    # Get tree-shaking analysis
    local tree_shake_data
    tree_shake_data=$(analyze_tree_shaking "$name" "$clean_ver" 2>/dev/null || echo '{}')

    # Extract values
    local size gzip is_heavy size_class tree_shakeable esm_support side_effects

    size=$(echo "$bundle_data" | jq -r '.size // 0')
    gzip=$(echo "$bundle_data" | jq -r '.gzip // 0')
    is_heavy=$(is_heavy_package "$gzip" "$HEAVY_THRESHOLD" && echo "true" || echo "false")
    size_class=$(classify_bundle_size "$gzip")

    tree_shakeable=$(echo "$tree_shake_data" | jq -r '.tree_shakeable // false')
    esm_support=$(echo "$tree_shake_data" | jq -r '.esm_support // "unknown"')
    side_effects=$(echo "$tree_shake_data" | jq -r '.side_effects // "unknown"')

    # Build result
    jq -n \
        --arg name "$name" \
        --arg version "$version" \
        --arg type "$dep_type" \
        --argjson size "$size" \
        --argjson gzip "$gzip" \
        --arg is_heavy "$is_heavy" \
        --arg size_class "$size_class" \
        --arg tree_shakeable "$tree_shakeable" \
        --arg esm_support "$esm_support" \
        --arg side_effects "$side_effects" \
        '{
            name: $name,
            version: $version,
            type: $type,
            size: $size,
            gzip: $gzip,
            is_heavy: ($is_heavy == "true"),
            size_class: $size_class,
            tree_shakeable: ($tree_shakeable == "true"),
            esm_support: $esm_support,
            side_effects: $side_effects,
            analyzed: true
        }'
}

# Generate recommendations based on analysis
generate_recommendations() {
    local packages_json="$1"
    local recommendations='[]'

    # Heavy packages without tree-shaking
    local heavy_no_treeshake
    heavy_no_treeshake=$(echo "$packages_json" | jq '[.[] | select(.is_heavy == true and .tree_shakeable == false)]')

    while read -r pkg; do
        [[ -z "$pkg" ]] && continue
        local name gzip
        name=$(echo "$pkg" | jq -r '.name')
        gzip=$(echo "$pkg" | jq -r '.gzip')

        recommendations=$(echo "$recommendations" | jq \
            --arg pkg "$name" \
            --arg gzip "$gzip" \
            '. + [{
                type: "heavy_package",
                package: $pkg,
                size_gzip: ($gzip | tonumber),
                issue: "Large package without tree-shaking support",
                suggestion: "Consider lighter alternatives or evaluate if full library is needed"
            }]')
    done < <(echo "$heavy_no_treeshake" | jq -c '.[]' 2>/dev/null)

    # Packages without ESM
    local no_esm
    no_esm=$(echo "$packages_json" | jq '[.[] | select(.esm_support == "none" and .analyzed == true)]')

    while read -r pkg; do
        [[ -z "$pkg" ]] && continue
        local name
        name=$(echo "$pkg" | jq -r '.name')

        recommendations=$(echo "$recommendations" | jq \
            --arg pkg "$name" \
            '. + [{
                type: "tree_shaking",
                package: $pkg,
                issue: "CommonJS only - cannot tree-shake",
                suggestion: "Use ESM version if available or granular imports"
            }]')
    done < <(echo "$no_esm" | jq -c '.[]' 2>/dev/null)

    echo "$recommendations"
}

# Main analysis function
analyze_bundle() {
    local target="$1"

    echo -e "${BLUE}Analyzing bundle for: ${CYAN}$target${NC}" >&2

    # Find package.json
    local package_json="$target/package.json"
    if [[ ! -f "$package_json" ]]; then
        echo '{"error": "No package.json found", "target": "'"$target"'"}'
        return 1
    fi

    # Get project name
    local project_name
    project_name=$(jq -r '.name // "unknown"' "$package_json")

    echo -e "${BLUE}Project: ${CYAN}$project_name${NC}" >&2

    # Parse dependencies
    local deps
    deps=$(parse_dependencies "$package_json")
    local total_deps
    total_deps=$(echo "$deps" | jq 'length')

    echo -e "${BLUE}Found ${CYAN}$total_deps${BLUE} dependencies${NC}" >&2

    # Set up caching if specified
    if [[ -n "$CACHE_DIR" ]]; then
        export BUNDLEPHOBIA_CACHE_DIR="$CACHE_DIR/bundlephobia"
        export NPM_CACHE_DIR="$CACHE_DIR/npm"
        mkdir -p "$BUNDLEPHOBIA_CACHE_DIR" "$NPM_CACHE_DIR" 2>/dev/null || true
    fi

    # Analyze each package
    local packages='[]'
    local analyzed=0
    local failed=0
    local total_size=0
    local total_gzip=0
    local heavy_count=0
    local tree_shakeable_count=0
    local not_tree_shakeable_count=0

    while read -r dep; do
        [[ -z "$dep" ]] && continue

        local name version dep_type
        name=$(echo "$dep" | jq -r '.name')
        version=$(echo "$dep" | jq -r '.version')
        dep_type=$(echo "$dep" | jq -r '.type')

        echo -e "  ${BLUE}Analyzing: ${CYAN}$name@$version${NC}" >&2

        local result
        result=$(analyze_package "$name" "$version" "$dep_type")

        if echo "$result" | jq -e '.analyzed == true' &>/dev/null; then
            analyzed=$((analyzed + 1))

            local pkg_size pkg_gzip pkg_heavy pkg_treeshake
            pkg_size=$(echo "$result" | jq -r '.size // 0')
            pkg_gzip=$(echo "$result" | jq -r '.gzip // 0')
            pkg_heavy=$(echo "$result" | jq -r '.is_heavy')
            pkg_treeshake=$(echo "$result" | jq -r '.tree_shakeable')

            total_size=$((total_size + pkg_size))
            total_gzip=$((total_gzip + pkg_gzip))

            [[ "$pkg_heavy" == "true" ]] && heavy_count=$((heavy_count + 1))
            if [[ "$pkg_treeshake" == "true" ]]; then
                tree_shakeable_count=$((tree_shakeable_count + 1))
            else
                not_tree_shakeable_count=$((not_tree_shakeable_count + 1))
            fi
        else
            failed=$((failed + 1))
        fi

        packages=$(echo "$packages" | jq --argjson pkg "$result" '. + [$pkg]')

    done < <(echo "$deps" | jq -c '.[]')

    echo -e "${GREEN}✓ Analyzed $analyzed packages${NC}" >&2

    # Get top largest packages
    local top_largest
    top_largest=$(echo "$packages" | jq '[.[] | select(.analyzed == true)] | sort_by(-.gzip) | .[0:10]')

    # Get tree-shaking issues
    local tree_shaking_issues
    tree_shaking_issues=$(echo "$packages" | jq '[.[] | select(.analyzed == true and .tree_shakeable == false and .esm_support == "none")] | sort_by(-.gzip)')

    # Generate recommendations
    local recommendations
    recommendations=$(generate_recommendations "$packages")

    # Build final output
    jq -n \
        --arg analyzer "bundle-analysis" \
        --arg version "$VERSION" \
        --arg timestamp "$(date -u +"%Y-%m-%dT%H:%M:%SZ")" \
        --arg target "$target" \
        --arg project_name "$project_name" \
        --argjson total_deps "$total_deps" \
        --argjson analyzed "$analyzed" \
        --argjson failed "$failed" \
        --argjson total_size "$total_size" \
        --argjson total_gzip "$total_gzip" \
        --argjson heavy_count "$heavy_count" \
        --argjson tree_shakeable "$tree_shakeable_count" \
        --argjson not_tree_shakeable "$not_tree_shakeable_count" \
        --argjson threshold "$HEAVY_THRESHOLD" \
        --argjson packages "$packages" \
        --argjson top_largest "$top_largest" \
        --argjson tree_shaking_issues "$tree_shaking_issues" \
        --argjson recommendations "$recommendations" \
        '{
            analyzer: $analyzer,
            version: $version,
            timestamp: $timestamp,
            target: $target,
            project_name: $project_name,
            summary: {
                total_dependencies: $total_deps,
                analyzed: $analyzed,
                failed: $failed,
                total_size_bytes: $total_size,
                total_gzip_bytes: $total_gzip,
                total_size_kb: (($total_size / 1024) | floor),
                total_gzip_kb: (($total_gzip / 1024) | floor),
                heavy_packages: $heavy_count,
                heavy_threshold_bytes: $threshold,
                tree_shakeable: $tree_shakeable,
                not_tree_shakeable: $not_tree_shakeable
            },
            packages: $packages,
            top_largest: $top_largest,
            tree_shaking_issues: $tree_shaking_issues,
            recommendations: $recommendations
        }'
}

# Parse arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --local-path)
                LOCAL_PATH="$2"
                shift 2
                ;;
            --repo)
                REPO="$2"
                shift 2
                ;;
            --org)
                ORG="$2"
                shift 2
                ;;
            -o|--output)
                OUTPUT_FILE="$2"
                shift 2
                ;;
            --threshold)
                HEAVY_THRESHOLD="$2"
                shift 2
                ;;
            --cache-dir)
                CACHE_DIR="$2"
                shift 2
                ;;
            -k|--keep-clone)
                CLEANUP=false
                shift
                ;;
            -h|--help)
                usage
                ;;
            *)
                TARGET="$1"
                shift
                ;;
        esac
    done
}

# Main execution
main() {
    parse_args "$@"

    # Determine target directory
    local target_dir=""

    if [[ -n "$LOCAL_PATH" ]]; then
        target_dir="$LOCAL_PATH"
    elif [[ -n "$REPO" ]]; then
        # Look in zero cache
        local zero_home="${ZERO_HOME:-$HOME/.zero}"
        target_dir="$zero_home/repos/$REPO/repo"
        if [[ ! -d "$target_dir" ]]; then
            echo '{"error": "Repository not found in zero cache", "repo": "'"$REPO"'"}'
            exit 1
        fi
    elif [[ -n "$ORG" ]]; then
        # Find first repo in org cache
        local zero_home="${ZERO_HOME:-$HOME/.zero}"
        local org_dir="$zero_home/repos/$ORG"
        if [[ -d "$org_dir" ]]; then
            target_dir=$(find "$org_dir" -maxdepth 2 -type d -name "repo" | head -1)
        fi
        if [[ -z "$target_dir" ]]; then
            echo '{"error": "No repositories found for org", "org": "'"$ORG"'"}'
            exit 1
        fi
    elif [[ -n "$TARGET" ]]; then
        if is_git_url "$TARGET"; then
            clone_repository "$TARGET"
            target_dir="$TEMP_DIR"
        elif [[ -d "$TARGET" ]]; then
            target_dir="$TARGET"
        else
            echo '{"error": "Target not found", "target": "'"$TARGET"'"}'
            exit 1
        fi
    else
        usage
    fi

    # Run analysis
    local output
    output=$(analyze_bundle "$target_dir")

    # Output result
    if [[ -n "$OUTPUT_FILE" ]]; then
        echo "$output" > "$OUTPUT_FILE"
        echo -e "${GREEN}✓ Output written to: ${CYAN}$OUTPUT_FILE${NC}" >&2
    else
        echo "$output"
    fi
}

# Run if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
