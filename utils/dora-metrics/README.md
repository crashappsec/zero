<!--
Copyright (c) 2024 Gibson Powers Contributors

SPDX-License-Identifier: GPL-3.0
-->

# DORA Metrics Analyzer

**Status**: 🔬 Experimental

Calculates and analyzes DevOps Research and Assessment (DORA) metrics for software delivery performance.

## ⚠️ Development Status

This utility is in **early development** and is not yet ready for Beta or production use. It provides basic DORA metrics calculation but lacks the comprehensive testing, documentation, and features of the Beta supply chain analyzer.

### What Works
- ✅ Basic DORA metrics calculation (Deployment Frequency, Lead Time, MTTR, Change Failure Rate)
- ✅ Performance classification (Elite/High/Medium/Low)
- ✅ AI-enhanced analysis with Claude (--claude flag)
- ✅ Cost tracking for API usage
- ✅ Unified tool (single binary, dual modes)

### What's Missing
- ❌ Comprehensive error handling
- ❌ Configuration system integration
- ❌ Multi-repository scanning
- ❌ Output format options
- ❌ Extensive testing
- ❌ Complete documentation

**Use at your own risk**. For Beta-quality analysis, use the [Supply Chain Security Analyzer](../supply-chain/).

## Overview

DORA metrics are four key metrics that indicate the performance of software delivery:

1. **Deployment Frequency**: How often code is deployed to production
2. **Lead Time for Changes**: Time from commit to production
3. **Mean Time to Recovery (MTTR)**: Time to recover from failures
4. **Change Failure Rate**: Percentage of deployments causing failures

## Quick Start

### Prerequisites

```bash
# Install GitHub CLI
brew install gh

# Authenticate
gh auth login
```

### Basic Usage

```bash
# Basic analysis (no API key required)
./dora-analyzer.sh deployment-data.json

# AI-enhanced analysis with insights and cost tracking
export ANTHROPIC_API_KEY="your-key"
./dora-analyzer.sh --claude deployment-data.json
./dora-analyzer-claude.sh owner/repo

# Compare base vs Claude analysis
./compare-analyzers.sh owner/repo
```

## Available Scripts

### dora-analyzer.sh

Base analyzer that calculates DORA metrics from Git history and GitHub API.

**Features**:
- Calculates all four DORA metrics
- Performance classification
- Benchmark comparison
- Trend analysis

**Usage**:
```bash
./dora-analyzer.sh <owner>/<repo>
```

**Output**:
```
===================================
DORA Metrics Report
===================================
Repository: owner/repo
Analysis Period: Last 90 days

Deployment Frequency: 15.3 per day (Elite)
Lead Time for Changes: 2.3 hours (Elite)
Mean Time to Recovery: 0.8 hours (Elite)
Change Failure Rate: 5.2% (Elite)

Overall Performance: Elite
```

### dora-analyzer-claude.sh

AI-enhanced analyzer that provides contextual insights and recommendations.

**Features**:
- All base analyzer features
- Pattern recognition
- Contextual insights
- Performance recommendations
- Trend interpretation

**Requires**: `ANTHROPIC_API_KEY` environment variable

**Usage**:
```bash
export ANTHROPIC_API_KEY="your-key"
./dora-analyzer-claude.sh owner/repo
```

### compare-analyzers.sh

Compare base and AI-enhanced analysis side-by-side.

**Usage**:
```bash
./compare-analyzers.sh owner/repo
```

## DORA Performance Levels

| Level | Deployment Freq | Lead Time | MTTR | Change Failure Rate |
|-------|----------------|-----------|------|---------------------|
| **Elite** | On-demand (multiple per day) | < 1 hour | < 1 hour | < 5% |
| **High** | Between once per day and once per week | < 1 day | < 1 day | < 10% |
| **Medium** | Between once per week and once per month | < 1 week | < 1 day | < 15% |
| **Low** | Between once per month and once every 6 months | > 1 week | > 1 day | > 15% |

## Known Limitations

### Current Limitations

1. **Single Repository Only**: Does not support multi-repo or organization scanning
2. **No Configuration System**: Cannot persist settings or use global configs
3. **Limited Output Formats**: Only text output, no JSON/markdown
4. **Basic Error Handling**: May fail unexpectedly on edge cases
5. **No Time Range Selection**: Fixed 90-day analysis window
6. **Incomplete Metrics**: Some calculations may not match all environments

### Data Source Limitations

- Relies on GitHub API and Git history
- May not capture all deployment events
- Requires proper tagging/releases for deployment tracking
- MTTR calculation assumes incident tracking in issues
- Change failure rate depends on issue labeling

## Roadmap to Production

### Phase 1: Core Functionality (Current)
- [x] Basic DORA metrics calculation
- [x] AI-enhanced analysis
- [ ] Comprehensive error handling
- [ ] Input validation

### Phase 2: Integration
- [ ] Hierarchical configuration system
- [ ] Multi-repository support
- [ ] Organization scanning
- [ ] Output format options (JSON, markdown, CSV)

### Phase 3: Testing & Documentation
- [ ] Unit tests
- [ ] Integration tests
- [ ] Comprehensive README
- [ ] Usage examples
- [ ] Troubleshooting guide

### Phase 4: Production Ready
- [ ] CI/CD integration examples
- [ ] Dashboard integration
- [ ] Historical tracking
- [ ] Alerting/notifications
- [ ] Performance optimization

## Development

### Architecture

```
dora-metrics/
├── dora-analyzer.sh              # Base analyzer
├── dora-analyzer-claude.sh       # AI-enhanced analyzer
└── compare-analyzers.sh          # Comparison tool
```

### Adding Features

This utility needs significant development before production use. Key areas:

1. **Configuration Integration**: Add support for global config system
2. **Multi-Repo Support**: Enable organization and batch scanning
3. **Error Handling**: Comprehensive validation and error messages
4. **Output Formats**: JSON, markdown, and custom formats
5. **Testing**: Unit and integration test suite
6. **Documentation**: Complete usage guide and examples

## Related Documentation

- [DORA Skill](../../skills/dora-metrics/)
- [DORA Research](https://dora.dev/)
- [Changelog](./CHANGELOG.md)

## Contributing

Contributions welcome! This utility needs significant work to reach production quality. See [CONTRIBUTING.md](../../CONTRIBUTING.md).

Priority areas:
- Configuration system integration
- Multi-repository support
- Comprehensive testing
- Error handling improvements
- Output format options

## License

GPL-3.0 - See [LICENSE](../../LICENSE) for details.

## Version

Current version: 1.1.0 (Experimental)

See [CHANGELOG.md](./CHANGELOG.md) for version history.

## 🔄 v3.1 Consolidation (Partially Tested)

**What Changed:**
- Single tool with dual modes (basic + Claude AI)
- Use `--claude` flag for AI-powered insights
- Cost tracking automatically displays API usage
- Removed separate `dora-analyzer-claude.sh` file

**Testing Status:**
- ✅ DORA Metrics: Fully tested with both modes
- ✅ Cost tracking verified

**Example:**
```bash
# Basic mode
./dora-analyzer.sh deployment-data.json

# Claude AI mode
./dora-analyzer.sh --claude deployment-data.json
```

### Test Organization

The [Gibson Powers Test Organization](https://github.com/Gibson-Powers-Test-Org) provides sample repositories for testing:

```bash
# Test with sample data
./dora-analyzer.sh deployment-data.json

# Test with Claude AI
./dora-analyzer.sh --claude deployment-data.json
```

### All Arguments

```
OPTIONS:
    -f, --format FORMAT     Output format: text|json|csv (default: text)
    -o, --output FILE       Write results to file
    --claude                Use Claude AI for advanced analysis
    -k, --api-key KEY       Anthropic API key
    -h, --help              Show this help message
```
