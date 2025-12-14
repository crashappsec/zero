# Packages Scanner

The Packages scanner provides comprehensive supply chain security analysis for all project dependencies. It **depends on the SBOM scanner** and does not generate its own SBOM.

## Overview

| Property | Value |
|----------|-------|
| **Name** | `packages` |
| **Version** | 3.0.0 |
| **Output File** | `packages.json` |
| **Dependencies** | `sbom` (required) |
| **Estimated Time** | 60-180 seconds |

## Features

### 1. Vulnerabilities (`vulns`)

Scans dependencies for known CVEs using OSV-Scanner.

**Configuration:**
```json
{
  "vulns": {
    "enabled": true,
    "include_dev": false,
    "check_reachability": true,
    "include_kev": true
  }
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable vulnerability scanning |
| `include_dev` | bool | `false` | Include dev dependencies |
| `check_reachability` | bool | `true` | Check if vulns are reachable |
| `include_kev` | bool | `true` | Enrich with CISA KEV data |

**Data Sources:**
- OSV (Open Source Vulnerabilities)
- CISA KEV (Known Exploited Vulnerabilities)

**Severity Classification:**
| CVSS Score | Severity |
|------------|----------|
| 9.0+ | Critical |
| 7.0-8.9 | High |
| 4.0-6.9 | Medium |
| 0.1-3.9 | Low |

### 2. Health (`health`)

Assesses package maintenance and community health using deps.dev API.

**Configuration:**
```json
{
  "health": {
    "enabled": true,
    "check_deprecated": true,
    "check_maintained": true,
    "check_downloads": true,
    "max_packages": 50
  }
}
```

**Health Signals:**
- Deprecation status
- Latest version availability
- OpenSSF Scorecard score
- Maintenance activity

**Health Statuses:**
| Status | Condition |
|--------|-----------|
| `critical` | Package is deprecated |
| `warning` | Health score < 5 |
| `healthy` | All checks pass |

### 3. Licenses (`licenses`)

Analyzes license compliance across the dependency tree.

**Configuration:**
```json
{
  "licenses": {
    "enabled": true,
    "include_dev": false,
    "blocked_licenses": ["GPL-3.0", "AGPL-3.0"],
    "allowed_licenses": []
  }
}
```

**Default Allowed Licenses:**
- MIT, Apache-2.0, BSD-2-Clause, BSD-3-Clause
- ISC, Unlicense, CC0-1.0, 0BSD

**Default Blocked Licenses:**
- GPL-2.0, GPL-3.0, AGPL-3.0, SSPL-1.0

**License Statuses:**
| Status | Description |
|--------|-------------|
| `allowed` | License is in allowed list |
| `denied` | License is in blocked list |
| `review` | License needs manual review |
| `unknown` | No license detected |

### 4. Malcontent (`malcontent`)

Behavioral analysis for malicious code patterns using the malcontent tool.

**Configuration:**
```json
{
  "malcontent": {
    "enabled": true,
    "min_risk": "medium"
  }
}
```

| Risk Level | Description |
|------------|-------------|
| `critical` | Active exploitation indicators |
| `high` | Data exfiltration, code execution |
| `medium` | Suspicious behaviors |
| `low` | Minor concerns |

**Detected Behaviors:**
- Data exfiltration patterns
- Code execution capabilities
- Persistence mechanisms
- Network communications
- File system operations
- Post-install script analysis

### 5. Dependency Confusion (`confusion`)

Detects dependency confusion attack vectors.

**Configuration:**
```json
{
  "confusion": {
    "enabled": true,
    "check_internal_names": true,
    "check_npm": true,
    "check_pypi": true
  }
}
```

**Detection Logic:**
- Identifies packages with internal-looking names (e.g., `internal-`, `private-`, `-internal`)
- Checks if those names exist on public registries (npm, PyPI)
- Flags potential dependency confusion risks

### 6. Typosquatting (`typosquats`)

Detects potential typosquatting attacks.

**Configuration:**
```json
{
  "typosquats": {
    "enabled": true,
    "check_similar_names": true,
    "check_new_packages": true
  }
}
```

**Checks:**
- Name similarity to popular packages (lodash, express, react, etc.)
- Package age (new packages < 30 days are flagged)

### 7. Deprecations (`deprecations`)

Identifies deprecated packages across ecosystems.

**Configuration:**
```json
{
  "deprecations": {
    "enabled": true,
    "check_npm": true,
    "check_pypi": true,
    "check_go": true
  }
}
```

**Ecosystem-Specific Checks:**
- npm: `deprecated` field in package metadata
- PyPI: Development Status classifiers (Inactive, Deprecated)
- Go: Retract directives in go.mod

### 8. Duplicates (`duplicates`)

Identifies duplicate dependencies and functionality overlap.

**Configuration:**
```json
{
  "duplicates": {
    "enabled": true,
    "check_versions": true,
    "check_functionality": true
  }
}
```

**Duplicate Types:**
- **Version duplicates**: Same package with multiple versions
- **Functionality duplicates**: Multiple packages serving same purpose (e.g., moment + dayjs + date-fns)

**Known Functional Groups:**
- Date: moment, dayjs, date-fns, luxon
- HTTP: axios, node-fetch, got, request, superagent
- Lodash: lodash, underscore, ramda
- Promise: bluebird, q, when

### 9. Reachability (`reachability`)

Filters vulnerabilities based on call graph reachability.

**Configuration:**
```json
{
  "reachability": {
    "enabled": true,
    "filter_unreachable": false
  }
}
```

**Supported Ecosystems:**
- Go
- Python
- Rust

Uses OSV-Scanner's experimental call analysis to determine if vulnerable code paths are actually called.

### 10. Provenance (`provenance`)

Verifies build provenance and supply chain integrity.

**Configuration:**
```json
{
  "provenance": {
    "enabled": true,
    "check_sigstore": true,
    "check_slsa": true
  }
}
```

**Checks:**
- Sigstore signatures
- SLSA provenance attestations

### 11. Bundle (`bundle`)

Analyzes bundle size impact for npm packages.

**Configuration:**
```json
{
  "bundle": {
    "enabled": true,
    "size_threshold_kb": 100
  }
}
```

Only applies to npm/JavaScript projects.

### 12. Recommendations (`recommendations`)

Generates actionable recommendations based on scan results.

**Configuration:**
```json
{
  "recommendations": {
    "enabled": true
  }
}
```

Recommendations are generated based on:
- Critical vulnerabilities found
- Deprecated packages detected
- Health issues identified

## How It Works

### Technical Flow

1. **SBOM Loading**: Loads component data from `sbom.cdx.json`
2. **Parallel Execution**: Runs vulns, health, licenses, malcontent, confusion, typosquats, deprecations, duplicates in parallel
3. **Sequential Execution**: Runs reachability, provenance, bundle, recommendations sequentially (order-dependent)
4. **Aggregation**: Combines results from all features
5. **Output**: Writes `packages.json` with all findings

### Architecture

```
SBOM Scanner Output
        │
        ▼
┌───────────────────┐
│  packages.json    │◄── SBOM components loaded
└───────────────────┘
        │
        ▼
┌─────────────────────────────────────────┐
│          Parallel Features              │
├─────────┬─────────┬─────────┬──────────┤
│  vulns  │ health  │licenses │malcontent│
├─────────┼─────────┼─────────┼──────────┤
│confusion│typosquat│deprecat.│duplicates│
└─────────┴─────────┴─────────┴──────────┘
        │
        ▼
┌─────────────────────────────────────────┐
│         Sequential Features             │
├─────────┬─────────┬─────────┬──────────┤
│reachab. │provenance│ bundle │recommend.│
└─────────┴─────────┴─────────┴──────────┘
        │
        ▼
    packages.json
```

## Usage

### Command Line

```bash
# Run packages scanner (requires sbom to run first)
./zero scan --scanner packages /path/to/repo

# Run packages profile (includes sbom + packages)
./zero hydrate owner/repo --profile packages-only
```

### Configuration Profiles

| Profile | Features Enabled |
|---------|------------------|
| `quick` | vulns, health (no malcontent, provenance, bundle, confusion, reachability) |
| `standard` | vulns, health, licenses, malcontent, confusion, typosquats, deprecations, duplicates (no bundle, reachability) |
| `security` | All features enabled |
| `full` | All features enabled |

## Output Format

```json
{
  "scanner": "packages",
  "version": "3.0.0",
  "metadata": {
    "features_run": ["vulns", "health", "licenses", "malcontent"],
    "sbom_source": "sbom scanner",
    "component_count": 245
  },
  "summary": {
    "vulns": {
      "total_vulnerabilities": 12,
      "critical": 1,
      "high": 3,
      "medium": 5,
      "low": 3,
      "kev_count": 1
    },
    "health": {
      "total_packages": 50,
      "analyzed_count": 50,
      "healthy_count": 45,
      "warning_count": 3,
      "critical_count": 2,
      "deprecated_count": 2,
      "outdated_count": 10
    },
    "licenses": {
      "total_packages": 245,
      "unique_licenses": 8,
      "allowed": 200,
      "denied": 5,
      "needs_review": 20,
      "unknown": 20,
      "policy_violations": 5
    },
    "malcontent": {
      "total_files": 1500,
      "files_with_risk": 12,
      "critical": 0,
      "high": 2,
      "medium": 5,
      "low": 5
    }
  },
  "findings": {
    "vulns": [
      {
        "id": "GHSA-xxxx-xxxx-xxxx",
        "aliases": ["CVE-2024-1234"],
        "package": "lodash",
        "version": "4.17.20",
        "ecosystem": "npm",
        "severity": "high",
        "title": "Prototype Pollution",
        "in_kev": false
      }
    ],
    "health": [...],
    "licenses": [...],
    "malcontent": [...]
  }
}
```

## Prerequisites

| Tool | Required For | Install Command |
|------|--------------|-----------------|
| osv-scanner | vulns, reachability | `go install github.com/google/osv-scanner/cmd/osv-scanner@latest` |
| malcontent | malcontent | `brew install malcontent` |

## Related Scanners

- **sbom**: Required - provides component data
- **crypto**: May overlap on some security findings
- **code-security**: May detect similar secrets/vulns in code

## See Also

- [SBOM Scanner](sbom.md) - Generates the component data this scanner consumes
- [AI Scanner](ai.md) - ML-BOM generation for AI dependencies
- [OSV Database](https://osv.dev/) - Vulnerability data source
- [deps.dev](https://deps.dev/) - Package health data source
