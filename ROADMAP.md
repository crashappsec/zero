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
| Hydrate Command | ✅ Complete | Clone + scan with profiles |
| Freshness Tracking | ✅ Complete | Fresh/stale/expired indicators |
| Feed Sync | ✅ Complete | Semgrep rules, RAG rule generation |

### Scanner Features

#### code-packages (14 features)
| Feature | Description | Status |
|---------|-------------|--------|
| generation | SBOM generation via cdxgen/syft | ✅ |
| integrity | Lock file integrity verification | ✅ |
| vulns | Vulnerability scanning via OSV.dev | ✅ |
| health | Package health scores via deps.dev | ✅ |
| licenses | License detection and compliance | ✅ |
| malcontent | Supply chain malware detection | ✅ |
| confusion | Dependency confusion detection | ✅ |
| typosquats | Typosquatting detection | ✅ |
| deprecations | Deprecated package detection | ✅ |
| duplicates | Duplicate dependency detection | ✅ |
| reachability | Vulnerable code path detection | ⏳ Planned |
| provenance | SLSA provenance verification | ✅ |
| bundle | Bundle size analysis | ✅ |
| recommendations | Package replacement suggestions | ✅ |

#### code-security (8 features)
| Feature | Description | Status |
|---------|-------------|--------|
| vulns | SAST via Semgrep (OWASP, CWE) | ✅ |
| secrets | Secret detection + git history | ✅ |
| api | API security (auth, injection, CORS) | ✅ |
| ciphers | Weak/deprecated cipher detection | ✅ |
| keys | Hardcoded cryptographic keys | ✅ |
| random | Insecure random number generation | ✅ |
| tls | TLS version and cipher suite analysis | ✅ |
| certificates | Certificate validation issues | ✅ |

#### code-quality (4 features)
| Feature | Description | Status |
|---------|-------------|--------|
| tech_debt | TODO/FIXME/HACK marker detection | ✅ |
| complexity | Cyclomatic/cognitive complexity | ✅ |
| test_coverage | Coverage report parsing | ⚠️ Basic |
| documentation | Doc comment coverage | ✅ |

#### devops (5 features)
| Feature | Description | Status |
|---------|-------------|--------|
| iac | IaC scanning via Checkov/Trivy | ✅ |
| containers | Dockerfile security linting | ✅ |
| github_actions | Action pinning, secrets, permissions | ✅ |
| dora | DORA metrics (deploy freq, lead time, MTTR, CFR) | ✅ |
| git | Git activity and contributor patterns | ✅ |

#### technology-identification (7 features)
| Feature | Description | Status |
|---------|-------------|--------|
| detection | Language/framework/tool detection | ✅ |
| models | ML model inventory (.pt, .onnx, .safetensors) | ✅ |
| frameworks | AI/ML framework detection | ✅ |
| datasets | Training dataset detection | ✅ |
| ai_security | Pickle RCE, unsafe loading patterns | ✅ |
| ai_governance | Model cards, responsible AI checks | ✅ |
| infrastructure | Microservice mapping, API contracts | ✅ |

#### code-ownership (6 features)
| Feature | Description | Status |
|---------|-------------|--------|
| contributors | Git contributor analysis | ✅ |
| bus_factor | Key person risk calculation | ✅ |
| codeowners | CODEOWNERS file validation | ✅ |
| orphans | Files without active maintainers | ✅ |
| churn | High-churn file detection | ✅ |
| patterns | Commit timing and patterns | ✅ |

#### developer-experience (3 features)
| Feature | Description | Status |
|---------|-------------|--------|
| onboarding | README quality, setup friction | ✅ |
| sprawl | Tool and technology sprawl analysis | ✅ |
| workflow | PR templates, local dev, hot reload | ✅ |

### Maturity Levels

| Component | Status | Description |
|-----------|--------|-------------|
| **Scanners** | Alpha | 7 super scanners with 45+ features |
| **AI Agents** | Alpha | 12 specialist agents for deep analysis |
| **CLI** | Alpha | Core commands working, APIs may change |
| **Web UI** | Experimental | Next.js dashboard, expect breaking changes |

---

## Active Priorities

### Priority 1: RAG System Improvements

**Rationale:** RAG patterns are the foundation for both agent knowledge and automated scanning.

#### Phase 1-3: ✅ Complete (branch: `rag-improvements`)

| Task | Status |
|------|--------|
| Pattern validator framework (`pkg/core/rag/validator.go`) | ✅ Complete |
| Test fixtures (`testdata/rag/`) | ✅ Complete |
| Fix regex cleaning (remove <4 char filter) | ✅ Complete |
| Expand language support (18 languages) | ✅ Complete |
| Dynamic RAG category discovery (23+ categories) | ✅ Complete |
| Preserve severity metadata (`original_severity`, `is_critical`) | ✅ Complete |
| Fix hardcoded paths in RAG loader | ✅ Complete |
| Agent RAG search capability (`rag-search`) | ✅ Complete |

**Result:** Rule generation now produces **3497 rules** (major increase).

#### Phase 4: Testing & Validation (Current)

| Task | Status |
|------|--------|
| Detection tests (`pkg/core/rag/detection_test.go`) | ⏳ In Progress |
| Integration tests (RAG → Semgrep → Run → Verify) | ⏳ In Progress |
| False positive test suite | ⏳ In Progress |
| Run generated rules against real repositories | ⏳ In Progress |

#### Phase 5: Pattern Coverage Expansion (Next)

| Task | Status |
|------|--------|
| Add missing language patterns (Rust, C#, PHP, Java) | Planned |
| Semantic Semgrep patterns (dataflow, multi-line) | Planned |
| Community rule sync implementation | Planned |

**Files:**
- `rag/` - 23 categories, 400+ pattern files
- `pkg/core/rag/validator.go` - Pattern validation
- `pkg/core/rag/loader.go` - RAG loading with path discovery
- `pkg/scanner/technology-identification/rag_converter.go` - RAG to Semgrep
- `pkg/agent/system_info.go` - Agent RAG search

---

### Priority 2: Test Coverage

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

### Priority 3: Web UI Completion

**Rationale:** Web UI is partially built and needs to be finished.

**Current State:**
- Next.js dashboard scaffolded
- API handlers exist but incomplete
- Performance issues (500-2000ms response times)

**Tasks:**
- Complete dashboard pages and components
- Finish API endpoints for all scanner data
- Add SQLite storage layer for performance
- Implement SWR request deduplication
- Polish UI/UX

**Target:** Functional dashboard with <500ms load times

---

### Priority 4: MCP Integration

**Rationale:** Enable Zero as MCP server for IDE integration.

**Current State:** Scaffolded in `pkg/mcp/`, not functional.

**Implementation:**
- MCP server exposing scanner results
- Tool definitions for each scanner
- Resource definitions for analysis data
- Integration with Claude Desktop / VS Code

---

### Priority 5: Reachability Analysis

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
