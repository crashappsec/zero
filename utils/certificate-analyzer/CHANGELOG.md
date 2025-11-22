<!--
Copyright (c) 2024 Gibson Powers Contributors

SPDX-License-Identifier: GPL-3.0
-->

# Changelog - Certificate Analyzer

All notable changes to the Certificate Analyzer will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

**Status**: 🔬 Experimental

### Planned
- Configuration system integration
- Bulk domain scanning (file input, multiple domains)
- Certificate monitoring and alerting
- Output format options (JSON, markdown, CSV)
- Historical tracking and trend analysis
- OCSP stapling validation
- CT log verification
- Dashboard integration
- Comprehensive error handling
- Complete test suite

## [1.0.0] - 2024-11-20

### Added
- Initial release of Certificate Analyzer
- TLS/SSL certificate validation
- Expiration checking and warnings
- Certificate chain validation
- Security assessment:
  - Protocol version checking (TLS 1.2+)
  - Cipher suite evaluation
  - Key length validation
  - Signature algorithm checking
- Common name and SAN validation
- AI-enhanced analysis with Claude (`cert-analyzer-claude.sh`)
- Port specification support

### Known Limitations
- Single domain analysis only
- No configuration system integration
- Limited output formats (text only)
- Basic error handling
- No continuous monitoring
- Limited OCSP checking
- No historical tracking
- Requires internet connectivity

---

For details on other utilities, see the [main CHANGELOG](../../CHANGELOG.md).
