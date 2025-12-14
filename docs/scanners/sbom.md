# SBOM Scanner

The SBOM (Software Bill of Materials) scanner is the **source of truth** for all package and component data in Zero. Other scanners that need dependency information (like the packages scanner) depend on the SBOM scanner's output.

## Overview

| Property | Value |
|----------|-------|
| **Name** | `sbom` |
| **Version** | 1.0.0 |
| **Output File** | `sbom.json` + `sbom.cdx.json` |
| **Dependencies** | None (runs first) |
| **Estimated Time** | 30-60 seconds |

## Features

### 1. Generation

Generates a CycloneDX-format SBOM from the repository.

**Configuration:**
```json
{
  "generation": {
    "enabled": true,
    "tool": "auto",
    "spec_version": "1.5",
    "fallback_to_syft": true,
    "include_dev": false,
    "deep": false
  }
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable SBOM generation |
| `tool` | string | `"auto"` | SBOM generator: `"auto"`, `"cdxgen"`, or `"syft"` |
| `spec_version` | string | `"1.5"` | CycloneDX spec version |
| `fallback_to_syft` | bool | `true` | Fall back to syft if cdxgen fails |
| `include_dev` | bool | `false` | Include dev dependencies |
| `deep` | bool | `false` | Enable deep analysis mode |

**Tool Selection:**
- `auto`: Prefers cdxgen, falls back to syft
- `cdxgen`: CycloneDX Generator (more detailed, slower)
- `syft`: Anchore Syft (faster, lighter)

**Output:** Creates `sbom.cdx.json` in CycloneDX format containing:
- All components (packages, libraries, frameworks)
- Package URLs (purls) for each component
- License information per component
- Hashes/checksums where available
- Dependency relationships

### 2. Integrity

Verifies SBOM completeness against lockfiles.

**Configuration:**
```json
{
  "integrity": {
    "enabled": true,
    "verify_lockfiles": true,
    "detect_drift": true,
    "check_completeness": true
  }
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable integrity verification |
| `verify_lockfiles` | bool | `true` | Compare SBOM against lockfiles |
| `detect_drift` | bool | `true` | Detect drift from expected state |
| `check_completeness` | bool | `true` | Verify all packages are in SBOM |

**Supported Lockfiles:**
- `package-lock.json` (npm v1 and v2+ formats)
- `yarn.lock`
- `go.sum`
- `requirements.txt`

## How It Works

### Technical Flow

1. **Tool Detection**: Checks for `cdxgen` or `syft` availability
2. **SBOM Generation**: Runs the selected tool against the repository
3. **Parsing**: Parses the generated CycloneDX JSON
4. **Component Extraction**: Extracts components, licenses, hashes, and dependencies
5. **Integrity Verification**: Compares SBOM against lockfiles (if enabled)
6. **Output**: Writes both `sbom.json` (summary) and `sbom.cdx.json` (full SBOM)

### Component Data

Each component extracted includes:
```go
type Component struct {
    Type       string     // library, framework, application
    Name       string     // package name
    Version    string     // package version
    Purl       string     // package URL (pkg:npm/lodash@4.17.21)
    Ecosystem  string     // npm, pypi, golang, etc.
    Scope      string     // required, optional, dev
    Licenses   []string   // detected licenses
    Hashes     []Hash     // SHA-256, SHA-512, etc.
    Properties []Property // additional metadata
}
```

### Lockfile Comparison

The integrity feature compares SBOM contents against lockfiles to detect:
- **Missing packages**: In lockfile but not in SBOM
- **Extra packages**: In SBOM but not in lockfile
- **Matched packages**: Successfully reconciled

## Usage

### Command Line

```bash
# Run SBOM scanner only
./zero scan --scanner sbom /path/to/repo

# Run with specific profile
./zero hydrate owner/repo --profile sbom-only
```

### Programmatic Usage

```go
import "github.com/crashappsec/zero/pkg/scanners/sbom"

opts := &scanner.ScanOptions{
    RepoPath:  "/path/to/repo",
    OutputDir: "/path/to/output",
    FeatureConfig: map[string]interface{}{
        "generation": map[string]interface{}{
            "enabled": true,
            "tool": "cdxgen",
        },
        "integrity": map[string]interface{}{
            "enabled": true,
        },
    },
}

scanner := &sbom.SBOMScanner{}
result, err := scanner.Run(ctx, opts)
```

### Loading SBOM from Other Scanners

Other scanners can load the SBOM output:

```go
import "github.com/crashappsec/zero/pkg/scanners/sbom"

// Get SBOM path
sbomPath := sbom.GetSBOMPath(outputDir)

// Load SBOM data
sbomData, err := sbom.LoadSBOM(sbomPath)
if err != nil {
    // Handle error
}

// Access components
for _, component := range sbomData.Components {
    fmt.Printf("%s@%s (%s)\n", component.Name, component.Version, component.Ecosystem)
}
```

## Output Format

### Summary (sbom.json)

```json
{
  "scanner": "sbom",
  "version": "1.0.0",
  "summary": {
    "generation": {
      "tool": "cdxgen",
      "spec_version": "1.5",
      "total_components": 245,
      "by_type": {
        "library": 240,
        "framework": 5
      },
      "by_ecosystem": {
        "npm": 200,
        "golang": 45
      },
      "has_dependencies": true,
      "sbom_path": ".zero/repos/project/analysis/sbom.cdx.json"
    },
    "integrity": {
      "is_complete": true,
      "lockfiles_found": 2,
      "missing_packages": 0,
      "extra_packages": 0
    }
  },
  "findings": {
    "generation": {
      "components": [...],
      "dependencies": [...],
      "metadata": {...}
    },
    "integrity": {
      "lockfile_comparisons": [
        {
          "lockfile": "package-lock.json",
          "ecosystem": "npm",
          "in_sbom": 200,
          "in_lockfile": 200,
          "matched": 200,
          "missing": 0,
          "extra": 0
        }
      ]
    }
  }
}
```

### Full SBOM (sbom.cdx.json)

Standard CycloneDX 1.5 format - see [CycloneDX Specification](https://cyclonedx.org/specification/overview/).

## Prerequisites

One of the following tools must be installed:

| Tool | Install Command |
|------|-----------------|
| cdxgen | `npm install -g @cyclonedx/cdxgen` |
| syft | `brew install syft` or `curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh \| sh` |

## Related Scanners

- **packages**: Depends on SBOM output for vulnerability scanning, health checks, and license analysis
- **ai**: Uses SBOM component data for ML-BOM correlation

## See Also

- [Packages Scanner](packages.md) - Dependency analysis using SBOM data
- [Scanner Architecture](../architecture/scanners.md) - How scanners work together
- [CycloneDX Specification](https://cyclonedx.org/specification/overview/) - SBOM format details
