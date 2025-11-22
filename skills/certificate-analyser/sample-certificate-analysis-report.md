<!--
Copyright (c) 2024 Gibson Powers Contributors

SPDX-License-Identifier: GPL-3.0
-->

# Certificate Analysis Report

**Domain**: example.com  
**Analysis Date**: 2024-11-19 14:30:22 UTC  
**Analyst**: Certificate Security Analyser v1.0

---

## Executive Summary

This report analyzes the complete certificate chain for **example.com**, comprising 3 certificate(s). 
The analysis focuses on compliance with current CA/Browser Forum policies, cryptographic strength, 
and operational security posture.

**Key Findings:**
- Certificate expires in **45 days** (2025-01-03)
- Validity period compliance: **✓ Compliant**
- Chain validation and detailed security analysis below

---

## Certificate Chain Overview

### End-Entity Certificate (Leaf)

- **Subject**: CN=example.com
- **Issuer**: CN=DigiCert TLS RSA SHA256 2020 CA1,O=DigiCert Inc,C=US
- **Serial Number**: 0F:A0:16:C4:46:F6:4E:07:8D:4F:9F:C4:8A:F3:D5:4C
- **Valid From**: Nov 20 00:00:00 2024 GMT
- **Valid Until**: Jan  3 23:59:59 2025 GMT
- **Days Remaining**: 45 days ✓
- **Total Validity Period**: 90 days
- **Signature Algorithm**: sha256WithRSAEncryption
- **Public Key**: RSA Public-Key: (2048 bit)

**Subject Alternative Names**:
```
DNS:example.com
DNS:www.example.com
```

### Intermediate Certificate 1

- **Subject**: CN=DigiCert TLS RSA SHA256 2020 CA1,O=DigiCert Inc,C=US
- **Issuer**: CN=DigiCert Global Root CA,OU=www.digicert.com,O=DigiCert Inc,C=US
- **Serial Number**: 06:D8:D9:04:D5:58:43:46:F6:8A:2F:A7:54:22:7E:C4
- **Valid From**: Sep 23 00:00:00 2020 GMT
- **Valid Until**: Sep 23 23:59:59 2030 GMT
- **Days Remaining**: 2134 days ✓
- **Total Validity Period**: 3653 days
- **Signature Algorithm**: sha256WithRSAEncryption
- **Public Key**: RSA Public-Key: (2048 bit)

### Root Certificate

- **Subject**: CN=DigiCert Global Root CA,OU=www.digicert.com,O=DigiCert Inc,C=US
- **Issuer**: CN=DigiCert Global Root CA,OU=www.digicert.com,O=DigiCert Inc,C=US
- **Serial Number**: 08:3B:E0:56:90:42:46:B1:A1:75:6A:C9:59:91:C7:4A
- **Valid From**: Nov 10 00:00:00 2006 GMT
- **Valid Until**: Nov 10 00:00:00 2031 GMT
- **Days Remaining**: 2548 days ✓
- **Total Validity Period**: 9131 days
- **Signature Algorithm**: sha1WithRSAEncryption
- **Public Key**: RSA Public-Key: (2048 bit)

---

## Detailed Analysis

### 1. Validity Period Compliance

| Metric | Value | Status |
|--------|-------|--------|
| Total Validity Period | 90 days | ✓ Compliant |
| Current Policy Limit | 398 days | Reference |
| Future Expected Limit | ~90-180 days | Planning |

**Analysis**: Certificate complies with current CA/Browser Forum policy (max 398 days). Industry is moving toward shorter validity periods (90-180 days expected by 2027) to improve security posture. This certificate is already well-positioned for future policy changes.

### 2. Cryptographic Strength

**Signature Algorithm**: sha256WithRSAEncryption
- ✓ SHA-256+ (compliant)

**Public Key**: RSA Public-Key: (2048 bit)
- ⚠️ RSA 2048 (minimum, consider upgrade)

**Analysis**: While currently acceptable, consider upgrading to stronger algorithms (RSA 3072+ or ECC P-256+) during next renewal. Modern clients support stronger cryptography, and the performance overhead is minimal with hardware acceleration.

### 3. Subject Alternative Names (SANs)

**Covered Domains**:
```
DNS:example.com
DNS:www.example.com
```

**Analysis**: Certificate covers both the apex domain and www subdomain. This is a common and appropriate configuration. Wildcard certificates (*.example.com) provide flexibility but should be managed carefully due to broader scope.

### 4. Certificate Transparency

**CT Log Status**: ✓ Present

**Analysis**: Certificate Transparency has been mandatory since April 2018. CT logs provide public, append-only records of certificates, enabling detection of mis-issuance. This certificate has been properly logged.

### 5. Revocation Checking

**OCSP**: ✓ http://ocsp.digicert.com
**CRL**: ✓ http://crl3.digicert.com/DigiCertGlobalRootCA.crl

**Analysis**: OCSP (Online Certificate Status Protocol) provides real-time revocation checking. Consider enabling OCSP stapling for improved performance and privacy. OCSP stapling allows the server to provide certificate status directly, reducing client lookup time and protecting privacy.

### 6. Chain Validation

**Status**: ✓ Valid

---

## Risk Assessment

### Critical Issues ❌

*No critical issues detected.*

### Warnings ⚠️

1. **Upcoming Expiration**: Certificate expires in 45 days - plan renewal
2. **Minimal Key Strength**: RSA 2048 is minimum acceptable - consider 3072+ or ECC

### Informational ℹ️

1. **Industry Trend**: Certificate lifespans are decreasing to improve security
2. **Automation**: Shorter certificates require automated renewal processes
3. **Best Practice**: Implement monitoring 30/14/7 days before expiration

---

## Recommendations

### Immediate Actions (0-30 days)

1. **Renew Certificate**: Expiration is imminent - initiate renewal process immediately
2. **Verify Domain Control**: Ensure domain validation methods (DNS, HTTP, email) are accessible
3. **Test Renewal Process**: Verify certificate deployment pipeline is functional

### Short-term Planning (30-90 days)

1. **Implement Automated Renewal**
   - Adopt ACME protocol (Let's Encrypt, commercial CAs supporting ACME)
   - Deploy automation tools (certbot, cert-manager, acme.sh)
   - Test automated renewal in staging environment

2. **Enable Certificate Monitoring**
   - Deploy certificate expiration monitoring (e.g., cert-manager, commercial services)
   - Set up alerting to multiple channels (email, Slack, PagerDuty)
   - Create runbooks for renewal procedures

3. **Security Enhancements**
   - Enable OCSP stapling on web servers
   - Implement Certificate Transparency monitoring
   - Review and update cipher suites

### Long-term Strategy (90+ days)

1. **Prepare for Shorter Certificates**
   - Future CA/Browser Forum policies will reduce validity to 90-180 days
   - Manual processes will not scale - automation is mandatory
   - Budget for automation tools and engineering time

2. **Infrastructure as Code**
   - Treat certificate management as code (Terraform, CloudFormation, Kubernetes)
   - Version control certificate configurations
   - Implement CI/CD for certificate rotation

3. **Consider Modern Alternatives**
   - Evaluate ECC certificates (smaller, faster, equally secure)
   - Investigate managed certificate services (AWS ACM, Google-managed certs)
   - For Kubernetes: implement cert-manager with automatic rotation

4. **Certificate Lifecycle Management**
   - Establish certificate inventory and tracking system
   - Define ownership and responsibilities for certificate renewals
   - Create disaster recovery procedures for certificate issues
   - Regular security audits and compliance checks

---

## Appendix: Raw Certificate Data

### Certificate 1

```
Certificate:
    Data:
        Version: 3 (0x2)
        Serial Number:
            0f:a0:16:c4:46:f6:4e:07:8d:4f:9f:c4:8a:f3:d5:4c
        Signature Algorithm: sha256WithRSAEncryption
        Issuer: C=US, O=DigiCert Inc, CN=DigiCert TLS RSA SHA256 2020 CA1
        Validity
            Not Before: Nov 20 00:00:00 2024 GMT
            Not After : Jan  3 23:59:59 2025 GMT
        Subject: CN=example.com
        Subject Public Key Info:
            Public Key Algorithm: rsaEncryption
                RSA Public-Key: (2048 bit)
                Modulus:
                    [truncated for brevity]
                Exponent: 65537 (0x10001)
        X509v3 extensions:
            X509v3 Authority Key Identifier: 
                keyid:B7:6B:A2:EA:A8:AA:84:8C:79:EA:B4:DA:0F:98:B2:C5:95:76:B9:F4
            X509v3 Subject Key Identifier: 
                73:6C:75:88:AC:A6:D3:E9:FC:69:4B:04:FC:61:BE:AC:01:55:EE:37
            X509v3 Subject Alternative Name: 
                DNS:example.com, DNS:www.example.com
            X509v3 Key Usage: critical
                Digital Signature, Key Encipherment
            X509v3 Extended Key Usage: 
                TLS Web Server Authentication, TLS Web Client Authentication
            X509v3 CRL Distribution Points: 
                Full Name:
                  URI:http://crl3.digicert.com/DigiCertTLSRSASHA2562020CA1-4.crl
            X509v3 Certificate Policies: 
                Policy: 2.23.140.1.2.2
                  CPS: http://www.digicert.com/CPS
            Authority Information Access: 
                OCSP - URI:http://ocsp.digicert.com
                CA Issuers - URI:http://cacerts.digicert.com/DigiCertTLSRSASHA2562020CA1-1.crt
            X509v3 Basic Constraints: critical
                CA:FALSE
            CT Precertificate SCTs: 
                [SCT data omitted for brevity]
    Signature Algorithm: sha256WithRSAEncryption
         [Signature data omitted for brevity]
```

### Certificate 2

```
Certificate:
    Data:
        Version: 3 (0x2)
        Serial Number:
            06:d8:d9:04:d5:58:43:46:f6:8a:2f:a7:54:22:7e:c4
        Signature Algorithm: sha256WithRSAEncryption
        Issuer: C=US, O=DigiCert Inc, OU=www.digicert.com, CN=DigiCert Global Root CA
        Validity
            Not Before: Sep 23 00:00:00 2020 GMT
            Not After : Sep 23 23:59:59 2030 GMT
        Subject: C=US, O=DigiCert Inc, CN=DigiCert TLS RSA SHA256 2020 CA1
        [Additional details omitted for brevity]
```

### Certificate 3

```
Certificate:
    Data:
        Version: 3 (0x2)
        Serial Number:
            08:3b:e0:56:90:42:46:b1:a1:75:6a:c9:59:91:c7:4a
        Signature Algorithm: sha1WithRSAEncryption
        Issuer: C=US, O=DigiCert Inc, OU=www.digicert.com, CN=DigiCert Global Root CA
        Validity
            Not Before: Nov 10 00:00:00 2006 GMT
            Not After : Nov 10 00:00:00 2031 GMT
        Subject: C=US, O=DigiCert Inc, OU=www.digicert.com, CN=DigiCert Global Root CA
        [Additional details omitted for brevity]
```

---

## Methodology Notes

**Tools Used**: OpenSSL 3.0.2  
**Standards Referenced**: 
- CA/Browser Forum Baseline Requirements
- RFC 5280 (X.509 PKI Certificate and CRL Profile)
- RFC 6960 (OCSP)

**Analysis Date**: 2024-11-19 14:30:22 UTC  
**Script Version**: 1.0

---

*This analysis was performed using automated tooling and should be reviewed by qualified security personnel for production environments.*
