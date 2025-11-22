<!--
Copyright (c) 2024 Gibson Powers Contributors

SPDX-License-Identifier: GPL-3.0
-->

# Package Health Analyzer

**Status**: 🔬 Experimental v1.0.1 - Core functionality stable, ready for testing

Comprehensive package health analysis tools for identifying risks and operational improvement opportunities across an organization's software dependencies.

**Current Version**: 1.0.1 (Bug fixes - now functional!)

## Overview

The Package Health Analyzer provides two-tiered analysis of software packages:

1. **Base Analyzer**: Fast, automated scanning with health scoring, deprecation detection, and version analysis
2. **AI-Enhanced Analyzer**: Deep analysis with Claude AI, providing contextual recommendations, migration strategies, and risk prioritization

### Key Features

- ✅ **Health Scoring**: Composite 0-100 score based on OpenSSF Scorecard, maintenance, security, freshness, and popularity
- ✅ **Deprecation Detection**: Identifies deprecated packages with suggested alternatives
- ✅ **Version Standardization**: Finds version inconsistencies across repositories
- ✅ **Chain of Reasoning**: Integrates vulnerability and provenance analysis
- ✅ **AI-Powered Recommendations**: Detailed migration guides and strategic insights
- ✅ **Multi-Ecosystem Support**: npm, PyPI, Maven, Cargo, Go
- ✅ **Organization-Wide Analysis**: Scan entire GitHub organizations

## Installation

### Prerequisites

```bash
# Required
- bash 4.0+
- jq 1.6+
- curl
- git

# For repository scanning
- syft - https://github.com/anchore/syft

# For AI-enhanced analysis
- ANTHROPIC_API_KEY in .env file (see setup below)

# NOT Required (common misconception)
- gh (GitHub CLI) - NO LONGER NEEDED! Uses standard git clone
```

### Setup

```bash
# 1. Clone repository
git clone https://github.com/crashappsec/skills-and-prompts-and-rag.git
cd skills-and-prompts-and-rag

# 2. Run bootstrap to set everything up
./bootstrap.sh

# 3. Set up API key for AI analysis (automatically loaded)
cp .env.example .env
# Edit .env and add your ANTHROPIC_API_KEY

# 4. Configure health scoring weights (optional)
cp utils/supply-chain/package-health-analysis/config.example.json \
   utils/supply-chain/config.json
# Edit config.json to customize health score weights
```

## Usage

### Quick Start

```bash
# Analyze single repository
./utils/supply-chain/package-health-analysis/package-health-analyzer.sh \
  --repo owner/repo

# AI-enhanced analysis
./utils/supply-chain/package-health-analysis/package-health-analyzer-claude.sh \
  --repo owner/repo \
  --output health-report.md

# Organization-wide scan
./utils/supply-chain/package-health-analysis/package-health-analyzer-claude.sh \
  --org myorg \
  --output org-health.md
```

### Base Analyzer

**Purpose**: Fast automated scanning for CI/CD and regular monitoring.

```bash
# Basic usage
./package-health-analyzer.sh --repo owner/repo

# Custom output format
./package-health-analyzer.sh --repo owner/repo --format markdown

# Analyze existing SBOM
./package-health-analyzer.sh --sbom path/to/sbom.json

# Organization scan
./package-health-analyzer.sh --org myorg --output org-scan.json

# Skip specific analyses
./package-health-analyzer.sh --repo owner/repo \
  --no-version-analysis \
  --no-deprecation-check
```

**Output Example** (JSON):
```json
{
  "scan_metadata": {
    "timestamp": "2024-11-21T10:30:00Z",
    "repositories_scanned": 1,
    "packages_analyzed": 42,
    "analyzer_version": "1.0.0"
  },
  "summary": {
    "total_packages": 42,
    "deprecated_packages": 3,
    "low_health_packages": 5,
    "version_inconsistencies": 0
  },
  "packages": [
    {
      "package": "request",
      "system": "npm",
      "version": "2.88.0",
      "health_score": 25,
      "health_grade": "Critical",
      "deprecated": true,
      "deprecation_message": "Package no longer supported",
      "component_scores": {
        "openssf": 0,
        "maintenance": 0,
        "security": 50,
        "freshness": 10,
        "popularity": 80
      }
    }
  ]
}
```

### AI-Enhanced Analyzer

**Purpose**: Comprehensive analysis with recommendations and strategic insights.

```bash
# Full analysis with recommendations
./package-health-analyzer-claude.sh --repo owner/repo

# Organization-wide with all analyses
./package-health-analyzer-claude.sh --org myorg \
  --output comprehensive-report.md

# Quick mode (skip provenance check)
./package-health-analyzer-claude.sh --repo owner/repo \
  --skip-prov-analysis

# JSON output
./package-health-analyzer-claude.sh --repo owner/repo \
  --format json \
  --output analysis.json
```

**Output Example** (Markdown):
```markdown
# Package Health Analysis Report (AI-Enhanced)

## Executive Summary

Analysis identified 3 critical issues requiring immediate attention:
- 'request' package deprecated in 5 repositories (high security risk)
- Version inconsistencies in 'lodash' (12 repos, 3 versions)
- Low health score packages affecting core functionality

## Risk Rankings

| Package  | Risk Level | Impact | Urgency | Affected Repos |
|----------|-----------|--------|---------|----------------|
| request  | Critical  | High   | Immediate | 5 |
| moment   | High      | Medium | 30 days | 8 |
| lodash   | Medium    | Low    | 90 days | 12 |

## Detailed Findings

### 1. Deprecated Package: request

**Risk**: Critical - No security updates, known vulnerabilities

**Recommended Alternatives**:
1. **axios** (Recommended)
   - Modern Promise-based API
   - Better error handling
   - Migration effort: 2-3 days per repo

2. **node-fetch**
   - Lightweight
   - Fetch API compatible
   - Migration effort: 1-2 days per repo

**Migration Guide**:
```javascript
// Before (request)
request.get('https://api.example.com', (err, res, body) => {
  if (err) return console.error(err);
  console.log(body);
});

// After (axios)
const axios = require('axios');
try {
  const response = await axios.get('https://api.example.com');
  console.log(response.data);
} catch (error) {
  console.error(error);
}
```

[... continues with detailed analysis ...]
```

### Comparison Tool

Compare base vs AI-enhanced analyzers:

```bash
./compare-analyzers.sh --repo owner/repo --output comparison.md
```

## Health Scoring Algorithm

Composite health score (0-100) based on weighted components:

```
Health Score =
  (OpenSSF Score × 0.30) +     # Security practices
  (Maintenance Score × 0.25) +  # Active development
  (Security Score × 0.25) +     # Vulnerability status
  (Freshness Score × 0.10) +    # Version currency
  (Popularity Score × 0.10)     # Community adoption
```

### Grade Thresholds

| Grade     | Score Range | Meaning |
|-----------|-------------|---------|
| Excellent | 90-100      | High quality, well-maintained |
| Good      | 75-89       | Solid choice, minor concerns |
| Fair      | 60-74       | Acceptable, monitor closely |
| Poor      | 40-59       | Consider alternatives |
| Critical  | 0-39        | Replace urgently |

## Chain of Reasoning

The AI-enhanced analyzer orchestrates multiple tools:

```
1. SBOM Generation (if needed)
   ↓
2. Base Package Health Analysis
   ↓
3. Vulnerability Analysis
   ↓
4. Provenance Analysis (optional)
   ↓
5. Context Preparation
   ↓
6. AI Analysis & Recommendations
```

Each stage builds on the previous, providing comprehensive context for Claude AI to generate actionable insights.

## Configuration

Configuration file: `config.example.json`

```json
{
  "package_health": {
    "health_score_weights": {
      "openssf": 0.30,
      "maintenance": 0.25,
      "security": 0.25,
      "freshness": 0.10,
      "popularity": 0.10
    },
    "thresholds": {
      "excellent": 90,
      "good": 75,
      "fair": 60,
      "poor": 40
    },
    "api": {
      "deps_dev_base_url": "https://api.deps.dev/v3alpha",
      "timeout": 30,
      "retry_attempts": 3
    },
    "cache": {
      "enabled": true,
      "ttl_hours": 24
    }
  }
}
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Package Health Check

on:
  schedule:
    - cron: '0 0 * * 0'  # Weekly
  workflow_dispatch:

jobs:
  health-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2

      - name: Install dependencies
        run: |
          # Install syft, gh, jq
          brew install syft gh jq

      - name: Run Package Health Analysis
        env:
          ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
        run: |
          ./utils/supply-chain/package-health-analysis/package-health-analyzer.sh \
            --repo ${{ github.repository }} \
            --format json > health-report.json

      - name: Check for Critical Issues
        run: |
          CRITICAL=$(jq '.summary.critical_health_packages' health-report.json)
          if [ "$CRITICAL" -gt 0 ]; then
            echo "::error::Found $CRITICAL critical package health issues"
            exit 1
          fi

      - name: Upload Report
        uses: actions/upload-artifact@v2
        with:
          name: health-report
          path: health-report.json
```

## Use Cases

### 1. Security Audits
Identify deprecated packages with security vulnerabilities:
```bash
./package-health-analyzer-claude.sh --org myorg | \
  grep -A 5 "deprecated.*true"
```

### 2. Version Standardization
Find and fix version inconsistencies:
```bash
./package-health-analyzer.sh --org myorg --format json | \
  jq '.version_inconsistencies[]'
```

### 3. Tech Debt Reduction
Prioritize package improvements:
```bash
./package-health-analyzer-claude.sh --repo owner/repo \
  --output tech-debt-plan.md
```

### 4. Pre-Release Validation
Ensure healthy dependencies before major releases:
```bash
./package-health-analyzer.sh --repo owner/repo | \
  jq '.packages[] | select(.health_score < 60)'
```

## Architecture

```
package-health-analysis/
├── package-health-analyzer.sh          # Base scanner
├── package-health-analyzer-claude.sh   # AI-enhanced
├── compare-analyzers.sh                # Comparison tool
├── lib/
│   ├── deps-dev-client.sh              # API client
│   ├── health-scoring.sh               # Scoring engine
│   ├── version-analysis.sh             # Version analysis
│   └── deprecation-checker.sh          # Deprecation detection
├── config.example.json                 # Configuration
├── README.md                           # This file
└── CHANGELOG.md                        # Version history
```

## API Integration

### deps.dev API

The analyzer integrates with [deps.dev](https://deps.dev) for:
- OpenSSF Scorecard data
- Package metadata
- Deprecation status
- Dependency information

**Rate Limits**: Cached to minimize API calls (24-hour TTL by default).

## Troubleshooting

### Common Issues

**Error: "syft not found"**
```bash
# Install syft
brew install syft
# or
curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b /usr/local/bin
```

**Error: "ANTHROPIC_API_KEY not set"**
```bash
export ANTHROPIC_API_KEY="your-api-key-here"
```

**Error: "gh not found"**
```bash
# Install GitHub CLI
brew install gh
# or download from https://cli.github.com/

# Authenticate
gh auth login
```

**Slow Performance**
- Enable caching in config
- Use base analyzer for quick scans
- Reduce scope (single repo vs org)

## Performance

### Benchmarks

| Scan Type | Packages | Base Time | AI Time |
|-----------|----------|-----------|---------|
| Small Repo | 10-20 | 30s | 2-3 min |
| Medium Repo | 50-100 | 2 min | 5-8 min |
| Large Repo | 200+ | 5 min | 15-20 min |
| Organization (10 repos) | 500+ | 10 min | 30-45 min |

*Times are approximate and vary based on API response times*

## Limitations

- **Ecosystem Coverage**: Best support for npm and PyPI; limited for others
- **API Rate Limits**: deps.dev may rate limit frequent requests
- **Accuracy**: Health scoring is heuristic-based, not definitive
- **Analysis Depth**: Base analyzer provides data; AI analyzer provides insight

## Contributing

See [CONTRIBUTING.md](../../../CONTRIBUTING.md) for guidelines.

## License

GPL-3.0 - See [LICENSE](../../../LICENSE) for details.

## References

- [Build Prompt](../../../prompts/supply-chain/BUILD-PACKAGE-HEALTH-ANALYZER.md)
- [Requirements](../../../prompts/supply-chain/package-health-analyzer-requirements.md)
- [deps.dev API](../../../rag/supply-chain/package-health/deps-dev-api.md)
- [Best Practices](../../../rag/supply-chain/package-health/package-management-best-practices.md)
- [Supply Chain Analyzer](../README.md)

## Support

- [Issues](https://github.com/crashappsec/skills-and-prompts-and-rag/issues)
- [Discussions](https://github.com/crashappsec/skills-and-prompts-and-rag/discussions)

---

**Made with ❤️ by the Crash Override community**
