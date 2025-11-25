<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Certificate Analysis Prompts

Comprehensive prompt templates for X.509 certificate analysis, TLS security assessment, and CA/Browser Forum compliance verification using the Certificate Analyser skill.

## Directory Structure

```
prompts/certificate-analysis/
├── security/              # Security-focused analysis
│   ├── security-audit.md
│   └── chain-validation.md
├── compliance/            # Compliance checking
│   ├── cab-forum-compliance.md
│   └── expiry-monitoring.md
├── operations/            # Operational tasks
│   ├── format-conversion.md
│   └── certificate-comparison.md
└── troubleshooting/       # Diagnosis and debugging
    └── tls-troubleshooting.md
```

## Categories

### 🔒 Security
Security analysis and vulnerability assessment for certificates.

- **[security-audit.md](security/security-audit.md)** - Comprehensive certificate security audit
- **[chain-validation.md](security/chain-validation.md)** - Certificate chain validation and trust analysis

### ✅ Compliance
Certificate compliance and policy verification.

- **[cab-forum-compliance.md](compliance/cab-forum-compliance.md)** - CA/Browser Forum Baseline Requirements compliance
- **[expiry-monitoring.md](compliance/expiry-monitoring.md)** - Certificate expiration monitoring and alerting

### ⚙️ Operations
Certificate operational tasks and management.

- **[format-conversion.md](operations/format-conversion.md)** - Convert between PEM, DER, PKCS#7, PKCS#12 formats
- **[certificate-comparison.md](operations/certificate-comparison.md)** - Compare certificates for changes or matches

### 🔧 Troubleshooting
Certificate issue diagnosis and resolution.

- **[tls-troubleshooting.md](troubleshooting/tls-troubleshooting.md)** - Diagnose TLS connection and certificate issues

## Quick Start

1. **Load the Certificate Analyser Skill** in Claude Code
2. **Choose a category** based on your needs
3. **Select a prompt template** for your specific task
4. **Copy and customize** the prompt with your target
5. **Execute and review** the analysis

## Common Workflows

### Security Assessment
```
1. security-audit.md          → Comprehensive security analysis
2. chain-validation.md        → Validate certificate chain
3. cab-forum-compliance.md    → Check policy compliance
```

### Certificate Renewal
```
1. expiry-monitoring.md       → Check expiration dates
2. security-audit.md          → Validate new certificate
3. certificate-comparison.md  → Compare old vs new
```

### Incident Response
```
1. tls-troubleshooting.md     → Diagnose connection issues
2. chain-validation.md        → Verify trust chain
3. security-audit.md          → Full security assessment
```

### Compliance Audit
```
1. cab-forum-compliance.md    → Full compliance check
2. security-audit.md          → Security posture review
3. expiry-monitoring.md       → Certificate lifecycle status
```

## Use Cases by Role

### Security Teams
- **Daily**: security-audit, chain-validation
- **Weekly**: cab-forum-compliance
- **As Needed**: tls-troubleshooting

### Compliance Officers
- **Pre-Audit**: cab-forum-compliance
- **Monthly**: expiry-monitoring
- **Vendor Review**: security-audit

### DevOps/SRE
- **Deployment**: certificate-comparison
- **CI/CD**: chain-validation
- **Troubleshooting**: tls-troubleshooting

### Engineering Managers
- **Release Readiness**: security-audit
- **Compliance**: cab-forum-compliance
- **Planning**: expiry-monitoring

## Integration with Certificate Analyser

```bash
# Basic domain analysis
./utils/certificate-analyser/cert-analyser.sh example.com

# Full compliance check
./utils/certificate-analyser/cert-analyser.sh --all-checks example.com

# CA/B Forum compliance
./utils/certificate-analyser/cert-analyser.sh --compliance example.com

# StartTLS SMTP
./utils/certificate-analyser/cert-analyser.sh --starttls smtp mail.example.com

# With Claude AI analysis
./utils/certificate-analyser/cert-analyser.sh --claude example.com
```

## RAG Knowledge Base

These prompts are supported by comprehensive RAG documentation:

- **[X.509 Certificates](../../rag/certificate-analysis/x509/)** - Certificate structure, extensions, algorithms
- **[CA/B Forum Requirements](../../rag/certificate-analysis/cab-forum/)** - Baseline Requirements, compliance
- **[TLS Security](../../rag/certificate-analysis/tls-security/)** - TLS best practices, cipher suites
- **[Certificate Formats](../../rag/certificate-analysis/formats/)** - PEM, DER, PKCS formats
- **[Revocation](../../rag/certificate-analysis/revocation/)** - OCSP, CRL documentation

## Output Formats

Prompts can generate:
- Markdown reports with risk categorization
- JSON data for automation
- Executive summaries
- Technical deep-dives
- Compliance checklists

## Best Practices

### Regular Monitoring
- **Daily**: Check critical certificates for expiry
- **Weekly**: Run compliance scans
- **Monthly**: Full security audits

### Certificate Lifecycle
- Alert at 30/14/7 days before expiry
- Automate renewal with ACME/Let's Encrypt
- Validate new certificates before deployment

### Documentation
- Keep audit reports for compliance evidence
- Track remediation actions
- Document certificate inventory

## Contributing

Have a useful certificate analysis prompt? Please contribute!

1. Choose the appropriate category
2. Follow the template structure in existing prompts
3. Include purpose, usage, examples, and variations
4. Test thoroughly before submitting
5. Submit a pull request

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for guidelines.
