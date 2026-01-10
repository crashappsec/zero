# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Zero is an **engineering intelligence platform** for comprehensive repository analysis, written in Go with a Next.js web frontend. It analyzes code across multiple dimensions including security, quality, supply chain, DevOps, technology stack, and team dynamics.

Zero provides 7 analyzers and 12 specialist AI agents (named after characters from the movie Hackers 1995). While security is one key dimension, Zero covers the full spectrum of engineering intelligence.

## Build and Development Commands

```bash
# Build the CLI
go build -o zero ./cmd/zero

# Run all tests with race detection
go test -v -race ./...

# Run a single test file
go test -v ./pkg/scanner/code-security/...

# Run a specific test
go test -v -run TestSecretDetection ./pkg/scanner/code-security/...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# Lint (uses golangci-lint)
golangci-lint run --timeout=5m

# Verify dependencies
go mod verify

# Web frontend (Next.js)
cd web && npm ci && npm run dev      # Development server
cd web && npm run build              # Production build
cd web && npm run lint               # ESLint
cd web && npm run type-check         # TypeScript check

# MCP server
cd mcp-server && npm ci && npm run build
cd mcp-server && npm run dev         # Development with tsx
```

**Testing notes:**
- Some integration tests skip when external tools aren't installed (Semgrep)
- RAG-based tests skip if rules aren't generated - run `./zero feeds rag` first
- Tests are organized by package under `pkg/`

## CLI Usage

```bash
./zero hydrate owner/repo [profile]  # Clone and analyze repository
./zero onboard owner/repo            # Alias for hydrate
./zero cache owner/repo              # Alias for hydrate
./zero hydrate owner/repo all-quick  # Fast analysis with all analyzers
./zero hydrate myorg --demo          # Analyze org repos, skip large ones
./zero scan owner/repo [profile]     # Re-analyze already cloned repo
./zero analyze owner/repo            # Alias for scan
./zero status                        # Show analyzed projects
./zero serve                         # Start web UI (localhost:3000)
./zero feeds semgrep                 # Sync Semgrep rules
./zero feeds rag                     # Generate rules from RAG patterns
./zero list                          # List available analyzers
./zero checkup                       # Verify setup and external tools
```

## Analyzer Architecture (v4.0)

Zero uses **7 consolidated analyzers** with configurable features. These analyzers produce JSON artifacts that MCP tools can then access:

| Analyzer | Features | Description |
|---------|----------|-------------|
| **code-packages** | generation, integrity, vulns, health, licenses, malcontent, confusion, typosquats, deprecations, duplicates, reachability, provenance, bundle, recommendations | SBOM generation + package/dependency analysis |
| **code-security** | vulns, secrets, api, ciphers, keys, random, tls, certificates | Security-focused code analysis + cryptography |
| **code-quality** | tech_debt, complexity, test_coverage, documentation | Code quality metrics |
| **devops** | iac, containers, github_actions, dora, git | DevOps and CI/CD security |
| **technology-identification** | detection, models, frameworks, datasets, ai_security, ai_governance, infrastructure | Technology detection and ML-BOM generation |
| **code-ownership** | contributors, bus_factor, codeowners, orphans, churn, patterns | Code ownership analysis |
| **devx** | onboarding, sprawl, workflow | Developer experience analysis (depends on technology-identification) |

**Key architecture notes:**
- `code-packages` analyzer generates SBOM internally and produces `sbom.cdx.json` (CycloneDX format) + `code-packages.json`
- `code-security` analyzer includes all crypto features (ciphers, keys, random, tls, certificates)
- `technology-identification` analyzer generates ML-BOM (Machine Learning Bill of Materials)
- `devx` analyzer depends on technology-identification for technology detection (tool vs technology sprawl)
- Each analyzer produces **one JSON output file** with all feature results
- MCP tools provide agent access to analyzer output (not directly callable from agents)

## Code Architecture

### Analyzer Framework

All analyzers implement `pkg/scanner/interface.go:Scanner` (note: interface name pending rename):
- `Name()` - Analyzer identifier
- `Run(ctx, opts)` - Execute analysis, return `*ScanResult`
- `Dependencies()` - Analyzers that must run first (for topological sort)
- `EstimateDuration(fileCount)` - Duration estimate

The `NativeRunner` (`pkg/scanner/runner.go`) executes analyzers in dependency order with parallel execution per level.

### Key Packages

| Package | Purpose |
|---------|---------|
| `cmd/zero` | CLI entry point using Cobra |
| `pkg/scanner/*` | Analyzer implementations (one dir per analyzer) |
| `pkg/core/` | Shared utilities (config, terminal, findings, languages) |
| `pkg/workflow/` | Hydrate, automation, freshness tracking, diff |
| `pkg/api/` | HTTP server with Chi router, WebSocket hub |
| `pkg/storage/sqlite/` | SQLite persistence layer |
| `pkg/mcp/` | Go-based MCP server with 16 tools for agent data access |

### Storage Layout

```
.zero/
└── repos/owner/repo/
    ├── repo/              # Cloned repository
    ├── analysis/          # Analyzer JSON output
    │   ├── sbom.cdx.json
    │   ├── code-packages.json
    │   ├── code-security.json
    │   └── ...
    └── freshness.json     # Analysis metadata
```

### Configuration

Profiles defined in `config/zero.config.json` specify which analyzers/features to run. Each analyzer has features that can be enabled/disabled via `feature_overrides`.

## Orchestrator: Zero

**Zero** (named after Zero Cool) is the master orchestrator who coordinates all specialist agents.
Use `/agent` to enter agent mode and chat with Zero directly.

## Specialist Agents

The following agents are available for specialized engineering intelligence tasks. Use the Task tool with the appropriate `subagent_type` to invoke them.

| Agent | Persona | Character | Expertise | Primary Analyzer |
|-------|---------|-----------|-----------|-----------------|
| `cereal` | Cereal | Cereal Killer | Supply chain, vulnerabilities, malcontent | **code-packages** |
| `razor` | Razor | Razor | Code security, SAST, secrets detection | **code-security** |
| `blade` | Blade | Blade | Compliance, SOC 2, ISO 27001, audit prep | code-packages, code-security |
| `phreak` | Phreak | Phantom Phreak | Legal, licenses, data privacy | **code-packages** (licenses) |
| `acid` | Acid | Acid Burn | Frontend, React, TypeScript, accessibility | **code-security**, **code-quality** |
| `flushot` | Flu Shot | Flu Shot | Backend, APIs, databases, Node.js, Python | **code-security** (api) |
| `nikon` | Nikon | Lord Nikon | Architecture, system design, patterns | **technology-identification** |
| `joey` | Joey | Joey | CI/CD, build optimization, caching | **devops** (github_actions) |
| `plague` | Plague | The Plague | DevOps, infrastructure, Kubernetes, IaC | **devops** |
| `gibson` | Gibson | The Gibson | DORA metrics, team health, engineering KPIs | **devops** (dora, git), **code-ownership** |
| `gill` | Gill | Gill Bates | Cryptography, ciphers, keys, TLS, random | **code-security** (crypto) |
| `hal` | Hal | Hal | AI/ML security, ML-BOM, model safety, LLM security | **technology-identification** |

### Agent Details

#### Cereal (Supply Chain Security)
**subagent_type:** `cereal`

Cereal Killer was paranoid about surveillance - perfect for watching for malware hiding in dependencies.
Specializes in dependency vulnerability analysis, malcontent findings investigation (supply chain compromise detection), package health assessment, license compliance, and typosquatting detection.

**Primary analyzer:** `code-packages`
**Required data:** `code-packages.json` (contains vulns, health, malcontent, licenses, etc.)

**Example invocation:**
```
Task tool with subagent_type: "cereal"
prompt: "Investigate the malcontent findings for expressjs/express. Focus on critical and high severity findings."
```

#### Razor (Code Security)
**subagent_type:** `razor`

Razor cuts through code to find vulnerabilities.
Specializes in static analysis, secret detection, code vulnerability assessment, and security code review.

**Primary analyzer:** `code-security`
**Required data:** `code-security.json` (contains vulns, secrets, api, git_history_security)

#### Gill (Cryptography Specialist)
**subagent_type:** `gill`

Gill Bates represented the corporate establishment in Hackers - now reformed and using vast crypto knowledge to help secure implementations.
Specializes in cryptographic security analysis, cipher review, key management, TLS configuration, and random number generation security.

**Primary analyzer:** `code-security` (crypto features)
**Required data:** `code-security.json` (contains ciphers, keys, random, tls, certificates)

**Example invocation:**
```
Task tool with subagent_type: "gill"
prompt: "Analyze the cryptographic security of this repository. Focus on hardcoded keys and weak ciphers."
```

#### Hal (AI/ML Security Specialist)
**subagent_type:** `hal`

Hal - the elusive hacker who speaks in machine code and sees patterns others miss. Uses deep understanding of machine learning to secure AI systems against emerging ML supply chain threats.
Specializes in ML model security, ML-BOM generation, AI framework analysis, LLM security, and AI governance.

**Primary analyzer:** `technology-identification`
**Required data:** `technology-identification.json` (contains models, frameworks, datasets, security, governance)

**Example invocation:**
```
Task tool with subagent_type: "hal"
prompt: "Analyze the AI/ML security of this repository. Check for unsafe pickle models and exposed API keys."
```

#### Plague (DevOps Engineer)
**subagent_type:** `plague`

The Plague controlled all the infrastructure (we reformed him).
Specializes in infrastructure, Kubernetes, IaC security, container security, and deployment automation.

**Primary analyzer:** `devops`
**Required data:** `devops.json` (contains iac, containers, github_actions, dora, git)

#### Joey (Build Engineer)
**subagent_type:** `joey`

Joey was learning the ropes - builds things, sometimes breaks them.
Specializes in CI/CD pipelines, build optimization, caching strategies, and build security.

**Primary analyzer:** `devops` (github_actions feature)
**Required data:** `devops.json`

#### Gibson (Engineering Leader)
**subagent_type:** `gibson`

The Gibson - the ultimate system that tracks everything.
Specializes in DORA metrics analysis, team health assessment, and engineering KPIs.

**Primary analyzers:** `devops` (dora, git features), `code-ownership`
**Required data:** `devops.json`, `code-ownership.json`

#### Nikon (Software Architect)
**subagent_type:** `nikon`

Lord Nikon had photographic memory - sees the big picture.
Specializes in system design, architectural patterns, trade-offs analysis, and design review.

**Primary analyzer:** `technology-identification`
**Required data:** `technology-identification.json`, `code-packages.json`

#### Blade (Internal Auditor)
**subagent_type:** `blade`

Blade is meticulous and detail-oriented - perfect for auditing.
Specializes in compliance assessment (SOC 2, ISO 27001), audit preparation, control testing, and policy gap analysis.

**Primary analyzers:** Multiple (package-analysis, code-security, devops)
**Required data:** `package-analysis.json`, `code-security.json`, `devops.json`

#### Phreak (General Counsel)
**subagent_type:** `phreak`

Phantom Phreak knew the legal angles and how systems really work.
Specializes in license compatibility analysis, data privacy assessment, and legal risk evaluation.

**Primary analyzer:** `package-analysis` (licenses feature)
**Required data:** `package-analysis.json`

#### Acid (Frontend Engineer)
**subagent_type:** `acid`

Acid Burn - sharp, stylish, the elite frontend hacker.
Specializes in React, TypeScript, component architecture, accessibility (a11y), and frontend security.

**Primary analyzer:** `code-security`, `code-quality`
**Required data:** `code-security.json`, `code-quality.json`

#### Flu Shot (Backend Engineer)
**subagent_type:** `flushot`

Flu Shot - one of the underground hackers from the pool party scene, methodical and reliable.
Specializes in APIs, databases, Node.js, Python, and backend architecture.

**Primary analyzer:** `code-security` (api feature)
**Required data:** `code-security.json`

## Slash Commands

### /agent

Enter agent mode to chat with Zero, the master orchestrator. Zero can delegate to any specialist agent.

### /zero

Master orchestrator for repository analysis. See `.claude/commands/zero.md` for full documentation.

Key commands:
- `./zero hydrate <repo> [profile]` - Clone and analyze a repository
- `./zero status` - Show hydrated projects with freshness indicators
- `./zero report <repo>` - Generate analysis reports

#### zero hydrate

Clone and scan repositories:

```bash
zero hydrate strapi/strapi              # Single repo
zero hydrate strapi/strapi all-quick    # With profile
zero hydrate zero-test-org              # All org repos (default limit: 25)
zero hydrate zero-test-org --limit 10   # Limit to 10 repos
zero hydrate zero-test-org --demo       # Demo mode: skip repos > 50MB
```

**Flags:**
- `--limit N` - Maximum repos to process in org mode (default: 25)
- `--demo` - Demo mode: skip repositories larger than 50MB, automatically fetch replacement repos to maintain the requested count

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

Manage analysis rules and knowledge bases:

```bash
zero feeds rag               # Generate rules from RAG knowledge base
zero feeds rag --force       # Force regenerate even if unchanged
zero feeds semgrep           # Sync Semgrep community rules (SAST)
zero feeds semgrep --force   # Force sync even if fresh
zero feeds status            # Show feed sync status
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

Zero generates markdown reports from analysis data.

### Generating Reports

```bash
# Generate report for a specific analyzer
zero report expressjs/express --analyzer code-security

# Generate aggregated report across all analyzers
zero report expressjs/express

# Generate report for a specific category (6 dimensions)
zero report expressjs/express --category security
zero report expressjs/express --category supply-chain
zero report expressjs/express --category quality
zero report expressjs/express --category devops
zero report expressjs/express --category technology
zero report expressjs/express --category team

# Output to file
zero report expressjs/express --output report.md
```

### Report Categories (6 Dimensions of Engineering Intelligence)

| Category | Analyzers | Content |
|----------|-----------|---------|
| **Security** | code-security | Vulnerabilities, secrets, crypto issues |
| **Supply Chain** | code-packages | Dependencies, licenses, malcontent, package health |
| **Quality** | code-quality | Tech debt, complexity, test coverage |
| **DevOps** | devops | IaC, containers, GitHub Actions, DORA metrics |
| **Technology** | technology-identification | Stack detection, ML-BOM, AI/ML findings |
| **Team** | code-ownership, devx | Bus factor, contributors, onboarding |

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

# Generate markdown report
docker run -v ~/.zero:/home/zero/.zero ghcr.io/crashappsec/zero report owner/repo

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
│   ├── core/                  # Foundation packages
│   │   ├── config/            # Configuration loading
│   │   ├── terminal/          # Terminal output
│   │   ├── status/            # Status display
│   │   ├── findings/          # Finding types
│   │   ├── sarif/             # SARIF export
│   │   ├── languages/         # Language detection
│   │   ├── rag/               # RAG patterns
│   │   ├── rules/             # Semgrep rules
│   │   ├── scoring/           # Health scoring
│   │   ├── github/            # GitHub API client
│   │   ├── liveapi/           # Live API queries (OSV)
│   │   └── feeds/             # Semgrep feed sync
│   ├── scanner/               # Scanner framework + implementations
│   │   ├── interface.go       # Scanner interface
│   │   ├── runner.go          # Scanner runner
│   │   ├── code-packages/     # SBOM + package analysis
│   │   ├── code-security/     # Code security scanner (includes crypto)
│   │   ├── code-quality/      # Code quality scanner
│   │   ├── devops/            # DevOps scanner
│   │   ├── technology-identification/  # Technology detection + ML-BOM
│   │   ├── code-ownership/    # Code ownership scanner
│   │   └── developer-experience/  # Developer experience
│   ├── workflow/              # Workflow management
│   │   ├── hydrate/           # Clone and scan
│   │   ├── automation/        # Watch mode
│   │   ├── freshness/         # Staleness tracking
│   │   └── diff/              # Scan comparison
│   └── mcp/                   # MCP server
├── web/                       # Next.js web UI
│   ├── src/
│   │   ├── app/               # App router pages
│   │   └── components/        # React components
│   └── package.json
├── rag/                       # Retrieval-Augmented Generation knowledge
│   └── technology-identification/  # Technology detection patterns
├── config/
│   └── zero.config.json       # Scanner configuration
├── docs/
│   └── DOCKER.md              # Docker usage documentation
├── Dockerfile                 # Docker build configuration
└── .claude/
    ├── agents/                # Claude Code agent definitions
    ├── commands/              # Slash commands
    └── settings.local.json    # Local settings
```

## Data Flow

```
./zero hydrate <repo>  # (aliases: onboard, cache)
         │
         ├─► Clone repository to .zero/repos/<project>/repo/
         │
         ├─► Run analyzers, store JSON in .zero/repos/<project>/analysis/
         │        │
         │        ├─► code-packages.json       (14 features) + sbom.cdx.json
         │        ├─► code-security.json      (8 features, includes crypto)
         │        ├─► code-quality.json       (4 features)
         │        ├─► devops.json             (5 features)
         │        ├─► technology-identification.json (7 features) - ML-BOM
         │        ├─► code-ownership.json     (6 features)
         │        └─► developer-experience.json (3 features)
         │
         └─► Record freshness metadata in .zero/repos/<project>/freshness.json

./zero report <repo>
         │
         ├─► Read analysis JSON from .zero/repos/<project>/analysis/
         │
         └─► Generate markdown report (stdout or --output file)
                  │
                  ├─► Overview             (Executive summary)
                  ├─► Security             (Vulnerabilities, secrets, crypto)
                  ├─► Supply Chain         (Dependencies, licenses, malcontent)
                  ├─► Quality              (Tech debt, complexity)
                  ├─► DevOps               (DORA, IaC, containers)
                  ├─► Technology           (Stack detection, AI/ML)
                  └─► Team                 (Ownership, bus factor)

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
| Cereal | Phreak (legal), Razor (security), Plague (devops), Nikon (architecture), Gill (code-crypto) |
| Razor | Cereal (supply chain), Blade (compliance), Nikon (architecture), Flu Shot (backend), Gill (code-crypto) |
| Blade | Cereal (supply chain), Razor (security), Phreak (legal), Gill (code-crypto) |
| Acid | Flu Shot (backend), Nikon (architecture), Razor (security) |
| Flu Shot | Acid (frontend), Nikon (architecture), Razor (security), Plague (devops) |
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

Profiles define which analyzers and features to run. Choose based on your use case:

### Use-Case Profiles (Recommended)

| Profile | Use Case | Description |
|---------|----------|-------------|
| `security-focused` | Security | Deep security analysis - vulnerabilities, secrets, crypto, malware |
| `engineering-health` | Engineering | Team productivity - quality, ownership, DORA metrics, DevX |
| `compliance` | Audit | Compliance readiness - licenses, vulnerabilities, IaC controls |
| `supply-chain` | Security | Dependency security - SBOMs, malware, provenance |

### General Profiles

| Profile | Analyzers | Description |
|---------|----------|-------------|
| `all-quick` | All 7 analyzers (limited features) | Fast initial assessment |
| `all-complete` | All 7 analyzers (all features) | Comprehensive analysis |

### Analyzer-Specific Profiles

| Profile | Analyzers | Description |
|---------|----------|-------------|
| `code-packages` | code-packages | SBOM + package analysis |
| `code-security` | code-security | SAST, secrets, and crypto |
| `code-quality` | code-quality | Quality metrics |
| `devops` | devops | IaC, containers, CI/CD |
| `technology-identification` | technology-identification | Technology detection, ML-BOM |
| `code-ownership` | code-ownership | Contributor analysis |
| `developer-experience` | technology-identification, developer-experience | Developer experience |

## Environment Variables

- `GITHUB_TOKEN` - Required for GitHub API access
- `ANTHROPIC_API_KEY` - Required for Claude-assisted analysis
- `ZERO_HOME` - Override default `.zero/` location (defaults to project root)
