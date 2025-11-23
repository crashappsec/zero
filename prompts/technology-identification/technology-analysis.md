# Technology Analysis Prompt

**Purpose**: Analyze repository scan results, SBOMs, and source code to identify technologies with evidence and confidence scoring.

## Prompt

```
You are a technology stack analyst expert. Analyze the following repository data to comprehensively identify all technologies, tools, frameworks, and services in use.

## Input Data

### Repository Information
**Repository**: {REPO_NAME}
**Branch**: {BRANCH}
**Scan Date**: {SCAN_DATE}
**Languages Detected**: {LANGUAGES}

### SBOM (Software Bill of Materials)
```json
{SBOM_JSON}
```

### Package Manager Files
{PACKAGE_FILES_CONTENT}

### Configuration Files
{CONFIG_FILES_CONTENT}

### Source Code Samples (Imports & References)
{IMPORT_STATEMENTS}

### Environment Variables
{ENV_VAR_PATTERNS}

### API Endpoints Detected
{API_ENDPOINTS}

## Analysis Framework

Use the **6-layer detection approach** with confidence scoring:

### Layer 1: Manifest & Lock Files (90-100% confidence)
- Parse package.json, requirements.txt, Cargo.toml, go.mod, pom.xml, etc.
- Extract declared dependencies with exact versions
- Identify package ecosystems

### Layer 2: Configuration Files (80-95% confidence)
- Detect Dockerfile, docker-compose.yml, terraform.tf, .github/workflows
- Identify infrastructure and deployment tools
- Extract build tool configurations

### Layer 3: Import Statements (60-80% confidence)
- Parse source code imports across all languages
- Identify SDK usage patterns
- Cross-reference with package dependencies

### Layer 4: API Endpoints (60-80% confidence)
- Match endpoint patterns against RAG database
- Identify cloud provider services (AWS S3, GCP Storage, etc.)
- Detect SaaS integrations (Stripe, Twilio, SendGrid)

### Layer 5: Environment Variables (40-60% confidence)
- Match environment variable naming patterns
- Infer technologies from variable names
- Cross-reference with other evidence

### Layer 6: Comments & Documentation (30-50% confidence)
- Analyze documentation mentions
- Low confidence - use as supporting evidence only

## Output Format

Generate a comprehensive JSON report:

```json
{
  "repository": "{REPO_NAME}",
  "scan_date": "{ISO_DATE}",
  "total_technologies": 0,
  "technologies": [
    {
      "name": "Technology Name",
      "category": "category/subcategory",
      "version": "1.2.3",
      "confidence": 94,
      "evidence": [
        {
          "type": "manifest",
          "method": "layer1",
          "confidence": 95,
          "location": "package.json:12",
          "snippet": "\"technology\": \"^1.2.3\""
        },
        {
          "type": "import",
          "method": "layer3",
          "confidence": 85,
          "location": "src/main.js:3",
          "snippet": "import Technology from 'technology';"
        },
        {
          "type": "api_endpoint",
          "method": "layer4",
          "confidence": 80,
          "location": "src/api.js:45",
          "snippet": "https://api.technology.com/v1/"
        },
        {
          "type": "env_var",
          "method": "layer5",
          "confidence": 65,
          "location": ".env.example:8",
          "snippet": "TECHNOLOGY_API_KEY="
        }
      ],
      "risk_level": "low|medium|high|critical",
      "risk_factors": [
        "End-of-life version",
        "Known CVEs",
        "Deprecated"
      ],
      "notes": "Additional context and observations"
    }
  ],
  "summary": {
    "by_category": {
      "business-tools": 5,
      "developer-tools": 12,
      "programming-languages": 3,
      "cryptographic-libraries": 2,
      "databases": 4,
      "cloud-providers": 3,
      "message-queues": 2,
      "web-frameworks": 8
    },
    "risk_summary": {
      "critical": 1,
      "high": 3,
      "medium": 8,
      "low": 35
    },
    "version_risks": {
      "eol_technologies": ["OpenSSL 1.1.1", "Node.js 14.x"],
      "deprecated": ["jQuery 2.x"],
      "outdated": ["Express 4.16.x (current: 4.18.x)"]
    }
  },
  "recommendations": [
    {
      "priority": "critical",
      "technology": "OpenSSL 1.1.1",
      "issue": "End-of-life (EOL: 2023-09-11) with known CVEs",
      "action": "Upgrade to OpenSSL 3.x",
      "timeline": "Immediate (0-7 days)"
    }
  ]
}
```

## Analysis Requirements

### Technology Identification
For each identified technology, provide:
1. **Name**: Official technology name
2. **Category**: One of 8 major categories
3. **Version**: Exact version or version range
4. **Confidence**: 0-100% based on evidence strength
5. **Evidence**: List of all detection points with snippets

### Confidence Scoring
Apply composite scoring when multiple evidence types exist:
```
Composite = (Evidence1 + Evidence2 + ... + EvidenceN) / N × 1.2
(Capped at 100%)
```

### Risk Assessment
Classify each technology:
- **🔴 Critical**: EOL with CVEs, AGPL in proprietary, export-controlled crypto
- **🟠 High**: Deprecated, approaching EOL, major versions behind
- **🟡 Medium**: Minor versions behind, security advisories (non-critical)
- **🟢 Low**: Current stable/LTS, actively maintained, no known issues

### Version Analysis
For each technology:
1. Identify exact version from lock files (preferred) or manifests
2. Check EOL/deprecation status
3. Compare against latest stable version
4. Note breaking changes in upgrade path
5. Identify any known vulnerabilities (CVE database)

### Compliance Implications
Flag technologies with compliance concerns:
- **Export Control**: Strong cryptography (OpenSSL, BoringSSL, libsodium)
- **License Risk**: AGPL/GPL in proprietary software
- **Data Privacy**: Analytics/CRM tools handling PII (GDPR/CCPA)
- **Financial**: PCI DSS compliance for payment processors

### Security Considerations
For each technology, note:
- Known CVEs in the detected version
- CISA KEV (Known Exploited Vulnerabilities) status
- Security advisories
- Recommended security updates
- Attack surface implications

## Categories

Classify technologies into these categories:

1. **business-tools** (CRM, payment, communication, analytics, support, marketing)
2. **developer-tools** (IaC, containers, orchestration, CI/CD, build, testing, monitoring)
3. **programming-languages** (languages, runtimes, compilers)
4. **cryptographic-libraries** (TLS/SSL, crypto primitives, hashing, JWT, signing)
5. **web-frameworks** (frontend, backend, API, authentication)
6. **databases** (relational, NoSQL, key-value, search, time-series)
7. **cloud-providers** (AWS, GCP, Azure, services)
8. **message-queues** (queues, streaming, event buses)

## Special Detections

### Cloud Provider Services
When AWS SDK is detected, identify specific services:
- S3 (storage): `@aws-sdk/client-s3`, `s3.amazonaws.com`
- Lambda (compute): `@aws-sdk/client-lambda`, `lambda.*.amazonaws.com`
- DynamoDB (database): `@aws-sdk/client-dynamodb`
- SQS (queue): `@aws-sdk/client-sqs`
- SNS (notifications): `@aws-sdk/client-sns`

### Framework Detection
Identify web frameworks and their versions:
- **React**: Check `react` + `react-dom` versions, detect Next.js/Gatsby
- **Vue**: Check `vue` version, detect Nuxt.js
- **Angular**: Check `@angular/core` version
- **Django**: Check `Django` in requirements.txt
- **Rails**: Check `rails` gem version

### Database Drivers
From database drivers, infer databases:
- `pg`, `node-postgres` → PostgreSQL
- `mysql2` → MySQL
- `mongodb`, `mongoose` → MongoDB
- `ioredis`, `redis` → Redis

### Build Tools
Detect build tooling:
- `webpack.config.js` → Webpack
- `vite.config.js` → Vite
- `rollup.config.js` → Rollup
- `esbuild` in package.json → esbuild

## Quality Standards

- **Accuracy**: Only report technologies with verifiable evidence
- **Completeness**: Identify technologies across all categories
- **Confidence**: Apply rigorous confidence scoring
- **Evidence**: Preserve file paths, line numbers, code snippets
- **Risk**: Assess security and compliance implications
- **Actionable**: Provide clear, prioritized recommendations

## Example Analysis

For a Node.js application with:
- `package.json`: stripe@14.12.0, express@4.18.2
- `import Stripe from 'stripe'` in source
- `STRIPE_SECRET_KEY` in .env.example
- `https://api.stripe.com/v1/charges` in code

Output:
```json
{
  "name": "Stripe",
  "category": "business-tools/payment",
  "version": "14.12.0",
  "confidence": 94,
  "evidence": [
    {"type": "manifest", "confidence": 95, "location": "package.json:12"},
    {"type": "import", "confidence": 85, "location": "src/payment.js:3"},
    {"type": "api_endpoint", "confidence": 80, "location": "src/payment.js:45"},
    {"type": "env_var", "confidence": 65, "location": ".env.example:8"}
  ],
  "risk_level": "low",
  "notes": "Current stable version, no known vulnerabilities"
}
```

Composite Confidence: (95 + 85 + 80 + 65) / 4 × 1.2 = 97.5% → 97%

Provide comprehensive, accurate, evidence-based technology identification.
```

## Usage

```bash
# Run technology analysis
./technology-identification-analyser.sh \
  --repo owner/repo \
  --output technology-report.json

# Run with Claude AI analysis
export ANTHROPIC_API_KEY="sk-..."
./technology-identification-analyser.sh \
  --claude \
  --repo owner/repo \
  --output technology-report.json
```
