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

    # Recommended tools
    for tool in osv-scanner syft gh; do
        if command -v "$tool" &> /dev/null; then
            echo -e "  ${GREEN}✓${NC} $tool"
        else
            echo -e "  ${YELLOW}○${NC} $tool (recommended)"
            tools_to_install+=("$tool")
        fi
    done

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
        echo "Missing tools: ${tools_to_install[*]}"
        read -p "Install via Homebrew? (y/n) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            for tool in "${tools_to_install[@]}"; do
                echo -e "${BLUE}Installing $tool...${NC}"
                brew install "$tool" 2>/dev/null || true
            done
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
    for tool in osv-scanner syft gh; do
        printf "  %-16s " "$tool"
        if command -v "$tool" &> /dev/null; then
            local version=$("$tool" --version 2>/dev/null | head -1 | grep -oE '[0-9]+\.[0-9]+(\.[0-9]+)?' | head -1)
            echo -e "${GREEN}✓${NC} ${version:-installed}"
        else
            echo -e "${YELLOW}○${NC} missing (recommended)"
            ((warnings++))
        fi
    done
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

            echo -e "${BOLD}$project_id${NC} ${DIM}($size)${NC}"

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
# Clean Functions
#############################################################################

run_clean() {
    print_phantom_banner
    echo -e "${BOLD}Clean${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo

    if [[ ! -d "$GIBSON_PROJECTS_DIR" ]]; then
        echo "No projects to clean."
        return 0
    fi

    local count=$(find "$GIBSON_PROJECTS_DIR" -mindepth 2 -maxdepth 2 -type d 2>/dev/null | wc -l | tr -d ' ')
    local size=$(du -sh "$GIBSON_DIR" 2>/dev/null | cut -f1)

    echo -e "${YELLOW}Warning:${NC} This will remove all analysis data!"
    echo
    echo "  Projects: $count"
    echo "  Size: $size"
    echo "  Location: ~/.phantom/"
    echo
    read -p "Are you sure? (y/n) " -n 1 -r
    echo

    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf "$GIBSON_DIR"
        echo -e "${GREEN}✓${NC} Cleaned all data"
    else
        echo "Cancelled."
    fi
}

#############################################################################
# Interactive Menu
#############################################################################

show_menu() {
    local first_run=true

    while true; do
        # Use animated banner on first display (random effect), static after
        if [[ "$first_run" == "true" ]]; then
            print_phantom_banner_animated  # Random effect each time
            first_run=false
        else
            print_phantom_banner
        fi
        echo -e "${BOLD}What would you like to do?${NC}"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo
        echo -e "  ${CYAN}1${NC}  Setup       Install tools and configure API keys"
        echo -e "  ${CYAN}2${NC}  Check       Verify everything is ready"
        echo -e "  ${CYAN}3${NC}  Hydrate     Analyze a repository"
        echo -e "  ${CYAN}4${NC}  Status      Show hydrated projects"
        echo -e "  ${CYAN}5${NC}  Clean       Remove all analysis data"
        echo
        echo -e "  ${CYAN}q${NC}  Quit"
        echo
        read -p "Choose an option: " -n 1 -r
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
            3)
                echo -e "Enter repository (e.g., ${CYAN}expressjs/express${NC})"
                echo -e "Or organization with ${CYAN}--org orgname${NC}"
                echo
                read -p "Target: " target
                if [[ -n "$target" ]]; then
                    # Run hydrate and return to menu when done
                    if [[ "$target" == --org* ]]; then
                        "$SCRIPT_DIR/hydrate.sh" $target || true
                    else
                        "$SCRIPT_DIR/bootstrap.sh" $target || true
                    fi
                    echo
                    read -p "Press Enter to continue..."
                fi
                ;;
            4)
                run_status
                echo
                read -p "Press Enter to continue..."
                ;;
            5)
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
    clean               Remove all analysis data
    help                Show this help

OPTIONS FOR HYDRATE:
    --org <name>        Process all repos in organization
    --limit <n>         Max repos to process (org mode)
    --quick             Fast analyzers only
    --force             Re-analyze even if exists

EXAMPLES:
    $(basename "$0")                              # Interactive mode
    $(basename "$0") setup                        # First-time setup
    $(basename "$0") hydrate lodash/lodash        # Single repo
    $(basename "$0") hydrate --org expressjs      # All org repos
    $(basename "$0") status                       # List projects

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
        clean)
            run_clean
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
