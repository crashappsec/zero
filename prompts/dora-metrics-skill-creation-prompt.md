<!--
Copyright (c) 2024 Gibson Powers Contributors

SPDX-License-Identifier: GPL-3.0
-->

# DORA Metrics Skill Creation Prompt

## Objective

Create a comprehensive Claude skill for analyzing, tracking, and improving DORA (DevOps Research and Assessment) metrics within the Crash Override platform. This skill should help engineering teams measure and optimize their software delivery and operational performance.

## Background

DORA metrics are the industry-standard measurements for assessing DevOps performance, based on years of research by the DevOps Research and Assessment team. These metrics are proven predictors of organizational performance and software delivery excellence.

## The Four Key DORA Metrics

### 1. Deployment Frequency (DF)
**Definition:** How often an organization successfully deploys code to production.

**Performance Levels:**
- **Elite:** On-demand (multiple deploys per day)
- **High:** Between once per day and once per week
- **Medium:** Between once per week and once per month
- **Low:** Between once per month and once every six months

**Measurement:** Count of successful production deployments over time period.

**Data Sources:**
- CI/CD pipeline data (GitHub Actions, GitLab CI, Jenkins, CircleCI)
- Deployment logs
- Release management systems
- Container orchestration platforms (Kubernetes, Docker)
- Git commit/tag history
- Chalk build metadata

### 2. Lead Time for Changes (LT)
**Definition:** The time it takes for a commit to get into production.

**Performance Levels:**
- **Elite:** Less than one hour
- **High:** Between one day and one week
- **Medium:** Between one week and one month
- **Low:** Between one month and six months

**Measurement:** Time from commit to production deployment.

**Data Sources:**
- Git commit timestamps
- CI/CD pipeline execution times
- Pull request merge times
- Deployment timestamps
- Issue tracking systems (Jira, Linear, GitHub Issues)
- Code review systems

### 3. Change Failure Rate (CFR)
**Definition:** The percentage of deployments that cause a failure in production requiring remediation.

**Performance Levels:**
- **Elite:** 0-15%
- **High:** 16-30%
- **Medium:** 31-45%
- **Low:** 46-60%

**Measurement:** (Failed deployments / Total deployments) × 100

**Data Sources:**
- Incident management systems (PagerDuty, Opsgenie, VictorOps)
- Application monitoring (Datadog, New Relic, Prometheus)
- Error tracking (Sentry, Rollbar, Bugsnag)
- Rollback/revert logs
- Post-mortem databases
- Support ticket systems

### 4. Time to Restore Service (MTTR)
**Definition:** How long it takes to restore service when an incident or defect impacts users.

**Performance Levels:**
- **Elite:** Less than one hour
- **High:** Less than one day
- **Medium:** Between one day and one week
- **Low:** More than one week

**Measurement:** Time from incident detection to resolution.

**Data Sources:**
- Incident management platforms
- On-call schedules and alerts
- Monitoring and observability tools
- Incident timelines and post-mortems
- Chat/communication logs (Slack, Teams)

## Skill Requirements

### Core Capabilities

1. **Data Collection & Integration**
   - Integrate with common CI/CD platforms (GitHub Actions, GitLab CI, Jenkins, CircleCI)
   - Parse deployment logs and metadata
   - Connect to incident management systems
   - Extract data from Git repositories
   - Process Chalk build artifacts for deployment tracking
   - Support manual data input for systems without APIs

2. **Metric Calculation**
   - Calculate all four DORA metrics accurately
   - Support multiple time periods (daily, weekly, monthly, quarterly, yearly)
   - Handle partial data gracefully
   - Provide confidence indicators based on data completeness
   - Support team-level and organization-level aggregation

3. **Benchmarking & Classification**
   - Classify performance levels (Elite, High, Medium, Low)
   - Compare against DORA research benchmarks
   - Track performance trends over time
   - Identify improvement opportunities
   - Set custom organizational benchmarks

4. **Analysis & Insights**
   - Identify patterns and anomalies
   - Correlate metrics (e.g., how DF affects CFR)
   - Detect performance regressions
   - Highlight best-performing teams
   - Root cause analysis for metric degradation

5. **Reporting & Visualization**
   - Executive summaries for leadership
   - Detailed metric breakdowns for engineering teams
   - Trend visualization (textual descriptions for Mermaid diagrams)
   - Comparison reports (team vs. team, period vs. period)
   - Custom report generation

6. **Improvement Recommendations**
   - Actionable suggestions based on current performance
   - Prioritized improvement roadmap
   - Best practices for each metric
   - Example implementations from high performers
   - Link to relevant research and case studies

### Advanced Features

1. **Team Comparisons**
   - Compare multiple teams side-by-side
   - Identify organizational bottlenecks
   - Share best practices across teams

2. **Predictive Analytics**
   - Forecast future performance based on trends
   - Predict impact of proposed changes
   - Alert on concerning trends before they become critical

3. **Custom Metrics**
   - Allow users to define additional metrics
   - Support custom performance thresholds
   - Industry or domain-specific benchmarks

4. **Integration Workflows**
   - Automated metric collection pipelines
   - Scheduled reporting
   - Alert thresholds and notifications
   - Export to BI tools (Tableau, Looker, etc.)

## Data Source Examples

### GitHub Actions
```yaml
# Extract deployment data from workflow runs
- Workflow completion events
- Deployment job timestamps
- Success/failure status
- Commit SHAs and authors
- Environment targets (production, staging)
```

### GitLab CI
```yaml
# Extract from pipeline data
- Pipeline start/end times
- Deployment job status
- Commit information
- Environment deployment tracking
```

### Incident Management (PagerDuty)
```json
{
  "incident_id": "ABC123",
  "created_at": "2024-11-20T10:00:00Z",
  "resolved_at": "2024-11-20T10:45:00Z",
  "severity": "high",
  "service": "api-gateway"
}
```

### Git History
```bash
# Extract commit to deployment data
git log --format="%H|%an|%ae|%ai|%s"
# Track deployment tags
git tag -l "deploy-*" --sort=-creatordate
```

## Output Format Examples

### Executive Summary
```
DORA Metrics Report - Q4 2024
Team: Platform Engineering

Overall Performance: HIGH PERFORMER

Deployment Frequency:    ELITE     (5.2 deploys/day)
Lead Time for Changes:   HIGH      (4.3 hours average)
Change Failure Rate:     ELITE     (8.2%)
Time to Restore Service: ELITE     (32 minutes average)

Key Achievements:
✓ Moved from High to Elite in Deployment Frequency
✓ Maintained elite-level Change Failure Rate
✓ Improved MTTR by 40% from last quarter

Top Priorities:
1. Reduce Lead Time to elite level (<1 hour)
2. Maintain current performance levels
3. Share deployment practices with Backend team
```

### Detailed Metric Analysis
```
Deployment Frequency Analysis
=============================

Current Performance: ELITE (5.2 deploys/day)

Trend: ↑ Improving
- Previous month: 4.8 deploys/day
- 3 months ago: 3.2 deploys/day
- Year ago: 1.4 deploys/day

Breakdown by Week:
Week 1: 6.1 deploys/day
Week 2: 5.8 deploys/day
Week 3: 4.2 deploys/day (holiday week)
Week 4: 4.8 deploys/day

Contributing Factors:
✓ Automated deployment pipeline (100% of deploys)
✓ Feature flags enabling smaller, safer releases
✓ Improved test coverage (92%)
✓ Reduced PR review time (avg 2.3 hours)

Recommendations:
- Continue current practices
- Document deployment process for other teams
- Monitor for signs of over-deployment (quality issues)
```

### Improvement Roadmap
```
DORA Metrics Improvement Roadmap
================================

Current State: HIGH PERFORMER
Target: ELITE PERFORMER
Timeline: 3-6 months

Priority 1: Reduce Lead Time (HIGH → ELITE)
Current: 4.3 hours | Target: <1 hour

Actions:
1. Implement continuous deployment (eliminate manual approvals)
   - Effort: Medium | Impact: High | Timeline: 2 weeks
2. Optimize CI pipeline (reduce build time by 50%)
   - Effort: Medium | Impact: Medium | Timeline: 3 weeks
3. Improve automated test reliability
   - Effort: High | Impact: High | Timeline: 6 weeks

Expected Outcome: Lead time reduced to 45 minutes

Priority 2: Maintain Elite-Level Performance
- Continue investment in testing infrastructure
- Regular deployment pipeline optimization
- Ongoing incident response training
```

## Integration with Crash Override Platform

### Skill Invocation Examples

```
User: "Show me our DORA metrics for last month"

Skill Response:
- Parse request for time period (last month)
- Collect data from connected sources
- Calculate all four metrics
- Classify performance levels
- Generate executive summary
- Provide trend analysis
- Offer improvement recommendations
```

```
User: "Why is our Change Failure Rate so high?"

Skill Response:
- Analyze CFR trend over time
- Identify recent spikes in failures
- Correlate with deployment patterns
- Review incident data for common causes
- Compare with test coverage trends
- Provide root cause analysis
- Suggest specific improvements
```

```
User: "Compare our Platform team vs Backend team DORA metrics"

Skill Response:
- Calculate metrics for both teams
- Create side-by-side comparison
- Identify performance gaps
- Highlight best practices from higher performer
- Suggest knowledge sharing opportunities
- Provide team-specific recommendations
```

### Data Source Configuration

```
User: "Connect to our GitHub Actions for deployment tracking"

Skill Response:
- Guide user through GitHub API authentication
- Identify relevant workflows
- Configure deployment detection patterns
- Test data extraction
- Confirm metric calculation
- Schedule automated updates
```

## Implementation Considerations

### Accuracy & Reliability
- Validate data sources before calculation
- Handle missing or incomplete data gracefully
- Provide confidence scores with metrics
- Flag potential data quality issues
- Support manual corrections/overrides

### Privacy & Security
- Respect data access permissions
- Anonymize individual developer data (focus on team metrics)
- Secure API credentials and tokens
- Support on-premise/self-hosted data sources

### Scalability
- Handle large organizations (100+ teams)
- Support high deployment volumes (1000+ deploys/day)
- Efficient data processing and caching
- Incremental updates vs. full recalculation

### Usability
- Clear, jargon-free explanations
- Actionable recommendations
- Context-sensitive help
- Progressive disclosure (summary → details)

## Success Criteria

The DORA metrics skill should:

1. **Accurately Calculate Metrics**
   - All four DORA metrics computed correctly
   - Proper classification against benchmarks
   - Handle edge cases and partial data

2. **Provide Actionable Insights**
   - Clear improvement recommendations
   - Prioritized action plans
   - Specific, implementable suggestions

3. **Support Multiple Use Cases**
   - Executive reporting
   - Engineering team optimization
   - Cross-team comparison
   - Trend analysis and prediction

4. **Integrate Seamlessly**
   - Connect to common platforms easily
   - Automated data collection
   - Minimal manual intervention

5. **Drive Improvement**
   - Help teams progress through performance levels
   - Track improvement over time
   - Celebrate wins and maintain momentum

## Example Skill Prompts

### Basic Queries
- "What are our current DORA metrics?"
- "How did we perform last quarter?"
- "Show me deployment frequency trends"
- "What's our average time to restore service?"

### Analysis Queries
- "Why has our lead time increased?"
- "What's causing our high change failure rate?"
- "How do we compare to industry benchmarks?"
- "Which team has the best DORA metrics?"

### Improvement Queries
- "How can we improve our deployment frequency?"
- "What's the fastest path to elite performer status?"
- "Give me a roadmap to reduce our MTTR"
- "What should we prioritize this quarter?"

### Configuration Queries
- "Connect to our CI/CD pipeline"
- "Add PagerDuty for incident tracking"
- "Set up automated weekly reports"
- "Configure custom benchmarks for our industry"

## Research & References

- DORA State of DevOps Reports (2014-2025)
- "Accelerate" by Nicole Forsgren, Jez Humble, Gene Kim
- DORA Core capabilities model
- Industry benchmarks by sector
- Academic research on software delivery performance

## Deliverables

1. **Skill File** (.skill or .md)
   - Complete DORA metrics knowledge base
   - Calculation methodologies
   - Benchmarking data
   - Integration patterns
   - Prompt examples

2. **Documentation**
   - README with usage examples
   - Metric definitions and calculations
   - Data source integration guides
   - Troubleshooting guide
   - FAQ

3. **Automation Scripts** (optional)
   - Data collection scripts for common platforms
   - Report generation tools
   - Comparison utilities

4. **Examples**
   - Sample DORA metric reports
   - Before/after improvement case studies
   - Team comparison examples
   - Trend analysis examples

---

**Use this prompt to guide the creation of a comprehensive DORA metrics skill that helps engineering teams measure, understand, and improve their software delivery performance using the Crash Override platform.**
