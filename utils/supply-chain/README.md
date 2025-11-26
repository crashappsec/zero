<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Supply Chain Security Analyser

**Status**: 🚀 Beta | **Version**: 3.1.0

Comprehensive supply chain security analysis toolkit with vulnerability scanning, SLSA provenance verification, package health analysis, and AI-powered insights.

Feature-complete with 9 analysis modules covering security, maintainability, and supply chain risk.

## Overview

The Supply Chain Security Analyser provides modular analysis capabilities for software supply chain security:

### Security Modules
- **Vulnerability Analysis** (`--vulnerability`): Identifies security vulnerabilities using OSV and deps.dev
- **Provenance Analysis** (`--provenance`): Verifies SLSA build provenance and cryptographic signatures
- **Typosquatting Detection** (`--typosquat`): Detects potential typosquatting attacks on dependencies

### Package Health Modules
- **Abandoned Package Detection** (`--abandoned`): Identifies unmaintained packages with security risks
- **Unused Dependency Analysis** (`--unused`): Finds dead code dependencies for removal
- **Technical Debt Scoring** (`--debt-score`): Quantifies dependency technical debt

### Developer Productivity Modules
- **Library Recommendations** (`--library-recommend`): Suggests modern alternatives for outdated packages
- **Container Image Analysis** (`--container-images`): Recommends secure base images (distroless, Chainguard)

### Cross-Cutting Features
- **Multi-Repository Scanning**: Analyze entire GitHub organizations or specific repositories
- **AI-Enhanced Analysis**: Claude-powered unified insights across all modules

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

# Run all security modules (vulnerability + provenance)
./supply-chain-scanner.sh --all --repo owner/repo

# Individual security modules
./supply-chain-scanner.sh --vulnerability --repo owner/repo
./supply-chain-scanner.sh --provenance --repo owner/repo
./supply-chain-scanner.sh --typosquat --repo owner/repo

# Package health modules
./supply-chain-scanner.sh --abandoned --repo owner/repo
./supply-chain-scanner.sh --unused --repo owner/repo
./supply-chain-scanner.sh --debt-score --repo owner/repo

# Developer productivity modules
./supply-chain-scanner.sh --library-recommend --repo owner/repo
./supply-chain-scanner.sh --container-images --repo owner/repo

# Combine multiple modules
./supply-chain-scanner.sh --abandoned --typosquat --debt-score --repo owner/repo

# Scan entire organization
./supply-chain-scanner.sh --all --org myorg

# With Claude AI enhancement (auto-enabled when ANTHROPIC_API_KEY is set)
export ANTHROPIC_API_KEY="your-api-key"
./supply-chain-scanner.sh --all --repo owner/repo
```

### Test Organization

Test with the [Gibson Powers Test Organization](https://github.com/Gibson-Powers-Test-Org):

```bash
# Test vulnerability analysis
./vulnerability-analysis/vulnerability-analyser.sh --org Gibson-Powers-Test-Org

# Test with Claude AI analysis
export ANTHROPIC_API_KEY="your-key"
./vulnerability-analysis/vulnerability-analyser.sh --claude --org Gibson-Powers-Test-Org

# Test provenance analysis
./provenance-analysis/provenance-analyser.sh --org Gibson-Powers-Test-Org
```

## Architecture

```
supply-chain/
├── supply-chain-scanner.sh              # Central orchestrator
├── config.example.json                  # Module configuration template
├── lib/
│   └── deps-dev-client.sh               # Global deps.dev API client
├── vulnerability-analysis/
│   └── vulnerability-analyser.sh        # Security vulnerability scanning
├── provenance-analysis/
│   └── provenance-analyser.sh           # SLSA provenance verification
├── package-health-analysis/
│   ├── package-health-analyser.sh       # Package health orchestrator
│   └── lib/
│       ├── abandonment-detector.sh      # Abandoned package detection
│       ├── typosquat-detector.sh        # Typosquatting risk detection
│       └── unused-detector.sh           # Unused dependency detection
├── bundle-analysis/
│   ├── bundle-analyzer.sh               # Bundle size analysis
│   └── lib/
│       └── debt-scorer.sh               # Technical debt scoring
├── library-recommendations/
│   ├── lib-recommend-analyser.sh        # Library recommendation engine
│   └── lib/
│       └── recommender.sh               # Recommendation algorithms
└── container-recommendations/
    ├── container-image-analyser.sh      # Container image analysis
    └── lib/
        └── image-recommender.sh         # Image recommendation engine
```

All modules support:
- **Standalone mode**: Run directly via `--module` flag
- **Combined mode**: Run multiple modules together
- **Claude mode**: AI-enhanced unified analysis when `ANTHROPIC_API_KEY` is set

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

**Usage**:
```bash
# Basic analysis (no API costs)
./vulnerability-analysis/vulnerability-analyser.sh --prioritize owner/repo

# AI-enhanced analysis with Claude
export ANTHROPIC_API_KEY="your-key"
./vulnerability-analysis/vulnerability-analyser.sh --claude --prioritize owner/repo

# Scan entire organization
./vulnerability-analysis/vulnerability-analyser.sh --claude --org myorg

# Generate JSON output
./vulnerability-analysis/vulnerability-analyser.sh --format json owner/repo

# All options
./vulnerability-analysis/vulnerability-analyser.sh --help
```

**Arguments**:
- `--org ORG`: Scan all repositories in GitHub organization
- `--repo OWNER/REPO`: Scan specific repository
- `--claude`: Enable AI-enhanced analysis (requires ANTHROPIC_API_KEY)
- `-t, --taint-analysis`: Enable call graph/taint analysis (Go projects)
- `-p, --prioritize`: Add intelligent prioritization (CISA KEV, CVSS)
- `-f, --format FORMAT`: Output format (table|json|markdown|sarif)
- `-o, --output FILE`: Write results to file
- `-h, --help`: Show help message

### Provenance Analysis

Verifies SLSA build provenance and supply chain attestations.

**Features**:
- SLSA level assessment (0-4)
- npm provenance verification
- Cryptographic signature validation (cosign)
- Transparency log verification (rekor)
- Trusted builder identification
- Package URL (purl) analysis

**Usage**:
```bash
# Basic SLSA provenance analysis
./provenance-analysis/provenance-analyser.sh owner/repo

# AI-enhanced analysis with Claude
export ANTHROPIC_API_KEY="your-key"
./provenance-analysis/provenance-analyser.sh --claude owner/repo

# Verify cryptographic signatures
./provenance-analysis/provenance-analyser.sh --verify-signatures owner/repo

# Require minimum SLSA level
./provenance-analysis/provenance-analyser.sh --min-level 2 --strict owner/repo

# Scan entire organization
./provenance-analysis/provenance-analyser.sh --claude --org myorg

# All options
./provenance-analysis/provenance-analyser.sh --help
```

**Arguments**:
- `--org ORG`: Scan all repositories in GitHub organization
- `--repo OWNER/REPO`: Scan specific repository
- `--claude`: Enable AI-enhanced analysis (requires ANTHROPIC_API_KEY)
- `--verify-signatures`: Cryptographically verify signatures (requires cosign)
- `--min-level LEVEL`: Require minimum SLSA level (0-4)
- `--strict`: Fail on missing provenance or low SLSA level
- `-f, --format FORMAT`: Output format (table|json|markdown)
- `-o, --output FILE`: Write results to file
- `-h, --help`: Show help message

### Typosquatting Detection

Detects potential typosquatting attacks on dependencies.

**Features**:
- Levenshtein distance analysis for similar package names
- Detection of common typosquatting patterns (character swaps, omissions, additions)
- Popular package similarity checking
- Registry-specific analysis (npm, PyPI, Go)

**Usage**:
```bash
# Detect typosquatting risks via main scanner
./supply-chain-scanner.sh --typosquat --repo owner/repo

# Scan multiple repositories
./supply-chain-scanner.sh --typosquat --org myorg
```

**Risk Indicators**:
- Edit distance ≤ 2 from popular packages
- Common typo patterns (lodahs → lodash)
- Scoped package impersonation (@loadsh/core)

### Abandoned Package Detection

Identifies packages that are no longer actively maintained.

**Features**:
- Last update date analysis (deps.dev API)
- OpenSSF Scorecard "Maintained" check integration
- Deprecated package detection
- Archived repository detection
- Risk level scoring (healthy, stale, abandoned, deprecated, archived)

**Usage**:
```bash
# Detect abandoned packages via main scanner
./supply-chain-scanner.sh --abandoned --repo owner/repo

# Combined with other health checks
./supply-chain-scanner.sh --abandoned --typosquat --repo owner/repo
```

**Thresholds**:
| Days Since Update | Status | Risk Level |
|-------------------|--------|------------|
| < 180 | Active | Low |
| 180-365 | Warning | Medium |
| 365-730 | Stale | High |
| > 730 | Abandoned | Critical |
| Archived repo | Archived | Critical |

### Unused Dependency Analysis

Finds dead code dependencies that can be safely removed.

**Features**:
- Import/require pattern analysis
- Call graph analysis (when available)
- Cross-reference with SBOM
- Safe-to-remove confidence scoring

**Usage**:
```bash
# Detect unused dependencies via main scanner
./supply-chain-scanner.sh --unused --repo owner/repo

# Combine with debt scoring
./supply-chain-scanner.sh --unused --debt-score --repo owner/repo
```

**Benefits**:
- Reduce attack surface by removing unused packages
- Decrease build times and bundle sizes
- Simplify dependency management

### Technical Debt Scoring

Quantifies dependency technical debt using weighted factors.

**Features**:
- Multi-factor scoring (abandonment, deprecation, security, outdated versions)
- OpenSSF Scorecard integration for maintenance scoring
- Replacement availability checking
- Project-level aggregation
- Debt reduction roadmap generation

**Usage**:
```bash
# Calculate technical debt via main scanner
./supply-chain-scanner.sh --debt-score --repo owner/repo

# Get debt reduction roadmap
./supply-chain-scanner.sh --debt-score --library-recommend --repo owner/repo
```

**Score Ranges**:
| Score | Level | Action Required |
|-------|-------|-----------------|
| 0-20 | Low | Monitor normally |
| 21-40 | Medium | Plan future review |
| 41-60 | High | Address in next sprint |
| 61-100 | Critical | Immediate action |

### Library Recommendations

Suggests modern alternatives for outdated or deprecated packages.

**Features**:
- Deprecated package replacement suggestions
- Modern alternative recommendations
- Migration effort estimation (trivial, easy, moderate, significant, major)
- API compatibility analysis
- Community adoption metrics

**Usage**:
```bash
# Get library recommendations via main scanner
./supply-chain-scanner.sh --library-recommend --repo owner/repo

# Combined with debt analysis
./supply-chain-scanner.sh --debt-score --library-recommend --repo owner/repo
```

**Example Recommendations**:
| Package | Status | Replacement | Migration Effort |
|---------|--------|-------------|------------------|
| request | Deprecated | axios, got | Easy |
| moment | Deprecated | date-fns, luxon | Moderate |
| underscore | Stale | lodash | Easy |

### Container Image Analysis

Recommends secure base images for containerized applications.

**Features**:
- Dockerfile analysis
- Base image security assessment
- Distroless image recommendations
- Chainguard image recommendations
- Alpine image recommendations
- Size and security tradeoff analysis

**Usage**:
```bash
# Analyze container images via main scanner
./supply-chain-scanner.sh --container-images --repo owner/repo

# Requires Dockerfile in repository
```

**Recommendations**:
| Current Image | Recommended | Rationale |
|---------------|-------------|-----------|
| node:18 | gcr.io/distroless/nodejs18-debian11 | Minimal attack surface |
| python:3.11 | cgr.dev/chainguard/python | Supply chain verified |
| ubuntu:22.04 | alpine:3.18 | Smaller footprint |

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

**Org-Wide Claude Analysis**: When using `--org` with Claude enabled, the AI analysis runs **once** after ALL repositories are scanned, providing strategic, team-level recommendations instead of per-repo analysis:

```bash
# Org-wide strategic analysis
export ANTHROPIC_API_KEY="your-key"
./supply-chain-scanner.sh --vulnerability --package-health --org myorg

# Output:
# ├── Repo 1 scanned → results collected
# ├── Repo 2 scanned → results collected
# ├── Repo N scanned → results collected
# └── 🏢 Organization-Wide Claude AI Analysis
#     ├── Portfolio Health Dashboard
#     ├── Systemic Issues (vulns in multiple repos)
#     ├── Repository Prioritization Matrix
#     └── Strategic Recommendations for Team
```

**What Org-Wide Analysis Provides**:
- **Portfolio Health Dashboard**: Aggregate metrics across all repos
- **Systemic Issues**: Vulnerabilities appearing in multiple repositories
- **Repository Prioritization**: Risk-ranked list of repos for remediation
- **Strategic Recommendations**: Team-level initiatives vs per-repo fixes
- **Automation Opportunities**: Org-wide CI/CD and policy improvements

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

**Claude AI is now enabled by default** when you set your API key. No `--claude` flag needed!

```bash
# Set Anthropic API key to enable Claude AI automatically
export ANTHROPIC_API_KEY="your-api-key"

# Or add to shell profile for persistence
echo 'export ANTHROPIC_API_KEY="your-key"' >> ~/.zshrc

# Get your API key at: https://console.anthropic.com/settings/keys
```

**How it works**:
```bash
# With API key set - Claude runs automatically
export ANTHROPIC_API_KEY="your-key"
./supply-chain-scanner.sh --all --repo owner/repo

# Output:
#   🤖 Claude AI: ENABLED (analyzing results with AI)
#   [standard scans run...]
#   [Claude AI Enhanced Analysis runs last using all results]

# Without API key - standard analysis only
unset ANTHROPIC_API_KEY
./supply-chain-scanner.sh --all --repo owner/repo

# Output:
#   ℹ️  Claude AI: DISABLED (no API key found)
#   [standard scans run...]
```

**Execution Order**:
1. ✅ API key check (first thing)
2. ✅ Standard security scans (vulnerability, provenance, package health)
3. ✅ Claude AI analysis (LAST - analyzes all previous results)

### Features

**Base Analysers** provide:
- Data-driven vulnerability identification
- CVSS scoring and KEV checking
- SLSA level assessment
- Statistical summaries

**Claude-Enhanced Analysers** add:
- Pattern recognition across vulnerabilities
- Supply chain risk narratives
- Trust assessment and builder analysis
- Contextual security insights
- Prioritization recommendations

### When to Use Each

**Use Base Analysers** when:
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
./vulnerability-analysis/vulnerability-analyser-claude.sh owner/repo
```

### Example 2: CI/CD Integration

```bash
# Fast vulnerability check with strict mode
./vulnerability-analysis/vulnerability-analyser.sh \
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
./provenance-analysis/provenance-analyser.sh \
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
./vulnerability-analysis/vulnerability-analyser-claude.sh \
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

# Or analyser will auto-generate if syft is installed
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

**Current Status**: 🚀 Beta | **Version**: 3.1.0

### Completed Features (v3.0.0)

#### Security Modules
- [x] Vulnerability analysis with OSV integration
- [x] SLSA provenance verification
- [x] **Typosquatting detection** (NEW)
- [x] CISA KEV integration
- [x] npm provenance support

#### Package Health Modules
- [x] **Abandoned package detection** (NEW)
- [x] **Unused dependency analysis** (NEW)
- [x] **Technical debt scoring** (NEW)
- [x] OpenSSF Scorecard integration
- [x] deps.dev API integration

#### Developer Productivity Modules
- [x] **Library recommendations** (NEW)
- [x] **Container image analysis** (NEW)
- [x] Migration effort estimation
- [x] Alternative package suggestions

#### Infrastructure
- [x] Multi-repository scanning
- [x] Organization scanning
- [x] Hierarchical configuration system
- [x] AI-enhanced unified analysis (Claude)
- [x] Multiple output formats (table, JSON, markdown)
- [x] Comprehensive RAG knowledge base
- [x] Production testing completed

### 🚧 In Progress

- [ ] Technology Identification System (Phase 2: Implementation)
- [ ] Additional package ecosystem support (PyPI, Go, Maven)
- [ ] SBOM diffing and change detection
- [ ] Integration with security dashboards

### 🔮 Planned Features

#### High Priority

- [ ] **Enhanced Taint Analysis Integration**
  - **Purpose**: Determine if vulnerabilities are actually exploitable
  - **Current**: Basic unused detection implemented
  - **Future**:
    - `osv-scanner --call-analysis=all` for call graph analysis
    - Cross-reference with SBOM to identify unused packages
    - Generate "unused dependency" report
    - Suggest safe-to-remove packages with confidence scores
  - **Output**:
    - Vulnerabilities: CALLED | NOT_CALLED | UNKNOWN
    - Dependencies: USED | UNUSED | UNKNOWN
    - Recommendations: Safe to remove packages

#### Standard Features

- [ ] Docker image provenance
- [ ] Container registry scanning
- [ ] SLSA Level 3+ verification
- [ ] Automated PR creation for fixes
- [ ] Policy-as-code enforcement
- [ ] Historical vulnerability tracking

## Testing

The supply chain analyser is comprehensively tested and ready for Beta use.

### Run Tests

```bash
# Test help output
./supply-chain-scanner.sh --help

# Test configuration
./supply-chain-scanner.sh --setup

# Test vulnerability analysis
./vulnerability-analysis/vulnerability-analyser.sh --help

# Test provenance analysis
./provenance-analysis/provenance-analyser.sh --help

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
          ./utils/supply-chain/vulnerability-analysis/vulnerability-analyser.sh \
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

Current version: 3.1.0

See [CHANGELOG.md](./CHANGELOG.md) for version history and release notes.
