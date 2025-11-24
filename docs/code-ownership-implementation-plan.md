<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Code Ownership Analysis Enhancement Plan

**Date**: 2025-11-24
**Status**: Planning Phase
**Objective**: Build comprehensive bus factor, risk identification, and knowledge distribution analysis

## Executive Summary

This plan enhances the code ownership analyzer to provide:
1. **Bus Factor Analysis** - Standard analyzer (non-Claude) with 6-criteria SPOF detection
2. **Risk Identification** - Claude-powered risk analysis using RAG patterns
3. **Knowledge Distribution Analysis** - Statistical analysis of knowledge silos and concentration

## Current State Assessment

### What Exists ✅
- Basic ownership analysis (commit-based + line-based)
- Simple Claude integration (basic prompt)
- Library modules (metrics, succession, analyser-core)
- Comprehensive RAG documentation (bus-factor-analysis.md, ownership-metrics.md)
- Configuration system
- Historical tracking

### What's Missing ❌
- **Comprehensive bus factor calculation** using 6-criteria SPOF assessment
- **Deep risk identification** leveraging Claude + RAG patterns
- **Knowledge distribution metrics** (Gini coefficient, concentration analysis)
- **Succession planning integration** with risk scoring
- **Regression testing** for library modifications

### Gap Analysis

Current Claude integration (line 876-950 in ownership-analyser.sh):
- Basic prompt asking for generic insights
- Doesn't use RAG patterns
- No structured risk assessment
- Missing SPOF 6-criteria analysis
- No knowledge distribution calculations

## Architecture Design

### Component Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                  ownership-analyser.sh                       │
│                    (Main Orchestrator)                       │
└─────────────────────────────────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
        ▼                   ▼                   ▼
┌──────────────┐  ┌──────────────────┐  ┌──────────────────┐
│   Standard   │  │   Claude-Enhanced │  │   Knowledge      │
│  Bus Factor  │  │  Risk Analysis    │  │  Distribution    │
│   Analyzer   │  │                   │  │   Analyzer       │
└──────────────┘  └──────────────────┘  └──────────────────┘
        │                   │                   │
        ▼                   ▼                   ▼
┌──────────────────────────────────────────────────────────┐
│              Library Modules (lib/)                       │
│  • metrics.sh - Calculations                             │
│  • spof-detector.sh - NEW: 6-criteria SPOF detection     │
│  • knowledge-analysis.sh - NEW: Distribution metrics     │
│  • succession.sh - Succession planning                   │
│  • claude-integration.sh - NEW: RAG-powered analysis     │
└──────────────────────────────────────────────────────────┘
```

### Data Flow

```
1. Git Repository → analyser-core.sh → Raw Ownership Data
                                              │
2. Raw Data → metrics.sh → Basic Metrics (coverage, health)
                                              │
3. Basic Metrics → spof-detector.sh → Bus Factor + SPOF List
                                              │
4. SPOF List → knowledge-analysis.sh → Distribution Analysis
                                              │
5. All Data + RAG → claude-integration.sh → Risk Assessment
                                              │
6. Combined Results → Output (JSON/Markdown/CSV)
```

## Implementation Plan

### Phase 1: Standard Bus Factor Analyzer (No Claude Required)

**Objective**: Implement research-backed bus factor calculation as a standard analyzer

#### 1.1 Create `lib/spof-detector.sh`

**Location**: `/Users/curphey/Documents/GitHub/gibson-powers/utils/code-ownership/lib/spof-detector.sh`

**Functions**:
```bash
# Calculate overall bus factor for repository
calculate_bus_factor() {
    local repo_path="$1"
    local ownership_data="$2"
    # Implementation based on RAG: rag/code-ownership/bus-factor-analysis.md
    # Algorithm:
    # 1. For each file, identify contributors with >10% ownership
    # 2. Sort contributors by total files owned
    # 3. Simulate removing contributors (highest impact first)
    # 4. Bus factor = number removed before >20% files have no expert
}

# Detect single points of failure using 6 criteria
detect_spofs() {
    local repo_path="$1"
    local ownership_data="$2"
    # Criteria from RAG (bus-factor-analysis.md lines 106-118):
    # 1. Single contributor ever
    # 2. Critical path (auth, payments, core)
    # 3. High complexity (>500 LOC)
    # 4. No backup owner (>15% knowledge)
    # 5. Low test coverage (<60%)
    # 6. No documentation

    # Returns: JSON array of SPOFs with risk levels
}

# Calculate component-level bus factor
calculate_component_bus_factor() {
    local component_path="$1"
    local ownership_data="$2"
    # Per-component bus factor calculation
}

# Assess concentration risk
calculate_concentration_risk() {
    local ownership_data="$1"
    # From RAG (bus-factor-analysis.md lines 155-170):
    # - Top 1 person owns >20% = high risk
    # - Top 3 people own >50% = high risk
    # - Returns risk score 0.0-1.0
}

# Calculate departure risk for contributors
calculate_departure_risk() {
    local contributor="$1"
    local activity_data="$2"
    # From RAG (bus-factor-analysis.md lines 172-189):
    # High: Known departure, extended absence (score 1.0)
    # Medium: Declining activity, no recent contributions (score 0.5)
    # Low: Normal activity (score 0.2)
}

# Prioritize SPOFs for succession planning
prioritize_spofs() {
    local spof_data="$1"
    # From RAG (bus-factor-analysis.md lines 238-246):
    # priority = criticality*0.4 + concentration*0.3 + departure*0.2 + docs_gap*0.1
}
```

**RAG References**:
- `rag/code-ownership/bus-factor-analysis.md` - Complete algorithms and thresholds
- Lines 14-65: Bus factor calculation
- Lines 105-151: SPOF detection criteria
- Lines 154-233: Risk assessment framework

#### 1.2 Create `lib/knowledge-analysis.sh`

**Location**: `/Users/curphey/Documents/GitHub/gibson-powers/utils/code-ownership/lib/knowledge-analysis.sh`

**Functions**:
```bash
# Calculate Gini coefficient for ownership distribution
calculate_gini_coefficient() {
    local ownership_data="$1"
    # From RAG (ownership-metrics.md):
    # Gini >0.7 = dangerous concentration
    # Gini 0.5-0.7 = moderate concentration
    # Gini <0.5 = healthy distribution
}

# Calculate top-N concentration metrics
calculate_top_n_concentration() {
    local ownership_data="$1"
    local n="$2"
    # Top 1, Top 3, Top 5 contributor percentages
}

# Identify knowledge silos
identify_knowledge_silos() {
    local ownership_data="$1"
    # Areas where knowledge is isolated to single person/team
    # No knowledge overlap with other contributors
}

# Calculate backup coverage
calculate_backup_coverage() {
    local ownership_data="$1"
    # For each primary owner, check if backup exists with >15% knowledge
}

# Analyze knowledge overlap
analyze_knowledge_overlap() {
    local ownership_data="$1"
    # Which contributors share knowledge in same files?
    # Collaboration patterns
}
```

**RAG References**:
- `rag/code-ownership/ownership-metrics.md` - Lines 76-100 (distribution metrics)
- `docs/code-ownership-improvements.md` - Lines 119-134 (Gini coefficient thresholds)

#### 1.3 Integrate into Main Analyzer

**Modify**: `utils/code-ownership/ownership-analyser.sh`

**Changes**:
```bash
# Add after line 108 (check_prerequisites):
source "$SCRIPT_DIR/lib/spof-detector.sh"
source "$SCRIPT_DIR/lib/knowledge-analysis.sh"

# New function to run standard analysis
run_standard_analysis() {
    local repo_path="$1"
    local ownership_data="$2"

    # 1. Calculate bus factor
    local bus_factor=$(calculate_bus_factor "$repo_path" "$ownership_data")

    # 2. Detect SPOFs
    local spofs=$(detect_spofs "$repo_path" "$ownership_data")

    # 3. Calculate knowledge distribution
    local gini=$(calculate_gini_coefficient "$ownership_data")
    local concentration=$(calculate_top_n_concentration "$ownership_data" 3)

    # 4. Identify knowledge silos
    local silos=$(identify_knowledge_silos "$ownership_data")

    # 5. Calculate backup coverage
    local backup=$(calculate_backup_coverage "$ownership_data")

    # 6. Combine results into JSON
    generate_standard_report "$bus_factor" "$spofs" "$gini" "$concentration" "$silos" "$backup"
}
```

### Phase 2: Claude-Enhanced Risk Identification

**Objective**: Leverage Claude with RAG patterns for intelligent risk analysis

#### 2.1 Create `lib/claude-integration.sh`

**Location**: `/Users/curphey/Documents/GitHub/gibson-powers/utils/code-ownership/lib/claude-integration.sh`

**Functions**:
```bash
# Load RAG patterns for context
load_rag_patterns() {
    local rag_dir="$1"
    # Load relevant RAG documents:
    # - bus-factor-analysis.md
    # - ownership-metrics.md
    # - codeowners-best-practices.md
}

# Build comprehensive analysis prompt
build_analysis_prompt() {
    local standard_analysis="$1"
    local rag_context="$2"
    # Combine standard analysis results + RAG patterns
    # See section 2.2 for detailed prompt
}

# Analyze with Claude using RAG context
analyze_with_claude_rag() {
    local standard_analysis="$1"
    local rag_context="$2"
    # Call Claude API with comprehensive prompt
    # Parse structured response
}

# Extract risk recommendations
extract_risk_recommendations() {
    local claude_response="$1"
    # Parse Claude response into structured format:
    # - Critical risks (immediate action)
    # - High risks (2 weeks)
    # - Medium risks (quarter)
    # - Low risks (monitor)
}

# Generate succession planning recommendations
generate_succession_recommendations() {
    local spof_data="$1"
    local claude_insights="$2"
    # Combine SPOF analysis + Claude insights
    # Use succession.sh library
}
```

#### 2.2 Enhanced Claude Prompts

**Prompt 1: Risk Identification**

**Location**: Create new file `prompts/code-ownership/analysis/risk-identification.md`

```markdown
# Prompt: Identify Code Ownership Risks

## Context
You are analyzing code ownership patterns to identify risks that could impact project continuity and team effectiveness.

## Input Data
You will receive:
1. Standard analysis results (bus factor, SPOFs, knowledge distribution)
2. RAG context with research-backed patterns and thresholds
3. Repository metadata (size, team size, activity level)

## Analysis Framework

### 1. Bus Factor Risk Assessment
Using the SPOF data and 6-criteria assessment, identify:
- **Critical Risks** (all 6 criteria met): Immediate action required
- **High Risks** (4-5 criteria): Action within 2 weeks
- **Medium Risks** (2-3 criteria): Action within quarter
- **Low Risks** (1 criterion): Monitor

For each risk:
- File/component path
- Current owner(s)
- Risk level and specific criteria met
- Business impact assessment (1-10 based on component criticality)
- Why this is a risk (evidence from data)

### 2. Knowledge Distribution Analysis
Analyze the Gini coefficient and concentration metrics:
- If Gini >0.7: "Dangerous concentration - knowledge highly centralized"
- If Top 1 >20%: "High risk - single person has excessive ownership"
- If Top 3 >50%: "Moderate risk - limited knowledge distribution"
- Identify specific knowledge silos and their impact

### 3. Succession Planning Priorities
Using the departure risk scores and activity data:
- Identify contributors with declining activity (>50% drop)
- Flag inactive owners (>90 days)
- Prioritize based on: criticality × concentration × departure risk × docs gap
- List top 10 areas needing succession planning

### 4. Collaboration & Knowledge Sharing Opportunities
- Identify files where knowledge overlap is low (<2 contributors)
- Suggest mentorship pairings (expert + learner on shared files)
- Recommend code review rotation strategies

## Output Format
```json
{
  "risk_summary": {
    "critical_count": 0,
    "high_count": 0,
    "medium_count": 0,
    "low_count": 0
  },
  "critical_risks": [
    {
      "path": "src/auth/oauth.ts",
      "risk_level": "critical",
      "criteria_met": 6,
      "owner": "alice@example.com",
      "business_impact": 10,
      "evidence": {
        "single_contributor": true,
        "critical_path": "authentication",
        "complexity_loc": 847,
        "backup_owner": false,
        "test_coverage": 0.42,
        "documentation": false
      },
      "impact_assessment": "Authentication system is business-critical. Single owner with no backup creates severe continuity risk.",
      "succession_priority": 0.95
    }
  ],
  "knowledge_distribution": {
    "gini_coefficient": 0.73,
    "assessment": "Dangerous concentration",
    "top_1_percentage": 24.5,
    "top_3_percentage": 58.2,
    "silos_identified": [
      {
        "area": "src/auth/",
        "owner": "alice@example.com",
        "overlap": 0.08,
        "risk": "High isolation"
      }
    ]
  },
  "succession_priorities": [
    {
      "area": "src/auth/oauth.ts",
      "current_owner": "alice@example.com",
      "priority_score": 0.95,
      "recommended_successor": "bob@example.com",
      "reasoning": "Bob has 12% familiarity, active in related auth files",
      "estimated_transfer_time": "6-8 weeks",
      "transfer_complexity": "high"
    }
  ],
  "collaboration_opportunities": [
    {
      "type": "mentorship",
      "mentor": "alice@example.com",
      "mentee": "charlie@example.com",
      "shared_files": 15,
      "focus_area": "src/auth/"
    }
  ]
}
```

## Rules
- Base all assessments on data and RAG thresholds
- Use research-backed criteria (6-criteria SPOF, Gini thresholds, etc.)
- Prioritize by business impact × risk score
- Provide evidence for every risk identified
- Be specific with file paths and owner names
```

**Prompt 2: Knowledge Distribution Analysis**

**Location**: Create `prompts/code-ownership/analysis/knowledge-distribution.md`

```markdown
# Prompt: Analyze Knowledge Distribution

## Purpose
Deep analysis of how knowledge is distributed across the team to identify concentration risks and collaboration opportunities.

## Input
- Ownership data (who owns what, to what degree)
- Contributor activity data (commits, reviews, dates)
- Repository structure (directories, components)

## Analysis

### 1. Distribution Metrics
Calculate and interpret:
- **Gini Coefficient**: Overall concentration measure
- **Top-N Concentration**: Percentage owned by top 1, 3, 5 contributors
- **Backup Coverage**: % of files with backup owners (>15% knowledge)
- **Knowledge Overlap**: Average number of contributors per file

### 2. Silo Detection
Identify areas where:
- Single contributor with >80% ownership
- No backup contributor with >15% knowledge
- Limited cross-team collaboration
- Knowledge not spreading despite time

### 3. Collaboration Patterns
Analyze:
- Which contributors frequently work on same files?
- Are there natural mentorship opportunities?
- Is knowledge transfer happening organically?
- Are teams isolated or collaborating?

### 4. Trend Analysis (if historical data available)
- Is distribution improving or worsening?
- Are silos forming or dissolving?
- Is new knowledge spreading to more people?

## Output
Structured analysis with:
- Quantitative metrics
- Risk assessment
- Collaboration opportunities
- Mentorship recommendations
```

#### 2.3 Modify Claude Integration in Main Analyzer

**Location**: `utils/code-ownership/ownership-analyser.sh`

**Modify function** `analyze_with_claude()` (currently line 875):

```bash
analyze_with_claude() {
    local standard_analysis="$1"

    # Load new library
    source "$SCRIPT_DIR/lib/claude-integration.sh"

    # Load RAG patterns
    local rag_context=$(load_rag_patterns "$SCRIPT_DIR/../../rag/code-ownership")

    # Build comprehensive prompt using risk-identification.md template
    local prompt=$(build_analysis_prompt "$standard_analysis" "$rag_context")

    # Call Claude with enhanced prompt
    local response=$(analyze_with_claude_rag "$prompt")

    # Extract structured risk data
    local risks=$(extract_risk_recommendations "$response")

    # Generate succession recommendations
    local succession=$(generate_succession_recommendations "$standard_analysis" "$response")

    # Combine with standard analysis
    merge_analyses "$standard_analysis" "$risks" "$succession"
}
```

### Phase 3: Integration & Testing

#### 3.1 Output Format Integration

**Modify**: Output generation functions to include new analyses

**JSON Output Structure**:
```json
{
  "metadata": {
    "version": "4.0.0",
    "analysis_date": "2025-11-24",
    "repository": "example-repo",
    "analysis_mode": "standard|claude-enhanced"
  },
  "standard_analysis": {
    "bus_factor": {
      "overall": 2,
      "by_component": {
        "src/auth": 1,
        "src/api": 3
      }
    },
    "spofs": [
      {
        "path": "src/auth/oauth.ts",
        "criteria_met": 6,
        "risk_level": "critical",
        "owner": "alice@example.com"
      }
    ],
    "knowledge_distribution": {
      "gini_coefficient": 0.73,
      "top_1_percentage": 24.5,
      "top_3_percentage": 58.2,
      "backup_coverage": 0.42,
      "silos_count": 5
    }
  },
  "claude_analysis": {
    "risk_summary": { ... },
    "critical_risks": [ ... ],
    "succession_priorities": [ ... ],
    "collaboration_opportunities": [ ... ]
  }
}
```

#### 3.2 Markdown Report Template

**Location**: Update `lib/markdown.sh`

**Add sections**:
```markdown
## Bus Factor Analysis
**Overall Bus Factor**: 2 ⚠️

This means the project would be at critical risk if 2 key contributors left.

### Component-Level Bus Factors
| Component | Bus Factor | Status |
|-----------|------------|--------|
| src/auth  | 1          | 🚨 Critical |
| src/api   | 3          | ✅ Healthy |

## Single Points of Failure
| File | Owner | Risk Level | Criteria Met |
|------|-------|------------|--------------|
| src/auth/oauth.ts | alice@example.com | 🚨 Critical | 6/6 |

## Knowledge Distribution
**Gini Coefficient**: 0.73 (Dangerous Concentration)
**Top 1 Contributor**: 24.5% of codebase
**Top 3 Contributors**: 58.2% of codebase

### Knowledge Silos
- **src/auth/** - Owned 92% by Alice (high isolation)

## Succession Planning Priorities
1. **src/auth/oauth.ts** (Priority: 0.95)
   - Current: alice@example.com
   - Recommended Successor: bob@example.com
   - Transfer Time: 6-8 weeks
```

### Phase 4: Regression Testing Plan

**Objective**: Ensure modifications to global libraries don't break existing functionality

#### 4.1 Test Strategy

**Test Levels**:
1. **Unit Tests** - Individual function testing
2. **Integration Tests** - Library integration testing
3. **End-to-End Tests** - Complete analyzer workflow
4. **Regression Tests** - Existing functionality verification

#### 4.2 New Test Files

**Location**: `utils/code-ownership/tests/`

**Create**:
1. `test-spof-detector.sh` - Unit tests for SPOF detection
2. `test-knowledge-analysis.sh` - Unit tests for knowledge distribution
3. `test-claude-integration.sh` - Integration tests for Claude analysis
4. `test-regression.sh` - Regression test suite

**Example**: `test-spof-detector.sh`
```bash
#!/bin/bash
# Test SPOF Detection Functions

source ../lib/spof-detector.sh

test_bus_factor_calculation() {
    # Test with known dataset
    # Expected: Bus factor = 2
    result=$(calculate_bus_factor "test-repo" "test-data.json")
    assert_equals "$result" "2"
}

test_spof_detection_6_criteria() {
    # Test with file meeting all 6 criteria
    # Expected: Critical SPOF identified
    result=$(detect_spofs "test-repo" "test-data.json")
    assert_contains "$result" "critical"
}

test_concentration_risk() {
    # Test with top contributor owning 25%
    # Expected: High risk (>20% threshold)
    result=$(calculate_concentration_risk "test-data.json")
    assert_greater_than "$result" "0.7"
}

# Run all tests
run_all_tests
```

#### 4.3 Regression Test Scenarios

**Scenarios to Test**:
1. ✅ Basic ownership analysis still works (no Claude)
2. ✅ CODEOWNERS validation unchanged
3. ✅ JSON output structure backward compatible
4. ✅ Existing metrics calculations unchanged
5. ✅ Historical trend tracking unaffected
6. ✅ Multi-repo analysis still works
7. ✅ CSV/Markdown export unchanged

**Test Data**:
- Use existing test repository
- Create synthetic datasets with known properties
- Test edge cases (empty repos, single contributor, etc.)

#### 4.4 CI/CD Integration

**Location**: `.github/workflows/test-code-ownership.yml`

```yaml
name: Code Ownership Tests

on:
  push:
    paths:
      - 'utils/code-ownership/**'
      - 'lib/codeowners-validator.sh'
  pull_request:
    paths:
      - 'utils/code-ownership/**'

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0

      - name: Install dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y jq bc

      - name: Run unit tests
        run: |
          cd utils/code-ownership/tests
          ./test-spof-detector.sh
          ./test-knowledge-analysis.sh

      - name: Run integration tests
        run: |
          cd utils/code-ownership/tests
          ./test-integration.sh

      - name: Run regression tests
        run: |
          cd utils/code-ownership/tests
          ./test-regression.sh

      - name: Test Claude integration (if API key available)
        if: ${{ secrets.ANTHROPIC_API_KEY }}
        env:
          ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
        run: |
          cd utils/code-ownership/tests
          ./test-claude-integration.sh
```

## Implementation Timeline

### Week 1: Foundation
- **Day 1-2**: Create `lib/spof-detector.sh` with bus factor calculation
- **Day 3-4**: Create `lib/knowledge-analysis.sh` with distribution metrics
- **Day 5**: Integration into main analyzer (standard mode)

### Week 2: Claude Enhancement
- **Day 1-2**: Create `lib/claude-integration.sh` with RAG loading
- **Day 3**: Write enhanced prompts (risk-identification.md, knowledge-distribution.md)
- **Day 4-5**: Modify Claude integration in main analyzer

### Week 3: Testing & Documentation
- **Day 1-2**: Create unit test suites
- **Day 3**: Create regression test suite
- **Day 4**: CI/CD integration
- **Day 5**: Documentation updates

## Dependencies & Prerequisites

### Required Libraries (New)
- `lib/spof-detector.sh` - NEW
- `lib/knowledge-analysis.sh` - NEW
- `lib/claude-integration.sh` - NEW

### Modified Libraries
- `utils/code-ownership/ownership-analyser.sh` - Enhanced Claude integration
- `lib/markdown.sh` - New report sections
- `lib/metrics.sh` - May need Gini calculation helpers

### RAG Documents (Existing, Reference Only)
- `rag/code-ownership/bus-factor-analysis.md`
- `rag/code-ownership/ownership-metrics.md`
- `rag/code-ownership/codeowners-best-practices.md`

### Test Infrastructure
- `tests/test-spof-detector.sh` - NEW
- `tests/test-knowledge-analysis.sh` - NEW
- `tests/test-claude-integration.sh` - NEW
- `tests/test-regression.sh` - NEW

## Success Criteria

### Functional Requirements
- ✅ Standard bus factor calculation without Claude
- ✅ 6-criteria SPOF detection
- ✅ Gini coefficient and concentration metrics
- ✅ Claude-enhanced risk identification
- ✅ Succession planning integration
- ✅ Backward compatibility maintained

### Quality Requirements
- ✅ All unit tests passing
- ✅ All regression tests passing
- ✅ Code coverage >80% for new modules
- ✅ Documentation complete
- ✅ CI/CD pipeline green

### Performance Requirements
- Analysis time <2 minutes for medium repos (1000 files)
- Claude API calls <3 per analysis
- Memory usage <500MB for large repos

## Risk Mitigation

### Risk 1: Breaking Changes to Global Libraries
**Impact**: High
**Mitigation**: Comprehensive regression testing, feature flags

### Risk 2: Claude API Rate Limits
**Impact**: Medium
**Mitigation**: Implement caching, graceful degradation to standard mode

### Risk 3: Complex Calculations (Gini) May Be Slow
**Impact**: Low
**Mitigation**: Optimize algorithms, consider sampling for large repos

### Risk 4: RAG Context Too Large for Claude
**Impact**: Medium
**Mitigation**: Summarize RAG patterns, use relevant excerpts only

## Next Steps

1. **Review & Approve** this plan
2. **Set Up Branch**: `feature/code-ownership-risk-analysis`
3. **Start Phase 1**: Begin with `lib/spof-detector.sh`
4. **Iterative Development**: Complete each phase, test, iterate
5. **Final Review**: Code review, documentation review, testing verification

## Related Documents

- [Code Ownership Improvements](./code-ownership-improvements.md) - Original improvement roadmap
- [Bus Factor Analysis RAG](../rag/code-ownership/bus-factor-analysis.md) - Algorithm reference
- [Ownership Metrics RAG](../rag/code-ownership/ownership-metrics.md) - Metrics reference

---

**Ready to implement? Let's start with Phase 1: Standard Bus Factor Analyzer**
