# Deployment Strategies Guide

## Strategy Overview

| Strategy | Risk | Rollback Speed | Resource Cost | Complexity |
|----------|------|----------------|---------------|------------|
| Rolling | Low | Slow | Low | Low |
| Blue-Green | Very Low | Instant | High (2x) | Medium |
| Canary | Very Low | Fast | Medium | High |
| Feature Flags | Very Low | Instant | Low | Medium |

## Rolling Deployment

Gradually replace old instances with new ones.

```
Before:  [v1] [v1] [v1] [v1]
During:  [v2] [v1] [v1] [v1]
During:  [v2] [v2] [v1] [v1]
During:  [v2] [v2] [v2] [v1]
After:   [v2] [v2] [v2] [v2]
```

**Kubernetes Example:**
```yaml
spec:
  replicas: 4
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
      maxSurge: 1
```

**Pros:**
- Simple to implement
- Low resource overhead
- Built into most platforms

**Cons:**
- Mixed versions during deploy
- Slow rollback (redeploy old version)
- Need backward-compatible changes

**Best for:** Most deployments, stateless services

---

## Blue-Green Deployment

Run two identical environments, switch traffic instantly.

```
Before:    Traffic → [Blue v1]    [Green idle]
Deploy:    Traffic → [Blue v1]    [Green v2] ← deploy here
Switch:    Traffic → [Green v2]   [Blue v1 standby]
```

**Implementation:**
```yaml
# Two deployments
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-blue
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-green
---
# Service points to active
apiVersion: v1
kind: Service
metadata:
  name: app
spec:
  selector:
    deployment: green  # Switch this
```

**Pros:**
- Instant rollback (switch back)
- Test in production before switching
- No mixed versions

**Cons:**
- 2x resource cost
- Database schema challenges
- Complex for stateful apps

**Best for:** Critical applications, compliance requirements

---

## Canary Deployment

Route small percentage of traffic to new version, gradually increase.

```
Start:     [v1 90%] ←── Traffic ──→ [v2 10%]
Progress:  [v1 50%] ←── Traffic ──→ [v2 50%]
Complete:  [v2 100%] ←─ Traffic
```

**Kubernetes with Istio:**
```yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
spec:
  http:
  - route:
    - destination:
        host: app
        subset: v1
      weight: 90
    - destination:
        host: app
        subset: v2
      weight: 10
```

**Progressive Rollout (Argo Rollouts):**
```yaml
spec:
  strategy:
    canary:
      steps:
      - setWeight: 10
      - pause: {duration: 10m}
      - setWeight: 30
      - pause: {duration: 10m}
      - setWeight: 100
```

**Pros:**
- Minimize blast radius
- Data-driven decisions
- Automatic rollback possible

**Cons:**
- Requires traffic management
- Complex monitoring setup
- Multiple versions running

**Best for:** High-traffic apps, A/B testing, risk-averse deploys

---

## Feature Flags

Deploy code but toggle features independently.

```javascript
// Code deployed but inactive
if (featureFlags.isEnabled('new-checkout')) {
  return newCheckout();
}
return oldCheckout();
```

**Providers:**
- LaunchDarkly
- Split.io
- Flagsmith
- Unleash

**Pros:**
- Instant enable/disable
- Target specific users
- Decouple deploy from release

**Cons:**
- Code complexity (flag debt)
- Testing combinatorial explosion
- Requires cleanup discipline

**Best for:** Gradual rollouts, A/B testing, kill switches

---

## Database Migration Strategies

### Expand-Contract Pattern

Safe database schema changes in three phases:

```
1. Expand:   Add new column (nullable), deploy code that writes to both
2. Migrate:  Backfill data, verify consistency
3. Contract: Remove old column, clean up code
```

**Example: Renaming a column**
```sql
-- Phase 1: Expand
ALTER TABLE users ADD COLUMN full_name VARCHAR(255);
-- App writes to both name and full_name

-- Phase 2: Migrate
UPDATE users SET full_name = name WHERE full_name IS NULL;
-- Verify data

-- Phase 3: Contract
ALTER TABLE users DROP COLUMN name;
-- App only uses full_name
```

### Forward-Only Migrations

Never rollback database changes. Instead:
- Test thoroughly before deploy
- Use expand-contract for breaking changes
- Keep rollback scripts for data issues

---

## Monitoring Deployments

### Key Metrics to Watch

| Metric | Threshold | Action |
|--------|-----------|--------|
| Error rate | > 1% increase | Pause/rollback |
| Latency p99 | > 20% increase | Investigate |
| CPU/Memory | > 80% | Scale up first |
| Traffic drop | > 10% | Investigate immediately |

### Deployment Automation

```yaml
# Argo Rollouts with analysis
spec:
  strategy:
    canary:
      analysis:
        templates:
        - templateName: success-rate
        args:
        - name: service-name
          value: my-app
---
apiVersion: argoproj.io/v1alpha1
kind: AnalysisTemplate
metadata:
  name: success-rate
spec:
  metrics:
  - name: success-rate
    interval: 1m
    successCondition: result >= 0.99
    provider:
      prometheus:
        query: |
          sum(rate(http_requests_total{status=~"2.*"}[5m])) /
          sum(rate(http_requests_total[5m]))
```

---

## Rollback Procedures

### Automated Rollback Triggers

- Error rate exceeds threshold
- Latency exceeds threshold
- Health checks failing
- Custom metric conditions

### Manual Rollback Steps

1. **Kubernetes rolling back:**
   ```bash
   kubectl rollout undo deployment/app
   ```

2. **Blue-green switch back:**
   ```bash
   kubectl patch service app -p '{"spec":{"selector":{"deployment":"blue"}}}'
   ```

3. **Canary abort:**
   ```bash
   kubectl argo rollouts abort app
   ```

### Post-Rollback Checklist

- [ ] Verify service restored
- [ ] Check error rates back to normal
- [ ] Notify stakeholders
- [ ] Create incident ticket
- [ ] Root cause analysis
