<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
SPDX-License-Identifier: GPL-3.0
-->

# Finding Templates for Supply Chain Security Audits

## Finding Severity Classification

### Severity Matrix

```
                        LIKELIHOOD OF OCCURRENCE
                        Low       Medium      High
                    ┌─────────┬─────────┬─────────┐
              High  │ Medium  │  High   │Critical │
    IMPACT          ├─────────┼─────────┼─────────┤
              Med   │  Low    │ Medium  │  High   │
                    ├─────────┼─────────┼─────────┤
              Low   │ Advisory│  Low    │ Medium  │
                    └─────────┴─────────┴─────────┘
```

### Severity Definitions

```
CRITICAL
─────────────────────────────────────────────────────────────────
• Material weakness in internal control
• Immediate risk of significant harm
• Regulatory non-compliance with penalty risk
• Requires executive attention
• Remediation: Immediate (< 30 days)

HIGH
─────────────────────────────────────────────────────────────────
• Significant deficiency in internal control
• Substantial risk if not addressed
• Potential regulatory concern
• Requires management attention
• Remediation: Near-term (30-60 days)

MEDIUM
─────────────────────────────────────────────────────────────────
• Control deficiency requiring attention
• Moderate risk level
• Process improvement needed
• Remediation: Within quarter (60-90 days)

LOW
─────────────────────────────────────────────────────────────────
• Minor control weakness
• Limited risk exposure
• Best practice recommendation
• Remediation: Within year

ADVISORY (Observation)
─────────────────────────────────────────────────────────────────
• Improvement opportunity
• No control deficiency identified
• Enhancement recommendation
• Remediation: At management discretion
```

## Standard Finding Template

### Full Finding Format

```
╔═══════════════════════════════════════════════════════════════════════════╗
║  FINDING: [Finding Number]                                                ║
║  [Short Descriptive Title]                                                ║
╠═══════════════════════════════════════════════════════════════════════════╣
║                                                                           ║
║  SEVERITY: [Critical/High/Medium/Low/Advisory]                           ║
║  CONTROL AREA: [Control ID/Name]                                         ║
║  OWNER: [Responsible Party]                                              ║
║                                                                           ║
╠═══════════════════════════════════════════════════════════════════════════╣
║  CRITERIA (What Should Happen)                                            ║
╠═══════════════════════════════════════════════════════════════════════════╣
║                                                                           ║
║  [Reference to policy, standard, regulation, or best practice that       ║
║   establishes what is expected]                                          ║
║                                                                           ║
╠═══════════════════════════════════════════════════════════════════════════╣
║  CONDITION (What We Found)                                                ║
╠═══════════════════════════════════════════════════════════════════════════╣
║                                                                           ║
║  [Factual description of what was observed]                              ║
║                                                                           ║
║  Evidence:                                                                ║
║  • [Specific evidence supporting the finding]                            ║
║  • [Quantification where applicable]                                     ║
║                                                                           ║
╠═══════════════════════════════════════════════════════════════════════════╣
║  CAUSE (Why It Happened)                                                  ║
╠═══════════════════════════════════════════════════════════════════════════╣
║                                                                           ║
║  [Root cause analysis]                                                   ║
║                                                                           ║
╠═══════════════════════════════════════════════════════════════════════════╣
║  EFFECT/RISK (Why It Matters)                                            ║
╠═══════════════════════════════════════════════════════════════════════════╣
║                                                                           ║
║  [Potential impact on the organization]                                  ║
║  • Financial impact                                                       ║
║  • Compliance impact                                                      ║
║  • Operational impact                                                     ║
║  • Reputational impact                                                    ║
║                                                                           ║
╠═══════════════════════════════════════════════════════════════════════════╣
║  RECOMMENDATION                                                           ║
╠═══════════════════════════════════════════════════════════════════════════╣
║                                                                           ║
║  [Specific, actionable remediation steps]                                ║
║                                                                           ║
║  1. [Step 1]                                                             ║
║  2. [Step 2]                                                             ║
║  3. [Step 3]                                                             ║
║                                                                           ║
╠═══════════════════════════════════════════════════════════════════════════╣
║  MANAGEMENT RESPONSE                                                      ║
╠═══════════════════════════════════════════════════════════════════════════╣
║                                                                           ║
║  Response: [Agree/Partially Agree/Disagree]                              ║
║                                                                           ║
║  [Management's planned corrective action]                                ║
║                                                                           ║
║  Target Completion Date: [Date]                                          ║
║  Responsible Party: [Name/Title]                                         ║
║                                                                           ║
╚═══════════════════════════════════════════════════════════════════════════╝
```

## Supply Chain Specific Findings

### Finding: Incomplete Vulnerability Scanning Coverage

```
╔═══════════════════════════════════════════════════════════════════════════╗
║  FINDING: SC-2024-001                                                     ║
║  Incomplete Vulnerability Scanning Coverage                               ║
╠═══════════════════════════════════════════════════════════════════════════╣
║  SEVERITY: High        CONTROL: VM-1 Vulnerability Scanning              ║
╠═══════════════════════════════════════════════════════════════════════════╣
║  CRITERIA                                                                 ║
║  ─────────────────────────────────────────────────────────────────────    ║
║  Per the organization's Vulnerability Management Policy (Section 3.2),   ║
║  "All production applications and their dependencies must be subject     ║
║  to continuous automated vulnerability scanning."                        ║
║                                                                           ║
╠═══════════════════════════════════════════════════════════════════════════╣
║  CONDITION                                                                ║
║  ─────────────────────────────────────────────────────────────────────    ║
║  During our testing of vulnerability scanning controls, we identified    ║
║  that 12 of 145 production repositories (8.3%) were not configured      ║
║  for automated vulnerability scanning.                                   ║
║                                                                           ║
║  Specifically:                                                           ║
║  • 8 repositories were not onboarded to the scanning tool                ║
║  • 4 repositories had scanning disabled due to configuration errors     ║
║                                                                           ║
║  The uncovered repositories include 3 applications handling customer    ║
║  payment data.                                                           ║
║                                                                           ║
╠═══════════════════════════════════════════════════════════════════════════╣
║  CAUSE                                                                    ║
║  ─────────────────────────────────────────────────────────────────────    ║
║  • No automated process to ensure new repositories are onboarded        ║
║  • Lack of periodic coverage verification                                ║
║  • Configuration errors not monitored or alerted                         ║
║                                                                           ║
╠═══════════════════════════════════════════════════════════════════════════╣
║  EFFECT/RISK                                                              ║
║  ─────────────────────────────────────────────────────────────────────    ║
║  • Vulnerabilities in uncovered applications may go undetected          ║
║  • Increased risk of security breach in payment-handling systems        ║
║  • Potential PCI DSS compliance violation (Req. 6.3)                    ║
║  • Estimated exposure: 3 critical applications for ~6 months            ║
║                                                                           ║
╠═══════════════════════════════════════════════════════════════════════════╣
║  RECOMMENDATION                                                           ║
║  ─────────────────────────────────────────────────────────────────────    ║
║  1. Immediately enable scanning for all 12 uncovered repositories       ║
║  2. Implement automated onboarding for new repositories                 ║
║  3. Create monitoring alert for scan configuration failures             ║
║  4. Establish monthly coverage verification process                      ║
║                                                                           ║
╠═══════════════════════════════════════════════════════════════════════════╣
║  MANAGEMENT RESPONSE                                                      ║
║  ─────────────────────────────────────────────────────────────────────    ║
║  Response: Agree                                                          ║
║                                                                           ║
║  All 12 repositories have been onboarded to scanning as of [Date].      ║
║  We are implementing automated onboarding through our CI/CD pipeline    ║
║  and will add coverage monitoring to our security dashboard.             ║
║                                                                           ║
║  Target Completion Date: [Date]                                          ║
║  Responsible Party: [Name], Security Engineering Manager                 ║
╚═══════════════════════════════════════════════════════════════════════════╝
```

### Finding: SLA Breaches for Critical Vulnerabilities

```
╔═══════════════════════════════════════════════════════════════════════════╗
║  FINDING: SC-2024-002                                                     ║
║  Critical Vulnerability Remediation SLA Breaches                          ║
╠═══════════════════════════════════════════════════════════════════════════╣
║  SEVERITY: Critical    CONTROL: VM-3 Vulnerability Remediation           ║
╠═══════════════════════════════════════════════════════════════════════════╣
║  CRITERIA                                                                 ║
║  ─────────────────────────────────────────────────────────────────────    ║
║  Per Vulnerability Management Policy (Section 4.1), critical severity   ║
║  vulnerabilities must be remediated within 24 hours of detection.       ║
║  Additionally, CISA BOD 22-01 requires federal suppliers to remediate   ║
║  Known Exploited Vulnerabilities (KEV) within 14 days.                  ║
║                                                                           ║
╠═══════════════════════════════════════════════════════════════════════════╣
║  CONDITION                                                                ║
║  ─────────────────────────────────────────────────────────────────────    ║
║  Of 8 critical vulnerabilities identified during the audit period,      ║
║  3 (37.5%) were not remediated within the 24-hour SLA:                  ║
║                                                                           ║
║  │ CVE ID          │ Component    │ Days to Fix │ KEV │ Reason         │ ║
║  │─────────────────┼──────────────┼─────────────┼─────┼────────────────│ ║
║  │ CVE-2024-0001   │ log4j 2.14   │     4       │ Yes │ Not assigned   │ ║
║  │ CVE-2024-0002   │ openssl 3.0  │     3       │ No  │ Dep conflict   │ ║
║  │ CVE-2024-0003   │ axios 0.21   │     2       │ No  │ Owner on PTO   │ ║
║                                                                           ║
║  CVE-2024-0001 was in the CISA KEV catalog and required federal        ║
║  compliance action.                                                      ║
║                                                                           ║
╠═══════════════════════════════════════════════════════════════════════════╣
║  CAUSE                                                                    ║
║  ─────────────────────────────────────────────────────────────────────    ║
║  • Inadequate escalation procedures for unassigned vulnerabilities      ║
║  • No backup assignee process for PTO coverage                          ║
║  • Dependency conflicts not identified during initial triage            ║
║                                                                           ║
╠═══════════════════════════════════════════════════════════════════════════╣
║  EFFECT/RISK                                                              ║
║  ─────────────────────────────────────────────────────────────────────    ║
║  • Extended exposure to actively exploited vulnerabilities              ║
║  • Potential compliance violation with federal requirements             ║
║  • Increased probability of security incident                            ║
║  • Each day of delay compounds exploit opportunity                       ║
║                                                                           ║
╠═══════════════════════════════════════════════════════════════════════════╣
║  RECOMMENDATION                                                           ║
║  ─────────────────────────────────────────────────────────────────────    ║
║  1. Implement automated escalation if critical vuln unassigned >4 hours ║
║  2. Establish on-call rotation for critical vulnerability response      ║
║  3. Add dependency conflict check to triage process                     ║
║  4. Create KEV-specific expedited workflow                              ║
║  5. Implement automated PTO coverage assignment                          ║
║                                                                           ║
╠═══════════════════════════════════════════════════════════════════════════╣
║  MANAGEMENT RESPONSE                                                      ║
║  ─────────────────────────────────────────────────────────────────────    ║
║  Response: Agree                                                          ║
║                                                                           ║
║  We acknowledge the severity of this finding and are implementing:      ║
║  - Immediate: On-call rotation effective [Date]                         ║
║  - 30 days: Automated escalation alerts                                 ║
║  - 60 days: KEV-specific workflow and dependency pre-check              ║
║                                                                           ║
║  Target Completion Date: [Date]                                          ║
║  Responsible Party: [Name], VP of Engineering                           ║
╚═══════════════════════════════════════════════════════════════════════════╝
```

### Finding: Missing SBOM for Software Releases

```
╔═══════════════════════════════════════════════════════════════════════════╗
║  FINDING: SC-2024-003                                                     ║
║  Software Releases Missing Required SBOM Documentation                    ║
╠═══════════════════════════════════════════════════════════════════════════╣
║  SEVERITY: Medium      CONTROL: SBOM-1 SBOM Generation                   ║
╠═══════════════════════════════════════════════════════════════════════════╣
║  CRITERIA                                                                 ║
║  ─────────────────────────────────────────────────────────────────────    ║
║  Per Executive Order 14028 and the organization's Software Supply       ║
║  Chain Policy (Section 2.3), all software releases must include a       ║
║  machine-readable SBOM in SPDX or CycloneDX format containing NTIA      ║
║  minimum elements.                                                       ║
║                                                                           ║
╠═══════════════════════════════════════════════════════════════════════════╣
║  CONDITION                                                                ║
║  ─────────────────────────────────────────────────────────────────────    ║
║  Of 45 production releases during the audit period, 8 (17.8%) did not   ║
║  have an associated SBOM. Additionally, of the 37 releases with SBOMs:  ║
║                                                                           ║
║  • 5 were missing component version information                         ║
║  • 3 were missing unique identifiers (PURL/CPE)                        ║
║  • 2 were missing supplier information                                  ║
║                                                                           ║
║  Total non-compliant releases: 18 of 45 (40%)                          ║
║                                                                           ║
╠═══════════════════════════════════════════════════════════════════════════╣
║  CAUSE                                                                    ║
║  ─────────────────────────────────────────────────────────────────────    ║
║  • SBOM generation not enforced in all CI/CD pipelines                  ║
║  • Legacy applications not yet integrated with SBOM tooling             ║
║  • SBOM validation not performed prior to release                       ║
║                                                                           ║
╠═══════════════════════════════════════════════════════════════════════════╣
║  EFFECT/RISK                                                              ║
║  ─────────────────────────────────────────────────────────────────────    ║
║  • Inability to respond to supply chain incidents efficiently           ║
║  • Non-compliance with federal software supply chain requirements       ║
║  • Limited visibility into software composition                          ║
║  • Potential customer contract violations                                ║
║                                                                           ║
╠═══════════════════════════════════════════════════════════════════════════╣
║  RECOMMENDATION                                                           ║
║  ─────────────────────────────────────────────────────────────────────    ║
║  1. Implement mandatory SBOM generation as CI/CD gate                   ║
║  2. Add SBOM validation step to verify minimum elements                 ║
║  3. Create remediation plan for legacy application integration          ║
║  4. Establish SBOM quality monitoring dashboard                          ║
║                                                                           ║
╠═══════════════════════════════════════════════════════════════════════════╣
║  MANAGEMENT RESPONSE                                                      ║
║  ─────────────────────────────────────────────────────────────────────    ║
║  Response: Agree                                                          ║
║                                                                           ║
║  SBOM generation will be enforced as a mandatory pipeline gate for all  ║
║  new releases effective [Date]. Legacy application integration is       ║
║  planned for completion by [Date].                                      ║
║                                                                           ║
║  Target Completion Date: [Date]                                          ║
║  Responsible Party: [Name], Director of DevOps                          ║
╚═══════════════════════════════════════════════════════════════════════════╝
```

### Finding: Abandoned Dependencies in Production

```
╔═══════════════════════════════════════════════════════════════════════════╗
║  FINDING: SC-2024-004                                                     ║
║  Production Systems Using Abandoned Dependencies                          ║
╠═══════════════════════════════════════════════════════════════════════════╣
║  SEVERITY: Medium      CONTROL: TPR-1 Vendor Security Assessment         ║
╠═══════════════════════════════════════════════════════════════════════════╣
║  CRITERIA                                                                 ║
║  ─────────────────────────────────────────────────────────────────────    ║
║  Per the Third-Party Software Policy (Section 5.2), dependencies must   ║
║  be actively maintained. Packages with no updates in 24+ months or      ║
║  explicitly archived by maintainers require documented risk acceptance  ║
║  and migration plan.                                                     ║
║                                                                           ║
╠═══════════════════════════════════════════════════════════════════════════╣
║  CONDITION                                                                ║
║  ─────────────────────────────────────────────────────────────────────    ║
║  Analysis of production dependencies identified 23 packages meeting     ║
║  abandonment criteria:                                                   ║
║                                                                           ║
║  • 15 packages: No commits in 24+ months                                ║
║  • 5 packages: Maintainer explicitly archived project                   ║
║  • 3 packages: No response to critical security issues >6 months        ║
║                                                                           ║
║  Of these:                                                               ║
║  • 8 have documented risk acceptance                                    ║
║  • 2 have active migration plans                                        ║
║  • 13 have no documentation (56.5%)                                     ║
║                                                                           ║
║  Notable: 2 undocumented packages handle authentication functions       ║
║                                                                           ║
╠═══════════════════════════════════════════════════════════════════════════╣
║  CAUSE                                                                    ║
║  ─────────────────────────────────────────────────────────────────────    ║
║  • No automated monitoring for dependency maintenance status            ║
║  • Risk acceptance process not enforced                                 ║
║  • Dependencies adopted before policy implementation                     ║
║                                                                           ║
╠═══════════════════════════════════════════════════════════════════════════╣
║  EFFECT/RISK                                                              ║
║  ─────────────────────────────────────────────────────────────────────    ║
║  • Security vulnerabilities may not be patched by maintainer           ║
║  • No support for compatibility with newer systems                      ║
║  • Increased future technical debt and migration complexity             ║
║  • Authentication packages pose elevated security risk                   ║
║                                                                           ║
╠═══════════════════════════════════════════════════════════════════════════╣
║  RECOMMENDATION                                                           ║
║  ─────────────────────────────────────────────────────────────────────    ║
║  1. Document risk acceptance for all 13 undocumented packages           ║
║  2. Prioritize migration of authentication-related packages             ║
║  3. Implement automated abandonment monitoring                          ║
║  4. Add abandonment check to new dependency approval process            ║
║                                                                           ║
╠═══════════════════════════════════════════════════════════════════════════╣
║  MANAGEMENT RESPONSE                                                      ║
║  ─────────────────────────────────────────────────────────────────────    ║
║  Response: Agree                                                          ║
║                                                                           ║
║  Risk acceptance documentation will be completed for all packages by    ║
║  [Date]. Authentication packages are prioritized for Q1 migration.      ║
║                                                                           ║
║  Target Completion Date: [Date]                                          ║
║  Responsible Party: [Name], Security Architect                          ║
╚═══════════════════════════════════════════════════════════════════════════╝
```

## Report Summary Templates

### Findings Summary Table

```
╔═══════════════════════════════════════════════════════════════════════════╗
║                    FINDINGS SUMMARY                                       ║
╠═══════════════════════════════════════════════════════════════════════════╣
║                                                                           ║
║  │ ID          │ Title                              │ Severity │ Status  │║
║  │─────────────┼────────────────────────────────────┼──────────┼─────────│║
║  │ SC-2024-001 │ Incomplete Scanning Coverage       │   High   │  Open   │║
║  │ SC-2024-002 │ Critical Vuln SLA Breaches         │ Critical │  Open   │║
║  │ SC-2024-003 │ Missing SBOM Documentation         │  Medium  │  Open   │║
║  │ SC-2024-004 │ Abandoned Dependencies             │  Medium  │  Open   │║
║                                                                           ║
║  SUMMARY BY SEVERITY                                                      ║
║  ───────────────────────────────────────────────────────────────────      ║
║  Critical:  1                                                             ║
║  High:      1                                                             ║
║  Medium:    2                                                             ║
║  Low:       0                                                             ║
║  Advisory:  0                                                             ║
║  ───────────────────────────────────────────────────────────────────      ║
║  Total:     4                                                             ║
║                                                                           ║
╚═══════════════════════════════════════════════════════════════════════════╝
```

### Management Action Plan

```
╔═══════════════════════════════════════════════════════════════════════════╗
║                    MANAGEMENT ACTION PLAN                                 ║
╠═══════════════════════════════════════════════════════════════════════════╣
║                                                                           ║
║  │Finding     │Action Item                       │Owner     │Due Date   │║
║  │────────────┼──────────────────────────────────┼──────────┼───────────│║
║  │SC-2024-001 │Enable scanning for 12 repos      │J. Smith  │2024-12-01 │║
║  │SC-2024-001 │Implement auto-onboarding         │J. Smith  │2024-12-15 │║
║  │SC-2024-001 │Create coverage monitoring        │J. Smith  │2024-12-15 │║
║  │SC-2024-002 │Implement on-call rotation        │K. Jones  │2024-11-20 │║
║  │SC-2024-002 │Deploy automated escalation       │K. Jones  │2024-12-15 │║
║  │SC-2024-002 │Create KEV workflow               │K. Jones  │2025-01-15 │║
║  │SC-2024-003 │Enforce SBOM in pipelines         │L. Chen   │2024-12-31 │║
║  │SC-2024-003 │Integrate legacy applications     │L. Chen   │2025-02-28 │║
║  │SC-2024-004 │Document risk acceptance          │M. Wilson │2024-12-15 │║
║  │SC-2024-004 │Migrate auth packages             │M. Wilson │2025-03-31 │║
║                                                                           ║
╚═══════════════════════════════════════════════════════════════════════════╝
```

## Quick Reference

### Finding Writing Checklist

```
□ Title is clear and descriptive
□ Severity is justified
□ Criteria cites specific policy/standard
□ Condition is factual and quantified
□ Cause identifies root issue
□ Effect articulates business risk
□ Recommendation is specific and actionable
□ Evidence is referenced
□ Management response obtained
□ Target date is realistic
□ Responsible party is identified
```
