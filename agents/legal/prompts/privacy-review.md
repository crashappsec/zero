# Privacy Review Prompt

## Purpose
Generate a data privacy compliance assessment for a software application.

## Output Format

```markdown
# Data Privacy Assessment

**Application:** {{application_name}}
**Date:** {{date}}
**Reviewer:** Harper (General Counsel Agent)

## Executive Summary

{{brief_summary_of_privacy_posture_and_key_findings}}

## Data Inventory

### Personal Data Collected

| Data Element | Category | Legal Basis | Retention | Encrypted |
|--------------|----------|-------------|-----------|-----------|
| {{element}} | {{PII/sensitive/special}} | {{consent/contract/legitimate}} | {{period}} | {{yes/no}} |

### Data Flows

| Source | Destination | Purpose | Safeguards |
|--------|-------------|---------|------------|
| {{source}} | {{destination}} | {{purpose}} | {{measures}} |

## Regulatory Applicability

### GDPR
- **Applicable:** {{yes/no/likely}}
- **Basis:** {{why_applicable}}
- **Key Requirements:** {{specific_requirements}}

### CCPA/CPRA
- **Applicable:** {{yes/no/likely}}
- **Basis:** {{why_applicable}}
- **Key Requirements:** {{specific_requirements}}

### Other Regulations
{{List other applicable regulations}}

## Compliance Assessment

### Data Subject Rights

| Right | Implemented | Method | Gap |
|-------|-------------|--------|-----|
| Access | {{yes/no/partial}} | {{how}} | {{gap_if_any}} |
| Deletion | {{yes/no/partial}} | {{how}} | {{gap_if_any}} |
| Portability | {{yes/no/partial}} | {{how}} | {{gap_if_any}} |
| Opt-out | {{yes/no/partial}} | {{how}} | {{gap_if_any}} |

### Consent Management

- **Consent captured:** {{yes/no}}
- **Granular options:** {{yes/no}}
- **Withdrawal mechanism:** {{yes/no}}
- **Records maintained:** {{yes/no}}

### Technical Safeguards

| Measure | Status | Notes |
|---------|--------|-------|
| Encryption at rest | {{implemented/missing/partial}} | {{details}} |
| Encryption in transit | {{implemented/missing/partial}} | {{details}} |
| Access controls | {{implemented/missing/partial}} | {{details}} |
| Audit logging | {{implemented/missing/partial}} | {{details}} |
| Data minimization | {{implemented/missing/partial}} | {{details}} |

## Third-Party Processors

| Vendor | Data Shared | DPA Status | Location | Risk |
|--------|-------------|------------|----------|------|
| {{vendor}} | {{data_types}} | {{signed/needed/na}} | {{location}} | {{risk_level}} |

## International Transfers

### Transfer Mechanisms

| Destination | Data Types | Mechanism | Status |
|-------------|------------|-----------|--------|
| {{country}} | {{data}} | {{SCCs/adequacy/BCR}} | {{compliant/action_needed}} |

## Gap Analysis

### Critical Gaps
{{List critical compliance gaps}}

### Medium Priority Gaps
{{List medium priority gaps}}

### Low Priority Gaps
{{List lower priority gaps}}

## Recommendations

### Immediate Actions
1. {{urgent_action}}

### Documentation Needed
1. {{document_requirement}}

### Process Improvements
1. {{process_recommendation}}

## Disclaimer

This assessment provides general information about data privacy compliance based on code and configuration analysis. It is not legal advice. Privacy requirements vary by jurisdiction and specific circumstances. Consult a qualified privacy attorney for definitive guidance.
```

## Analysis Guidelines

1. **Map data flows**: Understand what data goes where
2. **Identify legal bases**: Consent, contract, legitimate interest
3. **Check vendor relationships**: DPAs and international transfers
4. **Review retention**: Is data kept longer than necessary?
5. **Assess security measures**: Encryption, access controls, logging
