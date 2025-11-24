# Certificate Analysis Report

**Domain**: github.com  
**Analysis Date**: 2025-11-24 17:58:41 UTC  
**Analyst**: Certificate Security Analyser v1.0

---

## Executive Summary

This report analyzes the complete certificate chain for **github.com**, comprising        3 certificate(s). 
The analysis focuses on compliance with current CA/Browser Forum policies, cryptographic strength, 
and operational security posture.

**Key Findings:**
- Certificate expires in **-20416 days** ()
- Validity period compliance: **✓ Compliant**
- Chain validation and detailed security analysis below

🚨 **CRITICAL**: Certificate expires within 7 days. Immediate renewal required.

---

## Certificate Chain Overview

### End-Entity Certificate (Leaf)

- **Subject**:  CN=github.com
- **Issuer**:  CN=Sectigo ECC Domain Validation Secure Server CA,O=Sectigo Limited,L=Salford,ST=Greater Manchester,C=GB
- **Serial Number**: AB6686B5627BE80596821330128649F5
- **Valid From**: notBefore=Feb  5 00:00:00 2025 GMT
- **Valid Until**: notAfter=Feb  5 23:59:59 2026 GMT
- **Days Remaining**: -20416 days ❌
- **Total Validity Period**: 0 days
- **Signature Algorithm**: ecdsa-with-SHA256
- **Public Key**: Public-Key: (256 bit)

**Subject Alternative Names**:
```
None
```

### Intermediate Certificate 1

- **Subject**:  CN=Sectigo ECC Domain Validation Secure Server CA,O=Sectigo Limited,L=Salford,ST=Greater Manchester,C=GB
- **Issuer**:  CN=USERTrust ECC Certification Authority,O=The USERTRUST Network,L=Jersey City,ST=New Jersey,C=US
- **Serial Number**: F3644E6B6E0050237E0946BD7BE1F51D
- **Valid From**: notBefore=Nov  2 00:00:00 2018 GMT
- **Valid Until**: notAfter=Dec 31 23:59:59 2030 GMT
- **Days Remaining**: -20416 days ❌
- **Total Validity Period**: 0 days
- **Signature Algorithm**: ecdsa-with-SHA384
- **Public Key**: Public-Key: (256 bit)

---

## Detailed Analysis

### 1. Validity Period Compliance

| Metric | Value | Status |
|--------|-------|--------|
| Total Validity Period | 0 days | ✓ Compliant |
| Current Policy Limit | 398 days | Reference |
| Future Expected Limit | ~90-180 days | Planning |

**Analysis**: Certificate complies with current CA/Browser Forum policy (max 398 days). Industry is moving toward shorter validity periods (90-180 days expected by 2027) to improve security posture.

### 2. Cryptographic Strength

**Signature Algorithm**: ecdsa-with-SHA256
- ⚠️ Unknown algorithm

**Public Key**: Public-Key: (256 bit)
- ℹ️ Public-Key: (256 bit)

**Analysis**: While currently acceptable, consider upgrading to stronger algorithms (RSA 3072+ or ECC P-256+) during next renewal.

### 3. Subject Alternative Names (SANs)

**Covered Domains**:
```
None
```

**Analysis**: Certificate should cover all domains and subdomains that will use it. Wildcard certificates (*.example.com) provide flexibility but should be managed carefully.

### 4. Certificate Transparency

**CT Log Status**: ⚠️ Not detected

**Analysis**: Certificate Transparency has been mandatory since April 2018. CT logs provide public, append-only records of certificates, enabling detection of mis-issuance.

### 5. Revocation Checking

**OCSP**: ✓ http://ocsp.sectigo.com
**CRL**: ⚠️ Not available

**Analysis**: OCSP (Online Certificate Status Protocol) provides real-time revocation checking. Consider enabling OCSP stapling for improved performance and privacy.

### 6. Chain Validation

**Status**: ⚠️ Issues detected (may be due to missing root CA in local trust store)

---

## Risk Assessment

### Critical Issues ❌

1. **Imminent Expiration**: Certificate expires in -20416 days

### Warnings ⚠️

*No warnings.*

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
            ab:66:86:b5:62:7b:e8:05:96:82:13:30:12:86:49:f5
    Signature Algorithm: ecdsa-with-SHA256
        Issuer: C=GB, ST=Greater Manchester, L=Salford, O=Sectigo Limited, CN=Sectigo ECC Domain Validation Secure Server CA
        Validity
            Not Before: Feb  5 00:00:00 2025 GMT
            Not After : Feb  5 23:59:59 2026 GMT
        Subject: CN=github.com
        Subject Public Key Info:
            Public Key Algorithm: id-ecPublicKey
                Public-Key: (256 bit)
                pub: 
                    04:20:34:5c:46:ff:2c:cb:f8:24:9a:ae:f0:bb:2f:
                    77:a9:1f:97:21:36:71:ba:c2:26:18:c5:1e:43:fd:
                    9d:49:e0:cc:46:9c:85:fc:29:b4:f9:7c:28:0b:a3:
                    2c:c7:5c:bf:6f:e7:46:dd:04:8a:ba:cb:80:2d:37:
                    88:0d:ee:06:d6
                ASN1 OID: prime256v1
                NIST CURVE: P-256
        X509v3 extensions:
            X509v3 Authority Key Identifier: 
                keyid:F6:85:0A:3B:11:86:E1:04:7D:0E:AA:0B:2C:D2:EE:CC:64:7B:7B:AE

            X509v3 Subject Key Identifier: 
                53:C8:7F:DE:9E:98:4E:C7:4D:D6:BC:DE:AB:95:3E:30:3D:3D:D1:C8
            X509v3 Key Usage: critical
                Digital Signature
            X509v3 Basic Constraints: critical
                CA:FALSE
            X509v3 Extended Key Usage: 
                TLS Web Server Authentication, TLS Web Client Authentication
            X509v3 Certificate Policies: 
                Policy: 1.3.6.1.4.1.6449.1.2.2.7
                  CPS: https://sectigo.com/CPS
                Policy: 2.23.140.1.2.1

            Authority Information Access: 
                CA Issuers - URI:http://crt.sectigo.com/SectigoECCDomainValidationSecureServerCA.crt
                OCSP - URI:http://ocsp.sectigo.com

            1.3.6.1.4.1.11129.2.4.2: 
                ...j.h.u...d.UX...C.h7.Bw..:....6nF.?.........k.K.....F0D. ;..>..#....9m..?N!..wt.7.....\`.. bQ.F.~J..
..~..`t...}@.r.h.-a.p..w.....(.o...ox*M....-r1...]pA-%L.......k.......H0F.!........1........t.d...r.....aW...!...$..}K.t..~....,8=.F.m......F...v..8...|..D_[....n..Y.G
i.......X......k.%.....G0E.!..b..f..SI!...x.%..t.iQ.N....K.... RB~.H6.9.. .Gu.N[k`..A.WK..m].'
            X509v3 Subject Alternative Name: 
                DNS:github.com, DNS:www.github.com
    Signature Algorithm: ecdsa-with-SHA256
         30:44:02:20:71:8c:a7:6e:c1:04:12:75:df:9e:a5:09:ed:96:
         63:2c:d8:22:9f:df:00:e3:50:33:70:24:78:4f:df:ca:6d:2c:
         02:20:6d:55:f3:77:62:02:19:fa:77:87:11:fc:1c:46:18:73:
         e2:e0:e9:73:c1:7e:b4:a9:ad:71:e5:89:4a:27:0c:90
```

### Certificate 2

```
Certificate:
    Data:
        Version: 3 (0x2)
        Serial Number:
            f3:64:4e:6b:6e:00:50:23:7e:09:46:bd:7b:e1:f5:1d
    Signature Algorithm: ecdsa-with-SHA384
        Issuer: C=US, ST=New Jersey, L=Jersey City, O=The USERTRUST Network, CN=USERTrust ECC Certification Authority
        Validity
            Not Before: Nov  2 00:00:00 2018 GMT
            Not After : Dec 31 23:59:59 2030 GMT
        Subject: C=GB, ST=Greater Manchester, L=Salford, O=Sectigo Limited, CN=Sectigo ECC Domain Validation Secure Server CA
        Subject Public Key Info:
            Public Key Algorithm: id-ecPublicKey
                Public-Key: (256 bit)
                pub: 
                    04:79:18:93:ca:9f:6d:9e:6c:57:00:23:05:37:0b:
                    5f:0f:58:5a:c4:de:7f:55:a3:e9:1e:d6:d9:25:0a:
                    88:a0:20:4a:1d:7a:4f:05:30:8a:63:49:13:8c:64:
                    21:07:95:fd:3a:35:e1:4a:ce:90:f0:18:f7:3d:af:
                    68:a6:fb:d4:48
                ASN1 OID: prime256v1
                NIST CURVE: P-256
        X509v3 extensions:
            X509v3 Authority Key Identifier: 
                keyid:3A:E1:09:86:D4:CF:19:C2:96:76:74:49:76:DC:E0:35:C6:63:63:9A

            X509v3 Subject Key Identifier: 
                F6:85:0A:3B:11:86:E1:04:7D:0E:AA:0B:2C:D2:EE:CC:64:7B:7B:AE
            X509v3 Key Usage: critical
                Digital Signature, Certificate Sign, CRL Sign
            X509v3 Basic Constraints: critical
                CA:TRUE, pathlen:0
            X509v3 Extended Key Usage: 
                TLS Web Server Authentication, TLS Web Client Authentication
            X509v3 Certificate Policies: 
                Policy: X509v3 Any Policy
                Policy: 2.23.140.1.2.1

            X509v3 CRL Distribution Points: 

                Full Name:
                  URI:http://crl.usertrust.com/USERTrustECCCertificationAuthority.crl

            Authority Information Access: 
                CA Issuers - URI:http://crt.usertrust.com/USERTrustECCAddTrustCA.crt
                OCSP - URI:http://ocsp.usertrust.com

    Signature Algorithm: ecdsa-with-SHA384
         30:65:02:30:4b:e7:c7:71:5c:b1:5c:09:6d:9a:42:60:5f:73:
         e9:f0:d6:26:d4:b5:51:54:6c:71:2d:1c:85:60:4d:28:f1:4d:
         a6:f0:ca:76:b7:4a:45:ef:a8:02:4a:f6:8d:4f:ae:6e:02:31:
         00:e0:e1:79:2a:f6:5e:17:00:ee:8c:fd:1e:67:9d:19:d3:21:
         96:b7:7d:e1:3a:0a:15:b6:65:fb:f3:a7:14:5c:ea:9e:f3:a1:
         72:31:ef:0a:51:02:11:07:0a:99:cf:1f:98
```

### Certificate 3

```
Certificate:
    Data:
        Version: 3 (0x2)
        Serial Number:
            56:67:1d:04:ea:4f:99:4c:6f:10:81:47:59:d2:75:94
    Signature Algorithm: sha384WithRSAEncryption
        Issuer: C=GB, ST=Greater Manchester, L=Salford, O=Comodo CA Limited, CN=AAA Certificate Services
        Validity
            Not Before: Mar 12 00:00:00 2019 GMT
            Not After : Dec 31 23:59:59 2028 GMT
        Subject: C=US, ST=New Jersey, L=Jersey City, O=The USERTRUST Network, CN=USERTrust ECC Certification Authority
        Subject Public Key Info:
            Public Key Algorithm: id-ecPublicKey
                Public-Key: (384 bit)
                pub: 
                    04:1a:ac:54:5a:a9:f9:68:23:e7:7a:d5:24:6f:53:
                    c6:5a:d8:4b:ab:c6:d5:b6:d1:e6:73:71:ae:dd:9c:
                    d6:0c:61:fd:db:a0:89:03:b8:05:14:ec:57:ce:ee:
                    5d:3f:e2:21:b3:ce:f7:d4:8a:79:e0:a3:83:7e:2d:
                    97:d0:61:c4:f1:99:dc:25:91:63:ab:7f:30:a3:b4:
                    70:e2:c7:a1:33:9c:f3:bf:2e:5c:53:b1:5f:b3:7d:
                    32:7f:8a:34:e3:79:79
                ASN1 OID: secp384r1
                NIST CURVE: P-384
        X509v3 extensions:
            X509v3 Authority Key Identifier: 
                keyid:A0:11:0A:23:3E:96:F1:07:EC:E2:AF:29:EF:82:A5:7F:D0:30:A4:B4

            X509v3 Subject Key Identifier: 
                3A:E1:09:86:D4:CF:19:C2:96:76:74:49:76:DC:E0:35:C6:63:63:9A
            X509v3 Key Usage: critical
                Digital Signature, Certificate Sign, CRL Sign
            X509v3 Basic Constraints: critical
                CA:TRUE
            X509v3 Certificate Policies: 
                Policy: X509v3 Any Policy

            X509v3 CRL Distribution Points: 

                Full Name:
                  URI:http://crl.comodoca.com/AAACertificateServices.crl

            Authority Information Access: 
                OCSP - URI:http://ocsp.comodoca.com

    Signature Algorithm: sha384WithRSAEncryption
         19:ec:eb:9d:89:2c:20:0b:04:80:1d:18:de:42:99:72:99:16:
         32:bd:0e:9c:75:5b:2c:15:e2:29:40:6d:ee:ff:72:db:db:ab:
         90:1f:8c:95:f2:8a:3d:08:72:42:89:50:07:e2:39:15:6c:01:
         87:d9:16:1a:f5:c0:75:2b:c5:e6:56:11:07:df:d8:98:bc:7c:
         9f:19:39:df:8b:ca:00:64:73:bc:46:10:9b:93:23:8d:be:16:
         c3:2e:08:82:9c:86:33:74:76:3b:28:4c:8d:03:42:85:b3:e2:
         b2:23:42:d5:1f:7a:75:6a:1a:d1:7c:aa:67:21:c4:33:3a:39:
         6d:53:c9:a2:ed:62:22:a8:bb:e2:55:6c:99:6c:43:6b:91:97:
         d1:0c:0b:93:02:1d:d2:bc:69:77:49:e6:1b:4d:f7:bf:14:78:
         03:b0:a6:ba:0b:b4:e1:85:7f:2f:dc:42:3b:ad:74:01:48:de:
         d6:6c:e1:19:98:09:5e:0a:b3:67:47:fe:1c:e0:d5:c1:28:ef:
         4a:8b:44:31:26:04:37:8d:89:74:36:2e:ef:a5:22:0f:83:74:
         49:92:c7:f7:10:c2:0c:29:fb:b7:bd:ba:7f:e3:5f:d5:9f:f2:
         a9:f4:74:d5:b8:e1:b3:b0:81:e4:e1:a5:63:a3:cc:ea:04:78:
         90:6e:bf:f7
```

---

## Methodology Notes

**Tools Used**: OpenSSL 3.3.6  
**Standards Referenced**: 
- CA/Browser Forum Baseline Requirements
- RFC 5280 (X.509 PKI Certificate and CRL Profile)
- RFC 6960 (OCSP)

**Analysis Date**: 2025-11-24 17:58:41 UTC  
**Script Version**: 1.0

---

*This analysis was performed using automated tooling and should be reviewed by qualified security personnel for production environments.*
