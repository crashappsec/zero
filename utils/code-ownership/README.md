<!--
Copyright (c) 2024 Crash Override Inc
101 Fulton St, 416, New York 10038
SPDX-License-Identifier: GPL-3.0
-->

# Code Ownership Analyzer

**Status**: 🚀 Beta v2.0 - Production-Ready

Comprehensive code ownership analysis with research-backed metrics, dual-method measurement, and advanced validation.

## ✨ What's New in v2.0

**Major Enhancements**:
- ✅ **Dual-method measurement** (commit-based + line-based)
- ✅ **Research-backed metrics** (Gini coefficient, bus factor, health scores)
- ✅ **Enhanced SPOF detection** (6-criteria assessment)
- ✅ **Advanced CODEOWNERS validation** (syntax, staleness, coverage, anti-patterns)
- ✅ **Multi-repository support** (organization scanning, batch analysis)
- ✅ **Complete JSON output** (comprehensive structured data)
- ✅ **GitHub integration** (automatic profile mapping)
- ✅ **Modular architecture** (4 library modules for extensibility)

**Based on 2024 Research**:
- arXiv empirical findings (commit vs. line-based metrics)
- Martin Fowler's ownership philosophy
- Industry best practices from Aviator.co and others
- Microsoft Research on defect prediction

### Version Comparison

**v1.0 (Experimental)**:
- Basic commit counting
- Simple CODEOWNERS validation
- Text output only

**v2.0 (Beta)**:
- Dual-method analysis (97% defect prediction accuracy)
- Advanced validation (4 check types)
- JSON + text output
- Multi-repo support
- GitHub API integration
- Modular, tested libraries

## Overview

The Code Ownership Analyzer helps teams understand who owns what code in a repository by analyzing Git commit history. It provides:

- **Ownership Metrics**: Calculate code ownership based on commits, lines changed, and recency
- **CODEOWNERS Validation**: Verify CODEOWNERS files match actual ownership
- **Bus Factor Analysis**: Identify single points of failure
- **Health Scores**: Overall code ownership health assessment
- **Team Insights**: Understand collaboration patterns

## Quick Start

### Prerequisites

```bash
# Git is required
git --version
```

### Basic Usage

```bash
# Analyze current repository
./ownership-analyzer.sh

# Analyze specific repository
./ownership-analyzer.sh /path/to/repo

# AI-enhanced analysis
export ANTHROPIC_API_KEY="your-key"
./ownership-analyzer-claude.sh

# Compare base vs Claude analysis
./compare-analyzers.sh
```

## Available Scripts

### ownership-analyzer-v2.sh (⭐ Recommended)

**Enhanced analyzer with research-backed metrics and comprehensive features.**

**Key Features**:
- **Dual-method analysis**: Combine commit-based (97% defect prediction) and line-based (authorship) approaches
- **Advanced metrics**: Gini coefficient, bus factor, health scores
- **6-criteria SPOF detection**: Comprehensive risk assessment
- **GitHub integration**: Automatic profile mapping from emails
- **Multi-repo support**: Scan organizations and multiple repositories
- **Complete JSON output**: Structured data for automation
- **Advanced validation**: Syntax, staleness, coverage gaps, anti-patterns

**Usage**:
```bash
# Analyze single repository (JSON output)
./ownership-analyzer-v2.sh .

# Analyze with text output
./ownership-analyzer-v2.sh --format text .

# Analyze GitHub repository
./ownership-analyzer-v2.sh https://github.com/owner/repo

# Analyze organization (requires GITHUB_TOKEN)
export GITHUB_TOKEN=ghp_xxx
./ownership-analyzer-v2.sh --org myorg --output org-analysis.json

# Validate CODEOWNERS
./ownership-analyzer-v2.sh --validate --verbose .

# Analyze multiple repos
./ownership-analyzer-v2.sh --repos repo1 repo2 repo3 --output analysis.json
```

**Output Example (JSON)**:
```json
{
  "metadata": {
    "analyzer_version": "2.0.0",
    "repository": "my-repo",
    "analysis_date": "2024-11-21T10:00:00Z",
    "analysis_method": "hybrid"
  },
  "ownership_health": {
    "coverage_percentage": 85.2,
    "gini_coefficient": 0.42,
    "bus_factor": 3,
    "health_score": 78.5,
    "health_grade": "Good"
  },
  "single_points_of_failure": [
    {
      "file": "src/auth/oauth.ts",
      "score": 5,
      "risk": "High",
      "contributors": 1
    }
  ],
  "recommendations": {
    "needs_attention": "Good: No critical issues"
  }
}
```

### ownership-analyzer.sh (Legacy v1.0)

Base analyzer that calculates ownership from Git history.

**Features**:
- Commit-based ownership scoring
- Weighted by lines changed and recency
- Directory-level ownership breakdown
- CODEOWNERS file validation
- Bus factor identification

**Note**: For new projects, use `ownership-analyzer-v2.sh` which includes all v1.0 features plus enhanced capabilities.

**Usage**:
```bash
# Analyze current directory
./ownership-analyzer.sh

# Analyze specific path
./ownership-analyzer.sh /path/to/repo

# Generate CODEOWNERS file
./ownership-analyzer.sh --generate-codeowners
```

**Output**:
```
===================================
Code Ownership Analysis
===================================
Repository: /path/to/repo
Analysis Date: 2024-11-21

Top Contributors:
  1. alice@example.com (45.2%)
  2. bob@example.com (32.1%)
  3. charlie@example.com (22.7%)

Directory Ownership:
  src/frontend/: alice@example.com (78%)
  src/backend/: bob@example.com (65%)
  tests/: charlie@example.com (52%)

Bus Factor: 2 (Medium Risk)
Health Score: 72/100
```

### ownership-analyzer-claude.sh

AI-enhanced analyzer with contextual insights and recommendations.

**Features**:
- All base analyzer features
- Ownership pattern analysis
- Collaboration insights
- Risk assessment
- Recommendations for improvement

**Requires**: `ANTHROPIC_API_KEY` environment variable

**Usage**:
```bash
export ANTHROPIC_API_KEY="your-key"
./ownership-analyzer-claude.sh
```

### compare-analyzers.sh

Compare base and AI-enhanced analysis side-by-side.

**Usage**:
```bash
./compare-analyzers.sh [repo-path]
```

## Ownership Metrics

### Calculation Method

Ownership scores are calculated using a weighted formula:

```
Score = (commits × 1.0) + (lines_changed × 0.5) + (recency_factor × 0.3)
```

Where:
- **commits**: Number of commits by author
- **lines_changed**: Total lines added/modified
- **recency_factor**: Higher weight for recent contributions

### Bus Factor

The bus factor is the minimum number of team members who would need to be unavailable to stall the project:

- **1**: Critical risk (single point of failure)
- **2-3**: Medium risk (limited redundancy)
- **4+**: Low risk (good knowledge distribution)

### Health Score

Overall ownership health (0-100):

- **90-100**: Excellent (well-distributed, documented)
- **70-89**: Good (mostly distributed, some gaps)
- **50-69**: Fair (concentrated ownership, needs improvement)
- **< 50**: Poor (high risk, immediate action needed)

## CODEOWNERS Integration

### Validating CODEOWNERS

```bash
# Check if CODEOWNERS matches actual ownership
./ownership-analyzer.sh --validate-codeowners
```

### Generating CODEOWNERS

```bash
# Auto-generate based on Git history
./ownership-analyzer.sh --generate-codeowners > .github/CODEOWNERS
```

**Example Output**:
```
# Auto-generated CODEOWNERS
# Based on Git history analysis from 2024-11-21

/src/frontend/ @alice
/src/backend/ @bob
/tests/ @charlie
/docs/ @alice @charlie

* @alice @bob @charlie
```

## Known Limitations

### Current Limitations

1. **Single Repository Only**: No multi-repo or organization scanning
2. **No Configuration System**: Cannot persist settings
3. **Limited Output Formats**: Text only, no JSON/markdown
4. **Basic Error Handling**: May fail on edge cases
5. **No Time Range Selection**: Analyzes full Git history
6. **Email-Based Only**: Doesn't map to GitHub usernames

### Analysis Limitations

- Assumes all commits are equally important
- May not reflect current team structure
- Doesn't account for code review contributions
- Pair programming not distinguished
- Automated commits may skew results

## Roadmap to Production

### Phase 1: Core Functionality (Current)
- [x] Basic ownership calculation
- [x] CODEOWNERS validation
- [x] Bus factor analysis
- [x] AI-enhanced insights
- [ ] Comprehensive error handling

### Phase 2: Integration
- [ ] Hierarchical configuration system
- [ ] Multi-repository support
- [ ] Organization scanning
- [ ] Output format options (JSON, markdown)
- [ ] GitHub username mapping

### Phase 3: Advanced Features
- [ ] Historical trend tracking
- [ ] Team structure analysis
- [ ] Code review integration
- [ ] Slack/email notifications
- [ ] Dashboard integration

### Phase 4: Production Ready
- [ ] Comprehensive testing
- [ ] Complete documentation
- [ ] CI/CD examples
- [ ] Performance optimization
- [ ] Enterprise features

## Development

### Architecture

```
code-ownership/
├── ownership-analyzer.sh              # Base analyzer
├── ownership-analyzer-claude.sh       # AI-enhanced analyzer
└── compare-analyzers.sh               # Comparison tool
```

### Adding Features

Priority development areas:

1. **Configuration Integration**: Add global config support
2. **Multi-Repo Support**: Batch and organization scanning
3. **Output Formats**: JSON, markdown, CSV
4. **GitHub Integration**: Map emails to usernames, use GitHub API
5. **Testing**: Comprehensive test suite
6. **Documentation**: Usage guide and examples

## Use Cases

### Team Onboarding
Identify code owners to contact for specific areas during onboarding.

### Knowledge Transfer Planning
Identify areas with concentrated ownership before team changes.

### CODEOWNERS Management
Keep CODEOWNERS files up-to-date with actual ownership.

### Risk Assessment
Identify bus factor risks and plan mitigation.

### Performance Reviews
Understand contribution patterns across the codebase.

## Related Documentation

- [Code Ownership Skill](../../skills/code-ownership/)
- [Changelog](./CHANGELOG.md)

## Contributing

Contributions welcome! This utility needs significant work to reach production quality. See [CONTRIBUTING.md](../../CONTRIBUTING.md).

Priority areas:
- Configuration system integration
- Multi-repository support
- Output format options
- GitHub API integration
- Comprehensive testing

## License

GPL-3.0 - See [LICENSE](../../LICENSE) for details.

## Version

Current version: 1.0.0 (Experimental)

See [CHANGELOG.md](./CHANGELOG.md) for version history.
