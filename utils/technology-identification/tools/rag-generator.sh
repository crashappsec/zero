#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# RAG Pattern Generator
# Automated tool to generate technology detection patterns from various sources
# Supports: npm, PyPI, RubyGems, Maven, crates.io
#############################################################################

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
UTILS_ROOT="$(dirname "$SCRIPT_DIR")"
REPO_ROOT="$(dirname "$(dirname "$UTILS_ROOT")")"
RAG_ROOT="$REPO_ROOT/rag/technology-identification"

# Command modes
MODE=""
TECHNOLOGY=""
PACKAGE=""
ECOSYSTEM="npm"
CATEGORY=""
DOCS_URL=""
OUTPUT_DIR=""
BATCH_FILE=""

#############################################################################
# Helper Functions
#############################################################################

usage() {
    cat << EOF
RAG Pattern Generator - Automate technology pattern creation

USAGE:
    $0 from-registry --technology NAME --package PKG --ecosystem ECOSYSTEM --category CAT
    $0 from-docs --technology NAME --docs-url URL --category CAT
    $0 batch --input FILE

MODES:
    from-registry   Generate patterns from package registry (npm, PyPI, etc.)
    from-docs       Generate patterns from documentation (coming soon)
    batch           Batch generate from CSV file

OPTIONS:
    --technology NAME       Technology name (e.g., vue, django)
    --package PKG           Package name in registry
    --ecosystem ECOSYSTEM   Package ecosystem: npm, pypi, rubygems, maven, cargo
    --category CAT          Category path (e.g., web-frameworks/frontend)
    --docs-url URL          Official documentation URL
    --input FILE            CSV file for batch processing
    --output DIR            Output directory (default: rag/technology-identification/)

EXAMPLES:
    # Generate Vue.js patterns from npm
    $0 from-registry \\
        --technology vue \\
        --package vue \\
        --ecosystem npm \\
        --category web-frameworks/frontend

    # Generate Django patterns from PyPI
    $0 from-registry \\
        --technology django \\
        --package Django \\
        --ecosystem pypi \\
        --category web-frameworks/backend

    # Batch generate from CSV
    $0 batch --input technologies.csv

EOF
    exit 1
}

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

#############################################################################
# Registry Fetchers
#############################################################################

fetch_npm_data() {
    local package="$1"
    local url="https://registry.npmjs.org/$package"

    log_info "Fetching npm data for: $package"

    local data=$(curl -s "$url" 2>/dev/null)

    if [[ -z "$data" ]] || echo "$data" | grep -q "error"; then
        log_error "Failed to fetch npm data for $package"
        return 1
    fi

    echo "$data"
}

fetch_pypi_data() {
    local package="$1"
    local url="https://pypi.org/pypi/$package/json"

    log_info "Fetching PyPI data for: $package"

    local data=$(curl -s "$url" 2>/dev/null)

    if [[ -z "$data" ]] || echo "$data" | grep -q "error"; then
        log_error "Failed to fetch PyPI data for $package"
        return 1
    fi

    echo "$data"
}

#############################################################################
# Pattern Generators
#############################################################################

generate_package_patterns_npm() {
    local tech_name="$1"
    local package="$2"
    local category="$3"
    local npm_data="$4"

    local description=$(echo "$npm_data" | jq -r '.description // "No description available"' 2>/dev/null)
    local homepage=$(echo "$npm_data" | jq -r '.homepage // ""' 2>/dev/null)

    cat << EOF
{
  "technology": "$tech_name",
  "category": "$category",
  "description": "$description",
  "confidence": 95,
  "patterns": [
    {
      "ecosystem": "npm",
      "names": ["$package"],
      "scoped_names": [],
      "description": "Core $tech_name package"
    },
    {
      "ecosystem": "yarn",
      "names": ["$package"],
      "description": "Core $tech_name package via Yarn"
    },
    {
      "ecosystem": "pnpm",
      "names": ["$package"],
      "description": "Core $tech_name package via pnpm"
    }
  ],
  "related_packages": [],
  "official_registry": "https://www.npmjs.com/package/$package",
  "homepage": "$homepage",
  "detection_notes": [
    "Check for $package in dependencies or devDependencies",
    "Look for import/require statements",
    "Verify with configuration files"
  ]
}
EOF
}

generate_package_patterns_pypi() {
    local tech_name="$1"
    local package="$2"
    local category="$3"
    local pypi_data="$4"

    local description=$(echo "$pypi_data" | jq -r '.info.summary // "No description available"' 2>/dev/null)
    local homepage=$(echo "$pypi_data" | jq -r '.info.home_page // ""' 2>/dev/null)

    cat << EOF
{
  "technology": "$tech_name",
  "category": "$category",
  "description": "$description",
  "confidence": 95,
  "patterns": [
    {
      "ecosystem": "pypi",
      "names": ["$package"],
      "description": "Core $tech_name package"
    },
    {
      "ecosystem": "pip",
      "names": ["$package"],
      "description": "Install via pip"
    },
    {
      "ecosystem": "poetry",
      "names": ["$package"],
      "description": "Install via Poetry"
    }
  ],
  "related_packages": [],
  "official_registry": "https://pypi.org/project/$package/",
  "homepage": "$homepage",
  "detection_notes": [
    "Check for $package in requirements.txt, setup.py, or pyproject.toml",
    "Look for import statements in Python files",
    "Verify with configuration files"
  ]
}
EOF
}

generate_import_patterns_generic() {
    local tech_name="$1"
    local package="$2"
    local category="$3"

    cat << EOF
{
  "technology": "$tech_name",
  "category": "$category",
  "confidence": 85,
  "patterns": [
    {
      "language": "javascript",
      "file_extensions": [".js", ".mjs"],
      "patterns": [
        {
          "regex": "import\\\\s+.*\\\\s+from\\\\s+['\"]$package['\"]",
          "description": "ES6 import from $package",
          "example": "import $tech_name from '$package';"
        },
        {
          "regex": "require\\\\(['\"]$package['\"]\\\\)",
          "description": "CommonJS require",
          "example": "const $tech_name = require('$package');"
        }
      ]
    }
  ],
  "common_imports": [],
  "detection_notes": [
    "Look for $package imports in JavaScript/TypeScript files",
    "Check both ES6 and CommonJS patterns"
  ]
}
EOF
}

generate_import_patterns_python() {
    local tech_name="$1"
    local package="$2"
    local category="$3"

    cat << EOF
{
  "technology": "$tech_name",
  "category": "$category",
  "confidence": 85,
  "patterns": [
    {
      "language": "python",
      "file_extensions": [".py"],
      "patterns": [
        {
          "regex": "^import\\\\s+$package",
          "description": "Direct import",
          "example": "import $package"
        },
        {
          "regex": "^from\\\\s+$package\\\\s+import",
          "description": "From import",
          "example": "from $package import something"
        }
      ]
    }
  ],
  "common_imports": [],
  "detection_notes": [
    "Look for $package imports in Python files",
    "Check both 'import' and 'from...import' patterns"
  ]
}
EOF
}

generate_config_patterns_generic() {
    local tech_name="$1"
    local category="$2"

    cat << EOF
{
  "technology": "$tech_name",
  "category": "$category",
  "confidence": 80,
  "patterns": [],
  "directory_conventions": [],
  "detection_notes": [
    "Look for $tech_name-specific configuration files",
    "Check for common directory structures",
    "Verify with package.json or requirements.txt"
  ]
}
EOF
}

generate_env_patterns_generic() {
    local tech_name="$1"
    local category="$2"

    cat << EOF
{
  "technology": "$tech_name",
  "category": "$category",
  "confidence": 60,
  "patterns": [],
  "detection_notes": [
    "Look for $tech_name-related environment variables",
    "Check .env files and environment configuration"
  ]
}
EOF
}

generate_api_patterns_generic() {
    local tech_name="$1"
    local category="$2"

    cat << EOF
{
  "technology": "$tech_name",
  "category": "$category",
  "confidence": 50,
  "patterns": [],
  "notes": [
    "$tech_name may not have distinctive API endpoint patterns",
    "Check for CDN URLs or hosted service endpoints"
  ]
}
EOF
}

generate_versions_from_npm() {
    local tech_name="$1"
    local package="$2"
    local npm_data="$3"

    local current_version=$(echo "$npm_data" | jq -r '."dist-tags".latest // "unknown"' 2>/dev/null)
    local homepage=$(echo "$npm_data" | jq -r '.homepage // ""' 2>/dev/null)
    local repo=$(echo "$npm_data" | jq -r '.repository.url // ""' 2>/dev/null)

    # Get last 3 versions
    local versions=$(echo "$npm_data" | jq -r '.versions | keys_unsorted | .[-3:][]' 2>/dev/null)

    cat << EOF
{
  "technology": "$tech_name",
  "category": "auto-generated",
  "current_stable": "$current_version",
  "release_history": [
EOF

    local first=true
    while IFS= read -r version; do
        if [[ -n "$version" ]]; then
            local release_date=$(echo "$npm_data" | jq -r ".time[\"$version\"] // \"unknown\"" 2>/dev/null)

            if [[ "$first" == true ]]; then
                first=false
            else
                echo ","
            fi

            cat << INNER_EOF
    {
      "version": "$version",
      "release_date": "$release_date",
      "status": "stable",
      "eol_date": null,
      "features": [],
      "breaking_changes": []
    }
INNER_EOF
        fi
    done <<< "$versions"

    cat << EOF

  ],
  "version_detection": {
    "package_json": "$package"
  },
  "official_resources": {
    "npm": "https://www.npmjs.com/package/$package",
    "homepage": "$homepage",
    "repository": "$repo"
  }
}
EOF
}

generate_versions_from_pypi() {
    local tech_name="$1"
    local package="$2"
    local pypi_data="$3"

    local current_version=$(echo "$pypi_data" | jq -r '.info.version // "unknown"' 2>/dev/null)
    local homepage=$(echo "$pypi_data" | jq -r '.info.home_page // ""' 2>/dev/null)
    local repo=$(echo "$pypi_data" | jq -r '.info.project_urls.Repository // ""' 2>/dev/null)

    cat << EOF
{
  "technology": "$tech_name",
  "category": "auto-generated",
  "current_stable": "$current_version",
  "release_history": [
    {
      "version": "$current_version",
      "release_date": "unknown",
      "status": "stable",
      "eol_date": null,
      "features": [],
      "breaking_changes": []
    }
  ],
  "version_detection": {
    "requirements_txt": "$package"
  },
  "official_resources": {
    "pypi": "https://pypi.org/project/$package/",
    "homepage": "$homepage",
    "repository": "$repo"
  }
}
EOF
}

#############################################################################
# Main Generation Logic
#############################################################################

generate_from_registry() {
    log_info "Generating patterns for: $TECHNOLOGY"
    log_info "Package: $PACKAGE"
    log_info "Ecosystem: $ECOSYSTEM"
    log_info "Category: $CATEGORY"

    # Create output directory
    local tech_dir="$RAG_ROOT/$CATEGORY/$TECHNOLOGY"
    mkdir -p "$tech_dir"

    log_info "Output directory: $tech_dir"

    # Fetch data based on ecosystem
    local registry_data=""

    case "$ECOSYSTEM" in
        npm)
            registry_data=$(fetch_npm_data "$PACKAGE")
            if [[ $? -ne 0 ]]; then
                log_error "Failed to fetch npm data"
                return 1
            fi

            # Generate patterns
            generate_package_patterns_npm "$TECHNOLOGY" "$PACKAGE" "$CATEGORY" "$registry_data" \
                > "$tech_dir/package-patterns.json"
            log_success "Created package-patterns.json"

            generate_import_patterns_generic "$TECHNOLOGY" "$PACKAGE" "$CATEGORY" \
                > "$tech_dir/import-patterns.json"
            log_success "Created import-patterns.json"

            generate_versions_from_npm "$TECHNOLOGY" "$PACKAGE" "$registry_data" \
                > "$tech_dir/versions.json"
            log_success "Created versions.json"
            ;;

        pypi)
            registry_data=$(fetch_pypi_data "$PACKAGE")
            if [[ $? -ne 0 ]]; then
                log_error "Failed to fetch PyPI data"
                return 1
            fi

            # Generate patterns
            generate_package_patterns_pypi "$TECHNOLOGY" "$PACKAGE" "$CATEGORY" "$registry_data" \
                > "$tech_dir/package-patterns.json"
            log_success "Created package-patterns.json"

            generate_import_patterns_python "$TECHNOLOGY" "$PACKAGE" "$CATEGORY" \
                > "$tech_dir/import-patterns.json"
            log_success "Created import-patterns.json"

            generate_versions_from_pypi "$TECHNOLOGY" "$PACKAGE" "$registry_data" \
                > "$tech_dir/versions.json"
            log_success "Created versions.json"
            ;;

        *)
            log_error "Unsupported ecosystem: $ECOSYSTEM"
            return 1
            ;;
    esac

    # Generate generic patterns
    generate_config_patterns_generic "$TECHNOLOGY" "$CATEGORY" \
        > "$tech_dir/config-patterns.json"
    log_success "Created config-patterns.json"

    generate_env_patterns_generic "$TECHNOLOGY" "$CATEGORY" \
        > "$tech_dir/env-patterns.json"
    log_success "Created env-patterns.json"

    generate_api_patterns_generic "$TECHNOLOGY" "$CATEGORY" \
        > "$tech_dir/api-patterns.json"
    log_success "Created api-patterns.json"

    log_success "✓ Generated all 6 pattern files for $TECHNOLOGY"
    echo ""
    echo "Next steps:"
    echo "1. Review and enhance patterns in: $tech_dir"
    echo "2. Add related packages to package-patterns.json"
    echo "3. Add specific import patterns to import-patterns.json"
    echo "4. Add configuration file patterns to config-patterns.json"
    echo "5. Test detection with: ./technology-identification-analyser.sh"
}

#############################################################################
# Parse Arguments
#############################################################################

if [[ $# -eq 0 ]]; then
    usage
fi

MODE="$1"
shift

while [[ $# -gt 0 ]]; do
    case $1 in
        --technology)
            TECHNOLOGY="$2"
            shift 2
            ;;
        --package)
            PACKAGE="$2"
            shift 2
            ;;
        --ecosystem)
            ECOSYSTEM="$2"
            shift 2
            ;;
        --category)
            CATEGORY="$2"
            shift 2
            ;;
        --docs-url)
            DOCS_URL="$2"
            shift 2
            ;;
        --input)
            BATCH_FILE="$2"
            shift 2
            ;;
        --output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        -h|--help)
            usage
            ;;
        *)
            log_error "Unknown option: $1"
            usage
            ;;
    esac
done

#############################################################################
# Execute
#############################################################################

echo ""
echo "========================================="
echo "  RAG Pattern Generator"
echo "========================================="
echo ""

case "$MODE" in
    from-registry)
        if [[ -z "$TECHNOLOGY" ]] || [[ -z "$PACKAGE" ]] || [[ -z "$CATEGORY" ]]; then
            log_error "Missing required arguments"
            usage
        fi

        generate_from_registry
        ;;

    from-docs)
        log_warn "Documentation-based generation not yet implemented"
        log_info "Use from-registry mode for now"
        exit 1
        ;;

    batch)
        log_warn "Batch generation not yet implemented"
        log_info "Use from-registry mode for now"
        exit 1
        ;;

    *)
        log_error "Unknown mode: $MODE"
        usage
        ;;
esac

echo ""
log_success "Pattern generation complete!"
