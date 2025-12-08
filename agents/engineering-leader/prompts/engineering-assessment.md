# Engineering Assessment Prompt

## Context
You are assessing engineering operations from a leadership perspective, focusing on delivery performance, team health, and cost efficiency.

## Assessment Areas

### Delivery Performance (DORA)
- Deployment frequency
- Lead time for changes
- Mean time to recovery
- Change failure rate

### Developer Experience
- Onboarding time
- Build/test cycle time
- Toil and manual work
- Tooling satisfaction

### Cost Efficiency
- Cloud spend per engineer
- Tooling costs
- Build minutes usage
- Resource utilization

### Team Health
- Retention indicators
- On-call burden
- Meeting load
- Focus time availability

## Output Format

```markdown
## Engineering Assessment: [Team/Organization Name]

### Executive Summary
2-3 sentence overview of engineering health with key callouts.

### DORA Metrics Assessment

| Metric | Current | Tier | Target |
|--------|---------|------|--------|
| Deployment Frequency | | Elite/High/Medium/Low | |
| Lead Time | | Elite/High/Medium/Low | |
| MTTR | | Elite/High/Medium/Low | |
| Change Failure Rate | | Elite/High/Medium/Low | |

**Overall DORA Tier:** [Elite/High/Medium/Low]

### Developer Experience

| Area | Score (1-5) | Notes |
|------|-------------|-------|
| Onboarding | | |
| Local development | | |
| CI/CD experience | | |
| Documentation | | |
| Tooling | | |

**Key Pain Points:**
1. ...
2. ...

### Cost Analysis

| Category | Monthly Cost | Trend | Benchmark |
|----------|--------------|-------|-----------|
| Cloud infrastructure | $ | ↑↓→ | vs industry |
| CI/CD | $ | ↑↓→ | |
| SaaS tooling | $ | ↑↓→ | |
| Per-engineer total | $ | ↑↓→ | |

**Cost Optimization Opportunities:**
1. ... (estimated savings: $X/month)
2. ...

### Recommendations

| Priority | Initiative | Impact | Effort | ROI |
|----------|------------|--------|--------|-----|
| 1 | | High/Medium/Low | S/M/L | |
| 2 | | High/Medium/Low | S/M/L | |
| 3 | | High/Medium/Low | S/M/L | |

### Investment Asks

If improvements require investment:

| Initiative | Investment | Expected Return | Payback |
|------------|------------|-----------------|---------|
| | $ | $ or time saved | X months |

### Dashboard Metrics to Track

1. **Metric**: Target, Current, Trend
2. ...
```

## Example Output

```markdown
## Engineering Assessment: Platform Team

### Executive Summary
The Platform team is performing at a **Medium DORA tier** with strong deployment frequency but concerning lead time (5 days average). Developer experience scores high but CI/CD frustration is a major pain point. Cloud costs are 20% above benchmark, primarily from oversized RDS instances.

### DORA Metrics Assessment

| Metric | Current | Tier | Target |
|--------|---------|------|--------|
| Deployment Frequency | 2x/day | High | Maintain |
| Lead Time | 5 days | Medium | < 1 day |
| MTTR | 4 hours | High | Maintain |
| Change Failure Rate | 18% | High | < 15% |

**Overall DORA Tier:** Medium (Lead Time dragging down overall)

### Developer Experience

| Area | Score (1-5) | Notes |
|------|-------------|-------|
| Onboarding | 4 | Good docs, could automate more |
| Local development | 3 | Docker setup is painful |
| CI/CD experience | 2 | 25+ minute builds, flaky tests |
| Documentation | 4 | Well maintained |
| Tooling | 3 | Too many tools, overlapping |

**Key Pain Points:**
1. **CI/CD speed**: 25-minute average builds blocking iteration
2. **Flaky tests**: 15% of test runs fail then pass on retry
3. **Local setup**: 2+ hours for new engineers to run locally

### Cost Analysis

| Category | Monthly Cost | Trend | Benchmark |
|----------|--------------|-------|-----------|
| Cloud infrastructure | $45,000 | ↑ +15% | +20% vs benchmark |
| CI/CD | $3,200 | → | Normal |
| SaaS tooling | $8,500 | ↑ +8% | High |
| Per-engineer total | $750 | → | Average |

**Cost Optimization Opportunities:**
1. Right-size RDS instances (savings: $8K/month)
2. Enable spot for CI runners (savings: $2K/month)
3. Consolidate monitoring tools (savings: $3K/month)

### Recommendations

| Priority | Initiative | Impact | Effort | ROI |
|----------|------------|--------|--------|-----|
| 1 | Speed up CI (caching, parallelization) | High | M | High |
| 2 | Fix flaky tests | High | M | High |
| 3 | Right-size RDS | Medium | S | Very High |
| 4 | Consolidate monitoring | Low | L | Medium |

### Investment Asks

| Initiative | Investment | Expected Return | Payback |
|------------|------------|-----------------|---------|
| CI optimization | $15K (contractor) | 2 hrs/dev/week saved | 2 months |
| Local dev improvement | $10K | 1 day onboarding saved | 3 months |

### Dashboard Metrics to Track

1. **CI Build Time**: Target < 10 min, Current 25 min, ↑ trending worse
2. **Flaky Test Rate**: Target < 5%, Current 15%, → stable
3. **Cloud Cost/Engineer**: Target $600, Current $750, ↑ trending worse
4. **PR Lead Time**: Target < 24 hrs, Current 5 days, → stable
```
