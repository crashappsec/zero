#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Phantom Scan
# Run scanners to gather enrichment data from cloned repositories
#
# Usage:
#   ./scan.sh <owner/repo>           # Single repo
#   ./scan.sh --org <org-name>       # All cloned repos in an org
#
# Examples:
#   ./scan.sh expressjs/express
#   ./scan.sh expressjs/express --quick
#   ./scan.sh --org expressjs --standard
#############################################################################

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ZERO_UTILS_DIR="$(dirname "$SCRIPT_DIR")"

# Load Zero library (sets ZERO_DIR to ~/.zero data directory)
source "$ZERO_UTILS_DIR/lib/zero-lib.sh"

# Load config loader for dynamic profiles
source "$ZERO_UTILS_DIR/config/config-loader.sh"

# Load .env if available
UTILS_ROOT="$(dirname "$ZERO_UTILS_DIR")"
REPO_ROOT="$(dirname "$UTILS_ROOT")"
SCANNERS_DIR="$UTILS_ROOT/scanners"

if [[ -f "$REPO_ROOT/.env" ]]; then
    set -a
    source "$REPO_ROOT/.env"
    set +a
fi

#############################################################################
# Configuration
#############################################################################

ORG_MODE=false
ORG_NAME=""
TARGET=""
PROFILE="$(get_default_profile)"
FORCE=false

# All scanners loaded from config (with fallback)
ALL_SCANNERS="$(get_all_scanners)"

#############################################################################
# Usage
#############################################################################

usage() {
    cat << EOF
Phantom Scan - Run scanners to gather enrichment data

Usage: $0 <target> [options]
       $0 --org <org-name> [options]

MODES:
    Single Repo:    $0 owner/repo [options]
    Organization:   $0 --org <org-name> [options]

PROFILES (from zero.config.json):
EOF
    print_profile_help
    cat << EOF

OPTIONS:
    --org <name>    Scan all cloned repos in a GitHub organization
    --force         Re-scan even if results exist
    -h, --help      Show this help

EXAMPLES:
    $0 expressjs/express                    # Standard scan
    $0 expressjs/express --quick            # Quick scan
    $0 --org expressjs --security           # Security scan all org repos

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
            --org)
                ORG_MODE=true
                ORG_NAME="$2"
                shift 2
                ;;
            --force)
                FORCE=true
                shift
                ;;
            --*)
                # Try to match as a profile name
                local profile_name="${1#--}"
                if [[ " $available_profiles " =~ " $profile_name " ]]; then
                    PROFILE="$profile_name"
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

    # Validate arguments
    if [[ "$ORG_MODE" == "true" ]]; then
        if [[ -z "$ORG_NAME" ]]; then
            echo -e "${RED}Error: --org requires an organization name${NC}" >&2
            exit 1
        fi
    elif [[ -z "$TARGET" ]]; then
        echo -e "${RED}Error: No target specified${NC}" >&2
        echo "Usage: $0 <owner/repo> or $0 --org <org-name>"
        exit 1
    fi
}

#############################################################################
# Scanner Functions
#############################################################################

# Check if scanner is in current profile (uses dynamic config)
# Note: scanner_in_profile is already defined in config-loader.sh

# Get scanner display name (uses dynamic config)
get_scanner_display() {
    local scanner="$1"
    get_scanner_name "$scanner"
}

# Get scanner output file (uses dynamic config)
get_scanner_output() {
    local scanner="$1"
    local analysis_path="$2"
    local output_file=$(get_scanner_output_file "$scanner")
    echo "$analysis_path/$output_file"
}

# Run a single scanner
run_scanner() {
    local scanner="$1"
    local repo_path="$2"
    local analysis_path="$3"

    # Get script path from config (handles scanners in subdirectories)
    local script_rel=$(get_scanner_script "$scanner")
    local script_path="$REPO_ROOT/$script_rel"
    local output_file=$(get_scanner_output "$scanner" "$analysis_path")

    # Check if scanner exists
    if [[ ! -x "$script_path" ]]; then
        return 1
    fi

    # Build args
    local args=("--local-path" "$repo_path")

    # Pass SBOM for scanners that need it
    if [[ -f "$analysis_path/sbom.cdx.json" ]]; then
        case "$scanner" in
            tech-discovery|package-vulns|package-health|licenses)
                args+=("--sbom" "$analysis_path/sbom.cdx.json")
                ;;
        esac
    fi

    # Output file
    args+=("-o" "$output_file")

    # Run scanner
    "$script_path" "${args[@]}" 2>/dev/null
}

# Get result summary from scanner output
get_scanner_result() {
    local scanner="$1"
    local analysis_path="$2"
    local output_file=$(get_scanner_output "$scanner" "$analysis_path")

    if [[ ! -f "$output_file" ]]; then
        echo ""
        return
    fi

    case "$scanner" in
        package-sbom)
            local count=$(jq -r '.components | length // 0' "$output_file" 2>/dev/null)
            echo "$count packages"
            ;;
        tech-discovery)
            local count=$(jq -r '.technologies | length // 0' "$output_file" 2>/dev/null)
            echo "$count technologies"
            ;;
        package-vulns)
            local count=$(jq -r '.summary.total // .vulnerabilities | length // 0' "$output_file" 2>/dev/null)
            echo "$count found"
            ;;
        licenses)
            local status=$(jq -r '.summary.status // "unknown"' "$output_file" 2>/dev/null)
            echo "$status"
            ;;
        code-security)
            local count=$(jq -r '.summary.total // .findings | length // 0' "$output_file" 2>/dev/null)
            echo "$count findings"
            ;;
        code-secrets)
            local count=$(jq -r '.summary.total // .secrets | length // 0' "$output_file" 2>/dev/null)
            echo "$count found"
            ;;
        tech-debt)
            local score=$(jq -r '.summary.score // "unknown"' "$output_file" 2>/dev/null)
            echo "score: $score"
            ;;
        code-ownership)
            local owners=$(jq -r '.summary.total_owners // 0' "$output_file" 2>/dev/null)
            echo "$owners owners"
            ;;
        dora)
            local freq=$(jq -r '.summary.deployment_frequency // "unknown"' "$output_file" 2>/dev/null)
            echo "$freq"
            ;;
        package-malcontent)
            local files=$(jq -r '.summary.total_files // 0' "$output_file" 2>/dev/null)
            local critical=$(jq -r '.summary.by_risk.critical // 0' "$output_file" 2>/dev/null)
            local high=$(jq -r '.summary.by_risk.high // 0' "$output_file" 2>/dev/null)
            if [[ "$critical" != "0" ]] || [[ "$high" != "0" ]]; then
                echo "$files files, ${critical}C/${high}H"
            else
                echo "$files files"
            fi
            ;;
        *)
            echo "complete"
            ;;
    esac
}

#############################################################################
# Scan Functions
#############################################################################

# Scan a single repository using bootstrap.sh (which has proven scanner implementations)
scan_repo() {
    local repo="$1"
    local project_id=$(zero_project_id "$repo")
    local repo_path="$ZERO_PROJECTS_DIR/$project_id/repo"

    # Check if repo is cloned
    if [[ ! -d "$repo_path" ]]; then
        echo -e "  ${RED}✗${NC} Repository not cloned"
        echo -e "    Run: ${CYAN}./zero.sh clone $repo${NC}"
        return 1
    fi

    # Build bootstrap args
    local bootstrap_args=("--scan-only" "--$PROFILE")
    [[ "$FORCE" == "true" ]] && bootstrap_args+=("--force")
    bootstrap_args+=("$repo")

    # Delegate to bootstrap.sh which has the proven scanner implementations
    "$SCRIPT_DIR/bootstrap.sh" "${bootstrap_args[@]}"
}

# Update analysis manifest
update_manifest() {
    local project_id="$1"
    local analysis_path="$2"
    local manifest="$analysis_path/manifest.json"

    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Build analyses object
    local analyses="{}"
    for scanner in $ALL_SCANNERS; do
        local output_file=$(get_scanner_output "$scanner" "$analysis_path")
        if [[ -f "$output_file" ]]; then
            local mtime=$(stat -f %m "$output_file" 2>/dev/null || stat -c %Y "$output_file" 2>/dev/null)
            analyses=$(echo "$analyses" | jq --arg s "$scanner" --arg t "$timestamp" '. + {($s): {"status": "complete", "completed_at": $t}}')
        fi
    done

    # Write manifest
    jq -n \
        --arg pid "$project_id" \
        --arg mode "$PROFILE" \
        --arg ts "$timestamp" \
        --argjson analyses "$analyses" \
        '{
            project_id: $pid,
            mode: $mode,
            completed_at: $ts,
            analyses: $analyses
        }' > "$manifest"
}

#############################################################################
# Main Functions
#############################################################################

scan_single() {
    local repo="$1"
    local project_id=$(zero_project_id "$repo")
    local repo_path="$ZERO_PROJECTS_DIR/$project_id/repo"

    # Check if repo is cloned
    if [[ ! -d "$repo_path" ]]; then
        print_zero_banner
        echo -e "${RED}Error: Repository not cloned${NC}"
        echo -e "Run: ${CYAN}./zero.sh clone $repo${NC}"
        return 1
    fi

    # Build bootstrap args and delegate to bootstrap.sh
    local bootstrap_args=("--scan-only" "--$PROFILE")
    [[ "$FORCE" == "true" ]] && bootstrap_args+=("--force")
    bootstrap_args+=("$repo")

    exec "$SCRIPT_DIR/bootstrap.sh" "${bootstrap_args[@]}"
}

scan_org() {
    local org="$1"

    print_zero_banner
    echo -e "${BOLD}Scan Organization${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo

    # Find cloned repos in org
    local org_path="$ZERO_PROJECTS_DIR/$org"
    if [[ ! -d "$org_path" ]]; then
        echo -e "${RED}No cloned repos found for org: $org${NC}" >&2
        echo -e "Clone first with: ${CYAN}./zero.sh clone --org $org${NC}"
        exit 1
    fi

    local repos=()
    for repo_dir in "$org_path"/*/; do
        [[ ! -d "$repo_dir" ]] && continue
        [[ ! -d "$repo_dir/repo" ]] && continue
        local repo_name=$(basename "$repo_dir")
        repos+=("$org/$repo_name")
    done

    local repo_count=${#repos[@]}
    if [[ $repo_count -eq 0 ]]; then
        echo -e "${RED}No cloned repos found in: $org_path${NC}" >&2
        exit 1
    fi

    # Get parallel jobs from config
    local parallel_jobs=$(get_parallel_jobs)

    echo -e "Organization: ${CYAN}$org${NC}"
    echo -e "Repositories: ${CYAN}$repo_count${NC}"
    echo -e "Profile:      ${CYAN}$PROFILE${NC}"
    echo -e "Parallel:     ${CYAN}$parallel_jobs jobs${NC}"
    echo
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    # Check if GNU parallel is available for better parallelization
    if command -v parallel &> /dev/null && [[ "$parallel_jobs" -gt 1 ]]; then
        scan_org_parallel "$org" "${repos[@]}"
    else
        scan_org_sequential "$org" "${repos[@]}"
    fi
}

# Scan org repos in parallel using GNU parallel
scan_org_parallel() {
    local org="$1"
    shift
    local repos=("$@")
    local repo_count=${#repos[@]}
    local parallel_jobs=$(get_parallel_jobs)

    echo -e "${CYAN}Running parallel scan with $parallel_jobs concurrent jobs...${NC}"
    echo

    # Create temp dir for results
    local tmp_dir=$(mktemp -d)
    trap "rm -rf $tmp_dir" EXIT

    # Export variables for parallel
    export SCRIPT_DIR PROFILE FORCE

    # Run scans in parallel
    printf '%s\n' "${repos[@]}" | parallel -j "$parallel_jobs" --bar --tag \
        "$SCRIPT_DIR/bootstrap.sh" --scan-only --"$PROFILE" $([ "$FORCE" == "true" ] && echo "--force") {} \
        '>' "$tmp_dir/{#}.log" '2>&1' ';' \
        'echo $? > '"$tmp_dir/{#}.exit"

    # Count results
    local success=0
    local failed=0
    for i in $(seq 1 $repo_count); do
        if [[ -f "$tmp_dir/$i.exit" ]] && [[ $(cat "$tmp_dir/$i.exit") -eq 0 ]]; then
            ((success++))
        else
            ((failed++))
        fi
    done

    echo
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${GREEN}✓ Complete${NC}: $success scanned, $failed failed"
    echo
    echo -e "View results: ${CYAN}./zero.sh report --org $org${NC}"
}

# Scan org repos sequentially with background job parallelization
scan_org_sequential() {
    local org="$1"
    shift
    local repos=("$@")
    local repo_count=${#repos[@]}
    local parallel_jobs=$(get_parallel_jobs)

    local success=0
    local failed=0
    local current=0
    local pids=()
    local repo_map=()

    # Create temp dir for logs
    local tmp_dir=$(mktemp -d)
    trap "rm -rf $tmp_dir" EXIT

    for repo in "${repos[@]}"; do
        ((current++))

        # Build bootstrap args
        local bootstrap_args=("--scan-only" "--$PROFILE")
        [[ "$FORCE" == "true" ]] && bootstrap_args+=("--force")
        bootstrap_args+=("$repo")

        # Start background job
        (
            "$SCRIPT_DIR/bootstrap.sh" "${bootstrap_args[@]}" > "$tmp_dir/$current.log" 2>&1
            echo $? > "$tmp_dir/$current.exit"
        ) &

        pids+=($!)
        repo_map+=("$repo")

        echo -e "${BOLD}[$current/$repo_count]${NC} ${CYAN}$repo${NC} ${DIM}(started)${NC}"

        # Limit concurrent jobs
        if [[ ${#pids[@]} -ge $parallel_jobs ]]; then
            # Wait for oldest job to complete (ignore exit code - we check logs later)
            wait "${pids[0]}" 2>/dev/null || true
            pids=("${pids[@]:1}")
        fi
    done

    # Wait for remaining jobs
    echo
    echo -e "${DIM}Waiting for remaining scans to complete...${NC}"
    for pid in "${pids[@]}"; do
        wait "$pid" 2>/dev/null || true
    done

    # Count results and show summary
    for i in $(seq 1 $repo_count); do
        if [[ -f "$tmp_dir/$i.exit" ]] && [[ $(cat "$tmp_dir/$i.exit") -eq 0 ]]; then
            ((success++))
            echo -e "  ${GREEN}✓${NC} ${repo_map[$((i-1))]}"
        else
            ((failed++))
            echo -e "  ${RED}✗${NC} ${repo_map[$((i-1))]}"
        fi
    done

    echo
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${GREEN}✓ Complete${NC}: $success scanned, $failed failed"
    echo
    echo -e "View results: ${CYAN}./zero.sh report --org $org${NC}"
}

#############################################################################
# Main
#############################################################################

main() {
    parse_args "$@"

    if [[ "$ORG_MODE" == "true" ]]; then
        scan_org "$ORG_NAME"
    else
        scan_single "$TARGET"
    fi
}

main "$@"
