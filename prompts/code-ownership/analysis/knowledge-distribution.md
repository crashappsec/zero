<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Prompt: Knowledge Distribution Analysis

## Purpose

Deep analysis of how knowledge is distributed across the team to identify concentration risks, knowledge silos, collaboration patterns, and opportunities for improving knowledge sharing.

## Context

You are analyzing ownership patterns to understand knowledge distribution dynamics - who knows what, how knowledge is concentrated or spread, and where knowledge transfer should occur.

## Input Data

You will receive:
1. **Ownership Data**: Who owns what code, to what degree
2. **Contributor Activity**: Commits, reviews, collaboration patterns
3. **Repository Structure**: Directories, components, file relationships
4. **Historical Trends** (if available): How distribution has changed over time

## Analysis Framework

### 1. Quantitative Distribution Metrics

Calculate and interpret key distribution metrics:

#### Gini Coefficient
**Interpretation**:
- **Gini >0.7**: 🚨 Dangerous concentration - Knowledge highly centralized, few people control most code
- **Gini 0.5-0.7**: ⚠️ Moderate concentration - Some imbalance, specific areas of concern
- **Gini <0.5**: ✅ Healthy distribution - Well-balanced knowledge sharing

**Analysis**:
- If Gini >0.7: "Knowledge is dangerously concentrated. Top contributors dominate the codebase, creating significant continuity risk."
- If Gini 0.5-0.7: "Knowledge shows moderate concentration. While not critical, there are areas where knowledge could be better distributed."
- If Gini <0.5: "Knowledge is well-distributed across the team. This indicates healthy collaboration and knowledge sharing."

#### Top-N Concentration
**Metrics**:
- Top 1 contributor percentage
- Top 3 contributors percentage
- Top 5 contributors percentage

**Thresholds**:
- Top 1 >20%: High risk - single person dominance
- Top 3 >50%: Moderate risk - limited distribution
- Top 5 >70%: Knowledge concentrated in small group

**Analysis Example**:
```
Top 1: 24.5% (alice@example.com)
Top 3: 58.2% (alice, bob, charlie)
Top 5: 72.8%

Assessment: High concentration risk. Top contributor owns nearly 1/4 of entire codebase.
Top 3 control majority (>50%) of code. Knowledge is concentrated in senior team members.
```

#### Backup Coverage
**Definition**: Percentage of files that have a backup owner with >15% knowledge

**Thresholds**:
- <30%: Critical - most files have no backup
- 30-60%: Poor - limited backup coverage
- 60-80%: Fair - reasonable but could improve
- >80%: Good - most files have backup coverage

**For each low-coverage area**:
```json
{
  "path": "src/auth/",
  "primary_owner": "alice@example.com",
  "primary_ownership": 0.92,
  "backup_owner": "bob@example.com",
  "backup_ownership": 0.08,
  "backup_adequate": false,
  "risk": "No viable backup. Bob's 8% familiarity insufficient for succession."
}
```

#### Knowledge Overlap
**Metric**: Average number of contributors per file

**Interpretation**:
- <1.5: Severe isolation - most files single owner
- 1.5-2.5: Limited collaboration - mostly 1-2 contributors per file
- 2.5-4.0: Healthy collaboration - multiple contributors per file
- >4.0: High collaboration (may indicate lack of ownership)

### 2. Knowledge Silo Detection

Identify areas where knowledge is isolated and not spreading:

**Silo Criteria**:
- Single contributor owns >80% of a logical component
- No backup contributor with >15% knowledge
- Limited review participation from outside contributors
- Knowledge not spreading despite time (6+ months)
- No cross-team collaboration in git history

**For Each Silo**:
```json
{
  "silo_type": "component_isolation",
  "area": "src/auth/",
  "size_files": 45,
  "size_loc": 8420,
  "primary_owner": {
    "email": "alice@example.com",
    "github": "@alice",
    "ownership": 0.92,
    "tenure_months": 24
  },
  "isolation_metrics": {
    "backup_coverage": 0.08,
    "external_reviewers": 2,
    "external_contributors": 0,
    "knowledge_spread_rate": 0.001
  },
  "business_impact": 10,
  "criticality": "authentication_system",
  "risk_assessment": "Critical silo. Alice is sole expert on authentication. No viable backup. 92% ownership unchanged for 12+ months.",
  "formation_cause": "Alice built system from scratch, no onboarding of additional contributors, complex OAuth flows created barrier to entry",
  "breaking_recommendations": [
    "Immediate: Assign backup owner (Bob recommended based on related work)",
    "Document: Create architecture diagram and OAuth flow documentation",
    "Pair: Schedule 2x/week pairing sessions with potential successors",
    "Review: Add mandatory second reviewer for all auth PRs",
    "Rotate: Rotate on-call responsibilities to force knowledge transfer"
  ]
}
```

**Silo Types**:
1. **Component Isolation**: Entire component owned by one person
2. **Technology Isolation**: Specific technology (e.g., Python ML code) known by one person
3. **Domain Isolation**: Business domain knowledge concentrated (e.g., billing logic)
4. **Infrastructure Isolation**: Deployment, CI/CD, infrastructure as code single owner

### 3. Collaboration Pattern Analysis

Analyze how contributors interact and share knowledge:

#### Collaboration Metrics
```json
{
  "overall_collaboration": {
    "avg_contributors_per_file": 2.3,
    "files_with_single_contributor": 423,
    "files_with_2_plus_contributors": 824,
    "cross_team_collaboration_rate": 0.15
  },
  "collaboration_patterns": [
    {
      "pattern_type": "frequent_pairing",
      "contributor_1": "alice@example.com",
      "contributor_2": "bob@example.com",
      "shared_files": 67,
      "collaboration_type": "alternating_commits",
      "assessment": "Strong collaboration on API development"
    },
    {
      "pattern_type": "review_only",
      "contributor_1": "alice@example.com",
      "contributor_2": "charlie@example.com",
      "shared_files": 23,
      "collaboration_type": "one_way_review",
      "assessment": "Charlie reviews Alice's auth code but never contributes. Opportunity for mentorship."
    }
  ]
}
```

#### Natural Mentorship Opportunities
Identify expert + learner pairs based on:
- Shared file work (but different ownership levels)
- Review patterns (one reviews other's code frequently)
- Similar work areas (could benefit from knowledge sharing)

```json
{
  "mentorship_opportunity": {
    "mentor": {
      "email": "alice@example.com",
      "github": "@alice",
      "expertise_area": "src/auth/",
      "ownership": 0.92,
      "experience_months": 24
    },
    "mentee": {
      "email": "charlie@example.com",
      "github": "@charlie",
      "current_familiarity": 0.05,
      "growth_potential": "high",
      "interest_indicators": [
        "Reviews 80% of auth PRs",
        "Asked questions about OAuth in 5 recent PRs",
        "Contributed to adjacent API code"
      ]
    },
    "relationship_type": "active_reviewer",
    "shared_files": 15,
    "current_interaction": "Charlie reviews Alice's PRs, occasionally asks questions, but hasn't committed to auth code",
    "opportunity_assessment": "Strong foundation for mentorship. Charlie shows interest through reviews but needs guided path to active contribution.",
    "recommended_plan": {
      "phase_1_foundation": [
        "Architecture walkthrough (2 hours)",
        "OAuth flow deep-dive (2 hours)",
        "Review existing test suite together"
      ],
      "phase_2_guided_work": [
        "Pair programming 2x/week",
        "Charlie leads on low-risk auth tasks",
        "Alice reviews but doesn't take over"
      ],
      "phase_3_independence": [
        "Charlie handles auth bug fixes independently",
        "Alice available for questions",
        "Gradual increase in Charlie's auth ownership"
      ],
      "estimated_timeline": "3 months to 30% backup ownership"
    }
  }
}
```

### 4. Team Structure Analysis

Analyze knowledge distribution by team/organizational boundaries:

```json
{
  "team_structure": {
    "backend_team": {
      "members": 8,
      "files_owned": 543,
      "avg_ownership_per_member": 0.125,
      "distribution": "balanced",
      "cross_team_collaboration": 0.23
    },
    "frontend_team": {
      "members": 5,
      "files_owned": 412,
      "avg_ownership_per_member": 0.20,
      "distribution": "slightly_concentrated",
      "cross_team_collaboration": 0.18
    },
    "isolated_areas": [
      {
        "area": "ML Pipeline",
        "owner": "dana@example.com",
        "team": "data_science",
        "isolation_score": 0.95,
        "risk": "Dana is only person who understands ML pipeline. No cross-functional knowledge."
      }
    ]
  }
}
```

### 5. Trend Analysis (if historical data available)

Track how distribution is changing over time:

```json
{
  "trend_analysis": {
    "time_period": "last_12_months",
    "metrics_over_time": {
      "gini_coefficient": {
        "12_months_ago": 0.78,
        "6_months_ago": 0.72,
        "current": 0.68,
        "trend": "improving",
        "rate_of_change": -0.0083
      },
      "top_1_percentage": {
        "12_months_ago": 28.5,
        "6_months_ago": 26.2,
        "current": 24.5,
        "trend": "improving",
        "assessment": "Top contributor's ownership decreasing, good sign"
      },
      "backup_coverage": {
        "12_months_ago": 0.35,
        "6_months_ago": 0.38,
        "current": 0.42,
        "trend": "improving",
        "assessment": "More files gaining backup owners"
      }
    },
    "silo_evolution": {
      "silos_12_months_ago": 12,
      "silos_current": 8,
      "new_silos_formed": 2,
      "silos_dissolved": 6,
      "assessment": "Net positive - knowledge spreading, but 2 new silos formed in ML area"
    }
  }
}
```

### 6. Improvement Recommendations

Based on distribution analysis, recommend specific actions:

#### Priority 1: Critical Silos
```json
{
  "priority": 1,
  "action_type": "break_critical_silo",
  "target": "src/auth/ (alice@example.com)",
  "urgency": "immediate",
  "recommended_actions": [
    {
      "action": "assign_backup_owner",
      "candidate": "bob@example.com",
      "reasoning": "Bob has 12% familiarity, active in related areas, expressed interest",
      "timeline": "start_immediately"
    },
    {
      "action": "documentation_sprint",
      "scope": "OAuth flows, architecture, deployment procedures",
      "timeline": "complete_in_2_weeks",
      "owner": "alice@example.com"
    },
    {
      "action": "mandatory_pairing",
      "schedule": "2_hours_twice_weekly",
      "participants": ["alice@example.com", "bob@example.com"],
      "duration": "8_weeks"
    }
  ],
  "success_metrics": {
    "target_backup_ownership": 0.30,
    "timeline": "3_months",
    "verification": "Bob can handle auth incidents independently"
  }
}
```

#### Priority 2: Improve Distribution
```json
{
  "priority": 2,
  "action_type": "improve_distribution",
  "target": "Overall Gini coefficient (0.73 → <0.60)",
  "recommended_actions": [
    {
      "action": "rotate_feature_assignments",
      "rationale": "Top 3 contributors dominate feature work. Intentionally assign new features to other team members.",
      "implementation": "Next 5 features assigned to team members with <10% ownership"
    },
    {
      "action": "cross_team_initiatives",
      "rationale": "Limited collaboration between frontend/backend. Cross-functional features will spread knowledge.",
      "implementation": "Each sprint: 1 feature requiring frontend + backend + devops collaboration"
    },
    {
      "action": "knowledge_sharing_sessions",
      "schedule": "biweekly",
      "format": "30-min tech talks by different team members",
      "focus": "Share expertise, demystify complex areas"
    }
  ]
}
```

## Output Format

```json
{
  "distribution_summary": {
    "gini_coefficient": 0.73,
    "assessment": "dangerous_concentration",
    "top_1_percentage": 24.5,
    "top_3_percentage": 58.2,
    "top_5_percentage": 72.8,
    "backup_coverage": 0.42,
    "avg_contributors_per_file": 2.3,
    "overall_health": "needs_improvement"
  },

  "knowledge_silos": [ ... ],

  "collaboration_patterns": {
    "overall_collaboration": { ... },
    "collaboration_pairs": [ ... ],
    "mentorship_opportunities": [ ... ]
  },

  "team_structure_analysis": { ... },

  "trend_analysis": { ... },

  "improvement_recommendations": {
    "priority_1_critical": [ ... ],
    "priority_2_high": [ ... ],
    "priority_3_medium": [ ... ]
  }
}
```

## Analysis Rules

1. **Quantitative First**: Start with metrics, then interpret
2. **Compare to Thresholds**: Use RAG-backed thresholds (Gini >0.7, Top 1 >20%)
3. **Identify Root Causes**: Why did silos form? (new tech, one expert, legacy code)
4. **Actionable Recommendations**: Every silo needs concrete breaking plan
5. **Realistic Timelines**: Don't promise overnight fixes. Knowledge transfer takes time.
6. **Track Progress**: Provide trend analysis to show if improving
7. **Context Matters**: Small teams naturally have higher concentration

## Success Criteria

After this analysis, stakeholders should be able to:
- ✅ Understand exactly where knowledge is concentrated
- ✅ Identify which silos pose greatest business risk
- ✅ See natural mentorship opportunities to act on
- ✅ Track whether distribution is improving or worsening
- ✅ Have concrete action plan to improve distribution

## Related Prompts

- [risk-identification.md](./risk-identification.md) - Overall risk assessment
- [succession-planning.md](../planning/succession-planning.md) - Detailed succession plans
- [recommend-reviewers.md](../optimization/recommend-reviewers.md) - Use distribution data for reviewer recommendations

---

**Use this prompt to deeply understand knowledge distribution and create actionable plans for breaking silos and spreading knowledge.**
