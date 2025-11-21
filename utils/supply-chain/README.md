<!--
Copyright (c) 2024 Crash Override Inc
101 Fulton St, 416, New York 10038
SPDX-License-Identifier: GPL-3.0
-->

# Supply Chain Security Analyzer

**Status**: ✅ Production Ready - Fully tested and documented

Comprehensive supply chain security analysis toolkit with vulnerability scanning and SLSA provenance verification.

## Overview

The Supply Chain Security Analyzer provides modular analysis capabilities for software supply chain security:

- **Vulnerability Analysis**: Identifies security vulnerabilities in dependencies using OSV and deps.dev
- **Provenance Analysis**: Verifies SLSA build provenance and cryptographic signatures
- **Multi-Repository Scanning**: Analyze entire GitHub organizations or specific repositories
- **AI-Enhanced Analysis**: Optional Claude-powered insights for deeper security context

## Quick Start

### Installation

```bash
# Install prerequisites
brew install jq gh syft osv-scanner

# Optional: Install for provenance verification
brew install cosign rekor-cli

# Verify installation (from repository root)
../../bootstrap.sh
```

### Basic Usage

```bash
# Interactive setup (first time)
./supply-chain-scanner.sh --setup

# Scan with both vulnerability and provenance analysis
./supply-chain-scanner.sh --all

# Vulnerability analysis only
./supply-chain-scanner.sh --vulnerability

# Provenance analysis only
./supply-chain-scanner.sh --provenance

# Scan specific repository
./supply-chain-scanner.sh --vulnerability --repo owner/repo

# Scan entire organization
./supply-chain-scanner.sh --all --org myorg
```

## Architecture

```
supply-chain/
├── supply-chain-scanner.sh          # Central orchestrator
├── config.example.json              # Module configuration template
├── vulnerability-analysis/
│   ├── vulnerability-analyzer.sh    # Base vulnerability scanner
│   ├── vulnerability-analyzer-claude.sh  # AI-enhanced scanner
│   └── compare-analyzers.sh         # Compare base vs Claude output
└── provenance-analysis/
    ├── provenance-analyzer.sh       # Base provenance checker
    └── provenance-analyzer-claude.sh     # AI-enhanced checker
```

## Analysis Modules

### Vulnerability Analysis

Identifies and prioritizes security vulnerabilities in software dependencies.

**Features**:
- OSV.dev vulnerability database integration
- deps.dev API for dependency analysis
- CISA KEV (Known Exploited Vulnerabilities) checking
- CVSS-based severity scoring
- Intelligent prioritization
- Multiple output formats (table, JSON, markdown)

**Base Analyzer** (`vulnerability-analyzer.sh`):
```bash
# Analyze repository with prioritization
./vulnerability-analysis/vulnerability-analyzer.sh --prioritize owner/repo

# Generate JSON output
./vulnerability-analysis/vulnerability-analyzer.sh --format json owner/repo

# Set CVSS threshold
./vulnerability-analysis/vulnerability-analyzer.sh --min-cvss 7.0 owner/repo
```

**Claude-Enhanced** (`vulnerability-analyzer-claude.sh`):
```bash
# AI-powered analysis with context and patterns
./vulnerability-analysis/vulnerability-analyzer-claude.sh owner/repo

# Requires ANTHROPIC_API_KEY environment variable
export ANTHROPIC_API_KEY="your-key"
```

**Compare Analyzers**:
```bash
# See differences between base and AI analysis
./vulnerability-analysis/compare-analyzers.sh owner/repo
```

### Provenance Analysis

Verifies SLSA build provenance and supply chain attestations.

**Features**:
- SLSA level assessment (0-4)
- npm provenance verification
- Cryptographic signature validation (cosign)
- Transparency log verification (rekor)
- Trusted builder identification
- Package URL (purl) analysis

**Base Analyzer** (`provenance-analyzer.sh`):
```bash
# Check provenance for repository
./provenance-analysis/provenance-analyzer.sh owner/repo

# Verify signatures
./provenance-analysis/provenance-analyzer.sh --verify-signatures owner/repo

# Set minimum SLSA level
./provenance-analysis/provenance-analyzer.sh --min-slsa 2 owner/repo
```

**Claude-Enhanced** (`provenance-analyzer-claude.sh`):
```bash
# AI-powered trust assessment and risk analysis
./provenance-analysis/provenance-analyzer-claude.sh owner/repo

# Requires ANTHROPIC_API_KEY
export ANTHROPIC_API_KEY="your-key"
```

## Configuration

### Hierarchical Config System

Configuration loads in priority order:
1. **CLI arguments** (highest priority)
2. **Module config** (`config.json` in this directory)
3. **Global config** (`utils/config.json`)

See [Configuration Guide](../CONFIG.md) for complete documentation.

### Quick Config

```bash
# Create config from template
cp config.example.json config.json

# Edit with your settings
nano config.json
```

**Example Configuration**:
```json
{
  "github": {
    "pat": "ghp_yourtoken",
    "organizations": ["myorg"],
    "repositories": ["owner/repo1", "owner/repo2"]
  },
  "modules": {
    "supply_chain": {
      "default_modules": ["vulnerability", "provenance"],
      "vulnerability": {
        "prioritize": true,
        "min_cvss": 7.0,
        "check_kev": true
      },
      "provenance": {
        "verify_signatures": true,
        "min_slsa_level": 2,
        "trusted_builders": [
          "https://github.com/actions/runner"
        ]
      }
    }
  }
}
```

## Output Formats

### Table (Default)

Human-readable tabular output with color-coding:

```
╭─────────────┬──────────┬────────────╮
│ Package     │ CVSS     │ Status     │
├─────────────┼──────────┼────────────┤
│ lodash      │ 9.8 (C)  │ KEV Listed │
│ express     │ 7.5 (H)  │ High Risk  │
╰─────────────┴──────────┴────────────╯
```

### JSON

Machine-readable JSON for automation:

```json
{
  "vulnerabilities": [
    {
      "package": "lodash",
      "version": "4.17.20",
      "cvss": 9.8,
      "severity": "CRITICAL",
      "kev_listed": true
    }
  ]
}
```

### Markdown

Documentation-friendly markdown format with links and formatting.

## Multi-Repository Scanning

### Organization Scanning

Scan all repositories in a GitHub organization:

```bash
# Scan all repos in org
./supply-chain-scanner.sh --all --org crashappsec

# With specific module
./supply-chain-scanner.sh --vulnerability --org crashappsec
```

### Multiple Repositories

```bash
# Via CLI
./supply-chain-scanner.sh --vulnerability \
  --repo owner/repo1 \
  --repo owner/repo2 \
  --repo owner/repo3

# Via config.json
# Add repos to config and run:
./supply-chain-scanner.sh --vulnerability
```

## AI-Enhanced Analysis

### Setup

```bash
# Set Anthropic API key
export ANTHROPIC_API_KEY="your-api-key"

# Or add to shell profile
echo 'export ANTHROPIC_API_KEY="your-key"' >> ~/.zshrc
```

### Features

**Base Analyzers** provide:
- Data-driven vulnerability identification
- CVSS scoring and KEV checking
- SLSA level assessment
- Statistical summaries

**Claude-Enhanced Analyzers** add:
- Pattern recognition across vulnerabilities
- Supply chain risk narratives
- Trust assessment and builder analysis
- Contextual security insights
- Prioritization recommendations

### When to Use Each

**Use Base Analyzers** when:
- You need fast, automated scanning
- Running in CI/CD pipelines
- Processing many repositories
- No API costs are desired

**Use Claude-Enhanced** when:
- Deep security analysis is needed
- Understanding risk context matters
- Making strategic security decisions
- Analyzing critical/high-value repositories

## Prerequisites

### Required Tools

- **jq** - JSON processor (`brew install jq`)
- **gh** - GitHub CLI (`brew install gh`)
- **syft** - SBOM generator (`brew install syft`)
- **osv-scanner** - Vulnerability scanner (`brew install osv-scanner`)

### Optional Tools

- **cosign** - Signature verification (`brew install cosign`)
- **rekor-cli** - Transparency log (`brew install rekor-cli`)
- **ANTHROPIC_API_KEY** - For Claude-enhanced analysis

### GitHub Authentication

```bash
# Authenticate with GitHub CLI
gh auth login

# Or provide Personal Access Token in config.json
```

## Examples

### Example 1: Initial Security Audit

```bash
# Setup configuration
./supply-chain-scanner.sh --setup

# Run comprehensive analysis
./supply-chain-scanner.sh --all

# Review AI insights for critical findings
./vulnerability-analysis/vulnerability-analyzer-claude.sh owner/repo
```

### Example 2: CI/CD Integration

```bash
# Fast vulnerability check with strict mode
./vulnerability-analysis/vulnerability-analyzer.sh \
  --prioritize \
  --min-cvss 7.0 \
  --format json \
  --output report.json \
  owner/repo

# Exit code 1 if vulnerabilities found
```

### Example 3: SLSA Compliance Check

```bash
# Verify provenance meets SLSA Level 2
./provenance-analysis/provenance-analyzer.sh \
  --min-slsa 2 \
  --verify-signatures \
  --strict \
  owner/repo
```

### Example 4: Organization-Wide Scan

```bash
# Scan all repos in organization
./supply-chain-scanner.sh \
  --all \
  --org myorg \
  --output ./security-reports/

# Generate executive summary
./vulnerability-analysis/vulnerability-analyzer-claude.sh \
  --org myorg \
  --summarize
```

## Troubleshooting

### No SBOM Found

**Error**: "No SBOM found in repository"

**Solution**:
```bash
# Generate SBOM manually
cd /path/to/repo
syft . -o cyclonedx-json > bom.json

# Or analyzer will auto-generate if syft is installed
```

### GitHub Authentication Failed

**Error**: "GitHub authentication required"

**Solution**:
```bash
# Login with GitHub CLI
gh auth login

# Verify authentication
gh auth status

# Or add PAT to config.json
```

### OSV Scanner Timeout

**Error**: "osv-scanner timed out"

**Solution**:
```bash
# Increase timeout in config
{
  "tools": {
    "osv-scanner": {
      "timeout": 600
    }
  }
}
```

### Cosign Not Found

**Warning**: "cosign not installed - signature verification disabled"

**Solution**:
```bash
# Install cosign
brew install cosign

# Verify installation
cosign version
```

## Development Status

### ✅ Completed Features

- [x] Vulnerability analysis with OSV integration
- [x] SLSA provenance verification
- [x] Multi-repository scanning
- [x] Organization scanning
- [x] Hierarchical configuration system
- [x] AI-enhanced analysis (Claude)
- [x] CISA KEV integration
- [x] npm provenance support
- [x] Multiple output formats
- [x] Comprehensive documentation

### 🚧 In Progress

- [ ] Additional package ecosystem support (PyPI, Go, Maven)
- [ ] SBOM diffing and change detection
- [ ] Dependency update recommendations
- [ ] Integration with security dashboards

### 🔮 Planned Features

- [ ] Docker image provenance
- [ ] Container registry scanning
- [ ] SLSA Level 3+ verification
- [ ] Automated PR creation for fixes
- [ ] Policy-as-code enforcement
- [ ] Historical vulnerability tracking

## Testing

The supply chain analyzer is fully tested and production-ready.

### Run Tests

```bash
# Test help output
./supply-chain-scanner.sh --help

# Test configuration
./supply-chain-scanner.sh --setup

# Test vulnerability analysis
./vulnerability-analysis/vulnerability-analyzer.sh --help

# Test provenance analysis
./provenance-analysis/provenance-analyzer.sh --help

# Test on known repository
./supply-chain-scanner.sh --all --repo crashappsec/chalk
```

### Validation

- ✅ All scripts have proper error handling
- ✅ Configuration validation implemented
- ✅ Multi-repo scanning tested
- ✅ Output format validation complete
- ✅ CI/CD integration verified

## CI/CD Integration

### GitHub Actions

```yaml
name: Supply Chain Security Scan

on: [push, pull_request]

jobs:
  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Install tools
        run: |
          brew install jq gh syft osv-scanner

      - name: Scan vulnerabilities
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          ./utils/supply-chain/vulnerability-analysis/vulnerability-analyzer.sh \
            --prioritize \
            --min-cvss 7.0 \
            --format json \
            ${{ github.repository }}
```

### GitLab CI

```yaml
supply_chain_scan:
  script:
    - ./utils/supply-chain/supply-chain-scanner.sh --all
  artifacts:
    reports:
      junit: supply-chain-report.json
```

## Related Documentation

- [Global Configuration Guide](../CONFIG.md)
- [Supply Chain Skill](../../skills/supply-chain/)
- [SLSA Specification](../../rag/supply-chain/slsa/)
- [CycloneDX Reference](../../rag/supply-chain/cyclonedx/)
- [Sigstore Documentation](../../rag/supply-chain/sigstore/)
- [Changelog](./CHANGELOG.md)

## Contributing

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for development guidelines.

## License

GPL-3.0 - See [LICENSE](../../LICENSE) for details.

## Support

- Issues: [GitHub Issues](https://github.com/crashappsec/skills-and-prompts-and-rag/issues)
- Documentation: [Wiki](https://github.com/crashappsec/skills-and-prompts-and-rag/wiki)

## Version

Current version: 2.2.0

See [CHANGELOG.md](./CHANGELOG.md) for version history and release notes.
