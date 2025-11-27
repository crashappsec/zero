<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
SPDX-License-Identifier: GPL-3.0
-->

# Prioritization Frameworks for Supply Chain Security

## Risk-Based Prioritization

### CVSS + EPSS + Asset Criticality Model

```
Priority Score = CVSS_Base × EPSS_Multiplier × Asset_Criticality × Exposure

Where:
CVSS_Base:        0-10 (from CVE database)
EPSS_Multiplier:  1 + (EPSS × 10)  [ranges 1-11]
Asset_Criticality: 1-3 (low/medium/high)
Exposure:         1-3 (internal/limited/public)

Example:
CVE with CVSS 7.5, EPSS 0.15, critical asset, public-facing
Score = 7.5 × 2.5 × 3 × 3 = 168.75 (Very High Priority)
```

### Priority Tiers

```
Tier    Score Range    SLA           Resources
────────────────────────────────────────────────────
P0      > 150          4 hours       All hands, war room
P1      75-150         24 hours      Dedicated engineer
P2      25-75          7 days        Normal sprint work
P3      10-25          30 days       Technical debt backlog
P4      < 10           90 days       As capacity allows
```

## CISA KEV-First Approach

### Why KEV Takes Priority

CISA's Known Exploited Vulnerabilities catalog contains vulnerabilities actively exploited in the wild. These are **proven** threats, not theoretical.

```
Decision Tree:
                    ┌──────────────┐
                    │ Is it in KEV?│
                    └──────┬───────┘
                           │
              ┌────────────┴────────────┐
              │                         │
             YES                        NO
              │                         │
              ▼                         ▼
    ┌─────────────────┐       ┌─────────────────┐
    │ P0: Immediate   │       │ Apply standard  │
    │ action required │       │ prioritization  │
    └─────────────────┘       └─────────────────┘
```

### KEV Response SLAs

| KEV Category | Federal Requirement | Recommended |
|-------------|---------------------|-------------|
| New KEV entry | 14 days | 7 days |
| Critical + KEV | 14 days | 24-48 hours |
| KEV + public-facing | 14 days | 24 hours |

## Business Context Prioritization

### Asset Criticality Classification

```
Tier 1 - Business Critical:
• Revenue-generating systems
• Customer data stores
• Authentication services
• Payment processing
• Core APIs
→ Highest priority, zero tolerance for critical vulns

Tier 2 - Business Important:
• Internal tools with sensitive data
• Employee-facing applications
• Development infrastructure
• CI/CD systems
→ High priority, 48-hour SLA for critical

Tier 3 - Supporting:
• Documentation systems
• Internal utilities
• Development environments
• Test systems
→ Normal priority, follow standard SLAs
```

### Data Sensitivity Multiplier

```
Data Type          Multiplier    Rationale
─────────────────────────────────────────────────────
PII/PHI            3.0×          Regulatory exposure
Financial          2.5×          Fraud/compliance risk
Authentication     2.5×          Full compromise vector
Internal only      1.5×          Limited blast radius
Public data        1.0×          Baseline priority
```

## Sprint Planning Integration

### Security Debt Velocity

Track security work as percentage of sprint capacity:

```
Recommended Allocation:
• 20% of sprint capacity for security work
• Split: 15% remediation, 5% prevention

Example (10-person team, 2-week sprint):
• Total capacity: 200 story points
• Security allocation: 40 points
• Remediation: 30 points
• Prevention: 10 points
```

### Batching Strategy

```
Efficient Batching                 Inefficient Approach
─────────────────────────────────────────────────────────
Week 1: All lodash updates         Monday: Update lodash
        across all services        Tuesday: Update axios
        (shared PR template,       Wednesday: Update lodash again
        bulk testing)              (context switching, duplicate work)

Week 2: All axios updates
        across all services
```

### Breaking Ties

When multiple items have similar priority scores:

```
Tiebreaker Hierarchy:
1. Blast radius (more systems affected = higher)
2. Remediation effort (easier fix = higher)
3. Dependency depth (direct dep = higher than transitive)
4. Time since disclosure (older = higher)
5. Public exploit availability (PoC exists = higher)
```

## Capacity Planning

### Estimating Remediation Effort

```
Update Type                          Typical Effort
─────────────────────────────────────────────────────
Patch update (x.x.PATCH)             30 min - 2 hours
Minor update (x.MINOR.x)             2-4 hours
Major update (MAJOR.x.x)             1-3 days
Framework migration                  1-4 weeks
Full replacement                     2-8 weeks
```

### Team Sizing Model

```
Base Formula:
Engineers needed = (Monthly vulns × Avg hours) / (Hours per engineer × Efficiency)

Example:
• Monthly vulnerabilities: 50
• Average remediation: 4 hours
• Hours per engineer: 160/month
• Efficiency factor: 0.7 (meetings, etc.)

Engineers = (50 × 4) / (160 × 0.7) = 1.8 engineers

Recommendation: 2 FTE allocated to security remediation
```

## Backlog Management

### Vulnerability Backlog Health

```
Healthy Backlog Metrics:
• P0/P1 items: 0 (always empty)
• P2 items: < 10
• P3 items: < 30
• P4 items: < 100
• Average age: < 30 days
• Oldest item: < 90 days
```

### Grooming Process

**Weekly:**
- Review new vulnerabilities
- Assign priorities
- Update estimates
- Identify blockers

**Monthly:**
- Age analysis
- Re-prioritization based on new intelligence
- Capacity adjustment
- Exception review

**Quarterly:**
- Backlog cleanup (close won't-fix)
- Process improvement
- Tool evaluation
- Team retrospective

### Exception Management

```
Exception Request Template:
─────────────────────────────
Vulnerability: CVE-XXXX-XXXXX
Current SLA: 7 days
Requested extension: 30 days
Reason: [Business justification]
Compensating controls: [What's in place]
Residual risk: [What remains]
Approver: [Director+ level]
Expiration: [Hard date]
Review date: [When to reassess]
```

## Automation Opportunities

### Prioritize Automation Investment

```
High ROI Automation:
1. Patch updates with no breaking changes
   - Savings: 30+ hours/month
   - Risk: Low

2. Security scanning in CI/CD
   - Savings: Early detection
   - Risk: None

3. Dependabot/Renovate auto-merge for patch
   - Savings: 20+ hours/month
   - Risk: Low (with tests)

Medium ROI Automation:
4. Minor version updates with comprehensive tests
5. SBOM generation
6. License compliance checking

Low ROI Automation:
7. Major version updates (too variable)
8. Framework migrations (context-dependent)
```

### Auto-Merge Criteria

```yaml
# Safe for auto-merge:
auto_merge:
  conditions:
    - version_change: "patch"
    - tests_passing: true
    - no_breaking_changes: true
    - security_scan_clean: true
    - package_trusted: true  # e.g., in allowlist

# Requires review:
manual_review:
  conditions:
    - version_change: "minor" OR "major"
    - new_dependency: true
    - maintainer_changed: true
    - security_advisory: true
```

## Decision Frameworks

### Fix vs Accept vs Mitigate

```
                    ┌─────────────────┐
                    │ Can we fix it?  │
                    └────────┬────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
           Easy           Hard          Impossible
              │              │              │
              ▼              ▼              ▼
           Fix it     ┌───────────┐   ┌───────────┐
                      │Worth the  │   │ Mitigate  │
                      │effort?    │   │ or Accept │
                      └─────┬─────┘   └───────────┘
                            │
                  ┌─────────┴─────────┐
                 Yes                  No
                  │                   │
                  ▼                   ▼
               Fix it            Mitigate
```

### Replace vs Patch Decision

```
Consider Replacement When:
• Package abandoned (no updates in 2+ years)
• Recurring vulnerabilities (3+ in 12 months)
• Better-maintained alternative exists
• License concerns
• Performance issues compound security work

Stick with Patching When:
• Active maintainer with good response time
• Deep integration (high migration cost)
• No viable alternatives
• Temporary issue (maintainer vacation, etc.)
```

## Metrics for Leaders

### Leading Indicators

```
Watch These Weekly:
• New vulns discovered (trend)
• Backlog growth rate
• Sprint velocity on security items
• Blocked items count
• Auto-fix success rate
```

### Lagging Indicators

```
Monthly/Quarterly:
• MTTR by severity
• SLA compliance rate
• Security debt age
• Incidents from dependencies
• Audit findings
```

### Benchmarking

```
Industry Comparisons (Source: Various):
                        Average    Top 10%    Your Target
MTTR Critical           7 days     1 day      _________
MTTR High               30 days    7 days     _________
Automation rate         30%        60%        _________
SLA compliance          82%        95%        _________
Vulns per 1K deps       15         5          _________
```

## Quick Reference

### Priority Decision Checklist

- [ ] Check KEV catalog first
- [ ] Apply CVSS + EPSS scoring
- [ ] Consider asset criticality
- [ ] Factor in exposure level
- [ ] Evaluate remediation effort
- [ ] Check for available automation
- [ ] Assign appropriate SLA
- [ ] Document exceptions if needed
