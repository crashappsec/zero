<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
SPDX-License-Identifier: GPL-3.0
-->

# Risk Communication for Engineering Leaders

## Executive Risk Summary Framework

### The 3x3 Risk Matrix

Present supply chain risks in a format executives understand:

```
              IMPACT
              Low    Medium   High
         ┌────────┬────────┬────────┐
    High │ Medium │  High  │Critical│
         ├────────┼────────┼────────┤
LIKELIHOOD Medium │  Low   │ Medium │  High  │
         ├────────┼────────┼────────┤
    Low  │  Info  │  Low   │ Medium │
         └────────┴────────┴────────┘
```

### Translating Technical to Business Risk

| Technical Finding | Business Translation |
|------------------|---------------------|
| Critical CVE in auth library | Customer data breach risk |
| Abandoned dependency | Future maintenance cost increase |
| License violation | Legal liability exposure |
| Supply chain attack vector | Brand reputation damage |
| Outdated framework | Compliance audit failure |

### Risk Quantification Model

```
Business Risk Score = Likelihood × Impact × Exposure

Where:
- Likelihood: 1-5 (rare to almost certain)
- Impact: 1-5 (negligible to catastrophic)
- Exposure: 1-3 (internal only to public-facing)

Score Interpretation:
1-10:  Low - Monitor, address in normal sprint
11-25: Medium - Plan remediation within quarter
26-50: High - Priority for next sprint
51-75: Critical - Immediate action required
```

## Communicating with Different Stakeholders

### To the C-Suite

**Format:** Single-page summary with key metrics

```
┌─────────────────────────────────────────────────────────┐
│           SUPPLY CHAIN SECURITY STATUS                  │
│                  November 2024                          │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  Overall Health: 🟢 GREEN (improved from 🟡 YELLOW)     │
│                                                         │
│  Key Metrics:                                           │
│  • Critical vulnerabilities: 0 (target: 0) ✓           │
│  • Mean time to remediate: 4 days (target: 7) ✓        │
│  • Compliance coverage: 94% (target: 95%) ⚠            │
│                                                         │
│  Top Risk: Legacy auth library EOL in Q1                │
│  Action: Migration project approved, 60% complete       │
│                                                         │
│  Investment Ask: None this quarter                      │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

**Language Tips:**
- Lead with business impact, not technical details
- Use traffic light status (green/yellow/red)
- Compare to targets and benchmarks
- Be specific about asks

### To the Board

**Format:** Quarterly risk report with trends

```
Supply Chain Risk Trend (4 Quarters)
─────────────────────────────────────
Risk Score: 45 → 38 → 29 → 22 (↓51%)

Investments Made:
• Automated vulnerability scanning ($X)
• Dependency update automation ($Y)
• Security training program ($Z)

Return on Investment:
• 70% reduction in manual remediation time
• Zero security incidents from dependencies
• Compliance audit passed with no findings
```

### To Product Management

**Format:** Feature impact and timeline clarity

```
Impact Assessment: CVE-2024-XXXXX

Affected Products:
• Product A: Critical path, blocks release
• Product B: Not affected
• Product C: Medium impact, workaround available

Timeline Options:
1. Immediate patch: 2 days, delays Feature X by 1 week
2. Sprint inclusion: 2 weeks, no feature delay
3. Accept risk: Document and monitor (not recommended)

Recommendation: Option 1 - Customer data at risk
```

### To Engineering Teams

**Format:** Clear priorities with technical context

```
This Sprint's Security Priorities
─────────────────────────────────

P0 - Must complete:
• Upgrade lodash to 4.17.21 (prototype pollution)
  Files affected: 3, Estimated: 2 hours

P1 - Should complete:
• Update axios to 1.6.0 (SSRF fix)
  Files affected: 12, Estimated: 4 hours

P2 - Stretch goal:
• Replace moment.js with date-fns
  Files affected: 28, Estimated: 2 days
```

## Building the Business Case

### Cost of Inaction Model

```
Potential Incident Cost Calculation:

Direct Costs:
• Breach notification: $150 per affected customer
• Investigation: $50,000-$500,000
• Legal/regulatory: $100,000-$10,000,000
• Remediation: $200,000-$2,000,000

Indirect Costs:
• Customer churn: 3-7% post-breach
• Brand damage: 6-12 months recovery
• Stock impact: 5-15% (if public)
• Increased insurance: 20-50%

Example: 100,000 customers affected
Direct costs: $15M + $500K + $1M + $500K = $17M
Indirect costs: $5M annual revenue loss

Total risk exposure: $22M
```

### ROI of Prevention

```
Investment: Automated dependency management
Annual cost: $50,000

Benefits:
• Reduced MTTR: 10 hours/vuln → 2 hours/vuln
• Vulnerabilities per year: 50
• Hours saved: 400 hours × $150/hour = $60,000
• Avoided incidents: 1-2 per year = $100,000-$500,000

ROI = (Benefits - Cost) / Cost
ROI = ($160,000 - $50,000) / $50,000 = 220%
```

### Comparison to Industry

```
Your Organization vs Industry Benchmark
──────────────────────────────────────────
                    You    Industry  Target
MTTR (Critical)     4 days  7 days    2 days
Patch coverage      92%     78%       95%
Automation rate     45%     30%       60%
SLA compliance      96%     82%       99%

Status: Ahead of industry, room for improvement
```

## Risk Appetite and Thresholds

### Defining Risk Tolerance

```
Risk Category      Tolerance    Action Threshold
──────────────────────────────────────────────────
Critical vulns     Zero         Immediate response
High vulns         <5 open      48-hour SLA
Medium vulns       <20 open     30-day SLA
Low vulns          <50 open     90-day SLA
Abandoned deps     <10%         Quarterly review
License issues     Zero GPL     Pre-approval required
```

### Escalation Matrix

```
Severity    Team Lead    Director    VP    CISO    CEO
─────────────────────────────────────────────────────────
Low         Notified     -           -     -       -
Medium      Owns         Notified    -     -       -
High        Owns         Decides     Notified  -   -
Critical    Assists      Owns        Decides   Notified  -
KEV/Active  Assists      Assists     Owns      Decides   Notified
```

## Reporting Cadence

### Weekly Report (Team Level)

```
Supply Chain Status - Week 47
─────────────────────────────
New vulnerabilities: 3
Resolved: 7
Open critical: 0
Open high: 2

Actions this week:
• Patched axios (CVE-2024-xxx)
• Migrated 3 services to new auth library

Next week priorities:
• Complete auth library migration (2 remaining)
• Address high-priority lodash update
```

### Monthly Report (Director Level)

```
Monthly Supply Chain Summary - November 2024
─────────────────────────────────────────────

Portfolio Health:
• Repositories scanned: 145/150 (97%)
• Dependencies tracked: 4,823
• Average age of dependencies: 8 months

Vulnerability Metrics:
• New CVEs affecting us: 12
• CVEs remediated: 15
• MTTR: 4.2 days (↓ from 5.1)
• SLA compliance: 96%

Risk Posture:
• Open critical: 0
• Open high: 3 (2 in progress, 1 planned)
• Exceptions granted: 1 (expires 12/15)

Resource Utilization:
• Security engineering hours: 120
• Developer remediation hours: 80
• Automated fixes: 45%

Trends: Improving (3rd consecutive month)
```

### Quarterly Report (Executive Level)

```
Q4 2024 Supply Chain Security Report
─────────────────────────────────────

Executive Summary:
Supply chain security posture improved 15% quarter-over-quarter.
Zero security incidents from dependencies. Compliance audit passed.

Key Achievements:
✓ Implemented automated scanning across all repositories
✓ Reduced MTTR from 7 days to 4 days
✓ Eliminated all critical vulnerabilities

Challenges:
• Technical debt in legacy services delaying some updates
• Vendor slow to patch critical library

Strategic Initiatives:
• SBOM generation for all releases (85% complete)
• SLSA Level 2 compliance (60% complete)

Budget Status:
• Allocated: $200,000
• Spent: $180,000
• Forecast: On budget

Next Quarter Priorities:
1. Complete SBOM rollout
2. Achieve SLSA Level 2
3. Implement vendor security reviews
```

## Handling Difficult Conversations

### When Asked to Accept Risk

```
Response Framework:

1. Acknowledge the business pressure
   "I understand we need to ship by [date]..."

2. Clarify the risk clearly
   "This vulnerability allows [specific attack] which could result in [business impact]..."

3. Present options with trade-offs
   "We have three options:
    A) Delay 2 days, full fix
    B) Ship with compensating control, fix next sprint
    C) Accept risk with documented exception"

4. Recommend and own the decision
   "I recommend option B because..."

5. Document everything
   "I'll send a summary email with the decision and who approved it."
```

### When a Major Vulnerability Drops

```
Communication Timeline:

T+0h:  Alert security and engineering leadership
T+1h:  Initial impact assessment complete
T+2h:  Executive briefing (if critical)
T+4h:  Remediation plan defined
T+8h:  Status update to stakeholders
T+24h: Either resolved or exception documented
```

### When Budget is Challenged

```
Defense Framework:

Cost of Tools: $100,000/year
vs
Cost of One Incident: $500,000-$10,000,000

Probability of incident without tools: 20%/year
Expected loss: $1,000,000-$2,000,000/year

Insurance premium reduction with tools: $50,000/year

Net cost: $100,000 - $50,000 = $50,000
Risk reduction: $1,000,000+ per year

ROI: 20x
```

## Quick Reference

### Communication Checklist

- [ ] Lead with business impact
- [ ] Use consistent metrics
- [ ] Compare to baselines/targets
- [ ] Provide clear recommendations
- [ ] Include resource requirements
- [ ] Set expectations for timeline
- [ ] Document decisions and rationale
