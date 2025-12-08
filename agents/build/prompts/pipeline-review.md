# CI/CD Pipeline Review Prompt

## Context
You are reviewing CI/CD pipeline configuration for speed, reliability, and cost efficiency.

## Review Checklist

### Speed
- [ ] Dependencies cached properly
- [ ] Shallow clone where appropriate
- [ ] Parallel jobs for independent tasks
- [ ] Test splitting/sharding for large suites
- [ ] Affected-only builds in monorepos
- [ ] Docker layer caching

### Reliability
- [ ] Timeouts set on all jobs
- [ ] Retry logic for flaky operations
- [ ] Concurrency control configured
- [ ] Fail-fast disabled for matrix (if desired)
- [ ] Error handling for failed steps

### Security
- [ ] Secrets not exposed in logs
- [ ] Minimal permissions (least privilege)
- [ ] Dependencies pinned to versions
- [ ] Approved actions only

### Cost
- [ ] Appropriate runner types
- [ ] Redundant runs cancelled
- [ ] Artifacts cleaned up
- [ ] Path filters to skip unnecessary runs

## Output Format

```markdown
## Pipeline Review: [Workflow Name]

### Performance Summary

| Metric | Current | Target | Gap |
|--------|---------|--------|-----|
| Total Duration | X min | < Y min | Z min |
| Cache Hit Rate | X% | > 90% | |
| Parallelism | X jobs | Y jobs | |

### Quick Wins (< 1 hour to implement)

1. **[Optimization]**
   - Current: What's happening
   - Change: Specific fix
   - Impact: Estimated time savings

### Medium-Term Improvements

1. **[Optimization]**
   - Effort: X hours
   - Impact: Y minutes saved per run

### Bottleneck Analysis

```
[Timeline visualization]
Job A: ████████ (4 min)
Job B:         ████████████████ (8 min) ← BOTTLENECK
Job C:                         ████ (2 min)
Total: 14 min (could be 8 min with parallelization)
```

### Security Concerns

1. **[Issue]**
   - Risk: Description
   - Fix: Recommendation

### Recommendations

| Priority | Action | Effort | Impact |
|----------|--------|--------|--------|
| 1 | ... | Low | High |
| 2 | ... | Medium | Medium |
```

## Example Output

```markdown
## Pipeline Review: main.yml

### Performance Summary

| Metric | Current | Target | Gap |
|--------|---------|--------|-----|
| Total Duration | 12 min | < 8 min | 4 min |
| Cache Hit Rate | 45% | > 90% | 45% |
| Parallelism | 2 jobs | 4 jobs | 2 |

### Quick Wins (< 1 hour to implement)

1. **Add dependency caching**
   - Current: `npm ci` runs fresh every time (2.5 min)
   - Change: Add `cache: 'npm'` to setup-node action
   - Impact: ~2 minutes saved

2. **Shallow clone**
   - Current: Fetching full history (45 seconds)
   - Change: Add `fetch-depth: 1` to checkout
   - Impact: ~40 seconds saved

3. **Add concurrency control**
   - Current: Multiple runs for same branch
   - Change: Add concurrency group with cancel-in-progress
   - Impact: Reduces wasted minutes

### Medium-Term Improvements

1. **Split lint and test jobs**
   - Effort: 1 hour
   - Impact: Run in parallel, save 3 minutes

2. **Add test sharding**
   - Effort: 2 hours
   - Impact: 4 shards could reduce 6 min test to 2 min

### Bottleneck Analysis

```
Checkout:  ██ (0.8 min)
Install:   ████████████ (2.5 min) ← No caching!
Lint:      ████ (1 min)
Test:      ████████████████████████ (6 min) ← Could parallelize
Build:     ████████ (2 min)
─────────────────────────────────────
Total: 12.3 min (sequential)
Could be: ~6 min (with caching + parallel)
```

### Security Concerns

1. **Overly permissive token**
   - Risk: `contents: write` not needed for this workflow
   - Fix: Remove or scope to specific job

2. **Unpinned action version**
   - Risk: `uses: actions/checkout@main` could change
   - Fix: Pin to specific version `@v4.1.1`

### Recommendations

| Priority | Action | Effort | Impact |
|----------|--------|--------|--------|
| 1 | Add npm caching | 5 min | 2 min/run |
| 2 | Shallow clone | 2 min | 40s/run |
| 3 | Parallel lint/test | 30 min | 3 min/run |
| 4 | Test sharding | 2 hours | 4 min/run |

**Total potential savings: ~10 minutes per run (83% faster)**
```
