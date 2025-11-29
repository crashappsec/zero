#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# 
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Core Analysis Functions
# Dual-method ownership analysis (commit-based and line-based)
#############################################################################

# Analyze ownership using commit-based method
# Returns: author|email|file|commits|last_commit_date
analyze_commit_based_ownership() {
    local repo_path="$1"
    local since_date="$2"
    local output_file="$3"

    cd "$repo_path" || return 1

    # Get commit-based ownership for each file
    git ls-files | while read -r file; do
        # Get contributors to this file since the date
        git log --since="$since_date" --format="%an|%ae|%ad" --date=short --follow -- "$file" 2>/dev/null | \
        awk -F'|' -v file="$file" '
        {
            author=$1
            email=$2
            date=$3
            key=author"|"email

            commits[key]++
            files[key"|"file]++

            # Track most recent commit date
            if (last_date[key] == "" || date > last_date[key]) {
                last_date[key] = date
            }
        }
        END {
            for (key in commits) {
                for (f in files) {
                    if (index(f, key"|") == 1) {
                        fname = substr(f, length(key) + 2)
                        print key"|"fname"|"files[f]"|"last_date[key]
                    }
                }
            }
        }
        ' >> "$output_file"
    done
}

# Analyze ownership using line-based method (git blame)
# Returns: author|email|file|lines|percentage
analyze_line_based_ownership() {
    local repo_path="$1"
    local output_file="$2"

    cd "$repo_path" || return 1

    # Get line-based ownership for each file
    git ls-files | while read -r file; do
        # Skip binary files and very large files
        if file "$file" | grep -q "text"; then
            local total_lines=$(wc -l < "$file" 2>/dev/null || echo "0")

            if [[ $total_lines -gt 0 ]] && [[ $total_lines -lt 10000 ]]; then
                # Use git blame to get line ownership
                git blame --line-porcelain "$file" 2>/dev/null | \
                awk -v file="$file" -v total="$total_lines" '
                /^author / {author=substr($0, 8)}
                /^author-mail / {
                    email=substr($0, 14)
                    gsub(/[<>]/, "", email)
                    key=author"|"email
                    lines[key]++
                }
                END {
                    for (key in lines) {
                        percentage = (lines[key] / total) * 100
                        print key"|"file"|"lines[key]"|"percentage
                    }
                }
                ' >> "$output_file"
            fi
        fi
    done
}

# Combine commit-based and line-based results
# Creates hybrid ownership scores
combine_ownership_methods() {
    local commit_file="$1"
    local line_file="$2"
    local output_file="$3"

    # Process both files and combine
    {
        echo "# Commit-based contributions"
        cat "$commit_file"
        echo "# Line-based contributions"
        cat "$line_file"
    } | awk -F'|' '
    /^#/ {next}
    {
        author=$1
        email=$2
        file=$3
        key=author"|"email"|"file

        # Track both methods
        if (NF == 5) {
            # Commit-based: author|email|file|commits|date
            commit_count[key] = $4
            last_commit[key] = $5
        } else if (NF == 4) {
            # Line-based: author|email|file|lines|percentage
            line_count[key] = $4
            line_pct[key] = $5
        }
    }
    END {
        for (key in commit_count) {
            # Calculate hybrid score
            commits = commit_count[key]
            lines = line_count[key] + 0

            # Weighted score: commits 60%, lines 40%
            hybrid_score = (commits * 0.6) + (lines * 0.004)

            print key"|"commits"|"lines"|"hybrid_score"|"last_commit[key]
        }

        # Also output line-only entries
        for (key in line_count) {
            if (!(key in commit_count)) {
                print key"|0|"line_count[key]"|"(line_count[key] * 0.004)"|unknown"
            }
        }
    }
    ' > "$output_file"
}

# Calculate comprehensive contributor statistics
calculate_contributor_stats() {
    local repo_path="$1"
    local since_date="$2"
    local output_file="$3"

    cd "$repo_path" || return 1

    # Get detailed stats per contributor
    git log --since="$since_date" --format="%an|%ae|%ad" --date=short --numstat | \
    awk -F'|' '
    NF==3 {
        author=$1
        email=$2
        date=$3
        key=author"|"email
        next
    }
    NF==3 && $1 != "-" {
        # numstat output: additions deletions filename
        split($0, parts, "\t")
        added[key] += parts[1]
        deleted[key] += parts[2]
        files_changed[key]++
    }
    END {
        for (key in added) {
            net = added[key] - deleted[key]
            print key"|"added[key]"|"deleted[key]"|"net"|"files_changed[key]
        }
    }
    ' > "$output_file"
}

# Detect Single Points of Failure (SPOF) using 6 criteria
detect_spof() {
    local repo_path="$1"
    local ownership_file="$2"
    local output_file="$3"

    cd "$repo_path" || return 1

    # Analyze each file for SPOF criteria
    git ls-files | while read -r file; do
        # Criterion 1: Single contributor
        local contributor_count=$(grep "|$file|" "$ownership_file" | wc -l | tr -d ' ')

        # Criterion 2: Critical path (heuristic based on path)
        local is_critical=0
        if [[ "$file" =~ (auth|payment|security|crypto|login|passwd) ]]; then
            is_critical=1
        fi

        # Criterion 3: High complexity (>500 LOC)
        local loc=$(wc -l < "$file" 2>/dev/null || echo "0")
        local is_complex=0
        if [[ $loc -gt 500 ]]; then
            is_complex=1
        fi

        # Criterion 4: No backup owner (checked from ownership_file)
        local has_backup=0
        if [[ $contributor_count -gt 1 ]]; then
            has_backup=1
        fi

        # Criterion 5: Low test coverage (heuristic - check if test file exists)
        local has_tests=0
        local test_file=""
        for ext in test.js test.ts spec.js spec.ts _test.go test.py; do
            test_file="${file%.*}.$ext"
            if [[ -f "$test_file" ]]; then
                has_tests=1
                break
            fi
        done

        # Criterion 6: No documentation (check for README or comments)
        local has_docs=0
        local dir=$(dirname "$file")
        if [[ -f "$dir/README.md" ]] || [[ -f "$dir/README.txt" ]]; then
            has_docs=1
        fi

        # Count SPOF criteria met
        local spof_score=0
        [[ $contributor_count -eq 1 ]] && ((spof_score++))
        [[ $is_critical -eq 1 ]] && ((spof_score++))
        [[ $is_complex -eq 1 ]] && ((spof_score++))
        [[ $has_backup -eq 0 ]] && ((spof_score++))
        [[ $has_tests -eq 0 ]] && ((spof_score++))
        [[ $has_docs -eq 0 ]] && ((spof_score++))

        # Determine risk level
        local risk_level="Low"
        if [[ $spof_score -ge 6 ]]; then
            risk_level="Critical"
        elif [[ $spof_score -ge 4 ]]; then
            risk_level="High"
        elif [[ $spof_score -ge 2 ]]; then
            risk_level="Medium"
        fi

        # Output SPOF analysis
        if [[ $spof_score -ge 2 ]]; then
            echo "$file|$spof_score|$risk_level|$contributor_count|$is_critical|$is_complex|$has_backup|$has_tests|$has_docs|$loc" >> "$output_file"
        fi
    done
}

# Calculate repository-wide metrics
calculate_repository_metrics() {
    local repo_path="$1"
    local since_date="$2"
    local ownership_file="$3"

    cd "$repo_path" || return 1

    # Total files
    local total_files=$(git ls-files | wc -l | tr -d ' ')

    # Total commits in period
    local total_commits=$(git log --since="$since_date" --oneline | wc -l | tr -d ' ')

    # Active contributors
    local active_contributors=$(git log --since="$since_date" --format="%ae" | sort -u | wc -l | tr -d ' ')

    # Files with owners (from ownership file)
    local files_with_owners=$(cut -d'|' -f3 "$ownership_file" | sort -u | wc -l | tr -d ' ')

    # Coverage
    local coverage=0
    if [[ $total_files -gt 0 ]]; then
        coverage=$(echo "scale=2; ($files_with_owners / $total_files) * 100" | bc -l)
    fi

    # Output metrics
    echo "$total_files|$total_commits|$active_contributors|$files_with_owners|$coverage"
}

# Export functions
export -f analyze_commit_based_ownership
export -f analyze_line_based_ownership
export -f combine_ownership_methods
export -f calculate_contributor_stats
export -f detect_spof
export -f calculate_repository_metrics
