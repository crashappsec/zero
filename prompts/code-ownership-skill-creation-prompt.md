<!--
Copyright (c) 2024 Gibson Powers Contributors

SPDX-License-Identifier: GPL-3.0
-->

# Prompt: Create Code Ownership Analysis Skill

## Objective

Create a comprehensive skill for the Crash Override platform that enables users to analyze, validate, and manage code ownership across repositories. The skill should help teams understand who owns what code, validate CODEOWNERS files, identify knowledge gaps, and improve code review processes.

## Skill Overview

### Name
Code Ownership Analysis

### Purpose
Enable teams to analyze git repositories to identify code owners, validate CODEOWNERS file accuracy, track ownership metrics, and improve knowledge distribution across the codebase.

### Key Capabilities Required

#### 1. Smart Ownership Detection
Analyze git history to automatically identify code owners based on multiple signals:

**Commit Analysis:**
- Most frequent contributors by file/directory
- Recent activity weighting (last 30/60/90 days)
- Lines of code contributed (additions/modifications)
- Commit quality indicators (size, test coverage, documentation)
- Consistency of contributions over time

**Code Review Patterns:**
- PR approval patterns
- Review participation by file/directory
- Response time to review requests
- Quality of review comments (depth, thoroughness)
- Review coverage (what areas each person reviews)

**Ownership Heuristics:**
- Primary author (most commits to a file)
- Recent maintainer (most recent significant changes)
- Domain expert (reviews + commits combined)
- Historical owner (consistent contributions over time)
- Active vs. inactive owners (recency weighting)

**Scoring Algorithm:**
Develop a weighted scoring system that considers:
- Commit frequency: 30%
- Lines contributed: 20%
- Review participation: 25%
- Recency: 15%
- Consistency: 10%

#### 2. CODEOWNERS File Management

**Validation:**
- Parse and validate CODEOWNERS syntax (GitHub, GitLab, Bitbucket formats)
- Check for syntax errors and format issues
- Verify referenced users/teams exist in the organization
- Identify patterns that never match any files
- Flag overly broad patterns (e.g., `*` ownership)
- Detect conflicting ownership rules

**Accuracy Assessment:**
- Compare CODEOWNERS entries against actual contribution patterns
- Identify files with incorrect owners listed
- Find active contributors not listed as owners
- Detect listed owners who are no longer active
- Calculate accuracy percentage per directory/file

**Generation:**
- Auto-generate CODEOWNERS from git history analysis
- Suggest ownership patterns based on directory structure
- Recommend team-based ownership where appropriate
- Generate in GitHub/GitLab/Bitbucket formats
- Include confidence scores for suggestions

**Maintenance:**
- Suggest additions for new code without owners
- Recommend removals for inactive owners
- Propose ownership transfers based on activity shifts
- Generate diff showing proposed changes
- Provide rationale for each suggested change

#### 3. Ownership Metrics and Reporting

**Coverage Metrics:**
- Files with assigned owners vs. without
- Directories with complete ownership vs. gaps
- Percentage of codebase with clear ownership
- Coverage by component/module/service
- Trend analysis (improving/declining over time)

**Distribution Metrics:**
- Number of files per owner
- Concentration risk (% of code owned by top N people)
- Team balance (even distribution vs. imbalance)
- Cross-team ownership patterns
- Shared ownership vs. single owner files

**Quality Metrics:**
- Ownership staleness (time since owner last contributed)
- Owner responsiveness (PR review turnaround time)
- Owner engagement (active vs. passive ownership)
- Knowledge breadth (areas of expertise per person)
- Succession readiness (backup owners identified)

**Health Indicators:**
- Overall ownership health score (0-100)
- Risk areas requiring attention
- Best practices compliance
- Trend indicators (improving/stable/declining)
- Benchmarks against similar projects

#### 4. Knowledge and Risk Analysis

**Single Points of Failure:**
- Files with only one contributor ever
- Critical paths with concentrated ownership
- Areas with no active owners
- Undocumented code with single owner
- Components at risk if owner leaves

**Knowledge Distribution:**
- Map of who knows what
- Knowledge overlap between team members
- Expertise gaps requiring attention
- Mentoring opportunities (expert + learner pairs)
- Rotation candidates for knowledge sharing

**Succession Planning:**
- Identify owners without backups
- Suggest potential successors based on contribution patterns
- Areas needing knowledge transfer
- Prioritized list for documentation/pairing
- Runway before knowledge loss (based on activity)

#### 5. Code Review Optimization

**Reviewer Recommendations:**
- Suggest best reviewers for each PR based on file ownership
- Consider reviewer workload/availability
- Factor in domain expertise and recent contributions
- Identify required vs. optional reviewers
- Suggest round-robin for shared areas

**Review Coverage:**
- Ensure PRs reach appropriate owners
- Flag PRs missing owner reviews
- Track owner participation in reviews
- Identify review bottlenecks
- Suggest review process improvements

#### 6. Integration Capabilities

**Git Platforms:**
- GitHub (commits, PRs, reviews, teams, CODEOWNERS)
- GitLab (commits, MRs, reviews, groups, CODEOWNERS)
- Bitbucket (commits, PRs, reviews, CODEOWNERS)
- Azure DevOps (commits, PRs, reviews)

**CI/CD Integration:**
- Validate CODEOWNERS on every commit
- Check PR has owner approval before merge
- Generate ownership reports in pipeline
- Alert on ownership gaps for new files
- Block merges lacking proper review

**Notification Systems:**
- Slack/Teams notifications for ownership gaps
- Alert when owner becomes inactive
- Notify on CODEOWNERS file changes
- Weekly ownership health reports
- Critical risk alerts

**Data Export:**
- JSON/CSV/YAML output formats
- Markdown reports
- HTML dashboards
- Metrics for visualization tools (Grafana, Datadog)
- API endpoints for integration

## Technical Requirements

### Data Sources

**Git Repository Analysis:**
- Complete git history (commits, authors, timestamps)
- File change patterns (additions, deletions, modifications)
- Branch and merge patterns
- Commit metadata (messages, size, files touched)

**Pull Request/Merge Request Data:**
- PR author and reviewers
- Review comments and approvals
- PR size and complexity
- Time to review and merge
- Files changed per PR

**Organization Data:**
- Team structure and membership
- User account status (active/inactive)
- User contact information
- Team ownership patterns
- Organizational hierarchy

**CODEOWNERS Files:**
- Current CODEOWNERS content
- Historical changes to CODEOWNERS
- Format (GitHub/GitLab/Bitbucket)
- Location in repository

### Analysis Algorithms

**Contribution Scoring:**
```
owner_score = (
    commit_frequency * 0.30 +
    lines_contributed * 0.20 +
    review_participation * 0.25 +
    recency_factor * 0.15 +
    consistency * 0.10
)

recency_factor = exponential_decay(days_since_last_commit, half_life=90)
consistency = 1 - coefficient_of_variation(commits_over_time)
```

**Staleness Detection:**
- Flag owners with no commits in 90+ days
- Warn on owners with no commits in 60+ days
- Consider overall repository activity level
- Adjust thresholds based on project cadence

**Accuracy Calculation:**
```
accuracy = (correct_owners / total_files_with_owners) * 100

correct_owner = actual_top_contributor matches CODEOWNERS entry
```

**Health Score:**
```
health_score = (
    coverage_score * 0.35 +
    distribution_score * 0.25 +
    freshness_score * 0.20 +
    engagement_score * 0.20
)

coverage_score = (files_with_owners / total_files) * 100
distribution_score = 100 - gini_coefficient(files_per_owner)
freshness_score = avg(recency_factors_for_all_owners)
engagement_score = (responsive_owners / total_owners) * 100
```

### Performance Considerations

**Large Repositories:**
- Handle repositories with 100K+ commits
- Efficient git log parsing and analysis
- Incremental analysis (only new commits)
- Caching of analysis results
- Parallel processing where possible

**Rate Limiting:**
- Respect GitHub/GitLab API rate limits
- Implement exponential backoff
- Batch API requests efficiently
- Cache API responses appropriately

**Memory Efficiency:**
- Stream processing for large git histories
- Limit in-memory data structures
- Efficient data structures for graphs
- Garbage collection considerations

## Output Formats

### 1. Ownership Report

**Executive Summary:**
```markdown
# Code Ownership Analysis Report
**Repository:** organization/repository
**Analysis Date:** 2024-11-20
**Commit Range:** Last 90 days (1,234 commits)

## Overall Health: 78/100 (Good)

### Key Metrics
- Coverage: 85% (1,234 of 1,450 files)
- Active Owners: 45 of 52 (87%)
- Avg Files per Owner: 27
- CODEOWNERS Accuracy: 72%

### Top Risks
1. ⚠️ Authentication module: Single owner (alice@) inactive 120 days
2. ⚠️ Database layer: 3 files without owners
3. ⚠️ API gateway: Owner (bob@) owns 234 files (16% of codebase)
```

**Detailed Breakdown:**
- Coverage by directory
- Owner activity matrix
- Top contributors by component
- Review participation rates
- Staleness indicators

### 2. CODEOWNERS Validation Report

```markdown
# CODEOWNERS Validation Report

## Syntax Check: ✅ PASS
- Format: GitHub CODEOWNERS
- Location: .github/CODEOWNERS
- Lines: 123
- Patterns: 87

## Accuracy: 72% (63 of 87 patterns)

### Issues Found

#### ❌ Incorrect Owners (15)
1. `/src/auth/**` → `@alice`
   - Actual top contributor: @bob (145 commits vs 12)
   - Recommendation: Change to @bob or @auth-team

2. `/api/v2/**` → `@charlie`
   - Owner inactive for 180 days
   - Recommendation: Transfer to @diana (current maintainer)

#### ⚠️ Missing Patterns (9)
1. `/services/notifications/` (23 files)
   - No CODEOWNERS entry
   - Suggested owner: @eve (85% of commits)

#### ℹ️ Optimization Opportunities (6)
1. Multiple entries for same owner in `/src/`
   - Could consolidate to `/src/core/**` @alice
```

### 3. Ownership Map

**Visual Representation:**
```markdown
## Ownership Map

### /src (234 files)
├── /auth (45 files) - @alice (Primary), @bob (Backup)
│   └── /oauth (12 files) - @alice (95% commits)
├── /api (123 files) - @charlie (Primary)
│   ├── /v1 (67 files) - @charlie (Legacy, inactive)
│   └── /v2 (56 files) - @diana (Active)
└── /database (66 files) - @eve (Primary), @frank (Backup)

### /tests (345 files)
├── /unit (234 files) - @frank (Primary)
└── /integration (111 files) - NO OWNER ⚠️
```

**Metrics Table:**
```markdown
| Directory | Files | Primary Owner | Backup Owner | Coverage | Health |
|-----------|-------|---------------|--------------|----------|--------|
| /src/auth | 45 | @alice (78%) | @bob (15%) | 100% | ⚠️ 65 |
| /src/api | 123 | @charlie (45%) | @diana (40%) | 95% | ✅ 82 |
| /database | 66 | @eve (89%) | @frank (8%) | 100% | ✅ 92 |
| /tests | 345 | None | None | 0% | ❌ 20 |
```

### 4. Risk Assessment

```markdown
## Ownership Risks

### Critical (Immediate Attention)
1. **Single Point of Failure: Authentication**
   - Owner: @alice
   - Files: 45 critical files
   - Risk: No backup owner, last commit 120 days ago
   - Impact: High - core security component
   - Action: Assign backup owner, schedule knowledge transfer

### High (Address Soon)
2. **Concentrated Ownership: API Layer**
   - Owner: @charlie
   - Files: 234 (16% of codebase)
   - Risk: Overloaded owner, review bottleneck
   - Impact: Medium - delays in reviews
   - Action: Distribute ownership across @diana and @frank

### Medium (Monitor)
3. **No Ownership: Test Suite**
   - Files: 345 integration tests
   - Risk: No clear maintainer
   - Impact: Medium - test quality may degrade
   - Action: Assign test ownership to component owners
```

### 5. Suggested CODEOWNERS

```markdown
# Generated CODEOWNERS
# Confidence: High (based on 90 days of analysis)
# Generated: 2024-11-20

# Core authentication (alice: 78% commits, bob: 15%)
/src/auth/** @alice @bob

# API v2 (diana: 85% commits in last 90 days)
/src/api/v2/** @diana

# Database layer (eve: 89% commits, frank: 8%)
/src/database/** @eve @frank

# Frontend (shared ownership)
/src/frontend/** @frontend-team

# Infrastructure (george: 92% commits)
/.github/** @george
/docker/** @george
/k8s/** @george

# Documentation (multiple contributors, team-based)
/docs/** @tech-writers @engineering

# Tests should be owned by component owners
/tests/auth/** @alice @bob
/tests/api/** @diana
/tests/database/** @eve @frank
```

### 6. Actionable Recommendations

```markdown
## Recommendations

### Priority 1: Critical Ownership Gaps
1. **Assign backup owner for /src/auth**
   - Effort: Medium (2-3 weeks knowledge transfer)
   - Impact: High (reduces bus factor risk)
   - Suggested assignee: @bob (already has 15% contributions)
   - Action: Schedule pairing sessions, documentation review

2. **Transfer ownership of inactive components**
   - /src/api/v1: @charlie → @diana
   - Effort: Low (diana already maintaining)
   - Impact: Medium (improves accuracy)
   - Action: Update CODEOWNERS, notify team

### Priority 2: Distribution Issues
3. **Redistribute @charlie's ownership**
   - Current: 234 files (16%)
   - Target: <100 files (<7%)
   - Distribute to: @diana (API), @frank (shared components)
   - Impact: Reduces review bottleneck

### Priority 3: Coverage Improvements
4. **Add ownership for /tests directory**
   - Assign tests to component owners
   - 345 files currently unowned
   - Impact: Improves test quality and maintenance
```

## Use Cases

### 1. Repository Audit
**Scenario:** Engineering manager wants to understand code ownership across the repository.

**Query:**
```
Analyze our repository and tell me:
- What percentage of our code has clear ownership?
- Who are the top 10 owners by file count?
- What areas have no clear ownership?
- Are there any single points of failure?
```

**Expected Output:**
- Ownership coverage percentage
- Top contributors list with file counts
- List of directories/files without owners
- Risk assessment for critical unowned areas

### 2. CODEOWNERS Validation
**Scenario:** Team wants to validate their CODEOWNERS file is accurate.

**Query:**
```
Review our CODEOWNERS file at .github/CODEOWNERS and check:
- Is the syntax correct?
- Are the listed owners actually the active maintainers?
- Are there any files that should have owners but don't?
- Are any listed owners inactive?
```

**Expected Output:**
- Syntax validation results
- Accuracy score with specific discrepancies
- Suggested additions and removals
- Activity status for all listed owners

### 3. New Hire Onboarding
**Scenario:** New engineer joining the team needs to understand code ownership.

**Query:**
```
I'm new to the team. Can you create a map showing:
- Who owns each major component?
- Who should I ask about the authentication system?
- What areas have multiple knowledgeable people?
- Where might we need more backup owners?
```

**Expected Output:**
- Visual ownership map by component
- Contact list for each area
- Knowledge distribution assessment
- Areas needing attention

### 4. PR Review Assignment
**Scenario:** Developer needs to know who should review their PR.

**Query:**
```
I have a PR that changes files in:
- src/auth/oauth.ts
- src/api/v2/users.ts
- tests/integration/auth_test.go

Who should review this? Who are the owners of these files?
```

**Expected Output:**
- List of file owners for each changed file
- Suggested reviewers ranked by relevance
- Required vs. optional reviewers
- Alternate reviewers if primary is unavailable

### 5. Knowledge Transfer Planning
**Scenario:** Team member leaving, need to plan knowledge transfer.

**Query:**
```
@alice is leaving the company in 4 weeks.
- What code does she own?
- Who would be the best person to transfer ownership to?
- What's the priority order for knowledge transfer?
- How can we ensure continuity?
```

**Expected Output:**
- Complete list of alice's owned files/areas
- Suggested successors based on existing contributions
- Prioritized transfer plan
- Documentation and pairing recommendations

### 6. Team Health Check
**Scenario:** Quarterly review of code ownership health.

**Query:**
```
Generate a quarterly code ownership health report:
- Overall ownership coverage trends
- Changes in ownership distribution
- New single points of failure
- Improvements or regressions
- Recommended actions for next quarter
```

**Expected Output:**
- Trend analysis (coverage, distribution, engagement)
- Quarter-over-quarter comparison
- Risk changes (new/resolved)
- Action items for next quarter

### 7. CODEOWNERS Generation
**Scenario:** New repository needs a CODEOWNERS file.

**Query:**
```
We don't have a CODEOWNERS file yet. Based on the last 6 months of git history:
- Generate a CODEOWNERS file for us
- Use team-based ownership where it makes sense
- Focus on files with clear primary owners
- Flag areas where ownership is unclear
```

**Expected Output:**
- Generated CODEOWNERS file
- Confidence scores for each entry
- Areas needing manual review
- Suggested team groupings

### 8. Pre-merge Validation
**Scenario:** CI/CD check before merging PR.

**Query:**
```
Validate this PR before merge:
- Have all code owners approved?
- Are we adding files without owners?
- Does the CODEOWNERS change look reasonable?
- Any ownership concerns?
```

**Expected Output:**
- Owner approval status
- New files ownership check
- CODEOWNERS diff validation
- Go/no-go recommendation

## Integration Examples

### GitHub Actions Workflow
```yaml
name: Code Ownership Check

on:
  pull_request:
    paths:
      - '.github/CODEOWNERS'
      - 'src/**'

  schedule:
    - cron: '0 0 * * 1'  # Weekly on Monday

jobs:
  ownership_analysis:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0  # Full history

      - name: Validate CODEOWNERS
        run: |
          # Use skill to validate CODEOWNERS
          # Check for syntax errors
          # Verify accuracy

      - name: Check New Files
        if: github.event_name == 'pull_request'
        run: |
          # Check if PR adds files without owners
          # Alert if ownership gaps introduced

      - name: Weekly Report
        if: github.event_name == 'schedule'
        run: |
          # Generate ownership health report
          # Post to Slack
```

### Pre-commit Hook
```bash
#!/bin/bash
# .git/hooks/pre-commit

# Check if CODEOWNERS is being modified
if git diff --cached --name-only | grep -q "CODEOWNERS"; then
    echo "CODEOWNERS modified, validating..."
    # Run validation
    # Block commit if validation fails
fi
```

### Slack Bot Integration
```
/ownership @alice
> @alice owns 45 files across 3 components:
> - /src/auth (45 files, 78% of commits)
> - /src/oauth (12 files, 95% of commits)
> - /tests/auth (23 files, 45% of commits)
>
> Health: ⚠️ Warning - No backup owner for critical auth component
> Last activity: 3 days ago
```

## Best Practices to Include

### 1. Ownership Principles
- **Clear ownership doesn't mean single ownership**: Encourage primary + backup owners
- **Team-based for shared code**: Use teams for infrastructure, tooling, shared libraries
- **Ownership follows expertise**: Align ownership with domain knowledge
- **Active over historical**: Recent contributors > distant past contributors
- **Documented exceptions**: Some files legitimately have no owner (generated code, etc.)

### 2. CODEOWNERS Maintenance
- **Review quarterly**: Set calendar reminders to review ownership
- **Update on team changes**: Immediately update when people join/leave
- **Start broad, refine**: Begin with directory-level, refine to files as needed
- **Use teams for scalability**: Team-based ownership reduces maintenance burden
- **Version control**: Track CODEOWNERS changes like code

### 3. Metrics and Monitoring
- **Track trends**: Coverage, distribution, staleness over time
- **Set thresholds**: Alert when metrics exceed acceptable ranges
- **Context matters**: Adjust expectations by repository type/size
- **Celebrate improvements**: Recognize teams improving ownership health
- **Make it visible**: Dashboard ownership metrics

### 4. Knowledge Sharing
- **Identify knowledge silos early**: Proactive detection and mitigation
- **Facilitate pairing**: Connect experts with learners
- **Documentation requirements**: Require docs for single-owner critical code
- **Rotation programs**: Encourage periodic ownership rotation
- **Review participation**: Non-owners should review owned code

## Success Metrics

### Skill Effectiveness
- **Accuracy**: CODEOWNERS recommendations >80% accepted by teams
- **Coverage improvement**: Teams increase ownership coverage by 20%+
- **Time savings**: Reduce time to find right reviewer by 50%
- **Risk reduction**: Decrease single-owner critical code by 30%

### User Satisfaction
- **Adoption**: >70% of repositories using the skill monthly
- **Feedback**: >4.0/5.0 user satisfaction score
- **Engagement**: Teams act on >50% of recommendations

### Business Impact
- **Faster reviews**: Reduce average PR review time by 25%
- **Better knowledge distribution**: Decrease bus factor risk
- **Improved quality**: Fewer bugs in well-owned code
- **Onboarding**: New hires productive faster with clear ownership

## Example Conversations

### Example 1: Basic Audit
```
User: Analyze code ownership for our repository github.com/acme/platform

Skill: I'll analyze the code ownership for your platform repository. Let me examine the git history and CODEOWNERS file.

[Analysis results...]

# Code Ownership Analysis: acme/platform
**Analysis Date:** 2024-11-20
**Timeframe:** Last 90 days (2,456 commits)
**Total Files:** 3,245

## Overall Health: 68/100 (Needs Improvement)

### Ownership Coverage: 72%
- Files with owners: 2,336 (72%)
- Files without owners: 909 (28%)
- CODEOWNERS accuracy: 64%

### Top Owners (by file count)
1. @sarah - 456 files (14%) - ⚠️ Concentration risk
2. @mike - 234 files (7%)
3. @frontend-team - 198 files (6%)
4. @jennifer - 176 files (5%)
5. @david - 145 files (4%)

### Critical Gaps
1. `/services/billing/` - 145 files, no owner ❌
2. `/infrastructure/k8s/` - 89 files, owner @alex inactive 180 days ⚠️
3. `/api/v1/` - 267 files, owner @legacy-team doesn't exist ❌

### Recommendations
1. **Urgent**: Assign owner for billing service (critical component)
2. **High**: Transfer k8s ownership from inactive @alex
3. **Medium**: Update CODEOWNERS to remove non-existent teams

Would you like me to generate a detailed report or focus on any specific area?
```

### Example 2: CODEOWNERS Validation
```
User: Can you validate our CODEOWNERS file and tell me what's wrong with it?

Skill: I'll validate your CODEOWNERS file for syntax, accuracy, and completeness.

# CODEOWNERS Validation Report

## File Location: `.github/CODEOWNERS`
## Syntax: ✅ PASS (GitHub format)
## Accuracy: 58% (35 of 60 patterns)

### Critical Issues (Must Fix)

#### 1. Non-existent Teams (5 patterns)
```
/api/v1/** @legacy-api-team  ❌ Team doesn't exist
/mobile/** @mobile-squad     ❌ Team doesn't exist
```
**Action**: Update to existing teams or individual owners

#### 2. Inactive Owners (8 patterns)
```
/infrastructure/** @alex  ⚠️ Last commit 180 days ago
/docs/** @technical-writers ⚠️ Last activity 120 days ago
```
**Action**: Transfer to active maintainers

#### 3. Incorrect Ownership (22 patterns)
```
/services/auth/** @sarah
  Actual: @mike (245 commits vs 12 in last 90 days)
  Recommendation: Change to @mike

/frontend/dashboard/** @legacy-team
  Actual: @frontend-team (all commits in last year)
  Recommendation: Change to @frontend-team
```

### Missing Ownership (15 areas)
1. `/services/billing/**` - No entry, 145 files
   Suggested: @jennifer (78% of commits)

2. `/tests/e2e/**` - No entry, 234 files
   Suggested: @qa-team

Would you like me to generate a corrected CODEOWNERS file?
```

### Example 3: Knowledge Transfer Planning
```
User: @sarah is leaving in 3 weeks. Help me plan the knowledge transfer.

Skill: I'll analyze @sarah's code ownership and create a knowledge transfer plan.

# Knowledge Transfer Plan: @sarah

## Ownership Summary
@sarah owns 456 files (14% of codebase) across 12 components.

### Critical Areas (Priority 1 - Week 1)
1. **Authentication Service** (78 files)
   - Complexity: High
   - Risk: Critical (core security)
   - Suggested successor: @mike (already 25% familiar)
   - Actions:
     - Document auth flow and key decisions
     - Pair programming: 4 sessions
     - Review all pending PRs
     - Transfer on-call responsibilities

2. **Payment Gateway Integration** (45 files)
   - Complexity: High
   - Risk: Critical (revenue impact)
   - Suggested successor: @jennifer (billing domain expert)
   - Actions:
     - Document third-party integrations
     - Share API keys and credentials
     - Review error handling patterns
     - Test failover procedures

### Important Areas (Priority 2 - Week 2)
3. **User Profile Service** (123 files)
   - Complexity: Medium
   - Suggested successors: @david (40% familiar)
   - Actions:
     - Code walkthrough
     - Document database schema
     - Review recent changes

4. **Notification System** (67 files)
   - Complexity: Medium
   - Suggested successors: @backend-team (shared ownership)
   - Actions:
     - Document message queue setup
     - Share monitoring dashboard

### Lower Priority (Priority 3 - Week 3)
5-8. [Additional components...]

## Recommended Schedule
**Week 1**: Focus on authentication and payment systems
- Mon-Wed: Authentication deep dive
- Thu-Fri: Payment integration

**Week 2**: User profiles and notifications
- Mon-Tue: Profile service
- Wed-Thu: Notifications
- Fri: Q&A and documentation review

**Week 3**: Wrap-up and transition
- Update CODEOWNERS
- Final knowledge check
- Introduce successors to stakeholders

## Documentation Gaps
- Authentication service lacks architecture doc
- Payment error codes not documented
- Notification retry logic undocumented

Should I create detailed session agendas for each component?
```

### Example 4: PR Reviewer Recommendation
```
User: I have a PR touching these files:
- src/auth/oauth/provider.ts
- src/api/users/profile.ts
- src/database/migrations/20241120_add_oauth.sql

Who should review this?

Skill: I'll identify the best reviewers based on code ownership and expertise.

# PR Reviewer Recommendations

## Required Reviewers (Need approval)

### 1. @mike (Primary - OAuth expertise)
**Relevance**: High
**Files**: `src/auth/oauth/provider.ts`
- Owner of /src/auth/** (78% of commits)
- OAuth domain expert (authored original implementation)
- Avg review time: 4 hours
- **Status**: ✅ Available (last review 2 hours ago)

### 2. @jennifer (Primary - Database expertise)
**Relevance**: High
**Files**: `src/database/migrations/20241120_add_oauth.sql`
- Owner of /src/database/** (92% of commits)
- Migration expert
- Avg review time: 2 hours
- **Status**: ⚠️ Busy (6 pending reviews, suggest @david as backup)

## Optional Reviewers (Good to have)

### 3. @sarah (Secondary - Profile service)
**Relevance**: Medium
**Files**: `src/api/users/profile.ts`
- Owner of /src/api/users/** (65% of commits)
- Recently active in this area (3 commits this week)
- Avg review time: 6 hours

## Backup Reviewers (If primary unavailable)

- @david - Can review database migration (40% familiar)
- @alex - Can review OAuth if @mike unavailable (25% familiar)

## Recommendation
1. Request review from @mike and @jennifer (required)
2. Optional: Add @sarah for profile changes
3. If @jennifer unavailable, request @david for database review

**Tip**: Your PR touches 3 different ownership areas. Consider splitting into smaller PRs for faster reviews.
```

### Example 5: Generate CODEOWNERS
```
User: We don't have a CODEOWNERS file. Can you create one based on our git history?

Skill: I'll analyze your repository and generate a CODEOWNERS file based on contribution patterns.

# Generated CODEOWNERS File

Based on analysis of 2,456 commits over the last 90 days:

```
# Code Owners for acme/platform
# Generated: 2024-11-20
# Review quarterly and update as team changes

# Default owners for everything (fallback)
* @engineering-team

# Infrastructure and DevOps
/.github/** @devops-team
/docker/** @devops-team
/k8s/** @devops-team @mike
/.circleci/** @devops-team
/terraform/** @devops-team

# Core Services
/services/auth/** @mike @david
/services/billing/** @jennifer
/services/notifications/** @backend-team
/services/users/** @sarah @mike

# Frontend
/frontend/** @frontend-team
/frontend/dashboard/** @frontend-team @sarah
/frontend/mobile/** @mobile-team

# API
/api/v2/** @mike @jennifer
/api/graphql/** @sarah @backend-team

# Database
/database/** @jennifer @david
/database/migrations/** @jennifer

# Shared Libraries
/lib/** @backend-team
/lib/auth/** @mike

# Documentation
/docs/** @technical-writers @engineering
README.md @engineering-leads

# Configuration
*.yml @devops-team
*.yaml @devops-team
*.json @engineering-team

# Testing
/tests/** @qa-team
/tests/e2e/** @qa-team @frontend-team
```

## Confidence Scores
- High confidence (>80% commit share): 35 patterns
- Medium confidence (60-80%): 18 patterns
- Low confidence (<60%): 7 patterns ⚠️

## Areas Needing Manual Review
1. `/services/billing/**` - Multiple contributors, no clear primary (suggested @jennifer at 45%)
2. `/lib/utilities/**` - Very distributed ownership
3. `/scripts/**` - Multiple one-off contributors

## Recommendations
1. Review low-confidence patterns with team
2. Consider team-based ownership for shared code
3. Add backup owners for critical components
4. Set quarterly review reminder

Would you like me to explain any specific assignments or adjust the ownership structure?
```

## Implementation Notes

### Skill Development Checklist

**Phase 1: Core Analysis (Week 1-2)**
- [ ] Git history parsing and analysis
- [ ] Commit frequency calculation
- [ ] Contributor identification
- [ ] Basic ownership scoring algorithm
- [ ] File-level ownership detection

**Phase 2: CODEOWNERS Support (Week 3)**
- [ ] CODEOWNERS file parsing (GitHub/GitLab/Bitbucket)
- [ ] Syntax validation
- [ ] Accuracy comparison
- [ ] Generation capability
- [ ] Diff and recommendation engine

**Phase 3: Advanced Metrics (Week 4)**
- [ ] Coverage metrics
- [ ] Distribution analysis
- [ ] Staleness detection
- [ ] Health scoring
- [ ] Trend analysis

**Phase 4: Risk Analysis (Week 5)**
- [ ] Single point of failure detection
- [ ] Knowledge gap identification
- [ ] Succession planning
- [ ] Review optimization
- [ ] Team health assessment

**Phase 5: Integration (Week 6)**
- [ ] GitHub API integration
- [ ] GitLab API integration
- [ ] Report generation
- [ ] Export formats (JSON/CSV/Markdown)
- [ ] CI/CD examples

**Phase 6: Polish (Week 7)**
- [ ] Example conversations
- [ ] Documentation
- [ ] Best practices guide
- [ ] Sample prompts
- [ ] Testing with real repositories

### Testing Strategy

**Test with Multiple Repository Types:**
1. **Small repo** (100-500 files) - Basic functionality
2. **Medium repo** (1,000-5,000 files) - Performance testing
3. **Large monorepo** (10,000+ files) - Scalability testing
4. **Active repo** (100+ commits/month) - Recent activity weighting
5. **Legacy repo** (low activity) - Staleness detection
6. **Multi-team repo** - Complex ownership patterns

**Validation Approach:**
1. Compare generated CODEOWNERS against existing (if available)
2. Validate with repository maintainers
3. Test recommendations with real team members
4. Measure accuracy of reviewer suggestions
5. Track adoption of recommendations

### Data Privacy Considerations

- **No PII storage**: Don't persist user emails or personal data
- **Anonymization option**: Allow anonymous analysis
- **Access control**: Respect repository permissions
- **Data retention**: Clear analysis data after use
- **API key security**: Secure credential handling

## Deliverables

1. **Skill File** (`code-ownership-analyzer.skill`)
   - Complete ownership analysis capability
   - CODEOWNERS validation and generation
   - Risk and health assessment
   - Integration examples

2. **Documentation** (`skills/code-ownership/README.md`)
   - Purpose and capabilities
   - Usage instructions
   - API integration guide
   - Best practices
   - Troubleshooting

3. **Example Data** (`skills/code-ownership/examples/`)
   - Sample ownership reports
   - CODEOWNERS validation examples
   - Risk assessment reports
   - Knowledge transfer plans

4. **Automation Scripts** (`skills/code-ownership/`)
   - `ownership-analyzer.sh` - Basic CLI analysis
   - `ownership-analyzer-claude.sh` - AI-enhanced analysis
   - `codeowners-validator.sh` - Validation tool

5. **Sample Prompts** (`prompts/code-ownership/`)
   - Audit and analysis prompts
   - CODEOWNERS validation prompts
   - Risk assessment prompts
   - Knowledge transfer prompts

## Related Resources

- [GitHub CODEOWNERS documentation](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners)
- [GitLab Code Owners](https://docs.gitlab.com/ee/user/project/codeowners/)
- [Microsoft DevOps - Code Ownership](https://learn.microsoft.com/en-us/azure/devops/repos/git/require-branch-folders)
- Academic papers on code ownership and software quality
- Bus Factor analysis (related skill)

---

**Use this prompt to guide the development of a comprehensive Code Ownership Analysis skill that helps teams understand, validate, and improve their code ownership practices.**