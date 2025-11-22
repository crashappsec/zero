#!/bin/bash
# Copyright (c) 2024 Gibson Powers Contributors
# 
# SPDX-License-Identifier: GPL-3.0

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m'

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ANTHROPIC_API_KEY="${ANTHROPIC_API_KEY:-}"

usage() {
    cat << EOF
DORA Analyzer Comparison Tool

Runs both basic and Claude-enhanced analyzers to demonstrate value-add.

Usage: $0 [OPTIONS] <deployment-data.json>

OPTIONS:
    -k, --api-key KEY       Anthropic API key
    -h, --help              Show help

EOF
    exit 1
}

check_prerequisites() {
    if [[ ! -x "$SCRIPT_DIR/dora-analyzer.sh" ]]; then
        echo -e "${RED}Error: dora-analyzer.sh not found${NC}"
        exit 1
    fi

    if [[ ! -x "$SCRIPT_DIR/dora-analyzer-claude.sh" ]]; then
        echo -e "${RED}Error: dora-analyzer-claude.sh not found${NC}"
        exit 1
    fi

    if [[ -z "$ANTHROPIC_API_KEY" ]]; then
        echo -e "${RED}Error: ANTHROPIC_API_KEY not set${NC}"
        exit 1
    fi
}

# Parse arguments
DATA_FILE=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -k|--api-key)
            ANTHROPIC_API_KEY="$2"
            shift 2
            ;;
        -h|--help)
            usage
            ;;
        *)
            DATA_FILE="$1"
            shift
            ;;
    esac
done

if [[ -z "$DATA_FILE" ]]; then
    usage
fi

# Main
echo ""
echo "========================================="
echo "  DORA Analyzer Comparison"
echo "========================================="
echo ""

check_prerequisites

echo -e "${BLUE}Running basic DORA analyzer...${NC}"
echo ""
"$SCRIPT_DIR/dora-analyzer.sh" "$DATA_FILE"

echo ""
echo -e "${BLUE}Running Claude-enhanced analyzer...${NC}"
echo ""
ANTHROPIC_API_KEY="$ANTHROPIC_API_KEY" "$SCRIPT_DIR/dora-analyzer-claude.sh" "$DATA_FILE"

echo ""
echo -e "${MAGENTA}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${MAGENTA}Value-Add Summary${NC}"
echo -e "${MAGENTA}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "Basic analyzer provides:"
echo "  • Metric calculations"
echo "  • Performance classifications"
echo "  • Benchmark comparisons"
echo ""
echo "Claude-enhanced analyzer adds:"
echo "  • ${GREEN}Executive summaries${NC} for leadership"
echo "  • ${GREEN}Root cause analysis${NC} of performance"
echo "  • ${GREEN}Prioritized recommendations${NC} with effort/impact"
echo "  • ${GREEN}Actionable roadmaps${NC} for improvement"
echo "  • ${GREEN}Contextual insights${NC} and best practices"
echo ""
echo "Use basic for: Quick metrics, CI/CD integration, dashboards"
echo "Use Claude for: Reports, planning, stakeholder communication"
echo ""

echo "========================================="
echo -e "${GREEN}  Comparison Complete${NC}"
echo "========================================="
echo ""
