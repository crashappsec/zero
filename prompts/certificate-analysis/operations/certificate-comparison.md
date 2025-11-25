<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Certificate Comparison Prompt

## Purpose

Compare X.509 certificates to verify identity, detect changes, or validate deployments.

## Usage

### Basic Comparison

```
Compare these two certificates:

Certificate 1: [file1.pem or domain1]
Certificate 2: [file2.pem or domain2]

Compare:
1. Subject and Issuer
2. Serial number
3. Validity period
4. Public key (fingerprint)
5. SAN entries
6. Signature algorithm
7. Extensions

Highlight differences and determine if they are the same certificate.
```

### Using Certificate Analyser

```bash
# Compare two certificate files
./utils/certificate-analyser/cert-analyser.sh --compare cert1.pem cert2.pem

# Compare local file with remote
./utils/certificate-analyser/cert-analyser.sh --compare-remote cert.pem example.com
```

## Example Output

### Same Certificate

```
Certificate Comparison Report
=============================

Certificates: IDENTICAL ✓

| Property | Certificate 1 | Certificate 2 |
|----------|--------------|---------------|
| Subject | CN=example.com | CN=example.com |
| Serial | 0x1234... | 0x1234... |
| Fingerprint (SHA-256) | AB:CD:EF... | AB:CD:EF... |
| Issuer | DigiCert TLS CA | DigiCert TLS CA |
| Valid From | 2024-01-15 | 2024-01-15 |
| Valid To | 2025-01-14 | 2025-01-14 |

Conclusion: Same certificate (fingerprints match)
```

### Different Certificates

```
Certificate Comparison Report
=============================

Certificates: DIFFERENT ❌

| Property | Certificate 1 | Certificate 2 | Match |
|----------|--------------|---------------|-------|
| Subject | CN=example.com | CN=example.com | ✓ |
| Serial | 0x1234... | 0x5678... | ❌ |
| Fingerprint | AB:CD:EF... | 12:34:56... | ❌ |
| Issuer | DigiCert | Let's Encrypt | ❌ |
| Valid From | 2024-01-15 | 2024-06-01 | ❌ |
| Valid To | 2025-01-14 | 2024-08-30 | ❌ |
| Key Size | RSA 2048 | RSA 2048 | ✓ |

Conclusion: Different certificates for same domain
- Certificate 2 appears to be a renewal
- Different CA used
- Shorter validity period
```

## Variations

### Pre/Post Deployment Comparison

```
Compare deployed certificate with expected:

Expected (local file): [cert.pem]
Deployed (remote): [domain:port]

Verify:
1. Is the deployed certificate the expected one?
2. Same fingerprint?
3. Same validity dates?
4. Chain matches expected?

Report any mismatches.
```

### Renewal Validation

```
Verify certificate renewal for [domain]:

Old certificate: [old.pem]
New certificate: [new.pem]

Check:
1. Same domain coverage (SAN)?
2. Different serial number (should be)?
3. New validity period?
4. Same or stronger key?
5. Same CA or different?

Confirm renewal was successful.
```

### Cross-Environment Comparison

```
Compare certificates across environments:

Production: [prod-domain]
Staging: [staging-domain]
Development: [dev-domain]

Report:
1. Are they the same certificate?
2. Same CA?
3. Same validity?
4. Consistency issues?

Identify environment-specific differences.
```

### Wildcard Coverage Comparison

```
Compare wildcard certificate coverage:

Certificate 1: *.example.com
Certificate 2: *.example.com, example.com

Check:
1. Same wildcard scope?
2. Apex domain coverage?
3. Multi-level subdomains?
4. SAN list completeness?
```

### Chain Comparison

```
Compare certificate chains:

Chain 1: [from domain1 or file]
Chain 2: [from domain2 or file]

Compare:
1. Chain depth
2. Intermediate certificates
3. Root CA
4. Any missing intermediates?
5. Order differences?
```

## Comparison Properties

### Exact Match Required
- SHA-256 fingerprint
- Serial number
- Public key

### Should Match (for renewals)
- Subject (CN)
- SAN entries
- Key algorithm

### May Differ
- Validity dates
- Serial number
- CA (if switching providers)

### Should Improve
- Key size (same or larger)
- Signature algorithm (same or stronger)

## OpenSSL Commands

```bash
# Get certificate fingerprint
openssl x509 -in cert.pem -noout -fingerprint -sha256

# Compare serial numbers
openssl x509 -in cert1.pem -noout -serial
openssl x509 -in cert2.pem -noout -serial

# Compare subjects
openssl x509 -in cert1.pem -noout -subject
openssl x509 -in cert2.pem -noout -subject

# Compare SANs
openssl x509 -in cert1.pem -noout -text | grep -A1 "Subject Alternative Name"
openssl x509 -in cert2.pem -noout -text | grep -A1 "Subject Alternative Name"

# Compare public keys
openssl x509 -in cert1.pem -noout -pubkey | md5sum
openssl x509 -in cert2.pem -noout -pubkey | md5sum

# Compare from remote
echo | openssl s_client -connect domain:443 2>/dev/null | openssl x509 -fingerprint -sha256
```

## Comparison Script

```bash
#!/bin/bash
# compare-certs.sh - Compare two certificates

cert1="$1"
cert2="$2"

echo "Comparing certificates..."
echo ""

# Fingerprints
fp1=$(openssl x509 -in "$cert1" -noout -fingerprint -sha256 | cut -d= -f2)
fp2=$(openssl x509 -in "$cert2" -noout -fingerprint -sha256 | cut -d= -f2)

if [[ "$fp1" == "$fp2" ]]; then
    echo "Result: IDENTICAL (same fingerprint)"
else
    echo "Result: DIFFERENT"
    echo ""
    echo "Certificate 1: $fp1"
    echo "Certificate 2: $fp2"
fi

# Subjects
echo ""
echo "Subjects:"
echo "  1: $(openssl x509 -in "$cert1" -noout -subject)"
echo "  2: $(openssl x509 -in "$cert2" -noout -subject)"

# Validity
echo ""
echo "Validity:"
echo "  1: $(openssl x509 -in "$cert1" -noout -dates)"
echo "  2: $(openssl x509 -in "$cert2" -noout -dates)"
```

## Use Cases

### Deployment Verification
Ensure the correct certificate was deployed to production.

### Backup Validation
Confirm backup certificate matches production.

### Certificate Inventory
Identify duplicate certificates across systems.

### Incident Response
Verify if a compromised certificate is deployed.

### Migration Validation
Confirm certificate was correctly migrated to new infrastructure.

## Related Prompts

- [security-audit.md](../security/security-audit.md) - Full security audit
- [chain-validation.md](../security/chain-validation.md) - Chain analysis
- [expiry-monitoring.md](../compliance/expiry-monitoring.md) - Expiry tracking

## Related RAG

- [X.509 Certificates](../../../rag/certificate-analysis/x509/x509-certificates.md) - Certificate structure
- [Certificate Formats](../../../rag/certificate-analysis/formats/certificate-formats.md) - Format handling
