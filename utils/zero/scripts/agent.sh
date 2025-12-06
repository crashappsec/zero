#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Phantom Agent Chat
# Interactive agent selection and context preparation for Claude Code
#
# Usage: ./agent.sh [agent_name] [project_id]
#        ./agent.sh --interactive
#
# Note: Compatible with Bash 3.x (macOS default)
#############################################################################

set -e

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ZERO_DIR="$(dirname "$SCRIPT_DIR")"
UTILS_DIR="$(dirname "$ZERO_DIR")"
REPO_ROOT="$(dirname "$UTILS_DIR")"

# Load Phantom library
source "$ZERO_DIR/lib/zero-lib.sh"

# Load Agent loader
source "$ZERO_DIR/lib/agent-loader.sh"

#############################################################################
# Agent Registry (Bash 3.x compatible - no associative arrays)
#############################################################################

# Get agent description
agent_get_description() {
    local agent_name="$1"
    case "$agent_name" in
        zero)     echo "Master orchestrator - coordinates all agents (Zero Cool)" ;;
        cereal)   echo "Supply chain security - paranoid about dependencies (Cereal Killer)" ;;
        razor)    echo "Code security - cuts through vulnerabilities" ;;
        blade)    echo "Compliance - meticulous auditor" ;;
        phreak)   echo "Legal - licenses, knows the angles (Phantom Phreak)" ;;
        acid)     echo "Frontend - stylish code quality (Acid Burn)" ;;
        dade)     echo "Backend - calm, methodical systems (Crash Override)" ;;
        nikon)    echo "Architecture - photographic memory for patterns (Lord Nikon)" ;;
        joey)     echo "Build - eager to prove himself" ;;
        plague)   echo "DevOps - reformed villain, knows the threats (The Plague)" ;;
        gibson)   echo "Engineering metrics - the supercomputer sees all" ;;
        # Legacy aliases
        scout)    echo "Supply chain security (alias for cereal)" ;;
        sentinel) echo "Code security (alias for razor)" ;;
        quinn)    echo "Compliance (alias for blade)" ;;
        harper)   echo "Legal (alias for phreak)" ;;
        *)        echo "Unknown agent" ;;
    esac
}

# Get agent persona name
agent_get_persona() {
    local agent_name="$1"
    case "$agent_name" in
        zero)     echo "Zero Cool" ;;
        cereal)   echo "Cereal Killer" ;;
        razor)    echo "Razor" ;;
        blade)    echo "Blade" ;;
        phreak)   echo "Phantom Phreak" ;;
        acid)     echo "Acid Burn" ;;
        dade)     echo "Crash Override" ;;
        nikon)    echo "Lord Nikon" ;;
        joey)     echo "Joey" ;;
        plague)   echo "The Plague" ;;
        gibson)   echo "The Gibson" ;;
        # Legacy aliases
        scout)    echo "Cereal Killer" ;;
        sentinel) echo "Razor" ;;
        quinn)    echo "Blade" ;;
        harper)   echo "Phantom Phreak" ;;
        *)        echo "$agent_name" ;;
    esac
}

#############################################################################
# Functions
#############################################################################

usage() {
    cat << EOF
Phantom Agent Chat - Interactive agent conversations

Usage: $0 [options] [agent] [project]

OPTIONS:
    -i, --interactive    Interactive mode (select agent and project)
    -l, --list           List available agents
    -c, --context        Generate context file for Claude Code
    -h, --help           Show this help

AGENTS:
EOF
    for agent in zero cereal razor blade phreak acid dade nikon joey plague gibson; do
        if agent_exists "$agent"; then
            local persona=$(agent_get_persona "$agent")
            local desc=$(agent_get_description "$agent")
            printf "    %-10s %-10s %s\n" "$agent" "($persona)" "$desc"
        fi
    done
    cat << EOF

EXAMPLES:
    $0                          # Interactive mode
    $0 scout                    # Chat with Scout (uses active project)
    $0 scout expressjs/express  # Chat with Scout about Express
    $0 --list                   # List all agents
    $0 --context scout express  # Generate context file

EOF
    exit 0
}

# List available agents
list_agents() {
    print_zero_banner
    echo -e "${BOLD}Available Agents${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo

    local i=1
    for agent in zero cereal razor blade phreak acid dade nikon joey plague gibson; do
        if agent_exists "$agent"; then
            local persona=$(agent_get_persona "$agent")
            local desc=$(agent_get_description "$agent")
            local tools=$(agent_get_tools "$agent" 2>/dev/null || echo "all")
            echo -e "  ${CYAN}$i${NC}  ${BOLD}$persona${NC} ($agent)"
            echo -e "      ${DIM}$desc${NC}"
            echo -e "      ${DIM}Tools: $tools${NC}"
            echo
            ((i++))
        fi
    done
}

# Select agent interactively
select_agent() {
    echo -e "${BOLD}Select an Agent${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo

    local agents=""
    local agent_count=0
    local i=1
    for agent in zero cereal razor blade phreak acid dade nikon joey plague gibson; do
        if agent_exists "$agent"; then
            agents="$agents $agent"
            agent_count=$((agent_count + 1))
            local persona=$(agent_get_persona "$agent")
            local desc=$(agent_get_description "$agent")
            echo -e "  ${CYAN}$i${NC}  ${BOLD}$persona${NC} - $desc"
            ((i++))
        fi
    done
    echo
    read -p "Choose agent [1-$agent_count]: " choice

    if [[ "$choice" =~ ^[0-9]+$ ]] && [[ "$choice" -ge 1 ]] && [[ "$choice" -le $agent_count ]]; then
        # Get the nth agent from the list
        echo "$agents" | tr ' ' '\n' | sed -n "${choice}p"
    else
        echo ""
    fi
}

# Select project interactively
select_project() {
    echo
    echo -e "${BOLD}Select a Project${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo

    if [[ ! -d "$GIBSON_PROJECTS_DIR" ]]; then
        echo -e "${YELLOW}No projects hydrated yet.${NC}"
        echo "Run: ./zero.sh hydrate <owner/repo>"
        return 1
    fi

    local projects=""
    local project_count=0
    local i=1

    for org_dir in "$GIBSON_PROJECTS_DIR"/*/; do
        [[ ! -d "$org_dir" ]] && continue
        local org=$(basename "$org_dir")

        for repo_dir in "$org_dir"*/; do
            [[ ! -d "$repo_dir" ]] && continue
            local repo=$(basename "$repo_dir")
            local project_id="${org}/${repo}"
            projects="$projects $project_id"
            project_count=$((project_count + 1))

            # Get scan info
            local manifest="$repo_dir/analysis/manifest.json"
            local mode="unknown"
            if [[ -f "$manifest" ]]; then
                mode=$(jq -r '.mode // "standard"' "$manifest" 2>/dev/null)
            fi

            echo -e "  ${CYAN}$i${NC}  ${BOLD}$project_id${NC} ${DIM}[$mode]${NC}"
            ((i++))
        done
    done

    if [[ $project_count -eq 0 ]]; then
        echo -e "${YELLOW}No projects hydrated yet.${NC}"
        return 1
    fi

    echo
    read -p "Choose project [1-$project_count]: " choice

    if [[ "$choice" =~ ^[0-9]+$ ]] && [[ "$choice" -ge 1 ]] && [[ "$choice" -le $project_count ]]; then
        # Get the nth project from the list
        echo "$projects" | tr ' ' '\n' | sed -n "${choice}p"
    else
        echo ""
    fi
}

# Generate context for Claude Code
generate_context() {
    local agent_name="$1"
    local project_id="$2"

    if ! agent_exists "$agent_name"; then
        echo -e "${RED}Error: Unknown agent '$agent_name'${NC}" >&2
        return 1
    fi

    local project_path=$(gibson_project_path "$project_id")
    if [[ ! -d "$project_path/analysis" ]]; then
        echo -e "${RED}Error: Project '$project_id' not hydrated${NC}" >&2
        return 1
    fi

    # Load full context
    load_agent_context "$agent_name" "$project_id"
}

# Generate a prompt file for Claude Code
generate_prompt_file() {
    local agent_name="$1"
    local project_id="$2"
    local output_file="$3"

    local persona=$(agent_get_persona "$agent_name")
    local agent_dir=$(agent_get_dir "$agent_name")
    local project_path=$(gibson_project_path "$project_id")
    local repo_path="$project_path/repo"

    # Get agent definition
    local definition=$(agent_get_definition "$agent_name")

    # Get findings summary
    local summary=$(get_findings_summary "$agent_name" "$project_id")

    # Build the prompt
    cat > "$output_file" << EOF
# Chat with $persona

You are now chatting with **$persona**, a specialist agent.

## Agent Definition

$definition

## Project Context

**Project:** $project_id
**Repository:** $repo_path

## Current Findings Summary

\`\`\`json
$summary
\`\`\`

## Instructions

You are $persona. Respond to the user's questions using your expertise and the analysis data above.

- For simple questions, use the cached data to respond
- For investigation requests (investigate, trace, analyze, examine), use your tools to dig deeper:
  - **Read** files in the repository
  - **Grep** for patterns
  - **WebSearch** for CVE/advisory research
- Always cite specific file:line references when discussing findings
- Maintain your persona throughout the conversation

## Begin Conversation

The user wants to chat with you about $project_id. Wait for their question.
EOF

    echo "$output_file"
}

# Launch Claude with agent persona
launch_claude_chat() {
    local agent_name="$1"
    local project_id="${2:-}"

    local persona=$(agent_get_persona "$agent_name")

    # Build system prompt from agent definition
    local agent_md="$REPO_ROOT/agents/$agent_name/agent.md"
    if [[ ! -f "$agent_md" ]]; then
        echo -e "${RED}Error: Agent definition not found: $agent_md${NC}" >&2
        exit 1
    fi

    # Write system prompt to temp file to handle special characters
    local prompt_file=$(mktemp)
    cat "$agent_md" > "$prompt_file"

    # Add project context if available
    if [[ -n "$project_id" ]]; then
        local project_path=$(gibson_project_path "$project_id" 2>/dev/null || echo "")
        if [[ -n "$project_path" ]] && [[ -d "$project_path/analysis" ]]; then
            local summary=$(get_findings_summary "$agent_name" "$project_id" 2>/dev/null || echo "{}")
            cat >> "$prompt_file" << EOF

## Current Project: $project_id

Analysis data is available at: $project_path/analysis/

### Findings Summary
\`\`\`json
$summary
\`\`\`
EOF
        fi
    fi

    # Check if claude CLI is available
    if ! command -v claude &>/dev/null; then
        echo -e "${RED}Error: 'claude' CLI not found${NC}" >&2
        echo -e "Install Claude Code: ${CYAN}npm install -g @anthropic-ai/claude-code${NC}"
        rm -f "$prompt_file"
        exit 1
    fi

    # Check if we're in an interactive terminal
    if [[ ! -t 0 ]] || [[ ! -t 1 ]]; then
        echo -e "${RED}Error: Agent chat requires an interactive terminal${NC}" >&2
        echo -e "Run this command directly in your terminal, not from a script or pipe."
        rm -f "$prompt_file"
        exit 1
    fi

    echo -e "${GREEN}✓${NC} Launching chat with ${BOLD}$persona${NC}..."
    [[ -n "$project_id" ]] && echo -e "  Project: ${CYAN}$project_id${NC}"
    echo

    # Launch claude with the system prompt
    # Use --append-system-prompt to add persona to Claude's default behavior
    # Keep the temp file for passing to claude to avoid shell escaping issues
    claude --append-system-prompt "$(cat "$prompt_file")"
    local exit_code=$?

    rm -f "$prompt_file"
    exit $exit_code
}

# Interactive chat mode
run_interactive() {
    print_zero_banner
    echo -e "${BOLD}Agent Chat${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo
    echo "Start a conversation with a specialist agent about your project."
    echo

    # Select agent
    local agent=$(select_agent)
    if [[ -z "$agent" ]]; then
        echo -e "${RED}Invalid selection${NC}"
        exit 1
    fi

    # Select project (optional)
    local project=""
    if [[ -d "$GIBSON_PROJECTS_DIR" ]]; then
        echo
        read -p "Load project context? [y/N]: " load_project
        if [[ "$load_project" =~ ^[Yy] ]]; then
            project=$(select_project)
        fi
    fi

    # Launch Claude with agent
    launch_claude_chat "$agent" "$project"
}

#############################################################################
# Main
#############################################################################

main() {
    local agent=""
    local project=""
    local mode="interactive"

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            -i|--interactive)
                mode="interactive"
                shift
                ;;
            -l|--list)
                mode="list"
                shift
                ;;
            -c|--context)
                mode="context"
                shift
                ;;
            -h|--help)
                usage
                ;;
            -*)
                echo -e "${RED}Unknown option: $1${NC}" >&2
                exit 1
                ;;
            *)
                if [[ -z "$agent" ]]; then
                    agent="$1"
                elif [[ -z "$project" ]]; then
                    project="$1"
                fi
                shift
                ;;
        esac
    done

    case "$mode" in
        list)
            list_agents
            ;;
        context)
            if [[ -z "$agent" ]] || [[ -z "$project" ]]; then
                echo -e "${RED}Error: --context requires agent and project${NC}" >&2
                echo "Usage: $0 --context <agent> <project>"
                exit 1
            fi
            generate_context "$agent" "$project"
            ;;
        interactive)
            if [[ -n "$agent" ]]; then
                # Agent specified - launch Claude directly
                launch_claude_chat "$agent" "$project"
            else
                run_interactive
            fi
            ;;
    esac
}

main "$@"
