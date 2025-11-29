<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
SPDX-License-Identifier: GPL-3.0
-->

# Control Testing Procedures for Supply Chain Security

## Control Testing Framework

### Types of Testing

```
Design Effectiveness Testing
─────────────────────────────────────────────────────────────────
Purpose: Determine if control is suitably designed to meet objective

Methods:
• Inquiry - Interview process owners
• Observation - Watch control execution
• Inspection - Review documentation

Questions to Answer:
1. Is the control designed to address the risk?
2. Are the steps defined clearly?
3. Are responsibilities assigned?
4. Would the control prevent/detect issues if followed?


Operating Effectiveness Testing
─────────────────────────────────────────────────────────────────
Purpose: Determine if control operated consistently over the period

Methods:
• Re-performance - Execute control independently
• Sample testing - Test sample of transactions
• Data analytics - Analyze full population

Questions to Answer:
1. Was the control performed as designed?
2. Was it performed consistently throughout the period?
3. Did exceptions follow the defined process?
4. Did the control achieve its objective?
```

### Control Frequency Impact

```
Control Frequency    Minimum Sample Size    Testing Approach
─────────────────────────────────────────────────────────────────
Annual               1                      Test the occurrence
Quarterly            2-4                    Test each quarter
Monthly              2-5                    Test spread across period
Weekly               5-15                   Test from multiple weeks
Daily                20-40                  Random selection
Per transaction      25-60                  Based on population
Continuous           Full population        Data analytics
```

## Vulnerability Management Controls

### VM-1: Vulnerability Scanning

```
╔═══════════════════════════════════════════════════════════════════════════╗
║  CONTROL: VM-1 Vulnerability Scanning                                     ║
╠═══════════════════════════════════════════════════════════════════════════╣
║                                                                           ║
║  Control Statement:                                                       ║
║  Automated vulnerability scanning is performed [frequency] on all         ║
║  production repositories to identify known security vulnerabilities       ║
║  in third-party dependencies.                                             ║
║                                                                           ║
║  Control Owner: [Name/Role]                                               ║
║  Frequency: [Continuous/Daily/Weekly]                                     ║
║  Evidence Location: [System/Repository]                                   ║
║                                                                           ║
╚═══════════════════════════════════════════════════════════════════════════╝

DESIGN EFFECTIVENESS TEST
─────────────────────────────────────────────────────────────────
Procedure:
1. Obtain vulnerability scanning policy/procedure
2. Identify scanning tool(s) in use
3. Review scan configuration settings
4. Verify coverage requirements defined
5. Confirm alerting/notification configured

Workpaper Documentation:
□ Policy document reference
□ Tool name and version
□ Configuration screenshots
□ Coverage requirements
□ Alert configuration

Design Questions:
• What triggers a scan?
• What data sources are used for CVE data?
• How are results stored and retained?
• Who receives scan notifications?
• How are failed scans handled?


OPERATING EFFECTIVENESS TEST
─────────────────────────────────────────────────────────────────
Procedure:
1. Obtain scan execution logs for period
2. Select sample of [X] weeks/days
3. Verify scans executed as scheduled
4. Verify all in-scope repos scanned
5. Verify results recorded

Test Steps:
Step 1: Obtain scan schedule
        Expected: Defined schedule exists

Step 2: Pull scan execution logs
        Sample Period: [Start] to [End]
        Expected: Logs available for full period

Step 3: Verify execution frequency
        Test: Compare actual runs to schedule
        Expected: ≥95% scheduled scans completed

Step 4: Verify coverage
        Test: Match scanned repos to inventory
        Expected: 100% coverage (or documented exceptions)

Exception Handling:
If exception found:
• Document specific exception
• Determine root cause
• Assess impact on control effectiveness
• Determine if isolated or systemic
```

### VM-2: Vulnerability Triage

```
╔═══════════════════════════════════════════════════════════════════════════╗
║  CONTROL: VM-2 Vulnerability Triage                                       ║
╠═══════════════════════════════════════════════════════════════════════════╣
║                                                                           ║
║  Control Statement:                                                       ║
║  Identified vulnerabilities are triaged within [X] hours and assigned     ║
║  a severity rating based on defined criteria.                             ║
║                                                                           ║
╚═══════════════════════════════════════════════════════════════════════════╝

DESIGN EFFECTIVENESS TEST
─────────────────────────────────────────────────────────────────
Procedure:
1. Obtain triage procedure documentation
2. Review severity classification criteria
3. Verify SLA definitions by severity
4. Confirm assignment procedures
5. Review escalation paths

Design Questions:
• How is severity determined?
• Who performs initial triage?
• What factors influence priority?
• How are conflicts resolved?
• When is escalation required?


OPERATING EFFECTIVENESS TEST
─────────────────────────────────────────────────────────────────
Sample Selection:
• Population: All vulnerabilities identified in period
• Sample: [X] items stratified by severity
• Selection: Random within strata

Test Steps:
Step 1: For each sample item, obtain:
        □ Initial detection timestamp
        □ Triage completion timestamp
        □ Assigned severity
        □ Triage justification

Step 2: Calculate triage time
        Expected: ≤[X] hours per policy

Step 3: Verify severity assignment
        Re-perform: Apply criteria independently
        Expected: Consistent with auditor assessment

Step 4: Verify proper assignment
        Expected: Ticket assigned to appropriate owner

Results Documentation:
┌──────────┬───────────┬───────────┬──────────┬────────┬────────┐
│Sample #  │Detection  │Triage     │Time (hrs)│SLA Met │Correct │
├──────────┼───────────┼───────────┼──────────┼────────┼────────┤
│    1     │[DateTime] │[DateTime] │   4.2    │  Yes   │  Yes   │
│    2     │[DateTime] │[DateTime] │   6.8    │  Yes   │  Yes   │
│    3     │[DateTime] │[DateTime] │  12.1    │  No    │  Yes   │
└──────────┴───────────┴───────────┴──────────┴────────┴────────┘
```

### VM-3: Vulnerability Remediation

```
╔═══════════════════════════════════════════════════════════════════════════╗
║  CONTROL: VM-3 Vulnerability Remediation                                  ║
╠═══════════════════════════════════════════════════════════════════════════╣
║                                                                           ║
║  Control Statement:                                                       ║
║  Vulnerabilities are remediated within defined SLAs:                      ║
║  - Critical: [X] hours                                                    ║
║  - High: [X] days                                                         ║
║  - Medium: [X] days                                                       ║
║  - Low: [X] days                                                          ║
║                                                                           ║
╚═══════════════════════════════════════════════════════════════════════════╝

OPERATING EFFECTIVENESS TEST
─────────────────────────────────────────────────────────────────
Sample Selection:
• Population: All closed vulnerabilities in period
• Stratification:
  - Critical: All items (census)
  - High: 30% random sample
  - Medium: 15% random sample
  - Low: 5% random sample

Test Procedure:
Step 1: For each sample, obtain:
        □ Triage date/severity
        □ Remediation completion date
        □ Fix verification evidence
        □ Applicable SLA

Step 2: Calculate remediation time
        Formula: Completion Date - Triage Date

Step 3: Compare to SLA
        Expected: ≤SLA time for severity

Step 4: Verify fix effectiveness
        Evidence: Re-scan showing resolution

Step 5: For SLA breaches, verify:
        □ Exception approved
        □ Risk acceptance documented
        □ Compensating controls noted

Results Summary:
┌─────────┬────────────┬─────────┬──────────┬───────────────┐
│Severity │Population  │Sampled  │In SLA    │Compliance %   │
├─────────┼────────────┼─────────┼──────────┼───────────────┤
│Critical │     3      │    3    │    3     │    100%       │
│High     │    15      │    5    │    4     │     80%       │
│Medium   │    28      │    5    │    5     │    100%       │
│Low      │    22      │    5    │    5     │    100%       │
├─────────┼────────────┼─────────┼──────────┼───────────────┤
│Total    │    68      │   18    │   17     │     94%       │
└─────────┴────────────┴─────────┴──────────┴───────────────┘
```

## SBOM Controls

### SBOM-1: SBOM Generation

```
╔═══════════════════════════════════════════════════════════════════════════╗
║  CONTROL: SBOM-1 SBOM Generation                                          ║
╠═══════════════════════════════════════════════════════════════════════════╣
║                                                                           ║
║  Control Statement:                                                       ║
║  A Software Bill of Materials (SBOM) is generated for each software       ║
║  release in [format] containing all required NTIA minimum elements.       ║
║                                                                           ║
╚═══════════════════════════════════════════════════════════════════════════╝

DESIGN EFFECTIVENESS TEST
─────────────────────────────────────────────────────────────────
Procedure:
1. Review SBOM generation procedure
2. Identify generation tool(s)
3. Verify required fields configured
4. Confirm automation in CI/CD
5. Review storage/retention

Required NTIA Minimum Elements:
□ Supplier name
□ Component name
□ Component version
□ Unique identifier
□ Dependency relationship
□ Author of SBOM data
□ Timestamp


OPERATING EFFECTIVENESS TEST
─────────────────────────────────────────────────────────────────
Sample Selection:
• Population: All releases in audit period
• Sample: [X] releases across different repos/apps

Test Procedure:
Step 1: For each sample release, obtain:
        □ Release identifier
        □ Associated SBOM
        □ Generation timestamp

Step 2: Verify SBOM exists
        Expected: SBOM present for each release

Step 3: Verify completeness
        Test: Check for all NTIA minimum elements
        Expected: All fields populated

Step 4: Verify accuracy
        Test: Compare SBOM to actual build dependencies
        Expected: Complete match

Accuracy Testing Detail:
For [X] SBOMs, independently scan source and compare:
• Dependencies listed in SBOM but not in source: [Count]
• Dependencies in source but not in SBOM: [Count]
• Version mismatches: [Count]
• Expected: Zero discrepancies
```

## Change Management Controls

### CM-1: Dependency Updates

```
╔═══════════════════════════════════════════════════════════════════════════╗
║  CONTROL: CM-1 Dependency Update Authorization                            ║
╠═══════════════════════════════════════════════════════════════════════════╣
║                                                                           ║
║  Control Statement:                                                       ║
║  All dependency updates are reviewed and approved through the             ║
║  organization's change management process prior to deployment             ║
║  to production.                                                           ║
║                                                                           ║
╚═══════════════════════════════════════════════════════════════════════════╝

OPERATING EFFECTIVENESS TEST
─────────────────────────────────────────────────────────────────
Sample Selection:
• Population: All dependency updates deployed in period
• Sample: [X] updates across different types
  - Patch updates: [X]
  - Minor updates: [X]
  - Major updates: [X]
  - Security updates: [X]

Test Procedure:
Step 1: For each sample, obtain:
        □ Change request/PR
        □ Review/approval records
        □ Testing evidence
        □ Deployment record

Step 2: Verify authorization
        Expected: Approval by authorized individual
        Evidence: Approval record with timestamp

Step 3: Verify testing
        Expected: CI/CD tests passed
        Evidence: Test execution logs

Step 4: Verify segregation
        Expected: Approver ≠ Developer
        Test: Compare submitter to approver

Step 5: Verify sequence
        Expected: Approval before deployment
        Test: Compare approval timestamp to deploy timestamp

Results Matrix:
┌────────┬─────────────┬───────────┬────────────┬─────────────┬────────┐
│Sample  │Update Type  │Approved   │Tested      │Segregation  │Sequence│
├────────┼─────────────┼───────────┼────────────┼─────────────┼────────┤
│   1    │Patch        │   Yes     │   Yes      │    Yes      │  Yes   │
│   2    │Minor        │   Yes     │   Yes      │    Yes      │  Yes   │
│   3    │Major        │   Yes     │   Yes      │    Yes      │  Yes   │
│   4    │Security     │   Yes     │   Yes      │    N/A*     │  Yes   │
└────────┴─────────────┴───────────┴────────────┴─────────────┴────────┘
*Emergency change - documented exception
```

## Third-Party Risk Controls

### TPR-1: Vendor Security Assessment

```
╔═══════════════════════════════════════════════════════════════════════════╗
║  CONTROL: TPR-1 Vendor Security Assessment                                ║
╠═══════════════════════════════════════════════════════════════════════════╣
║                                                                           ║
║  Control Statement:                                                       ║
║  Critical third-party software dependencies are assessed for security     ║
║  risks prior to adoption and [annually/periodically] thereafter.          ║
║                                                                           ║
╚═══════════════════════════════════════════════════════════════════════════╝

DESIGN EFFECTIVENESS TEST
─────────────────────────────────────────────────────────────────
Procedure:
1. Review vendor assessment policy
2. Identify assessment criteria
3. Review assessment template/checklist
4. Verify approval requirements
5. Confirm reassessment schedule

Assessment Criteria Should Include:
□ Security track record
□ Maintenance activity
□ Vulnerability response history
□ License compatibility
□ Community health (open source)


OPERATING EFFECTIVENESS TEST
─────────────────────────────────────────────────────────────────
Sample Selection:
• Population: Critical dependencies in use
• Sample: [X] dependencies
• Include: Mix of new and existing

Test Procedure:
Step 1: Verify assessment exists
        Expected: Assessment on file

Step 2: Verify timeliness
        For new: Before production use
        For existing: Within assessment period

Step 3: Verify completeness
        Expected: All criteria addressed

Step 4: Verify approval
        Expected: Appropriate authority approved

Step 5: Verify risk documentation
        Expected: Identified risks documented
        Expected: Mitigations defined

Results:
┌────────────────────┬────────────┬──────────┬──────────┬──────────┐
│Dependency          │Assessment  │Current   │Complete  │Approved  │
├────────────────────┼────────────┼──────────┼──────────┼──────────┤
│react               │2024-03-15  │  Yes     │  Yes     │  Yes     │
│express             │2024-06-20  │  Yes     │  Yes     │  Yes     │
│lodash              │2023-08-10  │  No*     │  Yes     │  Yes     │
└────────────────────┴────────────┴──────────┴──────────┴──────────┘
*Exception: Reassessment overdue
```

## Testing Workpaper Templates

### Standard Workpaper Format

```
┌─────────────────────────────────────────────────────────────────────────┐
│  WORKPAPER: [WP Reference Number]                                       │
│  Control: [Control ID and Name]                                         │
│  Prepared By: [Name] Date: [Date]                                       │
│  Reviewed By: [Name] Date: [Date]                                       │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  OBJECTIVE:                                                             │
│  [State the testing objective]                                          │
│                                                                         │
│  SCOPE:                                                                 │
│  Period: [Start] to [End]                                               │
│  Population: [Description]                                              │
│  Sample: [Size and selection method]                                    │
│                                                                         │
│  SOURCE OF EVIDENCE:                                                    │
│  [System/person where evidence obtained]                                │
│                                                                         │
│  PROCEDURE PERFORMED:                                                   │
│  [Detailed steps performed]                                             │
│                                                                         │
│  RESULTS:                                                               │
│  [Findings from testing]                                                │
│                                                                         │
│  EXCEPTIONS:                                                            │
│  [Any exceptions identified]                                            │
│                                                                         │
│  CONCLUSION:                                                            │
│  [Pass/Fail and rationale]                                              │
│                                                                         │
│  ATTACHMENTS:                                                           │
│  [List of supporting evidence]                                          │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Exception Documentation Template

```
┌─────────────────────────────────────────────────────────────────────────┐
│  EXCEPTION: [Exception Reference Number]                                │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  Control Tested: [Control ID]                                           │
│  Sample Item: [Identifier]                                              │
│                                                                         │
│  DESCRIPTION OF EXCEPTION:                                              │
│  [What was the deviation from expected]                                 │
│                                                                         │
│  EXPECTED:                                                              │
│  [What should have happened]                                            │
│                                                                         │
│  ACTUAL:                                                                │
│  [What actually happened]                                               │
│                                                                         │
│  ROOT CAUSE:                                                            │
│  [Why did this happen]                                                  │
│                                                                         │
│  IMPACT ASSESSMENT:                                                     │
│  • Financial: [None/Low/Medium/High]                                    │
│  • Compliance: [None/Low/Medium/High]                                   │
│  • Operational: [None/Low/Medium/High]                                  │
│                                                                         │
│  ISOLATED OR SYSTEMIC:                                                  │
│  [Assessment and rationale]                                             │
│                                                                         │
│  MANAGEMENT RESPONSE:                                                   │
│  [Corrective action planned]                                            │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

## Quick Reference

### Testing Checklist

```
Before Testing:
□ Understand control objective
□ Review prior period results
□ Determine sample size
□ Identify evidence sources
□ Prepare testing templates

During Testing:
□ Document all steps performed
□ Obtain complete evidence
□ Note any limitations
□ Identify all exceptions
□ Validate with control owner

After Testing:
□ Conclude on effectiveness
□ Document exceptions fully
□ Determine root causes
□ Assess systemic vs. isolated
□ Clear review notes
```
