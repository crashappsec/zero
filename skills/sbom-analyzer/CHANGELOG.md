<!--
Copyright (c) 2024 Crash Override Inc
101 Fulton St, 416, New York 10038
SPDX-License-Identifier: GPL-3.0
-->

# Changelog

All notable changes to the SBOM/BOM Analyzer skill will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2024-11-20

### Added
- Initial release of SBOM/BOM Analyzer skill
- CycloneDX 1.7 (ECMA-424) format support
- SPDX format support
- OSV.dev API integration for vulnerability detection
  - Individual and batch query support
  - Version and commit-based queries
  - OpenSSF Vulnerability Format compatibility
- deps.dev API v3alpha integration
  - Package and version information queries
  - Dependency graph resolution
  - Security advisory correlation
  - OpenSSF Scorecard integration
  - SLSA provenance verification
  - Typosquatting detection
- CISA Known Exploited Vulnerabilities (KEV) catalog integration
  - Active exploitation detection
  - Remediation prioritization
  - Ransomware campaign tracking
- Comprehensive vulnerability analysis
  - Multi-source correlation (OSV.dev, deps.dev, CISA KEV)
  - CVSS severity scoring
  - Exploitability assessment
  - Transitive dependency tracking
- Dependency graph analysis
  - Direct and transitive dependency mapping
  - Circular dependency detection
  - Outdated package identification
  - Dependency depth analysis
- License compliance checking
  - SPDX expression validation
  - License conflict detection
  - Copyleft license flagging
  - Missing license identification
- Supply chain security assessment
  - Provenance verification
  - SLSA attestation checking
  - Component integrity validation
  - OpenSSF Scorecard evaluation
- Risk prioritization framework
  - Composite risk scoring
  - CISA KEV-based prioritization
  - Dependency criticality assessment
  - Maintenance status evaluation
- Multiple report formats
  - Executive summaries
  - Detailed vulnerability tables
  - Dependency tree visualizations (Mermaid)
  - License compliance matrices
  - Remediation action plans
- Comprehensive documentation
  - Usage examples
  - API integration guides
  - Best practices
  - Troubleshooting guidance

### Supported SBOM Formats
- CycloneDX JSON (`.cdx.json`, `bom.json`)
- CycloneDX XML (`.cdx.xml`, `bom.xml`)
- CycloneDX Protocol Buffers
- SPDX JSON
- SPDX YAML
- SPDX RDF
- SPDX Tag-Value

### Supported Package Ecosystems
- npm (Node.js)
- PyPI (Python)
- Maven (Java)
- Go modules
- Cargo (Rust)
- RubyGems (Ruby)
- NuGet (.NET)
- And all other ecosystems supported by OSV.dev

### Known Limitations
- API rate limits apply for OSV.dev and deps.dev queries
- Batch processing may be required for very large SBOMs (>1000 components)
- Vulnerability database coverage varies by ecosystem
- False positives possible depending on usage context
