#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Malcontent Scanner
#
# Supply chain compromise detection using Chainguard's malcontent tool.
# Identifies malicious code through contextual analysis, differential
# analysis, and 14,500+ YARA rules from security vendors.
#
# Modes:
#   - analyze: Enumerate program capabilities (default)
#   - diff:    Compare two versions for risky changes
#   - scan:    Basic malware scanning
#
# Usage: ./malcontent.sh [options] <repo_path>
#
# See: https://github.com/chainguard-dev/malcontent
#############################################################################

set -e

# Colors for terminal output (stderr only)
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
DIM='\033[0;90m'
NC='\033[0m'

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SCANNERS_ROOT="$(dirname "$SCRIPT_DIR")"
UTILS_ROOT="$(dirname "$SCANNERS_ROOT")"
REPO_ROOT="$(dirname "$UTILS_ROOT")"
PROFILES_DIR="$SCRIPT_DIR/profiles"

# Load shared libraries from utils/lib
if [[ -f "$UTILS_ROOT/lib/sbom.sh" ]]; then
    source "$UTILS_ROOT/lib/sbom.sh"
fi

if [[ -f "$UTILS_ROOT/lib/package-download.sh" ]]; then
    source "$UTILS_ROOT/lib/package-download.sh"
fi

# Default options
OUTPUT_FILE=""
REPO_PATH=""
SBOM_FILE=""
MODE="analyze"       # analyze, diff, scan
PROFILE="default"    # default, security, quick
MIN_RISK="medium"    # low, medium, high, critical
FORMAT="json"        # json, yaml, markdown, terminal
VERBOSE=false
SHOW_FINDINGS=false  # Print detailed findings to terminal
DIFF_COMPARE=""      # Path or version for diff mode
DOWNLOAD_PACKAGES=false  # Download and scan package artifacts
REPO_NAME=""         # Repository name for display (e.g., owner/repo)

usage() {
    cat << EOF
Malcontent Scanner - Supply chain compromise detection

Usage: $0 [OPTIONS] <repo_path>

OPTIONS:
    --mode MODE             Scan mode: analyze, diff, scan (default: analyze)
    --profile PROFILE       Scan profile: default, security, quick (default: default)
    --sbom FILE             SBOM file to scan packages from (CycloneDX JSON)
    --download-packages     Download and scan package artifacts from SBOM
    --diff PATH             Compare against this path/version (enables diff mode)
    --min-risk LEVEL        Minimum risk level: low, medium, high, critical (default: medium)
    --format FORMAT         Output format: json, yaml, markdown, terminal (default: json)
    --verbose               Show progress messages
    --show-findings         Print detailed findings to terminal with code snippets
    --repo-name NAME        Repository name for display (e.g., owner/repo)
    -o, --output FILE       Write output to file (default: stdout)
    -h, --help              Show this help

MODES:
    analyze             Enumerate program capabilities categorized by risk level
    diff                Compare two versions to identify risky changes
    scan                Basic malware scanning on directories or container images

PROFILES:
    default             Standard analysis with medium risk threshold
    security            Comprehensive security scan with low threshold
    quick               Fast scan with high threshold (critical only)

SBOM INTEGRATION:
    When --sbom is provided, the scanner will extract package information
    and scan downloaded package artifacts for malicious content.

EXAMPLES:
    $0 /path/to/repo
    $0 --mode scan /path/to/repo
    $0 --profile security --sbom sbom.cdx.json /path/to/repo
    $0 --diff /path/to/old-version /path/to/new-version
    $0 --min-risk critical -o findings.json /path/to/repo

OUTPUT:
    JSON object with:
    - summary: counts by risk level, category
    - findings: array of matches with file, risk, behavior
    - packages: package-level analysis (if SBOM provided)
    - diff: changes between versions (if diff mode)

EOF
    exit 0
}

# Find malcontent binary (installed as 'mal' via brew)
find_malcontent_bin() {
    if command -v mal &> /dev/null; then
        echo "mal"
    elif [[ -x "/opt/homebrew/bin/mal" ]]; then
        echo "/opt/homebrew/bin/mal"
    elif [[ -x "/usr/local/bin/mal" ]]; then
        echo "/usr/local/bin/mal"
    else
        echo ""
    fi
}

MAL_BIN=""

# Check if malcontent is installed
check_malcontent() {
    MAL_BIN=$(find_malcontent_bin)
    if [[ -z "$MAL_BIN" ]]; then
        echo '{"error": "malcontent not installed", "install": "brew install malcontent"}' >&2
        exit 1
    fi
}

# Parse arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --mode)
                MODE="$2"
                shift 2
                ;;
            --profile)
                PROFILE="$2"
                shift 2
                ;;
            --sbom)
                SBOM_FILE="$2"
                shift 2
                ;;
            --download-packages)
                DOWNLOAD_PACKAGES=true
                shift
                ;;
            --diff)
                DIFF_COMPARE="$2"
                MODE="diff"
                shift 2
                ;;
            --min-risk)
                MIN_RISK="$2"
                shift 2
                ;;
            --format)
                FORMAT="$2"
                shift 2
                ;;
            --verbose)
                VERBOSE=true
                shift
                ;;
            --show-findings)
                SHOW_FINDINGS=true
                shift
                ;;
            --repo-name)
                REPO_NAME="$2"
                shift 2
                ;;
            --local-path)
                # Alias for compatibility with other scanners
                REPO_PATH="$2"
                shift 2
                ;;
            -o|--output)
                OUTPUT_FILE="$2"
                shift 2
                ;;
            -h|--help)
                usage
                ;;
            *)
                if [[ -z "$REPO_PATH" ]]; then
                    REPO_PATH="$1"
                elif [[ "$MODE" == "diff" ]] && [[ -z "$DIFF_COMPARE" ]]; then
                    # In diff mode, second positional arg is compare target
                    DIFF_COMPARE="$REPO_PATH"
                    REPO_PATH="$1"
                fi
                shift
                ;;
        esac
    done

    if [[ -z "$REPO_PATH" ]]; then
        echo "Error: Repository path required" >&2
        usage
    fi

    if [[ ! -d "$REPO_PATH" ]] && [[ ! -f "$REPO_PATH" ]]; then
        echo "Error: Path not found: $REPO_PATH" >&2
        exit 1
    fi

    # Apply profile settings
    case "$PROFILE" in
        quick)
            MIN_RISK="critical"
            ;;
        security)
            MIN_RISK="low"
            ;;
    esac
}

# Log message to stderr if verbose
log() {
    if [[ "$VERBOSE" == true ]]; then
        echo -e "${BLUE}[malcontent]${NC} $1" >&2
    fi
}

# Get malcontent version and rule info
get_malcontent_info() {
    local version=$("$MAL_BIN" --version 2>/dev/null | head -1 | grep -oE '[0-9]+\.[0-9]+(\.[0-9]+)?' || echo "unknown")
    echo "$version"
}

# Print detailed findings to terminal
# Only shows critical and high findings to reduce noise
# Limits to first 10 files with option to see more
print_findings() {
    local findings_json="$1"
    local repo_name="$2"
    local target_path="$3"

    # Get counts by risk level
    local total_count=$(echo "$findings_json" | jq 'length' 2>/dev/null || echo "0")
    local critical_count=$(echo "$findings_json" | jq '[.[] | select(.risk == "critical")] | length' 2>/dev/null || echo "0")
    local high_count=$(echo "$findings_json" | jq '[.[] | select(.risk == "high")] | length' 2>/dev/null || echo "0")
    local medium_count=$(echo "$findings_json" | jq '[.[] | select(.risk == "medium")] | length' 2>/dev/null || echo "0")

    # Only show findings if there are critical or high
    local important_count=$((critical_count + high_count))
    if [[ "$important_count" -eq 0 ]]; then
        if [[ "$total_count" -gt 0 ]]; then
            echo -e "\n${GREEN}✓${NC} ${total_count} files scanned, ${medium_count} medium-risk findings (no critical/high)\n" >&2
        else
            echo -e "\n${GREEN}✓ No suspicious findings detected${NC}\n" >&2
        fi
        return
    fi

    # Header
    echo "" >&2
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}" >&2
    if [[ -n "$repo_name" ]]; then
        echo -e "${CYAN}MALCONTENT FINDINGS: ${NC}${repo_name}" >&2
    else
        echo -e "${CYAN}MALCONTENT FINDINGS${NC}" >&2
    fi
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}" >&2

    # Filter to only critical and high, limit to 10 files
    local shown=0
    local max_files=10
    echo "$findings_json" | jq -r '
        [.[] | select(.risk == "critical" or .risk == "high")] |
        sort_by(
            if .risk == "critical" then 0
            else 1 end
        ) | .[0:10] | .[] | @json
    ' 2>/dev/null | while IFS= read -r finding; do
        local file_path=$(echo "$finding" | jq -r '.path // "unknown"')
        local risk=$(echo "$finding" | jq -r '.risk // "unknown"')
        local rules_matched=$(echo "$finding" | jq -r '.rules_matched // 0')

        # Make path relative to target
        local rel_path="${file_path#$target_path/}"
        [[ "$rel_path" == "$file_path" ]] && rel_path="${file_path##*/}"
        # Also strip /private prefix on macOS
        rel_path="${rel_path#/private}"
        rel_path="${rel_path#$target_path/}"

        # Risk color
        local risk_color="$NC"
        case "$risk" in
            critical) risk_color="$RED" ;;
            high)     risk_color="$YELLOW" ;;
        esac

        echo "" >&2
        echo -e "${risk_color}[$risk]${NC} ${rel_path}" >&2

        # Print top 3 behaviors with code snippets
        echo "$finding" | jq -r '
            .behaviors // [] |
            [.[] | select(.risk_level == "critical" or .risk_level == "high")] |
            .[0:3] | .[] |
            "  → \(.description // .id)\n    \(.matches // [] | .[0:2] | map("\"" + . + "\"") | join(", "))"
        ' 2>/dev/null | while IFS= read -r line; do
            if [[ "$line" =~ ^"  →" ]]; then
                echo -e "  ${DIM}→${NC}${line#  →}" >&2
            else
                echo -e "    ${DIM}${line#    }${NC}" >&2
            fi
        done
    done

    # Check if there are more findings
    if [[ "$important_count" -gt "$max_files" ]]; then
        echo "" >&2
        echo -e "${DIM}  ... and $((important_count - max_files)) more critical/high findings${NC}" >&2
    fi

    # Summary footer
    echo "" >&2
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}" >&2
    echo -e "Summary: ${total_count} files scanned" >&2
    [[ "$critical_count" -gt 0 ]] && echo -e "  ${RED}● Critical: $critical_count${NC}" >&2
    [[ "$high_count" -gt 0 ]] && echo -e "  ${YELLOW}● High: $high_count${NC}" >&2
    [[ "$medium_count" -gt 0 ]] && echo -e "  ${DIM}● Medium: $medium_count${NC}" >&2
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}" >&2
    echo "" >&2
}

# Count files to be scanned
count_scannable_files() {
    local target="$1"
    # Count files that malcontent would scan (executables, scripts, etc.)
    find "$target" -type f \( \
        -name "*.py" -o -name "*.js" -o -name "*.ts" -o -name "*.rb" -o \
        -name "*.php" -o -name "*.pl" -o -name "*.sh" -o -name "*.bash" -o \
        -name "*.go" -o -name "*.rs" -o -name "*.c" -o -name "*.cpp" -o \
        -perm +111 \
    \) 2>/dev/null | wc -l | tr -d ' '
}

# Convert risk level to numeric for comparison
risk_to_num() {
    case "$1" in
        low)      echo 1 ;;
        medium)   echo 2 ;;
        high)     echo 3 ;;
        critical) echo 4 ;;
        *)        echo 0 ;;
    esac
}

# Run malcontent analyze
run_analyze() {
    local target="$1"

    # Get version and file count for logging
    local version=$(get_malcontent_info)
    local file_count=$(count_scannable_files "$target")

    log "Malcontent v${version} with 14,500+ YARA rules"
    log "Scanning $file_count files in: $target"
    log "Min risk level: $MIN_RISK"

    # Run malcontent (binary is 'mal')
    # Global flags come before the command
    local result
    result=$("$MAL_BIN" --format json --min-risk "$MIN_RISK" analyze "$target" 2>/dev/null) || result='{"Files": {}}'

    # Log summary of what was found
    local findings_count=$(echo "$result" | jq '(.Files // .files // {}) | keys | length' 2>/dev/null || echo "0")
    log "Analysis complete: $findings_count files with findings"

    echo "$result"
}

# Run malcontent diff
run_diff() {
    local old_path="$1"
    local new_path="$2"

    log "Running malcontent diff: $old_path -> $new_path"

    # Run malcontent (binary is 'mal')
    # Global flags come before the command
    "$MAL_BIN" --format json --min-risk "$MIN_RISK" diff "$old_path" "$new_path" 2>/dev/null || echo '{"diff": {}}'
}

# Run malcontent scan
run_scan() {
    local target="$1"

    log "Running malcontent scan on: $target"

    # Run malcontent (binary is 'mal')
    # Global flags come before the command
    "$MAL_BIN" --format json --min-risk "$MIN_RISK" scan "$target" 2>/dev/null || echo '{"files": {}}'
}

# Extract packages from SBOM and optionally scan them
run_sbom_scan() {
    local sbom_file="$1"
    local download="$2"

    if [[ ! -f "$sbom_file" ]]; then
        log "SBOM file not found: $sbom_file"
        echo '{"error": "SBOM file not found", "packages": []}'
        return
    fi

    log "Extracting packages from SBOM..."

    # Use shared library if available, otherwise fall back to basic extraction
    local packages=""
    if type package_dl_extract_from_sbom &>/dev/null; then
        packages=$(package_dl_extract_from_sbom "$sbom_file")
    else
        # Fallback: basic extraction
        packages=$(jq -r '.components[]? | "\(.name)|\(.version // "latest")|unknown|\(.purl // "")"' "$sbom_file" 2>/dev/null)
    fi

    if [[ -z "$packages" ]]; then
        log "No packages found in SBOM"
        echo '{"total_packages": 0, "packages": [], "scanned": 0}'
        return
    fi

    local package_count=$(echo "$packages" | wc -l | tr -d ' ')
    log "Found $package_count packages in SBOM"

    local scanned=0
    local scan_results=()

    # Build package list with optional scanning
    local package_list=$(echo "$packages" | while IFS='|' read -r name version ecosystem purl; do
        [[ -z "$name" ]] && continue

        # Use shared function to parse ecosystem if available
        if [[ "$ecosystem" == "unknown" ]] && type package_dl_parse_ecosystem &>/dev/null; then
            ecosystem=$(package_dl_parse_ecosystem "$purl")
        fi

        local scanned_flag="false"
        local scan_result='null'

        # Download and scan if requested
        if [[ "$download" == "true" ]] && [[ "$ecosystem" =~ ^(npm|pypi)$ ]]; then
            if type package_dl_download &>/dev/null; then
                local pkg_path
                pkg_path=$(package_dl_download "$ecosystem" "$name" "$version" 2>/dev/null)

                if [[ -n "$pkg_path" ]] && [[ -f "$pkg_path" ]]; then
                    # Scan with malcontent (binary is 'mal')
                    # Global flags come before the command
                    local mc_output
                    mc_output=$("$MAL_BIN" --format json --min-risk "$MIN_RISK" analyze "$pkg_path" 2>/dev/null || echo '{}')
                    scan_result="$mc_output"
                    scanned_flag="true"
                fi
            fi
        fi

        jq -n \
            --arg name "$name" \
            --arg version "$version" \
            --arg ecosystem "$ecosystem" \
            --arg purl "$purl" \
            --argjson scanned "$scanned_flag" \
            --argjson result "$scan_result" \
            '{
                name: $name,
                version: $version,
                ecosystem: $ecosystem,
                purl: $purl,
                scanned: $scanned,
                result: $result
            }'
    done | jq -s '.')

    # Count scanned packages
    scanned=$(echo "$package_list" | jq '[.[] | select(.scanned == true)] | length')

    jq -n \
        --argjson total "$package_count" \
        --argjson scanned "$scanned" \
        --argjson packages "$package_list" \
        --arg note "$(if [[ "$download" != "true" ]]; then echo "Use --download-packages to scan package artifacts"; else echo ""; fi)" \
        '{
            total_packages: $total,
            scanned_packages: $scanned,
            packages: $packages,
            note: (if $note != "" then $note else null end)
        } | with_entries(select(.value != null))'
}

# Process malcontent output into standardized format
process_output() {
    local raw_output="$1"
    local mode="$2"

    # Parse based on mode
    # Note: malcontent uses PascalCase keys (Files, Behaviors, RiskLevel, etc.)
    case "$mode" in
        analyze|scan)
            echo "$raw_output" | jq '
                # Extract files with findings (malcontent uses "Files" with capital F)
                (.Files // .files // {}) | to_entries | map({
                    path: .key,
                    risk: (.value.RiskLevel // .value.risk // "unknown") | ascii_downcase,
                    behaviors: [(.value.Behaviors // .value.behaviors // [])[] | {
                        id: .ID,
                        description: .Description,
                        risk_level: (.RiskLevel // "unknown") | ascii_downcase,
                        risk_score: .RiskScore,
                        matches: .MatchStrings
                    }],
                    rules_matched: (.value.Behaviors // .value.behaviors // []) | length,
                    sha256: .value.SHA256,
                    size: .value.Size
                }) |
                # Filter and sort by risk
                sort_by(
                    if .risk == "critical" then 0
                    elif .risk == "high" then 1
                    elif .risk == "medium" then 2
                    else 3 end
                )
            ' 2>/dev/null || echo '[]'
            ;;
        diff)
            echo "$raw_output" | jq '
                (.Diff // .diff // {}) | {
                    added: (.Added // .added // {}),
                    removed: (.Removed // .removed // {}),
                    modified: (.Modified // .modified // {})
                }
            ' 2>/dev/null || echo '{}'
            ;;
    esac
}

# Build summary statistics
build_summary() {
    local findings="$1"
    local mode="$2"

    if [[ "$mode" == "diff" ]]; then
        echo "$findings" | jq '{
            mode: "diff",
            added_count: (.added | keys | length),
            removed_count: (.removed | keys | length),
            modified_count: (.modified | keys | length)
        }' 2>/dev/null || echo '{}'
    else
        echo "$findings" | jq '
            {
                total_files: length,
                by_risk: (group_by(.risk) | map({key: .[0].risk, value: length}) | from_entries),
                total_rules_matched: (map(.rules_matched) | add // 0),
                behaviors: ([.[].behaviors[]] | unique | sort)
            }
        ' 2>/dev/null || echo '{}'
    fi
}

# Main
main() {
    parse_args "$@"
    check_malcontent

    local start_time=$(date +%s)

    log "Mode: $MODE"
    log "Profile: $PROFILE"
    log "Min Risk: $MIN_RISK"
    log "Target: $REPO_PATH"

    # Run appropriate mode
    local raw_output=""
    case "$MODE" in
        analyze)
            raw_output=$(run_analyze "$REPO_PATH")
            ;;
        diff)
            if [[ -z "$DIFF_COMPARE" ]]; then
                echo "Error: Diff mode requires --diff <compare_path>" >&2
                exit 1
            fi
            raw_output=$(run_diff "$DIFF_COMPARE" "$REPO_PATH")
            ;;
        scan)
            raw_output=$(run_scan "$REPO_PATH")
            ;;
        *)
            echo "Error: Unknown mode: $MODE" >&2
            exit 1
            ;;
    esac

    # Process findings
    local findings=$(process_output "$raw_output" "$MODE")

    # Build summary
    local summary=$(build_summary "$findings" "$MODE")

    # Process SBOM if provided
    local sbom_results='null'
    if [[ -n "$SBOM_FILE" ]]; then
        sbom_results=$(run_sbom_scan "$SBOM_FILE" "$DOWNLOAD_PACKAGES")
    fi

    local end_time=$(date +%s)
    local duration=$((end_time - start_time))

    log "Scan completed in ${duration}s"

    # Print detailed findings if requested
    if [[ "$SHOW_FINDINGS" == true ]]; then
        print_findings "$findings" "$REPO_NAME" "$REPO_PATH"
    fi

    # Build final output using temp files to avoid "Argument list too long" error
    # when findings are large (can exceed ARG_MAX with thousands of findings)
    local temp_findings=$(mktemp)
    local temp_summary=$(mktemp)
    local temp_sbom=$(mktemp)

    echo "$findings" > "$temp_findings"
    echo "$summary" > "$temp_summary"
    echo "$sbom_results" > "$temp_sbom"

    local output=$(jq -n \
        --arg analyzer "malcontent" \
        --arg version "1.0.0" \
        --arg timestamp "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        --arg target "$REPO_PATH" \
        --arg mode "$MODE" \
        --arg profile "$PROFILE" \
        --arg min_risk "$MIN_RISK" \
        --argjson duration "$duration" \
        --slurpfile summary "$temp_summary" \
        --slurpfile findings "$temp_findings" \
        --slurpfile sbom "$temp_sbom" \
        '{
            analyzer: $analyzer,
            version: $version,
            timestamp: $timestamp,
            target: $target,
            mode: $mode,
            profile: $profile,
            min_risk: $min_risk,
            duration_seconds: $duration,
            summary: $summary[0],
            findings: $findings[0],
            sbom_analysis: $sbom[0]
        }')

    rm -f "$temp_findings" "$temp_summary" "$temp_sbom"

    # Output
    if [[ -n "$OUTPUT_FILE" ]]; then
        echo "$output" > "$OUTPUT_FILE"
        log "Output written to: $OUTPUT_FILE"
    else
        echo "$output"
    fi
}

main "$@"
