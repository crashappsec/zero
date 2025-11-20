<!--
Copyright (c) 2024 Crash Override Inc
101 Fulton St, 416, New York 10038
SPDX-License-Identifier: GPL-3.0
-->

# Generate Executive Summary

## Purpose

Create a concise, executive-level summary of DORA metrics for leadership reporting.

## Prompt Template

```
Create an executive summary of our DORA metrics performance.

Team/Organization: [name]
Period: [time period]
Audience: [CTO/VPE/Board/etc.]

Current Metrics:
- Deployment Frequency: [value]
- Lead Time for Changes: [value]
- Change Failure Rate: [value]
- Time to Restore Service: [value]

Historical Context (optional):
- Previous period metrics: [values]
- Trends: [improving/stable/declining]

Business Context:
- Key projects this period: [list]
- Team changes: [headcount, reorganizations]
- Significant incidents: [if any]

Please provide:

1. Overall Performance Assessment
   - Performance classification (Elite/High/Medium/Low)
   - One-sentence summary
   - Comparison to industry benchmarks

2. Key Achievements (Top 3)
   - What improved
   - Why it matters
   - Quantified impact

3. Current Focus Areas (Top 3)
   - What needs attention
   - Why it's important
   - Proposed actions

4. Business Impact
   - How metrics correlate with business outcomes
   - Customer/user impact
   - Risk assessment

5. Forward-Looking
   - Next quarter goals
   - Expected improvements
   - Investment needs

Keep it concise: 1-2 pages maximum
Use business language, not technical jargon
Focus on "so what?" not just "what"
```

## Example Usage

**Quarterly Business Review:**
```
Create an executive summary for our Q4 Board presentation.

Engineering Department, Q4 2024
Audience: Board of Directors

Metrics:
- DF: 3.2/day (Elite)
- LT: 6 hours (High)
- CFR: 15% (Elite)
- MTTR: 1 hour (Elite)

Context:
- Launched 3 major features
- Grew team from 20 to 25 engineers
- Had one significant outage (resolved in 45 minutes)

Focus on: How our engineering excellence enables business goals
```

**Monthly Team Update:**
```
Create a summary for our monthly all-hands meeting.

Platform Team, November 2024
Audience: All Engineering

Key wins:
- Hit Elite deployment frequency for first time
- Reduced lead time by 30%
- Zero high-severity incidents

Make it celebrate successes while being honest about challenges
```

**Investor Update:**
```
Generate metrics summary for investor update.

Highlight:
- Engineering efficiency improvements
- Deployment capabilities vs. competitors
- How this enables faster go-to-market

Keep technical details minimal, focus on competitive advantage
```

## Expected Output

The skill should provide:
- Concise overall performance statement
- 3 key achievements with business impact
- 3 focus areas with proposed actions
- Business context and forward-looking statement
- Appropriate language for audience
- Suggested visuals or charts to include

## Formatting

**Should include:**
- Clear section headers
- Bullet points for scannability
- Quantified results where possible
- Trend indicators (↑ ↓ →)
- Color coding for status (✅ ⚠️ ❌)

**Should avoid:**
- Technical jargon
- Dense paragraphs
- Too much detail
- Blame or negativity
- Excuses

## Related Prompts

- [Team Comparison Report](./team-comparison.md)
- [Detailed Analysis Report](../analysis/full-analysis.md)
- [Trend Visualization](./trend-report.md)
