<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
SPDX-License-Identifier: GPL-3.0
-->

# Security Metrics and KPIs for Supply Chain Security

## Core Vulnerability Metrics

### Mean Time to Remediate (MTTR)

**Definition:** Average time from vulnerability discovery to verified remediation.

**Calculation:**
```
MTTR = Σ(Remediation Date - Discovery Date) / Number of Vulnerabilities
```

**Targets by Severity:**
| Severity | Target MTTR | Industry Benchmark |
|----------|-------------|-------------------|
| Critical | <24 hours | 7 days |
| High | <7 days | 30 days |
| Medium | <30 days | 60 days |
| Low | <90 days | 180 days |

**Tracking Template:**
```json
{
  "period": "2024-Q4",
  "mttr_by_severity": {
    "critical": {"target_hours": 24, "actual_hours": 18, "status": "met"},
    "high": {"target_days": 7, "actual_days": 5, "status": "met"},
    "medium": {"target_days": 30, "actual_days": 45, "status": "missed"},
    "low": {"target_days": 90, "actual_days": 75, "status": "met"}
  },
  "overall_mttr_days": 12
}
```

### Vulnerability Density

**Definition:** Number of vulnerabilities per unit of code or dependency.

**Calculations:**
```
Vuln per 1K Dependencies = (Total Vulns / Total Dependencies) × 1000
Vuln per 100K LOC = (Total Vulns / Lines of Code) × 100000
```

**Benchmarks:**
| Metric | Good | Acceptable | Needs Work |
|--------|------|------------|------------|
| Vulns per 1K deps | <5 | 5-15 | >15 |
| Critical vulns per repo | 0 | 1-2 | >2 |

### SLA Compliance Rate

**Definition:** Percentage of vulnerabilities remediated within SLA.

**Calculation:**
```
SLA Compliance = (Vulns Fixed Within SLA / Total Vulns) × 100
```

**Target:** >95% overall, 100% for Critical/KEV

## Supply Chain Specific Metrics

### SLSA Level Coverage

**Definition:** Percentage of artifacts meeting each SLSA level.

```
Level 1 Coverage = (Artifacts with Provenance / Total Artifacts) × 100
Level 2 Coverage = (Artifacts with Hosted Build / Total Artifacts) × 100
Level 3 Coverage = (Artifacts with Hardened Build / Total Artifacts) × 100
```

**Target Progression:**
- Year 1: 80% Level 1
- Year 2: 80% Level 2
- Year 3: 50% Level 3

### Dependency Freshness

**Definition:** Percentage of dependencies within N versions of latest.

```
Freshness = (Dependencies within 1 major version / Total) × 100
```

**Categories:**
| Status | Definition | Target % |
|--------|------------|----------|
| Current | Latest version | >30% |
| Recent | Within 1 minor | >50% |
| Outdated | >1 minor behind | <15% |
| Severely Outdated | >1 major behind | <5% |

### Abandoned Package Ratio

**Definition:** Percentage of dependencies from abandoned projects.

```
Abandonment Ratio = (Abandoned Packages / Total Packages) × 100
```

**Abandonment criteria:**
- No commits in 2+ years
- No response to security issues
- Maintainer explicitly abandoned

**Target:** <5% of dependencies abandoned

## Risk Metrics

### Risk Score Aggregation

**Definition:** Weighted risk score across the portfolio.

```
Portfolio Risk Score = Σ(Vuln Severity × Asset Criticality × Exposure)
```

**Weights:**
- Critical vuln: 10
- High vuln: 5
- Medium vuln: 2
- Low vuln: 1

**Asset Criticality:**
- Business critical: 3×
- Production: 2×
- Non-production: 1×

### Attack Surface Metrics

```
External Attack Surface = Internet-facing packages with vulns
Internal Attack Surface = Internal packages with vulns
Supply Chain Attack Surface = Third-party dependencies with vulns
```

## Operational Metrics

### Scanner Coverage

**Definition:** Percentage of repositories/artifacts being scanned.

```
Coverage = (Repos with Active Scanning / Total Repos) × 100
```

**Target:** 100% for production systems

### False Positive Rate

**Definition:** Percentage of reported vulnerabilities that are not actual risks.

```
FP Rate = (False Positives / Total Findings) × 100
```

**Target:** <10%

### Automation Rate

**Definition:** Percentage of vulnerabilities automatically remediated.

```
Automation Rate = (Auto-fixed Vulns / Total Vulns) × 100
```

**Target:** >40% for patch-level updates

## Reporting Dashboards

### Executive Dashboard

```
┌─────────────────────────────────────────────────────────┐
│              Security Posture Summary                   │
├─────────────────────────────────────────────────────────┤
│  Risk Score: 72/100 (↑5 from last month)               │
│                                                         │
│  ┌───────────────┬───────────────┬───────────────┐     │
│  │ Open Vulns    │ MTTR (days)   │ SLA Compliance│     │
│  │     23        │     4.2       │     96%       │     │
│  │  (↓12)        │   (↓1.3)      │    (↑2%)      │     │
│  └───────────────┴───────────────┴───────────────┘     │
│                                                         │
│  Critical Issues: 0  │  KEV Matches: 0  │  P0 Open: 0  │
└─────────────────────────────────────────────────────────┘
```

### Team Dashboard

```
┌─────────────────────────────────────────────────────────┐
│              Vulnerability Queue                         │
├─────────────────────────────────────────────────────────┤
│  Priority  │  Count  │  Oldest  │  Avg Age  │  SLA     │
│  P0        │    0    │    -     │     -     │   OK     │
│  P1        │    3    │   2d     │    1d     │   OK     │
│  P2        │   12    │   8d     │    5d     │   OK     │
│  P3        │   45    │  22d     │   14d     │   OK     │
│  P4        │   89    │  45d     │   30d     │   RISK   │
└─────────────────────────────────────────────────────────┘
```

## Trend Analysis

### Monthly Trend Template

```json
{
  "trends": {
    "total_vulns": [120, 115, 108, 95, 88],
    "critical_vulns": [5, 3, 2, 1, 0],
    "mttr_days": [8, 7.5, 6, 5.5, 4.2],
    "sla_compliance": [88, 90, 92, 94, 96]
  },
  "analysis": {
    "direction": "improving",
    "rate_of_change": "-15% vulns/month",
    "projection": "Zero critical by Q2"
  }
}
```

### Quarterly Business Review Metrics

1. **Risk Reduction:** % decrease in portfolio risk score
2. **Investment Efficiency:** Cost per vulnerability fixed
3. **Process Maturity:** Automation rate improvement
4. **Compliance Status:** Regulatory requirement coverage
5. **Benchmark Comparison:** Performance vs industry

## Alerting Thresholds

| Metric | Warning | Critical |
|--------|---------|----------|
| Open Critical Vulns | >0 | >2 |
| P0 Age (hours) | >12 | >24 |
| SLA Compliance | <95% | <90% |
| MTTR Critical | >24h | >48h |
| KEV Matches | >0 | - |
| Abandoned Deps | >10% | >20% |

## References

- OWASP SAMM: https://owaspsamm.org/
- BSIMM: https://www.bsimm.com/
- OpenSSF Scorecard: https://securityscorecards.dev/
