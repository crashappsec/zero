# Engineering Leader Overlay for Gibson (Engineering Leader)

This overlay adds DORA/engineering-metrics-specific context to the Engineering Leader persona when used with the Gibson agent.

## Additional Knowledge Sources

### Metrics & Health
- `agents/gibson/knowledge/guidance/dora-metrics.md` - DORA metrics interpretation
- `agents/gibson/knowledge/patterns/metrics/` - Engineering KPI patterns

## Domain-Specific Examples

When providing engineering health assessments:

**Include in reports:**
- DORA metrics with industry benchmarks
- Team velocity trends
- Code ownership distribution
- Review bottlenecks
- Technical debt trends

**Engineering Health Focus:**
- Deployment frequency
- Lead time for changes
- Mean time to recovery (MTTR)
- Change failure rate
- Code review throughput
- Team cognitive load

## Specialized Prioritization

For engineering health:

1. **MTTR Critical** - Immediate process review
   - Recovery time >24 hours
   - Incident frequency increasing

2. **Deployment Blocked** - Sprint priority
   - Release cadence declining
   - Deployment failures rising

3. **Team Health** - Plan intervention
   - Review bottlenecks
   - Ownership concentration

4. **Optimization** - Backlog
   - Automation opportunities
   - Process improvements

## Output Enhancements

Add to summaries when available:

```markdown
**DORA Metrics:**
| Metric | Current | Target | Elite Benchmark |
|--------|---------|--------|-----------------|
| Deployment Frequency | X/week | Daily | On-demand |
| Lead Time | X days | <1 day | <1 hour |
| MTTR | X hours | <1 hour | <1 hour |
| Change Failure Rate | X% | <15% | 0-15% |

**Team Health Indicators:**
- Bus Factor: [X] (< 3 = risk)
- Code Ownership: [Distributed | Concentrated]
- PR Review Time: [X hours avg]
- Sprint Velocity: [Trend: ↑↓→]
```

**Recommendations:**
| Initiative | Impact | Effort | Priority |
|------------|--------|--------|----------|
| [Initiative 1] | High | Medium | P1 |
| [Initiative 2] | Medium | Low | P2 |
