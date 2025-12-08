#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Technical Debt - Data Collector
# Scans for TODO/FIXME markers, deprecated code, and debt indicators
# Outputs JSON for agent analysis - NO AI calls
#
# Usage: ./tech-debt-data.sh [options] <target>
# Output: JSON with debt markers, outdated dependencies, and metrics
#############################################################################

set -e

# Colors for terminal output (stderr only)
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SCANNERS_ROOT="$(dirname "$SCRIPT_DIR")"
UTILS_ROOT="$(dirname "$SCANNERS_ROOT")"
REPO_ROOT="$(dirname "$UTILS_ROOT")"

# Default options
OUTPUT_FILE=""
LOCAL_PATH=""
REPO=""
ORG=""
TEMP_DIR=""
CLEANUP=true
TARGET=""
MAX_FILES=1000
EXCLUDE_PATTERNS="node_modules/**,vendor/**,.git/**,*.min.js,*.bundle.js,dist/**,build/**,__pycache__/**,*.pyc"

# Thresholds
LONG_FILE_LINES=500
LONG_FUNCTION_LINES=100

usage() {
    cat << EOF
Technical Debt - Data Collector (JSON output for agent analysis)

Usage: $0 [OPTIONS] <target>

TARGET:
    Git repository URL      Clone and analyze repository
    Local directory path    Analyze local repository

OPTIONS:
    --local-path PATH       Use pre-cloned repository (skips cloning)
    --repo OWNER/REPO       GitHub repository (looks in zero cache)
    --org ORG               GitHub org (uses first repo found in zero cache)
    --max-files N           Maximum files to scan (default: 1000)
    --exclude PATTERNS      Glob patterns to exclude (comma-separated)
    -o, --output FILE       Write JSON to file (default: stdout)
    -k, --keep-clone        Keep cloned repository
    -h, --help              Show this help

OUTPUT:
    JSON object with:
    - summary: debt score, marker counts
    - markers: TODO/FIXME/HACK/XXX annotations with context
    - deprecated: @deprecated annotations
    - long_files: files exceeding line threshold
    - duplication: duplicate code detection (if jscpd available)

EXAMPLES:
    $0 https://github.com/expressjs/express
    $0 --local-path ~/.zero/projects/foo/repo
    $0 -o tech-debt.json /path/to/project

EOF
    exit 0
}

# Clone repository
clone_repository() {
    local repo_url="$1"
    TEMP_DIR=$(mktemp -d)

    echo -e "${BLUE}Cloning repository...${NC}" >&2
    if git clone --depth 1 "$repo_url" "$TEMP_DIR" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Cloned${NC}" >&2
        return 0
    else
        echo '{"error": "Failed to clone repository"}'
        exit 1
    fi
}

# Cleanup
cleanup() {
    if [[ "$CLEANUP" == true ]] && [[ -n "$TEMP_DIR" ]] && [[ -d "$TEMP_DIR" ]]; then
        rm -rf "$TEMP_DIR"
    fi
}
trap cleanup EXIT

# Detect if target is a Git URL
is_git_url() {
    [[ "$1" =~ ^(https?|git)://.*\.git$ ]] || [[ "$1" =~ ^git@.*:.*\.git$ ]] || [[ "$1" =~ github\.com|gitlab\.com|bitbucket\.org ]]
}

# Get age of a line in days (using git blame)
get_line_age() {
    local repo_dir="$1"
    local file="$2"
    local line="$3"

    # Try to get commit date for the line
    if [[ -d "$repo_dir/.git" ]]; then
        local commit_date=$(cd "$repo_dir" && git blame -L "$line,$line" --porcelain "$file" 2>/dev/null | grep "^author-time" | awk '{print $2}')
        if [[ -n "$commit_date" ]]; then
            local now=$(date +%s)
            local age_seconds=$((now - commit_date))
            echo $((age_seconds / 86400))
            return
        fi
    fi
    echo "null"
}

# Scan for debt markers (TODO, FIXME, HACK, XXX)
scan_debt_markers() {
    local repo_dir="$1"
    local markers="[]"

    # Search patterns
    local patterns="TODO|FIXME|HACK|XXX|BUG|KLUDGE|OPTIMIZE|REFACTOR"

    # Find all source files
    local source_files=$(find "$repo_dir" -type f \( \
        -name "*.py" -o -name "*.js" -o -name "*.ts" -o -name "*.jsx" -o -name "*.tsx" \
        -o -name "*.java" -o -name "*.go" -o -name "*.rb" -o -name "*.php" \
        -o -name "*.c" -o -name "*.cpp" -o -name "*.h" -o -name "*.hpp" \
        -o -name "*.cs" -o -name "*.swift" -o -name "*.kt" -o -name "*.rs" \
        -o -name "*.scala" -o -name "*.sh" -o -name "*.bash" \
    \) ! -path "*node_modules*" ! -path "*vendor*" ! -path "*.git*" ! -path "*dist*" ! -path "*build*" 2>/dev/null)

    local todo_count=0
    local fixme_count=0
    local hack_count=0
    local xxx_count=0
    local other_count=0

    while IFS= read -r file; do
        [[ -z "$file" ]] && continue

        # Search for markers with context
        while IFS=: read -r line_num content; do
            [[ -z "$line_num" ]] && continue

            local rel_path="${file#$repo_dir/}"
            local marker_type=""

            # Determine marker type
            if echo "$content" | grep -qi "TODO"; then
                marker_type="TODO"
                ((todo_count++))
            elif echo "$content" | grep -qi "FIXME"; then
                marker_type="FIXME"
                ((fixme_count++))
            elif echo "$content" | grep -qi "HACK"; then
                marker_type="HACK"
                ((hack_count++))
            elif echo "$content" | grep -qi "XXX"; then
                marker_type="XXX"
                ((xxx_count++))
            else
                marker_type="OTHER"
                ((other_count++))
            fi

            # Clean up the content (remove leading comment chars)
            local clean_text=$(echo "$content" | sed 's/^[[:space:]]*[#/*-]*[[:space:]]*//' | sed 's/^\/\///')

            # Truncate long content
            if [[ ${#clean_text} -gt 200 ]]; then
                clean_text="${clean_text:0:197}..."
            fi

            markers=$(echo "$markers" | jq \
                --arg type "$marker_type" \
                --arg file "$rel_path" \
                --argjson line "$line_num" \
                --arg text "$clean_text" \
                '. + [{
                    "type": $type,
                    "file": $file,
                    "line": $line,
                    "text": $text
                }]')

        done < <(grep -n -E "($patterns)" "$file" 2>/dev/null | head -100)

    done <<< "$source_files"

    # Return markers and counts
    echo "$markers"
    echo "$todo_count" >&3
    echo "$fixme_count" >&4
    echo "$hack_count" >&5
    echo "$xxx_count" >&6
}

# Scan for @deprecated annotations
scan_deprecated() {
    local repo_dir="$1"
    local deprecated="[]"

    # Find files with @deprecated
    local files=$(grep -rl "@deprecated\|@Deprecated\|DEPRECATED" "$repo_dir" \
        --include="*.py" --include="*.js" --include="*.ts" --include="*.java" \
        --include="*.go" --include="*.rb" --include="*.php" \
        ! --include="*node_modules*" ! --include="*vendor*" 2>/dev/null | head -50)

    while IFS= read -r file; do
        [[ -z "$file" ]] && continue
        local rel_path="${file#$repo_dir/}"

        # Get line numbers and context
        while IFS=: read -r line_num content; do
            [[ -z "$line_num" ]] && continue

            deprecated=$(echo "$deprecated" | jq \
                --arg file "$rel_path" \
                --argjson line "$line_num" \
                --arg context "$content" \
                '. + [{
                    "file": $file,
                    "line": $line,
                    "context": $context
                }]')
        done < <(grep -n -i "@deprecated\|DEPRECATED" "$file" 2>/dev/null | head -20)

    done <<< "$files"

    echo "$deprecated"
}

# Find long files
find_long_files() {
    local repo_dir="$1"
    local threshold="$2"
    local long_files="[]"

    local source_files=$(find "$repo_dir" -type f \( \
        -name "*.py" -o -name "*.js" -o -name "*.ts" -o -name "*.jsx" -o -name "*.tsx" \
        -o -name "*.java" -o -name "*.go" -o -name "*.rb" -o -name "*.php" \
        -o -name "*.c" -o -name "*.cpp" -o -name "*.h" -o -name "*.hpp" \
        -o -name "*.cs" -o -name "*.swift" -o -name "*.kt" -o -name "*.rs" \
    \) ! -path "*node_modules*" ! -path "*vendor*" ! -path "*.git*" ! -path "*dist*" ! -path "*build*" 2>/dev/null)

    while IFS= read -r file; do
        [[ -z "$file" ]] && continue

        local line_count=$(wc -l < "$file" 2>/dev/null | tr -d ' ')
        if [[ "$line_count" -gt "$threshold" ]]; then
            local rel_path="${file#$repo_dir/}"
            long_files=$(echo "$long_files" | jq \
                --arg file "$rel_path" \
                --argjson lines "$line_count" \
                --argjson threshold "$threshold" \
                '. + [{
                    "file": $file,
                    "lines": $lines,
                    "threshold": $threshold,
                    "excess": ($lines - $threshold)
                }]')
        fi
    done <<< "$source_files"

    echo "$long_files"
}

# Calculate code statistics
calculate_stats() {
    local repo_dir="$1"

    local total_lines=0
    local total_files=0
    local comment_lines=0

    # Use cloc if available for accurate stats
    if command -v cloc &> /dev/null; then
        local cloc_output=$(cloc --json "$repo_dir" 2>/dev/null || echo '{}')
        if [[ "$cloc_output" != "{}" ]]; then
            total_lines=$(echo "$cloc_output" | jq -r '.SUM.code // 0')
            comment_lines=$(echo "$cloc_output" | jq -r '.SUM.comment // 0')
            total_files=$(echo "$cloc_output" | jq -r '.SUM.nFiles // 0')
        fi
    else
        # Fallback: simple line count
        total_lines=$(find "$repo_dir" -type f \( -name "*.py" -o -name "*.js" -o -name "*.ts" -o -name "*.java" -o -name "*.go" \) \
            ! -path "*node_modules*" ! -path "*vendor*" ! -path "*.git*" -exec cat {} + 2>/dev/null | wc -l | tr -d ' ')
        total_files=$(find "$repo_dir" -type f \( -name "*.py" -o -name "*.js" -o -name "*.ts" -o -name "*.java" -o -name "*.go" \) \
            ! -path "*node_modules*" ! -path "*vendor*" ! -path "*.git*" 2>/dev/null | wc -l | tr -d ' ')
    fi

    jq -n \
        --argjson total_lines "$total_lines" \
        --argjson total_files "$total_files" \
        --argjson comment_lines "$comment_lines" \
        '{
            "total_lines": $total_lines,
            "total_files": $total_files,
            "comment_lines": $comment_lines
        }'
}

# Check for code duplication using jscpd (if available)
check_duplication() {
    local repo_dir="$1"

    if ! command -v jscpd &> /dev/null; then
        echo '{"available": false, "note": "jscpd not installed - run: npm i -g jscpd"}'
        return
    fi

    # Run jscpd (redirect stdout to stderr to keep function output clean)
    local temp_report=$(mktemp -d)
    jscpd "$repo_dir" --reporters json --output "$temp_report" \
        --ignore "**/node_modules/**,**/vendor/**,**/.git/**,**/dist/**,**/build/**" \
        --min-lines 5 --min-tokens 50 >/dev/null 2>&1 || true

    if [[ -f "$temp_report/jscpd-report.json" ]]; then
        local report=$(cat "$temp_report/jscpd-report.json")
        local percentage=$(echo "$report" | jq -r '.statistics.total.percentage // 0')
        local duplicates=$(echo "$report" | jq -r '.statistics.total.clones // 0')

        rm -rf "$temp_report"

        # Ensure values are valid numbers, defaulting to 0
        percentage=${percentage:-0}
        duplicates=${duplicates:-0}
        # Handle non-numeric values
        case "$percentage" in
            ''|*[!0-9.]*) percentage=0 ;;
        esac
        case "$duplicates" in
            ''|*[!0-9]*) duplicates=0 ;;
        esac

        jq -n \
            --argjson percentage "$percentage" \
            --argjson duplicates "$duplicates" \
            '{
                "available": true,
                "percentage": $percentage,
                "duplicate_blocks": $duplicates
            }'
    else
        rm -rf "$temp_report"
        echo '{"available": true, "percentage": 0, "duplicate_blocks": 0}'
    fi
}

# Calculate weighted debt score using RAG-based thresholds (0-100)
# Category weights from rag/tech-debt/scoring/tech-debt-scoring-guide.md:
#   - Code Markers: 15%
#   - Code Complexity: 20% (future)
#   - Code Duplication: 15%
#   - File Size: 10%
#   - Test Coverage: 15% (from test-coverage-data.sh)
#   - Dependency Debt: 15% (from supply-chain)
#   - Code Churn: 10% (from git-insights)
calculate_debt_score() {
    local todo_count="$1"
    local fixme_count="$2"
    local hack_count="$3"
    local deprecated_count="$4"
    local long_files_count="$5"
    local total_lines="$6"
    local duplication_percent="${7:-0}"
    local test_ratio="${8:-0}"

    # === Category 1: Code Markers (weight: 15) ===
    # Weights from RAG: TODO=0.5, FIXME=1.5, HACK=3.0, XXX=3.0
    # Using integer math: multiply by 10, then divide by 10 at end
    local marker_raw=$(( (todo_count * 5 + fixme_count * 15 + hack_count * 30) / 10 ))
    local marker_score=$marker_raw
    [[ $marker_score -gt 100 ]] && marker_score=100

    # === Category 2: Deprecated Code (part of markers, weight: 5) ===
    # Deprecated: 1.5 points each (using integer: *15/10)
    local deprecated_score=$(( (deprecated_count * 15) / 10 ))
    [[ $deprecated_score -gt 100 ]] && deprecated_score=100

    # === Category 3: File Size (weight: 10) ===
    # From RAG: files >500 lines add to debt
    local file_score=0
    if [[ $long_files_count -gt 0 ]]; then
        # 2 points per long file, capped at 100
        file_score=$((long_files_count * 2))
        file_score=$((file_score > 100 ? 100 : file_score))
    fi

    # === Category 4: Duplication (weight: 15) ===
    # From RAG: 0-3%=0, 3-5%=15, 5-10%=35, 10-20%=65, >20%=100
    local dup_score=0
    local dup_int=${duplication_percent%.*}  # Convert to integer
    dup_int=${dup_int:-0}
    if [[ $dup_int -gt 20 ]]; then
        dup_score=100
    elif [[ $dup_int -gt 10 ]]; then
        dup_score=65
    elif [[ $dup_int -gt 5 ]]; then
        dup_score=35
    elif [[ $dup_int -gt 3 ]]; then
        dup_score=15
    fi

    # === Category 5: Test Coverage (weight: 15) - placeholder ===
    # From RAG: ratio >0.8=0, 0.5-0.8=20, 0.3-0.5=45, 0.1-0.3=70, <0.1=100
    local test_score=50  # Default to moderate if unknown
    # Convert ratio to percentage integer (0.06 -> 6) - use awk for portability
    local test_pct=$(awk -v ratio="$test_ratio" 'BEGIN {printf "%.0f", ratio * 100}')
    test_pct=${test_pct:-0}
    if [[ $test_pct -ge 80 ]]; then
        test_score=0
    elif [[ $test_pct -ge 50 ]]; then
        test_score=20
    elif [[ $test_pct -ge 30 ]]; then
        test_score=45
    elif [[ $test_pct -ge 10 ]]; then
        test_score=70
    elif [[ $test_pct -gt 0 ]]; then
        test_score=90
    else
        test_score=100
    fi

    # === Calculate weighted total ===
    # Weights: markers=15, deprecated=5, file_size=10, duplication=15, test=15
    # (complexity=20, dependency=15, churn=10 would come from other analyzers)
    local total_weight=60  # Sum of weights we calculate here
    local weighted_sum=$((marker_score * 15 + deprecated_score * 5 + file_score * 10 + dup_score * 15 + test_score * 15))
    local final_score=$((weighted_sum / total_weight))

    # Cap at 100
    final_score=$((final_score > 100 ? 100 : final_score))

    echo "$final_score"
}

# Get debt level from score
get_debt_level() {
    local score="$1"
    if [[ $score -le 20 ]]; then
        echo "excellent"
    elif [[ $score -le 40 ]]; then
        echo "good"
    elif [[ $score -le 60 ]]; then
        echo "moderate"
    elif [[ $score -le 80 ]]; then
        echo "high"
    else
        echo "critical"
    fi
}

# Calculate category scores for detailed breakdown
calculate_category_scores() {
    local todo_count="$1"
    local fixme_count="$2"
    local hack_count="$3"
    local deprecated_count="$4"
    local long_files_count="$5"
    local duplication_percent="${6:-0}"
    local test_ratio="${7:-0}"

    # Marker score (integer math)
    local marker_score=$(( (todo_count * 5 + fixme_count * 15 + hack_count * 30) / 10 ))
    [[ $marker_score -gt 100 ]] && marker_score=100

    # Deprecated score
    local deprecated_score=$(( (deprecated_count * 15) / 10 ))
    [[ $deprecated_score -gt 100 ]] && deprecated_score=100

    # File size score
    local file_score=$((long_files_count * 2))
    [[ $file_score -gt 100 ]] && file_score=100

    # Duplication score
    local dup_score=0
    local dup_int=${duplication_percent%.*}
    dup_int=${dup_int:-0}
    if [[ $dup_int -gt 20 ]]; then
        dup_score=100
    elif [[ $dup_int -gt 10 ]]; then
        dup_score=65
    elif [[ $dup_int -gt 5 ]]; then
        dup_score=35
    elif [[ $dup_int -gt 3 ]]; then
        dup_score=15
    fi

    # Test coverage score - use awk for portability
    local test_score=50
    local test_pct=$(awk -v ratio="$test_ratio" 'BEGIN {printf "%.0f", ratio * 100}')
    test_pct=${test_pct:-0}
    if [[ $test_pct -ge 80 ]]; then
        test_score=0
    elif [[ $test_pct -ge 50 ]]; then
        test_score=20
    elif [[ $test_pct -ge 30 ]]; then
        test_score=45
    elif [[ $test_pct -ge 10 ]]; then
        test_score=70
    elif [[ $test_pct -gt 0 ]]; then
        test_score=90
    else
        test_score=100
    fi

    jq -n \
        --argjson markers "$marker_score" \
        --argjson deprecated "$deprecated_score" \
        --argjson file_size "$file_score" \
        --argjson duplication "$dup_score" \
        --argjson test_coverage "$test_score" \
        '{
            "markers": {"score": $markers, "weight": 15, "level": (if $markers <= 20 then "excellent" elif $markers <= 40 then "good" elif $markers <= 60 then "moderate" elif $markers <= 80 then "high" else "critical" end)},
            "deprecated": {"score": $deprecated, "weight": 5, "level": (if $deprecated <= 20 then "excellent" elif $deprecated <= 40 then "good" elif $deprecated <= 60 then "moderate" elif $deprecated <= 80 then "high" else "critical" end)},
            "file_size": {"score": $file_size, "weight": 10, "level": (if $file_size <= 20 then "excellent" elif $file_size <= 40 then "good" elif $file_size <= 60 then "moderate" elif $file_size <= 80 then "high" else "critical" end)},
            "duplication": {"score": $duplication, "weight": 15, "level": (if $duplication <= 20 then "excellent" elif $duplication <= 40 then "good" elif $duplication <= 60 then "moderate" elif $duplication <= 80 then "high" else "critical" end)},
            "test_coverage": {"score": $test_coverage, "weight": 15, "level": (if $test_coverage <= 20 then "excellent" elif $test_coverage <= 40 then "good" elif $test_coverage <= 60 then "moderate" elif $test_coverage <= 80 then "high" else "critical" end)}
        }'
}

# Main analysis
analyze_target() {
    local repo_dir="$1"

    echo -e "${BLUE}Scanning for debt markers (TODO/FIXME/HACK/XXX)...${NC}" >&2

    # Create file descriptors for counts
    exec 3>&1 4>&1 5>&1 6>&1

    # Scan markers - capture both output and counts
    local todo_count=0
    local fixme_count=0
    local hack_count=0
    local xxx_count=0

    local markers="[]"
    local patterns="TODO|FIXME|HACK|XXX|BUG|KLUDGE|OPTIMIZE|REFACTOR"

    # Find all source files
    local source_files=$(find "$repo_dir" -type f \( \
        -name "*.py" -o -name "*.js" -o -name "*.ts" -o -name "*.jsx" -o -name "*.tsx" \
        -o -name "*.java" -o -name "*.go" -o -name "*.rb" -o -name "*.php" \
        -o -name "*.c" -o -name "*.cpp" -o -name "*.h" -o -name "*.hpp" \
        -o -name "*.cs" -o -name "*.swift" -o -name "*.kt" -o -name "*.rs" \
        -o -name "*.scala" -o -name "*.sh" -o -name "*.bash" \
    \) ! -path "*node_modules*" ! -path "*vendor*" ! -path "*.git*" ! -path "*dist*" ! -path "*build*" 2>/dev/null)

    while IFS= read -r file; do
        [[ -z "$file" ]] && continue

        while IFS=: read -r line_num content; do
            [[ -z "$line_num" ]] && continue

            local rel_path="${file#$repo_dir/}"
            local marker_type=""

            if echo "$content" | grep -qi "TODO"; then
                marker_type="TODO"
                ((todo_count++))
            elif echo "$content" | grep -qi "FIXME"; then
                marker_type="FIXME"
                ((fixme_count++))
            elif echo "$content" | grep -qi "HACK"; then
                marker_type="HACK"
                ((hack_count++))
            elif echo "$content" | grep -qi "XXX"; then
                marker_type="XXX"
                ((xxx_count++))
            else
                marker_type="OTHER"
            fi

            local clean_text=$(echo "$content" | sed 's/^[[:space:]]*[#/*-]*[[:space:]]*//' | sed 's/^\/\///')
            if [[ ${#clean_text} -gt 200 ]]; then
                clean_text="${clean_text:0:197}..."
            fi

            markers=$(echo "$markers" | jq \
                --arg type "$marker_type" \
                --arg file "$rel_path" \
                --argjson line "$line_num" \
                --arg text "$clean_text" \
                '. + [{
                    "type": $type,
                    "text": $text,
                    "file": $file,
                    "line": $line
                }]')

        done < <(grep -n -E "($patterns)" "$file" 2>/dev/null | head -100)

    done <<< "$source_files"

    local total_markers=$((todo_count + fixme_count + hack_count + xxx_count))
    echo -e "${GREEN}✓ Found $total_markers debt markers${NC}" >&2

    echo -e "${BLUE}Scanning for deprecated code...${NC}" >&2
    local deprecated=$(scan_deprecated "$repo_dir")
    local deprecated_count=$(echo "$deprecated" | jq 'length')
    echo -e "${GREEN}✓ Found $deprecated_count deprecated items${NC}" >&2

    echo -e "${BLUE}Finding long files (>${LONG_FILE_LINES} lines)...${NC}" >&2
    local long_files=$(find_long_files "$repo_dir" "$LONG_FILE_LINES")
    local long_files_count=$(echo "$long_files" | jq 'length')
    echo -e "${GREEN}✓ Found $long_files_count long files${NC}" >&2

    echo -e "${BLUE}Calculating code statistics...${NC}" >&2
    local stats=$(calculate_stats "$repo_dir")
    local total_lines=$(echo "$stats" | jq -r '.total_lines')
    echo -e "${GREEN}✓ $total_lines lines of code${NC}" >&2

    echo -e "${BLUE}Checking for code duplication...${NC}" >&2
    local duplication=$(check_duplication "$repo_dir")
    local dup_available=$(echo "$duplication" | jq -r '.available')
    local dup_percent=0
    if [[ "$dup_available" == "true" ]]; then
        dup_percent=$(echo "$duplication" | jq -r '.percentage')
        echo -e "${GREEN}✓ ${dup_percent}% code duplication${NC}" >&2
    else
        echo -e "${YELLOW}○ jscpd not installed (optional)${NC}" >&2
    fi

    # Calculate test-to-code ratio (simplified - count test files vs source files)
    echo -e "${BLUE}Estimating test coverage...${NC}" >&2
    local test_files=$(find "$repo_dir" -type f \( \
        -name "*.test.*" -o -name "*.spec.*" -o -name "test_*.py" -o -name "*_test.go" \
    \) ! -path "*node_modules*" ! -path "*vendor*" 2>/dev/null | wc -l | tr -d ' ')
    local total_source=$(echo "$stats" | jq -r '.total_files')
    local test_ratio=0
    if [[ $total_source -gt 0 ]]; then
        test_ratio=$(echo "scale=2; $test_files / $total_source" | bc)
    fi
    echo -e "${GREEN}✓ Test ratio: $test_ratio ($test_files test files)${NC}" >&2

    # Calculate debt score using weighted categories
    local debt_score=$(calculate_debt_score "$todo_count" "$fixme_count" "$hack_count" "$deprecated_count" "$long_files_count" "$total_lines" "$dup_percent" "$test_ratio")

    # Determine debt level using get_debt_level function
    local debt_level=$(get_debt_level "$debt_score")

    # Calculate category breakdown
    local category_scores=$(calculate_category_scores "$todo_count" "$fixme_count" "$hack_count" "$deprecated_count" "$long_files_count" "$dup_percent" "$test_ratio")

    # Display score with color based on level
    if [[ "$debt_level" == "critical" ]]; then
        echo -e "${RED}Debt Score: $debt_score/100 ($debt_level)${NC}" >&2
    elif [[ "$debt_level" == "high" ]]; then
        echo -e "${YELLOW}Debt Score: $debt_score/100 ($debt_level)${NC}" >&2
    else
        echo -e "${CYAN}Debt Score: $debt_score/100 ($debt_level)${NC}" >&2
    fi

    # Build final output
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    jq -n \
        --arg ts "$timestamp" \
        --arg tgt "$TARGET" \
        --arg ver "2.0.0" \
        --argjson debt_score "$debt_score" \
        --arg debt_level "$debt_level" \
        --argjson todo_count "$todo_count" \
        --argjson fixme_count "$fixme_count" \
        --argjson hack_count "$hack_count" \
        --argjson xxx_count "$xxx_count" \
        --argjson deprecated_count "$deprecated_count" \
        --argjson long_files_count "$long_files_count" \
        --argjson test_files "$test_files" \
        --arg test_ratio "$test_ratio" \
        --argjson markers "$markers" \
        --argjson deprecated "$deprecated" \
        --argjson long_files "$long_files" \
        --argjson stats "$stats" \
        --argjson duplication "$duplication" \
        --argjson category_scores "$category_scores" \
        '{
            analyzer: "tech-debt",
            version: $ver,
            timestamp: $ts,
            target: $tgt,
            summary: {
                debt_score: $debt_score,
                debt_level: $debt_level,
                todo_count: $todo_count,
                fixme_count: $fixme_count,
                hack_count: $hack_count,
                xxx_count: $xxx_count,
                deprecated_count: $deprecated_count,
                long_files_count: $long_files_count,
                test_files: $test_files,
                test_ratio: ($test_ratio | tonumber),
                total_markers: ($todo_count + $fixme_count + $hack_count + $xxx_count)
            },
            category_scores: $category_scores,
            code_stats: $stats,
            markers: $markers,
            deprecated: $deprecated,
            long_files: $long_files,
            duplication: $duplication
        }'
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help) usage ;;
        --local-path)
            LOCAL_PATH="$2"
            shift 2
            ;;
        --repo)
            REPO="$2"
            shift 2
            ;;
        --org)
            ORG="$2"
            shift 2
            ;;
        --max-files)
            MAX_FILES="$2"
            shift 2
            ;;
        --exclude)
            EXCLUDE_PATTERNS="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_FILE="$2"
            shift 2
            ;;
        -k|--keep-clone)
            CLEANUP=false
            shift
            ;;
        -*)
            echo "Unknown option: $1" >&2
            exit 1
            ;;
        *)
            TARGET="$1"
            shift
            ;;
    esac
done

# Main execution
scan_path=""

if [[ -n "$LOCAL_PATH" ]]; then
    [[ ! -d "$LOCAL_PATH" ]] && { echo '{"error": "Local path does not exist"}'; exit 1; }
    scan_path="$LOCAL_PATH"
    TARGET="$LOCAL_PATH"
elif [[ -n "$REPO" ]]; then
    # Look in zero cache
    REPO_ORG=$(echo "$REPO" | cut -d'/' -f1)
    REPO_NAME=$(echo "$REPO" | cut -d'/' -f2)
    ZERO_CACHE_PATH="$HOME/.zero/projects/$REPO_ORG/$REPO_NAME/repo"
    LEGACY_PATH="$HOME/.zero/projects/${REPO_ORG}-${REPO_NAME}/repo"

    if [[ -d "$ZERO_CACHE_PATH" ]]; then
        scan_path="$ZERO_CACHE_PATH"
        TARGET="$REPO"
    elif [[ -d "$LEGACY_PATH" ]]; then
        scan_path="$LEGACY_PATH"
        TARGET="$REPO"
    else
        echo '{"error": "Repository not found in cache. Clone it first or use --local-path"}'
        exit 1
    fi
elif [[ -n "$ORG" ]]; then
    # Scan ALL repos in the org
    ORG_PATH="$HOME/.zero/projects/$ORG"
    if [[ -d "$ORG_PATH" ]]; then
        # Collect repos with and without cloned code
        REPOS_TO_SCAN=()
        REPOS_NOT_CLONED=()
        for repo_dir in "$ORG_PATH"/*/; do
            repo_name=$(basename "$repo_dir")
            if [[ -d "$repo_dir/repo" ]]; then
                REPOS_TO_SCAN+=("$repo_name")
            else
                REPOS_NOT_CLONED+=("$repo_name")
            fi
        done

        # Check if there are uncloned repos and prompt user
        if [[ ${#REPOS_NOT_CLONED[@]} -gt 0 ]]; then
            echo -e "${YELLOW}Found ${#REPOS_NOT_CLONED[@]} repositories without cloned code:${NC}" >&2
            for repo in "${REPOS_NOT_CLONED[@]}"; do
                echo -e "  - $repo" >&2
            done
            echo "" >&2

            # Prompt user
            read -p "Would you like to hydrate these repos for analysis? [y/N] " -n 1 -r >&2
            echo "" >&2

            if [[ $REPLY =~ ^[Yy]$ ]]; then
                echo -e "${BLUE}Hydrating ${#REPOS_NOT_CLONED[@]} repositories...${NC}" >&2

                # Run hydration for each uncloned repo
                for repo in "${REPOS_NOT_CLONED[@]}"; do
                    echo -e "${CYAN}Cloning $ORG/$repo...${NC}" >&2
                    "$REPO_ROOT/utils/zero/hydrate.sh" --repo "$ORG/$repo" --quick >&2 2>&1 || true

                    # Check if clone succeeded
                    if [[ -d "$ORG_PATH/$repo/repo" ]]; then
                        REPOS_TO_SCAN+=("$repo")
                        echo -e "${GREEN}✓ $repo ready${NC}" >&2
                    else
                        echo -e "${RED}✗ Failed to clone $repo${NC}" >&2
                    fi
                done
                echo "" >&2
            else
                echo -e "${CYAN}Continuing with ${#REPOS_TO_SCAN[@]} already-cloned repositories...${NC}" >&2
            fi
        fi

        if [[ ${#REPOS_TO_SCAN[@]} -eq 0 ]]; then
            echo '{"error": "No repositories with cloned code found in org cache. Hydrate repos first."}'
            exit 1
        fi

        # Analyze each repo and aggregate results
        echo -e "${BLUE}Scanning ${#REPOS_TO_SCAN[@]} repositories in $ORG...${NC}" >&2

        all_results="[]"
        repo_count=0
        total_repos=${#REPOS_TO_SCAN[@]}

        for repo_name in "${REPOS_TO_SCAN[@]}"; do
            ((repo_count++))
            scan_path="$ORG_PATH/$repo_name/repo"
            TARGET="$ORG/$repo_name"

            echo -e "\n${CYAN}[$repo_count/$total_repos] Analyzing: $TARGET${NC}" >&2

            repo_json=$(analyze_target "$scan_path")

            # Add repo identifier to each result AND to each finding within (reordered for readability)
            repo_json=$(echo "$repo_json" | jq --arg repo "$TARGET" '
                . + {repository: $repo} |
                .markers = [.markers[] | {type, text, repository: $repo, file, line}] |
                .deprecated = [.deprecated[] | {type: "DEPRECATED", context, repository: $repo, file, line}] |
                .long_files = [.long_files[] | {type: "LONG_FILE", lines, repository: $repo, file, threshold, excess}]
            ')

            # Extract summary for display
            debt_score=$(echo "$repo_json" | jq -r '.summary.debt_score')
            debt_level=$(echo "$repo_json" | jq -r '.summary.debt_level')
            total_markers=$(echo "$repo_json" | jq -r '.summary.total_markers')

            echo -e "${GREEN}  → Score: $debt_score/100 ($debt_level) | $total_markers markers${NC}" >&2

            all_results=$(echo "$all_results" | jq --argjson repo "$repo_json" '. + [$repo]')
        done

        # Build aggregated output
        timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

        # Calculate org-wide totals
        total_todo=$(echo "$all_results" | jq '[.[].summary.todo_count] | add')
        total_fixme=$(echo "$all_results" | jq '[.[].summary.fixme_count] | add')
        total_hack=$(echo "$all_results" | jq '[.[].summary.hack_count] | add')
        total_xxx=$(echo "$all_results" | jq '[.[].summary.xxx_count] | add')
        total_deprecated=$(echo "$all_results" | jq '[.[].summary.deprecated_count] | add')
        total_long_files=$(echo "$all_results" | jq '[.[].summary.long_files_count] | add')
        total_test_files=$(echo "$all_results" | jq '[.[].summary.test_files] | add')
        avg_debt_score=$(echo "$all_results" | jq '[.[].summary.debt_score] | add / length | floor')

        # Calculate org-level category score averages
        org_category_scores=$(echo "$all_results" | jq '{
            markers: {
                score: ([.[].category_scores.markers.score] | add / length | floor),
                weight: 15,
                level: (if ([.[].category_scores.markers.score] | add / length) <= 20 then "excellent" elif ([.[].category_scores.markers.score] | add / length) <= 40 then "good" elif ([.[].category_scores.markers.score] | add / length) <= 60 then "moderate" elif ([.[].category_scores.markers.score] | add / length) <= 80 then "high" else "critical" end)
            },
            deprecated: {
                score: ([.[].category_scores.deprecated.score] | add / length | floor),
                weight: 5,
                level: (if ([.[].category_scores.deprecated.score] | add / length) <= 20 then "excellent" elif ([.[].category_scores.deprecated.score] | add / length) <= 40 then "good" elif ([.[].category_scores.deprecated.score] | add / length) <= 60 then "moderate" elif ([.[].category_scores.deprecated.score] | add / length) <= 80 then "high" else "critical" end)
            },
            file_size: {
                score: ([.[].category_scores.file_size.score] | add / length | floor),
                weight: 10,
                level: (if ([.[].category_scores.file_size.score] | add / length) <= 20 then "excellent" elif ([.[].category_scores.file_size.score] | add / length) <= 40 then "good" elif ([.[].category_scores.file_size.score] | add / length) <= 60 then "moderate" elif ([.[].category_scores.file_size.score] | add / length) <= 80 then "high" else "critical" end)
            },
            duplication: {
                score: ([.[].category_scores.duplication.score] | add / length | floor),
                weight: 15,
                level: (if ([.[].category_scores.duplication.score] | add / length) <= 20 then "excellent" elif ([.[].category_scores.duplication.score] | add / length) <= 40 then "good" elif ([.[].category_scores.duplication.score] | add / length) <= 60 then "moderate" elif ([.[].category_scores.duplication.score] | add / length) <= 80 then "high" else "critical" end)
            },
            test_coverage: {
                score: ([.[].category_scores.test_coverage.score] | add / length | floor),
                weight: 15,
                level: (if ([.[].category_scores.test_coverage.score] | add / length) <= 20 then "excellent" elif ([.[].category_scores.test_coverage.score] | add / length) <= 40 then "good" elif ([.[].category_scores.test_coverage.score] | add / length) <= 60 then "moderate" elif ([.[].category_scores.test_coverage.score] | add / length) <= 80 then "high" else "critical" end)
            }
        }')

        # Calculate org debt level
        org_debt_level="moderate"
        if [[ $avg_debt_score -le 20 ]]; then
            org_debt_level="excellent"
        elif [[ $avg_debt_score -le 40 ]]; then
            org_debt_level="good"
        elif [[ $avg_debt_score -le 60 ]]; then
            org_debt_level="moderate"
        elif [[ $avg_debt_score -le 80 ]]; then
            org_debt_level="high"
        else
            org_debt_level="critical"
        fi

        final_json=$(jq -n \
            --arg ts "$timestamp" \
            --arg org "$ORG" \
            --arg ver "2.0.0" \
            --argjson repo_count "$total_repos" \
            --argjson avg_debt_score "$avg_debt_score" \
            --arg debt_level "$org_debt_level" \
            --argjson total_todo "$total_todo" \
            --argjson total_fixme "$total_fixme" \
            --argjson total_hack "$total_hack" \
            --argjson total_xxx "$total_xxx" \
            --argjson total_deprecated "$total_deprecated" \
            --argjson total_long_files "$total_long_files" \
            --argjson total_test_files "$total_test_files" \
            --argjson category_scores "$org_category_scores" \
            --argjson repositories "$all_results" \
            '{
                analyzer: "tech-debt",
                version: $ver,
                timestamp: $ts,
                organization: $org,
                summary: {
                    repositories_scanned: $repo_count,
                    avg_debt_score: $avg_debt_score,
                    debt_level: $debt_level,
                    total_todo: $total_todo,
                    total_fixme: $total_fixme,
                    total_hack: $total_hack,
                    total_xxx: $total_xxx,
                    total_deprecated: $total_deprecated,
                    total_long_files: $total_long_files,
                    total_test_files: $total_test_files,
                    total_markers: ($total_todo + $total_fixme + $total_hack + $total_xxx)
                },
                category_scores: $category_scores,
                repositories: $repositories
            }')

        echo -e "\n${CYAN}=== Organization Summary ===${NC}" >&2
        echo -e "${CYAN}Repos: $total_repos | Avg Score: $avg_debt_score/100 ($org_debt_level) | Total Markers: $((total_todo + total_fixme + total_hack + total_xxx))${NC}" >&2

        # Output
        if [[ -n "$OUTPUT_FILE" ]]; then
            echo "$final_json" > "$OUTPUT_FILE"
            echo -e "${GREEN}✓ Results written to $OUTPUT_FILE${NC}" >&2
        else
            echo "$final_json"
        fi
        exit 0
    else
        echo '{"error": "Org not found in cache. Hydrate repos first."}'
        exit 1
    fi
elif [[ -n "$TARGET" ]]; then
    if is_git_url "$TARGET"; then
        clone_repository "$TARGET"
        scan_path="$TEMP_DIR"
    elif [[ -d "$TARGET" ]]; then
        scan_path="$TARGET"
    else
        echo '{"error": "Invalid target - must be URL or directory"}'
        exit 1
    fi
else
    echo '{"error": "No target specified"}'
    exit 1
fi

echo -e "${BLUE}Analyzing technical debt: $TARGET${NC}" >&2

final_json=$(analyze_target "$scan_path")

# Output
if [[ -n "$OUTPUT_FILE" ]]; then
    echo "$final_json" > "$OUTPUT_FILE"
    echo -e "${GREEN}✓ Results written to $OUTPUT_FILE${NC}" >&2
else
    echo "$final_json"
fi
