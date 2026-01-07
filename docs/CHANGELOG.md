<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Changelog

All notable changes to Zero are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

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
