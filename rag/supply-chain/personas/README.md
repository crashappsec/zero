<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
SPDX-License-Identifier: GPL-3.0
-->

# Supply Chain Scanner Personas

The supply chain scanner supports **persona-based analysis** that tailors Claude AI's output based on who is consuming the results.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Analysis Request                             │
│                    (repository + scan data)                          │
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
│    (rag/supply-chain/       │   │   (specialist-agents/knowledge) │
│         personas/)          │   │                                 │
│                             │   │  • Vulnerability patterns       │
│  • Output style/tone        │   │  • Compliance frameworks        │
│  • Template structure       │   │  • Ecosystem patterns           │
│  • Knowledge references     │   │  • Security metrics             │
│  • Prioritization rules     │   │  • CIS benchmarks               │
└─────────────────────────────┘   └─────────────────────────────────┘
                    │                               │
                    └───────────────┬───────────────┘
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                     Claude AI Analysis                               │
│         (Persona context + Knowledge + Scan data)                    │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      Persona-Tailored Output                         │
└─────────────────────────────────────────────────────────────────────┘
```

## Key Principle: No Content Duplication

**Personas define HOW to present information, not WHAT information exists.**

- **Knowledge Base** (`specialist-agents/knowledge/`): Single source of truth for all factual content (vulnerability patterns, compliance frameworks, severity definitions, etc.)
- **Personas** (`rag/supply-chain/personas/`): Define output style, templates, and which knowledge sources to use for each role

This separation ensures:
1. Knowledge updates only need to happen in one place
2. Personas can be added without duplicating content
3. Consistent factual information across all personas

## Available Personas

| Persona | File | Focus |
|---------|------|-------|
| Security Engineer | `security-engineer.md` | Technical vulnerability analysis, CVE triage, remediation |
| Software Engineer | `software-engineer.md` | Dependency updates, migration guides, CLI commands |
| Engineering Leader | `engineering-leader.md` | Portfolio health, metrics, strategic recommendations |
| Auditor | `auditor.md` | Compliance assessment, control testing, finding documentation |

## Persona File Structure

Each persona file contains:

```markdown
# [Role] Persona

## Role Description
Who this persona is and what they need

## Output Style
- Tone (technical, strategic, formal, etc.)
- Detail level (high, medium, low)
- Format preferences (tables, commands, narratives)
- Prioritization approach

## Knowledge Sources
References to specialist-agents/knowledge/ files this persona uses

## Output Template
Structured template for this persona's output format

## Prioritization Framework
Role-specific prioritization rules

## Key Questions to Answer
What questions this persona needs answered
```

## Usage

### Command Line
```bash
./supply-chain-scanner.sh --claude --persona security-engineer /path/to/repo
./supply-chain-scanner.sh --claude --persona software-engineer /path/to/repo
./supply-chain-scanner.sh --claude --persona engineering-leader /path/to/repo
./supply-chain-scanner.sh --claude --persona auditor /path/to/repo
```

### Interactive Selection
```bash
./supply-chain-scanner.sh --claude /path/to/repo

# Prompts:
# Select analysis persona:
# 1) Security Engineer
# 2) Software Engineer
# 3) Engineering Leader
# 4) Auditor
```

## Adding New Personas

1. Create persona definition file: `rag/supply-chain/personas/[persona-name].md`
2. Define output style, templates, and knowledge references
3. Add persona to `VALID_PERSONAS` in `supply-chain-scanner.sh`
4. Update interactive selection menu

**Do NOT duplicate knowledge content.** If new knowledge is needed, add it to `specialist-agents/knowledge/` and reference it from the persona.

## Related Documentation

- [Knowledge Architecture](../../../specialist-agents/knowledge/KNOWLEDGE-ARCHITECTURE.md) - Full knowledge base documentation
- [Output Formatting](../../../specialist-agents/knowledge/shared/output-formatting.md) - Detailed formatting guidelines
