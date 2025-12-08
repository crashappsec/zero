#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Phantom Preflight Check
# Verifies all required tools and API keys are configured
# Run this before hydrating a project
#
# Usage: ./preflight.sh [--fix]
#############################################################################

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ZERO_DIR="$(dirname "$SCRIPT_DIR")"
UTILS_ROOT="$(dirname "$ZERO_DIR")"
REPO_ROOT="$(dirname "$UTILS_ROOT")"

# Load Phantom library for banner
source "$ZERO_DIR/lib/zero-lib.sh"

# Load .env if it exists
if [[ -f "$REPO_ROOT/.env" ]]; then
    set -a
    source "$REPO_ROOT/.env"
    set +a
fi

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

# Counters
ERRORS=0
WARNINGS=0

#############################################################################
# Usage
#############################################################################

usage() {
    cat << EOF
Phantom Preflight Check - Verify tools and configuration

Usage: $0 [options]

OPTIONS:
    -h, --help  Show this help

CHECKS:
    Required Tools:
      • git         - Repository cloning
      • jq          - JSON processing
      • curl        - API requests

    Recommended Tools:
      • osv-scanner - Vulnerability scanning (Google OSV)
      • syft        - SBOM generation (Anchore)
      • gh          - GitHub CLI for enhanced features
      • semgrep     - AST-aware code scanning
      • checkov     - IaC security scanning
      • malcontent  - Supply chain compromise detection (Chainguard)
      • trivy       - Container image vulnerability scanning (Aqua)
      • hadolint    - Dockerfile linting

    API Keys:
      • GITHUB_TOKEN      - Required for private repos, recommended for rate limits
      • ANTHROPIC_API_KEY - Required for Claude-enhanced analysis

EOF
    exit 0
}

#############################################################################
# Check Functions
#############################################################################

# Arrays to track missing tools for installation prompt
MISSING_BREW_TOOLS=()
MISSING_PIP_TOOLS=()

check_tool() {
    local tool="$1"
    local required="$2"
    local install_cmd="$3"
    local description="$4"

    printf "  %-16s " "$tool"

    if command -v "$tool" &> /dev/null; then
        local version=$("$tool" --version 2>/dev/null | head -1 | grep -oE '[0-9]+\.[0-9]+(\.[0-9]+)?' | head -1)
        echo -e "${GREEN}✓${NC} ${version:-installed}"
        return 0
    else
        if [[ "$required" == "required" ]]; then
            echo -e "${RED}✗ missing (required)${NC}"
            ((ERRORS++))
        else
            echo -e "${YELLOW}○ missing${NC}"
            ((WARNINGS++))
        fi

        # Track for later installation
        if [[ "$install_cmd" == "brew install"* ]]; then
            MISSING_BREW_TOOLS+=("$tool")
        fi
        return 1
    fi
}

# Special check for checkov (may be in Python user paths)
check_checkov() {
    printf "  %-16s " "checkov"

    local checkov_bin=""
    if command -v checkov &> /dev/null; then
        checkov_bin="checkov"
    elif [[ -x "$HOME/Library/Python/3.9/bin/checkov" ]]; then
        checkov_bin="$HOME/Library/Python/3.9/bin/checkov"
    elif [[ -x "$HOME/Library/Python/3.10/bin/checkov" ]]; then
        checkov_bin="$HOME/Library/Python/3.10/bin/checkov"
    elif [[ -x "$HOME/Library/Python/3.11/bin/checkov" ]]; then
        checkov_bin="$HOME/Library/Python/3.11/bin/checkov"
    elif [[ -x "$HOME/Library/Python/3.12/bin/checkov" ]]; then
        checkov_bin="$HOME/Library/Python/3.12/bin/checkov"
    elif [[ -x "$HOME/.local/bin/checkov" ]]; then
        checkov_bin="$HOME/.local/bin/checkov"
    fi

    if [[ -n "$checkov_bin" ]]; then
        local version=$("$checkov_bin" --version 2>/dev/null | head -1 | grep -oE '[0-9]+\.[0-9]+(\.[0-9]+)?' | head -1)
        echo -e "${GREEN}✓${NC} ${version:-installed}"
        return 0
    else
        echo -e "${YELLOW}○ missing${NC}"
        MISSING_PIP_TOOLS+=("checkov")
        ((WARNINGS++))
        return 1
    fi
}

# Special check for malcontent (binary is 'mal' from brew install malcontent)
check_malcontent() {
    printf "  %-16s " "malcontent"

    local mal_bin=""
    if command -v mal &> /dev/null; then
        mal_bin="mal"
    elif [[ -x "/opt/homebrew/bin/mal" ]]; then
        mal_bin="/opt/homebrew/bin/mal"
    elif [[ -x "/usr/local/bin/mal" ]]; then
        mal_bin="/usr/local/bin/mal"
    fi

    if [[ -n "$mal_bin" ]]; then
        local version=$("$mal_bin" --version 2>/dev/null | head -1 | grep -oE '[0-9]+\.[0-9]+(\.[0-9]+)?' | head -1)
        echo -e "${GREEN}✓${NC} ${version:-installed}"
        return 0
    else
        echo -e "${YELLOW}○ missing${NC}"
        echo -e "    ${CYAN}Install: brew install malcontent${NC}"
        ((WARNINGS++))
        return 1
    fi
}

# Offer to install missing tools
offer_install() {
    # Only offer if running interactively
    [[ ! -t 0 ]] && return 0

    if [[ ${#MISSING_BREW_TOOLS[@]} -gt 0 ]] && command -v brew &> /dev/null; then
        echo
        echo -e "${BOLD}Missing tools: ${MISSING_BREW_TOOLS[*]}${NC}"
        read -p "Install via Homebrew? (y/n) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            for tool in "${MISSING_BREW_TOOLS[@]}"; do
                echo -e "  ${BLUE}Installing $tool...${NC}"
                if brew install "$tool" 2>/dev/null; then
                    echo -e "  ${GREEN}✓${NC} $tool installed"
                    ((WARNINGS--))
                else
                    echo -e "  ${RED}✗${NC} Failed to install $tool"
                fi
            done
        fi
    fi

    if [[ ${#MISSING_PIP_TOOLS[@]} -gt 0 ]] && command -v pip3 &> /dev/null; then
        echo
        echo -e "${BOLD}Missing Python tools: ${MISSING_PIP_TOOLS[*]}${NC}"
        read -p "Install via pip3? (y/n) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            for tool in "${MISSING_PIP_TOOLS[@]}"; do
                echo -e "  ${BLUE}Installing $tool...${NC}"
                if pip3 install --user "$tool" 2>/dev/null; then
                    echo -e "  ${GREEN}✓${NC} $tool installed"
                    ((WARNINGS--))
                else
                    echo -e "  ${RED}✗${NC} Failed to install $tool"
                fi
            done
        fi
    fi
}

check_api_key() {
    local key_name="$1"
    local required="$2"
    local description="$3"

    printf "  %-16s " "$key_name"

    local key_value="${!key_name}"

    if [[ -n "$key_value" ]]; then
        local masked="${key_value:0:8}...${key_value: -4}"
        echo -e "${GREEN}✓${NC} configured ($masked)"
        return 0
    else
        if [[ "$required" == "required" ]]; then
            echo -e "${RED}✗ not set (required)${NC}"
            ((ERRORS++))
        else
            echo -e "${YELLOW}○ not set (recommended)${NC}"
            ((WARNINGS++))
        fi

        if [[ -n "$description" ]]; then
            echo -e "    ${CYAN}$description${NC}"
        fi
        return 1
    fi
}

check_github_auth() {
    printf "  %-16s " "gh auth"

    if command -v gh &> /dev/null; then
        if gh auth status &> /dev/null; then
            local user=$(gh api user -q '.login' 2>/dev/null || echo "authenticated")
            echo -e "${GREEN}✓${NC} logged in as $user"
            return 0
        else
            echo -e "${YELLOW}○ not authenticated${NC}"
            echo -e "    ${CYAN}Run: gh auth login${NC}"
            ((WARNINGS++))
            return 1
        fi
    else
        echo -e "${YELLOW}○ gh not installed${NC}"
        return 1
    fi
}

check_directory() {
    local dir="$1"
    local description="$2"

    printf "  %-16s " "$description"

    if [[ -d "$dir" ]]; then
        local count=$(ls -1 "$dir" 2>/dev/null | wc -l | tr -d ' ')
        echo -e "${GREEN}✓${NC} exists ($count items)"
        return 0
    else
        echo -e "${YELLOW}○ not created yet${NC}"
        return 1
    fi
}

#############################################################################
# Main Checks
#############################################################################

run_checks() {
    print_zero_banner
    echo -e "${BOLD}Preflight Check${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo

    # Required Tools
    echo -e "${BOLD}Required Tools${NC}"
    check_tool "git" "required" "brew install git" "Version control"
    check_tool "jq" "required" "brew install jq" "JSON processing"
    check_tool "curl" "required" "brew install curl" "HTTP requests"
    echo

    # Recommended Tools (use || true to not fail on missing optional tools)
    echo -e "${BOLD}Recommended Tools${NC}"
    check_tool "osv-scanner" "recommended" "brew install osv-scanner" "Vulnerability scanning" || true
    check_tool "syft" "recommended" "brew install syft" "SBOM generation" || true
    check_tool "gh" "recommended" "brew install gh" "GitHub CLI" || true
    check_tool "semgrep" "recommended" "brew install semgrep" "AST-aware code scanning" || true
    check_checkov || true
    check_malcontent || true
    check_tool "trivy" "recommended" "brew install trivy" "Container vulnerability scanning" || true
    check_tool "hadolint" "recommended" "brew install hadolint" "Dockerfile linting" || true
    echo

    # API Keys (use || true to not fail on missing optional keys)
    echo -e "${BOLD}API Keys${NC}"
    check_api_key "GITHUB_TOKEN" "recommended" "Create at: https://github.com/settings/tokens" || true
    check_api_key "ANTHROPIC_API_KEY" "recommended" "Get from: https://console.anthropic.com/" || true
    echo

    # GitHub Authentication
    echo -e "${BOLD}GitHub Authentication${NC}"
    check_github_auth || true
    echo

    # Phantom Directory (these are auto-created on first run, so just informational)
    echo -e "${BOLD}Phantom Storage${NC}"
    check_directory "$HOME/.zero" "~/.zero" || true
    check_directory "$HOME/.zero/projects" "projects" || true
    echo

    # Offer to install missing tools
    offer_install

    # Summary
    echo
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    if [[ $ERRORS -eq 0 ]] && [[ $WARNINGS -eq 0 ]]; then
        echo -e "${GREEN}${BOLD}✓ All checks passed!${NC}"
        echo
        echo "Ready to hydrate a project:"
        echo -e "  ${CYAN}./zero.sh hydrate <owner/repo>${NC}"
        return 0
    elif [[ $ERRORS -eq 0 ]]; then
        echo -e "${YELLOW}${BOLD}⚠ $WARNINGS warnings${NC}"
        echo
        echo "You can proceed, but some features may be limited."
        echo "Ready to hydrate a project:"
        echo -e "  ${CYAN}./zero.sh hydrate <owner/repo>${NC}"
        return 0
    else
        echo -e "${RED}${BOLD}✗ $ERRORS errors, $WARNINGS warnings${NC}"
        echo
        echo "Please fix the errors above before continuing."
        return 1
    fi
}

#############################################################################
# Main
#############################################################################

while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help)
            usage
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

run_checks
