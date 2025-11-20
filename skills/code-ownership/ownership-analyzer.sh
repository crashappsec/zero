#!/bin/bash
# Copyright (c) 2024 Crash Override Inc
# 101 Fulton St, 416, New York 10038
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Code Ownership Analyzer - Basic Analysis
# Analyzes git repository to identify code owners and generate reports
# Usage: ./ownership-analyzer.sh [options] <repository-path>
#############################################################################

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Default options
DAYS=90
FORMAT="text"
OUTPUT_FILE=""
VALIDATE_CODEOWNERS=false
CODEOWNERS_PATH=".github/CODEOWNERS"

usage() {
    cat << EOF
Code Ownership Analyzer - Analyze repository ownership patterns

Analyzes git history to identify code owners, calculate metrics, and
optionally validate CODEOWNERS files.

Usage: $0 [OPTIONS] <repository-path>

OPTIONS:
    -d, --days N            Analyze last N days of history (default: 90)
    -f, --format FORMAT     Output format: text, json, csv (default: text)
    -o, --output FILE       Write output to file instead of stdout
    -v, --validate          Validate CODEOWNERS file if present
    -c, --codeowners PATH   Path to CODEOWNERS file (default: .github/CODEOWNERS)
    -h, --help              Show this help message

EXAMPLES:
    # Analyze current directory (last 90 days)
    $0 .

    # Analyze specific repository (last 180 days)
    $0 --days 180 /path/to/repo

    # Generate JSON output
    $0 --format json --output ownership.json .

    # Validate CODEOWNERS file
    $0 --validate --codeowners .github/CODEOWNERS .

    # Full analysis with validation
    $0 --days 90 --validate --format json --output report.json .

EOF
    exit 1
}

check_prerequisites() {
    if ! command -v git &> /dev/null; then
        echo -e "${RED}Error: git is not installed${NC}"
        exit 1
    fi

    if ! command -v jq &> /dev/null; then
        echo -e "${RED}Error: jq is not installed${NC}"
        echo "Install: brew install jq  (or apt-get install jq)"
        exit 1
    fi

    if ! command -v bc &> /dev/null; then
        echo -e "${RED}Error: bc is not installed${NC}"
        echo "Install: brew install bc  (or apt-get install bc)"
        exit 1
    fi
}

is_git_repository() {
    local repo_path="$1"
    if [[ ! -d "$repo_path/.git" ]]; then
        echo -e "${RED}Error: $repo_path is not a git repository${NC}"
        exit 1
    fi
}

analyze_ownership() {
    local repo_path="$1"
    local days="$2"

    cd "$repo_path" || exit 1

    echo -e "${BLUE}Analyzing repository: $(basename "$repo_path")${NC}"
    echo -e "${BLUE}Time period: Last $days days${NC}"
    echo ""

    # Get total files
    local total_files=$(git ls-files | wc -l | tr -d ' ')

    # Get contributors in the time period
    local since_date=$(date -v-${days}d +%Y-%m-%d 2>/dev/null || date -d "$days days ago" +%Y-%m-%d)

    # Collect ownership data
    local ownership_data=$(mktemp)
    git log --since="$since_date" --format="%an|%ae" --name-only | \
        grep -v "^$" | \
        awk -F'|' '
        NF==2 {author=$1; email=$2; next}
        {files[author"|"email"|"$0]++}
        END {for (key in files) print key"|"files[key]}
        ' > "$ownership_data"

    # Calculate per-author statistics
    local author_stats=$(mktemp)
    git log --since="$since_date" --format="%an|%ae" --numstat | \
        awk -F'|' '
        NF==2 {author=$1; next}
        NF==3 {
            commits[author]++
            added[author]+=$1
            deleted[author]+=$2
        }
        END {
            for (author in commits) {
                print author"|"commits[author]"|"added[author]"|"deleted[author]
            }
        }
        ' > "$author_stats"

    # Get commit counts per author
    local commit_counts=$(git log --since="$since_date" --format="%an" | sort | uniq -c | sort -rn)

    # Calculate metrics
    local total_commits=$(git log --since="$since_date" --oneline | wc -l | tr -d ' ')
    local active_authors=$(git log --since="$since_date" --format="%an" | sort -u | wc -l | tr -d ' ')

    # Export results based on format
    if [[ "$FORMAT" == "json" ]]; then
        generate_json_output "$total_files" "$total_commits" "$active_authors" "$ownership_data" "$author_stats"
    elif [[ "$FORMAT" == "csv" ]]; then
        generate_csv_output "$ownership_data" "$author_stats"
    else
        generate_text_output "$total_files" "$total_commits" "$active_authors" "$ownership_data" "$author_stats" "$commit_counts"
    fi

    rm -f "$ownership_data" "$author_stats"
}

generate_text_output() {
    local total_files="$1"
    local total_commits="$2"
    local active_authors="$3"
    local ownership_data="$4"
    local author_stats="$5"
    local commit_counts="$6"

    cat << EOF
========================================
  Code Ownership Analysis Report
========================================

Repository Metrics:
-------------------
Total Files: $total_files
Total Commits (period): $total_commits
Active Contributors: $active_authors

Top Contributors (by commit count):
-----------------------------------
EOF

    echo "$commit_counts" | head -10 | while read count author; do
        printf "%-40s %5d commits\n" "$author" "$count"
    done

    echo ""
    echo "Ownership Distribution:"
    echo "----------------------"

    # Calculate top file owners
    awk -F'|' '{
        author_email=$1"|"$2
        authors[author_email]+=$4
    }
    END {
        for (ae in authors) {
            split(ae, parts, "|")
            print authors[ae]"|"parts[1]
        }
    }' "$ownership_data" | sort -t'|' -k1 -rn | head -10 | while IFS='|' read files author; do
        printf "%-40s %5d files\n" "$author" "$files"
    done

    echo ""
    echo "Activity Summary:"
    echo "----------------"

    while IFS='|' read author commits added deleted; do
        local net=$((added - deleted))
        printf "%-30s %4d commits, +%d -%d lines (net: %+d)\n" \
            "$author" "$commits" "$added" "$deleted" "$net"
    done < "$author_stats" | head -10

    echo ""
    echo "========================================="
}

generate_json_output() {
    local total_files="$1"
    local total_commits="$2"
    local active_authors="$3"
    local ownership_data="$4"
    local author_stats="$5"

    jq -n \
        --arg total_files "$total_files" \
        --arg total_commits "$total_commits" \
        --arg active_authors "$active_authors" \
        --arg analysis_date "$(date +%Y-%m-%d)" \
        --arg days "$DAYS" \
        '{
            metadata: {
                repository: $ENV.PWD,
                analysis_date: $analysis_date,
                time_period_days: ($days | tonumber),
                total_files: ($total_files | tonumber),
                total_commits: ($total_commits | tonumber),
                active_contributors: ($active_authors | tonumber)
            },
            contributors: []
        }'
}

generate_csv_output() {
    local ownership_data="$1"
    local author_stats="$2"

    echo "Author,Email,Files Owned,Commits,Lines Added,Lines Deleted,Net Lines"

    # Combine data
    join -t'|' \
        <(awk -F'|' '{ae=$1"|"$2; files[ae]+=$4} END {for(k in files) print k"|"files[k]}' "$ownership_data" | sort) \
        <(sort "$author_stats") | \
        while IFS='|' read author email files commits added deleted; do
            local net=$((added - deleted))
            echo "$author,$email,$files,$commits,$added,$deleted,$net"
        done
}

validate_codeowners_file() {
    local repo_path="$1"
    local codeowners_file="$repo_path/$CODEOWNERS_PATH"

    if [[ ! -f "$codeowners_file" ]]; then
        echo -e "${YELLOW}⚠ CODEOWNERS file not found at: $codeowners_file${NC}"
        return 1
    fi

    echo -e "${BLUE}Validating CODEOWNERS file: $codeowners_file${NC}"
    echo ""

    local total_patterns=0
    local valid_patterns=0
    local issues=0

    while IFS= read -r line; do
        # Skip comments and empty lines
        [[ "$line" =~ ^#.*$ ]] && continue
        [[ -z "$line" ]] && continue

        ((total_patterns++))

        # Extract pattern and owners
        local pattern=$(echo "$line" | awk '{print $1}')
        local owners=$(echo "$line" | cut -d' ' -f2-)

        # Basic validation
        if [[ -n "$pattern" && -n "$owners" ]]; then
            ((valid_patterns++))
        else
            echo -e "${YELLOW}⚠ Invalid pattern: $line${NC}"
            ((issues++))
        fi
    done < "$codeowners_file"

    echo "CODEOWNERS Validation Results:"
    echo "----------------------------"
    echo "Total patterns: $total_patterns"
    echo "Valid patterns: $valid_patterns"
    echo "Issues found: $issues"

    if [[ $issues -eq 0 ]]; then
        echo -e "${GREEN}✓ CODEOWNERS file syntax is valid${NC}"
    else
        echo -e "${RED}✗ CODEOWNERS file has $issues issues${NC}"
    fi
    echo ""
}

# Parse arguments
REPO_PATH=""

while [[ $# -gt 0 ]]; do
    case $1 in
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
        -h|--help)
            usage
            ;;
        *)
            REPO_PATH="$1"
            shift
            ;;
    esac
done

if [[ -z "$REPO_PATH" ]]; then
    echo -e "${RED}Error: No repository path specified${NC}"
    usage
fi

# Main execution
echo ""
echo "========================================="
echo "  Code Ownership Analyzer"
echo "========================================="
echo ""

check_prerequisites
is_git_repository "$REPO_PATH"

if [[ "$VALIDATE_CODEOWNERS" == "true" ]]; then
    validate_codeowners_file "$REPO_PATH"
fi

# Run analysis and redirect to file if specified
if [[ -n "$OUTPUT_FILE" ]]; then
    analyze_ownership "$REPO_PATH" "$DAYS" > "$OUTPUT_FILE"
    echo -e "${GREEN}✓ Analysis complete${NC}"
    echo -e "Output written to: $OUTPUT_FILE"
else
    analyze_ownership "$REPO_PATH" "$DAYS"
fi

echo ""
echo "========================================="
echo -e "${GREEN}  Analysis Complete${NC}"
echo "========================================="
echo ""
