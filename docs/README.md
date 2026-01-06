<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
SPDX-License-Identifier: GPL-3.0
-->

# Zero Documentation

Comprehensive documentation for Zero - an engineering intelligence platform for repository assessment.

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

### Scan a Repository

```bash
# Scan a single repository (uses default 'all-quick' profile)
./zero hydrate strapi/strapi

# Scan with a specific profile
./zero hydrate strapi/strapi all-quick

# Scan an organization (default limit: 25 repos)
./zero hydrate zero-test-org all-quick

# Demo mode: skip repos > 50MB, auto-fetch replacements
./zero hydrate zero-test-org all-quick --demo

# Check scan status
./zero status

# Start web UI to view reports
./zero serve
```

### Available Profiles

| Profile | Description | Time |
|---------|-------------|------|
| `all-quick` | All scanners, limited features (default) | ~2 min |
| `all-complete` | All scanners, all features | ~12 min |
| `code-security` | SAST, secrets, API security, git history security | ~3 min |
| `packages` | SBOM + vulnerability analysis | ~3 min |
| `devops` | IaC, containers, GitHub Actions | ~3 min |

### Scan an Organization

```bash
# Scan all repos in an organization (default limit: 25)
./zero hydrate zero-test-org

# Scan with a specific profile
./zero hydrate zero-test-org all-quick

# Limit to 10 repos
./zero hydrate zero-test-org --limit 10

# Demo mode: skip repos > 50MB, fetch replacements automatically
./zero hydrate zero-test-org --demo
```

**Organization Flags:**
- `--limit N` - Maximum repos to process (default: 25)
- `--demo` - Demo mode: skip repositories larger than 50MB, automatically fetch replacement repos to maintain the requested count

### Enter Agent Mode (Claude Code)

In Claude Code, use the `/agent` slash command to chat with Zero, the AI orchestrator who can delegate to specialist agents.

## Documentation Index

### Architecture

Technical documentation for developers and contributors.

| Document | Description |
|----------|-------------|
| [System Overview](architecture/overview.md) | High-level architecture and component relationships |
| [Scanner Architecture](architecture/scanners.md) | How scanners work: Semgrep engine + RAG rules + wrappers |
| [RAG Pipeline](architecture/rag-pipeline.md) | Converting markdown patterns to Semgrep YAML rules |
| [Knowledge Base](architecture/knowledge-base.md) | Knowledge organization within agents |

### Scanners

Reference documentation for all available scanners.

| Document | Description |
|----------|-------------|
| [Scanner Reference](scanners/reference.md) | Complete list of all scanners with options and examples |
| [Output Formats](scanners/output-formats.md) | JSON schemas for scanner output with examples |

### Agents (Hackers-Themed)

Self-contained, portable AI agents named after characters from Hackers (1995):

| Agent | Character | Expertise | Documentation |
|-------|-----------|-----------|---------------|
| **Zero** | Zero Cool | Orchestrator | [agents/README.md](agents/README.md) |
| **Cereal** | Cereal Killer | Supply chain, CVEs, malware | [agents/supply-chain](../agents/supply-chain/) |
| **Razor** | Razor | Code security, SAST, secrets | [agents/code-security](../agents/code-security/) |
| **Gill** | Gill Bates | Cryptography, TLS, keys | [agents/cryptography](../agents/cryptography/) |
| **Blade** | Blade | Compliance, SOC 2, ISO 27001 | [agents/compliance](../agents/compliance/) |
| **Phreak** | Phantom Phreak | Legal, licenses, privacy | [agents/legal](../agents/legal/) |
| **Acid** | Acid Burn | Frontend, React, TypeScript | [agents/frontend](../agents/frontend/) |
| **Dade** | Dade Murphy | Backend, APIs, databases | [agents/backend](../agents/backend/) |
| **Nikon** | Lord Nikon | Architecture, system design | [agents/architecture](../agents/architecture/) |
| **Joey** | Joey | CI/CD, build optimization | [agents/build](../agents/build/) |
| **Plague** | The Plague | DevOps, Kubernetes, IaC | [agents/devops](../agents/devops/) |
| **Gibson** | The Gibson | DORA metrics, team health | [agents/engineering-leader](../agents/engineering-leader/) |

See [Agent Reference](agents/README.md) for full documentation including invocation examples.

### Integrations

How to integrate Zero with other tools.

| Document | Description |
|----------|-------------|
| [MCP Server](integrations/mcp.md) | Model Context Protocol server for Claude Code |

## Super Scanners (v4.0)

Zero uses 7 consolidated super scanners with configurable features:

| Scanner | Description |
|---------|-------------|
| `code-packages` | SBOM generation + dependency analysis (vulns, licenses, malcontent, health) |
| `code-security` | SAST, secrets, API security, cryptography (ciphers, keys, TLS) |
| `code-quality` | Tech debt markers, complexity, test coverage, documentation |
| `devops` | IaC security, containers, GitHub Actions, DORA metrics |
| `technology-identification` | Technology detection, ML-BOM generation, AI/ML frameworks |
| `code-ownership` | Contributors, bus factor, CODEOWNERS, code churn |
| `developer-experience` | Onboarding friction, tool sprawl, workflow analysis |

## Scan Profiles

| Profile | Scanners | Time |
|---------|----------|------|
| `all-quick` | All 7 scanners (limited features) | ~2 min |
| `all-complete` | All 7 scanners (all features) | ~12 min |
| `code-packages` | SBOM + dependency analysis | ~1 min |
| `code-security` | SAST, secrets, API, cryptography | ~2 min |
| `devops` | IaC, containers, GitHub Actions, DORA | ~3 min |

## Data Storage

Zero stores data in `.zero/` (project local by default, override with `ZERO_HOME`):

```
.zero/
└── repos/
    └── owner/
        └── repo/
            ├── repo/           # Cloned repository
            └── analysis/       # Scanner output
                ├── sbom.cdx.json               # CycloneDX SBOM
                ├── code-packages.json          # Package analysis
                ├── code-security.json          # Security findings
                ├── code-quality.json           # Quality metrics
                ├── devops.json                 # DevOps analysis
                ├── technology-identification.json  # Tech detection
                ├── code-ownership.json         # Ownership analysis
                └── developer-experience.json   # DevX metrics
```

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `GITHUB_TOKEN` | Yes | GitHub API access for cloning |
| `ANTHROPIC_API_KEY` | No | Claude API for AI-enhanced analysis |
| `ZERO_HOME` | No | Override default `.zero/` location |

## RAG Knowledge Base

Zero uses a RAG (Retrieval-Augmented Generation) knowledge base:

- **Technology detection patterns** covering cloud providers, APIs, frameworks, and services
- **Secret detection rules** generated from RAG patterns
- **Markdown source files** converted to Semgrep YAML rules
- **Human-readable patterns** that serve as both documentation and detection rules

Run `zero rag stats` to see current pattern counts.

See [RAG Pipeline](architecture/rag-pipeline.md) for details.

## See Also

- [CLAUDE.md](../CLAUDE.md) - Claude Code configuration
- [ROADMAP.md](../ROADMAP.md) - Project roadmap
