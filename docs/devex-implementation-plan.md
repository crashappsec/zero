# DevEx Scanner Implementation Plan

## Overview

The DevEx (Developer Experience) scanner analyzes repositories for onboarding friction, technology complexity, and workflow efficiency. It depends on the **tech-id scanner** as the single source of truth for technology detection.

This scanner aligns with modern developer productivity frameworks:
- **DORA metrics** (via devops scanner) - deployment frequency, lead time, MTTR, change failure rate
- **SPACE framework** - satisfaction, performance, activity, communication, efficiency

## Architecture

```text
tech-id scanner                    devops scanner
    │                                   │
    ├── Detects: languages,             ├── DORA metrics
    │   frameworks, databases,          │   - Deployment frequency
    │   cloud services, CI/CD           │   - Lead time
    │                                   │   - Change failure rate
    └── Outputs: tech-id.json           │   - MTTR
                    │                   │
                    ▼                   └── Outputs: devops.json
              devex scanner                        │
                    │                              │
                    ├── Loads: tech-id.json ◄──────┘ (optional)
                    │
                    ├── Analyzes: onboarding, technology sprawl, workflow
                    │
                    └── Outputs: devex.json
```

## Relationship to Productivity Frameworks

### DORA Metrics (Already Implemented in devops scanner)

The **devops scanner** already provides DORA metrics:

| Metric | Description | Location |
|--------|-------------|----------|
| Deployment Frequency | How often code is deployed to production | `devops.json` |
| Lead Time for Changes | Time from commit to production | `devops.json` |
| Mean Time to Recovery (MTTR) | Time to restore service after incident | `devops.json` |
| Change Failure Rate | % of deployments causing failures | `devops.json` |

DevEx can optionally consume DORA data to provide context (e.g., "high complexity but elite DORA performance").

### SPACE Framework

The [SPACE framework](https://queue.acm.org/detail.cfm?id=3454124) from Microsoft Research captures five dimensions of developer productivity:

| Dimension | What DevEx Measures | What Requires Surveys |
|-----------|--------------------|-----------------------|
| **S**atisfaction | - | Developer surveys (not automated) |
| **P**erformance | DORA metrics (via devops) | Quality outcomes |
| **A**ctivity | Build counts, release frequency | Sprint metrics |
| **C**ommunication | PR templates, code review setup | Team feedback |
| **E**fficiency | Onboarding time, tool sprawl, feedback loops | Flow state data |

DevEx focuses on **measurable infrastructure** that enables productivity, while SPACE's satisfaction dimension requires surveys.

## Features

### 1. Onboarding (friction analysis)

Measures how easy it is for new contributors to get started:

| Check | Description |
|-------|-------------|
| `check_readme_quality` | Analyzes README for install, usage, prerequisites, quick start sections |
| `check_contributing` | Checks for CONTRIBUTING.md |
| `check_env_setup` | Looks for .env.example, docker-compose |

> **Note**: `check_prerequisites` was removed - prerequisites are derived from tech-id scanner data.

**Output:**

- Onboarding score (0-100, higher = easier)
- Setup complexity level (low/medium/high)
- Setup barriers with suggestions

### 2. Sprawl Analysis (cognitive load)

Measures two distinct aspects of cognitive load:

#### 2a. Tool Sprawl (configuration burden)

Number of **development tools** that must be configured and used:

| Category | Examples |
|----------|----------|
| Linters | ESLint, Pylint, golangci-lint |
| Formatters | Prettier, Black, gofmt |
| Bundlers | Webpack, Vite, esbuild |
| Test frameworks | Jest, Pytest, Go test |
| CI/CD | GitHub Actions, CircleCI |
| Build tools | Make, Gradle, npm scripts |

**Impact**: Configuration overhead, context switching, maintenance burden

#### 2b. Technology Sprawl (learning curve)

Number of **technologies** a developer must understand:

| Category | Examples |
|----------|----------|
| Languages | Python, TypeScript, Go, Rust |
| Frameworks | React, FastAPI, Gin, Rails |
| Databases | PostgreSQL, Redis, MongoDB |
| Cloud services | AWS S3, GCP Pub/Sub, Azure Functions |
| Infrastructure | Docker, Kubernetes, Terraform |

**Impact**: Knowledge breadth required, onboarding time, hiring difficulty

#### Configuration

| Check | Description |
|-------|-------------|
| `check_tool_sprawl` | Counts dev tools from tech-id data |
| `check_technology_sprawl` | Counts all technologies from tech-id data |
| `check_config_complexity` | Analyzes config file line counts, nesting depth |
| `max_recommended_tools` | Threshold for tool sprawl warnings (default: 10) |
| `max_recommended_technologies` | Threshold for tech sprawl warnings (default: 15) |

**Output:**

- Tool sprawl index + level (low/moderate/high/excessive)
- Technology sprawl index + level
- Breakdown by category
- Combined complexity score

### 3. Workflow (efficiency analysis)

Measures development workflow quality (SPACE "Efficiency" dimension):

| Check | Description |
|-------|-------------|
| `check_pr_templates` | Looks for PR and issue templates |
| `check_local_dev` | Checks for docker-compose, devcontainer, Makefile |
| `check_feedback_loop` | Detects hot reload, watch mode tools |

**Output:**

- Workflow score (0-100)
- PR process score
- Local dev score
- Feedback loop score

## Implementation Status

### Completed

| Task | Description |
|------|-------------|
| ✅ Scanner skeleton | `pkg/scanners/devex/` with types, config, main scanner |
| ✅ Tech-id dependency | Declared in `Dependencies()`, loads tech-id.json |
| ✅ Onboarding feature | README analysis, env detection |
| ✅ Tooling feature | Uses tech-id data, config complexity analysis |
| ✅ Workflow feature | PR templates, local dev setup, feedback tools |
| ✅ Config integration | Added to `zero.config.json` |
| ✅ Profile integration | Added to `full` profile, new `devex-only` profile |

### TODO

#### DevEx Scanner Changes

| Task | Description |
|------|-------------|
| ⬜ Remove check_prerequisites | Use tech-id data instead of local detection |
| ⬜ Split sprawl metrics | Separate tool sprawl (dev tools) from technology sprawl (all tech) |
| ⬜ Add category breakdown | Show sprawl by category (languages, frameworks, etc.) |
| ⬜ DORA integration | Optionally load devops.json for context |
| ⬜ Learning curve score | Estimate based on technology count and complexity |

#### Tech-ID / RAG Pipeline Changes (Required for DevEx)

| Task | Description |
|------|-------------|
| ⬜ Audit all RAG patterns.md | Ensure all patterns have complete sections (see checklist below) |
| ⬜ Extend RAG converter for ALL sections | Parse and convert all RAG sections to Semgrep rules |
| ⬜ Add category metadata to RAG patterns | Ensure all patterns.md have `Category: developer-tools/linting` etc. |
| ⬜ Map RAG categories to tool/technology | `developer-tools/*` → Tool, `languages/*` → Technology |
| ⬜ Add missing RAG patterns | Build tools (Make, Gradle, Maven, CMake), formatters (Black, gofmt) |
| ⬜ Remove hardcoded configPatterns | Once RAG covers all config file detection |

#### RAG Pattern Sections Checklist

Each `patterns.md` should have these sections, and the RAG → Semgrep converter should handle ALL of them:

| Section | Currently Converted? | Semgrep Rule Type |
|---------|---------------------|-------------------|
| `Package Detection` | ❌ No | Could use SBOM data instead |
| `Configuration Files` | ❌ No | `paths:` patterns for file detection |
| `Import Detection` | ✅ Yes | `pattern:` for code imports |
| `Environment Variables` | ❌ No | `pattern-regex:` for env var usage |
| `Secrets Detection` | ✅ Yes | `pattern-regex:` for secret patterns |
| `Detection Confidence` | ⚠️ Partial | `metadata.confidence` in rules |

#### RAG Converter Enhancement Tasks

| Task | Description |
|------|-------------|
| ⬜ Parse "Configuration Files" section | Generate rules with `paths:` to match config filenames |
| ⬜ Parse "Package Detection" section | Optional: cross-reference with SBOM data |
| ⬜ Parse "Environment Variables" section | Generate rules to detect env var usage in code |
| ⬜ Preserve category hierarchy | `developer-tools/linting` → metadata for tool/tech classification |
| ⬜ Add file extension support | Use `File extensions:` hints for language targeting |
| ⬜ Handle "Detection Notes" | Add to rule `message` or `metadata.notes` |

#### Tech-ID Scanner Refactor Tasks

| Task | Description |
|------|-------------|
| ⬜ Remove hardcoded `configPatterns` map | Replace with RAG-generated Semgrep rules |
| ⬜ Remove hardcoded `extensionMap` | Use RAG patterns with file extension hints |
| ⬜ Remove hardcoded `sbomPatterns` | Use RAG "Package Detection" cross-referenced with SBOM |
| ⬜ Remove hardcoded `frameworkPatterns` | Use RAG "Import Detection" patterns |
| ⬜ Consolidate all detection to Semgrep | Single detection path: RAG → Semgrep → Results |
| ⬜ Use RAG category for output classification | `developer-tools/linting` → `{"category": "linter", "type": "tool"}` |
| ⬜ Preserve RAG confidence scores | Pass through to findings metadata |
| ⬜ Add RAG source URL to findings | Include `homepage` from RAG in output |

### Tech-ID Architecture & Gap Analysis

**Current Pipeline (has hardcoded patterns):**

```text
RAG Patterns (markdown)                    Hardcoded Patterns (Go)
rag/technology-identification/             pkg/scanners/tech-id/technology.go
├── languages/                             ├── configPatterns (config files) ❌
├── frameworks/                            ├── extensionMap (file extensions) ❌
├── databases/                             └── sbomPatterns (SBOM packages) ❌
├── developer-tools/
│   ├── bundlers/ ✅
│   ├── linting/ ✅
│   └── testing/ ✅
└── ai-ml/
        │
        ▼
    RAG → Semgrep Converter
    (rag_converter.go)
        │
        ├── Import detection rules ✅
        ├── Secret detection rules ✅
        └── Config file detection ❌
        │
        ▼
    Semgrep Rules (YAML)
        │
        ▼
    Tech-ID Scanner
    (runs Semgrep + hardcoded detection)
```

**Target Pipeline (all RAG, no hardcoded):**

```text
RAG Patterns (markdown)
rag/technology-identification/
├── languages/           → Category: language (Technology)
├── web-frameworks/      → Category: framework (Technology)
├── databases/           → Category: database (Technology)
├── developer-tools/
│   ├── bundlers/        → Category: bundler (Tool)
│   ├── linting/         → Category: linter (Tool)
│   ├── testing/         → Category: testing (Tool)
│   ├── cicd/            → Category: ci-cd (Tool)
│   └── infrastructure/  → Category: iac (Tool)
└── ai-ml/               → Category: ai-ml (Technology)
        │
        ▼
    RAG → Semgrep Converter (ENHANCED)
    (rag_converter.go)
        │
        ├── Import detection rules      (from "Import Detection")
        ├── Secret detection rules      (from "Secrets Detection")
        ├── Config file detection rules (from "Configuration Files")
        ├── Env var detection rules     (from "Environment Variables")
        └── Metadata: category, confidence, homepage, type (tool/tech)
        │
        ▼
    Semgrep Rules (YAML)
    - tech-discovery.yaml
    - secrets.yaml
    - ai-ml.yaml
    - config-files.yaml (NEW)
        │
        ▼
    Tech-ID Scanner (SIMPLIFIED)
    - Runs Semgrep only
    - No hardcoded patterns
    - Merges results with category metadata
        │
        ▼
    tech-id.json output
    - findings with category, type (tool/technology), confidence
    - source URLs from RAG homepage
```

**The Problem:**

RAG already has patterns for developer tools:
- `rag/technology-identification/developer-tools/linting/eslint/patterns.md` ✅
- `rag/technology-identification/developer-tools/bundlers/webpack/patterns.md` ✅
- `rag/technology-identification/developer-tools/testing/jest/patterns.md` ✅

But the RAG → Semgrep converter only generates rules for:
- **Import detection** (e.g., `import eslint`)
- **Secret detection** (e.g., API keys)

It does NOT generate rules for:
- **Config file detection** (e.g., `.eslintrc.json`)

Config file detection is still **hardcoded** in `technology.go` and doesn't use RAG categories.

**Solution:**

Option A: Extend RAG converter to generate config file detection rules
Option B: Have tech-id scanner read RAG config file patterns directly
Option C: Add config file patterns to hardcoded Go maps with proper categories

**Recommended: Option A** - Extend RAG converter to generate Semgrep rules for config file detection, keeping all intelligence in RAG.

**Categories tech-id output SHOULD have:**

| Category | Type | RAG Location |
|----------|------|--------------|
| `language` | Technology | `rag/technology-identification/languages/` |
| `framework` | Technology | `rag/technology-identification/web-frameworks/` |
| `database` | Technology | `rag/technology-identification/databases/` |
| `cloud` | Technology | Various |
| `container` | Tool | `rag/technology-identification/developer-tools/containers/` |
| `iac` | Tool | `rag/technology-identification/developer-tools/infrastructure/` |
| `ci-cd` | Tool | `rag/technology-identification/developer-tools/cicd/` |
| `linter` | Tool | `rag/technology-identification/developer-tools/linting/` |
| `formatter` | Tool | (needs patterns or merge with linting) |
| `bundler` | Tool | `rag/technology-identification/developer-tools/bundlers/` |
| `testing` | Tool | `rag/technology-identification/developer-tools/testing/` |
| `build` | Tool | (needs patterns for Make, Gradle, Maven) |

**Action Items:**

1. Extend `rag_converter.go` to parse "Configuration Files" section from RAG patterns
2. Generate Semgrep rules that detect config files (using `filename:` or path patterns)
3. Ensure RAG patterns include proper `Category` metadata that maps to tool/technology type
4. Remove hardcoded `configPatterns` from `technology.go` once RAG covers everything

## Configuration

### Scanner Config (zero.config.json)

```json
"devex": {
  "name": "Developer Experience",
  "description": "Developer experience analysis: onboarding friction, sprawl analysis, workflow efficiency",
  "estimated_time": "15-30s",
  "output_file": "devex.json",
  "dependencies": ["tech-id"],
  "features": {
    "onboarding": {
      "enabled": true,
      "check_readme_quality": true,
      "check_contributing": true,
      "check_env_setup": true
    },
    "sprawl": {
      "enabled": true,
      "check_tool_sprawl": true,
      "check_technology_sprawl": true,
      "check_config_complexity": true,
      "max_recommended_tools": 10,
      "max_recommended_technologies": 15,
      "tool_categories": ["linter", "formatter", "bundler", "test", "ci-cd", "build"],
      "technology_categories": ["language", "framework", "database", "cloud", "container", "infrastructure"]
    },
    "workflow": {
      "enabled": true,
      "check_pr_templates": true,
      "check_local_dev": true,
      "check_feedback_loop": true
    }
  }
}
```

### Profiles

| Profile | Includes DevEx | Notes |
|---------|---------------|-------|
| `full` | ✅ | Complete analysis |
| `devex-only` | ✅ | DevEx + tech-id only |
| `standard` | ❌ | Security/quality focus |
| `quick` | ❌ | Fast feedback |

## Output Schema

### devex.json

```json
{
  "scanner": "devex",
  "version": "1.0.0",
  "summary": {
    "onboarding": {
      "score": 75,
      "setup_complexity": "medium",
      "config_file_count": 12,
      "dependency_count": 45,
      "build_step_count": 3,
      "has_contributing": true,
      "has_env_example": true,
      "readme_quality_score": 80
    },
    "sprawl": {
      "combined_score": 62,
      "tool_sprawl": {
        "index": 8,
        "level": "moderate",
        "by_category": {
          "linter": 2,
          "formatter": 1,
          "bundler": 1,
          "test": 2,
          "ci-cd": 1,
          "build": 1
        }
      },
      "technology_sprawl": {
        "index": 12,
        "level": "moderate",
        "learning_curve": "moderate",
        "by_category": {
          "language": 2,
          "framework": 3,
          "database": 2,
          "cloud": 2,
          "container": 2,
          "infrastructure": 1
        }
      },
      "config_complexity": "medium"
    },
    "workflow": {
      "score": 70,
      "efficiency_level": "medium",
      "has_pr_templates": true,
      "has_devcontainer": false,
      "has_hot_reload": true
    }
  },
  "findings": {
    "onboarding": {
      "config_files": [],
      "setup_barriers": []
    },
    "sprawl": {
      "tools": [
        {"name": "ESLint", "category": "linter"},
        {"name": "Prettier", "category": "formatter"},
        {"name": "Jest", "category": "test"}
      ],
      "technologies": [
        {"name": "TypeScript", "category": "language"},
        {"name": "React", "category": "framework"},
        {"name": "PostgreSQL", "category": "database"}
      ],
      "sprawl_issues": [],
      "config_analysis": []
    },
    "workflow": {
      "pr_templates": [],
      "dev_setup": {},
      "workflow_issues": []
    }
  }
}
```

## SPACE Framework Alignment

How DevEx maps to SPACE dimensions:

| SPACE Dimension | DevEx Coverage | Notes |
|-----------------|---------------|-------|
| **Satisfaction** | ❌ Not covered | Requires developer surveys |
| **Performance** | ⚠️ Via devops | DORA metrics in devops.json |
| **Activity** | ⚠️ Via devops | Git activity in devops.json |
| **Communication** | ✅ Partial | PR templates, code review setup |
| **Efficiency** | ✅ Strong | Onboarding, feedback loops, workflow |

DevEx is strongest at measuring the **infrastructure that enables efficiency** - the tooling, documentation, and workflows that reduce friction.

## Future Enhancements

1. **DORA context integration**: Load devops.json to provide context like "high sprawl but elite deployment frequency"

2. **Learning curve estimation**: Score based on:
   - Number of languages (more = steeper)
   - Framework complexity (React vs vanilla JS)
   - Infrastructure requirements (K8s vs simple Docker)

3. **IDE config detection**: VS Code settings, devcontainer configs, recommended extensions

4. **Onboarding time estimation**: Estimate "time to first build" based on:
   - Dependency count
   - Build step count
   - Technology sprawl

5. **SPACE survey integration**: Optional webhook/API to collect satisfaction data

6. **Comparison mode**: Compare devex scores across repos or over time

## References

- [SPACE Framework (ACM Queue)](https://queue.acm.org/detail.cfm?id=3454124)
- [SPACE Metrics Explained (LinearB)](https://linearb.io/blog/space-framework)
- [DORA Metrics](https://dora.dev/research/)
- [Microsoft Developer Experience](https://developer.microsoft.com/en-us/developer-experience)

## Files

| File | Purpose |
|------|---------|
| `pkg/scanners/devex/devex.go` | Main scanner implementation |
| `pkg/scanners/devex/types.go` | Type definitions |
| `pkg/scanners/devex/config.go` | Feature configuration |
| `config/zero.config.json` | Scanner and profile config |
| `pkg/scanners/all.go` | Scanner registration |
