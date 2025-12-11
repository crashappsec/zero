#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Vulnerability Analyser - Data Collector
# Pure data collection using osv-scanner - outputs JSON for agent analysis
#
# This is the data-only version. AI analysis is handled by agents.
#
# Usage: ./vulnerability-analyser-data.sh [options] <target>
# Output: JSON with vulnerabilities, severity, KEV status, and metadata
#############################################################################

set -e

# Colors for terminal output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PARENT_DIR="$(dirname "$SCRIPT_DIR")"

# Default options
TAINT_ANALYSIS=false
OUTPUT_FORMAT="json"  # Always JSON for data collection
OUTPUT_FILE=""
LOCAL_PATH=""
CLEANUP=true
PRIORITIZE=true  # Always prioritize for structured output
KEV_CACHE=""
TEMP_DIR=""
TARGET=""

usage() {
    cat << EOF
Vulnerability Analyser - Data Collector (JSON output for agent analysis)

Usage: $0 [OPTIONS] <target>

TARGET:
    Git repository URL      Clone and analyze repository
    Local directory path    Analyze local repository
    SBOM file path          Analyze existing SBOM

OPTIONS:
    --local-path PATH       Use pre-cloned repository (skips cloning)
    -t, --taint-analysis    Enable call graph analysis (Go projects)
    -o, --output FILE       Write JSON to file (default: stdout)
    -k, --keep-clone        Keep cloned repository
    -h, --help              Show this help

OUTPUT:
    JSON object with:
    - metadata: scan timestamp, target, analyzer version
    - summary: counts by severity
    - vulnerabilities: array of findings with priority scores

EXAMPLES:
    $0 https://github.com/expressjs/express
    $0 --local-path ~/.gibson/projects/foo/repo
    $0 -o vulns.json /path/to/project

EOF
    exit 0
}

# Check dependencies
check_osv_scanner() {
    if ! command -v osv-scanner &> /dev/null; then
        echo '{"error": "osv-scanner not installed", "install": "go install github.com/google/osv-scanner/cmd/osv-scanner@latest"}' >&2
        exit 1
    fi
}

# Fetch CISA KEV catalog
fetch_kev_catalog() {
    KEV_CACHE=$(mktemp)
    curl -sf "https://www.cisa.gov/sites/default/files/feeds/known_exploited_vulnerabilities.json" \
        -o "$KEV_CACHE" 2>/dev/null || echo '{"vulnerabilities":[]}' > "$KEV_CACHE"
    echo -e "${BLUE}Fetched CISA KEV catalog${NC}" >&2
}

# Check if CVE is in KEV
is_in_kev() {
    local vuln_id="$1"
    [[ -f "$KEV_CACHE" ]] && grep -q "\"$vuln_id\"" "$KEV_CACHE" 2>/dev/null
}

# Fetch full vulnerability details from OSV.dev API
# Returns JSON with full details including description, references, affected versions, fix info
fetch_osv_details() {
    local vuln_id="$1"
    local osv_response

    # OSV.dev API endpoint
    osv_response=$(curl -sf "https://api.osv.dev/v1/vulns/${vuln_id}" 2>/dev/null)

    if [[ -n "$osv_response" ]]; then
        echo "$osv_response"
    else
        echo "{}"
    fi
}

# Detect target type
is_git_url() {
    [[ "$1" =~ ^(https?|git)://.*\.git$ ]] || [[ "$1" =~ ^git@.*:.*\.git$ ]] || [[ "$1" =~ github\.com|gitlab\.com|bitbucket\.org ]]
}

is_sbom_file() {
    local file="$1"
    [[ -f "$file" ]] && ([[ "$file" =~ \.json$ ]] || [[ "$file" =~ \.xml$ ]] || [[ "$file" =~ \.cdx\. ]] || [[ "$file" =~ bom\. ]])
}

extract_repo_name() {
    local target="$1"
    if [[ "$target" =~ github\.com[/:]([^/]+/[^/.]+) ]]; then
        echo "${BASH_REMATCH[1]}"
    elif [[ "$target" =~ ^([a-zA-Z0-9_-]+/[a-zA-Z0-9_.-]+)$ ]]; then
        echo "$target"
    elif [[ -d "$target/.git" ]]; then
        local remote=$(git -C "$target" remote get-url origin 2>/dev/null)
        if [[ "$remote" =~ github\.com[/:]([^/]+/[^/.]+) ]]; then
            echo "${BASH_REMATCH[1]}"
        else
            basename "$target"
        fi
    else
        basename "$target"
    fi
}

# Clone repository
clone_repository() {
    local repo_url="$1"
    TEMP_DIR=$(mktemp -d)

    echo -e "${BLUE}Cloning repository...${NC}" >&2
    if git clone --depth 1 "$repo_url" "$TEMP_DIR" 2>/dev/null; then
        echo -e "${GREEN}✓ Cloned${NC}" >&2
        return 0
    else
        echo '{"error": "Failed to clone repository"}'
        exit 1
    fi
}

# Find lockfiles in directory
find_lockfiles() {
    local target_path="$1"
    local lockfiles=()

    # Common lockfile patterns
    local patterns=(
        "package-lock.json"
        "yarn.lock"
        "pnpm-lock.yaml"
        "Gemfile.lock"
        "Cargo.lock"
        "poetry.lock"
        "Pipfile.lock"
        "composer.lock"
        "go.sum"
        "requirements.txt"
        "gradle.lockfile"
        "pom.xml"
        "packages.lock.json"
    )

    for pattern in "${patterns[@]}"; do
        while IFS= read -r -d '' file; do
            lockfiles+=("$file")
        done < <(find "$target_path" -name "$pattern" -type f -print0 2>/dev/null)
    done

    printf '%s\n' "${lockfiles[@]}"
}

# Run osv-scanner and get JSON
run_osv_scan() {
    local target_path="$1"
    local temp_output=$(mktemp)

    # Find all lockfiles
    local lockfiles=()
    while IFS= read -r file; do
        [[ -n "$file" ]] && lockfiles+=("$file")
    done < <(find_lockfiles "$target_path")

    if [[ ${#lockfiles[@]} -eq 0 ]]; then
        echo '{"results":[]}'
        rm -f "$temp_output"
        return
    fi

    # Build scan command with lockfile flags
    # osv-scanner v2+ uses: osv-scanner scan source -L file1 -L file2 ...
    local scan_cmd="osv-scanner scan source --format=json"

    for lockfile in "${lockfiles[@]}"; do
        scan_cmd="$scan_cmd -L \"$lockfile\""
    done

    if [[ "$TAINT_ANALYSIS" == true ]]; then
        scan_cmd="$scan_cmd --call-analysis=all"
    fi

    # Run scan (exit 1 when vulns found is normal)
    eval "$scan_cmd" > "$temp_output" 2>/dev/null || true

    # Extract JSON (skip any non-JSON prefix)
    if grep -q "^{" "$temp_output"; then
        grep -A 999999 "^{" "$temp_output"
    else
        echo '{"results":[]}'
    fi

    rm -f "$temp_output"
}

# Process scan results into structured JSON
process_results() {
    local raw_json="$1"
    local target="$2"
    local repo_name="$3"

    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    local vulns_array="[]"
    local total=0 critical=0 high=0 medium=0 low=0 kev_count=0

    # Process vulnerabilities
    while IFS= read -r vuln_json; do
        [[ -z "$vuln_json" ]] && continue

        local vuln_id=$(echo "$vuln_json" | jq -r '.id')
        local package=$(echo "$vuln_json" | jq -r '.package')
        local version=$(echo "$vuln_json" | jq -r '.version')
        local ecosystem=$(echo "$vuln_json" | jq -r '.ecosystem')
        local cvss=$(echo "$vuln_json" | jq -r '.cvss' | grep -oE '[0-9]+(\.[0-9]+)?' | head -1 || echo "0")
        local summary=$(echo "$vuln_json" | jq -r '.summary')

        # Calculate priority
        local priority_score=0
        local in_kev=false
        local severity="low"

        # CISA KEV check
        if is_in_kev "$vuln_id"; then
            priority_score=$((priority_score + 100))
            in_kev=true
            ((kev_count++))
        fi

        # CVSS scoring - use awk for portability instead of bc
        if [[ -n "$cvss" ]] && [[ "$cvss" != "0" ]]; then
            local cvss_level=$(awk -v score="$cvss" 'BEGIN {
                if (score >= 9.0) print "critical"
                else if (score >= 7.0) print "high"
                else if (score >= 4.0) print "medium"
                else print "low"
            }')
            case "$cvss_level" in
                critical)
                    priority_score=$((priority_score + 50))
                    severity="critical"
                    ;;
                high)
                    priority_score=$((priority_score + 30))
                    severity="high"
                    ;;
                medium)
                    priority_score=$((priority_score + 15))
                    severity="medium"
                    ;;
                *)
                    priority_score=$((priority_score + 5))
                    ;;
            esac
        fi

        # Override severity if in KEV
        if [[ "$in_kev" == true ]]; then
            severity="critical"
        fi

        # Count by severity
        ((total++))
        case "$severity" in
            critical) ((critical++)) ;;
            high) ((high++)) ;;
            medium) ((medium++)) ;;
            low) ((low++)) ;;
        esac

        # Fetch full details from OSV.dev
        echo -e "${DIM}  Fetching details for $vuln_id...${NC}" >&2
        local osv_details=$(fetch_osv_details "$vuln_id")

        # Extract additional fields from OSV response
        local description=$(echo "$osv_details" | jq -r '.details // .summary // ""' 2>/dev/null)
        local aliases=$(echo "$osv_details" | jq -c '[.aliases[]?] // []' 2>/dev/null)
        local references=$(echo "$osv_details" | jq -c '[.references[]? | {type: .type, url: .url}] // []' 2>/dev/null)
        local affected_ranges=$(echo "$osv_details" | jq -c '[.affected[]? | {package: .package.name, ecosystem: .package.ecosystem, ranges: .ranges, versions: .versions}] // []' 2>/dev/null)
        local fix_available=$(echo "$osv_details" | jq -r 'if .affected[0].ranges[0].events | map(select(.fixed)) | length > 0 then "yes" else "no" end' 2>/dev/null || echo "unknown")
        local fixed_version=$(echo "$osv_details" | jq -r '.affected[0].ranges[0].events[] | select(.fixed) | .fixed' 2>/dev/null | head -1)
        local published=$(echo "$osv_details" | jq -r '.published // ""' 2>/dev/null)
        local modified=$(echo "$osv_details" | jq -r '.modified // ""' 2>/dev/null)
        local severity_osv=$(echo "$osv_details" | jq -c '.severity // []' 2>/dev/null)

        # Use OSV description if available and longer than summary
        if [[ -n "$description" ]] && [[ ${#description} -gt ${#summary} ]]; then
            summary="$description"
        fi

        # Build vulnerability object with full details
        local vuln_obj=$(jq -n \
            --arg id "$vuln_id" \
            --arg pkg "$package" \
            --arg ver "$version" \
            --arg eco "$ecosystem" \
            --arg cvs "$cvss" \
            --arg sev "$severity" \
            --argjson score "$priority_score" \
            --argjson kev "$in_kev" \
            --arg sum "$summary" \
            --argjson aliases "$aliases" \
            --argjson references "$references" \
            --argjson affected "$affected_ranges" \
            --arg fix_available "$fix_available" \
            --arg fixed_version "$fixed_version" \
            --arg published "$published" \
            --arg modified "$modified" \
            --argjson severity_scores "$severity_osv" \
            '{
                id: $id,
                package: $pkg,
                version: $ver,
                ecosystem: $eco,
                cvss: $cvs,
                severity: $sev,
                priority_score: $score,
                in_cisa_kev: $kev,
                summary: $sum,
                aliases: $aliases,
                references: $references,
                affected: $affected,
                fix_available: $fix_available,
                fixed_version: (if $fixed_version == "" then null else $fixed_version end),
                published: (if $published == "" then null else $published end),
                modified: (if $modified == "" then null else $modified end),
                severity_scores: $severity_scores,
                osv_url: ("https://osv.dev/vulnerability/" + $id)
            }')

        vulns_array=$(echo "$vulns_array" | jq --argjson v "$vuln_obj" '. + [$v]')

    done < <(echo "$raw_json" | jq -r '.results[]? | .packages[]? | select(.vulnerabilities) |
        .package as $pkg | .groups as $groups | .vulnerabilities[] |
        . as $vuln |
        # Find max_severity from groups that contain this vuln id
        ($groups | map(select(.ids[]? == $vuln.id)) | .[0].max_severity // "0") as $cvss |
        {
            id: (.id // "N/A"),
            package: ($pkg.name // "N/A"),
            version: ($pkg.version // "N/A"),
            ecosystem: ($pkg.ecosystem // "N/A"),
            cvss: ($cvss | tostring),
            summary: (.summary // .details // "No description")
        } | @json' 2>/dev/null)

    # Sort by priority score descending
    vulns_array=$(echo "$vulns_array" | jq 'sort_by(-.priority_score)')

    # Build final output
    jq -n \
        --arg ts "$timestamp" \
        --arg tgt "$target" \
        --arg repo "$repo_name" \
        --arg ver "1.0.0" \
        --argjson tot "$total" \
        --argjson crit "$critical" \
        --argjson hi "$high" \
        --argjson med "$medium" \
        --argjson lo "$low" \
        --argjson kev "$kev_count" \
        --argjson vulns "$vulns_array" \
        '{
            analyzer: "vulnerability-analyser",
            version: $ver,
            timestamp: $ts,
            target: $tgt,
            repository: $repo,
            summary: {
                total: $tot,
                critical: $crit,
                high: $hi,
                medium: $med,
                low: $lo,
                cisa_kev: $kev
            },
            vulnerabilities: $vulns
        }'
}

# Cleanup
cleanup() {
    if [[ "$CLEANUP" == true ]] && [[ -n "$TEMP_DIR" ]] && [[ -d "$TEMP_DIR" ]]; then
        rm -rf "$TEMP_DIR"
    fi
    [[ -n "$KEV_CACHE" ]] && [[ -f "$KEV_CACHE" ]] && rm -f "$KEV_CACHE"
}
trap cleanup EXIT

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help) usage ;;
        --local-path)
            LOCAL_PATH="$2"
            shift 2
            ;;
        -t|--taint-analysis)
            TAINT_ANALYSIS=true
            shift
            ;;
        -o|--output)
            OUTPUT_FILE="$2"
            shift 2
            ;;
        -k|--keep-clone)
            CLEANUP=false
            shift
            ;;
        -*)
            echo "Unknown option: $1" >&2
            exit 1
            ;;
        *)
            TARGET="$1"
            shift
            ;;
    esac
done

# Main execution
check_osv_scanner
fetch_kev_catalog

# Determine scan path
scan_path=""
repo_name=""

if [[ -n "$LOCAL_PATH" ]]; then
    if [[ ! -d "$LOCAL_PATH" ]]; then
        echo '{"error": "Local path does not exist"}'
        exit 1
    fi
    scan_path="$LOCAL_PATH"
    repo_name=$(extract_repo_name "$LOCAL_PATH")
elif [[ -n "$TARGET" ]]; then
    if is_git_url "$TARGET"; then
        repo_name=$(extract_repo_name "$TARGET")
        clone_repository "$TARGET"
        scan_path="$TEMP_DIR"
    elif [[ -d "$TARGET" ]]; then
        scan_path="$TARGET"
        repo_name=$(extract_repo_name "$TARGET")
    elif is_sbom_file "$TARGET"; then
        # For SBOM, run osv-scanner directly with -L flag
        scan_path="$TARGET"
        repo_name="sbom:$(basename "$TARGET")"
    else
        echo '{"error": "Invalid target - must be URL, directory, or SBOM file"}'
        exit 1
    fi
else
    echo '{"error": "No target specified"}'
    exit 1
fi

echo -e "${BLUE}Scanning: $repo_name${NC}" >&2

# Run scan and process
raw_results=$(run_osv_scan "$scan_path")
final_json=$(process_results "$raw_results" "$TARGET" "$repo_name")

# Output
if [[ -n "$OUTPUT_FILE" ]]; then
    echo "$final_json" > "$OUTPUT_FILE"
    echo -e "${GREEN}✓ Results written to $OUTPUT_FILE${NC}" >&2
else
    echo "$final_json"
fi
