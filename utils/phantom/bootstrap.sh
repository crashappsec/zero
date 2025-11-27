#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Phantom Bootstrap
# Clone a repository and run all analyzers to prepare for agent queries
#
# Usage: ./bootstrap.sh <target> [options]
#
# Examples:
#   ./bootstrap.sh https://github.com/expressjs/express
#   ./bootstrap.sh expressjs/express
#   ./bootstrap.sh expressjs/express --branch v5.x
#   ./bootstrap.sh expressjs/express --quick
#############################################################################

set -e

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Load Gibson library
source "$SCRIPT_DIR/lib/gibson.sh"

# Load shared config if available
UTILS_ROOT="$(dirname "$SCRIPT_DIR")"
REPO_ROOT="$(dirname "$UTILS_ROOT")"

if [[ -f "$UTILS_ROOT/lib/config.sh" ]]; then
    source "$UTILS_ROOT/lib/config.sh"
fi

# Load .env if it exists
if [[ -f "$REPO_ROOT/.env" ]]; then
    set -a
    source "$REPO_ROOT/.env"
    set +a
fi

#############################################################################
# Configuration
#############################################################################

TARGET=""
BRANCH=""
DEPTH=""
MODE="full"  # full, quick, security-only
FORCE=false

#############################################################################
# Usage
#############################################################################

usage() {
    cat << EOF
Phantom Hydrate - Clone and analyze a repository for agent queries

Usage: $0 <target> [options]

TARGETS:
    GitHub URL:      https://github.com/owner/repo
    GitHub shorthand: owner/repo
    Local path:       /path/to/project or ./project

OPTIONS:
    --branch <name>     Clone specific branch (default: default branch)
    --depth <n>         Shallow clone depth (default: full for DORA metrics)
    --quick             Fast analyzers only (skip code-security, dora)
    --security-only     Security analyzers only
    --force             Re-hydrate even if project exists
    -h, --help          Show this help

EXAMPLES:
    $0 expressjs/express
    $0 https://github.com/lodash/lodash --branch main
    $0 owner/repo --quick
    $0 ./local-project

FLOW:
    1. Run preflight check (verifies tools/keys)
    2. Clone repository to ~/.phantom/projects/<id>/repo/
    3. Run analyzers and store JSON in ~/.phantom/projects/<id>/analysis/
    4. Set as active project for agent queries

EOF
    exit 0
}

#############################################################################
# Argument Parsing
#############################################################################

parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            -h|--help)
                usage
                ;;
            --branch)
                BRANCH="$2"
                shift 2
                ;;
            --depth)
                DEPTH="$2"
                shift 2
                ;;
            --quick)
                MODE="quick"
                shift
                ;;
            --security-only)
                MODE="security"
                shift
                ;;
            --force)
                FORCE=true
                shift
                ;;
            -*)
                echo -e "${RED}Error: Unknown option $1${NC}" >&2
                exit 1
                ;;
            *)
                if [[ -z "$TARGET" ]]; then
                    TARGET="$1"
                else
                    echo -e "${RED}Error: Multiple targets specified${NC}" >&2
                    exit 1
                fi
                shift
                ;;
        esac
    done

    if [[ -z "$TARGET" ]]; then
        echo -e "${RED}Error: No target specified${NC}" >&2
        echo "Usage: $0 <target> [options]"
        exit 1
    fi
}

#############################################################################
# Clone Functions
#############################################################################

clone_github_repo() {
    local url="$1"
    local dest="$2"
    local branch="$3"
    local depth="$4"

    local clone_args=()

    if [[ -n "$branch" ]]; then
        clone_args+=("--branch" "$branch")
    fi

    if [[ -n "$depth" ]]; then
        clone_args+=("--depth" "$depth")
    fi

    # Use GITHUB_TOKEN if available for private repos
    if [[ -n "${GITHUB_TOKEN:-}" ]]; then
        # Convert HTTPS URL to use token
        if echo "$url" | grep -q '^https://github\.com/'; then
            url=$(echo "$url" | sed "s|https://github.com|https://${GITHUB_TOKEN}@github.com|")
        fi
    fi

    git clone "${clone_args[@]}" "$url" "$dest" 2>&1
}

copy_local_project() {
    local source="$1"
    local dest="$2"

    # Resolve absolute path
    local abs_source=$(cd "$source" && pwd)

    # Copy entire directory
    cp -R "$abs_source" "$dest"

    # If it's a git repo, ensure we have the .git directory
    if [[ -d "$abs_source/.git" ]]; then
        echo "  Git repository detected"
    else
        echo "  Not a git repository (DORA metrics unavailable)"
    fi
}

get_git_info() {
    local repo_path="$1"

    if [[ ! -d "$repo_path/.git" ]]; then
        echo ""
        echo ""
        return
    fi

    local branch=$(git -C "$repo_path" rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
    local commit=$(git -C "$repo_path" rev-parse --short HEAD 2>/dev/null || echo "unknown")

    echo "$branch"
    echo "$commit"
}

#############################################################################
# Project Detection
#############################################################################

detect_project_type() {
    local repo_path="$1"

    local languages=()
    local frameworks=()
    local package_managers=()

    # Detect by files present
    if [[ -f "$repo_path/package.json" ]]; then
        package_managers+=("npm")
        languages+=("javascript")

        # Check for specific frameworks in package.json
        if grep -q '"react"' "$repo_path/package.json" 2>/dev/null; then
            frameworks+=("react")
        fi
        if grep -q '"express"' "$repo_path/package.json" 2>/dev/null; then
            frameworks+=("express")
        fi
        if grep -q '"next"' "$repo_path/package.json" 2>/dev/null; then
            frameworks+=("nextjs")
        fi
        if grep -q '"vue"' "$repo_path/package.json" 2>/dev/null; then
            frameworks+=("vue")
        fi
        if grep -q '"typescript"' "$repo_path/package.json" 2>/dev/null; then
            languages+=("typescript")
        fi
    fi

    if [[ -f "$repo_path/requirements.txt" ]] || [[ -f "$repo_path/setup.py" ]] || [[ -f "$repo_path/pyproject.toml" ]]; then
        package_managers+=("pip")
        languages+=("python")

        if grep -q 'django' "$repo_path/requirements.txt" 2>/dev/null; then
            frameworks+=("django")
        fi
        if grep -q 'flask' "$repo_path/requirements.txt" 2>/dev/null; then
            frameworks+=("flask")
        fi
        if grep -q 'fastapi' "$repo_path/requirements.txt" 2>/dev/null; then
            frameworks+=("fastapi")
        fi
    fi

    if [[ -f "$repo_path/go.mod" ]]; then
        package_managers+=("go")
        languages+=("go")
    fi

    if [[ -f "$repo_path/Cargo.toml" ]]; then
        package_managers+=("cargo")
        languages+=("rust")
    fi

    if [[ -f "$repo_path/pom.xml" ]] || [[ -f "$repo_path/build.gradle" ]]; then
        if [[ -f "$repo_path/pom.xml" ]]; then
            package_managers+=("maven")
        fi
        if [[ -f "$repo_path/build.gradle" ]]; then
            package_managers+=("gradle")
        fi
        languages+=("java")

        if grep -q 'spring' "$repo_path/pom.xml" 2>/dev/null || grep -q 'spring' "$repo_path/build.gradle" 2>/dev/null; then
            frameworks+=("spring")
        fi
    fi

    if [[ -f "$repo_path/Gemfile" ]]; then
        package_managers+=("bundler")
        languages+=("ruby")

        if grep -q 'rails' "$repo_path/Gemfile" 2>/dev/null; then
            frameworks+=("rails")
        fi
    fi

    if [[ -f "$repo_path/composer.json" ]]; then
        package_managers+=("composer")
        languages+=("php")

        if grep -q 'laravel' "$repo_path/composer.json" 2>/dev/null; then
            frameworks+=("laravel")
        fi
    fi

    # Convert to JSON arrays (handle empty arrays, compact output)
    local langs_json="[]"
    local fwks_json="[]"
    local pkgs_json="[]"

    if [[ ${#languages[@]} -gt 0 ]]; then
        langs_json=$(printf '%s\n' "${languages[@]}" | jq -R . | jq -sc 'unique')
    fi
    if [[ ${#frameworks[@]} -gt 0 ]]; then
        fwks_json=$(printf '%s\n' "${frameworks[@]}" | jq -R . | jq -sc 'unique')
    fi
    if [[ ${#package_managers[@]} -gt 0 ]]; then
        pkgs_json=$(printf '%s\n' "${package_managers[@]}" | jq -R . | jq -sc 'unique')
    fi

    echo "$langs_json"
    echo "$fwks_json"
    echo "$pkgs_json"
}

#############################################################################
# Analyzer Execution
#############################################################################

# Get list of analyzers to run based on mode
get_analyzers_for_mode() {
    local mode="$1"

    case "$mode" in
        quick)
            echo "technology dependencies vulnerabilities licenses"
            ;;
        security)
            echo "vulnerabilities package-health security-findings provenance"
            ;;
        full|*)
            echo "technology dependencies vulnerabilities package-health licenses security-findings ownership dora"
            ;;
    esac
}

# Run a single analyzer
run_analyzer() {
    local analyzer="$1"
    local repo_path="$2"
    local output_path="$3"
    local project_id="$4"

    local start_time=$(date +%s)
    local status="complete"
    local summary="null"

    # Record start
    local analyzer_script=""
    case "$analyzer" in
        technology)
            analyzer_script="technology-identification-analyser.sh"
            ;;
        dependencies)
            analyzer_script="(internal)"
            ;;
        vulnerabilities)
            analyzer_script="vulnerability-analyser.sh"
            ;;
        package-health)
            analyzer_script="package-health-analyser.sh"
            ;;
        licenses)
            analyzer_script="legal-analyser.sh"
            ;;
        security-findings)
            analyzer_script="code-security-analyser.sh"
            ;;
        ownership)
            analyzer_script="ownership-analyser.sh"
            ;;
        dora)
            analyzer_script="dora-analyser.sh"
            ;;
        provenance)
            analyzer_script="provenance-analyser.sh"
            ;;
    esac

    gibson_analysis_start "$project_id" "$analyzer" "$analyzer_script"

    # Run the analyzer
    case "$analyzer" in
        technology)
            run_technology_analyzer "$repo_path" "$output_path"
            ;;
        dependencies)
            run_dependency_extractor "$repo_path" "$output_path"
            ;;
        vulnerabilities)
            run_vulnerability_analyzer "$repo_path" "$output_path"
            ;;
        package-health)
            run_package_health_analyzer "$repo_path" "$output_path"
            ;;
        licenses)
            run_license_analyzer "$repo_path" "$output_path"
            ;;
        security-findings)
            run_code_security_analyzer "$repo_path" "$output_path"
            ;;
        ownership)
            run_ownership_analyzer "$repo_path" "$output_path"
            ;;
        dora)
            run_dora_analyzer "$repo_path" "$output_path"
            ;;
        provenance)
            run_provenance_analyzer "$repo_path" "$output_path"
            ;;
        *)
            status="failed"
            ;;
    esac

    local exit_code=$?
    local end_time=$(date +%s)
    local duration_sec=$((end_time - start_time))
    local duration=$((duration_sec * 1000))  # Convert to ms for manifest

    if [[ $exit_code -ne 0 ]]; then
        status="failed"
    fi

    # Extract summary from output if available
    local output_file="$output_path/${analyzer}.json"
    if [[ -f "$output_file" ]] && [[ "$status" == "complete" ]]; then
        summary=$(jq '.summary // null' "$output_file" 2>/dev/null || echo "null")
    fi

    gibson_analysis_complete "$project_id" "$analyzer" "$status" "$duration" "$summary"

    return $exit_code
}

#############################################################################
# Individual Analyzer Wrappers
#############################################################################

run_technology_analyzer() {
    local repo_path="$1"
    local output_path="$2"

    # Use new data-only analyzer
    local tech_script="$UTILS_ROOT/technology-identification/technology-identification-data.sh"

    if [[ -x "$tech_script" ]]; then
        "$tech_script" --local-path "$repo_path" -o "$output_path/technology.json" 2>/dev/null
    else
        # Fallback: create basic technology.json from detection
        local detection=$(detect_project_type "$repo_path")
        local languages=$(echo "$detection" | sed -n '1p')
        local frameworks=$(echo "$detection" | sed -n '2p')
        local package_managers=$(echo "$detection" | sed -n '3p')

        cat > "$output_path/technology.json" << EOF
{
  "analyzer": "technology-identification",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "languages": $languages,
  "frameworks": $frameworks,
  "package_managers": $package_managers,
  "summary": {
    "language_count": $(echo "$languages" | jq 'length'),
    "framework_count": $(echo "$frameworks" | jq 'length')
  }
}
EOF
    fi
}

run_dependency_extractor() {
    local repo_path="$1"
    local output_path="$2"

    local direct_count=0
    local total_count=0

    # Extract from package.json
    if [[ -f "$repo_path/package.json" ]]; then
        local pkg_deps=$(jq -r '.dependencies // {} | keys | length' "$repo_path/package.json" 2>/dev/null)
        [[ -n "$pkg_deps" ]] && direct_count=$pkg_deps

        # Count from lock file if present
        if [[ -f "$repo_path/package-lock.json" ]]; then
            local lock_count=$(jq '.packages | length' "$repo_path/package-lock.json" 2>/dev/null)
            [[ -n "$lock_count" ]] && total_count=$lock_count
        elif [[ -f "$repo_path/yarn.lock" ]]; then
            local yarn_count=$(grep -c '^"' "$repo_path/yarn.lock" 2>/dev/null)
            [[ -n "$yarn_count" ]] && total_count=$yarn_count
        else
            total_count=$direct_count
        fi
    fi

    # Extract from requirements.txt
    if [[ -f "$repo_path/requirements.txt" ]]; then
        local py_count=$(grep -c '^[^#]' "$repo_path/requirements.txt" 2>/dev/null)
        [[ -n "$py_count" ]] && direct_count=$((direct_count + py_count)) && total_count=$((total_count + py_count))
    fi

    # Extract from go.mod
    if [[ -f "$repo_path/go.mod" ]]; then
        local go_count=$(grep -c '^	' "$repo_path/go.mod" 2>/dev/null)
        [[ -n "$go_count" ]] && direct_count=$((direct_count + go_count)) && total_count=$((total_count + go_count))
    fi

    # Ensure we have valid numbers
    [[ -z "$direct_count" ]] && direct_count=0
    [[ -z "$total_count" ]] && total_count=0

    cat > "$output_path/dependencies.json" << EOF
{
  "analyzer": "dependency-extractor",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "direct_dependencies": $direct_count,
  "total_dependencies": $total_count,
  "summary": {
    "direct": $direct_count,
    "total": $total_count
  }
}
EOF
}

run_vulnerability_analyzer() {
    local repo_path="$1"
    local output_path="$2"

    # Use new data-only analyzer
    local vuln_script="$UTILS_ROOT/supply-chain/vulnerability-analysis/vulnerability-analyser-data.sh"

    if [[ -x "$vuln_script" ]]; then
        "$vuln_script" --local-path "$repo_path" -o "$output_path/vulnerabilities.json" 2>/dev/null
    else
        cat > "$output_path/vulnerabilities.json" << EOF
{
  "analyzer": "vulnerability-analyser",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "status": "analyzer_not_found",
  "vulnerabilities": [],
  "summary": {
    "critical": 0,
    "high": 0,
    "medium": 0,
    "low": 0
  }
}
EOF
    fi
}

run_package_health_analyzer() {
    local repo_path="$1"
    local output_path="$2"

    # TODO: Integrate with actual package health analyzer when JSON output is standardized
    cat > "$output_path/package-health.json" << EOF
{
  "analyzer": "package-health-analyser",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "status": "pending_integration",
  "note": "Run ./utils/supply-chain/package-health-analysis/package-health-analyser.sh for full report",
  "findings": [],
  "summary": {
    "abandoned": 0,
    "typosquat_risk": 0
  }
}
EOF
}

run_license_analyzer() {
    local repo_path="$1"
    local output_path="$2"

    # Use new data-only analyzer
    local legal_script="$UTILS_ROOT/legal-review/legal-analyser-data.sh"

    if [[ -x "$legal_script" ]]; then
        "$legal_script" --local-path "$repo_path" -o "$output_path/licenses.json" 2>/dev/null
    else
        cat > "$output_path/licenses.json" << EOF
{
  "analyzer": "legal-analyser",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "status": "analyzer_not_found",
  "licenses": [],
  "summary": {
    "overall_status": "unknown",
    "license_violations": 0
  }
}
EOF
    fi
}

run_code_security_analyzer() {
    local repo_path="$1"
    local output_path="$2"

    # Use new data-only analyzer
    local security_script="$UTILS_ROOT/code-security/code-security-data.sh"

    if [[ -x "$security_script" ]]; then
        "$security_script" --local-path "$repo_path" -o "$output_path/security-findings.json" 2>/dev/null
    else
        cat > "$output_path/security-findings.json" << EOF
{
  "analyzer": "code-security",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "status": "analyzer_not_found",
  "potential_issues": [],
  "summary": {
    "potential_issues": 0,
    "potential_secrets": 0
  }
}
EOF
    fi
}

run_ownership_analyzer() {
    local repo_path="$1"
    local output_path="$2"

    # Use new data-only analyzer
    local ownership_script="$UTILS_ROOT/code-ownership/ownership-analyser-data.sh"

    if [[ -x "$ownership_script" ]]; then
        "$ownership_script" --local-path "$repo_path" -o "$output_path/ownership.json" 2>/dev/null
    else
        cat > "$output_path/ownership.json" << EOF
{
  "analyzer": "code-ownership",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "status": "analyzer_not_found",
  "summary": {
    "total_files": 0,
    "active_contributors": 0
  }
}
EOF
    fi
}

run_dora_analyzer() {
    local repo_path="$1"
    local output_path="$2"

    # Use new data-only analyzer
    local dora_script="$UTILS_ROOT/dora-metrics/dora-analyser-data.sh"

    # Skip if not a git repo
    if [[ ! -d "$repo_path/.git" ]]; then
        cat > "$output_path/dora.json" << EOF
{
  "analyzer": "dora-metrics",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "status": "not_a_git_repo",
  "summary": {}
}
EOF
        return 0
    fi

    if [[ -x "$dora_script" ]]; then
        "$dora_script" --local-path "$repo_path" -o "$output_path/dora.json" 2>/dev/null
    else
        cat > "$output_path/dora.json" << EOF
{
  "analyzer": "dora-metrics",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "status": "analyzer_not_found",
  "summary": {}
}
EOF
    fi
}

run_provenance_analyzer() {
    local repo_path="$1"
    local output_path="$2"

    local prov_script="$UTILS_ROOT/supply-chain/provenance-analysis/provenance-analyser.sh"

    if [[ -x "$prov_script" ]]; then
        "$prov_script" --format json "$repo_path" > "$output_path/provenance.json" 2>/dev/null
    else
        cat > "$output_path/provenance.json" << EOF
{
  "analyzer": "provenance-analyser",
  "version": "1.0.0",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "status": "analyzer_not_found",
  "summary": {}
}
EOF
    fi
}

#############################################################################
# Run All Analyzers
#############################################################################

run_all_analyzers() {
    local repo_path="$1"
    local output_path="$2"
    local project_id="$3"
    local mode="$4"

    local analyzers=$(get_analyzers_for_mode "$mode")
    local analyzer_count=$(echo "$analyzers" | wc -w | tr -d ' ')
    local current=0

    echo -e "\n${BOLD}Running $analyzer_count analyzers...${NC}"
    echo

    for analyzer in $analyzers; do
        ((current++))

        # Show progress indicator
        printf "  [%d/%d] %-20s " "$current" "$analyzer_count" "$analyzer"

        local start_time=$(date +%s)
        run_analyzer "$analyzer" "$repo_path" "$output_path" "$project_id"
        local exit_code=$?
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))

        if [[ $exit_code -eq 0 ]]; then
            printf "${GREEN}✓${NC} %2ds  " "$duration"

            # Show inline summary
            local output_file="$output_path/${analyzer}.json"
            if [[ -f "$output_file" ]]; then
                case "$analyzer" in
                    vulnerabilities)
                        local c=$(jq -r '.summary.critical // 0' "$output_file" 2>/dev/null)
                        local h=$(jq -r '.summary.high // 0' "$output_file" 2>/dev/null)
                        local m=$(jq -r '.summary.medium // 0' "$output_file" 2>/dev/null)
                        local l=$(jq -r '.summary.low // 0' "$output_file" 2>/dev/null)
                        local total=$((c + h + m + l))
                        if [[ "$c" != "0" ]] || [[ "$h" != "0" ]]; then
                            echo -e "${RED}$c critical${NC}, ${YELLOW}$h high${NC}"
                        elif [[ "$total" -gt 0 ]]; then
                            echo "$total found"
                        else
                            echo -e "${GREEN}clean${NC}"
                        fi
                        ;;
                    dependencies)
                        local direct=$(jq -r '.direct_dependencies // 0' "$output_file" 2>/dev/null)
                        local total=$(jq -r '.total_dependencies // 0' "$output_file" 2>/dev/null)
                        echo "$direct direct, $total total"
                        ;;
                    technology)
                        local tech_count=$(jq -r '.summary.total // (.technologies | length) // 0' "$output_file" 2>/dev/null)
                        echo "$tech_count technologies"
                        ;;
                    licenses)
                        local status=$(jq -r '.summary.overall_status // "unknown"' "$output_file" 2>/dev/null)
                        local violations=$(jq -r '.summary.license_violations // 0' "$output_file" 2>/dev/null)
                        if [[ "$violations" -gt 0 ]]; then
                            echo -e "${RED}$violations violations${NC}"
                        elif [[ "$status" == "pass" ]]; then
                            echo -e "${GREEN}pass${NC}"
                        else
                            echo "$status"
                        fi
                        ;;
                    security-findings)
                        local issues=$(jq -r '.summary.potential_issues // 0' "$output_file" 2>/dev/null)
                        local secrets=$(jq -r '.summary.potential_secrets // 0' "$output_file" 2>/dev/null)
                        if [[ "$secrets" -gt 0 ]]; then
                            echo -e "${RED}$secrets secrets!${NC}"
                        elif [[ "$issues" -gt 0 ]]; then
                            echo -e "${YELLOW}$issues issues${NC}"
                        else
                            echo -e "${GREEN}clean${NC}"
                        fi
                        ;;
                    ownership)
                        local contributors=$(jq -r '.summary.active_contributors // 0' "$output_file" 2>/dev/null)
                        local bus=$(jq -r '.summary.estimated_bus_factor // 0' "$output_file" 2>/dev/null)
                        echo "$contributors contributors, bus factor $bus"
                        ;;
                    dora)
                        local perf=$(jq -r '.summary.overall_performance // "N/A"' "$output_file" 2>/dev/null)
                        local perf_color="$NC"
                        [[ "$perf" == "ELITE" ]] && perf_color="$GREEN"
                        [[ "$perf" == "HIGH" ]] && perf_color="$GREEN"
                        [[ "$perf" == "LOW" ]] && perf_color="$RED"
                        echo -e "${perf_color}$perf${NC}"
                        ;;
                    package-health)
                        echo "done"
                        ;;
                    *)
                        echo "done"
                        ;;
                esac
            else
                echo ""
            fi
        else
            printf "${RED}✗${NC} %2ds  " "$duration"
            echo -e "${RED}failed${NC}"
        fi
    done

    echo
}

#############################################################################
# Summary Generation
#############################################################################

generate_summary() {
    local project_id="$1"
    local output_path="$2"

    local risk_level="low"
    local total_deps=0
    local direct_deps=0
    local total_vulns=0
    local total_findings=0
    local license_status="unknown"
    local abandoned=0

    # Read from analysis outputs
    if [[ -f "$output_path/dependencies.json" ]]; then
        direct_deps=$(jq -r '.direct_dependencies // 0' "$output_path/dependencies.json" 2>/dev/null)
        total_deps=$(jq -r '.total_dependencies // 0' "$output_path/dependencies.json" 2>/dev/null)
    fi

    if [[ -f "$output_path/vulnerabilities.json" ]]; then
        local critical=$(jq -r '.summary.critical // 0' "$output_path/vulnerabilities.json" 2>/dev/null)
        local high=$(jq -r '.summary.high // 0' "$output_path/vulnerabilities.json" 2>/dev/null)
        local medium=$(jq -r '.summary.medium // 0' "$output_path/vulnerabilities.json" 2>/dev/null)
        local low_v=$(jq -r '.summary.low // 0' "$output_path/vulnerabilities.json" 2>/dev/null)
        total_vulns=$((critical + high + medium + low_v))

        if [[ "$critical" -gt 0 ]]; then
            risk_level="critical"
        elif [[ "$high" -gt 0 ]]; then
            risk_level="high"
        elif [[ "$medium" -gt 0 ]]; then
            risk_level="medium"
        fi
    fi

    if [[ -f "$output_path/security-findings.json" ]]; then
        # code-security-data.sh outputs potential_issues count
        total_findings=$(jq -r '.summary.potential_issues // 0' "$output_path/security-findings.json" 2>/dev/null)
        local secrets=$(jq -r '.summary.potential_secrets // 0' "$output_path/security-findings.json" 2>/dev/null)
        if [[ "$secrets" -gt 0 ]]; then
            risk_level="critical"
        fi
    fi

    if [[ -f "$output_path/licenses.json" ]]; then
        # legal-analyser-data.sh outputs overall_status
        license_status=$(jq -r '.summary.overall_status // "unknown"' "$output_path/licenses.json" 2>/dev/null)
        local violations=$(jq -r '.summary.license_violations // 0' "$output_path/licenses.json" 2>/dev/null)
        if [[ "$violations" -gt 0 ]]; then
            risk_level="high"
        fi
    fi

    if [[ -f "$output_path/package-health.json" ]]; then
        abandoned=$(jq -r '.summary.abandoned // 0' "$output_path/package-health.json" 2>/dev/null)
    fi

    # Update manifest summary
    gibson_update_summary "$project_id" "$risk_level" "$total_deps" "$direct_deps" "$total_vulns" "$total_findings" "$license_status" "$abandoned"
}

print_final_summary() {
    local project_id="$1"
    local output_path="$2"

    echo
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${BOLD}HYDRATION COMPLETE${NC} ${DIM}(static analysis - no AI)${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    local manifest="$output_path/manifest.json"
    if [[ -f "$manifest" ]]; then
        local risk=$(jq -r '.summary.risk_level' "$manifest" 2>/dev/null)
        local risk_color="$GREEN"
        case "$risk" in
            critical) risk_color="$RED" ;;
            high) risk_color="$RED" ;;
            medium) risk_color="$YELLOW" ;;
        esac
        local risk_upper=$(echo "$risk" | tr '[:lower:]' '[:upper:]')
        echo -e "\nRisk Level: ${risk_color}${risk_upper}${NC}"

        echo

        # Package Vulnerabilities (CVEs in dependencies)
        if [[ -f "$output_path/vulnerabilities.json" ]]; then
            local c=$(jq -r '.summary.critical // 0' "$output_path/vulnerabilities.json" 2>/dev/null)
            local h=$(jq -r '.summary.high // 0' "$output_path/vulnerabilities.json" 2>/dev/null)
            local m=$(jq -r '.summary.medium // 0' "$output_path/vulnerabilities.json" 2>/dev/null)
            local l=$(jq -r '.summary.low // 0' "$output_path/vulnerabilities.json" 2>/dev/null)
            local total=$((c + h + m + l))
            printf "  %-18s " "Pkg Vulnerabilities:"
            if [[ $c -gt 0 ]] || [[ $h -gt 0 ]]; then
                echo -e "${RED}$c critical${NC}, ${YELLOW}$h high${NC}, $m medium, $l low"
            elif [[ $total -gt 0 ]]; then
                echo "$total found ($m medium, $l low)"
            else
                echo -e "${GREEN}None found${NC}"
            fi
        fi

        # Security findings
        if [[ -f "$output_path/security-findings.json" ]]; then
            local issues=$(jq -r '.summary.potential_issues // 0' "$output_path/security-findings.json" 2>/dev/null)
            local secrets=$(jq -r '.summary.potential_secrets // 0' "$output_path/security-findings.json" 2>/dev/null)
            local files=$(jq -r '.summary.total_files // 0' "$output_path/security-findings.json" 2>/dev/null)
            printf "  %-18s " "Code Security:"
            if [[ $secrets -gt 0 ]]; then
                echo -e "${RED}$secrets potential secrets${NC}, $issues code issues"
            elif [[ $issues -gt 0 ]]; then
                echo -e "${YELLOW}$issues potential issues${NC} in $files files"
            else
                echo -e "${GREEN}Clean${NC} ($files files scanned)"
            fi
        fi

        # Dependencies
        if [[ -f "$output_path/dependencies.json" ]]; then
            local direct=$(jq -r '.direct_dependencies // 0' "$output_path/dependencies.json" 2>/dev/null)
            local total=$(jq -r '.total_dependencies // 0' "$output_path/dependencies.json" 2>/dev/null)
            printf "  %-18s %s\n" "Dependencies:" "$direct direct, $total total"
        fi

        # Licenses
        if [[ -f "$output_path/licenses.json" ]]; then
            local lic_status=$(jq -r '.summary.overall_status // "unknown"' "$output_path/licenses.json" 2>/dev/null)
            local violations=$(jq -r '.summary.license_violations // 0' "$output_path/licenses.json" 2>/dev/null)
            printf "  %-18s " "Licenses:"
            if [[ "$lic_status" == "fail" ]] || [[ $violations -gt 0 ]]; then
                echo -e "${RED}✗ $violations violations${NC}"
            elif [[ "$lic_status" == "warning" ]]; then
                echo -e "${YELLOW}⚠ Review needed${NC}"
            elif [[ "$lic_status" == "pass" ]]; then
                echo -e "${GREEN}✓ All clear${NC}"
            else
                echo "Unknown"
            fi
        fi

        # Ownership
        if [[ -f "$output_path/ownership.json" ]]; then
            local contributors=$(jq -r '.summary.active_contributors // 0' "$output_path/ownership.json" 2>/dev/null)
            local bus_factor=$(jq -r '.summary.estimated_bus_factor // 0' "$output_path/ownership.json" 2>/dev/null)
            local bus_risk=$(jq -r '.risk_assessment.bus_factor_risk // "unknown"' "$output_path/ownership.json" 2>/dev/null)
            local bus_color="$GREEN"
            [[ "$bus_risk" == "critical" ]] && bus_color="$RED"
            [[ "$bus_risk" == "high" ]] && bus_color="$RED"
            [[ "$bus_risk" == "medium" ]] && bus_color="$YELLOW"
            printf "  %-18s %d active, bus factor: ${bus_color}%d (%s)${NC}\n" "Ownership:" "$contributors" "$bus_factor" "$bus_risk"
        fi

        # DORA
        if [[ -f "$output_path/dora.json" ]]; then
            local dora_perf=$(jq -r '.summary.overall_performance // "N/A"' "$output_path/dora.json" 2>/dev/null)
            local dora_color="$GREEN"
            [[ "$dora_perf" == "LOW" ]] && dora_color="$RED"
            [[ "$dora_perf" == "MEDIUM" ]] && dora_color="$YELLOW"
            printf "  %-18s ${dora_color}%s${NC}\n" "DORA Performance:" "$dora_perf"
        fi
    fi

    echo
    echo -e "${BOLD}Agents ready:${NC}"
    echo -e "  Scout     → ${CYAN}/phantom ask scout ...${NC}"
    echo -e "  Sentinel  → ${CYAN}/phantom ask sentinel ...${NC}"
    echo -e "  Quinn     → ${CYAN}/phantom ask quinn ...${NC}"
    echo -e "  Harper    → ${CYAN}/phantom ask harper ...${NC}"

    echo
    local size=$(gibson_project_size "$project_id")
    echo -e "Storage: ${CYAN}~/.phantom/projects/$project_id/${NC} ($size)"
}

#############################################################################
# Main
#############################################################################

main() {
    parse_args "$@"

    # Run preflight check (silent unless errors)
    if [[ -x "$SCRIPT_DIR/preflight.sh" ]]; then
        if ! "$SCRIPT_DIR/preflight.sh" > /dev/null 2>&1; then
            echo -e "${RED}Preflight check failed. Run ./utils/phantom/preflight.sh to see details.${NC}"
            exit 1
        fi
    fi

    # Ensure Gibson is initialized
    gibson_ensure_initialized

    # Generate project ID
    local project_id=$(gibson_project_id "$TARGET")

    # Check if project already exists
    if gibson_project_exists "$project_id" && [[ "$FORCE" != "true" ]]; then
        echo -e "${YELLOW}Project '$project_id' already exists.${NC}"
        echo "Use --force to re-bootstrap, or /phantom refresh to update analysis."
        exit 1
    fi

    # Print header
    gibson_print_header
    echo -e "Target: ${CYAN}$TARGET${NC}"
    echo -e "Project ID: ${CYAN}$project_id${NC}"
    echo

    # Setup paths
    local project_path=$(gibson_project_path "$project_id")
    local repo_path=$(gibson_project_repo_path "$project_id")
    local analysis_path=$(gibson_project_analysis_path "$project_id")

    # Clean up if force
    if [[ "$FORCE" == "true" ]] && [[ -d "$project_path" ]]; then
        rm -rf "$project_path"
    fi

    # Create directories
    mkdir -p "$project_path"
    mkdir -p "$analysis_path"

    # Add to index
    gibson_index_add_project "$project_id" "$TARGET" "bootstrapping"

    # Clone or copy project
    echo -n "Cloning..."
    if gibson_is_local_source "$TARGET"; then
        copy_local_project "$TARGET" "$repo_path"
        echo -e " ${GREEN}✓${NC} (local copy)"
    else
        local clone_url=$(gibson_clone_url "$TARGET")
        if [[ -z "$clone_url" ]]; then
            echo -e " ${RED}✗${NC}"
            echo "Error: Could not determine clone URL for '$TARGET'"
            exit 1
        fi

        local clone_output
        clone_output=$(clone_github_repo "$clone_url" "$repo_path" "$BRANCH" "$DEPTH" 2>&1)
        if [[ $? -ne 0 ]]; then
            echo -e " ${RED}✗${NC}"
            echo "Clone failed: $clone_output"
            gibson_index_remove_project "$project_id"
            exit 1
        fi
        echo -e " ${GREEN}✓${NC}"
    fi

    # Get git info
    local git_info=$(get_git_info "$repo_path")
    local branch=$(echo "$git_info" | sed -n '1p')
    local commit=$(echo "$git_info" | sed -n '2p')

    if [[ -n "$branch" ]]; then
        echo -e "  Branch: ${CYAN}$branch${NC}"
        echo -e "  Commit: ${CYAN}$commit${NC}"
    fi

    # Detect project type (silent - no separate output line)
    local detection=$(detect_project_type "$repo_path")
    local languages=$(echo "$detection" | sed -n '1p')
    local frameworks=$(echo "$detection" | sed -n '2p')
    local package_managers=$(echo "$detection" | sed -n '3p')

    local langs_display=$(echo "$languages" | jq -r 'join(", ")' 2>/dev/null)
    local fwks_display=$(echo "$frameworks" | jq -r 'join(", ")' 2>/dev/null)
    # Always print Languages line (hydrate.sh uses this as phase marker)
    echo -e "  Languages: ${CYAN}${langs_display:-unknown}${NC}"
    [[ -n "$fwks_display" ]] && echo -e "  Frameworks: ${CYAN}$fwks_display${NC}"

    # Determine source type
    local source_type="github"
    gibson_is_local_source "$TARGET" && source_type="local"

    # Create project metadata
    gibson_create_project_metadata "$project_id" "$TARGET" "$source_type" "$branch" "$commit"
    gibson_update_project_type "$project_id" "$languages" "$frameworks" "$package_managers"

    # Initialize analysis manifest
    gibson_init_analysis_manifest "$project_id" "$commit"

    # Run analyzers
    run_all_analyzers "$repo_path" "$analysis_path" "$project_id" "$MODE"

    # Generate summary
    generate_summary "$project_id" "$analysis_path"

    # Finalize manifest
    gibson_finalize_manifest "$project_id"

    # Update index status
    gibson_index_update_status "$project_id" "ready"

    # Set as active project
    gibson_set_active_project "$project_id"

    # Print final summary
    print_final_summary "$project_id" "$analysis_path"
}

main "$@"
