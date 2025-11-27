# Output Formatting Guidelines

## Overview

This document defines standard output formats for specialist agent findings and recommendations. Consistent formatting ensures clear communication across different personas and use cases.

## General Principles

### 1. Clarity First
- Lead with the most important information
- Use plain language where possible
- Define technical terms on first use
- Avoid jargon when simpler terms work

### 2. Structured Output
- Use consistent headers and sections
- Group related findings
- Provide clear visual hierarchy
- Support both human and machine parsing

### 3. Actionable Recommendations
- Every finding should have a clear "what to do"
- Provide specific commands when applicable
- Include effort/impact assessment
- Prioritize by business impact

## Standard Finding Format

### Individual Finding

```markdown
### [SEVERITY] Finding Title

**Package:** package-name@version
**CVE:** CVE-YYYY-NNNNN (if applicable)
**Confidence:** High | Medium | Low

**Description:**
Brief explanation of the issue and why it matters.

**Impact:**
What could happen if this isn't addressed.

**Remediation:**
1. Specific step to fix
2. Alternative approach if applicable

**References:**
- [Link to advisory](url)
- [Additional context](url)
```

### Summary Format

```markdown
## Analysis Summary

| Severity | Count |
|----------|-------|
| Critical | N     |
| High     | N     |
| Medium   | N     |
| Low      | N     |
| Info     | N     |

### Key Findings
1. Most critical finding summary
2. Second most important finding
3. Third most important finding

### Recommended Actions
- [ ] Immediate action item
- [ ] Short-term action item
- [ ] Long-term action item
```

## Persona-Specific Formats

### Security Engineer

Focus: Technical depth, attack vectors, remediation details

```markdown
## Vulnerability Analysis

### Critical: CVE-2024-XXXXX - RCE in package-name

**CVSS:** 9.8 (Critical) | **EPSS:** 0.85 (High)
**CISA KEV:** Yes - Active Exploitation

**Attack Vector:**
- Network accessible (AV:N)
- No authentication required (PR:N)
- No user interaction (UI:N)

**Affected Versions:** < 2.0.0
**Fixed Version:** 2.0.1

**Exploitation:**
Known exploit available. Active exploitation observed in the wild.

**Remediation Priority:** IMMEDIATE

**Steps:**
```bash
npm update package-name@2.0.1
```

**Compensating Controls (if upgrade not possible):**
- Implement WAF rule for attack pattern
- Disable affected feature temporarily
- Apply virtual patching
```

### Software Engineer

Focus: Developer workflow, commands, migration guides

```markdown
## Dependency Updates Needed

### package-name: 1.2.3 → 2.0.0

**Why Update:** Security fix for CVE-2024-XXXXX

**Breaking Changes:**
- `oldMethod()` renamed to `newMethod()`
- Configuration format changed (see migration guide)

**Update Command:**
```bash
npm install package-name@2.0.0
```

**Migration Steps:**
1. Update import statements
2. Rename method calls
3. Update configuration file

**Testing Checklist:**
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual smoke test
```

### Engineering Leader

Focus: Executive summary, metrics, strategic recommendations

```markdown
## Supply Chain Health Report

### Portfolio Risk Score: 7.2/10 (Medium-High)

**Key Metrics:**
| Metric | Current | Target | Trend |
|--------|---------|--------|-------|
| Critical Vulns | 3 | 0 | ↑ |
| Mean Time to Remediate | 45 days | 7 days | → |
| Outdated Dependencies | 23% | <10% | ↓ |

### Risk Heat Map

| Repository | Critical | High | Medium | Risk Level |
|------------|----------|------|--------|------------|
| api-server | 2 | 5 | 12 | HIGH |
| frontend | 0 | 3 | 8 | MEDIUM |
| shared-lib | 1 | 2 | 4 | HIGH |

### Strategic Recommendations

1. **Immediate (This Sprint)**
   - Remediate 3 critical vulnerabilities
   - Resource need: 2 engineer-days

2. **Short-term (This Quarter)**
   - Implement automated vulnerability scanning
   - Reduce dependency count by 15%

3. **Long-term**
   - Establish dependency governance policy
   - Implement SLA for vulnerability remediation
```

### Auditor

Focus: Control effectiveness, compliance mapping, evidence

```markdown
## Audit Finding

### Finding ID: SC-2024-001
### Title: Vulnerable Dependencies in Production

**Criteria:**
Organization policy requires critical vulnerabilities to be remediated within 7 days (Policy SC-4.2).

**Condition:**
3 critical vulnerabilities have remained unpatched for 45+ days.

**Cause:**
- No automated scanning in CI/CD pipeline
- Manual tracking process is inefficient
- Unclear ownership for remediation

**Effect:**
- Non-compliance with internal policy
- Potential regulatory finding (SOC 2 CC7.1)
- Increased risk of security incident

**Recommendation:**
Implement automated dependency scanning with blocking on critical vulnerabilities.

**Management Response:**
[To be completed by management]

**Compliance Mapping:**
| Framework | Control | Status |
|-----------|---------|--------|
| SOC 2 | CC7.1 | Non-compliant |
| NIST CSF | ID.RA-1 | Partial |
| PCI DSS | 6.2 | Non-compliant |

**Evidence Required:**
- [ ] Vulnerability scan reports
- [ ] Remediation timelines
- [ ] Exception documentation
```

## Machine-Readable Output

### JSON Format

```json
{
  "scan_metadata": {
    "timestamp": "2025-01-15T10:30:00Z",
    "target": "repository-name",
    "scanner_version": "1.0.0"
  },
  "summary": {
    "total_findings": 15,
    "by_severity": {
      "critical": 2,
      "high": 5,
      "medium": 6,
      "low": 2
    }
  },
  "findings": [
    {
      "id": "FIND-001",
      "severity": "critical",
      "confidence": "high",
      "title": "RCE in package-name",
      "package": "package-name",
      "version": "1.2.3",
      "cve": "CVE-2024-XXXXX",
      "cvss": 9.8,
      "description": "...",
      "remediation": "...",
      "references": []
    }
  ]
}
```

## Color and Symbol Conventions

### Severity Colors
- Critical: Red (#dc3545)
- High: Orange (#fd7e14)
- Medium: Yellow (#ffc107)
- Low: Blue (#17a2b8)
- Info: Gray (#6c757d)

### Status Symbols
- ✓ Complete/Passed
- ✗ Failed/Blocked
- ⚠ Warning/Attention
- ℹ Information
- → In Progress
- ↑ Increasing (trend)
- ↓ Decreasing (trend)

### Confidence Indicators
- ◉ High confidence
- ◐ Medium confidence
- ○ Low confidence
- ? Speculative

## Accessibility

- Use sufficient color contrast
- Don't rely on color alone for meaning
- Provide text alternatives for symbols
- Support screen readers with semantic markup
