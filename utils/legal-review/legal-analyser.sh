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

# Scan licenses
scan_licenses() {
    local path="$1"

    log "Scanning licenses in $path"

    echo "## License Scan"
    echo ""
    echo "⏳ License scanning implementation pending"
    echo ""
    echo "**TODO**: Implement license detection using:"
    echo "- SPDX identifier detection in file headers"
    echo "- License file detection (LICENSE, COPYING, etc.)"
    echo "- Package manifest parsing (package.json, pom.xml, etc.)"
    echo "- Integration with ScanCode or Licensee"
    echo ""
    echo "See \`prompts/legal-review/BUILD-LEGAL-ANALYSER.md\` for implementation details."
    echo ""
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
