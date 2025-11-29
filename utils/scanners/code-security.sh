#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Code Security - Data Collector
# Collects security-relevant data using static analysis tools
# Outputs JSON for agent analysis - NO AI calls
#
# Usage: ./code-security-data.sh [options] <target>
# Output: JSON with file inventory, potential issues, and metadata
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

# Default options
OUTPUT_FILE=""
LOCAL_PATH=""
TEMP_DIR=""
CLEANUP=true
TARGET=""
MAX_FILES=500
EXCLUDE_PATTERNS="node_modules/**,vendor/**,.git/**,*.min.js,*.bundle.js,dist/**,build/**"

usage() {
    cat << EOF
Code Security - Data Collector (JSON output for agent analysis)

Usage: $0 [OPTIONS] <target>

TARGET:
    Git repository URL      Clone and analyze repository
    Local directory path    Analyze local repository

OPTIONS:
    --local-path PATH       Use pre-cloned repository (skips cloning)
    --max-files N           Maximum files to scan (default: 500)
    --exclude PATTERNS      Glob patterns to exclude (comma-separated)
    -o, --output FILE       Write JSON to file (default: stdout)
    -k, --keep-clone        Keep cloned repository
    -h, --help              Show this help

OUTPUT:
    JSON object with:
    - metadata: scan timestamp, target, file counts
    - file_inventory: list of source files with types
    - potential_issues: pattern-matched security concerns
    - secrets_scan: potential secret exposures

EXAMPLES:
    $0 https://github.com/expressjs/express
    $0 --local-path ~/.gibson/projects/foo/repo
    $0 -o security.json /path/to/project

EOF
    exit 0
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

# Cleanup
cleanup() {
    if [[ "$CLEANUP" == true ]] && [[ -n "$TEMP_DIR" ]] && [[ -d "$TEMP_DIR" ]]; then
        rm -rf "$TEMP_DIR"
    fi
}
trap cleanup EXIT

# Detect if target is a Git URL
is_git_url() {
    [[ "$1" =~ ^(https?|git)://.*\.git$ ]] || [[ "$1" =~ ^git@.*:.*\.git$ ]] || [[ "$1" =~ github\.com|gitlab\.com|bitbucket\.org ]]
}

# Get list of source files
get_source_files() {
    local repo_dir="$1"
    local max_files="$2"

    local extensions="py,js,ts,jsx,tsx,java,go,rb,php,c,cpp,h,hpp,cs,swift,kt,rs,scala,sh,bash"

    # Build find command
    local find_cmd="find \"$repo_dir\" -type f \\( "
    local first=true

    IFS=',' read -ra EXT_ARRAY <<< "$extensions"
    for ext in "${EXT_ARRAY[@]}"; do
        if [[ "$first" == "true" ]]; then
            find_cmd+="-name \"*.$ext\""
            first=false
        else
            find_cmd+=" -o -name \"*.$ext\""
        fi
    done
    find_cmd+=" \\)"

    # Add exclusions
    IFS=',' read -ra EXCLUDE_ARRAY <<< "$EXCLUDE_PATTERNS"
    for pattern in "${EXCLUDE_ARRAY[@]}"; do
        local path_pattern=$(echo "$pattern" | sed 's/\*\*/*/g')
        find_cmd+=" ! -path \"*/$path_pattern\""
    done

    eval "$find_cmd" 2>/dev/null | head -n "$max_files"
}

# Detect file type/language
detect_language() {
    local file="$1"
    case "$file" in
        *.py) echo "python" ;;
        *.js) echo "javascript" ;;
        *.ts) echo "typescript" ;;
        *.jsx) echo "javascript-react" ;;
        *.tsx) echo "typescript-react" ;;
        *.java) echo "java" ;;
        *.go) echo "go" ;;
        *.rb) echo "ruby" ;;
        *.php) echo "php" ;;
        *.c|*.h) echo "c" ;;
        *.cpp|*.hpp) echo "cpp" ;;
        *.cs) echo "csharp" ;;
        *.swift) echo "swift" ;;
        *.kt) echo "kotlin" ;;
        *.rs) echo "rust" ;;
        *.scala) echo "scala" ;;
        *.sh|*.bash) echo "shell" ;;
        *) echo "unknown" ;;
    esac
}

# Scan for potential security patterns (no AI, just pattern matching)
scan_security_patterns() {
    local repo_dir="$1"
    local findings="[]"

    # SQL injection patterns
    local sql_files=$(grep -rl "SELECT.*WHERE.*\$\|INSERT.*VALUES.*\$\|UPDATE.*SET.*\$" "$repo_dir" --include="*.py" --include="*.js" --include="*.php" --include="*.java" 2>/dev/null | head -20 || true)
    if [[ -n "$sql_files" ]]; then
        while IFS= read -r file; do
            [[ -z "$file" ]] && continue
            local rel_path="${file#$repo_dir/}"
            local line_nums=$(grep -n "SELECT.*WHERE.*\$\|INSERT.*VALUES.*\$\|UPDATE.*SET.*\$" "$file" 2>/dev/null | head -5 | cut -d: -f1 | tr '\n' ',' | sed 's/,$//')
            findings=$(echo "$findings" | jq --arg file "$rel_path" --arg lines "$line_nums" \
                '. + [{"type": "potential_sql_injection", "file": $file, "lines": $lines, "severity": "high", "confidence": "low", "note": "Pattern match only - requires manual review"}]')
        done <<< "$sql_files"
    fi

    # Command injection patterns
    local cmd_files=$(grep -rl "exec(\|system(\|popen(\|subprocess\|child_process\|os\.system" "$repo_dir" --include="*.py" --include="*.js" --include="*.php" --include="*.rb" 2>/dev/null | head -20 || true)
    if [[ -n "$cmd_files" ]]; then
        while IFS= read -r file; do
            [[ -z "$file" ]] && continue
            local rel_path="${file#$repo_dir/}"
            findings=$(echo "$findings" | jq --arg file "$rel_path" \
                '. + [{"type": "potential_command_execution", "file": $file, "severity": "high", "confidence": "low", "note": "Uses command execution - verify input sanitization"}]')
        done <<< "$cmd_files"
    fi

    # Eval patterns
    local eval_files=$(grep -rl "eval(\|new Function(" "$repo_dir" --include="*.js" --include="*.ts" --include="*.py" 2>/dev/null | head -20 || true)
    if [[ -n "$eval_files" ]]; then
        while IFS= read -r file; do
            [[ -z "$file" ]] && continue
            local rel_path="${file#$repo_dir/}"
            findings=$(echo "$findings" | jq --arg file "$rel_path" \
                '. + [{"type": "dynamic_code_execution", "file": $file, "severity": "medium", "confidence": "low", "note": "Uses eval or dynamic code - verify input source"}]')
        done <<< "$eval_files"
    fi

    # Weak crypto patterns
    local crypto_files=$(grep -rl "md5\|sha1\|DES\|RC4" "$repo_dir" --include="*.py" --include="*.js" --include="*.java" --include="*.go" 2>/dev/null | head -20 || true)
    if [[ -n "$crypto_files" ]]; then
        while IFS= read -r file; do
            [[ -z "$file" ]] && continue
            local rel_path="${file#$repo_dir/}"
            findings=$(echo "$findings" | jq --arg file "$rel_path" \
                '. + [{"type": "weak_cryptography", "file": $file, "severity": "medium", "confidence": "low", "note": "May use weak cryptographic algorithms"}]')
        done <<< "$crypto_files"
    fi

    # Hardcoded secrets patterns
    local secret_patterns="password.*=.*['\"][^'\"]+['\"]\|api_key.*=.*['\"][^'\"]+['\"]\|secret.*=.*['\"][^'\"]+['\"]\|token.*=.*['\"][^'\"]+['\"]"
    local secret_files=$(grep -ril "$secret_patterns" "$repo_dir" --include="*.py" --include="*.js" --include="*.ts" --include="*.java" --include="*.go" --include="*.rb" 2>/dev/null | head -20 || true)
    if [[ -n "$secret_files" ]]; then
        while IFS= read -r file; do
            [[ -z "$file" ]] && continue
            local rel_path="${file#$repo_dir/}"
            findings=$(echo "$findings" | jq --arg file "$rel_path" \
                '. + [{"type": "potential_hardcoded_secret", "file": $file, "severity": "high", "confidence": "low", "note": "May contain hardcoded credentials - verify"}]')
        done <<< "$secret_files"
    fi

    # Authentication security patterns (merged from auth scanner)

    # Insecure session/cookie settings
    local insecure_cookie_files=$(grep -rln "secure:\s*false\|httpOnly:\s*false\|sameSite:\s*['\"]none['\"]" "$repo_dir" --include="*.js" --include="*.ts" 2>/dev/null | head -20 || true)
    if [[ -n "$insecure_cookie_files" ]]; then
        while IFS= read -r file; do
            [[ -z "$file" ]] && continue
            local rel_path="${file#$repo_dir/}"
            findings=$(echo "$findings" | jq --arg file "$rel_path" \
                '. + [{"type": "insecure_cookie_settings", "file": $file, "severity": "high", "confidence": "medium", "note": "Cookie missing secure, httpOnly, or sameSite flags"}]')
        done <<< "$insecure_cookie_files"
    fi

    # JWT without expiration
    local jwt_no_exp=$(grep -rln "jwt\.\(sign\|encode\)" "$repo_dir" --include="*.js" --include="*.ts" --include="*.py" 2>/dev/null | while read f; do
        if ! grep -q "expiresIn\|exp:" "$f" 2>/dev/null; then echo "$f"; fi
    done | head -20)
    if [[ -n "$jwt_no_exp" ]]; then
        while IFS= read -r file; do
            [[ -z "$file" ]] && continue
            local rel_path="${file#$repo_dir/}"
            findings=$(echo "$findings" | jq --arg file "$rel_path" \
                '. + [{"type": "jwt_no_expiration", "file": $file, "severity": "high", "confidence": "medium", "note": "JWT signing without expiration - tokens may be valid indefinitely"}]')
        done <<< "$jwt_no_exp"
    fi

    # JWT algorithm none or disabled verification
    local jwt_algo_none=$(grep -rln "algorithm.*none\|alg.*none\|algorithms.*\[\s*\]" "$repo_dir" --include="*.js" --include="*.ts" --include="*.py" 2>/dev/null | head -10 || true)
    if [[ -n "$jwt_algo_none" ]]; then
        while IFS= read -r file; do
            [[ -z "$file" ]] && continue
            local rel_path="${file#$repo_dir/}"
            findings=$(echo "$findings" | jq --arg file "$rel_path" \
                '. + [{"type": "jwt_algorithm_none", "file": $file, "severity": "critical", "confidence": "medium", "note": "JWT with none algorithm or disabled verification - tokens can be forged"}]')
        done <<< "$jwt_algo_none"
    fi

    # OAuth callback without state parameter
    local oauth_no_state=$(grep -rln "oauth.*callback\|/callback.*oauth\|/auth/callback" "$repo_dir" --include="*.js" --include="*.ts" --include="*.py" 2>/dev/null | while read f; do
        if ! grep -q "state" "$f" 2>/dev/null; then echo "$f"; fi
    done | head -10)
    if [[ -n "$oauth_no_state" ]]; then
        while IFS= read -r file; do
            [[ -z "$file" ]] && continue
            local rel_path="${file#$repo_dir/}"
            findings=$(echo "$findings" | jq --arg file "$rel_path" \
                '. + [{"type": "oauth_missing_state", "file": $file, "severity": "high", "confidence": "low", "note": "OAuth callback may lack state parameter - vulnerable to CSRF"}]')
        done <<< "$oauth_no_state"
    fi

    # CORS wildcard origin
    local cors_wildcard=$(grep -rln "Access-Control-Allow-Origin.*\*\|cors.*origin.*['\"]\\*['\"]\|origin:\s*['\"]\\*['\"]" "$repo_dir" --include="*.js" --include="*.ts" --include="*.py" --include="*.java" --include="*.go" 2>/dev/null | head -20 || true)
    if [[ -n "$cors_wildcard" ]]; then
        while IFS= read -r file; do
            [[ -z "$file" ]] && continue
            local rel_path="${file#$repo_dir/}"
            findings=$(echo "$findings" | jq --arg file "$rel_path" \
                '. + [{"type": "cors_wildcard_origin", "file": $file, "severity": "medium", "confidence": "high", "note": "CORS allows any origin - may expose sensitive data to malicious sites"}]')
        done <<< "$cors_wildcard"
    fi

    # Weak password hashing (MD5/SHA1 for passwords specifically)
    local weak_pw_hash=$(grep -rln "md5.*password\|sha1.*password\|password.*md5\|password.*sha1\|hashlib\.md5.*password\|hashlib\.sha1.*password" "$repo_dir" --include="*.py" --include="*.js" --include="*.ts" --include="*.java" --include="*.php" 2>/dev/null | head -10 || true)
    if [[ -n "$weak_pw_hash" ]]; then
        while IFS= read -r file; do
            [[ -z "$file" ]] && continue
            local rel_path="${file#$repo_dir/}"
            findings=$(echo "$findings" | jq --arg file "$rel_path" \
                '. + [{"type": "weak_password_hashing", "file": $file, "severity": "critical", "confidence": "medium", "note": "MD5/SHA1 used for password hashing - use bcrypt/argon2/scrypt instead"}]')
        done <<< "$weak_pw_hash"
    fi

    echo "$findings"
}

# Scan for potential secrets
scan_secrets() {
    local repo_dir="$1"
    local findings="[]"

    # AWS keys pattern
    local aws_keys=$(grep -rn "AKIA[0-9A-Z]\{16\}" "$repo_dir" --include="*" 2>/dev/null | head -10 || true)
    if [[ -n "$aws_keys" ]]; then
        while IFS=: read -r file line content; do
            [[ -z "$file" ]] && continue
            local rel_path="${file#$repo_dir/}"
            findings=$(echo "$findings" | jq --arg file "$rel_path" --arg line "$line" \
                '. + [{"type": "aws_access_key", "file": $file, "line": ($line | tonumber), "severity": "critical"}]')
        done <<< "$aws_keys"
    fi

    # GitHub tokens
    local gh_tokens=$(grep -rn "gh[pousr]_[A-Za-z0-9_]\{36,\}" "$repo_dir" --include="*" 2>/dev/null | head -10 || true)
    if [[ -n "$gh_tokens" ]]; then
        while IFS=: read -r file line content; do
            [[ -z "$file" ]] && continue
            local rel_path="${file#$repo_dir/}"
            findings=$(echo "$findings" | jq --arg file "$rel_path" --arg line "$line" \
                '. + [{"type": "github_token", "file": $file, "line": ($line | tonumber), "severity": "critical"}]')
        done <<< "$gh_tokens"
    fi

    # Private keys
    local priv_keys=$(grep -rln "BEGIN.*PRIVATE KEY" "$repo_dir" --include="*" 2>/dev/null | head -10 || true)
    if [[ -n "$priv_keys" ]]; then
        while IFS= read -r file; do
            [[ -z "$file" ]] && continue
            local rel_path="${file#$repo_dir/}"
            findings=$(echo "$findings" | jq --arg file "$rel_path" \
                '. + [{"type": "private_key", "file": $file, "severity": "critical"}]')
        done <<< "$priv_keys"
    fi

    echo "$findings"
}

# Main analysis
analyze_target() {
    local repo_dir="$1"

    echo -e "${BLUE}Collecting file inventory...${NC}" >&2
    local files=$(get_source_files "$repo_dir" "$MAX_FILES")
    local file_count=0
    if [[ -n "$files" ]]; then
        file_count=$(echo "$files" | wc -l | tr -d ' ')
    fi
    echo -e "${GREEN}✓ Found $file_count source files${NC}" >&2

    # Build file inventory
    local file_inventory="[]"
    local lang_counts="{}"
    while IFS= read -r file; do
        [[ -z "$file" ]] && continue
        local rel_path="${file#$repo_dir/}"
        local lang=$(detect_language "$file")
        local size=$(wc -c < "$file" 2>/dev/null | tr -d ' ')

        file_inventory=$(echo "$file_inventory" | jq --arg path "$rel_path" --arg lang "$lang" --argjson size "$size" \
            '. + [{"path": $path, "language": $lang, "size_bytes": $size}]')

        # Count languages
        lang_counts=$(echo "$lang_counts" | jq --arg lang "$lang" '.[$lang] = ((.[$lang] // 0) + 1)')
    done <<< "$files"

    echo -e "${BLUE}Scanning for security patterns...${NC}" >&2
    local security_findings=$(scan_security_patterns "$repo_dir")
    local security_count=$(echo "$security_findings" | jq 'length')
    echo -e "${GREEN}✓ Found $security_count potential issues${NC}" >&2

    echo -e "${BLUE}Scanning for secrets...${NC}" >&2
    local secrets_findings=$(scan_secrets "$repo_dir")
    local secrets_count=$(echo "$secrets_findings" | jq 'length')
    if [[ "$secrets_count" -gt 0 ]]; then
        echo -e "${YELLOW}⚠ Found $secrets_count potential secrets${NC}" >&2
    else
        echo -e "${GREEN}✓ No obvious secrets found${NC}" >&2
    fi

    # Build final output
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    jq -n \
        --arg ts "$timestamp" \
        --arg tgt "$TARGET" \
        --arg ver "1.0.0" \
        --argjson file_count "$file_count" \
        --argjson lang_counts "$lang_counts" \
        --argjson files "$file_inventory" \
        --argjson security "$security_findings" \
        --argjson secrets "$secrets_findings" \
        '{
            analyzer: "code-security",
            version: $ver,
            timestamp: $ts,
            target: $tgt,
            summary: {
                total_files: $file_count,
                languages: $lang_counts,
                potential_issues: ($security | length),
                potential_secrets: ($secrets | length)
            },
            file_inventory: $files,
            potential_issues: $security,
            secrets_scan: $secrets,
            note: "Pattern-based detection only. All findings require manual review. AI analysis is performed by agents."
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
        --max-files)
            MAX_FILES="$2"
            shift 2
            ;;
        --exclude)
            EXCLUDE_PATTERNS="$2"
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

# Main execution
scan_path=""

if [[ -n "$LOCAL_PATH" ]]; then
    [[ ! -d "$LOCAL_PATH" ]] && { echo '{"error": "Local path does not exist"}'; exit 1; }
    scan_path="$LOCAL_PATH"
    TARGET="$LOCAL_PATH"
elif [[ -n "$TARGET" ]]; then
    if is_git_url "$TARGET"; then
        clone_repository "$TARGET"
        scan_path="$TEMP_DIR"
    elif [[ -d "$TARGET" ]]; then
        scan_path="$TARGET"
    else
        echo '{"error": "Invalid target - must be URL or directory"}'
        exit 1
    fi
else
    echo '{"error": "No target specified"}'
    exit 1
fi

echo -e "${BLUE}Analyzing: $TARGET${NC}" >&2

final_json=$(analyze_target "$scan_path")

# Output
if [[ -n "$OUTPUT_FILE" ]]; then
    echo "$final_json" > "$OUTPUT_FILE"
    echo -e "${GREEN}✓ Results written to $OUTPUT_FILE${NC}" >&2
else
    echo "$final_json"
fi
