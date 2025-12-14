# Health Scanner

The Health scanner provides an aggregate project health score by combining metrics from other scanners. It acts as a meta-scanner that synthesizes quality, ownership, and technology data into an overall health assessment.

## Overview

| Property | Value |
|----------|-------|
| **Name** | `health` |
| **Version** | 3.0.0 |
| **Output File** | `health.json` |
| **Dependencies** | Optionally uses: quality, ownership, technology |
| **Estimated Time** | 10-30 seconds |

## Features

### 1. Health Score (`score`)

Calculates an overall project health score (0-100) by aggregating metrics from other scanners.

**Configuration:**
```json
{
  "score": {
    "enabled": true,
    "include_quality": true,
    "include_ownership": true,
    "include_security": true,
    "weights": {
      "quality": 0.30,
      "ownership": 0.25,
      "security": 0.25,
      "activity": 0.20
    }
  }
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable health scoring |
| `include_quality` | bool | `true` | Include quality metrics |
| `include_ownership` | bool | `true` | Include ownership metrics |
| `include_security` | bool | `true` | Include security metrics |
| `weights` | object | (see above) | Category weights |

**Score Components:**

| Component | Weight | Source | Metrics |
|-----------|--------|--------|---------|
| Quality | 30% | quality scanner | Tech debt, complexity, test coverage, docs |
| Ownership | 25% | ownership scanner | Bus factor, CODEOWNERS, orphaned code |
| Security | 25% | packages, code-security | Vulnerability count, secret exposure |
| Activity | 20% | ownership scanner | Recent commits, active contributors |

**Health Grade Classification:**

| Score | Grade | Description |
|-------|-------|-------------|
| 90-100 | A | Excellent health |
| 80-89 | B | Good health |
| 70-79 | C | Fair health |
| 60-69 | D | Needs attention |
| 0-59 | F | Critical issues |

### 2. Summary Aggregation (`summary`)

Aggregates key metrics from all related scanners into a single summary.

**Configuration:**
```json
{
  "summary": {
    "enabled": true,
    "include_top_issues": true,
    "max_issues": 10
  }
}
```

**Aggregated Metrics:**
- Total vulnerabilities (from packages, code-security)
- Test coverage percentage (from quality)
- Documentation score (from quality)
- Bus factor (from ownership)
- Active contributors (from ownership)
- Technology count (from technology)
- Critical issues count

### 3. Recommendations (`recommendations`)

Generates actionable recommendations based on health analysis.

**Configuration:**
```json
{
  "recommendations": {
    "enabled": true,
    "priority_threshold": "medium",
    "max_recommendations": 10
  }
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable recommendations |
| `priority_threshold` | string | `"medium"` | Minimum priority to include |
| `max_recommendations` | int | `10` | Maximum recommendations |

**Recommendation Categories:**
- **Critical**: Immediate action required (vulnerabilities, secrets)
- **High**: Should be addressed soon (low bus factor, missing tests)
- **Medium**: Improve when possible (documentation, tech debt)
- **Low**: Nice to have improvements

**Example Recommendations:**
```json
{
  "recommendations": [
    {
      "priority": "critical",
      "category": "security",
      "title": "Fix critical vulnerabilities",
      "description": "3 critical vulnerabilities found in dependencies",
      "action": "Run `npm audit fix` or update affected packages"
    },
    {
      "priority": "high",
      "category": "ownership",
      "title": "Increase bus factor",
      "description": "Bus factor is 1 - single point of failure",
      "action": "Document critical systems and cross-train team members"
    },
    {
      "priority": "medium",
      "category": "quality",
      "title": "Improve test coverage",
      "description": "Test coverage is 65%, below 80% threshold",
      "action": "Add tests for uncovered critical paths"
    }
  ]
}
```

### 4. Trends (`trends`)

Tracks health metrics over time (requires historical data).

**Configuration:**
```json
{
  "trends": {
    "enabled": true,
    "compare_to_previous": true,
    "history_days": 30
  }
}
```

**Tracked Trends:**
- Health score change
- Vulnerability count change
- Test coverage change
- Active contributors change

## How It Works

### Technical Flow

1. **Scanner Data Loading**: Loads outputs from quality, ownership, technology, packages, code-security scanners
2. **Metric Extraction**: Extracts key metrics from each scanner
3. **Score Calculation**: Calculates weighted health score
4. **Recommendation Generation**: Identifies improvement areas
5. **Summary Creation**: Aggregates top-level metrics

### Architecture

```
Scanner Outputs
    │
    ├─► quality.json ───────► Quality Metrics ──────┐
    │                                                │
    ├─► ownership.json ─────► Ownership Metrics ────┼──► Health Score
    │                                                │
    ├─► packages.json ──────► Security Metrics ─────┤
    │                                                │
    ├─► code-security.json ─► Security Metrics ─────┤
    │                                                │
    └─► technology.json ────► Tech Inventory ───────┘
                                    │
                                    ▼
                            health.json
                        (Score + Summary + Recommendations)
```

### Score Calculation

```go
healthScore =
    (qualityScore × 0.30) +
    (ownershipScore × 0.25) +
    (securityScore × 0.25) +
    (activityScore × 0.20)
```

**Component Scoring:**

| Component | Calculation |
|-----------|-------------|
| Quality | `100 - (techDebtPenalty + complexityPenalty + (100 - testCoverage)/2)` |
| Ownership | `min(busFactor × 20, 100) + codeownersCoverage/2` |
| Security | `100 - (critical×25 + high×10 + medium×3 + low×1)` |
| Activity | Based on recent commit frequency and contributor diversity |

## Usage

### Command Line

```bash
# Run health scanner
./zero scan --scanner health /path/to/repo

# Run after other scanners for full aggregation
./zero hydrate owner/repo --profile full
./zero scan --scanner health /path/to/repo
```

### Programmatic Usage

```go
import "github.com/crashappsec/zero/pkg/scanners/health"

opts := &scanner.ScanOptions{
    RepoPath:  "/path/to/repo",
    OutputDir: "/path/to/output",
    FeatureConfig: map[string]interface{}{
        "score": map[string]interface{}{
            "enabled": true,
        },
        "summary": map[string]interface{}{
            "enabled": true,
        },
        "recommendations": map[string]interface{}{
            "enabled": true,
            "max_recommendations": 10,
        },
    },
}

scanner := &health.HealthScanner{}
result, err := scanner.Run(ctx, opts)
```

## Output Format

```json
{
  "scanner": "health",
  "version": "3.0.0",
  "metadata": {
    "features_run": ["score", "summary", "recommendations"],
    "scanners_loaded": ["quality", "ownership", "packages", "code-security", "technology"]
  },
  "summary": {
    "health_score": 78,
    "health_grade": "C",
    "component_scores": {
      "quality": 75,
      "ownership": 80,
      "security": 72,
      "activity": 85
    },
    "key_metrics": {
      "vulnerabilities": {
        "critical": 0,
        "high": 3,
        "medium": 12,
        "low": 25
      },
      "test_coverage": 78.5,
      "documentation_score": 85,
      "bus_factor": 4,
      "active_contributors_30d": 8,
      "technologies_detected": 45
    },
    "top_issues": [
      {
        "severity": "high",
        "category": "security",
        "description": "3 high-severity vulnerabilities in dependencies"
      },
      {
        "severity": "medium",
        "category": "quality",
        "description": "Test coverage below 80% threshold"
      }
    ]
  },
  "findings": {
    "recommendations": [
      {
        "priority": "high",
        "category": "security",
        "title": "Address high-severity vulnerabilities",
        "description": "3 packages have high-severity vulnerabilities",
        "action": "Update lodash, axios, and express to latest versions",
        "effort": "low"
      },
      {
        "priority": "medium",
        "category": "quality",
        "title": "Improve test coverage",
        "description": "Current coverage is 78.5%, target is 80%",
        "action": "Add tests for src/api/ and src/core/ directories",
        "effort": "medium"
      }
    ],
    "trends": {
      "health_score_change": +5,
      "vulnerability_change": -2,
      "coverage_change": +3.5
    }
  }
}
```

## Prerequisites

No external tools required. Best results when run after:
- quality scanner
- ownership scanner
- packages scanner
- code-security scanner
- technology scanner

## Profiles

The health scanner is included in most profiles as an aggregator:

| Profile | Included | Pre-requisites |
|---------|----------|----------------|
| `quick` | Yes | sbom, packages |
| `standard` | Yes | sbom, packages, code-security, quality |
| `full` | Yes | All scanners |
| `health-only` | Yes | None (limited data) |

## Related Scanners

- **quality**: Provides code quality metrics
- **ownership**: Provides ownership and contributor metrics
- **packages**: Provides dependency security metrics
- **code-security**: Provides code security metrics
- **technology**: Provides technology inventory

## See Also

- [Quality Scanner](quality.md) - Code quality analysis
- [Ownership Scanner](ownership.md) - Code ownership analysis
- [Packages Scanner](packages.md) - Dependency analysis
- [Code Security Scanner](code-security.md) - Security analysis
