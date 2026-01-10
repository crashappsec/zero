<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
SPDX-License-Identifier: GPL-3.0
-->

# Zero Documentation

Comprehensive documentation for Zero - an engineering intelligence platform for repository assessment.

## Documentation

| Document | Description |
|----------|-------------|
| [Getting Started](GETTING_STARTED.md) | Installation, setup, and first scan |
| [Docker](DOCKER.md) | Running Zero in containers |
| [Roadmap](ROADMAP.md) | Project roadmap and planned features |
| [Changelog](CHANGELOG.md) | Version history and release notes |

## Scanners

| Document | Description |
|----------|-------------|
| [Scanner Reference](scanners/reference.md) | Complete list of all scanners with options |
| [Output Formats](scanners/output-formats.md) | JSON schemas for scanner output |

## Quick Start

### Prerequisites

- Go 1.21+
- GitHub CLI (`gh`) - for authentication
- GitHub token (for cloning repositories)

```bash
# Authenticate with GitHub CLI (recommended)
gh auth login

# Or set your GitHub token directly
export GITHUB_TOKEN=ghp_your_token_here
```

### Build and Run

```bash
# Build the CLI
go build -o zero ./cmd/zero

# Verify it works
./zero --help
```

### Initialize Rules

Before scanning, sync Semgrep rules:

```bash
# Sync Semgrep community rules for SAST scanning
./zero feeds semgrep

# Generate rules from RAG knowledge base
./zero feeds rag
```

### Scan a Repository

```bash
# Scan a single repository
./zero hydrate strapi/strapi

# Scan with a specific profile
./zero hydrate strapi/strapi all-quick

# Check scan status
./zero status

# Generate markdown report
./zero report strapi/strapi
```

### Scan an Organization

```bash
# Scan all repos in an organization (default limit: 25)
./zero hydrate zero-test-org

# Demo mode: skip repos > 50MB, fetch replacements
./zero hydrate zero-test-org --demo
```

### Enter Agent Mode

In Claude Code, use the `/agent` slash command to chat with Zero and the specialist agents.

## Super Scanners (v4.0)

Zero uses 7 consolidated super scanners:

| Scanner | Description |
|---------|-------------|
| `code-packages` | SBOM generation + dependency analysis (vulns, licenses, malcontent, health) |
| `code-security` | SAST, secrets, API security, cryptography |
| `code-quality` | Tech debt, complexity, test coverage, documentation |
| `devops` | IaC security, containers, GitHub Actions, DORA metrics |
| `technology-identification` | Technology detection, ML-BOM generation |
| `code-ownership` | Contributors, bus factor, CODEOWNERS, code churn |
| `developer-experience` | Onboarding friction, tool sprawl, workflow |

## Scan Profiles

| Profile | Scanners | Time |
|---------|----------|------|
| `all-quick` | All 7 scanners (limited features) | ~2 min |
| `all-complete` | All 7 scanners (all features) | ~12 min |
| `code-packages` | SBOM + dependency analysis | ~1 min |
| `code-security` | SAST, secrets, API, cryptography | ~2 min |
| `devops` | IaC, containers, GitHub Actions, DORA | ~3 min |

## AI Agents

Zero includes 12 specialist agents named after characters from Hackers (1995):

| Agent | Character | Expertise |
|-------|-----------|-----------|
| **Zero** | Zero Cool | Master orchestrator |
| **Cereal** | Cereal Killer | Supply chain, CVEs, malware |
| **Razor** | Razor | Code security, SAST, secrets |
| **Gill** | Gill Bates | Cryptography, TLS, keys |
| **Hal** | Hal | AI/ML security, model safety |
| **Blade** | Blade | Compliance, SOC 2, ISO 27001 |
| **Phreak** | Phantom Phreak | Legal, licenses, privacy |
| **Acid** | Acid Burn | Frontend, React, TypeScript |
| **Flu Shot** | Flu Shot | Backend, APIs, databases |
| **Nikon** | Lord Nikon | Architecture, system design |
| **Joey** | Joey | CI/CD, build optimization |
| **Plague** | The Plague | DevOps, Kubernetes, IaC |
| **Gibson** | The Gibson | DORA metrics, team health |

## Data Storage

Zero stores data in `.zero/` (project local, override with `ZERO_HOME`):

```
.zero/
└── repos/
    └── owner/
        └── repo/
            ├── repo/           # Cloned repository
            └── analysis/       # Scanner output
                ├── sbom.cdx.json
                ├── code-packages.json
                ├── code-security.json
                ├── code-quality.json
                ├── devops.json
                ├── technology-identification.json
                ├── code-ownership.json
                └── developer-experience.json
```

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `GITHUB_TOKEN` | Yes | GitHub API access for cloning |
| `ANTHROPIC_API_KEY` | No | Claude API for AI-enhanced analysis |
| `ZERO_HOME` | No | Override default `.zero/` location |

## See Also

- [CLAUDE.md](../CLAUDE.md) - Claude Code configuration
- [Root README](../README.md) - Project overview
