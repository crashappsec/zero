<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
SPDX-License-Identifier: GPL-3.0
-->

# Compliance Frameworks for Supply Chain Security Audits

## SOC 2 Type II

### Overview

```
Purpose: Third-party assurance on controls for service organizations
Governing Body: AICPA
Report Users: Customers, prospects, regulators
Period: Typically 6-12 months of operating effectiveness
```

### Trust Service Criteria for Supply Chain

```
CC6 - Logical and Physical Access Controls
─────────────────────────────────────────────────────────────────
CC6.1 - The entity implements logical access security software,
        infrastructure, and architectures over protected information
        assets to protect them from security events

Supply Chain Controls:
• Access controls for package registries
• Authentication for artifact repositories
• Code signing key management
• CI/CD pipeline access restrictions

Testing Approach:
□ Review access provisioning process
□ Test access review effectiveness
□ Verify separation of duties
□ Examine key management procedures


CC6.6 - The entity implements logical access security measures to
        protect against threats from sources outside its system
        boundaries

Supply Chain Controls:
• Dependency source verification
• Package integrity checking
• Supply chain attack detection
• Third-party code review

Testing Approach:
□ Review dependency sourcing policy
□ Test integrity verification (checksums, signatures)
□ Examine supply chain monitoring tools
□ Verify third-party assessment process


CC7 - System Operations
─────────────────────────────────────────────────────────────────
CC7.1 - To meet its objectives, the entity uses detection and
        monitoring procedures to identify changes to configurations
        that result in the introduction of new vulnerabilities

Supply Chain Controls:
• Automated vulnerability scanning
• Dependency change monitoring
• Configuration drift detection
• Security advisory monitoring

Testing Approach:
□ Verify scanning coverage and frequency
□ Test detection of known vulnerabilities
□ Review alerting and notification
□ Examine response procedures


CC8 - Change Management
─────────────────────────────────────────────────────────────────
CC8.1 - The entity authorizes, designs, develops or acquires,
        configures, documents, tests, approves, and implements
        changes to infrastructure, data, software, and procedures

Supply Chain Controls:
• Dependency update authorization
• Version control procedures
• Testing requirements
• Approval workflows

Testing Approach:
□ Review change management policy
□ Test approval workflows
□ Verify testing requirements enforced
□ Examine emergency change procedures
```

### SOC 2 Testing Matrix

```
┌───────────────────────────────────────────────────────────────────────────┐
│                     SOC 2 SUPPLY CHAIN TESTING MATRIX                     │
├─────────────┬─────────────────────────────┬───────────────────────────────┤
│ TSC         │ Control                     │ Test Procedure                │
├─────────────┼─────────────────────────────┼───────────────────────────────┤
│ CC6.1       │ Registry access controls    │ Review access list, test      │
│             │                             │ provisioning sample           │
├─────────────┼─────────────────────────────┼───────────────────────────────┤
│ CC6.6       │ Dependency verification     │ Verify signature checking,    │
│             │                             │ test supply chain controls    │
├─────────────┼─────────────────────────────┼───────────────────────────────┤
│ CC7.1       │ Vulnerability scanning      │ Review coverage, test         │
│             │                             │ detection and alerting        │
├─────────────┼─────────────────────────────┼───────────────────────────────┤
│ CC7.2       │ Vulnerability remediation   │ Test sample of remediations,  │
│             │                             │ verify SLA compliance         │
├─────────────┼─────────────────────────────┼───────────────────────────────┤
│ CC8.1       │ Dependency change control   │ Test sample of updates,       │
│             │                             │ verify approval workflow      │
├─────────────┼─────────────────────────────┼───────────────────────────────┤
│ CC9.2       │ Vendor risk management      │ Review vendor assessment,     │
│             │                             │ test monitoring procedures    │
└─────────────┴─────────────────────────────┴───────────────────────────────┘
```

## PCI DSS 4.0

### Overview

```
Purpose: Protect cardholder data
Governing Body: PCI Security Standards Council
Applicability: Any entity storing, processing, or transmitting CHD
Current Version: 4.0 (March 2022)
Future Requirements: Some 4.0 items mandatory March 2025
```

### Supply Chain Relevant Requirements

```
Requirement 6: Develop and Maintain Secure Systems and Software
─────────────────────────────────────────────────────────────────

6.3.1 - Security vulnerabilities are identified and managed
        using industry-recognized sources

Applicability: All system components
Evidence Required:
• List of vulnerability information sources
• Process for monitoring sources
• Evidence of CVE/advisory monitoring

Testing Procedure:
1. Identify sources used (NVD, vendor advisories, etc.)
2. Verify monitoring process documented
3. Test that new vulnerabilities are identified
4. Review sample of identified vulnerabilities


6.3.2 - An inventory of custom and third-party software components
        is maintained

Applicability: All applications in CDE
Evidence Required:
• Software inventory (SBOM)
• Third-party component list
• Version information

Testing Procedure:
1. Obtain software inventory
2. Verify completeness against actual systems
3. Check third-party components documented
4. Verify update process for inventory


6.3.3 - Software is developed and maintained considering industry
        standards for secure development

Applicability: Bespoke and custom software
Evidence Required:
• Secure SDLC documentation
• Dependency management procedures
• Security testing requirements

Testing Procedure:
1. Review SDLC documentation
2. Verify security requirements for dependencies
3. Test security review process
4. Examine dependency update workflow


6.4.3 - All payment page scripts that are loaded and executed in
        the consumer's browser are managed (NEW in 4.0)

Applicability: Payment page scripts
Evidence Required:
• Script inventory
• Authorization documentation
• Integrity verification (SRI)

Testing Procedure:
1. Obtain payment page script inventory
2. Verify all scripts authorized
3. Test integrity mechanisms
4. Review change detection/alerting


6.5.1 - Changes to system components are made according to
        change management procedures

Applicability: All system components in scope
Evidence Required:
• Change management policy
• Approval records
• Testing evidence

Testing Procedure:
1. Select sample of dependency updates
2. Verify proper authorization
3. Confirm testing completed
4. Verify deployment approval
```

### PCI DSS 4.0 Evidence Requirements

```
┌───────────────────────────────────────────────────────────────────────────┐
│                   PCI DSS 4.0 SUPPLY CHAIN EVIDENCE                       │
├────────────┬──────────────────────────────────────────────────────────────┤
│ Requirement│ Evidence                                                     │
├────────────┼──────────────────────────────────────────────────────────────┤
│ 6.3.1      │ • Vulnerability source documentation                        │
│            │ • Monitoring procedure                                       │
│            │ • Sample vulnerability identifications                       │
├────────────┼──────────────────────────────────────────────────────────────┤
│ 6.3.2      │ • Software/component inventory (SBOM)                       │
│            │ • Inventory update process                                   │
│            │ • Sample inventory vs. actual comparison                    │
├────────────┼──────────────────────────────────────────────────────────────┤
│ 6.3.3      │ • Secure SDLC documentation                                 │
│            │ • Dependency security requirements                          │
│            │ • Security testing results                                   │
├────────────┼──────────────────────────────────────────────────────────────┤
│ 6.4.3      │ • Payment page script inventory                             │
│ (Mar 2025) │ • Script authorization records                              │
│            │ • SRI implementation evidence                               │
│            │ • Script change monitoring                                   │
├────────────┼──────────────────────────────────────────────────────────────┤
│ 6.5.1      │ • Change management policy                                  │
│            │ • Sample dependency update records                          │
│            │ • Approval evidence                                          │
└────────────┴──────────────────────────────────────────────────────────────┘
```

## NIST SP 800-53 Rev 5

### Overview

```
Purpose: Security and privacy controls for federal information systems
Governing Body: NIST
Applicability: Federal agencies, contractors, FedRAMP
Control Families: 20 families with 1000+ controls
```

### Supply Chain Controls (SA, SR Families)

```
SA-8 Security and Privacy Engineering Principles
─────────────────────────────────────────────────────────────────
Control: Apply security engineering principles in specification,
         design, development, implementation, and modification

Supply Chain Elements:
(a) Verify third-party component integrity
(b) Minimize attack surface from dependencies
(c) Implement least privilege for build systems

Assessment Objectives:
• Determine if principles address supply chain
• Verify principles applied to dependency selection
• Confirm build system security


SA-9 External System Services
─────────────────────────────────────────────────────────────────
Control: Define and document government and external provider
         oversight and user roles

Supply Chain Elements:
(a) Establish trust requirements for external services
(b) Define required security controls for providers
(c) Monitor external service security posture

Assessment Objectives:
• Review external service agreements
• Verify security requirements defined
• Test monitoring of external services


SR-1 Policy and Procedures
─────────────────────────────────────────────────────────────────
Control: Develop, document, disseminate supply chain risk
         management policy and procedures

Assessment Objectives:
• Policy addresses supply chain risks
• Procedures are documented and current
• Policy is disseminated to relevant personnel


SR-2 Supply Chain Risk Management Plan
─────────────────────────────────────────────────────────────────
Control: Develop a plan for managing supply chain risks

Required Elements:
(a) Identify risks and threat scenarios
(b) Protection strategies and countermeasures
(c) Due diligence methods for suppliers
(d) Incident response procedures

Assessment Objectives:
• Plan exists and is current
• All required elements addressed
• Plan implemented as documented


SR-3 Supply Chain Controls and Processes
─────────────────────────────────────────────────────────────────
Control: Establish processes to identify and address supply chain
         risks

Control Enhancements:
(1) Diverse supply base
(2) Limitation of harm (compartmentalization)
(3) Notification agreements with suppliers

Assessment Objectives:
• Processes documented
• Controls implemented effectively
• Supplier relationships managed


SR-4 Provenance
─────────────────────────────────────────────────────────────────
Control: Document, monitor, and maintain valid provenance of
         system components

Assessment Objectives:
• Provenance tracked for critical components
• Documentation complete and accurate
• Monitoring procedures effective


SR-5 Acquisition Strategies, Tools, and Methods
─────────────────────────────────────────────────────────────────
Control: Employ tailored acquisition strategies and tools to
         reduce supply chain risk

Assessment Objectives:
• Acquisition strategy addresses risk
• Tools employed effectively
• Methods align with risk tolerance


SR-11 Component Authenticity
─────────────────────────────────────────────────────────────────
Control: Develop and implement anti-counterfeit policy and
         authenticity verification

Assessment Objectives:
• Authenticity verification implemented
• Counterfeit detection procedures exist
• Verification documented
```

## ISO 27001:2022

### Overview

```
Purpose: Information security management system requirements
Governing Body: ISO/IEC
Applicability: Any organization seeking certification
Current Version: 2022 (October 2022)
Key Change: New controls for cloud and supply chain
```

### Annex A Supply Chain Controls

```
A.5.19 Information Security in Supplier Relationships
─────────────────────────────────────────────────────────────────
Control: Processes and procedures for managing information security
         risks associated with suppliers shall be defined and
         implemented

Implementation Guidance:
• Risk assessment for supplier relationships
• Security requirements in agreements
• Monitoring of supplier compliance

Audit Evidence:
□ Supplier security policy
□ Risk assessment documentation
□ Security clauses in contracts
□ Monitoring procedures


A.5.20 Addressing Information Security Within Supplier Agreements
─────────────────────────────────────────────────────────────────
Control: Relevant information security requirements shall be
         established and agreed with each supplier

Implementation Guidance:
• Identify security requirements
• Include in formal agreements
• Cover incident response
• Address access and data protection

Audit Evidence:
□ Standard security clauses
□ Sample supplier agreements
□ Negotiation records


A.5.21 Managing Information Security in ICT Supply Chain
─────────────────────────────────────────────────────────────────
Control: Processes and procedures shall be defined and implemented
         for managing information security risks associated with
         ICT products and services supply chain

Implementation Guidance:
• Require suppliers to propagate security to their suppliers
• Address software development security
• Verify component integrity

Audit Evidence:
□ ICT supply chain policy
□ Software security requirements
□ Integrity verification procedures


A.5.22 Monitoring, Review and Change Management of Supplier Services
─────────────────────────────────────────────────────────────────
Control: The organization shall regularly monitor, review, audit
         and evaluate supplier service delivery changes

Implementation Guidance:
• Review service level performance
• Audit supplier security compliance
• Manage changes to supplier services

Audit Evidence:
□ Monitoring procedures
□ Review records
□ Change management process


A.5.23 Information Security for Use of Cloud Services
─────────────────────────────────────────────────────────────────
Control: Acquisition, use, management and exit from cloud services
         shall be established

Implementation Guidance:
• Define cloud security requirements
• Assess cloud provider security
• Manage shared responsibility

Audit Evidence:
□ Cloud security policy
□ Provider assessments
□ Shared responsibility documentation
```

## FedRAMP

### Overview

```
Purpose: Standardized security for cloud services to federal government
Governing Body: GSA/FedRAMP PMO
Applicability: Cloud Service Providers (CSPs) selling to federal agencies
Levels: Low, Moderate, High (based on FIPS 199)
```

### Supply Chain Specific Requirements

```
SBOM Requirements (2024+)
─────────────────────────────────────────────────────────────────
Requirement: CSPs must generate and maintain SBOMs for all
             software offered to federal customers

Format: SPDX or CycloneDX (machine-readable)

Minimum Elements (NTIA):
□ Supplier name
□ Component name
□ Component version
□ Unique identifier (PURL/CPE)
□ Dependency relationship
□ Author of SBOM data
□ Timestamp

Frequency: Generated with each release

Assessment:
1. Verify SBOM generation process
2. Validate format compliance
3. Check minimum elements present
4. Test accuracy against actual components


Continuous Monitoring
─────────────────────────────────────────────────────────────────
Requirement: Monthly vulnerability scanning and reporting

Supply Chain Elements:
• Dependency vulnerability scanning
• Third-party component monitoring
• Patch/update status reporting

Evidence:
□ Monthly scan reports
□ POA&M for open vulnerabilities
□ Remediation timelines


Supply Chain Risk Management
─────────────────────────────────────────────────────────────────
Requirement: Implement NIST SP 800-53 SR family controls

Key Controls:
• SR-1: Policy and procedures
• SR-2: Supply chain risk management plan
• SR-3: Supply chain controls and processes
• SR-4: Provenance
• SR-5: Acquisition strategies

Assessment: Per NIST SP 800-53A procedures
```

## Framework Comparison Matrix

```
┌─────────────────────────────────────────────────────────────────────────────┐
│              SUPPLY CHAIN CONTROL CROSS-REFERENCE                           │
├──────────────────────┬──────────┬──────────┬──────────┬──────────┬─────────┤
│ Control Area         │ SOC 2    │ PCI DSS  │ NIST     │ ISO27001 │ FedRAMP │
├──────────────────────┼──────────┼──────────┼──────────┼──────────┼─────────┤
│ Vuln Management      │ CC7.1    │ 6.3.1    │ RA-5     │ A.8.8    │ RA-5    │
│ Patch Management     │ CC7.2    │ 6.3.3    │ SI-2     │ A.8.8    │ SI-2    │
│ Software Inventory   │ CC6.1    │ 6.3.2    │ CM-8     │ A.5.9    │ CM-8    │
│ Third-Party Risk     │ CC9.2    │ 12.8     │ SA-9     │ A.5.19   │ SA-9    │
│ Change Control       │ CC8.1    │ 6.5.1    │ CM-3     │ A.8.32   │ CM-3    │
│ SBOM/Provenance      │ -        │ -        │ SR-4     │ A.5.21   │ SR-4    │
│ Supply Chain Plan    │ -        │ -        │ SR-2     │ A.5.21   │ SR-2    │
│ Supplier Agreements  │ CC9.2    │ 12.8.2   │ SA-9     │ A.5.20   │ SA-9    │
└──────────────────────┴──────────┴──────────┴──────────┴──────────┴─────────┘
```

## Quick Reference

### Auditor Framework Selection Guide

```
Organization Type          Primary Framework    Secondary
─────────────────────────────────────────────────────────────────
SaaS/Cloud Provider        SOC 2               ISO 27001
Payment Processing         PCI DSS             SOC 2
Federal Contractor         FedRAMP             NIST SP 800-53
Healthcare                 HIPAA               SOC 2
Financial Services         SOC 2               PCI DSS
Global Enterprise          ISO 27001           SOC 2
Government (Direct)        NIST SP 800-53      FedRAMP
```

### Evidence Efficiency Tips

```
One Evidence, Many Uses:
• Vulnerability scan reports → SOC 2, PCI, NIST, ISO, FedRAMP
• SBOM documentation → NIST, ISO, FedRAMP
• Change management records → All frameworks
• Vendor assessments → SOC 2, PCI, NIST, ISO

Plan audits to collect evidence once, use across frameworks
```
