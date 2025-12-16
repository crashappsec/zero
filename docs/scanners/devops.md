# DevOps Scanner

The DevOps scanner provides comprehensive DevOps and CI/CD security analysis, including Infrastructure as Code (IaC) scanning, container security, GitHub Actions analysis, DORA metrics, and git insights.

## Overview

| Property | Value |
|----------|-------|
| **Name** | `devops` |
| **Version** | 3.0.0 |
| **Output File** | `devops.json` |
| **Dependencies** | None |
| **Estimated Time** | 60-180 seconds |

## Features

### 1. Infrastructure as Code (`iac`)

Scans IaC configurations for security misconfigurations.

**Configuration:**
```json
{
  "iac": {
    "enabled": true,
    "tool": "auto",
    "fallback_tool": true
  }
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable IaC scanning |
| `tool` | string | `"auto"` | Tool: `"auto"`, `"checkov"`, or `"trivy"` |
| `fallback_tool` | bool | `true` | Fall back to Trivy if Checkov fails |

**Tool Selection:**
- `auto`: Prefers Checkov, falls back to Trivy
- `checkov`: Bridgecrew Checkov (more comprehensive)
- `trivy`: Aqua Trivy (faster)

**Supported IaC Types:**

| Type | File Patterns |
|------|---------------|
| Terraform | `*.tf`, `*.tfvars` |
| Kubernetes | `*.yaml`, `*.yml` with k8s resources |
| Dockerfile | `Dockerfile`, `*.Dockerfile`, `Dockerfile.*` |
| CloudFormation | `*.yaml`, `*.json` with CF templates |
| Helm | `Chart.yaml`, templates |
| Azure ARM | `*.json` with ARM schemas |

**Severity Classification:**
Checkov findings are classified based on check ID patterns:
- **Critical**: public, encrypt, privileged, root, admin
- **High**: auth, secret, password, credential, key, token
- **Medium**: log, monitor, backup, version, ssl, tls
- **Low**: Other findings

### 2. Containers (`containers`)

Scans container images referenced in Dockerfiles.

**Configuration:**
```json
{
  "containers": {
    "enabled": true,
    "scan_base_images": true,
    "check_hardened": true
  }
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable container scanning |
| `scan_base_images` | bool | `true` | Scan base images for vulns |
| `check_hardened` | bool | `true` | Check for hardened images |

**Detection Process:**
1. Finds all Dockerfiles in repository
2. Extracts base images from `FROM` statements
3. Scans each unique image with Trivy
4. Reports vulnerabilities found in images

**Output Data:**
```go
type ContainerFinding struct {
    VulnID       string   // CVE ID
    Title        string   // Vulnerability title
    Description  string   // Description
    Severity     string   // critical, high, medium, low
    Image        string   // Base image name
    Dockerfile   string   // Source Dockerfile path
    Package      string   // Affected package
    Version      string   // Installed version
    FixedVersion string   // Version with fix
    CVSS         float64  // CVSS score
    References   []string // Links to advisories
}
```

### 3. GitHub Actions (`github_actions`)

Security analysis of GitHub Actions workflows.

**Configuration:**
```json
{
  "github_actions": {
    "enabled": true,
    "check_pinning": true,
    "check_secrets": true,
    "check_injection": true,
    "check_permissions": true
  }
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable GHA scanning |
| `check_pinning` | bool | `true` | Check action version pinning |
| `check_secrets` | bool | `true` | Check secret handling |
| `check_injection` | bool | `true` | Check for injection risks |
| `check_permissions` | bool | `true` | Check permission scope |

**Detected Issues:**

| Category | Severity | Pattern | Description |
|----------|----------|---------|-------------|
| `unpinned-action` | high | `uses: owner/action@v1` | Action not pinned to SHA |
| `secret-in-run` | high | `${{ secrets.X }}` in run | Secret may be exposed |
| `injection-risk` | critical | `${{ github.event.issue.* }}` | Command injection from untrusted input |
| `excessive-permissions` | high | `permissions: write-all` | Overly broad permissions |
| `write-permissions` | medium | `contents: write` | Write permission granted |

**Suggestions:**
- Pin actions to specific commit SHAs for security
- Pass secrets through environment variables, not directly in run
- Sanitize untrusted input before use
- Use minimal required permissions

### 4. DORA Metrics (`dora`)

Calculates DevOps Research and Assessment metrics.

**Configuration:**
```json
{
  "dora": {
    "enabled": true,
    "period_days": 90
  }
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable DORA metrics |
| `period_days` | int | `90` | Analysis period in days |

**Calculated Metrics:**

| Metric | Description | Measurement |
|--------|-------------|-------------|
| Deployment Frequency | How often deployments occur | Deployments per week |
| Lead Time for Changes | Time from commit to deploy | Hours |
| Change Failure Rate | Percentage of failed changes | Percentage |
| Mean Time to Recovery (MTTR) | Time to recover from failures | Hours |

**Classification:**

| Metric | Elite | High | Medium | Low |
|--------|-------|------|--------|-----|
| Deploy Frequency | ≥7/week | ≥1/week | ≥0.25/week | <0.25/week |
| Lead Time | <24h | <168h (1w) | <720h (30d) | ≥720h |
| Change Failure Rate | ≤5% | ≤10% | ≤15% | >15% |
| MTTR | <1h | <24h | <168h | ≥168h |

**Detection Method:**
- Uses git tags matching release patterns (`v1.0.0`, `1.2.3`)
- If no tags, uses weekly commit aggregation as proxy
- Detects fix deployments by keywords: fix, hotfix, patch, bugfix

### 5. Git Insights (`git`)

Analyzes git history for contributor patterns and code health.

**Configuration:**
```json
{
  "git": {
    "enabled": true,
    "include_churn": true,
    "include_age": true,
    "include_patterns": true,
    "include_branches": true
  }
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable git analysis |
| `include_churn` | bool | `true` | Analyze high-churn files |
| `include_age` | bool | `true` | Analyze code age |
| `include_patterns` | bool | `true` | Analyze commit patterns |
| `include_branches` | bool | `true` | Analyze branch info |

**Contributor Analysis:**
- Total commits by contributor
- Commits in last 30/90/365 days
- Lines added/removed (90 days)
- Bus factor calculation

**Bus Factor:**
Minimum number of contributors who account for 50% of commits. Low bus factor (1-2) indicates knowledge concentration risk.

**High Churn Files:**
Files with frequent changes (≥5 changes in 90 days) that may indicate:
- Instability
- Ongoing development
- Technical debt

**Code Age Distribution:**
Categorizes files by last modification:
- 0-30 days (recent)
- 31-90 days
- 91-365 days
- 365+ days (stale)

**Commit Patterns:**
- Most active day of week
- Most active hour
- Average commits per week
- First and last commit dates

**Activity Level Classification:**

| Level | 90-day Commits |
|-------|----------------|
| very_high | >500 |
| high | 200-500 |
| medium | 50-200 |
| low | <50 |

## How It Works

### Technical Flow

1. **Parallel Execution**: All 5 features run concurrently
2. **Tool Invocation**: Checkov/Trivy for IaC, Trivy for containers
3. **Git Analysis**: Uses go-git library for git analysis
4. **Pattern Matching**: Uses regex for GitHub Actions scanning
5. **Aggregation**: Combines results from all features

### Architecture

```
Repository
    │
    ├─► IaC Feature ───► Checkov/Trivy ───► IaC Findings
    │
    ├─► Containers ────► Find Dockerfiles ─► Trivy Image Scan ─► Container Findings
    │
    ├─► GHA Feature ───► .github/workflows/*.yml ───► GHA Findings
    │
    ├─► DORA Feature ──► Git Tags ─► Release Analysis ─► DORA Metrics
    │
    └─► Git Feature ───► Git History ─► Contributor/Churn Analysis ─► Git Insights
```

## Usage

### Command Line

```bash
# Run devops scanner only
./zero scan --scanner devops /path/to/repo

# Run devops profile
./zero hydrate owner/repo --profile devops-only
```

### Programmatic Usage

```go
import "github.com/crashappsec/zero/pkg/scanners/devops"

opts := &scanner.ScanOptions{
    RepoPath:  "/path/to/repo",
    OutputDir: "/path/to/output",
    FeatureConfig: map[string]interface{}{
        "iac": map[string]interface{}{
            "enabled": true,
            "tool": "checkov",
        },
        "containers": map[string]interface{}{
            "enabled": true,
            "scan_base_images": true,
        },
        "github_actions": map[string]interface{}{
            "enabled": true,
        },
        "dora": map[string]interface{}{
            "enabled": true,
            "period_days": 90,
        },
        "git": map[string]interface{}{
            "enabled": true,
        },
    },
}

scanner := &devops.DevOpsScanner{}
result, err := scanner.Run(ctx, opts)
```

## Output Format

```json
{
  "scanner": "devops",
  "version": "3.0.0",
  "metadata": {
    "features_run": ["iac", "containers", "github_actions", "dora", "git"]
  },
  "summary": {
    "iac": {
      "tool": "checkov",
      "files_scanned": 15,
      "total_findings": 23,
      "critical": 2,
      "high": 8,
      "medium": 10,
      "low": 3,
      "by_type": {
        "terraform": 15,
        "kubernetes": 5,
        "dockerfile": 3
      }
    },
    "containers": {
      "dockerfiles_scanned": 3,
      "images_scanned": 2,
      "total_findings": 45,
      "critical": 5,
      "high": 15,
      "medium": 20,
      "low": 5,
      "by_image": {
        "node:18": 25,
        "python:3.11": 20
      }
    },
    "github_actions": {
      "workflows_scanned": 5,
      "total_findings": 8,
      "critical": 1,
      "high": 4,
      "medium": 3,
      "by_category": {
        "unpinned-action": 4,
        "injection-risk": 1,
        "write-permissions": 3
      }
    },
    "dora": {
      "period_days": 90,
      "deployment_frequency": 2.5,
      "deployment_frequency_class": "high",
      "lead_time_hours": 48,
      "lead_time_class": "high",
      "change_failure_rate": 8.5,
      "change_failure_class": "high",
      "mttr_hours": 4.2,
      "mttr_class": "high",
      "overall_class": "high"
    },
    "git": {
      "total_commits": 1250,
      "total_contributors": 15,
      "active_contributors_30d": 8,
      "active_contributors_90d": 12,
      "commits_90d": 350,
      "bus_factor": 3,
      "activity_level": "high"
    }
  },
  "findings": {
    "iac": [...],
    "containers": [...],
    "github_actions": [...],
    "dora": {
      "total_deployments": 23,
      "total_commits": 350,
      "deployments": [...]
    },
    "git": {
      "contributors": [...],
      "high_churn_files": [...],
      "code_age": {...},
      "patterns": {...},
      "branches": {...}
    }
  }
}
```

## Prerequisites

| Tool | Required For | Install Command |
|------|--------------|-----------------|
| checkov | IaC scanning | `pip install checkov` |
| trivy | IaC fallback, container scanning | `brew install trivy` |

**Note:** DORA and Git features use go-git and require no external tools.

## Profiles

| Profile | iac | containers | github_actions | dora | git |
|---------|-----|------------|----------------|------|-----|
| `quick` | - | - | - | - | - |
| `standard` | - | - | - | - | - |
| `security` | Yes | Yes | Yes | - | - |
| `full` | Yes | Yes | Yes | Yes | Yes |
| `devops-only` | Yes | Yes | Yes | Yes | Yes |
| `ci-cd` | - | - | Yes | Yes | - |

## Related Scanners

- **code-ownership**: Overlaps on code ownership via git analysis
- **code-security**: May detect similar issues in IaC

## See Also

- [Code Ownership Scanner](ownership.md) - Code ownership analysis
- [DORA Metrics](https://dora.dev/research/) - DevOps performance metrics
- [Checkov](https://www.checkov.io/) - IaC security scanner
- [Trivy](https://trivy.dev/) - Container and IaC scanner
