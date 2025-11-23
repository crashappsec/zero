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
CONFIG_FILE="$SCRIPT_DIR/config.json"
CONFIG_EXAMPLE="$SCRIPT_DIR/config.example.json"

# Load config library
if [[ -f "$UTILS_ROOT/lib/config-loader.sh" ]]; then
    source "$UTILS_ROOT/lib/config-loader.sh"
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
USE_CLAUDE=false
ANTHROPIC_API_KEY="${ANTHROPIC_API_KEY:-}"
PARALLEL=false

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
    --provenance, -p        Run provenance analysis (SLSA)
    --package-health        Run package health analysis
    --all, -a               Run all analysis modules (default if none specified)

TARGETS:
    --org ORG_NAME          Scan all repos in GitHub organization
    --repo OWNER/REPO       Scan specific repository
    (If no targets specified, uses config.json)

OPTIONS:
    --output DIR, -o DIR    Output directory for reports
    --config FILE           Use alternate config file
    --claude                Use Claude AI for enhanced analysis (requires ANTHROPIC_API_KEY)
    --parallel              Enable batch API processing (faster, recommended)
    -h, --help              Show this help message

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
    "default_modules": ["vulnerability"],
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
    "default_modules": ["vulnerability"],
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
    echo -e "${BLUE}Fetching repositories for org: $org${NC}"

    local repos=$(gh repo list "$org" --limit 1000 --json nameWithOwner --jq '.[].nameWithOwner' 2>/dev/null || echo "")

    if [[ -z "$repos" ]]; then
        echo -e "${YELLOW}⚠ No repositories found for org: $org${NC}"
        return
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

# Function to clone repository once for sharing across modules
clone_shared_repository() {
    local repo_url="$1"

    # Only clone if running multiple modules
    if [[ ${#MODULES[@]} -le 1 ]]; then
        return 0
    fi

    # Create temp directory for shared clone
    SHARED_REPO_DIR=$(mktemp -d)

    echo -e "${BLUE}Cloning repository for shared analysis...${NC}"
    if git clone --depth 1 --quiet "$repo_url" "$SHARED_REPO_DIR" 2>/dev/null; then
        # Count files and calculate repo size
        local file_count=$(find "$SHARED_REPO_DIR" -type f | wc -l | tr -d ' ')
        local repo_size=$(du -sh "$SHARED_REPO_DIR" 2>/dev/null | cut -f1)
        echo -e "${GREEN}✓ Repository cloned: ${file_count} files, ${repo_size}${NC}"

        # Generate SBOM for shared use
        echo -e "${BLUE}Generating SBOM for shared analysis...${NC}"

        # Source SBOM library
        if [[ -f "$UTILS_ROOT/lib/sbom.sh" ]]; then
            source "$UTILS_ROOT/lib/sbom.sh"

            # Create temp SBOM file
            SHARED_SBOM_FILE=$(mktemp)

            # Generate SBOM
            if generate_sbom "$SHARED_REPO_DIR" "$SHARED_SBOM_FILE" "true" 2>&1 | grep -v "^\["; then
                # Display SBOM summary
                local package_count=$(jq '.components | length' "$SHARED_SBOM_FILE" 2>/dev/null || echo "0")
                local sbom_format=$(jq -r '.bomFormat // "unknown"' "$SHARED_SBOM_FILE" 2>/dev/null || echo "unknown")

                # Count packages by language/type
                local package_by_lang=$(jq -r '.components[]? | .type + "/" + (.purl // "unknown" | split(":")[0])' "$SHARED_SBOM_FILE" 2>/dev/null | sort | uniq -c | sort -rn)

                echo -e "${GREEN}✓ SBOM generated: ${package_count} components (${sbom_format} format)${NC}"

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

    # Add Claude flag if enabled
    if [[ "$USE_CLAUDE" == "true" ]]; then
        cmd="$cmd --claude"
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

    # Add Claude flag if enabled
    if [[ "$USE_CLAUDE" == "true" ]]; then
        cmd="$cmd --claude"
    fi

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

    # Add Claude flag if enabled
    if [[ "$USE_CLAUDE" == "true" ]]; then
        cmd="$cmd --claude"
    fi

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

# Function to run analysis on target
analyze_target() {
    local target="$1"

    echo ""
    echo -e "${CYAN}=========================================${NC}"
    echo -e "${CYAN}Analyzing: $target${NC}"
    echo -e "${CYAN}=========================================${NC}"
    echo ""

    # If running multiple modules, clone repository once for sharing
    if [[ ${#MODULES[@]} -gt 1 ]]; then
        local normalized=$(normalize_target "$target")
        # Only clone if it's a git URL (not a local directory or file)
        if [[ "$normalized" =~ ^https?:// ]] || [[ "$normalized" =~ ^git@ ]]; then
            clone_shared_repository "$normalized"
        fi
    fi

    for module in "${MODULES[@]}"; do
        echo ""
        case "$module" in
            vulnerability)
                run_vulnerability_analysis "$target"
                ;;
            provenance)
                run_provenance_analysis "$target"
                ;;
            package-health)
                run_package_health_analysis "$target"
                ;;
            *)
                echo -e "${YELLOW}⚠ Unknown module: $module${NC}"
                ;;
        esac
    done

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
        -a|--all)
            MODULES=("vulnerability" "provenance" "package-health")
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
echo ""
echo -e "${CYAN}=========================================${NC}"
echo -e "${CYAN}  Supply Chain Scanner${NC}"
echo -e "${CYAN}=========================================${NC}"
echo ""

check_prerequisites

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

    # If still no modules, default to all
    if [[ ${#MODULES[@]} -eq 0 ]]; then
        echo -e "${BLUE}No modules specified, running all modules${NC}"
        MODULES=("vulnerability" "provenance" "package-health")
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

# Process each target
for target in "${TARGETS[@]}"; do
    if [[ "$target" =~ ^org: ]]; then
        # Expand organization to repositories
        org="${target#org:}"
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

echo ""
echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}  Analysis Complete${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""
