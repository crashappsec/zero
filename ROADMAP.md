<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Zero Roadmap

**Vision**: Position Zero as the leading **open-source software analysis toolkit** — providing deep insights into what software is made of, how it's built, and its security posture.

Zero is the free, open-source component of the Crash Override platform. It provides analyzers for understanding software while adding AI capabilities via specialist agents.

---

## Maturity Levels

| Component | Status | Description |
|-----------|--------|-------------|
| **Scanners** | Alpha | 9 super scanners with 45+ features, changing fast |
| **AI Agents** | Alpha | 13 specialist agents for deep analysis |
| **CLI** | Alpha | Core commands working, APIs may change |
| **Reports** | Experimental | HTML reports via Evidence.dev, expect breaking changes |

---

## Planned Features

### Source Code Scanners

#### Reachability Analysis
Determine if vulnerable dependencies are actually used:
- Vulnerable Code Path Detection - Trace calls to vulnerable functions
- Call Graph Analysis - Map how vulnerabilities could be exploited
- Risk Prioritization - Focus on reachable vulnerabilities first

#### Advanced Architecture Analysis
- Dependency Graph Visualization - Interactive dependency explorer
- Circular Dependency Detection - Find problematic dependency cycles
- Layer Violation Identification - Detect architecture rule violations
- Microservice Mapping - Service-to-service communication from code
- Database Schema Analysis - Migration risks, schema drift from ORM models

#### Semgrep Rule Management
Better organization and discoverability of Semgrep rules:
- Unified Taxonomy - Consistent naming across code-security, tech-id, secrets, IaC
- Individual Rule Files - One file per rule for easy browsing and review
- Rule Browser - CLI command to list, search, and inspect rules
- Rule Metadata - Severity, category, references, test cases per rule
- Design TBD - Goal is making rules easy to browse, review, and contribute

#### Future Enhancements
- API Versioning Audit - Detect deprecated or sunset API endpoints
- Training Data Analysis - PII detection in ML datasets
- Jupyter Notebook Security - Secrets in `.ipynb` files
- Semgrep IaC Enhancement - Custom organizational policies via RAG patterns
- Secrets-in-IaC Detection - Semgrep rules for hardcoded secrets in IaC files

---

### Cloud & Runtime Scanners

These require cloud credentials or access to running infrastructure. **Not source code based.**

#### Cloud Asset Inventory
Connect to cloud providers to build infrastructure SBOMs:
- Multi-Cloud Discovery - Inventory AWS, Azure, GCP resources via [CloudQuery](https://github.com/cloudquery/cloudquery) or [Fix Inventory](https://github.com/someengineering/fixinventory)
- Cloud SBOM Generation - CycloneDX SBOMs for containers, functions, services
- Runtime vs Build-time Comparison - Compare deployed assets against source SBOMs
- Cloud Security Posture - Misconfigurations, exposed services, risky permissions
- Cross-Cloud Unified View - Normalize resource data across providers

#### Live Endpoint Scanning
- Certificate Expiry Monitoring - Check live SSL/TLS certificates
- Exposed Service Detection - Identify publicly accessible services
- DNS Security - DNSSEC, SPF, DKIM, DMARC validation

---

### Reports & Analytics

#### Report Enhancements
- PDF Export - Executive summaries for stakeholders
- Trend Analysis - Track security posture over time
- Compliance Dashboards - SOC 2, ISO 27001, NIST mapping

#### Developer Experience Metrics
- Flow Metrics - Cycle time, PR review time, deployment frequency
- Bottleneck Identification - Where is work getting stuck?
- Team Health Indicators - Code review patterns, on-call burden
- Investment Tracking - Where is engineering time spent?

---

## Integration Roadmap

### Ocular Integration

[Ocular](https://ocularproject.io) is a Crash Override project providing robust code synchronization and tool orchestration at scale.

- Replace Zero's hydration with Ocular's code sync
- Leverage Ocular's repository caching and versioning
- Delegate scanner execution to Ocular's orchestration layer
- Support for monorepos and multi-repo projects
- Real-time analysis as Ocular syncs changes

### Chalk Integration

[Chalk](https://github.com/crashappsec/chalk) provides build-time attestation and security metadata.

- Build-time security analysis integration
- Attestation enrichment with Zero findings
- CI/CD workflow templates
- SLSA compliance verification

### GitHub/GitLab Organization Analysis

- Repository security configuration audit
- Branch protection and access review
- GitHub Actions/GitLab CI security analysis
- Compliance mapping (SOC 2, ISO 27001)
- Organization-wide policy enforcement

### Database Backend

- SQLite for single-user deployments
- DuckDB for analytics and dashboards
- PostgreSQL for enterprise multi-user
- Cross-project queries and aggregation

---

## How to Contribute

1. **Submit Feature Requests**: [Create an issue](https://github.com/crashappsec/zero/issues/new)
2. **Comment on Existing Items**: Add use cases and implementation ideas
3. **Vote with Reactions**: Use 👍 to help prioritize
4. **Contribute Code**: Pick up any roadmap item and submit a PR

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

---

*Last Updated: 2025-12-24*
*Version: 3.7.0*

*"Hack the planet!"*
