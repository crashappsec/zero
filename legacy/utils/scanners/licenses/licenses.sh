#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Legal Review - Data Collector
# Scans for licenses and content policy issues
# Outputs JSON for agent analysis - NO AI calls
#
# Usage: ./legal-analyser-data.sh [options] <target>
# Output: JSON with license findings, content issues, and metadata
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
SBOM_FILE=""           # Path to SBOM file (CycloneDX JSON)
TEMP_DIR=""
CLEANUP=true
TARGET=""
SCAN_LICENSES=true
SCAN_CONTENT=true

# License policy (defaults)
ALLOWED_LICENSES=("MIT" "Apache-2.0" "BSD-2-Clause" "BSD-3-Clause" "ISC" "Unlicense" "CC0-1.0")
DENIED_LICENSES=("GPL-2.0" "GPL-3.0" "AGPL-3.0")
REVIEW_LICENSES=("LGPL-2.1" "LGPL-3.0" "MPL-2.0" "EPL-1.0" "EPL-2.0")

usage() {
    cat << EOF
Legal Review - Data Collector (JSON output for agent analysis)

Usage: $0 [OPTIONS] <target>

TARGET:
    Git repository URL      Clone and analyze repository
    Local directory path    Analyze local repository

OPTIONS:
    --local-path PATH       Use pre-cloned repository (skips cloning)
    --sbom FILE             Use existing SBOM file for dependency licenses
    --licenses-only         Scan licenses only
    --content-only          Scan content policy only
    -o, --output FILE       Write JSON to file (default: stdout)
    -k, --keep-clone        Keep cloned repository
    -h, --help              Show this help

OUTPUT:
    JSON object with:
    - metadata: scan timestamp, target
    - licenses: detected licenses and policy status
    - content_issues: profanity, non-inclusive language findings

EXAMPLES:
    $0 https://github.com/expressjs/express
    $0 --local-path ~/.gibson/projects/foo/repo
    $0 -o legal.json /path/to/project

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
    [[ "$1" =~ ^(https?|git)://.*\.git$ ]] || [[ "$1" =~ ^git@.*:.*\.git$ ]] || [[ "$1" =~ github\.com|gitlab\.com|bitbucket\.org ]]
}

# Check if license is in array
license_in_array() {
    local license="$1"
    shift
    local array=("$@")
    for item in "${array[@]}"; do
        [[ "$license" == "$item" ]] && return 0
    done
    return 1
}

# Identify license from text
identify_license_from_text() {
    local file="$1"

    # MIT License detection (multiple forms)
    if grep -qi "MIT License" "$file" 2>/dev/null; then
        echo "MIT"
    elif grep -qi "Permission is hereby granted, free of charge" "$file" 2>/dev/null && \
         grep -qi "THE SOFTWARE IS PROVIDED \"AS IS\"" "$file" 2>/dev/null; then
        echo "MIT"
    elif grep -qi "Apache License" "$file" 2>/dev/null && grep -qi "Version 2.0" "$file" 2>/dev/null; then
        echo "Apache-2.0"
    elif grep -qi "GNU GENERAL PUBLIC LICENSE" "$file" 2>/dev/null && grep -qi "Version 3" "$file" 2>/dev/null; then
        echo "GPL-3.0"
    elif grep -qi "GNU GENERAL PUBLIC LICENSE" "$file" 2>/dev/null && grep -qi "Version 2" "$file" 2>/dev/null; then
        echo "GPL-2.0"
    elif grep -qi "GNU LESSER GENERAL PUBLIC LICENSE" "$file" 2>/dev/null; then
        echo "LGPL"
    elif grep -qi "GNU AFFERO GENERAL PUBLIC LICENSE" "$file" 2>/dev/null; then
        echo "AGPL-3.0"
    elif grep -qi "BSD" "$file" 2>/dev/null; then
        if grep -qi "3-Clause" "$file" 2>/dev/null; then
            echo "BSD-3-Clause"
        elif grep -qi "2-Clause" "$file" 2>/dev/null; then
            echo "BSD-2-Clause"
        else
            echo "BSD"
        fi
    elif grep -qi "Mozilla Public License" "$file" 2>/dev/null; then
        echo "MPL-2.0"
    elif grep -qi "ISC License" "$file" 2>/dev/null; then
        echo "ISC"
    elif grep -qi "Unlicense" "$file" 2>/dev/null; then
        echo "Unlicense"
    else
        echo "Unknown"
    fi
}

# Scan licenses
scan_licenses() {
    local repo_path="$1"
    local findings="[]"
    local violations=0
    local warnings=0
    local primary_license=""
    local primary_license_file=""

    # Find license files
    local license_files=$(find "$repo_path" -maxdepth 2 -type f \( -iname "LICENSE*" -o -iname "COPYING*" -o -iname "COPYRIGHT*" -o -iname "NOTICE*" \) 2>/dev/null)

    if [[ -n "$license_files" ]]; then
        while IFS= read -r file; do
            [[ -z "$file" ]] && continue
            local rel_path="${file#$repo_path/}"
            local license=$(identify_license_from_text "$file")

            local status="unknown"
            local policy_action="review"

            if license_in_array "$license" "${ALLOWED_LICENSES[@]}"; then
                status="allowed"
                policy_action="none"
            elif license_in_array "$license" "${DENIED_LICENSES[@]}"; then
                status="denied"
                policy_action="remove_or_replace"
                ((violations++))
            elif license_in_array "$license" "${REVIEW_LICENSES[@]}"; then
                status="review_required"
                policy_action="legal_review"
                ((warnings++))
            else
                status="unknown"
                policy_action="identify_and_review"
                ((warnings++))
            fi

            # Track primary license (first LICENSE file at root)
            if [[ -z "$primary_license" ]] && [[ "$rel_path" =~ ^LICENSE ]]; then
                primary_license="$license"
                primary_license_file="$rel_path"
            fi

            findings=$(echo "$findings" | jq \
                --arg file "$rel_path" \
                --arg license "$license" \
                --arg status "$status" \
                --arg action "$policy_action" \
                '. + [{"file": $file, "license": $license, "status": $status, "policy_action": $action}]')
        done <<< "$license_files"
    fi

    # Check package.json
    if [[ -f "$repo_path/package.json" ]]; then
        local npm_license=$(jq -r '.license // empty' "$repo_path/package.json" 2>/dev/null)
        if [[ -n "$npm_license" ]]; then
            local status="unknown"
            local policy_action="review"

            if license_in_array "$npm_license" "${ALLOWED_LICENSES[@]}"; then
                status="allowed"
                policy_action="none"
            elif license_in_array "$npm_license" "${DENIED_LICENSES[@]}"; then
                status="denied"
                policy_action="remove_or_replace"
                ((violations++))
            elif license_in_array "$npm_license" "${REVIEW_LICENSES[@]}"; then
                status="review_required"
                policy_action="legal_review"
                ((warnings++))
            fi

            findings=$(echo "$findings" | jq \
                --arg file "package.json" \
                --arg license "$npm_license" \
                --arg status "$status" \
                --arg action "$policy_action" \
                '. + [{"file": $file, "license": $license, "status": $status, "policy_action": $action}]')
        fi
    fi

    # Check Cargo.toml
    if [[ -f "$repo_path/Cargo.toml" ]]; then
        # Use awk instead of grep -oP for macOS compatibility
        local cargo_license=$(awk -F'"' '/^license\s*=/ {print $2}' "$repo_path/Cargo.toml" 2>/dev/null | head -1)
        if [[ -n "$cargo_license" ]]; then
            local status="unknown"
            if license_in_array "$cargo_license" "${ALLOWED_LICENSES[@]}"; then
                status="allowed"
            elif license_in_array "$cargo_license" "${DENIED_LICENSES[@]}"; then
                status="denied"
                ((violations++))
            fi

            findings=$(echo "$findings" | jq \
                --arg file "Cargo.toml" \
                --arg license "$cargo_license" \
                --arg status "$status" \
                '. + [{"file": $file, "license": $license, "status": $status}]')
        fi
    fi

    # Output with primary license info
    jq -n \
        --argjson findings "$findings" \
        --argjson violations "$violations" \
        --argjson warnings "$warnings" \
        --arg primary_license "$primary_license" \
        --arg primary_license_file "$primary_license_file" \
        '{
            findings: $findings,
            violations: $violations,
            warnings: $warnings,
            primary_license: (if $primary_license == "" then null else $primary_license end),
            primary_license_file: (if $primary_license_file == "" then null else $primary_license_file end)
        }'
}

# Scan dependency licenses from SBOM
scan_dependency_licenses() {
    local repo_path="$1"
    local provided_sbom="$2"  # Optional: explicitly provided SBOM path
    local sbom_file=""

    # Use provided SBOM if available
    if [[ -n "$provided_sbom" ]] && [[ -f "$provided_sbom" ]]; then
        sbom_file="$provided_sbom"
        echo -e "${BLUE}Using provided SBOM: $sbom_file${NC}" >&2
    else
        # Look for SBOM file in analysis directory (if running as part of hydration)
        local analysis_dir="${repo_path%/repo}/analysis"
        if [[ -f "$analysis_dir/sbom.cdx.json" ]]; then
            sbom_file="$analysis_dir/sbom.cdx.json"
        elif [[ -f "$repo_path/sbom.cdx.json" ]]; then
            sbom_file="$repo_path/sbom.cdx.json"
        elif [[ -f "$repo_path/bom.json" ]]; then
            sbom_file="$repo_path/bom.json"
        fi
    fi

    if [[ -z "$sbom_file" ]] || [[ ! -f "$sbom_file" ]]; then
        echo '{"by_license": {}, "all_packages": [], "denied": [], "review": []}'
        return
    fi

    echo -e "${BLUE}Analyzing dependency licenses from SBOM...${NC}" >&2

    # Extract license info from SBOM and categorize
    local result=$(jq '
        # Extract components with licenses
        [.components[]? | select(.licenses) | {
            name: .name,
            version: .version,
            ecosystem: (.purl // "" | split(":")[0] | gsub("pkg/"; "")),
            licenses: [.licenses[]?.license | .id // .name // "Unknown"]
        }] |

        # Group by license
        group_by(.licenses[0]) |
        map({
            license: .[0].licenses[0],
            count: length,
            packages: [.[] | {name, version, ecosystem}]
        }) |

        # Separate into categories
        {
            by_license: (map({(.license): {count, packages}}) | add // {}),
            all_packages: (map(.packages) | flatten),
            denied: [.[] | select(.license | test("GPL|AGPL"; "i")) | {license, count, packages}],
            review: [.[] | select(.license | test("LGPL|MPL|EPL"; "i")) | {license, count, packages}]
        }
    ' "$sbom_file" 2>/dev/null || echo '{"by_license": {}, "all_packages": [], "denied": [], "review": []}')

    echo "$result"
}

# Scan content policy
scan_content() {
    local repo_path="$1"
    local profanity_findings="[]"
    local inclusive_findings="[]"

    # Profanity patterns (basic list)
    local profanity_terms=("fuck" "shit" "damn" "wtf" "crap" "ass")

    # Non-inclusive terms
    local inclusive_terms=("master" "slave" "whitelist" "blacklist" "grandfathered" "sanity")

    # Search source files
    local source_files=$(find "$repo_path" -type f \( -name "*.js" -o -name "*.ts" -o -name "*.py" -o -name "*.go" -o -name "*.rb" -o -name "*.java" -o -name "*.sh" -o -name "*.md" \) \
        ! -path "*/node_modules/*" ! -path "*/.git/*" ! -path "*/vendor/*" 2>/dev/null | head -100)

    [[ -z "$source_files" ]] && echo '{"profanity": [], "inclusive_language": [], "profanity_count": 0, "inclusive_count": 0}' && return

    local profanity_count=0
    local inclusive_count=0

    while IFS= read -r file; do
        [[ -z "$file" ]] && continue
        [[ ! -f "$file" ]] && continue

        local rel_path="${file#$repo_path/}"

        # Check profanity
        for term in "${profanity_terms[@]}"; do
            local matches=$(grep -in "\b$term\b" "$file" 2>/dev/null | head -3)
            if [[ -n "$matches" ]]; then
                while IFS=: read -r line_num content; do
                    [[ -z "$line_num" ]] && continue
                    ((profanity_count++))
                    [[ $profanity_count -gt 20 ]] && break 2

                    profanity_findings=$(echo "$profanity_findings" | jq \
                        --arg file "$rel_path" \
                        --arg line "$line_num" \
                        --arg term "$term" \
                        '. + [{"file": $file, "line": ($line | tonumber), "term": $term}]')
                done <<< "$matches"
            fi
        done

        # Check inclusive language
        for term in "${inclusive_terms[@]}"; do
            local matches=$(grep -in "\b$term\b" "$file" 2>/dev/null | head -3)
            if [[ -n "$matches" ]]; then
                # Skip common false positives
                if echo "$matches" | grep -qi "git.*$term\|IDE.*$term"; then
                    continue
                fi

                while IFS=: read -r line_num content; do
                    [[ -z "$line_num" ]] && continue
                    ((inclusive_count++))
                    [[ $inclusive_count -gt 20 ]] && break 2

                    inclusive_findings=$(echo "$inclusive_findings" | jq \
                        --arg file "$rel_path" \
                        --arg line "$line_num" \
                        --arg term "$term" \
                        '. + [{"file": $file, "line": ($line | tonumber), "term": $term}]')
                done <<< "$matches"
            fi
        done
    done <<< "$source_files"

    jq -n \
        --argjson profanity "$profanity_findings" \
        --argjson inclusive "$inclusive_findings" \
        --argjson p_count "$profanity_count" \
        --argjson i_count "$inclusive_count" \
        '{profanity: $profanity, inclusive_language: $inclusive, profanity_count: $p_count, inclusive_count: $i_count}'
}

# Main analysis
analyze_target() {
    local repo_path="$1"
    local sbom_path="$2"  # Optional: explicitly provided SBOM path

    local license_results='{"findings": [], "violations": 0, "warnings": 0}'
    local content_results='{"profanity": [], "inclusive_language": [], "profanity_count": 0, "inclusive_count": 0}'
    local dep_license_results='{"by_license": {}, "all_packages": [], "denied": [], "review": []}'

    if [[ "$SCAN_LICENSES" == true ]]; then
        echo -e "${BLUE}Scanning project licenses...${NC}" >&2
        license_results=$(scan_licenses "$repo_path")
        local license_count=$(echo "$license_results" | jq '.findings | length')
        local violations=$(echo "$license_results" | jq '.violations')
        echo -e "${GREEN}✓ Found $license_count license references ($violations violations)${NC}" >&2

        # Also scan dependency licenses from SBOM
        dep_license_results=$(scan_dependency_licenses "$repo_path" "$sbom_path")
        local dep_count=$(echo "$dep_license_results" | jq '.all_packages | length')
        local denied_count=$(echo "$dep_license_results" | jq '[.denied[].count] | add // 0')
        echo -e "${GREEN}✓ Analyzed $dep_count dependency licenses ($denied_count with denied licenses)${NC}" >&2
    fi

    if [[ "$SCAN_CONTENT" == true ]]; then
        echo -e "${BLUE}Scanning content policy...${NC}" >&2
        content_results=$(scan_content "$repo_path")
        local profanity_count=$(echo "$content_results" | jq '.profanity_count')
        local inclusive_count=$(echo "$content_results" | jq '.inclusive_count')
        echo -e "${GREEN}✓ Content scan: $profanity_count profanity, $inclusive_count non-inclusive terms${NC}" >&2
    fi

    # Build final output
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Determine overall status
    local overall_status="pass"
    local license_violations=$(echo "$license_results" | jq '.violations')
    local dep_violations=$(echo "$dep_license_results" | jq '[.denied[].count] | add // 0')
    local total_violations=$((license_violations + dep_violations))

    if [[ "$total_violations" -gt 0 ]]; then
        overall_status="fail"
    elif [[ $(echo "$license_results" | jq '.warnings') -gt 0 ]]; then
        overall_status="warning"
    elif [[ $(echo "$dep_license_results" | jq '.review | length') -gt 0 ]]; then
        overall_status="warning"
    fi

    # Extract primary license info
    local primary_license=$(echo "$license_results" | jq -r '.primary_license // empty')
    local primary_license_file=$(echo "$license_results" | jq -r '.primary_license_file // empty')

    jq -n \
        --arg ts "$timestamp" \
        --arg tgt "$TARGET" \
        --arg ver "1.0.0" \
        --arg status "$overall_status" \
        --arg primary_license "$primary_license" \
        --arg primary_license_file "$primary_license_file" \
        --argjson licenses "$(echo "$license_results" | jq '.findings')" \
        --argjson license_violations "$license_violations" \
        --argjson license_warnings "$(echo "$license_results" | jq '.warnings')" \
        --argjson dep_violations "$dep_violations" \
        --argjson content "$content_results" \
        --argjson dep_licenses "$dep_license_results" \
        '{
            analyzer: "legal-review",
            version: $ver,
            timestamp: $ts,
            target: $tgt,
            summary: {
                overall_status: $status,
                license_violations: $license_violations,
                dependency_license_violations: $dep_violations,
                license_warnings: $license_warnings,
                profanity_issues: $content.profanity_count,
                inclusive_language_issues: $content.inclusive_count,
                total_dependencies_with_licenses: ($dep_licenses.all_packages | length)
            },
            repository_license: {
                license: (if $primary_license == "" then null else $primary_license end),
                file: (if $primary_license_file == "" then null else $primary_license_file end)
            },
            licenses: $licenses,
            dependency_licenses: {
                by_license: $dep_licenses.by_license,
                denied: $dep_licenses.denied,
                review_required: $dep_licenses.review
            },
            content_policy: {
                profanity: $content.profanity,
                inclusive_language: $content.inclusive_language
            },
            policy: {
                allowed_licenses: ["MIT", "Apache-2.0", "BSD-2-Clause", "BSD-3-Clause", "ISC"],
                denied_licenses: ["GPL-2.0", "GPL-3.0", "AGPL-3.0"],
                review_required: ["LGPL-2.1", "LGPL-3.0", "MPL-2.0"]
            }
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
        --sbom)
            SBOM_FILE="$2"
            shift 2
            ;;
        --licenses-only)
            SCAN_CONTENT=false
            shift
            ;;
        --content-only)
            SCAN_LICENSES=false
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

final_json=$(analyze_target "$scan_path" "$SBOM_FILE")

# Output
if [[ -n "$OUTPUT_FILE" ]]; then
    echo "$final_json" > "$OUTPUT_FILE"
    echo -e "${GREEN}✓ Results written to $OUTPUT_FILE${NC}" >&2
else
    echo "$final_json"
fi
