#!/bin/bash
# Abandoned Package Detector
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0
#
# Identifies packages that are no longer actively maintained, posing security risks.
# Part of the Security & Risk Management module.

set -eo pipefail

# Get script directory for loading shared libraries
LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCANNER_DIR="$(dirname "$LIB_DIR")"
SCANNERS_ROOT="$(dirname "$SCANNER_DIR")"

# Load deps.dev client from shared libs (if not already loaded)
if ! command -v deps_dev_get_package_info &> /dev/null; then
    source "$SCANNERS_ROOT/shared/lib/deps-dev-client.sh"
fi

#############################################################################
# Configuration
#############################################################################

# Thresholds for abandonment detection (in days)
ABANDONED_THRESHOLD_DAYS=${ABANDONED_THRESHOLD_DAYS:-730}   # 2 years
STALE_THRESHOLD_DAYS=${STALE_THRESHOLD_DAYS:-365}           # 1 year
WARNING_THRESHOLD_DAYS=${WARNING_THRESHOLD_DAYS:-180}       # 6 months

# Minimum maintainer count for health
MIN_MAINTAINER_COUNT=${MIN_MAINTAINER_COUNT:-2}

#############################################################################
# Date Utility Functions
#############################################################################

# Calculate days since a given date
# Usage: days_since_date <ISO8601_date>
days_since_date() {
    local date_str="$1"

    if [[ -z "$date_str" || "$date_str" == "null" ]]; then
        echo "999999"  # Unknown date = very old
        return
    fi

    # Extract date part (handle various formats)
    date_str="${date_str%%T*}"  # Remove time portion

    # Calculate using Python for reliability across platforms
    python3 -c "
from datetime import datetime, date
try:
    d = datetime.fromisoformat('$date_str'.replace('Z', '+00:00'))
    days = (datetime.now(d.tzinfo) - d).days if d.tzinfo else (datetime.now() - d).days
    print(max(0, days))
except:
    # Try parsing just the date part
    try:
        d = datetime.strptime('$date_str'[:10], '%Y-%m-%d')
        days = (datetime.now() - d).days
        print(max(0, days))
    except:
        print(999999)
" 2>/dev/null || echo "999999"
}

#############################################################################
# Abandonment Detection Functions
#############################################################################

# Check if a package is explicitly deprecated
# Usage: check_deprecated <package> <ecosystem>
check_deprecated() {
    local pkg="$1"
    local ecosystem="$2"

    # Get package info from deps.dev
    local pkg_info=$(get_package_info "$ecosystem" "$pkg" 2>/dev/null)

    if [[ -z "$pkg_info" || "$pkg_info" == *"error"* ]]; then
        echo "false"
        return
    fi

    # Check deprecated flag
    local deprecated=$(echo "$pkg_info" | jq -r '.deprecated // false')
    echo "$deprecated"
}

# Get deprecation message if deprecated
# Usage: get_deprecation_reason <package> <ecosystem>
get_deprecation_reason() {
    local pkg="$1"
    local ecosystem="$2"

    local pkg_info=$(get_package_info "$ecosystem" "$pkg" 2>/dev/null)

    if [[ -z "$pkg_info" || "$pkg_info" == *"error"* ]]; then
        echo ""
        return
    fi

    echo "$pkg_info" | jq -r '.deprecationMessage // ""'
}

# Check if repository is archived
# Usage: check_archived <package> <ecosystem>
check_archived() {
    local pkg="$1"
    local ecosystem="$2"

    # Get package info to find repository URL
    local pkg_info=$(get_package_info "$ecosystem" "$pkg" 2>/dev/null)

    if [[ -z "$pkg_info" || "$pkg_info" == *"error"* ]]; then
        echo "false"
        return
    fi

    # Try to get project info which may have archived status
    local project_key=$(echo "$pkg_info" | jq -r '.projectKey // empty')

    if [[ -n "$project_key" ]]; then
        local project_info=$(get_project_info "$project_key" 2>/dev/null)
        if [[ -n "$project_info" && "$project_info" != *"error"* ]]; then
            # Check OpenSSF scorecard for archived status
            local maintained=$(echo "$project_info" | jq -r '.scorecard.checks[]? | select(.name=="Maintained") | .score // 0')
            if [[ "$maintained" == "0" ]]; then
                echo "true"
                return
            fi
        fi
    fi

    echo "false"
}

# Get last update date for a package
# Usage: get_last_update <package> <ecosystem>
get_last_update() {
    local pkg="$1"
    local ecosystem="$2"

    # Get package info
    local pkg_info=$(get_package_info "$ecosystem" "$pkg" 2>/dev/null)

    if [[ -z "$pkg_info" || "$pkg_info" == *"error"* ]]; then
        echo ""
        return
    fi

    # Get the latest version's publish date
    local latest_version=$(echo "$pkg_info" | jq -r '.versions[-1].versionKey.version // ""')

    if [[ -n "$latest_version" ]]; then
        local version_info=$(get_package_version "$ecosystem" "$pkg" "$latest_version" 2>/dev/null)
        if [[ -n "$version_info" && "$version_info" != *"error"* ]]; then
            echo "$version_info" | jq -r '.publishedAt // ""'
            return
        fi
    fi

    echo ""
}

# Get OpenSSF Scorecard maintenance info
# Usage: get_maintenance_metrics <package> <ecosystem>
get_maintenance_metrics() {
    local pkg="$1"
    local ecosystem="$2"

    local pkg_info=$(get_package_info "$ecosystem" "$pkg" 2>/dev/null)

    if [[ -z "$pkg_info" || "$pkg_info" == *"error"* ]]; then
        echo '{"error": "package_not_found"}'
        return
    fi

    # Extract scorecard data
    local scorecard_score=$(echo "$pkg_info" | jq -r '.scorecard.score // null')
    local scorecard_date=$(echo "$pkg_info" | jq -r '.scorecard.date // null')

    # Extract specific checks (use null if empty)
    local maintained_score=$(echo "$pkg_info" | jq -r '.scorecard.checks[]? | select(.name=="Maintained") | .score // null')
    local code_review_score=$(echo "$pkg_info" | jq -r '.scorecard.checks[]? | select(.name=="Code-Review") | .score // null')
    local branch_protection=$(echo "$pkg_info" | jq -r '.scorecard.checks[]? | select(.name=="Branch-Protection") | .score // null')

    # Get dependent count as popularity metric
    local dependent_count=$(echo "$pkg_info" | jq -r '.dependentCount // 0')

    # Ensure values are valid JSON (use null for empty/missing)
    [[ -z "$scorecard_score" || "$scorecard_score" == "null" ]] && scorecard_score="null"
    [[ -z "$maintained_score" ]] && maintained_score="null"
    [[ -z "$code_review_score" ]] && code_review_score="null"
    [[ -z "$branch_protection" ]] && branch_protection="null"
    [[ -z "$dependent_count" ]] && dependent_count="0"

    echo "{
        \"openssf_score\": $scorecard_score,
        \"scorecard_date\": \"$scorecard_date\",
        \"maintained_score\": $maintained_score,
        \"code_review_score\": $code_review_score,
        \"branch_protection_score\": $branch_protection,
        \"dependent_count\": $dependent_count
    }" | jq '.'
}

# Main abandonment status check
# Usage: check_abandonment_status <package> <ecosystem>
# Returns: JSON with status, risk_factors, and recommendations
check_abandonment_status() {
    local pkg="$1"
    local ecosystem="$2"

    local status="healthy"
    local risk_factors=()
    local risk_level="low"
    local recommendations=()

    # Check if explicitly deprecated
    local is_deprecated=$(check_deprecated "$pkg" "$ecosystem")
    if [[ "$is_deprecated" == "true" ]]; then
        status="deprecated"
        risk_factors+=("explicitly_deprecated")
        risk_level="high"
        local deprecation_msg=$(get_deprecation_reason "$pkg" "$ecosystem")
        if [[ -n "$deprecation_msg" ]]; then
            recommendations+=("Deprecation message: $deprecation_msg")
        fi
        recommendations+=("Find and migrate to recommended alternative")
    fi

    # Check if repository is archived
    local is_archived=$(check_archived "$pkg" "$ecosystem")
    if [[ "$is_archived" == "true" ]]; then
        status="archived"
        risk_factors+=("repository_archived")
        risk_level="critical"
        recommendations+=("Repository is archived - no future updates expected")
        recommendations+=("Immediately identify and migrate to alternative package")
    fi

    # Check last update date
    local last_update=$(get_last_update "$pkg" "$ecosystem")
    if [[ -n "$last_update" ]]; then
        local days_since=$(days_since_date "$last_update")

        if [[ $days_since -gt $ABANDONED_THRESHOLD_DAYS ]]; then
            status="abandoned"
            risk_factors+=("no_updates_${days_since}_days")
            risk_level="critical"
            recommendations+=("Package has not been updated in over 2 years")
            recommendations+=("Consider migrating to an actively maintained alternative")
        elif [[ $days_since -gt $STALE_THRESHOLD_DAYS ]]; then
            if [[ "$status" == "healthy" ]]; then
                status="stale"
            fi
            risk_factors+=("last_update_${days_since}_days_ago")
            if [[ "$risk_level" != "critical" ]]; then
                risk_level="high"
            fi
            recommendations+=("Package has not been updated in over 1 year")
            recommendations+=("Monitor for security issues and plan migration")
        elif [[ $days_since -gt $WARNING_THRESHOLD_DAYS ]]; then
            risk_factors+=("last_update_${days_since}_days_ago")
            if [[ "$risk_level" == "low" ]]; then
                risk_level="medium"
            fi
            recommendations+=("Consider monitoring package health more closely")
        fi
    else
        risk_factors+=("unknown_last_update")
        if [[ "$risk_level" == "low" ]]; then
            risk_level="medium"
        fi
    fi

    # Get maintenance metrics from OpenSSF Scorecard
    local metrics=$(get_maintenance_metrics "$pkg" "$ecosystem")
    local maintained_score=$(echo "$metrics" | jq -r '.maintained_score // null')

    if [[ "$maintained_score" != "null" && "$maintained_score" != "" ]]; then
        if [[ $(echo "$maintained_score < 3" | bc -l 2>/dev/null || echo "0") == "1" ]]; then
            risk_factors+=("low_maintained_score_${maintained_score}")
            if [[ "$risk_level" == "low" ]]; then
                risk_level="medium"
            fi
            recommendations+=("OpenSSF Scorecard indicates low maintenance activity")
        fi
    fi

    # Convert arrays to JSON (handle empty arrays)
    local risk_factors_json="[]"
    local recommendations_json="[]"
    if [[ ${#risk_factors[@]} -gt 0 ]]; then
        risk_factors_json=$(printf '%s\n' "${risk_factors[@]}" | jq -R . | jq -s '.')
    fi
    if [[ ${#recommendations[@]} -gt 0 ]]; then
        recommendations_json=$(printf '%s\n' "${recommendations[@]}" | jq -R . | jq -s '.')
    fi

    # Build final response
    echo "{
        \"package\": \"$pkg\",
        \"ecosystem\": \"$ecosystem\",
        \"status\": \"$status\",
        \"risk_level\": \"$risk_level\",
        \"risk_factors\": $risk_factors_json,
        \"recommendations\": $recommendations_json,
        \"metrics\": $metrics,
        \"last_update\": \"$last_update\"
    }" | jq '.'
}

# Batch check abandonment status for multiple packages
# Usage: check_abandonment_batch <packages_json>
# Input: [{"name": "lodash", "ecosystem": "npm"}, ...]
check_abandonment_batch() {
    local packages_json="$1"
    local results="[]"

    while IFS= read -r pkg; do
        local name=$(echo "$pkg" | jq -r '.name')
        local ecosystem=$(echo "$pkg" | jq -r '.ecosystem // "npm"')

        local status=$(check_abandonment_status "$name" "$ecosystem")
        results=$(echo "$results" | jq --argjson item "$status" '. + [$item]')
    done < <(echo "$packages_json" | jq -c '.[]')

    echo "$results"
}

# Generate abandonment report summary
# Usage: generate_abandonment_report <packages_json>
generate_abandonment_report() {
    local packages_json="$1"

    local results=$(check_abandonment_batch "$packages_json")

    local total=$(echo "$results" | jq 'length')
    local healthy=$(echo "$results" | jq '[.[] | select(.status == "healthy")] | length')
    local stale=$(echo "$results" | jq '[.[] | select(.status == "stale")] | length')
    local abandoned=$(echo "$results" | jq '[.[] | select(.status == "abandoned")] | length')
    local deprecated=$(echo "$results" | jq '[.[] | select(.status == "deprecated")] | length')
    local archived=$(echo "$results" | jq '[.[] | select(.status == "archived")] | length')

    local critical=$(echo "$results" | jq '[.[] | select(.risk_level == "critical")] | length')
    local high=$(echo "$results" | jq '[.[] | select(.risk_level == "high")] | length')
    local medium=$(echo "$results" | jq '[.[] | select(.risk_level == "medium")] | length')

    echo "{
        \"summary\": {
            \"total_packages\": $total,
            \"healthy\": $healthy,
            \"stale\": $stale,
            \"abandoned\": $abandoned,
            \"deprecated\": $deprecated,
            \"archived\": $archived,
            \"risk_breakdown\": {
                \"critical\": $critical,
                \"high\": $high,
                \"medium\": $medium,
                \"low\": $((total - critical - high - medium))
            }
        },
        \"packages\": $results
    }" | jq '.'
}

#############################################################################
# Export Functions
#############################################################################

export -f days_since_date
export -f check_deprecated
export -f get_deprecation_reason
export -f check_archived
export -f get_last_update
export -f get_maintenance_metrics
export -f check_abandonment_status
export -f check_abandonment_batch
export -f generate_abandonment_report
