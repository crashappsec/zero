#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Context Builder Library
# Build code context for Claude security analysis
#############################################################################

# Maximum context size (in characters)
MAX_CONTEXT_SIZE="${MAX_CONTEXT_SIZE:-100000}"
MAX_FILE_SIZE="${MAX_FILE_SIZE:-50000}"

# Build context for a single file
# Usage: build_file_context <file_path> [include_imports]
build_file_context() {
    local file="$1"
    local include_imports="${2:-true}"

    if [[ ! -f "$file" ]]; then
        echo "Error: File not found: $file" >&2
        return 1
    fi

    local content
    content=$(cat "$file" 2>/dev/null)

    # Truncate if too large
    if [[ ${#content} -gt $MAX_FILE_SIZE ]]; then
        content="${content:0:$MAX_FILE_SIZE}

... [truncated - file exceeds $MAX_FILE_SIZE characters]"
    fi

    local language
    language=$(get_file_language "$file" 2>/dev/null || echo "unknown")

    cat <<EOF
## File: $file
Language: $language
Lines: $(echo "$content" | wc -l | tr -d ' ')

\`\`\`$language
$content
\`\`\`
EOF
}

# Build context for multiple files
# Usage: build_multi_file_context <file1> <file2> ...
build_multi_file_context() {
    local total_size=0
    local context=""

    for file in "$@"; do
        if [[ ! -f "$file" ]]; then
            continue
        fi

        local file_context
        file_context=$(build_file_context "$file")
        local file_size=${#file_context}

        # Check if adding this file would exceed limit
        if [[ $((total_size + file_size)) -gt $MAX_CONTEXT_SIZE ]]; then
            echo "$context"
            echo ""
            echo "... [additional files truncated - context limit reached]"
            return 0
        fi

        context="$context$file_context

---

"
        total_size=$((total_size + file_size))
    done

    echo "$context"
}

# Extract imports from a file
extract_imports() {
    local file="$1"
    local language
    language=$(get_file_language "$file" 2>/dev/null || echo "unknown")

    case "$language" in
        python)
            grep -E "^(import|from)" "$file" 2>/dev/null | head -50
            ;;
        javascript|typescript)
            grep -E "^(import|require|export)" "$file" 2>/dev/null | head -50
            ;;
        java|kotlin|scala)
            grep -E "^import" "$file" 2>/dev/null | head -50
            ;;
        go)
            # Go imports can be multi-line
            sed -n '/^import/,/^)/p' "$file" 2>/dev/null | head -50
            ;;
        ruby)
            grep -E "^require" "$file" 2>/dev/null | head -50
            ;;
        php)
            grep -E "^(use|require|include)" "$file" 2>/dev/null | head -50
            ;;
        csharp)
            grep -E "^using" "$file" 2>/dev/null | head -50
            ;;
        rust)
            grep -E "^use" "$file" 2>/dev/null | head -50
            ;;
        *)
            echo ""
            ;;
    esac
}

# Extract function signatures from a file
extract_function_signatures() {
    local file="$1"
    local language
    language=$(get_file_language "$file" 2>/dev/null || echo "unknown")

    case "$language" in
        python)
            grep -E "^(def|async def|class)" "$file" 2>/dev/null | head -30
            ;;
        javascript|typescript)
            grep -E "(function|const.*=.*=>|class|async.*function)" "$file" 2>/dev/null | head -30
            ;;
        java|kotlin|scala)
            grep -E "(public|private|protected|static).*\(" "$file" 2>/dev/null | head -30
            ;;
        go)
            grep -E "^func" "$file" 2>/dev/null | head -30
            ;;
        ruby)
            grep -E "^(def|class|module)" "$file" 2>/dev/null | head -30
            ;;
        php)
            grep -E "(function|class)" "$file" 2>/dev/null | head -30
            ;;
        csharp)
            grep -E "(public|private|protected|internal).*\(" "$file" 2>/dev/null | head -30
            ;;
        rust)
            grep -E "^(pub )?fn" "$file" 2>/dev/null | head -30
            ;;
        *)
            echo ""
            ;;
    esac
}

# Build summary context (for large repos)
build_summary_context() {
    local dir="$1"
    local max_files="${2:-20}"

    cat <<EOF
# Repository Structure Summary

## Directory Structure
$(find "$dir" -type d -not -path "*/\.*" -not -path "*/node_modules/*" -not -path "*/vendor/*" 2>/dev/null | head -30)

## Key Files Found
EOF

    # Find and list key security-relevant files
    local key_patterns=(
        "*auth*"
        "*login*"
        "*password*"
        "*secret*"
        "*config*"
        "*routes*"
        "*api*"
        "*controller*"
        "*handler*"
    )

    for pattern in "${key_patterns[@]}"; do
        local files
        files=$(find "$dir" -type f -iname "$pattern" -not -path "*/node_modules/*" -not -path "*/vendor/*" 2>/dev/null | head -5)
        if [[ -n "$files" ]]; then
            echo ""
            echo "### Pattern: $pattern"
            echo "$files"
        fi
    done
}

# Extract security-relevant code snippets
extract_security_snippets() {
    local file="$1"
    local language
    language=$(get_file_language "$file" 2>/dev/null || echo "unknown")

    # Patterns that often indicate security-relevant code
    local patterns=(
        "password"
        "secret"
        "key"
        "token"
        "auth"
        "query"
        "execute"
        "eval"
        "exec"
        "system"
        "shell"
        "command"
        "sql"
        "html"
        "cookie"
        "session"
        "encrypt"
        "decrypt"
        "hash"
        "verify"
        "validate"
    )

    echo "## Security-Relevant Snippets from: $file"
    echo ""

    for pattern in "${patterns[@]}"; do
        local matches
        matches=$(grep -n -i "$pattern" "$file" 2>/dev/null | head -5)
        if [[ -n "$matches" ]]; then
            echo "### Pattern: $pattern"
            echo "\`\`\`"
            echo "$matches"
            echo "\`\`\`"
            echo ""
        fi
    done
}

# Build context with related files
# Given a file, find and include related files (imports, tests, configs)
build_related_context() {
    local file="$1"
    local dir
    dir=$(dirname "$file")
    local basename
    basename=$(basename "$file")
    local name="${basename%.*}"

    echo "## Primary File"
    build_file_context "$file"
    echo ""

    # Look for related test file
    local test_patterns=(
        "${dir}/test_${name}.py"
        "${dir}/${name}_test.py"
        "${dir}/__tests__/${name}.test.js"
        "${dir}/${name}.test.js"
        "${dir}/${name}.spec.js"
        "${dir}/../tests/${name}_test.go"
        "${dir}/${name}Test.java"
    )

    for test_file in "${test_patterns[@]}"; do
        if [[ -f "$test_file" ]]; then
            echo "## Related Test File"
            build_file_context "$test_file"
            echo ""
            break
        fi
    done

    # Look for related config in same directory
    for config in "$dir"/*.json "$dir"/*.yaml "$dir"/*.yml; do
        if [[ -f "$config" ]]; then
            local config_size
            config_size=$(wc -c < "$config" 2>/dev/null || echo "0")
            if [[ "$config_size" -lt 10000 ]]; then
                echo "## Related Config"
                build_file_context "$config"
                echo ""
            fi
            break
        fi
    done
}
