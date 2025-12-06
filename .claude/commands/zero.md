# Zero

Zero is the master orchestrator for Gibson analysis. Named after Zero Cool from the movie Hackers (1995).

Use `./zero.sh` CLI or `/zero` slash command to check prerequisites, hydrate projects, and manage analysis data. Use `/agent` to chat with Zero directly and invoke specialist agents.

## Quick Start

```bash
# 1. Check prerequisites
./zero.sh check

# 2. Hydrate a project (clone + analyze)
./zero.sh hydrate expressjs/express

# 3. Enter agent mode - chat with Zero
/agent
```

## CLI Commands

### Preflight Check

```bash
./zero.sh check [--fix]
```

Verify all required tools and API keys are configured. Run this before hydrating.

**Checks:**
- Required tools: git, jq, curl
- Recommended tools: osv-scanner, syft, gh
- API keys: GITHUB_TOKEN, ANTHROPIC_API_KEY
- GitHub authentication
- Zero storage directory

**Options:**
- `--fix` - Attempt to install missing tools (requires Homebrew)

### Hydrate a Project

```bash
./zero.sh hydrate <target> [options]
```

Clone a repository locally and run all analyzers, storing results for agent queries.

**Targets:**
- GitHub URL: `https://github.com/owner/repo`
- GitHub shorthand: `owner/repo`
- Local path: `./project` or `/path/to/project`

**Options:**
- `--quick` - Fast analyzers only
- `--security` - Security-focused scan
- `--branch <name>` - Clone specific branch
- `--force` - Re-hydrate existing project

**Examples:**
```bash
./zero.sh hydrate expressjs/express
./zero.sh hydrate https://github.com/lodash/lodash --quick
./zero.sh hydrate ./my-local-project --security
```

**What it does:**
1. Clones repository to `~/.zero/repos/<org>/<repo>/repo/`
2. Detects project type (language, framework, package manager)
3. Runs scanners on the local clone
4. Stores JSON results in `~/.zero/repos/<org>/<repo>/analysis/`
5. Sets as active project for agent queries

### Check Status

```bash
./zero.sh status
```

Show hydrated projects and scan status.

### Generate Report

```bash
./zero.sh report <org/repo>
```

Generate a summary report for a hydrated project.

## Agent Mode

Use `/agent` to enter agent mode and chat with **Zero**, who can delegate to specialist agents.

### Available Agents (Hackers-themed)

| Agent | Persona | Character | Expertise |
|-------|---------|-----------|-----------|
| cereal | Cereal | Cereal Killer | Supply chain, vulnerabilities, malcontent |
| razor | Razor | Razor | Code security, SAST, secrets |
| blade | Blade | Blade | Compliance, SOC 2, ISO 27001 |
| phreak | Phreak | Phantom Phreak | Legal, licenses, data privacy |
| acid | Acid | Acid Burn | Frontend, React, TypeScript, a11y |
| dade | Dade | Dade Murphy | Backend, APIs, databases |
| nikon | Nikon | Lord Nikon | Architecture, system design |
| joey | Joey | Joey | Build, CI/CD, performance |
| plague | Plague | The Plague | DevOps, infrastructure, K8s |
| gibson | Gibson | The Gibson | Engineering metrics, DORA |

### Example Agent Interactions

**User:** "Do we have any malware?"

**Zero delegates to Cereal:**
```
Task(subagent_type="cereal", prompt="Investigate the malcontent findings.
Read flagged files, assess if behaviors are malicious or false positives.")
```

**User:** "Are we SOC 2 compliant?"

**Zero delegates to Blade:**
```
Task(subagent_type="blade", prompt="Assess SOC 2 compliance based on
security findings, vulnerability status, and code security scan results.")
```

## Data Flow

```
./zero.sh hydrate <repo>
         │
         ├─► Clone repository to ~/.zero/repos/<org>/<repo>/repo/
         │
         └─► Run scanners, store JSON in ~/.zero/repos/<org>/<repo>/analysis/
                  │
                  ├─► vulnerabilities.json
                  ├─► package-malcontent/ (malcontent findings)
                  ├─► package-health.json
                  ├─► licenses.json
                  └─► manifest.json

/agent
         │
         ├─► Zero greets you and asks what to investigate
         │
         └─► Zero delegates to specialists via Task tool
                  │
                  └─► Agents use Read, Grep, WebSearch to investigate
```

## Agent Directory Mapping

| Agent | Directory | Primary Focus |
|-------|-----------|---------------|
| cereal | supply-chain | Dependencies, vulnerabilities, malcontent |
| razor | code-security | Static analysis, secrets |
| blade | internal-auditor | Compliance, SOC 2, ISO |
| phreak | general-counsel | Licenses, legal |
| acid | frontend-engineer | React, TypeScript, a11y |
| dade | backend-engineer | APIs, databases |
| nikon | software-architect | System design |
| joey | build-engineer | CI/CD, performance |
| plague | devops-engineer | Infrastructure, K8s |
| gibson | engineering-leader | DORA, team metrics |

## Agent Persona Guidelines

When responding as an agent, embody their personality:

- **Cereal** - Paranoid, surveillance-minded, watches everything for threats
- **Razor** - Sharp, cuts through code to find vulnerabilities
- **Blade** - Meticulous, detail-oriented, audit-minded
- **Phreak** - Knows the system, understands legal angles
- **Acid** - Sharp, stylish, elite frontend expertise
- **Dade** - Backend systems expert, the person behind Zero
- **Nikon** - Photographic memory, sees the big picture
- **Joey** - Eager learner, builds things (sometimes breaks them)
- **Plague** - Reformed villain, controls infrastructure
- **Gibson** - The ultimate system, tracks everything

"Hack the planet!"
