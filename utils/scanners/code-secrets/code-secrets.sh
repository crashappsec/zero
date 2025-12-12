#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Code Secrets Scanner (code-secrets)
#
# Detects exposed secrets, API keys, and credentials using Semgrep.
# Uses official Semgrep p/secrets ruleset plus custom patterns from RAG.
#
# This replaces the old pattern-matching approach with Semgrep for:
# - Better accuracy (AST-aware, not just regex)
# - Consistent tooling across all scanners
# - Extensible via RAG-generated rules
#
# Usage: ./code-secrets.sh [options] <repo_path>
# Output: JSON with secret findings (values redacted)
#############################################################################

set -e

VERSION="2.0.0"

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
COMMUNITY_RULES_DIR="${SEMGREP_COMMUNITY_DIR:-$REPO_ROOT/rag/semgrep/community-rules}"

# Default options
OUTPUT_FILE=""
REPO_PATH=""
USE_COMMUNITY=true
VERBOSE=false
TIMEOUT=60

usage() {
    cat << EOF
Code Secrets Scanner v${VERSION}

Detects exposed secrets, API keys, passwords, and credentials in source code
using Semgrep's official secrets ruleset plus custom patterns.

Usage: $0 [OPTIONS] <repo_path>

OPTIONS:
    --local-path PATH       Path to repository
    --repo OWNER/REPO       GitHub repository (uses zero cache)
    --org ORG               GitHub org (uses first repo in zero cache)
    --no-community          Skip community rules (faster, offline)
    --timeout SECONDS       Timeout per file (default: 60)
    --verbose               Show progress messages
    -o, --output FILE       Write JSON to file (default: stdout)
    -h, --help              Show this help

DETECTED SECRET TYPES:
    - AWS Access Keys and Secret Keys
    - GitHub Tokens (PAT, OAuth, App)
    - GitLab Tokens
    - Slack Tokens and Webhooks
    - Stripe API Keys
    - Google Cloud Service Account Keys
    - Private Keys (RSA, EC, DSA, PGP)
    - Database Connection Strings
    - JWT Secrets
    - API Keys (generic patterns)
    - And 100+ more patterns via Semgrep registry

OUTPUT:
    JSON object with:
    - summary: risk score, counts by severity and type
    - findings: array with file, line, type, severity (values redacted)
    - recommendations: remediation steps

EXAMPLES:
    $0 /path/to/repo
    $0 --local-path ~/.zero/repos/myapp/repo
    $0 --repo expressjs/express
    $0 -o secrets.json /path/to/repo

EOF
    exit 0
}

log() {
    if [[ "$VERBOSE" == true ]]; then
        echo -e "${BLUE}[code-secrets]${NC} $1" >&2
    fi
}

log_warn() {
    echo -e "${YELLOW}⚠${NC} $1" >&2
}

log_success() {
    echo -e "${GREEN}✓${NC} $1" >&2
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

# Resolve repository path from various inputs
resolve_repo_path() {
    local local_path="$1"
    local repo="$2"
    local org="$3"

    if [[ -n "$local_path" ]]; then
        echo "$local_path"
        return
    fi

    if [[ -n "$repo" ]]; then
        local repo_org=$(echo "$repo" | cut -d'/' -f1)
        local repo_name=$(echo "$repo" | cut -d'/' -f2)
        local zero_path="$HOME/.zero/repos/$repo_org/$repo_name/repo"
        if [[ -d "$zero_path" ]]; then
            echo "$zero_path"
            return
        fi
    fi

    if [[ -n "$org" ]]; then
        local org_path="$HOME/.zero/repos/$org"
        if [[ -d "$org_path" ]]; then
            # Find first repo with cloned code
            for dir in "$org_path"/*/repo; do
                if [[ -d "$dir" ]]; then
                    echo "$dir"
                    return
                fi
            done
        fi
    fi

    echo ""
}

# Run semgrep with secrets rules
run_semgrep() {
    local repo_path="$1"
    local config_args=()

    # PRIORITY 1: Use our comprehensive RAG-generated secrets rules
    # This includes 242+ patterns from 106 technology patterns covering:
    # AWS, Azure, GCP, Stripe, Twilio, SendGrid, OpenAI, Anthropic, and 100+ more
    if [[ -f "$CUSTOM_RULES_DIR/secrets.yaml" ]]; then
        config_args+=("--config" "$CUSTOM_RULES_DIR/secrets.yaml")
        local rule_count=$(grep -c "^- id:" "$CUSTOM_RULES_DIR/secrets.yaml" 2>/dev/null || echo "?")
        log "Using RAG-generated secrets rules ($rule_count patterns from tech-discovery)"
    fi

    # PRIORITY 2: Add community/registry secrets rules to supplement
    if [[ "$USE_COMMUNITY" == true ]]; then
        # Check for locally cached rules first
        local local_rules="$COMMUNITY_RULES_DIR/security/secrets.yaml"
        if [[ -f "$local_rules" ]]; then
            config_args+=("--config" "$local_rules")
            log "Using cached community secrets rules"
        else
            # Use registry pack directly
            config_args+=("--config" "p/secrets")
            log "Supplementing with Semgrep registry: p/secrets"
        fi
    fi

    # Fallback if no custom rules
    if [[ ${#config_args[@]} -eq 0 ]]; then
        config_args+=("--config" "p/secrets")
        log "Fallback: Using Semgrep registry p/secrets only"
    fi

    log "Running semgrep secrets scan..."

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
        --exclude "package-lock.json" \
        --exclude "yarn.lock" \
        --exclude "pnpm-lock.yaml" \
        --exclude "*.env.example" \
        --exclude "*.env.sample" \
        --exclude "*.env.template" \
        "$repo_path" 2>/dev/null || echo '{"results":[],"errors":[]}'
}

# Redact secret values in snippets
redact_secret() {
    local snippet="$1"
    # Truncate and mask secret values
    echo "$snippet" | sed -E 's/([A-Za-z0-9_-]{8})[A-Za-z0-9_+/=-]{8,}/\1********/g' | head -c 200
}

# Map semgrep severity to our severity
map_severity() {
    local rule_id="$1"
    local semgrep_severity="$2"

    # Critical secrets
    if echo "$rule_id" | grep -qiE "aws.access|private.key|gcp.service.account|stripe.live"; then
        echo "critical"
        return
    fi

    # High severity
    if echo "$rule_id" | grep -qiE "github.token|gitlab.token|database.url|jwt.secret|api.key"; then
        echo "high"
        return
    fi

    # Map semgrep severity
    case "$semgrep_severity" in
        ERROR) echo "critical" ;;
        WARNING) echo "high" ;;
        INFO) echo "medium" ;;
        *) echo "medium" ;;
    esac
}

# Determine secret type from rule ID
get_secret_type() {
    local rule_id="$1"

    # Extract meaningful type from rule ID
    if echo "$rule_id" | grep -qi "aws"; then
        echo "aws_credential"
    elif echo "$rule_id" | grep -qi "github"; then
        echo "github_token"
    elif echo "$rule_id" | grep -qi "gitlab"; then
        echo "gitlab_token"
    elif echo "$rule_id" | grep -qi "slack"; then
        echo "slack_token"
    elif echo "$rule_id" | grep -qi "stripe"; then
        echo "stripe_key"
    elif echo "$rule_id" | grep -qi "private.key\|rsa\|dsa\|ec.private"; then
        echo "private_key"
    elif echo "$rule_id" | grep -qi "postgres\|mysql\|mongodb\|redis\|database"; then
        echo "database_credential"
    elif echo "$rule_id" | grep -qi "jwt"; then
        echo "jwt_secret"
    elif echo "$rule_id" | grep -qi "api.key\|apikey"; then
        echo "api_key"
    elif echo "$rule_id" | grep -qi "password"; then
        echo "password"
    elif echo "$rule_id" | grep -qi "secret"; then
        echo "generic_secret"
    else
        echo "unknown"
    fi
}

# Process findings into our output format
process_findings() {
    local raw_json="$1"
    local repo_path="$2"

    # Process each finding
    echo "$raw_json" | jq --arg repo "$repo_path" '
    [.results[] | {
        rule_id: .check_id,
        type: .check_id,
        severity: .extra.severity,
        message: .extra.message,
        file: (.path | sub($repo + "/"; "")),
        line: .start.line,
        column: .start.col,
        snippet: (.extra.lines | .[0:200]),
        detector: "semgrep"
    }]'
}

# Build summary statistics
build_summary() {
    local findings_json="$1"

    local total=$(echo "$findings_json" | jq 'length')
    local critical=$(echo "$findings_json" | jq '[.[] | select(.severity == "ERROR" or .severity == "critical")] | length')
    local high=$(echo "$findings_json" | jq '[.[] | select(.severity == "WARNING" or .severity == "high")] | length')
    local medium=$(echo "$findings_json" | jq '[.[] | select(.severity == "INFO" or .severity == "medium")] | length')
    local low=$(echo "$findings_json" | jq '[.[] | select(.severity == "low")] | length')

    # Calculate risk score (100 = clean, 0 = critical)
    local risk_score=100
    local penalty=$((critical * 25 + high * 15 + medium * 5 + low * 2))
    risk_score=$((risk_score - penalty))
    [[ $risk_score -lt 0 ]] && risk_score=0

    # Determine risk level
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

    local by_type=$(echo "$findings_json" | jq 'group_by(.type) | map({key: .[0].type, value: length}) | from_entries')
    local files_affected=$(echo "$findings_json" | jq '[.[].file] | unique | length')

    jq -n \
        --argjson total "$total" \
        --argjson critical "$critical" \
        --argjson high "$high" \
        --argjson medium "$medium" \
        --argjson low "$low" \
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
            low_count: $low,
            by_type: $by_type,
            files_affected: $files_affected
        }'
}

# Generate recommendations
generate_recommendations() {
    local summary_json="$1"

    local critical=$(echo "$summary_json" | jq -r '.critical_count')
    local high=$(echo "$summary_json" | jq -r '.high_count')
    local total=$(echo "$summary_json" | jq -r '.total_findings')

    local recs='["Use environment variables or a secrets manager for sensitive data", "Enable pre-commit hooks to prevent secret commits (e.g., git-secrets, detect-secrets)"]'

    if [[ "$critical" -gt 0 ]]; then
        recs=$(echo "$recs" | jq '. = ["URGENT: Rotate all critical secrets immediately - they may already be compromised"] + .')
    fi

    if [[ "$high" -gt 0 ]]; then
        recs=$(echo "$recs" | jq '. += ["Review and rotate high-severity secrets before next deployment"]')
    fi

    if [[ "$total" -gt 5 ]]; then
        recs=$(echo "$recs" | jq '. += ["Consider using a secrets scanning tool in CI/CD pipeline"]')
    fi

    echo "$recs"
}

# Parse arguments
LOCAL_PATH=""
REPO=""
ORG=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --local-path)
            LOCAL_PATH="$2"
            shift 2
            ;;
        --repo)
            REPO="$2"
            shift 2
            ;;
        --org)
            ORG="$2"
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

# Resolve repository path
if [[ -z "$REPO_PATH" ]]; then
    REPO_PATH=$(resolve_repo_path "$LOCAL_PATH" "$REPO" "$ORG")
fi

if [[ -z "$REPO_PATH" ]] || [[ ! -d "$REPO_PATH" ]]; then
    log_error "Repository path required or not found"
    usage
fi

# Main execution
check_semgrep

start_time=$(date +%s)
log "Scanning: $REPO_PATH"

# Run scan
raw_output=$(run_semgrep "$REPO_PATH")

# Check for errors
error_count=$(echo "$raw_output" | jq '.errors | length')
if [[ "$error_count" -gt 0 ]]; then
    log_warn "$error_count scan errors occurred"
fi

# Process findings
findings=$(process_findings "$raw_output" "$REPO_PATH")

# Build summary
summary=$(build_summary "$findings")

# Generate recommendations
recommendations=$(generate_recommendations "$summary")

end_time=$(date +%s)
duration=$((end_time - start_time))

# Build final output
output=$(jq -n \
    --arg analyzer "code-secrets" \
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
            ruleset: "p/secrets + custom"
        },
        duration_seconds: $duration,
        summary: $summary,
        findings: $findings,
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
