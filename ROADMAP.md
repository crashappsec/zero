<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Zero Roadmap

**Vision**: Position Zero as the leading **open-source software analysis toolkit** — providing deep insights into what software is made of, how it's built, and its security posture.

Zero is the free, open-source component of the Crash Override platform. It provides analyzers for understanding software while adding AI capabilities via specialist agents. Zero serves as an on-ramp to the commercial Crash Override platform for organizations needing enterprise features.

---

## What's Complete

### Core Infrastructure ✅
- **Zero CLI** - Master orchestrator with hydrate, scan, report commands
- **Agent System** - 10 Hackers-themed specialist agents with Task tool integration
- **Storage** - Project hydration with analysis caching in `~/.zero/`
- **Profiles** - quick, standard, security, advanced, deep analysis modes

### Scanners ✅
| Scanner | Status | Description |
|---------|--------|-------------|
| tech-discovery | ✅ | 112+ technologies with multi-layer detection |
| vulnerabilities | ✅ | CVE scanning via OSV.dev and CISA KEV |
| package-malcontent | ✅ | Supply chain compromise detection (14,500+ YARA rules) |
| package-health | ✅ | Abandonment, typosquatting, health scoring |
| licenses | ✅ | SPDX license analysis |
| code-security | ✅ | AI-powered security review |
| secrets-scanner | ✅ | Pattern-based secret detection (22+ patterns) |
| package-sbom | ✅ | CycloneDX SBOM via Syft |
| dora | ✅ | DORA metrics calculation |
| code-ownership | ✅ | Contributor analysis, bus factor |
| iac-security | ✅ | Checkov integration (50+ frameworks) |
| tech-debt | ✅ | RAG-based weighted scoring |
| documentation | ✅ | README quality, API docs coverage |
| git-insights | ✅ | Contributor patterns, churn analysis |
| test-coverage | ✅ | Framework detection, coverage estimation |
| auth-analysis | ✅ | Auth provider and pattern detection |

### Agents ✅
All 10 specialist agents with knowledge bases and Claude Code integration:
- **Cereal** - Supply chain security
- **Razor** - Code security
- **Blade** - Compliance auditing
- **Phreak** - Legal counsel
- **Acid** - Frontend engineering
- **Dade** - Backend engineering
- **Nikon** - Software architecture
- **Joey** - Build engineering
- **Plague** - DevOps engineering
- **Gibson** - Engineering metrics

---

## In Progress

### Agent Autonomy
- [ ] Full investigation mode with tool access (Read, Grep, Glob, WebSearch)
- [ ] Agent-to-agent delegation for complex investigations
- [ ] Improved context loading from cached analysis data

### Report System
- [ ] HTML report generation with interactive visualizations
- [ ] PDF export for executive summaries
- [ ] Trend analysis across multiple scans

---

## Planned Features

### Q1 2025

#### Enhanced Secret Detection
- [ ] Claude-enhanced false positive reduction
- [ ] Context-aware severity assessment
- [ ] Git history deep scanning
- [ ] Secret rotation recommendations

#### Bundle Analysis (npm/JavaScript)
- [ ] Bundle size analysis via bundlephobia API
- [ ] Tree-shaking opportunity detection
- [ ] Heavy package identification with lighter alternatives
- [ ] Code splitting recommendations

### Q2 2025

#### Developer Experience Metrics
- [ ] Developer satisfaction surveys (Swarmia-inspired)
- [ ] Flow metrics (cycle time, PR review time)
- [ ] Bottleneck identification
- [ ] Working agreements monitoring

#### Container Security
- [ ] Dockerfile best practices analysis
- [ ] Base image vulnerability assessment
- [ ] Hardened image recommendations (Chainguard, Distroless)
- [ ] Multi-stage build optimization

### Q3 2025

#### Business Alignment
- [ ] Investment tracking (where is engineering time spent?)
- [ ] Initiative monitoring across teams
- [ ] OKR alignment and tracking
- [ ] Quarterly planning with capacity forecasting

#### Advanced Architecture Analysis
- [ ] Dependency graph visualization
- [ ] Circular dependency detection
- [ ] Layer violation identification
- [ ] API security analysis

### Q4 2025

#### Predictive Intelligence
- [ ] AI impact measurement
- [ ] Delivery timeline forecasting
- [ ] Security posture trending
- [ ] Supply chain risk forecasting

---

## Integration Roadmap

### Chalk Integration
- Build-time security analysis
- Attestation enrichment with Zero findings
- CI/CD workflow templates

### GitHub Organization Analysis
- Repository security configuration audit
- Branch protection and access review
- GitHub Actions security analysis
- Compliance mapping (SOC 2, ISO 27001)

### Database Backend (Research)
- SQLite for single-user deployments
- DuckDB for analytics and dashboards
- PostgreSQL for enterprise multi-user
- Enable cross-project queries

---

## How to Contribute

1. **Submit Feature Requests**: [Create an issue](https://github.com/crashappsec/zero/issues/new)
2. **Comment on Existing Items**: Add use cases and implementation ideas
3. **Vote with Reactions**: Use 👍 to help prioritize
4. **Contribute Code**: Pick up any roadmap item and submit a PR

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

---

*Last Updated: 2025-12-06*
*Version: 5.0.0*

*"Hack the planet!"*
