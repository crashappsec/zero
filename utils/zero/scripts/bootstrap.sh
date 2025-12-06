#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Phantom Bootstrap
# Clone a repository and run all analyzers to prepare for agent queries
#
# Usage: ./bootstrap.sh <target> [options]
#
# Examples:
#   ./bootstrap.sh https://github.com/expressjs/express
#   ./bootstrap.sh expressjs/express
#   ./bootstrap.sh expressjs/express --branch v5.x
#   ./bootstrap.sh expressjs/express --quick
#############################################################################

set -e

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ZERO_DIR="$(dirname "$SCRIPT_DIR")"

# Load Phantom library
source "$ZERO_DIR/lib/zero-lib.sh"

# Load shared config if available
UTILS_ROOT="$(dirname "$ZERO_DIR")"
REPO_ROOT="$(dirname "$UTILS_ROOT")"

if [[ -f "$UTILS_ROOT/lib/config.sh" ]]; then
    source "$UTILS_ROOT/lib/config.sh"
fi

# Load config loader for dynamic profiles
source "$ZERO_DIR/config/config-loader.sh"

# Load .env if it exists
if [[ -f "$REPO_ROOT/.env" ]]; then
    set -a
    source "$REPO_ROOT/.env"
    set +a
fi

#############################################################################
# Configuration
#############################################################################

TARGET=""
BRANCH=""
DEPTH=""
MODE="$(get_default_profile)"  # Load default from config
FORCE=false
ENRICH=false      # Incremental enrichment - only run missing collectors
CLONE_ONLY=false  # Just clone, don't scan
SCAN_ONLY=false   # Just scan existing clone, don't clone

# Canonical list of ALL scanners - loaded from config for display
# This ensures consistent output format across all profiles
ALL_SCANNERS="$(get_all_scanners)"

#############################################################################
# Usage
#############################################################################

usage() {
    cat << EOF
Phantom Hydrate - Clone and analyze a repository for agent queries

Usage: $0 <target> [options]

TARGETS:
    GitHub URL:      https://github.com/owner/repo
    GitHub shorthand: owner/repo
    Local path:       /path/to/project or ./project

ANALYSIS MODES (from zero.config.json):
EOF
    print_profile_help
    cat << EOF

OPTIONS:
    --branch <name>     Clone specific branch (default: default branch)
    --depth <n>         Shallow clone depth (default: full for DORA metrics)
    --force             Re-hydrate even if project exists
    --enrich            Only run collectors not previously run (incremental)
    --clone-only        Clone repository without scanning
    --scan-only         Scan existing clone without cloning
    -h, --help          Show this help

EXAMPLES:
    $0 expressjs/express                    # Standard analysis
    $0 owner/repo --quick                   # Fast scan
    $0 owner/repo --advanced                # Include package-health, provenance
    $0 owner/repo --deep                    # Claude-enhanced analysis
    $0 ./local-project --security           # Security-focused scan

FLOW:
    1. Clone repository to ~/.zero/projects/<id>/repo/
    2. Run analyzers and store JSON in ~/.zero/projects/<id>/analysis/
    3. Set as active project for agent queries

EOF
    exit 0
}

#############################################################################
# Argument Parsing
#############################################################################

parse_args() {
    # Get available profiles for dynamic matching
    local available_profiles=$(get_available_profiles)

    while [[ $# -gt 0 ]]; do
        case "$1" in
            -h|--help)
                usage
                ;;
            --branch)
                BRANCH="$2"
                shift 2
                ;;
            --depth)
                DEPTH="$2"
                shift 2
                ;;
            # Keep legacy aliases for backwards compatibility
            --thorough)
                MODE="advanced"
                shift
                ;;
            --ai-only)
                MODE="deep"
                export USE_CLAUDE=true
                shift
                ;;
            --security-only)
                MODE="security"
                shift
                ;;
            --security-deep)
                MODE="security-deep"
                export USE_CLAUDE=true
                shift
                ;;
            --collectors)
                MODE="custom"
                # Convert comma-separated to space-separated
                export CUSTOM_COLLECTORS="${2//,/ }"
                shift 2
                ;;
            --force)
                FORCE=true
                shift
                ;;
            --enrich)
                ENRICH=true
                shift
                ;;
            --clone-only)
                CLONE_ONLY=true
                shift
                ;;
            --scan-only)
                SCAN_ONLY=true
                shift
                ;;
            --*)
                # Try to match as a dynamic profile name from config
                local profile_name="${1#--}"
                if [[ " $available_profiles " =~ " $profile_name " ]]; then
                    MODE="$profile_name"
                    # Enable Claude if profile requires it
                    if profile_uses_claude "$profile_name"; then
                        export USE_CLAUDE=true
                    fi
                    shift
                else
                    echo -e "${RED}Error: Unknown option $1${NC}" >&2
                    echo -e "Available profiles: $available_profiles"
                    exit 1
                fi
                ;;
            -*)
                echo -e "${RED}Error: Unknown option $1${NC}" >&2
                exit 1
                ;;
            *)
                if [[ -z "$TARGET" ]]; then
                    TARGET="$1"
                else
                    echo -e "${RED}Error: Multiple targets specified${NC}" >&2
                    exit 1
                fi
                shift
                ;;
        esac
    done

    if [[ -z "$TARGET" ]]; then
        echo -e "${RED}Error: No target specified${NC}" >&2
        echo "Usage: $0 <target> [options]"
        exit 1
    fi
}

#############################################################################
# Clone Functions
#############################################################################

clone_github_repo() {
    local url="$1"
    local dest="$2"
    local branch="$3"
    local depth="$4"

    local clone_args=()

    if [[ -n "$branch" ]]; then
        clone_args+=("--branch" "$branch")
    fi

    if [[ -n "$depth" ]]; then
        clone_args+=("--depth" "$depth")
    fi

    # Use GITHUB_TOKEN if available for private repos
    if [[ -n "${GITHUB_TOKEN:-}" ]]; then
        # Convert HTTPS URL to use token
        if echo "$url" | grep -q '^https://github\.com/'; then
            url=$(echo "$url" | sed "s|https://github.com|https://${GITHUB_TOKEN}@github.com|")
        fi
    fi

    # Clone with progress display
    clone_args+=("--progress")

    # Run git clone with performance optimizations:
    # - checkout.workers=0: Use all CPU cores for parallel file checkout (huge speedup)
    # - GIT_PROGRESS_DELAY=0: Show progress immediately
    # Git uses \r for in-place updates, so we need to handle that
    GIT_PROGRESS_DELAY=0 git clone -c checkout.workers=0 "${clone_args[@]}" "$url" "$dest" 2>&1 | tr '\r' '\n' | while IFS= read -r line; do
        # Parse git progress lines like "Receiving objects:  45% (1234/2742), 12.50 MiB | 5.20 MiB/s"
        if echo "$line" | grep -qE 'Receiving objects:.*[0-9]+ [KMG]iB'; then
            # Extract percentage, size received, and speed
            local pct=$(echo "$line" | grep -oE '[0-9]+%' | head -1)
            local size=$(echo "$line" | grep -oE '[0-9]+(\.[0-9]+)? [KMG]iB' | head -1)
            local speed=$(echo "$line" | grep -oE '[0-9]+(\.[0-9]+)? [KMG]iB/s' | head -1)
            if [[ -n "$pct" ]] && [[ -n "$size" ]]; then
                printf "\r    Receiving: %s (%s @ %s)   " "$pct" "$size" "${speed:-...}"
            fi
        elif echo "$line" | grep -qE 'Receiving objects:.*%'; then
            # Progress without size (small repos)
            local pct=$(echo "$line" | grep -oE '[0-9]+%' | head -1)
            if [[ -n "$pct" ]]; then
                printf "\r    Receiving: %s              " "$pct"
            fi
        elif echo "$line" | grep -qE 'Resolving deltas:.*%'; then
            local pct=$(echo "$line" | grep -oE '[0-9]+%' | head -1)
            printf "\r    Resolving: %s              " "$pct"
        elif echo "$line" | grep -qE 'Cloning into'; then
            : # Skip this line
        elif echo "$line" | grep -qE 'Counting|Compressing'; then
            : # Skip these phases
        elif [[ -n "$line" ]]; then
            # Pass through other non-empty output
            echo "$line"
        fi
    done
    # Clear the progress line
    printf "\r                                              \r"
}

copy_local_project() {
    local source="$1"
    local dest="$2"

    # Resolve absolute path
    local abs_source=$(cd "$source" && pwd)

    # Copy entire directory
    cp -R "$abs_source" "$dest"

    # If it's a git repo, ensure we have the .git directory
    if [[ -d "$abs_source/.git" ]]; then
        echo "  Git repository detected"
    else
        echo "  Not a git repository (DORA metrics unavailable)"
    fi
}

get_git_info() {
    local repo_path="$1"

    if [[ ! -d "$repo_path/.git" ]]; then
        echo ""
        echo ""
        return
    fi

    local branch=$(git -C "$repo_path" rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
    local commit=$(git -C "$repo_path" rev-parse --short HEAD 2>/dev/null || echo "unknown")

    echo "$branch"
    echo "$commit"
}

#############################################################################
# Project Detection
#############################################################################

detect_project_type() {
    local repo_path="$1"

    local languages=()
    local frameworks=()
    local package_managers=()

    # Detect by files present
    if [[ -f "$repo_path/package.json" ]]; then
        package_managers+=("npm")
        languages+=("javascript")

        # Check for specific frameworks in package.json
        if grep -q '"react"' "$repo_path/package.json" 2>/dev/null; then
            frameworks+=("react")
        fi
        if grep -q '"express"' "$repo_path/package.json" 2>/dev/null; then
            frameworks+=("express")
        fi
        if grep -q '"next"' "$repo_path/package.json" 2>/dev/null; then
            frameworks+=("nextjs")
        fi
        if grep -q '"vue"' "$repo_path/package.json" 2>/dev/null; then
            frameworks+=("vue")
        fi
        if grep -q '"typescript"' "$repo_path/package.json" 2>/dev/null; then
            languages+=("typescript")
        fi
    fi

    if [[ -f "$repo_path/requirements.txt" ]] || [[ -f "$repo_path/setup.py" ]] || [[ -f "$repo_path/pyproject.toml" ]]; then
        package_managers+=("pip")
        languages+=("python")

        if grep -q 'django' "$repo_path/requirements.txt" 2>/dev/null; then
            frameworks+=("django")
        fi
        if grep -q 'flask' "$repo_path/requirements.txt" 2>/dev/null; then
            frameworks+=("flask")
        fi
        if grep -q 'fastapi' "$repo_path/requirements.txt" 2>/dev/null; then
            frameworks+=("fastapi")
        fi
    fi

    if [[ -f "$repo_path/go.mod" ]]; then
        package_managers+=("go")
        languages+=("go")
    fi

    if [[ -f "$repo_path/Cargo.toml" ]]; then
        package_managers+=("cargo")
        languages+=("rust")
    fi

    if [[ -f "$repo_path/pom.xml" ]] || [[ -f "$repo_path/build.gradle" ]]; then
        if [[ -f "$repo_path/pom.xml" ]]; then
            package_managers+=("maven")
        fi
        if [[ -f "$repo_path/build.gradle" ]]; then
            package_managers+=("gradle")
        fi
        languages+=("java")

        if grep -q 'spring' "$repo_path/pom.xml" 2>/dev/null || grep -q 'spring' "$repo_path/build.gradle" 2>/dev/null; then
            frameworks+=("spring")
        fi
    fi

    if [[ -f "$repo_path/Gemfile" ]]; then
        package_managers+=("bundler")
        languages+=("ruby")

        if grep -q 'rails' "$repo_path/Gemfile" 2>/dev/null; then
            frameworks+=("rails")
        fi
    fi

    if [[ -f "$repo_path/composer.json" ]]; then
        package_managers+=("composer")
        languages+=("php")

        if grep -q 'laravel' "$repo_path/composer.json" 2>/dev/null; then
            frameworks+=("laravel")
        fi
    fi

    # Convert to JSON arrays (handle empty arrays, compact output)
    local langs_json="[]"
    local fwks_json="[]"
    local pkgs_json="[]"

    if [[ ${#languages[@]} -gt 0 ]]; then
        langs_json=$(printf '%s\n' "${languages[@]}" | jq -R . | jq -sc 'unique')
    fi
    if [[ ${#frameworks[@]} -gt 0 ]]; then
        fwks_json=$(printf '%s\n' "${frameworks[@]}" | jq -R . | jq -sc 'unique')
    fi
    if [[ ${#package_managers[@]} -gt 0 ]]; then
        pkgs_json=$(printf '%s\n' "${package_managers[@]}" | jq -R . | jq -sc 'unique')
    fi

    echo "$langs_json"
    echo "$fwks_json"
    echo "$pkgs_json"
}

#############################################################################
# Analyzer Execution
#############################################################################

# Get list of analyzers to run based on mode
# Note: dependencies (SBOM) must run first as other analyzers use sbom.cdx.json
# Profiles are loaded dynamically from zero.config.json
get_analyzers_for_mode() {
    local mode="$1"

    # Special case: custom mode uses environment variable
    if [[ "$mode" == "custom" ]]; then
        echo "${CUSTOM_COLLECTORS:-package-sbom tech-discovery package-vulns licenses}"
        return
    fi

    # Use dynamic profile from config (with fallback)
    local scanners=$(get_profile_scanners "$mode")
    if [[ -n "$scanners" ]]; then
        echo "$scanners"
    else
        # Fallback to standard if profile not found
        get_profile_scanners "standard"
    fi
}

# Get analyzers that have already been run (completed successfully)
get_completed_analyzers() {
    local analysis_path="$1"
    local manifest="$analysis_path/manifest.json"

    if [[ ! -f "$manifest" ]]; then
        echo ""
        return
    fi

    # Get analyzers with status "complete"
    jq -r '.analyses // {} | to_entries[] | select(.value.status == "complete") | .key' "$manifest" 2>/dev/null | tr '\n' ' '
}

# Filter out already-completed analyzers from requested list
get_missing_analyzers() {
    local requested="$1"
    local completed="$2"

    local missing=""
    for analyzer in $requested; do
        if [[ ! " $completed " =~ " $analyzer " ]]; then
            [[ -n "$missing" ]] && missing+=" "
            missing+="$analyzer"
        fi
    done
    echo "$missing"
}

# Check if a scanner is included in a given scanner list (local helper)
# Note: This is different from config-loader's scanner_in_profile which takes profile name
scanner_in_list() {
    local scanner="$1"
    local scanner_list="$2"
    [[ " $scanner_list " =~ " $scanner " ]]
}

# Get display name for a scanner (uses dynamic config)
get_scanner_display_name() {
    local scanner="$1"
    get_scanner_name "$scanner"
}

# Run a single analyzer
run_analyzer() {
    local analyzer="$1"
    local repo_path="$2"
    local output_path="$3"
    local project_id="$4"

    local start_time=$(date +%s)
    local status="complete"
    local summary="null"

    # Record start - map scanner ID to script name for logging
    local analyzer_script=""
    case "$analyzer" in
        tech-discovery)
            analyzer_script="tech-discovery.sh"
            ;;
        package-sbom)
            analyzer_script="package-sbom.sh"
            ;;
        package-vulns)
            analyzer_script="package-vulns.sh"
            ;;
        package-health)
            analyzer_script="package-health.sh"
            ;;
        licenses)
            analyzer_script="licenses.sh"
            ;;
        code-security)
            analyzer_script="code-security.sh"
            ;;
        iac-security)
            analyzer_script="iac-security.sh"
            ;;
        tech-debt)
            analyzer_script="tech-debt.sh"
            ;;
        documentation)
            analyzer_script="documentation.sh"
            ;;
        git)
            analyzer_script="git.sh"
            ;;
        test-coverage)
            analyzer_script="test-coverage.sh"
            ;;
        code-secrets)
            analyzer_script="code-secrets.sh"
            ;;
        code-ownership)
            analyzer_script="code-ownership.sh"
            ;;
        bus-factor)
            analyzer_script="bus-factor.sh"
            ;;
        dora)
            analyzer_script="dora.sh"
            ;;
        package-provenance)
            analyzer_script="package-provenance.sh"
            ;;
        containers)
            analyzer_script="containers.sh"
            ;;
        chalk)
            analyzer_script="chalk.sh"
            ;;
        digital-certificates)
            analyzer_script="digital-certificates.sh"
            ;;
        package-malcontent)
            analyzer_script="package-malcontent/package-malcontent.sh"
            ;;
    esac

    gibson_analysis_start "$project_id" "$analyzer" "$analyzer_script"

    # Run the scanner
    case "$analyzer" in
        tech-discovery)
            run_technology_analyzer "$repo_path" "$output_path"
            ;;
        package-sbom)
            run_dependency_extractor "$repo_path" "$output_path"
            ;;
        package-vulns)
            run_vulnerability_analyzer "$repo_path" "$output_path"
            ;;
        package-health)
            run_package_health_analyzer "$repo_path" "$output_path"
            ;;
        licenses)
            run_license_analyzer "$repo_path" "$output_path"
            ;;
        code-security)
            run_code_security_analyzer "$repo_path" "$output_path"
            ;;
        iac-security)
            run_iac_security_analyzer "$repo_path" "$output_path"
            ;;
        tech-debt)
            run_tech_debt_analyzer "$repo_path" "$output_path"
            ;;
        documentation)
            run_documentation_analyzer "$repo_path" "$output_path"
            ;;
        git)
            run_git_insights_analyzer "$repo_path" "$output_path"
            ;;
        test-coverage)
            run_test_coverage_analyzer "$repo_path" "$output_path"
            ;;
        code-secrets)
            run_secrets_scanner "$repo_path" "$output_path"
            ;;
        code-ownership)
            run_ownership_analyzer "$repo_path" "$output_path"
            ;;
        bus-factor)
            run_bus_factor_analyzer "$repo_path" "$output_path"
            ;;
        dora)
            run_dora_analyzer "$repo_path" "$output_path"
            ;;
        package-provenance)
            run_provenance_analyzer "$repo_path" "$output_path"
            ;;
        containers)
            run_container_analyzer "$repo_path" "$output_path"
            ;;
        chalk)
            run_chalk_analyzer "$repo_path" "$output_path"
            ;;
        digital-certificates)
            run_certificate_analyzer "$repo_path" "$output_path"
            ;;
        package-malcontent)
            run_malcontent_analyzer "$repo_path" "$output_path" "$project_id"
            ;;
        *)
            status="failed"
            ;;
    esac

    local exit_code=$?
    local end_time=$(date +%s)
    local duration_sec=$((end_time - start_time))
    local duration=$((duration_sec * 1000))  # Convert to ms for manifest

    if [[ $exit_code -ne 0 ]]; then
        status="failed"
    fi

    # Extract summary from output if available
    local output_file="$output_path/${analyzer}.json"
    if [[ -f "$output_file" ]] && [[ "$status" == "complete" ]]; then
        summary=$(jq '.summary // null' "$output_file" 2>/dev/null || echo "null")
    fi

    gibson_analysis_complete "$project_id" "$analyzer" "$status" "$duration" "$summary"

    return $exit_code
}

#############################################################################
# Individual Analyzer Wrappers
#############################################################################

run_technology_analyzer() {
    local repo_path="$1"
    local output_path="$2"

    # Use Claude-enabled analyzer in deep mode, otherwise data-only
    local tech_script=""
    if [[ "${USE_CLAUDE:-}" == "true" ]]; then
        tech_script="$UTILS_ROOT/scanners/tech-discovery/tech-discovery.sh"
    else
        tech_script="$UTILS_ROOT/scanners/tech-discovery/tech-discovery.sh"
    fi

    if [[ -x "$tech_script" ]]; then
        # Pass existing SBOM if available (generated by dependencies analyzer)
        local sbom_arg=""
        if [[ -f "$output_path/sbom.cdx.json" ]]; then
            sbom_arg="--sbom $output_path/sbom.cdx.json"
        fi
        local claude_arg=""
        [[ "${USE_CLAUDE:-}" == "true" ]] && claude_arg="--claude"
        "$tech_script" --local-path "$repo_path" $sbom_arg $claude_arg -o "$output_path/tech-discovery.json" 2>/dev/null
    else
        # Fallback: create basic tech-discovery.json from detection
        local detection=$(detect_project_type "$repo_path")
        local languages=$(echo "$detection" | sed -n '1p')
        local frameworks=$(echo "$detection" | sed -n '2p')
        local package_managers=$(echo "$detection" | sed -n '3p')

        cat > "$output_path/tech-discovery.json" << EOF
{
  "analyzer": "tech-discovery",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "languages": $languages,
  "frameworks": $frameworks,
  "package_managers": $package_managers,
  "summary": {
    "language_count": $(echo "$languages" | jq 'length'),
    "framework_count": $(echo "$frameworks" | jq 'length')
  }
}
EOF
    fi
}

run_dependency_extractor() {
    local repo_path="$1"
    local output_path="$2"

    local direct_count=0
    local total_count=0
    local sbom_format="none"
    local sbom_file=""

    # Use syft for SBOM generation if available
    if command -v syft &> /dev/null; then
        sbom_file="$output_path/sbom.cdx.json"
        sbom_format="CycloneDX"

        # Generate SBOM with syft (CycloneDX format)
        if syft scan "$repo_path" -o cyclonedx-json="$sbom_file" 2>/dev/null; then
            # Count components from SBOM
            total_count=$(jq '.components | length // 0' "$sbom_file" 2>/dev/null)
            [[ -z "$total_count" ]] && total_count=0
        fi
    fi

    # Also count direct dependencies from manifest files
    # Extract from package.json
    if [[ -f "$repo_path/package.json" ]]; then
        local pkg_deps=$(jq -r '.dependencies // {} | keys | length' "$repo_path/package.json" 2>/dev/null)
        [[ -n "$pkg_deps" ]] && direct_count=$pkg_deps
    fi

    # Extract from requirements.txt
    if [[ -f "$repo_path/requirements.txt" ]]; then
        local py_count=$(grep -c '^[^#]' "$repo_path/requirements.txt" 2>/dev/null)
        [[ -n "$py_count" ]] && direct_count=$((direct_count + py_count))
    fi

    # Extract from go.mod
    if [[ -f "$repo_path/go.mod" ]]; then
        local go_count=$(grep -c 'require' "$repo_path/go.mod" 2>/dev/null)
        [[ -n "$go_count" ]] && direct_count=$((direct_count + go_count))
    fi

    # If no syft, use direct count as total
    if [[ "$total_count" -eq 0 ]]; then
        total_count=$direct_count
        sbom_format="manifest-only"
    fi

    # Ensure we have valid numbers
    [[ -z "$direct_count" ]] && direct_count=0
    [[ -z "$total_count" ]] && total_count=0

    cat > "$output_path/package-sbom.json" << EOF
{
  "analyzer": "package-sbom",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "sbom_format": "$sbom_format",
  "sbom_file": "$(basename "$sbom_file" 2>/dev/null)",
  "direct_dependencies": $direct_count,
  "total_dependencies": $total_count,
  "summary": {
    "format": "$sbom_format",
    "direct": $direct_count,
    "total": $total_count
  }
}
EOF
}

run_vulnerability_analyzer() {
    local repo_path="$1"
    local output_path="$2"

    # Use Claude-enabled analyzer in deep mode, otherwise data-only
    local vuln_script=""
    if [[ "${USE_CLAUDE:-}" == "true" ]]; then
        vuln_script="$UTILS_ROOT/scanners/package-vulns/package-vulns.sh"
    else
        vuln_script="$UTILS_ROOT/scanners/package-vulns/package-vulns.sh"
    fi

    if [[ -x "$vuln_script" ]]; then
        local claude_arg=""
        [[ "${USE_CLAUDE:-}" == "true" ]] && claude_arg="--claude"
        "$vuln_script" --local-path "$repo_path" $claude_arg -o "$output_path/package-vulns.json" 2>/dev/null
    else
        cat > "$output_path/package-vulns.json" << EOF
{
  "analyzer": "package-vulns",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "status": "analyzer_not_found",
  "vulnerabilities": [],
  "summary": {
    "critical": 0,
    "high": 0,
    "medium": 0,
    "low": 0
  }
}
EOF
    fi
}

run_package_health_analyzer() {
    local repo_path="$1"
    local output_path="$2"

    # TODO: Integrate with actual package health analyzer when JSON output is standardized
    cat > "$output_path/package-health.json" << EOF
{
  "analyzer": "package-health-analyser",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "status": "pending_integration",
  "note": "Run ./utils/supply-chain/package-health-analysis/package-health-analyser.sh for full report",
  "findings": [],
  "summary": {
    "abandoned": 0,
    "typosquat_risk": 0
  }
}
EOF
}

run_license_analyzer() {
    local repo_path="$1"
    local output_path="$2"

    # Use Claude-enabled analyzer in deep mode, otherwise data-only
    local legal_script=""
    if [[ "${USE_CLAUDE:-}" == "true" ]]; then
        legal_script="$UTILS_ROOT/scanners/licenses/licenses.sh"
    else
        legal_script="$UTILS_ROOT/scanners/licenses/licenses.sh"
    fi

    if [[ -x "$legal_script" ]]; then
        local claude_arg=""
        [[ "${USE_CLAUDE:-}" == "true" ]] && claude_arg="--claude"
        "$legal_script" --local-path "$repo_path" $claude_arg -o "$output_path/licenses.json" 2>/dev/null
    else
        cat > "$output_path/licenses.json" << EOF
{
  "analyzer": "legal-analyser",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "status": "analyzer_not_found",
  "licenses": [],
  "summary": {
    "overall_status": "unknown",
    "license_violations": 0
  }
}
EOF
    fi
}

run_code_security_analyzer() {
    local repo_path="$1"
    local output_path="$2"

    # Use Claude-enabled analyzer in deep mode, otherwise data-only
    local security_script=""
    if [[ "${USE_CLAUDE:-}" == "true" ]]; then
        security_script="$UTILS_ROOT/scanners/code-security/code-security.sh"
    else
        security_script="$UTILS_ROOT/scanners/code-security/code-security.sh"
    fi

    if [[ -x "$security_script" ]]; then
        local claude_arg=""
        [[ "${USE_CLAUDE:-}" == "true" ]] && claude_arg="--claude"
        "$security_script" --local-path "$repo_path" $claude_arg -o "$output_path/code-security.json" 2>/dev/null
    else
        cat > "$output_path/code-security.json" << EOF
{
  "analyzer": "code-security",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "status": "analyzer_not_found",
  "potential_issues": [],
  "summary": {
    "potential_issues": 0,
    "potential_secrets": 0
  }
}
EOF
    fi
}

run_iac_security_analyzer() {
    local repo_path="$1"
    local output_path="$2"

    local iac_script="$UTILS_ROOT/scanners/iac-security/iac-security.sh"

    if [[ -x "$iac_script" ]]; then
        "$iac_script" --local-path "$repo_path" -o "$output_path/iac-security.json" 2>/dev/null
    else
        cat > "$output_path/iac-security.json" << EOF
{
  "analyzer": "iac-security",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "status": "analyzer_not_found",
  "frameworks_detected": [],
  "summary": {
    "total_findings": 0,
    "by_severity": {"critical": 0, "high": 0, "medium": 0, "low": 0, "info": 0}
  },
  "findings": [],
  "compliance_mapping": {}
}
EOF
    fi
}

run_tech_debt_analyzer() {
    local repo_path="$1"
    local output_path="$2"

    local script="$UTILS_ROOT/scanners/tech-debt/tech-debt.sh"

    if [[ -x "$script" ]]; then
        "$script" --local-path "$repo_path" -o "$output_path/tech-debt.json" 2>/dev/null
    else
        cat > "$output_path/tech-debt.json" << EOF
{
  "analyzer": "tech-debt",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "status": "analyzer_not_found",
  "summary": {
    "debt_score": 0,
    "todo_count": 0,
    "fixme_count": 0,
    "hack_count": 0
  }
}
EOF
    fi
}

run_documentation_analyzer() {
    local repo_path="$1"
    local output_path="$2"

    local script="$UTILS_ROOT/scanners/documentation/documentation.sh"

    if [[ -x "$script" ]]; then
        "$script" --local-path "$repo_path" -o "$output_path/documentation.json" 2>/dev/null
    else
        cat > "$output_path/documentation.json" << EOF
{
  "analyzer": "documentation",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "status": "analyzer_not_found",
  "summary": {
    "documentation_score": 0,
    "readme_exists": false,
    "license_exists": false
  }
}
EOF
    fi
}

run_git_insights_analyzer() {
    local repo_path="$1"
    local output_path="$2"

    local script="$UTILS_ROOT/scanners/git/git.sh"

    if [[ -x "$script" ]]; then
        "$script" --local-path "$repo_path" -o "$output_path/git.json" 2>/dev/null
    else
        cat > "$output_path/git.json" << EOF
{
  "analyzer": "git",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "status": "analyzer_not_found",
  "summary": {
    "total_commits": 0,
    "total_contributors": 0,
    "bus_factor": 0
  }
}
EOF
    fi
}

run_test_coverage_analyzer() {
    local repo_path="$1"
    local output_path="$2"

    local script="$UTILS_ROOT/scanners/test-coverage/test-coverage.sh"

    if [[ -x "$script" ]]; then
        "$script" --local-path "$repo_path" -o "$output_path/test-coverage.json" 2>/dev/null
    else
        cat > "$output_path/test-coverage.json" << EOF
{
  "analyzer": "test-coverage",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "status": "analyzer_not_found",
  "summary": {
    "test_files": 0,
    "source_files": 0,
    "test_to_code_ratio": 0
  }
}
EOF
    fi
}

run_secrets_scanner() {
    local repo_path="$1"
    local output_path="$2"

    local script="$UTILS_ROOT/scanners/code-secrets/code-secrets.sh"

    if [[ -x "$script" ]]; then
        "$script" --local-path "$repo_path" --output "$output_path/code-secrets.json" 2>/dev/null
    else
        cat > "$output_path/code-secrets.json" << EOF
{
  "analyzer": "code-secrets",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "status": "analyzer_not_found",
  "summary": {
    "risk_score": 100,
    "risk_level": "unknown",
    "total_findings": 0
  }
}
EOF
    fi
}


run_ownership_analyzer() {
    local repo_path="$1"
    local output_path="$2"

    # Use Claude-enabled analyzer in deep mode, otherwise data-only
    local ownership_script=""
    if [[ "${USE_CLAUDE:-}" == "true" ]]; then
        ownership_script="$UTILS_ROOT/scanners/code-ownership/code-ownership.sh"
    else
        ownership_script="$UTILS_ROOT/scanners/code-ownership/code-ownership.sh"
    fi

    if [[ -x "$ownership_script" ]]; then
        local claude_arg=""
        [[ "${USE_CLAUDE:-}" == "true" ]] && claude_arg="--claude"
        "$ownership_script" --local-path "$repo_path" $claude_arg -o "$output_path/code-ownership.json" 2>/dev/null
    else
        cat > "$output_path/code-ownership.json" << EOF
{
  "analyzer": "code-ownership",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "status": "analyzer_not_found",
  "summary": {
    "total_files": 0,
    "active_contributors": 0
  }
}
EOF
    fi
}

run_bus_factor_analyzer() {
    local repo_path="$1"
    local output_path="$2"

    local bus_factor_script="$UTILS_ROOT/scanners/code-ownership/bus-factor.sh"

    if [[ -x "$bus_factor_script" ]]; then
        "$bus_factor_script" --local-path "$repo_path" -o "$output_path/bus-factor.json" 2>/dev/null
    else
        cat > "$output_path/bus-factor.json" << EOF
{
  "analyzer": "bus-factor",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "status": "analyzer_not_found",
  "summary": {
    "bus_factor": 0,
    "risk_level": "unknown",
    "active_contributors": 0
  }
}
EOF
    fi
}

run_dora_analyzer() {
    local repo_path="$1"
    local output_path="$2"

    # Use Claude-enabled analyzer in deep mode, otherwise data-only
    local dora_script=""
    if [[ "${USE_CLAUDE:-}" == "true" ]]; then
        dora_script="$UTILS_ROOT/scanners/dora/dora.sh"
    else
        dora_script="$UTILS_ROOT/scanners/dora/dora.sh"
    fi

    # Skip if not a git repo
    if [[ ! -d "$repo_path/.git" ]]; then
        cat > "$output_path/dora.json" << EOF
{
  "analyzer": "dora-metrics",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "status": "not_a_git_repo",
  "summary": {}
}
EOF
        return 0
    fi

    if [[ -x "$dora_script" ]]; then
        local claude_arg=""
        [[ "${USE_CLAUDE:-}" == "true" ]] && claude_arg="--claude"
        "$dora_script" --local-path "$repo_path" $claude_arg -o "$output_path/dora.json" 2>/dev/null
    else
        cat > "$output_path/dora.json" << EOF
{
  "analyzer": "dora-metrics",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "status": "analyzer_not_found",
  "summary": {}
}
EOF
    fi
}

run_provenance_analyzer() {
    local repo_path="$1"
    local output_path="$2"

    # Provenance analysis is complex and requires SLSA/sigstore checks
    # For now, create a placeholder with basic git signature info
    local signed_commits=0
    if [[ -d "$repo_path/.git" ]]; then
        # Count signed commits in recent history (last 100)
        signed_commits=$(git -C "$repo_path" log --oneline -100 --show-signature 2>/dev/null | grep -c "Good signature" || echo "0")
    fi

    cat > "$output_path/package-provenance.json" << EOF
{
  "analyzer": "package-provenance",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "status": "basic_check",
  "summary": {
    "signed_commits": $signed_commits,
    "slsa_level": null,
    "note": "Basic git signature check only. Full SLSA provenance verification available via supply-chain analyzer."
  }
}
EOF
}

run_provenance_analyzer_full() {
    local repo_path="$1"
    local output_path="$2"

    local prov_script="$UTILS_ROOT/scanners/package-provenance/package-provenance.sh"

    if [[ -x "$prov_script" ]]; then
        # Use local-path for pre-cloned repo, output JSON format
        "$prov_script" "$repo_path" -f json -o "$output_path/package-provenance.json" 2>/dev/null || {
            # If it fails, create a fallback
            cat > "$output_path/package-provenance.json" << EOF
{
  "analyzer": "package-provenance",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "status": "analysis_failed",
  "summary": {}
}
EOF
        }
    else
        cat > "$output_path/package-provenance.json" << EOF
{
  "analyzer": "package-provenance",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "status": "analyzer_not_found",
  "summary": {}
}
EOF
    fi
}

run_container_analyzer() {
    local repo_path="$1"
    local output_path="$2"

    local script="$UTILS_ROOT/scanners/containers/containers.sh"

    if [[ -x "$script" ]]; then
        "$script" --local-path "$repo_path" -o "$output_path/containers.json" 2>/dev/null
    else
        cat > "$output_path/containers.json" << EOF
{
  "analyzer": "containers",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "status": "analyzer_not_found",
  "summary": {
    "dockerfiles": 0,
    "compose_files": 0,
    "kubernetes_manifests": 0
  }
}
EOF
    fi
}

run_chalk_analyzer() {
    local repo_path="$1"
    local output_path="$2"

    # Use Claude-enabled analyzer in deep mode, otherwise data-only
    local script=""
    if [[ "${USE_CLAUDE:-}" == "true" ]]; then
        script="$UTILS_ROOT/scanners/chalk/chalk.sh"
    else
        script="$UTILS_ROOT/scanners/chalk/chalk.sh"
    fi

    if [[ -x "$script" ]]; then
        local claude_arg=""
        [[ "${USE_CLAUDE:-}" == "true" ]] && claude_arg="--claude"
        "$script" --local-path "$repo_path" $claude_arg -o "$output_path/chalk.json" 2>/dev/null
    else
        cat > "$output_path/chalk.json" << EOF
{
  "analyzer": "chalk",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "status": "analyzer_not_found",
  "summary": {
    "artifacts_found": 0
  }
}
EOF
    fi
}

run_certificate_analyzer() {
    local repo_path="$1"
    local output_path="$2"

    # Use Claude-enabled analyzer in deep mode, otherwise data-only
    local script=""
    if [[ "${USE_CLAUDE:-}" == "true" ]]; then
        script="$UTILS_ROOT/scanners/digital-certificates/digital-certificates.sh"
    else
        script="$UTILS_ROOT/scanners/digital-certificates/digital-certificates.sh"
    fi

    if [[ -x "$script" ]]; then
        local claude_arg=""
        [[ "${USE_CLAUDE:-}" == "true" ]] && claude_arg="--claude"
        "$script" --local-path "$repo_path" $claude_arg -o "$output_path/digital-certificates.json" 2>/dev/null
    else
        cat > "$output_path/digital-certificates.json" << EOF
{
  "analyzer": "digital-certificates",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "status": "analyzer_not_found",
  "summary": {
    "certificates_found": 0,
    "expired": 0,
    "expiring_soon": 0
  }
}
EOF
    fi
}

run_malcontent_analyzer() {
    local repo_path="$1"
    local output_path="$2"
    local project_id="$3"

    local script="$UTILS_ROOT/scanners/package-malcontent/package-malcontent.sh"

    if [[ -x "$script" ]]; then
        # Pass SBOM if available for package-level analysis
        local sbom_arg=""
        if [[ -f "$output_path/sbom.cdx.json" ]]; then
            sbom_arg="--sbom $output_path/sbom.cdx.json"
        fi

        # Run with verbose mode and show-findings for detailed terminal output
        "$script" --local-path "$repo_path" $sbom_arg --verbose --show-findings --repo-name "$project_id" -o "$output_path/package-malcontent.json" 2>&1
    else
        cat > "$output_path/package-malcontent.json" << EOF
{
  "analyzer": "package-malcontent",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "status": "analyzer_not_found",
  "install": "brew install malcontent",
  "summary": {
    "total_files": 0,
    "by_risk": {}
  },
  "findings": []
}
EOF
    fi
}

#############################################################################
# Run All Analyzers
#############################################################################

run_all_analyzers() {
    local repo_path="$1"
    local output_path="$2"
    local project_id="$3"
    local mode="$4"
    local enrich="${5:-false}"

    local requested_analyzers=$(get_analyzers_for_mode "$mode")
    local analyzers_to_run="$requested_analyzers"

    # In enrich mode, only run missing analyzers
    if [[ "$enrich" == "true" ]]; then
        local completed=$(get_completed_analyzers "$output_path")
        analyzers_to_run=$(get_missing_analyzers "$requested_analyzers" "$completed")

        if [[ -z "$analyzers_to_run" ]]; then
            echo -e "\n${GREEN}All requested analyzers already complete.${NC}"
            echo -e "${DIM}Use --force to re-run all analyzers.${NC}"
            return 0
        fi
    fi

    local profile_count=$(echo "$requested_analyzers" | wc -w | tr -d ' ')
    local total_scanners=$(echo "$ALL_SCANNERS" | wc -w | tr -d ' ')

    # Show mode in header
    local mode_display=""
    case "$mode" in
        quick)    mode_display=" ${DIM}(quick mode)${NC}" ;;
        standard) mode_display=" ${DIM}(standard mode)${NC}" ;;
        advanced) mode_display=" ${DIM}(advanced mode)${NC}" ;;
        deep)     mode_display=" ${DIM}(deep mode)${NC}" ;;
        security) mode_display=" ${DIM}(security mode)${NC}" ;;
        custom)   mode_display=" ${DIM}(custom selection)${NC}" ;;
    esac

    if [[ "$enrich" == "true" ]]; then
        echo -e "\n${BOLD}Running $profile_count analyzers${mode_display}${NC} ${CYAN}(enrichment)${NC}"
    else
        echo -e "\n${BOLD}Running $profile_count analyzers${mode_display}${NC}"
    fi
    echo

    # Iterate through ALL scanners for consistent output
    for analyzer in $ALL_SCANNERS; do
        local display_name=$(get_scanner_display_name "$analyzer")
        local in_profile=$(scanner_in_list "$analyzer" "$requested_analyzers" && echo "yes" || echo "no")
        local should_run=$(scanner_in_list "$analyzer" "$analyzers_to_run" && echo "yes" || echo "no")

        # Format: indicator + display name (padded) + result
        if [[ "$in_profile" != "yes" ]]; then
            # Not in this profile - show as dimmed
            printf "  ${DIM}○ %-24s not in profile${NC}\n" "$display_name"
        elif [[ "$should_run" != "yes" ]]; then
            # In profile but already complete (enrich mode)
            printf "  ${GREEN}✓${NC} %-24s ${DIM}(cached)${NC}\n" "$display_name"
        else
            # Run this analyzer - show waiting indicator
            printf "  ${WHITE}○${NC} %-24s ${DIM}running...${NC}" "$display_name"

            local start_time=$(date +%s)
            run_analyzer "$analyzer" "$repo_path" "$output_path" "$project_id"
            local exit_code=$?
            local end_time=$(date +%s)
            local duration=$((end_time - start_time))

            # Clear line and show result
            printf "\r\033[K"

            if [[ $exit_code -eq 0 ]]; then
                printf "  ${GREEN}✓${NC} %-24s " "$display_name"

                # Show inline summary
                local output_file="$output_path/${analyzer}.json"
                if [[ -f "$output_file" ]]; then
                    case "$analyzer" in
                        package-vulns)
                            local c=$(jq -r '.summary.critical // 0' "$output_file" 2>/dev/null)
                            local h=$(jq -r '.summary.high // 0' "$output_file" 2>/dev/null)
                            local m=$(jq -r '.summary.medium // 0' "$output_file" 2>/dev/null)
                            local l=$(jq -r '.summary.low // 0' "$output_file" 2>/dev/null)
                            local total=$((c + h + m + l))
                            # Always show full breakdown
                            if [[ "$c" -gt 0 ]]; then
                                printf "${RED}%dC${NC} " "$c"
                            else
                                printf "${DIM}0C${NC} "
                            fi
                            if [[ "$h" -gt 0 ]]; then
                                printf "${YELLOW}%dH${NC} " "$h"
                            else
                                printf "${DIM}0H${NC} "
                            fi
                            if [[ "$m" -gt 0 ]]; then
                                printf "%dM " "$m"
                            else
                                printf "${DIM}0M${NC} "
                            fi
                            if [[ "$l" -gt 0 ]]; then
                                printf "%dL" "$l"
                            else
                                printf "${DIM}0L${NC}"
                            fi
                            if [[ $total -eq 0 ]]; then
                                printf " ${GREEN}✓${NC}"
                            fi
                            ;;
                        package-sbom)
                            # Get total from package-sbom.json summary
                            local total=$(jq -r '.summary.total // .total_dependencies // 0' "$output_file" 2>/dev/null)
                            local direct=$(jq -r '.summary.direct // .direct_dependencies // 0' "$output_file" 2>/dev/null)
                            # Get ecosystem breakdown from sbom.cdx.json if it exists
                            local sbom_file="$output_path/sbom.cdx.json"
                            local ecosystems=""
                            if [[ -f "$sbom_file" ]]; then
                                ecosystems=$(jq -r '[.components[]? | .purl // empty | ltrimstr("pkg:") | split("/")[0]] | map(select(length > 0)) | group_by(.) | map({type: .[0], count: length}) | sort_by(-.count) | .[0:3] | map("\(.type):\(.count)") | join(" ")' "$sbom_file" 2>/dev/null)
                            fi
                            if [[ -n "$ecosystems" ]] && [[ "$ecosystems" != "null" ]] && [[ "$ecosystems" != "" ]]; then
                                printf "%d packages ${DIM}(%d direct, %s)${NC}" "$total" "$direct" "$ecosystems"
                            else
                                printf "%d packages ${DIM}(%d direct)${NC}" "$total" "$direct"
                            fi
                            ;;
                        tech-discovery)
                            # Show actual technologies found (top 5)
                            local techs=$(jq -r '.technologies // .summary.by_category // {} | if type == "array" then .[0:5] | map(.name // .technology // .) | join(", ") else (to_entries | .[0:5] | map(.key) | join(", ")) end' "$output_file" 2>/dev/null)
                            local tech_count=$(jq -r '.technologies | length // (.summary.total // 0)' "$output_file" 2>/dev/null)
                            if [[ -n "$techs" ]] && [[ "$techs" != "null" ]] && [[ "$techs" != "" ]]; then
                                if [[ $tech_count -gt 5 ]]; then
                                    printf "%s ${DIM}+%d more${NC}" "$techs" "$((tech_count - 5))"
                                else
                                    printf "%s" "$techs"
                                fi
                            else
                                printf "%d technologies" "$tech_count"
                            fi
                            ;;
                        licenses)
                            local total_licenses=$(jq -r '.licenses | length // 0' "$output_file" 2>/dev/null)
                            local violations=$(jq -r '.summary.license_violations // 0' "$output_file" 2>/dev/null)
                            local copyleft=$(jq -r '.summary.copyleft_count // 0' "$output_file" 2>/dev/null)
                            # Get license type breakdown (top 4)
                            local license_breakdown=$(jq -r '[.licenses[]?.license // empty] | group_by(.) | map({lic: .[0], count: length}) | sort_by(-.count) | .[0:4] | map("\(.lic):\(.count)") | join(" ")' "$output_file" 2>/dev/null)

                            if [[ "$violations" -gt 0 ]]; then
                                printf "${RED}%d violations${NC} " "$violations"
                            fi
                            if [[ -n "$license_breakdown" ]] && [[ "$license_breakdown" != "null" ]] && [[ "$license_breakdown" != "" ]]; then
                                printf "${DIM}%s${NC}" "$license_breakdown"
                            else
                                printf "%d licenses" "$total_licenses"
                            fi
                            if [[ "$copyleft" -gt 0 ]]; then
                                printf " ${YELLOW}(%d copyleft)${NC}" "$copyleft"
                            fi
                            ;;
                        code-secrets)
                            local secrets=$(jq -r '.summary.total_findings // (.findings | length) // 0' "$output_file" 2>/dev/null)
                            if [[ "$secrets" -gt 0 ]]; then
                                # Show breakdown by type
                                local by_type=$(jq -r '.summary.by_type // {} | to_entries | map("\(.key):\(.value)") | join(" ")' "$output_file" 2>/dev/null)
                                if [[ -n "$by_type" ]] && [[ "$by_type" != "null" ]] && [[ "$by_type" != "" ]]; then
                                    printf "${RED}%d secrets${NC} ${DIM}(%s)${NC}" "$secrets" "$by_type"
                                else
                                    printf "${RED}%d secrets!${NC}" "$secrets"
                                fi
                            else
                                printf "${GREEN}0 secrets ✓${NC}"
                            fi
                            ;;
                        code-security)
                            local total=$(jq -r '.summary.total // (.findings | length) // 0' "$output_file" 2>/dev/null)
                            local high=$(jq -r '.summary.high // 0' "$output_file" 2>/dev/null)
                            local medium=$(jq -r '.summary.medium // 0' "$output_file" 2>/dev/null)
                            if [[ "$total" -gt 0 ]]; then
                                if [[ "$high" -gt 0 ]]; then
                                    printf "${RED}%d high${NC} " "$high"
                                fi
                                if [[ "$medium" -gt 0 ]]; then
                                    printf "${YELLOW}%d medium${NC}" "$medium"
                                fi
                                if [[ "$high" -eq 0 ]] && [[ "$medium" -eq 0 ]]; then
                                    printf "%d findings" "$total"
                                fi
                            else
                                printf "${GREEN}0 issues ✓${NC}"
                            fi
                            ;;
                        iac-security)
                            local total=$(jq -r '.summary.total // (.findings | length) // 0' "$output_file" 2>/dev/null)
                            local high=$(jq -r '.summary.high // 0' "$output_file" 2>/dev/null)
                            if [[ "$total" -gt 0 ]]; then
                                if [[ "$high" -gt 0 ]]; then
                                    printf "${RED}%d high${NC} %d total" "$high" "$total"
                                else
                                    printf "%d findings" "$total"
                                fi
                            else
                                printf "${GREEN}0 issues ✓${NC}"
                            fi
                            ;;
                        code-ownership)
                            local contributors=$(jq -r '.summary.active_contributors // 0' "$output_file" 2>/dev/null)
                            local bus=$(jq -r '.summary.estimated_bus_factor // 0' "$output_file" 2>/dev/null)
                            local top_contrib=$(jq -r '.contributors[0].name // .contributors[0].author // empty' "$output_file" 2>/dev/null)
                            if [[ "$bus" -le 1 ]]; then
                                printf "${RED}bus factor %d${NC} " "$bus"
                            else
                                printf "bus factor %d " "$bus"
                            fi
                            printf "${DIM}%d contributors${NC}" "$contributors"
                            if [[ -n "$top_contrib" ]]; then
                                printf " ${DIM}(top: %s)${NC}" "$top_contrib"
                            fi
                            ;;
                        dora)
                            local perf=$(jq -r '.summary.overall_performance // "N/A"' "$output_file" 2>/dev/null)
                            local deploy_freq=$(jq -r '.summary.deployment_frequency // .metrics.deployment_frequency.level // "N/A"' "$output_file" 2>/dev/null)
                            if [[ "$perf" == "ELITE" ]]; then
                                printf "${GREEN}ELITE${NC}"
                            elif [[ "$perf" == "HIGH" ]]; then
                                printf "${GREEN}HIGH${NC}"
                            elif [[ "$perf" == "LOW" ]]; then
                                printf "${RED}LOW${NC}"
                            else
                                printf "%s" "$perf"
                            fi
                            if [[ "$deploy_freq" != "N/A" ]] && [[ "$deploy_freq" != "null" ]]; then
                                printf " ${DIM}(deploys: %s)${NC}" "$deploy_freq"
                            fi
                            ;;
                        tech-debt)
                            local debt_score=$(jq -r '.summary.debt_score // 0' "$output_file" 2>/dev/null)
                            local todos=$(jq -r '.summary.todo_count // 0' "$output_file" 2>/dev/null)
                            local fixmes=$(jq -r '.summary.fixme_count // 0' "$output_file" 2>/dev/null)
                            local deprecated=$(jq -r '.summary.deprecated_count // 0' "$output_file" 2>/dev/null)
                            local markers=$((todos + fixmes))
                            if [[ "$debt_score" -gt 70 ]]; then
                                printf "${RED}score %d${NC} " "$debt_score"
                            elif [[ "$debt_score" -gt 40 ]]; then
                                printf "${YELLOW}score %d${NC} " "$debt_score"
                            else
                                printf "score %d " "$debt_score"
                            fi
                            printf "${DIM}%d TODOs %d FIXMEs${NC}" "$todos" "$fixmes"
                            if [[ "$deprecated" -gt 0 ]]; then
                                printf " ${YELLOW}%d deprecated${NC}" "$deprecated"
                            fi
                            ;;
                        documentation)
                            local score=$(jq -r '.summary.score // 0' "$output_file" 2>/dev/null)
                            local has_readme=$(jq -r '.summary.has_readme // false' "$output_file" 2>/dev/null)
                            local has_license=$(jq -r '.summary.has_license // false' "$output_file" 2>/dev/null)
                            printf "score %d " "$score"
                            if [[ "$has_readme" == "true" ]]; then
                                printf "${GREEN}README${NC} "
                            else
                                printf "${RED}no README${NC} "
                            fi
                            if [[ "$has_license" == "true" ]]; then
                                printf "${GREEN}LICENSE${NC}"
                            else
                                printf "${YELLOW}no LICENSE${NC}"
                            fi
                            ;;
                        git)
                            local commits=$(jq -r '.summary.total_commits // 0' "$output_file" 2>/dev/null)
                            local contributors=$(jq -r '.summary.contributors // 0' "$output_file" 2>/dev/null)
                            local commits_90d=$(jq -r '.summary.commits_90d // 0' "$output_file" 2>/dev/null)
                            printf "%d commits " "$commits"
                            printf "${DIM}%d contributors${NC} " "$contributors"
                            if [[ "$commits_90d" -gt 0 ]]; then
                                printf "${DIM}(%d in 90d)${NC}" "$commits_90d"
                            fi
                            ;;
                        test-coverage)
                            local coverage=$(jq -r '.summary.coverage_percentage // 0' "$output_file" 2>/dev/null)
                            local has_tests=$(jq -r '.summary.has_tests // false' "$output_file" 2>/dev/null)
                            if [[ "$has_tests" == "true" ]]; then
                                if [[ "$coverage" -ge 80 ]]; then
                                    printf "${GREEN}%d%%${NC}" "$coverage"
                                elif [[ "$coverage" -ge 50 ]]; then
                                    printf "${YELLOW}%d%%${NC}" "$coverage"
                                else
                                    printf "${RED}%d%%${NC}" "$coverage"
                                fi
                            else
                                printf "${DIM}no tests found${NC}"
                            fi
                            ;;
                        package-provenance)
                            local signed=$(jq -r '.summary.signed_packages // 0' "$output_file" 2>/dev/null)
                            local slsa=$(jq -r '.summary.slsa_level // "none"' "$output_file" 2>/dev/null)
                            if [[ "$slsa" != "none" ]] && [[ "$slsa" != "null" ]]; then
                                printf "${GREEN}SLSA %s${NC} " "$slsa"
                            fi
                            printf "%d signed" "$signed"
                            ;;
                        package-malcontent)
                            local total_files=$(jq -r '.summary.total_files // 0' "$output_file" 2>/dev/null)
                            local total_rules=$(jq -r '.summary.total_rules_matched // 0' "$output_file" 2>/dev/null)
                            local critical=$(jq -r '.summary.by_risk.critical // 0' "$output_file" 2>/dev/null)
                            local high=$(jq -r '.summary.by_risk.high // 0' "$output_file" 2>/dev/null)
                            local medium=$(jq -r '.summary.by_risk.medium // 0' "$output_file" 2>/dev/null)
                            if [[ "$total_files" -gt 0 ]]; then
                                printf "%d files " "$total_files"
                                if [[ "$critical" -gt 0 ]]; then
                                    printf "${RED}%dC${NC} " "$critical"
                                fi
                                if [[ "$high" -gt 0 ]]; then
                                    printf "${YELLOW}%dH${NC} " "$high"
                                fi
                                if [[ "$medium" -gt 0 ]]; then
                                    printf "%dM " "$medium"
                                fi
                                printf "${DIM}(%d rules)${NC}" "$total_rules"
                            else
                                printf "${GREEN}0 findings ✓${NC}"
                            fi
                            ;;
                        *)
                            printf "done"
                            ;;
                    esac
                fi
                printf " ${DIM}%ds${NC}\n" "$duration"
            else
                printf "  ${RED}✗${NC} %-24s ${RED}failed${NC} ${DIM}%ds${NC}\n" "$display_name" "$duration"
            fi
        fi
    done

    echo
}

#############################################################################
# Summary Generation
#############################################################################

generate_summary() {
    local project_id="$1"
    local output_path="$2"

    local risk_level="low"
    local total_deps=0
    local direct_deps=0
    local total_vulns=0
    local total_findings=0
    local license_status="unknown"
    local abandoned=0

    # Read from analysis outputs
    if [[ -f "$output_path/dependencies.json" ]]; then
        direct_deps=$(jq -r '.direct_dependencies // 0' "$output_path/dependencies.json" 2>/dev/null)
        total_deps=$(jq -r '.total_dependencies // 0' "$output_path/dependencies.json" 2>/dev/null)
    fi

    if [[ -f "$output_path/vulnerabilities.json" ]]; then
        local critical=$(jq -r '.summary.critical // 0' "$output_path/vulnerabilities.json" 2>/dev/null)
        local high=$(jq -r '.summary.high // 0' "$output_path/vulnerabilities.json" 2>/dev/null)
        local medium=$(jq -r '.summary.medium // 0' "$output_path/vulnerabilities.json" 2>/dev/null)
        local low_v=$(jq -r '.summary.low // 0' "$output_path/vulnerabilities.json" 2>/dev/null)
        total_vulns=$((critical + high + medium + low_v))

        if [[ "$critical" -gt 0 ]]; then
            risk_level="critical"
        elif [[ "$high" -gt 0 ]]; then
            risk_level="high"
        elif [[ "$medium" -gt 0 ]]; then
            risk_level="medium"
        fi
    fi

    if [[ -f "$output_path/security-findings.json" ]]; then
        # code-security-data.sh outputs potential_issues count
        total_findings=$(jq -r '.summary.potential_issues // 0' "$output_path/security-findings.json" 2>/dev/null)
        local secrets=$(jq -r '.summary.potential_secrets // 0' "$output_path/security-findings.json" 2>/dev/null)
        if [[ "$secrets" -gt 0 ]]; then
            risk_level="critical"
        fi
    fi

    if [[ -f "$output_path/licenses.json" ]]; then
        # legal-analyser-data.sh outputs overall_status
        license_status=$(jq -r '.summary.overall_status // "unknown"' "$output_path/licenses.json" 2>/dev/null)
        local violations=$(jq -r '.summary.license_violations // 0' "$output_path/licenses.json" 2>/dev/null)
        if [[ "$violations" -gt 0 ]]; then
            risk_level="high"
        fi
    fi

    if [[ -f "$output_path/package-health.json" ]]; then
        abandoned=$(jq -r '.summary.abandoned // 0' "$output_path/package-health.json" 2>/dev/null)
    fi

    # Update manifest summary
    gibson_update_summary "$project_id" "$risk_level" "$total_deps" "$direct_deps" "$total_vulns" "$total_findings" "$license_status" "$abandoned"
}

print_final_summary() {
    local project_id="$1"
    local output_path="$2"

    echo
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    if [[ "${USE_CLAUDE:-}" == "true" ]]; then
        echo -e "${BOLD}HYDRATION COMPLETE${NC} ${CYAN}(Claude-assisted deep analysis)${NC}"
    else
        echo -e "${BOLD}HYDRATION COMPLETE${NC} ${DIM}(static analysis - no AI)${NC}"
    fi
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    local manifest="$output_path/manifest.json"
    if [[ -f "$manifest" ]]; then
        local risk=$(jq -r '.summary.risk_level' "$manifest" 2>/dev/null)
        local risk_color="$GREEN"
        case "$risk" in
            critical) risk_color="$RED" ;;
            high) risk_color="$RED" ;;
            medium) risk_color="$YELLOW" ;;
        esac
        local risk_upper=$(echo "$risk" | tr '[:lower:]' '[:upper:]')
        echo -e "\nRisk Level: ${risk_color}${risk_upper}${NC}"

        echo

        # Package Vulnerabilities (CVEs in dependencies)
        if [[ -f "$output_path/vulnerabilities.json" ]]; then
            local c=$(jq -r '.summary.critical // 0' "$output_path/vulnerabilities.json" 2>/dev/null)
            local h=$(jq -r '.summary.high // 0' "$output_path/vulnerabilities.json" 2>/dev/null)
            local m=$(jq -r '.summary.medium // 0' "$output_path/vulnerabilities.json" 2>/dev/null)
            local l=$(jq -r '.summary.low // 0' "$output_path/vulnerabilities.json" 2>/dev/null)
            local total=$((c + h + m + l))
            printf "  %-18s " "Pkg Vulnerabilities:"
            if [[ $c -gt 0 ]] || [[ $h -gt 0 ]]; then
                echo -e "${RED}$c critical${NC}, ${YELLOW}$h high${NC}, $m medium, $l low"
            elif [[ $total -gt 0 ]]; then
                echo "$total found ($m medium, $l low)"
            else
                echo -e "${GREEN}None found${NC}"
            fi
        fi

        # Security findings
        if [[ -f "$output_path/security-findings.json" ]]; then
            local issues=$(jq -r '.summary.potential_issues // 0' "$output_path/security-findings.json" 2>/dev/null)
            local secrets=$(jq -r '.summary.potential_secrets // 0' "$output_path/security-findings.json" 2>/dev/null)
            local files=$(jq -r '.summary.total_files // 0' "$output_path/security-findings.json" 2>/dev/null)
            printf "  %-18s " "Code Security:"
            if [[ $secrets -gt 0 ]]; then
                echo -e "${RED}$secrets potential secrets${NC}, $issues code issues"
            elif [[ $issues -gt 0 ]]; then
                echo -e "${YELLOW}$issues potential issues${NC} in $files files"
            else
                echo -e "${GREEN}Clean${NC} ($files files scanned)"
            fi
        fi

        # SBOM / Dependencies
        if [[ -f "$output_path/dependencies.json" ]]; then
            local total=$(jq -r '.total_dependencies // 0' "$output_path/dependencies.json" 2>/dev/null)
            local format=$(jq -r '.sbom_format // "unknown"' "$output_path/dependencies.json" 2>/dev/null)
            printf "  %-18s %s packages (%s)\n" "SBOM:" "$total" "$format"
        fi

        # Licenses
        if [[ -f "$output_path/licenses.json" ]]; then
            local lic_status=$(jq -r '.summary.overall_status // "unknown"' "$output_path/licenses.json" 2>/dev/null)
            local violations=$(jq -r '.summary.license_violations // 0' "$output_path/licenses.json" 2>/dev/null)
            printf "  %-18s " "Licenses:"
            if [[ "$lic_status" == "fail" ]] || [[ $violations -gt 0 ]]; then
                echo -e "${RED}✗ $violations violations${NC}"
            elif [[ "$lic_status" == "warning" ]]; then
                echo -e "${YELLOW}⚠ Review needed${NC}"
            elif [[ "$lic_status" == "pass" ]]; then
                echo -e "${GREEN}✓ All clear${NC}"
            else
                echo "Unknown"
            fi
        fi

        # Ownership
        if [[ -f "$output_path/ownership.json" ]]; then
            local contributors=$(jq -r '.summary.active_contributors // 0' "$output_path/ownership.json" 2>/dev/null)
            local bus_factor=$(jq -r '.summary.estimated_bus_factor // 0' "$output_path/ownership.json" 2>/dev/null)
            local bus_risk=$(jq -r '.risk_assessment.bus_factor_risk // "unknown"' "$output_path/ownership.json" 2>/dev/null)
            local bus_color="$GREEN"
            [[ "$bus_risk" == "critical" ]] && bus_color="$RED"
            [[ "$bus_risk" == "high" ]] && bus_color="$RED"
            [[ "$bus_risk" == "medium" ]] && bus_color="$YELLOW"
            printf "  %-18s %d active, bus factor: ${bus_color}%d (%s)${NC}\n" "Ownership:" "$contributors" "$bus_factor" "$bus_risk"
        fi

        # DORA
        if [[ -f "$output_path/dora.json" ]]; then
            local dora_perf=$(jq -r '.summary.overall_performance // "N/A"' "$output_path/dora.json" 2>/dev/null)
            local dora_color="$GREEN"
            [[ "$dora_perf" == "LOW" ]] && dora_color="$RED"
            [[ "$dora_perf" == "MEDIUM" ]] && dora_color="$YELLOW"
            printf "  %-18s ${dora_color}%s${NC}\n" "DORA Performance:" "$dora_perf"
        fi
    fi

    echo
    echo -e "${BOLD}Agents ready:${NC}"
    echo -e "  Scout     → ${CYAN}/zero ask scout ...${NC}"
    echo -e "  Sentinel  → ${CYAN}/zero ask sentinel ...${NC}"
    echo -e "  Quinn     → ${CYAN}/zero ask quinn ...${NC}"
    echo -e "  Harper    → ${CYAN}/zero ask harper ...${NC}"

    echo
    local size=$(gibson_project_size "$project_id")
    echo -e "Storage: ${CYAN}~/.zero/projects/$project_id/${NC} ($size)"
}

#############################################################################
# Main
#############################################################################

main() {
    parse_args "$@"

    # Run preflight check (silent unless errors)
    if [[ -x "$SCRIPT_DIR/preflight.sh" ]]; then
        if ! "$SCRIPT_DIR/preflight.sh" > /dev/null 2>&1; then
            echo -e "${RED}Preflight check failed. Run ./utils/zero/preflight.sh to see details.${NC}"
            exit 1
        fi
    fi

    # Regenerate Semgrep rules from RAG patterns (ensures latest rules)
    local rag_to_semgrep="$UTILS_ROOT/scanners/semgrep/rag-to-semgrep.py"
    local rag_dir="$REPO_ROOT/rag/technology-identification"
    local semgrep_rules_dir="$UTILS_ROOT/scanners/semgrep/rules"
    if [[ -x "$rag_to_semgrep" ]] || [[ -f "$rag_to_semgrep" ]]; then
        if command -v python3 &> /dev/null && [[ -d "$rag_dir" ]]; then
            python3 "$rag_to_semgrep" "$rag_dir" "$semgrep_rules_dir" > /dev/null 2>&1 || true
        fi
    fi

    # Sync Semgrep community rules (cached, only updates if stale)
    local community_rules="$UTILS_ROOT/scanners/semgrep/community-rules.sh"
    if [[ -x "$community_rules" ]] && command -v semgrep &> /dev/null; then
        "$community_rules" sync default > /dev/null 2>&1 || true
    fi

    # Ensure Gibson is initialized
    gibson_ensure_initialized

    # Generate project ID
    local project_id=$(gibson_project_id "$TARGET")

    # Setup paths
    local project_path=$(gibson_project_path "$project_id")
    local repo_path=$(gibson_project_repo_path "$project_id")
    local analysis_path=$(gibson_project_analysis_path "$project_id")

    # Handle enrich mode - project must already exist
    if [[ "$ENRICH" == "true" ]]; then
        if ! gibson_project_exists "$project_id"; then
            echo -e "${RED}Project '$project_id' does not exist. Cannot enrich.${NC}"
            echo "Use standard hydration first: ./zero.sh hydrate $TARGET"
            exit 1
        fi

        # Generate scan ID for tracking
        local scan_id=$(generate_scan_id)
        local scan_start_time=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

        # Get git context for history tracking
        local git_context=""
        if [[ -d "$repo_path/.git" ]]; then
            git_context=$(gibson_get_git_context "$repo_path")
        fi

        # Print header for enrichment
        gibson_print_header
        echo -e "Enriching: ${CYAN}$TARGET${NC}"
        echo -e "Project ID: ${CYAN}$project_id${NC}"
        echo -e "Scan ID: ${DIM}$scan_id${NC}"

        # Skip cloning, just run missing analyzers
        run_all_analyzers "$repo_path" "$analysis_path" "$project_id" "$MODE" "true"

        # Update summary and finalize
        generate_summary "$project_id" "$analysis_path"
        gibson_finalize_manifest "$project_id"

        # Record scan in history
        local scan_end_time=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
        # Calculate duration - use TZ=UTC for macOS timestamp parsing
        local start_epoch
        if [[ "$scan_start_time" == *Z ]]; then
            start_epoch=$(TZ=UTC date -j -f "%Y-%m-%dT%H:%M:%SZ" "$scan_start_time" +%s 2>/dev/null)
        fi
        [[ -z "$start_epoch" ]] && start_epoch=$(date -d "$scan_start_time" +%s 2>/dev/null || echo "0")
        local end_epoch=$(date +%s)
        local duration_seconds=$((end_epoch - start_epoch))
        [[ $duration_seconds -lt 0 ]] && duration_seconds=0

        # Extract git info from context
        local commit_hash=$(echo "$git_context" | jq -r '.commit_hash // ""' 2>/dev/null)
        local commit_short=$(echo "$git_context" | jq -r '.commit_short // ""' 2>/dev/null)
        local git_branch=$(echo "$git_context" | jq -r '.branch // ""' 2>/dev/null)

        # Convert scanners to JSON array
        local scanners_run=$(get_analyzers_for_mode "$MODE")
        local scanners_json=$(echo "$scanners_run" | tr ' ' '\n' | jq -R . | jq -sc '.')

        gibson_append_scan_history "$project_id" "$scan_id" "$commit_hash" "$commit_short" "$git_branch" "$scan_start_time" "$scan_end_time" "$duration_seconds" "$MODE" "$scanners_json" "complete" "{}"

        # Update org-level index
        gibson_update_org_index "$project_id"

        gibson_index_update_status "$project_id" "ready"
        gibson_set_active_project "$project_id"
        print_final_summary "$project_id" "$analysis_path"
        return 0
    fi

    # Handle scan-only mode - project must already be cloned
    if [[ "$SCAN_ONLY" == "true" ]]; then
        if ! gibson_project_exists "$project_id"; then
            echo -e "${RED}Project '$project_id' does not exist. Cannot scan.${NC}"
            echo "Clone first with: ./zero.sh clone $TARGET"
            exit 1
        fi

        if [[ ! -d "$repo_path" ]]; then
            echo -e "${RED}Repository not found at $repo_path${NC}"
            echo "Clone first with: ./zero.sh clone $TARGET"
            exit 1
        fi

        # Generate scan ID for tracking
        local scan_id=$(generate_scan_id)
        local scan_start_time=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

        # Get git context for history tracking
        local git_context=""
        if [[ -d "$repo_path/.git" ]]; then
            git_context=$(gibson_get_git_context "$repo_path")
        fi

        # Print header for scan
        gibson_print_header
        echo -e "Scanning: ${CYAN}$TARGET${NC}"
        echo -e "Project ID: ${CYAN}$project_id${NC}"
        echo -e "Mode: ${CYAN}$MODE${NC}"
        echo -e "Scan ID: ${DIM}$scan_id${NC}"

        # Create analysis directory if needed
        mkdir -p "$analysis_path"

        # Initialize/update analysis manifest
        local commit=$(echo "$git_context" | jq -r '.commit_hash // ""' 2>/dev/null)
        gibson_init_analysis_manifest "$project_id" "$commit" "$MODE" "$scan_id" "$git_context"

        # Run analyzers
        run_all_analyzers "$repo_path" "$analysis_path" "$project_id" "$MODE"

        # Update summary and finalize
        generate_summary "$project_id" "$analysis_path"
        gibson_finalize_manifest "$project_id"

        # Record scan in history
        local scan_end_time=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
        local start_epoch
        if [[ "$scan_start_time" == *Z ]]; then
            start_epoch=$(TZ=UTC date -j -f "%Y-%m-%dT%H:%M:%SZ" "$scan_start_time" +%s 2>/dev/null)
        fi
        [[ -z "$start_epoch" ]] && start_epoch=$(date -d "$scan_start_time" +%s 2>/dev/null || echo "0")
        local end_epoch=$(date +%s)
        local duration_seconds=$((end_epoch - start_epoch))
        [[ $duration_seconds -lt 0 ]] && duration_seconds=0

        local commit_hash=$(echo "$git_context" | jq -r '.commit_hash // ""' 2>/dev/null)
        local commit_short=$(echo "$git_context" | jq -r '.commit_short // ""' 2>/dev/null)
        local git_branch=$(echo "$git_context" | jq -r '.branch // ""' 2>/dev/null)

        local scanners_run=$(get_analyzers_for_mode "$MODE")
        local scanners_json=$(echo "$scanners_run" | tr ' ' '\n' | jq -R . | jq -sc '.')

        gibson_append_scan_history "$project_id" "$scan_id" "$commit_hash" "$commit_short" "$git_branch" "$scan_start_time" "$scan_end_time" "$duration_seconds" "$MODE" "$scanners_json" "complete" "{}"

        gibson_update_org_index "$project_id"
        gibson_index_update_status "$project_id" "ready"
        gibson_set_active_project "$project_id"
        print_final_summary "$project_id" "$analysis_path"
        return 0
    fi

    # Check if project already exists
    if gibson_project_exists "$project_id" && [[ "$FORCE" != "true" ]]; then
        echo -e "${YELLOW}Project '$project_id' already exists.${NC}"
        echo "Use --force to re-bootstrap, or --enrich to add missing analyzers."
        exit 1
    fi

    # Print header
    gibson_print_header
    echo -e "Target: ${CYAN}$TARGET${NC}"
    echo -e "Project ID: ${CYAN}$project_id${NC}"
    echo

    # Clean up if force
    if [[ "$FORCE" == "true" ]] && [[ -d "$project_path" ]]; then
        rm -rf "$project_path"
    fi

    # Create directories
    mkdir -p "$project_path"
    mkdir -p "$analysis_path"

    # Add to index
    gibson_index_add_project "$project_id" "$TARGET" "bootstrapping"

    # Clone or copy project
    echo -n "Cloning..."
    if gibson_is_local_source "$TARGET"; then
        copy_local_project "$TARGET" "$repo_path"
        echo -e " ${GREEN}✓${NC} (local copy)"
    else
        local clone_url=$(gibson_clone_url "$TARGET")
        if [[ -z "$clone_url" ]]; then
            echo -e " ${RED}✗${NC}"
            echo "Error: Could not determine clone URL for '$TARGET'"
            exit 1
        fi

        local clone_output
        clone_output=$(clone_github_repo "$clone_url" "$repo_path" "$BRANCH" "$DEPTH" 2>&1)
        if [[ $? -ne 0 ]]; then
            echo -e " ${RED}✗${NC}"
            echo "Clone failed: $clone_output"
            gibson_index_remove_project "$project_id"
            exit 1
        fi
        echo -e " ${GREEN}✓${NC}"
    fi

    # If clone-only mode, stop here
    if [[ "$CLONE_ONLY" == "true" ]]; then
        # Get basic git info for display
        local git_info=$(get_git_info "$repo_path")
        local branch=$(echo "$git_info" | sed -n '1p')
        local commit=$(echo "$git_info" | sed -n '2p')
        [[ -n "$branch" ]] && echo -e "  Branch: ${CYAN}$branch${NC}"
        [[ -n "$commit" ]] && echo -e "  Commit: ${CYAN}$commit${NC}"

        # Create basic project metadata
        local source_type="github"
        gibson_is_local_source "$TARGET" && source_type="local"
        gibson_create_project_metadata "$project_id" "$TARGET" "$source_type" "$branch" "$commit"
        gibson_index_update_status "$project_id" "cloned"

        echo
        echo -e "${GREEN}✓ Clone complete${NC}"
        echo -e "Run scanners with: ${CYAN}./zero.sh scan $TARGET${NC}"
        return 0
    fi

    # Generate scan ID for tracking
    local scan_id=$(generate_scan_id)
    local scan_start_time=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Get git info (full context for tracking)
    local git_info=$(get_git_info "$repo_path")
    local branch=$(echo "$git_info" | sed -n '1p')
    local commit=$(echo "$git_info" | sed -n '2p')

    # Get full git context JSON for history tracking
    local git_context=""
    if [[ -d "$repo_path/.git" ]]; then
        git_context=$(gibson_get_git_context "$repo_path")
    fi

    if [[ -n "$branch" ]]; then
        echo -e "  Branch: ${CYAN}$branch${NC}"
        echo -e "  Commit: ${CYAN}$commit${NC}"
    fi
    echo -e "  Scan ID: ${DIM}$scan_id${NC}"

    # Detect project type (silent - no separate output line)
    local detection=$(detect_project_type "$repo_path")
    local languages=$(echo "$detection" | sed -n '1p')
    local frameworks=$(echo "$detection" | sed -n '2p')
    local package_managers=$(echo "$detection" | sed -n '3p')

    local langs_display=$(echo "$languages" | jq -r 'join(", ")' 2>/dev/null)
    local fwks_display=$(echo "$frameworks" | jq -r 'join(", ")' 2>/dev/null)
    # Always print Languages line (hydrate.sh uses this as phase marker)
    echo -e "  Languages: ${CYAN}${langs_display:-unknown}${NC}"
    [[ -n "$fwks_display" ]] && echo -e "  Frameworks: ${CYAN}$fwks_display${NC}"

    # Determine source type
    local source_type="github"
    gibson_is_local_source "$TARGET" && source_type="local"

    # Create project metadata
    gibson_create_project_metadata "$project_id" "$TARGET" "$source_type" "$branch" "$commit"
    gibson_update_project_type "$project_id" "$languages" "$frameworks" "$package_managers"

    # Initialize analysis manifest (with mode, scan_id, and git context)
    gibson_init_analysis_manifest "$project_id" "$commit" "$MODE" "$scan_id" "$git_context"

    # Run analyzers
    run_all_analyzers "$repo_path" "$analysis_path" "$project_id" "$MODE"

    # Generate summary
    generate_summary "$project_id" "$analysis_path"

    # Finalize manifest
    gibson_finalize_manifest "$project_id"

    # Record scan in history
    local scan_end_time=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    # Calculate duration - use TZ=UTC for macOS timestamp parsing
    local start_epoch
    if [[ "$scan_start_time" == *Z ]]; then
        start_epoch=$(TZ=UTC date -j -f "%Y-%m-%dT%H:%M:%SZ" "$scan_start_time" +%s 2>/dev/null)
    fi
    [[ -z "$start_epoch" ]] && start_epoch=$(date -d "$scan_start_time" +%s 2>/dev/null || echo "0")
    local end_epoch=$(date +%s)
    local duration_seconds=$((end_epoch - start_epoch))
    [[ $duration_seconds -lt 0 ]] && duration_seconds=0

    # Extract git info from context
    local commit_hash=$(echo "$git_context" | jq -r '.commit_hash // ""' 2>/dev/null)
    local commit_short=$(echo "$git_context" | jq -r '.commit_short // ""' 2>/dev/null)
    local git_branch=$(echo "$git_context" | jq -r '.branch // ""' 2>/dev/null)

    # Convert scanners to JSON array
    local scanners_run=$(get_analyzers_for_mode "$MODE")
    local scanners_json=$(echo "$scanners_run" | tr ' ' '\n' | jq -R . | jq -sc '.')

    gibson_append_scan_history "$project_id" "$scan_id" "$commit_hash" "$commit_short" "$git_branch" "$scan_start_time" "$scan_end_time" "$duration_seconds" "$MODE" "$scanners_json" "complete" "{}"

    # Update org-level index for agent queries
    gibson_update_org_index "$project_id"

    # Update index status
    gibson_index_update_status "$project_id" "ready"

    # Set as active project
    gibson_set_active_project "$project_id"

    # Print final summary
    print_final_summary "$project_id" "$analysis_path"
}

main "$@"
