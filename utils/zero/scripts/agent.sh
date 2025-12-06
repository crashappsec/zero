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
PHANTOM_DIR="$(dirname "$SCRIPT_DIR")"
UTILS_DIR="$(dirname "$PHANTOM_DIR")"
REPO_ROOT="$(dirname "$UTILS_DIR")"

# Load Phantom library
source "$PHANTOM_DIR/lib/phantom-lib.sh"

# Load Agent loader
source "$PHANTOM_DIR/lib/agent-loader.sh"

#############################################################################
# Agent Registry (Bash 3.x compatible - no associative arrays)
#############################################################################

# Get agent description
agent_get_description() {
    local agent_name="$1"
    case "$agent_name" in
        scout)    echo "Supply chain security - vulnerabilities, malcontent, package health" ;;
        sentinel) echo "Code security - static analysis, secrets, SAST findings" ;;
        quinn)    echo "Compliance - SOC 2, ISO 27001, audit evidence" ;;
        harper)   echo "Legal - licenses, data privacy, contracts" ;;
        casey)    echo "Frontend - React, TypeScript, accessibility" ;;
        morgan)   echo "Backend - APIs, databases, data pipelines" ;;
        ada)      echo "Architecture - system design, patterns, trade-offs" ;;
        bailey)   echo "Build - CI/CD, performance, caching" ;;
        phoenix)  echo "DevOps - infrastructure, Kubernetes, incidents" ;;
        jordan)   echo "Engineering metrics - DORA, team health, KPIs" ;;
        *)        echo "Unknown agent" ;;
    esac
}

# Get agent persona name
agent_get_persona() {
    local agent_name="$1"
    case "$agent_name" in
        scout)    echo "Scout" ;;
        sentinel) echo "Sentinel" ;;
        quinn)    echo "Quinn" ;;
        harper)   echo "Harper" ;;
        casey)    echo "Casey" ;;
        morgan)   echo "Morgan" ;;
        ada)      echo "Ada" ;;
        bailey)   echo "Bailey" ;;
        phoenix)  echo "Phoenix" ;;
        jordan)   echo "Jordan" ;;
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
    for agent in scout sentinel quinn harper casey morgan ada bailey phoenix jordan; do
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
    print_phantom_banner
    echo -e "${BOLD}Available Agents${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo

    local i=1
    for agent in scout sentinel quinn harper casey morgan ada bailey phoenix jordan; do
        if agent_exists "$agent"; then
            local persona=$(agent_get_persona "$agent")
            local desc=$(agent_get_description "$agent")
            local tools=$(agent_get_tools "$agent")
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
    for agent in scout sentinel quinn harper casey morgan ada bailey phoenix jordan; do
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
        echo "Run: ./phantom.sh hydrate <owner/repo>"
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

# Interactive chat mode
run_interactive() {
    print_phantom_banner
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

    # Select project
    local project=$(select_project)
    if [[ -z "$project" ]]; then
        exit 1
    fi

    # Generate prompt file
    local persona=$(agent_get_persona "$agent")
    local prompt_file="/tmp/phantom-agent-${agent}-$(date +%s).md"
    generate_prompt_file "$agent" "$project" "$prompt_file"

    echo
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${GREEN}✓${NC} Agent context prepared"
    echo
    echo -e "  Agent:   ${BOLD}$persona${NC} ($agent)"
    echo -e "  Project: ${BOLD}$project${NC}"
    echo -e "  Prompt:  ${CYAN}$prompt_file${NC}"
    echo
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo
    echo -e "${BOLD}To start the conversation:${NC}"
    echo
    echo -e "  1. Copy this command:"
    echo -e "     ${CYAN}cat $prompt_file${NC}"
    echo
    echo -e "  2. Or use the /phantom slash command:"
    echo -e "     ${CYAN}/phantom ask $agent \"Your question here\"${NC}"
    echo
    echo -e "  3. For investigation mode (uses tools):"
    echo -e "     ${CYAN}/phantom ask $agent \"Investigate the malcontent findings\"${NC}"
    echo

    # Output the prompt file path for scripting
    echo "$prompt_file"
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
                # Agent specified, check if project specified
                if [[ -z "$project" ]]; then
                    # Try to get active project or prompt
                    print_phantom_banner
                    echo -e "${BOLD}Agent Chat${NC}"
                    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
                    echo
                    project=$(select_project)
                    if [[ -z "$project" ]]; then
                        exit 1
                    fi
                fi

                # Generate prompt for specified agent/project
                local persona=$(agent_get_persona "$agent")
                local prompt_file="/tmp/phantom-agent-${agent}-$(date +%s).md"
                generate_prompt_file "$agent" "$project" "$prompt_file"

                echo
                echo -e "${GREEN}✓${NC} Agent context prepared for ${BOLD}$persona${NC}"
                echo -e "  Project: $project"
                echo -e "  Prompt:  ${CYAN}$prompt_file${NC}"
                echo
                echo -e "Start with: ${CYAN}/phantom ask $agent \"Your question\"${NC}"
            else
                run_interactive
            fi
            ;;
    esac
}

main "$@"
