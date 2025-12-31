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

Multi-source secret and credential detection combining Semgrep rules, entropy analysis, and git history scanning. Includes optional AI-powered false positive reduction and rotation recommendations.

**Configuration:**
```json
{
  "secrets": {
    "enabled": true,
    "redact_secrets": true,
    "entropy_analysis": {
      "enabled": true,
      "min_length": 16,
      "high_threshold": 4.5,
      "med_threshold": 3.5
    },
    "git_history_scan": {
      "enabled": false,
      "max_commits": 1000,
      "max_age": "1y",
      "scan_removed": true
    },
    "ai_analysis": {
      "enabled": false,
      "max_findings": 50,
      "confidence_threshold": 0.8
    },
    "rotation_guidance": true
  }
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable secret detection |
| `redact_secrets` | bool | `true` | Mask secret values in output |

**Entropy Analysis (enabled by default):**

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `entropy_analysis.enabled` | bool | `true` | Enable Shannon entropy detection |
| `entropy_analysis.min_length` | int | `16` | Minimum string length to analyze |
| `entropy_analysis.high_threshold` | float | `4.5` | Entropy threshold for high confidence |
| `entropy_analysis.med_threshold` | float | `3.5` | Entropy threshold for medium confidence |

**Git History Scanning (disabled by default):**

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `git_history_scan.enabled` | bool | `false` | Enable git history scanning |
| `git_history_scan.max_commits` | int | `1000` | Maximum commits to scan |
| `git_history_scan.max_age` | string | `"1y"` | Maximum age (e.g., "90d", "6m", "2y") |
| `git_history_scan.scan_removed` | bool | `true` | Track if secrets were later removed |

**AI Analysis (disabled by default, requires ANTHROPIC_API_KEY):**

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `ai_analysis.enabled` | bool | `false` | Enable Claude-powered FP reduction |
| `ai_analysis.max_findings` | int | `50` | Maximum findings to analyze |
| `ai_analysis.confidence_threshold` | float | `0.8` | Threshold to mark as false positive |

**Rotation Guidance (enabled by default):**

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `rotation_guidance` | bool | `true` | Add rotation recommendations to findings |

**How Multi-Source Secrets Detection Works:**

1. **Semgrep Detection**: Pattern-based detection using Semgrep's secrets ruleset
2. **Entropy Analysis**: Shannon entropy calculation to find high-randomness strings
3. **Git History Scanning**: Scans commit history to find secrets that were committed (even if later removed)
4. **Deduplication**: Merges results from all sources, removing duplicates
5. **AI Analysis (Optional)**: Uses Claude to analyze findings and identify false positives
6. **Rotation Guidance**: Adds service-specific rotation instructions

**Entropy Analysis:**

Entropy analysis detects high-randomness strings that may be secrets:

| Entropy Level | Range | Description |
|---------------|-------|-------------|
| High | ≥4.5 | Very likely to be a secret (API keys, tokens) |
| Medium | 3.5-4.5 | Possible secret, needs review |
| Low | <3.5 | Unlikely to be a secret |

Built-in false positive filters:
- Placeholder patterns (`example`, `test`, `YOUR_KEY_HERE`)
- UUIDs and git SHAs
- Known example keys (e.g., `AKIAIOSFODNN7EXAMPLE`)
- Hash outputs with context indicators

**Git History Scanning:**

Scans git commit history to find secrets that were:
- Committed and still present
- Committed but later removed
- Exposed across multiple commits

| Pattern | Description | Severity |
|---------|-------------|----------|
| AWS Access Key | `AKIA[0-9A-Z]{16}` | Critical |
| GitHub Token | `ghp_[A-Za-z0-9]{36,}` | Critical |
| Stripe Live Key | `sk_live_[A-Za-z0-9]{24,}` | Critical |
| OpenAI Key | `sk-[A-Za-z0-9]{48,}` | Critical |
| Private Key | `-----BEGIN.*PRIVATE KEY-----` | Critical |
| Slack Token | `xox[baprs]-*` | Critical |
| Database URL | `postgres://.*:.*@` | Critical |
| JWT Token | `eyJ...` | Medium |

**AI-Powered False Positive Reduction:**

When `ANTHROPIC_API_KEY` is set and AI analysis is enabled:
- Analyzes finding context (surrounding code)
- Identifies placeholder/example values
- Assigns confidence score (0-1)
- Provides reasoning for determination
- Only analyzes medium+ severity findings

**Rotation Recommendations:**

Each finding includes service-specific rotation guidance:

```json
{
  "rotation": {
    "priority": "immediate",
    "steps": [
      "1. Log into AWS Console",
      "2. Navigate to IAM > Security credentials",
      "3. Create new access key pair",
      "4. Update applications",
      "5. Delete old key"
    ],
    "rotation_url": "https://console.aws.amazon.com/iam/",
    "cli_command": "aws iam create-access-key",
    "automation_hint": "Use AWS Secrets Manager"
  },
  "service_provider": "aws"
}
```

Supported providers: AWS, GitHub, Stripe, Slack, OpenAI, Anthropic, Google/GCP, Azure, databases, private keys, NPM, PyPI, Heroku, Vercel, and more.

**Redaction:**
When `redact_secrets` is enabled, secrets are partially masked:
```
Before: sk-abcdef123456789012345678901234567890123456789012
After:  sk-a****9012
```

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

### 3. Git History Security (`git_history_security`)

Scans git history for files that should have been purged - gitignore violations, sensitive files committed by mistake, and generates cleanup recommendations.

**Configuration:**
```json
{
  "secrets": {
    "git_history_security": {
      "enabled": true,
      "max_commits": 1000,
      "max_age": "1y",
      "scan_gitignore_history": true,
      "scan_sensitive_files": true,
      "generate_purge_report": true
    }
  }
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `false` | Enable git history security scanning |
| `max_commits` | int | `1000` | Maximum commits to scan |
| `max_age` | string | `"1y"` | Maximum age (e.g., "90d", "6m", "2y") |
| `scan_gitignore_history` | bool | `true` | Scan for files matching gitignore patterns |
| `scan_sensitive_files` | bool | `true` | Scan for sensitive file patterns |
| `generate_purge_report` | bool | `true` | Generate purge recommendations |

**Gitignore Violations:**

Detects files in git history that match current `.gitignore` rules but were committed before being ignored:

| Pattern | Category | Severity |
|---------|----------|----------|
| `.env`, `.env.*` | credentials | critical |
| `*.pem`, `*.key` | keys | critical |
| `*.p12`, `*.pfx` | certificates | high |
| `.aws/credentials` | credentials | critical |
| `id_rsa`, `id_dsa` | keys | critical |
| `*.sqlite`, `*.db` | database | medium |
| `node_modules/` | dependencies | low |

**Sensitive File Detection:**

Uses RAG-based patterns to detect sensitive files regardless of gitignore:

| Category | Examples | Severity |
|----------|----------|----------|
| Credentials | `.env`, `credentials.json`, `.htpasswd` | critical |
| Keys | `*.pem`, `*.key`, `id_rsa`, `private_key` | critical |
| Certificates | `*.p12`, `*.pfx`, `*.jks` | high |
| Database | `*.sqlite`, `*.db`, database dumps | medium |
| Backups | `*.bak`, `*.backup`, `*.old` | medium |
| IDE/Config | `.vscode/settings.json` (with secrets) | low |

**Purge Recommendations:**

When files should be removed from git history, the scanner generates commands for popular tools:

```json
{
  "purge_recommendations": [
    {
      "file": ".env",
      "reason": "Environment configuration files",
      "severity": "critical",
      "priority": 1,
      "command": "bfg --delete-files '.env'",
      "alternative": "git filter-repo --path '.env' --invert-paths",
      "affected_commits": 15
    }
  ]
}
```

**Shallow Clone Detection:**

The scanner detects shallow clones and returns a helpful message:
```
Repository is a shallow clone. Git history security scanning requires full history.
Use 'git fetch --unshallow' or clone with full depth to enable history scanning.
```

**Risk Score Calculation:**

| Finding | Impact |
|---------|--------|
| Critical file in history | -25 points |
| High severity file | -15 points |
| Medium severity file | -5 points |
| Low severity file | -2 points |

| Score | Risk Level |
|-------|------------|
| 0-39 | critical |
| 40-59 | high |
| 60-79 | medium |
| 80-94 | low |
| 95-100 | excellent |

### 4. API Security (`api`)

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
    "features_run": ["vulns", "secrets", "api", "git_history_security"],
    "git_history_security": {
      "gitignore_violations": [],
      "sensitive_files": [],
      "purge_recommendations": [],
      "timeline": [],
      "summary": {
        "total_violations": 1,
        "gitignore_violations": 0,
        "sensitive_files_found": 1,
        "files_to_purge": 1,
        "commits_scanned": 250,
        "risk_score": 75,
        "risk_level": "medium"
      }
    }
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
      "total_findings": 8,
      "critical": 1,
      "high": 4,
      "medium": 2,
      "low": 1,
      "files_affected": 5,
      "by_type": {
        "aws_credential": 1,
        "github_token": 2,
        "api_key": 2,
        "high_entropy_string": 3
      },
      "by_source": {
        "semgrep": 5,
        "entropy": 2,
        "git_history": 1
      },
      "entropy_findings": 2,
      "history_findings": 1,
      "removed_secrets": 1,
      "false_positives": 2,
      "confirmed_secrets": 3,
      "risk_score": 35,
      "risk_level": "critical"
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
        "snippet": "AKIA****MPLE",
        "entropy": 3.8,
        "entropy_level": "medium",
        "detection_source": "semgrep",
        "service_provider": "aws",
        "rotation": {
          "priority": "immediate",
          "steps": ["1. Create new key", "2. Update apps", "3. Delete old key"],
          "rotation_url": "https://console.aws.amazon.com/iam/"
        }
      },
      {
        "rule_id": "git-history-github_token",
        "type": "github_token",
        "severity": "critical",
        "message": "Secret found in git history",
        "file": "src/auth.js",
        "line": 42,
        "snippet": "ghp_****wxyz",
        "detection_source": "git_history",
        "commit_info": {
          "hash": "abc123def456",
          "short_hash": "abc123de",
          "author": "Developer",
          "date": "2025-01-15T10:30:00Z",
          "message": "Add auth config",
          "is_removed": true
        },
        "service_provider": "github"
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
