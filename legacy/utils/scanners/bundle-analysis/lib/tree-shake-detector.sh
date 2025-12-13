#!/usr/bin/env bash
# Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0
#
# Tree-Shaking Detector
# Analyzes packages for tree-shaking compatibility
#
# Usage:
#   source tree-shake-detector.sh
#   analyze_tree_shaking "react" "18.2.0"

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source npm registry client
source "$SCRIPT_DIR/npm-registry-client.sh"

# Analyze a package's tree-shaking compatibility
# Usage: analyze_tree_shaking "package-name" ["version"]
# Returns JSON with tree-shaking analysis
analyze_tree_shaking() {
    local package="$1"
    local version="${2:-}"

    local esm_info
    esm_info=$(get_esm_info "$package" "$version")

    if echo "$esm_info" | jq -e '.error' &>/dev/null; then
        echo "$esm_info"
        return 1
    fi

    # Extract fields
    local has_module has_exports has_type_module side_effects main_field

    has_module=$(echo "$esm_info" | jq -r '.module // empty' 2>/dev/null)
    has_exports=$(echo "$esm_info" | jq -r 'if .exports then "true" else "false" end' 2>/dev/null)
    has_type_module=$(echo "$esm_info" | jq -r 'if .type == "module" then "true" else "false" end' 2>/dev/null)
    side_effects=$(echo "$esm_info" | jq '.sideEffects' 2>/dev/null)
    main_field=$(echo "$esm_info" | jq -r '.main // empty' 2>/dev/null)

    # Determine tree-shaking status
    local is_tree_shakeable="false"
    local esm_support="none"
    local issues=()
    local recommendations=()

    # Check ESM support level
    if [[ "$has_type_module" == "true" ]]; then
        esm_support="full"
        is_tree_shakeable="true"
    elif [[ -n "$has_module" ]]; then
        esm_support="dual"  # Has both CJS and ESM
        is_tree_shakeable="true"
    elif [[ "$has_exports" == "true" ]]; then
        # Check if exports has ESM entry
        local exports_has_esm
        exports_has_esm=$(echo "$esm_info" | jq -r '.exports | if type == "object" then (if .import then "true" else "false" end) else "false" end' 2>/dev/null)
        if [[ "$exports_has_esm" == "true" ]]; then
            esm_support="conditional"
            is_tree_shakeable="true"
        fi
    fi

    # If no ESM support found
    if [[ "$esm_support" == "none" ]]; then
        issues+=("No ESM entry point detected (no 'module', 'exports.import', or 'type: module')")
        recommendations+=("Check for an ESM version of this package (e.g., ${package}-es or @esm-bundle/${package})")
        recommendations+=("Consider using granular imports if the package supports them")
    fi

    # Check sideEffects
    local side_effects_status="unknown"
    if [[ "$side_effects" == "false" ]]; then
        side_effects_status="none"
    elif [[ "$side_effects" == "true" ]]; then
        side_effects_status="all"
        if [[ "$is_tree_shakeable" == "true" ]]; then
            issues+=("Package declares sideEffects: true, limiting tree-shaking effectiveness")
        fi
    elif [[ "$side_effects" != "null" && "$side_effects" != "unknown" ]]; then
        side_effects_status="partial"
    fi

    # Build result JSON
    jq -n \
        --arg package "$package" \
        --arg version "$(echo "$esm_info" | jq -r '.version // empty')" \
        --arg is_tree_shakeable "$is_tree_shakeable" \
        --arg esm_support "$esm_support" \
        --arg side_effects_status "$side_effects_status" \
        --arg has_module_field "$([ -n "$has_module" ] && echo "true" || echo "false")" \
        --arg has_exports_field "$has_exports" \
        --arg has_type_module "$has_type_module" \
        --argjson issues "$(printf '%s\n' "${issues[@]:-}" | jq -R . | jq -s .)" \
        --argjson recommendations "$(printf '%s\n' "${recommendations[@]:-}" | jq -R . | jq -s .)" \
        '{
            package: $package,
            version: $version,
            tree_shakeable: ($is_tree_shakeable == "true"),
            esm_support: $esm_support,
            side_effects: $side_effects_status,
            metadata: {
                has_module_field: ($has_module_field == "true"),
                has_exports_field: ($has_exports_field == "true"),
                has_type_module: ($has_type_module == "true")
            },
            issues: $issues,
            recommendations: $recommendations
        }' 2>/dev/null
}

# Analyze multiple packages
# Usage: analyze_tree_shaking_batch "pkg1@version" "pkg2" ...
# Returns JSON array of analyses
analyze_tree_shaking_batch() {
    local packages=("$@")
    local results=()

    for pkg_spec in "${packages[@]}"; do
        local package="${pkg_spec%@*}"
        local version=""

        if [[ "$pkg_spec" == *"@"* ]]; then
            version="${pkg_spec##*@}"
        fi

        local result
        result=$(analyze_tree_shaking "$package" "$version")
        results+=("$result")
    done

    printf '%s\n' "${results[@]}" | jq -s '.' 2>/dev/null || echo '[]'
}

# Generate tree-shaking recommendation for a package
# Usage: get_tree_shaking_recommendation "package" "esm_support" "side_effects" "gzip_size"
get_tree_shaking_recommendation() {
    local package="$1"
    local esm_support="$2"
    local side_effects="$3"
    local gzip_size="${4:-0}"

    local recommendation=""
    local priority="low"

    case "$esm_support" in
        "none")
            priority="high"
            recommendation="Package '$package' is CommonJS only and cannot be tree-shaken. "
            if [[ "$gzip_size" -gt 50000 ]]; then
                recommendation+="Consider finding an ESM alternative or using granular imports."
            else
                recommendation+="Consider using granular imports if available."
            fi
            ;;
        "dual"|"conditional")
            if [[ "$side_effects" == "all" ]]; then
                priority="medium"
                recommendation="Package '$package' has ESM but declares side effects. Tree-shaking may be limited."
            elif [[ "$side_effects" == "unknown" ]]; then
                priority="low"
                recommendation="Package '$package' supports ESM but doesn't declare sideEffects. Tree-shaking should work but verify."
            fi
            ;;
        "full")
            if [[ "$side_effects" == "all" ]]; then
                priority="medium"
                recommendation="Package '$package' is ESM but declares side effects, limiting tree-shaking."
            fi
            ;;
    esac

    jq -n \
        --arg package "$package" \
        --arg recommendation "$recommendation" \
        --arg priority "$priority" \
        '{
            package: $package,
            priority: $priority,
            recommendation: $recommendation
        }' 2>/dev/null
}

# Export functions
export -f analyze_tree_shaking
export -f analyze_tree_shaking_batch
export -f get_tree_shaking_recommendation
