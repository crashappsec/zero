# Scanner Reference

Complete reference for Zero's 7 super scanners (v4.0). Each scanner provides multiple features that can be individually enabled or disabled.

## Architecture Overview

Zero uses a consolidated "super scanner" architecture where each scanner handles a specific domain with multiple configurable features:

```
┌─────────────────────────────────────────────────────────────────┐
│                        Zero Scanner Engine                       │
├─────────────────────────────────────────────────────────────────┤
│  code-packages ──► [parallel scanners]                          │
│                                                                  │
│  Parallel: code-security, code-quality, devops,                 │
│            technology-identification, code-ownership, devx       │
└─────────────────────────────────────────────────────────────────┘
```

The `code-packages` scanner generates the SBOM as **source of truth**. All other scanners can run in parallel. The `devx` scanner depends on `technology-identification`.

## Super Scanners

| Scanner | Features | Output | Docs |
|---------|----------|--------|------|
| **code-packages** | generation, integrity, vulns, health, licenses, malcontent, confusion, typosquats, deprecations, duplicates, reachability, provenance, bundle, recommendations | `code-packages.json` + `sbom.cdx.json` | [code-packages.md](code-packages.md) |
| **code-security** | vulns, secrets, api, ciphers, keys, random, tls, certificates | `code-security.json` | [code-security.md](code-security.md) |
| **code-quality** | tech_debt, complexity, test_coverage, documentation | `code-quality.json` | [code-quality.md](quality.md) |
| **devops** | iac, containers, github_actions, dora, git | `devops.json` | [devops.md](devops.md) |
| **technology-identification** | detection, models, frameworks, datasets, ai_security, ai_governance, infrastructure | `technology-identification.json` | [technology-identification.md](technology.md) |
| **code-ownership** | contributors, bus_factor, codeowners, orphans, churn, patterns | `code-ownership.json` | [code-ownership.md](ownership.md) |
| **devx** | onboarding, sprawl, workflow | `devx.json` | [devx.md](devx.md) |

## Quick Reference

### Code Packages Scanner

SBOM generation and package/dependency analysis (consolidated from sbom + package-analysis).

```bash
./zero scan --scanner code-packages /path/to/repo
```

**Key Features:**
- CycloneDX 1.5 SBOM generation (cdxgen)
- Vulnerability scanning (osv-scanner, CISA KEV)
- Package health assessment (deps.dev)
- License compliance checking
- Malcontent behavioral analysis
- Dependency confusion and typosquatting detection

[Full Documentation →](code-packages.md)

---

### Code Security Scanner

Static analysis, secret detection, and cryptographic security (consolidated from code-security + crypto).

```bash
./zero scan --scanner code-security /path/to/repo
```

**Key Features:**
- Vulnerability detection (Semgrep)
- Secret detection with redaction
- API security analysis (OWASP API Top 10)
- Weak cipher detection (DES, MD5, SHA-1)
- Hardcoded key detection
- TLS misconfiguration analysis

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

### Technology Identification Scanner

Technology detection and AI/ML analysis with ML-BOM generation.

```bash
./zero scan --scanner technology-identification /path/to/repo
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

### DevX Scanner

Developer experience analysis (depends on technology-identification).

```bash
./zero scan --scanner devx /path/to/repo
```

**Key Features:**
- Onboarding friction analysis
- Tool sprawl detection
- Workflow efficiency assessment

[Full Documentation →](devx.md)

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
| `all-quick` | All 7 scanners (limited features) | ~2min | Fast initial assessment |
| `all-complete` | All 7 scanners (all features) | ~12min | Complete analysis |
| `code-packages` | code-packages | ~1min | SBOM + dependency analysis |
| `code-security` | code-security | ~2min | Security assessment |
| `technology-identification` | technology-identification | ~1min | Technology detection, ML-BOM |
| `developer-experience` | technology-identification, devx | ~2min | Developer experience |

### Profile-Specific Configurations

Profiles can override default feature settings. For example, the `all-quick` profile disables slower features:

```json
{
  "all-quick": {
    "scanners": ["code-packages", "code-security", "code-quality", "devops", "technology-identification", "code-ownership", "devx"],
    "feature_overrides": {
      "code-packages": {
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
| cdxgen | code-packages | SBOM generation | `npm i -g @cyclonedx/cdxgen` |
| osv-scanner | code-packages | Vulnerability scanning | `go install github.com/google/osv-scanner/...` |
| malcontent | code-packages | Behavioral analysis | `brew install malcontent` |
| semgrep | code-security, code-quality | Pattern analysis | `pip install semgrep` |
| trufflehog | code-security | Secret detection | `brew install trufflehog` |
| checkov | devops | IaC scanning | `pip install checkov` |
| trivy | devops | Container/IaC scanning | `brew install trivy` |

## Configuration

Scanner configuration is managed in `config/zero.config.json`:

```json
{
  "scanners": {
    "code-packages": {
      "features": {
        "generation": {"enabled": true, "tool": "auto"},
        "integrity": {"enabled": true},
        "vulns": {"enabled": true},
        "health": {"enabled": true},
        "licenses": {"enabled": true, "blocked_licenses": ["GPL-3.0"]}
      }
    },
    "code-security": {
      "features": {
        "vulns": {"enabled": true},
        "secrets": {"enabled": true},
        "ciphers": {"enabled": true},
        "keys": {"enabled": true}
      }
    },
    "technology-identification": {
      "features": {
        "detection": {"enabled": true},
        "models": {"enabled": true},
        "ai_security": {"enabled": true}
      }
    },
    "code-ownership": {
      "features": {
        "contributors": {"enabled": true, "period_days": 90},
        "bus_factor": {"enabled": true},
        "codeowners": {"enabled": true}
      }
    }
  }
}
```

## Output Location

All scanner outputs are written to the analysis directory:

```
.zero/repos/<project>/analysis/
├── sbom.cdx.json                    # Full CycloneDX SBOM
├── code-packages.json               # Package analysis
├── code-security.json               # Code security + crypto
├── code-quality.json                # Code quality
├── devops.json                      # DevOps analysis
├── technology-identification.json   # Technology + AI/ML (ML-BOM)
├── code-ownership.json              # Code ownership
└── devx.json                        # Developer experience
```

## Scanner Relationships

```
    ┌───────────────┐   ┌───────────────┐   ┌─────────────────────┐
    │ code-quality  │   │code-ownership │   │  technology-id      │
    │  (debt,tests, │   │ (contributors,│   │ (detection, AI/ML,  │
    │   docs,cplx)  │   │  bus factor)  │   │    MLBOM)           │
    └───────────────┘   └───────────────┘   └──────────┬──────────┘
                                                       │
    ┌───────────────┐   ┌───────────────┐   ┌──────────▼──────────┐
    │ code-security │   │    devops     │   │        devx         │
    │(vulns,secrets,│   │(iac,containers│   │(onboarding, sprawl, │
    │ crypto, api)  │   │  gha,dora)    │   │     workflow)       │
    └───────────────┘   └───────────────┘   └─────────────────────┘

                        ┌─────────────────────────────────────────┐
                        │             code-packages               │
                        │    (sbom, vulns, health, licenses,      │
                        │      malcontent, confusion)             │
                        │           [source of truth]             │
                        └─────────────────────────────────────────┘
```

## See Also

- [Scanner Architecture](../architecture/scanners.md) - How scanners work
- [Output Formats](output-formats.md) - JSON output schemas
- [Getting Started](../GETTING_STARTED.md) - Quick start guide
- [RAG Technology Patterns](../../rag/technology-identification/README.md) - Technology detection patterns
