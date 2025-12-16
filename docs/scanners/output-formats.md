# Scanner Output Formats

Zero uses 8 super scanners, each outputting a single consolidated JSON file. This document describes the output formats with examples.

## Output Files

| Scanner | Output File | Description |
|---------|-------------|-------------|
| sbom | `sbom.json` + `sbom.cdx.json` | SBOM summary and CycloneDX SBOM |
| package-analysis | `package-analysis.json` | Vulnerabilities, health, licenses, malcontent |
| crypto | `crypto.json` | Ciphers, keys, TLS, random |
| code-security | `code-security.json` | SAST, secrets, API security |
| code-quality | `code-quality.json` | Tech debt, complexity, coverage, docs |
| devops | `devops.json` | IaC, containers, GitHub Actions, DORA |
| tech-id | `technology.json` | Technology detection, ML-BOM |
| code-ownership | `code-ownership.json` | Contributors, bus factor, CODEOWNERS |

## Common Structure

All scanner outputs share a common structure:

```json
{
  "scanner": "scanner-name",
  "version": "1.0.0",
  "timestamp": "2025-12-12T10:30:00Z",
  "repository": "owner/repo",
  "duration_seconds": 5,
  "metadata": {
    "features_run": ["feature1", "feature2"]
  },
  "summary": {
    // Scanner-specific summary by feature
  },
  "findings": {
    // Findings organized by feature
  }
}
```

## Manifest File

The `manifest.json` file tracks scan metadata:

```json
{
  "project_id": "expressjs/express",
  "scan_id": "20251212-103000-abc1",
  "schema_version": "3.0.0",
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
    "scanners_requested": ["sbom", "package-analysis", "code-security"],
    "scanners_completed": ["sbom", "package-analysis", "code-security"],
    "scanners_failed": []
  },
  "analyses": {
    "sbom": {
      "version": "1.0.0",
      "started_at": "2025-12-12T10:30:00Z",
      "completed_at": "2025-12-12T10:30:05Z",
      "duration_ms": 5000,
      "status": "complete",
      "output_file": "sbom.json",
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

## SBOM Scanner

### sbom.json

Software Bill of Materials summary:

```json
{
  "scanner": "sbom",
  "version": "1.0.0",
  "timestamp": "2025-12-12T10:30:00Z",
  "metadata": {
    "features_run": ["generation", "integrity"]
  },
  "summary": {
    "generation": {
      "format": "CycloneDX",
      "generator": "cdxgen",
      "sbom_file": "sbom.cdx.json",
      "direct_dependencies": 28,
      "total_dependencies": 156
    },
    "integrity": {
      "lockfile_drift": false,
      "lockfiles_checked": 2
    }
  }
}
```

## Package Analysis Scanner

### package-analysis.json

Consolidated supply chain security findings:

```json
{
  "scanner": "package-analysis",
  "version": "3.0.0",
  "timestamp": "2025-12-12T10:30:00Z",
  "metadata": {
    "features_run": ["vulns", "health", "licenses", "malcontent"],
    "sbom_source": "sbom scanner",
    "component_count": 156
  },
  "summary": {
    "vulns": {
      "total": 5,
      "critical": 0,
      "high": 2,
      "medium": 2,
      "low": 1,
      "cisa_kev": 1
    },
    "health": {
      "abandoned": 2,
      "typosquat_risk": 1,
      "deprecated": 0,
      "healthy": 153
    },
    "licenses": {
      "overall_status": "pass",
      "allowed": 150,
      "denied": 0,
      "needs_review": 6
    },
    "malcontent": {
      "total_files": 1500,
      "files_with_risk": 12,
      "critical": 0,
      "high": 2
    }
  },
  "findings": {
    "vulns": [
      {
        "id": "GHSA-abc1-2345-6789",
        "aliases": ["CVE-2024-1234"],
        "package": "lodash",
        "version": "4.17.20",
        "ecosystem": "npm",
        "severity": "high",
        "cvss_score": 7.5,
        "cisa_kev": false,
        "summary": "Prototype Pollution in lodash",
        "fixed_version": "4.17.21"
      }
    ],
    "health": [
      {
        "package": "left-pad",
        "version": "1.0.0",
        "ecosystem": "npm",
        "issue_type": "abandoned",
        "severity": "medium",
        "recommendation": "Consider replacing with String.prototype.padStart()"
      }
    ],
    "licenses": [
      {
        "package": "some-gpl-package",
        "version": "1.0.0",
        "license": "GPL-3.0",
        "severity": "warning",
        "issue": "Copyleft license may require source disclosure"
      }
    ]
  }
}
```

## Code Security Scanner

### code-security.json

Consolidated SAST, secrets, and API security findings:

```json
{
  "scanner": "code-security",
  "version": "1.0.0",
  "timestamp": "2025-12-12T10:30:00Z",
  "metadata": {
    "features_run": ["vulns", "secrets", "api"]
  },
  "summary": {
    "vulns": {
      "total": 8,
      "critical": 0,
      "high": 2,
      "medium": 4,
      "low": 2,
      "by_type": {
        "sql_injection": 1,
        "xss": 2,
        "path_traversal": 1
      }
    },
    "secrets": {
      "total": 3,
      "critical": 1,
      "high": 1,
      "medium": 1,
      "by_type": {
        "aws_credential": 1,
        "github_token": 1,
        "api_key": 1
      }
    },
    "api": {
      "total": 2,
      "owasp_api_violations": ["API1:2023", "API3:2023"]
    }
  },
  "findings": {
    "vulns": [
      {
        "rule_id": "javascript.express.security.sql-injection",
        "type": "sql_injection",
        "severity": "high",
        "message": "SQL injection vulnerability detected",
        "file": "src/db/users.js",
        "line": 42,
        "cwe": ["CWE-89"],
        "owasp": ["A03:2021"]
      }
    ],
    "secrets": [
      {
        "rule_id": "zero.cloud-providers.aws.secret.access-key",
        "type": "aws_credential",
        "severity": "critical",
        "message": "Potential AWS Access Key exposed",
        "file": "config/settings.js",
        "line": 15,
        "snippet": "accessKey: 'AKIA********EXAMPLE'"
      }
    ]
  }
}
```

## Crypto Scanner

### crypto.json

Consolidated cryptographic security findings:

```json
{
  "scanner": "crypto",
  "version": "1.0.0",
  "timestamp": "2025-12-12T10:30:00Z",
  "metadata": {
    "features_run": ["ciphers", "keys", "random", "tls", "certificates"]
  },
  "summary": {
    "ciphers": {
      "total": 4,
      "critical": 1,
      "high": 2,
      "medium": 1,
      "by_algorithm": {
        "DES": 1,
        "MD5": 1,
        "SHA-1": 1,
        "ECB_MODE": 1
      }
    },
    "keys": {
      "total": 3,
      "critical": 1,
      "high": 1,
      "medium": 1,
      "hardcoded_keys": 1,
      "private_keys": 1,
      "weak_keys": 1
    },
    "random": {
      "total": 2,
      "high": 1,
      "medium": 1,
      "by_language": {
        "javascript": 1,
        "python": 1
      }
    },
    "tls": {
      "total": 3,
      "critical": 1,
      "high": 1,
      "medium": 1,
      "cert_verification_disabled": 1,
      "deprecated_tls": 1
    },
    "certificates": {
      "total": 2,
      "expired": 0,
      "weak_key": 1
    }
  },
  "findings": {
    "ciphers": [
      {
        "rule_id": "crypto.deprecated-cipher.des",
        "severity": "critical",
        "algorithm": "DES",
        "message": "DES encryption detected - this cipher is broken",
        "file": "src/crypto/legacy.py",
        "line": 45,
        "cwe": ["CWE-327"],
        "suggestion": "Replace with AES-256-GCM"
      }
    ],
    "keys": [
      {
        "rule_id": "crypto.hardcoded-key.symmetric",
        "severity": "critical",
        "type": "HARDCODED_KEY",
        "message": "Hardcoded symmetric encryption key detected",
        "file": "src/utils/encrypt.js",
        "line": 12,
        "cwe": ["CWE-321"]
      }
    ],
    "random": [
      {
        "rule_id": "javascript.math-random.insecure",
        "severity": "high",
        "message": "Math.random() used in security context",
        "file": "src/auth/tokens.js",
        "line": 18,
        "cwe": ["CWE-330", "CWE-338"],
        "suggestion": "Use crypto.randomBytes() or crypto.getRandomValues()"
      }
    ],
    "tls": [
      {
        "rule_id": "python.requests.security.ssl-verify-false",
        "severity": "critical",
        "type": "CERT_VERIFICATION_DISABLED",
        "message": "Certificate verification disabled - vulnerable to MITM attacks",
        "file": "src/api/client.py",
        "line": 25,
        "cwe": ["CWE-295"]
      }
    ]
  }
}
```

## DevOps Scanner

### devops.json

Consolidated DevOps and infrastructure findings:

```json
{
  "scanner": "devops",
  "version": "1.0.0",
  "timestamp": "2025-12-12T10:30:00Z",
  "metadata": {
    "features_run": ["iac", "containers", "github_actions", "dora", "git"]
  },
  "summary": {
    "iac": {
      "total": 5,
      "critical": 1,
      "high": 2,
      "medium": 2,
      "by_resource_type": {
        "aws_s3_bucket": 2,
        "aws_security_group": 2,
        "aws_iam_policy": 1
      }
    },
    "containers": {
      "total": 3,
      "images_scanned": 2,
      "vulnerabilities": 3
    },
    "github_actions": {
      "total": 2,
      "injection_risks": 1,
      "permission_issues": 1
    },
    "dora": {
      "deployment_frequency": "weekly",
      "lead_time_for_changes": "2 days",
      "mttr": "4 hours",
      "change_failure_rate": 0.15
    },
    "git": {
      "total_commits_30d": 45,
      "active_contributors": 8
    }
  },
  "findings": {
    "iac": [
      {
        "check_id": "CKV_AWS_19",
        "severity": "critical",
        "message": "S3 bucket has public access enabled",
        "file": "terraform/s3.tf",
        "line": 5,
        "resource": "aws_s3_bucket.data"
      }
    ],
    "github_actions": [
      {
        "severity": "high",
        "type": "injection_risk",
        "file": ".github/workflows/ci.yml",
        "message": "Potential command injection via untrusted input"
      }
    ]
  }
}
```

## See Also

- [Scanner Reference](reference.md) - Available scanners
- [Scanner Architecture](../architecture/scanners.md) - How scanners work
- [Agent Reference](../agents/README.md) - Agents that consume this data
