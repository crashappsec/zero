# System Architecture Overview

## Introduction

Zero is a comprehensive toolkit for software analysis and engineering assistance. Named after Zero Cool from the movie Hackers (1995), the system uses AI agents enhanced with structured knowledge bases, deployable to Claude instances and the Crash Override platform.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                 ZERO                                         │
│                        (Master Orchestrator)                                 │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                      Security Agents                                 │   │
│  │  ┌────────────────┐ ┌────────────────┐ ┌────────────────┐           │   │
│  │  │ Cereal (Supply │ │ Razor (Code    │ │ Blade (Audit)  │           │   │
│  │  │ Chain)         │ │ Security)      │ │                │           │   │
│  │  └────────────────┘ └────────────────┘ └────────────────┘           │   │
│  │  ┌────────────────┐                                                 │   │
│  │  │ Phreak (Legal) │                                                 │   │
│  │  └────────────────┘                                                 │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                       Engineering Agents                             │   │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐  │   │
│  │  │  Acid    │ │  Dade    │ │  Nikon   │ │  Joey    │ │  Plague  │  │   │
│  │  │(Frontend)│ │(Backend) │ │(Architect│ │ (Build)  │ │ (DevOps) │  │   │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘  │   │
│  │                        ┌──────────────┐                             │   │
│  │                        │   Gibson     │                             │   │
│  │                        │  (Metrics)   │                             │   │
│  │                        └──────────────┘                             │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                         Shared Knowledge                             │   │
│  │         (Severity levels, confidence scoring, formatting)            │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                          DEPLOYMENT TARGETS                                  │
│  ┌─────────────────────┐  ┌─────────────────────┐  ┌──────────────────┐    │
│  │   Claude Instance   │  │  Crash Override     │  │   CLI Tools      │    │
│  │   (Claude Code)     │  │    Platform         │  │   (zero.sh)      │    │
│  └─────────────────────┘  └─────────────────────┘  └──────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. Agents (`agents/`)

Self-contained, portable AI agents named after Hackers (1995) characters:

```
agents/
├── cereal/                 # Cereal Killer - Supply chain security, CVEs, malcontent
├── razor/                  # Razor - Static analysis, secrets, SAST
├── blade/                  # Blade - Compliance, SOC 2, ISO 27001
├── phreak/                 # Phantom Phreak - Legal, licenses, data privacy
├── acid/                   # Acid Burn - React, TypeScript, accessibility
├── dade/                   # Dade Murphy - APIs, databases, data engineering
├── nikon/                  # Lord Nikon - System design, architecture patterns
├── joey/                   # Joey - CI/CD optimization, build speed
├── plague/                 # The Plague - Infrastructure, Kubernetes
├── gibson/                 # The Gibson - DORA metrics, team effectiveness
└── shared/                 # Cross-agent knowledge
```

### Agent Structure

Each agent follows this structure:

```
agent-name/
├── agent.md               # Identity, capabilities, behavior
├── VERSION                # Semantic version (0.1.0)
├── CHANGELOG.md           # Version history
├── knowledge/
│   ├── patterns/          # Detection (what things ARE)
│   └── guidance/          # Interpretation (what things MEAN)
└── prompts/               # Role-specific output formats (optional)
```

### 2. CLI Tools (`utils/`)

Shell-based analysis tools that can work standalone or feed data to agents:

| Tool | Purpose |
|------|---------|
| `supply-chain/` | Dependency scanning, vulnerability detection |
| `code-security/` | Static analysis, secret detection |
| `technology-identification/` | Detect frameworks, APIs, services |
| `legal-review/` | License analysis |
| `code-ownership/` | CODEOWNERS analysis |
| `dora-metrics/` | DevOps performance metrics |

### 3. RAG Content (`rag/`)

Retrieval-Augmented Generation content for domain-specific context:

| Directory | Purpose |
|-----------|---------|
| `technology-identification/` | 100+ technology detection patterns |
| `legal-review/` | License terms and obligations |
| `certificate-analysis/` | CA/Browser Forum requirements |

## Key Design Principles

### 1. Self-Contained Agents

Each agent directory contains everything needed to run. Copy an agent to another system and it works:

```bash
# Deploy supply-chain agent to another system
cp -r agents/supply-chain/ /other/system/agents/
```

### 2. Patterns vs Guidance

Clear separation between:
- **Patterns**: Detection/identification (what things ARE)
- **Guidance**: Interpretation/action (what things MEAN)

### 3. Independent Versioning

Each agent has its own version, enabling:
- Independent release cycles
- Backward compatibility tracking
- Clear upgrade paths

### 4. Shared Knowledge

Common definitions in `agents/shared/` ensure consistency:
- Severity levels (Critical, High, Medium, Low, Info)
- Confidence scoring (Confirmed, High, Medium, Low, Speculative)
- Output formatting conventions

## Data Flow

### Agent-Based Analysis

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Input      │────▶│    Agent     │────▶│   Output     │
│ (repo, code) │     │ + Knowledge  │     │  (analysis)  │
└──────────────┘     └──────────────┘     └──────────────┘
```

### Tool + Agent Pipeline

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Input      │────▶│  CLI Tool    │────▶│    Agent     │────▶│   Output     │
│   (repo)     │     │  (scanning)  │     │  (analysis)  │     │  (report)    │
└──────────────┘     └──────────────┘     └──────────────┘     └──────────────┘
```

## Directory Structure

```
zero/
├── zero.sh                 # Main CLI entry point
├── agents/                 # Self-contained AI agents (Hackers-themed)
│   ├── cereal/             # Cereal Killer - supply chain
│   ├── razor/              # Razor - code security
│   ├── blade/              # Blade - compliance
│   ├── phreak/             # Phantom Phreak - legal
│   ├── acid/               # Acid Burn - frontend
│   ├── dade/               # Dade Murphy - backend
│   ├── nikon/              # Lord Nikon - architecture
│   ├── joey/               # Joey - build
│   ├── plague/             # The Plague - devops
│   ├── gibson/             # The Gibson - metrics
│   └── shared/
├── docs/                   # Documentation
│   └── architecture/
├── utils/
│   ├── zero/               # Zero orchestrator
│   │   ├── lib/            # Libraries (zero-lib.sh, agent-loader.sh)
│   │   ├── scripts/        # CLI scripts (hydrate, scan, report)
│   │   └── config/         # Configuration files
│   └── scanners/           # Individual scanners
├── rag/                    # RAG content library
├── prompts/                # Prompt templates
└── .claude/
    └── commands/           # Slash commands (/agent, /zero)
```

## Storage

All analysis data is stored in `~/.zero/`:

```
~/.zero/
├── config.json                 # Global settings
├── index.json                  # Project index
└── repos/
    └── expressjs/
        └── express/
            ├── project.json    # Project metadata
            ├── repo/           # Cloned repository
            └── analysis/       # Analysis results
```

## Related Documentation

- [Agents README](../../agents/README.md) - Agent catalog and usage
- [Knowledge Base Architecture](knowledge-base.md) - Knowledge organization
- [Zero Architecture](../plans/zero-architecture.md) - Orchestrator design
