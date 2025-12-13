#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# API Security Scanner (api-security)
#
# Comprehensive API security analysis using Semgrep with RAG-generated rules.
# Covers OWASP API Security Top 10:
#   - API1: Broken Object Level Authorization (BOLA)
#   - API2: Broken Authentication
#   - API3: Broken Object Property Level Authorization
#   - API4: Unrestricted Resource Consumption
#   - API5: Broken Function Level Authorization
#   - API6: Unrestricted Access to Sensitive Business Flows
#   - API7: Server Side Request Forgery (SSRF)
#   - API8: Security Misconfiguration
#
# Categories:
#   - api-auth: Authentication/authorization flaws
#   - api-injection: SQL, NoSQL, command, LDAP injection
#   - api-data-exposure: Excessive data exposure, sensitive data leakage
#   - api-rate-limiting: Missing rate limiting, resource exhaustion
#   - api-mass-assignment: Mass assignment vulnerabilities
#   - api-ssrf: Server-side request forgery
#
# Usage: ./api-security.sh [options] <repo_path>
# Output: JSON with API security findings
#############################################################################

set -e

VERSION="1.0.0"

# Colors for terminal output (stderr only)
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SCANNERS_ROOT="$(dirname "$SCRIPT_DIR")"
UTILS_ROOT="$(dirname "$SCANNERS_ROOT")"
REPO_ROOT="$(dirname "$UTILS_ROOT")"

# Semgrep rules locations
SEMGREP_DIR="$SCANNERS_ROOT/semgrep"
CUSTOM_RULES_DIR="$SEMGREP_DIR/rules"

# Default options
OUTPUT_FILE=""
REPO_PATH=""
VERBOSE=false
TIMEOUT=120
CATEGORY=""  # Filter to specific category (optional)

usage() {
    cat << EOF
API Security Scanner v${VERSION}

Comprehensive API security analysis covering OWASP API Security Top 10.

Usage: $0 [OPTIONS] <repo_path>

OPTIONS:
    --local-path PATH       Path to repository
    --timeout SECONDS       Timeout per file (default: 120)
    --category CATEGORY     Filter to specific category (optional)
    --verbose               Show progress messages
    -o, --output FILE       Write JSON to file (default: stdout)
    -h, --help              Show this help

CATEGORIES:
    api-auth            Authentication/authorization issues
    api-injection       Injection vulnerabilities (SQL, NoSQL, command)
    api-data-exposure   Excessive data exposure
    api-rate-limiting   Missing rate limiting
    api-mass-assignment Mass assignment vulnerabilities
    api-ssrf            Server-side request forgery

OUTPUT:
    JSON object with:
    - summary: counts by severity and issue type
    - findings: array with file, line, issue_type, severity
    - recommendations: remediation steps

EXAMPLES:
    $0 /path/to/repo
    $0 --local-path ~/.zero/repos/myapp/repo
    $0 --category api-auth /path/to/repo
    $0 -o api-security.json /path/to/repo

EOF
    exit 0
}

log() {
    if [[ "$VERBOSE" == true ]]; then
        echo -e "${BLUE}[api-security]${NC} $1" >&2
    fi
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

# Run semgrep with API security rules
run_semgrep() {
    local repo_path="$1"
    local config_args=()

    # Use our comprehensive RAG-generated API security rules
    if [[ -f "$CUSTOM_RULES_DIR/api-security.yaml" ]]; then
        config_args+=("--config" "$CUSTOM_RULES_DIR/api-security.yaml")
        log "Using RAG-generated API security rules (128 patterns)"
    else
        log_warn "API security rules not found at $CUSTOM_RULES_DIR/api-security.yaml"
    fi

    # Supplement with Semgrep registry security rules
    config_args+=("--config" "p/security-audit")
    config_args+=("--config" "p/owasp-top-ten")

    log "Running semgrep API security scan..."

    # Run semgrep
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
        --exclude "test" \
        --exclude "tests" \
        --exclude "*_test.go" \
        --exclude "*_test.py" \
        "$repo_path" 2>/dev/null || echo '{"results":[],"errors":[]}'
}

# Filter findings to API security related issues
filter_api_findings() {
    local raw_json="$1"
    local category_filter="$2"

    if [[ -n "$category_filter" ]]; then
        # Filter to specific category
        echo "$raw_json" | jq --arg cat "$category_filter" '
        .results | map(select(
            .check_id | test("(api-auth|api-injection|api-data|api-rate|api-mass|api-ssrf|sql.injection|nosql|xss|command.injection|ssrf|bola|broken.auth|mass.assign|rate.limit)"; "i")
            and (.extra.metadata.category // .check_id | test($cat; "i"))
        ))'
    else
        # Return all API security related findings
        echo "$raw_json" | jq '
        .results | map(select(
            .check_id | test("(api-auth|api-injection|api-data|api-rate|api-mass|api-ssrf|sql.injection|nosql|xss|command.injection|ssrf|bola|broken.auth|mass.assign|rate.limit|jwt|auth|token|session|csrf|cors)"; "i")
        ))'
    fi
}

# Determine category from rule ID
get_category() {
    local rule_id="$1"

    if echo "$rule_id" | grep -qiE "(api-auth|jwt|token|session|auth|bola|broken.auth)"; then
        echo "api-auth"
    elif echo "$rule_id" | grep -qiE "(injection|sql|nosql|command|ldap|xpath|template)"; then
        echo "api-injection"
    elif echo "$rule_id" | grep -qiE "(data.exposure|sensitive|password|leak|debug|verbose)"; then
        echo "api-data-exposure"
    elif echo "$rule_id" | grep -qiE "(rate|limit|dos|resource|pagination|timeout)"; then
        echo "api-rate-limiting"
    elif echo "$rule_id" | grep -qiE "(mass.assign|prototype|spread|body)"; then
        echo "api-mass-assignment"
    elif echo "$rule_id" | grep -qiE "(ssrf|url|fetch|request|redirect)"; then
        echo "api-ssrf"
    else
        echo "api-security"
    fi
}

# Process findings into our output format
process_findings() {
    local findings_json="$1"
    local repo_path="$2"

    echo "$findings_json" | jq --arg repo "$repo_path" '
    [.[] | {
        rule_id: .check_id,
        severity: (if .extra.severity == "ERROR" then "critical"
                   elif .extra.severity == "WARNING" then "high"
                   else "medium" end),
        message: .extra.message,
        file: (.path | sub($repo + "/"; "")),
        line: .start.line,
        column: .start.col,
        code_snippet: (.extra.lines | .[0:200]),
        category: (.extra.metadata.category // "api-security"),
        cwe: (.extra.metadata.cwe // []),
        owasp: (.extra.metadata.owasp // []),
        confidence: (.extra.metadata.confidence // 80),
        detector: "semgrep"
    }]'
}

# Build summary statistics
build_summary() {
    local findings_json="$1"

    local total=$(echo "$findings_json" | jq 'length')
    local critical=$(echo "$findings_json" | jq '[.[] | select(.severity == "critical")] | length')
    local high=$(echo "$findings_json" | jq '[.[] | select(.severity == "high")] | length')
    local medium=$(echo "$findings_json" | jq '[.[] | select(.severity == "medium")] | length')
    local low=$(echo "$findings_json" | jq '[.[] | select(.severity == "low")] | length')

    # Calculate risk score
    local risk_score=100
    local penalty=$((critical * 25 + high * 15 + medium * 5 + low * 2))
    risk_score=$((risk_score - penalty))
    [[ $risk_score -lt 0 ]] && risk_score=0

    local risk_level="excellent"
    if [[ $risk_score -lt 40 ]]; then
        risk_level="critical"
    elif [[ $risk_score -lt 60 ]]; then
        risk_level="high"
    elif [[ $risk_score -lt 80 ]]; then
        risk_level="medium"
    elif [[ $risk_score -lt 95 ]]; then
        risk_level="low"
    fi

    local by_category=$(echo "$findings_json" | jq 'group_by(.category) | map({key: .[0].category, value: length}) | from_entries')
    local files_affected=$(echo "$findings_json" | jq '[.[].file] | unique | length')

    jq -n \
        --argjson total "$total" \
        --argjson critical "$critical" \
        --argjson high "$high" \
        --argjson medium "$medium" \
        --argjson low "$low" \
        --argjson risk_score "$risk_score" \
        --arg risk_level "$risk_level" \
        --argjson by_category "$by_category" \
        --argjson files_affected "$files_affected" \
        '{
            risk_score: $risk_score,
            risk_level: $risk_level,
            total_findings: $total,
            critical_count: $critical,
            high_count: $high,
            medium_count: $medium,
            low_count: $low,
            by_category: $by_category,
            files_affected: $files_affected
        }'
}

# Generate recommendations based on findings
generate_recommendations() {
    local summary_json="$1"
    local by_category=$(echo "$summary_json" | jq -r '.by_category')

    local recs='[]'

    # Check for specific categories
    if echo "$by_category" | jq -e '."api-auth" // 0 > 0' > /dev/null 2>&1; then
        recs=$(echo "$recs" | jq '. + ["Implement proper authentication middleware on all endpoints"]')
        recs=$(echo "$recs" | jq '. + ["Use JWT with proper validation (algorithm, expiration, signature)"]')
        recs=$(echo "$recs" | jq '. + ["Implement object-level authorization checks (BOLA prevention)"]')
    fi

    if echo "$by_category" | jq -e '."api-injection" // 0 > 0' > /dev/null 2>&1; then
        recs=$(echo "$recs" | jq '. + ["Use parameterized queries for all database operations"]')
        recs=$(echo "$recs" | jq '. + ["Implement input validation and sanitization"]')
        recs=$(echo "$recs" | jq '. + ["Use ORM query builders instead of raw queries"]')
    fi

    if echo "$by_category" | jq -e '."api-data-exposure" // 0 > 0' > /dev/null 2>&1; then
        recs=$(echo "$recs" | jq '. + ["Use DTOs/view models to control response data"]')
        recs=$(echo "$recs" | jq '. + ["Never expose internal IDs, passwords, or sensitive fields"]')
        recs=$(echo "$recs" | jq '. + ["Implement field-level authorization for sensitive data"]')
    fi

    if echo "$by_category" | jq -e '."api-rate-limiting" // 0 > 0' > /dev/null 2>&1; then
        recs=$(echo "$recs" | jq '. + ["Implement rate limiting on all endpoints (express-rate-limit, etc.)"]')
        recs=$(echo "$recs" | jq '. + ["Set request body size limits"]')
        recs=$(echo "$recs" | jq '. + ["Implement pagination for list endpoints"]')
    fi

    if echo "$by_category" | jq -e '."api-mass-assignment" // 0 > 0' > /dev/null 2>&1; then
        recs=$(echo "$recs" | jq '. + ["Use allowlists for accepted request body fields"]')
        recs=$(echo "$recs" | jq '. + ["Never pass req.body directly to database operations"]')
        recs=$(echo "$recs" | jq '. + ["Use validation schemas (Joi, Yup, Zod) for all inputs"]')
    fi

    if echo "$by_category" | jq -e '."api-ssrf" // 0 > 0' > /dev/null 2>&1; then
        recs=$(echo "$recs" | jq '. + ["Validate and allowlist URLs before fetching"]')
        recs=$(echo "$recs" | jq '. + ["Block requests to internal/private IP ranges"]')
        recs=$(echo "$recs" | jq '. + ["Disable HTTP redirect following for user-provided URLs"]')
    fi

    # Add general recommendations
    recs=$(echo "$recs" | jq '. + ["Review OWASP API Security Top 10 for comprehensive guidance"]')

    echo "$recs"
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --local-path)
            REPO_PATH="$2"
            shift 2
            ;;
        --timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        --category)
            CATEGORY="$2"
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
[[ -n "$CATEGORY" ]] && log "Filtering to category: $CATEGORY"

# Run scan
raw_output=$(run_semgrep "$REPO_PATH")

# Filter to API security findings
api_findings=$(filter_api_findings "$raw_output" "$CATEGORY")

# Process findings
findings=$(process_findings "$api_findings" "$REPO_PATH")

# Build summary
summary=$(build_summary "$findings")

# Generate recommendations
recommendations=$(generate_recommendations "$summary")

end_time=$(date +%s)
duration=$((end_time - start_time))

# Build final output
output=$(jq -n \
    --arg analyzer "api-security" \
    --arg version "$VERSION" \
    --arg timestamp "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    --arg repo "$REPO_PATH" \
    --argjson duration "$duration" \
    --argjson summary "$summary" \
    --argjson findings "$findings" \
    --argjson recommendations "$recommendations" \
    '{
        analyzer: $analyzer,
        version: $version,
        timestamp: $timestamp,
        repository: $repo,
        scanner: {
            engine: "semgrep",
            ruleset: "RAG api-security + p/owasp-top-ten + p/security-audit"
        },
        duration_seconds: $duration,
        summary: $summary,
        findings: $findings,
        recommendations: $recommendations
    }')

# Output
if [[ -n "$OUTPUT_FILE" ]]; then
    echo "$output" > "$OUTPUT_FILE"
    log "Results written to $OUTPUT_FILE"
else
    echo "$output"
fi

log "Scan completed in ${duration}s"
