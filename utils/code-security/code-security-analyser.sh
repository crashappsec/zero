#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Code Security Analyser
# AI-powered security code review using Claude
# Scans repositories for vulnerabilities, secrets, and security weaknesses
# Usage: ./code-security-analyser.sh [options] [targets...]
#############################################################################

set -e

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
UTILS_ROOT="$(dirname "$SCRIPT_DIR")"
REPO_ROOT="$(dirname "$UTILS_ROOT")"

# Load shared libraries
if [[ -f "$UTILS_ROOT/lib/config.sh" ]]; then
    source "$UTILS_ROOT/lib/config.sh"
fi

if [[ -f "$UTILS_ROOT/lib/github.sh" ]]; then
    source "$UTILS_ROOT/lib/github.sh"
fi

if [[ -f "$UTILS_ROOT/lib/config-loader.sh" ]]; then
    source "$UTILS_ROOT/lib/config-loader.sh"
fi

# Load .env file if it exists in repository root
if [[ -f "$REPO_ROOT/.env" ]]; then
    set -a
    source "$REPO_ROOT/.env"
    set +a
fi

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

# Default options
TARGET=""
TARGET_TYPE=""  # repo, org, local
OUTPUT_DIR="./code-security-reports"
OUTPUT_FORMAT="markdown"
MIN_SEVERITY="low"
FAIL_ON_SEVERITY=""
CATEGORIES="all"
EXCLUDE_PATTERNS="node_modules/**,vendor/**,.git/**,*.min.js,*.bundle.js"
MAX_FILES=500
INCLUDE_SUPPLY_CHAIN=false
TEMP_DIR=""

# Claude configuration
ANTHROPIC_API_KEY="${ANTHROPIC_API_KEY:-}"
USE_CLAUDE=true
if [[ -z "$ANTHROPIC_API_KEY" ]]; then
    USE_CLAUDE=false
fi

# Cleanup function
cleanup() {
    if [[ -n "$TEMP_DIR" ]] && [[ -d "$TEMP_DIR" ]]; then
        rm -rf "$TEMP_DIR"
    fi
}

trap cleanup EXIT

# Print usage
usage() {
    cat << EOF
Code Security Analyser - AI-powered security code review

Usage: $0 [OPTIONS] [TARGET]

TARGETS:
    --repo OWNER/REPO       Scan specific GitHub repository
    --org ORG_NAME          Scan all repos in GitHub organization
    --local PATH            Scan local directory
    (No target = scan current directory)

OPTIONS:
    --output DIR, -o DIR    Output directory for reports (default: ./code-security-reports)
    --format FORMAT         Output format: markdown|json|sarif (default: markdown)
    --severity LEVEL        Minimum severity to report: low|medium|high|critical (default: low)
    --fail-on LEVEL         Exit non-zero if findings >= level
    --categories LIST       Comma-separated categories to check (default: all)
                           Options: injection,auth,crypto,exposure,validation,secrets,config
    --exclude PATTERNS      Glob patterns to exclude (default: node_modules/**,vendor/**,.git/**)
    --max-files N           Maximum files to analyse (default: 500)
    --supply-chain          Include supply chain analysis (calls supply-chain-scanner.sh)
    --no-supply-chain       Skip supply chain analysis (default)
    --claude                Enable Claude AI analysis (default if ANTHROPIC_API_KEY set)
    --no-claude             Disable Claude AI analysis
    -h, --help              Show this help message

CATEGORIES:
    injection    - SQL, command, LDAP, XPath injection
    auth         - Authentication and authorization flaws
    crypto       - Cryptographic weaknesses
    exposure     - Data exposure and information leakage
    validation   - Input validation (XSS, path traversal, SSRF)
    secrets      - Hardcoded secrets and credentials
    config       - Security misconfigurations
    all          - All categories (default)

EXAMPLES:
    # Scan a GitHub repository
    $0 --repo owner/repo

    # Scan local project with supply chain analysis
    $0 --local /path/to/project --supply-chain

    # Scan with minimum severity filter
    $0 --repo owner/repo --severity high

    # CI/CD mode - fail on critical findings
    $0 --repo owner/repo --fail-on critical --format sarif

    # Scan specific categories only
    $0 --local . --categories injection,secrets,auth

ENVIRONMENT:
    ANTHROPIC_API_KEY       Required for Claude AI analysis

EOF
    exit 1
}

# Check prerequisites
check_prerequisites() {
    local missing=()

    if ! command -v jq &> /dev/null; then
        missing+=("jq")
    fi

    if ! command -v git &> /dev/null; then
        missing+=("git")
    fi

    if [[ ${#missing[@]} -gt 0 ]]; then
        echo -e "${RED}Error: Missing required tools:${NC}"
        for tool in "${missing[@]}"; do
            echo "  - $tool"
        done
        echo ""
        echo "Install missing tools:"
        echo "  brew install jq git"
        exit 1
    fi

    # Check for Claude API key
    if [[ "$USE_CLAUDE" == "true" ]] && [[ -z "$ANTHROPIC_API_KEY" ]]; then
        echo -e "${RED}Error: ANTHROPIC_API_KEY is required for Claude analysis${NC}"
        echo ""
        echo "To enable Claude AI analysis:"
        echo "  1. Get an API key at: https://console.anthropic.com/settings/keys"
        echo "  2. Export it: export ANTHROPIC_API_KEY='your-key'"
        exit 1
    fi
}

# Check Claude API key status
check_api_key() {
    if [[ -n "$ANTHROPIC_API_KEY" ]]; then
        local key_length=${#ANTHROPIC_API_KEY}
        echo -e "${GREEN}✓ ANTHROPIC_API_KEY is set (${key_length} chars)${NC}"
        return 0
    else
        echo -e "${YELLOW}⚠ ANTHROPIC_API_KEY is NOT set${NC}"
        echo -e "${YELLOW}  Claude AI analysis will be disabled${NC}"
        return 1
    fi
}

# Normalize target to consistent format
normalize_target() {
    local target="$1"

    # Already a URL
    if [[ "$target" =~ ^https?:// ]]; then
        echo "$target"
        return
    fi

    # GitHub SSH URL
    if [[ "$target" =~ ^git@ ]]; then
        echo "$target"
        return
    fi

    # owner/repo format -> GitHub URL
    if [[ "$target" =~ ^[a-zA-Z0-9_-]+/[a-zA-Z0-9_.-]+$ ]]; then
        echo "https://github.com/$target"
        return
    fi

    # Local path or other
    echo "$target"
}

# Clone repository to temp directory
clone_repository() {
    local target="$1"

    TEMP_DIR=$(mktemp -d)

    echo -e "${BLUE}Cloning repository...${NC}" >&2

    local repo_url=$(normalize_target "$target")

    if git clone --depth 1 --quiet "$repo_url" "$TEMP_DIR" 2>/dev/null; then
        local file_count=$(find "$TEMP_DIR" -type f | wc -l | tr -d ' ')
        local repo_size=$(du -sh "$TEMP_DIR" 2>/dev/null | cut -f1)
        echo -e "${GREEN}✓ Repository cloned: ${file_count} files, ${repo_size}${NC}" >&2
        echo "$TEMP_DIR"
        return 0
    else
        echo -e "${RED}✗ Failed to clone repository: $repo_url${NC}" >&2
        rm -rf "$TEMP_DIR"
        TEMP_DIR=""
        return 1
    fi
}

# Get list of files to scan
get_target_files() {
    local repo_dir="$1"
    local max_files="$2"

    # Source file extensions to scan (code + security-relevant config files)
    local extensions="py,js,ts,jsx,tsx,java,go,rb,php,c,cpp,h,hpp,cs,swift,kt,rs,scala,sh,bash,yaml,yml,json,toml,xml,tf,hcl,Dockerfile,docker-compose.yml,Makefile,gradle,pom.xml,Gemfile,requirements.txt,package.json,Cargo.toml,go.mod"

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
        # Convert glob to path pattern
        local path_pattern=$(echo "$pattern" | sed 's/\*\*/*/g')
        find_cmd+=" ! -path \"*/$path_pattern\""
    done

    # Execute and limit results
    eval "$find_cmd" 2>/dev/null | head -n "$max_files"
}

# Map severity to numeric value for comparison
severity_to_num() {
    case "$1" in
        critical) echo 4 ;;
        high) echo 3 ;;
        medium) echo 2 ;;
        low) echo 1 ;;
        *) echo 0 ;;
    esac
}

# Load RAG content for security analysis
load_security_rag() {
    local rag_dir="$RAG_DIR/code-security"

    if [[ ! -d "$rag_dir" ]]; then
        return 1
    fi

    local rag_content=""

    # Load vulnerability patterns
    if [[ -d "$rag_dir/vulnerability-patterns" ]]; then
        for file in "$rag_dir/vulnerability-patterns"/*.md; do
            if [[ -f "$file" ]]; then
                rag_content+="$(cat "$file")"$'\n\n'
            fi
        done
    fi

    # Load standards
    if [[ -d "$rag_dir/standards" ]]; then
        for file in "$rag_dir/standards"/*.md; do
            if [[ -f "$file" ]]; then
                rag_content+="$(cat "$file")"$'\n\n'
            fi
        done
    fi

    echo "$rag_content"
}

# Load security analysis prompt
load_security_prompt() {
    local prompt_file="$REPO_ROOT/prompts/code-security/analysis/security-review.md"

    if [[ -f "$prompt_file" ]]; then
        cat "$prompt_file"
    else
        # Fallback inline prompt
        cat << 'PROMPT'
You are a security expert analysing source code for vulnerabilities.

## Your Task

Analyse the provided code for security issues. For each finding:
1. Identify the specific vulnerability
2. Explain why it's a security risk
3. Rate the severity (Critical/High/Medium/Low)
4. Provide remediation guidance with code examples

## Vulnerability Categories

Check for these security issues:
- **Injection**: SQL, command, LDAP, XPath injection
- **Authentication**: Hardcoded credentials, weak auth, missing checks
- **Cryptography**: Weak algorithms, hardcoded keys, insecure random
- **Data Exposure**: Sensitive data in logs, PII exposure
- **Input Validation**: XSS, path traversal, SSRF, open redirects
- **Secrets**: Hardcoded API keys, passwords, tokens
- **Configuration**: Debug mode, insecure defaults, CORS issues

## Output Format

Return findings as a JSON array:
```json
[
  {
    "file": "path/to/file.py",
    "line": 42,
    "category": "injection",
    "type": "SQL Injection",
    "severity": "critical",
    "confidence": "high",
    "cwe": "CWE-89",
    "description": "User input directly concatenated into SQL query",
    "code_snippet": "query = \"SELECT * FROM users WHERE id=\" + user_id",
    "remediation": "Use parameterized queries",
    "exploitation": "Attacker can inject SQL to extract or modify data"
  }
]
```

If no security issues are found, return an empty array: []
PROMPT
    fi
}

# Analyse code with Claude
analyse_with_claude() {
    local code_content="$1"
    local file_path="$2"

    local prompt=$(load_security_prompt)
    local rag_content=$(load_security_rag)

    # Build the full prompt
    local full_prompt="$prompt"

    if [[ -n "$rag_content" ]]; then
        full_prompt+=$'\n\n## Reference Documentation\n\n'"$rag_content"
    fi

    full_prompt+=$'\n\n## File: '"$file_path"$'\n\n```\n'"$code_content"$'\n```'

    # Call Claude API
    local response=$(curl -s "https://api.anthropic.com/v1/messages" \
        -H "Content-Type: application/json" \
        -H "x-api-key: $ANTHROPIC_API_KEY" \
        -H "anthropic-version: 2023-06-01" \
        -d "$(jq -n \
            --arg prompt "$full_prompt" \
            '{
                model: "claude-sonnet-4-20250514",
                max_tokens: 4096,
                messages: [
                    {
                        role: "user",
                        content: $prompt
                    }
                ]
            }')")

    # Extract the response text
    echo "$response" | jq -r '.content[0].text // empty'
}

# Extract JSON findings from Claude response
extract_findings() {
    local response="$1"

    # Try to extract JSON array from response
    # Look for content between ``` markers or raw JSON array
    local json_content

    # First try: extract from code block
    json_content=$(echo "$response" | sed -n '/```json/,/```/p' | sed '1d;$d')

    # Second try: extract from plain code block
    if [[ -z "$json_content" ]] || ! echo "$json_content" | jq . &>/dev/null; then
        json_content=$(echo "$response" | sed -n '/```/,/```/p' | sed '1d;$d')
    fi

    # Third try: find raw JSON array
    if [[ -z "$json_content" ]] || ! echo "$json_content" | jq . &>/dev/null; then
        json_content=$(echo "$response" | grep -o '\[.*\]' | head -1)
    fi

    # Validate JSON
    if echo "$json_content" | jq . &>/dev/null; then
        echo "$json_content"
    else
        echo "[]"
    fi
}

# Run supply chain analysis
run_supply_chain_analysis() {
    local repo_dir="$1"
    local output_dir="$2"

    local scanner="$UTILS_ROOT/supply-chain/supply-chain-scanner.sh"

    if [[ ! -f "$scanner" ]]; then
        echo -e "${YELLOW}⚠ Supply chain scanner not found, skipping${NC}"
        return 1
    fi

    echo -e "${BLUE}Running supply chain analysis...${NC}"

    mkdir -p "$output_dir/supply-chain"

    # Call the supply chain scanner
    "$scanner" \
        --local "$repo_dir" \
        --vulnerability \
        --package-health \
        --output "$output_dir/supply-chain" \
        ${USE_CLAUDE:+--claude} \
        2>&1 || true

    echo -e "${GREEN}✓ Supply chain analysis complete${NC}"
}

# Generate markdown report
generate_markdown_report() {
    local findings_file="$1"
    local output_file="$2"
    local target="$3"
    local scan_time="$4"

    local total=$(jq 'length' "$findings_file")
    local critical=$(jq '[.[] | select(.severity == "critical")] | length' "$findings_file")
    local high=$(jq '[.[] | select(.severity == "high")] | length' "$findings_file")
    local medium=$(jq '[.[] | select(.severity == "medium")] | length' "$findings_file")
    local low=$(jq '[.[] | select(.severity == "low")] | length' "$findings_file")

    # Get unique files with findings
    local files_with_findings=$(jq -r '[.[].file] | unique | length' "$findings_file")

    # Get vulnerability types breakdown
    local vuln_types=$(jq -r '[.[].type] | group_by(.) | map({type: .[0], count: length}) | sort_by(-.count) | .[0:5] | .[] | "| \(.type) | \(.count) |"' "$findings_file")

    # Determine risk level
    local risk_level="Low"
    local risk_color="🟢"
    if [[ "$critical" -gt 0 ]]; then
        risk_level="Critical"
        risk_color="🔴"
    elif [[ "$high" -gt 0 ]]; then
        risk_level="High"
        risk_color="🟠"
    elif [[ "$medium" -gt 0 ]]; then
        risk_level="Medium"
        risk_color="🟡"
    fi

    cat > "$output_file" << EOF
# 🔒 Code Security Analysis Report

## Scan Information

| Property | Value |
|----------|-------|
| **Repository** | \`$target\` |
| **Scan Date** | $(date '+%Y-%m-%d %H:%M:%S %Z') |
| **Duration** | ${scan_time} seconds |
| **Tool** | Gibson Powers Code Security Analyser v1.0 |
| **Analysis Engine** | Claude AI (Anthropic) |

---

## Executive Summary

$risk_color **Overall Risk Level: $risk_level**

This security analysis identified **$total** potential security issue(s) across **$files_with_findings** file(s).

### Severity Distribution

| Severity | Count | Risk |
|----------|-------|------|
| 🔴 **Critical** | $critical | Immediate action required |
| 🟠 **High** | $high | Address within 24-48 hours |
| 🟡 **Medium** | $medium | Plan remediation this sprint |
| 🟢 **Low** | $low | Address as time permits |
| **Total** | **$total** | |

EOF

    if [[ "$total" -gt 0 ]]; then
        # Add vulnerability type breakdown
        cat >> "$output_file" << 'EOF'
### Top Vulnerability Types

| Type | Count |
|------|-------|
EOF
        echo "$vuln_types" >> "$output_file"

        # Add findings by file summary
        echo "" >> "$output_file"
        echo "### Affected Files" >> "$output_file"
        echo "" >> "$output_file"
        jq -r 'group_by(.file) | .[] | "- **\(.[0].file)**: \(length) finding(s)"' "$findings_file" >> "$output_file"

        echo "" >> "$output_file"
        echo "---" >> "$output_file"
        echo "" >> "$output_file"
        echo "## Detailed Findings" >> "$output_file"
        echo "" >> "$output_file"

        local finding_num=1

        # Group by severity
        for severity in critical high medium low; do
            local count=$(jq "[.[] | select(.severity == \"$severity\")] | length" "$findings_file")
            if [[ "$count" -gt 0 ]]; then
                local emoji="🟢"
                local severity_desc="Low Priority"
                case "$severity" in
                    critical) emoji="🔴"; severity_desc="Immediate Action Required" ;;
                    high) emoji="🟠"; severity_desc="High Priority" ;;
                    medium) emoji="🟡"; severity_desc="Medium Priority" ;;
                esac

                echo "### $emoji ${severity^} Severity Findings ($count)" >> "$output_file"
                echo "" >> "$output_file"
                echo "*$severity_desc*" >> "$output_file"
                echo "" >> "$output_file"

                # Output each finding
                jq -r --argjson start "$finding_num" ".[] | select(.severity == \"$severity\") | . as \$f | \"
<details>
<summary><strong>$emoji #\($start + (input_line_number - 1)) \(.type)</strong> - \(.file):\(.line // \"?\")</summary>

#### Details

| Property | Value |
|----------|-------|
| **File** | \`\(.file)\` |
| **Line** | \(.line // \"Unknown\") |
| **Category** | \(.category // \"N/A\") |
| **CWE** | [\(.cwe // \"N/A\")](https://cwe.mitre.org/data/definitions/\(.cwe // \"\" | gsub(\"CWE-\"; \"\")).html) |
| **Confidence** | \(.confidence // \"medium\") |

#### Description

\(.description)

#### Vulnerable Code

\\\`\\\`\\\`
\(.code_snippet // \"N/A\")
\\\`\\\`\\\`

#### Evidence

\(.evidence // \"See code snippet above.\")

#### Exploit Scenario

\(.exploit_scenario // .exploitation // \"Attacker could potentially exploit this vulnerability to compromise the application.\")

#### Remediation

\(.remediation)

</details>
\"" "$findings_file" | sed 's/\\`\\`\\`/```/g' >> "$output_file"

                finding_num=$((finding_num + count))
                echo "" >> "$output_file"
            fi
        done
    else
        cat >> "$output_file" << 'EOF'

---

## ✅ No Security Issues Found

The automated security analysis did not identify any security vulnerabilities in the scanned code.

**Note**: This does not guarantee the code is free of all security issues. Consider:
- Manual code review for business logic flaws
- Dynamic testing (DAST) for runtime vulnerabilities
- Penetration testing for complex attack scenarios

EOF
    fi

    cat >> "$output_file" << 'EOF'

---

## Methodology

This analysis was performed using AI-powered static analysis with the following approach:

1. **Context Understanding** - Identify code purpose, data handling, and trust boundaries
2. **Data Flow Analysis** - Trace user input from entry points to sensitive sinks
3. **Pattern Matching** - Detect known vulnerability patterns with CWE classification
4. **Confidence Filtering** - Report only high-confidence findings (≥80%)

### Vulnerability Categories Checked

- **Injection**: SQL, Command, NoSQL, LDAP, XPath, Template
- **Authentication & Authorization**: Broken access control, missing auth, IDOR
- **Cryptographic Failures**: Weak algorithms, hardcoded keys, insecure random
- **Code Execution**: Deserialization, eval/exec, prototype pollution
- **Input Validation**: XSS, path traversal, SSRF, XXE, open redirects
- **Secrets**: Hardcoded credentials, API keys, tokens
- **Configuration**: CORS, cookies, security headers, debug mode

---

*Report generated by [Gibson Powers Code Security Analyser](https://github.com/crashappsec/gibson-powers)*
*Powered by Claude AI (Anthropic)*
EOF
}

# Generate JSON report
generate_json_report() {
    local findings_file="$1"
    local output_file="$2"
    local target="$3"
    local scan_time="$4"

    jq -n \
        --arg target "$target" \
        --arg date "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" \
        --arg duration "$scan_time" \
        --slurpfile findings "$findings_file" \
        '{
            metadata: {
                target: $target,
                timestamp: $date,
                duration_seconds: ($duration | tonumber),
                tool: "Gibson Powers Code Security Analyser",
                version: "1.0.0"
            },
            summary: {
                total: ($findings[0] | length),
                critical: ([$findings[0][] | select(.severity == "critical")] | length),
                high: ([$findings[0][] | select(.severity == "high")] | length),
                medium: ([$findings[0][] | select(.severity == "medium")] | length),
                low: ([$findings[0][] | select(.severity == "low")] | length)
            },
            findings: $findings[0]
        }' > "$output_file"
}

# Generate SARIF report (for GitHub code scanning)
generate_sarif_report() {
    local findings_file="$1"
    local output_file="$2"
    local target="$3"

    jq -n \
        --slurpfile findings "$findings_file" \
        '{
            "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
            version: "2.1.0",
            runs: [{
                tool: {
                    driver: {
                        name: "Gibson Powers Code Security Analyser",
                        version: "1.0.0",
                        informationUri: "https://github.com/crashappsec/gibson-powers"
                    }
                },
                results: [
                    $findings[0][] | {
                        ruleId: .cwe,
                        level: (if .severity == "critical" then "error" elif .severity == "high" then "error" elif .severity == "medium" then "warning" else "note" end),
                        message: { text: .description },
                        locations: [{
                            physicalLocation: {
                                artifactLocation: { uri: .file },
                                region: { startLine: (.line // 1) }
                            }
                        }]
                    }
                ]
            }]
        }' > "$output_file"
}

# Main analysis function
run_analysis() {
    local repo_dir="$1"
    local output_dir="$2"

    mkdir -p "$output_dir"

    local start_time=$(date +%s)
    local all_findings="[]"

    echo -e "${CYAN}=========================================${NC}"
    echo -e "${CYAN}  Code Security Analysis${NC}"
    echo -e "${CYAN}=========================================${NC}"
    echo ""

    # Get files to scan
    echo -e "${BLUE}Identifying files to scan...${NC}"
    local files=$(get_target_files "$repo_dir" "$MAX_FILES")
    local file_count=0
    if [[ -n "$files" ]]; then
        file_count=$(echo "$files" | wc -l | tr -d ' ')
    fi

    echo -e "${GREEN}✓ Found $file_count files to analyse${NC}"
    echo ""

    if [[ "$file_count" -eq 0 ]]; then
        echo -e "${YELLOW}No source files found to analyse${NC}"
        return 0
    fi

    if [[ "$USE_CLAUDE" != "true" ]]; then
        echo -e "${RED}Error: Claude AI is required for security analysis${NC}"
        echo -e "${YELLOW}Set ANTHROPIC_API_KEY to enable analysis${NC}"
        return 1
    fi

    # Analyse files
    local analysed=0
    local findings_with_issues=0

    echo -e "${BLUE}Analysing files with Claude AI...${NC}"
    echo ""

    while IFS= read -r file; do
        [[ -z "$file" ]] && continue

        local relative_path="${file#$repo_dir/}"
        ((analysed++))

        printf "\r  [%d/%d] Analysing: %-50s" "$analysed" "$file_count" "${relative_path:0:50}"

        # Read file content
        local content=$(cat "$file" 2>/dev/null | head -c 50000)  # Limit to 50KB

        if [[ -z "$content" ]]; then
            continue
        fi

        # Analyse with Claude
        local response=$(analyse_with_claude "$content" "$relative_path")

        if [[ -n "$response" ]]; then
            local findings=$(extract_findings "$response")

            if [[ "$findings" != "[]" ]] && [[ -n "$findings" ]]; then
                # Add file path to findings if not present
                findings=$(echo "$findings" | jq --arg file "$relative_path" '[.[] | .file = $file]')

                # Display findings inline in terminal
                local file_finding_count=$(echo "$findings" | jq 'length')
                echo ""  # New line after progress indicator
                echo -e "${YELLOW}  ⚠ Found $file_finding_count issue(s) in $relative_path:${NC}"

                # Show each finding briefly
                echo "$findings" | jq -r '.[] | "    \(if .severity == "critical" then "🔴" elif .severity == "high" then "🟠" elif .severity == "medium" then "🟡" else "🟢" end) [\(.severity | ascii_upcase)] \(.type): \(.description | .[0:80])..."' 2>/dev/null || true

                # Merge findings
                all_findings=$(echo "$all_findings $findings" | jq -s 'add')
                ((findings_with_issues++))
            fi
        fi

        # Rate limiting - avoid hitting API limits
        sleep 0.5

    done <<< "$files"

    echo ""
    echo ""

    local end_time=$(date +%s)
    local scan_time=$((end_time - start_time))

    # Filter by minimum severity
    local min_sev_num=$(severity_to_num "$MIN_SEVERITY")
    all_findings=$(echo "$all_findings" | jq --argjson min "$min_sev_num" '[.[] | select(
        (.severity == "critical" and $min <= 4) or
        (.severity == "high" and $min <= 3) or
        (.severity == "medium" and $min <= 2) or
        (.severity == "low" and $min <= 1)
    )]')

    # Save findings
    local findings_file="$output_dir/findings.json"
    echo "$all_findings" > "$findings_file"

    # Generate reports
    local total_findings=$(echo "$all_findings" | jq 'length')

    echo -e "${GREEN}✓ Analysis complete: $total_findings findings${NC}"
    echo ""

    # Generate report based on format
    case "$OUTPUT_FORMAT" in
        markdown)
            local report_file="$output_dir/security-report.md"
            generate_markdown_report "$findings_file" "$report_file" "$TARGET" "$scan_time"
            echo -e "${GREEN}✓ Report saved: $report_file${NC}"
            ;;
        json)
            local report_file="$output_dir/security-report.json"
            generate_json_report "$findings_file" "$report_file" "$TARGET" "$scan_time"
            echo -e "${GREEN}✓ Report saved: $report_file${NC}"
            ;;
        sarif)
            local report_file="$output_dir/security-report.sarif"
            generate_sarif_report "$findings_file" "$report_file" "$TARGET"
            echo -e "${GREEN}✓ Report saved: $report_file${NC}"
            ;;
    esac

    # Print summary
    echo ""
    echo -e "${CYAN}Summary:${NC}"
    echo -e "  Files analysed: $analysed"
    echo -e "  Files with findings: $findings_with_issues"
    echo -e "  Total findings: $total_findings"
    echo -e "  Duration: ${scan_time}s"

    local critical=$(echo "$all_findings" | jq '[.[] | select(.severity == "critical")] | length')
    local high=$(echo "$all_findings" | jq '[.[] | select(.severity == "high")] | length')
    local medium=$(echo "$all_findings" | jq '[.[] | select(.severity == "medium")] | length')
    local low=$(echo "$all_findings" | jq '[.[] | select(.severity == "low")] | length')

    echo ""
    echo -e "  🔴 Critical: $critical"
    echo -e "  🟠 High: $high"
    echo -e "  🟡 Medium: $medium"
    echo -e "  🟢 Low: $low"

    # Run supply chain analysis if requested
    if [[ "$INCLUDE_SUPPLY_CHAIN" == "true" ]]; then
        echo ""
        run_supply_chain_analysis "$repo_dir" "$output_dir"
    fi

    # Check fail-on threshold
    if [[ -n "$FAIL_ON_SEVERITY" ]]; then
        local fail_threshold=$(severity_to_num "$FAIL_ON_SEVERITY")
        local should_fail=false

        if [[ "$fail_threshold" -le 4 ]] && [[ "$critical" -gt 0 ]]; then
            should_fail=true
        elif [[ "$fail_threshold" -le 3 ]] && [[ "$high" -gt 0 ]]; then
            should_fail=true
        elif [[ "$fail_threshold" -le 2 ]] && [[ "$medium" -gt 0 ]]; then
            should_fail=true
        elif [[ "$fail_threshold" -le 1 ]] && [[ "$low" -gt 0 ]]; then
            should_fail=true
        fi

        if [[ "$should_fail" == "true" ]]; then
            echo ""
            echo -e "${RED}✗ Failing due to findings at or above $FAIL_ON_SEVERITY severity${NC}"
            exit 1
        fi
    fi
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --repo)
                TARGET="$2"
                TARGET_TYPE="repo"
                shift 2
                ;;
            --org)
                TARGET="$2"
                TARGET_TYPE="org"
                shift 2
                ;;
            --local)
                TARGET="$2"
                TARGET_TYPE="local"
                shift 2
                ;;
            --output|-o)
                OUTPUT_DIR="$2"
                shift 2
                ;;
            --format)
                OUTPUT_FORMAT="$2"
                shift 2
                ;;
            --severity)
                MIN_SEVERITY="$2"
                shift 2
                ;;
            --fail-on)
                FAIL_ON_SEVERITY="$2"
                shift 2
                ;;
            --categories)
                CATEGORIES="$2"
                shift 2
                ;;
            --exclude)
                EXCLUDE_PATTERNS="$2"
                shift 2
                ;;
            --max-files)
                MAX_FILES="$2"
                shift 2
                ;;
            --supply-chain)
                INCLUDE_SUPPLY_CHAIN=true
                shift
                ;;
            --no-supply-chain)
                INCLUDE_SUPPLY_CHAIN=false
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
            -h|--help)
                usage
                ;;
            *)
                # Assume positional argument is target
                if [[ -z "$TARGET" ]]; then
                    TARGET="$1"
                    if [[ -d "$1" ]]; then
                        TARGET_TYPE="local"
                    else
                        TARGET_TYPE="repo"
                    fi
                fi
                shift
                ;;
        esac
    done

    # Default to current directory if no target specified
    if [[ -z "$TARGET" ]]; then
        TARGET="$(pwd)"
        TARGET_TYPE="local"
    fi
}

# Main entry point
main() {
    parse_args "$@"

    echo -e "${MAGENTA}"
    echo "╔═══════════════════════════════════════════════════════════╗"
    echo "║         Gibson Powers Code Security Analyser              ║"
    echo "║         AI-Powered Security Code Review                   ║"
    echo "╚═══════════════════════════════════════════════════════════╝"
    echo -e "${NC}"

    check_prerequisites

    echo -e "${BLUE}Configuration:${NC}"
    echo "  Target: $TARGET ($TARGET_TYPE)"
    echo "  Output: $OUTPUT_DIR"
    echo "  Format: $OUTPUT_FORMAT"
    echo "  Min Severity: $MIN_SEVERITY"
    echo "  Max Files: $MAX_FILES"
    echo "  Supply Chain: $INCLUDE_SUPPLY_CHAIN"
    echo ""

    check_api_key
    echo ""

    local repo_dir=""

    case "$TARGET_TYPE" in
        local)
            if [[ ! -d "$TARGET" ]]; then
                echo -e "${RED}Error: Directory not found: $TARGET${NC}"
                exit 1
            fi
            repo_dir="$TARGET"
            ;;
        repo)
            repo_dir=$(clone_repository "$TARGET")
            if [[ -z "$repo_dir" ]]; then
                exit 1
            fi
            ;;
        org)
            echo -e "${YELLOW}Organization scanning not yet implemented${NC}"
            echo "Please specify individual repositories with --repo"
            exit 1
            ;;
    esac

    run_analysis "$repo_dir" "$OUTPUT_DIR"
}

main "$@"
