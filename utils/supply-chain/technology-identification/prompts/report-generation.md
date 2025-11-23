# Report Generation Prompt

**Purpose**: Generate comprehensive, executive-level technology stack reports with risk assessment and recommendations.

## Prompt

```
You are a technology architecture analyst creating an executive report on the technology stack of a software repository. Generate a comprehensive report with actionable insights.

## Input Data

### Technology Findings (JSON)
```json
{TECHNOLOGY_FINDINGS_JSON}
```

### Repository Context
**Repository**: {REPO_NAME}
**Organization**: {ORG_NAME}
**Primary Language**: {PRIMARY_LANGUAGE}
**Total Dependencies**: {DEP_COUNT}
**Scan Date**: {SCAN_DATE}

### Analysis Scope
- Package Manager Files Analyzed: {MANIFEST_COUNT}
- Configuration Files Scanned: {CONFIG_COUNT}
- Source Files Analyzed: {SOURCE_FILE_COUNT}
- Lines of Code: {LOC}

## Report Requirements

Generate a **markdown report** with the following sections:

## 1. Executive Summary

**Format**:
```markdown
# Technology Stack Report

**Repository**: {REPO_NAME}
**Organization**: {ORG_NAME}
**Analysis Date**: {SCAN_DATE}
**Total Technologies Identified**: {COUNT}

## Executive Summary

This repository utilizes **{COUNT} technologies** across **{CATEGORY_COUNT} categories**, with a focus on {PRIMARY_FOCUS}.

**Key Findings**:
- ✅ **Strengths**: {2-3 positive observations}
- ⚠️ **Concerns**: {2-3 risk areas}
- 🔴 **Critical Issues**: {Number} technologies requiring immediate attention

**Overall Technology Health**: {RATING}/10

**Primary Risk**: {HIGHEST_RISK_SUMMARY}
```

## 2. Technology Breakdown by Category

For each category with technologies, provide:

**Format**:
```markdown
## Technology Breakdown

### Business Tools & Services ({COUNT})

#### Payment Processing
- **Stripe v14.12.0** (Confidence: 94%)
  - **Usage**: Payment processing and subscription management
  - **Evidence**: package.json dependency, src/payments/ usage
  - **Status**: ✅ Current stable version (released 2024-11-01)
  - **Risk**: 🟢 Low - No known vulnerabilities
  - **Compliance**: PCI DSS relevant - ensure proper key management

#### Communication
- **Twilio v4.5.0** (Confidence: 88%)
  - **Usage**: SMS notifications and voice calls
  - **Evidence**: requirements.txt, src/notifications.py
  - **Status**: ✅ Current version
  - **Risk**: 🟢 Low
  - **Compliance**: TCPA compliance for SMS, GDPR for contact data

### Developer Tools ({COUNT})

#### Infrastructure as Code
- **Terraform v1.6.4** (Confidence: 95%)
  - **Usage**: AWS infrastructure provisioning
  - **Evidence**: terraform/ directory, .terraform.lock.hcl
  - **Status**: ✅ Recent stable (current: v1.6.5)
  - **Risk**: 🟢 Low
  - **Note**: Consider upgrade to 1.6.5 for bug fixes

#### Containers
- **Docker v24.0.7** (Confidence: 90%)
  - **Usage**: Application containerization
  - **Evidence**: Dockerfile, docker-compose.yml
  - **Status**: ✅ Current stable
  - **Risk**: 🟢 Low

...

### Cryptographic Libraries ({COUNT})

#### TLS/SSL
- **OpenSSL 1.1.1q** (Confidence: 85%)
  - **Usage**: System-level TLS/SSL
  - **Evidence**: /usr/lib/libssl.so.1.1
  - **Status**: ⚠️ **END-OF-LIFE** (EOL: 2023-09-11)
  - **Risk**: 🔴 **CRITICAL**
  - **Vulnerabilities**: Multiple CVEs post-EOL
  - **Action Required**: Immediate upgrade to OpenSSL 3.x
  - **Impact**: System-wide security risk
```

## 3. Risk Assessment

Provide detailed risk analysis:

**Format**:
```markdown
## Risk Assessment

### Critical Risks ({COUNT})

#### 1. OpenSSL 1.1.1 End-of-Life
- **Technology**: OpenSSL 1.1.1q
- **Risk Level**: 🔴 **CRITICAL**
- **Issue**: Reached end-of-life on 2023-09-11, no longer receives security updates
- **Impact**:
  - All TLS/SSL connections vulnerable to unpatched CVEs
  - System-wide security exposure
  - Compliance violations (PCI DSS, SOC 2)
- **Known Vulnerabilities**:
  - CVE-2023-XXXX (CVSS 9.8) - Remote code execution
  - CVE-2023-YYYY (CVSS 7.5) - Information disclosure
- **Evidence**: /usr/lib/libssl.so.1.1
- **Recommendation**: Upgrade to OpenSSL 3.0.x or 3.1.x immediately
- **Migration Complexity**: Medium - Requires system update and testing
- **Timeline**: **Immediate** (0-7 days)

### High Risks ({COUNT})

#### 1. Node.js 14.x Approaching EOL
- **Technology**: Node.js 14.21.3
- **Risk Level**: 🟠 **HIGH**
- **Issue**: Node.js 14.x EOL on 2023-04-30
- **Impact**:
  - No security patches
  - NPM package compatibility issues
  - Developer ecosystem abandonment
- **Evidence**: .nvmrc, package.json engines field
- **Recommendation**: Upgrade to Node.js 20.x LTS
- **Migration Complexity**: Medium - Test for breaking changes
- **Timeline**: **Short-term** (30 days)

#### 2. MongoDB 4.2 Deprecated
- **Technology**: MongoDB 4.2.24
- **Risk Level**: 🟠 **HIGH**
- **Issue**: MongoDB 4.2 end of support (April 2023)
- **Impact**:
  - No security updates
  - Performance improvements unavailable
  - Limited cloud provider support
- **Evidence**: Dockerfile, docker-compose.yml
- **Recommendation**: Upgrade to MongoDB 6.0+
- **Migration Complexity**: High - Schema and query compatibility testing required
- **Timeline**: **Medium-term** (90 days)

### Medium Risks ({COUNT})

#### 1. Express.js Outdated Version
- **Technology**: Express 4.16.4
- **Risk Level**: 🟡 **MEDIUM**
- **Issue**: 2+ years behind current stable (4.18.x)
- **Impact**:
  - Missing security patches
  - Missing performance improvements
  - Accumulated technical debt
- **Evidence**: package.json
- **Recommendation**: Upgrade to Express 4.18.x
- **Migration Complexity**: Low - Backward compatible
- **Timeline**: **Medium-term** (60 days)

### Low Risks ({COUNT})
- All current/LTS versions
- No known vulnerabilities
- Active maintenance
```

## 4. Compliance & Security Implications

**Format**:
```markdown
## Compliance & Security Implications

### Export Control (ITAR/EAR)
- **OpenSSL 3.x**: Strong cryptography (>= 256-bit) - Generally EAR99 (export friendly)
- **BoringSSL**: Used by Google, export-cleared
- **Note**: Most modern cryptography is export-approved under EAR99

### License Compliance
- **AGPL Libraries**: None detected ✅
- **GPL Libraries**: None in proprietary code ✅
- **Permissive Licenses**: MIT, Apache-2.0 (98% of dependencies)
- **Risk**: Low - No copyleft concerns

### Data Privacy (GDPR/CCPA)
- **Analytics Tools**:
  - Google Analytics 4 - EU data transfers, cookie consent required
  - Mixpanel - PII handling, data retention policies needed
- **CRM**:
  - Salesforce API - Customer PII storage, DPA required
- **Communication**:
  - Twilio - Message content privacy, TCPA compliance
  - SendGrid - Email tracking, GDPR unsubscribe mechanisms

### Financial Compliance (PCI DSS)
- **Payment Processors**:
  - Stripe API - PCI DSS Level 1 compliant service
  - **Critical**: Never log credit card data
  - **Critical**: Secure API key storage (use secrets management)
  - Recommended: Use Stripe Elements (reduces PCI scope)

### Security Best Practices
- **Secrets Management**:
  - 🔴 Found: 3 hardcoded API keys in codebase (CRITICAL)
  - ✅ Using: dotenv for local development
  - ⚠️ Missing: Vault/AWS Secrets Manager for production
- **Authentication**:
  - Using OAuth 2.0 via Auth0 ✅
  - JWT tokens with proper expiration ✅
- **Dependency Security**:
  - ⚠️ 5 dependencies with known CVEs (see vulnerability report)
```

## 5. Technology Rationalization Opportunities

Identify opportunities to simplify the stack:

**Format**:
```markdown
## Technology Rationalization

### Consolidation Opportunities

#### 1. Multiple HTTP Clients
**Finding**: 3 different HTTP clients detected:
- `axios` (v0.27.2) - Used in 45 files
- `node-fetch` (v2.6.7) - Used in 12 files
- `request` (v2.88.0) - Used in 3 files (DEPRECATED)

**Recommendation**: Standardize on `axios`
- **Benefit**: Reduce bundle size (~50KB), improve maintainability
- **Effort**: Low - Simple refactoring
- **Timeline**: 30 days

#### 2. Duplicate Functionality
**Finding**: Both `moment.js` and `date-fns` for date handling
- `moment` - Legacy, larger bundle size
- `date-fns` - Modern, tree-shakeable

**Recommendation**: Migrate to `date-fns`
- **Benefit**: Reduce bundle size (~67KB), better performance
- **Effort**: Medium - Test date formatting across app
- **Timeline**: 60 days

### Deprecated Technology Removal

#### jQuery 2.1.4
- **Status**: Deprecated (2016)
- **Usage**: Legacy admin panel only
- **Recommendation**: Migrate to Vanilla JS or Vue.js
- **Benefit**: Security, performance, maintainability
- **Effort**: High - UI rewrite required
- **Timeline**: 120 days

### Cloud Provider Optimization

**Finding**: Using AWS, GCP, and Azure simultaneously
- **AWS**: Primary (90% of infrastructure)
- **GCP**: Firebase only (10%)
- **Azure**: Single blob storage service

**Recommendation**: Consolidate to AWS
- **Benefit**: Simplified billing, better volume discounts, reduced complexity
- **Effort**: Medium - Migrate GCP Firebase to AWS Amplify, Azure Blob to S3
- **Timeline**: 90 days
- **Savings**: Est. 15-20% infrastructure cost reduction
```

## 6. Recommendations

Provide prioritized, actionable recommendations:

**Format**:
```markdown
## Recommendations

### Immediate (0-7 days) 🔴

1. **Upgrade OpenSSL to 3.x**
   - **Why**: Critical security risk, EOL since 2023-09-11
   - **Action**: Update system packages, rebuild containers
   - **Owner**: DevOps team
   - **Verification**: `openssl version` shows 3.x

2. **Rotate Hardcoded API Keys**
   - **Why**: Security breach risk
   - **Action**: Move to environment variables, rotate compromised keys
   - **Owner**: Security team
   - **Verification**: No API keys in source code

### Short-term (30 days) 🟠

1. **Upgrade Node.js to 20.x LTS**
   - **Why**: Node 14.x EOL, missing security patches
   - **Action**: Update .nvmrc, package.json engines, test application
   - **Owner**: Backend team
   - **Verification**: `node --version` shows v20.x

2. **Remove Deprecated `request` Package**
   - **Why**: No longer maintained, known vulnerabilities
   - **Action**: Replace with `axios` in 3 files
   - **Owner**: Backend team
   - **Verification**: `npm ls request` shows not installed

### Medium-term (60-90 days) 🟡

1. **Upgrade MongoDB to 6.x**
   - **Why**: Version 4.2 EOL, performance improvements in 6.x
   - **Action**: Test compatibility, plan migration, execute upgrade
   - **Owner**: Database team
   - **Verification**: Query performance tests pass

2. **Implement Secrets Management**
   - **Why**: Improve security posture, compliance requirement
   - **Action**: Deploy Vault or AWS Secrets Manager
   - **Owner**: DevOps + Security teams
   - **Verification**: All secrets loaded from secure vault

3. **Standardize HTTP Client Library**
   - **Why**: Reduce bundle size, improve maintainability
   - **Action**: Migrate all requests to `axios`
   - **Owner**: Frontend + Backend teams
   - **Verification**: Only `axios` in dependency tree

### Long-term (120+ days) 🟢

1. **Migrate Legacy jQuery Admin Panel**
   - **Why**: Technical debt, security concerns
   - **Action**: Rewrite in Vue.js or React
   - **Owner**: Frontend team
   - **Verification**: jQuery removed from package.json

2. **Cloud Provider Consolidation**
   - **Why**: Cost optimization, reduced complexity
   - **Action**: Migrate GCP/Azure services to AWS
   - **Owner**: Cloud Architecture team
   - **Verification**: Single cloud provider for 95%+ of services
```

## 7. Technology Maturity Assessment

Assess overall technology stack maturity:

**Format**:
```markdown
## Technology Maturity Assessment

### Overall Score: {SCORE}/10

#### Modern Stack (8/10)
✅ **Strengths**:
- Using current LTS versions for most core technologies
- Modern frontend framework (React 18.x)
- Container-based deployment (Docker + Kubernetes)
- Infrastructure as Code (Terraform)
- Automated CI/CD (GitHub Actions)

⚠️ **Areas for Improvement**:
- Some EOL technologies (OpenSSL 1.1, Node 14)
- Lack of centralized secrets management
- Multiple overlapping tools (HTTP clients, date libraries)

### Security Posture (6/10)
✅ **Strengths**:
- OAuth 2.0 authentication
- HTTPS everywhere
- Dependency scanning in CI/CD

🔴 **Critical Gaps**:
- EOL cryptographic libraries
- Hardcoded secrets in codebase
- Missing secrets management solution
- 5 dependencies with known CVEs

### Maintainability (7/10)
✅ **Strengths**:
- Well-documented dependencies
- Standard project structure
- Automated testing

⚠️ **Areas for Improvement**:
- Deprecated packages (jQuery, request)
- Technology sprawl (3 HTTP clients)
- Legacy code accumulation

### Compliance Readiness (7/10)
✅ **Strengths**:
- PCI DSS via Stripe
- GDPR-compliant auth (Auth0)
- Audit logging

⚠️ **Gaps**:
- Export control documentation needed
- Data retention policies undefined
- DPA not established with all vendors
```

## 8. Next Steps

**Format**:
```markdown
## Next Steps

### Week 1
- [ ] Upgrade OpenSSL to 3.x
- [ ] Rotate hardcoded API keys
- [ ] Document export control technologies

### Month 1
- [ ] Upgrade Node.js to 20.x LTS
- [ ] Remove deprecated `request` package
- [ ] Implement secrets management solution

### Quarter 1 (3 months)
- [ ] Upgrade MongoDB to 6.x
- [ ] Standardize on single HTTP client
- [ ] Address all high-priority CVEs
- [ ] Establish data retention policies

### Quarter 2-4
- [ ] Migrate legacy jQuery admin panel
- [ ] Cloud provider consolidation
- [ ] Technology stack documentation
- [ ] Quarterly technology review process
```

---

## Report Metadata

```markdown
---

**Report Generated**: {ISO_DATE}
**Analysis Version**: {VERSION}
**Analyzer**: Technology Identification System v2.0
**Total Technologies Analyzed**: {COUNT}
**Confidence Threshold**: 60%
**Risk Framework**: NIST Cybersecurity Framework

For questions or clarifications, contact: {CONTACT_EMAIL}
```

## Quality Standards

- **Executive Focus**: Clear, non-technical language in executive summary
- **Evidence-Based**: Every claim backed by specific evidence
- **Actionable**: Concrete recommendations with timelines and owners
- **Prioritized**: Risk-based prioritization (Critical → High → Medium → Low)
- **Comprehensive**: Cover security, compliance, cost, and maintainability
- **Realistic**: Timeline estimates based on complexity and dependencies

Generate professional, actionable, executive-ready technology stack reports.
```

## Usage

```bash
# Generate comprehensive report
./technology-identification-analyser.sh \
  --repo owner/repo \
  --format markdown \
  --output tech-stack-report.md \
  --executive-summary

# Generate with Claude AI for enhanced insights
export ANTHROPIC_API_KEY="sk-..."
./technology-identification-analyser.sh \
  --claude \
  --repo owner/repo \
  --format markdown \
  --output tech-stack-report.md
```
