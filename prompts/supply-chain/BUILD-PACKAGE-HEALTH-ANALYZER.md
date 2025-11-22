<!--
Copyright (c) 2024 Gibson Powers Contributors

SPDX-License-Identifier: GPL-3.0
-->

# BUILD: Package Health Analyzer Implementation Prompt

**Purpose**: Build a comprehensive package health analysis system for identifying risks and operational improvement opportunities across an organization's software packages.

**Status**: 🔬 Experimental - New capability under development

## Overview

Build tools that analyze package health across repositories to identify:
- Version standardization opportunities
- Deprecated packages requiring updates
- Unused packages that can be retired
- Low health score packages posing risks
- Operational best practices gaps

## Architecture Requirements

### Two-Tiered Analysis System

**1. Base Scanner** (`package-health-analyzer.sh`):
- Collects package health data from multiple sources
- Performs basic analysis and scoring
- Generates structured reports
- Fast, automated, suitable for CI/CD

**2. AI-Enhanced Analyzer** (`package-health-analyzer-claude.sh`):
- Deep analysis using Claude
- Contextual recommendations
- Chain of reasoning with other supply chain tools
- Risk assessment and prioritization
- Migration planning and remediation guidance

### Integration with Existing Supply Chain Tools

**Chain of Reasoning Pattern**:
```bash
# The package health analyzer should orchestrate other tools:

1. Generate/use existing SBOM → supply-chain-scanner.sh
2. Analyze vulnerabilities → vulnerability-analyzer.sh
3. Check provenance → provenance-analyzer.sh (if needed)
4. Assess package health → package-health-analyzer.sh (new)
5. AI synthesis → package-health-analyzer-claude.sh (new)
```

## Directory Structure

Create the following structure in `utils/supply-chain/package-health-analysis/`:

```
utils/supply-chain/package-health-analysis/
├── package-health-analyzer.sh          # Base health analyzer
├── package-health-analyzer-claude.sh   # AI-enhanced analyzer
├── compare-analyzers.sh                # Compare base vs AI results
├── lib/
│   ├── deps-dev-client.sh              # deps.dev API client
│   ├── health-scoring.sh               # Health score calculation
│   ├── version-analysis.sh             # Version standardization
│   └── deprecation-checker.sh          # Deprecation detection
├── config.example.json                 # Configuration template
├── README.md                           # Complete documentation
└── CHANGELOG.md                        # Version history
```

Also create skill file in `skills/supply-chain/`:

```
skills/supply-chain/
└── package-health-analyzer.skill       # Claude skill file
```

## Implementation Specifications

### 1. Base Analyzer (`package-health-analyzer.sh`)

**Input Sources**:
- SBOM files (CycloneDX/SPDX)
- Repository package manifests (package.json, requirements.txt, etc.)
- Multiple repositories (org-wide scanning)

**Data Collection**:
```bash
# For each package discovered:
1. Query deps.dev API for package metadata
2. Retrieve OpenSSF Scorecard scores
3. Check deprecation status
4. Gather version information
5. Collect usage statistics (if available)
```

**Analysis Tasks**:

**Health Scoring**:
```bash
# Calculate composite health score (0-100)
Health Score =
  (OpenSSF Score * 0.30) +      # Security posture
  (Maintenance Score * 0.25) +   # Active maintenance
  (Security Score * 0.25) +      # Vulnerability status
  (Freshness Score * 0.10) +     # Version currency
  (Popularity Score * 0.10)      # Community adoption

Thresholds:
  90-100: Excellent
  75-89:  Good
  60-74:  Fair
  40-59:  Poor
  0-39:   Critical
```

**Version Standardization Analysis**:
```bash
# Identify version inconsistencies across repositories
Package: lodash
  repo-1: 4.17.21
  repo-2: 4.17.19
  repo-3: 4.17.21
  repo-4: 3.10.1 ⚠️  Outlier detected

Recommendation: Standardize on 4.17.21 (latest secure version)
```

**Deprecation Detection**:
```bash
# Check for deprecated packages
Sources:
- Package registry deprecation flags
- deps.dev deprecation data
- npm deprecate notices
- PyPI yanked releases

Output:
Package: request (npm)
Status: DEPRECATED
Since: 2020-02-11
Reason: "Package no longer supported"
Alternative: axios, node-fetch
Usage: Found in 5 repositories
```

**Unused Package Detection**:
```bash
# Identify packages that may be unused
Heuristics:
- No code references found (via grep/ast analysis)
- Orphaned dependencies (nothing depends on them)
- Dev dependencies in production manifests
- Leftover from removed features

Note: Base analyzer flags potential unused packages
      AI analyzer provides deeper code analysis
```

**Output Format**:
```json
{
  "scan_metadata": {
    "timestamp": "2024-11-21T10:30:00Z",
    "repositories_scanned": 15,
    "packages_analyzed": 342,
    "analyzer_version": "1.0.0"
  },
  "summary": {
    "total_packages": 342,
    "unique_packages": 87,
    "deprecated_packages": 5,
    "low_health_packages": 12,
    "version_inconsistencies": 23,
    "potential_unused": 8
  },
  "packages": [
    {
      "name": "lodash",
      "ecosystem": "npm",
      "versions_in_use": ["4.17.21", "4.17.19", "3.10.1"],
      "health_score": 85,
      "health_grade": "Good",
      "openssf_score": 7.8,
      "is_deprecated": false,
      "vulnerability_count": 0,
      "usage_count": 12,
      "repositories": ["repo-1", "repo-2", "repo-3"],
      "recommendations": ["Standardize on version 4.17.21"]
    },
    {
      "name": "request",
      "ecosystem": "npm",
      "versions_in_use": ["2.88.0"],
      "health_score": 25,
      "health_grade": "Critical",
      "openssf_score": null,
      "is_deprecated": true,
      "deprecated_since": "2020-02-11",
      "alternative_packages": ["axios", "node-fetch"],
      "usage_count": 5,
      "repositories": ["repo-4", "repo-7"],
      "recommendations": ["Migrate to axios or node-fetch"]
    }
  ],
  "version_inconsistencies": [
    {
      "package": "lodash",
      "versions": {
        "4.17.21": ["repo-1", "repo-2", "repo-3"],
        "4.17.19": ["repo-5"],
        "3.10.1": ["repo-4"]
      },
      "recommended_version": "4.17.21",
      "affected_repositories": 5
    }
  ],
  "deprecated_packages": [
    {
      "package": "request",
      "ecosystem": "npm",
      "deprecated_since": "2020-02-11",
      "usage_count": 5,
      "alternatives": ["axios", "node-fetch"],
      "repositories": ["repo-4", "repo-7"]
    }
  ]
}
```

### 2. AI-Enhanced Analyzer (`package-health-analyzer-claude.sh`)

**Purpose**: Provide deep analysis and actionable recommendations using Claude with chain of reasoning.

**Chain of Reasoning Workflow**:

```bash
#!/bin/bash
# package-health-analyzer-claude.sh

# Step 1: Generate base analysis
echo "Step 1/5: Running base package health analysis..."
BASE_RESULTS=$(./package-health-analyzer.sh "$@")

# Step 2: Integrate vulnerability analysis
echo "Step 2/5: Analyzing vulnerabilities..."
VULN_RESULTS=$(../vulnerability-analysis/vulnerability-analyzer.sh --sbom "$SBOM")

# Step 3: Check provenance (if needed)
echo "Step 3/5: Checking provenance..."
PROV_RESULTS=$(../provenance-analysis/provenance-analyzer.sh --sbom "$SBOM")

# Step 4: Prepare context for Claude
echo "Step 4/5: Preparing analysis context..."
CONTEXT=$(jq -s '.[0] * .[1] * .[2]' \
  <(echo "$BASE_RESULTS") \
  <(echo "$VULN_RESULTS") \
  <(echo "$PROV_RESULTS"))

# Step 5: AI analysis
echo "Step 5/5: Performing AI-enhanced analysis..."
claude_analyze "$CONTEXT"
```

**AI Analysis Tasks**:

**1. Risk Assessment**:
```
Prompt Claude to analyze:
- Which deprecated packages pose highest risk?
- What's the blast radius of each issue?
- What are the migration complexities?
- What are the priority order for remediation?

Output: Risk-ranked list with justifications
```

**2. Version Standardization Strategy**:
```
Prompt Claude to:
- Analyze version inconsistencies
- Consider breaking changes between versions
- Recommend standardization strategy
- Estimate migration effort
- Identify blockers or concerns

Output: Migration plan with phased approach
```

**3. Unused Package Analysis**:
```
Prompt Claude to:
- Review code references (if provided)
- Analyze dependency graphs
- Identify truly unused packages vs. indirect usage
- Assess removal safety

Output: Confident removal candidates with justification
```

**4. Alternative Package Recommendations**:
```
Prompt Claude to:
- Research modern alternatives for deprecated packages
- Compare features and API compatibility
- Consider organization's tech stack
- Evaluate migration paths

Output: Ranked alternatives with migration guides
```

**5. Operational Improvements**:
```
Prompt Claude to:
- Identify patterns in package management
- Recommend best practices
- Suggest policy improvements
- Propose automation opportunities

Output: Strategic recommendations
```

**AI Prompt Template**:

````markdown
# Package Health Analysis

## Context
You are analyzing package health across an organization to identify risks and operational improvements.

## Input Data

### Base Package Health Analysis
```json
{BASE_RESULTS}
```

### Vulnerability Analysis
```json
{VULN_RESULTS}
```

### Provenance Analysis (if available)
```json
{PROV_RESULTS}
```

## Analysis Tasks

### 1. Risk Assessment
Analyze all identified issues and provide:
- Risk ranking (Critical/High/Medium/Low)
- Business impact assessment
- Blast radius (how many repos/services affected)
- Urgency rating

For each high-risk item, explain:
- Why it's risky
- What could go wrong
- Timeline for action

### 2. Version Standardization Strategy
For packages with version inconsistencies:
- Recommended target version
- Breaking changes to consider
- Migration complexity (Simple/Moderate/Complex)
- Phased rollout plan
- Testing requirements

### 3. Deprecated Package Migration
For each deprecated package:
- Top 3 alternative packages with pros/cons
- Feature parity analysis
- API compatibility assessment
- Migration effort estimate (hours/days)
- Sample migration guide

### 4. Unused Package Assessment
For potentially unused packages:
- Confidence level (High/Medium/Low) that it's truly unused
- Rationale for assessment
- Safe removal steps
- Rollback plan if needed

### 5. Health Score Insights
For packages with low health scores:
- Root cause of low score
- Whether to keep, replace, or accept risk
- Monitoring recommendations

### 6. Operational Recommendations
Provide strategic guidance:
- Patterns observed across the organization
- Policy recommendations (version pinning, approval workflows)
- Automation opportunities
- Best practices to adopt

## Output Format

Provide a comprehensive markdown report with:
1. Executive Summary (2-3 paragraphs)
2. Risk Rankings (table format)
3. Detailed Findings (by category)
4. Action Plan (prioritized with effort estimates)
5. Long-term Recommendations

Use clear sections, bullet points, and code examples where helpful.
````

### 3. Integration Skill (`package-health-analyzer.skill`)

**Skill File Structure**:

```markdown
---
name: package-health-analyzer
description: Analyze package health across repositories to identify risks and operational improvements
version: 1.0.0
status: experimental
---

# Package Health Analyzer

Comprehensive analysis of software packages across your organization to identify:
- Deprecated packages requiring migration
- Version inconsistencies to standardize
- Low health score packages posing risks
- Unused packages that can be retired
- Operational best practices gaps

## How to Use This Skill

### Quick Start
```bash
# Analyze single repository
./utils/supply-chain/package-health-analysis/package-health-analyzer-claude.sh \
  --repo owner/repo

# Analyze entire organization
./utils/supply-chain/package-health-analysis/package-health-analyzer-claude.sh \
  --org myorg
```

## What I Can Do

1. **Discover Package Health Issues**
   - Scan repositories for package usage
   - Query deps.dev for health metrics
   - Calculate composite health scores
   - Identify deprecated packages

2. **Analyze Version Standardization**
   - Find version inconsistencies
   - Recommend target versions
   - Assess migration complexity
   - Plan phased rollouts

3. **Detect Unused Packages**
   - Identify potentially unused dependencies
   - Analyze code references
   - Assess removal safety
   - Recommend cleanup actions

4. **Provide Migration Guidance**
   - Research alternative packages
   - Compare features and compatibility
   - Estimate migration effort
   - Generate migration guides

5. **Strategic Recommendations**
   - Identify patterns and trends
   - Recommend operational improvements
   - Suggest automation opportunities
   - Propose policy enhancements

## Chain of Reasoning

This skill integrates with other supply chain tools:

```
1. SBOM Generation (if needed)
   ↓
2. Vulnerability Analysis
   ↓
3. Provenance Check (if needed)
   ↓
4. Package Health Analysis (this skill)
   ↓
5. AI-Enhanced Recommendations
```

## Configuration

See `utils/supply-chain/package-health-analysis/config.example.json`:

```json
{
  "package_health": {
    "health_score_weights": {
      "openssf": 0.30,
      "maintenance": 0.25,
      "security": 0.25,
      "freshness": 0.10,
      "popularity": 0.10
    },
    "thresholds": {
      "excellent": 90,
      "good": 75,
      "fair": 60,
      "poor": 40
    },
    "deprecation_sources": [
      "deps.dev",
      "package_registry"
    ]
  }
}
```

## Example Interaction

**User**: "Analyze package health across my organization and identify the top risks"

**Assistant**:
I'll analyze your organization's packages using the package health analyzer with AI enhancements.

*Runs: package-health-analyzer-claude.sh --org your-org*

Based on the analysis of 342 packages across 15 repositories:

**Top Risks Identified:**

1. **CRITICAL: "request" npm package (5 repos affected)**
   - Deprecated since 2020, no security updates
   - 2 known vulnerabilities (CVE-2023-XXXX)
   - Recommended migration: axios or node-fetch
   - Effort: 2-3 days per repository

2. **HIGH: "lodash" version inconsistency (12 repos affected)**
   - Versions: 3.10.1 (1 repo), 4.17.19 (4 repos), 4.17.21 (7 repos)
   - Version 3.10.1 has known vulnerabilities
   - Standardize on 4.17.21
   - Effort: 4-8 hours per repo

[... detailed findings continue ...]

**Recommended Action Plan:**

Week 1-2: Migrate away from "request" in critical services
Week 3-4: Standardize lodash versions
Week 5-6: Address remaining low-health packages

Would you like detailed migration guides for any of these packages?

## When to Use This Skill

- During security audits
- Before major releases
- As part of tech debt reduction initiatives
- When onboarding new teams
- For compliance reporting
- During dependency updates

## Integration with CI/CD

```yaml
# .github/workflows/package-health.yml
name: Package Health Check

on:
  schedule:
    - cron: '0 0 * * 0'  # Weekly
  workflow_dispatch:

jobs:
  health-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2

      - name: Run Package Health Analysis
        run: |
          ./utils/supply-chain/package-health-analysis/package-health-analyzer.sh \
            --repo ${{ github.repository }} \
            --format json > health-report.json

      - name: Check for Critical Issues
        run: |
          CRITICAL=$(jq '.summary.critical_packages' health-report.json)
          if [ "$CRITICAL" -gt 0 ]; then
            echo "::error::Found $CRITICAL critical package health issues"
            exit 1
          fi
```

## References

- [Requirements Document](../../prompts/supply-chain/package-health-analyzer-requirements.md)
- [deps.dev API Documentation](../../rag/supply-chain/package-health/deps-dev-api.md)
- [Best Practices](../../rag/supply-chain/package-health/package-management-best-practices.md)
- [Supply Chain Scanner](./supply-chain-analyzer.skill)
```

## Implementation Checklist

### Phase 1: Base Analyzer (Week 1-2)
- [ ] Create directory structure
- [ ] Implement deps.dev API client (`lib/deps-dev-client.sh`)
- [ ] Build health scoring engine (`lib/health-scoring.sh`)
- [ ] Implement base analyzer script (`package-health-analyzer.sh`)
- [ ] Add multi-repo scanning support
- [ ] Create JSON output format
- [ ] Write basic tests

### Phase 2: Analysis Libraries (Week 2-3)
- [ ] Build version analysis module (`lib/version-analysis.sh`)
- [ ] Implement deprecation checker (`lib/deprecation-checker.sh`)
- [ ] Add unused package detection
- [ ] Create comparison tool (`compare-analyzers.sh`)
- [ ] Add configuration system
- [ ] Write documentation

### Phase 3: AI Enhancement (Week 3-4)
- [ ] Implement Claude integration (`package-health-analyzer-claude.sh`)
- [ ] Build chain of reasoning workflow
- [ ] Create AI prompt templates
- [ ] Add context preparation logic
- [ ] Implement recommendation generation
- [ ] Test with real scenarios

### Phase 4: Skill Integration (Week 4)
- [ ] Create skill file (`package-health-analyzer.skill`)
- [ ] Write comprehensive README
- [ ] Add usage examples
- [ ] Create CHANGELOG
- [ ] Update main README to reference new capability
- [ ] CI/CD integration examples

### Phase 5: Testing & Documentation (Week 5)
- [ ] End-to-end testing with real repositories
- [ ] Performance optimization
- [ ] Error handling improvements
- [ ] Complete documentation review
- [ ] Example scenarios and walkthroughs

## Key Design Principles

### 1. Leverage Existing Infrastructure
```bash
# Reuse SBOM generation
if [ ! -f "$SBOM_FILE" ]; then
  echo "Generating SBOM..."
  syft scan --output cyclonedx-json="$SBOM_FILE" "$TARGET"
fi

# Reuse vulnerability analysis
VULNS=$(../vulnerability-analysis/vulnerability-analyzer.sh --sbom "$SBOM_FILE")

# Build on existing config system
CONFIG=$(../../lib/config-loader.sh load package-health-analysis)
```

### 2. Chain of Reasoning Pattern
```bash
# Each analyzer should:
1. Accept input from previous stage
2. Perform its specific analysis
3. Output structured JSON
4. Pass context to next stage

# Example:
SBOM → Vulnerabilities → Provenance → Health → AI Analysis
  ↓         ↓              ↓           ↓          ↓
JSON     + JSON        + JSON      + JSON    = Final Report
```

### 3. Graceful Degradation
```bash
# If deps.dev API fails:
- Fall back to package registry metadata
- Continue with available data
- Mark items as "needs manual review"
- Don't fail entire analysis
```

### 4. Progressive Enhancement
```bash
# Base analyzer: Fast, automated, basic insights
./package-health-analyzer.sh  # 30 seconds

# AI analyzer: Deeper, contextual, actionable
./package-health-analyzer-claude.sh  # 2-3 minutes
```

## API Integration Requirements

### deps.dev API Client

```bash
# lib/deps-dev-client.sh

get_package_info() {
  local system=$1  # npm, pypi, cargo, maven
  local package=$2

  curl -s "https://api.deps.dev/v3alpha/systems/${system}/packages/${package}"
}

get_package_version() {
  local system=$1
  local package=$2
  local version=$3

  curl -s "https://api.deps.dev/v3alpha/systems/${system}/packages/${package}/versions/${version}"
}

get_openssf_scorecard() {
  local package_info=$1

  echo "$package_info" | jq -r '.scorecard // null'
}

check_deprecation() {
  local package_info=$1

  echo "$package_info" | jq -r '.deprecated // false'
}
```

## Testing Strategy

### Unit Tests
```bash
# Test individual components
tests/
├── test-deps-dev-client.sh
├── test-health-scoring.sh
├── test-version-analysis.sh
└── test-deprecation-checker.sh
```

### Integration Tests
```bash
# Test with sample repositories
test-repos/
├── simple-node/          # Basic npm project
├── complex-python/       # Python with many deps
└── multi-language/       # Mixed ecosystems
```

### End-to-End Tests
```bash
# Test complete workflows
e2e-tests/
├── test-single-repo.sh
├── test-org-scan.sh
└── test-ai-analysis.sh
```

## Success Criteria

**Base Analyzer**:
- [ ] Scans 100+ packages in < 2 minutes
- [ ] Accurate health scoring (validated against manual assessment)
- [ ] Deprecation detection for major ecosystems (npm, PyPI, Maven)
- [ ] Version inconsistency detection across repos
- [ ] Clean JSON output format

**AI Analyzer**:
- [ ] Provides actionable recommendations
- [ ] Accurate risk assessment
- [ ] Practical migration guides
- [ ] Strategic insights beyond base analysis
- [ ] Clear, well-structured reports

**Skill Integration**:
- [ ] Works seamlessly with other supply chain skills
- [ ] Clear documentation and examples
- [ ] CI/CD integration examples
- [ ] Configuration system integration

## Example Use Cases

### Use Case 1: Security Audit
```bash
# Find all packages with security concerns
./package-health-analyzer-claude.sh \
  --org myorg \
  --filter "deprecated=true OR health_score<60" \
  --output security-audit.md
```

### Use Case 2: Version Standardization
```bash
# Identify version inconsistencies
./package-health-analyzer.sh \
  --org myorg \
  --analyze-versions \
  --output version-report.json
```

### Use Case 3: Tech Debt Reduction
```bash
# Full analysis with recommendations
./package-health-analyzer-claude.sh \
  --org myorg \
  --include-unused \
  --include-recommendations \
  --output tech-debt-plan.md
```

## Next Steps After Implementation

1. **Gather Feedback**: Run on real organizations, collect user feedback
2. **Iterate**: Improve accuracy and recommendations based on usage
3. **Expand Ecosystems**: Add support for more package ecosystems
4. **Policy Engine**: Add ability to define and enforce policies
5. **Dashboard**: Consider web UI for visualization
6. **Automation**: Build automated remediation where safe

## Questions for Implementation

1. **Ecosystem Priority**: Start with npm, PyPI, both, or more?
2. **API Rate Limits**: How to handle deps.dev rate limits at scale?
3. **Caching Strategy**: How long to cache deps.dev responses?
4. **Parallel Processing**: Should we parallelize API calls?
5. **Output Format**: Prefer JSON, Markdown, both, or user choice?
6. **Integration Points**: Any specific CI/CD systems to prioritize?

## References

- [Complete Requirements](./package-health-analyzer-requirements.md)
- [deps.dev API Reference](../../rag/supply-chain/package-health/deps-dev-api.md)
- [Package Management Best Practices](../../rag/supply-chain/package-health/package-management-best-practices.md)
- [Supply Chain Scanner](../../utils/supply-chain/supply-chain-scanner.sh)
- [Vulnerability Analyzer](../../utils/supply-chain/vulnerability-analysis/)

---

**Implementation Status**: Ready to begin
**Estimated Effort**: 4-5 weeks for full implementation
**Complexity**: Moderate - leverages existing infrastructure
**Priority**: High - addresses operational pain points
