# Engineering Cost Optimization Guide

## Cost Categories

### Cloud Infrastructure

| Category | Typical % | Optimization Potential |
|----------|-----------|----------------------|
| Compute (EC2, VMs) | 40-60% | High |
| Storage (S3, EBS) | 15-25% | Medium |
| Data transfer | 10-20% | Medium |
| Databases (RDS, etc) | 10-20% | High |
| Other services | 10-15% | Variable |

### Developer Tooling

| Category | Typical Cost | Optimization |
|----------|--------------|--------------|
| CI/CD minutes | $1-5K/month | Caching, parallelization |
| SaaS tools | $50-200/seat/month | Usage audit, consolidation |
| Code hosting | $4-21/seat/month | Team size optimization |
| Monitoring | $15-50/host/month | Host consolidation |

---

## Cloud Cost Optimization

### 1. Right-Sizing

Most organizations overprovision by 30-50%.

**Analysis Steps:**
1. Identify underutilized resources (< 20% CPU average)
2. Review memory utilization
3. Check network throughput needs
4. Consider burstable instances for variable workloads

**Common Findings:**
| Current | Right-Sized | Monthly Savings |
|---------|-------------|-----------------|
| m5.2xlarge | m5.large | $200 |
| r5.xlarge | m5.xlarge | $150 |
| t3.medium → t3.small | t3.small | $20 |

### 2. Reserved Capacity

Commit for predictable workloads.

| Plan | Discount | Use Case |
|------|----------|----------|
| 1-year no upfront | 30-40% | Baseline capacity |
| 1-year partial upfront | 40-50% | Stable workloads |
| 3-year all upfront | 60-70% | Long-term infrastructure |

**Strategy:**
- Reserved for baseline (always running)
- On-demand for variable load
- Spot for fault-tolerant workloads

### 3. Spot Instances

60-90% cheaper, but can be interrupted.

**Good for:**
- CI/CD runners
- Batch processing
- Dev/test environments
- Stateless web tiers (with fallback)

**Bad for:**
- Databases
- Stateful applications
- Single points of failure

### 4. Storage Optimization

| Strategy | Savings | Implementation |
|----------|---------|----------------|
| S3 lifecycle policies | 40-70% | Move to IA/Glacier |
| GP3 over GP2 (EBS) | 20% | Volume type change |
| Delete unused snapshots | 100% | Snapshot audit |
| Clean up orphaned volumes | 100% | Volume audit |

### 5. Database Costs

| Strategy | Savings | Consideration |
|----------|---------|---------------|
| Reserved RDS | 30-50% | 1-3 year commitment |
| Aurora Serverless v2 | Variable | Low-traffic databases |
| Read replicas (smaller) | 20-40% | Scale reads, not primary |
| Graviton instances | 20% | ARM-based, good performance |

---

## Developer Tooling Costs

### CI/CD Optimization

**GitHub Actions:**
| Optimization | Impact |
|--------------|--------|
| Caching dependencies | -50% build time |
| Skip redundant runs | -30% total runs |
| Use Ubuntu (not macOS) | 10x cheaper |
| Self-hosted runners | $0 compute |

**Calculator:**
```
Monthly cost = minutes × runners × $0.008 (Linux)
                                  × $0.08 (macOS)
```

### SaaS License Audit

1. **Identify unused licenses**
   - Check last login dates
   - Review activity metrics
   - Survey teams

2. **Consolidate tools**
   - Multiple monitoring tools? → Pick one
   - Overlapping features? → Negotiate

3. **Negotiate enterprise deals**
   - Volume discounts
   - Multi-year commitments
   - Feature bundles

### Per-Engineer Costs

**Benchmark:**
| Category | Low | Average | High |
|----------|-----|---------|------|
| Tooling | $200/mo | $400/mo | $800/mo |
| Cloud (attributed) | $100/mo | $300/mo | $1000/mo |
| Total cost | $300/mo | $700/mo | $1800/mo |

---

## Cost Visibility

### Tagging Strategy

**Required Tags:**
```
Environment: production | staging | development
Team: platform | product | data
Service: api | web | worker
Cost-Center: eng-platform | eng-product
```

### Allocation Methods

| Method | Pros | Cons |
|--------|------|------|
| Direct attribution | Accurate | Complex to implement |
| Team-based allocation | Simple | Can be unfair |
| Usage-based chargeback | Fair | Requires good metrics |

### Reporting

**Monthly Cost Report:**
1. Total spend vs budget
2. Cost by team/service
3. Month-over-month change
4. Top cost drivers
5. Optimization opportunities

---

## Optimization Playbook

### Quick Wins (< 1 week)

| Action | Expected Savings |
|--------|------------------|
| Delete unused resources | 5-10% |
| Enable S3 lifecycle policies | 3-5% |
| Resize obvious oversized instances | 5-10% |
| Stop dev environments after hours | 3-5% |

### Medium-Term (1-3 months)

| Action | Expected Savings |
|--------|------------------|
| Implement spot for CI/CD | 60-80% of CI costs |
| Right-size databases | 20-30% of DB costs |
| Purchase reserved capacity | 30-40% of compute |
| Consolidate SaaS tools | 20-30% of tools |

### Strategic (3-12 months)

| Action | Expected Savings |
|--------|------------------|
| Kubernetes (better utilization) | 20-40% |
| Serverless migration | Variable |
| Multi-cloud optimization | 10-20% |
| Build vs buy evaluation | Variable |

---

## Governance

### Budget Management

- Set budgets per team/project
- Alert at 80% threshold
- Review overages weekly
- Require approval for new significant spend

### Cost Review Process

**Weekly:** Quick check on anomalies
**Monthly:** Detailed review with teams
**Quarterly:** Strategic optimization planning

### Accountability

- Teams own their costs
- Cost visible in dashboards
- Include in sprint planning
- Celebrate cost savings

---

## ROI Calculations

### Engineer Time Savings

```
Value = hours_saved × hourly_cost × engineers_affected

Example:
10 min saved/day × $100/hr × 50 engineers
= 0.167 hr × $100 × 50 × 250 days
= $208,750/year
```

### Infrastructure Investment

```
ROI = (savings - investment) / investment × 100

Example:
$50K investment in automation
$200K annual savings
ROI = ($200K - $50K) / $50K = 300%
Payback period = $50K / ($200K/12) = 3 months
```
