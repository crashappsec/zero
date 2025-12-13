#!/usr/bin/env bash
# Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0
#
# Image Scanner
# Integrates with Trivy, Grype, and Syft for container image vulnerability scanning
#
# Usage:
#   source image-scanner.sh
#   scan_image "myapp:latest"

set -euo pipefail

# Check if trivy is available
has_trivy() {
    command -v trivy &>/dev/null
}

# Check if grype is available
has_grype() {
    command -v grype &>/dev/null
}

# Check if syft is available
has_syft() {
    command -v syft &>/dev/null
}

# Check if docker is available
has_docker() {
    command -v docker &>/dev/null
}

# Check if an image exists locally
# Usage: image_exists "myapp:latest"
image_exists() {
    local image="$1"

    if ! has_docker; then
        return 1
    fi

    docker image inspect "$image" &>/dev/null
}

# Scan image with Trivy
# Usage: scan_with_trivy "myapp:latest"
# Returns JSON vulnerability report
scan_with_trivy() {
    local image="$1"
    local severity="${2:-UNKNOWN,LOW,MEDIUM,HIGH,CRITICAL}"

    if ! has_trivy; then
        jq -n '{"error": "trivy not installed", "tool": "trivy"}'
        return 1
    fi

    # Run trivy with JSON output
    local output
    output=$(trivy image \
        --format json \
        --severity "$severity" \
        --quiet \
        "$image" 2>/dev/null) || {
        jq -n --arg img "$image" '{"error": "trivy scan failed", "image": $img, "tool": "trivy"}'
        return 1
    }

    # Transform trivy output to our format
    echo "$output" | jq '{
        tool: "trivy",
        image: .ArtifactName,
        vulnerabilities: (
            [.Results[]? | .Vulnerabilities[]? | {
                id: .VulnerabilityID,
                package: .PkgName,
                installed_version: .InstalledVersion,
                fixed_version: .FixedVersion,
                severity: .Severity,
                title: .Title,
                description: .Description,
                references: .References
            }] // []
        ),
        summary: {
            critical: ([.Results[]?.Vulnerabilities[]? | select(.Severity == "CRITICAL")] | length),
            high: ([.Results[]?.Vulnerabilities[]? | select(.Severity == "HIGH")] | length),
            medium: ([.Results[]?.Vulnerabilities[]? | select(.Severity == "MEDIUM")] | length),
            low: ([.Results[]?.Vulnerabilities[]? | select(.Severity == "LOW")] | length),
            unknown: ([.Results[]?.Vulnerabilities[]? | select(.Severity == "UNKNOWN")] | length)
        }
    }' 2>/dev/null || jq -n '{"error": "failed to parse trivy output", "tool": "trivy"}'
}

# Scan image with Grype
# Usage: scan_with_grype "myapp:latest"
# Returns JSON vulnerability report
scan_with_grype() {
    local image="$1"

    if ! has_grype; then
        jq -n '{"error": "grype not installed", "tool": "grype"}'
        return 1
    fi

    # Run grype with JSON output
    local output
    output=$(grype "$image" -o json 2>/dev/null) || {
        jq -n --arg img "$image" '{"error": "grype scan failed", "image": $img, "tool": "grype"}'
        return 1
    }

    # Transform grype output to our format
    echo "$output" | jq '{
        tool: "grype",
        image: .source.target.userInput,
        vulnerabilities: [
            .matches[]? | {
                id: .vulnerability.id,
                package: .artifact.name,
                installed_version: .artifact.version,
                fixed_version: (.vulnerability.fix.versions[0] // null),
                severity: .vulnerability.severity,
                description: .vulnerability.description,
                references: .vulnerability.urls
            }
        ],
        summary: {
            critical: ([.matches[]? | select(.vulnerability.severity == "Critical")] | length),
            high: ([.matches[]? | select(.vulnerability.severity == "High")] | length),
            medium: ([.matches[]? | select(.vulnerability.severity == "Medium")] | length),
            low: ([.matches[]? | select(.vulnerability.severity == "Low")] | length),
            negligible: ([.matches[]? | select(.vulnerability.severity == "Negligible")] | length)
        }
    }' 2>/dev/null || jq -n '{"error": "failed to parse grype output", "tool": "grype"}'
}

# Generate SBOM with Syft
# Usage: generate_sbom "myapp:latest"
# Returns JSON SBOM
generate_sbom() {
    local image="$1"

    if ! has_syft; then
        jq -n '{"error": "syft not installed", "tool": "syft"}'
        return 1
    fi

    # Run syft with JSON output
    local output
    output=$(syft "$image" -o json 2>/dev/null) || {
        jq -n --arg img "$image" '{"error": "syft scan failed", "image": $img, "tool": "syft"}'
        return 1
    }

    # Transform syft output to summary format
    echo "$output" | jq '{
        tool: "syft",
        image: .source.target.userInput,
        package_count: (.artifacts | length),
        packages: [
            .artifacts[]? | {
                name: .name,
                version: .version,
                type: .type,
                language: .language
            }
        ],
        by_type: (
            .artifacts | group_by(.type) | map({
                type: .[0].type,
                count: length
            })
        )
    }' 2>/dev/null || jq -n '{"error": "failed to parse syft output", "tool": "syft"}'
}

# Scan image with available tools
# Usage: scan_image "myapp:latest"
# Returns combined JSON report
scan_image() {
    local image="$1"
    local trivy_result='null'
    local grype_result='null'
    local sbom_result='null'

    # Check if image exists or can be pulled
    local image_available="false"
    if has_docker; then
        if image_exists "$image" || docker pull "$image" &>/dev/null; then
            image_available="true"
        fi
    fi

    # Scan with trivy if available
    if has_trivy; then
        trivy_result=$(scan_with_trivy "$image" 2>/dev/null || echo 'null')
    fi

    # Scan with grype if available
    if has_grype; then
        grype_result=$(scan_with_grype "$image" 2>/dev/null || echo 'null')
    fi

    # Generate SBOM if syft available
    if has_syft; then
        sbom_result=$(generate_sbom "$image" 2>/dev/null || echo 'null')
    fi

    # Combine results
    local combined_vulns='[]'
    local total_critical=0 total_high=0 total_medium=0 total_low=0

    # Prefer trivy results if available
    if [[ "$trivy_result" != "null" ]] && ! echo "$trivy_result" | jq -e '.error' &>/dev/null; then
        combined_vulns=$(echo "$trivy_result" | jq '.vulnerabilities // []')
        total_critical=$(echo "$trivy_result" | jq '.summary.critical // 0')
        total_high=$(echo "$trivy_result" | jq '.summary.high // 0')
        total_medium=$(echo "$trivy_result" | jq '.summary.medium // 0')
        total_low=$(echo "$trivy_result" | jq '.summary.low // 0')
    elif [[ "$grype_result" != "null" ]] && ! echo "$grype_result" | jq -e '.error' &>/dev/null; then
        combined_vulns=$(echo "$grype_result" | jq '.vulnerabilities // []')
        total_critical=$(echo "$grype_result" | jq '.summary.critical // 0')
        total_high=$(echo "$grype_result" | jq '.summary.high // 0')
        total_medium=$(echo "$grype_result" | jq '.summary.medium // 0')
        total_low=$(echo "$grype_result" | jq '.summary.low // 0')
    fi

    # Get package count from SBOM
    local package_count=0
    if [[ "$sbom_result" != "null" ]] && ! echo "$sbom_result" | jq -e '.error' &>/dev/null; then
        package_count=$(echo "$sbom_result" | jq '.package_count // 0')
    fi

    # Build combined result
    jq -n \
        --arg image "$image" \
        --arg image_available "$image_available" \
        --arg has_trivy "$(has_trivy && echo true || echo false)" \
        --arg has_grype "$(has_grype && echo true || echo false)" \
        --arg has_syft "$(has_syft && echo true || echo false)" \
        --argjson trivy "$trivy_result" \
        --argjson grype "$grype_result" \
        --argjson sbom "$sbom_result" \
        --argjson vulnerabilities "$combined_vulns" \
        --argjson critical "$total_critical" \
        --argjson high "$total_high" \
        --argjson medium "$total_medium" \
        --argjson low "$total_low" \
        --argjson package_count "$package_count" \
        '{
            image: $image,
            image_available: ($image_available == "true"),
            tools_available: {
                trivy: ($has_trivy == "true"),
                grype: ($has_grype == "true"),
                syft: ($has_syft == "true")
            },
            summary: {
                critical: $critical,
                high: $high,
                medium: $medium,
                low: $low,
                total: ($critical + $high + $medium + $low),
                package_count: $package_count
            },
            vulnerabilities: $vulnerabilities,
            raw_results: {
                trivy: $trivy,
                grype: $grype,
                sbom: $sbom
            }
        }'
}

# Get tool availability status
# Usage: get_scanner_status
get_scanner_status() {
    jq -n \
        --arg has_trivy "$(has_trivy && echo true || echo false)" \
        --arg has_grype "$(has_grype && echo true || echo false)" \
        --arg has_syft "$(has_syft && echo true || echo false)" \
        --arg has_docker "$(has_docker && echo true || echo false)" \
        '{
            trivy: {available: ($has_trivy == "true"), purpose: "Vulnerability scanning"},
            grype: {available: ($has_grype == "true"), purpose: "Vulnerability scanning (alternative)"},
            syft: {available: ($has_syft == "true"), purpose: "SBOM generation"},
            docker: {available: ($has_docker == "true"), purpose: "Image inspection"}
        }'
}

# Export functions
export -f has_trivy
export -f has_grype
export -f has_syft
export -f has_docker
export -f image_exists
export -f scan_with_trivy
export -f scan_with_grype
export -f generate_sbom
export -f scan_image
export -f get_scanner_status
