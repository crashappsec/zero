<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
SPDX-License-Identifier: GPL-3.0
-->

# Supply Chain Scanner Personas

This directory contains persona definitions for the supply chain scanner.

## Documentation

For complete documentation on the persona system, see:
- **[Personas Architecture](../../../docs/architecture/personas.md)** - How personas work
- **[Knowledge Base Architecture](../../../docs/architecture/knowledge-base.md)** - Where factual content lives
- **[System Overview](../../../docs/architecture/overview.md)** - How all components fit together

## Available Personas

| File | Role | Focus |
|------|------|-------|
| `security-engineer.md` | Security Engineer | Technical vulnerability analysis, CVE triage |
| `software-engineer.md` | Software Engineer | Dependency updates, CLI commands, migrations |
| `engineering-leader.md` | Engineering Leader | Portfolio metrics, strategic recommendations |
| `auditor.md` | Auditor | Compliance assessment, control testing |

## Quick Start

```bash
# Use a specific persona
./supply-chain-scanner.sh --claude --persona security-engineer /path/to/repo

# Interactive persona selection
./supply-chain-scanner.sh --claude /path/to/repo
```

## Key Principle

**Personas define HOW to present information, not WHAT information exists.**

- Factual content (patterns, frameworks, definitions) lives in `specialist-agents/knowledge/`
- Personas reference that knowledge and define output formatting
- This prevents duplication and ensures consistency
