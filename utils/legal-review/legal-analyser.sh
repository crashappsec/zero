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

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Temp directories tracking
TEMP_DIRS=()

# Cleanup function
cleanup() {
    if [[ ${#TEMP_DIRS[@]} -gt 0 ]]; then
        for temp_dir in "${TEMP_DIRS[@]}"; do
            if [[ -n "$temp_dir" ]] && [[ -d "$temp_dir" ]]; then
                rm -rf "$temp_dir"
            fi
        done
    fi
}

trap cleanup EXIT

# Usage
usage() {
    cat <<EOF
Legal Review Analyser - Comprehensive code legal compliance scanner

Usage: $0 [OPTIONS]

OPTIONS:
    --repo OWNER/REPO          Analyze GitHub repository
    --path PATH                Analyze local path
    --licenses-only            Scan licenses only
    --secrets-only             Scan secrets only
    --content-only             Scan content policy only
    --format FORMAT            Output format: markdown (default), json
    --output FILE              Write output to file
    --claude                   Use Claude AI for enhanced analysis
    --verbose                  Enable verbose output
    -h, --help                 Show this help message

EXAMPLES:
    # Full analysis
    $0 --repo owner/repo

    # License scan only
    $0 --repo owner/repo --licenses-only

    # Local path with JSON output
    $0 --path /path/to/code --format json --output report.json

    # Claude AI enhanced
    $0 --repo owner/repo --claude

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
            --path)
                TARGET_PATH="$2"
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

    printf '%s\n' "${license_files[@]}"
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

    echo "## Secret Scan"
    echo ""
    echo "⏳ Secret detection implementation pending"
    echo ""
    echo "**TODO**: Implement secret detection using:"
    echo "- Pattern-based detection (AWS keys, GitHub tokens, etc.)"
    echo "- Entropy-based detection for random strings"
    echo "- PII detection (SSN, credit cards, etc.)"
    echo "- Integration with TruffleHog or GitLeaks"
    echo ""
    echo "See \`prompts/legal-review/BUILD-LEGAL-ANALYSER.md\` for implementation details."
    echo ""
}

# Scan content policy
scan_content_policy() {
    local path="$1"

    log "Scanning content policy in $path"

    echo "## Content Policy Scan"
    echo ""
    echo "⏳ Content policy scanning implementation pending"
    echo ""
    echo "**TODO**: Implement content policy checks for:"
    echo "- Profanity in identifiers and comments"
    echo "- Non-inclusive language (master/slave, whitelist/blacklist, etc.)"
    echo "- Hate speech detection"
    echo "- Integration with woke or alex"
    echo ""
    echo "See \`prompts/legal-review/BUILD-LEGAL-ANALYSER.md\` for implementation details."
    echo ""
}

# Main analysis
main() {
    parse_args "$@"

    if [[ -z "$TARGET_REPO" ]] && [[ -z "$TARGET_PATH" ]]; then
        echo "Error: Must specify --repo or --path"
        usage
    fi

    load_config

    echo "# Legal Review Analysis Report"
    echo ""
    echo "**Generated**: $(date -u +"%Y-%m-%dT%H:%M:%SZ")"
    echo "**Target**: ${TARGET_REPO:-${TARGET_PATH}}"
    echo ""

    # Determine scan path
    local scan_path="$TARGET_PATH"

    if [[ -n "$TARGET_REPO" ]]; then
        log "Cloning repository $TARGET_REPO"
        local temp_dir=$(mktemp -d)
        TEMP_DIRS+=("$temp_dir")

        local clone_url="https://github.com/$TARGET_REPO"
        git clone --depth 1 --quiet "$clone_url" "$temp_dir/repo" 2>/dev/null || {
            echo "Error: Failed to clone $TARGET_REPO" >&2
            exit 1
        }

        scan_path="$temp_dir/repo"
    fi

    # Run scans
    if [[ "$SCAN_LICENSES" == true ]]; then
        scan_licenses "$scan_path"
    fi

    if [[ "$SCAN_SECRETS" == true ]]; then
        scan_secrets "$scan_path"
    fi

    if [[ "$SCAN_CONTENT" == true ]]; then
        scan_content_policy "$scan_path"
    fi

    echo "## Implementation Status"
    echo ""
    echo "✅ Legal review framework complete:"
    echo "- RAG documentation: 4 comprehensive guides"
    echo "- Configuration: \`config/legal-review-config.json\`"
    echo "- Skill: \`skills/legal-review/\`"
    echo "- Build prompt: \`prompts/legal-review/BUILD-LEGAL-ANALYSER.md\`"
    echo ""
    echo "⏳ Analyser implementation: In progress"
    echo ""
    echo "**Next Steps**:"
    echo "1. Review \`prompts/legal-review/BUILD-LEGAL-ANALYSER.md\`"
    echo "2. Implement license scanning (Phase 1)"
    echo "3. Implement secret detection (Phase 2)"
    echo "4. Implement content policy (Phase 3)"
    echo "5. Add Claude AI integration (Phase 4)"
    echo ""
    echo "**Use Claude Code to complete implementation**:"
    echo "\`\`\`bash"
    echo "# In Claude Code, use the build prompt to implement the analyser"
    echo "@legal-review implement the analyser using BUILD-LEGAL-ANALYSER.md"
    echo "\`\`\`"
    echo ""
}

main "$@"
