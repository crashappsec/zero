<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
SPDX-License-Identifier: GPL-3.0
-->

# Package Manager Commands Reference

## NPM (Node.js)

### Installation & Updates

```bash
# Install all dependencies
npm install
npm ci                    # Clean install from lock file (CI/CD)

# Install specific package
npm install package
npm install package@1.2.3  # Specific version
npm install package@^1.2.0 # Range

# Update packages
npm update                # All packages within semver
npm update package        # Single package
npm install package@latest # Force latest

# Uninstall
npm uninstall package
```

### Audit & Security

```bash
# Check for vulnerabilities
npm audit

# Fix vulnerabilities
npm audit fix             # Safe fixes only
npm audit fix --force     # Include breaking changes

# Generate audit report
npm audit --json > audit.json
```

### Dependency Analysis

```bash
# List outdated packages
npm outdated

# List all dependencies
npm list
npm list --depth=0        # Direct deps only
npm list package          # Specific package

# Why is package installed?
npm explain package
npm why package

# Check for unused dependencies
npx depcheck
```

### Lock File Management

```bash
# Regenerate lock file
rm package-lock.json && npm install

# Update lock file only (no node_modules change)
npm install --package-lock-only

# Deduplicate
npm dedupe
```

## Yarn (Node.js)

### Installation & Updates

```bash
# Install all dependencies
yarn install
yarn install --frozen-lockfile  # CI/CD

# Add package
yarn add package
yarn add package@1.2.3
yarn add -D package       # Dev dependency

# Update
yarn upgrade package
yarn upgrade-interactive  # Interactive updates

# Remove
yarn remove package
```

### Audit & Security

```bash
# Check vulnerabilities
yarn audit

# Fix (Yarn 2+)
yarn npm audit --all --recursive
```

### Dependency Analysis

```bash
# List outdated
yarn outdated

# Why is package installed?
yarn why package

# Deduplicate
yarn dedupe
```

## PNPM (Node.js)

### Installation & Updates

```bash
# Install
pnpm install
pnpm install --frozen-lockfile

# Add package
pnpm add package
pnpm add -D package       # Dev dependency

# Update
pnpm update
pnpm update package

# Remove
pnpm remove package
```

### Audit & Security

```bash
# Audit
pnpm audit

# Fix
pnpm audit --fix
```

## Python (pip)

### Installation & Updates

```bash
# Install from requirements
pip install -r requirements.txt

# Install package
pip install package
pip install package==1.2.3
pip install "package>=1.2,<2.0"

# Upgrade
pip install --upgrade package
pip install -U package

# Uninstall
pip uninstall package
```

### Audit & Security

```bash
# Install audit tool
pip install pip-audit

# Run audit
pip-audit
pip-audit -r requirements.txt

# Fix vulnerabilities
pip-audit --fix

# Safety (alternative)
pip install safety
safety check
```

### Dependency Analysis

```bash
# List outdated
pip list --outdated

# Show package info
pip show package

# Dependency tree
pip install pipdeptree
pipdeptree

# Generate requirements
pip freeze > requirements.txt
```

## Python (Poetry)

```bash
# Install dependencies
poetry install

# Add package
poetry add package
poetry add --dev package

# Update
poetry update
poetry update package

# Audit
poetry audit  # If available
# Or use pip-audit within poetry shell
```

## Go Modules

### Installation & Updates

```bash
# Initialize module
go mod init module-name

# Add dependency
go get package
go get package@v1.2.3
go get package@latest

# Update all
go get -u ./...

# Tidy up (remove unused)
go mod tidy

# Verify dependencies
go mod verify
```

### Audit & Security

```bash
# Install govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest

# Run audit
govulncheck ./...
```

### Dependency Analysis

```bash
# List dependencies
go list -m all

# Why is package needed?
go mod why package

# Graph
go mod graph
```

## Rust (Cargo)

### Installation & Updates

```bash
# Build (installs deps)
cargo build

# Add dependency
cargo add package
cargo add package@1.2.3

# Update
cargo update
cargo update -p package

# Remove
cargo remove package
```

### Audit & Security

```bash
# Install audit tool
cargo install cargo-audit

# Run audit
cargo audit

# Fix vulnerabilities
cargo audit fix
```

### Dependency Analysis

```bash
# List outdated
cargo outdated

# Dependency tree
cargo tree
cargo tree -p package    # Specific package
```

## Ruby (Bundler)

```bash
# Install
bundle install

# Add gem
bundle add gem_name

# Update
bundle update
bundle update gem_name

# Audit
bundle-audit check
bundle-audit update
```

## PHP (Composer)

```bash
# Install
composer install

# Add package
composer require package
composer require --dev package

# Update
composer update
composer update package

# Audit
composer audit
```

## Quick Reference Table

| Task | NPM | Yarn | Pip | Go |
|------|-----|------|-----|-----|
| Install all | `npm ci` | `yarn --frozen-lockfile` | `pip install -r req.txt` | `go mod tidy` |
| Add package | `npm i pkg` | `yarn add pkg` | `pip install pkg` | `go get pkg` |
| Update all | `npm update` | `yarn upgrade` | `pip install -U -r req.txt` | `go get -u ./...` |
| Audit | `npm audit` | `yarn audit` | `pip-audit` | `govulncheck` |
| Outdated | `npm outdated` | `yarn outdated` | `pip list -o` | `go list -m -u all` |
