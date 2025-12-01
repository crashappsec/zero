#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# 
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Chalk Build Analyser Script
# Analyzes Chalk build reports for performance metrics and bottlenecks
# Usage: ./chalk-build-analyser.sh [options] <chalk-report.json>
#############################################################################

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SCANNERS_ROOT="$(dirname "$SCRIPT_DIR")"
UTILS_ROOT="$(dirname "$SCANNERS_ROOT")"
REPO_ROOT="$(dirname "$UTILS_ROOT")"

# Load .env file if it exists in repository root
if [[ -f "$REPO_ROOT/.env" ]]; then
    set -a  # automatically export all variables
    source "$REPO_ROOT/.env"
    set +a  # stop automatically exporting
fi

# Default options
OUTPUT_FORMAT="markdown"
OUTPUT_FILE=""
COMPARE_MODE=false
BASELINE_FILE=""
USE_CLAUDE=false
ANTHROPIC_API_KEY="${ANTHROPIC_API_KEY:-}"

# Load cost tracking if using Claude
if [[ "$USE_CLAUDE" == "true" ]]; then
    REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
    if [ -f "$REPO_ROOT/lib/claude-cost.sh" ]; then
        source "$REPO_ROOT/lib/claude-cost.sh"
        init_cost_tracking
    fi
fi

# Function to print usage
usage() {
    cat << EOF
Chalk Build Analyser - Build performance analysis

Analyzes Chalk build reports to identify performance bottlenecks,
resource utilization issues, and engineering velocity metrics.

Usage: $0 [OPTIONS] <chalk-report.json>

SINGLE BUILD ANALYSIS:
    $0 build-report.json

COMPARISON MODE:
    $0 --compare baseline.json current.json

OPTIONS:
    -c, --compare BASELINE  Compare current build against baseline
    --claude                Use Claude AI for enhanced analysis (requires ANTHROPIC_API_KEY)
    -f, --format FORMAT     Output format: text|json|csv|markdown (default: markdown)
    -o, --output FILE       Write results to file
    -h, --help              Show this help message

EXAMPLES:
    # Analyze a single build
    $0 my-build.json

    # Compare two builds
    $0 --compare before.json after.json

    # Export to JSON
    $0 --format json --output analysis.json build.json

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

    # Check if it looks like a Chalk report
    if ! jq -e '.chalk_version or ._CHALK_VERSION or .build_id or .duration' "$report" &>/dev/null; then
        echo -e "${YELLOW}Warning: This may not be a Chalk build report${NC}"
    fi
}

# Function to extract build metrics
extract_metrics() {
    local report="$1"

    # Extract key metrics from Chalk report
    local build_id=$(jq -r '.build_id // .BUILD_ID // "unknown"' "$report")
    local duration=$(jq -r '.duration // .DURATION // 0' "$report")
    local status=$(jq -r '.status // .STATUS // "unknown"' "$report")
    local platform=$(jq -r '.platform // .PLATFORM // "unknown"' "$report")
    local branch=$(jq -r '.branch // .git_branch // .BRANCH // "unknown"' "$report")

    echo "build_id=$build_id"
    echo "duration=$duration"
    echo "status=$status"
    echo "platform=$platform"
    echo "branch=$branch"
}

# Function to analyze build performance
analyze_build() {
    local report="$1"

    echo -e "${BLUE}Analyzing build performance...${NC}"
    echo ""

    # Extract basic metrics
    local metrics=$(extract_metrics "$report")
    eval "$metrics"

    echo "========================================="
    echo "  CHALK BUILD ANALYSIS"
    echo "========================================="
    echo ""

    echo -e "${CYAN}BUILD SUMMARY${NC}"
    echo "  Build ID: $build_id"
    echo "  Status: $status"
    echo "  Duration: ${duration}s"
    echo "  Platform: $platform"
    echo "  Branch: $branch"
    echo ""

    # Analyze stage breakdown if available
    if jq -e '.stages' "$report" &>/dev/null; then
        echo -e "${CYAN}STAGE BREAKDOWN${NC}"
        jq -r '.stages | to_entries[] | "  \(.key): \(.value.duration)s"' "$report"
        echo ""
    fi

    # Analyze cache effectiveness
    if jq -e '.cache_hit_rate' "$report" &>/dev/null; then
        local cache_rate=$(jq -r '.cache_hit_rate' "$report")
        echo -e "${CYAN}CACHE EFFECTIVENESS${NC}"
        echo "  Cache Hit Rate: ${cache_rate}%"
        echo ""
    fi

    # Resource utilization
    if jq -e '.resources' "$report" &>/dev/null; then
        echo -e "${CYAN}RESOURCE UTILIZATION${NC}"
        jq -r '.resources | to_entries[] | "  \(.key): \(.value)"' "$report"
        echo ""
    fi

    # Basic recommendations
    echo -e "${CYAN}ANALYSIS${NC}"

    # Duration assessment
    if (( $(echo "$duration > 600" | bc -l) )); then
        echo "  ⚠️  Build duration is HIGH (>10 minutes)"
        echo "     Consider parallelization or caching improvements"
    elif (( $(echo "$duration > 300" | bc -l) )); then
        echo "  ⚡ Build duration is MODERATE (5-10 minutes)"
        echo "     Room for optimization"
    else
        echo "  ✓  Build duration is GOOD (<5 minutes)"
    fi

    # Cache assessment
    if jq -e '.cache_hit_rate' "$report" &>/dev/null; then
        local cache_rate=$(jq -r '.cache_hit_rate' "$report")
        if (( $(echo "$cache_rate < 70" | bc -l) )); then
            echo "  ⚠️  Cache hit rate is LOW (<70%)"
            echo "     Review dependency management and cache configuration"
        fi
    fi

    echo ""
}

# Function to compare two builds
compare_builds() {
    local baseline="$1"
    local current="$2"

    echo -e "${BLUE}Comparing builds...${NC}"
    echo ""

    # Extract metrics from both
    local baseline_metrics=$(extract_metrics "$baseline")
    local current_metrics=$(extract_metrics "$current")

    # Parse baseline
    eval "baseline_$baseline_metrics"
    # Parse current
    eval "current_$current_metrics"

    echo "========================================="
    echo "  BUILD COMPARISON"
    echo "========================================="
    echo ""

    echo -e "${CYAN}BASELINE BUILD${NC}"
    echo "  Build ID: $baseline_build_id"
    echo "  Duration: ${baseline_duration}s"
    echo ""

    echo -e "${CYAN}CURRENT BUILD${NC}"
    echo "  Build ID: $current_build_id"
    echo "  Duration: ${current_duration}s"
    echo ""

    # Calculate difference
    local diff=$(echo "$current_duration - $baseline_duration" | bc)
    local pct_change=$(echo "scale=1; ($diff / $baseline_duration) * 100" | bc)

    echo -e "${CYAN}PERFORMANCE CHANGE${NC}"
    if (( $(echo "$diff > 0" | bc -l) )); then
        echo -e "  Duration: ${RED}+${diff}s (+${pct_change}%)${NC}"
        echo "  ⚠️  Build got SLOWER"
    elif (( $(echo "$diff < 0" | bc -l) )); then
        local abs_diff=$(echo "$diff * -1" | bc)
        local abs_pct=$(echo "$pct_change * -1" | bc)
        echo -e "  Duration: ${GREEN}-${abs_diff}s (-${abs_pct}%)${NC}"
        echo "  ✓  Build got FASTER"
    else
        echo "  Duration: No change"
    fi

    echo ""

    # Regression detection
    if (( $(echo "$pct_change > 10" | bc -l) )); then
        echo -e "${RED}REGRESSION DETECTED${NC}"
        echo "  Severity: HIGH"
        echo "  Recommendation: Investigate recent changes"
    elif (( $(echo "$pct_change > 5" | bc -l) )); then
        echo -e "${YELLOW}POTENTIAL REGRESSION${NC}"
        echo "  Severity: MODERATE"
        echo "  Recommendation: Monitor trend"
    fi

    echo ""
}

# Function to export to JSON
export_json() {
    local report="$1"
    local output="$2"

    jq '{
        build_id: (.build_id // .BUILD_ID // "unknown"),
        duration: (.duration // .DURATION // 0),
        status: (.status // .STATUS // "unknown"),
        platform: (.platform // .PLATFORM // "unknown"),
        branch: (.branch // .git_branch // .BRANCH // "unknown"),
        stages: (.stages // {}),
        cache_hit_rate: (.cache_hit_rate // null),
        resources: (.resources // {}),
        analysis: {
            duration_category: (
                if (.duration // 0) > 600 then "HIGH"
                elif (.duration // 0) > 300 then "MODERATE"
                else "GOOD"
                end
            )
        }
    }' "$report" > "$output"

    echo -e "${GREEN}✓ Analysis exported to: $output${NC}"
}

# Function to export to Markdown
export_markdown() {
    local report="$1"
    local output="$2"

    # Extract key fields from the JSON report
    local build_id=$(jq -r '.build_id // .BUILD_ID // "unknown"' "$report")
    local duration=$(jq -r '.duration // .DURATION // 0' "$report")
    local status=$(jq -r '.status // .STATUS // "unknown"' "$report")
    local platform=$(jq -r '.platform // .PLATFORM // "unknown"' "$report")
    local branch=$(jq -r '.branch // .git_branch // .BRANCH // "unknown"' "$report")

    # Calculate duration category
    local duration_category="GOOD"
    if (( $(echo "$duration > 600" | bc -l) )); then
        duration_category="HIGH"
    elif (( $(echo "$duration > 300" | bc -l) )); then
        duration_category="MODERATE"
    fi

    cat > "$output" << EOF
# Chalk Build Analysis Report

## Build Information

- **Build ID**: $build_id
- **Status**: $status
- **Duration**: ${duration}s ($duration_category)
- **Platform**: $platform
- **Branch**: $branch

## Build Metrics

| Metric | Value |
|--------|-------|
| Build ID | $build_id |
| Status | $status |
| Duration | ${duration}s |
| Duration Category | $duration_category |
| Platform | $platform |
| Branch | $branch |

EOF

    # Add stages if present
    if jq -e '.stages' "$report" > /dev/null 2>&1; then
        echo "" >> "$output"
        echo "## Build Stages" >> "$output"
        echo "" >> "$output"
        jq -r '.stages | to_entries[] | "### \(.key)\n- Duration: \(.value.duration // "N/A")s\n- Status: \(.value.status // "N/A")\n"' "$report" >> "$output"
    fi

    echo -e "${GREEN}✓ Analysis exported to: $output${NC}"
}

#############################################################################
# Claude AI Analysis
#############################################################################

analyze_with_claude() {
    local data="$1"
    local model="claude-sonnet-4-20250514"

    if [[ -z "$ANTHROPIC_API_KEY" ]]; then
        echo -e "${RED}Error: ANTHROPIC_API_KEY required for --claude mode${NC}" >&2
        exit 1
    fi

    echo -e "${BLUE}Analyzing with Claude AI...${NC}" >&2

    local prompt="Analyze this Chalk build report data and provide insights on build performance and optimization. Focus on:
1. Build duration analysis and performance categorization
2. Stage-by-stage bottleneck identification
3. Cache effectiveness and optimization opportunities
4. Resource utilization patterns and inefficiencies
5. Trend analysis and regression detection
6. Specific recommendations for:
   - Parallelization opportunities
   - Caching improvements
   - Dependency optimization
   - Build pipeline hardening
7. Prioritized action items for reducing build time

Data:
$data"

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

    if command -v record_api_usage &> /dev/null; then
        record_api_usage "$response" "$model" > /dev/null
    fi

    echo "$response" | jq -r '.content[0].text // empty'
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
echo "  Chalk Build Analyser"
echo "========================================="
echo ""

check_prerequisites

# Capture output for Claude analysis if enabled
if [[ "$USE_CLAUDE" == "true" ]]; then
    analysis_output=$(
        if [[ "$COMPARE_MODE" == true ]]; then
            if [[ -z "$BASELINE_FILE" ]]; then
                echo "Error: No baseline file specified for comparison"
                exit 1
            fi

            validate_report "$BASELINE_FILE" 2>&1
            validate_report "$REPORT_FILE" 2>&1
            compare_builds "$BASELINE_FILE" "$REPORT_FILE" 2>&1
        else
            validate_report "$REPORT_FILE" 2>&1

            if [[ "$OUTPUT_FORMAT" == "json" ]]; then
                if [[ -z "$OUTPUT_FILE" ]]; then
                    OUTPUT_FILE="analysis.json"
                fi
                export_json "$REPORT_FILE" "$OUTPUT_FILE" 2>&1
            elif [[ "$OUTPUT_FORMAT" == "markdown" ]]; then
                if [[ -z "$OUTPUT_FILE" ]]; then
                    OUTPUT_FILE="analysis.md"
                fi
                export_markdown "$REPORT_FILE" "$OUTPUT_FILE" 2>&1
            else
                analyze_build "$REPORT_FILE" 2>&1
            fi
        fi
    )

    # Display original analysis
    echo "$analysis_output"

    echo ""
    echo "========================================="
    echo "  Claude AI Enhanced Analysis"
    echo "========================================="
    echo ""

    # Get Claude analysis
    claude_analysis=$(analyze_with_claude "$analysis_output")
    echo "$claude_analysis"

    # Display cost summary
    if command -v display_api_cost_summary &> /dev/null; then
        echo ""
        display_api_cost_summary
    fi
else
    if [[ "$COMPARE_MODE" == true ]]; then
        if [[ -z "$BASELINE_FILE" ]]; then
            echo -e "${RED}Error: No baseline file specified for comparison${NC}"
            exit 1
        fi

        validate_report "$BASELINE_FILE"
        validate_report "$REPORT_FILE"
        compare_builds "$BASELINE_FILE" "$REPORT_FILE"
    else
        validate_report "$REPORT_FILE"

        if [[ "$OUTPUT_FORMAT" == "json" ]]; then
            if [[ -z "$OUTPUT_FILE" ]]; then
                OUTPUT_FILE="analysis.json"
            fi
            export_json "$REPORT_FILE" "$OUTPUT_FILE"
        elif [[ "$OUTPUT_FORMAT" == "markdown" ]]; then
            if [[ -z "$OUTPUT_FILE" ]]; then
                OUTPUT_FILE="analysis.md"
            fi
            export_markdown "$REPORT_FILE" "$OUTPUT_FILE"
        else
            analyze_build "$REPORT_FILE"
        fi
    fi
fi

echo "========================================="
echo -e "${GREEN}  Analysis Complete${NC}"
echo "========================================="
echo ""
