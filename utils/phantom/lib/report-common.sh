#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Report Common Utilities
# Shared functions for report generation
#############################################################################

# Ensure we only source this once
[[ -n "$_REPORT_COMMON_LOADED" ]] && return 0
_REPORT_COMMON_LOADED=true

#############################################################################
# Constants
#############################################################################

REPORT_VERSION="1.0.0"

# Report types
REPORT_TYPES=("summary" "security" "licenses" "compliance" "sbom" "supply-chain" "dora" "code-ownership" "ai-adoption" "full")

# Output formats
REPORT_FORMATS=("terminal" "markdown" "json" "html" "csv")

# Risk level colors (terminal) - bash 3.x compatible
# Use functions instead of associative arrays
_get_risk_color_value() {
    case "$1" in
        critical) echo "$RED" ;;
        high) echo "$RED" ;;
        medium) echo "$YELLOW" ;;
        low) echo "$GREEN" ;;
        none) echo "$GREEN" ;;
        *) echo "$DIM" ;;
    esac
}

# Risk level emojis (markdown/html)
_get_risk_emoji_value() {
    case "$1" in
        critical) echo "🔴" ;;
        high) echo "🟠" ;;
        medium) echo "🟡" ;;
        low) echo "🟢" ;;
        none) echo "✅" ;;
        *) echo "⚪" ;;
    esac
}

# Severity order for sorting
_get_severity_order() {
    case "$1" in
        critical) echo 1 ;;
        high) echo 2 ;;
        medium) echo 3 ;;
        low) echo 4 ;;
        none) echo 5 ;;
        *) echo 6 ;;
    esac
}

#############################################################################
# Data Loading Functions
#############################################################################

# Load scanner data with fallback
# Usage: load_scanner_data <analysis_path> <scanner_name>
# Returns: JSON data or empty object
load_scanner_data() {
    local analysis_path="$1"
    local scanner="$2"
    local file="$analysis_path/${scanner}.json"

    if [[ -f "$file" ]]; then
        cat "$file"
    else
        echo '{}'
    fi
}

# Check if scanner data exists
# Usage: has_scanner_data <analysis_path> <scanner_name>
has_scanner_data() {
    local analysis_path="$1"
    local scanner="$2"
    [[ -f "$analysis_path/${scanner}.json" ]]
}

# Get summary field from scanner data
# Usage: get_summary_field <json_data> <field> [default]
get_summary_field() {
    local json="$1"
    local field="$2"
    local default="${3:-0}"

    local value=$(echo "$json" | jq -r ".summary.$field // \"$default\"" 2>/dev/null)
    [[ "$value" == "null" ]] && value="$default"
    echo "$value"
}

# Load manifest data
# Usage: load_manifest <analysis_path>
load_manifest() {
    local analysis_path="$1"
    local manifest="$analysis_path/manifest.json"

    if [[ -f "$manifest" ]]; then
        cat "$manifest"
    else
        echo '{"error": "No manifest found"}'
    fi
}

#############################################################################
# Time Utilities
#############################################################################

# Format relative time
# Usage: relative_time <iso_timestamp>
relative_time() {
    local timestamp="$1"
    if [[ -z "$timestamp" ]] || [[ "$timestamp" == "null" ]]; then
        echo "unknown"
        return
    fi

    # Try macOS date format first (with TZ=UTC for Z suffix), then GNU
    local ts_epoch
    if [[ "$timestamp" == *Z ]]; then
        # macOS needs TZ=UTC to correctly interpret Z suffix
        ts_epoch=$(TZ=UTC date -j -f "%Y-%m-%dT%H:%M:%SZ" "$timestamp" +%s 2>/dev/null)
    fi
    [[ -z "$ts_epoch" ]] && ts_epoch=$(date -d "$timestamp" +%s 2>/dev/null)
    [[ -z "$ts_epoch" ]] && { echo "unknown"; return; }

    local now_epoch=$(date +%s)
    local diff=$((now_epoch - ts_epoch))

    if [[ $diff -lt 60 ]]; then
        echo "just now"
    elif [[ $diff -lt 3600 ]]; then
        local mins=$((diff / 60))
        echo "${mins}m ago"
    elif [[ $diff -lt 86400 ]]; then
        local hours=$((diff / 3600))
        echo "${hours}h ago"
    elif [[ $diff -lt 604800 ]]; then
        local days=$((diff / 86400))
        echo "${days}d ago"
    else
        local weeks=$((diff / 604800))
        echo "${weeks}w ago"
    fi
}

# Format timestamp for display
# Usage: format_timestamp <iso_timestamp>
format_timestamp() {
    local timestamp="$1"
    if [[ -z "$timestamp" ]] || [[ "$timestamp" == "null" ]]; then
        echo "unknown"
        return
    fi

    # Extract date and time parts
    echo "$timestamp" | sed 's/T/ /' | cut -d':' -f1,2 | cut -d'.' -f1
}

# Check if data is stale (>24 hours)
# Usage: is_data_stale <iso_timestamp>
is_data_stale() {
    local timestamp="$1"
    if [[ -z "$timestamp" ]] || [[ "$timestamp" == "null" ]]; then
        return 0  # Unknown = stale
    fi

    local ts_epoch
    if [[ "$timestamp" == *Z ]]; then
        ts_epoch=$(TZ=UTC date -j -f "%Y-%m-%dT%H:%M:%SZ" "$timestamp" +%s 2>/dev/null)
    fi
    [[ -z "$ts_epoch" ]] && ts_epoch=$(date -d "$timestamp" +%s 2>/dev/null)
    [[ -z "$ts_epoch" ]] && return 0

    local now_epoch=$(date +%s)
    local diff=$((now_epoch - ts_epoch))

    [[ $diff -gt 86400 ]]
}

#############################################################################
# Risk Calculation
#############################################################################

# Calculate overall risk level from vulnerability counts
# Usage: calculate_risk_level <critical> <high> <medium>
calculate_risk_level() {
    local critical="${1:-0}"
    local high="${2:-0}"
    local medium="${3:-0}"

    if [[ $critical -gt 0 ]]; then
        echo "critical"
    elif [[ $high -gt 0 ]]; then
        echo "high"
    elif [[ $medium -gt 5 ]]; then
        echo "medium"
    elif [[ $medium -gt 0 ]]; then
        echo "low"
    else
        echo "none"
    fi
}

# Get risk color for terminal output
# Usage: get_risk_color <risk_level>
get_risk_color() {
    local risk="$1"
    _get_risk_color_value "$risk"
}

# Get risk emoji for markdown/html
# Usage: get_risk_emoji <risk_level>
get_risk_emoji() {
    local risk="$1"
    _get_risk_emoji_value "$risk"
}

#############################################################################
# Number Formatting
#############################################################################

# Format large numbers with K/M suffix
# Usage: format_number <number>
format_number() {
    local num="$1"
    if [[ $num -ge 1000000 ]]; then
        printf "%.1fM" "$(echo "scale=1; $num/1000000" | bc)"
    elif [[ $num -ge 1000 ]]; then
        printf "%.1fK" "$(echo "scale=1; $num/1000" | bc)"
    else
        echo "$num"
    fi
}

# Format percentage
# Usage: format_percentage <value> [decimals]
format_percentage() {
    local value="$1"
    local decimals="${2:-1}"
    printf "%.${decimals}f%%" "$value"
}

#############################################################################
# Progress Bar Utilities
#############################################################################

# Generate ASCII progress bar
# Usage: progress_bar <value> <max> [width]
progress_bar() {
    local value="$1"
    local max="$2"
    local width="${3:-20}"

    local filled=$((value * width / max))
    local empty=$((width - filled))

    printf "%s%s" \
        "$(printf '█%.0s' $(seq 1 $filled 2>/dev/null) || echo "")" \
        "$(printf '░%.0s' $(seq 1 $empty 2>/dev/null) || echo "")"
}

# Generate risk progress bar with color
# Usage: risk_bar <risk_level> [width]
risk_bar() {
    local risk="$1"
    local width="${2:-20}"

    local filled
    case "$risk" in
        critical) filled=$width ;;
        high) filled=$((width * 3 / 4)) ;;
        medium) filled=$((width / 2)) ;;
        low) filled=$((width / 4)) ;;
        *) filled=0 ;;
    esac

    local empty=$((width - filled))
    local color=$(get_risk_color "$risk")

    printf "${color}%s${NC}%s" \
        "$(printf '█%.0s' $(seq 1 $filled 2>/dev/null) || echo "")" \
        "$(printf '░%.0s' $(seq 1 $empty 2>/dev/null) || echo "")"
}

#############################################################################
# Table Utilities
#############################################################################

# Print a horizontal rule
# Usage: hr [char] [width]
hr() {
    local char="${1:-━}"
    local width="${2:-66}"
    printf '%*s\n' "$width" '' | tr ' ' "$char"
}

# Print centered text
# Usage: center_text <text> [width]
center_text() {
    local text="$1"
    local width="${2:-66}"
    local text_len=${#text}
    local padding=$(( (width - text_len) / 2 ))
    printf "%*s%s\n" $padding '' "$text"
}

# Print a key-value pair
# Usage: kv <key> <value> [key_width]
kv() {
    local key="$1"
    local value="$2"
    local key_width="${3:-14}"
    printf "  %-${key_width}s %s\n" "$key:" "$value"
}

# Print a key-value pair with colored value
# Usage: kv_color <key> <value> <color> [key_width]
kv_color() {
    local key="$1"
    local value="$2"
    local color="$3"
    local key_width="${4:-14}"
    printf "  %-${key_width}s ${color}%s${NC}\n" "$key:" "$value"
}

#############################################################################
# Validation
#############################################################################

# Validate report type
# Usage: validate_report_type <type>
validate_report_type() {
    local type="$1"
    for t in "${REPORT_TYPES[@]}"; do
        [[ "$t" == "$type" ]] && return 0
    done
    return 1
}

# Validate output format
# Usage: validate_format <format>
validate_format() {
    local format="$1"
    for f in "${REPORT_FORMATS[@]}"; do
        [[ "$f" == "$format" ]] && return 0
    done
    return 1
}

#############################################################################
# Report Header/Footer
#############################################################################

# Generate report header
# Usage: report_header <title> <project_id>
report_header() {
    local title="$1"
    local project_id="$2"

    echo
    echo -e "${BOLD}${title}${NC}"
    hr
    echo
}

# Generate report footer
# Usage: report_footer
report_footer() {
    echo
    hr "─" 66
    echo -e "${DIM}Generated by Phantom Report v${REPORT_VERSION}${NC}"
    echo -e "${DIM}$(date '+%Y-%m-%d %H:%M:%S %Z')${NC}"
    echo
}

#############################################################################
# Data Aggregation
#############################################################################

# Aggregate vulnerability counts from package-vulns.json
# Usage: aggregate_vulns <analysis_path>
# Output: JSON object with critical, high, medium, low, total
aggregate_vulns() {
    local analysis_path="$1"
    local vulns_file="$analysis_path/package-vulns.json"

    if [[ -f "$vulns_file" ]]; then
        jq '{
            critical: (.summary.critical // 0),
            high: (.summary.high // 0),
            medium: (.summary.medium // 0),
            low: (.summary.low // 0),
            total: (.summary.total // 0)
        }' "$vulns_file" 2>/dev/null
    else
        echo '{"critical":0,"high":0,"medium":0,"low":0,"total":0}'
    fi
}

# Aggregate dependency counts from package-sbom.json
# Usage: aggregate_deps <analysis_path>
aggregate_deps() {
    local analysis_path="$1"
    local sbom_file="$analysis_path/package-sbom.json"

    if [[ -f "$sbom_file" ]]; then
        jq '{
            total: (.total_dependencies // .summary.total // 0),
            direct: (.direct_dependencies // .summary.direct // 0)
        }' "$sbom_file" 2>/dev/null
    else
        echo '{"total":0,"direct":0}'
    fi
}

# Get top N issues for summary
# Usage: get_top_issues <analysis_path> [limit]
get_top_issues() {
    local analysis_path="$1"
    local limit="${2:-3}"
    local issues=()

    # Check critical vulns
    if has_scanner_data "$analysis_path" "package-vulns"; then
        local vulns=$(load_scanner_data "$analysis_path" "package-vulns")
        local critical=$(echo "$vulns" | jq -r '.summary.critical // 0')
        local high=$(echo "$vulns" | jq -r '.summary.high // 0')

        if [[ $critical -gt 0 ]]; then
            issues+=("🔴 $critical critical vulnerabilities require immediate attention")
        fi
        if [[ $high -gt 0 ]]; then
            issues+=("🟠 $high high severity vulnerabilities found")
        fi
    fi

    # Check secrets
    if has_scanner_data "$analysis_path" "code-secrets"; then
        local secrets=$(load_scanner_data "$analysis_path" "code-secrets")
        local total=$(echo "$secrets" | jq -r '.summary.total_findings // 0')
        if [[ $total -gt 0 ]]; then
            issues+=("🔑 $total exposed secrets detected")
        fi
    fi

    # Check abandoned packages
    if has_scanner_data "$analysis_path" "package-health"; then
        local health=$(load_scanner_data "$analysis_path" "package-health")
        local abandoned=$(echo "$health" | jq -r '.summary.abandoned // 0')
        if [[ $abandoned -gt 0 ]]; then
            issues+=("📦 $abandoned abandoned packages in dependencies")
        fi
    fi

    # Check license violations
    if has_scanner_data "$analysis_path" "licenses"; then
        local licenses=$(load_scanner_data "$analysis_path" "licenses")
        local violations=$(echo "$licenses" | jq -r '.summary.license_violations // 0')
        if [[ $violations -gt 0 ]]; then
            issues+=("⚖️ $violations license compliance violations")
        fi
    fi

    # Return top N
    printf '%s\n' "${issues[@]}" | head -n "$limit"
}

#############################################################################
# Export all functions
#############################################################################

export -f load_scanner_data
export -f has_scanner_data
export -f get_summary_field
export -f load_manifest
export -f relative_time
export -f format_timestamp
export -f is_data_stale
export -f calculate_risk_level
export -f get_risk_color
export -f get_risk_emoji
export -f format_number
export -f format_percentage
export -f progress_bar
export -f risk_bar
export -f hr
export -f center_text
export -f kv
export -f kv_color
export -f validate_report_type
export -f validate_format
export -f report_header
export -f report_footer
export -f aggregate_vulns
export -f aggregate_deps
export -f get_top_issues
