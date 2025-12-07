# User Personas

User personas define HOW analysis results should be presented based on WHO is consuming them. They customize output style, formatting, and prioritization without duplicating factual content (that lives in the Knowledge Base).

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

| Persona | File | Focus |
|---------|------|-------|
| Security Engineer | `security-engineer.md` | Technical vulnerability analysis, CVE triage, attacker perspective |
| Software Engineer | `software-engineer.md` | Practical remediation, CLI commands, migration guides |
| Engineering Leader | `engineering-leader.md` | Portfolio metrics, dashboards, strategic recommendations |
| Auditor | `auditor.md` | Compliance assessment, control testing, evidence documentation |

## Persona Structure

Each persona file contains:

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

Overlays in `overlays/[agent]/` add agent-specific customizations without duplicating the base persona. For example, `overlays/cereal/security-engineer-overlay.md` adds supply-chain-specific knowledge references.

See `overlays/README.md` for details.

## How Personas Work with Agents

```
User asks Zero: "Give me a security engineer report on vulnerabilities"

Zero invokes agent with persona:
  Task(subagent_type="cereal", persona="security-engineer")

System loads:
  1. agents/cereal/agent.md (agent identity & capabilities)
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

1. Create persona file: `rag/personas/[persona-name].md`
2. Follow the standard structure (see existing personas)
3. Create agent-specific overlays if needed in `overlays/[agent]/`
4. Test with sample analysis data

**Important:** Do NOT duplicate knowledge content. Reference knowledge files; don't copy them.
