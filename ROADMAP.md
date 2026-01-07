<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Zero Roadmap

**Version:** 5.0.0
**Last Updated:** 2026-01-07

**Vision**: Position Zero as the leading **open-source software analysis toolkit** — providing deep insights into what software is made of, how it's built, and its security posture.

---

## Current State

### What's Complete

| Component | Status | Notes |
|-----------|--------|-------|
| 7 Super Scanners | ✅ Complete | code-packages, code-security, code-quality, devops, technology-identification, code-ownership, devx |
| 12 Specialist Agents | ✅ Complete | Cereal, Razor, Gill, Turing, Blade, Phreak, Acid, Dade, Nikon, Joey, Plague, Gibson |
| Agent CLI (`./zero agent`) | ✅ Complete | Interactive agent mode with Zero orchestrator |
| Agent Self-Awareness | ✅ Complete | GetSystemInfo tool - agents can query RAG patterns, scanners, feeds, config |
| RAG Pattern System | ✅ Complete | 23 categories, 400+ patterns |
| Evidence Reports | ✅ Complete | HTML reports via Evidence.dev |
| Hydrate Command | ✅ Complete | Clone + scan with profiles |
| Freshness Tracking | ✅ Complete | Fresh/stale/expired indicators |
| Feed Sync | ✅ Complete | Semgrep rules, RAG rule generation |

### Maturity Levels

| Component | Status | Description |
|-----------|--------|-------------|
| **Scanners** | Alpha | 7 super scanners with 45+ features |
| **AI Agents** | Alpha | 12 specialist agents for deep analysis |
| **CLI** | Alpha | Core commands working, APIs may change |
| **Web UI** | Experimental | Next.js dashboard, expect breaking changes |

---

## Active Priorities

### Priority 1: Test Coverage

**Rationale:** Low coverage (6-47%) blocks confident releases and refactoring.

| Package | Current | Target | Complexity |
|---------|---------|--------|------------|
| `pkg/api/handlers` | 0% | 70% | Medium |
| `pkg/scanner/code-packages` | 8% | 70% | High |
| `pkg/scanner/code-security` | 28% | 70% | Medium |
| `pkg/core/scoring` | 0% | 70% | Low |
| `pkg/workflow/hydrate` | 17% | 70% | Medium |

**Approach:**
- Table-driven tests for scanner detection logic
- Mock external tools (semgrep, osv-scanner, cdxgen)
- Integration tests with test fixtures in `testdata/`

---

### Priority 2: Web UI Performance

**Rationale:** Dashboard unusable with current 500-2000ms response times.

**Plan:** `docs/PERFORMANCE-IMPLEMENTATION-PLAN.md`

| Phase | Description | Status |
|-------|-------------|--------|
| Phase 1 | SQLite storage layer | In Progress |
| Phase 2 | In-memory caching | Planned |
| Phase 3 | SWR request deduplication | Planned |

**Target:** <50ms API responses, <500ms dashboard load

---

### Priority 3: MCP Integration

**Rationale:** Enable Zero as MCP server for IDE integration.

**Current State:** Scaffolded in `pkg/mcp/`, not functional.

**Implementation:**
- MCP server exposing scanner results
- Tool definitions for each scanner
- Resource definitions for analysis data
- Integration with Claude Desktop / VS Code

---

### Priority 4: Reachability Analysis

**Rationale:** Prioritize actually-reachable vulnerabilities.

**Features:**
- Vulnerable code path detection
- Call graph analysis
- Risk prioritization based on reachability

---

## Planned Features

### Source Code Analysis

| Feature | Description | Status |
|---------|-------------|--------|
| Reachability Analysis | Trace calls to vulnerable functions | Planned |
| Dependency Graph Visualization | Interactive dependency explorer | Planned |
| Circular Dependency Detection | Find problematic cycles | Planned |
| Database Schema Analysis | Migration risks, schema drift | Future |
| Jupyter Notebook Security | Secrets in `.ipynb` files | Future |

### Cloud & Runtime

| Feature | Description | Status |
|---------|-------------|--------|
| Cloud Asset Inventory | AWS/Azure/GCP resource discovery | Planned |
| Cloud SBOM Generation | CycloneDX for cloud resources | Planned |
| Certificate Monitoring | Live SSL/TLS certificate expiry | Future |
| DNS Security | DNSSEC, SPF, DKIM, DMARC | Future |

### Reports & Analytics

| Feature | Description | Status |
|---------|-------------|--------|
| PDF Export | Executive summaries | Future |
| Trend Analysis | Track security posture over time | Future |
| Compliance Dashboards | SOC 2, ISO 27001, NIST mapping | Future |

---

## Integration Roadmap

### Ocular Integration

[Ocular](https://ocularproject.io) provides robust code synchronization at scale.

- Replace Zero's hydration with Ocular's code sync
- Leverage repository caching and versioning
- Support for monorepos and multi-repo projects

### Chalk Integration

[Chalk](https://github.com/crashappsec/chalk) provides build-time attestation.

- Build-time security analysis integration
- Attestation enrichment with Zero findings
- SLSA compliance verification

### Database Backend

- SQLite for single-user deployments
- DuckDB for analytics and dashboards
- PostgreSQL for enterprise multi-user

---

## Quick Reference

### CLI Commands

```bash
./zero hydrate owner/repo      # Clone and scan
./zero status                  # Check hydrated projects
./zero report owner/repo       # Generate HTML report
./zero agent                   # Enter agent mode
./zero feeds semgrep           # Sync Semgrep rules
./zero feeds rag               # Generate RAG rules
```

### Available Agents

| Agent | Domain | Scanner |
|-------|--------|---------|
| Cereal | Supply chain, vulnerabilities | code-packages |
| Razor | Code security, SAST, secrets | code-security |
| Gill | Cryptography, ciphers, TLS | code-security |
| Turing | AI/ML security, ML-BOM | technology-identification |
| Blade | Compliance, SOC 2, ISO 27001 | Multiple |
| Phreak | Legal, licenses, privacy | code-packages |
| Acid | Frontend, React, TypeScript | code-security, code-quality |
| Dade | Backend, APIs, databases | code-security |
| Nikon | Architecture, system design | technology-identification |
| Joey | Build, CI/CD, pipelines | devops |
| Plague | DevOps, infrastructure, K8s | devops |
| Gibson | DORA metrics, team health | devops, code-ownership |

---

## Contributing

1. **Submit Feature Requests**: [Create an issue](https://github.com/crashappsec/zero/issues/new)
2. **Comment on Existing Items**: Add use cases and implementation ideas
3. **Vote with Reactions**: Use 👍 to help prioritize
4. **Contribute Code**: Pick up any roadmap item and submit a PR

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

---

*"Hack the planet!"*
