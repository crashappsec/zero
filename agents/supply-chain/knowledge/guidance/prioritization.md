<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
SPDX-License-Identifier: GPL-3.0
-->

# CISA Known Exploited Vulnerabilities (KEV) Prioritization Guide

## Overview

The CISA Known Exploited Vulnerabilities (KEV) catalog is the authoritative source
for vulnerabilities that are actively being exploited in the wild. KEV entries
represent the highest priority for remediation because they represent real,
demonstrated attack capability.

## KEV Catalog Significance

### Why KEV Matters
- **Confirmed Exploitation:** Every KEV entry has verified, active exploitation
- **Federal Mandate:** FCEB agencies must remediate within defined timelines
- **Industry Standard:** Increasingly adopted as private sector baseline
- **Risk Reduction:** Addresses known attack paths, not theoretical risks

### KEV vs CVSS Priority
| Scenario | CVSS Score | KEV Status | Priority |
|----------|------------|------------|----------|
| Log4Shell | 10.0 | Yes | P0 - Immediate |
| Theoretical RCE | 9.8 | No | P1 - Urgent |
| Auth bypass | 8.5 | Yes | P0 - Immediate |
| XSS | 6.1 | No | P3 - Planned |

**Rule:** KEV status always elevates priority to P0, regardless of CVSS score.

## Remediation Timelines

### CISA Binding Operational Directive (BOD) 22-01 Timelines
- **Internet-facing systems:** 2 weeks from KEV addition
- **All other systems:** 3 weeks from KEV addition

### Recommended Private Sector Timelines
- **Critical infrastructure:** Match federal timelines
- **High-value targets:** 7 days for internet-facing
- **Standard systems:** 14 days for internet-facing
- **Internal systems:** 30 days maximum

## KEV Triage Process

### Step 1: Identify KEV Matches
```bash
# Check current KEV catalog
curl -s https://www.cisa.gov/sites/default/files/feeds/known_exploited_vulnerabilities.json | \
  jq '.vulnerabilities[] | {cve: .cveID, name: .vulnerabilityName, due: .dueDate}'
```

### Step 2: Assess Exposure
- Is the vulnerable component internet-facing?
- Is it in production or development only?
- What data/systems can be accessed through it?

### Step 3: Determine Remediation Path
1. **Patch available:** Apply patch immediately
2. **No patch:** Implement compensating controls
3. **Cannot patch:** Document risk acceptance at executive level

## Compensating Controls

When immediate patching is not possible:

### Network-Level Controls
- Web Application Firewall (WAF) rules
- Network segmentation
- Disable vulnerable service temporarily

### Detection Controls
- Enhanced logging on vulnerable systems
- IDS/IPS signatures for exploitation attempts
- SIEM alerts for suspicious patterns

### Access Controls
- Restrict network access to vulnerable systems
- Implement additional authentication
- Limit user privileges

## KEV Reporting Requirements

### Internal Reporting
- Report KEV findings to security leadership within 4 hours
- Document remediation plan within 24 hours
- Track remediation to completion

### External Reporting (if applicable)
- Federal contractors: Report per contract requirements
- Critical infrastructure: Coordinate with sector ISAC
- Regulated industries: Follow sector-specific requirements

## Integration with Vulnerability Management

### Priority Matrix
| KEV | CVSS Critical | CVSS High | CVSS Medium | CVSS Low |
|-----|---------------|-----------|-------------|----------|
| Yes | P0 (24h) | P0 (24h) | P0 (48h) | P0 (72h) |
| No | P1 (7d) | P2 (14d) | P3 (30d) | P4 (90d) |

### Escalation Path
1. KEV discovered → Immediate ticket creation
2. 4 hours: Security team assessment complete
3. 24 hours: Remediation plan documented
4. 48 hours: Production fix deployed or compensating control active
5. 72 hours: All environments remediated

## Resources

- CISA KEV Catalog: https://www.cisa.gov/known-exploited-vulnerabilities-catalog
- KEV JSON Feed: https://www.cisa.gov/sites/default/files/feeds/known_exploited_vulnerabilities.json
- BOD 22-01 Guidance: https://www.cisa.gov/news-events/directives/bod-22-01
