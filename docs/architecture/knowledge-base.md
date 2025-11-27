# Knowledge Architecture

## Overview

The knowledge system provides structured, reusable information that specialist agents and personas can reference during analysis. This document explains the architecture and how the different components work together.

## Core Principle: Single Source of Truth

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    KNOWLEDGE BASE (Single Source of Truth)               │
│                     specialist-agents/knowledge/                         │
│                                                                         │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐       │
│  │  Security   │ │Supply Chain │ │   DevOps    │ │ Engineering │       │
│  │ Patterns    │ │  Patterns   │ │  Patterns   │ │  Patterns   │       │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘       │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐                       │
│  │ Compliance  │ │Dependencies │ │   Shared    │                       │
│  │ Frameworks  │ │  Guidance   │ │ Definitions │                       │
│  └─────────────┘ └─────────────┘ └─────────────┘                       │
└─────────────────────────────────────────────────────────────────────────┘
                              │
            ┌─────────────────┼─────────────────┐
            ▼                 ▼                 ▼
┌───────────────────┐ ┌───────────────┐ ┌───────────────────┐
│     PERSONAS      │ │    AGENTS     │ │      SKILLS       │
│ (Output styling)  │ │(Analysis logic)│ │(Reusable prompts) │
│                   │ │               │ │                   │
│ rag/supply-chain/ │ │ specialist-   │ │ skills/           │
│ personas/         │ │ agents/       │ │                   │
└───────────────────┘ └───────────────┘ └───────────────────┘
```

**Key Rules:**
1. All factual content (patterns, frameworks, definitions) lives in the knowledge base
2. Personas, agents, and skills REFERENCE knowledge - they don't duplicate it
3. Updates to knowledge automatically apply everywhere it's used

## Directory Structure

```
specialist-agents/knowledge/
├── security/                    # Security-specific knowledge
│   ├── vulnerabilities/
│   │   ├── cwe-database.json   # CWE patterns and remediation
│   │   └── owasp-top-10.json   # OWASP Top 10 reference
│   ├── threats/
│   │   └── mitre-attack.json   # ATT&CK techniques
│   ├── secrets/
│   │   └── secret-patterns.json # Regex patterns for secrets
│   ├── vulnerability-scoring.md # CVSS/EPSS guidance
│   ├── cisa-kev-prioritization.md
│   ├── cve-remediation-workflows.md
│   ├── remediation-techniques.md
│   └── security-metrics.md
│
├── supply-chain/                # Supply chain knowledge
│   ├── ecosystems/
│   │   ├── npm-patterns.json   # npm-specific patterns
│   │   ├── pypi-patterns.json  # PyPI-specific patterns
│   │   └── registry-apis.md    # Registry API reference
│   ├── health/
│   │   ├── abandonment-signals.json
│   │   └── typosquat-patterns.json
│   └── licenses/
│       └── spdx-licenses.json
│
├── compliance/                  # Audit and compliance
│   ├── audit-standards.md
│   ├── compliance-frameworks.md
│   ├── control-testing.md
│   ├── evidence-collection.md
│   └── finding-templates.md
│
├── dependencies/                # Dependency management
│   ├── abandoned-package-detection.md
│   ├── deps-dev-api.md
│   ├── package-management-best-practices.md
│   ├── typosquatting-detection.md
│   └── upgrade-path-patterns.md
│
├── devops/                      # DevOps and infrastructure
│   ├── cicd/
│   │   ├── github-actions.json
│   │   ├── pipeline-patterns.json
│   │   └── secrets-in-ci.json
│   └── infrastructure/
│       ├── terraform-patterns.json
│       ├── aws-misconfigs.json
│       └── cis-benchmarks.json
│
├── engineering/                 # Code quality and performance
│   ├── code-quality/
│   │   └── code-smells.json
│   └── performance/
│       └── antipatterns.json
│
└── shared/                      # Cross-cutting definitions
    ├── severity-levels.json    # Universal severity definitions
    ├── confidence-levels.json  # Confidence scoring
    └── output-formatting.md    # Output format guidelines
```

## Knowledge Types

### 1. Pattern Databases (JSON)

Machine-readable patterns for detection and classification. These are designed to be loaded and queried programmatically.

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

Human-readable guidance and methodology. These provide context and detailed explanations.

```markdown
# STRIDE Threat Modeling

## Overview
STRIDE is a threat modeling methodology...

## Categories
### Spoofing
...
```

### 3. Definitions (JSON)

Standardized definitions used across all analysis types.

```json
{
  "severity_levels": {
    "critical": {
      "level": 5,
      "label": "Critical",
      "cvss_range": "9.0-10.0",
      "sla": {"remediate": "24 hours"}
    }
  }
}
```

## How Personas Use Knowledge

Personas are stored in `rag/supply-chain/personas/` and define:
- **Output style** (tone, detail level, format)
- **Knowledge references** (which files from the knowledge base to use)
- **Templates** (how to structure output)
- **Prioritization rules** (persona-specific ordering)

Example persona structure:
```markdown
# Security Engineer Persona

## Knowledge Sources
- `security/vulnerabilities/cwe-database.json`
- `security/vulnerability-scoring.md`
- `shared/severity-levels.json`

## Output Template
[Structured format for this persona]

## Prioritization
1. Critical + KEV → Immediate
2. Critical + High EPSS → 24 hours
...
```

**Personas NEVER duplicate knowledge content.** They only reference it.

## How Agents Use Knowledge

Specialist agents load relevant knowledge files at analysis time:

```python
# Pseudocode
knowledge = load_knowledge([
    "security/vulnerabilities/cwe-database.json",
    "shared/severity-levels.json"
])

for finding in scan_results:
    cwe_info = knowledge.lookup_cwe(finding.cwe_id)
    severity = knowledge.get_severity(finding.cvss)
    report_finding(finding, cwe_info, severity)
```

## Versioning and Updates

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
| CIS Benchmarks | Quarterly | Benchmark releases |
| Ecosystem Patterns | As needed | Registry changes |

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

## Adding New Knowledge

1. Determine the appropriate category directory
2. Create file following the type conventions (JSON for patterns, MD for docs)
3. Include complete metadata
4. Reference from relevant personas/agents
5. Submit PR for review

**Do not create knowledge that duplicates existing files.** If you need different presentation, create a persona that references the existing knowledge.

## Related Documentation

- [Personas Architecture](personas.md) - How personas use knowledge
- [System Architecture Overview](overview.md) - How all components fit together
- [Output Formatting](../../specialist-agents/knowledge/shared/output-formatting.md) - Output format guidelines
- [Severity Levels](../../specialist-agents/knowledge/shared/severity-levels.json) - Standard severity definitions
