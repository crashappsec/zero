#!/bin/bash
# Copyright (c) 2024 Crash Override Inc
# 101 Fulton St, 416, New York 10038
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Chalk Build Analyzer Script
# Analyzes Chalk build reports for performance metrics and bottlenecks
# Usage: ./chalk-build-analyzer.sh [options] <chalk-report.json>
#############################################################################

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Default options
OUTPUT_FORMAT="text"
OUTPUT_FILE=""
COMPARE_MODE=false
BASELINE_FILE=""

# Function to print usage
usage() {
    cat << EOF
Chalk Build Analyzer - Build performance analysis

Analyzes Chalk build reports to identify performance bottlenecks,
resource utilization issues, and engineering velocity metrics.

Usage: $0 [OPTIONS] <chalk-report.json>

SINGLE BUILD ANALYSIS:
    $0 build-report.json

COMPARISON MODE:
    $0 --compare baseline.json current.json

OPTIONS:
    -c, --compare BASELINE  Compare current build against baseline
    -f, --format FORMAT     Output format: text|json|csv (default: text)
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
echo "  Chalk Build Analyzer"
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
    compare_builds "$BASELINE_FILE" "$REPORT_FILE"
else
    validate_report "$REPORT_FILE"

    if [[ "$OUTPUT_FORMAT" == "json" ]]; then
        if [[ -z "$OUTPUT_FILE" ]]; then
            OUTPUT_FILE="analysis.json"
        fi
        export_json "$REPORT_FILE" "$OUTPUT_FILE"
    else
        analyze_build "$REPORT_FILE"
    fi
fi

echo "========================================="
echo -e "${GREEN}  Analysis Complete${NC}"
echo "========================================="
echo ""
