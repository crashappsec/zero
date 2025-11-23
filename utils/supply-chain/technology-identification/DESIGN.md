<!--
Copyright (c) 2025 Crash Override Inc.
https://crashoverride.com
SPDX-License-Identifier: GPL-3.0
-->

# Technology Identification System - Design Document

**Status**: 🎨 Design Phase
**Version**: 0.1.0
**Date**: 2025-11-23

## Overview

The Technology Identification System analyzes source code repositories and SBOMs to identify technologies, tools, frameworks, and services used by developers. It provides comprehensive visibility into the technology stack, including business tools, developer tools, programming languages, cryptographic libraries, and cloud services.

## Goals

1. **Comprehensive Detection**: Identify technologies across all categories
2. **Evidence-Based**: Record specific evidence with confidence scoring
3. **Version Tracking**: Track technology versions and update RAG documentation
4. **Pattern Matching**: Use API documentation and configuration patterns for detection
5. **Actionable Intelligence**: Provide insights for security, compliance, and architecture decisions
6. **Technology Governance**: Enforce approved/banned technology policies and identify policy violations
7. **Consolidation Opportunities**: Identify multiple tools in the same category for rationalization
8. **Deprecation Detection**: Flag old, deprecated, and end-of-life technologies requiring replacement
9. **Policy Compliance**: Maintain and enforce organizational technology standards

## Technology Categories

### 1. Business Tools & Services
- **CRM**: Salesforce, HubSpot, Zoho
- **Payment Processing**: Stripe, PayPal, Square, Braintree
- **Communication**: Twilio, SendGrid, Mailchimp, Slack API
- **Analytics**: Google Analytics, Mixpanel, Segment, Amplitude
- **Customer Support**: Zendesk, Intercom, Freshdesk
- **Marketing**: Marketo, Pardot, ActiveCampaign

### 2. Developer Tools
- **Infrastructure as Code**: Terraform, Pulumi, CloudFormation, Ansible
- **Containers**: Docker, Podman, containerd
- **Orchestration**: Kubernetes, Docker Compose, Nomad
- **CI/CD**: GitHub Actions, GitLab CI, Jenkins, CircleCI, Travis CI
- **Version Control**: Git, GitHub, GitLab, Bitbucket
- **Build Tools**: Webpack, Vite, esbuild, Rollup, Make, Gradle, Maven
- **Testing Frameworks**: Jest, Mocha, pytest, JUnit, RSpec
- **Monitoring**: Prometheus, Grafana, Datadog, New Relic, Sentry

### 3. Programming Languages & Runtimes
- **Languages**: Python, JavaScript, TypeScript, Go, Rust, Java, Ruby, PHP, C/C++
- **Runtimes**: Node.js, Deno, Bun, JVM, Python interpreter versions
- **Language Features**: async/await, type systems, memory management

### 4. Cryptographic Libraries
- **TLS/SSL**: OpenSSL, LibreSSL, BoringSSL
- **Crypto Libraries**: libsodium, NaCl, crypto++, ring (Rust)
- **Hashing**: bcrypt, argon2, scrypt
- **JWT**: jsonwebtoken, PyJWT, jose
- **Signing**: GPG, cosign, sigstore

### 5. Web Frameworks & Libraries
- **Frontend**: React, Vue, Angular, Svelte, Next.js, Nuxt.js
- **Backend**: Express, Django, Flask, FastAPI, Rails, Spring Boot, Laravel
- **API**: GraphQL, REST, gRPC, tRPC
- **Authentication**: OAuth, SAML, OpenID Connect, Passport.js, Auth0

### 6. Databases & Data Stores
- **Relational**: PostgreSQL, MySQL, MariaDB, SQLite, Oracle, SQL Server
- **NoSQL**: MongoDB, Cassandra, Couchbase
- **Key-Value**: Redis, Memcached, etcd
- **Search**: Elasticsearch, Solr, Meilisearch
- **Time Series**: InfluxDB, TimescaleDB, Prometheus TSDB

### 7. Cloud Providers & Services
- **Providers**: AWS, Google Cloud Platform, Azure, DigitalOcean, Heroku
- **Specific Services**:
  - AWS: S3, Lambda, EC2, RDS, DynamoDB, SQS, SNS
  - GCP: Cloud Storage, Cloud Functions, Compute Engine, Cloud SQL
  - Azure: Blob Storage, Functions, VMs, SQL Database
- **CDN**: CloudFlare, Fastly, Akamai

### 8. Message Queues & Event Systems
- **Queue**: RabbitMQ, Apache Kafka, AWS SQS, Google Pub/Sub, Azure Service Bus
- **Streaming**: Apache Kafka, Apache Pulsar, NATS, Redis Streams
- **Event Bus**: EventBridge, Apache Camel

## Detection Strategy

### Multi-Layered Detection Approach

#### Layer 1: Manifest & Lock File Analysis (High Confidence: 90-100%)
**Method**: Parse package manager files
- `package.json` / `package-lock.json` (npm)
- `requirements.txt` / `poetry.lock` (Python)
- `Cargo.toml` / `Cargo.lock` (Rust)
- `go.mod` / `go.sum` (Go)
- `pom.xml` / `build.gradle` (Java)

**Evidence**: Declared dependencies with exact versions
**Confidence**: 95-100% (declarative source of truth)

#### Layer 2: Configuration File Detection (High Confidence: 80-95%)
**Method**: Pattern matching configuration files
- Infrastructure: `terraform.tf`, `docker-compose.yml`, `Dockerfile`, `.gitlab-ci.yml`
- Cloud: `serverless.yml`, `.aws/config`, `gcloud config`
- Build: `webpack.config.js`, `tsconfig.json`, `Makefile`
- Testing: `jest.config.js`, `pytest.ini`, `.rspec`

**Evidence**: Configuration file presence and content
**Confidence**: 85-95% (explicit configuration)

#### Layer 3: Import & Reference Analysis (Medium Confidence: 60-80%)
**Method**: Parse source code imports and API calls
- Python: `import stripe`, `from twilio.rest import Client`
- JavaScript: `import express from 'express'`, `const stripe = require('stripe')`
- Go: `import "github.com/aws/aws-sdk-go/aws"`
- Rust: `use tokio::runtime::Runtime`

**Evidence**: Import statements and package references
**Confidence**: 70-85% (may be unused imports)

#### Layer 4: API Endpoint Detection (Medium Confidence: 60-80%)
**Method**: Pattern matching API endpoints in code
- AWS: `https://s3.amazonaws.com`, `https://sqs.us-east-1.amazonaws.com`
- Stripe: `https://api.stripe.com`
- Twilio: `https://api.twilio.com`
- SendGrid: `https://api.sendgrid.com`

**Evidence**: API endpoint strings in source code
**Confidence**: 65-80% (may be example/test code)

#### Layer 5: Environment Variable Pattern Detection (Low-Medium Confidence: 40-60%)
**Method**: Identify environment variable naming patterns
- `STRIPE_API_KEY`, `STRIPE_SECRET_KEY` → Stripe
- `TWILIO_ACCOUNT_SID`, `TWILIO_AUTH_TOKEN` → Twilio
- `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` → AWS
- `DATABASE_URL` → Database (need additional context)

**Evidence**: Environment variable names in `.env.example`, config files
**Confidence**: 50-65% (indirect evidence)

#### Layer 6: Comment & Documentation Analysis (Low Confidence: 30-50%)
**Method**: NLP on comments and documentation
- "Using Salesforce API to sync contacts"
- "Integrated with Stripe for payment processing"
- "Deployed to AWS Lambda"

**Evidence**: Documentation mentions
**Confidence**: 35-50% (may be aspirational or outdated)

### Confidence Scoring System

```
High Confidence (80-100%):
  - Manifest file declaration: 100%
  - Lock file entry: 95%
  - Configuration file: 90%
  - Direct import with usage: 85%
  - API endpoint with credentials: 80%

Medium Confidence (50-79%):
  - Import statement without usage verification: 75%
  - Environment variable pattern: 65%
  - API endpoint without auth: 60%
  - Binary detection (compiled tool): 55%

Low Confidence (0-49%):
  - Documentation mention: 45%
  - Comment reference: 40%
  - Similar naming pattern: 30%
  - Indirect evidence: 20%
```

### Composite Scoring

When multiple detection methods identify the same technology, use **Bayesian confidence aggregation**:

```
P(tech | evidence) = (P(evidence1) + P(evidence2) + ... + P(evidenceN)) / N * 1.2

Max confidence: 100%

Example:
- Stripe in package.json: 95%
- import stripe: 75%
- API endpoint: 65%
→ Composite: (95 + 75 + 65) / 3 * 1.2 = 94% (capped at 100%)
```

## Technology Governance & Policy Enforcement

### Overview

The Technology Identification System includes governance capabilities to enforce organizational technology standards, identify consolidation opportunities, and flag deprecated technologies.

### Approved/Banned Technology Lists

Organizations can maintain policies for allowed and prohibited technologies.

#### Policy Configuration Format

```json
{
  "technology_policy": {
    "approved": [
      {
        "name": "stripe",
        "category": "business-tools/payment",
        "reason": "Standardized payment processor",
        "approved_versions": [">=13.0.0"],
        "expires": "2026-12-31"
      },
      {
        "name": "terraform",
        "category": "developer-tools/infrastructure",
        "reason": "Standard IaC tool",
        "approved_versions": [">=1.5.0", "<2.0.0"],
        "required_for": ["infrastructure provisioning"]
      },
      {
        "name": "openssl",
        "category": "cryptographic-libraries/tls",
        "approved_versions": [">=3.0.0"],
        "reason": "Current LTS only"
      }
    ],
    "banned": [
      {
        "name": "openssl",
        "category": "cryptographic-libraries/tls",
        "versions": ["<3.0.0"],
        "reason": "End-of-life, known vulnerabilities",
        "banned_date": "2023-09-11",
        "alternative": "OpenSSL 3.x"
      },
      {
        "name": "jquery",
        "category": "web-frameworks/frontend",
        "versions": ["<3.0.0"],
        "reason": "Deprecated, security issues",
        "alternative": "Modern JavaScript or React/Vue"
      },
      {
        "name": "request",
        "category": "developer-tools",
        "versions": ["*"],
        "reason": "Deprecated npm package",
        "alternative": "axios or node-fetch"
      },
      {
        "name": "moment",
        "category": "web-frameworks",
        "versions": ["*"],
        "reason": "Bundle size, deprecated",
        "alternative": "date-fns or Temporal API"
      }
    ],
    "review_required": [
      {
        "name": "salesforce",
        "category": "business-tools/crm",
        "reason": "High cost, license review required",
        "approval_process": "Director level sign-off"
      },
      {
        "name": "mongodb",
        "category": "databases/nosql",
        "reason": "SSPL license considerations",
        "approval_process": "Legal review"
      }
    ],
    "preferred": [
      {
        "category": "developer-tools/containers",
        "preferred": ["docker"],
        "alternatives": ["podman", "containerd"],
        "reason": "Standardization and tooling support"
      },
      {
        "category": "programming-languages",
        "preferred": ["python", "go", "typescript"],
        "reason": "Team expertise and hiring"
      }
    ]
  }
}
```

#### Policy Enforcement Modes

**Strict Mode**:
- Fail scan if banned technology detected
- Require approval for non-approved technologies
- Block deployments with policy violations

**Advisory Mode** (default):
- Report policy violations as warnings
- Suggest alternatives
- Track violations over time

**Audit Mode**:
- Log all findings
- Generate compliance reports
- No blocking

#### Policy Violation Report

```json
{
  "repository": "owner/repo",
  "policy_violations": [
    {
      "severity": "critical",
      "type": "banned_technology",
      "technology": "OpenSSL 1.1.1",
      "category": "cryptographic-libraries/tls",
      "version": "1.1.1q",
      "confidence": 85,
      "evidence": [
        {
          "location": "/usr/lib/libssl.so.1.1",
          "type": "binary"
        }
      ],
      "policy": {
        "banned_date": "2023-09-11",
        "reason": "End-of-life, known vulnerabilities",
        "alternative": "OpenSSL 3.x"
      },
      "action_required": "Immediate upgrade to OpenSSL 3.x",
      "timeline": "0-7 days"
    },
    {
      "severity": "high",
      "type": "unapproved_technology",
      "technology": "MongoDB 4.2",
      "category": "databases/nosql",
      "version": "4.2.24",
      "confidence": 90,
      "evidence": [
        {
          "location": "docker-compose.yml:15",
          "snippet": "image: mongo:4.2"
        }
      ],
      "policy": {
        "status": "review_required",
        "reason": "SSPL license considerations",
        "approval_process": "Legal review"
      },
      "action_required": "Obtain legal approval or migrate to approved alternative",
      "timeline": "30 days"
    }
  ],
  "summary": {
    "critical": 1,
    "high": 2,
    "medium": 5,
    "low": 10
  }
}
```

### Technology Consolidation Detection

Identify multiple tools serving the same purpose for rationalization opportunities.

#### Consolidation Rules

```json
{
  "consolidation_rules": [
    {
      "category": "business-tools/payment",
      "max_allowed": 1,
      "rationale": "Single payment processor reduces complexity and cost",
      "exceptions": ["Multi-region requirements", "Compliance mandates"]
    },
    {
      "category": "developer-tools/cicd",
      "max_allowed": 2,
      "rationale": "Standardize on primary CI/CD platform",
      "preferred": ["github-actions", "gitlab-ci"]
    },
    {
      "category": "databases/relational",
      "max_allowed": 2,
      "rationale": "Limit database diversity for operational efficiency"
    },
    {
      "subcategory": "http-client",
      "max_allowed": 1,
      "rationale": "Single HTTP client reduces bundle size"
    },
    {
      "subcategory": "date-library",
      "max_allowed": 1,
      "rationale": "Standardize date handling"
    }
  ]
}
```

#### Consolidation Opportunities Report

```markdown
## Technology Consolidation Opportunities

### 1. Multiple HTTP Clients Detected

**Finding**: 3 different HTTP client libraries in use
- **axios** v0.27.2 - Used in 45 files (Primary)
- **node-fetch** v2.6.7 - Used in 12 files
- **request** v2.88.0 - Used in 3 files (DEPRECATED)

**Policy**: max_allowed = 1 per subcategory
**Violation**: Exceeds limit by 2 technologies

**Analysis**:
- `axios` is the dominant library (75% of usage)
- `request` is deprecated (security risk)
- `node-fetch` adds ~40KB to bundle

**Recommendation**: Standardize on `axios`
- Migrate 12 `node-fetch` usages to `axios`
- Remove deprecated `request` library
- **Benefit**: -40KB bundle size, improved maintainability
- **Effort**: Low - Simple API migration
- **Timeline**: 30 days

### 2. Duplicate Date Handling Libraries

**Finding**: Both `moment.js` and `date-fns` detected
- **moment.js** v2.29.4 - Used in 23 files (Legacy)
- **date-fns** v2.30.0 - Used in 15 files (Modern)

**Policy**: max_allowed = 1 per subcategory
**Violation**: Exceeds limit by 1 technology

**Analysis**:
- `moment` is deprecated and has large bundle (67KB)
- `date-fns` is modern, tree-shakeable, actively maintained
- Mixed usage creates confusion

**Recommendation**: Migrate to `date-fns`
- Replace 23 `moment` usages with `date-fns`
- **Benefit**: -67KB bundle size, better performance
- **Effort**: Medium - Test date formatting edge cases
- **Timeline**: 60 days

### 3. Multiple Payment Processors

**Finding**: 2 payment processors detected
- **Stripe** v14.12.0 - Primary (95% of transactions)
- **PayPal** v2.0.1 - Legacy integration (5%)

**Policy**: max_allowed = 1 for payment category
**Violation**: Exceeds limit by 1 technology

**Analysis**:
- Stripe handles majority of payments
- PayPal integration maintenance overhead
- Dual PCI compliance scope

**Recommendation**: Consolidate to Stripe
- Migrate PayPal customers to Stripe
- **Benefit**: Reduced compliance scope, lower fees, simplified codebase
- **Effort**: High - Customer migration required
- **Timeline**: 120 days
- **Risk**: Customer communication required

### 4. Cloud Provider Sprawl

**Finding**: 3 cloud providers in use
- **AWS** - Primary (90% of infrastructure)
- **Google Cloud** - Firebase only (8%)
- **Azure** - Single blob storage (2%)

**Policy**: Preferred single cloud provider for cost optimization
**Violation**: Advisory level

**Analysis**:
- AWS is dominant platform
- GCP used only for Firebase
- Azure used for single blob storage service

**Recommendation**: Consolidate to AWS
- Migrate Firebase → AWS Amplify/AppSync
- Migrate Azure Blob → AWS S3
- **Benefit**: 15-20% cost reduction, simplified architecture
- **Effort**: High - Service migration and testing
- **Timeline**: 90-120 days
```

### Deprecation and End-of-Life Tracking

Automatically track and flag deprecated, EOL, and outdated technologies.

#### Deprecation Database Format

```json
{
  "deprecation_database": [
    {
      "technology": "openssl",
      "versions": {
        "1.0.x": {
          "status": "end-of-life",
          "eol_date": "2019-12-31",
          "latest_version": "1.0.2u",
          "replacement": "3.0.x or 3.1.x",
          "critical_cves": [
            {"cve": "CVE-2022-XXXX", "cvss": 9.8},
            {"cve": "CVE-2021-YYYY", "cvss": 7.5}
          ]
        },
        "1.1.x": {
          "status": "end-of-life",
          "eol_date": "2023-09-11",
          "latest_version": "1.1.1w",
          "replacement": "3.0.x or 3.1.x",
          "critical_cves": [
            {"cve": "CVE-2023-XXXX", "cvss": 9.1}
          ]
        },
        "3.0.x": {
          "status": "lts",
          "eol_date": "2026-09-07",
          "latest_version": "3.0.12",
          "notes": "Long-term support"
        }
      }
    },
    {
      "technology": "node.js",
      "versions": {
        "14.x": {
          "status": "end-of-life",
          "eol_date": "2023-04-30",
          "replacement": "20.x LTS",
          "migration_guide": "https://nodejs.org/en/download/releases/"
        },
        "16.x": {
          "status": "end-of-life",
          "eol_date": "2023-09-11",
          "replacement": "20.x LTS"
        },
        "18.x": {
          "status": "lts",
          "eol_date": "2025-04-30",
          "notes": "Maintenance LTS"
        },
        "20.x": {
          "status": "lts",
          "eol_date": "2026-04-30",
          "notes": "Active LTS (recommended)"
        }
      }
    },
    {
      "technology": "jquery",
      "versions": {
        "1.x": {
          "status": "deprecated",
          "deprecated_date": "2016-06-09",
          "replacement": "Modern JavaScript or frameworks",
          "security_issues": true
        },
        "2.x": {
          "status": "deprecated",
          "deprecated_date": "2016-07-07",
          "replacement": "jQuery 3.x or modern frameworks"
        }
      }
    },
    {
      "technology": "request",
      "versions": {
        "*": {
          "status": "deprecated",
          "deprecated_date": "2020-02-11",
          "replacement": "axios, node-fetch, got",
          "reason": "Package no longer maintained"
        }
      }
    }
  ]
}
```

#### Deprecation Report

```markdown
## Deprecated & End-of-Life Technologies

### Critical - End-of-Life with CVEs (1)

#### OpenSSL 1.1.1q
- **Status**: 🔴 **END-OF-LIFE** (since 2023-09-11)
- **Confidence**: 85%
- **Evidence**: /usr/lib/libssl.so.1.1
- **Impact**: System-wide TLS/SSL security risk
- **Known CVEs**:
  - CVE-2023-XXXX (CVSS 9.1) - Critical vulnerability
  - No security patches available
- **Replacement**: OpenSSL 3.0.x or 3.1.x LTS
- **Action**: Immediate system upgrade required
- **Timeline**: 0-7 days

### High - End-of-Life (2)

#### Node.js 14.21.3
- **Status**: 🟠 **END-OF-LIFE** (since 2023-04-30)
- **Confidence**: 95%
- **Evidence**: package.json engines field, .nvmrc
- **Impact**: No security updates, npm compatibility issues
- **Replacement**: Node.js 20.x LTS (active LTS)
- **Migration Complexity**: Medium - test for breaking changes
- **Timeline**: 30 days

#### MongoDB 4.2.24
- **Status**: 🟠 **END-OF-LIFE** (since 2023-04-01)
- **Confidence**: 90%
- **Evidence**: docker-compose.yml
- **Impact**: No updates, limited cloud support
- **Replacement**: MongoDB 6.0+ or 7.0 LTS
- **Migration Complexity**: High - schema/query testing required
- **Timeline**: 90 days

### Medium - Deprecated (3)

#### jQuery 2.1.4
- **Status**: 🟡 **DEPRECATED** (since 2016-07-07)
- **Confidence**: 92%
- **Evidence**: package.json, vendor/jquery-2.1.4.min.js
- **Usage**: Legacy admin panel only
- **Replacement**: jQuery 3.x or modern framework (Vue.js, React)
- **Security**: Known XSS vulnerabilities
- **Timeline**: 60-90 days

#### request 2.88.0 (npm)
- **Status**: 🟡 **DEPRECATED** (since 2020-02-11)
- **Confidence**: 95%
- **Evidence**: package.json
- **Reason**: Package no longer maintained
- **Replacement**: axios, node-fetch, or got
- **Usage**: 3 files in src/legacy/
- **Timeline**: 30 days

#### moment.js 2.29.4
- **Status**: 🟡 **MAINTENANCE MODE** (since 2020-09-15)
- **Confidence**: 94%
- **Evidence**: package.json
- **Reason**: Large bundle size (67KB), deprecated
- **Replacement**: date-fns or Temporal API
- **Usage**: 23 files
- **Timeline**: 60 days

### Summary

- **Critical**: 1 technology (immediate action)
- **High**: 2 technologies (30-90 days)
- **Medium**: 3 technologies (30-90 days)
- **Total Deprecated/EOL**: 6 technologies

**Overall Risk**: 🔴 High - Immediate action required for OpenSSL
```

### Policy Compliance Reporting

Generate compliance reports against organizational policies.

#### Example Compliance Report

```markdown
# Technology Policy Compliance Report

**Repository**: myorg/myapp
**Scan Date**: 2024-11-23
**Policy Version**: 2.1.0

## Executive Summary

**Compliance Score**: 72/100 (C Grade)

- ✅ **Approved**: 35 technologies (74%)
- ⚠️ **Review Required**: 3 technologies (6%)
- 🔴 **Banned**: 1 technology (2%)
- ⚠️ **Unapproved**: 8 technologies (17%)

**Policy Violations**: 4 critical, 3 high, 5 medium

## Critical Violations (4)

1. **Banned Technology: OpenSSL 1.1.1**
   - **Policy**: Banned since 2023-09-11
   - **Reason**: End-of-life, known vulnerabilities
   - **Action**: Upgrade to OpenSSL 3.x immediately

2. **Banned Technology: request (npm)**
   - **Policy**: Banned since 2020-02-11
   - **Reason**: Deprecated, no longer maintained
   - **Action**: Replace with axios or node-fetch

3. **Consolidation Violation: 3 HTTP clients**
   - **Policy**: Maximum 1 HTTP client per repository
   - **Detected**: axios, node-fetch, request
   - **Action**: Standardize on axios

4. **Unapproved Technology: Custom JWT library**
   - **Policy**: Only approved auth libraries allowed
   - **Action**: Replace with jsonwebtoken or Auth0

## Review Required (3)

1. **MongoDB 5.0**
   - **Policy**: Requires legal review (SSPL license)
   - **Action**: Obtain legal sign-off or use DocumentDB

2. **Salesforce API**
   - **Policy**: Director-level approval required
   - **Status**: Approval pending
   - **Action**: Follow up on approval request

## Recommendations

1. **Immediate** (0-7 days):
   - Upgrade OpenSSL to 3.x
   - Replace banned `request` package

2. **Short-term** (30 days):
   - Consolidate HTTP clients to axios
   - Obtain approvals for review-required technologies

3. **Medium-term** (60-90 days):
   - Replace deprecated libraries
   - Document technology choices
```

## RAG Library Structure

```
rag/
└── technology-identification/
    ├── README.md                          # Overview and usage
    │
    ├── business-tools/
    │   ├── crm/
    │   │   ├── salesforce/
    │   │   │   ├── api-patterns.md        # API endpoints, auth patterns
    │   │   │   ├── config-patterns.md     # Configuration files
    │   │   │   ├── import-patterns.md     # Import/require patterns
    │   │   │   ├── env-variables.md       # Environment variable patterns
    │   │   │   └── versions.md            # Version history & changes
    │   │   ├── hubspot/
    │   │   └── zoho/
    │   │
    │   ├── payment/
    │   │   ├── stripe/
    │   │   │   ├── api-patterns.md
    │   │   │   ├── sdk-patterns.md
    │   │   │   ├── webhook-patterns.md
    │   │   │   └── versions.md
    │   │   ├── paypal/
    │   │   ├── square/
    │   │   └── braintree/
    │   │
    │   ├── communication/
    │   │   ├── twilio/
    │   │   ├── sendgrid/
    │   │   └── slack/
    │   │
    │   └── analytics/
    │       ├── google-analytics/
    │       ├── mixpanel/
    │       └── segment/
    │
    ├── developer-tools/
    │   ├── infrastructure/
    │   │   ├── terraform/
    │   │   │   ├── config-patterns.md     # .tf file patterns
    │   │   │   ├── provider-patterns.md   # Provider blocks
    │   │   │   ├── module-patterns.md     # Module usage
    │   │   │   └── versions.md
    │   │   ├── ansible/
    │   │   └── cloudformation/
    │   │
    │   ├── containers/
    │   │   ├── docker/
    │   │   │   ├── dockerfile-patterns.md
    │   │   │   ├── compose-patterns.md
    │   │   │   └── versions.md
    │   │   └── kubernetes/
    │   │       ├── manifest-patterns.md
    │   │       ├── helm-patterns.md
    │   │       └── versions.md
    │   │
    │   ├── cicd/
    │   │   ├── github-actions/
    │   │   │   ├── workflow-patterns.md
    │   │   │   ├── action-patterns.md
    │   │   │   └── versions.md
    │   │   ├── gitlab-ci/
    │   │   └── jenkins/
    │   │
    │   └── monitoring/
    │       ├── prometheus/
    │       ├── grafana/
    │       └── datadog/
    │
    ├── programming-languages/
    │   ├── python/
    │   │   ├── version-detection.md       # Version identification
    │   │   ├── stdlib-patterns.md         # Standard library usage
    │   │   ├── framework-patterns.md      # Django, Flask, FastAPI
    │   │   └── versions.md
    │   │
    │   ├── javascript/
    │   │   ├── runtime-detection.md       # Node.js, Deno, Bun
    │   │   ├── typescript-patterns.md     # TypeScript detection
    │   │   ├── framework-patterns.md      # React, Vue, Angular
    │   │   └── versions.md
    │   │
    │   ├── go/
    │   ├── rust/
    │   └── java/
    │
    ├── cryptographic-libraries/
    │   ├── tls/
    │   │   ├── openssl/
    │   │   │   ├── import-patterns.md
    │   │   │   ├── api-patterns.md
    │   │   │   ├── vulnerabilities.md     # Known CVEs per version
    │   │   │   └── versions.md
    │   │   ├── libressl/
    │   │   └── boringssl/
    │   │
    │   ├── crypto/
    │   │   ├── libsodium/
    │   │   ├── crypto-js/
    │   │   └── pycryptodome/
    │   │
    │   └── jwt/
    │       ├── jsonwebtoken/
    │       └── pyjwt/
    │
    ├── databases/
    │   ├── relational/
    │   │   ├── postgresql/
    │   │   │   ├── driver-patterns.md     # psycopg2, node-postgres
    │   │   │   ├── connection-patterns.md
    │   │   │   └── versions.md
    │   │   └── mysql/
    │   │
    │   ├── nosql/
    │   │   ├── mongodb/
    │   │   └── cassandra/
    │   │
    │   └── keyvalue/
    │       ├── redis/
    │       └── memcached/
    │
    └── cloud-providers/
        ├── aws/
        │   ├── sdk-patterns.md            # AWS SDK usage
        │   ├── service-patterns.md        # S3, Lambda, etc.
        │   ├── endpoint-patterns.md       # API endpoints
        │   ├── iam-patterns.md            # IAM usage
        │   └── versions.md
        │
        ├── gcp/
        │   ├── sdk-patterns.md
        │   ├── service-patterns.md
        │   └── versions.md
        │
        └── azure/
            ├── sdk-patterns.md
            ├── service-patterns.md
            └── versions.md
```

## Pattern Documentation Format

Each technology's pattern files follow a standard format:

### Example: `stripe/api-patterns.md`

```markdown
# Stripe API Patterns

**Category**: Business Tools → Payment Processing
**Confidence Base**: 80% (API endpoint detection)

## API Endpoints

### Production
- `https://api.stripe.com/v1/*`
- Confidence: 85%

### Test Mode
- `https://api.stripe.com/test/*`
- Confidence: 70% (may be example/test code)

## SDK Patterns

### JavaScript/Node.js
```javascript
// High confidence patterns
const stripe = require('stripe')('sk_...');
import Stripe from 'stripe';
const stripe = new Stripe('sk_...');

// Medium confidence (import without instantiation)
import Stripe from 'stripe';
```
Confidence:
- With API key: 95%
- Import only: 75%

### Python
```python
import stripe
stripe.api_key = "sk_..."
stripe.Charge.create(...)
```
Confidence: 85%

### Ruby
```ruby
require 'stripe'
Stripe.api_key = 'sk_...'
```
Confidence: 85%

## Environment Variables

- `STRIPE_API_KEY` - Confidence: 70%
- `STRIPE_SECRET_KEY` - Confidence: 70%
- `STRIPE_PUBLISHABLE_KEY` - Confidence: 65%

## Package Dependencies

### npm
- `stripe` - Confidence: 95%

### Python
- `stripe` - Confidence: 95%

### Ruby
- `stripe` (gem) - Confidence: 95%

## Webhook Patterns

```javascript
// Webhook endpoint signature verification
stripe.webhooks.constructEvent(body, signature, secret)
```
Confidence: 90% (strong indicator of Stripe integration)

## Configuration Files

```yaml
# Example: serverless.yml or similar
environment:
  STRIPE_SECRET_KEY: ${env:STRIPE_SECRET_KEY}
```
Confidence: 75%

## Comments & Documentation Patterns

- "Stripe payment", "stripe integration" - Confidence: 45%
- "Process payment with Stripe" - Confidence: 50%

## Version Detection

Extract from:
1. `package.json`: `"stripe": "^10.0.0"` → Version 10.x
2. Import: `stripe.VERSION` or API version header
3. API endpoint: `/v1/` → API version 1

## Detection Rules

1. **Definitive** (95-100%):
   - Package dependency declared
   - SDK import with instantiation
   - Webhook signature verification

2. **High Confidence** (80-94%):
   - API endpoint in code with authentication
   - Environment variables + import statement
   - Webhook endpoint patterns

3. **Medium Confidence** (60-79%):
   - Import without usage
   - Environment variables only
   - API endpoint without auth

4. **Low Confidence** (0-59%):
   - Documentation mentions
   - Comments referencing Stripe
```

## RAG Update Mechanism

### Automated Update System

```
utils/technology-identification/
└── rag-updater/
    ├── update-rag.sh                    # Main update script
    ├── sources/
    │   ├── npm-registry-scraper.sh      # Scrape npm for new versions
    │   ├── pypi-scraper.sh              # Scrape PyPI
    │   ├── github-scraper.sh            # Check GitHub releases
    │   └── api-documentation-fetcher.sh # Fetch API docs
    │
    ├── parsers/
    │   ├── parse-npm-package.sh         # Extract patterns from npm
    │   ├── parse-github-readme.sh       # Extract from README
    │   └── parse-api-docs.sh            # Parse OpenAPI/Swagger
    │
    └── generators/
        ├── generate-patterns.sh          # Generate pattern files
        └── generate-versions.sh          # Update version info
```

### Update Workflow

```bash
# Weekly automated update
./rag-updater/update-rag.sh --auto

# Manual update for specific technology
./rag-updater/update-rag.sh --tech stripe --source npm

# Update from API documentation
./rag-updater/update-rag.sh --tech aws --source api-docs
```

### Version Tracking Format

Each technology has a `versions.md`:

```markdown
# Stripe Versions

## Current Stable
- **Latest**: 14.12.0 (2024-11-20)
- **LTS**: 12.18.0 (2024-06-15)

## Version History

### v14.x (2024-09-01 - Current)
**Major Changes**:
- New Payment Intents API
- Removed legacy Charges API
**Pattern Changes**:
- `stripe.charges.*` → Deprecated
- `stripe.paymentIntents.*` → New pattern

### v13.x (2023-12-01 - 2024-08-31)
**Major Changes**:
- Enhanced webhook handling
**Pattern Changes**:
- Added `stripe.webhooks.constructEvent()`

## Deprecated Versions
- v10.x and earlier (End of support: 2023-12-31)

## Breaking Changes

### v13 → v14
- `stripe.charges.create()` deprecated
- Migrate to `stripe.paymentIntents.create()`

## Security Advisories
- CVE-2023-XXXX: Affects v10.x - v12.5.0
  - Impact: Webhook signature bypass
  - Fixed in: v12.5.1+
```

## Analyzer Implementation

### High-Level Architecture

```
technology-identification-analyser.sh
├── parse_arguments()
├── detect_technologies()
│   ├── scan_manifests()          # Layer 1
│   ├── scan_config_files()       # Layer 2
│   ├── scan_imports()            # Layer 3
│   ├── scan_api_endpoints()      # Layer 4
│   ├── scan_env_variables()      # Layer 5
│   └── scan_documentation()      # Layer 6
├── load_rag_patterns()           # Load detection patterns from RAG
├── calculate_confidence()        # Score each detection
├── aggregate_findings()          # Composite confidence scores
├── generate_report()             # Output findings
└── run_claude_analysis()         # Optional AI enhancement
```

### Detection Algorithm

```bash
detect_technology() {
    local tech_name="$1"
    local rag_dir="$RAG_BASE/$tech_category/$tech_name"

    # Load patterns from RAG
    local api_patterns=$(cat "$rag_dir/api-patterns.md")
    local import_patterns=$(cat "$rag_dir/import-patterns.md")
    local env_patterns=$(cat "$rag_dir/env-variables.md")

    # Multi-layer detection
    local findings=()

    # Layer 1: Package manifest
    if grep -q "$tech_name" package.json 2>/dev/null; then
        findings+=("manifest:95")
    fi

    # Layer 2: Configuration files
    for config_file in $(find . -name "*.config.js" -o -name "*.yml"); do
        if grep -q "$tech_name" "$config_file"; then
            findings+=("config:90")
        fi
    done

    # Layer 3: Import statements
    for src_file in $(find . -name "*.js" -o -name "*.py" -o -name "*.go"); do
        if grep -E "(import|require|from).*$tech_name" "$src_file"; then
            findings+=("import:75")
        fi
    done

    # Layer 4: API endpoints
    for src_file in $(find . -type f -name "*.js" -o -name "*.py"); do
        for endpoint in $(echo "$api_patterns" | grep -oP 'https://[^"]+'); do
            if grep -q "$endpoint" "$src_file"; then
                findings+=("endpoint:70")
            fi
        done
    done

    # Calculate composite confidence
    calculate_confidence "${findings[@]}"
}
```

## Claude AI Integration

### Analysis Prompts

**Prompt 1: Pattern Extraction**
```
Given the following technology documentation for {TECHNOLOGY_NAME}, extract:
1. API endpoint patterns
2. SDK import patterns
3. Configuration file patterns
4. Environment variable naming conventions
5. Version detection methods

Format the output as markdown pattern files following our RAG structure.
```

**Prompt 2: Technology Identification**
```
Analyze the following repository scan results and SBOM to identify technologies:

SBOM: {SBOM_JSON}
Config Files: {CONFIG_FILES}
Import Statements: {IMPORTS}
API Endpoints: {ENDPOINTS}

For each identified technology, provide:
1. Technology name and category
2. Version (if detectable)
3. Evidence list with confidence scores
4. Security considerations
5. Compliance implications
```

**Prompt 3: Report Generation**
```
Generate a comprehensive technology stack report:

Findings: {TECHNOLOGY_FINDINGS_JSON}

Provide:
1. Executive summary
2. Technology breakdown by category
3. Risk assessment (deprecated versions, security issues)
4. Compliance implications
5. Recommendations for technology rationalization
```

## Output Formats

### JSON Output
```json
{
  "repository": "owner/repo",
  "scan_date": "2024-11-23T10:00:00Z",
  "technologies": [
    {
      "name": "Stripe",
      "category": "business-tools/payment",
      "version": "14.12.0",
      "confidence": 94,
      "evidence": [
        {
          "type": "manifest",
          "location": "package.json",
          "confidence": 95
        },
        {
          "type": "import",
          "location": "src/payment.js:3",
          "confidence": 85,
          "snippet": "import Stripe from 'stripe';"
        },
        {
          "type": "api_endpoint",
          "location": "src/payment.js:15",
          "confidence": 80,
          "snippet": "https://api.stripe.com/v1/charges"
        }
      ],
      "risk_level": "low",
      "notes": "Current stable version, no known vulnerabilities"
    },
    {
      "name": "OpenSSL",
      "category": "cryptographic-libraries/tls",
      "version": "1.1.1",
      "confidence": 85,
      "evidence": [
        {
          "type": "binary",
          "location": "/usr/lib/libssl.so.1.1",
          "confidence": 90
        }
      ],
      "risk_level": "critical",
      "notes": "OpenSSL 1.1.1 reached end-of-life 2023-09-11. Upgrade to 3.x required."
    }
  ],
  "summary": {
    "total_technologies": 47,
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

### Markdown Report
```markdown
# Technology Stack Report

**Repository**: owner/repo
**Scan Date**: 2024-11-23
**Total Technologies**: 47

## Executive Summary

This repository uses 47 identified technologies across 6 categories.
**Critical Risk**: 1 technology (OpenSSL 1.1.1) requires immediate attention due to end-of-life status.

## Technology Breakdown

### Business Tools (5)

#### Payment Processing
- **Stripe v14.12.0** (Confidence: 94%)
  - Evidence: package.json, src/payment.js imports
  - Status: ✅ Current version
  - Risk: Low

#### Communication
- **Twilio v4.5.0** (Confidence: 88%)
  - Evidence: requirements.txt, src/notifications.py
  - Status: ✅ Current version
  - Risk: Low

### Cryptographic Libraries (2)

#### TLS/SSL
- **OpenSSL 1.1.1** (Confidence: 85%)
  - Evidence: /usr/lib/libssl.so.1.1
  - Status: ⚠️ End-of-Life (2023-09-11)
  - Risk: **CRITICAL**
  - Recommendation: Upgrade to OpenSSL 3.x immediately

### Developer Tools (12)
...

## Risk Assessment

### Critical (1)
1. OpenSSL 1.1.1 - End-of-life, multiple CVEs, upgrade required

### High (3)
1. Node.js 14.x - Approaching end-of-life (2024-04-30)
2. MongoDB 4.2 - Deprecated version, upgrade to 6.x
3. jQuery 2.x - Unsupported, migrate to 3.x

## Recommendations

1. **Immediate**: Upgrade OpenSSL to 3.x
2. **Short-term** (30 days): Update Node.js to 20.x LTS
3. **Medium-term** (90 days): Migrate MongoDB to 6.x, remove jQuery
```

## Security & Compliance Considerations

### Secret Detection Integration
- Cross-reference identified technologies with secret detection
- Flag hardcoded API keys for identified services (Stripe, AWS, etc.)
- Higher severity for business-critical tools

### License Compliance
- Technology identification feeds into license analysis
- Some business tools (Salesforce) have restrictive licensing
- Open source cryptographic libraries may have export restrictions

### Vulnerability Correlation
- Link identified technology versions to known CVEs
- Prioritize vulnerabilities in actively used technologies
- Highlight end-of-life technologies

## Next Steps

1. ✅ Design RAG library structure
2. Create initial RAG content for high-value technologies:
   - Stripe (payment)
   - AWS SDK (cloud)
   - OpenSSL (crypto)
   - Terraform (IaC)
   - Docker (containers)
3. Implement analyzer script
4. Build RAG update mechanism
5. Integrate with existing supply chain scanner
6. Add Claude AI analysis prompts
7. Test on real repositories
8. Document and deploy

---

**Document Version**: 0.1.0
**Last Updated**: 2025-11-23
**Status**: Design Phase - Ready for Implementation
