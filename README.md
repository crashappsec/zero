<!--
SPDX-License-Identifier: GPL-3.0
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
-->

# Zero

> **"Hack the planet!"** - A unified orchestrator for repository analysis, security scanning, and developer productivity insights powered by AI agents

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Status: Experimental](https://img.shields.io/badge/Status-Experimental-orange.svg)](https://github.com/crashappsec/zero)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

Named after **Zero Cool** from the movie Hackers (1995), Zero is a team of AI agents that analyze your code for security, compliance, and quality issues.

## What is Zero?

Zero is a set of open-source tools for software and security engineers. When combined with two other Crash Override projects—[Chalk](https://github.com/crashappsec/chalk) and Ocular—it can analyze code and builds to create a comprehensive understanding of what is happening in the development process.

[Crash Override](https://crashoverride.com) sells a commercial platform that includes advanced versions of these tools and other features for operations teams, providing a way to understand and improve software development at scale.

### Components

**Source Code Management**
- Local cloning utilities designed for laptop-based analysis
- The Ocular project handles code syncing at scale for enterprise environments

**Scanning Utilities**
- Standalone scripts and wrappers for security and analysis tools
- Many are unique solutions for specific data collection and complex business problems
- Zero orchestrates these scanners locally; Ocular handles orchestration at scale

**Data & Intelligence**
- Results stored in JSON format for easy consumption
- Comprehensive RAG (Retrieval-Augmented Generation) knowledge base
- Reference data for agents to perform specialized analysis tasks

**AI Agents**
- A team of specialist agents (named after Hackers characters)
- Each agent has deep expertise in their domain
- Zero coordinates investigations, delegates to specialists, and synthesizes findings into actionable insights

### The Team

| Agent | Character | Expertise |
|-------|-----------|-----------|
| **Zero** | Zero Cool | Master orchestrator - coordinates all agents |
| **Cereal** | Cereal Killer | Supply chain security, malware detection, CVEs |
| **Razor** | Razor | Code security, SAST, secrets detection |
| **Blade** | Blade | Compliance, SOC 2, ISO 27001 auditing |
| **Phreak** | Phantom Phreak | Legal, licenses, data privacy |
| **Acid** | Acid Burn | Frontend, React, TypeScript, accessibility |
| **Flu Shot** | Flu Shot | Backend, APIs, databases |
| **Nikon** | Lord Nikon | Architecture, system design |
| **Joey** | Joey | Build, CI/CD, performance |
| **Plague** | The Plague | DevOps, infrastructure, Kubernetes |
| **Gibson** | The Gibson | Engineering metrics, DORA |

## Features

### Supply Chain Security
- **Vulnerability Scanning** - CVE detection via OSV database with KEV (Known Exploited Vulnerabilities) prioritization
- **Malware Detection** - Supply chain compromise detection using [malcontent](https://github.com/chainguard-dev/malcontent) with 14,500+ YARA rules
- **Package Health** - Dependency health scoring, abandonment detection, typosquat warnings
- **Package Provenance** - SLSA attestation verification, build provenance analysis

### SBOM Generation
- **Default: cdxgen** - [cdxgen](https://github.com/CycloneDX/cdxgen) installs dependencies for complete transitive dependency analysis
- **Optional: Syft** - [Syft](https://github.com/anchore/syft) available via `--generator syft` for fast static analysis
- **CycloneDX Format** - Industry-standard SBOM with CPE identifiers for vulnerability matching

### Code Security
- **Static Analysis** - SAST findings via Semgrep with custom rules
- **Secrets Detection** - API keys, credentials, and token detection
- **IaC Security** - Terraform, Kubernetes, CloudFormation, and Dockerfile analysis

### Compliance & Legal
- **License Analysis** - SPDX license detection with policy enforcement (allowed/denied/review)
- **Content Policy** - Profanity and non-inclusive language detection

### Developer Productivity
- **Technology Detection** - Automated tech stack identification across 100+ frameworks and languages
- **Code Ownership** - Bus factor analysis, CODEOWNERS validation, contributor insights
- **DORA Metrics** - Deployment frequency, lead time, change failure rate, MTTR

## Quick Start

### Prerequisites

**Required:**
- Bash 3.2+ (macOS default works)
- Git, jq, curl
- [syft](https://github.com/anchore/syft) - Fast SBOM generation
- [osv-scanner](https://github.com/google/osv-scanner) - Vulnerability scanning
- [gh](https://cli.github.com/) - GitHub CLI

**Recommended:**
- [cdxgen](https://github.com/CycloneDX/cdxgen) - Deep SBOM generation (installs deps for complete analysis)
- [malcontent](https://github.com/chainguard-dev/malcontent) - Supply chain compromise detection
- [semgrep](https://github.com/returntocorp/semgrep) - Code security scanning
- [trivy](https://github.com/aquasecurity/trivy) - Container vulnerability scanning
- [hadolint](https://github.com/hadolint/hadolint) - Dockerfile linting
- [checkov](https://github.com/bridgecrewio/checkov) - IaC security scanning

### Installation

```bash
# Clone the repository
git clone https://github.com/crashappsec/zero.git
cd zero

# Check prerequisites (will offer to install missing tools)
./zero.sh check --fix

# Set up API keys
cp .env.example .env
# Edit .env with your GITHUB_TOKEN and ANTHROPIC_API_KEY
```

### Usage

**Interactive Mode:**
```bash
./zero.sh
```

**Command Line:**
```bash
# Hydrate (clone and analyze) a repository
./zero.sh hydrate expressjs/express

# With analysis profiles
./zero.sh hydrate owner/repo --quick      # ~30s - fast scan
./zero.sh hydrate owner/repo --standard   # ~2min - default
./zero.sh hydrate owner/repo --security   # ~3min - security-focused
./zero.sh hydrate owner/repo --advanced   # ~5min - all analyzers
./zero.sh hydrate owner/repo --deep       # ~10min - Claude-assisted

# Check status of hydrated projects
./zero.sh status

# Generate reports
./zero.sh report expressjs/express
```

**Agent Mode (Claude Code):**
```
/agent
```

This enters agent mode where you chat with Zero, who can delegate to specialist agents for deep analysis.

## Agent Mode

The real power of Zero is the agent system. Use the `/agent` slash command in Claude Code to chat with Zero:

```
You: Do we have any malware in our dependencies?

Zero: Let me check what projects are loaded and delegate to Cereal...
[Invokes Cereal agent to investigate malcontent findings]

Cereal: I've analyzed the malcontent scan results for express. Found 3 high-risk
behaviors flagged, but after reading the source files, all appear to be false
positives related to legitimate test fixtures...
```

### Example Queries

| Query | Agent | What Happens |
|-------|-------|--------------|
| "Any malware in our deps?" | Cereal | Investigates malcontent scanner findings |
| "Review code security" | Razor | Analyzes SAST findings and secrets |
| "Are we SOC 2 compliant?" | Blade | Assesses compliance posture |
| "License conflicts?" | Phreak | Reviews license compatibility |
| "Frontend architecture?" | Acid | Reviews React/TypeScript patterns |
| "API security issues?" | Flu Shot | Checks backend security |
| "System design concerns?" | Nikon | Architecture review |
| "CI/CD performance?" | Joey | Build pipeline analysis |
| "Infrastructure issues?" | Plague | DevOps/K8s review |
| "Team health metrics?" | Gibson | DORA metrics analysis |

## Storage

All analysis data is stored in `~/.zero/`:

```
~/.zero/
├── config.json                 # Global settings
├── index.json                  # Project index
└── repos/
    └── expressjs/
        └── express/
            ├── project.json    # Project metadata
            ├── repo/           # Cloned repository
            └── analysis/       # Analysis results
                ├── manifest.json
                ├── scanners/
                │   ├── vulnerabilities/
                │   ├── package-malcontent/
                │   ├── package-health/
                │   ├── licenses/
                │   ├── code-security/
                │   └── secrets-scanner/
                └── technology.json
```

## Configuration

### Environment Variables

```bash
# GitHub authentication (required for GitHub API)
export GITHUB_TOKEN="ghp_..."

# Claude API key (for deep analysis and agents)
export ANTHROPIC_API_KEY="sk-ant-..."
```

### Analysis Profiles

All profiles use cdxgen by default. Use `--generator syft` for faster but less complete SBOM generation.

| Profile | Time | Scanners |
|---------|------|----------|
| **quick** | ~30s | SBOM, tech-discovery, vulnerabilities, licenses |
| **standard** | ~2min | + package-health, code-secrets, code-ownership |
| **security** | ~3min | SBOM, vulns, code-security, iac-security, secrets, malcontent, container-security |
| **packages** | ~5min | SBOM, vulns, health, provenance, malcontent, bundle-optimization, recommendations |
| **advanced** | ~5min | All scanners except bundle-analysis |
| **deep** | ~10min | All scanners + Claude AI-assisted insights |

## Scanners

Zero includes 21 specialized scanners organized by category:

### Supply Chain Security
| Scanner | Description |
|---------|-------------|
| `package-sbom` | CycloneDX SBOM generation (syft or cdxgen) |
| `package-vulns` | CVE scanning via OSV with KEV prioritization |
| `package-health` | Dependency health scoring, abandonment, typosquats |
| `package-provenance` | SLSA attestations and build provenance |
| `package-malcontent` | Malware detection with 14,500+ YARA rules |
| `package-recommendations` | Alternative library suggestions |
| `package-bundle-optimization` | JavaScript bundle size analysis |

### Code Security
| Scanner | Description |
|---------|-------------|
| `code-security` | SAST analysis via Semgrep |
| `code-secrets` | API keys, credentials, token detection |
| `iac-security` | Terraform, K8s, CloudFormation analysis |
| `container-security` | Dockerfile best practices and hardening |

### Compliance & Legal
| Scanner | Description |
|---------|-------------|
| `licenses` | SPDX license detection with policy enforcement |
| `digital-certificates` | SSL/TLS certificate analysis |

### Developer Productivity
| Scanner | Description |
|---------|-------------|
| `tech-discovery` | Framework and language detection (100+ patterns) |
| `code-ownership` | CODEOWNERS, bus factor, contributor analysis |
| `dora` | Deployment frequency, lead time, MTTR |
| `git` | Commit patterns, contributor activity |
| `documentation` | README quality and docs coverage |
| `test-coverage` | Test framework detection and coverage estimates |
| `tech-debt` | Code duplication, complexity, TODO markers |

### Container & Infrastructure
| Scanner | Description |
|---------|-------------|
| `containers` | Container image analysis |
| `bundle-analysis` | Deep npm bundle analysis with tree-shaking |

## Repository Structure

```
zero/
├── zero.sh                     # Main CLI entry point
├── utils/
│   ├── zero/                   # Zero orchestrator
│   │   ├── lib/                # Libraries (zero-lib.sh, agent-loader.sh)
│   │   ├── scripts/            # CLI scripts (hydrate, scan, report)
│   │   └── config/             # Configuration files
│   └── scanners/               # Individual scanners
│       ├── vulnerabilities/
│       ├── package-malcontent/
│       ├── package-health/
│       ├── tech-discovery/
│       └── ...
├── agents/                     # Specialist agent definitions
│   ├── orchestrator/           # Zero - master orchestrator
│   ├── supply-chain/           # Cereal Killer - supply chain security
│   ├── code-security/          # Razor - code security
│   ├── compliance/             # Blade - compliance auditing
│   ├── legal/                  # Phantom Phreak - legal counsel
│   ├── frontend/               # Acid Burn - frontend engineer
│   ├── backend/                # Flu Shot - backend engineer
│   ├── architecture/           # Lord Nikon - software architect
│   ├── build/                  # Joey - build engineer
│   ├── devops/                 # The Plague - devops engineer
│   ├── engineering-leader/     # The Gibson - engineering metrics
│   └── shared/                 # Cross-agent knowledge
├── rag/                        # RAG knowledge base
│   ├── technology-identification/  # 100+ technology patterns
│   ├── supply-chain/           # Supply chain references
│   └── ...
└── .claude/
    └── commands/               # Slash commands (/agent, /zero)
```

## Claude Code Integration

Zero is designed to work with [Claude Code](https://claude.com/claude-code). Install the Claude Code extension and use:

- `/agent` - Chat with Zero and the specialist agents
- `/zero` - Access Zero commands and documentation

## Test Organization

We've created the [phantom-tests](https://github.com/phantom-tests) organization with sample repositories for safe testing:

```bash
./zero.sh hydrate phantom-tests/express
./zero.sh hydrate phantom-tests/mitmproxy
./zero.sh hydrate phantom-tests/openai-node
```

## Roadmap

See [ROADMAP.md](./ROADMAP.md) for planned features including:

- **Cloud Asset Inventory** - Multi-cloud discovery and SBOM generation for AWS, Azure, GCP
- **API Security Analysis** - OpenAPI/Swagger and GraphQL scanning
- **Reachability Analysis** - Determine if vulnerable dependencies are actually used
- **Report System** - HTML/PDF reports with trend analysis
- **Integration** - Ocular code sync, Chalk attestations, GitHub/GitLab org analysis

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

**Important:** All contributors must complete our [Contributor License Agreement](https://crashoverride.com/docs/other/contributing) before their contributions can be merged. This ensures the project remains properly licensed and protects both contributors and users.

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

The agents, knowledge base, and RAG database were built to augment the Crash Override platform.

---

**Status**: Experimental Preview
**Version**: 5.1.0
**Last Updated**: 2025-12-12

*"Hack the planet!"*
