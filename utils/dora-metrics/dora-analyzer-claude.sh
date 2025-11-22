#!/bin/bash
# Copyright (c) 2024 Gibson Powers Contributors
# 
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# DORA Metrics Analyzer with Claude AI Integration
# Analyzes DORA metrics with AI-enhanced insights and recommendations
# Usage: ./dora-analyzer-claude.sh [options] <deployment-data.json>
#############################################################################

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

ANTHROPIC_API_KEY="${ANTHROPIC_API_KEY:-}"

usage() {
    cat << EOF
DORA Metrics Analyzer with Claude AI - Enhanced performance analysis

Calculates DORA metrics and provides AI-powered insights:
- Performance classification and benchmarking
- Root cause analysis
- Improvement recommendations
- Executive summaries

Usage: $0 [OPTIONS] <deployment-data.json>

OPTIONS:
    -k, --api-key KEY       Anthropic API key (or set ANTHROPIC_API_KEY env var)
    -h, --help              Show this help message

ENVIRONMENT:
    ANTHROPIC_API_KEY       Your Anthropic API key

EXAMPLES:
    # Analyze with AI insights
    $0 deployment-data.json

    # Specify API key directly
    $0 --api-key sk-ant-xxx deployment-data.json

EOF
    exit 1
}

check_prerequisites() {
    if ! command -v jq &> /dev/null; then
        echo -e "${RED}Error: jq is not installed${NC}"
        exit 1
    fi

    if [[ -z "$ANTHROPIC_API_KEY" ]]; then
        echo -e "${RED}Error: ANTHROPIC_API_KEY not set${NC}"
        echo ""
        echo "Set your API key:"
        echo "  export ANTHROPIC_API_KEY=sk-ant-xxx"
        exit 1
    fi
}

analyze_with_claude() {
    local data_file="$1"

    echo ""
    echo -e "${BLUE}Analyzing DORA metrics with Claude AI...${NC}"

    local data_content=$(cat "$data_file")

    local prompt="I need you to analyze this DORA metrics deployment data.

Data:
\`\`\`json
$data_content
\`\`\`

Please provide:

1. Executive Summary
   - Overall performance classification (Elite/High/Medium/Low)
   - Key highlights and achievements
   - Primary concerns

2. Metric Analysis
   For each DORA metric:
   - Current performance and classification
   - Trend analysis if historical data available
   - How it compares to benchmarks

3. Strengths
   - What the team is doing well
   - Practices to maintain
   - Lessons to share

4. Improvement Opportunities
   - Prioritized recommendations (High/Medium/Low priority)
   - Specific actions to take
   - Expected impact of each recommendation
   - Implementation effort estimates

5. Roadmap
   - Short-term actions (1-4 weeks)
   - Medium-term initiatives (1-3 months)
   - Long-term goals (3-6 months)

Be specific, actionable, and data-driven."

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

    local analysis=$(echo "$response" | jq -r '.content[0].text // empty')

    if [[ -z "$analysis" ]]; then
        echo -e "${RED}✗ Claude API error${NC}"
        echo "$response" | jq .
        return 1
    fi

    echo -e "${GREEN}✓ Analysis complete${NC}"
    echo ""
    echo "========================================="
    echo "  Claude AI DORA Metrics Analysis"
    echo "========================================="
    echo ""
    echo "$analysis"
    echo ""
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
    echo -e "${RED}Error: No data file specified${NC}"
    usage
fi

# Main
echo ""
echo "========================================="
echo "  DORA Analyzer with Claude AI"
echo "========================================="
echo ""

check_prerequisites

if [[ ! -f "$DATA_FILE" ]]; then
    echo -e "${RED}Error: File not found: $DATA_FILE${NC}"
    exit 1
fi

analyze_with_claude "$DATA_FILE"

echo "========================================="
echo -e "${GREEN}  Analysis Complete${NC}"
echo "========================================="
echo ""
