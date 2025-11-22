<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Legal Review Prompts

Prompts for legal review of source code, including license compliance and content policy enforcement.

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
| License audit | N/A | `./legal-analyser.sh --path . --licenses-only` |
| Content policy | N/A | `./legal-analyser.sh --path . --content-only` |
| Full review | N/A | `./legal-analyser.sh --path .` |
| With Claude AI | N/A | `./legal-analyser.sh --path . --claude` |

### Severity Levels

- **Critical**: Severe license violations (GPL in proprietary code)
- **High**: Denied licenses, copyleft conflicts
- **Medium**: Non-inclusive language, unknown licenses
- **Low**: Profanity, style issues, recommendations

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
2. Review code for inappropriate content
3. Check for non-inclusive language
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

## Related Documentation

- [Legal Review RAG](../../rag/legal-review/)
- [Legal Review Skill](../../skills/legal-review/)
- [License Compliance Guide](../../rag/legal-review/license-compliance-guide.md)
- [Content Policy Guide](../../rag/legal-review/content-policy-guide.md)
