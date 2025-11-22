#!/bin/bash
# Copyright (c) 2024 Gibson Powers Contributors
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Skills Sync Script
# Automatically syncs skills to Claude Code without using the UI
#
# This script provides multiple sync methods:
# 1. Symlink (recommended for development)
# 2. Copy (one-time sync)
# 3. Watch mode (continuous sync on file changes)
#############################################################################

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SOURCE_SKILLS_DIR="$SCRIPT_DIR/skills"

# Detect Claude Code skills directory
detect_claude_skills_dir() {
    # Try common locations
    local candidates=(
        "$HOME/.config/claude-code/skills"
        "$HOME/Library/Application Support/Claude/skills"
        "$HOME/.claude/skills"
    )

    for dir in "${candidates[@]}"; do
        if [[ -d "$(dirname "$dir")" ]]; then
            echo "$dir"
            return 0
        fi
    done

    echo ""
    return 1
}

CLAUDE_SKILLS_DIR=$(detect_claude_skills_dir)

usage() {
    cat << EOF
Skills Sync Script - Automatically sync skills to Claude Code

Usage: $0 [MODE] [OPTIONS]

MODES:
    symlink     Create symlinks (recommended for development)
    copy        One-time copy of skills
    watch       Watch for changes and auto-sync
    status      Show current sync status
    clean       Remove all synced skills

OPTIONS:
    -d, --dir PATH      Override Claude Code skills directory
    -h, --help          Show this help

EXAMPLES:
    # Setup symlinks (recommended)
    $0 symlink

    # One-time copy
    $0 copy

    # Watch mode (continuous sync)
    $0 watch

    # Check status
    $0 status

NOTES:
    - Symlink mode is recommended for active development
    - Copy mode is better for stable releases
    - Watch mode requires fswatch (install: brew install fswatch)

EOF
    exit 0
}

log_info() {
    echo -e "${BLUE}ℹ${NC} $*"
}

log_success() {
    echo -e "${GREEN}✓${NC} $*"
}

log_warning() {
    echo -e "${YELLOW}⚠${NC} $*"
}

log_error() {
    echo -e "${RED}✗${NC} $*"
}

# Create symlinks for all skills
sync_symlink() {
    log_info "Creating symlinks from $SOURCE_SKILLS_DIR to $CLAUDE_SKILLS_DIR"

    # Create Claude skills directory if it doesn't exist
    mkdir -p "$CLAUDE_SKILLS_DIR"

    local skills_synced=0

    # Find all skill directories
    while IFS= read -r -d '' skill_dir; do
        local skill_name=$(basename "$skill_dir")

        # Skip if not a directory
        [[ ! -d "$skill_dir" ]] && continue

        local target="$CLAUDE_SKILLS_DIR/$skill_name"

        # Remove existing link or directory
        if [[ -L "$target" ]]; then
            rm "$target"
            log_info "Removed existing symlink: $skill_name"
        elif [[ -d "$target" ]]; then
            log_warning "Directory exists (not a symlink): $skill_name"
            read -p "Replace with symlink? (y/n) " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                rm -rf "$target"
            else
                continue
            fi
        fi

        # Create symlink
        ln -s "$skill_dir" "$target"
        log_success "Linked: $skill_name"
        ((skills_synced++))
    done < <(find "$SOURCE_SKILLS_DIR" -maxdepth 1 -type d ! -name "$(basename "$SOURCE_SKILLS_DIR")" -print0)

    log_success "Synced $skills_synced skills via symlink"
    log_info "Changes in $SOURCE_SKILLS_DIR will be immediately available in Claude Code"
}

# Copy skills (one-time sync)
sync_copy() {
    log_info "Copying skills from $SOURCE_SKILLS_DIR to $CLAUDE_SKILLS_DIR"

    # Create Claude skills directory if it doesn't exist
    mkdir -p "$CLAUDE_SKILLS_DIR"

    local skills_synced=0

    # Copy each skill directory
    while IFS= read -r -d '' skill_dir; do
        local skill_name=$(basename "$skill_dir")

        # Skip if not a directory
        [[ ! -d "$skill_dir" ]] && continue

        local target="$CLAUDE_SKILLS_DIR/$skill_name"

        # Copy with rsync (preserves permissions, excludes temp files)
        if rsync -a --delete \
            --exclude=".DS_Store" \
            --exclude="*.tmp" \
            --exclude=".git" \
            "$skill_dir/" "$target/"; then
            log_success "Copied: $skill_name"
            ((skills_synced++))
        else
            log_error "Failed to copy: $skill_name"
        fi
    done < <(find "$SOURCE_SKILLS_DIR" -maxdepth 1 -type d ! -name "$(basename "$SOURCE_SKILLS_DIR")" -print0)

    log_success "Synced $skills_synced skills via copy"
    log_warning "Run '$0 copy' again to update after making changes"
}

# Watch mode - continuous sync
sync_watch() {
    # Check if fswatch is installed
    if ! command -v fswatch &> /dev/null; then
        log_error "fswatch is required for watch mode"
        log_info "Install: brew install fswatch"
        exit 1
    fi

    log_info "Starting watch mode on $SOURCE_SKILLS_DIR"
    log_info "Press Ctrl+C to stop"
    log_info ""

    # Initial sync
    sync_copy

    # Watch for changes
    fswatch -r "$SOURCE_SKILLS_DIR" | while read -r changed_file; do
        local skill_name=$(echo "$changed_file" | sed "s|$SOURCE_SKILLS_DIR/||" | cut -d'/' -f1)
        local skill_dir="$SOURCE_SKILLS_DIR/$skill_name"

        if [[ -d "$skill_dir" ]]; then
            log_info "Change detected in: $skill_name"

            local target="$CLAUDE_SKILLS_DIR/$skill_name"
            if rsync -a --delete \
                --exclude=".DS_Store" \
                --exclude="*.tmp" \
                --exclude=".git" \
                "$skill_dir/" "$target/"; then
                log_success "Synced: $skill_name"
            fi
        fi
    done
}

# Show sync status
show_status() {
    log_info "Sync Status"
    echo ""
    echo "Source: $SOURCE_SKILLS_DIR"
    echo "Target: $CLAUDE_SKILLS_DIR"
    echo ""

    if [[ ! -d "$CLAUDE_SKILLS_DIR" ]]; then
        log_warning "Claude Code skills directory not found"
        log_info "Run '$0 symlink' or '$0 copy' to sync"
        return
    fi

    local total_skills=0
    local synced_skills=0
    local symlinked_skills=0

    # Check each skill
    while IFS= read -r -d '' skill_dir; do
        local skill_name=$(basename "$skill_dir")
        ((total_skills++))

        local target="$CLAUDE_SKILLS_DIR/$skill_name"

        if [[ -L "$target" ]]; then
            local link_target=$(readlink "$target")
            if [[ "$link_target" == "$skill_dir" ]]; then
                echo -e "${GREEN}→${NC} $skill_name (symlink)"
                ((synced_skills++))
                ((symlinked_skills++))
            else
                echo -e "${YELLOW}⚠${NC} $skill_name (symlink to different location)"
                ((synced_skills++))
            fi
        elif [[ -d "$target" ]]; then
            echo -e "${BLUE}→${NC} $skill_name (copied)"
            ((synced_skills++))
        else
            echo -e "${RED}✗${NC} $skill_name (not synced)"
        fi
    done < <(find "$SOURCE_SKILLS_DIR" -maxdepth 1 -type d ! -name "$(basename "$SOURCE_SKILLS_DIR")" -print0)

    echo ""
    log_info "Total skills: $total_skills"
    log_info "Synced: $synced_skills ($symlinked_skills via symlink)"
}

# Clean synced skills
clean_skills() {
    log_warning "This will remove all synced skills from Claude Code"
    read -p "Are you sure? (y/n) " -n 1 -r
    echo

    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Cancelled"
        exit 0
    fi

    local removed=0

    while IFS= read -r -d '' skill_dir; do
        local skill_name=$(basename "$skill_dir")
        local target="$CLAUDE_SKILLS_DIR/$skill_name"

        if [[ -L "$target" ]] || [[ -d "$target" ]]; then
            rm -rf "$target"
            log_success "Removed: $skill_name"
            ((removed++))
        fi
    done < <(find "$SOURCE_SKILLS_DIR" -maxdepth 1 -type d ! -name "$(basename "$SOURCE_SKILLS_DIR")" -print0)

    log_success "Removed $removed skills"
}

# Main
main() {
    # Parse options
    while [[ $# -gt 0 ]]; do
        case $1 in
            symlink)
                MODE="symlink"
                shift
                ;;
            copy)
                MODE="copy"
                shift
                ;;
            watch)
                MODE="watch"
                shift
                ;;
            status)
                MODE="status"
                shift
                ;;
            clean)
                MODE="clean"
                shift
                ;;
            -d|--dir)
                CLAUDE_SKILLS_DIR="$2"
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

    # Validate directories
    if [[ ! -d "$SOURCE_SKILLS_DIR" ]]; then
        log_error "Skills directory not found: $SOURCE_SKILLS_DIR"
        exit 1
    fi

    if [[ -z "$CLAUDE_SKILLS_DIR" ]]; then
        log_error "Could not detect Claude Code skills directory"
        log_info "Specify manually with: $0 -d /path/to/claude/skills <mode>"
        exit 1
    fi

    # Execute mode
    case "${MODE:-}" in
        symlink)
            sync_symlink
            ;;
        copy)
            sync_copy
            ;;
        watch)
            sync_watch
            ;;
        status)
            show_status
            ;;
        clean)
            clean_skills
            ;;
        *)
            usage
            ;;
    esac
}

main "$@"
