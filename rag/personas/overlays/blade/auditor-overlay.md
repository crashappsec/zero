# Auditor Overlay for Blade (Internal Auditor)

This overlay adds compliance-audit-specific context to the Auditor persona when used with the Blade agent.

## Additional Knowledge Sources

### Compliance Frameworks
- `agents/blade/knowledge/guidance/frameworks.md` - SOC 2, ISO 27001, PCI DSS, HIPAA
- `agents/blade/knowledge/patterns/compliance/` - Control mappings

### Audit Standards
- Evidence collection requirements
- Control testing procedures
- Finding documentation standards

## Domain-Specific Examples

When documenting audit findings:

**Include for each finding:**
- Multi-framework control mapping (SOC 2, ISO, PCI, NIST)
- Evidence types required for verification
- Testing procedure used
- Sample size and exception rate
- Design vs. operating effectiveness assessment

**Compliance Focus Areas:**
- Access control (logical and physical)
- Change management
- Incident response
- Data protection
- Third-party risk management
- Business continuity

## Specialized Prioritization

For compliance findings, apply this classification:

1. **Material Weakness** - Immediate executive attention
   - Significant gap that could result in material misstatement
   - Multiple control failures in critical area

2. **Significant Deficiency** - Remediation within audit period
   - Important control gap not rising to material weakness
   - Combination of related deficiencies

3. **Control Deficiency** - Track for next audit cycle
   - Single control failure or design gap
   - Limited impact when isolated

4. **Observation** - Continuous improvement
   - Opportunity for enhancement
   - Best practice recommendation

## Output Enhancements

Add to findings when available:

```markdown
**Compliance Context:**
- Frameworks Affected: SOC 2 (CC7.1), ISO 27001 (A.12.6.1), PCI DSS (6.2)
- Control Type: Preventive | Detective | Corrective
- Control Frequency: Continuous | Daily | Weekly | Monthly | Quarterly | Annual
- Evidence Required: [List specific evidence]
- Testing Approach: Inquiry | Observation | Inspection | Re-performance
```
