#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Test Certificate Generator
# Generates test certificates in various formats for certificate-analyser testing
#
# Usage: ./generate-test-certs.sh
#
# This script generates:
# - Root CA certificate
# - Intermediate CA certificate
# - End-entity certificates (valid, expired, weak crypto, etc.)
# - Certificates in multiple formats (PEM, DER, PKCS7, PKCS12)
#############################################################################

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
FIXTURES_DIR="$SCRIPT_DIR/fixtures"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Create fixtures directory
mkdir -p "$FIXTURES_DIR"
cd "$FIXTURES_DIR"

echo ""
echo "========================================="
echo "  Test Certificate Generator"
echo "========================================="
echo ""

#############################################################################
# 1. Generate Root CA
#############################################################################
log_info "Generating Root CA..."

# Root CA private key (RSA 4096)
openssl genrsa -out root-ca.key 4096 2>/dev/null

# Root CA certificate (10 years validity)
openssl req -x509 -new -nodes \
    -key root-ca.key \
    -sha256 \
    -days 3650 \
    -out root-ca.pem \
    -subj "/C=US/ST=California/L=San Francisco/O=Test Root CA/OU=Certificate Testing/CN=Test Root CA" \
    2>/dev/null

log_success "Root CA created: root-ca.pem"

#############################################################################
# 2. Generate Intermediate CA
#############################################################################
log_info "Generating Intermediate CA..."

# Intermediate CA private key
openssl genrsa -out intermediate-ca.key 4096 2>/dev/null

# Intermediate CA CSR
openssl req -new \
    -key intermediate-ca.key \
    -out intermediate-ca.csr \
    -subj "/C=US/ST=California/L=San Francisco/O=Test Root CA/OU=Certificate Testing/CN=Test Intermediate CA" \
    2>/dev/null

# Intermediate CA extensions config (LibreSSL compatible)
cat > intermediate-ca.ext << EOF
basicConstraints = critical, CA:TRUE, pathlen:0
keyUsage = critical, digitalSignature, cRLSign, keyCertSign
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid
EOF

# Sign Intermediate CA with Root CA (5 years)
openssl x509 -req \
    -in intermediate-ca.csr \
    -CA root-ca.pem \
    -CAkey root-ca.key \
    -CAcreateserial \
    -out intermediate-ca.pem \
    -days 1825 \
    -sha256 \
    -extfile intermediate-ca.ext \
    2>/dev/null

log_success "Intermediate CA created: intermediate-ca.pem"

#############################################################################
# 3. Generate Valid End-Entity Certificate (RSA 2048, SHA-256)
#############################################################################
log_info "Generating valid end-entity certificate..."

openssl genrsa -out valid-cert.key 2048 2>/dev/null

openssl req -new \
    -key valid-cert.key \
    -out valid-cert.csr \
    -subj "/C=US/ST=California/L=San Francisco/O=Test Organization/OU=IT Department/CN=test.example.com" \
    2>/dev/null

# End-entity extensions (LibreSSL compatible)
cat > end-entity.ext << EOF
basicConstraints = CA:FALSE
keyUsage = critical, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth, clientAuth
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid
subjectAltName = @alt_names

[alt_names]
DNS.1 = test.example.com
DNS.2 = www.test.example.com
DNS.3 = api.test.example.com
DNS.4 = *.test.example.com
IP.1 = 192.168.1.1
EOF

# Sign with Intermediate CA (397 days - CA/B Forum compliant)
openssl x509 -req \
    -in valid-cert.csr \
    -CA intermediate-ca.pem \
    -CAkey intermediate-ca.key \
    -CAcreateserial \
    -out valid-cert.pem \
    -days 397 \
    -sha256 \
    -extfile end-entity.ext \
    2>/dev/null

log_success "Valid certificate created: valid-cert.pem"

#############################################################################
# 4. Generate Strong Certificate (RSA 4096, SHA-384)
#############################################################################
log_info "Generating strong certificate (RSA 4096)..."

openssl genrsa -out strong-cert.key 4096 2>/dev/null

openssl req -new \
    -key strong-cert.key \
    -out strong-cert.csr \
    -subj "/C=US/ST=California/L=San Francisco/O=Secure Organization/CN=secure.example.com" \
    2>/dev/null

cat > strong-cert.ext << EOF
basicConstraints = CA:FALSE
keyUsage = critical, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid
subjectAltName = DNS:secure.example.com
EOF

openssl x509 -req \
    -in strong-cert.csr \
    -CA intermediate-ca.pem \
    -CAkey intermediate-ca.key \
    -CAcreateserial \
    -out strong-cert.pem \
    -days 365 \
    -sha384 \
    -extfile strong-cert.ext \
    2>/dev/null

log_success "Strong certificate created: strong-cert.pem"

#############################################################################
# 5. Generate ECC Certificate (P-256)
#############################################################################
log_info "Generating ECC certificate (P-256)..."

openssl ecparam -name prime256v1 -genkey -noout -out ecc-cert.key 2>/dev/null

openssl req -new \
    -key ecc-cert.key \
    -out ecc-cert.csr \
    -subj "/C=US/ST=California/L=San Francisco/O=Modern Organization/CN=ecc.example.com" \
    2>/dev/null

cat > ecc-cert.ext << EOF
basicConstraints = CA:FALSE
keyUsage = critical, digitalSignature
extendedKeyUsage = serverAuth
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid
subjectAltName = DNS:ecc.example.com
EOF

openssl x509 -req \
    -in ecc-cert.csr \
    -CA intermediate-ca.pem \
    -CAkey intermediate-ca.key \
    -CAcreateserial \
    -out ecc-cert.pem \
    -days 365 \
    -sha256 \
    -extfile ecc-cert.ext \
    2>/dev/null

log_success "ECC certificate created: ecc-cert.pem"

#############################################################################
# 6. Generate Expired Certificate
#############################################################################
log_info "Generating expired certificate..."

openssl genrsa -out expired-cert.key 2048 2>/dev/null

openssl req -new \
    -key expired-cert.key \
    -out expired-cert.csr \
    -subj "/C=US/ST=California/L=San Francisco/O=Expired Org/CN=expired.example.com" \
    2>/dev/null

cat > expired-cert.ext << EOF
basicConstraints = CA:FALSE
keyUsage = critical, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid
subjectAltName = DNS:expired.example.com
EOF

# Create certificate that's already expired (start 30 days ago, valid for 1 day)
openssl x509 -req \
    -in expired-cert.csr \
    -CA intermediate-ca.pem \
    -CAkey intermediate-ca.key \
    -CAcreateserial \
    -out expired-cert.pem \
    -days 1 \
    -sha256 \
    -extfile expired-cert.ext \
    2>/dev/null

# Backdate the certificate using faketime if available, otherwise note it
if command -v faketime &> /dev/null; then
    faketime -f '-30d' openssl x509 -req \
        -in expired-cert.csr \
        -CA intermediate-ca.pem \
        -CAkey intermediate-ca.key \
        -CAcreateserial \
        -out expired-cert.pem \
        -days 1 \
        -sha256 \
        -extfile expired-cert.ext \
        2>/dev/null
    log_success "Expired certificate created: expired-cert.pem (backdated)"
else
    log_warning "Expired certificate created but not backdated (install faketime for true expired cert)"
    # Create a note file
    echo "Note: This certificate was created with 1-day validity." > expired-cert.note
    echo "To create a truly expired cert, run this script with 'faketime' installed." >> expired-cert.note
fi

#############################################################################
# 7. Generate Long Validity Certificate (Non-compliant)
#############################################################################
log_info "Generating long validity certificate (non-compliant)..."

openssl genrsa -out long-validity-cert.key 2048 2>/dev/null

openssl req -new \
    -key long-validity-cert.key \
    -out long-validity-cert.csr \
    -subj "/C=US/ST=California/L=San Francisco/O=Legacy Org/CN=legacy.example.com" \
    2>/dev/null

cat > long-validity.ext << EOF
basicConstraints = CA:FALSE
keyUsage = critical, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid
subjectAltName = DNS:legacy.example.com
EOF

# 825 days - exceeds CA/B Forum 398-day limit
openssl x509 -req \
    -in long-validity-cert.csr \
    -CA intermediate-ca.pem \
    -CAkey intermediate-ca.key \
    -CAcreateserial \
    -out long-validity-cert.pem \
    -days 825 \
    -sha256 \
    -extfile long-validity.ext \
    2>/dev/null

log_success "Long validity certificate created: long-validity-cert.pem (825 days)"

#############################################################################
# 8. Generate Weak Certificate (RSA 1024, SHA-1) - DEPRECATED
#############################################################################
log_info "Generating weak certificate (RSA 1024, SHA-1)..."

openssl genrsa -out weak-cert.key 1024 2>/dev/null

openssl req -new \
    -key weak-cert.key \
    -out weak-cert.csr \
    -subj "/C=US/ST=California/L=San Francisco/O=Insecure Org/CN=weak.example.com" \
    2>/dev/null

cat > weak-cert.ext << EOF
basicConstraints = CA:FALSE
keyUsage = critical, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid
subjectAltName = DNS:weak.example.com
EOF

openssl x509 -req \
    -in weak-cert.csr \
    -CA intermediate-ca.pem \
    -CAkey intermediate-ca.key \
    -CAcreateserial \
    -out weak-cert.pem \
    -days 365 \
    -sha1 \
    -extfile weak-cert.ext \
    2>/dev/null

log_success "Weak certificate created: weak-cert.pem (RSA 1024, SHA-1)"

#############################################################################
# 9. Generate Self-Signed Certificate
#############################################################################
log_info "Generating self-signed certificate..."

openssl genrsa -out self-signed.key 2048 2>/dev/null

# Create config for self-signed with SAN (LibreSSL compatible)
cat > self-signed.cnf << EOF
[req]
distinguished_name = req_distinguished_name
x509_extensions = v3_req
prompt = no

[req_distinguished_name]
C = US
ST = California
L = San Francisco
O = Self Signed Org
CN = selfsigned.example.com

[v3_req]
subjectAltName = DNS:selfsigned.example.com
EOF

openssl req -x509 -new -nodes \
    -key self-signed.key \
    -sha256 \
    -days 365 \
    -out self-signed.pem \
    -config self-signed.cnf \
    -extensions v3_req \
    2>/dev/null

log_success "Self-signed certificate created: self-signed.pem"

#############################################################################
# 10. Generate Certificate Chain Bundle
#############################################################################
log_info "Creating certificate chain bundle..."

cat valid-cert.pem intermediate-ca.pem root-ca.pem > chain-bundle.pem

log_success "Chain bundle created: chain-bundle.pem"

#############################################################################
# 11. Convert to DER Format
#############################################################################
log_info "Converting certificates to DER format..."

openssl x509 -in valid-cert.pem -outform DER -out valid-cert.der 2>/dev/null
openssl x509 -in root-ca.pem -outform DER -out root-ca.der 2>/dev/null
openssl x509 -in ecc-cert.pem -outform DER -out ecc-cert.der 2>/dev/null

log_success "DER certificates created: valid-cert.der, root-ca.der, ecc-cert.der"

#############################################################################
# 12. Create PKCS7 Bundle
#############################################################################
log_info "Creating PKCS7 bundle..."

openssl crl2pkcs7 -nocrl \
    -certfile valid-cert.pem \
    -certfile intermediate-ca.pem \
    -certfile root-ca.pem \
    -out chain-bundle.p7b \
    2>/dev/null

# Also create DER-encoded PKCS7
openssl crl2pkcs7 -nocrl \
    -certfile valid-cert.pem \
    -certfile intermediate-ca.pem \
    -out chain-bundle-der.p7b \
    -outform DER \
    2>/dev/null

log_success "PKCS7 bundles created: chain-bundle.p7b, chain-bundle-der.p7b"

#############################################################################
# 13. Create PKCS12 Keystores
#############################################################################
log_info "Creating PKCS12 keystores..."

# Password for PKCS12 files
PKCS12_PASSWORD="testpassword123"

# Valid certificate with chain
openssl pkcs12 -export \
    -in valid-cert.pem \
    -inkey valid-cert.key \
    -certfile intermediate-ca.pem \
    -out valid-cert.p12 \
    -name "valid-cert" \
    -passout pass:$PKCS12_PASSWORD \
    2>/dev/null

# ECC certificate
openssl pkcs12 -export \
    -in ecc-cert.pem \
    -inkey ecc-cert.key \
    -out ecc-cert.p12 \
    -name "ecc-cert" \
    -passout pass:$PKCS12_PASSWORD \
    2>/dev/null

# Strong certificate with full chain (include CA certs without -chain flag for LibreSSL)
cat intermediate-ca.pem root-ca.pem > ca-bundle.pem
openssl pkcs12 -export \
    -in strong-cert.pem \
    -inkey strong-cert.key \
    -certfile ca-bundle.pem \
    -out strong-cert-chain.p12 \
    -name "strong-cert-chain" \
    -passout pass:$PKCS12_PASSWORD \
    2>/dev/null
rm -f ca-bundle.pem

# Also create a .pfx (same format, different extension)
cp valid-cert.p12 valid-cert.pfx

log_success "PKCS12 keystores created: valid-cert.p12, ecc-cert.p12, strong-cert-chain.p12"

#############################################################################
# 14. Create Password File
#############################################################################
log_info "Creating password file..."

echo "$PKCS12_PASSWORD" > keystore-password.txt
chmod 600 keystore-password.txt

log_success "Password file created: keystore-password.txt"

#############################################################################
# 15. Create Certificate with Multiple SANs
#############################################################################
log_info "Generating multi-SAN certificate..."

openssl genrsa -out multi-san.key 2048 2>/dev/null

openssl req -new \
    -key multi-san.key \
    -out multi-san.csr \
    -subj "/C=US/ST=California/L=San Francisco/O=Multi Domain Org/CN=primary.example.com" \
    2>/dev/null

cat > multi-san.ext << EOF
basicConstraints = CA:FALSE
keyUsage = critical, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid
subjectAltName = @alt_names

[alt_names]
DNS.1 = primary.example.com
DNS.2 = secondary.example.com
DNS.3 = tertiary.example.com
DNS.4 = *.wildcard.example.com
DNS.5 = api.example.com
DNS.6 = www.example.com
DNS.7 = mail.example.com
DNS.8 = ftp.example.com
IP.1 = 10.0.0.1
IP.2 = 10.0.0.2
IP.3 = 192.168.1.100
email.1 = admin@example.com
URI.1 = https://example.com
EOF

openssl x509 -req \
    -in multi-san.csr \
    -CA intermediate-ca.pem \
    -CAkey intermediate-ca.key \
    -CAcreateserial \
    -out multi-san.pem \
    -days 365 \
    -sha256 \
    -extfile multi-san.ext \
    2>/dev/null

log_success "Multi-SAN certificate created: multi-san.pem"

#############################################################################
# 16. Create Expiring Soon Certificate (7 days)
#############################################################################
log_info "Generating expiring-soon certificate (7 days)..."

openssl genrsa -out expiring-soon.key 2048 2>/dev/null

openssl req -new \
    -key expiring-soon.key \
    -out expiring-soon.csr \
    -subj "/C=US/ST=California/L=San Francisco/O=Expiring Org/CN=expiring.example.com" \
    2>/dev/null

cat > expiring-soon.ext << EOF
basicConstraints = CA:FALSE
keyUsage = critical, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid
subjectAltName = DNS:expiring.example.com
EOF

openssl x509 -req \
    -in expiring-soon.csr \
    -CA intermediate-ca.pem \
    -CAkey intermediate-ca.key \
    -CAcreateserial \
    -out expiring-soon.pem \
    -days 7 \
    -sha256 \
    -extfile expiring-soon.ext \
    2>/dev/null

log_success "Expiring-soon certificate created: expiring-soon.pem (7 days)"

#############################################################################
# 17. Create Expiring Warning Certificate (30 days)
#############################################################################
log_info "Generating expiring-warning certificate (30 days)..."

openssl genrsa -out expiring-warning.key 2048 2>/dev/null

openssl req -new \
    -key expiring-warning.key \
    -out expiring-warning.csr \
    -subj "/C=US/ST=California/L=San Francisco/O=Warning Org/CN=warning.example.com" \
    2>/dev/null

cat > expiring-warning.ext << EOF
basicConstraints = CA:FALSE
keyUsage = critical, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid
subjectAltName = DNS:warning.example.com
EOF

openssl x509 -req \
    -in expiring-warning.csr \
    -CA intermediate-ca.pem \
    -CAkey intermediate-ca.key \
    -CAcreateserial \
    -out expiring-warning.pem \
    -days 30 \
    -sha256 \
    -extfile expiring-warning.ext \
    2>/dev/null

log_success "Expiring-warning certificate created: expiring-warning.pem (30 days)"

#############################################################################
# 18. Create Client Authentication Certificate
#############################################################################
log_info "Generating client authentication certificate..."

openssl genrsa -out client-auth.key 2048 2>/dev/null

openssl req -new \
    -key client-auth.key \
    -out client-auth.csr \
    -subj "/C=US/ST=California/L=San Francisco/O=Client Org/CN=client@example.com" \
    2>/dev/null

cat > client-auth.ext << EOF
basicConstraints = CA:FALSE
keyUsage = critical, digitalSignature
extendedKeyUsage = clientAuth, emailProtection
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid
subjectAltName = email:client@example.com
EOF

openssl x509 -req \
    -in client-auth.csr \
    -CA intermediate-ca.pem \
    -CAkey intermediate-ca.key \
    -CAcreateserial \
    -out client-auth.pem \
    -days 365 \
    -sha256 \
    -extfile client-auth.ext \
    2>/dev/null

log_success "Client auth certificate created: client-auth.pem"

#############################################################################
# 19. Create Code Signing Certificate
#############################################################################
log_info "Generating code signing certificate..."

openssl genrsa -out code-signing.key 2048 2>/dev/null

openssl req -new \
    -key code-signing.key \
    -out code-signing.csr \
    -subj "/C=US/ST=California/L=San Francisco/O=Software Org/CN=Code Signing" \
    2>/dev/null

cat > code-signing.ext << EOF
basicConstraints = CA:FALSE
keyUsage = critical, digitalSignature
extendedKeyUsage = codeSigning
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid
EOF

openssl x509 -req \
    -in code-signing.csr \
    -CA intermediate-ca.pem \
    -CAkey intermediate-ca.key \
    -CAcreateserial \
    -out code-signing.pem \
    -days 365 \
    -sha256 \
    -extfile code-signing.ext \
    2>/dev/null

log_success "Code signing certificate created: code-signing.pem"

#############################################################################
# Cleanup temporary files
#############################################################################
log_info "Cleaning up temporary files..."

rm -f *.csr *.ext *.srl *.cnf

log_success "Cleanup complete"

#############################################################################
# Generate manifest file
#############################################################################
log_info "Generating manifest..."

cat > MANIFEST.md << 'EOF'
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

EOF

log_success "Manifest created: MANIFEST.md"

#############################################################################
# Summary
#############################################################################
echo ""
echo "========================================="
echo "  Certificate Generation Complete"
echo "========================================="
echo ""
echo "Generated certificates:"
echo "  - CA certificates: 2 (root + intermediate)"
echo "  - End-entity PEM: 12"
echo "  - DER format: 3"
echo "  - PKCS7 bundles: 2"
echo "  - PKCS12 keystores: 4"
echo ""
echo "Total files: $(ls -1 | wc -l | tr -d ' ')"
echo ""
echo "PKCS12 password: $PKCS12_PASSWORD"
echo "(Also stored in keystore-password.txt)"
echo ""
echo "See MANIFEST.md for full documentation"
echo "========================================="
