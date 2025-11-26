<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
SPDX-License-Identifier: GPL-3.0
-->

# Audit Standards for Supply Chain Security

## Relevant Audit Standards

### ISACA COBIT Framework

**Relevant Control Objectives:**

```
APO09 - Manage Service Agreements
─────────────────────────────────────────────────────────────────
Objective: Ensure IT suppliers deliver agreed services effectively

Supply Chain Relevance:
• Third-party software meets security requirements
• Vendor contracts include security provisions
• SLAs for vulnerability response

Audit Focus Areas:
□ Vendor security assessment process
□ Contract security clauses
□ SLA monitoring and enforcement
□ Periodic vendor reviews

APO12 - Manage Risk
─────────────────────────────────────────────────────────────────
Objective: Continually identify, assess, and reduce IT-related risk

Supply Chain Relevance:
• Dependency risk assessment
• Vulnerability prioritization
• Risk acceptance documentation

Audit Focus Areas:
□ Risk assessment methodology
□ Risk register completeness
□ Risk treatment decisions
□ Risk monitoring effectiveness

BAI06 - Manage Changes
─────────────────────────────────────────────────────────────────
Objective: Manage all changes in a controlled manner

Supply Chain Relevance:
• Dependency updates
• Version control
• Change authorization

Audit Focus Areas:
□ Change request documentation
□ Testing requirements
□ Approval workflows
□ Emergency change procedures
```

### NIST SP 800-53 Controls

**Supply Chain Specific Controls:**

```
SA-8 Security and Privacy Engineering Principles
─────────────────────────────────────────────────────────────────
Requirement: Apply security engineering principles in system development

Testing Procedures:
1. Review secure development lifecycle documentation
2. Verify dependency security review process
3. Examine secure coding standards
4. Test vulnerability scanning integration

Evidence Required:
• SDLC documentation
• Code review records
• Security testing results
• Vulnerability remediation records

SA-9 External System Services
─────────────────────────────────────────────────────────────────
Requirement: Define and document government and external provider oversight

Testing Procedures:
1. Identify all external dependencies
2. Review vendor security assessments
3. Verify contractual security requirements
4. Test monitoring of external services

Evidence Required:
• External service inventory
• Vendor assessments
• Security requirements in contracts
• Monitoring procedures

SR-3 Supply Chain Controls and Processes
─────────────────────────────────────────────────────────────────
Requirement: Establish and implement processes to protect against supply chain risks

Testing Procedures:
1. Review supply chain risk management policy
2. Test dependency verification processes
3. Examine provenance validation
4. Verify incident response for supply chain events

Evidence Required:
• Supply chain security policy
• Dependency verification records
• Provenance attestations
• Supply chain incident procedures
```

### ISO/IEC 27001:2022 Controls

**Annex A Supply Chain Controls:**

```
A.5.21 Managing Information Security in ICT Supply Chain
─────────────────────────────────────────────────────────────────
Control: Processes and procedures for managing ICT supply chain risks

Audit Approach:
1. Policy Review
   • Supply chain security policy exists
   • Covers open source and commercial software
   • Defines risk assessment criteria

2. Process Testing
   • Verify dependency review process
   • Test vulnerability scanning workflow
   • Examine update authorization process

3. Evidence Collection
   • Software inventory records
   • Risk assessment documentation
   • Vendor evaluation records

A.5.22 Monitoring, Review, and Change Management of Supplier Services
─────────────────────────────────────────────────────────────────
Control: Regularly monitor, review, and evaluate supplier service delivery

Audit Approach:
1. Monitoring Effectiveness
   • Continuous vulnerability scanning
   • Dependency health monitoring
   • Security advisory tracking

2. Change Management
   • Dependency update process
   • Version control procedures
   • Approval requirements

3. Review Cycles
   • Periodic vendor assessments
   • License compliance reviews
   • Security posture updates

A.8.28 Secure Coding
─────────────────────────────────────────────────────────────────
Control: Secure coding principles applied to software development

Audit Approach:
1. Standards Review
   • Secure coding guidelines exist
   • Dependency security requirements
   • Third-party component guidelines

2. Implementation Testing
   • Code review processes
   • Automated security scanning
   • Vulnerability remediation workflow
```

## Audit Program Structure

### Supply Chain Security Audit Program

```
╔═══════════════════════════════════════════════════════════════════════════╗
║                    AUDIT PROGRAM: SUPPLY CHAIN SECURITY                   ║
╠═══════════════════════════════════════════════════════════════════════════╣
║                                                                           ║
║  Audit Objective:                                                         ║
║  Evaluate the design and operating effectiveness of controls over         ║
║  software supply chain security                                           ║
║                                                                           ║
║  Scope:                                                                   ║
║  • Third-party dependency management                                      ║
║  • Vulnerability identification and remediation                           ║
║  • Software composition analysis                                          ║
║  • Software bill of materials                                             ║
║  • Provenance and integrity verification                                  ║
║                                                                           ║
║  Period: [Start Date] to [End Date]                                       ║
║  Prepared By: [Auditor Name]                                              ║
║  Reviewed By: [Reviewer Name]                                             ║
║                                                                           ║
╚═══════════════════════════════════════════════════════════════════════════╝
```

### Audit Phases

```
Phase 1: Planning (Week 1)
─────────────────────────────────────────────────────────────────
□ Understand the IT environment
□ Identify in-scope applications/repositories
□ Review prior audit findings
□ Identify key personnel
□ Request preliminary documentation
□ Develop audit procedures

Phase 2: Fieldwork (Weeks 2-4)
─────────────────────────────────────────────────────────────────
□ Control walkthrough documentation
□ Testing of design effectiveness
□ Testing of operating effectiveness
□ Sample selection and testing
□ Exception identification
□ Root cause analysis

Phase 3: Reporting (Week 5)
─────────────────────────────────────────────────────────────────
□ Draft findings with management
□ Validate exceptions
□ Determine finding severity
□ Document management response
□ Issue final report
□ Schedule follow-up
```

## Risk-Based Sampling

### Sample Size Determination

```
Population Size Based Sampling:
─────────────────────────────────────────────────────────────────
Population Size    Recommended Sample    Confidence Level
< 50               All items             100%
50-100             25-30 items           95%
101-500            30-50 items           95%
501-1000           50-60 items           95%
> 1000             60-100 items          95%

Risk-Adjusted Sampling:
─────────────────────────────────────────────────────────────────
Control Risk       Sample Increase
High               +50%
Medium             Baseline
Low                -25% (minimum 25 items)
```

### Sampling Criteria for Vulnerabilities

```
Stratified Sample Selection:
─────────────────────────────────────────────────────────────────
Stratum             Selection Criteria       Sample Proportion
Critical CVEs       All (census testing)     100%
High severity       Random selection         30%
Medium severity     Random selection         15%
Low severity        Random selection         5%

Sample Attributes to Test:
□ Detection timeliness (when discovered)
□ Triage accuracy (severity assignment)
□ Remediation timeliness (SLA compliance)
□ Verification completion (fix validated)
□ Documentation completeness
```

## Testing Procedures

### Walkthrough Testing Template

```
Control: Vulnerability Scanning
─────────────────────────────────────────────────────────────────

Walkthrough Procedure:
1. Observe a scan being initiated/scheduled
2. Review scan configuration settings
3. Observe scan results processing
4. Walk through triage decision
5. Follow remediation workflow
6. Verify closure and documentation

Questions for Process Owner:
• How often are scans performed?
• Who reviews scan results?
• How are false positives handled?
• What triggers an emergency response?
• How is scan coverage monitored?

Documentation to Obtain:
□ Scan schedule configuration
□ Sample scan output
□ Triage decision record
□ Remediation ticket
□ Verification evidence
```

### Operating Effectiveness Tests

```
Test 1: Vulnerability Detection Completeness
─────────────────────────────────────────────────────────────────
Objective: Verify scanning covers all in-scope repositories

Procedure:
1. Obtain complete repository inventory
2. Obtain scan coverage report
3. Compare and identify gaps
4. For gaps, determine if exception documented

Expected Evidence:
• Repository inventory list
• Scan coverage dashboard/report
• Exception documentation (if applicable)

Pass Criteria:
• 100% coverage OR
• All gaps have documented, approved exceptions


Test 2: Vulnerability Remediation Timeliness
─────────────────────────────────────────────────────────────────
Objective: Verify vulnerabilities remediated within SLA

Procedure:
1. Select sample of closed vulnerabilities
2. Obtain discovery date and closure date
3. Calculate remediation time
4. Compare to applicable SLA
5. For exceptions, verify approval

Expected Evidence:
• Vulnerability ticket/record
• Discovery timestamp
• Closure timestamp
• SLA definition
• Exception approval (if applicable)

Pass Criteria:
• [X]% within SLA OR
• Exceptions properly approved


Test 3: SBOM Accuracy
─────────────────────────────────────────────────────────────────
Objective: Verify SBOM accurately reflects actual dependencies

Procedure:
1. Select sample of repositories
2. Obtain generated SBOM
3. Independently scan repository
4. Compare SBOM to actual dependencies
5. Identify discrepancies

Expected Evidence:
• Repository SBOM
• Independent scan results
• Reconciliation analysis

Pass Criteria:
• 100% match between SBOM and actual dependencies
• OR discrepancies explained and documented
```

## Professional Standards

### AICPA SSAE 18 (SOC 2)

```
Relevant Trust Services Criteria for Supply Chain:
─────────────────────────────────────────────────────────────────
CC6.6 - The entity implements logical access security measures
        to protect against threats from sources outside its
        system boundaries

CC6.7 - The entity restricts the transmission, movement, and
        removal of information to authorized internal and
        external users and processes

CC7.1 - To meet its objectives, the entity uses detection and
        monitoring procedures to identify changes to configurations
        that result in the introduction of new vulnerabilities

CC7.2 - The entity monitors system components and the operation
        of those components for anomalies that are indicative of
        malicious acts, natural disasters, and errors affecting
        the entity's ability to meet its objectives

CC8.1 - The entity authorizes, designs, develops or acquires,
        configures, documents, tests, approves, and implements
        changes to infrastructure, data, software, and procedures
```

### IIA Standards

```
IPPF Standard 2100 - Nature of Work
─────────────────────────────────────────────────────────────────
The internal audit activity must evaluate and contribute to the
improvement of the organization's governance, risk management,
and control processes

Application to Supply Chain:
• Evaluate supply chain risk management effectiveness
• Assess controls over third-party software
• Review vulnerability management program
• Test incident response capabilities
```

## Quality Assurance

### Audit Workpaper Standards

```
Documentation Requirements:
─────────────────────────────────────────────────────────────────
Each workpaper must include:

□ Purpose - What the workpaper demonstrates
□ Scope - Period and population covered
□ Source - Where evidence was obtained
□ Procedure - Steps performed
□ Results - Findings from testing
□ Conclusion - Pass/fail determination

Cross-Reference Requirements:
□ Link to audit program step
□ Link to control being tested
□ Link to finding (if exception)
□ Review evidence attached
```

### Finding Classification

```
Finding Severity Matrix:
─────────────────────────────────────────────────────────────────
                    LIKELIHOOD
                    Low      Medium    High
              High  │ Medium │ High    │ Critical │
    IMPACT    Med   │ Low    │ Medium  │ High     │
              Low   │ Info   │ Low     │ Medium   │


Finding Categories:
─────────────────────────────────────────────────────────────────
• Control Deficiency - Control does not operate as designed
• Design Gap - Control is missing or inadequate by design
• Operating Exception - Control exists but was not followed
• Observation - Improvement opportunity (not a deficiency)
```

## Quick Reference

### Auditor Checklist

```
Pre-Audit:
□ Review applicable standards and regulations
□ Understand the technology environment
□ Identify key risks and controls
□ Develop risk-based audit program
□ Coordinate with management

During Audit:
□ Document walkthroughs completely
□ Maintain independence and objectivity
□ Test both design and operating effectiveness
□ Document exceptions clearly
□ Validate findings with management

Post-Audit:
□ Clear all review notes
□ Obtain management responses
□ Issue findings within timeline
□ Schedule follow-up testing
□ Update risk assessment
```
