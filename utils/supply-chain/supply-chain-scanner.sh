#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# 
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Supply Chain Scanner
# Central orchestrator for modular supply chain analysis
# Supports multi-repo/org scanning with persistent configuration
# Usage: ./supply-chain-scanner.sh [options] [targets...]
#############################################################################

set -e

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
UTILS_ROOT="$(dirname "$SCRIPT_DIR")"
REPO_ROOT="$(dirname "$UTILS_ROOT")"
CONFIG_FILE="$SCRIPT_DIR/config.json"
CONFIG_EXAMPLE="$SCRIPT_DIR/config.example.json"

# Load config library
if [[ -f "$UTILS_ROOT/lib/config-loader.sh" ]]; then
    source "$UTILS_ROOT/lib/config-loader.sh"
fi

# Load .env file if it exists in repository root
if [[ -f "$REPO_ROOT/.env" ]]; then
    set -a  # automatically export all variables
    source "$REPO_ROOT/.env"
    set +a  # stop automatically exporting
fi

# Load GitHub token from config.json if not in .env
# Priority: .env GITHUB_TOKEN > config.json github.pat
if [[ -f "$UTILS_ROOT/lib/config-loader.sh" ]]; then
    load_github_token 2>/dev/null || true
fi

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Default options
INTERACTIVE=false
SETUP_MODE=false
MODULES=()
TARGETS=()
OUTPUT_DIR=""
SHARED_REPO_DIR=""  # For sharing cloned repo across modules
SHARED_SBOM_FILE=""  # For sharing SBOM across modules
ANTHROPIC_API_KEY="${ANTHROPIC_API_KEY:-}"
# Claude enabled by default if API key is set
USE_CLAUDE=false
if [[ -n "$ANTHROPIC_API_KEY" ]]; then
    USE_CLAUDE=true
fi
PARALLEL=true

# Persona configuration
PERSONA=""
VALID_PERSONAS=("security-engineer" "software-engineer" "engineering-leader" "auditor" "all")

# Cleanup function
cleanup_shared_repo() {
    if [[ -n "$SHARED_REPO_DIR" ]] && [[ -d "$SHARED_REPO_DIR" ]]; then
        echo -e "${YELLOW}Cleaning up shared repository...${NC}"
        rm -rf "$SHARED_REPO_DIR"
    fi
    if [[ -n "$SHARED_SBOM_FILE" ]] && [[ -f "$SHARED_SBOM_FILE" ]]; then
        rm -f "$SHARED_SBOM_FILE"
    fi
}

# Ensure cleanup on script exit
trap cleanup_shared_repo EXIT

# Function to print usage
usage() {
    cat << EOF
Supply Chain Scanner - Modular supply chain security analysis

Usage: $0 [OPTIONS] [TARGETS...]

MODES:
    --setup                 Interactive setup of configuration
    --interactive, -i       Interactive mode (prompt for repos if not configured)

MODULES:
    --vulnerability, -v     Run vulnerability analysis
    --provenance, -p        Run provenance analysis (SLSA) - slow, queries npm registry per-package
    --package-health        Run package health analysis
    --legal                 Run legal compliance analysis (licenses, secrets, content)
    --all, -a               Run all analysis modules (includes provenance - may be slow)

ENHANCED ANALYSIS:
    --abandoned             Detect abandoned/deprecated packages
    --typosquat             Check for typosquatting risks
    --unused                Find unused dependencies
    --debt-score            Calculate technical debt scores
    --container-images      Analyze container images and recommend alternatives
    --library-recommend     Suggest library replacements

TARGETS:
    --org ORG_NAME          Scan all repos in GitHub organization
    --repo OWNER/REPO       Scan specific repository
    (If no targets specified, uses config.json)

OPTIONS:
    --output DIR, -o DIR    Output directory for reports
    --config FILE           Use alternate config file
    --claude                Use Claude AI for enhanced analysis (requires ANTHROPIC_API_KEY)
    --parallel              Batch API processing (default: enabled)
    --persona PERSONA       Analysis persona (security-engineer, software-engineer,
                           engineering-leader, auditor, all). Defaults to 'all'
                           which generates reports for all personas.
    --list-personas         Show available personas and descriptions
    -h, --help              Show this help message

PERSONAS:
    security-engineer       Technical vulnerability analysis with remediation priorities
    software-engineer       Developer-focused: dependencies, upgrades, build optimization
    engineering-leader      Executive summary with metrics and strategic recommendations
    auditor                 Compliance assessment with framework mappings and evidence
    all                     Generate reports for ALL personas (4 separate reports)

EXAMPLES:
    # Initial setup (configure GH PAT, repos, orgs)
    $0 --setup

    # Scan configured repos with vulnerability analysis
    $0 --vulnerability

    # Scan specific org
    $0 --vulnerability --org myorg

    # Scan specific repos
    $0 --vulnerability --repo owner/repo1 --repo owner/repo2

    # Interactive mode (prompt for selections)
    $0 --interactive --vulnerability

    # Run all modules on all configured repos
    $0 --all

    # Run with specific persona (security-focused analysis)
    $0 --vulnerability --repo owner/repo --persona security-engineer

    # Run with engineering leader persona (executive summary)
    $0 --all --repo owner/repo --persona engineering-leader

    # Generate reports for ALL personas (4 separate reports)
    $0 --vulnerability --repo owner/repo --persona all

    # List available personas
    $0 --list-personas

EOF
    exit 1
}

# Function to check prerequisites
check_prerequisites() {
    local missing=()

    if ! command -v jq &> /dev/null; then
        missing+=("jq")
    fi

    if ! command -v gh &> /dev/null; then
        missing+=("gh (GitHub CLI)")
    fi

    if [ ${#missing[@]} -gt 0 ]; then
        echo -e "${RED}Error: Missing required tools:${NC}"
        for tool in "${missing[@]}"; do
            echo "  - $tool"
        done
        echo ""
        echo "Install missing tools:"
        echo "  brew install jq gh"
        exit 1
    fi
}

# Function to check API key configuration
check_api_key() {
    echo -e "${BLUE}Checking Claude AI configuration...${NC}"

    if [ -n "$ANTHROPIC_API_KEY" ]; then
        local key_length=${#ANTHROPIC_API_KEY}
        echo -e "${GREEN}✓ ANTHROPIC_API_KEY is set (${key_length} chars)${NC}"
        echo -e "${GREEN}  Claude AI enhanced analysis is available${NC}"
        return 0
    else
        echo -e "${YELLOW}⚠ ANTHROPIC_API_KEY is NOT set${NC}"
        echo -e "${YELLOW}  Claude AI enhanced analysis will be disabled${NC}"
        echo ""
        echo "To enable Claude AI analysis:"
        echo "  1. Get an API key at: https://console.anthropic.com/settings/keys"
        echo "  2. Export it in your shell:"
        echo "     export ANTHROPIC_API_KEY='your-api-key-here'"
        echo "  3. Or add it to your shell profile (~/.zshrc or ~/.bashrc)"
        echo ""
        return 1
    fi
}

# Function to create default config
create_default_config() {
    if [[ ! -f "$CONFIG_EXAMPLE" ]]; then
        cat > "$CONFIG_EXAMPLE" << 'CONFIGEOF'
{
  "github": {
    "pat": "",
    "organizations": [],
    "repositories": []
  },
  "analysis": {
    "default_modules": ["vulnerability", "provenance", "package-health"],
    "output_dir": "./supply-chain-reports"
  }
}
CONFIGEOF
    fi

    if [[ ! -f "$CONFIG_FILE" ]]; then
        cp "$CONFIG_EXAMPLE" "$CONFIG_FILE"
        echo -e "${GREEN}✓ Created config file: $CONFIG_FILE${NC}"
    fi
}

# Function to setup configuration interactively
setup_config() {
    echo ""
    echo -e "${CYAN}=========================================${NC}"
    echo -e "${CYAN}  Supply Chain Scanner - Setup${NC}"
    echo -e "${CYAN}=========================================${NC}"
    echo ""

    create_default_config

    # Check GitHub authentication
    echo -e "${BLUE}Checking GitHub authentication...${NC}"
    if ! gh auth status &> /dev/null; then
        echo -e "${YELLOW}⚠ Not authenticated with GitHub${NC}"
        echo ""
        read -p "Would you like to authenticate now? (y/n) " -n 1 -r
        echo ""
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            gh auth login
        else
            echo -e "${RED}GitHub authentication required for org/repo access${NC}"
            exit 1
        fi
    else
        echo -e "${GREEN}✓ GitHub authentication valid${NC}"
    fi

    # Get GitHub PAT
    echo ""
    echo -e "${BLUE}GitHub Personal Access Token${NC}"
    echo "For API rate limit increases and private repo access"
    echo "Leave blank to skip (will use gh CLI authentication)"
    read -p "Enter PAT (or press Enter to skip): " -s gh_pat
    echo ""

    # List available orgs
    echo ""
    echo -e "${BLUE}Fetching your GitHub organizations...${NC}"
    local orgs=$(gh api user/orgs --jq '.[].login' 2>/dev/null || echo "")

    local selected_orgs=()
    if [[ -n "$orgs" ]]; then
        echo "Available organizations:"
        local i=1
        declare -a org_array
        while IFS= read -r org; do
            org_array+=("$org")
            echo "  $i) $org"
            ((i++))
        done <<< "$orgs"

        echo ""
        echo "Select organizations to scan (comma-separated numbers, or 'all', or press Enter to skip):"
        read -p "> " org_selection

        if [[ "$org_selection" == "all" ]]; then
            selected_orgs=("${org_array[@]}")
        elif [[ -n "$org_selection" ]]; then
            IFS=',' read -ra selections <<< "$org_selection"
            for sel in "${selections[@]}"; do
                sel=$(echo "$sel" | xargs) # trim whitespace
                if [[ "$sel" =~ ^[0-9]+$ ]] && [ "$sel" -ge 1 ] && [ "$sel" -le "${#org_array[@]}" ]; then
                    selected_orgs+=("${org_array[$((sel-1))]}")
                fi
            done
        fi
    fi

    # Add specific repositories
    echo ""
    echo -e "${BLUE}Add specific repositories (format: owner/repo)${NC}"
    echo "Enter repositories one per line. Press Enter on empty line to finish."

    local selected_repos=()
    while true; do
        read -p "> " repo
        if [[ -z "$repo" ]]; then
            break
        fi
        if [[ "$repo" =~ ^[a-zA-Z0-9_-]+/[a-zA-Z0-9_.-]+$ ]]; then
            selected_repos+=("$repo")
        else
            echo -e "${YELLOW}  Invalid format. Use: owner/repo${NC}"
        fi
    done

    # Build config JSON
    local orgs_json=$(printf '%s\n' "${selected_orgs[@]}" | jq -R . | jq -s .)
    local repos_json=$(printf '%s\n' "${selected_repos[@]}" | jq -R . | jq -s .)

    cat > "$CONFIG_FILE" << CONFIGEOF
{
  "github": {
    "pat": "$gh_pat",
    "organizations": $orgs_json,
    "repositories": $repos_json
  },
  "analysis": {
    "default_modules": ["vulnerability", "provenance", "package-health"],
    "output_dir": "./supply-chain-reports"
  }
}
CONFIGEOF

    echo ""
    echo -e "${GREEN}✓ Configuration saved to: $CONFIG_FILE${NC}"
    echo ""
    echo "Summary:"
    echo "  Organizations: ${#selected_orgs[@]}"
    echo "  Repositories: ${#selected_repos[@]}"
    echo ""
}

# Function to get targets from config or interactively
get_targets() {
    if [[ ${#TARGETS[@]} -gt 0 ]]; then
        # Targets provided via CLI
        return 0
    fi

    # Try to load config using hierarchical system
    if load_config "supply-chain" "$CONFIG_FILE" 2>/dev/null; then
        # Use hierarchical config loader
        local config_orgs=$(get_organizations)
        local config_repos=$(get_repositories)
    else
        # Fallback to direct config file reading
        if [[ ! -f "$CONFIG_FILE" ]]; then
            if [[ "$INTERACTIVE" == "true" ]]; then
                echo -e "${YELLOW}No config file found${NC}"
                setup_config
            else
                echo -e "${RED}Error: No config file found${NC}"
                echo "Run with --setup to create configuration, or specify targets via --org/--repo"
                exit 1
            fi
        fi

        config_orgs=$(jq -r '.github.organizations[]' "$CONFIG_FILE" 2>/dev/null || echo "")
        config_repos=$(jq -r '.github.repositories[]' "$CONFIG_FILE" 2>/dev/null || echo "")
    fi

    if [[ -z "$config_orgs" ]] && [[ -z "$config_repos" ]]; then
        if [[ "$INTERACTIVE" == "true" ]]; then
            echo -e "${YELLOW}No targets configured${NC}"
            setup_config
            # Reload after setup
            config_orgs=$(jq -r '.github.organizations[]' "$CONFIG_FILE" 2>/dev/null || echo "")
            config_repos=$(jq -r '.github.repositories[]' "$CONFIG_FILE" 2>/dev/null || echo "")
        else
            echo -e "${RED}Error: No targets configured${NC}"
            echo "Run with --setup or --interactive to configure targets"
            exit 1
        fi
    fi

    # Build targets array
    while IFS= read -r org; do
        [[ -n "$org" ]] && TARGETS+=("org:$org")
    done <<< "$config_orgs"

    while IFS= read -r repo; do
        [[ -n "$repo" ]] && TARGETS+=("repo:$repo")
    done <<< "$config_repos"
}

# Function to expand org into repos
expand_org_repos() {
    local org="$1"
    echo -e "${BLUE}Fetching repositories for org: $org${NC}" >&2

    # Check if gh is authenticated before attempting
    if ! gh auth status >/dev/null 2>&1 && [[ -z "${GH_TOKEN:-}" ]]; then
        echo -e "${RED}✗ Error: GitHub authentication required${NC}" >&2
        echo -e "${YELLOW}  Either run 'gh auth login' or set GH_TOKEN environment variable${NC}" >&2
        return 1
    fi

    local repos=$(gh repo list "$org" --limit 1000 --json nameWithOwner --jq '.[].nameWithOwner' 2>/dev/null)
    local exit_code=$?

    if [[ $exit_code -ne 0 ]] || [[ -z "$repos" ]]; then
        echo -e "${YELLOW}⚠ No repositories found for org: $org${NC}" >&2
        echo -e "${YELLOW}  Check that GH_TOKEN has read:org permissions${NC}" >&2
        return 1
    fi

    echo "$repos"
}

# Function to normalize target format
# Converts owner/repo to full GitHub URL if needed
normalize_target() {
    local target="$1"

    # If it's already a URL or a path, return as-is
    if [[ "$target" =~ ^https?:// ]] || [[ "$target" =~ ^git@ ]] || [[ -e "$target" ]]; then
        echo "$target"
        return
    fi

    # If it looks like owner/repo format, convert to GitHub URL
    if [[ "$target" =~ ^[a-zA-Z0-9_-]+/[a-zA-Z0-9_.-]+$ ]]; then
        echo "https://github.com/$target"
        return
    fi

    # Otherwise return as-is
    echo "$target"
}

# Function to extract repo name from URL
extract_repo_name() {
    local url="$1"
    if [[ "$url" =~ github\.com[/:]([^/]+/[^/.]+) ]]; then
        echo "${BASH_REMATCH[1]}"
    elif [[ "$url" =~ ^([a-zA-Z0-9_-]+/[a-zA-Z0-9_.-]+)$ ]]; then
        echo "$url"
    else
        basename "$url" .git
    fi
}

# Function to clone repository once for sharing across modules
clone_shared_repository() {
    local repo_url="$1"

    # Extract repo name for display
    SHARED_REPO_NAME=$(extract_repo_name "$repo_url")

    # Create temp directory for shared clone
    SHARED_REPO_DIR=$(mktemp -d)

    echo -e "${BLUE}Cloning ${SHARED_REPO_NAME}...${NC}"
    if git clone --depth 1 --quiet "$repo_url" "$SHARED_REPO_DIR" 2>/dev/null; then
        # Count files and calculate repo size
        local file_count=$(find "$SHARED_REPO_DIR" -type f | wc -l | tr -d ' ')
        local repo_size=$(du -sh "$SHARED_REPO_DIR" 2>/dev/null | cut -f1)
        echo -e "${GREEN}✓ Cloned ${SHARED_REPO_NAME}: ${file_count} files, ${repo_size}${NC}"

        # Generate SBOM for shared use
        echo -e "${BLUE}Generating SBOM for shared analysis...${NC}"

        # Source SBOM library
        if [[ -f "$UTILS_ROOT/lib/sbom.sh" ]]; then
            source "$UTILS_ROOT/lib/sbom.sh"

            # Create temp SBOM file
            SHARED_SBOM_FILE=$(mktemp)

            # Generate SBOM
            if generate_sbom "$SHARED_REPO_DIR" "$SHARED_SBOM_FILE" "true" 2>&1 | grep -v "^\["; then
                # Display SBOM summary - make package_count global for progress display
                SHARED_PACKAGE_COUNT=$(jq '.components | length' "$SHARED_SBOM_FILE" 2>/dev/null || echo "0")
                local sbom_format=$(jq -r '.bomFormat // "unknown"' "$SHARED_SBOM_FILE" 2>/dev/null || echo "unknown")

                # Count packages by language/type
                local package_by_lang=$(jq -r '.components[]? | .type + "/" + (.purl // "unknown" | split(":")[0])' "$SHARED_SBOM_FILE" 2>/dev/null | sort | uniq -c | sort -rn)

                echo -e "${GREEN}✓ SBOM generated: ${SHARED_PACKAGE_COUNT} components (${sbom_format} format)${NC}"

                # Display language breakdown
                if [[ -n "$package_by_lang" ]]; then
                    echo -e "${CYAN}  Package breakdown:${NC}"
                    while IFS= read -r line; do
                        local count=$(echo "$line" | awk '{print $1}')
                        local type=$(echo "$line" | awk '{print $2}')
                        echo -e "${CYAN}    - ${type}: ${count}${NC}"
                    done <<< "$package_by_lang"
                fi
            else
                echo -e "${YELLOW}⚠ SBOM generation failed, modules will generate their own${NC}"
                rm -f "$SHARED_SBOM_FILE"
                SHARED_SBOM_FILE=""
            fi
        else
            echo -e "${YELLOW}⚠ SBOM library not found, modules will generate their own${NC}"
        fi

        return 0
    else
        echo -e "${RED}✗ Failed to clone repository${NC}"
        rm -rf "$SHARED_REPO_DIR"
        SHARED_REPO_DIR=""
        return 1
    fi
}

# Function to run vulnerability analysis
run_vulnerability_analysis() {
    local target=$(normalize_target "$1")
    local analyser="$SCRIPT_DIR/vulnerability-analysis/vulnerability-analyser.sh"

    if [[ ! -f "$analyser" ]]; then
        echo -e "${RED}✗ Vulnerability analyser not found${NC}"
        return 1
    fi

    # Build command with optional flags
    local cmd="$analyser --prioritize"

    # Disable Claude in individual analyzer - we'll run it once at the end
    cmd="$cmd --no-claude"

    # Add parallel flag if enabled (uses OSV.dev batch API)
    if [[ "$PARALLEL" == "true" ]]; then
        cmd="$cmd --parallel"
    fi

    # Use shared SBOM if available for batch API mode
    if [[ -n "$SHARED_SBOM_FILE" ]] && [[ -f "$SHARED_SBOM_FILE" ]] && [[ "$PARALLEL" == "true" ]]; then
        cmd="$cmd --sbom $SHARED_SBOM_FILE"
    fi

    # Use shared repository if available
    if [[ -n "$SHARED_REPO_DIR" ]] && [[ -d "$SHARED_REPO_DIR" ]]; then
        cmd="$cmd --local-path $SHARED_REPO_DIR"
    fi

    cmd="$cmd $target"
    eval "$cmd"
}

# Function to run provenance analysis
run_provenance_analysis() {
    local target=$(normalize_target "$1")
    local analyser="$SCRIPT_DIR/provenance-analysis/provenance-analyser.sh"

    if [[ ! -f "$analyser" ]]; then
        echo -e "${RED}✗ Provenance analyser not found${NC}"
        return 1
    fi

    # Build command with optional flags
    local cmd="$analyser"

    # Disable Claude in individual analyzer - we'll run it once at the end
    cmd="$cmd --no-claude"

    # Add parallel flag if enabled
    if [[ "$PARALLEL" == "true" ]]; then
        cmd="$cmd --parallel"
    fi

    # Use shared SBOM if available
    if [[ -n "$SHARED_SBOM_FILE" ]] && [[ -f "$SHARED_SBOM_FILE" ]]; then
        cmd="$cmd --sbom-file $SHARED_SBOM_FILE"
    fi

    # Use shared repository if available
    if [[ -n "$SHARED_REPO_DIR" ]] && [[ -d "$SHARED_REPO_DIR" ]]; then
        cmd="$cmd --local-path $SHARED_REPO_DIR $target"
    else
        cmd="$cmd --repo $target"
    fi

    eval "$cmd"
}

# Function to run package health analysis
run_package_health_analysis() {
    local target=$(normalize_target "$1")
    local analyser="$SCRIPT_DIR/package-health-analysis/package-health-analyser.sh"

    if [[ ! -f "$analyser" ]]; then
        echo -e "${RED}✗ Package health analyser not found${NC}"
        return 1
    fi

    # Build command with optional flags
    local cmd="$analyser"

    # Disable Claude in individual analyzer - we'll run it once at the end
    cmd="$cmd --no-claude"

    # Add parallel flag if enabled
    if [[ "$PARALLEL" == "true" ]]; then
        cmd="$cmd --parallel"
    fi

    # Use shared SBOM if available
    if [[ -n "$SHARED_SBOM_FILE" ]] && [[ -f "$SHARED_SBOM_FILE" ]]; then
        cmd="$cmd --sbom-file $SHARED_SBOM_FILE"
    fi

    # Use shared repository if available
    if [[ -n "$SHARED_REPO_DIR" ]] && [[ -d "$SHARED_REPO_DIR" ]]; then
        # Package health analyser expects just the owner/repo format for the repo name
        # Strip the https://github.com/ prefix if present
        local repo_name="${target#https://github.com/}"
        cmd="$cmd --repo $repo_name --local-path $SHARED_REPO_DIR"
    else
        # Package health analyser expects just the owner/repo format
        # Strip the https://github.com/ prefix if present
        target="${target#https://github.com/}"
        cmd="$cmd --repo $target"
    fi

    eval "$cmd"
}

# Function to run legal compliance analysis
run_legal_analysis() {
    local target=$(normalize_target "$1")
    local analyser="$UTILS_ROOT/legal-review/legal-analyser.sh"

    if [[ ! -f "$analyser" ]]; then
        echo -e "${RED}✗ Legal analyser not found${NC}"
        return 1
    fi

    # Build command with optional flags
    local cmd="$analyser"

    # Add Claude flag if enabled
    if [[ "$USE_CLAUDE" == "true" ]]; then
        cmd="$cmd --claude"
    fi

    # Add parallel flag if enabled
    if [[ "$PARALLEL" == "true" ]]; then
        cmd="$cmd --parallel"
    fi

    # Use shared SBOM if available (for license extraction)
    if [[ -n "$SHARED_SBOM_FILE" ]] && [[ -f "$SHARED_SBOM_FILE" ]]; then
        cmd="$cmd --sbom $SHARED_SBOM_FILE"
    fi

    # Use shared repository if available
    if [[ -n "$SHARED_REPO_DIR" ]] && [[ -d "$SHARED_REPO_DIR" ]]; then
        cmd="$cmd --local-path $SHARED_REPO_DIR"
    else
        # Legal analyser expects just the owner/repo format
        local repo_name="${target#https://github.com/}"
        cmd="$cmd --repo $repo_name"
    fi

    eval "$cmd"
}

#############################################################################
# Enhanced Analysis Modules
#############################################################################

# Function to run abandoned package detection
run_abandoned_analysis() {
    local target=$(normalize_target "$1")
    local lib_file="$SCRIPT_DIR/package-health-analysis/lib/abandonment-detector.sh"

    if [[ ! -f "$lib_file" ]]; then
        echo -e "${RED}✗ Abandonment detector not found${NC}"
        return 1
    fi

    source "$lib_file"

    echo "# Abandoned Package Detection"
    echo ""

    # Get project directory
    local project_dir=""
    if [[ -n "$SHARED_REPO_DIR" ]] && [[ -d "$SHARED_REPO_DIR" ]]; then
        project_dir="$SHARED_REPO_DIR"
    else
        echo -e "${YELLOW}⚠ No local repository available. Clone the repo first.${NC}"
        return 1
    fi

    # Extract packages from manifest
    local packages_json="[]"
    if [[ -f "$project_dir/package.json" ]]; then
        local deps=$(jq -r '.dependencies // {} | keys[]' "$project_dir/package.json" 2>/dev/null)
        while IFS= read -r pkg; do
            [[ -z "$pkg" ]] && continue
            packages_json=$(echo "$packages_json" | jq --arg name "$pkg" '. + [{"name": $name, "ecosystem": "npm"}]')
        done <<< "$deps"
    fi

    if [[ "$packages_json" == "[]" ]]; then
        echo "No dependencies found to analyze."
        return 0
    fi

    # Generate report
    local report=$(generate_abandonment_report "$packages_json")
    echo "$report" | jq '.'
}

# Function to run typosquatting detection
run_typosquat_analysis() {
    local target=$(normalize_target "$1")
    local lib_file="$SCRIPT_DIR/package-health-analysis/lib/typosquat-detector.sh"

    if [[ ! -f "$lib_file" ]]; then
        echo -e "${RED}✗ Typosquat detector not found${NC}"
        return 1
    fi

    source "$lib_file"

    echo "# Typosquatting Risk Detection"
    echo ""

    # Get project directory
    local project_dir=""
    if [[ -n "$SHARED_REPO_DIR" ]] && [[ -d "$SHARED_REPO_DIR" ]]; then
        project_dir="$SHARED_REPO_DIR"
    else
        echo -e "${YELLOW}⚠ No local repository available. Clone the repo first.${NC}"
        return 1
    fi

    # Extract packages from manifest
    local packages_json="[]"
    if [[ -f "$project_dir/package.json" ]]; then
        local deps=$(jq -r '.dependencies // {} | keys[]' "$project_dir/package.json" 2>/dev/null)
        while IFS= read -r pkg; do
            [[ -z "$pkg" ]] && continue
            packages_json=$(echo "$packages_json" | jq --arg name "$pkg" '. + [{"name": $name, "ecosystem": "npm"}]')
        done <<< "$deps"
    fi

    if [[ "$packages_json" == "[]" ]]; then
        echo "No dependencies found to analyze."
        return 0
    fi

    # Generate report
    local report=$(generate_typosquat_report "$packages_json")
    echo "$report" | jq '.'
}

# Function to run unused dependency detection
run_unused_analysis() {
    local target=$(normalize_target "$1")
    local lib_file="$SCRIPT_DIR/package-health-analysis/lib/unused-detector.sh"

    if [[ ! -f "$lib_file" ]]; then
        echo -e "${RED}✗ Unused dependency detector not found${NC}"
        return 1
    fi

    source "$lib_file"

    echo "# Unused Dependency Detection"
    echo ""

    # Get project directory
    local project_dir=""
    if [[ -n "$SHARED_REPO_DIR" ]] && [[ -d "$SHARED_REPO_DIR" ]]; then
        project_dir="$SHARED_REPO_DIR"
    else
        echo -e "${YELLOW}⚠ No local repository available. Clone the repo first.${NC}"
        return 1
    fi

    # Generate report
    local report=$(generate_unused_report "$project_dir")
    echo "$report" | jq '.'
}

# Function to run technical debt scoring
run_debt_score_analysis() {
    local target=$(normalize_target "$1")
    local lib_file="$SCRIPT_DIR/bundle-analysis/lib/debt-scorer.sh"

    if [[ ! -f "$lib_file" ]]; then
        echo -e "${RED}✗ Debt scorer not found${NC}"
        return 1
    fi

    source "$lib_file"

    echo "# Technical Debt Score"
    echo ""

    # Get project directory
    local project_dir=""
    if [[ -n "$SHARED_REPO_DIR" ]] && [[ -d "$SHARED_REPO_DIR" ]]; then
        project_dir="$SHARED_REPO_DIR"
    else
        echo -e "${YELLOW}⚠ No local repository available. Clone the repo first.${NC}"
        return 1
    fi

    # Generate roadmap which includes debt scoring
    local report=$(generate_debt_roadmap "$project_dir")
    echo "$report" | jq '.'
}

# Function to run container image analysis
run_container_analysis() {
    local target=$(normalize_target "$1")
    local lib_file="$SCRIPT_DIR/container-analysis/lib/image-recommender.sh"

    if [[ ! -f "$lib_file" ]]; then
        echo -e "${RED}✗ Container image recommender not found${NC}"
        return 1
    fi

    source "$lib_file"

    echo "# Container Image Analysis"
    echo ""

    # Get project directory
    local project_dir=""
    if [[ -n "$SHARED_REPO_DIR" ]] && [[ -d "$SHARED_REPO_DIR" ]]; then
        project_dir="$SHARED_REPO_DIR"
    else
        echo -e "${YELLOW}⚠ No local repository available. Clone the repo first.${NC}"
        return 1
    fi

    # Analyze Dockerfile
    local report=$(analyze_dockerfile "$project_dir")
    echo "$report" | jq '.'
}

# Function to run library recommendation analysis
run_library_recommendation_analysis() {
    local target=$(normalize_target "$1")
    local lib_file="$SCRIPT_DIR/library-recommendations/lib/recommender.sh"

    if [[ ! -f "$lib_file" ]]; then
        echo -e "${RED}✗ Library recommender not found${NC}"
        return 1
    fi

    source "$lib_file"

    echo "# Library Recommendations"
    echo ""

    # Get project directory
    local project_dir=""
    if [[ -n "$SHARED_REPO_DIR" ]] && [[ -d "$SHARED_REPO_DIR" ]]; then
        project_dir="$SHARED_REPO_DIR"
    else
        echo -e "${YELLOW}⚠ No local repository available. Clone the repo first.${NC}"
        return 1
    fi

    # Generate migration plan
    local report=$(generate_migration_plan "$project_dir")
    echo "$report" | jq '.'
}

#############################################################################
# Unified Claude AI Analysis
#############################################################################

run_unified_claude_analysis() {
    local all_data="$1"
    local target="$2"
    local model="claude-sonnet-4-20250514"

    echo -e "${BLUE}Analyzing with Claude AI...${NC}"

    local repo_root="$(cd "$SCRIPT_DIR/.." && pwd)"
    local rag_dir="$repo_root/rag/supply-chain"

    # Source universal persona loader if persona is set
    local persona_context=""
    if [[ -n "$PERSONA" ]]; then
        local universal_loader="$REPO_ROOT/lib/universal-persona-loader.sh"

        if [[ -f "$universal_loader" ]]; then
            # Use universal chain-of-reasoning personas (from rag/personas/)
            source "$universal_loader"
            echo -e "${CYAN}Loading persona context: $(get_universal_persona_display_name "$PERSONA") (Chain of Reasoning)${NC}"
            # Note: We'll build the full prompt with build_persona_prompt() below
        else
            echo -e "${YELLOW}Warning: Universal persona loader not found at $universal_loader${NC}"
        fi
    fi

    # Load base RAG knowledge for comprehensive supply chain security
    local rag_context=""

    # Load all relevant base RAG documents
    if [[ -f "$rag_dir/sbom-generation-best-practices.md" ]]; then
        rag_context+="# SBOM Best Practices\n\n"
        rag_context+=$(head -150 "$rag_dir/sbom-generation-best-practices.md" | tail -n +1)
        rag_context+="\n\n"
    fi

    if [[ -f "$rag_dir/slsa/slsa-specification.md" ]]; then
        rag_context+="# SLSA Specification\n\n"
        rag_context+=$(head -150 "$rag_dir/slsa/slsa-specification.md" | tail -n +1)
        rag_context+="\n\n"
    fi

    if [[ -f "$rag_dir/package-health/package-management-best-practices.md" ]]; then
        rag_context+="# Package Management Best Practices\n\n"
        rag_context+=$(head -150 "$rag_dir/package-health/package-management-best-practices.md" | tail -n +1)
        rag_context+="\n\n"
    fi

    # Build the prompt - use universal chain-of-reasoning personas
    local prompt
    if [[ -n "$PERSONA" ]] && command -v build_persona_prompt &> /dev/null; then
        # Use universal chain-of-reasoning personas (from rag/personas/)
        # Add repository context to the scan data
        local scan_data_with_context="# Target Repository
Repository: $target

# Scan Results
$all_data"

        prompt=$(build_persona_prompt "$PERSONA" "$rag_context" "$scan_data_with_context" "Supply Chain Scanner")
    else
        # Default prompt (no persona selected)
        prompt="You are a supply chain security expert analyzing MULTIPLE security scans for repository: $target

# Supply Chain Security Knowledge Base
$rag_context

# Your Task
Analyze ALL the scan results together to provide a HOLISTIC, PRIORITY-BASED ACTION REPORT.

**CRITICAL**: Correlate findings across scans. For example:
- A package with vulnerabilities + poor health + no provenance = CRITICAL PRIORITY
- A package with only poor health but good provenance = LOWER PRIORITY
- Multiple issues on same package = consolidate into single high-priority action

# Analysis Requirements

For each finding, you MUST:
1. **Explain the issue** - What's wrong across all dimensions (vulnerability + health + provenance)?
2. **Correlate findings** - How do issues compound? (e.g., \"lodash has CVE-2021-23337 AND health score of 45 AND no SLSA provenance\")
3. **Justify priority** - Why this specific combination makes it critical/high/medium/low
4. **Reference knowledge base** - Cite relevant best practices
5. **Provide consolidated actions** - Single action plan addressing all issues for that package

# Output Format

## 🔴 CRITICAL PRIORITY (Immediate - 0-24 hours)

For each critical issue, provide:

**Package**: [name@version]

**Combined Issues**:
- 🚨 Vulnerability: [CVE details, CVSS score, KEV status]
- 💊 Health: [Score, specific health concerns]
- 🔒 Provenance: [SLSA level, attestation status]

**Why Critical**: [Explain how these issues compound - e.g., \"This package is exploitable (KEV listed), unmaintained (health score 35), and has no verifiable build provenance (SLSA 0). This represents a complete supply chain failure.\"]

**Best Practice References**: [Cite from knowledge base]

**Consolidated Action Plan**:
\`\`\`bash
# Step 1: Immediate mitigation
<specific commands>

# Step 2: Upgrade path
<specific version/alternative>

# Step 3: Verification
<how to verify fix addresses all issues>
\`\`\`

**Timeline**: Immediate (0-24h)

---

## 🟠 HIGH PRIORITY (Urgent - 1-7 days)
[Same structured format]

## 🟡 MEDIUM PRIORITY (Important - 1-30 days)
[Same structured format]

## 🟢 LOW PRIORITY (Monitor - 30+ days)
[Same structured format]

## 📊 Holistic Assessment & Strategic Recommendations

### Overall Supply Chain Posture
- Vulnerability Management: [assessment across all packages]
- Package Health: [overall health score distribution]
- Provenance Maturity: [SLSA level distribution]
- **Combined Risk Score**: [Critical/High/Medium/Low]

### Systemic Improvements
Based on patterns across ALL scans:
1. [e.g., \"80% of packages lack provenance - implement GitHub attestations org-wide\"]
2. [e.g., \"3 critical vulnerabilities in unmaintained packages - establish deprecation policy\"]
3. [e.g., \"Version inconsistencies across 12 packages - run npm dedupe\"]

### Automation Opportunities
- CI/CD Integration: [specific tools based on findings]
- Dependency Management: [Dependabot, Renovate, etc.]
- SLSA Compliance: [path to Level 3+]

### Effort Estimation
- Immediate fixes: [X hours]
- Short-term improvements: [X days]
- Long-term hardening: [X weeks]

# All Scan Results:

$all_data"
    fi

    local response=$(curl -s https://api.anthropic.com/v1/messages \
        -H "content-type: application/json" \
        -H "x-api-key: $ANTHROPIC_API_KEY" \
        -H "anthropic-version: 2023-06-01" \
        -d "{
            \"model\": \"$model\",
            \"max_tokens\": 8192,
            \"messages\": [{
                \"role\": \"user\",
                \"content\": $(echo "$prompt" | jq -Rs .)
            }]
        }")

    # Load cost tracking if available
    if [ -f "$repo_root/lib/claude-cost.sh" ]; then
        source "$repo_root/lib/claude-cost.sh"
        if command -v record_api_usage &> /dev/null; then
            record_api_usage "$response" "$model" > /dev/null
        fi
    fi

    # Extract and display analysis
    local analysis=$(echo "$response" | jq -r '.content[0].text // empty')

    if [[ -n "$analysis" ]]; then
        echo "$analysis"

        # Display cost summary if available
        if command -v display_api_cost_summary &> /dev/null; then
            echo ""
            display_api_cost_summary
        fi
    else
        echo -e "${RED}Error: No analysis returned from Claude API${NC}"
        echo "Response: $response"
    fi
}

# Function to display summary of all analysis results
display_analysis_summary() {
    local target="$1"
    local repo_files="$2"
    local repo_size="$3"
    local sbom_packages="$4"
    local vuln_total="$5"
    local vuln_critical="$6"
    local vuln_high="$7"
    local vuln_medium="$8"
    local vuln_low="$9"
    local vuln_kev="${10}"
    local pkg_deprecated="${11}"
    local pkg_low_health="${12}"
    local pkg_total="${13}"

    echo ""
    echo -e "${GREEN}=========================================${NC}"
    echo -e "${GREEN}  Analysis Summary${NC}"
    echo -e "${GREEN}=========================================${NC}"
    echo ""

    # Repository Information
    if [[ -n "$repo_files" ]] && [[ "$repo_files" != "0" ]]; then
        echo -e "${CYAN}Repository:${NC} $target"
        echo -e "  Files: ${repo_files}"
        echo -e "  Size: ${repo_size}"
        echo ""
    fi

    # SBOM Information
    if [[ -n "$sbom_packages" ]] && [[ "$sbom_packages" != "0" ]]; then
        echo -e "${CYAN}SBOM Generation:${NC}"
        echo -e "  Packages detected: ${sbom_packages}"
        echo ""
    fi

    # Vulnerability Information
    if [[ -n "$vuln_total" ]] && [[ "$vuln_total" != "0" ]]; then
        echo -e "${CYAN}Vulnerability Analysis:${NC}"
        echo -e "  Total vulnerabilities: ${vuln_total}"
        if [[ -n "$vuln_critical" ]] && [[ "$vuln_critical" != "0" ]]; then
            echo -e "  ${RED}⚠ Critical: ${vuln_critical}${NC}"
        fi
        if [[ -n "$vuln_high" ]] && [[ "$vuln_high" != "0" ]]; then
            echo -e "  ${YELLOW}⚠ High: ${vuln_high}${NC}"
        fi
        if [[ -n "$vuln_medium" ]]; then
            echo -e "  Medium: ${vuln_medium}"
        fi
        if [[ -n "$vuln_low" ]]; then
            echo -e "  Low: ${vuln_low}"
        fi
        if [[ -n "$vuln_kev" ]] && [[ "$vuln_kev" != "0" ]]; then
            echo -e "  ${RED}⚠ In CISA KEV: ${vuln_kev}${NC}"
        fi
        echo ""
    elif [[ "$vuln_total" == "0" ]]; then
        echo -e "${CYAN}Vulnerability Analysis:${NC}"
        echo -e "  ${GREEN}✓ No vulnerabilities found${NC}"
        echo ""
    fi

    # Package Health Information
    if [[ -n "$pkg_total" ]] && [[ "$pkg_total" != "0" ]]; then
        echo -e "${CYAN}Package Health Analysis:${NC}"
        echo -e "  Packages analyzed: ${pkg_total}"
        if [[ -n "$pkg_deprecated" ]] && [[ "$pkg_deprecated" != "0" ]]; then
            echo -e "  ${YELLOW}⚠ Deprecated: ${pkg_deprecated}${NC}"
        fi
        if [[ -n "$pkg_low_health" ]] && [[ "$pkg_low_health" != "0" ]]; then
            echo -e "  ${YELLOW}⚠ Low health score: ${pkg_low_health}${NC}"
        fi
        echo ""
    fi

    echo -e "${GREEN}=========================================${NC}"
    echo ""
}

# Function to run analysis on target
analyze_target() {
    local target="$1"

    # Initialize summary metrics
    local summary_repo_files="0"
    local summary_repo_size=""
    local summary_sbom_packages="0"
    local summary_vuln_total="0"
    local summary_vuln_critical="0"
    local summary_vuln_high="0"
    local summary_vuln_medium="0"
    local summary_vuln_low="0"
    local summary_vuln_kev="0"
    local summary_pkg_deprecated="0"
    local summary_pkg_low_health="0"
    local summary_pkg_total="0"

    echo ""
    echo -e "${CYAN}=========================================${NC}"
    echo -e "${CYAN}Analyzing: $target${NC}"
    if [[ "$USE_CLAUDE" == "true" ]] && [[ -n "$ANTHROPIC_API_KEY" ]]; then
        echo -e "${GREEN}Claude AI: ENABLED (will run after all scans)${NC}"
    elif [[ -z "$ANTHROPIC_API_KEY" ]]; then
        echo -e "${YELLOW}Claude AI: DISABLED (no API key)${NC}"
    fi
    if [[ "$PARALLEL" == "true" ]]; then
        echo -e "${CYAN}Parallel Mode: ENABLED${NC}"
    fi
    # Warn about slow modules
    for mod in "${MODULES[@]}"; do
        case "$mod" in
            provenance)
                echo -e "${YELLOW}⚠ Provenance analysis: queries npm registry per-package (may be slow)${NC}"
                ;;
        esac
    done
    echo -e "${CYAN}=========================================${NC}"
    echo ""

    # Clone repository for modules that need local access
    # Clone if: multiple modules, OR any of the new enhanced analysis modules
    local needs_clone=false
    if [[ ${#MODULES[@]} -gt 1 ]]; then
        needs_clone=true
    else
        for mod in "${MODULES[@]}"; do
            case "$mod" in
                abandoned|typosquat|unused|debt-score|container-images|library-recommend)
                    needs_clone=true
                    break
                    ;;
            esac
        done
    fi

    if [[ "$needs_clone" == "true" ]]; then
        local normalized=$(normalize_target "$target")
        # Only clone if it's a git URL (not a local directory or file)
        if [[ "$normalized" =~ ^https?:// ]] || [[ "$normalized" =~ ^git@ ]]; then
            clone_shared_repository "$normalized"
        fi
    fi

    # Capture all analysis outputs for unified Claude analysis
    local all_results=""

    for module in "${MODULES[@]}"; do
        echo ""
        # Show module header with package count and mode
        local mode_info=""
        if [[ "$PARALLEL" == "true" ]]; then
            mode_info=" (batch mode)"
        fi
        local pkg_info=""
        if [[ -n "$SHARED_PACKAGE_COUNT" ]] && [[ "$SHARED_PACKAGE_COUNT" -gt 0 ]]; then
            pkg_info=" - ${SHARED_PACKAGE_COUNT} packages"
        fi
        local repo_display=""
        if [[ -n "$SHARED_REPO_NAME" ]]; then
            repo_display=" on ${SHARED_REPO_NAME}"
        fi
        echo -e "${BLUE}▶ Running ${module} analysis${repo_display}${pkg_info}${mode_info}${NC}"

        # Run analysis - stderr (progress) goes to terminal, stdout captured
        # Use a temp file to capture output while allowing stderr to display
        local temp_output=$(mktemp)
        local module_output=""
        case "$module" in
            vulnerability)
                run_vulnerability_analysis "$target" 2>&1 | tee "$temp_output"
                module_output=$(cat "$temp_output")
                ;;
            provenance)
                run_provenance_analysis "$target" 2>&1 | tee "$temp_output"
                module_output=$(cat "$temp_output")
                ;;
            package-health)
                run_package_health_analysis "$target" 2>&1 | tee "$temp_output"
                module_output=$(cat "$temp_output")
                ;;
            legal)
                run_legal_analysis "$target" 2>&1 | tee "$temp_output"
                module_output=$(cat "$temp_output")
                ;;
            abandoned)
                run_abandoned_analysis "$target" 2>&1 | tee "$temp_output"
                module_output=$(cat "$temp_output")
                ;;
            typosquat)
                run_typosquat_analysis "$target" 2>&1 | tee "$temp_output"
                module_output=$(cat "$temp_output")
                ;;
            unused)
                run_unused_analysis "$target" 2>&1 | tee "$temp_output"
                module_output=$(cat "$temp_output")
                ;;
            debt-score)
                run_debt_score_analysis "$target" 2>&1 | tee "$temp_output"
                module_output=$(cat "$temp_output")
                ;;
            container-images)
                run_container_analysis "$target" 2>&1 | tee "$temp_output"
                module_output=$(cat "$temp_output")
                ;;
            library-recommend)
                run_library_recommendation_analysis "$target" 2>&1 | tee "$temp_output"
                module_output=$(cat "$temp_output")
                ;;
            *)
                echo -e "${YELLOW}⚠ Unknown module: $module${NC}"
                ;;
        esac

        # Clean up temp file
        rm -f "$temp_output"

        # Extract metrics from module output for summary
        case "$module" in
            vulnerability)
                # Extract vulnerability metrics - strip markdown formatting and extract numbers
                summary_vuln_total=$(echo "$module_output" | grep -i "Total vulnerabilities:" | sed -E 's/.*Total vulnerabilities:[*[:space:]]*([0-9]+).*/\1/' | head -1)
                summary_vuln_total=${summary_vuln_total:-0}
                # For severity levels, match lines with just severity label and number (skip table rows with |)
                summary_vuln_critical=$(echo "$module_output" | grep "Critical:" | grep -v "|" | sed -E 's/.*Critical:[*[:space:]]*([0-9]+).*/\1/' | head -1)
                summary_vuln_critical=${summary_vuln_critical:-0}
                summary_vuln_high=$(echo "$module_output" | grep "High:" | grep -v "|" | sed -E 's/.*High:[*[:space:]]*([0-9]+).*/\1/' | head -1)
                summary_vuln_high=${summary_vuln_high:-0}
                summary_vuln_medium=$(echo "$module_output" | grep "Medium:" | grep -v "|" | sed -E 's/.*Medium:[*[:space:]]*([0-9]+).*/\1/' | head -1)
                summary_vuln_medium=${summary_vuln_medium:-0}
                summary_vuln_low=$(echo "$module_output" | grep "Low:" | grep -v "|" | sed -E 's/.*Low:[*[:space:]]*([0-9]+).*/\1/' | head -1)
                summary_vuln_low=${summary_vuln_low:-0}
                summary_vuln_kev=$(echo "$module_output" | grep "In CISA KEV:" | sed -E 's/.*In CISA KEV:[*[:space:]]*([0-9]+).*/\1/' | head -1)
                summary_vuln_kev=${summary_vuln_kev:-0}
                ;;
            package-health)
                # Extract package health metrics
                summary_pkg_total=$(echo "$module_output" | grep -i "Total Packages:" | sed -E 's/.*Total Packages:[*[:space:]]*([0-9]+).*/\1/' | head -1)
                summary_pkg_total=${summary_pkg_total:-0}
                summary_pkg_deprecated=$(echo "$module_output" | grep -iE "Deprecated( Packages)?:" | sed -E 's/.*Deprecated( Packages)?:[*[:space:]]*([0-9]+).*/\2/' | head -1)
                summary_pkg_deprecated=${summary_pkg_deprecated:-0}
                summary_pkg_low_health=$(echo "$module_output" | grep -iE "Low Health( Packages)?:" | sed -E 's/.*Low Health( Packages)?:[*[:space:]]*([0-9]+).*/\2/' | head -1)
                summary_pkg_low_health=${summary_pkg_low_health:-0}
                ;;
        esac

        # Append to combined results for Claude
        if [[ -n "$module_output" ]]; then
            all_results+="
========================================
${module} Analysis Results
========================================
$module_output

"
        fi
    done

    # Extract repository and SBOM metrics if shared repo was created
    if [[ -n "$SHARED_REPO_DIR" ]] && [[ -d "$SHARED_REPO_DIR" ]]; then
        summary_repo_files=$(find "$SHARED_REPO_DIR" -type f 2>/dev/null | wc -l | tr -d ' ')
        summary_repo_size=$(du -sh "$SHARED_REPO_DIR" 2>/dev/null | cut -f1)
    fi
    if [[ -n "$SHARED_SBOM_FILE" ]] && [[ -f "$SHARED_SBOM_FILE" ]]; then
        summary_sbom_packages=$(jq '.components | length' "$SHARED_SBOM_FILE" 2>/dev/null || echo "0")
    fi

    # Run unified Claude analysis on ALL results
    if [[ "$USE_CLAUDE" == "true" ]] && [[ -n "$ANTHROPIC_API_KEY" ]] && [[ -n "$all_results" ]]; then
        if [[ "$PERSONA" == "all" ]]; then
            # Generate reports for all personas
            local all_persona_list=("security-engineer" "software-engineer" "engineering-leader" "auditor")
            local total_personas=${#all_persona_list[@]}
            local current_num=0

            echo ""
            echo -e "${GREEN}=========================================${NC}"
            echo -e "${GREEN}  🤖 Claude AI Multi-Persona Analysis${NC}"
            echo -e "${GREEN}  Generating ${total_personas} persona reports${NC}"
            echo -e "${GREEN}=========================================${NC}"
            echo ""

            for persona_item in "${all_persona_list[@]}"; do
                current_num=$((current_num + 1))
                echo ""
                echo -e "${CYAN}─────────────────────────────────────────${NC}"
                echo -e "${CYAN}  Report ${current_num}/${total_personas}: $(echo "$persona_item" | sed 's/-/ /g' | awk '{for(i=1;i<=NF;i++)sub(/./,toupper(substr($i,1,1)),$i)}1')${NC}"
                echo -e "${CYAN}─────────────────────────────────────────${NC}"
                echo ""

                # Temporarily set PERSONA for this analysis
                PERSONA="$persona_item"
                run_unified_claude_analysis "$all_results" "$target"
            done

            # Reset PERSONA
            PERSONA="all"
        else
            echo ""
            echo -e "${GREEN}=========================================${NC}"
            echo -e "${GREEN}  🤖 Claude AI Unified Analysis${NC}"
            echo -e "${GREEN}  Analyzing ALL scan results together${NC}"
            echo -e "${GREEN}=========================================${NC}"
            echo ""

            run_unified_claude_analysis "$all_results" "$target"
        fi
    fi

    # Display summary of all analysis results
    display_analysis_summary "$target" "$summary_repo_files" "$summary_repo_size" \
        "$summary_sbom_packages" "$summary_vuln_total" "$summary_vuln_critical" \
        "$summary_vuln_high" "$summary_vuln_medium" "$summary_vuln_low" \
        "$summary_vuln_kev" "$summary_pkg_deprecated" "$summary_pkg_low_health" \
        "$summary_pkg_total"

    # Clean up shared repository after all modules complete
    if [[ -n "$SHARED_REPO_DIR" ]] && [[ -d "$SHARED_REPO_DIR" ]]; then
        cleanup_shared_repo
        SHARED_REPO_DIR=""  # Reset for next target
        SHARED_SBOM_FILE=""  # Reset for next target
    fi
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --setup)
            SETUP_MODE=true
            shift
            ;;
        -i|--interactive)
            INTERACTIVE=true
            shift
            ;;
        -v|--vulnerability)
            MODULES+=("vulnerability")
            shift
            ;;
        -p|--provenance)
            MODULES+=("provenance")
            shift
            ;;
        --package-health)
            MODULES+=("package-health")
            shift
            ;;
        --legal)
            MODULES+=("legal")
            shift
            ;;
        --abandoned)
            MODULES+=("abandoned")
            shift
            ;;
        --typosquat)
            MODULES+=("typosquat")
            shift
            ;;
        --unused)
            MODULES+=("unused")
            shift
            ;;
        --debt-score)
            MODULES+=("debt-score")
            shift
            ;;
        --container-images)
            MODULES+=("container-images")
            shift
            ;;
        --library-recommend)
            MODULES+=("library-recommend")
            shift
            ;;
        -a|--all)
            MODULES=("vulnerability" "provenance" "package-health" "abandoned" "typosquat" "unused" "debt-score")
            shift
            ;;
        --org)
            TARGETS+=("org:$2")
            shift 2
            ;;
        --repo)
            TARGETS+=("repo:$2")
            shift 2
            ;;
        -o|--output)
            OUTPUT_DIR="$2"
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
        --parallel)
            PARALLEL=true
            shift
            ;;
        --persona)
            PERSONA="$2"
            # Validate persona
            if [[ ! " ${VALID_PERSONAS[*]} " =~ " ${PERSONA} " ]]; then
                echo -e "${RED}Error: Invalid persona '${PERSONA}'${NC}"
                echo "Valid personas: ${VALID_PERSONAS[*]}"
                exit 1
            fi
            shift 2
            ;;
        --list-personas)
            echo ""
            echo "Available Personas for Supply Chain Analysis"
            echo "============================================"
            echo ""
            echo "  security-engineer     Technical vulnerability analysis"
            echo "                        Focus: CVEs, CISA KEV, remediation priorities"
            echo "                        Output: Markdown + JSON for engineering handoff"
            echo ""
            echo "  software-engineer     Developer-focused dependency analysis"
            echo "                        Focus: Upgrade paths, breaking changes, build optimization"
            echo "                        Output: CLI commands, version tables, migration guides"
            echo ""
            echo "  engineering-leader    Executive strategic overview"
            echo "                        Focus: Portfolio health, cost savings, team efficiency"
            echo "                        Output: Executive summary report, metrics dashboard"
            echo ""
            echo "  auditor               Compliance and risk assessment"
            echo "                        Focus: Framework mappings (NIST, SOC2), evidence collection"
            echo "                        Output: Audit report with findings and evidence inventory"
            echo ""
            echo "  all                   Generate ALL persona reports"
            echo "                        Runs analysis 4 times, once per persona"
            echo "                        Output: Separate report files for each persona"
            echo ""
            echo "Usage: $0 --persona <persona-name> [other options]"
            echo ""
            exit 0
            ;;
        -h|--help)
            usage
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            usage
            ;;
    esac
done

# Main execution
check_prerequisites
check_api_key || true  # Don't exit if API key not set, just inform user

# Setup mode
if [[ "$SETUP_MODE" == "true" ]]; then
    setup_config
    exit 0
fi

# Validate modules - default to --all if none specified
if [[ ${#MODULES[@]} -eq 0 ]]; then
    # Try to load default modules from config
    if load_config "supply-chain" "$CONFIG_FILE" 2>/dev/null; then
        default_modules=$(get_default_modules)
        if [[ -n "$default_modules" ]]; then
            echo -e "${BLUE}Using default modules from config${NC}"
            while IFS= read -r mod; do
                [[ -n "$mod" ]] && MODULES+=("$mod")
            done <<< "$default_modules"
        fi
    fi

    # If still no modules, default to core modules (excludes slow provenance)
    if [[ ${#MODULES[@]} -eq 0 ]]; then
        echo -e "${BLUE}No modules specified, running core modules${NC}"
        echo -e "${CYAN}  (use --all or -p to include provenance analysis)${NC}"
        MODULES=("vulnerability" "package-health")
    fi
fi

# Get targets
get_targets

if [[ ${#TARGETS[@]} -eq 0 ]]; then
    echo -e "${RED}Error: No targets to analyze${NC}"
    exit 1
fi

echo "Analysis modules: ${MODULES[*]}"
echo "Targets: ${#TARGETS[@]}"
echo ""

# Default to all personas when Claude enabled but no persona specified
if [[ "$USE_CLAUDE" == "true" ]] && [[ -z "$PERSONA" ]]; then
    PERSONA="all"
    echo -e "${CYAN}No persona specified - generating reports for ALL personas${NC}"
    echo -e "${CYAN}  Use --persona <name> to generate a single persona report${NC}"
    echo ""
fi

# Process each target
for target in "${TARGETS[@]}"; do
    if [[ "$target" =~ ^org: ]]; then
        # Expand organization to repositories
        org="${target#org:}"

        # Parse org name from GitHub URL if needed
        if [[ "$org" =~ github\.com/([^/]+) ]]; then
            org="${BASH_REMATCH[1]}"
        fi

        repos=$(expand_org_repos "$org")

        while IFS= read -r repo; do
            [[ -n "$repo" ]] && analyze_target "$repo"
        done <<< "$repos"
    elif [[ "$target" =~ ^repo: ]]; then
        # Direct repository
        repo="${target#repo:}"
        analyze_target "$repo"
    else
        # Assume it's a repository URL or path
        analyze_target "$target"
    fi
done
