<!--
Copyright (c) 2024 Crash Override Inc
101 Fulton St, 416, New York 10038
SPDX-License-Identifier: GPL-3.0
-->

# Certificate Analyzer

**Status**: 🚧 Experimental - Not yet production-ready

Analyzes TLS/SSL certificates for security, validity, and best practices compliance.

## ⚠️ Development Status

This utility is in **early development** and is not yet ready for production use. It provides basic certificate analysis but lacks the comprehensive testing, documentation, and features of the production-ready supply chain analyzer.

### What Works
- ✅ TLS/SSL certificate validation
- ✅ Expiration checking and warnings
- ✅ Certificate chain analysis
- ✅ Security assessment
- ✅ AI-enhanced analysis with Claude

### What's Missing
- ❌ Configuration system integration
- ❌ Bulk domain scanning
- ❌ Certificate monitoring/alerting
- ❌ Output format options (JSON, markdown)
- ❌ Historical tracking
- ❌ Comprehensive testing
- ❌ Complete documentation

**Use at your own risk**. For production-grade analysis, use the [Supply Chain Security Analyzer](../supply-chain/).

## Overview

The Certificate Analyzer validates and analyzes SSL/TLS certificates to ensure:

- **Validity**: Certificate is not expired or invalid
- **Chain Trust**: Complete certificate chain verification
- **Encryption Strength**: Strong cipher suites and protocols
- **Best Practices**: Compliance with current security standards
- **Expiration Warnings**: Proactive expiration notifications

## Quick Start

### Prerequisites

```bash
# OpenSSL is required
openssl version

# Optional: For advanced features
brew install jq curl
```

### Basic Usage

```bash
# Analyze a domain
./cert-analyzer.sh example.com

# Analyze with port specification
./cert-analyzer.sh example.com 443

# AI-enhanced analysis
export ANTHROPIC_API_KEY="your-key"
./cert-analyzer-claude.sh example.com
```

## Available Scripts

### cert-analyzer.sh

Base analyzer that validates certificates using OpenSSL.

**Features**:
- Certificate validity checking
- Expiration date analysis
- Certificate chain validation
- Cipher suite evaluation
- Protocol version checking
- Common name and SAN validation

**Usage**:
```bash
# Analyze domain
./cert-analyzer.sh example.com

# Specify port
./cert-analyzer.sh example.com 8443

# Verbose output
./cert-analyzer.sh --verbose example.com
```

**Output**:
```
===================================
Certificate Analysis
===================================
Domain: example.com
Port: 443

Status: ✓ Valid
Issuer: Let's Encrypt
Valid From: 2024-09-01
Valid Until: 2024-12-01 (40 days remaining)
Common Name: example.com
Subject Alternative Names:
  - example.com
  - www.example.com

Certificate Chain: ✓ Valid (3 certificates)
Protocol: TLSv1.3
Cipher: TLS_AES_256_GCM_SHA384

Security Assessment: Good
```

### cert-analyzer-claude.sh

AI-enhanced analyzer with security insights and recommendations.

**Features**:
- All base analyzer features
- Security risk assessment
- Compliance checking
- Remediation recommendations
- Industry best practices comparison

**Requires**: `ANTHROPIC_API_KEY` environment variable

**Usage**:
```bash
export ANTHROPIC_API_KEY="your-key"
./cert-analyzer-claude.sh example.com
```

## Analysis Components

### Certificate Validation

Checks:
- Certificate is not expired
- Certificate is not revoked (OCSP)
- Certificate is trusted (chain validation)
- Valid date ranges
- Proper encoding and format

### Chain Analysis

Validates:
- Complete certificate chain
- Intermediate certificates present
- Root CA trust
- Chain ordering
- Cross-signing

### Security Assessment

Evaluates:
- Protocol version (TLS 1.2+)
- Cipher suite strength
- Key length (2048+ bits for RSA)
- Signature algorithm (SHA-256+)
- Forward secrecy support

### Expiration Warnings

Alerts when certificate expires in:
- ⚠️ Less than 30 days
- ⚠️⚠️ Less than 7 days
- 🚨 Expired

## Known Limitations

### Current Limitations

1. **Single Domain Only**: No bulk scanning or file input
2. **No Configuration System**: Cannot persist settings
3. **Limited Output Formats**: Text only
4. **No Monitoring**: No continuous checking or alerting
5. **Basic OCSP**: Limited revocation checking
6. **No Historical Data**: No tracking over time

### Analysis Limitations

- Relies on OpenSSL capabilities
- May not detect all vulnerabilities
- OCSP checking requires internet connectivity
- Some checks require external services
- May not work with non-standard ports

## Roadmap to Production

### Phase 1: Core Functionality (Current)
- [x] Basic certificate validation
- [x] Expiration checking
- [x] Chain validation
- [x] Security assessment
- [x] AI-enhanced analysis
- [ ] Comprehensive error handling

### Phase 2: Integration
- [ ] Hierarchical configuration system
- [ ] Bulk domain scanning
- [ ] File-based input (domains list)
- [ ] Output format options (JSON, markdown)

### Phase 3: Monitoring
- [ ] Continuous monitoring
- [ ] Expiration alerts
- [ ] Change detection
- [ ] Historical tracking
- [ ] Dashboard integration

### Phase 4: Production Ready
- [ ] Comprehensive testing
- [ ] Complete documentation
- [ ] CI/CD examples
- [ ] Performance optimization
- [ ] Enterprise features (LDAP, SAML)

## Development

### Architecture

```
certificate-analyzer/
├── cert-analyzer.sh              # Base analyzer
└── cert-analyzer-claude.sh       # AI-enhanced analyzer
```

### Adding Features

Priority development areas:

1. **Configuration Integration**: Add global config support
2. **Bulk Scanning**: Multiple domains, file input
3. **Output Formats**: JSON, markdown, CSV
4. **Monitoring**: Continuous checking and alerting
5. **Testing**: Comprehensive test suite
6. **Documentation**: Usage guide and examples

## Use Cases

### Certificate Expiration Monitoring
Monitor certificates and alert before expiration.

### Security Audit
Validate certificate security across infrastructure.

### Compliance Checking
Ensure certificates meet security standards (PCI-DSS, HIPAA, etc.).

### Incident Response
Quick certificate validation during security incidents.

### Migration Planning
Assess certificate status before infrastructure changes.

## Examples

### Example 1: Basic Check

```bash
./cert-analyzer.sh google.com
```

### Example 2: Custom Port

```bash
./cert-analyzer.sh internal.example.com 8443
```

### Example 3: Multiple Domains

```bash
for domain in example.com test.com demo.com; do
  echo "Checking $domain..."
  ./cert-analyzer.sh "$domain"
  echo ""
done
```

### Example 4: CI/CD Integration

```bash
#!/bin/bash
# Check production certificates
./cert-analyzer.sh prod.example.com | grep -q "Valid"
if [ $? -ne 0 ]; then
  echo "Certificate validation failed!"
  exit 1
fi
```

## Related Documentation

- [Certificate Skill](../../skills/certificate-analyzer/)
- [Changelog](./CHANGELOG.md)

## Contributing

Contributions welcome! This utility needs significant work to reach production quality. See [CONTRIBUTING.md](../../CONTRIBUTING.md).

Priority areas:
- Configuration system integration
- Bulk scanning support
- Output format options
- Monitoring capabilities
- Comprehensive testing

## License

GPL-3.0 - See [LICENSE](../../LICENSE) for details.

## Version

Current version: 1.0.0 (Experimental)

See [CHANGELOG.md](./CHANGELOG.md) for version history.
