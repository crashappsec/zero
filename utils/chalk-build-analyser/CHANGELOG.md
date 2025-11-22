<!--
Copyright (c) 2024 Gibson Powers Contributors

SPDX-License-Identifier: GPL-3.0
-->

# Changelog - Chalk Build Analyser

All notable changes to the Chalk Build Analyser will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

**Status**: 🔬 Experimental

### Planned
- Configuration system integration
- Multi-artifact scanning (bulk, directory)
- Output format options (JSON, markdown, CSV)
- Policy-as-code validation
- Artifact comparison
- Historical tracking
- Dashboard integration
- Alerting/notifications
- Comprehensive error handling
- Complete test suite

## [1.0.0] - 2024-11-20

### Added
- Initial release of Chalk Build Analyser
- Chalk metadata extraction from artifacts
- Build information display:
  - Build context (time, host, user)
  - Source repository details
  - Git commit, branch, tag information
  - Environment configuration
  - Tool versions
- Supply chain metadata insights
- AI-enhanced analysis with Claude (`chalk-build-analyser-claude.sh`)
- Comparison tool for base vs Claude analysis

### Known Limitations
- Single artifact analysis only
- No configuration system integration
- Limited output formats (text only)
- Basic error handling
- No policy validation
- No artifact comparison
- Requires Chalk tool installed
- No historical tracking

---

For details on other utilities, see the [main CHANGELOG](../../CHANGELOG.md).
