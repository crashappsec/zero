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
# SBOM generation (at least one required)
npm install -g @cyclonedx/cdxgen    # Preferred - complete dependency analysis
brew install syft                    # Fallback - fast static analysis

# Vulnerability scanning (at least one required)
brew install grype                   # Recommended
go install github.com/google/osv-scanner/cmd/osv-scanner@latest

# Code security (highly recommended)
brew install semgrep                 # SAST, secrets, API security
brew install gitleaks                # Secrets detection

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
go build -o main ./cmd/zero

# Verify installation
./main --help
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
./main checkup
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
  ✓ syft
  ✓ grype
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

## Quick Start

### Analyze a Repository

```bash
# Clone and scan a repository (uses default profile from config)
./main hydrate expressjs/express

# With a specific profile (profile is a positional argument)
./main hydrate expressjs/express code-security
./main hydrate expressjs/express code-packages
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
./main status
```

Example output:
```
Hydrated Projects
────────────────────────────────────────────────────────────────
  expressjs/express
    Path: .zero/repos/expressjs/express
    Last scanned: 2025-12-13 10:30:00
    Scanners: 12 completed

  phantom-tests/platform
    Path: .zero/repos/phantom-tests/platform
    Last scanned: 2025-12-13 09:15:00
    Scanners: 8 completed
```

### Generate Reports

```bash
./main report expressjs/express
```

## Scanning an Organization

Scan all repositories in a GitHub organization (target without `/` is treated as org):

```bash
# Scan all public repos (uses default profile)
./main hydrate myorganization

# With a specific profile
./main hydrate myorganization security
./main hydrate myorganization quick

# Limit number of repos
./main hydrate myorganization --limit 10

# Skip slow scanners for faster org scans
./main hydrate myorganization quick --skip-slow
```

## List Available Scanners

```bash
./main list
```

This shows all 7 super scanners with their descriptions.

## Common Use Cases

### Security Audit

```bash
# Full security scan
./main hydrate owner/repo security

# Check the report
./main report owner/repo
```

### Dependency Analysis

```bash
# Package-focused scan
./main hydrate owner/repo code-packages

# View vulnerabilities
cat .zero/repos/owner/repo/analysis/code-packages.json | jq '.summary'
```

### Pre-Merge Check

```bash
# Quick scan for PR review
./main hydrate owner/repo all-quick

# Check for critical issues
cat .zero/repos/owner/repo/analysis/code-security.json | jq '.summary'
```

### Organization-Wide Assessment

```bash
# Scan all repos with security profile
./main hydrate myorg security

# Check status
./main status
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

Create or edit `config/zero.config.json`:

```json
{
  "settings": {
    "default_profile": "all-quick",
    "scanner_timeout_seconds": 300,
    "parallel_jobs": 4
  },
  "scanners": {
    "code-packages": {
      "features": {
        "generation": { "enabled": true, "tool": "auto" },
        "vulns": { "enabled": true },
        "health": { "enabled": true }
      }
    }
  }
}
```

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

Install cdxgen or syft:
```bash
npm install -g @cyclonedx/cdxgen
# or
brew install syft
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
./main checkup
```

The checkup command shows exactly what permissions you need.

## Next Steps

1. **Run `./main checkup`** to understand your current capabilities
2. **Install missing tools** with `./main checkup --fix`
3. **Try `./main hydrate` on a test repo** to see it in action
4. **Explore agent mode** with `/agent` in Claude Code
5. **Review scanner results** in `.zero/repos/*/analysis/`

## Getting Help

- Run `./main --help` for CLI help
- Run `./main <command> --help` for command-specific help
- Check the [README](../README.md) for full documentation
- Report issues at https://github.com/crashappsec/zero/issues
