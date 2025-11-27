# Agents

Self-contained AI agents for software analysis and engineering. Each agent is portable and can be deployed to Claude instances or the Crash Override platform.

## Available Agents

### Security Agents

| Agent | Version | Description |
|-------|---------|-------------|
| [supply-chain](supply-chain/) | 0.1.0 | Supply chain security, vulnerabilities, licenses |
| [code-security](code-security/) | 0.1.0 | Static analysis, secrets, code vulnerabilities |

### Engineering Agents

| Agent | Version | Description |
|-------|---------|-------------|
| [frontend-engineer](frontend-engineer/) | 0.1.0 | React, TypeScript, web app development |
| [backend-engineer](backend-engineer/) | 0.1.0 | APIs, databases, data engineering |
| [architect](architect/) | 0.1.0 | System design, patterns, auth frameworks |
| [build-engineer](build-engineer/) | 0.1.0 | CI/CD optimization, build performance |
| [devops-engineer](devops-engineer/) | 0.1.0 | Deployments, infrastructure, operations |
| [engineering-leader](engineering-leader/) | 0.1.0 | Costs, metrics, team effectiveness |

### Shared Resources

| Resource | Description |
|----------|-------------|
| [shared/](shared/) | Cross-agent knowledge (severity, confidence, formatting) |

## Agent Architecture

Each agent is self-contained:

```
agent-name/
├── agent.md               # Agent definition and behavior
├── VERSION                # Semantic version
├── CHANGELOG.md           # Version history
├── knowledge/
│   ├── patterns/          # Detection patterns (what things ARE)
│   └── guidance/          # Analysis guidance (what things MEAN)
└── prompts/               # Role-specific output prompts (optional)
```

### Key Concepts

| Component | Purpose | Example |
|-----------|---------|---------|
| `agent.md` | Defines agent identity, capabilities, behavior | "You are a supply chain security analyst..." |
| `patterns/` | Detection signatures, regex, file patterns | npm package patterns, secret regexes |
| `guidance/` | Interpretation frameworks, scoring, remediation | CVSS interpretation, remediation workflows |
| `prompts/` | Role-specific output formatting | Security engineer vs auditor output |

### Patterns vs Guidance

- **Patterns** answer: "What is this?" (detection/identification)
- **Guidance** answers: "What does this mean?" (interpretation/action)

## Deployment

### To Claude Instance

Copy the agent directory to your Claude environment:

```bash
cp -r agents/supply-chain/ /path/to/claude/agents/
```

### To Crash Override Platform

Agents are packaged as-is. The platform reads:
1. `agent.md` for agent definition
2. `knowledge/` for context
3. `prompts/` for output customization

## Versioning

Each agent is independently versioned using [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking changes to agent behavior or knowledge schema
- **MINOR**: New capabilities, patterns, or guidance
- **PATCH**: Bug fixes, pattern updates, documentation

Version is stored in `VERSION` file and changes documented in `CHANGELOG.md`.

## Creating a New Agent

1. Create directory structure:
   ```bash
   mkdir -p agents/new-agent/{knowledge/{patterns,guidance},prompts}
   ```

2. Create `agent.md` with:
   - Agent identity and purpose
   - Capabilities and limitations
   - Knowledge references
   - Default behavior

3. Add knowledge:
   - `patterns/` - Detection patterns as JSON
   - `guidance/` - Interpretation docs as Markdown

4. Add prompts for different output styles (optional)

5. Create `VERSION` (start at `0.1.0`) and `CHANGELOG.md`

## Documentation

- [System Architecture](../docs/architecture/overview.md)
- [Knowledge Base Architecture](../docs/architecture/knowledge-base.md)
