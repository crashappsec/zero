# Zero: Master Orchestrator Agent Architecture

> Named after Zero Cool from the movie Hackers (1995) - "Hack the planet!"

## Overview

Zero is the master orchestrator agent - the single entry point for all analysis operations. Zero bootstraps projects, manages the analysis cache, and routes queries to specialist agents.

## Storage Architecture

All data stored in `~/.zero/`:

```
~/.zero/
├── config.json                          # Global settings
├── index.json                           # Quick lookup of all projects
│
└── repos/
    ├── expressjs/
    │   └── express/
    │       ├── project.json             # Project metadata
    │       ├── repo/                    # Cloned code (git intact)
    │       │   ├── .git/
    │       │   ├── src/
    │       │   └── package.json
    │       └── analysis/                # Analysis results
    │           ├── manifest.json        # What was analyzed, when
    │           ├── technology.json
    │           ├── dependencies.json
    │           ├── vulnerabilities/
    │           ├── package-health/
    │           ├── package-malcontent/
    │           ├── licenses/
    │           ├── code-security/
    │           ├── ownership/
    │           └── dora/
    │
    └── lodash/
        └── lodash/
            ├── project.json
            ├── repo/
            └── analysis/
```

### Design Rationale

| Decision | Choice | Why |
|----------|--------|-----|
| Location | `~/.zero/` | User-global, survives pwd changes, standard convention |
| Project ID | `owner/repo` path | Human-readable, unique, filesystem-safe |
| Code + analysis | Same parent | Easy correlation, atomic delete |
| Git preserved | `.git/` intact | Pull updates, switch branches, view history |
| Analysis portable | `analysis/` folder | Future: zip/commit for sharing |

### Schema Definitions

**Global Config** (`~/.zero/config.json`):
```json
{
  "version": "1.0.0",
  "default_analyzers": ["technology", "dependencies", "vulnerabilities", "licenses"],
  "analyzer_timeout_seconds": 300,
  "github_token_env": "GITHUB_TOKEN",
  "anthropic_key_env": "ANTHROPIC_API_KEY"
}
```

**Project Index** (`~/.zero/index.json`):
```json
{
  "projects": {
    "expressjs/express": {
      "source": "https://github.com/expressjs/express",
      "created_at": "2025-11-27T10:30:00Z",
      "last_analyzed": "2025-11-27T10:35:00Z",
      "status": "ready"
    },
    "lodash/lodash": {
      "source": "https://github.com/lodash/lodash",
      "created_at": "2025-11-26T14:00:00Z",
      "last_analyzed": "2025-11-26T14:05:00Z",
      "status": "ready"
    }
  },
  "active": "expressjs/express"
}
```

**Project Metadata** (`~/.zero/repos/<org>/<repo>/project.json`):
```json
{
  "id": "expressjs/express",
  "source": "https://github.com/expressjs/express",
  "source_type": "github",
  "cloned_at": "2025-11-27T10:30:00Z",
  "branch": "main",
  "commit": "a1b2c3d4e5f6",
  "path": "~/.zero/repos/expressjs/express/repo",
  "detected_type": {
    "languages": ["javascript"],
    "frameworks": ["express"],
    "package_managers": ["npm"]
  }
}
```

**Analysis Manifest** (`~/.zero/repos/<org>/<repo>/analysis/manifest.json`):
```json
{
  "project_id": "expressjs/express",
  "analyzed_commit": "a1b2c3d4e5f6",
  "analyses": {
    "technology": {
      "analyzer": "tech-discovery.sh",
      "version": "1.0.0",
      "started_at": "2025-11-27T10:30:05Z",
      "completed_at": "2025-11-27T10:30:08Z",
      "duration_ms": 2340,
      "status": "complete",
      "output_file": "technology.json"
    },
    "vulnerabilities": {
      "analyzer": "vulnerabilities.sh",
      "version": "1.2.0",
      "started_at": "2025-11-27T10:30:05Z",
      "completed_at": "2025-11-27T10:30:20Z",
      "duration_ms": 15230,
      "status": "complete",
      "output_dir": "vulnerabilities/",
      "summary": {
        "critical": 3,
        "high": 7,
        "medium": 12,
        "low": 5
      }
    }
  }
}
```

## Commands

### Hydrate

Clone and analyze a new project:

```bash
./zero.sh hydrate <target> [options]

Examples:
  ./zero.sh hydrate https://github.com/expressjs/express
  ./zero.sh hydrate expressjs/express
  ./zero.sh hydrate expressjs/express --branch v5.x
  ./zero.sh hydrate expressjs/express --quick
```

**Options:**
| Flag | Description |
|------|-------------|
| `--branch <name>` | Clone specific branch (default: default branch) |
| `--quick` | Fast analyzers only (~30s) |
| `--standard` | Default analysis profile (~2min) |
| `--security` | Security-focused scan (~3min) |
| `--advanced` | All analyzers (~5min) |
| `--deep` | Claude-assisted analysis (~10min) |
| `--force` | Re-hydrate existing project |

**Flow:**
```
1. Parse target → derive project path (expressjs/express)
2. Check if already exists in ~/.zero/repos/
   - If exists: "Project already hydrated. Use --force to re-hydrate."
3. Clone to ~/.zero/repos/<org>/<repo>/repo/
4. Detect project type (language, framework, package manager)
5. Run scanners based on profile
6. Write results to ~/.zero/repos/<org>/<repo>/analysis/
7. Update ~/.zero/index.json
8. Display summary
```

### Status

Show hydrated projects:

```bash
./zero.sh status

Output:
┌─────────────────────────────────────────────────────────────────┐
│ ZERO STATUS                                                      │
├─────────────────────────────────────────────────────────────────┤
│ Active: expressjs/express                                        │
├─────────────────────────────────────────────────────────────────┤
│ Projects:                                                        │
│                                                                  │
│   expressjs/express (active)                                     │
│   └── Last analyzed: 2 hours ago                                 │
│   └── Risk: HIGH (3 critical vulns)                              │
│   └── Agents: Cereal ✓  Razor ✓  Blade ✓  Phreak ✓              │
│                                                                  │
│   lodash/lodash                                                  │
│   └── Last analyzed: 1 day ago                                   │
│   └── Risk: MEDIUM                                               │
│   └── Agents: Cereal ✓  Razor ✓  Blade ✓  Phreak ✓              │
│                                                                  │
├─────────────────────────────────────────────────────────────────┤
│ Storage: ~/.zero/ (1.2 GB)                                       │
└─────────────────────────────────────────────────────────────────┘
```

### Report

Generate a summary report:

```bash
./zero.sh report <org/repo>

Example:
  ./zero.sh report expressjs/express
```

### Agent Mode

Enter agent mode to chat with Zero:

```bash
# In Claude Code
/agent
```

Zero will delegate to specialist agents based on the query.

## Agent Team (Hackers-Themed)

| Agent | Character | Expertise | Data Required |
|-------|-----------|-----------|---------------|
| **Cereal** | Cereal Killer | Supply chain security | vulnerabilities, package-health, package-malcontent, licenses |
| **Razor** | Razor | Code security | code-security, secrets-scanner, technology |
| **Blade** | Blade | Compliance auditing | vulnerabilities, licenses, code-security, iac-security |
| **Phreak** | Phantom Phreak | Legal counsel | licenses, dependencies |
| **Acid** | Acid Burn | Frontend engineering | technology, code-security |
| **Dade** | Dade Murphy | Backend engineering | technology, code-security |
| **Nikon** | Lord Nikon | Software architecture | technology, dependencies |
| **Joey** | Joey | Build engineering | technology, dora |
| **Plague** | The Plague | DevOps engineering | technology, dora, iac-security |
| **Gibson** | The Gibson | Engineering metrics | dora, code-ownership, git-insights |

## Scanner Integration

### Scanner Mapping

| Scanner | Output | Used By |
|---------|--------|---------|
| `tech-discovery` | `technology.json` | All agents |
| `vulnerabilities` | `vulnerabilities/` | Cereal, Blade |
| `package-malcontent` | `package-malcontent/` | Cereal |
| `package-health` | `package-health/` | Cereal |
| `licenses` | `licenses/` | Phreak, Blade |
| `code-security` | `code-security/` | Razor, Blade |
| `secrets-scanner` | `secrets-scanner/` | Razor |
| `package-sbom` | `sbom.cdx.json` | Cereal, Blade |
| `dora` | `dora/` | Gibson, Joey, Plague |
| `code-ownership` | `code-ownership/` | Gibson |

### Analysis Profiles

| Profile | Time | Scanners |
|---------|------|----------|
| **quick** | ~30s | tech-discovery, vulnerabilities, licenses |
| **standard** | ~2min | + package-health, code-ownership, dora |
| **security** | ~3min | vulnerabilities, package-malcontent, code-security, secrets-scanner |
| **advanced** | ~5min | All scanners |
| **deep** | ~10min | All scanners + Claude-assisted analysis |

## Hydrate Output Example

```
$ ./zero.sh hydrate expressjs/express

███████╗███████╗██████╗  ██████╗
╚══███╔╝██╔════╝██╔══██╗██╔═══██╗
  ███╔╝ █████╗  ██████╔╝██║   ██║
 ███╔╝  ██╔══╝  ██╔══██╗██║   ██║
███████╗███████╗██║  ██║╚██████╔╝
╚══════╝╚══════╝╚═╝  ╚═╝ ╚═════╝
crashoverride.com

Target: https://github.com/expressjs/express
Project: expressjs/express

Cloning...                                                    ✓
  Branch: main
  Commit: a1b2c3d

Detecting project type...                                     ✓
  Languages: JavaScript
  Framework: Express.js
  Package Manager: npm

Running scanners...
  ├── Technology identification                               ✓  3s
  ├── Dependency extraction (168 packages)                    ✓  2s
  ├── Vulnerability scan                                      ✓  15s
  │   └── 3 critical, 7 high, 12 medium, 5 low
  ├── Package health                                          ✓  12s
  │   └── 2 abandoned, 1 typosquat risk
  ├── Package malcontent                                      ✓  45s
  │   └── 3 high-risk behaviors flagged
  ├── License analysis                                        ✓  5s
  │   └── All MIT compatible
  ├── Code security                                           ✓  84s
  │   └── 1 critical, 4 high, 7 medium
  └── Code ownership                                          ✓  3s
      └── 67% coverage

Total time: 132s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SUMMARY
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Risk Level: HIGH

┌────────────────┬────────────────────────────────────────────┐
│ Critical       │ 3 CVEs + 1 code vulnerability              │
│ High           │ 7 CVEs + 4 code vulnerabilities            │
│ Abandoned      │ 2 packages with no updates in 2+ years     │
│ Malcontent     │ 3 behaviors flagged (investigate)          │
│ Licenses       │ ✓ All compatible (MIT)                     │
└────────────────┴────────────────────────────────────────────┘

Agents ready:
  Cereal   → /agent "Are there any malicious packages?"
  Razor    → /agent "Review code security"
  Blade    → /agent "Are we SOC 2 compliant?"
  Phreak   → /agent "Any license conflicts?"

Storage: ~/.zero/repos/expressjs/express/ (45 MB)
```

---

*"Hack the planet!"*
