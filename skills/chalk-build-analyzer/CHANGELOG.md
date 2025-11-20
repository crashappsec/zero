<!--
Copyright (c) 2024 Crash Override Inc
101 Fulton St, 416, New York 10038
SPDX-License-Identifier: GPL-3.0
-->

# Changelog

All notable changes to the Chalk Build Analyzer skill will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2024-11-20

### Added
- **Automation Scripts for CI/CD Integration**
  - `chalk-build-analyzer.sh` - Basic Chalk report analysis
    - Analyze Chalk build reports (JSON)
    - Compare two builds for regression detection
    - Stage-by-stage breakdown
    - Cache effectiveness metrics
    - Resource utilization analysis
    - Performance categorization (GOOD/MODERATE/HIGH)
    - JSON export capability
    - Threshold-based regression alerts

  - `chalk-build-analyzer-claude.sh` - AI-enhanced build analysis
    - All features from basic analyzer
    - Claude API integration (claude-sonnet-4-20250514)
    - Executive summaries with build health scores
    - Performance bottleneck identification
    - Engineering velocity metrics (DORA)
    - Actionable recommendations with effort/impact estimates
    - Cost & efficiency analysis with ROI calculations
    - Root cause analysis for regressions
    - Severity assessment with business context

  - `compare-analyzers.sh` - Comparison tool
    - Runs both basic and Claude-enhanced analyzers
    - Supports both single build and comparison modes
    - Side-by-side capability comparison
    - Value-add demonstration for engineering leaders
    - Comprehensive comparison report
    - Optional output file preservation
    - Use case recommendations

- **Enhanced Documentation**
  - Automation scripts section in README
  - Usage examples for all three scripts
  - CI/CD integration examples (GitHub Actions, GitLab CI)
  - Prerequisites and requirements documentation
  - Single build and comparison mode examples

### Requirements
- jq: `brew install jq` (or `apt-get install jq`)
- Anthropic API key (for Claude-enhanced analyzer)
- Chalk build reports in JSON format

### Use Cases
- **CI/CD Pipelines**: Automated build performance analysis in pipelines
- **Regression Detection**: Catch build performance regressions automatically
- **Engineering Velocity**: Track DORA metrics and team productivity
- **Performance Optimization**: Identify bottlenecks and optimization opportunities
- **Cost Management**: Resource waste analysis and savings recommendations
- **Comparison Analysis**: Demonstrate AI value-add to stakeholders

## [1.0.0] - 2024-11-20

### Added
- Initial release of Chalk Build Analyzer skill
- Build performance analysis from Chalk reports
- Stage-by-stage performance breakdown
- Bottleneck detection and identification
- Resource utilization monitoring
- Queue time analysis
- DORA metrics calculation
  - Deployment frequency tracking
  - Lead time measurement
  - MTTR (Mean Time To Recovery) calculation
  - Change failure rate monitoring
- Team productivity metrics
  - Builds per developer
  - Collaboration patterns
  - Success rate tracking
  - Velocity scoring
- Regression detection capabilities
  - Performance comparison between builds
  - Impact analysis
  - Quality change tracking
  - Efficiency monitoring
- Trend analysis and predictions
  - Historical pattern visualization
  - Cost trend tracking
  - Team pattern understanding
  - Predictive insights
- Python analysis scripts
  - `scripts/analyze_build.py` - Single build analyzer
  - `scripts/compare_builds.py` - Build comparison tool
  - `scripts/demo.sh` - Interactive demonstration
- Comprehensive documentation
  - Build metrics reference
  - Example Chalk reports
  - Usage examples and best practices
  - Integration guides

### Supported Metrics
- Build duration and timing
- Stage breakdown (compile, test, package)
- Cache effectiveness
- Parallelization efficiency
- Resource utilization (CPU, memory)
- Queue times and contention
- Test coverage and pass rates
- Build success/failure rates

### Known Limitations
- Requires Chalk build reports in JSON format
- DORA metrics calculation requires sufficient historical data
- Trend analysis needs multiple data points over time
- Python scripts require pandas, matplotlib for full functionality
