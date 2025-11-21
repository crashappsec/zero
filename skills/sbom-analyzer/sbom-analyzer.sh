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
