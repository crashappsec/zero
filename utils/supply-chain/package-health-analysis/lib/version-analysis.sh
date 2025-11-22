#!/bin/bash
# Version Analysis Module
# Copyright (c) 2024 Gibson Powers Contributors
# SPDX-License-Identifier: GPL-3.0

set -euo pipefail

# Get script directory
LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
UTILS_ROOT="$(cd "$LIB_DIR/../.." && pwd)"

# Load configuration
if [ -f "$UTILS_ROOT/lib/config-loader.sh" ]; then
    source "$UTILS_ROOT/lib/config-loader.sh"
    CONFIG=$(load_config "package-health-analysis")
else
    CONFIG="{}"
fi

MIN_REPOS=$(echo "$CONFIG" | jq -r '.package_health.analysis.minimum_repos_for_version_inconsistency // 2')

# Parse semantic version
# Usage: parse_semver <version>
parse_semver() {
    local version=$1

    # Remove leading 'v' if present
    version=${version#v}

    # Extract major, minor, patch
    local major=$(echo "$version" | cut -d. -f1 | sed 's/[^0-9]//g')
    local minor=$(echo "$version" | cut -d. -f2 | sed 's/[^0-9]//g')
    local patch=$(echo "$version" | cut -d. -f3 | sed 's/[^0-9]//g')

    # Default to 0 if empty
    major=${major:-0}
    minor=${minor:-0}
    patch=${patch:-0}

    echo "$major.$minor.$patch"
}

# Compare semantic versions
# Returns: -1 (v1 < v2), 0 (v1 == v2), 1 (v1 > v2)
# Usage: compare_versions <version1> <version2>
compare_versions() {
    local v1=$1
    local v2=$2

    # Parse versions
    local v1_parsed=$(parse_semver "$v1")
    local v2_parsed=$(parse_semver "$v2")

    local v1_major=$(echo "$v1_parsed" | cut -d. -f1)
    local v1_minor=$(echo "$v1_parsed" | cut -d. -f2)
    local v1_patch=$(echo "$v1_parsed" | cut -d. -f3)

    local v2_major=$(echo "$v2_parsed" | cut -d. -f1)
    local v2_minor=$(echo "$v2_parsed" | cut -d. -f2)
    local v2_patch=$(echo "$v2_parsed" | cut -d. -f3)

    # Compare major
    if [ "$v1_major" -lt "$v2_major" ]; then
        echo "-1"
        return
    elif [ "$v1_major" -gt "$v2_major" ]; then
        echo "1"
        return
    fi

    # Compare minor
    if [ "$v1_minor" -lt "$v2_minor" ]; then
        echo "-1"
        return
    elif [ "$v1_minor" -gt "$v2_minor" ]; then
        echo "1"
        return
    fi

    # Compare patch
    if [ "$v1_patch" -lt "$v2_patch" ]; then
        echo "-1"
        return
    elif [ "$v1_patch" -gt "$v2_patch" ]; then
        echo "1"
        return
    fi

    echo "0"
}

# Find the most common version
# Usage: find_most_common_version <versions_json>
find_most_common_version() {
    local versions_json=$1

    echo "$versions_json" | jq -r '
        group_by(.) |
        map({version: .[0], count: length}) |
        sort_by(.count) |
        reverse |
        .[0].version
    '
}

# Find the latest version
# Usage: find_latest_version <versions_array_json>
find_latest_version() {
    local versions_json=$1

    local latest=""
    while IFS= read -r version; do
        if [ -z "$latest" ]; then
            latest="$version"
        else
            local comparison=$(compare_versions "$version" "$latest")
            if [ "$comparison" = "1" ]; then
                latest="$version"
            fi
        fi
    done < <(echo "$versions_json" | jq -r '.[]')

    echo "$latest"
}

# Analyze version inconsistencies
# Usage: analyze_version_inconsistencies <package_usage_json>
# Input format: [{"repo": "repo1", "version": "1.2.3"}, ...]
analyze_version_inconsistencies() {
    local package_usage=$1

    # Count unique versions
    local version_count=$(echo "$package_usage" | jq -r '[.[].version] | unique | length')

    if [ "$version_count" -eq 1 ]; then
        # All repos use same version
        echo "{
            \"has_inconsistency\": false,
            \"unique_versions\": 1
        }"
        return
    fi

    # Get repo count
    local repo_count=$(echo "$package_usage" | jq 'length')

    if [ "$repo_count" -lt "$MIN_REPOS" ]; then
        # Not enough repos to call it an inconsistency
        echo "{
            \"has_inconsistency\": false,
            \"unique_versions\": $version_count,
            \"note\": \"insufficient_repos\"
        }"
        return
    fi

    # Group repos by version
    local version_groups=$(echo "$package_usage" | jq -r '
        group_by(.version) |
        map({
            version: .[0].version,
            repos: [.[].repo],
            count: length
        })
    ')

    # Find recommended version (most common)
    local most_common=$(echo "$version_groups" | jq -r 'sort_by(.count) | reverse | .[0].version')

    # Find latest version
    local all_versions=$(echo "$package_usage" | jq -r '[.[].version]')
    local latest_version=$(find_latest_version "$all_versions")

    # Check if latest is being used
    local latest_in_use=$(echo "$package_usage" | jq -r --arg latest "$latest_version" '
        map(select(.version == $latest)) | length > 0
    ')

    # Identify outliers (versions used by only 1 repo)
    local outliers=$(echo "$version_groups" | jq -r '
        map(select(.count == 1)) |
        [.[].version]
    ')

    # Calculate severity
    local severity="low"
    local outlier_count=$(echo "$outliers" | jq 'length')

    if [ "$outlier_count" -gt 0 ]; then
        # Check if outliers are major versions behind
        local has_major_outlier=false
        while IFS= read -r outlier; do
            [ -z "$outlier" ] && continue
            local comparison=$(compare_versions "$outlier" "$latest_version")
            local outlier_major=$(parse_semver "$outlier" | cut -d. -f1)
            local latest_major=$(parse_semver "$latest_version" | cut -d. -f1)

            if [ "$outlier_major" -lt "$latest_major" ]; then
                has_major_outlier=true
                break
            fi
        done < <(echo "$outliers" | jq -r '.[]')

        if [ "$has_major_outlier" = true ]; then
            severity="high"
        else
            severity="medium"
        fi
    fi

    # Build result
    jq -n \
        --argjson groups "$version_groups" \
        --arg recommended "$most_common" \
        --arg latest "$latest_version" \
        --argjson latest_in_use "$latest_in_use" \
        --argjson outliers "$outliers" \
        --arg severity "$severity" \
        --argjson unique "$version_count" \
        --argjson total "$repo_count" \
        '{
            has_inconsistency: true,
            unique_versions: $unique,
            total_repos: $total,
            version_distribution: $groups,
            recommended_version: $recommended,
            latest_version: $latest,
            latest_in_use: $latest_in_use,
            outlier_versions: $outliers,
            severity: $severity
        }'
}

# Calculate migration complexity
# Usage: calculate_migration_complexity <from_version> <to_version>
calculate_migration_complexity() {
    local from_version=$1
    local to_version=$2

    local from_parsed=$(parse_semver "$from_version")
    local to_parsed=$(parse_semver "$to_version")

    local from_major=$(echo "$from_parsed" | cut -d. -f1)
    local from_minor=$(echo "$from_parsed" | cut -d. -f2)
    local to_major=$(echo "$to_parsed" | cut -d. -f1)
    local to_minor=$(echo "$to_parsed" | cut -d. -f2)

    local complexity="simple"
    local breaking_change=false

    # Major version change
    if [ "$from_major" != "$to_major" ]; then
        complexity="complex"
        breaking_change=true
    # Multiple minor versions
    elif [ $((to_minor - from_minor)) -gt 3 ]; then
        complexity="moderate"
    # Single minor version or patch
    elif [ "$from_minor" != "$to_minor" ]; then
        complexity="simple"
    else
        complexity="trivial"
    fi

    echo "{
        \"complexity\": \"$complexity\",
        \"breaking_change\": $breaking_change,
        \"from_version\": \"$from_version\",
        \"to_version\": \"$to_version\"
    }"
}

# Generate standardization recommendations
# Usage: generate_standardization_recommendations <inconsistency_analysis>
generate_standardization_recommendations() {
    local analysis=$1

    local has_inconsistency=$(echo "$analysis" | jq -r '.has_inconsistency')

    if [ "$has_inconsistency" = "false" ]; then
        echo "{
            \"needs_standardization\": false,
            \"message\": \"All repositories use the same version\"
        }"
        return
    fi

    local recommended=$(echo "$analysis" | jq -r '.recommended_version')
    local latest=$(echo "$analysis" | jq -r '.latest_version')
    local severity=$(echo "$analysis" | jq -r '.severity')

    # Should we recommend latest or most common?
    local target_version="$recommended"
    local rationale="Most commonly used version across repositories"

    # If latest is significantly newer and secure, recommend it instead
    local comparison=$(compare_versions "$recommended" "$latest")
    if [ "$comparison" = "-1" ]; then
        local rec_major=$(parse_semver "$recommended" | cut -d. -f1)
        local latest_major=$(parse_semver "$latest" | cut -d. -f1)

        if [ "$rec_major" = "$latest_major" ]; then
            # Same major version, recommend latest
            target_version="$latest"
            rationale="Latest version in the same major series"
        fi
    fi

    # Generate migration tasks
    local version_groups=$(echo "$analysis" | jq -r '.version_distribution')
    local migration_tasks="[]"

    migration_tasks=$(echo "$version_groups" | jq -r --arg target "$target_version" '
        map(select(.version != $target) | {
            from_version: .version,
            to_version: $target,
            affected_repos: .repos,
            repo_count: .count
        })
    ')

    # Add complexity to each task
    local tasks_with_complexity="[]"
    while IFS= read -r task; do
        [ -z "$task" ] && continue

        local from=$(echo "$task" | jq -r '.from_version')
        local to=$(echo "$task" | jq -r '.to_version')
        local complexity=$(calculate_migration_complexity "$from" "$to")

        local enhanced_task=$(echo "$task" | jq --argjson comp "$complexity" '. + {complexity: $comp}')
        tasks_with_complexity=$(echo "$tasks_with_complexity" | jq --argjson task "$enhanced_task" '. + [$task]')
    done < <(echo "$migration_tasks" | jq -c '.[]')

    # Calculate total effort estimate (in hours)
    local total_repos=$(echo "$tasks_with_complexity" | jq '[.[].repo_count] | add // 0')
    local has_complex=$(echo "$tasks_with_complexity" | jq '[.[].complexity.breaking_change] | any')

    local effort_per_repo=2  # hours
    if [ "$has_complex" = "true" ]; then
        effort_per_repo=8  # More effort for breaking changes
    fi

    local total_effort=$((total_repos * effort_per_repo))

    jq -n \
        --arg target "$target_version" \
        --arg rationale "$rationale" \
        --argjson tasks "$tasks_with_complexity" \
        --argjson effort "$total_effort" \
        --arg severity "$severity" \
        '{
            needs_standardization: true,
            target_version: $target,
            rationale: $rationale,
            severity: $severity,
            migration_tasks: $tasks,
            estimated_effort_hours: $effort,
            recommendations: [
                "Review breaking changes in target version",
                "Test migration in development environment first",
                "Update version in stages, starting with non-critical repos",
                "Monitor for issues after each migration"
            ]
        }'
}

# Analyze all packages for version inconsistencies
# Usage: analyze_all_versions <packages_json>
# Input: {"package_name": [{"repo": "repo1", "version": "1.0.0"}, ...], ...}
analyze_all_versions() {
    local packages_json=$1

    local results="[]"

    # Process each package
    while IFS= read -r package_name; do
        [ -z "$package_name" ] && continue

        local usage=$(echo "$packages_json" | jq -r --arg pkg "$package_name" '.[$pkg]')
        local analysis=$(analyze_version_inconsistencies "$usage")

        local has_inconsistency=$(echo "$analysis" | jq -r '.has_inconsistency')

        if [ "$has_inconsistency" = "true" ]; then
            local recommendations=$(generate_standardization_recommendations "$analysis")

            local result=$(jq -n \
                --arg pkg "$package_name" \
                --argjson analysis "$analysis" \
                --argjson recommendations "$recommendations" \
                '{
                    package: $pkg,
                    analysis: $analysis,
                    recommendations: $recommendations
                }')

            results=$(echo "$results" | jq --argjson item "$result" '. + [$item]')
        fi
    done < <(echo "$packages_json" | jq -r 'keys[]')

    echo "$results"
}

# Export functions
export -f parse_semver
export -f compare_versions
export -f find_most_common_version
export -f find_latest_version
export -f analyze_version_inconsistencies
export -f calculate_migration_complexity
export -f generate_standardization_recommendations
export -f analyze_all_versions
