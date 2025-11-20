<!--
Copyright (c) 2024 Crash Override Inc
101 Fulton St, 416, New York 10038
SPDX-License-Identifier: GPL-3.0
-->

# Create DORA Improvement Roadmap

## Purpose

Generate a prioritized, actionable roadmap to improve DORA metrics and advance to the next performance level.

## Prompt Template

```
Create a DORA metrics improvement roadmap for our team.

Current Performance:
- Deployment Frequency: [value] ([classification])
- Lead Time: [value] ([classification])
- Change Failure Rate: [value] ([classification])
- MTTR: [value] ([classification])

Target Performance Level: [Elite/High/Medium]
Timeline: [3 months/6 months/1 year]

Context:
- Team size: [number]
- Tech stack: [languages, frameworks]
- Current practices: [brief description]
- Biggest pain points: [deployment speed/quality/recovery/etc.]

Please provide:

1. Priority Assessment
   - Which metric to focus on first
   - Rationale for prioritization
   - Expected impact on overall performance

2. Phased Improvement Plan
   - Quick wins (1-4 weeks)
   - Medium-term improvements (1-3 months)
   - Long-term initiatives (3-6 months)

3. Specific Action Items
   - What to implement
   - Estimated effort (Low/Medium/High)
   - Expected impact on metrics
   - Dependencies and prerequisites

4. Success Metrics
   - How to measure progress
   - Milestones to track
   - Target values for each phase

5. Resources Needed
   - Tools or services
   - Training requirements
   - Time allocation

6. Risk Mitigation
   - Potential challenges
   - How to address them
```

## Example Usage

**Moving from Medium to High:**
```
Create a roadmap to move our team from Medium to High performer.

Current: DF=0.5/day, LT=3 weeks, CFR=35%, MTTR=1.5 days
Target: High performer across all metrics
Timeline: 6 months
Team: 5 engineers, Node.js/React stack
Pain point: Manual QA testing is our biggest bottleneck
```

**Achieving Elite Status:**
```
We're High performers and want to reach Elite. Create a focused roadmap.

Current: DF=2/day (High), LT=4 hours (High), CFR=18% (High), MTTR=2 hours (High)
Target: Elite across all four metrics
Timeline: 3 months
We have strong automation but need optimization
```

**Single Metric Focus:**
```
Help us improve just our Lead Time from 2 weeks to <1 day.

Current LT breakdown:
- Code review: 3 days
- CI/CD: 1 day
- Manual testing: 7 days
- Deployment approval: 3 days

Create a focused improvement plan.
```

## Expected Output

The skill should provide:
- Clear prioritization with rationale
- Phased action plan with specific tasks
- Effort and impact estimates
- Success metrics and milestones
- Resource requirements
- Risk assessment
- Timeline with dependencies

## Related Prompts

- [Improve Deployment Frequency](./improve-deployment-frequency.md)
- [Reduce Lead Time](./reduce-lead-time.md)
- [Lower Change Failure Rate](./lower-cfr.md)
- [Improve MTTR](./improve-mttr.md)
