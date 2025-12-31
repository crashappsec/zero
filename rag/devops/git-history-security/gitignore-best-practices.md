# Gitignore Best Practices Patterns

**Category**: devops/git-history-security
**Description**: Files that should be in .gitignore and never committed to git history
**Type**: best-practice

These patterns detect files commonly found in git repositories that should have been gitignored.
When found in git history, these indicate potential data leaks or poor repository hygiene.

---

## Environment Files

### Environment Configuration Files
**Type**: filepath
**Severity**: critical
**Category**: credentials
**Pattern**: `\.env$`
- Environment files typically contain secrets, API keys, and database credentials
- These should NEVER be committed to version control
- Remediation: Add `.env` to .gitignore, use `.env.example` for templates

### Environment Variants
**Type**: filepath
**Severity**: critical
**Category**: credentials
**Pattern**: `\.env\.(local|dev|development|prod|production|staging|test|ci)$`
- Environment-specific configuration files
- Often contain environment-specific secrets
- Remediation: Gitignore all `.env.*` files except templates

### Dotenv Backup Files
**Type**: filepath
**Severity**: high
**Category**: credentials
**Pattern**: `\.env\.bak$|\.env\.backup$|\.env\.old$`
- Backup copies of environment files
- May contain outdated but still valid credentials
- Remediation: Remove and add to .gitignore

---

## IDE and Editor Files

### JetBrains IDE Directory
**Type**: filepath
**Severity**: low
**Category**: ide
**Pattern**: `\.idea/`
- JetBrains IDE configuration directory
- May contain project-specific settings or credentials
- Remediation: Add `.idea/` to .gitignore

### VS Code Directory
**Type**: filepath
**Severity**: low
**Category**: ide
**Pattern**: `\.vscode/`
- Visual Studio Code settings directory
- May contain local workspace settings
- Remediation: Add `.vscode/` to .gitignore (or be selective about what to commit)

### Vim Swap Files
**Type**: filepath
**Severity**: info
**Category**: ide
**Pattern**: `.*\.swp$|.*\.swo$`
- Vim editor swap files
- May contain partial file contents
- Remediation: Add `*.swp` and `*.swo` to .gitignore

### Sublime Text Files
**Type**: filepath
**Severity**: info
**Category**: ide
**Pattern**: `\.sublime-workspace$|\.sublime-project$`
- Sublime Text project files
- May contain local paths
- Remediation: Add `.sublime-*` to .gitignore

---

## Operating System Files

### macOS Metadata
**Type**: filepath
**Severity**: info
**Category**: os
**Pattern**: `\.DS_Store$`
- macOS Finder metadata files
- Contains folder settings, no security risk but clutters history
- Remediation: Add `.DS_Store` to global gitignore

### Windows Thumbnails
**Type**: filepath
**Severity**: info
**Category**: os
**Pattern**: `Thumbs\.db$|ehthumbs\.db$`
- Windows Explorer thumbnail cache
- Contains image previews
- Remediation: Add `Thumbs.db` to global gitignore

### Windows Desktop Config
**Type**: filepath
**Severity**: info
**Category**: os
**Pattern**: `[Dd]esktop\.ini$`
- Windows folder configuration
- Remediation: Add `Desktop.ini` to .gitignore

---

## Build Artifacts

### Node.js Dependencies
**Type**: filepath
**Severity**: medium
**Category**: dependencies
**Pattern**: `node_modules/`
- Node.js dependency directory
- Large, reproducible from package.json
- Remediation: Add `node_modules/` to .gitignore

### Python Virtual Environments
**Type**: filepath
**Severity**: medium
**Category**: dependencies
**Pattern**: `venv/|\.venv/|env/|\.env/|virtualenv/`
- Python virtual environment directories
- Large, reproducible from requirements.txt
- Remediation: Add `venv/` and `.venv/` to .gitignore

### Python Cache
**Type**: filepath
**Severity**: info
**Category**: build
**Pattern**: `__pycache__/|\.pyc$|\.pyo$`
- Python bytecode and cache files
- Automatically generated
- Remediation: Add `__pycache__/` and `*.pyc` to .gitignore

### Go Vendor Directory
**Type**: filepath
**Severity**: medium
**Category**: dependencies
**Pattern**: `vendor/`
- Go vendor directory (unless using vendor mode)
- Remediation: Add `vendor/` to .gitignore if not vendoring

### Java Build Output
**Type**: filepath
**Severity**: info
**Category**: build
**Pattern**: `target/|\.class$`
- Maven/Gradle build output
- Remediation: Add `target/` to .gitignore

### Rust Build Output
**Type**: filepath
**Severity**: info
**Category**: build
**Pattern**: `target/|Cargo\.lock$`
- Rust cargo build output
- Note: Cargo.lock should be committed for binaries, ignored for libraries
- Remediation: Add `target/` to .gitignore

### General Build Directories
**Type**: filepath
**Severity**: low
**Category**: build
**Pattern**: `dist/|build/|out/|\.build/`
- Common build output directories
- Generated from source code
- Remediation: Add `dist/`, `build/` to .gitignore

---

## Log Files

### General Log Files
**Type**: filepath
**Severity**: medium
**Category**: logs
**Pattern**: `.*\.log$`
- Log files may contain sensitive runtime data
- Example: Database queries, user data, stack traces
- Remediation: Add `*.log` to .gitignore

### npm Debug Logs
**Type**: filepath
**Severity**: low
**Category**: logs
**Pattern**: `npm-debug\.log.*`
- npm error and debug output
- May contain system paths
- Remediation: Add `npm-debug.log*` to .gitignore

### Yarn Logs
**Type**: filepath
**Severity**: low
**Category**: logs
**Pattern**: `yarn-error\.log$|yarn-debug\.log$`
- Yarn package manager logs
- Remediation: Add `yarn-*.log` to .gitignore

---

## Coverage and Test Reports

### Code Coverage Reports
**Type**: filepath
**Severity**: low
**Category**: test
**Pattern**: `coverage/|\.coverage$|htmlcov/`
- Test coverage report directories
- Generated from test runs
- Remediation: Add `coverage/` to .gitignore

### Jest Cache
**Type**: filepath
**Severity**: low
**Category**: test
**Pattern**: `\.jest/`
- Jest test framework cache
- Remediation: Add `.jest/` to .gitignore

### pytest Cache
**Type**: filepath
**Severity**: low
**Category**: test
**Pattern**: `\.pytest_cache/`
- pytest cache directory
- Remediation: Add `.pytest_cache/` to .gitignore

---

## Package Manager Lock Files (Contextual)

### Composer Lock
**Type**: filepath
**Severity**: info
**Category**: dependencies
**Pattern**: `composer\.lock$`
- PHP dependency lock file
- Should be committed for applications, often ignored for libraries
- Note: Context-dependent

### Gemfile Lock
**Type**: filepath
**Severity**: info
**Category**: dependencies
**Pattern**: `Gemfile\.lock$`
- Ruby dependency lock file
- Should be committed for applications
- Note: Context-dependent

---

## Temporary and Cache Files

### Temporary Files
**Type**: filepath
**Severity**: info
**Category**: temporary
**Pattern**: `.*\.tmp$|.*\.temp$|tmp/|temp/`
- Temporary files and directories
- Should never be committed
- Remediation: Add `*.tmp`, `*.temp`, `tmp/` to .gitignore

### Cache Directories
**Type**: filepath
**Severity**: info
**Category**: temporary
**Pattern**: `\.cache/|cache/`
- Cache directories
- Remediation: Add `.cache/` to .gitignore
