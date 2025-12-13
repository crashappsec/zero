# Zero Super Scanner Domain Knowledge

This directory contains consolidated RAG knowledge organized by the 5 super scanner domains.

## Super Scanner Domains

| Scanner | File | Features | Primary Agents |
|---------|------|----------|----------------|
| **packages** | [packages.md](packages.md) | SBOM, vulnerabilities, health, malcontent, provenance, bundle, licenses, duplicates, recommendations, typosquats, deprecations | Cereal, Phreak |
| **crypto** | [crypto.md](crypto.md) | ciphers, keys, random, tls, certificates | Gill |
| **code** | [code.md](code.md) | vulns, secrets, api, tech_debt | Razor, Dade, Acid |
| **infra** | [infra.md](infra.md) | iac, containers, github_actions, dora, git | Plague, Joey, Gibson |
| **health** | [health.md](health.md) | technology, documentation, tests, ownership | Nikon, Gibson |

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Zero Orchestrator                         │
└───────────────────────────┬─────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
        ▼                   ▼                   ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│   packages    │   │    crypto     │   │     code      │
│  (11 features)│   │  (5 features) │   │  (4 features) │
└───────────────┘   └───────────────┘   └───────────────┘
        │                   │                   │
        ▼                   ▼                   ▼
┌───────────────┐   ┌───────────────┐
│     infra     │   │    health     │
│  (5 features) │   │  (4 features) │
└───────────────┘   └───────────────┘
```

## Feature Configuration

Each super scanner supports feature-level configuration:

```json
{
  "scanners": ["packages", "code"],
  "feature_overrides": {
    "packages": {
      "malcontent": { "enabled": false },
      "bundle": { "enabled": false }
    },
    "code": {
      "tech_debt": { "enabled": false }
    }
  }
}
```

## Output Structure

Each super scanner produces a single JSON file:

- `packages.json` - All package/dependency analysis
- `crypto.json` - All cryptographic security analysis
- `code.json` - All code security analysis
- `infra.json` - All infrastructure analysis
- `health.json` - All project health analysis

## Agent-Domain Mapping

| Agent | Primary Domain | Secondary Domains |
|-------|----------------|-------------------|
| Cereal | packages | - |
| Phreak | packages (licenses) | - |
| Gill | crypto | code (secrets) |
| Razor | code | crypto, packages |
| Dade | code (api) | infra |
| Acid | code (frontend) | - |
| Plague | infra (iac, containers) | - |
| Joey | infra (github_actions) | - |
| Gibson | infra (dora, git), health | - |
| Nikon | health (technology) | All domains |
| Blade | packages, code | All domains (compliance) |

## Related RAG Directories

The domain files reference existing RAG knowledge:

| Directory | Related Domain(s) |
|-----------|-------------------|
| `rag/supply-chain/` | packages |
| `rag/cryptography/` | crypto |
| `rag/code-security/` | code |
| `rag/api-security/` | code |
| `rag/secrets-scanner/` | code, crypto |
| `rag/dora/`, `rag/dora-metrics/` | infra |
| `rag/technology-identification/` | health |
| `rag/code-ownership/` | health |
| `rag/legal-review/` | packages |
| `rag/certificate-analysis/` | crypto |
