# Engineering Leader Persona

## Role Description

An engineering leader (VP, Director, Engineering Manager) responsible for portfolio health, resource allocation, and strategic technical decisions. This persona needs executive summaries, metrics dashboards, and strategic recommendations.

**What they care about:**
- Overall risk posture
- Resource requirements
- Compliance impact
- Trends and benchmarks

## Output Style

- **Tone:** Strategic, business-focused, metric-driven
- **Detail Level:** Low - summaries and trends, not individual findings
- **Format:** Dashboards, tables, executive summaries
- **Prioritization:** By business impact and resource requirements

## Output Template

```markdown
## Executive Summary

**Report Date:** YYYY-MM-DD
**Overall Risk Level:** Critical/High/Medium/Low
**Trend:** Improving/Stable/Declining

[2-3 sentence summary of current state and key concerns]

### Key Metrics

| Metric | Current | Target | Trend | Status |
|--------|---------|--------|-------|--------|
| Critical Issues | X | 0 | ↑/↓/→ | Red/Yellow/Green |
| High Issues | X | <5 | ↑/↓/→ | Red/Yellow/Green |
| MTTR (Critical) | X days | 1 day | ↑/↓/→ | Red/Yellow/Green |
| Coverage | X% | >80% | ↑/↓/→ | Red/Yellow/Green |

### Risk Distribution

| Area | Critical | High | Medium | Risk Level |
|------|----------|------|--------|------------|
| [Area 1] | X | X | X | HIGH |
| [Area 2] | X | X | X | MEDIUM |
| [Area 3] | X | X | X | LOW |

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
| [Framework 1] | Compliant/At Risk | [Description] |
| [Framework 2] | Compliant/At Risk | [Description] |

### Next Review

**Recommended:** [Date or frequency]
```

## Prioritization Framework

1. **Business Critical + Compliance Risk** - Executive escalation
2. **Customer-Facing + Security Risk** - Prioritize resources
3. **Internal Tools + High Severity** - Plan remediation
4. **Technical Debt + Low Risk** - Batch in maintenance

## Key Questions to Answer

- What's our overall security/health posture?
- Where should we focus engineering resources?
- Are we meeting our SLAs?
- What's the compliance impact of our current state?
- How do we compare to industry benchmarks?
- What's the resource investment needed to improve?
