# Ownership Scanner

The Ownership scanner analyzes code ownership patterns, contributor activity, and team health metrics. It helps identify knowledge concentration risks, orphaned code, and ownership gaps.

## Overview

| Property | Value |
|----------|-------|
| **Name** | `ownership` |
| **Version** | 1.0.0 |
| **Output File** | `ownership.json` |
| **Dependencies** | None |
| **Estimated Time** | 30-60 seconds |

## Features

### 1. Contributors (`contributors`)

Analyzes git commit history to identify contributors and their activity.

**Configuration:**
```json
{
  "contributors": {
    "enabled": true,
    "period_days": 90,
    "include_lines_changed": true,
    "group_by_email_domain": false
  }
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable contributor analysis |
| `period_days` | int | `90` | Analysis period in days |
| `include_lines_changed` | bool | `true` | Track lines added/removed |
| `group_by_email_domain` | bool | `false` | Group contributors by email domain |

**Contributor Metrics:**
- Total commits (all time)
- Commits in last 30/90/365 days
- Lines added/removed (in period)
- First and last commit dates
- Primary file areas

**Output:**
```json
{
  "contributors": [
    {
      "name": "Jane Developer",
      "email": "jane@example.com",
      "total_commits": 450,
      "commits_30d": 25,
      "commits_90d": 85,
      "commits_365d": 350,
      "lines_added_90d": 5420,
      "lines_removed_90d": 2100,
      "first_commit": "2022-01-15T10:30:00Z",
      "last_commit": "2024-12-10T14:22:00Z",
      "primary_areas": ["src/api/", "src/services/"]
    }
  ]
}
```

### 2. Bus Factor (`bus_factor`)

Calculates the bus factor - the minimum number of contributors who account for 50% of commits.

**Configuration:**
```json
{
  "bus_factor": {
    "enabled": true,
    "period_days": 365,
    "weight_by_recency": true
  }
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable bus factor calculation |
| `period_days` | int | `365` | Period for calculation |
| `weight_by_recency` | bool | `true` | Weight recent commits higher |

**Risk Classification:**

| Bus Factor | Risk Level | Description |
|------------|------------|-------------|
| 1 | Critical | Single point of failure |
| 2 | High | High knowledge concentration |
| 3-4 | Medium | Moderate risk |
| 5+ | Low | Well-distributed knowledge |

**File-Level Bus Factor:**
Also calculates bus factor per directory/file area to identify concentrated ownership.

### 3. CODEOWNERS (`codeowners`)

Parses and validates CODEOWNERS file.

**Configuration:**
```json
{
  "codeowners": {
    "enabled": true,
    "check_coverage": true,
    "validate_owners": true,
    "check_conflicts": true
  }
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable CODEOWNERS analysis |
| `check_coverage` | bool | `true` | Calculate path coverage |
| `validate_owners` | bool | `true` | Validate owner references |
| `check_conflicts` | bool | `true` | Check for conflicting rules |

**CODEOWNERS Locations:**
- `.github/CODEOWNERS`
- `CODEOWNERS`
- `docs/CODEOWNERS`

**Analysis Output:**
```json
{
  "codeowners": {
    "file_found": true,
    "location": ".github/CODEOWNERS",
    "rules_count": 25,
    "coverage_percentage": 85.5,
    "owners": ["@frontend-team", "@backend-team", "@security-team"],
    "rules": [
      {
        "pattern": "*.js",
        "owners": ["@frontend-team"],
        "line": 5
      },
      {
        "pattern": "/src/api/",
        "owners": ["@backend-team", "@security-team"],
        "line": 8
      }
    ],
    "uncovered_paths": ["scripts/", "docs/internal/"],
    "issues": []
  }
}
```

### 4. Orphaned Code (`orphans`)

Identifies files and directories with no recent activity or clear ownership.

**Configuration:**
```json
{
  "orphans": {
    "enabled": true,
    "inactive_days": 180,
    "check_codeowners_coverage": true,
    "exclude_patterns": ["vendor/", "node_modules/", "*.generated.*"]
  }
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable orphan detection |
| `inactive_days` | int | `180` | Days without commits to flag |
| `check_codeowners_coverage` | bool | `true` | Check if covered by CODEOWNERS |
| `exclude_patterns` | []string | (see above) | Patterns to exclude |

**Orphan Criteria:**
- No commits in `inactive_days`
- Not covered by CODEOWNERS
- No identifiable owner from git history

**Output:**
```json
{
  "orphaned_files": [
    {
      "path": "src/legacy/old_module.py",
      "last_modified": "2023-06-15T10:00:00Z",
      "days_inactive": 545,
      "last_author": "former-employee@example.com",
      "in_codeowners": false,
      "lines_of_code": 850
    }
  ],
  "orphaned_directories": [
    {
      "path": "src/deprecated/",
      "files_count": 15,
      "total_lines": 3200,
      "last_activity": "2023-03-20T08:00:00Z"
    }
  ]
}
```

### 5. Code Churn (`churn`)

Identifies high-churn files that may indicate instability or ongoing issues.

**Configuration:**
```json
{
  "churn": {
    "enabled": true,
    "period_days": 90,
    "min_changes": 5,
    "top_n": 30
  }
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable churn analysis |
| `period_days` | int | `90` | Analysis period |
| `min_changes` | int | `5` | Minimum changes to flag |
| `top_n` | int | `30` | Number of top files to report |

**Churn Indicators:**
- Change frequency (commits touching file)
- Number of contributors modifying file
- Lines changed (adds + deletes)
- Churn ratio (changes / file size)

**Output:**
```json
{
  "high_churn_files": [
    {
      "file": "src/api/handlers.go",
      "changes_90d": 45,
      "contributors": 8,
      "lines_added": 2500,
      "lines_removed": 1800,
      "churn_ratio": 2.5,
      "risk_level": "high"
    }
  ]
}
```

### 6. Activity Patterns (`patterns`)

Analyzes commit patterns to understand team activity.

**Configuration:**
```json
{
  "patterns": {
    "enabled": true,
    "include_time_distribution": true,
    "include_message_analysis": true
  }
}
```

**Analyzed Patterns:**
- Most active day of week
- Most active hour
- Commit frequency trends
- Weekend/after-hours commits
- Commit message patterns (fix, feat, refactor, etc.)

**Output:**
```json
{
  "patterns": {
    "most_active_day": "Tuesday",
    "most_active_hour": 14,
    "avg_commits_per_week": 45,
    "weekend_commits_percentage": 5.2,
    "commit_types": {
      "feat": 35,
      "fix": 28,
      "refactor": 15,
      "docs": 12,
      "other": 10
    },
    "first_commit": "2021-03-15T09:00:00Z",
    "last_commit": "2024-12-14T16:30:00Z"
  }
}
```

## How It Works

### Technical Flow

1. **Git Repository Open**: Uses go-git to access repository
2. **Commit Enumeration**: Walks commit history within period
3. **Contributor Analysis**: Aggregates commits by author
4. **File Diff Analysis**: Calculates lines changed per commit
5. **CODEOWNERS Parsing**: Parses and validates ownership rules
6. **Orphan Detection**: Cross-references activity with ownership
7. **Churn Calculation**: Identifies frequently modified files
8. **Pattern Analysis**: Analyzes temporal commit patterns

### Architecture

```
Git Repository
    │
    ├─► Contributors ─────► Author aggregation ─────► Contributor List
    │
    ├─► Bus Factor ───────► Commit distribution ────► Risk Score
    │
    ├─► CODEOWNERS ───────► File parsing ───────────► Ownership Rules
    │
    ├─► Orphans ──────────► Activity analysis ──────► Orphaned Code
    │
    ├─► Churn ────────────► Change frequency ───────► High-Churn Files
    │
    └─► Patterns ─────────► Temporal analysis ──────► Activity Patterns
```

## Usage

### Command Line

```bash
# Run ownership scanner
./zero scan --scanner ownership /path/to/repo

# Run with specific period
./zero scan --scanner ownership --period-days 180 /path/to/repo
```

### Programmatic Usage

```go
import "github.com/crashappsec/zero/pkg/scanners/ownership"

opts := &scanner.ScanOptions{
    RepoPath:  "/path/to/repo",
    OutputDir: "/path/to/output",
    FeatureConfig: map[string]interface{}{
        "contributors": map[string]interface{}{
            "enabled": true,
            "period_days": 90,
        },
        "bus_factor": map[string]interface{}{
            "enabled": true,
        },
        "codeowners": map[string]interface{}{
            "enabled": true,
            "check_coverage": true,
        },
        "orphans": map[string]interface{}{
            "enabled": true,
            "inactive_days": 180,
        },
    },
}

scanner := &ownership.OwnershipScanner{}
result, err := scanner.Run(ctx, opts)
```

## Output Format

```json
{
  "scanner": "ownership",
  "version": "1.0.0",
  "metadata": {
    "features_run": ["contributors", "bus_factor", "codeowners", "orphans", "churn", "patterns"],
    "analysis_period_days": 90
  },
  "summary": {
    "contributors": {
      "total_contributors": 25,
      "active_30d": 12,
      "active_90d": 18,
      "top_contributor": "jane@example.com"
    },
    "bus_factor": {
      "overall": 4,
      "risk_level": "medium",
      "critical_areas": ["src/core/"]
    },
    "codeowners": {
      "file_found": true,
      "rules_count": 25,
      "coverage_percentage": 85.5,
      "unique_owners": 5
    },
    "orphans": {
      "orphaned_files": 12,
      "orphaned_directories": 2,
      "total_orphaned_lines": 4500
    },
    "churn": {
      "high_churn_files": 8,
      "files_analyzed": 450
    },
    "patterns": {
      "avg_commits_per_week": 45,
      "most_active_day": "Tuesday"
    }
  },
  "findings": {
    "contributors": [...],
    "bus_factor": {
      "overall": 4,
      "by_directory": {
        "src/api/": 2,
        "src/core/": 1,
        "src/web/": 3
      }
    },
    "codeowners": {...},
    "orphans": {...},
    "churn": {...},
    "patterns": {...}
  }
}
```

## Prerequisites

No external tools required. Uses:
- go-git library for git analysis
- File system access for CODEOWNERS parsing

## Profiles

| Profile | contributors | bus_factor | codeowners | orphans | churn | patterns |
|---------|--------------|------------|------------|---------|-------|----------|
| `quick` | - | - | - | - | - | - |
| `standard` | Yes | Yes | Yes | - | - | - |
| `full` | Yes | Yes | Yes | Yes | Yes | Yes |
| `ownership-only` | Yes | Yes | Yes | Yes | Yes | Yes |

## Related Scanners

- **devops**: DORA metrics complement ownership data
- **health**: Uses ownership for project health scoring
- **quality**: Code quality often correlates with ownership

## See Also

- [DevOps Scanner](devops.md) - DORA metrics and git analysis
- [Health Scanner](health.md) - Overall project health
- [CODEOWNERS documentation](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners)
