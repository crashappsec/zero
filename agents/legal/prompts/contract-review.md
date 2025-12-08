# Contract Review Prompt

## Purpose
Generate a technical contract review focusing on software-related terms and risks.

## Output Format

```markdown
# Technical Contract Review

**Document:** {{contract_name}}
**Counterparty:** {{party_name}}
**Type:** {{SaaS Agreement/License/DPA/API Terms/etc}}
**Date:** {{date}}
**Reviewer:** Harper (General Counsel Agent)

## Executive Summary

{{brief_summary_of_contract_and_key_concerns}}

## Contract Classification

- **Agreement Type:** {{type}}
- **Our Role:** {{customer/vendor/partner}}
- **Data Involved:** {{yes/no}} - {{types_if_yes}}
- **Integration Level:** {{API/SDK/data_exchange/none}}

## Key Terms Analysis

### Service Levels

| Metric | Committed | Penalty | Industry Standard | Assessment |
|--------|-----------|---------|-------------------|------------|
| Uptime | {{%}} | {{remedy}} | {{typical}} | {{adequate/concerning}} |
| Response Time | {{time}} | {{remedy}} | {{typical}} | {{adequate/concerning}} |
| Support | {{level}} | {{escalation}} | {{typical}} | {{adequate/concerning}} |

### Data Terms

| Aspect | Contract Terms | Risk | Recommendation |
|--------|----------------|------|----------------|
| Data ownership | {{terms}} | {{risk_level}} | {{action}} |
| Data use rights | {{terms}} | {{risk_level}} | {{action}} |
| Data location | {{terms}} | {{risk_level}} | {{action}} |
| Subprocessors | {{terms}} | {{risk_level}} | {{action}} |
| Return/deletion | {{terms}} | {{risk_level}} | {{action}} |

### Security Requirements

| Requirement | Present | Adequate | Notes |
|-------------|---------|----------|-------|
| Encryption standards | {{yes/no}} | {{yes/no}} | {{details}} |
| Access controls | {{yes/no}} | {{yes/no}} | {{details}} |
| Audit rights | {{yes/no}} | {{yes/no}} | {{details}} |
| Breach notification | {{yes/no}} | {{yes/no}} | {{details}} |
| Certifications | {{yes/no}} | {{yes/no}} | {{details}} |

### Intellectual Property

| Aspect | Terms | Risk | Notes |
|--------|-------|------|-------|
| Pre-existing IP | {{terms}} | {{risk}} | {{notes}} |
| Developed IP | {{terms}} | {{risk}} | {{notes}} |
| License scope | {{terms}} | {{risk}} | {{notes}} |
| Open source | {{terms}} | {{risk}} | {{notes}} |

### Liability and Risk

| Provision | Terms | Assessment |
|-----------|-------|------------|
| Liability cap | {{amount/formula}} | {{adequate/inadequate}} |
| Indemnification | {{scope}} | {{balanced/one-sided}} |
| Insurance | {{requirements}} | {{adequate/inadequate}} |
| Exclusions | {{carve-outs}} | {{reasonable/concerning}} |

## Red Flags

{{List any concerning terms that require attention}}

### Critical Issues
- {{critical_issue_1}}
- {{critical_issue_2}}

### Significant Concerns
- {{concern_1}}
- {{concern_2}}

### Minor Issues
- {{minor_1}}

## Missing Terms

{{Important provisions not present in the contract}}

- [ ] {{missing_provision_1}}
- [ ] {{missing_provision_2}}

## Negotiation Points

### Must-Have Changes
1. {{essential_change}}

### Should-Have Changes
1. {{important_change}}

### Nice-to-Have Changes
1. {{optional_improvement}}

## Comparison to Standard

| Aspect | This Contract | Market Standard | Delta |
|--------|---------------|-----------------|-------|
| {{aspect}} | {{this}} | {{standard}} | {{better/worse/neutral}} |

## Recommendations

### Before Signing
1. {{pre-signature_action}}

### Post-Signature
1. {{implementation_action}}

### Ongoing Monitoring
1. {{monitoring_requirement}}

## Disclaimer

This review provides general information about contract terms from a technical perspective. It is not legal advice. Contract interpretation requires consideration of specific circumstances, governing law, and business context. Consult a qualified attorney for legal guidance.
```

## Analysis Guidelines

1. **Focus on technical terms**: SLAs, security, data handling
2. **Identify imbalances**: One-sided provisions
3. **Check for missing terms**: What should be there but isn't
4. **Consider exit**: Can you leave without data loss?
5. **Review change provisions**: How can terms change?
