#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Technology Identification Scanner
# Detects technologies, frameworks, and services used in a codebase
#
# This scanner uses the unified scanner-ux library for consistent UX.
#
# Usage: ./tech-discovery.sh [options] <target>
# Output: JSON with detected technologies, categories, and confidence scores
#############################################################################

set -e

# Get script directory and load libraries
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SCANNERS_ROOT="$(dirname "$SCRIPT_DIR")"
UTILS_ROOT="$(dirname "$SCANNERS_ROOT")"
REPO_ROOT="$(dirname "$UTILS_ROOT")"

# Load unified scanner UX library
source "$UTILS_ROOT/lib/scanner-ux.sh"
source "$UTILS_ROOT/lib/scanner-reports.sh"

# Load scanner-specific libraries
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
OUTPUT_FORMAT="json"
VERBOSE=false
QUIET=false
AGENT="nikon"  # Lord Nikon - architecture/patterns, fits tech-discovery

usage() {
    cat << EOF
Technology Identification Scanner

Detects technologies, frameworks, and services used in a codebase.

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
    --format FORMAT         Output format: json, markdown, terminal, html (default: json)
    --agent AGENT           Enable agent personality (nikon, cereal, etc.)
    -o, --output FILE       Write output to file (default: stdout)
    -k, --keep-clone        Keep cloned repository
    -v, --verbose           Verbose output
    -q, --quiet             Suppress progress output
    -h, --help              Show this help

DETECTION LAYERS:
    1. SBOM packages (via syft)
    2. Config files (Dockerfile, package.json, etc.)
    3. Environment variables (.env files)
    4. Docker image scanning (optional, --scan-docker-images)
    5. AI import patterns (scans source files for AI library imports)

EXAMPLES:
    $0 https://github.com/expressjs/express
    $0 --local-path ~/.zero/repos/foo/repo
    $0 --format markdown -o report.md /path/to/project

EOF
    exit 0
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
        --format|-f)
            OUTPUT_FORMAT="$2"
            shift 2
            ;;
        --agent|-a)
            AGENT="$2"
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
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -q|--quiet)
            QUIET=true
            shift
            ;;
        -*)
            scanner_error "Unknown option: $1"
            exit 1
            ;;
        *)
            TARGET="$1"
            shift
            ;;
    esac
done
export SCAN_DOCKER_IMAGES

# Initialize scanner with agent personality
scanner_init "tech-discovery" "2.0.0" \
    $(if $VERBOSE; then echo "--verbose"; fi) \
    $(if $QUIET; then echo "--quiet"; fi) \
    --agent "$AGENT"

# Check dependencies
scanner_require "jq" "brew install jq" || exit 1
scanner_require "syft" "brew install syft" || exit 1

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

    scanner_step "Cloning repository"
    if git clone --depth 1 "$repo_url" "$TEMP_DIR" > /dev/null 2>&1; then
        scanner_step "Cloning repository" "success"
        return 0
    else
        scanner_step "Cloning repository" "error"
        scanner_error "Failed to clone repository"
        exit 1
    fi
}

# Generate SBOM if needed
generate_sbom_file() {
    local target_dir="$1"
    local sbom_file=$(mktemp)

    scanner_step "Generating SBOM"
    if syft scan "$target_dir" -o cyclonedx-json > "$sbom_file" 2>/dev/null; then
        scanner_step "Generating SBOM" "success"
        echo "$sbom_file"
        return 0
    else
        scanner_step "Generating SBOM" "warning"
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

    local components=$(jq -c '.components[]?' "$sbom_file" 2>/dev/null)
    [[ -z "$components" ]] && { echo "[]"; return; }

    local total=$(echo "$components" | wc -l | tr -d ' ')
    scanner_progress_start "Scanning packages" "$total"

    local count=0
    while IFS= read -r component; do
        ((count++))
        local name=$(echo "$component" | jq -r '.name // ""')
        local version=$(echo "$component" | jq -r '.version // ""')

        scanner_progress_update "$count" "$name"

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

    scanner_progress_end
    echo "$findings"
}

# Layer 2: Scan for configuration files
scan_config_files() {
    local repo_path="$1"
    local findings="[]"

    scanner_step "Scanning configuration files"

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

    scanner_step "Scanning configuration files" "success"
    echo "$findings"
}

# Layer 3: Scan environment variables
scan_env_variables() {
    local repo_path="$1"
    local findings="[]"

    local env_files=$(find "$repo_path" -maxdepth 3 -type f \( -name ".env*" -o -name "*.env" \) 2>/dev/null)
    [[ -z "$env_files" ]] && { echo "[]"; return; }

    scanner_step "Scanning environment variables"

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

    scanner_step "Scanning environment variables" "success"
    echo "$findings"
}

# Layer 4: Scan Docker images for packages
scan_docker_images() {
    local repo_path="$1"
    local findings="[]"

    local dockerfiles=$(find "$repo_path" -maxdepth 3 -type f -name "Dockerfile*" 2>/dev/null)
    [[ -z "$dockerfiles" ]] && { echo "[]"; return; }

    scanner_step "Scanning Docker images"

    local images=""
    for df in $dockerfiles; do
        local img=$(grep -E "^FROM " "$df" 2>/dev/null | head -1 | awk '{print $2}' | sed 's/:.*$//')
        [[ -n "$img" ]] && images="$images $img"
    done

    if [[ -f "$repo_path/docker-compose.yml" ]] || [[ -f "$repo_path/docker-compose.yaml" ]]; then
        local compose_file="$repo_path/docker-compose.yml"
        [[ -f "$repo_path/docker-compose.yaml" ]] && compose_file="$repo_path/docker-compose.yaml"
        local compose_images=$(grep -E "^\s+image:" "$compose_file" 2>/dev/null | sed 's/.*image:\s*//' | sed 's/:.*$//' | tr -d '"' | tr -d "'")
        images="$images $compose_images"
    fi

    images=$(echo "$images" | tr ' ' '\n' | sort -u | grep -v '^$')
    [[ -z "$images" ]] && { scanner_step "Scanning Docker images" "warning"; echo "[]"; return; }

    for image in $images; do
        [[ "$image" =~ ^(alpine|ubuntu|debian|busybox|scratch)$ ]] && continue

        local temp_sbom=$(mktemp)
        if timeout 60 syft scan "registry:$image:latest" -o cyclonedx-json > "$temp_sbom" 2>/dev/null; then
            scanner_debug "Scanned Docker image: $image"
            local image_findings=$(scan_sbom_packages "$temp_sbom")
            image_findings=$(echo "$image_findings" | jq --arg img "$image" '
                map(. + {
                    detection_method: "docker-image-sbom",
                    evidence: (.evidence + ["from Docker image: " + $img])
                })
            ')
            findings=$(echo "$findings $image_findings" | jq -s 'add')
        else
            scanner_debug "Could not scan Docker image: $image"
        fi
        rm -f "$temp_sbom"
    done

    scanner_step "Scanning Docker images" "success"
    echo "$findings"
}

# Layer 5: Scan source files for AI import patterns
scan_ai_imports() {
    local repo_path="$1"
    local findings="[]"

    scanner_step "Scanning AI import patterns"

    local patterns=(
        'import.*openai|OpenAI|ai-ml/llm-apis|*.py'
        'from openai import|OpenAI|ai-ml/llm-apis|*.py'
        "import.*from ['\"]openai['\"]|OpenAI|ai-ml/llm-apis|*.ts,*.js,*.tsx,*.jsx"
        'import anthropic|Anthropic|ai-ml/llm-apis|*.py'
        'from anthropic import|Anthropic|ai-ml/llm-apis|*.py'
        "import.*from ['\"]@anthropic-ai/sdk['\"]|Anthropic|ai-ml/llm-apis|*.ts,*.js,*.tsx,*.jsx"
        'from langchain|LangChain|ai-ml/frameworks|*.py'
        'import langchain|LangChain|ai-ml/frameworks|*.py'
        'from llama_index|LlamaIndex|ai-ml/frameworks|*.py'
        'import google.generativeai|Google AI|ai-ml/llm-apis|*.py'
        'import cohere|Cohere|ai-ml/llm-apis|*.py'
        'from mistralai|Mistral|ai-ml/llm-apis|*.py'
        'from pinecone|Pinecone|ai-ml/vectordb|*.py'
        'import weaviate|Weaviate|ai-ml/vectordb|*.py'
        'import chromadb|ChromaDB|ai-ml/vectordb|*.py'
        'from qdrant_client|Qdrant|ai-ml/vectordb|*.py'
        'from transformers|Hugging Face|ai-ml/mlops|*.py'
        'import wandb|Weights & Biases|ai-ml/mlops|*.py'
    )

    for pattern_spec in "${patterns[@]}"; do
        IFS='|' read -r pattern tech category globs <<< "$pattern_spec"

        local find_args=()
        IFS=',' read -ra glob_array <<< "$globs"
        for g in "${glob_array[@]}"; do
            find_args+=(-name "$g" -o)
        done
        unset 'find_args[${#find_args[@]}-1]'

        local matches=$(find "$repo_path" -type f \( "${find_args[@]}" \) 2>/dev/null | \
            xargs grep -l -E "$pattern" 2>/dev/null | head -5)

        if [[ -n "$matches" ]]; then
            local file_count=$(echo "$matches" | wc -l | tr -d ' ')
            local first_file=$(echo "$matches" | head -1 | sed "s|$repo_path/||")
            local evidence="import found in $file_count file(s): $first_file"
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

    scanner_step "Scanning AI import patterns" "success"
    echo "$findings"
}

# Aggregate and deduplicate findings
aggregate_findings() {
    local all_findings="$1"

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

    if [[ -n "$provided_sbom" ]] && [[ -f "$provided_sbom" ]]; then
        scanner_info "Using provided SBOM"
        sbom_file="$provided_sbom"
    else
        sbom_file=$(generate_sbom_file "$repo_path")
        should_cleanup_sbom=true
    fi

    # Run all detection layers
    local layer1="[]"
    local layer2="[]"
    local layer3="[]"
    local layer4="[]"
    local layer5="[]"

    if [[ -n "$sbom_file" ]] && [[ -f "$sbom_file" ]]; then
        layer1=$(scan_sbom_packages "$sbom_file")
        [[ "$should_cleanup_sbom" == "true" ]] && rm -f "$sbom_file"
    fi

    layer2=$(scan_config_files "$repo_path")
    layer3=$(scan_env_variables "$repo_path")

    if [[ "${SCAN_DOCKER_IMAGES:-false}" == "true" ]]; then
        layer4=$(scan_docker_images "$repo_path")
    fi

    layer5=$(scan_ai_imports "$repo_path")

    local all_findings=$(echo "$layer1 $layer2 $layer3 $layer4 $layer5" | jq -s 'add')
    local results=$(aggregate_findings "$all_findings")
    results=$(echo "$results" | jq --argjson threshold "$CONFIDENCE_THRESHOLD" 'map(select(.confidence >= $threshold))')

    echo "$results"
}

# Generate final output
generate_output() {
    local findings="$1"
    local target="$2"

    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    local total=$(echo "$findings" | jq 'length')
    local by_category=$(echo "$findings" | jq 'group_by(.category) | map({key: .[0].category, value: length}) | from_entries')
    local high=$(echo "$findings" | jq '[.[] | select(.confidence >= 80)] | length')
    local medium=$(echo "$findings" | jq '[.[] | select(.confidence >= 60 and .confidence < 80)] | length')
    local low=$(echo "$findings" | jq '[.[] | select(.confidence < 60)] | length')

    jq -n \
        --arg ts "$timestamp" \
        --arg tgt "$target" \
        --arg ver "2.0.0" \
        --argjson total "$total" \
        --argjson by_cat "$by_category" \
        --argjson hi "$high" \
        --argjson med "$medium" \
        --argjson lo "$low" \
        --argjson techs "$findings" \
        '{
            analyzer: "tech-discovery",
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

#############################################################################
# MAIN EXECUTION
#############################################################################

# Load RAG patterns if available
if [[ -d "$RAG_ROOT" ]] && type load_all_patterns &>/dev/null; then
    scanner_debug "Loading technology patterns from RAG"
    load_all_patterns "$RAG_ROOT" 2>/dev/null || true
fi

scan_path=""
repo_name=""

if [[ -n "$LOCAL_PATH" ]]; then
    [[ ! -d "$LOCAL_PATH" ]] && { scanner_error "Local path does not exist"; exit 1; }
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
        scanner_header "$TARGET"
        findings=$(scan_sbom_packages "$TARGET")
        final_json=$(generate_output "$findings" "$TARGET")

        if [[ "$OUTPUT_FORMAT" == "json" ]]; then
            if [[ -n "$OUTPUT_FILE" ]]; then
                echo "$final_json" > "$OUTPUT_FILE"
                scanner_success "Results written to $OUTPUT_FILE"
            else
                echo "$final_json"
            fi
        else
            report_init "tech-discovery" "2.0.0" "$TARGET"
            report_add_section "Technologies" "$findings"
            if [[ -n "$OUTPUT_FILE" ]]; then
                report_save "$OUTPUT_FILE" "$OUTPUT_FORMAT"
            else
                report_generate "$OUTPUT_FORMAT"
            fi
        fi
        scanner_footer "success"
        exit 0
    else
        scanner_error "Invalid target - must be URL, directory, or SBOM file"
        exit 1
    fi
else
    scanner_error "No target specified"
    usage
fi

# Display header
scanner_header "$repo_name"

# Run analysis
findings=$(analyze_target "$scan_path" "$SBOM_FILE")
final_json=$(generate_output "$findings" "${TARGET:-$scan_path}")

# Display summary
scanner_summary_start "Results"
scanner_summary_metric "Technologies found" "$(echo "$final_json" | jq -r '.summary.total')" "info"
scanner_summary_metric "High confidence" "$(echo "$final_json" | jq -r '.summary.confidence_distribution.high')" "good"
scanner_summary_metric "Medium confidence" "$(echo "$final_json" | jq -r '.summary.confidence_distribution.medium')" "warning"
scanner_summary_metric "Low confidence" "$(echo "$final_json" | jq -r '.summary.confidence_distribution.low')" "info"
scanner_summary_end

# Output results
if [[ "$OUTPUT_FORMAT" == "json" ]]; then
    if [[ -n "$OUTPUT_FILE" ]]; then
        echo "$final_json" > "$OUTPUT_FILE"
        scanner_success "Results written to $OUTPUT_FILE"
    else
        echo "$final_json"
    fi
else
    # Use report library for other formats
    report_init "tech-discovery" "2.0.0" "$repo_name"
    report_set_metadata_json "summary" "$(echo "$final_json" | jq '.summary')"
    report_add_section "Technologies Detected" "$(echo "$final_json" | jq '.technologies')"

    if [[ -n "$OUTPUT_FILE" ]]; then
        report_save "$OUTPUT_FILE" "$OUTPUT_FORMAT"
    else
        report_generate "$OUTPUT_FORMAT"
    fi
fi

# Pass finding count for easter eggs
total_found=$(echo "$final_json" | jq -r '.summary.total // 0')
scanner_footer "success" "$total_found"
