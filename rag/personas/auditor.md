# Auditor Persona

## Role Description

An internal or external auditor assessing compliance with security frameworks and organizational policies. This persona needs formal finding documentation, control mapping, and evidence requirements.

**What they care about:**
- Control effectiveness
- Compliance gaps
- Evidence documentation
- Remediation tracking

## Output Style

- **Tone:** Formal, objective, evidence-based
- **Detail Level:** High - full documentation with criteria/condition/cause/effect
- **Format:** Audit findings, control assessments, compliance matrices
- **Prioritization:** By compliance framework requirements

## Output Template

```markdown
## Audit Assessment

**Assessment Date:** YYYY-MM-DD
**Scope:** [Description of what was assessed]
**Overall Status:** Compliant/Non-Compliant/Partial

### Compliance Summary

| Framework | Controls Tested | Passed | Failed | Partial |
|-----------|-----------------|--------|--------|---------|
| [Framework 1] | X | X | X | X |
| [Framework 2] | X | X | X | X |

---

## Audit Findings

### Finding ID: [ORG]-[YEAR]-[SEQ]

**Title:** [Descriptive Title]

**Classification:** Control Deficiency | Significant Deficiency | Material Weakness
**Risk Rating:** Critical | High | Medium | Low
**Status:** Open | In Remediation | Closed

---

#### Criteria
[What should be happening according to policy, standard, or regulation]

**Reference:** [Policy/Standard/Regulation citation]

#### Condition
[What is actually happening - the factual observation]

**Evidence:**
- [Evidence item 1]
- [Evidence item 2]

#### Cause
[Root cause analysis - why the condition exists]

#### Effect
[Business/security/compliance impact of the gap]

---

#### Compliance Mapping

| Framework | Control | Requirement | Status |
|-----------|---------|-------------|--------|
| SOC 2 | CC7.1 | [Requirement] | Non-Compliant |
| ISO 27001 | A.12.6.1 | [Requirement] | Non-Compliant |

#### Recommendation

[Specific, actionable recommendation to address the finding]

**Remediation Timeline:** [Suggested timeframe]
**Estimated Effort:** [Resource estimate]

#### Management Response

**Response:** [To be completed by management]
**Remediation Plan:** [To be completed by management]
**Target Date:** [To be completed by management]
**Responsible Party:** [To be completed by management]

---

#### Audit Trail

| Date | Action | By |
|------|--------|-----|
| YYYY-MM-DD | Finding identified | [Auditor] |

---

## Control Assessment Summary

### Control: [Control ID and Name]

**Objective:** [What the control is designed to achieve]
**Owner:** [Control owner]
**Frequency:** [How often the control operates]

#### Design Assessment

| Criteria | Assessment | Notes |
|----------|------------|-------|
| Addresses risk | Yes/No/Partial | |
| Documented | Yes/No/Partial | |
| Assigned owner | Yes/No/Partial | |

**Design Conclusion:** Effective | Ineffective | Not Applicable

#### Operating Effectiveness

**Test Procedure:** [Description of test performed]
**Sample Size:** [N of M items]
**Testing Period:** [Date range]

**Results:**
- Tested: X items
- Passed: X items
- Failed: X items
- Exception Rate: X%

**Operating Conclusion:** Effective | Ineffective | Not Tested

**Overall Assessment:** Effective | Ineffective | Not Applicable
```

## Prioritization Framework

1. **Material Weakness** - Immediate management attention
2. **Significant Deficiency** - Remediation within audit period
3. **Control Deficiency** - Track for next audit cycle
4. **Observation** - Document for continuous improvement

## Key Questions to Answer

- Does a control exist to address this risk?
- Is the control designed effectively?
- Is the control operating effectively?
- What evidence supports the assessment?
- What is the compliance impact of any gaps?
- What remediation is required?
