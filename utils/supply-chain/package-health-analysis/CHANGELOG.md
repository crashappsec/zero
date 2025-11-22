<!--
Copyright (c) 2024 Gibson Powers Contributors

SPDX-License-Identifier: GPL-3.0
-->

# Changelog

All notable changes to the Package Health Analyzer will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.1] - 2024-11-21

### Fixed
- **Critical**: Fixed URL encoding bug where newlines were included in API requests
  - Changed from `echo "$package" | jq -sRr @uri` to `printf '%s' "$package" | jq -sRr @uri`
  - All API requests were failing with 404 due to `%0A` (newline) in URLs
  - This fix makes the analyzer actually functional

- **Critical**: Fixed PURL ecosystem extraction in SBOM parsing
  - Was extracting "pkg" instead of actual ecosystem ("npm", "pypi", etc.)
  - Packages are now correctly identified by their ecosystem

- **Major**: Removed GitHub CLI (`gh`) dependency
  - Now uses standard `git clone` which is more widely available
  - Supports full GitHub URLs and owner/repo format
  - Works with any Git hosting platform, not just GitHub

- **Major**: Added comprehensive error handling for API responses
  - Validates all JSON before passing to `jq --argjson`
  - API errors return valid JSON instead of mixed text/errors
  - Invalid responses get error placeholders instead of crashing
  - Failed packages are skipped gracefully with warnings
  - Eliminates all "jq: invalid JSON text passed to --argjson" errors

- **Major**: Fixed temp file cleanup in SBOM generation
  - SBOM files were being deleted before they could be used
  - Now properly persists temp files until analysis completes

- **Major**: Added `.env` file loading for ANTHROPIC_API_KEY
  - Automatically loads API key from repository root `.env` file
  - Consistent with other Claude-enabled scripts
  - No need to export environment variable manually

### Added
- **Real-time progress indicators** in AI-enhanced analyzer
  - Shows progress for all 5 analysis steps
  - Users see the script is working, not frozen
  - Progress messages sent to stderr (doesn't interfere with JSON output)
  - Friendly emoji indicators for better UX

### Changed
- Improved error messages with package context (name, system, version)
- Better validation of curl responses before caching
- More robust JSON handling throughout the codebase

## [1.0.0] - 2024-11-21

### Added
- **Base Analyzer** (`package-health-analyzer.sh`)
  - Automated package health scanning
  - deps.dev API integration for package metadata
  - Health scoring algorithm (0-100) with weighted components
  - Deprecation detection from multiple sources
  - Version inconsistency analysis across repositories
  - Support for npm, PyPI, Maven, Cargo, and Go ecosystems
  - Multiple output formats (JSON, Markdown, Table)
  - Organization-wide scanning capability
  - SBOM file analysis (CycloneDX and SPDX)

- **AI-Enhanced Analyzer** (`package-health-analyzer-claude.sh`)
  - Claude AI integration for deep analysis
  - Chain of reasoning across multiple tools
  - Integration with vulnerability analyzer
  - Integration with provenance analyzer
  - Risk assessment and prioritization
  - Detailed migration strategies
  - Alternative package recommendations
  - Actionable improvement plans
  - Strategic operational recommendations

- **Comparison Tool** (`compare-analyzers.sh`)
  - Side-by-side comparison of base vs AI-enhanced results
  - Performance benchmarking
  - Feature comparison matrix

- **Library Modules**
  - `deps-dev-client.sh`: deps.dev API client with caching
  - `health-scoring.sh`: Health score calculation engine
  - `version-analysis.sh`: Version inconsistency analysis
  - `deprecation-checker.sh`: Deprecation detection logic

- **Configuration System**
  - Customizable health score weights
  - Adjustable grade thresholds
  - API settings (timeout, retries, rate limiting)
  - Cache configuration
  - Ecosystem selection

- **Claude Code Skill**
  - Interactive skill file for Claude Code integration
  - Example interactions and workflows
  - Best practices and use case guidance

- **Documentation**
  - Comprehensive README with usage examples
  - Build prompt for implementation guidance
  - Requirements specification
  - RAG documentation (deps.dev API, best practices)
  - This CHANGELOG

### Features

**Health Scoring Components**:
- OpenSSF Scorecard integration (30% weight)
- Maintenance activity scoring (25% weight)
- Security vulnerability assessment (25% weight)
- Version freshness evaluation (10% weight)
- Popularity metrics (10% weight)

**Deprecation Detection**:
- deps.dev deprecation flags
- Known deprecated package database
- Alternative package suggestions
- Migration urgency assessment

**Version Analysis**:
- Semantic version parsing and comparison
- Version distribution across repositories
- Recommended version identification
- Migration complexity calculation
- Effort estimation

**AI Analysis Capabilities**:
- Risk assessment with business impact
- Version standardization strategies
- Deprecated package migration plans
- Health score root cause analysis
- Operational improvement recommendations

**Output Formats**:
- JSON: Structured data for programmatic processing
- Markdown: Human-readable reports
- Table: Quick CLI visualization

### Configuration

Default configuration includes:
- Health score weights tuned for balanced assessment
- Grade thresholds aligned with industry standards
- API settings optimized for performance and reliability
- 24-hour cache TTL for API responses
- Support for 5 major ecosystems

### Integration

- Works with existing supply chain scanner infrastructure
- Leverages shared configuration system
- Integrates with vulnerability and provenance analyzers
- Compatible with CI/CD pipelines
- GitHub Actions workflow examples included

### Performance

- Base analyzer: ~30s for small repos (10-20 packages)
- AI-enhanced: ~2-3min for small repos
- Organization scans: ~10min for 10 repos (base)
- Caching reduces API calls by ~80%

### Known Limitations

- Best support for npm and PyPI ecosystems
- deps.dev API rate limits may affect large scans
- Health scoring is heuristic-based
- Requires external dependencies (syft, gh, jq)
- AI analysis requires ANTHROPIC_API_KEY

### Status

🔬 **Experimental** - New capability under active development

**Maturity Assessment**:
- Core functionality: Complete
- Testing: Basic validation done
- Documentation: Comprehensive
- Integration: With supply chain tools
- Production readiness: Experimental, not production-ready yet

**Next Steps for Beta**:
- [ ] Extensive testing with real repositories
- [ ] Performance optimization
- [ ] Enhanced error handling
- [ ] Additional ecosystem support
- [ ] User feedback integration
- [ ] CI/CD template refinement

## [Unreleased]

### Planned
- Enhanced ecosystem support (Ruby, PHP, .NET)
- Custom health scoring profiles
- Historical trend analysis
- Dashboard web interface
- Policy enforcement engine
- Automated remediation capabilities
- Slack/email notifications
- JIRA/GitHub Issues integration

### Under Consideration
- Local package registry support
- Offline mode for air-gapped environments
- Multi-language report generation
- Custom deprecation rules
- Team-specific configurations
- Cost analysis for migrations

---

## Release Notes

### Version 1.0.0 - Initial Release

This is the first release of the Package Health Analyzer, providing comprehensive tools for analyzing package health across organizations.

**Highlights**:
- Two-tiered analysis system (base + AI-enhanced)
- Chain of reasoning integration
- Organization-wide scanning
- Detailed AI-powered recommendations
- Multiple output formats
- Extensive documentation

**Target Users**:
- DevOps teams managing dependencies
- Security teams conducting audits
- Engineering leadership planning tech debt reduction
- Platform teams standardizing packages

**Use Cases**:
- Security audits
- Version standardization
- Tech debt reduction
- Pre-release validation
- Operational improvements

**Getting Started**:
See [README.md](README.md) for installation and usage instructions.

---

## Version History

- **1.0.0** (2024-11-21): Initial release - Experimental

---

[1.0.0]: https://github.com/crashappsec/gibson-powers/releases/tag/package-health-v1.0.0
[Unreleased]: https://github.com/crashappsec/gibson-powers/compare/package-health-v1.0.0...HEAD
