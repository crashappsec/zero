#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Dynamic Pattern Loader for Technology Identification
# Loads technology detection patterns from RAG Markdown files
# Parses patterns.md files for package names, imports, env vars, etc.
#############################################################################

# Get script directory
PATTERN_LOADER_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
UTILS_ROOT="$(dirname "$PATTERN_LOADER_DIR")"
REPO_ROOT="$(dirname "$(dirname "$UTILS_ROOT")")"
RAG_ROOT="$REPO_ROOT/rag/technology-identification"

# Global data structures for loaded patterns (bash 3.2 compatible)
# Using temp directory for pattern storage
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

# Convert package name to safe filename (handle scoped packages with /)
pkg_to_filename() {
    local pkg_name="$1"
    # Replace / with __SLASH__ for safe filenames
    echo "${pkg_name//\//__SLASH__}"
}

# Convert filename back to package name
filename_to_pkg() {
    local filename="$1"
    # Replace __SLASH__ back to /
    echo "${filename//__SLASH__//}"
}

#############################################################################
# Markdown Parsing Functions
#############################################################################

# Extract value after a markdown header pattern like "**Key**: Value"
extract_md_value() {
    local content="$1"
    local key="$2"
    echo "$content" | grep -m1 "^\*\*$key\*\*:" | sed "s/^\*\*$key\*\*: *//"
}

# Extract category path from directory structure
get_category_from_path() {
    local tech_dir="$1"
    # Convert /path/to/rag/technology-identification/ai-ml/apis/anthropic
    # to ai-ml/apis
    local rel_path="${tech_dir#$RAG_ROOT/}"
    # Remove the technology name (last component)
    dirname "$rel_path"
}

# Parse a patterns.md file and extract technology info
parse_patterns_md() {
    local md_file="$1"
    local tech_dir=$(dirname "$md_file")
    local tech_name=$(basename "$tech_dir")

    if [[ ! -f "$md_file" ]]; then
        return 1
    fi

    local content=$(cat "$md_file")

    # Extract main metadata
    local technology=$(echo "$content" | head -5 | grep "^# " | sed 's/^# //')
    local category=$(extract_md_value "$content" "Category")
    local description=$(extract_md_value "$content" "Description")
    local homepage=$(extract_md_value "$content" "Homepage")

    # If category not found in file, derive from path
    if [[ -z "$category" ]]; then
        category=$(get_category_from_path "$tech_dir")
    fi

    # Default confidence
    local confidence=95

    # Create tech info JSON
    local tech_info=$(cat << EOF
{"technology":"$technology","category":"$category","confidence":$confidence,"description":"$description"}
EOF
)

    # Store technology info
    echo "$tech_info" > "$PATTERN_CACHE_DIR/tech/$tech_name.info"

    # Extract package names from "## Package Detection" section
    local in_package_section=false
    local current_ecosystem=""

    while IFS= read -r line; do
        # Check for section headers
        if [[ "$line" =~ ^##[[:space:]]Package ]]; then
            in_package_section=true
            continue
        elif [[ "$line" =~ ^##[[:space:]] ]] && [[ ! "$line" =~ ^##[[:space:]]Package ]]; then
            in_package_section=false
            continue
        fi

        if [[ "$in_package_section" == true ]]; then
            # Check for ecosystem headers (### NPM, ### PYPI, etc.)
            if [[ "$line" =~ ^###[[:space:]](.+)$ ]]; then
                current_ecosystem=$(echo "${BASH_REMATCH[1]}" | tr '[:upper:]' '[:lower:]' | tr -d ' ')
                continue
            fi

            # Check for package names (- `package-name`)
            if [[ "$line" =~ ^-[[:space:]]\`([^\`]+)\` ]]; then
                local pkg_name="${BASH_REMATCH[1]}"
                local safe_name=$(pkg_to_filename "$pkg_name")
                echo "$tech_info" > "$PATTERN_CACHE_DIR/package/$safe_name"
            fi
        fi
    done <<< "$content"

    # Mark technology as loaded
    echo "loaded" > "$PATTERN_CACHE_DIR/tech/$tech_name"

    return 0
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

    # Find all patterns.md files
    while IFS= read -r md_file; do
        if [[ -n "$md_file" ]] && [[ -f "$md_file" ]]; then
            if parse_patterns_md "$md_file"; then
                tech_count=$((tech_count + 1))
            fi
        fi
    done < <(find "$rag_root" -type f -name "patterns.md")

    # Count loaded packages
    pattern_count=$(ls "$PATTERN_CACHE_DIR/package" 2>/dev/null | wc -l | tr -d ' ')

    echo "Loaded $tech_count technologies with $pattern_count package patterns" >&2
    return 0
}

# Load patterns for a specific technology directory
load_technology_patterns() {
    local tech_dir="$1"
    local tech_name=$(basename "$tech_dir")

    init_pattern_cache

    # Skip if already loaded
    if [[ -f "$PATTERN_CACHE_DIR/tech/$tech_name" ]]; then
        return 0
    fi

    local md_file="$tech_dir/patterns.md"
    if [[ -f "$md_file" ]]; then
        parse_patterns_md "$md_file"
    fi

    return 0
}

#############################################################################
# Pattern Matching Functions
#############################################################################

# Match package name against loaded patterns
match_package_name() {
    local package_name="$1"

    init_pattern_cache

    # Convert to safe filename
    local safe_name=$(pkg_to_filename "$package_name")

    # Direct match
    if [[ -f "$PATTERN_CACHE_DIR/package/$safe_name" ]]; then
        cat "$PATTERN_CACHE_DIR/package/$safe_name"
        return 0
    fi

    # Fuzzy matching for scoped packages (@org/package)
    if [[ "$package_name" =~ ^@.*/.* ]]; then
        local base_name=$(echo "$package_name" | sed 's|^@[^/]*/||')
        local safe_base=$(pkg_to_filename "$base_name")
        if [[ -f "$PATTERN_CACHE_DIR/package/$safe_base" ]]; then
            cat "$PATTERN_CACHE_DIR/package/$safe_base"
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

    # For now, return basic confidence if tech is loaded
    if [[ -f "$PATTERN_CACHE_DIR/tech/$tech_name.info" ]]; then
        echo "90"
        return 0
    fi

    return 1
}

# Check if config file matches patterns
match_config_file() {
    local file_path="$1"
    local tech_name="$2"

    init_pattern_cache

    # For now, return basic confidence if tech is loaded
    if [[ -f "$PATTERN_CACHE_DIR/tech/$tech_name.info" ]]; then
        echo "85"
        return 0
    fi

    return 1
}

# Check if environment variable matches patterns
match_env_variable() {
    local var_name="$1"
    local tech_name="$2"

    init_pattern_cache

    # For now, return basic confidence if tech is loaded
    if [[ -f "$PATTERN_CACHE_DIR/tech/$tech_name.info" ]]; then
        echo "80"
        return 0
    fi

    return 1
}

# Get technology information by name
get_technology_info() {
    local tech_name="$1"

    init_pattern_cache

    if [[ -f "$PATTERN_CACHE_DIR/tech/$tech_name.info" ]]; then
        cat "$PATTERN_CACHE_DIR/tech/$tech_name.info"
        return 0
    fi

    return 1
}

# Get version information for a technology
get_version_info() {
    local tech_name="$1"

    init_pattern_cache

    if [[ -f "$PATTERN_CACHE_DIR/version/$tech_name" ]]; then
        cat "$PATTERN_CACHE_DIR/version/$tech_name"
        return 0
    fi

    return 1
}

# List all loaded technologies
list_loaded_technologies() {
    init_pattern_cache

    if [[ -d "$PATTERN_CACHE_DIR/tech" ]]; then
        ls "$PATTERN_CACHE_DIR/tech" 2>/dev/null | grep -v "\.info$" | sort
    fi
}

# Get pattern statistics
get_pattern_statistics() {
    init_pattern_cache

    local tech_count=$(ls "$PATTERN_CACHE_DIR/tech" 2>/dev/null | grep -v "\.info$" | wc -l | tr -d ' ')
    local package_count=$(ls "$PATTERN_CACHE_DIR/package" 2>/dev/null | wc -l | tr -d ' ')

    cat << EOF
{
  "technologies_loaded": $tech_count,
  "package_patterns": $package_count
}
EOF
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
export -f pkg_to_filename filename_to_pkg
export -f load_all_patterns load_technology_patterns
export -f match_package_name match_import_statement match_config_file match_env_variable
export -f get_technology_info get_version_info
export -f list_loaded_technologies get_pattern_statistics
