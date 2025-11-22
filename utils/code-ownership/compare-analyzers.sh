#!/bin/bash
# Copyright (c) 2024 Gibson Powers Contributors
# 
# SPDX-License-Identifier: GPL-3.0

set -e

# Load environment variables from .env file if it exists
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
if [ -f "$REPO_ROOT/.env" ]; then
    source "$REPO_ROOT/.env"
fi

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m'

ANTHROPIC_API_KEY="${ANTHROPIC_API_KEY:-}"

usage() {
    cat << EOF
Code Ownership Analyzer Comparison Tool

Runs both basic and Claude-enhanced analyzers to demonstrate value-add.

Usage: $0 [OPTIONS] <repository-path>

OPTIONS:
    -k, --api-key KEY       Anthropic API key
    -d, --days N            Analyze last N days (default: 90)
    -h, --help              Show help

EOF
    exit 1
}

check_prerequisites() {
    if [[ ! -x "$SCRIPT_DIR/ownership-analyzer.sh" ]]; then
        echo -e "${RED}Error: ownership-analyzer.sh not found${NC}"
        exit 1
    fi

    if [[ ! -x "$SCRIPT_DIR/ownership-analyzer-claude.sh" ]]; then
        echo -e "${RED}Error: ownership-analyzer-claude.sh not found${NC}"
        exit 1
    fi

    if [[ -z "$ANTHROPIC_API_KEY" ]]; then
        echo -e "${RED}Error: ANTHROPIC_API_KEY not set${NC}"
        exit 1
    fi
}

# Parse arguments
REPO_PATH=""
DAYS=90

while [[ $# -gt 0 ]]; do
    case $1 in
        -k|--api-key)
            ANTHROPIC_API_KEY="$2"
            shift 2
            ;;
        -d|--days)
            DAYS="$2"
            shift 2
            ;;
        -h|--help)
            usage
            ;;
        *)
            REPO_PATH="$1"
            shift
            ;;
    esac
done

if [[ -z "$REPO_PATH" ]]; then
    usage
fi

# Main
echo ""
echo "========================================="
echo "  Code Ownership Analyzer Comparison"
echo "========================================="
echo ""

check_prerequisites

echo -e "${BLUE}Running basic ownership analyzer...${NC}"
echo ""
"$SCRIPT_DIR/ownership-analyzer.sh" --days "$DAYS" "$REPO_PATH"

echo ""
echo -e "${BLUE}Running Claude-enhanced analyzer...${NC}"
echo ""
ANTHROPIC_API_KEY="$ANTHROPIC_API_KEY" "$SCRIPT_DIR/ownership-analyzer-claude.sh" --days "$DAYS" "$REPO_PATH"

echo ""
echo -e "${MAGENTA}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${MAGENTA}Value-Add Summary${NC}"
echo -e "${MAGENTA}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "Basic analyzer provides:"
echo "  • Contribution statistics and metrics"
echo "  • Top contributors by commit count"
echo "  • Ownership distribution data"
echo "  • Activity summary"
echo "  • CODEOWNERS syntax validation"
echo ""
echo "Claude-enhanced analyzer adds:"
echo "  • ${GREEN}Executive summary${NC} with health assessment"
echo "  • ${GREEN}Risk analysis${NC} and bus factor calculation"
echo "  • ${GREEN}CODEOWNERS accuracy validation${NC} against actual contributions"
echo "  • ${GREEN}Prioritized recommendations${NC} with effort/impact estimates"
echo "  • ${GREEN}Actionable improvement plans${NC} (Priority 1/2/3)"
echo "  • ${GREEN}Knowledge transfer guidance${NC}"
echo "  • ${GREEN}Contextual insights${NC} and best practices"
echo ""
echo "Use basic for:"
echo "  - Quick metrics and statistics"
echo "  - CI/CD integration for automated checks"
echo "  - Data export (JSON/CSV) for dashboards"
echo ""
echo "Use Claude for:"
echo "  - Quarterly ownership audits"
echo "  - Strategic planning and improvements"
echo "  - Risk assessments for leadership"
echo "  - Knowledge transfer planning"
echo "  - CODEOWNERS validation with fix recommendations"
echo ""

echo "========================================="
echo -e "${GREEN}  Comparison Complete${NC}"
echo "========================================="
echo ""
