# GitHub Actions CI/CD Plan

> Plan for automated quality, security, and performance checks using Claude as code review agent.

## Overview

Implement GitHub Actions workflows that:
1. Run on every PR and push to main
2. Enforce code quality standards
3. Use Claude API for intelligent code review
4. Block merges on critical issues

---

## Workflow Structure

```
.github/workflows/
├── ci.yml              # Main CI pipeline (build, test, lint)
├── code-review.yml     # Claude-powered code review
├── security.yml        # Security scanning with Zero
└── release.yml         # Release automation
```

---

## Workflow 1: CI Pipeline (`ci.yml`)

**Triggers:** Push to main, PR to main

### Jobs

#### 1. Build & Test (Go)
```yaml
build-go:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with:
        go-version: '1.22'
    - run: go build -v ./...
    - run: go test -v -race -coverprofile=coverage.out ./...
    - uses: codecov/codecov-action@v4
      with:
        files: coverage.out
```

#### 2. Build & Test (Web)
```yaml
build-web:
  runs-on: ubuntu-latest
  defaults:
    run:
      working-directory: web
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-node@v4
      with:
        node-version: '20'
        cache: 'npm'
        cache-dependency-path: web/package-lock.json
    - run: npm ci
    - run: npm run lint
    - run: npm run type-check
    - run: npm run build
```

#### 3. Lint (Go)
```yaml
lint-go:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: golangci/golangci-lint-action@v4
      with:
        version: latest
        args: --timeout=5m
```

---

## Workflow 2: Claude Code Review (`code-review.yml`)

**Triggers:** PR opened, synchronized

### Implementation Options

#### Option A: Claude GitHub Action (Recommended)
Use official Anthropic GitHub Action when available, or build custom action.

```yaml
name: Claude Code Review

on:
  pull_request:
    types: [opened, synchronize]

jobs:
  review:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      pull-requests: write
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Get changed files
        id: changed
        run: |
          echo "files=$(git diff --name-only origin/main...HEAD | tr '\n' ' ')" >> $GITHUB_OUTPUT

      - name: Claude Review
        uses: anthropics/claude-code-review-action@v1  # hypothetical
        with:
          anthropic-api-key: ${{ secrets.ANTHROPIC_API_KEY }}
          files: ${{ steps.changed.outputs.files }}
          review-type: 'security,quality,performance'
          fail-on: 'critical'
```

#### Option B: Custom Script with Claude API

```yaml
- name: Run Claude Review
  env:
    ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
  run: |
    # Get diff
    git diff origin/main...HEAD > diff.patch

    # Call Claude API for review
    python scripts/claude_review.py diff.patch --output review.json

    # Post review comments
    python scripts/post_review.py review.json

- name: Check Review Results
  run: |
    if grep -q '"severity": "critical"' review.json; then
      echo "Critical issues found"
      exit 1
    fi
```

### Review Prompt Template

```markdown
You are a code reviewer for Zero, a security analysis toolkit.

Review the following code changes for:
1. **Security issues** - vulnerabilities, exposed secrets, unsafe patterns
2. **Quality issues** - error handling, resource leaks, code smells
3. **Performance issues** - inefficient algorithms, memory leaks, N+1 queries
4. **Best practices** - Go idioms, React patterns, TypeScript safety

For each issue found, provide:
- File and line number
- Severity (critical/high/medium/low)
- Description of the issue
- Suggested fix

Code diff:
```diff
{diff_content}
```

Respond in JSON format:
{
  "summary": "Brief summary of review",
  "issues": [
    {
      "file": "path/to/file.go",
      "line": 42,
      "severity": "high",
      "category": "security",
      "description": "Description of issue",
      "suggestion": "How to fix"
    }
  ],
  "approved": true/false
}
```

---

## Workflow 3: Security Scanning (`security.yml`)

**Triggers:** Push to main, PR to main, weekly schedule

### Jobs

#### 1. Zero Self-Scan
```yaml
zero-scan:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with:
        go-version: '1.22'
    - run: go build -o zero ./cmd/zero
    - run: ./zero scan . --profile security --output sarif
    - uses: github/codeql-action/upload-sarif@v3
      with:
        sarif_file: zero-results.sarif
```

#### 2. Dependency Scanning
```yaml
deps-scan:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - name: Go vulnerabilities
      run: govulncheck ./...
    - name: npm audit
      working-directory: web
      run: npm audit --audit-level=high
```

#### 3. Secret Scanning
```yaml
secrets-scan:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
      with:
        fetch-depth: 0
    - uses: trufflesecurity/trufflehog@main
      with:
        extra_args: --only-verified
```

---

## Workflow 4: Release (`release.yml`)

**Triggers:** Tag push (v*)

```yaml
release:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
    - name: Build binaries
      run: |
        GOOS=linux GOARCH=amd64 go build -o zero-linux-amd64 ./cmd/zero
        GOOS=darwin GOARCH=amd64 go build -o zero-darwin-amd64 ./cmd/zero
        GOOS=darwin GOARCH=arm64 go build -o zero-darwin-arm64 ./cmd/zero
    - uses: softprops/action-gh-release@v1
      with:
        files: zero-*
```

---

## Required Secrets

| Secret | Description |
|--------|-------------|
| `ANTHROPIC_API_KEY` | Claude API key for code review |
| `CODECOV_TOKEN` | Codecov upload token |
| `GITHUB_TOKEN` | Auto-provided by GitHub Actions |

---

## Quality Gates

### PR Requirements
- [ ] All CI checks pass
- [ ] Claude review approves (no critical/high issues)
- [ ] Test coverage doesn't decrease
- [ ] No new security vulnerabilities

### Branch Protection Rules
```
main:
  - Require PR before merging
  - Require status checks: [build-go, build-web, lint-go, claude-review]
  - Require conversation resolution
  - Require linear history
```

---

## Implementation Tasks

### Phase 1: Basic CI
- [ ] Create `.github/workflows/ci.yml`
- [ ] Set up Go build and test
- [ ] Set up web build and lint
- [ ] Configure Codecov integration

### Phase 2: Claude Code Review
- [ ] Create `scripts/claude_review.py`
- [ ] Create `.github/workflows/code-review.yml`
- [ ] Add ANTHROPIC_API_KEY to repository secrets
- [ ] Test on sample PR

### Phase 3: Security Scanning
- [ ] Create `.github/workflows/security.yml`
- [ ] Integrate Zero self-scan
- [ ] Set up SARIF upload to GitHub Security

### Phase 4: Release Automation
- [ ] Create `.github/workflows/release.yml`
- [ ] Set up multi-platform builds
- [ ] Configure GitHub Releases

---

## Cost Considerations

### Claude API Usage
- Estimated tokens per review: ~10,000-50,000
- Cost per review: ~$0.15-0.75 (Claude Sonnet)
- Monthly estimate (50 PRs): ~$15-40

### Optimization Strategies
- Only review changed files, not entire codebase
- Cache common patterns to reduce repeated analysis
- Use Claude Haiku for initial triage, Sonnet for deep review
- Skip review for documentation-only changes
