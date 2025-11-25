<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Code Security Analyser

AI-powered security code review using Claude to identify vulnerabilities in source code.

## Overview

The Code Security Analyser scans repositories for security vulnerabilities, including:
- Injection attacks (SQL, command, XSS)
- Authentication and authorization flaws
- Cryptographic weaknesses
- Hardcoded secrets and credentials
- Data exposure risks
- Input validation issues
- Security misconfigurations

## Installation

### Prerequisites

- Bash 4.0+
- jq
- Git
- Anthropic API key

### Setup

1. Set your Anthropic API key:
```bash
export ANTHROPIC_API_KEY='your-api-key'
```

2. Make the script executable:
```bash
chmod +x code-security-analyser.sh
```

## Usage

```bash
./code-security-analyser.sh [OPTIONS] [TARGET]
```

### Targets

| Option | Description |
|--------|-------------|
| `--repo OWNER/REPO` | Scan a GitHub repository |
| `--org ORG_NAME` | Scan all repos in organization |
| `--local PATH` | Scan a local directory |
| (no target) | Scan current directory |

### Options

| Option | Description | Default |
|--------|-------------|---------|
| `--output DIR` | Output directory | `./code-security-reports` |
| `--format FORMAT` | markdown, json, sarif | `markdown` |
| `--severity LEVEL` | Minimum: low, medium, high, critical | `low` |
| `--fail-on LEVEL` | Exit non-zero if findings >= level | (none) |
| `--categories LIST` | Categories to check | `all` |
| `--exclude PATTERNS` | Glob patterns to exclude | `node_modules/**,...` |
| `--max-files N` | Maximum files to scan | `500` |
| `--supply-chain` | Include dependency analysis | (disabled) |
| `--no-claude` | Disable AI analysis | (enabled) |

### Examples

```bash
# Scan a GitHub repository
./code-security-analyser.sh --repo owner/repo

# Scan local project
./code-security-analyser.sh --local /path/to/project

# Scan with supply chain analysis
./code-security-analyser.sh --local . --supply-chain

# Filter by severity
./code-security-analyser.sh --repo owner/repo --severity high

# CI/CD mode - fail on critical findings
./code-security-analyser.sh --repo owner/repo --fail-on critical --format sarif

# Scan specific categories
./code-security-analyser.sh --local . --categories injection,secrets,auth
```

## Output

### Markdown Report

```markdown
# Code Security Analysis Report

**Target**: owner/repo
**Date**: 2025-01-15 10:30:00

## Summary

| Severity | Count |
|----------|-------|
| 🔴 Critical | 2 |
| 🟠 High | 5 |
| 🟡 Medium | 8 |
| 🟢 Low | 3 |

## Findings

### 🔴 Critical Severity

#### SQL Injection
**File**: src/db.py:42
**CWE**: CWE-89
...
```

### JSON Report

```json
{
  "metadata": {
    "target": "owner/repo",
    "timestamp": "2025-01-15T10:30:00Z",
    "tool": "Gibson Powers Code Security Analyser"
  },
  "summary": {
    "total": 18,
    "critical": 2,
    "high": 5,
    "medium": 8,
    "low": 3
  },
  "findings": [...]
}
```

### SARIF Report

SARIF format for GitHub code scanning integration.

## Categories

| Category | Description | Examples |
|----------|-------------|----------|
| `injection` | Injection vulnerabilities | SQL, command, XSS, LDAP |
| `auth` | Authentication issues | Broken auth, weak passwords |
| `crypto` | Cryptographic weaknesses | Weak algorithms, hardcoded keys |
| `exposure` | Data exposure | Sensitive data in logs |
| `validation` | Input validation | Path traversal, SSRF |
| `secrets` | Hardcoded secrets | API keys, passwords |
| `config` | Misconfigurations | Debug mode, CORS |

## CI/CD Integration

### GitHub Actions

```yaml
- name: Security Scan
  run: |
    ./utils/code-security/code-security-analyser.sh \
      --local . \
      --format sarif \
      --fail-on high

- name: Upload SARIF
  uses: github/codeql-action/upload-sarif@v3
  with:
    sarif_file: code-security-reports/security-report.sarif
```

### GitLab CI

```yaml
security-scan:
  script:
    - ./utils/code-security/code-security-analyser.sh --local . --fail-on critical
  artifacts:
    reports:
      sast: code-security-reports/security-report.json
```

## Related

- [Skill](../../skills/code-security/) - Claude Code skill definition
- [Prompts](../../prompts/code-security/) - Analysis prompts
- [RAG](../../rag/code-security/) - Knowledge base
