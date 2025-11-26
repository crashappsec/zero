# Specialist Agents for Supply Chain Analysis

Autonomous specialist agents invoked via Claude Code's Task tool. Each agent operates with full autonomy within defined guardrails to provide deep analysis of supply chain security issues.

## Quick Start

Invoke an agent using Claude Code's Task tool:

```
Use the Task tool with subagent_type="security-analyst" and provide the vulnerability scan results.
```

## Available Agents

| Agent | Purpose | Primary Tools |
|-------|---------|---------------|
| **security-analyst** | CVE deep-dive, exploit research, attack chain analysis | Read, Grep, WebSearch |
| **dependency-investigator** | Package health, abandonment detection, alternatives | Read, WebFetch, Bash |
| **compliance-auditor** | License audit, policy verification, SBOM assessment | Read, Grep, WebFetch |
| **remediation-planner** | Prioritized fix plans, upgrade paths, PR suggestions | Read, Grep, Bash |

## Directory Structure

```
specialist-agents/
├── definitions/           # Agent prompt definitions
│   ├── security-analyst.md
│   ├── dependency-investigator.md
│   ├── compliance-auditor.md
│   └── remediation-planner.md
├── guardrails/           # Safety and output constraints
│   ├── allowed-tools.json
│   ├── forbidden-actions.md
│   └── output-schemas/   # JSON schemas for each agent
├── knowledge/            # Domain-specific knowledge bases
│   ├── security/         # CVE, CVSS, exploit knowledge
│   ├── dependencies/     # Package health, typosquatting
│   ├── compliance/       # License, audit frameworks
│   └── remediation/      # Upgrade paths, fix patterns
└── examples/             # Few-shot examples (per agent)
```

## Agent Capabilities

### Security Analyst
- Analyze CVE details and CVSS breakdowns
- Research exploit availability via web search
- Assess reachability in target codebase
- Correlate with CISA KEV catalog
- Identify attack chains across dependencies

### Dependency Investigator
- Fetch live data from package registries
- Detect abandoned and deprecated packages
- Identify typosquatting attempts
- Research superior alternatives
- Assess migration complexity

### Compliance Auditor
- Analyze license compatibility chains
- Detect copyleft infection risks
- Verify SBOM completeness
- Check against organization policies
- Identify disclosure requirements

### Remediation Planner
- Prioritize by risk × effort matrix
- Generate specific fix commands
- Identify safe upgrade paths
- Suggest logical PR groupings
- Include rollback procedures

## Guardrails

All agents operate under strict guardrails defined in `guardrails/`:

### Universal Restrictions
- **No file modifications**: Agents cannot write, edit, or delete files
- **No code execution**: No arbitrary commands that change state
- **No credential access**: Cannot read secrets or tokens
- **Source citations required**: All claims must reference evidence

### Per-Agent Permissions
Tool permissions are defined in `allowed-tools.json` and vary by agent. For example:
- Security Analyst: Read, Grep, Glob, WebSearch, WebFetch
- Dependency Investigator: Read, Grep, Glob, WebFetch, Bash (read-only commands)

## Output Schemas

Each agent has a defined JSON output schema in `guardrails/output-schemas/`. Agents are instructed to format their responses according to these schemas for consistent, parseable output.

## Knowledge Base

The `knowledge/` directory contains domain-specific documentation that agents reference during analysis. This content is derived from the broader RAG knowledge base but focused on each agent's specialization.

## Integration

To integrate with the supply-chain-scanner:

```bash
# Future: supply-chain-scanner.sh --agent security-analyst
# Currently: Invoke via Claude Code Task tool
```

## Development

### Adding a New Agent

1. Create definition in `definitions/<agent-name>.md`
2. Add tool permissions to `guardrails/allowed-tools.json`
3. Create output schema in `guardrails/output-schemas/<agent-name>.json`
4. Add relevant knowledge to `knowledge/<domain>/`
5. Add examples to `examples/<agent-name>/`

### Testing Agents

Test agents against known vulnerable repositories:
- OWASP Juice Shop
- Damn Vulnerable Web Application
- Known CVE test cases

## Comparison: Agents vs Personas

| Aspect | Personas (Previous) | Specialist Agents |
|--------|---------------------|-------------------|
| Invocation | Flag on scanner | Task tool |
| Autonomy | Format output only | Full analysis capability |
| Data gathering | Pre-collected | Can explore codebase |
| Tool access | None | Read, Grep, WebFetch, etc. |
| Depth | Single pass | Multi-step investigation |

## License

Copyright (c) 2025 Crash Override Inc.
SPDX-License-Identifier: GPL-3.0
