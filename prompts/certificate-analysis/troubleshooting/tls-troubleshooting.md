<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# TLS Troubleshooting Prompt

## Purpose

Diagnose and resolve TLS connection issues, certificate errors, and configuration problems.

## Usage

### General TLS Troubleshooting

```
Troubleshoot TLS connection issues for [domain]:

Symptoms:
[Describe the error or behavior]

Diagnose:
1. Connection establishment
2. Certificate validity
3. Chain completeness
4. Protocol/cipher compatibility
5. SNI configuration
6. OCSP/revocation status

Provide specific fixes for identified issues.
```

### Using Certificate Analyser

```bash
# Full diagnostic
./utils/certificate-analyser/cert-analyser.sh --all-checks example.com

# StartTLS for mail servers
./utils/certificate-analyser/cert-analyser.sh --starttls smtp mail.example.com

# Verbose output for debugging
./utils/certificate-analyser/cert-analyser.sh --verbose example.com
```

## Common Error Patterns

### Certificate Not Trusted

**Error Messages**:
- "certificate verify failed"
- "unable to get local issuer certificate"
- "self signed certificate in certificate chain"
- "certificate has expired"

**Diagnosis**:
```bash
# Check certificate chain
openssl s_client -connect example.com:443 -showcerts

# Verify against system store
openssl s_client -connect example.com:443 -verify_return_error
```

**Common Causes**:
1. Missing intermediate certificate
2. Self-signed certificate
3. Expired certificate
4. Root CA not in trust store

### Hostname Mismatch

**Error Messages**:
- "hostname mismatch"
- "certificate is not valid for this name"
- "SSL_ERROR_BAD_CERT_DOMAIN"

**Diagnosis**:
```bash
# Check certificate SAN entries
openssl s_client -connect example.com:443 2>/dev/null | \
    openssl x509 -noout -text | grep -A1 "Subject Alternative Name"

# Check Common Name
openssl s_client -connect example.com:443 2>/dev/null | \
    openssl x509 -noout -subject
```

**Common Causes**:
1. Wrong certificate deployed
2. Missing SAN entry
3. Accessing via IP instead of hostname
4. Wildcard doesn't cover subdomain level

### Connection Refused/Timeout

**Error Messages**:
- "Connection refused"
- "Connection timed out"
- "no peer certificate available"

**Diagnosis**:
```bash
# Test connectivity
nc -zv example.com 443

# Test with openssl
openssl s_client -connect example.com:443 -servername example.com

# Check if service is listening
curl -v https://example.com 2>&1 | head -20
```

**Common Causes**:
1. Firewall blocking port 443
2. Service not running
3. Wrong port configuration
4. DNS resolution issues

### Protocol/Cipher Mismatch

**Error Messages**:
- "no shared cipher"
- "wrong version number"
- "unsupported protocol"
- "sslv3 alert handshake failure"

**Diagnosis**:
```bash
# Test specific TLS versions
openssl s_client -connect example.com:443 -tls1_2
openssl s_client -connect example.com:443 -tls1_3

# List supported ciphers
nmap --script ssl-enum-ciphers -p 443 example.com
```

**Common Causes**:
1. TLS 1.0/1.1 disabled on server
2. Client doesn't support TLS 1.2/1.3
3. No common cipher suites
4. Server requires specific ciphers

## Troubleshooting Workflows

### Certificate Error Workflow

```
TLS certificate error for [domain]:
Error: [exact error message]

Step through diagnosis:
1. Is the certificate expired?
2. Does hostname match SAN/CN?
3. Is the chain complete?
4. Is the CA trusted?
5. Is the certificate revoked?

Provide the specific fix for the root cause.
```

### StartTLS Workflow

```
Troubleshoot StartTLS for [protocol] on [hostname:port]:

Protocol: [SMTP/IMAP/POP3/LDAP/FTP]
Port: [port number]

Check:
1. Does server advertise StartTLS?
2. Can StartTLS handshake complete?
3. Certificate valid for hostname?
4. Protocol-specific requirements?
```

### Mixed Content / HTTPS Issues

```
HTTPS issues for [website]:

Symptoms:
[Describe issues - mixed content, redirect loops, etc.]

Diagnose:
1. Certificate valid for all domains?
2. HSTS configured correctly?
3. Redirect configuration?
4. Mixed content sources?
```

## Diagnostic Commands

### Basic Connection Test

```bash
# Full handshake with verbose output
openssl s_client -connect example.com:443 -servername example.com -state -debug

# Show certificate chain
openssl s_client -connect example.com:443 -showcerts

# Check certificate details
echo | openssl s_client -connect example.com:443 2>/dev/null | \
    openssl x509 -noout -text

# Quick validity check
echo | openssl s_client -connect example.com:443 2>/dev/null | \
    openssl x509 -noout -dates
```

### Protocol Testing

```bash
# Test TLS 1.2
openssl s_client -connect example.com:443 -tls1_2

# Test TLS 1.3
openssl s_client -connect example.com:443 -tls1_3

# Test specific cipher
openssl s_client -connect example.com:443 -cipher ECDHE-RSA-AES256-GCM-SHA384
```

### StartTLS Testing

```bash
# SMTP
openssl s_client -connect mail.example.com:587 -starttls smtp

# IMAP
openssl s_client -connect mail.example.com:143 -starttls imap

# POP3
openssl s_client -connect mail.example.com:110 -starttls pop3

# LDAP
openssl s_client -connect ldap.example.com:389 -starttls ldap

# FTP
openssl s_client -connect ftp.example.com:21 -starttls ftp
```

### OCSP Testing

```bash
# Check OCSP stapling
openssl s_client -connect example.com:443 -status

# Manual OCSP check
openssl ocsp -issuer issuer.pem -cert cert.pem -url http://ocsp.example.com
```

## Error Resolution Quick Reference

| Error | Likely Cause | Fix |
|-------|--------------|-----|
| Certificate expired | Validity period ended | Renew certificate |
| Hostname mismatch | Wrong cert or missing SAN | Deploy correct cert |
| Self-signed | No CA signature | Use CA-signed cert |
| Unknown CA | Missing chain or untrusted | Add intermediates |
| Revoked | Certificate revoked | Get new certificate |
| Handshake failure | Protocol mismatch | Update TLS config |
| No shared cipher | Cipher mismatch | Update cipher list |
| Connection refused | Port blocked/service down | Check firewall/service |

## Variations

### Client-Side Issues

```
Troubleshoot TLS from client perspective:

Client: [browser/curl/application]
Error: [error message]
Server: [domain]

Diagnose:
1. Client TLS capabilities
2. Trust store status
3. Proxy interference
4. Certificate pinning issues
```

### Load Balancer/CDN Issues

```
TLS issues behind [load balancer/CDN]:

Symptoms:
[Describe intermittent or proxy-related issues]

Check:
1. Origin certificate vs edge certificate
2. SSL termination point
3. Backend TLS requirements
4. Certificate propagation
```

### Internal/Private CA Issues

```
Troubleshoot internal PKI certificate:

Server: [internal hostname]
Internal CA: [CA name]

Issues to check:
1. Client trusts internal CA?
2. Chain includes internal intermediate?
3. CRL/OCSP accessible internally?
4. Name constraints followed?
```

## Debugging Tips

### Increase Verbosity

```bash
# Maximum OpenSSL verbosity
openssl s_client -connect example.com:443 -state -debug -msg

# curl verbose mode
curl -v https://example.com

# With certificate details
curl -v --cert-status https://example.com
```

### Capture Traffic

```bash
# Capture TLS handshake with tcpdump
tcpdump -i eth0 -w tls.pcap port 443

# Analyze with Wireshark
# Set up SSLKEYLOGFILE for decryption
export SSLKEYLOGFILE=/tmp/keys.log
curl https://example.com
```

### Check System Time

Certificate validation requires accurate system time:
```bash
# Check system time
date

# NTP sync status
timedatectl status
```

## Related Prompts

- [security-audit.md](../security/security-audit.md) - Full security audit
- [chain-validation.md](../security/chain-validation.md) - Chain validation
- [cab-forum-compliance.md](../compliance/cab-forum-compliance.md) - Compliance check

## Related RAG

- [TLS Security](../../../rag/certificate-analysis/tls-security/best-practices.md) - TLS best practices
- [Certificate Formats](../../../rag/certificate-analysis/formats/certificate-formats.md) - Format issues
- [Revocation](../../../rag/certificate-analysis/revocation/ocsp-crl.md) - OCSP/CRL issues
