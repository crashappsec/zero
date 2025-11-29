<!--
SPDX-License-Identifier: GPL-3.0
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
-->

# Phantom

> **Experimental Preview** - A unified orchestrator for repository analysis, security scanning, and developer productivity insights powered by AI

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Status: Experimental](https://img.shields.io/badge/Status-Experimental-orange.svg)](https://github.com/crashappsec/gibson-powers)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)
[![GitHub Discussions](https://img.shields.io/badge/GitHub-Discussions-181717?logo=github)](https://github.com/crashappsec/gibson-powers/discussions)

## What is Phantom?

Phantom is a modern Developer-First Platform Intelligence (DPI) system built on AI agents, skills, RAG knowledge bases, and specialized prompts. It serves as a master orchestrator for repository analysis - bootstrapping projects, running analyzers, and querying results through specialist agents. It includes utilities to extract structured data from source code that agents use for comprehensive analysis of security, compliance, code ownership, and development practices.

**Note:** Phantom's open-source agents operate on source code alone, but deliver significantly more sophisticated insights when combined with [Crash Override's](https://crashoverride.com) platform - particularly its deep build inspection technology and cloud services integration - providing enterprise-grade DPI for modern DevOps engineering teams.

```
┌─────────────────────────────────────────────────────────────────────┐
│  PHANTOM                                                             │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│                                                                      │
│  ./phantom.sh hydrate expressjs/express                              │
│                                                                      │
│  [1/7] dependencies     ✓  1s   21 packages (CycloneDX)             │
│  [2/7] technology       ✓  6s   2 technologies                      │
│  [3/7] vulnerabilities  ✓  1s   clean                               │
│  [4/7] licenses         ✓  2s   pass                                │
│  [5/7] security         ✓  1s   2 issues                            │
│  [6/7] ownership        ✓  3s   20 contributors, bus factor 2       │
│  [7/7] dora             ✓  0s   ELITE                               │
│                                                                      │
│  Risk Level: LOW                                                     │
│  Storage: ~/.phantom/projects/expressjs/express/ (11 MB)             │
└─────────────────────────────────────────────────────────────────────┘
```

## Features

- **SBOM Generation** - Software Bill of Materials in CycloneDX format
- **Vulnerability Scanning** - CVE detection via OSV.dev integration
- **Technology Identification** - Automated tech stack detection with RAG-powered patterns
- **License Compliance** - SPDX license analysis and compliance checking
- **Code Security** - Static analysis for security issues and secrets
- **Code Ownership** - Bus factor analysis and contributor insights
- **DORA Metrics** - Deployment frequency, lead time, and performance metrics
- **Package Health** - Abandoned package and typosquatting detection
- **Provenance** - Git signature verification and supply chain integrity

## Quick Start

### Prerequisites

- Bash 4.0+
- Git
- jq
- curl
- [syft](https://github.com/anchore/syft) (recommended for SBOM generation)
- [osv-scanner](https://github.com/google/osv-scanner) (recommended for vulnerability scanning)
- [gh](https://cli.github.com/) (recommended for GitHub integration)

### Installation

```bash
# Clone the repository
git clone https://github.com/crashappsec/gibson-powers.git
cd gibson-powers

# Run setup to install dependencies and configure
./utils/phantom/phantom.sh setup

# Verify everything is ready
./utils/phantom/phantom.sh check
```

### Usage

**Interactive Mode:**
```bash
./utils/phantom/phantom.sh
```

**Command Line:**
```bash
# Hydrate (clone and analyze) a repository
./utils/phantom/phantom.sh hydrate expressjs/express

# With analysis depth options
./utils/phantom/phantom.sh hydrate owner/repo --quick      # ~30s - fast scan
./utils/phantom/phantom.sh hydrate owner/repo --standard   # ~2min - default
./utils/phantom/phantom.sh hydrate owner/repo --advanced   # ~5min - all analyzers
./utils/phantom/phantom.sh hydrate owner/repo --deep       # ~10min - Claude-assisted
./utils/phantom/phantom.sh hydrate owner/repo --security   # Security-focused

# Check status of hydrated projects
./utils/phantom/phantom.sh status

# Clean all analysis data
./utils/phantom/phantom.sh clean
```

### Analysis Modes

| Mode | Time | Analyzers | Description |
|------|------|-----------|-------------|
| **Quick** | ~30s | 4 | Dependencies, technology, vulnerabilities, licenses |
| **Standard** | ~2min | 7 | + security, ownership, DORA metrics |
| **Advanced** | ~5min | 9 | + package health, provenance |
| **Deep** | ~10min | 9 | Claude-assisted analysis (requires API key) |
| **Security** | ~3min | 5 | Security-focused subset |

## Storage

All analysis data is stored in `~/.phantom/`:

```
~/.phantom/
├── config.json                 # Global settings
├── index.json                  # Project index
└── projects/
    └── expressjs/
        └── express/
            ├── project.json    # Project metadata
            ├── repo/           # Cloned repository
            └── analysis/       # Analysis results
                ├── manifest.json
                ├── sbom.cdx.json
                ├── dependencies.json
                ├── technology.json
                ├── vulnerabilities.json
                ├── licenses.json
                ├── security-findings.json
                ├── ownership.json
                └── dora.json
```

## Configuration

### Environment Variables

```bash
# GitHub authentication (for private repos)
export GITHUB_TOKEN="ghp_..."

# Claude API key (for deep analysis mode)
export ANTHROPIC_API_KEY="sk-ant-..."
```

Create a `.env` file in the repository root:

```bash
cp .env.example .env
# Edit .env with your API keys
```

## Available Analyzers

| Analyzer | Output | Description |
|----------|--------|-------------|
| `dependencies` | `sbom.cdx.json`, `dependencies.json` | SBOM generation via syft |
| `technology` | `technology.json` | Tech stack identification |
| `vulnerabilities` | `vulnerabilities.json` | CVE scanning via OSV |
| `licenses` | `licenses.json` | SPDX license analysis |
| `security-findings` | `security-findings.json` | Code security analysis |
| `ownership` | `ownership.json` | Contributor and bus factor analysis |
| `dora` | `dora.json` | DORA metrics calculation |
| `package-health` | `package-health.json` | Dependency health checks |
| `provenance` | `provenance.json` | Git signature verification |

## Standalone Utilities

Each analyzer is also available as a standalone utility:

```bash
# Technology identification
./utils/technology-identification/technology-identification-analyser.sh --repo owner/repo

# Vulnerability analysis
./utils/supply-chain/vulnerability-analysis/vulnerability-analyser.sh --repo owner/repo

# License compliance
./utils/legal-review/legal-analyser.sh --repo owner/repo

# Code ownership
./utils/code-ownership/ownership-analyser.sh /path/to/repo

# DORA metrics
./utils/dora-metrics/dora-analyser.sh /path/to/repo

# Certificate analysis
./utils/certificate-analyser/cert-analyser.sh api.example.com
```

## Claude Integration

For AI-enhanced analysis, set your Anthropic API key and use `--deep` mode:

```bash
export ANTHROPIC_API_KEY="sk-ant-..."
./utils/phantom/phantom.sh hydrate owner/repo --deep
```

Or use individual analyzers with `--claude`:

```bash
./utils/legal-review/legal-analyser.sh --claude --repo owner/repo
./utils/code-ownership/ownership-analyser.sh --claude /path/to/repo
```

## Test Organization

We've created the [phantom-tests](https://github.com/phantom-tests) organization with sample repositories for safe testing:

```bash
./utils/phantom/phantom.sh hydrate phantom-tests/express
./utils/phantom/phantom.sh hydrate phantom-tests/mitmproxy
```

## Repository Structure

```
phantom/
├── utils/
│   ├── phantom/                    # Main orchestrator
│   │   ├── phantom.sh              # Interactive CLI
│   │   ├── bootstrap.sh            # Single repo hydration
│   │   ├── hydrate.sh              # Batch/org hydration
│   │   └── lib/                    # Shared libraries
│   ├── supply-chain/               # Supply chain analysis
│   │   ├── vulnerability-analysis/
│   │   ├── provenance-analysis/
│   │   └── package-health-analysis/
│   ├── technology-identification/  # Tech stack detection
│   ├── legal-review/               # License compliance
│   ├── code-ownership/             # Ownership analysis
│   ├── dora-metrics/               # DORA metrics
│   ├── code-security/              # Security analysis
│   └── certificate-analyser/       # TLS/cert analysis
├── rag/                            # RAG knowledge base
│   ├── technology-identification/  # 112+ technology patterns
│   ├── supply-chain/               # Supply chain references
│   ├── legal-review/               # Legal compliance
│   └── ...
├── prompts/                        # Reusable prompt templates
├── skills/                         # Claude Code skills
├── agents/                         # Specialist agent definitions
└── docs/                           # Documentation
```

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

```bash
# Run validation
./utils/validation/check-copyright.sh

# Run tests
./utils/code-ownership/tests/run-all-tests.sh
```

## License

Phantom is licensed under the [GNU General Public License v3.0](./LICENSE).

```
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
```

## About

Phantom is maintained by the open source community and sponsored by [Crash Override](https://crashoverride.com).

---

**Status**: Experimental Preview
**Version**: 4.0.0
**Last Updated**: 2025-11-27
