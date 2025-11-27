# DORA Metrics Guide

## Overview

DORA (DevOps Research and Assessment) metrics are four key metrics that indicate software delivery performance and organizational performance.

## The Four Metrics

### 1. Deployment Frequency

**What it measures:** How often code deploys to production

**Why it matters:** Higher frequency indicates:
- Smaller batch sizes
- Lower risk per deployment
- Faster value delivery
- Better CI/CD maturity

**How to improve:**
1. Automate deployments
2. Adopt trunk-based development
3. Reduce batch size
4. Use feature flags to decouple deploy from release

**Benchmarks:**
| Tier | Frequency |
|------|-----------|
| Elite | Multiple per day |
| High | Daily to weekly |
| Medium | Weekly to monthly |
| Low | Monthly to quarterly |

---

### 2. Lead Time for Changes

**What it measures:** Time from code commit to production

**Why it matters:** Shorter lead time means:
- Faster feedback loops
- Quicker time to market
- Better developer experience
- More competitive advantage

**How to improve:**
1. Speed up CI/CD pipelines (target < 10 min)
2. Automate testing
3. Reduce PR review time (< 24 hours)
4. Keep PRs small (< 200 lines)

**Benchmarks:**
| Tier | Lead Time |
|------|-----------|
| Elite | < 1 hour |
| High | < 1 week |
| Medium | < 1 month |
| Low | > 1 month |

---

### 3. Mean Time to Recovery (MTTR)

**What it measures:** Time to restore service after incident

**Why it matters:** Lower MTTR indicates:
- Better incident response
- Stronger observability
- Resilient architecture
- Less customer impact

**How to improve:**
1. Implement feature flags for instant rollback
2. Build strong observability (metrics, logs, traces)
3. Create runbooks for common issues
4. Practice incident response
5. Automate rollback procedures

**Benchmarks:**
| Tier | MTTR |
|------|------|
| Elite | < 1 hour |
| High | < 1 day |
| Medium | < 1 week |
| Low | > 1 week |

---

### 4. Change Failure Rate

**What it measures:** Percentage of deployments causing failures

**Why it matters:** Lower failure rate indicates:
- Better quality gates
- Effective testing
- Lower risk deployments
- Higher customer trust

**How to improve:**
1. Increase test coverage
2. Add pre-production environments
3. Implement canary deployments
4. Use feature flags
5. Conduct chaos engineering

**Benchmarks:**
| Tier | Failure Rate |
|------|--------------|
| Elite | 0-15% |
| High | 16-30% |
| Medium | 31-45% |
| Low | > 45% |

---

## Measuring DORA Metrics

### Data Sources

| Metric | Sources |
|--------|---------|
| Deployment Frequency | CI/CD logs, Git tags, K8s events |
| Lead Time | Git commits, PR timestamps, deploy logs |
| MTTR | Incident tools, alert history, status pages |
| Change Failure Rate | Rollbacks, hotfixes, incident correlation |

### Tools

- **DORA Quick Check**: Google's self-assessment
- **LinearB, Jellyfish**: Engineering analytics platforms
- **Sleuth, Faros AI**: DORA-focused tools
- **DIY**: Custom dashboards from CI/CD data

### Calculation Examples

**Deployment Frequency:**
```
deployments_per_week = COUNT(deployments) / weeks_in_period
```

**Lead Time:**
```
lead_time = deploy_timestamp - commit_timestamp
median_lead_time = MEDIAN(all lead_times)
```

**MTTR:**
```
mttr = incident_resolved_time - incident_detected_time
mean_mttr = AVG(all incident recovery times)
```

**Change Failure Rate:**
```
cfr = (deployments_causing_incidents / total_deployments) * 100
```

---

## Improvement Roadmap

### From Low to Medium

1. **Automate basic CI/CD**
   - Set up automated testing
   - Implement basic deployment automation
   - Add monitoring basics

2. **Reduce batch size**
   - Encourage smaller PRs
   - More frequent merges

3. **Improve observability**
   - Add logging
   - Basic alerting
   - Error tracking

### From Medium to High

1. **Speed up pipelines**
   - Parallel testing
   - Caching optimizations
   - Test splitting

2. **Improve deployment process**
   - Blue-green or canary deployments
   - Feature flags
   - Automated rollback

3. **Strengthen incident response**
   - On-call rotations
   - Runbooks
   - Blameless postmortems

### From High to Elite

1. **Full automation**
   - Trunk-based development
   - Continuous deployment
   - Progressive delivery

2. **Culture shift**
   - Everyone deploys
   - Psychological safety
   - Learning from failures

3. **Advanced practices**
   - Chaos engineering
   - SLOs and error budgets
   - Platform engineering

---

## Common Pitfalls

### Gaming Metrics

**Problem:** Teams optimize for metrics, not outcomes

**Examples:**
- Many tiny deployments that don't deliver value
- Not counting failures as failures
- Cherry-picking measurement periods

**Solution:** Focus on trends, not absolute numbers. Combine with outcome metrics (customer satisfaction, revenue).

### Vanity Metrics

**Problem:** Measuring what's easy, not what matters

**Examples:**
- Lines of code
- Number of commits
- Build minutes

**Solution:** Stick to DORA metrics and business outcomes.

### Measurement Burden

**Problem:** Spending more time measuring than improving

**Solution:** Automate data collection. Start simple and iterate.

---

## Connecting to Business Outcomes

DORA metrics are **leading indicators** that predict:

| DORA Metric | Business Outcome |
|-------------|------------------|
| Deployment Frequency | Time to market, competitive advantage |
| Lead Time | Developer productivity, innovation speed |
| MTTR | Customer trust, revenue protection |
| Change Failure Rate | Quality, customer satisfaction |

Research shows elite performers are:
- 208x more likely to deploy on demand
- 106x faster from commit to deploy
- 7x lower change failure rate
- 2,604x faster recovery from incidents
