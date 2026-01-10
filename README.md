<!--
SPDX-License-Identifier: GPL-3.0
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
-->

# Zero

> **"Hack the planet!"** - Engineering intelligence platform for repository analysis

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Status: Alpha](https://img.shields.io/badge/Status-Alpha-orange.svg)](https://github.com/crashappsec/zero)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8.svg)](https://go.dev)

Named after **Zero Cool** from Hackers (1995), Zero provides engineering intelligence tools and specialist AI agents for comprehensive repository assessment.

## Quick Start

```bash
# Clone and build
git clone https://github.com/crashappsec/zero.git
cd zero
go build -o zero ./cmd/zero

# Authenticate with GitHub
gh auth login

# Initialize rules from RAG knowledge base
./zero feeds semgrep    # Sync Semgrep community rules
./zero feeds rag        # Generate rules from RAG patterns

# Scan a repository
./zero hydrate strapi/strapi

# View results
./zero serve            # Open http://localhost:3000
```

**[Full Getting Started Guide](docs/GETTING_STARTED.md)** - Prerequisites, installation, profiles, and troubleshooting.

## What is Zero?

Zero is a Go CLI that provides **Engineering Intelligence** through 7 analyzers with 45+ features, aligned with industry frameworks (DORA, SPACE, LinearB).

### 6 Pillars of Engineering Intelligence

```
┌────────────────────────────────────────────────────────────┐
│     Productivity Pillars    │      Technical Pillars       │
├────────────────────────────────────────────────────────────┤
│  Speed │ Quality │  Team   │ Security │ Supply  │  Tech   │
│ (DORA) │ (Health)│(People) │  (Risk)  │  Chain  │ (Stack) │
└────────────────────────────────────────────────────────────┘
```

| Pillar | Analyzer | Key Metrics |
|--------|----------|-------------|
| **Speed** | devops | DORA metrics, cycle time, deployment frequency |
| **Quality** | code-quality | Tech debt, complexity, test coverage |
| **Team** | code-ownership, devx | Bus factor, contributors, onboarding |
| **Security** | code-security | Vulnerabilities, secrets, crypto issues |
| **Supply Chain** | code-packages | Dependencies, licenses, malware, SBOM |
| **Technology** | technology-identification | Stack detection, ML-BOM, AI security |

### AI Agents

12 specialist agents (named after Hackers characters) for deep analysis:

| Agent | Expertise |
|-------|-----------|
| **Zero** | Master orchestrator |
| **Cereal** | Supply chain, malware, CVEs |
| **Razor** | Code security, SAST, secrets |
| **Gill** | Cryptography, TLS, keys |
| **Plague** | DevOps, Kubernetes, IaC |
| **Hal** | AI/ML security, ML-BOM |

Use `/agent` in Claude Code to chat with Zero.

## Scan Profiles

```bash
./zero hydrate owner/repo all-quick       # All scanners, fast (~2min)
./zero hydrate owner/repo all-complete    # All scanners, thorough (~12min)
./zero hydrate owner/repo code-security   # Security only
./zero hydrate myorg --demo               # Organization scan, skip large repos
```

## Documentation

| Document | Description |
|----------|-------------|
| **[Getting Started](docs/GETTING_STARTED.md)** | Installation, prerequisites, first scan |
| **[Documentation Index](docs/README.md)** | Full documentation |
| **[Scanner Reference](docs/scanners/reference.md)** | All scanners and features |
| **[Agent Reference](docs/agents/README.md)** | AI agent system |
| **[Configuration](config/README.md)** | Profiles, settings, customization |

## Commands

```bash
./zero hydrate <target> [profile]  # Clone and scan
./zero status                       # Show analyzed projects
./zero checkup                      # Verify setup and tools
./zero serve                        # Start web UI
./zero feeds rag                    # Generate rules from RAG
./zero feeds semgrep                # Sync Semgrep rules
./zero list                         # List available scanners
```

## Storage

```
.zero/
└── repos/owner/repo/
    ├── repo/              # Cloned repository
    └── analysis/          # Scanner results (JSON)
        ├── sbom.cdx.json
        ├── code-packages.json
        ├── code-security.json
        └── ...
```

## Contributing

Contributions welcome! See [CONTRIBUTING.md](./CONTRIBUTING.md).

All contributors must complete our [Contributor License Agreement](https://crashoverride.com/docs/other/contributing).

## License

[GNU General Public License v3.0](./LICENSE)

```
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
```

---

**Status**: Alpha | **Version**: 6.0.0 | *"Hack the planet!"*
