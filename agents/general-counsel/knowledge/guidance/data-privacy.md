# Data Privacy Guide

## Overview

Data privacy regulations govern how organizations collect, process, store, and share personal information. Key regulations include:

| Regulation | Jurisdiction | Scope |
|------------|--------------|-------|
| GDPR | EU/EEA | Any org processing EU residents' data |
| CCPA/CPRA | California | Businesses meeting revenue/data thresholds |
| LGPD | Brazil | Similar to GDPR, Brazil residents |
| PIPEDA | Canada | Commercial activity in Canada |
| State Laws | US States | Virginia, Colorado, Connecticut, etc. |

## GDPR (General Data Protection Regulation)

### Key Principles

1. **Lawfulness, fairness, transparency** - Legal basis, clear communication
2. **Purpose limitation** - Collect for specified purposes only
3. **Data minimization** - Only collect what's necessary
4. **Accuracy** - Keep data accurate and up to date
5. **Storage limitation** - Don't keep longer than needed
6. **Integrity and confidentiality** - Appropriate security
7. **Accountability** - Demonstrate compliance

### Lawful Bases for Processing

| Basis | Description | Common Use |
|-------|-------------|------------|
| Consent | Freely given, specific, informed | Marketing emails |
| Contract | Necessary for contract performance | Service delivery |
| Legal obligation | Required by law | Tax records |
| Vital interests | Protect life | Emergency contact |
| Public task | Public authority functions | Government services |
| Legitimate interests | Balanced against data subject rights | Fraud prevention |

### Data Subject Rights

| Right | Description | Response Time |
|-------|-------------|---------------|
| Access | Copy of their data | 1 month |
| Rectification | Correct inaccurate data | 1 month |
| Erasure | Delete data ("right to be forgotten") | 1 month |
| Restriction | Limit processing | 1 month |
| Portability | Receive data in machine-readable format | 1 month |
| Object | Object to processing | Without delay |

### Breach Notification

- **Authority notification**: Within 72 hours of awareness
- **Data subject notification**: Without undue delay (if high risk)
- **Documentation**: Record all breaches regardless of notification

### Data Processing Agreements

When using processors (vendors), contracts must include:
- Subject matter and duration
- Nature and purpose of processing
- Types of personal data
- Categories of data subjects
- Controller's obligations and rights
- Security measures
- Sub-processor requirements
- Audit rights

---

## CCPA/CPRA (California)

### Applicability

Applies to businesses that:
- Have gross revenue > $25 million, OR
- Buy/sell data of 100,000+ consumers, OR
- Derive 50%+ revenue from selling personal information

### Consumer Rights

| Right | Description |
|-------|-------------|
| Know | What personal information is collected |
| Delete | Request deletion of personal information |
| Opt-out | Opt out of sale/sharing of personal information |
| Non-discrimination | Equal service regardless of privacy choices |
| Correct | Correct inaccurate personal information (CPRA) |
| Limit | Limit use of sensitive personal information (CPRA) |

### "Sale" of Personal Information

Broadly defined to include:
- Sharing data with third parties for monetary consideration
- Sharing data for other valuable consideration
- Sharing for cross-context behavioral advertising (CPRA)

**Requires:** "Do Not Sell My Personal Information" link

### Service Provider Requirements

- Written contract required
- Cannot retain/use/disclose data except as specified
- Must certify understanding of restrictions

---

## Privacy by Design

### Principles

1. **Proactive not reactive** - Prevent privacy issues
2. **Privacy as default** - No action required from user
3. **Privacy embedded** - Built into design
4. **Full functionality** - Not zero-sum (privacy vs. functionality)
5. **End-to-end security** - Lifecycle protection
6. **Visibility and transparency** - Verifiable practices
7. **User-centric** - Respect user interests

### Implementation

**Data Collection:**
- Collect minimum necessary data
- Clear purpose at collection
- Consent where required

**Data Storage:**
- Encryption at rest
- Access controls
- Retention policies
- Secure deletion

**Data Processing:**
- Purpose limitation
- Access logging
- Anonymization/pseudonymization where possible

**Data Sharing:**
- DPAs with processors
- Legitimate basis for transfers
- Appropriate safeguards for international transfers

---

## International Data Transfers

### EU to Non-EU Transfers

**Mechanisms:**
1. **Adequacy decision** - Country deemed adequate (UK, Canada, etc.)
2. **Standard Contractual Clauses (SCCs)** - Approved contract terms
3. **Binding Corporate Rules** - Intra-group transfers
4. **Derogations** - Explicit consent, contract necessity, etc.

**Post-Schrems II Requirements:**
- Transfer Impact Assessment
- Supplementary measures if needed
- Document decision-making

### US Data Transfers

- **EU-US Data Privacy Framework** - Participating companies
- **SCCs** - With supplementary measures
- **Case-by-case assessment** - Document legal basis

---

## Compliance Checklist

### Documentation
- [ ] Privacy policy (external)
- [ ] Data processing records (Article 30)
- [ ] Lawful basis documentation
- [ ] Consent records
- [ ] DPAs with vendors
- [ ] DPIA (Data Protection Impact Assessment) for high-risk processing

### Technical Measures
- [ ] Encryption (transit and rest)
- [ ] Access controls
- [ ] Audit logging
- [ ] Data minimization
- [ ] Retention/deletion automation
- [ ] Breach detection

### Operational
- [ ] Privacy training for staff
- [ ] Data subject request process
- [ ] Vendor assessment process
- [ ] Breach response plan
- [ ] Regular privacy reviews

### Governance
- [ ] DPO appointed (if required)
- [ ] Privacy impact assessments
- [ ] Privacy in product development
- [ ] Regular compliance audits

---

## Common Issues

### Website Compliance
- Cookie consent (not just notice)
- Privacy policy accessibility
- Do Not Sell link (if applicable)
- Consent before data collection

### Marketing
- Consent for marketing emails
- Easy unsubscribe
- Honor opt-outs promptly
- Separate consent per purpose

### Vendor Management
- Due diligence before engagement
- DPA in place
- Regular reviews
- Sub-processor tracking

### Data Retention
- Define retention periods
- Automate deletion
- Document justification
- Regular review

---

## Disclaimer

This guide provides general information about data privacy regulations. It is not legal advice. Privacy law is complex and varies by jurisdiction. The requirements depend on your specific circumstances. Consult a qualified attorney for legal advice regarding your privacy compliance obligations.
