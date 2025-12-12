# Scanner Output Formats

All Zero scanners output JSON files following consistent schemas. This document describes the output formats with examples.

## Common Structure

All scanner outputs share a common structure:

```json
{
  "analyzer": "scanner-name",
  "version": "1.0.0",
  "timestamp": "2025-12-12T10:30:00Z",
  "repository": "owner/repo",
  "duration_seconds": 5,
  "summary": {
    // Scanner-specific summary
  },
  "findings": [
    // Array of findings
  ],
  "recommendations": [
    // Remediation recommendations
  ]
}
```

## Manifest File

The `manifest.json` file tracks scan metadata:

```json
{
  "project_id": "expressjs/express",
  "scan_id": "20251212-103000-abc1",
  "schema_version": "2.0.0",
  "git": {
    "commit_hash": "aa907945cd1727483a888a0a6481f9f4861593f8",
    "commit_short": "aa907945",
    "branch": "master",
    "tag": null,
    "commit_date": "2025-08-22T09:12:09+02:00",
    "commit_author": "John Doe <john@example.com>"
  },
  "scan": {
    "started_at": "2025-12-12T10:30:00Z",
    "completed_at": "2025-12-12T10:35:00Z",
    "duration_seconds": 300,
    "profile": "security",
    "scanners_requested": ["package-sbom", "package-vulns", "code-secrets"],
    "scanners_completed": ["package-sbom", "package-vulns", "code-secrets"],
    "scanners_failed": []
  },
  "analyses": {
    "package-sbom": {
      "analyzer": "package-sbom.sh",
      "version": "1.0.0",
      "started_at": "2025-12-12T10:30:00Z",
      "completed_at": "2025-12-12T10:30:05Z",
      "duration_ms": 5000,
      "status": "complete",
      "output_file": "package-sbom.json",
      "summary": {
        "format": "CycloneDX",
        "generator": "cdxgen",
        "direct": 28,
        "total": 156
      }
    }
  },
  "summary": {
    "risk_level": "medium",
    "total_dependencies": 156,
    "direct_dependencies": 28,
    "total_vulnerabilities": 5,
    "total_security_findings": 12,
    "critical_count": 0,
    "high_count": 2,
    "license_status": "pass"
  }
}
```

## Package Scanners

### package-sbom.json

Software Bill of Materials:

```json
{
  "analyzer": "package-sbom",
  "version": "1.0.0",
  "timestamp": "2025-12-12T10:30:00Z",
  "sbom_format": "CycloneDX",
  "sbom_generator": "cdxgen",
  "sbom_file": "sbom.cdx.json",
  "direct_dependencies": 28,
  "total_dependencies": 156,
  "summary": {
    "format": "CycloneDX",
    "generator": "cdxgen",
    "direct": 28,
    "total": 156
  }
}
```

### package-vulns.json

Vulnerability findings:

```json
{
  "analyzer": "vulnerability-analyser",
  "version": "1.0.0",
  "timestamp": "2025-12-12T10:30:00Z",
  "repository": "expressjs/express",
  "summary": {
    "total": 5,
    "critical": 0,
    "high": 2,
    "medium": 2,
    "low": 1,
    "cisa_kev": 1
  },
  "vulnerabilities": [
    {
      "id": "GHSA-abc1-2345-6789",
      "aliases": ["CVE-2024-1234"],
      "package": {
        "name": "lodash",
        "version": "4.17.20",
        "ecosystem": "npm"
      },
      "severity": "high",
      "cvss_score": 7.5,
      "cvss_vector": "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:N/A:H",
      "epss_score": 0.023,
      "cisa_kev": false,
      "summary": "Prototype Pollution in lodash",
      "details": "Versions before 4.17.21 are vulnerable to Prototype Pollution...",
      "affected_versions": "<4.17.21",
      "fixed_version": "4.17.21",
      "references": [
        "https://github.com/advisories/GHSA-abc1-2345-6789"
      ],
      "published": "2024-03-15T00:00:00Z"
    }
  ]
}
```

### package-health.json

Package health assessment:

```json
{
  "analyzer": "package-health-analyser",
  "version": "1.0.0",
  "timestamp": "2025-12-12T10:30:00Z",
  "summary": {
    "abandoned": 2,
    "typosquat_risk": 1,
    "deprecated": 0,
    "healthy": 153
  },
  "findings": [
    {
      "package": "left-pad",
      "version": "1.0.0",
      "ecosystem": "npm",
      "issue_type": "abandoned",
      "severity": "medium",
      "last_publish": "2016-03-23T00:00:00Z",
      "days_since_publish": 3186,
      "recommendation": "Consider replacing with String.prototype.padStart()"
    },
    {
      "package": "loadsh",
      "version": "1.0.0",
      "ecosystem": "npm",
      "issue_type": "typosquat",
      "severity": "high",
      "similar_to": "lodash",
      "recommendation": "Verify this is the intended package, not lodash"
    }
  ]
}
```

### licenses.json

License compliance:

```json
{
  "analyzer": "licenses",
  "version": "1.0.0",
  "timestamp": "2025-12-12T10:30:00Z",
  "summary": {
    "overall_status": "pass",
    "license_violations": 0,
    "dependency_license_violations": 0,
    "license_warnings": 2,
    "total_dependencies_with_licenses": 156
  },
  "project_license": {
    "spdx_id": "MIT",
    "name": "MIT License",
    "category": "permissive"
  },
  "dependency_licenses": {
    "MIT": 120,
    "Apache-2.0": 25,
    "ISC": 8,
    "BSD-3-Clause": 3
  },
  "findings": [
    {
      "package": "some-gpl-package",
      "version": "1.0.0",
      "license": "GPL-3.0",
      "severity": "warning",
      "issue": "Copyleft license may require source disclosure",
      "recommendation": "Review GPL-3.0 obligations for your use case"
    }
  ]
}
```

## Code Scanners

### code-vulns.json

SAST findings:

```json
{
  "analyzer": "code-vulns",
  "version": "1.0.0",
  "timestamp": "2025-12-12T10:30:00Z",
  "repository": "/path/to/repo",
  "scanner": {
    "engine": "semgrep",
    "ruleset": "p/security-audit + p/owasp-top-ten"
  },
  "duration_seconds": 45,
  "summary": {
    "risk_score": 72,
    "risk_level": "medium",
    "total_findings": 8,
    "critical_count": 0,
    "high_count": 2,
    "medium_count": 4,
    "low_count": 2,
    "by_type": {
      "sql_injection": 1,
      "xss": 2,
      "path_traversal": 1,
      "hardcoded_secret": 2,
      "insecure_random": 2
    },
    "files_affected": 5
  },
  "findings": [
    {
      "rule_id": "javascript.express.security.sql-injection",
      "type": "sql_injection",
      "severity": "high",
      "message": "SQL injection vulnerability detected",
      "file": "src/db/users.js",
      "line": 42,
      "column": 15,
      "code_snippet": "db.query(`SELECT * FROM users WHERE id = ${userId}`)",
      "cwe": ["CWE-89"],
      "owasp": ["A03:2021"],
      "detector": "semgrep"
    }
  ],
  "recommendations": [
    "Use parameterized queries to prevent SQL injection",
    "Implement input validation on all user-supplied data",
    "Enable pre-commit hooks with security linters"
  ]
}
```

### code-secrets.json

Secret detection:

```json
{
  "analyzer": "code-secrets",
  "version": "2.0.0",
  "timestamp": "2025-12-12T10:30:00Z",
  "repository": "/path/to/repo",
  "scanner": {
    "engine": "semgrep",
    "ruleset": "p/secrets + custom"
  },
  "duration_seconds": 30,
  "summary": {
    "risk_score": 45,
    "risk_level": "high",
    "total_findings": 3,
    "critical_count": 1,
    "high_count": 1,
    "medium_count": 1,
    "low_count": 0,
    "by_type": {
      "aws_credential": 1,
      "github_token": 1,
      "api_key": 1
    },
    "files_affected": 3
  },
  "findings": [
    {
      "rule_id": "zero.cloud-providers.aws.secret.access-key",
      "type": "aws_credential",
      "severity": "critical",
      "message": "Potential AWS Access Key exposed",
      "file": "config/settings.js",
      "line": 15,
      "column": 20,
      "snippet": "accessKey: 'AKIA********EXAMPLE'",
      "detector": "semgrep"
    }
  ],
  "recommendations": [
    "URGENT: Rotate all critical secrets immediately - they may already be compromised",
    "Use environment variables or a secrets manager for sensitive data",
    "Enable pre-commit hooks to prevent secret commits (e.g., git-secrets, detect-secrets)"
  ]
}
```

## Cryptography Scanners

### crypto-ciphers.json

Weak cipher detection:

```json
{
  "analyzer": "crypto-ciphers",
  "version": "1.0.0",
  "timestamp": "2025-12-12T10:30:00Z",
  "repository": "/path/to/repo",
  "scanner": {
    "engine": "semgrep",
    "ruleset": "p/security-audit + crypto-security.yaml"
  },
  "duration_seconds": 15,
  "summary": {
    "risk_score": 60,
    "risk_level": "medium",
    "total_findings": 4,
    "critical_count": 1,
    "high_count": 2,
    "medium_count": 1,
    "by_issue_type": {
      "DEPRECATED_CIPHER": 2,
      "WEAK_HASH": 1,
      "ECB_MODE": 1
    },
    "files_affected": 3
  },
  "findings": [
    {
      "rule_id": "crypto.deprecated-cipher.des",
      "severity": "critical",
      "message": "DES encryption detected - this cipher is broken",
      "file": "src/crypto/legacy.py",
      "line": 45,
      "column": 10,
      "code_snippet": "cipher = DES.new(key, DES.MODE_ECB)",
      "issue_type": "DEPRECATED_CIPHER",
      "cipher": "DES",
      "cwe": ["CWE-327"],
      "detector": "semgrep"
    }
  ],
  "recommendations": [
    "CRITICAL: Replace DES with AES-256-GCM",
    "Replace MD5/SHA1 with SHA-256 or SHA-3 for security purposes",
    "Avoid ECB mode - use GCM or CBC with random IV"
  ]
}
```

### crypto-keys.json

Key management findings:

```json
{
  "analyzer": "crypto-keys",
  "version": "1.0.0",
  "timestamp": "2025-12-12T10:30:00Z",
  "repository": "/path/to/repo",
  "summary": {
    "risk_score": 55,
    "risk_level": "high",
    "total_findings": 3,
    "critical_count": 1,
    "high_count": 1,
    "medium_count": 1,
    "by_issue_type": {
      "HARDCODED_KEY": 1,
      "PRIVATE_KEY": 1,
      "WEAK_KEY_LENGTH": 1
    }
  },
  "findings": [
    {
      "rule_id": "crypto.hardcoded-key.symmetric",
      "severity": "critical",
      "message": "Hardcoded symmetric encryption key detected",
      "file": "src/utils/encrypt.js",
      "line": 12,
      "code_snippet": "const key = 'supersecretkey123';",
      "issue_type": "HARDCODED_KEY",
      "cwe": ["CWE-321"],
      "detector": "semgrep"
    }
  ],
  "recommendations": [
    "CRITICAL: Move hardcoded keys to environment variables or secrets manager",
    "Never commit private keys to version control",
    "Use key derivation functions (KDF) for password-based encryption"
  ]
}
```

### crypto-tls.json

TLS configuration findings:

```json
{
  "analyzer": "crypto-tls",
  "version": "1.0.0",
  "timestamp": "2025-12-12T10:30:00Z",
  "repository": "/path/to/repo",
  "scanner": {
    "engine": "semgrep",
    "ruleset": "p/insecure-transport + p/security-audit"
  },
  "summary": {
    "risk_score": 50,
    "risk_level": "high",
    "total_findings": 3,
    "critical_count": 1,
    "high_count": 1,
    "medium_count": 1,
    "by_issue_type": {
      "CERT_VERIFICATION_DISABLED": 1,
      "DEPRECATED_TLS_VERSION": 1,
      "INSECURE_HTTP": 1
    }
  },
  "findings": [
    {
      "rule_id": "python.requests.security.ssl-verify-false",
      "severity": "critical",
      "message": "Certificate verification disabled - vulnerable to MITM attacks",
      "file": "src/api/client.py",
      "line": 25,
      "code_snippet": "requests.get(url, verify=False)",
      "issue_type": "CERT_VERIFICATION_DISABLED",
      "cwe": ["CWE-295"],
      "detector": "semgrep"
    }
  ],
  "recommendations": [
    "CRITICAL: Enable certificate verification - disabled verification allows MITM attacks",
    "TLS 1.0/1.1 are deprecated - set minimum version to TLS 1.2",
    "Use HTTPS for all external communication"
  ]
}
```

### crypto-random.json

Insecure random detection:

```json
{
  "analyzer": "crypto-random",
  "version": "1.0.0",
  "timestamp": "2025-12-12T10:30:00Z",
  "repository": "/path/to/repo",
  "summary": {
    "risk_score": 75,
    "risk_level": "medium",
    "total_findings": 2,
    "high_count": 1,
    "medium_count": 1,
    "low_count": 0,
    "by_language": {
      "javascript": 1,
      "python": 1
    }
  },
  "findings": [
    {
      "rule_id": "javascript.math-random.insecure",
      "severity": "high",
      "message": "Math.random() used in security context",
      "file": "src/auth/tokens.js",
      "line": 18,
      "code_snippet": "const token = Math.random().toString(36).substring(2);",
      "language": "javascript",
      "cwe": ["CWE-330", "CWE-338"],
      "detector": "semgrep"
    }
  ],
  "recommendations": [
    "JavaScript: Use crypto.randomBytes() or crypto.getRandomValues() instead of Math.random()",
    "Python: Use secrets module or os.urandom() instead of random module for security"
  ]
}
```

## Infrastructure Scanners

### iac-security.json

IaC security findings:

```json
{
  "analyzer": "iac-security",
  "version": "1.0.0",
  "timestamp": "2025-12-12T10:30:00Z",
  "repository": "/path/to/repo",
  "scanner": {
    "engine": "checkov",
    "version": "3.2.0"
  },
  "summary": {
    "total_findings": 5,
    "critical_count": 1,
    "high_count": 2,
    "medium_count": 2,
    "by_resource_type": {
      "aws_s3_bucket": 2,
      "aws_security_group": 2,
      "aws_iam_policy": 1
    }
  },
  "findings": [
    {
      "check_id": "CKV_AWS_19",
      "severity": "critical",
      "message": "S3 bucket has public access enabled",
      "file": "terraform/s3.tf",
      "line": 5,
      "resource": "aws_s3_bucket.data",
      "guideline": "https://docs.bridgecrew.io/docs/bc_aws_s3_19"
    }
  ]
}
```

## See Also

- [Scanner Reference](reference.md) - Available scanners
- [Scanner Architecture](../architecture/scanners.md) - How scanners work
- [Agent Reference](../agents/README.md) - Agents that consume this data
