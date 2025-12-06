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

# Zero root directory - defaults to .zero in the gibson-powers repo root
# This makes zero data visible in IDEs and keeps it project-scoped
# Can be overridden with ZERO_HOME environment variable
_ZERO_LIB_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
_ZERO_REPO_ROOT="$(dirname "$(dirname "$(dirname "$_ZERO_LIB_DIR")")")"
export GIBSON_DIR="${ZERO_HOME:-$_ZERO_REPO_ROOT/.zero}"
export GIBSON_REPOS_DIR="$GIBSON_DIR/repos"
# Legacy alias for compatibility
export GIBSON_PROJECTS_DIR="$GIBSON_REPOS_DIR"
export GIBSON_VERSION="1.0.0"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
WHITE='\033[0;37m'
BOLD='\033[1m'
DIM='\033[2m'
NC='\033[0m'

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

# Initialize ~/.zero/ directory structure
# Creates all necessary directories and config files if they don't exist
gibson_init() {
    local force="${1:-false}"

    # Check if already initialized
    if [[ -f "$GIBSON_DIR/config.json" ]] && [[ "$force" != "true" ]]; then
        return 0
    fi

    echo -e "${CYAN}Initializing Zero directory at ~/.zero...${NC}"

    # Create directory structure
    mkdir -p "$GIBSON_DIR"
    mkdir -p "$GIBSON_PROJECTS_DIR"
    mkdir -p "$GIBSON_DIR/cache"

    # Create config.json if it doesn't exist
    if [[ ! -f "$GIBSON_DIR/config.json" ]]; then
        cat > "$GIBSON_DIR/config.json" << 'EOF'
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
            jq --arg ts "$timestamp" '.created_at = $ts' "$GIBSON_DIR/config.json" > "$tmp" && mv "$tmp" "$GIBSON_DIR/config.json"
        fi
    fi

    # Create index.json if it doesn't exist
    if [[ ! -f "$GIBSON_DIR/index.json" ]]; then
        cat > "$GIBSON_DIR/index.json" << 'EOF'
{
  "version": "1.0.0",
  "projects": {},
  "active": null
}
EOF
    fi

    echo -e "${GREEN}✓${NC} Zero initialized at ~/.zero"
    return 0
}

# Check if Gibson is initialized
gibson_is_initialized() {
    [[ -d "$GIBSON_DIR" ]] && [[ -f "$GIBSON_DIR/config.json" ]] && [[ -f "$GIBSON_DIR/index.json" ]]
}

# Ensure Gibson is initialized (auto-init if not)
gibson_ensure_initialized() {
    if ! gibson_is_initialized; then
        gibson_init
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
gibson_project_id() {
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
gibson_clone_url() {
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
gibson_is_local_source() {
    local source="$1"
    [[ -d "$source" ]] || [[ "$source" =~ ^\./ ]] || [[ "$source" =~ ^/ ]]
}

#############################################################################
# Project Management
#############################################################################

# Get project directory path
gibson_project_path() {
    local project_id="$1"
    echo "$GIBSON_PROJECTS_DIR/$project_id"
}

# Get project repo path
gibson_project_repo_path() {
    local project_id="$1"
    echo "$GIBSON_PROJECTS_DIR/$project_id/repo"
}

# Get project analysis path
gibson_project_analysis_path() {
    local project_id="$1"
    echo "$GIBSON_PROJECTS_DIR/$project_id/analysis"
}

# Check if project exists
gibson_project_exists() {
    local project_id="$1"
    local project_path=$(gibson_project_path "$project_id")
    [[ -d "$project_path" ]]
}

# Check if project is fully hydrated (has repo and completed analysis)
# Returns 0 if hydrated, 1 if not
gibson_is_hydrated() {
    local project_id="$1"

    local project_path=$(gibson_project_path "$project_id")
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
gibson_require_hydrated() {
    local project_id=$(gibson_active_project)

    if [[ -z "$project_id" ]]; then
        echo -e "${RED}No active project.${NC}" >&2
        echo -e "Run ${CYAN}/zero hydrate <repo>${NC} first." >&2
        return 1
    fi

    if ! gibson_is_hydrated "$project_id"; then
        echo -e "${RED}Project '$project_id' is not fully hydrated.${NC}" >&2
        echo -e "Run ${CYAN}/zero hydrate $project_id --force${NC} to complete hydration." >&2
        return 1
    fi

    return 0
}

# Get hydration status as JSON
gibson_hydration_status() {
    local project_id="$1"

    if [[ -z "$project_id" ]]; then
        project_id=$(gibson_active_project)
    fi

    if [[ -z "$project_id" ]]; then
        echo '{"hydrated": false, "reason": "no_active_project"}'
        return
    fi

    local project_path=$(gibson_project_path "$project_id")
    local has_project=$(gibson_project_exists "$project_id" && echo "true" || echo "false")
    local has_repo=$([[ -d "$project_path/repo" ]] && echo "true" || echo "false")
    local has_manifest=$([[ -f "$project_path/analysis/manifest.json" ]] && echo "true" || echo "false")
    local proj_status=$(jq -r --arg id "$project_id" '.projects[$id].status // "unknown"' "$GIBSON_DIR/index.json" 2>/dev/null)

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
gibson_list_hydrated() {
    local projects=$(gibson_list_projects)

    if [[ -z "$projects" ]]; then
        echo "[]"
        return
    fi

    local hydrated_list="[]"

    while IFS= read -r project_id; do
        [[ -z "$project_id" ]] && continue
        if gibson_is_hydrated "$project_id"; then
            hydrated_list=$(echo "$hydrated_list" | jq --arg id "$project_id" '. + [$id]')
        fi
    done <<< "$projects"

    echo "$hydrated_list"
}

# List all projects by scanning directory structure
gibson_list_projects() {
    if [[ ! -d "$GIBSON_PROJECTS_DIR" ]]; then
        return
    fi

    # Scan for org/repo directories that have an analysis folder
    for org_dir in "$GIBSON_PROJECTS_DIR"/*/; do
        [[ ! -d "$org_dir" ]] && continue
        local org=$(basename "$org_dir")

        for repo_dir in "$org_dir"*/; do
            [[ ! -d "$repo_dir" ]] && continue
            local repo=$(basename "$repo_dir")
            echo "${org}/${repo}"
        done
    done
}

# Get active project
gibson_active_project() {
    if [[ ! -f "$GIBSON_DIR/index.json" ]]; then
        echo ""
        return
    fi
    jq -r '.active // ""' "$GIBSON_DIR/index.json" 2>/dev/null || echo ""
}

# Set active project
gibson_set_active_project() {
    local project_id="$1"

    if [[ ! -f "$GIBSON_DIR/index.json" ]]; then
        return 1
    fi

    local tmp=$(mktemp)
    jq --arg id "$project_id" '.active = $id' "$GIBSON_DIR/index.json" > "$tmp" && mv "$tmp" "$GIBSON_DIR/index.json"
}

# Add project to index
gibson_index_add_project() {
    local project_id="$1"
    local source="$2"
    local status="${3:-bootstrapping}"

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
       }' "$GIBSON_DIR/index.json" > "$tmp" && mv "$tmp" "$GIBSON_DIR/index.json"
}

# Update project status in index
gibson_index_update_status() {
    local project_id="$1"
    local status="$2"

    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    local tmp=$(mktemp)
    jq --arg id "$project_id" \
       --arg st "$status" \
       --arg ts "$timestamp" \
       '.projects[$id].status = $st | .projects[$id].last_analyzed = $ts' \
       "$GIBSON_DIR/index.json" > "$tmp" && mv "$tmp" "$GIBSON_DIR/index.json"
}

# Remove project from index
gibson_index_remove_project() {
    local project_id="$1"

    local tmp=$(mktemp)
    jq --arg id "$project_id" 'del(.projects[$id])' "$GIBSON_DIR/index.json" > "$tmp" && mv "$tmp" "$GIBSON_DIR/index.json"

    # Clear active if it was this project
    local active=$(gibson_active_project)
    if [[ "$active" == "$project_id" ]]; then
        gibson_set_active_project ""
    fi
}

#############################################################################
# Project Metadata
#############################################################################

# Create project.json for a new project
gibson_create_project_metadata() {
    local project_id="$1"
    local source="$2"
    local source_type="$3"  # github, local
    local branch="$4"
    local commit="$5"

    local project_path=$(gibson_project_path "$project_id")
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
gibson_update_project_type() {
    local project_id="$1"
    local languages="$2"      # JSON array string
    local frameworks="$3"     # JSON array string
    local package_managers="$4"  # JSON array string

    local project_path=$(gibson_project_path "$project_id")
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
gibson_read_project_metadata() {
    local project_id="$1"
    local project_path=$(gibson_project_path "$project_id")

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
gibson_init_analysis_manifest() {
    local project_id="$1"
    local commit="$2"
    local mode="${3:-standard}"
    local scan_id="${4:-}"
    local git_context="${5:-}"

    local analysis_path=$(gibson_project_analysis_path "$project_id")
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
gibson_analysis_start() {
    local project_id="$1"
    local analysis_type="$2"
    local analyzer_script="$3"
    local analyzer_version="${4:-1.0.0}"

    local analysis_path=$(gibson_project_analysis_path "$project_id")
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
gibson_analysis_complete() {
    local project_id="$1"
    local analysis_type="$2"
    local status="$3"  # complete, failed, partial
    local duration_ms="$4"
    local summary="$5"  # JSON object string

    local analysis_path=$(gibson_project_analysis_path "$project_id")
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
gibson_finalize_manifest() {
    local project_id="$1"

    local analysis_path=$(gibson_project_analysis_path "$project_id")
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
gibson_update_summary() {
    local project_id="$1"
    local risk_level="$2"
    local total_deps="$3"
    local direct_deps="$4"
    local total_vulns="$5"
    local total_findings="$6"
    local license_status="$7"
    local abandoned="$8"

    local analysis_path=$(gibson_project_analysis_path "$project_id")
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
gibson_print_header() {
    print_zero_banner
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo
}

# Print status line with checkmark or X
gibson_print_status() {
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
gibson_time_ago() {
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
gibson_project_size() {
    local project_id="$1"
    local project_path=$(gibson_project_path "$project_id")

    if [[ -d "$project_path" ]]; then
        du -sh "$project_path" 2>/dev/null | cut -f1
    else
        echo "0"
    fi
}

# Calculate total Gibson disk usage
gibson_total_size() {
    if [[ -d "$GIBSON_DIR" ]]; then
        du -sh "$GIBSON_DIR" 2>/dev/null | cut -f1
    else
        echo "0"
    fi
}

#############################################################################
# Repository Freshness Check
#############################################################################

# Check if a cached repo needs updating by comparing local and remote HEAD
# Returns: "up-to-date", "needs-update", "error", or "no-remote"
gibson_check_repo_freshness() {
    local project_id="$1"
    local repo_path=$(gibson_project_repo_path "$project_id")

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
gibson_repo_freshness_json() {
    local project_id="$1"
    local repo_path=$(gibson_project_repo_path "$project_id")

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

    local freshness=$(gibson_check_repo_freshness "$project_id")
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
gibson_update_repo_if_needed() {
    local project_id="$1"
    local force="${2:-false}"

    local repo_path=$(gibson_project_repo_path "$project_id")

    if [[ ! -d "$repo_path/.git" ]]; then
        echo -e "${RED}Error: Not a git repository${NC}" >&2
        return 1
    fi

    cd "$repo_path" || return 1

    local freshness=$(gibson_check_repo_freshness "$project_id")

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
                local project_path=$(gibson_project_path "$project_id")
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
# Usage: gibson_get_cached_repo <source> [--check-fresh]
# Returns: path to repo if cached, empty string if not
gibson_get_cached_repo() {
    local source="$1"
    local check_fresh="${2:-}"

    local project_id=$(gibson_project_id "$source")
    local repo_path=$(gibson_project_repo_path "$project_id")

    if [[ ! -d "$repo_path" ]]; then
        echo ""
        return 1
    fi

    if [[ "$check_fresh" == "--check-fresh" ]]; then
        local freshness=$(gibson_check_repo_freshness "$project_id")
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
gibson_get_git_context() {
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
gibson_init_history() {
    local project_id="$1"
    local analysis_path=$(gibson_project_analysis_path "$project_id")
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
gibson_append_scan_history() {
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

    local analysis_path=$(gibson_project_analysis_path "$project_id")
    local history_file="$analysis_path/history.json"

    # Initialize if doesn't exist
    gibson_init_history "$project_id"

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
gibson_get_scan_history() {
    local project_id="$1"
    local limit="${2:-10}"

    local history_file="$(gibson_project_analysis_path "$project_id")/history.json"

    if [[ -f "$history_file" ]]; then
        jq --argjson limit "$limit" '.scans[:$limit]' "$history_file"
    else
        echo "[]"
    fi
}

# Get scans for a specific commit
gibson_get_scans_for_commit() {
    local project_id="$1"
    local commit="$2"

    local history_file="$(gibson_project_analysis_path "$project_id")/history.json"

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
gibson_update_org_index() {
    local project_id="$1"

    # Extract org from project_id (org/repo format)
    local org=$(echo "$project_id" | cut -d'/' -f1)
    local repo=$(echo "$project_id" | cut -d'/' -f2)
    local org_dir="$GIBSON_PROJECTS_DIR/$org"
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
    local manifest="$(gibson_project_analysis_path "$project_id")/manifest.json"
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
    local vulns_file="$(gibson_project_analysis_path "$project_id")/package-vulns.json"
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
    gibson_recalculate_org_aggregates "$org"
}

# Recalculate org aggregate stats
gibson_recalculate_org_aggregates() {
    local org="$1"
    local index_file="$GIBSON_PROJECTS_DIR/$org/_index.json"

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
gibson_get_org_index() {
    local org="$1"
    local index_file="$GIBSON_PROJECTS_DIR/$org/_index.json"

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
gibson_list_org_projects() {
    local org="$1"
    local org_dir="$GIBSON_PROJECTS_DIR/$org"

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
gibson_clean_project() {
    local project_id="$1"
    local project_path=$(gibson_project_path "$project_id")

    if [[ ! -d "$project_path" ]]; then
        return 1
    fi

    # Remove project directory
    rm -rf "$project_path"

    # Remove from index
    gibson_index_remove_project "$project_id"

    # Extract org and recalculate org index
    local org=$(echo "$project_id" | cut -d'/' -f1)
    local org_dir="$GIBSON_PROJECTS_DIR/$org"

    # Remove org directory if empty (except for _index.json)
    local remaining=$(find "$org_dir" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | wc -l)
    if [[ "$remaining" -eq 0 ]]; then
        rm -rf "$org_dir"
    else
        # Recalculate org aggregates
        gibson_recalculate_org_aggregates "$org"
    fi

    return 0
}

# Clean all projects in an organization
gibson_clean_org() {
    local org="$1"
    local org_dir="$GIBSON_PROJECTS_DIR/$org"
    local count=0

    if [[ ! -d "$org_dir" ]]; then
        return 1
    fi

    # Clean each project in the org
    for repo in $(gibson_list_org_projects "$org"); do
        gibson_clean_project "$org/$repo"
        ((count++))
    done

    # Remove org directory
    rm -rf "$org_dir"

    echo "$count"
}

# Get list of all orgs
gibson_list_orgs() {
    if [[ ! -d "$GIBSON_PROJECTS_DIR" ]]; then
        return
    fi

    for org_dir in "$GIBSON_PROJECTS_DIR"/*/; do
        [[ ! -d "$org_dir" ]] && continue
        basename "$org_dir"
    done
}
