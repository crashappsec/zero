#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Unified File Scanner Library
# Shared file discovery and pattern matching for all scanners
#
# Features:
# - Centralized file extension and exclusion lists
# - Efficient file discovery with caching
# - Batch pattern matching with ripgrep
# - Parallel processing support
# - Consistent behavior across all scanners
#############################################################################

# Ensure we only source this once
[[ -n "$_FILE_SCANNER_LOADED" ]] && return 0
_FILE_SCANNER_LOADED=true

#############################################################################
# Constants - Single Source of Truth
#############################################################################

# Source code file extensions (space-separated for bash 3.x compatibility)
SOURCE_EXTENSIONS="py js ts jsx tsx java go rb php c cpp h hpp cs swift kt rs scala sh bash"

# Config/infrastructure file patterns
CONFIG_FILES="Dockerfile docker-compose.yml docker-compose.yaml package.json requirements.txt pyproject.toml go.mod Gemfile pom.xml build.gradle Cargo.toml"

# Documentation file extensions
DOC_EXTENSIONS="md rst txt adoc"

# Directories to always exclude
EXCLUDE_DIRS="node_modules vendor .git dist build __pycache__ .venv venv .cache .next .nuxt coverage .tox .pytest_cache .mypy_cache .ruff_cache target"

# File patterns to exclude
EXCLUDE_PATTERNS="*.min.js *.bundle.js *.map *.lock package-lock.json yarn.lock pnpm-lock.yaml Gemfile.lock poetry.lock"

#############################################################################
# File Discovery Functions
#############################################################################

# Build find exclusion arguments
# Usage: _build_find_excludes
# Returns: find-compatible exclusion arguments
_build_find_excludes() {
    local excludes=""
    for dir in $EXCLUDE_DIRS; do
        excludes="$excludes -path \"*/$dir/*\" -prune -o"
    done
    echo "$excludes"
}

# Get source files in a directory
# Usage: get_source_files <directory> [extensions]
# Returns: Newline-separated list of file paths
get_source_files() {
    local dir="$1"
    local extensions="${2:-$SOURCE_EXTENSIONS}"

    [[ ! -d "$dir" ]] && return 1

    # Build -name arguments for extensions
    local name_args=""
    local first=true
    for ext in $extensions; do
        if $first; then
            name_args="-name \"*.$ext\""
            first=false
        else
            name_args="$name_args -o -name \"*.$ext\""
        fi
    done

    # Build exclusion paths
    local exclude_args=""
    for exc_dir in $EXCLUDE_DIRS; do
        exclude_args="$exclude_args ! -path \"*/$exc_dir/*\""
    done

    # Execute find with eval to handle quoting
    eval "find \"$dir\" -type f \\( $name_args \\) $exclude_args 2>/dev/null"
}

# Get config files in a directory
# Usage: get_config_files <directory>
# Returns: Newline-separated list of config file paths
get_config_files() {
    local dir="$1"

    [[ ! -d "$dir" ]] && return 1

    local name_args=""
    local first=true
    for file in $CONFIG_FILES; do
        if $first; then
            name_args="-name \"$file\""
            first=false
        else
            name_args="$name_args -o -name \"$file\""
        fi
    done

    # Build exclusion paths
    local exclude_args=""
    for exc_dir in $EXCLUDE_DIRS; do
        exclude_args="$exclude_args ! -path \"*/$exc_dir/*\""
    done

    eval "find \"$dir\" -type f \\( $name_args \\) $exclude_args 2>/dev/null"
}

# Get all scannable files (source + config)
# Usage: get_all_files <directory>
get_all_files() {
    local dir="$1"

    [[ ! -d "$dir" ]] && return 1

    {
        get_source_files "$dir"
        get_config_files "$dir"
    } | sort -u
}

# Get file language from extension
# Usage: get_file_language <file_path>
# Returns: Language name (lowercase)
get_file_language() {
    local file="$1"
    local ext="${file##*.}"

    case "$ext" in
        py) echo "python" ;;
        js|jsx|mjs|cjs) echo "javascript" ;;
        ts|tsx) echo "typescript" ;;
        java) echo "java" ;;
        go) echo "go" ;;
        rb) echo "ruby" ;;
        php) echo "php" ;;
        c|h) echo "c" ;;
        cpp|hpp|cc|cxx) echo "cpp" ;;
        cs) echo "csharp" ;;
        swift) echo "swift" ;;
        kt|kts) echo "kotlin" ;;
        rs) echo "rust" ;;
        scala) echo "scala" ;;
        sh|bash|zsh) echo "shell" ;;
        md|markdown) echo "markdown" ;;
        json) echo "json" ;;
        yaml|yml) echo "yaml" ;;
        xml) echo "xml" ;;
        *) echo "unknown" ;;
    esac
}

#############################################################################
# Pattern Matching Functions (Ripgrep-based)
#############################################################################

# Check if ripgrep is available
has_ripgrep() {
    command -v rg &>/dev/null
}

# Batch pattern search with ripgrep
# Usage: batch_grep_patterns <directory> <pattern1> [pattern2] ...
# Returns: JSON array of matches with file, line, pattern info
batch_grep_patterns() {
    local dir="$1"
    shift
    local patterns=("$@")

    [[ ${#patterns[@]} -eq 0 ]] && { echo "[]"; return; }

    if has_ripgrep; then
        _batch_grep_rg "$dir" "${patterns[@]}"
    else
        _batch_grep_fallback "$dir" "${patterns[@]}"
    fi
}

# Ripgrep implementation
_batch_grep_rg() {
    local dir="$1"
    shift
    local patterns=("$@")

    # Build ripgrep args
    local rg_args=("--json" "--no-heading" "--line-number")

    # Add exclusions
    for exc_dir in $EXCLUDE_DIRS; do
        rg_args+=("-g" "!$exc_dir/**")
    done

    # Add patterns (use -e for each)
    for pattern in "${patterns[@]}"; do
        rg_args+=("-e" "$pattern")
    done

    # Run ripgrep and convert to simplified JSON
    rg "${rg_args[@]}" "$dir" 2>/dev/null | jq -c -s '
        [.[] | select(.type == "match") | {
            file: .data.path.text,
            line_number: .data.line_number,
            line_text: .data.lines.text,
            submatches: [.data.submatches[] | {
                match: .match.text,
                start: .start,
                end: .end
            }]
        }]
    ' 2>/dev/null || echo "[]"
}

# Fallback grep implementation (slower)
_batch_grep_fallback() {
    local dir="$1"
    shift
    local patterns=("$@")
    local results="[]"

    # Build combined pattern
    local combined_pattern
    combined_pattern=$(IFS='|'; echo "${patterns[*]}")

    # Get files and search
    local files
    files=$(get_source_files "$dir")

    while IFS= read -r file; do
        [[ -z "$file" ]] && continue

        local matches
        matches=$(grep -n -E "$combined_pattern" "$file" 2>/dev/null || true)

        while IFS= read -r match_line; do
            [[ -z "$match_line" ]] && continue
            local line_num="${match_line%%:*}"
            local line_text="${match_line#*:}"

            local entry
            entry=$(jq -n \
                --arg file "$file" \
                --argjson line "$line_num" \
                --arg text "$line_text" \
                '{file: $file, line_number: $line, line_text: $text, submatches: []}')

            results=$(echo "$results" | jq --argjson e "$entry" '. + [$e]')
        done <<< "$matches"
    done <<< "$files"

    echo "$results"
}

# Search for files matching a pattern (by content)
# Usage: files_matching_pattern <directory> <pattern> [file_glob]
# Returns: Newline-separated list of matching files
files_matching_pattern() {
    local dir="$1"
    local pattern="$2"
    local glob="${3:-}"

    if has_ripgrep; then
        local rg_args=("-l" "--no-heading")

        # Add exclusions
        for exc_dir in $EXCLUDE_DIRS; do
            rg_args+=("-g" "!$exc_dir/**")
        done

        # Add glob filter if provided
        [[ -n "$glob" ]] && rg_args+=("-g" "$glob")

        rg "${rg_args[@]}" "$pattern" "$dir" 2>/dev/null || true
    else
        # Fallback to grep
        get_source_files "$dir" | xargs grep -l -E "$pattern" 2>/dev/null || true
    fi
}

# Count matches for a pattern
# Usage: count_pattern_matches <directory> <pattern>
# Returns: Integer count
count_pattern_matches() {
    local dir="$1"
    local pattern="$2"

    if has_ripgrep; then
        local rg_args=("--count-matches" "--no-heading")
        for exc_dir in $EXCLUDE_DIRS; do
            rg_args+=("-g" "!$exc_dir/**")
        done
        rg "${rg_args[@]}" "$pattern" "$dir" 2>/dev/null | awk -F: '{sum += $2} END {print sum+0}'
    else
        get_source_files "$dir" | xargs grep -c -E "$pattern" 2>/dev/null | awk -F: '{sum += $2} END {print sum+0}'
    fi
}

#############################################################################
# Parallel Processing Functions
#############################################################################

# Get optimal worker count for parallel processing
# Usage: get_worker_count
# Returns: Integer (number of CPU cores, capped at 8)
get_worker_count() {
    local cores
    cores=$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 4)
    # Cap at 8 to avoid overwhelming the system
    [[ $cores -gt 8 ]] && cores=8
    echo "$cores"
}

# Process files in parallel with a callback
# Usage: parallel_file_process <directory> <callback_function> [worker_count]
# Note: callback_function receives file path as argument
parallel_file_process() {
    local dir="$1"
    local callback="$2"
    local workers="${3:-$(get_worker_count)}"

    # Export callback function for xargs subprocess
    export -f "$callback" 2>/dev/null || true

    get_source_files "$dir" | xargs -P "$workers" -I {} bash -c "$callback \"{}\""
}

# Batch process patterns across files in parallel
# Usage: parallel_pattern_scan <directory> <pattern_file> [worker_count]
# pattern_file: One pattern per line
# Returns: JSON array of all matches
parallel_pattern_scan() {
    local dir="$1"
    local pattern_file="$2"
    local workers="${3:-$(get_worker_count)}"

    [[ ! -f "$pattern_file" ]] && { echo "[]"; return 1; }

    if has_ripgrep; then
        local rg_args=("--json" "--no-heading" "-f" "$pattern_file")

        for exc_dir in $EXCLUDE_DIRS; do
            rg_args+=("-g" "!$exc_dir/**")
        done

        # Use ripgrep's built-in parallelism
        rg "${rg_args[@]}" "$dir" 2>/dev/null | jq -c -s '
            [.[] | select(.type == "match") | {
                file: .data.path.text,
                line_number: .data.line_number,
                line_text: .data.lines.text
            }]
        ' 2>/dev/null || echo "[]"
    else
        # Fallback: simple parallel grep
        local results="[]"
        local combined_pattern
        combined_pattern=$(cat "$pattern_file" | paste -sd '|' -)

        get_source_files "$dir" | xargs -P "$workers" grep -n -E "$combined_pattern" 2>/dev/null | while IFS= read -r line; do
            local file="${line%%:*}"
            local rest="${line#*:}"
            local line_num="${rest%%:*}"
            local text="${rest#*:}"

            jq -n --arg f "$file" --argjson n "$line_num" --arg t "$text" \
                '{file: $f, line_number: $n, line_text: $t}'
        done | jq -s '.'
    fi
}

#############################################################################
# File Caching (for multi-scanner efficiency)
#############################################################################

# Global file cache
_FILE_CACHE_DIR=""
_FILE_CACHE_VALID=false

# Initialize file cache for a directory
# Usage: init_file_cache <directory>
init_file_cache() {
    local dir="$1"
    _FILE_CACHE_DIR=$(mktemp -d)

    # Cache source files
    get_source_files "$dir" > "$_FILE_CACHE_DIR/source_files.txt"

    # Cache config files
    get_config_files "$dir" > "$_FILE_CACHE_DIR/config_files.txt"

    # Mark cache as valid
    _FILE_CACHE_VALID=true

    echo "$_FILE_CACHE_DIR"
}

# Get cached source files
# Usage: get_cached_source_files
get_cached_source_files() {
    if [[ "$_FILE_CACHE_VALID" == true ]] && [[ -f "$_FILE_CACHE_DIR/source_files.txt" ]]; then
        cat "$_FILE_CACHE_DIR/source_files.txt"
    else
        return 1
    fi
}

# Get cached config files
# Usage: get_cached_config_files
get_cached_config_files() {
    if [[ "$_FILE_CACHE_VALID" == true ]] && [[ -f "$_FILE_CACHE_DIR/config_files.txt" ]]; then
        cat "$_FILE_CACHE_DIR/config_files.txt"
    else
        return 1
    fi
}

# Cleanup file cache
# Usage: cleanup_file_cache
cleanup_file_cache() {
    if [[ -n "$_FILE_CACHE_DIR" ]] && [[ -d "$_FILE_CACHE_DIR" ]]; then
        rm -rf "$_FILE_CACHE_DIR"
    fi
    _FILE_CACHE_VALID=false
}

#############################################################################
# Statistics Functions
#############################################################################

# Get file statistics for a directory
# Usage: get_file_stats <directory>
# Returns: JSON with file counts by type
get_file_stats() {
    local dir="$1"
    local stats='{"source_files": 0, "config_files": 0, "total_lines": 0}'

    # Count source files
    local source_count
    source_count=$(get_source_files "$dir" | wc -l | tr -d ' ')

    # Count config files
    local config_count
    config_count=$(get_config_files "$dir" | wc -l | tr -d ' ')

    # Count total lines (approximation using wc -l)
    local total_lines=0
    if [[ $source_count -lt 5000 ]]; then
        total_lines=$(get_source_files "$dir" | head -1000 | xargs wc -l 2>/dev/null | tail -1 | awk '{print $1}')
    fi

    jq -n \
        --argjson src "$source_count" \
        --argjson cfg "$config_count" \
        --argjson lines "$total_lines" \
        '{source_files: $src, config_files: $cfg, total_lines: $lines}'
}

# Get extension distribution
# Usage: get_extension_distribution <directory>
# Returns: JSON object with extension counts
get_extension_distribution() {
    local dir="$1"

    get_source_files "$dir" | \
        sed 's/.*\.//' | \
        sort | \
        uniq -c | \
        sort -rn | \
        head -20 | \
        awk '{print "{\"" $2 "\": " $1 "}"}' | \
        jq -s 'add // {}'
}

#############################################################################
# Export Functions
#############################################################################

export -f get_source_files
export -f get_config_files
export -f get_all_files
export -f get_file_language
export -f has_ripgrep
export -f batch_grep_patterns
export -f files_matching_pattern
export -f count_pattern_matches
export -f get_worker_count
export -f parallel_file_process
export -f parallel_pattern_scan
export -f init_file_cache
export -f get_cached_source_files
export -f get_cached_config_files
export -f cleanup_file_cache
export -f get_file_stats
export -f get_extension_distribution

# Export constants
export SOURCE_EXTENSIONS
export CONFIG_FILES
export DOC_EXTENSIONS
export EXCLUDE_DIRS
export EXCLUDE_PATTERNS
