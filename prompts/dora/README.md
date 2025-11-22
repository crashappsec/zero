<!--
Copyright (c) 2024 Gibson Powers Contributors

SPDX-License-Identifier: GPL-3.0
-->

# DORA Metrics Prompt Templates

Ready-to-use prompt templates for analyzing, improving, and reporting on DORA (DevOps Research and Assessment) metrics.

## Overview

These prompts help you leverage the DORA Metrics skill to:
- Calculate and classify performance levels
- Diagnose issues and identify improvements
- Create reports for different audiences
- Build improvement roadmaps
- Monitor trends and progress

## Directory Structure

```
dora/
├── analysis/          # Calculate and analyze metrics
├── improvement/       # Create improvement plans
├── reporting/         # Generate reports and summaries
└── troubleshooting/   # Diagnose problems and regressions
```

## Quick Start

### 1. Calculate Your Metrics

Start with the basics - calculate your current DORA metrics:

**Prompt:** [analysis/calculate-metrics.md](./analysis/calculate-metrics.md)

```
Calculate DORA metrics for our Platform team for November 2024:
- 45 deployments in 30 days
- 4 failures (8.9% failure rate)
- Median lead time: 2.5 hours
- Median MTTR: 42 minutes

Classify our performance level and compare to benchmarks.
```

### 2. Create an Improvement Plan

Once you know your current state, create a roadmap:

**Prompt:** [improvement/create-roadmap.md](./improvement/create-roadmap.md)

```
Create a roadmap to move from High to Elite performer.

Current: DF=2/day, LT=4 hours, CFR=18%, MTTR=2 hours
Target: Elite across all metrics
Timeline: 3 months
```

### 3. Generate Reports

Share results with stakeholders:

**Prompt:** [reporting/executive-summary.md](./reporting/executive-summary.md)

```
Create an executive summary for our Q4 Board presentation.

Metrics: DF=3.2/day, LT=6 hours, CFR=15%, MTTR=1 hour
Focus on: How engineering excellence enables business goals
```

### 4. Diagnose Issues

When metrics regress, find out why:

**Prompt:** [troubleshooting/diagnose-metric-regression.md](./troubleshooting/diagnose-metric-regression.md)

```
Our deployment frequency dropped from 5/day to 0.8/week.

Context: New security policy, larger PRs, team on vacation
Help identify root causes and create recovery plan.
```

## Available Prompts

### Analysis

| Prompt | Purpose | Use When |
|--------|---------|----------|
| [calculate-metrics.md](./analysis/calculate-metrics.md) | Calculate all four DORA metrics | You have deployment data and want baseline metrics |

**Coming Soon:**
- trend-analysis.md - Track metrics over time
- team-comparison.md - Compare multiple teams
- classify-performance.md - Determine performance level

### Improvement

| Prompt | Purpose | Use When |
|--------|---------|----------|
| [create-roadmap.md](./improvement/create-roadmap.md) | Build improvement plan | You want to advance to next performance level |

**Coming Soon:**
- improve-deployment-frequency.md - Increase deploy frequency
- reduce-lead-time.md - Speed up delivery
- lower-cfr.md - Improve quality/stability
- improve-mttr.md - Faster incident recovery

### Reporting

| Prompt | Purpose | Use When |
|--------|---------|----------|
| [executive-summary.md](./reporting/executive-summary.md) | Create leadership report | Quarterly reviews, board meetings, investor updates |

**Coming Soon:**
- team-comparison.md - Compare team performance
- trend-report.md - Show progress over time
- detailed-analysis.md - Deep-dive technical report

### Troubleshooting

| Prompt | Purpose | Use When |
|--------|---------|----------|
| [diagnose-metric-regression.md](./troubleshooting/diagnose-metric-regression.md) | Find why metrics degraded | Performance suddenly drops |

**Coming Soon:**
- root-cause-analysis.md - Deep-dive into specific issues
- bottleneck-identification.md - Find pipeline slowdowns
- incident-pattern-analysis.md - Understand failure patterns

## Usage Tips

### Providing Good Context

The more context you provide, the better the analysis:

**Good:**
```
Calculate DORA metrics:
- Deployments: 45 in 30 days, 4 failures
- Lead time data: [paste JSON]
- Team: 5 engineers, Node.js stack
- Previous quarter: DF=1.2/day, CFR=22%
```

**Better:**
```
Calculate DORA metrics:
- Deployments: [detailed JSON data with timestamps]
- Incidents: [incident data with MTTR]
- Team context: 5 engineers, recently adopted CI/CD
- Business context: Major feature launch mid-month
- Previous quarter comparison data included
- Specific concerns: Our CFR seems high
```

### Iterating on Results

Use follow-up questions to dig deeper:

```
Initial: "Calculate our DORA metrics from this data"
↓
Follow-up: "Why is our lead time in the High range instead of Elite?"
↓
Follow-up: "Create a specific plan to reduce lead time to <1 hour"
↓
Follow-up: "What's the ROI of implementing build caching?"
```

### Combining Prompts

Chain prompts together for comprehensive analysis:

1. Calculate metrics → Understand current state
2. Diagnose regression → Find root causes
3. Create roadmap → Plan improvements
4. Generate summary → Communicate to stakeholders

## DORA Metrics Reference

### Performance Levels

| Metric | Elite | High | Medium | Low |
|--------|-------|------|--------|-----|
| **Deployment Frequency** | On-demand (multiple/day) | Daily to weekly | Weekly to monthly | Monthly to semi-annually |
| **Lead Time for Changes** | < 1 hour | 1 day - 1 week | 1 week - 1 month | 1 month - 6 months |
| **Change Failure Rate** | 0-15% | 16-30% | 31-45% | 46-60% |
| **Time to Restore Service** | < 1 hour | < 1 day | 1 day - 1 week | > 1 week |

### What Each Metric Measures

**Deployment Frequency:**
- Team agility and ability to deliver value
- Deployment automation and confidence
- Batch size (smaller = more frequent)

**Lead Time:**
- Efficiency of delivery pipeline
- Process bottlenecks and waste
- Ability to respond to market changes

**Change Failure Rate:**
- Quality and reliability of deployments
- Effectiveness of testing
- Deployment process maturity

**MTTR (Time to Restore):**
- Incident response capability
- System observability
- Recovery automation and procedures

## Common Patterns

### "We Want to Improve Everything"

**Don't:** Try to improve all four metrics simultaneously

**Do:** Focus on one or two metrics with biggest impact

**Prompt:**
```
We want to improve all our DORA metrics. Help us prioritize.

Current: DF=0.5/day (Medium), LT=2 weeks (Medium), CFR=35% (Medium), MTTR=1 day (Medium)

Which metric should we focus on first and why?
```

### "We're Stuck at High Performance"

**Don't:** Assume incremental improvements will get you to Elite

**Do:** Identify fundamental shifts needed

**Prompt:**
```
We're consistently High performers but can't reach Elite.

Current: DF=1/day, LT=1 day, CFR=20%, MTTR=4 hours

What fundamental changes are needed to reach Elite?
```

### "Management Doesn't Understand DORA"

**Don't:** Use technical jargon and raw metrics

**Do:** Translate to business impact

**Prompt:**
```
Create an executive summary explaining DORA metrics in business terms.

Show how our metrics (DF=3/day, LT=6hrs, CFR=15%, MTTR=1hr)
translate to:
- Faster time to market
- Lower risk
- Better customer experience
```

## Best Practices

### Data Collection

**Automated > Manual:**
- Extract from CI/CD tools automatically
- Use deployment tracking systems
- Instrument your pipelines

**Consistency Matters:**
- Define metrics clearly
- Measure the same way every time
- Document your methodology

**Quality Over Quantity:**
- Accurate data is better than more data
- Validate your data sources
- Note confidence levels

### Analysis

**Look at Trends:**
- Single point-in-time = snapshot
- Trends over time = real insight
- Compare periods meaningfully

**Consider Context:**
- Holidays affect deployment frequency
- Team changes impact all metrics
- Major projects skew data

**Avoid Vanity Metrics:**
- Don't game the numbers
- Focus on meaningful improvements
- Metrics serve the work, not vice versa

### Improvement

**Start Small:**
- Pick one metric to improve
- Focus on quick wins first
- Build momentum with successes

**Measure Impact:**
- Track before and after
- Attribute changes to initiatives
- Learn from what doesn't work

**Cultural Change:**
- Metrics are for learning, not punishment
- Celebrate improvements
- Share learnings across teams

## Troubleshooting

### "The skill says I'm a Low performer but I think we're doing well"

**Check:**
- Are you measuring the right things?
- Do your definitions match DORA's?
- Is your industry fundamentally different?

**Remember:** DORA benchmarks apply across all industries. Elite performers exist in regulated industries too.

### "Our metrics improved but business results didn't"

**Consider:**
- Is there a lag between metrics and business outcomes?
- Are you measuring the right business outcomes?
- Are other factors at play?

**Remember:** DORA metrics correlate with business performance, but aren't the only factor.

### "We can't improve without [tool/budget/people]"

**Try:**
- Identify no-cost improvements first
- Calculate ROI of investments
- Find creative solutions within constraints

**Remember:** Many improvements are process/cultural, not tool-based.

## Resources

### Official DORA Research
- [DORA State of DevOps Reports](https://dora.dev/research/)
- [Accelerate book](https://www.oreilly.com/library/view/accelerate/9781457191435/)
- [DORA capabilities](https://dora.dev/capabilities/)

### Related Skills
- [SBOM Analyzer](../../skills/sbom-analyzer/) - Supply chain security
- [Chalk Build Analyzer](../../skills/chalk-build-analyzer/) - Build performance
- [Certificate Analyzer](../../skills/certificate-analyzer/) - TLS/SSL analysis

## Contributing

Have a useful prompt template? Submit a pull request!

**Good prompt templates:**
- Solve a specific, common problem
- Include clear examples
- Provide expected outputs
- Link to related prompts

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for guidelines.

## Support

- **Questions:** Open an issue or discussion
- **Bug reports:** [GitHub Issues](https://github.com/crashappsec/skills-and-prompts/issues)
- **Contact:** mark@crashoverride.com

---

**Start improving your software delivery performance today!**
