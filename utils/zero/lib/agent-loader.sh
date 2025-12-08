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
PERSONAS_DIR="$REPO_ROOT/rag/personas"

# Load zero-lib if not already loaded
if ! type gibson_project_path &>/dev/null; then
    source "$ZERO_DIR/lib/zero-lib.sh"
fi

#############################################################################
# Agent Registry Functions (Bash 3.x compatible - no associative arrays)
#############################################################################

# Get the directory for an agent
# Usage: agent_get_dir "cereal"
# Accepts both character names (cereal) and functional names (supply-chain)
agent_get_dir() {
    local agent_name="$1"
    local dir=""

    case "$agent_name" in
        # Functional directory names (primary)
        orchestrator|zero)           dir="orchestrator" ;;      # Zero Cool - master orchestrator
        supply-chain|cereal)         dir="supply-chain" ;;      # Cereal Killer - supply chain security
        code-security|razor)         dir="code-security" ;;     # Razor - code security/SAST
        compliance|blade)            dir="compliance" ;;        # Blade - compliance/auditor
        legal|phreak)                dir="legal" ;;             # Phantom Phreak - legal/licenses
        frontend|acid)               dir="frontend" ;;          # Acid Burn - frontend engineer
        backend|dade|flushot)        dir="backend" ;;           # Flu Shot - backend engineer
        architecture|nikon)          dir="architecture" ;;      # Lord Nikon - software architect
        build|joey)                  dir="build" ;;             # Joey - build/CI engineer
        devops|plague)               dir="devops" ;;            # The Plague - devops engineer
        engineering-leader|gibson)   dir="engineering-leader" ;;# The Gibson - engineering metrics
        *)                           dir="" ;;
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
        zero)         echo "all" ;;  # Zero can access all data to delegate
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
        zero)         echo "Read Grep Glob Bash WebSearch WebFetch Task" ;;  # Full orchestrator capability
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
# Persona Functions
#############################################################################

# List available personas
# Usage: persona_list
persona_list() {
    echo "security-engineer software-engineer engineering-leader auditor"
}

# Check if a persona exists
# Usage: persona_exists "security-engineer"
persona_exists() {
    local persona_name="$1"
    [[ -f "$PERSONAS_DIR/${persona_name}.md" ]]
}

# Get persona file path
# Usage: persona_get_path "security-engineer"
persona_get_path() {
    local persona_name="$1"
    local path="$PERSONAS_DIR/${persona_name}.md"

    if [[ -f "$path" ]]; then
        echo "$path"
    else
        echo ""
        return 1
    fi
}

# Get persona overlay path for a specific agent
# Usage: persona_get_overlay_path "cereal" "security-engineer"
persona_get_overlay_path() {
    local agent_name="$1"
    local persona_name="$2"
    local path="$PERSONAS_DIR/overlays/${agent_name}/${persona_name}-overlay.md"

    if [[ -f "$path" ]]; then
        echo "$path"
    else
        echo ""
        return 1
    fi
}

# Load persona definition content
# Usage: load_persona "security-engineer"
load_persona() {
    local persona_name="$1"
    local path=$(persona_get_path "$persona_name")

    if [[ -z "$path" ]]; then
        echo ""
        return 1
    fi

    cat "$path"
}

# Load persona overlay for an agent (if exists)
# Usage: load_persona_overlay "cereal" "security-engineer"
load_persona_overlay() {
    local agent_name="$1"
    local persona_name="$2"
    local path=$(persona_get_overlay_path "$agent_name" "$persona_name")

    if [[ -z "$path" ]]; then
        echo ""
        return 0  # Not an error - overlays are optional
    fi

    cat "$path"
}

# Build complete persona context (base + overlay)
# Usage: build_persona_context "cereal" "security-engineer"
build_persona_context() {
    local agent_name="$1"
    local persona_name="$2"

    local base=$(load_persona "$persona_name")
    local overlay=$(load_persona_overlay "$agent_name" "$persona_name")

    if [[ -z "$base" ]]; then
        echo '{"error": "Persona not found: '"$persona_name"'"}'
        return 1
    fi

    # Build JSON context
    jq -n \
        --arg persona "$persona_name" \
        --arg base "$base" \
        --arg overlay "$overlay" \
        --arg agent "$agent_name" \
        '{
            persona: $persona,
            agent: $agent,
            definition: $base,
            overlay: (if $overlay != "" then $overlay else null end),
            has_overlay: ($overlay != "")
        }' 2>/dev/null || echo '{"error": "Failed to build persona context"}'
}

# Load agent context WITH persona
# Usage: load_agent_context_with_persona "cereal" "expressjs/express" "security-engineer"
load_agent_context_with_persona() {
    local agent_name="$1"
    local project_id="$2"
    local persona_name="$3"

    # Get base agent context
    local agent_context=$(load_agent_context "$agent_name" "$project_id")

    if [[ -z "$persona_name" ]]; then
        echo "$agent_context"
        return 0
    fi

    # Get persona context
    local persona_context=$(build_persona_context "$agent_name" "$persona_name")

    # Merge contexts
    echo "$agent_context" | jq --argjson persona "$persona_context" '. + {persona: $persona}' 2>/dev/null || echo "$agent_context"
}

#############################################################################
# Voice Mode Functions
#############################################################################

# Get the current voice mode from config
# Usage: get_voice_mode
# Returns: "full", "minimal", or "neutral" (default: "full")
get_voice_mode() {
    local config_file="$ZERO_DIR/config/zero.config.json"
    if [[ -f "$config_file" ]]; then
        local mode=$(jq -r '.settings.voice_mode // "full"' "$config_file" 2>/dev/null)
        # Validate mode
        case "$mode" in
            full|minimal|neutral) echo "$mode" ;;
            *) echo "full" ;;
        esac
    else
        echo "full"
    fi
}

# Set the voice mode in config
# Usage: set_voice_mode "minimal"
set_voice_mode() {
    local mode="$1"
    local config_file="$ZERO_DIR/config/zero.config.json"

    # Validate mode
    case "$mode" in
        full|minimal|neutral) ;;
        *)
            echo "Error: Invalid voice mode '$mode'. Use: full, minimal, or neutral" >&2
            return 1
            ;;
    esac

    if [[ ! -f "$config_file" ]]; then
        echo "Error: Config file not found: $config_file" >&2
        return 1
    fi

    # Update config file
    local tmp=$(mktemp)
    if jq --arg mode "$mode" '.settings.voice_mode = $mode' "$config_file" > "$tmp" 2>/dev/null; then
        mv "$tmp" "$config_file"
        echo "Voice mode set to: $mode"
    else
        rm -f "$tmp"
        echo "Error: Failed to update config" >&2
        return 1
    fi
}

# Extract agent definition with voice mode applied
# Usage: get_agent_definition_with_voice "supply-chain" "minimal"
# The agent.md file should have voice sections marked with:
#   <!-- VOICE:full -->  ... content ...  <!-- /VOICE:full -->
#   <!-- VOICE:minimal --> ... content ... <!-- /VOICE:minimal -->
#   <!-- VOICE:neutral --> ... content ... <!-- /VOICE:neutral -->
get_agent_definition_with_voice() {
    local agent_name="$1"
    local voice_mode="${2:-$(get_voice_mode)}"
    local agent_dir=$(agent_get_dir "$agent_name")
    local agent_md="$agent_dir/agent.md"

    if [[ ! -f "$agent_md" ]]; then
        echo ""
        return 1
    fi

    local content=$(cat "$agent_md")

    # Check if file has voice markers
    if ! grep -q '<!-- VOICE:' "$agent_md" 2>/dev/null; then
        # No voice markers - return full content (legacy support)
        echo "$content"
        return 0
    fi

    # Extract base content (everything outside voice sections)
    # Then append the selected voice section
    local base_content=""
    local in_voice_section=false
    local current_voice=""
    local selected_voice_content=""

    while IFS= read -r line; do
        # Check for voice section start
        if [[ "$line" =~ \<!--\ VOICE:([a-z]+)\ --\> ]]; then
            in_voice_section=true
            current_voice="${BASH_REMATCH[1]}"
            continue
        fi

        # Check for voice section end
        if [[ "$line" =~ \<!--\ /VOICE:[a-z]+\ --\> ]]; then
            in_voice_section=false
            current_voice=""
            continue
        fi

        # Collect content
        if [[ "$in_voice_section" == true ]]; then
            if [[ "$current_voice" == "$voice_mode" ]]; then
                selected_voice_content+="$line"$'\n'
            fi
        else
            base_content+="$line"$'\n'
        fi
    done < "$agent_md"

    # Output base content + selected voice content
    echo "$base_content"
    echo "$selected_voice_content"
}

# Get voice mode description
# Usage: get_voice_mode_description "minimal"
get_voice_mode_description() {
    local mode="${1:-$(get_voice_mode)}"
    case "$mode" in
        full)    echo "Full Hackers character voice with quotes, catchphrases, and roleplay" ;;
        minimal) echo "Agent names retained, but no quotes, catchphrases, or heavy roleplay" ;;
        neutral) echo "Professional tone with no character references" ;;
        *)       echo "Unknown voice mode" ;;
    esac
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
export -f persona_list
export -f persona_exists
export -f persona_get_path
export -f persona_get_overlay_path
export -f load_persona
export -f load_persona_overlay
export -f build_persona_context
export -f load_agent_context_with_persona
export -f get_voice_mode
export -f set_voice_mode
export -f get_agent_definition_with_voice
export -f get_voice_mode_description
