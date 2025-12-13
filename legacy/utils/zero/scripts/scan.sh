#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Phantom Scan
# Run scanners to gather enrichment data from cloned repositories
#
# Usage:
#   ./scan.sh <owner/repo>           # Single repo
#   ./scan.sh --org <org-name>       # All cloned repos in an org
#
# Examples:
#   ./scan.sh expressjs/express
#   ./scan.sh expressjs/express --quick
#   ./scan.sh --org expressjs --standard
#############################################################################

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ZERO_UTILS_DIR="$(dirname "$SCRIPT_DIR")"

# Load Zero library (sets ZERO_DIR to .zero data directory in project root)
source "$ZERO_UTILS_DIR/lib/zero-lib.sh"

# Load config loader for dynamic profiles
source "$ZERO_UTILS_DIR/config/config-loader.sh"

# Load .env if available
UTILS_ROOT="$(dirname "$ZERO_UTILS_DIR")"
REPO_ROOT="$(dirname "$UTILS_ROOT")"
SCANNERS_DIR="$UTILS_ROOT/scanners"

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
TARGET=""
PROFILE="$(get_default_profile)"
FORCE=false

# All scanners loaded from config (with fallback)
ALL_SCANNERS="$(get_all_scanners)"

#############################################################################
# Usage
#############################################################################

usage() {
    cat << EOF
Phantom Scan - Run scanners to gather enrichment data

Usage: $0 <target> [options]
       $0 --org <org-name> [options]

MODES:
    Single Repo:    $0 owner/repo [options]
    Organization:   $0 --org <org-name> [options]

PROFILES (from zero.config.json):
EOF
    print_profile_help
    cat << EOF

OPTIONS:
    --org <name>    Scan all cloned repos in a GitHub organization
    --force         Re-scan even if results exist
    -h, --help      Show this help

EXAMPLES:
    $0 expressjs/express                    # Standard scan
    $0 expressjs/express --quick            # Quick scan
    $0 --org expressjs --security           # Security scan all org repos

EOF
    exit 0
}

#############################################################################
# Argument Parsing
#############################################################################

parse_args() {
    # Get available profiles for dynamic matching
    local available_profiles=$(get_available_profiles)

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
            --force)
                FORCE=true
                shift
                ;;
            --*)
                # Try to match as a profile name
                local profile_name="${1#--}"
                if [[ " $available_profiles " =~ " $profile_name " ]]; then
                    PROFILE="$profile_name"
                    # Enable Claude if profile requires it
                    if profile_uses_claude "$profile_name"; then
                        export USE_CLAUDE=true
                    fi
                    shift
                else
                    echo -e "${RED}Error: Unknown option $1${NC}" >&2
                    echo -e "Available profiles: $available_profiles"
                    exit 1
                fi
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
    fi
}

#############################################################################
# Scanner Functions
#############################################################################

# Check if scanner is in current profile (uses dynamic config)
# Note: scanner_in_profile is already defined in config-loader.sh

# Get scanner display name (uses dynamic config)
get_scanner_display() {
    local scanner="$1"
    get_scanner_name "$scanner"
}

# Get scanner output file (uses dynamic config)
get_scanner_output() {
    local scanner="$1"
    local analysis_path="$2"
    local output_file=$(get_scanner_output_file "$scanner")
    echo "$analysis_path/$output_file"
}

# Run a single scanner
run_scanner() {
    local scanner="$1"
    local repo_path="$2"
    local analysis_path="$3"

    # Get script path from config (handles scanners in subdirectories)
    local script_rel=$(get_scanner_script "$scanner")
    local script_path="$REPO_ROOT/$script_rel"
    local output_file=$(get_scanner_output "$scanner" "$analysis_path")

    # Check if scanner exists
    if [[ ! -x "$script_path" ]]; then
        return 1
    fi

    # Build args
    local args=("--local-path" "$repo_path")

    # Pass SBOM for scanners that need it
    if [[ -f "$analysis_path/sbom.cdx.json" ]]; then
        case "$scanner" in
            tech-discovery|package-vulns|package-health|licenses)
                args+=("--sbom" "$analysis_path/sbom.cdx.json")
                ;;
        esac
    fi

    # Output file
    args+=("-o" "$output_file")

    # Run scanner
    "$script_path" "${args[@]}" 2>/dev/null
}

# Get result summary from scanner output
get_scanner_result() {
    local scanner="$1"
    local analysis_path="$2"
    local output_file=$(get_scanner_output "$scanner" "$analysis_path")

    if [[ ! -f "$output_file" ]]; then
        echo ""
        return
    fi

    case "$scanner" in
        package-sbom)
            local count=$(jq -r '.components | length // 0' "$output_file" 2>/dev/null)
            echo "$count packages"
            ;;
        tech-discovery)
            local count=$(jq -r '.technologies | length // 0' "$output_file" 2>/dev/null)
            echo "$count technologies"
            ;;
        package-vulns)
            local count=$(jq -r '.summary.total // .vulnerabilities | length // 0' "$output_file" 2>/dev/null)
            echo "$count found"
            ;;
        licenses)
            local status=$(jq -r '.summary.status // "unknown"' "$output_file" 2>/dev/null)
            echo "$status"
            ;;
        code-security)
            local count=$(jq -r '.summary.total // .findings | length // 0' "$output_file" 2>/dev/null)
            echo "$count findings"
            ;;
        code-secrets)
            local count=$(jq -r '.summary.total // .secrets | length // 0' "$output_file" 2>/dev/null)
            echo "$count found"
            ;;
        tech-debt)
            local score=$(jq -r '.summary.score // "unknown"' "$output_file" 2>/dev/null)
            echo "score: $score"
            ;;
        code-ownership)
            local owners=$(jq -r '.summary.total_owners // 0' "$output_file" 2>/dev/null)
            echo "$owners owners"
            ;;
        dora)
            local freq=$(jq -r '.summary.deployment_frequency // "unknown"' "$output_file" 2>/dev/null)
            echo "$freq"
            ;;
        package-malcontent)
            local files=$(jq -r '.summary.total_files // 0' "$output_file" 2>/dev/null)
            local critical=$(jq -r '.summary.by_risk.critical // 0' "$output_file" 2>/dev/null)
            local high=$(jq -r '.summary.by_risk.high // 0' "$output_file" 2>/dev/null)
            if [[ "$critical" != "0" ]] || [[ "$high" != "0" ]]; then
                echo "$files files, ${critical}C/${high}H"
            else
                echo "$files files"
            fi
            ;;
        *)
            echo "complete"
            ;;
    esac
}

#############################################################################
# Scan Functions
#############################################################################

# Scan a single repository using bootstrap.sh (which has proven scanner implementations)
scan_repo() {
    local repo="$1"
    local project_id=$(zero_project_id "$repo")
    local repo_path="$ZERO_PROJECTS_DIR/$project_id/repo"

    # Check if repo is cloned
    if [[ ! -d "$repo_path" ]]; then
        echo -e "  ${RED}✗${NC} Repository not cloned"
        echo -e "    Run: ${CYAN}./zero.sh clone $repo${NC}"
        return 1
    fi

    # Build bootstrap args
    local bootstrap_args=("--scan-only" "--$PROFILE")
    [[ "$FORCE" == "true" ]] && bootstrap_args+=("--force")
    bootstrap_args+=("$repo")

    # Delegate to bootstrap.sh which has the proven scanner implementations
    "$SCRIPT_DIR/bootstrap.sh" "${bootstrap_args[@]}"
}

# Update analysis manifest
update_manifest() {
    local project_id="$1"
    local analysis_path="$2"
    local manifest="$analysis_path/manifest.json"

    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Build analyses object
    local analyses="{}"
    for scanner in $ALL_SCANNERS; do
        local output_file=$(get_scanner_output "$scanner" "$analysis_path")
        if [[ -f "$output_file" ]]; then
            local mtime=$(stat -f %m "$output_file" 2>/dev/null || stat -c %Y "$output_file" 2>/dev/null)
            analyses=$(echo "$analyses" | jq --arg s "$scanner" --arg t "$timestamp" '. + {($s): {"status": "complete", "completed_at": $t}}')
        fi
    done

    # Write manifest
    jq -n \
        --arg pid "$project_id" \
        --arg mode "$PROFILE" \
        --arg ts "$timestamp" \
        --argjson analyses "$analyses" \
        '{
            project_id: $pid,
            mode: $mode,
            completed_at: $ts,
            analyses: $analyses
        }' > "$manifest"
}

#############################################################################
# Main Functions
#############################################################################

scan_single() {
    local repo="$1"
    local project_id=$(zero_project_id "$repo")
    local repo_path="$ZERO_PROJECTS_DIR/$project_id/repo"

    # Check if repo is cloned
    if [[ ! -d "$repo_path" ]]; then
        print_zero_banner
        echo -e "${RED}Error: Repository not cloned${NC}"
        echo -e "Run: ${CYAN}./zero.sh clone $repo${NC}"
        return 1
    fi

    # Build bootstrap args and delegate to bootstrap.sh
    local bootstrap_args=("--scan-only" "--$PROFILE")
    [[ "$FORCE" == "true" ]] && bootstrap_args+=("--force")
    bootstrap_args+=("$repo")

    exec "$SCRIPT_DIR/bootstrap.sh" "${bootstrap_args[@]}"
}

scan_org() {
    local org="$1"

    print_zero_banner
    echo -e "${BOLD}Scan Organization${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo

    # Find cloned repos in org
    local org_path="$ZERO_PROJECTS_DIR/$org"
    if [[ ! -d "$org_path" ]]; then
        echo -e "${RED}No cloned repos found for org: $org${NC}" >&2
        echo -e "Clone first with: ${CYAN}./zero.sh clone --org $org${NC}"
        exit 1
    fi

    local repos=()
    for repo_dir in "$org_path"/*/; do
        [[ ! -d "$repo_dir" ]] && continue
        [[ ! -d "$repo_dir/repo" ]] && continue
        local repo_name=$(basename "$repo_dir")
        repos+=("$org/$repo_name")
    done

    local repo_count=${#repos[@]}
    if [[ $repo_count -eq 0 ]]; then
        echo -e "${RED}No cloned repos found in: $org_path${NC}" >&2
        exit 1
    fi

    # Get parallel jobs from config
    local parallel_jobs=$(get_parallel_jobs)

    echo -e "Organization: ${CYAN}$org${NC}"
    echo -e "Repositories: ${CYAN}$repo_count${NC}"
    echo -e "Profile:      ${CYAN}$PROFILE${NC}"
    echo -e "Parallel:     ${CYAN}$parallel_jobs jobs${NC}"
    echo
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    # Check if GNU parallel is available for better parallelization
    if command -v parallel &> /dev/null && [[ "$parallel_jobs" -gt 1 ]]; then
        scan_org_parallel "$org" "${repos[@]}"
    else
        scan_org_sequential "$org" "${repos[@]}"
    fi
}

# Scan org repos with progress bar and grouped output
scan_org_parallel() {
    local org="$1"
    shift
    local repos=("$@")
    local repo_count=${#repos[@]}
    local parallel_jobs=$(get_parallel_jobs)

    echo -e "${CYAN}Scanning $repo_count repositories with $parallel_jobs concurrent jobs${NC}"
    echo -e "${DIM}Progress bar mode • Grouped output • Press Ctrl+C to cancel${NC}"
    echo
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo

    # Create temp dirs
    local tmp_dir=$(mktemp -d)
    local status_dir=$(mktemp -d)
    local buffer_dir=$(init_output_buffer)

    # Initialize dashboard
    init_org_scan_dashboard "${repos[*]}" "$status_dir"

    local completed_count=0

    # Start background scans with status tracking
    local pids=()
    local running=0
    local next_idx=0

    # Setup cleanup handler for background processes
    cleanup_org_scan() {
        local exit_code=$?

        # Clear any progress bar
        clear_progress_line 2>/dev/null || true

        # If interrupted (Ctrl+C), show cancellation message
        if [[ $exit_code -eq 130 ]] || [[ "${SCAN_INTERRUPTED:-}" == "true" ]]; then
            echo -e "\n${YELLOW}⚠${NC} Scan cancelled by user" >&2
        fi

        # Kill all background scan jobs
        for pid in "${pids[@]}"; do
            kill "$pid" 2>/dev/null || true
        done
        # Wait briefly for jobs to terminate
        sleep 0.2
        # Force kill any remaining jobs
        for pid in "${pids[@]}"; do
            kill -9 "$pid" 2>/dev/null || true
        done
        # Clean up temp directories
        rm -rf "$tmp_dir" "$status_dir" "$buffer_dir" 2>/dev/null || true
    }

    # Handle Ctrl+C gracefully
    handle_interrupt() {
        SCAN_INTERRUPTED=true
        exit 130
    }

    trap cleanup_org_scan EXIT
    trap handle_interrupt INT TERM

    # Function to monitor a scan and buffer output
    monitor_scan() {
        local repo="$1"
        local tmp_dir="$2"
        local status_dir="$3"
        local buffer_dir="$4"
        local idx="$5"

        local start_time=$(date +%s)
        update_repo_scan_status "$status_dir" "$repo" "running" "starting" "" "0"

        # Initialize buffer for this repo
        start_buffer "$buffer_dir" "$repo"

        # Run scan and capture output
        local scan_output=$(mktemp)
        "$SCRIPT_DIR/bootstrap.sh" --scan-only --"$PROFILE" $([ "$FORCE" == "true" ] && echo "--force") "$repo" > "$scan_output" 2>&1
        local exit_code=$?

        local end_time=$(date +%s)
        local duration=$((end_time - start_time))

        # Store output in buffer
        append_buffer "$buffer_dir" "$repo" "$(cat "$scan_output")"

        # Parse output for final summary
        local summary=""
        if [[ $exit_code -eq 0 ]]; then
            summary=$(grep -E "✓|complete" "$scan_output" 2>/dev/null | wc -l | tr -d ' ')
            summary="${summary} scanners"
            update_repo_scan_status "$status_dir" "$repo" "complete" "" "$summary" "$duration"
        else
            update_repo_scan_status "$status_dir" "$repo" "failed" "" "" "$duration"
        fi

        rm -f "$scan_output"
        echo "$exit_code" > "$tmp_dir/$idx.exit"
    }

    # Start initial batch
    while [[ $next_idx -lt ${#repos[@]} ]] && [[ $running -lt $parallel_jobs ]]; do
        local repo="${repos[$next_idx]}"
        monitor_scan "$repo" "$tmp_dir" "$status_dir" "$buffer_dir" "$next_idx" &
        pids+=($!)
        ((running++))
        ((next_idx++))
    done

    # Get total scanner count for the profile
    local profile_scanners=$(get_profile_scanners "$PROFILE")
    local total_scanners=$(echo "$profile_scanners" | wc -w | tr -d ' ')

    # Track which repos have been displayed
    local displayed_repos=""

    # Monitor with dashboard and display completed repos
    echo  # Initial spacing
    while repos_still_scanning "$status_dir" "${repos[*]}"; do
        # Count completed scanners
        completed_count=0
        for repo in "${repos[@]}"; do
            local safe_name=$(echo "$repo" | sed 's/\//__/g')
            local status_file="$status_dir/$safe_name.status"
            if [[ -f "$status_file" ]]; then
                local status=$(cut -d'|' -f1 "$status_file")
                if [[ "$status" == "complete" ]] || [[ "$status" == "failed" ]]; then
                    ((completed_count++))

                    # Display repo output if not already displayed
                    if [[ ! " $displayed_repos " =~ " $repo " ]]; then
                        # Clear dashboard before showing repo output
                        if [[ -n "${DASHBOARD_LINES:-}" ]]; then
                            clear_lines "$DASHBOARD_LINES"
                        fi

                        echo
                        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
                        echo -e "${BOLD}${CYAN}$repo${NC}"
                        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
                        flush_buffer "$buffer_dir" "$repo"
                        echo
                        displayed_repos="$displayed_repos $repo"

                        # Reset dashboard line tracking
                        unset DASHBOARD_LINES
                    fi
                fi
            fi
        done

        # Render dashboard
        render_scan_dashboard "$status_dir" "$total_scanners" "${repos[@]}"

        # Check for completed jobs and start new ones
        for i in "${!pids[@]}"; do
            local pid="${pids[$i]}"
            if ! kill -0 "$pid" 2>/dev/null; then
                unset 'pids[$i]'
                ((running--))

                # Start next scan if available
                if [[ $next_idx -lt ${#repos[@]} ]]; then
                    local repo="${repos[$next_idx]}"
                    monitor_scan "$repo" "$tmp_dir" "$status_dir" "$buffer_dir" "$next_idx" &
                    pids+=($!)
                    ((running++))
                    ((next_idx++))
                fi
            fi
        done

        # Compact pids array
        pids=("${pids[@]}")

        sleep 0.3
    done

    # Clear dashboard
    if [[ -n "${DASHBOARD_LINES:-}" ]]; then
        clear_lines "$DASHBOARD_LINES"
    fi

    # Display any remaining repos that haven't been shown yet
    for repo in "${repos[@]}"; do
        if [[ ! " $displayed_repos " =~ " $repo " ]]; then
            echo
            echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
            echo -e "${BOLD}${CYAN}$repo${NC}"
            echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
            flush_buffer "$buffer_dir" "$repo"
            echo
        fi
    done

    # Count results
    local success=0
    local failed=0
    for i in $(seq 0 $((repo_count - 1))); do
        if [[ -f "$tmp_dir/$i.exit" ]] && [[ $(cat "$tmp_dir/$i.exit") -eq 0 ]]; then
            ((success++))
        else
            ((failed++))
        fi
    done

    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${GREEN}✓ Complete${NC}: $success scanned, $failed failed"
    echo
    echo -e "View results: ${CYAN}./zero.sh report --org $org${NC}"
}

# Scan org repos sequentially with grouped output
scan_org_sequential() {
    local org="$1"
    shift
    local repos=("$@")
    local repo_count=${#repos[@]}
    local parallel_jobs=$(get_parallel_jobs)

    # Initialize todo-style display
    init_todo_display "$org" "${repos[@]}"

    # Reset progress display flag for this scan
    SCAN_STATUS_RENDERED=0
    TODO_DISPLAY_LINES=0

    local success=0
    local failed=0
    local current=0
    local completed_count=0
    local pids=()
    local repo_map=()

    # Create temp dirs
    local tmp_dir=$(mktemp -d)
    local buffer_dir=$(init_output_buffer)
    local status_dir=$(mktemp -d)

    # Track completed repos
    local displayed_repos=""

    # Setup cleanup handler for background processes
    cleanup_seq_scan() {
        local exit_code=$?

        # Clear any progress bar
        clear_progress_line 2>/dev/null || true

        # If interrupted (Ctrl+C), show cancellation message
        if [[ $exit_code -eq 130 ]] || [[ "${SCAN_INTERRUPTED:-}" == "true" ]]; then
            echo -e "\n${YELLOW}⚠${NC} Scan cancelled by user" >&2
        fi

        # Kill all background scan jobs
        for pid in "${pids[@]}"; do
            kill "$pid" 2>/dev/null || true
        done
        # Wait briefly for jobs to terminate
        sleep 0.2
        # Force kill any remaining jobs
        for pid in "${pids[@]}"; do
            kill -9 "$pid" 2>/dev/null || true
        done
        # Clean up temp directories
        rm -rf "$tmp_dir" "$buffer_dir" "$status_dir" 2>/dev/null || true
    }

    # Handle Ctrl+C gracefully
    handle_interrupt() {
        SCAN_INTERRUPTED=true
        exit 130
    }

    trap cleanup_seq_scan EXIT
    trap handle_interrupt INT TERM

    # Show initial todo display
    render_todo_display "$status_dir" "$org" "0" "${repos[@]}"

    # Track start time for elapsed counter
    local scan_start_time=$(date +%s)

    for repo in "${repos[@]}"; do
        ((current++))

        # Initialize buffer for this repo
        start_buffer "$buffer_dir" "$repo"

        # Initialize status for this repo
        update_repo_scan_status "$status_dir" "$repo" "running" "scanning" "" "0"

        # Start background job - capture output and update status
        (
            local scan_output=$(mktemp)
            local start_time=$(date +%s)

            "$SCRIPT_DIR/bootstrap.sh" --scan-only --"$PROFILE" $([ "$FORCE" == "true" ] && echo "--force") --status-dir "$status_dir" "$repo" > "$scan_output" 2>&1
            local exit_code=$?

            local end_time=$(date +%s)
            local duration=$((end_time - start_time))

            # Store output in buffer
            append_buffer "$buffer_dir" "$repo" "$(cat "$scan_output")"

            # Update final status
            if [[ $exit_code -eq 0 ]]; then
                local summary=$(grep -E "✓|complete" "$scan_output" 2>/dev/null | wc -l | tr -d ' ')
                update_repo_scan_status "$status_dir" "$repo" "complete" "" "${summary} scanners" "$duration"
            else
                update_repo_scan_status "$status_dir" "$repo" "failed" "" "" "$duration"
            fi

            rm -f "$scan_output"
            echo $exit_code > "$tmp_dir/$current.exit"
        ) &

        pids+=($!)
        repo_map+=("$repo")

        # Limit concurrent jobs
        if [[ ${#pids[@]} -ge $parallel_jobs ]]; then
            # Wait for ANY job to complete (not just the first one)
            # This ensures we maintain full parallelism
            local completed_pid=""
            local completed_index=-1

            # Poll with progress updates until any job completes
            while true; do
                local elapsed=$(($(date +%s) - scan_start_time))
                render_todo_display "$status_dir" "$org" "$elapsed" "${repos[@]}"

                # Check if any PID has completed
                for i in "${!pids[@]}"; do
                    if ! kill -0 "${pids[$i]}" 2>/dev/null; then
                        completed_pid="${pids[$i]}"
                        completed_index=$i
                        break 2  # Break out of both loops
                    fi
                done

                sleep 0.5
            done

            # Wait for the completed job to clean up
            wait "$completed_pid" 2>/dev/null || true

            # Track completion (todo display shows it automatically)
            local completed_repo="${repo_map[$completed_index]}"
            ((completed_count++))
            displayed_repos="$displayed_repos $completed_repo"

            # Remove completed job from arrays
            unset 'pids[$completed_index]'
            unset 'repo_map[$completed_index]'
            pids=("${pids[@]}")  # Re-index array
            repo_map=("${repo_map[@]}")  # Re-index array
        fi
    done

    # Wait for remaining jobs
    while [[ ${#pids[@]} -gt 0 ]]; do
        local elapsed=$(($(date +%s) - scan_start_time))
        render_todo_display "$status_dir" "$org" "$elapsed" "${repos[@]}"

        # Check each remaining job
        for i in "${!pids[@]}"; do
            local pid="${pids[$i]}"
            if ! kill -0 "$pid" 2>/dev/null; then
                # Job completed
                wait "$pid" 2>/dev/null || true
                local completed_repo="${repo_map[$i]}"
                ((completed_count++))
                displayed_repos="$displayed_repos $completed_repo"

                # Remove from arrays
                unset 'pids[$i]'
                unset 'repo_map[$i]'
            fi
        done

        # Compact arrays
        pids=("${pids[@]}")
        repo_map=("${repo_map[@]}")

        sleep 0.5
    done

    # Count results
    for i in $(seq 1 $repo_count); do
        if [[ -f "$tmp_dir/$i.exit" ]] && [[ $(cat "$tmp_dir/$i.exit") -eq 0 ]]; then
            ((success++))
        else
            ((failed++))
        fi
    done

    # Calculate total duration
    local total_elapsed=$(($(date +%s) - scan_start_time))
    local duration=$(format_duration $total_elapsed)

    # Show final summary with todo-style
    finalize_todo_display "$org" "$repo_count" "$success" "$failed" "$duration"
}

#############################################################################
# Main
#############################################################################

main() {
    parse_args "$@"

    if [[ "$ORG_MODE" == "true" ]]; then
        scan_org "$ORG_NAME"
    else
        scan_single "$TARGET"
    fi
}

main "$@"
