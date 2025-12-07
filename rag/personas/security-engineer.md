# Security Engineer Persona

## Role Description

A security engineer focused on technical vulnerability analysis, risk assessment, and remediation prioritization. This persona needs deep technical details, exploit context, and actionable remediation steps.

**What they care about:**
- Attack vectors and exploitability
- Severity scores and risk metrics
- Compensating controls
- Remediation SLAs

## Output Style

- **Tone:** Technical, direct, action-oriented
- **Detail Level:** High - include identifiers, severity scores, exploit context
- **Format:** Structured findings with clear severity and remediation
- **Prioritization:** Risk-based using severity, exploitability, and active exploitation status

## Output Template

```markdown
## Security Analysis

### Executive Summary

**Risk Level:** Critical/High/Medium/Low
**Total Findings:** X (Y critical, Z high)
**Immediate Action Required:** Yes/No

### Findings

#### [SEVERITY] Finding Title

**Risk Score:** [Severity metric] | [Exploitability score]
**Affected:** [Component/package/file/system]
**Confidence:** High/Medium/Low
**Status:** New/Known/In Remediation

**Attack Vector:**
- [Technical description of how this can be exploited]
- [Attack prerequisites and complexity]

**Impact:**
- [Confidentiality/Integrity/Availability impact]
- [Specific impact to this system/application]
- [Potential blast radius]

**Evidence:**
```
[Relevant code, config, or scan output]
```

**Remediation:**
[Specific steps or commands to remediate]

**Timeline:** [SLA based on severity]

**Compensating Controls (if remediation blocked):**
- [Alternative mitigations]
- [Detection mechanisms]

---
```

## Prioritization Framework

1. **Critical + Actively Exploited** - Immediate (within hours)
2. **Critical + High Exploitability** - Within 24 hours
3. **High + Network Accessible + No Auth** - Within 7 days
4. **High + Authentication Required** - Within 14 days
5. **Medium** - Within 30 days
6. **Low** - Within 90 days

## Key Questions to Answer

- What can an attacker do with this finding?
- Is this being actively exploited in the wild?
- What's the fastest path to remediation?
- Are there compensating controls if we can't remediate immediately?
- What's the blast radius if exploited?
- What detection mechanisms should we implement?
