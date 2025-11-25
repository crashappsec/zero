<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Certificate Security Audit Prompt

## Purpose

Perform a comprehensive security audit of X.509 certificates, covering cryptographic strength, configuration, compliance, and operational security.

## Usage

### Basic Security Audit

```
Perform a comprehensive security audit of the certificate for [domain].

Analyze:
1. Cryptographic security (key size, signature algorithm, curves)
2. Certificate configuration (extensions, validity, SAN)
3. CA/Browser Forum compliance
4. Certificate chain trust
5. Revocation status (OCSP/CRL)
6. Certificate Transparency (SCT presence)

Provide:
- Risk categorization (Critical/Warning/Info)
- Prioritized remediation recommendations
- Timeline for addressing issues
```

### Using Certificate Analyser

```bash
# Run full security audit
./utils/certificate-analyser/cert-analyser.sh --all-checks example.com

# With Claude AI analysis
./utils/certificate-analyser/cert-analyser.sh --claude --all-checks example.com
```

## Example Output

### Executive Summary

| Metric | Value | Status |
|--------|-------|--------|
| Domain | example.com | ✓ |
| Days to Expiry | 45 | ⚠️ Warning |
| Key Strength | RSA 2048-bit | ⚠️ Minimum |
| Signature | SHA-256 | ✓ Compliant |
| CT Logs | 2 SCTs | ✓ Present |
| OCSP Status | Good | ✓ Valid |

### Critical Findings

| Issue | Risk | Recommendation |
|-------|------|----------------|
| RSA 2048-bit key | Medium | Upgrade to RSA 4096 or ECC P-384 |
| Expiry in 45 days | Medium | Schedule renewal |

### Recommendations

**Immediate (0-7 days)**:
- None

**Short-term (7-30 days)**:
- Schedule certificate renewal
- Consider key algorithm upgrade

**Long-term (30+ days)**:
- Implement automated certificate renewal
- Plan migration to shorter validity certificates

## Variations

### Quick Security Check

```
Quickly assess the certificate security for [domain]:
- Is the key strong enough? (RSA 2048+ or ECC P-256+)
- Is SHA-256 or stronger used?
- Is the certificate close to expiry?
- Are there any critical issues?

Just the essentials, no deep dive needed.
```

### Detailed Cryptographic Analysis

```
Analyze the cryptographic strength of the certificate for [domain]:

1. Public key algorithm and size
2. Signature algorithm
3. Hash function strength
4. Key exchange parameters
5. ECDSA curve (if applicable)
6. Comparison to current NIST/NSA recommendations

Include vulnerability exposure (ROBOT, DROWN, etc.) based on configuration.
```

### Internal Certificate Audit

```
Audit the internal certificate for [hostname]:

Since this is an internal/private certificate:
- Note: CA/B Forum rules may not apply
- Focus on: key strength, expiry, purpose
- Check: internal CA trust, chain completeness
- Evaluate: appropriate for use case

Include recommendations for internal PKI best practices.
```

### Multi-Domain Audit

```
Perform security audits for the following domains:
- api.example.com
- www.example.com
- mail.example.com
- admin.example.com

Provide:
1. Summary table of all certificates
2. Common issues across domains
3. Inconsistencies (different issuers, key sizes, etc.)
4. Unified remediation plan
```

## Analysis Checklist

### Cryptographic Security
- [ ] Key algorithm (RSA, ECDSA, EdDSA)
- [ ] Key size meets minimum (RSA 2048+, ECC P-256+)
- [ ] Signature algorithm (SHA-256+)
- [ ] No deprecated algorithms (MD5, SHA-1)

### Certificate Configuration
- [ ] Valid not-before and not-after dates
- [ ] Appropriate validity period (≤398 days)
- [ ] Subject Alternative Names present
- [ ] Basic Constraints correct for type
- [ ] Key Usage appropriate
- [ ] Extended Key Usage includes serverAuth

### Compliance
- [ ] CA/B Forum Baseline Requirements
- [ ] Certificate Transparency SCTs
- [ ] OCSP responder accessible
- [ ] CRL distribution point (if required)

### Trust & Chain
- [ ] Trusted root CA
- [ ] Complete intermediate chain
- [ ] Proper chain order
- [ ] No expired chain certificates

### Operational Security
- [ ] Reasonable expiry timeline
- [ ] Renewal process documented
- [ ] OCSP stapling enabled (server)
- [ ] HSTS configured (server)

## Risk Levels

### Critical ❌
Issues requiring immediate attention:
- Expired certificate
- Key size < 2048 bits
- SHA-1 or MD5 signature
- No trusted chain
- Revoked certificate

### Warning ⚠️
Issues to address soon:
- Expiry within 30 days
- RSA 2048 (minimum acceptable)
- Long validity period (>398 days)
- Missing CT logs
- OCSP failures

### Informational ℹ️
Best practices and recommendations:
- Consider ECC for performance
- Plan for shorter validity periods
- Enable OCSP stapling
- Implement automation

## Related Prompts

- [chain-validation.md](chain-validation.md) - Detailed chain analysis
- [cab-forum-compliance.md](../compliance/cab-forum-compliance.md) - Compliance check
- [tls-troubleshooting.md](../troubleshooting/tls-troubleshooting.md) - Issue diagnosis

## Related RAG

- [X.509 Certificates](../../../rag/certificate-analysis/x509/x509-certificates.md)
- [TLS Security](../../../rag/certificate-analysis/tls-security/best-practices.md)
- [CA/B Forum Requirements](../../../rag/certificate-analysis/cab-forum/baseline-requirements.md)
