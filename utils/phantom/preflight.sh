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
UTILS_ROOT="$(dirname "$SCRIPT_DIR")"
REPO_ROOT="$(dirname "$UTILS_ROOT")"

# Load Gibson library for banner
source "$SCRIPT_DIR/lib/gibson.sh"

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

FIX_MODE=false
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
    --fix       Attempt to install missing tools (requires Homebrew on macOS)
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

    API Keys:
      • GITHUB_TOKEN      - Required for private repos, recommended for rate limits
      • ANTHROPIC_API_KEY - Required for Claude-enhanced analysis

EOF
    exit 0
}

#############################################################################
# Check Functions
#############################################################################

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
            echo -e "${YELLOW}○ missing (recommended)${NC}"
            ((WARNINGS++))
        fi

        if [[ "$FIX_MODE" == true ]] && [[ -n "$install_cmd" ]]; then
            echo -e "    ${BLUE}Installing...${NC}"
            if eval "$install_cmd" > /dev/null 2>&1; then
                echo -e "    ${GREEN}✓ Installed${NC}"
                return 0
            else
                echo -e "    ${RED}✗ Install failed${NC}"
            fi
        elif [[ -n "$install_cmd" ]]; then
            echo -e "    ${CYAN}Install: $install_cmd${NC}"
        fi
        return 1
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
    print_phantom_banner
    echo -e "${BOLD}Preflight Check${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo

    # Required Tools
    echo -e "${BOLD}Required Tools${NC}"
    check_tool "git" "required" "brew install git" "Version control"
    check_tool "jq" "required" "brew install jq" "JSON processing"
    check_tool "curl" "required" "brew install curl" "HTTP requests"
    echo

    # Recommended Tools
    echo -e "${BOLD}Recommended Tools${NC}"
    check_tool "osv-scanner" "recommended" "brew install osv-scanner" "Vulnerability scanning"
    check_tool "syft" "recommended" "brew install syft" "SBOM generation"
    check_tool "gh" "recommended" "brew install gh" "GitHub CLI"
    echo

    # API Keys
    echo -e "${BOLD}API Keys${NC}"
    check_api_key "GITHUB_TOKEN" "recommended" "Create at: https://github.com/settings/tokens"
    check_api_key "ANTHROPIC_API_KEY" "recommended" "Get from: https://console.anthropic.com/"
    echo

    # GitHub Authentication
    echo -e "${BOLD}GitHub Authentication${NC}"
    check_github_auth
    echo

    # Phantom Directory (these are auto-created on first run, so just informational)
    echo -e "${BOLD}Phantom Storage${NC}"
    check_directory "$HOME/.phantom" "~/.phantom" || true
    check_directory "$HOME/.phantom/projects" "projects" || true
    echo

    # Summary
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    if [[ $ERRORS -eq 0 ]] && [[ $WARNINGS -eq 0 ]]; then
        echo -e "${GREEN}${BOLD}✓ All checks passed!${NC}"
        echo
        echo "Ready to hydrate a project:"
        echo -e "  ${CYAN}./utils/phantom/hydrate.sh <owner/repo>${NC}"
        return 0
    elif [[ $ERRORS -eq 0 ]]; then
        echo -e "${YELLOW}${BOLD}⚠ $WARNINGS warnings${NC}"
        echo
        echo "You can proceed, but some features may be limited."
        echo "Ready to hydrate a project:"
        echo -e "  ${CYAN}./utils/phantom/hydrate.sh <owner/repo>${NC}"
        return 0
    else
        echo -e "${RED}${BOLD}✗ $ERRORS errors, $WARNINGS warnings${NC}"
        echo
        echo "Please fix the errors above before continuing."
        if [[ "$FIX_MODE" != true ]]; then
            echo -e "Run with ${CYAN}--fix${NC} to attempt automatic installation."
        fi
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
        --fix)
            FIX_MODE=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

run_checks
