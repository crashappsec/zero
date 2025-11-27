# System Architecture Overview

## Introduction

This repository provides a comprehensive toolkit for software analysis, including supply chain security, code security, technology identification, and more. The system uses Claude AI enhanced with structured knowledge bases and role-based personas.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Analysis Scanners                               │
│  ┌────────────────┐ ┌────────────────┐ ┌────────────────┐ ┌──────────────┐ │
│  │  Supply Chain  │ │  Code Security │ │   Technology   │ │    Legal     │ │
│  │    Scanner     │ │    Scanner     │ │ Identification │ │   Analyser   │ │
│  └────────────────┘ └────────────────┘ └────────────────┘ └──────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           Claude AI Analysis                                 │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                         Persona Layer                                │   │
│  │    (Output styling, templates, prioritization per role)              │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                    │                                        │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                       Knowledge Base                                 │   │
│  │    (Patterns, frameworks, definitions - single source of truth)     │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                    │                                        │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                      RAG Content Library                             │   │
│  │    (Technology patterns, legal terms, compliance references)        │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                            Analysis Output                                   │
│         (Tailored to persona: technical, executive, compliance, etc.)       │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. Analysis Scanners (`utils/`)

Shell-based tools that gather data from repositories:

| Scanner | Location | Purpose |
|---------|----------|---------|
| Supply Chain | `utils/supply-chain/` | Dependency vulnerabilities, license compliance, package health |
| Code Security | `utils/code-security/` | Static analysis, secret detection, security patterns |
| Technology ID | `utils/technology-identification/` | Detect frameworks, APIs, services in use |
| Legal Review | `utils/legal-review/` | License analysis, compliance checking |
| Code Ownership | `utils/code-ownership/` | CODEOWNERS analysis, contribution patterns |
| Certificate | `utils/certificate-analyser/` | TLS/SSL certificate validation |
| DORA Metrics | `utils/dora-metrics/` | DevOps performance metrics |

### 2. Knowledge Base (`specialist-agents/knowledge/`)

**Single source of truth** for all factual content used in analysis:

```
specialist-agents/knowledge/
├── security/           # Vulnerability patterns, CWE, OWASP, MITRE ATT&CK
├── supply-chain/       # Ecosystem patterns, health signals, licenses
├── compliance/         # Audit standards, frameworks, control testing
├── dependencies/       # Package management, upgrade patterns
├── devops/            # CI/CD security, infrastructure patterns
├── engineering/       # Code quality, performance patterns
└── shared/            # Severity levels, confidence scores, formatting
```

See [Knowledge Base Architecture](knowledge-base.md) for details.

### 3. Personas (`rag/supply-chain/personas/`)

Define how to present analysis results for different audiences:

- **Security Engineer**: Technical depth, CVE details, remediation commands
- **Software Engineer**: Practical commands, migration guides, checklists
- **Engineering Leader**: Metrics dashboards, strategic recommendations
- **Auditor**: Compliance mapping, control assessment, evidence requirements

See [Personas Architecture](personas.md) for details.

### 4. RAG Content Library (`rag/`)

Retrieval-Augmented Generation content for domain-specific context:

```
rag/
├── technology-identification/  # 100+ technology detection patterns
├── legal-review/              # License terms and obligations
├── certificate-analysis/      # CA/Browser Forum requirements
├── code-security/            # Vulnerability patterns by language
└── supply-chain/             # Ecosystem-specific guidance
```

### 5. Skills (`skills/`)

Reusable prompt components for Claude Code integration.

### 6. Prompts (`prompts/`)

Structured prompts for different analysis scenarios.

## Data Flow

1. **Scanner Execution**: User runs scanner on a repository
2. **Data Collection**: Scanner gathers dependency, code, or configuration data
3. **Claude Analysis**: Data sent to Claude with:
   - Relevant knowledge base content
   - Selected persona definition
   - RAG content for domain context
4. **Output Generation**: Claude produces analysis tailored to persona
5. **Report Delivery**: Results formatted per persona preferences

## Key Design Principles

### Single Source of Truth
All factual content (patterns, frameworks, definitions) lives in the knowledge base. Personas and other components reference it - never duplicate it.

### Separation of Concerns
- **Scanners**: Data collection only
- **Knowledge Base**: Facts and patterns
- **Personas**: Presentation and prioritization
- **RAG Content**: Domain-specific context

### Extensibility
- Add new scanners without changing knowledge
- Add new personas without duplicating content
- Add new knowledge without modifying existing personas

## Directory Structure

```
gibson-powers/
├── docs/                    # All documentation (you are here)
│   ├── architecture/       # Architecture documentation
│   ├── guides/            # How-to guides
│   ├── plans/             # Implementation plans
│   └── references/        # Reference documentation
├── utils/                  # Analysis scanners
├── specialist-agents/      # Agent definitions and knowledge
│   └── knowledge/         # Knowledge base (single source of truth)
├── rag/                    # RAG content library
│   └── supply-chain/
│       └── personas/      # Persona definitions
├── skills/                 # Claude Code skills
├── prompts/               # Prompt templates
└── config/                # Configuration files
```

## Related Documentation

- [Knowledge Base Architecture](knowledge-base.md)
- [Personas Architecture](personas.md)
- [Supply Chain Scanner](../../utils/supply-chain/README.md)
- [Code Security Scanner](../../utils/code-security/README.md)
