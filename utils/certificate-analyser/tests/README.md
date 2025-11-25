# Certificate Analyser Test Suite

This directory contains test certificates and utilities for testing the certificate-analyser tool.

## Directory Structure

```
tests/
├── README.md                    # This file
├── generate-test-certs.sh       # Script to generate test certificates
└── fixtures/                    # Generated test certificates
    ├── MANIFEST.md              # Detailed certificate inventory
    ├── *.pem                    # PEM format certificates
    ├── *.der                    # DER format certificates
    ├── *.p7b                    # PKCS7 bundles
    ├── *.p12, *.pfx             # PKCS12 keystores
    └── *.key                    # Private keys (test only!)
```

## Generating Test Certificates

To generate (or regenerate) all test certificates:

```bash
cd utils/certificate-analyser/tests
./generate-test-certs.sh
```

This creates certificates for testing:
- **Format detection**: PEM, DER, PKCS7, PKCS12
- **Compliance checking**: Valid, non-compliant (long validity), weak crypto
- **Expiry warnings**: 7-day, 30-day expiring certs
- **Certificate types**: Server, client auth, code signing
- **Key types**: RSA 1024/2048/4096, ECC P-256
- **Signature algorithms**: SHA-1 (deprecated), SHA-256, SHA-384

## PKCS12 Password

All PKCS12/PFX files use the password: `testpassword123`

This is also stored in `fixtures/keystore-password.txt`

## Test Scenarios

See `fixtures/MANIFEST.md` for detailed test scenarios and usage examples.

## Security Notice

These certificates are for **testing purposes only**. The private keys are included and should **never** be used in production environments.
