# Knowledge Architecture

## Overview

The knowledge system provides structured, reusable information that specialist agents can reference during analysis. Knowledge is organized by domain, with both human-readable documentation and machine-readable pattern databases.

## Directory Structure

```
knowledge/
├── security/
│   ├── vulnerabilities/
│   │   ├── cwe-database.json        # CWE patterns and remediation
│   │   ├── owasp-top-10.json        # OWASP Top 10 reference
│   │   └── cvss-guide.md            # CVSS scoring guidance
│   ├── threats/
│   │   ├── mitre-attack.json        # ATT&CK techniques
│   │   ├── stride-methodology.md    # STRIDE threat modeling
│   │   └── attack-patterns.json     # Common attack patterns
│   ├── secrets/
│   │   ├── secret-patterns.json     # Regex patterns for secrets
│   │   └── rotation-guides.md       # Secret rotation procedures
│   └── containers/
│       ├── cis-docker.json          # CIS Docker Benchmark
│       ├── base-images.json         # Recommended base images
│       └── k8s-security.json        # K8s security patterns
│
├── supply-chain/
│   ├── licenses/
│   │   ├── spdx-licenses.json       # License database
│   │   ├── compatibility-matrix.json # License compatibility
│   │   └── obligations.json         # License obligations
│   ├── ecosystems/
│   │   ├── npm-patterns.json        # npm-specific patterns
│   │   ├── pypi-patterns.json       # PyPI-specific patterns
│   │   └── registry-apis.md         # Registry API reference
│   └── health/
│       ├── abandonment-signals.json # Abandonment indicators
│       └── typosquat-patterns.json  # Typosquatting detection
│
├── engineering/
│   ├── code-quality/
│   │   ├── code-smells.json         # Code smell catalog
│   │   ├── refactoring-patterns.json # Refactoring catalog
│   │   └── complexity-thresholds.json # Complexity limits
│   ├── testing/
│   │   ├── coverage-guidelines.md   # Coverage best practices
│   │   ├── test-patterns.json       # Test pattern catalog
│   │   └── flaky-test-patterns.json # Flaky test indicators
│   └── performance/
│       ├── complexity-guide.md      # Big O reference
│       ├── antipatterns.json        # Performance antipatterns
│       └── optimization-patterns.json # Optimization techniques
│
├── devops/
│   ├── infrastructure/
│   │   ├── terraform-patterns.json  # Terraform security patterns
│   │   ├── aws-misconfigs.json      # AWS misconfiguration patterns
│   │   └── cis-benchmarks.json      # CIS cloud benchmarks
│   └── cicd/
│       ├── github-actions.json      # GHA security patterns
│       ├── pipeline-patterns.json   # General CI/CD patterns
│       └── secrets-in-ci.json       # CI secrets patterns
│
└── shared/
    ├── severity-levels.json         # Universal severity definitions
    ├── confidence-levels.json       # Confidence level definitions
    └── output-formatting.md         # Output format guidelines
```

## Knowledge Types

### 1. Pattern Databases (JSON)
Machine-readable patterns for detection and classification.

```json
{
  "metadata": {
    "version": "1.0.0",
    "updated": "2025-01-15",
    "category": "security/vulnerabilities"
  },
  "patterns": [
    {
      "id": "PATTERN-001",
      "name": "SQL Injection",
      "description": "...",
      "regex": "...",
      "severity": "critical",
      "cwe": "CWE-89",
      "remediation": "..."
    }
  ]
}
```

### 2. Reference Documentation (Markdown)
Human-readable guidance and methodology.

```markdown
# STRIDE Threat Modeling

## Overview
STRIDE is a threat modeling methodology...

## Categories
### Spoofing
...
```

### 3. Compatibility/Mapping Tables (JSON)
Relationship data between entities.

```json
{
  "license_compatibility": {
    "MIT": {
      "compatible_with": ["Apache-2.0", "BSD-3-Clause", "ISC"],
      "incompatible_with": []
    },
    "GPL-3.0": {
      "compatible_with": ["GPL-2.0", "LGPL-3.0"],
      "incompatible_with": ["Apache-2.0"]
    }
  }
}
```

## Usage Patterns

### Agent Definition Reference
```markdown
## Knowledge Base

This agent uses the following knowledge sources:
- `security/vulnerabilities/cwe-database.json` - CWE patterns
- `security/vulnerabilities/owasp-top-10.json` - OWASP reference
- `security/vulnerabilities/cvss-guide.md` - Scoring guidance

Key patterns are loaded at runtime for detection.
```

### Pattern Matching
```python
# Pseudocode for how agents use patterns
patterns = load_knowledge("security/secrets/secret-patterns.json")
for pattern in patterns:
    matches = grep(codebase, pattern.regex)
    for match in matches:
        report_finding(pattern, match)
```

### Dynamic Enrichment
```markdown
## Analysis Framework

1. Load CWE database from knowledge/security/vulnerabilities/cwe-database.json
2. For each finding, enrich with CWE details
3. Fetch current CVE data via WebFetch (dynamic)
4. Combine static patterns + dynamic data for complete analysis
```

## Knowledge Update Process

### Versioning
Each knowledge file includes metadata with version and update date:
```json
{
  "metadata": {
    "version": "1.2.0",
    "updated": "2025-01-15",
    "source": "OWASP Top 10 2021",
    "maintainer": "security-team"
  }
}
```

### Update Frequency
| Knowledge Type | Update Frequency | Trigger |
|----------------|------------------|---------|
| CWE Database | Quarterly | MITRE releases |
| OWASP Top 10 | On release | OWASP updates |
| Secret Patterns | Monthly | New patterns discovered |
| License Data | Annually | SPDX updates |
| Code Smells | Rarely | Framework changes |
| CIS Benchmarks | Quarterly | Benchmark releases |

### Contribution Process
1. Create PR with updated knowledge file
2. Bump version in metadata
3. Document changes in knowledge CHANGELOG
4. Review for accuracy
5. Merge and tag release

## Quality Standards

### Pattern Quality
- Each pattern must have a unique ID
- Regex patterns must be tested against positive and negative cases
- Severity must align with `shared/severity-levels.json`
- Remediation must be actionable

### Documentation Quality
- Clear, concise explanations
- Examples for complex concepts
- References to authoritative sources
- Regular review for accuracy
