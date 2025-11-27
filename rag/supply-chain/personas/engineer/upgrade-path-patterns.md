<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
SPDX-License-Identifier: GPL-3.0
-->

# Upgrade Path Patterns for Dependencies

## Semantic Versioning (SemVer) Guide

### Version Format: MAJOR.MINOR.PATCH

```
2.4.1
│ │ │
│ │ └─ PATCH: Bug fixes, security patches (backward compatible)
│ └─── MINOR: New features (backward compatible)
└───── MAJOR: Breaking changes (may require code changes)
```

### Safe Upgrade Rules

| Current → Target | Safe? | Action Required |
|------------------|-------|-----------------|
| 1.2.3 → 1.2.4 | Yes | Apply immediately |
| 1.2.3 → 1.3.0 | Usually | Review changelog, run tests |
| 1.2.3 → 2.0.0 | No | Read migration guide, allocate time |

## Upgrade Strategies by Risk Level

### Low Risk: Patch Updates

**Characteristics:**
- Same major.minor version
- Bug fixes and security patches only
- No API changes

**Process:**
```bash
# Safe to batch update
npm update  # All patch updates
pip install --upgrade-strategy only-if-needed package
```

**Testing:** Smoke tests sufficient

### Medium Risk: Minor Updates

**Characteristics:**
- Same major version
- New features added
- Deprecation warnings possible

**Process:**
1. Review changelog for deprecated APIs
2. Update an2d run full test suite
3. Check for new warnings in logs

**Testing:** Full test suite required

### High Risk: Major Updates

**Characteristics:**
- Breaking API changes
- Removed features
- Different behavior

**Process:**
1. Read migration guide thoroughly
2. Create feature branch
3. Update incrementally if multiple majors behind
4. Run comprehensive tests
5. Deploy to staging first
6. Monitor for issues

**Testing:** Full test suite + manual verification + staging deployment

## Upgrade Path Planning

### Single Major Jump (Recommended)

```
1.x → 2.x → 3.x → 4.x (current)
     ↑     ↑     ↑
   One at a time
```

**Benefits:**
- Migration guides written for N-1 → N
- Easier to debug issues
- Can roll back to intermediate version

### Skip-Version Upgrades

**When acceptable:**
- Package has LTS strategy (Node.js, Python)
- Well-documented multi-version migration
- Adequate test coverage

**Example (Node.js):**
```
Node 14 LTS → Node 18 LTS (skip 16)
# OK because LTS-to-LTS upgrade paths documented
```

## Common Upgrade Patterns

### Pattern 1: Drop-in Replacement

Package API unchanged, just version bump.

```json
// Before
"axios": "0.21.1"

// After
"axios": "0.27.2"
```

**Indicators:**
- No breaking changes in changelog
- Same function signatures
- Tests pass without changes

### Pattern 2: Deprecation Migration

Old APIs still work but deprecated.

```javascript
// Before (deprecated)
moment().format('YYYY-MM-DD')

// After (recommended)
dayjs().format('YYYY-MM-DD')  // Mostly compatible
```

**Process:**
1. Update dependency
2. Fix deprecation warnings
3. Verify functionality

### Pattern 3: API Rewrite

Complete API change requiring code updates.

```javascript
// Before (request library - deprecated)
request.get('https://api.example.com', (err, res) => { })

// After (axios - replacement)
const response = await axios.get('https://api.example.com')
```

**Process:**
1. Create adapter layer
2. Migrate usage incrementally
3. Remove adapter when complete

### Pattern 4: Ecosystem Migration

Framework or runtime upgrade affecting multiple packages.

```
React 17 → React 18
├── Update react
├── Update react-dom
├── Update react-router (if using)
├── Update testing-library
└── Update all react-* plugins
```

**Process:**
1. Identify all ecosystem dependencies
2. Check compatibility matrix
3. Update together in coordinated release

## Handling Breaking Changes

### Common Breaking Change Types

1. **Removed API**
   ```javascript
   // Old
   library.deprecatedMethod()

   // New - method removed
   // Must use alternative
   library.newMethod()
   ```

2. **Changed Function Signature**
   ```javascript
   // Old
   func(callback)

   // New
   func(options, callback)
   // or
   await func(options)  // Promise-based
   ```

3. **Changed Default Behavior**
   ```javascript
   // Old: returned null
   // New: throws error
   // Must add try/catch or null check
   ```

4. **Changed Return Type**
   ```javascript
   // Old: returned string
   // New: returns object
   const result = await func()
   // result.value instead of result
   ```

### Migration Checklist

- [ ] Read full changelog from current to target version
- [ ] Identify all breaking changes affecting your code
- [ ] Search codebase for affected APIs
- [ ] Create migration plan with estimates
- [ ] Write/update tests for changed behavior
- [ ] Update code to new APIs
- [ ] Run full test suite
- [ ] Deploy to staging
- [ ] Monitor for issues
- [ ] Deploy to production

## Lock File Strategies

### When to Regenerate Lock Files

**Regenerate:**
- Major dependency update
- Resolving complex conflicts
- Security audit failures
- CI/CD inconsistencies

**Command:**
```bash
# NPM
rm package-lock.json
npm install

# Yarn
rm yarn.lock
yarn install

# PNPM
rm pnpm-lock.yaml
pnpm install
```

### Lock File Hygiene

```bash
# Verify lock file is in sync
npm ci  # Fails if lock file mismatch

# Update lock file only
npm install --package-lock-only
```

## Rollback Strategies

### Quick Rollback

```bash
# Git-based rollback
git checkout HEAD~1 -- package.json package-lock.json
npm ci

# Or revert specific update
npm install package@previous-version
```

### Emergency Procedures

1. **Identify failing version**
2. **Pin to last known good**
3. **Deploy rollback**
4. **Investigate root cause**
5. **Plan proper upgrade**

```json
// Emergency pin in package.json
{
  "overrides": {
    "problematic-package": "1.2.3"
  }
}
```
