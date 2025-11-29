# Compliance Framework Guide

## Framework Overview

| Framework | Focus | Certification | Typical Audience |
|-----------|-------|---------------|------------------|
| SOC 2 | Service organization controls | Type I / Type II | SaaS vendors |
| ISO 27001 | Information security management | Certification | Global enterprises |
| NIST CSF | Cybersecurity maturity | Self-assessment | US organizations |
| PCI-DSS | Cardholder data protection | SAQ / ROC | Payment processors |
| HIPAA | Health information | N/A (regulatory) | Healthcare |

## SOC 2

### Trust Service Criteria

**Security (Common Criteria)** - Required for all SOC 2 reports
- CC1: Control Environment
- CC2: Communication and Information
- CC3: Risk Assessment
- CC4: Monitoring Activities
- CC5: Control Activities
- CC6: Logical and Physical Access
- CC7: System Operations
- CC8: Change Management
- CC9: Risk Mitigation

**Availability** - System uptime commitments
- A1: Availability commitments and requirements

**Processing Integrity** - Complete, accurate processing
- PI1: Processing integrity commitments

**Confidentiality** - Information protection
- C1: Confidentiality commitments

**Privacy** - Personal information handling
- P1-P8: Privacy criteria

### Type I vs Type II

| Aspect | Type I | Type II |
|--------|--------|---------|
| Scope | Design of controls | Design and operating effectiveness |
| Period | Point in time | Period of time (typically 12 months) |
| Testing | Control description | Control testing with samples |
| Assurance | Lower | Higher |
| Use case | Initial compliance, M&A | Ongoing assurance |

### Common Control Mappings

| Control Objective | Typical Controls |
|-------------------|------------------|
| CC6.1 Logical access | SSO, MFA, RBAC |
| CC6.2 User registration | Provisioning workflows |
| CC6.3 Access removal | Deprovisioning procedures |
| CC7.1 Detection | SIEM, monitoring, alerting |
| CC7.2 Incident response | IR plan, runbooks |
| CC8.1 Change management | CAB, approvals, testing |

---

## ISO 27001

### Structure

- **Clauses 4-10**: ISMS requirements
- **Annex A**: 93 controls (2022 version)

### Key Clauses

| Clause | Focus |
|--------|-------|
| 4 | Context of organization |
| 5 | Leadership commitment |
| 6 | Risk assessment and treatment |
| 7 | Support (resources, competence) |
| 8 | Operational planning |
| 9 | Performance evaluation |
| 10 | Improvement |

### Annex A Control Themes (2022)

1. Organizational controls (37)
2. People controls (8)
3. Physical controls (14)
4. Technological controls (34)

### Certification Process

1. Gap assessment
2. ISMS implementation
3. Internal audit
4. Stage 1 audit (documentation review)
5. Stage 2 audit (implementation verification)
6. Certification
7. Surveillance audits (annual)
8. Recertification (3 years)

---

## NIST Cybersecurity Framework

### Functions

| Function | Purpose | Example Categories |
|----------|---------|-------------------|
| **Identify** | Know your assets and risks | Asset management, risk assessment |
| **Protect** | Safeguard delivery | Access control, training, data security |
| **Detect** | Identify incidents | Anomaly detection, monitoring |
| **Respond** | Take action | Response planning, communications |
| **Recover** | Restore capabilities | Recovery planning, improvements |

### Implementation Tiers

| Tier | Description | Characteristics |
|------|-------------|-----------------|
| 1 - Partial | Ad hoc, reactive | Limited awareness, informal |
| 2 - Risk Informed | Aware but informal | Some processes, not org-wide |
| 3 - Repeatable | Formal policies | Consistent, org-wide |
| 4 - Adaptive | Continuous improvement | Predictive, risk-informed decisions |

### Profiles

- **Current Profile**: Present cybersecurity posture
- **Target Profile**: Desired future state
- **Gap Analysis**: Roadmap from current to target

---

## PCI-DSS

### 12 Requirements

| Req | Domain | Description |
|-----|--------|-------------|
| 1 | Network security | Firewall configuration |
| 2 | Secure configuration | No vendor defaults |
| 3 | Data protection | Protect stored cardholder data |
| 4 | Encryption | Encrypt transmission |
| 5 | Malware | Anti-malware |
| 6 | Secure development | Secure systems and applications |
| 7 | Access control | Restrict access need-to-know |
| 8 | Authentication | Identify and authenticate |
| 9 | Physical security | Restrict physical access |
| 10 | Logging | Track and monitor access |
| 11 | Testing | Regular security testing |
| 12 | Policies | Information security policy |

### Validation Types

| Type | Who | Method |
|------|-----|--------|
| SAQ A | Card-not-present, fully outsourced | Self-assessment |
| SAQ A-EP | E-commerce, redirected | Self-assessment |
| SAQ D | All others | Self-assessment or ROC |
| ROC | Level 1 merchants | QSA assessment |

---

## HIPAA

### Safeguard Categories

**Administrative Safeguards**
- Risk analysis and management
- Workforce security
- Information access management
- Security awareness training
- Security incident procedures
- Contingency planning

**Physical Safeguards**
- Facility access controls
- Workstation security
- Device and media controls

**Technical Safeguards**
- Access control
- Audit controls
- Integrity controls
- Transmission security

### Key Requirements

| Requirement | Description |
|-------------|-------------|
| Risk Assessment | Annual risk analysis |
| BAA | Business Associate Agreements |
| Encryption | Data at rest and in transit |
| Access Controls | Minimum necessary access |
| Audit Logs | Activity tracking |
| Incident Response | Breach notification (60 days) |

---

## Framework Mapping

### Common Control Alignment

| Control Area | SOC 2 | ISO 27001 | NIST CSF | PCI-DSS |
|--------------|-------|-----------|----------|---------|
| Access Control | CC6.1-6.3 | A.9 | PR.AC | 7, 8 |
| Change Management | CC8.1 | A.12.1.2 | PR.IP-3 | 6.4 |
| Incident Response | CC7.2-7.5 | A.16 | RS.RP | 12.10 |
| Risk Assessment | CC3.1-3.4 | 6.1 | ID.RA | 12.2 |
| Logging | CC7.1 | A.12.4 | DE.CM | 10 |
| Encryption | CC6.7 | A.10 | PR.DS | 3, 4 |

### Using Multiple Frameworks

1. **Identify primary framework** based on business needs
2. **Map controls** to other frameworks
3. **Implement once**, document for multiple frameworks
4. **Unified evidence** collection
5. **Single audit** covering multiple frameworks where possible
