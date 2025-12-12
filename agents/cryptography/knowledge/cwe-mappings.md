# CWE Mappings for Cryptographic Vulnerabilities

## Primary CWEs

### CWE-327: Use of a Broken or Risky Cryptographic Algorithm
**Description**: Using a broken, weak, or risky algorithm (DES, RC4, MD5, SHA1).

**Examples**:
- DES encryption
- RC4 stream cipher
- MD5 for integrity/authentication
- SHA1 for digital signatures

**Detection Patterns**:
- `Cipher.getInstance("DES")`
- `crypto.createHash('md5')`
- `hashlib.sha1()`

**Remediation**: Replace with modern algorithms (AES-GCM, SHA-256).

---

### CWE-328: Reversible One-Way Hash
**Description**: Using a weak hash that allows collisions or preimage attacks.

**Examples**:
- MD5 for password storage
- SHA1 for digital signatures

**Related CVEs**:
- CVE-2004-2761 (MD5 collision)
- CVE-2017-15361 (SHA1 collision - SHAttered)

---

### CWE-321: Use of Hard-coded Cryptographic Key
**Description**: Cryptographic key embedded directly in source code.

**Examples**:
- `key = "mySecretKey123"`
- `AES_KEY = b'\x00\x01\x02...'`
- Private keys in repositories

**Impact**: Anyone with source code access can decrypt data.

**Remediation**: Use secrets managers, environment variables, HSMs.

---

### CWE-798: Use of Hard-coded Credentials
**Description**: Password, API key, or credential embedded in code.

**Related to**: CWE-321 but broader scope.

**Examples**:
- `PASSWORD = "admin123"`
- `API_KEY = "sk-..."`
- Connection strings with credentials

---

### CWE-330: Use of Insufficiently Random Values
**Description**: Using non-cryptographic random for security purposes.

**Examples**:
- `Math.random()` for session tokens
- `random.randint()` for password generation
- `java.util.Random` for key generation

**Impact**: Attackers can predict "random" values.

**Remediation**: Use cryptographically secure RNG (SecureRandom, secrets, crypto/rand).

---

### CWE-338: Use of Cryptographically Weak PRNG
**Description**: Similar to CWE-330, specific to weak PRNGs.

**Examples**:
- Mersenne Twister (Python's random)
- LCG (Linear Congruential Generator)

---

### CWE-326: Inadequate Encryption Strength
**Description**: Key length or algorithm provides insufficient security.

**Examples**:
- RSA 1024-bit keys
- AES-128 when 256 is needed
- 64-bit block ciphers (Blowfish, 3DES)

**Thresholds**:
- RSA: >= 2048 bits
- ECC: >= 256 bits
- Symmetric: >= 128 bits (256 preferred)

---

### CWE-295: Improper Certificate Validation
**Description**: Not verifying SSL/TLS certificates properly.

**Examples**:
- `verify=False` in requests
- `InsecureSkipVerify: true` in Go
- `rejectUnauthorized: false` in Node.js
- Custom TrustManager accepting all certs

**Impact**: MITM attacks possible.

---

### CWE-757: Selection of Less-Secure Algorithm During Negotiation
**Description**: Allowing weak protocols/ciphers in TLS negotiation.

**Examples**:
- SSLv3 enabled
- TLS 1.0/1.1 as minimum
- Weak cipher suites allowed

---

### CWE-311: Missing Encryption of Sensitive Data
**Description**: Sensitive data transmitted or stored without encryption.

**Examples**:
- HTTP instead of HTTPS
- Passwords stored in plaintext
- PII in logs without redaction

---

### CWE-329: Not Using a Random IV with CBC Mode
**Description**: Reusing or using predictable IVs with CBC mode.

**Examples**:
- Static IV: `iv = b'\x00' * 16`
- Hardcoded IV in source

**Impact**: IV reuse can leak plaintext patterns.

**Remediation**: Generate random IV per encryption, store with ciphertext.

---

## CWE Hierarchy

```
CWE-310: Cryptographic Issues
├── CWE-311: Missing Encryption of Sensitive Data
├── CWE-312: Cleartext Storage of Sensitive Information
├── CWE-319: Cleartext Transmission of Sensitive Information
├── CWE-320: Key Management Errors
│   ├── CWE-321: Use of Hard-coded Cryptographic Key
│   └── CWE-322: Key Exchange without Entity Authentication
├── CWE-326: Inadequate Encryption Strength
├── CWE-327: Use of a Broken or Risky Cryptographic Algorithm
│   └── CWE-328: Reversible One-Way Hash
├── CWE-329: Not Using a Random IV with CBC Mode
├── CWE-330: Use of Insufficiently Random Values
│   └── CWE-338: Use of Cryptographically Weak PRNG
├── CWE-347: Improper Verification of Cryptographic Signature
└── CWE-757: Selection of Less-Secure Algorithm During Negotiation
```

## OWASP Mapping

| OWASP Top 10 2021 | Related CWEs |
|-------------------|--------------|
| A02: Cryptographic Failures | CWE-310 family |
| A07: Identification and Authentication Failures | CWE-798, CWE-321 |
| A09: Security Logging and Monitoring Failures | CWE-311, CWE-312 |
