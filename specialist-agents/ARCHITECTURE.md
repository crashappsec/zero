# Specialist Agent Architecture

## Overview

A modular agent architecture for security and software engineering analysis. Agents are autonomous, specialized units invoked via Claude Code's Task tool, each operating within defined guardrails.

## Agent Taxonomy

```
specialist-agents/
├── security/                    # Security-focused agents
│   ├── vulnerability-analyst    # CVE analysis, exploit research
│   ├── threat-modeler          # Attack surface, threat scenarios
│   ├── code-auditor            # SAST-style code review
│   ├── secrets-scanner         # Credential and secret detection
│   ├── container-security      # Container/image security
│   └── compliance-checker      # Security compliance (SOC2, etc.)
│
├── supply-chain/               # Dependency and supply chain
│   ├── dependency-investigator # Package health, alternatives
│   ├── license-auditor         # License compliance
│   ├── sbom-analyst           # SBOM generation and analysis
│   └── provenance-verifier    # SLSA, sigstore verification
│
├── engineering/                # Software engineering agents
│   ├── code-reviewer          # PR review, best practices
│   ├── refactoring-advisor    # Code quality improvements
│   ├── test-strategist        # Test coverage, strategies
│   ├── performance-analyst    # Performance bottlenecks
│   ├── api-designer           # API design review
│   └── documentation-writer   # Technical documentation
│
├── devops/                     # Infrastructure and operations
│   ├── infrastructure-auditor # IaC review (Terraform, etc.)
│   ├── ci-cd-optimizer        # Pipeline optimization
│   ├── observability-advisor  # Logging, metrics, tracing
│   └── incident-responder     # Incident analysis, RCA
│
└── planning/                   # Meta-agents for coordination
    ├── remediation-planner    # Prioritized fix plans
    ├── project-planner        # Feature implementation plans
    └── migration-architect    # System migration strategies
```

## Agent Categories

### 1. Security Agents
Focus: Identifying and analyzing security issues

| Agent | Input | Output | Tools |
|-------|-------|--------|-------|
| vulnerability-analyst | CVE data, code | Risk assessment, exploit analysis | Read, Grep, WebSearch |
| threat-modeler | Architecture, code | Threat model, attack trees | Read, Glob, WebFetch |
| code-auditor | Source code | Security findings, CWE mapping | Read, Grep, Glob |
| secrets-scanner | Repository | Exposed secrets report | Read, Grep, Glob |
| container-security | Dockerfiles, images | Security recommendations | Read, Grep, WebFetch |
| compliance-checker | Code, configs | Compliance gaps | Read, Grep, Glob |

### 2. Supply Chain Agents
Focus: Dependency and build pipeline security

| Agent | Input | Output | Tools |
|-------|-------|--------|-------|
| dependency-investigator | Manifests | Health report, alternatives | Read, WebFetch, Bash |
| license-auditor | Dependencies | License inventory, conflicts | Read, Grep, WebFetch |
| sbom-analyst | Codebase | SBOM, completeness report | Read, Grep, Bash |
| provenance-verifier | Packages | SLSA levels, trust assessment | Read, WebFetch, Bash |

### 3. Engineering Agents
Focus: Code quality and best practices

| Agent | Input | Output | Tools |
|-------|-------|--------|-------|
| code-reviewer | PR/code diff | Review comments, suggestions | Read, Grep, Glob |
| refactoring-advisor | Code | Improvement opportunities | Read, Grep, Glob |
| test-strategist | Code, tests | Coverage gaps, test plans | Read, Grep, Glob |
| performance-analyst | Code, profiles | Bottleneck analysis | Read, Grep, WebFetch |
| api-designer | API specs | Design review, improvements | Read, Grep, WebFetch |
| documentation-writer | Code | Documentation drafts | Read, Grep, Glob |

### 4. DevOps Agents
Focus: Infrastructure and operations

| Agent | Input | Output | Tools |
|-------|-------|--------|-------|
| infrastructure-auditor | IaC files | Security/cost findings | Read, Grep, Glob |
| ci-cd-optimizer | Pipeline configs | Optimization suggestions | Read, Grep, WebFetch |
| observability-advisor | Code, configs | Monitoring recommendations | Read, Grep, Glob |
| incident-responder | Logs, metrics | RCA, remediation steps | Read, Grep, WebFetch |

### 5. Planning Agents
Focus: Coordination and strategic planning

| Agent | Input | Output | Tools |
|-------|-------|--------|-------|
| remediation-planner | Findings | Prioritized fix plans | Read, Grep, WebFetch |
| project-planner | Requirements | Implementation plan | Read, Grep, Glob |
| migration-architect | Codebase | Migration strategy | Read, Grep, WebFetch |

## Agent Definition Structure

Each agent follows a standard definition format:

```markdown
# Agent Name

## Identity
Brief description of the agent's role and expertise.

## Objective
Primary goal in 1-2 sentences.

## Capabilities
- What the agent CAN do
- Specific analysis types
- Output formats

## Guardrails
- What the agent MUST NOT do
- Required behaviors
- Citation requirements

## Tools Available
- Specific tools and restrictions
- Allowed commands/domains

## Knowledge Base
- Domain-specific knowledge
- Reference frameworks
- Best practices

## Analysis Framework
- Step-by-step methodology
- Decision criteria

## Output Requirements
- Required sections
- JSON schema reference

## Examples
- Few-shot examples of good analysis
```

## Invocation Patterns

### Direct Invocation
```
Task tool:
  subagent_type: "security/vulnerability-analyst"
  prompt: "Analyze CVE-2024-1234 in the context of this codebase"
```

### Chained Analysis
```
1. security/code-auditor → findings
2. planning/remediation-planner → fix plans
3. engineering/code-reviewer → PR review
```

### Parallel Analysis
```
Invoke simultaneously:
- security/vulnerability-analyst
- supply-chain/dependency-investigator
- security/secrets-scanner
→ Aggregate results in remediation-planner
```

## Guardrail Levels

### Level 1: Read-Only (Most Restrictive)
- Tools: Read, Grep, Glob only
- Use: Security analysis, auditing
- Agents: code-auditor, secrets-scanner, compliance-checker

### Level 2: Web Access
- Tools: Level 1 + WebFetch, WebSearch
- Use: Research, threat intel
- Agents: vulnerability-analyst, threat-modeler

### Level 3: Limited Commands
- Tools: Level 2 + Bash (allowlisted commands)
- Use: Package queries, version checks
- Agents: dependency-investigator, sbom-analyst

### Level 4: Extended (Future)
- Tools: Level 3 + Write (to specific directories)
- Use: Report generation, SBOM output
- Agents: TBD with explicit user approval

## Output Schemas

All agents produce structured JSON output for:
1. **Consistency**: Predictable format for downstream processing
2. **Validation**: Schema validation for quality assurance
3. **Aggregation**: Combining outputs from multiple agents
4. **Automation**: Integration with CI/CD and ticketing systems

Schema location: `guardrails/output-schemas/<agent-name>.json`

## Knowledge Base Organization

```
knowledge/
├── security/
│   ├── cve-analysis/
│   ├── threat-modeling/
│   ├── secure-coding/
│   └── compliance/
├── supply-chain/
│   ├── package-ecosystems/
│   ├── licensing/
│   └── provenance/
├── engineering/
│   ├── code-quality/
│   ├── testing/
│   └── performance/
└── devops/
    ├── infrastructure/
    ├── ci-cd/
    └── observability/
```

## Integration Points

### CI/CD Integration
```yaml
# Example GitHub Action
- name: Security Analysis
  run: |
    claude-code --agent security/vulnerability-analyst \
      --input scan-results.json \
      --output security-report.json
```

### IDE Integration
```
VSCode Command Palette:
> Gibson: Run Code Auditor
> Gibson: Analyze Dependencies
```

### API Integration
```bash
# Future: REST API
POST /api/v1/agents/security/vulnerability-analyst
{
  "input": { "cve": "CVE-2024-1234", "repo": "owner/repo" },
  "options": { "depth": "thorough" }
}
```

## Versioning

Agents are versioned independently:
- `security/vulnerability-analyst@1.0.0`
- Breaking changes increment major version
- Knowledge base updates increment minor version
- Bug fixes increment patch version

## Development Workflow

1. **Define**: Create agent definition in `definitions/`
2. **Schema**: Add output schema in `guardrails/output-schemas/`
3. **Knowledge**: Add domain knowledge to `knowledge/`
4. **Examples**: Add few-shot examples to `examples/`
5. **Test**: Validate against test cases
6. **Document**: Update catalog and README
