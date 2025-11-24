#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# 
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# DORA Metrics Analyser Script
# Calculates DORA metrics from deployment data
# Usage: ./dora-analyser.sh [options] <deployment-data.json>
#############################################################################

set -e

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
UTILS_ROOT="$(dirname "$SCRIPT_DIR")"
REPO_ROOT="$(dirname "$UTILS_ROOT")"

# Load .env file if it exists in repository root
if [[ -f "$REPO_ROOT/.env" ]]; then
    set -a  # automatically export all variables
    source "$REPO_ROOT/.env"
    set +a  # stop automatically exporting
fi

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Default options
OUTPUT_FORMAT="markdown"
OUTPUT_FILE=""
USE_CLAUDE=false
COMPARE_MODE=false
ANTHROPIC_API_KEY="${ANTHROPIC_API_KEY:-}"

# Function to print usage
usage() {
    cat << EOF
DORA Metrics Analyser - Calculate software delivery performance metrics

Calculates the four key DORA metrics from deployment data:
- Deployment Frequency
- Lead Time for Changes
- Change Failure Rate
- Time to Restore Service (MTTR)

Usage: $0 [OPTIONS] <deployment-data.json>

OPTIONS:
    -f, --format FORMAT     Output format: text|json|csv|markdown (default: markdown)
    -o, --output FILE       Write results to file
    --claude                Use Claude AI for advanced analysis (requires ANTHROPIC_API_KEY)
    --compare               Run both basic and Claude modes side-by-side for comparison
    -k, --api-key KEY       Anthropic API key (or set ANTHROPIC_API_KEY env var)
    -h, --help              Show this help message

INPUT FORMAT:
    JSON file with deployment and incident data. See examples/ for schema.

EXAMPLES:
    # Analyze deployment data
    $0 deployment-data.json

    # Export to JSON
    $0 --format json --output metrics.json deployment-data.json

    # Export to CSV for spreadsheet
    $0 --format csv --output metrics.csv deployment-data.json

    # Use Claude AI for advanced insights
    $0 --claude deployment-data.json

EOF
    exit 1
}

# Function to check prerequisites
check_prerequisites() {
    if ! command -v jq &> /dev/null; then
        echo -e "${RED}Error: jq is not installed${NC}"
        echo "Install: brew install jq  (or apt-get install jq)"
        exit 1
    fi

    if ! command -v bc &> /dev/null; then
        echo -e "${RED}Error: bc is not installed${NC}"
        echo "Install: brew install bc  (or apt-get install bc)"
        exit 1
    fi
}

# Function to validate input file
validate_input() {
    local file="$1"

    if [[ ! -f "$file" ]]; then
        echo -e "${RED}Error: File not found: $file${NC}"
        exit 1
    fi

    if ! jq empty "$file" 2>/dev/null; then
        echo -e "${RED}Error: Invalid JSON file${NC}"
        exit 1
    fi
}

# Function to classify deployment frequency
classify_df() {
    local df="$1"

    if (( $(echo "$df >= 1" | bc -l) )); then
        echo "ELITE"
    elif (( $(echo "$df >= 0.14" | bc -l) )); then  # ~1/week
        echo "HIGH"
    elif (( $(echo "$df >= 0.03" | bc -l) )); then  # ~1/month
        echo "MEDIUM"
    else
        echo "LOW"
    fi
}

# Function to classify lead time
classify_lt() {
    local lt_hours="$1"

    if (( $(echo "$lt_hours < 1" | bc -l) )); then
        echo "ELITE"
    elif (( $(echo "$lt_hours < 168" | bc -l) )); then  # <1 week
        echo "HIGH"
    elif (( $(echo "$lt_hours < 730" | bc -l) )); then  # <1 month
        echo "MEDIUM"
    else
        echo "LOW"
    fi
}

# Function to classify change failure rate
classify_cfr() {
    local cfr="$1"

    if (( $(echo "$cfr <= 15" | bc -l) )); then
        echo "ELITE"
    elif (( $(echo "$cfr <= 30" | bc -l) )); then
        echo "HIGH"
    elif (( $(echo "$cfr <= 45" | bc -l) )); then
        echo "MEDIUM"
    else
        echo "LOW"
    fi
}

# Function to classify MTTR
classify_mttr() {
    local mttr_hours="$1"

    if (( $(echo "$mttr_hours < 1" | bc -l) )); then
        echo "ELITE"
    elif (( $(echo "$mttr_hours < 24" | bc -l) )); then
        echo "HIGH"
    elif (( $(echo "$mttr_hours < 168" | bc -l) )); then  # <1 week
        echo "MEDIUM"
    else
        echo "LOW"
    fi
}

# Function to calculate overall performance
calculate_overall() {
    local df_class="$1"
    local lt_class="$2"
    local cfr_class="$3"
    local mttr_class="$4"

    local elite=0 high=0 medium=0 low=0

    for class in "$df_class" "$lt_class" "$cfr_class" "$mttr_class"; do
        case $class in
            ELITE) ((elite++)) ;;
            HIGH) ((high++)) ;;
            MEDIUM) ((medium++)) ;;
            LOW) ((low++)) ;;
        esac
    done

    if (( elite >= 3 )); then
        echo "ELITE"
    elif (( elite + high >= 3 )); then
        echo "HIGH"
    elif (( low <= 1 )); then
        echo "MEDIUM"
    else
        echo "LOW"
    fi
}

# Function to analyze deployment data
analyze_data() {
    local data_file="$1"

    echo -e "${BLUE}Analyzing DORA metrics...${NC}"
    echo ""

    # Extract summary data
    local total_deployments=$(jq -r '.summary.total_deployments // 0' "$data_file")
    local failed_deployments=$(jq -r '.summary.failed_deployments // 0' "$data_file")
    local total_days=$(jq -r '.metadata.total_days // 1' "$data_file")
    local median_lead_time_hours=$(jq -r '.summary.median_lead_time_hours // 0' "$data_file")
    local median_mttr_minutes=$(jq -r '.summary.median_mttr_minutes // 0' "$data_file")

    # Calculate metrics
    local df=$(echo "scale=2; $total_deployments / $total_days" | bc)
    local cfr=0
    if (( total_deployments > 0 )); then
        cfr=$(echo "scale=1; ($failed_deployments / $total_deployments) * 100" | bc)
    fi
    local mttr_hours=$(echo "scale=2; $median_mttr_minutes / 60" | bc)

    # Classify performance
    local df_class=$(classify_df "$df")
    local lt_class=$(classify_lt "$median_lead_time_hours")
    local cfr_class=$(classify_cfr "$cfr")
    local mttr_class=$(classify_mttr "$mttr_hours")
    local overall=$(calculate_overall "$df_class" "$lt_class" "$cfr_class" "$mttr_class")

    # Store for export
    METRICS_DF="$df"
    METRICS_DF_CLASS="$df_class"
    METRICS_LT="$median_lead_time_hours"
    METRICS_LT_CLASS="$lt_class"
    METRICS_CFR="$cfr"
    METRICS_CFR_CLASS="$cfr_class"
    METRICS_MTTR="$mttr_hours"
    METRICS_MTTR_CLASS="$mttr_class"
    METRICS_OVERALL="$overall"

    # Display results
    echo "========================================="
    echo "  DORA METRICS ANALYSIS"
    echo "========================================="
    echo ""

    echo -e "${CYAN}DATA SUMMARY${NC}"
    local team=$(jq -r '.team // "Unknown"' "$data_file")
    local period=$(jq -r '.period // "Unknown"' "$data_file")
    echo "  Team: $team"
    echo "  Period: $period"
    echo "  Total Deployments: $total_deployments"
    echo "  Analysis Period: $total_days days"
    echo ""

    echo -e "${CYAN}DORA METRICS${NC}"
    echo ""

    echo "Deployment Frequency:"
    echo "  Value: $df deploys/day"
    echo "  Classification: $(get_colored_class "$df_class")"
    echo ""

    echo "Lead Time for Changes:"
    echo "  Value: $median_lead_time_hours hours (median)"
    echo "  Classification: $(get_colored_class "$lt_class")"
    echo ""

    echo "Change Failure Rate:"
    echo "  Value: $cfr%"
    echo "  Failures: $failed_deployments of $total_deployments"
    echo "  Classification: $(get_colored_class "$cfr_class")"
    echo ""

    echo "Time to Restore Service:"
    echo "  Value: $median_mttr_minutes minutes (median)"
    echo "  In Hours: $mttr_hours hours"
    echo "  Classification: $(get_colored_class "$mttr_class")"
    echo ""

    echo -e "${CYAN}OVERALL PERFORMANCE${NC}"
    echo "  Level: $(get_colored_class "$overall")"
    echo ""

    echo -e "${CYAN}BENCHMARKS${NC}"
    echo "  ELITE:  Multiple deploys/day, <1h lead time, <15% CFR, <1h MTTR"
    echo "  HIGH:   Daily to weekly, 1d-1w lead time, 16-30% CFR, <1d MTTR"
    echo "  MEDIUM: Weekly to monthly, 1w-1m lead time, 31-45% CFR, 1d-1w MTTR"
    echo "  LOW:    Monthly+ deploys, >1m lead time, >45% CFR, >1w MTTR"
    echo ""
}

# Function to get colored classification
get_colored_class() {
    local class="$1"
    case $class in
        ELITE)
            echo -e "${GREEN}ELITE${NC}"
            ;;
        HIGH)
            echo -e "${BLUE}HIGH${NC}"
            ;;
        MEDIUM)
            echo -e "${YELLOW}MEDIUM${NC}"
            ;;
        LOW)
            echo -e "${RED}LOW${NC}"
            ;;
    esac
}

# Function to export to JSON
export_json() {
    local output="$1"

    cat > "$output" << EOF
{
  "deployment_frequency": {
    "value": $METRICS_DF,
    "unit": "deploys_per_day",
    "classification": "$METRICS_DF_CLASS"
  },
  "lead_time_for_changes": {
    "value": $METRICS_LT,
    "unit": "hours",
    "classification": "$METRICS_LT_CLASS"
  },
  "change_failure_rate": {
    "value": $METRICS_CFR,
    "unit": "percent",
    "classification": "$METRICS_CFR_CLASS"
  },
  "time_to_restore_service": {
    "value": $METRICS_MTTR,
    "unit": "hours",
    "classification": "$METRICS_MTTR_CLASS"
  },
  "overall_performance": "$METRICS_OVERALL"
}
EOF

    echo -e "${GREEN}✓ Metrics exported to: $output${NC}"
}

# Function to export to CSV
export_csv() {
    local output="$1"

    cat > "$output" << EOF
Metric,Value,Unit,Classification
Deployment Frequency,$METRICS_DF,deploys/day,$METRICS_DF_CLASS
Lead Time for Changes,$METRICS_LT,hours,$METRICS_LT_CLASS
Change Failure Rate,$METRICS_CFR,percent,$METRICS_CFR_CLASS
Time to Restore Service,$METRICS_MTTR,hours,$METRICS_MTTR_CLASS
Overall Performance,N/A,N/A,$METRICS_OVERALL
EOF

    echo -e "${GREEN}✓ Metrics exported to: $output${NC}"
}

# Function to export to Markdown
export_markdown() {
    local output="$1"

    cat > "$output" << EOF
# DORA Metrics Analysis

## Summary

Overall Performance: **$METRICS_OVERALL**

## Metrics

### Deployment Frequency
- **Value**: $METRICS_DF deploys/day
- **Classification**: $METRICS_DF_CLASS

### Lead Time for Changes
- **Value**: $METRICS_LT hours
- **Classification**: $METRICS_LT_CLASS

### Change Failure Rate
- **Value**: $METRICS_CFR%
- **Classification**: $METRICS_CFR_CLASS

### Time to Restore Service
- **Value**: $METRICS_MTTR hours
- **Classification**: $METRICS_MTTR_CLASS

## Classifications

| Metric | Value | Classification |
|--------|-------|----------------|
| Deployment Frequency | $METRICS_DF deploys/day | $METRICS_DF_CLASS |
| Lead Time for Changes | $METRICS_LT hours | $METRICS_LT_CLASS |
| Change Failure Rate | $METRICS_CFR% | $METRICS_CFR_CLASS |
| Time to Restore Service | $METRICS_MTTR hours | $METRICS_MTTR_CLASS |

EOF

    echo -e "${GREEN}✓ Metrics exported to: $output${NC}"
}

#############################################################################
# Claude AI Functions (only used when --claude flag is set)
#############################################################################

analyze_with_claude() {
    local data_file="$1"
    local model="claude-sonnet-4-20250514"

    if [[ -z "$ANTHROPIC_API_KEY" ]]; then
        echo -e "${RED}Error: ANTHROPIC_API_KEY is required for Claude analysis${NC}"
        echo "Set it with: export ANTHROPIC_API_KEY=your-key"
        echo "Or use: --api-key your-key"
        exit 1
    fi

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
2. Metric Analysis
3. Strengths
4. Improvement Opportunities
5. Roadmap

Be specific, actionable, and data-driven."

    local response=$(curl -s https://api.anthropic.com/v1/messages \
        -H "content-type: application/json" \
        -H "x-api-key: $ANTHROPIC_API_KEY" \
        -H "anthropic-version: 2023-06-01" \
        -d "{
            \"model\": \"$model\",
            \"max_tokens\": 4096,
            \"messages\": [{
                \"role\": \"user\",
                \"content\": $(echo "$prompt" | jq -Rs .)
            }]
        }")

    # Record API usage for cost tracking
    if command -v record_api_usage &> /dev/null; then
        record_api_usage "$response" "$model" > /dev/null
    fi

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

# Parse command line arguments
DATA_FILE=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -f|--format)
            OUTPUT_FORMAT="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_FILE="$2"
            shift 2
            ;;
        --claude)
            USE_CLAUDE=true
            shift
            ;;
        --compare)
            COMPARE_MODE=true
            shift
            ;;
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

# Validate arguments
if [[ -z "$DATA_FILE" ]]; then
    echo -e "${RED}Error: No data file specified${NC}"
    usage
fi

# Check compare mode requirements
if [[ "$COMPARE_MODE" == "true" ]]; then
    if [[ -z "$ANTHROPIC_API_KEY" ]]; then
        echo -e "${RED}Error: --compare mode requires ANTHROPIC_API_KEY${NC}"
        echo "Set environment variable or use -k flag"
        exit 1
    fi
    USE_CLAUDE=false  # Start with basic
fi

# Load cost tracking library if using Claude or compare mode
if [[ "$USE_CLAUDE" == "true" ]] || [[ "$COMPARE_MODE" == "true" ]]; then
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
    if [ -f "$REPO_ROOT/lib/claude-cost.sh" ]; then
        source "$REPO_ROOT/lib/claude-cost.sh"
        init_cost_tracking
    fi
fi

# Main
echo ""
echo "========================================="
if [[ "$COMPARE_MODE" == "true" ]]; then
    echo "  DORA Metrics Analyser (Comparison Mode)"
elif [[ "$USE_CLAUDE" == "true" ]]; then
    echo "  DORA Metrics Analyser (Claude AI Mode)"
else
    echo "  DORA Metrics Analyser"
fi
echo "========================================="
echo ""

check_prerequisites
validate_input "$DATA_FILE"

if [[ "$COMPARE_MODE" == "true" ]]; then
    # Comparison mode: run both basic and Claude
    echo -e "${BLUE}Running basic DORA analysis...${NC}"
    echo ""
    USE_CLAUDE=false
    analyze_data "$DATA_FILE"

    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo -e "${BLUE}Running Claude AI analysis...${NC}"
    echo ""
    USE_CLAUDE=true
    analyze_with_claude "$DATA_FILE"

    # Display comparison summary
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${CYAN}Comparison Summary${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "Basic analyser provides:"
    echo "  • Four key DORA metrics calculation"
    echo "  • Performance level classification"
    echo "  • Benchmark comparison (Elite/High/Medium/Low)"
    echo "  • Raw metrics and percentiles"
    echo ""
    echo "Claude-enhanced analyser adds:"
    echo "  • ${GREEN}Contextual insights${NC} and trend interpretation"
    echo "  • ${GREEN}Pattern recognition${NC} across metrics"
    echo "  • ${GREEN}Root cause analysis${NC} for performance issues"
    echo "  • ${GREEN}Specific improvement recommendations${NC}"
    echo "  • ${GREEN}Industry best practices${NC} application"
    echo "  • ${GREEN}Prioritized action items${NC}"
    echo ""

    if command -v display_api_cost_summary &> /dev/null; then
        echo "API Cost:"
        display_api_cost_summary
    fi

    echo ""
    echo "Use basic for: CI/CD integration, dashboards, trending"
    echo "Use Claude for: Strategic planning, performance reviews, root cause analysis"
    echo ""

elif [[ "$USE_CLAUDE" == "true" ]]; then
    # Claude AI analysis mode
    analyze_with_claude "$DATA_FILE"

    # Display cost summary
    if command -v display_api_cost_summary &> /dev/null; then
        display_api_cost_summary
    fi
else
    # Standard analysis mode
    analyze_data "$DATA_FILE"
fi

# Export if requested
if [[ -n "$OUTPUT_FILE" ]]; then
    case $OUTPUT_FORMAT in
        json)
            export_json "$OUTPUT_FILE"
            ;;
        csv)
            export_csv "$OUTPUT_FILE"
            ;;
        markdown)
            export_markdown "$OUTPUT_FILE"
            ;;
    esac
fi

echo "========================================="
echo -e "${GREEN}  Analysis Complete${NC}"
echo "========================================="
echo ""
