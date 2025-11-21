#!/bin/bash
# Copyright (c) 2024 Crash Override Inc
# 101 Fulton St, 416, New York 10038
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Provenance Analyzer Script with Claude AI Integration
# Analyzes SLSA provenance and enhances with Claude analysis
# Usage: ./provenance-analyzer-claude.sh [options] <target>
#############################################################################

set -e

# Load environment variables from .env file if it exists
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PARENT_DIR="$(dirname "$SCRIPT_DIR")"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
CONFIG_FILE="$PARENT_DIR/config.json"

if [ -f "$REPO_ROOT/.env" ]; then
    source "$REPO_ROOT/.env"
fi

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Default options
VERIFY_SIGNATURES=false
MIN_SLSA_LEVEL=0
TEMP_DIR=""
CLEANUP=true
ANTHROPIC_API_KEY="${ANTHROPIC_API_KEY:-}"
MULTI_REPO_MODE=false
TARGETS_LIST=()

# Function to print usage
usage() {
    cat << EOF
Provenance Analyzer with Claude AI - Enhanced SLSA provenance analysis

Usage: $0 [OPTIONS] [target]

TARGET:
    SBOM file path          Analyze an existing SBOM (JSON/XML)
    Git repository URL      Clone and analyze repository
    Local directory path    Analyze local repository
    Package URL (purl)      Analyze specific package
    (If no target specified, uses config.json)

MULTI-REPO OPTIONS:
    --org ORG_NAME          Scan all repos in GitHub organization
    --repo OWNER/REPO       Scan specific repository
    --config FILE           Use alternate config file

ANALYSIS OPTIONS:
    --verify-signatures     Cryptographically verify signatures (requires cosign)
    --min-level LEVEL       Require minimum SLSA level (0-4)
    -k, --api-key KEY       Anthropic API key (or set ANTHROPIC_API_KEY env var)
    --keep-clone            Keep cloned repository (don't cleanup)
    -h, --help              Show this help message

ENVIRONMENT:
    ANTHROPIC_API_KEY       Your Anthropic API key

EXAMPLES:
    # Analyze SBOM with AI insights
    $0 /path/to/sbom.json

    # Analyze repository
    $0 https://github.com/org/repo

    # Scan organization
    $0 --org myorg --min-level 2

    # Analyze specific package
    $0 pkg:npm/express@4.17.1

EOF
    exit 1
}

# Function to check prerequisites
check_prerequisites() {
    # Check base analyzer
    if [[ ! -x "$SCRIPT_DIR/provenance-analyzer.sh" ]]; then
        echo -e "${RED}Error: provenance-analyzer.sh not found or not executable${NC}"
        exit 1
    fi

    # Check API key
    if [[ -z "$ANTHROPIC_API_KEY" ]]; then
        echo -e "${RED}Error: ANTHROPIC_API_KEY not set${NC}"
        echo ""
        echo "Set your API key:"
        echo "  export ANTHROPIC_API_KEY=sk-ant-xxx"
        echo ""
        echo "Or use --api-key option"
        exit 1
    fi

    # Check jq
    if ! command -v jq &> /dev/null; then
        echo -e "${RED}Error: jq is not installed${NC}"
        echo "Install: brew install jq"
        exit 1
    fi
}

# Function to expand org into repos
expand_org_repos() {
    local org="$1"
    echo -e "${BLUE}Fetching repositories for org: $org${NC}" >&2

    if ! command -v gh &> /dev/null; then
        echo -e "${RED}Error: gh (GitHub CLI) is required for org scanning${NC}" >&2
        echo "Install: brew install gh" >&2
        return 1
    fi

    local repos=$(gh repo list "$org" --limit 1000 --json nameWithOwner --jq '.[].nameWithOwner' 2>/dev/null || echo "")

    if [[ -z "$repos" ]]; then
        echo -e "${YELLOW}⚠ No repositories found for org: $org${NC}" >&2
        return 1
    fi

    echo "$repos"
}

# Function to load targets from config
load_config_targets() {
    if [[ ! -f "$CONFIG_FILE" ]]; then
        return 1
    fi

    local config_orgs=$(jq -r '.github.organizations[]?' "$CONFIG_FILE" 2>/dev/null || echo "")
    local config_repos=$(jq -r '.github.repositories[]?' "$CONFIG_FILE" 2>/dev/null || echo "")

    while IFS= read -r org; do
        [[ -n "$org" ]] && TARGETS_LIST+=("org:$org")
    done <<< "$config_orgs"

    while IFS= read -r repo; do
        [[ -n "$repo" ]] && TARGETS_LIST+=("repo:$repo")
    done <<< "$config_repos"

    return 0
}

# Function to run base analyzer and capture output
run_base_analyzer() {
    local target="$1"
    local output_file=$(mktemp)

    echo -e "${BLUE}Running provenance analysis...${NC}" >&2
    echo "" >&2

    local cmd="$SCRIPT_DIR/provenance-analyzer.sh"

    if [[ "$VERIFY_SIGNATURES" == "true" ]]; then
        cmd="$cmd --verify-signatures"
    fi

    if [[ $MIN_SLSA_LEVEL -gt 0 ]]; then
        cmd="$cmd --min-level $MIN_SLSA_LEVEL"
    fi

    cmd="$cmd $target"

    # Run and capture output
    if eval "$cmd" > "$output_file" 2>&1; then
        echo -e "${GREEN}✓ Base analysis complete${NC}" >&2
    else
        echo -e "${YELLOW}⚠ Analysis completed with findings${NC}" >&2
    fi

    echo "$output_file"
}

# Function to call Claude API for analysis
analyze_with_claude() {
    local scan_results="$1"
    local target_desc="$2"

    echo ""
    echo -e "${BLUE}Analyzing with Claude AI...${NC}"

    # Read scan results
    local results_content=$(cat "$scan_results")

    # Prepare prompt
    local prompt="Analyze these SLSA provenance scan results. Focus on AI-driven insights about supply chain trust and security.

Target: $target_desc

Provenance Analysis Results:
\`\`\`
$results_content
\`\`\`

Provide CONTEXTUAL ANALYSIS focusing on:

1. **Trust Assessment**
   - Builder identity patterns and trustworthiness signals
   - Source repository health indicators
   - Organizational trust patterns across packages
   - Anomalies or inconsistencies in provenance data

2. **Supply Chain Position Analysis**
   - Critical path packages and their provenance status
   - Direct vs transitive dependency provenance patterns
   - Ecosystem-specific trust models
   - Package popularity vs provenance quality correlation

3. **SLSA Level Distribution Context**
   - What the SLSA level distribution indicates about maturity
   - Gaps between industry best practices and current state
   - Comparison to ecosystem norms
   - Temporal patterns (older packages vs newer ones)

4. **Pattern Recognition**
   - Unusual build configurations or patterns
   - Inconsistencies across package versions
   - Clustering of provenance issues
   - Ecosystem-specific risk patterns

5. **Risk Narrative**
   - What story does this provenance data tell?
   - Systemic issues vs isolated gaps
   - Supply chain maturity assessment
   - Trust fabric analysis

NOTE: Focus on INSIGHTS, PATTERNS, and CONTEXT that require understanding and reasoning.
Do NOT provide specific remediation steps or prescriptive actions."

    # Call Claude API
    local response=$(curl -s https://api.anthropic.com/v1/messages \
        -H "content-type: application/json" \
        -H "x-api-key: $ANTHROPIC_API_KEY" \
        -H "anthropic-version: 2023-06-01" \
        -d "{
            \"model\": \"claude-sonnet-4-20250514\",
            \"max_tokens\": 4096,
            \"messages\": [{
                \"role\": \"user\",
                \"content\": $(echo "$prompt" | jq -Rs .)
            }]
        }")

    # Extract response
    local analysis=$(echo "$response" | jq -r '.content[0].text // empty')

    if [[ -z "$analysis" ]]; then
        echo -e "${RED}✗ Claude API error${NC}"
        echo "$response" | jq .
        return 1
    fi

    echo -e "${GREEN}✓ Analysis complete${NC}"
    echo ""
    echo "========================================="
    echo "  Claude AI Analysis"
    echo "========================================="
    echo ""
    echo "$analysis"
    echo ""
}

# Function to analyze a single target
analyze_single_target() {
    local target="$1"
    local TARGET_DESC="$target"

    # Run base analyzer
    local SCAN_RESULTS=$(run_base_analyzer "$target")

    if [[ -n "$SCAN_RESULTS" ]] && [[ -f "$SCAN_RESULTS" ]]; then
        analyze_with_claude "$SCAN_RESULTS" "$TARGET_DESC"
        rm -f "$SCAN_RESULTS"
    else
        echo -e "${RED}✗ Scan failed or produced no results${NC}"
    fi
}

# Parse command line arguments
TARGET=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --verify-signatures)
            VERIFY_SIGNATURES=true
            shift
            ;;
        --min-level)
            MIN_SLSA_LEVEL="$2"
            shift 2
            ;;
        -k|--api-key)
            ANTHROPIC_API_KEY="$2"
            shift 2
            ;;
        --keep-clone)
            CLEANUP=false
            shift
            ;;
        --org)
            TARGETS_LIST+=("org:$2")
            MULTI_REPO_MODE=true
            shift 2
            ;;
        --repo)
            TARGETS_LIST+=("repo:$2")
            MULTI_REPO_MODE=true
            shift 2
            ;;
        --config)
            CONFIG_FILE="$2"
            shift 2
            ;;
        -h|--help)
            usage
            ;;
        *)
            TARGET="$1"
            shift
            ;;
    esac
done

# Determine targets
if [[ "$MULTI_REPO_MODE" == false ]] && [[ -z "$TARGET" ]] && [[ ${#TARGETS_LIST[@]} -eq 0 ]]; then
    if ! load_config_targets; then
        echo -e "${RED}Error: No targets specified and no config file found${NC}"
        echo ""
        echo "Specify targets via:"
        echo "  - Single target: $0 <sbom|repo-url|directory|purl>"
        echo "  - Organization:  $0 --org myorg"
        echo "  - Repositories:  $0 --repo owner/repo1 --repo owner/repo2"
        echo "  - Config file:   Create $CONFIG_FILE with targets"
        echo ""
        exit 1
    fi
    MULTI_REPO_MODE=true
fi

# Main
echo ""
echo "========================================="
echo "  Provenance Analyzer with Claude AI"
echo "========================================="
echo ""

check_prerequisites

# Multi-repo mode or single target
if [[ "$MULTI_REPO_MODE" == true ]]; then
    echo -e "${BLUE}Multi-repository mode: ${#TARGETS_LIST[@]} target(s)${NC}"
    echo ""

    for target_spec in "${TARGETS_LIST[@]}"; do
        if [[ "$target_spec" =~ ^org: ]]; then
            org="${target_spec#org:}"
            repos=$(expand_org_repos "$org")

            if [[ -z "$repos" ]]; then
                continue
            fi

            while IFS= read -r repo; do
                if [[ -n "$repo" ]]; then
                    echo ""
                    echo -e "${CYAN}=========================================${NC}"
                    echo -e "${CYAN}Analyzing: $repo${NC}"
                    echo -e "${CYAN}=========================================${NC}"
                    echo ""
                    analyze_single_target "https://github.com/$repo"
                fi
            done <<< "$repos"

        elif [[ "$target_spec" =~ ^repo: ]]; then
            repo="${target_spec#repo:}"
            echo ""
            echo -e "${CYAN}=========================================${NC}"
            echo -e "${CYAN}Analyzing: $repo${NC}"
            echo -e "${CYAN}=========================================${NC}"
            echo ""
            analyze_single_target "https://github.com/$repo"
        fi
    done
else
    analyze_single_target "$TARGET"
fi

echo ""
echo "========================================="
echo -e "${GREEN}  Analysis Complete${NC}"
echo "========================================="
echo ""
