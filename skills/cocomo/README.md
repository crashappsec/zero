<!--
Copyright (c) 2024 Crash Override Inc
101 Fulton St, 416, New York 10038
SPDX-License-Identifier: GPL-3.0
-->

# COCOMO Estimation Skill

**Status**: 📋 Planned - Future development

Software cost, effort, and schedule estimation using COCOMO II models with automated code analysis.

## Overview

**COCOMO** (Constructive Cost Model) provides algorithmic software cost estimation based on project size, complexity, and various cost drivers. This skill will enable automated estimation from code repositories.

## Planned Capabilities

### Analysis & Estimation
- **Automated Size Calculation**: Lines of code (LOC) counting by language
- **Complexity Assessment**: Code complexity metrics for CPLX factor
- **Effort Estimation**: Person-months required for development
- **Schedule Estimation**: Calendar time for project completion
- **Team Sizing**: Optimal team size recommendations
- **Cost Estimation**: Budget estimation with salary inputs

### Repository Analysis
- **Multi-Language Support**: Accurate counting for all major languages
- **Dependency Analysis**: Include third-party and reused code appropriately
- **Historical Calibration**: Learn from past projects to improve accuracy
- **Change Impact**: Estimate effort for modifications and enhancements
- **Maintenance Estimation**: Annual maintenance effort calculation

### COCOMO II Integration
- **Scale Factors**: Precedentedness, flexibility, risk resolution, team cohesion, process maturity
- **Cost Drivers**: 17 factors across product, platform, personnel, and project categories
- **Multiple Models**: Application composition, early design, post-architecture
- **Calibration Support**: Customize for organizational environment

## Planned Features

### Core Estimation
- Calculate SLOC (Source Lines of Code) from repositories
- Convert between SLOC and Function Points
- Apply COCOMO II effort equation with scale factors
- Calculate development schedule (TDEV)
- Estimate optimal team size
- Provide uncertainty ranges (optimistic, most likely, pessimistic)

### Intelligent Analysis
- Detect project characteristics automatically
- Suggest appropriate scale factor ratings
- Analyze codebase complexity for CPLX rating
- Consider reuse and modification factors
- Account for different development models (waterfall, agile, iterative)

### Repository Integration
- Clone and analyze GitHub repositories
- Scan organization repos in bulk
- Track estimates over time
- Compare estimates to actuals
- Generate estimation reports

### AI-Enhanced Features
- Smart cost driver rating suggestions
- Pattern recognition from similar projects
- Risk factor identification
- Alternative scenario modeling
- Effort breakdown by component/module

## Planned Utility Scripts

```
utils/cocomo/
├── cocomo-estimator.sh              # Base estimator
├── cocomo-estimator-claude.sh       # AI-enhanced estimator
├── compare-estimators.sh            # Compare approaches
└── calibration/
    └── calibrate-model.sh           # Local calibration tool
```

## Example Usage (Planned)

```bash
# Estimate single repository
./cocomo-estimator.sh owner/repo

# Estimate with custom factors
./cocomo-estimator.sh owner/repo \
  --prec 3 \
  --flex 2 \
  --team 4 \
  --cost-per-month 15000

# Organization-wide estimation
./cocomo-estimator.sh --org myorg

# AI-enhanced with smart suggestions
./cocomo-estimator-claude.sh owner/repo
```

## Planned Output

```
===========================================
COCOMO II Estimation Report
===========================================
Repository: owner/repo
Analysis Date: 2024-11-21

Code Size:
  Total SLOC: 45,250
  New Code: 42,000 (93%)
  Reused: 2,500 (5%)
  Modified: 750 (2%)

  Language Breakdown:
    JavaScript: 25,000 SLOC
    Python: 15,000 SLOC
    SQL: 5,250 SLOC

Scale Factors (B = 1.08):
  Precedentedness: 3 (Nominal)
  Flexibility: 2 (High)
  Risk Resolution: 3 (Nominal)
  Team Cohesion: 4 (High)
  Process Maturity: 3 (Nominal)

Effort Multiplier: 1.02
  Product: 1.05 (Above average complexity)
  Platform: 0.98 (Standard environment)
  Personnel: 0.95 (Experienced team)
  Project: 1.02 (Good tools, collocated)

Estimates:
  Effort: 142 person-months
  Duration: 14.2 months
  Team Size: 10 people (average)

  Effort Distribution:
    Requirements: 10 PM (7%)
    Design: 23 PM (16%)
    Implementation: 37 PM (26%)
    Testing: 44 PM (31%)
    Management: 28 PM (20%)

Uncertainty Range:
  Optimistic: 114 PM (11.4 months, 10 people)
  Most Likely: 142 PM (14.2 months, 10 people)
  Pessimistic: 178 PM (17.7 months, 10 people)

Cost Estimate (at $15,000/PM):
  Most Likely: $2,130,000
  Range: $1,710,000 - $2,670,000
```

## Planned Integration Points

### With Supply Chain Analyzer
- Factor in dependency count for complexity
- Consider technical debt for maintenance
- Assess security vulnerabilities impact

### With DORA Metrics
- Use deployment frequency for SCED factor
- Factor in lead time for team experience
- Consider MTTR for quality assessment

### With Code Ownership
- Use bus factor for PCON (personnel continuity)
- Team size from ownership metrics
- Experience levels from contribution history

## Future Enhancements

### Phase 1: Basic Estimation
- [ ] SLOC counting by language
- [ ] Basic COCOMO II calculation
- [ ] Manual cost driver input
- [ ] Single repository support

### Phase 2: Automation
- [ ] Automated complexity analysis
- [ ] Cost driver suggestions
- [ ] Multi-repository support
- [ ] Historical tracking

### Phase 3: AI Enhancement
- [ ] Claude-powered cost driver rating
- [ ] Similar project comparison
- [ ] Risk assessment
- [ ] Scenario modeling

### Phase 4: Advanced Features
- [ ] Local calibration tools
- [ ] Agile/sprint estimation
- [ ] Real-time tracking
- [ ] Integration with project management tools

## Research Areas

- Modern language productivity factors (Rust, Go, TypeScript)
- Cloud/serverless effort multipliers
- AI-assisted development impact on productivity
- DevOps/automation effect on maintenance
- Microservices architecture complexity
- Infrastructure as Code estimation

## Why COCOMO?

**Industry Standard**: Widely accepted, research-backed model
**Algorithmic**: Objective, repeatable estimates
**Calibratable**: Adaptable to organization specifics
**Comprehensive**: Considers multiple project factors
**Proven**: 40+ years of validation and refinement

**Limitations to Address**:
- Subjective factor ratings (AI can help)
- Early estimation difficulty (uncertainty ranges help)
- Model assumptions (document deviations)

## References

- [COCOMO II Model Definition](http://csse.usc.edu/tools/)
- [Software Cost Estimation with COCOMO II](http://www.amazon.com/Software-Cost-Estimation-Cocomo-II/dp/0137025769)
- [RAG Documentation](../../rag/cocomo/)

## Contributing

Once development begins, contributions will be welcome for:
- Language-specific LOC counting improvements
- Calibration data from real projects
- Cost driver rating heuristics
- Integration with other tools

## License

GPL-3.0 - See [LICENSE](../../LICENSE) for details.

## Status

**Current**: 📋 Planned - Specification and research phase
**Timeline**: TBD based on demand and resource availability
**Priority**: Community feedback will guide prioritization

This is a **future capability** - the skill and utilities do not yet exist. This documentation serves as a specification for eventual implementation.
