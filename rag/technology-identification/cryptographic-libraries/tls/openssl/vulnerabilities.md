# OpenSSL Vulnerabilities and Security Information

## Critical Vulnerabilities (Historical)

### Heartbleed (CVE-2014-0160)
- **Date**: April 2014
- **Severity**: CRITICAL (10.0 CVSS)
- **Affected Versions**: OpenSSL 1.0.1 through 1.0.1f
- **Fixed Versions**: 1.0.1g and later
- **Description**: Buffer over-read vulnerability allowing memory disclosure
- **Impact**: Private keys, passwords, and sensitive data could be leaked
- **Detection**: OpenSSL 1.0.1 through 1.0.1f

### POODLE (CVE-2014-3566)
- **Date**: October 2014
- **Severity**: HIGH
- **Affected**: SSLv3 protocol
- **Mitigation**: Disable SSLv3
- **Description**: Padding oracle attack on SSLv3
- **Impact**: Man-in-the-middle attacks possible

### DROWN (CVE-2016-0800)
- **Date**: March 2016
- **Severity**: HIGH (5.9 CVSS)
- **Affected Versions**: OpenSSL 1.0.2 before 1.0.2g, 1.0.1 before 1.0.1s
- **Fixed Versions**: 1.0.2g, 1.0.1s
- **Description**: Cross-protocol attack on TLS using SSLv2
- **Impact**: Decryption of TLS connections
- **Mitigation**: Disable SSLv2

### Sweet32 (CVE-2016-2183)
- **Date**: August 2016
- **Severity**: MEDIUM
- **Affected**: 64-bit block ciphers (3DES, Blowfish)
- **Mitigation**: Disable 64-bit block ciphers
- **Description**: Birthday attack on 64-bit block ciphers
- **Impact**: Plaintext recovery from long-lived connections

### Heartbleed-like (CVE-2016-6309)
- **Date**: September 2016
- **Severity**: HIGH
- **Affected Versions**: OpenSSL 1.1.0a
- **Fixed Versions**: 1.1.0b
- **Description**: Memory corruption in MDC2_Update()

### ROBOT (CVE-2017-13098)
- **Date**: December 2017
- **Severity**: MEDIUM
- **Description**: Return Of Bleichenbacher's Oracle Threat
- **Impact**: RSA decryption and signing

### Bleichenbacher's CAT (CVE-2022-4304)
- **Date**: February 2023
- **Severity**: MEDIUM (5.9 CVSS)
- **Affected Versions**: 1.0.2, 1.1.1, 3.0.0-3.0.7
- **Fixed Versions**: 1.0.2zg, 1.1.1t, 3.0.8
- **Description**: Timing Oracle in RSA Decryption

### Recent CVEs (2023-2025)

#### CVE-2023-0286
- **Date**: February 2023
- **Severity**: HIGH (7.4 CVSS)
- **Affected**: OpenSSL 3.0.0-3.0.8, 1.1.1-1.1.1s
- **Fixed**: 3.0.9, 1.1.1t
- **Description**: X.400 address type confusion

#### CVE-2023-2650
- **Date**: May 2023
- **Severity**: MEDIUM (6.5 CVSS)
- **Affected**: OpenSSL 3.0, 3.1
- **Description**: Possible DoS via certificate verification

#### CVE-2024-0727
- **Date**: January 2024
- **Severity**: MEDIUM (5.5 CVSS)
- **Affected**: OpenSSL 3.0, 3.1, 3.2
- **Description**: PKCS12 parsing NULL pointer dereference

## Version Support and EOL Dates

### OpenSSL 3.x

#### OpenSSL 3.2
- **Released**: November 2023
- **EOL**: TBD (supported)
- **Status**: Active development
- **LTS**: No

#### OpenSSL 3.1
- **Released**: March 2023
- **EOL**: March 2025 (1 year after 3.2)
- **Status**: Security fixes only
- **LTS**: No

#### OpenSSL 3.0 (LTS)
- **Released**: September 2021
- **EOL**: September 7, 2026
- **Status**: Long-term support
- **LTS**: Yes

### OpenSSL 1.1.1 (LTS - EOL)

#### OpenSSL 1.1.1
- **Released**: September 2018
- **EOL**: September 11, 2023 (END OF LIFE)
- **Final Version**: 1.1.1w
- **Status**: No longer supported
- **Security Risk**: HIGH - No security updates

### OpenSSL 1.1.0 (EOL)

#### OpenSSL 1.1.0
- **Released**: August 2016
- **EOL**: September 11, 2019 (END OF LIFE)
- **Final Version**: 1.1.0l
- **Status**: No longer supported
- **Security Risk**: CRITICAL - Multiple unpatched vulnerabilities

### OpenSSL 1.0.2 (LTS - EOL)

#### OpenSSL 1.0.2
- **Released**: January 2015
- **EOL**: December 31, 2019 (END OF LIFE)
- **Final Version**: 1.0.2u
- **Status**: No longer supported
- **Security Risk**: CRITICAL - Multiple unpatched vulnerabilities

### OpenSSL 1.0.1 (LTS - EOL)

#### OpenSSL 1.0.1
- **Released**: March 2012
- **EOL**: December 31, 2016 (END OF LIFE)
- **Final Version**: 1.0.1u
- **Status**: No longer supported
- **Security Risk**: CRITICAL - Heartbleed and many others

### OpenSSL 1.0.0 and Earlier (EOL)
- **Status**: END OF LIFE
- **Security Risk**: CRITICAL
- **Recommendation**: DO NOT USE

## Deprecated Features and Functions

### Deprecated in OpenSSL 3.0
- Low-level APIs (use EVP APIs instead)
- ENGINE API (use provider API)
- MD2, MD4, MDC2, RIPEMD160
- DES, RC2, RC4, RC5, CAST, Blowfish, IDEA
- SEED cipher
- Compression

### Deprecated Protocols
- SSLv2 (removed in 1.1.0)
- SSLv3 (disabled by default, POODLE)
- TLS 1.0 (deprecated, use TLS 1.2+)
- TLS 1.1 (deprecated, use TLS 1.2+)

### Insecure Algorithms
- MD5 (collision attacks)
- SHA-1 (collision attacks, deprecated for signatures)
- DES/3DES (small key size, Sweet32)
- RC4 (biases in keystream)
- Export ciphers (intentionally weak)

## Security Best Practices

### Version Selection
```
✅ RECOMMENDED:
- OpenSSL 3.0.x (LTS, supported until 2026)
- OpenSSL 3.2.x (latest stable)

⚠️ AVOID:
- OpenSSL 1.1.1 (EOL September 2023)

❌ DO NOT USE:
- OpenSSL 1.1.0 and earlier (EOL, critical vulnerabilities)
```

### TLS Configuration

#### Recommended Protocol Versions
```c
// Use TLS 1.2 and 1.3 only
SSL_CTX_set_min_proto_version(ctx, TLS1_2_VERSION);
SSL_CTX_set_max_proto_version(ctx, TLS1_3_VERSION);
```

#### Recommended Cipher Suites (TLS 1.3)
```
TLS_AES_256_GCM_SHA384
TLS_CHACHA20_POLY1305_SHA256
TLS_AES_128_GCM_SHA256
```

#### Recommended Cipher Suites (TLS 1.2)
```
ECDHE-RSA-AES256-GCM-SHA384
ECDHE-RSA-AES128-GCM-SHA256
ECDHE-RSA-CHACHA20-POLY1305
DHE-RSA-AES256-GCM-SHA384
DHE-RSA-AES128-GCM-SHA256
```

#### Cipher String
```
TLSv1.3:TLSv1.2:!SSLv3:!SSLv2:!MD5:!RC4:!DES:!3DES:!NULL:!EXPORT
```

### Certificate Validation
```c
// Always verify certificates
SSL_CTX_set_verify(ctx, SSL_VERIFY_PEER | SSL_VERIFY_FAIL_IF_NO_PEER_CERT, NULL);

// Set certificate verification depth
SSL_CTX_set_verify_depth(ctx, 4);

// Load trusted CA certificates
SSL_CTX_load_verify_locations(ctx, ca_file, ca_path);
```

### Key Size Recommendations
- **RSA**: Minimum 2048 bits (3072+ recommended)
- **ECDSA**: Minimum 256 bits (P-256, P-384, P-521)
- **DH**: Minimum 2048 bits (3072+ recommended)

## FIPS Mode

### FIPS 140-2/140-3
```c
// Enable FIPS mode (OpenSSL 1.0.2+)
FIPS_mode_set(1);

// Check if FIPS mode is enabled
if (FIPS_mode()) {
    // FIPS mode is active
}
```

### FIPS Validated Modules
- OpenSSL FIPS Object Module 2.0 (validated)
- OpenSSL 3.0 FIPS Provider (validation in progress)

## Detection Patterns

### Vulnerable Version Detection

#### Heartbleed (1.0.1 - 1.0.1f)
```bash
openssl version | grep -E "1\.0\.1[a-f]"
```

#### EOL Versions
```bash
# OpenSSL 1.1.1 (EOL September 2023)
openssl version | grep "1.1.1"

# OpenSSL 1.1.0 and earlier (EOL)
openssl version | grep -E "1\.[01]\."
```

### Configuration Issues
- SSLv2/SSLv3 enabled
- Weak ciphers enabled
- Certificate verification disabled
- Using deprecated APIs
- Using removed/deprecated ciphers

## Upgrade Recommendations

### From 1.1.1 to 3.0
- **Priority**: HIGH (1.1.1 is EOL)
- **Compatibility**: Some API changes
- **Action**: Test thoroughly, update deprecated API usage
- **Timeline**: Immediate

### From 3.0 to 3.2
- **Priority**: LOW (3.0 is LTS)
- **Compatibility**: Generally compatible
- **Action**: Opportunistic upgrade
- **Timeline**: Before 3.0 EOL (2026)

### From 1.0.x
- **Priority**: CRITICAL
- **Compatibility**: Major API changes
- **Action**: Complete rewrite may be needed
- **Timeline**: Immediate (already EOL)

## Compliance Requirements

### PCI DSS
- Requires TLS 1.2 or higher
- Prohibits SSLv3, TLS 1.0, TLS 1.1
- Regular security updates required

### HIPAA
- Requires encryption in transit
- Security updates required
- Risk assessment for cryptographic systems

### NIST Guidelines
- TLS 1.2 minimum (SP 800-52 Rev. 2)
- TLS 1.3 recommended
- Strong cipher suites required

## Detection Confidence

- **HIGH**: Version string match for known vulnerable versions
- **HIGH**: EOL version detection
- **MEDIUM**: Protocol/cipher configuration analysis
- **MEDIUM**: API usage patterns indicating deprecated features
- **LOW**: Generic SSL/TLS usage without version information
