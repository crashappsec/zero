<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Changelog

All notable changes to Zero will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [4.1.0] - 2026-01-05

### Web UI & API Server

Added a complete web interface for visualizing scan results and managing projects.

#### Added

- **Web UI** (`web/`): Next.js dashboard for visualization
  - Project browser with search and filtering
  - Real-time scan progress via WebSocket
  - Scanner-specific analysis views
  - Dark/light theme support
  - Keyboard shortcuts and export functionality

- **API Server** (`./zero serve`): REST API + WebSocket for real-time updates
  - `GET /api/projects` - List all analyzed projects
  - `GET /api/projects/:id` - Get project details with scan data
  - `POST /api/scans` - Start a new scan
  - `WS /ws` - Real-time scan progress updates
  - SQLite storage layer for performance

- **Configuration System**: Multi-source configuration loading
  - `config/defaults/scanners.json` - Scanner feature defaults
  - `config/zero.config.json` - Main config with profiles
  - `~/.zero/config.json` - User overrides (optional)
  - Profile-based scanner selection

- **Credentials Management** (`./zero config`):
  - `zero config` - View current credentials
  - `zero config set github_token` - Set GitHub token
  - `zero config set anthropic_key` - Set Anthropic API key
  - Stored securely at `~/.zero/credentials.json` (0600 permissions)

- **Enhanced Checkup** (`./zero checkup`):
  - Lists all accessible repos for GitHub token
  - Shows public/private visibility per repo
  - Recommends fine-grained PATs over classic tokens
  - Security warning for overly-broad classic PATs

- **Demo Scripts** (`demo/`):
  - `LOOM_SCRIPT.md` - Step-by-step demo recording guide
  - `COMMANDS.sh` - All demo commands in sequence

### Changed

- Scanner consolidation from 9 to 7 super scanners (v4.0 architecture)
- All documentation updated for `./zero` binary (was `./main`)
- Configuration now uses profiles instead of inline scanner configs

### Fixed

- `./zero report` removed - use `./zero serve` for web UI
- Config loading priority clarified in documentation
- Scanner names consistent across codebase and docs

---

## [4.0.0] - 2025-12-28

### Super Scanner Architecture v4.0

Consolidated from 9 to **7 super scanners** with cleaner feature organization.

#### Scanner Changes

| Scanner | Features | Description |
|---------|----------|-------------|
| **code-packages** | generation, vulns, health, licenses, malcontent, typosquats, deprecations, duplicates, confusion, provenance, reachability | SBOM + package analysis |
| **code-security** | vulns, secrets, api, ciphers, keys, random, tls, certificates | Code security + cryptography |
| **code-quality** | tech_debt, complexity, test_coverage, code_docs | Quality metrics |
| **devops** | iac, containers, github_actions, dora, git | DevOps analysis |
| **technology-identification** | detection, models, frameworks, datasets, ai_security, ai_governance, infrastructure | Tech detection + ML-BOM |
| **code-ownership** | contributors, bus_factor, codeowners, orphans, churn, patterns | Ownership analysis |
| **developer-experience** | onboarding, sprawl, workflow | DevX analysis |

#### Key Changes

- **code-crypto merged into code-security**: Cryptography features now under code-security scanner
- **devx renamed to developer-experience**: Clearer naming
- **technology-identification**: Renamed from tech-id for clarity
- Output files follow scanner names exactly

---

## [3.6.0] - 2025-12-20

### Checkup Command Enhancements

#### Added

- **Scanner Requirements**: Checkup shows which tools each scanner needs
- **Feature-Level Status**: Reports which features are available/limited
- **Auto-Fix**: `./zero checkup --fix` installs missing tools
- **Permission Checking**: Validates GitHub token scopes

---

## [3.5.0] - 2025-12-16

### Tech-ID Scanner & Code Ownership

#### Added

- **Semgrep Integration**: RAG-to-Semgrep converter for technology detection
- **ML-BOM Generation**: Machine Learning Bill of Materials
- **Code Ownership Enhancements**: Adaptive period detection, activity status
- **Hal Agent**: AI/ML security specialist
- **Gill Agent**: Cryptography specialist (named after Gill Bates from Hackers)

#### Changed

- Scanner names updated for consistency
- Output files renamed to match scanner names

---

## [3.0.0] - 2025-12-13

### Super Scanner Architecture v3.0

Consolidated from 26+ individual scanners into **6 super scanners**.

#### Added

- **SBOM Super Scanner**: Standalone CycloneDX SBOM generation
- **Scanner Dependencies**: `Dependencies() []string` interface
- **Packages Scanner**: Depends on SBOM output

#### Removed

- Legacy individual scanners absorbed into super scanners

---

## [2.0.0] - 2025-12-06

### Phantom → Zero Rebranding

Renamed from **Phantom** to **Zero** with Hackers (1995) themed agents.

#### Added

- **Zero Orchestrator**: Master orchestrator (Zero Cool)
- **Agent Team**: 12 specialists named after Hackers characters
  - Cereal (supply chain), Razor (security), Blade (compliance)
  - Phreak (legal), Acid (frontend), Dade (backend)
  - Nikon (architecture), Joey (build), Plague (devops)
  - Gibson (metrics), Gill (crypto), Hal (AI/ML)

- **Malcontent Integration**: 14,500+ YARA rules for supply chain detection

#### Changed

- `.phantom/` → `.zero/`
- `PHANTOM_HOME` → `ZERO_HOME`
- All agent names updated to Hackers theme

---

## [1.0.0] - 2025-11-20

### Initial Release

- Go CLI with modular scanner architecture
- CycloneDX SBOM generation
- Vulnerability scanning via OSV.dev
- Secret detection
- DORA metrics calculation
- Code ownership analysis

---

For detailed documentation, see [docs/README.md](docs/README.md).
