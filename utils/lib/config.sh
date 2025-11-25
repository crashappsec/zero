#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Global Configuration Library
# Provides common paths and configuration for all analysers
#############################################################################

# Detect repository root (works from any analyser location)
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

# RAG (Retrieval-Augmented Generation) configuration
# Can be overridden via environment variable or .env file
export RAG_DIR="${RAG_DIR:-$REPO_ROOT/rag}"

# RAG Server configuration (for production deployments with vector databases)
export RAG_SERVER_URL="${RAG_SERVER_URL:-}"
export RAG_API_KEY="${RAG_API_KEY:-}"
export RAG_SERVER_TYPE="${RAG_SERVER_TYPE:-local}"
export RAG_COLLECTION="${RAG_COLLECTION:-gibson-powers-knowledge}"

# RAG (Retrieval-Augmented Generation) knowledge base paths
export RAG_SUPPLY_CHAIN_DIR="$RAG_DIR/supply-chain"
export RAG_DORA_DIR="$RAG_DIR/dora-metrics"
export RAG_CODE_OWNERSHIP_DIR="$RAG_DIR/code-ownership"
export RAG_COCOMO_DIR="$RAG_DIR/cocomo"
export RAG_CERTIFICATE_DIR="$RAG_DIR/certificate-analysis"
export RAG_TECHNOLOGY_DIR="$RAG_DIR/technology-identification"
export RAG_LEGAL_DIR="$RAG_DIR/legal-review"

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

#############################################################################
# RAG Server Functions (for future vector database support)
#############################################################################

# Check if RAG server is configured and reachable
# Returns 0 if server is available, 1 if should use local filesystem
is_rag_server_available() {
    # No server URL configured - use local
    if [[ -z "$RAG_SERVER_URL" ]]; then
        return 1
    fi

    # Try to reach the server (quick health check)
    if curl -s --connect-timeout 2 --max-time 5 "$RAG_SERVER_URL/health" >/dev/null 2>&1; then
        return 0
    fi

    # Server not reachable - fall back to local
    echo "Warning: RAG server not reachable, using local filesystem" >&2
    return 1
}

# Query RAG server for relevant content
# Usage: query_rag_server "certificate security best practices" "certificate-analysis"
query_rag_server() {
    local query="$1"
    local topic="${2:-}"
    local limit="${3:-5}"

    if [[ -z "$RAG_SERVER_URL" ]]; then
        return 1
    fi

    local endpoint="$RAG_SERVER_URL/query"
    local payload

    # Build query payload based on server type
    case "$RAG_SERVER_TYPE" in
        pinecone)
            payload=$(jq -n \
                --arg query "$query" \
                --arg namespace "$topic" \
                --argjson topK "$limit" \
                '{query: $query, namespace: $namespace, topK: $topK}')
            ;;
        weaviate|chromadb|qdrant)
            payload=$(jq -n \
                --arg query "$query" \
                --arg collection "$RAG_COLLECTION" \
                --argjson limit "$limit" \
                '{query: $query, collection: $collection, limit: $limit}')
            ;;
        *)
            # Generic format
            payload=$(jq -n \
                --arg query "$query" \
                --arg topic "$topic" \
                --argjson limit "$limit" \
                '{query: $query, topic: $topic, limit: $limit}')
            ;;
    esac

    # Query the server
    local response
    response=$(curl -s --connect-timeout 5 --max-time 30 \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $RAG_API_KEY" \
        -d "$payload" \
        "$endpoint")

    if [[ $? -eq 0 ]] && [[ -n "$response" ]]; then
        echo "$response" | jq -r '.results[].content // .documents[].content // empty' 2>/dev/null
        return 0
    fi

    return 1
}

# Get RAG content with server fallback to local filesystem
# Usage: get_rag_content_smart "certificate-analysis" "query about certificates"
get_rag_content_smart() {
    local topic="$1"
    local query="${2:-}"

    # Try RAG server first if configured
    if is_rag_server_available && [[ -n "$query" ]]; then
        local server_content
        server_content=$(query_rag_server "$query" "$topic")
        if [[ -n "$server_content" ]]; then
            echo "$server_content"
            return 0
        fi
    fi

    # Fall back to local filesystem
    get_all_rag_content "$topic"
}

#############################################################################
# RAG Initialization and Validation
#############################################################################

# Initialize RAG configuration
# Called by bootstrap.sh to set up RAG paths
init_rag_config() {
    local quiet="${1:-false}"

    # Validate RAG directory exists
    if [[ ! -d "$RAG_DIR" ]]; then
        if [[ "$quiet" != "true" ]]; then
            echo -e "${YELLOW}Warning: RAG directory not found: $RAG_DIR${NC}" >&2
        fi
        return 1
    fi

    # Count available RAG topics
    local topic_count=0
    local file_count=0

    for dir in "$RAG_DIR"/*/; do
        if [[ -d "$dir" ]]; then
            ((topic_count++))
            file_count=$((file_count + $(find "$dir" -type f -name "*.md" 2>/dev/null | wc -l)))
        fi
    done

    if [[ "$quiet" != "true" ]]; then
        echo -e "${GREEN}✓${NC} RAG initialized: $topic_count topics, $file_count documents"
    fi

    return 0
}

# Get RAG status summary
get_rag_status() {
    echo "RAG Configuration:"
    echo "  Directory: $RAG_DIR"
    echo "  Server URL: ${RAG_SERVER_URL:-not configured}"
    echo "  Server Type: $RAG_SERVER_TYPE"

    if [[ -d "$RAG_DIR" ]]; then
        echo "  Topics:"
        for dir in "$RAG_DIR"/*/; do
            if [[ -d "$dir" ]]; then
                local topic=$(basename "$dir")
                local count=$(find "$dir" -type f -name "*.md" 2>/dev/null | wc -l | tr -d ' ')
                echo "    - $topic: $count documents"
            fi
        done
    else
        echo "  Status: Directory not found"
    fi
}

# Export functions
export -f has_rag_content
export -f get_rag_content
export -f list_rag_files
export -f get_all_rag_content
export -f include_rag_in_prompt
export -f is_rag_server_available
export -f query_rag_server
export -f get_rag_content_smart
export -f init_rag_config
export -f get_rag_status
