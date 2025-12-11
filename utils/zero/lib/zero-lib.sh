#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Zero Directory Management Library
# Manages .zero/ directory structure for Zero orchestrator
# Named after Zero Cool from the movie Hackers (1995)
#############################################################################

# Zero root directory - defaults to .zero in project root
# Can be overridden with ZERO_HOME environment variable
_ZERO_LIB_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
_ZERO_REPO_ROOT="$(dirname "$(dirname "$(dirname "$_ZERO_LIB_DIR")")")"
export ZERO_DIR="${ZERO_HOME:-$_ZERO_REPO_ROOT/.zero}"
export ZERO_REPOS_DIR="$ZERO_DIR/repos"
# Legacy alias for compatibility
export ZERO_PROJECTS_DIR="$ZERO_REPOS_DIR"
export ZERO_VERSION="1.0.0"

# Color codes (using ANSI-C quoting for proper escape sequence interpretation)
RED=$'\033[0;31m'
GREEN=$'\033[0;32m'
YELLOW=$'\033[1;33m'
BLUE=$'\033[0;34m'
CYAN=$'\033[0;36m'
MAGENTA=$'\033[0;35m'
WHITE=$'\033[0;37m'
BOLD=$'\033[1m'
DIM=$'\033[2m'
NC=$'\033[0m'

#############################################################################
# ZERO Banner
#############################################################################

# Raw banner text (no colors, for piping to effects)
ZERO_BANNER='  ███████╗███████╗██████╗  ██████╗
  ╚══███╔╝██╔════╝██╔══██╗██╔═══██╗
    ███╔╝ █████╗  ██████╔╝██║   ██║
   ███╔╝  ██╔══╝  ██╔══██╗██║   ██║
  ███████╗███████╗██║  ██║╚██████╔╝
  ╚══════╝╚══════╝╚═╝  ╚═╝ ╚═════╝
  crashoverride.com'

# Check if terminal text effects is available
has_tte() {
    python3 -c "import terminaltexteffects" 2>/dev/null
}

# Available effects for random selection
ZERO_EFFECTS=(burn decrypt slide)

# Print animated banner using terminal text effects
# Falls back to static banner if tte not available
print_zero_banner_animated() {
    local effect="${1:-random}"

    # Pick random effect if not specified or "random"
    if [[ "$effect" == "random" ]]; then
        effect="${ZERO_EFFECTS[$RANDOM % ${#ZERO_EFFECTS[@]}]}"
    fi

    # Check if tte is available and terminal is interactive
    if has_tte && [[ -t 1 ]]; then
        case "$effect" in
            burn)
                echo "$ZERO_BANNER" | python3 -m terminaltexteffects burn \
                    --starting-color 444444 \
                    --burn-colors ff6600 ff3300 cc0000 660000 \
                    --final-gradient-stops 00ff00 00cc00 \
                    --final-gradient-steps 8 \
                    --final-gradient-direction vertical 2>/dev/null
                ;;
            decrypt)
                echo "$ZERO_BANNER" | python3 -m terminaltexteffects decrypt \
                    --typing-speed 2 \
                    --ciphertext-colors 00ff00 00cc00 009900 \
                    --final-gradient-stops 00ff00 00cc00 \
                    --final-gradient-steps 8 2>/dev/null
                ;;
            slide)
                echo "$ZERO_BANNER" | python3 -m terminaltexteffects slide \
                    --movement-speed 0.5 \
                    --grouping row \
                    --final-gradient-stops 00ff00 00cc00 \
                    --final-gradient-direction vertical 2>/dev/null
                ;;
            *)
                # Unknown effect, use static
                print_zero_banner
                return
                ;;
        esac
        echo
    else
        # Fallback to static banner
        print_zero_banner
    fi
}

# Print static banner (original behavior)
print_zero_banner() {
    echo -e "${GREEN}"
    cat << 'BANNER'
  ███████╗███████╗██████╗  ██████╗
  ╚══███╔╝██╔════╝██╔══██╗██╔═══██╗
    ███╔╝ █████╗  ██████╔╝██║   ██║
   ███╔╝  ██╔══╝  ██╔══██╗██║   ██║
  ███████╗███████╗██║  ██║╚██████╔╝
  ╚══════╝╚══════╝╚═╝  ╚═╝ ╚═════╝
BANNER
    echo -e "${DIM}  crashoverride.com${NC}"
    echo
}

# Backwards compatibility aliases
print_phantom_banner() { print_zero_banner; }
print_phantom_banner_animated() { print_zero_banner_animated "$@"; }

#############################################################################
# Directory Initialization
#############################################################################

# Initialize .zero/ directory structure
# Creates all necessary directories and config files if they don't exist
zero_init() {
    local force="${1:-false}"

    # Check if already initialized
    if [[ -f "$ZERO_DIR/config.json" ]] && [[ "$force" != "true" ]]; then
        return 0
    fi

    echo -e "${CYAN}Initializing Zero directory at $ZERO_DIR...${NC}"

    # Create directory structure
    mkdir -p "$ZERO_DIR"
    mkdir -p "$ZERO_PROJECTS_DIR"
    mkdir -p "$ZERO_DIR/cache"

    # Create config.json if it doesn't exist
    if [[ ! -f "$ZERO_DIR/config.json" ]]; then
        cat > "$ZERO_DIR/config.json" << 'EOF'
{
  "version": "1.0.0",
  "created_at": null,
  "settings": {
    "default_analyzers": [
      "technology",
      "dependencies",
      "vulnerabilities",
      "licenses"
    ],
    "full_analyzers": [
      "technology",
      "dependencies",
      "vulnerabilities",
      "package-health",
      "licenses",
      "security-findings",
      "ownership",
      "dora"
    ],
    "quick_analyzers": [
      "technology",
      "dependencies",
      "vulnerabilities",
      "licenses"
    ],
    "security_analyzers": [
      "vulnerabilities",
      "package-health",
      "security-findings",
      "provenance"
    ],
    "analyzer_timeout_seconds": 300,
    "parallel_jobs": 4
  },
  "environment": {
    "github_token_env": "GITHUB_TOKEN",
    "anthropic_key_env": "ANTHROPIC_API_KEY"
  }
}
EOF
        # Update created_at timestamp
        local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
        if command -v jq &> /dev/null; then
            local tmp=$(mktemp)
            jq --arg ts "$timestamp" '.created_at = $ts' "$ZERO_DIR/config.json" > "$tmp" && mv "$tmp" "$ZERO_DIR/config.json"
        fi
    fi

    # Create index.json if it doesn't exist
    if [[ ! -f "$ZERO_DIR/index.json" ]]; then
        cat > "$ZERO_DIR/index.json" << 'EOF'
{
  "version": "1.0.0",
  "projects": {},
  "active": null
}
EOF
    fi

    echo -e "${GREEN}✓${NC} Zero initialized at $ZERO_DIR"
    return 0
}

# Check if Gibson is initialized
zero_is_initialized() {
    [[ -d "$ZERO_DIR" ]] && [[ -f "$ZERO_DIR/config.json" ]] && [[ -f "$ZERO_DIR/index.json" ]]
}

# Ensure Gibson is initialized (auto-init if not)
zero_ensure_initialized() {
    if ! zero_is_initialized; then
        zero_init
    fi
}

#############################################################################
# Project ID Generation
#############################################################################

# Generate project ID from source URL or path
# Returns owner/repo format for nested directory structure
# Examples:
#   https://github.com/expressjs/express -> expressjs/express
#   git@github.com:lodash/lodash.git -> lodash/lodash
#   expressjs/express -> expressjs/express
#   /path/to/local/project -> local/project
zero_project_id() {
    local source="$1"
    local project_id=""
    local owner=""
    local repo=""

    # GitHub HTTPS URL: https://github.com/owner/repo or https://github.com/owner/repo.git
    if echo "$source" | grep -qE '^https://github\.com/'; then
        # Extract owner/repo from URL
        local path=$(echo "$source" | sed 's|https://github.com/||' | sed 's|\.git$||' | sed 's|/$||')
        owner=$(echo "$path" | cut -d'/' -f1)
        repo=$(echo "$path" | cut -d'/' -f2)
        project_id="${owner}/${repo}"
    # GitHub SSH URL: git@github.com:owner/repo.git
    elif echo "$source" | grep -qE '^git@github\.com:'; then
        local path=$(echo "$source" | sed 's|git@github.com:||' | sed 's|\.git$||')
        owner=$(echo "$path" | cut -d'/' -f1)
        repo=$(echo "$path" | cut -d'/' -f2)
        project_id="${owner}/${repo}"
    # Local path (starts with / or . or ~)
    elif [[ -d "$source" ]] || echo "$source" | grep -qE '^[./~]'; then
        # Use directory name, prefixed with "local/"
        local dirname=$(basename "$(cd "$source" 2>/dev/null && pwd || echo "$source")")
        project_id="local/${dirname}"
    # GitHub shorthand: owner/repo (contains exactly one /, no special prefix)
    elif echo "$source" | grep -qE '^[^/]+/[^/]+$'; then
        owner=$(echo "$source" | cut -d'/' -f1)
        repo=$(echo "$source" | cut -d'/' -f2)
        project_id="${owner}/${repo}"
    else
        # Fallback: sanitize the input
        project_id="other/$(echo "$source" | sed 's/[^a-zA-Z0-9]/-/g' | sed 's/--*/-/g' | sed 's/^-//' | sed 's/-$//')"
    fi

    # Ensure lowercase (compatible with bash 3.2+)
    echo "$project_id" | tr '[:upper:]' '[:lower:]'
}

# Get the GitHub clone URL from various input formats
zero_clone_url() {
    local source="$1"

    # Already a full URL
    if [[ "$source" =~ ^https:// ]] || [[ "$source" =~ ^git@ ]]; then
        echo "$source"
    # Shorthand owner/repo
    elif [[ "$source" =~ ^([^/]+)/([^/]+)$ ]]; then
        echo "https://github.com/$source"
    else
        echo ""
    fi
}

# Check if source is a local path
zero_is_local_source() {
    local source="$1"
    [[ -d "$source" ]] || [[ "$source" =~ ^\./ ]] || [[ "$source" =~ ^/ ]]
}

#############################################################################
# Project Management
#############################################################################

# Get project directory path
zero_project_path() {
    local project_id="$1"
    echo "$ZERO_PROJECTS_DIR/$project_id"
}

# Get project repo path
zero_project_repo_path() {
    local project_id="$1"
    echo "$ZERO_PROJECTS_DIR/$project_id/repo"
}

# Get project analysis path
zero_project_analysis_path() {
    local project_id="$1"
    echo "$ZERO_PROJECTS_DIR/$project_id/analysis"
}

# Check if project exists
zero_project_exists() {
    local project_id="$1"
    local project_path=$(zero_project_path "$project_id")
    [[ -d "$project_path" ]]
}

# Check if project is fully hydrated (has repo and completed analysis)
# Returns 0 if hydrated, 1 if not
zero_is_hydrated() {
    local project_id="$1"

    local project_path=$(zero_project_path "$project_id")
    local repo_path="$project_path/repo"
    local analysis_path="$project_path/analysis"

    # Check repo exists
    if [[ ! -d "$repo_path" ]]; then
        return 1
    fi

    # Check analysis directory exists with manifest
    if [[ ! -f "$analysis_path/manifest.json" ]]; then
        return 1
    fi

    # Check if analysis actually completed (completed_at is not null)
    local completed_at
    completed_at=$(jq -r '.completed_at // "null"' "$analysis_path/manifest.json" 2>/dev/null)
    if [[ "$completed_at" == "null" ]] || [[ -z "$completed_at" ]]; then
        return 1
    fi

    return 0
}

# Check if active project is hydrated and ready for queries
# Prints error message and returns 1 if not ready
zero_require_hydrated() {
    local project_id=$(zero_active_project)

    if [[ -z "$project_id" ]]; then
        echo -e "${RED}No active project.${NC}" >&2
        echo -e "Run ${CYAN}/zero hydrate <repo>${NC} first." >&2
        return 1
    fi

    if ! zero_is_hydrated "$project_id"; then
        echo -e "${RED}Project '$project_id' is not fully hydrated.${NC}" >&2
        echo -e "Run ${CYAN}/zero hydrate $project_id --force${NC} to complete hydration." >&2
        return 1
    fi

    return 0
}

# Get hydration status as JSON
zero_hydration_status() {
    local project_id="$1"

    if [[ -z "$project_id" ]]; then
        project_id=$(zero_active_project)
    fi

    if [[ -z "$project_id" ]]; then
        echo '{"hydrated": false, "reason": "no_active_project"}'
        return
    fi

    local project_path=$(zero_project_path "$project_id")
    local has_project=$(zero_project_exists "$project_id" && echo "true" || echo "false")
    local has_repo=$([[ -d "$project_path/repo" ]] && echo "true" || echo "false")
    local has_manifest=$([[ -f "$project_path/analysis/manifest.json" ]] && echo "true" || echo "false")
    local proj_status=$(jq -r --arg id "$project_id" '.projects[$id].status // "unknown"' "$ZERO_DIR/index.json" 2>/dev/null)

    local hydrated="false"
    local reason="unknown"

    if [[ "$has_project" != "true" ]]; then
        reason="project_not_found"
    elif [[ "$has_repo" != "true" ]]; then
        reason="repo_not_cloned"
    elif [[ "$has_manifest" != "true" ]]; then
        reason="analysis_incomplete"
    elif [[ "$proj_status" != "ready" ]]; then
        reason="status_not_ready"
    else
        hydrated="true"
        reason="ok"
    fi

    jq -n \
        --arg id "$project_id" \
        --argjson hydrated "$hydrated" \
        --arg reason "$reason" \
        --argjson has_project "$has_project" \
        --argjson has_repo "$has_repo" \
        --argjson has_manifest "$has_manifest" \
        --arg proj_status "$proj_status" \
        '{
            project_id: $id,
            hydrated: $hydrated,
            reason: $reason,
            checks: {
                project_exists: $has_project,
                repo_cloned: $has_repo,
                analysis_complete: $has_manifest,
                status: $proj_status
            }
        }'
}

# List all hydrated projects
zero_list_hydrated() {
    local projects=$(zero_list_projects)

    if [[ -z "$projects" ]]; then
        echo "[]"
        return
    fi

    local hydrated_list="[]"

    while IFS= read -r project_id; do
        [[ -z "$project_id" ]] && continue
        if zero_is_hydrated "$project_id"; then
            hydrated_list=$(echo "$hydrated_list" | jq --arg id "$project_id" '. + [$id]')
        fi
    done <<< "$projects"

    echo "$hydrated_list"
}

# List all projects by scanning directory structure
zero_list_projects() {
    if [[ ! -d "$ZERO_PROJECTS_DIR" ]]; then
        return
    fi

    # Scan for org/repo directories that have an analysis folder
    for org_dir in "$ZERO_PROJECTS_DIR"/*/; do
        [[ ! -d "$org_dir" ]] && continue
        local org=$(basename "$org_dir")

        for repo_dir in "$org_dir"*/; do
            [[ ! -d "$repo_dir" ]] && continue
            local repo=$(basename "$repo_dir")
            echo "${org}/${repo}"
        done
    done
}

# Ensure index.json exists
zero_ensure_index() {
    if [[ ! -f "$ZERO_DIR/index.json" ]]; then
        mkdir -p "$ZERO_DIR"
        cat > "$ZERO_DIR/index.json" << 'EOF'
{
  "version": "1.0.0",
  "projects": {},
  "active": null
}
EOF
    fi
}

# Get active project
zero_active_project() {
    if [[ ! -f "$ZERO_DIR/index.json" ]]; then
        echo ""
        return
    fi
    jq -r '.active // ""' "$ZERO_DIR/index.json" 2>/dev/null || echo ""
}

# Set active project
zero_set_active_project() {
    local project_id="$1"

    zero_ensure_index

    local tmp=$(mktemp)
    jq --arg id "$project_id" '.active = $id' "$ZERO_DIR/index.json" > "$tmp" && mv "$tmp" "$ZERO_DIR/index.json"
}

# Add project to index
zero_index_add_project() {
    local project_id="$1"
    local source="$2"
    local status="${3:-bootstrapping}"

    zero_ensure_index

    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    local tmp=$(mktemp)
    jq --arg id "$project_id" \
       --arg src "$source" \
       --arg ts "$timestamp" \
       --arg st "$status" \
       '.projects[$id] = {
         "source": $src,
         "created_at": $ts,
         "last_analyzed": null,
         "status": $st
       }' "$ZERO_DIR/index.json" > "$tmp" && mv "$tmp" "$ZERO_DIR/index.json"
}

# Update project status in index
zero_index_update_status() {
    local project_id="$1"
    local status="$2"

    zero_ensure_index

    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    local tmp=$(mktemp)
    jq --arg id "$project_id" \
       --arg st "$status" \
       --arg ts "$timestamp" \
       '.projects[$id].status = $st | .projects[$id].last_analyzed = $ts' \
       "$ZERO_DIR/index.json" > "$tmp" && mv "$tmp" "$ZERO_DIR/index.json"
}

# Remove project from index
zero_index_remove_project() {
    local project_id="$1"

    zero_ensure_index

    local tmp=$(mktemp)
    jq --arg id "$project_id" 'del(.projects[$id])' "$ZERO_DIR/index.json" > "$tmp" && mv "$tmp" "$ZERO_DIR/index.json"

    # Clear active if it was this project
    local active=$(zero_active_project)
    if [[ "$active" == "$project_id" ]]; then
        zero_set_active_project ""
    fi
}

#############################################################################
# Project Metadata
#############################################################################

# Create project.json for a new project
zero_create_project_metadata() {
    local project_id="$1"
    local source="$2"
    local source_type="$3"  # github, local
    local branch="$4"
    local commit="$5"

    local project_path=$(zero_project_path "$project_id")
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    cat > "$project_path/project.json" << EOF
{
  "id": "$project_id",
  "source": "$source",
  "source_type": "$source_type",
  "cloned_at": "$timestamp",
  "branch": "$branch",
  "commit": "$commit",
  "path": "$project_path/repo",
  "detected_type": {
    "languages": [],
    "frameworks": [],
    "package_managers": []
  }
}
EOF
}

# Update detected project type in project.json
zero_update_project_type() {
    local project_id="$1"
    local languages="$2"      # JSON array string
    local frameworks="$3"     # JSON array string
    local package_managers="$4"  # JSON array string

    local project_path=$(zero_project_path "$project_id")
    local project_json="$project_path/project.json"

    if [[ ! -f "$project_json" ]]; then
        return 1
    fi

    local tmp=$(mktemp)
    jq --argjson langs "$languages" \
       --argjson fwks "$frameworks" \
       --argjson pkgs "$package_managers" \
       '.detected_type.languages = $langs | .detected_type.frameworks = $fwks | .detected_type.package_managers = $pkgs' \
       "$project_json" > "$tmp" && mv "$tmp" "$project_json"
}

# Read project.json
zero_read_project_metadata() {
    local project_id="$1"
    local project_path=$(zero_project_path "$project_id")

    if [[ -f "$project_path/project.json" ]]; then
        cat "$project_path/project.json"
    else
        echo "{}"
    fi
}

#############################################################################
# Analysis Manifest
#############################################################################

# Initialize analysis manifest for a project
zero_init_analysis_manifest() {
    local project_id="$1"
    local commit="$2"
    local mode="${3:-standard}"
    local scan_id="${4:-}"
    local git_context="${5:-}"

    local analysis_path=$(zero_project_analysis_path "$project_id")
    mkdir -p "$analysis_path"

    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Generate scan_id if not provided (for backwards compatibility)
    if [[ -z "$scan_id" ]]; then
        scan_id=$(generate_scan_id)
    fi

    # Build git context JSON block
    local git_json="null"
    if [[ -n "$git_context" ]] && [[ "$git_context" != "null" ]]; then
        git_json="$git_context"
    fi

    # Create enhanced manifest (schema v2.0.0)
    jq -n \
        --arg project_id "$project_id" \
        --arg scan_id "$scan_id" \
        --arg schema_version "2.0.0" \
        --arg mode "$mode" \
        --arg started_at "$timestamp" \
        --argjson git "$git_json" \
        '{
            project_id: $project_id,
            scan_id: $scan_id,
            schema_version: $schema_version,
            git: $git,
            scan: {
                started_at: $started_at,
                completed_at: null,
                duration_seconds: null,
                profile: $mode,
                scanners_requested: [],
                scanners_completed: [],
                scanners_failed: []
            },
            analyses: {},
            summary: {
                risk_level: "unknown",
                total_dependencies: 0,
                direct_dependencies: 0,
                total_vulnerabilities: 0,
                total_security_findings: 0,
                license_status: "unknown",
                abandoned_packages: 0,
                vulnerability_count: 0,
                critical_count: 0,
                high_count: 0,
                dependency_count: 0
            }
        }' > "$analysis_path/manifest.json"
}

# Record analysis start
zero_analysis_start() {
    local project_id="$1"
    local analysis_type="$2"
    local analyzer_script="$3"
    local analyzer_version="${4:-1.0.0}"

    local analysis_path=$(zero_project_analysis_path "$project_id")
    local manifest="$analysis_path/manifest.json"
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    if [[ ! -f "$manifest" ]]; then
        return 1
    fi

    local tmp=$(mktemp)
    jq --arg type "$analysis_type" \
       --arg script "$analyzer_script" \
       --arg ver "$analyzer_version" \
       --arg ts "$timestamp" \
       '.analyses[$type] = {
         "analyzer": $script,
         "version": $ver,
         "started_at": $ts,
         "completed_at": null,
         "duration_ms": null,
         "status": "running",
         "output_file": ($type + ".json"),
         "summary": null
       }' "$manifest" > "$tmp" && mv "$tmp" "$manifest"
}

# Record analysis completion
zero_analysis_complete() {
    local project_id="$1"
    local analysis_type="$2"
    local status="$3"  # complete, failed, partial
    local duration_ms="$4"
    local summary="$5"  # JSON object string

    local analysis_path=$(zero_project_analysis_path "$project_id")
    local manifest="$analysis_path/manifest.json"
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    if [[ ! -f "$manifest" ]]; then
        return 1
    fi

    local tmp=$(mktemp)
    if [[ -n "$summary" ]] && [[ "$summary" != "null" ]]; then
        jq --arg type "$analysis_type" \
           --arg st "$status" \
           --arg ts "$timestamp" \
           --argjson dur "$duration_ms" \
           --argjson sum "$summary" \
           '.analyses[$type].completed_at = $ts | .analyses[$type].status = $st | .analyses[$type].duration_ms = $dur | .analyses[$type].summary = $sum' \
           "$manifest" > "$tmp" && mv "$tmp" "$manifest"
    else
        jq --arg type "$analysis_type" \
           --arg st "$status" \
           --arg ts "$timestamp" \
           --argjson dur "$duration_ms" \
           '.analyses[$type].completed_at = $ts | .analyses[$type].status = $st | .analyses[$type].duration_ms = $dur' \
           "$manifest" > "$tmp" && mv "$tmp" "$manifest"
    fi
}

# Finalize analysis manifest
zero_finalize_manifest() {
    local project_id="$1"

    local analysis_path=$(zero_project_analysis_path "$project_id")
    local manifest="$analysis_path/manifest.json"
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    if [[ ! -f "$manifest" ]]; then
        return 1
    fi

    # Calculate duration if started_at is available
    local started_at=$(jq -r '.scan.started_at // .started_at // empty' "$manifest" 2>/dev/null)
    local duration_seconds="null"
    if [[ -n "$started_at" ]]; then
        # Parse ISO 8601 UTC timestamp (ends with Z)
        # On macOS, set TZ=UTC to correctly interpret the Z suffix
        local start_epoch
        if [[ "$started_at" == *Z ]]; then
            # macOS date doesn't handle Z suffix correctly, use UTC timezone
            start_epoch=$(TZ=UTC date -j -f "%Y-%m-%dT%H:%M:%SZ" "$started_at" +%s 2>/dev/null)
        fi
        # Fallback for GNU date on Linux
        [[ -z "$start_epoch" ]] && start_epoch=$(date -d "$started_at" +%s 2>/dev/null)

        local end_epoch=$(date +%s)
        if [[ -n "$start_epoch" ]]; then
            duration_seconds=$((end_epoch - start_epoch))
            # Ensure non-negative
            [[ $duration_seconds -lt 0 ]] && duration_seconds=0
        fi
    fi

    # Get completed and failed scanners from analyses
    local scanners_completed=$(jq -r '[.analyses // {} | to_entries[] | select(.value.status == "complete") | .key] | unique' "$manifest" 2>/dev/null)
    local scanners_failed=$(jq -r '[.analyses // {} | to_entries[] | select(.value.status == "failed") | .key] | unique' "$manifest" 2>/dev/null)

    local tmp=$(mktemp)
    jq --arg ts "$timestamp" \
       --argjson duration "$duration_seconds" \
       --argjson completed "$scanners_completed" \
       --argjson failed "$scanners_failed" \
       '
       # Update both old and new schema fields for compatibility
       .completed_at = $ts |
       .scan.completed_at = $ts |
       .scan.duration_seconds = $duration |
       .scan.scanners_completed = $completed |
       .scan.scanners_failed = $failed
       ' "$manifest" > "$tmp" && mv "$tmp" "$manifest"
}

# Update manifest summary
zero_update_summary() {
    local project_id="$1"
    local risk_level="$2"
    local total_deps="$3"
    local direct_deps="$4"
    local total_vulns="$5"
    local total_findings="$6"
    local license_status="$7"
    local abandoned="$8"

    local analysis_path=$(zero_project_analysis_path "$project_id")
    local manifest="$analysis_path/manifest.json"

    if [[ ! -f "$manifest" ]]; then
        return 1
    fi

    local tmp=$(mktemp)
    jq --arg risk "$risk_level" \
       --argjson tdeps "$total_deps" \
       --argjson ddeps "$direct_deps" \
       --argjson vulns "$total_vulns" \
       --argjson finds "$total_findings" \
       --arg lic "$license_status" \
       --argjson aband "$abandoned" \
       '.summary = {
         "risk_level": $risk,
         "total_dependencies": $tdeps,
         "direct_dependencies": $ddeps,
         "total_vulnerabilities": $vulns,
         "total_security_findings": $finds,
         "license_status": $lic,
         "abandoned_packages": $aband
       }' "$manifest" > "$tmp" && mv "$tmp" "$manifest"
}

#############################################################################
# Utility Functions
#############################################################################

# Print Gibson status header
zero_print_header() {
    print_zero_banner
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo
}

# Print status line with checkmark or X
zero_print_status() {
    local message="$1"
    local status="$2"  # ok, fail, warn, running
    local detail="${3:-}"

    local status_icon=""
    local status_color=""

    case "$status" in
        ok|complete)
            status_icon="✓"
            status_color="$GREEN"
            ;;
        fail|failed)
            status_icon="✗"
            status_color="$RED"
            ;;
        warn|partial)
            status_icon="⚠"
            status_color="$YELLOW"
            ;;
        running)
            status_icon="○"
            status_color="$CYAN"
            ;;
        *)
            status_icon="·"
            status_color="$NC"
            ;;
    esac

    printf "  %-50s ${status_color}%s${NC}" "$message" "$status_icon"
    if [[ -n "$detail" ]]; then
        echo -e "  ${CYAN}$detail${NC}"
    else
        echo
    fi
}

# Get human-readable time ago
zero_time_ago() {
    local timestamp="$1"
    local now=$(date +%s)
    local then

    # Handle UTC timestamps (ending with Z) - macOS needs TZ=UTC
    if [[ "$timestamp" == *Z ]]; then
        then=$(TZ=UTC date -j -f "%Y-%m-%dT%H:%M:%SZ" "$timestamp" +%s 2>/dev/null)
    fi
    # Fallback for GNU date on Linux
    [[ -z "$then" ]] && then=$(date -d "$timestamp" +%s 2>/dev/null)

    if [[ -z "$then" ]]; then
        echo "unknown"
        return
    fi

    local diff=$((now - then))

    if [[ $diff -lt 60 ]]; then
        echo "just now"
    elif [[ $diff -lt 3600 ]]; then
        echo "$((diff / 60)) minutes ago"
    elif [[ $diff -lt 86400 ]]; then
        echo "$((diff / 3600)) hours ago"
    elif [[ $diff -lt 604800 ]]; then
        echo "$((diff / 86400)) days ago"
    else
        echo "$((diff / 604800)) weeks ago"
    fi
}

# Calculate disk usage for a project
zero_project_size() {
    local project_id="$1"
    local project_path=$(zero_project_path "$project_id")

    if [[ -d "$project_path" ]]; then
        du -sh "$project_path" 2>/dev/null | cut -f1
    else
        echo "0"
    fi
}

# Calculate total Gibson disk usage
zero_total_size() {
    if [[ -d "$ZERO_DIR" ]]; then
        du -sh "$ZERO_DIR" 2>/dev/null | cut -f1
    else
        echo "0"
    fi
}

#############################################################################
# Dynamic Status Display for Parallel Scanners
#############################################################################

# Spinner characters for running scanners
SPINNER_CHARS=('⠋' '⠙' '⠹' '⠸' '⠼' '⠴' '⠦' '⠧' '⠇' '⠏')

# Get spinner character based on iteration
get_spinner_char() {
    local iteration=$1
    local index=$((iteration % ${#SPINNER_CHARS[@]}))
    echo "${SPINNER_CHARS[$index]}"
}

# Initialize scanner status display
# Usage: init_scanner_status_display <scanner_list>
# Creates a status file for tracking scanner states
init_scanner_status_display() {
    local scanners="$1"
    local status_dir=$(mktemp -d)

    # Create status file for each scanner
    for scanner in $scanners; do
        echo "queued" > "$status_dir/$scanner.status"
    done

    echo "$status_dir"
}

# Update scanner status
# Usage: update_scanner_status <status_dir> <scanner> <status> [result] [duration]
update_scanner_status() {
    local status_dir="$1"
    local scanner="$2"
    local status="$3"
    local result="${4:-}"
    local duration="${5:-0}"

    echo "$status|$result|$duration" > "$status_dir/$scanner.status"
}

# Render single scanner status line
# Usage: render_scanner_line <scanner> <display_name> <status> <result> <duration> <spinner_iteration>
render_scanner_line() {
    local scanner="$1"
    local display_name="$2"
    local status="$3"
    local result="$4"
    local duration="$5"
    local spinner_iter="${6:-0}"

    local icon=""
    local color=""
    local status_text=""

    case "$status" in
        queued)
            icon="${WHITE}○${NC}"
            status_text="${DIM}queued...${NC}"
            ;;
        running)
            local spinner=$(get_spinner_char "$spinner_iter")
            icon="${CYAN}${spinner}${NC}"
            status_text="${DIM}running...${NC}"
            ;;
        complete)
            icon="${GREEN}✓${NC}"
            status_text="${GREEN}${result}${NC}"
            if [[ "$duration" != "0" ]]; then
                status_text="$status_text ${DIM}${duration}s${NC}"
            fi
            ;;
        failed)
            icon="${RED}✗${NC}"
            status_text="${RED}failed${NC}"
            if [[ "$duration" != "0" ]]; then
                status_text="$status_text ${DIM}${duration}s${NC}"
            fi
            ;;
        *)
            icon="${DIM}○${NC}"
            status_text="${DIM}unknown${NC}"
            ;;
    esac

    printf "  %b %-24s %b\n" "$icon" "$display_name" "$status_text"
}

# Render all scanner status lines
# Usage: render_all_scanner_lines <status_dir> <scanner_list> <display_names> <spinner_iteration>
render_all_scanner_lines() {
    local status_dir="$1"
    local scanners="$2"
    local spinner_iter="${3:-0}"

    for scanner in $scanners; do
        local status_file="$status_dir/$scanner.status"
        if [[ -f "$status_file" ]]; then
            IFS='|' read -r status result duration < "$status_file"
        else
            status="queued"
            result=""
            duration="0"
        fi

        # Get display name dynamically
        local display_name=$(get_scanner_display_name "$scanner" 2>/dev/null || echo "$scanner")

        render_scanner_line "$scanner" "$display_name" "$status" "$result" "$duration" "$spinner_iter"
    done
}

# Move cursor up N lines
move_cursor_up() {
    local lines=$1
    printf "\033[%dA" "$lines"
}

# Clear current line
clear_line() {
    printf "\r\033[K"
}

# Save cursor position
save_cursor() {
    printf "\033[s"
}

# Restore cursor position
restore_cursor() {
    printf "\033[u"
}

# Check if any scanners are still running
# Usage: scanners_still_running <status_dir> <scanner_list>
scanners_still_running() {
    local status_dir="$1"
    local scanners="$2"

    for scanner in $scanners; do
        local status_file="$status_dir/$scanner.status"
        if [[ -f "$status_file" ]]; then
            local status=$(cut -d'|' -f1 "$status_file")
            if [[ "$status" == "running" ]] || [[ "$status" == "queued" ]]; then
                return 0
            fi
        fi
    done

    return 1
}

#############################################################################
# Org Scan Dashboard Display
#############################################################################

# Sanitize repo name for use as filename (replace / with --)
# Usage: sanitize_repo_name <repo>
sanitize_repo_name() {
    echo "$1" | sed 's/\//__/g'
}

# Initialize org scan dashboard
# Usage: init_org_scan_dashboard <repo_list> <status_dir>
init_org_scan_dashboard() {
    local repos="$1"
    local status_dir="$2"

    # Create status file for each repo
    for repo in $repos; do
        local safe_name=$(sanitize_repo_name "$repo")
        # Format: status|current_scanner|progress|duration
        echo "queued|||0" > "$status_dir/$safe_name.status"
    done
}

# Update repo scan status
# Usage: update_repo_scan_status <status_dir> <repo> <status> <current_scanner> <progress> <duration> [clone_status]
update_repo_scan_status() {
    local status_dir="$1"
    local repo="$2"
    local status="$3"
    local current_scanner="${4:-}"
    local progress="${5:-}"
    local duration="${6:-0}"
    local clone_status="${7:-}"

    local safe_name=$(sanitize_repo_name "$repo")
    local status_file="$status_dir/$safe_name.status"

    # Preserve existing clone_status if not provided
    if [[ -z "$clone_status" ]] && [[ -f "$status_file" ]]; then
        local existing_clone_status
        existing_clone_status=$(cut -d'|' -f5 "$status_file" 2>/dev/null)
        [[ -n "$existing_clone_status" ]] && clone_status="$existing_clone_status"
    fi

    echo "$status|$current_scanner|$progress|$duration|$clone_status" > "$status_file"
}

# Render single repo status line for dashboard
# Usage: render_repo_dashboard_line <repo> <status> <current_scanner> <progress> <duration> <spinner_iter>
render_repo_dashboard_line() {
    local repo="$1"
    local status="$2"
    local current_scanner="$3"
    local progress="$4"
    local duration="$5"
    local spinner_iter="${6:-0}"

    # Truncate repo name if too long
    local repo_display="$repo"
    if [[ ${#repo_display} -gt 30 ]]; then
        repo_display="${repo_display:0:27}..."
    fi

    local icon=""
    local status_text=""

    case "$status" in
        queued)
            icon="${WHITE}○${NC}"
            status_text="${DIM}queued...${NC}"
            ;;
        running)
            local spinner=$(get_spinner_char "$spinner_iter")
            icon="${CYAN}${spinner}${NC}"
            if [[ -n "$current_scanner" ]]; then
                status_text="${CYAN}$current_scanner${NC} ${DIM}$progress${NC}"
            else
                status_text="${DIM}running...${NC}"
            fi
            ;;
        complete)
            icon="${GREEN}✓${NC}"
            status_text="${GREEN}complete${NC} ${DIM}${duration}s${NC}"
            if [[ -n "$progress" ]]; then
                status_text="$status_text ${DIM}• $progress${NC}"
            fi
            ;;
        failed)
            icon="${RED}✗${NC}"
            status_text="${RED}failed${NC}"
            ;;
        *)
            icon="${DIM}○${NC}"
            status_text="${DIM}unknown${NC}"
            ;;
    esac

    printf "  %b %-30s %b\n" "$icon" "$repo_display" "$status_text"
}

# Render full org scan dashboard
# Usage: render_org_scan_dashboard <status_dir> <repo_list> <spinner_iteration>
render_org_scan_dashboard() {
    local status_dir="$1"
    local repos="$2"
    local spinner_iter="${3:-0}"

    for repo in $repos; do
        local safe_name=$(sanitize_repo_name "$repo")
        local status_file="$status_dir/$safe_name.status"
        if [[ -f "$status_file" ]]; then
            IFS='|' read -r status current_scanner progress duration < "$status_file"
        else
            status="queued"
            current_scanner=""
            progress=""
            duration="0"
        fi

        render_repo_dashboard_line "$repo" "$status" "$current_scanner" "$progress" "$duration" "$spinner_iter"
    done
}

# Check if any repos are still scanning
# Usage: repos_still_scanning <status_dir> <repo_list>
repos_still_scanning() {
    local status_dir="$1"
    local repos="$2"

    for repo in $repos; do
        local safe_name=$(sanitize_repo_name "$repo")
        local status_file="$status_dir/$safe_name.status"
        if [[ -f "$status_file" ]]; then
            local status=$(cut -d'|' -f1 "$status_file")
            if [[ "$status" == "running" ]] || [[ "$status" == "queued" ]]; then
                return 0
            fi
        fi
    done

    return 1
}

#############################################################################
# Progress Bar Display
#############################################################################

# Render a progress bar
# Usage: render_progress_bar <current> <total> <width> [label]
render_progress_bar() {
    local current=$1
    local total=$2
    local width=${3:-50}
    local label="${4:-}"

    local pct=0
    if [[ $total -gt 0 ]]; then
        pct=$((current * 100 / total))
    fi

    local filled=$((current * width / total))
    [[ $filled -gt $width ]] && filled=$width
    local empty=$((width - filled))

    # Build progress bar
    local bar=""
    for ((i=0; i<filled; i++)); do bar+="="; done
    if [[ $filled -lt $width ]]; then
        bar+=">"
        ((empty--))
    fi
    for ((i=0; i<empty; i++)); do bar+=" "; done

    # Print progress bar
    if [[ -n "$label" ]]; then
        printf "\r%s [%s] %d%% (%d/%d)" "$label" "$bar" "$pct" "$current" "$total"
    else
        printf "\r[%s] %d%% (%d/%d)" "$bar" "$pct" "$current" "$total"
    fi
}

# Clear current line and move to beginning
clear_progress_line() {
    printf "\r\033[K"
}

# Format duration in seconds to human-readable format (e.g., "3m 24s")
format_duration() {
    local seconds=$1
    local minutes=$((seconds / 60))
    local secs=$((seconds % 60))

    if [[ $minutes -gt 0 ]]; then
        echo "${minutes}m ${secs}s"
    else
        echo "${secs}s"
    fi
}

# Format bytes to human readable (e.g., "245mb", "1.1gb")
format_size_lower() {
    local bytes=$1
    if [[ $bytes -ge 1073741824 ]]; then
        local gb=$((bytes / 1073741824))
        local remainder_bytes=$((bytes % 1073741824))
        # Divide first to avoid overflow, then get single decimal digit
        local decimal=$(( remainder_bytes / 107374182 ))  # 1073741824/10
        if [[ $decimal -gt 0 ]]; then
            echo "${gb}.${decimal}gb"
        else
            echo "${gb}gb"
        fi
    elif [[ $bytes -ge 1048576 ]]; then
        echo "$(( bytes / 1048576 ))mb"
    elif [[ $bytes -ge 1024 ]]; then
        echo "$(( bytes / 1024 ))kb"
    else
        echo "${bytes}b"
    fi
}

# Get repo stats (size in bytes and file count) for a cloned repo
# Usage: get_repo_stats <repo_path>
# Returns: "size_bytes|file_count" or "0|0" if not found
get_repo_stats() {
    local repo_path="$1"

    if [[ ! -d "$repo_path" ]]; then
        echo "0|0"
        return
    fi

    # Get size in bytes (excluding .git for accuracy)
    local size_bytes=$(du -sk "$repo_path" 2>/dev/null | cut -f1)
    size_bytes=$((size_bytes * 1024))  # Convert KB to bytes

    # Get file count (excluding .git directory)
    local file_count=$(find "$repo_path" -type f ! -path '*/.git/*' 2>/dev/null | wc -l | tr -d ' ')

    echo "${size_bytes}|${file_count}"
}

# Format repo stats for display
# Usage: format_repo_stats <repo_path>
# Returns: "245mb, 900 files" or empty if not found
format_repo_stats() {
    local repo_path="$1"
    local stats=$(get_repo_stats "$repo_path")

    local size_bytes="${stats%%|*}"
    local file_count="${stats##*|}"

    if [[ "$size_bytes" == "0" ]] && [[ "$file_count" == "0" ]]; then
        echo ""
        return
    fi

    local size_str=$(format_size_lower "$size_bytes")
    printf "%s, %s files" "$size_str" "$file_count"
}

# Calculate progress including partial repo completion
calculate_partial_progress() {
    local status_dir="$1"
    local completed_count="$2"
    shift 2
    local all_repos=("$@")

    local total_progress=0

    for repo in "${all_repos[@]}"; do
        local safe_name=$(sanitize_repo_name "$repo")
        local status_file="$status_dir/$safe_name.status"

        if [[ -f "$status_file" ]]; then
            IFS='|' read -r status current_scanner progress duration < "$status_file"

            if [[ "$progress" =~ ^([0-9]+)/([0-9]+)$ ]]; then
                local current="${BASH_REMATCH[1]}"
                local total="${BASH_REMATCH[2]}"
                if [[ $total -gt 0 ]]; then
                    total_progress=$((total_progress + (current * 100 / total)))
                fi
            fi
        fi
    done

    local partial=$((total_progress / 100))
    echo $((completed_count + partial))
}

# Track if first render for clearing
SCAN_STATUS_FIRST_RENDER=1
SCAN_STATUS_PREV_LINES=0

# Render multi-line scan status with one line per running repo
# Line 1: spinner + overall progress (repos)
# Lines 2+: Each running repo with scanner, duration, completed/queued scanners
render_scan_status_line() {
    local status_dir="$1"
    local repo_count="$2"
    local completed_count="$3"
    local elapsed="$4"
    shift 4
    local all_repos=("$@")

    # Count repos by status and collect running repo info
    local running_count=0
    local running_repos=()

    for repo in "${all_repos[@]}"; do
        local safe_name=$(sanitize_repo_name "$repo")
        local status_file="$status_dir/$safe_name.status"

        if [[ -f "$status_file" ]]; then
            IFS='|' read -r status current_scanner progress duration < "$status_file"
            if [[ "$status" == "running" ]]; then
                ((running_count++))
                # Store: repo|scanner|progress|duration
                running_repos+=("$repo|$current_scanner|$progress|$duration")
            fi
        fi
    done

    # Calculate queued count: total - running - completed
    local queued_count=$((repo_count - running_count - completed_count))

    # Get spinner frame
    local spinner=$(get_spinner_frame)

    # Calculate partial progress
    local display_progress=$(calculate_partial_progress "$status_dir" "$completed_count" "${all_repos[@]}")
    if [[ $repo_count -eq 0 ]]; then
        local pct=0
    else
        local pct=$((display_progress * 100 / repo_count))
    fi

    # Build progress bar (compact version)
    local bar_width=30
    if [[ $repo_count -eq 0 ]]; then
        local filled=0
    else
        local filled=$((display_progress * bar_width / repo_count))
    fi
    [[ $filled -gt $bar_width ]] && filled=$bar_width
    local empty=$((bar_width - filled))
    local bar=""
    for ((i=0; i<filled; i++)); do bar+="="; done
    if [[ $filled -lt $bar_width ]] && [[ $filled -gt 0 ]]; then
        bar+=">"
        ((empty--))
    fi
    for ((i=0; i<empty; i++)); do bar+=" "; done

    # Calculate number of lines we'll draw
    local new_lines=$((1 + ${#running_repos[@]}))  # 1 header + N repos

    # Clear previous lines if not first render
    if [[ $SCAN_STATUS_FIRST_RENDER -eq 0 ]] && [[ $SCAN_STATUS_PREV_LINES -gt 0 ]]; then
        # Move up and clear all previous lines
        for ((i=0; i<$SCAN_STATUS_PREV_LINES; i++)); do
            printf "\r\033[2K\033[1A"
        done
        printf "\r\033[2K"
    elif [[ $SCAN_STATUS_FIRST_RENDER -eq 1 ]]; then
        printf "\r\033[2K"
        SCAN_STATUS_FIRST_RENDER=0
    fi

    # Store line count for next render
    SCAN_STATUS_PREV_LINES=$new_lines

    # Line 1: Status summary (repo counts)
    printf "%s Scanning: %d running • %d complete • %d queued (%ds) [%s] %d%%\n" \
        "$spinner" "$running_count" "$completed_count" "$queued_count" "$elapsed" \
        "$bar" "$pct"

    # Lines 2+: One line per running repo
    for repo_info in "${running_repos[@]}"; do
        IFS='|' read -r repo scanner progress duration <<< "$repo_info"
        local short_repo="${repo##*/}"

        # Parse progress to get completed and queued scanners
        local scanners_complete=0
        local scanners_queued=0
        if [[ "$progress" =~ ^([0-9]+)/([0-9]+)$ ]]; then
            local current="${BASH_REMATCH[1]}"
            local total="${BASH_REMATCH[2]}"
            # Current scanner is running, so complete = current - 1
            scanners_complete=$((current - 1))
            # Queued = total - current
            scanners_queued=$((total - current))
        fi

        # Format duration
        local duration_str="${duration}s"

        # Print repo line with colored active scanner
        printf "  └─ %s (${CYAN}%s: %s${NC}) ${GREEN}%d complete${NC}, %d queued\n" \
            "$short_repo" "$scanner" "$duration_str" "$scanners_complete" "$scanners_queued"
    done
}

#############################################################################
# Todo-Style Progress Display
# Clean, Claude Code-inspired todo list format for org scans
#############################################################################

# Track line count for clearing (Bash 3 compatible - no associative arrays)
TODO_DISPLAY_LINES=0

# Initialize todo display with header
# Usage: init_todo_display [org_name] [repo_count] [profile] [parallel_jobs]
init_todo_display() {
    TODO_DISPLAY_LINES=0

    local org_name="${1:-}"
    local repo_count="${2:-}"
    local profile="${3:-}"
    local parallel_jobs="${4:-}"

    # Print header if org info provided
    if [[ -n "$org_name" ]]; then
        printf "\n"
        printf "${BOLD}Scan Organization${NC}\n"
        printf "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
        printf "\n"
        printf "Organization: ${CYAN}%s${NC}\n" "$org_name"
        [[ -n "$repo_count" ]] && printf "Repositories: ${CYAN}%s${NC}\n" "$repo_count"
        [[ -n "$profile" ]] && printf "Profile:      ${CYAN}%s${NC}\n" "$profile"
        [[ -n "$parallel_jobs" ]] && printf "Parallel:     ${CYAN}%s jobs${NC}\n" "$parallel_jobs"
        printf "\n"
    fi
}

# Render todo-style progress display with separate clone and scan sections
# Usage: render_todo_display status_dir org_name elapsed_seconds repo1 repo2 ...
# Status file format: status|scanner|progress|duration|clone_status
# clone_status: cloned|skipped|pending (optional field)
render_todo_display() {
    local status_dir="$1"
    local org_name="$2"
    local elapsed="$3"
    shift 3
    local all_repos=("$@")

    local total=${#all_repos[@]}

    # Arrays for clone status
    local clone_complete=()
    local clone_running=()
    local clone_waiting=()

    # Arrays for scan status
    local scan_running=()
    local scan_complete=()
    local scan_waiting=()

    local complete_count=0

    for current_repo in "${all_repos[@]}"; do
        local safe_name=$(sanitize_repo_name "$current_repo")
        local short="${current_repo##*/}"
        local status_file="$status_dir/$safe_name.status"

        if [[ -f "$status_file" ]]; then
            IFS='|' read -r status scanner progress duration clone_status < "$status_file"

            # Clone status - include full repo name for stats lookup
            if [[ "$clone_status" == "cloned" ]]; then
                clone_complete+=("$short|Cloned|$current_repo")
            elif [[ "$clone_status" == "skipped" ]]; then
                clone_complete+=("$short|Already cloned|$current_repo")
            elif [[ "$scanner" == "cloning" ]]; then
                clone_running+=("$short")
            fi

            # Scan status
            if [[ "$status" == "complete" ]]; then
                ((complete_count++))
                scan_complete+=("$short|$duration")
            elif [[ "$status" == "running" ]] && [[ "$scanner" != "cloning" ]]; then
                scan_running+=("$short|$scanner|$progress")
            elif [[ "$status" == "pending" ]] || [[ -z "$status" ]]; then
                scan_waiting+=("$short")
            fi
        else
            # No status file yet - check if repo is already cloned on disk
            local project_id=$(zero_project_id "$current_repo")
            local repo_path="$ZERO_PROJECTS_DIR/$project_id/repo"
            if [[ -d "$repo_path" ]]; then
                clone_complete+=("$short|Already cloned|$current_repo")
            else
                clone_waiting+=("$short")
            fi
            scan_waiting+=("$short")
        fi
    done

    # Clear previous lines
    if [[ $TODO_DISPLAY_LINES -gt 0 ]]; then
        for ((i=0; i<$TODO_DISPLAY_LINES; i++)); do
            printf "\033[1A\033[2K"
        done
    fi

    local spinner=$(get_spinner_frame)
    local line_count=0

    # Clone section - show all repos with status and stats
    local clone_total=${#clone_complete[@]}
    local clone_running_count=${#clone_running[@]}
    local clone_waiting_count=${#clone_waiting[@]}

    # If all cloning is done (no running, no waiting), show full list with summary
    if [[ $clone_running_count -eq 0 ]] && [[ $clone_waiting_count -eq 0 ]] && [[ $clone_total -gt 0 ]]; then
        printf "\n"; ((line_count++))

        # Show all cloned repos (no stats in live display - too slow)
        for info in "${clone_complete[@]}"; do
            IFS='|' read -r short clone_msg repo_full <<< "$info"

            if [[ "$clone_msg" == "Cloned" ]]; then
                printf "    ${GREEN}✓${NC}  %-24s ${GREEN}Complete${NC}\n" "$short"
            else
                printf "    ${GREEN}✓${NC}  %-24s ${DIM}Previously complete${NC}\n" "$short"
            fi
            ((line_count++))
        done

        printf "\n"; ((line_count++))

        # Count cloned vs skipped
        local freshly_cloned=0
        local already_cloned=0
        for info in "${clone_complete[@]}"; do
            IFS='|' read -r short clone_msg repo_full <<< "$info"
            if [[ "$clone_msg" == "Cloned" ]]; then
                ((freshly_cloned++))
            else
                ((already_cloned++))
            fi
        done

        printf "  ${GREEN}✓${NC} ${BOLD}Cloning complete${NC}"
        local clone_parts=()
        [[ $freshly_cloned -gt 0 ]] && clone_parts+=("${freshly_cloned} cloned")
        [[ $already_cloned -gt 0 ]] && clone_parts+=("${already_cloned} already cloned")
        if [[ ${#clone_parts[@]} -gt 0 ]]; then
            printf " ("
            local first=true
            for part in "${clone_parts[@]}"; do
                [[ "$first" != "true" ]] && printf ", "
                printf "%s" "$part"
                first=false
            done
            printf ")"
        fi
        printf "\n"; ((line_count++))
        printf "\n"; ((line_count++))
    else
        printf "\n"; ((line_count++))
        printf "  ${BOLD}%s Cloning repos${NC}\n" "$spinner"; ((line_count++))
        printf "\n"; ((line_count++))

        # Show ALL repos in order: running, complete, waiting
        # Show clone running
        for short in "${clone_running[@]}"; do
            printf "    ${YELLOW}*${NC}  %-24s ${CYAN}Cloning...${NC}\n" "$short"; ((line_count++))
        done

        # Show clone complete (no stats in live display - too slow)
        for info in "${clone_complete[@]}"; do
            IFS='|' read -r short clone_msg repo_full <<< "$info"

            if [[ "$clone_msg" == "Cloned" ]]; then
                printf "    ${GREEN}✓${NC}  %-24s ${GREEN}Complete${NC}\n" "$short"
            else
                printf "    ${GREEN}✓${NC}  %-24s ${DIM}Previously complete${NC}\n" "$short"
            fi
            ((line_count++))
        done

        # Show clone waiting
        for short in "${clone_waiting[@]}"; do
            printf "    ${DIM}○${NC}  ${DIM}%-24s Waiting${NC}\n" "$short"; ((line_count++))
        done
    fi

    # Scan section - simple stable list, one line per repo
    printf "\n"; ((line_count++))

    for current_repo in "${all_repos[@]}"; do
        local short="${current_repo##*/}"
        local found_status="waiting"
        local found_info=""

        # Check if complete
        for repo_info in "${scan_complete[@]}"; do
            IFS='|' read -r check_short duration <<< "$repo_info"
            if [[ "$check_short" == "$short" ]]; then
                found_status="complete"
                found_info="$duration"
                break
            fi
        done

        # Check if running
        if [[ "$found_status" == "waiting" ]]; then
            for repo_info in "${scan_running[@]}"; do
                IFS='|' read -r check_short scanner progress <<< "$repo_info"
                if [[ "$check_short" == "$short" ]]; then
                    found_status="running"
                    found_info="$scanner"
                    break
                fi
            done
        fi

        case "$found_status" in
            complete)
                printf "  ${GREEN}✓${NC} ${DIM}%-20s${NC} ${DIM}done (%ss)${NC}\n" "$short" "$found_info"
                ;;
            running)
                printf "  ${YELLOW}○${NC} %-20s ${CYAN}%s${NC}\n" "$short" "$found_info"
                ;;
            waiting)
                printf "  ${DIM}○ %-20s pending${NC}\n" "$short"
                ;;
        esac
        ((line_count++))
    done

    TODO_DISPLAY_LINES=$line_count
}

# Finalize todo display - show final summary with repo stats
# Usage: finalize_todo_display org_name total success failed duration repos_array
finalize_todo_display() {
    local org_name="$1"
    local total="$2"
    local success="$3"
    local failed="$4"
    local duration="$5"
    shift 5
    local repos=("$@")

    # Clear the last progress display
    if [[ $TODO_DISPLAY_LINES -gt 0 ]]; then
        for ((i=0; i<$TODO_DISPLAY_LINES; i++)); do
            printf "\033[1A\033[2K"
        done
    fi
    TODO_DISPLAY_LINES=0

    printf "\n"

    # Show all repos with stats (only calculated once at the end)
    if [[ ${#repos[@]} -gt 0 ]]; then
        for repo in "${repos[@]}"; do
            local short="${repo##*/}"
            local project_id=$(zero_project_id "$repo")
            local repo_path="$ZERO_PROJECTS_DIR/$project_id/repo"
            local stats=$(format_repo_stats "$repo_path")

            printf "    ${GREEN}✓${NC}  %-24s" "$short"
            [[ -n "$stats" ]] && printf " ${DIM}(%s)${NC}" "$stats"
            printf "\n"
        done
        printf "\n"
    fi

    printf "${GREEN}${BOLD}✓ Hydrate complete${NC}\n"
    printf "\n"
    printf "  Organization: ${CYAN}%s${NC}\n" "$org_name"
    printf "  Repositories: %d processed in %s\n" "$total" "$duration"
    printf "  Results: ${GREEN}%d success${NC}" "$success"
    [[ $failed -gt 0 ]] && printf " • ${RED}%d failed${NC}" "$failed"
    printf "\n"
    printf "\n"
    printf "  ${DIM}View results: ./zero.sh report --org %s${NC}\n" "$org_name"
    printf "\n"
}

#############################################################################
# Todo-Style Clone Display
# Consistent display format for org cloning
#############################################################################

# Track line count for clone display
CLONE_DISPLAY_LINES=0

# Initialize clone display
init_clone_display() {
    CLONE_DISPLAY_LINES=0
}

# Render todo-style clone progress display showing ALL repos
# Usage: render_clone_display org_name elapsed all_repos_array
# all_repos_array format: "repo1|status1|stats1" where status is pending|cloning|complete|skipped|failed
# stats format: "245mb, 900 files" (empty for pending/failed)
render_clone_display() {
    local org_name="$1"
    local elapsed="$2"
    shift 2
    local all_repos=("$@")

    local total=${#all_repos[@]}

    # Count statuses
    local completed=0
    local cloned=0
    local skipped=0
    local failed=0
    local cloning=0

    for repo_info in "${all_repos[@]}"; do
        IFS='|' read -r repo status stats <<< "$repo_info"
        case "$status" in
            complete) ((completed++)); ((cloned++)) ;;
            skipped) ((completed++)); ((skipped++)) ;;
            failed) ((completed++)); ((failed++)) ;;
            cloning) ((cloning++)) ;;
        esac
    done

    # Calculate percentage
    local pct=0
    [[ $total -gt 0 ]] && pct=$((completed * 100 / total))

    # Get spinner
    local spinner=$(get_spinner_frame)

    # Clear previous lines
    if [[ $CLONE_DISPLAY_LINES -gt 0 ]]; then
        for ((i=0; i<$CLONE_DISPLAY_LINES; i++)); do
            printf "\033[1A\033[2K"
        done
    fi

    # Count lines as we print them
    local line_count=0

    # Header (3 lines: blank, header text, blank)
    printf "\n"; ((line_count++))
    printf "${BOLD}%s Cloning %s${NC} (%d/%d repos, %d%%, %ds)\n" \
        "$spinner" "$org_name" "$completed" "$total" "$pct" "$elapsed"; ((line_count++))
    printf "\n"; ((line_count++))

    # Show ALL repos with their status
    for repo_info in "${all_repos[@]}"; do
        IFS='|' read -r repo status stats <<< "$repo_info"
        local short="${repo##*/}"

        case "$status" in
            complete)
                printf "    ${GREEN}✓${NC}  %-24s ${GREEN}Complete${NC}" "$short"
                [[ -n "$stats" ]] && printf " ${DIM}(%s)${NC}" "$stats"
                printf "\n"
                ;;
            skipped)
                printf "    ${GREEN}✓${NC}  %-24s ${DIM}Previously complete${NC}" "$short"
                [[ -n "$stats" ]] && printf " ${DIM}(%s)${NC}" "$stats"
                printf "\n"
                ;;
            failed)
                printf "    ${RED}✗${NC}  %-24s ${RED}Failed${NC}\n" "$short"
                ;;
            cloning)
                printf "    ${YELLOW}*${NC}  %-24s ${CYAN}Cloning...${NC}\n" "$short"
                ;;
            pending)
                printf "    ${DIM}○${NC}  ${DIM}%-24s Waiting${NC}" "$short"
                [[ -n "$stats" ]] && printf " ${DIM}(%s)${NC}" "$stats"
                printf "\n"
                ;;
        esac
        ((line_count++))
    done

    CLONE_DISPLAY_LINES=$line_count
}

# Finalize clone display - show full list then summary
# Usage: finalize_clone_display org_name duration all_repos_array
finalize_clone_display() {
    local org_name="$1"
    local duration="$2"
    shift 2
    local all_repos=("$@")

    local total=${#all_repos[@]}

    # Count statuses
    local cloned=0
    local skipped=0
    local failed=0

    for repo_info in "${all_repos[@]}"; do
        IFS='|' read -r repo status stats <<< "$repo_info"
        case "$status" in
            complete) ((cloned++)) ;;
            skipped) ((skipped++)) ;;
            failed) ((failed++)) ;;
        esac
    done

    # Clear the last progress display
    if [[ $CLONE_DISPLAY_LINES -gt 0 ]]; then
        for ((i=0; i<$CLONE_DISPLAY_LINES; i++)); do
            printf "\033[1A\033[2K"
        done
    fi
    CLONE_DISPLAY_LINES=0

    printf "\n"

    # Show ALL repos with their final status
    for repo_info in "${all_repos[@]}"; do
        IFS='|' read -r repo status stats <<< "$repo_info"
        local short="${repo##*/}"

        case "$status" in
            complete)
                printf "    ${GREEN}✓${NC}  %-24s ${GREEN}Complete${NC}" "$short"
                [[ -n "$stats" ]] && printf " ${DIM}(%s)${NC}" "$stats"
                printf "\n"
                ;;
            skipped)
                printf "    ${GREEN}✓${NC}  %-24s ${DIM}Previously complete${NC}" "$short"
                [[ -n "$stats" ]] && printf " ${DIM}(%s)${NC}" "$stats"
                printf "\n"
                ;;
            failed)
                printf "    ${RED}✗${NC}  %-24s ${RED}Failed${NC}\n" "$short"
                ;;
        esac
    done

    printf "\n"
    printf "${GREEN}${BOLD}✓ Cloning complete${NC}"

    # Build summary parts
    local summary_parts=()
    [[ $cloned -gt 0 ]] && summary_parts+=("${cloned} cloned")
    [[ $skipped -gt 0 ]] && summary_parts+=("${skipped} already cloned")
    [[ $failed -gt 0 ]] && summary_parts+=("${failed} failed")

    if [[ ${#summary_parts[@]} -gt 0 ]]; then
        printf " ("
        local first=true
        for part in "${summary_parts[@]}"; do
            [[ "$first" != "true" ]] && printf ", "
            printf "%s" "$part"
            first=false
        done
        printf ")"
    fi
    printf "\n"

    printf "\n"
    printf "  ${DIM}Run scanners: ./zero.sh scan --org %s${NC}\n" "$org_name"
    printf "\n"
}

#############################################################################
# Dashboard Display for Parallel Scans (DEPRECATED - use render_todo_display)
#############################################################################

# Spinner frames for animated progress
SPINNER_FRAMES=('⠋' '⠙' '⠹' '⠸' '⠼' '⠴' '⠦' '⠧' '⠇' '⠏')
SPINNER_FRAME=0

# Get next spinner frame
get_spinner_frame() {
    echo "${SPINNER_FRAMES[$SPINNER_FRAME]}"
    SPINNER_FRAME=$(( (SPINNER_FRAME + 1) % ${#SPINNER_FRAMES[@]} ))
}

# ANSI cursor control
move_cursor_up() {
    local lines=$1
    printf "\033[%dA" "$lines"
}

clear_lines() {
    local lines=$1
    for ((i=0; i<lines; i++)); do
        printf "\033[K\n"
    done
    move_cursor_up "$lines"
}

# Render dashboard showing parallel scan progress
# Usage: render_scan_dashboard <status_dir> <total_scanners> <repos...>
render_scan_dashboard() {
    local status_dir="$1"
    local total_scanners="$2"
    shift 2
    local all_repos=("$@")

    # Dashboard config
    local max_slots=4  # Show up to 4 active scans
    local bar_width=30

    # Find running and queued repos
    local running_repos=()
    local queued_repos=()
    local completed_repos=()

    for repo in "${all_repos[@]}"; do
        local safe_name=$(sanitize_repo_name "$repo")
        local status_file="$status_dir/$safe_name.status"

        if [[ -f "$status_file" ]]; then
            local status=$(cut -d'|' -f1 "$status_file")
            case "$status" in
                running)
                    running_repos+=("$repo")
                    ;;
                complete|failed)
                    completed_repos+=("$repo")
                    ;;
            esac
        else
            queued_repos+=("$repo")
        fi
    done

    # Clear previous dashboard
    if [[ -n "${DASHBOARD_LINES:-}" ]]; then
        clear_lines "$DASHBOARD_LINES"
    fi

    local lines_printed=0

    # Render active scan slots
    for ((slot=0; slot<max_slots; slot++)); do
        if [[ $slot -lt ${#running_repos[@]} ]]; then
            # Render active scan
            local repo="${running_repos[$slot]}"
            local safe_name=$(sanitize_repo_name "$repo")
            local status_file="$status_dir/$safe_name.status"

            if [[ -f "$status_file" ]]; then
                IFS='|' read -r status current_scanner progress duration < "$status_file"

                # Parse progress (e.g., "10/18")
                local current=0
                local total=$total_scanners
                if [[ "$progress" =~ ^([0-9]+)/([0-9]+)$ ]]; then
                    current="${BASH_REMATCH[1]}"
                    total="${BASH_REMATCH[2]}"
                fi

                # Calculate percentage
                local pct=0
                if [[ $total -gt 0 ]]; then
                    pct=$((current * 100 / total))
                fi

                # Build mini progress bar
                local filled=$((current * bar_width / total))
                [[ $filled -gt $bar_width ]] && filled=$bar_width
                local empty=$((bar_width - filled))
                local bar=""
                for ((i=0; i<filled; i++)); do bar+="="; done
                if [[ $filled -lt $bar_width ]]; then
                    bar+=">"
                    ((empty--))
                fi
                for ((i=0; i<empty; i++)); do bar+=" "; done

                # Print scan line with spinner
                local spinner=$(get_spinner_frame)
                printf "[%s] %-35s [%s] %3d%% %s\n" \
                    "${CYAN}${spinner}${NC}" \
                    "${repo:0:35}" \
                    "$bar" \
                    "$pct" \
                    "${DIM}${current_scanner}${NC}"
            fi
        elif [[ $slot -lt $((${#running_repos[@]} + ${#queued_repos[@]})) ]]; then
            # Show queued repo
            local queue_idx=$((slot - ${#running_repos[@]}))
            local repo="${queued_repos[$queue_idx]}"
            printf "[${DIM}⋯${NC}] %-35s ${DIM}queued...${NC}\n" "${repo:0:35}"
        else
            # Empty slot
            printf "${DIM}[−] (empty)${NC}\n"
        fi
        ((lines_printed++))
    done

    # Show overall progress
    local total_repos=${#all_repos[@]}
    local completed_count=${#completed_repos[@]}
    local overall_pct=0
    if [[ $total_repos -gt 0 ]]; then
        overall_pct=$((completed_count * 100 / total_repos))
    fi

    echo
    printf "${BOLD}Overall:${NC} %d/%d repos complete (%d%%) " \
        "$completed_count" "$total_repos" "$overall_pct"

    # Show queued count if any
    if [[ ${#queued_repos[@]} -gt $((max_slots - ${#running_repos[@]})) ]]; then
        local hidden=$((${#queued_repos[@]} - (max_slots - ${#running_repos[@]})))
        printf "${DIM}(+%d queued)${NC}" "$hidden"
    fi

    echo
    ((lines_printed+=2))

    # Store line count for next clear
    export DASHBOARD_LINES=$lines_printed
}

#############################################################################
# Buffered Output System
#############################################################################

# Initialize output buffer directory
# Returns: path to buffer directory
init_output_buffer() {
    local buffer_dir=$(mktemp -d)
    echo "$buffer_dir"
}

# Start buffering output for a specific item
# Usage: start_buffer <buffer_dir> <item_id>
start_buffer() {
    local buffer_dir="$1"
    local item_id="$2"

    # Sanitize item_id for filename
    local safe_id=$(echo "$item_id" | sed 's/[^a-zA-Z0-9_-]/_/g')
    echo "" > "$buffer_dir/$safe_id.output"
}

# Append to buffer
# Usage: append_buffer <buffer_dir> <item_id> <text>
append_buffer() {
    local buffer_dir="$1"
    local item_id="$2"
    local text="$3"

    local safe_id=$(echo "$item_id" | sed 's/[^a-zA-Z0-9_-]/_/g')
    echo "$text" >> "$buffer_dir/$safe_id.output"
}

# Flush buffer to stdout
# Usage: flush_buffer <buffer_dir> <item_id>
flush_buffer() {
    local buffer_dir="$1"
    local item_id="$2"

    local safe_id=$(echo "$item_id" | sed 's/[^a-zA-Z0-9_-]/_/g')
    local buffer_file="$buffer_dir/$safe_id.output"

    if [[ -f "$buffer_file" ]]; then
        cat "$buffer_file"
    fi
}

#############################################################################
# Repository Freshness Check
#############################################################################

# Check if a cached repo needs updating by comparing local and remote HEAD
# Returns: "up-to-date", "needs-update", "error", or "no-remote"
zero_check_repo_freshness() {
    local project_id="$1"
    local repo_path=$(zero_project_repo_path "$project_id")

    # Check if repo exists
    if [[ ! -d "$repo_path/.git" ]]; then
        echo "error:not-a-git-repo"
        return 1
    fi

    cd "$repo_path" || { echo "error:cannot-access"; return 1; }

    # Get local HEAD
    local local_head=$(git rev-parse HEAD 2>/dev/null)
    if [[ -z "$local_head" ]]; then
        echo "error:no-local-head"
        return 1
    fi

    # Get current branch
    local branch=$(git rev-parse --abbrev-ref HEAD 2>/dev/null)
    if [[ -z "$branch" ]] || [[ "$branch" == "HEAD" ]]; then
        branch="master"  # Fallback for detached HEAD
    fi

    # Check remote URL exists
    local remote_url=$(git remote get-url origin 2>/dev/null)
    if [[ -z "$remote_url" ]]; then
        echo "no-remote"
        return 0
    fi

    # Fetch remote refs without updating working tree (lightweight check)
    # Note: We skip the dry-run fetch and just use ls-remote which is faster

    # Get remote HEAD for the branch
    local remote_head=$(git ls-remote origin "$branch" 2>/dev/null | cut -f1)
    if [[ -z "$remote_head" ]]; then
        # Try refs/heads/branch
        remote_head=$(git ls-remote origin "refs/heads/$branch" 2>/dev/null | cut -f1)
    fi

    if [[ -z "$remote_head" ]]; then
        echo "error:no-remote-head"
        return 1
    fi

    # Compare local and remote
    if [[ "$local_head" == "$remote_head" ]]; then
        echo "up-to-date"
    else
        # Check if remote is ahead (local is behind)
        if git merge-base --is-ancestor "$local_head" "$remote_head" 2>/dev/null; then
            echo "needs-update:$remote_head"
        else
            # Local has diverged or is ahead
            echo "diverged"
        fi
    fi

    return 0
}

# Get detailed freshness info as JSON
zero_repo_freshness_json() {
    local project_id="$1"
    local repo_path=$(zero_project_repo_path "$project_id")

    if [[ ! -d "$repo_path/.git" ]]; then
        echo '{"status": "error", "message": "not a git repository"}'
        return
    fi

    cd "$repo_path" || { echo '{"status": "error", "message": "cannot access repo"}'; return; }

    local local_head=$(git rev-parse HEAD 2>/dev/null)
    local local_short=$(git rev-parse --short HEAD 2>/dev/null)
    local branch=$(git rev-parse --abbrev-ref HEAD 2>/dev/null)
    local remote_url=$(git remote get-url origin 2>/dev/null | sed 's/[^@]*@/***@/' | sed 's/github_pat_[^@]*@/***@/')  # Mask tokens
    local last_commit_date=$(git log -1 --format=%ci 2>/dev/null)

    local freshness=$(zero_check_repo_freshness "$project_id")
    local fresh_status="unknown"
    local remote_head=""

    case "$freshness" in
        up-to-date)
            fresh_status="current"
            ;;
        needs-update:*)
            fresh_status="behind"
            remote_head=$(echo "$freshness" | cut -d: -f2)
            ;;
        diverged)
            fresh_status="diverged"
            ;;
        no-remote)
            fresh_status="local-only"
            ;;
        error:*)
            fresh_status="error"
            ;;
    esac

    jq -n \
        --arg status "$fresh_status" \
        --arg branch "$branch" \
        --arg local_head "$local_short" \
        --arg remote_head "${remote_head:0:7}" \
        --arg remote_url "$remote_url" \
        --arg last_commit "$last_commit_date" \
        '{
            status: $status,
            branch: $branch,
            local_commit: $local_head,
            remote_commit: (if $remote_head != "" then $remote_head else null end),
            remote_url: $remote_url,
            last_commit_date: $last_commit
        }'
}

# Update cached repo if remote is ahead
# Returns 0 if updated or already up-to-date, 1 on error
zero_update_repo_if_needed() {
    local project_id="$1"
    local force="${2:-false}"

    local repo_path=$(zero_project_repo_path "$project_id")

    if [[ ! -d "$repo_path/.git" ]]; then
        echo -e "${RED}Error: Not a git repository${NC}" >&2
        return 1
    fi

    cd "$repo_path" || return 1

    local freshness=$(zero_check_repo_freshness "$project_id")

    case "$freshness" in
        up-to-date)
            echo -e "${GREEN}✓${NC} Repository is up to date"
            return 0
            ;;
        needs-update:*)
            local remote_head=$(echo "$freshness" | cut -d: -f2)
            echo -e "${BLUE}Updating repository...${NC}"
            if git pull --ff-only origin 2>/dev/null; then
                local new_head=$(git rev-parse --short HEAD)
                echo -e "${GREEN}✓${NC} Updated to $new_head"

                # Update project.json with new commit
                local project_path=$(zero_project_path "$project_id")
                if [[ -f "$project_path/project.json" ]]; then
                    local tmp=$(mktemp)
                    jq --arg commit "$new_head" '.commit = $commit' "$project_path/project.json" > "$tmp" && mv "$tmp" "$project_path/project.json"
                fi
                return 0
            else
                echo -e "${RED}✗${NC} Failed to update (try --force for full re-clone)" >&2
                return 1
            fi
            ;;
        diverged)
            if [[ "$force" == "true" ]]; then
                echo -e "${YELLOW}⚠${NC} Repository has diverged, resetting to remote..."
                local branch=$(git rev-parse --abbrev-ref HEAD)
                git fetch origin
                git reset --hard "origin/$branch"
                return 0
            else
                echo -e "${YELLOW}⚠${NC} Repository has diverged from remote (use --force to reset)"
                return 1
            fi
            ;;
        no-remote)
            echo -e "${CYAN}ℹ${NC} Local-only repository (no remote to check)"
            return 0
            ;;
        error:*)
            local error_type=$(echo "$freshness" | cut -d: -f2)
            echo -e "${RED}✗${NC} Cannot check freshness: $error_type" >&2
            return 1
            ;;
    esac
}

# Get cached repo path if available and optionally check freshness
# Usage: zero_get_cached_repo <source> [--check-fresh]
# Returns: path to repo if cached, empty string if not
zero_get_cached_repo() {
    local source="$1"
    local check_fresh="${2:-}"

    local project_id=$(zero_project_id "$source")
    local repo_path=$(zero_project_repo_path "$project_id")

    if [[ ! -d "$repo_path" ]]; then
        echo ""
        return 1
    fi

    if [[ "$check_fresh" == "--check-fresh" ]]; then
        local freshness=$(zero_check_repo_freshness "$project_id")
        if [[ "$freshness" == needs-update:* ]]; then
            echo -e "${YELLOW}⚠ Cached repo is behind remote${NC}" >&2
        fi
    fi

    echo "$repo_path"
    return 0
}

#############################################################################
# Scan ID Generation
#############################################################################

# Generate unique scan ID
# Format: YYYYMMDD-HHMMSS-XXXX (timestamp + 4-char random)
generate_scan_id() {
    local timestamp=$(date -u +"%Y%m%d-%H%M%S")
    local random=$(head -c 2 /dev/urandom | xxd -p 2>/dev/null || echo "$(date +%N | cut -c1-4)")
    echo "${timestamp}-${random:0:4}"
}

# Get full git context from a repository
# Returns JSON with commit, branch, tag, date info
zero_get_git_context() {
    local repo_path="$1"

    if [[ ! -d "$repo_path/.git" ]]; then
        echo '{"error": "not a git repository"}'
        return 1
    fi

    cd "$repo_path" || { echo '{"error": "cannot access repo"}'; return 1; }

    local commit_hash=$(git rev-parse HEAD 2>/dev/null)
    local commit_short=$(git rev-parse --short HEAD 2>/dev/null)
    local branch=$(git rev-parse --abbrev-ref HEAD 2>/dev/null)
    local tag=$(git describe --tags --exact-match HEAD 2>/dev/null || echo "")
    local commit_date=$(git log -1 --format=%cI 2>/dev/null)
    local commit_author=$(git log -1 --format='%an <%ae>' 2>/dev/null)

    # Handle detached HEAD
    [[ "$branch" == "HEAD" ]] && branch=""

    jq -n \
        --arg hash "$commit_hash" \
        --arg short "$commit_short" \
        --arg branch "$branch" \
        --arg tag "$tag" \
        --arg date "$commit_date" \
        --arg author "$commit_author" \
        '{
            commit_hash: $hash,
            commit_short: $short,
            branch: (if $branch != "" then $branch else null end),
            tag: (if $tag != "" then $tag else null end),
            commit_date: $date,
            commit_author: $author
        }'
}

#############################################################################
# Scan History Management
#############################################################################

# Initialize history.json for a project
zero_init_history() {
    local project_id="$1"
    local analysis_path=$(zero_project_analysis_path "$project_id")
    local history_file="$analysis_path/history.json"

    mkdir -p "$analysis_path"

    if [[ ! -f "$history_file" ]]; then
        cat > "$history_file" << EOF
{
  "schema_version": "1.0.0",
  "project_id": "$project_id",
  "total_scans": 0,
  "first_scan_at": null,
  "last_scan_at": null,
  "scans": [],
  "by_commit": {}
}
EOF
    fi
}

# Append a scan record to history
zero_append_scan_history() {
    local project_id="$1"
    local scan_id="$2"
    local commit_hash="$3"
    local commit_short="$4"
    local branch="$5"
    local started_at="$6"
    local completed_at="$7"
    local duration_seconds="$8"
    local profile="$9"
    local scanners_run="${10}"  # JSON array
    local status="${11}"
    local summary="${12}"       # JSON object

    local analysis_path=$(zero_project_analysis_path "$project_id")
    local history_file="$analysis_path/history.json"

    # Initialize if doesn't exist
    zero_init_history "$project_id"

    # Create scan record
    local scan_record=$(jq -n \
        --arg scan_id "$scan_id" \
        --arg commit_hash "$commit_hash" \
        --arg commit_short "$commit_short" \
        --arg branch "$branch" \
        --arg started_at "$started_at" \
        --arg completed_at "$completed_at" \
        --argjson duration "$duration_seconds" \
        --arg profile "$profile" \
        --argjson scanners "$scanners_run" \
        --arg status "$status" \
        --argjson summary "$summary" \
        '{
            scan_id: $scan_id,
            commit_hash: $commit_hash,
            commit_short: $commit_short,
            branch: (if $branch != "" then $branch else null end),
            started_at: $started_at,
            completed_at: $completed_at,
            duration_seconds: $duration,
            profile: $profile,
            scanners_run: $scanners,
            status: $status,
            summary: $summary
        }')

    # Update history file
    local tmp=$(mktemp)
    jq --argjson scan "$scan_record" \
       --arg commit "$commit_hash" \
       --arg scan_id "$scan_id" \
       --arg completed_at "$completed_at" \
       '
       .scans = [$scan] + .scans |
       .total_scans = (.scans | length) |
       .last_scan_at = $completed_at |
       .first_scan_at = (.first_scan_at // $completed_at) |
       .by_commit[$commit] = ((.by_commit[$commit] // []) + [$scan_id])
       ' "$history_file" > "$tmp" && mv "$tmp" "$history_file"
}

# Get scan history for a project
zero_get_scan_history() {
    local project_id="$1"
    local limit="${2:-10}"

    local history_file="$(zero_project_analysis_path "$project_id")/history.json"

    if [[ -f "$history_file" ]]; then
        jq --argjson limit "$limit" '.scans[:$limit]' "$history_file"
    else
        echo "[]"
    fi
}

# Get scans for a specific commit
zero_get_scans_for_commit() {
    local project_id="$1"
    local commit="$2"

    local history_file="$(zero_project_analysis_path "$project_id")/history.json"

    if [[ -f "$history_file" ]]; then
        jq --arg commit "$commit" '.scans | map(select(.commit_hash == $commit or .commit_short == $commit))' "$history_file"
    else
        echo "[]"
    fi
}

#############################################################################
# Organization Index
#############################################################################

# Update org-level index after a scan
zero_update_org_index() {
    local project_id="$1"

    # Extract org from project_id (org/repo format)
    local org=$(echo "$project_id" | cut -d'/' -f1)
    local repo=$(echo "$project_id" | cut -d'/' -f2)
    local org_dir="$ZERO_PROJECTS_DIR/$org"
    local index_file="$org_dir/_index.json"

    # Create org index if doesn't exist
    if [[ ! -f "$index_file" ]]; then
        cat > "$index_file" << EOF
{
  "org": "$org",
  "updated_at": null,
  "project_count": 0,
  "aggregate": {
    "total_vulnerabilities": 0,
    "critical": 0,
    "high": 0,
    "total_dependencies": 0,
    "repos_at_risk": []
  },
  "projects": {}
}
EOF
    fi

    # Get data from manifest
    local manifest="$(zero_project_analysis_path "$project_id")/manifest.json"
    if [[ ! -f "$manifest" ]]; then
        return 1
    fi

    local scan_id=$(jq -r '.scan_id // ""' "$manifest" 2>/dev/null)
    local completed_at=$(jq -r '.completed_at // ""' "$manifest" 2>/dev/null)
    local commit=$(jq -r '.git.commit_short // .analyzed_commit // ""' "$manifest" 2>/dev/null)
    local risk_level=$(jq -r '.summary.risk_level // "unknown"' "$manifest" 2>/dev/null)
    local critical=$(jq -r '.summary.total_vulnerabilities // 0' "$manifest" 2>/dev/null)
    local deps=$(jq -r '.summary.total_dependencies // 0' "$manifest" 2>/dev/null)

    # Get vuln breakdown from package-vulns.json if available
    local vulns_file="$(zero_project_analysis_path "$project_id")/package-vulns.json"
    local crit=0 high=0
    if [[ -f "$vulns_file" ]]; then
        crit=$(jq -r '.summary.critical // 0' "$vulns_file" 2>/dev/null)
        high=$(jq -r '.summary.high // 0' "$vulns_file" 2>/dev/null)
    fi

    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Update org index
    local tmp=$(mktemp)
    jq --arg repo "$repo" \
       --arg scan_id "$scan_id" \
       --arg scan_at "$completed_at" \
       --arg commit "$commit" \
       --arg risk "$risk_level" \
       --argjson crit "$crit" \
       --argjson high "$high" \
       --argjson deps "$deps" \
       --arg ts "$timestamp" \
       '
       .updated_at = $ts |
       .projects[$repo] = {
           last_scan_id: $scan_id,
           last_scan_at: $scan_at,
           commit: $commit,
           risk_level: $risk,
           vulns: { critical: $crit, high: $high },
           deps: $deps
       } |
       .project_count = (.projects | length)
       ' "$index_file" > "$tmp" && mv "$tmp" "$index_file"

    # Recalculate aggregates
    zero_recalculate_org_aggregates "$org"
}

# Recalculate org aggregate stats
zero_recalculate_org_aggregates() {
    local org="$1"
    local index_file="$ZERO_PROJECTS_DIR/$org/_index.json"

    if [[ ! -f "$index_file" ]]; then
        return 1
    fi

    local tmp=$(mktemp)
    jq '
    .aggregate.total_vulnerabilities = ([.projects[].vulns.critical, .projects[].vulns.high] | add // 0) |
    .aggregate.critical = ([.projects[].vulns.critical] | add // 0) |
    .aggregate.high = ([.projects[].vulns.high] | add // 0) |
    .aggregate.total_dependencies = ([.projects[].deps] | add // 0) |
    .aggregate.repos_at_risk = [.projects | to_entries[] | select(.value.vulns.critical > 0 or .value.vulns.high > 0) | .key]
    ' "$index_file" > "$tmp" && mv "$tmp" "$index_file"
}

# Get org index
zero_get_org_index() {
    local org="$1"
    local index_file="$ZERO_PROJECTS_DIR/$org/_index.json"

    if [[ -f "$index_file" ]]; then
        cat "$index_file"
    else
        echo '{"error": "org index not found"}'
    fi
}

#############################################################################
# Project/Org Clean Functions
#############################################################################

# List all projects in an organization
zero_list_org_projects() {
    local org="$1"
    local org_dir="$ZERO_PROJECTS_DIR/$org"

    if [[ ! -d "$org_dir" ]]; then
        return
    fi

    for repo_dir in "$org_dir"/*/; do
        [[ ! -d "$repo_dir" ]] && continue
        [[ "$(basename "$repo_dir")" == "_index.json" ]] && continue
        local repo=$(basename "$repo_dir")
        [[ "$repo" != "_"* ]] && echo "$repo"
    done
}

# Clean a single project
zero_clean_project() {
    local project_id="$1"
    local project_path=$(zero_project_path "$project_id")

    if [[ ! -d "$project_path" ]]; then
        return 1
    fi

    # Remove project directory
    rm -rf "$project_path"

    # Remove from index
    zero_index_remove_project "$project_id"

    # Extract org and recalculate org index
    local org=$(echo "$project_id" | cut -d'/' -f1)
    local org_dir="$ZERO_PROJECTS_DIR/$org"

    # Remove org directory if empty (except for _index.json)
    local remaining=$(find "$org_dir" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | wc -l)
    if [[ "$remaining" -eq 0 ]]; then
        rm -rf "$org_dir"
    else
        # Recalculate org aggregates
        zero_recalculate_org_aggregates "$org"
    fi

    return 0
}

# Clean all projects in an organization
zero_clean_org() {
    local org="$1"
    local org_dir="$ZERO_PROJECTS_DIR/$org"
    local count=0

    if [[ ! -d "$org_dir" ]]; then
        return 1
    fi

    # Clean each project in the org
    for repo in $(zero_list_org_projects "$org"); do
        zero_clean_project "$org/$repo"
        ((count++))
    done

    # Remove org directory
    rm -rf "$org_dir"

    echo "$count"
}

# Get list of all orgs
zero_list_orgs() {
    if [[ ! -d "$ZERO_PROJECTS_DIR" ]]; then
        return
    fi

    for org_dir in "$ZERO_PROJECTS_DIR"/*/; do
        [[ ! -d "$org_dir" ]] && continue
        basename "$org_dir"
    done
}
