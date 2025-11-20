<!--
Copyright (c) 2024 Crash Override Inc
101 Fulton St, 416, New York 10038
SPDX-License-Identifier: GPL-3.0
-->

# Changelog

All notable changes to the SBOM/BOM Analyzer skill will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.3.0] - 2024-11-20

### Added
- **Taint Analysis Capability**
  - osv-scanner integration for call graph/taint analysis
  - Reachability determination (CALLED, NOT CALLED, UNKNOWN)
  - Vulnerability exploitability assessment beyond SBOM component presence
  - Support for Go projects with experimental call analysis
  - Integration workflow combining SBOM scanning with reachability analysis
  - Differentiation between present vulnerabilities and actually exploitable ones

- **Automation Scripts for CI/CD Integration**
  - `sbom-analyzer.sh` - Basic SBOM scanning with osv-scanner
    - SBOM file analysis (JSON/XML)
    - Git repository scanning with auto-cloning
    - Local directory scanning
    - Taint analysis flag (`--taint-analysis`)
    - Multiple output formats (table, JSON, markdown, SARIF)
    - Flexible target handling (SBOM files, Git URLs, local paths)

  - `sbom-analyzer-claude.sh` - AI-enhanced SBOM analysis
    - All features from basic analyzer
    - Claude API integration (claude-sonnet-4-20250514)
    - Executive summaries with risk assessment
    - Critical findings prioritization
    - Remediation guidance with specific version upgrades
    - CISA KEV correlation and exploitation context
    - Supply chain risk assessment
    - Actionable recommendations

  - `compare-analyzers.sh` - Comparison tool
    - Runs both basic and Claude-enhanced analyzers
    - Side-by-side capability comparison
    - Value-add demonstration
    - Comprehensive comparison report
    - Optional output file preservation
    - Use case recommendations

- **Enhanced Documentation**
  - Automation scripts section in README
  - Usage examples for all three scripts
  - CI/CD integration examples (GitHub Actions, GitLab CI)
  - Prerequisites and requirements documentation
  - osv-scanner taint analysis procedures in skill file

### Requirements
- osv-scanner: `go install github.com/google/osv-scanner/cmd/osv-scanner@latest`
- jq: `brew install jq` (or `apt-get install jq`)
- Anthropic API key (for Claude-enhanced analyzer)

### Use Cases
- **CI/CD Pipelines**: Automated SBOM scanning in GitHub Actions, GitLab CI, etc.
- **Local Development**: Quick vulnerability checks before committing
- **Security Audits**: Comprehensive analysis with AI-enhanced insights
- **Taint Analysis**: Reachability testing for Go projects
- **Comparison Analysis**: Demonstrate AI value-add to stakeholders

## [1.2.0] - 2024-11-20

### Added
- **SLSA (Supply-chain Levels for Software Artifacts) Expertise**
  - Comprehensive knowledge of SLSA v1.0 framework
  - All SLSA levels (0-4) with requirements and verification
  - SLSA provenance format understanding and validation
  - Integration with CycloneDX and SPDX for SLSA documentation
  - SLSA verification in SBOM analysis (Level 1-3)
  - Build platform identification and assessment
  - Use cases: procurement, internal compliance, risk assessment, incident response
  - Level-specific recommendations for improving security posture
  - SLSA and vulnerability management integration
  - Tools and standards (slsa-verifier, in-toto, SigStore, GUAC)

## [1.1.0] - 2024-11-20

### Added
- **Format Conversion Capabilities** (CycloneDX ↔ SPDX)
  - Bidirectional conversion between CycloneDX and SPDX formats
  - Intelligent field mapping with format-specific handling
  - Support for all CycloneDX formats (JSON, XML, Protocol Buffers)
  - Support for all SPDX formats (JSON, YAML, RDF/XML, Tag-Value)
  - Preservation of metadata and provenance during conversion
  - PURL (Package URL) maintenance across formats
  - Vulnerability data conversion
  - License expression translation
  - Conversion validation and reporting
  - Handling of format-specific features with annotations

- **Version Upgrade Capabilities**
  - CycloneDX version upgrades (1.0-1.6 → 1.7)
    - Automated addition of new required fields
    - Structure modernization
    - Schema validation
    - Backwards compatibility preservation
  - SPDX version upgrades (2.0-2.2 → 2.3)
    - Relationship type updates
    - Security reference additions
    - Enhanced metadata fields
  - Upgrade reporting with detailed change logs
  - Best practices enforcement
  - Missing data handling with appropriate defaults
  - Bidirectional conversion workflows for complex upgrades

- **New Prompt Templates**
  - `format-conversion.md`: Comprehensive format conversion guide
  - `version-upgrade.md`: Version upgrade procedures and best practices
  - Reorganized prompt structure by category (security/operations/compliance)

### Changed
- Expanded skill scope beyond security to include operational SBOM management
- Reorganized prompt directory structure:
  - `prompts/sbom/security/` - Security-focused prompts
  - `prompts/sbom/operations/` - Operational SBOM management
  - `prompts/sbom/compliance/` - License and compliance prompts
- Updated skill introduction to reflect broader capabilities
- Enhanced documentation with conversion and upgrade workflows

### Migration Guide
- Old location: `prompts/security/sbom/`
- New locations:
  - Security prompts: `prompts/sbom/security/`
  - Operations prompts: `prompts/sbom/operations/`
  - Compliance prompts: `prompts/sbom/compliance/`

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
