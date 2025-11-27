<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
SPDX-License-Identifier: GPL-3.0
-->

# Compliance Mapping for Supply Chain Security

## Framework Overview

### Common Compliance Requirements

```
Framework        Supply Chain Focus                    Typical Orgs
─────────────────────────────────────────────────────────────────────
SOC 2            Vendor management, change control    SaaS, Cloud
PCI DSS          Patch management, software inventory Payments
HIPAA            Third-party risk, data protection    Healthcare
FedRAMP          SBOM, provenance, continuous mon.    Government
NIST CSF         Supply chain risk management         All
ISO 27001        Supplier relationships, asset mgmt   Enterprise
GDPR             Data processor requirements          EU operations
```

## SOC 2 Mapping

### Trust Service Criteria - Supply Chain Relevant

```
╔═══════════════════════════════════════════════════════════════════════════╗
║  CC6.1 - Logical and Physical Access Controls                             ║
╠═══════════════════════════════════════════════════════════════════════════╣
║                                                                           ║
║  Requirement: Access to software and data is restricted                   ║
║                                                                           ║
║  Supply Chain Evidence:                                                   ║
║  • Package registry access controls (npm, PyPI auth)                     ║
║  • Dependency lockfile enforcement                                        ║
║  • Private artifact repository controls                                   ║
║  • Code signing verification                                              ║
║                                                                           ║
║  Scanner Mapping:                                                         ║
║  • Provenance validation                                                  ║
║  • SBOM generation showing authorized sources                             ║
║                                                                           ║
╚═══════════════════════════════════════════════════════════════════════════╝

╔═══════════════════════════════════════════════════════════════════════════╗
║  CC7.1 - System Components are Monitored                                  ║
╠═══════════════════════════════════════════════════════════════════════════╣
║                                                                           ║
║  Requirement: Detect and report system anomalies                          ║
║                                                                           ║
║  Supply Chain Evidence:                                                   ║
║  • Continuous vulnerability scanning                                      ║
║  • Dependency change monitoring                                           ║
║  • Maintainer change alerts                                               ║
║  • Suspicious package detection                                           ║
║                                                                           ║
║  Scanner Mapping:                                                         ║
║  • Vulnerability scan reports with timestamps                             ║
║  • Package health monitoring reports                                      ║
║  • Alert configurations and response records                              ║
║                                                                           ║
╚═══════════════════════════════════════════════════════════════════════════╝

╔═══════════════════════════════════════════════════════════════════════════╗
║  CC8.1 - Change Management                                                ║
╠═══════════════════════════════════════════════════════════════════════════╣
║                                                                           ║
║  Requirement: Changes are authorized, tested, approved                    ║
║                                                                           ║
║  Supply Chain Evidence:                                                   ║
║  • Dependency update review process                                       ║
║  • Version pinning policies                                               ║
║  • Automated testing for dependency changes                               ║
║  • Approval workflows for major updates                                   ║
║                                                                           ║
║  Scanner Mapping:                                                         ║
║  • PR/MR records for dependency updates                                   ║
║  • CI/CD logs showing test execution                                      ║
║  • Lockfile change history                                                ║
║                                                                           ║
╚═══════════════════════════════════════════════════════════════════════════╝
```

### SOC 2 Evidence Checklist

```
□ Vulnerability scanning policy document
□ Scan schedule and coverage report
□ Remediation SLA documentation
□ Historical vulnerability metrics
□ Dependency management policy
□ Third-party software inventory
□ Vendor security assessment records
□ Incident response procedures (supply chain specific)
```

## PCI DSS 4.0 Mapping

### Requirement 6: Develop and Maintain Secure Systems

```
╔═══════════════════════════════════════════════════════════════════════════╗
║  6.3.1 - Security Vulnerabilities Identified and Managed                  ║
╠═══════════════════════════════════════════════════════════════════════════╣
║                                                                           ║
║  Requirement:                                                             ║
║  "Security vulnerabilities are identified and managed as follows:         ║
║   - New security vulnerabilities are identified using industry-recognized ║
║     sources for security vulnerability information"                       ║
║                                                                           ║
║  Supply Chain Evidence:                                                   ║
║  • Integration with NVD, GitHub Advisory DB                              ║
║  • CVE monitoring and alerting                                            ║
║  • Vulnerability scanner configuration                                    ║
║                                                                           ║
║  Scanner Reports Needed:                                                  ║
║  • Vulnerability scan results with CVE references                         ║
║  • Data source configuration documentation                                ║
║                                                                           ║
╚═══════════════════════════════════════════════════════════════════════════╝

╔═══════════════════════════════════════════════════════════════════════════╗
║  6.3.3 - Critical and High Vulnerabilities Addressed                      ║
╠═══════════════════════════════════════════════════════════════════════════╣
║                                                                           ║
║  Requirement:                                                             ║
║  "Applicable patches/updates for critical and high-risk vulnerabilities   ║
║   are installed within one month of release"                              ║
║                                                                           ║
║  Supply Chain Evidence:                                                   ║
║  • Remediation timelines by severity                                      ║
║  • SLA compliance reports                                                 ║
║  • Historical MTTR metrics                                                ║
║                                                                           ║
║  Scanner Reports Needed:                                                  ║
║  • Vulnerability aging report                                             ║
║  • Remediation completion dates vs. disclosure dates                      ║
║                                                                           ║
╚═══════════════════════════════════════════════════════════════════════════╝

╔═══════════════════════════════════════════════════════════════════════════╗
║  6.4.3 - All Payment Page Scripts Managed                                 ║
╠═══════════════════════════════════════════════════════════════════════════╣
║                                                                           ║
║  Requirement (NEW in 4.0):                                                ║
║  "All payment page scripts that are loaded and executed in the consumer's ║
║   browser are managed as follows..."                                      ║
║                                                                           ║
║  Supply Chain Evidence:                                                   ║
║  • Client-side dependency inventory                                       ║
║  • Subresource Integrity (SRI) implementation                            ║
║  • CDN/third-party script authorization                                   ║
║                                                                           ║
║  Scanner Reports Needed:                                                  ║
║  • Frontend dependency SBOM                                               ║
║  • SRI hash verification                                                  ║
║                                                                           ║
╚═══════════════════════════════════════════════════════════════════════════╝
```

### PCI DSS Evidence Checklist

```
□ Software inventory including all third-party components
□ Vulnerability scan reports (quarterly at minimum)
□ Patch/update timeline documentation
□ Remediation SLA policy and compliance records
□ Critical/high vulnerability closure evidence
□ Client-side script inventory (for payment pages)
□ SRI implementation records
□ Third-party vendor security assessments
```

## NIST Cybersecurity Framework

### Supply Chain Risk Management (ID.SC)

```
╔═══════════════════════════════════════════════════════════════════════════╗
║  ID.SC-1 - Cyber supply chain risk management processes                   ║
╠═══════════════════════════════════════════════════════════════════════════╣
║                                                                           ║
║  Requirement:                                                             ║
║  "Cyber supply chain risk management processes are identified,            ║
║   established, assessed, managed, and agreed to by organizational         ║
║   stakeholders"                                                           ║
║                                                                           ║
║  Supply Chain Evidence:                                                   ║
║  • Supply chain security policy                                           ║
║  • Risk assessment methodology                                            ║
║  • Stakeholder RACI matrix                                                ║
║                                                                           ║
║  Scanner Contribution:                                                    ║
║  • Automated risk scoring                                                 ║
║  • Continuous monitoring capabilities                                     ║
║                                                                           ║
╚═══════════════════════════════════════════════════════════════════════════╝

╔═══════════════════════════════════════════════════════════════════════════╗
║  ID.SC-2 - Suppliers and Third-Party Partners Identified                  ║
╠═══════════════════════════════════════════════════════════════════════════╣
║                                                                           ║
║  Requirement:                                                             ║
║  "Suppliers and third-party partners of information systems, components,  ║
║   and services are identified, prioritized, and assessed"                 ║
║                                                                           ║
║  Supply Chain Evidence:                                                   ║
║  • Complete dependency inventory                                          ║
║  • Criticality classification                                             ║
║  • Vendor assessment records                                              ║
║                                                                           ║
║  Scanner Contribution:                                                    ║
║  • SBOM generation                                                        ║
║  • Dependency tree visualization                                          ║
║  • Package health scores                                                  ║
║                                                                           ║
╚═══════════════════════════════════════════════════════════════════════════╝

╔═══════════════════════════════════════════════════════════════════════════╗
║  ID.SC-4 - Suppliers and Partners Meet Obligations                        ║
╠═══════════════════════════════════════════════════════════════════════════╣
║                                                                           ║
║  Requirement:                                                             ║
║  "Suppliers and third-party partners are routinely assessed using audits, ║
║   test results, or other forms of evaluations"                            ║
║                                                                           ║
║  Supply Chain Evidence:                                                   ║
║  • OpenSSF Scorecard results                                              ║
║  • Maintainer activity monitoring                                         ║
║  • Security advisory response tracking                                    ║
║                                                                           ║
║  Scanner Contribution:                                                    ║
║  • Package health analysis                                                ║
║  • Abandonment detection                                                  ║
║  • Vulnerability response metrics                                         ║
║                                                                           ║
╚═══════════════════════════════════════════════════════════════════════════╝
```

## FedRAMP Requirements

### SBOM Requirements (New)

```
╔═══════════════════════════════════════════════════════════════════════════╗
║  FedRAMP SBOM Requirements                                                ║
╠═══════════════════════════════════════════════════════════════════════════╣
║                                                                           ║
║  Effective: 2024 (phased implementation)                                  ║
║                                                                           ║
║  Minimum SBOM Fields (NTIA Minimum Elements):                            ║
║  • Supplier name                                                          ║
║  • Component name                                                         ║
║  • Component version                                                      ║
║  • Unique identifier (PURL or CPE)                                       ║
║  • Dependency relationship                                                ║
║  • Author of SBOM data                                                    ║
║  • Timestamp                                                              ║
║                                                                           ║
║  Additional FedRAMP Requirements:                                         ║
║  • Machine-readable format (SPDX or CycloneDX)                           ║
║  • Updated with each release                                              ║
║  • Vulnerability correlation capability                                   ║
║  • Accessible to agency upon request                                      ║
║                                                                           ║
║  Scanner Contribution:                                                    ║
║  • Automated SBOM generation                                              ║
║  • Format compliance validation                                           ║
║  • Vulnerability matching to SBOM components                              ║
║                                                                           ║
╚═══════════════════════════════════════════════════════════════════════════╝
```

### SLSA Requirements

```
╔═══════════════════════════════════════════════════════════════════════════╗
║  Supply Chain Levels for Software Artifacts (SLSA)                        ║
╠═══════════════════════════════════════════════════════════════════════════╣
║                                                                           ║
║  Level 1: Documentation                                                   ║
║  • Build process documented                                               ║
║  • Provenance generated (can be manually)                                ║
║  → Scanner: Validate provenance exists                                    ║
║                                                                           ║
║  Level 2: Build Service                                                   ║
║  • Version control used                                                   ║
║  • Hosted build service                                                   ║
║  • Provenance generated by service                                        ║
║  → Scanner: Validate provenance source                                    ║
║                                                                           ║
║  Level 3: Build Platform                                                  ║
║  • Hardened build platform                                                ║
║  • Non-falsifiable provenance                                             ║
║  • Isolated builds                                                        ║
║  → Scanner: Cryptographic provenance verification                         ║
║                                                                           ║
╚═══════════════════════════════════════════════════════════════════════════╝
```

## Compliance Report Templates

### Audit Evidence Package

```
Supply Chain Security Audit Evidence
Framework: [SOC 2 / PCI DSS / NIST / etc.]
Period: [Start Date] to [End Date]
Prepared: [Date]
Prepared By: [Name/Team]

TABLE OF CONTENTS
1. Executive Summary
2. Policy Documentation
3. Technical Controls Evidence
4. Operational Evidence
5. Metrics and Trends
6. Exception Documentation

SECTION 3: TECHNICAL CONTROLS EVIDENCE

3.1 Vulnerability Scanning
────────────────────────────────────────
Tool(s) in use: [Scanner names]
Scan frequency: [Daily/Weekly/etc.]
Coverage: [X]% of repositories

Evidence:
• Scan configuration screenshots
• Sample scan reports (attached)
• Coverage report

3.2 Software Bill of Materials
────────────────────────────────────────
SBOM Format: [SPDX/CycloneDX]
Generation: [Manual/Automated]
Update frequency: [Per release/Daily]

Evidence:
• Sample SBOM files (attached)
• SBOM generation process documentation
• Artifact correlation report

3.3 Vulnerability Remediation
────────────────────────────────────────
SLA by Severity:
• Critical: [X] hours
• High: [X] days
• Medium: [X] days
• Low: [X] days

Evidence:
• SLA compliance report
• MTTR trend analysis
• Sample remediation tickets

[Continue for each control area...]
```

### Compliance Dashboard Data

```json
{
  "compliance_status": {
    "framework": "SOC 2",
    "period": "2024-Q4",
    "overall_compliance": 94,
    "controls": [
      {
        "id": "CC6.1",
        "name": "Access Controls",
        "status": "compliant",
        "evidence_count": 12,
        "last_review": "2024-11-15"
      },
      {
        "id": "CC7.1",
        "name": "System Monitoring",
        "status": "compliant",
        "evidence_count": 8,
        "last_review": "2024-11-15"
      },
      {
        "id": "CC8.1",
        "name": "Change Management",
        "status": "partial",
        "evidence_count": 15,
        "gaps": ["Manual approval documentation incomplete"],
        "remediation_date": "2024-12-01"
      }
    ]
  }
}
```

## Cross-Framework Mapping

### Control Mapping Matrix

```
Control Area          │ SOC 2  │ PCI DSS │ NIST CSF │ ISO 27001
──────────────────────┼────────┼─────────┼──────────┼──────────
Vulnerability Mgmt    │ CC7.1  │ 6.3.1   │ ID.RA-1  │ A.12.6.1
Patch Management      │ CC7.2  │ 6.3.3   │ PR.IP-12 │ A.12.6.1
Software Inventory    │ CC6.1  │ 6.3.2   │ ID.AM-2  │ A.8.1.1
Third-Party Risk      │ CC9.2  │ 12.8    │ ID.SC-4  │ A.15.1.1
Change Control        │ CC8.1  │ 6.4.5   │ PR.IP-3  │ A.12.1.2
Incident Response     │ CC7.4  │ 12.10   │ RS.RP-1  │ A.16.1.1
SBOM/Inventory        │ CC6.1  │ 6.3.2   │ ID.AM-2  │ A.8.1.1
```

### Evidence Reuse Guide

```
Evidence Type                  Applicable Frameworks
───────────────────────────────────────────────────────────────
Vulnerability scan reports     SOC 2, PCI DSS, NIST, ISO, FedRAMP
SBOM documentation            NIST, FedRAMP, CISA requirements
Remediation metrics           All frameworks
Dependency inventory          All frameworks
Patch timeline records        PCI DSS, SOC 2, NIST
Third-party assessments       SOC 2, PCI DSS, NIST, ISO
Incident response records     All frameworks
Policy documentation          All frameworks
```

## Quick Reference

### Compliance Preparation Checklist

```
Pre-Audit (4-6 weeks before):
□ Verify scan coverage is complete
□ Run gap analysis against requirements
□ Compile evidence packages
□ Review and close aging vulnerabilities
□ Update policy documentation
□ Prepare exception documentation

During Audit:
□ Provide requested evidence promptly
□ Demonstrate scanning capabilities
□ Walk through remediation workflow
□ Show metrics dashboards
□ Explain exception process

Post-Audit:
□ Address findings promptly
□ Update procedures based on feedback
□ Document lessons learned
□ Plan for continuous compliance
```
