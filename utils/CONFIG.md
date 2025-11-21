# Configuration System

The Skills and Prompts utilities use a hierarchical configuration system that allows for global settings with module-specific overrides.

## Overview

```
utils/
├── config.json                    # Global configuration (user-created)
├── config.example.json            # Global configuration template
├── lib/
│   └── config-loader.sh          # Configuration loading library
└── <module>/
    ├── config.json               # Module-specific config (optional, user-created)
    └── config.example.json       # Module-specific config template
```

## Configuration Hierarchy

Configuration is loaded in the following order (later overrides earlier):

1. **Global Config** (`utils/config.json`)
   - Provides base settings for all modules
   - GitHub authentication, organizations, repositories
   - Default output formats and directories
   - Tool configurations (cosign, rekor, osv-scanner, syft)

2. **Module Config** (`utils/<module>/config.json`)
   - Module-specific overrides
   - Can override GitHub settings, output locations, etc.
   - Can be disabled via global `ignore_module_configs` flag

3. **Command-Line Arguments**
   - Highest priority
   - `--org`, `--repo`, `--output`, etc.
   - Always override config file settings

## Quick Start

### Initial Setup

1. Create your global config from the template:
```bash
cd utils
cp config.example.json config.json
```

2. Edit `config.json` with your settings:
```json
{
  "github": {
    "pat": "ghp_yourtoken",
    "organizations": ["myorg"],
    "repositories": ["owner/repo1", "owner/repo2"]
  },
  "modules": {
    "supply-chain": {
      "default_modules": ["vulnerability", "provenance"]
    }
  }
}
```

3. Run any utility - it will automatically use your config:
```bash
./supply-chain/supply-chain-scanner.sh --vulnerability
```

### Interactive Setup

Most utilities provide an interactive setup wizard:

```bash
./supply-chain/supply-chain-scanner.sh --setup
```

This will guide you through:
- GitHub authentication
- Organization selection
- Repository configuration
- Module defaults

## Global Configuration

### Location
`utils/config.json` (created from `utils/config.example.json`)

### Structure

```json
{
  "version": "1.0",
  "config_behavior": {
    "ignore_module_configs": false
  },
  "github": {
    "pat": "",
    "organizations": [],
    "repositories": []
  },
  "output": {
    "default_dir": "./reports",
    "formats": ["table", "json", "markdown"],
    "default_format": "table"
  },
  "modules": {
    "supply-chain": {
      "enabled": true,
      "default_modules": ["vulnerability", "provenance"],
      "output_dir": "./supply-chain-reports"
    },
    "dora-metrics": {
      "enabled": true,
      "output_dir": "./dora-reports"
    },
    "code-ownership": {
      "enabled": true,
      "output_dir": "./ownership-reports"
    }
  },
  "tools": {
    "cosign": {
      "enabled": true,
      "keyless": true
    },
    "rekor": {
      "enabled": true,
      "url": "https://rekor.sigstore.dev"
    }
  }
}
```

### Key Fields

#### `config_behavior.ignore_module_configs`
- **Type**: Boolean
- **Default**: `false`
- **Purpose**: When `true`, only uses global config and ignores all module-specific configs
- **Use Case**: Enforce organization-wide settings

#### `github.pat`
- **Type**: String
- **Purpose**: GitHub Personal Access Token for API access
- **Optional**: Can use `gh` CLI authentication instead
- **Scope Required**: `repo`, `read:org`

#### `github.organizations`
- **Type**: Array of strings
- **Purpose**: GitHub organizations to scan
- **Format**: `["org1", "org2"]`
- **Note**: Tools will expand to all repos in the org

#### `github.repositories`
- **Type**: Array of strings
- **Purpose**: Specific repositories to scan
- **Format**: `["owner/repo1", "owner/repo2"]`

#### `modules.<module>.default_modules`
- **Type**: Array of strings
- **Purpose**: Default analysis modules to run if none specified on CLI
- **Example**: `["vulnerability", "provenance"]`

## Module-Specific Configuration

### Location
`utils/<module>/config.json` (created from `utils/<module>/config.example.json`)

### Purpose
Override global settings for specific modules without affecting others.

### Example: Supply Chain Module

```json
{
  "version": "1.0",
  "description": "Supply chain module-specific configuration",
  "github": {
    "repositories": ["security-team/critical-repo"]
  },
  "modules": {
    "supply_chain": {
      "enabled": true,
      "default_modules": ["vulnerability", "provenance"],
      "vulnerability": {
        "prioritize": true,
        "min_cvss": 7.0,
        "check_kev": true,
        "taint_analysis": false
      },
      "provenance": {
        "verify_signatures": true,
        "min_slsa_level": 2,
        "strict_mode": true,
        "trusted_builders": [
          "https://github.com/actions/runner"
        ]
      }
    }
  }
}
```

### Merge Behavior

When both global and module configs exist:
- Module config values **override** global config values
- Arrays are **replaced** (not merged)
- Objects are **deep merged**

Example:
```json
// Global config
{
  "github": {
    "organizations": ["org1", "org2"],
    "repositories": ["owner/repo1"]
  }
}

// Module config
{
  "github": {
    "repositories": ["owner/repo2"]
  }
}

// Final merged config
{
  "github": {
    "organizations": ["org1", "org2"],
    "repositories": ["owner/repo2"]  // Module override
  }
}
```

## Using the Config Loader Library

### In Shell Scripts

```bash
#!/bin/bash

# Source the config loader
UTILS_ROOT="$(dirname "$SCRIPT_DIR")"
source "$UTILS_ROOT/lib/config-loader.sh"

# Load config with optional module-specific override
load_config "supply-chain" "$MODULE_CONFIG_FILE"

# Access config values
organizations=$(get_organizations)
repositories=$(get_repositories)
default_modules=$(get_default_modules)

# Get specific config paths
cvss_threshold=$(get_module_config 'vulnerability.min_cvss')

# Check if module is enabled
if is_module_enabled; then
    echo "Module is enabled"
fi
```

### Available Functions

#### `load_config <module_name> [module_config_path]`
- Loads global config and optionally merges module-specific config
- Exports `CONFIG_JSON` and `CONFIG_MODULE` environment variables
- Returns 0 on success, 1 on error

#### `get_config <json_path>`
- Retrieves value from loaded config using jq path syntax
- Example: `get_config '.github.pat'`

#### `get_config_default <json_path> <default_value>`
- Like `get_config` but returns default if value is null/empty

#### `get_module_config <key>`
- Gets module-specific config value
- Automatically uses current module name from `load_config`
- Example: `get_module_config 'default_modules[]'`

#### `is_module_enabled`
- Returns 0 (true) if module is enabled, 1 (false) otherwise
- Checks `.modules.<module>.enabled` in config

#### `get_organizations`
- Returns list of GitHub organizations from config
- One per line

#### `get_repositories`
- Returns list of GitHub repositories from config
- Format: `owner/repo`, one per line

#### `get_default_modules`
- Returns list of default analysis modules for current module
- One per line

## Configuration Precedence Examples

### Example 1: Default Behavior
```bash
# No CLI args, uses config defaults
./supply-chain-scanner.sh
# Runs: vulnerability + provenance (from config.default_modules)
# Targets: All repos/orgs from config
```

### Example 2: CLI Override
```bash
# CLI overrides config
./supply-chain-scanner.sh --vulnerability --repo owner/specific-repo
# Runs: Only vulnerability (overrides default_modules)
# Targets: owner/specific-repo (overrides config repos/orgs)
```

### Example 3: Module Override
```bash
# Global config: organizations: ["org1", "org2"]
# Module config: organizations: ["org3"]

./supply-chain-scanner.sh
# Uses: org3 (module overrides global)
```

### Example 4: Ignore Module Configs
```bash
# Global config: ignore_module_configs: true
# Module config: <any settings>

./supply-chain-scanner.sh
# Uses: Only global config (module config ignored)
```

## Security Considerations

### GitHub Personal Access Token (PAT)

1. **File Permissions**: Config files containing PATs should be protected:
```bash
chmod 600 utils/config.json
```

2. **Git Ignore**: Config files are automatically ignored:
```bash
# .gitignore already contains:
**/config.json
!**/config.example.json
```

3. **Alternative**: Use `gh` CLI authentication instead of PAT in config:
```bash
gh auth login
# Leave config.github.pat empty
```

### Environment Variables

For CI/CD or automation, override config via environment:
```bash
export GITHUB_TOKEN="ghp_token"
export CONFIG_ORG="myorg"
./supply-chain-scanner.sh --org "$CONFIG_ORG"
```

## Module-Specific Settings

### Supply Chain

```json
{
  "modules": {
    "supply-chain": {
      "default_modules": ["vulnerability", "provenance"],
      "output_dir": "./supply-chain-reports",
      "vulnerability": {
        "prioritize": true,
        "min_cvss": 0,
        "check_kev": true,
        "taint_analysis": false,
        "formats": ["table", "json", "markdown"],
        "strict_mode": false
      },
      "provenance": {
        "verify_signatures": false,
        "min_slsa_level": 0,
        "strict_mode": false,
        "trusted_builders": [
          "https://github.com/actions/runner",
          "https://cloudbuild.googleapis.com"
        ],
        "ecosystems": ["npm", "pypi", "go", "maven", "docker"]
      }
    }
  }
}
```

#### Vulnerability Settings

- `prioritize`: Enable intelligent prioritization (KEV, CVSS)
- `min_cvss`: Minimum CVSS score to report (0-10)
- `check_kev`: Check CISA Known Exploited Vulnerabilities catalog
- `taint_analysis`: Enable osv-scanner call analysis (Go projects)
- `formats`: Output formats to generate
- `strict_mode`: Fail on any vulnerability found

#### Provenance Settings

- `verify_signatures`: Enable cryptographic signature verification
- `min_slsa_level`: Minimum SLSA level required (0-4)
- `strict_mode`: Fail if provenance doesn't meet min_slsa_level
- `trusted_builders`: List of trusted build platform URLs
- `ecosystems`: Package ecosystems to analyze

### DORA Metrics

```json
{
  "modules": {
    "dora-metrics": {
      "enabled": true,
      "output_dir": "./dora-reports"
    }
  }
}
```

### Code Ownership

```json
{
  "modules": {
    "code-ownership": {
      "enabled": true,
      "output_dir": "./ownership-reports"
    }
  }
}
```

## Troubleshooting

### Config Not Found

**Error**: "No config file found"

**Solution**:
```bash
# Create from template
cp utils/config.example.json utils/config.json

# Or use interactive setup
./supply-chain/supply-chain-scanner.sh --setup
```

### Module Config Ignored

**Check**: `config_behavior.ignore_module_configs` in global config

```bash
# Verify setting
jq '.config_behavior.ignore_module_configs' utils/config.json
```

### Organizations/Repos Not Found

**Verify**:
```bash
# Check config
jq '.github.organizations[]' utils/config.json
jq '.github.repositories[]' utils/config.json

# Test GitHub auth
gh auth status
```

### jq Parse Errors

**Cause**: Invalid JSON in config file

**Fix**:
```bash
# Validate JSON
jq . utils/config.json

# If error, compare with example
diff utils/config.json utils/config.example.json
```

## Migration from Old Config

If you have old module-specific configs (pre-hierarchical system):

```bash
# Old: utils/supply-chain/config.json had everything
# New: Split between global and module-specific

# 1. Create global config
cp utils/config.example.json utils/config.json

# 2. Move GitHub settings to global
jq '{github: .github}' old-config.json |
  jq -s '.[0] * .[1]' utils/config.json - > new-config.json
mv new-config.json utils/config.json

# 3. Keep module-specific settings in module config
jq '{modules: .modules}' old-config.json > utils/supply-chain/config.json
```

## Best Practices

1. **Version Control**
   - ✅ Commit: `config.example.json` files
   - ❌ Never commit: `config.json` files with PATs

2. **Organization Standards**
   - Use global config for org-wide settings
   - Enable `ignore_module_configs` for standardization
   - Distribute global config via secure channels

3. **Team Workflows**
   - Share `config.example.json` with sensible defaults
   - Document custom settings in team wiki
   - Use `--setup` for onboarding new team members

4. **CI/CD Integration**
   - Use CLI args instead of config files
   - Pass secrets via CI environment variables
   - Example: `--repo $REPO --output $OUTPUT_DIR`

5. **Testing**
   - Test with both global and module configs
   - Verify `ignore_module_configs` behavior
   - Check CLI argument override precedence

## See Also

- [Supply Chain README](./supply-chain/README.md)
- [DORA Metrics README](./dora-metrics/README.md)
- [Code Ownership README](./code-ownership/README.md)
- [Global Changelog](../CHANGELOG.md)
