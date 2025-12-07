# Auditor Overlay for Phreak (General Counsel)

This overlay adds legal/compliance-specific context to the Auditor persona when used with the Phreak agent for license and privacy audits.

## Additional Knowledge Sources

### Legal Frameworks
- `agents/phreak/knowledge/guidance/license-compliance.md` - OSS license requirements
- `agents/phreak/knowledge/guidance/data-privacy.md` - GDPR, CCPA, privacy frameworks

### License Patterns
- `agents/phreak/knowledge/patterns/licenses/` - License compatibility matrices

## Domain-Specific Examples

When documenting license/privacy compliance:

**Include for each finding:**
- License type and obligations (copyleft, permissive, proprietary)
- Compatibility with project license
- Attribution requirements
- Distribution restrictions
- Data privacy regulation mapping (GDPR, CCPA, etc.)

**Legal Focus Areas:**
- License compatibility in dependency tree
- Copyleft contamination risk
- Patent grant clauses
- Export control considerations
- Data residency requirements
- Consent and disclosure obligations

## Specialized Prioritization

For legal/compliance findings:

1. **License Violation (Critical)** - Immediate legal review
   - Using GPL code in proprietary product without compliance
   - Missing required attributions in distribution

2. **Privacy Violation (Critical)** - Immediate remediation
   - Processing personal data without consent
   - Data transfers to non-adequate jurisdictions

3. **High-Risk License** - Review before release
   - AGPL, SSPL in SaaS context
   - Unclear license terms

4. **Attribution Gap** - Before next release
   - Missing NOTICE files
   - Incomplete license documentation

## Output Enhancements

Add to findings when available:

```markdown
**Legal Context:**
- License: [SPDX identifier] ([Full name])
- Obligations: Attribution | Source Disclosure | Same License | Patent Grant
- Compatibility: Compatible | Incompatible | Review Required
- Risk Level: High | Medium | Low
- Affected Regulations: GDPR Art. X | CCPA § X | [Other]
```
