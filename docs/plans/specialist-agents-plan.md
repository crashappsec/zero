# Plan: Specialist Agents for Supply Chain Analysis

## Overview

Transition from output-formatting personas to autonomous specialist agents invoked via Claude Code's Task tool. Each agent operates with full autonomy within defined guardrails, capable of exploring codebases, fetching data, and producing comprehensive analyses.

## Architecture Decision

**Chosen Approach**: Claude Code Subagents (Task tool)
- **Invocation**: `Task` tool with `subagent_type` parameter
- **Autonomy**: Full autonomy with guardrails
- **Complexity**: Medium (prompt engineering focus, no model training)

### Why Agents Over Personas

| Factor | Personas | Specialist Agents |
|--------|----------|-------------------|
| Data gathering | Pre-collected only | Can explore and fetch |
| Depth of analysis | Single pass | Multi-step investigation |
| Tool access | None | Read, Grep, Glob, Bash, WebFetch |
| Adaptability | Fixed output format | Dynamic based on findings |
| Maintenance | Lower | Slightly higher |

## Proposed Agent Architecture

```
specialist-agents/
├── definitions/
│   ├── security-analyst.md          # CVE deep-dive, exploit research
│   ├── dependency-investigator.md   # Package health, alternatives
│   ├── compliance-auditor.md        # License audit, policy checks
│   └── remediation-planner.md       # Prioritized fixes, PR suggestions
├── guardrails/
│   ├── allowed-tools.json           # Per-agent tool permissions
│   ├── forbidden-actions.md         # Universal restrictions
│   └── output-schemas/              # Required output formats
├── knowledge/
│   ├── security/                    # Migrated from rag/supply-chain/personas/security-engineer
│   ├── compliance/                  # Migrated from auditor persona
│   └── remediation/                 # New: fix patterns, upgrade guides
└── examples/
    ├── security-analyst-examples/   # Few-shot examples for each agent
    └── ...
```

## Agent Definitions

### 1. Security Analyst Agent

**Purpose**: Deep vulnerability analysis with exploit context

**Capabilities**:
- Analyze CVE details and CVSS breakdowns
- Research exploit availability (via WebSearch)
- Assess reachability in target codebase
- Correlate with CISA KEV catalog
- Identify attack chains across dependencies

**Guardrails**:
- Cannot modify files
- Cannot execute arbitrary code
- Must cite sources for exploit claims
- Output must include confidence levels

**Tools Allowed**: Read, Grep, Glob, WebSearch, WebFetch

**Output Schema**:
```json
{
  "critical_findings": [...],
  "exploit_analysis": {...},
  "attack_surface_assessment": {...},
  "recommendations": [...],
  "confidence": "high|medium|low",
  "sources": [...]
}
```

### 2. Dependency Investigator Agent

**Purpose**: Package health analysis and alternative recommendations

**Capabilities**:
- Fetch live data from npm/PyPI/Go registries
- Analyze maintainer activity and commit patterns
- Detect abandonment signals
- Research alternative packages
- Compare security track records

**Guardrails**:
- Cannot modify package.json/requirements.txt
- Must verify alternatives exist and are maintained
- Include migration complexity assessment

**Tools Allowed**: Read, Grep, Glob, WebFetch, Bash (npm info, pip show)

**Output Schema**:
```json
{
  "packages_analyzed": [...],
  "health_assessments": {...},
  "recommended_alternatives": [...],
  "migration_guides": [...],
  "staleness_warnings": [...]
}
```

### 3. Compliance Auditor Agent

**Purpose**: License compliance and policy verification

**Capabilities**:
- Analyze license compatibility chains
- Detect license conflicts
- Verify SBOM completeness
- Check against organization policies
- Identify disclosure requirements

**Guardrails**:
- Cannot make legal determinations (flag for legal review)
- Must distinguish permissive vs copyleft
- Include confidence levels on complex cases

**Tools Allowed**: Read, Grep, Glob, WebFetch

**Output Schema**:
```json
{
  "license_inventory": {...},
  "compatibility_issues": [...],
  "policy_violations": [...],
  "disclosure_requirements": [...],
  "legal_review_flags": [...]
}
```

### 4. Remediation Planner Agent

**Purpose**: Prioritized fix plans with implementation guidance

**Capabilities**:
- Prioritize vulnerabilities by risk and effort
- Generate specific fix commands
- Create dependency upgrade paths
- Suggest PR descriptions
- Estimate breaking change risk

**Guardrails**:
- Cannot directly modify files (suggest only)
- Must verify upgrade paths exist
- Include rollback considerations

**Tools Allowed**: Read, Grep, Glob, WebFetch, Bash (version checks)

**Output Schema**:
```json
{
  "priority_matrix": {...},
  "fix_plans": [...],
  "upgrade_paths": {...},
  "breaking_change_risks": [...],
  "suggested_pr_descriptions": [...]
}
```

## Implementation Phases

### Phase 1: Foundation (Week 1)
- [ ] Create `specialist-agents/` directory structure
- [ ] Define guardrails framework (allowed tools, forbidden actions)
- [ ] Migrate relevant RAG content to agent knowledge base
- [ ] Create output schema validators

### Phase 2: First Agent - Security Analyst (Week 2)
- [ ] Write security-analyst.md agent definition
- [ ] Include few-shot examples from real CVE analyses
- [ ] Test against known vulnerable repos
- [ ] Iterate on prompt based on output quality
- [ ] Document invocation pattern

### Phase 3: Remaining Agents (Week 3-4)
- [ ] Dependency Investigator agent
- [ ] Compliance Auditor agent
- [ ] Remediation Planner agent
- [ ] Cross-agent coordination patterns

### Phase 4: Integration (Week 5)
- [ ] Update supply-chain-scanner.sh to invoke agents
- [ ] Add `--agent` flag for specialist mode
- [ ] Create agent orchestration for full analysis
- [ ] Performance optimization (parallel agent execution)

### Phase 5: Validation & Refinement (Week 6)
- [ ] Test suite against diverse repos
- [ ] Collect output quality metrics
- [ ] Refine prompts based on failures
- [ ] Document best practices

## Agent Definition Template

```markdown
# [Agent Name]

## Identity
You are a [role] specialist agent focused on [domain].

## Objective
[Primary goal in 1-2 sentences]

## Capabilities
You can:
- [Capability 1]
- [Capability 2]
- ...

## Guardrails
You must NOT:
- [Restriction 1]
- [Restriction 2]
- ...

## Knowledge Base
[Inline critical knowledge or reference to knowledge files]

## Output Requirements
Your response must include:
1. [Required section 1]
2. [Required section 2]
...

Format your output as:
```json
{schema}
```

## Examples
[2-3 few-shot examples of good analysis]
```

## Migration from Personas

### What Transfers
- Reasoning frameworks (3-phase chain)
- Domain expertise content
- Output structure patterns
- Quality criteria

### What Changes
- Passive → Active (can gather own data)
- Format-only → Full analysis
- Single-pass → Multi-step
- Manual invocation → Task tool invocation

## Success Metrics

1. **Output Quality**: Agent findings match expert manual review (>90%)
2. **Autonomy**: Agents complete analysis without human intervention
3. **Accuracy**: No false positives in critical findings
4. **Completeness**: All relevant issues identified
5. **Actionability**: Recommendations are implementable

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Agent exceeds scope | Strict tool permissions + prompt guardrails |
| Hallucinated vulnerabilities | Require source citations + confidence levels |
| Infinite loops | Timeout limits on Task tool |
| Poor output quality | Few-shot examples + output validation |

## Next Steps

1. Review and approve this plan
2. Create directory structure
3. Start with Security Analyst agent as proof of concept
4. Iterate based on output quality
