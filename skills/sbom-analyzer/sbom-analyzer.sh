#!/bin/bash
# Copyright (c) 2024 Crash Override Inc
# 101 Fulton St, 416, New York 10038
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# SBOM Analyzer Script
# Analyzes SBOMs and repositories for vulnerabilities using osv-scanner
# Supports taint analysis for reachability determination
# Usage: ./sbom-analyzer.sh [options] <target>
#############################################################################

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default options
TAINT_ANALYSIS=false
OUTPUT_FORMAT="table"
TEMP_DIR=""
CLEANUP=true
PRIORITIZE=false
KEV_CACHE=""

# Function to print usage
usage() {
    cat << EOF
SBOM Analyzer - Vulnerability scanning with osv-scanner

Usage: $0 [OPTIONS] <target>

TARGET:
    SBOM file path          Analyze an existing SBOM (JSON/XML)
    Git repository URL      Clone and analyze repository
    Local directory path    Analyze local repository

OPTIONS:
    -t, --taint-analysis    Enable call graph/taint analysis (Go projects)
    -p, --prioritize        Add intelligent prioritization (CISA KEV, CVSS, exploitability)
    -f, --format FORMAT     Output format: table|json|markdown|sarif (default: table)
    -o, --output FILE       Write results to file
    -k, --keep-clone        Keep cloned repository (don't cleanup)
    -h, --help              Show this help message

EXAMPLES:
    # Analyze an SBOM file
    $0 /path/to/sbom.json

    # Analyze a Git repository with taint analysis
    $0 --taint-analysis https://github.com/org/repo

    # Analyze local directory with JSON output
    $0 --format json ./my-project

    # Analyze and save results
    $0 --output results.json --format json /path/to/sbom.cdx.xml

EOF
    exit 1
}

# Function to check if osv-scanner is installed
check_osv_scanner() {
    if ! command -v osv-scanner &> /dev/null; then
        echo -e "${RED}Error: osv-scanner is not installed${NC}"
        echo ""
        echo "Install osv-scanner:"
        echo "  go install github.com/google/osv-scanner/cmd/osv-scanner@latest"
        echo ""
        echo "Or with Homebrew:"
        echo "  brew install osv-scanner"
        echo ""
        exit 1
    fi
}

# Function to check if jq is installed
check_jq() {
    if ! command -v jq &> /dev/null; then
        echo -e "${RED}Error: jq is not installed (required for prioritization)${NC}"
        echo ""
        echo "Install jq:"
        echo "  brew install jq  (macOS)"
        echo "  apt-get install jq  (Debian/Ubuntu)"
        echo ""
        exit 1
    fi
}

# Function to fetch CISA KEV catalog
fetch_kev_catalog() {
    KEV_CACHE=$(mktemp)
    echo -e "${BLUE}Fetching CISA KEV catalog...${NC}" >&2

    if curl -s "https://www.cisa.gov/sites/default/files/feeds/known_exploited_vulnerabilities.json" -o "$KEV_CACHE"; then
        echo -e "${GREEN}✓ KEV catalog downloaded${NC}" >&2
        return 0
    else
        echo -e "${YELLOW}⚠ Failed to fetch KEV catalog${NC}" >&2
        return 1
    fi
}

# Function to check if CVE is in CISA KEV
is_in_kev() {
    local cve_id="$1"
    if [[ -z "$KEV_CACHE" ]] || [[ ! -f "$KEV_CACHE" ]]; then
        return 1
    fi
    jq -e ".vulnerabilities[] | select(.cveID == \"$cve_id\")" "$KEV_CACHE" > /dev/null 2>&1
}

# Function to prioritize vulnerabilities
prioritize_results() {
    local scan_output="$1"
    local prioritized_output=$(mktemp)

    echo ""
    echo -e "${BLUE}Analyzing and prioritizing vulnerabilities...${NC}"
    echo ""

    # Parse vulnerabilities and add priority scores
    jq -r '.results[] | select(.vulnerabilities) | .vulnerabilities[] |
        {
            id: (.id // "N/A"),
            package: (.package.name // "N/A"),
            version: (.package.version // "N/A"),
            ecosystem: (.package.ecosystem // "N/A"),
            cvss: ((.database_specific.cvss // .database_specific.severity // "0") | tostring),
            summary: (.summary // .details // "No description available")
        } | @json' "$scan_output" 2>/dev/null | while IFS= read -r vuln_json; do

        # Extract fields
        local vuln_id=$(echo "$vuln_json" | jq -r '.id')
        local package=$(echo "$vuln_json" | jq -r '.package')
        local version=$(echo "$vuln_json" | jq -r '.version')
        local cvss=$(echo "$vuln_json" | jq -r '.cvss' | grep -oE '[0-9]+(\.[0-9]+)?' | head -1 || echo "0")
        local summary=$(echo "$vuln_json" | jq -r '.summary' | head -c 100)

        # Calculate priority score
        local priority_score=0
        local priority_label="LOW"
        local flags=""

        # CISA KEV check (highest priority)
        if is_in_kev "$vuln_id"; then
            priority_score=$((priority_score + 100))
            flags="${flags}[KEV] "
        fi

        # CVSS score
        if [[ -n "$cvss" ]] && [[ "$cvss" != "0" ]]; then
            # CVSS to priority: 9-10=Critical(50), 7-8.9=High(30), 4-6.9=Medium(15), 0-3.9=Low(5)
            if (( $(echo "$cvss >= 9.0" | bc -l 2>/dev/null || echo 0) )); then
                priority_score=$((priority_score + 50))
            elif (( $(echo "$cvss >= 7.0" | bc -l 2>/dev/null || echo 0) )); then
                priority_score=$((priority_score + 30))
            elif (( $(echo "$cvss >= 4.0" | bc -l 2>/dev/null || echo 0) )); then
                priority_score=$((priority_score + 15))
            else
                priority_score=$((priority_score + 5))
            fi
        fi

        # Determine priority label
        if [[ $priority_score -ge 100 ]]; then
            priority_label="CRITICAL"
        elif [[ $priority_score -ge 50 ]]; then
            priority_label="HIGH"
        elif [[ $priority_score -ge 30 ]]; then
            priority_label="MEDIUM"
        fi

        # Store with priority
        echo "$priority_score|$priority_label|$flags|$vuln_id|$package|$version|$cvss|$summary" >> "$prioritized_output"
    done

    # Sort by priority score (descending) and display
    if [[ -f "$prioritized_output" ]] && [[ -s "$prioritized_output" ]]; then
        echo -e "${GREEN}Prioritized Vulnerabilities:${NC}"
        echo ""
        echo "Priority | Flags | CVE ID | Package | Version | CVSS | Description"
        echo "---------|-------|--------|---------|---------|------|-------------"

        sort -t'|' -k1 -rn "$prioritized_output" | while IFS='|' read -r score label flags vuln_id package version cvss summary; do
            # Color based on priority
            local color="$NC"
            case "$label" in
                CRITICAL) color="$RED" ;;
                HIGH) color="$YELLOW" ;;
                MEDIUM) color="$BLUE" ;;
            esac

            printf "${color}%-8s${NC} | %-5s | %-15s | %-20s | %-10s | %-4s | %-50s\n" \
                "$label" "$flags" "$vuln_id" "${package:0:20}" "${version:0:10}" "$cvss" "${summary:0:50}"
        done

        echo ""

        # Summary statistics
        local total=$(wc -l < "$prioritized_output")
        local critical=$(grep -c "^[0-9]*|CRITICAL|" "$prioritized_output" || echo 0)
        local high=$(grep -c "^[0-9]*|HIGH|" "$prioritized_output" || echo 0)
        local medium=$(grep -c "^[0-9]*|MEDIUM|" "$prioritized_output" || echo 0)
        local low=$(grep -c "^[0-9]*|LOW|" "$prioritized_output" || echo 0)
        local kev=$(grep -c "\[KEV\]" "$prioritized_output" || echo 0)

        echo -e "${GREEN}Summary:${NC}"
        echo "  Total vulnerabilities: $total"
        echo "  Critical: $critical"
        echo "  High: $high"
        echo "  Medium: $medium"
        echo "  Low: $low"
        echo "  In CISA KEV: $kev"

        rm -f "$prioritized_output"
    else
        echo -e "${GREEN}No vulnerabilities found${NC}"
    fi
}

# Function to detect if target is a Git URL
is_git_url() {
    [[ "$1" =~ ^(https?|git)://.*\.git$ ]] || [[ "$1" =~ ^git@.*:.*\.git$ ]] || [[ "$1" =~ github\.com|gitlab\.com|bitbucket\.org ]]
}

# Function to detect if target is an SBOM file
is_sbom_file() {
    local file="$1"
    [[ -f "$file" ]] && ([[ "$file" =~ \.json$ ]] || [[ "$file" =~ \.xml$ ]] || [[ "$file" =~ \.cdx\. ]] || [[ "$file" =~ bom\. ]])
}

# Function to clone repository
clone_repository() {
    local repo_url="$1"
    TEMP_DIR=$(mktemp -d)

    echo -e "${BLUE}Cloning repository...${NC}"
    if git clone --depth 1 "$repo_url" "$TEMP_DIR"; then
        echo -e "${GREEN}✓ Repository cloned to: $TEMP_DIR${NC}"
        echo ""
        return 0
    else
        echo -e "${RED}✗ Failed to clone repository${NC}"
        echo -e "${YELLOW}Note: For private repositories, ensure you have proper SSH keys or authentication${NC}"
        return 1
    fi
}

# Function to run osv-scanner on SBOM
scan_sbom() {
    local sbom_file="$1"
    local output_file="$2"

    echo -e "${BLUE}Scanning SBOM: $sbom_file${NC}"
    echo ""

    # If prioritization is enabled, force JSON output to temp file
    if [[ "$PRIORITIZE" == true ]]; then
        local temp_json=$(mktemp)
        if osv-scanner -L "$sbom_file" --format=json > "$temp_json" 2>&1; then
            prioritize_results "$temp_json"
        else
            cat "$temp_json"
        fi
        rm -f "$temp_json"
        return
    fi

    # Normal osv-scanner output
    local cmd="osv-scanner -L $sbom_file"

    if [[ "$OUTPUT_FORMAT" != "table" ]]; then
        cmd="$cmd --format=$OUTPUT_FORMAT"
    fi

    if [[ -n "$output_file" ]]; then
        cmd="$cmd --output=$output_file"
    fi

    eval "$cmd" || true
}

# Function to run osv-scanner on repository
scan_repository() {
    local repo_path="$1"
    local output_file="$2"

    echo -e "${BLUE}Scanning repository: $repo_path${NC}"
    echo ""

    # If prioritization is enabled, force JSON output to temp file
    if [[ "$PRIORITIZE" == true ]]; then
        local temp_json=$(mktemp)
        local scan_cmd="osv-scanner --recursive $repo_path"

        if [[ "$TAINT_ANALYSIS" == true ]]; then
            echo -e "${YELLOW}Enabling call graph/taint analysis...${NC}"
            scan_cmd="osv-scanner --call-analysis=all $repo_path"
        fi

        if eval "$scan_cmd --format=json" > "$temp_json" 2>&1; then
            prioritize_results "$temp_json"
        else
            cat "$temp_json"
        fi
        rm -f "$temp_json"
        return
    fi

    # Normal osv-scanner output
    local cmd="osv-scanner --recursive $repo_path"

    if [[ "$TAINT_ANALYSIS" == true ]]; then
        echo -e "${YELLOW}Enabling call graph/taint analysis...${NC}"
        cmd="osv-scanner --call-analysis=all $repo_path"
    fi

    if [[ "$OUTPUT_FORMAT" != "table" ]]; then
        cmd="$cmd --format=$OUTPUT_FORMAT"
    fi

    if [[ -n "$output_file" ]]; then
        cmd="$cmd --output=$output_file"
    fi

    eval "$cmd" || true
}

# Function to cleanup temporary files
cleanup() {
    if [[ "$CLEANUP" == true ]] && [[ -n "$TEMP_DIR" ]] && [[ -d "$TEMP_DIR" ]]; then
        echo ""
        echo -e "${YELLOW}Cleaning up temporary files...${NC}"
        rm -rf "$TEMP_DIR"
        echo -e "${GREEN}✓ Cleanup complete${NC}"
    fi

    # Clean up KEV cache
    if [[ -n "$KEV_CACHE" ]] && [[ -f "$KEV_CACHE" ]]; then
        rm -f "$KEV_CACHE"
    fi
}

# Parse command line arguments
OUTPUT_FILE=""
TARGET=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--taint-analysis)
            TAINT_ANALYSIS=true
            shift
            ;;
        -p|--prioritize)
            PRIORITIZE=true
            shift
            ;;
        -f|--format)
            OUTPUT_FORMAT="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_FILE="$2"
            shift 2
            ;;
        -k|--keep-clone)
            CLEANUP=false
            shift
            ;;
        -h|--help)
            usage
            ;;
        *)
            TARGET="$1"
            shift
            ;;
    esac
done

# Validate target
if [[ -z "$TARGET" ]]; then
    echo -e "${RED}Error: No target specified${NC}"
    echo ""
    usage
fi

# Main script
echo ""
echo "========================================="
echo "  SBOM Analyzer (osv-scanner)"
echo "========================================="
echo ""

# Check prerequisites
check_osv_scanner

if [[ "$PRIORITIZE" == true ]]; then
    check_jq
    fetch_kev_catalog || echo -e "${YELLOW}Continuing without KEV data${NC}"
fi

# Determine target type and scan
if is_sbom_file "$TARGET"; then
    echo -e "${GREEN}Target type: SBOM file${NC}"
    scan_sbom "$TARGET" "$OUTPUT_FILE"
elif is_git_url "$TARGET"; then
    echo -e "${GREEN}Target type: Git repository${NC}"
    if clone_repository "$TARGET"; then
        scan_repository "$TEMP_DIR" "$OUTPUT_FILE"
        cleanup
    fi
elif [[ -d "$TARGET" ]]; then
    echo -e "${GREEN}Target type: Local directory${NC}"
    scan_repository "$TARGET" "$OUTPUT_FILE"
else
    echo -e "${RED}Error: Invalid target${NC}"
    echo "Target must be:"
    echo "  - Path to SBOM file (.json, .xml, .cdx.*)"
    echo "  - Git repository URL"
    echo "  - Local directory path"
    exit 1
fi

echo ""
echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}  Analysis Complete${NC}"
echo -e "${GREEN}=========================================${NC}"

if [[ -n "$OUTPUT_FILE" ]]; then
    echo ""
    echo -e "Results saved to: ${BLUE}$OUTPUT_FILE${NC}"
fi

echo ""
