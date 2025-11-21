#!/bin/bash
# Health Scoring Engine
# Copyright (c) 2024 Crash Override Inc
# SPDX-License-Identifier: GPL-3.0

set -euo pipefail

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
UTILS_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Load configuration
if [ -f "$UTILS_ROOT/lib/config-loader.sh" ]; then
    source "$UTILS_ROOT/lib/config-loader.sh"
    CONFIG=$(load_config "package-health-analysis")
else
    CONFIG="{}"
fi

# Load deps.dev client
source "$SCRIPT_DIR/deps-dev-client.sh"

# Get scoring weights from config
WEIGHT_OPENSSF=$(echo "$CONFIG" | jq -r '.package_health.health_score_weights.openssf // 0.30')
WEIGHT_MAINTENANCE=$(echo "$CONFIG" | jq -r '.package_health.health_score_weights.maintenance // 0.25')
WEIGHT_SECURITY=$(echo "$CONFIG" | jq -r '.package_health.health_score_weights.security // 0.25')
WEIGHT_FRESHNESS=$(echo "$CONFIG" | jq -r '.package_health.health_score_weights.freshness // 0.10')
WEIGHT_POPULARITY=$(echo "$CONFIG" | jq -r '.package_health.health_score_weights.popularity // 0.10')

# Get thresholds from config
THRESHOLD_EXCELLENT=$(echo "$CONFIG" | jq -r '.package_health.thresholds.excellent // 90')
THRESHOLD_GOOD=$(echo "$CONFIG" | jq -r '.package_health.thresholds.good // 75')
THRESHOLD_FAIR=$(echo "$CONFIG" | jq -r '.package_health.thresholds.fair // 60')
THRESHOLD_POOR=$(echo "$CONFIG" | jq -r '.package_health.thresholds.poor // 40')

# Calculate OpenSSF component score (0-100)
# Usage: calculate_openssf_score <openssf_score>
calculate_openssf_score() {
    local openssf_score=$1

    if [ "$openssf_score" = "null" ] || [ -z "$openssf_score" ]; then
        # No OpenSSF score available - return neutral score
        echo "50"
        return
    fi

    # OpenSSF score is 0-10, convert to 0-100
    echo "scale=2; $openssf_score * 10" | bc
}

# Calculate maintenance score (0-100)
# Usage: calculate_maintenance_score <package_json>
calculate_maintenance_score() {
    local package_json=$1

    local score=50  # Start neutral

    # Check if deprecated
    local deprecated=$(echo "$package_json" | jq -r '.deprecated // false')
    if [ "$deprecated" = "true" ]; then
        echo "0"
        return
    fi

    # Check for recent releases (if version history available)
    local versions_count=$(echo "$package_json" | jq -r '.versions | length // 0')
    if [ "$versions_count" -gt 0 ]; then
        # Get latest version date
        local latest_date=$(echo "$package_json" | jq -r '.versions[-1].publishedAt // ""')

        if [ -n "$latest_date" ]; then
            local days_since_update=$(calculate_days_since "$latest_date")

            if [ "$days_since_update" -lt 180 ]; then
                # Updated within 6 months - good
                score=85
            elif [ "$days_since_update" -lt 365 ]; then
                # Updated within 1 year - fair
                score=70
            elif [ "$days_since_update" -lt 730 ]; then
                # Updated within 2 years - poor
                score=50
            else
                # No updates in 2+ years - very poor
                score=30
            fi
        fi
    fi

    echo "$score"
}

# Calculate security score (0-100)
# Usage: calculate_security_score <version_json>
calculate_security_score() {
    local version_json=$1

    # Get vulnerability count
    local vuln_count=$(echo "$version_json" | jq -r '.advisories | length // 0')

    if [ "$vuln_count" -eq 0 ]; then
        echo "100"
    elif [ "$vuln_count" -eq 1 ]; then
        echo "75"
    elif [ "$vuln_count" -eq 2 ]; then
        echo "50"
    elif [ "$vuln_count" -eq 3 ]; then
        echo "25"
    else
        echo "0"
    fi
}

# Calculate freshness score (0-100)
# Usage: calculate_freshness_score <current_version> <latest_version>
calculate_freshness_score() {
    local current_version=$1
    local latest_version=$2

    if [ "$current_version" = "$latest_version" ]; then
        echo "100"
    else
        # Compare versions - simplified scoring
        # If major version behind: 0-25
        # If minor version behind: 50-75
        # If patch version behind: 75-90

        local current_major=$(echo "$current_version" | cut -d. -f1 | sed 's/[^0-9]//g')
        local current_minor=$(echo "$current_version" | cut -d. -f2 | sed 's/[^0-9]//g')
        local latest_major=$(echo "$latest_version" | cut -d. -f1 | sed 's/[^0-9]//g')
        local latest_minor=$(echo "$latest_version" | cut -d. -f2 | sed 's/[^0-9]//g')

        # Default to 0 if parsing fails
        current_major=${current_major:-0}
        current_minor=${current_minor:-0}
        latest_major=${latest_major:-0}
        latest_minor=${latest_minor:-0}

        if [ "$current_major" -lt "$latest_major" ]; then
            # Major version behind
            local diff=$((latest_major - current_major))
            if [ "$diff" -eq 1 ]; then
                echo "25"
            else
                echo "10"
            fi
        elif [ "$current_minor" -lt "$latest_minor" ]; then
            # Minor version behind
            local diff=$((latest_minor - current_minor))
            if [ "$diff" -eq 1 ]; then
                echo "75"
            elif [ "$diff" -eq 2 ]; then
                echo "65"
            else
                echo "50"
            fi
        else
            # Patch version behind
            echo "90"
        fi
    fi
}

# Calculate popularity score (0-100)
# Usage: calculate_popularity_score <dependent_count>
calculate_popularity_score() {
    local dependent_count=$1

    # Logarithmic scale for popularity
    if [ "$dependent_count" -ge 10000 ]; then
        echo "100"
    elif [ "$dependent_count" -ge 5000 ]; then
        echo "90"
    elif [ "$dependent_count" -ge 1000 ]; then
        echo "80"
    elif [ "$dependent_count" -ge 500 ]; then
        echo "70"
    elif [ "$dependent_count" -ge 100 ]; then
        echo "60"
    elif [ "$dependent_count" -ge 50 ]; then
        echo "50"
    elif [ "$dependent_count" -ge 10 ]; then
        echo "40"
    else
        echo "30"
    fi
}

# Calculate days since a date
# Usage: calculate_days_since <iso8601_date>
calculate_days_since() {
    local date_string=$1

    if [ -z "$date_string" ]; then
        echo "9999"
        return
    fi

    # Parse date and calculate days
    local target_epoch=$(date -j -f "%Y-%m-%dT%H:%M:%S" "$(echo "$date_string" | cut -d'.' -f1)" "+%s" 2>/dev/null || echo "0")
    local now_epoch=$(date "+%s")

    if [ "$target_epoch" -eq 0 ]; then
        echo "9999"
        return
    fi

    local diff_seconds=$((now_epoch - target_epoch))
    local diff_days=$((diff_seconds / 86400))

    echo "$diff_days"
}

# Calculate composite health score
# Usage: calculate_health_score <package_summary> <version_info> <current_version>
calculate_health_score() {
    local package_summary=$1
    local version_info=$2
    local current_version=$3

    # Extract component scores
    local openssf_raw=$(echo "$package_summary" | jq -r '.openssf_score // null')
    local openssf_score=$(calculate_openssf_score "$openssf_raw")

    local maintenance_score=$(calculate_maintenance_score "$package_summary")

    local security_score=$(calculate_security_score "$version_info")

    local latest_version=$(echo "$package_summary" | jq -r '.latest_version // "unknown"')
    local freshness_score=$(calculate_freshness_score "$current_version" "$latest_version")

    local dependent_count=$(echo "$package_summary" | jq -r '.dependent_count // 0')
    local popularity_score=$(calculate_popularity_score "$dependent_count")

    # Calculate weighted score
    local composite_score=$(echo "scale=2; \
        ($openssf_score * $WEIGHT_OPENSSF) + \
        ($maintenance_score * $WEIGHT_MAINTENANCE) + \
        ($security_score * $WEIGHT_SECURITY) + \
        ($freshness_score * $WEIGHT_FRESHNESS) + \
        ($popularity_score * $WEIGHT_POPULARITY)" | bc)

    # Round to integer
    printf "%.0f" "$composite_score"
}

# Get health grade from score
# Usage: get_health_grade <score>
get_health_grade() {
    local score=$1

    if [ "$score" -ge "$THRESHOLD_EXCELLENT" ]; then
        echo "Excellent"
    elif [ "$score" -ge "$THRESHOLD_GOOD" ]; then
        echo "Good"
    elif [ "$score" -ge "$THRESHOLD_FAIR" ]; then
        echo "Fair"
    elif [ "$score" -ge "$THRESHOLD_POOR" ]; then
        echo "Poor"
    else
        echo "Critical"
    fi
}

# Get health grade color (for display)
# Usage: get_health_color <grade>
get_health_color() {
    local grade=$1

    case $grade in
        "Excellent")
            echo "green"
            ;;
        "Good")
            echo "blue"
            ;;
        "Fair")
            echo "yellow"
            ;;
        "Poor")
            echo "orange"
            ;;
        "Critical")
            echo "red"
            ;;
        *)
            echo "gray"
            ;;
    esac
}

# Analyze single package health
# Usage: analyze_package_health <system> <package> <version>
analyze_package_health() {
    local system=$1
    local package=$2
    local version=$3

    # Get package summary
    local package_summary=$(get_package_summary "$system" "$package")

    if echo "$package_summary" | jq -e '.error' > /dev/null 2>&1; then
        local error=$(echo "$package_summary" | jq -r '.error')
        echo "{\"error\": \"$error\", \"package\": \"$package\", \"system\": \"$system\"}"
        return 1
    fi

    # Get version info
    local version_info=$(get_package_version "$system" "$package" "$version")

    # Calculate health score
    local health_score=$(calculate_health_score "$package_summary" "$version_info" "$version")
    local health_grade=$(get_health_grade "$health_score")

    # Get individual component scores for details
    local openssf_raw=$(echo "$package_summary" | jq -r '.openssf_score // null')
    local openssf_score=$(calculate_openssf_score "$openssf_raw")
    local maintenance_score=$(calculate_maintenance_score "$package_summary")
    local security_score=$(calculate_security_score "$version_info")
    local latest_version=$(echo "$package_summary" | jq -r '.latest_version')
    local freshness_score=$(calculate_freshness_score "$version" "$latest_version")
    local dependent_count=$(echo "$package_summary" | jq -r '.dependent_count')
    local popularity_score=$(calculate_popularity_score "$dependent_count")

    # Build result JSON
    jq -n \
        --arg name "$package" \
        --arg system "$system" \
        --arg version "$version" \
        --argjson score "$health_score" \
        --arg grade "$health_grade" \
        --argjson openssf "$openssf_score" \
        --argjson openssf_raw "$openssf_raw" \
        --argjson maintenance "$maintenance_score" \
        --argjson security "$security_score" \
        --argjson freshness "$freshness_score" \
        --argjson popularity "$popularity_score" \
        --arg latest "$latest_version" \
        --argjson deprecated "$(echo "$package_summary" | jq -r '.deprecated')" \
        --arg deprecation_msg "$(echo "$package_summary" | jq -r '.deprecation_message // ""')" \
        --argjson dependent_count "$dependent_count" \
        '{
            package: $name,
            system: $system,
            version: $version,
            health_score: $score,
            health_grade: $grade,
            component_scores: {
                openssf: $openssf,
                openssf_raw: $openssf_raw,
                maintenance: $maintenance,
                security: $security,
                freshness: $freshness,
                popularity: $popularity
            },
            latest_version: $latest,
            deprecated: $deprecated,
            deprecation_message: $deprecation_msg,
            dependent_count: $dependent_count
        }'
}

# Export functions
export -f calculate_openssf_score
export -f calculate_maintenance_score
export -f calculate_security_score
export -f calculate_freshness_score
export -f calculate_popularity_score
export -f calculate_days_since
export -f calculate_health_score
export -f get_health_grade
export -f get_health_color
export -f analyze_package_health
