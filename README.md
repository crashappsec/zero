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

Phantom is a set of open-source tools for software and security engineers. It's core is a set of AI agents, the knowledge they use and a RAG database with authoritatve information they can use when reasoning. It also consists of a set of shell-script utilities that can be run stand-alone or using Claude, leveraging all of some of the capabilities in this repo. 

The agents, knowledge and RAG was built to augment the Crash Override platform at crashoverride.com. Crash Override works by connecting source-code and the cooud through deep build inspection so it has a single source of truth about everything that is happening in an SDLC. Phantom just operates on the source-code and, when used on it's own has no knowledge of the deep build inspection data or the cloud, however it can be used in conjunction with two other Crash Override projects, Chalk at chalkproject.io which is a core part if deep build inspection and Ocular ocularproject.io that orchestrates code syncing and tools orchestration at scale. 

Phantoms utilities are also designed to provide richer meta-data about source code that can be used by the AI agents and as stand-alone functionality, that is beyond currently available tools. For instance SCA or supply chain security tools match vulnerabilites to open-source libraries, Phantom extends that to provide health score about libraries, can make recomendations of alternative healthier libraries and now developers can reduce their use of packages to improve build times and reduce risk.

## Features

Features are being rapidly developed but currently include:

- **Code Sync** - Clone and sync source code from Github to analyze locally and access to the Github API for further data
- **SBOM** - Software Bill of Materials generation and analysis
- **Package Vulnerabilities** - CVE detection via OSV.dev integration
- **Package Health** - OSSF scorecards and reliability
- **Technology Identification** - Automated tech stack detection with RAG-powered patterns
- **License Compliance** - SPDX license analysis and compliance checking
- **Code Security** - Static analysis for security issues and secrets
- **IaC Security** - Analysis of infrastretcure as code
- **Secret Detection** - Discover secrets in code
- **Code Ownership** - Bus factor analysis and contributor insights
- **DORA Metrics** - Deployment frequency, lead time, and performance metrics
- **Provenance** - Git signature verification and supply chain integrity
- **Code Quality** - including test coverage and documentation 

Many features have two modes, a basic mode that can operate independently and an advanced mode that levaerages Claude. 

You can also use the Agents, providing a full LLM experience to ask quetions about and analyze all of the data. 

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

Phantom provides an interactive terminal menu that can check for pre-requsities, install thewm automatically if asked and run the core utilities. Any individual task can be run stand alone and are availbale in the /utils folder. 

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

## Report Generation

Generate comprehensive reports in multiple formats:

```bash
# Interactive mode - select project and options
./utils/phantom/report.sh -i

# Command line - specify project and options
./utils/phantom/report.sh expressjs/express                    # Summary report (terminal)
./utils/phantom/report.sh expressjs/express -t security        # Security report
./utils/phantom/report.sh expressjs/express -t security -f html -o report.html

# Report types: summary, security, licenses, sbom, compliance, supply-chain, dora, full
# Output formats: terminal, markdown, json, html, csv
```

| Report Type | Description |
|-------------|-------------|
| **summary** | High-level overview of all findings |
| **security** | Vulnerabilities, secrets, code security issues |
| **licenses** | License compliance and dependency licenses |
| **sbom** | Software Bill of Materials |
| **compliance** | Documentation, ownership, policy compliance |
| **supply-chain** | Dependencies, health, provenance |
| **dora** | DevOps metrics and performance |
| **full** | Comprehensive report (all sections) |

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
**Version**: 4.1.0
**Last Updated**: 2025-12-03
