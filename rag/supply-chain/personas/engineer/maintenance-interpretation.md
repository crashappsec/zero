<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
SPDX-License-Identifier: GPL-3.0
-->

# Understanding Package Maintenance Signals

## OpenSSF Scorecard Interpretation

### What is OpenSSF Scorecard?

The OpenSSF Scorecard evaluates open source project security practices with automated checks.

**Running Scorecard:**
```bash
# Install
go install github.com/ossf/scorecard/v4/cmd/scorecard@latest

# Run against a repo
scorecard --repo=github.com/owner/repo

# Or use the REST API
curl "https://api.securityscorecards.dev/projects/github.com/owner/repo"
```

### Score Categories (0-10 scale)

| Check | What It Measures | Weight |
|-------|------------------|--------|
| **Code-Review** | PRs reviewed before merge | High |
| **Maintained** | Recent commits, issue responses | High |
| **Branch-Protection** | Protected main branch | High |
| **Vulnerabilities** | Known unfixed vulns | Critical |
| **Dependency-Update-Tool** | Dependabot/Renovate enabled | Medium |
| **Fuzzing** | Fuzz testing implemented | Medium |
| **SAST** | Static analysis in CI | Medium |
| **Token-Permissions** | Minimal GitHub token scope | Medium |
| **Signed-Releases** | Cryptographic signatures | Medium |
| **Binary-Artifacts** | No checked-in binaries | Low |
| **Pinned-Dependencies** | Hash-pinned deps in CI | Low |

### Interpreting Scores

**Overall Score Ranges:**
```
8-10: Excellent - Safe to use, well-maintained
6-7:  Good - Minor improvements needed, generally safe
4-5:  Fair - Some concerns, evaluate necessity
2-3:  Poor - Significant risk, consider alternatives
0-1:  Critical - Avoid if possible
```

**Critical Checks (Must-Haves):**
```
✓ Maintained > 5      (Project is active)
✓ Vulnerabilities = 10 (No known unfixed CVEs)
✓ Code-Review > 5     (Changes are reviewed)
✓ Branch-Protection > 5 (Main branch protected)
```

### Common Score Issues and What They Mean

**Low Maintained Score:**
```
Score: 2/10
Reason: No commits in 90 days, no issue responses

What this means:
- Security issues may not be fixed
- Bugs won't be addressed
- API may become stale

Action: Check if project is archived, find alternatives
```

**Low Code-Review Score:**
```
Score: 3/10
Reason: 40% of commits lack review

What this means:
- Single maintainer pushing directly
- Quality may be inconsistent
- Malicious code harder to catch

Action: Evaluate commit history, test thoroughly
```

**Low Branch-Protection Score:**
```
Score: 1/10
Reason: No branch protection on default branch

What this means:
- Anyone with write access can push to main
- No required reviews
- Force pushes allowed

Action: Fork and maintain if critical dependency
```

## NPM Package Health Signals

### Registry Metadata Analysis

```bash
# Get package info
npm view package-name

# Key fields to check:
npm view package-name time.modified  # Last publish
npm view package-name maintainers    # Who maintains it
npm view package-name repository     # Source code location
```

### Health Indicators

**Positive Signals:**
```
✓ Regular releases (monthly or quarterly)
✓ Multiple maintainers (bus factor > 1)
✓ Linked GitHub repository
✓ TypeScript definitions (@types or built-in)
✓ Good weekly download trends
✓ Recent npm publish (< 6 months)
✓ Semantic versioning followed
✓ CHANGELOG maintained
```

**Warning Signals:**
```
⚠ Single maintainer
⚠ No releases in 12+ months
⚠ No linked repository
⚠ Declining download trends
⚠ Many open issues (100+)
⚠ Security advisories unfixed
⚠ Breaking changes in patch versions
```

**Red Flags:**
```
🚨 Maintainer account compromised (check news)
🚨 Package hijacked/transferred recently
🚨 Tarball differs from repository
🚨 Suspicious postinstall scripts
🚨 Dependency on known malicious packages
```

### NPM Provenance Checking

```bash
# Check if package has provenance
npm view package-name --json | jq '.dist.attestations'

# Provenance tells you:
# - Build system that created the package
# - Source repository commit
# - Build workflow used
```

## GitHub Repository Health

### Automated Checks

```bash
# Community health files
curl https://api.github.com/repos/owner/repo/community/profile

# Returns presence of:
# - README
# - CODE_OF_CONDUCT
# - CONTRIBUTING
# - LICENSE
# - ISSUE_TEMPLATE
# - PULL_REQUEST_TEMPLATE
```

### Manual Assessment Checklist

**Activity Metrics:**
```
□ Commits in last 90 days?
□ Issues responded to within 7 days?
□ PRs reviewed and merged regularly?
□ Releases follow schedule?
□ Security advisories addressed?
```

**Documentation Quality:**
```
□ Clear README with usage examples?
□ API documentation current?
□ Migration guides for major versions?
□ CHANGELOG maintained?
□ Security policy (SECURITY.md)?
```

**Community Health:**
```
□ Multiple active contributors?
□ Healthy issue/PR ratio?
□ Code of conduct present?
□ Contributing guidelines clear?
□ Discussion forum active?
```

## deps.dev Integration

### What deps.dev Provides

[deps.dev](https://deps.dev) aggregates package health data:

```bash
# API endpoint
curl "https://api.deps.dev/v3/systems/npm/packages/express"

# Returns:
# - Version history
# - Dependencies
# - Dependents count
# - Security advisories
# - OpenSSF Scorecard (if available)
# - License information
```

### Key Metrics from deps.dev

**Dependency Graph:**
```
Direct dependencies: 30
Transitive dependencies: 127
Dependents: 1.2M packages

Analysis:
- Low direct deps = less attack surface
- High dependents = well-tested, important to ecosystem
- Many transitive deps = harder to audit
```

**Security Signals:**
```json
{
  "securityAdvisories": 2,
  "openIssues": 45,
  "lastRelease": "2024-01-15",
  "developmentStatus": "active"
}
```

## Evaluating Maintenance for Decisions

### Decision Matrix

```
                    High Usage          Low Usage
                    ──────────          ─────────
Well Maintained  │  ✓ Safe to use    │  ✓ Safe to use
                 │                    │
Poorly Maintained│  ⚠ Monitor closely │  ✗ Find alternative
                 │    or fork         │    or remove
```

### Questions to Ask

1. **Is this package critical to my application?**
   - Yes → Requires high maintenance standards
   - No → Lower standards acceptable

2. **Are there well-maintained alternatives?**
   - Yes → Consider switching if current is poorly maintained
   - No → May need to fork or contribute

3. **What's the blast radius of compromise?**
   - High (auth, crypto, data) → Strict requirements
   - Low (dev tooling) → More tolerance

### Maintenance Thresholds by Risk

**High-Risk Dependencies** (auth, crypto, user data):
```
Required:
- OpenSSF Scorecard > 7
- Last commit < 90 days
- Multiple maintainers
- Quick security response history
- Provenance available
```

**Medium-Risk Dependencies** (core functionality):
```
Required:
- OpenSSF Scorecard > 5
- Last commit < 180 days
- Active issue responses
- Security advisories addressed
```

**Low-Risk Dependencies** (dev tools, utilities):
```
Acceptable:
- OpenSSF Scorecard > 3
- Last commit < 1 year
- No critical security issues
- Stable API
```

## Monitoring Package Health Over Time

### Setting Up Alerts

**Dependabot alerts:**
```yaml
# .github/dependabot.yml
version: 2
updates:
  - package-ecosystem: "npm"
    directory: "/"
    schedule:
      interval: "weekly"
    open-pull-requests-limit: 10
```

**Socket.dev integration:**
- Monitors for supply chain attacks
- Alerts on maintainer changes
- Detects suspicious package updates

### Periodic Review Process

**Monthly:**
- Review Dependabot PRs
- Check for security advisories
- Update critical dependencies

**Quarterly:**
- Run full dependency audit
- Review OpenSSF Scorecards for critical deps
- Evaluate alternatives for poorly maintained deps
- Update dependency policy

**Annually:**
- Full supply chain review
- Vendor assessment for critical dependencies
- Update minimum maintenance requirements

## Quick Reference

### Red Flag Combinations

```
🚨 IMMEDIATE ACTION REQUIRED:
- Known vulnerability + no maintainer response
- Recent maintainer transfer + suspicious changes
- Declining scorecard + critical usage

⚠️ PLAN MIGRATION:
- No commits in 2 years + active alternatives exist
- Multiple unfixed CVEs + slow response time
- Deprecated by maintainer

📋 MONITOR CLOSELY:
- Single maintainer + high criticality
- Scorecard declining trend
- Unusual release pattern
```

### Maintenance Assessment Commands

```bash
# Quick health check script
npm view $PKG version time.modified maintainers

# Check for vulnerabilities
npm audit --json | jq '.vulnerabilities'

# OpenSSF Scorecard
scorecard --repo=github.com/$OWNER/$REPO --checks=Maintained,Vulnerabilities

# deps.dev API
curl "https://api.deps.dev/v3/systems/npm/packages/$PKG"
```
