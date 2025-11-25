<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Certificate Format Conversion Prompt

## Purpose

Convert certificates between different formats (PEM, DER, PKCS#7, PKCS#12) for platform compatibility and key management.

## Usage

### Format Identification and Conversion

```
Identify the format of this certificate file and provide conversion commands:

File: [filename]
Target format: [PEM/DER/PKCS7/PKCS12]

Provide:
1. Current format detection
2. OpenSSL conversion command
3. Verification command
4. Common issues to watch for
```

### Using Certificate Analyser

```bash
# Analyze certificate file (auto-detects format)
./utils/certificate-analyser/cert-analyser.sh --file certificate.pem

# Analyze PKCS12 file
./utils/certificate-analyser/cert-analyser.sh --file keystore.p12 --password secret
```

## Format Quick Reference

| Format | Extensions | Contains | Usage |
|--------|------------|----------|-------|
| PEM | .pem, .crt, .cer | Certs, keys | Most common |
| DER | .der, .cer | Single cert | Java, Windows |
| PKCS#7 | .p7b, .p7c | Cert chain | Certificate bundles |
| PKCS#12 | .p12, .pfx | Cert + key | Key exchange |

## Conversion Commands

### PEM Conversions

```bash
# PEM to DER
openssl x509 -in cert.pem -outform DER -out cert.der

# PEM to PKCS#7 (certificate bundle)
openssl crl2pkcs7 -nocrl -certfile cert.pem -out cert.p7b

# PEM to PKCS#12 (with private key)
openssl pkcs12 -export -in cert.pem -inkey key.pem -out cert.p12

# PEM to PKCS#12 (with chain)
openssl pkcs12 -export -in cert.pem -inkey key.pem -certfile chain.pem -out cert.p12
```

### DER Conversions

```bash
# DER to PEM
openssl x509 -in cert.der -inform DER -outform PEM -out cert.pem

# DER to PKCS#7
# First convert to PEM, then to PKCS#7
openssl x509 -in cert.der -inform DER -outform PEM -out cert.pem
openssl crl2pkcs7 -nocrl -certfile cert.pem -out cert.p7b
```

### PKCS#7 Conversions

```bash
# PKCS#7 to PEM (extract certificates)
openssl pkcs7 -in bundle.p7b -print_certs -out certs.pem

# PKCS#7 (DER) to PEM
openssl pkcs7 -in bundle.p7b -inform DER -print_certs -out certs.pem
```

### PKCS#12 Conversions

```bash
# PKCS#12 to PEM (certificate only)
openssl pkcs12 -in keystore.p12 -clcerts -nokeys -out cert.pem

# PKCS#12 to PEM (private key only)
openssl pkcs12 -in keystore.p12 -nocerts -out key.pem

# PKCS#12 to PEM (all contents)
openssl pkcs12 -in keystore.p12 -out all.pem

# PKCS#12 to PEM (CA certificates)
openssl pkcs12 -in keystore.p12 -cacerts -nokeys -out chain.pem
```

## Variations

### Identify Unknown Format

```
I have a certificate file but don't know its format:

File: [filename]
First few bytes (hex): [if available]

Help me identify:
1. What format is this?
2. How can I view its contents?
3. What tools can read it?
```

### Create PKCS#12 Keystore

```
Create a PKCS#12 keystore from these components:

- Certificate: cert.pem
- Private key: key.pem
- Intermediate chain: chain.pem
- Friendly name: [name]

Provide the command with appropriate options.
```

### Extract from PKCS#12

```
Extract all components from PKCS#12 file:

File: [keystore.p12]

Extract:
1. End-entity certificate
2. Private key (encrypted and unencrypted)
3. Certificate chain
4. CA certificates

Provide verification commands for each extracted file.
```

### Java KeyStore Conversion

```
Convert between PKCS#12 and Java KeyStore:

Direction: [JKS to PKCS12 / PKCS12 to JKS]
Source: [filename]
Target: [filename]
Alias: [certificate alias]

Provide keytool commands.
```

### Platform-Specific Formats

```
Convert certificate for use on:
- [ ] Windows IIS (needs .pfx)
- [ ] Apache/Nginx (needs .pem)
- [ ] Java application (needs .jks or .p12)
- [ ] AWS ACM (needs .pem)
- [ ] Azure (needs .pfx)

Source file: [filename]
Current format: [format]
```

## Format Detection

### By Extension

| Extension | Likely Format |
|-----------|---------------|
| .pem | PEM |
| .crt | PEM or DER |
| .cer | PEM or DER |
| .der | DER |
| .p7b, .p7c | PKCS#7 |
| .p12, .pfx | PKCS#12 |
| .jks | Java KeyStore |
| .key | PEM private key |

### By Content

```bash
# Check if text (PEM) or binary
file certificate.unknown

# Check first bytes for DER
xxd -l 4 certificate.unknown
# 30 82 xx xx = likely DER

# Check for PEM headers
grep -q "BEGIN CERTIFICATE" certificate.unknown && echo "PEM"

# Try parsing as different formats
openssl x509 -in cert -text -noout 2>/dev/null && echo "PEM certificate"
openssl x509 -in cert -inform DER -text -noout 2>/dev/null && echo "DER certificate"
openssl pkcs7 -in cert -print_certs 2>/dev/null && echo "PKCS7"
openssl pkcs12 -in cert -info -noout 2>/dev/null && echo "PKCS12"
```

## Common Issues

### Password Protection

```bash
# PKCS#12 usually requires password
openssl pkcs12 -in keystore.p12 -passin pass:mypassword -out cert.pem

# Create PKCS#12 with password
openssl pkcs12 -export -in cert.pem -inkey key.pem -out keystore.p12 -passout pass:mypassword

# Create without password (not recommended)
openssl pkcs12 -export -in cert.pem -inkey key.pem -out keystore.p12 -passout pass:
```

### Chain Order

When creating bundles, ensure correct order:
1. End-entity certificate (leaf)
2. Intermediate certificate(s)
3. Root certificate (optional)

```bash
# Combine in correct order
cat cert.pem intermediate.pem > chain.pem
```

### Windows Compatibility

Windows may require specific PKCS#12 settings:

```bash
# For older Windows versions
openssl pkcs12 -export -in cert.pem -inkey key.pem -out cert.pfx \
    -legacy -keypbe PBE-SHA1-3DES -certpbe PBE-SHA1-3DES
```

## Verification Commands

```bash
# Verify PEM certificate
openssl x509 -in cert.pem -text -noout

# Verify DER certificate
openssl x509 -in cert.der -inform DER -text -noout

# Verify PKCS#7
openssl pkcs7 -in bundle.p7b -print_certs

# Verify PKCS#12
openssl pkcs12 -in keystore.p12 -info -noout

# Verify private key matches certificate
cert_mod=$(openssl x509 -noout -modulus -in cert.pem | md5sum)
key_mod=$(openssl rsa -noout -modulus -in key.pem | md5sum)
[[ "$cert_mod" == "$key_mod" ]] && echo "Match" || echo "Mismatch"
```

## Related Prompts

- [security-audit.md](../security/security-audit.md) - Security analysis
- [certificate-comparison.md](certificate-comparison.md) - Compare certificates
- [chain-validation.md](../security/chain-validation.md) - Chain validation

## Related RAG

- [Certificate Formats](../../../rag/certificate-analysis/formats/certificate-formats.md) - Detailed format guide
- [X.509 Certificates](../../../rag/certificate-analysis/x509/x509-certificates.md) - Certificate structure
