<!--
Copyright (c) 2024 Gibson Powers Contributors

SPDX-License-Identifier: GPL-3.0
-->

# Code Ownership Analyzer

**Status**: 🚀 Production-Ready v3.0

Enterprise-grade code ownership analysis with historical tracking, multi-format reporting, and intelligent CODEOWNERS generation.

## ✨ What's New in v3.0

**Phase 3 Advanced Features**:
- ✅ **Historical trend tracking** (time-series analysis, snapshot storage, metric change detection)
- ✅ **Markdown reports** (GitHub-flavored, badges, tables, visualizations)
- ✅ **CSV export** (multi-sheet, Excel/Sheets compatible, time-series data)
- ✅ **Strategic CODEOWNERS generator** (pattern detection, intelligent rules, validation)
- ✅ **Trend visualization** (ASCII charts, linear regression, predictions)
- ✅ **Contributor trend analysis** (emerging, declining, churned classification)

**Phase 2 Enterprise Features (v2.5)**:
- ✅ **Hierarchical configuration system** (global, local, environment)
- ✅ **Enhanced 5-component ownership score** (commits, lines, reviews, recency, consistency)
- ✅ **Succession planning module** (identify successors, mentorship recommendations, risk detection)
- ✅ **GitHub review metrics** (PR participation tracking, review scores)
- ✅ **Comprehensive test suite** (unit tests + integration tests)
- ✅ **Advanced metrics** (consistency scoring, readiness assessment)

**Phase 1 Features (v2.0)**:
- ✅ **Dual-method measurement** (commit-based + line-based)
- ✅ **Research-backed metrics** (Gini coefficient, bus factor, health scores)
- ✅ **Enhanced SPOF detection** (6-criteria assessment)
- ✅ **Advanced CODEOWNERS validation** (syntax, staleness, coverage, anti-patterns)
- ✅ **Multi-repository support** (organization scanning, batch analysis)
- ✅ **Complete JSON output** (comprehensive structured data)
- ✅ **GitHub integration** (automatic profile mapping)
- ✅ **Modular architecture** (5 library modules for extensibility)

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
- 4 modular libraries

**v2.5 (Production-Ready)**:
- Configuration system
- 5-component ownership score
- Succession planning
- GitHub review metrics
- Comprehensive test suite
- 5 modular libraries + tests

**v3.0 (Production-Ready)**:
- Historical trend tracking
- Markdown & CSV export
- Strategic CODEOWNERS generation
- Trend visualization
- Contributor analysis
- 9 modular libraries + tests

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

## Configuration System

### Hierarchical Configuration

The analyzer supports hierarchical configuration with the following priority (highest to lowest):

1. Command-line arguments
2. Environment variables (`CODE_OWNERSHIP_*`)
3. Local config (`.code-ownership.conf` in repo)
4. Global config (`~/.config/code-ownership/config`)
5. System config (`/etc/code-ownership/config`)
6. Built-in defaults

### Creating Configuration Files

```bash
# Generate default configuration
cd /path/to/repo
cat > .code-ownership.conf << EOF
# Analysis Settings
analysis_method=hybrid
analysis_days=90
output_format=json

# Thresholds
staleness_threshold_days=90
bus_factor_threshold=3
coverage_target=90

# Health Score Weights (must sum to 1.0)
health_score_weights_coverage=0.35
health_score_weights_distribution=0.25
health_score_weights_freshness=0.20
health_score_weights_engagement=0.20

# GitHub Integration
github_api_enabled=true
include_github_profiles=true
EOF
```

### Environment Variables

```bash
# Override settings with environment variables
export CODE_OWNERSHIP_ANALYSIS_METHOD=commit
export CODE_OWNERSHIP_ANALYSIS_DAYS=120
export CODE_OWNERSHIP_COVERAGE_TARGET=95

./ownership-analyzer-v2.sh .
```

### Available Configuration Options

See `lib/config.sh` for complete list of configurable options.

## Succession Planning

### Overview

The succession planning module identifies potential successors for code owners and generates knowledge transfer plans.

### Features

- **Successor Identification**: Automatically identifies potential successors based on contribution patterns
- **Readiness Scoring**: Calculates readiness scores (0-100) based on:
  - Contribution frequency (30%)
  - Recency (25%)
  - Code familiarity (25%)
  - Collaboration history (20%)
- **Risk Detection**: Identifies files with no successors or inadequate coverage
- **Mentorship Recommendations**: Suggests mentor-mentee pairings based on shared files

### Usage

```bash
# Using the succession planning library directly
source lib/succession.sh

# Generate succession report
generate_succession_report "/path/to/repo" "2024-01-01" "json"

# Identify successors for specific file
identify_successors "/path/to/repo" "src/main.js" "2024-01-01"

# Get mentorship recommendations
recommend_mentorships "/path/to/repo" "2024-01-01" "mentorships.txt"
```

### Example Output

```json
{
  "succession_coverage": 72.5,
  "risk_summary": {
    "critical_risks": 3,
    "high_risks": 8,
    "total_risks": 11
  },
  "mentorship_recommendations": [
    {
      "mentor": "alice@example.com",
      "mentee": "bob@example.com",
      "shared_files": 15,
      "files": ["src/auth.js", "src/api.js", ...]
    }
  ]
}
```

## Historical Trend Tracking

### Overview

Track ownership changes over time to identify patterns and predict future trends.

### Features

- **Snapshot Storage**: Automatic storage of analysis snapshots
- **Time-Series Analysis**: Track metrics over time (coverage, health, bus factor)
- **Contributor Trends**: Classify contributors as emerging, declining, stable, or churned
- **Linear Regression**: Predict future trends based on historical data
- **ASCII Visualizations**: Charts showing metric changes

### Usage

```bash
# Using the trends library
source lib/trends.sh

# Store current analysis as snapshot
store_snapshot "/path/to/repo" "$analysis_json" "2024-11-21"

# Generate trend report
generate_trend_report "/path/to/repo" "json" > trend-report.json

# Analyze contributor trends
analyze_contributor_trends "/path/to/repo" 90

# Create ASCII chart for metric
create_ascii_chart "Health Score" 65 70 75 78 82
```

### Example Output

```json
{
  "snapshot_count": 10,
  "trends": {
    "coverage": {
      "slope": 0.5,
      "direction": "improving"
    },
    "health": {
      "slope": 1.2,
      "direction": "improving"
    }
  },
  "contributor_trends": [
    {
      "email": "alice@example.com",
      "trend": "emerging",
      "file_delta": 15
    }
  ]
}
```

## Report Formats

### Markdown Reports

Generate human-readable GitHub-flavored markdown reports with tables, badges, and visualizations.

```bash
# Using the markdown library
source lib/markdown.sh

# Generate full markdown report
generate_markdown_report "$analysis_json" > report.md

# Generate comparison report
generate_comparison_report "$snapshot1" "$snapshot2" > comparison.md
```

**Markdown Features**:
- Executive summary with health badges
- Repository metrics tables
- Top contributors table
- SPOF table with risk indicators
- Distribution visualization
- Recommendations section
- Collapsible detailed metrics

### CSV Export

Export analysis data to CSV for Excel, Google Sheets, or data analysis tools.

```bash
# Using the CSV library
source lib/csv.sh

# Export all data (creates multiple CSV files)
export_all_csv "$analysis_json" "report" "/path/to/repo"

# Individual exports
export_ownership_csv "$analysis_json" "ownership.csv"
export_metrics_csv "$analysis_json" "metrics.csv"
export_spof_csv "$analysis_json" "spofs.csv"
export_timeseries_csv "/path/to/repo" "timeseries.csv"
```

**CSV Outputs**:
- `*_ownership.csv`: Contributor ownership percentages
- `*_metrics.csv`: Key metrics and status
- `*_spofs.csv`: Single points of failure analysis
- `*_contributors_detailed.csv`: Detailed contributor stats
- `*_timeseries.csv`: Historical metric data
- `*_files.csv`: File-level ownership details

## Strategic CODEOWNERS Generation

### Overview

Intelligently generate CODEOWNERS files based on actual ownership patterns detected in your repository.

### Features

- **Pattern Detection**: Automatically detect directory and file extension patterns
- **Ownership Analysis**: Identify primary owners for each pattern
- **GitHub Username Mapping**: Convert emails to @mentions
- **Validation**: Check generated CODEOWNERS for common issues
- **Best Practices**: Follow GitHub CODEOWNERS conventions

### Usage

```bash
# Using the CODEOWNERS generator
source lib/codeowners-generator.sh

# Generate strategic CODEOWNERS
generate_strategic_codeowners "/path/to/repo" ".github/CODEOWNERS"

# With custom ownership threshold (default 60%)
generate_strategic_codeowners "/path/to/repo" ".github/CODEOWNERS" "2024-01-01" 75

# Validate generated file
validate_generated_codeowners ".github/CODEOWNERS"
```

**Example Output**:
```
# CODEOWNERS
# Auto-generated by Code Ownership Analyzer

# Global Owners
* @alice @bob @charlie

# Pattern-Based Ownership
*.js                @alice        # 85.2% ownership
*.py                @bob          # 92.1% ownership
*.md                @charlie      # 78.5% ownership

# Directory-Based Ownership
/src/frontend/      @alice        # 89.3% ownership
/src/backend/       @bob          # 91.7% ownership
/docs/              @charlie      # 95.2% ownership
```

## Testing

### Running Tests

```bash
# Run all tests (unit + integration)
cd utils/code-ownership/tests
./run-all-tests.sh

# Run individual test suites
./test-metrics.sh         # Metrics library unit tests
./test-config.sh          # Configuration library unit tests
./test-integration.sh     # Full analyzer integration tests
```

### Test Coverage

**Unit Tests**:
- `test-metrics.sh`: Tests all metric calculation functions
  - Recency factors
  - Gini coefficients
  - Bus factors
  - Health scores
  - Ownership scores
- `test-config.sh`: Tests configuration system
  - Loading hierarchy
  - Validation
  - Type conversions

**Integration Tests**:
- `test-integration.sh`: End-to-end analyzer tests
  - Basic analysis workflow
  - CODEOWNERS validation
  - Different analysis methods
  - Output formats
  - Configuration integration
  - Library loading

### Continuous Integration

```yaml
# Example GitHub Actions workflow
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Install dependencies
        run: |
          brew install jq bc
      - name: Run tests
        run: |
          cd utils/code-ownership/tests
          ./run-all-tests.sh
```

## Known Limitations

### Resolved in v2.5

- ~~**No Configuration System**~~ → ✅ Hierarchical config system implemented
- ~~**Limited Output Formats**~~ → ✅ JSON, text, and markdown support
- ~~**Email-Based Only**~~ → ✅ GitHub username mapping implemented
- ~~**No Succession Planning**~~ → ✅ Full succession planning module

### Resolved in v3.0

- ~~**No Historical Trend Tracking**~~ → ✅ Full trend tracking with time-series analysis
- ~~**Limited Report Formats**~~ → ✅ Markdown and CSV export implemented
- ~~**Manual CODEOWNERS Generation**~~ → ✅ Strategic pattern-based generation

### Current Limitations

1. **Review Metrics Require GitHub Token**: PR review data needs `GITHUB_TOKEN`
2. **Basic Pattern Matching**: CODEOWNERS glob patterns use simplified matching
3. **Rate Limited API Calls**: GitHub API subject to rate limits
4. **Single Platform Support**: Currently Git/GitHub only (GitLab/Bitbucket planned)

### Analysis Limitations

- Assumes all commits are equally important (weighted by recency in v2.5)
- May not reflect current team structure
- Pair programming not distinguished
- Automated commits may skew results

## Roadmap

### Phase 1: Core Functionality ✅ COMPLETE
- [x] Basic ownership calculation
- [x] CODEOWNERS validation
- [x] Bus factor analysis
- [x] AI-enhanced insights
- [x] Dual-method measurement
- [x] Multi-repository support
- [x] GitHub integration
- [x] JSON output

### Phase 2: Enterprise Features ✅ COMPLETE
- [x] Hierarchical configuration system
- [x] Enhanced 5-component ownership score
- [x] Succession planning module
- [x] GitHub review metrics
- [x] Comprehensive unit tests
- [x] Integration test suite
- [x] Complete documentation

### Phase 3: Advanced Features ✅ COMPLETE
- [x] Historical trend tracking
- [x] Trend visualization
- [x] Markdown report format
- [x] CSV export
- [x] Strategic CODEOWNERS generation
- [ ] Platform support (GitLab, Bitbucket)

### Phase 4: Production Optimization
- [ ] Performance optimization
- [ ] Dashboard integration
- [ ] Slack/email notifications
- [ ] Team structure analysis
- [ ] CI/CD examples
- [ ] Enterprise features

## Development

### Architecture

```
code-ownership/
├── ownership-analyzer.sh              # Legacy v1.0 analyzer
├── ownership-analyzer-v2.sh           # ⭐ Enhanced v3.0 analyzer (recommended)
├── ownership-analyzer-claude.sh       # AI-enhanced analyzer
├── compare-analyzers.sh               # Comparison tool
├── lib/                               # Library modules (9 total)
│   ├── metrics.sh                     # Research-backed metric calculations
│   ├── github.sh                      # GitHub API integration
│   ├── analyzer-core.sh               # Dual-method analysis engine
│   ├── codeowners-validator.sh        # Advanced validation
│   ├── config.sh                      # Configuration system
│   ├── succession.sh                  # Succession planning
│   ├── trends.sh                      # ⭐ Historical trend tracking
│   ├── markdown.sh                    # ⭐ Markdown report generation
│   ├── csv.sh                         # ⭐ CSV export functionality
│   └── codeowners-generator.sh        # ⭐ Strategic CODEOWNERS generation
└── tests/                             # Test suite
    ├── run-all-tests.sh               # Test runner
    ├── test-metrics.sh                # Metrics unit tests
    ├── test-config.sh                 # Config unit tests
    └── test-integration.sh            # Integration tests
```

### Adding Features

Completed in v3.0:
- ✅ Historical trend tracking (time-series analysis, snapshots)
- ✅ Markdown report generation (badges, tables, visualizations)
- ✅ CSV export (multi-sheet, Excel/Sheets compatible)
- ✅ Strategic CODEOWNERS generation (pattern detection, intelligent rules)

Completed in v2.5:
- ✅ Configuration system (hierarchical)
- ✅ GitHub API integration (profile mapping, review metrics)
- ✅ Comprehensive test suite (unit + integration)
- ✅ Enhanced documentation

Priority development areas for Phase 4:

1. **Performance Optimization**: Large repository handling
2. **Dashboard Integration**: Web UI for visualizations
3. **Notification Systems**: Slack/email alerts
4. **Team Structure Analysis**: Organizational insights
5. **Platform Support**: GitLab, Bitbucket integration

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

Current version: 3.0.0 (Production-Ready)

See [CHANGELOG.md](./CHANGELOG.md) for version history.

### Version History

- **v3.0.0** (Phase 3): Historical trend tracking, markdown/CSV reports, strategic CODEOWNERS generation, trend visualization
- **v2.5.0** (Phase 2): Configuration system, succession planning, enhanced metrics, test suite
- **v2.0.0** (Phase 1): Dual-method analysis, multi-repo, GitHub integration, JSON output
- **v1.0.0** (Experimental): Basic ownership analysis, CODEOWNERS validation
