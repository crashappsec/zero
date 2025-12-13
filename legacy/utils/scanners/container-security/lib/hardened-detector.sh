#!/usr/bin/env bash
# Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0
#
# Hardened Image Detector
# Detects Chainguard, Distroless, Alpine, and other hardened base images
#
# Usage:
#   source hardened-detector.sh
#   analyze_image_hardening "node:18-alpine"

set -euo pipefail

# Chainguard image prefixes
CHAINGUARD_PREFIXES=(
    "cgr.dev/chainguard/"
    "images.chainguard.dev/"
    "chainguard/"
)

# Distroless image prefixes
DISTROLESS_PREFIXES=(
    "gcr.io/distroless/"
    "distroless/"
)

# Detect if an image is from Chainguard
# Usage: is_chainguard "cgr.dev/chainguard/node"
is_chainguard() {
    local image="$1"

    for prefix in "${CHAINGUARD_PREFIXES[@]}"; do
        if [[ "$image" == "$prefix"* ]]; then
            return 0
        fi
    done
    return 1
}

# Detect if an image is Distroless
# Usage: is_distroless "gcr.io/distroless/nodejs18-debian12"
is_distroless() {
    local image="$1"

    for prefix in "${DISTROLESS_PREFIXES[@]}"; do
        if [[ "$image" == "$prefix"* ]]; then
            return 0
        fi
    done
    return 1
}

# Detect if an image is Alpine-based
# Usage: is_alpine "node:18-alpine"
is_alpine() {
    local image="$1"

    # Check for -alpine suffix or alpine: base
    if [[ "$image" =~ -alpine ]] || [[ "$image" == alpine:* ]] || [[ "$image" == alpine ]]; then
        return 0
    fi
    return 1
}

# Detect if an image is a slim variant
# Usage: is_slim "python:3.11-slim"
is_slim() {
    local image="$1"

    if [[ "$image" =~ -slim ]] || [[ "$image" =~ -slim- ]]; then
        return 0
    fi
    return 1
}

# Detect if an image uses a scratch base
# Usage: is_scratch "scratch"
is_scratch() {
    local image="$1"

    if [[ "$image" == "scratch" ]]; then
        return 0
    fi
    return 1
}

# Classify the hardening level of an image
# Usage: classify_image "node:18-alpine"
# Returns: chainguard, distroless, scratch, alpine, slim, standard
classify_image() {
    local image="$1"

    if is_chainguard "$image"; then
        echo "chainguard"
    elif is_distroless "$image"; then
        echo "distroless"
    elif is_scratch "$image"; then
        echo "scratch"
    elif is_alpine "$image"; then
        echo "alpine"
    elif is_slim "$image"; then
        echo "slim"
    else
        echo "standard"
    fi
}

# Get security rating for image type
# Usage: get_security_rating "chainguard"
get_security_rating() {
    local image_type="$1"

    case "$image_type" in
        chainguard) echo "very_high" ;;
        distroless) echo "very_high" ;;
        scratch)    echo "very_high" ;;
        alpine)     echo "high" ;;
        slim)       echo "medium" ;;
        *)          echo "low" ;;
    esac
}

# Detect the language/runtime from an image name
# Usage: detect_language "node:18-alpine"
detect_language() {
    local image="$1"

    # Common patterns
    if [[ "$image" =~ ^node:|nodejs|/node ]]; then
        echo "nodejs"
    elif [[ "$image" =~ ^python:|/python ]]; then
        echo "python"
    elif [[ "$image" =~ ^golang:|^go:|/go: ]]; then
        echo "go"
    elif [[ "$image" =~ ^openjdk:|^java:|^eclipse-temurin:|/jre|/jdk ]]; then
        echo "java"
    elif [[ "$image" =~ ^ruby:|/ruby ]]; then
        echo "ruby"
    elif [[ "$image" =~ ^php:|/php ]]; then
        echo "php"
    elif [[ "$image" =~ ^rust:|/rust ]]; then
        echo "rust"
    elif [[ "$image" =~ ^dotnet|^mcr.microsoft.com/dotnet ]]; then
        echo "dotnet"
    elif [[ "$image" =~ ^nginx:|/nginx ]]; then
        echo "nginx"
    elif [[ "$image" =~ ^httpd:|^apache:|/httpd ]]; then
        echo "apache"
    elif [[ "$image" =~ ^postgres:|^postgresql:|/postgres ]]; then
        echo "postgres"
    elif [[ "$image" =~ ^mysql:|^mariadb:|/mysql ]]; then
        echo "mysql"
    elif [[ "$image" =~ ^redis:|/redis ]]; then
        echo "redis"
    elif [[ "$image" =~ ^mongo:|/mongo ]]; then
        echo "mongodb"
    else
        echo "unknown"
    fi
}

# Get hardened alternatives for a language
# Usage: get_hardened_alternatives "nodejs"
get_hardened_alternatives() {
    local language="$1"

    case "$language" in
        nodejs)
            jq -n '[
                {"image": "cgr.dev/chainguard/node:latest", "type": "chainguard", "security_benefit": "Zero known CVEs, minimal attack surface, daily rebuilds"},
                {"image": "gcr.io/distroless/nodejs22-debian12", "type": "distroless", "security_benefit": "No shell, no package manager, reduced attack surface"},
                {"image": "node:22-alpine", "type": "alpine", "security_benefit": "Smaller image size, fewer packages than debian"}
            ]'
            ;;
        python)
            jq -n '[
                {"image": "cgr.dev/chainguard/python:latest", "type": "chainguard", "security_benefit": "Zero known CVEs, minimal attack surface"},
                {"image": "gcr.io/distroless/python3-debian12", "type": "distroless", "security_benefit": "No shell, no package manager"},
                {"image": "python:3.12-alpine", "type": "alpine", "security_benefit": "Smaller image size"}
            ]'
            ;;
        go)
            jq -n '[
                {"image": "cgr.dev/chainguard/go:latest", "type": "chainguard", "security_benefit": "Zero known CVEs for build stage"},
                {"image": "gcr.io/distroless/static-debian12", "type": "distroless", "security_benefit": "Minimal runtime for static Go binaries"},
                {"image": "scratch", "type": "scratch", "security_benefit": "Absolutely minimal - just your binary"}
            ]'
            ;;
        java)
            jq -n '[
                {"image": "cgr.dev/chainguard/jre:latest", "type": "chainguard", "security_benefit": "Zero known CVEs, minimal JRE"},
                {"image": "gcr.io/distroless/java21-debian12", "type": "distroless", "security_benefit": "No shell, no package manager"},
                {"image": "eclipse-temurin:21-jre-alpine", "type": "alpine", "security_benefit": "Smaller image size"}
            ]'
            ;;
        rust)
            jq -n '[
                {"image": "cgr.dev/chainguard/rust:latest", "type": "chainguard", "security_benefit": "Zero known CVEs for build stage"},
                {"image": "gcr.io/distroless/cc-debian12", "type": "distroless", "security_benefit": "Minimal C runtime for Rust binaries"},
                {"image": "scratch", "type": "scratch", "security_benefit": "Minimal runtime for static binaries"}
            ]'
            ;;
        ruby)
            jq -n '[
                {"image": "cgr.dev/chainguard/ruby:latest", "type": "chainguard", "security_benefit": "Zero known CVEs"},
                {"image": "ruby:3.3-alpine", "type": "alpine", "security_benefit": "Smaller image size"}
            ]'
            ;;
        php)
            jq -n '[
                {"image": "cgr.dev/chainguard/php:latest", "type": "chainguard", "security_benefit": "Zero known CVEs"},
                {"image": "php:8.3-alpine", "type": "alpine", "security_benefit": "Smaller image size"}
            ]'
            ;;
        nginx)
            jq -n '[
                {"image": "cgr.dev/chainguard/nginx:latest", "type": "chainguard", "security_benefit": "Zero known CVEs"},
                {"image": "nginx:alpine", "type": "alpine", "security_benefit": "Smaller image size"}
            ]'
            ;;
        *)
            jq -n '[
                {"image": "cgr.dev/chainguard/static:latest", "type": "chainguard", "security_benefit": "Minimal static base image"},
                {"image": "gcr.io/distroless/static-debian12", "type": "distroless", "security_benefit": "Minimal runtime base"}
            ]'
            ;;
    esac
}

# Analyze a single image for hardening status
# Usage: analyze_image_hardening "node:18-alpine"
analyze_image_hardening() {
    local image="$1"

    local image_type
    image_type=$(classify_image "$image")

    local security_rating
    security_rating=$(get_security_rating "$image_type")

    local language
    language=$(detect_language "$image")

    local is_hardened="false"
    if [[ "$image_type" == "chainguard" ]] || [[ "$image_type" == "distroless" ]] || [[ "$image_type" == "scratch" ]]; then
        is_hardened="true"
    fi

    local alternatives
    alternatives=$(get_hardened_alternatives "$language")

    # Filter out current image type from alternatives if already using hardened
    if [[ "$is_hardened" == "true" ]]; then
        alternatives=$(echo "$alternatives" | jq --arg type "$image_type" '[.[] | select(.type != $type)]')
    fi

    jq -n \
        --arg image "$image" \
        --arg image_type "$image_type" \
        --arg security_rating "$security_rating" \
        --arg language "$language" \
        --arg is_hardened "$is_hardened" \
        --arg is_chainguard "$(is_chainguard "$image" && echo true || echo false)" \
        --arg is_distroless "$(is_distroless "$image" && echo true || echo false)" \
        --arg is_alpine "$(is_alpine "$image" && echo true || echo false)" \
        --arg is_slim "$(is_slim "$image" && echo true || echo false)" \
        --arg is_scratch "$(is_scratch "$image" && echo true || echo false)" \
        --argjson alternatives "$alternatives" \
        '{
            image: $image,
            classification: $image_type,
            security_rating: $security_rating,
            detected_language: $language,
            is_hardened: ($is_hardened == "true"),
            details: {
                is_chainguard: ($is_chainguard == "true"),
                is_distroless: ($is_distroless == "true"),
                is_alpine: ($is_alpine == "true"),
                is_slim: ($is_slim == "true"),
                is_scratch: ($is_scratch == "true")
            },
            recommended_alternatives: $alternatives
        }'
}

# Analyze multiple images
# Usage: analyze_images_hardening '["node:18", "python:3.11-slim"]'
analyze_images_hardening() {
    local images_json="$1"
    local results='[]'

    while IFS= read -r image; do
        [[ -z "$image" ]] && continue
        local analysis
        analysis=$(analyze_image_hardening "$image")
        results=$(echo "$results" | jq --argjson a "$analysis" '. + [$a]')
    done < <(echo "$images_json" | jq -r '.[]')

    echo "$results"
}

# Calculate overall hardening score (0-100)
# Usage: calculate_hardening_score '["chainguard", "alpine", "standard"]'
calculate_hardening_score() {
    local types_json="$1"

    # Score weights
    local total=0
    local count=0

    while IFS= read -r type; do
        [[ -z "$type" ]] && continue
        count=$((count + 1))

        case "$type" in
            chainguard) total=$((total + 100)) ;;
            distroless) total=$((total + 100)) ;;
            scratch)    total=$((total + 100)) ;;
            alpine)     total=$((total + 70)) ;;
            slim)       total=$((total + 50)) ;;
            *)          total=$((total + 20)) ;;
        esac
    done < <(echo "$types_json" | jq -r '.[]')

    if [[ "$count" -eq 0 ]]; then
        echo "0"
        return
    fi

    echo $((total / count))
}

# Export functions
export -f is_chainguard
export -f is_distroless
export -f is_alpine
export -f is_slim
export -f is_scratch
export -f classify_image
export -f get_security_rating
export -f detect_language
export -f get_hardened_alternatives
export -f analyze_image_hardening
export -f analyze_images_hardening
export -f calculate_hardening_score
