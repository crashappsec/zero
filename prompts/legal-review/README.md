<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Legal Review Prompts

Prompts for legal review of source code, including license compliance, secret detection, and content policy enforcement.

## Build Prompts

### [BUILD-LEGAL-ANALYSER.md](BUILD-LEGAL-ANALYSER.md)
Comprehensive prompt for building the legal review analyser tool.

**Use**: Creating the legal analyser from scratch
**Output**: Complete legal-analyser.sh tool

## Operation Prompts

### License Analysis
```bash
# Audit repository licenses
@legal-review audit licenses in this repository against our policy

# Check license compatibility
@legal-review analyze license compatibility for dependencies in package.json

# Generate attribution file
@legal-review create NOTICE file with all required attributions
```

### Secret Detection
```bash
# Scan for secrets
@legal-review scan for hardcoded secrets and credentials

# Check specific file
@legal-review check config.py for exposed secrets

# Git history scan
@legal-review scan git history for accidentally committed secrets
```

### Content Policy
```bash
# Check for inappropriate content
@legal-review scan for profanity and non-inclusive language

# Inclusive language audit
@legal-review review code for non-inclusive technical terms

# Full content review
@legal-review perform comprehensive content policy check
```

### Compliance Reporting
```bash
# Generate compliance report
@legal-review create compliance report for stakeholders

# Pre-release audit
@legal-review perform pre-release legal audit

# M&A due diligence
@legal-review generate legal due diligence report
```

## Quick Reference

### Common Tasks

| Task | Prompt File | Command |
|------|-------------|---------|
| Build analyser | BUILD-LEGAL-ANALYSER.md | Use to create tool |
| License audit | N/A | `./legal-analyser.sh --licenses` |
| Secret scan | N/A | `./legal-analyser.sh --secrets` |
| Full review | N/A | `./legal-analyser.sh --all` |

### Severity Levels

- **Critical**: Hardcoded secrets, severe license violations
- **High**: Denied licenses, PII exposure
- **Medium**: Non-inclusive language, unknown licenses
- **Low**: Style issues, recommendations

## Integration

These prompts work with:
- Legal review skill (`skills/legal-review/`)
- Legal analyser tool (`utils/legal-review/`)
- RAG documentation (`rag/legal-review/`)

## Examples

### Example 1: Pre-Release Audit
```
Please perform a comprehensive legal audit of this repository before release:
1. Check all licenses against our approved list
2. Scan for any hardcoded secrets or credentials
3. Review code for inappropriate content
4. Generate executive summary for legal team
```

### Example 2: New Dependency Review
```
I want to add the package "example-lib" to our project. Please:
1. Identify its license
2. Check compatibility with our MIT license
3. Review any patent or attribution requirements
4. Advise if it's safe to use
```

### Example 3: Incident Response
```
A developer accidentally committed AWS credentials. Please:
1. Scan git history to find all occurrences
2. Identify which commits need remediation
3. Generate list of secrets to rotate
4. Provide step-by-step cleanup instructions
```

## Related Documentation

- [Legal Review RAG](../../rag/legal-review/)
- [Legal Review Skill](../../skills/legal-review/)
- [License Compliance Guide](../../rag/legal-review/license-compliance-guide.md)
- [Content Policy Guide](../../rag/legal-review/content-policy-guide.md)
