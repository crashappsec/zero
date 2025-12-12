<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
SPDX-License-Identifier: GPL-3.0
-->

# Zero Documentation

Comprehensive documentation for the Zero security analysis toolkit.

## Quick Start

```bash
# Clone and analyze a repository
./zero.sh hydrate expressjs/express --security

# Check status
./zero.sh status

# Enter agent mode
/agent
```

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
| `code-vulns` | SAST vulnerabilities (Semgrep p/security-audit, p/owasp-top-ten) |
| `code-secrets` | Secret detection (242+ RAG patterns + p/secrets) |
| `tech-discovery` | Technology stack identification |
| `tech-debt` | TODO, FIXME, complexity markers |

### Cryptography Scanners

| Scanner | Description |
|---------|-------------|
| `crypto-ciphers` | Weak/deprecated cipher detection (DES, RC4, MD5) |
| `crypto-keys` | Hardcoded keys, weak key lengths |
| `crypto-random` | Insecure random number generation |
| `crypto-tls` | TLS/SSL misconfiguration |

### Package Scanners

| Scanner | Description |
|---------|-------------|
| `package-sbom` | CycloneDX SBOM generation (syft/cdxgen) |
| `package-vulns` | CVE detection (OSV database) |
| `package-health` | Abandonment, typosquatting risk |
| `package-malcontent` | Supply chain malware detection |
| `package-provenance` | SLSA provenance verification |
| `licenses` | SPDX license compliance |

### Infrastructure Scanners

| Scanner | Description |
|---------|-------------|
| `iac-security` | Terraform, CloudFormation, K8s (Checkov) |
| `container-security` | Dockerfile, images (Trivy, Hadolint) |

## Scan Profiles

| Profile | Scanners | Time |
|---------|----------|------|
| `quick` | tech-discovery, package-sbom, package-vulns, licenses | ~2 min |
| `standard` | quick + package-health, code-secrets, code-vulns | ~5 min |
| `security` | standard + package-malcontent, package-provenance, iac, container | ~10 min |
| `deep` | All scanners | ~15 min |
| `crypto` | crypto-ciphers, crypto-keys, crypto-random, crypto-tls, code-secrets | ~5 min |
| `packages` | package-sbom, package-vulns, package-health, package-malcontent, licenses | ~8 min |

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
                ├── crypto-ciphers.json
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

- **106 technology patterns** covering AWS, Azure, GCP, Stripe, OpenAI, and 100+ more
- **242+ secret detection rules** generated from patterns
- **Markdown source files** converted to Semgrep YAML rules
- **Human-readable patterns** that serve as both documentation and detection rules

See [RAG Pipeline](architecture/rag-pipeline.md) for details.

## See Also

- [CLAUDE.md](../CLAUDE.md) - Claude Code configuration
- [ROADMAP.md](../ROADMAP.md) - Project roadmap
