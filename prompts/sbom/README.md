<!--
Copyright (c) 2024 Crash Override Inc
101 Fulton St, 416, New York 10038
SPDX-License-Identifier: GPL-3.0
-->

# SBOM Management Prompts

Comprehensive prompt templates for Software Bill of Materials (SBOM) analysis, operations, security, and compliance using the SBOM Analyzer skill.

## Directory Structure

```
prompts/sbom/
├── security/           # Security-focused SBOM analysis
│   ├── vulnerability-scan.md
│   └── supply-chain-security.md
├── operations/         # SBOM operational tasks
│   ├── dependency-analysis.md
│   ├── format-conversion.md
│   └── version-upgrade.md
└── compliance/         # Compliance and licensing
    └── license-compliance.md
```

## Categories

### 🔒 Security
Security analysis and vulnerability management for SBOMs.

- **[vulnerability-scan.md](security/vulnerability-scan.md)** - Comprehensive vulnerability scanning with CISA KEV prioritization
- **[supply-chain-security.md](security/supply-chain-security.md)** - Supply chain security posture assessment, SLSA compliance, provenance

### ⚙️ Operations
SBOM transformation, conversion, and management operations.

- **[dependency-analysis.md](operations/dependency-analysis.md)** - Dependency graph analysis, outdated packages, circular dependencies
- **[format-conversion.md](operations/format-conversion.md)** - Convert between CycloneDX and SPDX formats
- **[version-upgrade.md](operations/version-upgrade.md)** - Upgrade SBOMs to latest specification versions

### ✅ Compliance
License compliance and regulatory requirements.

- **[license-compliance.md](compliance/license-compliance.md)** - License analysis, conflict detection, copyleft obligations

## Quick Start

1. **Load the SBOM Analyzer Skill** in Crash Override
2. **Choose a category** based on your needs (Security, Operations, Compliance)
3. **Select a prompt template** for your specific task
4. **Copy and customize** the prompt with your SBOM
5. **Execute and review** the analysis

## Common Workflows

### Security Assessment
```
1. vulnerability-scan.md       → Identify vulnerabilities
2. supply-chain-security.md    → Assess supply chain risks
3. dependency-analysis.md      → Understand impact through dependencies
```

### SBOM Modernization
```
1. version-upgrade.md          → Upgrade to latest version
2. format-conversion.md        → Convert to preferred format (if needed)
3. vulnerability-scan.md       → Validate security posture
```

### Compliance Review
```
1. license-compliance.md       → Check license obligations
2. supply-chain-security.md    → Verify provenance and attestations
3. dependency-analysis.md      → Review dependency structure
```

### Procurement Evaluation
```
1. vulnerability-scan.md       → Security risk assessment
2. license-compliance.md       → Legal compliance check
3. supply-chain-security.md    → Trust and provenance verification
```

## Use Cases by Role

### Security Teams
- **Daily**: vulnerability-scan, supply-chain-security
- **Weekly**: dependency-analysis for risk trending
- **As Needed**: format-conversion for tool integration

### Compliance Officers
- **Pre-Release**: license-compliance
- **Audits**: supply-chain-security (provenance)
- **Vendor Review**: All compliance prompts

### DevOps/SRE
- **CI/CD Integration**: vulnerability-scan, version-upgrade
- **Operations**: format-conversion, dependency-analysis
- **Monitoring**: supply-chain-security (scorecard tracking)

### Engineering Managers
- **Sprint Planning**: dependency-analysis (tech debt)
- **Release Readiness**: vulnerability-scan, license-compliance
- **Vendor Management**: supply-chain-security

## Best Practices

### Regular Scanning
- **Continuous**: Scan on every build (vulnerability-scan)
- **Daily**: Check CISA KEV for new exploits
- **Weekly**: Dependency analysis for outdated packages
- **Monthly**: License compliance review

### SBOM Maintenance
- **Version Control**: Track SBOM changes with Git
- **Format Standardization**: Use format-conversion for consistency
- **Version Currency**: Apply version-upgrade regularly
- **Enrichment**: Continuously enhance with metadata

### Integration
- **CI/CD Pipelines**: Automate vulnerability scanning
- **Security Tools**: Convert formats for tool compatibility
- **Compliance Systems**: Extract license data programmatically
- **Dashboards**: Track metrics over time

## Prerequisites

- SBOM Analyzer skill loaded in Crash Override
- SBOM document (CycloneDX or SPDX format)
- Internet access for API queries (OSV.dev, deps.dev, CISA KEV)

## Data Sources

All prompts leverage:
- **OSV.dev**: Vulnerability database
- **deps.dev**: Dependency intelligence
- **CISA KEV**: Known exploited vulnerabilities
- **OpenSSF Scorecard**: Project security metrics
- **SPDX License List**: License information

## Output Formats

Prompts can generate:
- Markdown reports
- JSON/CSV data exports
- Mermaid dependency graphs
- Executive summaries
- Technical deep-dives

## Related Resources

### Skills
- [SBOM Analyzer Skill](../../skills/sbom-analyzer/) - Complete documentation
- [Examples](../../skills/sbom-analyzer/examples/) - Sample SBOMs and reports

### External Documentation
- [CycloneDX Specification](https://cyclonedx.org/specification/overview/)
- [SPDX Specification](https://spdx.dev/specifications/)
- [OSV.dev](https://google.github.io/osv.dev/)
- [deps.dev API](https://docs.deps.dev/api/v3alpha/)
- [CISA KEV](https://www.cisa.gov/known-exploited-vulnerabilities-catalog)

## Contributing

Have a useful SBOM prompt? Please contribute!

1. Choose the appropriate category (security/operations/compliance)
2. Follow the template structure in existing prompts
3. Include purpose, usage, examples, and variations
4. Test thoroughly before submitting
5. Submit a pull request

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for guidelines.

## Support

For questions or issues:
- Review [SBOM Analyzer documentation](../../skills/sbom-analyzer/README.md)
- Check existing [discussions](https://github.com/crashappsec/skills-and-prompts/discussions)
- Open an [issue](https://github.com/crashappsec/skills-and-prompts/issues)
- Contact: mark@crashoverride.com

---

**Empowering secure software supply chains with comprehensive SBOM management**
