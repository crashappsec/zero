<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
SPDX-License-Identifier: GPL-3.0
-->

# Resource Planning for Supply Chain Security

## Team Structure Models

### Centralized Security Model

```
┌─────────────────────────────────────────────────┐
│            Security Engineering Team            │
│                                                 │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐        │
│  │Vuln Mgmt│  │ AppSec  │  │  SAST/  │        │
│  │   Lead  │  │Engineer │  │  DAST   │        │
│  └─────────┘  └─────────┘  └─────────┘        │
│                                                 │
│  Responsibilities:                              │
│  • Vulnerability triage and prioritization     │
│  • Security tooling and automation             │
│  • Policy and standards                        │
│  • Training and enablement                     │
│                                                 │
└─────────────────────────────────────────────────┘
                      │
                      ▼ Tickets/PRs/Guidance
┌─────────────────────────────────────────────────┐
│           Product Engineering Teams              │
│                                                 │
│  Team A    Team B    Team C    Team D          │
│  (fixes)   (fixes)   (fixes)   (fixes)         │
│                                                 │
└─────────────────────────────────────────────────┘

Pros: Consistent standards, deep expertise
Cons: Bottleneck potential, context switching
Best for: < 50 engineers, compliance-focused
```

### Embedded Security Champions

```
┌─────────────────────────────────────────────────┐
│         Central Security Team (Small)           │
│                                                 │
│  ┌──────────────┐  ┌──────────────┐            │
│  │Security Lead │  │ Tooling Eng  │            │
│  └──────────────┘  └──────────────┘            │
│                                                 │
│  Responsibilities:                              │
│  • Strategy and standards                      │
│  • Tool management                             │
│  • Champion training                           │
│  • Escalation support                          │
└─────────────────────────────────────────────────┘
                      │
                      ▼ Training/Support
┌─────────────────────────────────────────────────┐
│           Product Engineering Teams              │
│                                                 │
│  Team A         Team B         Team C          │
│  ┌───────┐      ┌───────┐      ┌───────┐      │
│  │Champ🛡│      │Champ🛡│      │Champ🛡│      │
│  └───────┘      └───────┘      └───────┘      │
│                                                 │
│  Champions spend 20% on security               │
│                                                 │
└─────────────────────────────────────────────────┘

Pros: Scales well, embedded context, faster fixes
Cons: Inconsistent quality, training overhead
Best for: 50-500 engineers, fast-moving orgs
```

### Hybrid Platform Model

```
┌─────────────────────────────────────────────────┐
│          Security Platform Team                 │
│                                                 │
│  ┌───────┐ ┌───────┐ ┌───────┐ ┌───────┐      │
│  │ Lead  │ │ Tools │ │Policy │ │ Data  │      │
│  └───────┘ └───────┘ └───────┘ └───────┘      │
│                                                 │
│  Responsibilities:                              │
│  • Self-service security tooling               │
│  • Automated scanning infrastructure           │
│  • Metrics and dashboards                      │
│  • Policy as code                              │
└─────────────────────────────────────────────────┘
                      │
                      ▼ Platform/APIs/Automation
┌─────────────────────────────────────────────────┐
│           Product Teams + Champions              │
│                                                 │
│  Autonomous remediation via platform tools     │
│  Escalation for complex issues only            │
│                                                 │
└─────────────────────────────────────────────────┘

Pros: Self-service, scalable, measurable
Cons: Platform investment, change management
Best for: > 500 engineers, platform-oriented orgs
```

## Staffing Ratios

### Security-to-Developer Ratios

```
Industry Benchmarks:
• Average: 1:100 (one security engineer per 100 devs)
• Good: 1:50
• Strong: 1:25
• Elite: 1:15

Adjust for:
• Regulatory requirements: +50% staffing
• Handling sensitive data: +25% staffing
• High velocity releases: +25% staffing
• Legacy systems: +30% staffing
```

### Supply Chain Focus Allocation

```
Of total security staffing, allocate for supply chain:

Organization Size    Supply Chain FTE    % of Security
──────────────────────────────────────────────────────
< 50 developers      0.25 FTE            Part of AppSec
50-200 developers    0.5-1 FTE           20%
200-500 developers   1-2 FTE             20-25%
500+ developers      2-5 FTE             15-20%
```

## Budget Planning

### Tool Cost Categories

```
Category                    Annual Cost Range    Notes
─────────────────────────────────────────────────────────────
SCA Scanner (commercial)    $30K-$200K          Per-seat or per-repo
SBOM Management            $20K-$100K          Scale with artifacts
Container Scanning         $20K-$100K          Per-node or per-scan
License Compliance         $10K-$50K           Often bundled
Dependency Automation      $0-$50K             GitHub native or Renovate
Security Intelligence      $20K-$100K          CVE feeds, threat intel
─────────────────────────────────────────────────────────────
Typical Enterprise Total:  $100K-$500K/year
```

### Build vs Buy Analysis

```
                        Build           Buy (Commercial)
──────────────────────────────────────────────────────────
Initial cost            $50K-$200K      $30K-$100K/year
Ongoing maintenance     $30K-$80K/year  Included
Time to value           3-6 months      1-2 weeks
Customization           Unlimited       Limited
Support                 Internal        Vendor SLA
Expertise required      High            Low

Decision Framework:
• Buy if: Core competency is not security tooling
• Build if: Unique requirements, strong platform team
```

### ROI Calculation Template

```
Investment: Security tool/process
Cost: $X per year

Quantified Benefits:
──────────────────────────────────────────────────────────
Benefit                              Value
──────────────────────────────────────────────────────────
Reduced MTTR (hours × hourly rate)   $________
Avoided incidents (probability × cost) $________
Audit efficiency (hours saved × rate) $________
Developer productivity (time saved)   $________
Insurance reduction                   $________
──────────────────────────────────────────────────────────
Total Annual Benefit                  $________

ROI = (Benefits - Cost) / Cost × 100 = _____%
Payback Period = Cost / Monthly Benefit = ____ months
```

## Capacity Planning

### Vulnerability Load Forecasting

```
Historical Analysis Template:

Month    New Vulns    Remediated    Net Change    Backlog
────────────────────────────────────────────────────────────
Jan      45           50            -5            120
Feb      38           42            -4            116
Mar      52           55            -3            113
Apr      48           52            -4            109
May      55           48            +7            116
Jun      42           50            -8            108

Trends:
• Average new per month: 47
• Average resolved per month: 50
• Backlog trend: Decreasing (-2.5/month)
• Seasonal spike: Q1 (post-disclosure cycles)
```

### Sprint Allocation Formula

```
Security Sprint Points = Base + Variable + Buffer

Base (Maintenance):
• Scanning and triage: 5 points
• Tool updates: 3 points
• Reporting: 2 points
Total base: 10 points

Variable (Remediation):
• Expected vulns × avg points per vuln
• Example: 12 vulns × 2 points = 24 points

Buffer (Unknowns):
• 15% of (Base + Variable)
• Example: 34 × 0.15 = 5 points

Total: 10 + 24 + 5 = 39 points/sprint
```

### Incident Capacity Reserve

```
Reserve Allocation:
• Normal operations: 80% planned, 20% reserve
• During incidents: Can surge to 100%

Reserve should cover:
• Critical CVE response (2-3 per quarter)
• Security incidents (1-2 per quarter)
• Audit support (1-2 weeks per quarter)
• Tool outages and troubleshooting
```

## Skill Development

### Security Champion Program

```
Program Structure:
──────────────────────────────────────────────────────────
Phase          Duration    Content
──────────────────────────────────────────────────────────
Foundation     4 weeks     Security basics, OWASP Top 10
Supply Chain   2 weeks     Dependencies, SBOMs, scanning
Advanced       4 weeks     Threat modeling, code review
Ongoing        Weekly      Office hours, new threats
──────────────────────────────────────────────────────────

Time Commitment: 4-8 hours/month after initial training
Recognition: Title, career progression, conference budget

ROI of Champion Program:
• Training cost: $2K per champion
• Time cost: 80 hours × $75/hr = $6K
• Total investment: $8K per champion
• Benefit: 1 champion reduces security team load by 10%
• At 5 champions: 50% load reduction = 0.5 FTE saved
• Savings: ~$75K/year (0.5 FTE)
• ROI: (75K - 40K) / 40K = 87.5%
```

### Training Budget Guidelines

```
Per Engineer Annually:
• Security awareness: $200-$500
• Technical training: $1,000-$3,000
• Conferences: $2,000-$5,000
• Certifications: $500-$2,000

Security Team Members:
• Advanced training: $5,000-$10,000
• Conferences: $5,000-$10,000
• Certifications: $2,000-$5,000
• Tools/lab access: $1,000-$3,000

Recommended Budget: 3-5% of security team salary
```

## Outsourcing Considerations

### What to Keep In-House

```
Always In-House:
• Strategic decisions
• Risk acceptance authority
• Incident response lead
• Vendor management
• Architecture review

Can Outsource:
• Scanning operations
• Initial triage
• Reporting/dashboards
• Tool implementation
• Compliance documentation
```

### Managed Service Evaluation

```
Evaluation Criteria:
─────────────────────────────────────────────────
Criterion           Weight    Vendor A    Vendor B
─────────────────────────────────────────────────
Coverage            20%       ____        ____
Accuracy (FP rate)  20%       ____        ____
Integration ease    15%       ____        ____
Response time       15%       ____        ____
Expertise depth     15%       ____        ____
Cost                15%       ____        ____
─────────────────────────────────────────────────
Total Score                   ____        ____
```

## Resource Request Template

### Headcount Request

```
Position: Supply Chain Security Engineer
Level: Senior
Cost: $XXX,XXX fully loaded

Current State:
• Vulnerabilities per month: 50
• Current remediation capacity: 40
• Backlog growth: +10/month
• Current team: 1 FTE

Business Impact of Gap:
• Backlog will reach 120 in 12 months
• SLA compliance dropping (currently 85%)
• Risk of compliance finding
• Developer productivity impact

With This Hire:
• Remediation capacity: 60/month
• Backlog reduction: -10/month
• SLA compliance: 95%+
• Automation projects enabled

ROI:
• Reduced incident probability: $200K/year avoided cost
• Compliance penalty avoidance: $100K
• Developer productivity: $50K
• Total benefit: $350K
• Cost: $200K
• ROI: 75%
```

### Tool Investment Request

```
Tool: [Name]
Category: Software Composition Analysis
Annual Cost: $XX,XXX

Current Challenge:
• Manual dependency review: 20 hours/week
• Vulnerability detection: 3-5 day delay
• No SBOM capability
• Compliance gap for [regulation]

Proposed Solution Benefits:
• Automated scanning: Save 15 hours/week
• Real-time detection: Reduce to <1 hour
• SBOM generation: Meet compliance
• Integration with CI/CD: Shift left

Quantified Value:
• Labor savings: 15 hrs × $75 × 52 = $58,500
• Faster detection: Reduced breach probability
• Compliance: Avoid $50K+ finding
• Total annual value: $100K+

Payback Period: 6 months
3-Year TCO: $XXX,XXX
3-Year Value: $XXX,XXX
```

## Quick Reference

### Resource Planning Checklist

- [ ] Define team structure model
- [ ] Calculate staffing ratios
- [ ] Allocate supply chain FTE
- [ ] Budget for tools
- [ ] Plan for training
- [ ] Establish champion program
- [ ] Set aside incident reserve
- [ ] Document outsourcing strategy
- [ ] Create growth projections
