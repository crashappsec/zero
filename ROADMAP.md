<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Zero Roadmap

**Vision**: Position Zero as the leading **open-source software analysis toolkit** — providing deep insights into what software is made of, how it's built, and its security posture.

Zero is the free, open-source component of the Crash Override platform. It provides analyzers for understanding software while adding AI capabilities via specialist agents.

---

## Current Capabilities

Zero currently includes **21 scanners** across these categories:

| Category | Scanners |
|----------|----------|
| **Supply Chain Security** | package-sbom, package-vulns, package-health, package-provenance, package-malcontent, package-recommendations, package-bundle-optimization |
| **Code Security** | code-security, code-secrets, iac-security, container-security |
| **Compliance & Legal** | licenses, digital-certificates |
| **Developer Productivity** | tech-discovery, code-ownership, dora, git, documentation, test-coverage, tech-debt |
| **Container & Infrastructure** | containers, bundle-analysis |

Plus **10 AI specialist agents** for deep analysis (Cereal, Razor, Blade, Phreak, Acid, Dade, Nikon, Joey, Plague, Gibson).

---

## Planned Features

### Source Code Scanners

These scanners analyze repositories and work with Zero's existing hydrate workflow.

#### API Security Analysis
Scan OpenAPI specs, GraphQL schemas, and route definitions in source code:
- [ ] **OpenAPI/Swagger Scanning** - Parse API specs for security issues (auth, rate limiting, input validation)
- [ ] **GraphQL Security** - Introspection exposure, query complexity, authorization gaps
- [ ] **Authentication Analysis** - OAuth/OIDC configuration review in code
- [ ] **API Versioning Audit** - Detect deprecated or sunset API endpoints

#### AI/ML Model Security
Scan ML models, datasets, and AI pipelines in repositories:
- [ ] **ML-BOM Generation** - CycloneDX inventory of models, datasets, frameworks
- [ ] **Pickle/Model File Scanning** - Detect unsafe deserialization in `.pkl`, `.pt` files
- [ ] **ML Framework Vulnerabilities** - CVEs in PyTorch, TensorFlow, scikit-learn
- [ ] **Training Data Analysis** - PII detection, provenance tracking
- [ ] **Jupyter Notebook Security** - Secrets, credentials, unsafe code in `.ipynb`

#### Cryptography Audit
Scan source code for weak cryptographic patterns:
- [ ] **Weak Cipher Detection** - Flag MD5, SHA1, DES, RC4, ECB mode usage
- [ ] **Hardcoded Keys/IVs** - Detect cryptographic keys in source code
- [ ] **TLS Configuration** - Insecure TLS versions, weak cipher suites in configs
- [ ] **Certificate Validation** - Disabled cert verification, pinning issues

#### Enhanced Secret Detection
- [ ] **Claude-enhanced False Positive Reduction** - AI-powered context analysis
- [ ] **Git History Deep Scanning** - Find secrets in commit history
- [ ] **Secret Rotation Recommendations** - Suggest remediation steps
- [ ] **Entropy Analysis** - Detect high-entropy strings that may be secrets

#### Reachability Analysis
Determine if vulnerable dependencies are actually used:
- [ ] **Vulnerable Code Path Detection** - Trace calls to vulnerable functions
- [ ] **Call Graph Analysis** - Map how vulnerabilities could be exploited
- [ ] **Risk Prioritization** - Focus on reachable vulnerabilities first
- [ ] **VEX Generation** - Auto-generate Vulnerability Exploitability eXchange documents

#### Advanced Architecture Analysis
- [ ] **Dependency Graph Visualization** - Interactive dependency explorer
- [ ] **Circular Dependency Detection** - Find problematic dependency cycles
- [ ] **Layer Violation Identification** - Detect architecture rule violations
- [ ] **Microservice Mapping** - Service-to-service communication from code
- [ ] **Database Schema Analysis** - Migration risks, schema drift from ORM models

---

### Cloud & Runtime Scanners

These require cloud credentials or access to running infrastructure. **Not source code based.**

#### Cloud Asset Inventory
Connect to cloud providers to build infrastructure SBOMs:
- [ ] **Multi-Cloud Discovery** - Inventory AWS, Azure, GCP resources via [CloudQuery](https://github.com/cloudquery/cloudquery) or [Fix Inventory](https://github.com/someengineering/fixinventory)
- [ ] **Cloud SBOM Generation** - CycloneDX SBOMs for containers, functions, services
- [ ] **Runtime vs Build-time Comparison** - Compare deployed assets against source SBOMs
- [ ] **Cloud Security Posture** - Misconfigurations, exposed services, risky permissions
- [ ] **Cross-Cloud Unified View** - Normalize resource data across providers

#### Live Endpoint Scanning
- [ ] **Certificate Expiry Monitoring** - Check live SSL/TLS certificates
- [ ] **Exposed Service Detection** - Identify publicly accessible services
- [ ] **DNS Security** - DNSSEC, SPF, DKIM, DMARC validation

---

### Reports & Analytics

#### Report System
- [ ] **HTML Report Generation** - Interactive visualizations with drill-down
- [ ] **PDF Export** - Executive summaries for stakeholders
- [ ] **Trend Analysis** - Track security posture over time
- [ ] **Delta Detection** - New/fixed findings between scans
- [ ] **Compliance Dashboards** - SOC 2, ISO 27001, NIST mapping

#### Developer Experience Metrics
- [ ] **Flow Metrics** - Cycle time, PR review time, deployment frequency
- [ ] **Bottleneck Identification** - Where is work getting stuck?
- [ ] **Team Health Indicators** - Code review patterns, on-call burden
- [ ] **Investment Tracking** - Where is engineering time spent?

---

## Integration Roadmap

### Ocular Integration

[Ocular](https://ocularproject.io) is a Crash Override project providing robust code synchronization and tool orchestration at scale.

- [ ] Replace Zero's hydration with Ocular's code sync
- [ ] Leverage Ocular's repository caching and versioning
- [ ] Delegate scanner execution to Ocular's orchestration layer
- [ ] Support for monorepos and multi-repo projects
- [ ] Real-time analysis as Ocular syncs changes

### Chalk Integration

[Chalk](https://github.com/crashappsec/chalk) provides build-time attestation and security metadata.

- [ ] Build-time security analysis integration
- [ ] Attestation enrichment with Zero findings
- [ ] CI/CD workflow templates
- [ ] SLSA compliance verification

### GitHub/GitLab Organization Analysis

- [ ] Repository security configuration audit
- [ ] Branch protection and access review
- [ ] GitHub Actions/GitLab CI security analysis
- [ ] Compliance mapping (SOC 2, ISO 27001)
- [ ] Organization-wide policy enforcement

### Database Backend

- [ ] SQLite for single-user deployments
- [ ] DuckDB for analytics and dashboards
- [ ] PostgreSQL for enterprise multi-user
- [ ] Cross-project queries and aggregation

---

## How to Contribute

1. **Submit Feature Requests**: [Create an issue](https://github.com/crashappsec/zero/issues/new)
2. **Comment on Existing Items**: Add use cases and implementation ideas
3. **Vote with Reactions**: Use 👍 to help prioritize
4. **Contribute Code**: Pick up any roadmap item and submit a PR

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

---

*Last Updated: 2025-12-12*
*Version: 5.1.0*

*"Hack the planet!"*
