#!/bin/bash
# Package Health Analyser - Base Scanner
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0

set -euo pipefail

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
UTILS_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
REPO_ROOT="$(cd "$UTILS_ROOT/.." && pwd)"

# Load global libraries
source "$REPO_ROOT/lib/sbom.sh"

# Load local libraries
source "$SCRIPT_DIR/lib/deps-dev-client.sh"
source "$SCRIPT_DIR/lib/health-scoring.sh"
source "$SCRIPT_DIR/lib/version-analysis.sh"
source "$SCRIPT_DIR/lib/deprecation-checker.sh"

# Default values
REPO=""
ORG=""
SBOM_FILE=""
LOCAL_PATH=""
OUTPUT_FORMAT="markdown"
VERBOSE=false
ANALYZE_VERSIONS=true
CHECK_DEPRECATION=true
OUTPUT_FILE=""
ANTHROPIC_API_KEY="${ANTHROPIC_API_KEY:-}"
# Claude enabled by default if API key is set
USE_CLAUDE=false
if [[ -n "$ANTHROPIC_API_KEY" ]]; then
    USE_CLAUDE=true
fi
COMPARE_MODE=false
PARALLEL=true
BATCH_SIZE=1000  # Process in batches to avoid overwhelming the API

# Track temporary directories for cleanup
TEMP_DIRS=()

# Cleanup function
cleanup() {
    if [[ ${#TEMP_DIRS[@]} -gt 0 ]]; then
        for temp_dir in "${TEMP_DIRS[@]}"; do
            if [[ -n "$temp_dir" ]] && [[ -d "$temp_dir" ]]; then
                rm -rf "$temp_dir"
            fi
        done
    fi
}

# Ensure cleanup on script exit (normal, error, or interrupt)
trap cleanup EXIT

# Usage information
usage() {
    cat <<EOF
Package Health Analyser - Base Scanner

Usage: $0 [OPTIONS]

OPTIONS:
    --repo OWNER/REPO          Analyze single repository
    --org ORGANIZATION         Analyze all repositories in organization
    --sbom FILE                Analyze existing SBOM file
    --local-path PATH          Use pre-cloned repository at PATH (skips cloning)
    --format FORMAT            Output format: json, markdown (default), table
    --output FILE              Write output to file (default: stdout)
    --no-version-analysis      Skip version inconsistency analysis
    --no-deprecation-check     Skip deprecation checking
    --claude                   Use Claude AI for advanced analysis (requires ANTHROPIC_API_KEY)
    --compare                  Run both basic and Claude modes side-by-side for comparison
    -k, --api-key KEY          Anthropic API key (or set ANTHROPIC_API_KEY env var)
    --parallel                 Enable batch API processing (faster, recommended)
    --verbose                  Enable verbose output
    -h, --help                 Show this help message

EXAMPLES:
    # Analyze single repository
    $0 --repo owner/repo

    # Analyze with batch API (faster, recommended)
    $0 --repo owner/repo --parallel

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
            --sbom|--sbom-file)
                SBOM_FILE="$2"
                shift 2
                ;;
            --local-path)
                LOCAL_PATH="$2"
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
            --claude)
                USE_CLAUDE=true
                shift
                ;;
            --no-claude)
                USE_CLAUDE=false
                shift
                ;;
            --compare)
                COMPARE_MODE=true
                shift
                ;;
            -k|--api-key)
                ANTHROPIC_API_KEY="$2"
                shift 2
                ;;
            --parallel)
                PARALLEL=true
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

    # Check compare mode requirements
    if [[ "$COMPARE_MODE" == "true" ]]; then
        if [[ -z "$ANTHROPIC_API_KEY" ]]; then
            echo "Error: --compare mode requires ANTHROPIC_API_KEY" >&2
            echo "Set environment variable or use -k flag" >&2
            exit 1
        fi
        USE_CLAUDE=false  # Start with basic
    fi
}

# Log message if verbose
log() {
    if [ "$VERBOSE" = true ]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*" >&2
    fi
}

# Strip temporary file paths from package names for display
# Removes paths like /var/folders/.../tmp.xxx/ and var/folders/.../tmp.xxx/
clean_package_name() {
    local package_name="$1"

    # Remove absolute temp paths (e.g., /var/folders/.../tmp.xxx/)
    package_name="${package_name#/var/folders/*/tmp.*/}"
    package_name="${package_name#/tmp/tmp.*/}"
    package_name="${package_name#/private/var/folders/*/tmp.*/}"

    # Remove relative temp paths (e.g., var/folders/.../tmp.xxx/)
    package_name="${package_name#var/folders/*/tmp.*/}"
    package_name="${package_name#tmp/tmp.*/}"

    # Remove any remaining leading slashes or path separators
    package_name="${package_name#/}"

    echo "$package_name"
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

    # If LOCAL_PATH is set, use it instead of cloning
    if [[ -n "$LOCAL_PATH" ]]; then
        if [[ ! -d "$LOCAL_PATH" ]]; then
            echo "Error: Local path does not exist: $LOCAL_PATH" >&2
            exit 1
        fi

        log "Using pre-cloned repository at $LOCAL_PATH"

        # Generate SBOM using global SBOM library
        local sbom_file=$(mktemp)
        log "Generating SBOM using global library"

        # Use the generate_sbom function from utils/lib/sbom.sh
        if ! generate_sbom "$LOCAL_PATH" "$sbom_file" "true" >&2; then
            rm -f "$sbom_file"
            echo "Error: Failed to generate SBOM" >&2
            exit 1
        fi

        echo "$sbom_file"
        return 0
    fi

    log "Generating SBOM for $repo"

    # Create temp directory for cloning
    local temp_dir=$(mktemp -d)
    TEMP_DIRS+=("$temp_dir")

    # Convert repo URL to git clone format
    local clone_url="$repo"
    if [[ ! "$repo" =~ ^https?:// ]] && [[ ! "$repo" =~ ^git@ ]]; then
        # Assume it's in owner/repo format, convert to HTTPS
        clone_url="https://github.com/$repo"
    fi

    log "Cloning repository from $clone_url to $temp_dir"
    if ! git clone --depth 1 --quiet "$clone_url" "$temp_dir/repo" 2>/dev/null; then
        echo "Error: Failed to clone repository: $clone_url" >&2
        exit 1
    fi

    # Generate SBOM using global SBOM library
    local sbom_file=$(mktemp)
    log "Generating SBOM using global library"

    # Use the generate_sbom function from utils/lib/sbom.sh
    if ! generate_sbom "$temp_dir/repo" "$sbom_file" "true" >&2; then
        rm -f "$sbom_file"
        echo "Error: Failed to generate SBOM" >&2
        exit 1
    fi

    echo "$sbom_file"
}

# Analyze single package
analyze_package() {
    local system=$1
    local package=$2
    local version=$3

    local display_name=$(clean_package_name "$package")
    log "Analyzing $display_name@$version ($system)"

    # Get health analysis
    local health_result=$(analyze_package_health "$system" "$package" "$version" 2>/dev/null || echo '{"error": "analysis_failed"}')

    # Validate health_result is valid JSON
    if ! jq empty <<< "$health_result" 2>/dev/null; then
        log "Warning: Invalid JSON from health analysis for $display_name, using error placeholder"
        health_result='{"error": "invalid_response", "package": "'$package'", "system": "'$system'", "version": "'$version'"}'
    fi

    # Check deprecation
    local deprecation_result="{}"
    if [ "$CHECK_DEPRECATION" = true ]; then
        deprecation_result=$(comprehensive_deprecation_check "$system" "$package" 2>/dev/null || echo '{"deprecated": false}')

        # Validate deprecation_result is valid JSON
        if ! jq empty <<< "$deprecation_result" 2>/dev/null; then
            log "Warning: Invalid JSON from deprecation check for $display_name"
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

    # Count total packages for progress tracking
    local total_packages=$(echo "$packages" | jq -s 'length')
    local current_package=0

    # Analyze each package
    local package_results="[]"
    local package_usage="{}"  # For version analysis: {package: [{repo, version}, ...]}

    if [ "$PARALLEL" = true ]; then
        # Batch API processing mode
        echo -e "\033[0;32mBatch mode enabled: processing up to $BATCH_SIZE packages per batch\033[0m" >&2

        # Prepare packages for batch request - only include ecosystems supported by deps.dev
        # Supported: npm, pypi, cargo, maven, go, nuget
        local batch_packages=$(echo "$packages" | jq -s 'map({
            system: ((.ecosystem | ascii_downcase) as $eco |
                if $eco == "npm" or $eco == "javascript" or $eco == "node" then "npm"
                elif $eco == "pypi" or $eco == "python" then "pypi"
                elif $eco == "cargo" or $eco == "rust" or $eco == "crates.io" then "cargo"
                elif $eco == "maven" or $eco == "java" then "maven"
                elif $eco == "go" or $eco == "golang" then "go"
                elif $eco == "nuget" or $eco == "dotnet" then "nuget"
                else "unsupported" end
            ),
            name: .package,
            version: .version
        }) | map(select(.system != "unsupported" and .name != null and .version != "unknown"))')

        local valid_packages_count=$(echo "$batch_packages" | jq 'length')

        # If no valid packages for batch API, fall back to sequential
        if [[ "$valid_packages_count" -eq 0 ]]; then
            echo -e "\033[0;33mNo packages with supported ecosystems (npm, pypi, cargo, maven, go, nuget)\033[0m" >&2
            echo -e "\033[0;33mFalling back to sequential processing for unsupported ecosystems\033[0m" >&2
            PARALLEL=false
        else
            echo -e "\033[0;34mFetching version data for $valid_packages_count packages via batch API...\033[0m" >&2

            # Get batch version data
            local batch_response=$(get_versions_batch "$batch_packages")

            if echo "$batch_response" | jq -e '.error' > /dev/null 2>&1; then
                echo -e "\033[0;33mWarning: Batch API failed, falling back to sequential processing\033[0m" >&2
                PARALLEL=false
            else
            # Build a lookup map of version data: {package@version: data}
            local version_lookup=$(echo "$batch_response" | jq -r '
                [.responses[]? | select(.versionKey) | {
                    key: ("\(.versionKey.name)@\(.versionKey.version)"),
                    value: .
                }] | from_entries
            ')

            # Now fetch package summary data (no batch endpoint, use sequential with cache)
            local current_package=0
            local batch_start_time=$(date +%s)
            local batch_processed=0

            while IFS= read -r pkg_info; do
                [ -z "$pkg_info" ] && continue

                local package=$(jq -r '.package' <<< "$pkg_info")
                local version=$(jq -r '.version' <<< "$pkg_info")
                local ecosystem=$(jq -r '.ecosystem' <<< "$pkg_info")
                local system=$(map_ecosystem "$ecosystem")

                # Skip if invalid
                if [ "$system" = "unknown" ] || [ -z "$package" ] || [ "$version" = "unknown" ]; then
                    local display_name=$(clean_package_name "$package")
                    log "Skipping invalid package: $display_name"
                    continue
                fi

                ((current_package++))
                ((batch_processed++))
                local display_name=$(clean_package_name "$package")

                # Calculate ETA
                local eta_str=""
                if [ $batch_processed -gt 1 ]; then
                    local elapsed=$(($(date +%s) - batch_start_time))
                    local avg_time=$((elapsed / batch_processed))
                    local remaining=$((valid_packages_count - current_package))
                    local eta_secs=$((avg_time * remaining))
                    if [ $eta_secs -ge 60 ]; then
                        eta_str=" (ETA: $((eta_secs / 60))m $((eta_secs % 60))s)"
                    elif [ $eta_secs -gt 0 ]; then
                        eta_str=" (ETA: ${eta_secs}s)"
                    fi
                fi

                printf "\r\033[K\033[0;34mProcessing package %d of %d: %s%s\033[0m" \
                    "$current_package" "$valid_packages_count" "${display_name:0:35}" "$eta_str" >&2

                # Get package summary (uses cache)
                local package_summary=$(get_package_summary "$system" "$package")

                # Get version info from batch response
                local lookup_key="${package}@${version}"
                local version_info=$(echo "$version_lookup" | jq -r --arg key "$lookup_key" '.[$key] // {}')

                # If version info not in batch (shouldn't happen), fetch individually
                if [ "$version_info" = "{}" ]; then
                    version_info=$(get_package_version "$system" "$package" "$version")
                fi

                # Calculate health score
                local health_score=$(calculate_health_score "$package_summary" "$version_info" "$version")
                local health_grade=$(get_health_grade "$health_score")

                # Get individual component scores
                local openssf_raw=$(echo "$package_summary" | jq -r '.openssf_score // null')
                local openssf_score=$(calculate_openssf_score "$openssf_raw")
                local maintenance_score=$(calculate_maintenance_score "$package_summary")
                local security_score=$(calculate_security_score "$version_info")
                local latest_version=$(echo "$package_summary" | jq -r '.latest_version')
                local freshness_score=$(calculate_freshness_score "$version" "$latest_version")
                local dependent_count=$(echo "$package_summary" | jq -r '.dependent_count')
                local popularity_score=$(calculate_popularity_score "$dependent_count")

                # Check deprecation if enabled
                local deprecation_result="{\"deprecated\": false}"
                if [ "$CHECK_DEPRECATION" = true ]; then
                    deprecation_result=$(comprehensive_deprecation_check "$system" "$package" 2>/dev/null || echo '{"deprecated": false}')
                fi

                # Build result
                local result=$(jq -n \
                    --arg name "$package" \
                    --arg system "$system" \
                    --arg version "$version" \
                    --argjson score "$health_score" \
                    --arg grade "$health_grade" \
                    --argjson openssf "$openssf_score" \
                    --argjson openssf_raw "$openssf_raw" \
                    --argjson maintenance "$maintenance_score" \
                    --argjson security "$security_score" \
                    --argjson freshness "$freshness_score" \
                    --argjson popularity "$popularity_score" \
                    --arg latest "$latest_version" \
                    --argjson deprecated "$(echo "$package_summary" | jq -r '.deprecated')" \
                    --arg deprecation_msg "$(echo "$package_summary" | jq -r '.deprecation_message // ""')" \
                    --argjson dependent_count "$dependent_count" \
                    --argjson deprecation "$deprecation_result" \
                    '{
                        package: $name,
                        system: $system,
                        version: $version,
                        health_score: $score,
                        health_grade: $grade,
                        component_scores: {
                            openssf: $openssf,
                            openssf_raw: $openssf_raw,
                            maintenance: $maintenance,
                            security: $security,
                            freshness: $freshness,
                            popularity: $popularity
                        },
                        latest_version: $latest,
                        deprecated: $deprecated,
                        deprecation_message: $deprecation_msg,
                        dependent_count: $dependent_count,
                        deprecation: $deprecation
                    }')

                # Add to results
                package_results=$(jq --argjson item "$result" '. + [$item]' <<< "$package_results" 2>/dev/null || echo "$package_results")

                # Track for version analysis
                if [ "$ANALYZE_VERSIONS" = true ]; then
                    local usage_item=$(jq -n --arg repo "$repo_name" --arg ver "$version" '{repo: $repo, version: $ver}')
                    package_usage=$(jq --arg pkg "$package" --argjson item "$usage_item" '
                        .[$pkg] = ((.[$pkg] // []) + [$item])
                    ' <<< "$package_usage")
                fi

            done < <(jq -c '.' <<< "$packages")

            # Print newline after progress indicator
            echo "" >&2
            fi  # closes .error check
        fi  # closes valid_packages_count check
    fi  # closes PARALLEL check

    # Sequential processing mode (if not parallel or fallback)
    if [ "$PARALLEL" = false ]; then
        # Track start time for ETA calculation
        local start_time=$(date +%s)
        local processed_count=0

        # Sequential processing mode (original)
        while IFS= read -r pkg_info; do
            [ -z "$pkg_info" ] && continue

            local package=$(jq -r '.package' <<< "$pkg_info")
            local version=$(jq -r '.version' <<< "$pkg_info")
            local ecosystem=$(jq -r '.ecosystem' <<< "$pkg_info")

            # Map ecosystem to deps.dev system
            local system=$(map_ecosystem "$ecosystem")

            # Skip if invalid
            if [ "$system" = "unknown" ] || [ -z "$package" ] || [ "$version" = "unknown" ]; then
                local display_name=$(clean_package_name "$package")
                log "Skipping invalid package: $display_name"
                continue
            fi

            # Increment counter and display progress with ETA
            ((current_package++))
            ((processed_count++))
            local display_name=$(clean_package_name "$package")

            # Calculate ETA
            local eta_str=""
            if [ $processed_count -gt 1 ]; then
                local elapsed=$(($(date +%s) - start_time))
                local avg_time_per_pkg=$((elapsed / processed_count))
                local remaining_pkgs=$((total_packages - current_package))
                local eta_seconds=$((avg_time_per_pkg * remaining_pkgs))
                if [ $eta_seconds -ge 60 ]; then
                    local eta_mins=$((eta_seconds / 60))
                    local eta_secs=$((eta_seconds % 60))
                    eta_str=" (ETA: ${eta_mins}m ${eta_secs}s)"
                elif [ $eta_seconds -gt 0 ]; then
                    eta_str=" (ETA: ${eta_seconds}s)"
                fi
            fi

            # Use \r to overwrite line for cleaner output
            printf "\r\033[K\033[0;34mAnalyzing package %d of %d: %s@%s%s\033[0m" \
                "$current_package" "$total_packages" "${display_name:0:30}" "${version:0:10}" "$eta_str" >&2

            # Analyze package
            local result=$(analyze_package "$system" "$package" "$version")

            # Validate result before adding
            if jq empty <<< "$result" 2>/dev/null; then
                # Add to results only if valid JSON
                package_results=$(jq --argjson item "$result" '. + [$item]' <<< "$package_results" 2>/dev/null || echo "$package_results")
            else
                log "Warning: Skipping invalid result for $display_name@$version"
            fi

            # Track for version analysis
            if [ "$ANALYZE_VERSIONS" = true ]; then
                local usage_item=$(jq -n --arg repo "$repo_name" --arg ver "$version" '{repo: $repo, version: $ver}')
                package_usage=$(jq --arg pkg "$package" --argjson item "$usage_item" '
                    .[$pkg] = ((.[$pkg] // []) + [$item])
                ' <<< "$package_usage")
            fi

        done < <(jq -c '.' <<< "$packages")

        # Print newline after progress indicator
        echo "" >&2
    fi

    # Analyze version inconsistencies (if multiple repos)
    local version_analysis="[]"
    if [ "$ANALYZE_VERSIONS" = true ]; then
        version_analysis=$(analyze_all_versions "$package_usage" 2>/dev/null || echo '[]')
    fi

    # Generate summary
    local total_packages=$(jq 'length' <<< "$package_results")
    local deprecated_count=$(jq '[.[] | select(.deprecation.deprecated == true)] | length' <<< "$package_results")
    local low_health_count=$(jq '[.[] | select(.health_score < 60)] | length' <<< "$package_results")
    local critical_count=$(jq '[.[] | select(.health_grade == "Critical")] | length' <<< "$package_results")

    # Use temp files to avoid "Argument list too long" error with large datasets
    local temp_packages=$(mktemp)
    local temp_versions=$(mktemp)
    echo "$package_results" > "$temp_packages"
    echo "$version_analysis" > "$temp_versions"

    # Build final result using slurpfile to read from temp files
    jq -n \
        --arg timestamp "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        --arg repo "$repo_name" \
        --argjson total "$total_packages" \
        --argjson deprecated "$deprecated_count" \
        --argjson low_health "$low_health_count" \
        --argjson critical "$critical_count" \
        --slurpfile packages "$temp_packages" \
        --slurpfile versions "$temp_versions" \
        '{
            scan_metadata: {
                timestamp: $timestamp,
                repositories_scanned: 1,
                packages_analyzed: $total,
                analyser_version: "1.0.0",
                analyser_type: "base"
            },
            summary: {
                total_packages: $total,
                deprecated_packages: $deprecated,
                low_health_packages: $low_health,
                critical_health_packages: $critical,
                version_inconsistencies: ($versions[0] | length)
            },
            repositories: [
                {
                    name: $repo,
                    package_count: $total
                }
            ],
            packages: $packages[0],
            version_inconsistencies: $versions[0]
        }'

    # Clean up temp files
    rm -f "$temp_packages" "$temp_versions"
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
                analyser_version: "1.0.0",
                analyser_type: "base"
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

#############################################################################
# Claude AI Analysis
#############################################################################

analyze_with_claude() {
    local data="$1"
    local model="claude-sonnet-4-20250514"

    if [[ -z "$ANTHROPIC_API_KEY" ]]; then
        echo "Error: ANTHROPIC_API_KEY required for --claude mode" >&2
        exit 1
    fi

    echo "Analyzing with Claude AI..." >&2

    # Load RAG knowledge for supply chain security best practices
    local rag_context=""
    local repo_root="$(cd "$SCRIPT_DIR/../../.." && pwd)"
    local rag_dir="$repo_root/rag/supply-chain"

    if [[ -f "$rag_dir/package-health/package-management-best-practices.md" ]]; then
        rag_context+="# Package Management Best Practices\n\n"
        rag_context+=$(head -200 "$rag_dir/package-health/package-management-best-practices.md" | tail -n +1)
        rag_context+="\n\n"
    fi

    if [[ -f "$rag_dir/sbom-generation-best-practices.md" ]]; then
        rag_context+="# SBOM Best Practices\n\n"
        rag_context+=$(head -150 "$rag_dir/sbom-generation-best-practices.md" | tail -n +1)
        rag_context+="\n\n"
    fi

    local prompt="You are a supply chain security expert. Analyze this package health data and produce a PRIORITY-BASED ACTION REPORT.

# Supply Chain Security Knowledge Base
$rag_context

# Analysis Requirements

For each finding, you MUST:
1. **Explain the issue** - What is the package health concern and why is it a problem?
2. **Justify the recommendation** - Why this action is needed based on package management best practices
3. **Reference knowledge base** - Cite relevant best practices from the knowledge base above
4. **Provide specific actions** - Exact commands, version numbers, or dependency changes

# Output Format

## 🔴 CRITICAL PRIORITY (Immediate - 0-24 hours)
- Deprecated packages blocking security, critical health issues (score < 30)
- For each item:
  * **Issue**: [Package name, version, and problem]
  * **Risk**: [Why this is critical for supply chain]
  * **Best Practice Reference**: [Cite from knowledge base]
  * **Action**: [Specific migration/update commands]
  * **Timeline**: Immediate

## 🟠 HIGH PRIORITY (Urgent - 1-7 days)
- Low health packages (score 30-60), version inconsistencies, major version drift
- Same structured format as above

## 🟡 MEDIUM PRIORITY (Important - 1-30 days)
- Moderate health issues (score 60-75), minor updates, deduplication opportunities
- Same structured format as above

## 🟢 LOW PRIORITY (Monitor - 30+ days)
- Optimization opportunities, documentation, healthy packages needing minor updates
- Same structured format as above

## 📊 Summary & Strategic Recommendations
- Overall dependency health posture
- Version standardization opportunities (npm dedupe, overrides)
- Lock file management and automation
- Dependency update policies (Dependabot, Renovate)
- Testing strategy for updates

# Scan Data:
$data"

    local response=$(curl -s https://api.anthropic.com/v1/messages \
        -H "content-type: application/json" \
        -H "x-api-key: $ANTHROPIC_API_KEY" \
        -H "anthropic-version: 2023-06-01" \
        -d "{
            \"model\": \"$model\",
            \"max_tokens\": 4096,
            \"messages\": [{
                \"role\": \"user\",
                \"content\": $(echo "$prompt" | jq -Rs .)
            }]
        }")

    if command -v record_api_usage &> /dev/null; then
        record_api_usage "$response" "$model" > /dev/null
    fi

    # Check for API errors
    if echo "$response" | jq -e '.error' > /dev/null 2>&1; then
        local error_type=$(echo "$response" | jq -r '.error.type')
        local error_message=$(echo "$response" | jq -r '.error.message')
        echo "Error: Claude API request failed - $error_type: $error_message" >&2
        return 1
    fi

    # Extract and return the analysis
    local analysis=$(echo "$response" | jq -r '.content[0].text // empty')

    if [[ -z "$analysis" ]]; then
        echo "Error: No analysis returned from Claude API" >&2
        echo "Response: $response" >&2
        return 1
    fi

    echo "$analysis"
}

# Main execution
main() {
    # Load cost tracking if using Claude or compare mode
    if [[ "$USE_CLAUDE" == "true" ]] || [[ "$COMPARE_MODE" == "true" ]]; then
        REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
        if [ -f "$REPO_ROOT/lib/claude-cost.sh" ]; then
            source "$REPO_ROOT/lib/claude-cost.sh"
            init_cost_tracking
        fi
    fi

    parse_args "$@"

    echo ""
    echo "========================================="
    echo "  Package Health Analysis"
    echo "========================================="
    echo ""

    # Check Claude AI status first
    if [[ "$USE_CLAUDE" == "true" ]] && [[ -n "$ANTHROPIC_API_KEY" ]] && [[ "$COMPARE_MODE" != "true" ]]; then
        echo -e "\033[0;32m🤖 Claude AI: ENABLED (analyzing results with AI)\033[0m"
        echo ""
    elif [[ -z "$ANTHROPIC_API_KEY" ]] && [[ "$COMPARE_MODE" != "true" ]]; then
        echo -e "\033[1;33mℹ️  Claude AI: DISABLED (no API key found)\033[0m"
        echo -e "\033[0;36m   Set ANTHROPIC_API_KEY to enable AI-enhanced analysis\033[0m"
        echo -e "\033[0;36m   Get your key at: https://console.anthropic.com/settings/keys\033[0m"
        echo ""
        USE_CLAUDE=false
    fi

    # Inform about batch mode for faster processing
    if [[ "$PARALLEL" != "true" ]]; then
        echo -e "\033[0;36m⚡ Batch API mode available with --parallel flag\033[0m"
        echo -e "\033[0;36m   Processes packages 6-10x faster using deps.dev batch API\033[0m"
        echo -e "\033[0;36m   Recommended for repositories with many dependencies\033[0m"
        echo ""
    fi

    # Validate input
    if [ -z "$REPO" ] && [ -z "$ORG" ] && [ -z "$SBOM_FILE" ]; then
        echo "Error: Must specify --repo, --org, or --sbom" >&2
        usage
    fi

    local result=""
    local temp_sbom=""

    # Extract org name from URL if provided
    if [ -n "$ORG" ]; then
        if [[ "$ORG" =~ github\.com/orgs/([^/]+) ]]; then
            ORG="${BASH_REMATCH[1]}"
        elif [[ "$ORG" =~ github\.com/([^/]+) ]]; then
            ORG="${BASH_REMATCH[1]}"
        fi
        ORG="${ORG%/}"  # Remove trailing slashes
    fi

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
    if [[ "$COMPARE_MODE" == "true" ]]; then
        # Comparison mode: run both basic and Claude
        echo "========================================="
        echo "  Package Health Analysis (Comparison)"
        echo "========================================="
        echo ""
        echo "Running basic analysis..."
        echo ""

        USE_CLAUDE=false
        local formatted=$(format_output "$result" "$OUTPUT_FORMAT")
        echo "$formatted"

        echo ""
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo ""
        echo "Running Claude AI analysis..."
        echo ""

        USE_CLAUDE=true
        local claude_analysis=$(analyze_with_claude "$formatted")
        echo "$claude_analysis"

        # Display comparison summary
        echo ""
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "Comparison Summary"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo ""
        echo "Basic analyser provides:"
        echo "  • Package health scores and metrics"
        echo "  • Vulnerability counts and severity"
        echo "  • Deprecation status"
        echo "  • Version analysis"
        echo "  • Raw data export (JSON/CSV)"
        echo ""
        echo "Claude-enhanced analyser adds:"
        echo "  • Risk prioritization and assessment"
        echo "  • Contextual security insights"
        echo "  • Specific remediation recommendations"
        echo "  • Upgrade path guidance"
        echo "  • Impact analysis"
        echo "  • Prioritized action plan"
        echo ""

        if command -v display_api_cost_summary &> /dev/null; then
            echo "API Cost:"
            display_api_cost_summary
        fi

        echo ""
        echo "Use basic for: Automation, CI/CD, dashboards"
        echo "Use Claude for: Security reviews, upgrade planning, risk assessment"
        echo ""

    elif [[ "$USE_CLAUDE" == "true" ]]; then
        # Claude AI analysis mode
        local formatted=$(format_output "$result" "$OUTPUT_FORMAT")
        local claude_analysis=$(analyze_with_claude "$formatted")

        echo "========================================="
        echo "  Package Health Analysis (Claude AI)"
        echo "========================================="
        echo ""
        echo "$claude_analysis"

        # Display cost summary
        if command -v display_api_cost_summary &> /dev/null; then
            display_api_cost_summary
        fi
    else
        # Standard analysis mode
        local formatted=$(format_output "$result" "$OUTPUT_FORMAT")

        if [ -n "$OUTPUT_FILE" ]; then
            echo "$formatted" > "$OUTPUT_FILE"
            log "Report written to $OUTPUT_FILE"
        else
            echo "$formatted"
        fi
    fi
}

# Run main function
main "$@"
