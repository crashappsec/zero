#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Container Security Scanner
# Analyzes Dockerfiles for best practices, detects hardened images,
# scans for vulnerabilities, and provides optimization recommendations
#
# Usage: ./container-security.sh [options] <target>
# Output: JSON with Dockerfile analysis, vulnerabilities, and recommendations
#############################################################################

set -e

# Colors for terminal output (stderr only)
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SCANNERS_ROOT="$(dirname "$SCRIPT_DIR")"
UTILS_ROOT="$(dirname "$SCANNERS_ROOT")"
REPO_ROOT="$(dirname "$UTILS_ROOT")"

# Source library files
source "$SCRIPT_DIR/lib/dockerfile-analyzer.sh"
source "$SCRIPT_DIR/lib/hardened-detector.sh"
source "$SCRIPT_DIR/lib/multistage-analyzer.sh"
source "$SCRIPT_DIR/lib/image-scanner.sh"

# Default options
OUTPUT_FILE=""
LOCAL_PATH=""
REPO=""
ORG=""
TEMP_DIR=""
CLEANUP=true
TARGET=""
SCAN_IMAGES=false

# Version
VERSION="1.0.0"

usage() {
    cat << EOF
Container Security Scanner - Analyze Dockerfiles and container images

Usage: $0 [OPTIONS] <target>

TARGET:
    Git repository URL      Clone and analyze repository
    Local directory path    Analyze local repository

OPTIONS:
    --local-path PATH       Use pre-cloned repository (skips cloning)
    --repo OWNER/REPO       GitHub repository (looks in zero cache)
    --org ORG               GitHub org (uses first repo found in zero cache)
    -o, --output FILE       Write JSON to file (default: stdout)
    --scan-images           Scan container images with trivy/grype (slower)
    -k, --keep-clone        Keep cloned repository
    -h, --help              Show this help

OUTPUT:
    JSON object with:
    - summary: Dockerfile counts, vulnerability totals, hardening score
    - dockerfiles: Analysis of each Dockerfile found
    - hardening_analysis: Base image security assessment
    - multistage_analysis: Multi-stage build recommendations
    - images: Vulnerability scan results (if --scan-images)

EXAMPLES:
    $0 https://github.com/expressjs/express
    $0 --local-path ~/.zero/repos/foo/repo
    $0 -o container-security.json /path/to/project
    $0 --scan-images myproject/  # Also scan built images

EOF
    exit 0
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
    local url="$1"
    [[ "$url" =~ ^(https?://|git@|git://) ]]
}

# Main analysis function
analyze_container_security() {
    local target="$1"

    echo -e "${BLUE}Analyzing container security for: ${CYAN}$target${NC}" >&2

    # Find Dockerfiles
    local dockerfiles
    dockerfiles=$(find_dockerfiles "$target")
    local dockerfile_count
    if [[ -z "$dockerfiles" ]]; then
        dockerfile_count=0
    else
        dockerfile_count=$(echo "$dockerfiles" | wc -l | tr -d ' ')
    fi

    echo -e "${BLUE}Found ${CYAN}$dockerfile_count${BLUE} Dockerfile(s)${NC}" >&2

    if [[ "$dockerfile_count" -eq 0 ]]; then
        jq -n \
            --arg analyzer "container-security" \
            --arg version "$VERSION" \
            --arg timestamp "$(date -u +"%Y-%m-%dT%H:%M:%SZ")" \
            --arg target "$target" \
            '{
                analyzer: $analyzer,
                version: $version,
                timestamp: $timestamp,
                target: $target,
                status: "no_dockerfiles_found",
                summary: {
                    dockerfiles_found: 0,
                    images_analyzed: 0,
                    total_vulnerabilities: 0,
                    by_severity: {critical: 0, high: 0, medium: 0, low: 0},
                    dockerfile_issues: 0,
                    uses_multistage: false,
                    uses_hardened_base: false,
                    hardening_score: 0
                },
                dockerfiles: [],
                hardening_analysis: {},
                multistage_analysis: {},
                images: [],
                recommendations: ["No Dockerfiles found in this repository"]
            }'
        return
    fi

    # Analyze all Dockerfiles
    echo -e "${BLUE}Analyzing Dockerfiles...${NC}" >&2
    local dockerfile_analyses
    dockerfile_analyses=$(analyze_all_dockerfiles "$target")

    # Collect all base images
    local all_base_images
    all_base_images=$(echo "$dockerfile_analyses" | jq '[.[].final_base] | unique')

    # Analyze hardening for each base image
    echo -e "${BLUE}Analyzing image hardening...${NC}" >&2
    local hardening_analyses
    hardening_analyses=$(analyze_images_hardening "$all_base_images")

    # Check if any are hardened
    local uses_hardened
    uses_hardened=$(echo "$hardening_analyses" | jq '[.[].is_hardened] | any')

    # Get image types for hardening score
    local image_types
    image_types=$(echo "$hardening_analyses" | jq '[.[].classification]')
    local hardening_score
    hardening_score=$(calculate_hardening_score "$image_types")

    # Analyze multi-stage builds
    echo -e "${BLUE}Analyzing multi-stage builds...${NC}" >&2
    local multistage_results='[]'
    local any_multistage="false"

    while IFS= read -r dockerfile; do
        [[ -z "$dockerfile" ]] && continue
        local rel_path="${dockerfile#$target/}"

        # Detect language from Dockerfile for better recommendations
        local first_base
        first_base=$(grep -iE "^FROM " "$dockerfile" | head -1 | awk '{print $2}' || echo "")
        local language
        language=$(detect_language "$first_base")

        local ms_analysis
        ms_analysis=$(analyze_multistage "$dockerfile" "$language")

        if echo "$ms_analysis" | jq -e '.is_multistage == true' &>/dev/null; then
            any_multistage="true"
        fi

        multistage_results=$(echo "$multistage_results" | jq \
            --arg path "$rel_path" \
            --argjson analysis "$ms_analysis" \
            '. + [{path: $path, analysis: $analysis}]')
    done < <(echo "$dockerfiles")

    # Count total issues
    local total_issues
    total_issues=$(echo "$dockerfile_analyses" | jq '[.[].issue_counts.total] | add // 0')
    local error_count warning_count info_count
    error_count=$(echo "$dockerfile_analyses" | jq '[.[].issue_counts.error] | add // 0')
    warning_count=$(echo "$dockerfile_analyses" | jq '[.[].issue_counts.warning] | add // 0')
    info_count=$(echo "$dockerfile_analyses" | jq '[.[].issue_counts.info] | add // 0')

    # Image scanning (if enabled)
    local image_scans='[]'
    local total_vulns=0
    local vuln_critical=0 vuln_high=0 vuln_medium=0 vuln_low=0

    if [[ "$SCAN_IMAGES" == "true" ]]; then
        echo -e "${BLUE}Scanning images for vulnerabilities...${NC}" >&2

        while IFS= read -r image; do
            [[ -z "$image" ]] && continue
            [[ "$image" == "null" ]] && continue

            echo -e "  ${CYAN}Scanning: $image${NC}" >&2
            local scan_result
            scan_result=$(scan_image "$image")

            if ! echo "$scan_result" | jq -e '.error' &>/dev/null; then
                vuln_critical=$((vuln_critical + $(echo "$scan_result" | jq '.summary.critical // 0')))
                vuln_high=$((vuln_high + $(echo "$scan_result" | jq '.summary.high // 0')))
                vuln_medium=$((vuln_medium + $(echo "$scan_result" | jq '.summary.medium // 0')))
                vuln_low=$((vuln_low + $(echo "$scan_result" | jq '.summary.low // 0')))
            fi

            image_scans=$(echo "$image_scans" | jq --argjson scan "$scan_result" '. + [$scan]')
        done < <(echo "$all_base_images" | jq -r '.[]')

        total_vulns=$((vuln_critical + vuln_high + vuln_medium + vuln_low))
    fi

    # Get scanner status
    local scanner_status
    scanner_status=$(get_scanner_status)

    # Generate overall recommendations
    local recommendations='[]'

    # Hardening recommendations
    if [[ "$uses_hardened" != "true" ]]; then
        recommendations=$(echo "$recommendations" | jq '. + ["Consider using hardened base images (Chainguard or Distroless) for improved security"]')
    fi

    # Multi-stage recommendations
    if [[ "$any_multistage" != "true" ]]; then
        recommendations=$(echo "$recommendations" | jq '. + ["Use multi-stage builds to reduce final image size and attack surface"]')
    fi

    # Issue-based recommendations
    if [[ "$error_count" -gt 0 ]]; then
        recommendations=$(echo "$recommendations" | jq --argjson count "$error_count" '. + [("Fix " + ($count | tostring) + " critical Dockerfile issues")]')
    fi

    if [[ "$warning_count" -gt 5 ]]; then
        recommendations=$(echo "$recommendations" | jq '. + ["Review and address Dockerfile best practice warnings"]')
    fi

    # Vulnerability recommendations
    if [[ "$vuln_critical" -gt 0 ]]; then
        recommendations=$(echo "$recommendations" | jq --argjson count "$vuln_critical" '. + [("Address " + ($count | tostring) + " critical vulnerabilities immediately")]')
    fi

    echo -e "${GREEN}✓ Analysis complete${NC}" >&2

    # Build final output
    jq -n \
        --arg analyzer "container-security" \
        --arg version "$VERSION" \
        --arg timestamp "$(date -u +"%Y-%m-%dT%H:%M:%SZ")" \
        --arg target "$target" \
        --argjson dockerfile_count "$dockerfile_count" \
        --argjson images_analyzed "$(echo "$image_scans" | jq 'length')" \
        --argjson total_vulns "$total_vulns" \
        --argjson vuln_critical "$vuln_critical" \
        --argjson vuln_high "$vuln_high" \
        --argjson vuln_medium "$vuln_medium" \
        --argjson vuln_low "$vuln_low" \
        --argjson dockerfile_issues "$total_issues" \
        --argjson error_count "$error_count" \
        --argjson warning_count "$warning_count" \
        --argjson info_count "$info_count" \
        --arg uses_multistage "$any_multistage" \
        --arg uses_hardened "$uses_hardened" \
        --argjson hardening_score "$hardening_score" \
        --argjson dockerfiles "$dockerfile_analyses" \
        --argjson hardening_analyses "$hardening_analyses" \
        --argjson multistage_results "$multistage_results" \
        --argjson image_scans "$image_scans" \
        --argjson scanner_status "$scanner_status" \
        --argjson recommendations "$recommendations" \
        '{
            analyzer: $analyzer,
            version: $version,
            timestamp: $timestamp,
            target: $target,
            status: "scan_completed",
            summary: {
                dockerfiles_found: $dockerfile_count,
                images_analyzed: $images_analyzed,
                total_vulnerabilities: $total_vulns,
                by_severity: {
                    critical: $vuln_critical,
                    high: $vuln_high,
                    medium: $vuln_medium,
                    low: $vuln_low
                },
                dockerfile_issues: $dockerfile_issues,
                issue_breakdown: {
                    error: $error_count,
                    warning: $warning_count,
                    info: $info_count
                },
                uses_multistage: ($uses_multistage == "true"),
                uses_hardened_base: ($uses_hardened == "true"),
                hardening_score: $hardening_score
            },
            dockerfiles: $dockerfiles,
            hardening_analysis: $hardening_analyses,
            multistage_analysis: $multistage_results,
            images: $image_scans,
            scanner_tools: $scanner_status,
            recommendations: $recommendations
        }'
}

# Parse arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
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
            -o|--output)
                OUTPUT_FILE="$2"
                shift 2
                ;;
            --scan-images)
                SCAN_IMAGES=true
                shift
                ;;
            -k|--keep-clone)
                CLEANUP=false
                shift
                ;;
            -h|--help)
                usage
                ;;
            *)
                TARGET="$1"
                shift
                ;;
        esac
    done
}

# Main execution
main() {
    parse_args "$@"

    # Determine target directory
    local target_dir=""

    if [[ -n "$LOCAL_PATH" ]]; then
        target_dir="$LOCAL_PATH"
    elif [[ -n "$REPO" ]]; then
        # Look in zero cache
        local zero_home="${ZERO_HOME:-$HOME/.zero}"
        target_dir="$zero_home/repos/$REPO/repo"
        if [[ ! -d "$target_dir" ]]; then
            echo '{"error": "Repository not found in zero cache", "repo": "'"$REPO"'"}'
            exit 1
        fi
    elif [[ -n "$ORG" ]]; then
        # Find first repo in org cache
        local zero_home="${ZERO_HOME:-$HOME/.zero}"
        local org_dir="$zero_home/repos/$ORG"
        if [[ -d "$org_dir" ]]; then
            target_dir=$(find "$org_dir" -maxdepth 2 -type d -name "repo" | head -1)
        fi
        if [[ -z "$target_dir" ]]; then
            echo '{"error": "No repositories found for org", "org": "'"$ORG"'"}'
            exit 1
        fi
    elif [[ -n "$TARGET" ]]; then
        if is_git_url "$TARGET"; then
            clone_repository "$TARGET"
            target_dir="$TEMP_DIR"
        elif [[ -d "$TARGET" ]]; then
            target_dir="$TARGET"
        else
            echo '{"error": "Target not found", "target": "'"$TARGET"'"}'
            exit 1
        fi
    else
        usage
    fi

    # Run analysis
    local output
    output=$(analyze_container_security "$target_dir")

    # Output result
    if [[ -n "$OUTPUT_FILE" ]]; then
        echo "$output" > "$OUTPUT_FILE"
        echo -e "${GREEN}✓ Output written to: ${CYAN}$OUTPUT_FILE${NC}" >&2
    else
        echo "$output"
    fi
}

# Run if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
