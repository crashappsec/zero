<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Gibson Powers Configuration

Global configuration files for Gibson Powers analysers and tools.

## SBOM Configuration

### Quick Start

```bash
# 1. Copy example configuration
cp config/sbom-config.example.json config/sbom-config.json

# 2. Edit configuration
# Customize settings based on your needs

# 3. Use with analysers
./utils/supply-chain/package-health-analysis/package-health-analyser.sh --repo owner/repo
```

### Configuration Files

- **`sbom-config.json`** - Main SBOM configuration (create from example)
- **`sbom-config.example.json`** - Example configuration with documentation

### Available Presets

The SBOM configuration includes several presets for common use cases:

#### 1. Production (Default)
- **Use case**: Production releases, customer delivery
- **Includes**: Only production runtime dependencies
- **Format**: CycloneDX JSON
- **Lock files**: Enabled

```json
{
  "preset": "production"
}
```

#### 2. Compliance
- **Use case**: Regulatory compliance, CISA NTIA requirements
- **Includes**: Production dependencies with full metadata
- **Format**: SPDX JSON
- **Validation**: CISA NTIA minimum elements

```json
{
  "preset": "compliance"
}
```

#### 3. Development
- **Use case**: Internal development, full dependency tree
- **Includes**: All dependencies (dev, test, runtime)
- **Format**: CycloneDX JSON

```json
{
  "preset": "development"
}
```

#### 4. Security Audit
- **Use case**: Security analysis, vulnerability scanning
- **Includes**: All dependencies with security metadata
- **Format**: CycloneDX JSON with CPEs
- **Integrations**: OSV vulnerability checking

```json
{
  "preset": "security_audit"
}
```

## Configuration Structure

### Core Settings

```json
{
  "sbom": {
    "format": "cyclonedx-json",           // Output format
    "use_lock_files": true,               // Use lock files for accuracy
    "include_dev_deps": false,            // Include dev dependencies
    "include_test_deps": false,           // Include test dependencies
    "scan_node_modules": false,           // Scan node_modules directory
    "resolve_versions": true,             // Resolve version ranges
    "output_dir": "sbom-output"          // Output directory
  }
}
```

### Supported Formats

- **CycloneDX**: `cyclonedx-json`, `cyclonedx-xml`
- **SPDX**: `spdx-json`, `spdx-tag-value`
- **Syft**: `syft-json`
- **Table**: `table` (human-readable)

### Package Managers

Automatic detection and lock file support for:

| Ecosystem | Package Manager | Lock File | Status |
|-----------|----------------|-----------|--------|
| JavaScript | npm | package-lock.json | ✅ Full |
| JavaScript | yarn | yarn.lock | ✅ Full |
| JavaScript | pnpm | pnpm-lock.yaml | ✅ Full |
| JavaScript | bun | bun.lockb | ✅ Full |
| Python | pip | requirements.txt | ⚠️ Basic |
| Python | poetry | poetry.lock | ✅ Full |
| Python | pipenv | Pipfile.lock | ✅ Full |
| Rust | cargo | Cargo.lock | ✅ Full |
| Go | go | go.sum | ✅ Full |
| Ruby | bundler | Gemfile.lock | ✅ Full |
| Java | maven | pom.xml | ⚠️ Basic |
| Java | gradle | gradle.lockfile | ✅ Full |
| PHP | composer | composer.lock | ✅ Full |

### Lock File vs Manifest

**Why use lock files?**

Lock files provide:
- ✅ **Exact versions** - No version range ambiguity
- ✅ **Reproducibility** - Same dependencies every time
- ✅ **Transitive deps** - Complete dependency tree
- ✅ **Integrity hashes** - Checksum verification

**When to use manifests:**
- Package manager doesn't support lock files
- Quick analysis where exact versions aren't critical
- Lock file is missing or corrupted

### Dependency Types

Configure which dependencies to include:

```json
{
  "dependencies": {
    "production": {
      "include": true,
      "transitive": true
    },
    "development": {
      "include": false,
      "transitive": false
    },
    "test": {
      "include": false,
      "transitive": false
    },
    "optional": {
      "include": true
    },
    "peer": {
      "include": true
    }
  }
}
```

### Metadata Options

Control what information is included in the SBOM:

```json
{
  "metadata": {
    "include_licenses": true,
    "include_checksums": true,
    "include_purls": true,
    "include_cpes": false,

    "checksum_algorithms": ["sha256", "sha1"],

    "component_metadata": {
      "supplier": true,
      "author": true,
      "description": true,
      "homepage": true,
      "repository": true,
      "download_location": true
    }
  }
}
```

### Validation

Enable SBOM validation and compliance checks:

```json
{
  "validation": {
    "enabled": true,
    "strict_mode": false,

    "checks": {
      "minimum_elements": true,
      "cisa_ntia_compliance": true,
      "license_presence": true,
      "checksum_presence": true,
      "purl_format": true,
      "lock_file_consistency": true
    }
  }
}
```

### Performance

Optimize SBOM generation performance:

```json
{
  "performance": {
    "cache": {
      "enabled": true,
      "directory": ".sbom-cache",
      "ttl_hours": 24
    },

    "parallel": {
      "enabled": true,
      "max_workers": 4
    },

    "timeout": {
      "generation": 600,
      "api_calls": 30
    }
  }
}
```

## Common Scenarios

### Scenario 1: Production Release SBOM

```json
{
  "sbom": {
    "format": "cyclonedx-json",
    "use_lock_files": true,
    "include_dev_deps": false,
    "include_test_deps": false,
    "metadata": {
      "include_licenses": true,
      "include_checksums": true
    }
  }
}
```

### Scenario 2: Security Audit

```json
{
  "sbom": {
    "format": "cyclonedx-json",
    "use_lock_files": true,
    "include_dev_deps": true,
    "include_test_deps": true,
    "metadata": {
      "include_licenses": true,
      "include_checksums": true,
      "include_cpes": true
    },
    "integrations": {
      "osv": {
        "enabled": true,
        "vulnerability_check": true
      }
    }
  }
}
```

### Scenario 3: Compliance (CISA NTIA)

```json
{
  "sbom": {
    "format": "spdx-json",
    "use_lock_files": true,
    "include_dev_deps": false,
    "validation": {
      "enabled": true,
      "checks": {
        "cisa_ntia_compliance": true
      }
    }
  }
}
```

### Scenario 4: Monorepo

```json
{
  "sbom": {
    "format": "cyclonedx-json",
    "use_lock_files": true,
    "scanning": {
      "include_paths": [
        "packages/*",
        "apps/*"
      ]
    },
    "lock_files": {
      "package_managers": {
        "npm": {
          "include_workspaces": true
        }
      }
    }
  }
}
```

## Environment Variables

Override configuration with environment variables:

```bash
# Format
export SBOM_FORMAT="spdx-json"

# Lock file usage
export SBOM_USE_LOCK_FILES=true

# Dependencies
export SBOM_INCLUDE_DEV_DEPS=false
export SBOM_INCLUDE_TEST_DEPS=false

# Scanning
export SBOM_SCAN_NODE_MODULES=false

# Output
export SBOM_OUTPUT_DIR="custom-sbom-output"
```

## Troubleshooting

### Issue: SBOM generation fails

**Solution**: Check that syft is installed and lock files exist

```bash
# Install syft
brew install syft

# Check for lock files
ls -la package-lock.json yarn.lock pnpm-lock.yaml
```

### Issue: Missing dependencies in SBOM

**Solution**: Ensure lock file exists and `use_lock_files` is enabled

```bash
# Generate lock file if missing
npm install
# or
yarn install
```

### Issue: SBOM too large

**Solution**: Exclude dev and test dependencies

```json
{
  "sbom": {
    "include_dev_deps": false,
    "include_test_deps": false
  }
}
```

### Issue: Slow generation

**Solution**: Enable caching and parallel processing

```json
{
  "performance": {
    "cache": {
      "enabled": true
    },
    "parallel": {
      "enabled": true,
      "max_workers": 4
    }
  }
}
```

## Best Practices

1. **Always use lock files** for production SBOMs
2. **Exclude dev/test deps** for release artifacts
3. **Include licenses** for compliance
4. **Enable validation** for quality assurance
5. **Sign SBOMs** for authenticity
6. **Cache results** for performance
7. **Version control** SBOM configurations
8. **Document** custom configurations

## References

- [SBOM Generation Best Practices](../rag/supply-chain/sbom-generation-best-practices.md)
- [Package Manager Specifications](../rag/supply-chain/package-manager-specifications.md)
- [CycloneDX Specification](https://cyclonedx.org/specification/overview/)
- [SPDX Specification](https://spdx.dev/specifications/)
- [CISA NTIA Minimum Elements](https://www.ntia.gov/files/ntia/publications/sbom_minimum_elements_report.pdf)

## Support

For issues or questions:
- [GitHub Issues](https://github.com/crashappsec/gibson-powers/issues)
- [Documentation](../README.md)
