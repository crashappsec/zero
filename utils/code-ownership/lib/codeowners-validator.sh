#!/bin/bash
# Copyright (c) 2024 Crash Override Inc
# 101 Fulton St, 416, New York 10038
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# CODEOWNERS Validation Library
# Advanced validation based on research-backed best practices
#############################################################################

# Parse CODEOWNERS file and extract patterns
parse_codeowners() {
    local codeowners_file="$1"
    local output_file="$2"

    if [[ ! -f "$codeowners_file" ]]; then
        return 1
    fi

    # Extract non-comment, non-empty lines
    grep -v "^#" "$codeowners_file" | grep -v "^$" | while IFS= read -r line; do
        # Extract pattern (first field) and owners (remaining fields)
        local pattern=$(echo "$line" | awk '{print $1}')
        local owners=$(echo "$line" | cut -d' ' -f2-)

        echo "$pattern|$owners"
    done > "$output_file"
}

# Validate CODEOWNERS syntax
validate_syntax() {
    local codeowners_file="$1"
    local output_file="$2"

    local line_num=0
    local errors=0

    while IFS= read -r line; do
        ((line_num++))

        # Skip comments and empty lines
        [[ "$line" =~ ^#.*$ ]] && continue
        [[ -z "$line" ]] && continue

        # Check for pattern
        local pattern=$(echo "$line" | awk '{print $1}')
        if [[ -z "$pattern" ]]; then
            echo "Line $line_num: Missing pattern" >> "$output_file"
            ((errors++))
            continue
        fi

        # Check for owners
        local owners=$(echo "$line" | cut -d' ' -f2-)
        if [[ -z "$owners" ]]; then
            echo "Line $line_num: Pattern '$pattern' has no owners" >> "$output_file"
            ((errors++))
            continue
        fi

        # Validate owner format (@username or @org/team)
        for owner in $owners; do
            if [[ ! "$owner" =~ ^@[a-zA-Z0-9_-]+(/[a-zA-Z0-9_-]+)?$ ]]; then
                echo "Line $line_num: Invalid owner format '$owner' (should be @username or @org/team)" >> "$output_file"
                ((errors++))
            fi
        done
    done < "$codeowners_file"

    echo "$errors"
}

# Check for stale owners (inactive >90 days)
check_stale_owners() {
    local repo_path="$1"
    local codeowners_file="$2"
    local staleness_days="${3:-90}"
    local output_file="$4"

    cd "$repo_path" || return 1

    local stale_count=0
    local patterns_file=$(mktemp)
    parse_codeowners "$codeowners_file" "$patterns_file"

    while IFS='|' read -r pattern owners; do
        # For each owner in the pattern
        for owner in $owners; do
            # Strip @ from owner
            local username="${owner#@}"

            # Try to find git email for this username
            # This is a heuristic - we look for commits with username in email
            local last_commit_date=$(git log --all --format="%ae|%ad" --date=short | \
                grep -i "$username" | head -1 | cut -d'|' -f2)

            if [[ -n "$last_commit_date" ]]; then
                # Calculate days since last commit
                local days_since=$(( ($(date +%s) - $(date -j -f "%Y-%m-%d" "$last_commit_date" +%s 2>/dev/null || date -d "$last_commit_date" +%s)) / 86400 ))

                if [[ $days_since -gt $staleness_days ]]; then
                    echo "$pattern|$owner|$days_since" >> "$output_file"
                    ((stale_count++))
                fi
            fi
        done
    done < "$patterns_file"

    rm -f "$patterns_file"
    echo "$stale_count"
}

# Check for coverage gaps (files without ownership)
check_coverage_gaps() {
    local repo_path="$1"
    local codeowners_file="$2"
    local output_file="$3"

    cd "$repo_path" || return 1

    local patterns_file=$(mktemp)
    parse_codeowners "$codeowners_file" "$patterns_file"

    # Get all files in repo
    git ls-files | while read -r file; do
        local has_owner=false

        # Check if file matches any pattern
        while IFS='|' read -r pattern owners; do
            # Convert CODEOWNERS pattern to shell glob
            # This is simplified - full implementation would need more sophisticated matching
            if [[ "$file" == $pattern ]] || [[ "$file" == ${pattern#/} ]]; then
                has_owner=true
                break
            fi

            # Check for wildcard patterns
            if [[ "$pattern" == *"*"* ]]; then
                # Simple wildcard matching
                local pattern_regex=$(echo "$pattern" | sed 's/\*/.*/' | sed 's/\//\\\//g')
                if echo "$file" | grep -q "^$pattern_regex$"; then
                    has_owner=true
                    break
                fi
            fi
        done < "$patterns_file"

        if [[ "$has_owner" == "false" ]]; then
            echo "$file" >> "$output_file"
        fi
    done

    rm -f "$patterns_file"

    # Count gaps
    if [[ -f "$output_file" ]]; then
        wc -l < "$output_file" | tr -d ' '
    else
        echo "0"
    fi
}

# Detect anti-patterns in CODEOWNERS
detect_antipatterns() {
    local codeowners_file="$1"
    local output_file="$2"

    local issues=0

    # Anti-pattern 1: Overly broad catch-all with single owner
    if grep -q "^\* @[a-zA-Z0-9_-]*$" "$codeowners_file"; then
        echo "Anti-pattern: Single owner for all files (*)" >> "$output_file"
        ((issues++))
    fi

    # Anti-pattern 2: Too many owners on single pattern (>5)
    grep -v "^#" "$codeowners_file" | grep -v "^$" | while IFS= read -r line; do
        local owner_count=$(echo "$line" | awk '{print NF-1}')
        if [[ $owner_count -gt 5 ]]; then
            echo "Anti-pattern: Too many owners ($owner_count) on pattern: $(echo "$line" | awk '{print $1}')" >> "$output_file"
            ((issues++))
        fi
    done

    # Anti-pattern 3: No default owner pattern
    if ! grep -q "^\*" "$codeowners_file"; then
        echo "Anti-pattern: No default owner pattern (*) specified" >> "$output_file"
        ((issues++))
    fi

    # Anti-pattern 4: Duplicate patterns
    local patterns=$(grep -v "^#" "$codeowners_file" | grep -v "^$" | awk '{print $1}' | sort)
    local duplicates=$(echo "$patterns" | uniq -d)
    if [[ -n "$duplicates" ]]; then
        echo "Anti-pattern: Duplicate patterns found: $duplicates" >> "$output_file"
        ((issues++))
    fi

    echo "$issues"
}

# Generate validation report
generate_validation_report() {
    local codeowners_file="$1"
    local repo_path="$2"
    local format="${3:-text}"

    local syntax_errors=$(mktemp)
    local stale_owners=$(mktemp)
    local coverage_gaps=$(mktemp)
    local antipatterns=$(mktemp)

    # Run all validations
    local syntax_error_count=$(validate_syntax "$codeowners_file" "$syntax_errors")
    local stale_owner_count=$(check_stale_owners "$repo_path" "$codeowners_file" 90 "$stale_owners")
    local coverage_gap_count=$(check_coverage_gaps "$repo_path" "$codeowners_file" "$coverage_gaps")
    local antipattern_count=$(detect_antipatterns "$codeowners_file" "$antipatterns")

    if [[ "$format" == "json" ]]; then
        # Generate JSON report
        jq -n \
            --arg file "$codeowners_file" \
            --arg syntax_errors "$syntax_error_count" \
            --arg stale_owners "$stale_owner_count" \
            --arg coverage_gaps "$coverage_gap_count" \
            --arg antipatterns "$antipattern_count" \
            --arg syntax_details "$(cat "$syntax_errors" 2>/dev/null || echo "")" \
            --arg stale_details "$(cat "$stale_owners" 2>/dev/null || echo "")" \
            --arg gap_details "$(cat "$coverage_gaps" 2>/dev/null || echo "")" \
            --arg antipattern_details "$(cat "$antipatterns" 2>/dev/null || echo "")" \
            '{
                file: $file,
                validation: {
                    syntax_errors: ($syntax_errors | tonumber),
                    stale_owners: ($stale_owners | tonumber),
                    coverage_gaps: ($coverage_gaps | tonumber),
                    antipatterns: ($antipatterns | tonumber),
                    total_issues: (($syntax_errors | tonumber) + ($stale_owners | tonumber) + ($coverage_gaps | tonumber) + ($antipatterns | tonumber))
                },
                details: {
                    syntax: $syntax_details,
                    stale: $stale_details,
                    gaps: $gap_details,
                    antipatterns: $antipattern_details
                }
            }'
    else
        # Generate text report
        cat << EOF
========================================
CODEOWNERS Validation Report
========================================

File: $codeowners_file

Summary:
--------
Syntax Errors: $syntax_error_count
Stale Owners: $stale_owner_count
Coverage Gaps: $coverage_gap_count
Anti-patterns: $antipattern_count

$(if [[ $syntax_error_count -gt 0 ]]; then
    echo "Syntax Issues:"
    echo "-------------"
    cat "$syntax_errors"
    echo ""
fi)

$(if [[ $stale_owner_count -gt 0 ]]; then
    echo "Stale Owners (>90 days inactive):"
    echo "--------------------------------"
    cat "$stale_owners"
    echo ""
fi)

$(if [[ $coverage_gap_count -gt 0 ]]; then
    echo "Coverage Gaps (first 10 files without owners):"
    echo "---------------------------------------------"
    head -10 "$coverage_gaps"
    echo ""
fi)

$(if [[ $antipattern_count -gt 0 ]]; then
    echo "Anti-patterns Detected:"
    echo "----------------------"
    cat "$antipatterns"
    echo ""
fi)

Overall Status: $(if [[ $((syntax_error_count + stale_owner_count + coverage_gap_count + antipattern_count)) -eq 0 ]]; then echo "✓ PASS"; else echo "✗ ISSUES FOUND"; fi)

========================================
EOF
    fi

    # Cleanup
    rm -f "$syntax_errors" "$stale_owners" "$coverage_gaps" "$antipatterns"
}

# Export functions
export -f parse_codeowners
export -f validate_syntax
export -f check_stale_owners
export -f check_coverage_gaps
export -f detect_antipatterns
export -f generate_validation_report
