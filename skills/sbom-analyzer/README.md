<!--
Copyright (c) 2024 Crash Override Inc
101 Fulton St, 416, New York 10038
SPDX-License-Identifier: GPL-3.0
-->

# SBOM/BOM Analyzer Skill

Expert analysis of Software Bill of Materials (SBOM) and Bill of Materials (BOM) documents using industry-standard formats, comprehensive vulnerability databases, and dependency intelligence tools.

## Purpose

This skill enables deep analysis of software supply chains by:
- **Vulnerability Detection**: Identifying security risks across all components
- **Dependency Analysis**: Understanding direct and transitive dependency relationships
- **License Compliance**: Ensuring license compatibility and identifying risks
- **Supply Chain Security**: Verifying component provenance and integrity
- **Risk Prioritization**: Focusing remediation on critical, exploited vulnerabilities

## Standards and Specifications

### SBOM Formats

#### CycloneDX 1.7 (ECMA-424)
- **Formats**: JSON, XML, Protocol Buffers
- **Components**: 12 structural elements including metadata, components, services, dependencies, vulnerabilities
- **Features**: Complete supply chain transparency, vulnerability tracking, provenance documentation

#### SPDX (Software Package Data Exchange)
- **Formats**: JSON, YAML, RDF, Tag-Value
- **Features**: License compliance, package relationships, file-level documentation
- **Standard**: ISO/IEC 5962:2021

## Data Sources

### OSV.dev (Open Source Vulnerabilities)
Comprehensive vulnerability database aggregating data from multiple sources:
- **Coverage**: All major package ecosystems (npm, PyPI, Maven, Go, Cargo, etc.)
- **API**: RESTful endpoints for queries, batch processing, version determination
- **Format**: OpenSSF Vulnerability Schema
- **Updates**: Continuously updated with new vulnerability disclosures

**Key Features:**
- Query by package name, version, or commit hash
- Batch processing for large SBOM analysis
- Affected version ranges with precision
- Ecosystem-specific vulnerability details

### deps.dev API (v3alpha)
Dependency intelligence and package metadata:
- **Ecosystems**: Go, RubyGems, npm, Cargo, Maven, PyPI, NuGet
- **Features**: Dependency graphs, security advisories, license data, OpenSSF Scorecard
- **Capabilities**: Typosquatting detection, SLSA provenance, deprecation tracking

**Key Endpoints:**
- Package and version information
- Resolved dependency graphs
- Security advisory integration
- Project metadata (GitHub, GitLab, Bitbucket)

### CISA Known Exploited Vulnerabilities (KEV)
Authoritative catalog of vulnerabilities with confirmed exploitation:
- **Purpose**: Prioritize remediation for actively exploited vulnerabilities
- **Updates**: Continuous additions as new exploits are discovered
- **Data**: CVE IDs, vendor info, exploitation details, remediation deadlines
- **Access**: JSON feed, CSV, web interface

## Prerequisites

- SBOM document in CycloneDX or SPDX format
- Internet access for API queries (OSV.dev, deps.dev, CISA KEV)
- Basic understanding of software dependencies and vulnerabilities

## Usage

### Load the Skill

In Crash Override, load the SBOM Analyzer skill to enable expert SBOM analysis capabilities.

### Basic Analysis

1. **Provide an SBOM**
   ```
   Please analyze this SBOM for vulnerabilities and risks:
   [paste SBOM content or attach file]
   ```

2. **Get Comprehensive Report**
   The skill will:
   - Parse and validate the SBOM structure
   - Extract all components and dependencies
   - Query vulnerability databases (OSV.dev, deps.dev)
   - Check CISA KEV for exploited vulnerabilities
   - Analyze licenses and compliance
   - Provide prioritized remediation recommendations

### Advanced Usage

#### Focused Vulnerability Scan
```
Scan this SBOM for vulnerabilities and prioritize based on CISA KEV listings.
Include CVSS scores and remediation guidance.
```

#### Dependency Analysis
```
Analyze the dependency graph in this SBOM. Identify:
- Transitive dependencies with vulnerabilities
- Outdated packages
- Circular dependencies
- Typosquatting risks
```

#### License Compliance Check
```
Review this SBOM for license compliance issues.
Identify any GPL/copyleft licenses and potential conflicts.
```

#### Supply Chain Security Assessment
```
Evaluate the supply chain security posture of this SBOM:
- Check for SLSA attestations
- Verify provenance information
- Review OpenSSF Scorecard metrics
- Identify unsigned components
```

#### Comparative Analysis
```
Compare these two SBOMs (before/after upgrade) and show:
- New vulnerabilities introduced
- Vulnerabilities resolved
- Dependency changes
- License modifications
```

## Analysis Workflow

The skill follows a systematic approach:

1. **Parse & Validate**
   - Identify SBOM format (CycloneDX, SPDX)
   - Validate structure and completeness
   - Extract metadata

2. **Inventory Components**
   - List all components with versions
   - Build dependency graph
   - Categorize component types

3. **Vulnerability Scanning**
   - Query OSV.dev for each component
   - Cross-reference deps.dev advisories
   - Check CISA KEV for known exploits
   - Deduplicate findings

4. **Risk Analysis**
   - Calculate severity scores
   - Assess exploitability
   - Evaluate impact based on dependency position
   - Factor in CISA KEV presence

5. **License Review**
   - Extract license declarations
   - Identify conflicts
   - Flag compliance issues

6. **Supply Chain Evaluation**
   - Review provenance and attestations
   - Check OpenSSF Scorecard
   - Assess component integrity

7. **Generate Recommendations**
   - Prioritize remediation actions
   - Suggest version upgrades
   - Recommend alternatives
   - Provide compliance guidance

## Output Formats

### Vulnerability Report
- Comprehensive table of vulnerabilities with CVE, CVSS, affected components
- Severity classification (Critical, High, Medium, Low)
- Exploitation status (CISA KEV flagging)
- Remediation recommendations with specific version upgrades

### Dependency Analysis
- Dependency tree visualization (Mermaid diagrams)
- Direct vs. transitive dependency breakdown
- Outdated package identification
- Circular dependency detection

### License Matrix
- Component license inventory
- License compatibility analysis
- Compliance risk assessment
- Missing license flagging

### Executive Summary
- High-level risk overview
- Critical findings count
- Top priorities for remediation
- Compliance status

## Examples

See the [examples/](examples/) directory for:
- Sample SBOM documents (CycloneDX and SPDX)
- Example vulnerability reports
- Dependency analysis outputs
- License compliance reviews
- Before/after comparison analyses

## Common Use Cases

### Security Audit
Comprehensive security assessment of a software project's dependencies:
```
Perform a complete security audit of this SBOM. Include:
1. All vulnerabilities with CVSS scores
2. CISA KEV matches requiring immediate action
3. Risk prioritization
4. Specific remediation steps
```

### Compliance Check
Ensure license compliance before deployment:
```
Check this SBOM for license compliance issues:
- Identify all GPL/AGPL components
- Flag license conflicts
- Verify SPDX expression validity
- List components with missing licenses
```

### Continuous Monitoring
Regular SBOM scanning as part of CI/CD:
```
Compare this SBOM against the baseline from last week:
- New vulnerabilities introduced
- Remediated issues
- Dependency drift
- License changes
```

### Procurement Review
Assess third-party software before acquisition:
```
Evaluate this vendor-provided SBOM:
- Overall security posture
- Maintenance quality (OpenSSF Scorecard)
- License obligations
- Supply chain risks
```

### Incident Response
Quickly assess impact of disclosed vulnerabilities:
```
Check if CVE-2024-XXXXX affects any components in this SBOM.
If yes, provide:
- Affected component versions
- Attack vector and severity
- Available patches
- Workarounds if no patch exists
```

## Best Practices

1. **Scan Regularly**: Vulnerabilities are disclosed continuously; rescan SBOMs frequently
2. **Prioritize CISA KEV**: Vulnerabilities with known exploitation require immediate attention
3. **Consider Context**: Not all vulnerabilities are exploitable in every deployment scenario
4. **Update Dependencies**: Keep components current to minimize vulnerability exposure
5. **Verify SBOM Completeness**: Ensure SBOMs include transitive dependencies
6. **Track Remediation**: Monitor vulnerability resolution over time
7. **Automate Where Possible**: Integrate SBOM analysis into CI/CD pipelines

## Limitations

- **API Rate Limits**: OSV.dev and deps.dev have rate limits for API queries
- **False Positives**: Vulnerability scanners may report issues not applicable to specific usage
- **Coverage Gaps**: Not all package ecosystems have complete vulnerability data
- **Timing**: There may be delays between vulnerability disclosure and database updates
- **Transitive Dependencies**: Some SBOM generators may not capture complete dependency trees

## Troubleshooting

### "Format not recognized"
Ensure the SBOM is valid CycloneDX or SPDX format. Validate against the official schemas.

### "No vulnerabilities found"
This could mean:
- Components are secure (verify with manual spot checks)
- SBOM is missing version information
- Package names don't match vulnerability database naming conventions

### "API query failed"
Check:
- Internet connectivity
- API service status (OSV.dev, deps.dev)
- Rate limiting (wait and retry)

### "Incomplete dependency graph"
The SBOM may only include direct dependencies. Regenerate with a tool that captures transitive dependencies.

## Resources

### Documentation
- [CycloneDX Specification](https://cyclonedx.org/specification/overview/)
- [SPDX Specification](https://spdx.dev/specifications/)
- [OSV.dev Documentation](https://google.github.io/osv.dev/)
- [deps.dev API v3alpha](https://docs.deps.dev/api/v3alpha/)
- [CISA KEV Catalog](https://www.cisa.gov/known-exploited-vulnerabilities-catalog)

### Tools
- **SBOM Generators**: Syft, CycloneDX CLI, SPDX tools
- **Vulnerability Scanners**: OSV-Scanner, Grype, Trivy
- **API Clients**: curl, Postman, custom scripts

### Related Skills
- [Certificate Analyzer](../certificate-analyzer/) - TLS/SSL certificate analysis
- [Chalk Build Analyzer](../chalk-build-analyzer/) - Build artifact analysis

## Contributing

Improvements to this skill are welcome! Consider contributing:
- Additional analysis capabilities
- New vulnerability data sources
- Enhanced reporting formats
- Example SBOMs and analyses
- Integration with other tools

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for guidelines.

## License

This skill is licensed under GPL-3.0. See [LICENSE](../../LICENSE) for details.

## Support

For questions, issues, or feature requests:
- Open an issue in the [GitHub repository](https://github.com/crashappsec/skills-and-prompts/issues)
- Review existing [discussions](https://github.com/crashappsec/skills-and-prompts/discussions)
- Contact: mark@crashoverride.com

---

**Made with expertise by the Crash Override community**
