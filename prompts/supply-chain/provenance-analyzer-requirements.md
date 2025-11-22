<!--
Copyright (c) 2024 Gibson Powers Contributors

SPDX-License-Identifier: GPL-3.0
-->

# SLSA Provenance Analyzer - Implementation Requirements

## Overview

Create a provenance verification tool that analyzes SBOMs and repositories to check package provenance using SLSA (Supply-chain Levels for Software Artifacts) standards. The tool should verify build provenance, assess SLSA levels, and validate attestations.

## Context

This tool will be part of the modular supply chain analysis suite located at:
- **Module location**: `utils/supply-chain/provenance-analysis/`
- **Integration**: Works standalone and through `supply-chain-scanner.sh`
- **Architecture pattern**: Follow the same structure as `vulnerability-analysis/`

### Reference Implementations
Use these as templates for structure and patterns:
- `utils/supply-chain/vulnerability-analysis/vulnerability-analyzer.sh`
- `utils/supply-chain/vulnerability-analysis/vulnerability-analyzer-claude.sh`
- `utils/supply-chain/supply-chain-scanner.sh`

## Core Requirements

### 1. SLSA Provenance Verification

**What is SLSA Provenance?**
SLSA provenance is cryptographically signed attestation about how a software artifact was built:
- **Build metadata**: Build platform, tools, parameters, timestamps
- **Source identity**: Repository, commit SHA, branch
- **Builder identity**: Who/what performed the build
- **Materials**: Dependencies and inputs used during build
- **Signature**: Cryptographic proof of authenticity

**SLSA Levels (0-4):**
- **Level 0**: No guarantees
- **Level 1**: Documentation of build process
- **Level 2**: Signed provenance, tamper resistance
- **Level 3**: Hardened builds, non-falsifiable provenance
- **Level 4**: Two-party review, hermetic builds

### 2. Script Functionality

Create two scripts following the established pattern:

#### `provenance-analyzer.sh` (Base Analyzer)
Basic provenance checking without AI:
- Verify provenance attestations exist
- Validate signatures using cosign/sigstore
- Check SLSA level compliance
- Verify build platform claims
- Validate source repository matches
- Check material completeness
- Assess SLSA level (0-4)
- Generate structured reports (JSON, table, markdown)

#### `provenance-analyzer-claude.sh` (AI-Enhanced)
Add Claude AI analysis for:
- **Trust Assessment**: Evaluate builder identity trustworthiness
- **Risk Context**: Analyze supply chain position and exposure
- **Pattern Recognition**: Identify suspicious build patterns
- **Policy Recommendations**: Suggest provenance improvements
- **Compliance Gaps**: Identify missing SLSA requirements
- **Comparative Analysis**: Compare provenance across ecosystem

### 3. Input Sources

Support multiple input types:

#### A. SBOM Files (CycloneDX/SPDX)
- Parse components from SBOM
- Extract package identifiers (purl)
- Look up provenance from registries
- Cross-reference with SBOM metadata

#### B. Git Repositories
- Detect package manifests (package.json, go.mod, requirements.txt, etc.)
- Generate SBOM if not present (using syft)
- Check repository's own provenance (for publishing)
- Analyze build configuration files

#### C. Direct Package Queries
- Accept package URLs (purl format)
- Support multiple ecosystems: npm, PyPI, Maven, Go, Docker, etc.
- Query registry APIs for provenance

### 4. Provenance Sources

#### Package Registries
- **npm**: Check for provenance via registry API
- **PyPI**: Look for attestations and signatures
- **Maven Central**: Check for PGP signatures
- **Go Modules**: Verify using go.sum and transparency log
- **Docker Hub/GHCR**: Check for image signatures and attestations
- **GitHub Releases**: Look for SLSA provenance attestations

#### Sigstore Integration
- Verify signatures using cosign
- Check transparency log (Rekor)
- Validate certificates
- Query fulcio for certificate info

#### SLSA Provenance Format
Support standard SLSA provenance formats:
- in-toto attestations
- SLSA Provenance v0.2 and v1.0
- GitHub SLSA provenance
- Google Cloud Build provenance

### 5. Verification Checks

Implement these verification steps:

#### Signature Verification
```bash
# Verify artifact signature
cosign verify --key <public-key> <artifact>

# Verify with keyless (sigstore)
cosign verify <artifact>

# Check transparency log
rekor-cli search --artifact <artifact>
```

#### Provenance Content Validation
- **Builder identity**: Trusted builder (GitHub Actions, Cloud Build, etc.)
- **Source repo**: Matches expected repository
- **Commit SHA**: Points to valid commit
- **Build parameters**: No suspicious flags
- **Materials**: All dependencies declared
- **Timestamps**: Within reasonable bounds

#### SLSA Level Assessment
Automated scoring against SLSA requirements:
```
SLSA Level 1:
✓ Build process documented
✓ Provenance exists

SLSA Level 2:
✓ Provenance signed
✓ Tamper-resistant
✓ Build service identity

SLSA Level 3:
✓ Hardened build platform
✓ Non-falsifiable provenance
✓ Isolated build

SLSA Level 4:
✓ Two-party review
✓ Hermetic builds
✓ Reproducible
```

### 6. Command-Line Interface

Follow the same pattern as vulnerability analyzers:

```bash
# Basic usage
./provenance-analyzer.sh <sbom|repo|package-url>

# Options
-f, --format FORMAT     Output format: table|json|markdown|sarif
-o, --output FILE       Write results to file
--strict               Fail on any missing provenance
--min-level LEVEL      Require minimum SLSA level (1-4)
--verify-signatures    Cryptographically verify all signatures
--check-builders LIST  Only trust specific builders
-k, --keep-clone       Keep cloned repositories

# Multi-repo scanning
--org ORG_NAME         Scan all repos in organization
--repo OWNER/REPO      Scan specific repository
--config FILE          Use alternate config file

# Examples
./provenance-analyzer.sh /path/to/sbom.json
./provenance-analyzer.sh --min-level 2 https://github.com/org/repo
./provenance-analyzer.sh --verify-signatures pkg:npm/express@4.17.1
./provenance-analyzer.sh --org myorg --min-level 1
```

### 7. Output Format

#### Basic Analyzer Output (Table)
```
========================================
  Provenance Analysis Results
========================================

Package: express@4.17.1 (npm)
---------------------------------------
Provenance:      ✓ Found
Signature:       ✓ Verified (cosign)
Builder:         GitHub Actions
Source Repo:     expressjs/express
Commit:          abc123def456
SLSA Level:      2
Transparency:    ✓ Rekor entry found

Checks:
  ✓ Signature valid
  ✓ Builder trusted
  ✓ Source matches expected
  ✓ Build reproducible
  ⚠ Not hermetically sealed
  ✗ No two-party review

Risk Assessment: MEDIUM
- Missing SLSA Level 3+ protections
- Supply chain transparency good
- Trusted builder and signature

========================================
Summary:
  Total packages:        25
  With provenance:       18 (72%)
  Signatures verified:   15 (60%)
  SLSA Level 2+:         12 (48%)
  Missing provenance:    7 (28%)
```

#### JSON Output
```json
{
  "scan_metadata": {
    "timestamp": "2024-11-21T10:30:00Z",
    "scanner": "provenance-analyzer v1.0.0",
    "target": "sbom.json",
    "total_packages": 25
  },
  "packages": [
    {
      "purl": "pkg:npm/express@4.17.1",
      "provenance": {
        "found": true,
        "format": "slsa-v1.0",
        "builder": {
          "id": "https://github.com/actions/runner",
          "trusted": true
        },
        "source": {
          "repo": "https://github.com/expressjs/express",
          "commit": "abc123def456",
          "verified": true
        },
        "signature": {
          "verified": true,
          "method": "cosign",
          "certificate": "fulcio-cert-xyz"
        },
        "slsa_level": 2,
        "checks": {
          "signature_valid": true,
          "builder_trusted": true,
          "source_matches": true,
          "build_reproducible": true,
          "hermetic_build": false,
          "two_party_review": false
        },
        "transparency_log": {
          "rekor_entry": "https://rekor.sigstore.dev/...",
          "verified": true
        }
      },
      "risk_level": "medium",
      "recommendations": [
        "Consider requiring SLSA Level 3 for production dependencies",
        "Hermetic builds would improve supply chain security"
      ]
    }
  ],
  "summary": {
    "with_provenance": 18,
    "signatures_verified": 15,
    "slsa_level_distribution": {
      "0": 7,
      "1": 6,
      "2": 8,
      "3": 3,
      "4": 1
    },
    "missing_provenance": 7
  }
}
```

### 8. Claude AI Analysis Focus

The Claude-enhanced version should provide:

#### Trust Assessment
- Evaluate builder reputation and history
- Assess source repository health signals
- Analyze organizational trust patterns

#### Risk Context
- Position in dependency tree (direct vs transitive)
- Package popularity and maintenance
- Historical provenance patterns

#### Pattern Recognition
- Unusual build configurations
- Inconsistent provenance across versions
- Ecosystem-specific risks

#### Policy Recommendations
```
Based on this analysis:

1. SLSA Level Distribution Concern
   - Only 48% of packages meet SLSA Level 2+
   - Critical path dependencies should require Level 3

2. Provenance Gaps
   - 7 packages without any provenance (28%)
   - Consider alternatives with better supply chain security

3. Builder Trust
   - Mix of GitHub Actions and unknown builders
   - Establish builder allowlist policy

4. Improvement Roadmap
   - Phase 1: Require provenance for all direct dependencies
   - Phase 2: Enforce SLSA Level 2 minimum
   - Phase 3: Migrate critical dependencies to Level 3+
```

### 9. Tool Dependencies

Required external tools:
- **cosign**: Signature verification (sigstore)
- **rekor-cli**: Transparency log queries
- **syft**: SBOM generation
- **jq**: JSON parsing
- **curl**: API queries
- **gh**: GitHub integration (optional, for org scanning)

Check and install in bootstrap.sh:
```bash
# Check cosign
if ! command -v cosign &> /dev/null; then
    echo "Install cosign: brew install cosign"
fi

# Check rekor-cli
if ! command -v rekor-cli &> /dev/null; then
    echo "Install rekor-cli: brew install rekor-cli"
fi
```

### 10. Integration with Central Orchestrator

Update `supply-chain-scanner.sh` to support provenance module:

```bash
# Add to MODULES
--provenance, -p     Run provenance analysis

# Usage
./supply-chain-scanner.sh --provenance --org myorg
./supply-chain-scanner.sh --vulnerability --provenance --org myorg
```

### 11. Configuration Support

Extend `config.json` with provenance settings:

```json
{
  "github": {
    "pat": "",
    "organizations": [],
    "repositories": []
  },
  "analysis": {
    "default_modules": ["vulnerability", "provenance"],
    "output_dir": "./supply-chain-reports"
  },
  "provenance": {
    "min_slsa_level": 2,
    "verify_signatures": true,
    "trusted_builders": [
      "https://github.com/actions/runner",
      "https://cloudbuild.googleapis.com"
    ],
    "fail_on_missing": false,
    "ecosystems": ["npm", "pypi", "go", "maven", "docker"]
  }
}
```

### 12. Registry API Integration

#### npm Registry
```bash
# Check npm provenance
curl -s "https://registry.npmjs.org/$package/$version" | \
  jq '.dist.signatures, .dist.attestations'
```

#### GitHub Packages
```bash
# Query GitHub for SLSA provenance
gh api repos/$owner/$repo/attestations/$sha
```

#### Sigstore/Rekor
```bash
# Search transparency log
rekor-cli search --artifact $artifact_hash
```

### 13. Error Handling

Graceful handling of:
- Missing provenance (common, not fatal)
- Signature verification failures (warning)
- Unreachable registries (retry with backoff)
- Invalid attestation formats (report and skip)
- Network timeouts (configurable)

### 14. Testing Strategy

Create test cases for:
- Packages with SLSA provenance (express, sigstore examples)
- Packages without provenance (legacy packages)
- Different SLSA levels (0-4)
- Multiple ecosystems (npm, PyPI, Go, Maven)
- Signature verification (valid, invalid, missing)
- Multi-repo scanning

### 15. Documentation

Create comprehensive docs:
- `utils/supply-chain/provenance-analysis/README.md`
- Usage examples for each ecosystem
- Troubleshooting guide
- SLSA level explanations
- Builder trust model

### 16. Performance Considerations

- Parallel package checking (batch API requests)
- Cache provenance lookups (temporary file)
- Rate limit handling for registries
- Async signature verification
- Progress indicators for large scans

## Implementation Checklist

### Phase 1: Basic Functionality
- [ ] Create directory structure
- [ ] Implement SBOM parsing
- [ ] Add package purl extraction
- [ ] Implement npm registry queries
- [ ] Basic provenance detection
- [ ] SLSA level assessment
- [ ] Table output format

### Phase 2: Signature Verification
- [ ] Integrate cosign for verification
- [ ] Add Rekor transparency log checks
- [ ] Certificate validation
- [ ] Keyless verification support
- [ ] Builder identity checking

### Phase 3: Multi-Ecosystem Support
- [ ] PyPI provenance checking
- [ ] Go module verification
- [ ] Maven/Java support
- [ ] Docker image attestations
- [ ] GitHub Packages integration

### Phase 4: Multi-Repo Support
- [ ] Add --org and --repo flags
- [ ] Config file integration
- [ ] GitHub CLI integration
- [ ] Batch processing
- [ ] Progress reporting

### Phase 5: Claude AI Integration
- [ ] Create Claude analyzer script
- [ ] Trust assessment prompts
- [ ] Risk context analysis
- [ ] Policy recommendations
- [ ] Comparison with base analyzer

### Phase 6: Integration & Polish
- [ ] Update central orchestrator
- [ ] Add to bootstrap.sh
- [ ] Update documentation
- [ ] Add to CHANGELOG.md
- [ ] Create examples

## Success Criteria

The implementation is complete when:
1. ✅ Can verify provenance for npm, PyPI, Go, Maven packages
2. ✅ Validates signatures using cosign
3. ✅ Accurately assesses SLSA levels (0-4)
4. ✅ Supports all input types (SBOM, repo, package URL)
5. ✅ Works standalone and through central orchestrator
6. ✅ Multi-repo scanning functional
7. ✅ Claude AI provides actionable insights
8. ✅ Clear documentation and examples
9. ✅ Comprehensive error handling
10. ✅ Performance acceptable for large SBOMs (100+ packages)

## Resources

### SLSA Specifications
- https://slsa.dev/spec/v1.0/
- https://slsa.dev/provenance/

### Sigstore Documentation
- https://docs.sigstore.dev/
- https://github.com/sigstore/cosign
- https://github.com/sigstore/rekor

### Package Ecosystems
- npm: https://github.blog/2023-04-19-introducing-npm-package-provenance/
- PyPI: https://blog.pypi.org/posts/2023-05-23-introducing-trusted-publishers/
- Go: https://go.dev/blog/module-mirror-launch
- Maven: https://central.sonatype.org/publish/requirements/

### In-toto Attestations
- https://in-toto.io/
- https://github.com/in-toto/attestation

## Example Test Cases

### Test Case 1: Well-Signed npm Package
```bash
./provenance-analyzer.sh pkg:npm/sigstore@latest
# Expected: SLSA Level 3, verified signature, GitHub Actions builder
```

### Test Case 2: Legacy Package Without Provenance
```bash
./provenance-analyzer.sh pkg:npm/left-pad@1.3.0
# Expected: SLSA Level 0, no provenance, warning issued
```

### Test Case 3: Full Repository Scan
```bash
./provenance-analyzer.sh --min-level 2 https://github.com/myorg/myrepo
# Expected: Analyze all dependencies, report Level 2+ compliance
```

### Test Case 4: Organization Scan
```bash
./provenance-analyzer.sh --org myorg --verify-signatures
# Expected: Scan all repos, verify all signatures, summary report
```

## Notes

- Follow the same code style and patterns as vulnerability analyzers
- Use consistent error messages and color coding
- Maintain compatibility with existing config.json
- Ensure all scripts are POSIX-compliant bash
- Add proper copyright headers to all files
- Update central CHANGELOG.md with new features
