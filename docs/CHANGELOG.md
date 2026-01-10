<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Changelog

All notable changes to Zero are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

---

## [6.0.0] - 2026-01-10

### Added

- **Engineering Intelligence Framework**: 6-pillar organization aligned with DORA, SPACE, LinearB
  - Speed (DORA metrics), Quality (code health), Team (people), Security (risk), Supply Chain (dependencies), Technology (stack)
  - Benchmark tiers: Elite, Good, Fair, Needs Focus based on LinearB 2026 benchmarks
  - Pillar-to-analyzer mapping for all 7 super scanners

- **Benchmark Visualization**: BenchmarkTier component across all pillar pages
  - Visual tier indicators with color coding (green/blue/yellow/red)
  - LinearB 2026 benchmark thresholds for all metrics
  - Security, Supply Chain, Quality, and Team benchmark configurations

- **Web UI Reorganization**
  - Sidebar navigation reordered: Speed, Quality, Team, Security, Supply Chain, Technology
  - New Speed page combining DORA metrics with LinearB benchmarks
  - Benchmark tier cards on all pillar pages

- **Finding Validation System**: Comprehensive suppression and validation
  - SARIF-format suppression files per project
  - Reasons: false_positive, wont_fix, acceptable_risk, test_code
  - Global and path-specific suppression rules
  - CLI commands: `zero validate add`, `zero validate remove`, `zero validate list`

- **Markdown Reports**: CLI-generated reports by category
  - `zero report <repo>` generates markdown reports
  - Category-specific: `--category security|supply-chain|quality|speed`

### Changed

- Removed Evidence.dev dependency (legacy report system)
- Updated all documentation to v6.0 framework
- Agent pillar assignments for domain clarity

### Fixed

- BenchmarkTier component handles all metric types (higher/lower is better)
- Consistent tier thresholds across web UI and reports

---

## [4.0.0] - 2026-01-07

### Added

- **RAG Pattern System**: 3500+ detection rules generated from knowledge base
  - Pattern validation framework with regex compilation checks
  - RAG-to-Semgrep conversion for all 23 categories
  - Agent RAG search capability for pattern discovery
  - Detection tests and false positive reduction

- **Agent System Enhancements**
  - `rag-search` capability for searching patterns by name, severity, or technology
  - `rag-detail` capability for retrieving full pattern details
  - GetSystemInfo tool for querying scanners, profiles, and RAG statistics

- **Report Generation**: Interactive HTML reports via Evidence.dev
  - Executive summary with risk scores
  - Security findings with severity breakdown
  - Dependency analysis with license distribution
  - Supply chain threat detection
  - DORA metrics visualization

- **Automation Commands**
  - `zero watch` - Watch directory for changes and auto-scan
  - `zero refresh` - Refresh stale scan data
  - `zero feeds rag` - Generate rules from RAG knowledge base
  - `zero feeds semgrep` - Sync Semgrep community rules

- **Freshness Tracking**: Know when scan data is stale
  - Fresh (< 24h), Stale (1-7d), Very Stale (7-30d), Expired (> 30d)

### Changed

- Consolidated to 7 super scanners (v4.0 architecture)
- Standardized RAG tooling to cdxgen, osv-scanner, semgrep
- Improved agent context loading with summary/critical/full modes
- Default hydrate limit changed to 25 repos

### Fixed

- False positive reduction in AWS, Kubernetes, NATS patterns
- Agent tool validation and empty input handling
- GetAnalysis context overflow prevention

---

## [3.0.0] - 2025-12-20

### Added

- **Super Scanner Architecture**: Consolidated from 9 to 7 scanners
  - code-packages (SBOM + 14 features)
  - code-security (SAST + secrets + crypto)
  - code-quality (metrics + coverage)
  - devops (IaC + containers + CI/CD)
  - technology-identification (tech detection + ML-BOM)
  - code-ownership (contributors + bus factor)
  - devx (developer experience)

- **Agent Definitions**: 12 specialist agents
  - Zero (orchestrator), Cereal (supply chain), Razor (security)
  - Blade (compliance), Phreak (legal), Acid (frontend)
  - Dade (backend), Nikon (architecture), Joey (CI/CD)
  - Plague (DevOps), Gibson (DORA), Gill (crypto), Hal (AI/ML)

- **Hydrate Command**: Clone and scan repositories
  - Profile-based scanning (all-quick, all-complete, individual scanners)
  - Org-level scanning with configurable limits
  - Demo mode for large repositories

- **Docker Support**: Run Zero in containers
  - All dependencies bundled
  - Consistent execution environment

### Changed

- Moved from individual scanners to feature-based super scanners
- Unified configuration in `zero.config.json`
- Standardized output format across all scanners

---

## [2.0.0] - 2025-11-15

### Added

- Initial scanner implementations
- Basic CLI commands (hydrate, status)
- RAG knowledge base foundation

### Changed

- Restructured project layout
- Introduced pkg-based architecture

---

## [1.0.0] - 2025-10-01

### Added

- Project initialization
- Core framework design
- Agent persona definitions

---

*"Hack the planet!"*
