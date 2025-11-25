#!/bin/bash
# Unused Dependency Detector
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0
#
# Detects unused dependencies in projects to reduce attack surface and build times.
# Part of the Developer Productivity module.

set -eo pipefail

# Get script directory
LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

#############################################################################
# NPM/Node.js Unused Dependency Detection
#############################################################################

# Detect unused npm dependencies using depcheck
# Usage: detect_unused_npm <project_dir>
detect_unused_npm() {
    local project_dir="$1"

    if [[ ! -d "$project_dir" ]]; then
        echo '{"error": "project_directory_not_found"}'
        return 1
    fi

    if [[ ! -f "$project_dir/package.json" ]]; then
        echo '{"error": "no_package_json_found"}'
        return 1
    fi

    # Check if depcheck is available
    if command -v npx &> /dev/null; then
        # Run depcheck
        local result=$(cd "$project_dir" && npx depcheck --json 2>/dev/null || echo '{"error": "depcheck_failed"}')

        if [[ "$result" == *"error"* ]]; then
            # Fallback to manual analysis
            detect_unused_npm_manual "$project_dir"
        else
            echo "$result" | jq '{
                unused_dependencies: .dependencies,
                unused_devDependencies: .devDependencies,
                missing_dependencies: .missing,
                tool: "depcheck"
            }' 2>/dev/null || echo '{"error": "parse_failed"}'
        fi
    else
        # Fallback to manual analysis
        detect_unused_npm_manual "$project_dir"
    fi
}

# Manual npm unused detection (without depcheck)
# Usage: detect_unused_npm_manual <project_dir>
detect_unused_npm_manual() {
    local project_dir="$1"
    local package_json="$project_dir/package.json"

    # Get declared dependencies
    local deps=$(jq -r '.dependencies // {} | keys[]' "$package_json" 2>/dev/null)
    local dev_deps=$(jq -r '.devDependencies // {} | keys[]' "$package_json" 2>/dev/null)

    local unused_deps=()
    local unused_dev_deps=()

    # Check each dependency by searching for imports
    while IFS= read -r dep; do
        [[ -z "$dep" ]] && continue

        # Normalize dependency name for search (handle scoped packages)
        local search_pattern="$dep"
        if [[ "$dep" == @*/* ]]; then
            # For scoped packages, search for the full name
            search_pattern=$(echo "$dep" | sed 's/@/\\@/g; s/\//\\//g')
        fi

        # Search for require/import statements
        local found=false
        if grep -rq "require(['\"]$search_pattern" "$project_dir" --include="*.js" --include="*.ts" --include="*.jsx" --include="*.tsx" --include="*.mjs" --include="*.cjs" 2>/dev/null; then
            found=true
        fi
        if grep -rq "from ['\"]$search_pattern" "$project_dir" --include="*.js" --include="*.ts" --include="*.jsx" --include="*.tsx" --include="*.mjs" --include="*.cjs" 2>/dev/null; then
            found=true
        fi
        if grep -rq "import ['\"]$search_pattern" "$project_dir" --include="*.js" --include="*.ts" --include="*.jsx" --include="*.tsx" --include="*.mjs" --include="*.cjs" 2>/dev/null; then
            found=true
        fi

        # Check package.json scripts for CLI tools
        if grep -q "\"$dep\"" "$package_json" 2>/dev/null; then
            found=true
        fi

        if [[ "$found" == "false" ]]; then
            unused_deps+=("$dep")
        fi
    done <<< "$deps"

    # Check dev dependencies
    while IFS= read -r dep; do
        [[ -z "$dep" ]] && continue

        local search_pattern="$dep"
        if [[ "$dep" == @*/* ]]; then
            search_pattern=$(echo "$dep" | sed 's/@/\\@/g; s/\//\\//g')
        fi

        local found=false

        # Dev deps might be used in test files or build configs
        if grep -rq "$search_pattern" "$project_dir" --include="*.js" --include="*.ts" --include="*.json" --include="*.config.*" 2>/dev/null; then
            found=true
        fi

        if [[ "$found" == "false" ]]; then
            unused_dev_deps+=("$dep")
        fi
    done <<< "$dev_deps"

    # Convert to JSON
    local unused_deps_json=$(printf '%s\n' "${unused_deps[@]}" | jq -R . | jq -s '.')
    local unused_dev_deps_json=$(printf '%s\n' "${unused_dev_deps[@]}" | jq -R . | jq -s '.')

    echo "{
        \"unused_dependencies\": $unused_deps_json,
        \"unused_devDependencies\": $unused_dev_deps_json,
        \"tool\": \"manual_analysis\",
        \"note\": \"Manual analysis may have false positives for dynamically loaded modules\"
    }" | jq '.'
}

#############################################################################
# Python Unused Dependency Detection
#############################################################################

# Detect unused Python dependencies
# Usage: detect_unused_python <project_dir>
detect_unused_python() {
    local project_dir="$1"

    if [[ ! -d "$project_dir" ]]; then
        echo '{"error": "project_directory_not_found"}'
        return 1
    fi

    # Find requirements file
    local req_file=""
    if [[ -f "$project_dir/requirements.txt" ]]; then
        req_file="$project_dir/requirements.txt"
    elif [[ -f "$project_dir/pyproject.toml" ]]; then
        req_file="$project_dir/pyproject.toml"
    elif [[ -f "$project_dir/setup.py" ]]; then
        req_file="$project_dir/setup.py"
    else
        echo '{"error": "no_requirements_file_found"}'
        return 1
    fi

    # Check if pipreqs is available
    if command -v pipreqs &> /dev/null; then
        # Generate requirements from actual imports
        local temp_file=$(mktemp)
        pipreqs "$project_dir" --print --force 2>/dev/null > "$temp_file" || true

        if [[ -f "$req_file" && "$req_file" == *requirements.txt ]]; then
            # Compare declared vs actual
            local declared=$(cut -d'=' -f1 "$req_file" 2>/dev/null | cut -d'>' -f1 | cut -d'<' -f1 | cut -d'~' -f1 | tr '[:upper:]' '[:lower:]' | sort -u)
            local actual=$(cut -d'=' -f1 "$temp_file" 2>/dev/null | tr '[:upper:]' '[:lower:]' | sort -u)

            # Find unused (declared but not in actual)
            local unused=$(comm -23 <(echo "$declared") <(echo "$actual"))

            rm -f "$temp_file"

            local unused_json=$(echo "$unused" | jq -R . | jq -s '.')

            echo "{
                \"unused_dependencies\": $unused_json,
                \"tool\": \"pipreqs\"
            }" | jq '.'
        else
            rm -f "$temp_file"
            detect_unused_python_manual "$project_dir"
        fi
    else
        detect_unused_python_manual "$project_dir"
    fi
}

# Manual Python unused detection
# Usage: detect_unused_python_manual <project_dir>
detect_unused_python_manual() {
    local project_dir="$1"

    local req_file="$project_dir/requirements.txt"
    if [[ ! -f "$req_file" ]]; then
        echo '{"error": "no_requirements_txt", "tool": "manual_analysis"}'
        return 1
    fi

    # Get declared packages (normalize names)
    local declared=$(cut -d'=' -f1 "$req_file" 2>/dev/null | cut -d'>' -f1 | cut -d'<' -f1 | cut -d'~' -f1 | cut -d'[' -f1 | tr '[:upper:]' '[:lower:]' | tr '_' '-' | grep -v '^#' | grep -v '^$')

    local unused=()

    # Check each package
    while IFS= read -r pkg; do
        [[ -z "$pkg" ]] && continue

        # Normalize package name for import (replace - with _)
        local import_name=$(echo "$pkg" | tr '-' '_')

        # Search for imports
        local found=false
        if grep -rq "^import $import_name" "$project_dir" --include="*.py" 2>/dev/null; then
            found=true
        fi
        if grep -rq "^from $import_name" "$project_dir" --include="*.py" 2>/dev/null; then
            found=true
        fi
        # Also check for partial imports like "from pkg.module"
        if grep -rq "^from $import_name\." "$project_dir" --include="*.py" 2>/dev/null; then
            found=true
        fi

        if [[ "$found" == "false" ]]; then
            unused+=("$pkg")
        fi
    done <<< "$declared"

    local unused_json=$(printf '%s\n' "${unused[@]}" | jq -R . | jq -s '.')

    echo "{
        \"unused_dependencies\": $unused_json,
        \"tool\": \"manual_analysis\",
        \"note\": \"Manual analysis may miss dynamically imported modules\"
    }" | jq '.'
}

#############################################################################
# Go Unused Dependency Detection
#############################################################################

# Detect unused Go modules
# Usage: detect_unused_go <project_dir>
detect_unused_go() {
    local project_dir="$1"

    if [[ ! -f "$project_dir/go.mod" ]]; then
        echo '{"error": "no_go_mod_found"}'
        return 1
    fi

    # Check if go is available
    if command -v go &> /dev/null; then
        # Use go mod tidy in dry-run mode to detect unused
        local temp_dir=$(mktemp -d)
        cp "$project_dir/go.mod" "$temp_dir/"
        [[ -f "$project_dir/go.sum" ]] && cp "$project_dir/go.sum" "$temp_dir/"

        local result=$(cd "$project_dir" && go mod why -m all 2>/dev/null || echo "")

        # Find modules that are "not needed"
        local unused=$(echo "$result" | grep -B1 "not needed" | grep -v "not needed" | grep -v "^--$" | tr -d '()')

        local unused_json=$(echo "$unused" | jq -R . | jq -s '.')

        rm -rf "$temp_dir"

        echo "{
            \"unused_dependencies\": $unused_json,
            \"tool\": \"go_mod_why\"
        }" | jq '.'
    else
        echo '{"error": "go_not_installed"}'
    fi
}

#############################################################################
# Main Entry Point
#############################################################################

# Detect unused dependencies (auto-detect ecosystem)
# Usage: detect_unused_dependencies <project_dir> [ecosystem]
detect_unused_dependencies() {
    local project_dir="$1"
    local ecosystem="${2:-auto}"

    if [[ ! -d "$project_dir" ]]; then
        echo '{"error": "project_directory_not_found"}'
        return 1
    fi

    # Auto-detect ecosystem if not specified
    if [[ "$ecosystem" == "auto" ]]; then
        if [[ -f "$project_dir/package.json" ]]; then
            ecosystem="npm"
        elif [[ -f "$project_dir/requirements.txt" || -f "$project_dir/pyproject.toml" || -f "$project_dir/setup.py" ]]; then
            ecosystem="python"
        elif [[ -f "$project_dir/go.mod" ]]; then
            ecosystem="go"
        else
            echo '{"error": "could_not_detect_ecosystem"}'
            return 1
        fi
    fi

    case "$ecosystem" in
        npm|node)
            detect_unused_npm "$project_dir"
            ;;
        python|pypi)
            detect_unused_python "$project_dir"
            ;;
        go|golang)
            detect_unused_go "$project_dir"
            ;;
        *)
            echo '{"error": "unsupported_ecosystem", "ecosystem": "'"$ecosystem"'"}'
            return 1
            ;;
    esac
}

# Generate unused dependencies report with recommendations
# Usage: generate_unused_report <project_dir> [ecosystem]
generate_unused_report() {
    local project_dir="$1"
    local ecosystem="${2:-auto}"

    local result=$(detect_unused_dependencies "$project_dir" "$ecosystem")

    if [[ $(echo "$result" | jq -e '.error' 2>/dev/null) ]]; then
        echo "$result"
        return 1
    fi

    local unused=$(echo "$result" | jq -r '.unused_dependencies // []')
    local unused_count=$(echo "$unused" | jq 'length')

    local unused_dev=$(echo "$result" | jq -r '.unused_devDependencies // []')
    local unused_dev_count=$(echo "$unused_dev" | jq 'length' 2>/dev/null || echo "0")

    local recommendations=()
    local security_impact=""
    local build_impact=""

    if [[ $unused_count -gt 0 ]]; then
        recommendations+=("Remove unused dependencies to reduce attack surface")
        recommendations+=("Run 'npm prune' or equivalent to clean up")
        security_impact="Each unused dependency is a potential attack vector that adds risk without benefit"
        build_impact="Unused dependencies increase install time and bundle size"
    fi

    if [[ $unused_dev_count -gt 0 ]]; then
        recommendations+=("Consider removing unused dev dependencies to speed up CI/CD")
    fi

    local recommendations_json=$(printf '%s\n' "${recommendations[@]}" 2>/dev/null | jq -R . | jq -s '.' || echo "[]")

    echo "{
        \"summary\": {
            \"unused_dependencies_count\": $unused_count,
            \"unused_devDependencies_count\": $unused_dev_count,
            \"total_unused\": $((unused_count + unused_dev_count))
        },
        \"unused_dependencies\": $unused,
        \"unused_devDependencies\": $unused_dev,
        \"impact\": {
            \"security\": \"$security_impact\",
            \"build\": \"$build_impact\"
        },
        \"recommendations\": $recommendations_json,
        \"tool\": $(echo "$result" | jq '.tool')
    }" | jq '.'
}

#############################################################################
# Export Functions
#############################################################################

export -f detect_unused_npm
export -f detect_unused_npm_manual
export -f detect_unused_python
export -f detect_unused_python_manual
export -f detect_unused_go
export -f detect_unused_dependencies
export -f generate_unused_report
