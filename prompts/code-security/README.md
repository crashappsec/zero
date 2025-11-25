<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Code Security Prompts

Prompts for AI-powered security code analysis using Claude.

## Structure

```
prompts/code-security/
├── analysis/
│   ├── security-review.md      # Main security analysis prompt
│   └── secrets-detection.md    # Secrets/credentials detection
├── reporting/
│   └── security-report.md      # Report generation prompt
└── remediation/
    └── fix-recommendations.md  # Remediation guidance prompt
```

## Usage

These prompts are used by:
- `utils/code-security/code-security-analyser.sh` - Main analyser script
- `skills/code-security/` - Claude Code skill

## Categories

The security analysis covers these vulnerability categories:

| Category | Description |
|----------|-------------|
| **injection** | SQL, command, LDAP, XPath injection |
| **auth** | Authentication and authorization flaws |
| **crypto** | Cryptographic weaknesses |
| **exposure** | Data exposure and information leakage |
| **validation** | Input validation (XSS, path traversal, SSRF) |
| **secrets** | Hardcoded secrets and credentials |
| **config** | Security misconfigurations |

## Related

- [RAG Knowledge Base](../../rag/code-security/) - Vulnerability patterns and remediation
- [Skill Definition](../../skills/code-security/) - Claude Code integration
- [Analyser Script](../../utils/code-security/) - CLI tool
