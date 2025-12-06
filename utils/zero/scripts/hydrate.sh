#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Phantom Hydrate
# Clone and scan repositories in one command
#
# This is a convenience wrapper that calls clone.sh + scan.sh
#
# Usage:
#   ./hydrate.sh <owner/repo>           # Clone and scan single repo
#   ./hydrate.sh --org <org-name>       # Clone and scan all repos in org
#
# Examples:
#   ./hydrate.sh expressjs/express
#   ./hydrate.sh expressjs/express --quick
#   ./hydrate.sh --org expressjs --standard
#############################################################################

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ZERO_DIR="$(dirname "$SCRIPT_DIR")"

# Load Phantom library
source "$ZERO_DIR/lib/zero-lib.sh"

#############################################################################
# Configuration
#############################################################################

ORG_MODE=false
ORG_NAME=""
TARGET=""
CLONE_ONLY=false

# Arguments to pass through to clone.sh and scan.sh
CLONE_ARGS=()
SCAN_ARGS=()

#############################################################################
# Usage
#############################################################################

usage() {
    cat << EOF
Phantom Hydrate - Clone and scan repositories

Usage: $0 <target> [options]
       $0 --org <org-name> [options]

This command combines clone + scan into a single step.

MODES:
    Single Repo:    $0 owner/repo [options]
    Organization:   $0 --org <org-name> [options]

CLONE OPTIONS:
    --branch <name>     Clone specific branch
    --depth <n>         Shallow clone depth
    --clone-only        Clone without scanning

SCAN OPTIONS:
    --quick             Fast scan (~30s)
    --standard          Standard scan (~2min) [default]
    --advanced          Full scan (~5min)
    --deep              Deep scan with Claude (~10min)
    --security          Security-focused scan (~3min)
    --security-deep     Deep security analysis with Claude (~10min)
    --compliance        License and policy compliance (~2min)
    --devops            CI/CD and operational metrics (~3min)
    --malcontent        Supply chain compromise detection (~2min)

COMMON OPTIONS:
    --org <name>        Process all repos in organization
    --limit <n>         Max repos in org mode
    --force             Re-clone and re-scan
    -h, --help          Show this help

EXAMPLES:
    $0 expressjs/express                    # Clone + standard scan
    $0 expressjs/express --quick            # Clone + quick scan
    $0 --org expressjs --limit 10           # Clone + scan first 10 repos
    $0 expressjs/express --clone-only       # Clone only, no scan

EOF
    exit 0
}

#############################################################################
# Argument Parsing
#############################################################################

parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            -h|--help)
                usage
                ;;
            --org)
                ORG_MODE=true
                ORG_NAME="$2"
                CLONE_ARGS+=("--org" "$2")
                SCAN_ARGS+=("--org" "$2")
                shift 2
                ;;
            --limit)
                CLONE_ARGS+=("--limit" "$2")
                shift 2
                ;;
            --branch)
                CLONE_ARGS+=("--branch" "$2")
                shift 2
                ;;
            --depth)
                CLONE_ARGS+=("--depth" "$2")
                shift 2
                ;;
            --clone-only)
                CLONE_ONLY=true
                shift
                ;;
            --quick|--standard|--advanced|--deep|--security|--security-deep|--compliance|--devops|--malcontent)
                SCAN_ARGS+=("$1")
                shift
                ;;
            --force)
                CLONE_ARGS+=("--force")
                SCAN_ARGS+=("--force")
                shift
                ;;
            -*)
                echo -e "${RED}Error: Unknown option $1${NC}" >&2
                exit 1
                ;;
            *)
                if [[ -z "$TARGET" ]]; then
                    TARGET="$1"
                    CLONE_ARGS+=("$1")
                    SCAN_ARGS+=("$1")
                else
                    echo -e "${RED}Error: Multiple targets specified${NC}" >&2
                    exit 1
                fi
                shift
                ;;
        esac
    done

    # Validate arguments
    if [[ "$ORG_MODE" != "true" ]] && [[ -z "$TARGET" ]]; then
        echo -e "${RED}Error: No target specified${NC}" >&2
        echo "Usage: $0 <owner/repo> or $0 --org <org-name>"
        exit 1
    fi
}

#############################################################################
# Main
#############################################################################

main() {
    parse_args "$@"

    # Step 1: Clone
    echo -e "${BOLD}Step 1: Clone${NC}"
    "$SCRIPT_DIR/clone.sh" "${CLONE_ARGS[@]}"

    # Step 2: Scan (unless clone-only)
    if [[ "$CLONE_ONLY" != "true" ]]; then
        echo
        echo -e "${BOLD}Step 2: Scan${NC}"
        "$SCRIPT_DIR/scan.sh" "${SCAN_ARGS[@]}"
    fi
}

main "$@"
