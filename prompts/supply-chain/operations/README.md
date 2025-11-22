<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# SBOM Operations Prompts

Prompts for operational SBOM management including format conversion, version upgrades, and dependency analysis.

## Available Prompts

### [dependency-analysis.md](dependency-analysis.md)
Analyze dependency graphs, identify outdated packages, detect circular dependencies, and optimize dependency trees.

**Use for:**
- Understanding transitive dependencies
- Identifying technical debt
- Planning dependency updates
- Risk assessment based on dependency position

### [format-conversion.md](format-conversion.md)
Convert SBOMs between CycloneDX and SPDX formats while preserving data integrity.

**Use for:**
- Tool compatibility requirements
- Standardizing SBOM collections
- Vendor integration
- Format-specific analysis needs

### [version-upgrade.md](version-upgrade.md)
Upgrade SBOMs from older specification versions to current versions.

**Use for:**
- Modernizing legacy SBOMs
- Leveraging new specification features
- Ensuring tool compatibility
- Meeting updated compliance requirements

## Common Operations Workflows

### SBOM Modernization
```
1. version-upgrade → Upgrade to latest spec version
2. format-conversion → Convert to preferred format
3. dependency-analysis → Validate and optimize
```

### Tool Integration
```
1. format-conversion → Convert to required format
2. dependency-analysis → Verify completeness
```

### Dependency Management
```
1. dependency-analysis → Identify outdated packages
2. Apply updates to dependencies
3. dependency-analysis → Verify improvements
```

## Best Practices

- **Validate Before and After**: Always validate SBOMs before and after operations
- **Maintain Originals**: Keep original SBOMs for reference and audit
- **Document Changes**: Use annotations to track transformations
- **Test Thoroughly**: Verify converted/upgraded SBOMs with target tools
- **Automate Where Possible**: Integrate into CI/CD pipelines

## Related
- [Security Prompts](../security/) - Vulnerability and supply chain security
- [Compliance Prompts](../compliance/) - License and regulatory compliance
