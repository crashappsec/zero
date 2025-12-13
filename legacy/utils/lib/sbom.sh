#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# SBOM Generation Library
# Comprehensive Software Bill of Materials generation with lock file support
#############################################################################

# Load global config
if [ -f "${REPO_ROOT:-}/utils/lib/config.sh" ]; then
    source "${REPO_ROOT:-}/utils/lib/config.sh"
fi

# Default SBOM configuration
SBOM_FORMAT="${SBOM_FORMAT:-cyclonedx-json}"
SBOM_USE_LOCK_FILES="${SBOM_USE_LOCK_FILES:-true}"
SBOM_INCLUDE_DEV_DEPS="${SBOM_INCLUDE_DEV_DEPS:-false}"
SBOM_INCLUDE_TEST_DEPS="${SBOM_INCLUDE_TEST_DEPS:-false}"
SBOM_SCAN_NODE_MODULES="${SBOM_SCAN_NODE_MODULES:-false}"
SBOM_RESOLVE_VERSIONS="${SBOM_RESOLVE_VERSIONS:-true}"
SBOM_OUTPUT_DIR="${SBOM_OUTPUT_DIR:-sbom-output}"
SBOM_GENERATOR="${SBOM_GENERATOR:-cdxgen}"  # cdxgen (default), syft, or auto

# Load SBOM configuration if available
load_sbom_config() {
    local config_file="${1:-${REPO_ROOT}/utils/scanners/package-sbom/config/sbom-config.json}"

    if [[ -f "$config_file" ]]; then
        # Parse JSON config
        SBOM_FORMAT=$(jq -r '.sbom.format // "cyclonedx-json"' "$config_file" 2>/dev/null)
        SBOM_USE_LOCK_FILES=$(jq -r '.sbom.use_lock_files // true' "$config_file" 2>/dev/null)
        SBOM_INCLUDE_DEV_DEPS=$(jq -r '.sbom.include_dev_deps // false' "$config_file" 2>/dev/null)
        SBOM_INCLUDE_TEST_DEPS=$(jq -r '.sbom.include_test_deps // false' "$config_file" 2>/dev/null)
        SBOM_SCAN_NODE_MODULES=$(jq -r '.sbom.scan_node_modules // false' "$config_file" 2>/dev/null)
        SBOM_RESOLVE_VERSIONS=$(jq -r '.sbom.resolve_versions // true' "$config_file" 2>/dev/null)

        return 0
    fi
    return 1
}

# Detect package manager and lock files
detect_package_manager() {
    local repo_path="$1"
    local detected=""

    cd "$repo_path" || return 1

    # JavaScript/Node.js
    if [[ -f "package.json" ]]; then
        if [[ -f "pnpm-lock.yaml" ]]; then
            detected="pnpm"
        elif [[ -f "yarn.lock" ]]; then
            detected="yarn"
        elif [[ -f "package-lock.json" ]]; then
            detected="npm"
        elif [[ -f "bun.lockb" ]]; then
            detected="bun"
        else
            detected="npm"  # Default to npm if no lock file
        fi
    # Python
    elif [[ -f "pyproject.toml" ]] && [[ -f "poetry.lock" ]]; then
        detected="poetry"
    elif [[ -f "Pipfile" ]] && [[ -f "Pipfile.lock" ]]; then
        detected="pipenv"
    elif [[ -f "requirements.txt" ]]; then
        detected="pip"
    # Rust
    elif [[ -f "Cargo.toml" ]]; then
        detected="cargo"
    # Go
    elif [[ -f "go.mod" ]]; then
        detected="go"
    # Ruby
    elif [[ -f "Gemfile" ]]; then
        detected="bundler"
    # Java
    elif [[ -f "pom.xml" ]]; then
        detected="maven"
    elif [[ -f "build.gradle" ]] || [[ -f "build.gradle.kts" ]]; then
        detected="gradle"
    # PHP
    elif [[ -f "composer.json" ]]; then
        detected="composer"
    fi

    echo "$detected"
}

# Get lock file for package manager
get_lock_file() {
    local package_manager="$1"
    local repo_path="${2:-.}"

    case "$package_manager" in
        npm)
            echo "$repo_path/package-lock.json"
            ;;
        yarn)
            echo "$repo_path/yarn.lock"
            ;;
        pnpm)
            echo "$repo_path/pnpm-lock.yaml"
            ;;
        bun)
            echo "$repo_path/bun.lockb"
            ;;
        poetry)
            echo "$repo_path/poetry.lock"
            ;;
        pipenv)
            echo "$repo_path/Pipfile.lock"
            ;;
        pip)
            echo "$repo_path/requirements.txt"
            ;;
        cargo)
            echo "$repo_path/Cargo.lock"
            ;;
        go)
            echo "$repo_path/go.sum"
            ;;
        bundler)
            echo "$repo_path/Gemfile.lock"
            ;;
        maven)
            echo "$repo_path/pom.xml"
            ;;
        gradle)
            if [[ -f "$repo_path/gradle.lockfile" ]]; then
                echo "$repo_path/gradle.lockfile"
            else
                echo "$repo_path/build.gradle"
            fi
            ;;
        composer)
            echo "$repo_path/composer.lock"
            ;;
        *)
            echo ""
            ;;
    esac
}

# Build syft command with appropriate options
build_syft_command() {
    local repo_path="$1"
    local output_file="$2"
    local package_manager="$3"

    local syft_cmd="syft scan"
    local syft_opts=()

    # Cataloger selection based on package manager and config
    if [[ "$SBOM_USE_LOCK_FILES" == "true" ]]; then
        case "$package_manager" in
            npm|yarn|pnpm|bun)
                syft_opts+=("--catalogers" "javascript-lock")
                ;;
            poetry|pipenv)
                syft_opts+=("--catalogers" "python-lock")
                ;;
            cargo)
                syft_opts+=("--catalogers" "rust-cargo-lock")
                ;;
            go)
                syft_opts+=("--catalogers" "go-module-binary,go-mod-file")
                ;;
            bundler)
                syft_opts+=("--catalogers" "ruby-gemfile-lock")
                ;;
            *)
                # Use all catalogers for package manager
                ;;
        esac
    fi

    # Exclude dev dependencies if configured
    if [[ "$SBOM_INCLUDE_DEV_DEPS" != "true" ]]; then
        case "$package_manager" in
            npm|yarn|pnpm)
                syft_opts+=("--exclude" "devDependencies")
                ;;
            poetry)
                syft_opts+=("--exclude" "dev-dependencies")
                ;;
        esac
    fi

    # Exclude test dependencies if configured
    if [[ "$SBOM_INCLUDE_TEST_DEPS" != "true" ]]; then
        syft_opts+=("--exclude" "**/test/**")
        syft_opts+=("--exclude" "**/tests/**")
    fi

    # Node modules scanning
    if [[ "$SBOM_SCAN_NODE_MODULES" != "true" ]] && [[ "$package_manager" =~ ^(npm|yarn|pnpm|bun)$ ]]; then
        syft_opts+=("--exclude" "**/node_modules")
    fi

    # Build final command
    # Note: Use --output FORMAT=PATH format (new syft syntax)
    echo "$syft_cmd" "${syft_opts[@]}" "-o" "$SBOM_FORMAT=$output_file" "dir:$repo_path"
}

# Generate SBOM for repository
generate_sbom() {
    local repo_path="$1"
    local output_file="${2:-sbom.json}"
    local force="${3:-false}"

    # Check if syft is installed
    if ! command -v syft &> /dev/null; then
        echo "Error: syft is not installed" >&2
        echo "Install: brew install syft" >&2
        return 1
    fi

    # Create output directory if needed
    mkdir -p "$(dirname "$output_file")"

    # Check if SBOM already exists
    if [[ -f "$output_file" ]] && [[ "$force" != "true" ]]; then
        echo "SBOM already exists: $output_file" >&2
        echo "Use force=true to regenerate" >&2
        return 0
    fi

    # Detect package manager
    local package_manager=$(detect_package_manager "$repo_path")

    if [[ -z "$package_manager" ]]; then
        echo "Warning: Could not detect package manager, using default catalogers" >&2
        package_manager="unknown"
    else
        echo "Detected package manager: $package_manager" >&2
    fi

    # Check for lock file
    local lock_file=$(get_lock_file "$package_manager" "$repo_path")

    if [[ -n "$lock_file" ]] && [[ -f "$lock_file" ]]; then
        echo "Found lock file: $lock_file" >&2
    else
        echo "Warning: No lock file found for $package_manager" >&2
        if [[ "$SBOM_USE_LOCK_FILES" == "true" ]]; then
            echo "Recommendation: Generate lock file for accurate dependency resolution" >&2
        fi
    fi

    # Extract repository name and version for syft to avoid warnings
    local repo_name=$(basename "$repo_path")
    local repo_version="latest"

    # Try to get version from git if it's a git repository
    if [[ -d "$repo_path/.git" ]]; then
        repo_version=$(cd "$repo_path" && git describe --tags --always 2>/dev/null || echo "latest")
    fi

    # Build syft command directly without using build_syft_command
    local syft_opts=()

    # Add source name and version to avoid warning
    syft_opts+=("--source-name" "$repo_name")
    syft_opts+=("--source-version" "$repo_version")

    # Exclude test dependencies if configured
    if [[ "$SBOM_INCLUDE_TEST_DEPS" != "true" ]]; then
        syft_opts+=("--exclude" "**/test/**")
        syft_opts+=("--exclude" "**/tests/**")
    fi

    # Node modules scanning
    if [[ "$SBOM_SCAN_NODE_MODULES" != "true" ]] && [[ "$package_manager" =~ ^(npm|yarn|pnpm|bun)$ ]]; then
        syft_opts+=("--exclude" "**/node_modules")
    fi

    echo "Generating SBOM..." >&2

    if syft scan "${syft_opts[@]}" -o "$SBOM_FORMAT=$output_file" "dir:$repo_path" 2>&1; then
        echo "SBOM generated successfully" >&2
        return 0
    else
        echo "Error: SBOM generation failed" >&2
        return 1
    fi
}

# Generate SBOM using cdxgen (more accurate, installs dependencies)
generate_sbom_cdxgen() {
    local repo_path="$1"
    local output_file="${2:-sbom.json}"
    local force="${3:-false}"

    # Check if cdxgen is installed
    if ! command -v cdxgen &> /dev/null; then
        echo "Error: cdxgen is not installed" >&2
        echo "Install: npm install -g @cyclonedx/cdxgen" >&2
        return 1
    fi

    # Create output directory if needed
    mkdir -p "$(dirname "$output_file")"

    # Check if SBOM already exists
    if [[ -f "$output_file" ]] && [[ "$force" != "true" ]]; then
        echo "SBOM already exists: $output_file" >&2
        echo "Use force=true to regenerate" >&2
        return 0
    fi

    echo "Generating SBOM with cdxgen (this may take longer but is more accurate)..." >&2

    # cdxgen options
    local cdxgen_opts=()
    cdxgen_opts+=("-o" "$output_file")
    cdxgen_opts+=("--spec-version" "1.5")

    # Use deep mode for better accuracy
    cdxgen_opts+=("--deep")

    # Exclude test directories
    if [[ "$SBOM_INCLUDE_TEST_DEPS" != "true" ]]; then
        cdxgen_opts+=("--exclude" "**/test/**,**/tests/**,**/__tests__/**")
    fi

    if cdxgen "${cdxgen_opts[@]}" "$repo_path" 2>&1; then
        echo "SBOM generated successfully with cdxgen" >&2
        return 0
    else
        echo "Error: cdxgen SBOM generation failed" >&2
        return 1
    fi
}

# Check if project is complex (multiple package managers = benefits from cdxgen)
is_complex_project() {
    local repo_path="$1"
    local manifest_count=0

    # Count different package manager manifests
    [[ -f "$repo_path/package.json" ]] && ((manifest_count++))
    [[ -f "$repo_path/requirements.txt" || -f "$repo_path/pyproject.toml" ]] && ((manifest_count++))
    [[ -f "$repo_path/go.mod" ]] && ((manifest_count++))
    [[ -f "$repo_path/Cargo.toml" ]] && ((manifest_count++))
    [[ -f "$repo_path/pom.xml" || -f "$repo_path/build.gradle" ]] && ((manifest_count++))
    [[ -f "$repo_path/Gemfile" ]] && ((manifest_count++))
    [[ -f "$repo_path/composer.json" ]] && ((manifest_count++))

    # Complex if 2+ different ecosystems
    [[ $manifest_count -ge 2 ]]
}

# Smart SBOM generation - selects best generator based on config/project
# Usage: generate_sbom_smart <repo_path> <output_file> [generator] [force]
# generator: "syft" (default, fast), "cdxgen" (accurate), "auto" (smart selection)
generate_sbom_smart() {
    local repo_path="$1"
    local output_file="${2:-sbom.json}"
    local generator="${3:-$SBOM_GENERATOR}"
    local force="${4:-false}"

    case "$generator" in
        cdxgen)
            if command -v cdxgen &> /dev/null; then
                generate_sbom_cdxgen "$repo_path" "$output_file" "$force"
            else
                echo "Warning: cdxgen not installed, falling back to syft" >&2
                generate_sbom "$repo_path" "$output_file" "$force"
            fi
            ;;
        auto)
            # Use cdxgen if available AND project is complex
            if command -v cdxgen &> /dev/null && is_complex_project "$repo_path"; then
                echo "Complex project detected, using cdxgen for better accuracy" >&2
                generate_sbom_cdxgen "$repo_path" "$output_file" "$force"
            else
                generate_sbom "$repo_path" "$output_file" "$force"
            fi
            ;;
        syft|*)
            generate_sbom "$repo_path" "$output_file" "$force"
            ;;
    esac
}

# Validate SBOM against lock file
validate_sbom_against_lock() {
    local sbom_file="$1"
    local lock_file="$2"
    local package_manager="$3"

    if [[ ! -f "$sbom_file" ]]; then
        echo "Error: SBOM file not found: $sbom_file" >&2
        return 1
    fi

    if [[ ! -f "$lock_file" ]]; then
        echo "Error: Lock file not found: $lock_file" >&2
        return 1
    fi

    # Extract packages from SBOM
    local sbom_packages=$(jq -r '.components[]? | "\(.name)@\(.version)"' "$sbom_file" 2>/dev/null | sort)

    if [[ -z "$sbom_packages" ]]; then
        echo "Error: Could not extract packages from SBOM" >&2
        return 1
    fi

    echo "SBOM contains $(echo "$sbom_packages" | wc -l | tr -d ' ') packages" >&2

    # Package manager specific validation
    case "$package_manager" in
        npm|yarn|pnpm)
            # Extract from package-lock.json or yarn.lock
            local lock_packages=$(jq -r '.packages | to_entries[] | "\(.key)"' "$lock_file" 2>/dev/null | grep -v "^$" | sort)
            ;;
        *)
            echo "Validation not yet implemented for: $package_manager" >&2
            return 0
            ;;
    esac

    # Compare
    local missing=$(comm -13 <(echo "$sbom_packages") <(echo "$lock_packages") | head -10)

    if [[ -n "$missing" ]]; then
        echo "Warning: Packages in lock file but not in SBOM:" >&2
        echo "$missing" >&2
    fi

    return 0
}

# Get SBOM statistics
get_sbom_stats() {
    local sbom_file="$1"

    if [[ ! -f "$sbom_file" ]]; then
        echo "Error: SBOM file not found: $sbom_file" >&2
        return 1
    fi

    local format=$(jq -r '.bomFormat // "unknown"' "$sbom_file" 2>/dev/null)
    local spec_version=$(jq -r '.specVersion // "unknown"' "$sbom_file" 2>/dev/null)
    local component_count=$(jq '.components | length' "$sbom_file" 2>/dev/null)
    local dependency_count=$(jq '.dependencies | length' "$sbom_file" 2>/dev/null)

    cat << EOF
SBOM Statistics:
  Format: $format
  Spec Version: $spec_version
  Components: $component_count
  Dependencies: $dependency_count
EOF
}

# Export functions
export -f load_sbom_config
export -f detect_package_manager
export -f get_lock_file
export -f build_syft_command
export -f generate_sbom
export -f generate_sbom_cdxgen
export -f generate_sbom_smart
export -f is_complex_project
export -f validate_sbom_against_lock
export -f get_sbom_stats
