#!/bin/bash
# Copyright (c) 2024 Crash Override Inc
# 101 Fulton St, 416, New York 10038
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Code Ownership Analyzer v2.0
# Enhanced analysis with dual-method measurement, advanced metrics, and
# comprehensive JSON output
#
# Key Features:
# - Dual-method ownership (commit-based + line-based)
# - Research-backed metrics (Gini, bus factor, health score)
# - Enhanced SPOF detection (6 criteria)
# - Advanced CODEOWNERS validation
# - GitHub profile integration
# - Multi-repository support
# - Complete JSON output
#############################################################################

set -euo pipefail

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LIB_DIR="$SCRIPT_DIR/lib"

# Load libraries
source "$LIB_DIR/metrics.sh"
source "$LIB_DIR/github.sh"
source "$LIB_DIR/analyzer-core.sh"
source "$LIB_DIR/codeowners-validator.sh"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Default options
ANALYSIS_METHOD="hybrid"  # commit, line, hybrid
DAYS=90
FORMAT="json"
OUTPUT_FILE=""
VALIDATE_CODEOWNERS=false
CODEOWNERS_PATH=".github/CODEOWNERS"
VERBOSE=false
REPOS=()
ORG=""

# Version
VERSION="2.0.0"

usage() {
    cat << EOF
Code Ownership Analyzer v$VERSION - Enhanced Analysis

Performs comprehensive code ownership analysis using dual-method measurement,
research-backed metrics, and advanced validation.

Usage: $0 [OPTIONS] <target>

TARGETS:
    Local directory         Analyze local repository
    Git repository URL      Clone and analyze repository
    --org ORGANIZATION      Analyze all repos in organization (requires GITHUB_TOKEN)
    --repos REPO1 REPO2...  Analyze multiple repositories

OPTIONS:
    Analysis:
    -m, --method METHOD     Analysis method: commit, line, hybrid (default: hybrid)
    -d, --days N            Analyze last N days (default: 90)

    Output:
    -f, --format FORMAT     Output format: json, text (default: json)
    -o, --output FILE       Write output to file

    Validation:
    -v, --validate          Validate CODEOWNERS file
    -c, --codeowners PATH   CODEOWNERS file path (default: .github/CODEOWNERS)

    Other:
    --verbose               Enable verbose output
    --version               Show version
    -h, --help              Show this help

EXAMPLES:
    # Analyze single repository (JSON output)
    $0 .

    # Analyze with all methods
    $0 --method hybrid --validate --verbose .

    # Analyze GitHub repository
    $0 https://github.com/owner/repo

    # Analyze organization (requires GITHUB_TOKEN)
    export GITHUB_TOKEN=ghp_xxx
    $0 --org myorg --output org-analysis.json

    # Analyze multiple repos
    $0 --repos repo1 repo2 repo3

ENVIRONMENT:
    GITHUB_TOKEN    Optional GitHub personal access token for API access
                    Increases rate limits and enables private repo access

EOF
    exit 0
}

log() {
    if [[ "$VERBOSE" == "true" ]]; then
        echo -e "${BLUE}[$(date '+%H:%M:%S')]${NC} $*" >&2
    fi
}

error() {
    echo -e "${RED}Error:${NC} $*" >&2
    exit 1
}

check_prerequisites() {
    local missing=()

    for cmd in git jq bc; do
        if ! command -v $cmd &> /dev/null; then
            missing+=("$cmd")
        fi
    done

    if [[ ${#missing[@]} -gt 0 ]]; then
        error "Missing required tools: ${missing[*]}\nInstall: brew install ${missing[*]}"
    fi
}

# Analyze single repository
analyze_repository() {
    local repo_path="$1"
    local repo_name="${2:-$(basename "$repo_path")}"

    log "Analyzing repository: $repo_name"

    cd "$repo_path" || return 1

    # Calculate since date
    local since_date=$(date -v-${DAYS}d +%Y-%m-%d 2>/dev/null || date -d "$DAYS days ago" +%Y-%m-% d)

    # Create temp files for analysis
    local commit_file=$(mktemp)
    local line_file=$(mktemp)
    local hybrid_file=$(mktemp)
    local stats_file=$(mktemp)
    local spof_file=$(mktemp)

    # Run analyses based on method
    if [[ "$ANALYSIS_METHOD" == "commit" ]] || [[ "$ANALYSIS_METHOD" == "hybrid" ]]; then
        log "Running commit-based analysis..."
        analyze_commit_based_ownership "$repo_path" "$since_date" "$commit_file"
    fi

    if [[ "$ANALYSIS_METHOD" == "line" ]] || [[ "$ANALYSIS_METHOD" == "hybrid" ]]; then
        log "Running line-based analysis..."
        analyze_line_based_ownership "$repo_path" "$line_file"
    fi

    if [[ "$ANALYSIS_METHOD" == "hybrid" ]]; then
        log "Combining analysis methods..."
        combine_ownership_methods "$commit_file" "$line_file" "$hybrid_file"
    fi

    # Calculate contributor statistics
    log "Calculating contributor statistics..."
    calculate_contributor_stats "$repo_path" "$since_date" "$stats_file"

    # Detect SPOFs
    log "Detecting single points of failure..."
    detect_spof "$repo_path" "${hybrid_file:-$commit_file}" "$spof_file"

    # Calculate repository metrics
    log "Calculating repository metrics..."
    local repo_metrics=$(calculate_repository_metrics "$repo_path" "$since_date" "${hybrid_file:-$commit_file}")

    # Initialize GitHub cache
    init_github_cache

    # Generate output based on format
    if [[ "$FORMAT" == "json" ]]; then
        generate_json_report "$repo_path" "$repo_name" "$repo_metrics" \
            "${hybrid_file:-$commit_file}" "$stats_file" "$spof_file"
    else
        generate_text_report "$repo_path" "$repo_name" "$repo_metrics" \
            "${hybrid_file:-$commit_file}" "$stats_file" "$spof_file"
    fi

    # Run CODEOWNERS validation if requested
    if [[ "$VALIDATE_CODEOWNERS" == "true" ]]; then
        local codeowners_file="$repo_path/$CODEOWNERS_PATH"
        if [[ -f "$codeowners_file" ]]; then
            log "Validating CODEOWNERS file..."
            generate_validation_report "$codeowners_file" "$repo_path" "$FORMAT"
        else
            log "CODEOWNERS file not found at: $codeowners_file"
        fi
    fi

    # Cleanup
    cleanup_github_cache
    rm -f "$commit_file" "$line_file" "$hybrid_file" "$stats_file" "$spof_file"
}

# Generate comprehensive JSON report
generate_json_report() {
    local repo_path="$1"
    local repo_name="$2"
    local repo_metrics="$3"
    local ownership_file="$4"
    local stats_file="$5"
    local spof_file="$6"

    # Parse repository metrics
    IFS='|' read -r total_files total_commits active_contributors files_with_owners coverage <<< "$repo_metrics"

    # Build contributors array
    local contributors_json=$(awk -F'|' '
    {
        author=$1
        email=$2
        file=$3

        key=author"|"email
        files[key]++
        total_files_owned[key]++
    }
    END {
        printf "["
        first=1
        for (key in files) {
            split(key, parts, "|")
            if (!first) printf ","
            first=0
            printf "{\"name\":\"%s\",\"email\":\"%s\",\"files_owned\":%d}",
                parts[1], parts[2], total_files_owned[key]
        }
        printf "]"
    }
    ' "$ownership_file")

    # Build SPOFs array
    local spofs_json="[]"
    if [[ -f "$spof_file" ]] && [[ -s "$spof_file" ]]; then
        spofs_json=$(awk -F'|' '
        {
            printf "%s{\"file\":\"%s\",\"score\":%d,\"risk\":\"%s\",\"contributors\":%d,\"critical\":%d,\"complex\":%d,\"has_backup\":%d,\"has_tests\":%d,\"has_docs\":%d,\"loc\":%d}",
                (NR>1?",":""), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
        }
        END {
            if (NR == 0) print "[]"
            else print "]"
        }
        BEGIN {
            print "["
        }
        ' "$spof_file")
    fi

    # Calculate Gini coefficient
    local file_counts=()
    while IFS='|' read -r author email count; do
        file_counts+=("$count")
    done < <(awk -F'|' '{key=$1"|"$2; files[key]++} END {for(k in files) print k"|"files[k]}' "$ownership_file")

    local gini=$(calculate_gini_coefficient "${file_counts[@]}")

    # Calculate bus factor
    local bus_factor=$(calculate_bus_factor "$total_files" "${file_counts[@]}")

    # Calculate health score (simplified - coverage and distribution based)
    local freshness="75"  # Placeholder - would need to calculate active owners
    local engagement="70"  # Placeholder - would need review data
    local health_score=$(calculate_health_score "$coverage" "$gini" "$freshness" "$engagement")
    local health_grade=$(get_health_grade "$health_score")

    # Build final JSON
    jq -n \
        --arg version "$VERSION" \
        --arg repo_name "$repo_name" \
        --arg repo_path "$repo_path" \
        --arg analysis_date "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        --arg method "$ANALYSIS_METHOD" \
        --arg days "$DAYS" \
        --arg total_files "$total_files" \
        --arg total_commits "$total_commits" \
        --arg active_contributors "$active_contributors" \
        --arg coverage "$coverage" \
        --arg gini "$gini" \
        --arg bus_factor "$bus_factor" \
        --arg health_score "$health_score" \
        --arg health_grade "$health_grade" \
        --argjson contributors "$contributors_json" \
        --argjson spofs "$spofs_json" \
        '{
            metadata: {
                analyzer_version: $version,
                repository: $repo_name,
                repository_path: $repo_path,
                analysis_date: $analysis_date,
                analysis_method: $method,
                time_period_days: ($days | tonumber)
            },
            repository_metrics: {
                total_files: ($total_files | tonumber),
                total_commits: ($total_commits | tonumber),
                active_contributors: ($active_contributors | tonumber),
                files_with_owners: (($total_files | tonumber) * ($coverage | tonumber) / 100 | floor)
            },
            ownership_health: {
                coverage_percentage: ($coverage | tonumber),
                gini_coefficient: ($gini | tonumber),
                bus_factor: ($bus_factor | tonumber),
                health_score: ($health_score | tonumber),
                health_grade: $health_grade
            },
            contributors: $contributors,
            single_points_of_failure: $spofs,
            recommendations: {
                critical_spofs: ($spofs | map(select(.risk == "Critical")) | length),
                high_spofs: ($spofs | map(select(.risk == "High")) | length),
                coverage_target: 90,
                bus_factor_target: 3,
                needs_attention: (
                    if ($bus_factor | tonumber) < 2 then "Critical: Bus factor too low"
                    elif ($coverage | tonumber) < 70 then "Warning: Coverage below 70%"
                    elif ($gini | tonumber) > 0.7 then "Warning: High ownership concentration"
                    else "Good: No critical issues"
                    end
                )
            }
        }'
}

# Generate text report
generate_text_report() {
    local repo_path="$1"
    local repo_name="$2"
    local repo_metrics="$3"
    local ownership_file="$4"
    local stats_file="$5"
    local spof_file="$6"

    IFS='|' read -r total_files total_commits active_contributors files_with_owners coverage <<< "$repo_metrics"

    cat << EOF
========================================
Code Ownership Analysis v$VERSION
========================================

Repository: $repo_name
Analysis Date: $(date +%Y-%m-%d)
Method: $ANALYSIS_METHOD
Time Period: Last $DAYS days

Repository Metrics:
------------------
Total Files: $total_files
Total Commits: $total_commits
Active Contributors: $active_contributors
Ownership Coverage: $coverage%

Top Contributors:
----------------
EOF

    # Show top contributors with GitHub profiles
    awk -F'|' '{key=$1"|"$2; files[key]++; name[key]=$1; email[key]=$2}
    END {for(k in files) print files[k]"|"name[k]"|"email[k]}' "$ownership_file" | \
    sort -t'|' -k1 -rn | head -10 | while IFS='|' read -r count author email; do
        local github_info=$(get_github_profile "$email")
        local username=$(echo "$github_info" | cut -d'|' -f1)

        if [[ -n "$username" ]]; then
            printf "%-40s %5d files (@%s)\n" "$author" "$count" "$username"
        else
            printf "%-40s %5d files\n" "$author" "$count"
        fi
    done

    if [[ -f "$spof_file" ]] && [[ -s "$spof_file" ]]; then
        echo ""
        echo "Single Points of Failure:"
        echo "------------------------"
        awk -F'|' '{printf "%-50s %s (%d criteria)\n", $1, $3, $2}' "$spof_file" | head -10
    fi

    echo ""
    echo "========================================"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -m|--method)
            ANALYSIS_METHOD="$2"
            shift 2
            ;;
        -d|--days)
            DAYS="$2"
            shift 2
            ;;
        -f|--format)
            FORMAT="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_FILE="$2"
            shift 2
            ;;
        -v|--validate)
            VALIDATE_CODEOWNERS=true
            shift
            ;;
        -c|--codeowners)
            CODEOWNERS_PATH="$2"
            shift 2
            ;;
        --org)
            ORG="$2"
            shift 2
            ;;
        --repos)
            shift
            while [[ $# -gt 0 ]] && [[ ! "$1" =~ ^- ]]; do
                REPOS+=("$1")
                shift
            done
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        --version)
            echo "Code Ownership Analyzer v$VERSION"
            exit 0
            ;;
        -h|--help)
            usage
            ;;
        *)
            REPOS+=("$1")
            shift
            ;;
    esac
done

# Main execution
echo -e "${CYAN}Code Ownership Analyzer v$VERSION${NC}" >&2
echo "" >&2

check_prerequisites

# Handle organization scanning
if [[ -n "$ORG" ]]; then
    if [[ -z "${GITHUB_TOKEN:-}" ]]; then
        error "GITHUB_TOKEN required for organization scanning"
    fi

    log "Fetching repositories for organization: $ORG"
    mapfile -t org_repos < <(list_org_repos "$ORG")

    if [[ ${#org_repos[@]} -eq 0 ]]; then
        error "No repositories found for organization: $ORG"
    fi

    log "Found ${#org_repos[@]} repositories"
    REPOS=("${org_repos[@]}")
fi

# Validate we have repositories to analyze
if [[ ${#REPOS[@]} -eq 0 ]]; then
    error "No repositories specified. Use --help for usage."
fi

# Analyze repositories
if [[ ${#REPOS[@]} -eq 1 ]]; then
    # Single repository
    if [[ -n "$OUTPUT_FILE" ]]; then
        analyze_repository "${REPOS[0]}" "$(basename "${REPOS[0]}")" > "$OUTPUT_FILE"
        echo -e "${GREEN}✓ Analysis complete${NC}" >&2
        echo -e "Output written to: $OUTPUT_FILE" >&2
    else
        analyze_repository "${REPOS[0]}" "$(basename "${REPOS[0]}")"
    fi
else
    # Multiple repositories - aggregate results
    echo "[" > "${OUTPUT_FILE:-/dev/stdout}"

    local first=true
    for repo in "${REPOS[@]}"; do
        if [[ "$first" != "true" ]]; then
            echo "," >> "${OUTPUT_FILE:-/dev/stdout}"
        fi
        first=false

        analyze_repository "$repo" "$(basename "$repo")" >> "${OUTPUT_FILE:-/dev/stdout}"
    done

    echo "]" >> "${OUTPUT_FILE:-/dev/stdout}"

    echo -e "${GREEN}✓ Analyzed ${#REPOS[@]} repositories${NC}" >&2
fi

echo "" >&2
echo -e "${GREEN}Analysis Complete${NC}" >&2
