#!/bin/bash
# Typosquatting Detector
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0
#
# Detects potentially malicious packages with names similar to popular packages.
# Part of the Security & Risk Management module.

set -eo pipefail

# Get script directory for loading shared libraries
LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCANNER_DIR="$(dirname "$LIB_DIR")"
SCANNERS_ROOT="$(dirname "$SCANNER_DIR")"

# Load popular packages library (from package-sbom)
if [[ -f "$SCANNERS_ROOT/package-sbom/lib/popular-packages.sh" ]]; then
    source "$SCANNERS_ROOT/package-sbom/lib/popular-packages.sh"
fi

# Load deps.dev client from shared libs (if not already loaded)
if ! command -v deps_dev_get_package_info &> /dev/null; then
    source "$UTILS_ROOT/lib/deps-dev-client.sh"
fi

#############################################################################
# Configuration
#############################################################################

# Levenshtein distance thresholds
TYPOSQUAT_THRESHOLD=${TYPOSQUAT_THRESHOLD:-2}
TYPOSQUAT_HIGH_RISK_THRESHOLD=${TYPOSQUAT_HIGH_RISK_THRESHOLD:-1}

# Age threshold for suspicious new packages (in days)
NEW_PACKAGE_THRESHOLD_DAYS=${NEW_PACKAGE_THRESHOLD_DAYS:-30}

# Minimum download count for established packages
MIN_DOWNLOAD_THRESHOLD=${MIN_DOWNLOAD_THRESHOLD:-1000}

#############################################################################
# Enhanced Typosquatting Detection Functions
#############################################################################

# Check for scope confusion attack vectors
# Usage: check_scope_confusion <package> <ecosystem>
check_scope_confusion() {
    local pkg="$1"
    local ecosystem="${2:-npm}"
    local confusions=()

    # Only relevant for npm ecosystem
    if [[ "$ecosystem" != "npm" && "$ecosystem" != "node" ]]; then
        echo "[]"
        return
    fi

    # Check if unscoped package might be confused with scoped one
    if [[ ! "$pkg" =~ ^@ ]]; then
        # Common scope prefixes to check
        local scopes=("@aws-sdk" "@google-cloud" "@azure" "@anthropic-ai" "@openai" "@types" "@babel" "@emotion" "@mui" "@testing-library")

        for scope in "${scopes[@]}"; do
            local scoped="${scope}/${pkg}"
            if [[ $(is_popular_package "$scoped" "$ecosystem" 2>/dev/null) == "true" ]]; then
                confusions+=("{\"type\": \"missing_scope\", \"popular_package\": \"$scoped\"}")
            fi
        done
    fi

    # Check if scoped package might be impersonating different scope
    if [[ "$pkg" =~ ^@([^/]+)/(.+)$ ]]; then
        local current_scope="${BASH_REMATCH[1]}"
        local package_name="${BASH_REMATCH[2]}"

        # Check if the unscoped version is popular
        if [[ $(is_popular_package "$package_name" "$ecosystem" 2>/dev/null) == "true" ]]; then
            confusions+=("{\"type\": \"suspicious_scope\", \"popular_package\": \"$package_name\", \"suspicious_scope\": \"$current_scope\"}")
        fi
    fi

    if [[ ${#confusions[@]} -gt 0 ]]; then
        printf '%s\n' "${confusions[@]}" | jq -s '.'
    else
        echo "[]"
    fi
}

# Check for common typosquatting patterns
# Usage: check_common_patterns <package> <ecosystem>
check_common_patterns() {
    local pkg="$1"
    local ecosystem="${2:-npm}"
    local patterns=()

    # Get popular packages list
    local popular_list=$(get_popular_packages "$ecosystem" 2>/dev/null)

    # Pattern 1: Missing hyphen (date-fns -> datefns)
    local no_hyphen="${pkg//-/}"
    if [[ "$no_hyphen" != "$pkg" ]]; then
        while IFS= read -r popular; do
            local popular_no_hyphen="${popular//-/}"
            if [[ "$no_hyphen" == "$popular_no_hyphen" && "$pkg" != "$popular" ]]; then
                patterns+=("{\"pattern\": \"hyphen_removal\", \"similar_to\": \"$popular\"}")
            fi
        done <<< "$popular_list"
    fi

    # Pattern 2: Added hyphen (lodash -> lo-dash)
    while IFS= read -r popular; do
        local popular_no_hyphen="${popular//-/}"
        local pkg_no_hyphen="${pkg//-/}"
        if [[ "$pkg_no_hyphen" == "$popular_no_hyphen" && "$pkg" != "$popular" ]]; then
            patterns+=("{\"pattern\": \"hyphen_addition\", \"similar_to\": \"$popular\"}")
        fi
    done <<< "$popular_list"

    # Pattern 3: Plural/singular confusion (request -> requests, colors -> color)
    while IFS= read -r popular; do
        if [[ "${pkg}s" == "$popular" || "$pkg" == "${popular}s" ]]; then
            patterns+=("{\"pattern\": \"plural_confusion\", \"similar_to\": \"$popular\"}")
        fi
    done <<< "$popular_list"

    # Pattern 4: Common character substitutions
    # Check for: rn->m, l->1, o->0, i->1, etc.
    local pkg_normalized=$(echo "$pkg" | sed 's/rn/m/g; s/1/l/g; s/0/o/g')
    while IFS= read -r popular; do
        local popular_normalized=$(echo "$popular" | sed 's/rn/m/g; s/1/l/g; s/0/o/g')
        if [[ "$pkg_normalized" == "$popular_normalized" && "$pkg" != "$popular" ]]; then
            patterns+=("{\"pattern\": \"character_substitution\", \"similar_to\": \"$popular\"}")
        fi
    done <<< "$popular_list"

    # Pattern 5: JS/TS suffix confusion
    while IFS= read -r popular; do
        if [[ "$pkg" == "${popular}js" || "$pkg" == "${popular}-js" || "$pkg" == "${popular}.js" ]]; then
            patterns+=("{\"pattern\": \"js_suffix\", \"similar_to\": \"$popular\"}")
        fi
        if [[ "$pkg" == "${popular}ts" || "$pkg" == "${popular}-ts" || "$pkg" == "${popular}.ts" ]]; then
            patterns+=("{\"pattern\": \"ts_suffix\", \"similar_to\": \"$popular\"}")
        fi
    done <<< "$popular_list"

    if [[ ${#patterns[@]} -gt 0 ]]; then
        printf '%s\n' "${patterns[@]}" | jq -s '.'
    else
        echo "[]"
    fi
}

# Get package publish date
# Usage: get_package_age <package> <ecosystem>
get_package_age() {
    local pkg="$1"
    local ecosystem="$2"

    # Get package info from deps.dev
    local pkg_info=$(get_package_info "$ecosystem" "$pkg" 2>/dev/null)

    if [[ -z "$pkg_info" || "$pkg_info" == *"error"* ]]; then
        echo "unknown"
        return
    fi

    # Get first version's publish date
    local first_version=$(echo "$pkg_info" | jq -r '.versions[0].versionKey.version // ""')

    if [[ -n "$first_version" ]]; then
        local version_info=$(get_package_version "$ecosystem" "$pkg" "$first_version" 2>/dev/null)
        if [[ -n "$version_info" && "$version_info" != *"error"* ]]; then
            local publish_date=$(echo "$version_info" | jq -r '.publishedAt // ""')
            if [[ -n "$publish_date" ]]; then
                # Calculate days since publish
                python3 -c "
from datetime import datetime
try:
    d = datetime.fromisoformat('$publish_date'.replace('Z', '+00:00'))
    days = (datetime.now(d.tzinfo) - d).days if d.tzinfo else (datetime.now() - d).days
    print(max(0, days))
except:
    print('unknown')
" 2>/dev/null
                return
            fi
        fi
    fi

    echo "unknown"
}

# Get package download/dependent count as popularity metric
# Usage: get_package_popularity <package> <ecosystem>
get_package_popularity() {
    local pkg="$1"
    local ecosystem="$2"

    local pkg_info=$(get_package_info "$ecosystem" "$pkg" 2>/dev/null)

    if [[ -z "$pkg_info" || "$pkg_info" == *"error"* ]]; then
        echo "0"
        return
    fi

    echo "$pkg_info" | jq -r '.dependentCount // 0'
}

# Comprehensive typosquatting analysis
# Usage: analyze_typosquat_risk <package> <ecosystem>
analyze_typosquat_risk() {
    local pkg="$1"
    local ecosystem="${2:-npm}"

    # Check if package is itself popular (safe)
    if [[ $(is_popular_package "$pkg" "$ecosystem" 2>/dev/null) == "true" ]]; then
        echo '{
            "package": "'"$pkg"'",
            "ecosystem": "'"$ecosystem"'",
            "suspicious": false,
            "reason": "is_popular_package",
            "risk_level": "none"
        }' | jq '.'
        return
    fi

    # Check if known malicious
    if [[ $(is_known_malicious "$pkg" "$ecosystem" 2>/dev/null) == "true" ]]; then
        echo '{
            "package": "'"$pkg"'",
            "ecosystem": "'"$ecosystem"'",
            "suspicious": true,
            "reason": "known_malicious_package",
            "risk_level": "critical",
            "recommendations": ["Immediately remove this package", "Audit your codebase for malicious activity", "Check for compromised credentials or data exfiltration"]
        }' | jq '.'
        return
    fi

    local risk_factors=()
    local risk_level="none"
    local recommendations=()

    # Check Levenshtein distance similarity
    local similar=$(find_similar_packages "$pkg" "$ecosystem" "$TYPOSQUAT_THRESHOLD" 2>/dev/null)
    local similar_count=$(echo "$similar" | jq 'length' 2>/dev/null || echo "0")

    if [[ $similar_count -gt 0 ]]; then
        local min_distance=$(echo "$similar" | jq '[.[].distance] | min' 2>/dev/null || echo "999")

        if [[ $min_distance -le $TYPOSQUAT_HIGH_RISK_THRESHOLD ]]; then
            risk_level="high"
            risk_factors+=("very_similar_to_popular_package")
            recommendations+=("Package name is very similar to a popular package - verify source before use")
        else
            risk_level="medium"
            risk_factors+=("similar_to_popular_package")
            recommendations+=("Package name is similar to a popular package - verify this is the intended package")
        fi
    fi

    # Check scope confusion
    local scope_issues=$(check_scope_confusion "$pkg" "$ecosystem")
    local scope_count=$(echo "$scope_issues" | jq 'length' 2>/dev/null || echo "0")

    if [[ $scope_count -gt 0 ]]; then
        if [[ "$risk_level" == "none" ]]; then
            risk_level="medium"
        elif [[ "$risk_level" == "medium" ]]; then
            risk_level="high"
        fi
        risk_factors+=("scope_confusion_risk")
        recommendations+=("Check if this is the correct package - similar scoped packages exist")
    fi

    # Check common patterns
    local patterns=$(check_common_patterns "$pkg" "$ecosystem")
    local pattern_count=$(echo "$patterns" | jq 'length' 2>/dev/null || echo "0")

    if [[ $pattern_count -gt 0 ]]; then
        if [[ "$risk_level" == "none" ]]; then
            risk_level="medium"
        elif [[ "$risk_level" == "medium" ]]; then
            risk_level="high"
        fi
        risk_factors+=("matches_typosquat_pattern")
        recommendations+=("Package name matches common typosquatting patterns")
    fi

    # Check package age (new + similar = high risk)
    local age=$(get_package_age "$pkg" "$ecosystem")
    if [[ "$age" != "unknown" && $age -lt $NEW_PACKAGE_THRESHOLD_DAYS ]]; then
        if [[ ${#risk_factors[@]} -gt 0 ]]; then
            risk_level="critical"
            risk_factors+=("new_package_${age}_days_old")
            recommendations+=("CRITICAL: New package with suspicious name - high likelihood of typosquatting attack")
        fi
    fi

    # Check popularity (low popularity + similar to popular = suspicious)
    local popularity=$(get_package_popularity "$pkg" "$ecosystem")
    if [[ $popularity -lt $MIN_DOWNLOAD_THRESHOLD && ${#risk_factors[@]} -gt 0 ]]; then
        if [[ "$risk_level" == "high" ]]; then
            risk_level="critical"
        fi
        risk_factors+=("low_popularity_${popularity}_dependents")
        recommendations+=("Low adoption combined with suspicious name increases risk")
    fi

    # Determine if suspicious
    local suspicious="false"
    if [[ ${#risk_factors[@]} -gt 0 ]]; then
        suspicious="true"
    fi

    # Convert arrays to JSON
    local risk_factors_json=$(printf '%s\n' "${risk_factors[@]}" 2>/dev/null | jq -R . | jq -s '.' || echo "[]")
    local recommendations_json=$(printf '%s\n' "${recommendations[@]}" 2>/dev/null | jq -R . | jq -s '.' || echo "[]")

    # Build response
    echo "{
        \"package\": \"$pkg\",
        \"ecosystem\": \"$ecosystem\",
        \"suspicious\": $suspicious,
        \"risk_level\": \"$risk_level\",
        \"risk_factors\": $risk_factors_json,
        \"similar_packages\": $similar,
        \"scope_issues\": $scope_issues,
        \"pattern_matches\": $patterns,
        \"package_age_days\": \"$age\",
        \"dependent_count\": $popularity,
        \"recommendations\": $recommendations_json
    }" | jq '.'
}

# Batch analyze typosquat risk
# Usage: analyze_typosquat_batch <packages_json>
# Input: [{"name": "lodash", "ecosystem": "npm"}, ...]
analyze_typosquat_batch() {
    local packages_json="$1"
    local results="[]"

    while IFS= read -r pkg; do
        local name=$(echo "$pkg" | jq -r '.name')
        local ecosystem=$(echo "$pkg" | jq -r '.ecosystem // "npm"')

        local analysis=$(analyze_typosquat_risk "$name" "$ecosystem")
        results=$(echo "$results" | jq --argjson item "$analysis" '. + [$item]')
    done < <(echo "$packages_json" | jq -c '.[]')

    echo "$results"
}

# Generate typosquat risk report
# Usage: generate_typosquat_report <packages_json>
generate_typosquat_report() {
    local packages_json="$1"

    local results=$(analyze_typosquat_batch "$packages_json")

    local total=$(echo "$results" | jq 'length')
    local suspicious=$(echo "$results" | jq '[.[] | select(.suspicious == true)] | length')
    local safe=$(echo "$results" | jq '[.[] | select(.suspicious == false)] | length')

    local critical=$(echo "$results" | jq '[.[] | select(.risk_level == "critical")] | length')
    local high=$(echo "$results" | jq '[.[] | select(.risk_level == "high")] | length')
    local medium=$(echo "$results" | jq '[.[] | select(.risk_level == "medium")] | length')

    # Get high-risk packages
    local high_risk_packages=$(echo "$results" | jq '[.[] | select(.risk_level == "critical" or .risk_level == "high")]')

    echo "{
        \"summary\": {
            \"total_packages\": $total,
            \"suspicious_packages\": $suspicious,
            \"safe_packages\": $safe,
            \"risk_breakdown\": {
                \"critical\": $critical,
                \"high\": $high,
                \"medium\": $medium,
                \"none\": $((total - critical - high - medium))
            }
        },
        \"high_risk_packages\": $high_risk_packages,
        \"all_packages\": $results
    }" | jq '.'
}

#############################################################################
# Export Functions
#############################################################################

export -f check_scope_confusion
export -f check_common_patterns
export -f get_package_age
export -f get_package_popularity
export -f analyze_typosquat_risk
export -f analyze_typosquat_batch
export -f generate_typosquat_report
