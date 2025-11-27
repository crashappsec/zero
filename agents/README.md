# Agents

Self-contained AI agents for software analysis and engineering. Each agent is portable and can be deployed to Claude instances or the Crash Override platform.

Each agent has a human persona name for easy invocation: "Ask Scout about dependencies" or "Casey, review this component"

## Security Agents

### Scout - Supply Chain Security
**Agent:** [supply-chain](supply-chain/) | **Version:** 0.1.0

Scout is your vigilant dependency analyst who hunts down hidden risks in your software supply chain. Scout obsesses over package health, vulnerability exposure, and license compliance. When you add a new dependency, Scout wants to know: Is it maintained? Is it safe? Will it cause legal issues?

Scout cares about: CVEs and vulnerability severity, abandoned packages, typosquatting risks, license compatibility, SBOM completeness, and upgrade paths.

*"Ask Scout to scan dependencies" or "Scout, is this package safe?"*

---

### Sentinel - Code Security
**Agent:** [code-security](code-security/) | **Version:** 0.1.0

Sentinel stands guard over your codebase, watching for security vulnerabilities before they reach production. Sentinel thinks like an attacker to find weaknesses in your code—injection flaws, hardcoded secrets, insecure configurations, and dangerous patterns.

Sentinel cares about: OWASP Top 10, CWE classifications, secret exposure, input validation, authentication flaws, and secure coding practices.

*"Ask Sentinel to review security" or "Sentinel, find vulnerabilities in this code"*

---

## Engineering Agents

### Casey - Frontend Engineer
**Agent:** [frontend-engineer](frontend-engineer/) | **Version:** 0.1.0

Casey is your React expert who lives and breathes component architecture, TypeScript, and modern frontend patterns. Casey obsesses over render performance, accessibility, and developer experience. When reviewing frontend code, Casey asks: Is this component reusable? Is it accessible? Will it perform well?

Casey cares about: React best practices, bundle size, Core Web Vitals, WCAG accessibility, state management patterns, and testing strategies.

*"Ask Casey about this component" or "Casey, review this React code"*

---

### Morgan - Backend Engineer
**Agent:** [backend-engineer](backend-engineer/) | **Version:** 0.1.0

Morgan builds the backbone of your application—APIs, databases, and data pipelines. Morgan thinks in terms of data flows, query optimization, and system reliability. When designing backend systems, Morgan asks: Will this scale? Is the data model right? Are we handling errors properly?

Morgan cares about: REST/GraphQL design, database optimization, data pipelines, caching strategies, observability, and error handling.

*"Ask Morgan about the API design" or "Morgan, optimize this query"*

---

### Ada - Software Architect
**Agent:** [software-architect](software-architect/) | **Version:** 0.1.0

Named after Ada Lovelace, Ada sees the big picture. Ada designs systems that balance competing concerns—scalability vs. simplicity, security vs. usability, speed vs. cost. When evaluating architecture, Ada asks: What are the trade-offs? Will this decision age well? What happens when we 10x?

Ada cares about: System design patterns, service boundaries, authentication architecture, data architecture, technical debt, and architectural decision records (ADRs).

*"Ask Ada to review the architecture" or "Ada, what pattern should we use?"*

---

### Bailey - Build Engineer
**Agent:** [build-engineer](build-engineer/) | **Version:** 0.1.0

Bailey makes your builds fast and reliable. Bailey hates waiting for CI and despises flaky tests. When looking at your pipeline, Bailey asks: Why does this take so long? What can we cache? Where are the bottlenecks?

Bailey cares about: Build speed, CI/CD optimization, caching strategies, test parallelization, flaky test detection, and pipeline costs.

*"Ask Bailey to speed up the build" or "Bailey, why is CI so slow?"*

---

### Phoenix - DevOps Engineer
**Agent:** [devops-engineer](devops-engineer/) | **Version:** 0.1.0

Phoenix rises from incidents stronger. Phoenix orchestrates the journey from code to production and keeps systems running when things go wrong. When evaluating infrastructure, Phoenix asks: Can we deploy safely? Will we know when it breaks? How fast can we recover?

Phoenix cares about: Deployment strategies, infrastructure as code, Kubernetes, observability, incident response, and disaster recovery.

*"Ask Phoenix about the deployment" or "Phoenix, help with this Terraform"*

---

### Jordan - Engineering Leader
**Agent:** [engineering-leader](engineering-leader/) | **Version:** 0.1.0

Jordan sees engineering through a business lens. Jordan tracks the metrics that matter—delivery velocity, team health, and engineering costs. When evaluating engineering operations, Jordan asks: Are we shipping effectively? Is the team happy? Are we spending wisely?

Jordan cares about: DORA metrics, developer experience, cloud costs, team effectiveness, engineering productivity, and technical investment decisions.

*"Ask Jordan about the metrics" or "Jordan, how's our DORA score?"*

---

## Shared Resources

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
