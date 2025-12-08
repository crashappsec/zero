<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Prompt: Analyze Repository Ownership

## Purpose

Perform a comprehensive analysis of code ownership across a git repository to understand who owns what code and identify risks through objective, data-driven assessment.

**Analysis Approach**: This prompt focuses on providing factual observations about ownership patterns, coverage, and risks. It does NOT generate recommendations or action items - use specific planning prompts (knowledge-transfer.md, succession-planning.md) for prescriptive guidance.

## When to Use

- Quarterly ownership audits
- New repository assessment
- Team health checks
- Pre-release reviews
- Organizational planning

## Prompt

```
Analyze code ownership for our repository.

Repository: [repository-name, path, or Git URL]
Time Period: Last [90|180|365] days

Please provide:

1. **Executive Summary**
   - Overall ownership health score (0-100)
   - Key findings and highlights
   - Critical risks identified

2. **Ownership Coverage**
   - What percentage of files have assigned owners?
   - Which components/directories have full coverage?
   - What areas lack ownership?

3. **Ownership Distribution**
   - Who are the top 10 contributors by file count?
   - Include GitHub usernames (if available) in format: Name (@github_username)
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
   - If no: Note the absence
   - Gap analysis between file and actual contributions

IMPORTANT: Provide ONLY factual analysis and observations. Do NOT include:
- Recommendations or action items
- Implementation priorities or timelines
- Suggested fixes or improvements
- "Should" or "must" statements

Focus exclusively on what IS, not what should be done about it.
```

## Expected Output

- Comprehensive ownership analysis report
- Health score with grading (Excellent/Good/Fair/Poor)
- Coverage and distribution metrics with GitHub usernames
- Risk assessment with severity ratings
- CODEOWNERS file accuracy analysis

**Note**: For actionable recommendations and planning, use dedicated planning prompts:
- [knowledge-transfer.md](../planning/knowledge-transfer.md) - Transfer planning
- [succession-planning.md](../planning/succession-planning.md) - Risk mitigation
- [recommend-reviewers.md](../optimization/recommend-reviewers.md) - PR reviewer suggestions

## Variations

### Quick Health Check
```
Quick ownership health check:
- Overall coverage %
- Top 5 contributors (with GitHub usernames if available)
- Any critical risks identified?
- Bus factor score
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
- Top 3 risks identified
- Bus factor and distribution metrics
- For presentation to leadership
```

## Examples

### Example 1: Initial Repository Audit

**Input:**
```
Analyze code ownership for our platform repository.
Repository: https://github.com/myorg/platform
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
- Well-distributed across 23 active contributors (@alice, @bob, @charlie, etc.)
- 3 critical components with single owner (risk)

Critical Risks Identified:
- Authentication service: No backup owner (Alice Smith @alice only)
- Billing module: Owner inactive 120 days (Bob Jones @bjones)
- No CODEOWNERS file present

## Ownership Coverage: 85%
- 1,247 of 1,467 files have identifiable owners
- Core services: 95% coverage
- Tests directory: 62% coverage (gap identified)
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
Gap: 5% (primarily /tests directory which has 62% coverage)

## Changes Since Q3
- Backup owners added for 5 critical components
- CODEOWNERS file updated (accuracy improved from 64% to 87%)
- Inactive owners reduced from 15 to 8
- Top contributors now include GitHub usernames for easier identification
...
```

## Related Prompts

- [health-assessment.md](./health-assessment.md) - Focused health scoring
- [validate-codeowners.md](../validation/validate-codeowners.md) - CODEOWNERS validation
- [succession-planning.md](../planning/succession-planning.md) - Risk mitigation planning

## Tips

- **Be Specific**: Provide repository path, name, or Git URL
- **Git URL Support**: CLI scripts accept GitHub/GitLab/Bitbucket URLs directly (auto-clones with full history)
- **Set Context**: Mention if first audit or regular review
- **State Goals**: What metrics are you tracking?
- **Request Format**: Specify if you need specific output format (markdown, executive summary, etc.)
- **Time Period**: Adjust days based on repository activity level
  - Active repos: 30-90 days
  - Slower repos: 90-180 days
  - Legacy repos: 180-365 days
- **Full Clone Required**: Never use shallow clones (--depth 1) - ownership analysis requires complete commit history
- **GitHub Profiles**: Request contributor GitHub usernames for easier team identification
- **Objective Analysis**: This prompt provides factual assessment only - use planning prompts for recommendations

---

*Use this prompt to understand your repository's ownership landscape and identify improvement opportunities.*
