<!--
Copyright (c) 2024 Crash Override Inc
101 Fulton St, 416, New York 10038
SPDX-License-Identifier: GPL-3.0
-->

# Code Ownership Analysis Skill

Comprehensive code ownership analysis for understanding who owns what code, validating CODEOWNERS files, identifying risks, and optimizing code review processes.

## Purpose

This skill enables teams to:

- **Analyze Ownership**: Understand code ownership patterns across repositories
- **Validate CODEOWNERS**: Check accuracy and completeness of CODEOWNERS files
- **Assess Risk**: Identify single points of failure and knowledge gaps
- **Plan Succession**: Prepare for team member transitions
- **Optimize Reviews**: Recommend best reviewers for pull requests
- **Track Health**: Monitor ownership metrics and trends

## Key Capabilities

### 1. Repository Ownership Analysis
- Parse git history to identify contributors
- Calculate ownership scores based on commits, reviews, and recency
- Generate ownership maps by component/directory
- Measure coverage (% of files with owners)
- Analyze distribution (concentration vs. balance)
- Track owner activity and staleness

### 2. CODEOWNERS File Management
- Validate syntax across GitHub/GitLab/Bitbucket formats
- Compare CODEOWNERS against actual contribution patterns
- Identify inactive or non-existent owners
- Detect missing ownership coverage
- Generate updated CODEOWNERS files
- Suggest improvements and corrections

### 3. Risk and Health Assessment
- Calculate bus factor (knowledge concentration risk)
- Identify single points of failure (SPOF)
- Detect knowledge gaps and silos
- Assess owner engagement and responsiveness
- Score overall ownership health (0-100)
- Track trends over time

### 4. Knowledge Transfer Planning
- Inventory departing team member's ownership
- Prioritize components by criticality and complexity
- Recommend successors based on familiarity
- Create detailed transfer timelines
- Identify documentation gaps
- Generate handoff checklists

### 5. PR Review Optimization
- Recommend best reviewers based on ownership
- Consider reviewer availability and workload
- Suggest required vs. optional reviewers
- Identify backup reviewers
- Balance review distribution

## Prerequisites

- Repository with git history
- Optional: Existing CODEOWNERS file (for validation)
- Optional: GitHub/GitLab API access (for enhanced analysis)

## Usage

### With the Skill

Load the Code Ownership Analysis skill in Crash Override and use natural language:

#### Repository Audit

```
Analyze code ownership for our repository.
Show me:
- Overall ownership coverage
- Top contributors
- Any single points of failure
- Inactive owners
```

#### CODEOWNERS Validation

```
Validate our CODEOWNERS file at .github/CODEOWNERS.
Check for:
- Syntax errors
- Inactive or non-existent owners
- Missing coverage
- Accuracy compared to actual contributions
```

#### Knowledge Transfer

```
@alice is leaving in 3 weeks. Help me plan the knowledge transfer.
What code does she own?
Who should take over?
What's the priority order?
```

#### Reviewer Recommendation

```
I have a PR changing these files:
- src/auth/oauth.ts
- src/api/users/profile.ts
- database/migrations/20241120_oauth.sql

Who should review this?
```

#### Generate CODEOWNERS

```
We don't have a CODEOWNERS file. Generate one based on the last 90 days of git history.
Use team-based ownership where it makes sense.
```

## Analysis Workflow

The skill follows a systematic approach:

1. **Parse Repository**
   - Analyze git log (commits, authors, dates)
   - Extract PR and review data (if available)
   - Parse existing CODEOWNERS (if present)

2. **Calculate Ownership**
   - Score contributors by commit frequency, LOC, reviews, recency
   - Weight recent contributions higher (exponential decay)
   - Identify primary and backup owners
   - Map ownership to files/directories

3. **Assess Health**
   - Calculate coverage metrics
   - Measure distribution (Gini coefficient)
   - Check owner activity/staleness
   - Compute overall health score

4. **Identify Risks**
   - Detect single points of failure
   - Calculate bus factor
   - Find knowledge gaps
   - Flag inactive owners

5. **Generate Recommendations**
   - Prioritize actions by impact and effort
   - Suggest specific ownership changes
   - Provide implementation guidance
   - Estimate effort required

## Output Formats

### Ownership Analysis Report
- Executive summary with health score
- Coverage breakdown by component
- Distribution analysis (top owners, concentration)
- Activity analysis (active/inactive owners)
- Risk assessment (critical/high/medium)
- Prioritized recommendations

### CODEOWNERS Validation Report
- Syntax validation results
- Accuracy score with specific issues
- Missing pattern identification
- Suggested corrections
- Generated updated CODEOWNERS file

### Knowledge Transfer Plan
- Ownership inventory for departing member
- Prioritized component list
- Suggested successors with reasoning
- Week-by-week timeline
- Documentation gap analysis
- Verification checklist

### Reviewer Recommendations
- Primary reviewers (required approval)
- Secondary reviewers (optional)
- Backup reviewers (if primary unavailable)
- Reasoning for each suggestion
- Availability and workload considerations

## Metrics and Scoring

### Ownership Score

```
owner_score = (
    commit_frequency * 0.30 +
    lines_contributed * 0.20 +
    review_participation * 0.25 +
    recency * 0.15 +
    consistency * 0.10
)
```

### Health Score (0-100)

```
health_score = (
    coverage * 0.35 +        # % of files with owners
    distribution * 0.25 +    # Balance (1 - Gini coefficient)
    freshness * 0.20 +       # % of active owners
    engagement * 0.20        # % of responsive owners
)
```

**Grading:**
- **Excellent:** 85-100 (Well-managed, low risk)
- **Good:** 70-84 (Healthy, minor improvements needed)
- **Fair:** 50-69 (Needs attention, moderate risk)
- **Poor:** <50 (Critical issues, high risk)

### Bus Factor

Minimum number of team members who must leave before project knowledge is critically lost.

**Thresholds:**
- **Healthy:** >3 (multiple owners for most components)
- **Acceptable:** 3 (reasonable redundancy)
- **Risky:** 2 (limited backup)
- **Critical:** 1 (single point of failure)

## Examples

### Example Reports

- [Ownership Analysis Report](./examples/example-ownership-analysis.md) - Complete repository audit
- [CODEOWNERS Validation Report](./examples/example-codeowners-validation.md) - Validation with issues and fixes
- [Updated CODEOWNERS File](./examples/example-updated-codeowners.txt) - Generated file example

## Common Use Cases

### Quarterly Ownership Audit
Regular health check of code ownership:
```
Generate a quarterly ownership health report:
- Current health score and trends
- New risks since last quarter
- Ownership distribution changes
- Recommendations for next quarter
```

### New Hire Onboarding
Help new team members understand ownership:
```
I'm new to the team. Create an ownership map showing:
- Who owns each major component
- Who to ask about the authentication system
- What areas have good knowledge distribution
- Where we might need backup owners
```

### Pre-Release Checklist
Validate ownership before major release:
```
We're releasing v2.0 next month. Check:
- Do all new features have clear owners?
- Are there any SPOFs in critical paths?
- Is our CODEOWNERS file up to date?
- Do we have adequate review coverage?
```

### Team Restructuring
Plan ownership changes during reorganization:
```
We're splitting the backend team into Services and Infrastructure teams.
Help me:
- Identify which files each new team should own
- Update CODEOWNERS accordingly
- Find any gaps in the new structure
```

## Best Practices

### Ownership Principles
- **Clear but Not Single**: Primary owner + backup for critical code
- **Team-Based for Shared**: Use teams for infrastructure, tooling
- **Active Over Historical**: Recent contributors matter more
- **Balance Distribution**: No single person >15% of codebase
- **Document Exceptions**: Some files legitimately have no owner

### CODEOWNERS Maintenance
- **Review Quarterly**: Set calendar reminders (Mar, Jun, Sep, Dec)
- **Update on Changes**: Immediately when people join/leave
- **Start Broad**: Directory-level first, refine as needed
- **Use Teams**: Reduces maintenance burden
- **Version Control**: Treat changes like code (review, test)

### Metrics and Monitoring
- **Track Trends**: Coverage, distribution, staleness over time
- **Set Thresholds**: Alert when metrics exceed ranges
- **Context Matters**: Adjust expectations by repo type/size
- **Make Visible**: Dashboard ownership metrics
- **Regular Audits**: Full audit annually minimum

### Knowledge Sharing
- **Identify Silos Early**: Proactive detection
- **Facilitate Pairing**: Connect experts with learners
- **Require Docs**: Critical single-owner code needs documentation
- **Rotation Programs**: Encourage periodic ownership rotation
- **Review Participation**: Non-owners should review owned code

## Integration Opportunities

### GitHub Actions

```yaml
name: Code Ownership Check

on:
  pull_request:
    paths: ['.github/CODEOWNERS']
  schedule:
    - cron: '0 0 1 * *'  # Monthly

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0
      - name: Validate CODEOWNERS
        run: |
          # Use skill to validate CODEOWNERS
          # Check accuracy, completeness
          # Alert on issues
```

### Pre-commit Hook

```bash
#!/bin/bash
# Validate CODEOWNERS before commit

if git diff --cached --name-only | grep -q "CODEOWNERS"; then
    echo "CODEOWNERS modified, validating..."
    # Run validation
    # Block if critical issues found
fi
```

### Slack Bot

```
/ownership @alice
> @alice owns 234 files across 5 components:
> - /services/auth (78 files)
> - /api/users (67 files)
> ...
> Health: ⚠️ No backup owner for auth service
```

## Limitations

- **Requires Git History**: Accurate analysis needs complete git log
- **Platform-Specific**: CODEOWNERS format varies by platform
- **Context Needed**: Can't always determine ownership intent
- **API Rate Limits**: GitHub/GitLab APIs have rate limits
- **Heuristic-Based**: Ownership scoring is probabilistic, not definitive

## Troubleshooting

### "Can't determine primary owner"
- Multiple contributors with similar commit counts
- Check recency - recent activity may clarify
- Consider review patterns, not just commits
- May legitimately need co-ownership

### "Health score seems low"
- Common for new/small repos or specialized projects
- Check individual metrics (coverage, distribution, etc.)
- Consider context (acceptable varies by project type)
- Focus on trends, not absolute scores

### "Generated CODEOWNERS too specific"
- Adjust to broader directory patterns
- Use team-based ownership
- Consolidate similar patterns
- Remove low-confidence suggestions

## Resources

### Documentation
- [GitHub CODEOWNERS](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners)
- [GitLab Code Owners](https://docs.gitlab.com/ee/user/project/codeowners/)
- [Bus Factor Research](https://en.wikipedia.org/wiki/Bus_factor)

### Related Skills
- [Bus Factor Analysis](../../ROADMAP.md#2-bus-factor-analysis) - Complementary risk analysis
- [DORA Metrics](../dora-metrics/) - Team performance metrics

## Contributing

Improvements to this skill are welcome! Consider contributing:
- Additional CODEOWNERS format support
- Enhanced ownership algorithms
- New metric calculations
- Example reports and analyses

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for guidelines.

## License

This skill is licensed under GPL-3.0. See [LICENSE](../../LICENSE) for details.

## Support

For questions, issues, or feature requests:
- Open an issue in the [GitHub repository](https://github.com/crashappsec/skills-and-prompts-and-rag/issues)
- Review existing [discussions](https://github.com/crashappsec/skills-and-prompts-and-rag/discussions)
- Contact: mark@crashoverride.com

---

**Understand who owns your code and improve your team's ownership practices today!**
