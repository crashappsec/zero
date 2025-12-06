#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Zero - Unified CLI for repository analysis
#
# Named after Zero Cool from the movie Hackers (1995) - the legendary hacker
# who coordinates a team of specialists to hack the planet.
#
# Usage:
#   ./zero.sh                    # Interactive mode
#   ./zero.sh check              # Verify tools and API keys
#   ./zero.sh clone <repo>       # Clone a repository (no scanning)
#   ./zero.sh scan <repo>        # Scan an already-cloned repository
#   ./zero.sh hydrate <repo>     # Clone and scan a repository
#   ./zero.sh hydrate --org <n>  # Analyze all repos in an org
#   ./zero.sh status             # Show hydrated projects
#   ./zero.sh clean              # Remove all analysis data
#############################################################################

set -e

REPO_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
UTILS_ROOT="$REPO_ROOT/utils"
ZERO_DIR="$UTILS_ROOT/zero"

# Load Zero library
source "$ZERO_DIR/lib/zero-lib.sh"

# Load .env if available
if [[ -f "$REPO_ROOT/.env" ]]; then
    set -a
    source "$REPO_ROOT/.env"
    set +a
fi

#############################################################################
# Check Functions - delegates to preflight.sh
# Use --fix to install missing tools
#############################################################################

run_check() {
    "$ZERO_DIR/scripts/preflight.sh" "$@"
}

#############################################################################
# Status Functions
#############################################################################

run_status() {
    print_zero_banner
    echo -e "${BOLD}Hydrated Projects${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo

    if [[ ! -d "$GIBSON_PROJECTS_DIR" ]]; then
        echo -e "${YELLOW}No projects hydrated yet.${NC}"
        echo
        echo "Hydrate a repository:"
        echo -e "  ${CYAN}./zero.sh hydrate owner/repo${NC}"
        return 0
    fi

    local count=0
    for org_dir in "$GIBSON_PROJECTS_DIR"/*/; do
        [[ ! -d "$org_dir" ]] && continue
        local org=$(basename "$org_dir")

        for repo_dir in "$org_dir"*/; do
            [[ ! -d "$repo_dir" ]] && continue
            local repo=$(basename "$repo_dir")
            local project_id="${org}/${repo}"
            ((count++))

            # Get project info
            local size=$(du -sh "$repo_dir" 2>/dev/null | cut -f1)
            local analysis_path="$repo_dir/analysis"
            local manifest="$analysis_path/manifest.json"

            # Get mode from manifest
            local mode="unknown"
            if [[ -f "$manifest" ]]; then
                mode=$(jq -r '.mode // "standard"' "$manifest" 2>/dev/null)
            fi

            # Mode display with color
            local mode_display=""
            case "$mode" in
                quick)    mode_display="${DIM}quick${NC}" ;;
                standard) mode_display="${CYAN}standard${NC}" ;;
                advanced) mode_display="${BLUE}advanced${NC}" ;;
                deep)     mode_display="${MAGENTA}deep${NC}" ;;
                security) mode_display="${YELLOW}security${NC}" ;;
                *)        mode_display="${DIM}$mode${NC}" ;;
            esac

            echo -e "${BOLD}$project_id${NC} ${DIM}($size)${NC} [${mode_display}]"

            # Show key metrics if available
            if [[ -f "$analysis_path/vulnerabilities.json" ]]; then
                local critical=$(jq -r '.summary.critical // 0' "$analysis_path/vulnerabilities.json" 2>/dev/null)
                local high=$(jq -r '.summary.high // 0' "$analysis_path/vulnerabilities.json" 2>/dev/null)
                if [[ "$critical" != "0" ]] || [[ "$high" != "0" ]]; then
                    echo -e "  Vulnerabilities: ${RED}$critical critical${NC}, ${YELLOW}$high high${NC}"
                else
                    echo -e "  Vulnerabilities: ${GREEN}clean${NC}"
                fi
            fi

            if [[ -f "$analysis_path/dependencies.json" ]]; then
                local deps=$(jq -r '.total_dependencies // 0' "$analysis_path/dependencies.json" 2>/dev/null)
                echo -e "  Dependencies: $deps"
            fi
            echo
        done
    done

    if [[ $count -eq 0 ]]; then
        echo -e "${YELLOW}No projects hydrated yet.${NC}"
        echo
        echo "Hydrate a repository:"
        echo -e "  ${CYAN}./zero.sh hydrate owner/repo${NC}"
    else
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo -e "Total: ${BOLD}$count${NC} projects"
        echo -e "Storage: ${CYAN}~/.zero/projects/${NC}"
    fi
}

#############################################################################
# Report Functions
#############################################################################

run_report() {
    exec "$ZERO_DIR/scripts/report.sh" "$@"
}

#############################################################################
# History Functions
#############################################################################

run_history() {
    local target="$1"
    local limit="${2:-10}"

    if [[ -z "$target" ]]; then
        echo -e "${RED}Error: No target specified${NC}" >&2
        echo "Usage: $(basename "$0") history <org/repo>"
        exit 1
    fi

    local project_id=$(gibson_project_id "$target")
    local history=$(gibson_get_scan_history "$project_id" "$limit")

    if [[ -z "$history" ]] || [[ "$history" == "null" ]]; then
        echo -e "${RED}Error: No scan history found for '$project_id'${NC}" >&2
        exit 1
    fi

    print_zero_banner
    echo -e "${BOLD}Scan History: $project_id${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo

    local total_scans=$(echo "$history" | jq -r '.total_scans // 0')
    local first_scan=$(echo "$history" | jq -r '.first_scan_at // "unknown"')
    local last_scan=$(echo "$history" | jq -r '.last_scan_at // "unknown"')

    printf "  %-14s %s\n" "Total Scans:" "$total_scans"
    printf "  %-14s %s\n" "First Scan:" "$(echo "$first_scan" | cut -d'T' -f1)"
    printf "  %-14s %s\n" "Last Scan:" "$(echo "$last_scan" | cut -d'T' -f1)"
    echo

    echo -e "${BOLD}Recent Scans${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    echo "$history" | jq -r '.scans // [] | .[] | "\(.scan_id)\t\(.started_at | split("T")[0])\t\(.profile)\t\(.status)\t\(.summary.vulnerability_count // 0) vulns"' 2>/dev/null | \
    while IFS=$'\t' read -r scan_id date profile status vulns; do
        local status_color="$GREEN"
        [[ "$status" == "failed" ]] && status_color="$RED"
        [[ "$status" == "partial" ]] && status_color="$YELLOW"

        printf "  %-24s %-12s %-10s ${status_color}%-10s${NC} %s\n" "$scan_id" "$date" "$profile" "$status" "$vulns"
    done

    echo
}

#############################################################################
# Clean Functions
#############################################################################

run_clean() {
    local target=""
    local org=""
    local dry_run=false
    local skip_confirm=false

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --org)
                org="$2"
                shift 2
                ;;
            --dry-run)
                dry_run=true
                shift
                ;;
            --yes|-y)
                skip_confirm=true
                shift
                ;;
            -*)
                echo -e "${RED}Error: Unknown option $1${NC}" >&2
                exit 1
                ;;
            *)
                target="$1"
                shift
                ;;
        esac
    done

    print_zero_banner
    echo -e "${BOLD}Clean${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo

    if [[ ! -d "$GIBSON_PROJECTS_DIR" ]]; then
        echo "No projects to clean."
        return 0
    fi

    # Determine what to clean
    if [[ -n "$target" ]]; then
        # Clean single project
        local project_id=$(gibson_project_id "$target")
        local project_path=$(gibson_project_path "$project_id")

        if [[ ! -d "$project_path" ]]; then
            echo -e "${RED}Error: Project '$project_id' not found${NC}"
            exit 1
        fi

        local size=$(du -sh "$project_path" 2>/dev/null | cut -f1)
        echo "  Project: $project_id"
        echo "  Size: $size"
        echo

        if [[ "$dry_run" == "true" ]]; then
            echo -e "${CYAN}[DRY RUN]${NC} Would remove: $project_path"
            return 0
        fi

        if [[ "$skip_confirm" != "true" ]]; then
            read -p "Remove this project? (y/n) " -n 1 -r
            echo
            [[ ! $REPLY =~ ^[Yy]$ ]] && { echo "Cancelled."; return 0; }
        fi

        gibson_clean_project "$project_id"
        echo -e "${GREEN}✓${NC} Cleaned project: $project_id"

    elif [[ -n "$org" ]]; then
        # Clean entire org
        local projects=$(gibson_list_org_projects "$org")
        if [[ -z "$projects" ]]; then
            echo -e "${RED}Error: No projects found for org '$org'${NC}"
            exit 1
        fi

        local count=$(echo "$projects" | wc -w | tr -d ' ')
        local size=$(du -sh "$GIBSON_PROJECTS_DIR/$org" 2>/dev/null | cut -f1)

        echo "  Organization: $org"
        echo "  Projects: $count"
        echo "  Size: $size"
        echo
        echo "  Projects to remove:"
        for repo in $projects; do
            echo "    - $org/$repo"
        done
        echo

        if [[ "$dry_run" == "true" ]]; then
            echo -e "${CYAN}[DRY RUN]${NC} Would remove: $GIBSON_PROJECTS_DIR/$org/"
            return 0
        fi

        if [[ "$skip_confirm" != "true" ]]; then
            read -p "Remove all projects in '$org'? (y/n) " -n 1 -r
            echo
            [[ ! $REPLY =~ ^[Yy]$ ]] && { echo "Cancelled."; return 0; }
        fi

        gibson_clean_org "$org"
        echo -e "${GREEN}✓${NC} Cleaned org: $org ($count projects)"

    else
        # Clean everything
        local count=$(find "$GIBSON_PROJECTS_DIR" -mindepth 2 -maxdepth 2 -type d 2>/dev/null | wc -l | tr -d ' ')
        local size=$(du -sh "$GIBSON_DIR" 2>/dev/null | cut -f1)

        echo -e "${YELLOW}Warning:${NC} This will remove ALL analysis data!"
        echo
        echo "  Projects: $count"
        echo "  Size: $size"
        echo "  Location: $GIBSON_DIR"
        echo

        if [[ "$dry_run" == "true" ]]; then
            echo -e "${CYAN}[DRY RUN]${NC} Would remove: $GIBSON_DIR"
            return 0
        fi

        if [[ "$skip_confirm" != "true" ]]; then
            read -p "Are you sure? (y/n) " -n 1 -r
            echo
            [[ ! $REPLY =~ ^[Yy]$ ]] && { echo "Cancelled."; return 0; }
        fi

        rm -rf "$GIBSON_DIR"
        echo -e "${GREEN}✓${NC} Cleaned all data"
    fi
}

#############################################################################
# Interactive Menu
#############################################################################

#############################################################################
# Helper: Get hydration status for a target
#############################################################################

get_hydration_status() {
    local target="$1"
    local project_id=""

    # Determine project_id from target
    if [[ "$target" == --org* ]]; then
        # Org mode - can't check individual status
        echo ""
        return
    fi

    # Convert target to project_id format
    if [[ "$target" =~ ^https://github\.com/(.+)$ ]]; then
        project_id="${BASH_REMATCH[1]%.git}"
    elif [[ "$target" =~ ^([^/]+)/([^/]+)$ ]]; then
        project_id="$target"
    else
        echo ""
        return
    fi

    # Check if project exists
    local project_path="$GIBSON_PROJECTS_DIR/${project_id//\//_}"
    project_path="$GIBSON_PROJECTS_DIR/$(echo "$project_id" | tr '/' '/')"

    # Parse as org/repo
    local org=$(echo "$project_id" | cut -d'/' -f1)
    local repo=$(echo "$project_id" | cut -d'/' -f2)
    project_path="$GIBSON_PROJECTS_DIR/$org/$repo"

    if [[ -d "$project_path/analysis" ]]; then
        local manifest="$project_path/analysis/manifest.json"
        if [[ -f "$manifest" ]]; then
            local mode=$(jq -r '.mode // "standard"' "$manifest" 2>/dev/null)
            local completed=$(jq -r '.completed_at // ""' "$manifest" 2>/dev/null)
            if [[ -n "$completed" ]] && [[ "$completed" != "null" ]]; then
                echo "$mode"
                return
            fi
        fi
    fi
    echo ""
}

# Get mode display with status
get_mode_display() {
    local mode="$1"
    local current_mode="$2"
    local mode_name="$3"
    local time_est="$4"
    local description="$5"

    if [[ "$current_mode" == "$mode" ]]; then
        echo -e "  ${CYAN}$1${NC}  ${mode_name}   ${time_est}  ${description} ${GREEN}[hydrated]${NC}"
    else
        echo -e "  ${CYAN}$1${NC}  ${mode_name}   ${time_est}  ${description}"
    fi
}

# Configuration file path (unified config)
CONFIG_FILE="$ZERO_DIR/config/zero.config.json"

# Semgrep rules configuration
SEMGREP_DIR="$UTILS_ROOT/scanners/semgrep"
SEMGREP_RULES_DIR="$SEMGREP_DIR/rules"
RAG_TECH_DIR="$REPO_ROOT/rag/technology-identification"

#############################################################################
# Semgrep Rules Update Functions
#############################################################################

# Get the last modification time of semgrep rules
get_semgrep_rules_mtime() {
    local rules_file="$SEMGREP_RULES_DIR/tech-discovery.yaml"
    if [[ -f "$rules_file" ]]; then
        stat -f %m "$rules_file" 2>/dev/null || stat -c %Y "$rules_file" 2>/dev/null || echo "0"
    else
        echo "0"
    fi
}

# Get the newest modification time from RAG patterns
get_rag_newest_mtime() {
    local newest=0
    while IFS= read -r file; do
        local mtime=$(stat -f %m "$file" 2>/dev/null || stat -c %Y "$file" 2>/dev/null || echo "0")
        if [[ $mtime -gt $newest ]]; then
            newest=$mtime
        fi
    done < <(find "$RAG_TECH_DIR" -name "patterns.md" -type f 2>/dev/null)
    echo "$newest"
}

# Get list of RAG files modified since last rules generation
get_modified_rag_files() {
    local rules_mtime="$1"
    local modified_files=()

    while IFS= read -r file; do
        local mtime=$(stat -f %m "$file" 2>/dev/null || stat -c %Y "$file" 2>/dev/null || echo "0")
        if [[ $mtime -gt $rules_mtime ]]; then
            # Extract technology name from path
            local rel_path="${file#$RAG_TECH_DIR/}"
            local tech_path="${rel_path%/patterns.md}"
            modified_files+=("$tech_path")
        fi
    done < <(find "$RAG_TECH_DIR" -name "patterns.md" -type f 2>/dev/null)

    printf '%s\n' "${modified_files[@]}"
}

# Check if semgrep rules need updating
check_semgrep_rules_status() {
    local rules_mtime=$(get_semgrep_rules_mtime)
    local rag_mtime=$(get_rag_newest_mtime)

    if [[ "$rules_mtime" == "0" ]]; then
        echo "missing"
    elif [[ $rag_mtime -gt $rules_mtime ]]; then
        echo "outdated"
    else
        echo "current"
    fi
}

# Run semgrep rules update
run_semgrep_update() {
    print_zero_banner
    echo -e "${BOLD}Semgrep Rules Update${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo

    # Check if RAG directory exists
    if [[ ! -d "$RAG_TECH_DIR" ]]; then
        echo -e "${RED}Error: RAG directory not found: $RAG_TECH_DIR${NC}"
        return 1
    fi

    # Check if converter script exists
    if [[ ! -f "$SEMGREP_DIR/rag-to-semgrep.py" ]]; then
        echo -e "${RED}Error: Converter script not found: $SEMGREP_DIR/rag-to-semgrep.py${NC}"
        return 1
    fi

    # Get current status
    local rules_mtime=$(get_semgrep_rules_mtime)
    local status=$(check_semgrep_rules_status)

    # Count RAG patterns
    local rag_count=$(find "$RAG_TECH_DIR" -name "patterns.md" -type f 2>/dev/null | wc -l | tr -d ' ')

    # Count current rules
    local current_rules=0
    if [[ -f "$SEMGREP_RULES_DIR/tech-discovery.yaml" ]]; then
        current_rules=$(grep -c "^- id:" "$SEMGREP_RULES_DIR/tech-discovery.yaml" 2>/dev/null || echo "0")
    fi

    echo -e "  RAG patterns:     ${CYAN}$rag_count${NC} technologies"
    echo -e "  Current rules:    ${CYAN}$current_rules${NC} semgrep rules"
    echo

    case "$status" in
        missing)
            echo -e "  Status:           ${YELLOW}Rules not generated${NC}"
            echo
            ;;
        outdated)
            echo -e "  Status:           ${YELLOW}Rules outdated${NC}"
            echo

            # Show modified files
            local modified_files=$(get_modified_rag_files "$rules_mtime")
            local modified_count=$(echo "$modified_files" | grep -c . || echo "0")

            echo -e "${BOLD}Modified RAG patterns ($modified_count files):${NC}"
            echo "$modified_files" | head -20 | while read -r tech; do
                [[ -n "$tech" ]] && echo -e "  ${DIM}•${NC} $tech"
            done
            if [[ $modified_count -gt 20 ]]; then
                echo -e "  ${DIM}... and $((modified_count - 20)) more${NC}"
            fi
            echo
            ;;
        current)
            echo -e "  Status:           ${GREEN}Rules up to date${NC}"
            echo
            echo -e "${DIM}No update needed. RAG patterns have not changed since last generation.${NC}"
            return 0
            ;;
    esac

    # Prompt for update
    read -p "Generate updated semgrep rules? (y/n) " -n 1 -r
    echo
    echo

    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Cancelled."
        return 0
    fi

    # Check Python is available
    if ! command -v python3 &> /dev/null; then
        echo -e "${RED}Error: python3 is required but not installed${NC}"
        return 1
    fi

    # Check PyYAML is available
    if ! python3 -c "import yaml" 2>/dev/null; then
        echo -e "${YELLOW}Installing PyYAML...${NC}"
        pip3 install pyyaml --quiet || {
            echo -e "${RED}Error: Failed to install PyYAML${NC}"
            return 1
        }
    fi

    echo -e "${BLUE}Generating semgrep rules from RAG patterns...${NC}"
    echo

    # Run the converter
    if python3 "$SEMGREP_DIR/rag-to-semgrep.py" "$RAG_TECH_DIR" "$SEMGREP_RULES_DIR"; then
        echo
        echo -e "${GREEN}✓${NC} Semgrep rules updated successfully"

        # Show new rule count
        local new_rules=$(grep -c "^- id:" "$SEMGREP_RULES_DIR/tech-discovery.yaml" 2>/dev/null || echo "0")
        echo -e "  Generated:        ${CYAN}$new_rules${NC} rules"
    else
        echo
        echo -e "${RED}✗${NC} Failed to generate semgrep rules"
        return 1
    fi
}

#############################################################################
# Malcontent Update Functions
#############################################################################

# Get installed malcontent version
get_malcontent_version() {
    local mal_bin=""
    if command -v mal &> /dev/null; then
        mal_bin="mal"
    elif [[ -x "/opt/homebrew/bin/mal" ]]; then
        mal_bin="/opt/homebrew/bin/mal"
    elif [[ -x "/usr/local/bin/mal" ]]; then
        mal_bin="/usr/local/bin/mal"
    fi

    if [[ -n "$mal_bin" ]]; then
        "$mal_bin" --version 2>/dev/null | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' | head -1
    else
        echo ""
    fi
}

# Check if malcontent update is available via brew
check_malcontent_update() {
    local current=$(get_malcontent_version)
    if [[ -z "$current" ]]; then
        echo "not_installed"
        return
    fi

    # Check brew for latest version
    if command -v brew &> /dev/null; then
        local latest=$(brew info malcontent --json 2>/dev/null | jq -r '.[0].versions.stable // empty' 2>/dev/null)
        if [[ -n "$latest" ]]; then
            # Compare versions (strip 'v' prefix)
            local current_clean="${current#v}"
            if [[ "$current_clean" != "$latest" ]]; then
                echo "outdated:$latest"
                return
            fi
        fi
    fi
    echo "current"
}

# Get malcontent rules count (from help output)
get_malcontent_rules_info() {
    # Malcontent embeds ~14,500 YARA rules from multiple vendors
    echo "~14,500 YARA rules (embedded)"
}

# Run malcontent update
run_malcontent_update() {
    print_zero_banner
    echo -e "${BOLD}Malcontent Update${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo

    local current_version=$(get_malcontent_version)
    local status=$(check_malcontent_update)

    if [[ "$status" == "not_installed" ]]; then
        echo -e "  Status:           ${RED}Not installed${NC}"
        echo
        echo "Malcontent is a supply chain compromise detection tool from Chainguard."
        echo "It includes ~14,500 YARA rules from security vendors:"
        echo "  • Avast, Elastic, FireEye, Mandiant, ReversingLabs"
        echo
        read -p "Install malcontent via brew? (y/n) " -n 1 -r
        echo
        echo

        if [[ $REPLY =~ ^[Yy]$ ]]; then
            echo -e "${BLUE}Installing malcontent...${NC}"
            if brew install malcontent; then
                echo
                echo -e "${GREEN}✓${NC} Malcontent installed successfully"
                local new_version=$(get_malcontent_version)
                echo -e "  Version:          ${CYAN}$new_version${NC}"
            else
                echo
                echo -e "${RED}✗${NC} Failed to install malcontent"
                return 1
            fi
        fi
        return 0
    fi

    echo -e "  Current version:  ${CYAN}$current_version${NC}"
    echo -e "  Rules:            $(get_malcontent_rules_info)"
    echo

    if [[ "$status" == "current" ]]; then
        echo -e "  Status:           ${GREEN}Up to date${NC}"
        echo
        echo -e "${DIM}Malcontent rules are embedded in the binary and updated with each release.${NC}"
        echo -e "${DIM}Check https://github.com/chainguard-dev/malcontent/releases for changelog.${NC}"
        return 0
    fi

    # Parse outdated status
    if [[ "$status" == outdated:* ]]; then
        local latest="${status#outdated:}"
        echo -e "  Latest version:   ${YELLOW}$latest${NC}"
        echo -e "  Status:           ${YELLOW}Update available${NC}"
        echo
        read -p "Update malcontent to $latest? (y/n) " -n 1 -r
        echo
        echo

        if [[ $REPLY =~ ^[Yy]$ ]]; then
            echo -e "${BLUE}Updating malcontent...${NC}"
            if brew upgrade malcontent; then
                echo
                echo -e "${GREEN}✓${NC} Malcontent updated successfully"
                local new_version=$(get_malcontent_version)
                echo -e "  Version:          ${CYAN}$new_version${NC}"
            else
                echo
                echo -e "${RED}✗${NC} Failed to update malcontent"
                return 1
            fi
        fi
    fi
}

# Get profile info from zero.config.json
get_profile_info() {
    local profile="$1"
    local field="$2"
    if [[ -f "$CONFIG_FILE" ]]; then
        jq -r --arg p "$profile" --arg f "$field" '.profiles[$p][$f] // empty' "$CONFIG_FILE" 2>/dev/null
    fi
}

# Get scanners for a profile
get_profile_scanners() {
    local profile="$1"
    if [[ -f "$CONFIG_FILE" ]]; then
        jq -r --arg p "$profile" '.profiles[$p].scanners // [] | join(" ")' "$CONFIG_FILE" 2>/dev/null
    fi
}

# Check if profile requires Claude API (claude_mode is "enabled" or "required")
profile_requires_claude() {
    local profile="$1"
    if [[ -f "$CONFIG_FILE" ]]; then
        local mode=$(jq -r --arg p "$profile" '.profiles[$p].claude_mode // "none"' "$CONFIG_FILE" 2>/dev/null)
        [[ "$mode" == "enabled" || "$mode" == "required" ]]
    else
        [[ "$profile" == "deep" ]]
    fi
}

# Get scanners that should use Claude for a profile
get_claude_scanners() {
    local profile="$1"
    if [[ -f "$CONFIG_FILE" ]]; then
        jq -r --arg p "$profile" '.profiles[$p].claude_scanners // [] | join(" ")' "$CONFIG_FILE" 2>/dev/null
    fi
}

# Check if a scanner supports Claude (has claude_mode "optional" or "required")
scanner_supports_claude() {
    local scanner="$1"
    if [[ -f "$CONFIG_FILE" ]]; then
        local mode=$(jq -r --arg s "$scanner" '.scanners[$s].claude_mode // "none"' "$CONFIG_FILE" 2>/dev/null)
        [[ "$mode" == "optional" || "$mode" == "required" ]]
    else
        return 1
    fi
}

# List all scanners that support Claude
list_claude_capable_scanners() {
    if [[ -f "$CONFIG_FILE" ]]; then
        jq -r '.scanners | to_entries[] | select(.value.claude_mode == "optional" or .value.claude_mode == "required") | .key' "$CONFIG_FILE" 2>/dev/null
    fi
}

# Get config setting
get_config_setting() {
    local key="$1"
    local default="$2"
    if [[ -f "$CONFIG_FILE" ]]; then
        local value=$(jq -r ".settings.$key // empty" "$CONFIG_FILE" 2>/dev/null)
        if [[ -n "$value" ]] && [[ "$value" != "null" ]]; then
            echo "$value"
            return 0
        fi
    fi
    echo "$default"
}

show_menu() {
    local first_run=true

    while true; do
        print_zero_banner

        # Get hydrated project count
        local hydrated_count=0
        if [[ -d "$GIBSON_PROJECTS_DIR" ]]; then
            hydrated_count=$(find "$GIBSON_PROJECTS_DIR" -mindepth 2 -maxdepth 2 -type d 2>/dev/null | wc -l | tr -d ' ')
        fi

        echo -e "${BOLD}What would you like to do?${NC}"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo
        echo -e "  ${CYAN}1${NC}  Check       Verify tools & config"
        echo -e "  ${CYAN}2${NC}  Clone       Clone a repository"
        echo -e "  ${CYAN}3${NC}  Scan        Scan a cloned repository"
        echo
        if [[ $hydrated_count -gt 0 ]]; then
            echo -e "  ${CYAN}a${NC}  Agent       Chat with a specialist agent ${DIM}($hydrated_count projects)${NC}"
            echo -e "  ${CYAN}r${NC}  Report      Generate analysis reports ${DIM}($hydrated_count projects)${NC}"
            echo -e "  ${CYAN}s${NC}  Status      Show hydrated projects ${DIM}($hydrated_count projects)${NC}"
        else
            echo -e "  ${CYAN}a${NC}  Agent       Chat with a specialist agent"
            echo -e "  ${CYAN}r${NC}  Report      Generate analysis reports"
            echo -e "  ${CYAN}s${NC}  Status      Show hydrated projects"
        fi
        echo -e "  ${CYAN}x${NC}  Clean       Remove all analysis data"
        echo
        # Check semgrep rules status for display
        local semgrep_status=$(check_semgrep_rules_status)
        local semgrep_indicator=""
        case "$semgrep_status" in
            missing)  semgrep_indicator=" ${YELLOW}(not generated)${NC}" ;;
            outdated) semgrep_indicator=" ${YELLOW}(update available)${NC}" ;;
            current)  semgrep_indicator=" ${GREEN}(up to date)${NC}" ;;
        esac
        echo -e "  ${CYAN}u${NC}  Update      Update semgrep rules from RAG${semgrep_indicator}"

        # Check malcontent status for display
        local malcontent_status=$(check_malcontent_update)
        local malcontent_indicator=""
        case "$malcontent_status" in
            not_installed) malcontent_indicator=" ${RED}(not installed)${NC}" ;;
            outdated:*)    malcontent_indicator=" ${YELLOW}(update available)${NC}" ;;
            current)       malcontent_indicator=" ${GREEN}($(get_malcontent_version))${NC}" ;;
        esac
        echo -e "  ${CYAN}m${NC}  Malcontent  Update malcontent YARA rules${malcontent_indicator}"
        echo
        echo -e "  ${CYAN}q${NC}  Quit"
        echo
        read -p "Choose an option: " -r
        echo
        echo

        # Build profile list for reuse
        local profile_keys=()
        if [[ -f "$CONFIG_FILE" ]]; then
            local ordered_profiles=("quick" "standard" "advanced" "deep" "security" "security-deep" "compliance" "devops" "malcontent")
            for profile in "${ordered_profiles[@]}"; do
                if jq -e --arg p "$profile" '.profiles[$p]' "$CONFIG_FILE" &>/dev/null; then
                    profile_keys+=("$profile")
                fi
            done
        else
            profile_keys=("quick" "standard" "advanced" "deep" "security" "security-deep" "compliance" "devops" "malcontent")
        fi

        case $REPLY in
            1)
                run_check || true
                echo
                read -p "Press Enter to continue..."
                ;;
            2)
                # Clone only - with depth option
                echo -e "${BOLD}Clone Repository${NC}"
                echo
                echo -e "Clone type:"
                echo -e "  ${CYAN}1${NC}  Single repository  (owner/repo)"
                echo -e "  ${CYAN}2${NC}  Organization       (all repos in org)"
                echo
                read -p "Choose [1-2, default=1]: " clone_type
                [[ -z "$clone_type" ]] && clone_type=1
                echo

                if [[ "$clone_type" == "2" ]]; then
                    # Org cloning
                    read -p "Enter organization name: " org_name
                    if [[ -n "$org_name" ]]; then
                        echo
                        read -p "Max repos to clone [default=all]: " limit

                        echo
                        echo -e "Clone depth:"
                        echo -e "  ${CYAN}1${NC}  Shallow  ${DIM}(faster, smaller, no git history)${NC}"
                        echo -e "  ${CYAN}2${NC}  Full     ${DIM}(complete history for DORA metrics)${NC}"
                        echo
                        read -p "Choose [1-2, default=1]: " depth_choice

                        # Build command arguments as array
                        local cmd_args=("--org" "$org_name")
                        [[ -n "$limit" ]] && cmd_args+=("--limit" "$limit")
                        [[ "$depth_choice" != "2" ]] && cmd_args+=("--depth" "1")

                        echo
                        "$ZERO_DIR/scripts/clone.sh" "${cmd_args[@]}" || true
                    fi
                else
                    # Single repo cloning
                    read -p "Enter repository (owner/repo): " target
                    if [[ -n "$target" ]]; then
                        echo
                        echo -e "Clone depth:"
                        echo -e "  ${CYAN}1${NC}  Shallow  ${DIM}(faster, smaller, no git history)${NC}"
                        echo -e "  ${CYAN}2${NC}  Full     ${DIM}(complete history for DORA metrics)${NC}"
                        echo
                        read -p "Choose [1-2, default=2]: " depth_choice

                        # Build command arguments as array
                        local cmd_args=()
                        [[ "$depth_choice" == "1" ]] && cmd_args+=("--depth" "1")
                        cmd_args+=("$target")

                        echo
                        "$ZERO_DIR/scripts/clone.sh" "${cmd_args[@]}" || true
                    fi
                fi
                echo
                read -p "Press Enter to continue..."
                ;;
            3)
                # Scan only - select profile
                echo -e "${BOLD}Scan Repository${NC}"
                echo
                echo -e "Scan type:"
                echo -e "  ${CYAN}1${NC}  Single repository  (owner/repo)"
                echo -e "  ${CYAN}2${NC}  Organization       (all cloned repos in org)"
                echo
                read -p "Choose [1-2, default=1]: " scan_type
                [[ -z "$scan_type" ]] && scan_type=1
                echo

                if [[ "$scan_type" == "2" ]]; then
                    # Org scanning - scan all cloned repos
                    read -p "Enter organization name: " org_name
                    if [[ -n "$org_name" ]]; then
                        # Check if org has cloned repos
                        if [[ ! -d "$GIBSON_REPOS_DIR/$org_name" ]]; then
                            echo -e "${RED}Error: No cloned repos found for org '$org_name'${NC}"
                            echo -e "Clone first with: ${CYAN}./zero.sh clone --org $org_name${NC}"
                        else
                            echo
                            echo -e "Select scan profile:"
                            local i=1
                            for profile in "${profile_keys[@]}"; do
                                local name=$(get_profile_info "$profile" "name")
                                local time=$(get_profile_info "$profile" "estimated_time")
                                local markers=""
                                [[ "$profile" == "standard" ]] && markers=" ${DIM}(recommended)${NC}"
                                if profile_requires_claude "$profile"; then
                                    markers="${markers} ${DIM}(requires API key)${NC}"
                                fi
                                printf "  ${CYAN}%s${NC}  %-12s %-10s%s\n" "$i" "${name:-$profile}" "${time:-}" "$markers"
                                ((i++))
                            done
                            echo
                            read -p "Choose [1-${#profile_keys[@]}, default=2]: " profile_choice

                            [[ -z "$profile_choice" ]] && profile_choice=2
                            local selected_idx=$((profile_choice - 1))
                            if [[ $selected_idx -ge 0 ]] && [[ $selected_idx -lt ${#profile_keys[@]} ]]; then
                                local selected_profile="${profile_keys[$selected_idx]}"
                                "$ZERO_DIR/scripts/scan.sh" --"$selected_profile" --org "$org_name" || true
                            else
                                echo -e "${RED}Invalid selection${NC}"
                            fi
                        fi
                    fi
                else
                    # Single repo scanning
                    read -p "Enter repository (owner/repo): " target
                    if [[ -n "$target" ]]; then
                        echo
                        echo -e "Select scan profile:"
                        local i=1
                        for profile in "${profile_keys[@]}"; do
                            local name=$(get_profile_info "$profile" "name")
                            local time=$(get_profile_info "$profile" "estimated_time")
                            local markers=""
                            [[ "$profile" == "standard" ]] && markers=" ${DIM}(recommended)${NC}"
                            if profile_requires_claude "$profile"; then
                                markers="${markers} ${DIM}(requires API key)${NC}"
                            fi
                            printf "  ${CYAN}%s${NC}  %-12s %-10s%s\n" "$i" "${name:-$profile}" "${time:-}" "$markers"
                            ((i++))
                        done
                        echo
                        read -p "Choose [1-${#profile_keys[@]}, default=2]: " profile_choice

                        [[ -z "$profile_choice" ]] && profile_choice=2
                        local selected_idx=$((profile_choice - 1))
                        if [[ $selected_idx -ge 0 ]] && [[ $selected_idx -lt ${#profile_keys[@]} ]]; then
                            local selected_profile="${profile_keys[$selected_idx]}"
                            "$ZERO_DIR/scripts/scan.sh" --"$selected_profile" "$target" || true
                        else
                            echo -e "${RED}Invalid selection${NC}"
                        fi
                    fi
                fi
                echo
                read -p "Press Enter to continue..."
                ;;
            a|A)
                # Agent mode info
                echo -e "${BOLD}Agent Mode${NC}"
                echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
                echo
                echo "Agent mode launches Zero, the master orchestrator."
                echo "(Named after Zero Cool from the movie Hackers)"
                echo
                echo -e "To start, use the slash command in Claude Code:"
                echo -e "  ${CYAN}/agent${NC}"
                echo
                echo "Zero coordinates these specialist agents:"
                echo "  • Cereal   - Supply chain, malware, vulnerabilities"
                echo "  • Razor    - Code security, secrets, SAST"
                echo "  • Blade    - Compliance, SOC 2, ISO 27001"
                echo "  • Phreak   - Legal, licenses, data privacy"
                echo "  • Acid     - Frontend, React, accessibility"
                echo "  • Dade     - Backend, APIs, databases"
                echo "  • Nikon    - Architecture, system design"
                echo "  • Joey     - Build, CI/CD, performance"
                echo "  • Plague   - DevOps, infrastructure"
                echo "  • Gibson   - Engineering metrics, DORA"
                echo
                read -p "Press Enter to continue..."
                ;;
            r|R)
                # Run report generator in interactive mode
                "$ZERO_DIR/scripts/report.sh" --interactive || true
                echo
                read -p "Press Enter to continue..."
                ;;
            s|S)
                run_status
                echo
                read -p "Press Enter to continue..."
                ;;
            x|X)
                run_clean
                echo
                read -p "Press Enter to continue..."
                ;;
            u|U)
                run_semgrep_update || true
                echo
                read -p "Press Enter to continue..."
                ;;
            m|M)
                run_malcontent_update || true
                echo
                read -p "Press Enter to continue..."
                ;;
            q|Q)
                echo "Goodbye!"
                exit 0
                ;;
            *)
                echo "Invalid option"
                sleep 1
                ;;
        esac
    done
}

#############################################################################
# Usage
#############################################################################

usage() {
    cat << EOF
Zero - Repository Analysis CLI
Named after Zero Cool from the movie Hackers (1995)

Usage: $(basename "$0") [command] [options]

COMMANDS:
    (none)              Chat with Zero Cool - tell him what you need
    menu                Interactive CLI menu
    check               Verify tools and configuration
    clone <repo>        Clone a repository (no scanning)
    scan <repo>         Scan an already-cloned repository
    hydrate <repo>      Clone and scan a repository (e.g., expressjs/express)
    hydrate --org <n>   Analyze all repos in an organization
    agent               Chat with a specialist agent (interactive picker)
    agent <name>        Chat with a specific agent (cereal, razor, zero, etc.)
    status              Show hydrated projects
    report <repo>       Generate summary report for a project
    history <repo>      Show scan history for a project
    clean               Remove analysis data (all, org, or project)
    update-rules        Update semgrep rules from RAG patterns
    update-malcontent   Update malcontent YARA rules (via brew)
    help                Show this help

OPTIONS FOR HYDRATE:
    --org <name>        Process all repos in organization
    --limit <n>         Max repos to process (org mode)
    --quick             Fast static analysis (~30s)
    --standard          Most scanners (~2min) [default]
    --advanced          All static scanners + health/provenance (~5min)
    --deep              Claude-assisted analysis (~10min)
    --security          Security-focused analysis (~3min)
    --compliance        License and policy compliance (~2min)
    --devops            CI/CD and operational metrics (~3min)
    --malcontent        Supply chain compromise detection (~2min)
    --force             Re-analyze even if exists

OPTIONS FOR REPORT:
    <org/repo>          Report for a specific project
    --org <name>        Aggregate report for an organization
    --json              Output in JSON format

OPTIONS FOR CLEAN:
    (no args)           Clean all data (with confirmation)
    <org/repo>          Clean a specific project
    --org <name>        Clean all projects in an organization
    --dry-run           Preview what would be deleted
    --yes               Skip confirmation prompt

CONFIGURATION:
    All settings are in utils/zero/config/zero.config.json
    See zero.config.example.json for full documentation
    Create custom profiles by adding entries to the profiles section

AGENTS (Hackers movie inspired):
    cereal    Supply chain, malware detection (Cereal Killer)
    razor     Code security, SAST, secrets (Razor)
    blade     Compliance, SOC 2, ISO 27001 (Blade)
    phreak    Legal, licenses, data privacy (Phantom Phreak)
    acid      Frontend, React, accessibility (Acid Burn)
    dade      Backend, APIs, databases (Dade Murphy)
    nikon     Architecture, system design (Lord Nikon)
    joey      Build, CI/CD, performance (Joey)
    plague    DevOps, infrastructure (The Plague)
    gibson    Engineering metrics, DORA (The Gibson)

EXAMPLES:
    $(basename "$0")                              # Interactive mode
    $(basename "$0") setup                        # First-time setup
    $(basename "$0") hydrate lodash/lodash        # Single repo
    $(basename "$0") hydrate --org expressjs      # All org repos
    $(basename "$0") agent                        # Interactive agent selection
    $(basename "$0") agent cereal                 # Chat with Cereal
    $(basename "$0") agent cereal expressjs/express # Cereal on specific project
    $(basename "$0") status                       # List projects
    $(basename "$0") report expressjs/express     # Project report
    $(basename "$0") report --org expressjs       # Org report
    $(basename "$0") history expressjs/express    # Scan history
    $(basename "$0") clean expressjs/express      # Clean one project
    $(basename "$0") clean --org expressjs        # Clean org
    $(basename "$0") update-rules                 # Update semgrep rules from RAG

STORAGE:
    Analysis data is stored in ~/.zero/projects/

EOF
    exit 0
}

#############################################################################
# Main
#############################################################################

main() {
    case "${1:-}" in
        "")
            # Default: launch Zero chat - the universal orchestrator
            exec "$ZERO_DIR/scripts/agent.sh" zero
            ;;
        menu)
            # Interactive menu for CLI browsing
            show_menu
            ;;
        setup)
            # setup is now just an alias for check
            run_check
            ;;
        check|preflight)
            run_check
            ;;
        clone)
            shift
            if [[ $# -eq 0 ]]; then
                echo -e "${RED}Error: No target specified${NC}"
                echo "Usage: $(basename "$0") clone owner/repo"
                echo "       $(basename "$0") clone --org orgname"
                exit 1
            fi
            # Delegate to clone.sh
            exec "$ZERO_DIR/scripts/clone.sh" "$@"
            ;;
        scan)
            shift
            if [[ $# -eq 0 ]]; then
                echo -e "${RED}Error: No target specified${NC}"
                echo "Usage: $(basename "$0") scan owner/repo [--quick|--standard|--deep]"
                echo "       $(basename "$0") scan --org orgname [--quick|--standard|--deep]"
                exit 1
            fi
            # Check preflight first
            if ! run_check > /dev/null 2>&1; then
                echo -e "${YELLOW}Warning: Preflight check has issues. Run './zero.sh check' to see details.${NC}"
                echo
            fi
            # Delegate to scan.sh
            exec "$ZERO_DIR/scripts/scan.sh" "$@"
            ;;
        hydrate|bootstrap)
            shift
            if [[ $# -eq 0 ]]; then
                echo -e "${RED}Error: No target specified${NC}"
                echo "Usage: $(basename "$0") hydrate owner/repo"
                echo "       $(basename "$0") hydrate --org orgname"
                exit 1
            fi
            # Check preflight first
            if ! run_check > /dev/null 2>&1; then
                echo -e "${YELLOW}Warning: Preflight check has issues. Run './zero.sh check' to see details.${NC}"
                echo
            fi
            # Delegate to hydrate.sh (clone + scan)
            exec "$ZERO_DIR/scripts/hydrate.sh" "$@"
            ;;
        status|list)
            run_status
            ;;
        agent|chat)
            # Launch interactive agent chat
            shift
            exec "$ZERO_DIR/scripts/agent.sh" "$@"
            ;;
        ask)
            shift
            exec "$ZERO_DIR/scripts/agent.sh" "$@"
            ;;
        report)
            shift
            run_report "$@"
            ;;
        history)
            shift
            run_history "$@"
            ;;
        clean)
            shift
            run_clean "$@"
            ;;
        update-rules|update)
            run_semgrep_update
            ;;
        update-malcontent)
            run_malcontent_update
            ;;
        -h|--help|help)
            usage
            ;;
        *)
            echo -e "${RED}Unknown command: $1${NC}"
            echo "Run '$(basename "$0") help' for usage"
            exit 1
            ;;
    esac
}

main "$@"
