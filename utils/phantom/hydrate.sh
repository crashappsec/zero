#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Phantom Hydrate
# Clone repository/repositories and run scanners for agent queries
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

# Note: We don't use set -e here because we want to continue processing
# remaining repos even if one fails
# set -e

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
    --quick             Fast static analysis (~30s)
    --standard          Most analyzers (~2min) [default]
    --advanced          All static analyzers + health/provenance (~5min)
    --deep              Claude-assisted analysis (~10min)
    --security          Security-focused analysis (~3min)
    --compliance        License and policy compliance (~2min)
    --devops            CI/CD and operational metrics (~3min)
    --force             Re-hydrate even if project exists
    --clean             Remove ALL hydrated data before starting (fresh start)
    -h, --help          Show this help

CONFIGURATION:
    All settings are in utils/phantom/config/phantom.config.json
    See phantom.config.example.json for full documentation
    Create custom profiles by adding entries to the profiles section

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

# Hide cursor (reduces visual noise during animation)
hide_cursor() {
    printf "\033[?25l"
}

# Show cursor
show_cursor() {
    printf "\033[?25h"
}

# Fixed width line output - prevents flashing by overwriting with spaces
# Usage: fixed_line "content" [width]
fixed_line() {
    local content="$1"
    local width="${2:-100}"
    # Strip ANSI codes to get actual visible length
    local visible_content=$(echo -e "$content" | sed 's/\x1b\[[0-9;]*m//g')
    local visible_len=${#visible_content}
    local padding=$((width - visible_len))
    if [[ $padding -lt 0 ]]; then padding=0; fi
    printf "\r%b%${padding}s\n" "$content" ""
}

# Format a scanner line with aligned columns
# Column 1: Status indicator (3 chars)
# Column 2: Scanner name (25 chars)
# Column 3: Result/Status (right-aligned)
format_scanner_line() {
    local indicator="$1"
    local name="$2"
    local result="$3"
    local name_width=26

    # Pad the name to fixed width
    local padded_name=$(printf "%-${name_width}s" "$name")

    printf "  %b %s %b" "$indicator" "$padded_name" "$result"
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

# Get shimmer color code for active text (green theme)
# Returns ANSI color code based on tick for smooth color cycling
get_shimmer_color() {
    local tick="$1"
    local frame=$((tick % 6))
    case $frame in
        0) echo "32" ;;      # green
        1) echo "92" ;;      # bright green
        2) echo "97" ;;      # bright white
        3) echo "92" ;;      # bright green
        4) echo "32" ;;      # green
        5) echo "92" ;;      # bright green
    esac
}

# Format text with shimmer effect (green)
shimmer_text() {
    local text="$1"
    local tick="$2"
    local color=$(get_shimmer_color "$tick")
    printf "\033[${color}m%s\033[0m" "$text"
}

# Get green glow indicator for active tasks
get_green_glow() {
    local tick="$1"
    local frame=$((tick % 4))
    case $frame in
        0) printf "\033[0;32m●\033[0m" ;;   # dim green dot
        1) printf "\033[1;32m◉\033[0m" ;;   # bright green ring
        2) printf "\033[1;97m●\033[0m" ;;   # bright white dot
        3) printf "\033[1;32m◉\033[0m" ;;   # bright green ring
    esac
}

# Print scanner list with status indicators
# Status: queued (dim), active (shimmer), completed (green check), failed (red x)
print_scanner_status() {
    local scanner="$1"
    local status="$2"  # queued, active, completed, failed
    local tick="${3:-0}"
    local result="${4:-}"

    local display=$(get_phase_display "$scanner")
    local estimate=$(get_phase_estimate "$scanner")

    case "$status" in
        queued)
            printf "  ${DIM}○${NC} %-24s ${DIM}(est: %s)${NC}" "$display" "$estimate"
            ;;
        active)
            local glow=$(get_glow $tick)
            printf "  %s %-24s ${DIM}(est: %s)${NC}" "$glow" "$display..." "$estimate"
            ;;
        completed)
            if [[ -n "$result" ]]; then
                printf "  ${GREEN}✓${NC} %-24s %b" "$display" "$result"
            else
                printf "  ${GREEN}✓${NC} %-24s" "$display"
            fi
            ;;
        failed)
            printf "  ${RED}✗${NC} %-24s ${RED}failed${NC}" "$display"
            ;;
    esac
}

# Get output file for a scanner
get_scanner_output_file() {
    local scanner="$1"
    local analysis_path="$2"

    case "$scanner" in
        package-sbom)     echo "$analysis_path/package-sbom.json" ;;
        tech-discovery)   echo "$analysis_path/tech-discovery.json" ;;
        package-vulns)    echo "$analysis_path/package-vulns.json" ;;
        package-health)   echo "$analysis_path/package-health.json" ;;
        licenses)         echo "$analysis_path/licenses.json" ;;
        code-security)    echo "$analysis_path/code-security.json" ;;
        code-secrets)     echo "$analysis_path/code-secrets.json" ;;
        code-ownership)   echo "$analysis_path/code-ownership.json" ;;
        dora)             echo "$analysis_path/dora.json" ;;
        package-provenance) echo "$analysis_path/package-provenance.json" ;;
        git)              echo "$analysis_path/git.json" ;;
        test-coverage)    echo "$analysis_path/test-coverage.json" ;;
        iac-security)     echo "$analysis_path/iac-security.json" ;;
        tech-debt)        echo "$analysis_path/tech-debt.json" ;;
        documentation)    echo "$analysis_path/documentation.json" ;;
        containers)       echo "$analysis_path/containers.json" ;;
        chalk)            echo "$analysis_path/chalk.json" ;;
        digital-certificates) echo "$analysis_path/digital-certificates.json" ;;
        *)                echo "$analysis_path/${scanner}.json" ;;
    esac
}

# Move cursor up N lines
cursor_up() {
    local n="${1:-1}"
    printf "\033[%dA" "$n"
}

# Move cursor down N lines
cursor_down() {
    local n="${1:-1}"
    printf "\033[%dB" "$n"
}

# Save cursor position
cursor_save() {
    printf "\033[s"
}

# Restore cursor position
cursor_restore() {
    printf "\033[u"
}

# Configuration file path (unified config)
CONFIG_FILE="$SCRIPT_DIR/config/phantom.config.json"

# Get list of scanners to run based on profile name
# Loads from phantom.config.json configuration file
get_scanners_for_mode() {
    local mode="$1"

    # Load from phantom.config.json if available
    if [[ -f "$CONFIG_FILE" ]]; then
        local scanners=$(jq -r --arg m "$mode" '.profiles[$m].scanners // empty | join(" ")' "$CONFIG_FILE" 2>/dev/null)
        if [[ -n "$scanners" ]]; then
            echo "$scanners"
            return 0
        fi
    fi

    # Fallback defaults if phantom.config.json not available or profile not found
    case "$mode" in
        quick)
            echo "package-sbom tech-discovery package-vulns licenses tech-debt"
            ;;
        standard|full)
            echo "package-sbom tech-discovery package-vulns licenses code-security code-secrets tech-debt code-ownership dora"
            ;;
        advanced)
            echo "package-sbom tech-discovery package-vulns package-health licenses code-security iac-security code-secrets tech-debt documentation git test-coverage code-ownership dora package-provenance"
            ;;
        deep)
            echo "package-sbom tech-discovery package-vulns package-health licenses code-security iac-security code-secrets tech-debt documentation git test-coverage code-ownership dora package-provenance"
            ;;
        security)
            echo "package-sbom package-vulns licenses code-security iac-security code-secrets"
            ;;
        compliance)
            echo "package-sbom licenses code-security documentation code-ownership"
            ;;
        devops)
            echo "package-sbom tech-discovery iac-security git test-coverage dora package-provenance"
            ;;
        *)
            echo "package-sbom tech-discovery package-vulns licenses code-security code-secrets tech-debt code-ownership dora"
            ;;
    esac
}

# Get profile display info (name, description, estimated_time)
get_profile_info() {
    local profile="$1"
    local field="$2"

    if [[ -f "$CONFIG_FILE" ]]; then
        jq -r --arg p "$profile" --arg f "$field" '.profiles[$p][$f] // empty' "$CONFIG_FILE" 2>/dev/null
    fi
}

# List all available profiles
list_profiles() {
    if [[ -f "$CONFIG_FILE" ]]; then
        jq -r '.profiles | keys[]' "$CONFIG_FILE" 2>/dev/null
    else
        echo "quick standard advanced deep security compliance devops"
    fi
}

# Check if profile requires Claude API
profile_requires_claude() {
    local profile="$1"

    if [[ -f "$CONFIG_FILE" ]]; then
        local requires=$(jq -r --arg p "$profile" '.profiles[$p].requires_claude // false' "$CONFIG_FILE" 2>/dev/null)
        [[ "$requires" == "true" ]]
    else
        [[ "$profile" == "deep" ]]
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

# Get display name for a phase
get_phase_display() {
    local phase="$1"
    case "$phase" in
        "Cloning") echo "Cloning repository" ;;
        "package-sbom") echo "SBOM generation" ;;
        "tech-discovery") echo "Tech discovery" ;;
        "package-vulns") echo "Package vulnerabilities" ;;
        "package-health") echo "Package health" ;;
        "licenses") echo "License scan" ;;
        "code-security") echo "Code security" ;;
        "code-ownership") echo "Code ownership" ;;
        "dora") echo "DORA metrics" ;;
        "package-provenance") echo "Provenance check" ;;
        "git") echo "Git insights" ;;
        "test-coverage") echo "Test coverage" ;;
        "iac-security") echo "IaC security" ;;
        "code-secrets") echo "Secrets scan" ;;
        "tech-debt") echo "Tech debt" ;;
        "documentation") echo "Documentation" ;;
        "containers") echo "Container security" ;;
        "chalk") echo "Chalk artifacts" ;;
        "digital-certificates") echo "Certificates" ;;
        *) echo "$phase" ;;
    esac
}

# Get time estimate for a phase
get_phase_estimate() {
    local phase="$1"
    case "$phase" in
        "Cloning") echo "~30s" ;;
        "package-sbom") echo "~30s" ;;  # SBOM generation with syft
        "tech-discovery") echo "~5s" ;;
        "package-vulns") echo "~10s" ;;
        "package-health") echo "~5s" ;;
        "licenses") echo "~5s" ;;
        "code-security") echo "~15s" ;;
        "code-ownership") echo "~15s" ;;
        "dora") echo "~20s" ;;
        "package-provenance") echo "~10s" ;;
        "git") echo "~10s" ;;
        "test-coverage") echo "~10s" ;;
        "iac-security") echo "~15s" ;;
        "code-secrets") echo "~20s" ;;
        "tech-debt") echo "~30s" ;;
        "documentation") echo "~10s" ;;
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
        "tech-discovery")
            if [[ -f "$analysis_path/tech-discovery.json" ]]; then
                local count=$(jq -r '.technologies | length // 0' "$analysis_path/tech-discovery.json" 2>/dev/null)
                printf "\033[2m%s technologies detected\033[0m" "$count"
            fi
            ;;
        "package-sbom")
            if [[ -f "$analysis_path/package-sbom.json" ]]; then
                local total=$(jq -r '.total_dependencies // 0' "$analysis_path/package-sbom.json" 2>/dev/null)
                local format=$(jq -r '.sbom_format // "unknown"' "$analysis_path/package-sbom.json" 2>/dev/null)
                printf "\033[2m%s packages (%s)\033[0m" "$total" "$format"
            fi
            ;;
        "package-vulns")
            if [[ -f "$analysis_path/package-vulns.json" ]]; then
                local total=$(jq -r '.summary.total // 0' "$analysis_path/package-vulns.json" 2>/dev/null)
                local c=$(jq -r '.summary.critical // 0' "$analysis_path/package-vulns.json" 2>/dev/null)
                local h=$(jq -r '.summary.high // 0' "$analysis_path/package-vulns.json" 2>/dev/null)
                local m=$(jq -r '.summary.medium // 0' "$analysis_path/package-vulns.json" 2>/dev/null)
                local l=$(jq -r '.summary.low // 0' "$analysis_path/package-vulns.json" 2>/dev/null)
                if [[ "$total" == "0" ]]; then
                    printf "\033[0;32mclean\033[0m"
                else
                    # Show breakdown with colors: critical=red, high=orange, medium=yellow, low=dim
                    local parts=()
                    [[ "$c" != "0" ]] && parts+=("\033[1;31m${c}C\033[0m")
                    [[ "$h" != "0" ]] && parts+=("\033[0;31m${h}H\033[0m")
                    [[ "$m" != "0" ]] && parts+=("\033[1;33m${m}M\033[0m")
                    [[ "$l" != "0" ]] && parts+=("\033[2m${l}L\033[0m")
                    local IFS=','
                    printf "%b" "${parts[*]}"
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
        "code-security")
            if [[ -f "$analysis_path/code-security.json" ]]; then
                local high=$(jq -r '.findings | map(select(.severity == "high")) | length // 0' "$analysis_path/code-security.json" 2>/dev/null)
                local medium=$(jq -r '.findings | map(select(.severity == "medium")) | length // 0' "$analysis_path/code-security.json" 2>/dev/null)
                local secrets=$(jq -r '.secrets | length // 0' "$analysis_path/code-security.json" 2>/dev/null)
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
        "code-ownership")
            if [[ -f "$analysis_path/code-ownership.json" ]]; then
                local contributors=$(jq -r '.contributors | length // 0' "$analysis_path/code-ownership.json" 2>/dev/null)
                local bus_factor=$(jq -r '.bus_factor // 0' "$analysis_path/code-ownership.json" 2>/dev/null)
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
        "package-provenance")
            if [[ -f "$analysis_path/package-provenance.json" ]]; then
                local signed=$(jq -r '.summary.signed_commits // 0' "$analysis_path/package-provenance.json" 2>/dev/null)
                local slsa=$(jq -r '.summary.slsa_level // "none"' "$analysis_path/package-provenance.json" 2>/dev/null)
                local status=$(jq -r '.status // "unknown"' "$analysis_path/package-provenance.json" 2>/dev/null)
                if [[ "$status" == "analyzer_not_found" ]]; then
                    printf "\033[2mskipped\033[0m"
                elif [[ "$slsa" != "none" ]] && [[ "$slsa" != "null" ]]; then
                    printf "\033[0;32mSLSA %s\033[0m" "$slsa"
                elif [[ "$signed" -gt 0 ]]; then
                    printf "\033[0;32m%s signed commits\033[0m" "$signed"
                else
                    printf "\033[2mno attestations\033[0m"
                fi
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

# Check if a mode flag is already specified
has_mode_flag() {
    local args=("$@")
    for arg in "${args[@]}"; do
        case "$arg" in
            --quick|--standard|--advanced|--deep|--security|--compliance|--devops)
                return 0
                ;;
        esac
    done
    return 1
}

# Extract mode from args array
get_mode_from_args() {
    local args=("$@")
    for arg in "${args[@]}"; do
        case "$arg" in
            --quick) echo "quick"; return ;;
            --standard) echo "standard"; return ;;
            --advanced) echo "advanced"; return ;;
            --deep) echo "deep"; return ;;
            --security) echo "security"; return ;;
            --compliance) echo "compliance"; return ;;
            --devops) echo "devops"; return ;;
        esac
    done
    echo "standard"  # Default
}

# Prompt user to select analysis mode interactively
select_analysis_mode() {
    echo
    echo -e "${BOLD}Select analysis profile:${NC}"
    echo

    # Build profile menu dynamically from profiles.json
    local profiles=()
    local idx=0

    if [[ -f "$CONFIG_FILE" ]]; then
        # Get profiles in preferred display order
        local ordered_profiles=("quick" "standard" "advanced" "deep" "security" "compliance" "devops")

        for profile in "${ordered_profiles[@]}"; do
            # Check if profile exists in config
            if jq -e --arg p "$profile" '.profiles[$p]' "$CONFIG_FILE" &>/dev/null; then
                ((idx++))
                profiles+=("$profile")

                local name=$(get_profile_info "$profile" "name")
                local time=$(get_profile_info "$profile" "estimated_time")
                local desc=$(get_profile_info "$profile" "description")
                local scanners=$(get_scanners_for_mode "$profile")
                local scanner_count=$(echo "$scanners" | wc -w | tr -d ' ')

                # Format display
                local default_marker=""
                [[ "$profile" == "standard" ]] && default_marker=" ${DIM}(default)${NC}"

                local claude_marker=""
                if profile_requires_claude "$profile"; then
                    claude_marker=" ${DIM}(requires API key)${NC}"
                fi

                printf "  ${CYAN}%d${NC}  %-10s %-8s %s%s%s\n" "$idx" "$name" "$time" "$desc" "$default_marker" "$claude_marker"

                # Show scanner list in dim text
                printf "      ${DIM}→ %s${NC}\n" "$scanners"
            fi
        done

        # Also check for any custom profiles not in the ordered list
        while IFS= read -r profile; do
            if [[ ! " ${ordered_profiles[*]} " =~ " $profile " ]]; then
                ((idx++))
                profiles+=("$profile")

                local name=$(get_profile_info "$profile" "name")
                local time=$(get_profile_info "$profile" "estimated_time")
                local desc=$(get_profile_info "$profile" "description")
                local scanners=$(get_scanners_for_mode "$profile")

                local claude_marker=""
                if profile_requires_claude "$profile"; then
                    claude_marker=" ${DIM}(requires API key)${NC}"
                fi

                printf "  ${CYAN}%d${NC}  %-10s %-8s %s%s\n" "$idx" "$name" "$time" "$desc" "$claude_marker"
                printf "      ${DIM}→ %s${NC}\n" "$scanners"
            fi
        done < <(jq -r '.profiles | keys[]' "$CONFIG_FILE" 2>/dev/null)
    else
        # Fallback if no phantom.config.json
        echo -e "  ${CYAN}1${NC}  Quick      ~30s   Fast static analysis"
        echo -e "  ${CYAN}2${NC}  Standard   ~2min  Most scanners ${DIM}(default)${NC}"
        echo -e "  ${CYAN}3${NC}  Advanced   ~5min  All static scanners + package health"
        echo -e "  ${CYAN}4${NC}  Deep       ~10min Claude-assisted analysis ${DIM}(requires API key)${NC}"
        echo -e "  ${CYAN}5${NC}  Security   ~3min  Security-focused analysis"
        profiles=("quick" "standard" "advanced" "deep" "security")
    fi

    echo
    read -p "Choose profile [2]: " -r mode_choice
    echo
    echo

    # Handle selection
    if [[ -z "$mode_choice" ]]; then
        echo "--standard"
    elif [[ "$mode_choice" =~ ^[0-9]+$ ]] && [[ $mode_choice -ge 1 ]] && [[ $mode_choice -le ${#profiles[@]} ]]; then
        echo "--${profiles[$((mode_choice-1))]}"
    else
        echo "--standard"
    fi
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

    # If no mode specified and running interactively, prompt for mode
    if ! has_mode_flag "${extra_args[@]}" && [[ -t 0 ]]; then
        local selected_mode=$(select_analysis_mode)
        extra_args+=("$selected_mode")
    fi

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

    # Determine current mode for progress display and summary
    local current_mode=$(get_mode_from_args "${extra_args[@]}")
    local mode_scanners=$(get_scanners_for_mode "$current_mode")
    local scanner_count=$(echo "$mode_scanners" | wc -w | tr -d ' ')

    # Get mode display name from phantom.config.json or use fallback
    local mode_display=""
    if [[ -f "$CONFIG_FILE" ]]; then
        mode_display=$(get_profile_info "$current_mode" "name")
    fi
    if [[ -z "$mode_display" ]]; then
        case "$current_mode" in
            quick)      mode_display="Quick" ;;
            standard)   mode_display="Standard" ;;
            advanced)   mode_display="Advanced" ;;
            deep)       mode_display="Deep" ;;
            security)   mode_display="Security" ;;
            compliance) mode_display="Compliance" ;;
            devops)     mode_display="DevOps" ;;
            *)          mode_display="Standard" ;;
        esac
    fi

    # Show org summary
    echo
    echo -e "${BOLD}Organization Summary${NC}"
    echo -e "  Repositories:  ${CYAN}$repo_count${NC}"
    echo -e "  Total Size:    ${CYAN}$total_size${NC}"
    echo -e "  Languages:     ${CYAN}$languages${NC}"
    echo -e "  Scan Profile:  ${CYAN}$mode_display${NC} ${DIM}($scanner_count scanners)${NC}"
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
                local total=$(jq -r '.summary.total // 0' "$analysis_path/vulnerabilities.json" 2>/dev/null)
                local c=$(jq -r '.summary.critical // 0' "$analysis_path/vulnerabilities.json" 2>/dev/null)
                local h=$(jq -r '.summary.high // 0' "$analysis_path/vulnerabilities.json" 2>/dev/null)
                if [[ "$total" == "0" ]]; then
                    results+="  ${GREEN}✓${NC} Package vulnerabilities: ${GREEN}clean${NC}\n"
                elif [[ "$c" != "0" ]] || [[ "$h" != "0" ]]; then
                    results+="  ${YELLOW}!${NC} Package vulnerabilities: ${RED}$total found${NC} ($c critical, $h high)\n"
                else
                    results+="  ${YELLOW}!${NC} Package vulnerabilities: ${YELLOW}$total found${NC} (low/medium)\n"
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

            # Provenance
            if [[ -f "$analysis_path/provenance.json" ]]; then
                local prov_status=$(jq -r '.status // "unknown"' "$analysis_path/provenance.json" 2>/dev/null)
                if [[ "$prov_status" != "analyzer_not_found" ]]; then
                    local slsa=$(jq -r '.summary.slsa_level // "none"' "$analysis_path/provenance.json" 2>/dev/null)
                    if [[ "$slsa" != "none" ]] && [[ "$slsa" != "null" ]]; then
                        results+="  ${GREEN}✓${NC} Provenance: ${GREEN}SLSA $slsa${NC}\n"
                    else
                        results+="  ${GREEN}✓${NC} Provenance: no attestations\n"
                    fi
                fi
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

        local analysis_path="$GIBSON_PROJECTS_DIR/$project_id/analysis"
        local tick=0

        # Convert scanner list to array for easier manipulation
        local -a scanners_array=($mode_scanners)
        local total_scanners=${#scanners_array[@]}

        # Get all available scanners for display
        local all_scanners="package-sbom tech-discovery package-vulns package-health licenses code-security code-ownership dora package-provenance git test-coverage iac-security code-secrets tech-debt documentation containers chalk digital-certificates"

        # Build array of scanners NOT in current profile
        local -a skipped_scanners=()
        for scanner in $all_scanners; do
            local in_profile=false
            for active in "${scanners_array[@]}"; do
                if [[ "$scanner" == "$active" ]]; then
                    in_profile=true
                    break
                fi
            done
            if [[ "$in_profile" == "false" ]]; then
                skipped_scanners+=("$scanner")
            fi
        done

        # Print initial scanner list (cloning + all profile scanners)
        # Queued profile scanners are MAGENTA, skipped are DIM gray
        echo -e "$(format_scanner_line "${MAGENTA}○${NC}" "${MAGENTA}Cloning repository...${NC}" "${MAGENTA}est: ~30s${NC}")"
        for scanner in "${scanners_array[@]}"; do
            local display=$(get_phase_display "$scanner")
            local estimate=$(get_phase_estimate "$scanner")
            # Check if we have cached data from previous scan
            local output_file=$(get_scanner_output_file "$scanner" "$analysis_path")
            if [[ -f "$output_file" ]] && [[ "$force_mode" != "true" ]]; then
                local result=$(get_phase_result "$scanner" "$analysis_path" "$project_id")
                echo -e "$(format_scanner_line "${GREEN}✓${NC}" "${GREEN}${display}${NC}" "$result ${DIM}(cached)${NC}")"
            else
                echo -e "$(format_scanner_line "${MAGENTA}○${NC}" "${MAGENTA}${display}${NC}" "${MAGENTA}est: ${estimate}${NC}")"
            fi
        done

        # Print separator and skipped scanners if there are any
        if [[ ${#skipped_scanners[@]} -gt 0 ]]; then
            echo -e "  ${DIM}─────────────────────────────────────${NC}"
            for scanner in "${skipped_scanners[@]}"; do
                local display=$(get_phase_display "$scanner")
                # Check if we have cached data from previous scan
                local output_file=$(get_scanner_output_file "$scanner" "$analysis_path")
                if [[ -f "$output_file" ]]; then
                    local result=$(get_phase_result "$scanner" "$analysis_path" "$project_id")
                    echo -e "$(format_scanner_line "${GREEN}✓${NC}" "${GREEN}${display}${NC}" "$result ${DIM}(cached)${NC}")"
                else
                    echo -e "$(format_scanner_line "${DIM}○${NC}" "${DIM}${display}${NC}" "${DIM}skipped in profile${NC}")"
                fi
            done
        fi

        # Total lines to manage: 1 (cloning) + total_scanners + separator + skipped_scanners
        local total_lines=$((total_scanners + 1))
        if [[ ${#skipped_scanners[@]} -gt 0 ]]; then
            total_lines=$((total_lines + 1 + ${#skipped_scanners[@]}))  # +1 for separator line
        fi

        # Hide cursor during animation to reduce visual noise
        hide_cursor

        # Cleanup function for proper exit
        cleanup_animation() {
            show_cursor
            # Kill the background process if running
            if kill -0 $pid 2>/dev/null; then
                kill $pid 2>/dev/null
                wait $pid 2>/dev/null
            fi
            rm -f "$log_file"
            exit 130  # Standard exit code for Ctrl+C
        }

        # Trap to ensure cursor is shown and process killed on exit (ctrl-c, etc)
        trap 'cleanup_animation' INT TERM

        # Monitor progress
        while kill -0 $pid 2>/dev/null; do
            local now=$(date +%s)
            local elapsed=$((now - start_time))
            local clean_log=$(sed 's/\x1b\[[0-9;]*m//g' "$log_file" 2>/dev/null)

            # Check if cloning is done (repo exists and Languages printed)
            local cloning_done=false
            if [[ -d "$GIBSON_PROJECTS_DIR/$project_id/repo" ]] && echo "$clean_log" | grep -q "Languages:"; then
                cloning_done=true
            fi

            # Move cursor to start of scanner list
            cursor_up $total_lines

            # Update cloning status - use fixed_line to prevent flashing
            if [[ "$cloning_done" == "true" ]]; then
                local repo_path="$GIBSON_PROJECTS_DIR/$project_id/repo"
                local size=$(du -sh "$repo_path" 2>/dev/null | cut -f1)
                local files=$(find "$repo_path" -type f ! -path '*/.git/*' 2>/dev/null | wc -l | tr -d ' ')
                fixed_line "$(format_scanner_line "${GREEN}✓${NC}" "Cloning repository" "${DIM}${size}, ${files} files${NC}")"
            else
                local glow=$(get_green_glow $tick)
                local shimmer_name=$(shimmer_text "Cloning repository..." $tick)
                fixed_line "$(format_scanner_line "$glow" "$shimmer_name" "${DIM}${elapsed}s (est: ~30s)${NC}")"
            fi

            # Update each scanner's status
            local active_found=false
            for scanner in "${scanners_array[@]}"; do
                local display=$(get_phase_display "$scanner")
                local estimate=$(get_phase_estimate "$scanner")
                local output_file=$(get_scanner_output_file "$scanner" "$analysis_path")

                if [[ -f "$output_file" ]]; then
                    # Scanner completed - show result in green
                    local result=$(get_phase_result "$scanner" "$analysis_path" "$project_id")
                    fixed_line "$(format_scanner_line "${GREEN}✓${NC}" "${GREEN}${display}${NC}" "$result")"
                elif [[ "$cloning_done" == "true" ]] && [[ "$active_found" == "false" ]]; then
                    # First incomplete scanner after cloning - mark as active with green shimmer
                    active_found=true
                    local glow=$(get_green_glow $tick)
                    local shimmer_name=$(shimmer_text "${display}..." $tick)
                    fixed_line "$(format_scanner_line "$glow" "$shimmer_name" "${DIM}${elapsed}s (est: ${estimate})${NC}")"
                else
                    # Still queued - purple/magenta for profile scanners
                    fixed_line "$(format_scanner_line "${MAGENTA}○${NC}" "${MAGENTA}${display}${NC}" "${MAGENTA}est: ${estimate}${NC}")"
                fi
            done

            # Update separator and skipped scanners if present
            if [[ ${#skipped_scanners[@]} -gt 0 ]]; then
                fixed_line "  ${DIM}─────────────────────────────────────${NC}"
                for scanner in "${skipped_scanners[@]}"; do
                    local display=$(get_phase_display "$scanner")
                    local output_file=$(get_scanner_output_file "$scanner" "$analysis_path")
                    if [[ -f "$output_file" ]]; then
                        local result=$(get_phase_result "$scanner" "$analysis_path" "$project_id")
                        fixed_line "$(format_scanner_line "${GREEN}✓${NC}" "${GREEN}${display}${NC}" "$result ${DIM}(cached)${NC}")"
                    else
                        # Skipped scanners stay dim gray
                        fixed_line "$(format_scanner_line "${DIM}○${NC}" "${DIM}${display}${NC}" "${DIM}skipped in profile${NC}")"
                    fi
                done
            fi

            ((tick++))
            sleep 0.3
        done

        # Show cursor again after animation and clear trap
        show_cursor
        trap - INT TERM

        # Get exit status
        wait $pid
        local exit_status=$?
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))

        # Final update - move cursor back and update all lines with final status
        cursor_up $total_lines

        # Show cloning as complete
        local repo_path="$GIBSON_PROJECTS_DIR/$project_id/repo"
        if [[ -d "$repo_path" ]]; then
            local size=$(du -sh "$repo_path" 2>/dev/null | cut -f1)
            local files=$(find "$repo_path" -type f ! -path '*/.git/*' 2>/dev/null | wc -l | tr -d ' ')
            fixed_line "$(format_scanner_line "${GREEN}✓${NC}" "Cloning repository" "${DIM}${size}, ${files} files${NC}")"
        else
            fixed_line "$(format_scanner_line "${RED}✗${NC}" "${RED}Cloning repository${NC}" "${RED}failed${NC}")"
        fi

        # Show all scanners with final status
        for scanner in "${scanners_array[@]}"; do
            local display=$(get_phase_display "$scanner")
            local output_file=$(get_scanner_output_file "$scanner" "$analysis_path")

            if [[ -f "$output_file" ]]; then
                local result=$(get_phase_result "$scanner" "$analysis_path" "$project_id")
                fixed_line "$(format_scanner_line "${GREEN}✓${NC}" "${GREEN}${display}${NC}" "$result")"
            else
                fixed_line "$(format_scanner_line "${RED}✗${NC}" "${RED}${display}${NC}" "${RED}failed${NC}")"
            fi
        done

        # Show separator and skipped scanners with final status
        if [[ ${#skipped_scanners[@]} -gt 0 ]]; then
            fixed_line "  ${DIM}─────────────────────────────────────${NC}"
            for scanner in "${skipped_scanners[@]}"; do
                local display=$(get_phase_display "$scanner")
                local output_file=$(get_scanner_output_file "$scanner" "$analysis_path")
                if [[ -f "$output_file" ]]; then
                    local result=$(get_phase_result "$scanner" "$analysis_path" "$project_id")
                    fixed_line "$(format_scanner_line "${GREEN}✓${NC}" "${GREEN}${display}${NC}" "$result ${DIM}(cached)${NC}")"
                else
                    fixed_line "$(format_scanner_line "${DIM}○${NC}" "${DIM}${display}${NC}" "${DIM}skipped in profile${NC}")"
                fi
            done
        fi

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
            --branch|--depth|--quick|--standard|--advanced|--deep|--security|--compliance|--devops|--force)
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
