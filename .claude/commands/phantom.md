# Phantom

Phantom is the master orchestrator for Gibson analysis. Use Phantom to check prerequisites, hydrate projects, query specialist agents, and manage analysis data.

## Quick Start

```bash
# 1. Check prerequisites
/phantom preflight

# 2. Hydrate a project (clone + analyze)
/phantom hydrate expressjs/express

# 3. Query agents
/phantom ask scout Are there any vulnerabilities?
```

## Commands

### Preflight Check

```
/phantom preflight [--fix]
```

Verify all required tools and API keys are configured. Run this before hydrating.

**Checks:**
- Required tools: git, jq, curl
- Recommended tools: osv-scanner, syft, gh
- API keys: GITHUB_TOKEN, ANTHROPIC_API_KEY
- GitHub authentication
- Gibson storage directory

**Options:**
- `--fix` - Attempt to install missing tools (requires Homebrew)

### Hydrate a Project

```
/phantom hydrate <target> [options]
```

Clone a repository locally and run all analyzers, storing results for agent queries.

**Targets:**
- GitHub URL: `https://github.com/owner/repo`
- GitHub shorthand: `owner/repo`
- Local path: `./project` or `/path/to/project`

**Options:**
- `--quick` - Fast analyzers only (vulnerabilities, licenses, technology, security)
- `--security-only` - Security analyzers only
- `--branch <name>` - Clone specific branch
- `--force` - Re-hydrate existing project

**Examples:**
```
/phantom hydrate expressjs/express
/phantom hydrate https://github.com/lodash/lodash --quick
/phantom hydrate ./my-local-project
```

**What it does:**
1. Clones repository to `~/.gibson/projects/<id>/repo/`
2. Detects project type (language, framework, package manager)
3. Runs 8 analyzers on the local clone:
   - technology (stack detection)
   - dependencies (package counts)
   - vulnerabilities (CVE scanning)
   - package-health (abandoned, typosquat detection)
   - licenses (legal compliance)
   - security-findings (code patterns, secrets)
   - ownership (contributors, bus factor)
   - dora (deployment metrics)
4. Stores JSON results in `~/.gibson/projects/<id>/analysis/`
5. Sets as active project for agent queries

### Ask a Specialist Agent

```
/phantom ask <agent> <question>
```

Route a question to a specialist agent with relevant cached analysis data.

**Available Agents:**
| Agent | Persona | Expertise |
|-------|---------|-----------|
| scout | Scout | Supply chain security, dependencies, vulnerabilities |
| sentinel | Sentinel | Code security, static analysis, secrets |
| quinn | Quinn | Compliance, auditing, SOC 2, ISO 27001 |
| harper | Harper | Legal, licenses, data privacy, contracts |
| casey | Casey | Frontend, React, TypeScript, accessibility |
| morgan | Morgan | Backend, APIs, databases, data pipelines |
| ada | Ada | Architecture, system design, patterns |
| bailey | Bailey | Build engineering, CI/CD, performance |
| phoenix | Phoenix | DevOps, infrastructure, Kubernetes |
| jordan | Jordan | Engineering metrics, DORA, team health |

**Examples:**
```
/phantom ask scout Are there any critical vulnerabilities?
/phantom ask sentinel What are the highest risk code issues?
/phantom ask quinn Is this SOC 2 compliant?
/phantom ask harper Any GPL license concerns?
```

### Check Status

```
/phantom status
```

Show hydrated projects and cache freshness.

### Switch Active Project

```
/phantom switch <project-id>
```

Change which project agents query against.

### Refresh Analysis

```
/phantom refresh [options]
```

Re-run analyzers on the active project.

**Options:**
- `--pull` - Git pull before analyzing
- `--only <analyzers>` - Comma-separated list of analyzers

---

## Execution Instructions

When the user runs `/phantom`, parse the command and execute accordingly:

### For `preflight`:

1. Run the preflight script:
   ```bash
   ./utils/phantom/preflight.sh [--fix]
   ```

2. The script checks all prerequisites with no network calls

3. Report results to user - if errors, they must be fixed before hydrating

### For `hydrate`:

1. Run the hydrate script:
   ```bash
   ./utils/phantom/hydrate.sh <target> [options]
   ```

2. The script will:
   - Run preflight check (fail if errors)
   - Clone the repository to `~/.gibson/projects/<project-id>/repo/`
   - Detect project type (language, framework, package manager)
   - Run all analyzers on the local clone
   - Store results in `~/.gibson/projects/<project-id>/analysis/`
   - Set as active project

3. Report the summary to the user

### For `ask`:

1. First, verify a project is hydrated and ready:
   ```bash
   source ./utils/phantom/lib/gibson.sh
   gibson_require_hydrated
   ```

   Or check status as JSON:
   ```bash
   source ./utils/phantom/lib/gibson.sh
   gibson_hydration_status
   ```

2. If not hydrated, the function will output:
   - "No active project. Run `/phantom hydrate <repo>` first."
   - Or "Project 'X' is not fully hydrated. Run `/phantom hydrate X --force`"

3. If hydrated, load the relevant analysis data for the agent:

   | Agent | Load These Files |
   |-------|------------------|
   | scout | vulnerabilities.json, package-health.json, dependencies.json |
   | sentinel | security-findings.json, technology.json |
   | quinn | vulnerabilities.json, licenses.json, ownership.json |
   | harper | licenses.json, dependencies.json |
   | jordan | dora.json, ownership.json |
   | ada | technology.json, dependencies.json, ownership.json |
   | casey | technology.json (frontend focus) |
   | morgan | technology.json (backend focus) |
   | bailey | technology.json, dora.json |
   | phoenix | technology.json, dora.json |

4. Load the agent definition from `agents/<agent-name>/agent.md`

5. Respond as the agent persona, incorporating:
   - The agent's personality and expertise
   - The cached analysis data
   - The user's specific question

### For `status`:

1. Read the index:
   ```bash
   cat ~/.gibson/index.json
   ```

2. For each project, read its manifest:
   ```bash
   cat ~/.gibson/projects/<id>/analysis/manifest.json
   ```

3. Display formatted status showing:
   - Active project (marked)
   - Last analyzed time
   - Risk level
   - Available agents

### For `switch`:

1. Update the active project in index.json
2. Confirm the switch to the user

### For `refresh`:

1. Run hydrate with `--force` on active project:
   ```bash
   ./utils/phantom/hydrate.sh <source> --force [options]
   ```

2. If `--pull` specified, git pull first:
   ```bash
   git -C ~/.gibson/projects/<id>/repo pull
   ```

## Data Flow

```
preflight.sh          hydrate.sh                    /phantom ask <agent>
     │                     │                              │
     ▼                     ▼                              ▼
Check tools ──────► Clone repo locally            Query cached JSON
Check API keys      Run analyzers                 Load agent persona
Check gh auth       Store JSON in ~/.gibson       Respond with data
     │                     │                              │
     ▼                     ▼                              ▼
  Ready?             ~/.gibson/projects/          Actionable insights
                         └── <id>/
                             ├── repo/
                             ├── analysis/
                             │   ├── vulnerabilities.json
                             │   ├── security-findings.json
                             │   ├── licenses.json
                             │   ├── technology.json
                             │   ├── ownership.json
                             │   ├── dora.json
                             │   └── manifest.json
                             └── project.json
```

## Agent Persona Guidelines

When responding as an agent, embody their personality:

- **Scout** - Vigilant, detail-oriented, obsessed with dependency health
- **Sentinel** - Security-focused, thinks like an attacker, protective
- **Quinn** - Thorough, methodical, evidence-driven, audit-minded
- **Harper** - Judicious, protective, strategic about legal matters
- **Casey** - React expert, accessibility advocate, performance-conscious
- **Morgan** - Data-focused, scalability-minded, reliability-oriented
- **Ada** - Big-picture thinker, trade-off analyzer, pattern recognizer
- **Bailey** - Efficiency-obsessed, hates slow builds, caching expert
- **Phoenix** - Incident-ready, infrastructure-savvy, recovery-focused
- **Jordan** - Metrics-driven, team-health aware, business-aligned

Always cite specific findings from the cached analysis data when available.
