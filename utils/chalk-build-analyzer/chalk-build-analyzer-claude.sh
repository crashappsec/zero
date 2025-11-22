#!/bin/bash
# Copyright (c) 2024 Gibson Powers Contributors
# 
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Chalk Build Analyzer Script with Claude AI Integration
# Analyzes Chalk build reports with AI-enhanced insights
# Usage: ./chalk-build-analyzer-claude.sh [options] <chalk-report.json>
#############################################################################

set -e

# Load environment variables from .env file if it exists
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
if [ -f "$REPO_ROOT/.env" ]; then
    source "$REPO_ROOT/.env"
fi

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default options
COMPARE_MODE=false
BASELINE_FILE=""
ANTHROPIC_API_KEY="${ANTHROPIC_API_KEY:-}"

# Function to print usage
usage() {
    cat << EOF
Chalk Build Analyzer with Claude AI - Enhanced build performance analysis

Analyzes Chalk build reports using AI to provide intelligent insights,
actionable recommendations, and engineering velocity metrics.

Usage: $0 [OPTIONS] <chalk-report.json>

SINGLE BUILD ANALYSIS:
    $0 build-report.json

COMPARISON MODE:
    $0 --compare baseline.json current.json

OPTIONS:
    -c, --compare BASELINE  Compare current build against baseline
    -k, --api-key KEY       Anthropic API key (or set ANTHROPIC_API_KEY env var)
    -h, --help              Show this help message

ENVIRONMENT:
    ANTHROPIC_API_KEY       Your Anthropic API key

EXAMPLES:
    # Analyze a single build with AI insights
    $0 my-build.json

    # Compare builds with AI-powered regression analysis
    $0 --compare before.json after.json

    # Specify API key directly
    $0 --api-key sk-ant-xxx build.json

EOF
    exit 1
}

# Function to check prerequisites
check_prerequisites() {
    # Check jq
    if ! command -v jq &> /dev/null; then
        echo -e "${RED}Error: jq is not installed${NC}"
        echo "Install: brew install jq  (or apt-get install jq)"
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
}

# Function to validate Chalk report
validate_report() {
    local report="$1"

    if [[ ! -f "$report" ]]; then
        echo -e "${RED}Error: File not found: $report${NC}"
        exit 1
    fi

    # Check if valid JSON
    if ! jq empty "$report" 2>/dev/null; then
        echo -e "${RED}Error: Invalid JSON file${NC}"
        exit 1
    fi
}

# Function to call Claude API for single build analysis
analyze_with_claude() {
    local report="$1"

    echo ""
    echo -e "${BLUE}Analyzing with Claude AI...${NC}"

    # Read report
    local report_content=$(cat "$report")

    # Prepare prompt
    local prompt="I need you to analyze this Chalk build performance report.

Build Report:
\`\`\`json
$report_content
\`\`\`

Please provide:

1. Executive Summary
   - Build status and overall performance assessment
   - Key metrics (duration, success rate, efficiency)
   - Overall build health score

2. Performance Analysis
   - Build duration breakdown by stage
   - Identification of bottlenecks
   - Queue time and resource contention
   - Cache effectiveness

3. Engineering Velocity Metrics
   - Where available, calculate DORA metrics
   - Team productivity indicators
   - Build reliability and success rate
   - Velocity scoring

4. Bottleneck Identification
   - Which stages are consuming the most time
   - Resource utilization issues
   - Parallelization opportunities
   - Cache misses and inefficiencies

5. Actionable Recommendations
   - Prioritized list of improvements (High/Medium/Low)
   - Expected impact of each recommendation
   - Implementation effort estimates
   - Quick wins vs. strategic improvements

6. Cost & Efficiency
   - Resource waste analysis
   - Potential cost savings
   - ROI of suggested optimizations

Please be specific, data-driven, and focus on actionable insights."

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
    echo "  Claude AI Build Analysis"
    echo "========================================="
    echo ""
    echo "$analysis"
    echo ""
}

# Function to call Claude API for build comparison
compare_with_claude() {
    local baseline="$1"
    local current="$2"

    echo ""
    echo -e "${BLUE}Comparing builds with Claude AI...${NC}"

    # Read reports
    local baseline_content=$(cat "$baseline")
    local current_content=$(cat "$current")

    # Prepare prompt
    local prompt="I need you to perform a regression analysis comparing two Chalk build reports.

Baseline Build:
\`\`\`json
$baseline_content
\`\`\`

Current Build:
\`\`\`json
$current_content
\`\`\`

Please provide:

1. Comparison Summary
   - Build IDs and time gap between builds
   - Same branch/environment verification
   - Overall performance change

2. Performance Regression Detection
   - Build duration change (absolute and percentage)
   - Stage-by-stage comparison
   - Identify which stages got slower/faster
   - Statistical significance assessment

3. Root Cause Analysis
   - What changed to cause performance differences
   - Code changes vs. environmental factors
   - Test additions/removals impact
   - Resource allocation changes

4. Impact Assessment
   - Severity rating (NONE/LOW/MEDIUM/HIGH/CRITICAL)
   - Impact on team velocity
   - Cost implications
   - Reliability changes

5. Recommendations
   - Should this build be blocked?
   - What needs to be fixed?
   - Prioritized action items
   - Prevention strategies

6. Efficiency Comparison
   - Cache effectiveness changes
   - Resource utilization changes
   - Parallelization changes
   - Overall efficiency trend

Please be specific about what regressed and why."

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

    echo -e "${GREEN}✓ Comparison complete${NC}"
    echo ""
    echo "========================================="
    echo "  Claude AI Regression Analysis"
    echo "========================================="
    echo ""
    echo "$analysis"
    echo ""
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
        -h|--help)
            usage
            ;;
        *)
            REPORT_FILE="$1"
            shift
            ;;
    esac
done

# Validate arguments
if [[ -z "$REPORT_FILE" ]]; then
    echo -e "${RED}Error: No report file specified${NC}"
    usage
fi

# Main
echo ""
echo "========================================="
echo "  Chalk Build Analyzer with Claude AI"
echo "========================================="
echo ""

check_prerequisites

if [[ "$COMPARE_MODE" == true ]]; then
    if [[ -z "$BASELINE_FILE" ]]; then
        echo -e "${RED}Error: No baseline file specified for comparison${NC}"
        exit 1
    fi

    validate_report "$BASELINE_FILE"
    validate_report "$REPORT_FILE"
    compare_with_claude "$BASELINE_FILE" "$REPORT_FILE"
else
    validate_report "$REPORT_FILE"
    analyze_with_claude "$REPORT_FILE"
fi

echo "========================================="
echo -e "${GREEN}  Analysis Complete${NC}"
echo "========================================="
echo ""
