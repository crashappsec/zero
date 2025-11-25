<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Competitive Analysis: Developer Productivity Intelligence Tools

**Date**: 2025-11-25
**Purpose**: Understand the DPI market landscape to identify opportunities for Gibson Powers
**Status**: Strategic Planning

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Market Overview](#market-overview)
3. [The Five Core DPI Capabilities](#the-five-core-dpi-capabilities)
4. [Feature Matrix by Core Capability](#feature-matrix-by-core-capability)
5. [Feature Matrix: Security & Supply Chain](#feature-matrix-security--supply-chain-capabilities)
6. [Detailed Competitor Analysis](#detailed-competitor-analysis)
7. [Internal Developer Portal Comparison](#internal-developer-portal-comparison)
8. [Gibson Powers: Strategic Positioning](#gibson-powers-strategic-positioning)
9. [Gap Analysis](#gap-analysis-market-gaps-gibson-powers-addresses)
10. [Conclusion](#conclusion)

---

## Executive Summary

This document analyzes the Developer Productivity Intelligence (DPI) market to understand where **Gibson Powers** fits within the broader software analysis landscape.

### What Gibson Powers Is

**Gibson Powers is NOT a DPI tool.** It is:

- An **open-source software analysis toolkit** (GPL-3.0)
- The **free, open-source component** of the Crash Override platform
- A set of **analyzers** for understanding software: technology detection, supply chain analysis, code ownership, certificate analysis, and more
- An **AI-enhanced** analysis platform with RAG-powered insights
- An **on-ramp** to the commercial Crash Override platform for teams needing enterprise features

### Why This Analysis Matters

Understanding the DPI market reveals:
1. **Gaps** that Gibson Powers uniquely addresses (security, supply chain, technology detection)
2. **Features** that the commercial Crash Override platform could offer
3. **Market positioning** for Gibson Powers as complementary to (not competing with) DPI tools

### Key Insight

**DPI tools focus on productivity metrics. Gibson Powers focuses on software analysis.**

They answer different questions:
- **DPI Tools**: "How productive is my team?" / "Where does engineering time go?"
- **Gibson Powers**: "What is this software made of?" / "Is it secure?" / "What technologies does it use?"

---

## Market Overview

### What is a DPI/SEI Platform?

Developer Productivity Intelligence (DPI) or Software Engineering Intelligence (SEI) platforms help engineering leaders with five core capabilities:

1. **Measure** engineering productivity and team health
2. **Align** engineering work with business objectives
3. **Identify** bottlenecks and areas for improvement
4. **Track** DORA metrics and other delivery indicators
5. **Understand** where engineering investment goes

### Gartner Market Guide (2024)

Gartner released its first [**Market Guide for Software Engineering Intelligence Platforms**](https://www.gartner.com/en/documents/5276563) in March 2024, recognizing SEI as an emerging category with significant growth potential.

**Key Findings**:
- By 2027, SEI platform adoption expected to rise to **50%** (up from 5% in 2024)
- Client interactions on SEI doubled from 2022 to 2023
- Market is small but growing rapidly
- Existing DevOps and agile tools evolving to include SEI features

**Gartner Definition**: SEI platforms provide software engineering leaders with data-driven visibility into the engineering team's use of time and resources, operational effectiveness, and progress on deliverables.

**Representative Vendors**: DX, Jellyfish, LinearB, Swarmia (among others)

### Market Leaders

| Tool | Focus | Pricing Model | Notable Customers |
|------|-------|---------------|-------------------|
| **DX** | Research-based developer experience | Enterprise | Pfizer, eBay (acquired by Atlassian 2025) |
| **Jellyfish** | Business alignment & resource allocation | Enterprise | Clari, Hootsuite, Priceline, PagerDuty (500+ orgs) |
| **Swarmia** | Engineering effectiveness | $39/eng/mo | Miro, Docker, Webflow |
| **LinearB** | Workflow automation & DORA | Free tier available | 3,000+ engineering leaders |

---

## The Five Core DPI Capabilities

All DPI/SEI platforms are evaluated against these five core capabilities:

### 1. MEASURE - Engineering Productivity & Team Health

**What It Includes**:
- Developer experience surveys and feedback
- SPACE framework metrics (Satisfaction, Performance, Activity, Communication, Efficiency)
- Team health indicators
- Work patterns analysis
- Collaboration metrics
- AI tool adoption tracking (Copilot, Cursor, etc.)

**Why It Matters**: Understanding how productive and healthy your teams are is foundational to improvement.

### 2. ALIGN - Engineering Work with Business Objectives

**What It Includes**:
- OKR alignment and tracking
- Initiative/project-level monitoring
- Strategic vs tactical work categorization
- Business outcome correlation
- Quarterly planning support
- Roadmap alignment

**Why It Matters**: Engineering leaders must demonstrate that work aligns with business strategy.

### 3. IDENTIFY - Bottlenecks & Areas for Improvement

**What It Includes**:
- Lifecycle bottleneck detection
- Code review delays
- CI/CD performance issues
- Merge queue analysis
- Work-in-progress limits
- Process friction identification
- Working agreements monitoring

**Why It Matters**: Finding and eliminating friction accelerates delivery.

### 4. TRACK - DORA Metrics & Delivery Indicators

**What It Includes**:
- **Deployment Frequency** - How often code deploys to production
- **Lead Time for Changes** - Time from commit to production
- **Mean Time to Recovery (MTTR)** - Time to restore service after incident
- **Change Failure Rate** - Percentage of deployments causing failures
- Additional metrics: Cycle time, PR size, merge frequency

**Why It Matters**: DORA metrics are the industry standard for measuring software delivery performance.

### 5. UNDERSTAND - Where Engineering Investment Goes

**What It Includes**:
- Resource allocation tracking
- Investment categorization (features, bugs, tech debt, security)
- Software capitalization (CapEx) reporting
- Cost per project/initiative
- FTE-based effort modeling
- Time distribution analysis

**Why It Matters**: Engineering is expensive - leaders need to justify and optimize investment.

---

## Feature Matrix by Core Capability

### 1. MEASURE - Engineering Productivity & Team Health

| Feature | DX | Jellyfish | Swarmia | LinearB | Backstage |
|---------|----|-----------|---------|---------|-----------|
| **Developer Surveys** | ✅ Strong | ❌ | ✅ 32q | ❌ | ❌ |
| **SPACE Framework** | ✅ Core | ⚠️ Partial | ✅ | ⚠️ | ❌ |
| **Team Health Metrics** | ✅ | ✅ | ✅ | ⚠️ | ❌ |
| **Real-time Feedback** | ✅ Unique | ❌ | ⚠️ | ❌ | ❌ |
| **AI Tool Adoption** | ✅ | ❌ | ✅ | ❌ | ❌ |
| **Work Patterns** | ✅ | ✅ | ✅ | ⚠️ | ❌ |
| **Benchmarking** | ✅ Direct™ | ✅ | ⚠️ | ⚠️ | ❌ |

**Leader**: **DX** - Founded by DORA/SPACE researchers, strongest measurement framework

---

### 2. ALIGN - Engineering Work with Business Objectives

| Feature | DX | Jellyfish | Swarmia | LinearB | Backstage |
|---------|----|-----------|---------|---------|-----------|
| **OKR Alignment** | ❌ | ✅ Strong | ⚠️ | ❌ | ❌ |
| **Initiative Tracking** | ⚠️ | ✅ Strong | ✅ | ⚠️ | ❌ |
| **Quarterly Planning** | ❌ | ✅ Capacity | ✅ | ❌ | ❌ |
| **Roadmap Visibility** | ⚠️ | ✅ | ✅ | ⚠️ | ❌ |
| **Business Outcome Tracking** | ⚠️ | ✅ Strong | ✅ | ⚠️ | ❌ |
| **Executive Dashboards** | ✅ | ✅ | ✅ | ⚠️ | ❌ |

**Leader**: **Jellyfish** - Strongest business alignment with patented resource allocation

---

### 3. IDENTIFY - Bottlenecks & Areas for Improvement

| Feature | DX | Jellyfish | Swarmia | LinearB | Backstage |
|---------|----|-----------|---------|---------|-----------|
| **Lifecycle Bottlenecks** | ✅ | ✅ Explorer | ✅ | ✅ | ❌ |
| **Code Review Delays** | ✅ | ✅ | ✅ Auto | ✅ Auto | ❌ |
| **CI/CD Performance** | ⚠️ | ✅ | ✅ Strong | ⚠️ | Plugin |
| **PR Analytics** | ✅ | ✅ | ✅ Timeline | ✅ | ❌ |
| **Working Agreements** | ⚠️ | ❌ | ✅ Strong | ⚠️ | ❌ |
| **Workflow Automation** | ⚠️ | ⚠️ | ✅ Slack | ✅ gitStream | ❌ |

**Leader**: **Swarmia** - Best combination of bottleneck detection + automation

---

### 4. TRACK - DORA Metrics & Delivery Indicators

| Feature | DX | Jellyfish | Swarmia | LinearB | Backstage |
|---------|----|-----------|---------|---------|-----------|
| **Deployment Frequency** | ✅ | ✅ | ✅ | ✅ Free | Plugin |
| **Lead Time for Changes** | ✅ | ✅ | ✅ | ✅ Free | Plugin |
| **MTTR** | ✅ | ✅ | ✅ | ✅ Free | ❌ |
| **Change Failure Rate** | ✅ | ✅ | ✅ | ✅ Free | ❌ |
| **Cycle Time** | ✅ | ✅ | ✅ | ✅ | Plugin |
| **Custom Metrics** | ✅ | ⚠️ | ⚠️ | ⚠️ | Plugin |
| **Real-time Dashboards** | ✅ | ✅ | ✅ | ✅ | Plugin |

**Leader**: **LinearB** - Only vendor offering FREE DORA metrics

---

### 5. UNDERSTAND - Where Engineering Investment Goes

| Feature | DX | Jellyfish | Swarmia | LinearB | Backstage |
|---------|----|-----------|---------|---------|-----------|
| **Resource Allocation** | ❌ | ✅ Patented | ✅ | ❌ | ❌ |
| **Investment Categories** | ⚠️ | ✅ Strong | ✅ AI-powered | ⚠️ | ❌ |
| **Software Capitalization** | ❌ | ✅ Audit-ready | ✅ | ❌ | ❌ |
| **FTE Effort Model** | ❌ | ✅ | ✅ New | ❌ | ❌ |
| **Cost Per Project** | ❌ | ✅ | ✅ | ❌ | ❌ |
| **Time Distribution** | ✅ | ✅ | ✅ | ⚠️ | ❌ |
| **Code Ownership** | ⚠️ | ✅ | ⚠️ | ⚠️ | Plugin |

**Leader**: **Jellyfish** - Patented resource allocation, audit-ready CapEx reporting

---

## Feature Matrix: Security & Supply Chain Capabilities

**This is where DPI tools have a significant gap.** None offer security or supply chain analysis.

| Feature | DX | Jellyfish | Swarmia | LinearB | Backstage |
|---------|----|-----------|---------|---------|-----------|
| **SBOM Generation** | ❌ | ❌ | ❌ | ❌ | Plugin |
| **SBOM Analysis** | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Vulnerability Tracking** | ❌ | ❌ | ❌ | ❌ | Plugin |
| **Dependency Health** | ❌ | ❌ | ❌ | ❌ | Plugin |
| **License Compliance** | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Technology Detection** | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Deep Build Inspection** | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Supply Chain Analysis** | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Certificate Analysis** | ❌ | ❌ | ❌ | ❌ | ❌ |

### Key Insight

**No DPI tool offers security or supply chain analysis.** This represents a significant market gap.

---

## Detailed Competitor Analysis

### 1. DX (getdx.com)

**Founded by**: DORA and SPACE framework researchers
**Status**: Acquired by Atlassian (2025)
**Market Position**: Research-backed developer intelligence platform

#### Capability Scores

| Core Capability | Score | Notes |
|-----------------|-------|-------|
| **1. MEASURE** | ⭐⭐⭐⭐⭐ | **Leader** - DX Core 4, DXI, real-time feedback |
| **2. ALIGN** | ⭐⭐⭐ | Limited business alignment features |
| **3. IDENTIFY** | ⭐⭐⭐⭐ | Good bottleneck detection |
| **4. TRACK** | ⭐⭐⭐⭐⭐ | DORA/SPACE research foundation |
| **5. UNDERSTAND** | ⭐⭐ | Limited investment tracking |
| **SECURITY** | ⭐ | None |

#### Key Features
- **DX Core 4**: Proprietary measurement framework
- **DXI**: Developer Experience Index for benchmarking
- **Direct Benchmarking™**: Compare against peer companies
- **Real-time Feedback**: Captures feedback during tool interaction
- **AI Insights**: Analysis, recommendations, custom reports

#### Strengths
- ✅ Strongest research foundation (DORA, SPACE creators)
- ✅ Proven results (6x lead time reduction at Pfizer)
- ✅ Now part of Atlassian ecosystem
- ✅ Best developer survey/feedback capabilities

#### Weaknesses
- ❌ Proprietary/closed source
- ❌ Enterprise pricing only
- ❌ Limited business alignment vs Jellyfish
- ❌ No security or supply chain features
- ❌ No technology detection

#### Sources
- [DX Platform](https://getdx.com/)
- [Atlassian + DX Announcement](https://www.atlassian.com/blog/announcements/atlassian-acquires-dx)

---

### 2. Jellyfish

**Customers**: 500+ organizations, 35,000+ engineers
**Market Position**: Business alignment and strategic resource allocation

#### Capability Scores

| Core Capability | Score | Notes |
|-----------------|-------|-------|
| **1. MEASURE** | ⭐⭐⭐ | Team health but no surveys |
| **2. ALIGN** | ⭐⭐⭐⭐⭐ | **Leader** - OKR, initiatives, quarterly planning |
| **3. IDENTIFY** | ⭐⭐⭐⭐ | Life Cycle Explorer for bottlenecks |
| **4. TRACK** | ⭐⭐⭐⭐ | Full DORA implementation |
| **5. UNDERSTAND** | ⭐⭐⭐⭐⭐ | **Leader** - Patented allocation, CapEx |
| **SECURITY** | ⭐ | None |

#### Key Features
- **Patented Resource Allocation**: Automatic analysis of engineering signals
- **Capacity Planner**: Multi-deliverable planning
- **Software Capitalization**: Audit-ready CapEx reports
- **Life Cycle Explorer**: Bottleneck identification
- **DevFinOps**: Self-serve financial reporting

#### Strengths
- ✅ Strongest business outcome focus
- ✅ Patented resource allocation algorithm
- ✅ Comprehensive software capitalization (audit-ready)
- ✅ Large customer base with proven scale

#### Weaknesses
- ❌ Proprietary/closed source
- ❌ No developer surveys (unlike DX, Swarmia)
- ❌ Complex setup and onboarding
- ❌ Expensive enterprise pricing
- ❌ No security features

#### Sources
- [Jellyfish Platform](https://jellyfish.co/)
- [Engineering Management Platform](https://jellyfish.co/platform/engineering-management-platform/)

---

### 3. Swarmia

**Customers**: Miro, Docker, Webflow
**Market Position**: Engineering effectiveness with developer experience focus

#### Capability Scores

| Core Capability | Score | Notes |
|-----------------|-------|-------|
| **1. MEASURE** | ⭐⭐⭐⭐⭐ | 32-question surveys, AI tool tracking |
| **2. ALIGN** | ⭐⭐⭐⭐ | Investment tracking, initiatives |
| **3. IDENTIFY** | ⭐⭐⭐⭐⭐ | **Leader** - Working agreements, automation |
| **4. TRACK** | ⭐⭐⭐⭐ | Full DORA + CI insights |
| **5. UNDERSTAND** | ⭐⭐⭐⭐ | FTE model, cost per project |
| **SECURITY** | ⭐ | None |

#### Key Features
- **32-Question Survey Framework**: Comprehensive DX measurement
- **Working Agreements**: Define and monitor team norms
- **AI Tool Tracking**: Copilot/Cursor adoption monitoring
- **FTE-based Effort Model**: New 2024 feature
- **GitHub-Slack Integration**: Two-way, automated reminders

#### Strengths
- ✅ Best balance of quantitative + qualitative
- ✅ Strongest workflow automation (working agreements)
- ✅ AI tool adoption measurement
- ✅ Modern, clean interface
- ✅ Strong CI/CD insights

#### Weaknesses
- ❌ Proprietary/closed source
- ❌ Per-engineer pricing ($39/mo) - expensive at scale
- ❌ Less business alignment depth than Jellyfish
- ❌ No security features

#### Sources
- [Swarmia](https://www.swarmia.com/)
- [2024 in Review](https://www.swarmia.com/blog/2024-in-review/)

---

### 4. LinearB

**Customers**: 3,000+ engineering leaders
**Market Position**: DORA metrics and workflow automation

#### Capability Scores

| Core Capability | Score | Notes |
|-----------------|-------|-------|
| **1. MEASURE** | ⭐⭐ | Team-level only, no surveys |
| **2. ALIGN** | ⭐⭐ | Limited business features |
| **3. IDENTIFY** | ⭐⭐⭐⭐ | gitStream automation |
| **4. TRACK** | ⭐⭐⭐⭐⭐ | **Leader** - FREE DORA metrics |
| **5. UNDERSTAND** | ⭐⭐ | Basic time tracking |
| **SECURITY** | ⭐ | None |

#### Key Features
- **Free DORA Dashboard**: Only vendor with free DORA
- **gitStream**: Automated code review routing
- **Leading Indicators**: Merge frequency, PR size
- **Team-Level Focus**: No individual micromanagement

#### Strengths
- ✅ **FREE DORA metrics** - unique in market
- ✅ Strong workflow automation (gitStream)
- ✅ Fast time-to-value
- ✅ Proven ROI (47% cycle time decrease)

#### Weaknesses
- ❌ Proprietary/closed source
- ❌ Limited business alignment
- ❌ No developer surveys
- ❌ Git-centric only
- ❌ No security features

#### Sources
- [LinearB](https://linearb.io/)
- [DORA Metrics Platform](https://linearb.io/platform/dora-metrics)

---

## Internal Developer Portal Comparison

### Backstage.io (IDP Comparison)

**What It Is**: Open-source framework for building Internal Developer Portals (IDPs)
**Created By**: Spotify (donated to CNCF)
**Note**: Backstage is NOT a DPI tool - it's an IDP. Comparing it highlights the gaps.

#### Capability Scores

| Core Capability | Score | Notes |
|-----------------|-------|-------|
| **1. MEASURE** | ⭐ | None native |
| **2. ALIGN** | ⭐ | None |
| **3. IDENTIFY** | ⭐⭐ | Via plugins only |
| **4. TRACK** | ⭐⭐ | Via plugins only |
| **5. UNDERSTAND** | ⭐ | None |
| **SECURITY** | ⭐⭐ | Via plugins (fragmented) |
| **DISCOVERABILITY** | ⭐⭐⭐⭐⭐ | **Core strength** |

#### Core Features
- **Software Catalog**: Centralized service inventory
- **TechDocs**: Documentation alongside code
- **Software Templates**: Project scaffolding
- **Plugin Ecosystem**: 100+ plugins

#### Key Insight: The Gap

**Backstage focuses on DISCOVERABILITY (what do we have?)**
**DPI tools focus on PRODUCTIVITY (how well is it working?)**
**Neither focuses on SOFTWARE ANALYSIS (what is it made of? is it secure?)**

#### Sources
- [Backstage.io](https://backstage.io/)
- [What is Backstage?](https://backstage.io/docs/overview/what-is-backstage/)

---

## Gibson Powers: Strategic Positioning

### What Gibson Powers Is

Gibson Powers is the **open-source software analysis toolkit** from Crash Override. It is:

| Aspect | Description |
|--------|-------------|
| **Type** | Software analysis toolkit (NOT a DPI tool) |
| **License** | GPL-3.0 (100% open source) |
| **Role** | Free, open-source component of Crash Override platform |
| **Purpose** | Analyze software: what it's made of, how it's built, is it secure |
| **AI Features** | RAG-powered analysis with Claude integration |
| **On-ramp** | Entry point to commercial Crash Override platform |

### What Gibson Powers Analyzes

Gibson Powers provides **analyzers** that answer questions DPI tools cannot:

| Analyzer | Question It Answers |
|----------|---------------------|
| **Technology Identification** | "What technologies does this software use?" (112+ technologies) |
| **Supply Chain Scanner** | "What dependencies does this have? Are they vulnerable?" |
| **Code Ownership** | "Who owns what code? Who should review changes?" |
| **Certificate Analyser** | "Are certificates valid? When do they expire?" |
| **DORA Metrics** | "What are the delivery metrics for this repository?" |
| **Legal Review** | "What licenses are in use? Are there compliance issues?" |

### How Gibson Powers Differs from DPI Tools

| Aspect | DPI Tools | Gibson Powers |
|--------|-----------|---------------|
| **Primary Question** | "How productive is my team?" | "What is this software made of?" |
| **Focus** | People & process metrics | Software & security analysis |
| **Data Sources** | Git, issue trackers, surveys | SBOM, code, configs, manifests |
| **Output** | Dashboards, reports | Analysis reports, AI insights |
| **Pricing** | Per-seat ($39+/mo) or enterprise | Free (open source) |
| **Deployment** | SaaS only | Self-hosted |
| **Security Focus** | None | Core capability |

### Crash Override Platform Integration

Gibson Powers serves as the **open-source on-ramp** to the commercial Crash Override platform:

```
┌─────────────────────────────────────────────────────────────────┐
│                    CRASH OVERRIDE ECOSYSTEM                      │
│                                                                  │
│  ┌─────────────────────┐      ┌─────────────────────────────┐   │
│  │   GIBSON POWERS     │      │   CRASH OVERRIDE PLATFORM   │   │
│  │   (Open Source)     │ ───► │   (Commercial)              │   │
│  │                     │      │                             │   │
│  │ • Technology ID     │      │ • Enterprise dashboards     │   │
│  │ • Supply Chain      │      │ • Team collaboration        │   │
│  │ • Code Ownership    │      │ • Historical analytics      │   │
│  │ • Certificate       │      │ • Multi-org support         │   │
│  │ • DORA Metrics      │      │ • Advanced AI features      │   │
│  │ • AI Analysis       │      │ • Integrations              │   │
│  │                     │      │ • Support & SLA             │   │
│  └─────────────────────┘      └─────────────────────────────┘   │
│         FREE                         COMMERCIAL                  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Gibson Powers Capabilities

| Capability | Status | Notes |
|------------|--------|-------|
| **Technology Detection** | ✅ Implemented | 112+ technologies, multi-layer detection |
| **SBOM Generation** | ✅ Implemented | Via syft integration |
| **SBOM Analysis** | ✅ Implemented | Vulnerability, license, health scoring |
| **Code Ownership** | ✅ Implemented | CODEOWNERS analysis, contribution patterns |
| **Certificate Analysis** | ✅ Implemented | X.509, chain validation, expiry tracking |
| **DORA Metrics** | ✅ Implemented | Basic git-based metrics |
| **AI Analysis** | ✅ Implemented | RAG-powered Claude integration |
| **Supply Chain Security** | ✅ Implemented | Dependency analysis, provenance |
| **Legal Review** | ✅ Implemented | License detection, compliance |

---

## Gap Analysis: Market Gaps Gibson Powers Addresses

### What DPI Tools Miss

| Gap | DPI Reality | Gibson Powers Solution |
|-----|-------------|------------------------|
| **Security analysis** | Not offered | SBOM, vulnerability tracking, supply chain |
| **Technology detection** | Not offered | 112+ technologies detected automatically |
| **Build inspection** | Git/issue metrics only | Deep CI/CD and build analysis |
| **Certificate management** | Not offered | X.509 analysis, chain validation |
| **License compliance** | Not offered | License detection and compliance checking |
| **Open source option** | All proprietary | GPL-3.0, fully open source |
| **Self-hosted** | SaaS only | Self-hosted by design |

### Complementary Positioning

Gibson Powers **complements** DPI tools rather than competing with them:

| Use Case | DPI Tool Role | Gibson Powers Role |
|----------|---------------|-------------------|
| **M&A Due Diligence** | Team productivity assessment | Technology stack & security analysis |
| **Security Audit** | N/A (not supported) | Full supply chain & vulnerability analysis |
| **Tech Debt Assessment** | Time spent on tech debt | What technologies are outdated |
| **Compliance** | N/A (not supported) | License compliance, certificate validity |
| **New Team Onboarding** | Team metrics history | Codebase analysis, ownership mapping |

### Potential Crash Override Platform Features

Based on DPI market analysis, the commercial Crash Override platform could offer:

| Feature | Inspired By | Crash Override Angle |
|---------|-------------|---------------------|
| **Team Dashboards** | All DPI tools | Security + productivity unified view |
| **Historical Trends** | Jellyfish, DX | Tech stack evolution over time |
| **Benchmarking** | DX Direct™ | Security posture benchmarking |
| **Alerts & Notifications** | Swarmia | Certificate expiry, vulnerability alerts |
| **Investment Tracking** | Jellyfish | Security investment ROI |

---

## Conclusion

### DPI Market Summary

| Capability | Leader |
|------------|--------|
| **1. MEASURE** | DX |
| **2. ALIGN** | Jellyfish |
| **3. IDENTIFY** | Swarmia |
| **4. TRACK** | LinearB (free) |
| **5. UNDERSTAND** | Jellyfish |
| **SECURITY** | **None** |

### Gibson Powers Position

**Gibson Powers is not competing with DPI tools.** It occupies a different space:

| Tool Category | Focus | Example |
|---------------|-------|---------|
| **DPI Tools** | Team productivity | DX, Jellyfish, Swarmia, LinearB |
| **IDPs** | Service discovery | Backstage |
| **Security Tools** | Vulnerability scanning | Snyk, Trivy |
| **Gibson Powers** | Software analysis | Technology, supply chain, security |

### Strategic Value

1. **Unique Capabilities**: Technology detection, supply chain analysis, certificate management - features no DPI tool offers
2. **Open Source**: The only open-source option in this space
3. **On-ramp**: Entry point to commercial Crash Override platform
4. **Complementary**: Works alongside DPI tools, not against them
5. **Security-First**: Built by security experts, security is native not bolted-on

### The Opportunity

DPI tools answer: *"How productive is my team?"*

Gibson Powers answers: *"What is my software made of, and is it secure?"*

**Both questions matter. Gibson Powers answers the one DPI tools ignore.**

---

## Sources

### Gartner
- [Gartner Market Guide for Software Engineering Intelligence Platforms (2024)](https://www.gartner.com/en/documents/5276563)
- [Gartner Peer Insights: SEI Platforms](https://www.gartner.com/reviews/market/software-engineering-intelligence-platforms)

### DPI Tool Sources
- [DX Platform](https://getdx.com/)
- [Atlassian + DX Announcement](https://www.atlassian.com/blog/announcements/atlassian-acquires-dx)
- [Jellyfish](https://jellyfish.co/)
- [Swarmia](https://www.swarmia.com/)
- [Swarmia 2024 Review](https://www.swarmia.com/blog/2024-in-review/)
- [LinearB](https://linearb.io/)
- [LinearB DORA Metrics](https://linearb.io/platform/dora-metrics)

### IDP Sources
- [Backstage.io](https://backstage.io/)
- [What is Backstage?](https://backstage.io/docs/overview/what-is-backstage/)

### Security Tool Sources
- [Anchore SBOM](https://anchore.com/sbom/)
- [OWASP Dependency-Track](https://dependencytrack.org/)
- [Trivy Scanner](https://github.com/aquasecurity/trivy)

---

*Last Updated: 2025-11-25*
