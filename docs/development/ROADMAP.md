<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Zero Roadmap

**Version:** 6.0.0
**Last Updated:** 2026-01-10

**Vision**: Position Zero as the leading **open-source engineering intelligence platform** — providing deep insights into software composition, security posture, delivery performance, and team health through the lens of industry frameworks (DORA, SPACE, DX Core 4).

---

## Engineering Intelligence Framework

Zero organizes analysis around **6 Pillars** that unify engineering productivity frameworks (DORA, SPACE, LinearB) with technical security analysis:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    ENGINEERING INTELLIGENCE FRAMEWORK                        │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                          │
│  │   SPEED     │  │  QUALITY    │  │   TEAM      │   ← Productivity         │
│  │   (Flow)    │  │  (Health)   │  │  (People)   │     Pillars              │
│  └─────────────┘  └─────────────┘  └─────────────┘                          │
│                                                                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                          │
│  │  SECURITY   │  │ SUPPLY CHAIN│  │ TECHNOLOGY  │   ← Technical            │
│  │  (Risk)     │  │ (Dependency)│  │  (Stack)    │     Pillars              │
│  └─────────────┘  └─────────────┘  └─────────────┘                          │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Pillar → Analyzer Mapping

| Pillar | Primary Analyzers | Key Metrics |
|--------|-------------------|-------------|
| **Speed** | devops | DORA metrics, cycle time, deploy frequency |
| **Quality** | code-quality | Tech debt, complexity, test coverage |
| **Team** | code-ownership, developer-experience | Bus factor, ownership, onboarding |
| **Security** | code-security | Vulnerabilities, secrets, crypto issues |
| **Supply Chain** | code-packages | Package health, licenses, malcontent |
| **Technology** | technology-identification | Stack detection, AI/ML security |

### Benchmark Tiers (LinearB 2026)

All metrics are classified into four performance tiers:

| Tier | Description | Color |
|------|-------------|-------|
| **Elite** | Top 25% performers | Green |
| **Good** | Above average | Blue |
| **Fair** | Average | Yellow |
| **Needs Focus** | Below average | Red |

---

## Current State

### What's Complete

| Component | Status | Notes |
|-----------|--------|-------|
| 7 Super Analyzers | ✅ Complete | code-packages, code-security, code-quality, devops, technology-identification, code-ownership, devx |
| 12 Specialist Agents | ✅ Complete | Cereal, Razor, Gill, Hal, Blade, Phreak, Acid, Flu Shot, Nikon, Joey, Plague, Gibson |
| 6-Pillar UI Navigation | ✅ Complete | Speed, Quality, Team, Security, Supply Chain, Technology |
| Benchmark Tier Component | ✅ Complete | Visual tier classification with LinearB benchmarks |
| Markdown Reports | ✅ Complete | CLI-generated markdown reports by category |
| Agent CLI (`./zero agent`) | ✅ Complete | Interactive agent mode with Zero orchestrator |
| RAG Pattern System | ✅ Complete | 23 categories, 400+ patterns |
| Hydrate Command | ✅ Complete | Clone + scan with profiles |
| Freshness Tracking | ✅ Complete | Fresh/stale/expired indicators |

### Analyzer Features (45+ total)

#### code-packages (14 features) → Supply Chain Pillar
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

#### code-security (8 features) → Security Pillar
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

#### devops (5 features) → Speed Pillar
| Feature | Description | Status |
|---------|-------------|--------|
| iac | IaC scanning via Checkov/Trivy | ✅ |
| containers | Dockerfile security linting | ✅ |
| github_actions | Action pinning, secrets, permissions | ✅ |
| dora | DORA metrics (deploy freq, lead time, MTTR, CFR) | ✅ |
| git | Git activity and contributor patterns | ✅ |

#### code-quality (4 features) → Quality Pillar
| Feature | Description | Status |
|---------|-------------|--------|
| tech_debt | TODO/FIXME/HACK marker detection | ✅ |
| complexity | Cyclomatic/cognitive complexity | ✅ |
| test_coverage | Coverage report parsing | ⚠️ Basic |
| documentation | Doc comment coverage | ✅ |

#### technology-identification (7 features) → Technology Pillar
| Feature | Description | Status |
|---------|-------------|--------|
| detection | Language/framework/tool detection | ✅ |
| models | ML model inventory (.pt, .onnx, .safetensors) | ✅ |
| frameworks | AI/ML framework detection | ✅ |
| datasets | Training dataset detection | ✅ |
| ai_security | Pickle RCE, unsafe loading patterns | ✅ |
| ai_governance | Model cards, responsible AI checks | ✅ |
| infrastructure | Microservice mapping, API contracts | ✅ |

#### code-ownership (6 features) → Team Pillar
| Feature | Description | Status |
|---------|-------------|--------|
| contributors | Git contributor analysis | ✅ |
| bus_factor | Key person risk calculation | ✅ |
| codeowners | CODEOWNERS file validation | ✅ |
| orphans | Files without active maintainers | ✅ |
| churn | High-churn file detection | ✅ |
| patterns | Commit timing and patterns | ✅ |

#### developer-experience (3 features) → Team Pillar
| Feature | Description | Status |
|---------|-------------|--------|
| onboarding | README quality, setup friction | ✅ |
| sprawl | Tool and technology sprawl analysis | ✅ |
| workflow | PR templates, local dev, hot reload | ✅ |

---

## Implementation Phases

### Phase 1: Framework Alignment (P0) - Current
**Status:** ✅ Complete

| Task | Status | Issue |
|------|--------|-------|
| Reorganize UI around 6 pillars | ✅ Complete | - |
| Create Speed page with DORA focus | ✅ Complete | - |
| Add BenchmarkTier component | ✅ Complete | - |
| Update sidebar navigation order | ✅ Complete | - |
| Remove Evidence.dev (legacy) | ✅ Complete | - |
| Add markdown report generator | ✅ Complete | - |

### Phase 2: Benchmark Visualization (P0) - In Progress
**Status:** 🔄 In Progress

| Task | Status | Issue |
|------|--------|-------|
| Add benchmark tiers to Security page | ⏳ Pending | [#55](https://github.com/crashappsec/zero/issues/55) |
| Add benchmark tiers to Supply Chain page | ⏳ Pending | [#55](https://github.com/crashappsec/zero/issues/55) |
| Add benchmark tiers to Quality page | ⏳ Pending | [#55](https://github.com/crashappsec/zero/issues/55) |
| Add benchmark tiers to Team page | ⏳ Pending | [#55](https://github.com/crashappsec/zero/issues/55) |
| Add benchmark reference footer to all pillar pages | ⏳ Pending | [#55](https://github.com/crashappsec/zero/issues/55) |

### Phase 3: PR-Level Metrics (P1)
**Status:** Planned

| Task | Status | Issue |
|------|--------|-------|
| Add pickup time per PR | Planned | [#56](https://github.com/crashappsec/zero/issues/56) |
| Add review time per PR | Planned | [#56](https://github.com/crashappsec/zero/issues/56) |
| Add merge time per PR | Planned | [#56](https://github.com/crashappsec/zero/issues/56) |
| Add PR size distribution | Planned | [#56](https://github.com/crashappsec/zero/issues/56) |
| Add rework rate calculation | Planned | [#57](https://github.com/crashappsec/zero/issues/57) |

### Phase 4: Composite Scores (P2)
**Status:** Planned

| Task | Status | Issue |
|------|--------|-------|
| Speed Score (0-100) | Planned | [#58](https://github.com/crashappsec/zero/issues/58) |
| Quality Score (0-100) | Planned | [#58](https://github.com/crashappsec/zero/issues/58) |
| Team Score (0-100) | Planned | [#58](https://github.com/crashappsec/zero/issues/58) |
| Security Score (0-100) | Planned | [#58](https://github.com/crashappsec/zero/issues/58) |
| Supply Chain Score (0-100) | Planned | [#58](https://github.com/crashappsec/zero/issues/58) |
| Technology Score (0-100) | Planned | [#58](https://github.com/crashappsec/zero/issues/58) |
| Executive dashboard with all scores | Planned | [#58](https://github.com/crashappsec/zero/issues/58) |

### Phase 5: Framework-Specific Reports (P3)
**Status:** Planned

| Task | Status | Issue |
|------|--------|-------|
| DORA-aligned report format | Planned | [#59](https://github.com/crashappsec/zero/issues/59) |
| LinearB-aligned report format | Planned | [#59](https://github.com/crashappsec/zero/issues/59) |
| SPACE-aligned report format | Planned | [#59](https://github.com/crashappsec/zero/issues/59) |
| Executive summary report | Planned | [#59](https://github.com/crashappsec/zero/issues/59) |
| Add --benchmark flag to CLI | Planned | [#59](https://github.com/crashappsec/zero/issues/59) |
| Add --framework flag to CLI | Planned | [#59](https://github.com/crashappsec/zero/issues/59) |

---

## Legacy Phases (From Modernization)

These phases are from the previous modernization effort and remain active:

### Phase 1-8: Modernization (Existing)

| Phase | Title | Status | Issue |
|-------|-------|--------|-------|
| Phase 1 | Non-Breaking Quick Wins | Open | #43 |
| Phase 2 | Hybrid Cache Architecture | Open | #44 |
| Phase 3 | Finding Validation System | ✅ Complete | #45 |
| Phase 4 | Flatten Analyzers + Knowledge Co-location | Open | #46 |
| Phase 5 | New Analyzers | Open | #47 |
| Phase 6 | Eliminate RAG → Semgrep Pipeline | Open | #48 |
| Phase 7 | Full Migration (scanner → analyzer) | Open | #49 |
| Phase 8 | Web UI & Polish | Open | #50 |

---

## Benchmark Reference

### LinearB 2026 Delivery Metrics

| Metric | Elite | Good | Fair | Needs Focus |
|--------|-------|------|------|-------------|
| Cycle Time | < 25h | 25-72h | 73-161h | > 161h |
| Deploy Time | < 16h | 16-106h | 107-277h | > 277h |
| Pickup Time | < 1h | 1-4h | 5-16h | > 16h |
| Review Time | < 3h | 3-14h | 15-24h | > 24h |
| Merge Time | < 1h | 1-3h | 4-16h | > 16h |
| Change Failure Rate | < 1% | 1-4% | 5-17% | > 17% |
| PR Size | < 100 | 100-155 | 156-228 | > 228 |
| Rework Rate | < 3% | 3-5% | 6-8% | > 8% |
| Deploy Freq/day | > 1.2 | 0.5-1.2 | 0.2-0.5 | < 0.2 |

### Zero Security Benchmarks

| Metric | Elite | Good | Fair | Needs Focus |
|--------|-------|------|------|-------------|
| Critical Vulnerabilities | 0 | 1-3 | 4-10 | > 10 |
| High Vulnerabilities | < 5 | 5-15 | 16-50 | > 50 |
| Secrets Exposed | 0 | 0 | 1-5 | > 5 |
| Weak Crypto Instances | 0 | 1-3 | 4-10 | > 10 |

### Zero Supply Chain Benchmarks

| Metric | Elite | Good | Fair | Needs Focus |
|--------|-------|------|------|-------------|
| Package Health | > 85% | 70-85% | 50-70% | < 50% |
| Vulnerable Dependencies | < 3% | 3-10% | 10-25% | > 25% |
| License Violations | 0 | 0-2 | 3-10 | > 10 |
| KEV (Exploited) Vulns | 0 | 0 | 1-2 | > 2 |

---

## Framework Alignment

### DORA Metrics Coverage

| Metric | Zero Analyzer | Feature | Status |
|--------|---------------|---------|--------|
| Deployment Frequency | devops | dora | ✅ |
| Lead Time for Changes | devops | dora | ✅ |
| Change Failure Rate | devops | dora | ✅ |
| Mean Time to Recovery | devops | dora | ✅ |
| Reliability | - | - | ❌ Gap |
| Rework Rate | devops | dora | ⏳ Planned |

### SPACE Dimensions Coverage

| Dimension | Zero Analyzer | Coverage |
|-----------|---------------|----------|
| **S**atisfaction | - | ❌ Requires surveys |
| **P**erformance | code-quality, code-security | ✅ Partial |
| **A**ctivity | code-ownership | ✅ Commits, PRs |
| **C**ommunication | code-ownership | ⚠️ Basic |
| **E**fficiency | devops, developer-experience | ✅ Cycle time, setup friction |

### LinearB Metrics Coverage

| Category | Coverage | Gaps |
|----------|----------|------|
| Delivery | ⚠️ Partial | PR-level breakdown (pickup, review, merge) |
| Predictability | ⚠️ Partial | Rework rate, capacity accuracy |
| Project Management | ❌ None | Needs issue tracker integration |

---

## Quick Reference

### CLI Commands

```bash
./zero hydrate owner/repo      # Clone and scan
./zero status                  # Check hydrated projects
./zero report owner/repo       # Generate markdown report
./zero report owner/repo --category security  # Category-specific report
./zero agent                   # Enter agent mode
./zero feeds semgrep           # Sync Semgrep rules
./zero feeds rag               # Generate RAG rules
```

### Available Agents

| Agent | Domain | Pillar |
|-------|--------|--------|
| Cereal | Supply chain, vulnerabilities | Supply Chain |
| Razor | Code security, SAST, secrets | Security |
| Gill | Cryptography, ciphers, TLS | Security |
| Hal | AI/ML security, ML-BOM | Technology |
| Blade | Compliance, SOC 2, ISO 27001 | Multiple |
| Phreak | Legal, licenses, privacy | Supply Chain |
| Acid | Frontend, React, TypeScript | Quality |
| Flu Shot | Backend, APIs, databases | Security |
| Nikon | Architecture, system design | Technology |
| Joey | Build, CI/CD, pipelines | Speed |
| Plague | DevOps, infrastructure, K8s | Speed |
| Gibson | DORA metrics, team health | Speed, Team |

---

## Contributing

1. **Submit Feature Requests**: [Create an issue](https://github.com/crashappsec/zero/issues/new)
2. **Comment on Existing Items**: Add use cases and implementation ideas
3. **Vote with Reactions**: Use 👍 to help prioritize
4. **Contribute Code**: Pick up any roadmap item and submit a PR

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

---

*"Hack the planet!"*
