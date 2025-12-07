#!/usr/bin/env bash
#
# Secrets Scanner Data Extractor
# Scans repository for exposed secrets and credentials
# Integrates with TruffleHog when available, falls back to pattern matching
#
# Usage: ./secrets-scanner-data.sh [--local-path <path>] [--repo <owner/repo>]
#
# Output: JSON with detected secrets (redacted), severity, and locations

set -eo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VERSION="1.1.0"

# Source common utilities if available
if [[ -f "$SCRIPT_DIR/../zero/lib/common.sh" ]]; then
    source "$SCRIPT_DIR/../zero/lib/common.sh"
fi

# Default values
LOCAL_PATH=""
REPO=""
ORG=""
OUTPUT_FILE=""
VERBOSE=false
MAX_FILE_SIZE=1048576  # 1MB - skip larger files for pattern matching
MAX_FILES=5000         # Safety limit to prevent scanning huge repos forever
SCAN_TIMEOUT=120       # Timeout in seconds for the entire scan

usage() {
    cat << EOF
Secrets Scanner Data Extractor v${VERSION}

Usage: $(basename "$0") [OPTIONS]

Options:
    --local-path <path>    Path to local repository
    --repo <owner/repo>    GitHub repository (requires gh CLI)
    --org <org>            GitHub org (uses first repo found in zero cache)
    --output <file>        Output file (default: stdout)
    --verbose              Enable verbose output
    --help                 Show this help message

Examples:
    $(basename "$0") --local-path ./my-project
    $(basename "$0") --repo expressjs/express
EOF
    exit 0
}

log() {
    if [[ "$VERBOSE" == "true" ]]; then
        echo "[secrets-scanner] $*" >&2
    fi
}

error() {
    echo "[ERROR] $*" >&2
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --local-path)
            LOCAL_PATH="$2"
            shift 2
            ;;
        --repo)
            REPO="$2"
            shift 2
            ;;
        --org)
            ORG="$2"
            shift 2
            ;;
        --output)
            OUTPUT_FILE="$2"
            shift 2
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        --help)
            usage
            ;;
        *)
            error "Unknown option: $1"
            usage
            ;;
    esac
done

# Determine repository path
REPO_PATH=""
if [[ -n "$LOCAL_PATH" ]]; then
    REPO_PATH="$LOCAL_PATH"
elif [[ -n "$REPO" ]]; then
    # Check zero cache first
    REPO_ORG=$(echo "$REPO" | cut -d'/' -f1)
    REPO_NAME=$(echo "$REPO" | cut -d'/' -f2)
    ZERO_CACHE_PATH="$HOME/.zero/projects/$REPO_ORG/$REPO_NAME/repo"
    GIBSON_PATH="$HOME/.gibson/projects/${REPO_ORG}-${REPO_NAME}/repo"

    if [[ -d "$ZERO_CACHE_PATH" ]]; then
        REPO_PATH="$ZERO_CACHE_PATH"
    elif [[ -d "$GIBSON_PATH" ]]; then
        REPO_PATH="$GIBSON_PATH"
    else
        error "Repository not found. Clone it first or use --local-path"
        exit 1
    fi
elif [[ -n "$ORG" ]]; then
    # Look in zero cache for repos in org
    ORG_PATH="$HOME/.zero/projects/$ORG"
    REPO_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"

    # Colors for output
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    CYAN='\033[0;36m'
    NC='\033[0m'

    if [[ -d "$ORG_PATH" ]]; then
        # Collect repos with and without cloned code
        REPOS_TO_SCAN=()
        REPOS_NOT_CLONED=()
        for repo_dir in "$ORG_PATH"/*/; do
            repo_name=$(basename "$repo_dir")
            if [[ -d "$repo_dir/repo" ]]; then
                REPOS_TO_SCAN+=("$repo_name")
            else
                REPOS_NOT_CLONED+=("$repo_name")
            fi
        done

        # Check if there are uncloned repos and prompt user
        if [[ ${#REPOS_NOT_CLONED[@]} -gt 0 ]]; then
            echo -e "${YELLOW}Found ${#REPOS_NOT_CLONED[@]} repositories without cloned code:${NC}" >&2
            for repo in "${REPOS_NOT_CLONED[@]}"; do
                echo -e "  - $repo" >&2
            done
            echo "" >&2

            # Only prompt if interactive terminal
            if [[ -t 0 ]]; then
                read -p "Would you like to hydrate these repos for analysis? [y/N] " -n 1 -r >&2
                echo "" >&2
            else
                echo -e "${CYAN}Non-interactive mode: skipping uncloned repos${NC}" >&2
                REPLY="n"
            fi

            if [[ $REPLY =~ ^[Yy]$ ]]; then
                echo -e "${BLUE}Hydrating ${#REPOS_NOT_CLONED[@]} repositories...${NC}" >&2
                for repo in "${REPOS_NOT_CLONED[@]}"; do
                    echo -e "${CYAN}Cloning $ORG/$repo...${NC}" >&2
                    "$REPO_ROOT/utils/zero/hydrate.sh" --repo "$ORG/$repo" --quick >&2 2>&1 || true
                    if [[ -d "$ORG_PATH/$repo/repo" ]]; then
                        REPOS_TO_SCAN+=("$repo")
                        echo -e "${GREEN}✓ $repo ready${NC}" >&2
                    else
                        echo -e "${RED}✗ Failed to clone $repo${NC}" >&2
                    fi
                done
                echo "" >&2
            else
                echo -e "${CYAN}Continuing with ${#REPOS_TO_SCAN[@]} already-cloned repositories...${NC}" >&2
            fi
        fi

        if [[ ${#REPOS_TO_SCAN[@]} -eq 0 ]]; then
            error "No repositories with cloned code found in org cache. Hydrate repos first."
            exit 1
        fi

        # Use first available repo (secrets scanner doesn't support multi-repo yet)
        FIRST_REPO="${REPOS_TO_SCAN[0]}"
        REPO_PATH="$ORG_PATH/$FIRST_REPO/repo"
        echo -e "${BLUE}Scanning: $ORG/$FIRST_REPO${NC}" >&2

        if [[ ${#REPOS_TO_SCAN[@]} -gt 1 ]]; then
            echo -e "${YELLOW}Note: secrets-scanner currently analyzes one repo at a time. Run separately for other repos.${NC}" >&2
        fi
    else
        error "Org not found in cache. Hydrate repos first."
        exit 1
    fi
else
    error "Either --local-path, --repo, or --org is required"
    usage
fi

if [[ ! -d "$REPO_PATH" ]]; then
    error "Repository path does not exist: $REPO_PATH"
    exit 1
fi

log "Scanning repository: $REPO_PATH"

# Initialize counters
critical_count=0
high_count=0
medium_count=0
low_count=0

# Store findings in a temp file for building JSON
findings_file=$(mktemp)
echo "[]" > "$findings_file"

# Function to add finding
add_finding() {
    local type="$1"
    local file="$2"
    local line="$3"
    local severity="$4"
    local snippet="$5"
    local detector="$6"

    # Redact the actual secret in snippet - truncate long matches
    local redacted_snippet
    redacted_snippet=$(echo "$snippet" | sed -E 's/([A-Za-z0-9_-]{8})[A-Za-z0-9_+/=-]{8,}/\1********/g' | head -c 200)

    # Escape for JSON
    redacted_snippet=$(echo "$redacted_snippet" | sed 's/\\/\\\\/g; s/"/\\"/g; s/	/\\t/g' | tr '\n' ' ' | tr '\r' ' ')
    local rel_file
    rel_file=$(echo "$file" | sed "s|$REPO_PATH/||")

    # Update counts
    case "$severity" in
        critical) ((critical_count++)) || true ;;
        high) ((high_count++)) || true ;;
        medium) ((medium_count++)) || true ;;
        low) ((low_count++)) || true ;;
    esac

    # Append to findings
    local finding="{\"type\": \"$type\", \"file\": \"$rel_file\", \"line\": $line, \"severity\": \"$severity\", \"snippet\": \"$redacted_snippet\", \"detector\": \"$detector\"}"

    local current
    current=$(cat "$findings_file")
    if [[ "$current" == "[]" ]]; then
        echo "[$finding]" > "$findings_file"
    else
        echo "${current%]}, $finding]" > "$findings_file"
    fi
}

# Get severity for pattern type
get_severity() {
    local pattern_type="$1"
    case "$pattern_type" in
        aws_access_key|aws_secret_key|private_key|pgp_private|gcp_service_account|azure_storage|stripe_key)
            echo "critical"
            ;;
        postgres_url|mysql_url|mongodb_url|redis_url|github_token|github_pat|gitlab_token|anthropic_key|openai_key|npm_token|pypi_token|digitalocean_token|jwt_secret|password_assignment)
            echo "high"
            ;;
        stripe_test_key|slack_token|slack_webhook|sendgrid_key|twilio_sid|twilio_token|heroku_key|bearer_token|api_key_assignment|secret_assignment|docker_auth)
            echo "medium"
            ;;
        mailchimp_key)
            echo "low"
            ;;
        *)
            echo "medium"
            ;;
    esac
}

log "Running pattern-based scan..."

# Define patterns as separate arrays (portable approach)
pattern_names=(
    "aws_access_key"
    "github_token"
    "github_pat"
    "gitlab_token"
    "slack_token"
    "slack_webhook"
    "stripe_key"
    "stripe_test_key"
    "sendgrid_key"
    "gcp_service_account"
    "private_key"
    "postgres_url"
    "mysql_url"
    "mongodb_url"
    "redis_url"
    "anthropic_key"
    "openai_key"
    "npm_token"
    "pypi_token"
    "digitalocean_token"
    "password_assignment"
    "api_key_assignment"
)

pattern_regexes=(
    'AKIA[0-9A-Z]{16}'
    'gh[pousr]_[A-Za-z0-9_]{36,}'
    'github_pat_[A-Za-z0-9_]{22,}'
    'glpat-[A-Za-z0-9-]{20,}'
    'xox[baprs]-[0-9]{10,13}-[0-9]{10,13}[a-zA-Z0-9-]*'
    'https://hooks\.slack\.com/services/T[A-Z0-9]+/B[A-Z0-9]+/[A-Za-z0-9]+'
    'sk_live_[0-9a-zA-Z]{24,}'
    'sk_test_[0-9a-zA-Z]{24,}'
    'SG\.[A-Za-z0-9_-]{22}\.[A-Za-z0-9_-]{43}'
    '"type":[[:space:]]*"service_account"'
    '-----BEGIN (RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----'
    'postgres(ql)?://[^:]+:[^@]+@[^/]+/[^[:space:]]+'
    'mysql://[^:]+:[^@]+@[^/]+/[^[:space:]]+'
    'mongodb(\+srv)?://[^:]+:[^@]+@[^[:space:]]+'
    'redis://[^:]+:[^@]+@[^[:space:]]+'
    'sk-ant-[A-Za-z0-9_-]{90,}'
    'sk-[A-Za-z0-9]{48}'
    'npm_[A-Za-z0-9]{36}'
    'pypi-AgEIcHlwaS5vcmc[A-Za-z0-9_-]+'
    'dop_v1_[a-f0-9]{64}'
    '[pP]assword[[:space:]]*[:=][[:space:]]*["\047][^"\047]{8,}["\047]'
    '[aA]pi[_-]?[kK]ey[[:space:]]*[:=][[:space:]]*["\047][A-Za-z0-9_-]{16,}["\047]'
)

# Create combined pattern file for batch grep (much faster than per-pattern grep)
patterns_file=$(mktemp)
for pattern in "${pattern_regexes[@]}"; do
    echo "$pattern" >> "$patterns_file"
done

# Get list of files to scan (with limit for safety)
log "Finding files to scan..."
files_to_scan=$(find "$REPO_PATH" \
    -type f \
    ! -path "*/.git/*" \
    ! -path "*/node_modules/*" \
    ! -path "*/vendor/*" \
    ! -path "*/__pycache__/*" \
    ! -path "*/venv/*" \
    ! -path "*/.venv/*" \
    ! -path "*/dist/*" \
    ! -path "*/build/*" \
    ! -path "*/.next/*" \
    ! -path "*/coverage/*" \
    ! -path "*/.nyc_output/*" \
    ! -name "*.min.js" \
    ! -name "*.min.css" \
    ! -name "*.map" \
    ! -name "*.png" ! -name "*.jpg" ! -name "*.jpeg" ! -name "*.gif" \
    ! -name "*.ico" ! -name "*.svg" ! -name "*.woff" ! -name "*.woff2" \
    ! -name "*.ttf" ! -name "*.eot" ! -name "*.pdf" \
    ! -name "*.zip" ! -name "*.tar" ! -name "*.gz" \
    ! -name "*.exe" ! -name "*.dll" ! -name "*.so" ! -name "*.dylib" \
    ! -name "package-lock.json" ! -name "yarn.lock" ! -name "pnpm-lock.yaml" \
    2>/dev/null | head -n "$MAX_FILES" || true)

total_files=$(echo "$files_to_scan" | grep -c . || echo "0")
log "Found $total_files files to scan (max: $MAX_FILES)"

scanned_files=0
skipped_files=0
file_count=0

# Identify pattern from matched content
identify_pattern() {
    local content="$1"
    for i in "${!pattern_regexes[@]}"; do
        if echo "$content" | grep -qE "${pattern_regexes[$i]}" 2>/dev/null; then
            echo "${pattern_names[$i]}"
            return
        fi
    done
    echo "unknown"
}

while IFS= read -r file; do
    [[ -z "$file" ]] && continue
    ((file_count++)) || true

    # Progress indicator for large repos
    if [[ $((file_count % 500)) -eq 0 ]]; then
        log "Progress: $file_count/$total_files files..."
    fi

    # Skip files that are too large
    file_size=$(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null || echo "0")
    if [[ "$file_size" -gt "$MAX_FILE_SIZE" ]]; then
        ((skipped_files++)) || true
        continue
    fi

    # Check if file is binary (faster check using head + file)
    if head -c 512 "$file" 2>/dev/null | file - 2>/dev/null | grep -q "binary"; then
        ((skipped_files++)) || true
        continue
    fi

    ((scanned_files++)) || true

    # Check if this is a test/example file (for severity adjustment)
    is_test_file=false
    if echo "$file" | grep -qE '(test|spec|example|sample|mock|fixture)'; then
        is_test_file=true
    fi

    # Skip .env.example files entirely
    if echo "$file" | grep -qE '\.env\.(example|sample|template)'; then
        continue
    fi

    # Use grep -f for batch pattern matching (one grep call per file instead of 22)
    matches=$(grep -n -E -f "$patterns_file" "$file" 2>/dev/null || true)

    while IFS= read -r match; do
        [[ -z "$match" ]] && continue

        line_num=$(echo "$match" | cut -d: -f1)
        content=$(echo "$match" | cut -d: -f2-)

        # Skip common false positives (placeholder values)
        if echo "$content" | grep -qiE '(YOUR_|REPLACE_|EXAMPLE_|xxx|placeholder|dummy|fake|test123|<[^>]+>)'; then
            continue
        fi

        # Identify which pattern matched
        pattern_name=$(identify_pattern "$content")
        severity=$(get_severity "$pattern_name")

        # Downgrade severity for test files
        if [[ "$is_test_file" == "true" ]]; then
            case "$severity" in
                critical) severity="high" ;;
                high) severity="medium" ;;
                medium) severity="low" ;;
            esac
        fi

        add_finding "$pattern_name" "$file" "$line_num" "$severity" "$content" "pattern"
    done <<< "$matches"
done <<< "$files_to_scan"

# Cleanup
rm -f "$patterns_file"

# Check for .env files (which shouldn't be committed)
env_files=$(find "$REPO_PATH" \
    -type f \
    -name ".env" \
    ! -path "*/.git/*" \
    ! -path "*/node_modules/*" \
    2>/dev/null || true)

env_files_count=0
while IFS= read -r env_file; do
    [[ -z "$env_file" ]] && continue
    ((env_files_count++)) || true

    # Check if .env is in .gitignore
    gitignore="$REPO_PATH/.gitignore"
    env_ignored=false
    if [[ -f "$gitignore" ]]; then
        if grep -qE '^\.env$|^\*\.env$' "$gitignore" 2>/dev/null; then
            env_ignored=true
        fi
    fi

    if [[ "$env_ignored" == "false" ]]; then
        add_finding "env_file_committed" "$env_file" 1 "high" ".env file may be committed to repository" "pattern"
    fi
done <<< "$env_files"

# Calculate risk score
total_findings=$((critical_count + high_count + medium_count + low_count))
risk_score=100

if [[ $total_findings -gt 0 ]]; then
    # Deduct points based on severity
    critical_penalty=$((critical_count * 25))
    high_penalty=$((high_count * 15))
    medium_penalty=$((medium_count * 5))
    low_penalty=$((low_count * 2))

    total_penalty=$((critical_penalty + high_penalty + medium_penalty + low_penalty))
    risk_score=$((100 - total_penalty))

    # Clamp to 0-100
    [[ $risk_score -lt 0 ]] && risk_score=0
fi

# Determine risk level
risk_level="excellent"
if [[ $risk_score -lt 40 ]]; then
    risk_level="critical"
elif [[ $risk_score -lt 60 ]]; then
    risk_level="high"
elif [[ $risk_score -lt 80 ]]; then
    risk_level="medium"
elif [[ $risk_score -lt 95 ]]; then
    risk_level="low"
fi

# Build recommendations array
recommendations='["Use environment variables or secret managers for sensitive data","Enable pre-commit hooks to prevent secret commits"]'

if [[ $critical_count -gt 0 ]]; then
    recommendations='["URGENT: Rotate all critical secrets immediately",'${recommendations:1}
fi
if [[ $high_count -gt 0 ]]; then
    recommendations=${recommendations%]}
    recommendations+=',"Review and rotate high-severity secrets"]'
fi
if [[ $env_files_count -gt 0 ]]; then
    recommendations=${recommendations%]}
    recommendations+=',"Ensure .env files are in .gitignore"]'
fi

# Get findings JSON
findings_json=$(cat "$findings_file")
rm -f "$findings_file"

# Generate output JSON
output=$(cat << EOF
{
    "analyzer": "secrets-scanner",
    "version": "$VERSION",
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "repository": "$REPO_PATH",
    "scanner": {
        "pattern_matching": true,
        "patterns_count": ${#pattern_names[@]},
        "max_files": $MAX_FILES
    },
    "summary": {
        "risk_score": $risk_score,
        "risk_level": "$risk_level",
        "total_findings": $total_findings,
        "critical_count": $critical_count,
        "high_count": $high_count,
        "medium_count": $medium_count,
        "low_count": $low_count,
        "files_scanned": $scanned_files,
        "files_skipped": $skipped_files,
        "env_files_found": $env_files_count
    },
    "findings": $findings_json,
    "recommendations": $recommendations
}
EOF
)

# Output results
if [[ -n "$OUTPUT_FILE" ]]; then
    echo "$output" > "$OUTPUT_FILE"
    log "Results written to $OUTPUT_FILE"
else
    echo "$output"
fi
