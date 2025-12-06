#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Scanner UX Library
# Unified terminal user experience for all scanners
#
# Usage:
#   source "$UTILS_ROOT/lib/scanner-ux.sh"
#   scanner_init "my-scanner" "1.0.0"
#   scanner_header "expressjs/express"
#   scanner_progress_start "Analyzing packages" 100
#   scanner_progress_update 50
#   scanner_progress_end
#   scanner_summary_start
#   scanner_summary_metric "Vulnerabilities" "12" "warning"
#   scanner_summary_end
#   scanner_footer "success"
#
# Design Principles:
# - All user messages to stderr, data to stdout
# - Consistent colors across all scanners
# - Unified progress indicators
# - Standardized report formats
#############################################################################

#############################################################################
# COLOR PALETTE
# Standard 7-color palette used across all scanners
#############################################################################

# Only set colors if terminal supports them and not in CI
if [[ -t 2 ]] && [[ -z "${NO_COLOR:-}" ]]; then
    SCANNER_RED='\033[0;31m'
    SCANNER_GREEN='\033[0;32m'
    SCANNER_YELLOW='\033[1;33m'
    SCANNER_BLUE='\033[0;34m'
    SCANNER_CYAN='\033[0;36m'
    SCANNER_DIM='\033[0;90m'
    SCANNER_BOLD='\033[1m'
    SCANNER_NC='\033[0m'
else
    SCANNER_RED=''
    SCANNER_GREEN=''
    SCANNER_YELLOW=''
    SCANNER_BLUE=''
    SCANNER_CYAN=''
    SCANNER_DIM=''
    SCANNER_BOLD=''
    SCANNER_NC=''
fi

# Export for subshells
export SCANNER_RED SCANNER_GREEN SCANNER_YELLOW SCANNER_BLUE SCANNER_CYAN SCANNER_DIM SCANNER_BOLD SCANNER_NC

#############################################################################
# STATUS INDICATORS
# Unicode symbols for status messages
#############################################################################

SCANNER_CHECK="${SCANNER_GREEN}✓${SCANNER_NC}"
SCANNER_CROSS="${SCANNER_RED}✗${SCANNER_NC}"
SCANNER_WARN="${SCANNER_YELLOW}⚠${SCANNER_NC}"
SCANNER_INFO="${SCANNER_BLUE}ℹ${SCANNER_NC}"
SCANNER_ARROW="${SCANNER_CYAN}→${SCANNER_NC}"

# Risk level indicators
SCANNER_RISK_CRITICAL="${SCANNER_RED}●${SCANNER_NC}"
SCANNER_RISK_HIGH="${SCANNER_YELLOW}●${SCANNER_NC}"
SCANNER_RISK_MEDIUM="${SCANNER_YELLOW}○${SCANNER_NC}"
SCANNER_RISK_LOW="${SCANNER_GREEN}○${SCANNER_NC}"

#############################################################################
# SCANNER STATE
# Internal state tracking
#############################################################################

_SCANNER_NAME=""
_SCANNER_VERSION=""
_SCANNER_START_TIME=""
_SCANNER_VERBOSE=false
_SCANNER_QUIET=false
_SCANNER_FORMAT="terminal"  # terminal, json, markdown

# Progress tracking
_PROGRESS_LABEL=""
_PROGRESS_TOTAL=0
_PROGRESS_CURRENT=0
_PROGRESS_START_TIME=""

#############################################################################
# INITIALIZATION
#############################################################################

# Initialize scanner with name and version
# Usage: scanner_init "my-scanner" "1.0.0" [--verbose] [--quiet] [--format FORMAT]
scanner_init() {
    _SCANNER_NAME="${1:-scanner}"
    _SCANNER_VERSION="${2:-1.0.0}"
    shift 2 || true

    while [[ $# -gt 0 ]]; do
        case "$1" in
            --verbose|-v) _SCANNER_VERBOSE=true; shift ;;
            --quiet|-q) _SCANNER_QUIET=true; shift ;;
            --format|-f) _SCANNER_FORMAT="$2"; shift 2 ;;
            *) shift ;;
        esac
    done

    _SCANNER_START_TIME=$(date +%s)
}

#############################################################################
# LOGGING & MESSAGES
# Status messages to stderr
#############################################################################

# Log debug message (only if verbose)
scanner_debug() {
    [[ "$_SCANNER_VERBOSE" == "true" ]] || return 0
    echo -e "${SCANNER_DIM}[debug] $*${SCANNER_NC}" >&2
}

# Log info message
scanner_info() {
    [[ "$_SCANNER_QUIET" == "true" ]] && return 0
    echo -e "${SCANNER_BLUE}${SCANNER_INFO}${SCANNER_NC} $*" >&2
}

# Log success message
scanner_success() {
    [[ "$_SCANNER_QUIET" == "true" ]] && return 0
    echo -e "${SCANNER_GREEN}${SCANNER_CHECK}${SCANNER_NC} $*" >&2
}

# Log warning message
scanner_warn() {
    echo -e "${SCANNER_YELLOW}${SCANNER_WARN}${SCANNER_NC} $*" >&2
}

# Log error message
scanner_error() {
    echo -e "${SCANNER_RED}${SCANNER_CROSS}${SCANNER_NC} $*" >&2
}

# Log step (for multi-step processes)
scanner_step() {
    local step_name="$1"
    local status="${2:-running}"  # running, success, warning, error

    case "$status" in
        running)
            echo -e "${SCANNER_BLUE}→${SCANNER_NC} ${step_name}..." >&2
            ;;
        success)
            echo -e "${SCANNER_GREEN}✓${SCANNER_NC} ${step_name}" >&2
            ;;
        warning)
            echo -e "${SCANNER_YELLOW}⚠${SCANNER_NC} ${step_name}" >&2
            ;;
        error)
            echo -e "${SCANNER_RED}✗${SCANNER_NC} ${step_name}" >&2
            ;;
    esac
}

#############################################################################
# HEADERS & FOOTERS
#############################################################################

# Print scanner header
# Usage: scanner_header "target-name" [--no-banner]
scanner_header() {
    local target="${1:-}"
    local show_banner=true

    [[ "${2:-}" == "--no-banner" ]] && show_banner=false
    [[ "$_SCANNER_QUIET" == "true" ]] && return 0

    if [[ "$show_banner" == "true" ]]; then
        echo -e "${SCANNER_CYAN}" >&2
        cat >&2 << 'EOF'
███████╗███████╗██████╗  ██████╗
╚══███╔╝██╔════╝██╔══██╗██╔═══██╗
  ███╔╝ █████╗  ██████╔╝██║   ██║
 ███╔╝  ██╔══╝  ██╔══██╗██║   ██║
███████╗███████╗██║  ██║╚██████╔╝
╚══════╝╚══════╝╚═╝  ╚═╝ ╚═════╝
EOF
        echo -e "${SCANNER_NC}" >&2
        echo -e "${SCANNER_DIM}crashoverride.com${SCANNER_NC}" >&2
        echo "" >&2
    fi

    echo -e "${SCANNER_BOLD}Scanner:${SCANNER_NC} $_SCANNER_NAME v$_SCANNER_VERSION" >&2

    if [[ -n "$target" ]]; then
        echo -e "${SCANNER_BOLD}Target:${SCANNER_NC}  $target" >&2
    fi

    echo -e "${SCANNER_BOLD}Started:${SCANNER_NC} $(date '+%Y-%m-%d %H:%M:%S')" >&2
    echo "" >&2
}

# Print scanner footer with duration
# Usage: scanner_footer "success" | "warning" | "error"
scanner_footer() {
    local status="${1:-success}"
    local end_time=$(date +%s)
    local duration=$((end_time - _SCANNER_START_TIME))

    [[ "$_SCANNER_QUIET" == "true" ]] && return 0

    echo "" >&2

    local status_icon status_text status_color
    case "$status" in
        success)
            status_icon="$SCANNER_CHECK"
            status_text="Completed successfully"
            status_color="$SCANNER_GREEN"
            ;;
        warning)
            status_icon="$SCANNER_WARN"
            status_text="Completed with warnings"
            status_color="$SCANNER_YELLOW"
            ;;
        error)
            status_icon="$SCANNER_CROSS"
            status_text="Failed"
            status_color="$SCANNER_RED"
            ;;
    esac

    echo -e "${status_color}${status_icon} ${status_text}${SCANNER_NC} ${SCANNER_DIM}(${duration}s)${SCANNER_NC}" >&2
}

#############################################################################
# PROGRESS INDICATORS
#############################################################################

# Start progress tracking
# Usage: scanner_progress_start "Analyzing packages" 100
scanner_progress_start() {
    _PROGRESS_LABEL="${1:-Processing}"
    _PROGRESS_TOTAL="${2:-100}"
    _PROGRESS_CURRENT=0
    _PROGRESS_START_TIME=$(date +%s)

    [[ "$_SCANNER_QUIET" == "true" ]] && return 0

    # Initial display
    printf "\r${SCANNER_BLUE}${_PROGRESS_LABEL}${SCANNER_NC}: 0/${_PROGRESS_TOTAL}" >&2
}

# Update progress
# Usage: scanner_progress_update 50 ["optional item name"]
scanner_progress_update() {
    _PROGRESS_CURRENT="${1:-$((_PROGRESS_CURRENT + 1))}"
    local item_name="${2:-}"

    [[ "$_SCANNER_QUIET" == "true" ]] && return 0

    # Calculate percentage
    local pct=0
    if [[ $_PROGRESS_TOTAL -gt 0 ]]; then
        pct=$(( (_PROGRESS_CURRENT * 100) / _PROGRESS_TOTAL ))
    fi

    # Build progress bar (20 chars wide)
    local filled=$(( (pct * 20) / 100 ))
    local empty=$((20 - filled))
    local bar=""
    for ((i=0; i<filled; i++)); do bar+="█"; done
    for ((i=0; i<empty; i++)); do bar+="░"; done

    # Calculate ETA
    local elapsed=$(($(date +%s) - _PROGRESS_START_TIME))
    local eta=""
    if [[ $_PROGRESS_CURRENT -gt 0 ]] && [[ $elapsed -gt 0 ]]; then
        local remaining=$(( (elapsed * (_PROGRESS_TOTAL - _PROGRESS_CURRENT)) / _PROGRESS_CURRENT ))
        if [[ $remaining -gt 60 ]]; then
            eta=" ETA: $((remaining / 60))m"
        elif [[ $remaining -gt 0 ]]; then
            eta=" ETA: ${remaining}s"
        fi
    fi

    # Clear line and print progress
    if [[ -n "$item_name" ]]; then
        # Truncate item name if too long
        [[ ${#item_name} -gt 30 ]] && item_name="${item_name:0:27}..."
        printf "\r\033[K${SCANNER_BLUE}${_PROGRESS_LABEL}${SCANNER_NC} [${bar}] ${pct}%% ${_PROGRESS_CURRENT}/${_PROGRESS_TOTAL}${eta} ${SCANNER_DIM}${item_name}${SCANNER_NC}" >&2
    else
        printf "\r\033[K${SCANNER_BLUE}${_PROGRESS_LABEL}${SCANNER_NC} [${bar}] ${pct}%% ${_PROGRESS_CURRENT}/${_PROGRESS_TOTAL}${eta}" >&2
    fi
}

# End progress tracking
scanner_progress_end() {
    [[ "$_SCANNER_QUIET" == "true" ]] && return 0

    local elapsed=$(($(date +%s) - _PROGRESS_START_TIME))
    printf "\r\033[K${SCANNER_GREEN}✓${SCANNER_NC} ${_PROGRESS_LABEL}: ${_PROGRESS_TOTAL} items ${SCANNER_DIM}(${elapsed}s)${SCANNER_NC}\n" >&2

    # Reset
    _PROGRESS_LABEL=""
    _PROGRESS_TOTAL=0
    _PROGRESS_CURRENT=0
}

# Simple spinner for indeterminate progress
# Usage: scanner_spinner_start "Loading..."
#        ... do work ...
#        scanner_spinner_stop
_SPINNER_PID=""

scanner_spinner_start() {
    local message="${1:-Processing}"

    [[ "$_SCANNER_QUIET" == "true" ]] && return 0

    (
        local frames=("⠋" "⠙" "⠹" "⠸" "⠼" "⠴" "⠦" "⠧" "⠇" "⠏")
        local i=0
        while true; do
            printf "\r${SCANNER_BLUE}${frames[$i]}${SCANNER_NC} ${message}" >&2
            i=$(( (i + 1) % ${#frames[@]} ))
            sleep 0.1
        done
    ) &
    _SPINNER_PID=$!
    disown $_SPINNER_PID 2>/dev/null || true
}

scanner_spinner_stop() {
    if [[ -n "$_SPINNER_PID" ]]; then
        kill $_SPINNER_PID 2>/dev/null || true
        wait $_SPINNER_PID 2>/dev/null || true
        _SPINNER_PID=""
        printf "\r\033[K" >&2
    fi
}

#############################################################################
# SUMMARY SECTION
# Terminal summary display
#############################################################################

_SUMMARY_METRICS=()

# Start summary section
scanner_summary_start() {
    local title="${1:-Summary}"

    [[ "$_SCANNER_QUIET" == "true" ]] && return 0

    echo "" >&2
    echo -e "${SCANNER_BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${SCANNER_NC}" >&2
    echo -e "${SCANNER_BOLD}${title}${SCANNER_NC}" >&2
    echo -e "${SCANNER_BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${SCANNER_NC}" >&2

    _SUMMARY_METRICS=()
}

# Add metric to summary
# Usage: scanner_summary_metric "Name" "Value" "status"
# Status: good, warning, error, info
scanner_summary_metric() {
    local name="$1"
    local value="$2"
    local status="${3:-info}"

    [[ "$_SCANNER_QUIET" == "true" ]] && return 0

    local status_icon
    case "$status" in
        good|success)   status_icon="${SCANNER_GREEN}●${SCANNER_NC}" ;;
        warning)        status_icon="${SCANNER_YELLOW}●${SCANNER_NC}" ;;
        error|critical) status_icon="${SCANNER_RED}●${SCANNER_NC}" ;;
        *)              status_icon="${SCANNER_BLUE}●${SCANNER_NC}" ;;
    esac

    printf "  ${status_icon} %-28s %s\n" "$name:" "$value" >&2

    _SUMMARY_METRICS+=("$name|$value|$status")
}

# Add section break in summary
scanner_summary_section() {
    local title="$1"

    [[ "$_SCANNER_QUIET" == "true" ]] && return 0

    echo "" >&2
    echo -e "  ${SCANNER_DIM}─── ${title} ───${SCANNER_NC}" >&2
}

# End summary section
scanner_summary_end() {
    [[ "$_SCANNER_QUIET" == "true" ]] && return 0

    echo "" >&2
}

#############################################################################
# TABLES
# ASCII table rendering
#############################################################################

# Print a simple table
# Usage: scanner_table "Header1|Header2|Header3" "Row1Col1|Row1Col2|Row1Col3" "Row2..."
scanner_table() {
    local header="$1"
    shift

    [[ "$_SCANNER_QUIET" == "true" ]] && return 0

    # Calculate column widths
    IFS='|' read -ra cols <<< "$header"
    local num_cols=${#cols[@]}
    local widths=()

    # Initialize with header widths
    for col in "${cols[@]}"; do
        widths+=("${#col}")
    done

    # Update with data widths
    for row in "$@"; do
        IFS='|' read -ra cells <<< "$row"
        for i in "${!cells[@]}"; do
            local len=${#cells[$i]}
            [[ $len -gt ${widths[$i]:-0} ]] && widths[$i]=$len
        done
    done

    # Print header
    local border="+"
    local header_row="|"
    for i in "${!cols[@]}"; do
        local w=$((${widths[$i]} + 2))
        border+=$(printf '%*s' "$w" | tr ' ' '-')+"+"
        header_row+=$(printf " %-${widths[$i]}s |" "${cols[$i]}")
    done

    echo "$border" >&2
    echo "$header_row" >&2
    echo "$border" >&2

    # Print data rows
    for row in "$@"; do
        IFS='|' read -ra cells <<< "$row"
        local data_row="|"
        for i in "${!widths[@]}"; do
            data_row+=$(printf " %-${widths[$i]}s |" "${cells[$i]:-}")
        done
        echo "$data_row" >&2
    done

    echo "$border" >&2
}

#############################################################################
# OUTPUT FORMATTING
# Format data for different output types
#############################################################################

# Output JSON to stdout
scanner_output_json() {
    local data="$1"
    echo "$data"
}

# Add scanner metadata to JSON output
scanner_wrap_json() {
    local data="$1"

    local end_time=$(date +%s)
    local duration=$((end_time - _SCANNER_START_TIME))

    jq -n \
        --arg scanner "$_SCANNER_NAME" \
        --arg version "$_SCANNER_VERSION" \
        --arg timestamp "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        --argjson duration "$duration" \
        --argjson data "$data" \
        '{
            metadata: {
                scanner: $scanner,
                version: $version,
                timestamp: $timestamp,
                duration_seconds: $duration
            },
            data: $data
        }'
}

#############################################################################
# DEPENDENCY CHECKING
#############################################################################

# Check if a command exists
# Usage: scanner_require "jq" "brew install jq"
scanner_require() {
    local cmd="$1"
    local install_hint="${2:-}"

    if ! command -v "$cmd" &>/dev/null; then
        scanner_error "Required command not found: $cmd"
        if [[ -n "$install_hint" ]]; then
            echo -e "  ${SCANNER_DIM}Install: ${install_hint}${SCANNER_NC}" >&2
        fi
        return 1
    fi
    return 0
}

# Check multiple dependencies
# Usage: scanner_require_all "jq" "curl" "git"
scanner_require_all() {
    local missing=()
    for cmd in "$@"; do
        command -v "$cmd" &>/dev/null || missing+=("$cmd")
    done

    if [[ ${#missing[@]} -gt 0 ]]; then
        scanner_error "Missing required commands: ${missing[*]}"
        return 1
    fi
    return 0
}

#############################################################################
# RISK DISPLAY HELPERS
#############################################################################

# Get risk indicator
# Usage: scanner_risk_indicator "critical"
scanner_risk_indicator() {
    local level="$1"

    case "${level,,}" in
        critical) echo -e "${SCANNER_RED}●${SCANNER_NC} Critical" ;;
        high)     echo -e "${SCANNER_RED}○${SCANNER_NC} High" ;;
        medium)   echo -e "${SCANNER_YELLOW}●${SCANNER_NC} Medium" ;;
        low)      echo -e "${SCANNER_GREEN}○${SCANNER_NC} Low" ;;
        info)     echo -e "${SCANNER_BLUE}○${SCANNER_NC} Info" ;;
        *)        echo -e "${SCANNER_DIM}○${SCANNER_NC} Unknown" ;;
    esac
}

# Display risk counts
# Usage: scanner_risk_summary 3 7 12 5  # critical high medium low
scanner_risk_summary() {
    local critical="${1:-0}"
    local high="${2:-0}"
    local medium="${3:-0}"
    local low="${4:-0}"

    [[ "$_SCANNER_QUIET" == "true" ]] && return 0

    [[ $critical -gt 0 ]] && echo -e "  ${SCANNER_RED}●${SCANNER_NC} Critical: $critical" >&2
    [[ $high -gt 0 ]]     && echo -e "  ${SCANNER_RED}○${SCANNER_NC} High: $high" >&2
    [[ $medium -gt 0 ]]   && echo -e "  ${SCANNER_YELLOW}●${SCANNER_NC} Medium: $medium" >&2
    [[ $low -gt 0 ]]      && echo -e "  ${SCANNER_GREEN}○${SCANNER_NC} Low: $low" >&2
}

#############################################################################
# EXPORT FUNCTIONS
#############################################################################

export -f scanner_init
export -f scanner_debug
export -f scanner_info
export -f scanner_success
export -f scanner_warn
export -f scanner_error
export -f scanner_step
export -f scanner_header
export -f scanner_footer
export -f scanner_progress_start
export -f scanner_progress_update
export -f scanner_progress_end
export -f scanner_spinner_start
export -f scanner_spinner_stop
export -f scanner_summary_start
export -f scanner_summary_metric
export -f scanner_summary_section
export -f scanner_summary_end
export -f scanner_table
export -f scanner_output_json
export -f scanner_wrap_json
export -f scanner_require
export -f scanner_require_all
export -f scanner_risk_indicator
export -f scanner_risk_summary
