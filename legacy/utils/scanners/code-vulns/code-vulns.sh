#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Code Vulnerabilities Scanner (code-vulns)
#
# Static analysis for security vulnerabilities using Semgrep.
# Uses official Semgrep rulesets: OWASP Top 10, CWE Top 25, and more.
#
# This replaces the old code-security scanner that used raw grep patterns.
#
# Usage: ./code-vulns.sh [options] <repo_path>
# Output: JSON with vulnerability findings
#############################################################################

set -e

VERSION="2.0.0"

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

# Semgrep rules locations
SEMGREP_DIR="$SCANNERS_ROOT/semgrep"
CUSTOM_RULES_DIR="$SEMGREP_DIR/rules"
COMMUNITY_RULES_DIR="${SEMGREP_COMMUNITY_DIR:-$REPO_ROOT/rag/semgrep/community-rules}"

# Default options
OUTPUT_FILE=""
REPO_PATH=""
PROFILE="standard"  # quick, standard, comprehensive
USE_COMMUNITY=true
VERBOSE=false
MAX_FILES=2000
TIMEOUT=60

usage() {
    cat << EOF
Code Vulnerabilities Scanner v${VERSION}

Scans source code for security vulnerabilities using Semgrep with official
OWASP Top 10, CWE Top 25, and other security-focused rulesets.

Usage: $0 [OPTIONS] <repo_path>

OPTIONS:
    --local-path PATH       Path to repository (alternative to positional arg)
    --profile PROFILE       Scan profile: quick, standard, comprehensive
                           (default: standard)
    --no-community          Skip community rules (faster, offline)
    --timeout SECONDS       Timeout per file (default: 60)
    --verbose               Show progress messages
    -o, --output FILE       Write JSON to file (default: stdout)
    -h, --help              Show this help

PROFILES:
    quick                   Fast scan: p/security-audit only (~30s)
    standard                Balanced: security-audit + OWASP + secrets (~2min)
    comprehensive           Full scan: all security packs + language-specific (~5min)

OUTPUT:
    JSON object with:
    - summary: counts by severity, category, CWE
    - findings: array with file, line, rule, severity, message, fix suggestion
    - recommendations: prioritized remediation steps

EXAMPLES:
    $0 /path/to/repo
    $0 --profile comprehensive /path/to/repo
    $0 --profile quick --no-community /path/to/repo
    $0 -o vulns.json --local-path ~/.zero/repos/express/repo

ENVIRONMENT:
    SEMGREP_COMMUNITY_DIR   Override community rules cache location

EOF
    exit 0
}

log() {
    if [[ "$VERBOSE" == true ]]; then
        echo -e "${BLUE}[code-vulns]${NC} $1" >&2
    fi
}

log_success() {
    echo -e "${GREEN}✓${NC} $1" >&2
}

log_warn() {
    echo -e "${YELLOW}⚠${NC} $1" >&2
}

log_error() {
    echo -e "${RED}✗${NC} $1" >&2
}

# Check if semgrep is installed
check_semgrep() {
    if ! command -v semgrep &> /dev/null; then
        echo '{"error": "semgrep not installed", "install": "brew install semgrep"}'
        exit 1
    fi
}

# Get rule packs for profile
get_rule_packs() {
    local profile="$1"

    case "$profile" in
        quick)
            echo "p/security-audit"
            ;;
        standard)
            echo "p/security-audit p/owasp-top-ten p/secrets"
            ;;
        comprehensive)
            echo "p/security-audit p/owasp-top-ten p/cwe-top-25 p/secrets p/supply-chain p/command-injection p/sql-injection p/xss p/insecure-transport p/jwt"
            ;;
        *)
            echo "p/security-audit p/owasp-top-ten p/secrets"
            ;;
    esac
}

# Run semgrep with specified rules
run_semgrep() {
    local repo_path="$1"
    local profile="$2"
    local config_args=()

    # Add custom rules if they exist
    if [[ -f "$CUSTOM_RULES_DIR/code-vulns.yaml" ]]; then
        config_args+=("--config" "$CUSTOM_RULES_DIR/code-vulns.yaml")
        log "Using custom rules: code-vulns.yaml"
    fi

    # Add community/registry rules
    if [[ "$USE_COMMUNITY" == true ]]; then
        local packs=$(get_rule_packs "$profile")

        # Check for locally cached rules first
        local local_rules_dir="$COMMUNITY_RULES_DIR/security"
        if [[ -d "$local_rules_dir" ]] && [[ $(find "$local_rules_dir" -name "*.yaml" -type f 2>/dev/null | wc -l) -gt 0 ]]; then
            log "Using cached community rules"
            config_args+=("--config" "$local_rules_dir")
        else
            # Use registry packs directly
            log "Using Semgrep registry rules: $packs"
            for pack in $packs; do
                config_args+=("--config" "$pack")
            done
        fi
    fi

    # If no configs, use default security audit
    if [[ ${#config_args[@]} -eq 0 ]]; then
        config_args+=("--config" "p/security-audit")
    fi

    log "Running semgrep scan (profile: $profile)..."

    # Run semgrep with JSON output
    semgrep "${config_args[@]}" \
        --json \
        --metrics=off \
        --timeout "$TIMEOUT" \
        --max-memory 4096 \
        --exclude "node_modules" \
        --exclude "vendor" \
        --exclude ".git" \
        --exclude "dist" \
        --exclude "build" \
        --exclude "*.min.js" \
        --exclude "*.bundle.js" \
        "$repo_path" 2>/dev/null || echo '{"results":[],"errors":[]}'
}

# Process findings into our output format
process_findings() {
    local raw_json="$1"
    local repo_path="$2"

    echo "$raw_json" | jq --arg repo "$repo_path" '
    {
        findings: [.results[] | {
            rule_id: .check_id,
            severity: .extra.severity,
            message: .extra.message,
            file: (.path | sub($repo + "/"; "")),
            line: .start.line,
            column: .start.col,
            end_line: .end.line,
            code_snippet: .extra.lines,
            category: (.extra.metadata.category // "security"),
            cwe: (.extra.metadata.cwe // []),
            owasp: (.extra.metadata.owasp // []),
            confidence: (.extra.metadata.confidence // "MEDIUM"),
            references: (.extra.metadata.references // []),
            fix: (.extra.fix // null)
        }],
        errors: [.errors[] | {
            type: .type,
            message: .message,
            path: .path
        }]
    }'
}

# Build summary statistics
build_summary() {
    local findings_json="$1"

    echo "$findings_json" | jq '
    {
        total_findings: (.findings | length),
        by_severity: (.findings | group_by(.severity) | map({key: .[0].severity, value: length}) | from_entries),
        by_category: (.findings | group_by(.category) | map({key: .[0].category, value: length}) | from_entries),
        by_cwe: ([.findings[].cwe[] | select(. != null)] | group_by(.) | map({key: .[0], value: length}) | from_entries),
        critical_count: ([.findings[] | select(.severity == "ERROR")] | length),
        high_count: ([.findings[] | select(.severity == "WARNING")] | length),
        medium_count: ([.findings[] | select(.severity == "INFO")] | length),
        files_affected: (.findings | map(.file) | unique | length),
        scan_errors: (.errors | length)
    }'
}

# Generate recommendations
generate_recommendations() {
    local summary_json="$1"

    local critical=$(echo "$summary_json" | jq -r '.critical_count')
    local high=$(echo "$summary_json" | jq -r '.high_count')
    local total=$(echo "$summary_json" | jq -r '.total_findings')

    local recs='[]'

    if [[ "$critical" -gt 0 ]]; then
        recs=$(echo "$recs" | jq '. + ["URGENT: Address critical severity findings immediately - these represent high-risk vulnerabilities"]')
    fi

    if [[ "$high" -gt 0 ]]; then
        recs=$(echo "$recs" | jq '. + ["Review and remediate high severity findings before next release"]')
    fi

    if [[ "$total" -gt 20 ]]; then
        recs=$(echo "$recs" | jq '. + ["Consider implementing automated SAST in CI/CD pipeline to catch issues earlier"]')
    fi

    # Check for common vulnerability types
    local sql_count=$(echo "$summary_json" | jq -r '.by_category["sql-injection"] // 0')
    local xss_count=$(echo "$summary_json" | jq -r '.by_category["xss"] // 0')
    local cmd_count=$(echo "$summary_json" | jq -r '.by_category["command-injection"] // 0')

    if [[ "$sql_count" -gt 0 ]]; then
        recs=$(echo "$recs" | jq '. + ["SQL Injection detected: Use parameterized queries or prepared statements"]')
    fi

    if [[ "$xss_count" -gt 0 ]]; then
        recs=$(echo "$recs" | jq '. + ["XSS vulnerabilities found: Implement proper output encoding and Content Security Policy"]')
    fi

    if [[ "$cmd_count" -gt 0 ]]; then
        recs=$(echo "$recs" | jq '. + ["Command injection risk: Avoid shell execution with user input, use safe APIs"]')
    fi

    # Add default recommendation if none generated
    if [[ $(echo "$recs" | jq 'length') -eq 0 ]]; then
        recs='["No critical issues found - continue regular security reviews"]'
    fi

    echo "$recs"
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --local-path)
            REPO_PATH="$2"
            shift 2
            ;;
        --profile)
            PROFILE="$2"
            shift 2
            ;;
        --no-community)
            USE_COMMUNITY=false
            shift
            ;;
        --timeout)
            TIMEOUT="$2"
            shift 2
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
        -*)
            log_error "Unknown option: $1"
            usage
            ;;
        *)
            REPO_PATH="$1"
            shift
            ;;
    esac
done

# Validate input
if [[ -z "$REPO_PATH" ]]; then
    log_error "Repository path required"
    usage
fi

if [[ ! -d "$REPO_PATH" ]]; then
    echo '{"error": "Repository path does not exist"}'
    exit 1
fi

# Main execution
check_semgrep

start_time=$(date +%s)
log "Scanning: $REPO_PATH"
log "Profile: $PROFILE"

# Run scan
raw_output=$(run_semgrep "$REPO_PATH" "$PROFILE")

# Process findings
processed=$(process_findings "$raw_output" "$REPO_PATH")

# Build summary
summary=$(build_summary "$processed")

# Generate recommendations
recommendations=$(generate_recommendations "$summary")

end_time=$(date +%s)
duration=$((end_time - start_time))

# Build final output
output=$(jq -n \
    --arg analyzer "code-vulns" \
    --arg version "$VERSION" \
    --arg timestamp "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    --arg repo "$REPO_PATH" \
    --arg profile "$PROFILE" \
    --argjson duration "$duration" \
    --argjson summary "$summary" \
    --argjson findings "$(echo "$processed" | jq '.findings')" \
    --argjson errors "$(echo "$processed" | jq '.errors')" \
    --argjson recommendations "$recommendations" \
    '{
        analyzer: $analyzer,
        version: $version,
        timestamp: $timestamp,
        repository: $repo,
        profile: $profile,
        duration_seconds: $duration,
        summary: $summary,
        findings: $findings,
        errors: $errors,
        recommendations: $recommendations
    }')

# Output
if [[ -n "$OUTPUT_FILE" ]]; then
    echo "$output" > "$OUTPUT_FILE"
    log_success "Results written to $OUTPUT_FILE"
else
    echo "$output"
fi

log "Scan completed in ${duration}s"
