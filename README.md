<!--
SPDX-License-Identifier: GPL-3.0
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
-->

# Zero

> **"Hack the planet!"** - A unified orchestrator for repository analysis, security scanning, and developer productivity insights powered by AI agents

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Status: Experimental](https://img.shields.io/badge/Status-Experimental-orange.svg)](https://github.com/crashappsec/zero)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8.svg)](https://go.dev)

Named after **Zero Cool** from the movie Hackers (1995), Zero is a team of AI agents that analyze your code for security, compliance, and quality issues.

## What is Zero?

Zero is a Go-based CLI tool for software and security engineers. It provides 25+ security scanners, AI-powered analysis agents, and integrates with tools like cdxgen, syft, semgrep, and grype to provide comprehensive security assessments.

### Key Capabilities

- **25+ Security Scanners** - SBOM generation, vulnerability scanning, secrets detection, SAST, IaC security, and more
- **AI Agent System** - Specialist agents (named after Hackers characters) for deep security analysis
- **Configurable** - JSON configuration for scanner options, profiles, and tool preferences
- **Token-Aware** - The `roadmap` command shows which scanners work with your GitHub token permissions

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/crashappsec/zero.git
cd zero

# Build the CLI
go build -o main ./cmd/zero

# Check prerequisites and install missing tools
./main checkup --fix

# Verify your GitHub token and see what scanners will work
./main checkup
```

### Prerequisites

**Required:**
- Go 1.22+
- Git
- GitHub CLI (`gh`) - for authentication

**Recommended Tools** (install with `./main checkup --fix`):
| Tool | Purpose | Install |
|------|---------|---------|
| [cdxgen](https://github.com/CycloneDX/cdxgen) | SBOM generation (preferred) | `npm install -g @cyclonedx/cdxgen` |
| [syft](https://github.com/anchore/syft) | SBOM generation (fallback) | `brew install syft` |
| [grype](https://github.com/anchore/grype) | Vulnerability scanning | `brew install grype` |
| [osv-scanner](https://github.com/google/osv-scanner) | Vulnerability scanning | `go install github.com/google/osv-scanner/cmd/osv-scanner@latest` |
| [semgrep](https://github.com/returntocorp/semgrep) | Code security scanning | `brew install semgrep` |
| [gitleaks](https://github.com/gitleaks/gitleaks) | Secrets detection | `brew install gitleaks` |
| [malcontent](https://github.com/chainguard-dev/malcontent) | Supply chain malware detection | `go install github.com/chainguard-dev/malcontent/cmd/mal@latest` |
| [trivy](https://github.com/aquasecurity/trivy) | Container scanning | `brew install trivy` |
| [checkov](https://github.com/bridgecrewio/checkov) | IaC security | `pip install checkov` |

### Basic Usage

```bash
# Hydrate (clone and scan) a repository
./main hydrate expressjs/express

# With analysis profiles (profile is a positional argument)
./main hydrate expressjs/express quick        # Fast scan (~30s)
./main hydrate expressjs/express security     # Security-focused (~3min)
./main hydrate expressjs/express packages     # Package analysis (~5min)

# Scan an entire GitHub organization (no "/" means org)
./main hydrate phantom-tests                  # All repos in org
./main hydrate phantom-tests quick            # With profile
./main hydrate phantom-tests --limit 10       # Limit repos

# Check status of analyzed projects
./main status

# Generate reports
./main report expressjs/express

# See what scanners work with your token
./main checkup

# List all available scanners
./main list
```

## Commands

| Command | Description |
|---------|-------------|
| `hydrate <target> [profile]` | Clone and scan (target: `owner/repo` or `org-name`) |
| `scan <target> [profile]` | Re-scan already-cloned repos |
| `status` | Show all analyzed projects |
| `report <owner/repo>` | Generate security report |
| `checkup` | Check setup, token permissions, and install missing tools |
| `list` | List all available scanners |
| `clean <owner/repo>` | Remove analysis data |
| `history <owner/repo>` | Show scan history |

**Target detection:** If target contains `/`, it's a single repo. Otherwise, it's an organization.

Profiles are defined in `config/zero.config.json` and can be customized.

## Scanners

Zero includes 25 specialized scanners organized by category:

### Supply Chain Security
| Scanner | Description | External Tool |
|---------|-------------|---------------|
| `package-sbom` | CycloneDX SBOM generation | cdxgen or syft |
| `package-vulns` | CVE scanning via OSV database | osv-scanner or grype |
| `package-health` | Dependency health scoring, abandonment detection | - |
| `package-provenance` | SLSA attestations and build provenance | - |
| `package-malcontent` | Malware detection (14,500+ YARA rules) | malcontent |
| `package-recommendations` | Alternative library suggestions | - |
| `package-bundle-optimization` | JavaScript bundle size analysis | - |

### Code Security
| Scanner | Description | External Tool |
|---------|-------------|---------------|
| `code-vulns` | SAST analysis (OWASP, CWE) | semgrep |
| `code-secrets` | API keys, credentials, token detection | semgrep or gitleaks |
| `api-security` | OWASP API Security Top 10 | semgrep |
| `iac-security` | Terraform, K8s, CloudFormation analysis | checkov or trivy |
| `container-security` | Dockerfile security and best practices | trivy or hadolint |
| `containers` | Container image analysis | trivy |

### Cryptography Analysis
| Scanner | Description | External Tool |
|---------|-------------|---------------|
| `crypto-ciphers` | Weak/deprecated algorithms (DES, RC4, MD5) | semgrep |
| `crypto-keys` | Hardcoded keys, weak key lengths | semgrep |
| `crypto-random` | Insecure random number generation | semgrep |
| `crypto-tls` | TLS configuration issues | semgrep |
| `digital-certificates` | X.509 certificate analysis | - |

### Developer Productivity
| Scanner | Description | External Tool |
|---------|-------------|---------------|
| `tech-discovery` | Framework and language detection (100+ patterns) | - |
| `code-ownership` | CODEOWNERS, bus factor, contributor analysis | - |
| `dora` | DORA metrics (deployment freq, lead time, MTTR) | GitHub API |
| `git` | Commit patterns, contributor activity | - |
| `documentation` | README quality and docs coverage | - |
| `test-coverage` | Test framework detection and coverage | - |
| `tech-debt` | Code duplication, complexity, TODO markers | - |
| `licenses` | SPDX license detection with policy | - |

## Configuration

### Environment Variables

```bash
# GitHub authentication (required for GitHub API scanners)
export GITHUB_TOKEN="ghp_..."
# Or use: gh auth login

# Claude API key (for AI agent analysis)
export ANTHROPIC_API_KEY="sk-ant-..."
```

### Configuration File

Zero uses `config/zero.config.json` for scanner configuration:

```json
{
  "settings": {
    "default_profile": "standard",
    "scanner_timeout_seconds": 300,
    "parallel_jobs": 4
  },
  "scanners": {
    "package-sbom": {
      "options": {
        "sbom": {
          "tool": "auto",
          "spec_version": "1.5",
          "recurse": true,
          "install_deps": false,
          "fallback_to_syft": true
        }
      }
    }
  }
}
```

### SBOM Options

The SBOM scanner supports extensive configuration:

| Option | Default | Description |
|--------|---------|-------------|
| `tool` | `auto` | Tool preference: `cdxgen`, `syft`, or `auto` |
| `spec_version` | `1.5` | CycloneDX spec version |
| `format` | `json` | Output format: `json` or `xml` |
| `recurse` | `true` | Recurse into subdirectories (mono-repos) |
| `install_deps` | `false` | Install dependencies before scanning |
| `babel_analysis` | `false` | Run babel for JS/TS analysis |
| `deep` | `false` | Deep analysis for C/C++, OCI images |
| `evidence` | `false` | Generate evidence in SBOM |
| `profile` | `generic` | cdxgen profile: `generic`, `research`, `appsec` |
| `fallback_to_syft` | `true` | Fall back to syft if cdxgen fails |

### Scan Profiles

| Profile | Scanners |
|---------|----------|
| `quick` | sbom, tech-discovery, vulnerabilities, licenses |
| `standard` | + package-health, code-secrets, code-ownership |
| `security` | sbom, vulns, code-security, iac-security, secrets, malcontent, container-security |
| `packages` | sbom, vulns, health, provenance, malcontent, bundle-optimization, recommendations |
| `advanced` | All scanners |

## Checkup Command

The `checkup` command helps you understand what scanners will work with your current setup:

```bash
./main checkup
```

This shows:
- **Token Status** - Whether your GitHub token is valid and its type (classic PAT, fine-grained PAT, OAuth)
- **Token Permissions** - Scopes and permissions available
- **External Tools** - Which required tools are installed
- **Scanner Compatibility** - Which scanners are ready, limited, or unavailable
- **Recommendations** - What permissions or tools to add

## AI Agent System

Zero includes specialist AI agents (powered by Claude) for deep analysis:

| Agent | Character | Expertise |
|-------|-----------|-----------|
| **Zero** | Zero Cool | Master orchestrator |
| **Cereal** | Cereal Killer | Supply chain security, malware, CVEs |
| **Razor** | Razor | Code security, SAST, secrets |
| **Blade** | Blade | Compliance, SOC 2, ISO 27001 |
| **Phreak** | Phantom Phreak | Legal, licenses, data privacy |
| **Acid** | Acid Burn | Frontend, React, TypeScript |
| **Dade** | Dade Murphy | Backend, APIs, databases |
| **Nikon** | Lord Nikon | Architecture, system design |
| **Joey** | Joey | Build, CI/CD, performance |
| **Plague** | The Plague | DevOps, infrastructure, Kubernetes |
| **Gibson** | The Gibson | Engineering metrics, DORA |
| **Gill** | Gill Bates | Cryptography, TLS, keys |

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
                ├── package-sbom.json
                ├── package-vulns.json
                ├── code-secrets.json
                └── ...
```

## Architecture

```
zero/
├── cmd/zero/                   # CLI entry point
│   └── cmd/                    # Cobra commands
├── pkg/
│   ├── scanner/                # Scanner framework
│   ├── scanners/               # Scanner implementations (25+)
│   │   ├── sbom/
│   │   ├── vulns/
│   │   ├── secrets/
│   │   └── ...
│   ├── config/                 # Configuration handling
│   ├── github/                 # GitHub API client
│   ├── hydrate/                # Clone + scan orchestration
│   └── terminal/               # Terminal UI
├── agents/                     # AI agent definitions
├── rag/                        # RAG knowledge base
├── config/                     # Configuration files
├── legacy/                     # Old shell-based implementation
```

## Development

### Building

```bash
# Build
go build -o main ./cmd/zero

# Run tests
go test ./...

# Run specific scanner tests
go test ./pkg/scanners/sbom/...
```

### Adding a New Scanner

1. Create a new package in `pkg/scanners/<name>/`
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
4. Import in `pkg/scanners/all.go`

## Test Organization

We maintain [phantom-tests](https://github.com/phantom-tests) for safe testing:

```bash
./main hydrate phantom-tests/express
./main hydrate phantom-tests/platform
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

**Status**: Experimental Preview
**Version**: 6.0.0 (Go rewrite)
**Last Updated**: 2025-12-13

*"Hack the planet!"*
