#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# File Scanner Library
# Functions for scanning and filtering source files for security analysis
#############################################################################

# Security-relevant file extensions by language
declare -A SECURITY_FILE_EXTENSIONS=(
    ["python"]="py"
    ["javascript"]="js,jsx,mjs,cjs"
    ["typescript"]="ts,tsx"
    ["java"]="java"
    ["go"]="go"
    ["ruby"]="rb"
    ["php"]="php"
    ["csharp"]="cs"
    ["rust"]="rs"
    ["shell"]="sh,bash"
    ["sql"]="sql"
    ["kotlin"]="kt,kts"
    ["swift"]="swift"
    ["scala"]="scala"
)

# Configuration file patterns (often contain secrets)
CONFIG_FILE_PATTERNS=(
    "*.json"
    "*.yaml"
    "*.yml"
    "*.toml"
    "*.ini"
    "*.cfg"
    "*.conf"
    "*.config"
    "*.properties"
    ".env*"
    "Dockerfile*"
    "docker-compose*.yml"
)

# Default exclusion patterns
DEFAULT_EXCLUSIONS=(
    "node_modules/**"
    "vendor/**"
    ".git/**"
    "__pycache__/**"
    "*.pyc"
    ".venv/**"
    "venv/**"
    "dist/**"
    "build/**"
    ".next/**"
    "target/**"
    "*.min.js"
    "*.min.css"
    "*.map"
    "package-lock.json"
    "yarn.lock"
    "Gemfile.lock"
    "poetry.lock"
    "Cargo.lock"
    "go.sum"
)

# Get all supported file extensions as find pattern
get_find_extensions() {
    local extensions=""
    for lang in "${!SECURITY_FILE_EXTENSIONS[@]}"; do
        IFS=',' read -ra exts <<< "${SECURITY_FILE_EXTENSIONS[$lang]}"
        for ext in "${exts[@]}"; do
            if [[ -n "$extensions" ]]; then
                extensions="$extensions -o"
            fi
            extensions="$extensions -name \"*.$ext\""
        done
    done
    echo "$extensions"
}

# Build exclusion arguments for find
build_exclusion_args() {
    local custom_exclusions="$1"
    local args=""

    # Add default exclusions
    for pattern in "${DEFAULT_EXCLUSIONS[@]}"; do
        args="$args -not -path \"*/$pattern\""
    done

    # Add custom exclusions
    if [[ -n "$custom_exclusions" ]]; then
        IFS=',' read -ra customs <<< "$custom_exclusions"
        for pattern in "${customs[@]}"; do
            args="$args -not -path \"*/$pattern\""
        done
    fi

    echo "$args"
}

# Scan directory for security-relevant files
# Usage: scan_for_files <directory> [exclusions] [max_files]
scan_for_files() {
    local dir="$1"
    local exclusions="${2:-}"
    local max_files="${3:-500}"

    local exclusion_args
    exclusion_args=$(build_exclusion_args "$exclusions")

    # Build find command
    local find_cmd="find \"$dir\" -type f \\( -name \"*.py\" -o -name \"*.js\" -o -name \"*.jsx\" -o -name \"*.ts\" -o -name \"*.tsx\" -o -name \"*.java\" -o -name \"*.go\" -o -name \"*.rb\" -o -name \"*.php\" -o -name \"*.cs\" -o -name \"*.rs\" -o -name \"*.sh\" -o -name \"*.sql\" -o -name \"*.kt\" -o -name \"*.swift\" -o -name \"*.scala\" \\) $exclusion_args"

    # Execute and limit results
    eval "$find_cmd" 2>/dev/null | head -n "$max_files"
}

# Scan for configuration files (potential secrets)
scan_for_config_files() {
    local dir="$1"
    local exclusions="${2:-}"
    local max_files="${3:-100}"

    local exclusion_args
    exclusion_args=$(build_exclusion_args "$exclusions")

    # Build find command for config files
    local find_cmd="find \"$dir\" -type f \\( -name \"*.json\" -o -name \"*.yaml\" -o -name \"*.yml\" -o -name \"*.toml\" -o -name \"*.ini\" -o -name \"*.cfg\" -o -name \"*.conf\" -o -name \"*.properties\" -o -name \".env*\" -o -name \"Dockerfile*\" \\) $exclusion_args"

    eval "$find_cmd" 2>/dev/null | head -n "$max_files"
}

# Get file language from extension
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
        cs) echo "csharp" ;;
        rs) echo "rust" ;;
        sh|bash) echo "shell" ;;
        sql) echo "sql" ;;
        kt|kts) echo "kotlin" ;;
        swift) echo "swift" ;;
        scala) echo "scala" ;;
        json|yaml|yml|toml|ini|cfg|conf|properties) echo "config" ;;
        *) echo "unknown" ;;
    esac
}

# Filter files by size (skip very large files)
filter_by_size() {
    local max_size="${1:-100000}"  # 100KB default

    while IFS= read -r file; do
        if [[ -f "$file" ]]; then
            local size
            size=$(wc -c < "$file" 2>/dev/null || echo "0")
            if [[ "$size" -le "$max_size" ]]; then
                echo "$file"
            fi
        fi
    done
}

# Filter files by line count
filter_by_lines() {
    local max_lines="${1:-5000}"

    while IFS= read -r file; do
        if [[ -f "$file" ]]; then
            local lines
            lines=$(wc -l < "$file" 2>/dev/null || echo "0")
            if [[ "$lines" -le "$max_lines" ]]; then
                echo "$file"
            fi
        fi
    done
}

# Prioritize files for analysis (entry points, routes, auth, etc.)
prioritize_files() {
    local -a high_priority=()
    local -a medium_priority=()
    local -a low_priority=()

    while IFS= read -r file; do
        local basename
        basename=$(basename "$file")
        local dirname
        dirname=$(dirname "$file")

        # High priority: auth, routes, controllers, API endpoints
        if [[ "$basename" =~ (auth|login|session|token|password|credential|secret) ]] ||
           [[ "$dirname" =~ (auth|api|routes|controllers|handlers|endpoints) ]]; then
            high_priority+=("$file")
        # Medium priority: models, services, utils
        elif [[ "$dirname" =~ (models|services|utils|helpers|middleware) ]]; then
            medium_priority+=("$file")
        # Low priority: everything else
        else
            low_priority+=("$file")
        fi
    done

    # Output in priority order
    printf '%s\n' "${high_priority[@]}"
    printf '%s\n' "${medium_priority[@]}"
    printf '%s\n' "${low_priority[@]}"
}

# Count files by language
count_files_by_language() {
    local dir="$1"

    echo "Files by language:"
    for lang in "${!SECURITY_FILE_EXTENSIONS[@]}"; do
        IFS=',' read -ra exts <<< "${SECURITY_FILE_EXTENSIONS[$lang]}"
        local count=0
        for ext in "${exts[@]}"; do
            local c
            c=$(find "$dir" -type f -name "*.$ext" 2>/dev/null | wc -l)
            count=$((count + c))
        done
        if [[ "$count" -gt 0 ]]; then
            printf "  %-12s: %d\n" "$lang" "$count"
        fi
    done
}
