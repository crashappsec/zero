<!--
SPDX-License-Identifier: GPL-3.0
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
-->

# Zero

> **"Hack the planet!"** - Developer intelligence platform for repository analysis, powered by AI agents

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Status: Alpha](https://img.shields.io/badge/Status-Alpha-orange.svg)](https://github.com/crashappsec/zero)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8.svg)](https://go.dev)

Named after **Zero Cool** from the movie Hackers (1995), Zero provides engineering intelligence tools and specialist AI agents for comprehensive repository assessment.

### Maturity

| Component | Status | Notes |
|-----------|--------|-------|
| **Scanners** | Alpha | 7 super scanners, 45+ features, changing fast |
| **AI Agents** | Alpha | 12 specialists for deep analysis |
| **CLI** | Alpha | Core commands working, APIs may change |
| **Reports** | Experimental | HTML reports, expect breaking changes |

## What is Zero?

Zero is a Go-based CLI tool for software engineers. It provides 7 consolidated "super scanners" with 45+ configurable features, AI-powered analysis agents, and integrates with tools like cdxgen, syft, semgrep, and grype to provide comprehensive engineering intelligence.

### Key Capabilities

- **7 Super Scanners** - Consolidated scanners covering dependencies, security, quality, DevOps, technology identification, code ownership, and developer experience
- **AI Agent System** - 12 specialist agents (named after Hackers characters) for deep analysis
- **Configurable** - JSON configuration for scanner options, profiles, and feature toggles
- **ML-BOM Generation** - Machine Learning Bill of Materials for AI/ML projects

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/crashappsec/zero.git
cd zero

# Build the CLI
go build -o main ./cmd/zero

# Check prerequisites and install missing tools
./main checkup --fix

# Verify your GitHub token and see what scanners will work
./main checkup
```

### Prerequisites

**Required:**
- Go 1.22+
- Git
- GitHub CLI (`gh`) - for authentication

**Recommended Tools** (install with `./main checkup --fix`):
| Tool | Purpose | Install |
|------|---------|---------|
| [cdxgen](https://github.com/CycloneDX/cdxgen) | SBOM generation (preferred) | `npm install -g @cyclonedx/cdxgen` |
| [syft](https://github.com/anchore/syft) | SBOM generation (fallback) | `brew install syft` |
| [grype](https://github.com/anchore/grype) | Vulnerability scanning | `brew install grype` |
| [osv-scanner](https://github.com/google/osv-scanner) | Vulnerability scanning | `go install github.com/google/osv-scanner/cmd/osv-scanner@latest` |
| [semgrep](https://github.com/returntocorp/semgrep) | Code security scanning | `brew install semgrep` |
| [gitleaks](https://github.com/gitleaks/gitleaks) | Secrets detection | `brew install gitleaks` |
| [malcontent](https://github.com/chainguard-dev/malcontent) | Supply chain malware detection | `go install github.com/chainguard-dev/malcontent/cmd/mal@latest` |
| [trivy](https://github.com/aquasecurity/trivy) | Container scanning | `brew install trivy` |
| [checkov](https://github.com/bridgecrewio/checkov) | IaC security | `pip install checkov` |

### Basic Usage

```bash
# Hydrate (clone and scan) a repository
./main hydrate <owner/repo>

# With analysis profiles (profile is a positional argument)
./main hydrate <owner/repo> all-quick       # All scanners, limited features (~2min)
./main hydrate <owner/repo> all-complete    # All scanners, all features (~12min)
./main hydrate <owner/repo> code-packages   # SBOM + dependency analysis
./main hydrate <owner/repo> code-security   # Security scanning only

# Scan an entire GitHub organization (no "/" means org)
./main hydrate <org>                        # All repos in org
./main hydrate <org> all-quick              # With profile
./main hydrate <org> --limit 10             # Limit repos

# Check status of analyzed projects
./main status

# Generate reports
./main report <owner/repo>

# See what scanners work with your token
./main checkup

# List all available scanners
./main list
```

## Commands

| Command | Description |
|---------|-------------|
| `hydrate <target> [profile]` | Clone and scan (target: `owner/repo` or `org-name`) |
| `scan <target> [profile]` | Re-scan already-cloned repos |
| `status` | Show all analyzed projects |
| `report <owner/repo>` | Generate security report |
| `checkup` | Check setup, token permissions, and install missing tools |
| `list` | List all available scanners |
| `clean <owner/repo>` | Remove analysis data |
| `history <owner/repo>` | Show scan history |

**Target detection:** If target contains `/`, it's a single repo. Otherwise, it's an organization.

Profiles are defined in `config/zero.config.json` and can be customized.

## Scanners

Zero uses **7 consolidated super scanners** (v4.0 architecture), each with multiple configurable features:

| Scanner | Features | Description | External Tools |
|---------|----------|-------------|----------------|
| **code-packages** | generation, integrity, vulns, health, licenses, malcontent, confusion, typosquats, deprecations, duplicates, reachability, provenance, bundle, recommendations | SBOM generation + package/dependency analysis | cdxgen, syft, grype, osv-scanner, malcontent |
| **code-security** | vulns, secrets, api, ciphers, keys, random, tls, certificates | Code analysis + cryptographic security | semgrep, gitleaks |
| **code-quality** | tech_debt, complexity, test_coverage, documentation | Code quality metrics | - |
| **devops** | iac, containers, github_actions, dora, git | DevOps and CI/CD analysis | trivy, checkov |
| **technology-identification** | detection, models, frameworks, datasets, ai_security, ai_governance, infrastructure | Technology detection and ML-BOM generation | - |
| **code-ownership** | contributors, bus_factor, codeowners, orphans, churn, patterns | Code ownership analysis | - |
| **devx** | onboarding, sprawl, workflow | Developer experience analysis (depends on technology-identification) | - |

### Feature Details

**code-packages** (generates SBOM + analyzes dependencies):
- `generation` - SBOM generation in CycloneDX format
- `vulns` - CVE scanning via OSV database
- `health` - Dependency health scoring, abandonment detection
- `licenses` - SPDX license detection with policy enforcement
- `malcontent` - Malware detection (14,500+ YARA rules)
- `provenance` - SLSA attestations and build provenance
- `bundle` - JavaScript bundle size analysis
- `typosquats` - Package name typosquatting detection
- `confusion` - Dependency confusion detection

**code-security** (includes cryptography analysis):
- `vulns` - SAST analysis (OWASP Top 10, CWE)
- `secrets` - API keys, credentials, token detection
- `api` - OWASP API Security Top 10
- `ciphers` - Weak/deprecated algorithms (DES, RC4, MD5)
- `keys` - Hardcoded keys, weak key lengths
- `random` - Insecure random number generation
- `tls` - TLS configuration issues
- `certificates` - X.509 certificate analysis

**technology-identification** (generates ML-BOM):
- `detection` - Language and framework detection (100+ patterns)
- `models` - ML model file detection (.pt, .onnx, .safetensors, .gguf)
- `frameworks` - AI/ML framework detection (PyTorch, TensorFlow, LangChain)
- `datasets` - Training dataset detection
- `ai_security` - Pickle RCE, unsafe model loading, API key exposure
- `ai_governance` - Model cards, licenses, dataset provenance

## Configuration

### Environment Variables

```bash
# GitHub authentication (required for GitHub API scanners)
export GITHUB_TOKEN="ghp_..."
# Or use: gh auth login

# Claude API key (for AI agent analysis)
export ANTHROPIC_API_KEY="sk-ant-..."
```

### Configuration File

Zero uses `config/zero.config.json` for scanner configuration. Each scanner has multiple features that can be enabled/disabled:

```json
{
  "settings": {
    "default_profile": "standard",
    "scanner_timeout_seconds": 300,
    "parallel_scanners": 4
  },
  "scanners": {
    "code-packages": {
      "features": {
        "generation": { "enabled": true, "tool": "auto", "spec_version": "1.5" },
        "integrity": { "enabled": true, "verify_lockfiles": true },
        "vulns": { "enabled": true, "include_dev": false },
        "health": { "enabled": true },
        "malcontent": { "enabled": true, "min_risk": "medium" },
        "licenses": { "enabled": true, "blocked_licenses": ["GPL-3.0", "AGPL-3.0"] }
      }
    },
    "code-security": {
      "features": {
        "vulns": { "enabled": true },
        "secrets": { "enabled": true },
        "ciphers": { "enabled": true }
      }
    }
  }
}
```

### Scan Profiles

| Profile | Scanners | Description |
|---------|----------|-------------|
| `all-quick` | All 7 scanners (limited features) | Fast scan of everything (~2min) |
| `all-complete` | All 7 scanners (all features) | Complete analysis (~12min) |
| `code-packages` | code-packages | SBOM + dependency analysis |
| `code-security` | code-security | SAST, secrets, and crypto |
| `code-quality` | code-quality | Quality metrics |
| `devops` | devops | IaC, containers, CI/CD |
| `technology-identification` | technology-identification | Technology detection, ML-BOM |
| `code-ownership` | code-ownership | Contributor analysis |
| `developer-experience` | technology-identification, devx | Developer experience |

## Checkup Command

The `checkup` command helps you understand what scanners will work with your current setup:

```bash
./main checkup
```

This shows:
- **Token Status** - Whether your GitHub token is valid and its type (classic PAT, fine-grained PAT, OAuth)
- **Token Permissions** - Scopes and permissions available
- **External Tools** - Which required tools are installed
- **Scanner Compatibility** - Which scanners are ready, limited, or unavailable
- **Recommendations** - What permissions or tools to add

## AI Agent System

Zero includes 12 specialist AI agents (powered by Claude) for deep analysis:

| Agent | Character | Expertise | Primary Scanner |
|-------|-----------|-----------|-----------------|
| **Zero** | Zero Cool | Master orchestrator | All |
| **Cereal** | Cereal Killer | Supply chain, malware, CVEs | code-packages |
| **Razor** | Razor | Code security, SAST, secrets | code-security |
| **Blade** | Blade | Compliance, SOC 2, ISO 27001 | code-packages, code-security |
| **Phreak** | Phantom Phreak | Legal, licenses, data privacy | code-packages (licenses) |
| **Acid** | Acid Burn | Frontend, React, TypeScript, a11y | code-security, code-quality |
| **Dade** | Dade Murphy | Backend, APIs, databases | code-security (api) |
| **Nikon** | Lord Nikon | Architecture, system design | technology-identification |
| **Joey** | Joey | Build, CI/CD, performance | devops (github_actions) |
| **Plague** | The Plague | DevOps, infrastructure, Kubernetes | devops |
| **Gibson** | The Gibson | Engineering metrics, DORA | devops (dora, git), code-ownership |
| **Gill** | Gill Bates | Cryptography, TLS, keys | code-security (crypto) |
| **Turing** | Alan Turing | AI/ML security, ML-BOM, LLM security | technology-identification |

### Agent Mode (Claude Code)

Use the `/agent` slash command in Claude Code to chat with Zero:

```
You: Do we have any malware in our dependencies?

Zero: Let me check what projects are loaded and delegate to Cereal...
[Invokes Cereal agent to investigate malcontent findings]

Cereal: I've analyzed the malcontent scan results. Found 3 high-risk
behaviors flagged, but after reading the source files, all appear to be
false positives related to legitimate test fixtures...
```

## Storage

Analysis data is stored in `.zero/` (configurable):

```
.zero/
├── index.json                  # Project index
└── repos/
    └── expressjs/
        └── express/
            ├── project.json    # Project metadata
            ├── repo/           # Cloned repository
            └── analysis/       # Scanner results (JSON)
                ├── sbom.cdx.json                    # CycloneDX SBOM
                ├── code-packages.json               # Package analysis results
                ├── code-security.json               # Code security + crypto results
                ├── code-quality.json                # Code quality results
                ├── devops.json                      # DevOps analysis results
                ├── technology-identification.json   # Technology/ML-BOM results
                ├── code-ownership.json              # Ownership results
                └── devx.json                        # Developer experience results
```

## Architecture

```
zero/
├── cmd/zero/                   # CLI entry point
│   └── cmd/                    # Cobra commands
├── pkg/
│   ├── scanner/                # Scanner framework + implementations
│   │   ├── code-packages/      # SBOM + package/dependency analysis
│   │   ├── code-security/      # Code security (SAST, secrets, crypto)
│   │   ├── code-quality/       # Code quality metrics
│   │   ├── devops/             # DevOps and CI/CD
│   │   ├── technology-identification/  # Technology detection, ML-BOM
│   │   ├── code-ownership/     # Code ownership
│   │   └── developer-experience/  # Developer experience
│   ├── core/                   # Core packages (config, terminal, etc.)
│   ├── workflow/               # Hydrate, automation, freshness
│   └── api/                    # REST API and handlers
├── agents/                     # AI agent definitions
├── rag/                        # RAG knowledge base
├── config/                     # Configuration files
```

## Development

### Building

```bash
# Build
go build -o main ./cmd/zero

# Run tests
go test ./...

# Run specific scanner tests
go test ./pkg/scanner/code-packages/...
```

### Adding a New Scanner

1. Create a new package in `pkg/scanner/<name>/`
2. Implement the `scanner.Scanner` interface:
   ```go
   type Scanner interface {
       Name() string
       Description() string
       Dependencies() []string
       EstimateDuration(fileCount int) time.Duration
       Run(ctx context.Context, opts *ScanOptions) (*ScanResult, error)
   }
   ```
3. Register in `init()`:
   ```go
   func init() {
       scanner.Register(&MyScanner{})
   }
   ```
4. Import in `pkg/scanner/all.go`

## Test Organization

We maintain [phantom-tests](https://github.com/phantom-tests) for safe testing:

```bash
./main hydrate phantom-tests/express
./main hydrate phantom-tests/platform
```

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

**Important:** All contributors must complete our [Contributor License Agreement](https://crashoverride.com/docs/other/contributing).

## License

Zero is licensed under the [GNU General Public License v3.0](./LICENSE).

```
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
```

## About

Zero is maintained by the open source community and sponsored by [Crash Override](https://crashoverride.com).

---

**Status**: Alpha (Reports: Experimental)
**Version**: 4.0.0 (Super Scanner Architecture)
**Last Updated**: 2026-01-04

*"Hack the planet!"*
