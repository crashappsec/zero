#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Crypto Ciphers Scanner (crypto-ciphers)
#
# Detects weak, deprecated, or broken cryptographic ciphers using Semgrep.
# Patterns are sourced from RAG and include:
# - DES, 3DES, RC4, Blowfish (broken/deprecated ciphers)
# - MD5, SHA1 (broken/deprecated hashes)
# - ECB mode (insecure block cipher mode)
#
# Usage: ./crypto-ciphers.sh [options] <repo_path>
# Output: JSON with cipher findings
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
Crypto Ciphers Scanner v${VERSION}

Detects weak, deprecated, or broken cryptographic ciphers in source code.

Usage: $0 [OPTIONS] <repo_path>

OPTIONS:
    --local-path PATH       Path to repository
    --timeout SECONDS       Timeout per file (default: 60)
    --verbose               Show progress messages
    -o, --output FILE       Write JSON to file (default: stdout)
    -h, --help              Show this help

DETECTED ISSUES:
    - DES, 3DES, RC4, Blowfish cipher usage
    - MD5, SHA1 hash usage for security purposes
    - ECB mode usage (patterns visible in ciphertext)
    - Weak cipher configurations

OUTPUT:
    JSON object with:
    - summary: counts by severity and cipher type
    - findings: array with file, line, cipher, severity
    - recommendations: remediation steps

EXAMPLES:
    $0 /path/to/repo
    $0 --local-path ~/.zero/repos/myapp/repo
    $0 -o ciphers.json /path/to/repo

EOF
    exit 0
}

log() {
    if [[ "$VERBOSE" == true ]]; then
        echo -e "${BLUE}[crypto-ciphers]${NC} $1" >&2
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

# Run semgrep with cipher-focused rules
run_semgrep() {
    local repo_path="$1"
    local config_args=()

    # PRIORITY 1: Use our comprehensive RAG-generated secrets rules (includes crypto patterns)
    if [[ -f "$CUSTOM_RULES_DIR/secrets.yaml" ]]; then
        config_args+=("--config" "$CUSTOM_RULES_DIR/secrets.yaml")
        log "Using RAG-generated secrets rules (242+ patterns)"
    fi

    # PRIORITY 2: Use custom crypto rules if they exist
    if [[ -f "$CUSTOM_RULES_DIR/crypto-security.yaml" ]]; then
        config_args+=("--config" "$CUSTOM_RULES_DIR/crypto-security.yaml")
        log "Using custom rules: crypto-security.yaml"
    fi

    # PRIORITY 3: Supplement with Semgrep registry security rules
    config_args+=("--config" "p/security-audit")

    log "Running semgrep cipher scan..."

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

# Filter findings to only cipher-related issues
filter_cipher_findings() {
    local raw_json="$1"

    # Filter for cipher/hash related rules
    echo "$raw_json" | jq '
    .results | map(select(
        .check_id | test("(des|rc4|md5|sha1|ecb|blowfish|3des|arcfour|weak.*crypt|insecure.*cipher|broken.*hash)"; "i")
    ))'
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
        cipher_type: (
            if .check_id | test("des"; "i") then "DES"
            elif .check_id | test("rc4|arcfour"; "i") then "RC4"
            elif .check_id | test("md5"; "i") then "MD5"
            elif .check_id | test("sha1"; "i") then "SHA1"
            elif .check_id | test("ecb"; "i") then "ECB"
            elif .check_id | test("blowfish"; "i") then "Blowfish"
            elif .check_id | test("3des"; "i") then "3DES"
            else "weak_cipher"
            end
        ),
        cwe: (.extra.metadata.cwe // ["CWE-327"]),
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

    # Calculate risk score
    local risk_score=100
    local penalty=$((critical * 25 + high * 15 + medium * 5))
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

    local by_type=$(echo "$findings_json" | jq 'group_by(.cipher_type) | map({key: .[0].cipher_type, value: length}) | from_entries')
    local files_affected=$(echo "$findings_json" | jq '[.[].file] | unique | length')

    jq -n \
        --argjson total "$total" \
        --argjson critical "$critical" \
        --argjson high "$high" \
        --argjson medium "$medium" \
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
            medium_count: $medium,
            by_cipher_type: $by_type,
            files_affected: $files_affected
        }'
}

# Generate recommendations
generate_recommendations() {
    local summary_json="$1"
    local by_type=$(echo "$summary_json" | jq -r '.by_cipher_type')

    local recs='[]'

    # Check for specific cipher types
    if echo "$by_type" | jq -e '.DES // .["3DES"]' > /dev/null 2>&1; then
        recs=$(echo "$recs" | jq '. + ["Replace DES/3DES with AES-256-GCM or ChaCha20-Poly1305"]')
    fi

    if echo "$by_type" | jq -e '.RC4' > /dev/null 2>&1; then
        recs=$(echo "$recs" | jq '. + ["RC4 is broken - migrate to AES-GCM immediately"]')
    fi

    if echo "$by_type" | jq -e '.MD5' > /dev/null 2>&1; then
        recs=$(echo "$recs" | jq '. + ["Replace MD5 with SHA-256 or SHA-3 for integrity, or Argon2/bcrypt for passwords"]')
    fi

    if echo "$by_type" | jq -e '.SHA1' > /dev/null 2>&1; then
        recs=$(echo "$recs" | jq '. + ["SHA1 is deprecated - use SHA-256 or SHA-3 for new implementations"]')
    fi

    if echo "$by_type" | jq -e '.ECB' > /dev/null 2>&1; then
        recs=$(echo "$recs" | jq '. + ["ECB mode reveals patterns - use CBC with random IV or preferably GCM mode"]')
    fi

    # Add general recommendation
    recs=$(echo "$recs" | jq '. + ["Follow NIST guidelines for cryptographic algorithm selection"]')

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

# Filter to cipher findings
cipher_findings=$(filter_cipher_findings "$raw_output")

# Process findings
findings=$(process_findings "$cipher_findings" "$REPO_PATH")

# Build summary
summary=$(build_summary "$findings")

# Generate recommendations
recommendations=$(generate_recommendations "$summary")

end_time=$(date +%s)
duration=$((end_time - start_time))

# Build final output
output=$(jq -n \
    --arg analyzer "crypto-ciphers" \
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
            ruleset: "p/security-audit + crypto patterns"
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
