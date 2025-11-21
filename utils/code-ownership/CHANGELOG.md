<!--
Copyright (c) 2024 Crash Override Inc
101 Fulton St, 416, New York 10038
SPDX-License-Identifier: GPL-3.0
-->

# Changelog - Code Ownership Analyzer

All notable changes to the Code Ownership Analyzer will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

**Status**: 🚧 Experimental - Not production-ready

### Planned
- Configuration system integration
- Multi-repository scanning support
- Organization-wide analysis
- Output format options (JSON, markdown, CSV)
- GitHub username mapping
- Historical trend tracking
- Team structure analysis
- Code review integration
- Dashboard integration
- Comprehensive error handling
- Complete test suite

## [1.0.0] - 2024-11-20

### Added
- Initial release of Code Ownership analyzer
- Git history analysis with weighted scoring:
  - Commit count weighting
  - Lines changed weighting
  - Recency factor
- CODEOWNERS file validation
- CODEOWNERS file generation
- Ownership metrics by directory
- Health scores calculation
- Bus factor risk identification
- AI-enhanced analysis with Claude (`ownership-analyzer-claude.sh`)
- Comparison tool for base vs Claude analysis

### Known Limitations
- Single repository analysis only
- No configuration system integration
- Limited output formats (text only)
- Basic error handling
- Email-based only (no GitHub username mapping)
- Analyzes full Git history (no time range selection)
- Doesn't account for code review contributions
- May not reflect current team structure

---

For details on other utilities, see the [main CHANGELOG](../../CHANGELOG.md).
