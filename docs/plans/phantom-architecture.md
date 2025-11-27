# Phantom: Master Orchestrator Agent Architecture

## Overview

Phantom is the master orchestrator agent - the single entry point for all analysis operations. Phantom bootstraps projects, manages the analysis cache, and routes queries to specialist agents.

## Storage Architecture

All data stored in `~/.phantom/`:

```
~/.phantom/
├── config.json                          # Global settings
├── index.json                           # Quick lookup of all projects
│
└── projects/
    ├── expressjs-express/
    │   ├── project.json                 # Project metadata
    │   ├── repo/                        # Cloned code (git intact)
    │   │   ├── .git/
    │   │   ├── src/
    │   │   └── package.json
    │   └── analysis/                    # Analysis results
    │       ├── manifest.json            # What was analyzed, when
    │       ├── technology.json
    │       ├── dependencies.json
    │       ├── vulnerabilities.json
    │       ├── package-health.json
    │       ├── licenses.json
    │       ├── security-findings.json
    │       ├── ownership.json
    │       └── dora.json
    │
    └── my-org-private-repo/
        ├── project.json
        ├── repo/
        └── analysis/
```

### Design Rationale

| Decision | Choice | Why |
|----------|--------|-----|
| Location | `~/.phantom/` | User-global, survives pwd changes, standard convention |
| Project ID | `owner-repo` slug | Human-readable, unique, filesystem-safe |
| Code + analysis | Same parent | Easy correlation, atomic delete |
| Git preserved | `.git/` intact | Pull updates, switch branches, view history |
| Analysis portable | `analysis/` folder | Future: zip/commit for sharing |

### Schema Definitions

**Global Config** (`~/.phantom/config.json`):
```json
{
  "version": "1.0.0",
  "default_analyzers": ["technology", "dependencies", "vulnerabilities", "licenses"],
  "analyzer_timeout_seconds": 300,
  "github_token_env": "GITHUB_TOKEN",
  "anthropic_key_env": "ANTHROPIC_API_KEY"
}
```

**Project Index** (`~/.phantom/index.json`):
```json
{
  "projects": {
    "expressjs-express": {
      "source": "https://github.com/expressjs/express",
      "created_at": "2025-11-27T10:30:00Z",
      "last_analyzed": "2025-11-27T10:35:00Z",
      "status": "ready"
    },
    "lodash-lodash": {
      "source": "https://github.com/lodash/lodash",
      "created_at": "2025-11-26T14:00:00Z",
      "last_analyzed": "2025-11-26T14:05:00Z",
      "status": "ready"
    }
  },
  "active": "expressjs-express"
}
```

**Project Metadata** (`~/.phantom/projects/<id>/project.json`):
```json
{
  "id": "expressjs-express",
  "source": "https://github.com/expressjs/express",
  "source_type": "github",
  "cloned_at": "2025-11-27T10:30:00Z",
  "branch": "main",
  "commit": "a1b2c3d4e5f6",
  "path": "~/.phantom/projects/expressjs-express/repo",
  "detected_type": {
    "languages": ["javascript"],
    "frameworks": ["express"],
    "package_managers": ["npm"]
  }
}
```

**Analysis Manifest** (`~/.phantom/projects/<id>/analysis/manifest.json`):
```json
{
  "project_id": "expressjs-express",
  "analyzed_commit": "a1b2c3d4e5f6",
  "analyses": {
    "technology": {
      "analyzer": "technology-identification-analyser.sh",
      "version": "1.0.0",
      "started_at": "2025-11-27T10:30:05Z",
      "completed_at": "2025-11-27T10:30:08Z",
      "duration_ms": 2340,
      "status": "complete",
      "output_file": "technology.json"
    },
    "vulnerabilities": {
      "analyzer": "vulnerability-analyser.sh",
      "version": "1.2.0",
      "started_at": "2025-11-27T10:30:05Z",
      "completed_at": "2025-11-27T10:30:20Z",
      "duration_ms": 15230,
      "status": "complete",
      "output_file": "vulnerabilities.json",
      "summary": {
        "critical": 3,
        "high": 7,
        "medium": 12,
        "low": 5
      }
    },
    "security-findings": {
      "analyzer": "code-security-analyser.sh",
      "version": "1.0.0",
      "started_at": "2025-11-27T10:30:21Z",
      "completed_at": "2025-11-27T10:31:45Z",
      "duration_ms": 84000,
      "status": "complete",
      "output_file": "security-findings.json",
      "summary": {
        "critical": 1,
        "high": 4,
        "medium": 7,
        "low": 0
      }
    }
  },
  "summary": {
    "risk_level": "high",
    "total_dependencies": 168,
    "direct_dependencies": 24,
    "total_vulnerabilities": 27,
    "total_security_findings": 12,
    "license_status": "compatible",
    "abandoned_packages": 2
  }
}
```

## Commands

### Bootstrap

Clone and analyze a new project:

```
/phantom bootstrap <target> [options]

Examples:
  /phantom bootstrap https://github.com/expressjs/express
  /phantom bootstrap expressjs/express
  /phantom bootstrap expressjs/express --branch v5.x
  /phantom bootstrap expressjs/express --quick
```

**Options:**
| Flag | Description |
|------|-------------|
| `--branch <name>` | Clone specific branch (default: default branch) |
| `--quick` | Fast analyzers only (skip code-security, dora) |
| `--security-only` | Security analyzers only |
| `--depth <n>` | Shallow clone (default: full history for DORA) |

**Flow:**
```
1. Parse target → derive project ID (expressjs-express)
2. Check if already exists in ~/.phantom/projects/
   - If exists: "Project already bootstrapped. Use /phantom refresh to update."
3. Clone to ~/.phantom/projects/<id>/repo/
4. Detect project type (language, framework, package manager)
5. Run analyzers in parallel where possible
6. Write results to ~/.phantom/projects/<id>/analysis/
7. Update ~/.phantom/index.json
8. Display summary
```

### Refresh

Re-analyze an existing project:

```
/phantom refresh [project-id] [options]

Examples:
  /phantom refresh                              # Refresh active project
  /phantom refresh expressjs-express            # Refresh specific project
  /phantom refresh --pull                       # Git pull then analyze
  /phantom refresh --only vulnerabilities       # Single analyzer
  /phantom refresh --only vulnerabilities,licenses  # Multiple analyzers
```

**Options:**
| Flag | Description |
|------|-------------|
| `--pull` | Git pull before analyzing |
| `--only <analyzers>` | Comma-separated list of analyzers to run |
| `--force` | Re-run even if recent analysis exists |

### Status

Show bootstrapped projects:

```
/phantom status

Output:
┌─────────────────────────────────────────────────────────────────┐
│ PHANTOM STATUS                                                   │
├─────────────────────────────────────────────────────────────────┤
│ Active: expressjs-express                                        │
├─────────────────────────────────────────────────────────────────┤
│ Projects:                                                        │
│                                                                  │
│   expressjs-express (active)                                     │
│   └── Last analyzed: 2 hours ago                                 │
│   └── Risk: HIGH (3 critical vulns)                              │
│   └── Agents: Scout ✓  Sentinel ✓  Quinn ✓  Harper ✓            │
│                                                                  │
│   lodash-lodash                                                  │
│   └── Last analyzed: 1 day ago                                   │
│   └── Risk: MEDIUM                                               │
│   └── Agents: Scout ✓  Sentinel ✓  Quinn ✓  Harper ✓            │
│                                                                  │
├─────────────────────────────────────────────────────────────────┤
│ Storage: ~/.phantom/ (1.2 GB)                                    │
└─────────────────────────────────────────────────────────────────┘
```

### Switch

Change active project:

```
/phantom switch <project-id>

Example:
  /phantom switch lodash-lodash
```

### Ask

Route question to specialist agent:

```
/phantom ask <agent> <question>

Examples:
  /phantom ask scout about the critical vulnerabilities
  /phantom ask sentinel what are the highest risk code issues
  /phantom ask quinn is this SOC 2 compliant
  /phantom ask harper any GPL license concerns
```

**Flow:**
```
1. Check active project is set
2. Load relevant analysis files for the agent
3. Load agent definition from agents/<agent>/agent.md
4. Construct prompt with analysis context
5. Route to agent
6. Return response
```

### Remove

Delete a bootstrapped project:

```
/phantom remove <project-id> [--keep-analysis]

Options:
  --keep-analysis    Delete cloned code but keep analysis results
```

## Analyzer Integration

### Analyzer Output Contract

Each analyzer must output JSON to stdout with this structure:

```json
{
  "analyzer": "vulnerability-analyser",
  "version": "1.2.0",
  "timestamp": "2025-11-27T10:30:20Z",
  "target": "/path/to/repo",
  "status": "complete",
  "summary": {
    // Analyzer-specific summary
  },
  "findings": [
    // Analyzer-specific findings
  ]
}
```

### Analyzer Mapping

| Analysis Type | Script | Output File | Used By |
|--------------|--------|-------------|---------|
| technology | `technology-identification-analyser.sh` | `technology.json` | All agents |
| dependencies | (extracted from package files) | `dependencies.json` | Scout, Quinn |
| vulnerabilities | `vulnerability-analyser.sh` | `vulnerabilities.json` | Scout, Sentinel |
| package-health | `package-health-analyser.sh` | `package-health.json` | Scout |
| provenance | `provenance-analyser.sh` | `provenance.json` | Scout, Quinn |
| licenses | `legal-analyser.sh` | `licenses.json` | Harper, Quinn |
| security-findings | `code-security-analyser.sh` | `security-findings.json` | Sentinel |
| ownership | `ownership-analyser.sh` | `ownership.json` | Jordan, Ada |
| dora | `dora-analyser.sh` | `dora.json` | Jordan |

### Parallel Execution

Bootstrap runs analyzers in parallel groups:

```
Group 1 (fast, no network): ~5s
├── technology-identification
├── dependency-extraction
├── ownership-analysis
└── license-detection

Group 2 (network required): ~20s
├── vulnerability-scan (OSV API)
└── package-health (deps.dev API)

Group 3 (slow, CPU intensive): ~60-120s
├── code-security-analysis
└── dora-metrics (git log processing)
```

Quick mode (`--quick`) skips Group 3.

## Agent Data Access

When Phantom routes to an agent, it loads relevant analysis:

| Agent | Analysis Files Loaded |
|-------|----------------------|
| Scout | vulnerabilities.json, package-health.json, dependencies.json, provenance.json |
| Sentinel | security-findings.json, technology.json |
| Quinn | vulnerabilities.json, licenses.json, provenance.json, ownership.json |
| Harper | licenses.json, dependencies.json |
| Jordan | dora.json, ownership.json |
| Ada | technology.json, dependencies.json, ownership.json |
| Casey | technology.json (frontend focus) |
| Morgan | technology.json (backend focus) |
| Bailey | technology.json, dora.json |
| Phoenix | technology.json, dora.json |

## Bootstrap Output Example

```
$ /phantom bootstrap expressjs/express

🔮 PHANTOM BOOTSTRAP
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Target: https://github.com/expressjs/express
Project ID: expressjs-express

Cloning...                                                    ✓
  Branch: main
  Commit: a1b2c3d

Detecting project type...                                     ✓
  Languages: JavaScript
  Framework: Express.js
  Package Manager: npm

Running analyzers...
  ├── Technology identification                               ✓  3s
  ├── Dependency extraction (168 packages)                    ✓  2s
  ├── Vulnerability scan                                      ✓  15s
  │   └── 3 critical, 7 high, 12 medium, 5 low
  ├── Package health                                          ✓  12s
  │   └── 2 abandoned, 1 typosquat risk
  ├── License analysis                                        ✓  5s
  │   └── All MIT compatible
  ├── Code security                                           ✓  84s
  │   └── 1 critical, 4 high, 7 medium
  ├── Code ownership                                          ✓  3s
  │   └── 67% coverage
  └── DORA metrics                                            ✓  8s
      └── Deployment frequency: Daily

Total time: 132s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SUMMARY
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Risk Level: HIGH

┌────────────────┬────────────────────────────────────────────┐
│ Critical       │ 3 CVEs + 1 code vulnerability              │
│ High           │ 7 CVEs + 4 code vulnerabilities            │
│ Abandoned      │ 2 packages with no updates in 2+ years     │
│ Licenses       │ ✓ All compatible (MIT)                     │
└────────────────┴────────────────────────────────────────────┘

Agents ready:
  Scout     → /phantom ask scout ...
  Sentinel  → /phantom ask sentinel ...
  Quinn     → /phantom ask quinn ...
  Harper    → /phantom ask harper ...

Storage: ~/.phantom/projects/expressjs-express/ (45 MB)
```

## Implementation Plan

### Phase 1: Core Infrastructure
- [ ] Create `~/.phantom/` directory structure on first run
- [ ] Implement project ID generation (slug from URL)
- [ ] Create config.json and index.json schemas
- [ ] Write project.json and manifest.json on bootstrap

### Phase 2: Bootstrap Command
- [ ] Clone handler (git clone with options)
- [ ] Project type detection
- [ ] Parallel analyzer execution
- [ ] Progress output formatting
- [ ] Error handling for partial failures

### Phase 3: Analyzer Standardization
- [ ] Define JSON output contract for all analyzers
- [ ] Update analyzers to write to analysis/ directory
- [ ] Add version tracking to analyzer output
- [ ] Create summary extraction for manifest

### Phase 4: Agent Integration
- [ ] Create Phantom agent definition
- [ ] Implement ask command routing
- [ ] Load analysis context per agent
- [ ] Test agent responses with cached data

### Phase 5: Additional Commands
- [ ] refresh command
- [ ] status command
- [ ] switch command
- [ ] remove command

### Phase 6: Slash Command
- [ ] Create `.claude/commands/phantom.md`
- [ ] Wire up command parsing
- [ ] Test end-to-end flow
