<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
SPDX-License-Identifier: GPL-3.0
-->

# Zero Documentation

Comprehensive documentation for Zero - an engineering intelligence platform for repository assessment.

## Quick Start

### Prerequisites

- Go 1.21+
- GitHub token (for cloning repositories)

```bash
# Set your GitHub token
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
./zero hydrate phantom-tests/juice-shop

# Scan with a specific profile
./zero hydrate phantom-tests/juice-shop code-security

# Check scan status
./zero status

# View the HTML report
./zero report phantom-tests/juice-shop
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
# Scan all repos in an organization
./zero hydrate phantom-tests

# Limit to 5 repos with code-security profile
./zero hydrate phantom-tests code-security --limit 5
```

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

## Scanner Categories

### Code Scanners

| Scanner | Description |
|---------|-------------|
| `code-security` | SAST vulnerabilities, secret detection, API security, git history security |
| `code-crypto` | Cryptographic security: ciphers, keys, random, TLS, certificates |
| `tech-id` | Technology stack identification, AI/ML detection |
| `code-quality` | TODO, FIXME, complexity markers, test coverage |

### Package Scanners

| Scanner | Description |
|---------|-------------|
| `package-sbom` | CycloneDX SBOM generation (syft/cdxgen) |
| `package-vulns` | CVE detection (OSV database) |
| `package-health` | Abandonment, typosquatting risk |
| `package-malcontent` | Supply chain malware detection |
| `package-provenance` | SLSA provenance verification |
| `package-licenses` | SPDX license compliance (feature of packages scanner) |

### Infrastructure Scanners

| Scanner | Description |
|---------|-------------|
| `iac-security` | Terraform, CloudFormation, K8s (Checkov) |
| `container-security` | Dockerfile, images (Trivy, Hadolint) |

## Scan Profiles

| Profile | Scanners | Time |
|---------|----------|------|
| `all-quick` | All 9 scanners (limited features) | ~2 min |
| `all-complete` | All 9 scanners (all features) | ~12 min |
| `code-crypto` | Cryptographic security | ~2 min |
| `code-security` | SAST, secrets, API, git history security | ~3 min |
| `packages` | sbom, package analysis | ~3 min |

## Data Storage

Zero stores data in `~/.zero/`:

```
~/.zero/
└── repos/
    └── owner/
        └── repo/
            ├── repo/           # Cloned repository
            └── analysis/       # Scanner output
                ├── manifest.json
                ├── package-sbom.json
                ├── package-vulns.json
                ├── code-secrets.json
                ├── code-crypto.json
                └── ...
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
