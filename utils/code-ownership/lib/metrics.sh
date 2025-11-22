#!/bin/bash
# Copyright (c) 2024 Gibson Powers Contributors
# 
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Code Ownership Metrics Library
# Research-backed metrics for code ownership analysis
# Based on 2024 arXiv research and industry best practices
#############################################################################

# Calculate recency factor using exponential decay
# Formula: e^(-ln(2) * days_since_last_commit / half_life)
# Default half-life: 90 days
calculate_recency_factor() {
    local days_since="$1"
    local half_life="${2:-90}"

    # Using bc for floating point calculation
    # recency_factor = e^(-ln(2) * days / half_life)
    echo "scale=4; e(-l(2) * $days_since / $half_life)" | bc -l
}

# Calculate Gini coefficient for ownership distribution
# Measures concentration/inequality of ownership
# Returns value between 0 (perfect equality) and 1 (perfect inequality)
calculate_gini_coefficient() {
    local -a values=("$@")
    local n=${#values[@]}

    if [[ $n -eq 0 ]]; then
        echo "0"
        return
    fi

    # Sort values
    IFS=$'\n' sorted=($(sort -n <<<"${values[*]}"))
    unset IFS

    # Calculate mean
    local sum=0
    for val in "${sorted[@]}"; do
        sum=$((sum + val))
    done
    local mean=$((sum / n))

    if [[ $mean -eq 0 ]]; then
        echo "0"
        return
    fi

    # Calculate Gini
    local numerator=0
    for ((i=0; i<n; i++)); do
        for ((j=0; j<n; j++)); do
            local diff=$((sorted[i] - sorted[j]))
            numerator=$((numerator + (diff < 0 ? -diff : diff)))
        done
    done

    local denominator=$((2 * n * n * mean))
    echo "scale=4; $numerator / $denominator" | bc -l
}

# Calculate bus factor
# Returns the minimum number of people who need to leave before project stalls
calculate_bus_factor() {
    local total_files="$1"
    shift
    local -a owner_file_counts=("$@")

    # Sort by file count (descending)
    IFS=$'\n' sorted=($(sort -rn <<<"${owner_file_counts[*]}"))
    unset IFS

    local removed_count=0
    local uncovered_files=0

    # Simulate removing contributors one by one
    for count in "${sorted[@]}"; do
        ((removed_count++))
        uncovered_files=$((uncovered_files + count))

        # Check if >20% of files now have no expert owner
        local uncovered_percentage=$(echo "scale=2; $uncovered_files / $total_files" | bc -l)
        local threshold=$(echo "scale=2; $uncovered_percentage > 0.20" | bc -l)

        if [[ "$threshold" == "1" ]]; then
            echo "$removed_count"
            return
        fi
    done

    # If we get here, bus factor is the number of contributors
    echo "${#sorted[@]}"
}

# Calculate overall ownership health score (0-100)
# Components: coverage, distribution, freshness, engagement
calculate_health_score() {
    local coverage="$1"        # Ownership coverage percentage (0-100)
    local gini="$2"            # Gini coefficient (0-1)
    local freshness="$3"       # Active owners percentage (0-100)
    local engagement="$4"      # Responsive owners percentage (0-100)

    # Weights (must sum to 1.0)
    local coverage_weight="0.35"
    local distribution_weight="0.25"
    local freshness_weight="0.20"
    local engagement_weight="0.20"

    # Convert Gini to distribution score (invert, so lower Gini = higher score)
    local distribution=$(echo "scale=2; (1 - $gini) * 100" | bc -l)

    # Calculate weighted score
    echo "scale=2; ($coverage * $coverage_weight) + ($distribution * $distribution_weight) + ($freshness * $freshness_weight) + ($engagement * $engagement_weight)" | bc -l
}

# Get health grade from score
get_health_grade() {
    local score="$1"

    if (( $(echo "$score >= 85" | bc -l) )); then
        echo "Excellent"
    elif (( $(echo "$score >= 70" | bc -l) )); then
        echo "Good"
    elif (( $(echo "$score >= 50" | bc -l) )); then
        echo "Fair"
    else
        echo "Poor"
    fi
}

# Calculate top-N concentration
# Returns percentage of files owned by top N contributors
calculate_top_n_concentration() {
    local total_files="$1"
    local n="$2"
    shift 2
    local -a owner_file_counts=("$@")

    # Sort by file count (descending)
    IFS=$'\n' sorted=($(sort -rn <<<"${owner_file_counts[*]}"))
    unset IFS

    local top_n_files=0
    local count=0

    for files in "${sorted[@]}"; do
        if [[ $count -lt $n ]]; then
            top_n_files=$((top_n_files + files))
            ((count++))
        else
            break
        fi
    done

    echo "scale=2; ($top_n_files / $total_files) * 100" | bc -l
}

# Assess staleness category
get_staleness_category() {
    local days_since="$1"

    if [[ $days_since -lt 30 ]]; then
        echo "Active"
    elif [[ $days_since -lt 60 ]]; then
        echo "Recent"
    elif [[ $days_since -lt 90 ]]; then
        echo "Stale"
    elif [[ $days_since -lt 180 ]]; then
        echo "Inactive"
    else
        echo "Abandoned"
    fi
}

# Calculate commit frequency score for a file/contributor
# Returns normalized score 0-100 based on commit count
calculate_commit_frequency_score() {
    local commits="$1"
    local max_commits="$2"

    if [[ $max_commits -eq 0 ]]; then
        echo "0"
        return
    fi

    # Normalize to 0-100 scale
    echo "scale=2; ($commits / $max_commits) * 100" | bc -l
}

# Calculate lines contributed score
# Returns normalized score 0-100 based on lines changed
calculate_lines_score() {
    local lines_changed="$1"
    local max_lines="$2"

    if [[ $max_lines -eq 0 ]]; then
        echo "0"
        return
    fi

    # Normalize to 0-100 scale with logarithmic scaling for very large changes
    local normalized=$(echo "scale=2; ($lines_changed / $max_lines) * 100" | bc -l)

    # Cap at 100
    if (( $(echo "$normalized > 100" | bc -l) )); then
        echo "100"
    else
        echo "$normalized"
    fi
}

# Calculate review participation score
# Based on number of reviews given and received
calculate_review_participation_score() {
    local reviews_given="$1"
    local reviews_received="$2"
    local max_reviews_given="$3"
    local max_reviews_received="$4"

    local given_score=0
    local received_score=0

    if [[ $max_reviews_given -gt 0 ]]; then
        given_score=$(echo "scale=2; ($reviews_given / $max_reviews_given) * 50" | bc -l)
    fi

    if [[ $max_reviews_received -gt 0 ]]; then
        received_score=$(echo "scale=2; ($reviews_received / $max_reviews_received) * 50" | bc -l)
    fi

    # Combined score (50% for giving reviews, 50% for receiving)
    echo "scale=2; $given_score + $received_score" | bc -l
}

# Calculate consistency score based on commit distribution over time
# Returns 0-1 where 1 is perfectly consistent, 0 is highly sporadic
calculate_consistency_score() {
    local -a commit_dates=("$@")
    local n=${#commit_dates[@]}

    if [[ $n -lt 2 ]]; then
        echo "0"
        return
    fi

    # Calculate standard deviation of time gaps between commits
    # Lower std dev = more consistent = higher score

    # Calculate gaps
    local -a gaps=()
    local prev_date="${commit_dates[0]}"

    for ((i=1; i<n; i++)); do
        local current_date="${commit_dates[i]}"
        local gap=$(( $(date -j -f "%Y-%m-%d" "$current_date" +%s 2>/dev/null || date -d "$current_date" +%s) - $(date -j -f "%Y-%m-%d" "$prev_date" +%s 2>/dev/null || date -d "$prev_date" +%s) ))
        gaps+=("$gap")
        prev_date="$current_date"
    done

    # Calculate mean gap
    local sum=0
    for gap in "${gaps[@]}"; do
        sum=$((sum + gap))
    done
    local mean=$((sum / ${#gaps[@]}))

    # Calculate standard deviation
    local variance_sum=0
    for gap in "${gaps[@]}"; do
        local diff=$((gap - mean))
        variance_sum=$((variance_sum + diff * diff))
    done
    local std_dev=$(echo "scale=2; sqrt($variance_sum / ${#gaps[@]})" | bc -l)

    # Convert to consistency score (0-1)
    # Lower std dev relative to mean = higher consistency
    local coefficient_variation=$(echo "scale=4; $std_dev / $mean" | bc -l)

    # Invert and normalize (assuming CV > 2.0 is very inconsistent)
    local consistency=$(echo "scale=4; 1 - ($coefficient_variation / 2)" | bc -l)

    # Clamp to 0-1
    if (( $(echo "$consistency < 0" | bc -l) )); then
        echo "0"
    elif (( $(echo "$consistency > 1" | bc -l) )); then
        echo "1"
    else
        echo "$consistency"
    fi
}

# Calculate ownership score using enhanced 5-component formula
# Based on 2024 research findings
calculate_ownership_score() {
    local commit_freq="$1"      # Commit frequency score (0-100)
    local lines_contrib="$2"    # Lines contributed score (0-100)
    local review_partic="$3"    # Review participation score (0-100)
    local recency="$4"          # Recency factor (0-1)
    local consistency="$5"      # Consistency score (0-1)

    # Weights (sum to 1.0)
    echo "scale=2; ($commit_freq * 0.30) + ($lines_contrib * 0.20) + ($review_partic * 0.25) + ($recency * 100 * 0.15) + ($consistency * 100 * 0.10)" | bc -l
}

# Calculate comprehensive ownership metrics for a contributor to a file
# Returns: commit_score|lines_score|review_score|recency_factor|consistency_score|total_score
calculate_comprehensive_ownership() {
    local repo_path="$1"
    local file_path="$2"
    local contributor_email="$3"
    local since_date="$4"

    cd "$repo_path" || return 1

    # Get commit count
    local commits=$(git log --since="$since_date" --author="$contributor_email" --oneline -- "$file_path" 2>/dev/null | wc -l | tr -d ' ')

    # Get max commits by anyone to this file
    local max_commits=$(git log --since="$since_date" --format="%ae" -- "$file_path" 2>/dev/null | sort | uniq -c | sort -rn | head -1 | awk '{print $1}')
    max_commits="${max_commits:-1}"

    # Get lines changed
    local lines_changed=$(git log --since="$since_date" --author="$contributor_email" --numstat -- "$file_path" 2>/dev/null | awk '{sum+=$1+$2} END {print sum+0}')

    # Get max lines by anyone
    local max_lines=$(git log --since="$since_date" --numstat -- "$file_path" 2>/dev/null | awk -v email="$contributor_email" '{lines[$4]+=$1+$2} END {for(e in lines) if(lines[e]>max) max=lines[e]; print max+0}')
    max_lines="${max_lines:-1}"

    # Get last commit date for recency
    local last_commit=$(git log --since="$since_date" --author="$contributor_email" --format="%ad" --date=short -- "$file_path" 2>/dev/null | head -1)
    local days_since=90
    if [[ -n "$last_commit" ]]; then
        days_since=$(( ($(date +%s) - $(date -j -f "%Y-%m-%d" "$last_commit" +%s 2>/dev/null || date -d "$last_commit" +%s)) / 86400 ))
    fi

    # Get commit dates for consistency
    mapfile -t commit_dates < <(git log --since="$since_date" --author="$contributor_email" --format="%ad" --date=short -- "$file_path" 2>/dev/null)

    # Calculate components
    local commit_score=$(calculate_commit_frequency_score "$commits" "$max_commits")
    local lines_score=$(calculate_lines_score "$lines_changed" "$max_lines")
    local review_score="50"  # Placeholder - would need GitHub API integration
    local recency=$(calculate_recency_factor "$days_since")
    local consistency=$(calculate_consistency_score "${commit_dates[@]}")

    # Calculate total ownership score
    local total=$(calculate_ownership_score "$commit_score" "$lines_score" "$review_score" "$recency" "$consistency")

    echo "$commit_score|$lines_score|$review_score|$recency|$consistency|$total"
}

# Export functions for use in other scripts
export -f calculate_recency_factor
export -f calculate_gini_coefficient
export -f calculate_bus_factor
export -f calculate_health_score
export -f get_health_grade
export -f calculate_top_n_concentration
export -f get_staleness_category
export -f calculate_commit_frequency_score
export -f calculate_lines_score
export -f calculate_review_participation_score
export -f calculate_consistency_score
export -f calculate_ownership_score
export -f calculate_comprehensive_ownership
