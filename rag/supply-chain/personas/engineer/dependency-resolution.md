<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
SPDX-License-Identifier: GPL-3.0
-->

# Dependency Resolution and Conflict Management

## Understanding Dependency Conflicts

### What Causes Conflicts

```
your-app
├── package-a@1.0.0
│   └── shared-dep@2.x
└── package-b@1.0.0
    └── shared-dep@3.x  ← Conflict!
```

**Common causes:**
1. Different packages require incompatible versions
2. Direct dependency conflicts with transitive dependency
3. Peer dependency mismatches
4. Platform-specific version requirements

## NPM Resolution Strategies

### Default Resolution (Hoisting)

NPM tries to install a single version at the root:
```
node_modules/
├── shared-dep@3.0.0      ← Hoisted (higher version)
├── package-a/
│   └── node_modules/
│       └── shared-dep@2.0.0  ← Nested (can't use 3.x)
└── package-b/
```

### Override Conflicting Versions

**package.json overrides (NPM 8.3+):**
```json
{
  "overrides": {
    "shared-dep": "3.0.0"
  }
}
```

**Scoped overrides:**
```json
{
  "overrides": {
    "package-a": {
      "shared-dep": "3.0.0"
    }
  }
}
```

### Yarn Resolutions

**package.json:**
```json
{
  "resolutions": {
    "shared-dep": "3.0.0",
    "**/shared-dep": "3.0.0",
    "package-a/shared-dep": "3.0.0"
  }
}
```

### PNPM Overrides

**package.json:**
```json
{
  "pnpm": {
    "overrides": {
      "shared-dep": "3.0.0"
    }
  }
}
```

## Peer Dependency Handling

### What Are Peer Dependencies?

Peer dependencies declare "I need this, but you should install it":

```json
// plugin-package/package.json
{
  "peerDependencies": {
    "react": "^17.0.0 || ^18.0.0"
  }
}
```

### Peer Dependency Conflicts

**Error:**
```
npm WARN Could not resolve dependency:
peer react@"^17.0.0" from package-old@1.0.0
```

**Resolution options:**

1. **Upgrade the dependent package:**
   ```bash
   npm install package-old@latest  # May support newer peer
   ```

2. **Use legacy peer deps (temporary):**
   ```bash
   npm install --legacy-peer-deps
   ```

3. **Override peer dependency:**
   ```json
   {
     "overrides": {
       "package-old": {
         "react": "$react"  // Use app's version
       }
     }
   }
   ```

## Python Dependency Resolution

### Pip Conflict Detection

```bash
# Check for conflicts
pip check

# Output:
# package-a 1.0.0 requires shared-dep>=2.0, which is not installed.
```

### Resolution with Constraints

**constraints.txt:**
```
shared-dep==3.0.0
```

**Install with constraints:**
```bash
pip install -c constraints.txt -r requirements.txt
```

### Poetry Resolution

Poetry solves dependencies automatically:
```bash
poetry lock
poetry install
```

**Force resolution:**
```toml
# pyproject.toml
[tool.poetry.dependencies]
shared-dep = "^3.0.0"  # Direct specification wins
```

## Go Module Resolution

### Version Selection

Go uses **Minimal Version Selection (MVS)**:
- Takes the minimum version that satisfies all requirements
- Deterministic and reproducible

**Force specific version:**
```bash
go get shared-dep@v3.0.0
go mod tidy
```

**Replace directive:**
```go
// go.mod
replace shared-dep => shared-dep v3.0.0
```

## Conflict Resolution Strategies

### Strategy 1: Upgrade Everything

Best for greenfield or regular maintenance:

```bash
# NPM
rm -rf node_modules package-lock.json
npm install

# Pip
pip install --upgrade -r requirements.txt
```

### Strategy 2: Pin Problematic Package

When one package causes issues:

```json
{
  "dependencies": {
    "problematic": "1.2.3"
  },
  "overrides": {
    "problematic": "1.2.3"
  }
}
```

### Strategy 3: Replace Package

When conflict is unresolvable:

```bash
# Find alternative package
npx npm-check-updates -u problematic-package
# Research alternatives on npm
```

### Strategy 4: Fork and Fix

Last resort for abandoned packages:

1. Fork the repository
2. Update dependencies in fork
3. Use fork temporarily:
   ```json
   {
     "dependencies": {
       "package": "github:your-org/package-fork#fix-deps"
     }
   }
   ```

## Debugging Dependency Issues

### NPM

```bash
# Why is this version installed?
npm explain package-name

# Dependency tree
npm list package-name

# Full tree
npm list --all
```

### Yarn

```bash
yarn why package-name
yarn list --pattern package-name
```

### Pip

```bash
pipdeptree
pipdeptree -r -p package-name  # Reverse tree
```

### Go

```bash
go mod why package
go mod graph | grep package
```

## Version Pinning Best Practices

### When to Pin

**Pin (exact version):**
- CI/CD builds
- Production deployments
- After debugging issues

**Range (caret/tilde):**
- Development
- Libraries (give consumers flexibility)

### Pinning Examples

```json
// NPM
{
  "dependencies": {
    "exact": "1.2.3",        // Exactly 1.2.3
    "patch": "~1.2.3",       // 1.2.x
    "minor": "^1.2.3",       // 1.x.x
    "range": ">=1.2.3 <2.0.0"
  }
}
```

```
# requirements.txt
exact==1.2.3
minimum>=1.2.3
range>=1.2.3,<2.0.0
compatible~=1.2.3  # Same as >=1.2.3,<1.3.0
```

## Lock File Conflicts in Git

### Resolving Lock File Merge Conflicts

**NPM:**
```bash
# Accept current, regenerate
git checkout --ours package-lock.json
npm install

# Or regenerate completely
rm package-lock.json
npm install
git add package-lock.json
```

**Yarn:**
```bash
git checkout --ours yarn.lock
yarn install
git add yarn.lock
```

### Prevention

```bash
# Before merging
git fetch origin
npm ci  # Will fail if lock file mismatch

# After merge
npm install
git add package-lock.json
git commit --amend  # Add lock file to merge commit
```
