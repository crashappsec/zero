#!/bin/bash
# Package Health Analyzer - AI-Enhanced with Chain of Reasoning
# Copyright (c) 2024 Crash Override Inc
# SPDX-License-Identifier: GPL-3.0

set -euo pipefail

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
UTILS_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Check for API key
if [ -z "${ANTHROPIC_API_KEY:-}" ]; then
    echo "Error: ANTHROPIC_API_KEY environment variable not set" >&2
    echo "Please set your Anthropic API key:" >&2
    echo "  export ANTHROPIC_API_KEY='your-api-key'" >&2
    exit 1
fi

# Default values
REPO=""
ORG=""
SBOM_FILE=""
OUTPUT_FORMAT="markdown"
OUTPUT_FILE=""
VERBOSE=false
SKIP_VULN_ANALYSIS=false
SKIP_PROV_ANALYSIS=false

# Usage information
usage() {
    cat <<EOF
Package Health Analyzer - AI-Enhanced

Performs comprehensive package health analysis with AI-powered recommendations
using chain of reasoning across multiple supply chain tools.

Usage: $0 [OPTIONS]

OPTIONS:
    --repo OWNER/REPO          Analyze single repository
    --org ORGANIZATION         Analyze all repositories in organization
    --sbom FILE                Analyze existing SBOM file
    --format FORMAT            Output format: markdown (default), json
    --output FILE              Write output to file (default: stdout)
    --skip-vuln-analysis       Skip vulnerability analysis step
    --skip-prov-analysis       Skip provenance analysis step
    --verbose                  Enable verbose output
    -h, --help                 Show this help message

REQUIREMENTS:
    - ANTHROPIC_API_KEY environment variable must be set
    - Base analyzer and dependencies must be installed

EXAMPLES:
    # Analyze repository with full chain of reasoning
    $0 --repo owner/repo

    # Analyze organization
    $0 --org myorg --output org-health-report.md

    # Quick analysis (skip provenance)
    $0 --repo owner/repo --skip-prov-analysis

EOF
    exit 0
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --repo)
                REPO="$2"
                shift 2
                ;;
            --org)
                ORG="$2"
                shift 2
                ;;
            --sbom)
                SBOM_FILE="$2"
                shift 2
                ;;
            --format)
                OUTPUT_FORMAT="$2"
                shift 2
                ;;
            --output)
                OUTPUT_FILE="$2"
                shift 2
                ;;
            --skip-vuln-analysis)
                SKIP_VULN_ANALYSIS=true
                shift
                ;;
            --skip-prov-analysis)
                SKIP_PROV_ANALYSIS=true
                shift
                ;;
            --verbose)
                VERBOSE=true
                shift
                ;;
            -h|--help)
                usage
                ;;
            *)
                echo "Error: Unknown option: $1" >&2
                usage
                ;;
        esac
    done
}

# Log message
log() {
    if [ "$VERBOSE" = true ]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*" >&2
    fi
}

# Call Claude API
call_claude() {
    local prompt=$1
    local max_tokens=${2:-4096}

    log "Calling Claude API"

    local response=$(curl -s https://api.anthropic.com/v1/messages \
        -H "Content-Type: application/json" \
        -H "x-api-key: $ANTHROPIC_API_KEY" \
        -H "anthropic-version: 2023-06-01" \
        -d @- <<EOF
{
    "model": "claude-sonnet-4-20250514",
    "max_tokens": $max_tokens,
    "messages": [
        {
            "role": "user",
            "content": "$prompt"
        }
    ]
}
EOF
    )

    # Check for errors
    if echo "$response" | jq -e '.error' > /dev/null 2>&1; then
        local error_msg=$(echo "$response" | jq -r '.error.message // "Unknown error"')
        echo "Error calling Claude API: $error_msg" >&2
        return 1
    fi

    # Extract content
    echo "$response" | jq -r '.content[0].text'
}

# Step 1: Run base package health analysis
run_base_analysis() {
    local args=$1

    log "Step 1/5: Running base package health analysis"

    "$SCRIPT_DIR/package-health-analyzer.sh" $args --format json
}

# Step 2: Run vulnerability analysis
run_vulnerability_analysis() {
    local sbom_file=$1

    if [ "$SKIP_VULN_ANALYSIS" = true ]; then
        log "Step 2/5: Skipping vulnerability analysis"
        echo '{"skipped": true}'
        return
    fi

    log "Step 2/5: Running vulnerability analysis"

    if [ ! -f "$UTILS_ROOT/vulnerability-analysis/vulnerability-analyzer.sh" ]; then
        log "Warning: Vulnerability analyzer not found, skipping"
        echo '{"error": "analyzer_not_found"}'
        return
    fi

    "$UTILS_ROOT/vulnerability-analysis/vulnerability-analyzer.sh" \
        --sbom "$sbom_file" \
        --format json 2>/dev/null || echo '{"error": "analysis_failed"}'
}

# Step 3: Run provenance analysis
run_provenance_analysis() {
    local sbom_file=$1

    if [ "$SKIP_PROV_ANALYSIS" = true ]; then
        log "Step 3/5: Skipping provenance analysis"
        echo '{"skipped": true}'
        return
    fi

    log "Step 3/5: Running provenance analysis"

    if [ ! -f "$UTILS_ROOT/provenance-analysis/provenance-analyzer.sh" ]; then
        log "Warning: Provenance analyzer not found, skipping"
        echo '{"error": "analyzer_not_found"}'
        return
    fi

    "$UTILS_ROOT/provenance-analysis/provenance-analyzer.sh" \
        --sbom "$sbom_file" \
        --format json 2>/dev/null || echo '{"error": "analysis_failed"}'
}

# Step 4: Prepare context for Claude
prepare_context() {
    local base_results=$1
    local vuln_results=$2
    local prov_results=$3

    log "Step 4/5: Preparing analysis context for AI"

    # Combine all results
    jq -n \
        --argjson base "$base_results" \
        --argjson vuln "$vuln_results" \
        --argjson prov "$prov_results" \
        '{
            package_health: $base,
            vulnerabilities: $vuln,
            provenance: $prov
        }'
}

# Step 5: AI analysis with Claude
ai_analysis() {
    local context=$1

    log "Step 5/5: Performing AI-enhanced analysis"

    # Escape context for JSON
    local escaped_context=$(echo "$context" | jq -Rs .)

    # Build comprehensive prompt
    local prompt=$(cat <<EOF
# Package Health Analysis

## Context
You are analyzing package health across an organization to identify risks and operational improvements.

## Input Data

### Base Package Health Analysis
\`\`\`json
$context
\`\`\`

## Analysis Tasks

Please provide a comprehensive analysis covering:

### 1. Risk Assessment
Analyze all identified issues and provide:
- Risk ranking (Critical/High/Medium/Low) for each deprecated or low-health package
- Business impact assessment
- Blast radius (how many repos/services affected)
- Urgency rating and recommended timeline for action

### 2. Version Standardization Strategy
For packages with version inconsistencies:
- Recommended target version with justification
- Breaking changes to consider
- Migration complexity assessment (Simple/Moderate/Complex)
- Phased rollout plan
- Testing requirements

### 3. Deprecated Package Migration
For each deprecated package:
- Top 3 alternative packages with pros/cons comparison
- Feature parity analysis
- API compatibility assessment
- Migration effort estimate (hours/days)
- Sample migration guide or approach

### 4. Health Score Insights
For packages with low health scores:
- Root cause analysis of low score
- Whether to keep, replace, or accept risk with mitigation
- Monitoring recommendations
- Action items to improve score

### 5. Operational Recommendations
Provide strategic guidance:
- Patterns observed across the organization
- Policy recommendations (version pinning, approval workflows, update cadence)
- Automation opportunities
- Best practices to adopt
- Technical debt reduction strategy

## Output Format

Please structure your response as a detailed markdown report with:

1. **Executive Summary** (2-3 paragraphs highlighting key findings)
2. **Risk Rankings** (table format with package, risk level, impact, urgency)
3. **Detailed Findings** (organized by category above)
4. **Prioritized Action Plan** (with effort estimates and dependencies)
5. **Long-term Recommendations** (strategic improvements)

Use clear sections, bullet points, tables, and code examples where helpful.
Focus on actionable, specific recommendations rather than generic advice.
EOF
    )

    # Call Claude
    local ai_response=$(call_claude "$prompt" 8192)

    if [ $? -ne 0 ]; then
        echo "Error: AI analysis failed" >&2
        return 1
    fi

    echo "$ai_response"
}

# Generate final report
generate_report() {
    local base_results=$1
    local ai_analysis=$2
    local format=$3

    case $format in
        json)
            jq -n \
                --argjson base "$base_results" \
                --arg ai "$ai_analysis" \
                '{
                    base_analysis: $base,
                    ai_analysis: $ai,
                    generated_at: (now | strftime("%Y-%m-%dT%H:%M:%SZ")),
                    analyzer_type: "ai-enhanced"
                }'
            ;;
        markdown)
            cat <<EOF
# Package Health Analysis Report (AI-Enhanced)

**Generated:** $(date -u +%Y-%m-%dT%H:%M:%SZ)
**Analyzer:** AI-Enhanced with Chain of Reasoning

---

## Base Analysis Summary

$(echo "$base_results" | jq -r '
"- **Repositories Scanned:** \(.scan_metadata.repositories_scanned)
- **Total Packages:** \(.summary.total_packages)
- **Deprecated Packages:** \(.summary.deprecated_packages)
- **Low Health Packages:** \(.summary.low_health_packages)
- **Version Inconsistencies:** \(.summary.version_inconsistencies)"
')

---

## AI-Enhanced Analysis

$ai_analysis

---

## Raw Data

<details>
<summary>View Complete Base Analysis Data</summary>

\`\`\`json
$(echo "$base_results" | jq '.')
\`\`\`

</details>

---

*Report generated by Package Health Analyzer with Claude AI*
EOF
            ;;
        *)
            echo "Error: Unknown format: $format" >&2
            exit 1
            ;;
    esac
}

# Main execution with chain of reasoning
main() {
    parse_args "$@"

    # Validate input
    if [ -z "$REPO" ] && [ -z "$ORG" ] && [ -z "$SBOM_FILE" ]; then
        echo "Error: Must specify --repo, --org, or --sbom" >&2
        usage
    fi

    # Prepare arguments for base analyzer
    local base_args=""
    if [ -n "$REPO" ]; then
        base_args="--repo $REPO"
    elif [ -n "$ORG" ]; then
        base_args="--org $ORG"
    elif [ -n "$SBOM_FILE" ]; then
        base_args="--sbom $SBOM_FILE"
    fi

    # Step 1: Base analysis
    local base_results=$(run_base_analysis "$base_args")

    # Determine SBOM file for subsequent analyses
    local sbom_for_analysis=""
    if [ -n "$SBOM_FILE" ]; then
        sbom_for_analysis="$SBOM_FILE"
    else
        # Generate SBOM for repo/org
        log "Generating SBOM for vulnerability/provenance analysis"

        if ! command -v syft &> /dev/null; then
            log "Warning: syft not found, skipping downstream analyses"
            SKIP_VULN_ANALYSIS=true
            SKIP_PROV_ANALYSIS=true
        else
            local temp_sbom=$(mktemp)
            trap "rm -f $temp_sbom" EXIT

            if [ -n "$REPO" ]; then
                local temp_dir=$(mktemp -d)
                trap "rm -rf $temp_dir" EXIT
                gh repo clone "$REPO" "$temp_dir/$REPO" -- --depth 1 --quiet 2>/dev/null
                syft scan "$temp_dir/$REPO" --output cyclonedx-json="$temp_sbom" --quiet 2>/dev/null
            fi

            sbom_for_analysis="$temp_sbom"
        fi
    fi

    # Step 2: Vulnerability analysis (chain of reasoning)
    local vuln_results=$(run_vulnerability_analysis "$sbom_for_analysis")

    # Step 3: Provenance analysis (chain of reasoning)
    local prov_results=$(run_provenance_analysis "$sbom_for_analysis")

    # Step 4: Prepare context
    local context=$(prepare_context "$base_results" "$vuln_results" "$prov_results")

    # Step 5: AI analysis
    local ai_analysis=$(ai_analysis "$context")

    # Generate final report
    local report=$(generate_report "$base_results" "$ai_analysis" "$OUTPUT_FORMAT")

    # Output
    if [ -n "$OUTPUT_FILE" ]; then
        echo "$report" > "$OUTPUT_FILE"
        log "Report written to $OUTPUT_FILE"
        echo "✓ Analysis complete. Report saved to: $OUTPUT_FILE"
    else
        echo "$report"
    fi
}

# Run main function
main "$@"
