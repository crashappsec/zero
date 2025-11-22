<!--
Copyright (c) 2024 Gibson Powers Contributors

SPDX-License-Identifier: GPL-3.0
-->

# COCOMO Estimator Utilities

**Status**: 📋 Planned - Not yet implemented

Automated software cost estimation tools using COCOMO II models.

## Overview

These utilities will provide command-line tools for estimating software project effort, schedule, and cost using the COCOMO II (Constructive Cost Model) methodology with automated code analysis.

## Planned Architecture

```
utils/cocomo/
├── cocomo-estimator.sh              # Base COCOMO II estimator
├── cocomo-estimator-claude.sh       # AI-enhanced estimator
├── compare-estimators.sh            # Compare base vs AI
├── calibration/
│   ├── calibrate-model.sh           # Local calibration tool
│   └── historical-tracker.sh        # Track estimates vs actuals
├── config.example.json              # Configuration template
├── README.md                        # This file
└── CHANGELOG.md                     # Version history (future)
```

## Planned Features

### Base Estimator (`cocomo-estimator.sh`)

**Capabilities**:
- Automated SLOC (Source Lines of Code) counting
- Multi-language support with appropriate conversion factors
- COCOMO II effort and schedule calculation
- Cost driver assessment
- Uncertainty range estimation
- Multiple output formats (table, JSON, markdown)

**Usage** (planned):
```bash
# Basic estimation
./cocomo-estimator.sh owner/repo

# With custom cost drivers
./cocomo-estimator.sh owner/repo \
  --prec 3 \
  --flex 2 \
  --resl 4 \
  --team 4 \
  --pmat 3

# With cost per month
./cocomo-estimator.sh owner/repo --cost-per-month 15000

# JSON output
./cocomo-estimator.sh owner/repo --format json

# Multiple repositories
./cocomo-estimator.sh --org myorg

# Specific files only
./cocomo-estimator.sh owner/repo --include "src/**/*.js"
```

### AI-Enhanced Estimator (`cocomo-estimator-claude.sh`)

**Additional Capabilities**:
- Smart cost driver rating suggestions
- Complexity analysis from code patterns
- Similar project comparison
- Risk factor identification
- Alternative scenario modeling
- Detailed breakdown recommendations
- Calibration suggestions

**Usage** (planned):
```bash
# AI-enhanced estimation
export ANTHROPIC_API_KEY="your-key"
./cocomo-estimator-claude.sh owner/repo

# With scenario analysis
./cocomo-estimator-claude.sh owner/repo --scenarios

# With similar project comparison
./cocomo-estimator-claude.sh owner/repo --compare-similar
```

### Calibration Tools

**Local Calibration**:
```bash
# Initialize calibration database
./calibration/calibrate-model.sh --init

# Add actual project data
./calibration/calibrate-model.sh --add \
  --repo owner/repo \
  --actual-effort 150 \
  --actual-duration 15

# Generate calibration factors
./calibration/calibrate-model.sh --calibrate

# Track accuracy over time
./calibration/historical-tracker.sh
```

## Planned Analysis Process

### 1. Code Size Calculation
```bash
# Count lines of code by language
- Use cloc or similar tool
- Apply language-specific filters
- Exclude comments, blanks (configurable)
- Calculate logical SLOC
```

### 2. Scale Factor Assessment
```bash
# 5 scale factors (0-5 rating each):
- Precedentedness (PREC)
- Development Flexibility (FLEX)
- Architecture/Risk Resolution (RESL)
- Team Cohesion (TEAM)
- Process Maturity (PMAT)

# Calculate B exponent:
B = 0.91 + 0.01 × Σ SF
```

### 3. Cost Driver Evaluation
```bash
# 17 cost drivers across 4 categories:
Product: RELY, DATA, CPLX, RUSE, DOCU
Platform: TIME, STOR, PVOL
Personnel: ACAP, PCAP, PCON, APEX, PLEX, LTEX
Project: TOOL, SITE, SCED

# Calculate effort multiplier:
EM = ∏ EM_i
```

### 4. Effort Calculation
```bash
# COCOMO II formula:
Effort = 2.94 × Size^B × EM

# Where:
- Size in KSLOC (thousands of SLOC)
- B from scale factors
- EM from cost drivers
```

### 5. Schedule Calculation
```bash
# Development time:
TDEV = 3.67 × Effort^(0.28 + 0.2×(B-0.91))

# Team size:
Team = Effort / TDEV
```

## Planned Output Formats

### Table Format (Default)
```
╭─────────────────────┬───────────────╮
│ Metric              │ Value         │
├─────────────────────┼───────────────┤
│ Total SLOC          │ 45,250        │
│ Effort (PM)         │ 142           │
│ Duration (months)   │ 14.2          │
│ Team Size           │ 10            │
│ Cost                │ $2,130,000    │
╰─────────────────────┴───────────────╯
```

### JSON Format
```json
{
  "repository": "owner/repo",
  "analysis_date": "2024-11-21",
  "size": {
    "total_sloc": 45250,
    "new": 42000,
    "reused": 2500,
    "modified": 750,
    "languages": {
      "javascript": 25000,
      "python": 15000,
      "sql": 5250
    }
  },
  "scale_factors": {
    "prec": 3,
    "flex": 2,
    "resl": 3,
    "team": 4,
    "pmat": 3,
    "exponent": 1.08
  },
  "effort_multiplier": 1.02,
  "estimates": {
    "effort_pm": 142,
    "duration_months": 14.2,
    "team_size": 10,
    "cost": 2130000
  },
  "uncertainty": {
    "optimistic": {
      "effort_pm": 114,
      "duration_months": 11.4
    },
    "pessimistic": {
      "effort_pm": 178,
      "duration_months": 17.7
    }
  }
}
```

### Markdown Format
```markdown
# COCOMO Estimation Report

**Repository**: owner/repo
**Date**: 2024-11-21

## Code Size
- Total: 45,250 SLOC
- New: 42,000 (93%)
- Reused: 2,500 (5%)
- Modified: 750 (2%)

## Estimates
- **Effort**: 142 person-months
- **Duration**: 14.2 months
- **Team**: 10 people
- **Cost**: $2,130,000
```

## Planned Configuration

**Example `config.json`**:
```json
{
  "cocomo": {
    "version": "cocomo_ii",
    "default_scale_factors": {
      "prec": 3,
      "flex": 3,
      "resl": 3,
      "team": 3,
      "pmat": 3
    },
    "default_cost_drivers": {
      "rely": "nominal",
      "data": "nominal",
      "cplx": "nominal"
    },
    "cost_per_month": 15000,
    "language_factors": {
      "javascript": 47,
      "python": 38,
      "java": 53,
      "go": 40,
      "rust": 35
    },
    "exclude_patterns": [
      "**/node_modules/**",
      "**/vendor/**",
      "**/test/**",
      "**/*.min.js"
    ]
  }
}
```

## Planned Integration

### With Supply Chain Analyzer
- Factor in dependency count
- Consider technical debt
- Assess security impact on effort

### With DORA Metrics
- Use metrics for experience factors
- Deployment frequency for SCED
- Lead time for team assessment

### With Code Ownership
- Bus factor for personnel continuity
- Team size from ownership
- Experience from history

## Implementation Phases

### Phase 1: Basic Estimator
- [ ] SLOC counting (cloc integration)
- [ ] Basic COCOMO II calculation
- [ ] Manual factor input
- [ ] Table output format

### Phase 2: Automation
- [ ] Automated complexity analysis
- [ ] Multi-repo support
- [ ] JSON/markdown output
- [ ] Configuration system

### Phase 3: AI Enhancement
- [ ] Claude integration
- [ ] Smart factor suggestions
- [ ] Pattern recognition
- [ ] Risk assessment

### Phase 4: Advanced Features
- [ ] Calibration tools
- [ ] Historical tracking
- [ ] Comparison reports
- [ ] Dashboard integration

## Dependencies

**Required** (when implemented):
- `cloc` - Line counting
- `jq` - JSON processing
- `gh` - GitHub API access

**Optional**:
- `radon` - Python complexity
- `eslint` - JavaScript complexity
- `sonarqube-scanner` - Multi-language metrics

## Research Questions

1. How to automatically determine scale factors from code/repo?
2. Modern language productivity (Rust, Go, TypeScript)?
3. AI/Copilot impact on effort multipliers?
4. Microservices architecture complexity factor?
5. Cloud-native development differences?

## Why This Matters

**Business Value**:
- Budget planning and resource allocation
- Project timeline estimation
- Risk assessment and contingency
- Contract negotiation support

**Engineering Value**:
- Effort tracking and improvement
- Team sizing and planning
- Technology choice impact analysis
- Process improvement validation

## Limitations

Even when implemented, COCOMO estimation will have limitations:
- Accuracy depends on size estimation (most critical)
- Subjective factor ratings (AI helps but not perfect)
- Model assumptions may not fit all projects
- Works best for traditional development (adjustments needed for modern practices)

## References

- [COCOMO II Model](http://csse.usc.edu/tools/)
- [RAG Documentation](../../rag/cocomo/)
- [COCOMO Skill](../../skills/cocomo/)

## Current Status

**Status**: 📋 Planned - Not yet implemented

This directory and documentation serve as a specification for future implementation. No utilities currently exist.

**Timeline**: TBD based on:
- Community demand and feedback
- Resource availability
- Integration opportunities with existing tools

**Contributing**: Once implementation begins, contributions will be welcome. Until then, feedback on requirements and use cases is valuable.

## License

GPL-3.0 - See [LICENSE](../../LICENSE) for details.
