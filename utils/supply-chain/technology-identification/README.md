<!--
Copyright (c) 2025 Crash Override Inc.
https://crashoverride.com
SPDX-License-Identifier: GPL-3.0
-->

# Technology Identification System

**Status**: 🚀 Beta (v0.2.0)
**Created**: 2025-11-23
**Last Updated**: 2025-11-23

Comprehensive technology stack analysis toolkit that identifies technologies, tools, frameworks, and services used in software development through multi-layered detection with confidence scoring.

## Overview

The Technology Identification System analyzes source code repositories and SBOMs to provide complete visibility into your technology stack across 8 major categories:

1. **Business Tools** - CRM, payment processors, communication services, analytics
2. **Developer Tools** - IaC, containers, orchestration, CI/CD, build tools, monitoring
3. **Programming Languages** - Languages, runtimes, type systems, compilers
4. **Cryptographic Libraries** - TLS/SSL, crypto primitives, hashing, JWT, signing
5. **Web Frameworks** - Frontend, backend, API frameworks, authentication
6. **Databases** - Relational, NoSQL, key-value stores, search engines, time series
7. **Cloud Providers** - AWS, GCP, Azure, and their specific services
8. **Message Queues** - Queues, streaming platforms, event buses

### Key Features

- **Multi-Layered Detection** (6 layers with confidence scoring)
- **Evidence-Based Analysis** (file locations, line numbers, code snippets)
- **Version Tracking** (exact versions, EOL detection, security advisories)
- **Risk Assessment** (Critical → High → Medium → Low)
- **Compliance Implications** (export control, licenses, data privacy)
- **AI-Enhanced Analysis** (Claude-powered insights and recommendations)

## Quick Start

```bash
# Install prerequisites
# (Most tools already available from supply chain scanner)

# Run technology identification scan
./technology-identification-analyser.sh --repo owner/repo

# With Claude AI enhancement
export ANTHROPIC_API_KEY="your-api-key"
./technology-identification-analyser.sh --claude --repo owner/repo

# Generate executive report
./technology-identification-analyser.sh \
  --claude \
  --repo owner/repo \
  --format markdown \
  --output tech-stack-report.md \
  --executive-summary
```

## Architecture

### Integration with Supply Chain Infrastructure

The Technology Identification module **leverages existing supply chain libraries** for consistency:

- **Repository Management**: Uses `lib/github.sh` for cloning and GitHub API access
- **SBOM Generation**: Uses `lib/sbom.sh` for consistent package manager detection and SBOM creation
- **Shared Resources**: Single repository clone and SBOM shared across all analyzers
- **Configuration**: Unified `config.json` hierarchy (module → utils → global)

```
supply-chain-scanner.sh (orchestrator)
├── Clone repository once (lib/github.sh) → SHARED_REPO_DIR
├── Generate SBOM once (lib/sbom.sh) → SHARED_SBOM_FILE
└── Run analyzers in sequence:
    ├── vulnerability-analyser.sh
    ├── provenance-analyser.sh
    ├── package-health-analyser.sh
    └── technology-identification-analyser.sh ← Uses shared SBOM + repo
```

### Module Structure

```
technology-identification/
├── README.md                          # This file
├── DESIGN.md                          # Comprehensive design document
│
├── technology-identification-analyser.sh    # Main analyzer script
│   ├── Sources: lib/github.sh, lib/sbom.sh, lib/config-loader.sh
│   └── Uses: SHARED_REPO_DIR, SHARED_SBOM_FILE
│
├── prompts/
│   ├── pattern-extraction.md         # Extract patterns from docs
│   ├── technology-analysis.md        # Analyze repositories
│   └── report-generation.md          # Generate reports
│
├── rag-updater/                      # RAG maintenance tools
│   ├── update-rag.sh                 # Main update script
│   ├── sources/                      # Data source scrapers
│   ├── parsers/                      # Documentation parsers
│   └── generators/                   # Pattern generators
│
└── config.json                       # Configuration settings
```

## Detection Strategy

### Multi-Layered Approach

The system uses 6 detection layers with decreasing confidence levels:

#### Layer 1: Manifest & Lock Files (90-100% Confidence)
**What**: Parse package manager dependency files
**Examples**:
- `package.json`, `package-lock.json` (npm)
- `requirements.txt`, `poetry.lock` (Python)
- `Cargo.toml`, `Cargo.lock` (Rust)
- `go.mod`, `go.sum` (Go)
- `pom.xml`, `build.gradle` (Java)

**Confidence**: 95-100% (declarative source of truth)

#### Layer 2: Configuration Files (80-95% Confidence)
**What**: Pattern match configuration files
**Examples**:
- Infrastructure: `terraform.tf`, `docker-compose.yml`
- CI/CD: `.github/workflows/*.yml`, `.gitlab-ci.yml`
- Build: `webpack.config.js`, `tsconfig.json`

**Confidence**: 85-95% (explicit configuration)

#### Layer 3: Import Statements (60-80% Confidence)
**What**: Parse source code imports
**Examples**:
```javascript
import Stripe from 'stripe';
const express = require('express');
```
```python
import boto3
from twilio.rest import Client
```

**Confidence**: 70-85% (may include unused imports)

#### Layer 4: API Endpoints (60-80% Confidence)
**What**: Match API endpoint patterns
**Examples**:
- `https://api.stripe.com/v1/*` → Stripe
- `https://s3.amazonaws.com/*` → AWS S3
- `https://api.twilio.com/*` → Twilio

**Confidence**: 65-80% (may be examples/test code)

#### Layer 5: Environment Variables (40-60% Confidence)
**What**: Identify environment variable patterns
**Examples**:
- `STRIPE_API_KEY` → Stripe
- `AWS_ACCESS_KEY_ID` → AWS
- `TWILIO_ACCOUNT_SID` → Twilio

**Confidence**: 50-65% (indirect evidence)

#### Layer 6: Comments & Documentation (30-50% Confidence)
**What**: NLP on comments and documentation
**Examples**:
- "Using Salesforce API to sync contacts"
- "Integrated with Stripe for payment processing"

**Confidence**: 35-50% (may be outdated or aspirational)

### Confidence Scoring

**Composite Confidence Calculation**:
```
When multiple evidence types exist:
Composite = (Evidence1 + Evidence2 + ... + EvidenceN) / N × 1.2
(Capped at 100%)

Example - Stripe Detection:
- package.json: 95%
- import statement: 75%
- API endpoint: 65%
→ Composite: (95 + 75 + 65) / 3 × 1.2 = 94%
```

## Usage

### Basic Scanning

```bash
# Scan single repository
./technology-identification-analyser.sh --repo owner/repo

# Scan entire organization
./technology-identification-analyser.sh --org myorg

# Scan with specific format
./technology-identification-analyser.sh \
  --repo owner/repo \
  --format json \
  --output tech-report.json
```

### Advanced Options

```bash
# With Claude AI analysis (requires ANTHROPIC_API_KEY)
export ANTHROPIC_API_KEY="your-api-key"
./technology-identification-analyser.sh \
  --claude \
  --repo owner/repo \
  --format markdown \
  --output tech-stack-report.md

# With executive summary
./technology-identification-analyser.sh \
  --claude \
  --repo owner/repo \
  --executive-summary

# Custom confidence threshold
./technology-identification-analyser.sh \
  --repo owner/repo \
  --min-confidence 70

# Include risk assessment
./technology-identification-analyser.sh \
  --repo owner/repo \
  --risk-assessment

# All options combined
./technology-identification-analyser.sh \
  --claude \
  --org myorg \
  --format markdown \
  --output reports/tech-stacks/ \
  --min-confidence 65 \
  --risk-assessment \
  --executive-summary
```

## Report Audience

### Primary: Head of Engineering
Strategic technology decision-maker who needs:
- **Executive Summary**: High-level findings in business terms
- **Clear Risk Classification**: Critical → High → Medium → Low with business impact
- **Actionable Recommendations**: What to do, when, and why
- **Effort Estimates**: Complexity and timeline for remediation

### Secondary: Internal Audit
Compliance and risk-focused team requiring:
- **Evidence Trail**: File paths, line numbers, detection methods
- **Policy Compliance**: Approved/banned technology violations
- **Regulatory Implications**: Export control, licensing, data privacy
- **Audit-Ready Documentation**: Confidence scores, timestamps, accountability

### Report Characteristics
- **Executive Summary**: 1 page, non-technical language, business-focused
- **Technical Details**: Engineering context with evidence and recommendations
- **Audit Trail**: Structured compliance reporting with policy mapping

## Output Formats

### JSON (Machine-Readable)

```json
{
  "repository": "owner/repo",
  "scan_date": "2024-11-23T10:00:00Z",
  "total_technologies": 47,
  "technologies": [
    {
      "name": "Stripe",
      "category": "business-tools/payment",
      "version": "14.12.0",
      "confidence": 94,
      "evidence": [
        {
          "type": "manifest",
          "confidence": 95,
          "location": "package.json:12",
          "snippet": "\"stripe\": \"^14.12.0\""
        },
        {
          "type": "import",
          "confidence": 85,
          "location": "src/payments.js:3",
          "snippet": "import Stripe from 'stripe';"
        }
      ],
      "risk_level": "low",
      "notes": "Current stable version, no known vulnerabilities"
    }
  ],
  "summary": {
    "by_category": {
      "business-tools": 5,
      "developer-tools": 12,
      "programming-languages": 3,
      "cryptographic-libraries": 2,
      "databases": 4,
      "cloud-providers": 3
    },
    "risk_summary": {
      "critical": 1,
      "high": 3,
      "medium": 8,
      "low": 35
    }
  }
}
```

### Markdown (Executive-Focused)

```markdown
# Technology Stack Report

**Repository**: owner/repo
**Prepared For**: Head of Engineering
**Scan Date**: 2024-11-23
**Total Technologies**: 47

## Executive Summary

Your application uses **47 technologies** across 6 categories. We identified **1 critical issue** requiring immediate attention.

**Key Findings**:
- ✅ **Strengths**: Modern frontend (React 18), containerized deployment
- ⚠️ **Concerns**: Using 3 different payment processors increases cost/complexity
- 🔴 **Critical**: OpenSSL 1.1.1 reached end-of-life in September 2023

**Immediate Action Required**:
Upgrade OpenSSL to version 3.x within 7 days to address critical security vulnerabilities.

## Technology Inventory

### Business Tools (5)
**Payment Processing**:
- Stripe v14.12.0 (Primary - 95% of transactions)
  - Status: ✅ Current version
  - Risk: Low

**Communication**:
- Twilio v4.5.0
  - Status: ✅ Current version
  - Risk: Low

### Cryptographic Libraries (2)
**TLS/SSL**:
- OpenSSL 1.1.1q - 🔴 **CRITICAL**
  - **Issue**: End-of-life since September 2023
  - **Impact**: All HTTPS connections at risk, no security patches
  - **Business Risk**: Data breach potential, regulatory violations (PCI DSS, SOC 2)
  - **Action**: Upgrade to OpenSSL 3.x within 7 days
  - **Effort**: Medium (2-3 days)

## Consolidation Opportunities

### Multiple Payment Processors
**Finding**: 3 payment processors detected
- Stripe (95% of transactions) - Primary
- PayPal (5% of transactions) - Legacy
- Square (unused, legacy integration)

**Recommendation**: Consolidate to Stripe
- **Benefit**: Reduced PCI compliance scope, lower transaction fees, simplified codebase
- **Effort**: High - Customer migration required
- **Timeline**: 120 days
- **Estimated Savings**: 15-20% reduction in payment processing costs

## Policy Compliance

**Compliance Score**: 72/100 (C Grade)
- ✅ Approved: 35 technologies (74%)
- 🔴 Banned: 1 technology (OpenSSL 1.1.1)
- ⚠️ Review Required: 3 technologies
- ⚠️ Unapproved: 8 technologies

**Critical Violation**: OpenSSL 1.1.1 (banned since 2023-09-11)
```

### Table (Console)

```
╭─────────────────┬──────────┬────────────┬──────────────╮
│ Technology      │ Version  │ Confidence │ Risk Level   │
├─────────────────┼──────────┼────────────┼──────────────┤
│ Stripe          │ 14.12.0  │ 94%        │ 🟢 Low       │
│ OpenSSL         │ 1.1.1    │ 85%        │ 🔴 Critical  │
│ Terraform       │ 1.6.4    │ 90%        │ 🟢 Low       │
╰─────────────────┴──────────┴────────────┴──────────────╯
```

## RAG Library Structure

Technology detection patterns are stored in a hierarchical RAG library:

```
rag/technology-identification/
├── business-tools/
│   ├── crm/
│   │   ├── salesforce/
│   │   │   ├── api-patterns.md
│   │   │   ├── import-patterns.md
│   │   │   ├── config-patterns.md
│   │   │   ├── env-variables.md
│   │   │   └── versions.md
│   │   ├── hubspot/
│   │   └── zoho/
│   ├── payment/
│   │   ├── stripe/
│   │   ├── paypal/
│   │   └── square/
│   └── communication/
│       ├── twilio/
│       ├── sendgrid/
│       └── slack/
│
├── developer-tools/
│   ├── infrastructure/
│   │   ├── terraform/
│   │   ├── ansible/
│   │   └── cloudformation/
│   ├── containers/
│   │   ├── docker/
│   │   └── kubernetes/
│   └── cicd/
│       ├── github-actions/
│       ├── gitlab-ci/
│       └── jenkins/
│
├── cryptographic-libraries/
│   ├── tls/
│   │   ├── openssl/
│   │   ├── libressl/
│   │   └── boringssl/
│   └── crypto/
│       ├── libsodium/
│       └── crypto-js/
│
└── [other categories...]
```

Each technology directory contains:
- **api-patterns.md** - API endpoint patterns and authentication
- **import-patterns.md** - Import/require syntax across languages
- **config-patterns.md** - Configuration file patterns
- **env-variables.md** - Environment variable naming conventions
- **versions.md** - Version history, EOL dates, breaking changes

## RAG Update Mechanism

The RAG library is kept current through automated updates:

```bash
# Update all RAG patterns (weekly automated run)
./rag-updater/update-rag.sh --auto

# Update specific technology
./rag-updater/update-rag.sh --tech stripe --source npm

# Update from API documentation
./rag-updater/update-rag.sh --tech aws --source api-docs

# Add new technology
./rag-updater/add-technology.sh \
  --name datadog \
  --category developer-tools/monitoring \
  --docs-url https://docs.datadoghq.com
```

## Risk Assessment

Technologies are classified into risk levels:

### 🔴 Critical Risk
- End-of-life (EOL) with known CVEs
- Deprecated cryptographic libraries (OpenSSL 1.0.x, 1.1.x)
- Unsupported runtimes (Python 2.7, Node.js <16)
- AGPL libraries in proprietary software

### 🟠 High Risk
- Approaching EOL (within 6 months)
- Multiple major versions behind
- Deprecated but functioning
- Community-abandoned projects

### 🟡 Medium Risk
- One major version behind
- Security advisories (non-critical)
- Limited community support

### 🟢 Low Risk
- Current stable/LTS version
- Active maintenance
- No known vulnerabilities
- Strong community support

## Compliance Implications

### Export Control (ITAR/EAR)
- Flags strong cryptography (>= 256-bit)
- Identifies OpenSSL, BoringSSL, libsodium usage
- Most modern crypto is EAR99 (export-friendly)

### License Compliance
- AGPL in SaaS → Network copyleft risk
- GPL in proprietary → License violation risk
- Proprietary business tools → License cost

### Data Privacy (GDPR/CCPA)
- Analytics tools → PII handling requirements
- CRM systems → Customer data storage
- Communication services → Message content privacy

### Financial (PCI DSS)
- Payment processors → Secure key storage
- Credit card handling → Compliance requirements

## Integration with Supply Chain Scanner

The Technology Identification module integrates with the existing supply chain scanner:

```bash
# Run as part of supply chain analysis
./supply-chain-scanner.sh \
  --all \
  --technology \
  --repo owner/repo

# Technology identification + vulnerability analysis
./supply-chain-scanner.sh \
  --vulnerability \
  --technology \
  --repo owner/repo

# Full stack analysis with Claude AI
export ANTHROPIC_API_KEY="your-api-key"
./supply-chain-scanner.sh \
  --all \
  --technology \
  --claude \
  --repo owner/repo
```

## Examples

### Example 1: Initial Technology Audit

```bash
# Comprehensive technology stack analysis
export ANTHROPIC_API_KEY="your-key"
./technology-identification-analyser.sh \
  --claude \
  --repo myorg/myapp \
  --format markdown \
  --output reports/tech-audit-2024-11.md \
  --executive-summary \
  --risk-assessment

# Review findings
cat reports/tech-audit-2024-11.md
```

### Example 2: CI/CD Integration

```bash
# Fast technology check with strict thresholds
./technology-identification-analyser.sh \
  --repo owner/repo \
  --format json \
  --output tech-check.json \
  --min-confidence 80 \
  --fail-on-critical

# Exit code 1 if critical risks found
```

### Example 3: Organization-Wide Scan

```bash
# Scan all repositories in organization
./technology-identification-analyser.sh \
  --org myorg \
  --format json \
  --output reports/org-tech-stack/ \
  --parallel

# Generate consolidated report
./technology-identification-analyser.sh \
  --claude \
  --consolidate reports/org-tech-stack/*.json \
  --output org-tech-summary.md \
  --executive-summary
```

### Example 4: Technology Migration Planning

```bash
# Identify all OpenSSL 1.1.x usage across organization
./technology-identification-analyser.sh \
  --org myorg \
  --filter "OpenSSL 1.1" \
  --format csv \
  --output openssl-migration-plan.csv

# Generate migration recommendations
./technology-identification-analyser.sh \
  --claude \
  --input openssl-migration-plan.csv \
  --migration-plan \
  --output openssl-upgrade-guide.md
```

## Prerequisites

### Required Tools

Already installed from supply chain scanner:
- **jq** - JSON processor
- **gh** - GitHub CLI
- **syft** - SBOM generator
- **osv-scanner** - Vulnerability scanner

### Optional Tools

- **ANTHROPIC_API_KEY** - For Claude-enhanced analysis
- **cosign** - Signature verification
- **rekor-cli** - Transparency log

### GitHub Authentication

```bash
# Authenticate with GitHub CLI
gh auth login

# Or provide Personal Access Token in config.json
```

## Configuration

```json
{
  "technology_identification": {
    "min_confidence": 60,
    "categories": [
      "business-tools",
      "developer-tools",
      "programming-languages",
      "cryptographic-libraries",
      "web-frameworks",
      "databases",
      "cloud-providers",
      "message-queues"
    ],
    "risk_assessment": true,
    "compliance_checks": true,
    "output_format": "json",
    "rag_update_frequency": "weekly",
    "version_tracking": true
  }
}
```

## Development Status

**Current Status**: 🚀 Beta (v0.2.0)

### Phase 1: Design & Architecture ✅
- [x] Multi-layered detection strategy
- [x] Confidence scoring system
- [x] RAG library structure
- [x] Category taxonomy
- [x] Evidence documentation format
- [x] Risk assessment framework

### Phase 2: Implementation ✅
- [x] Core analyzer script (998 lines)
- [x] Confidence scoring engine
- [x] Report generation (JSON, Markdown, Table formats)
- [x] Integration with supply chain scanner
- [ ] RAG pattern extraction tools (planned)
- [ ] Version tracking system (planned)

### Phase 3: RAG Library Population (In Progress)
- [x] Business tools patterns (Stripe - 4 files)
- [x] Developer tools patterns (Terraform, Docker - 6 files)
- [x] Cryptographic library patterns (OpenSSL - 3 files)
- [x] Cloud provider patterns (AWS SDK - 4 files)
- [ ] Programming language patterns (Python, JavaScript, Go)
- [ ] Additional business tools (Salesforce, Twilio, SendGrid)
- [ ] Additional developer tools (Kubernetes, Ansible)

**Total RAG Patterns**: 17 files (~31,000 lines) covering 5 technologies

### Phase 4: Testing & Validation (In Progress)
- [x] Test on real-world repositories (crashappsec/chalk)
- [x] Integration testing with supply chain scanner
- [ ] Validate confidence scoring accuracy
- [ ] Benchmark against manual audits
- [ ] Performance optimization for large repositories

### Phase 5: Production Deployment (Pending)
- [x] Core documentation (README, DESIGN, PROGRESS)
- [ ] Expand RAG library to 50+ technologies
- [ ] CI/CD integration examples
- [ ] Training materials
- [ ] Production rollout

## Documentation

- [DESIGN.md](./DESIGN.md) - Comprehensive design document
- [prompts/pattern-extraction.md](./prompts/pattern-extraction.md) - Extract patterns from docs
- [prompts/technology-analysis.md](./prompts/technology-analysis.md) - Analyze repositories
- [prompts/report-generation.md](./prompts/report-generation.md) - Generate reports
- [Skill File](../../skills/technology-identification/technology-identification.skill) - Claude AI skill definition

## Contributing

See [CONTRIBUTING.md](../../../CONTRIBUTING.md) for development guidelines.

### Adding New Technology Patterns

```bash
# Create new technology pattern directory
mkdir -p rag/technology-identification/category/subcategory/technology-name

# Generate pattern files from documentation
./rag-updater/extract-patterns.sh \
  --technology technology-name \
  --category category/subcategory \
  --docs-url https://docs.example.com

# Validate patterns
./rag-updater/validate-patterns.sh \
  rag/technology-identification/category/subcategory/technology-name/

# Submit pull request
git checkout -b add-technology-name-patterns
git add rag/technology-identification/category/subcategory/technology-name/
git commit -m "feat: Add technology-name identification patterns"
git push origin add-technology-name-patterns
```

## License

GPL-3.0 - See [LICENSE](../../../LICENSE) for details.

## Support

- Issues: [GitHub Issues](https://github.com/crashappsec/gibson-powers/issues)
- Documentation: [Wiki](https://github.com/crashappsec/gibson-powers/wiki)

## Version

**Current Version**: 0.2.0 (Beta)

See [CHANGELOG.md](./CHANGELOG.md) for version history.

---

**Created**: 2025-11-23
**Last Updated**: 2025-11-23
**Status**: 🚀 Beta - Core implementation complete, RAG library expansion ongoing
