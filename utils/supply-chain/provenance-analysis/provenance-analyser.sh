#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# 
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Provenance Analyser Script
# Analyzes SBOMs and repositories for SLSA provenance using sigstore
# Verifies build attestations and assesses SLSA levels
# Usage: ./provenance-analyser.sh [options] <target>
#############################################################################

set -e

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PARENT_DIR="$(dirname "$SCRIPT_DIR")"
REPO_ROOT="$(cd "$PARENT_DIR/.." && pwd)"
CONFIG_FILE="$PARENT_DIR/config.json"

# Load global libraries
source "$REPO_ROOT/lib/sbom.sh"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Default options
OUTPUT_FORMAT="markdown"
VERIFY_SIGNATURES=false
ANTHROPIC_API_KEY="${ANTHROPIC_API_KEY:-}"
# Claude enabled by default if API key is set
USE_CLAUDE=false
if [[ -n "$ANTHROPIC_API_KEY" ]]; then
    USE_CLAUDE=true
fi
MIN_SLSA_LEVEL=0
TEMP_DIR=""
LOCAL_PATH=""
SBOM_FILE=""
CLEANUP=true
STRICT_MODE=false
MULTI_REPO_MODE=false
TARGETS_LIST=()
PARALLEL=true
PARALLEL_JOBS=$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo "4")

# Function to print usage
usage() {
    cat << EOF
Provenance Analyser - SLSA provenance verification and assessment

Usage: $0 [OPTIONS] [target]

TARGET:
    SBOM file path          Analyze an existing SBOM (JSON/XML)
    Git repository URL      Clone and analyze repository
    Local directory path    Analyze local repository
    Package URL (purl)      Analyze specific package
    (If no target specified, uses config.json)

MULTI-REPO OPTIONS:
    --org ORG_NAME          Scan all repos in GitHub organization
    --repo OWNER/REPO       Scan specific repository
    --config FILE           Use alternate config file

ANALYSIS OPTIONS:
    --verify-signatures     Cryptographically verify signatures (requires cosign)
    --min-level LEVEL       Require minimum SLSA level (0-4)
    --strict                Fail on missing provenance or low SLSA level
    --claude                Use Claude AI for enhanced analysis (requires ANTHROPIC_API_KEY)
    --parallel              Enable parallel package analysis (faster)
    --jobs N                Number of parallel jobs (default: CPU count)
    --local-path PATH       Use pre-cloned repository at PATH (skips cloning)
    -f, --format FORMAT     Output format: table|json|markdown (default: markdown)
    -o, --output FILE       Write results to file
    -k, --keep-clone        Keep cloned repository (don't cleanup)
    -h, --help              Show this help message

EXAMPLES:
    # Analyze an SBOM file
    $0 /path/to/sbom.json

    # Analyze with signature verification
    $0 --verify-signatures https://github.com/org/repo

    # Require minimum SLSA level
    $0 --min-level 2 --strict sbom.json

    # Scan entire GitHub organization
    $0 --org myorg --min-level 1

    # Analyze specific package
    $0 pkg:npm/express@4.17.1

EOF
    exit 1
}

# Function to check if cosign is installed
check_cosign() {
    if ! command -v cosign &> /dev/null; then
        echo -e "${YELLOW}⚠ cosign not installed - signature verification disabled${NC}" >&2
        echo "  Install: brew install cosign" >&2
        return 1
    fi
    return 0
}

# Function to check if rekor-cli is installed
check_rekor() {
    if ! command -v rekor-cli &> /dev/null; then
        echo -e "${YELLOW}⚠ rekor-cli not installed - transparency log checks disabled${NC}" >&2
        echo "  Install: brew install rekor-cli" >&2
        return 1
    fi
    return 0
}

# Function to check if syft is installed
check_syft() {
    if ! command -v syft &> /dev/null; then
        echo -e "${RED}Error: syft is not installed${NC}"
        echo "Install: brew install syft"
        exit 1
    fi
}

# Function to expand org into repos
expand_org_repos() {
    local org="$1"
    echo -e "${BLUE}Fetching repositories for org: $org${NC}" >&2

    if ! command -v gh &> /dev/null; then
        echo -e "${RED}Error: gh (GitHub CLI) is required for org scanning${NC}" >&2
        echo "Install: brew install gh" >&2
        return 1
    fi

    local repos=$(gh repo list "$org" --limit 1000 --json nameWithOwner --jq '.[].nameWithOwner' 2>/dev/null || echo "")

    if [[ -z "$repos" ]]; then
        echo -e "${YELLOW}⚠ No repositories found for org: $org${NC}" >&2
        return 1
    fi

    echo "$repos"
}

# Function to load targets from config
load_config_targets() {
    if [[ ! -f "$CONFIG_FILE" ]]; then
        return 1
    fi

    local config_orgs=$(jq -r '.github.organizations[]?' "$CONFIG_FILE" 2>/dev/null || echo "")
    local config_repos=$(jq -r '.github.repositories[]?' "$CONFIG_FILE" 2>/dev/null || echo "")

    # Add orgs
    while IFS= read -r org; do
        [[ -n "$org" ]] && TARGETS_LIST+=("org:$org")
    done <<< "$config_orgs"

    # Add repos
    while IFS= read -r repo; do
        [[ -n "$repo" ]] && TARGETS_LIST+=("repo:$repo")
    done <<< "$config_repos"

    return 0
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

# Function to detect if target is a package URL
is_purl() {
    [[ "$1" =~ ^pkg: ]]
}

# Function to clone repository
clone_repository() {
    local repo_url="$1"
    TEMP_DIR=$(mktemp -d)

    echo -e "${BLUE}Cloning repository...${NC}"
    if git clone --depth 1 "$repo_url" "$TEMP_DIR" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Repository cloned${NC}"
        return 0
    else
        echo -e "${RED}✗ Failed to clone repository${NC}"
        return 1
    fi
}

# Function to generate SBOM if not exists
# Uses global SBOM library from utils/lib/sbom.sh
generate_sbom_for_provenance() {
    local target_dir="$1"
    local output_file="$2"

    echo -e "${BLUE}Generating SBOM using global library...${NC}"

    # Call global generate_sbom function (from utils/lib/sbom.sh)
    # Pass force=true to regenerate even if exists
    if generate_sbom "$target_dir" "$output_file" "true" 2>&1 | grep -v "^$"; then
        if [[ -f "$output_file" ]]; then
            echo -e "${GREEN}✓ SBOM generated${NC}"
            return 0
        fi
    fi

    echo -e "${RED}✗ SBOM generation failed${NC}"
    return 1
}

# Function to cleanup temporary files
cleanup() {
    if [[ "$CLEANUP" == true ]] && [[ -n "$TEMP_DIR" ]] && [[ -d "$TEMP_DIR" ]]; then
        rm -rf "$TEMP_DIR"
    fi
}

# Ensure cleanup on script exit (normal, error, or interrupt)
trap cleanup EXIT

# Function to parse purl components
parse_purl() {
    local purl="$1"

    # Extract ecosystem and package
    if [[ "$purl" =~ ^pkg:([^/]+)/(.+)@(.+)$ ]]; then
        echo "${BASH_REMATCH[1]}|${BASH_REMATCH[2]}|${BASH_REMATCH[3]}"
    elif [[ "$purl" =~ ^pkg:([^/]+)/(.+)$ ]]; then
        echo "${BASH_REMATCH[1]}|${BASH_REMATCH[2]}|latest"
    else
        return 1
    fi
}

# Function to check npm provenance
check_npm_provenance() {
    local package="$1"
    local version="$2"

    local registry_data=$(curl -s "https://registry.npmjs.org/$package/$version" 2>/dev/null)

    if [[ -z "$registry_data" ]]; then
        echo "UNAVAILABLE|0|No registry data"
        return
    fi

    local has_signatures=$(echo "$registry_data" | jq -r '.dist.signatures // empty' 2>/dev/null)
    local has_attestations=$(echo "$registry_data" | jq -r '.dist.attestations // empty' 2>/dev/null)

    if [[ -n "$has_attestations" ]]; then
        # npm native provenance (SLSA Level 3)
        echo "VERIFIED|3|npm native provenance"
    elif [[ -n "$has_signatures" ]]; then
        echo "SIGNED|2|Package signed"
    else
        echo "NONE|0|No provenance found"
    fi
}

# Function to check GitHub provenance
check_github_provenance() {
    local repo="$1"

    if ! command -v gh &> /dev/null; then
        echo "UNAVAILABLE|0|gh CLI not available"
        return
    fi

    # Check if repo has any releases with attestations
    local latest_release=$(gh api "repos/$repo/releases/latest" 2>/dev/null | jq -r '.tag_name // empty')

    if [[ -z "$latest_release" ]]; then
        echo "NONE|0|No releases found"
        return
    fi

    # Try to get attestations for latest release
    local attestations=$(gh api "repos/$repo/attestations" 2>/dev/null || echo "")

    if [[ -n "$attestations" ]]; then
        echo "VERIFIED|3|GitHub attestations found"
    else
        echo "NONE|0|No attestations found"
    fi
}

# Function to assess SLSA level based on checks
assess_slsa_level() {
    local has_provenance="$1"
    local is_signed="$2"
    local has_attestation="$3"
    local trusted_builder="$4"

    if [[ "$has_attestation" == "true" ]] && [[ "$trusted_builder" == "true" ]]; then
        echo "3"
    elif [[ "$is_signed" == "true" ]] && [[ "$has_provenance" == "true" ]]; then
        echo "2"
    elif [[ "$has_provenance" == "true" ]]; then
        echo "1"
    else
        echo "0"
    fi
}

# Function to analyze SBOM for provenance
analyze_sbom() {
    local sbom_file="$1"

    echo -e "${BLUE}Analyzing SBOM for provenance...${NC}"
    echo ""

    # Parse SBOM and extract components
    local components=$(jq -r '.components[]? | @json' "$sbom_file" 2>/dev/null)

    if [[ -z "$components" ]]; then
        echo -e "${RED}No components found in SBOM${NC}"
        return 1
    fi

    local total=0
    local with_provenance=0
    local verified=0
    local level_0=0
    local level_1=0
    local level_2=0
    local level_3=0
    local level_4=0

    echo "Package Analysis:"
    echo "==============================================="
    echo ""

    if [[ "$PARALLEL" == "true" ]]; then
        # Parallel processing mode
        echo -e "${CYAN}Parallel mode enabled: using $PARALLEL_JOBS workers${NC}"
        echo ""

        # Create temp directory for results
        local results_dir=$(mktemp -d)

        # Export functions for subshells
        export -f check_npm_provenance
        export -f parse_purl

        # Process components in parallel
        echo "$components" | nl -w1 -s'|' | xargs -P "$PARALLEL_JOBS" -I {} bash -c '
            IFS="|" read -r num component <<< "{}"

            name=$(echo "$component" | jq -r ".name // \"unknown\"")
            version=$(echo "$component" | jq -r ".version // \"unknown\"")
            purl=$(echo "$component" | jq -r ".purl // empty")

            echo "Package: $name@$version" >&2

            if [[ -n "$purl" ]]; then
                IFS="|" read -r ecosystem pkg ver <<< "$(parse_purl "$purl")"

                case "$ecosystem" in
                    npm)
                        IFS="|" read -r status level detail <<< "$(check_npm_provenance "$pkg" "$ver")"
                        ;;
                    *)
                        status="UNSUPPORTED"
                        level="0"
                        detail="Ecosystem not yet supported"
                        ;;
                esac

                echo "  Provenance:    $status" >&2
                echo "  SLSA Level:    $level" >&2
                echo "  Details:       $detail" >&2

                # Write results to file
                echo "$status|$level" > "'"$results_dir"'/result_$num.txt"
            else
                echo "  Provenance:    UNKNOWN (no purl)" >&2
                echo "  SLSA Level:    0" >&2
                echo "UNKNOWN|0" > "'"$results_dir"'/result_$num.txt"
            fi

            echo "" >&2
        '

        # Aggregate results
        for result_file in "$results_dir"/result_*.txt; do
            if [[ -f "$result_file" ]]; then
                ((total++))
                IFS='|' read -r status level < "$result_file"

                if [[ "$status" != "NONE" ]] && [[ "$status" != "UNSUPPORTED" ]] && [[ "$status" != "UNKNOWN" ]]; then
                    ((with_provenance++))
                fi

                if [[ "$status" == "VERIFIED" ]]; then
                    ((verified++))
                fi

                case "$level" in
                    0) ((level_0++)) ;;
                    1) ((level_1++)) ;;
                    2) ((level_2++)) ;;
                    3) ((level_3++)) ;;
                    4) ((level_4++)) ;;
                esac
            fi
        done

        # Cleanup
        rm -rf "$results_dir"

    else
        # Sequential processing mode
        while IFS= read -r component; do
            ((total++))

            local name=$(echo "$component" | jq -r '.name // "unknown"')
            local version=$(echo "$component" | jq -r '.version // "unknown"')
            local purl=$(echo "$component" | jq -r '.purl // empty')

            echo "Package: $name@$version"

            if [[ -n "$purl" ]]; then
                IFS='|' read -r ecosystem pkg ver <<< "$(parse_purl "$purl")"

                case "$ecosystem" in
                    npm)
                        IFS='|' read -r status level detail <<< "$(check_npm_provenance "$pkg" "$ver")"
                        ;;
                    *)
                        status="UNSUPPORTED"
                        level="0"
                        detail="Ecosystem not yet supported"
                        ;;
                esac

                echo "  Provenance:    $status"
                echo "  SLSA Level:    $level"
                echo "  Details:       $detail"

                if [[ "$status" != "NONE" ]] && [[ "$status" != "UNSUPPORTED" ]]; then
                    ((with_provenance++))
                fi

                if [[ "$status" == "VERIFIED" ]]; then
                    ((verified++))
                fi

                case "$level" in
                    0) ((level_0++)) ;;
                    1) ((level_1++)) ;;
                    2) ((level_2++)) ;;
                    3) ((level_3++)) ;;
                    4) ((level_4++)) ;;
                esac

            else
                echo "  Provenance:    UNKNOWN (no purl)"
                echo "  SLSA Level:    0"
                ((level_0++))
            fi

            echo ""
        done <<< "$components"
    fi

    echo "==============================================="
    echo "Summary:"
    echo "  Total packages:        $total"
    echo "  With provenance:       $with_provenance ($(( total > 0 ? with_provenance * 100 / total : 0 ))%)"
    echo "  Verified:              $verified ($(( total > 0 ? verified * 100 / total : 0 ))%)"
    echo ""
    echo "SLSA Level Distribution:"
    echo "  Level 0 (No guarantees):         $level_0"
    echo "  Level 1 (Documentation):         $level_1"
    echo "  Level 2 (Signed provenance):     $level_2"
    echo "  Level 3 (Hardened builds):       $level_3"
    echo "  Level 4 (Highest assurance):     $level_4"
    echo ""

    # Check against minimum level if specified
    if [[ $MIN_SLSA_LEVEL -gt 0 ]]; then
        local meeting_min=0
        for ((i=MIN_SLSA_LEVEL; i<=4; i++)); do
            eval "meeting_min=\$((meeting_min + level_$i))"
        done

        local pct=$(( total > 0 ? meeting_min * 100 / total : 0 ))
        echo "Packages meeting minimum SLSA Level $MIN_SLSA_LEVEL: $meeting_min ($pct%)"

        if [[ "$STRICT_MODE" == "true" ]] && [[ $meeting_min -lt $total ]]; then
            echo -e "${RED}✗ STRICT MODE: Not all packages meet minimum SLSA level${NC}"
            return 1
        fi
    fi
}

# Function to analyze repository
analyze_repository() {
    local repo_path="$1"

    echo -e "${BLUE}Analyzing repository for provenance...${NC}"
    echo ""

    # Use shared SBOM if provided, otherwise generate new one
    local sbom_file=""
    local cleanup_sbom=false

    if [[ -n "$SBOM_FILE" ]] && [[ -f "$SBOM_FILE" ]]; then
        echo -e "${GREEN}Using shared SBOM file${NC}"
        sbom_file="$SBOM_FILE"
        cleanup_sbom=false
    else
        # Generate SBOM using global library
        sbom_file="$repo_path/generated-sbom.json"
        if generate_sbom_for_provenance "$repo_path" "$sbom_file"; then
            cleanup_sbom=true
        else
            echo -e "${RED}Failed to generate SBOM for analysis${NC}"
            return 1
        fi
    fi

    # Analyze the SBOM
    analyze_sbom "$sbom_file"

    # Clean up only if we generated it
    if [[ "$cleanup_sbom" == "true" ]]; then
        rm -f "$sbom_file"
    fi
}

# Function to analyze single package
analyze_package() {
    local purl="$1"

    echo -e "${BLUE}Analyzing package: $purl${NC}"
    echo ""

    IFS='|' read -r ecosystem pkg ver <<< "$(parse_purl "$purl")"

    if [[ -z "$ecosystem" ]]; then
        echo -e "${RED}Invalid package URL format${NC}"
        return 1
    fi

    echo "Package: $pkg@$ver"
    echo "Ecosystem: $ecosystem"
    echo ""

    case "$ecosystem" in
        npm)
            IFS='|' read -r status level detail <<< "$(check_npm_provenance "$pkg" "$ver")"
            ;;
        *)
            status="UNSUPPORTED"
            level="0"
            detail="Ecosystem not yet supported (npm only currently)"
            ;;
    esac

    echo "Provenance Status: $status"
    echo "SLSA Level: $level"
    echo "Details: $detail"
    echo ""

    if [[ $level -lt $MIN_SLSA_LEVEL ]] && [[ "$STRICT_MODE" == "true" ]]; then
        echo -e "${RED}✗ STRICT MODE: Package does not meet minimum SLSA level $MIN_SLSA_LEVEL${NC}"
        return 1
    fi
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

    # Load RAG knowledge for supply chain security best practices
    local rag_context=""
    local repo_root="$( cd "$SCRIPT_DIR/../.." && pwd )"
    local rag_dir="$repo_root/rag/supply-chain"

    if [[ -f "$rag_dir/slsa/slsa-specification.md" ]]; then
        rag_context+="# SLSA Specification\n\n"
        rag_context+=$(head -200 "$rag_dir/slsa/slsa-specification.md" | tail -n +1)
        rag_context+="\n\n"
    fi

    if [[ -f "$rag_dir/sigstore/sigstore-reference.md" ]]; then
        rag_context+="# Sigstore Reference\n\n"
        rag_context+=$(head -200 "$rag_dir/sigstore/sigstore-reference.md" | tail -n +1)
        rag_context+="\n\n"
    fi

    local prompt="You are a supply chain security expert. Analyze this SLSA provenance scan data and produce a PRIORITY-BASED ACTION REPORT.

# Supply Chain Security Knowledge Base
$rag_context

# Analysis Requirements

For each finding, you MUST:
1. **Explain the issue** - What is the provenance/SLSA gap and why is it a problem?
2. **Justify the recommendation** - Why this action is needed based on SLSA/supply chain best practices
3. **Reference knowledge base** - Cite relevant SLSA levels or best practices from the knowledge base above
4. **Provide specific actions** - Exact configurations, tooling changes, or attestation improvements

# Output Format

## 🔴 CRITICAL PRIORITY (Immediate - 0-24 hours)
- No provenance, SLSA Level 0 packages, or unverified attestations
- For each item:
  * **Issue**: [Description]
  * **Risk**: [Why this threatens supply chain integrity]
  * **Best Practice Reference**: [Cite SLSA requirements or best practices]
  * **Action**: [Specific steps to add provenance/attestations]
  * **Timeline**: Immediate

## 🟠 HIGH PRIORITY (Urgent - 1-7 days)
- Low SLSA levels (1-2), missing signatures, or weak build environments
- Same structured format as above

## 🟡 MEDIUM PRIORITY (Important - 1-30 days)
- SLSA level improvements (moving from 2→3 or 3→4)
- Same structured format as above

## 🟢 LOW PRIORITY (Monitor - 30+ days)
- Documentation improvements, enhanced transparency
- Same structured format as above

## 📊 Summary & Strategic Recommendations
- Overall SLSA maturity posture
- Build pipeline hardening opportunities
- Automation recommendations (GitHub attestations, Sigstore integration)
- Path to SLSA Level 3+ compliance

# Scan Data:
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

# Function to analyze a single target
analyze_single_target() {
    local target="$1"

    # If LOCAL_PATH is set, use it instead of target
    if [[ -n "$LOCAL_PATH" ]]; then
        if [[ ! -d "$LOCAL_PATH" ]]; then
            echo -e "${RED}Error: Local path does not exist: $LOCAL_PATH${NC}"
            return 1
        fi
        echo -e "${GREEN}Target: Local directory (provided)${NC}"
        echo ""
        analyze_repository "$LOCAL_PATH"
        return $?
    fi

    if is_purl "$target"; then
        analyze_package "$target"
    elif is_sbom_file "$target"; then
        echo -e "${GREEN}Target: SBOM file${NC}"
        echo ""
        analyze_sbom "$target"
    elif is_git_url "$target"; then
        echo -e "${GREEN}Target: Git repository${NC}"
        echo ""
        if clone_repository "$target"; then
            analyze_repository "$TEMP_DIR"
        fi
    elif [[ -d "$target" ]]; then
        echo -e "${GREEN}Target: Local directory${NC}"
        echo ""
        analyze_repository "$target"
    else
        echo -e "${RED}Error: Invalid target${NC}"
        echo "Target must be:"
        echo "  - Path to SBOM file (.json, .xml)"
        echo "  - Git repository URL"
        echo "  - Local directory path"
        echo "  - Package URL (pkg:ecosystem/package@version)"
        return 1
    fi
}

# Load cost tracking if using Claude
if [[ "$USE_CLAUDE" == "true" ]]; then
    REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
    if [ -f "$REPO_ROOT/lib/claude-cost.sh" ]; then
        source "$REPO_ROOT/lib/claude-cost.sh"
        init_cost_tracking
    fi
fi

# Parse command line arguments
OUTPUT_FILE=""
TARGET=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --verify-signatures)
            VERIFY_SIGNATURES=true
            shift
            ;;
        --min-level)
            MIN_SLSA_LEVEL="$2"
            shift 2
            ;;
        --strict)
            STRICT_MODE=true
            shift
            ;;
        --local-path)
            LOCAL_PATH="$2"
            CLEANUP=false  # Don't cleanup externally managed directory
            shift 2
            ;;
        --sbom-file)
            SBOM_FILE="$2"
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
        -k|--keep-clone)
            CLEANUP=false
            shift
            ;;
        --org)
            TARGETS_LIST+=("org:$2")
            MULTI_REPO_MODE=true
            shift 2
            ;;
        --repo)
            TARGETS_LIST+=("repo:$2")
            MULTI_REPO_MODE=true
            shift 2
            ;;
        --config)
            CONFIG_FILE="$2"
            shift 2
            ;;
        --claude)
            USE_CLAUDE=true
            shift
            ;;
        --no-claude)
            USE_CLAUDE=false
            shift
            ;;
        --parallel)
            PARALLEL=true
            shift
            ;;
        --jobs)
            PARALLEL_JOBS="$2"
            shift 2
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

# Determine targets
if [[ "$MULTI_REPO_MODE" == false ]] && [[ -z "$TARGET" ]] && [[ ${#TARGETS_LIST[@]} -eq 0 ]]; then
    if ! load_config_targets; then
        echo -e "${RED}Error: No targets specified and no config file found${NC}"
        echo ""
        echo "Specify targets via:"
        echo "  - Single target: $0 <sbom|repo-url|directory|purl>"
        echo "  - Organization:  $0 --org myorg"
        echo "  - Repositories:  $0 --repo owner/repo1 --repo owner/repo2"
        echo "  - Config file:   Create $CONFIG_FILE with targets"
        echo ""
        exit 1
    fi
    MULTI_REPO_MODE=true
fi

# Main script
echo ""
echo "========================================="
echo "  Provenance Analyser (SLSA)"
echo "========================================="
echo ""

# Check Claude AI status first
if [[ "$USE_CLAUDE" == "true" ]] && [[ -n "$ANTHROPIC_API_KEY" ]]; then
    echo -e "${GREEN}🤖 Claude AI: ENABLED (analyzing results with AI)${NC}"
    echo ""
elif [[ -z "$ANTHROPIC_API_KEY" ]]; then
    echo -e "${YELLOW}ℹ️  Claude AI: DISABLED (no API key found)${NC}"
    echo -e "${CYAN}   Set ANTHROPIC_API_KEY to enable AI-enhanced analysis${NC}"
    echo -e "${CYAN}   Get your key at: https://console.anthropic.com/settings/keys${NC}"
    echo ""
    USE_CLAUDE=false
fi

# Check prerequisites
if [[ "$VERIFY_SIGNATURES" == "true" ]]; then
    check_cosign || VERIFY_SIGNATURES=false
    check_rekor
fi

if [[ "$MULTI_REPO_MODE" == false ]]; then
    check_syft
fi

# Capture output for Claude analysis if enabled
if [[ "$USE_CLAUDE" == "true" ]]; then
    analysis_output=$(
        # Multi-repo mode or single target
        if [[ "$MULTI_REPO_MODE" == true ]]; then
            echo "Multi-repository mode: ${#TARGETS_LIST[@]} target(s)"
            echo ""

            for target_spec in "${TARGETS_LIST[@]}"; do
                if [[ "$target_spec" =~ ^org: ]]; then
                    org="${target_spec#org:}"
                    # Extract org name from URL if provided
                    if [[ "$org" =~ github\.com/orgs/([^/]+) ]]; then
                        org="${BASH_REMATCH[1]}"
                    elif [[ "$org" =~ github\.com/([^/]+) ]]; then
                        org="${BASH_REMATCH[1]}"
                    fi
                    org="${org%/}"  # Remove trailing slashes
                    repos=$(expand_org_repos "$org" 2>&1)

                    if [[ -z "$repos" ]]; then
                        continue
                    fi

                    while IFS= read -r repo; do
                        if [[ -n "$repo" ]]; then
                            echo ""
                            echo "========================================="
                            echo "Analyzing: $repo"
                            echo "========================================="
                            echo ""
                            analyze_single_target "https://github.com/$repo" 2>&1
                        fi
                    done <<< "$repos"

                elif [[ "$target_spec" =~ ^repo: ]]; then
                    repo="${target_spec#repo:}"
                    echo ""
                    echo "========================================="
                    echo "Analyzing: $repo"
                    echo "========================================="
                    echo ""
                    analyze_single_target "https://github.com/$repo" 2>&1
                fi
            done
        else
            analyze_single_target "$TARGET" 2>&1
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
    # Multi-repo mode or single target
    if [[ "$MULTI_REPO_MODE" == true ]]; then
        echo -e "${BLUE}Multi-repository mode: ${#TARGETS_LIST[@]} target(s)${NC}"
        echo ""

        for target_spec in "${TARGETS_LIST[@]}"; do
            if [[ "$target_spec" =~ ^org: ]]; then
                org="${target_spec#org:}"
                repos=$(expand_org_repos "$org")

                if [[ -z "$repos" ]]; then
                    continue
                fi

                while IFS= read -r repo; do
                    if [[ -n "$repo" ]]; then
                        echo ""
                        echo -e "${CYAN}=========================================${NC}"
                        echo -e "${CYAN}Analyzing: $repo${NC}"
                        echo -e "${CYAN}=========================================${NC}"
                        echo ""
                        analyze_single_target "https://github.com/$repo"
                    fi
                done <<< "$repos"

            elif [[ "$target_spec" =~ ^repo: ]]; then
                repo="${target_spec#repo:}"
                echo ""
                echo -e "${CYAN}=========================================${NC}"
                echo -e "${CYAN}Analyzing: $repo${NC}"
                echo -e "${CYAN}=========================================${NC}"
                echo ""
                analyze_single_target "https://github.com/$repo"
            fi
        done
    else
        analyze_single_target "$TARGET"
    fi
fi

echo ""
echo -e "${GREEN}✓ Analysis Complete${NC}"

if [[ -n "$OUTPUT_FILE" ]]; then
    echo ""
    echo -e "Results saved to: ${BLUE}$OUTPUT_FILE${NC}"
fi

echo ""
