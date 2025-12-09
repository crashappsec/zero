#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Phantom Clone
# Clone repositories for analysis
#
# Usage:
#   ./clone.sh <owner/repo>           # Single repo
#   ./clone.sh --org <org-name>       # All repos in an org
#
# Examples:
#   ./clone.sh expressjs/express
#   ./clone.sh --org expressjs
#   ./clone.sh --org my-company --limit 10
#############################################################################

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ZERO_UTILS_DIR="$(dirname "$SCRIPT_DIR")"

# Load Zero library (sets ZERO_DIR to .zero data directory in project root)
source "$ZERO_UTILS_DIR/lib/zero-lib.sh"

# Load .env if available
UTILS_ROOT="$(dirname "$ZERO_UTILS_DIR")"
REPO_ROOT="$(dirname "$UTILS_ROOT")"
if [[ -f "$REPO_ROOT/.env" ]]; then
    set -a
    source "$REPO_ROOT/.env"
    set +a
fi

#############################################################################
# Configuration
#############################################################################

ORG_MODE=false
ORG_NAME=""
LIMIT=0
TARGET=""
BRANCH=""
DEPTH=""
FORCE=false

#############################################################################
# Usage
#############################################################################

usage() {
    cat << EOF
Phantom Clone - Clone repositories for analysis

Usage: $0 <target> [options]
       $0 --org <org-name> [options]

MODES:
    Single Repo:    $0 owner/repo [options]
    Organization:   $0 --org <org-name> [options]

OPTIONS:
    --org <name>        Clone all repos in a GitHub organization
    --limit <n>         Max repos to clone in org mode (default: all)
    --branch <name>     Clone specific branch (default: default branch)
    --depth <n>         Shallow clone depth (default: full clone)
    --force             Re-clone even if repo exists
    -h, --help          Show this help

EXAMPLES:
    $0 expressjs/express                    # Single repo
    $0 expressjs/express --depth 1          # Shallow clone
    $0 --org expressjs                      # All repos in expressjs org
    $0 --org my-company --limit 10          # First 10 repos in org

EOF
    exit 0
}

#############################################################################
# Argument Parsing
#############################################################################

parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            -h|--help)
                usage
                ;;
            --org)
                ORG_MODE=true
                ORG_NAME="$2"
                shift 2
                ;;
            --limit)
                LIMIT="$2"
                shift 2
                ;;
            --branch)
                BRANCH="$2"
                shift 2
                ;;
            --depth)
                DEPTH="$2"
                shift 2
                ;;
            --force)
                FORCE=true
                shift
                ;;
            -*)
                echo -e "${RED}Error: Unknown option $1${NC}" >&2
                exit 1
                ;;
            *)
                if [[ -z "$TARGET" ]]; then
                    TARGET="$1"
                else
                    echo -e "${RED}Error: Multiple targets specified${NC}" >&2
                    exit 1
                fi
                shift
                ;;
        esac
    done

    # Validate arguments
    if [[ "$ORG_MODE" == "true" ]]; then
        if [[ -z "$ORG_NAME" ]]; then
            echo -e "${RED}Error: --org requires an organization name${NC}" >&2
            exit 1
        fi
    elif [[ -z "$TARGET" ]]; then
        echo -e "${RED}Error: No target specified${NC}" >&2
        echo "Usage: $0 <owner/repo> or $0 --org <org-name>"
        exit 1
    elif [[ ! "$TARGET" =~ ^[^/]+/[^/]+$ ]]; then
        echo -e "${RED}Error: Invalid target format '${TARGET}'${NC}" >&2
        echo "Expected format: owner/repo (e.g., expressjs/express)"
        echo ""
        echo "Did you mean to use --org mode?"
        echo "  ./zero.sh hydrate --org $TARGET"
        exit 1
    fi
}

#############################################################################
# GitHub Functions
#############################################################################

# Check gh CLI is available and authenticated
check_gh_cli() {
    if ! command -v gh &> /dev/null; then
        echo -e "${RED}Error: GitHub CLI (gh) is required for --org mode${NC}" >&2
        echo -e "Install with: ${CYAN}brew install gh${NC}" >&2
        exit 1
    fi

    if gh auth status &> /dev/null; then
        return 0
    fi

    if [[ -n "${GITHUB_TOKEN:-}" ]]; then
        export GH_TOKEN="$GITHUB_TOKEN"
        return 0
    fi

    echo -e "${RED}Error: GitHub CLI not authenticated${NC}" >&2
    echo -e "Run: ${CYAN}gh auth login${NC}" >&2
    exit 1
}

# Fetch org repos with metadata
fetch_org_repos() {
    local org="$1"
    local limit="$2"

    local gh_limit=1000
    [[ $limit -gt 0 ]] && gh_limit=$limit

    gh repo list "$org" \
        --json nameWithOwner,diskUsage,primaryLanguage \
        --limit "$gh_limit" 2>/dev/null
}

# Format bytes to human readable
format_size() {
    local bytes=$1
    if [[ $bytes -ge 1073741824 ]]; then
        echo "$(( bytes / 1073741824 ))GB"
    elif [[ $bytes -ge 1048576 ]]; then
        echo "$(( bytes / 1048576 ))MB"
    elif [[ $bytes -ge 1024 ]]; then
        echo "$(( bytes / 1024 ))KB"
    else
        echo "${bytes}B"
    fi
}

#############################################################################
# Clone Functions
#############################################################################

# Clone a single repository
clone_repo() {
    local repo="$1"
    local project_id=$(zero_project_id "$repo")
    local repo_path="$ZERO_PROJECTS_DIR/$project_id/repo"

    # Check if already cloned
    if [[ -d "$repo_path" ]] && [[ "$FORCE" != "true" ]]; then
        local size=$(du -sh "$repo_path" 2>/dev/null | cut -f1)
        echo -e "  ${GREEN}✓${NC} Already cloned ${DIM}($size)${NC}"
        return 0
    fi

    # Remove existing if force
    if [[ -d "$repo_path" ]]; then
        rm -rf "$repo_path"
    fi

    # Create project directory
    mkdir -p "$ZERO_PROJECTS_DIR/$project_id"

    # Build clone URL
    local clone_url="https://github.com/$repo.git"
    if [[ -n "${GITHUB_TOKEN:-}" ]]; then
        clone_url="https://${GITHUB_TOKEN}@github.com/$repo.git"
    fi

    # Build clone args
    local clone_args=("--progress")
    [[ -n "$BRANCH" ]] && clone_args+=("--branch" "$BRANCH")
    [[ -n "$DEPTH" ]] && clone_args+=("--depth" "$DEPTH")

    # Clone with progress
    local start_time=$(date +%s)

    echo -e "  ${CYAN}Cloning...${NC}"

    if GIT_PROGRESS_DELAY=0 git clone -c checkout.workers=0 "${clone_args[@]}" "$clone_url" "$repo_path" 2>&1 | \
        tr '\r' '\n' | while IFS= read -r line; do
            # Show key progress updates
            if [[ "$line" =~ "Receiving objects:" ]] && [[ "$line" =~ [0-9]+% ]]; then
                local pct=$(echo "$line" | grep -oE '[0-9]+%' | head -1)
                printf "\r  ${CYAN}Receiving: %s${NC}          " "$pct"
            elif [[ "$line" =~ "Resolving deltas:" ]] && [[ "$line" =~ [0-9]+% ]]; then
                local pct=$(echo "$line" | grep -oE '[0-9]+%' | head -1)
                printf "\r  ${CYAN}Resolving: %s${NC}          " "$pct"
            fi
        done; then

        local duration=$(($(date +%s) - start_time))
        local size=$(du -sh "$repo_path" 2>/dev/null | cut -f1)
        local files=$(find "$repo_path" -type f ! -path '*/.git/*' 2>/dev/null | wc -l | tr -d ' ')
        printf "\r  ${GREEN}✓${NC} Cloned ${DIM}($size, $files files, ${duration}s)${NC}          \n"

        # Create analysis directory
        mkdir -p "$ZERO_PROJECTS_DIR/$project_id/analysis"
        return 0
    else
        printf "\r  ${RED}✗${NC} Clone failed                    \n"
        return 1
    fi
}

#############################################################################
# Main Functions
#############################################################################

clone_single() {
    local repo="$1"

    print_zero_banner
    echo -e "${BOLD}Clone Repository${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo
    echo -e "Target: ${CYAN}$repo${NC}"
    echo

    clone_repo "$repo"
    local status=$?

    echo
    if [[ $status -eq 0 ]]; then
        echo -e "${GREEN}✓ Clone complete${NC}"
        echo -e "Run scanners with: ${CYAN}./zero.sh scan $repo${NC}"
    else
        echo -e "${RED}✗ Clone failed${NC}"
    fi

    return $status
}

clone_org() {
    local org="$1"

    check_gh_cli

    print_zero_banner
    echo -e "${BOLD}Clone Organization${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo

    echo -e "${BLUE}Fetching repositories for ${CYAN}$org${BLUE}...${NC}"
    local repos_json=$(fetch_org_repos "$org" "$LIMIT")

    if [[ -z "$repos_json" ]] || [[ "$repos_json" == "[]" ]]; then
        echo -e "${RED}No repos found for organization: $org${NC}" >&2
        exit 1
    fi

    # Parse stats
    local repo_count=$(echo "$repos_json" | jq 'length')
    local total_size_kb=$(echo "$repos_json" | jq '[.[].diskUsage // 0] | add')
    local total_size=$(format_size $((total_size_kb * 1024)))

    echo
    echo -e "${BOLD}Organization Summary${NC}"
    echo -e "  Repositories: ${CYAN}$repo_count${NC}"
    echo -e "  Total Size:   ${CYAN}$total_size${NC}"
    echo
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo

    # Clone each repo
    local success=0
    local failed=0
    local skipped=0
    local current=0

    echo "$repos_json" | jq -r '.[].nameWithOwner' | while IFS= read -r repo; do
        [[ -z "$repo" ]] && continue
        ((current++))

        local lang=$(echo "$repos_json" | jq -r --arg r "$repo" '.[] | select(.nameWithOwner == $r) | .primaryLanguage.name // "Unknown"')

        echo -e "${BOLD}[$current/$repo_count]${NC} $repo ${DIM}($lang)${NC}"

        if clone_repo "$repo"; then
            ((success++))
        else
            ((failed++))
        fi
        echo
    done

    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${GREEN}✓ Complete${NC}: $success cloned, $failed failed"
    echo
    echo -e "Run scanners with: ${CYAN}./zero.sh scan --org $org${NC}"
}

#############################################################################
# Main
#############################################################################

main() {
    parse_args "$@"

    if [[ "$ORG_MODE" == "true" ]]; then
        clone_org "$ORG_NAME"
    else
        clone_single "$TARGET"
    fi
}

main "$@"
