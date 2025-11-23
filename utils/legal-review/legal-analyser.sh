#!/bin/bash
# Legal Review Analyser - Code Legal Compliance Scanner
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0

set -euo pipefail

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
UTILS_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
REPO_ROOT="$(cd "$UTILS_ROOT/.." && pwd)"

# Load global libraries
source "$REPO_ROOT/utils/lib/config.sh"
source "$REPO_ROOT/utils/lib/github.sh"

# Configuration
LEGAL_CONFIG="${REPO_ROOT}/config/legal-review-config.json"
VERBOSE=false
OUTPUT_FORMAT="markdown"
OUTPUT_FILE=""
SCAN_LICENSES=true
SCAN_SECRETS=true
SCAN_CONTENT=true
USE_CLAUDE=false
TARGET_REPO=""
TARGET_PATH=""
TARGET_ORG=""
LOCAL_PATH=""
SBOM_FILE=""
COMPARE_MODE=false
PARALLEL=false
PARALLEL_JOBS=$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo "4")
ANTHROPIC_API_KEY="${ANTHROPIC_API_KEY:-}"
REPORTS_DIR="${REPO_ROOT}/reports/legal-review"
AUTO_SAVE_REPORTS=true

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Temp directories tracking
TEMP_DIRS=()

# Scan results storage (for Claude AI enhancement)
SCAN_RESULTS_FILE=""
LICENSE_VIOLATIONS=()
CONTENT_ISSUES=()

# Cleanup function
cleanup() {
    if [[ ${#TEMP_DIRS[@]} -gt 0 ]]; then
        for temp_dir in "${TEMP_DIRS[@]}"; do
            if [[ -n "$temp_dir" ]] && [[ -d "$temp_dir" ]]; then
                rm -rf "$temp_dir"
            fi
        done
    fi
    [[ -n "$SCAN_RESULTS_FILE" ]] && rm -f "$SCAN_RESULTS_FILE"
    # Cost tracking cleanup is handled by temp file cleanup
}

trap cleanup EXIT

# Usage
usage() {
    cat <<EOF
Legal Review Analyser - Comprehensive code legal compliance scanner

Usage: $0 [OPTIONS]

TARGET OPTIONS:
    --repo OWNER/REPO          Analyze single GitHub repository
    --org ORG_NAME             Analyze all repositories in GitHub organization
    --path PATH                Analyze local path
    --local-path PATH          Use pre-cloned repository (skip cloning)
    --sbom FILE                Use existing SBOM file for dependency analysis

SCAN OPTIONS:
    --licenses-only            Scan licenses only
    --secrets-only             Scan secrets only
    --content-only             Scan content policy only
    --parallel                 Enable parallel file processing (faster)
    --jobs N                   Number of parallel jobs (default: CPU count)

OUTPUT OPTIONS:
    --format FORMAT            Output format: markdown (default), json, table
    --output FILE              Write output to file
    --verbose                  Enable verbose output

CLAUDE AI OPTIONS:
    --claude                   Use Claude AI for enhanced analysis
    --compare                  Run both basic and Claude modes side-by-side
    -k, --api-key KEY          Anthropic API key (or set ANTHROPIC_API_KEY env var)

OTHER OPTIONS:
    -h, --help                 Show this help message

EXAMPLES:
    # Full analysis of single repository
    $0 --repo owner/repo

    # Analyze all repos in organization
    $0 --org my-organization

    # License scan only with parallel processing
    $0 --repo owner/repo --licenses-only --parallel

    # Local path with JSON output
    $0 --path /path/to/code --format json --output report.json

    # Claude AI enhanced analysis
    $0 --repo owner/repo --claude

    # Compare basic vs Claude analysis
    $0 --repo owner/repo --compare

    # Use pre-cloned repository
    $0 --local-path /tmp/cloned-repo

    # Multi-repo analysis with table output
    $0 --org my-org --format table --parallel

EOF
    exit 0
}

# Parse arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --repo)
                TARGET_REPO="$2"
                shift 2
                ;;
            --org)
                TARGET_ORG="$2"
                shift 2
                ;;
            --path)
                TARGET_PATH="$2"
                shift 2
                ;;
            --local-path)
                LOCAL_PATH="$2"
                shift 2
                ;;
            --sbom)
                SBOM_FILE="$2"
                shift 2
                ;;
            --licenses-only)
                SCAN_SECRETS=false
                SCAN_CONTENT=false
                shift
                ;;
            --secrets-only)
                SCAN_LICENSES=false
                SCAN_CONTENT=false
                shift
                ;;
            --content-only)
                SCAN_LICENSES=false
                SCAN_SECRETS=false
                shift
                ;;
            --parallel)
                PARALLEL=true
                shift
                ;;
            --jobs)
                PARALLEL_JOBS="$2"
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
            --claude)
                USE_CLAUDE=true
                shift
                ;;
            --compare)
                COMPARE_MODE=true
                USE_CLAUDE=true
                shift
                ;;
            -k|--api-key)
                ANTHROPIC_API_KEY="$2"
                shift 2
                ;;
            --verbose)
                VERBOSE=true
                shift
                ;;
            -h|--help)
                usage
                ;;
            *)
                echo "Unknown option: $1"
                usage
                ;;
        esac
    done
}

# Log function
log() {
    if [[ "$VERBOSE" == true ]]; then
        echo -e "${BLUE}[INFO]${NC} $*" >&2
    fi
}

# Load configuration
load_config() {
    if [[ ! -f "$LEGAL_CONFIG" ]]; then
        echo -e "${YELLOW}⚠ Config not found: $LEGAL_CONFIG${NC}" >&2
        echo -e "${YELLOW}  Using default settings${NC}" >&2
        return 1
    fi

    log "Loaded configuration from $LEGAL_CONFIG"
    return 0
}

# License policy arrays
ALLOWED_LICENSES=()
DENIED_LICENSES=()
REVIEW_LICENSES=()

# Load license policy from config
load_license_policy() {
    if [[ ! -f "$LEGAL_CONFIG" ]]; then
        # Default policy
        ALLOWED_LICENSES=("MIT" "Apache-2.0" "BSD-2-Clause" "BSD-3-Clause" "ISC")
        DENIED_LICENSES=("GPL-2.0" "GPL-3.0" "AGPL-3.0")
        REVIEW_LICENSES=("LGPL-2.1" "LGPL-3.0" "MPL-2.0")
        return
    fi

    # Parse JSON config
    if command -v jq &> /dev/null; then
        while IFS= read -r license; do
            [[ -n "$license" ]] && ALLOWED_LICENSES+=("$license")
        done < <(jq -r '.legal_review.licenses.allowed.list[]?' "$LEGAL_CONFIG" 2>/dev/null || echo "MIT")

        while IFS= read -r license; do
            [[ -n "$license" ]] && DENIED_LICENSES+=("$license")
        done < <(jq -r '.legal_review.licenses.denied.list[]?' "$LEGAL_CONFIG" 2>/dev/null || echo "GPL-3.0")

        while IFS= read -r license; do
            [[ -n "$license" ]] && REVIEW_LICENSES+=("$license")
        done < <(jq -r '.legal_review.licenses.review_required.list[]?' "$LEGAL_CONFIG" 2>/dev/null)
    fi
}

# Check if license is in array
license_in_array() {
    local license="$1"
    shift
    local array=("$@")

    for item in "${array[@]}"; do
        if [[ "$license" == "$item" ]]; then
            return 0
        fi
    done
    return 1
}

# Detect license files
detect_license_files() {
    local path="$1"

    local license_files=()
    local patterns=("LICENSE" "LICENSE.txt" "LICENSE.md" "COPYING" "COPYING.txt" "COPYRIGHT" "NOTICE")

    for pattern in "${patterns[@]}"; do
        while IFS= read -r -d '' file; do
            license_files+=("$file")
        done < <(find "$path" -maxdepth 2 -iname "$pattern" -type f -print0 2>/dev/null)
    done

    if [[ ${#license_files[@]} -gt 0 ]]; then
        printf '%s\n' "${license_files[@]}"
    fi
}

# Extract SPDX identifier from file
extract_spdx_from_file() {
    local file="$1"

    # Look for SPDX-License-Identifier in first 20 lines
    head -20 "$file" 2>/dev/null | grep -i "SPDX-License-Identifier:" | sed -E 's/.*SPDX-License-Identifier:[[:space:]]*([A-Za-z0-9\.\-]+).*/\1/' | head -1
}

# Detect license from package.json
detect_npm_license() {
    local path="$1"

    if [[ -f "$path/package.json" ]]; then
        if command -v jq &> /dev/null; then
            jq -r '.license // empty' "$path/package.json" 2>/dev/null
        else
            grep -oP '"license"\s*:\s*"\K[^"]+' "$path/package.json" 2>/dev/null | head -1
        fi
    fi
}

# Detect license from Cargo.toml
detect_cargo_license() {
    local path="$1"

    if [[ -f "$path/Cargo.toml" ]]; then
        grep -oP '^license\s*=\s*"\K[^"]+' "$path/Cargo.toml" 2>/dev/null | head -1
    fi
}

# Detect license from pom.xml
detect_maven_license() {
    local path="$1"

    if [[ -f "$path/pom.xml" ]]; then
        grep -oP '<name>\K[^<]+' "$path/pom.xml" 2>/dev/null | grep -i license | head -1
    fi
}

# Identify license from text
identify_license_from_text() {
    local file="$1"

    # Simple pattern matching for common licenses
    if grep -qi "MIT License" "$file" 2>/dev/null; then
        echo "MIT"
    elif grep -qi "Apache License.*Version 2.0" "$file" 2>/dev/null || \
         (grep -qi "Apache License" "$file" 2>/dev/null && grep -qi "Version 2.0" "$file" 2>/dev/null); then
        echo "Apache-2.0"
    elif grep -qi "GNU GENERAL PUBLIC LICENSE.*Version 3" "$file" 2>/dev/null || \
         (grep -qi "GNU GENERAL PUBLIC LICENSE" "$file" 2>/dev/null && grep -qi "Version 3" "$file" 2>/dev/null); then
        echo "GPL-3.0"
    elif grep -qi "GNU GENERAL PUBLIC LICENSE.*Version 2" "$file" 2>/dev/null || \
         (grep -qi "GNU GENERAL PUBLIC LICENSE" "$file" 2>/dev/null && grep -qi "Version 2" "$file" 2>/dev/null); then
        echo "GPL-2.0"
    elif grep -qi "GNU LESSER GENERAL PUBLIC LICENSE.*Version 3" "$file" 2>/dev/null || \
         (grep -qi "GNU LESSER GENERAL PUBLIC LICENSE" "$file" 2>/dev/null && grep -qi "Version 3" "$file" 2>/dev/null); then
        echo "LGPL-3.0"
    elif grep -qi "GNU AFFERO GENERAL PUBLIC LICENSE" "$file" 2>/dev/null && grep -qi "Version 3" "$file" 2>/dev/null; then
        echo "AGPL-3.0"
    elif grep -qi "BSD.*License" "$file" 2>/dev/null; then
        if grep -qi "3-Clause" "$file" 2>/dev/null; then
            echo "BSD-3-Clause"
        elif grep -qi "2-Clause" "$file" 2>/dev/null; then
            echo "BSD-2-Clause"
        else
            echo "BSD"
        fi
    elif grep -qi "Mozilla Public License.*Version 2.0" "$file" 2>/dev/null || \
         (grep -qi "Mozilla Public License" "$file" 2>/dev/null && grep -qi "Version 2.0" "$file" 2>/dev/null); then
        echo "MPL-2.0"
    else
        echo "Unknown"
    fi
}

# Scan licenses
scan_licenses() {
    local path="$1"

    log "Scanning licenses in $path"
    load_license_policy

    local license_findings=()
    local license_status="✅ PASS"
    local has_violations=false

    echo "## License Compliance Scan"
    echo ""

    # Detect license files
    log "Detecting license files..."
    local license_files=($(detect_license_files "$path"))

    if [[ ${#license_files[@]} -eq 0 ]]; then
        echo "⚠️ **Warning**: No license files found"
        echo ""
        license_status="⚠️ WARNING"
    else
        echo "### License Files Found"
        echo ""
        for file in "${license_files[@]}"; do
            local rel_path="${file#$path/}"
            local detected_license=$(identify_license_from_text "$file")
            echo "- \`$rel_path\` - **$detected_license**"

            # Check policy
            if license_in_array "$detected_license" "${DENIED_LICENSES[@]}"; then
                echo "  - ❌ **VIOLATION**: License is on denied list"
                has_violations=true
                license_status="❌ FAIL"
                LICENSE_VIOLATIONS+=("$rel_path: $detected_license (denied)")
            elif license_in_array "$detected_license" "${REVIEW_LICENSES[@]}"; then
                echo "  - ⚠️ **REVIEW REQUIRED**: License requires legal review"
                license_status="⚠️ WARNING"
            elif license_in_array "$detected_license" "${ALLOWED_LICENSES[@]}"; then
                echo "  - ✅ Approved license"
            elif [[ "$detected_license" == "Unknown" ]]; then
                echo "  - ⚠️ **Unknown license** - manual review needed"
                license_status="⚠️ WARNING"
            fi
        done
        echo ""
    fi

    # Check package manifests
    echo "### Package Manifest Licenses"
    echo ""

    local manifest_licenses=()

    # npm
    local npm_license=$(detect_npm_license "$path")
    if [[ -n "$npm_license" ]]; then
        echo "- **npm** (package.json): \`$npm_license\`"
        manifest_licenses+=("$npm_license")

        if license_in_array "$npm_license" "${DENIED_LICENSES[@]}"; then
            echo "  - ❌ **VIOLATION**: Denied license"
            has_violations=true
            license_status="❌ FAIL"
            LICENSE_VIOLATIONS+=("package.json: $npm_license (denied)")
        elif license_in_array "$npm_license" "${ALLOWED_LICENSES[@]}"; then
            echo "  - ✅ Approved"
        fi
    fi

    # Cargo
    local cargo_license=$(detect_cargo_license "$path")
    if [[ -n "$cargo_license" ]]; then
        echo "- **Cargo** (Cargo.toml): \`$cargo_license\`"
        manifest_licenses+=("$cargo_license")

        if license_in_array "$cargo_license" "${DENIED_LICENSES[@]}"; then
            echo "  - ❌ **VIOLATION**: Denied license"
            has_violations=true
            license_status="❌ FAIL"
            LICENSE_VIOLATIONS+=("Cargo.toml: $cargo_license (denied)")
        elif license_in_array "$cargo_license" "${ALLOWED_LICENSES[@]}"; then
            echo "  - ✅ Approved"
        fi
    fi

    if [[ ${#manifest_licenses[@]} -eq 0 ]]; then
        echo "*No package manifests with license information found*"
    fi
    echo ""

    # Scan source files for SPDX identifiers (sample)
    echo "### SPDX Identifiers in Source Files (Sample)"
    echo ""

    local spdx_count=0
    local spdx_files=()
    local file_count=0
    local max_files=100

    # Find common source files with SPDX identifiers
    while IFS= read -r -d '' file; do
        ((file_count++))
        [[ $file_count -gt $max_files ]] && break

        local spdx=$(extract_spdx_from_file "$file")
        if [[ -n "$spdx" ]]; then
            spdx_files+=("$file")
            ((spdx_count++))

            if [[ $spdx_count -le 5 ]]; then
                local rel_path="${file#$path/}"
                echo "- \`$rel_path\`: \`$spdx\`"

                if license_in_array "$spdx" "${DENIED_LICENSES[@]}"; then
                    echo "  - ❌ **VIOLATION**: Denied license"
                    has_violations=true
                    license_status="❌ FAIL"
                    LICENSE_VIOLATIONS+=("$rel_path: $spdx (SPDX identifier)")
                fi
            fi
        fi
    done < <(find "$path" -type f \( -name "*.js" -o -name "*.ts" -o -name "*.py" -o -name "*.rs" -o -name "*.go" -o -name "*.sh" -o -name "*.md" \) -print0 2>/dev/null)

    if [[ $spdx_count -gt 5 ]]; then
        echo "- ... and $((spdx_count - 5)) more files with SPDX identifiers"
    elif [[ $spdx_count -eq 0 ]]; then
        echo "*No SPDX identifiers found in source files*"
    fi
    echo ""

    # Summary
    echo "### Summary"
    echo ""
    echo "**Status**: $license_status"
    echo ""
    echo "**Policy Configuration**:"
    echo "- Allowed licenses: ${#ALLOWED_LICENSES[@]} (${ALLOWED_LICENSES[*]})"
    echo "- Denied licenses: ${#DENIED_LICENSES[@]} (${DENIED_LICENSES[*]})"
    echo "- Review required: ${#REVIEW_LICENSES[@]} (${REVIEW_LICENSES[*]})"
    echo ""

    if [[ "$has_violations" == true ]]; then
        echo "**⚠️ Action Required**: Address license violations before distribution"
        echo ""
    fi
}

# Scan secrets
scan_secrets() {
    local path="$1"

    log "Scanning for secrets in $path"

    echo "## Secret Detection"
    echo ""
    echo "ℹ️ Secret detection feature has been moved to the roadmap for future implementation."
    echo ""
    echo "**Planned Features**:"
    echo "- Pattern-based detection (AWS keys, GitHub tokens, private keys)"
    echo "- Entropy-based detection for high-entropy strings"
    echo "- PII detection (SSN, credit cards, emails)"
    echo "- Integration with TruffleHog or GitLeaks"
    echo "- False positive filtering"
    echo ""
    echo "See \`ROADMAP.md\` for timeline and details."
    echo ""
}

# Load content policy from config
load_content_policy() {
    PROFANITY_TERMS=()
    INCLUSIVE_TERMS=()

    if [[ ! -f "$LEGAL_CONFIG" ]]; then
        # Default profanity terms
        PROFANITY_TERMS=("fuck" "shit" "damn" "wtf")
        # Default non-inclusive terms
        INCLUSIVE_TERMS=("master" "slave" "whitelist" "blacklist" "grandfathered" "sanity check")
        return
    fi

    # Parse JSON config for profanity
    if command -v jq &> /dev/null; then
        while IFS= read -r term; do
            [[ -n "$term" ]] && PROFANITY_TERMS+=("$term")
        done < <(jq -r '.legal_review.content_policy.profanity.patterns[]?.term' "$LEGAL_CONFIG" 2>/dev/null)

        # Parse JSON config for inclusive language
        while IFS= read -r term; do
            [[ -n "$term" ]] && INCLUSIVE_TERMS+=("$term")
        done < <(jq -r '.legal_review.content_policy.inclusive_language.replacements[]?.term' "$LEGAL_CONFIG" 2>/dev/null)
    fi
}

# Get alternatives for a term
get_alternatives() {
    local term="$1"
    local type="$2"  # "profanity" or "inclusive_language"

    if [[ ! -f "$LEGAL_CONFIG" ]] || ! command -v jq &> /dev/null; then
        case "$term" in
            "master") echo "primary, main, leader" ;;
            "slave") echo "replica, follower, secondary" ;;
            "whitelist") echo "allowlist, permitted" ;;
            "blacklist") echo "denylist, blocked" ;;
            "fuck") echo "broken, problematic" ;;
            "shit") echo "poor quality, problematic" ;;
            *) echo "N/A" ;;
        esac
        return
    fi

    local alternatives=$(jq -r ".legal_review.content_policy.$type.replacements[] | select(.term == \"$term\") | .alternatives | join(\", \")" "$LEGAL_CONFIG" 2>/dev/null)
    [[ -z "$alternatives" ]] && alternatives=$(jq -r ".legal_review.content_policy.$type.patterns[] | select(.term == \"$term\") | .alternatives | join(\", \")" "$LEGAL_CONFIG" 2>/dev/null)

    echo "${alternatives:-N/A}"
}

# Scan content policy
scan_content_policy() {
    local path="$1"

    log "Scanning content policy in $path"
    load_content_policy

    local content_status="✅ PASS"
    local has_issues=false
    local profanity_count=0
    local inclusive_count=0
    local file_count=0
    local max_files=100

    echo "## Content Policy Scan"
    echo ""

    # Scan for profanity
    echo "### Profanity Detection"
    echo ""

    local profanity_findings=()

    # Build grep pattern for profanity (case-insensitive)
    local profanity_pattern=""
    for term in "${PROFANITY_TERMS[@]}"; do
        if [[ -z "$profanity_pattern" ]]; then
            profanity_pattern="$term"
        else
            profanity_pattern="$profanity_pattern\|$term"
        fi
    done

    if [[ -n "$profanity_pattern" ]]; then
        while IFS= read -r -d '' file; do
            ((file_count++))
            [[ $file_count -gt $max_files ]] && break

            # Skip binary files and large files
            [[ ! -f "$file" ]] && continue
            file -b "$file" | grep -qi "text" || continue

            # Search for profanity terms
            while IFS=: read -r line_num line_content; do
                for term in "${PROFANITY_TERMS[@]}"; do
                    if echo "$line_content" | grep -qi "\b$term\b"; then
                        ((profanity_count++))
                        if [[ $profanity_count -le 10 ]]; then
                            local rel_path="${file#$path/}"
                            local alternatives=$(get_alternatives "$term" "profanity")
                            profanity_findings+=("- \`$rel_path:$line_num\` - **$term** → Alternatives: $alternatives")
                            CONTENT_ISSUES+=("$rel_path:$line_num - profanity: $term")
                            has_issues=true
                            content_status="⚠️ WARNING"
                        fi
                    fi
                done
            done < <(grep -ni "$profanity_pattern" "$file" 2>/dev/null)
        done < <(find "$path" -type f \( -name "*.js" -o -name "*.ts" -o -name "*.py" -o -name "*.rs" -o -name "*.go" -o -name "*.sh" -o -name "*.md" -o -name "*.java" -o -name "*.c" -o -name "*.cpp" \) -print0 2>/dev/null)
    fi

    if [[ ${#profanity_findings[@]} -gt 0 ]]; then
        printf '%s\n' "${profanity_findings[@]}"
        if [[ $profanity_count -gt 10 ]]; then
            echo "- ... and $((profanity_count - 10)) more instances"
        fi
        echo ""
    else
        echo "✅ No profanity detected"
        echo ""
    fi

    # Scan for non-inclusive language
    echo "### Inclusive Language Check"
    echo ""

    local inclusive_findings=()
    file_count=0

    # Build grep pattern for inclusive language terms
    local inclusive_pattern=""
    for term in "${INCLUSIVE_TERMS[@]}"; do
        if [[ -z "$inclusive_pattern" ]]; then
            inclusive_pattern="$term"
        else
            inclusive_pattern="$inclusive_pattern\|$term"
        fi
    done

    if [[ -n "$inclusive_pattern" ]]; then
        while IFS= read -r -d '' file; do
            ((file_count++))
            [[ $file_count -gt $max_files ]] && break

            # Skip binary files
            [[ ! -f "$file" ]] && continue
            file -b "$file" | grep -qi "text" || continue

            # Search for non-inclusive terms
            while IFS=: read -r line_num line_content; do
                for term in "${INCLUSIVE_TERMS[@]}"; do
                    if echo "$line_content" | grep -qi "\b$term\b"; then
                        # Skip if it's in an exemption context (like "git master")
                        if echo "$line_content" | grep -qi "git $term\|IDE $term\|Bluetooth $term"; then
                            continue
                        fi

                        ((inclusive_count++))
                        if [[ $inclusive_count -le 10 ]]; then
                            local rel_path="${file#$path/}"
                            local alternatives=$(get_alternatives "$term" "inclusive_language")
                            inclusive_findings+=("- \`$rel_path:$line_num\` - **$term** → Alternatives: $alternatives")
                            CONTENT_ISSUES+=("$rel_path:$line_num - non-inclusive: $term")
                            has_issues=true
                            content_status="⚠️ WARNING"
                        fi
                    fi
                done
            done < <(grep -ni "$inclusive_pattern" "$file" 2>/dev/null)
        done < <(find "$path" -type f \( -name "*.js" -o -name "*.ts" -o -name "*.py" -o -name "*.rs" -o -name "*.go" -o -name "*.sh" -o -name "*.md" -o -name "*.java" -o -name "*.c" -o -name "*.cpp" \) -print0 2>/dev/null)
    fi

    if [[ ${#inclusive_findings[@]} -gt 0 ]]; then
        printf '%s\n' "${inclusive_findings[@]}"
        if [[ $inclusive_count -gt 10 ]]; then
            echo "- ... and $((inclusive_count - 10)) more instances"
        fi
        echo ""
    else
        echo "✅ All language is inclusive"
        echo ""
    fi

    # Summary
    echo "### Summary"
    echo ""
    echo "**Status**: $content_status"
    echo ""
    echo "**Findings**:"
    echo "- Profanity instances: $profanity_count"
    echo "- Non-inclusive terms: $inclusive_count"
    echo ""

    if [[ "$has_issues" == true ]]; then
        echo "**⚠️ Action Required**: Review and update flagged content"
        echo ""
        echo "**Best Practices**:"
        echo "- Use professional, inclusive language in all code and comments"
        echo "- Replace non-inclusive terms with modern alternatives"
        echo "- Consider audience and context when writing documentation"
        echo ""
    fi
}

# Parallel content policy scanner
scan_content_policy_parallel() {
    local path="$1"
    local jobs="$2"

    log "Scanning content policy in $path (parallel mode with $jobs workers)"
    load_content_policy

    local content_status="✅ PASS"
    local has_issues=false
    local profanity_count=0
    local inclusive_count=0

    echo "## Content Policy Scan (Parallel Mode)"
    echo ""

    # Create temp directory for results
    local results_dir=$(mktemp -d)
    TEMP_DIRS+=("$results_dir")

    # Find all relevant files first
    local file_list="$results_dir/files.txt"
    find "$path" -type f \( \
        -name "*.js" -o -name "*.ts" -o -name "*.py" -o \
        -name "*.rs" -o -name "*.go" -o -name "*.sh" -o \
        -name "*.md" -o -name "*.java" -o -name "*.c" -o -name "*.cpp" \
    \) > "$file_list" 2>/dev/null

    local total_files=$(wc -l < "$file_list")
    log "Found $total_files files to scan"

    # Export variables and functions for parallel execution
    export -f get_alternatives
    export PROFANITY_TERMS
    export INCLUSIVE_TERMS
    export LEGAL_CONFIG
    export REPO_ROOT

    # Build grep patterns
    local profanity_pattern=""
    for term in "${PROFANITY_TERMS[@]}"; do
        if [[ -z "$profanity_pattern" ]]; then
            profanity_pattern="$term"
        else
            profanity_pattern="$profanity_pattern\|$term"
        fi
    done

    local inclusive_pattern=""
    for term in "${INCLUSIVE_TERMS[@]}"; do
        if [[ -z "$inclusive_pattern" ]]; then
            inclusive_pattern="$term"
        else
            inclusive_pattern="$inclusive_pattern\|$term"
        fi
    done

    export profanity_pattern
    export inclusive_pattern
    export path

    # Scan for profanity in parallel
    echo "### Profanity Detection"
    echo ""

    if [[ -n "$profanity_pattern" ]]; then
        cat "$file_list" | xargs -P "$jobs" -I {} bash -c '
            file="$1"
            results_dir="$2"
            profanity_pattern="$3"

            # Skip binary files
            [[ ! -f "$file" ]] && exit 0
            file -b "$file" | grep -qi "text" || exit 0

            # Create result file for this file
            result_file="$results_dir/profanity_$(echo "$file" | md5 -q).txt"

            # Search for profanity terms
            grep -ni "$profanity_pattern" "$file" 2>/dev/null | while IFS=: read -r line_num line_content; do
                # Check each profanity term
                for term in '"${PROFANITY_TERMS[@]}"'; do
                    if echo "$line_content" | grep -qi "\b$term\b"; then
                        rel_path="${file#$path/}"
                        echo "PROFANITY|$rel_path|$line_num|$term" >> "$result_file"
                    fi
                done
            done
        ' bash {} "$results_dir" "$profanity_pattern"

        # Aggregate profanity results
        local profanity_findings=()
        for result_file in "$results_dir"/profanity_*.txt; do
            [[ ! -f "$result_file" ]] && continue
            while IFS='|' read -r type rel_path line_num term; do
                ((profanity_count++))
                if [[ $profanity_count -le 10 ]]; then
                    local alternatives=$(get_alternatives "$term" "profanity")
                    profanity_findings+=("- \`$rel_path:$line_num\` - **$term** → Alternatives: $alternatives")
                    CONTENT_ISSUES+=("$rel_path:$line_num - profanity: $term")
                    has_issues=true
                    content_status="⚠️ WARNING"
                fi
            done < "$result_file"
        done

        if [[ ${#profanity_findings[@]} -gt 0 ]]; then
            printf '%s\n' "${profanity_findings[@]}"
            if [[ $profanity_count -gt 10 ]]; then
                echo "- ... and $((profanity_count - 10)) more instances"
            fi
            echo ""
        else
            echo "✅ No profanity detected"
            echo ""
        fi
    else
        echo "✅ No profanity patterns configured"
        echo ""
    fi

    # Scan for non-inclusive language in parallel
    echo "### Inclusive Language Check"
    echo ""

    if [[ -n "$inclusive_pattern" ]]; then
        # Clean up previous profanity results
        rm -f "$results_dir"/profanity_*.txt

        cat "$file_list" | xargs -P "$jobs" -I {} bash -c '
            file="$1"
            results_dir="$2"
            inclusive_pattern="$3"

            # Skip binary files
            [[ ! -f "$file" ]] && exit 0
            file -b "$file" | grep -qi "text" || exit 0

            # Create result file for this file
            result_file="$results_dir/inclusive_$(echo "$file" | md5 -q).txt"

            # Search for non-inclusive terms
            grep -ni "$inclusive_pattern" "$file" 2>/dev/null | while IFS=: read -r line_num line_content; do
                # Skip exemption contexts
                if echo "$line_content" | grep -qi "git master\|IDE master\|Bluetooth master"; then
                    continue
                fi

                # Check each inclusive term
                for term in '"${INCLUSIVE_TERMS[@]}"'; do
                    if echo "$line_content" | grep -qi "\b$term\b"; then
                        rel_path="${file#$path/}"
                        echo "INCLUSIVE|$rel_path|$line_num|$term" >> "$result_file"
                    fi
                done
            done
        ' bash {} "$results_dir" "$inclusive_pattern"

        # Aggregate inclusive language results
        local inclusive_findings=()
        for result_file in "$results_dir"/inclusive_*.txt; do
            [[ ! -f "$result_file" ]] && continue
            while IFS='|' read -r type rel_path line_num term; do
                ((inclusive_count++))
                if [[ $inclusive_count -le 10 ]]; then
                    local alternatives=$(get_alternatives "$term" "inclusive_language")
                    inclusive_findings+=("- \`$rel_path:$line_num\` - **$term** → Alternatives: $alternatives")
                    CONTENT_ISSUES+=("$rel_path:$line_num - non-inclusive: $term")
                    has_issues=true
                    content_status="⚠️ WARNING"
                fi
            done < "$result_file"
        done

        if [[ ${#inclusive_findings[@]} -gt 0 ]]; then
            printf '%s\n' "${inclusive_findings[@]}"
            if [[ $inclusive_count -gt 10 ]]; then
                echo "- ... and $((inclusive_count - 10)) more instances"
            fi
            echo ""
        else
            echo "✅ All language is inclusive"
            echo ""
        fi
    else
        echo "✅ No inclusive language patterns configured"
        echo ""
    fi

    # Summary
    echo "### Summary"
    echo ""
    echo "**Status**: $content_status"
    echo "**Files Scanned**: $total_files"
    echo ""
    echo "**Findings**:"
    echo "- Profanity instances: $profanity_count"
    echo "- Non-inclusive terms: $inclusive_count"
    echo ""

    if [[ "$has_issues" == true ]]; then
        echo "**⚠️ Action Required**: Review and update flagged content"
        echo ""
        echo "**Best Practices**:"
        echo "- Use professional, inclusive language in all code and comments"
        echo "- Replace non-inclusive terms with modern alternatives"
        echo "- Consider audience and context when writing documentation"
        echo ""
    fi
}

# Load RAG documentation for Claude context
load_rag_context() {
    local context=""

    # Load supply chain best practices (relevant for all analysis)
    if [[ -f "$REPO_ROOT/rag/supply-chain/supply-chain-best-practices.md" ]]; then
        context+="# Supply Chain Best Practices\n\n"
        context+=$(head -300 "$REPO_ROOT/rag/supply-chain/supply-chain-best-practices.md")
        context+="\n\n"
    fi

    # Load license compliance guide
    if [[ -f "$REPO_ROOT/rag/legal-review/license-compliance-guide.md" ]]; then
        context+="# License Compliance Guide\n\n"
        context+=$(head -400 "$REPO_ROOT/rag/legal-review/license-compliance-guide.md")
        context+="\n\n"
    fi

    # Load content policy guide
    if [[ -f "$REPO_ROOT/rag/legal-review/content-policy-guide.md" ]]; then
        context+="# Content Policy Guide\n\n"
        context+=$(head -300 "$REPO_ROOT/rag/legal-review/content-policy-guide.md")
        context+="\n\n"
    fi

    echo -e "$context"
}

# Call Claude API for enhanced analysis
call_claude_api() {
    local prompt="$1"
    local model="${2:-claude-sonnet-4-5-20250929}"

    # Check for API key
    if [[ -z "${ANTHROPIC_API_KEY:-}" ]]; then
        echo "Error: ANTHROPIC_API_KEY environment variable not set" >&2
        return 1
    fi

    log "Calling Claude API with model: $model"

    # Prepare API request
    local request_body=$(jq -n \
        --arg model "$model" \
        --arg prompt "$prompt" \
        '{
            model: $model,
            max_tokens: 4096,
            messages: [{
                role: "user",
                content: $prompt
            }]
        }')

    # Call API
    local response=$(curl -s -X POST https://api.anthropic.com/v1/messages \
        -H "x-api-key: $ANTHROPIC_API_KEY" \
        -H "anthropic-version: 2023-06-01" \
        -H "content-type: application/json" \
        -d "$request_body")

    # Check for API errors
    if echo "$response" | jq -e '.error' > /dev/null 2>&1; then
        local error_type=$(echo "$response" | jq -r '.error.type')
        local error_message=$(echo "$response" | jq -r '.error.message')
        echo "Error: Claude API request failed - $error_type: $error_message" >&2
        return 1
    fi

    # Record API usage for cost tracking
    if command -v record_api_usage &> /dev/null; then
        record_api_usage "$response" "$model" > /dev/null
    fi

    # Extract and return the analysis
    local analysis=$(echo "$response" | jq -r '.content[0].text // empty')

    if [[ -z "$analysis" ]]; then
        echo "Error: No analysis returned from Claude API" >&2
        echo "Response: $response" >&2
        return 1
    fi

    echo "$analysis"
}

# Enhanced license analysis with Claude AI
claude_analyze_licenses() {
    local scan_results="$1"

    log "Enhancing license analysis with Claude AI..."

    # Load RAG context
    local rag_context=$(load_rag_context)

    # Build comprehensive prompt with RAG best practices
    local prompt="Analyze this license compliance data following best practices from our RAG documentation.

${rag_context}

## Analysis Focus Areas:

### 1. License Compatibility & Conflicts
- Identify incompatible license combinations (e.g., GPL + proprietary, GPL + Apache, AGPL + commercial)
- Explain copyleft implications for commercial use and distribution
- Flag viral license risks (AGPL, GPL v3, GPL v2)
- Check weak copyleft compatibility (LGPL, MPL) - safe if dynamic linking, risky if static
- Assess permissive license combinations (MIT, Apache, BSD)

### 2. Risk Assessment by License Category
- **Permissive** (MIT, Apache-2.0, BSD): Generally safe, attribution required, patent grant considerations
- **Weak Copyleft** (LGPL, MPL-2.0): Safe if dynamic linking, risky if static linking or modification
- **Strong Copyleft** (GPL v2/v3): Derivative code must be open-sourced, distribution triggers obligations
- **Network Copyleft** (AGPL-3.0): Triggers on network use, stricter than GPL
- **Proprietary/Custom**: Requires detailed legal review, commercial restrictions likely
- **Unknown/Missing**: Critical compliance risk, blocks deployment

### 3. Compliance Requirements
- Attribution requirements (copyright notices, license file inclusion, NOTICE files)
- Source code disclosure obligations (when triggered, what must be disclosed)
- Patent grant implications (Apache-2.0 patent protection, GPL patent clauses)
- Trademark restrictions (use of project names, logos)
- Redistribution conditions (binary vs source, modifications)
- License file propagation requirements

### 4. Business Impact Assessment
- Commercial use restrictions (can we sell products using this?)
- SaaS deployment implications (AGPL network trigger, GPL distribution)
- Distribution requirements (app stores, customer deployments)
- Derivative work definitions (what modifications trigger copyleft?)
- Sublicensing restrictions
- Vendor lock-in considerations

### 5. Remediation Recommendations
For each violation or high-risk license:
- **Specific action required**: Remove dependency, replace library, add attribution, refactor code
- **Alternative libraries**: Suggest compatible replacements with similar functionality
- **Migration complexity estimate**: Easy (drop-in replacement), Medium (API changes), Hard (architectural changes)
- **Timeline for remediation**: Immediate (blocking), Short-term (1-7 days), Medium-term (1-4 weeks)
- **Cost-benefit analysis**: License compliance value vs engineering effort

## Output Format:

### Executive Summary
- **Overall compliance status**: Pass / Warning / Fail
- **Critical issues requiring immediate attention** (blocks release, legal exposure)
- **Total licenses detected and breakdown by category**:
  - Permissive: X packages
  - Weak Copyleft: X packages
  - Strong Copyleft: X packages
  - Network Copyleft: X packages
  - Unknown/Proprietary: X packages

### License Inventory
Present as markdown table:
| Package | Version | License | Category | Risk Level | Status | Action Required |

### Compatibility Analysis
- **License conflict matrix**: Which combinations are problematic?
- **Copyleft implications**: What code must be open-sourced?
- **Commercial use assessment**: Any restrictions on commercial distribution?

### Prioritized Action Items
1. **Critical** (0-24h): Violations blocking release, immediate legal exposure
2. **High** (1-7d): Significant compliance risks, policy violations
3. **Medium** (1-30d): Attribution fixes, documentation updates, policy clarifications
4. **Low** (30-90d): Best practice improvements, license optimization, future planning

## Scan Results:
${scan_results}

Provide actionable, specific recommendations in markdown format with clear next steps."

    # Call Claude API
    call_claude_api "$prompt"
}

# Enhanced content policy analysis with Claude AI
claude_analyze_content() {
    local scan_results="$1"

    log "Enhancing content policy analysis with Claude AI..."

    # Load RAG context
    local rag_context=$(load_rag_context)

    # Build comprehensive prompt with RAG best practices
    local prompt="Analyze this content policy compliance data with focus on professional standards and inclusive language.

${rag_context}

## Analysis Focus Areas:

### 1. Profanity & Offensive Language
- **Direct profanity** in code, comments, variable names, function names
- **Inappropriate technical metaphors** (e.g., \"kill\", \"abort\" might be OK in context)
- **Offensive variable/function names** that lack professionalism
- **Inappropriate commit messages** visible in git history
- **Context evaluation**: Distinguish technical necessity from offensive usage
- **Cultural sensitivity**: Terms that may be offensive in different cultures

### 2. Non-Inclusive Language
Common patterns and modern alternatives:
- **Master/slave** → primary/replica, leader/follower, main/secondary
- **Whitelist/blacklist** → allowlist/denylist, permitted/blocked
- **Blackhat/whitehat** → malicious/ethical, adversarial/defensive
- **Grandfathered** → legacy, established, existing
- **Sanity check** → validation, verification, consistency check
- **Gendered terms** (e.g., \"guys\", \"man-hours\") → gender-neutral alternatives
- **Cultural idioms** that may not translate well globally

### 3. Business Risk Assessment
- **PR and brand reputation impact**: How would this look in a news article?
- **Team morale and inclusivity**: Does this create an unwelcoming environment?
- **Customer perception risks**: Impact on diverse customer base
- **Legal/HR compliance issues**: Potential harassment or discrimination concerns
- **Recruitment impact**: Does this deter diverse talent?
- **Partner/vendor relationships**: Professional image with stakeholders

### 4. Context-Aware Recommendations
- **Distinguish technical terms from violations**:
  - \"git master\" branch (technical, but consider \"main\")
  - \"master key\" in cryptography (technical term)
  - \"slave device\" in hardware (legacy term, replace if possible)
  - Bluetooth \"master/slave\" (spec terminology, but replaceable)
- **Preserve necessary technical terminology** when no alternative exists
- **Provide context-appropriate alternatives** that maintain clarity
- **Suggest gradual migration paths** for large codebases
- **Consider API stability**: Public APIs require deprecation paths

## Best Practices from RAG:
- Use professional, inclusive language in all code and documentation
- Update terminology proactively during refactoring (not big-bang changes)
- Add linter rules (ESLint, Pylint plugins) to prevent future violations
- Document exceptions with clear justification (e.g., third-party API requirements)
- Include in code review guidelines
- Update style guides to reflect modern standards
- Provide team training on inclusive language

## Output Format:

### Executive Summary
- **Content policy compliance status**: Pass / Warning / Fail
- **Total violations by category**:
  - Profanity/Offensive: X instances
  - Non-Inclusive Language: X instances
- **Immediate actions required**
- **Overall risk level** (reputation, legal, team impact)

### Findings by Category

#### Profanity/Offensive Language
For each issue:
- **Location**: file:line
- **Current term**: The flagged content
- **Context**: Why it's problematic
- **Recommended replacement**: Professional alternative
- **Priority level**: High/Medium/Low based on visibility

#### Non-Inclusive Language
For each issue:
- **Location**: file:line
- **Current term**: The flagged terminology
- **Recommended replacement**: Modern inclusive alternative
- **Context analysis**: Is this technical jargon or problematic usage?
- **Migration path**: Immediate change vs gradual deprecation
- **Priority level**: High (public API/docs), Medium (internal code), Low (private comments)

### Remediation Plan

1. **High Priority** (Immediate - 1-7 days)
   - Offensive language in user-facing code/docs
   - Non-inclusive terms in public APIs
   - Violations visible to customers/partners
   - Code review required before merging

2. **Medium Priority** (1-30 days)
   - Non-inclusive terms in internal code
   - Comments and documentation
   - Variable/function names in private modules
   - Batch updates during planned refactoring

3. **Low Priority** (30-90 days)
   - Historical code, comments (gradual migration)
   - Third-party dependencies (influence upstream)
   - Legacy systems with limited maintenance

### Automation Recommendations
- **Linter configuration**: Specific rules to add (e.g., ESLint no-profanity plugin)
- **Pre-commit hooks**: Block new violations from being committed
- **CI/CD checks**: Fail builds with new violations
- **IDE integration**: Real-time warnings during development

### Team Enablement
- **Style guide updates**: Document inclusive language standards
- **Code review checklist**: Add content policy items
- **Training resources**: Links to inclusive language guides
- **Communication template**: How to discuss with team (non-judgmental, educational)

## Scan Results:
${scan_results}

Provide practical, actionable guidance in markdown format with specific file:line references and clear next steps."

    # Call Claude API
    call_claude_api "$prompt"
}

# Comprehensive Claude AI analysis
claude_enhanced_analysis() {
    local scan_path="$1"

    if [[ "$USE_CLAUDE" != true ]]; then
        return 0
    fi

    echo ""
    echo "## 🤖 Claude AI Enhanced Analysis"
    echo ""

    # Check for API key
    if [[ -z "${ANTHROPIC_API_KEY:-}" ]]; then
        echo "⚠️ **Warning**: ANTHROPIC_API_KEY not set. Skipping AI enhancement."
        echo ""
        echo "Set your API key to enable Claude AI analysis:"
        echo "\`\`\`bash"
        echo "export ANTHROPIC_API_KEY='your-api-key'"
        echo "\`\`\`"
        echo ""
        return 0
    fi

    # Create summary of scan results
    local scan_summary="# Legal Review Scan Results Summary

## License Findings
- Total violations: ${#LICENSE_VIOLATIONS[@]}
- Critical issues: $(echo "${LICENSE_VIOLATIONS[@]}" | grep -c "GPL" || echo "0")

Violations:
$(if [[ ${#LICENSE_VIOLATIONS[@]} -gt 0 ]]; then
    for violation in "${LICENSE_VIOLATIONS[@]}"; do echo "- $violation"; done
else
    echo "- None detected"
fi)

## Content Policy Findings
- Total issues: ${#CONTENT_ISSUES[@]}

Issues:
$(if [[ ${#CONTENT_ISSUES[@]} -gt 0 ]]; then
    for issue in "${CONTENT_ISSUES[@]}"; do echo "- $issue"; done
else
    echo "- None detected"
fi)

## Repository Context
- Path: $scan_path
- Scan types: $(
    local types=""
    [[ "$SCAN_LICENSES" == true ]] && types="${types}licenses, "
    [[ "$SCAN_SECRETS" == true ]] && types="${types}secrets, "
    [[ "$SCAN_CONTENT" == true ]] && types="${types}content policy, "
    echo "${types%, }"
)
"

    # Always run Claude analysis for comprehensive insights
    echo "### 📋 License Compliance Analysis"
    echo ""
    if [[ ${#LICENSE_VIOLATIONS[@]} -gt 0 ]]; then
        echo "**Analysis of detected violations and recommendations:**"
        echo ""
    else
        echo "**No violations detected. Validating best practices:**"
        echo ""
    fi
    local license_analysis=$(claude_analyze_licenses "$scan_summary")
    echo "$license_analysis"
    echo ""

    echo "### ✍️ Content Policy Analysis"
    echo ""
    if [[ ${#CONTENT_ISSUES[@]} -gt 0 ]]; then
        echo "**Analysis of detected issues and remediation:**"
        echo ""
    else
        echo "**No issues detected. Reviewing for best practices:**"
        echo ""
    fi
    local content_analysis=$(claude_analyze_content "$scan_summary")
    echo "$content_analysis"
    echo ""

    echo "**Note**: Analysis includes supply chain best practices from RAG knowledge base"
    echo ""
}

# Format results as table
format_table_output() {
    local target_name="$1"
    local license_violations="${#LICENSE_VIOLATIONS[@]}"
    local content_issues="${#CONTENT_ISSUES[@]}"

    # Calculate overall status
    local overall_status="PASS"
    if [[ $license_violations -gt 0 ]] || [[ $content_issues -gt 0 ]]; then
        overall_status="WARNING"
    fi
    if echo "${LICENSE_VIOLATIONS[@]}" | grep -q "GPL"; then
        overall_status="FAIL"
    fi

    # Summary table
    echo "╔════════════════════════════════════════════════════════════════╗"
    echo "║         Legal Compliance Analysis Summary                      ║"
    echo "╠════════════════════════════════════════════════════════════════╣"
    printf "║ %-30s %-30s ║\n" "Target:" "$target_name"
    printf "║ %-30s %-30s ║\n" "Timestamp:" "$(date -u +"%Y-%m-%d %H:%M:%S UTC")"
    printf "║ %-30s %-30s ║\n" "Overall Status:" "$overall_status"
    echo "╠════════════════════════════════════════════════════════════════╣"
    printf "║ %-30s %-30s ║\n" "License Violations:" "$license_violations"
    printf "║ %-30s %-30s ║\n" "Content Policy Issues:" "$content_issues"
    printf "║ %-30s %-30s ║\n" "Secret Exposures:" "0 (not implemented)"
    echo "╚════════════════════════════════════════════════════════════════╝"
    echo ""

    # Detailed findings
    if [[ $license_violations -gt 0 ]]; then
        echo "License Violations:"
        echo "┌────────────────────────────────────────────────────────────────┐"
        for violation in "${LICENSE_VIOLATIONS[@]}"; do
            printf "│ %-62s │\n" "$violation"
        done
        echo "└────────────────────────────────────────────────────────────────┘"
        echo ""
    fi

    if [[ $content_issues -gt 0 ]]; then
        echo "Content Policy Issues:"
        echo "┌────────────────────────────────────────────────────────────────┐"
        local count=0
        for issue in "${CONTENT_ISSUES[@]}"; do
            printf "│ %-62s │\n" "$issue"
            ((count++))
            [[ $count -ge 20 ]] && break
        done
        if [[ ${#CONTENT_ISSUES[@]} -gt 20 ]]; then
            printf "│ %-62s │\n" "... and $((${#CONTENT_ISSUES[@]} - 20)) more"
        fi
        echo "└────────────────────────────────────────────────────────────────┘"
        echo ""
    fi
}

# Format results as JSON
format_json_output() {
    local target_name="$1"
    local license_violations="${#LICENSE_VIOLATIONS[@]}"
    local content_issues="${#CONTENT_ISSUES[@]}"

    # Calculate overall status
    local overall_status="pass"
    if [[ $license_violations -gt 0 ]] || [[ $content_issues -gt 0 ]]; then
        overall_status="warning"
    fi
    if echo "${LICENSE_VIOLATIONS[@]}" | grep -q "GPL"; then
        overall_status="fail"
    fi

    # Build license violations JSON array
    local license_violations_json="[]"
    if [[ $license_violations -gt 0 ]]; then
        license_violations_json=$(printf '%s\n' "${LICENSE_VIOLATIONS[@]}" | jq -R . | jq -s .)
    fi

    # Build content issues JSON array
    local content_issues_json="[]"
    if [[ $content_issues -gt 0 ]]; then
        content_issues_json=$(printf '%s\n' "${CONTENT_ISSUES[@]}" | jq -R . | jq -s .)
    fi

    # Build complete JSON structure
    jq -n \
        --arg timestamp "$(date -u +"%Y-%m-%dT%H:%M:%SZ")" \
        --arg target "$target_name" \
        --arg scan_licenses "$SCAN_LICENSES" \
        --arg scan_secrets "$SCAN_SECRETS" \
        --arg scan_content "$SCAN_CONTENT" \
        --arg use_claude "$USE_CLAUDE" \
        --arg parallel "$PARALLEL" \
        --arg overall_status "$overall_status" \
        --argjson license_count "$license_violations" \
        --argjson content_count "$content_issues" \
        --argjson license_violations "$license_violations_json" \
        --argjson content_issues "$content_issues_json" \
        '{
            scan_metadata: {
                timestamp: $timestamp,
                target: $target,
                scan_types: [
                    (if $scan_licenses == "true" then "licenses" else empty end),
                    (if $scan_secrets == "true" then "secrets" else empty end),
                    (if $scan_content == "true" then "content_policy" else empty end)
                ],
                analyser_version: "1.0.0",
                analyser_type: (if $use_claude == "true" then "claude" else "basic" end),
                parallel_mode: ($parallel == "true")
            },
            summary: {
                overall_status: $overall_status,
                license_violations: $license_count,
                content_policy_issues: $content_count,
                secret_exposures: 0
            },
            findings: {
                license_violations: $license_violations,
                content_policy_issues: $content_issues,
                secrets: []
            }
        }'
}

# Run analysis on a single repository/path
analyze_single_target() {
    local scan_path="$1"
    local target_name="$2"

    # For table and JSON formats, suppress markdown scan output but show Claude
    local suppress_scan_output=false
    if [[ "$OUTPUT_FORMAT" == "table" ]] || [[ "$OUTPUT_FORMAT" == "json" ]]; then
        suppress_scan_output=true
    fi

    # Run scans (with optional output suppression)
    if [[ "$suppress_scan_output" == true ]]; then
        # Suppress scan output, collect results in arrays
        {
            if [[ "$SCAN_LICENSES" == true ]]; then
                scan_licenses "$scan_path"
            fi

            if [[ "$SCAN_SECRETS" == true ]]; then
                scan_secrets "$scan_path"
            fi

            if [[ "$SCAN_CONTENT" == true ]]; then
                if [[ "$PARALLEL" == true ]]; then
                    scan_content_policy_parallel "$scan_path" "$PARALLEL_JOBS"
                else
                    scan_content_policy "$scan_path"
                fi
            fi
        } > /dev/null 2>&1
    else
        # Normal markdown output
        echo "# Legal Review Analysis Report"
        echo ""
        echo "**Generated**: $(date -u +"%Y-%m-%dT%H:%M:%SZ")"
        echo "**Target**: ${target_name}"
        echo ""
        if [[ "$PARALLEL" == true ]]; then
            echo "**Mode**: Parallel processing ($PARALLEL_JOBS workers)"
            echo ""
        fi

        if [[ "$SCAN_LICENSES" == true ]]; then
            scan_licenses "$scan_path"
        fi

        if [[ "$SCAN_SECRETS" == true ]]; then
            scan_secrets "$scan_path"
        fi

        if [[ "$SCAN_CONTENT" == true ]]; then
            if [[ "$PARALLEL" == true ]]; then
                scan_content_policy_parallel "$scan_path" "$PARALLEL_JOBS"
            else
                scan_content_policy "$scan_path"
            fi
        fi
    fi

    # Claude AI enhanced analysis (always run if enabled, regardless of format)
    if [[ "$COMPARE_MODE" != true ]] && [[ "$USE_CLAUDE" == true ]]; then
        claude_enhanced_analysis "$scan_path"
    fi

    # Output formatted results
    if [[ "$OUTPUT_FORMAT" == "table" ]]; then
        format_table_output "$target_name"
    elif [[ "$OUTPUT_FORMAT" == "json" ]]; then
        format_json_output "$target_name"
    fi
}

# Save report to reports directory with sequence number
save_report() {
    local content="$1"
    local target_name="$2"
    local format="$3"

    # Skip if auto-save is disabled or output file is already specified
    if [[ "$AUTO_SAVE_REPORTS" != true ]] || [[ -n "$OUTPUT_FILE" ]]; then
        return 0
    fi

    # Create reports directory if it doesn't exist
    mkdir -p "$REPORTS_DIR"

    # Get next sequence number
    local sequence=1
    if [[ -f "$REPORTS_DIR/.sequence" ]]; then
        sequence=$(cat "$REPORTS_DIR/.sequence")
        ((sequence++))
    fi
    echo "$sequence" > "$REPORTS_DIR/.sequence"

    # Format timestamp
    local timestamp=$(date +"%Y%m%d-%H%M%S")

    # Sanitize target name for filename
    local sanitized_target=$(echo "$target_name" | tr '/' '_' | tr ':' '_')

    # Determine scan types
    local scan_types=""
    [[ "$SCAN_LICENSES" == true ]] && scan_types="${scan_types}L"
    [[ "$SCAN_SECRETS" == true ]] && scan_types="${scan_types}S"
    [[ "$SCAN_CONTENT" == true ]] && scan_types="${scan_types}C"
    [[ "$USE_CLAUDE" == true ]] && scan_types="${scan_types}+AI"

    # Determine file extension
    local extension="md"
    case "$format" in
        json) extension="json" ;;
        table) extension="txt" ;;
        markdown) extension="md" ;;
    esac

    # Build filename: {sequence}_{scan_types}_{target}_{timestamp}.{ext}
    local filename="${sequence}_${scan_types}_${sanitized_target}_${timestamp}.${extension}"
    local filepath="$REPORTS_DIR/$filename"

    # Save report
    echo "$content" > "$filepath"

    log "Report saved to: $filepath"
    echo "" >&2
    echo -e "${GREEN}✓ Report saved: reports/legal-review/$filename${NC}" >&2
    echo "" >&2
}

# Main analysis
main() {
    parse_args "$@"

    # Validate target options
    if [[ -z "$TARGET_REPO" ]] && [[ -z "$TARGET_PATH" ]] && [[ -z "$TARGET_ORG" ]] && [[ -z "$LOCAL_PATH" ]]; then
        echo "Error: Must specify one of: --repo, --org, --path, or --local-path"
        usage
    fi

    # Load Claude cost tracking if using Claude
    if [[ "$USE_CLAUDE" == "true" ]] || [[ "$COMPARE_MODE" == "true" ]]; then
        if [ -f "$REPO_ROOT/utils/lib/claude-cost.sh" ]; then
            source "$REPO_ROOT/utils/lib/claude-cost.sh"
            init_cost_tracking
        fi
    fi

    load_config

    # Determine scan path and target name
    local scan_path=""
    local target_name=""

    # Priority: LOCAL_PATH > TARGET_PATH > TARGET_REPO > TARGET_ORG
    if [[ -n "$LOCAL_PATH" ]]; then
        # Use pre-cloned repository
        if [[ ! -d "$LOCAL_PATH" ]]; then
            echo "Error: Local path does not exist: $LOCAL_PATH" >&2
            exit 1
        fi
        scan_path="$LOCAL_PATH"
        target_name="$LOCAL_PATH"
        log "Using pre-cloned repository at $LOCAL_PATH"

    elif [[ -n "$TARGET_PATH" ]]; then
        # Use local path
        if [[ ! -d "$TARGET_PATH" ]]; then
            echo "Error: Path does not exist: $TARGET_PATH" >&2
            exit 1
        fi
        scan_path="$TARGET_PATH"
        target_name="$TARGET_PATH"

    elif [[ -n "$TARGET_REPO" ]]; then
        # Clone single repository
        log "Cloning repository $TARGET_REPO"
        local temp_dir=$(mktemp -d)
        TEMP_DIRS+=("$temp_dir")

        local repo_dir="$temp_dir/repo"
        if ! github_clone_repository "$TARGET_REPO" "$repo_dir" --depth 1; then
            echo "Error: Failed to clone $TARGET_REPO" >&2
            exit 1
        fi
        echo ""  # Add blank line after clone output

        scan_path="$repo_dir"
        target_name="$TARGET_REPO"

    elif [[ -n "$TARGET_ORG" ]]; then
        # Analyze all repositories in organization
        echo "# Legal Review Analysis Report - Organization"
        echo ""
        echo "**Generated**: $(date -u +"%Y-%m-%dT%H:%M:%SZ")"
        echo "**Organization**: ${TARGET_ORG}"
        echo ""

        log "Fetching repositories for organization: $TARGET_ORG"

        # Get all repos in org using github.sh library
        local repos=($(github_list_org_repos "$TARGET_ORG"))

        if [[ ${#repos[@]} -eq 0 ]]; then
            echo "Error: No repositories found for organization $TARGET_ORG" >&2
            exit 1
        fi

        echo "Found ${#repos[@]} repositories in organization $TARGET_ORG"
        echo ""

        # Analyze each repository
        for repo in "${repos[@]}"; do
            echo "---"
            echo ""
            echo "## Repository: $repo"
            echo ""

            local temp_dir=$(mktemp -d)
            TEMP_DIRS+=("$temp_dir")

            local repo_dir="$temp_dir/repo"
            if github_clone_repository "$TARGET_ORG/$repo" "$repo_dir" --depth 1; then
                analyze_single_target "$repo_dir" "$TARGET_ORG/$repo"
            else
                echo "⚠️ **Warning**: Failed to clone $TARGET_ORG/$repo - skipping"
            fi
            echo ""
        done

        # Display cost summary if using Claude
        if command -v display_api_cost_summary &> /dev/null; then
            display_api_cost_summary
        fi

        return 0
    fi

    # Single target analysis - capture output for saving
    local report_content=""

    if [[ "$COMPARE_MODE" == true ]]; then
        # Run comparison mode: basic vs Claude side-by-side
        report_content=$(
            echo "# Legal Review Analysis - Comparison Mode"
            echo ""
            echo "**Generated**: $(date -u +"%Y-%m-%dT%H:%M:%SZ")"
            echo "**Target**: ${target_name}"
            echo ""
            echo "---"
            echo ""
            echo "## Basic Analysis (Without Claude AI)"
            echo ""

            # Temporarily disable Claude for basic analysis
            local original_use_claude="$USE_CLAUDE"
            USE_CLAUDE=false
            analyze_single_target "$scan_path" "$target_name"
            USE_CLAUDE="$original_use_claude"

            echo ""
            echo "---"
            echo ""
            echo "## Claude AI Enhanced Analysis"
            echo ""

            # Reset violation tracking for Claude analysis
            LICENSE_VIOLATIONS=()
            CONTENT_ISSUES=()

            analyze_single_target "$scan_path" "$target_name"
            claude_enhanced_analysis "$scan_path"
        )
    else
        # Normal mode - capture output
        report_content=$(analyze_single_target "$scan_path" "$target_name")
    fi

    # Display the report
    echo "$report_content"

    # Save report to file
    save_report "$report_content" "$target_name" "$OUTPUT_FORMAT"

    # Display cost summary if using Claude
    if command -v display_api_cost_summary &> /dev/null; then
        display_api_cost_summary
    fi
}

main "$@"
