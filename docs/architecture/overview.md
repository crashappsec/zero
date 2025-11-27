# System Architecture Overview

## Introduction

This repository provides a comprehensive toolkit for software analysis and engineering assistance. The system uses AI agents enhanced with structured knowledge bases, deployable to Claude instances and the Crash Override platform.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                               AGENTS                                         │
│                          (Self-contained, portable)                          │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                        Security Agents                               │   │
│  │  ┌────────────────┐  ┌────────────────┐                             │   │
│  │  │  Supply Chain  │  │  Code Security │                             │   │
│  │  └────────────────┘  └────────────────┘                             │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                       Engineering Agents                             │   │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐  │   │
│  │  │ Frontend │ │ Backend  │ │Architect │ │  Build   │ │  DevOps  │  │   │
│  │  │ Engineer │ │ Engineer │ │          │ │ Engineer │ │ Engineer │  │   │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘  │   │
│  │                        ┌──────────────┐                             │   │
│  │                        │ Engineering  │                             │   │
│  │                        │   Leader     │                             │   │
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
│  │   (Claude Code)     │  │    Platform         │  │   (utils/)       │    │
│  └─────────────────────┘  └─────────────────────┘  └──────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. Agents (`agents/`)

Self-contained, portable AI agents. Each agent includes everything needed to run:

```
agents/
├── supply-chain/           # Security: dependency vulnerabilities, licenses
├── code-security/          # Security: static analysis, secrets, patterns
├── frontend-engineer/      # Engineering: React, TypeScript, web apps
├── backend-engineer/       # Engineering: APIs, databases, data engineering
├── architect/              # Engineering: system design, patterns, auth
├── build-engineer/         # Engineering: CI/CD optimization, build speed
├── devops-engineer/        # Engineering: deployments, infrastructure
├── engineering-leader/     # Engineering: costs, metrics, team effectiveness
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
phantom/
├── agents/                 # Self-contained AI agents
│   ├── supply-chain/
│   ├── code-security/
│   ├── frontend-engineer/
│   ├── backend-engineer/
│   ├── architect/
│   ├── build-engineer/
│   ├── devops-engineer/
│   ├── engineering-leader/
│   └── shared/
├── docs/                   # Documentation
│   └── architecture/
├── utils/                  # CLI analysis tools
├── rag/                    # RAG content library
├── skills/                 # Claude Code skills
├── prompts/               # Prompt templates
└── config/                # Configuration
```

## Related Documentation

- [Agents README](../../agents/README.md) - Agent catalog and usage
- [Knowledge Base Architecture](knowledge-base.md) - Knowledge organization
