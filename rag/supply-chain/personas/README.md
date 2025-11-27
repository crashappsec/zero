<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
SPDX-License-Identifier: GPL-3.0
-->

# Supply Chain Scanner Personas

The supply chain scanner supports **persona-based analysis** that tailors Claude AI's output based on who is consuming the results.

## Available Personas

### 1. Security Engineer (`security-engineer`)
**Focus:** Technical vulnerability analysis with remediation priorities

Output includes:
- CVE triage with CVSS/EPSS scores
- CISA KEV prioritization
- Attack surface analysis
- Risk-based remediation ordering
- Compensating controls when patches aren't available

RAG content: `security/`
- CISA KEV prioritization
- CVE remediation workflows
- Vulnerability scoring models
- Remediation techniques
- Security metrics (MTTR, SLA compliance)

### 2. Software Engineer (`software-engineer`)
**Focus:** Developer-friendly guidance with copy-paste commands

Output includes:
- Ready-to-run CLI commands (npm, pip, cargo, etc.)
- Version upgrade tables with breaking change notes
- Dependency conflict resolution
- Bundle size impact analysis
- Migration hints and code examples

RAG content: `engineer/`
- Upgrade path patterns
- Package manager commands
- Dependency resolution
- Build optimization
- Maintenance interpretation

### 3. Engineering Leader (`engineering-leader`)
**Focus:** Executive summary with metrics and strategic recommendations

Output includes:
- Portfolio health dashboard
- Risk heat maps
- Team performance comparisons
- Resource planning implications
- Compliance status
- Strategic recommendations

RAG content: `leader/`
- Risk communication
- Prioritization frameworks
- Resource planning
- Portfolio dashboards
- Compliance mapping

### 4. Auditor (`auditor`)
**Focus:** Compliance assessment with framework mappings

Output includes:
- Control effectiveness assessment
- SOC 2, PCI DSS, NIST, ISO mappings
- Evidence requirements
- Finding documentation (criteria, condition, cause, effect)
- Audit workpaper structure

RAG content: `auditor/`
- Audit standards
- Control testing procedures
- Evidence collection
- Finding templates
- Compliance frameworks

## Usage

### Interactive Selection
When using `--claude` without specifying a persona, you'll be prompted:

```bash
./supply-chain-scanner.sh --claude /path/to/repo

# Prompts:
# Select analysis persona:
# 1) Security Engineer
# 2) Software Engineer
# 3) Engineering Leader
# 4) Auditor
```

### Command Line
Specify persona directly:

```bash
./supply-chain-scanner.sh --claude --persona security-engineer /path/to/repo
./supply-chain-scanner.sh --claude --persona software-engineer /path/to/repo
./supply-chain-scanner.sh --claude --persona engineering-leader /path/to/repo
./supply-chain-scanner.sh --claude --persona auditor /path/to/repo
```

### List Available Personas
```bash
./supply-chain-scanner.sh --list-personas
```

## Architecture

```
personas/
├── security/          # Security Engineer RAG
│   ├── cisa-kev-prioritization.md
│   ├── cve-remediation-workflows.md
│   ├── vulnerability-scoring.md
│   ├── remediation-techniques.md
│   └── security-metrics.md
├── engineer/          # Software Engineer RAG
│   ├── upgrade-path-patterns.md
│   ├── package-manager-commands.md
│   ├── dependency-resolution.md
│   ├── build-optimization.md
│   └── maintenance-interpretation.md
├── leader/            # Engineering Leader RAG
│   ├── risk-communication.md
│   ├── prioritization-frameworks.md
│   ├── resource-planning.md
│   ├── portfolio-dashboards.md
│   └── compliance-mapping.md
└── auditor/           # Auditor RAG
    ├── audit-standards.md
    ├── control-testing.md
    ├── evidence-collection.md
    ├── finding-templates.md
    └── compliance-frameworks.md
```

### Related Files

- **Prompt Templates:** `/prompts/supply-chain/personas/*.md`
- **Skill Overlays:** `/skills/supply-chain/overlays/*.md`
- **Persona Loader:** `/utils/supply-chain/lib/persona-loader.sh`

## How It Works

When a persona is selected:

1. **Skill Overlay** is loaded - modifies Claude's behavioral approach
2. **Persona RAG** is loaded - provides specialized knowledge
3. **Prompt Template** is loaded - defines output structure
4. All content is combined with scan data and sent to Claude API

The persona loader library (`lib/persona-loader.sh`) provides:
- `load_persona_prompt()` - Load prompt template
- `load_persona_rag()` - Load RAG content
- `load_persona_overlay()` - Load skill overlay
- `load_persona_context()` - Load all persona content combined

## Adding New Personas

1. Add persona to `VALID_PERSONAS` in `supply-chain-scanner.sh`
2. Create prompt template in `/prompts/supply-chain/personas/`
3. Create RAG directory and files in `/rag/supply-chain/personas/`
4. Create skill overlay in `/skills/supply-chain/overlays/`
5. Update `get_persona_rag_subdir()` in `lib/persona-loader.sh`
6. Update interactive selection menu in scanner

## Extending to Other Scanners

The persona system is designed to be reusable. To add personas to other scanners:

1. Copy/adapt `lib/persona-loader.sh`
2. Create scanner-specific RAG content
3. Create scanner-specific prompt templates
4. Integrate into scanner's Claude analysis function
