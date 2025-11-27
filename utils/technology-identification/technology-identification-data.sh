#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Technology Identification - Data Collector
# Pure data collection - outputs JSON for agent analysis
#
# This is the data-only version. AI analysis is handled by agents.
#
# Usage: ./technology-identification-data.sh [options] <target>
# Output: JSON with detected technologies, categories, and confidence scores
#############################################################################

set -e

# Colors for terminal output (stderr only)
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
UTILS_ROOT="$(dirname "$SCRIPT_DIR")"
REPO_ROOT="$(dirname "$UTILS_ROOT")"

# Load global libraries
source "$UTILS_ROOT/lib/sbom.sh" 2>/dev/null || true
source "$SCRIPT_DIR/lib/pattern-loader.sh" 2>/dev/null || true

# RAG patterns directory
RAG_ROOT="$REPO_ROOT/rag/technology-identification"

# Default options
OUTPUT_FILE=""
LOCAL_PATH=""
TEMP_DIR=""
CLEANUP=true
TARGET=""
CONFIDENCE_THRESHOLD=50

usage() {
    cat << EOF
Technology Identification - Data Collector (JSON output for agent analysis)

Usage: $0 [OPTIONS] <target>

TARGET:
    Git repository URL      Clone and analyze repository
    Local directory path    Analyze local repository
    SBOM file path          Analyze existing SBOM

OPTIONS:
    --local-path PATH       Use pre-cloned repository (skips cloning)
    --confidence N          Minimum confidence threshold (0-100, default: 50)
    -o, --output FILE       Write JSON to file (default: stdout)
    -k, --keep-clone        Keep cloned repository
    -h, --help              Show this help

OUTPUT:
    JSON object with:
    - metadata: scan timestamp, target, analyzer version
    - summary: counts by category
    - technologies: array of findings with confidence scores

EXAMPLES:
    $0 https://github.com/expressjs/express
    $0 --local-path ~/.gibson/projects/foo/repo
    $0 -o tech.json /path/to/project

EOF
    exit 0
}

# Check if syft is installed
check_syft() {
    if ! command -v syft &> /dev/null; then
        echo '{"error": "syft not installed", "install": "brew install syft"}' >&2
        exit 1
    fi
}

# Detect if target is a Git URL
is_git_url() {
    [[ "$1" =~ ^(https?|git)://.*\.git$ ]] || [[ "$1" =~ ^git@.*:.*\.git$ ]] || [[ "$1" =~ github\.com|gitlab\.com|bitbucket\.org ]]
}

# Detect if target is an SBOM file
is_sbom_file() {
    local file="$1"
    [[ -f "$file" ]] && ([[ "$file" =~ \.json$ ]] || [[ "$file" =~ \.xml$ ]] || [[ "$file" =~ \.cdx\. ]] || [[ "$file" =~ bom\. ]])
}

# Clone repository
clone_repository() {
    local repo_url="$1"
    TEMP_DIR=$(mktemp -d)

    echo -e "${BLUE}Cloning repository...${NC}" >&2
    if git clone --depth 1 "$repo_url" "$TEMP_DIR" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Cloned${NC}" >&2
        return 0
    else
        echo '{"error": "Failed to clone repository"}'
        exit 1
    fi
}

# Generate SBOM if needed
generate_sbom() {
    local target_dir="$1"
    local sbom_file=$(mktemp)

    echo -e "${BLUE}Generating SBOM...${NC}" >&2
    if syft scan "$target_dir" -o json > "$sbom_file" 2>/dev/null; then
        echo -e "${GREEN}✓ SBOM generated${NC}" >&2
        echo "$sbom_file"
        return 0
    else
        echo -e "${YELLOW}⚠ SBOM generation failed${NC}" >&2
        return 1
    fi
}

# Cleanup
cleanup() {
    if [[ "$CLEANUP" == true ]] && [[ -n "$TEMP_DIR" ]] && [[ -d "$TEMP_DIR" ]]; then
        rm -rf "$TEMP_DIR"
    fi
}
trap cleanup EXIT

# Layer 1: Scan SBOM for package dependencies
scan_sbom_packages() {
    local sbom_file="$1"
    local findings="[]"

    # Extract components from SBOM
    local components=$(jq -c '.components[]?' "$sbom_file" 2>/dev/null)

    [[ -z "$components" ]] && { echo "[]"; return; }

    while IFS= read -r component; do
        local name=$(echo "$component" | jq -r '.name // ""')
        local version=$(echo "$component" | jq -r '.version // ""')

        # Try RAG pattern match
        local match_result=""
        if type match_package_name &>/dev/null; then
            match_result=$(match_package_name "$name" 2>/dev/null || echo "")
        fi

        local tech_category=""
        local tech_name=""
        local confidence=95

        if [[ -n "$match_result" ]]; then
            tech_name=$(echo "$match_result" | jq -r '.technology // ""' 2>/dev/null)
            tech_category=$(echo "$match_result" | jq -r '.category // ""' 2>/dev/null)
            confidence=$(echo "$match_result" | jq -r '.confidence // 95' 2>/dev/null)
        else
            # Fallback pattern matching for common packages
            case "$name" in
                stripe) tech_category="business-tools/payment"; tech_name="Stripe" ;;
                express) tech_category="web-frameworks/backend"; tech_name="Express" ;;
                react|react-dom) tech_category="web-frameworks/frontend"; tech_name="React" ;;
                vue) tech_category="web-frameworks/frontend"; tech_name="Vue.js" ;;
                angular|@angular/*) tech_category="web-frameworks/frontend"; tech_name="Angular" ;;
                next) tech_category="web-frameworks/frontend"; tech_name="Next.js" ;;
                django) tech_category="web-frameworks/backend"; tech_name="Django" ;;
                flask) tech_category="web-frameworks/backend"; tech_name="Flask" ;;
                fastapi) tech_category="web-frameworks/backend"; tech_name="FastAPI" ;;
                pg|postgres|postgresql) tech_category="databases/relational"; tech_name="PostgreSQL" ;;
                mysql|mysql2) tech_category="databases/relational"; tech_name="MySQL" ;;
                mongodb|mongoose) tech_category="databases/nosql"; tech_name="MongoDB" ;;
                redis|ioredis) tech_category="databases/keyvalue"; tech_name="Redis" ;;
                boto3|botocore|aws-sdk|@aws-sdk/*) tech_category="cloud-providers/aws"; tech_name="AWS SDK" ;;
                openai) tech_category="ai-ml/llm-apis"; tech_name="OpenAI" ;;
                anthropic|@anthropic-ai/sdk) tech_category="ai-ml/llm-apis"; tech_name="Anthropic" ;;
                langchain|langchain-*) tech_category="ai-ml/frameworks"; tech_name="LangChain" ;;
                tensorflow|tf-*) tech_category="ai-ml/frameworks"; tech_name="TensorFlow" ;;
                pytorch|torch) tech_category="ai-ml/frameworks"; tech_name="PyTorch" ;;
                jest) tech_category="testing/unit"; tech_name="Jest" ;;
                pytest) tech_category="testing/unit"; tech_name="Pytest" ;;
                cypress) tech_category="testing/e2e"; tech_name="Cypress" ;;
                playwright) tech_category="testing/e2e"; tech_name="Playwright" ;;
                *) continue ;;
            esac
        fi

        [[ -z "$tech_category" ]] && continue

        local finding=$(jq -n \
            --arg name "$tech_name" \
            --arg category "$tech_category" \
            --arg version "$version" \
            --argjson confidence "$confidence" \
            --arg method "sbom-package" \
            --arg evidence "package dependency: $name@$version" \
            '{name: $name, category: $category, version: $version, confidence: $confidence, detection_method: $method, evidence: [$evidence]}')

        findings=$(echo "$findings" | jq --argjson f "$finding" '. + [$f]')
    done <<< "$components"

    echo "$findings"
}

# Layer 2: Scan for configuration files
scan_config_files() {
    local repo_path="$1"
    local findings="[]"

    # Dockerfile
    if [[ -f "$repo_path/Dockerfile" ]] || find "$repo_path" -name "Dockerfile*" -type f 2>/dev/null | grep -q .; then
        findings=$(echo "$findings" | jq '. + [{"name": "Docker", "category": "developer-tools/containers", "confidence": 90, "detection_method": "config-file", "evidence": ["Dockerfile found"]}]')
    fi

    # docker-compose
    if [[ -f "$repo_path/docker-compose.yml" ]] || [[ -f "$repo_path/docker-compose.yaml" ]]; then
        findings=$(echo "$findings" | jq '. + [{"name": "Docker Compose", "category": "developer-tools/containers", "confidence": 90, "detection_method": "config-file", "evidence": ["docker-compose.yml found"]}]')
    fi

    # Terraform
    if find "$repo_path" -name "*.tf" -type f 2>/dev/null | grep -q .; then
        findings=$(echo "$findings" | jq '. + [{"name": "Terraform", "category": "developer-tools/infrastructure", "confidence": 90, "detection_method": "config-file", "evidence": ["*.tf files found"]}]')
    fi

    # Kubernetes
    if find "$repo_path" -name "*.yaml" -o -name "*.yml" -type f 2>/dev/null | xargs grep -l "kind:" 2>/dev/null | grep -q .; then
        findings=$(echo "$findings" | jq '. + [{"name": "Kubernetes", "category": "developer-tools/containers", "confidence": 85, "detection_method": "config-file", "evidence": ["Kubernetes manifests found"]}]')
    fi

    # GitHub Actions
    if [[ -d "$repo_path/.github/workflows" ]]; then
        findings=$(echo "$findings" | jq '. + [{"name": "GitHub Actions", "category": "developer-tools/cicd", "confidence": 95, "detection_method": "config-file", "evidence": [".github/workflows directory found"]}]')
    fi

    # GitLab CI
    if [[ -f "$repo_path/.gitlab-ci.yml" ]]; then
        findings=$(echo "$findings" | jq '. + [{"name": "GitLab CI", "category": "developer-tools/cicd", "confidence": 95, "detection_method": "config-file", "evidence": [".gitlab-ci.yml found"]}]')
    fi

    # package.json (Node.js)
    if [[ -f "$repo_path/package.json" ]]; then
        findings=$(echo "$findings" | jq '. + [{"name": "Node.js", "category": "languages/runtime", "confidence": 95, "detection_method": "config-file", "evidence": ["package.json found"]}]')
    fi

    # requirements.txt / pyproject.toml (Python)
    if [[ -f "$repo_path/requirements.txt" ]] || [[ -f "$repo_path/pyproject.toml" ]]; then
        findings=$(echo "$findings" | jq '. + [{"name": "Python", "category": "languages/runtime", "confidence": 95, "detection_method": "config-file", "evidence": ["Python project files found"]}]')
    fi

    # go.mod (Go)
    if [[ -f "$repo_path/go.mod" ]]; then
        findings=$(echo "$findings" | jq '. + [{"name": "Go", "category": "languages/runtime", "confidence": 95, "detection_method": "config-file", "evidence": ["go.mod found"]}]')
    fi

    # Cargo.toml (Rust)
    if [[ -f "$repo_path/Cargo.toml" ]]; then
        findings=$(echo "$findings" | jq '. + [{"name": "Rust", "category": "languages/runtime", "confidence": 95, "detection_method": "config-file", "evidence": ["Cargo.toml found"]}]')
    fi

    echo "$findings"
}

# Layer 3: Scan environment variables
scan_env_variables() {
    local repo_path="$1"
    local findings="[]"

    local env_files=$(find "$repo_path" -maxdepth 3 -type f \( -name ".env*" -o -name "*.env" \) 2>/dev/null)
    [[ -z "$env_files" ]] && { echo "[]"; return; }

    local env_content=$(cat $env_files 2>/dev/null)

    if echo "$env_content" | grep -q "STRIPE"; then
        findings=$(echo "$findings" | jq '. + [{"name": "Stripe", "category": "business-tools/payment", "confidence": 65, "detection_method": "env-variable", "evidence": ["STRIPE_* environment variables found"]}]')
    fi

    if echo "$env_content" | grep -q "AWS_"; then
        findings=$(echo "$findings" | jq '. + [{"name": "AWS", "category": "cloud-providers/aws", "confidence": 65, "detection_method": "env-variable", "evidence": ["AWS_* environment variables found"]}]')
    fi

    if echo "$env_content" | grep -q "OPENAI"; then
        findings=$(echo "$findings" | jq '. + [{"name": "OpenAI", "category": "ai-ml/llm-apis", "confidence": 65, "detection_method": "env-variable", "evidence": ["OPENAI_* environment variables found"]}]')
    fi

    if echo "$env_content" | grep -q "ANTHROPIC"; then
        findings=$(echo "$findings" | jq '. + [{"name": "Anthropic", "category": "ai-ml/llm-apis", "confidence": 65, "detection_method": "env-variable", "evidence": ["ANTHROPIC_* environment variables found"]}]')
    fi

    if echo "$env_content" | grep -q "DATABASE_URL\|DB_"; then
        findings=$(echo "$findings" | jq '. + [{"name": "Database", "category": "databases", "confidence": 60, "detection_method": "env-variable", "evidence": ["Database environment variables found"]}]')
    fi

    echo "$findings"
}

# Aggregate and deduplicate findings
aggregate_findings() {
    local all_findings="$1"

    # Group by technology name and calculate composite confidence
    echo "$all_findings" | jq '
        group_by(.name) |
        map({
            name: .[0].name,
            category: .[0].category,
            version: (map(.version // "") | map(select(length > 0)) | .[0] // ""),
            confidence: (
                map(.confidence) |
                (add / length * 1.2) |
                if . > 100 then 100 else . end |
                floor
            ),
            detection_methods: map(.detection_method) | unique,
            evidence: map(.evidence[]) | unique
        }) |
        sort_by(-.confidence)
    ' 2>/dev/null || echo "[]"
}

# Main analysis
analyze_target() {
    local repo_path="$1"
    local sbom_file=""

    # Generate SBOM
    sbom_file=$(generate_sbom "$repo_path")

    echo -e "${BLUE}Running technology detection...${NC}" >&2

    # Run all detection layers
    local layer1="[]"
    local layer2="[]"
    local layer3="[]"

    if [[ -n "$sbom_file" ]] && [[ -f "$sbom_file" ]]; then
        layer1=$(scan_sbom_packages "$sbom_file")
        rm -f "$sbom_file"
    fi

    layer2=$(scan_config_files "$repo_path")
    layer3=$(scan_env_variables "$repo_path")

    # Combine all findings
    local all_findings=$(echo "$layer1 $layer2 $layer3" | jq -s 'add')

    # Aggregate and deduplicate
    local results=$(aggregate_findings "$all_findings")

    # Filter by confidence threshold
    results=$(echo "$results" | jq --argjson threshold "$CONFIDENCE_THRESHOLD" 'map(select(.confidence >= $threshold))')

    echo "$results"
}

# Generate final JSON output
generate_output() {
    local findings="$1"
    local target="$2"

    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    local total=$(echo "$findings" | jq 'length')

    # Count by category
    local by_category=$(echo "$findings" | jq 'group_by(.category) | map({key: .[0].category, value: length}) | from_entries')

    # Confidence distribution
    local high=$(echo "$findings" | jq '[.[] | select(.confidence >= 80)] | length')
    local medium=$(echo "$findings" | jq '[.[] | select(.confidence >= 60 and .confidence < 80)] | length')
    local low=$(echo "$findings" | jq '[.[] | select(.confidence < 60)] | length')

    jq -n \
        --arg ts "$timestamp" \
        --arg tgt "$target" \
        --arg ver "1.0.0" \
        --argjson total "$total" \
        --argjson by_cat "$by_category" \
        --argjson hi "$high" \
        --argjson med "$medium" \
        --argjson lo "$low" \
        --argjson techs "$findings" \
        '{
            analyzer: "technology-identification",
            version: $ver,
            timestamp: $ts,
            target: $tgt,
            summary: {
                total: $total,
                by_category: $by_cat,
                confidence_distribution: {
                    high: $hi,
                    medium: $med,
                    low: $lo
                }
            },
            technologies: $techs
        }'
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help) usage ;;
        --local-path)
            LOCAL_PATH="$2"
            shift 2
            ;;
        --confidence)
            CONFIDENCE_THRESHOLD="$2"
            shift 2
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

# Load RAG patterns if available
if [[ -d "$RAG_ROOT" ]] && type load_all_patterns &>/dev/null; then
    echo -e "${BLUE}Loading technology patterns...${NC}" >&2
    load_all_patterns "$RAG_ROOT" 2>/dev/null || true
fi

# Main execution
check_syft

scan_path=""
repo_name=""

if [[ -n "$LOCAL_PATH" ]]; then
    [[ ! -d "$LOCAL_PATH" ]] && { echo '{"error": "Local path does not exist"}'; exit 1; }
    scan_path="$LOCAL_PATH"
    repo_name=$(basename "$LOCAL_PATH")
elif [[ -n "$TARGET" ]]; then
    if is_git_url "$TARGET"; then
        repo_name=$(echo "$TARGET" | sed -E 's|.*[:/]([^/]+/[^/]+)(\.git)?$|\1|')
        clone_repository "$TARGET"
        scan_path="$TEMP_DIR"
    elif [[ -d "$TARGET" ]]; then
        scan_path="$TARGET"
        repo_name=$(basename "$TARGET")
    elif is_sbom_file "$TARGET"; then
        # Direct SBOM analysis
        echo -e "${BLUE}Analyzing SBOM file...${NC}" >&2
        findings=$(scan_sbom_packages "$TARGET")
        final_json=$(generate_output "$findings" "$TARGET")
        if [[ -n "$OUTPUT_FILE" ]]; then
            echo "$final_json" > "$OUTPUT_FILE"
            echo -e "${GREEN}✓ Results written to $OUTPUT_FILE${NC}" >&2
        else
            echo "$final_json"
        fi
        exit 0
    else
        echo '{"error": "Invalid target - must be URL, directory, or SBOM file"}'
        exit 1
    fi
else
    echo '{"error": "No target specified"}'
    exit 1
fi

echo -e "${BLUE}Scanning: $repo_name${NC}" >&2

# Run analysis
findings=$(analyze_target "$scan_path")
final_json=$(generate_output "$findings" "${TARGET:-$scan_path}")

# Output
if [[ -n "$OUTPUT_FILE" ]]; then
    echo "$final_json" > "$OUTPUT_FILE"
    echo -e "${GREEN}✓ Results written to $OUTPUT_FILE${NC}" >&2
else
    echo "$final_json"
fi
