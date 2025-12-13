#!/usr/bin/env bash
# Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0
#
# Dockerfile Analyzer
# Parses Dockerfiles for best practices, base images, and multi-stage detection
#
# Usage:
#   source dockerfile-analyzer.sh
#   analyze_dockerfile "/path/to/Dockerfile"

set -euo pipefail

# Check if hadolint is available
has_hadolint() {
    command -v hadolint &>/dev/null
}

# Find all Dockerfiles in a directory
# Usage: find_dockerfiles "/path/to/project"
find_dockerfiles() {
    local target="$1"

    # Find Dockerfile, Dockerfile.*, *.dockerfile, and docker/Dockerfile patterns
    find "$target" \
        -type f \
        \( -name "Dockerfile" -o -name "Dockerfile.*" -o -name "*.dockerfile" \) \
        -not -path "*/node_modules/*" \
        -not -path "*/.git/*" \
        -not -path "*/vendor/*" \
        2>/dev/null | sort
}

# Extract FROM instructions from a Dockerfile
# Usage: extract_from_instructions "/path/to/Dockerfile"
# Returns JSON array of base images
extract_from_instructions() {
    local dockerfile="$1"
    local images=()
    local line_num=0

    while IFS= read -r line || [[ -n "$line" ]]; do
        line_num=$((line_num + 1))
        # Match FROM instruction (case insensitive)
        if [[ "$line" =~ ^[[:space:]]*[Ff][Rr][Oo][Mm][[:space:]]+([^[:space:]]+) ]]; then
            local image="${BASH_REMATCH[1]}"
            local alias=""

            # Check for AS alias
            if [[ "$line" =~ [Aa][Ss][[:space:]]+([^[:space:]]+) ]]; then
                alias="${BASH_REMATCH[1]}"
            fi

            images+=("{\"image\": \"$image\", \"line\": $line_num, \"alias\": \"$alias\"}")
        fi
    done < "$dockerfile"

    # Output as JSON array
    printf '%s\n' "${images[@]}" | jq -s '.'
}

# Check best practices without external tools (regex-based)
# Usage: check_best_practices_basic "/path/to/Dockerfile"
# Returns JSON object with best practice checks
check_best_practices_basic() {
    local dockerfile="$1"
    local content
    content=$(cat "$dockerfile")

    # Initialize checks
    local has_user="false"
    local has_healthcheck="false"
    local uses_add="false"
    local has_unpinned_packages="false"
    local runs_as_root="true"
    local has_workdir="false"
    local has_label="false"
    local copies_sensitive_files="false"

    # Check for USER instruction (non-root)
    if grep -qiE "^[[:space:]]*USER[[:space:]]+" "$dockerfile"; then
        has_user="true"
        # Check if it's not root
        if ! grep -qiE "^[[:space:]]*USER[[:space:]]+(root|0)" "$dockerfile"; then
            runs_as_root="false"
        fi
    fi

    # Check for HEALTHCHECK
    if grep -qiE "^[[:space:]]*HEALTHCHECK[[:space:]]+" "$dockerfile"; then
        has_healthcheck="true"
    fi

    # Check for ADD (should use COPY instead)
    if grep -qiE "^[[:space:]]*ADD[[:space:]]+" "$dockerfile"; then
        uses_add="true"
    fi

    # Check for unpinned package versions
    if grep -qE "(apt-get install|apk add|yum install|dnf install)" "$dockerfile"; then
        # Check if packages have versions pinned
        if grep -E "(apt-get install|apk add)" "$dockerfile" | grep -qvE "=[0-9]"; then
            has_unpinned_packages="true"
        fi
    fi

    # Check for WORKDIR
    if grep -qiE "^[[:space:]]*WORKDIR[[:space:]]+" "$dockerfile"; then
        has_workdir="true"
    fi

    # Check for LABEL (metadata)
    if grep -qiE "^[[:space:]]*LABEL[[:space:]]+" "$dockerfile"; then
        has_label="true"
    fi

    # Check for copying sensitive files
    if grep -qiE "COPY.*(\\.env|credentials|secrets|password|\.pem|\.key)" "$dockerfile"; then
        copies_sensitive_files="true"
    fi

    # Build JSON result
    jq -n \
        --arg has_user "$has_user" \
        --arg has_healthcheck "$has_healthcheck" \
        --arg uses_add "$uses_add" \
        --arg has_unpinned_packages "$has_unpinned_packages" \
        --arg runs_as_root "$runs_as_root" \
        --arg has_workdir "$has_workdir" \
        --arg has_label "$has_label" \
        --arg copies_sensitive_files "$copies_sensitive_files" \
        '{
            has_user: ($has_user == "true"),
            has_healthcheck: ($has_healthcheck == "true"),
            uses_add: ($uses_add == "true"),
            has_unpinned_packages: ($has_unpinned_packages == "true"),
            runs_as_root: ($runs_as_root == "true"),
            has_workdir: ($has_workdir == "true"),
            has_label: ($has_label == "true"),
            copies_sensitive_files: ($copies_sensitive_files == "true")
        }'
}

# Generate issues from basic checks
# Usage: generate_basic_issues <best_practices_json>
generate_basic_issues() {
    local practices="$1"
    local issues='[]'

    # Check each practice and generate issues
    if [[ $(echo "$practices" | jq -r '.runs_as_root') == "true" ]]; then
        issues=$(echo "$issues" | jq '. + [{
            "rule": "NO_USER",
            "severity": "warning",
            "line": 0,
            "message": "Container runs as root. Add USER instruction to run as non-root user",
            "category": "security"
        }]')
    fi

    if [[ $(echo "$practices" | jq -r '.has_healthcheck') == "false" ]]; then
        issues=$(echo "$issues" | jq '. + [{
            "rule": "NO_HEALTHCHECK",
            "severity": "info",
            "line": 0,
            "message": "No HEALTHCHECK instruction. Consider adding for container orchestration",
            "category": "best_practice"
        }]')
    fi

    if [[ $(echo "$practices" | jq -r '.uses_add') == "true" ]]; then
        issues=$(echo "$issues" | jq '. + [{
            "rule": "USE_COPY",
            "severity": "warning",
            "line": 0,
            "message": "Using ADD instruction. Prefer COPY unless you need ADD features (tar extraction, URLs)",
            "category": "best_practice"
        }]')
    fi

    if [[ $(echo "$practices" | jq -r '.has_unpinned_packages') == "true" ]]; then
        issues=$(echo "$issues" | jq '. + [{
            "rule": "PIN_VERSIONS",
            "severity": "warning",
            "line": 0,
            "message": "Package versions not pinned. Pin versions for reproducible builds",
            "category": "reproducibility"
        }]')
    fi

    if [[ $(echo "$practices" | jq -r '.has_workdir') == "false" ]]; then
        issues=$(echo "$issues" | jq '. + [{
            "rule": "NO_WORKDIR",
            "severity": "info",
            "line": 0,
            "message": "No WORKDIR instruction. Consider setting explicit working directory",
            "category": "best_practice"
        }]')
    fi

    if [[ $(echo "$practices" | jq -r '.copies_sensitive_files') == "true" ]]; then
        issues=$(echo "$issues" | jq '. + [{
            "rule": "SENSITIVE_FILES",
            "severity": "error",
            "line": 0,
            "message": "Potentially copying sensitive files (.env, credentials, keys). Use secrets management instead",
            "category": "security"
        }]')
    fi

    echo "$issues"
}

# Run hadolint if available
# Usage: run_hadolint "/path/to/Dockerfile"
# Returns JSON array of issues
run_hadolint() {
    local dockerfile="$1"

    if ! has_hadolint; then
        echo '[]'
        return
    fi

    # Run hadolint with JSON output
    local output
    output=$(hadolint -f json "$dockerfile" 2>/dev/null || echo '[]')

    # Transform hadolint output to our format
    echo "$output" | jq '[.[] | {
        rule: .code,
        severity: .level,
        line: .line,
        message: .message,
        category: "hadolint"
    }]' 2>/dev/null || echo '[]'
}

# Analyze a single Dockerfile
# Usage: analyze_dockerfile "/path/to/Dockerfile"
# Returns complete JSON analysis
analyze_dockerfile() {
    local dockerfile="$1"
    local relative_path="${2:-$dockerfile}"

    if [[ ! -f "$dockerfile" ]]; then
        jq -n --arg path "$relative_path" '{
            path: $path,
            error: "File not found",
            analyzed: false
        }'
        return 1
    fi

    # Extract FROM instructions
    local from_instructions
    from_instructions=$(extract_from_instructions "$dockerfile")

    # Get base images array
    local base_images
    base_images=$(echo "$from_instructions" | jq '[.[].image]')

    # Get final base (last FROM)
    local final_base
    final_base=$(echo "$from_instructions" | jq -r '.[-1].image // "unknown"')

    # Count stages
    local stage_count
    stage_count=$(echo "$from_instructions" | jq 'length')

    # Check if multi-stage
    local is_multistage="false"
    if [[ "$stage_count" -gt 1 ]]; then
        is_multistage="true"
    fi

    # Get aliases
    local aliases
    aliases=$(echo "$from_instructions" | jq '[.[] | select(.alias != "") | .alias]')

    # Run best practices checks
    local best_practices
    best_practices=$(check_best_practices_basic "$dockerfile")

    # Generate basic issues
    local basic_issues
    basic_issues=$(generate_basic_issues "$best_practices")

    # Run hadolint if available
    local hadolint_issues
    hadolint_issues=$(run_hadolint "$dockerfile")

    # Merge issues
    local all_issues
    all_issues=$(echo "$basic_issues" "$hadolint_issues" | jq -s 'add')

    # Count issues by severity
    local error_count warning_count info_count
    error_count=$(echo "$all_issues" | jq '[.[] | select(.severity == "error")] | length')
    warning_count=$(echo "$all_issues" | jq '[.[] | select(.severity == "warning")] | length')
    info_count=$(echo "$all_issues" | jq '[.[] | select(.severity == "info")] | length')

    # Build result
    jq -n \
        --arg path "$relative_path" \
        --argjson base_images "$base_images" \
        --arg final_base "$final_base" \
        --arg is_multistage "$is_multistage" \
        --argjson stages "$stage_count" \
        --argjson aliases "$aliases" \
        --argjson from_instructions "$from_instructions" \
        --argjson issues "$all_issues" \
        --argjson best_practices "$best_practices" \
        --argjson error_count "$error_count" \
        --argjson warning_count "$warning_count" \
        --argjson info_count "$info_count" \
        --arg has_hadolint "$(has_hadolint && echo true || echo false)" \
        '{
            path: $path,
            base_images: $base_images,
            final_base: $final_base,
            is_multistage: ($is_multistage == "true"),
            stages: $stages,
            stage_aliases: $aliases,
            from_instructions: $from_instructions,
            issues: $issues,
            issue_counts: {
                error: $error_count,
                warning: $warning_count,
                info: $info_count,
                total: ($error_count + $warning_count + $info_count)
            },
            best_practices: $best_practices,
            analyzed_with_hadolint: ($has_hadolint == "true"),
            analyzed: true
        }'
}

# Analyze all Dockerfiles in a project
# Usage: analyze_all_dockerfiles "/path/to/project"
analyze_all_dockerfiles() {
    local target="$1"
    local results='[]'

    while IFS= read -r dockerfile; do
        [[ -z "$dockerfile" ]] && continue

        # Get relative path
        local rel_path="${dockerfile#$target/}"

        local analysis
        analysis=$(analyze_dockerfile "$dockerfile" "$rel_path")
        results=$(echo "$results" | jq --argjson a "$analysis" '. + [$a]')
    done < <(find_dockerfiles "$target")

    echo "$results"
}

# Export functions
export -f has_hadolint
export -f find_dockerfiles
export -f extract_from_instructions
export -f check_best_practices_basic
export -f generate_basic_issues
export -f run_hadolint
export -f analyze_dockerfile
export -f analyze_all_dockerfiles
