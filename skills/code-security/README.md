<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Code Security Analyser Skill

AI-powered security code review using Claude to identify vulnerabilities in source code.

## Overview

The Code Security Analyser skill enables comprehensive security analysis of source code repositories, identifying vulnerabilities like injection attacks, authentication flaws, cryptographic weaknesses, and hardcoded secrets.

## Usage

### Via Skill

Use the code-security skill in Claude Code for interactive security analysis:

```
Use the code-security skill to analyse this repository for security vulnerabilities.
```

### Via CLI

```bash
# Scan a GitHub repository
./utils/code-security/code-security-analyser.sh --repo owner/repo

# Scan local directory
./utils/code-security/code-security-analyser.sh --local /path/to/project

# Scan with supply chain analysis
./utils/code-security/code-security-analyser.sh --local . --supply-chain

# Filter by severity
./utils/code-security/code-security-analyser.sh --repo owner/repo --severity high

# CI/CD mode (fail on critical)
./utils/code-security/code-security-analyser.sh --repo owner/repo --fail-on critical --format sarif
```

## Capabilities

### Source Code Analysis

Detects vulnerabilities in first-party code:

| Category | Examples |
|----------|----------|
| **Injection** | SQL, Command, XSS, LDAP, XPath |
| **Authentication** | Broken auth, missing checks, session issues |
| **Authorization** | Access control flaws, IDOR, privilege escalation |
| **Cryptography** | Weak algorithms, hardcoded keys, insecure random |
| **Data Exposure** | Sensitive data in logs, information disclosure |
| **Input Validation** | Path traversal, SSRF, open redirects |
| **Secrets** | Hardcoded credentials, API keys, tokens |
| **Configuration** | Debug mode, insecure defaults, CORS issues |

### Supply Chain Integration

Optionally integrates with the Supply Chain Scanner for:
- Dependency vulnerabilities (CVEs)
- Package health analysis
- License compliance
- Provenance verification

### Output Formats

- **Markdown** - Human-readable reports
- **JSON** - Structured data for processing
- **SARIF** - GitHub code scanning integration

## Supported Languages

- Python
- JavaScript/TypeScript
- Java
- Go
- Ruby
- PHP
- C/C++
- C#
- Swift
- Kotlin
- Rust
- Scala
- Shell/Bash

## Examples

See the `examples/` directory for sample security reports.

## Related

- [Prompts](../../prompts/code-security/) - Claude prompts for security analysis
- [RAG Knowledge Base](../../rag/code-security/) - Vulnerability patterns and remediation
- [Analyser Script](../../utils/code-security/) - CLI tool

## Requirements

- Anthropic API key (`ANTHROPIC_API_KEY` environment variable)
- jq (JSON processing)
- Git (repository cloning)
