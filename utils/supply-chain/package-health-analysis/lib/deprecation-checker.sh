#!/bin/bash
# Deprecation Checker
# Copyright (c) 2024 Gibson Powers Contributors
# SPDX-License-Identifier: GPL-3.0

set -euo pipefail

# Get script directory
LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Load deps.dev client (if not already loaded)
if ! command -v get_package_info &> /dev/null; then
    source "$LIB_DIR/deps-dev-client.sh"
fi

# Check if package is deprecated via deps.dev
# Usage: check_package_deprecation <system> <package>
check_package_deprecation() {
    local system=$1
    local package=$2

    local package_info=$(get_package_info "$system" "$package")

    if echo "$package_info" | jq -e '.error' > /dev/null 2>&1; then
        echo "{
            \"deprecated\": false,
            \"error\": \"could_not_fetch_data\",
            \"source\": \"deps.dev\"
        }"
        return 1
    fi

    local is_deprecated=$(echo "$package_info" | jq -r '.deprecated // false')
    local deprecation_message=$(echo "$package_info" | jq -r '.deprecationMessage // ""')

    if [ "$is_deprecated" = "true" ]; then
        # Try to extract deprecation date and alternatives from message
        local alternatives=$(extract_alternatives_from_message "$deprecation_message")

        jq -n \
            --argjson deprecated true \
            --arg message "$deprecation_message" \
            --argjson alternatives "$alternatives" \
            '{
                deprecated: $deprecated,
                deprecation_message: $message,
                alternative_packages: $alternatives,
                source: "deps.dev"
            }'
    else
        echo "{
            \"deprecated\": false,
            \"source\": \"deps.dev\"
        }"
    fi
}

# Extract alternative package names from deprecation message
# Usage: extract_alternatives_from_message <message>
extract_alternatives_from_message() {
    local message=$1

    # Common patterns in deprecation messages:
    # "use X instead"
    # "migrate to X"
    # "replaced by X"
    # "switch to X"

    local alternatives="[]"

    # Try to extract package names after common phrases
    if echo "$message" | grep -qi "use.*instead"; then
        local extracted=$(echo "$message" | sed -n 's/.*[Uu]se \([a-zA-Z0-9_-]*\).*/\1/p')
        if [ -n "$extracted" ]; then
            alternatives=$(echo "$alternatives" | jq --arg alt "$extracted" '. + [$alt]')
        fi
    fi

    if echo "$message" | grep -qi "migrate to"; then
        local extracted=$(echo "$message" | sed -n 's/.*[Mm]igrate to \([a-zA-Z0-9_-]*\).*/\1/p')
        if [ -n "$extracted" ]; then
            alternatives=$(echo "$alternatives" | jq --arg alt "$extracted" '. + [$alt]')
        fi
    fi

    if echo "$message" | grep -qi "replaced by"; then
        local extracted=$(echo "$message" | sed -n 's/.*[Rr]eplaced by \([a-zA-Z0-9_-]*\).*/\1/p')
        if [ -n "$extracted" ]; then
            alternatives=$(echo "$alternatives" | jq --arg alt "$extracted" '. + [$alt]')
        fi
    fi

    # Return unique alternatives
    echo "$alternatives" | jq 'unique'
}

# Find known deprecated packages for ecosystem
# Returns predefined list of commonly deprecated packages
# Usage: get_known_deprecated_packages <system>
get_known_deprecated_packages() {
    local system=$1

    case $system in
        npm)
            echo '[
                "request",
                "node-uuid",
                "coffee-script",
                "bower",
                "gulp-util",
                "natives",
                "nsp"
            ]'
            ;;
        pypi)
            echo '[
                "pycrypto",
                "python-memcached",
                "oslo.config"
            ]'
            ;;
        *)
            echo '[]'
            ;;
    esac
}

# Check if package is in known deprecated list
# Usage: is_known_deprecated <system> <package>
is_known_deprecated() {
    local system=$1
    local package=$2

    local known=$(get_known_deprecated_packages "$system")
    local found=$(echo "$known" | jq -r --arg pkg "$package" 'map(select(. == $pkg)) | length > 0')

    echo "$found"
}

# Get suggested alternatives for deprecated packages
# Usage: get_suggested_alternatives <system> <package>
get_suggested_alternatives() {
    local system=$1
    local package=$2

    # Known alternatives mapping
    case $system in
        npm)
            case $package in
                request)
                    echo '["axios", "node-fetch", "got"]'
                    ;;
                node-uuid)
                    echo '["uuid"]'
                    ;;
                coffee-script)
                    echo '["coffeescript"]'
                    ;;
                bower)
                    echo '["npm", "yarn"]'
                    ;;
                gulp-util)
                    echo '["gulp-plugin utilities"]'
                    ;;
                *)
                    echo '[]'
                    ;;
            esac
            ;;
        pypi)
            case $package in
                pycrypto)
                    echo '["pycryptodome", "cryptography"]'
                    ;;
                python-memcached)
                    echo '["pymemcache", "python3-memcached"]'
                    ;;
                *)
                    echo '[]'
                    ;;
            esac
            ;;
        *)
            echo '[]'
            ;;
    esac
}

# Comprehensive deprecation check
# Usage: comprehensive_deprecation_check <system> <package>
comprehensive_deprecation_check() {
    local system=$1
    local package=$2

    # Check via deps.dev API
    local api_result=$(check_package_deprecation "$system" "$package")
    local api_deprecated=$(echo "$api_result" | jq -r '.deprecated // false')

    # Check against known deprecated list
    local known_deprecated=$(is_known_deprecated "$system" "$package")

    # Determine final status
    local is_deprecated="false"
    local deprecation_message=""
    local alternatives="[]"
    local confidence="low"

    if [ "$api_deprecated" = "true" ]; then
        is_deprecated="true"
        deprecation_message=$(echo "$api_result" | jq -r '.deprecation_message // ""')
        alternatives=$(echo "$api_result" | jq -r '.alternative_packages // []')
        confidence="high"
    elif [ "$known_deprecated" = "true" ]; then
        is_deprecated="true"
        deprecation_message="Package is known to be deprecated"
        alternatives=$(get_suggested_alternatives "$system" "$package")
        confidence="medium"
    fi

    # If no alternatives found yet, try to suggest based on known mappings
    if [ "$(echo "$alternatives" | jq 'length')" -eq 0 ] && [ "$is_deprecated" = "true" ]; then
        alternatives=$(get_suggested_alternatives "$system" "$package")
    fi

    # Build result
    jq -n \
        --arg pkg "$package" \
        --arg sys "$system" \
        --argjson deprecated "$is_deprecated" \
        --arg message "$deprecation_message" \
        --argjson alternatives "$alternatives" \
        --arg confidence "$confidence" \
        '{
            package: $pkg,
            system: $sys,
            deprecated: $deprecated,
            deprecation_message: $message,
            alternative_packages: $alternatives,
            confidence: $confidence
        }'
}

# Check multiple packages for deprecation
# Usage: check_multiple_packages <packages_json>
# Input format: [{"system": "npm", "package": "request"}, ...]
check_multiple_packages() {
    local packages_json=$1

    local results="[]"

    while IFS= read -r item; do
        [ -z "$item" ] && continue

        local system=$(echo "$item" | jq -r '.system')
        local package=$(echo "$item" | jq -r '.package')

        local result=$(comprehensive_deprecation_check "$system" "$package")

        results=$(echo "$results" | jq --argjson item "$result" '. + [$item]')
    done < <(echo "$packages_json" | jq -c '.[]')

    echo "$results"
}

# Generate deprecation report
# Usage: generate_deprecation_report <deprecation_results>
generate_deprecation_report() {
    local results=$1

    local total=$(echo "$results" | jq 'length')
    local deprecated_count=$(echo "$results" | jq '[.[] | select(.deprecated == true)] | length')
    local with_alternatives=$(echo "$results" | jq '[.[] | select(.deprecated == true and (.alternative_packages | length > 0))] | length')

    # Group by confidence
    local high_confidence=$(echo "$results" | jq '[.[] | select(.deprecated == true and .confidence == "high")] | length')
    local medium_confidence=$(echo "$results" | jq '[.[] | select(.deprecated == true and .confidence == "medium")] | length')

    jq -n \
        --argjson total "$total" \
        --argjson deprecated "$deprecated_count" \
        --argjson with_alts "$with_alternatives" \
        --argjson high "$high_confidence" \
        --argjson medium "$medium_confidence" \
        --argjson details "$results" \
        '{
            summary: {
                total_packages_checked: $total,
                deprecated_packages: $deprecated,
                with_alternatives: $with_alts,
                high_confidence: $high,
                medium_confidence: $medium
            },
            deprecated_packages: [
                $details[] | select(.deprecated == true)
            ]
        }'
}

# Assess migration urgency
# Usage: assess_migration_urgency <package> <usage_count> <has_vulnerabilities>
assess_migration_urgency() {
    local package=$1
    local usage_count=$2
    local has_vulnerabilities=$3

    local urgency="low"
    local score=0

    # Factor 1: Usage count (higher usage = higher urgency)
    if [ "$usage_count" -ge 10 ]; then
        score=$((score + 3))
    elif [ "$usage_count" -ge 5 ]; then
        score=$((score + 2))
    elif [ "$usage_count" -ge 2 ]; then
        score=$((score + 1))
    fi

    # Factor 2: Vulnerabilities (critical factor)
    if [ "$has_vulnerabilities" = "true" ]; then
        score=$((score + 5))
    fi

    # Factor 3: Known critical packages (e.g., request)
    case $package in
        request|pycrypto)
            score=$((score + 2))
            ;;
    esac

    # Determine urgency
    if [ $score -ge 7 ]; then
        urgency="critical"
    elif [ $score -ge 5 ]; then
        urgency="high"
    elif [ $score -ge 3 ]; then
        urgency="medium"
    fi

    echo "$urgency"
}

# Export functions
export -f check_package_deprecation
export -f extract_alternatives_from_message
export -f get_known_deprecated_packages
export -f is_known_deprecated
export -f get_suggested_alternatives
export -f comprehensive_deprecation_check
export -f check_multiple_packages
export -f generate_deprecation_report
export -f assess_migration_urgency
