<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Zero Roadmap

> Last updated: 2026-01-07 | Current version: 4.0.0

## Overview

Zero is an engineering intelligence platform that provides AI-assisted code analysis with 12 specialist agents powered by 7 super scanners.

---

## What's Available Now

### Core Features

| Feature | Status | Description |
|---------|--------|-------------|
| 7 Super Scanners | ✅ Complete | code-packages, code-security, code-quality, devops, technology-identification, code-ownership, devx |
| Report Generation | ✅ Complete | Interactive HTML reports via Evidence.dev |
| Agent Mode | ✅ Complete | `/agent` chat with Zero and 12 specialists |
| Hydrate Command | ✅ Complete | Clone + scan with configurable profiles |
| Freshness Tracking | ✅ Complete | Know when data is stale |
| Automation | ✅ Complete | `zero watch`, `zero refresh` commands |
| RAG Pattern Detection | ✅ Complete | 3500+ detection rules from knowledge base |
| Docker Support | ✅ Complete | Run Zero anywhere with consistent dependencies |

### Scanners

| Scanner | Features |
|---------|----------|
| **code-packages** | SBOM generation, vulnerability scanning, license analysis, malcontent detection, typosquatting, package health |
| **code-security** | SAST, secrets detection, cryptography analysis, API security |
| **code-quality** | Tech debt, complexity, test coverage, documentation |
| **devops** | IaC security, container security, GitHub Actions, DORA metrics |
| **technology-identification** | Tech detection, ML-BOM, AI security, framework detection |
| **code-ownership** | Contributors, bus factor, code owners, orphaned code |
| **devx** | Developer experience, onboarding, workflow analysis |

### AI Agents

| Agent | Specialty |
|-------|-----------|
| **Zero** | Master orchestrator - coordinates all specialists |
| **Cereal** | Supply chain security, malcontent, vulnerabilities |
| **Razor** | Code security, SAST, secrets detection |
| **Gill** | Cryptography, ciphers, keys, TLS |
| **Hal** | AI/ML security, model safety, LLM security |
| **Blade** | Compliance, SOC 2, ISO 27001 |
| **Phreak** | Legal, licenses, data privacy |
| **Plague** | DevOps, infrastructure, Kubernetes |
| **Joey** | CI/CD, build optimization |
| **Nikon** | Architecture, system design |
| **Acid** | Frontend, React, accessibility |
| **Dade** | Backend, APIs, databases |
| **Gibson** | DORA metrics, team health |

---

## In Progress

### Priority 1: Test Coverage

Improving test coverage across all packages to enable confident refactoring and releases.

**Current Progress**:
- Core packages (sarif, errors, feeds): 80%+
- Scanner packages: 20-50%
- Workflow packages: 15-30%

**Target**: 70% coverage across critical packages

### Priority 2: MCP Integration

Enable Zero as an MCP server for IDE integration with Claude Desktop and VS Code.

**Status**: Architecture defined, implementation pending

---

## Planned

### Near-term

| Feature | Description |
|---------|-------------|
| CI/CD Integration | GitHub Actions workflow for automated scanning |
| SARIF Export | Standard security format for CI integration |
| Custom Rules | User-defined Semgrep rules |
| Incremental Scanning | Only scan changed files |

### Long-term

| Feature | Description |
|---------|-------------|
| Web Dashboard | Interactive web UI for results |
| Multi-Repo Analysis | Compare security across repositories |
| Remediation Automation | Auto-fix PRs for common issues |
| Cloud Service | Hosted Zero with API access |
| Enterprise Features | SSO, teams, policies |

---

## Quick Start

```bash
# Scan a repository
zero hydrate owner/repo

# Generate interactive report
zero report owner/repo

# Chat with AI agents
/agent
```

See [GETTING_STARTED.md](GETTING_STARTED.md) for detailed usage.

---

*"Hack the planet!"*
