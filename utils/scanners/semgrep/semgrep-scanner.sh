#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Unified Semgrep Scanner
#
# A single-pass scanner that combines:
# - Technology detection (imports, configs)
# - Secrets scanning
# - Tech debt markers (TODO, FIXME, etc.)
#
# Outputs JSON with git enrichment (who, when, commit message)
#
# Usage: ./semgrep-scanner.sh [options] <repo_path>
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
REPO_ROOT="$(dirname "$(dirname "$(dirname "$SCRIPT_DIR")")")"
RULES_DIR="$SCRIPT_DIR/rules"
PROFILES_DIR="$SCRIPT_DIR/profiles"
COMMUNITY_RULES_DIR="${SEMGREP_COMMUNITY_DIR:-$REPO_ROOT/rag/semgrep/community-rules}"

# Default options
OUTPUT_FILE=""
REPO_PATH=""
SCAN_TYPES="all"  # all, tech, secrets, debt
PROFILE="default"  # default, security, quick
ENRICH_GIT=true
USE_COMMUNITY=true
VERBOSE=false

usage() {
    cat << EOF
Unified Semgrep Scanner - Single-pass code analysis with git enrichment

Usage: $0 [OPTIONS] <repo_path>

OPTIONS:
    --profile PROFILE   Scan profile: default, security, quick (default: default)
    --type TYPE         Scan type: all, tech, secrets, debt (default: all)
    --no-git            Skip git enrichment (faster)
    --no-community      Skip community rules (faster, offline)
    --verbose           Show progress messages
    -o, --output FILE   Write JSON to file (default: stdout)
    -h, --help          Show this help

PROFILES:
    default             Tech discovery + basic security (balanced)
    security            Comprehensive security scan (OWASP, CWE, secrets)
    quick               Fast tech discovery only (no community rules)

OUTPUT:
    JSON object with:
    - summary: counts by technology, category
    - findings: array of matches with file, line, metadata
    - technologies: aggregated technology list
    - git_info: author, date, commit message (if --git enabled)

EXAMPLES:
    $0 /path/to/repo
    $0 --profile security /path/to/repo
    $0 --type tech --no-git /path/to/repo
    $0 --type secrets -o secrets.json /path/to/repo

EOF
    exit 0
}

# Check if semgrep is installed
check_semgrep() {
    if ! command -v semgrep &> /dev/null; then
        echo '{"error": "semgrep not installed", "install": "brew install semgrep"}' >&2
        exit 1
    fi
}

# Parse arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --profile)
                PROFILE="$2"
                shift 2
                ;;
            --type)
                SCAN_TYPES="$2"
                shift 2
                ;;
            --no-git)
                ENRICH_GIT=false
                shift
                ;;
            --no-community)
                USE_COMMUNITY=false
                shift
                ;;
            --verbose)
                VERBOSE=true
                shift
                ;;
            -o|--output)
                OUTPUT_FILE="$2"
                shift 2
                ;;
            -h|--help)
                usage
                ;;
            *)
                REPO_PATH="$1"
                shift
                ;;
        esac
    done

    if [[ -z "$REPO_PATH" ]]; then
        echo "Error: Repository path required" >&2
        usage
    fi

    if [[ ! -d "$REPO_PATH" ]]; then
        echo "Error: Directory not found: $REPO_PATH" >&2
        exit 1
    fi

    # Quick profile disables community rules by default
    if [[ "$PROFILE" == "quick" ]]; then
        USE_COMMUNITY=false
        ENRICH_GIT=false
    fi
}

# Log message to stderr if verbose
log() {
    if [[ "$VERBOSE" == true ]]; then
        echo -e "${BLUE}[semgrep]${NC} $1" >&2
    fi
}

# Get git blame info for a file:line
# Returns: author|email|timestamp|commit_hash|commit_message
get_git_blame() {
    local repo_path="$1"
    local file="$2"
    local line="$3"

    if [[ ! -d "$repo_path/.git" ]]; then
        echo "||||"
        return
    fi

    cd "$repo_path"

    # Get blame info for the specific line
    local blame_info=$(git blame -L "$line,$line" --porcelain "$file" 2>/dev/null || echo "")

    if [[ -z "$blame_info" ]]; then
        echo "||||"
        return
    fi

    local commit_hash=$(echo "$blame_info" | head -1 | awk '{print $1}')
    local author=$(echo "$blame_info" | grep "^author " | sed 's/^author //')
    local author_email=$(echo "$blame_info" | grep "^author-mail " | sed 's/^author-mail //' | tr -d '<>')
    local author_time=$(echo "$blame_info" | grep "^author-time " | awk '{print $2}')

    # Get commit message
    local commit_message=$(git log -1 --format='%s' "$commit_hash" 2>/dev/null | head -1 | cut -c1-100)

    # Convert timestamp to ISO format
    local iso_date=""
    if [[ -n "$author_time" ]]; then
        iso_date=$(date -r "$author_time" -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "")
    fi

    echo "$author|$author_email|$iso_date|$commit_hash|$commit_message"
}

# Get community rule packs for profile
get_community_packs() {
    local profile="$1"

    case "$profile" in
        security)
            echo "p/security-audit p/owasp-top-ten p/cwe-top-25 p/secrets p/supply-chain p/command-injection p/sql-injection p/xss p/insecure-transport"
            ;;
        default)
            echo "p/secrets p/security-audit"
            ;;
        quick|*)
            echo ""  # No community rules for quick profile
            ;;
    esac
}

# Run semgrep scan
run_semgrep_scan() {
    local repo_path="$1"
    local scan_type="$2"
    local config_args=()

    # Add custom rules based on scan type
    case "$scan_type" in
        tech)
            config_args+=("--config" "$RULES_DIR/tech-discovery.yaml")
            ;;
        secrets)
            config_args+=("--config" "$RULES_DIR/secrets.yaml")
            ;;
        debt)
            config_args+=("--config" "$RULES_DIR/tech-debt.yaml")
            ;;
        all)
            config_args+=("--config" "$RULES_DIR")
            ;;
    esac

    # Verify custom rules exist
    if [[ ! -d "$RULES_DIR" ]] && [[ ! -f "${config_args[1]}" ]]; then
        log "Warning: Custom rules not found at $RULES_DIR"
    fi

    # Add community rules if enabled
    if [[ "$USE_COMMUNITY" == true ]]; then
        local community_packs=$(get_community_packs "$PROFILE")

        if [[ -n "$community_packs" ]]; then
            # Check for locally cached rules first
            local local_rules_dir="$COMMUNITY_RULES_DIR/$PROFILE"
            if [[ -d "$local_rules_dir" ]] && [[ $(find "$local_rules_dir" -name "*.yaml" -type f 2>/dev/null | wc -l) -gt 0 ]]; then
                log "Using cached community rules from $local_rules_dir"
                config_args+=("--config" "$local_rules_dir")
            else
                # Fall back to registry packs (requires network)
                log "Using Semgrep registry rules: $community_packs"
                for pack in $community_packs; do
                    config_args+=("--config" "$pack")
                done
            fi
        fi
    fi

    log "Running semgrep scan (profile: $PROFILE, type: $scan_type)..."
    log "Config: ${config_args[*]}"

    # Run semgrep with JSON output, no metrics
    # Add timeout and memory limits for large repos
    semgrep "${config_args[@]}" \
        --json \
        --metrics=off \
        --timeout 30 \
        --max-memory 4096 \
        "$repo_path" 2>/dev/null
}

# Enrich findings with git blame
enrich_with_git() {
    local repo_path="$1"
    local findings_json="$2"

    log "Enriching findings with git data..."

    # Create temp file for collecting enriched findings
    local temp_file=$(mktemp)

    # Process each finding and add git info
    echo "$findings_json" | jq -c '.results[]' | while read -r finding; do
        local file=$(echo "$finding" | jq -r '.path')
        local line=$(echo "$finding" | jq -r '.start.line')

        # Make file path relative to repo
        local rel_file="${file#$repo_path/}"

        # Get git blame
        local blame=$(get_git_blame "$repo_path" "$rel_file" "$line")
        IFS='|' read -r author email timestamp commit_hash commit_msg <<< "$blame"

        # Add git info to finding
        echo "$finding" | jq \
            --arg author "$author" \
            --arg email "$email" \
            --arg timestamp "$timestamp" \
            --arg commit "$commit_hash" \
            --arg message "$commit_msg" \
            '. + {git: {author: $author, email: $email, timestamp: $timestamp, commit: $commit, message: $message}}'
    done > "$temp_file"

    # Convert newline-separated JSON to array
    jq -s '.' "$temp_file"
    rm -f "$temp_file"
}

# Aggregate findings by technology
aggregate_technologies() {
    local findings="$1"

    echo "$findings" | jq '
        [.[] | select(.extra.metadata.detection_type == "import")] |
        group_by(.extra.metadata.technology) |
        map({
            name: .[0].extra.metadata.technology,
            category: .[0].extra.metadata.category,
            confidence: .[0].extra.metadata.confidence,
            file_count: (map(.path) | unique | length),
            files: (map(.path) | unique),
            first_seen: (if .[0].git.timestamp then ([.[].git.timestamp | select(. != "")] | sort | first) else null end)
        }) |
        sort_by(-.file_count)
    '
}

# Build summary statistics
build_summary() {
    local findings="$1"
    local scan_type="$2"

    local total=$(echo "$findings" | jq 'length')

    local by_category=$(echo "$findings" | jq '
        group_by(.extra.metadata.category) |
        map({key: .[0].extra.metadata.category, value: length}) |
        from_entries
    ')

    local by_technology=$(echo "$findings" | jq '
        group_by(.extra.metadata.technology) |
        map({key: (.[0].extra.metadata.technology // "unknown"), value: length}) |
        from_entries
    ')

    local by_severity=$(echo "$findings" | jq '
        group_by(.extra.severity) |
        map({key: .[0].extra.severity, value: length}) |
        from_entries
    ')

    jq -n \
        --argjson total "$total" \
        --argjson by_category "$by_category" \
        --argjson by_technology "$by_technology" \
        --argjson by_severity "$by_severity" \
        --arg scan_type "$scan_type" \
        '{
            total_findings: $total,
            scan_type: $scan_type,
            by_category: $by_category,
            by_technology: $by_technology,
            by_severity: $by_severity
        }'
}

# Main
main() {
    parse_args "$@"
    check_semgrep

    local start_time=$(date +%s)

    log "Scanning: $REPO_PATH"
    log "Scan type: $SCAN_TYPES"
    log "Git enrichment: $ENRICH_GIT"

    # Run semgrep
    local raw_output=$(run_semgrep_scan "$REPO_PATH" "$SCAN_TYPES")

    # Check for errors
    if echo "$raw_output" | jq -e '.errors | length > 0' &>/dev/null; then
        local error_count=$(echo "$raw_output" | jq '.errors | length')
        log "Warning: $error_count scan errors"
    fi

    # Extract findings
    local findings=$(echo "$raw_output" | jq '.results')
    local findings_count=$(echo "$findings" | jq 'length')

    log "Found $findings_count findings"

    # Enrich with git if enabled
    local enriched_findings="$findings"
    if [[ "$ENRICH_GIT" == true ]] && [[ -d "$REPO_PATH/.git" ]]; then
        # Only enrich first 100 findings to avoid slowdown
        local sample_size=100
        if [[ $findings_count -gt $sample_size ]]; then
            log "Enriching sample of $sample_size findings (of $findings_count total)..."
            enriched_findings=$(echo "$findings" | jq ".[:$sample_size]")
            enriched_findings=$(enrich_with_git "$REPO_PATH" "{\"results\": $enriched_findings}")

            # Merge back with remaining findings (without git info)
            local remaining=$(echo "$findings" | jq ".[$sample_size:]")
            enriched_findings=$(echo "$enriched_findings $remaining" | jq -s 'add')
        else
            enriched_findings=$(enrich_with_git "$REPO_PATH" "{\"results\": $findings}")
        fi
    else
        enriched_findings=$(echo "$findings" | jq '[.[] | . + {git: {}}]')
    fi

    # Aggregate technologies
    local technologies=$(aggregate_technologies "$enriched_findings")

    # Build summary
    local summary=$(build_summary "$enriched_findings" "$SCAN_TYPES")

    local end_time=$(date +%s)
    local duration=$((end_time - start_time))

    # Build final output
    local output=$(jq -n \
        --arg analyzer "semgrep-scanner" \
        --arg version "1.0.0" \
        --arg timestamp "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        --arg repo "$REPO_PATH" \
        --argjson duration "$duration" \
        --argjson summary "$summary" \
        --argjson technologies "$technologies" \
        --argjson findings "$enriched_findings" \
        '{
            analyzer: $analyzer,
            version: $version,
            timestamp: $timestamp,
            target: $repo,
            duration_seconds: $duration,
            summary: $summary,
            technologies: $technologies,
            findings: $findings
        }')

    # Output
    if [[ -n "$OUTPUT_FILE" ]]; then
        echo "$output" > "$OUTPUT_FILE"
        log "Output written to: $OUTPUT_FILE"
    else
        echo "$output"
    fi

    log "Scan completed in ${duration}s"
}

main "$@"
