# Compliance Assessment Prompt

## Context
You are conducting a compliance assessment against one or more frameworks (SOC 2, ISO 27001, NIST CSF, PCI-DSS, HIPAA).

## Assessment Process

### 1. Scope Definition
- Which framework(s) apply?
- What systems/processes are in scope?
- What is the assessment period?

### 2. Control Identification
- Map applicable controls from framework
- Identify control owners
- Determine testing approach

### 3. Evidence Collection
- Request relevant documentation
- Gather system configurations
- Collect logs and reports

### 4. Control Testing
- Test design effectiveness
- Test operating effectiveness (for Type II)
- Document test results

### 5. Finding Documentation
- Rate findings by severity
- Identify root causes
- Recommend remediation

## Output Format

```markdown
## Compliance Assessment Report

### Executive Summary
Brief overview of assessment scope, methodology, and key findings.

### Assessment Scope

| Attribute | Value |
|-----------|-------|
| Framework(s) | |
| Period | |
| Systems in Scope | |
| Assessment Type | |

### Control Assessment Summary

| Control Area | Controls Tested | Passed | Failed | N/A |
|--------------|-----------------|--------|--------|-----|
| Access Management | | | | |
| Change Management | | | | |
| Operations | | | | |
| Security | | | | |

### Findings

#### Finding 1: [Title]

| Attribute | Value |
|-----------|-------|
| Control Reference | CC6.1, ISO A.9.1.1 |
| Severity | Critical/High/Medium/Low |
| Status | Open/Remediated |

**Observation:**
Description of the finding.

**Risk:**
Impact if not addressed.

**Evidence:**
- Evidence item 1
- Evidence item 2

**Recommendation:**
Specific remediation steps.

**Management Response:**
[To be completed by management]

**Target Remediation Date:**
[To be completed by management]

---

### Remediation Tracking

| Finding | Severity | Owner | Target Date | Status |
|---------|----------|-------|-------------|--------|
| | | | | |

### Conclusion

Summary of overall compliance posture and key recommendations.

### Appendix: Evidence Log

| Evidence ID | Description | Date Obtained | Source |
|-------------|-------------|---------------|--------|
| | | | |
```

## Example Output

```markdown
## Compliance Assessment Report

### Executive Summary
This SOC 2 Type II assessment evaluated XYZ Corp's controls over the period January 1 - December 31, 2024. The assessment identified **2 high-severity** and **4 medium-severity** findings. Key concerns include incomplete access reviews and missing change approval documentation. Overall, the control environment is **maturing** but requires attention to access management controls.

### Assessment Scope

| Attribute | Value |
|-----------|-------|
| Framework(s) | SOC 2 Type II (Security, Availability) |
| Period | Jan 1 - Dec 31, 2024 |
| Systems in Scope | Production AWS environment, GitHub, Okta |
| Assessment Type | Annual compliance audit |

### Control Assessment Summary

| Control Area | Controls Tested | Passed | Failed | N/A |
|--------------|-----------------|--------|--------|-----|
| Access Management | 8 | 5 | 3 | 0 |
| Change Management | 6 | 4 | 2 | 0 |
| Operations | 5 | 4 | 1 | 0 |
| Security | 7 | 7 | 0 | 0 |

### Findings

#### Finding 1: Incomplete Quarterly Access Reviews

| Attribute | Value |
|-----------|-------|
| Control Reference | CC6.2, CC6.3 |
| Severity | High |
| Status | Open |

**Observation:**
Quarterly access reviews were not completed for Q2 and Q3 2024. Q1 and Q4 reviews were completed but Q2 showed only 60% completion rate with several manager attestations missing.

**Risk:**
Inappropriate access may persist undetected, increasing risk of unauthorized data access or actions.

**Evidence:**
- Access review completion report (E-001)
- Q2 review with missing attestations (E-002)
- Policy requiring quarterly reviews (E-003)

**Recommendation:**
1. Complete retroactive review for Q2/Q3 immediately
2. Implement automated reminders for reviewers
3. Establish escalation process for incomplete reviews
4. Consider access review tooling to improve completion rates

**Management Response:**
[To be completed]

**Target Remediation Date:**
[To be completed]

---

#### Finding 2: Missing Change Approvals

| Attribute | Value |
|-----------|-------|
| Control Reference | CC8.1 |
| Severity | High |
| Status | Open |

**Observation:**
Of 25 changes sampled, 5 (20%) lacked documented approval before deployment. In 3 cases, approval was obtained retroactively; in 2 cases, no approval was documented.

**Risk:**
Unauthorized or untested changes may be deployed to production, potentially causing outages or security vulnerabilities.

**Evidence:**
- Change sample list with test results (E-004)
- Change tickets lacking approval (E-005, E-006)

**Recommendation:**
1. Implement technical control preventing deployment without approval
2. Enforce approval requirement in CI/CD pipeline
3. Retrain team on change management policy
4. Review and approve the 2 undocumented changes retroactively

---

### Conclusion

XYZ Corp demonstrates a generally effective control environment with strong security controls. However, **access management** and **change management** require improvement. The organization should prioritize:

1. Completing access review remediation immediately
2. Implementing automated change approval gates
3. Enhancing monitoring of control execution

We recommend a follow-up assessment in 6 months to verify remediation effectiveness.
```
