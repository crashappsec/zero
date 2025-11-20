#!/bin/bash
# Copyright (c) 2024 Crash Override Inc
# 101 Fulton St, 416, New York 10038
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Chalk Analyzer Comparison Script
# Runs both basic and Claude-enhanced versions and compares results
# Usage: ./compare-analyzers.sh [options] <chalk-report.json>
#############################################################################

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Default options
COMPARE_MODE=false
BASELINE_FILE=""
KEEP_OUTPUTS=false
ANTHROPIC_API_KEY="${ANTHROPIC_API_KEY:-}"

# Function to print usage
usage() {
    cat << EOF
Chalk Analyzer Comparison Tool

Runs both basic build analysis and Claude-enhanced analysis,
then compares the results to show the value-add of AI enhancement.

Usage: $0 [OPTIONS] <chalk-report.json>

SINGLE BUILD ANALYSIS:
    $0 build-report.json

COMPARISON MODE:
    $0 --compare baseline.json current.json

OPTIONS:
    -c, --compare BASELINE  Compare two builds (runs both analyzers in compare mode)
    -k, --api-key KEY       Anthropic API key (or set ANTHROPIC_API_KEY env var)
    --keep-outputs          Keep output files for manual inspection
    -h, --help              Show this help message

EXAMPLES:
    # Compare basic vs Claude analysis of a single build
    $0 my-build.json

    # Compare regression detection between basic and Claude
    $0 --compare before.json after.json

    # Keep output files for review
    $0 --keep-outputs build.json

EOF
    exit 1
}

# Function to check prerequisites
check_prerequisites() {
    # Check if basic analyzer exists
    if [[ ! -x "$SCRIPT_DIR/chalk-build-analyzer.sh" ]]; then
        echo -e "${RED}Error: chalk-build-analyzer.sh not found or not executable${NC}"
        exit 1
    fi

    # Check if Claude analyzer exists
    if [[ ! -x "$SCRIPT_DIR/chalk-build-analyzer-claude.sh" ]]; then
        echo -e "${RED}Error: chalk-build-analyzer-claude.sh not found or not executable${NC}"
        exit 1
    fi

    # Check API key for Claude version
    if [[ -z "$ANTHROPIC_API_KEY" ]]; then
        echo -e "${RED}Error: ANTHROPIC_API_KEY not set${NC}"
        echo ""
        echo "Set your API key:"
        echo "  export ANTHROPIC_API_KEY=sk-ant-xxx"
        echo ""
        echo "Or use --api-key option"
        exit 1
    fi
}

# Function to run basic analyzer
run_basic_analyzer() {
    local report="$1"
    local baseline="$2"
    local output_file="$3"

    echo -e "${BLUE}Running basic build analysis...${NC}"
    echo ""

    local cmd="$SCRIPT_DIR/chalk-build-analyzer.sh"

    if [[ -n "$baseline" ]]; then
        cmd="$cmd --compare $baseline"
    fi

    cmd="$cmd $report"

    # Run and capture output
    if eval "$cmd" > "$output_file" 2>&1; then
        echo -e "${GREEN}✓ Basic analysis complete${NC}"
    else
        echo -e "${YELLOW}⚠ Basic analysis completed with warnings${NC}"
    fi

    echo ""
}

# Function to run Claude analyzer
run_claude_analyzer() {
    local report="$1"
    local baseline="$2"
    local output_file="$3"

    echo -e "${BLUE}Running Claude-enhanced analysis...${NC}"
    echo ""

    local cmd="ANTHROPIC_API_KEY=$ANTHROPIC_API_KEY $SCRIPT_DIR/chalk-build-analyzer-claude.sh"

    if [[ -n "$baseline" ]]; then
        cmd="$cmd --compare $baseline"
    fi

    cmd="$cmd $report"

    # Run and capture output
    if eval "$cmd" > "$output_file" 2>&1; then
        echo -e "${GREEN}✓ Claude analysis complete${NC}"
    else
        echo -e "${YELLOW}⚠ Claude analysis completed with warnings${NC}"
    fi

    echo ""
}

# Function to generate comparison report
generate_comparison() {
    local basic_output="$1"
    local claude_output="$2"
    local report="$3"
    local mode="$4"

    echo ""
    echo "========================================="
    echo "  COMPARISON REPORT"
    echo "========================================="
    echo ""

    echo -e "${CYAN}Target:${NC} $report"
    if [[ "$mode" == "compare" ]]; then
        echo -e "${CYAN}Mode:${NC} Regression Analysis"
    else
        echo -e "${CYAN}Mode:${NC} Single Build Analysis"
    fi
    echo ""

    echo -e "${MAGENTA}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${MAGENTA}Basic Analyzer Results${NC}"
    echo -e "${MAGENTA}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo "Output Format: Structured text analysis"
    echo "Provides:"
    echo "  • Build metrics (duration, status, platform)"
    echo "  • Stage breakdown when available"
    echo "  • Cache effectiveness metrics"
    echo "  • Resource utilization data"
    echo "  • Basic performance categorization"
    if [[ "$mode" == "compare" ]]; then
        echo "  • Numeric regression detection"
        echo "  • Simple threshold-based alerts"
    fi
    echo ""

    echo -e "${MAGENTA}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${MAGENTA}Claude-Enhanced Results${NC}"
    echo -e "${MAGENTA}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo "Output Format: AI-powered natural language analysis"
    echo ""
    echo "Additional Intelligence Provided:"
    echo "  • ${GREEN}Executive Summary${NC} - High-level build health assessment"
    echo "  • ${GREEN}Performance Analysis${NC} - Deep bottleneck identification"
    echo "  • ${GREEN}Engineering Velocity${NC} - DORA metrics and team productivity"
    echo "  • ${GREEN}Actionable Recommendations${NC} - Prioritized with effort/impact"
    echo "  • ${GREEN}Cost & Efficiency${NC} - Resource waste and ROI analysis"

    if [[ "$mode" == "compare" ]]; then
        echo "  • ${GREEN}Root Cause Analysis${NC} - Why builds got slower/faster"
        echo "  • ${GREEN}Severity Assessment${NC} - Should the build be blocked?"
        echo "  • ${GREEN}Prevention Strategies${NC} - How to avoid future regressions"
    fi
    echo ""

    echo -e "${MAGENTA}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${MAGENTA}Value-Add Summary${NC}"
    echo -e "${MAGENTA}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo "The Claude-enhanced version provides:"
    echo ""
    echo "1. ${CYAN}Contextual Understanding${NC}"
    echo "   Transform raw metrics into business insights about team velocity"
    echo ""
    echo "2. ${CYAN}Intelligent Bottleneck Detection${NC}"
    echo "   Not just slow stages - understand WHY and WHAT to do"
    echo "   - Root cause identification"
    echo "   - Parallelization opportunities"
    echo "   - Cache optimization strategies"
    echo ""
    echo "3. ${CYAN}Engineering Velocity Metrics${NC}"
    echo "   DORA metrics calculation and interpretation"
    echo "   - Deployment frequency"
    echo "   - Lead time for changes"
    echo "   - Team productivity indicators"
    echo ""
    echo "4. ${CYAN}Prioritized Action Plan${NC}"
    echo "   Specific recommendations with:"
    echo "   - Priority level (High/Medium/Low)"
    echo "   - Expected impact (time/cost savings)"
    echo "   - Implementation effort"
    echo "   - Quick wins vs. strategic improvements"
    echo ""
    echo "5. ${CYAN}Cost & ROI Analysis${NC}"
    echo "   Resource waste identification and savings opportunities"
    echo ""

    if [[ "$mode" == "compare" ]]; then
        echo "6. ${CYAN}Regression Intelligence${NC}"
        echo "   Beyond simple threshold alerts:"
        echo "   - Root cause identification"
        echo "   - Statistical significance"
        echo "   - Severity assessment with business context"
        echo "   - Block/proceed recommendations"
        echo ""
    fi

    echo -e "${MAGENTA}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${MAGENTA}Recommendation${NC}"
    echo -e "${MAGENTA}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo "Use ${GREEN}basic analyzer${NC} when:"
    echo "  • Quick metrics needed for dashboards"
    echo "  • Automated CI/CD checks"
    echo "  • High-frequency monitoring"
    echo "  • Cost optimization (no API calls)"
    echo ""
    echo "Use ${GREEN}Claude-enhanced analyzer${NC} when:"
    echo "  • Creating reports for engineering leaders"
    echo "  • Investigating build performance issues"
    echo "  • Making infrastructure investment decisions"
    echo "  • Understanding team velocity trends"
    echo "  • Analyzing complex regressions"
    echo "  • Need actionable recommendations"
    echo ""

    if [[ "$KEEP_OUTPUTS" == "true" ]]; then
        echo -e "${CYAN}Output Files Saved:${NC}"
        echo "  Basic:  $basic_output"
        echo "  Claude: $claude_output"
        echo ""
    fi
}

# Function to cleanup
cleanup() {
    if [[ "$KEEP_OUTPUTS" == "false" ]]; then
        rm -f /tmp/chalk_basic_*.txt
        rm -f /tmp/chalk_claude_*.txt
    fi
}

# Parse command line arguments
REPORT_FILE=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -c|--compare)
            COMPARE_MODE=true
            BASELINE_FILE="$2"
            shift 2
            ;;
        -k|--api-key)
            ANTHROPIC_API_KEY="$2"
            shift 2
            ;;
        --keep-outputs)
            KEEP_OUTPUTS=true
            shift
            ;;
        -h|--help)
            usage
            ;;
        *)
            REPORT_FILE="$1"
            shift
            ;;
    esac
done

# Validate
if [[ -z "$REPORT_FILE" ]]; then
    echo -e "${RED}Error: No report file specified${NC}"
    usage
fi

# Main
echo ""
echo "========================================="
echo "  Chalk Analyzer Comparison"
echo "========================================="
echo ""

check_prerequisites

# Create temporary output files
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BASIC_OUTPUT="/tmp/chalk_basic_${TIMESTAMP}.txt"
CLAUDE_OUTPUT="/tmp/chalk_claude_${TIMESTAMP}.txt"

# Determine mode
if [[ "$COMPARE_MODE" == true ]]; then
    MODE="compare"
else
    MODE="single"
fi

# Run both analyzers
if [[ "$COMPARE_MODE" == true ]]; then
    run_basic_analyzer "$REPORT_FILE" "$BASELINE_FILE" "$BASIC_OUTPUT"
    run_claude_analyzer "$REPORT_FILE" "$BASELINE_FILE" "$CLAUDE_OUTPUT"
else
    run_basic_analyzer "$REPORT_FILE" "" "$BASIC_OUTPUT"
    run_claude_analyzer "$REPORT_FILE" "" "$CLAUDE_OUTPUT"
fi

# Generate comparison report
generate_comparison "$BASIC_OUTPUT" "$CLAUDE_OUTPUT" "$REPORT_FILE" "$MODE"

# Cleanup if requested
if [[ "$KEEP_OUTPUTS" == "false" ]]; then
    cleanup
fi

echo "========================================="
echo -e "${GREEN}  Comparison Complete${NC}"
echo "========================================="
echo ""
