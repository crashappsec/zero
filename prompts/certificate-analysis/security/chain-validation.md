<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Certificate Chain Validation Prompt

## Purpose

Validate X.509 certificate chains, verify trust paths, and diagnose chain-related issues.

## Usage

### Basic Chain Validation

```
Validate the certificate chain for [domain].

Check:
1. Chain completeness (leaf → intermediate → root)
2. Proper chain ordering
3. Signature verification at each level
4. Validity period alignment
5. Trust anchor in system store
6. Name chaining (subject → issuer match)

Report any broken links or trust issues.
```

### Using Certificate Analyser

```bash
# Validate certificate chain
./utils/certificate-analyser/cert-analyser.sh --verify-chain example.com

# Full validation with OCSP
./utils/certificate-analyser/cert-analyser.sh --verify-chain --check-ocsp example.com
```

## Example Output

### Chain Structure

```
Certificate Chain for example.com:

[0] Leaf Certificate
    Subject: CN=example.com
    Issuer: CN=DigiCert TLS RSA SHA256 2020 CA1
    Valid: 2024-01-15 to 2025-01-14
    Key: RSA 2048-bit

[1] Intermediate CA
    Subject: CN=DigiCert TLS RSA SHA256 2020 CA1
    Issuer: CN=DigiCert Global Root CA
    Valid: 2020-04-14 to 2031-04-13
    Key: RSA 2048-bit

[2] Root CA (in trust store)
    Subject: CN=DigiCert Global Root CA
    Issuer: CN=DigiCert Global Root CA (self-signed)
    Valid: 2006-11-10 to 2031-11-10
    Key: RSA 2048-bit
```

### Validation Results

| Check | Status | Details |
|-------|--------|---------|
| Chain Complete | ✓ Pass | All certificates present |
| Chain Order | ✓ Pass | Correct leaf → root order |
| Signature Valid | ✓ Pass | All signatures verified |
| Trust Anchor | ✓ Pass | DigiCert Global Root CA trusted |
| Name Chaining | ✓ Pass | Subject/Issuer match verified |
| Validity Overlap | ✓ Pass | No gaps in validity periods |

## Variations

### Troubleshoot Chain Issues

```
The certificate for [domain] is showing trust errors.

Diagnose:
1. Is the chain complete?
2. Are intermediates missing?
3. Is the root CA trusted?
4. Is any certificate in the chain expired?
5. Are there any signature verification failures?

Provide specific fixes for any issues found.
```

### Verify Against Specific Trust Store

```
Validate the certificate chain for [domain] against:
- [ ] System trust store
- [ ] Mozilla CA bundle
- [ ] Java cacerts
- [ ] Custom CA bundle: [path]

Report which trust stores accept the chain and which reject it.
```

### Cross-Signed Certificate Analysis

```
Analyze the certificate chain for [domain] for cross-signing:

1. Are there multiple trust paths?
2. Which root CAs can anchor the chain?
3. Is the cross-signed intermediate included?
4. What happens if old root is distrusted?

This is important for Let's Encrypt and similar cross-signed scenarios.
```

### Internal CA Chain Validation

```
Validate the certificate chain for internal certificate [hostname]:

Given our internal CA structure:
- Root CA: [internal-root.pem]
- Intermediate CA: [internal-intermediate.pem]

Check:
1. Chain builds to our internal root
2. All certificates properly signed
3. pathLenConstraint respected
4. Name constraints followed
```

### Compare Chains

```
Compare certificate chains between:
- Production: [prod-domain]
- Staging: [staging-domain]

Check:
1. Same CA issuer?
2. Same trust path?
3. Consistent chain depth?
4. Validity alignment?
```

## Chain Validation Checks

### Structure Checks
- [ ] Leaf certificate present
- [ ] All intermediate certificates present
- [ ] Chain in correct order (leaf first)
- [ ] Root certificate in trust store
- [ ] No unnecessary certificates included

### Cryptographic Checks
- [ ] All signatures valid
- [ ] Key sizes meet requirements
- [ ] Signature algorithms acceptable
- [ ] No self-signed leaf (unless intentional)

### Policy Checks
- [ ] Basic Constraints correct for each level
- [ ] CA:TRUE for all non-leaf certificates
- [ ] pathLenConstraint respected
- [ ] Key Usage includes keyCertSign for CAs
- [ ] Name constraints followed (if present)

### Validity Checks
- [ ] No expired certificates
- [ ] Leaf validity within CA validity
- [ ] Reasonable validity periods
- [ ] Not-yet-valid check passed

## Common Chain Issues

### Missing Intermediate

**Symptom**: Browser shows untrusted error, but certificate is valid.

**Diagnosis**:
```bash
# Check what's being sent
openssl s_client -connect example.com:443 -showcerts

# Should show leaf + intermediate(s), not just leaf
```

**Fix**: Configure server to send full chain.

### Wrong Chain Order

**Symptom**: Some clients fail, others succeed.

**Diagnosis**: Leaf should be first, then intermediates, root optional.

**Fix**: Reorder certificates in chain file.

### Expired Intermediate

**Symptom**: Certificate valid but chain fails validation.

**Diagnosis**:
```bash
# Check each certificate's validity
openssl x509 -in intermediate.pem -noout -dates
```

**Fix**: Replace intermediate with valid one from CA.

### Cross-Signed Root Issue

**Symptom**: Works on new devices, fails on old ones.

**Diagnosis**: Old devices may not have new root.

**Fix**: Include cross-signed intermediate for backwards compatibility.

## OpenSSL Commands

```bash
# Show full chain from server
openssl s_client -connect example.com:443 -showcerts

# Verify chain against system store
openssl verify -CApath /etc/ssl/certs server.crt

# Verify chain with intermediates
openssl verify -CAfile ca-bundle.crt -untrusted intermediate.crt server.crt

# Check specific certificate in chain
openssl x509 -in cert.pem -noout -subject -issuer -dates

# Build chain from AIA (Authority Information Access)
# (requires downloading intermediate from AIA URL)
```

## Related Prompts

- [security-audit.md](security-audit.md) - Full security audit
- [tls-troubleshooting.md](../troubleshooting/tls-troubleshooting.md) - Connection issues
- [cab-forum-compliance.md](../compliance/cab-forum-compliance.md) - Compliance check

## Related RAG

- [X.509 Certificates](../../../rag/certificate-analysis/x509/x509-certificates.md) - Chain structure
- [Certificate Formats](../../../rag/certificate-analysis/formats/certificate-formats.md) - Format handling
- [TLS Security](../../../rag/certificate-analysis/tls-security/best-practices.md) - Configuration
