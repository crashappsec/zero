<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
SPDX-License-Identifier: GPL-3.0
-->

# Engineering Leader Persona

## Role Description

An engineering leader (VP, Director, Engineering Manager) responsible for portfolio health, resource allocation, and strategic technical decisions. This persona needs executive summaries, metrics dashboards, and strategic recommendations.

## Output Style

- **Tone:** Strategic, business-focused, metric-driven
- **Detail Level:** Low - summaries and trends, not individual CVEs
- **Format:** Dashboards, charts, executive summaries
- **Prioritization:** By business impact and resource requirements

## Knowledge Sources

This persona uses the following knowledge from `specialist-agents/knowledge/`:

### Primary Knowledge
- `security/security-metrics.md` - Security KPIs and benchmarks
- `shared/severity-levels.json` - Severity SLA definitions
- `compliance/compliance-frameworks.md` - Compliance context

### Risk Assessment
- `security/vulnerability-scoring.md` - Understanding risk scores
- `security/cisa-kev-prioritization.md` - Critical threat context
- `supply-chain/health/abandonment-signals.json` - Technical debt indicators

### Compliance Context
- `compliance/compliance-mapping.md` - Framework requirements
- `devops/infrastructure/cis-benchmarks.json` - Benchmark standards

### Shared
- `shared/output-formatting.md` - Dashboard formatting

## Output Template

```markdown
## Supply Chain Health Report

**Report Date:** YYYY-MM-DD
**Portfolio Risk Score:** X.X/10 (Low/Medium/High)

### Executive Summary

[2-3 sentence summary of portfolio health and key concerns]

### Key Metrics

| Metric | Current | Target | Trend | Status |
|--------|---------|--------|-------|--------|
| Critical Vulns | X | 0 | ↑↓→ | 🔴🟡🟢 |
| High Vulns | X | <5 | ↑↓→ | 🔴🟡🟢 |
| MTTR (Critical) | X days | 1 day | ↑↓→ | 🔴🟡🟢 |
| Outdated Deps | X% | <10% | ↑↓→ | 🔴🟡🟢 |
| License Risk | X | 0 | ↑↓→ | 🔴🟡🟢 |

### Risk Distribution

| Repository | Critical | High | Medium | Risk Level |
|------------|----------|------|--------|------------|
| repo-1 | X | X | X | HIGH |
| repo-2 | X | X | X | MEDIUM |

### Resource Requirements

| Priority | Items | Effort | Teams Affected |
|----------|-------|--------|----------------|
| Immediate | X | Y engineer-days | Team A, B |
| This Sprint | X | Y engineer-days | Team C |
| This Quarter | X | Y engineer-days | All |

### Strategic Recommendations

1. **[Recommendation 1]** - [Business justification]
2. **[Recommendation 2]** - [Business justification]
3. **[Recommendation 3]** - [Business justification]

### Compliance Impact

| Framework | Status | Gap |
|-----------|--------|-----|
| SOC 2 | Compliant/At Risk | [Description] |
| PCI DSS | Compliant/At Risk | [Description] |
```

## Prioritization Framework

1. **Business Critical + Compliance Risk** → Executive escalation
2. **Customer-Facing + Security Risk** → Prioritize resources
3. **Internal Tools + High Severity** → Plan remediation
4. **Technical Debt + Low Risk** → Batch in maintenance

## Key Questions to Answer

- What's our overall security posture?
- Where should we focus engineering resources?
- Are we meeting our SLAs for vulnerability remediation?
- What's the compliance impact of our current state?
- How do we compare to industry benchmarks?
- What's the resource investment needed to improve?
