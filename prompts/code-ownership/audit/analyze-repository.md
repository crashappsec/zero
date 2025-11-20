<!--
Copyright (c) 2024 Crash Override Inc
101 Fulton St, 416, New York 10038
SPDX-License-Identifier: GPL-3.0
-->

# Prompt: Analyze Repository Ownership

## Purpose

Perform a comprehensive analysis of code ownership across a git repository to understand who owns what code, identify risks, and generate actionable recommendations.

## When to Use

- Quarterly ownership audits
- New repository assessment
- Team health checks
- Pre-release reviews
- Organizational planning

## Prompt

```
Analyze code ownership for our repository.

Repository: [repository-name or path]
Time Period: Last [90|180|365] days

Please provide:

1. **Executive Summary**
   - Overall ownership health score (0-100)
   - Key findings and highlights
   - Critical concerns requiring immediate attention

2. **Ownership Coverage**
   - What percentage of files have assigned owners?
   - Which components/directories have full coverage?
   - What areas lack ownership?

3. **Ownership Distribution**
   - Who are the top 10 contributors by file count?
   - Is ownership well-distributed or concentrated?
   - Calculate Gini coefficient if possible
   - Any concentration risks (single person owning >15%)?

4. **Owner Activity**
   - How many owners are active (< 30 days)?
   - Any inactive owners (> 90 days)?
   - Staleness by component

5. **Risk Assessment**
   - Single points of failure (SPOFs)
   - Bus factor calculation
   - Knowledge gaps
   - Critical components at risk

6. **CODEOWNERS File Status**
   - Does one exist?
   - If yes: Accuracy compared to actual contributions
   - If no: Should we create one?

7. **Recommendations**
   - Priority 1 (This Week): Urgent actions
   - Priority 2 (This Month): Important improvements
   - Priority 3 (This Quarter): Strategic initiatives

   For each: specific action, effort, impact, suggested owner
```

## Expected Output

- Comprehensive ownership analysis report
- Health score with grading (Excellent/Good/Fair/Poor)
- Coverage and distribution metrics
- Risk assessment with priorities
- Actionable recommendations with timelines

## Variations

### Quick Health Check
```
Quick ownership health check:
- Overall coverage %
- Top 5 contributors
- Any critical risks?
- One key recommendation
```

### Component-Specific Analysis
```
Analyze ownership for the /services/auth directory:
- Who owns this component?
- Is there backup coverage?
- How active are the owners?
- Any risks specific to this component?
```

### Trend Analysis
```
Compare ownership metrics:
- 90 days ago vs today
- Coverage trending up or down?
- Distribution improving or worsening?
- New risks emerged?
```

### Executive Summary Only
```
Generate an executive summary of repository ownership:
- One-paragraph health assessment
- Top 3 risks
- Top 3 recommendations
- For presentation to leadership
```

## Examples

### Example 1: Initial Repository Audit

**Input:**
```
Analyze code ownership for our platform repository.
Last 90 days.
This is our first ownership audit.
```

**Expected Output:**
```
# Code Ownership Analysis: platform

## Executive Summary
Overall Health: 72/100 (Good)

Key Findings:
- 85% of files have clear ownership
- Well-distributed across 23 active contributors
- 3 critical components with single owner (risk)

Critical Concerns:
- Authentication service: No backup owner
- Billing module: Owner inactive 120 days
- No CODEOWNERS file present

## Ownership Coverage: 85%
...
```

### Example 2: Quarterly Review

**Input:**
```
Quarterly ownership review for platform repo:
- Compare to last quarter (90 days vs 180 days ago)
- Has our health score improved?
- Track our Q4 OKR: reach 90% coverage
```

**Expected Output:**
```
# Q4 2024 Ownership Review

## Progress vs Q3
Health Score: 68 → 72 (+4 points)
Coverage: 78% → 85% (+7%)
Bus Factor: 1.8 → 2.1 (improving)

## OKR Progress
Target: 90% coverage by end of Q4
Current: 85%
Gap: 5%
On Track: Yes, if we assign owners to /tests directory

## Improvements Since Q3
✓ Added backup owners for 5 critical components
✓ Updated CODEOWNERS file
✓ Reduced inactive owners from 15 → 8
...
```

## Related Prompts

- [health-assessment.md](./health-assessment.md) - Focused health scoring
- [validate-codeowners.md](../validation/validate-codeowners.md) - CODEOWNERS validation
- [succession-planning.md](../planning/succession-planning.md) - Risk mitigation planning

## Tips

- **Be Specific**: Provide repository path or name
- **Set Context**: Mention if first audit or regular review
- **State Goals**: What are you trying to improve?
- **Request Format**: Specify if you need specific output format (markdown, executive summary, etc.)
- **Time Period**: Adjust days based on repository activity level
  - Active repos: 30-90 days
  - Slower repos: 90-180 days
  - Legacy repos: 180-365 days

---

*Use this prompt to understand your repository's ownership landscape and identify improvement opportunities.*
