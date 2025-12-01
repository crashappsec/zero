#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# IaC Security - Data Collector
# Scans Infrastructure-as-Code files for security misconfigurations
# Uses Checkov for static analysis of Terraform, CloudFormation, Kubernetes, etc.
# Outputs JSON for agent analysis - NO AI calls
#
# Usage: ./iac-security-data.sh [options] <target>
# Output: JSON with IaC findings, framework detection, and compliance mapping
#############################################################################

set -e

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

# Default options
OUTPUT_FILE=""
LOCAL_PATH=""
TEMP_DIR=""
CLEANUP=true
TARGET=""
FRAMEWORKS=""  # Auto-detect by default
COMPACT=false

usage() {
    cat << EOF
IaC Security - Data Collector (JSON output for agent analysis)

Scans Infrastructure-as-Code files for security misconfigurations using Checkov.

Usage: $0 [OPTIONS] <target>

TARGET:
    Git repository URL      Clone and analyze repository
    Local directory path    Analyze local repository

OPTIONS:
    --local-path PATH       Use pre-cloned repository (skips cloning)
    --framework TYPE        Specific framework to scan (terraform, cloudformation,
                           kubernetes, dockerfile, helm, etc.)
    --compact               Output compact JSON (no pretty-print)
    -o, --output FILE       Write JSON to file (default: stdout)
    -k, --keep-clone        Keep cloned repository
    -h, --help              Show this help

SUPPORTED FRAMEWORKS:
    Checkov supports 50+ frameworks including:
    - terraform, terraform_plan
    - cloudformation, serverless
    - kubernetes, helm, kustomize
    - dockerfile, docker_compose
    - arm, bicep (Azure)
    - github_actions, gitlab_ci, circleci
    - secrets (embedded secrets detection)

OUTPUT:
    JSON object with:
    - metadata: scan timestamp, target, checkov version
    - frameworks_detected: IaC frameworks found in repo
    - findings: security misconfigurations by severity
    - compliance: mapping to frameworks (CIS, SOC2, HIPAA, etc.)

EXAMPLES:
    $0 https://github.com/hashicorp/terraform-provider-aws
    $0 --local-path ~/.gibson/projects/foo/repo
    $0 --framework terraform -o iac-security.json /path/to/project

REQUIREMENTS:
    - Checkov must be installed: pip install checkov
    - Or: brew install checkov

EOF
    exit 0
}

# Find checkov binary - check PATH and common locations
CHECKOV_BIN=""
find_checkov() {
    # Check PATH first
    if command -v checkov &> /dev/null; then
        CHECKOV_BIN="checkov"
        return 0
    fi

    # Check common pip install locations
    local possible_paths=(
        "$HOME/Library/Python/3.9/bin/checkov"
        "$HOME/Library/Python/3.10/bin/checkov"
        "$HOME/Library/Python/3.11/bin/checkov"
        "$HOME/Library/Python/3.12/bin/checkov"
        "$HOME/.local/bin/checkov"
        "/usr/local/bin/checkov"
        "/opt/homebrew/bin/checkov"
    )

    for path in "${possible_paths[@]}"; do
        if [[ -x "$path" ]]; then
            CHECKOV_BIN="$path"
            return 0
        fi
    done

    return 1
}

# Check if Checkov is installed
check_checkov() {
    find_checkov
}

# Clone repository
clone_repository() {
    local repo_url="$1"
    TEMP_DIR=$(mktemp -d)

    echo -e "${BLUE}Cloning repository...${NC}" >&2
    if git clone --depth 1 "$repo_url" "$TEMP_DIR" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Cloned${NC}" >&2
        return 0
    else
        echo '{"error": "Failed to clone repository"}'
        exit 1
    fi
}

# Cleanup
cleanup() {
    if [[ "$CLEANUP" == true ]] && [[ -n "$TEMP_DIR" ]] && [[ -d "$TEMP_DIR" ]]; then
        rm -rf "$TEMP_DIR"
    fi
}
trap cleanup EXIT

# Detect if target is a Git URL
is_git_url() {
    [[ "$1" =~ ^(https?|git)://.*\.git$ ]] || [[ "$1" =~ ^git@.*:.*\.git$ ]] || [[ "$1" =~ github\.com|gitlab\.com|bitbucket\.org ]]
}

# Detect IaC frameworks present in repository
detect_frameworks() {
    local repo_dir="$1"
    local frameworks="[]"

    # Terraform
    if find "$repo_dir" -name "*.tf" -type f 2>/dev/null | head -1 | grep -q .; then
        frameworks=$(echo "$frameworks" | jq '. + ["terraform"]')
    fi

    # CloudFormation
    if find "$repo_dir" \( -name "*.yaml" -o -name "*.yml" -o -name "*.json" \) -type f 2>/dev/null | \
       xargs grep -l "AWSTemplateFormatVersion\|AWS::CloudFormation" 2>/dev/null | head -1 | grep -q .; then
        frameworks=$(echo "$frameworks" | jq '. + ["cloudformation"]')
    fi

    # Kubernetes
    if find "$repo_dir" \( -name "*.yaml" -o -name "*.yml" \) -type f 2>/dev/null | \
       xargs grep -l "apiVersion:\|kind:" 2>/dev/null | head -1 | grep -q .; then
        frameworks=$(echo "$frameworks" | jq '. + ["kubernetes"]')
    fi

    # Dockerfile
    if find "$repo_dir" \( -name "Dockerfile" -o -name "Dockerfile.*" -o -name "*.dockerfile" \) -type f 2>/dev/null | head -1 | grep -q .; then
        frameworks=$(echo "$frameworks" | jq '. + ["dockerfile"]')
    fi

    # Docker Compose
    if find "$repo_dir" \( -name "docker-compose.yml" -o -name "docker-compose.yaml" -o -name "compose.yml" -o -name "compose.yaml" \) -type f 2>/dev/null | head -1 | grep -q .; then
        frameworks=$(echo "$frameworks" | jq '. + ["docker_compose"]')
    fi

    # Helm
    if find "$repo_dir" -name "Chart.yaml" -type f 2>/dev/null | head -1 | grep -q .; then
        frameworks=$(echo "$frameworks" | jq '. + ["helm"]')
    fi

    # Kustomize
    if find "$repo_dir" -name "kustomization.yaml" -type f 2>/dev/null | head -1 | grep -q .; then
        frameworks=$(echo "$frameworks" | jq '. + ["kustomize"]')
    fi

    # ARM (Azure)
    if find "$repo_dir" -name "*.json" -type f 2>/dev/null | \
       xargs grep -l '"$schema".*schema.management.azure.com' 2>/dev/null | head -1 | grep -q .; then
        frameworks=$(echo "$frameworks" | jq '. + ["arm"]')
    fi

    # Bicep (Azure)
    if find "$repo_dir" -name "*.bicep" -type f 2>/dev/null | head -1 | grep -q .; then
        frameworks=$(echo "$frameworks" | jq '. + ["bicep"]')
    fi

    # GitHub Actions
    if [[ -d "$repo_dir/.github/workflows" ]]; then
        frameworks=$(echo "$frameworks" | jq '. + ["github_actions"]')
    fi

    # GitLab CI
    if [[ -f "$repo_dir/.gitlab-ci.yml" ]] || [[ -f "$repo_dir/.gitlab-ci.yaml" ]]; then
        frameworks=$(echo "$frameworks" | jq '. + ["gitlab_ci"]')
    fi

    # Serverless Framework
    if find "$repo_dir" \( -name "serverless.yml" -o -name "serverless.yaml" \) -type f 2>/dev/null | head -1 | grep -q .; then
        frameworks=$(echo "$frameworks" | jq '. + ["serverless"]')
    fi

    # Ansible
    if find "$repo_dir" -name "ansible.cfg" -o -name "playbook.yml" -o -name "site.yml" -type f 2>/dev/null | head -1 | grep -q .; then
        frameworks=$(echo "$frameworks" | jq '. + ["ansible"]')
    fi

    echo "$frameworks"
}

# Run Checkov and parse results
run_checkov() {
    local repo_dir="$1"
    local frameworks="$2"

    local checkov_output
    local checkov_args=(
        "--directory" "$repo_dir"
        "--output" "json"
        "--quiet"
        "--compact"
    )

    # Add framework filter if specified
    if [[ -n "$frameworks" ]]; then
        checkov_args+=("--framework" "$frameworks")
    fi

    # Run checkov and capture output
    checkov_output=$("$CHECKOV_BIN" "${checkov_args[@]}" 2>/dev/null) || true

    echo "$checkov_output"
}

# Parse Checkov JSON output into our format
parse_checkov_results() {
    local checkov_json="$1"

    # Check if we got valid JSON
    if ! echo "$checkov_json" | jq empty 2>/dev/null; then
        echo '{"passed": [], "failed": [], "skipped": []}'
        return
    fi

    # Checkov outputs an array of results per framework
    # Normalize to our format
    local all_passed="[]"
    local all_failed="[]"
    local all_skipped="[]"

    # Handle both array and object responses
    if echo "$checkov_json" | jq -e 'type == "array"' > /dev/null 2>&1; then
        # Array of framework results
        all_passed=$(echo "$checkov_json" | jq '[.[] | .results.passed_checks // [] | .[]] | unique_by(.check_id + .file_path + (.file_line_range | tostring))')
        all_failed=$(echo "$checkov_json" | jq '[.[] | .results.failed_checks // [] | .[]] | unique_by(.check_id + .file_path + (.file_line_range | tostring))')
        all_skipped=$(echo "$checkov_json" | jq '[.[] | .results.skipped_checks // [] | .[]] | unique_by(.check_id + .file_path)')
    else
        # Single framework result
        all_passed=$(echo "$checkov_json" | jq '.results.passed_checks // []')
        all_failed=$(echo "$checkov_json" | jq '.results.failed_checks // []')
        all_skipped=$(echo "$checkov_json" | jq '.results.skipped_checks // []')
    fi

    # Build normalized findings
    jq -n \
        --argjson passed "$all_passed" \
        --argjson failed "$all_failed" \
        --argjson skipped "$all_skipped" \
        '{
            passed: $passed,
            failed: $failed,
            skipped: $skipped
        }'
}

# Extract compliance framework mappings from findings
extract_compliance_mappings() {
    local findings="$1"

    # Extract unique guideline references from failed checks
    local guidelines=$(echo "$findings" | jq -r '
        .failed // [] |
        [.[].guideline // empty] |
        unique |
        map(select(. != null and . != ""))
    ')

    # Extract check IDs and categorize by compliance framework
    local check_ids=$(echo "$findings" | jq -r '
        .failed // [] |
        [.[].check_id // empty] |
        unique
    ')

    # Map to compliance frameworks based on check prefixes
    local compliance='{}'

    # CIS checks
    local cis_count=$(echo "$check_ids" | jq '[.[] | select(startswith("CKV_") or contains("CIS"))] | length')

    # AWS checks
    local aws_count=$(echo "$check_ids" | jq '[.[] | select(startswith("CKV_AWS"))] | length')

    # Azure checks
    local azure_count=$(echo "$check_ids" | jq '[.[] | select(startswith("CKV_AZURE") or startswith("CKV_ARM"))] | length')

    # GCP checks
    local gcp_count=$(echo "$check_ids" | jq '[.[] | select(startswith("CKV_GCP"))] | length')

    # Docker checks
    local docker_count=$(echo "$check_ids" | jq '[.[] | select(startswith("CKV_DOCKER"))] | length')

    # Kubernetes checks
    local k8s_count=$(echo "$check_ids" | jq '[.[] | select(startswith("CKV_K8S"))] | length')

    jq -n \
        --argjson cis "$cis_count" \
        --argjson aws "$aws_count" \
        --argjson azure "$azure_count" \
        --argjson gcp "$gcp_count" \
        --argjson docker "$docker_count" \
        --argjson k8s "$k8s_count" \
        '{
            cis_benchmark: $cis,
            aws_best_practices: $aws,
            azure_best_practices: $azure,
            gcp_best_practices: $gcp,
            docker_best_practices: $docker,
            kubernetes_security: $k8s
        }'
}

# Categorize findings by severity
categorize_by_severity() {
    local findings="$1"

    # Checkov uses severity levels: CRITICAL, HIGH, MEDIUM, LOW, INFO
    local critical=$(echo "$findings" | jq '[.failed // [] | .[] | select(.severity == "CRITICAL")] | length')
    local high=$(echo "$findings" | jq '[.failed // [] | .[] | select(.severity == "HIGH")] | length')
    local medium=$(echo "$findings" | jq '[.failed // [] | .[] | select(.severity == "MEDIUM" or .severity == null)] | length')
    local low=$(echo "$findings" | jq '[.failed // [] | .[] | select(.severity == "LOW")] | length')
    local info=$(echo "$findings" | jq '[.failed // [] | .[] | select(.severity == "INFO")] | length')

    jq -n \
        --argjson critical "$critical" \
        --argjson high "$high" \
        --argjson medium "$medium" \
        --argjson low "$low" \
        --argjson info "$info" \
        '{
            critical: $critical,
            high: $high,
            medium: $medium,
            low: $low,
            info: $info
        }'
}

# Format findings for output
format_findings() {
    local findings="$1"
    local repo_dir="$2"

    # Format failed checks with relative paths
    echo "$findings" | jq --arg base "$repo_dir" '
        .failed // [] | map({
            check_id: .check_id,
            check_name: .check_name,
            check_type: .check_type,
            severity: (.severity // "MEDIUM"),
            file: (.file_path | sub($base + "/"; "")),
            resource: .resource,
            line_range: .file_line_range,
            guideline: .guideline,
            bc_category: .bc_category
        })
    '
}

# Main analysis
analyze_target() {
    local repo_dir="$1"

    echo -e "${BLUE}Detecting IaC frameworks...${NC}" >&2
    local detected_frameworks=$(detect_frameworks "$repo_dir")
    local framework_count=$(echo "$detected_frameworks" | jq 'length')
    echo -e "${GREEN}✓ Found $framework_count IaC frameworks${NC}" >&2

    if [[ "$framework_count" -eq 0 ]]; then
        echo -e "${YELLOW}No IaC files detected${NC}" >&2
        # Return empty result
        local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
        jq -n \
            --arg ts "$timestamp" \
            --arg tgt "$TARGET" \
            --arg ver "1.0.0" \
            '{
                analyzer: "iac-security",
                version: $ver,
                timestamp: $ts,
                target: $tgt,
                status: "no_iac_detected",
                frameworks_detected: [],
                summary: {
                    total_findings: 0,
                    by_severity: {critical: 0, high: 0, medium: 0, low: 0, info: 0}
                },
                findings: [],
                compliance_mapping: {},
                note: "No Infrastructure-as-Code files were detected in this repository."
            }'
        return
    fi

    echo -e "${BLUE}Running Checkov security scan...${NC}" >&2

    # Run checkov with specific frameworks or auto-detect
    local framework_arg=""
    if [[ -n "$FRAMEWORKS" ]]; then
        framework_arg="$FRAMEWORKS"
    fi

    local checkov_output=$(run_checkov "$repo_dir" "$framework_arg")

    # Check if checkov returned valid results
    if [[ -z "$checkov_output" ]] || ! echo "$checkov_output" | jq empty 2>/dev/null; then
        echo -e "${YELLOW}⚠ Checkov returned no results${NC}" >&2
        local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
        jq -n \
            --arg ts "$timestamp" \
            --arg tgt "$TARGET" \
            --arg ver "1.0.0" \
            --argjson frameworks "$detected_frameworks" \
            '{
                analyzer: "iac-security",
                version: $ver,
                timestamp: $ts,
                target: $tgt,
                status: "scan_completed",
                frameworks_detected: $frameworks,
                summary: {
                    total_findings: 0,
                    by_severity: {critical: 0, high: 0, medium: 0, low: 0, info: 0}
                },
                findings: [],
                compliance_mapping: {},
                note: "Checkov scan completed but returned no findings."
            }'
        return
    fi

    echo -e "${BLUE}Parsing scan results...${NC}" >&2
    local parsed_results=$(parse_checkov_results "$checkov_output")

    local passed_count=$(echo "$parsed_results" | jq '.passed | length')
    local failed_count=$(echo "$parsed_results" | jq '.failed | length')
    local skipped_count=$(echo "$parsed_results" | jq '.skipped | length')

    echo -e "${GREEN}✓ Passed: $passed_count, Failed: $failed_count, Skipped: $skipped_count${NC}" >&2

    # Extract additional metadata
    local severity_breakdown=$(categorize_by_severity "$parsed_results")
    local compliance_mapping=$(extract_compliance_mappings "$parsed_results")
    local formatted_findings=$(format_findings "$parsed_results" "$repo_dir")

    # Get Checkov version
    local checkov_version=$("$CHECKOV_BIN" --version 2>/dev/null | head -1 || echo "unknown")

    # Build final output
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    jq -n \
        --arg ts "$timestamp" \
        --arg tgt "$TARGET" \
        --arg ver "1.0.0" \
        --arg checkov_ver "$checkov_version" \
        --argjson frameworks "$detected_frameworks" \
        --argjson passed "$passed_count" \
        --argjson failed "$failed_count" \
        --argjson skipped "$skipped_count" \
        --argjson severity "$severity_breakdown" \
        --argjson findings "$formatted_findings" \
        --argjson compliance "$compliance_mapping" \
        '{
            analyzer: "iac-security",
            version: $ver,
            timestamp: $ts,
            target: $tgt,
            checkov_version: $checkov_ver,
            status: "scan_completed",
            frameworks_detected: $frameworks,
            summary: {
                total_checks: ($passed + $failed + $skipped),
                passed: $passed,
                failed: $failed,
                skipped: $skipped,
                by_severity: $severity
            },
            findings: $findings,
            compliance_mapping: $compliance,
            note: "Checkov IaC security scan. All findings require manual review."
        }'
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help) usage ;;
        --local-path)
            LOCAL_PATH="$2"
            shift 2
            ;;
        --framework)
            FRAMEWORKS="$2"
            shift 2
            ;;
        --compact)
            COMPACT=true
            shift
            ;;
        -o|--output)
            OUTPUT_FILE="$2"
            shift 2
            ;;
        -k|--keep-clone)
            CLEANUP=false
            shift
            ;;
        -*)
            echo "Unknown option: $1" >&2
            exit 1
            ;;
        *)
            TARGET="$1"
            shift
            ;;
    esac
done

# Check for Checkov
if ! check_checkov; then
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    result=$(jq -n \
        --arg ts "$timestamp" \
        --arg tgt "${TARGET:-unknown}" \
        '{
            analyzer: "iac-security",
            version: "1.0.0",
            timestamp: $ts,
            target: $tgt,
            status: "analyzer_not_found",
            error: "Checkov is not installed",
            install_instructions: {
                pip: "pip install checkov",
                brew: "brew install checkov",
                docs: "https://www.checkov.io/2.Basics/Installing%20Checkov.html"
            },
            summary: {
                total_findings: 0,
                by_severity: {critical: 0, high: 0, medium: 0, low: 0, info: 0}
            },
            findings: [],
            compliance_mapping: {}
        }')

    if [[ -n "$OUTPUT_FILE" ]]; then
        echo "$result" > "$OUTPUT_FILE"
        echo -e "${YELLOW}⚠ Checkov not installed - results indicate missing analyzer${NC}" >&2
    else
        echo "$result"
    fi
    exit 0
fi

# Main execution
scan_path=""

if [[ -n "$LOCAL_PATH" ]]; then
    [[ ! -d "$LOCAL_PATH" ]] && { echo '{"error": "Local path does not exist"}'; exit 1; }
    scan_path="$LOCAL_PATH"
    TARGET="$LOCAL_PATH"
elif [[ -n "$TARGET" ]]; then
    if is_git_url "$TARGET"; then
        clone_repository "$TARGET"
        scan_path="$TEMP_DIR"
    elif [[ -d "$TARGET" ]]; then
        scan_path="$TARGET"
    else
        echo '{"error": "Invalid target - must be URL or directory"}'
        exit 1
    fi
else
    echo '{"error": "No target specified"}'
    exit 1
fi

echo -e "${BLUE}Analyzing: $TARGET${NC}" >&2

final_json=$(analyze_target "$scan_path")

# Output
if [[ -n "$OUTPUT_FILE" ]]; then
    if [[ "$COMPACT" == true ]]; then
        echo "$final_json" | jq -c '.' > "$OUTPUT_FILE"
    else
        echo "$final_json" > "$OUTPUT_FILE"
    fi
    echo -e "${GREEN}✓ Results written to $OUTPUT_FILE${NC}" >&2
else
    if [[ "$COMPACT" == true ]]; then
        echo "$final_json" | jq -c '.'
    else
        echo "$final_json"
    fi
fi
