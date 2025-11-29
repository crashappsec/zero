#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Gibson Directory Management Library
# Manages ~/.phantom/ directory structure for Phantom orchestrator
#############################################################################

# Gibson root directory - defaults to ~/.phantom (user's home directory)
# Can be overridden with PHANTOM_HOME environment variable
export GIBSON_DIR="${PHANTOM_HOME:-$HOME/.phantom}"
export GIBSON_PROJECTS_DIR="$GIBSON_DIR/projects"
export GIBSON_VERSION="1.0.0"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
BOLD='\033[1m'
DIM='\033[2m'
NC='\033[0m'

#############################################################################
# PHANTOM Banner
#############################################################################

# Raw banner text (no colors, for piping to effects)
PHANTOM_BANNER='  ██████╗ ██╗  ██╗ █████╗ ███╗   ██╗████████╗ ██████╗ ███╗   ███╗
  ██╔══██╗██║  ██║██╔══██╗████╗  ██║╚══██╔══╝██╔═══██╗████╗ ████║
  ██████╔╝███████║███████║██╔██╗ ██║   ██║   ██║   ██║██╔████╔██║
  ██╔═══╝ ██╔══██║██╔══██║██║╚██╗██║   ██║   ██║   ██║██║╚██╔╝██║
  ██║     ██║  ██║██║  ██║██║ ╚████║   ██║   ╚██████╔╝██║ ╚═╝ ██║
  ╚═╝     ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═══╝   ╚═╝    ╚═════╝ ╚═╝     ╚═╝
  crashoverride.com'

# Check if terminal text effects is available
has_tte() {
    python3 -c "import terminaltexteffects" 2>/dev/null
}

# Available effects for random selection
PHANTOM_EFFECTS=(burn decrypt slide)

# Print animated banner using terminal text effects
# Falls back to static banner if tte not available
print_phantom_banner_animated() {
    local effect="${1:-random}"

    # Pick random effect if not specified or "random"
    if [[ "$effect" == "random" ]]; then
        effect="${PHANTOM_EFFECTS[$RANDOM % ${#PHANTOM_EFFECTS[@]}]}"
    fi

    # Check if tte is available and terminal is interactive
    if has_tte && [[ -t 1 ]]; then
        case "$effect" in
            burn)
                echo "$PHANTOM_BANNER" | python3 -m terminaltexteffects burn \
                    --starting-color 444444 \
                    --burn-colors ff6600 ff3300 cc0000 660000 \
                    --final-gradient-stops 9933ff cc66ff \
                    --final-gradient-steps 8 \
                    --final-gradient-direction vertical 2>/dev/null
                ;;
            decrypt)
                echo "$PHANTOM_BANNER" | python3 -m terminaltexteffects decrypt \
                    --typing-speed 2 \
                    --ciphertext-colors 00ff00 00cc00 009900 \
                    --final-gradient-stops 9933ff cc66ff \
                    --final-gradient-steps 8 2>/dev/null
                ;;
            slide)
                echo "$PHANTOM_BANNER" | python3 -m terminaltexteffects slide \
                    --movement-speed 0.5 \
                    --grouping row \
                    --final-gradient-stops 9933ff cc66ff \
                    --final-gradient-direction vertical 2>/dev/null
                ;;
            *)
                # Unknown effect, use static
                print_phantom_banner
                return
                ;;
        esac
        echo
    else
        # Fallback to static banner
        print_phantom_banner
    fi
}

# Print static banner (original behavior)
print_phantom_banner() {
    echo -e "${MAGENTA}"
    cat << 'BANNER'
  ██████╗ ██╗  ██╗ █████╗ ███╗   ██╗████████╗ ██████╗ ███╗   ███╗
  ██╔══██╗██║  ██║██╔══██╗████╗  ██║╚══██╔══╝██╔═══██╗████╗ ████║
  ██████╔╝███████║███████║██╔██╗ ██║   ██║   ██║   ██║██╔████╔██║
  ██╔═══╝ ██╔══██║██╔══██║██║╚██╗██║   ██║   ██║   ██║██║╚██╔╝██║
  ██║     ██║  ██║██║  ██║██║ ╚████║   ██║   ╚██████╔╝██║ ╚═╝ ██║
  ╚═╝     ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═══╝   ╚═╝    ╚═════╝ ╚═╝     ╚═╝
BANNER
    echo -e "${DIM}  crashoverride.com${NC}"
    echo
}

#############################################################################
# Directory Initialization
#############################################################################

# Initialize ~/.phantom/ directory structure
# Creates all necessary directories and config files if they don't exist
gibson_init() {
    local force="${1:-false}"

    # Check if already initialized
    if [[ -f "$GIBSON_DIR/config.json" ]] && [[ "$force" != "true" ]]; then
        return 0
    fi

    echo -e "${CYAN}Initializing Phantom directory at ~/.phantom...${NC}"

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

    echo -e "${GREEN}✓${NC} Phantom initialized at ~/.phantom"
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
        echo -e "Run ${CYAN}/phantom hydrate <repo>${NC} first." >&2
        return 1
    fi

    if ! gibson_is_hydrated "$project_id"; then
        echo -e "${RED}Project '$project_id' is not fully hydrated.${NC}" >&2
        echo -e "Run ${CYAN}/phantom hydrate $project_id --force${NC} to complete hydration." >&2
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

    local analysis_path=$(gibson_project_analysis_path "$project_id")
    mkdir -p "$analysis_path"

    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    cat > "$analysis_path/manifest.json" << EOF
{
  "project_id": "$project_id",
  "analyzed_commit": "$commit",
  "mode": "$mode",
  "started_at": "$timestamp",
  "completed_at": null,
  "analyses": {},
  "summary": {
    "risk_level": "unknown",
    "total_dependencies": 0,
    "direct_dependencies": 0,
    "total_vulnerabilities": 0,
    "total_security_findings": 0,
    "license_status": "unknown",
    "abandoned_packages": 0
  }
}
EOF
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

    local tmp=$(mktemp)
    jq --arg ts "$timestamp" '.completed_at = $ts' "$manifest" > "$tmp" && mv "$tmp" "$manifest"
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
    print_phantom_banner
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
    local then=$(date -j -f "%Y-%m-%dT%H:%M:%SZ" "$timestamp" +%s 2>/dev/null || date -d "$timestamp" +%s 2>/dev/null)

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
