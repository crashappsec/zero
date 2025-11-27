#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Severity Classifier Library
# Functions for classifying and filtering security finding severity
#############################################################################

# Severity levels (ordered from lowest to highest)
declare -A SEVERITY_LEVELS=(
    ["info"]=0
    ["low"]=1
    ["medium"]=2
    ["high"]=3
    ["critical"]=4
)

# Severity colors for output
declare -A SEVERITY_COLORS=(
    ["info"]='\033[0;37m'      # Gray
    ["low"]='\033[0;32m'       # Green
    ["medium"]='\033[0;33m'    # Yellow
    ["high"]='\033[0;31m'      # Red
    ["critical"]='\033[1;31m'  # Bold Red
)

NC='\033[0m'

# Get numeric severity level
# Usage: get_severity_level <severity>
get_severity_level() {
    local severity="${1,,}"  # lowercase
    echo "${SEVERITY_LEVELS[$severity]:-0}"
}

# Compare severity levels
# Usage: compare_severity <severity1> <severity2>
# Returns: 0 if equal, 1 if first > second, 2 if first < second
compare_severity() {
    local level1
    level1=$(get_severity_level "$1")
    local level2
    level2=$(get_severity_level "$2")

    if [[ "$level1" -eq "$level2" ]]; then
        echo 0
    elif [[ "$level1" -gt "$level2" ]]; then
        echo 1
    else
        echo 2
    fi
}

# Check if severity meets minimum threshold
# Usage: meets_severity_threshold <severity> <minimum>
meets_severity_threshold() {
    local severity="$1"
    local minimum="$2"

    local level
    level=$(get_severity_level "$severity")
    local min_level
    min_level=$(get_severity_level "$minimum")

    [[ "$level" -ge "$min_level" ]]
}

# Filter findings by minimum severity (expects JSON input)
# Usage: echo "$findings_json" | filter_by_severity <minimum>
filter_by_severity() {
    local minimum="$1"
    local min_level
    min_level=$(get_severity_level "$minimum")

    jq --argjson min "$min_level" '
        def severity_level:
            if . == "critical" then 4
            elif . == "high" then 3
            elif . == "medium" then 2
            elif . == "low" then 1
            else 0
            end;
        [.[] | select((.severity | ascii_downcase | severity_level) >= $min)]
    '
}

# Sort findings by severity (highest first)
# Usage: echo "$findings_json" | sort_by_severity
sort_by_severity() {
    jq '
        def severity_level:
            if . == "critical" then 4
            elif . == "high" then 3
            elif . == "medium" then 2
            elif . == "low" then 1
            else 0
            end;
        sort_by(-(.severity | ascii_downcase | severity_level))
    '
}

# Count findings by severity
# Usage: echo "$findings_json" | count_by_severity
count_by_severity() {
    jq '
        group_by(.severity | ascii_downcase) |
        map({
            severity: .[0].severity,
            count: length
        }) |
        sort_by(
            if .severity == "critical" then 0
            elif .severity == "high" then 1
            elif .severity == "medium" then 2
            elif .severity == "low" then 3
            else 4
            end
        )
    '
}

# Get severity summary as text
# Usage: echo "$findings_json" | get_severity_summary
get_severity_summary() {
    local findings
    findings=$(cat)

    local critical high medium low

    critical=$(echo "$findings" | jq '[.[] | select(.severity == "critical" or .severity == "CRITICAL")] | length')
    high=$(echo "$findings" | jq '[.[] | select(.severity == "high" or .severity == "HIGH")] | length')
    medium=$(echo "$findings" | jq '[.[] | select(.severity == "medium" or .severity == "MEDIUM")] | length')
    low=$(echo "$findings" | jq '[.[] | select(.severity == "low" or .severity == "LOW")] | length')

    echo "Severity Summary:"
    echo -e "  ${SEVERITY_COLORS[critical]}Critical: $critical${NC}"
    echo -e "  ${SEVERITY_COLORS[high]}High:     $high${NC}"
    echo -e "  ${SEVERITY_COLORS[medium]}Medium:   $medium${NC}"
    echo -e "  ${SEVERITY_COLORS[low]}Low:      $low${NC}"
}

# Print colored severity label
# Usage: print_severity <severity>
print_severity() {
    local severity="${1,,}"
    local color="${SEVERITY_COLORS[$severity]:-$NC}"
    echo -e "${color}${severity^^}${NC}"
}

# Calculate overall risk score from findings
# Usage: echo "$findings_json" | calculate_risk_score
calculate_risk_score() {
    jq '
        def severity_weight:
            if . == "critical" or . == "CRITICAL" then 10
            elif . == "high" or . == "HIGH" then 5
            elif . == "medium" or . == "MEDIUM" then 2
            elif . == "low" or . == "LOW" then 1
            else 0
            end;

        if length == 0 then 0
        else
            (map(.severity | severity_weight) | add) / length * 10 | floor
        end
    '
}

# Get risk level from score
# Usage: get_risk_level <score>
get_risk_level() {
    local score="$1"

    if [[ "$score" -ge 80 ]]; then
        echo "critical"
    elif [[ "$score" -ge 50 ]]; then
        echo "high"
    elif [[ "$score" -ge 25 ]]; then
        echo "medium"
    elif [[ "$score" -gt 0 ]]; then
        echo "low"
    else
        echo "none"
    fi
}

# Check if findings warrant failure (for CI/CD)
# Usage: echo "$findings_json" | should_fail <fail_on_severity>
should_fail() {
    local fail_on="$1"
    local findings
    findings=$(cat)

    local count
    count=$(echo "$findings" | filter_by_severity "$fail_on" | jq 'length')

    [[ "$count" -gt 0 ]]
}

# CVSS to severity mapping
cvss_to_severity() {
    local cvss="$1"

    # Handle both integer and decimal CVSS scores
    local score
    score=$(echo "$cvss" | cut -d. -f1)

    if [[ "$score" -ge 9 ]]; then
        echo "critical"
    elif [[ "$score" -ge 7 ]]; then
        echo "high"
    elif [[ "$score" -ge 4 ]]; then
        echo "medium"
    elif [[ "$score" -ge 1 ]]; then
        echo "low"
    else
        echo "info"
    fi
}

# Confidence adjustment
# Adjusts severity based on confidence level
# Usage: adjust_severity_by_confidence <severity> <confidence>
adjust_severity_by_confidence() {
    local severity="$1"
    local confidence="${2,,}"

    # Low confidence downgrades severity by one level
    if [[ "$confidence" == "low" ]]; then
        case "$severity" in
            critical) echo "high" ;;
            high) echo "medium" ;;
            medium) echo "low" ;;
            *) echo "$severity" ;;
        esac
    else
        echo "$severity"
    fi
}
