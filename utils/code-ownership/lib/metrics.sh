#!/bin/bash
# Copyright (c) 2024 Crash Override Inc
# 101 Fulton St, 416, New York 10038
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

# Export functions for use in other scripts
export -f calculate_recency_factor
export -f calculate_gini_coefficient
export -f calculate_bus_factor
export -f calculate_health_score
export -f get_health_grade
export -f calculate_top_n_concentration
export -f get_staleness_category
export -f calculate_ownership_score
