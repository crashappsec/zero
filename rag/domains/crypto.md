# Cryptography Domain Knowledge

This document consolidates RAG knowledge for the **crypto** super scanner.

## Features Covered
- **ciphers**: Weak/deprecated cipher detection
- **keys**: Hardcoded keys and weak key lengths
- **random**: Insecure random number generation
- **tls**: TLS/SSL misconfiguration
- **certificates**: X.509 certificate analysis

## Related RAG Directories

### Cryptography
- `rag/cryptography/` - Core crypto knowledge
  - `weak-ciphers/` - DES, 3DES, RC4, Blowfish, MD5, SHA1, ECB mode
  - `insecure-random/` - Math.random, random module, time-based seeds
  - `hardcoded-keys/` - Embedded keys, key derivation issues
  - `tls-misconfig/` - Certificate verification, protocol versions

### Certificate Analysis
- `rag/certificate-analysis/` - X.509 certificate knowledge
  - Certificate chain validation
  - Key usage and extensions
  - Expiration monitoring

## Key Concepts

### Weak Ciphers
| Algorithm | Status | Replacement |
|-----------|--------|-------------|
| DES | Deprecated | AES-256 |
| 3DES | Deprecated | AES-256 |
| RC4 | Broken | AES-GCM |
| MD5 | Broken | SHA-256+ |
| SHA1 | Deprecated | SHA-256+ |
| ECB mode | Insecure | CBC, GCM, CTR |
| Blowfish | Deprecated | AES |

### Key Security
- **Minimum RSA**: 2048 bits (3072+ recommended)
- **Minimum ECDSA**: 256 bits
- **Never hardcode**: Private keys, symmetric keys, API keys
- **Key derivation**: Use PBKDF2, Argon2, scrypt (not raw passwords)

### Random Number Generation
| Language | Insecure | Secure |
|----------|----------|--------|
| JavaScript | Math.random() | crypto.getRandomValues() |
| Python | random.* | secrets.*, os.urandom() |
| Go | math/rand | crypto/rand |
| Java | java.util.Random | java.security.SecureRandom |

### TLS Configuration
- **Minimum version**: TLS 1.2 (1.3 preferred)
- **Deprecated**: SSL 2.0, SSL 3.0, TLS 1.0, TLS 1.1
- **Certificate verification**: Must be enabled
- **Hostname verification**: Must be enabled
- **Cipher suites**: Prefer AEAD (GCM, ChaCha20-Poly1305)

### Certificate Requirements
- Valid chain to trusted root
- Not expired
- Proper key usage extensions
- RSA 2048+ or ECDSA 256+ key size
- SHA-256+ signature algorithm

## Agent Expertise

### Gill Agent
The **Gill** agent (cryptography specialist) should be consulted for:
- Weak cipher analysis
- Key management recommendations
- TLS configuration review
- Certificate security assessment

### Razor Agent
The **Razor** agent (code security) may assist with:
- Hardcoded secret detection correlation
- Security vulnerability context

## Output Schema

The crypto scanner produces a single `crypto.json` file with:
```json
{
  "features_run": ["ciphers", "keys", "random", "tls", "certificates"],
  "summary": {
    "ciphers": { "total_findings": N, "critical": N, ... },
    "keys": { "total_findings": N, "hardcoded": N, ... },
    "random": { "total_findings": N, ... },
    "tls": { "total_findings": N, ... },
    "certificates": { "total_found": N, "issues": N, ... }
  },
  "findings": {
    "ciphers": [...],
    "keys": [...],
    "random": [...],
    "tls": [...],
    "certificates": [...]
  }
}
```

## Severity Classification

| Finding Type | Critical | High | Medium | Low |
|--------------|----------|------|--------|-----|
| Cipher | Broken (MD5, RC4) | Deprecated (DES, 3DES) | Weak mode (ECB) | Legacy |
| Keys | Hardcoded private key | Hardcoded symmetric key | Weak key length | Static IV |
| Random | Crypto with Math.random | Seeding with time | Predictable seed | - |
| TLS | Verify disabled | SSL/TLS 1.0 | TLS 1.1 | Missing HSTS |
| Certificate | Self-signed in prod | Expired | Weak signature | Near expiry |

## Detection Patterns

### Semgrep Rules
The scanner uses Semgrep with:
- `p/secrets` for hardcoded credentials
- Custom rules for crypto-specific patterns
- Language-specific crypto API detection

### Pattern Examples
```yaml
# Weak cipher detection
patterns:
  - pattern: DES.new(...)
  - pattern: Cipher.getInstance("DES")
  - pattern: crypto.createCipher("des", ...)
```
