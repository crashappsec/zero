#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Phantom Hydrate
# Clone repository/repositories and run analyzers for agent queries
#
# Usage:
#   ./hydrate.sh <owner/repo>           # Single repo
#   ./hydrate.sh --org <org-name>       # All repos in an org
#
# Examples:
#   ./hydrate.sh expressjs/express
#   ./hydrate.sh --org expressjs
#   ./hydrate.sh --org my-company --limit 10
#############################################################################

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Load Gibson library
source "$SCRIPT_DIR/lib/gibson.sh"

# Load .env if available
UTILS_ROOT="$(dirname "$SCRIPT_DIR")"
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
CLEAN_MODE=false
PASS_THROUGH_ARGS=()

#############################################################################
# Usage
#############################################################################

usage() {
    cat << EOF
Phantom Hydrate - Clone and analyze repositories for agent queries

Usage: $0 <target> [options]
       $0 --org <org-name> [options]

MODES:
    Single Repo:    $0 owner/repo [options]
    Organization:   $0 --org <org-name> [options]

OPTIONS:
    --org <name>        Hydrate all repos in a GitHub organization
    --limit <n>         Max repos to hydrate in org mode (default: all)
    --branch <name>     Clone specific branch (default: default branch)
    --quick             Fast analyzers only (skip code-security, dora)
    --security-only     Security analyzers only
    --force             Re-hydrate even if project exists
    --clean             Remove ALL hydrated data before starting (fresh start)
    -h, --help          Show this help

EXAMPLES:
    $0 expressjs/express                    # Single repo
    $0 expressjs/express --quick            # Single repo, quick mode
    $0 --org expressjs                      # All repos in expressjs org
    $0 --org my-company --limit 10          # First 10 repos in org
    $0 --org my-company --quick --force     # Re-hydrate all, quick mode

REQUIREMENTS:
    - GitHub CLI (gh) must be installed and authenticated for --org mode
    - Run ./preflight.sh first to verify all tools are ready

EOF
    exit 0
}

#############################################################################
# Helper Functions
#############################################################################

# Extract org name from various URL formats
# Supports:
#   - https://github.com/orgs/org-name/
#   - https://github.com/orgs/org-name
#   - https://github.com/org-name
#   - org-name
extract_org_name() {
    local input="$1"

    # Remove trailing slash
    input="${input%/}"

    # https://github.com/orgs/org-name
    if [[ "$input" =~ github\.com/orgs/([^/]+) ]]; then
        echo "${BASH_REMATCH[1]}"
        return 0
    fi

    # https://github.com/org-name (no repo specified)
    if [[ "$input" =~ github\.com/([^/]+)$ ]]; then
        echo "${BASH_REMATCH[1]}"
        return 0
    fi

    # Plain org name
    if [[ ! "$input" =~ / ]] && [[ ! "$input" =~ github\.com ]]; then
        echo "$input"
        return 0
    fi

    # Not an org URL
    return 1
}

# Check if input looks like an org URL (not a repo)
is_org_url() {
    local input="$1"

    # Explicit /orgs/ path
    if [[ "$input" =~ github\.com/orgs/ ]]; then
        return 0
    fi

    # github.com/something with no repo part
    if [[ "$input" =~ ^https?://github\.com/([^/]+)/?$ ]]; then
        return 0
    fi

    return 1
}

#############################################################################
# Org Mode Functions
#############################################################################

# Check gh CLI is available and authenticated
check_gh_cli() {
    if ! command -v gh &> /dev/null; then
        echo -e "${RED}Error: GitHub CLI (gh) is required for --org mode${NC}" >&2
        echo -e "Install with: ${CYAN}brew install gh${NC}" >&2
        exit 1
    fi

    if ! gh auth status &> /dev/null; then
        echo -e "${RED}Error: GitHub CLI not authenticated${NC}" >&2
        echo -e "Run: ${CYAN}gh auth login${NC}" >&2
        exit 1
    fi
}

# Fetch org repos with metadata (name, size, language, updated)
fetch_org_repos_with_stats() {
    local org="$1"
    local limit="$2"

    local gh_limit=1000
    [[ $limit -gt 0 ]] && gh_limit=$limit

    gh repo list "$org" \
        --json nameWithOwner,diskUsage,primaryLanguage,pushedAt \
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

# Progress bar function
progress_bar() {
    local current=$1
    local total=$2
    local width=30
    local percent=$((current * 100 / total))
    local filled=$((current * width / total))
    local empty=$((width - filled))

    printf "["
    printf "%${filled}s" | tr ' ' '█'
    printf "%${empty}s" | tr ' ' '░'
    printf "] %3d%%" "$percent"
}

# Clear current line and move cursor to beginning
clear_line() {
    printf "\r\033[K"
}

# Glow effect based on time (creates pulsing animation)
get_glow() {
    local tick="$1"
    local frame=$((tick % 4))
    case $frame in
        0) printf "\033[0;36m●\033[0m" ;;   # dim cyan dot
        1) printf "\033[1;36m◉\033[0m" ;;   # bright cyan ring
        2) printf "\033[1;37m●\033[0m" ;;   # bright white dot
        3) printf "\033[1;36m◉\033[0m" ;;   # bright cyan ring
    esac
}

# Get display name for a phase
get_phase_display() {
    local phase="$1"
    case "$phase" in
        "Cloning") echo "Cloning repository" ;;
        "technology") echo "Technology scan" ;;
        "dependencies") echo "SBOM generation" ;;
        "vulnerabilities") echo "Package vulnerabilities" ;;
        "package-health") echo "Package health" ;;
        "licenses") echo "License scan" ;;
        "security-findings") echo "Code security" ;;
        "ownership") echo "Code ownership" ;;
        "dora") echo "DORA metrics" ;;
        *) echo "$phase" ;;
    esac
}

# Get time estimate for a phase
get_phase_estimate() {
    local phase="$1"
    case "$phase" in
        "Cloning") echo "~30s" ;;
        "technology") echo "~5s" ;;
        "dependencies") echo "~3s" ;;
        "vulnerabilities") echo "~30s" ;;
        "package-health") echo "~10s" ;;
        "licenses") echo "~5s" ;;
        "security-findings") echo "~45s" ;;
        "ownership") echo "~15s" ;;
        "dora") echo "~20s" ;;
        *) echo "~10s" ;;
    esac
}

# Get result summary for a completed phase
# Uses printf to properly handle escape sequences
get_phase_result() {
    local phase="$1"
    local analysis_path="$2"
    local project_id="$3"
    local repo_path="$GIBSON_PROJECTS_DIR/$project_id/repo"

    case "$phase" in
        "Cloning")
            # Show repo stats after cloning
            if [[ -d "$repo_path" ]]; then
                local size=$(du -sh "$repo_path" 2>/dev/null | cut -f1)
                local files=$(find "$repo_path" -type f ! -path '*/.git/*' 2>/dev/null | wc -l | tr -d ' ')
                printf "\033[2m%s, %s files\033[0m" "$size" "$files"
            fi
            ;;
        "technology")
            if [[ -f "$analysis_path/technology.json" ]]; then
                local count=$(jq -r '.technologies | length // 0' "$analysis_path/technology.json" 2>/dev/null)
                printf "\033[2m%s technologies detected\033[0m" "$count"
            fi
            ;;
        "dependencies")
            if [[ -f "$analysis_path/dependencies.json" ]]; then
                local total=$(jq -r '.total_dependencies // 0' "$analysis_path/dependencies.json" 2>/dev/null)
                local format=$(jq -r '.sbom_format // "unknown"' "$analysis_path/dependencies.json" 2>/dev/null)
                printf "\033[2m%s packages (%s)\033[0m" "$total" "$format"
            fi
            ;;
        "vulnerabilities")
            if [[ -f "$analysis_path/vulnerabilities.json" ]]; then
                local c=$(jq -r '.summary.critical // 0' "$analysis_path/vulnerabilities.json" 2>/dev/null)
                local h=$(jq -r '.summary.high // 0' "$analysis_path/vulnerabilities.json" 2>/dev/null)
                local m=$(jq -r '.summary.medium // 0' "$analysis_path/vulnerabilities.json" 2>/dev/null)
                if [[ "$c" == "0" ]] && [[ "$h" == "0" ]]; then
                    printf "\033[0;32mclean\033[0m"
                else
                    printf "\033[0;31m%s critical\033[0m, \033[1;33m%s high\033[0m, \033[2m%s medium\033[0m" "$c" "$h" "$m"
                fi
            fi
            ;;
        "package-health")
            if [[ -f "$analysis_path/package-health.json" ]]; then
                local abandoned=$(jq -r '.abandoned | length // 0' "$analysis_path/package-health.json" 2>/dev/null)
                local typosquat=$(jq -r '.typosquat_suspects | length // 0' "$analysis_path/package-health.json" 2>/dev/null)
                if [[ "$abandoned" == "0" ]] && [[ "$typosquat" == "0" ]]; then
                    printf "\033[0;32mhealthy\033[0m"
                else
                    local first=true
                    if [[ "$abandoned" != "0" ]]; then
                        printf "\033[1;33m%s abandoned\033[0m" "$abandoned"
                        first=false
                    fi
                    if [[ "$typosquat" != "0" ]]; then
                        [[ "$first" == "false" ]] && printf ", "
                        printf "\033[0;31m%s suspect\033[0m" "$typosquat"
                    fi
                fi
            fi
            ;;
        "licenses")
            if [[ -f "$analysis_path/licenses.json" ]]; then
                local restrictive=$(jq -r '.restrictive | length // 0' "$analysis_path/licenses.json" 2>/dev/null)
                local unknown=$(jq -r '.unknown | length // 0' "$analysis_path/licenses.json" 2>/dev/null)
                if [[ "$restrictive" == "0" ]] && [[ "$unknown" == "0" ]]; then
                    printf "\033[0;32mcompliant\033[0m"
                else
                    local first=true
                    if [[ "$restrictive" != "0" ]]; then
                        printf "\033[1;33m%s restrictive\033[0m" "$restrictive"
                        first=false
                    fi
                    if [[ "$unknown" != "0" ]]; then
                        [[ "$first" == "false" ]] && printf ", "
                        printf "\033[2m%s unknown\033[0m" "$unknown"
                    fi
                fi
            fi
            ;;
        "security-findings")
            if [[ -f "$analysis_path/security-findings.json" ]]; then
                local high=$(jq -r '.findings | map(select(.severity == "high")) | length // 0' "$analysis_path/security-findings.json" 2>/dev/null)
                local medium=$(jq -r '.findings | map(select(.severity == "medium")) | length // 0' "$analysis_path/security-findings.json" 2>/dev/null)
                local secrets=$(jq -r '.secrets | length // 0' "$analysis_path/security-findings.json" 2>/dev/null)
                if [[ "$high" == "0" ]] && [[ "$secrets" == "0" ]]; then
                    printf "\033[0;32msecure\033[0m"
                else
                    local first=true
                    if [[ "$high" != "0" ]]; then
                        printf "\033[0;31m%s high risk\033[0m" "$high"
                        first=false
                    fi
                    if [[ "$secrets" != "0" ]]; then
                        [[ "$first" == "false" ]] && printf ", "
                        printf "\033[0;31m%s secrets\033[0m" "$secrets"
                        first=false
                    fi
                    if [[ "$medium" != "0" ]]; then
                        [[ "$first" == "false" ]] && printf ", "
                        printf "\033[1;33m%s medium\033[0m" "$medium"
                    fi
                fi
            fi
            ;;
        "ownership")
            if [[ -f "$analysis_path/ownership.json" ]]; then
                local contributors=$(jq -r '.contributors | length // 0' "$analysis_path/ownership.json" 2>/dev/null)
                local bus_factor=$(jq -r '.bus_factor // 0' "$analysis_path/ownership.json" 2>/dev/null)
                printf "\033[2m%s contributors, bus factor: %s\033[0m" "$contributors" "$bus_factor"
            fi
            ;;
        "dora")
            if [[ -f "$analysis_path/dora.json" ]]; then
                local freq=$(jq -r '.deployment_frequency // "unknown"' "$analysis_path/dora.json" 2>/dev/null)
                local lead=$(jq -r '.lead_time // "unknown"' "$analysis_path/dora.json" 2>/dev/null)
                printf "\033[2mdeploy: %s, lead: %s\033[0m" "$freq" "$lead"
            fi
            ;;
    esac
}

# Clean all hydration data
clean_all_data() {
    echo -e "${YELLOW}Cleaning all hydration data...${NC}"

    if [[ -d "$GIBSON_PROJECTS_DIR" ]]; then
        # Count org directories (each org is a subdirectory)
        local project_count=$(find "$GIBSON_PROJECTS_DIR" -mindepth 2 -maxdepth 2 -type d 2>/dev/null | wc -l | tr -d ' ')

        # Remove all projects
        rm -rf "$GIBSON_PROJECTS_DIR"/*

        echo -e "${GREEN}✓${NC} Removed $project_count projects from ~/.phantom/projects"
    else
        echo -e "${YELLOW}No projects directory found at ~/.phantom/projects${NC}"
    fi
    echo
}

# Hydrate all repos in an organization
hydrate_org() {
    local org="$1"
    local limit="$2"
    shift 2
    local extra_args=("$@")

    # Check prerequisites
    check_gh_cli

    print_phantom_banner
    echo -e "${BOLD}Organization Mode${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo

    # Fetch repos with stats
    echo -e "${BLUE}Fetching repository information for ${CYAN}$org${BLUE}...${NC}"
    local repos_json=$(fetch_org_repos_with_stats "$org" "$limit")

    if [[ -z "$repos_json" ]] || [[ "$repos_json" == "[]" ]]; then
        echo -e "${RED}No repos found for organization: $org${NC}" >&2
        exit 1
    fi

    # Parse stats
    local repo_count=$(echo "$repos_json" | jq 'length')
    local total_size_kb=$(echo "$repos_json" | jq '[.[].diskUsage // 0] | add')
    local total_size=$(format_size $((total_size_kb * 1024)))
    local all_languages=$(echo "$repos_json" | jq -r '[.[].primaryLanguage.name // "Unknown"] | group_by(.) | map({lang: .[0], count: length}) | sort_by(-.count)')
    local lang_count=$(echo "$all_languages" | jq 'length')
    local languages=$(echo "$all_languages" | jq -r '.[0:5] | map("\(.lang): \(.count)") | join(", ")')
    [[ $lang_count -gt 5 ]] && languages="$languages (+$((lang_count - 5)) more)"

    # Show org summary
    echo
    echo -e "${BOLD}Organization Summary${NC}"
    echo -e "  Repositories:  ${CYAN}$repo_count${NC}"
    echo -e "  Total Size:    ${CYAN}$total_size${NC}"
    echo -e "  Languages:     ${CYAN}$languages${NC}"
    [[ $limit -gt 0 ]] && echo -e "  Limit:         ${CYAN}$limit repos${NC}"
    echo
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo

    # Extract repo names
    local repos=$(echo "$repos_json" | jq -r '.[].nameWithOwner')

    # Hydrate each repo
    local success=0
    local failed=0
    local skipped=0
    local current=0
    local total_start_time=$(date +%s)

    while IFS= read -r repo; do
        [[ -z "$repo" ]] && continue
        ((current++))

        local project_id=$(gibson_project_id "$repo")
        local repo_size_kb=$(echo "$repos_json" | jq -r --arg r "$repo" '.[] | select(.nameWithOwner == $r) | .diskUsage // 0')
        local repo_size=$(format_size $((repo_size_kb * 1024)))
        local repo_lang=$(echo "$repos_json" | jq -r --arg r "$repo" '.[] | select(.nameWithOwner == $r) | .primaryLanguage.name // "Unknown"')

        # Check if already hydrated (unless --force)
        if gibson_is_hydrated "$project_id" && [[ ! " ${extra_args[*]} " =~ " --force " ]]; then
            # Show cached results for already hydrated repos
            local analysis_path="$GIBSON_PROJECTS_DIR/$project_id/analysis"

            echo -e "${DIM}[$current/$repo_count]${NC} ${CYAN}●${NC} $repo ${DIM}(cached - use --force to re-analyze)${NC}"

            # Show all available analyzer results
            local results=""

            # SBOM / Dependencies
            if [[ -f "$analysis_path/dependencies.json" ]]; then
                local deps=$(jq -r '.total_dependencies // 0' "$analysis_path/dependencies.json" 2>/dev/null)
                local format=$(jq -r '.sbom_format // "unknown"' "$analysis_path/dependencies.json" 2>/dev/null)
                results+="  ${GREEN}✓${NC} SBOM ($format): $deps packages\n"
            fi

            # Package vulnerabilities
            if [[ -f "$analysis_path/vulnerabilities.json" ]]; then
                local c=$(jq -r '.summary.critical // 0' "$analysis_path/vulnerabilities.json" 2>/dev/null)
                local h=$(jq -r '.summary.high // 0' "$analysis_path/vulnerabilities.json" 2>/dev/null)
                if [[ "$c" == "0" ]] && [[ "$h" == "0" ]]; then
                    results+="  ${GREEN}✓${NC} Package vulnerabilities: ${GREEN}clean${NC}\n"
                else
                    results+="  ${GREEN}✓${NC} Package vulnerabilities: ${RED}$c critical${NC}, ${YELLOW}$h high${NC}\n"
                fi
            fi

            # Code security
            if [[ -f "$analysis_path/security-findings.json" ]]; then
                local issues=$(jq -r '.findings | length // 0' "$analysis_path/security-findings.json" 2>/dev/null)
                if [[ "$issues" == "0" ]]; then
                    results+="  ${GREEN}✓${NC} Code security: ${GREEN}clean${NC}\n"
                else
                    results+="  ${GREEN}✓${NC} Code security: ${YELLOW}$issues issues${NC}\n"
                fi
            fi

            # Licenses
            if [[ -f "$analysis_path/licenses.json" ]]; then
                local status=$(jq -r '.status // "unknown"' "$analysis_path/licenses.json" 2>/dev/null)
                if [[ "$status" == "pass" ]]; then
                    results+="  ${GREEN}✓${NC} Licenses: ${GREEN}pass${NC}\n"
                else
                    results+="  ${GREEN}✓${NC} Licenses: ${YELLOW}$status${NC}\n"
                fi
            fi

            # Technology
            if [[ -f "$analysis_path/technology.json" ]]; then
                local tech_count=$(jq -r '.technologies | length // 0' "$analysis_path/technology.json" 2>/dev/null)
                results+="  ${GREEN}✓${NC} Technology: $tech_count detected\n"
            fi

            # DORA
            if [[ -f "$analysis_path/dora.json" ]]; then
                local perf=$(jq -r '.performance_level // "unknown"' "$analysis_path/dora.json" 2>/dev/null)
                results+="  ${GREEN}✓${NC} DORA metrics: $perf\n"
            fi

            echo -e "$results"
            ((skipped++))
            continue
        fi

        # Show repo being processed
        echo -e "${BOLD}[$current/$repo_count]${NC} $repo ${DIM}($repo_size, $repo_lang)${NC}"

        # Run bootstrap and capture output for progress parsing
        local log_file=$(mktemp)
        local start_time=$(date +%s)

        # Run bootstrap in background
        "$SCRIPT_DIR/bootstrap.sh" "$repo" "${extra_args[@]}" > "$log_file" 2>&1 &
        local pid=$!

        # Monitor progress by watching log file
        local last_phase=""
        local phase_start_time=$start_time
        local analysis_path="$GIBSON_PROJECTS_DIR/$project_id/analysis"
        local tick=0

        while kill -0 $pid 2>/dev/null; do
            if [[ -f "$log_file" ]]; then
                # Check for current phase by looking at log and output files
                local clean_log=$(sed 's/\x1b\[[0-9;]*m//g' "$log_file" 2>/dev/null)
                local current_phase=""

                # Determine phase by checking what exists:
                # 1. If repo doesn't exist yet or Languages not printed -> Cloning
                # 2. If analysis started -> check which analyzer files exist

                if [[ ! -d "$GIBSON_PROJECTS_DIR/$project_id/repo" ]] || ! echo "$clean_log" | grep -q "Languages:"; then
                    current_phase="Cloning"
                else
                    # Check which output files exist to determine current analyzer
                    local analyzers="technology dependencies vulnerabilities package-health licenses security-findings ownership dora"
                    current_phase=""
                    for analyzer in $analyzers; do
                        local output_file="$analysis_path/${analyzer}.json"
                        if [[ ! -f "$output_file" ]]; then
                            # This analyzer hasn't completed yet - might be running
                            if echo "$clean_log" | grep -q "Running.*analyzers"; then
                                current_phase="$analyzer"
                                break
                            fi
                        fi
                    done
                fi

                local now=$(date +%s)
                local phase_elapsed=$((now - phase_start_time))

                if [[ -n "$current_phase" ]] && [[ "$current_phase" != "$last_phase" ]]; then
                    # Print result of completed phase (if any)
                    if [[ -n "$last_phase" ]]; then
                        clear_line
                        local display_name=$(get_phase_display "$last_phase")
                        local result=$(get_phase_result "$last_phase" "$analysis_path" "$project_id")
                        printf "  ${GREEN}✓${NC} %-22s %b\n" "$display_name" "$result"
                    fi

                    phase_start_time=$now
                    phase_elapsed=0
                    last_phase="$current_phase"
                fi

                # Update display with elapsed time, estimate, and glow effect
                if [[ -n "$current_phase" ]]; then
                    local estimate=$(get_phase_estimate "$current_phase")
                    local display=$(get_phase_display "$current_phase")
                    local glow=$(get_glow $tick)
                    clear_line
                    printf "  %s %-22s ${DIM}%ds${NC} ${DIM}(est: %s)${NC}" "$glow" "$display..." "$phase_elapsed" "$estimate"
                fi
            fi
            ((tick++))
            sleep 0.3
        done

        # Print result of final phase
        if [[ -n "$last_phase" ]]; then
            clear_line
            local display_name=$(get_phase_display "$last_phase")
            local result=$(get_phase_result "$last_phase" "$analysis_path" "$project_id")
            printf "  ${GREEN}✓${NC} %-22s %b\n" "$display_name" "$result"
        fi

        # Get exit status
        wait $pid
        local exit_status=$?
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))

        # Clear progress line
        clear_line

        if [[ $exit_status -eq 0 ]]; then
            echo -e "  ${GREEN}━━━ Complete${NC} ${DIM}(${duration}s total)${NC}"
            ((success++))
        else
            echo -e "  ${RED}✗${NC} Failed (${duration}s)"
            # Show error excerpt
            local error_line=$(tail -1 "$log_file" 2>/dev/null | head -c 60)
            [[ -n "$error_line" ]] && echo -e "    ${DIM}$error_line${NC}"
            ((failed++))
        fi

        rm -f "$log_file"
        echo

    done <<< "$repos"

    # Calculate total time
    local total_end_time=$(date +%s)
    local total_duration=$((total_end_time - total_start_time))
    local duration_min=$((total_duration / 60))
    local duration_sec=$((total_duration % 60))

    # Final Summary
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${BOLD}ORGANIZATION HYDRATION COMPLETE${NC} ${DIM}(static analysis - no AI)${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo
    echo -e "  ${GREEN}✓ Hydrated:${NC}  $success"
    echo -e "  ${YELLOW}⊘ Skipped:${NC}   $skipped"
    echo -e "  ${RED}✗ Failed:${NC}    $failed"
    echo -e "  ${CYAN}⏱ Duration:${NC}  ${duration_min}m ${duration_sec}s"
    echo

    # Show aggregate stats if we have hydrated repos
    if [[ $success -gt 0 ]]; then
        echo -e "${BOLD}Aggregate Analysis${NC}"

        local total_vulns=0
        local total_critical=0
        local total_high=0
        local total_deps=0

        for proj in $(gibson_list_hydrated | jq -r '.[]'); do
            local proj_path="$GIBSON_PROJECTS_DIR/$proj/analysis"
            if [[ -f "$proj_path/vulnerabilities.json" ]]; then
                total_critical=$((total_critical + $(jq -r '.summary.critical // 0' "$proj_path/vulnerabilities.json" 2>/dev/null)))
                total_high=$((total_high + $(jq -r '.summary.high // 0' "$proj_path/vulnerabilities.json" 2>/dev/null)))
            fi
            if [[ -f "$proj_path/dependencies.json" ]]; then
                total_deps=$((total_deps + $(jq -r '.total_dependencies // 0' "$proj_path/dependencies.json" 2>/dev/null)))
            fi
        done

        echo -e "  Total Dependencies:   ${CYAN}$total_deps${NC}"
        if [[ $total_critical -gt 0 ]] || [[ $total_high -gt 0 ]]; then
            echo -e "  Critical Vulns:       ${RED}$total_critical${NC}"
            echo -e "  High Vulns:           ${YELLOW}$total_high${NC}"
        else
            echo -e "  Vulnerabilities:      ${GREEN}No critical/high issues${NC}"
        fi
        echo
    fi

    echo -e "${BOLD}Hydrated Projects:${NC}"
    gibson_list_hydrated | jq -r '.[]' | sed 's/^/  /'
    echo
    echo -e "Query with: ${CYAN}/phantom ask scout <question>${NC}"
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
            --clean)
                CLEAN_MODE=true
                shift
                ;;
            --branch|--depth|--quick|--security-only|--force)
                PASS_THROUGH_ARGS+=("$1")
                if [[ "$1" == "--branch" ]] || [[ "$1" == "--depth" ]]; then
                    PASS_THROUGH_ARGS+=("$2")
                    shift
                fi
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

    # Validate args
    if [[ "$ORG_MODE" == true ]]; then
        if [[ -z "$ORG_NAME" ]]; then
            echo -e "${RED}Error: --org requires organization name${NC}" >&2
            exit 1
        fi
        # Normalize org name from URL if needed
        local extracted_org
        if extracted_org=$(extract_org_name "$ORG_NAME"); then
            ORG_NAME="$extracted_org"
        fi
    elif [[ -z "$TARGET" ]]; then
        echo -e "${RED}Error: No target specified${NC}" >&2
        echo "Usage: $0 <owner/repo> or $0 --org <org-name>"
        exit 1
    elif is_org_url "$TARGET"; then
        # User passed an org URL without --org flag, auto-detect
        ORG_MODE=true
        ORG_NAME=$(extract_org_name "$TARGET")
        TARGET=""
    fi
}

#############################################################################
# Main
#############################################################################

main() {
    parse_args "$@"

    # Clean all data if requested
    if [[ "$CLEAN_MODE" == true ]]; then
        clean_all_data
    fi

    if [[ "$ORG_MODE" == true ]]; then
        hydrate_org "$ORG_NAME" "$LIMIT" "${PASS_THROUGH_ARGS[@]}"
    else
        # Single repo mode - forward to bootstrap.sh
        exec "$SCRIPT_DIR/bootstrap.sh" "$TARGET" "${PASS_THROUGH_ARGS[@]}"
    fi
}

main "$@"
