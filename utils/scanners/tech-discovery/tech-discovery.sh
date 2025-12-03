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
SCANNERS_ROOT="$(dirname "$SCRIPT_DIR")"
UTILS_ROOT="$(dirname "$SCANNERS_ROOT")"
REPO_ROOT="$(dirname "$UTILS_ROOT")"

# Load global libraries
source "$UTILS_ROOT/lib/sbom.sh" 2>/dev/null || true
source "$SCRIPT_DIR/lib/pattern-loader.sh" 2>/dev/null || true

# RAG patterns directory
RAG_ROOT="$REPO_ROOT/rag/technology-identification"

# Default options
OUTPUT_FILE=""
LOCAL_PATH=""
SBOM_FILE=""
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
    --sbom FILE             Use existing SBOM file (skips syft generation)
    --confidence N          Minimum confidence threshold (0-100, default: 50)
    --scan-docker-images    Also scan Docker images referenced in Dockerfile/compose
    -o, --output FILE       Write JSON to file (default: stdout)
    -k, --keep-clone        Keep cloned repository
    -h, --help              Show this help

OUTPUT:
    JSON object with:
    - metadata: scan timestamp, target, analyzer version
    - summary: counts by category
    - technologies: array of findings with confidence scores

DETECTION LAYERS:
    1. SBOM packages (via syft)
    2. Config files (Dockerfile, package.json, etc.)
    3. Environment variables (.env files)
    4. Docker image scanning (optional, --scan-docker-images)
    5. AI import patterns (scans source files for AI library imports)

EXAMPLES:
    $0 https://github.com/expressjs/express
    $0 --local-path ~/.gibson/projects/foo/repo
    $0 --scan-docker-images /path/to/docker/project
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
    if syft scan "$target_dir" -o cyclonedx-json > "$sbom_file" 2>/dev/null; then
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
            # No RAG pattern match - skip this package
            # All technology patterns should be defined in rag/technology-identification/
            continue
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

# Layer 4: Scan Docker images for packages
scan_docker_images() {
    local repo_path="$1"
    local findings="[]"

    # Find Dockerfiles
    local dockerfiles=$(find "$repo_path" -maxdepth 3 -type f -name "Dockerfile*" 2>/dev/null)
    [[ -z "$dockerfiles" ]] && { echo "[]"; return; }

    # Extract image names from FROM statements and docker-compose
    local images=""

    # From Dockerfiles
    for df in $dockerfiles; do
        local img=$(grep -E "^FROM " "$df" 2>/dev/null | head -1 | awk '{print $2}' | sed 's/:.*$//')
        [[ -n "$img" ]] && images="$images $img"
    done

    # From docker-compose.yml
    if [[ -f "$repo_path/docker-compose.yml" ]] || [[ -f "$repo_path/docker-compose.yaml" ]]; then
        local compose_file="$repo_path/docker-compose.yml"
        [[ -f "$repo_path/docker-compose.yaml" ]] && compose_file="$repo_path/docker-compose.yaml"
        local compose_images=$(grep -E "^\s+image:" "$compose_file" 2>/dev/null | sed 's/.*image:\s*//' | sed 's/:.*$//' | tr -d '"' | tr -d "'")
        images="$images $compose_images"
    fi

    # Unique images
    images=$(echo "$images" | tr ' ' '\n' | sort -u | grep -v '^$')
    [[ -z "$images" ]] && { echo "[]"; return; }

    echo -e "${BLUE}Scanning Docker images for AI packages...${NC}" >&2

    for image in $images; do
        # Skip base images that won't have interesting packages
        [[ "$image" =~ ^(alpine|ubuntu|debian|busybox|scratch)$ ]] && continue

        # Try to scan the image with syft (will pull if not local)
        local temp_sbom=$(mktemp)
        if timeout 60 syft scan "registry:$image:latest" -o cyclonedx-json > "$temp_sbom" 2>/dev/null; then
            echo -e "${GREEN}✓ Scanned Docker image: $image${NC}" >&2

            # Scan for AI packages in the image SBOM
            local image_findings=$(scan_sbom_packages "$temp_sbom")

            # Add source info to evidence
            image_findings=$(echo "$image_findings" | jq --arg img "$image" '
                map(. + {
                    detection_method: "docker-image-sbom",
                    evidence: (.evidence + ["from Docker image: " + $img])
                })
            ')

            findings=$(echo "$findings $image_findings" | jq -s 'add')
        else
            echo -e "${YELLOW}⚠ Could not scan Docker image: $image${NC}" >&2
        fi
        rm -f "$temp_sbom"
    done

    echo "$findings"
}

# Layer 5: Scan source files for AI import patterns
scan_ai_imports() {
    local repo_path="$1"
    local findings="[]"

    echo -e "${BLUE}Scanning source files for AI imports...${NC}" >&2

    # AI-related import patterns with their technology mappings
    # Format: pattern|technology|category|file_glob
    local patterns=(
        # OpenAI
        'import.*openai|OpenAI|ai-ml/llm-apis|*.py'
        'from openai import|OpenAI|ai-ml/llm-apis|*.py'
        "import.*from ['\"]openai['\"]|OpenAI|ai-ml/llm-apis|*.ts,*.js,*.tsx,*.jsx"
        'new OpenAI\(|OpenAI|ai-ml/llm-apis|*.ts,*.js,*.tsx,*.jsx'

        # Anthropic
        'import anthropic|Anthropic|ai-ml/llm-apis|*.py'
        'from anthropic import|Anthropic|ai-ml/llm-apis|*.py'
        "import.*from ['\"]@anthropic-ai/sdk['\"]|Anthropic|ai-ml/llm-apis|*.ts,*.js,*.tsx,*.jsx"

        # LangChain
        'from langchain|LangChain|ai-ml/frameworks|*.py'
        'import langchain|LangChain|ai-ml/frameworks|*.py'
        "import.*from ['\"]langchain['\"]|LangChain|ai-ml/frameworks|*.ts,*.js,*.tsx,*.jsx"
        "import.*from ['\"]@langchain/|LangChain|ai-ml/frameworks|*.ts,*.js,*.tsx,*.jsx"

        # LlamaIndex
        'from llama_index|LlamaIndex|ai-ml/frameworks|*.py'
        'import llama_index|LlamaIndex|ai-ml/frameworks|*.py'

        # Google AI / Gemini
        'import google.generativeai|Google AI|ai-ml/llm-apis|*.py'
        'from google.generativeai|Google AI|ai-ml/llm-apis|*.py'
        "import.*from ['\"]@google/generative-ai['\"]|Google AI|ai-ml/llm-apis|*.ts,*.js,*.tsx,*.jsx"

        # Cohere
        'import cohere|Cohere|ai-ml/llm-apis|*.py'
        'from cohere import|Cohere|ai-ml/llm-apis|*.py'

        # Mistral
        'from mistralai|Mistral|ai-ml/llm-apis|*.py'
        'import mistralai|Mistral|ai-ml/llm-apis|*.py'

        # Pinecone
        'from pinecone|Pinecone|ai-ml/vectordb|*.py'
        'import pinecone|Pinecone|ai-ml/vectordb|*.py'
        "import.*from ['\"]@pinecone-database/pinecone['\"]|Pinecone|ai-ml/vectordb|*.ts,*.js,*.tsx,*.jsx"

        # Weaviate
        'import weaviate|Weaviate|ai-ml/vectordb|*.py'
        'from weaviate|Weaviate|ai-ml/vectordb|*.py'

        # ChromaDB
        'import chromadb|ChromaDB|ai-ml/vectordb|*.py'
        'from chromadb|ChromaDB|ai-ml/vectordb|*.py'

        # Qdrant
        'from qdrant_client|Qdrant|ai-ml/vectordb|*.py'
        'import qdrant_client|Qdrant|ai-ml/vectordb|*.py'

        # Hugging Face
        'from transformers|Hugging Face|ai-ml/mlops|*.py'
        'import transformers|Hugging Face|ai-ml/mlops|*.py'
        'from huggingface_hub|Hugging Face|ai-ml/mlops|*.py'

        # Weights & Biases
        'import wandb|Weights & Biases|ai-ml/mlops|*.py'
    )

    for pattern_spec in "${patterns[@]}"; do
        IFS='|' read -r pattern tech category globs <<< "$pattern_spec"

        # Convert glob patterns to find arguments
        local find_args=()
        IFS=',' read -ra glob_array <<< "$globs"
        for g in "${glob_array[@]}"; do
            find_args+=(-name "$g" -o)
        done
        # Remove last -o
        unset 'find_args[${#find_args[@]}-1]'

        # Search for pattern in matching files
        local matches=$(find "$repo_path" -type f \( "${find_args[@]}" \) 2>/dev/null | \
            xargs grep -l -E "$pattern" 2>/dev/null | head -5)

        if [[ -n "$matches" ]]; then
            local file_count=$(echo "$matches" | wc -l | tr -d ' ')
            # Get all file paths relative to repo root
            local file_list=$(echo "$matches" | sed "s|$repo_path/||g" | head -10 | tr '\n' ',' | sed 's/,$//')
            local first_file=$(echo "$matches" | head -1 | sed "s|$repo_path/||")
            local evidence="import found in $file_count file(s): $first_file"

            # Create files array for the finding
            local files_json=$(echo "$matches" | head -10 | sed "s|$repo_path/||g" | jq -R . | jq -s .)

            local finding=$(jq -n \
                --arg name "$tech" \
                --arg category "$category" \
                --arg method "import-pattern" \
                --arg evidence "$evidence" \
                --argjson files "$files_json" \
                --argjson file_count "$file_count" \
                '{name: $name, category: $category, version: "", confidence: 88, detection_method: $method, evidence: [$evidence], files: $files, file_count: $file_count}')

            findings=$(echo "$findings" | jq --argjson f "$finding" '. + [$f]')
        fi
    done

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
            evidence: map(.evidence[]) | unique,
            files: [map(.files // [])[] | .[]] | unique
        }) |
        sort_by(-.confidence)
    ' 2>/dev/null || echo "[]"
}

# Main analysis
analyze_target() {
    local repo_path="$1"
    local provided_sbom="$2"
    local sbom_file=""
    local should_cleanup_sbom=false

    # Use provided SBOM or generate new one
    if [[ -n "$provided_sbom" ]] && [[ -f "$provided_sbom" ]]; then
        echo -e "${BLUE}Using provided SBOM...${NC}" >&2
        sbom_file="$provided_sbom"
    else
        # Generate SBOM
        sbom_file=$(generate_sbom "$repo_path")
        should_cleanup_sbom=true
    fi

    echo -e "${BLUE}Running technology detection...${NC}" >&2

    # Run all detection layers
    local layer1="[]"
    local layer2="[]"
    local layer3="[]"
    local layer4="[]"
    local layer5="[]"

    if [[ -n "$sbom_file" ]] && [[ -f "$sbom_file" ]]; then
        layer1=$(scan_sbom_packages "$sbom_file")
        # Only cleanup if we generated the SBOM ourselves
        [[ "$should_cleanup_sbom" == "true" ]] && rm -f "$sbom_file"
    fi

    layer2=$(scan_config_files "$repo_path")
    layer3=$(scan_env_variables "$repo_path")

    # Layer 4: Docker image scanning (optional - can be slow)
    if [[ "${SCAN_DOCKER_IMAGES:-false}" == "true" ]]; then
        layer4=$(scan_docker_images "$repo_path")
    fi

    # Layer 5: AI import pattern scanning (always enabled for AI adoption tracking)
    layer5=$(scan_ai_imports "$repo_path")

    # Combine all findings
    local all_findings=$(echo "$layer1 $layer2 $layer3 $layer4 $layer5" | jq -s 'add')

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
SCAN_DOCKER_IMAGES=false
while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help) usage ;;
        --local-path)
            LOCAL_PATH="$2"
            shift 2
            ;;
        --sbom)
            SBOM_FILE="$2"
            shift 2
            ;;
        --confidence)
            CONFIDENCE_THRESHOLD="$2"
            shift 2
            ;;
        --scan-docker-images)
            SCAN_DOCKER_IMAGES=true
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
export SCAN_DOCKER_IMAGES

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

# Run analysis (pass SBOM_FILE if provided)
findings=$(analyze_target "$scan_path" "$SBOM_FILE")
final_json=$(generate_output "$findings" "${TARGET:-$scan_path}")

# Output
if [[ -n "$OUTPUT_FILE" ]]; then
    echo "$final_json" > "$OUTPUT_FILE"
    echo -e "${GREEN}✓ Results written to $OUTPUT_FILE${NC}" >&2
else
    echo "$final_json"
fi
