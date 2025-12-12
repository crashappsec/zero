#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Crypto Random Scanner (crypto-random)
#
# Detects insecure random number generation using Semgrep.
# Patterns are sourced from RAG and include:
# - Math.random() in JavaScript
# - random module in Python (vs secrets/os.urandom)
# - java.util.Random (vs SecureRandom)
# - math/rand in Go (vs crypto/rand)
#
# Usage: ./crypto-random.sh [options] <repo_path>
# Output: JSON with insecure random findings
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
Crypto Random Scanner v${VERSION}

Detects insecure random number generation in source code.

Usage: $0 [OPTIONS] <repo_path>

OPTIONS:
    --local-path PATH       Path to repository
    --timeout SECONDS       Timeout per file (default: 60)
    --verbose               Show progress messages
    -o, --output FILE       Write JSON to file (default: stdout)
    -h, --help              Show this help

DETECTED ISSUES:
    - JavaScript Math.random() for security purposes
    - Python random module (should use secrets/os.urandom)
    - Java java.util.Random (should use SecureRandom)
    - Go math/rand (should use crypto/rand)
    - Ruby rand/Random (should use SecureRandom)
    - Hardcoded or time-based seeds

OUTPUT:
    JSON object with:
    - summary: counts by severity and language
    - findings: array with file, line, pattern, severity
    - recommendations: remediation steps

EXAMPLES:
    $0 /path/to/repo
    $0 --local-path ~/.zero/repos/myapp/repo
    $0 -o random.json /path/to/repo

EOF
    exit 0
}

log() {
    if [[ "$VERBOSE" == true ]]; then
        echo -e "${BLUE}[crypto-random]${NC} $1" >&2
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

# Run semgrep with random-focused rules
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

    log "Running semgrep random scan..."

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

# Filter findings to only random-related issues
filter_random_findings() {
    local raw_json="$1"

    # Filter for random/PRNG related rules
    echo "$raw_json" | jq '
    .results | map(select(
        .check_id | test("(random|prng|Math\\.random|SecureRandom|insecure.*random|weak.*random|predictable)"; "i")
        or (.extra.message // "" | test("(random|prng|predictable|insecure random)"; "i"))
    ))'
}

# Detect language from file extension
detect_language() {
    local file="$1"
    case "$file" in
        *.py) echo "python" ;;
        *.js|*.jsx|*.ts|*.tsx) echo "javascript" ;;
        *.java) echo "java" ;;
        *.go) echo "go" ;;
        *.rb) echo "ruby" ;;
        *.php) echo "php" ;;
        *.c|*.cpp|*.cc|*.h|*.hpp) echo "c" ;;
        *.cs) echo "csharp" ;;
        *) echo "unknown" ;;
    esac
}

# Process findings into our output format
process_findings() {
    local findings_json="$1"
    local repo_path="$2"

    echo "$findings_json" | jq --arg repo "$repo_path" '
    [.[] | {
        rule_id: .check_id,
        severity: (if .extra.severity == "ERROR" then "high"
                   elif .extra.severity == "WARNING" then "medium"
                   else "low" end),
        message: .extra.message,
        file: (.path | sub($repo + "/"; "")),
        line: .start.line,
        column: .start.col,
        code_snippet: (.extra.lines | .[0:200]),
        language: (
            if .path | test("\\.py$") then "python"
            elif .path | test("\\.(js|jsx|ts|tsx)$") then "javascript"
            elif .path | test("\\.java$") then "java"
            elif .path | test("\\.go$") then "go"
            elif .path | test("\\.rb$") then "ruby"
            elif .path | test("\\.php$") then "php"
            elif .path | test("\\.(c|cpp|h|hpp)$") then "c"
            elif .path | test("\\.cs$") then "csharp"
            else "unknown"
            end
        ),
        cwe: ["CWE-330", "CWE-338"],
        detector: "semgrep"
    }]'
}

# Build summary statistics
build_summary() {
    local findings_json="$1"

    local total=$(echo "$findings_json" | jq 'length')
    local high=$(echo "$findings_json" | jq '[.[] | select(.severity == "high")] | length')
    local medium=$(echo "$findings_json" | jq '[.[] | select(.severity == "medium")] | length')
    local low=$(echo "$findings_json" | jq '[.[] | select(.severity == "low")] | length')

    # Calculate risk score
    local risk_score=100
    local penalty=$((high * 15 + medium * 8 + low * 3))
    risk_score=$((risk_score - penalty))
    [[ $risk_score -lt 0 ]] && risk_score=0

    local risk_level="excellent"
    if [[ $risk_score -lt 40 ]]; then
        risk_level="high"
    elif [[ $risk_score -lt 60 ]]; then
        risk_level="medium"
    elif [[ $risk_score -lt 80 ]]; then
        risk_level="low"
    fi

    local by_language=$(echo "$findings_json" | jq 'group_by(.language) | map({key: .[0].language, value: length}) | from_entries')
    local files_affected=$(echo "$findings_json" | jq '[.[].file] | unique | length')

    jq -n \
        --argjson total "$total" \
        --argjson high "$high" \
        --argjson medium "$medium" \
        --argjson low "$low" \
        --argjson risk_score "$risk_score" \
        --arg risk_level "$risk_level" \
        --argjson by_language "$by_language" \
        --argjson files_affected "$files_affected" \
        '{
            risk_score: $risk_score,
            risk_level: $risk_level,
            total_findings: $total,
            high_count: $high,
            medium_count: $medium,
            low_count: $low,
            by_language: $by_language,
            files_affected: $files_affected
        }'
}

# Generate recommendations
generate_recommendations() {
    local summary_json="$1"
    local by_language=$(echo "$summary_json" | jq -r '.by_language')

    local recs='[]'

    # Check for specific languages
    if echo "$by_language" | jq -e '.javascript' > /dev/null 2>&1; then
        recs=$(echo "$recs" | jq '. + ["JavaScript: Use crypto.randomBytes() or crypto.getRandomValues() instead of Math.random()"]')
    fi

    if echo "$by_language" | jq -e '.python' > /dev/null 2>&1; then
        recs=$(echo "$recs" | jq '. + ["Python: Use secrets module or os.urandom() instead of random module for security"]')
    fi

    if echo "$by_language" | jq -e '.java' > /dev/null 2>&1; then
        recs=$(echo "$recs" | jq '. + ["Java: Use java.security.SecureRandom instead of java.util.Random"]')
    fi

    if echo "$by_language" | jq -e '.go' > /dev/null 2>&1; then
        recs=$(echo "$recs" | jq '. + ["Go: Use crypto/rand instead of math/rand for security-sensitive operations"]')
    fi

    if echo "$by_language" | jq -e '.ruby' > /dev/null 2>&1; then
        recs=$(echo "$recs" | jq '. + ["Ruby: Use SecureRandom instead of rand() for security purposes"]')
    fi

    if echo "$by_language" | jq -e '.php' > /dev/null 2>&1; then
        recs=$(echo "$recs" | jq '. + ["PHP: Use random_bytes() or random_int() instead of rand()/mt_rand()"]')
    fi

    if echo "$by_language" | jq -e '.c' > /dev/null 2>&1; then
        recs=$(echo "$recs" | jq '. + ["C/C++: Use OS-provided CSPRNG (/dev/urandom, CryptGenRandom) instead of rand()"]')
    fi

    if echo "$by_language" | jq -e '.csharp' > /dev/null 2>&1; then
        recs=$(echo "$recs" | jq '. + ["C#: Use RNGCryptoServiceProvider or RandomNumberGenerator instead of Random"]')
    fi

    # Add general recommendation if no findings
    if [[ $(echo "$summary_json" | jq -r '.total_findings') -eq 0 ]]; then
        recs='["No insecure random usage detected - ensure all security-sensitive code uses cryptographic RNG"]'
    else
        recs=$(echo "$recs" | jq '. + ["Review each finding to determine if the random value is used for security purposes"]')
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

# Filter to random findings
random_findings=$(filter_random_findings "$raw_output")

# Process findings
findings=$(process_findings "$random_findings" "$REPO_PATH")

# Build summary
summary=$(build_summary "$findings")

# Generate recommendations
recommendations=$(generate_recommendations "$summary")

end_time=$(date +%s)
duration=$((end_time - start_time))

# Build final output
output=$(jq -n \
    --arg analyzer "crypto-random" \
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
            ruleset: "p/security-audit + random patterns"
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
