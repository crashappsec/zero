<!--
Copyright (c) 2024 Crash Override Inc
101 Fulton St, 416, New York 10038
SPDX-License-Identifier: GPL-3.0
-->

# Certificate Analysis System - Project Overview

## What I've Built

A complete certificate analysis system consisting of:

1. **Technical Prompt** (`certificate-analysis-prompt.md`)
   - Comprehensive prompt for Claude to analyze certificates
   - Detailed instructions on data retrieval, analysis criteria, and report generation
   - Can be used standalone with Claude or as a skill

2. **Bash Script** (`cert-analyzer.sh`)
   - Production-ready executable script
   - Automated certificate retrieval and analysis
   - Generates professional markdown reports
   - ~600 lines of robust bash with error handling

3. **Documentation** (`cert-analyzer-README.md`)
   - Complete usage guide with examples
   - Troubleshooting section
   - Integration examples (CI/CD, Kubernetes, scheduled monitoring)
   - Best practices and security considerations

4. **Sample Report** (`sample-certificate-analysis-report.md`)
   - Example output showing what the analysis looks like
   - Demonstrates all report sections and formatting

## Key Features

### Compliance Analysis
- ✓ CA/Browser Forum 398-day policy compliance checking
- ✓ Future policy preparation (90-180 day certificates)
- ✓ Legacy certificate identification (pre-2020)

### Security Assessment
- ✓ Signature algorithm strength (SHA-1 vs SHA-256+)
- ✓ Public key analysis (RSA 2048/3072/4096, ECC P-256+)
- ✓ Certificate Transparency verification
- ✓ OCSP and CRL availability checking
- ✓ Chain validation

### Risk Management
- ✓ Expiration monitoring (7-day critical, 30-day warning)
- ✓ Prioritized recommendations (immediate/short-term/long-term)
- ✓ Risk categorization (Critical ❌ / Warning ⚠️ / Info ℹ️)

### Professional Reporting
- ✓ Executive summary for leadership
- ✓ Technical details for security teams
- ✓ Actionable recommendations
- ✓ Raw certificate data appendix
- ✓ Markdown format (portable, version-controllable)

## How to Use

### Quick Start

```bash
# Make the script executable
chmod +x cert-analyzer.sh

# Analyze a domain
./cert-analyzer.sh example.com

# View the generated report
cat certificate-analysis-example.com-*.md
```

### What Happens

1. Script connects to domain:443
2. Retrieves complete certificate chain via OpenSSL
3. Parses each certificate for 20+ data points
4. Analyzes against CA/B Forum policies
5. Generates comprehensive markdown report
6. Provides console summary with color-coded status

### Example Output

```
[INFO] Starting certificate analysis for example.com
[SUCCESS] Retrieved 3 certificate(s) from chain
[INFO] Generating analysis report...
[SUCCESS] Report generated: certificate-analysis-example.com-20241119.md

═══════════════════════════════════════════════════════════
  Certificate Analysis Summary
═══════════════════════════════════════════════════════════
  Domain: example.com
  Certificates in chain: 3
  Days until expiration: 45
  Validity period: 90 days
  Compliance: ✓ Compliant
═══════════════════════════════════════════════════════════
```

## Report Structure

Each report contains 7 major sections:

### 1. Executive Summary
- High-level findings
- Critical alerts
- Days until expiration
- Compliance status

### 2. Certificate Chain Overview
- Leaf certificate (end-entity)
- Intermediate certificate(s)
- Root certificate
- All metadata: subject, issuer, serial, dates, algorithms

### 3. Detailed Analysis
- **Validity Compliance**: Current and future policy alignment
- **Cryptographic Strength**: Algorithm and key size assessment
- **SANs Coverage**: Domain coverage analysis
- **Certificate Transparency**: CT log verification
- **Revocation Checking**: OCSP and CRL availability
- **Chain Validation**: Trust chain verification

### 4. Risk Assessment
- Critical issues requiring immediate action
- Warnings for near-term attention
- Informational items and best practices

### 5. Recommendations
- **Immediate (0-30 days)**: Urgent actions
- **Short-term (30-90 days)**: Implementation planning
- **Long-term (90+ days)**: Strategic initiatives

### 6. Appendix
- Complete raw certificate data
- OpenSSL text output for verification

### 7. Methodology Notes
- Tools and versions used
- Standards referenced
- Analysis timestamp

## Use Cases

### 1. Security Audits
Run before audits to identify certificate issues proactively:
```bash
./cert-analyzer.sh production-domain.com
grep -E "❌|⚠️" certificate-analysis-*.md
```

### 2. Compliance Reporting
Generate evidence of certificate compliance:
```bash
for domain in $(cat domains.txt); do
    ./cert-analyzer.sh "$domain"
done
# Aggregate reports for compliance documentation
```

### 3. Incident Response
Quickly analyze certificates during security incidents:
```bash
./cert-analyzer.sh suspicious-domain.com
# Review signature algorithms, issuers, CT logs
```

### 4. Certificate Inventory
Build comprehensive certificate inventory:
```bash
# Scan all production domains
./cert-analyzer.sh api.example.com
./cert-analyzer.sh www.example.com
./cert-analyzer.sh admin.example.com
# Store reports in git for historical tracking
```

### 5. Renewal Planning
Proactive certificate lifecycle management:
```bash
# Daily cron job
0 6 * * * /usr/local/bin/cert-analyzer.sh production.com | grep "⚠️\|❌"
```

## Integration Patterns

### CI/CD Pipeline
```yaml
# .gitlab-ci.yml
certificate-check:
  stage: test
  script:
    - ./cert-analyzer.sh $CI_ENVIRONMENT_URL
    - if grep -q "❌" certificate-*.md; then exit 1; fi
  artifacts:
    paths:
      - certificate-*.md
```

### Kubernetes Monitoring
```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: cert-monitor
spec:
  schedule: "0 6 * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: cert-analyzer
            image: alpine:latest
            command: ["/bin/sh", "-c"]
            args:
              - apk add --no-cache bash openssl curl;
                curl -o cert-analyzer.sh https://your-repo/cert-analyzer.sh;
                chmod +x cert-analyzer.sh;
                ./cert-analyzer.sh your-domain.com;
```

### Slack Notifications
```bash
#!/bin/bash
# cert-monitor-slack.sh

DOMAIN="$1"
WEBHOOK_URL="$2"

./cert-analyzer.sh "$DOMAIN"

if grep -q "❌" certificate-*.md; then
    curl -X POST "$WEBHOOK_URL" \
        -H 'Content-Type: application/json' \
        -d "{\"text\":\"🚨 Critical certificate issues found for $DOMAIN\"}"
fi
```

## Technical Details

### Dependencies
- **OpenSSL**: Certificate retrieval and parsing
- **Standard Unix utilities**: awk, sed, grep, date
- **Bash 4.0+**: Modern bash features

### Architecture
The script follows a modular design:

```
retrieve_certificates()      # OpenSSL s_client
    ↓
parse_certificates()         # Extract metadata
    ↓
analyze_compliance()         # CA/B Forum policies
    ↓
assess_security()           # Crypto strength
    ↓
generate_report()           # Markdown output
```

### Configuration Variables
```bash
CURRENT_MAX_VALIDITY=398  # Days (adjustable)
EXPIRY_CRITICAL=7         # Alert threshold
EXPIRY_WARNING=30         # Warning threshold
```

### Error Handling
- Network timeout protection
- Graceful failure modes
- Informative error messages
- Partial data analysis capability

## CA/Browser Forum Context

### Current Policy (Sept 2020 - Present)
- Maximum validity: **398 days**
- Required for all publicly trusted certificates
- Issued after September 1, 2020

### Future Direction (2024-2027)
- Expected reduction to **90-180 days**
- Automation will be mandatory
- Manual processes won't scale

### Why Shorter Certificates?
1. **Reduced Risk Window**: Compromised keys have shorter exposure
2. **Forced Validation**: More frequent domain ownership checks
3. **Crypto Agility**: Easier to deprecate weak algorithms
4. **Incident Response**: Faster recovery from CA compromises

## Best Practices

### For Operations Teams
1. Run analysis weekly for all production domains
2. Set up alerting for expiration warnings
3. Document renewal procedures in runbooks
4. Track certificate inventory in asset management

### For Security Teams
1. Include in security audit procedures
2. Monitor for weak cryptography
3. Verify Certificate Transparency compliance
4. Review chain validation regularly

### For Development Teams
1. Integrate into CI/CD pipelines
2. Test certificate renewals in staging
3. Automate using ACME protocol
4. Use Infrastructure as Code for certificate management

### For Management
1. Use executive summaries for reporting
2. Budget for certificate automation tools
3. Understand compliance requirements
4. Plan for shorter certificate lifespans

## Roadmap / Future Enhancements

Potential improvements (contributions welcome):

- [ ] **JSON output format** for easier parsing/integration
- [ ] **Local certificate analysis** (analyze .pem files directly)
- [ ] **Bulk domain scanning** (analyze multiple domains at once)
- [ ] **CT log monitoring** (alert on new certificates)
- [ ] **CAA record verification** (DNS authorization)
- [ ] **HSTS header checking** (security best practices)
- [ ] **Cipher suite analysis** (TLS configuration)
- [ ] **Historical comparison** (track changes over time)
- [ ] **API integration** (CRT.sh, SSLLabs, etc.)
- [ ] **Dashboard generation** (HTML summary page)

## Support Matrix

### Tested Environments
- ✓ Ubuntu 20.04, 22.04, 24.04
- ✓ Debian 11, 12
- ✓ RHEL 8, 9
- ✓ Alpine Linux 3.18+
- ✓ macOS 12+ (with Homebrew OpenSSL)

### Known Limitations
- Requires outbound HTTPS connectivity
- Cannot analyze certificates behind authentication
- Root CA validation depends on local trust store
- Some CAs may rate-limit certificate requests

## Security Considerations

**What This Tool Does:**
- ✓ Analyzes certificate metadata
- ✓ Checks compliance with policies
- ✓ Validates cryptographic parameters
- ✓ Provides recommendations

**What This Tool Doesn't Do:**
- ✗ Active security testing
- ✗ Vulnerability scanning
- ✗ TLS handshake analysis
- ✗ Full PKI auditing

**Complement With:**
- testssl.sh for comprehensive TLS testing
- SSLLabs for public-facing analysis
- Qualys SSL Labs API for automation
- OpenVAS or Nessus for vulnerability scanning

## Quick Reference

### Command Syntax
```bash
./cert-analyzer.sh <domain>
```

### Example Domains
```bash
./cert-analyzer.sh example.com          # Basic usage
./cert-analyzer.sh www.github.com       # With subdomain
./cert-analyzer.sh api.stripe.com       # API endpoint
./cert-analyzer.sh https://google.com   # Auto-strips protocol
```

### Report File Naming
```
certificate-analysis-{domain}-{YYYYMMDD-HHMMSS}.md
```

### Status Indicators
- ✓ **Pass/Compliant**: Meets standards
- ⚠️ **Warning**: Attention needed
- ❌ **Critical**: Urgent issue
- ℹ️ **Info**: Additional context

## File Manifest

```
certificate-analysis-system/
├── certificate-analysis-prompt.md     # Technical prompt for Claude
├── cert-analyzer.sh                   # Executable bash script
├── cert-analyzer-README.md            # Usage documentation
├── sample-certificate-analysis-report.md  # Example output
└── PROJECT-OVERVIEW.md                # This file
```

## Getting Started Checklist

- [ ] Download `cert-analyzer.sh`
- [ ] Make executable: `chmod +x cert-analyzer.sh`
- [ ] Test on a domain: `./cert-analyzer.sh example.com`
- [ ] Review generated report
- [ ] Set up scheduled monitoring
- [ ] Integrate into CI/CD
- [ ] Document renewal procedures
- [ ] Train team on usage

## Contact & Feedback

This tool was created to help teams manage certificates in light of CA/Browser Forum policy changes. Feedback, bug reports, and contributions are welcome.

**Version**: 1.0  
**Last Updated**: November 2024  
**Author**: Certificate Security Analyzer Project

---

*For additional support, refer to the detailed README or consult with your security team.*
