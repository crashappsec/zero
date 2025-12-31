# Scanner Reference

Complete reference for Zero's 8 super scanners. Each scanner provides multiple features that can be individually enabled or disabled.

## Architecture Overview

Zero uses a consolidated "super scanner" architecture where each scanner handles a specific domain with multiple configurable features:

```
┌─────────────────────────────────────────────────────────────────┐
│                        Zero Scanner Engine                       │
├─────────────────────────────────────────────────────────────────┤
│  sbom ──► package-analysis ──► [parallel scanners]              │
│                                                                  │
│  Parallel: crypto, code-security, code-quality, devops,         │
│            tech-id, code-ownership                               │
└─────────────────────────────────────────────────────────────────┘
```

The `sbom` scanner runs first as the **source of truth**, and `package-analysis` depends on its output. All other scanners can run in parallel.

## Super Scanners

| Scanner | Features | Output | Docs |
|---------|----------|--------|------|
| **sbom** | generation, integrity | `sbom.json` + `sbom.cdx.json` | [sbom.md](sbom.md) |
| **package-analysis** | vulns, health, licenses, malcontent, confusion, typosquats, deprecations, duplicates, reachability, provenance, bundle, recommendations | `package-analysis.json` | [package-analysis.md](package-analysis.md) |
| **crypto** | ciphers, keys, random, tls, certificates | `crypto.json` | [crypto.md](crypto.md) |
| **code-security** | vulns, secrets, api, git_history_security | `code-security.json` | [code-security.md](code-security.md) |
| **code-quality** | tech_debt, complexity, test_coverage, documentation | `code-quality.json` | [code-quality.md](quality.md) |
| **devops** | iac, containers, github_actions, dora, git | `devops.json` | [devops.md](devops.md) |
| **tech-id** | technology, models, frameworks, datasets, security, governance, semgrep_rules | `technology.json` | [tech-id.md](technology.md) |
| **code-ownership** | contributors, bus_factor, codeowners, orphans, churn, patterns | `code-ownership.json` | [code-ownership.md](ownership.md) |

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

### Package Analysis Scanner

Supply chain security analysis for dependencies.

```bash
./zero scan --scanner package-analysis /path/to/repo  # Requires sbom
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
- Git history security scanning (gitignore violations, sensitive files, purge recommendations)

[Full Documentation →](code-security.md)

---

### Code Quality Scanner

Code quality, test coverage, and documentation analysis.

```bash
./zero scan --scanner code-quality /path/to/repo
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

### Tech-ID Scanner

Technology identification and AI/ML analysis with ML-BOM generation.

```bash
./zero scan --scanner tech-id /path/to/repo
```

**Key Features:**
- Semgrep-powered technology detection using RAG patterns
- AI model detection with false positive reduction
- ML framework and dataset tracking
- AI security (pickle files, unsafe loading, API key exposure)
- AI governance (model cards, licenses)
- ML-BOM generation

[Full Documentation →](technology.md)

---

### Code Ownership Scanner

Code ownership and contributor analysis.

```bash
./zero scan --scanner code-ownership /path/to/repo
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

## Scan Profiles

Profiles combine scanners with specific feature configurations for common use cases.

| Profile | Scanners | Time | Use Case |
|---------|----------|------|----------|
| `quick` | sbom, package-analysis (limited) | ~30s | Fast initial assessment |
| `standard` | sbom, package-analysis, code-security, code-quality | ~2min | Balanced analysis |
| `security` | sbom, package-analysis, crypto, code-security, devops | ~4min | Security assessment |
| `full` | All 8 scanners | ~12min | Complete analysis |
| `ai-security` | sbom, package-analysis, code-security, tech-id | ~3min | AI/ML projects |
| `supply-chain` | sbom, package-analysis, tech-id | ~2min | Supply chain analysis |
| `compliance` | sbom, package-analysis, tech-id | ~2min | License/compliance |

### Profile-Specific Configurations

Profiles can override default feature settings. For example, the `quick` profile disables slower features:

```json
{
  "quick": {
    "scanners": ["sbom", "package-analysis"],
    "feature_overrides": {
      "package-analysis": {
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
    "package-analysis": {
      "features": {
        "vulns": {"enabled": true},
        "health": {"enabled": true},
        "licenses": {"enabled": true, "blocked_licenses": ["GPL-3.0"]}
      }
    },
    "tech-id": {
      "features": {
        "technology": {"enabled": true},
        "models": {"enabled": true},
        "security": {"enabled": true}
      }
    },
    "code-ownership": {
      "features": {
        "contributors": {"enabled": true, "period_days": 90},
        "bus_factor": {"enabled": true},
        "codeowners": {"enabled": true}
      }
    },
    "code-quality": {
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
├── sbom.json              # SBOM summary
├── sbom.cdx.json          # Full CycloneDX SBOM
├── package-analysis.json  # Package analysis
├── crypto.json            # Crypto analysis
├── code-security.json     # Code security
├── code-quality.json      # Code quality (includes tests, docs)
├── devops.json            # DevOps analysis
├── tech-id.json           # Technology ID + AI/ML (ML-BOM)
└── code-ownership.json    # Code ownership
```

## Scanner Relationships

```
    ┌───────────────┐   ┌───────────────┐   ┌───────────────┐   ┌───────────────┐
    │ code-quality  │   │code-ownership │   │    tech-id    │   │    crypto     │
    │  (debt,tests, │   │ (contributors,│   │ (detection,   │   │(ciphers,keys, │
    │   docs,cplx)  │   │  bus factor)  │   │  AI/ML, MLBOM)│   │  tls,certs)   │
    └───────────────┘   └───────────────┘   └───────────────┘   └───────────────┘

    ┌───────────────┐   ┌───────────────┐   ┌─────────────────────────────────┐
    │ code-security │   │    devops     │   │       package-analysis          │
    │(vulns,secrets,│   │(iac,containers│   │    (vulns, health, licenses,    │
    │     api)      │   │  gha,dora)    │   │     malcontent, confusion)      │
    └───────────────┘   └───────────────┘   └─────────────────┬───────────────┘
                                                              │
                                                              ▼
                                                      ┌───────────────┐
                                                      │     sbom      │
                                                      │  (source of   │
                                                      │    truth)     │
                                                      └───────────────┘
```

## See Also

- [Scanner Architecture](../architecture/scanners.md) - How scanners work
- [Output Formats](output-formats.md) - JSON output schemas
- [Getting Started](../GETTING_STARTED.md) - Quick start guide
- [RAG Technology Patterns](../../rag/technology-identification/README.md) - Technology detection patterns
