# Zero - Claude Code Configuration

Zero provides security analysis tools and specialist AI agents for repository assessment.
Named after characters from the movie Hackers (1995) - "Hack the planet!"

## Orchestrator: Zero

**Zero** (named after Zero Cool) is the master orchestrator who coordinates all specialist agents.
Use `/agent` to enter agent mode and chat with Zero directly.

## Specialist Agents

The following agents are available for specialized analysis tasks. Use the Task tool with the appropriate `subagent_type` to invoke them.

| Agent | Persona | Character | Expertise | Tools |
|-------|---------|-----------|-----------|-------|
| `cereal` | Cereal | Cereal Killer | Supply chain, vulnerabilities, malcontent, package health | Read, Grep, Glob, WebSearch, WebFetch |
| `razor` | Razor | Razor | Code security, SAST, secrets detection | Read, Grep, Glob, WebSearch |
| `blade` | Blade | Blade | Compliance, SOC 2, ISO 27001, audit prep | Read, Grep, Glob, WebFetch |
| `phreak` | Phreak | Phantom Phreak | Legal, licenses, data privacy, contracts | Read, Grep, WebFetch |
| `acid` | Acid | Acid Burn | Frontend, React, TypeScript, accessibility | Read, Grep, Glob |
| `dade` | Dade | Dade Murphy | Backend, APIs, databases, Node.js, Python | Read, Grep, Glob |
| `nikon` | Nikon | Lord Nikon | Architecture, system design, patterns | Read, Grep, Glob |
| `joey` | Joey | Joey | CI/CD, build optimization, caching | Read, Grep, Glob, Bash |
| `plague` | Plague | The Plague | DevOps, infrastructure, Kubernetes, IaC | Read, Grep, Glob, Bash |
| `gibson` | Gibson | The Gibson | DORA metrics, team health, engineering KPIs | Read, Grep, Glob |

### Agent Details

#### Cereal (Supply Chain Security)
**subagent_type:** `cereal`

Cereal Killer was paranoid about surveillance - perfect for watching for malware hiding in dependencies.
Specializes in dependency vulnerability analysis, malcontent findings investigation (supply chain compromise detection), package health assessment, license compliance, and typosquatting detection.

**Required data:** vulnerabilities, package-health, dependencies, package-malcontent, licenses, package-sbom

**Example invocation:**
```
Task tool with subagent_type: "cereal"
prompt: "Investigate the malcontent findings for expressjs/express. Focus on critical and high severity findings."
```

#### Razor (Code Security)
**subagent_type:** `razor`

Razor cuts through code to find vulnerabilities.
Specializes in static analysis, secret detection, code vulnerability assessment, and security code review.

**Required data:** code-security, code-secrets, technology, secrets-scanner

#### Blade (Internal Auditor)
**subagent_type:** `blade`

Blade is meticulous and detail-oriented - perfect for auditing.
Specializes in compliance assessment (SOC 2, ISO 27001), audit preparation, control testing, and policy gap analysis.

**Required data:** vulnerabilities, licenses, package-sbom, iac-security, code-security

#### Phreak (General Counsel)
**subagent_type:** `phreak`

Phantom Phreak knew the legal angles and how systems really work.
Specializes in license compatibility analysis, data privacy assessment, and legal risk evaluation.

**Required data:** licenses, dependencies, package-sbom

#### Acid (Frontend Engineer)
**subagent_type:** `acid`

Acid Burn - sharp, stylish, the elite frontend hacker.
Specializes in React, TypeScript, component architecture, accessibility (a11y), and frontend security.

**Required data:** technology, code-security

#### Dade (Backend Engineer)
**subagent_type:** `dade`

Dade Murphy - the person behind Zero Cool, backend systems expert.
Specializes in APIs, databases, Node.js, Python, and backend architecture.

**Required data:** technology, code-security

#### Nikon (Software Architect)
**subagent_type:** `nikon`

Lord Nikon had photographic memory - sees the big picture.
Specializes in system design, architectural patterns, trade-offs analysis, and design review.

**Required data:** technology, dependencies, package-sbom

#### Joey (Build Engineer)
**subagent_type:** `joey`

Joey was learning the ropes - builds things, sometimes breaks them.
Specializes in CI/CD pipelines, build optimization, caching strategies, and build security.

**Required data:** technology, dora, code-security

#### Plague (DevOps Engineer)
**subagent_type:** `plague`

The Plague controlled all the infrastructure (we reformed him).
Specializes in infrastructure, Kubernetes, IaC security, and deployment automation.

**Required data:** technology, dora, iac-security

#### Gibson (Engineering Leader)
**subagent_type:** `gibson`

The Gibson - the ultimate system that tracks everything.
Specializes in DORA metrics analysis, team health assessment, and engineering KPIs.

**Required data:** dora, code-ownership, git-insights

## Slash Commands

### /agent

Enter agent mode to chat with Zero, the master orchestrator. Zero can delegate to any specialist agent.

### /zero

Master orchestrator for repository analysis. See `.claude/commands/zero.md` for full documentation.

Key commands:
- `./zero.sh hydrate <repo>` - Clone and analyze a repository
- `./zero.sh status` - Show hydrated projects
- `./zero.sh report <repo>` - Generate analysis reports

### /security-review

AI-powered security code review. See `.claude/commands/security-review.md` for details.

## Project Structure

```
zero/
├── agents/                    # Agent definitions and knowledge
│   ├── supply-chain/          # Cereal agent
│   │   ├── agent.md           # Agent definition
│   │   ├── knowledge/         # Domain knowledge
│   │   └── prompts/           # Output templates
│   └── shared/                # Shared knowledge (severity, confidence)
├── utils/
│   ├── zero/                  # Zero orchestrator
│   │   ├── lib/               # Libraries (zero-lib.sh, agent-loader.sh)
│   │   └── scripts/           # CLI scripts (hydrate, scan, report)
│   └── scanners/              # Individual scanners
│       ├── package-malcontent/
│       ├── vulnerabilities/
│       └── ...
├── rag/                       # Retrieval-Augmented Generation knowledge
└── .claude/
    ├── agents/                # Claude Code agent definitions
    ├── commands/              # Slash commands
    └── settings.local.json    # Local settings
```

## Data Flow

```
./zero.sh hydrate <repo>
         │
         ├─► Clone repository to ~/.zero/repos/<project>/repo/
         │
         └─► Run scanners, store JSON in ~/.zero/repos/<project>/analysis/
                  │
                  ├─► vulnerabilities.json
                  ├─► package-malcontent/ (malcontent findings)
                  ├─► package-health.json
                  ├─► licenses.json
                  └─► ...

/agent
         │
         ├─► Zero greets you and asks what to investigate
         │
         └─► Zero delegates to specialists via Task tool
                  │
                  └─► Agents use Read, Grep, WebSearch to investigate
```

## Environment Variables

- `GITHUB_TOKEN` - Required for GitHub API access
- `ANTHROPIC_API_KEY` - Required for Claude-assisted analysis
- `ZERO_HOME` - Override default `~/.zero/` location
