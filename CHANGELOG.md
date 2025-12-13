<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Changelog

All notable changes to Zero will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [6.0.0] - 2025-12-13

### Major Architecture: Super Scanner v3.0

Consolidated the scanner architecture from 26+ individual scanners into **6 super scanners** with a clean dependency model.

### Added
- **SBOM Super Scanner** (`pkg/scanners/sbom/`): New standalone scanner that generates the authoritative SBOM
  - `generation` feature: Generates CycloneDX 1.5 SBOM using cdxgen/syft
  - `integrity` feature: Verifies SBOM against actual lockfiles
  - Produces `sbom.cdx.json` as source of truth for all package data
  - Other scanners can depend on SBOM output via `Dependencies()` interface

- **Scanner Dependencies**: New `Dependencies() []string` interface method
  - Scanners can declare dependencies on other scanners
  - Runner ensures dependent scanners complete first
  - `packages` scanner depends on `sbom` output

### Changed
- **Packages Scanner** (`pkg/scanners/packages/`): No longer generates SBOM
  - Removed SBOM generation - now uses `sbom.LoadSBOM()` from sbom scanner
  - Depends on sbom scanner via `Dependencies() []string { return []string{"sbom"} }`
  - 12 features: vulns, health, licenses, malcontent, confusion, typosquats, deprecations, duplicates, reachability, provenance, bundle, recommendations

- **Renamed Infra → DevOps** (`pkg/scanners/devops/`):
  - Better reflects scope: DevOps, CI/CD, infrastructure
  - Absorbed `github-actions-security` scanner
  - 5 features: iac, containers, github_actions, dora, git

- **Documentation**: Updated CLAUDE.md for v3.0 architecture

### Removed
- **Standalone scanners absorbed into super scanners**:
  - `github-actions-security` → absorbed into `devops` scanner
  - `dependency-confusion` → absorbed into `packages` scanner
  - `reachability-analysis` → absorbed into `packages` scanner
  - `sbom-integrity` → absorbed into `sbom` scanner

- **Old infra directory**: Renamed to `devops`

### Scanner Architecture (v3.0)

| Scanner | Features | Dependencies |
|---------|----------|--------------|
| **sbom** | generation, integrity | none (runs first) |
| **packages** | vulns, health, licenses, malcontent, confusion, typosquats, deprecations, duplicates, reachability, provenance, bundle, recommendations | sbom |
| **crypto** | ciphers, keys, random, tls, certificates | none |
| **code** | vulns, secrets, api, tech_debt | none |
| **devops** | iac, containers, github_actions, dora, git | none |
| **health** | technology, documentation, tests, ownership | none |

---

## [5.0.0] - 2025-12-06

### Major Rebranding: Phantom → Zero

The project has been renamed from **Phantom** to **Zero**, with all agents renamed after characters from the movie **Hackers (1995)**. "Hack the planet!"

### Added
- **Zero Orchestrator**: New master orchestrator named after Zero Cool
  - `/agent` slash command to chat with Zero
  - Zero delegates to specialist agents automatically
  - Full investigation capability with Read, Grep, Glob, WebSearch tools

- **Hackers-Themed Agent Team**:
  | New Name | Old Name | Character | Role |
  |----------|----------|-----------|------|
  | Zero | (new) | Zero Cool | Master orchestrator |
  | Cereal | Scout | Cereal Killer | Supply chain security |
  | Razor | Sentinel | Razor | Code security |
  | Blade | Quinn | Blade | Compliance auditing |
  | Phreak | Harper | Phantom Phreak | Legal counsel |
  | Acid | Casey | Acid Burn | Frontend engineer |
  | Dade | Morgan | Dade Murphy | Backend engineer |
  | Nikon | Ada | Lord Nikon | Software architect |
  | Joey | Bailey | Joey | Build engineer |
  | Plague | Phoenix | The Plague | DevOps engineer |
  | Gibson | Jordan | The Gibson | Engineering metrics |

- **Malcontent Scanner Integration**: Supply chain compromise detection with 14,500+ YARA rules
- **Agent Loader Library**: `utils/zero/lib/agent-loader.sh` for loading agent context
- **Enhanced Agent Definitions**: All 10 agents with knowledge bases and prompts

### Changed
- **Directory Structure**:
  - `utils/phantom/` → `utils/zero/`
  - `phantom.sh` → `zero.sh`
  - `.phantom/` → `.zero/`
  - `utils/zero/lib/phantom-lib.sh` → `utils/zero/lib/zero-lib.sh`
  - `utils/zero/config/phantom.config.json` → `utils/zero/config/zero.config.json`

- **Slash Commands**:
  - `/phantom` → `/zero`
  - New `/agent` for agent mode

- **Environment Variables**:
  - `PHANTOM_HOME` → `ZERO_HOME`
  - Data stored in `~/.zero/` instead of `~/.phantom/`

- **CLI Branding**: New Zero ASCII banner with green color scheme

### Removed
- **Skills Directory**: Removed `skills/` - functionality replaced by agent knowledge bases
- **Obsolete Planning Docs**: Removed 13+ implemented planning documents from `docs/`
- **Test Reports**: Removed old `test-reports/` and `code-security-reports/` directories

### Migration Guide

1. Rename your data directory:
   ```bash
   mv ~/.phantom ~/.zero
   ```

2. Update any scripts using `phantom.sh`:
   ```bash
   # Old
   ./phantom.sh hydrate owner/repo

   # New
   ./zero.sh hydrate owner/repo
   ```

3. Update agent names in prompts:
   - `scout` → `cereal`
   - `sentinel` → `razor`
   - `quinn` → `blade`
   - etc.

---

## [4.1.0] - 2025-12-03

### Changed
- **Status Indicator Updates**: Clarified development maturity levels
  - 🚀 **Beta**: Feature-complete, comprehensively tested, ready for broader use (Supply Chain, Better Prompts)
  - 🔬 **Experimental**: Early development, basic functionality working, not yet ready for production (DORA Metrics, Code Ownership, Certificate Analyser, Chalk Build Analyser)
  - 🧪 **Alpha**: (Reserved for very early prototypes)
  - Removed "Production Ready" designation - all tools under active development

### Added
- **Individual Utility Documentation**: Each utility now has comprehensive documentation
  - README.md in each utils subdirectory with status, usage, and roadmap
  - CHANGELOG.md in each utils subdirectory tracking module-specific changes
  - Clear development status indicators (Beta vs Experimental)
  - Supply Chain marked as 🚀 Beta - feature-complete and tested
  - DORA Metrics, Code Ownership, Certificate Analyser, Chalk Build Analyser marked as 🔬 Experimental
  - Maintains aggregated CHANGELOG (this file) for cross-utility changes

### Changed
- **Repository Restructure**
  - Renamed `tools/` to `utils/` for better clarity
  - Moved all scripts from `skills/*/` to `utils/*/` organized by topic
  - Skills directory now contains only skill files and documentation
  - Utils directory contains all executable scripts and utilities
  - Renamed "SBOM" to "Supply Chain" throughout for broader scope
  - Created central CHANGELOG.md (this file) consolidating all skill changelogs

## Supply Chain Analyser

### [2.2.0] - 2024-11-21

#### Added
- **Hierarchical Configuration System**: Global and module-specific config architecture
  - Global config at `utils/config.json` for organization-wide settings
  - Module-specific configs at `utils/<module>/config.json` for overrides
  - Config loading library at `utils/lib/config-loader.sh`
  - Configuration precedence: CLI args > module config > global config
  - `ignore_module_configs` flag to force global-only settings
  - Helper functions: `get_organizations()`, `get_repositories()`, `get_default_modules()`
- **Configuration Documentation**: Comprehensive `utils/CONFIG.md` guide
  - Setup instructions and quick start
  - Security considerations for PAT storage
  - Migration guide from old configs
  - Troubleshooting and best practices
- **Default Module Settings**: All analysis engines included by default
  - Supply chain: `["vulnerability", "provenance"]`
  - Automatic loading when no CLI modules specified
  - Configurable per module in global or module config

#### Changed
- **Config Loading**: All supply chain scripts now use hierarchical config system
  - `supply-chain-scanner.sh`: Integrated config-loader library
  - `vulnerability-analyser.sh`: Uses config for defaults
  - `vulnerability-analyser-claude.sh`: Inherits config integration
  - `provenance-analyser.sh`: Uses config for trust settings
  - `provenance-analyser-claude.sh`: Inherits config integration
- **Module Defaults**: Config-driven instead of hardcoded
  - Loads `default_modules` from config if no CLI flags
  - Supports per-module customization
  - Backward compatible with CLI-only usage

#### Technical Details
- Config merge algorithm: Deep merge with module override
- Array replacement (not concatenation) for lists
- Environment variable support via config-loader
- jq-based JSON parsing and validation
- Exported functions for cross-script usage

#### Migration
- Old: Module-specific configs only
- New: Global config with optional module overrides
- Action: Copy `utils/config.example.json` to `utils/config.json`
- No breaking changes: CLI-only usage still works

### [2.1.0] - 2024-11-21

#### Added
- **Provenance Analysis Module**: New SLSA provenance verification
  - `provenance-analyser.sh`: Base analyser with SLSA level assessment (0-4)
  - `provenance-analyser-claude.sh`: AI-enhanced with trust assessment and risk analysis
  - npm provenance checking with registry API integration
  - Signature verification support (cosign/rekor)
  - Multi-repo and organization scanning
  - Package URL (purl) analysis
- **RAG Knowledge Base**: Technical specifications optimized for AI consumption
  - SLSA v1.0 specification with provenance formats
  - CycloneDX v1.7 reference
  - Sigstore (cosign/rekor/fulcio) documentation
  - Structured for semantic search and RAG systems
- **Central Orchestrator Updates**:
  - Added `--provenance/-p` module flag
  - Integrated provenance analysis into `--all` option
  - Consistent multi-repo architecture

#### Changed
- Updated supply chain skill documentation with provenance analysis
- Enhanced bootstrap.sh with cosign and rekor-cli checks
- Expanded tool ecosystem coverage

#### Technical Details
- SLSA level compliance checking (0-4)
- Provenance attestation validation
- Builder identity verification
- Transparency log integration (Rekor)
- Multi-ecosystem support foundation (npm, more coming)

### [2.0.0] - 2024-11-21

#### Breaking Changes
- **Directory Restructure**: Renamed skills/supply-chain-analyser → skills/supply-chain
- **Script Renames**: supply-chain-analyser → vulnerability-analyser (moved to vulnerability-analysis subdirectory)
- **Modular Architecture**: Scripts reorganized into single-purpose modules with central orchestrator

#### Added
- **Central Orchestrator**: supply-chain-scanner.sh for unified entry point
  - `--setup`: Interactive configuration wizard with GitHub auth
  - `--interactive`: Prompt for repos if not configured
  - Module flags: `--vulnerability`, `--all` (extensible for future modules)
  - Multi-repo support: `--org` and `--repo` flags
- **Configuration Management**: config.json for persistent settings
  - GitHub Personal Access Token storage
  - Organizations and repositories lists
  - Default modules and output directories
  - Automatic config loading and validation
- **Multi-Repository Scanning**: Both analysers support organization/multi-repo scanning
  - GitHub CLI integration for org expansion (lists all repos in org)
  - Batch processing across multiple repositories
  - Individual repo targeting with `--repo owner/repo` flag
  - Config-based scanning for regular workflows
- **Interactive Setup**: Guided configuration wizard
  - GitHub authentication check
  - Organization selection from user's orgs
  - Manual repository entry
  - PAT configuration (optional)

#### Changed
- **Modular Architecture**: utils/supply-chain/ now contains:
  - vulnerability-analysis/ - Vulnerability scanning module (with both analysers)
  - config.example.json - Configuration template
  - supply-chain-scanner.sh - Central orchestrator
- **Script Organization**: Single-purpose scripts in feature subdirectories
- **Script Naming**: Clearer module-specific names (supply-chain-analyser → vulnerability-analyser)
- **Execution Model**: Scripts work standalone OR through central orchestrator
- **Output Headers**: Color-coded with CYAN for multi-repo section headers

#### Technical Improvements
- Consistent error handling across multi-repo workflows
- GitHub CLI (gh) integration for organization scanning
- jq-based configuration parsing
- Fallback to interactive mode when config missing
- Improved path resolution for nested script directories

#### Migration Guide
- Old path: `utils/supply-chain/supply-chain-analyser.sh`
- New path: `utils/supply-chain/vulnerability-analysis/vulnerability-analyser.sh`
- Or use central orchestrator: `utils/supply-chain/supply-chain-scanner.sh --vulnerability`
- Run `./utils/supply-chain/supply-chain-scanner.sh --setup` for interactive configuration

### [1.4.0] - 2024-11-21

#### Added
- **Intelligent Prioritization in Base Analyser**
  - `--prioritize` flag for data-driven vulnerability ranking
  - CISA KEV catalog integration (auto-fetched on demand)
  - Algorithmic priority scoring based on KEV presence and CVSS scores
  - Color-coded output with priority levels
  - Summary statistics (total, by severity, KEV count)

#### Changed
- **Refocused Claude Analyser on AI-Specific Value**
  - Moved basic prioritization (CVSS, KEV, counting) to base analyser
  - Claude now focuses on pattern analysis, supply chain context, and risk narratives
  - Clear separation: Base analyser (data-driven) vs Claude (AI insights)

### [1.3.1] - 2024-11-21

#### Fixed
- Script execution issues with `find_sbom()` and `set -e` compatibility
- SBOM filename compatibility (changed to `bom.json` per osv-scanner spec)
- Updated osv-scanner flag from deprecated `--sbom` to `-L`
- Fixed output capture in `run_osv_scanner()`
- Added JSON extraction from osv-scanner mixed output

#### Added
- SBOM generation integration with syft
- Automatic SBOM generation when no SBOM exists
- Enhanced documentation with syft usage and best practices

### [1.3.0] - 2024-11-20

#### Added
- Taint analysis capability with osv-scanner
- Reachability determination (CALLED, NOT CALLED, UNKNOWN)
- Automation scripts for CI/CD integration
- Support for Go projects with experimental call analysis

### [1.2.0] - 2024-11-20

#### Added
- SLSA (Supply-chain Levels for Software Artifacts) expertise
- SLSA provenance format understanding and validation
- Build platform identification and assessment

### [1.1.0] - 2024-11-20

#### Added
- Format conversion capabilities (CycloneDX ↔ SPDX)
- Version upgrade capabilities for both formats
- Bidirectional conversion workflows

### [1.0.0] - 2024-11-20

#### Added
- Initial release of Supply Chain Analyser
- CycloneDX 1.7 and SPDX format support
- OSV.dev, deps.dev, and CISA KEV integration
- Vulnerability analysis and license compliance
- Dependency graph analysis

## DORA Metrics

### [1.1.0] - 2024-11-20

#### Added
- Automation scripts for command-line DORA analysis
- CI/CD integration support
- Comparison tool for basic vs Claude-enhanced analysis

### [1.0.0] - 2024-11-20

#### Added
- Initial release of DORA Metrics analyser
- All four key metrics calculation
- Performance classification (Elite/High/Medium/Low)
- Benchmark comparison and trend analysis

## Code Ownership

### [1.0.0] - 2024-11-20

#### Added
- Initial release of Code Ownership analyser
- Git history analysis with weighted scoring
- CODEOWNERS file validation and generation
- Ownership metrics and health scores
- Bus factor risk identification

## Certificate Analyser

### [1.0.0] - 2024-11-20

#### Added
- Initial release of Certificate Analyser
- TLS/SSL certificate validation
- Expiration checking and security assessment

## Chalk Build Analyser

### [1.0.0] - 2024-11-20

#### Added
- Initial release of Chalk Build Analyser
- Build artifact analysis
- Supply chain metadata insights

## Better Prompts

### [1.0.0] - 2024-11-20

#### Added
- Initial release of Better Prompts skill
- Prompt engineering techniques
- Before/after examples and conversation patterns

---

For detailed feature documentation, see individual skill README files in `skills/` directory.
