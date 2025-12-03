#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Phantom - Unified CLI for repository analysis
#
# Usage:
#   ./phantom.sh                    # Interactive mode
#   ./phantom.sh setup              # Install tools and configure
#   ./phantom.sh check              # Verify tools and API keys
#   ./phantom.sh hydrate <repo>     # Analyze a single repository
#   ./phantom.sh hydrate --org <n>  # Analyze all repos in an org
#   ./phantom.sh status             # Show hydrated projects
#   ./phantom.sh clean              # Remove all analysis data
#############################################################################

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
UTILS_ROOT="$(dirname "$SCRIPT_DIR")"
REPO_ROOT="$(dirname "$UTILS_ROOT")"

# Load Gibson library
source "$SCRIPT_DIR/lib/gibson.sh"

# Load .env if available
if [[ -f "$REPO_ROOT/.env" ]]; then
    set -a
    source "$REPO_ROOT/.env"
    set +a
fi

#############################################################################
# Setup Functions (from setup.sh)
#############################################################################

setup_make_scripts_executable() {
    echo -e "${BLUE}Making scripts executable...${NC}"
    local count=0
    while IFS= read -r -d '' script; do
        chmod +x "$script"
        ((count++))
    done < <(find "$REPO_ROOT/utils" -type f -name "*.sh" -print0 2>/dev/null)
    echo -e "${GREEN}✓${NC} Made $count scripts executable"
}

setup_check_homebrew() {
    if [[ "$OSTYPE" != "darwin"* ]]; then
        return 0
    fi

    if command -v brew &> /dev/null; then
        echo -e "${GREEN}✓${NC} Homebrew installed"
        return 0
    fi

    echo -e "${YELLOW}○${NC} Homebrew not installed"
    echo
    read -p "Install Homebrew? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    fi
}

setup_install_tools() {
    local tools_to_install=()

    echo -e "\n${BOLD}Checking required tools...${NC}"

    # Required tools
    for tool in git jq curl; do
        if command -v "$tool" &> /dev/null; then
            echo -e "  ${GREEN}✓${NC} $tool"
        else
            echo -e "  ${RED}✗${NC} $tool (required)"
            tools_to_install+=("$tool")
        fi
    done

    echo -e "\n${BOLD}Checking recommended tools...${NC}"

    # Recommended tools (brew installable)
    for tool in osv-scanner syft gh cloc bc; do
        if command -v "$tool" &> /dev/null; then
            echo -e "  ${GREEN}✓${NC} $tool"
        else
            echo -e "  ${YELLOW}○${NC} $tool (recommended)"
            tools_to_install+=("$tool")
        fi
    done

    # jscpd (npm package) - track separately for npm install
    local need_jscpd=false
    if command -v jscpd &> /dev/null; then
        echo -e "  ${GREEN}✓${NC} jscpd"
    else
        echo -e "  ${YELLOW}○${NC} jscpd (recommended for duplicate detection)"
        need_jscpd=true
    fi

    # Checkov (pip installable)
    if command -v checkov &> /dev/null || [[ -x "$HOME/Library/Python/3.9/bin/checkov" ]] || [[ -x "$HOME/.local/bin/checkov" ]]; then
        echo -e "  ${GREEN}✓${NC} checkov"
    else
        echo -e "  ${YELLOW}○${NC} checkov (recommended for IaC security)"
        echo -e "     Install: ${CYAN}pip3 install checkov${NC} or ${CYAN}brew install checkov${NC}"
    fi

    echo -e "\n${BOLD}Checking optional tools...${NC}"

    # Python3 check
    if command -v python3 &> /dev/null; then
        local py_ver=$(python3 --version 2>&1 | grep -oE '[0-9]+\.[0-9]+')
        echo -e "  ${GREEN}✓${NC} python3 ($py_ver)"
    else
        echo -e "  ${YELLOW}○${NC} python3 (optional, for visual effects)"
    fi

    # Terminal text effects (for animated banner)
    if python3 -c "import terminaltexteffects" 2>/dev/null; then
        echo -e "  ${GREEN}✓${NC} terminaltexteffects (visual effects)"
    else
        echo -e "  ${YELLOW}○${NC} terminaltexteffects (optional visual effects)"
        echo -e "     Install: ${CYAN}pip3 install terminaltexteffects${NC}"
    fi

    if [[ ${#tools_to_install[@]} -gt 0 ]] && command -v brew &> /dev/null; then
        echo
        echo "Missing brew tools: ${tools_to_install[*]}"
        read -p "Install via Homebrew? (y/n) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            for tool in "${tools_to_install[@]}"; do
                echo -e "${BLUE}Installing $tool...${NC}"
                brew install "$tool" 2>/dev/null || true
            done
        fi
    fi

    # Offer to install jscpd via npm
    if [[ "$need_jscpd" == "true" ]]; then
        if command -v npm &> /dev/null; then
            echo
            read -p "Install jscpd via npm? (y/n) " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                echo -e "${BLUE}Installing jscpd...${NC}"
                npm install -g jscpd 2>/dev/null || true
            fi
        elif command -v brew &> /dev/null; then
            echo
            read -p "Install Node.js (required for jscpd)? (y/n) " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                echo -e "${BLUE}Installing node...${NC}"
                brew install node 2>/dev/null || true
                if command -v npm &> /dev/null; then
                    echo -e "${BLUE}Installing jscpd...${NC}"
                    npm install -g jscpd 2>/dev/null || true
                fi
            fi
        else
            echo -e "\n  ${DIM}To install jscpd: Install Node.js first, then run: npm i -g jscpd${NC}"
        fi
    fi
}

setup_configure_env() {
    echo -e "\n${BOLD}Checking API configuration...${NC}"

    if [[ -f "$REPO_ROOT/.env" ]]; then
        echo -e "  ${GREEN}✓${NC} .env file exists"
    else
        echo -e "  ${YELLOW}○${NC} .env file missing"
        if [[ -f "$REPO_ROOT/.env.example" ]]; then
            read -p "  Create from .env.example? (y/n) " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                cp "$REPO_ROOT/.env.example" "$REPO_ROOT/.env"
                echo -e "  ${GREEN}✓${NC} Created .env"
            fi
        fi
    fi

    # Check API keys
    if [[ -n "${ANTHROPIC_API_KEY:-}" ]]; then
        echo -e "  ${GREEN}✓${NC} ANTHROPIC_API_KEY configured"
    else
        echo -e "  ${YELLOW}○${NC} ANTHROPIC_API_KEY not set"
        echo -e "     Get one from: ${CYAN}https://console.anthropic.com/${NC}"
    fi

    if [[ -n "${GITHUB_TOKEN:-}" ]]; then
        echo -e "  ${GREEN}✓${NC} GITHUB_TOKEN configured"
    else
        echo -e "  ${YELLOW}○${NC} GITHUB_TOKEN not set (needed for private repos)"
    fi
}

run_setup() {
    print_phantom_banner
    echo -e "${BOLD}Setup${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo

    setup_make_scripts_executable
    echo
    setup_check_homebrew
    setup_install_tools
    setup_configure_env

    echo
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${GREEN}Setup complete!${NC}"
    echo
    echo "Next: Run ${CYAN}./phantom.sh check${NC} to verify everything is ready"
}

#############################################################################
# Check Functions (from preflight.sh)
#############################################################################

run_check() {
    print_phantom_banner
    echo -e "${BOLD}Preflight Check${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo

    local errors=0
    local warnings=0

    # Required Tools
    echo -e "${BOLD}Required Tools${NC}"
    for tool in git jq curl; do
        printf "  %-16s " "$tool"
        if command -v "$tool" &> /dev/null; then
            local version=$("$tool" --version 2>/dev/null | head -1 | grep -oE '[0-9]+\.[0-9]+(\.[0-9]+)?' | head -1)
            echo -e "${GREEN}✓${NC} ${version:-installed}"
        else
            echo -e "${RED}✗${NC} missing (required)"
            ((errors++))
        fi
    done
    echo

    # Recommended Tools
    echo -e "${BOLD}Recommended Tools${NC}"
    for tool in osv-scanner syft gh cloc bc; do
        printf "  %-16s " "$tool"
        if command -v "$tool" &> /dev/null; then
            local version=$("$tool" --version 2>/dev/null | head -1 | grep -oE '[0-9]+\.[0-9]+(\.[0-9]+)?' | head -1)
            echo -e "${GREEN}✓${NC} ${version:-installed}"
        else
            echo -e "${YELLOW}○${NC} missing (recommended)"
            ((warnings++))
        fi
    done

    # jscpd (npm package, check differently)
    printf "  %-16s " "jscpd"
    if command -v jscpd &> /dev/null; then
        local version=$(jscpd --version 2>/dev/null | head -1 | grep -oE '[0-9]+\.[0-9]+(\.[0-9]+)?' | head -1)
        echo -e "${GREEN}✓${NC} ${version:-installed}"
    elif command -v npx &> /dev/null && npx jscpd --version &> /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} via npx"
    else
        echo -e "${YELLOW}○${NC} missing ${DIM}(npm i -g jscpd)${NC}"
        ((warnings++))
    fi

    # Checkov (check multiple locations)
    printf "  %-16s " "checkov"
    local checkov_bin=""
    if command -v checkov &> /dev/null; then
        checkov_bin="checkov"
    elif [[ -x "$HOME/Library/Python/3.9/bin/checkov" ]]; then
        checkov_bin="$HOME/Library/Python/3.9/bin/checkov"
    elif [[ -x "$HOME/Library/Python/3.10/bin/checkov" ]]; then
        checkov_bin="$HOME/Library/Python/3.10/bin/checkov"
    elif [[ -x "$HOME/Library/Python/3.11/bin/checkov" ]]; then
        checkov_bin="$HOME/Library/Python/3.11/bin/checkov"
    elif [[ -x "$HOME/.local/bin/checkov" ]]; then
        checkov_bin="$HOME/.local/bin/checkov"
    fi
    if [[ -n "$checkov_bin" ]]; then
        local version=$("$checkov_bin" --version 2>/dev/null | head -1 | grep -oE '[0-9]+\.[0-9]+(\.[0-9]+)?' | head -1)
        echo -e "${GREEN}✓${NC} ${version:-installed}"
    else
        echo -e "${YELLOW}○${NC} missing (for IaC security)"
        ((warnings++))
    fi
    echo

    # Optional Tools
    echo -e "${BOLD}Optional Tools${NC}"
    printf "  %-16s " "python3"
    if command -v python3 &> /dev/null; then
        local py_ver=$(python3 --version 2>&1 | grep -oE '[0-9]+\.[0-9]+')
        echo -e "${GREEN}✓${NC} $py_ver"
    else
        echo -e "${DIM}○${NC} not installed"
    fi

    printf "  %-16s " "tte (effects)"
    if python3 -c "import terminaltexteffects" 2>/dev/null; then
        echo -e "${GREEN}✓${NC} installed"
    else
        echo -e "${DIM}○${NC} not installed ${DIM}(pip3 install terminaltexteffects)${NC}"
    fi
    echo

    # API Keys
    echo -e "${BOLD}API Keys${NC}"
    printf "  %-16s " "GITHUB_TOKEN"
    if [[ -n "${GITHUB_TOKEN:-}" ]]; then
        local masked="${GITHUB_TOKEN:0:8}...${GITHUB_TOKEN: -4}"
        echo -e "${GREEN}✓${NC} configured ($masked)"
    else
        echo -e "${YELLOW}○${NC} not set (recommended)"
        ((warnings++))
    fi

    printf "  %-16s " "ANTHROPIC_API_KEY"
    if [[ -n "${ANTHROPIC_API_KEY:-}" ]]; then
        local masked="${ANTHROPIC_API_KEY:0:8}...${ANTHROPIC_API_KEY: -4}"
        echo -e "${GREEN}✓${NC} configured ($masked)"
    else
        echo -e "${YELLOW}○${NC} not set (recommended)"
        ((warnings++))
    fi
    echo

    # GitHub Authentication
    echo -e "${BOLD}GitHub Authentication${NC}"
    printf "  %-16s " "gh auth"
    if command -v gh &> /dev/null; then
        if gh auth status &> /dev/null; then
            local user=$(gh api user -q '.login' 2>/dev/null || echo "authenticated")
            echo -e "${GREEN}✓${NC} logged in as $user"
        else
            echo -e "${YELLOW}○${NC} not authenticated"
            ((warnings++))
        fi
    else
        echo -e "${YELLOW}○${NC} gh not installed"
    fi
    echo

    # Phantom Storage
    echo -e "${BOLD}Phantom Storage${NC}"
    printf "  %-16s " "~/.phantom"
    if [[ -d "$HOME/.phantom" ]]; then
        local count=$(find "$HOME/.phantom/projects" -mindepth 2 -maxdepth 2 -type d 2>/dev/null | wc -l | tr -d ' ')
        echo -e "${GREEN}✓${NC} exists ($count projects)"
    else
        echo -e "${YELLOW}○${NC} not created yet (auto-created on first hydrate)"
    fi
    echo

    # Summary
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    if [[ $errors -eq 0 ]] && [[ $warnings -eq 0 ]]; then
        echo -e "${GREEN}${BOLD}✓ All checks passed!${NC}"
    elif [[ $errors -eq 0 ]]; then
        echo -e "${YELLOW}${BOLD}⚠ $warnings warnings${NC} (you can proceed)"
    else
        echo -e "${RED}${BOLD}✗ $errors errors, $warnings warnings${NC}"
        echo "Run ${CYAN}./phantom.sh setup${NC} to fix issues"
        return 1
    fi
    echo
    echo "Ready to analyze: ${CYAN}./phantom.sh hydrate owner/repo${NC}"
    return 0
}

#############################################################################
# Status Functions
#############################################################################

run_status() {
    print_phantom_banner
    echo -e "${BOLD}Hydrated Projects${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo

    if [[ ! -d "$GIBSON_PROJECTS_DIR" ]]; then
        echo -e "${YELLOW}No projects hydrated yet.${NC}"
        echo
        echo "Hydrate a repository:"
        echo -e "  ${CYAN}./phantom.sh hydrate owner/repo${NC}"
        return 0
    fi

    local count=0
    for org_dir in "$GIBSON_PROJECTS_DIR"/*/; do
        [[ ! -d "$org_dir" ]] && continue
        local org=$(basename "$org_dir")

        for repo_dir in "$org_dir"*/; do
            [[ ! -d "$repo_dir" ]] && continue
            local repo=$(basename "$repo_dir")
            local project_id="${org}/${repo}"
            ((count++))

            # Get project info
            local size=$(du -sh "$repo_dir" 2>/dev/null | cut -f1)
            local analysis_path="$repo_dir/analysis"
            local manifest="$analysis_path/manifest.json"

            # Get mode from manifest
            local mode="unknown"
            if [[ -f "$manifest" ]]; then
                mode=$(jq -r '.mode // "standard"' "$manifest" 2>/dev/null)
            fi

            # Mode display with color
            local mode_display=""
            case "$mode" in
                quick)    mode_display="${DIM}quick${NC}" ;;
                standard) mode_display="${CYAN}standard${NC}" ;;
                advanced) mode_display="${BLUE}advanced${NC}" ;;
                deep)     mode_display="${MAGENTA}deep${NC}" ;;
                security) mode_display="${YELLOW}security${NC}" ;;
                *)        mode_display="${DIM}$mode${NC}" ;;
            esac

            echo -e "${BOLD}$project_id${NC} ${DIM}($size)${NC} [${mode_display}]"

            # Show key metrics if available
            if [[ -f "$analysis_path/vulnerabilities.json" ]]; then
                local critical=$(jq -r '.summary.critical // 0' "$analysis_path/vulnerabilities.json" 2>/dev/null)
                local high=$(jq -r '.summary.high // 0' "$analysis_path/vulnerabilities.json" 2>/dev/null)
                if [[ "$critical" != "0" ]] || [[ "$high" != "0" ]]; then
                    echo -e "  Vulnerabilities: ${RED}$critical critical${NC}, ${YELLOW}$high high${NC}"
                else
                    echo -e "  Vulnerabilities: ${GREEN}clean${NC}"
                fi
            fi

            if [[ -f "$analysis_path/dependencies.json" ]]; then
                local deps=$(jq -r '.total_dependencies // 0' "$analysis_path/dependencies.json" 2>/dev/null)
                echo -e "  Dependencies: $deps"
            fi
            echo
        done
    done

    if [[ $count -eq 0 ]]; then
        echo -e "${YELLOW}No projects hydrated yet.${NC}"
        echo
        echo "Hydrate a repository:"
        echo -e "  ${CYAN}./phantom.sh hydrate owner/repo${NC}"
    else
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo -e "Total: ${BOLD}$count${NC} projects"
        echo -e "Storage: ${CYAN}~/.phantom/projects/${NC}"
    fi
}

#############################################################################
# Report Functions
#############################################################################

run_report() {
    exec "$SCRIPT_DIR/report.sh" "$@"
}

#############################################################################
# History Functions
#############################################################################

run_history() {
    local target="$1"
    local limit="${2:-10}"

    if [[ -z "$target" ]]; then
        echo -e "${RED}Error: No target specified${NC}" >&2
        echo "Usage: $(basename "$0") history <org/repo>"
        exit 1
    fi

    local project_id=$(gibson_project_id "$target")
    local history=$(gibson_get_scan_history "$project_id" "$limit")

    if [[ -z "$history" ]] || [[ "$history" == "null" ]]; then
        echo -e "${RED}Error: No scan history found for '$project_id'${NC}" >&2
        exit 1
    fi

    print_phantom_banner
    echo -e "${BOLD}Scan History: $project_id${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo

    local total_scans=$(echo "$history" | jq -r '.total_scans // 0')
    local first_scan=$(echo "$history" | jq -r '.first_scan_at // "unknown"')
    local last_scan=$(echo "$history" | jq -r '.last_scan_at // "unknown"')

    printf "  %-14s %s\n" "Total Scans:" "$total_scans"
    printf "  %-14s %s\n" "First Scan:" "$(echo "$first_scan" | cut -d'T' -f1)"
    printf "  %-14s %s\n" "Last Scan:" "$(echo "$last_scan" | cut -d'T' -f1)"
    echo

    echo -e "${BOLD}Recent Scans${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    echo "$history" | jq -r '.scans // [] | .[] | "\(.scan_id)\t\(.started_at | split("T")[0])\t\(.profile)\t\(.status)\t\(.summary.vulnerability_count // 0) vulns"' 2>/dev/null | \
    while IFS=$'\t' read -r scan_id date profile status vulns; do
        local status_color="$GREEN"
        [[ "$status" == "failed" ]] && status_color="$RED"
        [[ "$status" == "partial" ]] && status_color="$YELLOW"

        printf "  %-24s %-12s %-10s ${status_color}%-10s${NC} %s\n" "$scan_id" "$date" "$profile" "$status" "$vulns"
    done

    echo
}

#############################################################################
# Clean Functions
#############################################################################

run_clean() {
    local target=""
    local org=""
    local dry_run=false
    local skip_confirm=false

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --org)
                org="$2"
                shift 2
                ;;
            --dry-run)
                dry_run=true
                shift
                ;;
            --yes|-y)
                skip_confirm=true
                shift
                ;;
            -*)
                echo -e "${RED}Error: Unknown option $1${NC}" >&2
                exit 1
                ;;
            *)
                target="$1"
                shift
                ;;
        esac
    done

    print_phantom_banner
    echo -e "${BOLD}Clean${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo

    if [[ ! -d "$GIBSON_PROJECTS_DIR" ]]; then
        echo "No projects to clean."
        return 0
    fi

    # Determine what to clean
    if [[ -n "$target" ]]; then
        # Clean single project
        local project_id=$(gibson_project_id "$target")
        local project_path=$(gibson_project_path "$project_id")

        if [[ ! -d "$project_path" ]]; then
            echo -e "${RED}Error: Project '$project_id' not found${NC}"
            exit 1
        fi

        local size=$(du -sh "$project_path" 2>/dev/null | cut -f1)
        echo "  Project: $project_id"
        echo "  Size: $size"
        echo

        if [[ "$dry_run" == "true" ]]; then
            echo -e "${CYAN}[DRY RUN]${NC} Would remove: $project_path"
            return 0
        fi

        if [[ "$skip_confirm" != "true" ]]; then
            read -p "Remove this project? (y/n) " -n 1 -r
            echo
            [[ ! $REPLY =~ ^[Yy]$ ]] && { echo "Cancelled."; return 0; }
        fi

        gibson_clean_project "$project_id"
        echo -e "${GREEN}✓${NC} Cleaned project: $project_id"

    elif [[ -n "$org" ]]; then
        # Clean entire org
        local projects=$(gibson_list_org_projects "$org")
        if [[ -z "$projects" ]]; then
            echo -e "${RED}Error: No projects found for org '$org'${NC}"
            exit 1
        fi

        local count=$(echo "$projects" | wc -w | tr -d ' ')
        local size=$(du -sh "$GIBSON_PROJECTS_DIR/$org" 2>/dev/null | cut -f1)

        echo "  Organization: $org"
        echo "  Projects: $count"
        echo "  Size: $size"
        echo
        echo "  Projects to remove:"
        for repo in $projects; do
            echo "    - $org/$repo"
        done
        echo

        if [[ "$dry_run" == "true" ]]; then
            echo -e "${CYAN}[DRY RUN]${NC} Would remove: $GIBSON_PROJECTS_DIR/$org/"
            return 0
        fi

        if [[ "$skip_confirm" != "true" ]]; then
            read -p "Remove all projects in '$org'? (y/n) " -n 1 -r
            echo
            [[ ! $REPLY =~ ^[Yy]$ ]] && { echo "Cancelled."; return 0; }
        fi

        gibson_clean_org "$org"
        echo -e "${GREEN}✓${NC} Cleaned org: $org ($count projects)"

    else
        # Clean everything
        local count=$(find "$GIBSON_PROJECTS_DIR" -mindepth 2 -maxdepth 2 -type d 2>/dev/null | wc -l | tr -d ' ')
        local size=$(du -sh "$GIBSON_DIR" 2>/dev/null | cut -f1)

        echo -e "${YELLOW}Warning:${NC} This will remove ALL analysis data!"
        echo
        echo "  Projects: $count"
        echo "  Size: $size"
        echo "  Location: ~/.phantom/"
        echo

        if [[ "$dry_run" == "true" ]]; then
            echo -e "${CYAN}[DRY RUN]${NC} Would remove: $GIBSON_DIR"
            return 0
        fi

        if [[ "$skip_confirm" != "true" ]]; then
            read -p "Are you sure? (y/n) " -n 1 -r
            echo
            [[ ! $REPLY =~ ^[Yy]$ ]] && { echo "Cancelled."; return 0; }
        fi

        rm -rf "$GIBSON_DIR"
        echo -e "${GREEN}✓${NC} Cleaned all data"
    fi
}

#############################################################################
# Interactive Menu
#############################################################################

#############################################################################
# Helper: Get hydration status for a target
#############################################################################

get_hydration_status() {
    local target="$1"
    local project_id=""

    # Determine project_id from target
    if [[ "$target" == --org* ]]; then
        # Org mode - can't check individual status
        echo ""
        return
    fi

    # Convert target to project_id format
    if [[ "$target" =~ ^https://github\.com/(.+)$ ]]; then
        project_id="${BASH_REMATCH[1]%.git}"
    elif [[ "$target" =~ ^([^/]+)/([^/]+)$ ]]; then
        project_id="$target"
    else
        echo ""
        return
    fi

    # Check if project exists
    local project_path="$GIBSON_PROJECTS_DIR/${project_id//\//_}"
    project_path="$GIBSON_PROJECTS_DIR/$(echo "$project_id" | tr '/' '/')"

    # Parse as org/repo
    local org=$(echo "$project_id" | cut -d'/' -f1)
    local repo=$(echo "$project_id" | cut -d'/' -f2)
    project_path="$GIBSON_PROJECTS_DIR/$org/$repo"

    if [[ -d "$project_path/analysis" ]]; then
        local manifest="$project_path/analysis/manifest.json"
        if [[ -f "$manifest" ]]; then
            local mode=$(jq -r '.mode // "standard"' "$manifest" 2>/dev/null)
            local completed=$(jq -r '.completed_at // ""' "$manifest" 2>/dev/null)
            if [[ -n "$completed" ]] && [[ "$completed" != "null" ]]; then
                echo "$mode"
                return
            fi
        fi
    fi
    echo ""
}

# Get mode display with status
get_mode_display() {
    local mode="$1"
    local current_mode="$2"
    local mode_name="$3"
    local time_est="$4"
    local description="$5"

    if [[ "$current_mode" == "$mode" ]]; then
        echo -e "  ${CYAN}$1${NC}  ${mode_name}   ${time_est}  ${description} ${GREEN}[hydrated]${NC}"
    else
        echo -e "  ${CYAN}$1${NC}  ${mode_name}   ${time_est}  ${description}"
    fi
}

# Configuration file path (unified config)
CONFIG_FILE="$SCRIPT_DIR/config/phantom.config.json"

# Get profile info from phantom.config.json
get_profile_info() {
    local profile="$1"
    local field="$2"
    if [[ -f "$CONFIG_FILE" ]]; then
        jq -r --arg p "$profile" --arg f "$field" '.profiles[$p][$f] // empty' "$CONFIG_FILE" 2>/dev/null
    fi
}

# Get scanners for a profile
get_profile_scanners() {
    local profile="$1"
    if [[ -f "$CONFIG_FILE" ]]; then
        jq -r --arg p "$profile" '.profiles[$p].scanners // [] | join(" ")' "$CONFIG_FILE" 2>/dev/null
    fi
}

# Check if profile requires Claude API (claude_mode is "enabled" or "required")
profile_requires_claude() {
    local profile="$1"
    if [[ -f "$CONFIG_FILE" ]]; then
        local mode=$(jq -r --arg p "$profile" '.profiles[$p].claude_mode // "none"' "$CONFIG_FILE" 2>/dev/null)
        [[ "$mode" == "enabled" || "$mode" == "required" ]]
    else
        [[ "$profile" == "deep" ]]
    fi
}

# Get scanners that should use Claude for a profile
get_claude_scanners() {
    local profile="$1"
    if [[ -f "$CONFIG_FILE" ]]; then
        jq -r --arg p "$profile" '.profiles[$p].claude_scanners // [] | join(" ")' "$CONFIG_FILE" 2>/dev/null
    fi
}

# Check if a scanner supports Claude (has claude_mode "optional" or "required")
scanner_supports_claude() {
    local scanner="$1"
    if [[ -f "$CONFIG_FILE" ]]; then
        local mode=$(jq -r --arg s "$scanner" '.scanners[$s].claude_mode // "none"' "$CONFIG_FILE" 2>/dev/null)
        [[ "$mode" == "optional" || "$mode" == "required" ]]
    else
        return 1
    fi
}

# List all scanners that support Claude
list_claude_capable_scanners() {
    if [[ -f "$CONFIG_FILE" ]]; then
        jq -r '.scanners | to_entries[] | select(.value.claude_mode == "optional" or .value.claude_mode == "required") | .key' "$CONFIG_FILE" 2>/dev/null
    fi
}

# Get config setting
get_config_setting() {
    local key="$1"
    local default="$2"
    if [[ -f "$CONFIG_FILE" ]]; then
        local value=$(jq -r ".settings.$key // empty" "$CONFIG_FILE" 2>/dev/null)
        if [[ -n "$value" ]] && [[ "$value" != "null" ]]; then
            echo "$value"
            return 0
        fi
    fi
    echo "$default"
}

show_menu() {
    local first_run=true

    while true; do
        print_phantom_banner

        # Get hydrated project count
        local hydrated_count=0
        if [[ -d "$GIBSON_PROJECTS_DIR" ]]; then
            hydrated_count=$(find "$GIBSON_PROJECTS_DIR" -mindepth 2 -maxdepth 2 -type d 2>/dev/null | wc -l | tr -d ' ')
        fi

        echo -e "${BOLD}What would you like to do?${NC}"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo
        echo -e "  ${CYAN}1${NC}  Setup       Install tools and configure API keys"
        echo -e "  ${CYAN}2${NC}  Check       Verify everything is ready"
        echo
        echo -e "  ${BOLD}Hydrate a repository:${NC}"

        # Display profiles dynamically from phantom.config.json
        local profile_keys=()
        local menu_num=3

        if [[ -f "$CONFIG_FILE" ]]; then
            # Standard profile order
            local ordered_profiles=("quick" "standard" "advanced" "deep" "ai-only" "security" "security-deep" "compliance" "devops")

            for profile in "${ordered_profiles[@]}"; do
                if jq -e --arg p "$profile" '.profiles[$p]' "$CONFIG_FILE" &>/dev/null; then
                    profile_keys+=("$profile")

                    local name=$(get_profile_info "$profile" "name")
                    local time=$(get_profile_info "$profile" "estimated_time")
                    local desc=$(get_profile_info "$profile" "description")

                    local markers=""
                    [[ "$profile" == "standard" ]] && markers=" ${DIM}(recommended)${NC}"
                    if profile_requires_claude "$profile"; then
                        markers="${markers} ${DIM}(requires API key)${NC}"
                    fi

                    printf "  ${CYAN}%s${NC}  %-10s %-7s %s%s\n" "$menu_num" "$name" "$time" "$desc" "$markers"
                    ((menu_num++))
                fi
            done
        else
            # Fallback if no phantom.config.json
            echo -e "  ${CYAN}3${NC}  Quick       ~30s   Fast scan (deps, tech, vulns, licenses)"
            echo -e "  ${CYAN}4${NC}  Standard    ~2min  Most scanners ${DIM}(recommended)${NC}"
            echo -e "  ${CYAN}5${NC}  Advanced    ~5min  All static scanners + health/provenance"
            echo -e "  ${CYAN}6${NC}  Deep        ~10min Claude-assisted analysis ${DIM}(requires API key)${NC}"
            echo -e "  ${CYAN}7${NC}  Security    ~3min  Security-focused (vulns, code security)"
            profile_keys=("quick" "standard" "advanced" "deep" "security")
            menu_num=8
        fi

        echo -e "  ${CYAN}c${NC}  Choose      Custom Select specific collectors (checkboxes)"
        echo
        if [[ $hydrated_count -gt 0 ]]; then
            echo -e "  ${CYAN}r${NC}  Report      Generate analysis reports ${DIM}($hydrated_count projects)${NC}"
            echo -e "  ${CYAN}s${NC}  Status      Show hydrated projects ${DIM}($hydrated_count projects)${NC}"
        else
            echo -e "  ${CYAN}r${NC}  Report      Generate analysis reports"
            echo -e "  ${CYAN}s${NC}  Status      Show hydrated projects"
        fi
        echo -e "  ${CYAN}x${NC}  Clean       Remove all analysis data"
        echo
        echo -e "  ${CYAN}q${NC}  Quit"
        echo
        read -p "Choose an option: " -r
        echo
        echo

        case $REPLY in
            1)
                run_setup
                echo
                read -p "Press Enter to continue..."
                ;;
            2)
                run_check || true
                echo
                read -p "Press Enter to continue..."
                ;;
            [3-9])
                # Handle dynamic profile selection (3-9 = profiles)
                local profile_idx=$((REPLY - 3))
                if [[ $profile_idx -lt ${#profile_keys[@]} ]]; then
                    local selected_profile="${profile_keys[$profile_idx]}"
                    local mode_flag="--$selected_profile"
                    local mode_name=$(get_profile_info "$selected_profile" "name")
                    [[ -z "$mode_name" ]] && mode_name="$selected_profile"

                    # Show scanners for this profile
                    local scanners=$(get_profile_scanners "$selected_profile")
                    echo -e "${BOLD}$mode_name Hydration${NC}"
                    echo -e "${DIM}Scanners: $scanners${NC}"
                    echo
                    echo -e "Enter repository (e.g., ${CYAN}expressjs/express${NC})"
                    echo -e "Or organization with ${CYAN}--org orgname${NC}"
                    echo
                    read -p "Target: " target

                    if [[ -n "$target" ]]; then
                        local should_run=true
                        local force_flag=""

                        # Check if already hydrated (for single repo, not org)
                        if [[ "$target" != --org* ]]; then
                            local current_status=$(get_hydration_status "$target")
                            if [[ -n "$current_status" ]]; then
                                echo
                                echo -e "${YELLOW}This repository is already hydrated${NC} with mode: ${CYAN}$current_status${NC}"
                                echo
                                echo -e "  ${CYAN}1${NC}  Skip (use existing analysis)"
                                echo -e "  ${CYAN}2${NC}  Re-hydrate with $mode_name mode"
                                echo
                                read -p "Choose [1]: " -n 1 -r override_choice
                                echo

                                case "${override_choice:-1}" in
                                    2)
                                        force_flag="--force"
                                        echo -e "Re-hydrating with ${CYAN}$mode_name${NC} mode..."
                                        ;;
                                    *)
                                        should_run=false
                                        echo "Skipped."
                                        ;;
                                esac
                            fi
                        fi

                        if [[ "$should_run" == "true" ]]; then
                            # Check if org mode
                            if [[ "$target" == --org* ]]; then
                                "$SCRIPT_DIR/hydrate.sh" $target $mode_flag $force_flag || true
                            else
                                "$SCRIPT_DIR/bootstrap.sh" $target $mode_flag $force_flag || true
                            fi
                        fi
                        echo
                        read -p "Press Enter to continue..."
                    fi
                else
                    echo "Invalid option"
                    sleep 1
                fi
                ;;
            c|C)
                # Custom collector selection mode
                echo -e "${BOLD}Custom Hydration${NC}"
                echo
                echo -e "Enter repository (e.g., ${CYAN}expressjs/express${NC})"
                echo -e "Or organization with ${CYAN}--org orgname${NC}"
                echo
                read -p "Target: " target

                if [[ -n "$target" ]]; then
                    local force_flag=""

                    # Check if already hydrated (for single repo, not org)
                    if [[ "$target" != --org* ]]; then
                        local current_status=$(get_hydration_status "$target")
                        if [[ -n "$current_status" ]]; then
                            echo
                            echo -e "${YELLOW}This repository is already hydrated${NC} with mode: ${CYAN}$current_status${NC}"
                            echo
                            echo -e "  ${CYAN}1${NC}  Skip (use existing analysis)"
                            echo -e "  ${CYAN}2${NC}  Re-hydrate with custom collectors"
                            echo
                            read -p "Choose [1]: " -n 1 -r override_choice
                            echo

                            case "${override_choice:-1}" in
                                2)
                                    force_flag="--force"
                                    ;;
                                *)
                                    echo "Skipped."
                                    read -p "Press Enter to continue..."
                                    continue
                                    ;;
                            esac
                        fi
                    fi

                    # Run with --choose flag for interactive collector selection
                    if [[ "$target" == --org* ]]; then
                        "$SCRIPT_DIR/hydrate.sh" $target --choose $force_flag || true
                    else
                        "$SCRIPT_DIR/hydrate.sh" $target --choose $force_flag || true
                    fi
                    echo
                    read -p "Press Enter to continue..."
                fi
                ;;
            r|R)
                # Run report generator in interactive mode
                "$SCRIPT_DIR/report.sh" --interactive || true
                echo
                read -p "Press Enter to continue..."
                ;;
            s|S)
                run_status
                echo
                read -p "Press Enter to continue..."
                ;;
            x|X)
                run_clean
                echo
                read -p "Press Enter to continue..."
                ;;
            q|Q)
                echo "Goodbye!"
                exit 0
                ;;
            *)
                echo "Invalid option"
                sleep 1
                ;;
        esac
    done
}

#############################################################################
# Usage
#############################################################################

usage() {
    cat << EOF
Phantom - Repository Analysis CLI

Usage: $(basename "$0") [command] [options]

COMMANDS:
    (none)              Interactive menu
    setup               Install tools and configure API keys
    check               Verify tools and configuration
    hydrate <repo>      Analyze a repository (e.g., expressjs/express)
    hydrate --org <n>   Analyze all repos in an organization
    status              Show hydrated projects
    report <repo>       Generate summary report for a project
    history <repo>      Show scan history for a project
    clean               Remove analysis data (all, org, or project)
    help                Show this help

OPTIONS FOR HYDRATE:
    --org <name>        Process all repos in organization
    --limit <n>         Max repos to process (org mode)
    --quick             Fast static analysis (~30s)
    --standard          Most scanners (~2min) [default]
    --advanced          All static scanners + health/provenance (~5min)
    --deep              Claude-assisted analysis (~10min)
    --security          Security-focused analysis (~3min)
    --compliance        License and policy compliance (~2min)
    --devops            CI/CD and operational metrics (~3min)
    --force             Re-analyze even if exists

OPTIONS FOR REPORT:
    <org/repo>          Report for a specific project
    --org <name>        Aggregate report for an organization
    --json              Output in JSON format

OPTIONS FOR CLEAN:
    (no args)           Clean all data (with confirmation)
    <org/repo>          Clean a specific project
    --org <name>        Clean all projects in an organization
    --dry-run           Preview what would be deleted
    --yes               Skip confirmation prompt

CONFIGURATION:
    All settings are in utils/phantom/config/phantom.config.json
    See phantom.config.example.json for full documentation
    Create custom profiles by adding entries to the profiles section

EXAMPLES:
    $(basename "$0")                              # Interactive mode
    $(basename "$0") setup                        # First-time setup
    $(basename "$0") hydrate lodash/lodash        # Single repo
    $(basename "$0") hydrate --org expressjs      # All org repos
    $(basename "$0") status                       # List projects
    $(basename "$0") report expressjs/express     # Project report
    $(basename "$0") report --org expressjs       # Org report
    $(basename "$0") history expressjs/express    # Scan history
    $(basename "$0") clean expressjs/express      # Clean one project
    $(basename "$0") clean --org expressjs        # Clean org

STORAGE:
    Analysis data is stored in ~/.phantom/projects/

EOF
    exit 0
}

#############################################################################
# Main
#############################################################################

main() {
    case "${1:-}" in
        "")
            show_menu
            ;;
        setup)
            run_setup
            ;;
        check|preflight)
            run_check
            ;;
        hydrate|bootstrap)
            shift
            if [[ $# -eq 0 ]]; then
                echo -e "${RED}Error: No target specified${NC}"
                echo "Usage: $(basename "$0") hydrate owner/repo"
                echo "       $(basename "$0") hydrate --org orgname"
                exit 1
            fi
            # Check preflight first
            if ! run_check > /dev/null 2>&1; then
                echo -e "${YELLOW}Warning: Preflight check has issues. Run './phantom.sh check' to see details.${NC}"
                echo
            fi
            # Delegate to hydrate.sh for single repos or orgs
            if [[ "$1" == "--org" ]]; then
                exec "$SCRIPT_DIR/hydrate.sh" "$@"
            else
                exec "$SCRIPT_DIR/bootstrap.sh" "$@"
            fi
            ;;
        status|list)
            run_status
            ;;
        report)
            shift
            run_report "$@"
            ;;
        history)
            shift
            run_history "$@"
            ;;
        clean)
            shift
            run_clean "$@"
            ;;
        -h|--help|help)
            usage
            ;;
        *)
            echo -e "${RED}Unknown command: $1${NC}"
            echo "Run '$(basename "$0") help' for usage"
            exit 1
            ;;
    esac
}

main "$@"
