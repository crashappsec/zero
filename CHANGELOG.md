<!--
Copyright (c) 2024 Crash Override Inc
101 Fulton St, 416, New York 10038
SPDX-License-Identifier: GPL-3.0
-->

# Changelog

All notable changes to the Skills and Prompts repository will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- **Repository Restructure**
  - Renamed `tools/` to `utils/` for better clarity
  - Moved all scripts from `skills/*/` to `utils/*/` organized by topic
  - Skills directory now contains only skill files and documentation
  - Utils directory contains all executable scripts and utilities
  - Renamed "SBOM" to "Supply Chain" throughout for broader scope
  - Created central CHANGELOG.md (this file) consolidating all skill changelogs

## Supply Chain Analyzer

### [2.0.0] - 2024-11-21

#### Breaking Changes
- **Directory Restructure**: Renamed skills/supply-chain-analyzer → skills/supply-chain
- **Script Renames**: supply-chain-analyzer → vulnerability-analyzer (moved to vulnerability-analysis subdirectory)
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
- **Multi-Repository Scanning**: Both analyzers support organization/multi-repo scanning
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
  - vulnerability-analysis/ - Vulnerability scanning module (with both analyzers)
  - config.example.json - Configuration template
  - supply-chain-scanner.sh - Central orchestrator
- **Script Organization**: Single-purpose scripts in feature subdirectories
- **Script Naming**: Clearer module-specific names (supply-chain-analyzer → vulnerability-analyzer)
- **Execution Model**: Scripts work standalone OR through central orchestrator
- **Output Headers**: Color-coded with CYAN for multi-repo section headers

#### Technical Improvements
- Consistent error handling across multi-repo workflows
- GitHub CLI (gh) integration for organization scanning
- jq-based configuration parsing
- Fallback to interactive mode when config missing
- Improved path resolution for nested script directories

#### Migration Guide
- Old path: `utils/supply-chain/supply-chain-analyzer.sh`
- New path: `utils/supply-chain/vulnerability-analysis/vulnerability-analyzer.sh`
- Or use central orchestrator: `utils/supply-chain/supply-chain-scanner.sh --vulnerability`
- Run `./utils/supply-chain/supply-chain-scanner.sh --setup` for interactive configuration

### [1.4.0] - 2024-11-21

#### Added
- **Intelligent Prioritization in Base Analyzer**
  - `--prioritize` flag for data-driven vulnerability ranking
  - CISA KEV catalog integration (auto-fetched on demand)
  - Algorithmic priority scoring based on KEV presence and CVSS scores
  - Color-coded output with priority levels
  - Summary statistics (total, by severity, KEV count)

#### Changed
- **Refocused Claude Analyzer on AI-Specific Value**
  - Moved basic prioritization (CVSS, KEV, counting) to base analyzer
  - Claude now focuses on pattern analysis, supply chain context, and risk narratives
  - Clear separation: Base analyzer (data-driven) vs Claude (AI insights)

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
- Initial release of Supply Chain Analyzer
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
- Initial release of DORA Metrics analyzer
- All four key metrics calculation
- Performance classification (Elite/High/Medium/Low)
- Benchmark comparison and trend analysis

## Code Ownership

### [1.0.0] - 2024-11-20

#### Added
- Initial release of Code Ownership analyzer
- Git history analysis with weighted scoring
- CODEOWNERS file validation and generation
- Ownership metrics and health scores
- Bus factor risk identification

## Certificate Analyzer

### [1.0.0] - 2024-11-20

#### Added
- Initial release of Certificate Analyzer
- TLS/SSL certificate validation
- Expiration checking and security assessment

## Chalk Build Analyzer

### [1.0.0] - 2024-11-20

#### Added
- Initial release of Chalk Build Analyzer
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
