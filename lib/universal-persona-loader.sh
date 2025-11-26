#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Universal Persona Loader
# Loads persona definitions and reasoning framework for any scanner
# Implements chain-of-reasoning approach for consistent persona outputs
#############################################################################

# Get repository root
PERSONA_LOADER_SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PERSONA_LOADER_REPO_ROOT="$(dirname "$PERSONA_LOADER_SCRIPT_DIR")"

# Persona directories - stored in RAG for content management
UNIVERSAL_PERSONAS_DIR="${PERSONA_LOADER_REPO_ROOT}/rag/personas"
PERSONA_DEFINITIONS_DIR="${UNIVERSAL_PERSONAS_DIR}/definitions"
PERSONA_REASONING_DIR="${UNIVERSAL_PERSONAS_DIR}/reasoning"
PERSONA_OUTPUT_FORMATS_DIR="${UNIVERSAL_PERSONAS_DIR}/output-formats"

# Valid personas
UNIVERSAL_PERSONAS=("security-engineer" "software-engineer" "engineering-leader" "auditor")

# Check if persona is valid
is_valid_universal_persona() {
    local persona="$1"
    [[ " ${UNIVERSAL_PERSONAS[*]} " =~ " ${persona} " ]] || [[ "$persona" == "all" ]]
}

# Get persona display name
get_universal_persona_display_name() {
    local persona="$1"
    case "$persona" in
        security-engineer) echo "Security Engineer" ;;
        software-engineer) echo "Software Engineer" ;;
        engineering-leader) echo "Engineering Leader" ;;
        auditor) echo "Auditor" ;;
        all) echo "All Personas" ;;
        *) echo "$persona" ;;
    esac
}

# Load persona definition (Phase 1: Understand Your Audience)
load_persona_definition() {
    local persona="$1"
    local definition_file="${PERSONA_DEFINITIONS_DIR}/${persona}.md"

    if [[ -f "$definition_file" ]]; then
        cat "$definition_file"
    else
        echo "# Default Persona"
        echo "Provide a clear, actionable analysis of the scan results."
    fi
}

# Load reasoning framework
load_reasoning_framework() {
    local framework_file="${PERSONA_REASONING_DIR}/analysis-framework.md"

    if [[ -f "$framework_file" ]]; then
        cat "$framework_file"
    fi
}

# Load output format template (if exists)
load_output_format() {
    local persona="$1"
    local format_file="${PERSONA_OUTPUT_FORMATS_DIR}/${persona}.md"

    if [[ -f "$format_file" ]]; then
        cat "$format_file"
    fi
}

# Build chain-of-reasoning prompt for a persona
# This is the main function that scanners should call
build_persona_prompt() {
    local persona="$1"
    local domain_knowledge="$2"  # RAG content from the specific scanner
    local scan_data="$3"         # Raw scan output to analyze
    local scanner_name="${4:-Scanner}"  # Name of the scanner for context

    local prompt=""

    # Phase 1: Persona Context (Understand Your Audience)
    prompt+="# PHASE 1: UNDERSTAND YOUR AUDIENCE\n\n"
    prompt+="You are generating a report for a specific audience. Before analyzing any data, understand who you are advising:\n\n"
    prompt+="$(load_persona_definition "$persona")\n\n"

    # Load reasoning framework
    prompt+="# CHAIN OF REASONING FRAMEWORK\n\n"
    prompt+="$(load_reasoning_framework)\n\n"

    # Phase 2: Domain Knowledge
    if [[ -n "$domain_knowledge" ]]; then
        prompt+="# PHASE 2: DOMAIN KNOWLEDGE\n\n"
        prompt+="Apply this domain-specific expertise through your audience's lens:\n\n"
        prompt+="$domain_knowledge\n\n"
    fi

    # Phase 3: Analysis Task
    prompt+="# PHASE 3: ANALYSIS TASK\n\n"
    prompt+="Now analyze the following ${scanner_name} output and generate a report that serves your audience's needs.\n\n"
    prompt+="Remember:\n"
    prompt+="- FILTER: Include what matters to this persona, exclude what doesn't\n"
    prompt+="- FORMAT: Use their preferred terminology and output format\n"
    prompt+="- FOCUS: Lead with what's most important to them\n"
    prompt+="- FRAME: Connect findings to their decision context\n\n"

    # Load output format if available
    local output_format=$(load_output_format "$persona")
    if [[ -n "$output_format" ]]; then
        prompt+="## Output Format Requirements\n\n"
        prompt+="$output_format\n\n"
    fi

    prompt+="## Scan Data to Analyze\n\n"
    prompt+="$scan_data"

    echo -e "$prompt"
}

# Get list of all personas (for --persona all option)
get_all_personas() {
    echo "${UNIVERSAL_PERSONAS[@]}"
}

# Check if persona resources exist
check_persona_resources() {
    local persona="$1"
    local missing=()

    if [[ ! -f "${PERSONA_DEFINITIONS_DIR}/${persona}.md" ]]; then
        missing+=("definition: ${PERSONA_DEFINITIONS_DIR}/${persona}.md")
    fi

    if [[ ${#missing[@]} -gt 0 ]]; then
        echo "Missing persona resources for ${persona}:"
        for item in "${missing[@]}"; do
            echo "  - $item"
        done
        return 1
    fi

    return 0
}

# List available personas with descriptions
list_universal_personas() {
    echo "security-engineer|Security Engineer|Technical vulnerability analysis with CVE details and remediation"
    echo "software-engineer|Software Engineer|Developer-focused: commands, versions, breaking changes"
    echo "engineering-leader|Engineering Leader|Executive dashboard with metrics and strategic recommendations"
    echo "auditor|Auditor|Compliance assessment with framework mappings and evidence"
}

# Print persona summary for help text
print_persona_help() {
    echo ""
    echo "Available Personas (Universal - work with any scanner)"
    echo "======================================================="
    echo ""
    echo "  security-engineer     Technical security analysis"
    echo "                        Focus: CVEs, CVSS, KEV, remediation specifics"
    echo "                        Output: Technical report with tables and commands"
    echo ""
    echo "  software-engineer     Developer-focused guidance"
    echo "                        Focus: Commands, versions, breaking changes, effort"
    echo "                        Output: Copy-paste commands, migration guides"
    echo ""
    echo "  engineering-leader    Executive strategic overview"
    echo "                        Focus: Metrics, trends, resource needs, ROI"
    echo "                        Output: Dashboard with aggregated data"
    echo ""
    echo "  auditor               Compliance assessment"
    echo "                        Focus: Controls, frameworks, evidence, findings"
    echo "                        Output: Formal audit report"
    echo ""
    echo "  all                   Generate reports for ALL personas"
    echo ""
}
