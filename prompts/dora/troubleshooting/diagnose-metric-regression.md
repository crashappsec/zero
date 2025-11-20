<!--
Copyright (c) 2024 Crash Override Inc
101 Fulton St, 416, New York 10038
SPDX-License-Identifier: GPL-3.0
-->

# Diagnose Metric Regression

## Purpose

Investigate why a DORA metric has regressed and identify root causes.

## Prompt Template

```
Help me diagnose why our [metric name] has regressed.

Metric: [Deployment Frequency/Lead Time/CFR/MTTR]

Performance Change:
- Previous: [value] ([classification])
- Current: [value] ([classification])
- Change: [percentage or absolute change]
- Time period: [when regression occurred]

Context:
- Team changes: [new members, departures, reorganization]
- Process changes: [new tools, policy changes, workflow updates]
- Technical changes: [new services, tech stack changes]
- Business changes: [new features, priorities, deadlines]

Additional Data (if available):
[Paste relevant data: deployment logs, incident reports, etc.]

Please analyze:

1. Root Cause Analysis
   - What changed to cause this regression?
   - Are there multiple contributing factors?
   - Is this a temporary blip or systemic issue?

2. Pattern Detection
   - When did the regression start?
   - Is it getting worse or stable?
   - Are there related metrics also affected?

3. Impact Assessment
   - How severe is this regression?
   - What's the business impact?
   - What risks does it introduce?

4. Remediation Plan
   - What should we do immediately?
   - What are longer-term fixes?
   - How do we prevent recurrence?

5. Monitoring
   - What should we track closely?
   - What are early warning signs?
   - When should we expect recovery?
```

## Example Usage

**Deployment Frequency Regression:**
```
Our deployment frequency dropped from Elite to Medium in just 2 weeks.

Previous: 5.2 deploys/day (Elite)
Current: 0.8 deploys/week (Medium)
Change: -96% deployments

Context:
- New security policy requiring manual sign-off
- Two senior engineers on vacation
- Major feature development (larger PRs)

What's causing this dramatic drop and how do we fix it?
```

**Lead Time Increase:**
```
Lead time jumped from 4 hours to 3 days.

Previous: 4 hours (High)
Current: 3 days (Medium)
Change: +1700%

Noticed:
- CI pipeline hasn't changed
- More PRs waiting for review
- New approval process for prod deploys

Help me identify the bottleneck.
```

**CFR Spike:**
```
Change failure rate suddenly increased from 12% to 35%.

Previous: 12% (Elite)
Current: 35% (Medium)
Timeframe: Last 2 weeks

What changed:
- Deployed 15 times
- 5 failures (all different issues)
- No process changes that I know of

Why are we suddenly having so many failures?
```

**MTTR Degradation:**
```
Time to restore service has doubled.

Previous: 30 minutes (Elite)
Current: 2 hours (High)
Change: +300%

Context:
- Team is same size
- On-call rotation unchanged
- No major new incidents types
- Same monitoring tools

What's slowing down our incident response?
```

## Expected Output

The skill should provide:
- Clear identification of root causes
- Data-driven analysis of contributing factors
- Assessment of whether regression is temporary or systemic
- Prioritized remediation steps
- Preventive measures for future
- Monitoring recommendations
- Timeline for expected recovery

## Analysis Framework

**For each regression, consider:**

1. **People Factors**
   - Team capacity changes
   - Skill gaps
   - Knowledge silos
   - Workload/burnout

2. **Process Factors**
   - New approval gates
   - Policy changes
   - Workflow modifications
   - Communication breakdowns

3. **Technical Factors**
   - Tool changes
   - System complexity
   - Technical debt
   - Infrastructure issues

4. **External Factors**
   - Business pressures
   - Deadline crunches
   - Organizational changes
   - Vendor/dependency issues

## Related Prompts

- [Improve Deployment Frequency](../improvement/improve-deployment-frequency.md)
- [Reduce Lead Time](../improvement/reduce-lead-time.md)
- [Lower CFR](../improvement/lower-cfr.md)
- [Improve MTTR](../improvement/improve-mttr.md)
- [Root Cause Analysis](./root-cause-analysis.md)
