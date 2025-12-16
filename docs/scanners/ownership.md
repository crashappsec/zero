# Ownership Scanner

The Ownership scanner analyzes code ownership patterns, contributor activity, programming languages, and developer competency. It helps identify knowledge concentration risks, orphaned code, ownership gaps, and developer expertise by language.

## Overview

| Property | Value |
|----------|-------|
| **Name** | `code-ownership` |
| **Version** | 1.0.0 |
| **Output File** | `code-ownership.json` |
| **Dependencies** | None |
| **Estimated Time** | 30-60 seconds |

## Features

### 1. Contributors (`contributors`)

Analyzes git commit history to identify contributors and their activity.

**Configuration:**
```json
{
  "analyze_contributors": true,
  "period_days": 90
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `analyze_contributors` | bool | `true` | Enable contributor analysis |
| `period_days` | int | `90` | Analysis period in days |

**Contributor Metrics:**
- Total commits in period
- Files touched
- Lines added/removed
- Primary file areas

**Output:**
```json
{
  "contributors": [
    {
      "name": "Jane Developer",
      "email": "jane@example.com",
      "commits": 45,
      "files_touched": 120,
      "lines_added": 5420,
      "lines_removed": 2100
    }
  ]
}
```

### 2. Languages (`languages`)

Detects programming languages used in the repository using go-enry (GitHub Linguist port).

**Configuration:**
```json
{
  "detect_languages": true
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `detect_languages` | bool | `true` | Enable language detection |

**Language Detection Features:**
- Accurate language identification using GitHub's Linguist algorithms
- Excludes vendored files (node_modules, vendor/, etc.)
- Excludes generated files
- Filters to programming languages only (excludes data, markup, prose)
- File count and percentage by language

**Output (Summary):**
```json
{
  "summary": {
    "languages_detected": 7,
    "top_languages": [
      {
        "name": "Go",
        "file_count": 245,
        "percentage": 65.5
      },
      {
        "name": "TypeScript",
        "file_count": 89,
        "percentage": 23.8
      },
      {
        "name": "Python",
        "file_count": 32,
        "percentage": 8.6
      }
    ]
  }
}
```

### 3. Developer Competency (`competency`)

Analyzes developer expertise by tracking commits per language and commit types (features, bug fixes, refactors).

**Configuration:**
```json
{
  "analyze_competency": true,
  "period_days": 90
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `analyze_competency` | bool | `true` | Enable competency analysis |
| `period_days` | int | `90` | Analysis period in days |

**Competency Metrics:**
- Commits by language
- Feature vs bug fix vs refactor ratio
- Competency score based on:
  - Total commit volume
  - Bug fix ratio (indicates deeper understanding)
  - Language breadth (number of languages)

**Commit Classification:**
Commits are classified based on message patterns:
- **Feature**: `feat`, `feature`, `add`, `implement`, `create`, `new`, `introduce`, `support`
- **Bug Fix**: `fix`, `bug`, `issue`, `patch`, `hotfix`, `resolve`, `closes #`, `fixes #`
- **Refactor**: `refactor`, `cleanup`, `clean up`, `reorganize`, `restructure`, `simplify`, `optimize`

**Output:**
```json
{
  "competencies": [
    {
      "name": "Jane Developer",
      "email": "jane@example.com",
      "total_commits": 85,
      "feature_commits": 45,
      "bug_fix_commits": 28,
      "refactor_commits": 8,
      "other_commits": 4,
      "top_language": "Go",
      "languages": [
        {
          "language": "Go",
          "file_count": 45,
          "commits": 52,
          "feature_commits": 30,
          "bug_fix_commits": 15,
          "percentage": 61.2
        },
        {
          "language": "TypeScript",
          "file_count": 28,
          "commits": 25,
          "feature_commits": 12,
          "bug_fix_commits": 10,
          "percentage": 29.4
        }
      ],
      "competency_score": 142.5
    }
  ]
}
```

**Competency Score Formula:**
```
score = commits * (1 + bug_fix_bonus) * language_bonus

where:
  bug_fix_bonus = (bug_fix_commits / total_commits) * 0.5  # Up to 50% bonus
  language_bonus = 1.0 + (language_count - 1) * 0.1        # 10% per additional language
```

### 4. CODEOWNERS (`codeowners`)

Parses and validates CODEOWNERS file.

**Configuration:**
```json
{
  "check_codeowners": true
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `check_codeowners` | bool | `true` | Enable CODEOWNERS analysis |

**CODEOWNERS Locations:**
- `CODEOWNERS`
- `.github/CODEOWNERS`
- `docs/CODEOWNERS`

**Output:**
```json
{
  "codeowners": [
    {
      "pattern": "*.js",
      "owners": ["@frontend-team"]
    },
    {
      "pattern": "/src/api/",
      "owners": ["@backend-team", "@security-team"]
    }
  ]
}
```

### 5. Orphaned Code (`orphans`)

Identifies files with no recent activity or clear ownership.

**Configuration:**
```json
{
  "detect_orphans": true,
  "period_days": 90
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `detect_orphans` | bool | `true` | Enable orphan detection |
| `period_days` | int | `90` | Days without commits to flag |

**Orphan Criteria:**
- No commits in analysis period
- No identifiable owner from git history

**Output:**
```json
{
  "orphaned_files": [
    "src/legacy/old_module.py",
    "src/deprecated/unused.go"
  ]
}
```

### 6. File Ownership (`file_owners`)

Tracks which developers have contributed to each file.

**Output:**
```json
{
  "file_owners": [
    {
      "path": "src/api/handlers.go",
      "top_contributors": ["jane@example.com", "bob@example.com"],
      "commit_count": 45
    }
  ]
}
```

## How It Works

### Technical Flow

1. **Language Detection**: Scans repository using go-enry library
2. **Git Repository Open**: Uses go-git to access repository
3. **Commit Enumeration**: Walks commit history within period
4. **Contributor Analysis**: Aggregates commits by author
5. **Competency Tracking**: Tracks per-language contributions and commit types
6. **File Diff Analysis**: Calculates files changed per commit
7. **CODEOWNERS Parsing**: Parses ownership rules
8. **Orphan Detection**: Identifies files with no recent activity

### Architecture

```
Repository
    │
    ├─► Languages ────────► go-enry detection ───────► Language Stats
    │
    ├─► Contributors ─────► Author aggregation ──────► Contributor List
    │
    ├─► Competency ───────► Per-language tracking ───► Developer Profiles
    │
    ├─► CODEOWNERS ───────► File parsing ────────────► Ownership Rules
    │
    ├─► Orphans ──────────► Activity analysis ───────► Orphaned Files
    │
    └─► File Owners ──────► Commit attribution ──────► File Ownership
```

### Language Detection

The scanner uses [go-enry](https://github.com/go-enry/go-enry), the official Go port of GitHub's Linguist library:

- **Accurate detection**: Uses same algorithms as GitHub language detection
- **Filename-based**: Fast path using filename patterns (Makefile, Dockerfile, etc.)
- **Extension-based**: Falls back to file extension matching
- **Content-based**: Can analyze file content for ambiguous cases
- **Filtering**: Excludes vendored, generated, and documentation files

## Usage

### Command Line

```bash
# Run ownership scanner
./zero scan --scanner code-ownership /path/to/repo

# Run with specific period
./zero scan --scanner code-ownership --period-days 180 /path/to/repo
```

### Programmatic Usage

```go
import codeownership "github.com/crashappsec/zero/pkg/scanners/code-ownership"

scanner := &codeownership.OwnershipScanner{}
opts := &scanner.ScanOptions{
    RepoPath:  "/path/to/repo",
    OutputDir: "/path/to/output",
}

result, err := scanner.Run(ctx, opts)
```

## Output Format

```json
{
  "analyzer": "code-ownership",
  "version": "1.0.0",
  "timestamp": "2024-12-14T10:00:00Z",
  "duration_seconds": 5,
  "repository": "/path/to/repo",
  "summary": {
    "total_contributors": 25,
    "files_analyzed": 450,
    "has_codeowners": true,
    "codeowners_rules": 15,
    "orphaned_files": 12,
    "period_days": 90,
    "languages_detected": 7,
    "top_languages": [
      {"name": "Go", "file_count": 245, "percentage": 65.5},
      {"name": "TypeScript", "file_count": 89, "percentage": 23.8}
    ]
  },
  "findings": {
    "contributors": [...],
    "codeowners": [...],
    "orphaned_files": [...],
    "file_owners": [...],
    "competencies": [...]
  },
  "metadata": {
    "features_run": ["ownership", "languages", "competency"],
    "period_days": 90
  }
}
```

## Configuration

### FeatureConfig

```go
type FeatureConfig struct {
    Enabled             bool // Enable scanner (default: true)
    AnalyzeContributors bool // Analyze git contributors (default: true)
    CheckCodeowners     bool // Validate CODEOWNERS file (default: true)
    DetectOrphans       bool // Find files with no recent commits (default: true)
    AnalyzeCompetency   bool // Analyze developer competency by language (default: true)
    DetectLanguages     bool // Detect programming languages in repo (default: true)
    PeriodDays          int  // Analysis period in days (default: 90)
}
```

### Config Presets

| Preset | Description |
|--------|-------------|
| `DefaultConfig()` | All features enabled, 90-day period |
| `QuickConfig()` | Languages and CODEOWNERS only (fast) |
| `FullConfig()` | All features, 180-day period |

## Shared Language Detection Library

The language detection functionality is available as a shared library for use by other scanners:

```go
import "github.com/crashappsec/zero/pkg/languages"

// Detect language from file path
lang := languages.DetectFromPath("src/main.go")  // Returns "Go"

// Detect with file content for accuracy
lang := languages.DetectFromFile("/path/to/file.py")

// Check language type
if languages.IsProgrammingLanguage(lang) {
    // Process programming language
}

// Scan directory for language statistics
opts := languages.DefaultScanOptions()
stats, err := languages.ScanDirectory("/path/to/repo", opts)
```

## Prerequisites

No external tools required. Uses:
- go-git library for git analysis
- go-enry library for language detection
- File system access for CODEOWNERS parsing

## Profiles

| Profile | languages | contributors | competency | codeowners | orphans |
|---------|-----------|--------------|------------|------------|---------|
| `quick` | Yes | - | - | Yes | - |
| `standard` | Yes | Yes | - | Yes | - |
| `full` | Yes | Yes | Yes | Yes | Yes |

## Related Scanners

- **devops**: DORA metrics complement ownership data
- **code-quality**: Code quality often correlates with ownership

## See Also

- [DevOps Scanner](devops.md) - DORA metrics and git analysis
- [Code Quality Scanner](quality.md) - Code quality analysis
- [go-enry documentation](https://github.com/go-enry/go-enry)
- [CODEOWNERS documentation](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners)
