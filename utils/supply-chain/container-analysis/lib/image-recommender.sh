#!/bin/bash
# Container Image Recommender
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0
#
# Recommends secure, standardized container base images based on project requirements.
# Part of the Reliability & Standardization module.
#
# Gold images are loaded from RAG:
#   - rag/supply-chain/hardened-images/providers/google-distroless.json
#   - rag/supply-chain/hardened-images/providers/chainguard.json
#   - rag/supply-chain/hardened-images/providers/official-images.json
#   - rag/supply-chain/hardened-images/deprecated-images.json

set -eo pipefail

# Get script directory and repo root
LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$LIB_DIR/../../../.." && pwd)"
RAG_IMAGES_DIR="$REPO_ROOT/rag/supply-chain/hardened-images"

# Cache for fetched versions (populated dynamically)
DYNAMIC_VERSIONS_FETCHED=false
declare -A LATEST_VERSIONS

#############################################################################
# Dynamic Version Fetching - Get Latest Image Tags from Registries
#############################################################################

# Fetch latest tag from Docker Hub
# Usage: fetch_dockerhub_latest <image>
fetch_dockerhub_latest() {
    local image="$1"
    local namespace="${image%%/*}"
    local repo="${image#*/}"

    # Handle official images (no namespace)
    if [[ "$image" != *"/"* ]]; then
        namespace="library"
        repo="$image"
    fi

    # Query Docker Hub API for tags
    local response=$(curl -s "https://hub.docker.com/v2/repositories/${namespace}/${repo}/tags?page_size=100" 2>/dev/null)

    if [[ -n "$response" ]] && echo "$response" | jq -e '.results' >/dev/null 2>&1; then
        # Find latest non-latest stable tag (prefer alpine, then debian, then version numbers)
        local latest_tag=$(echo "$response" | jq -r '
            .results[] |
            select(.name != "latest") |
            select(.name | test("^[0-9]")) |
            select(.name | test("rc|alpha|beta|dev") | not) |
            .name
        ' 2>/dev/null | head -1)

        if [[ -n "$latest_tag" ]]; then
            echo "$latest_tag"
            return 0
        fi
    fi

    echo ""
    return 1
}

# Fetch latest version for Google Distroless images
# Usage: fetch_gcr_latest <image>
fetch_gcr_latest() {
    local image="$1"

    # gcr.io images typically use debian12 suffix for latest
    # The images are versioned by the debian base, not by tag
    # We'll check if the image exists
    local image_name="${image#gcr.io/distroless/}"

    # Most distroless images use latest or nonroot tags
    # Return the current recommended stable version
    case "$image_name" in
        static*|base*|cc*)
            echo "debian12"
            ;;
        nodejs*)
            # Check for latest Node LTS versions
            echo "22" # Current LTS
            ;;
        python*)
            echo "3.12"
            ;;
        java*)
            echo "21"
            ;;
        *)
            echo "latest"
            ;;
    esac
}

# Fetch latest versions from Chainguard
# Usage: fetch_chainguard_latest <image>
fetch_chainguard_latest() {
    local image="$1"
    local image_name="${image#chainguard/}"
    image_name="${image_name#cgr.dev/chainguard/}"
    image_name="${image_name%%:*}"

    # Chainguard uses :latest which is always current
    # But we can check their catalog for specific versions
    # For now, return latest as Chainguard rebuilds daily
    echo "latest"
}

# Fetch latest Alpine version
fetch_alpine_latest() {
    local response=$(curl -s "https://hub.docker.com/v2/repositories/library/alpine/tags?page_size=20" 2>/dev/null)

    if [[ -n "$response" ]]; then
        local latest=$(echo "$response" | jq -r '
            .results[] |
            select(.name | test("^3\\.[0-9]+$")) |
            .name
        ' 2>/dev/null | sort -V | tail -1)

        if [[ -n "$latest" ]]; then
            echo "$latest"
            return 0
        fi
    fi

    echo "3.20"  # Fallback to known recent version
}

# Fetch latest Node.js LTS version
fetch_node_lts() {
    # Query Node.js release schedule
    local response=$(curl -s "https://nodejs.org/dist/index.json" 2>/dev/null | head -c 50000)

    if [[ -n "$response" ]]; then
        local lts_version=$(echo "$response" | jq -r '
            [.[] | select(.lts != false)] | .[0].version | ltrimstr("v") | split(".")[0]
        ' 2>/dev/null)

        if [[ -n "$lts_version" && "$lts_version" != "null" ]]; then
            echo "$lts_version"
            return 0
        fi
    fi

    echo "22"  # Fallback to known LTS
}

# Fetch latest Python version
fetch_python_latest() {
    local response=$(curl -s "https://hub.docker.com/v2/repositories/library/python/tags?page_size=50" 2>/dev/null)

    if [[ -n "$response" ]]; then
        local latest=$(echo "$response" | jq -r '
            .results[] |
            select(.name | test("^3\\.[0-9]+$")) |
            .name
        ' 2>/dev/null | sort -V | tail -1)

        if [[ -n "$latest" ]]; then
            echo "$latest"
            return 0
        fi
    fi

    echo "3.13"  # Fallback
}

# Fetch latest Go version
fetch_go_latest() {
    local response=$(curl -s "https://go.dev/dl/?mode=json" 2>/dev/null)

    if [[ -n "$response" ]]; then
        local latest=$(echo "$response" | jq -r '.[0].version | ltrimstr("go")' 2>/dev/null)

        if [[ -n "$latest" && "$latest" != "null" ]]; then
            # Return major.minor only
            echo "${latest%.*}"
            return 0
        fi
    fi

    echo "1.23"  # Fallback
}

# Fetch latest Java LTS version
fetch_java_lts() {
    # Eclipse Temurin releases
    local response=$(curl -s "https://api.adoptium.net/v3/info/available_releases" 2>/dev/null)

    if [[ -n "$response" ]]; then
        local lts=$(echo "$response" | jq -r '.most_recent_lts' 2>/dev/null)

        if [[ -n "$lts" && "$lts" != "null" ]]; then
            echo "$lts"
            return 0
        fi
    fi

    echo "21"  # Fallback to known LTS
}

# Fetch all latest versions and cache them
# This is called once when the analyzer starts
fetch_latest_versions() {
    if [[ "$DYNAMIC_VERSIONS_FETCHED" == "true" ]]; then
        return
    fi

    echo -e "\033[0;34mFetching latest container image versions...\033[0m" >&2

    # Fetch in parallel for speed
    LATEST_VERSIONS["alpine"]=$(fetch_alpine_latest)
    LATEST_VERSIONS["node"]=$(fetch_node_lts)
    LATEST_VERSIONS["python"]=$(fetch_python_latest)
    LATEST_VERSIONS["go"]=$(fetch_go_latest)
    LATEST_VERSIONS["java"]=$(fetch_java_lts)

    # Log fetched versions
    echo -e "\033[0;32m✓ Latest versions: Alpine ${LATEST_VERSIONS["alpine"]}, Node ${LATEST_VERSIONS["node"]}, Python ${LATEST_VERSIONS["python"]}, Go ${LATEST_VERSIONS["go"]}, Java ${LATEST_VERSIONS["java"]}\033[0m" >&2

    DYNAMIC_VERSIONS_FETCHED=true
}

# Get dynamically versioned image recommendation
# Usage: get_versioned_image <base_image>
get_versioned_image() {
    local base_image="$1"

    # Ensure versions are fetched
    fetch_latest_versions

    case "$base_image" in
        *alpine*)
            echo "${base_image/alpine:*/alpine:${LATEST_VERSIONS["alpine"]}}"
            ;;
        *node:*alpine*)
            echo "node:${LATEST_VERSIONS["node"]}-alpine"
            ;;
        *node:*)
            echo "node:${LATEST_VERSIONS["node"]}"
            ;;
        *python:*alpine*)
            echo "python:${LATEST_VERSIONS["python"]}-alpine"
            ;;
        *python:*)
            echo "python:${LATEST_VERSIONS["python"]}"
            ;;
        *golang:*|*go:*)
            echo "golang:${LATEST_VERSIONS["go"]}-alpine"
            ;;
        *temurin:*|*java:*|*jdk:*|*jre:*)
            echo "eclipse-temurin:${LATEST_VERSIONS["java"]}-jre-alpine"
            ;;
        gcr.io/distroless/nodejs*)
            echo "gcr.io/distroless/nodejs${LATEST_VERSIONS["node"]}-debian12"
            ;;
        gcr.io/distroless/java*)
            echo "gcr.io/distroless/java${LATEST_VERSIONS["java"]}-debian12"
            ;;
        gcr.io/distroless/python*)
            echo "gcr.io/distroless/python3-debian12"
            ;;
        *)
            echo "$base_image"
            ;;
    esac
}

#############################################################################
# Load Gold Images from RAG
#############################################################################

# Load images from provider JSON files
load_gold_images_from_rag() {
    local language="$1"
    local rag_images=""

    # Load from each provider
    for provider_file in "$RAG_IMAGES_DIR/providers"/*.json; do
        [[ -f "$provider_file" ]] || continue

        # Extract images for the specified language
        local provider_images=$(jq -r --arg lang "$language" '
            .images[] |
            select(.languages as $langs |
                ($langs | length == 0) or
                ($langs | map(ascii_downcase) | index($lang | ascii_downcase))
            ) |
            "\(.name)|\(.stage)|\(.security_rating // "high")|\(.notes)"
        ' "$provider_file" 2>/dev/null || echo "")

        if [[ -n "$provider_images" ]]; then
            if [[ -n "$rag_images" ]]; then
                rag_images+=$'\n'
            fi
            rag_images+="$provider_images"
        fi
    done

    echo "$rag_images"
}

# Load deprecated patterns from RAG
load_deprecated_patterns() {
    local deprecated_file="$RAG_IMAGES_DIR/deprecated-images.json"

    if [[ -f "$deprecated_file" ]]; then
        jq -r '.deprecated_patterns[] | "\(.pattern)|\(.reason)"' "$deprecated_file" 2>/dev/null || echo ""
    else
        echo ""
    fi
}

# Cache for loaded images (populated on first use)
declare -A GOLD_IMAGE_CACHE
RAG_LOADED=false

# Initialize gold images from RAG (called once)
init_gold_images() {
    if [[ "$RAG_LOADED" == "true" ]]; then
        return
    fi

    # Always fetch latest versions first for dynamic recommendations
    fetch_latest_versions

    if [[ -d "$RAG_IMAGES_DIR/providers" ]]; then
        # Load from RAG
        NODEJS_GOLD_IMAGES=$(load_gold_images_from_rag "nodejs")
        PYTHON_GOLD_IMAGES=$(load_gold_images_from_rag "python")
        GO_GOLD_IMAGES=$(load_gold_images_from_rag "go")
        JAVA_GOLD_IMAGES=$(load_gold_images_from_rag "java")
        RUBY_GOLD_IMAGES=$(load_gold_images_from_rag "ruby")
        RUST_GOLD_IMAGES=$(load_gold_images_from_rag "rust")
        DOTNET_GOLD_IMAGES=$(load_gold_images_from_rag "dotnet")
        GENERIC_GOLD_IMAGES=$(load_gold_images_from_rag "")
        DEPRECATED_IMAGES=$(load_deprecated_patterns)
        RAG_LOADED=true
    else
        # Fallback to embedded defaults if RAG not available
        init_default_images
    fi
}

# Fallback default images (used if RAG files not found)
# Uses dynamically fetched versions for up-to-date recommendations
init_default_images() {
    # Fetch latest versions first
    fetch_latest_versions

    local node_ver="${LATEST_VERSIONS["node"]:-22}"
    local python_ver="${LATEST_VERSIONS["python"]:-3.13}"
    local go_ver="${LATEST_VERSIONS["go"]:-1.23}"
    local java_ver="${LATEST_VERSIONS["java"]:-21}"
    local alpine_ver="${LATEST_VERSIONS["alpine"]:-3.20}"

    NODEJS_GOLD_IMAGES="node:${node_ver}-alpine|production|high|Minimal Alpine-based, LTS version
gcr.io/distroless/nodejs${node_ver}-debian12|production|very_high|Google Distroless
chainguard/node:latest|production|very_high|Chainguard hardened image"

    PYTHON_GOLD_IMAGES="python:${python_ver}-alpine|production|high|Minimal Alpine-based
gcr.io/distroless/python3-debian12|production|very_high|Google Distroless
chainguard/python:latest|production|very_high|Chainguard hardened image"

    GO_GOLD_IMAGES="golang:${go_ver}-alpine|build|high|Build stage only
gcr.io/distroless/static-debian12|production|very_high|For static Go binaries
chainguard/go:latest|build|very_high|Chainguard hardened build image"

    JAVA_GOLD_IMAGES="eclipse-temurin:${java_ver}-jre-alpine|production|high|Eclipse Temurin JRE
gcr.io/distroless/java${java_ver}-debian12|production|very_high|Google Distroless Java
chainguard/jre:latest|production|very_high|Chainguard hardened JRE"

    RUBY_GOLD_IMAGES="ruby:3.3-alpine|production|high|Minimal Alpine-based
chainguard/ruby:latest|production|very_high|Chainguard hardened image"

    RUST_GOLD_IMAGES="rust:1.82-alpine|build|high|Build stage only
gcr.io/distroless/static-debian12|production|very_high|For static Rust binaries
scratch|production|very_high|Empty image for static binaries"

    DOTNET_GOLD_IMAGES="mcr.microsoft.com/dotnet/aspnet:9.0-alpine|production|high|ASP.NET runtime
chainguard/dotnet-runtime:latest|production|very_high|Chainguard hardened"

    GENERIC_GOLD_IMAGES="gcr.io/distroless/static-debian12|production|very_high|Static binaries
alpine:${alpine_ver}|production|high|Minimal general purpose
chainguard/static:latest|production|very_high|Chainguard static base"

    DEPRECATED_IMAGES="*:latest|Use specific version tag
openjdk:*|Use eclipse-temurin instead
*-stretch*|Debian Stretch is EOL
*-buster*|Debian Buster approaching EOL
*-bullseye*|Debian Bullseye approaching EOL - use bookworm
ubuntu:18.04|Ubuntu 18.04 is EOL
ubuntu:20.04|Ubuntu 20.04 approaching EOL
centos:*|CentOS is EOL
node:16*|Node.js 16 is EOL
node:18*|Node.js 18 approaching EOL - upgrade to ${node_ver}
python:3.8*|Python 3.8 is EOL
python:3.9*|Python 3.9 approaching EOL"

    RAG_LOADED=true
}

# Security anti-patterns (kept inline as they're detection patterns, not data)
SECURITY_ANTIPATTERNS="*:latest|Unpinned version is a security risk
FROM scratch AS|Multi-stage builds recommended
apt-get install|Pin package versions
npm install -g|Global installs may have permission issues
pip install|Use requirements.txt with pinned versions
curl.*|.*sh|Piping to shell is risky"

#############################################################################
# Detection Functions
#############################################################################

# Detect project language/framework from files
# Usage: detect_project_type <project_dir>
detect_project_type() {
    local project_dir="$1"
    local detected=()

    if [[ ! -d "$project_dir" ]]; then
        echo '{"error": "directory_not_found"}'
        return 1
    fi

    # Node.js detection
    if [[ -f "$project_dir/package.json" ]]; then
        local framework="nodejs"
        # Check for specific frameworks
        if grep -q '"next"' "$project_dir/package.json" 2>/dev/null; then
            framework="nextjs"
        elif grep -q '"nuxt"' "$project_dir/package.json" 2>/dev/null; then
            framework="nuxt"
        elif grep -q '"express"' "$project_dir/package.json" 2>/dev/null; then
            framework="express"
        elif grep -q '"fastify"' "$project_dir/package.json" 2>/dev/null; then
            framework="fastify"
        elif grep -q '"nest"' "$project_dir/package.json" 2>/dev/null; then
            framework="nestjs"
        fi
        detected+=("$framework")
    fi

    # Python detection
    if [[ -f "$project_dir/requirements.txt" || -f "$project_dir/pyproject.toml" || -f "$project_dir/setup.py" ]]; then
        local framework="python"
        if grep -rq "django" "$project_dir" --include="*.txt" --include="*.toml" 2>/dev/null; then
            framework="django"
        elif grep -rq "flask" "$project_dir" --include="*.txt" --include="*.toml" 2>/dev/null; then
            framework="flask"
        elif grep -rq "fastapi" "$project_dir" --include="*.txt" --include="*.toml" 2>/dev/null; then
            framework="fastapi"
        fi
        detected+=("$framework")
    fi

    # Go detection
    if [[ -f "$project_dir/go.mod" ]]; then
        detected+=("go")
    fi

    # Java/Maven detection
    if [[ -f "$project_dir/pom.xml" ]]; then
        local framework="java"
        if grep -q "spring-boot" "$project_dir/pom.xml" 2>/dev/null; then
            framework="spring-boot"
        fi
        detected+=("$framework")
    fi

    # Java/Gradle detection
    if [[ -f "$project_dir/build.gradle" || -f "$project_dir/build.gradle.kts" ]]; then
        local framework="java"
        if grep -rq "spring" "$project_dir" --include="*.gradle*" 2>/dev/null; then
            framework="spring-boot"
        fi
        detected+=("$framework")
    fi

    # Ruby detection
    if [[ -f "$project_dir/Gemfile" ]]; then
        local framework="ruby"
        if grep -q "rails" "$project_dir/Gemfile" 2>/dev/null; then
            framework="rails"
        fi
        detected+=("$framework")
    fi

    # Rust detection
    if [[ -f "$project_dir/Cargo.toml" ]]; then
        detected+=("rust")
    fi

    # .NET detection
    if ls "$project_dir"/*.csproj 1>/dev/null 2>&1 || ls "$project_dir"/*.fsproj 1>/dev/null 2>&1; then
        detected+=("dotnet")
    fi

    # Convert to JSON array
    if [[ ${#detected[@]} -eq 0 ]]; then
        echo '{"detected": [], "primary": null}'
    else
        local detected_json=$(printf '%s\n' "${detected[@]}" | jq -R . | jq -s '.')
        local primary="${detected[0]}"
        echo "{\"detected\": $detected_json, \"primary\": \"$primary\"}"
    fi
}

# Parse Dockerfile to extract base images
# Usage: parse_dockerfile <dockerfile_path>
parse_dockerfile() {
    local dockerfile="$1"

    if [[ ! -f "$dockerfile" ]]; then
        echo '{"error": "dockerfile_not_found"}'
        return 1
    fi

    local images=()
    local stages=()
    local current_stage=""

    while IFS= read -r line; do
        # Skip comments and empty lines
        [[ "$line" =~ ^[[:space:]]*# ]] && continue
        [[ -z "$line" ]] && continue

        # Match FROM instructions
        if [[ "$line" =~ ^[[:space:]]*FROM[[:space:]]+([^[:space:]]+)([[:space:]]+[aA][sS][[:space:]]+([^[:space:]]+))? ]]; then
            local image="${BASH_REMATCH[1]}"
            local stage="${BASH_REMATCH[3]:-}"

            # Handle ARG-based images
            if [[ "$image" =~ ^\$ ]]; then
                image="(dynamic:$image)"
            fi

            images+=("$image")
            if [[ -n "$stage" ]]; then
                stages+=("$stage:$image")
            fi
        fi
    done < "$dockerfile"

    local images_json=$(printf '%s\n' "${images[@]}" 2>/dev/null | jq -R . | jq -s '.' || echo "[]")
    local stages_json=$(printf '%s\n' "${stages[@]}" 2>/dev/null | jq -R . | jq -s '.' || echo "[]")

    local final_image=""
    if [[ ${#images[@]} -gt 0 ]]; then
        final_image="${images[-1]}"
    fi

    echo "{
        \"base_images\": $images_json,
        \"stages\": $stages_json,
        \"final_image\": \"$final_image\",
        \"multi_stage\": $([ ${#images[@]} -gt 1 ] && echo "true" || echo "false")
    }" | jq '.'
}

#############################################################################
# Recommendation Functions
#############################################################################

# Get gold images for a language/framework
# Usage: get_gold_images <language>
get_gold_images() {
    local language="$1"
    local images=""

    # Initialize gold images from RAG on first call
    init_gold_images

    case "$language" in
        nodejs|nextjs|nuxt|express|fastify|nestjs)
            images="$NODEJS_GOLD_IMAGES"
            ;;
        python|django|flask|fastapi)
            images="$PYTHON_GOLD_IMAGES"
            ;;
        go|golang)
            images="$GO_GOLD_IMAGES"
            ;;
        java|spring-boot)
            images="$JAVA_GOLD_IMAGES"
            ;;
        ruby|rails)
            images="$RUBY_GOLD_IMAGES"
            ;;
        rust)
            images="$RUST_GOLD_IMAGES"
            ;;
        dotnet|csharp|fsharp)
            images="$DOTNET_GOLD_IMAGES"
            ;;
        *)
            images="$GENERIC_GOLD_IMAGES"
            ;;
    esac

    # Parse into JSON
    local result="[]"
    while IFS= read -r line; do
        [[ -z "$line" ]] && continue
        local image=$(echo "$line" | cut -d'|' -f1)
        local stage=$(echo "$line" | cut -d'|' -f2)
        local rating=$(echo "$line" | cut -d'|' -f3)
        local notes=$(echo "$line" | cut -d'|' -f4)

        result=$(echo "$result" | jq --arg img "$image" --arg stg "$stage" --arg rat "$rating" --arg note "$notes" \
            '. + [{"image": $img, "stage": $stg, "security_rating": $rat, "notes": $note}]')
    done <<< "$images"

    echo "$result"
}

# Check if an image is deprecated or has security issues
# Usage: check_image_issues <image>
check_image_issues() {
    local image="$1"
    local issues=()

    # Initialize from RAG on first call
    init_gold_images

    # Check deprecated images
    while IFS= read -r pattern; do
        [[ -z "$pattern" ]] && continue
        local img_pattern=$(echo "$pattern" | cut -d'|' -f1)
        local reason=$(echo "$pattern" | cut -d'|' -f2)

        # Convert glob pattern to regex
        local regex=$(echo "$img_pattern" | sed 's/\*/.*/')

        if [[ "$image" =~ $regex ]]; then
            issues+=("{\"type\": \"deprecated\", \"reason\": \"$reason\"}")
        fi
    done <<< "$DEPRECATED_IMAGES"

    # Check for :latest tag
    if [[ "$image" == *":latest" || ! "$image" == *":"* ]]; then
        issues+=("{\"type\": \"unpinned_version\", \"reason\": \"Use specific version tag instead of latest\"}")
    fi

    # Check for known vulnerable base OS versions
    if [[ "$image" == *"stretch"* ]]; then
        issues+=("{\"type\": \"eol_base\", \"reason\": \"Debian Stretch is End of Life\"}")
    fi
    if [[ "$image" == *"buster"* ]]; then
        issues+=("{\"type\": \"eol_warning\", \"reason\": \"Debian Buster is approaching End of Life\"}")
    fi
    if [[ "$image" == *"ubuntu:18"* || "$image" == *"ubuntu:16"* ]]; then
        issues+=("{\"type\": \"eol_base\", \"reason\": \"Ubuntu version is End of Life\"}")
    fi

    if [[ ${#issues[@]} -gt 0 ]]; then
        printf '%s\n' "${issues[@]}" | jq -s '.'
    else
        echo "[]"
    fi
}

# Generate image recommendation based on current image
# Usage: recommend_replacement <current_image> <language>
recommend_replacement() {
    local current_image="$1"
    local language="${2:-}"

    local issues=$(check_image_issues "$current_image")
    local issue_count=$(echo "$issues" | jq 'length')

    # Detect language from image if not provided
    if [[ -z "$language" ]]; then
        if [[ "$current_image" == *"node"* ]]; then
            language="nodejs"
        elif [[ "$current_image" == *"python"* ]]; then
            language="python"
        elif [[ "$current_image" == *"golang"* || "$current_image" == *"go:"* ]]; then
            language="go"
        elif [[ "$current_image" == *"java"* || "$current_image" == *"jdk"* || "$current_image" == *"jre"* || "$current_image" == *"temurin"* ]]; then
            language="java"
        elif [[ "$current_image" == *"ruby"* ]]; then
            language="ruby"
        elif [[ "$current_image" == *"rust"* ]]; then
            language="rust"
        elif [[ "$current_image" == *"dotnet"* || "$current_image" == *"aspnet"* ]]; then
            language="dotnet"
        fi
    fi

    local gold_images=$(get_gold_images "$language")
    local recommended=$(echo "$gold_images" | jq '.[0]')

    echo "{
        \"current_image\": \"$current_image\",
        \"detected_language\": \"$language\",
        \"issues\": $issues,
        \"has_issues\": $([ $issue_count -gt 0 ] && echo "true" || echo "false"),
        \"recommended_images\": $gold_images,
        \"primary_recommendation\": $recommended
    }" | jq '.'
}

#############################################################################
# Analysis Functions
#############################################################################

# Analyze a project's Dockerfile and provide recommendations
# Usage: analyze_dockerfile <project_dir>
analyze_dockerfile() {
    local project_dir="$1"
    local dockerfile="$project_dir/Dockerfile"

    # Check for Dockerfile variants
    if [[ ! -f "$dockerfile" ]]; then
        if [[ -f "$project_dir/dockerfile" ]]; then
            dockerfile="$project_dir/dockerfile"
        elif [[ -f "$project_dir/Containerfile" ]]; then
            dockerfile="$project_dir/Containerfile"
        else
            echo '{"error": "no_dockerfile_found", "recommendation": "Consider adding a Dockerfile for containerization"}'
            return 1
        fi
    fi

    # Detect project type
    local project_type=$(detect_project_type "$project_dir")
    local primary_language=$(echo "$project_type" | jq -r '.primary // ""')

    # Parse Dockerfile
    local parsed=$(parse_dockerfile "$dockerfile")
    local final_image=$(echo "$parsed" | jq -r '.final_image')
    local base_images=$(echo "$parsed" | jq -r '.base_images')
    local is_multi_stage=$(echo "$parsed" | jq -r '.multi_stage')

    # Analyze each base image
    local image_analyses="[]"
    while IFS= read -r image; do
        [[ -z "$image" || "$image" == "null" ]] && continue
        local analysis=$(recommend_replacement "$image" "$primary_language")
        image_analyses=$(echo "$image_analyses" | jq --argjson a "$analysis" '. + [$a]')
    done < <(echo "$base_images" | jq -r '.[]')

    # Overall recommendations
    local recommendations=()
    local risk_level="low"

    # Check for multi-stage build
    if [[ "$is_multi_stage" != "true" ]]; then
        recommendations+=("Consider using multi-stage builds to reduce final image size")
    fi

    # Check final image issues
    local final_issues=$(check_image_issues "$final_image")
    local final_issue_count=$(echo "$final_issues" | jq 'length')

    if [[ $final_issue_count -gt 0 ]]; then
        risk_level="medium"
        if echo "$final_issues" | jq -e '.[] | select(.type == "eol_base")' >/dev/null 2>&1; then
            risk_level="high"
            recommendations+=("CRITICAL: Final image uses End of Life base - upgrade immediately")
        fi
        if echo "$final_issues" | jq -e '.[] | select(.type == "unpinned_version")' >/dev/null 2>&1; then
            recommendations+=("Pin image version for reproducible builds")
        fi
    fi

    # Get gold image recommendation
    local gold_images=$(get_gold_images "$primary_language")
    local primary_gold=$(echo "$gold_images" | jq '.[0]')

    local recommendations_json=$(printf '%s\n' "${recommendations[@]}" 2>/dev/null | jq -R . | jq -s '.' || echo "[]")

    echo "{
        \"project_dir\": \"$project_dir\",
        \"dockerfile\": \"$dockerfile\",
        \"project_type\": $project_type,
        \"current_images\": $parsed,
        \"image_analyses\": $image_analyses,
        \"risk_level\": \"$risk_level\",
        \"recommendations\": $recommendations_json,
        \"gold_images\": {
            \"language\": \"$primary_language\",
            \"recommended\": $gold_images
        }
    }" | jq '.'
}

# Batch analyze multiple projects
# Usage: analyze_projects <projects_json>
# Input: ["/path/to/project1", "/path/to/project2", ...]
analyze_projects() {
    local projects_json="$1"
    local results="[]"

    while IFS= read -r project_dir; do
        [[ -z "$project_dir" || "$project_dir" == "null" ]] && continue
        local analysis=$(analyze_dockerfile "$project_dir" 2>/dev/null || echo "{\"project_dir\": \"$project_dir\", \"error\": \"analysis_failed\"}")
        results=$(echo "$results" | jq --argjson a "$analysis" '. + [$a]')
    done < <(echo "$projects_json" | jq -r '.[]')

    echo "$results"
}

# Generate summary report
# Usage: generate_container_report <projects_json>
generate_container_report() {
    local projects_json="$1"

    local results=$(analyze_projects "$projects_json")
    local total=$(echo "$results" | jq 'length')
    local with_issues=$(echo "$results" | jq '[.[] | select(.risk_level != "low" and .risk_level != null)] | length')
    local high_risk=$(echo "$results" | jq '[.[] | select(.risk_level == "high")] | length')
    local medium_risk=$(echo "$results" | jq '[.[] | select(.risk_level == "medium")] | length')

    echo "{
        \"summary\": {
            \"total_projects\": $total,
            \"projects_with_issues\": $with_issues,
            \"risk_breakdown\": {
                \"high\": $high_risk,
                \"medium\": $medium_risk,
                \"low\": $((total - high_risk - medium_risk))
            }
        },
        \"projects\": $results
    }" | jq '.'
}

#############################################################################
# Export Functions
#############################################################################

export -f fetch_latest_versions
export -f get_versioned_image
export -f detect_project_type
export -f parse_dockerfile
export -f get_gold_images
export -f check_image_issues
export -f recommend_replacement
export -f analyze_dockerfile
export -f analyze_projects
export -f generate_container_report
