<!--
Copyright (c) 2024 Crash Override Inc
101 Fulton St, 416, New York 10038
SPDX-License-Identifier: GPL-3.0
-->

# Code Ownership Prompts

Comprehensive prompt templates for code ownership analysis, CODEOWNERS validation, knowledge transfer planning, and review optimization using the Code Ownership Analysis skill.

## Directory Structure

```
prompts/code-ownership/
├── audit/              # Repository ownership audits
│   ├── analyze-repository.md
│   └── health-assessment.md
├── validation/         # CODEOWNERS file validation
│   ├── validate-codeowners.md
│   └── generate-codeowners.md
├── planning/           # Knowledge transfer and succession
│   ├── knowledge-transfer.md
│   └── succession-planning.md
└── optimization/       # Review process optimization
    └── recommend-reviewers.md
```

## Categories

### 🔍 Audit
Repository-level ownership analysis and health assessments.

- **[analyze-repository.md](audit/analyze-repository.md)** - Complete repository ownership audit
- **[health-assessment.md](audit/health-assessment.md)** - Ownership health scoring and trends

### ✅ Validation
CODEOWNERS file validation, generation, and maintenance.

- **[validate-codeowners.md](validation/validate-codeowners.md)** - Syntax and accuracy validation
- **[generate-codeowners.md](validation/generate-codeowners.md)** - Auto-generate CODEOWNERS from git history

### 📋 Planning
Knowledge transfer and succession planning for team transitions.

- **[knowledge-transfer.md](planning/knowledge-transfer.md)** - Plan transfers for departing team members
- **[succession-planning.md](planning/succession-planning.md)** - Identify risks and prepare successors

### ⚡ Optimization
Code review process optimization and reviewer recommendations.

- **[recommend-reviewers.md](optimization/recommend-reviewers.md)** - PR reviewer suggestions based on ownership

## Quick Start

1. **Load the Code Ownership Analysis Skill** in Crash Override
2. **Choose a category** based on your needs (Audit, Validation, Planning, Optimization)
3. **Select a prompt template** for your specific task
4. **Copy and customize** the prompt with your repository details
5. **Execute and review** the analysis

## Common Workflows

### Quarterly Ownership Audit
```
1. analyze-repository.md       → Full repository audit
2. validate-codeowners.md       → Check CODEOWNERS accuracy
3. health-assessment.md         → Track trends vs last quarter
```

### New Team Member Onboarding
```
1. analyze-repository.md        → Show ownership structure
2. health-assessment.md         → Explain ownership health
```

### Team Member Departure
```
1. knowledge-transfer.md        → Create transfer plan
2. succession-planning.md       → Identify successors
3. generate-codeowners.md       → Update CODEOWNERS
```

### PR Review Assignment
```
1. recommend-reviewers.md       → Get reviewer suggestions for specific PR
```

### CODEOWNERS Maintenance
```
1. validate-codeowners.md       → Check current file
2. generate-codeowners.md       → Generate updated version
```

## Use Cases by Role

### Engineering Managers
- **Monthly**: health-assessment for team metrics
- **Quarterly**: analyze-repository for comprehensive review
- **Departures**: knowledge-transfer planning
- **As Needed**: succession-planning for risk mitigation

### DevOps/Platform Teams
- **CI/CD Integration**: validate-codeowners in pipelines
- **Automation**: generate-codeowners for new repos
- **Monitoring**: health-assessment for metrics

### Individual Contributors
- **PR Creation**: recommend-reviewers for finding right reviewers
- **Code Questions**: analyze-repository to find owners

### Technical Leads
- **Team Health**: health-assessment for ownership distribution
- **Transitions**: knowledge-transfer and succession-planning
- **Process Improvement**: validate-codeowners for accuracy

## Best Practices

### Regular Audits
- **Monthly**: Quick health check (health-assessment)
- **Quarterly**: Full audit (analyze-repository + validate-codeowners)
- **Annually**: Comprehensive review with strategic planning

### CODEOWNERS Maintenance
- **On Team Changes**: Update within 1 week of join/leave/transfer
- **Quarterly Review**: Validate accuracy against contributions
- **Version Control**: Treat CODEOWNERS changes like code (review, test)

### Knowledge Transfer
- **Start Early**: Begin planning 4-6 weeks before departure
- **Prioritize**: Focus on critical and high-risk components first
- **Document**: Create artifacts during transfer process
- **Verify**: Ensure successor can operate independently

### Metrics and Monitoring
- **Track Trends**: Monitor coverage, distribution, staleness over time
- **Set Thresholds**: Alert on coverage <70%, Gini >0.5, inactive owners >10%
- **Make Visible**: Dashboard key metrics
- **Celebrate Wins**: Recognize improvements

## Prerequisites

- Code Ownership Analysis skill loaded in Crash Override
- Repository access (for git history analysis)
- Optional: CODEOWNERS file (for validation)

## Data Sources

Prompts leverage git repository data:
- Commit history (authors, dates, files changed)
- PR and review data (if available via API)
- CODEOWNERS file (if present)
- Branch and merge patterns

## Output Formats

Prompts generate:
- Markdown reports with metrics and recommendations
- JSON/CSV data for export
- CODEOWNERS files (validated or generated)
- Knowledge transfer plans with timelines
- Risk assessments with priorities

## Related Resources

### Skills
- [Code Ownership Analysis](../../skills/code-ownership/) - Complete skill documentation
- [Examples](../../skills/code-ownership/examples/) - Sample reports and outputs

### Automation
- [ownership-analyzer.sh](../../skills/code-ownership/ownership-analyzer.sh) - Basic CLI analysis
- [ownership-analyzer-claude.sh](../../skills/code-ownership/ownership-analyzer-claude.sh) - AI-enhanced CLI

## Contributing

Have a useful ownership prompt? Please contribute!

1. Choose the appropriate category (audit/validation/planning/optimization)
2. Follow the template structure in existing prompts
3. Include purpose, usage, examples, and variations
4. Test thoroughly before submitting
5. Submit a pull request

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for guidelines.

## Support

For questions or issues:
- Review [Code Ownership skill documentation](../../skills/code-ownership/README.md)
- Check existing [discussions](https://github.com/crashappsec/skills-and-prompts-and-rag/discussions)
- Open an [issue](https://github.com/crashappsec/skills-and-prompts-and-rag/issues)
- Contact: mark@crashoverride.com

---

**Improve your code ownership practices with proven prompt templates!**
