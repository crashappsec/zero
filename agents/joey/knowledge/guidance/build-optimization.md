# Build Optimization Guide

## Quick Wins

### 1. Enable Dependency Caching

Every CI run should cache dependencies. This alone can save 1-5 minutes.

**GitHub Actions:**
```yaml
- uses: actions/setup-node@v4
  with:
    node-version: '20'
    cache: 'npm'  # Automatic caching!

# Or manual caching for more control
- uses: actions/cache@v4
  with:
    path: ~/.npm
    key: ${{ runner.os }}-node-${{ hashFiles('**/package-lock.json') }}
    restore-keys: |
      ${{ runner.os }}-node-
```

### 2. Shallow Clone

Don't fetch full git history unless needed.

```yaml
- uses: actions/checkout@v4
  with:
    fetch-depth: 1  # Only latest commit
```

### 3. Cancel Redundant Builds

Don't waste minutes on outdated commits.

```yaml
concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true
```

### 4. Set Timeouts

Prevent stuck builds from running for hours.

```yaml
jobs:
  test:
    timeout-minutes: 15
```

## Parallelization

### Parallel Jobs

Run independent jobs simultaneously:

```yaml
jobs:
  lint:
    runs-on: ubuntu-latest
    steps: [...]

  test:
    runs-on: ubuntu-latest  # Runs parallel to lint
    steps: [...]

  build:
    needs: [lint, test]  # Waits for both
    steps: [...]
```

### Matrix Builds

Test across configurations in parallel:

```yaml
jobs:
  test:
    strategy:
      matrix:
        node: [18, 20]
        os: [ubuntu-latest, windows-latest]
      fail-fast: false  # Don't cancel others on failure
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/setup-node@v4
        with:
          node-version: ${{ matrix.node }}
```

### Test Sharding

Split tests across parallel runners:

```yaml
jobs:
  test:
    strategy:
      matrix:
        shard: [1, 2, 3, 4]
    steps:
      - run: npm test -- --shard=${{ matrix.shard }}/4
```

## Caching Strategies

### What to Cache

| Item | Cache Key | Impact |
|------|-----------|--------|
| npm/yarn/pnpm | lock file hash | 1-3 min |
| Docker layers | Dockerfile hash | 2-10 min |
| Build outputs | source hash | 1-5 min |
| Compiled assets | config hash | 30s-2 min |

### Cache Key Patterns

```yaml
# Exact match preferred, fallback to partial
key: ${{ runner.os }}-npm-${{ hashFiles('**/package-lock.json') }}
restore-keys: |
  ${{ runner.os }}-npm-  # Partial match fallback
```

### Cache vs Artifacts

| Feature | Cache | Artifacts |
|---------|-------|-----------|
| Lifetime | 7 days unused | Configurable |
| Size limit | 10GB per repo | Per-artifact |
| Speed | Fast restore | Slower upload/download |
| Use case | Dependencies | Build outputs between jobs |

## Monorepo Optimization

### Affected-Only Builds

Only build/test what changed:

```yaml
# Nx
- run: npx nx affected --target=test --base=origin/main

# Turbo
- run: npx turbo run test --filter='...[origin/main]'

# Lerna
- run: npx lerna run test --since origin/main
```

### Package-Level Caching

```yaml
# Turborepo remote caching
- run: npx turbo run build --cache-dir=.turbo

# Nx remote caching
- run: npx nx affected --target=build
  env:
    NX_CLOUD_ACCESS_TOKEN: ${{ secrets.NX_TOKEN }}
```

## Docker Build Optimization

### Layer Caching

```dockerfile
# Order: least changing → most changing
FROM node:20-alpine

# Dependencies first (changes rarely)
COPY package*.json ./
RUN npm ci

# Source last (changes often)
COPY . .
RUN npm run build
```

### BuildKit Caching

```yaml
- uses: docker/build-push-action@v5
  with:
    cache-from: type=gha
    cache-to: type=gha,mode=max
```

### Multi-Stage Builds

```dockerfile
# Build stage
FROM node:20 AS builder
WORKDIR /app
COPY . .
RUN npm ci && npm run build

# Production stage (smaller image)
FROM node:20-alpine
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/node_modules ./node_modules
CMD ["node", "dist/index.js"]
```

## Flaky Test Management

### Detection

```yaml
- run: npm test -- --bail  # Stop on first failure
- run: npm test -- --retry=3  # Retry failed tests
```

### Quarantine

```javascript
// Jest
test.skip.failing('flaky test', () => {...})

// Or use test tags
test('stable test', { tags: ['stable'] }, () => {...})
```

### Reporting

Track flaky test rate over time. Target: < 1% flaky rate.

## Metrics to Track

| Metric | Target | How to Measure |
|--------|--------|----------------|
| Build duration | < 10 min | CI analytics |
| Cache hit rate | > 90% | Cache action output |
| Flaky test rate | < 1% | Test retries needed |
| Queue time | < 2 min | Time to runner |
| Success rate | > 95% | CI dashboard |

## Cost Optimization

### Runner Selection

- Use `ubuntu-latest` when possible (1x cost)
- Avoid `macos-latest` unless needed (10x cost)
- Consider self-hosted for heavy workloads

### Storage Management

```yaml
# Clean up old artifacts
- uses: actions/github-script@v6
  with:
    script: |
      const artifacts = await github.rest.actions.listArtifactsForRepo({...});
      // Delete artifacts older than 7 days
```

### Workflow Efficiency

- Skip CI for docs-only changes
- Use path filters to skip irrelevant workflows
- Consolidate similar workflows

```yaml
on:
  push:
    paths-ignore:
      - '**.md'
      - 'docs/**'
```

## Debugging Slow Builds

1. **Enable timing output**
   ```yaml
   - run: npm ci
     env:
       CI: true
       DEBUG: '*'
   ```

2. **Add step timing**
   ```yaml
   - name: Install
     run: time npm ci
   ```

3. **Profile builds**
   ```bash
   npm run build -- --profile
   ```

4. **Analyze cache effectiveness**
   - Check "Post Cache" step for cache size
   - Verify restore vs save times
