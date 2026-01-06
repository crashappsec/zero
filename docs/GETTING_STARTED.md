# Getting Started with Zero

This guide will help you get Zero up and running quickly.

## Prerequisites

### Required

- **Go 1.22+** - [Install Go](https://go.dev/doc/install)
- **Git** - For cloning repositories
- **GitHub CLI** - For authentication (`brew install gh`)

### Recommended Security Tools

Install these for full scanner functionality:

```bash
# SBOM generation (required)
npm install -g @cyclonedx/cdxgen    # CycloneDX generator - complete dependency analysis

# Vulnerability scanning (required)
go install github.com/google/osv-scanner/cmd/osv-scanner@latest

# Code security (highly recommended)
brew install semgrep                 # SAST, secrets (via RAG patterns), API security

# Supply chain security
go install github.com/chainguard-dev/malcontent/cmd/mal@latest

# Container security
brew install trivy

# IaC security
pip install checkov
```

## Installation

```bash
# Clone the repository
git clone https://github.com/crashappsec/zero.git
cd zero

# Build the CLI
go build -o zero ./cmd/zero

# Verify installation
./zero --help
```

## Authentication

Zero uses GitHub for repository access. Set up authentication:

```bash
# Option 1: GitHub CLI (recommended)
gh auth login

# Option 2: Environment variable
export GITHUB_TOKEN="ghp_your_token_here"
```

For AI agent features, set up Anthropic:

```bash
export ANTHROPIC_API_KEY="sk-ant-your_key_here"
```

## Check Your Setup

Run the checkup command to see what scanners will work with your current setup:

```bash
./zero checkup
```

This shows:
- Whether your GitHub token is valid
- What permissions your token has
- Which external tools are installed
- Which scanners are ready, limited, or unavailable

Example output:
```
GitHub Token Status
────────────────────────────────────────────────────────────────
  ✓ Status: Valid
    User: yourname
    Type: Classic PAT
    Scopes: repo, read:org

External Tools
────────────────────────────────────────────────────────────────
  ✓ cdxgen
  ✓ osv-scanner
  ✓ semgrep
  ✗ malcontent (not installed)

Scanner Compatibility
────────────────────────────────────────────────────────────────
  Ready
    ✓ code-packages
    ✓ code-security
    ✓ technology-identification
    ...

  Limited
    ⚠ code-packages (malcontent)
      Missing tool: malcontent
```

## Initialize Rules

Before scanning, sync Semgrep rules:

```bash
# Sync Semgrep community rules for SAST scanning (SQL injection, XSS, etc.)
./zero feeds semgrep

# Generate rules from RAG knowledge base (Zero's custom patterns)
./zero feeds rag

# Check feed status
./zero feeds status
```

Zero uses two sources of Semgrep rules:
- **Semgrep community**: Official SAST rules from semgrep.dev (vulnerabilities, secrets)
- **RAG patterns**: Custom rules generated from Zero's knowledge base (technology detection, etc.)

## Quick Start

### Analyze a Repository or Organization

```bash
# Single repository
./zero hydrate expressjs/express

# Entire organization (scans all public repos)
./zero hydrate zero-test-org

# With a specific profile
./zero hydrate expressjs/express code-security
./zero hydrate zero-test-org code-packages

# Organization with options
./zero hydrate zero-test-org --limit 10        # Limit to 10 repos
./zero hydrate zero-test-org --demo            # Skip repos > 50MB
```

### Available Profiles

Profiles are defined in `config/zero.config.json` and specify which scanners to run:

| Profile | Description | Typical Time |
|---------|-------------|--------------|
| `all-quick` | All 7 scanners (limited features) | ~2 minutes |
| `all-complete` | All 7 scanners (all features) | ~12 minutes |
| `code-packages` | SBOM + dependency analysis | ~1 minute |
| `code-security` | SAST, secrets, and crypto | ~2 minutes |
| `technology-identification` | Technology detection, ML-BOM | ~1 minute |
| `code-quality` | Quality metrics | ~1 minute |
| `devops` | IaC, containers, CI/CD, DORA | ~3 minutes |
| `developer-experience` | DevX analysis (depends on tech-id) | ~2 minutes |

### Check Analysis Status

```bash
./zero status
```

Example output:
```
Hydrated Projects
────────────────────────────────────────────────────────────────
  expressjs/express
    Path: .zero/repos/expressjs/express
    Last scanned: 2025-12-13 10:30:00
    Scanners: 7 completed

  zero-test-org/sample-app
    Path: .zero/repos/zero-test-org/sample-app
    Last scanned: 2025-12-13 09:15:00
    Scanners: 7 completed
```

### View Reports

Start the web UI to view interactive reports:

```bash
./zero serve
```

Then open http://localhost:3000 in your browser.

## Organization Scanning Options

When scanning an organization, these flags control behavior:

| Flag | Default | Description |
|------|---------|-------------|
| `--limit N` | 25 | Maximum repos to process |
| `--demo` | false | Skip repos > 50MB, fetch replacements to maintain count |

```bash
# Examples
./zero hydrate zero-test-org                    # All public repos (up to 25)
./zero hydrate zero-test-org --limit 5          # Only 5 repos
./zero hydrate zero-test-org --demo             # Demo mode for presentations
./zero hydrate zero-test-org code-security      # Security profile on all repos
```

## List Available Scanners

```bash
./zero list
```

This shows all 7 super scanners with their descriptions.

## Common Use Cases

### Security Audit

```bash
# Full security scan
./zero hydrate owner/repo code-security

# View in web UI
./zero serve
```

### Dependency Analysis

```bash
# Package-focused scan
./zero hydrate owner/repo code-packages

# View vulnerabilities
cat .zero/repos/owner/repo/analysis/code-packages.json | jq '.summary'
```

### Pre-Merge Check

```bash
# Quick scan for PR review
./zero hydrate owner/repo all-quick

# Check for critical issues
cat .zero/repos/owner/repo/analysis/code-security.json | jq '.summary'
```

### Organization-Wide Assessment

```bash
# Scan all repos with security profile
./zero hydrate zero-test-org code-security

# Check status
./zero status
```

## Understanding Scanner Output

Scanner results are stored in `.zero/repos/<owner>/<repo>/analysis/`:

```json
// code-packages.json
{
  "scanner": "code-packages",
  "version": "4.0.0",
  "timestamp": "2026-01-04T10:30:00Z",
  "summary": {
    "total_vulnerabilities": 12,
    "critical": 0,
    "high": 2,
    "medium": 5,
    "low": 5
  },
  "findings": [...]
}
```

## Configuration

Configuration is loaded from multiple sources (later overrides earlier):
1. `config/defaults/scanners.json` - Scanner feature defaults
2. `config/zero.config.json` - Main config with settings and profiles
3. `~/.zero/config.json` - User overrides (optional)

Example user override in `~/.zero/config.json`:

```json
{
  "settings": {
    "parallel_repos": 4,
    "scanner_timeout_seconds": 600
  },
  "profiles": {
    "my-custom": {
      "name": "My Custom Profile",
      "scanners": ["code-security", "code-packages"]
    }
  }
}
```

See `config/README.md` for full configuration documentation.

## Agent Mode (Claude Code)

If you have Claude Code, use the agent system for interactive analysis:

```
/agent
```

Then chat with Zero:
```
You: Are there any critical vulnerabilities in our dependencies?

Zero: Let me delegate to Cereal to analyze the vulnerability data...
```

## Troubleshooting

### "No SBOM tool available"

Install cdxgen:
```bash
npm install -g @cyclonedx/cdxgen
```

### "GitHub token invalid"

Re-authenticate:
```bash
gh auth login
# or
export GITHUB_TOKEN="ghp_new_token"
```

### Scanner times out

Increase timeout in config:
```json
{
  "settings": {
    "scanner_timeout_seconds": 600
  }
}
```

### "Permission denied" errors

Check your token permissions:
```bash
./zero checkup
```

The checkup command shows exactly what permissions you need.

## Next Steps

1. **Run `./zero checkup`** to understand your current capabilities
2. **Install missing tools** with `./zero checkup --fix`
3. **Try `./zero hydrate` on a test repo** to see it in action
4. **Explore agent mode** with `/agent` in Claude Code
5. **Review scanner results** in `.zero/repos/*/analysis/`

## Getting Help

- Run `./zero --help` for CLI help
- Run `./zero <command> --help` for command-specific help
- Check the [README](../README.md) for full documentation
- Report issues at https://github.com/crashappsec/zero/issues
