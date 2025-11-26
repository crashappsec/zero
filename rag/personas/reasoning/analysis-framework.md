# Chain of Reasoning Analysis Framework

This document defines the three-phase reasoning process that Claude should follow when generating persona-specific reports from any scanner output.

## Phase 1: Understand Your Audience

Before analyzing any data, internalize who you are advising:

```
<phase_1_audience_understanding>
Read the persona definition carefully. Understand:

1. WHO is this person?
   - Their job title and daily responsibilities
   - Who they report to
   - What decisions they make

2. WHAT do they care about?
   - High priority items (MUST include)
   - Medium priority items (include when relevant)
   - Low priority items (minimize or omit)

3. HOW do they communicate?
   - Terminology they use
   - Level of technical detail they expect
   - Format preferences (tables, charts, prose)

4. WHY do they need this report?
   - Decisions they'll make based on it
   - Actions they'll take
   - People they'll share it with

Take a moment to truly embody this persona's perspective before proceeding.
</phase_1_audience_understanding>
```

## Phase 2: Apply Domain Knowledge

With your audience clearly in mind, now apply the relevant domain expertise:

```
<phase_2_domain_knowledge>
Consider the domain-specific knowledge provided:

1. STANDARDS and BEST PRACTICES
   - What are the relevant industry standards?
   - What does "good" look like in this domain?
   - What frameworks or specifications apply?

2. RISK CONTEXT
   - What are the real-world implications of findings?
   - How do issues compound or interact?
   - What's the severity hierarchy in this domain?

3. REMEDIATION PATTERNS
   - What are common fixes for this type of issue?
   - What are the tradeoffs of different approaches?
   - What's the typical effort level?

4. INDUSTRY BENCHMARKS
   - How do these findings compare to industry norms?
   - What would be considered good/bad/acceptable?

Apply this knowledge through the lens of your audience.
A Security Engineer needs technical depth.
An Engineering Leader needs strategic context.
</phase_2_domain_knowledge>
```

## Phase 3: Generate Persona-Appropriate Output

Now transform the raw scan data into a report that serves your audience:

```
<phase_3_output_generation>
Transform the data according to your audience's needs:

1. FILTER - What to include vs. exclude
   - Include everything marked "high priority" for this persona
   - Include "medium priority" items when they're relevant
   - EXCLUDE or minimize "low priority" items
   - Don't include things just because they're in the data

2. FORMAT - How to present information
   - Use the output format specified for this persona
   - Use their preferred terminology
   - Use appropriate visualizations (tables, charts, prose)
   - Match the formality level they expect

3. FOCUS - What to emphasize
   - Lead with what matters most to this audience
   - Provide depth on their high-priority areas
   - Summarize or aggregate their low-priority areas
   - Connect findings to their decision context

4. FRAME - How to contextualize
   - Use language appropriate to their role
   - Connect to their success metrics
   - Anticipate their follow-up questions
   - Enable the actions they need to take
</phase_3_output_generation>
```

## Example: Same Data, Different Outputs

Given the same vulnerability scan showing 50 CVEs:

### Security Engineer Output
```
## Vulnerability Triage Report

### CISA KEV Findings (P0 - Immediate)
| CVE | Package | Version | CVSS | KEV Listed | Remediation |
|-----|---------|---------|------|------------|-------------|
| CVE-2024-1234 | lodash | 4.17.15 | 9.8 | Yes | Upgrade to 4.17.21 |

### Critical Vulnerabilities (P1 - 24h)
[Detailed table with CVEs, CVSS scores, CWEs, fix versions]

### Remediation Commands
```bash
npm install lodash@4.17.21 express@4.18.2
```
```

### Software Engineer Output
```
## Dependency Update Report

### Quick Wins (Safe to Update Now)
```bash
npm install lodash@4.17.21 axios@1.6.0
```
| Package | Current | Target | Breaking? | Effort |
|---------|---------|--------|-----------|--------|
| lodash | 4.17.15 | 4.17.21 | No | 5 min |

### Major Updates (Needs Review)
- express 4.x → 5.x: Breaking changes in middleware
```

### Engineering Leader Output
```
## Security Health Dashboard

**Overall Score: 72/100** (↑ 5 from last month)

### Risk Distribution
Critical: ██░░░░░░░░ 2
High:     ████░░░░░░ 8
Medium:   ██████████ 40

**Key Insight:** 2 critical issues require immediate attention.
Estimated remediation: 3 engineer-days.

### Recommendations
1. Allocate 1 engineer this week for critical fixes
2. Schedule dependency update sprint for Q1
```

### Auditor Output
```
## Compliance Assessment Report

**Assessment ID:** VUL-2024-001
**Framework:** NIST CSF / SOC 2

### Finding F-001: Critical Vulnerabilities Detected
**Severity:** Critical
**Control:** ID.RA-1

**Criteria:** Per vulnerability management policy, critical
vulnerabilities must be remediated within 24 hours.

**Condition:** 2 critical vulnerabilities identified exceeding SLA.

**Evidence:** Scan report VUL-2024-001-A (attached)
```

## Implementation Notes

1. **Always start with Phase 1** - Understanding the audience shapes everything else
2. **Phase 2 is domain-specific** - Load appropriate RAG content per scanner
3. **Phase 3 is where transformation happens** - Same data, different presentation
4. **Maintain persona consistency** - A security engineer report should ALWAYS look like a security engineer report, regardless of the domain
