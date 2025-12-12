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

    # Clone or update
    local project_id=$(zero_project_id "$current_repo")
    local repo_path="$ZERO_PROJECTS_DIR/$project_id/repo"

    if [[ -d "$repo_path" ]] && [[ "$FORCE" != "true" ]]; then
        # Update existing repo
        echo -e "  ${CYAN}Updating...${NC}"
        local old_commit=$(cd "$repo_path" && git rev-parse --short HEAD 2>/dev/null)
        if (cd "$repo_path" && git pull -q 2>/dev/null); then
            local new_commit=$(cd "$repo_path" && git rev-parse --short HEAD 2>/dev/null)
            if [[ "$old_commit" != "$new_commit" ]]; then
                echo -e "  ${GREEN}✓${NC} Updated (${old_commit} → ${new_commit})"
            else
                echo -e "  ${GREEN}✓${NC} Already up to date (${new_commit})"
            fi
        else
            echo -e "  ${GREEN}✓${NC} Using cached repo (pull skipped)"
        fi
        CLONE_STATUS="updated"
    else
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

    echo -e "Fetching repositories for ${CYAN}$org${NC}..."
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
    local scan_id="scan-$(date +%Y%m%d-%H%M%S)"

    # Print header
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${BOLD}Hydrate Organization: ${CYAN}$org${NC}"
    echo -e "Scan ID:      ${DIM}$scan_id${NC}"
    echo -e "Repositories: ${CYAN}$repo_count${NC}"
    echo -e "Profile:      ${CYAN}$PROFILE${NC}"
    echo -e "Parallel:     ${CYAN}$parallel_jobs jobs${NC}"
    echo ""

    # Create temp dir for tracking
    local tmp_dir=$(mktemp -d)

    # Setup cleanup handler
    cleanup_hydrate() {
        rm -rf "$tmp_dir" 2>/dev/null || true
    }
    trap cleanup_hydrate EXIT

    local start_time=$(date +%s)
    local clone_success=0
    local clone_failed=0
    local scan_success=0
    local scan_failed=0
    local total_disk_size=0
    local total_file_count=0

    #---------------------------------------------------------------------------
    # PHASE 1: Cloning
    #---------------------------------------------------------------------------
    echo -e "${BOLD}Cloning${NC}"
    echo ""

    for current_repo in "${repos[@]}"; do
        local short="${current_repo##*/}"
        local project_id=$(zero_project_id "$current_repo")
        local repo_path="$ZERO_PROJECTS_DIR/$project_id/repo"

        if [[ -d "$repo_path" ]] && [[ "$FORCE" != "true" ]]; then
            # Update existing repo with git pull
            printf "  ${YELLOW}*${NC} %s ${DIM}updating...${NC}" "$short"
            local old_commit=$(cd "$repo_path" && git rev-parse --short HEAD 2>/dev/null)
            if (cd "$repo_path" && git pull -q 2>/dev/null); then
                local new_commit=$(cd "$repo_path" && git rev-parse --short HEAD 2>/dev/null)
                local stats=$(format_repo_stats "$repo_path")
                if [[ "$old_commit" != "$new_commit" ]]; then
                    printf "\r  ${GREEN}✓${NC} %s ${DIM}%s${NC} ${CYAN}(updated: %s → %s)${NC}                    \n" "$short" "$stats" "$old_commit" "$new_commit"
                else
                    printf "\r  ${GREEN}✓${NC} %s ${DIM}%s${NC} ${DIM}(up to date)${NC}                    \n" "$short" "$stats"
                fi
            else
                local stats=$(format_repo_stats "$repo_path")
                printf "\r  ${GREEN}✓${NC} %s ${DIM}%s${NC} ${DIM}(pull skipped)${NC}                    \n" "$short" "$stats"
            fi
            clone_success=$((clone_success + 1))
        else
            printf "  ${YELLOW}*${NC} %s ${DIM}cloning...${NC}" "$short"

            [[ -d "$repo_path" ]] && rm -rf "$repo_path"
            mkdir -p "$ZERO_PROJECTS_DIR/$project_id"

            local clone_url="https://github.com/$current_repo.git"
            [[ -n "${GITHUB_TOKEN:-}" ]] && clone_url="https://${GITHUB_TOKEN}@github.com/$current_repo.git"

            local clone_args=("-q")
            [[ -n "$BRANCH" ]] && clone_args+=("--branch" "$BRANCH")
            [[ -n "$DEPTH" ]] && clone_args+=("--depth" "$DEPTH")

            if git clone "${clone_args[@]}" "$clone_url" "$repo_path" 2>/dev/null; then
                mkdir -p "$ZERO_PROJECTS_DIR/$project_id/analysis"
                local stats=$(format_repo_stats "$repo_path")
                printf "\r  ${GREEN}✓${NC} %s ${DIM}%s${NC}                    \n" "$short" "$stats"
                clone_success=$((clone_success + 1))
            else
                printf "\r  ${RED}✗${NC} %s ${RED}clone failed${NC}           \n" "$short"
                echo "$current_repo" >> "$tmp_dir/failed_repos"
                clone_failed=$((clone_failed + 1))
            fi
        fi
    done

    echo ""
    echo -e "${GREEN}✓${NC} ${BOLD}Cloning complete${NC} ${DIM}($clone_success repos now available locally)${NC}"
    echo ""

    # Skip scanning if clone-only
    if [[ "$CLONE_ONLY" == "true" ]]; then
        local total_elapsed=$(($(date +%s) - start_time))
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo -e "${GREEN}${BOLD}✓ Hydrate complete${NC} (clone only, $(format_duration $total_elapsed))"
        return 0
    fi

    #---------------------------------------------------------------------------
    # PHASE 2: Scanning with real-time per-scanner progress
    #---------------------------------------------------------------------------
    echo -e "${BOLD}Scanning Organization:${NC} ${CYAN}$org${NC} ${DIM}with profile ${PROFILE}${NC}"
    echo ""

    # Build list of repos to scan
    local scan_repos=()
    for current_repo in "${repos[@]}"; do
        if [[ -f "$tmp_dir/failed_repos" ]] && grep -q "^$current_repo$" "$tmp_dir/failed_repos" 2>/dev/null; then
            continue
        fi
        scan_repos+=("$current_repo")
    done

    local total_to_scan=${#scan_repos[@]}
    local scanned=0
    local pids=()
    local pid_to_repo=()

    # Helper to extract scanner summary from JSON
    get_scanner_summary() {
        local json_file="$1"
        local scanner_name=$(basename "$json_file" .json)
        local summary=""

        case "$scanner_name" in
            package-vulns)
                local total=$(jq -r '.summary.total // 0' "$json_file" 2>/dev/null)
                local critical=$(jq -r '.summary.critical // 0' "$json_file" 2>/dev/null)
                local high=$(jq -r '.summary.high // 0' "$json_file" 2>/dev/null)
                if [[ "$total" -gt 0 ]]; then
                    summary="${RED}$critical critical${NC}, ${YELLOW}$high high${NC}"
                fi
                ;;
            code-security)
                local count=$(jq '.findings | length' "$json_file" 2>/dev/null || echo "0")
                [[ "$count" -gt 0 ]] && summary="${YELLOW}$count issues${NC}"
                ;;
            code-secrets)
                local count=$(jq '.findings | length' "$json_file" 2>/dev/null || echo "0")
                [[ "$count" -gt 0 ]] && summary="${RED}$count secrets${NC}"
                ;;
            iac-security)
                local count=$(jq '.findings | length' "$json_file" 2>/dev/null || echo "0")
                [[ "$count" -gt 0 ]] && summary="${YELLOW}$count issues${NC}"
                ;;
            licenses)
                local violations=$(jq -r '.summary.license_violations // 0' "$json_file" 2>/dev/null)
                local deps=$(jq -r '.summary.total_dependencies_with_licenses // 0' "$json_file" 2>/dev/null)
                [[ "$violations" -gt 0 ]] && summary="${RED}$violations violations${NC}" || summary="${DIM}$deps deps${NC}"
                ;;
            package-malcontent)
                local critical=$(jq '[.packages[]?.findings[]? | select(.severity == "critical")] | length' "$json_file" 2>/dev/null || echo "0")
                local high=$(jq '[.packages[]?.findings[]? | select(.severity == "high")] | length' "$json_file" 2>/dev/null || echo "0")
                [[ "$critical" -gt 0 || "$high" -gt 0 ]] && summary="${RED}$critical critical${NC}, ${YELLOW}$high high${NC}"
                ;;
            container-security)
                local count=$(jq '.vulnerabilities | length' "$json_file" 2>/dev/null || echo "0")
                [[ "$count" -gt 0 ]] && summary="${YELLOW}$count vulns${NC}"
                ;;
            package-sbom)
                local total=$(jq -r '.summary.total // .total_dependencies // 0' "$json_file" 2>/dev/null)
                # Get ecosystem breakdown from the sbom.cdx.json if available, or from summary
                local ecosystems=""
                local analysis_dir=$(dirname "$json_file")
                local sbom_file="$analysis_dir/sbom.cdx.json"
                if [[ -f "$sbom_file" ]]; then
                    ecosystems=$(jq -r '[.components[]? | .purl // empty | ltrimstr("pkg:") | split("/")[0]] | map(select(length > 0)) | group_by(.) | map({type: .[0], count: length}) | sort_by(-.count) | .[0:3] | map("\(.type):\(.count)") | join(" ")' "$sbom_file" 2>/dev/null)
                else
                    ecosystems=$(jq -r '[.summary.ecosystems // {} | to_entries[] | "\(.key):\(.value)"] | join(" ")' "$json_file" 2>/dev/null)
                fi
                if [[ "$total" -gt 0 ]]; then
                    if [[ -n "$ecosystems" ]] && [[ "$ecosystems" != "null" ]] && [[ "$ecosystems" != "" ]]; then
                        summary="${DIM}$total packages ($ecosystems)${NC}"
                    else
                        summary="${DIM}$total packages${NC}"
                    fi
                fi
                ;;
        esac

        echo "$summary"
    }

    # Track which scanners we've already printed for each repo
    # Format: "repo:scanner repo:scanner ..."
    local printed_scanners=""

    # Print scanner result as it completes (called during polling)
    print_scanner_result() {
        local repo="$1"
        local json_file="$2"
        local scanner_name=$(basename "$json_file" .json)

        # Skip manifest and sbom.cdx
        [[ "$scanner_name" == "manifest" || "$scanner_name" == "sbom.cdx" ]] && return

        # Check if already printed
        [[ "$printed_scanners" =~ "$repo:$scanner_name " ]] && return

        # Mark as printed
        printed_scanners+="$repo:$scanner_name "

        local summary=$(get_scanner_summary "$json_file")
        if [[ -n "$summary" ]]; then
            echo -e "      ${DIM}└─${NC} ${DIM}${repo}/${NC}$scanner_name ${summary}"
        else
            echo -e "      ${DIM}└─${NC} ${DIM}${repo}/${NC}$scanner_name ${GREEN}no findings${NC}"
        fi
    }

    # Check for new scanner completions across all active repos
    print_scanner_progress() {
        for i in "${!pids[@]}"; do
            local repo="${pid_to_repo[$i]}"
            local short="${repo##*/}"
            local project_id=$(zero_project_id "$repo")
            local analysis_dir="$ZERO_PROJECTS_DIR/$project_id/analysis"

            # Check for any new JSON files
            if [[ -d "$analysis_dir" ]]; then
                for json_file in "$analysis_dir"/*.json; do
                    [[ -f "$json_file" ]] || continue
                    print_scanner_result "$short" "$json_file"
                done
            fi
        done
    }

    # Function to check and print completed jobs
    print_completed() {
        local new_pids=()
        local new_pid_to_repo=()

        for i in "${!pids[@]}"; do
            local pid="${pids[$i]}"
            local repo="${pid_to_repo[$i]}"

            if ! kill -0 "$pid" 2>/dev/null; then
                wait "$pid" 2>/dev/null || true
                local short="${repo##*/}"
                local project_id=$(zero_project_id "$repo")
                local exit_code=$(cat "$tmp_dir/${short}.exit" 2>/dev/null || echo "1")
                local analysis_dir="$ZERO_PROJECTS_DIR/$project_id/analysis"

                # Print any remaining scanners that weren't caught in progress
                if [[ -d "$analysis_dir" ]]; then
                    for json_file in "$analysis_dir"/*.json; do
                        [[ -f "$json_file" ]] || continue
                        print_scanner_result "$short" "$json_file"
                    done
                fi

                if [[ "$exit_code" == "0" ]]; then
                    echo -e "  ${GREEN}✓${NC} ${BOLD}${short}${NC} ${DIM}complete${NC}"
                    scan_success=$((scan_success + 1))
                else
                    echo -e "  ${RED}✗${NC} ${short} ${RED}(scan failed)${NC}"
                    scan_failed=$((scan_failed + 1))
                fi
                scanned=$((scanned + 1))
            else
                new_pids+=("$pid")
                new_pid_to_repo+=("$repo")
            fi
        done

        pids=("${new_pids[@]}")
        pid_to_repo=("${new_pid_to_repo[@]}")
    }

    # Launch and manage parallel jobs
    for current_repo in "${scan_repos[@]}"; do
        local short="${current_repo##*/}"
        local project_id=$(zero_project_id "$current_repo")
        local analysis_dir="$ZERO_PROJECTS_DIR/$project_id/analysis"

        # Clear old analysis JSON files to ensure fresh results (preserve scans history)
        if [[ -d "$analysis_dir" ]]; then
            rm -f "$analysis_dir"/*.json "$analysis_dir"/sbom.cdx.json 2>/dev/null || true
        fi

        # Show immediate feedback that scan is starting
        echo -e "  ${YELLOW}◐${NC} ${BOLD}${short}${NC} ${DIM}scanning...${NC}"

        # Start scan in background
        (
            local force_arg=""
            [[ "$FORCE" == "true" ]] && force_arg="--force"
            "$SCRIPT_DIR/bootstrap.sh" --scan-only --"$PROFILE" $force_arg "$current_repo" > "$tmp_dir/${short}.out" 2>&1
            echo $? > "$tmp_dir/${short}.exit"
        ) &

        pids+=($!)
        pid_to_repo+=("$current_repo")

        # If at capacity, wait for one to finish while showing scanner progress
        while [[ ${#pids[@]} -ge $parallel_jobs ]]; do
            sleep 0.5
            print_scanner_progress
            print_completed
        done
    done

    # Wait for remaining jobs while showing scanner progress
    while [[ ${#pids[@]} -gt 0 ]]; do
        sleep 0.5
        print_scanner_progress
        print_completed
    done

    echo ""
    echo -e "${GREEN}✓${NC} ${BOLD}Scanning complete${NC}"
    echo ""

    #---------------------------------------------------------------------------
    # SUMMARY REPORT
    #---------------------------------------------------------------------------
    local total_elapsed=$(($(date +%s) - start_time))

    # Calculate totals across all repos
    local total_vulns_critical=0
    local total_vulns_high=0
    local total_vulns_medium=0
    local total_vulns_low=0
    local total_secrets=0
    local total_code_issues=0
    local total_iac_issues=0
    local total_license_violations=0
    local total_packages=0
    local total_malcontent_critical=0
    local total_malcontent_high=0
    local scanners_used=""

    for current_repo in "${scan_repos[@]}"; do
        local project_id=$(zero_project_id "$current_repo")
        local analysis_dir="$ZERO_PROJECTS_DIR/$project_id/analysis"

        [[ -d "$analysis_dir" ]] || continue

        # Package vulnerabilities
        if [[ -f "$analysis_dir/package-vulns.json" ]]; then
            total_vulns_critical=$((total_vulns_critical + $(jq -r '.summary.critical // 0' "$analysis_dir/package-vulns.json" 2>/dev/null || echo 0)))
            total_vulns_high=$((total_vulns_high + $(jq -r '.summary.high // 0' "$analysis_dir/package-vulns.json" 2>/dev/null || echo 0)))
            total_vulns_medium=$((total_vulns_medium + $(jq -r '.summary.medium // 0' "$analysis_dir/package-vulns.json" 2>/dev/null || echo 0)))
            total_vulns_low=$((total_vulns_low + $(jq -r '.summary.low // 0' "$analysis_dir/package-vulns.json" 2>/dev/null || echo 0)))
            [[ ! "$scanners_used" =~ "package-vulns" ]] && scanners_used+=" package-vulns"
        fi

        # Secrets
        if [[ -f "$analysis_dir/code-secrets.json" ]]; then
            total_secrets=$((total_secrets + $(jq '.findings | length' "$analysis_dir/code-secrets.json" 2>/dev/null || echo 0)))
            [[ ! "$scanners_used" =~ "code-secrets" ]] && scanners_used+=" code-secrets"
        fi

        # Code security
        if [[ -f "$analysis_dir/code-security.json" ]]; then
            total_code_issues=$((total_code_issues + $(jq '.findings | length' "$analysis_dir/code-security.json" 2>/dev/null || echo 0)))
            [[ ! "$scanners_used" =~ "code-security" ]] && scanners_used+=" code-security"
        fi

        # IaC security
        if [[ -f "$analysis_dir/iac-security.json" ]]; then
            total_iac_issues=$((total_iac_issues + $(jq '.findings | length' "$analysis_dir/iac-security.json" 2>/dev/null || echo 0)))
            [[ ! "$scanners_used" =~ "iac-security" ]] && scanners_used+=" iac-security"
        fi

        # Licenses
        if [[ -f "$analysis_dir/licenses.json" ]]; then
            total_license_violations=$((total_license_violations + $(jq -r '.summary.license_violations // 0' "$analysis_dir/licenses.json" 2>/dev/null || echo 0)))
            [[ ! "$scanners_used" =~ "licenses" ]] && scanners_used+=" licenses"
        fi

        # Package count from SBOM (stored in manifest.json)
        if [[ -f "$analysis_dir/manifest.json" ]]; then
            local pkg_count=$(jq -r '.analyses["package-sbom"].summary.total // 0' "$analysis_dir/manifest.json" 2>/dev/null || echo 0)
            total_packages=$((total_packages + pkg_count))
            [[ ! "$scanners_used" =~ "package-sbom" ]] && scanners_used+=" package-sbom"
        fi

        # Malcontent
        if [[ -f "$analysis_dir/package-malcontent.json" ]]; then
            total_malcontent_critical=$((total_malcontent_critical + $(jq '[.packages[]?.findings[]? | select(.severity == "critical")] | length' "$analysis_dir/package-malcontent.json" 2>/dev/null || echo 0)))
            total_malcontent_high=$((total_malcontent_high + $(jq '[.packages[]?.findings[]? | select(.severity == "high")] | length' "$analysis_dir/package-malcontent.json" 2>/dev/null || echo 0)))
            [[ ! "$scanners_used" =~ "package-malcontent" ]] && scanners_used+=" package-malcontent"
        fi

        # Container security
        if [[ -f "$analysis_dir/container-security.json" ]]; then
            [[ ! "$scanners_used" =~ "container-security" ]] && scanners_used+=" container-security"
        fi
    done

    # Calculate disk usage
    local disk_usage=$(du -sh "$ZERO_PROJECTS_DIR" 2>/dev/null | awk '{print $1}')
    local file_count=$(find "$ZERO_PROJECTS_DIR" -type f 2>/dev/null | wc -l | tr -d ' ')
    local scanner_count=$(echo $scanners_used | wc -w | tr -d ' ')

    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${GREEN}${BOLD}✓ Hydrate Complete${NC}"
    echo ""
    echo -e "${BOLD}Summary${NC}"
    echo -e "  Organization:    ${CYAN}$org${NC}"
    echo -e "  Duration:        $(format_duration $total_elapsed)"
    echo -e "  Repos scanned:   ${GREEN}$scan_success success${NC}$([[ $scan_failed -gt 0 ]] && echo -e ", ${RED}$scan_failed failed${NC}")"
    echo -e "  Disk usage:      ${disk_usage}"
    echo -e "  Total files:     $(format_number $file_count)"
    echo -e "  Scanners ran:    $scanner_count (${DIM}${scanners_used# }${NC})"
    echo ""

    # Only show findings section if there are findings
    local has_findings=false
    [[ $total_vulns_critical -gt 0 || $total_vulns_high -gt 0 || $total_secrets -gt 0 || $total_code_issues -gt 0 || $total_malcontent_critical -gt 0 || $total_malcontent_high -gt 0 ]] && has_findings=true

    local total_vulns=$((total_vulns_critical + total_vulns_high + total_vulns_medium + total_vulns_low))
    echo -e "${BOLD}Findings Summary${NC}"
    echo -e "  Package vulnerabilities: $(format_number $total_vulns) total"
    [[ $total_vulns_critical -gt 0 ]] && echo -e "    ${RED}● $(format_number $total_vulns_critical) critical${NC}"
    [[ $total_vulns_high -gt 0 ]] && echo -e "    ${YELLOW}● $(format_number $total_vulns_high) high${NC}"
    [[ $total_vulns_medium -gt 0 ]] && echo -e "    ${DIM}● $(format_number $total_vulns_medium) medium${NC}"
    [[ $total_vulns_low -gt 0 ]] && echo -e "    ${DIM}● $(format_number $total_vulns_low) low${NC}"

    echo -e "  Secrets detected:        $(format_number $total_secrets)$([[ $total_secrets -gt 0 ]] && echo -e " ${RED}⚠${NC}")"
    echo -e "  Code security issues:    $(format_number $total_code_issues)"
    echo -e "  IaC security issues:     $(format_number $total_iac_issues)"
    echo -e "  License violations:      $(format_number $total_license_violations)"
    echo -e "  Packages analyzed:       $(format_number $total_packages)"

    if [[ $total_malcontent_critical -gt 0 || $total_malcontent_high -gt 0 ]]; then
        echo -e "  Supply chain risks:"
        [[ $total_malcontent_critical -gt 0 ]] && echo -e "    ${RED}● $(format_number $total_malcontent_critical) critical${NC}"
        [[ $total_malcontent_high -gt 0 ]] && echo -e "    ${YELLOW}● $(format_number $total_malcontent_high) high${NC}"
    fi

    echo ""
    echo -e "${DIM}Full details written to: ${NC}${CYAN}.zero/repos/${NC}"
    echo -e "${DIM}View detailed report:    ${NC}${CYAN}./zero.sh report --org $org${NC}"
    echo ""
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
