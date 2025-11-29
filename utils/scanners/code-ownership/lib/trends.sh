#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# 
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Historical Trend Tracking Library
# Track ownership changes over time and identify patterns
#
# Key Features:
# - Historical snapshot storage
# - Metric change tracking (ownership, bus factor, health)
# - Contributor activity trends (emerging, declining, churned)
# - Time-series analysis
# - Trend prediction (linear regression)
# - ASCII visualization
#############################################################################

# Default trend storage location
TRENDS_DIR="${TRENDS_DIR:-$HOME/.code-ownership/trends}"

# Initialize trends storage
init_trends() {
    local repo_path="$1"

    # Create trends directory structure
    local repo_hash=$(echo "$repo_path" | md5sum | cut -d' ' -f1 2>/dev/null || echo "$repo_path" | md5 | cut -d' ' -f1)
    local repo_trends_dir="$TRENDS_DIR/$repo_hash"

    mkdir -p "$repo_trends_dir/snapshots"
    mkdir -p "$repo_trends_dir/reports"

    # Store repo metadata
    cat > "$repo_trends_dir/metadata.json" << EOF
{
    "repository_path": "$repo_path",
    "first_snapshot": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "last_updated": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF

    echo "$repo_trends_dir"
}

# Get trends directory for repository
get_trends_dir() {
    local repo_path="$1"
    local repo_hash=$(echo "$repo_path" | md5sum | cut -d' ' -f1 2>/dev/null || echo "$repo_path" | md5 | cut -d' ' -f1)
    echo "$TRENDS_DIR/$repo_hash"
}

# Store snapshot of current analysis
store_snapshot() {
    local repo_path="$1"
    local analysis_json="$2"
    local snapshot_date="${3:-$(date -u +%Y-%m-%d)}"

    local trends_dir=$(get_trends_dir "$repo_path")

    # Create trends dir if doesn't exist
    if [[ ! -d "$trends_dir" ]]; then
        trends_dir=$(init_trends "$repo_path")
    fi

    # Store snapshot with timestamp
    local snapshot_file="$trends_dir/snapshots/${snapshot_date}.json"

    # Add snapshot metadata
    jq ". + {snapshot_date: \"$snapshot_date\"}" <<< "$analysis_json" > "$snapshot_file"

    # Update last_updated in metadata
    jq ".last_updated = \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"" \
        "$trends_dir/metadata.json" > "$trends_dir/metadata.json.tmp"
    mv "$trends_dir/metadata.json.tmp" "$trends_dir/metadata.json"

    echo "$snapshot_file"
}

# Get all snapshots for repository
list_snapshots() {
    local repo_path="$1"
    local trends_dir=$(get_trends_dir "$repo_path")

    if [[ ! -d "$trends_dir/snapshots" ]]; then
        return 1
    fi

    find "$trends_dir/snapshots" -name "*.json" -type f | sort
}

# Get snapshot by date
get_snapshot() {
    local repo_path="$1"
    local snapshot_date="$2"
    local trends_dir=$(get_trends_dir "$repo_path")

    local snapshot_file="$trends_dir/snapshots/${snapshot_date}.json"

    if [[ -f "$snapshot_file" ]]; then
        cat "$snapshot_file"
        return 0
    fi

    return 1
}

# Calculate changes between two snapshots
calculate_snapshot_diff() {
    local snapshot1="$1"
    local snapshot2="$2"

    # Extract key metrics from both snapshots
    local date1=$(jq -r '.snapshot_date' "$snapshot1")
    local date2=$(jq -r '.snapshot_date' "$snapshot2")

    local coverage1=$(jq -r '.ownership_health.coverage_percentage' "$snapshot1")
    local coverage2=$(jq -r '.ownership_health.coverage_percentage' "$snapshot2")

    local gini1=$(jq -r '.ownership_health.gini_coefficient' "$snapshot1")
    local gini2=$(jq -r '.ownership_health.gini_coefficient' "$snapshot2")

    local bus_factor1=$(jq -r '.ownership_health.bus_factor' "$snapshot1")
    local bus_factor2=$(jq -r '.ownership_health.bus_factor' "$snapshot2")

    local health1=$(jq -r '.ownership_health.health_score' "$snapshot1")
    local health2=$(jq -r '.ownership_health.health_score' "$snapshot2")

    # Calculate deltas
    local coverage_delta=$(echo "scale=2; $coverage2 - $coverage1" | bc -l)
    local gini_delta=$(echo "scale=4; $gini2 - $gini1" | bc -l)
    local bus_factor_delta=$(echo "$bus_factor2 - $bus_factor1" | bc)
    local health_delta=$(echo "scale=2; $health2 - $health1" | bc -l)

    # Calculate contributor changes
    local contributors1=$(jq -r '.repository_metrics.active_contributors' "$snapshot1")
    local contributors2=$(jq -r '.repository_metrics.active_contributors' "$snapshot2")
    local contributor_delta=$(echo "$contributors2 - $contributors1" | bc)

    # Build diff JSON
    jq -n \
        --arg date1 "$date1" \
        --arg date2 "$date2" \
        --arg coverage_delta "$coverage_delta" \
        --arg gini_delta "$gini_delta" \
        --arg bus_factor_delta "$bus_factor_delta" \
        --arg health_delta "$health_delta" \
        --arg contributor_delta "$contributor_delta" \
        '{
            from_date: $date1,
            to_date: $date2,
            changes: {
                coverage: {
                    from: ($coverage1 | tonumber),
                    to: ($coverage2 | tonumber),
                    delta: ($coverage_delta | tonumber)
                },
                gini_coefficient: {
                    from: ($gini1 | tonumber),
                    to: ($gini2 | tonumber),
                    delta: ($gini_delta | tonumber)
                },
                bus_factor: {
                    from: ($bus_factor1 | tonumber),
                    to: ($bus_factor2 | tonumber),
                    delta: ($bus_factor_delta | tonumber)
                },
                health_score: {
                    from: ($health1 | tonumber),
                    to: ($health2 | tonumber),
                    delta: ($health_delta | tonumber)
                },
                active_contributors: {
                    from: ($contributors1 | tonumber),
                    to: ($contributors2 | tonumber),
                    delta: ($contributor_delta | tonumber)
                }
            }
        }' \
        --arg coverage1 "$coverage1" \
        --arg coverage2 "$coverage2" \
        --arg gini1 "$gini1" \
        --arg gini2 "$gini2" \
        --arg bus_factor1 "$bus_factor1" \
        --arg bus_factor2 "$bus_factor2" \
        --arg health1 "$health1" \
        --arg health2 "$health2" \
        --arg contributors1 "$contributors1" \
        --arg contributors2 "$contributors2"
}

# Analyze contributor trends (emerging, declining, churned)
analyze_contributor_trends() {
    local repo_path="$1"
    local lookback_days="${2:-90}"

    local trends_dir=$(get_trends_dir "$repo_path")

    if [[ ! -d "$trends_dir/snapshots" ]]; then
        echo "[]"
        return 1
    fi

    # Get recent snapshots
    local snapshots=($(find "$trends_dir/snapshots" -name "*.json" -type f | sort | tail -5))

    if [[ ${#snapshots[@]} -lt 2 ]]; then
        echo "[]"
        return 1
    fi

    # Extract contributor data from each snapshot
    local temp_contributors=$(mktemp)

    for snapshot in "${snapshots[@]}"; do
        local date=$(jq -r '.snapshot_date' "$snapshot")
        jq -r --arg date "$date" '.contributors[] | "\(.email)|\($date)|\(.files_owned)"' "$snapshot" >> "$temp_contributors"
    done

    # Analyze trends
    awk -F'|' '
    {
        email = $1
        date = $2
        files = $3

        # Track files over time
        if (!(email in first_seen)) {
            first_seen[email] = date
            first_files[email] = files
        }
        last_seen[email] = date
        last_files[email] = files
        total_snapshots[email]++
    }
    END {
        for (email in first_seen) {
            delta = last_files[email] - first_files[email]
            rate = (total_snapshots[email] > 1) ? delta / (total_snapshots[email] - 1) : 0

            # Classify trend
            if (rate > 2) trend = "emerging"
            else if (rate < -2) trend = "declining"
            else if (last_files[email] == 0) trend = "churned"
            else trend = "stable"

            printf "{\"email\":\"%s\",\"trend\":\"%s\",\"file_delta\":%d,\"rate\":%.2f}\n",
                email, trend, delta, rate
        }
    }
    ' "$temp_contributors" | jq -s '.'

    rm -f "$temp_contributors"
}

# Calculate linear regression for metric prediction
calculate_trend_line() {
    local -a x_values=("$@")
    local n=${#x_values[@]}

    if [[ $n -lt 2 ]]; then
        echo "0|0"  # slope|intercept
        return
    fi

    # Calculate means
    local sum_x=0
    local sum_y=0
    for ((i=0; i<n; i++)); do
        sum_x=$((sum_x + i))
        sum_y=$(echo "scale=4; $sum_y + ${x_values[i]}" | bc -l)
    done
    local mean_x=$(echo "scale=4; $sum_x / $n" | bc -l)
    local mean_y=$(echo "scale=4; $sum_y / $n" | bc -l)

    # Calculate slope
    local numerator=0
    local denominator=0
    for ((i=0; i<n; i++)); do
        local x_diff=$(echo "scale=4; $i - $mean_x" | bc -l)
        local y_diff=$(echo "scale=4; ${x_values[i]} - $mean_y" | bc -l)
        numerator=$(echo "scale=4; $numerator + ($x_diff * $y_diff)" | bc -l)
        denominator=$(echo "scale=4; $denominator + ($x_diff * $x_diff)" | bc -l)
    done

    local slope=0
    if [[ $(echo "$denominator != 0" | bc -l) -eq 1 ]]; then
        slope=$(echo "scale=4; $numerator / $denominator" | bc -l)
    fi

    # Calculate intercept
    local intercept=$(echo "scale=4; $mean_y - ($slope * $mean_x)" | bc -l)

    echo "$slope|$intercept"
}

# Generate trend report
generate_trend_report() {
    local repo_path="$1"
    local format="${2:-json}"

    local trends_dir=$(get_trends_dir "$repo_path")

    if [[ ! -d "$trends_dir/snapshots" ]]; then
        echo "{\"error\": \"No historical data available\"}"
        return 1
    fi

    # Get all snapshots
    local snapshots=($(list_snapshots "$repo_path"))
    local snapshot_count=${#snapshots[@]}

    if [[ $snapshot_count -lt 2 ]]; then
        echo "{\"error\": \"Need at least 2 snapshots for trend analysis\"}"
        return 1
    fi

    # Extract time-series data
    local -a coverage_series
    local -a health_series
    local -a bus_factor_series
    local -a dates

    for snapshot in "${snapshots[@]}"; do
        dates+=($(jq -r '.snapshot_date' "$snapshot"))
        coverage_series+=($(jq -r '.ownership_health.coverage_percentage' "$snapshot"))
        health_series+=($(jq -r '.ownership_health.health_score' "$snapshot"))
        bus_factor_series+=($(jq -r '.ownership_health.bus_factor' "$snapshot"))
    done

    # Calculate trend lines
    local coverage_trend=$(calculate_trend_line "${coverage_series[@]}")
    local health_trend=$(calculate_trend_line "${health_series[@]}")

    IFS='|' read -r coverage_slope coverage_intercept <<< "$coverage_trend"
    IFS='|' read -r health_slope health_intercept <<< "$health_trend"

    # Analyze contributor trends
    local contributor_trends=$(analyze_contributor_trends "$repo_path")

    # Calculate diff between first and last
    local first_snapshot="${snapshots[0]}"
    local last_snapshot="${snapshots[-1]}"
    local snapshot_diff=$(calculate_snapshot_diff "$first_snapshot" "$last_snapshot")

    if [[ "$format" == "json" ]]; then
        jq -n \
            --arg snapshot_count "$snapshot_count" \
            --arg first_date "${dates[0]}" \
            --arg last_date "${dates[-1]}" \
            --arg coverage_slope "$coverage_slope" \
            --arg health_slope "$health_slope" \
            --argjson diff "$snapshot_diff" \
            --argjson contributor_trends "$contributor_trends" \
            '{
                snapshot_count: ($snapshot_count | tonumber),
                date_range: {
                    from: $first_date,
                    to: $last_date
                },
                trends: {
                    coverage: {
                        slope: ($coverage_slope | tonumber),
                        direction: (if ($coverage_slope | tonumber) > 0 then "improving" elif ($coverage_slope | tonumber) < 0 then "declining" else "stable" end)
                    },
                    health: {
                        slope: ($health_slope | tonumber),
                        direction: (if ($health_slope | tonumber) > 0 then "improving" elif ($health_slope | tonumber) < 0 then "declining" else "stable" end)
                    }
                },
                changes: $diff.changes,
                contributor_trends: $contributor_trends,
                summary: {
                    status: (
                        if ($health_slope | tonumber) > 1 then "Strong improvement"
                        elif ($health_slope | tonumber) > 0 then "Improving"
                        elif ($health_slope | tonumber) < -1 then "Declining"
                        else "Stable"
                        end
                    )
                }
            }'
    else
        # Text format
        cat << EOF
Historical Trend Report
======================

Date Range: ${dates[0]} to ${dates[-1]}
Snapshots: $snapshot_count

Trends:
-------
Coverage: $(if (( $(echo "$coverage_slope > 0" | bc -l) )); then echo "↑ Improving"; elif (( $(echo "$coverage_slope < 0" | bc -l) )); then echo "↓ Declining"; else echo "→ Stable"; fi) (slope: $coverage_slope)
Health:   $(if (( $(echo "$health_slope > 0" | bc -l) )); then echo "↑ Improving"; elif (( $(echo "$health_slope < 0" | bc -l) )); then echo "↓ Declining"; else echo "→ Stable"; fi) (slope: $health_slope)

Contributor Trends:
------------------
$(echo "$contributor_trends" | jq -r '.[] | "\(.email): \(.trend) (\(.file_delta) files)"')

EOF
    fi
}

# Create ASCII chart for metric over time
create_ascii_chart() {
    local metric_name="$1"
    shift
    local -a values=("$@")

    local n=${#values[@]}
    if [[ $n -eq 0 ]]; then
        return
    fi

    # Find min/max
    local min=${values[0]}
    local max=${values[0]}
    for val in "${values[@]}"; do
        if (( $(echo "$val < $min" | bc -l) )); then
            min=$val
        fi
        if (( $(echo "$val > $max" | bc -l) )); then
            max=$val
        fi
    done

    local range=$(echo "scale=4; $max - $min" | bc -l)
    if (( $(echo "$range == 0" | bc -l) )); then
        range=1
    fi

    # Chart height
    local height=10

    echo "$metric_name ($min - $max)"
    echo "────────────────────────────────────────"

    # Create chart
    for ((row=height; row>=0; row--)); do
        local threshold=$(echo "scale=4; $min + ($range * $row / $height)" | bc -l)
        printf "%6.1f │" "$threshold"

        for val in "${values[@]}"; do
            if (( $(echo "$val >= $threshold" | bc -l) )); then
                printf "█"
            else
                printf " "
            fi
        done
        echo ""
    done

    printf "       └"
    for ((i=0; i<n; i++)); do
        printf "─"
    done
    echo ""
}

# Export functions
export -f init_trends
export -f get_trends_dir
export -f store_snapshot
export -f list_snapshots
export -f get_snapshot
export -f calculate_snapshot_diff
export -f analyze_contributor_trends
export -f calculate_trend_line
export -f generate_trend_report
export -f create_ascii_chart
