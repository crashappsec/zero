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

**Supported Ecosystems and Lockfiles:**

| Ecosystem | Manager | Manifest | Lock File | Native Tool |
|-----------|---------|----------|-----------|-------------|
| JavaScript | npm | `package.json` | `package-lock.json` | `npm sbom` |
| JavaScript | yarn | `package.json` | `yarn.lock` | cyclonedx-yarn |
| JavaScript | pnpm | `package.json` | `pnpm-lock.yaml` | cyclonedx-pnpm |
| Python | pip | `requirements.txt` | (pip freeze) | cyclonedx-py |
| Python | poetry | `pyproject.toml` | `poetry.lock` | cyclonedx-py |
| Python | uv | `pyproject.toml` | `uv.lock` | cyclonedx-py |
| Rust | cargo | `Cargo.toml` | `Cargo.lock` | cargo-cyclonedx |
| Go | go mod | `go.mod` | `go.sum` | cyclonedx-gomod |
| Java | maven | `pom.xml` | (effective-pom) | cyclonedx-maven |
| Java | gradle | `build.gradle` | `gradle.lockfile` | cyclonedx-gradle |
| Ruby | bundler | `Gemfile` | `Gemfile.lock` | cyclonedx-ruby |
| PHP | composer | `composer.json` | `composer.lock` | cyclonedx-php |
| .NET | nuget | `*.csproj` | `packages.lock.json` | CycloneDX .NET |

See [Package Manager Patterns](/rag/supply-chain/package-managers/) for detailed ecosystem documentation.

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

## Configuration

The SBOM scanner uses `config/sbom.config.json` for detailed configuration of package manager support, dependency inclusion, and output format.

### SBOM Configuration File

```json
{
  "$schema": "./sbom.config.schema.json",
  "_version": "1.0.0",

  "output": {
    "format": "cyclonedx",
    "spec_version": "1.5",
    "output_format": "json"
  },

  "dependencies": {
    "include_dev": false,
    "include_test": false,
    "include_optional": true,
    "include_transitive": true
  },

  "metadata": {
    "include_licenses": true,
    "include_hashes": true,
    "include_purl": true
  },

  "package_managers": {
    "npm": {
      "use_native_sbom": true,
      "omit": ["dev"]
    },
    "poetry": {
      "groups": ["main"],
      "from_lock_file": true
    }
  }
}
```

See [sbom.config.schema.json](/config/sbom.config.schema.json) for the full schema.

### Package Manager-Specific Options

Each ecosystem has specific configuration options:

| Ecosystem | Key Options |
|-----------|-------------|
| npm/yarn/pnpm | `omit`, `workspaces`, `package_lock_only` |
| pip/poetry/uv | `groups`, `extras`, `from_lock_file` |
| cargo | `all_features`, `features`, `include_dev` |
| go | `include_test`, `include_std`, `from_binary` |
| maven/gradle | `include_compile`, `include_runtime`, `configurations` |
| bundler | `exclude_groups`, `include_groups` |
| composer | `include_dev`, `fetch_packagist_metadata` |
| nuget | `target_frameworks`, `use_lock_file` |

## Prerequisites

One of the following tools must be installed:

| Tool | Install Command |
|------|-----------------|
| cdxgen | `npm install -g @cyclonedx/cdxgen` |
| syft | `brew install syft` or `curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh \| sh` |

### Ecosystem-Specific Tools (Optional)

For more accurate SBOMs, install ecosystem-specific tools:

| Ecosystem | Tool | Install Command |
|-----------|------|-----------------|
| npm | npm sbom | Built into npm 9+ |
| Python | cyclonedx-py | `pip install cyclonedx-bom` |
| Rust | cargo-cyclonedx | `cargo install cargo-cyclonedx` |
| Go | cyclonedx-gomod | `go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest` |
| Java | cyclonedx-maven | Add plugin to pom.xml |
| Java | cyclonedx-gradle | Add plugin to build.gradle |
| Ruby | cyclonedx-ruby | `gem install cyclonedx-ruby` |
| PHP | cyclonedx-php | `composer global require cyclonedx/cyclonedx-php-composer` |
| .NET | CycloneDX | `dotnet tool install --global CycloneDX` |

## Related Scanners

- **packages**: Depends on SBOM output for vulnerability scanning, health checks, and license analysis
- **technology**: Uses SBOM component data for technology detection and ML-BOM correlation

## RAG Knowledge

Detailed package manager patterns are available in the RAG knowledge base:

```
rag/supply-chain/package-managers/
├── README.md                    # Overview and supported ecosystems
├── npm/patterns.md              # npm (JavaScript)
├── yarn/patterns.md             # Yarn Classic & Berry
├── pnpm/patterns.md             # pnpm (JavaScript)
├── pip/patterns.md              # pip (Python)
├── poetry/patterns.md           # Poetry (Python)
├── uv/patterns.md               # uv (Python, Rust-based)
├── cargo/patterns.md            # Cargo (Rust)
├── go/patterns.md               # Go Modules
├── maven/patterns.md            # Maven (Java)
├── gradle/patterns.md           # Gradle (Java/Kotlin)
├── bundler/patterns.md          # Bundler (Ruby)
├── composer/patterns.md         # Composer (PHP)
└── nuget/patterns.md            # NuGet (.NET)
```

Each pattern file includes:
- **TIER 1**: Manifest detection patterns
- **TIER 2**: Lock file structure and parsing
- **TIER 3**: Configuration extraction (registries, auth)
- **SBOM Generation**: Tool-specific commands
- **Best Practices**: Reproducible build recommendations
- **Troubleshooting**: Common issues and solutions

## See Also

- [Packages Scanner](packages.md) - Dependency analysis using SBOM data
- [Scanner Architecture](../architecture/scanners.md) - How scanners work together
- [CycloneDX Specification](https://cyclonedx.org/specification/overview/) - SBOM format details
- [SBOM Configuration Schema](/config/sbom.config.schema.json) - Full configuration options
- [SBOM Generation Best Practices](/rag/supply-chain/sbom-generation-best-practices.md) - Industry best practices
- [Package Manager Specifications](/rag/supply-chain/package-manager-specifications.md) - Detailed specifications
