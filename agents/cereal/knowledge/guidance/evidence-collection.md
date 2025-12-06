<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
SPDX-License-Identifier: GPL-3.0
-->

# Evidence Collection Guide for Supply Chain Security Audits

## Evidence Standards

### Qualities of Audit Evidence

```
RRIA Framework for Evidence Quality:
─────────────────────────────────────────────────────────────────

Relevant
• Evidence directly supports the control being tested
• Addresses the audit objective
• Pertains to the audit period

Reliable
• Source is independent (preferable)
• System-generated (vs manually created)
• Obtained directly by auditor
• Tamper-evident

Independent
• From party independent of control performance
• Cross-corroborated where possible
• Not solely from control owner

Appropriate
• Sufficient quantity for conclusion
• Timely (from within audit period)
• Complete (not partial or redacted)
```

### Evidence Hierarchy

```
Strongest Evidence (Most Reliable)
────────────────────────────────────
1. External third-party confirmation
2. System-generated reports (with timestamps)
3. Original documents examined by auditor
4. Re-performance by auditor
5. Observation of control performance

↓

Weaker Evidence (Less Reliable)
────────────────────────────────────
6. Management-prepared analysis
7. Internal reports without timestamps
8. Inquiry responses alone
9. Undated or unsigned documents
10. Photocopies without verification
```

## Supply Chain Evidence Types

### Vulnerability Scanning Evidence

```
Required Evidence:
─────────────────────────────────────────────────────────────────
1. Scan Configuration
   □ Tool configuration settings
   □ Scan schedule/frequency
   □ Coverage scope definition
   □ Data source configuration (NVD, etc.)

2. Scan Execution
   □ Scan logs with timestamps
   □ Completion status records
   □ Error/failure logs
   □ Coverage reports

3. Scan Results
   □ Vulnerability findings
   □ Severity assignments
   □ Affected components
   □ Recommended fixes

Collection Method:
• Export directly from scanning tool
• Include system timestamp
• Capture full report (not summary)
• Obtain both success and failure examples


Example Evidence Package:
┌─────────────────────────────────────────────────────────────────────────┐
│  Evidence: Vulnerability Scanning Execution                             │
│  Reference: VS-001                                                      │
├─────────────────────────────────────────────────────────────────────────┤
│  Source: [Scanning Tool Name] - Admin Console                          │
│  Obtained By: [Auditor Name]                                           │
│  Date Obtained: [Date]                                                  │
│  Period Covered: [Start] to [End]                                       │
│                                                                         │
│  Contents:                                                              │
│  • VS-001a: Scan configuration export (JSON)                           │
│  • VS-001b: Scan execution log (Week 1-4)                              │
│  • VS-001c: Sample scan results (5 repositories)                       │
│  • VS-001d: Coverage report                                            │
│                                                                         │
│  Verification: Auditor logged into system directly to obtain           │
└─────────────────────────────────────────────────────────────────────────┘
```

### SBOM Evidence

```
Required Evidence:
─────────────────────────────────────────────────────────────────
1. SBOM Policy/Procedure
   □ Format requirements (SPDX/CycloneDX)
   □ Generation trigger (build/release)
   □ Required fields
   □ Storage location

2. SBOM Generation Process
   □ Tool configuration
   □ CI/CD integration evidence
   □ Automation logs

3. SBOM Artifacts
   □ Actual SBOM files
   □ Associated release/build identifiers
   □ Generation timestamps
   □ Digital signatures (if applicable)

4. Accuracy Verification
   □ Independent dependency scan
   □ Comparison analysis
   □ Discrepancy documentation


SBOM Validation Checklist:
□ Format is valid (schema validation)
□ All NTIA minimum fields present
□ Supplier name populated
□ Component names populated
□ Versions populated
□ Unique identifiers present
□ Relationships defined
□ Author identified
□ Timestamp present
□ Matches actual dependencies
```

### Remediation Evidence

```
Required Evidence per Vulnerability:
─────────────────────────────────────────────────────────────────
1. Discovery/Detection
   □ Initial scan finding
   □ Discovery timestamp
   □ CVE identifier
   □ Affected component/version

2. Triage
   □ Triage record/ticket
   □ Severity assignment
   □ Triage timestamp
   □ Assignee

3. Remediation
   □ Fix implementation (PR/commit)
   □ Code review approval
   □ Test execution results
   □ Deployment record

4. Verification
   □ Post-fix scan
   □ Closure timestamp
   □ Verification by independent party


Timeline Documentation:
┌────────────────────────────────────────────────────────────────────────┐
│  CVE: CVE-2024-XXXXX                                                   │
│  Component: axios@0.21.1                                               │
│  Severity: High                                                        │
├────────────────────────────────────────────────────────────────────────┤
│                                                                        │
│  TIMELINE                                                              │
│  ──────────────────────────────────────────────────────────────       │
│  [2024-03-15 08:00] Detected by automated scan                         │
│  [2024-03-15 10:30] Triage completed, assigned High severity           │
│  [2024-03-15 14:00] Fix PR created (#1234)                            │
│  [2024-03-15 16:00] PR approved by reviewer                           │
│  [2024-03-15 17:30] Deployed to production                            │
│  [2024-03-16 08:00] Verified by post-deployment scan                   │
│                                                                        │
│  Total Remediation Time: 24 hours                                      │
│  SLA: 7 days | Status: COMPLIANT                                       │
│                                                                        │
├────────────────────────────────────────────────────────────────────────┤
│  EVIDENCE ATTACHED:                                                    │
│  • REM-001a: Initial scan finding                                      │
│  • REM-001b: Triage ticket (JIRA-1234)                                │
│  • REM-001c: Pull request #1234                                        │
│  • REM-001d: Deployment log                                            │
│  • REM-001e: Verification scan                                         │
└────────────────────────────────────────────────────────────────────────┘
```

## Collection Methods

### System Export Procedures

```
Scanning Tool Exports:
─────────────────────────────────────────────────────────────────
Snyk:
• Reports > Export > Select JSON format
• Include vulnerability details
• Capture license findings
• Export organization settings

GitHub Advanced Security:
• Security > Code scanning > Export
• Dependabot > Export alerts
• Include dismissed items with reasons

npm audit:
• Run: npm audit --json > audit_report.json
• Include: package-lock.json
• Run: npm ls --all --json > dependencies.json

GitLab:
• Security Dashboard > Export
• Pipeline > Download artifacts
• Include MR approval records
```

### Screenshot Evidence Standards

```
Screenshot Requirements:
─────────────────────────────────────────────────────────────────
□ Capture full screen (not cropped)
□ Include URL bar/address
□ Include system date/time
□ Include user identification
□ Use PNG format (no compression artifacts)
□ Do not edit or annotate originals
□ Create separate annotated copy if needed

Naming Convention:
[ControlID]_[Description]_[YYYYMMDD].png
Example: VM01_ScanConfig_20241115.png
```

### Interview Evidence

```
Interview Documentation:
─────────────────────────────────────────────────────────────────
Pre-Interview:
□ Prepare specific questions
□ Identify corroborating evidence needed
□ Schedule appropriate participants

During Interview:
□ Document date, time, participants
□ Record responses verbatim where critical
□ Note items for follow-up
□ Identify documents to request

Post-Interview:
□ Provide summary for confirmation
□ Obtain requested documents
□ Cross-reference to other evidence
□ Note any inconsistencies


Interview Memo Template:
┌─────────────────────────────────────────────────────────────────────────┐
│  INTERVIEW MEMORANDUM                                                   │
│  Reference: INT-001                                                     │
├─────────────────────────────────────────────────────────────────────────┤
│  Date: [Date]                                                           │
│  Time: [Start] - [End]                                                  │
│  Location/Method: [In-person/Video/Phone]                               │
│                                                                         │
│  Participants:                                                          │
│  • [Name], [Title] (Interviewee)                                       │
│  • [Name], [Title] (Auditor)                                           │
│                                                                         │
│  Purpose: [Topic/Control being discussed]                               │
│                                                                         │
│  DISCUSSION SUMMARY:                                                    │
│  [Key points discussed]                                                 │
│                                                                         │
│  FOLLOW-UP ITEMS:                                                       │
│  1. [Document/evidence to be provided]                                  │
│  2. [Clarification needed]                                              │
│                                                                         │
│  Prepared By: [Name]                                                    │
│  Date: [Date]                                                           │
└─────────────────────────────────────────────────────────────────────────┘
```

## Evidence Organization

### Workpaper Index Structure

```
Audit: Supply Chain Security 2024
─────────────────────────────────────────────────────────────────
A - Planning
    A-1     Engagement letter
    A-2     Audit program
    A-3     Risk assessment
    A-4     Materiality memo
    A-5     Timeline/milestones

B - Understanding
    B-1     Process narratives
    B-2     System documentation
    B-3     Control inventory
    B-4     Prior audit results

C - Testing - Vulnerability Management
    C-1     VM-1 Scanning (Design)
    C-2     VM-1 Scanning (Operating)
    C-3     VM-2 Triage (Design)
    C-4     VM-2 Triage (Operating)
    C-5     VM-3 Remediation (Design)
    C-6     VM-3 Remediation (Operating)

D - Testing - SBOM
    D-1     SBOM-1 Generation (Design)
    D-2     SBOM-1 Generation (Operating)
    D-3     SBOM-2 Accuracy (Operating)

E - Testing - Change Management
    E-1     CM-1 Authorization (Design)
    E-2     CM-1 Authorization (Operating)

F - Testing - Third-Party Risk
    F-1     TPR-1 Assessment (Design)
    F-2     TPR-1 Assessment (Operating)

G - Findings
    G-1     Finding summary
    G-2     Individual finding workpapers
    G-3     Management responses

H - Conclusion
    H-1     Summary of testing results
    H-2     Opinion memo
    H-3     Report draft
```

### Evidence Request List (PBC)

```
╔═══════════════════════════════════════════════════════════════════════════╗
║  PREPARED BY CLIENT (PBC) LIST                                            ║
║  Audit: Supply Chain Security                                             ║
║  Period: [Date] to [Date]                                                 ║
╠═══════════════════════════════════════════════════════════════════════════╣

│ #  │ Description                          │ Due Date │ Status │ Owner    │
├────┼──────────────────────────────────────┼──────────┼────────┼──────────┤
│ 1  │ Vulnerability management policy      │ [Date]   │        │ [Name]   │
│ 2  │ Scan configuration documentation     │ [Date]   │        │ [Name]   │
│ 3  │ Scan execution logs (period)         │ [Date]   │        │ [Name]   │
│ 4  │ Vulnerability inventory (period)     │ [Date]   │        │ [Name]   │
│ 5  │ Remediation tickets (sample)         │ [Date]   │        │ [Name]   │
│ 6  │ SBOM generation procedure            │ [Date]   │        │ [Name]   │
│ 7  │ Sample SBOMs (5 releases)            │ [Date]   │        │ [Name]   │
│ 8  │ Repository inventory                 │ [Date]   │        │ [Name]   │
│ 9  │ Change management policy             │ [Date]   │        │ [Name]   │
│ 10 │ Dependency update PRs (sample)       │ [Date]   │        │ [Name]   │
│ 11 │ Vendor assessment records            │ [Date]   │        │ [Name]   │
│ 12 │ Critical dependency inventory        │ [Date]   │        │ [Name]   │
│ 13 │ Exception/risk acceptance records    │ [Date]   │        │ [Name]   │
│ 14 │ Training records (security team)     │ [Date]   │        │ [Name]   │
│ 15 │ Org chart - security function        │ [Date]   │        │ [Name]   │

╚═══════════════════════════════════════════════════════════════════════════╝
```

## Special Considerations

### Sensitive Data Handling

```
Handling Requirements:
─────────────────────────────────────────────────────────────────
• Vulnerability details may be sensitive - handle appropriately
• Do not include actual credentials or secrets in workpapers
• Redact PII if present in evidence
• Follow organization's data classification
• Secure evidence storage and transmission
• Destroy/return evidence per agreement
```

### System Access Documentation

```
Access Log Template:
─────────────────────────────────────────────────────────────────
System: [System Name]
Access Granted: [Date]
Access Revoked: [Date]
Account Used: [Username]
Access Level: [Read/Write/Admin]
Purpose: [Audit testing]
Authorized By: [Name/Title]

Activities Performed:
[Date/Time] - [Activity]
[Date/Time] - [Activity]

Evidence Obtained:
• [File/Report name]
• [File/Report name]
```

### Chain of Custody

```
For Sensitive/Legal Evidence:
─────────────────────────────────────────────────────────────────
Evidence Item: [Description]
Original Source: [System/Person]

Transfer Record:
┌────────────┬─────────────┬─────────────┬───────────────────────┐
│ Date/Time  │ From        │ To          │ Purpose               │
├────────────┼─────────────┼─────────────┼───────────────────────┤
│[DateTime]  │[System]     │[Auditor]    │Initial collection     │
│[DateTime]  │[Auditor]    │[Workpaper]  │Documentation          │
│[DateTime]  │[Auditor]    │[Reviewer]   │Review                 │
└────────────┴─────────────┴─────────────┴───────────────────────┘

Storage Location: [Secure location/system]
Retention Period: [Per policy]
Destruction Date: [If applicable]
```

## Quick Reference

### Evidence Collection Checklist

```
Per Evidence Item:
□ Verify relevance to control/objective
□ Obtain from reliable source
□ Include date/timestamp
□ Document source
□ Verify completeness
□ Cross-reference to workpaper
□ Store securely

Per Control Tested:
□ Design evidence obtained
□ Operating evidence obtained
□ Sufficient sample size
□ Period coverage complete
□ Exceptions documented
□ Conclusion supported
```
