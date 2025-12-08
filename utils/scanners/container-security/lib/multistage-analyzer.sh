#!/usr/bin/env bash
# Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0
#
# Multi-stage Build Analyzer
# Analyzes Dockerfiles for multi-stage build patterns and optimization opportunities
#
# Usage:
#   source multistage-analyzer.sh
#   analyze_multistage "/path/to/Dockerfile"

set -euo pipefail

# Parse stages from a Dockerfile
# Usage: parse_stages "/path/to/Dockerfile"
# Returns JSON array of stages
parse_stages() {
    local dockerfile="$1"
    local stages='[]'
    local current_stage=""
    local current_alias=""
    local current_image=""
    local stage_index=0
    local line_num=0
    local stage_start=0
    local instructions='[]'

    while IFS= read -r line || [[ -n "$line" ]]; do
        line_num=$((line_num + 1))

        # Match FROM instruction
        if [[ "$line" =~ ^[[:space:]]*[Ff][Rr][Oo][Mm][[:space:]]+([^[:space:]]+) ]]; then
            # Save previous stage if exists
            if [[ -n "$current_image" ]]; then
                stages=$(echo "$stages" | jq \
                    --arg image "$current_image" \
                    --arg alias "$current_alias" \
                    --argjson index "$stage_index" \
                    --argjson start "$stage_start" \
                    --argjson end "$((line_num - 1))" \
                    --argjson instructions "$instructions" \
                    '. + [{
                        index: $index,
                        image: $image,
                        alias: $alias,
                        start_line: $start,
                        end_line: $end,
                        instructions: $instructions
                    }]')
                stage_index=$((stage_index + 1))
                instructions='[]'
            fi

            current_image="${BASH_REMATCH[1]}"
            current_alias=""
            stage_start=$line_num

            # Check for AS alias
            if [[ "$line" =~ [Aa][Ss][[:space:]]+([^[:space:]]+) ]]; then
                current_alias="${BASH_REMATCH[1]}"
            fi
        elif [[ -n "$current_image" ]]; then
            # Track instructions in current stage
            local instruction=""
            if [[ "$line" =~ ^[[:space:]]*([A-Z]+)[[:space:]] ]]; then
                instruction="${BASH_REMATCH[1]}"
                instructions=$(echo "$instructions" | jq --arg i "$instruction" '. + [$i]')
            fi
        fi
    done < "$dockerfile"

    # Save last stage
    if [[ -n "$current_image" ]]; then
        stages=$(echo "$stages" | jq \
            --arg image "$current_image" \
            --arg alias "$current_alias" \
            --argjson index "$stage_index" \
            --argjson start "$stage_start" \
            --argjson end "$line_num" \
            --argjson instructions "$instructions" \
            '. + [{
                index: $index,
                image: $image,
                alias: $alias,
                start_line: $start,
                end_line: $end,
                instructions: $instructions
            }]')
    fi

    echo "$stages"
}

# Find COPY --from references
# Usage: find_copy_from_refs "/path/to/Dockerfile"
# Returns JSON array of copy-from references
find_copy_from_refs() {
    local dockerfile="$1"
    local refs='[]'
    local line_num=0

    while IFS= read -r line || [[ -n "$line" ]]; do
        line_num=$((line_num + 1))

        # Match COPY --from=<stage>
        if [[ "$line" =~ [Cc][Oo][Pp][Yy][[:space:]]+--from=([^[:space:]]+) ]]; then
            local source="${BASH_REMATCH[1]}"
            refs=$(echo "$refs" | jq \
                --arg source "$source" \
                --argjson line "$line_num" \
                '. + [{source: $source, line: $line}]')
        fi
    done < "$dockerfile"

    echo "$refs"
}

# Classify stage purpose based on instructions and alias
# Usage: classify_stage '{"alias": "builder", "instructions": ["RUN", "COPY"]}'
classify_stage() {
    local stage_json="$1"

    local alias
    alias=$(echo "$stage_json" | jq -r '.alias // ""')

    local instructions
    instructions=$(echo "$stage_json" | jq -r '.instructions // []')

    # Check alias for hints
    if [[ -n "$alias" ]]; then
        local alias_lower
        alias_lower=$(echo "$alias" | tr '[:upper:]' '[:lower:]')
        if [[ "$alias_lower" =~ build|compile|builder ]]; then
            echo "build"
            return
        elif [[ "$alias_lower" =~ test|testing ]]; then
            echo "test"
            return
        elif [[ "$alias_lower" =~ prod|production|runtime|final|release ]]; then
            echo "runtime"
            return
        elif [[ "$alias_lower" =~ deps|dependencies ]]; then
            echo "dependencies"
            return
        fi
    fi

    # Check instructions for hints
    local has_npm_build has_go_build has_mvn has_cargo has_pip
    has_npm_build=$(echo "$instructions" | jq 'map(select(. == "RUN")) | length')

    # Simple heuristic: stages with many RUN commands are likely build stages
    local run_count
    run_count=$(echo "$instructions" | jq '[.[] | select(. == "RUN")] | length')

    if [[ "$run_count" -gt 3 ]]; then
        echo "build"
    else
        echo "runtime"
    fi
}

# Detect build tools in a stage
# Usage: detect_build_tools "/path/to/Dockerfile" <start_line> <end_line>
detect_build_tools() {
    local dockerfile="$1"
    local start="$2"
    local end="$3"
    local tools='[]'

    local content
    content=$(sed -n "${start},${end}p" "$dockerfile")

    # Check for various build tools
    if echo "$content" | grep -qE "npm run build|npm build|yarn build|pnpm build"; then
        tools=$(echo "$tools" | jq '. + ["npm"]')
    fi
    if echo "$content" | grep -qE "go build|go install"; then
        tools=$(echo "$tools" | jq '. + ["go"]')
    fi
    if echo "$content" | grep -qE "mvn |gradle |gradlew"; then
        tools=$(echo "$tools" | jq '. + ["maven/gradle"]')
    fi
    if echo "$content" | grep -qE "cargo build"; then
        tools=$(echo "$tools" | jq '. + ["cargo"]')
    fi
    if echo "$content" | grep -qE "pip install|poetry install|pipenv"; then
        tools=$(echo "$tools" | jq '. + ["pip"]')
    fi
    if echo "$content" | grep -qE "gcc |g\+\+ |make "; then
        tools=$(echo "$tools" | jq '. + ["c/c++"]')
    fi
    if echo "$content" | grep -qE "dotnet build|dotnet publish"; then
        tools=$(echo "$tools" | jq '. + ["dotnet"]')
    fi

    echo "$tools"
}

# Check for potential artifact leakage
# Usage: check_artifact_leakage '{"stages": [...], "copy_refs": [...]}'
check_artifact_leakage() {
    local analysis_json="$1"
    local issues='[]'

    local stages
    stages=$(echo "$analysis_json" | jq '.stages')

    local stage_count
    stage_count=$(echo "$stages" | jq 'length')

    # Single stage = potential issue if it's a build
    if [[ "$stage_count" -eq 1 ]]; then
        local stage
        stage=$(echo "$stages" | jq '.[0]')
        local purpose
        purpose=$(classify_stage "$stage")

        if [[ "$purpose" == "build" ]]; then
            issues=$(echo "$issues" | jq '. + [{
                "type": "artifact_leakage",
                "severity": "warning",
                "message": "Single-stage build may include build tools and dependencies in final image",
                "recommendation": "Consider multi-stage build to separate build and runtime"
            }]')
        fi
    fi

    # Check for dev dependencies in final stage
    local final_stage
    final_stage=$(echo "$stages" | jq '.[-1]')
    local final_instructions
    final_instructions=$(echo "$final_stage" | jq -r '.instructions | join(" ")')

    # Check if final stage installs dev dependencies (simplified check)
    if echo "$final_instructions" | grep -qE "npm install|yarn install|pip install"; then
        # Check if --production flag is NOT present (dev deps may be included)
        if ! echo "$final_instructions" | grep -qE -- "--production|--prod"; then
            issues=$(echo "$issues" | jq '. + [{
                "type": "dev_dependencies",
                "severity": "warning",
                "message": "Final stage may include development dependencies",
                "recommendation": "Use --production flag or separate requirements files"
            }]')
        fi
    fi

    echo "$issues"
}

# Generate multi-stage recommendations
# Usage: generate_multistage_recommendations <is_multistage> <stage_count> <detected_language>
generate_multistage_recommendations() {
    local is_multistage="$1"
    local stage_count="$2"
    local language="${3:-unknown}"
    local recommendations='[]'

    if [[ "$is_multistage" != "true" ]]; then
        recommendations=$(echo "$recommendations" | jq '. + [{
            "priority": "high",
            "message": "Use multi-stage builds to reduce final image size and attack surface"
        }]')

        # Language-specific recommendations
        case "$language" in
            nodejs)
                recommendations=$(echo "$recommendations" | jq '. + [{
                    "priority": "medium",
                    "message": "For Node.js: Use builder stage for npm install and build, copy only dist/ to runtime"
                }]')
                ;;
            go)
                recommendations=$(echo "$recommendations" | jq '. + [{
                    "priority": "medium",
                    "message": "For Go: Build static binary in builder stage, use scratch or distroless for runtime"
                }]')
                ;;
            java)
                recommendations=$(echo "$recommendations" | jq '. + [{
                    "priority": "medium",
                    "message": "For Java: Build JAR/WAR in builder stage, use JRE-only image for runtime"
                }]')
                ;;
            python)
                recommendations=$(echo "$recommendations" | jq '. + [{
                    "priority": "medium",
                    "message": "For Python: Install dependencies in builder, copy venv to runtime stage"
                }]')
                ;;
            rust)
                recommendations=$(echo "$recommendations" | jq '. + [{
                    "priority": "medium",
                    "message": "For Rust: Build release binary in builder, use scratch for final image"
                }]')
                ;;
        esac
    elif [[ "$stage_count" -eq 2 ]]; then
        # Already multi-stage, check for improvements
        recommendations=$(echo "$recommendations" | jq '. + [{
            "priority": "low",
            "message": "Consider adding a dedicated test stage for CI/CD integration"
        }]')
    fi

    echo "$recommendations"
}

# Analyze multi-stage build patterns
# Usage: analyze_multistage "/path/to/Dockerfile" [detected_language]
analyze_multistage() {
    local dockerfile="$1"
    local language="${2:-unknown}"

    if [[ ! -f "$dockerfile" ]]; then
        jq -n '{
            "error": "File not found",
            "is_multistage": false
        }'
        return 1
    fi

    # Parse stages
    local stages
    stages=$(parse_stages "$dockerfile")

    local stage_count
    stage_count=$(echo "$stages" | jq 'length')

    local is_multistage="false"
    if [[ "$stage_count" -gt 1 ]]; then
        is_multistage="true"
    fi

    # Find COPY --from references
    local copy_refs
    copy_refs=$(find_copy_from_refs "$dockerfile")

    # Classify each stage
    local classified_stages='[]'
    while IFS= read -r stage; do
        [[ -z "$stage" ]] && continue

        local purpose
        purpose=$(classify_stage "$stage")

        local start_line end_line
        start_line=$(echo "$stage" | jq -r '.start_line')
        end_line=$(echo "$stage" | jq -r '.end_line')

        local build_tools
        build_tools=$(detect_build_tools "$dockerfile" "$start_line" "$end_line")

        classified_stages=$(echo "$classified_stages" | jq \
            --argjson stage "$stage" \
            --arg purpose "$purpose" \
            --argjson build_tools "$build_tools" \
            '. + [$stage + {purpose: $purpose, build_tools: $build_tools}]')
    done < <(echo "$stages" | jq -c '.[]')

    # Check for artifact leakage
    local leakage_analysis
    leakage_analysis=$(jq -n --argjson stages "$classified_stages" --argjson refs "$copy_refs" \
        '{stages: $stages, copy_refs: $refs}')
    local leakage_issues
    leakage_issues=$(check_artifact_leakage "$leakage_analysis")

    # Generate recommendations
    local recommendations
    recommendations=$(generate_multistage_recommendations "$is_multistage" "$stage_count" "$language")

    # Build result
    jq -n \
        --arg is_multistage "$is_multistage" \
        --argjson stage_count "$stage_count" \
        --argjson stages "$classified_stages" \
        --argjson copy_refs "$copy_refs" \
        --argjson issues "$leakage_issues" \
        --argjson recommendations "$recommendations" \
        '{
            is_multistage: ($is_multistage == "true"),
            stage_count: $stage_count,
            stages: $stages,
            copy_from_refs: $copy_refs,
            issues: $issues,
            recommendations: $recommendations
        }'
}

# Export functions
export -f parse_stages
export -f find_copy_from_refs
export -f classify_stage
export -f detect_build_tools
export -f check_artifact_leakage
export -f generate_multistage_recommendations
export -f analyze_multistage
