#!/bin/bash
# Batch create all skill templates

set -euo pipefail

TEMPLATES_DIR="$HOME/claude-templates"

echo "=== Creating All Skill Templates ==="
echo "Output: $TEMPLATES_DIR"
echo ""

# Create all category directories
mkdir -p "$TEMPLATES_DIR"/{supply-chain,dora-metrics,code-ownership,security,meta}

#############################################################################
# DORA Metrics Template
#############################################################################

cat > "$TEMPLATES_DIR/dora-metrics/dora-analysis.md" << 'EOF'
# DORA Metrics Analysis - Portable Template

## Purpose
Measure and improve software delivery performance using the four key DORA metrics: Deployment Frequency, Lead Time for Changes, Change Failure Rate, and Mean Time to Recovery.

## When to Use
- Measure team/organization DevOps performance
- Benchmark against industry standards (Elite/High/Medium/Low performers)
- Track delivery improvements over time
- Compare multiple teams
- Generate executive reports on delivery health

## Required Context
**Minimum**: Repository URL or Git history access
**Optimal**: CI/CD logs, incident tracking data, deployment records

## Prompt Template

---

I need a comprehensive DORA metrics analysis for [REPOSITORY_URL/TEAM_NAME].

Please analyze the following four key metrics:

### 1. Deployment Frequency (DF)
How often does the team deploy to production?
- Calculate: deployments per day/week/month
- Classify: Elite (multiple/day), High (daily-weekly), Medium (weekly-monthly), Low (monthly+)
- Trends: Compare last 30/60/90 days

### 2. Lead Time for Changes (LT)
Time from commit to production deployment
- Calculate: Average time from commit timestamp to deployment
- Classify: Elite (<1 hour), High (1 day-1 week), Medium (1 week-1 month), Low (1+ month)
- Bottlenecks: Identify where delays occur (review, testing, approval, deployment)

### 3. Change Failure Rate (CFR)
Percentage of deployments causing production failures
- Calculate: (Failed deployments / Total deployments) × 100
- Classify: Elite (0-15%), High (16-30%), Medium (31-45%), Low (46%+)
- Patterns: When do failures occur? What types?

### 4. Mean Time to Recovery (MTTR)
How quickly the team recovers from failures
- Calculate: Average time from incident detection to resolution
- Classify: Elite (<1 hour), High (1 hour-1 day), Medium (1 day-1 week), Low (1+ week)
- Analysis: What slows recovery? Rollback vs. fix-forward?

## Analysis Requirements

**Performance Classification**: Overall rating (Elite/High/Medium/Low)
**Benchmark Comparison**: How does this compare to DORA research findings?
**Trends**: Are metrics improving, stable, or declining?
**Insights**: What patterns emerge? What's working well?
**Bottlenecks**: Where are the constraints in the delivery pipeline?
**Risks**: What threatens performance sustainability?

## Output Format

1. **Executive Summary** (3-5 sentences)
   - Overall performance level
   - Key strengths and concerns
   - Primary recommendation

2. **Metrics Dashboard**
   ```
   Metric                    Current    Classification    Trend
   ─────────────────────────────────────────────────────────────
   Deployment Frequency      X/day      Elite/High/...    ↑/→/↓
   Lead Time for Changes     X hours    Elite/High/...    ↑/→/↓
   Change Failure Rate       X%         Elite/High/...    ↑/→/↓
   Mean Time to Recovery     X hours    Elite/High/...    ↑/→/↓
   ```

3. **Detailed Analysis** per metric
   - Current performance with data
   - Classification and gap to next level
   - Trend analysis
   - Root causes of current state

4. **Improvement Roadmap**
   - Quick wins (0-30 days)
   - Medium-term improvements (1-3 months)
   - Strategic initiatives (3-6 months)

Format as markdown with tables and visualizations where helpful.

---

## Usage Examples

### Claude Desktop
1. Copy template above
2. Replace `[REPOSITORY_URL/TEAM_NAME]` with your repository or team name
3. Paste into Claude Desktop
4. **Optional**: Attach deployment logs, CI/CD data, incident reports

### Terminal
```bash
cat ~/claude-templates/dora-metrics/dora-analysis.md | \
  sed 's|\[REPOSITORY_URL/TEAM_NAME\]|https://github.com/myorg/api|' | \
  claude
```

### CI/CD Integration
```yaml
# Monthly DORA metrics report
- name: Generate DORA Report
  run: |
    PROMPT=$(cat .templates/dora-analysis.md | \
             sed "s|\[REPOSITORY_URL/TEAM_NAME\]|${GITHUB_REPOSITORY}|")
    # Call Claude API with deployment data attached
```

## Customization

**Focus on specific metric**: Remove sections for metrics you don't need
**Team comparison**: Run for multiple teams and compare results
**Historical analysis**: Request analysis across multiple time periods
**Custom thresholds**: Define your own Elite/High/Medium/Low criteria

## Related Templates
- `deployment-frequency-deep-dive.md` - DF-specific analysis
- `lead-time-breakdown.md` - LT bottleneck identification
- `change-failure-analysis.md` - CFR root cause analysis

**Source**: Exported from `skills/dora-metrics/`
**Last Updated**: 2025-11-21
EOF

echo "✓ Created: dora-metrics/dora-analysis.md"

#############################################################################
# Code Ownership Template
#############################################################################

cat > "$TEMPLATES_DIR/code-ownership/ownership-analysis.md" << 'EOF'
# Code Ownership Analysis - Portable Template

## Purpose
Analyze code ownership distribution, identify bus factor risks, validate CODEOWNERS files, and plan knowledge transfer.

## When to Use
- Assess code ownership health and distribution
- Identify single points of failure (bus factor analysis)
- Validate/generate CODEOWNERS files
- Plan succession and knowledge transfer
- Onboard new team members (find code owners)
- Before organizational changes (layoffs, restructuring)

## Required Context
**Minimum**: Repository URL or local path
**Optimal**: CODEOWNERS file (if exists), team structure, time period

## Prompt Template

---

I need a comprehensive code ownership analysis for [REPOSITORY_URL].

Please analyze the following areas:

### 1. Ownership Distribution
- Who owns what code (by file/directory)?
- How is ownership distributed across the team?
- Gini coefficient (0=perfect equality, 1=one person owns everything)
- Top contributors by ownership percentage

### 2. Bus Factor Analysis
- What is the bus factor? (minimum people who could leave before project stalls)
- Which files/areas have single points of failure?
- Critical files with only one contributor
- Risk assessment: Critical/High/Medium/Low

### 3. CODEOWNERS Validation
[If CODEOWNERS file exists]
- Does CODEOWNERS match actual ownership?
- Coverage: what % of files have owners?
- Staleness: are owners still active?
- Conflicts: multiple strong owners for same area?

[If no CODEOWNERS]
- Generate CODEOWNERS based on actual Git history
- Suggest ownership patterns (by directory, file type)

### 4. Knowledge Transfer Risks
- Files with no backup owners
- Areas where knowledge is concentrated
- Recommended mentorship pairings
- Succession planning priorities

### 5. Collaboration Patterns
- Which files have healthy collaboration (multiple contributors)?
- Isolated silos (one person, no collaboration)?
- Cross-team boundaries and interfaces

## Analysis Period
**Default**: Last 90 days
**Specify if different**: [TIME_PERIOD]

## Output Format

1. **Executive Summary**
   - Overall health score (0-100)
   - Bus factor number
   - Critical risks requiring attention
   - Top recommendation

2. **Ownership Metrics**
   ```
   Contributor          Files Owned    Percentage    Status
   ────────────────────────────────────────────────────────
   alice@company.com    245            35%           ⚠️ High concentration
   bob@company.com      180            25%           ✓ Healthy
   ```

3. **Bus Factor & SPOFs**
   - Bus factor: X people
   - Single points of failure (SPOFs):
     - `src/auth/core.ts` - Only alice (Critical)
     - `db/migrations/` - Only bob (High)

4. **CODEOWNERS Assessment**
   - Coverage: X% of files
   - Accuracy: X% match actual ownership
   - Recommended updates

5. **Succession Planning**
   - Priority areas for knowledge transfer
   - Recommended mentor/mentee pairs
   - Action items with timeline

Format as markdown with tables and clear risk indicators.

---

## Usage Examples

### Claude Desktop
1. Copy template
2. Replace `[REPOSITORY_URL]` with your repository
3. Replace `[TIME_PERIOD]` if not 90 days (e.g., "last 6 months")
4. Paste into Claude Desktop

### Terminal
```bash
cat ~/claude-templates/code-ownership/ownership-analysis.md | \
  sed 's|\[REPOSITORY_URL\]|https://github.com/myorg/backend|' | \
  sed 's|\[TIME_PERIOD\]|last 180 days|' | \
  claude
```

### Automation
```bash
# Quarterly ownership audit
./utils/code-ownership/ownership-analyzer-v2.sh . --format json | \
  claude "Analyze this ownership data: $(cat ownership.json)"
```

## Customization

**Focus areas**:
- Security-critical code only
- Specific directories (e.g., `/src/api/`)
- New team members (identify who to contact)

**Time periods**:
- Recent activity: 30 days
- Standard: 90 days
- Historical: 6-12 months

## Related Templates
- `bus-factor-mitigation.md` - Reducing SPOF risks
- `codeowners-generation.md` - Automated CODEOWNERS creation
- `knowledge-transfer-plan.md` - Succession planning

**Source**: Exported from `skills/code-ownership/`
**Last Updated**: 2025-11-21
EOF

echo "✓ Created: code-ownership/ownership-analysis.md"

#############################################################################
# Supply Chain Comprehensive Template
#############################################################################

cat > "$TEMPLATES_DIR/supply-chain/supply-chain-comprehensive.md" << 'EOF'
# Supply Chain Comprehensive Analysis - Portable Template

## Purpose
Complete software supply chain analysis including SBOM review, vulnerability assessment, provenance verification, and dependency health evaluation.

## When to Use
- Security audits and compliance reviews
- Pre-acquisition due diligence
- Vendor/dependency evaluation
- Incident response (supply chain compromise)
- Regulatory compliance (EO 14028, SSDF)
- Open source risk assessment

## Required Context
**Minimum**: Repository URL OR SBOM file (CycloneDX/SPDX)
**Optimal**: Build logs, SLSA attestations, vulnerability scan results

## Prompt Template

---

I need a comprehensive supply chain security analysis for [REPOSITORY_URL/SBOM_FILE].

Please analyze across these dimensions:

### 1. SBOM Quality & Completeness
- Format and version (CycloneDX 1.x, SPDX 2.x)
- Completeness: components, dependencies, metadata
- Missing information (licenses, versions, purls)
- Generation tool and quality

### 2. Vulnerability Assessment
- Known CVEs in dependencies (direct + transitive)
- Severity distribution (Critical/High/Medium/Low)
- Exploitability analysis (EPSS scores where available)
- Patch availability and version gaps
- Taint analysis: are vulnerable functions actually used?

### 3. Dependency Health
- Maintenance status (active, stale, abandoned, deprecated)
- Update recency and release frequency
- Community health (contributors, stars, issues)
- Version currency (latest vs. in-use)
- Technical debt (outdated dependencies)

### 4. License Compliance
- License distribution (MIT, Apache, GPL, proprietary)
- License conflicts and incompatibilities
- Copyleft obligations
- Commercial/proprietary dependencies
- Undeclared or unknown licenses

### 5. Provenance & Integrity
- Package source verification
- SLSA provenance level (if available)
- Signature verification status
- Build reproducibility
- Dependency confusion risks
- Typosquatting checks

### 6. Operational Risks
- Deprecated packages requiring migration
- Single points of failure (critical deps with single maintainer)
- Orphaned or unmaintained dependencies
- Supply chain attack surface
- Transitive dependency depth

## Risk Classification
Use this scale for findings:
- **Critical**: Immediate action required (known exploitation, critical CVE)
- **High**: Address within days (high-severity CVE, abandoned critical dep)
- **Medium**: Address within weeks (moderate CVE, technical debt)
- **Low**: Monitor or address opportunistically

## Output Format

1. **Executive Summary** (5 bullet points max)
   - Overall risk posture (Critical/High/Medium/Low)
   - Most critical findings
   - Compliance status
   - Key recommendations

2. **Risk Dashboard**
   ```
   Category              Critical  High  Medium  Low  Status
   ──────────────────────────────────────────────────────────
   Vulnerabilities       X         X     X       X    ⚠️/✓
   License Compliance    X         X     X       X    ⚠️/✓
   Dependency Health     X         X     X       X    ⚠️/✓
   Provenance           X         X     X       X    ⚠️/✓
   ```

3. **Critical Findings**
   Top 10 issues requiring immediate attention with:
   - Description and location
   - Risk level and business impact
   - Evidence/data
   - Recommended action

4. **Detailed Analysis** per category
   - Objective findings with data
   - Severity assessment
   - Trend information where available

5. **Dependency Inventory**
   Table of all dependencies with health scores

6. **Recommendations** (prioritized)
   - Immediate actions
   - Short-term improvements (1-3 months)
   - Strategic initiatives

Format as markdown with clear sections, tables, and severity indicators.

---

## Usage Examples

### Claude Desktop - Quick Scan
```
I need a supply chain analysis for https://github.com/expressjs/express
Focus on: Critical vulnerabilities and license compliance
```

### Claude Desktop - Deep Dive
1. Copy full template
2. Replace [REPOSITORY_URL/SBOM_FILE]
3. Attach SBOM file if you have one
4. Paste into conversation

### Terminal with SBOM
```bash
# Generate SBOM first
syft dir:. -o cyclonedx-json > sbom.json

# Analyze with template
cat ~/claude-templates/supply-chain/supply-chain-comprehensive.md | \
  sed 's|\[REPOSITORY_URL/SBOM_FILE\]|sbom.json (attached)|' | \
  claude --attach sbom.json
```

### Automated Pipeline
```yaml
- name: Supply Chain Audit
  run: |
    syft dir:. -o cyclonedx-json > sbom.json
    # Call Claude API with template + SBOM
    # Generate report and create issue if risks found
```

## Customization Options

**Quick scans**: Focus on specific categories
```
Focus on: Vulnerability assessment only
Skip: License compliance, provenance
```

**Specific ecosystems**:
```
Analyze only: npm dependencies
Ignore: dev dependencies
```

**Compliance-focused**:
```
Primary concern: SLSA Level 2 compliance
Secondary: License GPL conflicts
```

## Related Templates
- `vulnerability-deep-dive.md` - Detailed CVE analysis
- `license-audit.md` - License compliance only
- `provenance-verification.md` - SLSA/in-toto analysis
- Package health analysis (already created)

## Tools Integration
Works with data from:
- Syft (SBOM generation)
- Grype (vulnerability scanning)
- OSV.dev (vulnerability database)
- deps.dev (package metadata)
- Sigstore/cosign (provenance verification)

**Source**: Exported from `skills/supply-chain/`
**Last Updated**: 2025-11-21
EOF

echo "✓ Created: supply-chain/supply-chain-comprehensive.md"

#############################################################################
# Certificate Security Analysis Template
#############################################################################

cat > "$TEMPLATES_DIR/security/certificate-analysis.md" << 'EOF'
# Certificate Security Analysis - Portable Template

## Purpose
Analyze X.509 certificates and TLS configurations for security issues, compliance, and operational risks.

## When to Use
- Security audits of web services/APIs
- Certificate expiration monitoring
- TLS configuration review
- Compliance validation (PCI-DSS, HIPAA, SOC 2)
- Incident response (certificate-related outages)
- Certificate lifecycle management

## Required Context
**Minimum**: Domain name OR certificate file OR certificate chain
**Optimal**: Full TLS handshake data, server configuration

## Prompt Template

---

I need a certificate security analysis for [DOMAIN/CERTIFICATE_FILE].

Please analyze:

### 1. Certificate Validity & Trust
- Expiration date and time until expiry
- Issuer and certificate authority trust chain
- Signature algorithm strength
- Key size and algorithm (RSA 2048+, ECDSA P-256+)
- Certificate transparency (CT) log presence

### 2. Subject & Identity
- Common Name (CN) and Subject Alternative Names (SANs)
- Wildcard usage and scope
- Organization validation level (DV/OV/EV)
- Domain ownership verification

### 3. TLS Configuration
- Supported TLS versions (TLS 1.2, 1.3 only?)
- Cipher suite strength and ordering
- Perfect Forward Secrecy (PFS) support
- HTTP Strict Transport Security (HSTS) headers
- Certificate pinning configuration

### 4. Security Issues
- Weak algorithms (MD5, SHA-1, RC4)
- Small key sizes (<2048 RSA, <256 EC)
- Deprecated TLS versions (SSL, TLS 1.0/1.1)
- Missing security headers
- Certificate revocation status (OCSP, CRL)

### 5. Operational Risks
- Expiration warnings (30/60/90 days)
- Certificate chain completeness
- Renewal process and automation
- Monitoring and alerting gaps
- Multi-domain coverage issues

## Risk Levels
- **Critical**: Active security vulnerability, imminent expiration
- **High**: Weak crypto, deprecated protocols, expiring <30 days
- **Medium**: Best practice violations, expiring 30-90 days
- **Low**: Recommendations, expiring >90 days

## Output Format

1. **Executive Summary**
   - Overall security posture (Secure/At Risk/Vulnerable)
   - Critical issues count
   - Expiration status
   - Primary recommendation

2. **Certificate Details**
   ```
   Subject:         example.com
   Issuer:          Let's Encrypt
   Valid From:      2024-11-01
   Valid Until:     2025-02-01 (72 days remaining)
   Key Type:        RSA 2048
   Signature:       SHA-256 with RSA
   Status:          ✓ Valid
   ```

3. **Security Assessment**
   - TLS version support
   - Cipher suite analysis
   - Certificate chain validation
   - Revocation status

4. **Issues & Recommendations**
   Prioritized list with:
   - Issue description
   - Risk level
   - Business impact
   - Remediation steps

Format as markdown with status indicators (✓/⚠️/❌)

---

## Usage Examples

### Claude Desktop - Domain Analysis
```
Analyze certificate for: api.example.com
Include: TLS configuration review
```

### Terminal - Certificate File
```bash
# Analyze cert file
openssl x509 -in cert.pem -text -noout > cert-details.txt

cat ~/claude-templates/security/certificate-analysis.md | \
  sed 's|\[DOMAIN/CERTIFICATE_FILE\]|cert-details.txt (attached)|' | \
  claude --attach cert-details.txt
```

### Bulk Analysis
```bash
# Check multiple domains
for domain in api.example.com www.example.com; do
  echo "=== $domain ==="
  echo | openssl s_client -connect $domain:443 -servername $domain 2>/dev/null | \
    openssl x509 -text -noout | \
    claude "Analyze this certificate: $(cat -)"
done
```

## Customization

**Compliance-focused**:
```
Focus on: PCI-DSS compliance
Requirements: TLS 1.2+, strong ciphers only
```

**Expiration monitoring**:
```
Alert threshold: 30 days
Include: Renewal procedure validation
```

**Internal PKI**:
```
Certificate type: Internal CA
Validate against: Corporate PKI policy
```

## Related Templates
- `tls-configuration-hardening.md` - TLS security improvements
- `certificate-lifecycle.md` - Certificate management process

**Source**: Exported from `skills/certificate-analyzer/`
**Last Updated**: 2025-11-21
EOF

echo "✓ Created: security/certificate-analysis.md"

#############################################################################
# Chalk Build Analysis Template
#############################################################################

cat > "$TEMPLATES_DIR/security/chalk-build-analysis.md" << 'EOF'
# Chalk Build Attestation Analysis - Portable Template

## Purpose
Analyze Chalk build attestations and metadata for supply chain security, build provenance, and compliance verification.

## When to Use
- Verify software build provenance
- SLSA compliance assessment
- Build artifact validation
- CI/CD security review
- Supply chain attack detection
- Reproducible builds verification

## Required Context
**Minimum**: Chalk mark/attestation file OR artifact with embedded Chalk mark
**Optimal**: Build logs, expected build parameters

## Prompt Template

---

I need a Chalk build attestation analysis for [ARTIFACT/CHALK_MARK_FILE].

Please analyze:

### 1. Build Provenance
- Build timestamp and duration
- Build environment details (OS, architecture)
- Builder identity and authentication
- Source code reference (commit SHA, branch)
- Build reproducibility indicators

### 2. Artifact Metadata
- Artifact identity (hash, size, type)
- Dependencies captured
- Embedded credentials scan
- File integrity verification
- Metadata completeness

### 3. Supply Chain Security
- SLSA provenance level achieved
- Build isolation and hermeticity
- Dependency pinning status
- Secrets detection
- Unauthorized modification indicators

### 4. Compliance & Attestation
- Signature verification status
- Attestation format (in-toto, SLSA)
- Certificate chain validation
- Timestamp authority verification
- Policy compliance check

### 5. Build Quality Indicators
- Build warnings/errors
- Test execution evidence
- Code coverage metadata
- Security scan results
- Quality gate passage

## Output Format

1. **Executive Summary**
   - Provenance verified: Yes/No
   - SLSA level: L0/L1/L2/L3/L4
   - Issues found: count by severity
   - Trust assessment

2. **Provenance Details**
   ```
   Artifact:      example-app:1.2.3
   Built:         2024-11-21 14:30:00 UTC
   Builder:       GitHub Actions
   Source:        github.com/org/repo@abc123
   SLSA Level:    L3
   Verified:      ✓ Signature valid
   ```

3. **Security Analysis**
   - Secrets detection results
   - Dependency verification
   - Build environment security
   - Tampering indicators

4. **Compliance Assessment**
   - SLSA requirements met/unmet
   - Policy violations
   - Recommended improvements

Format as markdown with verification indicators.

---

## Usage Examples

### Claude Desktop
1. Extract Chalk mark from artifact:
   ```bash
   chalk extract artifact.bin > chalk-mark.json
   ```
2. Copy template, attach chalk-mark.json
3. Paste into Claude Desktop

### Terminal
```bash
chalk extract my-binary > chalk-mark.json

cat ~/claude-templates/security/chalk-build-analysis.md | \
  sed 's|\[ARTIFACT/CHALK_MARK_FILE\]|chalk-mark.json|' | \
  claude --attach chalk-mark.json
```

### CI/CD Verification
```yaml
- name: Verify Build Attestation
  run: |
    chalk extract artifact > chalk-mark.json
    # Analyze with Claude for compliance
    # Fail build if issues found
```

## Customization

**SLSA-focused**:
```
Requirement: SLSA Level 3 compliance
Validate: All SLSA L3 requirements
```

**Security-focused**:
```
Priority: Secrets detection, dependency verification
Flag: Any embedded credentials or unsigned dependencies
```

**Source**: Exported from `skills/chalk-build-analyzer/`
**Last Updated**: 2025-11-21
EOF

echo "✓ Created: security/chalk-build-analysis.md"

#############################################################################
# Better Prompts Template
#############################################################################

cat > "$TEMPLATES_DIR/meta/improve-prompt.md" << 'EOF'
# Prompt Improvement - Portable Template

## Purpose
Improve prompts for clarity, effectiveness, and better Claude responses using prompt engineering best practices.

## When to Use
- Getting inconsistent results from prompts
- Need to optimize existing prompts
- Creating new prompts for complex tasks
- Improving template quality
- Training team on prompt engineering

## Required Context
**Input**: Your current prompt or task description

## Prompt Template

---

Please help me improve this prompt using best practices:

**Current Prompt:**
```
[PASTE_YOUR_CURRENT_PROMPT_HERE]
```

**Task Context:**
[DESCRIBE_WHAT_YOU'RE_TRYING_TO_ACHIEVE]

**Current Issues** (optional):
[WHAT_PROBLEMS_ARE_YOU_EXPERIENCING]

Please analyze and improve this prompt by:

### 1. Clarity & Structure
- Is the request clear and unambiguous?
- Is it well-organized with logical sections?
- Are there vague terms that should be specific?

### 2. Context & Constraints
- Does it provide sufficient context?
- Are constraints and requirements clearly stated?
- Is the scope well-defined?

### 3. Output Specification
- Is the desired output format clear?
- Are examples provided where helpful?
- Are success criteria defined?

### 4. Best Practices Application
- Role assignment (if beneficial)
- Chain of thought prompting
- Few-shot examples
- Output formatting
- Error handling

### 5. Prompt Optimization
- Remove ambiguity
- Add missing context
- Improve structure
- Enhance clarity

## Output Format

Provide:

1. **Analysis** of current prompt
   - Strengths
   - Weaknesses
   - Improvement opportunities

2. **Improved Prompt**
   - Rewritten version
   - Structured sections
   - Clear expectations

3. **Explanation** of changes made
   - What was changed and why
   - Expected improvements

4. **Usage Tips**
   - When to use this version
   - How to customize further
   - Common pitfalls to avoid

---

## Usage Examples

### Claude Desktop
Copy template, paste your prompt in [PASTE_YOUR_CURRENT_PROMPT_HERE], add context, send.

### Terminal
```bash
cat ~/claude-templates/meta/improve-prompt.md | \
  sed 's|\[PASTE_YOUR_CURRENT_PROMPT_HERE\]|Analyze this code for bugs|' | \
  sed 's|\[DESCRIBE_WHAT_YOU'RE_TRYING_TO_ACHIEVE\]|Finding security vulnerabilities in Python|' | \
  claude
```

### Iterative Improvement
1. Test original prompt
2. Use this template to improve
3. Test improved version
4. Iterate if needed

## Customization

**Focus areas**:
```
Improve specifically: Output format consistency
Keep: Current tone and style
```

**Constraints**:
```
Must: Be under 500 words
Should: Include examples
```

**Source**: Exported from `skills/better-prompts/`
**Last Updated**: 2025-11-21
EOF

echo "✓ Created: meta/improve-prompt.md"

echo ""
echo "=== Template Creation Complete ==="
echo ""
echo "Created 7 portable templates:"
echo "  • dora-metrics/dora-analysis.md"
echo "  • code-ownership/ownership-analysis.md"
echo "  • supply-chain/supply-chain-comprehensive.md"
echo "  • supply-chain/package-health-analysis.md (already existed)"
echo "  • security/certificate-analysis.md"
echo "  • security/chalk-build-analysis.md"
echo "  • meta/improve-prompt.md"
echo ""
echo "Location: $TEMPLATES_DIR"
echo ""
echo "Next steps:"
echo "  1. Review templates: ls -R $TEMPLATES_DIR"
echo "  2. Try one: open $TEMPLATES_DIR/dora-metrics/dora-analysis.md"
echo "  3. See examples: $TEMPLATES_DIR/DEMO-USAGE.sh"
echo ""
EOF

chmod +x batch-create-templates.sh
