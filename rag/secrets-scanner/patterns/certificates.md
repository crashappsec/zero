# Private Keys and Certificates

## Private Keys

### RSA Private Keys
```
Pattern: -----BEGIN RSA PRIVATE KEY-----
End: -----END RSA PRIVATE KEY-----
Severity: critical
```

Traditional RSA private key format (PKCS#1).

### Generic Private Keys
```
Pattern: -----BEGIN PRIVATE KEY-----
End: -----END PRIVATE KEY-----
Severity: critical
```

PKCS#8 format, can contain RSA, EC, or other key types.

### EC Private Keys
```
Pattern: -----BEGIN EC PRIVATE KEY-----
End: -----END EC PRIVATE KEY-----
Severity: critical
```

Elliptic Curve private keys.

### DSA Private Keys
```
Pattern: -----BEGIN DSA PRIVATE KEY-----
End: -----END DSA PRIVATE KEY-----
Severity: critical
```

### OpenSSH Private Keys
```
Pattern: -----BEGIN OPENSSH PRIVATE KEY-----
End: -----END OPENSSH PRIVATE KEY-----
Severity: critical
```

Modern OpenSSH format (since OpenSSH 7.8).

### Encrypted Private Keys
```
Pattern: -----BEGIN ENCRYPTED PRIVATE KEY-----
End: -----END ENCRYPTED PRIVATE KEY-----
Severity: high
```

Encrypted keys are lower severity but still sensitive.

### PuTTY Private Keys
```
Pattern: PuTTY-User-Key-File-[0-9]+:
Severity: critical
```

PuTTY PPK format.

---

## PGP/GPG Keys

### PGP Private Key Block
```
Pattern: -----BEGIN PGP PRIVATE KEY BLOCK-----
End: -----END PGP PRIVATE KEY BLOCK-----
Severity: critical
```

### PGP Secret Key
```
Pattern: -----BEGIN PGP SECRET KEY BLOCK-----
Severity: critical
```

Older format name.

---

## Certificates (Lower Severity)

### X.509 Certificates
```
Pattern: -----BEGIN CERTIFICATE-----
End: -----END CERTIFICATE-----
Severity: informational
```

Public certificates are not secrets, but may indicate key presence nearby.

### Certificate Requests
```
Pattern: -----BEGIN CERTIFICATE REQUEST-----
End: -----END CERTIFICATE REQUEST-----
Severity: informational
```

CSRs are not secret.

---

## PKCS#12 / PFX Files

### File Extensions
```
Pattern: \.p12$|\.pfx$
Severity: high
```

Binary format containing private key and certificate chain. Password-protected but often with weak passwords.

---

## SSH Keys

### Common File Names
```
id_rsa
id_dsa
id_ecdsa
id_ed25519
*.pem (may contain keys)
```

### SSH Public Keys (informational)
```
Pattern: ssh-(rsa|dsa|ecdsa|ed25519) AAAA[A-Za-z0-9+/]+
Severity: informational
```

Public keys are not secrets but may indicate private key presence.

---

## SSL/TLS Related

### Key Files
```
Pattern: \.key$ (file extension)
Context: Often contains private keys
Severity: high (investigate content)
```

### Combined Key+Cert Files
```
Pattern: \.pem$ (file extension)
Context: May contain private key and certificate
Severity: high (investigate content)
```

### JKS Keystores
```
Pattern: \.jks$ (file extension)
Context: Java KeyStore, password-protected
Severity: high
```

### Keystore Passwords
```
Pattern: storepass[=:]["']?[^\s"']+
Pattern: keypass[=:]["']?[^\s"']+
Severity: high
```

---

## Code Signing

### Apple Certificates
```
Pattern: \.p12$|\.mobileprovision$
Context: iOS/macOS code signing
Severity: critical
```

### Android Keystores
```
Pattern: \.keystore$|\.jks$
Context: Android app signing
Severity: critical
```

### Signing Key Passwords
```
Pattern: ANDROID_KEY_PASSWORD|KEYSTORE_PASSWORD
Context: CI/CD environment variables
Severity: critical
```

---

## AWS/Cloud Certificates

### AWS Certificate Manager
```
Pattern: arn:aws:acm:[a-z0-9-]+:[0-9]+:certificate/[a-f0-9-]+
Severity: informational
```

Reference to ACM cert, not the cert itself.

### Let's Encrypt
```
Pattern: /etc/letsencrypt/live/[^/]+/privkey\.pem
Severity: critical
```

Path to Let's Encrypt private key.

---

## Detection Notes

### Critical Files to Flag
- Any file containing `BEGIN.*PRIVATE KEY`
- `.pem` files in non-standard locations
- `.p12`, `.pfx`, `.key` files in repositories
- `id_rsa`, `id_ecdsa` without `.pub` extension

### Common Locations (should not be in repo)
- `~/.ssh/`
- `/etc/ssl/private/`
- `/etc/letsencrypt/`
- `./certs/`, `./keys/`, `./ssl/`

### False Positives
- Documentation about key formats
- Test certificates in test fixtures
- Public certificates (BEGIN CERTIFICATE)
- Empty placeholder files

### Immediate Actions Required
1. **Revoke/rotate** any exposed private key immediately
2. **Check git history** - key may be in old commits
3. **Update certificates** that used exposed keys
4. **Add to .gitignore** to prevent future commits
