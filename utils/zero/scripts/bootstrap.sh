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

[[ -n "${DEBUG_BOOTSTRAP:-}" ]] && echo "[DEBUG] Script started" >&2

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ZERO_UTILS_DIR="$(dirname "$SCRIPT_DIR")"

[[ -n "${DEBUG_BOOTSTRAP:-}" ]] && echo "[DEBUG] Loading zero-lib.sh..." >&2
# Load Zero library (sets ZERO_DIR to .zero data directory in project root)
source "$ZERO_UTILS_DIR/lib/zero-lib.sh"
[[ -n "${DEBUG_BOOTSTRAP:-}" ]] && echo "[DEBUG] zero-lib.sh loaded" >&2

# Load shared config if available
UTILS_ROOT="$(dirname "$ZERO_UTILS_DIR")"
REPO_ROOT="$(dirname "$UTILS_ROOT")"

if [[ -f "$UTILS_ROOT/lib/config.sh" ]]; then
    [[ -n "${DEBUG_BOOTSTRAP:-}" ]] && echo "[DEBUG] Loading config.sh..." >&2
    source "$UTILS_ROOT/lib/config.sh"
    [[ -n "${DEBUG_BOOTSTRAP:-}" ]] && echo "[DEBUG] config.sh loaded" >&2
fi

[[ -n "${DEBUG_BOOTSTRAP:-}" ]] && echo "[DEBUG] Loading config-loader.sh..." >&2
# Load config loader for dynamic profiles
source "$ZERO_UTILS_DIR/config/config-loader.sh"
[[ -n "${DEBUG_BOOTSTRAP:-}" ]] && echo "[DEBUG] config-loader.sh loaded" >&2

# Load .env if it exists
if [[ -f "$REPO_ROOT/.env" ]]; then
    [[ -n "${DEBUG_BOOTSTRAP:-}" ]] && echo "[DEBUG] Loading .env..." >&2
    set -a
    source "$REPO_ROOT/.env"
    set +a
    [[ -n "${DEBUG_BOOTSTRAP:-}" ]] && echo "[DEBUG] .env loaded" >&2
fi

#############################################################################
# Configuration
#############################################################################

TARGET=""
BRANCH=""
DEPTH=""
[[ -n "${DEBUG_BOOTSTRAP:-}" ]] && echo "[DEBUG] Getting default profile..." >&2
MODE="$(get_default_profile)"  # Load default from config
[[ -n "${DEBUG_BOOTSTRAP:-}" ]] && echo "[DEBUG] Default profile: $MODE" >&2
FORCE=false
ENRICH=false      # Incremental enrichment - only run missing collectors
CLONE_ONLY=false  # Just clone, don't scan
SCAN_ONLY=false   # Just scan existing clone, don't clone
PARALLEL=true     # Run scanners in parallel (default: true)
STATUS_DIR=""     # Optional status directory for progress tracking

# Canonical list of ALL scanners - loaded from config for display
# This ensures consistent output format across all profiles
[[ -n "${DEBUG_BOOTSTRAP:-}" ]] && echo "[DEBUG] Getting all scanners..." >&2
ALL_SCANNERS="$(get_all_scanners)"
[[ -n "${DEBUG_BOOTSTRAP:-}" ]] && echo "[DEBUG] Scanners: $ALL_SCANNERS" >&2

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
    --parallel          Run scanners in parallel (default: enabled)
    --no-parallel       Run scanners sequentially
    -h, --help          Show this help

EXAMPLES:
    $0 expressjs/express                    # Standard analysis
    $0 owner/repo --quick                   # Fast scan
    $0 owner/repo --advanced                # Include package-health, provenance
    $0 owner/repo --deep                    # Claude-enhanced analysis
    $0 ./local-project --security           # Security-focused scan

FLOW:
    1. Clone repository to .zero/repos/<id>/repo/
    2. Run analyzers and store JSON in .zero/repos/<id>/analysis/
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
            --parallel)
                PARALLEL=true
                shift
                ;;
            --no-parallel)
                PARALLEL=false
                shift
                ;;
            --status-dir)
                STATUS_DIR="$2"
                shift 2
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

#############################################################################
# Parallel Execution Support
#############################################################################

# Scanners that depend on SBOM (must wait for package-sbom to complete)
SBOM_DEPENDENT_SCANNERS="tech-discovery package-vulns package-health licenses package-malcontent package-provenance"

# Check if scanner depends on SBOM
scanner_needs_sbom() {
    local scanner="$1"
    [[ " $SBOM_DEPENDENT_SCANNERS " =~ " $scanner " ]]
}

# Run scanner in background and save result to temp file
# Usage: run_scanner_background <analyzer> <repo_path> <output_path> <project_id> <result_dir>
run_scanner_background() {
    local analyzer="$1"
    local repo_path="$2"
    local output_path="$3"
    local project_id="$4"
    local result_dir="$5"

    local start_time=$(date +%s)
    run_analyzer "$analyzer" "$repo_path" "$output_path" "$project_id"
    local exit_code=$?
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))

    # Save result to temp file
    echo "$exit_code $duration" > "$result_dir/$analyzer.result"
}

# Display scanner result from temp file or status file
# Usage: display_scanner_result <analyzer> <output_path> <result_dir> [status_dir] [error_dir]
display_scanner_result() {
    local analyzer="$1"
    local output_path="$2"
    local result_dir="$3"
    local status_dir="${4:-}"
    local error_dir="${5:-}"
    local display_name=$(get_scanner_display_name "$analyzer")

    # Try status file first (new parallel scan method)
    if [[ -n "$status_dir" ]] && [[ -f "$status_dir/$analyzer.status" ]]; then
        local status_file="$status_dir/$analyzer.status"
        IFS='|' read -r status result duration < "$status_file"

        if [[ "$status" == "complete" ]]; then
            printf "  ${GREEN}✓${NC} %-24s %s ${DIM}%ds${NC}\n" "$display_name" "$result" "$duration"
            return 0
        elif [[ "$status" == "failed" ]]; then
            printf "  ${RED}✗${NC} %-24s ${RED}failed${NC} ${DIM}%ds${NC}\n" "$display_name" "$duration"
            # Show error message if available
            if [[ -n "$result" ]]; then
                printf "      ${DIM}└─ %s${NC}\n" "$result"
            fi
            # Check for full error log
            if [[ -n "$error_dir" ]] && [[ -f "$error_dir/$analyzer.err" ]]; then
                local error_log="$error_dir/$analyzer.err"
                if [[ -s "$error_log" ]]; then
                    printf "      ${DIM}└─ See log: %s${NC}\n" "$error_log"
                fi
            fi
            return 1
        fi
    fi

    # Fall back to old result file method
    local result_file="$result_dir/$analyzer.result"
    if [[ ! -f "$result_file" ]]; then
        printf "  ${RED}✗${NC} %-24s ${RED}no result${NC}\n" "$display_name"
        return 1
    fi

    read exit_code duration < "$result_file"

    if [[ "$exit_code" -eq 0 ]]; then
        printf "  ${GREEN}✓${NC} %-24s " "$display_name"
        display_scanner_summary "$analyzer" "$output_path"
        printf " ${DIM}%ds${NC}\n" "$duration"
    else
        printf "  ${RED}✗${NC} %-24s ${RED}failed${NC} ${DIM}%ds${NC}\n" "$display_name" "$duration"
    fi

    return $exit_code
}

# Display inline summary for a scanner (extracted from run_all_analyzers)
display_scanner_summary() {
    local analyzer="$1"
    local output_path="$2"
    local output_file="$output_path/${analyzer}.json"

    if [[ ! -f "$output_file" ]]; then
        printf "done"
        return
    fi

    case "$analyzer" in
        package-vulns)
            local c=$(jq -r '.summary.critical // 0' "$output_file" 2>/dev/null)
            local h=$(jq -r '.summary.high // 0' "$output_file" 2>/dev/null)
            local m=$(jq -r '.summary.medium // 0' "$output_file" 2>/dev/null)
            local l=$(jq -r '.summary.low // 0' "$output_file" 2>/dev/null)
            local total=$((c + h + m + l))
            if [[ "$c" -gt 0 ]]; then printf "${RED}%dC${NC} " "$c"; else printf "${DIM}0C${NC} "; fi
            if [[ "$h" -gt 0 ]]; then printf "${YELLOW}%dH${NC} " "$h"; else printf "${DIM}0H${NC} "; fi
            if [[ "$m" -gt 0 ]]; then printf "%dM " "$m"; else printf "${DIM}0M${NC} "; fi
            if [[ "$l" -gt 0 ]]; then printf "%dL" "$l"; else printf "${DIM}0L${NC}"; fi
            [[ $total -eq 0 ]] && printf " ${GREEN}✓${NC}" || true
            ;;
        package-sbom)
            local total=$(jq -r '.summary.total // .total_dependencies // 0' "$output_file" 2>/dev/null)
            local ecosystems=$(jq -r '[.summary.ecosystems // {} | to_entries[] | "\(.key):\(.value)"] | join(" ")' "$output_file" 2>/dev/null)
            printf "%d deps" "$total"
            [[ -n "$ecosystems" ]] && printf " ${DIM}(%s)${NC}" "$ecosystems" || true
            ;;
        package-health)
            local abandoned=$(jq -r '.summary.abandoned // 0' "$output_file" 2>/dev/null)
            local deprecated=$(jq -r '.summary.deprecated // 0' "$output_file" 2>/dev/null)
            local unmaintained=$(jq -r '.summary.unmaintained // 0' "$output_file" 2>/dev/null)
            local total=$((abandoned + deprecated + unmaintained))
            if [[ $total -gt 0 ]]; then
                [[ "$abandoned" -gt 0 ]] && printf "${RED}%d abandoned${NC} " "$abandoned" || true
                [[ "$deprecated" -gt 0 ]] && printf "${YELLOW}%d deprecated${NC} " "$deprecated" || true
                [[ "$unmaintained" -gt 0 ]] && printf "%d unmaintained " "$unmaintained" || true
            else
                printf "${GREEN}healthy ✓${NC}"
            fi
            ;;
        licenses)
            local status=$(jq -r '.summary.status // .summary.overall_status // "unknown"' "$output_file" 2>/dev/null)
            local violations=$(jq -r '.summary.violations // .summary.license_violations // 0' "$output_file" 2>/dev/null)
            if [[ "$violations" -gt 0 ]]; then
                printf "${RED}%d violations${NC}" "$violations"
            elif [[ "$status" == "compliant" ]] || [[ "$status" == "clean" ]]; then
                printf "${GREEN}compliant ✓${NC}"
            else
                printf "%s" "$status"
            fi
            ;;
        tech-discovery)
            local count=$(jq -r '.technologies | length // 0' "$output_file" 2>/dev/null)
            local primary=$(jq -r '.primary_language // .summary.primary_language // empty' "$output_file" 2>/dev/null)
            printf "%d technologies" "$count"
            [[ -n "$primary" ]] && printf " ${DIM}(%s)${NC}" "$primary" || true
            ;;
        code-security)
            local total=$(jq -r '.summary.total // (.findings | length) // 0' "$output_file" 2>/dev/null)
            local high=$(jq -r '.summary.high // 0' "$output_file" 2>/dev/null)
            if [[ "$total" -gt 0 ]]; then
                [[ "$high" -gt 0 ]] && printf "${RED}%d high${NC} " "$high" || true
                printf "%d findings" "$total"
            else
                printf "${GREEN}0 issues ✓${NC}"
            fi
            ;;
        code-secrets)
            local total=$(jq -r '.summary.total // (.secrets | length) // 0' "$output_file" 2>/dev/null)
            if [[ "$total" -gt 0 ]]; then
                printf "${RED}%d secrets${NC}" "$total"
            else
                printf "${GREEN}0 secrets ✓${NC}"
            fi
            ;;
        package-malcontent)
            local total_files=$(jq -r '.summary.total_files // 0' "$output_file" 2>/dev/null)
            local critical=$(jq -r '.summary.by_risk.critical // 0' "$output_file" 2>/dev/null)
            local high=$(jq -r '.summary.by_risk.high // 0' "$output_file" 2>/dev/null)
            if [[ "$total_files" -gt 0 ]]; then
                printf "%d files " "$total_files"
                [[ "$critical" -gt 0 ]] && printf "${RED}%dC${NC} " "$critical" || true
                [[ "$high" -gt 0 ]] && printf "${YELLOW}%dH${NC}" "$high" || true
            else
                printf "${GREEN}0 findings ✓${NC}"
            fi
            ;;
        *)
            printf "done"
            ;;
    esac
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
        digital-certificates)
            analyzer_script="digital-certificates.sh"
            ;;
        package-malcontent)
            analyzer_script="package-malcontent/package-malcontent.sh"
            ;;
        bundle-analysis)
            analyzer_script="bundle-analysis/bundle-analysis.sh"
            ;;
        container-security)
            analyzer_script="container-security/container-security.sh"
            ;;
    esac

    zero_analysis_start "$project_id" "$analyzer" "$analyzer_script"

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
        digital-certificates)
            run_certificate_analyzer "$repo_path" "$output_path"
            ;;
        package-malcontent)
            run_malcontent_analyzer "$repo_path" "$output_path" "$project_id"
            ;;
        bundle-analysis)
            run_bundle_analyzer "$repo_path" "$output_path"
            ;;
        container-security)
            run_container_security_analyzer "$repo_path" "$output_path"
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

    zero_analysis_complete "$project_id" "$analyzer" "$status" "$duration" "$summary"

    return $exit_code
}

#############################################################################
# Individual Analyzer Wrappers
#############################################################################

run_technology_analyzer() {
    local repo_path="$1"
    local output_path="$2"

    local tech_script="$UTILS_ROOT/scanners/tech-discovery/tech-discovery.sh"

    if [[ -x "$tech_script" ]]; then
        local args=("--local-path" "$repo_path" "-o" "$output_path/tech-discovery.json")

        # Pass shared SBOM if available, fallback to local
        if [[ -n "$SHARED_SBOM_FILE" ]] && [[ -f "$SHARED_SBOM_FILE" ]]; then
            args+=("--sbom" "$SHARED_SBOM_FILE")
        elif [[ -f "$output_path/sbom.cdx.json" ]]; then
            args+=("--sbom" "$output_path/sbom.cdx.json")
        fi

        [[ "${USE_CLAUDE:-}" == "true" ]] && args+=("--claude")
        "$tech_script" "${args[@]}" 2>/dev/null
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

# Global SBOM file path for sharing with dependent scanners
SHARED_SBOM_FILE=""

run_dependency_extractor() {
    local repo_path="$1"
    local output_path="$2"

    local direct_count=0
    local total_count=0
    local sbom_format="none"
    local sbom_file=""
    local sbom_generator_used=""

    # Get SBOM generator for current profile
    local sbom_generator=$(get_profile_sbom_generator "$MODE")

    # Load SBOM library for smart generation
    if [[ -f "$UTILS_ROOT/lib/sbom.sh" ]]; then
        source "$UTILS_ROOT/lib/sbom.sh"
    fi

    sbom_file="$output_path/sbom.cdx.json"
    sbom_format="CycloneDX"

    # Use smart generator if available, fallback to syft directly
    if type generate_sbom_smart &>/dev/null; then
        sbom_generator_used="$sbom_generator"
        if generate_sbom_smart "$repo_path" "$sbom_file" "$sbom_generator" "true" 2>/dev/null; then
            # Count components from SBOM
            total_count=$(jq '.components | length // 0' "$sbom_file" 2>/dev/null)
            [[ -z "$total_count" ]] && total_count=0
            # Store path for dependent scanners
            SHARED_SBOM_FILE="$sbom_file"
            export SHARED_SBOM_FILE
        fi
    elif command -v syft &> /dev/null; then
        sbom_generator_used="syft"
        # Fallback to syft directly
        if syft scan "$repo_path" -o cyclonedx-json="$sbom_file" 2>/dev/null; then
            total_count=$(jq '.components | length // 0' "$sbom_file" 2>/dev/null)
            [[ -z "$total_count" ]] && total_count=0
            SHARED_SBOM_FILE="$sbom_file"
            export SHARED_SBOM_FILE
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

    # If no SBOM generated, use direct count as total
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
  "sbom_generator": "$sbom_generator_used",
  "sbom_file": "$(basename "$sbom_file" 2>/dev/null)",
  "direct_dependencies": $direct_count,
  "total_dependencies": $total_count,
  "summary": {
    "format": "$sbom_format",
    "generator": "$sbom_generator_used",
    "direct": $direct_count,
    "total": $total_count
  }
}
EOF
}

run_vulnerability_analyzer() {
    local repo_path="$1"
    local output_path="$2"

    local vuln_script="$UTILS_ROOT/scanners/package-vulns/package-vulns.sh"

    if [[ -x "$vuln_script" ]]; then
        local args=("--local-path" "$repo_path" "-o" "$output_path/package-vulns.json")
        [[ "${USE_CLAUDE:-}" == "true" ]] && args+=("--claude")
        # Pass shared SBOM if available
        if [[ -n "$SHARED_SBOM_FILE" ]] && [[ -f "$SHARED_SBOM_FILE" ]]; then
            args+=("--sbom" "$SHARED_SBOM_FILE")
        fi
        "$vuln_script" "${args[@]}" 2>/dev/null
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

    local legal_script="$UTILS_ROOT/scanners/licenses/licenses.sh"

    if [[ -x "$legal_script" ]]; then
        local args=("--local-path" "$repo_path" "-o" "$output_path/licenses.json")
        [[ "${USE_CLAUDE:-}" == "true" ]] && args+=("--claude")
        # Pass shared SBOM if available
        if [[ -n "$SHARED_SBOM_FILE" ]] && [[ -f "$SHARED_SBOM_FILE" ]]; then
            args+=("--sbom" "$SHARED_SBOM_FILE")
        fi
        "$legal_script" "${args[@]}" 2>/dev/null
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

    local prov_script="$UTILS_ROOT/scanners/package-provenance/package-provenance.sh"

    if [[ -x "$prov_script" ]]; then
        local args=("$repo_path" "-f" "json" "-o" "$output_path/package-provenance.json")

        # Pass shared SBOM if available
        if [[ -n "$SHARED_SBOM_FILE" ]] && [[ -f "$SHARED_SBOM_FILE" ]]; then
            args+=("--sbom" "$SHARED_SBOM_FILE")
        fi

        "$prov_script" "${args[@]}" 2>/dev/null || {
            # If it fails, fall back to basic git signature check
            local signed_commits=0
            if [[ -d "$repo_path/.git" ]]; then
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
    "note": "Full provenance analysis failed, showing basic git signature check."
  }
}
EOF
        }
    else
        # Fallback to basic git signature check
        local signed_commits=0
        if [[ -d "$repo_path/.git" ]]; then
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
    "note": "Basic git signature check only. Install provenance scanner for full SLSA verification."
  }
}
EOF
    fi
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
        local args=("--local-path" "$repo_path" "--repo-name" "$project_id" "-o" "$output_path/package-malcontent.json")

        # Pass shared SBOM if available, fallback to local
        if [[ -n "$SHARED_SBOM_FILE" ]] && [[ -f "$SHARED_SBOM_FILE" ]]; then
            args+=("--sbom" "$SHARED_SBOM_FILE")
        elif [[ -f "$output_path/sbom.cdx.json" ]]; then
            args+=("--sbom" "$output_path/sbom.cdx.json")
        fi

        # Run with verbose mode and show-findings only when not in org scan mode
        if [[ -z "$STATUS_DIR" ]]; then
            args+=("--verbose" "--show-findings")
        fi

        "$script" "${args[@]}" 2>&1
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

run_bundle_analyzer() {
    local repo_path="$1"
    local output_path="$2"

    local script="$UTILS_ROOT/scanners/bundle-analysis/bundle-analysis.sh"

    if [[ -x "$script" ]]; then
        "$script" --local-path "$repo_path" -o "$output_path/bundle-analysis.json" 2>/dev/null
    else
        cat > "$output_path/bundle-analysis.json" << EOF
{
  "analyzer": "bundle-analysis",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "status": "analyzer_not_found",
  "summary": {
    "total_dependencies": 0,
    "analyzed": 0,
    "total_size_bytes": 0
  },
  "packages": []
}
EOF
    fi
}

run_container_security_analyzer() {
    local repo_path="$1"
    local output_path="$2"

    local script="$UTILS_ROOT/scanners/container-security/container-security.sh"

    if [[ -x "$script" ]]; then
        "$script" --local-path "$repo_path" -o "$output_path/container-security.json" 2>/dev/null
    else
        cat > "$output_path/container-security.json" << EOF
{
  "analyzer": "container-security",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "status": "analyzer_not_found",
  "summary": {
    "dockerfiles_found": 0,
    "images_analyzed": 0,
    "total_vulnerabilities": 0
  },
  "dockerfiles": [],
  "images": []
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

    # Track progress for status updates
    local scanner_index=0
    local total_to_run=0

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

    # Count total scanners that will actually run
    total_to_run=$(echo "$analyzers_to_run" | wc -w | tr -d ' ')

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

            # Update status if status dir provided
            ((scanner_index++))
            if [[ -n "$STATUS_DIR" ]]; then
                update_repo_scan_status "$STATUS_DIR" "$project_id" "running" "$analyzer" "$scanner_index/$total_to_run" "0"
            fi

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

# Run analyzers in parallel phases
# Phase 1: package-sbom (generates SBOM for dependent scanners)
# Phase 2: All other scanners in parallel batches
run_all_analyzers_parallel() {
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
    local parallel_jobs=$(get_parallel_jobs)
    local scanner_timeout=$(get_scanner_timeout)

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

    # Only show header when not in org scan mode
    if [[ -z "$STATUS_DIR" ]]; then
        if [[ "$enrich" == "true" ]]; then
            echo -e "\n${BOLD}Running $profile_count analyzers${mode_display}${NC} ${CYAN}(enrichment, $parallel_jobs parallel)${NC}"
        else
            echo -e "\n${BOLD}Running $profile_count analyzers${mode_display}${NC} ${CYAN}($parallel_jobs parallel)${NC}"
        fi
        echo
    fi

    # Create temp directory for results
    local result_dir=$(mktemp -d)

    # Setup cleanup handler for background processes
    cleanup_parallel_scan() {
        local exit_code=$?

        # Clear any progress bar
        clear_progress_line 2>/dev/null || true

        # If interrupted (Ctrl+C), show cancellation message
        if [[ $exit_code -eq 130 ]] || [[ "${SCAN_INTERRUPTED:-}" == "true" ]]; then
            echo -e "\n${YELLOW}⚠${NC} Scan cancelled by user" >&2
        fi

        # Kill all background scanner jobs
        for pid in "${pids[@]}"; do
            kill "$pid" 2>/dev/null || true
        done
        # Wait briefly for jobs to terminate
        sleep 0.2
        # Force kill any remaining jobs
        for pid in "${pids[@]}"; do
            kill -9 "$pid" 2>/dev/null || true
        done
        # Clean up temp directories
        rm -rf "$result_dir" "$status_dir" "$buffer_dir" "$error_dir" 2>/dev/null || true
    }

    # Handle Ctrl+C gracefully
    handle_interrupt() {
        SCAN_INTERRUPTED=true
        exit 130
    }

    trap cleanup_parallel_scan EXIT
    trap handle_interrupt INT TERM

    # Phase 1: Run package-sbom first if needed (other scanners depend on it)
    local needs_sbom=false
    local has_sbom_scanner=false
    if scanner_in_list "package-sbom" "$analyzers_to_run"; then
        has_sbom_scanner=true
    fi
    for analyzer in $analyzers_to_run; do
        if scanner_needs_sbom "$analyzer"; then
            needs_sbom=true
            break
        fi
    done

    # Track total scanners for progress (will be calculated after remaining_scanners)
    local scanner_index=0
    local total_scanners_to_run=$(echo "$analyzers_to_run" | wc -w | tr -d ' ')

    if [[ "$has_sbom_scanner" == "true" ]]; then
        local display_name=$(get_scanner_display_name "package-sbom")

        # Only show inline progress when not in org scan mode
        if [[ -z "$STATUS_DIR" ]]; then
            printf "  ${WHITE}○${NC} %-24s ${DIM}running...${NC}" "$display_name"
        fi

        # Update global STATUS_DIR for scan.sh
        ((scanner_index++))
        if [[ -n "$STATUS_DIR" ]]; then
            update_repo_scan_status "$STATUS_DIR" "$project_id" "running" "package-sbom" "$scanner_index/$total_scanners_to_run" "0"
        fi

        local start_time=$(date +%s)
        run_analyzer "package-sbom" "$repo_path" "$output_path" "$project_id"
        local exit_code=$?
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))

        # Only show result line when not in org scan mode
        if [[ -z "$STATUS_DIR" ]]; then
            printf "\r\033[K"
            if [[ $exit_code -eq 0 ]]; then
                printf "  ${GREEN}✓${NC} %-24s " "$display_name"
                display_scanner_summary "package-sbom" "$output_path"
                printf " ${DIM}%ds${NC}\n" "$duration"
            else
                printf "  ${RED}✗${NC} %-24s ${RED}failed${NC} ${DIM}%ds${NC}\n" "$display_name" "$duration"
            fi
        fi
    fi

    # Build list of remaining scanners (excluding package-sbom)
    local remaining_scanners=""
    for analyzer in $analyzers_to_run; do
        [[ "$analyzer" == "package-sbom" ]] && continue
        remaining_scanners="$remaining_scanners $analyzer"
    done
    remaining_scanners=$(echo "$remaining_scanners" | xargs)  # trim

    if [[ -z "$remaining_scanners" ]]; then
        echo
        return 0
    fi

    # Initialize status display and output buffer
    local status_dir=$(init_scanner_status_display "$remaining_scanners")
    local buffer_dir=$(init_output_buffer)
    local error_dir=$(mktemp -d)

    # Only show progress bar mode message when not in org scan mode
    if [[ -z "$STATUS_DIR" ]]; then
        echo -e "${DIM}• Progress bar mode • Overall progress displayed •${NC}"
        echo
    fi

    local total_scanners=$(echo "$remaining_scanners" | wc -w | tr -d ' ')
    local completed_count=0

    # Phase 2: Run remaining scanners in parallel batches
    local pids=()
    local pid_to_analyzer=()
    local analyzer_start_times=()
    local running=0
    local next_analyzer_idx=0
    local analyzer_array=($remaining_scanners)

    # Start initial batch of scanners
    while [[ $next_analyzer_idx -lt ${#analyzer_array[@]} ]] && [[ $running -lt $parallel_jobs ]]; do
        local analyzer="${analyzer_array[$next_analyzer_idx]}"
        local start_time=$(date +%s)

        # Update status to running (local status_dir for this bootstrap run)
        update_scanner_status "$status_dir" "$analyzer" "running"

        # Update global STATUS_DIR for scan.sh
        ((scanner_index++))
        if [[ -n "$STATUS_DIR" ]]; then
            update_repo_scan_status "$STATUS_DIR" "$project_id" "running" "$analyzer" "$scanner_index/$total_scanners_to_run" "0"
        fi

        # Start scanner in background
        (
            local error_log="$error_dir/$analyzer.err"
            run_analyzer "$analyzer" "$repo_path" "$output_path" "$project_id" 2>"$error_log"
            local exit_code=$?
            local end_time=$(date +%s)
            local duration=$((end_time - start_time))

            # Get result summary
            local summary=$(display_scanner_summary "$analyzer" "$output_path" 2>/dev/null || echo "complete")

            if [[ $exit_code -eq 0 ]]; then
                update_scanner_status "$status_dir" "$analyzer" "complete" "$summary" "$duration"
            else
                # Capture last few lines of error for status
                local error_msg=""
                if [[ -s "$error_log" ]]; then
                    error_msg=$(tail -3 "$error_log" | tr '\n' ' ' | head -c 100)
                fi
                update_scanner_status "$status_dir" "$analyzer" "failed" "$error_msg" "$duration"
            fi
        ) &

        pids+=($!)
        pid_to_analyzer+=("$analyzer")
        analyzer_start_times+=("$start_time")
        ((running++))
        ((next_analyzer_idx++))
    done

    # Monitor and update progress bar
    local current_time
    while scanners_still_running "$status_dir" "$remaining_scanners"; do
        current_time=$(date +%s)

        # Check for per-scanner timeouts and kill timed-out scanners
        for i in "${!pids[@]}"; do
            local pid="${pids[$i]}"
            local analyzer="${pid_to_analyzer[$i]}"
            local start_time="${analyzer_start_times[$i]}"
            local elapsed=$((current_time - start_time))

            # Check if scanner has exceeded timeout
            if [[ $elapsed -gt $scanner_timeout ]] && kill -0 "$pid" 2>/dev/null; then
                # Kill the timed-out scanner and its children (using pkill to get child processes)
                # First kill all child processes of the subshell
                pkill -TERM -P "$pid" 2>/dev/null || true
                sleep 0.3
                pkill -KILL -P "$pid" 2>/dev/null || true
                # Then kill the subshell itself
                kill -TERM "$pid" 2>/dev/null || true
                sleep 0.2
                kill -KILL "$pid" 2>/dev/null || true

                # Mark as timeout in status file
                update_scanner_status "$status_dir" "$analyzer" "timeout" "exceeded ${scanner_timeout}s limit" "$elapsed"

                # Update manifest with timeout status
                local duration_ms=$((elapsed * 1000))
                zero_analysis_complete "$project_id" "$analyzer" "timeout" "$duration_ms" '{"error": "Scanner exceeded timeout limit"}'

                # Log timeout (only when not in org scan mode)
                if [[ -z "$STATUS_DIR" ]]; then
                    echo -e "\n${YELLOW}⚠ Scanner '$analyzer' timed out after ${elapsed}s (limit: ${scanner_timeout}s)${NC}" >&2
                fi

                # Remove from tracking
                unset 'pids[$i]'
                unset 'pid_to_analyzer[$i]'
                unset 'analyzer_start_times[$i]'
                ((running--))

                # Start next scanner if available
                if [[ $next_analyzer_idx -lt ${#analyzer_array[@]} ]]; then
                    local next_analyzer="${analyzer_array[$next_analyzer_idx]}"
                    local new_start_time=$(date +%s)

                    update_scanner_status "$status_dir" "$next_analyzer" "running"
                    ((scanner_index++))
                    if [[ -n "$STATUS_DIR" ]]; then
                        update_repo_scan_status "$STATUS_DIR" "$project_id" "running" "$next_analyzer" "$scanner_index/$total_scanners_to_run" "0"
                    fi

                    (
                        local error_log="$error_dir/$next_analyzer.err"
                        run_analyzer "$next_analyzer" "$repo_path" "$output_path" "$project_id" 2>"$error_log"
                        local exit_code=$?
                        local end_time=$(date +%s)
                        local duration=$((end_time - new_start_time))
                        local summary=$(display_scanner_summary "$next_analyzer" "$output_path" 2>/dev/null || echo "complete")

                        if [[ $exit_code -eq 0 ]]; then
                            update_scanner_status "$status_dir" "$next_analyzer" "complete" "$summary" "$duration"
                        else
                            local error_msg=""
                            [[ -s "$error_log" ]] && error_msg=$(tail -3 "$error_log" | tr '\n' ' ' | head -c 100)
                            update_scanner_status "$status_dir" "$next_analyzer" "failed" "$error_msg" "$duration"
                        fi
                    ) &

                    pids+=($!)
                    pid_to_analyzer+=("$next_analyzer")
                    analyzer_start_times+=("$new_start_time")
                    ((running++))
                    ((next_analyzer_idx++))
                fi
            fi
        done

        # Count completed scanners (including timeouts)
        completed_count=0
        for analyzer in $remaining_scanners; do
            local status_file="$status_dir/$analyzer.status"
            if [[ -f "$status_file" ]]; then
                local status=$(cut -d'|' -f1 "$status_file")
                if [[ "$status" == "complete" ]] || [[ "$status" == "failed" ]] || [[ "$status" == "timeout" ]]; then
                    ((completed_count++))
                fi
            fi
        done

        # Render progress bar (only when not in org scan mode)
        if [[ -z "$STATUS_DIR" ]]; then
            render_progress_bar "$completed_count" "$total_scanners" 50 "${CYAN}Scanning${NC}"
        fi

        # Check for completed jobs and start new ones
        for i in "${!pids[@]}"; do
            local pid="${pids[$i]}"
            if ! kill -0 "$pid" 2>/dev/null; then
                # Get analyzer info before removing from tracking
                local completed_analyzer="${pid_to_analyzer[$i]}"
                local completed_start_time="${analyzer_start_times[$i]}"

                # Check if status file was updated (if not, the process crashed)
                local status_file="$status_dir/$completed_analyzer.status"
                if [[ -f "$status_file" ]]; then
                    local current_status=$(cut -d'|' -f1 "$status_file")
                    if [[ "$current_status" == "running" ]]; then
                        # Process exited without updating status - mark as failed
                        local elapsed=$((current_time - completed_start_time))
                        update_scanner_status "$status_dir" "$completed_analyzer" "failed" "process exited unexpectedly" "$elapsed"
                        local duration_ms=$((elapsed * 1000))
                        zero_analysis_complete "$project_id" "$completed_analyzer" "failed" "$duration_ms" '{"error": "Process exited unexpectedly"}'
                    fi
                fi

                # Job completed, remove from tracking
                unset 'pids[$i]'
                unset 'pid_to_analyzer[$i]'
                unset 'analyzer_start_times[$i]'
                ((running--))

                # Start next scanner if available
                if [[ $next_analyzer_idx -lt ${#analyzer_array[@]} ]]; then
                    local analyzer="${analyzer_array[$next_analyzer_idx]}"
                    local start_time=$(date +%s)

                    # Update status to running (local status_dir for this bootstrap run)
                    update_scanner_status "$status_dir" "$analyzer" "running"

                    # Update global STATUS_DIR for scan.sh
                    ((scanner_index++))
                    if [[ -n "$STATUS_DIR" ]]; then
                        update_repo_scan_status "$STATUS_DIR" "$project_id" "running" "$analyzer" "$scanner_index/$total_scanners_to_run" "0"
                    fi

                    # Start scanner in background
                    (
                        local error_log="$error_dir/$analyzer.err"
                        run_analyzer "$analyzer" "$repo_path" "$output_path" "$project_id" 2>"$error_log"
                        local exit_code=$?
                        local end_time=$(date +%s)
                        local duration=$((end_time - start_time))

                        # Get result summary
                        local summary=$(display_scanner_summary "$analyzer" "$output_path" 2>/dev/null || echo "complete")

                        if [[ $exit_code -eq 0 ]]; then
                            update_scanner_status "$status_dir" "$analyzer" "complete" "$summary" "$duration"
                        else
                            # Capture last few lines of error for status
                            local error_msg=""
                            if [[ -s "$error_log" ]]; then
                                error_msg=$(tail -3 "$error_log" | tr '\n' ' ' | head -c 100)
                            fi
                            update_scanner_status "$status_dir" "$analyzer" "failed" "$error_msg" "$duration"
                        fi
                    ) &

                    pids+=($!)
                    pid_to_analyzer+=("$analyzer")
                    analyzer_start_times+=("$start_time")
                    ((running++))
                    ((next_analyzer_idx++))
                fi
            fi
        done

        # Compact arrays (remove empty slots)
        pids=("${pids[@]}")
        pid_to_analyzer=("${pid_to_analyzer[@]}")
        analyzer_start_times=("${analyzer_start_times[@]}")

        sleep 0.1
    done

    # Clear progress bar and show completed results (only when not in org scan mode)
    if [[ -z "$STATUS_DIR" ]]; then
        clear_progress_line
        echo -e "${GREEN}✓${NC} Scanning complete"
        echo
    fi

    # Display all scanner results in grouped format (only when not in org scan mode)
    if [[ -z "$STATUS_DIR" ]]; then
        for analyzer in $remaining_scanners; do
            display_scanner_result "$analyzer" "$output_path" "$result_dir" "$status_dir" "$error_dir"
        done
    fi

    # Show scanners not in profile (only when not in org scan mode)
    if [[ -z "$STATUS_DIR" ]]; then
        for analyzer in $ALL_SCANNERS; do
            if ! scanner_in_list "$analyzer" "$requested_analyzers"; then
                local display_name=$(get_scanner_display_name "$analyzer")
                printf "  ${DIM}○ %-24s not in profile${NC}\n" "$display_name"
            fi
        done
        echo
    fi
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
    zero_update_summary "$project_id" "$risk_level" "$total_deps" "$direct_deps" "$total_vulns" "$total_findings" "$license_status" "$abandoned"
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
    local size=$(zero_project_size "$project_id")
    echo -e "Storage: ${CYAN}$ZERO_DIR/repos/$project_id/${NC} ($size)"
}

#############################################################################
# Main
#############################################################################

main() {
    [[ -n "${DEBUG_BOOTSTRAP:-}" ]] && echo "[DEBUG] main() started" >&2
    parse_args "$@"
    [[ -n "${DEBUG_BOOTSTRAP:-}" ]] && echo "[DEBUG] parse_args() completed" >&2

    # Run preflight check (silent unless errors)
    if [[ -x "$SCRIPT_DIR/preflight.sh" ]]; then
        [[ -n "${DEBUG_BOOTSTRAP:-}" ]] && echo "[DEBUG] Running preflight check..." >&2
        if ! "$SCRIPT_DIR/preflight.sh" > /dev/null 2>&1; then
            echo -e "${RED}Preflight check failed. Run ./utils/zero/preflight.sh to see details.${NC}"
            exit 1
        fi
        [[ -n "${DEBUG_BOOTSTRAP:-}" ]] && echo "[DEBUG] Preflight check passed" >&2
    fi

    # Regenerate Semgrep rules from RAG patterns (ensures latest rules)
    local rag_to_semgrep="$UTILS_ROOT/scanners/semgrep/rag-to-semgrep.py"
    local rag_dir="$REPO_ROOT/rag/technology-identification"
    local semgrep_rules_dir="$UTILS_ROOT/scanners/semgrep/rules"
    if [[ -x "$rag_to_semgrep" ]] || [[ -f "$rag_to_semgrep" ]]; then
        if command -v python3 &> /dev/null && [[ -d "$rag_dir" ]]; then
            [[ -n "${DEBUG_BOOTSTRAP:-}" ]] && echo "[DEBUG] Regenerating Semgrep rules..." >&2
            python3 "$rag_to_semgrep" "$rag_dir" "$semgrep_rules_dir" > /dev/null 2>&1 || true
            [[ -n "${DEBUG_BOOTSTRAP:-}" ]] && echo "[DEBUG] Semgrep rules regenerated" >&2
        fi
    fi

    # Sync Semgrep community rules (cached, only updates if stale)
    local community_rules="$UTILS_ROOT/scanners/semgrep/community-rules.sh"
    if [[ -x "$community_rules" ]] && command -v semgrep &> /dev/null; then
        [[ -n "${DEBUG_BOOTSTRAP:-}" ]] && echo "[DEBUG] Syncing Semgrep community rules..." >&2
        "$community_rules" sync default > /dev/null 2>&1 || true
        [[ -n "${DEBUG_BOOTSTRAP:-}" ]] && echo "[DEBUG] Semgrep community rules synced" >&2
    fi

    # Ensure Gibson is initialized
    zero_ensure_initialized

    # Generate project ID
    local project_id=$(zero_project_id "$TARGET")

    # Setup paths
    local project_path=$(zero_project_path "$project_id")
    local repo_path=$(zero_project_repo_path "$project_id")
    local analysis_path=$(zero_project_analysis_path "$project_id")

    # Handle enrich mode - project must already exist
    if [[ "$ENRICH" == "true" ]]; then
        if ! zero_project_exists "$project_id"; then
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
            git_context=$(zero_get_git_context "$repo_path")
        fi

        # Print header for enrichment
        zero_print_header
        echo -e "Enriching: ${CYAN}$TARGET${NC}"
        echo -e "Project ID: ${CYAN}$project_id${NC}"
        echo -e "Scan ID: ${DIM}$scan_id${NC}"

        # Get commit short for directory name
        local commit_short=$(echo "$git_context" | jq -r '.commit_short // ""' 2>/dev/null)

        # Initialize versioned scan directory (analysis/scans/<scan_id>_<commit>/)
        local scan_output_path=$(zero_init_scan_directory "$project_id" "$scan_id" "$commit_short")

        # Skip cloning, just run missing analyzers - output to versioned directory
        if [[ "$PARALLEL" == "true" ]]; then
            run_all_analyzers_parallel "$repo_path" "$scan_output_path" "$project_id" "$MODE" "true"
        else
            run_all_analyzers "$repo_path" "$scan_output_path" "$project_id" "$MODE" "true"
        fi

        # Copy results to analysis root for backwards compatibility
        for f in "$scan_output_path"/*.json; do
            [[ -f "$f" ]] && cp "$f" "$analysis_path/" 2>/dev/null || true
        done
        [[ -f "$scan_output_path/sbom.cdx.json" ]] && cp "$scan_output_path/sbom.cdx.json" "$analysis_path/" 2>/dev/null || true

        # Update summary and finalize
        generate_summary "$project_id" "$analysis_path"
        zero_finalize_manifest "$project_id"

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

        zero_append_scan_history "$project_id" "$scan_id" "$commit_hash" "$commit_short" "$git_branch" "$scan_start_time" "$scan_end_time" "$duration_seconds" "$MODE" "$scanners_json" "complete" "{}"

        # Update org-level index
        zero_update_org_index "$project_id"

        zero_index_update_status "$project_id" "ready"
        zero_set_active_project "$project_id"
        print_final_summary "$project_id" "$analysis_path"
        return 0
    fi

    # Handle scan-only mode - project must already be cloned
    if [[ "$SCAN_ONLY" == "true" ]]; then
        if ! zero_project_exists "$project_id"; then
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
            git_context=$(zero_get_git_context "$repo_path")
        fi

        # Print header for scan (suppress in org scan mode - STATUS_DIR is set)
        if [[ -z "$STATUS_DIR" ]]; then
            zero_print_header
            echo -e "Scanning: ${CYAN}$TARGET${NC}"
            echo -e "Project ID: ${CYAN}$project_id${NC}"
            echo -e "Mode: ${CYAN}$MODE${NC}"
            echo -e "Scan ID: ${DIM}$scan_id${NC}"
        fi

        # Create analysis directory if needed
        mkdir -p "$analysis_path"

        # Initialize/update analysis manifest
        local commit=$(echo "$git_context" | jq -r '.commit_hash // ""' 2>/dev/null)
        local commit_short=$(echo "$git_context" | jq -r '.commit_short // ""' 2>/dev/null)

        # Initialize versioned scan directory (analysis/scans/<scan_id>_<commit>/)
        local scan_output_path=$(zero_init_scan_directory "$project_id" "$scan_id" "$commit_short")

        zero_init_analysis_manifest "$project_id" "$commit" "$MODE" "$scan_id" "$git_context"

        # Run analyzers - output to versioned scan directory
        if [[ "$PARALLEL" == "true" ]]; then
            run_all_analyzers_parallel "$repo_path" "$scan_output_path" "$project_id" "$MODE"
        else
            run_all_analyzers "$repo_path" "$scan_output_path" "$project_id" "$MODE"
        fi

        # Copy results to analysis root for backwards compatibility
        for f in "$scan_output_path"/*.json; do
            [[ -f "$f" ]] && cp "$f" "$analysis_path/" 2>/dev/null || true
        done
        # Also copy sbom.cdx.json if it exists
        [[ -f "$scan_output_path/sbom.cdx.json" ]] && cp "$scan_output_path/sbom.cdx.json" "$analysis_path/" 2>/dev/null || true

        # Update summary and finalize
        generate_summary "$project_id" "$analysis_path"
        zero_finalize_manifest "$project_id"

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

        zero_append_scan_history "$project_id" "$scan_id" "$commit_hash" "$commit_short" "$git_branch" "$scan_start_time" "$scan_end_time" "$duration_seconds" "$MODE" "$scanners_json" "complete" "{}"

        zero_update_org_index "$project_id"
        zero_index_update_status "$project_id" "ready"
        zero_set_active_project "$project_id"
        # Only show final summary when not in org scan mode
        if [[ -z "$STATUS_DIR" ]]; then
            print_final_summary "$project_id" "$analysis_path"
        fi
        return 0
    fi

    # Check if project already exists
    if zero_project_exists "$project_id" && [[ "$FORCE" != "true" ]]; then
        echo -e "${YELLOW}Project '$project_id' already exists.${NC}"
        echo "Use --force to re-bootstrap, or --enrich to add missing analyzers."
        exit 1
    fi

    # Print header
    zero_print_header
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
    zero_index_add_project "$project_id" "$TARGET" "bootstrapping"

    # Clone or copy project
    echo -n "Cloning..."
    if zero_is_local_source "$TARGET"; then
        copy_local_project "$TARGET" "$repo_path"
        echo -e " ${GREEN}✓${NC} (local copy)"
    else
        local clone_url=$(zero_clone_url "$TARGET")
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
            zero_index_remove_project "$project_id"
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
        zero_is_local_source "$TARGET" && source_type="local"
        zero_create_project_metadata "$project_id" "$TARGET" "$source_type" "$branch" "$commit"
        zero_index_update_status "$project_id" "cloned"

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
        git_context=$(zero_get_git_context "$repo_path")
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
    zero_is_local_source "$TARGET" && source_type="local"

    # Create project metadata
    zero_create_project_metadata "$project_id" "$TARGET" "$source_type" "$branch" "$commit"
    zero_update_project_type "$project_id" "$languages" "$frameworks" "$package_managers"

    # Get commit short for directory name
    local commit_short=$(echo "$git_context" | jq -r '.commit_short // ""' 2>/dev/null)

    # Initialize versioned scan directory (analysis/scans/<scan_id>_<commit>/)
    local scan_output_path=$(zero_init_scan_directory "$project_id" "$scan_id" "$commit_short")

    # Initialize analysis manifest (with mode, scan_id, and git context)
    zero_init_analysis_manifest "$project_id" "$commit" "$MODE" "$scan_id" "$git_context"

    # Run analyzers - output to versioned scan directory
    if [[ "$PARALLEL" == "true" ]]; then
        run_all_analyzers_parallel "$repo_path" "$scan_output_path" "$project_id" "$MODE"
    else
        run_all_analyzers "$repo_path" "$scan_output_path" "$project_id" "$MODE"
    fi

    # Copy results to analysis root for backwards compatibility
    for f in "$scan_output_path"/*.json; do
        [[ -f "$f" ]] && cp "$f" "$analysis_path/" 2>/dev/null || true
    done
    # Also copy sbom.cdx.json if it exists
    [[ -f "$scan_output_path/sbom.cdx.json" ]] && cp "$scan_output_path/sbom.cdx.json" "$analysis_path/" 2>/dev/null || true

    # Generate summary
    generate_summary "$project_id" "$analysis_path"

    # Finalize manifest
    zero_finalize_manifest "$project_id"

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

    zero_append_scan_history "$project_id" "$scan_id" "$commit_hash" "$commit_short" "$git_branch" "$scan_start_time" "$scan_end_time" "$duration_seconds" "$MODE" "$scanners_json" "complete" "{}"

    # Update org-level index for agent queries
    zero_update_org_index "$project_id"

    # Update index status
    zero_index_update_status "$project_id" "ready"

    # Set as active project
    zero_set_active_project "$project_id"

    # Print final summary
    print_final_summary "$project_id" "$analysis_path"
}

main "$@"
