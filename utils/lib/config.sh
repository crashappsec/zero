#!/bin/bash
# Copyright (c) 2024 Gibson Powers Contributors
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Global Configuration Library
# Provides common paths and configuration for all analyzers
#############################################################################

# Detect repository root (works from any analyzer location)
if [[ -n "${REPO_ROOT:-}" ]]; then
    # Already set by caller
    :
elif [[ -n "${BASH_SOURCE[0]}" ]]; then
    # Calculate from this file's location (utils/lib/config.sh)
    REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
else
    # Fallback: assume current directory is repo root
    REPO_ROOT="$(pwd)"
fi

# Global paths
export REPO_ROOT
export UTILS_DIR="$REPO_ROOT/utils"
export LIB_DIR="$REPO_ROOT/utils/lib"
export RAG_DIR="$REPO_ROOT/rag"

# RAG (Retrieval-Augmented Generation) knowledge base paths
export RAG_SUPPLY_CHAIN_DIR="$RAG_DIR/supply-chain"
export RAG_DORA_DIR="$RAG_DIR/dora-metrics"
export RAG_CODE_OWNERSHIP_DIR="$RAG_DIR/code-ownership"
export RAG_COCOMO_DIR="$RAG_DIR/cocomo"

# Common configuration defaults
export DEFAULT_OUTPUT_FORMAT="markdown"
export DEFAULT_DAYS=90

# Color codes for consistent output
export RED='\033[0;31m'
export GREEN='\033[0;32m'
export YELLOW='\033[1;33m'
export BLUE='\033[0;34m'
export CYAN='\033[0;36m'
export NC='\033[0m' # No Color

# Check if RAG directory exists
has_rag_content() {
    [[ -d "$RAG_DIR" ]]
}

# Get RAG content for a specific topic
# Usage: get_rag_content "code-ownership" "codeowners-best-practices.md"
get_rag_content() {
    local topic="$1"
    local file="$2"
    local rag_file="$RAG_DIR/$topic/$file"

    if [[ -f "$rag_file" ]]; then
        cat "$rag_file"
        return 0
    else
        return 1
    fi
}

# List available RAG files for a topic
# Usage: list_rag_files "code-ownership"
list_rag_files() {
    local topic="$1"
    local topic_dir="$RAG_DIR/$topic"

    if [[ -d "$topic_dir" ]]; then
        find "$topic_dir" -type f -name "*.md" -exec basename {} \;
    fi
}

# Get all RAG content for a topic (concatenated)
# Usage: get_all_rag_content "code-ownership"
get_all_rag_content() {
    local topic="$1"
    local topic_dir="$RAG_DIR/$topic"

    if [[ ! -d "$topic_dir" ]]; then
        return 1
    fi

    echo "# Reference Documentation for $topic"
    echo ""

    while IFS= read -r file; do
        if [[ -f "$file" ]]; then
            echo "## $(basename "$file" .md)"
            echo ""
            cat "$file"
            echo ""
            echo "---"
            echo ""
        fi
    done < <(find "$topic_dir" -type f -name "*.md" | sort)
}

# Include RAG context in Claude prompts
# Usage: include_rag_in_prompt "code-ownership" "$existing_prompt"
include_rag_in_prompt() {
    local topic="$1"
    local prompt="$2"
    local rag_content

    if rag_content=$(get_all_rag_content "$topic"); then
        cat << EOF
${prompt}

## Reference Documentation

The following reference documentation is provided for context and best practices:

${rag_content}

Please use this reference documentation to inform your analysis and recommendations.
EOF
    else
        # No RAG content available, return original prompt
        echo "$prompt"
    fi
}

# Export functions
export -f has_rag_content
export -f get_rag_content
export -f list_rag_files
export -f get_all_rag_content
export -f include_rag_in_prompt
