<!--
Copyright (c) 2024 Crash Override Inc
101 Fulton St, 416, New York 10038
SPDX-License-Identifier: GPL-3.0
-->

# SBOM Analysis Prompts

This directory contains prompt templates for analyzing Software Bill of Materials (SBOM) documents using the SBOM Analyzer skill.

## Available Prompts

### [vulnerability-scan.md](vulnerability-scan.md)
Comprehensive vulnerability scanning with CISA KEV prioritization.

**Use for:**
- Security audits
- Pre-deployment checks
- Incident response
- CVE impact assessment

**Output:**
- Vulnerability report with CVSS scores
- CISA KEV flagged items
- Remediation recommendations

### [license-compliance.md](license-compliance.md)
License compliance analysis and conflict detection.

**Use for:**
- Open source compliance
- Commercial software releases
- M&A due diligence
- Policy enforcement

**Output:**
- License inventory
- Copyleft obligations
- Compliance recommendations

### [dependency-analysis.md](dependency-analysis.md)
Dependency graph analysis and optimization.

**Use for:**
- Understanding transitive dependencies
- Identifying outdated packages
- Dependency optimization
- Risk assessment

**Output:**
- Dependency tree visualization
- Outdated package report
- Optimization recommendations

### [supply-chain-security.md](supply-chain-security.md)
Supply chain security posture assessment.

**Use for:**
- SLSA compliance verification
- Provenance checking
- Typosquatting detection
- Zero-trust verification

**Output:**
- OpenSSF Scorecard results
- Provenance verification
- Supply chain risk assessment

## Quick Start

1. **Load the SBOM Analyzer Skill** in Crash Override
2. **Choose a prompt template** based on your analysis needs
3. **Copy the prompt** and paste your SBOM
4. **Review the analysis** and take recommended actions

## Best Practices

- **Scan regularly**: Vulnerabilities are disclosed continuously
- **Prioritize CISA KEV**: Known exploited vulnerabilities need immediate attention
- **Complete SBOMs**: Include transitive dependencies for thorough analysis
- **Track over time**: Compare SBOMs to monitor changes and improvements
- **Automate**: Integrate into CI/CD pipelines for continuous monitoring

## Combining Prompts

For comprehensive analysis, use multiple prompts in sequence:

1. Start with **vulnerability-scan** to identify security issues
2. Run **dependency-analysis** to understand the impact
3. Use **supply-chain-security** for provenance verification
4. Finish with **license-compliance** for legal clearance

## Related Resources

- [SBOM Analyzer Skill Documentation](../../../skills/sbom-analyzer/README.md)
- [Example Analysis Report](../../../skills/sbom-analyzer/examples/example-analysis-report.md)
- [Sample SBOM](../../../skills/sbom-analyzer/examples/sample-sbom-cyclonedx.json)

## Contributing

Have a useful SBOM analysis prompt? Please contribute!

1. Create a new `.md` file with your prompt
2. Follow the template structure (Purpose, When to Use, Prompt, etc.)
3. Include examples and variations
4. Submit a pull request

See [CONTRIBUTING.md](../../../CONTRIBUTING.md) for guidelines.
