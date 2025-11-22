<!--
Copyright (c) 2024 Gibson Powers Contributors

SPDX-License-Identifier: GPL-3.0
-->

# DORA Metrics Skill

Comprehensive DORA (DevOps Research and Assessment) metrics analysis for measuring and improving software delivery performance.

## Purpose

This skill enables teams to measure, analyze, and improve their software delivery and operational performance using the four key DORA metrics - industry-standard measurements proven to correlate with organizational success.

### What You Can Do

- **Calculate DORA Metrics**: Deployment Frequency, Lead Time, Change Failure Rate, MTTR
- **Performance Classification**: Elite, High, Medium, or Low performer status
- **Benchmark Comparison**: Compare against DORA research findings
- **Trend Analysis**: Track improvements over time
- **Team Comparison**: Compare multiple teams side-by-side
- **Root Cause Analysis**: Understand why metrics regress
- **Improvement Planning**: Create actionable roadmaps
- **Executive Reporting**: Generate leadership summaries

## The Four Key DORA Metrics

### 1. Deployment Frequency (DF)
**How often you deploy to production**

- **Elite:** Multiple deploys per day
- **High:** Daily to weekly deploys
- **Medium:** Weekly to monthly deploys
- **Low:** Monthly to semi-annually

### 2. Lead Time for Changes (LT)
**Time from commit to production**

- **Elite:** < 1 hour
- **High:** 1 day - 1 week
- **Medium:** 1 week - 1 month
- **Low:** 1 month - 6 months

### 3. Change Failure Rate (CFR)
**Percentage of deployments causing failures**

- **Elite:** 0-15%
- **High:** 16-30%
- **Medium:** 31-45%
- **Low:** 46-60%

### 4. Time to Restore Service (MTTR)
**Time to recover from incidents**

- **Elite:** < 1 hour
- **High:** < 1 day
- **Medium:** 1 day - 1 week
- **Low:** > 1 week

## Prerequisites

- Deployment data from CI/CD systems (GitHub Actions, GitLab CI, Jenkins, etc.)
- Incident data from monitoring/alerting systems (PagerDuty, Datadog, etc.)
- Git commit history
- Basic understanding of software delivery metrics

## Usage

### With the Skill

Load the DORA Metrics skill in Crash Override and use natural language:

```
Calculate our DORA metrics from this deployment data:
- 45 deployments in 30 days
- 4 failures
- Median lead time: 2.5 hours
- Median MTTR: 42 minutes
```

```
Why is our change failure rate so high? It went from 12% to 35% in two weeks.
Context: Deployed 15 times, had 5 different types of failures
```

```
Create an improvement roadmap to move from High to Elite performer.
Current: DF=2/day, LT=4 hours, CFR=18%, MTTR=2 hours
Timeline: 3 months
```

### With Automation Scripts

The skill includes bash scripts for automated analysis:

#### dora-analyser.sh

Basic DORA metrics calculation without AI enhancement.

```bash
# Analyze deployment data
./dora-analyser.sh deployment-data.json

# Export to JSON
./dora-analyser.sh --format json --output metrics.json deployment-data.json

# Export to CSV
./dora-analyser.sh --format csv --output metrics.csv deployment-data.json
```

**Requirements:**
- jq: `brew install jq`
- bc: `brew install bc`

#### dora-analyser-claude.sh

AI-enhanced analysis with insights and recommendations.

```bash
# Set API key
export ANTHROPIC_API_KEY=sk-ant-xxx

# Analyze with AI insights
./dora-analyser-claude.sh deployment-data.json

# Specify API key directly
./dora-analyser-claude.sh --api-key sk-ant-xxx deployment-data.json
```

**Output Includes:**
1. Executive Summary - Performance overview
2. Metric Analysis - Detailed breakdown
3. Strengths - What's working well
4. Improvement Opportunities - Prioritized recommendations
5. Roadmap - Short/medium/long-term actions

**Requirements:**
- Same as basic analyser
- Anthropic API key

#### compare-analysers.sh

Comparison tool showing value-add of AI enhancement.

```bash
# Compare basic vs Claude analysis
./compare-analysers.sh deployment-data.json
```

### Test with Safe Repositories

🧪 **Practice DORA metrics analysis safely:**

The [Gibson Powers Test Organization](https://github.com/Gibson-Powers-Test-Org) provides sample repositories with realistic commit and deployment patterns for testing.

```bash
# Analyze test repository deployment patterns
./dora-analyser.sh --repo https://github.com/Gibson-Powers-Test-Org/sample-repo

# Generate sample deployment data from test repo
./dora-analyser-claude.sh \
  --extract-from-git https://github.com/Gibson-Powers-Test-Org/sample-repo
```

Perfect for:
- Learning DORA metrics analysis
- Testing configurations
- Creating example reports
- Contributing examples

## Data Format

Input data should be JSON with deployment and incident information:

```json
{
  "team": "Platform Engineering",
  "period": "2024-Q4",
  "deployments": [
    {
      "commit_time": "2024-11-01T09:00:00Z",
      "production_time": "2024-11-01T10:30:00Z",
      "status": "success"
    }
  ],
  "incidents": [
    {
      "detected_at": "2024-11-05T14:00:00Z",
      "resolved_at": "2024-11-05T14:45:00Z",
      "severity": "high"
    }
  ],
  "summary": {
    "total_deployments": 45,
    "failed_deployments": 4,
    "median_lead_time_hours": 2.5,
    "median_mttr_minutes": 42
  }
}
```

See [examples/sample-deployment-data.json](./examples/sample-deployment-data.json) for complete schema.

## Examples

### Example Reports

- [Full DORA Analysis Report](./examples/example-dora-analysis.md) - Comprehensive Q4 analysis
- [Team Comparison Report](./examples/example-team-comparison.md) - Multi-team comparison

### Sample Prompts

See [prompts/dora/](../../prompts/dora/) for prompt templates:

- **Analysis**: Calculate metrics, classify performance
- **Improvement**: Create roadmaps, optimize metrics
- **Reporting**: Executive summaries, team comparisons
- **Troubleshooting**: Diagnose regressions, root cause analysis

## Common Use Cases

### Quarterly Business Review

```
Generate an executive summary for our Q4 board presentation.

Team: Engineering Department
Metrics: DF=3.2/day (Elite), LT=6 hours (High), CFR=15% (Elite), MTTR=1 hour (Elite)

Focus on how engineering excellence enables business goals.
```

### Performance Improvement

```
We're stuck at High performance. Create a roadmap to reach Elite.

Current: All metrics at High level
Timeline: 6 months
Biggest challenge: Lead time still 4 hours
```

### Incident Investigation

```
Our MTTR doubled from 30 minutes to 2 hours. Why?

Context: Same team size, same tools, no new incident types
Help diagnose what changed.
```

### Team Optimization

```
Compare our Platform team vs Backend team.

Platform: DF=5.6/day, LT=1.8hrs, CFR=11%, MTTR=42min
Backend: DF=0.8/day, LT=2.1days, CFR=12%, MTTR=8hrs

Identify best practices to share and areas needing help.
```

## CI/CD Integration

### GitHub Actions

```yaml
- name: Calculate DORA Metrics
  run: |
    ./dora-analyser.sh --format json --output metrics.json deployment-data.json

- name: AI Analysis (on main)
  if: github.ref == 'refs/heads/main'
  env:
    ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
  run: |
    ./dora-analyser-claude.sh deployment-data.json > dora-report.txt
```

### GitLab CI

```yaml
dora_metrics:
  script:
    - ./dora-analyser.sh deployment-data.json
  artifacts:
    reports:
      metrics: metrics.json
```

## Best Practices

### Measurement

- **Automate data collection** from CI/CD and monitoring tools
- **Define metrics clearly** and consistently
- **Use median over mean** for time-based metrics
- **Track confidence levels** with measurements
- **Validate data quality** regularly

### Analysis

- **Focus on trends** not point-in-time snapshots
- **Consider context** (holidays, team changes, projects)
- **Correlate with business outcomes**
- **Segment by team/service/product**

### Improvement

- **Start with one metric** to avoid overwhelm
- **Celebrate small wins** to build momentum
- **Share successes** across teams
- **Address culture** not just process/tools
- **Measure impact** of improvement initiatives

### Communication

- **Tailor to audience** (executives vs engineers)
- **Show progress** not just current state
- **Make recommendations actionable**
- **Use for learning** not punishment
- **Foster psychological safety**

## Resources

### Official DORA Research
- [DORA.dev](https://dora.dev/) - Official research site
- [State of DevOps Reports](https://dora.dev/research/) - Annual research findings
- [Accelerate book](https://www.oreilly.com/library/view/accelerate/9781457191435/) - Foundational research

### Related Skills
- [Chalk Build Analyser](../chalk-build-analyser/) - Build performance metrics
- [SBOM Analyser](../sbom-analyser/) - Supply chain security
- [Certificate Analyser](../certificate-analyser/) - TLS/SSL analysis

## Contributing

Improvements welcome! Consider contributing:
- Additional data source integrations
- Enhanced analysis capabilities
- More example reports
- Additional prompt templates

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for guidelines.

## License

This skill is licensed under GPL-3.0. See [LICENSE](../../LICENSE) for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/crashappsec/skills-and-prompts/issues)
- **Discussions**: [GitHub Discussions](https://github.com/crashappsec/skills-and-prompts/discussions)
- **Contact**: mark@crashoverride.com

---

**Start measuring and improving your software delivery performance today!**

## 🔄 Recent Updates (v3.1)

**Tool Consolidation:**
- Single `dora-analyser.sh` tool replaces separate versions
- `--claude` flag enables AI-powered analysis
- Cost tracking automatically displays when using Claude
- Note: Claude features fully tested

**Usage:**
```bash
# Basic DORA metrics
./dora-analyser.sh deployment-data.json

# AI-enhanced with cost tracking
./dora-analyser.sh --claude deployment-data.json
```
