<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Code Security Analyser - Implementation Plan

## Overview

Build an AI-powered code security review system using Claude to identify vulnerabilities, security weaknesses, and potential exploits in source code. This analyser follows the same architecture as other Gibson Powers analysers (supply-chain, technology-identification, etc.) - clone repository and perform comprehensive analysis.

## Architecture Alignment

### Same Pattern as Other Analysers

The Code Security Analyser will follow the established Gibson Powers structure:

```
# Utilities (shell scripts)
utils/code-security/
├── code-security-analyser.sh      # Main analyser script
├── lib/
│   ├── file-scanner.sh            # Scan files for security patterns
│   ├── context-builder.sh         # Build code context for Claude
│   ├── severity-classifier.sh     # Classify finding severity
│   └── report-generator.sh        # Generate markdown/JSON reports
└── tests/
    └── run-tests.sh

# Prompts (Claude prompts - follows existing /prompts structure)
prompts/code-security/
├── README.md
├── analysis/
│   ├── security-review.md         # Main security analysis prompt
│   └── secrets-detection.md       # Secrets/credentials detection
├── reporting/
│   └── security-report.md         # Report generation prompt
└── remediation/
    └── fix-recommendations.md     # Remediation guidance prompt

# Skills (Claude Code skills - follows existing /skills structure)
skills/code-security/
├── code-security.skill            # Main skill definition
├── README.md
└── examples/
    ├── example-security-report.md
    └── example-vulnerability-finding.md

# RAG Knowledge Base (follows existing /rag structure)
rag/code-security/
├── vulnerability-patterns/        # Known vulnerability patterns
├── framework-security/            # Framework-specific security
├── remediation-examples/          # Fix examples
└── standards/                     # OWASP, CWE references
```

### Key Design Decisions

| Decision | Approach | Rationale |
|----------|----------|-----------|
| **Scanning Model** | Clone & analyse full repo | Same as other analysers, comprehensive analysis |
| **AI Integration** | Claude-only | Security analysis requires sophisticated reasoning |
| **Cloning** | Use `github_clone_repository()` from `utils/lib/github.sh` | Reuse existing code |
| **Config Loading** | Use `utils/lib/config.sh` and `config-loader.sh` | Consistent with other analysers |
| **Output Formats** | Markdown + JSON | Same as other analysers |
| **RAG Integration** | Load from `rag/code-security/` | Same pattern as certificate-analyser |
| **Supply Chain** | Call existing `supply-chain-scanner.sh` | Avoid duplication, inherit improvements |

### Integration with Supply Chain Scanner

The Code Security Analyser will **call the existing supply chain scanner** for dependency/vulnerability analysis rather than duplicating that functionality:

```bash
# Code Security Analyser calls supply chain scanner internally
run_supply_chain_analysis() {
    local repo_dir="$1"
    local output_dir="$2"

    "$UTILS_ROOT/supply-chain/supply-chain-scanner.sh" \
        --local "$repo_dir" \
        --vulnerability \
        --package-health \
        --output "$output_dir/supply-chain" \
        --claude
}
```

**Benefits:**
- Single source of truth for dependency analysis
- Automatic inheritance of supply chain scanner improvements
- Consistent vulnerability detection across tools
- No code duplication

**Analysis Split:**
| Analysis Type | Handled By |
|--------------|------------|
| Source code vulnerabilities (injection, XSS, etc.) | Code Security Analyser |
| Dependency vulnerabilities (CVEs) | Supply Chain Scanner |
| Package health & provenance | Supply Chain Scanner |
| Secrets detection | Code Security Analyser |
| License compliance | Supply Chain Scanner (legal module) |

## Reference Implementation

Based on [Anthropic's claude-code-security-review](https://github.com/anthropics/claude-code-security-review) methodology:

- **Security categories**: Injection, auth, crypto, data exposure, input validation, business logic
- **False positive filtering**: AI-powered noise reduction
- **Severity classification**: Critical/High/Medium/Low with confidence scores
- **Remediation guidance**: Specific fix recommendations

Key difference: We scan entire repositories rather than PR diffs.

## Implementation Plan

### Phase 1: Core Infrastructure (Week 1-2)

#### 1.1 Main Analyser Script

Create `utils/code-security/code-security-analyser.sh` following supply-chain-scanner.sh pattern:

```bash
#!/bin/bash
# Code Security Analyser
# Usage: ./code-security-analyser.sh [OPTIONS] [TARGETS...]

# Standard header (same as other analysers)
set -e
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
UTILS_ROOT="$(dirname "$SCRIPT_DIR")"
REPO_ROOT="$(dirname "$UTILS_ROOT")"

# Load shared libraries
source "$UTILS_ROOT/lib/config.sh"
source "$UTILS_ROOT/lib/github.sh"
if [[ -f "$UTILS_ROOT/lib/config-loader.sh" ]]; then
    source "$UTILS_ROOT/lib/config-loader.sh"
fi

# Load .env
if [[ -f "$REPO_ROOT/.env" ]]; then
    set -a && source "$REPO_ROOT/.env" && set +a
fi
```

#### 1.2 CLI Interface

```bash
./code-security-analyser.sh [OPTIONS] [TARGETS...]

TARGETS:
    --repo OWNER/REPO       Scan specific repository
    --org ORG_NAME          Scan all repos in organization
    --local PATH            Scan local directory
    (No target = scan current directory)

OPTIONS:
    --output DIR, -o DIR    Output directory for reports
    --format FORMAT         Output format: markdown|json|sarif (default: markdown)
    --severity LEVEL        Minimum severity: low|medium|high|critical
    --fail-on LEVEL         Exit non-zero if findings >= level
    --categories LIST       Comma-separated: injection,auth,crypto,exposure,validation,secrets
    --exclude PATTERNS      Glob patterns to exclude (e.g., "test/**,vendor/**")
    --max-files N           Max files to analyse (default: 500)
    --supply-chain          Include supply chain analysis (calls supply-chain-scanner.sh)
    --no-supply-chain       Skip supply chain analysis (code-only)
    --claude                Enable Claude AI analysis (default if API key set)
    --no-claude             Disable Claude AI (pattern-only analysis)
    -h, --help              Show help

EXAMPLES:
    # Scan a GitHub repository (code + supply chain)
    ./code-security-analyser.sh --repo owner/repo --supply-chain

    # Scan local project (code only)
    ./code-security-analyser.sh --local /path/to/project

    # Scan with minimum severity
    ./code-security-analyser.sh --repo owner/repo --severity high

    # Full security scan with all modules
    ./code-security-analyser.sh --repo owner/repo --supply-chain --format sarif

    # CI/CD mode - fail on critical findings
    ./code-security-analyser.sh --repo owner/repo --fail-on critical --format sarif
```

#### 1.3 Core Functions

```bash
# Clone repository (reuse existing)
clone_and_prepare() {
    local target="$1"
    if [[ -d "$target" ]]; then
        REPO_DIR="$target"
    else
        TEMP_DIR=$(mktemp -d)
        github_clone_repository "$target" "$TEMP_DIR" --depth 1
        REPO_DIR="$TEMP_DIR"
    fi
}

# Identify security-relevant files
identify_target_files() {
    local repo_dir="$1"
    # Find source files by extension
    # Exclude: node_modules, vendor, .git, tests, etc.
}

# Build context for Claude analysis
build_analysis_context() {
    local file="$1"
    # Extract file content + surrounding context
    # Include imports, function signatures, etc.
}

# Run Claude security analysis
analyse_file_with_claude() {
    local file="$1"
    local context="$2"
    # Call Claude API with security prompt + RAG content
}

# Run supply chain analysis (delegate to existing scanner)
run_supply_chain_analysis() {
    local repo_dir="$1"
    local output_dir="$2"

    echo -e "${BLUE}Running supply chain analysis...${NC}"

    # Call existing supply chain scanner
    "$UTILS_ROOT/supply-chain/supply-chain-scanner.sh" \
        --local "$repo_dir" \
        --vulnerability \
        --package-health \
        --output "$output_dir/supply-chain" \
        ${USE_CLAUDE:+--claude}

    # Results will be in $output_dir/supply-chain/
}

# Merge findings from code analysis and supply chain
merge_security_findings() {
    local code_findings="$1"
    local supply_chain_dir="$2"
    local output_file="$3"

    # Combine:
    # - Code security findings (injection, XSS, etc.)
    # - Vulnerability findings from supply chain scanner
    # - Package health issues
    # Into unified security report
}

# Generate report
generate_report() {
    local findings="$1"
    local format="$2"
    # Output markdown, JSON, or SARIF
}
```

### Phase 2: Claude Integration (Week 2-3)

#### 2.1 Security Analysis Prompt

Create `prompts/code-security/analysis/security-review.md` (follows existing prompts structure):

```markdown
You are a security expert analysing source code for vulnerabilities.

## Your Task

Analyse the following code for security issues. For each finding:
1. Identify the specific vulnerability
2. Explain why it's a security risk
3. Rate the severity (Critical/High/Medium/Low)
4. Provide remediation guidance with code examples

## Vulnerability Categories

Check for these security issues:

### Injection
- SQL injection (string concatenation in queries)
- Command injection (shell commands with user input)
- LDAP/XPath injection
- Expression language injection

### Authentication & Authorization
- Hardcoded credentials
- Weak password handling
- Missing authentication checks
- Broken access control
- Session management issues

### Data Exposure
- Sensitive data in logs
- PII exposure
- Information disclosure in errors
- Insecure data storage

### Cryptography
- Weak algorithms (MD5, SHA1, DES)
- Hardcoded keys/IVs
- Insecure random generation
- Missing encryption

### Input Validation
- Cross-Site Scripting (XSS)
- Path traversal
- Open redirects
- SSRF vulnerabilities
- Regex DoS (ReDoS)

### Business Logic
- Race conditions
- TOCTOU vulnerabilities
- Integer overflow/underflow
- Unsafe deserialization

## Output Format

Return findings as JSON array:
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
    "remediation": "Use parameterized queries: cursor.execute(\"SELECT * FROM users WHERE id=?\", (user_id,))",
    "exploitation": "Attacker can inject SQL to extract or modify database contents"
  }
]
```

## Code to Analyse

{code_content}
```

#### 2.2 RAG Content Loading

Follow certificate-analyser pattern for RAG integration:

```bash
# Load RAG content for security analysis
load_security_rag() {
    local query="$1"
    if has_rag_content; then
        get_rag_content_smart "code-security" "$query"
    fi
}
```

### Phase 3: RAG Knowledge Base (Week 3-4)

#### 3.1 Directory Structure

```
rag/code-security/
├── vulnerability-patterns/
│   ├── injection-patterns.md
│   ├── authentication-flaws.md
│   ├── cryptographic-weaknesses.md
│   ├── data-exposure-risks.md
│   ├── input-validation-issues.md
│   └── business-logic-flaws.md
├── framework-security/
│   ├── react-security.md
│   ├── django-security.md
│   ├── express-security.md
│   ├── rails-security.md
│   ├── spring-security.md
│   └── flask-security.md
├── remediation-examples/
│   ├── sql-injection-fixes.md
│   ├── xss-prevention.md
│   ├── auth-best-practices.md
│   └── crypto-usage.md
└── standards/
    ├── owasp-top-10-2021.md
    ├── cwe-top-25.md
    └── secure-coding-guidelines.md
```

#### 3.2 Sample RAG Content

`rag/code-security/vulnerability-patterns/injection-patterns.md`:

```markdown
# Injection Vulnerability Patterns

## SQL Injection

### Dangerous Patterns
- String concatenation in queries: `"SELECT * FROM users WHERE id=" + userId`
- String formatting: `f"SELECT * FROM users WHERE id={user_id}"`
- Template strings in queries

### Safe Patterns
- Parameterized queries: `cursor.execute("SELECT * FROM users WHERE id=?", (user_id,))`
- ORM methods: `User.objects.filter(id=user_id)`
- Prepared statements

### Language-Specific Examples

#### Python
```python
# VULNERABLE
cursor.execute(f"SELECT * FROM users WHERE name='{name}'")

# SAFE
cursor.execute("SELECT * FROM users WHERE name=%s", (name,))
```

#### JavaScript
```javascript
// VULNERABLE
db.query(`SELECT * FROM users WHERE id=${userId}`)

// SAFE
db.query('SELECT * FROM users WHERE id=$1', [userId])
```
```

### Phase 4: Testing & Documentation (Week 4-5)

#### 4.1 Test Suite

Create test cases with intentionally vulnerable code samples:

```
utils/code-security/tests/
├── test-samples/
│   ├── sql-injection.py
│   ├── xss-vulnerable.js
│   ├── command-injection.sh
│   ├── hardcoded-secrets.py
│   └── weak-crypto.java
├── expected-results/
│   └── *.json
└── run-tests.sh
```

#### 4.2 Documentation

- README.md with usage examples
- CATEGORIES.md explaining each vulnerability type
- INTEGRATION.md for CI/CD setup

### Phase 5: Skill & Slash Command (Week 5-6)

#### 5.1 Skill Definition

Create `skills/code-security/code-security.skill` (follows existing skills structure):

```yaml
name: code-security
description: AI-powered security code review for repositories
version: 1.0.0

# Reference the prompts
prompts:
  - prompts/code-security/analysis/security-review.md
  - prompts/code-security/analysis/secrets-detection.md
  - prompts/code-security/reporting/security-report.md

# Reference the utility
utility: utils/code-security/code-security-analyser.sh

# RAG knowledge base
rag: rag/code-security/

capabilities:
  - Analyse source code for security vulnerabilities
  - Detect hardcoded secrets and credentials
  - Identify injection, XSS, and authentication flaws
  - Integrate with supply chain scanner for dependency analysis
  - Generate security reports with remediation guidance
```

#### 5.2 Skill README

Create `skills/code-security/README.md`:

```markdown
# Code Security Analyser Skill

AI-powered security code review using Claude to identify vulnerabilities in source code.

## Usage

### Via Skill
Use the code-security skill for comprehensive security analysis.

### Via CLI
\`\`\`bash
./utils/code-security/code-security-analyser.sh --repo owner/repo
\`\`\`

## Capabilities

- **Source Code Analysis**: Injection, XSS, auth flaws, crypto weaknesses
- **Secrets Detection**: Hardcoded credentials, API keys, tokens
- **Supply Chain Integration**: Delegates to supply-chain-scanner for dependencies
- **Remediation Guidance**: Specific fixes with code examples

## Examples

See `examples/` for sample security reports.
```

## Deliverables

| Deliverable | Description | Path |
|-------------|-------------|------|
| Main Analyser | Shell script following standard pattern | `utils/code-security/code-security-analyser.sh` |
| Library Functions | Reusable security scanning functions | `utils/code-security/lib/*.sh` |
| Prompts | Security analysis prompts | `prompts/code-security/` |
| Skill | Skill definition and docs | `skills/code-security/` |
| RAG Knowledge Base | Vulnerability patterns and remediation | `rag/code-security/` |
| Tests | Test suite with vulnerable samples | `utils/code-security/tests/` |

## Success Metrics

- Detection rate for known vulnerability patterns (>80%)
- False positive rate (<20%)
- Analysis performance (<5 min for typical repository)
- User satisfaction with finding quality

## Dependencies

- Anthropic API (ANTHROPIC_API_KEY) - **Required**
- jq for JSON processing
- Git for repository cloning
- Bash 4.0+

## Timeline

| Week | Deliverables |
|------|-------------|
| 1-2 | Core analyser script, CLI interface, basic file scanning |
| 2-3 | Claude integration, security prompts, analysis pipeline |
| 3-4 | RAG knowledge base (vulnerability patterns, remediation) |
| 4-5 | Testing, documentation, CI/CD examples |
| 5-6 | Slash command, polish, release |

## Next Steps

1. ✅ Create feature branch: `feature/code-security`
2. ✅ Create implementation plan
3. Set up directory structure
4. Implement core analyser script (following supply-chain-scanner.sh pattern)
5. Develop Claude security prompt
6. Build RAG knowledge base
7. Create slash command
8. Write tests and documentation
