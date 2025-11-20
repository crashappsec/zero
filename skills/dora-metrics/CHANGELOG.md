<!--
Copyright (c) 2024 Crash Override Inc
101 Fulton St, 416, New York 10038
SPDX-License-Identifier: GPL-3.0
-->

# Changelog

All notable changes to the DORA Metrics skill will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2024-11-20

### Added
- Initial release of DORA Metrics skill
- Comprehensive knowledge of all four DORA metrics
  - Deployment Frequency (DF)
  - Lead Time for Changes (LT)
  - Change Failure Rate (CFR)
  - Time to Restore Service (MTTR)
- Performance level classifications (Elite, High, Medium, Low)
- Detailed benchmark comparisons against DORA research
- Metric calculation methodologies
- Root cause analysis capabilities
- Improvement recommendation framework
- Team comparison functionality
- Trend analysis and forecasting
- Executive reporting capabilities

- **Automation Scripts for CI/CD Integration**
  - `dora-analyzer.sh` - Basic DORA metrics calculation
    - Calculate all four metrics from deployment data
    - Performance classification
    - Benchmark comparison
    - Multiple output formats (text, JSON, CSV)
    - Automated data validation

  - `dora-analyzer-claude.sh` - AI-enhanced analysis
    - All features from basic analyzer
    - Claude API integration (claude-sonnet-4-20250514)
    - Executive summaries with business context
    - Root cause analysis
    - Prioritized improvement recommendations
    - Actionable roadmaps (short/medium/long-term)
    - Team-specific insights

  - `compare-analyzers.sh` - Comparison tool
    - Runs both basic and Claude-enhanced analyzers
    - Side-by-side capability comparison
    - Value-add demonstration
    - Use case recommendations

- **Example Data and Reports**
  - `examples/sample-deployment-data.json` - Example data format
  - `examples/example-dora-analysis.md` - Comprehensive analysis report
  - `examples/example-team-comparison.md` - Multi-team comparison report

- **Sample Prompts**
  - `prompts/dora/analysis/calculate-metrics.md` - Metric calculation
  - `prompts/dora/improvement/create-roadmap.md` - Improvement planning
  - `prompts/dora/reporting/executive-summary.md` - Leadership reporting
  - `prompts/dora/troubleshooting/diagnose-metric-regression.md` - Problem diagnosis
  - `prompts/dora/README.md` - Comprehensive prompt guide

- **Comprehensive Documentation**
  - Complete skill file with DORA expertise
  - README with usage examples
  - Metric definitions and calculations
  - Best practices and recommendations
  - CI/CD integration examples
  - Troubleshooting guides

### Performance Benchmarks Included

**Elite Performers:**
- Deployment Frequency: On-demand (multiple per day)
- Lead Time: < 1 hour
- Change Failure Rate: 0-15%
- MTTR: < 1 hour

**High Performers:**
- Deployment Frequency: Daily to weekly
- Lead Time: 1 day - 1 week
- Change Failure Rate: 16-30%
- MTTR: < 1 day

**Medium Performers:**
- Deployment Frequency: Weekly to monthly
- Lead Time: 1 week - 1 month
- Change Failure Rate: 31-45%
- MTTR: 1 day - 1 week

**Low Performers:**
- Deployment Frequency: Monthly to semi-annually
- Lead Time: 1 month - 6 months
- Change Failure Rate: 46-60%
- MTTR: > 1 week

### Key Capabilities

1. **Metric Calculation**
   - Accurate calculation from deployment data
   - Proper handling of edge cases
   - Multiple time period support
   - Confidence indicators

2. **Performance Classification**
   - Benchmark against DORA research
   - Overall performance assessment
   - Metric-specific classifications
   - Industry comparisons

3. **Analysis and Insights**
   - Pattern detection
   - Anomaly identification
   - Correlation analysis
   - Trend forecasting

4. **Improvement Planning**
   - Prioritized recommendations
   - Effort and impact estimates
   - Phased roadmaps
   - Best practice guidance

5. **Reporting**
   - Executive summaries
   - Detailed technical reports
   - Team comparisons
   - Trend visualizations

### Data Sources Supported

- CI/CD platforms (GitHub Actions, GitLab CI, Jenkins, CircleCI, Travis CI)
- Incident management (PagerDuty, Opsgenie, VictorOps)
- Monitoring tools (Datadog, New Relic, Prometheus)
- Version control (Git commit history)
- Build systems (Chalk metadata)
- Cloud platforms (AWS, GCP, Azure)

### Requirements

- jq: `brew install jq` (or `apt-get install jq`)
- bc: `brew install bc` (or `apt-get install bc`)
- Anthropic API key (for Claude-enhanced analyzer)
- Deployment data in JSON format

### Use Cases

- **Quarterly Reviews**: Executive performance summaries
- **Continuous Improvement**: Ongoing metric tracking
- **Team Optimization**: Cross-team comparisons
- **Incident Analysis**: Root cause investigation
- **Strategic Planning**: Long-term improvement roadmaps
- **Leadership Reporting**: Board and investor updates
- **CI/CD Integration**: Automated metric monitoring

### Known Limitations

- Requires structured deployment data
- API rate limits apply for Claude integration
- Trend analysis needs historical data (minimum 30 days recommended)
- Team comparisons require consistent measurement across teams
- Some industry-specific contexts may need adjustment

## Future Enhancements

Planned for future releases:
- Additional data source integrations
- Predictive analytics
- Custom benchmark support
- Automated data collection scripts
- Dashboard integration
- Slack/Teams notifications
- Advanced visualization support
