# Zero Configuration

Configuration files for Zero scanners, profiles, and credentials.

## Configuration Files

| File | Purpose |
|------|---------|
| `zero.config.json` | Main config: settings and scan profiles |
| `defaults/scanners.json` | Scanner feature defaults |
| `~/.zero/config.json` | User overrides (optional) |
| `~/.zero/credentials.json` | API keys and tokens |

## Configuration Loading

Zero loads configuration from multiple sources with the following priority:

1. **Scanner defaults** from `config/defaults/scanners.json`
2. **Main config** from `config/zero.config.json`
3. **User overrides** from `~/.zero/config.json` (if present)

Later sources override earlier ones.

## Main Configuration (zero.config.json)

### Settings

```json
{
  "settings": {
    "default_profile": "all-quick",
    "storage_path": ".zero",
    "parallel_repos": 8,
    "parallel_scanners": 4,
    "scanner_timeout_seconds": 300,
    "cache_ttl_hours": 24
  }
}
```

| Setting | Description | Default |
|---------|-------------|---------|
| `default_profile` | Profile to use when none specified | `all-quick` |
| `storage_path` | Where to store cloned repos and results | `.zero` |
| `parallel_repos` | Max repos to process concurrently | `8` |
| `parallel_scanners` | Max scanners per repo | `4` |
| `scanner_timeout_seconds` | Timeout for each scanner | `300` |
| `cache_ttl_hours` | How long to cache results | `24` |

### Profiles

Profiles define which scanners to run and feature overrides:

```json
{
  "profiles": {
    "all-quick": {
      "name": "All Quick",
      "description": "All scanners with fast defaults",
      "scanners": ["code-packages", "code-security", "code-quality", "devops", "technology-identification", "code-ownership", "developer-experience"],
      "feature_overrides": {
        "code-packages": {
          "malcontent": {"enabled": false}
        }
      }
    }
  }
}
```

#### Built-in Profiles

| Profile | Scanners | Description |
|---------|----------|-------------|
| `all-quick` | All 7 | Fast scan with slow features disabled |
| `all-complete` | All 7 | Complete scan with all features |
| `code-packages` | code-packages | SBOM and package analysis only |
| `code-security` | code-security | SAST, secrets, crypto only |
| `code-quality` | code-quality | Tech debt, complexity only |
| `devops` | devops | IaC, containers, CI/CD only |
| `technology-identification` | technology-identification | Tech detection only |
| `code-ownership` | code-ownership | Contributor analysis only |
| `developer-experience` | technology-identification, developer-experience | DevX analysis |

## Credentials

Credentials are managed separately from configuration for security.

### Priority Order

1. Environment variables (`GITHUB_TOKEN`, `ANTHROPIC_API_KEY`)
2. Config file (`~/.zero/credentials.json`)
3. GitHub CLI (`gh auth token`)

### Managing Credentials

```bash
# View current credentials
zero config

# Set credentials
zero config set github_token
zero config set anthropic_key

# Get specific credential
zero config get github_token

# Clear all credentials
zero config clear
```

### Credentials File Format

```json
{
  "github_token": "ghp_xxxxxxxxxxxx",
  "anthropic_api_key": "sk-ant-xxxxxxxxxxxx"
}
```

The credentials file is stored at `~/.zero/credentials.json` with restricted permissions (0600).

## Scanner Defaults (defaults/scanners.json)

Each scanner has configurable features with defaults:

```json
{
  "code-security": {
    "name": "Code Security",
    "description": "Security-focused code analysis",
    "output_file": "code-security.json",
    "features": {
      "vulns": {"enabled": true},
      "secrets": {
        "enabled": true,
        "redact_secrets": true,
        "git_history_scan": {"enabled": false}
      }
    }
  }
}
```

### Available Scanners and Features

#### code-packages
SBOM generation and package/dependency analysis.

| Feature | Default | Description |
|---------|---------|-------------|
| `generation` | enabled | SBOM generation (CycloneDX) |
| `vulns` | enabled | Vulnerability scanning |
| `health` | enabled | Package health scores |
| `licenses` | enabled | License compliance |
| `malcontent` | enabled | Malware detection |
| `typosquats` | enabled | Typosquatting detection |
| `deprecations` | enabled | Deprecated package detection |
| `duplicates` | enabled | Duplicate dependency detection |
| `confusion` | enabled | Dependency confusion detection |
| `provenance` | disabled | Package provenance verification |
| `reachability` | disabled | Reachability analysis |

#### code-security
Security-focused code analysis.

| Feature | Default | Description |
|---------|---------|-------------|
| `vulns` | enabled | SAST vulnerability detection |
| `secrets` | enabled | Secret detection |
| `api` | enabled | API security analysis |
| `ciphers` | enabled | Weak cipher detection |
| `keys` | enabled | Hardcoded key detection |
| `random` | enabled | Insecure random detection |
| `tls` | enabled | TLS configuration analysis |
| `certificates` | enabled | Certificate analysis |

#### code-quality
Code quality metrics.

| Feature | Default | Description |
|---------|---------|-------------|
| `tech_debt` | enabled | TODO/FIXME markers |
| `complexity` | enabled | Cyclomatic complexity |
| `test_coverage` | enabled | Test coverage analysis |
| `code_docs` | enabled | Documentation coverage |

#### devops
DevOps and infrastructure analysis.

| Feature | Default | Description |
|---------|---------|-------------|
| `iac` | enabled | Infrastructure as Code scanning |
| `containers` | enabled | Dockerfile analysis |
| `github_actions` | enabled | CI/CD security |
| `dora` | enabled | DORA metrics |
| `git` | enabled | Git repository analysis |

#### technology-identification
Technology detection and ML-BOM.

| Feature | Default | Description |
|---------|---------|-------------|
| `detection` | enabled | Technology detection |
| `models` | enabled | ML model detection |
| `frameworks` | enabled | Framework detection |
| `datasets` | enabled | Dataset detection |
| `ai_security` | enabled | AI security analysis |
| `ai_governance` | enabled | AI governance checks |
| `infrastructure` | enabled | Infrastructure detection |

#### code-ownership
Code ownership analysis.

| Feature | Default | Description |
|---------|---------|-------------|
| `contributors` | enabled | Contributor analysis |
| `bus_factor` | enabled | Bus factor calculation |
| `codeowners` | enabled | CODEOWNERS validation |
| `orphans` | enabled | Orphaned code detection |
| `churn` | enabled | Code churn analysis |
| `patterns` | enabled | Ownership patterns |

#### developer-experience
Developer experience analysis.

| Feature | Default | Description |
|---------|---------|-------------|
| `onboarding` | enabled | Onboarding friction |
| `sprawl` | enabled | Tool/technology sprawl |
| `workflow` | enabled | Workflow analysis |

## User Overrides

Create `~/.zero/config.json` to override settings without modifying the main config:

```json
{
  "settings": {
    "parallel_repos": 4,
    "cache_ttl_hours": 48
  },
  "profiles": {
    "my-custom": {
      "name": "My Custom Profile",
      "description": "Custom scanner selection",
      "scanners": ["code-security", "code-packages"]
    }
  }
}
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `GITHUB_TOKEN` | GitHub API token |
| `ANTHROPIC_API_KEY` | Anthropic API key for agents |
| `ZERO_HOME` | Override default storage path |

## Examples

### Run with specific profile

```bash
zero hydrate owner/repo all-complete
zero hydrate owner/repo code-security
```

### Create custom profile for security audits

Add to `~/.zero/config.json`:

```json
{
  "profiles": {
    "security-audit": {
      "name": "Security Audit",
      "description": "Deep security scan",
      "scanners": ["code-packages", "code-security"],
      "feature_overrides": {
        "code-security": {
          "secrets": {
            "git_history_scan": {"enabled": true, "max_commits": 5000}
          }
        },
        "code-packages": {
          "malcontent": {"enabled": true},
          "provenance": {"enabled": true}
        }
      }
    }
  }
}
```

Then run:
```bash
zero hydrate owner/repo security-audit
```
