#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Phantom Hydrate
# Clone and scan repositories with unified progress display
#
# Usage:
#   ./hydrate.sh <owner/repo>           # Clone and scan single repo
#   ./hydrate.sh --org <org-name>       # Clone and scan all repos in org
#
# Examples:
#   ./hydrate.sh expressjs/express
#   ./hydrate.sh expressjs/express --quick
#   ./hydrate.sh --org expressjs --standard
#############################################################################

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ZERO_UTILS_DIR="$(dirname "$SCRIPT_DIR")"

# Load Zero library (sets ZERO_DIR to .zero data directory in project root)
source "$ZERO_UTILS_DIR/lib/zero-lib.sh"
source "$ZERO_UTILS_DIR/config/config-loader.sh"

#############################################################################
# Configuration
#############################################################################

ORG_MODE=false
ORG_NAME=""
TARGET=""
CLONE_ONLY=false
LIMIT=0
BRANCH=""
DEPTH=""
FORCE=false
PROFILE="standard"

#############################################################################
# Usage
#############################################################################

usage() {
    cat << EOF
Phantom Hydrate - Clone and scan repositories

Usage: $0 <target> [options]
       $0 --org <org-name> [options]

This command combines clone + scan into a single step.

MODES:
    Single Repo:    $0 owner/repo [options]
    Organization:   $0 --org <org-name> [options]

CLONE OPTIONS:
    --branch <name>     Clone specific branch
    --depth <n>         Shallow clone depth
    --clone-only        Clone without scanning

SCAN OPTIONS:
    --quick             Fast scan (~30s)
    --standard          Standard scan (~2min) [default]
    --advanced          Full scan (~5min)
    --deep              Deep scan with Claude (~10min)
    --security          Security-focused scan (~3min)
    --security-deep     Deep security analysis with Claude (~10min)
    --compliance        License and policy compliance (~2min)
    --devops            CI/CD and operational metrics (~3min)
    --malcontent        Supply chain compromise detection (~2min)

COMMON OPTIONS:
    --org <name>        Process all repos in organization
    --limit <n>         Max repos in org mode
    --force             Re-clone and re-scan
    -h, --help          Show this help

EXAMPLES:
    $0 expressjs/express                    # Clone + standard scan
    $0 expressjs/express --quick            # Clone + quick scan
    $0 --org expressjs --limit 10           # Clone + scan first 10 repos
    $0 expressjs/express --clone-only       # Clone only, no scan

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
            --clone-only)
                CLONE_ONLY=true
                shift
                ;;
            --quick)
                PROFILE="quick"
                shift
                ;;
            --standard)
                PROFILE="standard"
                shift
                ;;
            --advanced)
                PROFILE="advanced"
                shift
                ;;
            --deep)
                PROFILE="deep"
                shift
                ;;
            --security)
                PROFILE="security"
                shift
                ;;
            --security-deep)
                PROFILE="security-deep"
                shift
                ;;
            --compliance)
                PROFILE="compliance"
                shift
                ;;
            --devops)
                PROFILE="devops"
                shift
                ;;
            --malcontent)
                PROFILE="malcontent"
                shift
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
    if [[ "$ORG_MODE" != "true" ]] && [[ -z "$TARGET" ]]; then
        echo -e "${RED}Error: No target specified${NC}" >&2
        echo "Usage: $0 <owner/repo> or $0 --org <org-name>"
        exit 1
    fi
}

#############################################################################
# Clone Functions
#############################################################################

# Clone a single repo silently, return status
# Sets CLONE_STATUS to: cloned|skipped|failed
clone_repo_silent() {
    local current_repo="$1"
    local project_id=$(zero_project_id "$current_repo")
    local repo_path="$ZERO_PROJECTS_DIR/$project_id/repo"

    if [[ -d "$repo_path" ]] && [[ "$FORCE" != "true" ]]; then
        CLONE_STATUS="skipped"
        return 0
    fi

    # Remove existing if force
    if [[ -d "$repo_path" ]]; then
        rm -rf "$repo_path"
    fi

    mkdir -p "$ZERO_PROJECTS_DIR/$project_id"

    local clone_url="https://github.com/$current_repo.git"
    if [[ -n "${GITHUB_TOKEN:-}" ]]; then
        clone_url="https://${GITHUB_TOKEN}@github.com/$current_repo.git"
    fi

    local clone_args=("-q")
    [[ -n "$BRANCH" ]] && clone_args+=("--branch" "$BRANCH")
    [[ -n "$DEPTH" ]] && clone_args+=("--depth" "$DEPTH")

    if git clone "${clone_args[@]}" "$clone_url" "$repo_path" 2>/dev/null; then
        mkdir -p "$ZERO_PROJECTS_DIR/$project_id/analysis"
        CLONE_STATUS="cloned"
        return 0
    else
        CLONE_STATUS="failed"
        return 1
    fi
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

#############################################################################
# Hydrate Functions
#############################################################################

hydrate_single() {
    local current_repo="$1"

    print_zero_banner
    echo -e "${BOLD}Hydrate Repository${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo
    echo -e "Target: ${CYAN}$current_repo${NC}"
    echo

    # Clone
    echo -e "  ${CYAN}Cloning...${NC}"
    if clone_repo_silent "$current_repo"; then
        if [[ "$CLONE_STATUS" == "skipped" ]]; then
            echo -e "  ${GREEN}✓${NC} Already cloned"
        else
            echo -e "  ${GREEN}✓${NC} Cloned"
        fi
    else
        echo -e "  ${RED}✗${NC} Clone failed"
        return 1
    fi

    # Scan (unless clone-only)
    if [[ "$CLONE_ONLY" != "true" ]]; then
        echo
        echo -e "  ${CYAN}Scanning...${NC}"
        local force_arg=""
        [[ "$FORCE" == "true" ]] && force_arg="--force"
        "$SCRIPT_DIR/bootstrap.sh" --scan-only --"$PROFILE" $force_arg "$current_repo"
    fi

    echo
    echo -e "${GREEN}✓ Hydrate complete${NC}"
    echo -e "View results: ${CYAN}./zero.sh report $current_repo${NC}"
}

hydrate_org() {
    local org="$1"

    # Check gh CLI
    if ! command -v gh &> /dev/null; then
        echo -e "${RED}Error: GitHub CLI (gh) is required for --org mode${NC}" >&2
        exit 1
    fi

    echo -e "${BLUE}Fetching repositories for ${CYAN}$org${BLUE}...${NC}"
    local repos_json=$(fetch_org_repos "$org" "$LIMIT")

    if [[ -z "$repos_json" ]] || [[ "$repos_json" == "[]" ]]; then
        echo -e "${RED}No repos found for organization: $org${NC}" >&2
        exit 1
    fi

    local repo_count=$(echo "$repos_json" | jq 'length')

    # Get repos into array
    local repos=()
    while IFS= read -r r; do
        [[ -n "$r" ]] && repos+=("$r")
    done < <(echo "$repos_json" | jq -r '.[].nameWithOwner')

    local parallel_jobs=$(get_parallel_jobs)

    # Initialize display with header
    init_todo_display "$org" "$repo_count" "$PROFILE" "$parallel_jobs"

    local success=0
    local failed=0
    local pids=()
    local repo_map=()

    # Create temp dirs
    local tmp_dir=$(mktemp -d)
    local buffer_dir=$(init_output_buffer)
    local status_dir=$(mktemp -d)

    # Setup cleanup handler
    cleanup_hydrate() {
        for pid in "${pids[@]}"; do
            kill "$pid" 2>/dev/null || true
        done
        sleep 0.2
        for pid in "${pids[@]}"; do
            kill -9 "$pid" 2>/dev/null || true
        done
        rm -rf "$tmp_dir" "$buffer_dir" "$status_dir" 2>/dev/null || true
    }

    trap cleanup_hydrate EXIT
    trap 'exit 130' INT TERM

    # Show initial display
    local start_time=$(date +%s)
    render_todo_display "$status_dir" "$org" "0" "${repos[@]}"

    local current=0
    for current_repo in "${repos[@]}"; do
        ((current++))

        # Initialize buffer
        start_buffer "$buffer_dir" "$current_repo"

        # Start background job: clone then scan
        (
            local scan_output=$(mktemp)
            local job_start=$(date +%s)

            # First clone
            local clone_status="skipped"
            local project_id=$(zero_project_id "$current_repo")
            local repo_path="$ZERO_PROJECTS_DIR/$project_id/repo"

            if [[ ! -d "$repo_path" ]] || [[ "$FORCE" == "true" ]]; then
                [[ -d "$repo_path" ]] && rm -rf "$repo_path"
                mkdir -p "$ZERO_PROJECTS_DIR/$project_id"

                local clone_url="https://github.com/$current_repo.git"
                [[ -n "${GITHUB_TOKEN:-}" ]] && clone_url="https://${GITHUB_TOKEN}@github.com/$current_repo.git"

                local clone_args=("-q")
                [[ -n "$BRANCH" ]] && clone_args+=("--branch" "$BRANCH")
                [[ -n "$DEPTH" ]] && clone_args+=("--depth" "$DEPTH")

                if git clone "${clone_args[@]}" "$clone_url" "$repo_path" 2>/dev/null; then
                    mkdir -p "$ZERO_PROJECTS_DIR/$project_id/analysis"
                    clone_status="cloned"
                else
                    clone_status="failed"
                fi
            fi

            # Update status with clone info
            update_repo_scan_status "$status_dir" "$current_repo" "running" "cloning" "" "0" "$clone_status"

            # If clone failed, mark as failed
            if [[ "$clone_status" == "failed" ]]; then
                local duration=$(($(date +%s) - job_start))
                update_repo_scan_status "$status_dir" "$current_repo" "failed" "" "" "$duration" "$clone_status"
                echo "1" > "$tmp_dir/$current.exit"
                exit 1
            fi

            # Skip scan if clone-only
            if [[ "$CLONE_ONLY" == "true" ]]; then
                local duration=$(($(date +%s) - job_start))
                update_repo_scan_status "$status_dir" "$current_repo" "complete" "" "cloned" "$duration" "$clone_status"
                echo "0" > "$tmp_dir/$current.exit"
                exit 0
            fi

            # Update status to scanning
            update_repo_scan_status "$status_dir" "$current_repo" "running" "scanning" "" "0" "$clone_status"

            # Run scan
            local force_arg=""
            [[ "$FORCE" == "true" ]] && force_arg="--force"
            "$SCRIPT_DIR/bootstrap.sh" --scan-only --"$PROFILE" $force_arg --status-dir "$status_dir" "$current_repo" > "$scan_output" 2>&1
            local exit_code=$?

            local duration=$(($(date +%s) - job_start))

            # Store output
            append_buffer "$buffer_dir" "$current_repo" "$(cat "$scan_output")"

            # Update final status
            if [[ $exit_code -eq 0 ]]; then
                local summary=$(grep -E "✓|complete" "$scan_output" 2>/dev/null | wc -l | tr -d ' ')
                update_repo_scan_status "$status_dir" "$current_repo" "complete" "" "${summary} scanners" "$duration" "$clone_status"
            else
                update_repo_scan_status "$status_dir" "$current_repo" "failed" "" "" "$duration" "$clone_status"
            fi

            rm -f "$scan_output"
            echo $exit_code > "$tmp_dir/$current.exit"
        ) &

        pids+=($!)
        repo_map+=("$current_repo")

        # Limit concurrent jobs
        if [[ ${#pids[@]} -ge $parallel_jobs ]]; then
            while true; do
                local elapsed=$(($(date +%s) - start_time))
                render_todo_display "$status_dir" "$org" "$elapsed" "${repos[@]}"

                for i in "${!pids[@]}"; do
                    if ! kill -0 "${pids[$i]}" 2>/dev/null; then
                        wait "${pids[$i]}" 2>/dev/null || true
                        local exit_file="$tmp_dir/$((i+1)).exit"
                        if [[ -f "$exit_file" ]] && [[ "$(cat "$exit_file")" == "0" ]]; then
                            ((success++))
                        else
                            ((failed++))
                        fi
                        unset 'pids[i]'
                        pids=("${pids[@]}")
                        unset 'repo_map[i]'
                        repo_map=("${repo_map[@]}")
                        break 2
                    fi
                done
                sleep 0.3
            done
        fi
    done

    # Wait for remaining jobs
    while [[ ${#pids[@]} -gt 0 ]]; do
        local elapsed=$(($(date +%s) - start_time))
        render_todo_display "$status_dir" "$org" "$elapsed" "${repos[@]}"

        for i in "${!pids[@]}"; do
            if ! kill -0 "${pids[$i]}" 2>/dev/null; then
                wait "${pids[$i]}" 2>/dev/null || true
                local exit_file="$tmp_dir/$((i+1)).exit"
                if [[ -f "$exit_file" ]] && [[ "$(cat "$exit_file")" == "0" ]]; then
                    ((success++))
                else
                    ((failed++))
                fi
                unset 'pids[i]'
                pids=("${pids[@]}")
                break
            fi
        done
        sleep 0.3
    done

    # Final display
    local total_elapsed=$(($(date +%s) - start_time))
    local duration=$(format_duration $total_elapsed)
    finalize_todo_display "$org" "$repo_count" "$success" "$failed" "$duration"
}

#############################################################################
# Main
#############################################################################

main() {
    parse_args "$@"

    if [[ "$ORG_MODE" == "true" ]]; then
        hydrate_org "$ORG_NAME"
    else
        hydrate_single "$TARGET"
    fi
}

main "$@"
