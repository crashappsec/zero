<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# CA/Browser Forum Compliance Prompt

## Purpose

Verify certificate compliance with CA/Browser Forum Baseline Requirements for publicly trusted TLS/SSL certificates.

## Usage

### Full Compliance Check

```
Check the CA/Browser Forum Baseline Requirements compliance for [domain].

Verify:
1. Validity period (≤398 days)
2. Key size (RSA ≥2048, ECC ≥P-256)
3. Signature algorithm (SHA-256+)
4. Subject Alternative Name present
5. Basic Constraints correct
6. Key Usage appropriate
7. Certificate Transparency (SCTs present)
8. OCSP/CRL availability
9. Certificate Policy OID

Report compliance status for each requirement with specific failures.
```

### Using Certificate Analyser

```bash
# CA/B Forum compliance check
./utils/certificate-analyser/cert-analyser.sh --compliance example.com

# Full check with all validations
./utils/certificate-analyser/cert-analyser.sh --all-checks example.com
```

## Example Output

### Compliance Summary

```
CA/Browser Forum Baseline Requirements Compliance
=================================================
Domain: example.com
Analyzed: 2024-11-25

Overall: 7/7 Requirements PASSED ✓
```

### Detailed Results

| Requirement | Status | Details |
|-------------|--------|---------|
| Validity Period | ✓ Pass | 365 days (≤398 required) |
| Key Strength | ✓ Pass | RSA 2048-bit (≥2048 required) |
| Signature Algorithm | ✓ Pass | SHA-256 (SHA-256+ required) |
| SAN Extension | ✓ Pass | Present with DNS names |
| Basic Constraints | ✓ Pass | CA:FALSE for end-entity |
| Key Usage | ✓ Pass | digitalSignature, keyEncipherment |
| Certificate Transparency | ✓ Pass | 2 SCTs embedded |

### Non-Compliant Example

| Requirement | Status | Details |
|-------------|--------|---------|
| Validity Period | ❌ Fail | 730 days (exceeds 398 day limit) |
| Key Strength | ⚠️ Minimum | RSA 2048-bit (recommend 3072+) |
| Signature Algorithm | ✓ Pass | SHA-256 |
| Certificate Transparency | ❌ Fail | No SCTs found |

## Variations

### Pre-Issuance Compliance Check

```
Before issuing a certificate with these parameters:
- Domain: [domain]
- Validity: [days] days
- Key: [algorithm] [size]-bit
- SAN: [list of names]

Check if this would comply with CA/B Forum Baseline Requirements.
Flag any issues before the certificate is created.
```

### Audit Documentation

```
Generate CA/Browser Forum compliance documentation for audit purposes:

Domain: [domain]

For each BR requirement, provide:
1. Requirement number and description
2. Certificate evidence (exact values)
3. Compliance determination
4. Auditor notes

Format suitable for compliance evidence package.
```

### Historical Compliance

```
Check if certificate for [domain] complied with CA/B Forum requirements
at time of issuance.

Certificate was issued: [issue date]

Note: Requirements change over time:
- Before Sept 2020: 825-day max validity
- Before March 2018: 39-month max validity
- SHA-1 prohibited since Jan 2017

Evaluate against rules in effect at issuance.
```

### Validation Type Check

```
Determine the validation level of the certificate for [domain]:

- DV (Domain Validated): OID 2.23.140.1.2.1
- OV (Organization Validated): OID 2.23.140.1.2.2
- EV (Extended Validation): OID 2.23.140.1.1

Verify the certificate meets requirements for its claimed validation level.
```

### Wildcard Compliance

```
Check wildcard certificate compliance for [*.domain]:

BR requirements for wildcards:
- Only in leftmost label
- Cannot be registry-controlled (*.com invalid)
- No multi-level wildcards (*.*.domain invalid)
- Private-label-only wildcards require DNS validation

Verify compliance with wildcard-specific rules.
```

## CA/B Forum Requirements Checklist

### Section 6.1.5 - Key Sizes

**RSA**:
- [ ] Minimum 2048 bits
- [ ] Recommended 3072+ bits for new keys

**ECDSA**:
- [ ] Minimum P-256 curve
- [ ] P-384 and P-521 also acceptable

### Section 6.1.6 - Signature Algorithms

**Acceptable**:
- [ ] sha256WithRSAEncryption
- [ ] sha384WithRSAEncryption
- [ ] sha512WithRSAEncryption
- [ ] ecdsa-with-SHA256
- [ ] ecdsa-with-SHA384
- [ ] ecdsa-with-SHA512

**Prohibited**:
- [ ] No MD2, MD4, MD5
- [ ] No SHA-1 (since Jan 2017)

### Section 6.3.2 - Validity Period

- [ ] Maximum 398 days (since Sept 2020)
- [ ] notBefore not more than 48 hours before issuance
- [ ] notAfter correctly calculated

### Section 7.1 - Certificate Profile

**Subject Alternative Name**:
- [ ] Extension present
- [ ] Contains all FQDNs
- [ ] dNSName entries for domains
- [ ] iPAddress entries for IPs (if applicable)

**Basic Constraints**:
- [ ] Present for CA certificates (critical)
- [ ] CA:FALSE or absent for end-entity
- [ ] pathLenConstraint appropriate

**Key Usage**:
- [ ] Present (should be critical)
- [ ] digitalSignature for TLS
- [ ] keyEncipherment (RSA) or keyAgreement (ECDSA)

**Extended Key Usage**:
- [ ] Present for end-entity
- [ ] id-kp-serverAuth (1.3.6.1.5.5.7.3.1) required
- [ ] id-kp-clientAuth optional
- [ ] No anyExtendedKeyUsage

### Section 7.1.2.3 - Certificate Transparency

- [ ] At least 2 SCTs from different logs
- [ ] Logs must be qualified by browser programs
- [ ] SCTs embedded, via TLS extension, or OCSP

### Section 4.9 - Revocation

- [ ] OCSP responder URL in AIA
- [ ] CRL Distribution Point (recommended)
- [ ] OCSP response validity ≤10 days

## Policy OIDs

| Type | OID | Description |
|------|-----|-------------|
| DV | 2.23.140.1.2.1 | Domain Validated |
| OV | 2.23.140.1.2.2 | Organization Validated |
| EV | 2.23.140.1.1 | Extended Validation |
| IV | 2.23.140.1.2.3 | Individual Validated |

## Timeline of Policy Changes

| Date | Change |
|------|--------|
| 2016-06-01 | SHA-1 prohibited for new certs |
| 2017-01-01 | SHA-1 completely prohibited |
| 2018-03-01 | 825-day max validity |
| 2018-04-30 | CT required for all certs |
| 2020-09-01 | 398-day max validity |
| 2024+ | Shorter validity expected |

## Non-Compliance Consequences

### For CAs
- Root program removal
- Audit failures
- Trust store distrust

### For Certificate Holders
- Browser warnings
- Connection failures
- Compliance violations

## Related Prompts

- [security-audit.md](../security/security-audit.md) - Full security audit
- [expiry-monitoring.md](expiry-monitoring.md) - Expiration tracking
- [chain-validation.md](../security/chain-validation.md) - Chain validation

## Related RAG

- [CA/B Forum Requirements](../../../rag/certificate-analysis/cab-forum/baseline-requirements.md) - Full requirements
- [X.509 Certificates](../../../rag/certificate-analysis/x509/x509-certificates.md) - Certificate structure
- [Revocation](../../../rag/certificate-analysis/revocation/ocsp-crl.md) - OCSP/CRL
