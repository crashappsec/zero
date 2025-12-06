#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Semgrep Community Rules Manager
#
# Downloads and maintains Semgrep community rules from the official registry.
# Rules are cached locally and updated periodically.
#
# Usage: ./community-rules.sh [command] [options]
#
# Commands:
#   sync        Download/update community rules (default)
#   status      Show current rules status
#   list        List available rule packs
#   clean       Remove cached rules
#############################################################################

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_ROOT="$(dirname "$(dirname "$(dirname "$SCRIPT_DIR")")")"

# Community rules storage location - in RAG under semgrep
COMMUNITY_RULES_DIR="${SEMGREP_COMMUNITY_DIR:-$REPO_ROOT/rag/semgrep/community-rules}"
RULES_CACHE_FILE="$COMMUNITY_RULES_DIR/.cache-info.json"
CACHE_MAX_AGE_HOURS=24  # Update rules if older than this

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
DIM='\033[0;90m'
BOLD='\033[1m'
NC='\033[0m'

#############################################################################
# Rule Pack Definitions
#
# Semgrep registry rule packs - curated collections for different use cases
# See: https://semgrep.dev/explore
#############################################################################

# Security-focused rule packs (for --security profile)
SECURITY_PACKS=(
    "p/security-audit"           # Comprehensive security audit rules
    "p/owasp-top-ten"            # OWASP Top 10 vulnerabilities
    "p/cwe-top-25"               # CWE Top 25 dangerous weaknesses
    "p/secrets"                  # Secret/credential detection
    "p/supply-chain"             # Supply chain security
    "p/command-injection"        # Command injection patterns
    "p/sql-injection"            # SQL injection patterns
    "p/xss"                      # Cross-site scripting
    "p/insecure-transport"       # TLS/SSL issues
)

# Code quality rule packs
QUALITY_PACKS=(
    "p/best-practices"           # Language best practices
    "p/maintainability"          # Code maintainability
)

# Language-specific security packs
LANG_SECURITY_PACKS=(
    "p/python"                   # Python security + quality
    "p/javascript"               # JavaScript/TypeScript security
    "p/java"                     # Java security
    "p/go"                       # Go security
    "p/ruby"                     # Ruby security
    "p/php"                      # PHP security
    "p/c"                        # C/C++ security
    "p/rust"                     # Rust security
)

# Default packs for standard scanning
DEFAULT_PACKS=(
    "p/security-audit"
    "p/secrets"
    "p/owasp-top-ten"
)

#############################################################################
# Helper Functions
#############################################################################

log() {
    echo -e "${BLUE}[semgrep-rules]${NC} $1" >&2
}

log_success() {
    echo -e "${GREEN}✓${NC} $1" >&2
}

log_warn() {
    echo -e "${YELLOW}⚠${NC} $1" >&2
}

log_error() {
    echo -e "${RED}✗${NC} $1" >&2
}

check_semgrep() {
    if ! command -v semgrep &> /dev/null; then
        log_error "Semgrep not installed. Install with: brew install semgrep"
        exit 1
    fi
}

ensure_dirs() {
    mkdir -p "$COMMUNITY_RULES_DIR"
    mkdir -p "$COMMUNITY_RULES_DIR/security"
    mkdir -p "$COMMUNITY_RULES_DIR/quality"
    mkdir -p "$COMMUNITY_RULES_DIR/languages"
}

# Check if rules need updating based on cache age
needs_update() {
    local profile="${1:-default}"
    local cache_file="$COMMUNITY_RULES_DIR/$profile/.last-update"

    if [[ ! -f "$cache_file" ]]; then
        return 0  # No cache, needs update
    fi

    local last_update=$(cat "$cache_file" 2>/dev/null)
    local now=$(date +%s)
    local age_hours=$(( (now - last_update) / 3600 ))

    if [[ $age_hours -ge $CACHE_MAX_AGE_HOURS ]]; then
        return 0  # Cache is stale
    fi

    return 1  # Cache is fresh
}

# Record update timestamp
record_update() {
    local profile="${1:-default}"
    mkdir -p "$COMMUNITY_RULES_DIR/$profile"
    date +%s > "$COMMUNITY_RULES_DIR/$profile/.last-update"
}

# Download a rule pack from Semgrep registry
download_pack() {
    local pack="$1"
    local output_dir="$2"
    local pack_name="${pack#p/}"  # Remove p/ prefix for filename
    local output_file="$output_dir/${pack_name}.yaml"

    log "Downloading $pack..."

    # Use semgrep to fetch and validate the rules
    # --dump-rule-yaml exports the rules in YAML format
    if semgrep --config "$pack" --validate --metrics=off > /dev/null 2>&1; then
        # Download rules using curl from registry API
        local registry_url="https://semgrep.dev/c/$pack"

        if curl -sL "$registry_url" -o "$output_file.tmp" 2>/dev/null; then
            # Validate the downloaded YAML
            if semgrep --config "$output_file.tmp" --validate --metrics=off 2>/dev/null; then
                mv "$output_file.tmp" "$output_file"
                local rule_count=$(grep -c "^  - id:" "$output_file" 2>/dev/null || echo "?")
                log_success "$pack_name: $rule_count rules"
                return 0
            else
                rm -f "$output_file.tmp"
                log_warn "$pack_name: validation failed, using registry reference"
            fi
        fi
    fi

    # Fallback: just store the pack reference for semgrep to fetch at runtime
    echo "# Semgrep registry reference: $pack" > "$output_file"
    echo "# This file tells semgrep-scanner to use: semgrep --config $pack" >> "$output_file"
    echo "rules: []" >> "$output_file"
    log_warn "$pack_name: using registry reference (will fetch at scan time)"
    return 0
}

#############################################################################
# Commands
#############################################################################

cmd_sync() {
    local profile="${1:-default}"
    local force="${2:-false}"

    check_semgrep
    ensure_dirs

    echo -e "${BOLD}Syncing Semgrep Community Rules${NC}"
    echo -e "${DIM}Profile: $profile${NC}"
    echo

    # Check if update needed
    if [[ "$force" != "true" ]] && ! needs_update "$profile"; then
        local last_update=$(cat "$COMMUNITY_RULES_DIR/$profile/.last-update" 2>/dev/null)
        local age_hours=$(( ($(date +%s) - last_update) / 3600 ))
        log "Rules are up to date (${age_hours}h old, max ${CACHE_MAX_AGE_HOURS}h)"
        log "Use --force to update anyway"
        return 0
    fi

    local packs=()
    local output_dir=""

    case "$profile" in
        security)
            packs=("${SECURITY_PACKS[@]}")
            output_dir="$COMMUNITY_RULES_DIR/security"
            ;;
        quality)
            packs=("${QUALITY_PACKS[@]}")
            output_dir="$COMMUNITY_RULES_DIR/quality"
            ;;
        languages)
            packs=("${LANG_SECURITY_PACKS[@]}")
            output_dir="$COMMUNITY_RULES_DIR/languages"
            ;;
        all)
            # Sync all profiles
            cmd_sync "security" "$force"
            cmd_sync "quality" "$force"
            cmd_sync "languages" "$force"
            return 0
            ;;
        default|*)
            packs=("${DEFAULT_PACKS[@]}")
            output_dir="$COMMUNITY_RULES_DIR/default"
            ;;
    esac

    mkdir -p "$output_dir"

    local success=0
    local failed=0

    for pack in "${packs[@]}"; do
        if download_pack "$pack" "$output_dir"; then
            ((success++))
        else
            ((failed++))
        fi
    done

    record_update "$profile"

    echo
    echo -e "${GREEN}✓${NC} Synced $success rule packs to $output_dir"
    [[ $failed -gt 0 ]] && echo -e "${YELLOW}⚠${NC} $failed packs had issues"

    # Update cache info
    update_cache_info
}

cmd_status() {
    echo -e "${BOLD}Semgrep Community Rules Status${NC}"
    echo

    if [[ ! -d "$COMMUNITY_RULES_DIR" ]]; then
        echo -e "${YELLOW}No community rules downloaded yet.${NC}"
        echo "Run: ./community-rules.sh sync"
        return 0
    fi

    echo -e "Location: ${CYAN}$COMMUNITY_RULES_DIR${NC}"
    echo

    for profile in default security quality languages; do
        local profile_dir="$COMMUNITY_RULES_DIR/$profile"
        if [[ -d "$profile_dir" ]]; then
            local rule_count=$(find "$profile_dir" -name "*.yaml" -type f | wc -l | tr -d ' ')
            local last_update_file="$profile_dir/.last-update"
            local age="never"

            if [[ -f "$last_update_file" ]]; then
                local last_ts=$(cat "$last_update_file")
                local now=$(date +%s)
                local age_hours=$(( (now - last_ts) / 3600 ))
                local age_mins=$(( ((now - last_ts) % 3600) / 60 ))

                if [[ $age_hours -gt 0 ]]; then
                    age="${age_hours}h ${age_mins}m ago"
                else
                    age="${age_mins}m ago"
                fi
            fi

            printf "  %-12s %3d rule files  ${DIM}(updated %s)${NC}\n" "$profile:" "$rule_count" "$age"
        else
            printf "  %-12s ${DIM}not synced${NC}\n" "$profile:"
        fi
    done

    echo
    local total_size=$(du -sh "$COMMUNITY_RULES_DIR" 2>/dev/null | cut -f1)
    echo -e "Total size: ${CYAN}$total_size${NC}"
}

cmd_list() {
    echo -e "${BOLD}Available Rule Packs${NC}"
    echo

    echo -e "${CYAN}Security Packs:${NC}"
    for pack in "${SECURITY_PACKS[@]}"; do
        echo "  $pack"
    done
    echo

    echo -e "${CYAN}Quality Packs:${NC}"
    for pack in "${QUALITY_PACKS[@]}"; do
        echo "  $pack"
    done
    echo

    echo -e "${CYAN}Language Security Packs:${NC}"
    for pack in "${LANG_SECURITY_PACKS[@]}"; do
        echo "  $pack"
    done
    echo

    echo -e "${DIM}See more at: https://semgrep.dev/explore${NC}"
}

cmd_clean() {
    if [[ -d "$COMMUNITY_RULES_DIR" ]]; then
        local size=$(du -sh "$COMMUNITY_RULES_DIR" 2>/dev/null | cut -f1)
        rm -rf "$COMMUNITY_RULES_DIR"
        log_success "Removed $size of cached rules"
    else
        log "No cached rules to remove"
    fi
}

# Get rules path for a profile (used by semgrep-scanner.sh)
cmd_path() {
    local profile="${1:-default}"
    local profile_dir="$COMMUNITY_RULES_DIR/$profile"

    if [[ -d "$profile_dir" ]]; then
        echo "$profile_dir"
    else
        # Return empty - caller should handle missing rules
        return 1
    fi
}

# Update cache info JSON
update_cache_info() {
    local now=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    local profiles=()

    for profile in default security quality languages; do
        local profile_dir="$COMMUNITY_RULES_DIR/$profile"
        if [[ -d "$profile_dir" ]]; then
            local rule_count=$(find "$profile_dir" -name "*.yaml" -type f | wc -l | tr -d ' ')
            local last_update=""
            if [[ -f "$profile_dir/.last-update" ]]; then
                last_update=$(date -r "$(cat "$profile_dir/.last-update")" -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "unknown")
            fi
            profiles+=("\"$profile\": {\"rules\": $rule_count, \"updated\": \"$last_update\"}")
        fi
    done

    cat > "$RULES_CACHE_FILE" << EOF
{
  "version": "1.0.0",
  "generated": "$now",
  "location": "$COMMUNITY_RULES_DIR",
  "profiles": {
    $(IFS=,; echo "${profiles[*]}")
  }
}
EOF
}

#############################################################################
# Main
#############################################################################

usage() {
    cat << EOF
Semgrep Community Rules Manager

Usage: $0 [command] [options]

COMMANDS:
    sync [profile]    Download/update rules (default command)
    status            Show current rules status
    list              List available rule packs
    path [profile]    Print rules directory path
    clean             Remove cached rules

PROFILES:
    default           Basic security rules (OWASP, secrets, audit)
    security          Comprehensive security rules
    quality           Code quality rules
    languages         Language-specific security rules
    all               All profiles

OPTIONS:
    --force           Force update even if cache is fresh
    -h, --help        Show this help

EXAMPLES:
    $0 sync                    # Sync default rules
    $0 sync security           # Sync security-focused rules
    $0 sync all --force        # Force sync all profiles
    $0 status                  # Show rules status
    $0 path security           # Get path for security rules

ENVIRONMENT:
    SEMGREP_COMMUNITY_DIR      Override rules storage location
                               (default: ~/.zero/semgrep-rules)

EOF
    exit 0
}

main() {
    local cmd="${1:-sync}"
    shift || true

    local force=false
    local args=()

    # Parse options
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --force)
                force=true
                shift
                ;;
            -h|--help)
                usage
                ;;
            *)
                args+=("$1")
                shift
                ;;
        esac
    done

    case "$cmd" in
        sync)
            cmd_sync "${args[0]:-default}" "$force"
            ;;
        status)
            cmd_status
            ;;
        list)
            cmd_list
            ;;
        path)
            cmd_path "${args[0]:-default}"
            ;;
        clean)
            cmd_clean
            ;;
        -h|--help)
            usage
            ;;
        *)
            log_error "Unknown command: $cmd"
            usage
            ;;
    esac
}

main "$@"
