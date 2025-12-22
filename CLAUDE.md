# Zero - Claude Code Configuration

Zero provides security analysis tools and specialist AI agents for repository assessment.
Named after characters from the movie Hackers (1995) - "Hack the planet!"

## Super Scanner Architecture (v3.6)

Zero uses **9 consolidated super scanners** with configurable features:

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
| **devx** | onboarding, sprawl, workflow | Developer experience analysis (depends on tech-id) |

**Key architecture notes:**
- `sbom` scanner runs first and generates `sbom.cdx.json` (CycloneDX format)
- `package-analysis` scanner depends on sbom output - does not generate its own SBOM
- `tech-id` scanner generates ML-BOM (Machine Learning Bill of Materials)
- `devx` scanner depends on tech-id for technology detection (tool vs technology sprawl)
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
- `./zero.sh status` - Show hydrated projects with freshness indicators
- `./zero.sh report <repo>` - Generate analysis reports

### Automation Commands

Zero includes automation features to keep scan data fresh:

#### zero watch

Watch a directory for file changes and automatically trigger scans:

```bash
zero watch                        # Watch current directory
zero watch /path/to/repo          # Watch specific path
zero watch --debounce 5           # Wait 5 seconds after last change
zero watch --scanners sbom,code-security   # Only run specific scanners
zero watch --profile quick        # Use quick profile
```

#### zero refresh

Refresh repositories with stale scan data:

```bash
zero refresh                      # Refresh all stale repos
zero refresh owner/repo           # Refresh specific repo
zero refresh --force              # Force refresh even if fresh
zero refresh --all                # Refresh all repos (not just stale)
zero refresh --profile security   # Use specific scan profile
```

#### zero feeds

Manage external security data feeds:

```bash
zero feeds sync              # Sync all enabled feeds (Semgrep rules)
zero feeds sync --force      # Force sync even if fresh
zero feeds status            # Show feed sync status
zero feeds rules             # Generate rules from RAG patterns
```

**Note:** Vulnerability data is queried LIVE via OSV.dev during scans, not cached.

### Freshness Tracking

Zero tracks scan freshness with four levels:

| Level | Threshold | Status Indicator |
|-------|-----------|------------------|
| Fresh | < 24 hours | Green ● |
| Stale | 1-7 days | Yellow ● |
| Very Stale | 7-30 days | Red ● |
| Expired | > 30 days | Red ○ |

Run `zero status` to see freshness indicators for all hydrated projects.

## Reports

Zero generates interactive HTML reports using [Evidence](https://evidence.dev).

### Generating Reports

```bash
# Generate and open report (starts HTTP server, press Ctrl+C to stop)
zero report expressjs/express

# Force regenerate
zero report expressjs/express --regenerate

# Generate without opening browser
zero report expressjs/express --open=false

# Start live dev server (hot reload for editing)
zero report expressjs/express --serve
```

**Note:** Reports require HTTP to render properly (JavaScript loads data via fetch). The `zero report` command automatically starts a local HTTP server and opens your browser. Press Ctrl+C to stop the server when done viewing.

### Report Pages

| Page | Description | Data Sources |
|------|-------------|--------------|
| **Executive Summary** | Security posture overview with severity breakdown | All scanners |
| **Security Findings** | Vulnerabilities, secrets, crypto issues | code-security, crypto |
| **Dependencies & SBOM** | Package inventory, license distribution | sbom, package-analysis |
| **Supply Chain** | Malcontent detection, package health | package-analysis |
| **DevOps** | DORA metrics, IaC, GitHub Actions, containers | devops |
| **Code Quality** | Quality metrics, devx, technologies, ownership | code-quality, tech-id, code-ownership, devx |
| **AI/ML Security** | ML models, frameworks, AI security findings | tech-id |

### Report Data Sources

Reports use JavaScript data sources in `reports/template/sources/zero/`:

| Source | Scanner | Description |
|--------|---------|-------------|
| `severity_counts.js` | All | Aggregate severity counts |
| `scanner_summary.js` | All | Per-scanner summary |
| `vulnerabilities.js` | package-analysis, code-security | Combined vulnerabilities |
| `secrets.js` | code-security | Detected secrets |
| `crypto_findings.js` | crypto | Cryptographic issues |
| `licenses.js` | package-analysis | License distribution |
| `malcontent.js` | package-analysis | Supply chain threats |
| `dora_metrics.js` | devops | DORA performance metrics |
| `iac_findings.js` | devops | Infrastructure as Code issues |
| `github_actions_findings.js` | devops | CI/CD security |
| `container_findings.js` | devops | Container security |
| `technologies.js` | tech-id | Detected technologies |
| `contributors.js` | code-ownership | Top contributors |
| `ownership_summary.js` | code-ownership | Bus factor, ownership |
| `code_quality.js` | code-quality | Quality metrics |
| `devx_metrics.js` | devx | Developer experience |
| `ai_security.js` | tech-id | AI/ML security findings |
| `ml_models.js` | tech-id | Detected ML models |

## Docker

Zero is available as a Docker image for consistent, dependency-free execution.

### Quick Start

```bash
# Pull the image
docker pull ghcr.io/crashappsec/zero:latest

# Create alias
alias zero='docker run -v ~/.zero:/home/zero/.zero -e GITHUB_TOKEN ghcr.io/crashappsec/zero'

# Use normally
zero hydrate expressjs/express
zero report expressjs/express
```

### Commands

```bash
# Hydrate (clone + scan)
docker run -v ~/.zero:/home/zero/.zero -e GITHUB_TOKEN ghcr.io/crashappsec/zero hydrate owner/repo

# Generate report
docker run -v ~/.zero:/home/zero/.zero ghcr.io/crashappsec/zero report owner/repo

# Start report server
docker run -v ~/.zero:/home/zero/.zero -p 3000:3000 ghcr.io/crashappsec/zero report owner/repo --serve

# Agent mode (interactive)
docker run -it -v ~/.zero:/home/zero/.zero -e ANTHROPIC_API_KEY ghcr.io/crashappsec/zero agent
```

See `docs/DOCKER.md` for full documentation.

## Project Structure

```
zero/
├── agents/                    # Agent definitions and knowledge
│   ├── supply-chain/          # Cereal agent
│   │   ├── agent.md           # Agent definition
│   │   ├── knowledge/         # Domain knowledge
│   │   └── prompts/           # Output templates
│   └── shared/                # Shared knowledge (severity, confidence)
├── pkg/
│   ├── automation/            # Watch mode and scheduled scanning
│   ├── evidence/              # Evidence.dev report generator
│   ├── feeds/                 # External feed sync (Semgrep rules)
│   ├── findings/              # Standardized finding types
│   ├── freshness/             # Staleness detection and tracking
│   ├── rules/                 # Semgrep rule generation from RAG
│   └── scanners/              # Go scanner implementations (9 super scanners)
│       ├── sbom/              # SBOM super scanner (source of truth)
│       ├── package-analysis/  # Package analysis (depends on sbom)
│       ├── crypto/            # Crypto super scanner
│       ├── code-security/     # Security-focused code analysis
│       ├── code-quality/      # Code quality metrics
│       ├── devops/            # DevOps super scanner
│       ├── tech-id/           # Technology detection and ML-BOM
│       ├── code-ownership/    # Code ownership analysis
│       └── devx/              # Developer experience analysis
├── reports/
│   └── template/              # Evidence report template
│       ├── pages/             # Report pages (index, security, etc.)
│       ├── sources/zero/      # JavaScript data sources
│       ├── package.json       # Evidence dependencies
│       └── evidence.config.yaml
├── rag/                       # Retrieval-Augmented Generation knowledge
│   └── tech-id/               # Technology detection patterns
├── config/
│   └── zero.config.json       # Scanner configuration
├── docs/
│   ├── DOCKER.md              # Docker usage documentation
│   └── EVIDENCE-INTEGRATION-PLAN.md  # Report system design
├── Dockerfile                 # Docker build configuration
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
         ├─► Run super scanners, store JSON in .zero/repos/<project>/analysis/
         │        │
         │        ├─► sbom.json               (2 features) + sbom.cdx.json
         │        ├─► package-analysis.json   (12 features, depends on sbom)
         │        ├─► crypto.json             (5 features)
         │        ├─► code-security.json      (3 features)
         │        ├─► code-quality.json       (4 features)
         │        ├─► devops.json             (5 features)
         │        ├─► technology.json         (7 features) - ML-BOM
         │        ├─► code-ownership.json     (6 features)
         │        └─► devx.json               (3 features)
         │
         └─► Record freshness metadata in .zero/repos/<project>/freshness.json

./zero report <repo>
         │
         ├─► Copy Evidence template to .zero/repos/<project>/.evidence-build/
         │
         ├─► Symlink analysis JSON to sources/zero/data/
         │
         ├─► Run Evidence build (npm run sources && npm run build)
         │
         └─► Output HTML report to .zero/repos/<project>/report/
                  │
                  ├─► index.html           (Executive summary)
                  ├─► security/            (Vulnerabilities, secrets, crypto)
                  ├─► dependencies/        (SBOM, licenses)
                  ├─► supply-chain/        (Malcontent, package health)
                  ├─► devops/              (DORA, IaC, containers)
                  ├─► quality/             (Code quality, ownership)
                  └─► ai-security/         (ML models, AI findings)

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
