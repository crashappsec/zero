#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Crypto Keys Scanner (crypto-keys)
#
# Detects hardcoded cryptographic keys and weak key generation using Semgrep.
# Patterns are sourced from RAG and include:
# - Hardcoded AES/RSA/EC private keys
# - Hardcoded IVs and salts
# - Weak key lengths (RSA < 2048)
# - JWT and HMAC secrets in code
#
# Usage: ./crypto-keys.sh [options] <repo_path>
# Output: JSON with key findings
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
TIMEOUT=60

usage() {
    cat << EOF
Crypto Keys Scanner v${VERSION}

Detects hardcoded cryptographic keys and weak key generation in source code.

Usage: $0 [OPTIONS] <repo_path>

OPTIONS:
    --local-path PATH       Path to repository
    --timeout SECONDS       Timeout per file (default: 60)
    --verbose               Show progress messages
    -o, --output FILE       Write JSON to file (default: stdout)
    -h, --help              Show this help

DETECTED ISSUES:
    - RSA/EC/DSA private keys in source
    - Hardcoded AES keys (hex and base64)
    - Hardcoded IVs and nonces
    - Weak RSA key lengths (< 2048 bits)
    - JWT signing secrets
    - HMAC keys in code

OUTPUT:
    JSON object with:
    - summary: counts by severity and key type
    - findings: array with file, line, key_type, severity
    - recommendations: remediation steps

EXAMPLES:
    $0 /path/to/repo
    $0 --local-path ~/.zero/repos/myapp/repo
    $0 -o keys.json /path/to/repo

EOF
    exit 0
}

log() {
    if [[ "$VERBOSE" == true ]]; then
        echo -e "${BLUE}[crypto-keys]${NC} $1" >&2
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

# Run semgrep with key-focused rules
run_semgrep() {
    local repo_path="$1"
    local config_args=()

    # PRIORITY 1: Use our comprehensive RAG-generated secrets rules (242+ patterns from 106 tech patterns)
    if [[ -f "$CUSTOM_RULES_DIR/secrets.yaml" ]]; then
        config_args+=("--config" "$CUSTOM_RULES_DIR/secrets.yaml")
        log "Using RAG-generated secrets rules (242+ patterns)"
    fi

    # PRIORITY 2: Use custom crypto rules if they exist
    if [[ -f "$CUSTOM_RULES_DIR/crypto-security.yaml" ]]; then
        config_args+=("--config" "$CUSTOM_RULES_DIR/crypto-security.yaml")
        log "Using custom rules: crypto-security.yaml"
    fi

    # PRIORITY 3: Supplement with Semgrep registry secrets rules
    config_args+=("--config" "p/secrets")

    log "Running semgrep key scan..."

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
        --exclude "*.env.example" \
        --exclude "*.env.sample" \
        "$repo_path" 2>/dev/null || echo '{"results":[],"errors":[]}'
}

# Filter findings to only key-related issues
filter_key_findings() {
    local raw_json="$1"

    # Filter for key/secret related rules
    echo "$raw_json" | jq '
    .results | map(select(
        .check_id | test("(private.key|secret.key|hardcoded|rsa|aes|jwt|hmac|encryption.key|crypto.key|api.key|iv|nonce|salt|pkcs|pem)"; "i")
        or (.extra.message // "" | test("(private key|secret|hardcoded|encryption)"; "i"))
    ))'
}

# Redact key values
redact_key() {
    local snippet="$1"
    # Redact potential key material
    echo "$snippet" | sed -E 's/(-----BEGIN[^-]+-----).*$/\1[REDACTED]/g' | \
        sed -E 's/([A-Za-z0-9+/=]{20})[A-Za-z0-9+/=]+/\1[REDACTED]/g' | \
        head -c 200
}

# Process findings into our output format
process_findings() {
    local findings_json="$1"
    local repo_path="$2"

    echo "$findings_json" | jq --arg repo "$repo_path" '
    [.[] | {
        rule_id: .check_id,
        severity: "critical",
        message: .extra.message,
        file: (.path | sub($repo + "/"; "")),
        line: .start.line,
        column: .start.col,
        code_snippet: (.extra.lines | .[0:100] | gsub("[A-Za-z0-9+/=]{30,}"; "[REDACTED]")),
        key_type: (
            if .check_id | test("rsa|RSA"; "i") then "RSA_PRIVATE_KEY"
            elif .check_id | test("ec|EC|elliptic"; "i") then "EC_PRIVATE_KEY"
            elif .check_id | test("dsa|DSA"; "i") then "DSA_PRIVATE_KEY"
            elif .check_id | test("private.key"; "i") then "PRIVATE_KEY"
            elif .check_id | test("jwt"; "i") then "JWT_SECRET"
            elif .check_id | test("hmac"; "i") then "HMAC_KEY"
            elif .check_id | test("aes"; "i") then "AES_KEY"
            elif .check_id | test("iv|nonce"; "i") then "IV_NONCE"
            elif .check_id | test("salt"; "i") then "SALT"
            else "CRYPTO_KEY"
            end
        ),
        cwe: ["CWE-321", "CWE-798"],
        detector: "semgrep"
    }]'
}

# Build summary statistics
build_summary() {
    local findings_json="$1"

    local total=$(echo "$findings_json" | jq 'length')
    local critical=$(echo "$findings_json" | jq '[.[] | select(.severity == "critical")] | length')
    local high=$(echo "$findings_json" | jq '[.[] | select(.severity == "high")] | length')

    # Calculate risk score - hardcoded keys are very serious
    local risk_score=100
    local penalty=$((critical * 30 + high * 20))
    risk_score=$((risk_score - penalty))
    [[ $risk_score -lt 0 ]] && risk_score=0

    local risk_level="excellent"
    if [[ $risk_score -lt 30 ]]; then
        risk_level="critical"
    elif [[ $risk_score -lt 50 ]]; then
        risk_level="high"
    elif [[ $risk_score -lt 70 ]]; then
        risk_level="medium"
    elif [[ $risk_score -lt 90 ]]; then
        risk_level="low"
    fi

    local by_type=$(echo "$findings_json" | jq 'group_by(.key_type) | map({key: .[0].key_type, value: length}) | from_entries')
    local files_affected=$(echo "$findings_json" | jq '[.[].file] | unique | length')

    jq -n \
        --argjson total "$total" \
        --argjson critical "$critical" \
        --argjson high "$high" \
        --argjson risk_score "$risk_score" \
        --arg risk_level "$risk_level" \
        --argjson by_type "$by_type" \
        --argjson files_affected "$files_affected" \
        '{
            risk_score: $risk_score,
            risk_level: $risk_level,
            total_findings: $total,
            critical_count: $critical,
            high_count: $high,
            by_key_type: $by_type,
            files_affected: $files_affected
        }'
}

# Generate recommendations
generate_recommendations() {
    local summary_json="$1"
    local total=$(echo "$summary_json" | jq -r '.total_findings')
    local by_type=$(echo "$summary_json" | jq -r '.by_key_type')

    local recs='[]'

    if [[ "$total" -gt 0 ]]; then
        recs=$(echo "$recs" | jq '. + ["URGENT: Rotate all exposed keys immediately - they may be compromised"]')
    fi

    # Check for specific key types
    if echo "$by_type" | jq -e '.RSA_PRIVATE_KEY // .EC_PRIVATE_KEY // .DSA_PRIVATE_KEY // .PRIVATE_KEY' > /dev/null 2>&1; then
        recs=$(echo "$recs" | jq '. + ["Move private keys to secure key storage (HSM, KMS, or secrets manager)"]')
        recs=$(echo "$recs" | jq '. + ["Use git-secrets or pre-commit hooks to prevent future commits of keys"]')
    fi

    if echo "$by_type" | jq -e '.JWT_SECRET // .HMAC_KEY' > /dev/null 2>&1; then
        recs=$(echo "$recs" | jq '. + ["Store JWT/HMAC secrets in environment variables or secrets manager"]')
    fi

    if echo "$by_type" | jq -e '.AES_KEY' > /dev/null 2>&1; then
        recs=$(echo "$recs" | jq '. + ["Use key derivation (PBKDF2, Argon2) instead of hardcoded symmetric keys"]')
    fi

    if echo "$by_type" | jq -e '.IV_NONCE' > /dev/null 2>&1; then
        recs=$(echo "$recs" | jq '. + ["Generate random IVs/nonces per encryption - never hardcode them"]')
    fi

    if echo "$by_type" | jq -e '.SALT' > /dev/null 2>&1; then
        recs=$(echo "$recs" | jq '. + ["Generate unique salts per user/password - never reuse or hardcode"]')
    fi

    # Add general recommendations
    if [[ "$total" -eq 0 ]]; then
        recs='["No hardcoded keys detected - continue using secure key management practices"]'
    else
        recs=$(echo "$recs" | jq '. + ["Audit git history for previously committed keys using git-secrets or trufflehog"]')
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

# Run scan
raw_output=$(run_semgrep "$REPO_PATH")

# Filter to key findings
key_findings=$(filter_key_findings "$raw_output")

# Process findings
findings=$(process_findings "$key_findings" "$REPO_PATH")

# Build summary
summary=$(build_summary "$findings")

# Generate recommendations
recommendations=$(generate_recommendations "$summary")

end_time=$(date +%s)
duration=$((end_time - start_time))

# Build final output
output=$(jq -n \
    --arg analyzer "crypto-keys" \
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
            ruleset: "p/secrets + crypto patterns"
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
