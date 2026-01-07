# Personas Architecture

## Overview

Personas customize how analysis results are presented based on who is consuming them. They define output style, formatting, and prioritization - but do NOT contain factual content (that lives in the Knowledge Base).

**Key distinction:**
- **Agents** (Cereal, Razor, Blade, etc.) = Workers who do the analysis
- **User Personas** (Security Engineer, Auditor, etc.) = Define what the USER cares about

## Core Principle

**Personas define HOW to present information, not WHAT information exists.**

```
                    Analysis Request
                          │
                          ▼
              ┌───────────────────────┐
              │   Persona Selection   │
              │  ┌─────┬─────┬─────┐  │
              │  │Sec  │Soft │Eng  │  │
              │  │Eng  │Eng  │Lead │  │
              │  └─────┴─────┴─────┘  │
              └───────────────────────┘
                          │
          ┌───────────────┴───────────────┐
          ▼                               ▼
┌─────────────────────┐     ┌─────────────────────┐
│  Persona Definition │     │   Knowledge Base    │
│                     │     │                     │
│  • Output style     │     │  • Vulnerability    │
│  • Template         │     │    patterns         │
│  • Prioritization   │     │  • Compliance       │
│                     │     │    frameworks       │
└─────────────────────┘     └─────────────────────┘
          │                               │
          └───────────────┬───────────────┘
                          ▼
              ┌───────────────────────┐
              │ Persona-Tailored      │
              │ Output                │
              └───────────────────────┘
```

## Available Personas

| Persona | Location | Focus |
|---------|----------|-------|
| Security Engineer | `rag/personas/security-engineer.md` | Technical vulnerability analysis, attacker perspective |
| Software Engineer | `rag/personas/software-engineer.md` | Practical remediation, CLI commands, migrations |
| Engineering Leader | `rag/personas/engineering-leader.md` | Portfolio metrics, dashboards, strategic recommendations |
| Auditor | `rag/personas/auditor.md` | Compliance assessment, control testing, evidence |

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

## Output Template
Structured template showing how to format output for this persona

## Prioritization Framework
Role-specific rules for ordering findings

## Key Questions to Answer
What questions this persona needs answered
```

## Agent Overlays

Overlays add agent-specific customizations without duplicating base personas.

**Location:** `rag/personas/overlays/[agent]/[persona]-overlay.md`

**Example:** `rag/personas/overlays/cereal/security-engineer-overlay.md` adds supply-chain-specific knowledge references and examples.

Overlays contain:
- Additional knowledge sources
- Domain-specific examples
- Specialized prioritization rules

See `rag/personas/overlays/README.md` for details.

## How It Works

```
User asks Zero: "Give me a security engineer report on vulnerabilities"

Zero invokes: Task(subagent_type="cereal", persona="security-engineer")

System loads:
  1. agents/cereal/agent.md (agent identity)
  2. rag/personas/security-engineer.md (output style)
  3. rag/personas/overlays/cereal/security-engineer-overlay.md (if exists)

Agent outputs: Security-engineer-formatted report
```

## Persona Selection Logic

When the user's role is unclear, infer from context:

| Question Type | Default Persona |
|---------------|-----------------|
| Technical questions about CVEs/code | `security-engineer` |
| "How do I fix this?" | `software-engineer` |
| "Give me a summary" | `engineering-leader` |
| "Are we compliant?" | `auditor` |

## Adding a New Persona

1. **Create persona file**: `rag/personas/[persona-name].md`
2. **Define the structure**:
   - Role description
   - Output style preferences
   - Output templates
   - Prioritization rules
3. **Create overlays** (optional): `rag/personas/overlays/[agent]/[persona]-overlay.md`
4. **Test** with sample analysis data

**Important**: Do NOT duplicate knowledge content. Knowledge lives in agent knowledge bases; personas only define presentation.

## Directory Structure

```
rag/personas/
├── README.md                    # Persona system documentation
├── security-engineer.md         # Universal personas
├── software-engineer.md
├── engineering-leader.md
├── auditor.md
└── overlays/                    # Agent-specific customizations
    ├── README.md
    └── cereal/
        └── security-engineer-overlay.md
```

## Related Documentation

- [Knowledge Base Architecture](knowledge-base.md) - Where factual content lives
- [System Architecture Overview](overview.md) - How all components fit together
- [Agent Architecture](agents.md) - How agents work
