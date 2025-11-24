<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Prompt: Code Ownership Risk Identification

## Purpose

Comprehensive risk analysis of code ownership patterns to identify critical vulnerabilities that could impact project continuity, using research-backed methodologies and structured assessment criteria.

## Context

You are analyzing code ownership data from a git repository to identify and prioritize risks using the 6-criteria SPOF (Single Point of Failure) framework, knowledge distribution metrics, and succession planning priorities.

## Input Data Structure

You will receive JSON containing:
1. **Standard Analysis Results**:
   - Bus factor (overall and per-component)
   - SPOF candidates with 6-criteria assessment
   - Knowledge distribution metrics (Gini coefficient, concentration)
   - Contributor activity data
   - Test coverage and documentation metrics

2. **RAG Context** (Research-backed patterns):
   - Bus factor thresholds and calculation methods
   - SPOF risk level definitions
   - Knowledge transfer time estimates
   - Best practices for succession planning

3. **Repository Metadata**:
   - Repository size (files, LOC)
   - Team size
   - Activity level (commits/month)
   - Technology stack

## Analysis Framework

### 1. Bus Factor Risk Assessment

Analyze the bus factor data and identify critical risks:

**Overall Bus Factor Interpretation**:
- **Bus Factor = 1**: 🚨 Critical - Single point of failure for entire project
- **Bus Factor = 2**: ⚠️ High Risk - Vulnerable to 2 departures
- **Bus Factor = 3**: ⚡ Acceptable - Can sustain 3 departures
- **Bus Factor ≥4**: ✅ Healthy - Good knowledge distribution

**Component-Level Analysis**:
For each component with bus factor <3:
- Identify the component path
- List current owners
- Assess business criticality (1-10):
  - 10: Authentication, payments, core transactions
  - 8-9: Revenue-generating, customer-facing
  - 6-7: Important but not critical
  - 4-5: Internal tools, nice-to-have
  - 1-3: Experimental, documentation
- Calculate risk score: `bus_factor_risk = criticality * (1 / bus_factor)`

### 2. SPOF Detection Using 6 Criteria

For each file flagged as potential SPOF, assess against 6 criteria:

**Criteria**:
1. **Single Contributor**: Only one person has ever committed to this file
2. **Critical Path**: File is in critical system (auth, payments, core API, security)
3. **High Complexity**: File has >500 lines of code
4. **No Backup Owner**: No other contributor has >15% knowledge of this file
5. **Low Test Coverage**: File has <60% test coverage
6. **No Documentation**: File lacks documentation (no README, minimal comments)

**Risk Level Classification**:
- **Critical SPOF** (6/6 criteria): Immediate action required within 1 week
- **High SPOF** (4-5 criteria): Action required within 2-4 weeks
- **Medium SPOF** (2-3 criteria): Plan action within quarter
- **Low SPOF** (1 criterion): Monitor and track

**For Each SPOF**:
```json
{
  "path": "src/auth/oauth.ts",
  "risk_level": "critical",
  "criteria_met": 6,
  "criteria_details": {
    "single_contributor": true,
    "critical_path": "authentication",
    "complexity_loc": 847,
    "backup_owner": false,
    "test_coverage": 0.42,
    "documentation": false
  },
  "current_owner": {
    "email": "alice@example.com",
    "github": "@alice",
    "activity_status": "active",
    "last_commit": "2024-11-20"
  },
  "business_impact": 10,
  "impact_rationale": "Authentication is business-critical. Failure would prevent all user logins. No alternative authentication mechanism exists.",
  "succession_priority": 0.95
}
```

### 3. Knowledge Distribution Analysis

Analyze concentration and distribution metrics:

**Gini Coefficient Interpretation**:
- **Gini >0.7**: 🚨 Dangerous concentration - Knowledge highly centralized
- **Gini 0.5-0.7**: ⚠️ Moderate concentration - Some imbalance
- **Gini <0.5**: ✅ Healthy distribution - Well-balanced

**Top-N Concentration Analysis**:
- **Top 1 >20%**: 🚨 High risk - Single person has excessive ownership
- **Top 1 15-20%**: ⚠️ Moderate risk - Monitor carefully
- **Top 1 <15%**: ✅ Healthy
- **Top 3 >50%**: ⚠️ Moderate risk - Limited knowledge distribution
- **Top 3 <50%**: ✅ Healthy

**Knowledge Silos**:
Identify areas where:
- Single contributor owns >80% of a component
- No backup contributor with >15% knowledge
- Limited cross-team collaboration
- Knowledge not spreading despite time

For each silo:
```json
{
  "area": "src/auth/",
  "primary_owner": "alice@example.com",
  "ownership_percentage": 92.5,
  "backup_coverage": 0.08,
  "knowledge_overlap": "minimal",
  "risk_assessment": "High isolation - no viable successor",
  "business_impact": 10
}
```

### 4. Succession Planning Priorities

Calculate succession priority for each SPOF using the formula:
```
priority = criticality * 0.40 +
           concentration_risk * 0.30 +
           departure_risk * 0.20 +
           documentation_gap * 0.10
```

**Departure Risk Scoring**:
- **High (1.0)**: Known departure, extended absence, transfer announced
- **Medium (0.5)**: Declining activity (>50% drop), no commits >90 days
- **Low (0.2)**: Normal activity, regular contributions

**Documentation Gap**:
```
docs_gap = 1 - (docs_quality_score / 10)

Docs Quality (0-10):
- Has README: +2
- Has architecture diagram: +2
- Has API docs: +2
- Has inline comments: +2
- Has runbook: +2
```

**For Each Priority**:
```json
{
  "area": "src/auth/oauth.ts",
  "priority_score": 0.95,
  "current_owner": {
    "email": "alice@example.com",
    "github": "@alice",
    "departure_risk": 0.2
  },
  "recommended_successor": {
    "email": "bob@example.com",
    "github": "@bob",
    "current_familiarity": 0.12,
    "reasoning": "Bob has 12% existing familiarity from related auth work, actively reviews auth PRs, and has expressed interest in learning OAuth"
  },
  "transfer_estimate": {
    "complexity": "high",
    "estimated_weeks": "6-8",
    "rationale": "Complex OAuth flows, limited documentation, requires understanding of multiple external integrations"
  },
  "immediate_actions": [
    "Schedule architecture walkthrough with Alice",
    "Create OAuth flow documentation",
    "Add Bob as required reviewer on auth PRs",
    "Pair programming sessions 2x/week"
  ]
}
```

### 5. Collaboration & Knowledge Sharing Opportunities

Identify mentorship and knowledge sharing opportunities:

**Mentorship Pairing**:
Look for:
- Expert (>60% ownership) + Learner (<20% ownership) on shared files
- Contributors who review each other's code frequently
- Natural collaboration patterns in git history

**Code Review Rotation**:
- Areas with single reviewer → Recommend rotating reviewers
- High-risk files → Require multiple reviewers
- Knowledge silos → Mandate cross-team reviews

**Knowledge Overlap Analysis**:
- Files with <2 contributors → Flag for knowledge sharing
- Components with no cross-training → Recommend pair programming
- Teams working in isolation → Suggest collaboration initiatives

```json
{
  "type": "mentorship",
  "mentor": {
    "email": "alice@example.com",
    "github": "@alice",
    "expertise_area": "src/auth/",
    "ownership": 0.92
  },
  "mentee": {
    "email": "charlie@example.com",
    "github": "@charlie",
    "current_familiarity": 0.05,
    "growth_potential": "high"
  },
  "shared_files": 15,
  "collaboration_pattern": "Charlie reviews Alice's PRs but rarely commits to auth code",
  "recommendation": "Move from passive review to active pairing. Start with 2-hour sessions weekly focusing on OAuth fundamentals."
}
```

## Output Format

Provide structured JSON output:

```json
{
  "analysis_metadata": {
    "version": "4.0.0",
    "analysis_date": "2025-11-24T10:00:00Z",
    "repository": "example-repo",
    "team_size": 15,
    "files_analyzed": 1247
  },

  "risk_summary": {
    "overall_bus_factor": 2,
    "bus_factor_status": "high_risk",
    "critical_risks": 3,
    "high_risks": 8,
    "medium_risks": 15,
    "low_risks": 42,
    "total_spofs": 68
  },

  "critical_risks": [
    {
      "path": "src/auth/oauth.ts",
      "risk_level": "critical",
      "criteria_met": 6,
      "criteria_details": { ... },
      "current_owner": { ... },
      "business_impact": 10,
      "impact_rationale": "...",
      "succession_priority": 0.95
    }
  ],

  "high_risks": [ ... ],
  "medium_risks": [ ... ],

  "knowledge_distribution": {
    "gini_coefficient": 0.73,
    "assessment": "dangerous_concentration",
    "top_1_percentage": 24.5,
    "top_3_percentage": 58.2,
    "backup_coverage": 0.42,
    "silos_identified": [ ... ]
  },

  "succession_priorities": [
    {
      "area": "src/auth/oauth.ts",
      "priority_score": 0.95,
      "current_owner": { ... },
      "recommended_successor": { ... },
      "transfer_estimate": { ... },
      "immediate_actions": [ ... ]
    }
  ],

  "collaboration_opportunities": [
    {
      "type": "mentorship",
      "mentor": { ... },
      "mentee": { ... },
      "shared_files": 15,
      "recommendation": "..."
    },
    {
      "type": "code_review_rotation",
      "area": "src/api/",
      "current_reviewers": [ ... ],
      "recommended_rotation": [ ... ]
    }
  ],

  "component_analysis": {
    "src/auth": {
      "bus_factor": 1,
      "criticality": 10,
      "risk_score": 10.0,
      "primary_owners": [ ... ],
      "recommendations": "Urgent: Establish backup ownership for authentication"
    },
    "src/api": {
      "bus_factor": 3,
      "criticality": 8,
      "risk_score": 2.67,
      "primary_owners": [ ... ],
      "recommendations": "Healthy: Maintain current distribution"
    }
  }
}
```

## Analysis Rules

1. **Evidence-Based**: Every risk must be backed by specific data (metrics, thresholds from RAG)
2. **Prioritize by Impact**: Business impact × risk probability
3. **Use RAG Thresholds**: Apply research-backed thresholds (Gini >0.7, Top 1 >20%, etc.)
4. **Be Specific**: Always include file paths, owner names/emails, GitHub handles
5. **Actionable**: Succession priorities should include concrete next steps
6. **Realistic**: Transfer time estimates based on complexity (1-2 weeks simple, 6-8 weeks complex)
7. **Context-Aware**: Consider team size, repository size when assessing risk

## Success Metrics

The analysis should enable stakeholders to:
- ✅ Identify top 10 most critical risks immediately
- ✅ Understand which components need immediate succession planning
- ✅ Prioritize knowledge transfer activities
- ✅ Track improvement in risk metrics over time
- ✅ Make data-driven decisions about resource allocation

## Related Prompts

- [knowledge-distribution.md](./knowledge-distribution.md) - Deep distribution analysis
- [succession-planning.md](../planning/succession-planning.md) - Detailed succession planning
- [analyze-repository.md](../audit/analyze-repository.md) - General ownership analysis

---

**Use this prompt to generate comprehensive risk assessments that enable proactive risk mitigation and succession planning.**
