<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Legal Review Skill

Expert legal review of source code for license compliance and content policy enforcement.

## Capabilities

This skill provides:
- **License compliance analysis** - Detect and analyze open source licenses
- **License compatibility checking** - Evaluate license combinations
- **Content policy enforcement** - Check for inappropriate content
- **Compliance reporting** - Generate legal review reports
- **Risk assessment** - Evaluate legal risks in code
- **Remediation recommendations** - Suggest fixes for violations

## Quick Start

### License Analysis
```bash
# Audit repository licenses
@legal-review audit licenses in this repository

# Check specific dependency
@legal-review what license does package "express" use and is it compatible with MIT?

# Generate attribution file
@legal-review create NOTICE file with all required attributions
```

### Content Policy
```bash
# Check inappropriate content
@legal-review scan for profanity and non-inclusive language

# Specific check
@legal-review is the term "master-slave" acceptable in our code?

# Get recommendations
@legal-review what are the inclusive alternatives for "whitelist/blacklist"?
```

### Compliance
```bash
# Pre-release audit
@legal-review perform comprehensive pre-release legal audit

# Specific compliance
@legal-review check GDPR compliance for PII handling in user_service.py

# M&A due diligence
@legal-review generate legal due diligence report for this codebase
```

## Detailed Examples

### Example 1: Adding a New Dependency

**Scenario**: You want to add a new npm package to your project.

**Prompt**:
```
I want to add the package "axios" to our project. Our project is MIT licensed. Please:
1. Identify axios's license
2. Check if it's compatible with MIT
3. List any attribution requirements
4. Advise if it's safe to use
```

**Expected Analysis**:
- License identification (MIT)
- Compatibility assessment (compatible)
- Attribution requirements (copyright notice, license text)
- Recommendation (approved for use)

### Example 2: License Violation Investigation

**Scenario**: Dependency scan flagged a GPL license.

**Prompt**:
```
Our dependency scan found that package "readline" uses GPL-3.0 license.
Our product is proprietary software distributed to customers.

Please analyze:
1. Is this a license violation?
2. What are the implications of GPL-3.0?
3. What are our options to resolve this?
4. Are there alternative packages we could use?
```

**Expected Analysis**:
- Explain GPL-3.0 copyleft requirements
- Assess violation (likely yes for proprietary distribution)
- Options: remove dependency, find alternative, get commercial license, release as open source
- Suggest MIT/Apache licensed alternatives

### Example 3: Content Policy Review

**Scenario**: Code review found potentially offensive variable names.

**Prompt**:
```
During code review, I found these variable names:
- shitty_hack
- wtf_counter
- master_db / slave_db

Please:
1. Assess if these violate our content policy
2. Provide inclusive/professional alternatives
3. Explain why these should be changed
```

**Expected Analysis**:
- Flag profanity (shitty, wtf)
- Flag non-inclusive language (master/slave)
- Suggest alternatives (poor_workaround, unexpected_counter, primary_db/replica_db)
- Explain professional standards and inclusive language importance

### Example 4: Compliance Audit

**Scenario**: Preparing for a code review and release.

**Prompt**:
```
We're preparing for a major release. Please review our codebase for:
1. License compliance issues
2. Profanity or unprofessional language
3. Non-inclusive terminology
4. Missing license attributions

Generate a compliance checklist and highlight any critical issues.
```

**Expected Analysis**:
- Comprehensive scan summary
- Critical issues list
- Compliance checklist
- Remediation priorities
- Timeline recommendations

## Knowledge Base Access

This skill has access to comprehensive legal review documentation:

### License Compliance
- **Guide**: `rag/legal-review/license-compliance-guide.md`
- **Topics**: License detection, compliance requirements, compatibility matrix, attribution, SPDX, best practices

### Content Policy
- **Guide**: `rag/legal-review/content-policy-guide.md`
- **Topics**: Profanity detection, hate speech, legal risks, PII, secrets, inclusive language, remediation

### Tools & Automation
- **Guide**: `rag/legal-review/legal-review-tools.md`
- **Topics**: ScanCode, TruffleHog, woke, FOSSA, CI/CD integration, custom scripts

### Configuration
- **Config**: `utils/scanners/licenses/config/legal-review-config.json`
- **Contains**: License policies, secret patterns, content rules, compliance settings

## Integration with Tools

### Legal Analyser Tool

Run automated scans with the legal analyser:
```bash
# Full scan (licenses + content policy)
./utils/legal-review/legal-analyser.sh --path .

# License scan only
./utils/legal-review/legal-analyser.sh --path . --licenses-only

# Content policy scan only
./utils/legal-review/legal-analyser.sh --path . --content-only

# GitHub repository scan
./utils/legal-review/legal-analyser.sh --repo owner/repo

# With Claude AI enhancement (Phase 4)
export ANTHROPIC_API_KEY='your-api-key'
./utils/legal-review/legal-analyser.sh --path . --claude
```

**Claude AI Features** (when using --claude):
- License compatibility analysis with RAG context
- Risk assessment and prioritization
- Specific remediation recommendations
- Policy exception evaluation
- Content policy enhancement with context-aware analysis

### Pre-commit Hooks

```bash
# .git/hooks/pre-commit
./utils/legal-review/legal-analyser.sh --quick --staged
```

### CI/CD Integration

```yaml
# .github/workflows/legal-review.yml
- name: Legal Review
  run: |
    ./utils/legal-review/legal-analyser.sh \
      --repo . \
      --fail-on-critical \
      --output legal-report.md
```

## Common Scenarios

| Scenario | Command | Expected Outcome |
|----------|---------|------------------|
| Check new dependency | `@legal-review can I use package X?` | License analysis, compatibility |
| Pre-release audit | `@legal-review audit this repo` | Comprehensive report |
| Content policy | `@legal-review check language` | Profanity/inclusivity check |
| License conflict | `@legal-review GPL + MIT compatible?` | Compatibility analysis |
| M&A due diligence | `@legal-review due diligence report` | Legal risk assessment |

## Policy Configuration

Configure your organization's legal policies in `utils/scanners/licenses/config/legal-review-config.json`:

### Allowed Licenses
```json
{
  "legal_review": {
    "licenses": {
      "allowed": ["MIT", "Apache-2.0", "BSD-3-Clause"],
      "denied": ["GPL-3.0", "AGPL-3.0"]
    }
  }
}
```

### Content Policy
```json
{
  "content_policy": {
    "profanity": {"severity": "medium"},
    "inclusive_language": {"severity": "high"}
  }
}
```

## Best Practices

1. **Regular Audits** - Run legal review before each release
2. **Pre-commit Checks** - Catch issues early in development
3. **Document Decisions** - Record license exemptions and approvals
4. **Educate Team** - Train developers on license implications
5. **Automate** - Integrate into CI/CD pipeline
6. **Update Policies** - Review and update policies quarterly

## Limitations

- **Accuracy**: Automated detection may have false positives/negatives
- **Context**: Some violations require human judgment
- **Legal Advice**: This tool provides guidance, not legal advice
- **Scope**: Cannot detect all possible legal issues
- **Updates**: License landscape changes; keep policies current

## When to Escalate to Legal

Escalate to legal counsel when:
- Uncertainty about license compatibility
- Potential GPL violation in proprietary code
- Export control concerns (ITAR, EAR)
- Suspected intellectual property infringement
- M&A due diligence findings
- Regulatory compliance questions (GDPR, HIPAA)

## Support & Resources

### Internal
- Legal review configuration: `utils/scanners/licenses/config/legal-review-config.json`
- RAG documentation: `rag/legal-review/`
- Build prompts: `prompts/legal-review/`

### External
- SPDX License List: https://spdx.org/licenses/
- tl;drLegal: https://tldrlegal.com/
- Choose a License: https://choosealicense.com/
- Inclusive Naming: https://inclusivenaming.org/

### Getting Help
- GitHub Issues: https://github.com/crashappsec/gibson-powers/issues
- Documentation: See README in `rag/legal-review/`
- Legal Team: Contact legal@company.com for complex cases

## Examples Directory

See `examples/` for:
- License audit examples
- Secret detection scenarios
- Content policy checks
- Compliance reports
- Remediation guides

## Version

**Current Version**: 1.0.0
**Status**: Production Ready
**Last Updated**: 2025-01-01

## License

This skill is part of Gibson Powers and is licensed under GPL-3.0.
See [LICENSE](../../LICENSE) for details.
