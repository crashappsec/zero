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
go build -o zero ./cmd/zero

# Check prerequisites and install missing tools
./zero checkup --fix

# Verify your GitHub token and see what scanners will work
./zero checkup
```

### Prerequisites

**Required:**
- Go 1.22+
- Git
- GitHub CLI (`gh`) - for authentication

**Recommended Tools** (install with `./zero checkup --fix`):
| Tool | Purpose | Install |
|------|---------|---------|
| [cdxgen](https://github.com/CycloneDX/cdxgen) | SBOM generation (preferred) | `npm install -g @cyclonedx/cdxgen` |
| [syft](https://github.com/anchore/syft) | SBOM generation (fallback) | `brew install syft` |
| [osv-scanner](https://github.com/google/osv-scanner) | Vulnerability scanning | `go install github.com/google/osv-scanner/cmd/osv-scanner@latest` |
| [semgrep](https://github.com/returntocorp/semgrep) | Code security scanning | `brew install semgrep` |
| [malcontent](https://github.com/chainguard-dev/malcontent) | Supply chain malware detection | `go install github.com/chainguard-dev/malcontent/cmd/mal@latest` |
| [trivy](https://github.com/aquasecurity/trivy) | Container/IaC scanning | `brew install trivy` |
| [checkov](https://github.com/bridgecrewio/checkov) | IaC security | `pip install checkov` |

### Basic Usage

```bash
# Hydrate (clone and scan) a repository
./zero hydrate <owner/repo>

# With analysis profiles (profile is a positional argument)
./zero hydrate <owner/repo> all-quick       # All scanners, limited features (~2min)
./zero hydrate <owner/repo> all-complete    # All scanners, all features (~12min)
./zero hydrate <owner/repo> code-packages   # SBOM + dependency analysis
./zero hydrate <owner/repo> code-security   # Security scanning only

# Scan an entire GitHub organization (no "/" means org)
./zero hydrate <org>                        # All repos in org
./zero hydrate <org> all-quick              # With profile
./zero hydrate <org> --limit 10             # Limit repos

# Check status of analyzed projects
./zero status

# See what scanners work with your token
./zero checkup

# List all available scanners
./zero list
```

## Servers

Zero includes three server components for different use cases:

### API Server

The API server provides a REST API and WebSocket endpoints for real-time scan progress:

```bash
# Start API server (default port 3001)
./zero serve

# Custom port
./zero serve --port 8080

# Development mode (enables CORS for frontend dev server)
./zero serve --dev
```

**API Endpoints:**
- `GET /api/projects` - List all analyzed projects
- `GET /api/projects/:id` - Get project details
- `POST /api/scans` - Start a new scan
- `GET /api/scans/:id` - Get scan status
- `WS /ws` - WebSocket for real-time updates

### Web UI

Zero includes a Next.js web dashboard for visualizing analysis results:

```bash
# Install dependencies (first time only)
cd web && npm install

# Start development server (default port 3000)
npm run dev

# Production build
npm run build && npm start
```

The Web UI connects to the API server, so start both:
```bash
# Terminal 1: Start API server
./zero serve --dev

# Terminal 2: Start Web UI
cd web && npm run dev
```

Then open http://localhost:3000 in your browser.

### MCP Server (Claude Desktop Integration)

Zero provides an MCP (Model Context Protocol) server for Claude Desktop integration:

```bash
# Start MCP server (for testing)
./zero mcp
```

**Claude Desktop Configuration:**

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "zero": {
      "command": "/path/to/zero",
      "args": ["mcp"]
    }
  }
}
```

This enables Claude Desktop to query your analysis data directly.

## Commands

| Command | Description |
|---------|-------------|
| `hydrate <target> [profile]` | Clone and scan (target: `owner/repo` or `org-name`) |
| `scan <target> [profile]` | Re-scan already-cloned repos |
| `status` | Show all analyzed projects |
| `serve` | Start the API server |
| `checkup` | Check setup, token permissions, and install missing tools |
| `list` | List all available scanners |
| `mcp` | Start MCP server for Claude Desktop |
| `clean <owner/repo>` | Remove analysis data |
| `history <owner/repo>` | Show scan history |

**Target detection:** If target contains `/`, it's a single repo. Otherwise, it's an organization.

Profiles are defined in `config/zero.config.json` and can be customized.

## Scanners

Zero uses **7 consolidated super scanners** (v4.0 architecture), each with multiple configurable features:

| Scanner | Features | Description | External Tools |
|---------|----------|-------------|----------------|
| **code-packages** | generation, integrity, vulns, health, licenses, malcontent, confusion, typosquats, deprecations, duplicates, reachability, provenance, bundle, recommendations | SBOM generation + package/dependency analysis | cdxgen, syft, osv-scanner, malcontent |
| **code-security** | vulns, secrets, api, ciphers, keys, random, tls, certificates | Code analysis + cryptographic security | semgrep |
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

### Configuration Files

Configuration is loaded from multiple sources (later overrides earlier):

1. `config/defaults/scanners.json` - Scanner feature defaults
2. `config/zero.config.json` - Main config with settings and profiles
3. `~/.zero/config.json` - User overrides (optional)

**Main config (`config/zero.config.json`):**
```json
{
  "settings": {
    "default_profile": "all-quick",
    "parallel_repos": 8,
    "parallel_scanners": 4,
    "scanner_timeout_seconds": 300
  },
  "profiles": {
    "all-quick": {
      "name": "All Quick",
      "description": "All scanners with fast defaults",
      "scanners": ["code-packages", "code-security", "code-quality", "devops", "technology-identification", "code-ownership", "developer-experience"]
    }
  }
}
```

**User overrides (`~/.zero/config.json`):**
```json
{
  "profiles": {
    "my-custom": {
      "name": "My Custom Profile",
      "scanners": ["code-security", "code-packages"],
      "feature_overrides": {
        "code-security": {
          "secrets": { "git_history_scan": { "enabled": true } }
        }
      }
    }
  }
}
```

See `config/README.md` for full documentation.

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
| `developer-experience` | technology-identification, developer-experience | Developer experience |

## Checkup Command

The `checkup` command helps you understand what scanners will work with your current setup:

```bash
./zero checkup
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
go build -o zero ./cmd/zero

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
./zero hydrate phantom-tests/express
./zero hydrate phantom-tests/platform
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

**Status**: Alpha (Web UI: Experimental)
**Version**: 4.1.0 (Super Scanner Architecture)
**Last Updated**: 2026-01-05

*"Hack the planet!"*
