#!/bin/bash
# Copyright (c) 2024 Crash Override Inc
# 101 Fulton St, 416, New York 10038
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# SBOM Analyzer Script with Claude AI Integration
# Analyzes SBOMs using osv-scanner and enhances with Claude analysis
# Usage: ./sbom-analyzer-claude.sh [options] <target>
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
TAINT_ANALYSIS=false
DEEP_ANALYSIS=false
TEMP_DIR=""
CLEANUP=true
ANTHROPIC_API_KEY="${ANTHROPIC_API_KEY:-}"

# Function to print usage
usage() {
    cat << EOF
SBOM Analyzer with Claude AI - Enhanced vulnerability analysis

Usage: $0 [OPTIONS] <target>

TARGET:
    SBOM file path          Analyze an existing SBOM (JSON/XML)
    Git repository URL      Clone and analyze repository (scans lock files)
    Local directory path    Analyze local repository (scans lock files)

ANALYSIS MODES:
    (default)               Vulnerability analysis only
    -t, --taint-analysis    Add taint/reachability analysis (requires source code)
    -d, --deep              Run all available analysis (vulnerability + taint + more)

OTHER OPTIONS:
    -k, --api-key KEY       Anthropic API key (or set ANTHROPIC_API_KEY env var)
    --keep-clone            Keep cloned repository (don't cleanup)
    -h, --help              Show this help message

ENVIRONMENT:
    ANTHROPIC_API_KEY       Your Anthropic API key

DEFAULT BEHAVIOR (Vulnerability Analysis):
    1. If SBOM exists: Scans the SBOM for vulnerabilities
    2. If no SBOM: osv-scanner scans lock files directly
       - Supports: package-lock.json, go.mod, Cargo.lock, requirements.txt, etc.
    3. Reports vulnerabilities with severity, CVEs, CVSS scores
    4. Checks against CISA KEV for known exploitation

TAINT ANALYSIS (--taint-analysis):
    - Adds call graph/reachability analysis
    - Determines if vulnerable code is actually called
    - Reports: CALLED, NOT CALLED, UNKNOWN reachability
    - Currently best supported for Go projects
    - Requires source code (not just SBOM)

DEEP ANALYSIS (--deep):
    - Runs all available analysis types
    - Vulnerability scanning
    - Taint/reachability analysis
    - Maximum detail and insights

EXAMPLES:
    # Default: Vulnerability analysis only
    $0 https://github.com/org/repo

    # Add taint analysis (Go projects recommended)
    $0 --taint-analysis https://github.com/org/go-project

    # Deep analysis with everything
    $0 --deep https://github.com/org/repo

    # Analyze existing SBOM
    $0 /path/to/sbom.json

    # Analyze local directory
    $0 /path/to/project

DEPENDENCIES:
    - osv-scanner (required): Install via bootstrap.sh or:
      go install github.com/google/osv-scanner/cmd/osv-scanner@latest

EOF
    exit 1
}

# Function to check prerequisites
check_prerequisites() {
    # Check osv-scanner
    if ! command -v osv-scanner &> /dev/null; then
        echo -e "${RED}Error: osv-scanner is not installed${NC}"
        echo "Install: go install github.com/google/osv-scanner/cmd/osv-scanner@latest"
        exit 1
    fi

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
        echo -e "${GREEN}✓ Repository cloned${NC}"
        return 0
    else
        echo -e "${RED}✗ Failed to clone repository${NC}"
        echo -e "${YELLOW}Note: For private repositories, ensure you have proper SSH keys or authentication${NC}"
        return 1
    fi
}

# Function to find existing SBOM in directory
find_sbom() {
    local dir="$1"

    # Look for common SBOM file patterns
    local sbom_patterns=(
        "sbom.json"
        "bom.json"
        "*.cdx.json"
        "sbom.xml"
        "bom.xml"
        "*.cdx.xml"
        "sbom.cdx.json"
        "bom.cdx.json"
    )

    for pattern in "${sbom_patterns[@]}"; do
        local found=$(find "$dir" -maxdepth 2 -name "$pattern" -type f 2>/dev/null | head -1)
        if [[ -n "$found" ]]; then
            echo "$found"
            return 0
        fi
    done

    return 1
}

# Function to check if directory has lock files that osv-scanner can use
has_scannable_files() {
    local dir="$1"

    # Check for common lock files and manifest files that osv-scanner supports
    local lock_patterns=(
        "package-lock.json"
        "yarn.lock"
        "pnpm-lock.yaml"
        "Gemfile.lock"
        "Cargo.lock"
        "go.mod"
        "go.sum"
        "requirements.txt"
        "poetry.lock"
        "Pipfile.lock"
        "composer.lock"
        "pom.xml"
        "gradle.lockfile"
        "mix.lock"
    )

    for pattern in "${lock_patterns[@]}"; do
        if find "$dir" -name "$pattern" -type f 2>/dev/null | grep -q .; then
            return 0
        fi
    done

    return 1
}

# Function to run osv-scanner
run_osv_scanner() {
    local target="$1"
    local is_sbom="$2"
    local output_file=$(mktemp)

    echo -e "${BLUE}Running osv-scanner...${NC}"

    local cmd
    if [[ "$is_sbom" == "true" ]]; then
        cmd="osv-scanner --sbom=$target --format=json"
    elif [[ "$TAINT_ANALYSIS" == "true" ]]; then
        cmd="osv-scanner --call-analysis=all $target --format=json"
    else
        cmd="osv-scanner --recursive $target --format=json"
    fi

    if eval "$cmd" > "$output_file" 2>/dev/null || [[ -s "$output_file" ]]; then
        echo -e "${GREEN}✓ Scan complete${NC}"
        echo "$output_file"
    else
        # Even with no vulnerabilities, osv-scanner returns non-zero
        if [[ -s "$output_file" ]]; then
            echo -e "${GREEN}✓ Scan complete${NC}"
            echo "$output_file"
        else
            echo -e "${YELLOW}⚠ Scan completed with warnings${NC}"
            echo "$output_file"
        fi
    fi
}

# Function to call Claude API
analyze_with_claude() {
    local scan_results="$1"
    local target_desc="$2"

    echo ""
    echo -e "${BLUE}Analyzing with Claude AI...${NC}"

    # Read scan results
    local results_content=$(cat "$scan_results")

    # Determine analysis mode description
    local analysis_mode="Vulnerability Analysis Only"
    if [[ "$DEEP_ANALYSIS" == "true" ]]; then
        analysis_mode="Deep Analysis (Vulnerability + Taint + All Available)"
    elif [[ "$TAINT_ANALYSIS" == "true" ]]; then
        analysis_mode="Vulnerability + Taint/Reachability Analysis"
    fi

    # Prepare prompt
    local prompt="I need you to analyze these SBOM vulnerability scan results from osv-scanner.

Target: $target_desc
Analysis Mode: $analysis_mode
Taint Analysis Enabled: $TAINT_ANALYSIS

Scan Results:
\`\`\`json
$results_content
\`\`\`

Please provide a purely objective vulnerability analysis with:

1. **Executive Summary**
   - Total vulnerabilities found
   - Breakdown by severity (Critical, High, Medium, Low)
   - Key risk indicators identified

2. **Critical Findings**
   - List critical and high severity vulnerabilities
   - Include CVE IDs, affected packages, and CVSS scores
   - Note any CISA KEV matches if applicable
   - If taint analysis enabled: indicate reachability status (CALLED, NOT CALLED, UNKNOWN)

3. **Taint Analysis Results** (if enabled)
   - Vulnerabilities that are CALLED (actually reachable from code)
   - Vulnerabilities that are NOT CALLED (present but not used)
   - Vulnerabilities with UNKNOWN reachability
   - Impact of reachability on actual risk

4. **Vulnerability Distribution**
   - Breakdown by package/dependency
   - Direct vs. transitive dependencies
   - Severity distribution across components

5. **Risk Assessment**
   - Overall security posture based on findings
   - Supply chain risks identified
   - Known exploitation (CISA KEV) presence

IMPORTANT: Provide ONLY factual analysis and observations. Do NOT include:
- Recommendations or action items
- Remediation priorities or timelines
- Suggested version upgrades or fixes
- \"Should\" or \"must\" statements

Be specific and data-driven. Focus exclusively on what IS, not what should be done about it."

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
    echo "  Claude AI Analysis"
    echo "========================================="
    echo ""
    echo "$analysis"
    echo ""
}

# Function to cleanup
cleanup() {
    if [[ "$CLEANUP" == true ]] && [[ -n "$TEMP_DIR" ]] && [[ -d "$TEMP_DIR" ]]; then
        rm -rf "$TEMP_DIR"
    fi
}

# Parse command line arguments
TARGET=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--taint-analysis)
            TAINT_ANALYSIS=true
            shift
            ;;
        -d|--deep)
            DEEP_ANALYSIS=true
            TAINT_ANALYSIS=true  # Deep includes taint
            shift
            ;;
        -k|--api-key)
            ANTHROPIC_API_KEY="$2"
            shift 2
            ;;
        --keep-clone)
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

# Validate
if [[ -z "$TARGET" ]]; then
    echo -e "${RED}Error: No target specified${NC}"
    usage
fi

# Main
echo ""
echo "========================================="
echo "  SBOM Analyzer with Claude AI"
echo "========================================="
echo ""

# Show analysis mode
if [[ "$DEEP_ANALYSIS" == "true" ]]; then
    echo -e "${BLUE}Analysis Mode: Deep (Vulnerability + Taint + All)${NC}"
elif [[ "$TAINT_ANALYSIS" == "true" ]]; then
    echo -e "${BLUE}Analysis Mode: Vulnerability + Taint/Reachability${NC}"
else
    echo -e "${BLUE}Analysis Mode: Vulnerability Only (default)${NC}"
fi
echo ""

check_prerequisites

# Determine target type
if is_sbom_file "$TARGET"; then
    echo -e "${GREEN}Target: SBOM file${NC}"
    TARGET_DESC="SBOM: $TARGET"
    SCAN_RESULTS=$(run_osv_scanner "$TARGET" "true")
    analyze_with_claude "$SCAN_RESULTS" "$TARGET_DESC"
    rm -f "$SCAN_RESULTS"
elif is_git_url "$TARGET"; then
    echo -e "${GREEN}Target: Git repository${NC}"
    if clone_repository "$TARGET"; then
        TARGET_DESC="Repository: $TARGET"
        echo ""

        # Check for existing SBOM first
        echo -e "${BLUE}Checking for existing SBOM...${NC}"
        EXISTING_SBOM=$(find_sbom "$TEMP_DIR")
        if [[ -n "$EXISTING_SBOM" ]]; then
            echo -e "${GREEN}✓ Found existing SBOM: ${EXISTING_SBOM#$TEMP_DIR/}${NC}"
            echo ""
            SCAN_RESULTS=$(run_osv_scanner "$EXISTING_SBOM" "true")
        else
            echo -e "${YELLOW}⚠ No existing SBOM found${NC}"
            echo ""

            # Check for lock files
            echo -e "${BLUE}Checking for lock files...${NC}"
            if has_scannable_files "$TEMP_DIR"; then
                echo -e "${GREEN}✓ Found scannable lock files${NC}"
                echo -e "${BLUE}Using osv-scanner to scan dependencies from lock files...${NC}"
            else
                echo -e "${YELLOW}⚠ No lock files found. osv-scanner will do its best to find dependencies.${NC}"
            fi
            echo ""

            SCAN_RESULTS=$(run_osv_scanner "$TEMP_DIR" "false")
        fi

        if [[ -n "$SCAN_RESULTS" ]] && [[ -f "$SCAN_RESULTS" ]]; then
            analyze_with_claude "$SCAN_RESULTS" "$TARGET_DESC"
            rm -f "$SCAN_RESULTS"
        else
            echo -e "${RED}✗ Scan failed or produced no results${NC}"
        fi

        cleanup
    fi
elif [[ -d "$TARGET" ]]; then
    echo -e "${GREEN}Target: Local directory${NC}"
    TARGET_DESC="Directory: $TARGET"
    echo ""

    # Check for existing SBOM first
    echo -e "${BLUE}Checking for existing SBOM...${NC}"
    EXISTING_SBOM=$(find_sbom "$TARGET")
    if [[ -n "$EXISTING_SBOM" ]]; then
        echo -e "${GREEN}✓ Found existing SBOM: ${EXISTING_SBOM#$TARGET/}${NC}"
        echo ""
        SCAN_RESULTS=$(run_osv_scanner "$EXISTING_SBOM" "true")
    else
        echo -e "${YELLOW}⚠ No existing SBOM found${NC}"
        echo ""

        # Check for lock files
        echo -e "${BLUE}Checking for lock files...${NC}"
        if has_scannable_files "$TARGET"; then
            echo -e "${GREEN}✓ Found scannable lock files${NC}"
            echo -e "${BLUE}Using osv-scanner to scan dependencies from lock files...${NC}"
        else
            echo -e "${YELLOW}⚠ No lock files found. osv-scanner will do its best to find dependencies.${NC}"
        fi
        echo ""

        SCAN_RESULTS=$(run_osv_scanner "$TARGET" "false")
    fi

    if [[ -n "$SCAN_RESULTS" ]] && [[ -f "$SCAN_RESULTS" ]]; then
        analyze_with_claude "$SCAN_RESULTS" "$TARGET_DESC"
        rm -f "$SCAN_RESULTS"
    else
        echo -e "${RED}✗ Scan failed or produced no results${NC}"
    fi
else
    echo -e "${RED}Error: Invalid target${NC}"
    exit 1
fi

echo "========================================="
echo -e "${GREEN}  Analysis Complete${NC}"
echo "========================================="
echo ""
