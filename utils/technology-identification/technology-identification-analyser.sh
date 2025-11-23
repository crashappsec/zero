#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Technology Identification Analyser Script
# Analyzes SBOMs and repositories to identify technologies, frameworks, and tools
# Detects business tools, developer tools, languages, and cloud services
# Usage: ./technology-identification-analyser.sh [options] <target>
#############################################################################

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
UTILS_ROOT="$(dirname "$SCRIPT_DIR")"
REPO_ROOT="$(dirname "$UTILS_ROOT")"

# Load global libraries
source "$UTILS_ROOT/lib/sbom.sh"
source "$UTILS_ROOT/lib/github.sh"

# Default options
OUTPUT_FORMAT="markdown"
TEMP_DIR=""
LOCAL_PATH=""
SBOM_FILE=""
CLEANUP=true
ANTHROPIC_API_KEY="${ANTHROPIC_API_KEY:-}"
# Claude enabled by default if API key is set
USE_CLAUDE=false
if [[ -n "$ANTHROPIC_API_KEY" ]]; then
    USE_CLAUDE=true
fi
MULTI_REPO_MODE=false
TARGETS_LIST=()
OUTPUT_FILE=""
CONFIDENCE_THRESHOLD=50

# Function to print usage
usage() {
    cat << EOF
Technology Identification Analyser - Detect technologies in repositories

Usage: $0 [OPTIONS] [target]

TARGET:
    SBOM file path          Analyze an existing SBOM (JSON/XML)
    Git repository URL      Clone and analyze repository
    Local directory path    Analyze local repository
    (If no target specified, uses shared environment variables)

MULTI-REPO OPTIONS:
    --org ORG_NAME          Scan all repos in GitHub organization
    --repo OWNER/REPO       Scan specific repository

ANALYSIS OPTIONS:
    --claude                Use Claude AI for enhanced analysis (requires ANTHROPIC_API_KEY)
    --local-path PATH       Use pre-cloned repository at PATH (skips cloning)
    --sbom-file FILE        Use existing SBOM file (skips generation)
    --confidence N          Minimum confidence threshold (0-100, default: 50)
    -f, --format FORMAT     Output format: json|markdown|table (default: markdown)
    -o, --output FILE       Write results to file
    -k, --keep-clone        Keep cloned repository (don't cleanup)
    -h, --help              Show this help message

EXAMPLES:
    # Analyze an SBOM file
    $0 /path/to/sbom.json

    # Analyze repository
    $0 https://github.com/org/repo

    # Analyze with Claude AI
    $0 --claude https://github.com/org/repo

    # Scan entire GitHub organization
    $0 --org myorg

    # Use pre-cloned repository
    $0 --local-path /path/to/repo

EOF
    exit 1
}

# Function to check if syft is installed
check_syft() {
    if ! command -v syft &> /dev/null; then
        echo -e "${RED}Error: syft is not installed${NC}"
        echo "Install: brew install syft"
        exit 1
    fi
}

# Function to detect if target is a Git URL
is_git_url() {
    [[ "$1" =~ ^(https?|git)://.*\.git$ ]] || [[ "$1" =~ ^git@.*:.*\.git$ ]] || [[ "$1" =~ github\.com|gitlab\.com|bitbucket\.org ]]
}

# Function to detect if target is an SBOM file
is_sbom_file() {
    local file="$1"
    [[ -f "$file" ]] && ([[ "$file" =~ \.json$ ]] || [[ "$file" =~ \.xml$ ]] || [[ "$file" =~ \.cdx\. ]] || [[ "$file" =~ bom\. ]])
}

# Function to clone repository
clone_repository() {
    local repo_url="$1"
    TEMP_DIR=$(mktemp -d)

    echo -e "${BLUE}Cloning repository...${NC}"
    if git clone --depth 1 "$repo_url" "$TEMP_DIR" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Repository cloned${NC}"
        return 0
    else
        echo -e "${RED}✗ Failed to clone repository${NC}"
        return 1
    fi
}

# Function to cleanup temporary files
cleanup() {
    if [[ "$CLEANUP" == true ]] && [[ -n "$TEMP_DIR" ]] && [[ -d "$TEMP_DIR" ]]; then
        rm -rf "$TEMP_DIR"
    fi
}

# Ensure cleanup on script exit
trap cleanup EXIT

# Function to generate or locate SBOM
get_or_generate_sbom() {
    local target_dir="$1"

    # If SBOM_FILE is already set (from shared env), use it
    if [[ -n "$SBOM_FILE" ]] && [[ -f "$SBOM_FILE" ]]; then
        echo -e "${GREEN}Using shared SBOM file${NC}"
        echo "$SBOM_FILE"
        return 0
    fi

    # Generate new SBOM
    local sbom_file=$(mktemp)
    echo -e "${BLUE}Generating SBOM...${NC}"

    if generate_sbom "$target_dir" "$sbom_file" "true" 2>&1 | grep -v "^$"; then
        if [[ -f "$sbom_file" ]]; then
            echo -e "${GREEN}✓ SBOM generated${NC}"
            echo "$sbom_file"
            return 0
        fi
    fi

    echo -e "${RED}✗ SBOM generation failed${NC}"
    return 1
}

# Layer 1: Scan SBOM for package dependencies
scan_sbom_packages() {
    local sbom_file="$1"
    local findings=()

    # Extract components from SBOM
    local components=$(jq -r '.components[]? | @json' "$sbom_file" 2>/dev/null)

    if [[ -z "$components" ]]; then
        echo "[]"
        return
    fi

    while IFS= read -r component; do
        local name=$(echo "$component" | jq -r '.name // ""')
        local version=$(echo "$component" | jq -r '.version // ""')
        local purl=$(echo "$component" | jq -r '.purl // ""')

        # Extract ecosystem from purl
        local ecosystem=""
        if [[ -n "$purl" ]]; then
            ecosystem=$(echo "$purl" | sed -n 's/^pkg:\([^/]*\).*/\1/p')
        fi

        # Map package names to technology categories
        local tech_category=""
        local tech_name=""
        local confidence=95

        # Use pattern matching that works in bash case statements
        case "$name" in
            # Business Tools - Payment
            stripe) tech_category="business-tools/payment"; tech_name="Stripe" ;;
            paypal|paypal-*) tech_category="business-tools/payment"; tech_name="PayPal" ;;
            square) tech_category="business-tools/payment"; tech_name="Square" ;;
            braintree) tech_category="business-tools/payment"; tech_name="Braintree" ;;

            # Business Tools - Communication
            twilio) tech_category="business-tools/communication"; tech_name="Twilio" ;;
            sendgrid) tech_category="business-tools/communication"; tech_name="SendGrid" ;;
            mailgun) tech_category="business-tools/communication"; tech_name="Mailgun" ;;

            # Developer Tools - Infrastructure
            terraform|terraform-*) tech_category="developer-tools/infrastructure"; tech_name="Terraform" ;;
            ansible|ansible-*) tech_category="developer-tools/infrastructure"; tech_name="Ansible" ;;
            pulumi|pulumi-*) tech_category="developer-tools/infrastructure"; tech_name="Pulumi" ;;

            # Developer Tools - Containers
            docker|docker-*) tech_category="developer-tools/containers"; tech_name="Docker" ;;
            kubernetes|kubernetes-*|k8s|k8s-*) tech_category="developer-tools/containers"; tech_name="Kubernetes" ;;

            # Web Frameworks - Frontend
            react|react-dom) tech_category="web-frameworks/frontend"; tech_name="React" ;;
            vue) tech_category="web-frameworks/frontend"; tech_name="Vue.js" ;;
            angular|angular-*|@angular/*) tech_category="web-frameworks/frontend"; tech_name="Angular" ;;
            svelte) tech_category="web-frameworks/frontend"; tech_name="Svelte" ;;
            next) tech_category="web-frameworks/frontend"; tech_name="Next.js" ;;

            # Web Frameworks - Backend
            express) tech_category="web-frameworks/backend"; tech_name="Express" ;;
            fastapi) tech_category="web-frameworks/backend"; tech_name="FastAPI" ;;
            django) tech_category="web-frameworks/backend"; tech_name="Django" ;;
            flask) tech_category="web-frameworks/backend"; tech_name="Flask" ;;
            rails) tech_category="web-frameworks/backend"; tech_name="Ruby on Rails" ;;

            # Databases - Relational
            pg|postgres|postgresql) tech_category="databases/relational"; tech_name="PostgreSQL" ;;
            mysql|mysql2) tech_category="databases/relational"; tech_name="MySQL" ;;
            sqlite|sqlite3|sqlite-*) tech_category="databases/relational"; tech_name="SQLite" ;;

            # Databases - NoSQL
            mongodb|mongoose) tech_category="databases/nosql"; tech_name="MongoDB" ;;
            redis) tech_category="databases/keyvalue"; tech_name="Redis" ;;

            # Cloud - AWS
            aws-sdk|aws-sdk-*|@aws-sdk/*|boto3|botocore) tech_category="cloud-providers/aws"; tech_name="AWS SDK" ;;

            # Cloud - GCP
            @google-cloud/*|google-cloud-*) tech_category="cloud-providers/gcp"; tech_name="Google Cloud SDK" ;;

            # Cloud - Azure
            @azure/*|azure-*) tech_category="cloud-providers/azure"; tech_name="Azure SDK" ;;

            # Cryptographic Libraries
            openssl|openssl-*) tech_category="cryptographic-libraries/tls"; tech_name="OpenSSL" ;;
            jsonwebtoken|pyjwt) tech_category="cryptographic-libraries/jwt"; tech_name="JWT Library" ;;
            bcrypt|bcrypt-*|bcryptjs) tech_category="cryptographic-libraries/hashing"; tech_name="bcrypt" ;;

            # Message Queues
            amqp|amqp-*|rabbitmq|rabbitmq-*) tech_category="message-queues"; tech_name="RabbitMQ" ;;
            kafka|kafka-*|kafkajs) tech_category="message-queues"; tech_name="Apache Kafka" ;;

            *)
                # Unknown package, skip
                continue
                ;;
        esac

        if [[ -n "$tech_category" ]]; then
            local finding=$(jq -n \
                --arg name "$tech_name" \
                --arg category "$tech_category" \
                --arg version "$version" \
                --argjson confidence "$confidence" \
                --arg method "sbom-package" \
                --arg evidence "package.json dependency: $name@$version" \
                '{
                    name: $name,
                    category: $category,
                    version: $version,
                    confidence: $confidence,
                    detection_method: $method,
                    evidence: [$evidence]
                }')
            findings+=("$finding")
        fi
    done <<< "$components"

    # Return findings as JSON array
    if [[ ${#findings[@]} -eq 0 ]]; then
        echo "[]"
    else
        printf '%s\n' "${findings[@]}" | jq -s '.'
    fi
}

# Layer 2: Scan for configuration files
scan_config_files() {
    local repo_path="$1"
    local findings=()

    # Dockerfile detection
    if [[ -f "$repo_path/Dockerfile" ]] || find "$repo_path" -name "Dockerfile*" -type f 2>/dev/null | grep -q .; then
        local finding=$(jq -n \
            --arg name "Docker" \
            --arg category "developer-tools/containers" \
            --argjson confidence 90 \
            --arg method "config-file" \
            --arg evidence "Dockerfile found" \
            '{
                name: $name,
                category: $category,
                confidence: $confidence,
                detection_method: $method,
                evidence: [$evidence]
            }')
        findings+=("$finding")
    fi

    # docker-compose.yml detection
    if [[ -f "$repo_path/docker-compose.yml" ]] || [[ -f "$repo_path/docker-compose.yaml" ]]; then
        local finding=$(jq -n \
            --arg name "Docker Compose" \
            --arg category "developer-tools/containers" \
            --argjson confidence 90 \
            --arg method "config-file" \
            --arg evidence "docker-compose.yml found" \
            '{
                name: $name,
                category: $category,
                confidence: $confidence,
                detection_method: $method,
                evidence: [$evidence]
            }')
        findings+=("$finding")
    fi

    # Terraform detection
    if find "$repo_path" -name "*.tf" -type f 2>/dev/null | grep -q .; then
        local version=""
        local tf_file=$(find "$repo_path" -name "*.tf" -type f 2>/dev/null | head -1)
        if [[ -n "$tf_file" ]]; then
            version=$(grep -oP 'required_version\s*=\s*"\K[^"]+' "$tf_file" 2>/dev/null | head -1)
        fi

        local finding=$(jq -n \
            --arg name "Terraform" \
            --arg category "developer-tools/infrastructure" \
            --arg version "$version" \
            --argjson confidence 90 \
            --arg method "config-file" \
            --arg evidence "*.tf files found" \
            '{
                name: $name,
                category: $category,
                version: $version,
                confidence: $confidence,
                detection_method: $method,
                evidence: [$evidence]
            }')
        findings+=("$finding")
    fi

    # Kubernetes detection
    if find "$repo_path" -name "*.yaml" -o -name "*.yml" -type f 2>/dev/null | xargs grep -l "kind:" 2>/dev/null | grep -q .; then
        local finding=$(jq -n \
            --arg name "Kubernetes" \
            --arg category "developer-tools/containers" \
            --argjson confidence 85 \
            --arg method "config-file" \
            --arg evidence "Kubernetes manifests found" \
            '{
                name: $name,
                category: $category,
                confidence: $confidence,
                detection_method: $method,
                evidence: [$evidence]
            }')
        findings+=("$finding")
    fi

    # GitHub Actions detection
    if [[ -d "$repo_path/.github/workflows" ]]; then
        local finding=$(jq -n \
            --arg name "GitHub Actions" \
            --arg category "developer-tools/cicd" \
            --argjson confidence 95 \
            --arg method "config-file" \
            --arg evidence ".github/workflows directory found" \
            '{
                name: $name,
                category: $category,
                confidence: $confidence,
                detection_method: $method,
                evidence: [$evidence]
            }')
        findings+=("$finding")
    fi

    # GitLab CI detection
    if [[ -f "$repo_path/.gitlab-ci.yml" ]]; then
        local finding=$(jq -n \
            --arg name "GitLab CI" \
            --arg category "developer-tools/cicd" \
            --argjson confidence 95 \
            --arg method "config-file" \
            --arg evidence ".gitlab-ci.yml found" \
            '{
                name: $name,
                category: $category,
                confidence: $confidence,
                detection_method: $method,
                evidence: [$evidence]
            }')
        findings+=("$finding")
    fi

    # Return findings as JSON array
    if [[ ${#findings[@]} -eq 0 ]]; then
        echo "[]"
    else
        printf '%s\n' "${findings[@]}" | jq -s '.'
    fi
}

# Layer 3: Scan for import statements
scan_imports() {
    local repo_path="$1"
    local findings=()

    # Search for common import patterns in source files
    # Limit search to avoid scanning large node_modules, etc.
    local search_opts="-type f"
    search_opts="$search_opts -not -path '*/node_modules/*'"
    search_opts="$search_opts -not -path '*/vendor/*'"
    search_opts="$search_opts -not -path '*/.git/*'"
    search_opts="$search_opts -not -path '*/dist/*'"
    search_opts="$search_opts -not -path '*/build/*'"

    # JavaScript/TypeScript imports
    local js_imports=$(find "$repo_path" $search_opts \( -name "*.js" -o -name "*.ts" -o -name "*.jsx" -o -name "*.tsx" \) -exec grep -h "import.*from\|require(" {} \; 2>/dev/null | head -100)

    # Check for AWS imports
    if echo "$js_imports" | grep -q "aws-sdk\|@aws-sdk"; then
        local finding=$(jq -n \
            --arg name "AWS SDK" \
            --arg category "cloud-providers/aws" \
            --argjson confidence 80 \
            --arg method "import-statement" \
            --arg evidence "AWS SDK import found in source files" \
            '{
                name: $name,
                category: $category,
                confidence: $confidence,
                detection_method: $method,
                evidence: [$evidence]
            }')
        findings+=("$finding")
    fi

    # Check for Stripe imports
    if echo "$js_imports" | grep -q "stripe"; then
        local finding=$(jq -n \
            --arg name "Stripe" \
            --arg category "business-tools/payment" \
            --argjson confidence 80 \
            --arg method "import-statement" \
            --arg evidence "Stripe import found in source files" \
            '{
                name: $name,
                category: $category,
                confidence: $confidence,
                detection_method: $method,
                evidence: [$evidence]
            }')
        findings+=("$finding")
    fi

    # Python imports
    local py_imports=$(find "$repo_path" $search_opts -name "*.py" -exec grep -h "^import\|^from.*import" {} \; 2>/dev/null | head -100)

    # Check for boto3 (AWS Python SDK)
    if echo "$py_imports" | grep -q "boto3\|botocore"; then
        local finding=$(jq -n \
            --arg name "AWS SDK (boto3)" \
            --arg category "cloud-providers/aws" \
            --argjson confidence 80 \
            --arg method "import-statement" \
            --arg evidence "boto3 import found in Python files" \
            '{
                name: $name,
                category: $category,
                confidence: $confidence,
                detection_method: $method,
                evidence: [$evidence]
            }')
        findings+=("$finding")
    fi

    # Check for Django
    if echo "$py_imports" | grep -q "django"; then
        local finding=$(jq -n \
            --arg name "Django" \
            --arg category "web-frameworks/backend" \
            --argjson confidence 85 \
            --arg method "import-statement" \
            --arg evidence "Django import found in Python files" \
            '{
                name: $name,
                category: $category,
                confidence: $confidence,
                detection_method: $method,
                evidence: [$evidence]
            }')
        findings+=("$finding")
    fi

    # Return findings as JSON array
    if [[ ${#findings[@]} -eq 0 ]]; then
        echo "[]"
    else
        printf '%s\n' "${findings[@]}" | jq -s '.'
    fi
}

# Layer 4: Scan for API endpoint patterns
scan_api_endpoints() {
    local repo_path="$1"
    local findings=()

    # Search for API endpoint URLs in source files
    local search_opts="-type f"
    search_opts="$search_opts -not -path '*/node_modules/*'"
    search_opts="$search_opts -not -path '*/vendor/*'"
    search_opts="$search_opts -not -path '*/.git/*'"

    # Stripe API
    if find "$repo_path" $search_opts -exec grep -q "api\.stripe\.com" {} \; 2>/dev/null; then
        local finding=$(jq -n \
            --arg name "Stripe" \
            --arg category "business-tools/payment" \
            --argjson confidence 75 \
            --arg method "api-endpoint" \
            --arg evidence "api.stripe.com endpoint found" \
            '{
                name: $name,
                category: $category,
                confidence: $confidence,
                detection_method: $method,
                evidence: [$evidence]
            }')
        findings+=("$finding")
    fi

    # AWS S3
    if find "$repo_path" $search_opts -exec grep -q "s3\.amazonaws\.com\|s3://\|amazonaws\.com" {} \; 2>/dev/null; then
        local finding=$(jq -n \
            --arg name "AWS S3" \
            --arg category "cloud-providers/aws" \
            --argjson confidence 70 \
            --arg method "api-endpoint" \
            --arg evidence "AWS S3 endpoint found" \
            '{
                name: $name,
                category: $category,
                confidence: $confidence,
                detection_method: $method,
                evidence: [$evidence]
            }')
        findings+=("$finding")
    fi

    # Twilio API
    if find "$repo_path" $search_opts -exec grep -q "api\.twilio\.com" {} \; 2>/dev/null; then
        local finding=$(jq -n \
            --arg name "Twilio" \
            --arg category "business-tools/communication" \
            --argjson confidence 75 \
            --arg method "api-endpoint" \
            --arg evidence "api.twilio.com endpoint found" \
            '{
                name: $name,
                category: $category,
                confidence: $confidence,
                detection_method: $method,
                evidence: [$evidence]
            }')
        findings+=("$finding")
    fi

    # Return findings as JSON array
    if [[ ${#findings[@]} -eq 0 ]]; then
        echo "[]"
    else
        printf '%s\n' "${findings[@]}" | jq -s '.'
    fi
}

# Layer 5: Scan for environment variable patterns
scan_env_variables() {
    local repo_path="$1"
    local findings=()

    # Search .env, .env.example, .env.template files
    local env_files=$(find "$repo_path" -maxdepth 3 -type f \( -name ".env*" -o -name "*.env" \) 2>/dev/null)

    if [[ -z "$env_files" ]]; then
        echo "[]"
        return
    fi

    local env_content=$(cat $env_files 2>/dev/null)

    # Stripe
    if echo "$env_content" | grep -q "STRIPE"; then
        local finding=$(jq -n \
            --arg name "Stripe" \
            --arg category "business-tools/payment" \
            --argjson confidence 65 \
            --arg method "env-variable" \
            --arg evidence "STRIPE_* environment variables found" \
            '{
                name: $name,
                category: $category,
                confidence: $confidence,
                detection_method: $method,
                evidence: [$evidence]
            }')
        findings+=("$finding")
    fi

    # AWS
    if echo "$env_content" | grep -q "AWS_"; then
        local finding=$(jq -n \
            --arg name "AWS" \
            --arg category "cloud-providers/aws" \
            --argjson confidence 65 \
            --arg method "env-variable" \
            --arg evidence "AWS_* environment variables found" \
            '{
                name: $name,
                category: $category,
                confidence: $confidence,
                detection_method: $method,
                evidence: [$evidence]
            }')
        findings+=("$finding")
    fi

    # Database
    if echo "$env_content" | grep -q "DATABASE_URL\|DB_"; then
        local db_type="PostgreSQL"
        if echo "$env_content" | grep -q "postgres://\|postgresql://"; then
            db_type="PostgreSQL"
        elif echo "$env_content" | grep -q "mysql://"; then
            db_type="MySQL"
        elif echo "$env_content" | grep -q "mongodb://\|mongo://"; then
            db_type="MongoDB"
        fi

        local finding=$(jq -n \
            --arg name "$db_type" \
            --arg category "databases" \
            --argjson confidence 60 \
            --arg method "env-variable" \
            --arg evidence "Database environment variables found" \
            '{
                name: $name,
                category: $category,
                confidence: $confidence,
                detection_method: $method,
                evidence: [$evidence]
            }')
        findings+=("$finding")
    fi

    # Return findings as JSON array
    if [[ ${#findings[@]} -eq 0 ]]; then
        echo "[]"
    else
        printf '%s\n' "${findings[@]}" | jq -s '.'
    fi
}

# Aggregate findings from all layers
aggregate_findings() {
    local layer1="$1"
    local layer2="$2"
    local layer3="$3"
    local layer4="$4"
    local layer5="$5"

    # Combine all findings - flatten all arrays into one
    local all_findings=$(jq -s 'map(select(. != null)) | add' \
        <(echo "$layer1") \
        <(echo "$layer2") \
        <(echo "$layer3") \
        <(echo "$layer4") \
        <(echo "$layer5"))

    # Handle case where all_findings might be null
    if [[ "$all_findings" == "null" ]] || [[ -z "$all_findings" ]]; then
        all_findings="[]"
    fi

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
    '
}

# Analyze repository or SBOM
analyze_target() {
    local target="$1"
    local repo_path=""
    local sbom_file=""

    # Determine target type and prepare
    if [[ -n "$LOCAL_PATH" ]]; then
        if [[ ! -d "$LOCAL_PATH" ]]; then
            echo -e "${RED}Error: Local path does not exist: $LOCAL_PATH${NC}"
            return 1
        fi
        echo -e "${GREEN}Target: Local directory${NC}"
        repo_path="$LOCAL_PATH"
    elif is_sbom_file "$target"; then
        echo -e "${GREEN}Target: SBOM file${NC}"
        sbom_file="$target"
        # For SBOM-only analysis, we can only do Layer 1
        echo ""
        echo -e "${BLUE}Analyzing SBOM for technologies...${NC}"
        local layer1=$(scan_sbom_packages "$sbom_file")
        local results=$(aggregate_findings "$layer1" "[]" "[]" "[]" "[]")
        echo "$results"
        return 0
    elif is_git_url "$target"; then
        echo -e "${GREEN}Target: Git repository${NC}"
        if clone_repository "$target"; then
            repo_path="$TEMP_DIR"
        else
            return 1
        fi
    elif [[ -d "$target" ]]; then
        echo -e "${GREEN}Target: Local directory${NC}"
        repo_path="$target"
    else
        echo -e "${RED}Error: Invalid target${NC}"
        return 1
    fi

    # Generate or use existing SBOM
    echo ""
    sbom_file=$(get_or_generate_sbom "$repo_path")
    if [[ $? -ne 0 ]]; then
        return 1
    fi

    # Run all detection layers
    echo ""
    echo -e "${BLUE}Running multi-layer technology detection...${NC}"

    echo -e "${CYAN}Layer 1: Scanning SBOM packages...${NC}"
    local layer1=$(scan_sbom_packages "$sbom_file")

    echo -e "${CYAN}Layer 2: Scanning configuration files...${NC}"
    local layer2=$(scan_config_files "$repo_path")

    echo -e "${CYAN}Layer 3: Scanning import statements...${NC}"
    local layer3=$(scan_imports "$repo_path")

    echo -e "${CYAN}Layer 4: Scanning API endpoints...${NC}"
    local layer4=$(scan_api_endpoints "$repo_path")

    echo -e "${CYAN}Layer 5: Scanning environment variables...${NC}"
    local layer5=$(scan_env_variables "$repo_path")

    # Aggregate and deduplicate findings
    echo ""
    echo -e "${BLUE}Aggregating findings...${NC}"
    local results=$(aggregate_findings "$layer1" "$layer2" "$layer3" "$layer4" "$layer5")

    # Filter by confidence threshold
    results=$(echo "$results" | jq --argjson threshold "$CONFIDENCE_THRESHOLD" 'map(select(.confidence >= $threshold))')

    echo "$results"
}

# Generate markdown report
generate_markdown_report() {
    local findings="$1"
    local target="${2:-unknown}"

    cat << EOF
# Technology Stack Analysis Report

**Repository**: $target
**Scan Date**: $(date -u +%Y-%m-%dT%H:%M:%SZ)
**Total Technologies**: $(echo "$findings" | jq 'length')

## Executive Summary

This report identifies technologies detected across multiple layers of analysis, including package dependencies, configuration files, import statements, API endpoints, and environment variables.

## Technologies Detected

EOF

    # Group by category
    local categories=$(echo "$findings" | jq -r '[.[].category] | unique | .[]' | sort)

    while IFS= read -r category; do
        if [[ -z "$category" ]]; then
            continue
        fi

        echo ""
        echo "### $(echo "$category" | tr '/' ' ' | sed 's/\b\(.\)/\u\1/g')"
        echo ""

        echo "$findings" | jq -r --arg cat "$category" '
            map(select(.category == $cat)) |
            .[] |
            "#### \(.name)\(.version | if length > 0 then " v\(.)" else "" end)
- **Confidence**: \(.confidence)%
- **Detection Methods**: \(.detection_methods | join(", "))
- **Evidence**: \(.evidence | map("  - \(.)") | join("\n"))
"
        '
    done <<< "$categories"

    echo ""
    echo "## Summary by Category"
    echo ""
    echo "| Category | Count |"
    echo "|----------|-------|"

    echo "$findings" | jq -r '
        group_by(.category) |
        map({category: .[0].category, count: length}) |
        .[] |
        "| \(.category) | \(.count) |"
    '

    echo ""
    echo "## Confidence Distribution"
    echo ""

    local high=$(echo "$findings" | jq '[.[] | select(.confidence >= 80)] | length')
    local medium=$(echo "$findings" | jq '[.[] | select(.confidence >= 60 and .confidence < 80)] | length')
    local low=$(echo "$findings" | jq '[.[] | select(.confidence < 60)] | length')

    echo "- **High Confidence (80-100%)**: $high technologies"
    echo "- **Medium Confidence (60-79%)**: $medium technologies"
    echo "- **Low Confidence (<60%)**: $low technologies"
}

# Generate JSON report
generate_json_report() {
    local findings="$1"
    local target="${2:-unknown}"

    # Ensure findings is valid JSON, default to empty array if not
    if ! echo "$findings" | jq empty 2>/dev/null; then
        findings="[]"
    fi

    jq -n \
        --arg timestamp "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        --arg target "$target" \
        --argjson technologies "$findings" \
        '{
            scan_metadata: {
                timestamp: $timestamp,
                repository: $target,
                analyser_version: "1.0.0"
            },
            summary: {
                total_technologies: ($technologies | length),
                by_category: ($technologies | group_by(.category) | map({key: .[0].category, value: length}) | from_entries),
                confidence_distribution: {
                    high: ($technologies | map(select(.confidence >= 80)) | length),
                    medium: ($technologies | map(select(.confidence >= 60 and .confidence < 80)) | length),
                    low: ($technologies | map(select(.confidence < 60)) | length)
                }
            },
            technologies: $technologies
        }'
}

# Claude AI Analysis
analyze_with_claude() {
    local data="$1"
    local model="claude-sonnet-4-20250514"

    if [[ -z "$ANTHROPIC_API_KEY" ]]; then
        echo -e "${RED}Error: ANTHROPIC_API_KEY required for --claude mode${NC}" >&2
        exit 1
    fi

    echo -e "${BLUE}Analyzing with Claude AI...${NC}" >&2

    local prompt="You are a technology stack analyst. Analyze this technology detection data and provide insights.

# Analysis Requirements

For the detected technologies, provide:
1. **Technology Stack Overview** - Categorize and summarize the technology choices
2. **Architecture Patterns** - Identify architectural patterns (microservices, monolith, serverless, etc.)
3. **Technology Maturity Assessment** - Flag deprecated, end-of-life, or outdated technologies
4. **Security Considerations** - Identify technologies with known security concerns
5. **Consolidation Opportunities** - Identify redundant or overlapping technologies
6. **Recommendations** - Suggest improvements or modernization opportunities

# Output Format

## Technology Stack Overview
- Summary of detected technologies by category
- Key technology choices and their purposes

## Architecture Assessment
- Identified architectural patterns
- Technology alignment with modern practices

## Risk Assessment
- Deprecated or end-of-life technologies requiring attention
- Security considerations for detected technologies

## Recommendations
- High-priority actions (critical updates, security fixes)
- Medium-priority improvements (modernization, consolidation)
- Long-term strategic suggestions

# Detection Data:
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

    echo "$response" | jq -r '.content[0].text // empty'
}

# Load cost tracking if using Claude
if [[ "$USE_CLAUDE" == "true" ]]; then
    if [ -f "$UTILS_ROOT/lib/claude-cost.sh" ]; then
        source "$UTILS_ROOT/lib/claude-cost.sh"
        init_cost_tracking
    fi
fi

# Parse command line arguments
TARGET=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --local-path)
            LOCAL_PATH="$2"
            CLEANUP=false
            shift 2
            ;;
        --sbom-file)
            SBOM_FILE="$2"
            shift 2
            ;;
        -f|--format)
            OUTPUT_FORMAT="$2"
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
        --org)
            TARGETS_LIST+=("org:$2")
            MULTI_REPO_MODE=true
            shift 2
            ;;
        --repo)
            TARGETS_LIST+=("repo:$2")
            MULTI_REPO_MODE=true
            shift 2
            ;;
        --claude)
            USE_CLAUDE=true
            shift
            ;;
        --no-claude)
            USE_CLAUDE=false
            shift
            ;;
        --confidence)
            CONFIDENCE_THRESHOLD="$2"
            shift 2
            ;;
        -h|--help)
            usage
            ;;
        *)
            TARGET="$1"
            shift
            ;;
    esac
done

# Main script
echo ""
echo "========================================="
echo "  Technology Identification Analyser"
echo "========================================="
echo ""

# Check Claude AI status
if [[ "$USE_CLAUDE" == "true" ]] && [[ -n "$ANTHROPIC_API_KEY" ]]; then
    echo -e "${GREEN}🤖 Claude AI: ENABLED (analyzing results with AI)${NC}"
    echo ""
elif [[ -z "$ANTHROPIC_API_KEY" ]]; then
    echo -e "${YELLOW}ℹ️  Claude AI: DISABLED (no API key found)${NC}"
    echo -e "${CYAN}   Set ANTHROPIC_API_KEY to enable AI-enhanced analysis${NC}"
    echo ""
    USE_CLAUDE=false
fi

# Check prerequisites
check_syft

# Handle --repo argument by converting to TARGET
if [[ "$MULTI_REPO_MODE" == true ]] && [[ ${#TARGETS_LIST[@]} -eq 1 ]]; then
    target_spec="${TARGETS_LIST[0]}"
    if [[ "$target_spec" == repo:* ]]; then
        # Extract repo name (e.g., "repo:owner/repo" -> "owner/repo")
        repo_name="${target_spec#repo:}"
        TARGET="https://github.com/$repo_name"
        MULTI_REPO_MODE=false
        echo -e "${CYAN}Converted --repo $repo_name to $TARGET${NC}"
        echo ""
    fi
fi

# Single target mode
if [[ "$MULTI_REPO_MODE" == false ]]; then
    if [[ -z "$TARGET" ]]; then
        echo -e "${RED}Error: No target specified${NC}"
        echo "Use --repo, --org, or provide a target path/URL"
        usage
    fi

    # Analyze target
    findings=$(analyze_target "$TARGET")

    # Generate report
    case "$OUTPUT_FORMAT" in
        json)
            output=$(generate_json_report "$findings" "$TARGET")
            ;;
        markdown)
            output=$(generate_markdown_report "$findings" "$TARGET")
            ;;
        *)
            echo -e "${RED}Error: Unknown format: $OUTPUT_FORMAT${NC}"
            exit 1
            ;;
    esac

    # Output results
    if [[ -n "$OUTPUT_FILE" ]]; then
        echo "$output" > "$OUTPUT_FILE"
        echo ""
        echo -e "${GREEN}Report saved to: $OUTPUT_FILE${NC}"
    else
        echo ""
        echo "$output"
    fi

    # Claude AI analysis if enabled
    if [[ "$USE_CLAUDE" == "true" ]]; then
        echo ""
        echo "========================================="
        echo "  Claude AI Enhanced Analysis"
        echo "========================================="
        echo ""

        claude_analysis=$(analyze_with_claude "$output")
        echo "$claude_analysis"

        if command -v display_api_cost_summary &> /dev/null; then
            echo ""
            display_api_cost_summary
        fi
    fi
fi

echo ""
echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}  Analysis Complete${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""
