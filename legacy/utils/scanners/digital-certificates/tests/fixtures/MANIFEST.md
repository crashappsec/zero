# Test Certificate Fixtures

This directory contains test certificates for the certificate-analyser tool.

## Certificate Inventory

### Certificate Authority Certificates
| File | Description | Key Type | Validity |
|------|-------------|----------|----------|
| `root-ca.pem` | Root CA certificate | RSA 4096 | 10 years |
| `root-ca.key` | Root CA private key | RSA 4096 | N/A |
| `intermediate-ca.pem` | Intermediate CA certificate | RSA 4096 | 5 years |
| `intermediate-ca.key` | Intermediate CA private key | RSA 4096 | N/A |

### End-Entity Certificates (PEM)
| File | Description | Key Type | Signature | Validity | Compliance |
|------|-------------|----------|-----------|----------|------------|
| `valid-cert.pem` | Standard valid certificate | RSA 2048 | SHA-256 | 397 days | Compliant |
| `strong-cert.pem` | Strong crypto certificate | RSA 4096 | SHA-384 | 365 days | Compliant |
| `ecc-cert.pem` | ECC certificate | P-256 | SHA-256 | 365 days | Compliant |
| `weak-cert.pem` | Weak crypto (deprecated) | RSA 1024 | SHA-1 | 365 days | Non-compliant |
| `expired-cert.pem` | Expired certificate | RSA 2048 | SHA-256 | 1 day | Expired |
| `long-validity-cert.pem` | Long validity (non-compliant) | RSA 2048 | SHA-256 | 825 days | Non-compliant |
| `self-signed.pem` | Self-signed certificate | RSA 2048 | SHA-256 | 365 days | Self-signed |
| `multi-san.pem` | Multiple SANs certificate | RSA 2048 | SHA-256 | 365 days | Compliant |
| `expiring-soon.pem` | Expiring in 7 days | RSA 2048 | SHA-256 | 7 days | Warning |
| `expiring-warning.pem` | Expiring in 30 days | RSA 2048 | SHA-256 | 30 days | Warning |
| `client-auth.pem` | Client authentication | RSA 2048 | SHA-256 | 365 days | Compliant |
| `code-signing.pem` | Code signing certificate | RSA 2048 | SHA-256 | 365 days | Compliant |

### DER Format
| File | Description |
|------|-------------|
| `valid-cert.der` | Valid certificate in DER format |
| `root-ca.der` | Root CA in DER format |
| `ecc-cert.der` | ECC certificate in DER format |

### PKCS7 Bundles
| File | Description |
|------|-------------|
| `chain-bundle.p7b` | Certificate chain in PKCS7 PEM format |
| `chain-bundle-der.p7b` | Certificate chain in PKCS7 DER format |

### PKCS12 Keystores
| File | Description | Password |
|------|-------------|----------|
| `valid-cert.p12` | Valid cert with key | `testpassword123` |
| `valid-cert.pfx` | Same as above (.pfx extension) | `testpassword123` |
| `ecc-cert.p12` | ECC cert with key | `testpassword123` |
| `strong-cert-chain.p12` | Strong cert with full chain | `testpassword123` |

### Other Files
| File | Description |
|------|-------------|
| `chain-bundle.pem` | Full chain bundle (leaf + intermediate + root) |
| `keystore-password.txt` | Password for PKCS12 files |

## Usage Examples

### Reading PEM certificate
```bash
openssl x509 -in valid-cert.pem -text -noout
```

### Reading DER certificate
```bash
openssl x509 -in valid-cert.der -inform DER -text -noout
```

### Reading PKCS7 bundle
```bash
openssl pkcs7 -in chain-bundle.p7b -print_certs -noout
```

### Reading PKCS12 keystore
```bash
openssl pkcs12 -in valid-cert.p12 -info -passin pass:testpassword123
```

### Verifying certificate chain
```bash
openssl verify -CAfile root-ca.pem -untrusted intermediate-ca.pem valid-cert.pem
```

## Test Scenarios

1. **Format Detection**: Test auto-detection with `.pem`, `.der`, `.p7b`, `.p12` files
2. **Chain Validation**: Use `chain-bundle.pem` to test chain verification
3. **Compliance Checking**: Compare `valid-cert.pem` (compliant) vs `long-validity-cert.pem` (non-compliant)
4. **Expiry Warnings**: Test with `expiring-soon.pem` (7 days) and `expiring-warning.pem` (30 days)
5. **Crypto Strength**: Compare `strong-cert.pem` vs `weak-cert.pem`
6. **Self-Signed Detection**: Use `self-signed.pem`
7. **ECC Support**: Use `ecc-cert.pem`
8. **Multiple SANs**: Use `multi-san.pem`
9. **Certificate Types**: Use `client-auth.pem` and `code-signing.pem`

## Regenerating Certificates

To regenerate all test certificates:
```bash
cd utils/certificate-analyser/tests
./generate-test-certs.sh
```

Note: Some certificates (like `expired-cert.pem`) may need the `faketime` utility for accurate backdating.

## Security Notice

These certificates are for **testing purposes only**. Never use these certificates or keys in production environments.

