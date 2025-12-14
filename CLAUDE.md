# Zero - Claude Code Configuration

Zero provides security analysis tools and specialist AI agents for repository assessment.
Named after characters from the movie Hackers (1995) - "Hack the planet!"

## Super Scanner Architecture (v3.5)

Zero uses **8 consolidated super scanners** with configurable features:

| Scanner | Features | Description |
|---------|----------|-------------|
| **sbom** | generation, integrity | SBOM generation (source of truth) |
| **package-analysis** | vulns, health, licenses, malcontent, confusion, typosquats, deprecations, duplicates, reachability, provenance, bundle, recommendations | Package/dependency analysis (depends on sbom) |
| **crypto** | ciphers, keys, random, tls, certificates | Cryptographic security |
| **code-security** | vulns, secrets, api | Security-focused code analysis |
| **code-quality** | tech_debt, complexity, test_coverage, documentation | Code quality metrics |
| **devops** | iac, containers, github_actions, dora, git | DevOps and CI/CD security |
| **tech-id** | detection, models, frameworks, datasets, ai_security, ai_governance, infrastructure | Technology detection and ML-BOM generation |
| **code-ownership** | contributors, bus_factor, codeowners, orphans, churn, patterns | Code ownership analysis |

**Key architecture notes:**
- `sbom` scanner runs first and generates `sbom.cdx.json` (CycloneDX format)
- `package-analysis` scanner depends on sbom output - does not generate its own SBOM
- `tech-id` scanner generates ML-BOM (Machine Learning Bill of Materials)
- Each scanner produces **one JSON output file** with all feature results

## Orchestrator: Zero

**Zero** (named after Zero Cool) is the master orchestrator who coordinates all specialist agents.
Use `/agent` to enter agent mode and chat with Zero directly.

## Specialist Agents

The following agents are available for specialized analysis tasks. Use the Task tool with the appropriate `subagent_type` to invoke them.

| Agent | Persona | Character | Expertise | Primary Scanner |
|-------|---------|-----------|-----------|-----------------|
| `cereal` | Cereal | Cereal Killer | Supply chain, vulnerabilities, malcontent | **sbom**, **package-analysis** |
| `razor` | Razor | Razor | Code security, SAST, secrets detection | **code-security** |
| `blade` | Blade | Blade | Compliance, SOC 2, ISO 27001, audit prep | package-analysis, code-security |
| `phreak` | Phreak | Phantom Phreak | Legal, licenses, data privacy | **package-analysis** (licenses) |
| `acid` | Acid | Acid Burn | Frontend, React, TypeScript, accessibility | **code-security**, **code-quality** |
| `dade` | Dade | Dade Murphy | Backend, APIs, databases, Node.js, Python | **code-security** (api) |
| `nikon` | Nikon | Lord Nikon | Architecture, system design, patterns | **tech-id** |
| `joey` | Joey | Joey | CI/CD, build optimization, caching | **devops** (github_actions) |
| `plague` | Plague | The Plague | DevOps, infrastructure, Kubernetes, IaC | **devops** |
| `gibson` | Gibson | The Gibson | DORA metrics, team health, engineering KPIs | **devops** (dora, git), **code-ownership** |
| `gill` | Gill | Gill Bates | Cryptography, ciphers, keys, TLS, random | **crypto** |
| `turing` | Turing | Alan Turing | AI/ML security, ML-BOM, model safety, LLM security | **tech-id** |

### Agent Details

#### Cereal (Supply Chain Security)
**subagent_type:** `cereal`

Cereal Killer was paranoid about surveillance - perfect for watching for malware hiding in dependencies.
Specializes in dependency vulnerability analysis, malcontent findings investigation (supply chain compromise detection), package health assessment, license compliance, and typosquatting detection.

**Primary scanners:** `sbom`, `package-analysis`
**Required data:** `sbom.json`, `package-analysis.json` (contains vulns, health, malcontent, licenses, etc.)

**Example invocation:**
```
Task tool with subagent_type: "cereal"
prompt: "Investigate the malcontent findings for expressjs/express. Focus on critical and high severity findings."
```

#### Razor (Code Security)
**subagent_type:** `razor`

Razor cuts through code to find vulnerabilities.
Specializes in static analysis, secret detection, code vulnerability assessment, and security code review.

**Primary scanner:** `code-security`
**Required data:** `code-security.json` (contains vulns, secrets, api)

#### Gill (Cryptography Specialist)
**subagent_type:** `gill`

Gill Bates represented the corporate establishment in Hackers - now reformed and using vast crypto knowledge to help secure implementations.
Specializes in cryptographic security analysis, cipher review, key management, TLS configuration, and random number generation security.

**Primary scanner:** `crypto`
**Required data:** `crypto.json` (contains ciphers, keys, random, tls, certificates)

**Example invocation:**
```
Task tool with subagent_type: "gill"
prompt: "Analyze the cryptographic security of this repository. Focus on hardcoded keys and weak ciphers."
```

#### Turing (AI/ML Security Specialist)
**subagent_type:** `turing`

Alan Turing - the father of artificial intelligence and legendary codebreaker. Uses deep understanding of machine learning to secure AI systems against emerging ML supply chain threats.
Specializes in ML model security, ML-BOM generation, AI framework analysis, LLM security, and AI governance.

**Primary scanner:** `tech-id`
**Required data:** `technology.json` (contains models, frameworks, datasets, security, governance)

**Example invocation:**
```
Task tool with subagent_type: "turing"
prompt: "Analyze the AI/ML security of this repository. Check for unsafe pickle models and exposed API keys."
```

#### Plague (DevOps Engineer)
**subagent_type:** `plague`

The Plague controlled all the infrastructure (we reformed him).
Specializes in infrastructure, Kubernetes, IaC security, container security, and deployment automation.

**Primary scanner:** `devops`
**Required data:** `devops.json` (contains iac, containers, github_actions, dora, git)

#### Joey (Build Engineer)
**subagent_type:** `joey`

Joey was learning the ropes - builds things, sometimes breaks them.
Specializes in CI/CD pipelines, build optimization, caching strategies, and build security.

**Primary scanner:** `devops` (github_actions feature)
**Required data:** `devops.json`

#### Gibson (Engineering Leader)
**subagent_type:** `gibson`

The Gibson - the ultimate system that tracks everything.
Specializes in DORA metrics analysis, team health assessment, and engineering KPIs.

**Primary scanners:** `devops` (dora, git features), `code-ownership`
**Required data:** `devops.json`, `code-ownership.json`

#### Nikon (Software Architect)
**subagent_type:** `nikon`

Lord Nikon had photographic memory - sees the big picture.
Specializes in system design, architectural patterns, trade-offs analysis, and design review.

**Primary scanner:** `tech-id`
**Required data:** `technology.json`, `package-analysis.json`

#### Blade (Internal Auditor)
**subagent_type:** `blade`

Blade is meticulous and detail-oriented - perfect for auditing.
Specializes in compliance assessment (SOC 2, ISO 27001), audit preparation, control testing, and policy gap analysis.

**Primary scanners:** Multiple (package-analysis, code-security, devops)
**Required data:** `package-analysis.json`, `code-security.json`, `devops.json`

#### Phreak (General Counsel)
**subagent_type:** `phreak`

Phantom Phreak knew the legal angles and how systems really work.
Specializes in license compatibility analysis, data privacy assessment, and legal risk evaluation.

**Primary scanner:** `package-analysis` (licenses feature)
**Required data:** `package-analysis.json`

#### Acid (Frontend Engineer)
**subagent_type:** `acid`

Acid Burn - sharp, stylish, the elite frontend hacker.
Specializes in React, TypeScript, component architecture, accessibility (a11y), and frontend security.

**Primary scanner:** `code-security`, `code-quality`
**Required data:** `code-security.json`, `code-quality.json`

#### Dade (Backend Engineer)
**subagent_type:** `dade`

Dade Murphy - the person behind Zero Cool, backend systems expert.
Specializes in APIs, databases, Node.js, Python, and backend architecture.

**Primary scanner:** `code-security` (api feature)
**Required data:** `code-security.json`

## Slash Commands

### /agent

Enter agent mode to chat with Zero, the master orchestrator. Zero can delegate to any specialist agent.

### /zero

Master orchestrator for repository analysis. See `.claude/commands/zero.md` for full documentation.

Key commands:
- `./zero.sh hydrate <repo>` - Clone and analyze a repository
- `./zero.sh status` - Show hydrated projects
- `./zero.sh report <repo>` - Generate analysis reports

## Project Structure

```
zero/
├── agents/                    # Agent definitions and knowledge
│   ├── supply-chain/          # Cereal agent
│   │   ├── agent.md           # Agent definition
│   │   ├── knowledge/         # Domain knowledge
│   │   └── prompts/           # Output templates
│   └── shared/                # Shared knowledge (severity, confidence)
├── pkg/scanners/              # Go scanner implementations (8 super scanners)
│   ├── sbom/                  # SBOM super scanner (source of truth)
│   ├── package-analysis/      # Package analysis (depends on sbom)
│   ├── crypto/                # Crypto super scanner
│   ├── code-security/         # Security-focused code analysis
│   ├── code-quality/          # Code quality metrics
│   ├── devops/                # DevOps super scanner
│   ├── tech-id/  # Technology detection and ML-BOM
│   └── code-ownership/        # Code ownership analysis
├── rag/                       # Retrieval-Augmented Generation knowledge
│   └── tech-id/  # Technology detection patterns
├── config/
│   └── zero.config.json       # Scanner configuration
└── .claude/
    ├── agents/                # Claude Code agent definitions
    ├── commands/              # Slash commands
    └── settings.local.json    # Local settings
```

## Data Flow

```
./zero hydrate <repo>
         │
         ├─► Clone repository to .zero/repos/<project>/repo/
         │
         └─► Run super scanners, store JSON in .zero/repos/<project>/analysis/
                  │
                  ├─► sbom.json               (2 features) + sbom.cdx.json
                  ├─► package-analysis.json   (12 features, depends on sbom)
                  ├─► crypto.json             (5 features)
                  ├─► code-security.json      (3 features)
                  ├─► code-quality.json       (4 features)
                  ├─► devops.json             (5 features)
                  ├─► technology.json         (7 features) - ML-BOM
                  └─► code-ownership.json     (6 features)

/agent
         │
         ├─► Zero greets you and asks what to investigate
         │
         └─► Zero delegates to specialists via Task tool
                  │
                  └─► Agents use Read, Grep, WebSearch to investigate
```

## Agent Autonomy

Agents support autonomous investigation with full tool access and agent-to-agent delegation.

### Investigation Mode

When investigation is triggered, agents can autonomously:

1. **Read source files** - Examine flagged code, trace data flows
2. **Search the codebase** - Use Grep/Glob to find patterns, entry points, callers
3. **Research externally** - Use WebSearch to find CVEs, advisories, attacks
4. **Fetch documentation** - Use WebFetch to retrieve security bulletins

**Investigation triggers:**
- Critical/high severity findings
- Suspicious network behavior
- Obfuscated or encrypted code
- Post-install scripts with external calls

### Agent-to-Agent Delegation

Specialists can delegate to other agents when cross-domain expertise is needed:

| Agent | Can Delegate To |
|-------|-----------------|
| Cereal | Phreak (legal), Razor (security), Plague (devops), Nikon (architecture), Gill (crypto) |
| Razor | Cereal (supply chain), Blade (compliance), Nikon (architecture), Dade (backend), Gill (crypto) |
| Blade | Cereal (supply chain), Razor (security), Phreak (legal), Gill (crypto) |
| Acid | Dade (backend), Nikon (architecture), Razor (security) |
| Dade | Acid (frontend), Nikon (architecture), Razor (security), Plague (devops) |
| Nikon | All technical domains |
| Gill | Razor (security), Cereal (supply chain), Plague (devops), Blade (compliance) |

**Delegation example:**
```
Task(subagent_type="phreak", prompt="Analyze license compatibility of MIT and GPL-3.0 in this dependency tree")
```

### Context Loading Modes

Agents receive cached analysis data in three modes:

| Mode | Description | Use Case |
|------|-------------|----------|
| `summary` | Only summary sections | Quick Q&A, status checks |
| `critical` | Only critical/high findings | Triage, urgent issues |
| `full` | Complete data | Deep investigation |

Mode is automatically selected based on query keywords:
- "investigate", "analyze", "trace" → `full` mode
- "critical", "urgent", "priority" → `critical` mode
- Default → `summary` mode

## Configuration Profiles

Profiles define which scanners and features to run:

| Profile | Scanners | Use Case |
|---------|----------|----------|
| `quick` | sbom, package-analysis (limited) | Fast feedback |
| `standard` | sbom, package-analysis, code-security, code-quality | Balanced analysis |
| `security` | sbom, package-analysis, crypto, code-security, devops | Security-focused |
| `full` | All 8 scanners | Complete analysis |
| `sbom-only` | sbom | SBOM generation only |
| `package-analysis-only` | sbom, package-analysis | Dependency analysis only |
| `crypto-only` | crypto | Crypto security only |
| `ai-security` | sbom, package-analysis, code-security, tech-id | AI/ML security with ML-BOM |
| `supply-chain` | sbom, package-analysis, tech-id | Supply chain analysis |
| `compliance` | sbom, package-analysis, tech-id | License/compliance |

## Environment Variables

- `GITHUB_TOKEN` - Required for GitHub API access
- `ANTHROPIC_API_KEY` - Required for Claude-assisted analysis
- `ZERO_HOME` - Override default `.zero/` location (defaults to project root)
