# Knowledge Base Architecture

## Overview

Each agent contains its own knowledge base, making agents self-contained and portable. Knowledge is organized into **patterns** (what things are) and **guidance** (what things mean).

## Core Principle: Self-Contained Agents

```
agents/
├── supply-chain/
│   ├── agent.md                    # Agent definition
│   ├── VERSION                     # Semantic version
│   ├── CHANGELOG.md                # Version history
│   └── knowledge/
│       ├── patterns/               # Detection (what things ARE)
│       │   ├── ecosystems/        # npm, pypi patterns
│       │   ├── health/            # abandonment, typosquat
│       │   └── licenses/          # SPDX patterns
│       └── guidance/               # Interpretation (what things MEAN)
│           ├── vulnerability-scoring.md
│           ├── prioritization.md
│           └── remediation-techniques.md
│
├── code-security/
│   └── knowledge/
│       ├── patterns/
│       │   ├── vulnerabilities/   # CWE, OWASP
│       │   ├── secrets/           # Secret detection
│       │   └── devops/            # CI/CD, IaC patterns
│       └── guidance/
│
└── shared/                         # Cross-agent knowledge
    ├── severity-levels.json
    ├── confidence-levels.json
    ├── output-formatting.md
    └── guardrails/                # Safety constraints
```

## Knowledge Types

### 1. Patterns (Detection)

Machine-readable patterns that answer: **"What is this?"**

Location: `knowledge/patterns/`

```json
{
  "metadata": {
    "version": "1.0.0",
    "updated": "2025-01-15",
    "category": "vulnerabilities"
  },
  "patterns": [
    {
      "id": "CWE-89",
      "name": "SQL Injection",
      "regex": "...",
      "severity": "critical",
      "remediation": "Use parameterized queries"
    }
  ]
}
```

**Examples:**
- `ecosystems/npm-patterns.json` - npm package detection
- `vulnerabilities/cwe-database.json` - CWE vulnerability patterns
- `secrets/secret-patterns.json` - Credential detection regex

### 2. Guidance (Interpretation)

Human-readable documentation that answers: **"What does this mean?"**

Location: `knowledge/guidance/`

```markdown
# Vulnerability Scoring

## CVSS Interpretation

| Score | Severity | Response Time |
|-------|----------|---------------|
| 9.0+  | Critical | Immediate     |
| 7.0+  | High     | 7 days        |
...
```

**Examples:**
- `vulnerability-scoring.md` - CVSS/EPSS interpretation
- `prioritization.md` - Risk-based prioritization
- `remediation-techniques.md` - How to fix issues

### 3. Shared Knowledge

Cross-agent definitions ensuring consistency.

Location: `agents/shared/`

| File | Purpose |
|------|---------|
| `severity-levels.json` | Universal severity definitions |
| `confidence-levels.json` | Finding confidence scoring |
| `output-formatting.md` | Output conventions |
| `guardrails/` | Agent safety constraints |

## How Agents Use Knowledge

### Agent Definition References

In `agent.md`:

```markdown
## Knowledge Base

### Patterns (Detection)
- `knowledge/patterns/ecosystems/` - Package ecosystem patterns
- `knowledge/patterns/health/` - Package health signals

### Guidance (Interpretation)
- `knowledge/guidance/vulnerability-scoring.md` - Risk assessment
- `knowledge/guidance/prioritization.md` - Triage guidance

### Shared
- `../shared/severity-levels.json` - Severity definitions
```

### Runtime Loading

Agents load knowledge at analysis time:

```
1. Load agent definition (agent.md)
2. Load relevant patterns from knowledge/patterns/
3. Load interpretation guidance from knowledge/guidance/
4. Load shared definitions from ../shared/
5. Apply to analysis input
6. Generate output using guidance
```

## Adding Knowledge to an Agent

### Adding Patterns

1. Create JSON file in `knowledge/patterns/[category]/`
2. Include metadata (version, updated, category)
3. Follow existing schema conventions
4. Reference in agent.md

### Adding Guidance

1. Create Markdown file in `knowledge/guidance/`
2. Use clear headings and examples
3. Reference in agent.md

### Using Shared Knowledge

Reference shared files with relative path:
```markdown
- `../shared/severity-levels.json`
```

## Quality Standards

### Pattern Files
- Unique IDs for each pattern
- Tested regex against positive/negative cases
- Severity aligned with shared definitions
- Actionable remediation

### Guidance Files
- Clear, concise explanations
- Concrete examples
- References to authoritative sources
- Regular accuracy reviews

## Versioning

Each agent is independently versioned:

- `VERSION` file contains semantic version
- `CHANGELOG.md` documents changes
- Knowledge changes trigger version bumps

| Change Type | Version Bump |
|-------------|--------------|
| Breaking schema change | MAJOR |
| New patterns/guidance | MINOR |
| Pattern updates, fixes | PATCH |

## Related Documentation

- [System Overview](overview.md) - Full architecture
- [Agents README](../../agents/README.md) - Agent catalog
