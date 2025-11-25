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
source "$SCRIPT_DIR/lib/pattern-loader.sh"

# Load .env file if it exists in repository root
if [[ -f "$REPO_ROOT/.env" ]]; then
    set -a  # automatically export all variables
    source "$REPO_ROOT/.env"
    set +a  # stop automatically exporting
fi

# RAG patterns directory
RAG_ROOT="$REPO_ROOT/rag/technology-identification"

# Load RAG patterns at startup
echo -e "${BLUE}Loading technology patterns...${NC}" >&2
if load_all_patterns "$RAG_ROOT" 2>&1 | grep -v "^$" | grep -v "^load_" >&2; then
    echo -e "${GREEN}✓ Patterns loaded${NC}" >&2
else
    echo -e "${YELLOW}⚠ Pattern loading had warnings${NC}" >&2
fi

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

    echo -e "${BLUE}Cloning repository...${NC}" >&2
    if git clone --depth 1 "$repo_url" "$TEMP_DIR" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Repository cloned${NC}" >&2
        return 0
    else
        echo -e "${RED}✗ Failed to clone repository${NC}" >&2
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
        echo -e "${GREEN}Using shared SBOM file${NC}" >&2
        echo "$SBOM_FILE"
        return 0
    fi

    # Generate new SBOM
    local sbom_file=$(mktemp)
    echo -e "${BLUE}Generating SBOM...${NC}" >&2

    if generate_sbom "$target_dir" "$sbom_file" "true" 2>&1 | grep -v "^$" >&2; then
        if [[ -f "$sbom_file" ]]; then
            echo -e "${GREEN}✓ SBOM generated${NC}" >&2
            echo "$sbom_file"
            return 0
        fi
    fi

    echo -e "${RED}✗ SBOM generation failed${NC}" >&2
    return 1
}

# Layer 1: Scan SBOM for package dependencies
scan_sbom_packages() {
    local sbom_file="$1"
    local findings=()

    # Extract components from SBOM
    local components=$(jq -c '.components[]?' "$sbom_file" 2>/dev/null)

    if [[ -z "$components" ]]; then
        echo "[]"
        return
    fi

    while IFS= read -r component; do
        local name=$(echo "$component" | jq -r '.name // ""' 2>/dev/null || echo "")
        local version=$(echo "$component" | jq -r '.version // ""' 2>/dev/null || echo "")
        local purl=$(echo "$component" | jq -r '.purl // ""' 2>/dev/null || echo "")

        # Extract ecosystem from purl
        local ecosystem=""
        if [[ -n "$purl" ]]; then
            ecosystem=$(echo "$purl" | sed -n 's/^pkg:\([^/]*\).*/\1/p')
        fi

        # Try to match package using RAG patterns
        local match_result=$(match_package_name "$name" 2>/dev/null)

        local tech_category=""
        local tech_name=""
        local confidence=95

        if [[ -n "$match_result" ]]; then
            # Extract info from RAG pattern match
            tech_name=$(echo "$match_result" | jq -r '.technology // ""' 2>/dev/null)
            tech_category=$(echo "$match_result" | jq -r '.category // ""' 2>/dev/null)
            confidence=$(echo "$match_result" | jq -r '.confidence // 95' 2>/dev/null)

            # If we got a match, continue to create finding
            if [[ -z "$tech_name" ]] || [[ -z "$tech_category" ]]; then
                continue
            fi
        else
            # No RAG match - skip this package
            continue
        fi

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
        printf '%s\n' "${findings[@]}" | jq -s '.' 2>/dev/null || echo "[]"
    fi
}

# Layer 1b: Scan manifest files with osv-scanner (supplement to SBOM)
scan_manifest_files() {
    local repo_path="$1"
    local findings=()

    # Check if osv-scanner is available
    if ! command -v osv-scanner &> /dev/null; then
        echo "[]"
        return
    fi

    # Run osv-scanner to detect packages from manifest files
    # osv-scanner returns exit 1 when vulnerabilities found, so use || true
    # Also, it may have progress messages before JSON, so extract JSON portion
    local temp_output=$(mktemp)
    osv-scanner --recursive "$repo_path" --format=json > "$temp_output" 2>&1 || true

    # Extract JSON from output (skip progress messages)
    local osv_output=$(grep -A 999999 "^{" "$temp_output" 2>/dev/null || echo '{"results":[]}')
    rm -f "$temp_output"

    # Extract packages from osv-scanner output
    # Structure: .results[].packages[].package
    local packages=$(echo "$osv_output" | jq -r '.results[]?.packages[]?.package | select(. != null) | @json' 2>/dev/null)

    if [[ -z "$packages" ]]; then
        echo "[]"
        return
    fi

    while IFS= read -r pkg_json; do
        local name=$(echo "$pkg_json" | jq -r '.name // ""' 2>/dev/null)
        local version=$(echo "$pkg_json" | jq -r '.version // ""' 2>/dev/null)
        local ecosystem=$(echo "$pkg_json" | jq -r '.ecosystem // ""' 2>/dev/null)

        [[ -z "$name" ]] && continue

        # Map package names to technology categories (same logic as scan_sbom_packages)
        local tech_category=""
        local tech_name=""
        local confidence=95

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

            # Developer Tools - Containers
            docker|docker-*) tech_category="developer-tools/containers"; tech_name="Docker" ;;
            kubernetes|kubectl|k8s-*) tech_category="developer-tools/containers"; tech_name="Kubernetes" ;;

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

            # Databases
            pg|postgres|postgresql) tech_category="databases/relational"; tech_name="PostgreSQL" ;;
            mysql|mysql2) tech_category="databases/relational"; tech_name="MySQL" ;;
            mongodb|mongoose) tech_category="databases/nosql"; tech_name="MongoDB" ;;
            redis) tech_category="databases/keyvalue"; tech_name="Redis" ;;

            # Cryptographic Libraries
            openssl|pyopenssl) tech_category="cryptographic-libraries/tls"; tech_name="OpenSSL" ;;
            cryptography) tech_category="cryptographic-libraries/general"; tech_name="cryptography" ;;
            pycrypto|pycryptodome) tech_category="cryptographic-libraries/general"; tech_name="PyCrypto" ;;
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
                --arg method "manifest-file" \
                --arg evidence "Detected in $ecosystem manifest: $name@$version" \
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
    done <<< "$packages"

    # Return findings as JSON array
    if [[ ${#findings[@]} -eq 0 ]]; then
        echo "[]"
    else
        printf '%s\n' "${findings[@]}" | jq -s '.' 2>/dev/null || echo "[]"
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
        printf '%s\n' "${findings[@]}" | jq -s '.' 2>/dev/null || echo "[]"
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
        printf '%s\n' "${findings[@]}" | jq -s '.' 2>/dev/null || echo "[]"
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
        printf '%s\n' "${findings[@]}" | jq -s '.' 2>/dev/null || echo "[]"
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
        printf '%s\n' "${findings[@]}" | jq -s '.' 2>/dev/null || echo "[]"
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
        <(echo "$layer5") 2>/dev/null)

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
    ' 2>/dev/null || echo "[]"
}

# Analyze repository or SBOM
analyze_target() {
    local target="$1"
    local repo_path=""
    local sbom_file=""

    # Determine target type and prepare
    if [[ -n "$LOCAL_PATH" ]]; then
        if [[ ! -d "$LOCAL_PATH" ]]; then
            echo -e "${RED}Error: Local path does not exist: $LOCAL_PATH${NC}" >&2
            return 1
        fi
        echo -e "${GREEN}Target: Local directory${NC}" >&2
        repo_path="$LOCAL_PATH"
    elif is_sbom_file "$target"; then
        echo -e "${GREEN}Target: SBOM file${NC}" >&2
        sbom_file="$target"
        # For SBOM-only analysis, we can only do Layer 1
        echo "" >&2
        echo -e "${BLUE}Analyzing SBOM for technologies...${NC}" >&2
        local layer1=$(scan_sbom_packages "$sbom_file")
        local results=$(aggregate_findings "$layer1" "[]" "[]" "[]" "[]")
        echo "$results"
        return 0
    elif is_git_url "$target"; then
        echo -e "${GREEN}Target: Git repository${NC}" >&2
        if clone_repository "$target"; then
            repo_path="$TEMP_DIR"
        else
            return 1
        fi
    elif [[ -d "$target" ]]; then
        echo -e "${GREEN}Target: Local directory${NC}" >&2
        repo_path="$target"
    else
        echo -e "${RED}Error: Invalid target${NC}" >&2
        return 1
    fi

    # Generate or use existing SBOM
    echo "" >&2
    sbom_file=$(get_or_generate_sbom "$repo_path")
    if [[ $? -ne 0 ]]; then
        return 1
    fi

    # Run all detection layers
    echo "" >&2
    echo -e "${BLUE}Running multi-layer technology detection...${NC}" >&2

    echo -e "${CYAN}Layer 1a: Scanning SBOM packages...${NC}" >&2
    local layer1a=$(scan_sbom_packages "$sbom_file")
    local layer1a_count=$(echo "$layer1a" | jq 'length' 2>/dev/null || echo "0")
    echo -e "${CYAN}  Layer 1a found: $layer1a_count technologies${NC}" >&2

    echo -e "${CYAN}Layer 1b: Scanning manifest files (osv-scanner)...${NC}" >&2
    local layer1b=$(scan_manifest_files "$repo_path")
    local layer1b_count=$(echo "$layer1b" | jq 'length' 2>/dev/null || echo "0")
    echo -e "${CYAN}  Layer 1b found: $layer1b_count technologies${NC}" >&2

    # Merge layer 1a and 1b results
    local layer1=$(echo "$layer1a" "$layer1b" | jq -s 'add' 2>/dev/null || echo "[]")
    local layer1_count=$(echo "$layer1" | jq 'length' 2>/dev/null || echo "0")
    echo -e "${CYAN}  Layer 1 merged: $layer1_count technologies${NC}" >&2

    echo -e "${CYAN}Layer 2: Scanning configuration files...${NC}" >&2
    local layer2=$(scan_config_files "$repo_path")

    echo -e "${CYAN}Layer 3: Scanning import statements...${NC}" >&2
    local layer3=$(scan_imports "$repo_path")

    echo -e "${CYAN}Layer 4: Scanning API endpoints...${NC}" >&2
    local layer4=$(scan_api_endpoints "$repo_path")

    echo -e "${CYAN}Layer 5: Scanning environment variables...${NC}" >&2
    local layer5=$(scan_env_variables "$repo_path")

    # Aggregate and deduplicate findings
    echo "" >&2
    echo -e "${BLUE}Aggregating findings...${NC}" >&2
    local results=$(aggregate_findings "$layer1" "$layer2" "$layer3" "$layer4" "$layer5")

    # Filter by confidence threshold
    results=$(echo "$results" | jq --argjson threshold "$CONFIDENCE_THRESHOLD" 'map(select(.confidence >= $threshold))' 2>/dev/null || echo "[]")

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
**Total Technologies**: $(echo "$findings" | jq 'length' 2>/dev/null || echo "0")

## Executive Summary

This report identifies technologies detected across multiple layers of analysis, including package dependencies, configuration files, import statements, API endpoints, and environment variables.

## Technologies Detected

EOF

    # Group by category
    local categories=$(echo "$findings" | jq -r '[.[].category] | unique | .[]' 2>/dev/null | sort)

    while IFS= read -r category; do
        if [[ -z "$category" ]]; then
            continue
        fi

        echo ""
        # Format category name: replace / with space and capitalize each word
        local formatted_category=$(echo "$category" | tr '/' ' ' | awk '{for(i=1;i<=NF;i++) $i=toupper(substr($i,1,1)) tolower(substr($i,2))}1')
        echo "### $formatted_category"
        echo ""

        echo "$findings" | jq -r --arg cat "$category" '
            map(select(.category == $cat)) |
            .[] |
            "#### \(.name)\(.version | if length > 0 then " v\(.)" else "" end)
- **Confidence**: \(.confidence)%
- **Detection Methods**: \(.detection_methods | join(", "))
- **Evidence**: \(.evidence | map("  - \(.)") | join("\n"))
"
        ' 2>/dev/null
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
    ' 2>/dev/null

    echo ""
    echo "## Confidence Distribution"
    echo ""

    local high=$(echo "$findings" | jq '[.[] | select(.confidence >= 80)] | length' 2>/dev/null || echo "0")
    local medium=$(echo "$findings" | jq '[.[] | select(.confidence >= 60 and .confidence < 80)] | length' 2>/dev/null || echo "0")
    local low=$(echo "$findings" | jq '[.[] | select(.confidence < 60)] | length' 2>/dev/null || echo "0")

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
        }' 2>/dev/null || echo '{"error": "Failed to generate JSON report"}'
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

    echo "$response" | jq -r '.content[0].text // empty' 2>/dev/null
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

        # Check if it's already a full URL
        if [[ "$repo_name" =~ ^https?:// ]] || [[ "$repo_name" =~ ^git@ ]]; then
            TARGET="$repo_name"
        else
            # Convert owner/repo to full GitHub URL
            TARGET="https://github.com/$repo_name"
        fi

        MULTI_REPO_MODE=false
        echo -e "${CYAN}Using repository: $TARGET${NC}"
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

# Multi-repo mode
if [[ "$MULTI_REPO_MODE" == true ]]; then
    echo -e "${CYAN}Multi-repository mode enabled${NC}"
    echo ""

    # Expand targets (convert org:* to individual repos)
    declare -a REPO_URLS=()

    for target_spec in "${TARGETS_LIST[@]}"; do
        if [[ "$target_spec" == org:* ]]; then
            # Extract organization name
            org_name="${target_spec#org:}"

            # Strip GitHub URL if provided (support both https://github.com/org and https://github.com/org/)
            org_name="${org_name#https://github.com/}"
            org_name="${org_name#http://github.com/}"
            org_name="${org_name%/}"  # Remove trailing slash

            echo -e "${CYAN}Fetching repositories for organization: $org_name${NC}" >&2

            # Get list of repos from organization
            repo_list=$(list_org_repos "$org_name")

            if [[ -z "$repo_list" ]]; then
                echo -e "${YELLOW}Warning: No repositories found for organization $org_name${NC}" >&2
                continue
            fi

            # Convert each repo to a GitHub URL
            while IFS= read -r repo; do
                [[ -z "$repo" ]] && continue
                REPO_URLS+=("https://github.com/$repo")
            done <<< "$repo_list"

            echo -e "${GREEN}Found $(echo "$repo_list" | wc -l | tr -d ' ') repositories${NC}" >&2

        elif [[ "$target_spec" == repo:* ]]; then
            # Extract repo name
            repo_name="${target_spec#repo:}"

            # Check if it's already a full URL
            if [[ "$repo_name" =~ ^https?:// ]] || [[ "$repo_name" =~ ^git@ ]]; then
                REPO_URLS+=("$repo_name")
            else
                # Convert owner/repo to full GitHub URL
                REPO_URLS+=("https://github.com/$repo_name")
            fi
        fi
    done

    # Check if we have any repos to process
    if [[ ${#REPO_URLS[@]} -eq 0 ]]; then
        echo -e "${RED}Error: No repositories to analyze${NC}" >&2
        exit 1
    fi

    echo ""
    echo -e "${GREEN}Processing ${#REPO_URLS[@]} repositories...${NC}" >&2
    echo ""

    # Process each repository
    declare -a all_findings=()
    declare -a all_repo_names=()
    declare -A cumulative_tech_repos=()  # Map technology -> list of repos where found
    repo_count=0
    success_count=0

    for repo_url in "${REPO_URLS[@]}"; do
        repo_count=$((repo_count + 1))
        repo_name=$(echo "$repo_url" | sed 's|https://github.com/||')

        echo -e "${BLUE}[$repo_count/${#REPO_URLS[@]}] Analyzing: $repo_url${NC}" >&2
        echo ""

        # Analyze this repository (capture JSON on stdout, let status go to stderr)
        findings=$(analyze_target "$repo_url")
        analyze_status=$?

        if [[ $analyze_status -eq 0 ]] && [[ -n "$findings" ]]; then
            # Validate it's proper JSON
            if echo "$findings" | jq '.' >/dev/null 2>&1; then
                all_findings+=("$findings")
                all_repo_names+=("$repo_name")
                success_count=$((success_count + 1))

                # Extract technologies from this repo and add to cumulative map with repo info
                repo_techs=$(echo "$findings" | jq -r '.[].name' 2>/dev/null | sort -u)
                # Get short repo name (just the repo part, not org/repo)
                short_repo_name=$(basename "$repo_name")
                while IFS= read -r tech; do
                    [[ -z "$tech" ]] && continue
                    # Add repo to the technology's repo list
                    if [[ -z "${cumulative_tech_repos[$tech]}" ]]; then
                        cumulative_tech_repos[$tech]="$short_repo_name"
                    else
                        # Append repo if not already in list
                        if [[ ! "${cumulative_tech_repos[$tech]}" =~ (^|, )$short_repo_name(,|$) ]]; then
                            cumulative_tech_repos[$tech]="${cumulative_tech_repos[$tech]}, $short_repo_name"
                        fi
                    fi
                done <<< "$repo_techs"

                # Print summary for this repository
                echo "" >&2
                echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}" >&2
                echo -e "${GREEN}  Summary: $repo_name${NC}" >&2
                echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}" >&2

                # Extract and display key technologies
                tech_count=$(echo "$findings" | jq 'length' 2>/dev/null || echo "0")
                echo -e "  ${CYAN}Technologies detected in this repo: $tech_count${NC}" >&2

                # Show top technologies by confidence
                if [[ "$tech_count" != "0" ]] && [[ "$tech_count" != "null" ]] && [[ "$tech_count" -gt 0 ]]; then
                    echo "" >&2
                    echo -e "  ${CYAN}Technologies found:${NC}" >&2
                    echo "$findings" | jq -r 'sort_by(-.confidence) | .[] | "    • \(.name) (\(.category)) - \(.confidence)%"' 2>/dev/null >&2
                fi

                # Show cumulative technology list across all repos scanned so far
                echo "" >&2
                echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}" >&2
                echo -e "${YELLOW}  Cumulative Technologies (${#cumulative_tech_repos[@]} unique across $success_count repos)${NC}" >&2
                echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}" >&2

                # Display cumulative technologies sorted alphabetically with repo info
                for tech in $(printf '%s\n' "${!cumulative_tech_repos[@]}" | sort); do
                    echo -e "    ${NC}• $tech ${CYAN}[${cumulative_tech_repos[$tech]}]${NC}" >&2
                done

                echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}" >&2
                echo "" >&2
            else
                echo -e "${YELLOW}⚠ Analysis returned invalid JSON${NC}" >&2
            fi
        else
            echo -e "${YELLOW}⚠ Analysis failed or returned no data${NC}" >&2
        fi

        echo ""
    done

    # Generate combined report
    echo ""
    echo -e "${CYAN}Generating combined report...${NC}" >&2
    echo ""

    case "$OUTPUT_FORMAT" in
        json)
            # Create JSON array of all findings
            echo "["
            first=true
            for finding in "${all_findings[@]}"; do
                if [[ "$first" == true ]]; then
                    first=false
                else
                    echo ","
                fi
                echo "$finding"
            done
            echo "]"
            ;;
        markdown)
            # Create combined markdown report
            echo "# Multi-Repository Technology Analysis"
            echo ""
            echo "**Date**: $(date '+%Y-%m-%d %H:%M:%S')"
            echo "**Repositories Analyzed**: $success_count / ${#REPO_URLS[@]}"
            echo ""
            echo "---"
            echo ""

            repo_index=0
            for finding in "${all_findings[@]}"; do
                repo_index=$((repo_index + 1))
                echo "## Repository $repo_index: ${REPO_URLS[$((repo_index - 1))]}"
                echo ""
                echo "$finding"
                echo ""
                echo "---"
                echo ""
            done
            ;;
        *)
            echo -e "${RED}Error: Unknown format: $OUTPUT_FORMAT${NC}" >&2
            exit 1
            ;;
    esac
fi

echo ""
echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}  Analysis Complete${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""
