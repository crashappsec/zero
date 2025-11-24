#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Dynamic Pattern Loader for Technology Identification
# Loads technology detection patterns from RAG JSON files
# Replaces hardcoded case statements with data-driven detection
#############################################################################

# Get script directory
PATTERN_LOADER_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
UTILS_ROOT="$(dirname "$PATTERN_LOADER_DIR")"
REPO_ROOT="$(dirname "$(dirname "$UTILS_ROOT")")"
RAG_ROOT="$REPO_ROOT/rag/technology-identification"

# Global data structures for loaded patterns (bash 3.2 compatible)
# Using temp directory for pattern storage as bash 3.2 doesn't support associative arrays
PATTERN_CACHE_DIR=""

# Initialize pattern cache
init_pattern_cache() {
    if [[ -z "$PATTERN_CACHE_DIR" ]]; then
        PATTERN_CACHE_DIR=$(mktemp -d)
        mkdir -p "$PATTERN_CACHE_DIR"/{package,import,config,env,version,tech}
    fi
}

# Cleanup pattern cache
cleanup_pattern_cache() {
    if [[ -n "$PATTERN_CACHE_DIR" ]] && [[ -d "$PATTERN_CACHE_DIR" ]]; then
        rm -rf "$PATTERN_CACHE_DIR"
    fi
}

#############################################################################
# Pattern Loading Functions
#############################################################################

# Load all RAG patterns from directory structure
load_all_patterns() {
    local rag_root="${1:-$RAG_ROOT}"

    init_pattern_cache

    if [[ ! -d "$rag_root" ]]; then
        echo "Warning: RAG directory not found: $rag_root" >&2
        return 1
    fi

    local pattern_count=0
    local tech_count=0

    # Find all technology directories (contain package-patterns.json)
    local tech_dirs=$(find "$rag_root" -type f -name "package-patterns.json" -exec dirname {} \;)

    while IFS= read -r tech_dir; do
        if [[ -n "$tech_dir" ]]; then
            # Load patterns for this technology
            if load_technology_patterns "$tech_dir"; then
                tech_count=$((tech_count + 1))
                pattern_count=$((pattern_count + 6))
            fi
        fi
    done <<< "$tech_dirs"

    echo "Loaded $tech_count technologies with $pattern_count pattern files" >&2
    return 0
}

# Load all pattern files for a specific technology
load_technology_patterns() {
    local tech_dir="$1"
    local tech_name=$(basename "$tech_dir")

    init_pattern_cache

    # Skip if already loaded
    if [[ -f "$PATTERN_CACHE_DIR/tech/$tech_name" ]]; then
        return 0
    fi

    # Load each pattern type
    load_package_patterns "$tech_dir/package-patterns.json" "$tech_name"
    load_import_patterns "$tech_dir/import-patterns.json" "$tech_name"
    load_config_patterns "$tech_dir/config-patterns.json" "$tech_name"
    load_env_patterns "$tech_dir/env-patterns.json" "$tech_name"
    load_version_info "$tech_dir/versions.json" "$tech_name"

    # Mark as loaded
    echo "loaded" > "$PATTERN_CACHE_DIR/tech/$tech_name"
    return 0
}

# Load package patterns from JSON file
load_package_patterns() {
    local pattern_file="$1"
    local tech_name="$2"

    if [[ ! -f "$pattern_file" ]]; then
        return 1
    fi

    local tech_info=$(jq -c '{
        technology: .technology,
        category: .category,
        confidence: .confidence,
        description: .description
    }' "$pattern_file" 2>/dev/null)

    if [[ -z "$tech_info" ]] || [[ "$tech_info" == "null" ]]; then
        return 1
    fi

    # Extract all package names and write to cache files
    local package_names=$(jq -r '.patterns[].names[]? // empty' "$pattern_file" 2>/dev/null)

    while IFS= read -r pkg_name; do
        if [[ -n "$pkg_name" ]]; then
            echo "$tech_info" > "$PATTERN_CACHE_DIR/package/$pkg_name"
        fi
    done <<< "$package_names"

    # Also handle related packages with slightly lower confidence
    local related=$(jq -r '.related_packages[]? // empty' "$pattern_file" 2>/dev/null)
    local related_info=$(echo "$tech_info" | jq '.confidence = (.confidence - 10)')

    while IFS= read -r pkg_name; do
        if [[ -n "$pkg_name" ]]; then
            echo "$related_info" > "$PATTERN_CACHE_DIR/package/$pkg_name"
        fi
    done <<< "$related"

    return 0
}

# Load import patterns from JSON file
load_import_patterns() {
    local pattern_file="$1"
    local tech_name="$2"

    if [[ ! -f "$pattern_file" ]]; then
        return 1
    fi

    # Copy patterns to cache
    cp "$pattern_file" "$PATTERN_CACHE_DIR/import/$tech_name.json" 2>/dev/null
    return 0
}

# Load config patterns from JSON file
load_config_patterns() {
    local pattern_file="$1"
    local tech_name="$2"

    if [[ ! -f "$pattern_file" ]]; then
        return 1
    fi

    # Copy patterns to cache
    cp "$pattern_file" "$PATTERN_CACHE_DIR/config/$tech_name.json" 2>/dev/null
    return 0
}

# Load environment variable patterns from JSON file
load_env_patterns() {
    local pattern_file="$1"
    local tech_name="$2"

    if [[ ! -f "$pattern_file" ]]; then
        return 1
    fi

    # Copy patterns to cache
    cp "$pattern_file" "$PATTERN_CACHE_DIR/env/$tech_name.json" 2>/dev/null
    return 0
}

# Load version information from JSON file
load_version_info() {
    local version_file="$1"
    local tech_name="$2"

    if [[ ! -f "$version_file" ]]; then
        return 1
    fi

    # Copy version info to cache
    cp "$version_file" "$PATTERN_CACHE_DIR/version/$tech_name.json" 2>/dev/null
    return 0
}

#############################################################################
# Pattern Matching Functions
#############################################################################

# Match package name against loaded patterns
match_package_name() {
    local package_name="$1"

    init_pattern_cache

    # Direct match
    if [[ -f "$PATTERN_CACHE_DIR/package/$package_name" ]]; then
        cat "$PATTERN_CACHE_DIR/package/$package_name"
        return 0
    fi

    # Fuzzy matching for scoped packages (@org/package)
    if [[ "$package_name" =~ ^@.*/.* ]]; then
        local base_name=$(echo "$package_name" | sed 's|^@[^/]*/||')
        if [[ -f "$PATTERN_CACHE_DIR/package/$base_name" ]]; then
            cat "$PATTERN_CACHE_DIR/package/$base_name"
            return 0
        fi
    fi

    return 1
}

# Check if import statement matches patterns for a technology
match_import_statement() {
    local import_line="$1"
    local tech_name="$2"
    local file_extension="$3"

    init_pattern_cache

    local patterns_file="$PATTERN_CACHE_DIR/import/$tech_name.json"

    if [[ ! -f "$patterns_file" ]]; then
        return 1
    fi

    local patterns=$(cat "$patterns_file")

    # Find patterns for this file extension
    local language_patterns=$(echo "$patterns" | jq -c --arg ext "$file_extension" '
        .patterns[] |
        select(.file_extensions | index($ext)) |
        .patterns[]
    ' 2>/dev/null)

    if [[ -z "$language_patterns" ]]; then
        return 1
    fi

    # Test each pattern
    while IFS= read -r pattern_obj; do
        local regex=$(echo "$pattern_obj" | jq -r '.regex // empty' 2>/dev/null)

        if [[ -n "$regex" ]]; then
            if echo "$import_line" | grep -qE "$regex" 2>/dev/null; then
                # Get confidence from parent
                local confidence=$(echo "$patterns" | jq -r '.confidence' 2>/dev/null)
                echo "$confidence"
                return 0
            fi
        fi
    done <<< "$language_patterns"

    return 1
}

# Check if config file matches patterns
match_config_file() {
    local file_path="$1"
    local tech_name="$2"

    init_pattern_cache

    local patterns_file="$PATTERN_CACHE_DIR/config/$tech_name.json"

    if [[ ! -f "$patterns_file" ]]; then
        return 1
    fi

    local patterns=$(cat "$patterns_file")
    local basename=$(basename "$file_path")

    # Check file pattern matches
    local file_patterns=$(echo "$patterns" | jq -r '.patterns[].file_patterns[]? // empty' 2>/dev/null)

    while IFS= read -r pattern; do
        if [[ -n "$pattern" ]]; then
            # Handle glob patterns
            if [[ "$pattern" == *"*"* ]]; then
                local regex="${pattern//\*/.*}"
                regex="${regex//\?/.}"
                if [[ "$file_path" =~ $regex ]]; then
                    local confidence=$(echo "$patterns" | jq -r '.confidence' 2>/dev/null)
                    echo "$confidence"
                    return 0
                fi
            elif [[ "$basename" == "$pattern" ]] || [[ "$file_path" == *"$pattern"* ]]; then
                local confidence=$(echo "$patterns" | jq -r '.confidence' 2>/dev/null)
                echo "$confidence"
                return 0
            fi
        fi
    done <<< "$file_patterns"

    return 1
}

# Check if environment variable matches patterns
match_env_variable() {
    local var_name="$1"
    local tech_name="$2"

    init_pattern_cache

    local patterns_file="$PATTERN_CACHE_DIR/env/$tech_name.json"

    if [[ ! -f "$patterns_file" ]]; then
        return 1
    fi

    local patterns=$(cat "$patterns_file")

    # Check variable name patterns
    local var_names=$(echo "$patterns" | jq -r '.patterns[].variable_names[]? // empty' 2>/dev/null)

    while IFS= read -r pattern_name; do
        if [[ "$var_name" == "$pattern_name" ]]; then
            local confidence=$(echo "$patterns" | jq -r '.confidence' 2>/dev/null)
            echo "$confidence"
            return 0
        fi
    done <<< "$var_names"

    # Check prefix patterns
    local prefixes=$(echo "$patterns" | jq -r '.patterns[].prefix? // empty' 2>/dev/null)

    while IFS= read -r prefix; do
        if [[ -n "$prefix" ]] && [[ "$var_name" == "$prefix"* ]]; then
            local confidence=$(echo "$patterns" | jq -r '.confidence' 2>/dev/null)
            echo "$confidence"
            return 0
        fi
    done <<< "$prefixes"

    return 1
}

# Get technology information by name
get_technology_info() {
    local tech_name="$1"

    init_pattern_cache

    # Try to find from package patterns first
    local pkg_file=$(find "$PATTERN_CACHE_DIR/package" -type f -exec grep -l "\"technology\":\"$tech_name\"" {} \; 2>/dev/null | head -1)

    if [[ -n "$pkg_file" ]]; then
        cat "$pkg_file"
        return 0
    fi

    # Try import patterns
    if [[ -f "$PATTERN_CACHE_DIR/import/$tech_name.json" ]]; then
        jq -c '{
            technology: .technology,
            category: .category,
            confidence: .confidence
        }' "$PATTERN_CACHE_DIR/import/$tech_name.json" 2>/dev/null
        return 0
    fi

    return 1
}

# Get version information for a technology
get_version_info() {
    local tech_name="$1"

    init_pattern_cache

    if [[ -f "$PATTERN_CACHE_DIR/version/$tech_name.json" ]]; then
        cat "$PATTERN_CACHE_DIR/version/$tech_name.json"
        return 0
    fi

    return 1
}

# List all loaded technologies
list_loaded_technologies() {
    init_pattern_cache

    if [[ -d "$PATTERN_CACHE_DIR/tech" ]]; then
        ls "$PATTERN_CACHE_DIR/tech" 2>/dev/null | sort
    fi
}

# Get pattern statistics
get_pattern_statistics() {
    init_pattern_cache

    local tech_count=$(ls "$PATTERN_CACHE_DIR/tech" 2>/dev/null | wc -l | tr -d ' ')
    local package_count=$(ls "$PATTERN_CACHE_DIR/package" 2>/dev/null | wc -l | tr -d ' ')
    local import_count=$(ls "$PATTERN_CACHE_DIR/import" 2>/dev/null | wc -l | tr -d ' ')
    local config_count=$(ls "$PATTERN_CACHE_DIR/config" 2>/dev/null | wc -l | tr -d ' ')
    local env_count=$(ls "$PATTERN_CACHE_DIR/env" 2>/dev/null | wc -l | tr -d ' ')

    jq -n \
        --argjson tech_count "$tech_count" \
        --argjson package_count "$package_count" \
        --argjson import_count "$import_count" \
        --argjson config_count "$config_count" \
        --argjson env_count "$env_count" \
        '{
            technologies_loaded: $tech_count,
            package_patterns: $package_count,
            import_patterns: $import_count,
            config_patterns: $config_count,
            env_patterns: $env_count
        }'
}

#############################################################################
# Initialization
#############################################################################

# Auto-load patterns when library is sourced
if [[ "${BASH_SOURCE[0]}" != "${0}" ]]; then
    # Library is being sourced, load patterns automatically
    load_all_patterns "$RAG_ROOT" 2>/dev/null || true
fi

# Export functions
export -f load_all_patterns load_technology_patterns
export -f match_package_name match_import_statement match_config_file match_env_variable
export -f get_technology_info get_version_info
export -f list_loaded_technologies get_pattern_statistics
