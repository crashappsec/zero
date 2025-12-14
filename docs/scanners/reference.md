# Scanner Reference

Complete reference for Zero's 9 super scanners. Each scanner provides multiple features that can be individually enabled or disabled.

## Architecture Overview

Zero uses a consolidated "super scanner" architecture where each scanner handles a specific domain with multiple configurable features:

```
┌─────────────────────────────────────────────────────────────────┐
│                        Zero Scanner Engine                       │
├─────────────────────────────────────────────────────────────────┤
│  sbom ──► packages ──► [parallel scanners] ──► health           │
│                                                                  │
│  Parallel: crypto, code-security, quality, devops,              │
│            technology, ownership                                 │
└─────────────────────────────────────────────────────────────────┘
```

The `sbom` scanner runs first as the **source of truth**, `packages` depends on its output, and `health` runs last to aggregate results from all other scanners.

## Super Scanners

| Scanner | Features | Output | Docs |
|---------|----------|--------|------|
| **sbom** | generation, integrity | `sbom.json` + `sbom.cdx.json` | [sbom.md](sbom.md) |
| **packages** | vulns, health, licenses, malcontent, confusion, typosquats, deprecations, duplicates, reachability, provenance, bundle, recommendations | `packages.json` | [packages.md](packages.md) |
| **crypto** | ciphers, keys, random, tls, certificates | `crypto.json` | [crypto.md](crypto.md) |
| **code-security** | vulns, secrets, api | `code-security.json` | [code-security.md](code-security.md) |
| **quality** | tech_debt, complexity, test_coverage, documentation | `quality.json` | [quality.md](quality.md) |
| **devops** | iac, containers, github_actions, dora, git | `devops.json` | [devops.md](devops.md) |
| **technology** | detection, models, frameworks, datasets, ai_security, ai_governance, infrastructure | `technology.json` | [technology.md](technology.md) |
| **ownership** | contributors, bus_factor, codeowners, orphans, churn, patterns | `ownership.json` | [ownership.md](ownership.md) |
| **health** | score, summary, recommendations, trends | `health.json` | [health.md](health.md) |

## Quick Reference

### SBOM Scanner

Generates Software Bill of Materials in CycloneDX format.

```bash
./zero scan --scanner sbom /path/to/repo
```

**Key Features:**
- CycloneDX 1.5 SBOM generation (cdxgen or syft)
- Lockfile integrity verification
- Drift detection

[Full Documentation →](sbom.md)

---

### Packages Scanner

Supply chain security analysis for dependencies.

```bash
./zero scan --scanner packages /path/to/repo  # Requires sbom
```

**Key Features:**
- Vulnerability scanning (OSV, CISA KEV)
- Package health assessment (deps.dev)
- License compliance checking
- Malcontent behavioral analysis
- Dependency confusion detection
- Typosquatting detection

[Full Documentation →](packages.md)

---

### Crypto Scanner

Cryptographic security analysis.

```bash
./zero scan --scanner crypto /path/to/repo
```

**Key Features:**
- Weak cipher detection (DES, MD5, SHA-1)
- Hardcoded key detection
- Insecure random number generation
- TLS misconfiguration
- Certificate analysis

[Full Documentation →](crypto.md)

---

### Code Security Scanner

Static Application Security Testing (SAST).

```bash
./zero scan --scanner code-security /path/to/repo
```

**Key Features:**
- Vulnerability detection (Semgrep)
- Secret detection with redaction
- API security analysis (OWASP API Top 10)

[Full Documentation →](code-security.md)

---

### Quality Scanner

Code quality, test coverage, and documentation analysis.

```bash
./zero scan --scanner quality /path/to/repo
```

**Key Features:**
- Technical debt markers (TODO, FIXME, HACK)
- Complexity analysis (cyclomatic, cognitive)
- Test coverage parsing and analysis
- Documentation quality scoring

[Full Documentation →](quality.md)

---

### DevOps Scanner

DevOps and CI/CD security analysis.

```bash
./zero scan --scanner devops /path/to/repo
```

**Key Features:**
- IaC security (Checkov/Trivy)
- Container image scanning
- GitHub Actions security
- DORA metrics calculation
- Git activity analysis

[Full Documentation →](devops.md)

---

### Technology Scanner

Technology identification and AI/ML analysis with ML-BOM generation.

```bash
./zero scan --scanner technology /path/to/repo
```

**Key Features:**
- RAG-based technology detection (119+ technologies)
- Tiered detection (quick, deep, extract)
- AI model detection and registry queries
- ML framework and dataset tracking
- AI security (pickle files, API key exposure)
- AI governance (model cards, licenses)
- ML-BOM generation

[Full Documentation →](technology.md)

---

### Ownership Scanner

Code ownership and contributor analysis.

```bash
./zero scan --scanner ownership /path/to/repo
```

**Key Features:**
- Contributor activity analysis
- Bus factor calculation
- CODEOWNERS parsing and validation
- Orphaned code detection
- Code churn analysis
- Commit pattern analysis

[Full Documentation →](ownership.md)

---

### Health Scanner

Aggregate project health scoring and recommendations.

```bash
./zero scan --scanner health /path/to/repo
```

**Key Features:**
- Aggregate health score (0-100)
- Multi-scanner metric aggregation
- Actionable recommendations
- Trend tracking

[Full Documentation →](health.md)

---

## Scan Profiles

Profiles combine scanners with specific feature configurations for common use cases.

| Profile | Scanners | Time | Use Case |
|---------|----------|------|----------|
| `quick` | sbom, packages, health | ~30s | Fast initial assessment |
| `standard` | sbom, packages, code-security, quality, health | ~2min | Balanced analysis |
| `security` | sbom, packages, crypto, code-security, devops | ~4min | Security assessment |
| `full` | All 9 scanners | ~12min | Complete analysis |
| `ai-security` | sbom, packages, code-security, technology | ~3min | AI/ML projects |
| `ownership-only` | ownership | ~1min | Ownership analysis |
| `quality-only` | quality | ~2min | Code quality only |

### Profile-Specific Configurations

Profiles can override default feature settings. For example, the `quick` profile disables slower features:

```json
{
  "quick": {
    "scanners": ["sbom", "packages", "health"],
    "feature_overrides": {
      "packages": {
        "malcontent": {"enabled": false},
        "provenance": {"enabled": false}
      }
    }
  }
}
```

See `config/zero.config.json` for all profile definitions.

## Tool Dependencies

| Tool | Used By | Purpose | Install |
|------|---------|---------|---------|
| cdxgen | sbom | SBOM generation | `npm i -g @cyclonedx/cdxgen` |
| syft | sbom | SBOM generation (fallback) | `brew install syft` |
| osv-scanner | packages | Vulnerability scanning | `go install github.com/google/osv-scanner/...` |
| malcontent | packages | Behavioral analysis | `brew install malcontent` |
| semgrep | crypto, code-security, quality | Pattern analysis | `pip install semgrep` |
| checkov | devops | IaC scanning | `pip install checkov` |
| trivy | devops | Container/IaC scanning | `brew install trivy` |

## Configuration

Scanner configuration is managed in `config/zero.config.json`:

```json
{
  "scanners": {
    "sbom": {
      "features": {
        "generation": {"enabled": true, "tool": "auto"},
        "integrity": {"enabled": true}
      }
    },
    "packages": {
      "features": {
        "vulns": {"enabled": true},
        "health": {"enabled": true},
        "licenses": {"enabled": true, "blocked_licenses": ["GPL-3.0"]}
      }
    },
    "technology": {
      "features": {
        "detection": {"enabled": true, "tier": "auto"},
        "models": {"enabled": true},
        "ai_security": {"enabled": true}
      }
    },
    "ownership": {
      "features": {
        "contributors": {"enabled": true, "period_days": 90},
        "bus_factor": {"enabled": true},
        "codeowners": {"enabled": true}
      }
    },
    "quality": {
      "features": {
        "tech_debt": {"enabled": true},
        "test_coverage": {"enabled": true},
        "documentation": {"enabled": true}
      }
    }
  }
}
```

## Output Location

All scanner outputs are written to the analysis directory:

```
.zero/repos/<project>/analysis/
├── sbom.json           # SBOM summary
├── sbom.cdx.json       # Full CycloneDX SBOM
├── packages.json       # Package analysis
├── crypto.json         # Crypto analysis
├── code-security.json  # Code security
├── quality.json        # Code quality (includes tests, docs)
├── devops.json         # DevOps analysis
├── technology.json     # Technology ID + AI/ML (ML-BOM)
├── ownership.json      # Code ownership
└── health.json         # Aggregate health score
```

## Scanner Relationships

```
                    ┌─────────────────────────────────────────┐
                    │                  health                  │
                    │    (aggregates all scanner outputs)      │
                    └─────────────────────────────────────────┘
                                        ▲
            ┌───────────────────────────┼───────────────────────────┐
            │                           │                           │
    ┌───────┴───────┐           ┌───────┴───────┐           ┌───────┴───────┐
    │    quality    │           │   ownership   │           │   technology  │
    │  (debt,tests, │           │ (contributors,│           │ (detection,   │
    │   docs,cplx)  │           │  bus factor)  │           │  AI/ML, MLBOM)│
    └───────────────┘           └───────────────┘           └───────────────┘

            ┌───────────────────────────────────────────────────────┐
            │                                                       │
    ┌───────┴───────┐   ┌───────────────┐   ┌───────────────┐   ┌───┴───────────┐
    │ code-security │   │    crypto     │   │    devops     │   │   packages    │
    │(vulns,secrets,│   │(ciphers,keys, │   │(iac,containers│   │(vulns,health, │
    │     api)      │   │  tls,certs)   │   │  gha,dora)    │   │licenses,etc)  │
    └───────────────┘   └───────────────┘   └───────────────┘   └───────┬───────┘
                                                                        │
                                                                        ▼
                                                                ┌───────────────┐
                                                                │     sbom      │
                                                                │(source of     │
                                                                │   truth)      │
                                                                └───────────────┘
```

## See Also

- [Scanner Architecture](../architecture/scanners.md) - How scanners work
- [Output Formats](output-formats.md) - JSON output schemas
- [Getting Started](../GETTING_STARTED.md) - Quick start guide
- [RAG Technology Patterns](../../rag/technology-identification/README.md) - Technology detection patterns
