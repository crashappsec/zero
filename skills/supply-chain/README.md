<!--
Copyright (c) 2024 Gibson Powers Contributors

SPDX-License-Identifier: GPL-3.0
-->

# Supply Chain Analyzer Skill

Comprehensive SBOM/BOM management including analysis, conversion, version upgrades, and supply chain security assessment using industry-standard formats, vulnerability databases, and security frameworks.

## Purpose

This skill enables complete SBOM lifecycle management and supply chain analysis:

### Analysis & Security
- **Vulnerability Detection**: Identifying security risks across all components using OSV.dev, deps.dev, and CISA KEV
- **Provenance Analysis**: Verifying build attestations and SLSA compliance using sigstore
- **Dependency Analysis**: Understanding direct and transitive dependency relationships
- **License Compliance**: Ensuring license compatibility and identifying risks
- **Supply Chain Security**: Verifying component provenance, integrity, and SLSA compliance
- **Risk Prioritization**: Focusing remediation on critical, exploited vulnerabilities

### SBOM Operations
- **Format Conversion**: Bidirectional conversion between CycloneDX and SPDX formats
- **Version Upgrades**: Modernize SBOMs to latest specification versions (CycloneDX 1.7, SPDX 2.3)
- **SBOM Transformation**: Merge, split, and enrich SBOMs
- **Validation**: Schema compliance and data integrity verification

### SLSA Framework
- **SLSA Assessment**: Evaluate SLSA levels (0-4) and provenance
- **Build Security**: Verify build platform integrity and isolation
- **Provenance Validation**: Check signatures and attestations
- **Compliance Recommendations**: Guidance for improving SLSA posture

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

## SLSA Framework Integration

### SLSA (Supply-chain Levels for Software Artifacts)
Comprehensive understanding of SLSA v1.0 for supply chain security:

**SLSA Levels:**
- **Level 0**: No guarantees (baseline)
- **Level 1**: Provenance exists (build documentation)
- **Level 2**: Signed provenance (tamper-resistant)
- **Level 3**: Hardened builds (isolated, ephemeral environments)
- **Level 4**: Two-party review + hermetic builds (aspirational)

**Assessment Capabilities:**
- Verify SLSA level compliance
- Validate provenance signatures
- Check build platform security
- Assess build isolation and ephemeral environments
- Provide recommendations for level advancement

**Integration:**
- CycloneDX: Formulation, Declarations, Attestations, Citations
- SPDX: External References, Annotations, Package Verification
- Tools: slsa-verifier, in-toto, SigStore, GUAC

## Prerequisites

- SBOM document in CycloneDX or SPDX format
- Internet access for API queries (OSV.dev, deps.dev, CISA KEV)
- Basic understanding of software dependencies and vulnerabilities

## Usage

### Load the Skill

In Crash Override, load the Supply Chain Analyzer skill to enable expert SBOM analysis capabilities.

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

#### Format Conversion
```
Convert this CycloneDX SBOM to SPDX 2.3 format:
- Preserve all component and dependency information
- Maintain license data
- Convert vulnerability information
- Provide conversion validation report

[Paste SBOM]
```

#### Version Upgrade
```
Upgrade this CycloneDX 1.2 SBOM to version 1.7:
- Add new required fields
- Modernize structure
- Validate against 1.7 schema
- Generate upgrade report showing changes

[Paste SBOM]
```

#### SLSA Compliance Assessment
```
Assess the SLSA compliance level of this SBOM:
- Identify SLSA level (0-4)
- Verify provenance and signatures
- Check build platform security
- Provide recommendations for next level

[Paste SBOM]
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

## Automation Scripts

The Supply Chain Analyzer includes command-line automation scripts for CI/CD integration and rapid analysis:

### vulnerability-analyzer.sh

Intelligent SBOM vulnerability scanning using osv-scanner with data-driven prioritization.

**Features:**
- Analyze SBOM files (JSON/XML)
- Scan Git repositories (auto-cloning)
- Scan local directories
- Taint analysis for Go projects (call graph/reachability)
- **Intelligent prioritization** (CISA KEV, CVSS severity, exploitability)
- Multiple output formats (table, JSON, markdown, SARIF)

**Usage:**
```bash
# Analyze an SBOM file
./vulnerability-analyzer.sh /path/to/sbom.json

# Analyze with intelligent prioritization (KEV + CVSS scoring)
./vulnerability-analyzer.sh --prioritize /path/to/sbom.json

# Analyze repository with taint analysis and prioritization
./vulnerability-analyzer.sh --taint-analysis --prioritize https://github.com/org/repo

# JSON output to file
./vulnerability-analyzer.sh --format json --output results.json ./my-project
```

**Prioritization Output:**
When using `--prioritize`, vulnerabilities are ranked by risk score:
- **CRITICAL**: In CISA KEV catalog (actively exploited)
- **HIGH**: CVSS 9-10 or KEV + High CVSS
- **MEDIUM**: CVSS 7-8.9
- **LOW**: CVSS < 7

Includes summary statistics: total vulnerabilities, severity breakdown, KEV matches.

**Requirements:**
- osv-scanner: `go install github.com/google/osv-scanner/cmd/osv-scanner@latest`
- syft (for SBOM generation): `brew install syft`
- jq: `brew install jq`

**Note:** For repositories without existing SBOMs, the scripts will automatically generate one using syft (if installed). SBOMs are generated with standard filenames (`bom.json`) for osv-scanner compatibility.

### vulnerability-analyzer-claude.sh

AI-enhanced SBOM analysis with Claude for contextual insights and pattern analysis.

**Features:**
- All features from basic analyzer
- **Pattern analysis** across vulnerabilities and dependencies
- **Supply chain context** and ecosystem health assessment
- **Exploitability context** with attack surface analysis
- **Risk narratives** identifying systemic issues
- **Business impact context** and maturity assessment
- Dependency relationship analysis
- Temporal trend identification

**What Claude Adds:**
- Pattern recognition across vulnerabilities
- Contextual understanding of supply chain risks
- Attack feasibility assessment
- Systemic issue identification
- Security posture narratives

**What's in Base Analyzer:**
- CISA KEV prioritization
- CVSS severity scoring
- Vulnerability counts and statistics
- Basic categorization

**Setup:**
```bash
# Option 1: Use .env file (recommended)
# Copy .env.example to .env and add your API key
cp ../../.env.example ../../.env
# Edit .env and set ANTHROPIC_API_KEY=sk-ant-xxx

# Option 2: Export environment variable
export ANTHROPIC_API_KEY=sk-ant-xxx
```

**Usage:**
```bash
# Analyze with AI insights (uses .env file or environment variable)
./vulnerability-analyzer-claude.sh /path/to/sbom.json

# Analyze repository with taint analysis
./vulnerability-analyzer-claude.sh --taint-analysis https://github.com/org/repo

# Or specify API key directly (overrides .env)
./vulnerability-analyzer-claude.sh --api-key sk-ant-xxx sbom.json
```

**Output Includes:**
1. **Pattern Analysis** - Vulnerability clustering, dependency chain risks, ecosystem health
2. **Supply Chain Context** - Critical path vulnerabilities, maintainer patterns, ecosystem-specific risks
3. **Exploitability Context** - Attack surface analysis, reachability insights, feasibility assessment
4. **Risk Narrative** - Systemic issues, security posture assessment, concerning patterns
5. **Business Impact** - Real-world risk evaluation, maturity assessment

**Requirements:**
- osv-scanner: `go install github.com/google/osv-scanner/cmd/osv-scanner@latest`
- syft (for SBOM generation): `brew install syft`
- jq: `brew install jq`
- Anthropic API key

**Note:** Run `./bootstrap.sh` from repository root to automatically check for and install all required dependencies.

### compare-analyzers.sh

Comparison tool that runs both basic and Claude-enhanced analyzers to demonstrate value-add.

**Features:**
- Runs both analyzers in parallel
- Compares outputs and capabilities
- Shows AI value-add with specific examples
- Generates comprehensive comparison report
- Optional output file preservation

**Usage:**
```bash
# Compare basic vs Claude analysis
./compare-analyzers.sh /path/to/sbom.json

# With taint analysis
./compare-analyzers.sh --taint-analysis sbom.json

# Keep output files for review
./compare-analyzers.sh --keep-outputs sbom.json
```

**Output:**
- Side-by-side capability comparison
- Value-add summary
- Use case recommendations
- Detailed output files (if --keep-outputs)

### Test with Safe Repositories

🧪 **Practice supply chain analysis safely:**

The [Gibson Powers Test Organization](https://github.com/Gibson-Powers-Test-Org) provides sample repositories with various dependency patterns for testing.

```bash
# Analyze test repository for vulnerabilities
./vulnerability-analyzer.sh \
  https://github.com/Gibson-Powers-Test-Org/sample-repo

# Run AI-enhanced analysis with prioritization
./vulnerability-analyzer-claude.sh --prioritize \
  https://github.com/Gibson-Powers-Test-Org/sample-repo

# Test SBOM analysis
./vulnerability-analyzer.sh /path/to/test-sbom.json
```

Perfect for:
- Learning vulnerability scanning
- Testing SBOM analysis
- Practicing risk prioritization
- Creating example reports

### CI/CD Integration

**GitHub Actions Example:**
```yaml
- name: SBOM Analysis
  run: |
    ./vulnerability-analyzer.sh --format json --output scan.json sbom.json

- name: AI-Enhanced Analysis (on main)
  if: github.ref == 'refs/heads/main'
  env:
    ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
  run: |
    ./vulnerability-analyzer-claude.sh sbom.json > analysis-report.txt
```

**GitLab CI Example:**
```yaml
sbom_scan:
  script:
    - ./vulnerability-analyzer.sh --format json --output scan.json sbom.json
  artifacts:
    reports:
      dependency_scanning: scan.json
```

## Resources

### Documentation
- [CycloneDX Specification](https://cyclonedx.org/specification/overview/)
- [SPDX Specification](https://spdx.dev/specifications/)
- [OSV.dev Documentation](https://google.github.io/osv.dev/)
- [deps.dev API v3alpha](https://docs.deps.dev/api/v3alpha/)
- [CISA KEV Catalog](https://www.cisa.gov/known-exploited-vulnerabilities-catalog)

### Tools
- **SBOM Generators**:
  - **syft** by Anchore (recommended) - Fast, accurate SBOM generation
  - CycloneDX CLI
  - SPDX tools
- **Vulnerability Scanners**:
  - **osv-scanner** (official OSV CLI) - Scan SBOMs and repositories
  - Grype by Anchore
  - Trivy
- **API Clients**: curl, Postman, custom scripts
- **Automation Scripts**: vulnerability-analyzer.sh, vulnerability-analyzer-claude.sh, compare-analyzers.sh

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
- Open an issue in the [GitHub repository](https://github.com/crashappsec/skills-and-prompts-and-rag/issues)
- Review existing [discussions](https://github.com/crashappsec/skills-and-prompts-and-rag/discussions)
- Contact: mark@crashoverride.com

---

**Made with expertise by the Crash Override community**
