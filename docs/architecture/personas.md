# Personas Architecture

## Overview

Personas customize how analysis results are presented based on who is consuming them. They define output style, formatting, and prioritization - but do NOT contain factual content (that lives in the [Knowledge Base](knowledge-base.md)).

## Core Principle

**Personas define HOW to present information, not WHAT information exists.**

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Analysis Request                             │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        Persona Selection                             │
│     ┌──────────────┬──────────────┬──────────────┬──────────────┐  │
│     │   Security   │   Software   │  Engineering │    Auditor   │  │
│     │   Engineer   │   Engineer   │    Leader    │              │  │
│     └──────────────┴──────────────┴──────────────┴──────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                    ┌───────────────┴───────────────┐
                    ▼                               ▼
┌─────────────────────────────┐   ┌─────────────────────────────────┐
│      Persona Definition     │   │        Knowledge Base           │
│                             │   │                                 │
│  • Output style/tone        │   │  • Vulnerability patterns       │
│  • Template structure       │   │  • Compliance frameworks        │
│  • Knowledge references     │   │  • Ecosystem patterns           │
│  • Prioritization rules     │   │  • Security metrics             │
└─────────────────────────────┘   └─────────────────────────────────┘
                    │                               │
                    └───────────────┬───────────────┘
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      Persona-Tailored Output                         │
└─────────────────────────────────────────────────────────────────────┘
```

## Available Personas

| Persona | Location | Focus |
|---------|----------|-------|
| Security Engineer | `rag/supply-chain/personas/security-engineer.md` | Technical vulnerability analysis, CVE triage |
| Software Engineer | `rag/supply-chain/personas/software-engineer.md` | Dependency updates, CLI commands, migrations |
| Engineering Leader | `rag/supply-chain/personas/engineering-leader.md` | Portfolio metrics, strategic recommendations |
| Auditor | `rag/supply-chain/personas/auditor.md` | Compliance assessment, control testing |

## Persona File Structure

Each persona definition file contains:

```markdown
# [Role] Persona

## Role Description
Who this persona is and what they need from analysis

## Output Style
- Tone (technical, strategic, formal, practical)
- Detail level (high, medium, low)
- Format preferences (tables, commands, narratives)
- Prioritization approach

## Knowledge Sources
References to specialist-agents/knowledge/ files this persona uses
(This is the ONLY connection to factual content)

## Output Template
Structured template showing how to format output for this persona

## Prioritization Framework
Role-specific rules for ordering findings

## Key Questions to Answer
What questions this persona needs answered
```

## How Personas Reference Knowledge

Personas specify which knowledge files they need:

```markdown
## Knowledge Sources

This persona uses the following knowledge from `specialist-agents/knowledge/`:

### Primary Knowledge
- `security/vulnerabilities/cwe-database.json` - CWE patterns
- `security/vulnerability-scoring.md` - CVSS/EPSS interpretation
- `shared/severity-levels.json` - Severity definitions
```

The analysis system loads these knowledge files and combines them with the persona's output templates to generate tailored results.

## Adding a New Persona

1. **Create persona file**: `rag/supply-chain/personas/[persona-name].md`
2. **Define the structure**:
   - Role description
   - Output style preferences
   - Knowledge source references
   - Output templates
   - Prioritization rules
3. **Register the persona** in the scanner's valid personas list
4. **Test** with sample analysis data

**Important**: Do NOT duplicate knowledge content. If new factual content is needed, add it to `specialist-agents/knowledge/` and reference it from the persona.

## Scanner Integration

The supply chain scanner uses personas via command line:

```bash
# Specify persona directly
./supply-chain-scanner.sh --claude --persona security-engineer /path/to/repo

# Interactive selection
./supply-chain-scanner.sh --claude /path/to/repo
# Prompts user to select persona
```

## Related Documentation

- [Knowledge Base Architecture](knowledge-base.md) - Where factual content lives
- [System Architecture Overview](overview.md) - How all components fit together
