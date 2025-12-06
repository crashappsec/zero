#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Agent Loader Library
# Functions for loading agent definitions and context for Claude Code
#
# Usage:
#   source utils/zero/lib/agent-loader.sh
#   load_agent_context "cereal" "expressjs/express"
#
# Note: Compatible with Bash 3.x (macOS default)
#############################################################################

# Get script directory for relative paths
AGENT_LOADER_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ZERO_DIR="$(dirname "$AGENT_LOADER_DIR")"
UTILS_DIR="$(dirname "$ZERO_DIR")"
REPO_ROOT="$(dirname "$UTILS_DIR")"
AGENTS_DIR="$REPO_ROOT/agents"

# Load zero-lib if not already loaded
if ! type gibson_project_path &>/dev/null; then
    source "$ZERO_DIR/lib/zero-lib.sh"
fi

#############################################################################
# Agent Registry Functions (Bash 3.x compatible - no associative arrays)
#############################################################################

# Get the directory for an agent
# Usage: agent_get_dir "cereal"
agent_get_dir() {
    local agent_name="$1"
    local dir=""

    case "$agent_name" in
        # Primary agents (Hackers movie inspired names - directories match agent names)
        cereal)   dir="cereal" ;;    # Cereal Killer - paranoid, watches for malware in deps
        razor)    dir="razor" ;;     # Razor - cuts through code to find vulnerabilities
        blade)    dir="blade" ;;     # Blade - meticulous, detail-oriented for auditing
        phreak)   dir="phreak" ;;    # Phantom Phreak - knows the legal angles
        acid)     dir="acid" ;;      # Acid Burn - sharp, stylish frontend expert
        dade)     dir="dade" ;;      # Dade Murphy - backend systems expert
        nikon)    dir="nikon" ;;     # Lord Nikon - photographic memory, sees big picture
        joey)     dir="joey" ;;      # Joey - builds things, sometimes breaks them
        plague)   dir="plague" ;;    # The Plague - controls infrastructure (reformed)
        gibson)   dir="gibson" ;;    # The Gibson - tracks everything
        *)        dir="" ;;
    esac

    if [[ -n "$dir" ]]; then
        echo "$AGENTS_DIR/$dir"
    else
        echo ""
        return 1
    fi
}

# Get data requirements for an agent
# Usage: agent_get_required_data "cereal"
agent_get_required_data() {
    local agent_name="$1"

    case "$agent_name" in
        cereal)       echo "vulnerabilities package-health dependencies package-malcontent licenses package-sbom" ;;
        cereal-basic) echo "vulnerabilities package-health dependencies licenses package-sbom" ;;
        razor)        echo "code-security code-secrets technology secrets-scanner" ;;
        blade)        echo "vulnerabilities licenses package-sbom iac-security code-security" ;;
        phreak)       echo "licenses dependencies package-sbom" ;;
        acid)         echo "technology code-security" ;;
        dade)         echo "technology code-security" ;;
        nikon)        echo "technology dependencies package-sbom" ;;
        joey)         echo "technology dora code-security" ;;
        plague)       echo "technology dora iac-security" ;;
        gibson)       echo "dora code-ownership git-insights" ;;
        *)            echo "" ;;
    esac
}

# Get allowed tools for an agent
# Usage: agent_get_tools "cereal"
agent_get_tools() {
    local agent_name="$1"

    case "$agent_name" in
        cereal)       echo "Read Grep Glob WebSearch WebFetch" ;;  # Full investigation capability
        cereal-basic) echo "Read Grep Glob" ;;
        razor)        echo "Read Grep Glob WebSearch" ;;           # Security research
        blade)        echo "Read Grep Glob WebFetch" ;;            # Compliance docs
        phreak)       echo "Read Grep WebFetch" ;;                 # Legal research
        acid)         echo "Read Grep Glob" ;;                     # Frontend code review
        dade)         echo "Read Grep Glob" ;;                     # Backend code review
        nikon)        echo "Read Grep Glob" ;;                     # Architecture review
        joey)         echo "Read Grep Glob Bash" ;;                # Build/CI testing
        plague)       echo "Read Grep Glob Bash" ;;                # Infrastructure commands
        gibson)       echo "Read Grep Glob" ;;                     # Metrics analysis
        *)            echo "Read" ;;
    esac
}

#############################################################################
# Core Functions
#############################################################################

# Check if an agent exists
# Usage: agent_exists "cereal"
agent_exists() {
    local agent_name="$1"
    local agent_dir=$(agent_get_dir "$agent_name")

    [[ -n "$agent_dir" ]] && [[ -d "$agent_dir" ]] && [[ -f "$agent_dir/agent.md" ]]
}

# List all available agents
# Usage: agent_list
agent_list() {
    local agents="cereal razor blade phreak acid dade nikon joey plague gibson"
    for agent in $agents; do
        if agent_exists "$agent"; then
            echo "$agent"
        fi
    done
}

# Get agent definition content
# Usage: agent_get_definition "cereal"
agent_get_definition() {
    local agent_name="$1"
    local agent_dir=$(agent_get_dir "$agent_name")

    if [[ -z "$agent_dir" ]] || [[ ! -f "$agent_dir/agent.md" ]]; then
        echo ""
        return 1
    fi

    cat "$agent_dir/agent.md"
}

# Get paths to agent knowledge files
# Usage: agent_get_knowledge_paths "cereal"
agent_get_knowledge_paths() {
    local agent_name="$1"
    local agent_dir=$(agent_get_dir "$agent_name")

    if [[ -z "$agent_dir" ]]; then
        return 1
    fi

    # Find all knowledge files
    find "$agent_dir/knowledge" -type f \( -name "*.md" -o -name "*.json" \) 2>/dev/null | sort
}

#############################################################################
# Context Loading Functions
#############################################################################

# Load scanner data for a project
# Usage: load_scanner_data_for_agent "cereal" "expressjs/express"
load_scanner_data_for_agent() {
    local agent_name="$1"
    local project_id="$2"
    local required_data=$(agent_get_required_data "$agent_name")

    local project_path=$(gibson_project_path "$project_id")
    local analysis_path="$project_path/analysis"
    local scanners_path="$analysis_path/scanners"

    local result='{}'

    for scanner in $required_data; do
        local data_file=""

        # Check scanners directory first (new structure)
        if [[ -d "$scanners_path/$scanner" ]]; then
            # Find the most recent JSON file in scanner directory
            data_file=$(find "$scanners_path/$scanner" -name "*.json" -type f 2>/dev/null | head -1)
        fi

        # Fall back to analysis root (old structure)
        if [[ -z "$data_file" ]] && [[ -f "$analysis_path/${scanner}.json" ]]; then
            data_file="$analysis_path/${scanner}.json"
        fi

        # Also check without hyphens
        local scanner_alt="${scanner//-/_}"
        if [[ -z "$data_file" ]] && [[ -f "$analysis_path/${scanner_alt}.json" ]]; then
            data_file="$analysis_path/${scanner_alt}.json"
        fi

        if [[ -n "$data_file" ]] && [[ -f "$data_file" ]]; then
            # Add scanner data to result
            local scanner_data=$(cat "$data_file" 2>/dev/null)
            if [[ -n "$scanner_data" ]] && echo "$scanner_data" | jq . &>/dev/null; then
                result=$(echo "$result" | jq --arg key "$scanner" --argjson data "$scanner_data" '. + {($key): $data}' 2>/dev/null || echo "$result")
            fi
        fi
    done

    echo "$result"
}

# Load full agent context for Claude Code
# Usage: load_agent_context "cereal" "expressjs/express"
# Returns JSON with agent definition, knowledge paths, and scanner data
load_agent_context() {
    local agent_name="$1"
    local project_id="$2"

    if ! agent_exists "$agent_name"; then
        echo '{"error": "Agent not found: '"$agent_name"'"}'
        return 1
    fi

    local agent_dir=$(agent_get_dir "$agent_name")
    local definition=$(agent_get_definition "$agent_name")
    local tools=$(agent_get_tools "$agent_name")
    local knowledge_paths=$(agent_get_knowledge_paths "$agent_name" | jq -R . 2>/dev/null | jq -s . 2>/dev/null || echo '[]')
    local scanner_data=$(load_scanner_data_for_agent "$agent_name" "$project_id")

    # Get project info
    local project_path=$(gibson_project_path "$project_id")
    local repo_path="$project_path/repo"
    local manifest_path="$project_path/analysis/manifest.json"

    local project_info='{}'
    if [[ -f "$manifest_path" ]]; then
        project_info=$(cat "$manifest_path" 2>/dev/null || echo '{}')
    fi

    # Build context JSON
    jq -n \
        --arg agent "$agent_name" \
        --arg agent_dir "$agent_dir" \
        --arg definition "$definition" \
        --arg tools "$tools" \
        --argjson knowledge_paths "$knowledge_paths" \
        --argjson scanner_data "$scanner_data" \
        --arg project_id "$project_id" \
        --arg repo_path "$repo_path" \
        --argjson project_info "$project_info" \
        '{
            agent: {
                name: $agent,
                directory: $agent_dir,
                definition: $definition,
                tools_allowed: ($tools | split(" ")),
                knowledge_paths: $knowledge_paths
            },
            project: {
                id: $project_id,
                repo_path: $repo_path,
                info: $project_info
            },
            scanner_data: $scanner_data
        }' 2>/dev/null || echo '{"error": "Failed to build context"}'
}

# Get a summary of findings for quick Q&A mode
# Usage: get_findings_summary "cereal" "expressjs/express"
get_findings_summary() {
    local agent_name="$1"
    local project_id="$2"

    local scanner_data=$(load_scanner_data_for_agent "$agent_name" "$project_id")

    # Build summary based on agent type
    case "$agent_name" in
        cereal)
            echo "$scanner_data" | jq '{
                vulnerabilities: (.vulnerabilities.summary // {}),
                malcontent: (."package-malcontent".summary // {}),
                package_health: (."package-health".summary // {}),
                dependencies: (.dependencies.summary // {}),
                licenses: (.licenses.summary // {})
            }' 2>/dev/null || echo '{}'
            ;;
        razor)
            echo "$scanner_data" | jq '{
                security_findings: (."code-security".summary // {}),
                secrets: (."secrets-scanner".summary // {}),
                technology: (.technology.summary // {})
            }' 2>/dev/null || echo '{}'
            ;;
        *)
            # Generic summary - return all summaries
            echo "$scanner_data" | jq 'to_entries | map({key: .key, value: .value.summary}) | from_entries' 2>/dev/null || echo '{}'
            ;;
    esac
}

# Check if investigation mode should be triggered
# Based on keywords in the query
# Usage: should_investigate "Investigate the crypto behavior in lodash"
should_investigate() {
    local query="$1"
    local query_lower=$(echo "$query" | tr '[:upper:]' '[:lower:]')

    # Investigation trigger keywords
    local triggers="investigate trace analyze examine inspect deep-dive research explore why how"

    for trigger in $triggers; do
        if [[ "$query_lower" == *"$trigger"* ]]; then
            return 0  # Should investigate
        fi
    done

    return 1  # Simple Q&A mode
}

#############################################################################
# Export functions
#############################################################################

export -f agent_get_dir
export -f agent_exists
export -f agent_list
export -f agent_get_definition
export -f agent_get_knowledge_paths
export -f agent_get_required_data
export -f agent_get_tools
export -f load_scanner_data_for_agent
export -f load_agent_context
export -f get_findings_summary
export -f should_investigate
