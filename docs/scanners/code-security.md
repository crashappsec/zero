# Code Security Scanner

The Code Security scanner provides security-focused static code analysis (SAST), detecting vulnerabilities, exposed secrets, and API security issues.

## Overview

| Property | Value |
|----------|-------|
| **Name** | `code-security` |
| **Version** | 3.2.0 |
| **Output File** | `code-security.json` |
| **Dependencies** | None |
| **Estimated Time** | 60-180 seconds |

## Features

### 1. Vulnerabilities (`vulns`)

Static Application Security Testing (SAST) using Semgrep.

**Configuration:**
```json
{
  "vulns": {
    "enabled": true,
    "rulesets": ["p/security-audit", "p/owasp-top-ten"],
    "severity_minimum": "low"
  }
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable vulnerability scanning |
| `rulesets` | []string | `["p/security-audit", "p/owasp-top-ten"]` | Semgrep rule packs |
| `severity_minimum` | string | `"low"` | Minimum severity to report |

**Severity Mapping (Semgrep to Zero):**

| Semgrep | Zero |
|---------|------|
| ERROR | critical |
| WARNING | high |
| INFO | medium |
| (other) | low |

**Detected Vulnerability Categories:**
- Injection flaws (SQL, NoSQL, Command, LDAP, XPath)
- Cross-site scripting (XSS)
- Authentication issues
- Authorization bypass
- Path traversal
- Server-Side Request Forgery (SSRF)
- Insecure deserialization
- Security misconfiguration
- OWASP Top 10 coverage

**CWE and OWASP Enrichment:**
- Findings are automatically enriched with CWE IDs from Semgrep metadata
- OWASP mapping is extracted when available

### 2. Secrets (`secrets`)

RAG-based secret and credential detection with optional API validation to reduce false positives.

**Configuration:**
```json
{
  "secrets": {
    "enabled": true,
    "detection_method": "rag",
    "rag_patterns_path": "rag/technology-identification",
    "validate_with_api": true,
    "redact_secrets": true,
    "min_confidence": 0.8
  }
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable secret detection |
| `detection_method` | string | `"rag"` | Detection method (`rag` for RAG-based patterns) |
| `rag_patterns_path` | string | `"rag/technology-identification"` | Path to RAG patterns |
| `validate_with_api` | bool | `true` | Validate secrets via API to reduce false positives |
| `redact_secrets` | bool | `true` | Mask secret values in output |
| `min_confidence` | float | `0.8` | Minimum confidence for detection (0-1) |

**How RAG-Based Secrets Detection Works:**

1. **Technology Detection**: First identifies technologies in the codebase (e.g., OpenAI, AWS, Stripe)
2. **Pattern Matching**: Uses technology-specific secret patterns from RAG database
3. **API Validation**: Optionally validates secrets against provider APIs to confirm they're active
4. **Confidence Scoring**: Assigns confidence based on pattern specificity and context

**Detected Secret Types (from RAG patterns):**

| Technology | Pattern | Severity | Validation |
|------------|---------|----------|------------|
| OpenAI | `sk-[A-Za-z0-9]{48}` | Critical | `/v1/models` endpoint |
| OpenAI (proj) | `sk-proj-[A-Za-z0-9_-]{100,}` | Critical | `/v1/models` endpoint |
| Anthropic | `sk-ant-[A-Za-z0-9_-]{90,}` | Critical | Messages API |
| AWS | `AKIA[0-9A-Z]{16}` | Critical | STS GetCallerIdentity |
| GitHub | `ghp_[A-Za-z0-9]{36}` | High | User API |
| Stripe | `sk_live_[A-Za-z0-9]{24,}` | Critical | Account API |
| Stripe (test) | `sk_test_[A-Za-z0-9]{24,}` | Medium | Account API |
| Slack | `xoxb-[A-Za-z0-9-]+` | High | Auth test API |
| SendGrid | `SG\.[A-Za-z0-9_-]{22}\.[A-Za-z0-9_-]{43}` | High | API validation |
| Twilio | `SK[a-f0-9]{32}` | High | Account validation |
| Datadog | `[a-f0-9]{32}` (with context) | High | Validate endpoint |
| HuggingFace | `hf_[A-Za-z0-9]{34}` | High | WhoAmI endpoint |

**API Validation Benefits:**

- **Reduces false positives**: Only reports secrets that are actually valid
- **Identifies active vs revoked**: Distinguishes between active and inactive credentials
- **Low cost**: Uses free/minimal-cost validation endpoints
- **Optional**: Can be disabled for air-gapped environments

**Redaction:**
When `redact_secrets` is enabled, secrets longer than 16 characters are partially masked:
```
Before: sk-abcdef123456789012345678901234567890123456789012
After:  sk-********
```

**Confidence Levels:**

| Confidence | Description |
|------------|-------------|
| 95-100% | Very distinctive pattern (e.g., `sk-ant-`, `AKIA`) |
| 85-94% | Strong pattern with context |
| 70-84% | Moderate pattern, may need validation |
| <70% | Weak pattern, high false positive risk |

**Risk Score Calculation:**
```
Score = 100 - (critical × 25 + high × 15 + medium × 5 + low × 2)
```

| Score | Risk Level |
|-------|------------|
| 0-39 | critical |
| 40-59 | high |
| 60-79 | medium |
| 80-94 | low |
| 95-100 | excellent |

### 3. API Security (`api`)

API-specific security analysis targeting OWASP API Top 10.

**Configuration:**
```json
{
  "api": {
    "enabled": true,
    "check_openapi": true,
    "check_graphql": true,
    "check_owasp_api": true
  }
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable API security analysis |
| `check_openapi` | bool | `true` | Check OpenAPI/Swagger patterns |
| `check_graphql` | bool | `true` | Check GraphQL patterns |
| `check_owasp_api` | bool | `true` | Check OWASP API Top 10 |

**OWASP API Top 10 Mapping:**

| Category | Patterns Detected |
|----------|-------------------|
| API1 - BOLA | Object-level authorization issues |
| API2 - Broken Authentication | Auth, JWT, token, session issues |
| API3 - Broken Property Auth | Property-level access control |
| API4 - Resource Consumption | Rate limiting, DoS concerns |
| API5 - Broken Function Auth | Admin privilege escalation |
| API6 - Mass Assignment | Mass assignment, binding issues |
| API7 - SSRF | Server-side request forgery |
| API8 - Security Misconfiguration | CORS, headers, configs |
| API9 - Improper Inventory | Endpoint, version management |
| API10 - Unsafe Consumption | External/third-party risks |

**Finding Categories:**
- `authentication` - Auth-related issues
- `authorization` - Access control issues
- `injection` - SQL, NoSQL, command injection
- `data-exposure` - Sensitive data exposure
- `rate-limiting` - DoS/brute force concerns
- `ssrf` - Request forgery
- `mass-assignment` - Over-posting risks
- `misconfiguration` - CORS, headers, TLS

## How It Works

### Technical Flow

1. **Tool Check**: Verifies Semgrep is installed (for vulns/api features)
2. **Parallel Execution**: Runs vulns, secrets, and api features concurrently
3. **Vulnerability Scan**: Semgrep runs with security rulesets
4. **Secrets Scan**: RAG-based pattern matching with optional API validation
5. **API Scan**: Filters findings for API-relevant issues
6. **Enrichment**: Adds CWE, OWASP mappings, severity classifications
7. **Aggregation**: Combines results into single output

### Semgrep Configuration (for vulns/api)

The scanner uses the following Semgrep options:
```bash
semgrep --json --metrics=off --timeout 60 --max-memory 4096 \
  --exclude node_modules --exclude vendor --exclude .git \
  --exclude dist --exclude build --exclude "*.min.js" \
  --config <ruleset> <repo_path>
```

### RAG Secrets Detection Flow

1. **Load RAG Patterns**: Read secret patterns from `rag/technology-identification/*/patterns.md`
2. **Scan Files**: Search for patterns with technology-specific context
3. **Score Matches**: Assign confidence based on pattern specificity
4. **Validate (Optional)**: Call provider APIs to confirm validity
5. **Redact Output**: Mask secrets before writing to output

File exclusions for secrets:
```
node_modules/, vendor/, .git/, dist/, build/
package-lock.json, yarn.lock, pnpm-lock.yaml
*.env.example, *.env.sample, *.env.template
```

## Usage

### Command Line

```bash
# Run code-security scanner only
./zero scan --scanner code-security /path/to/repo

# Run code-security profile
./zero hydrate owner/repo --profile code-security-only
```

### Programmatic Usage

```go
import codesecurity "github.com/crashappsec/zero/pkg/scanners/code-security"

opts := &scanner.ScanOptions{
    RepoPath:  "/path/to/repo",
    OutputDir: "/path/to/output",
    FeatureConfig: map[string]interface{}{
        "vulns": map[string]interface{}{
            "enabled": true,
            "severity_minimum": "medium",
        },
        "secrets": map[string]interface{}{
            "enabled": true,
            "redact_secrets": true,
        },
        "api": map[string]interface{}{
            "enabled": true,
        },
    },
}

scanner := &codesecurity.CodeSecurityScanner{}
result, err := scanner.Run(ctx, opts)
```

## Output Format

```json
{
  "scanner": "code-security",
  "version": "3.2.0",
  "metadata": {
    "features_run": ["vulns", "secrets", "api"]
  },
  "summary": {
    "vulns": {
      "total_findings": 15,
      "critical": 2,
      "high": 5,
      "medium": 6,
      "low": 2,
      "by_cwe": {
        "CWE-79": 3,
        "CWE-89": 2,
        "CWE-22": 1
      },
      "by_category": {
        "injection": 3,
        "xss": 3,
        "path-traversal": 1
      }
    },
    "secrets": {
      "total_findings": 5,
      "critical": 1,
      "high": 3,
      "medium": 1,
      "low": 0,
      "files_affected": 3,
      "by_type": {
        "aws_credential": 1,
        "github_token": 2,
        "api_key": 2
      },
      "risk_score": 45,
      "risk_level": "high"
    },
    "api": {
      "total_findings": 8,
      "critical": 1,
      "high": 3,
      "medium": 3,
      "low": 1,
      "by_category": {
        "authentication": 2,
        "injection": 3,
        "rate-limiting": 2,
        "misconfiguration": 1
      }
    },
    "errors": []
  },
  "findings": {
    "vulns": [
      {
        "rule_id": "python.flask.security.injection.sql-injection-flask",
        "title": "sql injection flask",
        "description": "Detected SQL injection in Flask application",
        "severity": "critical",
        "file": "app/routes.py",
        "line": 42,
        "column": 15,
        "category": "injection",
        "cwe": ["CWE-89"],
        "owasp": ["A03:2021"]
      }
    ],
    "secrets": [
      {
        "rule_id": "generic.secrets.security.detected-aws-access-key",
        "type": "aws_credential",
        "severity": "critical",
        "message": "AWS Access Key detected",
        "file": "config/settings.py",
        "line": 15,
        "column": 10,
        "snippet": "AWS_ACCESS_K********"
      }
    ],
    "api": [
      {
        "rule_id": "javascript.express.security.audit.express-no-rate-limit",
        "title": "express no rate limit",
        "description": "Express route without rate limiting",
        "severity": "medium",
        "file": "src/routes/api.js",
        "line": 28,
        "category": "rate-limiting",
        "owasp_api": "API4 - Resource Consumption"
      }
    ]
  }
}
```

## Prerequisites

| Tool | Required | Install Command |
|------|----------|-----------------|
| semgrep | Yes | `pip install semgrep` or `brew install semgrep` |

**Note:** The scanner will report errors if Semgrep is not installed but will not fail completely.

## Profiles

| Profile | vulns | secrets | api |
|---------|-------|---------|-----|
| `quick` | - | - | - |
| `standard` | Yes | Yes | Yes |
| `security` | Yes | Yes | Yes |
| `full` | Yes | Yes | Yes |
| `code-security-only` | Yes | Yes | Yes |

## Related Scanners

- **crypto**: Overlaps on cryptographic key detection
- **quality**: Complements with code quality analysis
- **packages**: Checks for vulnerable dependencies

## See Also

- [Quality Scanner](quality.md) - Code quality analysis
- [Crypto Scanner](crypto.md) - Cryptographic security analysis
- [Technology Scanner](technology.md) - Technology detection and secrets patterns source
- [RAG Technology Patterns](../../rag/technology-identification/README.md) - Secret pattern definitions
- [Semgrep Rules](https://semgrep.dev/explore) - Available rule packs (for vulns/api)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/) - Web vulnerability reference
- [OWASP API Top 10](https://owasp.org/API-Security/) - API security reference
