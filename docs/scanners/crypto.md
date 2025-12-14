# Crypto Scanner

The Crypto scanner provides comprehensive cryptographic security analysis, detecting weak ciphers, hardcoded keys, insecure random number generation, TLS misconfigurations, and certificate issues.

## Overview

| Property | Value |
|----------|-------|
| **Name** | `crypto` |
| **Version** | 3.0.0 |
| **Output File** | `crypto.json` |
| **Dependencies** | None |
| **Estimated Time** | 30-90 seconds |

## Features

### 1. Ciphers (`ciphers`)

Detects weak and deprecated cryptographic algorithms.

**Configuration:**
```json
{
  "ciphers": {
    "enabled": true,
    "check_weak": true,
    "check_deprecated": true,
    "use_semgrep": true,
    "use_patterns": true
  }
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable cipher analysis |
| `check_weak` | bool | `true` | Check for weak ciphers |
| `check_deprecated` | bool | `true` | Check for deprecated algorithms |
| `use_semgrep` | bool | `true` | Use Semgrep for AST-based detection |
| `use_patterns` | bool | `true` | Use regex pattern matching |

**Detected Algorithms:**

| Algorithm | Severity | CWE | Recommendation |
|-----------|----------|-----|----------------|
| DES/3DES | High | CWE-327 | Use AES-256-GCM |
| RC4/RC2 | Critical | CWE-327 | Use AES-256-GCM |
| MD5 | High | CWE-328 | Use SHA-256/SHA-3 for hashing, bcrypt/argon2 for passwords |
| SHA-1 | Medium | CWE-328 | Use SHA-256 or SHA-3 |
| ECB Mode | High | CWE-327 | Use GCM or CBC with HMAC |
| Blowfish/CAST5/IDEA | Medium | CWE-327 | Use AES-256-GCM |
| 1024-bit RSA | High | CWE-326 | Use 2048+ bit RSA or ECDSA |
| Weak Padding | Medium | CWE-327 | Use OAEP for RSA, PKCS7 for block ciphers |

### 2. Keys (`keys`)

Detects hardcoded cryptographic keys and secrets.

**Configuration:**
```json
{
  "keys": {
    "enabled": true,
    "check_hardcoded": true,
    "check_weak_length": true,
    "min_rsa_bits": 2048,
    "min_ec_bits": 256,
    "check_api_keys": true,
    "check_private": true,
    "check_aws": true,
    "check_signing": true,
    "redact_matches": true
  }
}
```

**Detected Key Types:**

| Type | Severity | Example Pattern |
|------|----------|-----------------|
| API Key | Critical | `api_key = "sk_live_..."` |
| Secret Key | Critical | `secret_key = "..."` |
| Encryption Key | Critical | `aes_key = "..."` |
| Private Key | Critical | `-----BEGIN RSA PRIVATE KEY-----` |
| EC Private Key | Critical | `-----BEGIN EC PRIVATE KEY-----` |
| AWS Access Key | Critical | `AKIA...` |
| Static IV/Nonce | High | `iv = "..."` |
| Signing/HMAC Key | Critical | `hmac_key = "..."` |
| Master/Root Key | Critical | `master_key = "..."` |

### 3. Random (`random`)

Detects insecure random number generation.

**Configuration:**
```json
{
  "random": {
    "enabled": true,
    "check_insecure": true
  }
}
```

**Insecure Random Patterns:**

| Language | Insecure | Secure Alternative | Severity |
|----------|----------|-------------------|----------|
| JavaScript | `Math.random()` | `crypto.getRandomValues()` | High |
| Python | `random.random()` | `secrets` module | High |
| C | `rand()` / `srand()` | `arc4random()` / `getrandom()` | High |
| Java | `java.util.Random` | `SecureRandom` | High |
| Go | `math/rand` | `crypto/rand` | Medium |
| UUID | `uuid.uuid1()` | `uuid.uuid4()` | Medium |

All findings include CWE-338 (Use of Cryptographically Weak PRNG).

### 4. TLS (`tls`)

Detects TLS/SSL misconfigurations.

**Configuration:**
```json
{
  "tls": {
    "enabled": true,
    "check_verification": true,
    "check_protocols": true,
    "min_version": "1.2",
    "check_cipher_suites": true,
    "check_insecure_urls": true
  }
}
```

**Detected Issues:**

| Issue | Severity | CWE | Pattern |
|-------|----------|-----|---------|
| Deprecated Protocol (SSLv2/v3, TLS 1.0) | Critical | CWE-327 | `SSLv3`, `TLSv1.0` |
| TLS 1.1 | High | CWE-327 | `TLSv1.1` |
| Disabled Verification (Go) | Critical | CWE-295 | `InsecureSkipVerify: true` |
| Disabled Verification (Python) | Critical | CWE-295 | `verify=False`, `CERT_NONE` |
| Disabled Verification (Node.js) | Critical | CWE-295 | `rejectUnauthorized: false` |
| Weak Minimum Version | High | CWE-327 | `MinVersion: TLS10` |
| HTTP URLs | Medium | CWE-319 | `http://example.com` |
| NULL Cipher | Critical | CWE-327 | `cipher.*NULL` |
| Weak Cipher Suites | High | CWE-327 | `EXPORT`, `ANON`, `aNULL` |

### 5. Certificates (`certificates`)

Analyzes X.509 certificates in the codebase.

**Configuration:**
```json
{
  "certificates": {
    "enabled": true,
    "check_expiry": true,
    "expiry_warning_days": 30,
    "check_key_strength": true,
    "check_signature_algo": true,
    "check_self_signed": true,
    "check_validity_period": true
  }
}
```

**Supported Certificate Formats:**
- PEM (`.pem`, `.crt`, `.cer`, `.cert`)
- DER (`.der`, `.p7b`, `.p7c`)

**Certificate Checks:**

| Check | Severity | Description |
|-------|----------|-------------|
| Expired | Critical | Certificate has expired |
| Expiring Soon (<30 days) | High | Certificate expires within warning threshold |
| Weak RSA Key (<2048 bits) | Critical | RSA key too small |
| Weak RSA Key (<4096 bits) | Low | RSA key below recommended size |
| Weak ECDSA Key (<256 bits) | High | ECDSA key too small |
| Weak Signature (MD5/SHA1) | High | Weak signature algorithm |
| Self-Signed | Medium | Non-CA self-signed certificate |
| Long Validity (>825 days) | Low | Exceeds CA/Browser Forum guidelines |
| Wildcard | Low | Wildcard certificate detected |
| Private Key in Cert File | High | Private key found in certificate file |

**Certificate Information Extracted:**
```go
type CertInfo struct {
    File          string
    Subject       string
    Issuer        string
    NotBefore     time.Time
    NotAfter      time.Time
    DaysUntilExp  int
    KeyType       string    // RSA, ECDSA, Ed25519
    KeySize       int       // bits
    SignatureAlgo string
    IsSelfSigned  bool
    IsCA          bool
    DNSNames      []string
    Serial        string
}
```

## How It Works

### Technical Flow

1. **Parallel Execution**: All 5 features run concurrently
2. **Semgrep Analysis**: Uses `p/security-audit` and `p/secrets` rules for AST-based detection (ciphers feature)
3. **Pattern Matching**: Scans code files using regex patterns
4. **Certificate Parsing**: Parses PEM/DER certificate files using Go's `crypto/x509`
5. **Deduplication**: Removes duplicate findings from overlapping detection methods
6. **Aggregation**: Combines all findings with severity counts

### File Types Scanned

**Code Files:**
```
.go, .py, .js, .ts, .java, .rb, .php, .cs, .cpp, .c, .h, .hpp, .rs, .swift, .kt
```

**Config Files:**
```
.yaml, .yml, .json, .xml, .conf, .config, .ini, .properties
```

**Certificate Files:**
```
.pem, .crt, .cer, .cert, .der, .p7b, .p7c
```

### Excluded Directories

- `.git`
- `node_modules`
- `vendor`

## Usage

### Command Line

```bash
# Run crypto scanner only
./zero scan --scanner crypto /path/to/repo

# Run crypto profile
./zero hydrate owner/repo --profile crypto-only
```

### Programmatic Usage

```go
import "github.com/crashappsec/zero/pkg/scanners/crypto"

opts := &scanner.ScanOptions{
    RepoPath:  "/path/to/repo",
    OutputDir: "/path/to/output",
    FeatureConfig: map[string]interface{}{
        "ciphers": map[string]interface{}{
            "enabled": true,
            "use_semgrep": true,
        },
        "keys": map[string]interface{}{
            "enabled": true,
            "redact_matches": true,
        },
        "tls": map[string]interface{}{
            "enabled": true,
            "check_verification": true,
        },
    },
}

scanner := &crypto.CryptoScanner{}
result, err := scanner.Run(ctx, opts)
```

## Output Format

```json
{
  "scanner": "crypto",
  "version": "3.0.0",
  "metadata": {
    "features_run": ["ciphers", "keys", "random", "tls", "certificates"]
  },
  "summary": {
    "ciphers": {
      "total_findings": 5,
      "by_severity": {"high": 3, "medium": 2},
      "by_algorithm": {"MD5": 2, "DES/3DES": 1, "SHA-1": 2},
      "used_semgrep": true
    },
    "keys": {
      "total_findings": 3,
      "by_severity": {"critical": 2, "high": 1},
      "by_type": {"api-key": 1, "private-key": 1, "aws-access-key": 1}
    },
    "random": {
      "total_findings": 2,
      "by_severity": {"high": 2},
      "by_type": {"js-math-random": 1, "python-random": 1}
    },
    "tls": {
      "total_findings": 4,
      "by_severity": {"critical": 2, "medium": 2},
      "by_type": {"disabled-verification": 2, "insecure-url": 2}
    },
    "certificates": {
      "total_certificates": 2,
      "total_findings": 1,
      "expired": 0,
      "expiring_soon": 1,
      "weak_key": 0
    }
  },
  "findings": {
    "ciphers": [
      {
        "algorithm": "MD5",
        "severity": "high",
        "file": "src/auth/hash.go",
        "line": 42,
        "description": "MD5 is cryptographically broken for security purposes",
        "match": "crypto.MD5",
        "suggestion": "Use SHA-256 or SHA-3 for hashing, bcrypt/argon2 for passwords",
        "cwe": "CWE-328",
        "source": "semgrep"
      }
    ],
    "keys": [...],
    "random": [...],
    "tls": [...],
    "certificates": {
      "certificates": [...],
      "findings": [...]
    }
  }
}
```

## Prerequisites

| Tool | Required For | Install Command |
|------|--------------|-----------------|
| semgrep | Enhanced cipher detection | `pip install semgrep` or `brew install semgrep` |

Note: The scanner works without Semgrep but provides better detection with it.

## Related Scanners

- **code-security**: May detect overlapping secrets and security issues
- **packages**: Checks for vulnerable crypto libraries in dependencies

## See Also

- [Code Security Scanner](code-security.md) - Complements with SAST analysis
- [OWASP Cryptographic Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Cryptographic_Storage_Cheat_Sheet.html)
- [CWE-327: Use of Broken Crypto Algorithm](https://cwe.mitre.org/data/definitions/327.html)
