# Agents

Self-contained AI agents for software analysis and engineering. Each agent is portable and can be deployed to Claude instances or the Crash Override platform.

All agents are named after characters from the movie **Hackers (1995)** - "Hack the planet!"

## The Team

| Agent | Character | Role | Directory |
|-------|-----------|------|-----------|
| **Zero** | Zero Cool | Master orchestrator | (Claude Code) |
| **Cereal** | Cereal Killer | Supply chain security | [cereal/](cereal/) |
| **Razor** | Razor | Code security | [razor/](razor/) |
| **Blade** | Blade | Compliance auditor | [blade/](blade/) |
| **Phreak** | Phantom Phreak | Legal counsel | [phreak/](phreak/) |
| **Acid** | Acid Burn | Frontend engineer | [acid/](acid/) |
| **Dade** | Dade Murphy | Backend engineer | [dade/](dade/) |
| **Nikon** | Lord Nikon | Software architect | [nikon/](nikon/) |
| **Joey** | Joey | Build engineer | [joey/](joey/) |
| **Plague** | The Plague | DevOps engineer | [plague/](plague/) |
| **Gibson** | The Gibson | Engineering leader | [gibson/](gibson/) |

## Security Agents

### Cereal - Supply Chain Security
**Agent:** [cereal/](cereal/) | **Character:** Cereal Killer

Cereal Killer was paranoid about surveillance - perfect for watching for malware hiding in dependencies. Cereal hunts down hidden risks in your software supply chain, obsessing over package health, vulnerability exposure, and license compliance.

Cereal cares about: CVEs and vulnerability severity, malcontent (supply chain compromise), abandoned packages, typosquatting risks, license compatibility, SBOM completeness.

*"Ask Cereal to scan dependencies" or "Cereal, is this package safe?"*

---

### Razor - Code Security
**Agent:** [razor/](razor/) | **Character:** Razor

Razor cuts through code to find vulnerabilities. Stands guard over your codebase, watching for security issues before they reach production. Thinks like an attacker to find weaknesses—injection flaws, hardcoded secrets, insecure configurations.

Razor cares about: OWASP Top 10, CWE classifications, secret exposure, input validation, authentication flaws, secure coding practices.

*"Ask Razor to review security" or "Razor, find vulnerabilities in this code"*

---

### Blade - Compliance Auditor
**Agent:** [blade/](blade/) | **Character:** Blade

Blade is meticulous and detail-oriented - perfect for auditing. Ensures your systems meet compliance requirements for SOC 2, ISO 27001, and other standards.

Blade cares about: Control testing, evidence collection, audit preparation, compliance gaps, policy enforcement.

*"Ask Blade about SOC 2 compliance" or "Blade, are we audit-ready?"*

---

### Phreak - Legal Counsel
**Agent:** [phreak/](phreak/) | **Character:** Phantom Phreak

Phantom Phreak knew the legal angles and how systems really work. Analyzes license compatibility, data privacy requirements, and legal risks.

Phreak cares about: License compatibility, GPL concerns, data privacy, contract terms, legal risk assessment.

*"Ask Phreak about license conflicts" or "Phreak, can we use this library?"*

---

## Engineering Agents

### Acid - Frontend Engineer
**Agent:** [acid/](acid/) | **Character:** Acid Burn

Acid Burn - sharp, stylish, the elite frontend hacker. Lives and breathes component architecture, TypeScript, and modern frontend patterns.

Acid cares about: React best practices, bundle size, Core Web Vitals, WCAG accessibility, state management patterns.

*"Ask Acid about this component" or "Acid, review this React code"*

---

### Dade - Backend Engineer
**Agent:** [dade/](dade/) | **Character:** Dade Murphy

Dade Murphy - the person behind Zero Cool, backend systems expert. Builds the backbone of your application—APIs, databases, and data pipelines.

Dade cares about: REST/GraphQL design, database optimization, data pipelines, caching strategies, error handling.

*"Ask Dade about the API design" or "Dade, optimize this query"*

---

### Nikon - Software Architect
**Agent:** [nikon/](nikon/) | **Character:** Lord Nikon

Lord Nikon had photographic memory - sees the big picture. Designs systems that balance competing concerns—scalability vs. simplicity, security vs. usability.

Nikon cares about: System design patterns, service boundaries, data architecture, technical debt, architectural decision records.

*"Ask Nikon to review the architecture" or "Nikon, what pattern should we use?"*

---

### Joey - Build Engineer
**Agent:** [joey/](joey/) | **Character:** Joey

Joey was learning the ropes - builds things, sometimes breaks them. Makes your builds fast and reliable, hates waiting for CI.

Joey cares about: Build speed, CI/CD optimization, caching strategies, test parallelization, pipeline costs.

*"Ask Joey to speed up the build" or "Joey, why is CI so slow?"*

---

### Plague - DevOps Engineer
**Agent:** [plague/](plague/) | **Character:** The Plague

The Plague controlled all the infrastructure (we reformed him). Orchestrates the journey from code to production and keeps systems running.

Plague cares about: Deployment strategies, infrastructure as code, Kubernetes, observability, disaster recovery.

*"Ask Plague about the deployment" or "Plague, help with this Terraform"*

---

### Gibson - Engineering Leader
**Agent:** [gibson/](gibson/) | **Character:** The Gibson

The Gibson - the ultimate system that tracks everything. Sees engineering through a business lens, tracking the metrics that matter.

Gibson cares about: DORA metrics, developer experience, cloud costs, team effectiveness, engineering productivity.

*"Ask Gibson about the metrics" or "Gibson, how's our DORA score?"*

---

## Shared Resources

| Resource | Description |
|----------|-------------|
| [shared/](shared/) | Cross-agent knowledge (severity, confidence, formatting) |

## Agent Architecture

Each agent is self-contained:

```
cereal/                    # Directory named after agent
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
| `agent.md` | Defines agent identity, capabilities, behavior | "You are Cereal, a supply chain security analyst..." |
| `patterns/` | Detection signatures, regex, file patterns | npm package patterns, secret regexes |
| `guidance/` | Interpretation frameworks, scoring, remediation | CVSS interpretation, remediation workflows |
| `prompts/` | Role-specific output formatting | Security engineer vs auditor output |

### Patterns vs Guidance

- **Patterns** answer: "What is this?" (detection/identification)
- **Guidance** answers: "What does this mean?" (interpretation/action)

## Invoking Agents

### Via Zero (Agent Mode)

Use `/agent` in Claude Code to chat with Zero, who delegates to specialists:

```
You: Do we have any malware in our codebase?
Zero: Let me delegate to Cereal to investigate...
```

### Via Task Tool

Directly invoke agents using the Task tool:

```
Task tool with:
- subagent_type: "cereal"
- prompt: "Investigate the malcontent findings for expressjs/express"
```

### Agent-to-Data Mapping

| Agent | Required Scanner Data |
|-------|----------------------|
| Cereal | vulnerabilities, package-health, package-malcontent, licenses, package-sbom |
| Razor | code-security, code-secrets, technology, secrets-scanner |
| Blade | vulnerabilities, licenses, package-sbom, iac-security, code-security |
| Phreak | licenses, dependencies, package-sbom |
| Acid | technology, code-security |
| Dade | technology, code-security |
| Nikon | technology, dependencies, package-sbom |
| Joey | technology, dora, code-security |
| Plague | technology, dora, iac-security |
| Gibson | dora, code-ownership, git-insights |

## Versioning

Each agent is independently versioned using [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking changes to agent behavior or knowledge schema
- **MINOR**: New capabilities, patterns, or guidance
- **PATCH**: Bug fixes, pattern updates, documentation

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

4. Register in `utils/zero/lib/agent-loader.sh`

5. Create `VERSION` (start at `0.1.0`) and `CHANGELOG.md`

---

*"Hack the planet!"*
