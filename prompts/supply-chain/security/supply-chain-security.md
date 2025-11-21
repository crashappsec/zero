<!--
Copyright (c) 2024 Crash Override Inc
101 Fulton St, 416, New York 10038
SPDX-License-Identifier: GPL-3.0
-->

# Prompt: Supply Chain Security Assessment

## Purpose
Evaluate the supply chain security posture of software components in an SBOM using provenance, attestations, and security metrics.

## When to Use
- Pre-deployment security review
- Vendor software assessment
- Compliance with supply chain security frameworks (SLSA, SSDF)
- Risk assessment for critical applications
- Zero-trust security verification

## Prompt

```
Please perform a comprehensive supply chain security assessment of this SBOM:

1. Provenance Verification:
   - Check for SLSA attestations
   - Verify build provenance data
   - Identify unsigned or unverified components

2. Component Integrity:
   - Validate cryptographic signatures where available
   - Check for tamper indicators
   - Verify package hashes

3. Security Metrics:
   - Query OpenSSF Scorecard for each component
   - Assess maintainer reputation
   - Check for security policy presence
   - Evaluate vulnerability disclosure process

4. Risk Indicators:
   - Abandoned or unmaintained packages
   - Packages with known supply chain attacks
   - Typosquatting risks (use deps.dev GetSimilarlyNamedPackages)
   - Single-maintainer dependencies
   - Recent ownership changes

5. Recommendations:
   - Components requiring additional verification
   - Alternative packages with better supply chain security
   - Risk mitigation strategies

[Paste SBOM content here]
```

## Expected Output
- Provenance verification summary
- OpenSSF Scorecard results table
- Supply chain risk assessment
- Unsigned/unverified component list
- Security posture score
- Actionable recommendations

## Variations

### SLSA Compliance Check
```
Check this SBOM for SLSA (Supply-chain Levels for Software Artifacts) compliance.
Identify components with SLSA attestations and their levels.
Flag components without provenance data.

[Paste SBOM]
```

### Typosquatting Detection
```
Check this SBOM for potential typosquatting risks.
For each component, use deps.dev to find similarly-named packages.
Flag any suspicious packages or naming that could indicate typosquatting.

[Paste SBOM]
```

### Maintainer Trust Assessment
```
Assess the trustworthiness of package maintainers in this SBOM:
- Number of maintainers per package
- Maintainer reputation metrics
- Recent ownership transfers
- GitHub/GitLab organization vs. individual accounts

[Paste SBOM]
```

### Zero Trust Verification
```
Perform a zero-trust verification of this SBOM:
- Assume all components are untrusted
- Verify every signature and attestation
- Check for reproducible builds
- Validate all cryptographic artifacts
- Identify components that cannot be verified

[Paste SBOM]
```

## Examples

### Example Usage
```
Please perform a comprehensive supply chain security assessment...

[CycloneDX SBOM with provenance data]
```

### Example Output Structure
```markdown
# Supply Chain Security Assessment

## Provenance Summary
- Components with SLSA attestations: 12 / 45 (27%)
- Signed components: 15 / 45 (33%)
- Verified build provenance: 8 / 45 (18%)
- Unverified components: 30 / 45 (67%)

## OpenSSF Scorecard Results

| Component | Overall Score | Security Policy | Vuln Response | Signed Releases |
|-----------|--------------|-----------------|---------------|-----------------|
| express | 6.7/10 | ✅ | ⚠️ | ❌ |
| axios | 7.2/10 | ✅ | ✅ | ⚠️ |
| lodash | 5.9/10 | ⚠️ | ❌ | ❌ |

## Supply Chain Risks

### 🔴 Critical Risks
- 30 components without signature verification
- 2 packages with single maintainer
- 1 package with recent ownership change

### 🟠 Medium Risks
- 15 packages without OpenSSF Scorecard data
- 5 packages using deprecated build systems

## Typosquatting Check
✅ No suspicious package names detected

## Recommendations
1. [HIGH] Require SLSA Level 2+ for all future dependencies
2. [MEDIUM] Request signed SBOMs from vendors
3. [MEDIUM] Monitor ownership changes for critical deps
...
```

## Tips
- Specify your organization's supply chain security requirements
- Mention compliance frameworks you need to meet (SLSA, NIST SSDF, etc.)
- Include risk tolerance levels
- Request specific verification depth based on criticality
