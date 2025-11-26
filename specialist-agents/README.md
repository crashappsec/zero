# Specialist Agents

Autonomous specialist agents for security and software engineering analysis. Each agent operates with full autonomy within defined guardrails, invoked via Claude Code's Task tool.

## Overview

This framework provides 14 specialist agents across 5 categories:
- **Security**: Vulnerability analysis, threat modeling, code auditing, secrets scanning
- **Supply Chain**: Dependency health, license compliance
- **Engineering**: Code review, refactoring, testing, performance
- **DevOps**: Infrastructure auditing, CI/CD optimization
- **Planning**: Remediation and fix planning

See [CATALOG.md](./CATALOG.md) for detailed agent documentation.
See [ARCHITECTURE.md](./ARCHITECTURE.md) for system design.

## Quick Start

Invoke an agent using Claude Code's Task tool:

```
Task: security/vulnerability-analyst
Prompt: "Analyze CVE-2024-1234 in the context of this codebase..."
```

## Agent Catalog

### Security Agents
| Agent | Purpose | Guardrails |
|-------|---------|------------|
| `security/vulnerability-analyst` | CVE analysis, exploit research | Read, Grep, WebSearch |
| `security/threat-modeler` | STRIDE analysis, attack trees | Read, Grep, WebSearch |
| `security/code-auditor` | SAST-style security review | Read, Grep only |
| `security/secrets-scanner` | Credential detection | Read, Grep only |
| `security/container-security` | Docker/K8s security | Read, Grep, WebFetch |

### Supply Chain Agents
| Agent | Purpose | Guardrails |
|-------|---------|------------|
| `supply-chain/dependency-investigator` | Package health, alternatives | Read, WebFetch, Bash |
| `supply-chain/license-auditor` | License compliance | Read, Grep, WebFetch |

### Engineering Agents
| Agent | Purpose | Guardrails |
|-------|---------|------------|
| `engineering/code-reviewer` | PR review, best practices | Read, Grep only |
| `engineering/refactoring-advisor` | Tech debt, improvements | Read, Grep only |
| `engineering/test-strategist` | Coverage gaps, strategies | Read, Grep only |
| `engineering/performance-analyst` | Bottlenecks, optimization | Read, Grep, WebFetch |

### DevOps Agents
| Agent | Purpose | Guardrails |
|-------|---------|------------|
| `devops/infrastructure-auditor` | IaC security, cost | Read, Grep, WebFetch |
| `devops/ci-cd-optimizer` | Pipeline optimization | Read, Grep, WebFetch |

### Planning Agents
| Agent | Purpose | Guardrails |
|-------|---------|------------|
| `planning/remediation-planner` | Fix prioritization | Read, Grep, WebFetch, Bash |

## Directory Structure

```
specialist-agents/
├── definitions/              # Agent prompt definitions
│   ├── security/            # Security agents
│   ├── supply-chain/        # Supply chain agents
│   ├── engineering/         # Engineering agents
│   ├── devops/              # DevOps agents
│   └── planning/            # Planning agents
├── guardrails/              # Safety constraints
│   ├── allowed-tools.json   # Per-agent tool permissions
│   ├── forbidden-actions.md # Universal restrictions
│   └── output-schemas/      # JSON output schemas
├── knowledge/               # Domain knowledge
│   ├── security/
│   ├── supply-chain/
│   ├── engineering/
│   └── devops/
├── examples/                # Few-shot examples
├── CATALOG.md              # Complete agent catalog
└── ARCHITECTURE.md         # System architecture
```

## Guardrail Levels

| Level | Tools | Use Case |
|-------|-------|----------|
| **Level 1** | Read, Grep, Glob | Security audits (no network) |
| **Level 2** | Level 1 + WebFetch, WebSearch | Research, threat intel |
| **Level 3** | Level 2 + Bash (allowlisted) | Package queries |

### Universal Restrictions
All agents **MUST NOT**:
- Modify any files
- Execute arbitrary commands
- Access credentials or secrets
- Make claims without evidence

All agents **MUST**:
- Cite sources for security claims
- Include confidence levels
- Note assumptions and limitations

## Chaining Agents

### Security Assessment
```
1. security/secrets-scanner
2. security/code-auditor
3. security/container-security
4. planning/remediation-planner
```

### Full Code Review
```
1. engineering/code-reviewer
2. security/code-auditor
3. engineering/test-strategist
4. engineering/performance-analyst
```

### Supply Chain Analysis
```
1. supply-chain/dependency-investigator
2. security/vulnerability-analyst
3. supply-chain/license-auditor
4. planning/remediation-planner
```

## Adding New Agents

1. Create definition in `definitions/<category>/<agent-name>.md`
2. Add tool permissions to `guardrails/allowed-tools.json`
3. Create output schema in `guardrails/output-schemas/<category>/<agent-name>.json`
4. Add knowledge to `knowledge/<domain>/`
5. Add examples to `examples/<category>/`
6. Update CATALOG.md

## Testing Agents

Test against known repositories:
- OWASP Juice Shop (security)
- Damn Vulnerable Web Application (security)
- Known CVE test cases (vulnerability analysis)
- Open source projects (engineering review)

## License

Copyright (c) 2025 Crash Override Inc.
SPDX-License-Identifier: GPL-3.0
