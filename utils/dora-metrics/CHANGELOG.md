<!--
Copyright (c) 2024 Crash Override Inc
101 Fulton St, 416, New York 10038
SPDX-License-Identifier: GPL-3.0
-->

# Changelog - DORA Metrics Analyzer

All notable changes to the DORA Metrics Analyzer will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

**Status**: 🚧 Experimental - Not production-ready

### Planned
- Configuration system integration
- Multi-repository scanning support
- Organization-wide analysis
- Output format options (JSON, markdown, CSV)
- Historical tracking and trend analysis
- Dashboard integration
- Comprehensive error handling
- Complete test suite

## [1.1.0] - 2024-11-20

### Added
- Automation scripts for command-line DORA analysis
- CI/CD integration support
- Comparison tool for basic vs Claude-enhanced analysis

## [1.0.0] - 2024-11-20

### Added
- Initial release of DORA Metrics analyzer
- All four key metrics calculation:
  - Deployment Frequency
  - Lead Time for Changes
  - Mean Time to Recovery (MTTR)
  - Change Failure Rate
- Performance classification (Elite/High/Medium/Low)
- Benchmark comparison and trend analysis
- AI-enhanced analysis with Claude (`dora-analyzer-claude.sh`)
- Basic comparison tool

### Known Limitations
- Single repository analysis only
- No configuration system integration
- Limited output formats (text only)
- Basic error handling
- Fixed 90-day analysis window
- Requires proper GitHub tagging/releases

---

For details on other utilities, see the [main CHANGELOG](../../CHANGELOG.md).
