#!/bin/bash
# Package Health Analyzer - Base Scanner
# Copyright (c) 2024 Crash Override Inc
# SPDX-License-Identifier: GPL-3.0

set -euo pipefail

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
UTILS_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Load libraries
source "$SCRIPT_DIR/lib/deps-dev-client.sh"
source "$SCRIPT_DIR/lib/health-scoring.sh"
source "$SCRIPT_DIR/lib/version-analysis.sh"
source "$SCRIPT_DIR/lib/deprecation-checker.sh"

# Default values
REPO=""
ORG=""
SBOM_FILE=""
OUTPUT_FORMAT="json"
VERBOSE=false
ANALYZE_VERSIONS=true
CHECK_DEPRECATION=true
OUTPUT_FILE=""

# Usage information
usage() {
    cat <<EOF
Package Health Analyzer - Base Scanner

Usage: $0 [OPTIONS]

OPTIONS:
    --repo OWNER/REPO          Analyze single repository
    --org ORGANIZATION         Analyze all repositories in organization
    --sbom FILE                Analyze existing SBOM file
    --format FORMAT            Output format: json (default), markdown, table
    --output FILE              Write output to file (default: stdout)
    --no-version-analysis      Skip version inconsistency analysis
    --no-deprecation-check     Skip deprecation checking
    --verbose                  Enable verbose output
    -h, --help                 Show this help message

EXAMPLES:
    # Analyze single repository
    $0 --repo owner/repo

    # Analyze organization
    $0 --org myorg

    # Analyze existing SBOM
    $0 --sbom sbom.json

    # Custom output
    $0 --repo owner/repo --format markdown --output report.md

EOF
    exit 0
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --repo)
                REPO="$2"
                shift 2
                ;;
            --org)
                ORG="$2"
                shift 2
                ;;
            --sbom)
                SBOM_FILE="$2"
                shift 2
                ;;
            --format)
                OUTPUT_FORMAT="$2"
                shift 2
                ;;
            --output)
                OUTPUT_FILE="$2"
                shift 2
                ;;
            --no-version-analysis)
                ANALYZE_VERSIONS=false
                shift
                ;;
            --no-deprecation-check)
                CHECK_DEPRECATION=false
                shift
                ;;
            --verbose)
                VERBOSE=true
                shift
                ;;
            -h|--help)
                usage
                ;;
            *)
                echo "Error: Unknown option: $1" >&2
                usage
                ;;
        esac
    done
}

# Log message if verbose
log() {
    if [ "$VERBOSE" = true ]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*" >&2
    fi
}

# Extract packages from SBOM
# Returns: {"package": "name", "version": "1.0.0", "ecosystem": "npm"}
extract_packages_from_sbom() {
    local sbom_file=$1

    log "Extracting packages from SBOM: $sbom_file"

    # Check if file exists and is valid JSON
    if [ ! -f "$sbom_file" ]; then
        echo "Error: SBOM file not found: $sbom_file" >&2
        return 1
    fi

    if ! jq empty "$sbom_file" 2>/dev/null; then
        echo "Error: SBOM file is not valid JSON: $sbom_file" >&2
        return 1
    fi

    # Detect SBOM format
    local format=$(jq -r 'if .bomFormat then "cyclonedx" elif .spdxVersion then "spdx" else "unknown" end' "$sbom_file" 2>/dev/null || echo "unknown")

    case $format in
        cyclonedx)
            jq -r '
                .components[] |
                {
                    package: .name,
                    version: .version,
                    ecosystem: (
                        if .purl then
                            (.purl | split(":")[1] | split("/")[0])
                        else
                            "unknown"
                        end
                    )
                }
            ' "$sbom_file"
            ;;
        spdx)
            jq -r '
                .packages[] |
                select(.name != null) |
                {
                    package: .name,
                    version: (.versionInfo // "unknown"),
                    ecosystem: (
                        if .externalRefs then
                            (.externalRefs[] | select(.referenceType == "purl") | .referenceLocator | split(":")[0] | sub("pkg:";""))
                        else
                            "unknown"
                        end
                    )
                }
            ' "$sbom_file"
            ;;
        *)
            echo "Error: Unknown or unsupported SBOM format" >&2
            echo "Supported formats: CycloneDX, SPDX" >&2
            return 1
            ;;
    esac
}

# Map ecosystem names to deps.dev system names
map_ecosystem() {
    local ecosystem=$1

    case $ecosystem in
        npm|javascript|node)
            echo "npm"
            ;;
        pypi|python)
            echo "pypi"
            ;;
        cargo|rust|crates.io)
            echo "cargo"
            ;;
        maven|java)
            echo "maven"
            ;;
        go|golang)
            echo "go"
            ;;
        *)
            echo "$ecosystem"
            ;;
    esac
}

# Generate or locate SBOM for repository
generate_sbom_for_repo() {
    local repo=$1

    log "Generating SBOM for $repo"

    # Check if syft is available
    if ! command -v syft &> /dev/null; then
        echo "Error: syft is required but not installed" >&2
        echo "Install: https://github.com/anchore/syft" >&2
        exit 1
    fi

    # Create temp directory for cloning
    local temp_dir=$(mktemp -d)

    # Convert repo URL to git clone format
    local clone_url="$repo"
    if [[ ! "$repo" =~ ^https?:// ]] && [[ ! "$repo" =~ ^git@ ]]; then
        # Assume it's in owner/repo format, convert to HTTPS
        clone_url="https://github.com/$repo"
    fi

    log "Cloning repository from $clone_url to $temp_dir"
    if ! git clone --depth 1 --quiet "$clone_url" "$temp_dir/repo" 2>/dev/null; then
        rm -rf "$temp_dir"
        echo "Error: Failed to clone repository: $clone_url" >&2
        exit 1
    fi

    # Generate SBOM to a persistent temp file (caller will clean up)
    local sbom_file=$(mktemp)
    log "Running syft scan"
    if ! syft scan "$temp_dir/repo" --output cyclonedx-json="$sbom_file" --quiet 2>/dev/null; then
        rm -rf "$temp_dir"
        rm -f "$sbom_file"
        echo "Error: Failed to generate SBOM" >&2
        exit 1
    fi

    # Clean up cloned repo but keep SBOM
    rm -rf "$temp_dir"

    echo "$sbom_file"
}

# Analyze single package
analyze_package() {
    local system=$1
    local package=$2
    local version=$3

    log "Analyzing $package@$version ($system)"

    # Get health analysis
    local health_result=$(analyze_package_health "$system" "$package" "$version" 2>/dev/null || echo '{"error": "analysis_failed"}')

    # Validate health_result is valid JSON
    if ! echo "$health_result" | jq empty 2>/dev/null; then
        log "Warning: Invalid JSON from health analysis for $package, using error placeholder"
        health_result='{"error": "invalid_response", "package": "'$package'", "system": "'$system'", "version": "'$version'"}'
    fi

    # Check deprecation
    local deprecation_result="{}"
    if [ "$CHECK_DEPRECATION" = true ]; then
        deprecation_result=$(comprehensive_deprecation_check "$system" "$package" 2>/dev/null || echo '{"deprecated": false}')

        # Validate deprecation_result is valid JSON
        if ! echo "$deprecation_result" | jq empty 2>/dev/null; then
            log "Warning: Invalid JSON from deprecation check for $package"
            deprecation_result='{"deprecated": false}'
        fi
    fi

    # Combine results safely
    echo "$health_result" | jq --argjson deprecation "$deprecation_result" '. + {deprecation: $deprecation}' 2>/dev/null || \
        jq -n --arg pkg "$package" --arg sys "$system" --arg ver "$version" \
            '{"error": "failed_to_combine_results", "package": $pkg, "system": $sys, "version": $ver}'
}

# Analyze packages from SBOM
analyze_from_sbom() {
    local sbom_file=$1
    local repo_name=${2:-"unknown"}

    log "Starting analysis of SBOM"

    # Extract packages
    local packages=$(extract_packages_from_sbom "$sbom_file")

    # Analyze each package
    local package_results="[]"
    local package_usage="{}"  # For version analysis: {package: [{repo, version}, ...]}

    while IFS= read -r pkg_info; do
        [ -z "$pkg_info" ] && continue

        local package=$(echo "$pkg_info" | jq -r '.package')
        local version=$(echo "$pkg_info" | jq -r '.version')
        local ecosystem=$(echo "$pkg_info" | jq -r '.ecosystem')

        # Map ecosystem to deps.dev system
        local system=$(map_ecosystem "$ecosystem")

        # Skip if invalid
        if [ "$system" = "unknown" ] || [ -z "$package" ] || [ "$version" = "unknown" ]; then
            log "Skipping invalid package: $package"
            continue
        fi

        # Analyze package
        local result=$(analyze_package "$system" "$package" "$version")

        # Validate result before adding
        if echo "$result" | jq empty 2>/dev/null; then
            # Add to results only if valid JSON
            package_results=$(echo "$package_results" | jq --argjson item "$result" '. + [$item]' 2>/dev/null || echo "$package_results")
        else
            log "Warning: Skipping invalid result for $package@$version"
        fi

        # Track for version analysis
        if [ "$ANALYZE_VERSIONS" = true ]; then
            local usage_item=$(jq -n --arg repo "$repo_name" --arg ver "$version" '{repo: $repo, version: $ver}')
            package_usage=$(echo "$package_usage" | jq --arg pkg "$package" --argjson item "$usage_item" '
                .[$pkg] = ((.[$pkg] // []) + [$item])
            ')
        fi

    done < <(echo "$packages" | jq -c '.')

    # Analyze version inconsistencies (if multiple repos)
    local version_analysis="[]"
    if [ "$ANALYZE_VERSIONS" = true ]; then
        version_analysis=$(analyze_all_versions "$package_usage" 2>/dev/null || echo '[]')
    fi

    # Generate summary
    local total_packages=$(echo "$package_results" | jq 'length')
    local deprecated_count=$(echo "$package_results" | jq '[.[] | select(.deprecation.deprecated == true)] | length')
    local low_health_count=$(echo "$package_results" | jq '[.[] | select(.health_score < 60)] | length')
    local critical_count=$(echo "$package_results" | jq '[.[] | select(.health_grade == "Critical")] | length')

    # Build final result
    jq -n \
        --arg timestamp "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        --arg repo "$repo_name" \
        --argjson total "$total_packages" \
        --argjson deprecated "$deprecated_count" \
        --argjson low_health "$low_health_count" \
        --argjson critical "$critical_count" \
        --argjson packages "$package_results" \
        --argjson versions "$version_analysis" \
        '{
            scan_metadata: {
                timestamp: $timestamp,
                repositories_scanned: 1,
                packages_analyzed: $total,
                analyzer_version: "1.0.0",
                analyzer_type: "base"
            },
            summary: {
                total_packages: $total,
                deprecated_packages: $deprecated,
                low_health_packages: $low_health,
                critical_health_packages: $critical,
                version_inconsistencies: ($versions | length)
            },
            repositories: [
                {
                    name: $repo,
                    package_count: $total
                }
            ],
            packages: $packages,
            version_inconsistencies: $versions
        }'
}

# Analyze organization (multiple repositories)
analyze_organization() {
    local org=$1

    log "Analyzing organization: $org"

    # Get all repos in org
    local repos=$(gh repo list "$org" --limit 100 --json name --jq '.[].name')

    if [ -z "$repos" ]; then
        echo "Error: No repositories found for organization $org" >&2
        exit 1
    fi

    log "Found $(echo "$repos" | wc -l) repositories"

    # Analyze each repo and combine results
    local all_packages="[]"
    local all_version_data="{}"
    local repo_count=0

    while IFS= read -r repo_name; do
        [ -z "$repo_name" ] && continue

        log "Processing repository: $repo_name"

        local sbom_file=$(generate_sbom_for_repo "$org/$repo_name")
        local repo_result=$(analyze_from_sbom "$sbom_file" "$org/$repo_name")

        # Extract packages and version data
        local repo_packages=$(echo "$repo_result" | jq -r '.packages')
        all_packages=$(echo "$all_packages $repo_packages" | jq -s 'add | unique_by(.package)')

        # Merge version usage data
        local packages_list=$(echo "$repo_packages" | jq -r '.[].package')
        while IFS= read -r pkg; do
            [ -z "$pkg" ] && continue
            local version=$(echo "$repo_packages" | jq -r --arg p "$pkg" '.[] | select(.package == $p) | .version')

            local usage_item=$(jq -n --arg repo "$org/$repo_name" --arg ver "$version" '{repo: $repo, version: $ver}')
            all_version_data=$(echo "$all_version_data" | jq --arg pkg "$pkg" --argjson item "$usage_item" '
                .[$pkg] = ((.[$pkg] // []) + [$item])
            ')
        done < <(echo "$packages_list")

        ((repo_count++))

    done < <(echo "$repos")

    # Analyze version inconsistencies across all repos
    local version_analysis="[]"
    if [ "$ANALYZE_VERSIONS" = true ]; then
        log "Analyzing version inconsistencies"
        version_analysis=$(analyze_all_versions "$all_version_data" 2>/dev/null || echo '[]')
    fi

    # Generate summary
    local total_packages=$(echo "$all_packages" | jq 'length')
    local deprecated_count=$(echo "$all_packages" | jq '[.[] | select(.deprecation.deprecated == true)] | length')
    local low_health_count=$(echo "$all_packages" | jq '[.[] | select(.health_score < 60)] | length')

    # Build final result
    jq -n \
        --arg timestamp "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        --arg org "$org" \
        --argjson repos "$repo_count" \
        --argjson total "$total_packages" \
        --argjson deprecated "$deprecated_count" \
        --argjson low_health "$low_health_count" \
        --argjson packages "$all_packages" \
        --argjson versions "$version_analysis" \
        '{
            scan_metadata: {
                timestamp: $timestamp,
                organization: $org,
                repositories_scanned: $repos,
                packages_analyzed: $total,
                analyzer_version: "1.0.0",
                analyzer_type: "base"
            },
            summary: {
                total_packages: $total,
                unique_packages: $total,
                deprecated_packages: $deprecated,
                low_health_packages: $low_health,
                version_inconsistencies: ($versions | length)
            },
            packages: $packages,
            version_inconsistencies: $versions
        }'
}

# Format output
format_output() {
    local json_data=$1
    local format=$2

    case $format in
        json)
            echo "$json_data" | jq '.'
            ;;
        markdown)
            format_markdown "$json_data"
            ;;
        table)
            format_table "$json_data"
            ;;
        *)
            echo "Error: Unknown format: $format" >&2
            exit 1
            ;;
    esac
}

# Format as markdown
format_markdown() {
    local json_data=$1

    cat <<EOF
# Package Health Analysis Report

**Generated:** $(echo "$json_data" | jq -r '.scan_metadata.timestamp')
**Repositories Scanned:** $(echo "$json_data" | jq -r '.scan_metadata.repositories_scanned')

## Summary

- **Total Packages:** $(echo "$json_data" | jq -r '.summary.total_packages')
- **Deprecated Packages:** $(echo "$json_data" | jq -r '.summary.deprecated_packages')
- **Low Health Packages:** $(echo "$json_data" | jq -r '.summary.low_health_packages')
- **Version Inconsistencies:** $(echo "$json_data" | jq -r '.summary.version_inconsistencies')

## Packages by Health Grade

EOF

    # Group by health grade
    local critical=$(echo "$json_data" | jq -r '[.packages[] | select(.health_grade == "Critical")] | length')
    local poor=$(echo "$json_data" | jq -r '[.packages[] | select(.health_grade == "Poor")] | length')
    local fair=$(echo "$json_data" | jq -r '[.packages[] | select(.health_grade == "Fair")] | length')
    local good=$(echo "$json_data" | jq -r '[.packages[] | select(.health_grade == "Good")] | length')
    local excellent=$(echo "$json_data" | jq -r '[.packages[] | select(.health_grade == "Excellent")] | length')

    cat <<EOF
| Grade      | Count |
|------------|-------|
| Critical   | $critical |
| Poor       | $poor |
| Fair       | $fair |
| Good       | $good |
| Excellent  | $excellent |

## Deprecated Packages

EOF

    echo "$json_data" | jq -r '
        .packages[] |
        select(.deprecation.deprecated == true) |
        "- **\(.package)** (v\(.version)): \(.deprecation.deprecation_message // "No message provided")"
    '

    if [ "$(echo "$json_data" | jq '.version_inconsistencies | length')" -gt 0 ]; then
        cat <<EOF

## Version Inconsistencies

EOF
        echo "$json_data" | jq -r '
            .version_inconsistencies[] |
            "### \(.package)\n\n" +
            "- **Severity:** \(.analysis.severity)\n" +
            "- **Unique Versions:** \(.analysis.unique_versions)\n" +
            "- **Recommended:** \(.recommendations.target_version)\n" +
            "- **Estimated Effort:** \(.recommendations.estimated_effort_hours) hours\n"
        '
    fi
}

# Format as table
format_table() {
    local json_data=$1

    echo "╔════════════════════════════════════════════════════════════╗"
    echo "║            Package Health Analysis Summary                 ║"
    echo "╠════════════════════════════════════════════════════════════╣"
    printf "║ Total Packages:           %-32s ║\n" "$(echo "$json_data" | jq -r '.summary.total_packages')"
    printf "║ Deprecated:               %-32s ║\n" "$(echo "$json_data" | jq -r '.summary.deprecated_packages')"
    printf "║ Low Health:               %-32s ║\n" "$(echo "$json_data" | jq -r '.summary.low_health_packages')"
    printf "║ Version Inconsistencies:  %-32s ║\n" "$(echo "$json_data" | jq -r '.summary.version_inconsistencies')"
    echo "╚════════════════════════════════════════════════════════════╝"

    echo ""
    echo "Top Issues:"
    echo "$json_data" | jq -r '
        .packages[] |
        select(.health_score < 60 or .deprecation.deprecated == true) |
        "  - \(.package) v\(.version): \(.health_grade) (Score: \(.health_score))"
    ' | head -10
}

# Main execution
main() {
    parse_args "$@"

    # Validate input
    if [ -z "$REPO" ] && [ -z "$ORG" ] && [ -z "$SBOM_FILE" ]; then
        echo "Error: Must specify --repo, --org, or --sbom" >&2
        usage
    fi

    local result=""
    local temp_sbom=""

    # Determine analysis mode
    if [ -n "$SBOM_FILE" ]; then
        result=$(analyze_from_sbom "$SBOM_FILE")
    elif [ -n "$REPO" ]; then
        temp_sbom=$(generate_sbom_for_repo "$REPO")
        trap "rm -f $temp_sbom" EXIT
        result=$(analyze_from_sbom "$temp_sbom" "$REPO")
    elif [ -n "$ORG" ]; then
        result=$(analyze_organization "$ORG")
    fi

    # Format and output
    local formatted=$(format_output "$result" "$OUTPUT_FORMAT")

    if [ -n "$OUTPUT_FILE" ]; then
        echo "$formatted" > "$OUTPUT_FILE"
        log "Report written to $OUTPUT_FILE"
    else
        echo "$formatted"
    fi
}

# Run main function
main "$@"
