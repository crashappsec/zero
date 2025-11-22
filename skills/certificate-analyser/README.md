<!--
Copyright (c) 2024 Gibson Powers Contributors

SPDX-License-Identifier: GPL-3.0
-->

# Certificate Analyser with Claude AI Integration - Complete Package

## 🎯 Overview

This is a complete certificate analysis system that combines automated certificate parsing with Claude AI's security expertise to provide comprehensive, actionable recommendations for certificate lifecycle management.

**What's New in v2.0:**
- ✨ **Claude AI Integration**: Intelligent analysis and recommendations
- 🔒 **CA/Browser Forum Compliance**: Automated policy checking
- 📊 **Risk Assessment**: Critical/Warning/Info categorization
- 🚀 **Future-Ready**: Preparation for 90-180 day certificates
- 📦 **Skill File**: Distributable .skill file for easy import into Claude

## 📦 Package Contents

### 1. **certificate-analyser-skill.skill** (Recommended)
The complete skill package ready for import into Claude. This is the easiest way to use the analyser.

**To Install:**
1. Go to Claude settings
2. Navigate to Skills section
3. Click "Import Skill"
4. Upload `certificate-analyser-skill.skill`

### 2. **cert-analyser-claude.sh** (Standalone Script)
Enhanced bash script with Claude API integration. Can be used independently without the skill.

**Features:**
- Automated certificate chain retrieval
- OpenSSL-based parsing and analysis
- Claude API integration for intelligent recommendations
- Comprehensive markdown report generation
- Color-coded console output

### 3. **Documentation**
- `cert-analyser-README.md` - Complete usage guide
- `certificate-analysis-prompt.md` - Technical prompt specification
- `sample-certificate-analysis-report.md` - Example output
- `PROJECT-OVERVIEW.md` - High-level project summary

## 🚀 Quick Start

### Option 1: Using the Skill (Easiest)

1. **Import the skill file** into Claude
2. **Ask Claude** to analyze a domain:
   ```
   "Please analyze the certificate for example.com"
   ```
3. Claude will automatically use the skill to:
   - Retrieve certificates
   - Perform security analysis
   - Generate comprehensive report with recommendations

### Option 2: Using the Standalone Script

```bash
# Set your Anthropic API key
export ANTHROPIC_API_KEY="sk-ant-api03-..."

# Make script executable
chmod +x cert-analyser-claude.sh

# Analyze a domain
./cert-analyser-claude.sh example.com

# View the generated report
cat certificate-analysis-example.com-*.md
```

### Option 3: Basic Analysis (No API Key)

```bash
# Run without Claude AI enhancement
./cert-analyser-claude.sh --no-claude example.com
```

## 🔑 API Key Setup

The Claude-enhanced analysis requires an Anthropic API key.

### Getting an API Key

1. Go to [console.anthropic.com](https://console.anthropic.com)
2. Sign up or log in
3. Navigate to API Keys
4. Create a new API key
5. Copy the key (starts with `sk-ant-`)

### Setting the API Key

**Option A: .env File** (Recommended)
```bash
# Copy .env.example to .env and add your API key
cp ../../.env.example ../../.env
# Edit .env and set ANTHROPIC_API_KEY=sk-ant-xxx
```

The script will automatically load the API key from the .env file when run.

**Option B: Environment Variable**
```bash
export ANTHROPIC_API_KEY="sk-ant-api03-..."
```

Add to your `~/.bashrc` or `~/.zshrc` for persistence:
```bash
echo 'export ANTHROPIC_API_KEY="sk-ant-..."' >> ~/.bashrc
```

**Option C: Command Line Flag**
```bash
./cert-analyser-claude.sh --api-key sk-ant-... example.com
```

**Option D: Skill Usage**
When using the skill within Claude, the API key is automatically available.

## 📊 What You Get

### Comprehensive Analysis Report

Each analysis generates a markdown report with:

1. **Executive Summary**
   - Domain information
   - Days until expiration
   - Compliance status
   - Critical alerts

2. **Certificate Chain Overview**
   - Leaf certificate details
   - Intermediate certificate(s)
   - Root certificate
   - Full technical specifications

3. **Claude AI Security Analysis**
   - CA/Browser Forum compliance assessment
   - Cryptographic security evaluation
   - Operational security posture review
   - Risk categorization (Critical/Warning/Info)
   - Prioritized recommendations by timeline

4. **Raw Certificate Data**
   - Complete OpenSSL output
   - Full certificate text for verification
   - Technical details for auditing

### Status Indicators

- ✓ **Compliant/Good**: Meets current standards
- ⚠️ **Warning**: Acceptable but requires attention
- ❌ **Critical**: Non-compliant or urgent issue
- ℹ️ **Informational**: Best practices and opportunities

## 🎯 Use Cases

### 1. Security Audits
Analyze all production domains before compliance audits:
```bash
for domain in api.example.com www.example.com admin.example.com; do
    ./cert-analyser-claude.sh "$domain"
done
```

### 2. Expiration Monitoring
Check if certificates need renewal:
```bash
./cert-analyser-claude.sh production.example.com
```

### 3. Incident Response
Investigate suspicious certificates during security events:
```bash
./cert-analyser-claude.sh suspicious-domain.com
```

### 4. Compliance Reporting
Generate evidence for auditors:
```bash
# Run analysis
./cert-analyser-claude.sh example.com

# Extract key findings
grep -E "✓|⚠️|❌" certificate-analysis-*.md
```

### 5. Automation Planning
Get recommendations for certificate automation:
- Claude analyzes your current setup
- Provides specific tool recommendations
- Suggests implementation timeline
- Considers your infrastructure

## 🔧 Requirements

### System Requirements
- Linux, macOS, or WSL on Windows
- Bash 4.0+
- Internet connectivity (port 443 outbound)

### Dependencies
- `openssl` - Certificate retrieval and parsing
- `curl` - Claude API communication
- `jq` - JSON processing
- `awk`, `sed`, `grep`, `date` - Standard utilities

### Installing Dependencies

**Ubuntu/Debian:**
```bash
sudo apt-get update
sudo apt-get install openssl curl jq
```

**macOS:**
```bash
brew install openssl curl jq
```

**RHEL/CentOS:**
```bash
sudo yum install openssl curl jq
```

## 📖 Usage Examples

### Basic Domain Analysis
```bash
./cert-analyser-claude.sh example.com
```

### Multiple Domains
```bash
#!/bin/bash
domains=("api.example.com" "www.example.com" "admin.example.com")
for domain in "${domains[@]}"; do
    ./cert-analyser-claude.sh "$domain"
done
```

### CI/CD Integration
```yaml
# .gitlab-ci.yml
certificate-check:
  stage: security
  script:
    - ./cert-analyser-claude.sh $PRODUCTION_DOMAIN
    - |
      if grep -q "❌" certificate-analysis-*.md; then
        echo "Critical certificate issues found"
        exit 1
      fi
  artifacts:
    paths:
      - certificate-analysis-*.md
    expire_in: 30 days
```

### Scheduled Monitoring
```bash
#!/bin/bash
# /etc/cron.daily/cert-monitor

export ANTHROPIC_API_KEY="sk-ant-..."
cd /opt/cert-analyser

for domain in $(cat /etc/cert-domains.txt); do
    ./cert-analyser-claude.sh "$domain"
    
    # Alert on critical issues
    if grep -q "❌" certificate-analysis-${domain}-*.md; then
        cat certificate-analysis-${domain}-*.md | \
            mail -s "CRITICAL: Certificate issues for $domain" ops@example.com
    fi
done
```

## 🎓 Understanding the Analysis

### CA/Browser Forum Compliance

**Current Policy (2020-Present):**
- Maximum validity: **398 days**
- Applies to certificates issued after September 1, 2020
- Older certificates may be grandfathered in

**Future Direction (2024-2027):**
- Expected reduction to **90-180 days**
- Manual processes won't scale
- Automation will be mandatory

### Cryptographic Strength

**Signature Algorithms:**
- ✓ SHA-256, SHA-384, SHA-512 (compliant)
- ❌ SHA-1 (deprecated since 2017)
- ❌ MD5 (broken, never use)

**Public Key Sizes:**
- ✓ RSA 3072+, RSA 4096, ECC P-256+ (strong)
- ⚠️ RSA 2048 (minimum acceptable)
- ❌ RSA 1024 or less (weak)

### Risk Levels

**Critical (❌) - Immediate Action Required:**
- Certificate expires within 7 days
- Weak cryptography (SHA-1, RSA <2048)
- Missing Certificate Transparency logs
- Invalid certificate chain

**Warning (⚠️) - Address Soon:**
- Certificate expires within 30 days
- RSA 2048 (consider upgrade)
- Missing OCSP stapling
- Long validity period

**Informational (ℹ️) - Best Practices:**
- Industry trends
- Optimization opportunities
- Future planning guidance

## 🔒 Security Considerations

### What This Tool Does
- ✓ Analyzes certificate metadata and compliance
- ✓ Checks cryptographic parameters
- ✓ Validates against industry policies
- ✓ Provides strategic recommendations

### What This Tool Doesn't Do
- ✗ Active security testing
- ✗ Vulnerability scanning
- ✗ Complete TLS/SSL configuration analysis
- ✗ Real-time threat detection

### Complement With
- **testssl.sh** - Comprehensive TLS testing
- **SSL Labs** - Public-facing analysis
- **OpenVAS/Nessus** - Vulnerability scanning
- **Commercial tools** - Venafi, Keyfactor, etc.

## 🐛 Troubleshooting

### "Failed to retrieve certificates"

**Possible Causes:**
- Domain doesn't exist or no DNS record
- No HTTPS service on port 443
- Firewall blocking outbound connections
- Network connectivity issues

**Solutions:**
```bash
# Check DNS
dig example.com

# Test HTTPS connectivity
curl -I https://example.com

# Test port 443
nc -zv example.com 443
```

### "No API key provided"

**Solution:**
```bash
# Set environment variable
export ANTHROPIC_API_KEY="sk-ant-..."

# Or use flag
./cert-analyser-claude.sh --api-key sk-ant-... example.com
```

### "jq: command not found"

**Solution:**
```bash
# Ubuntu/Debian
sudo apt-get install jq

# macOS
brew install jq
```

### Claude API Errors

**401 Unauthorized:**
- Invalid or expired API key
- Check key at console.anthropic.com

**429 Too Many Requests:**
- Rate limit exceeded
- Add delays between requests
- Consider request volume

**500 Server Error:**
- Anthropic service issue
- Retry later
- Use `--no-claude` for basic analysis

## 📈 Advanced Usage

### Bulk Domain Analysis
```bash
#!/bin/bash
# analyze-all-domains.sh

while IFS= read -r domain; do
    echo "Analyzing $domain..."
    ./cert-analyser-claude.sh "$domain"
    sleep 5  # Rate limiting
done < domains.txt

# Generate summary report
echo "## Certificate Analysis Summary" > summary.md
echo "" >> summary.md
for report in certificate-analysis-*.md; do
    echo "### $(basename $report .md)" >> summary.md
    grep -A 3 "Executive Summary" "$report" >> summary.md
    echo "" >> summary.md
done
```

### Extract Critical Issues
```bash
# Find all critical issues across reports
grep -h "❌" certificate-analysis-*.md | sort -u

# Count warnings per domain
for file in certificate-analysis-*.md; do
    count=$(grep -c "⚠️" "$file")
    echo "$file: $count warnings"
done
```

### Integration with Monitoring Systems
```bash
# Prometheus textfile collector format
#!/bin/bash
OUTPUT="/var/lib/node_exporter/textfile_collector/certificates.prom"

for domain in $DOMAINS; do
    days=$(./cert-analyser-claude.sh --no-claude "$domain" 2>&1 | \
           grep "Days until expiration" | awk '{print $4}')
    
    echo "certificate_days_remaining{domain=\"$domain\"} $days" >> "$OUTPUT"
done
```

## 📚 Additional Resources

### Documentation
- [CA/Browser Forum Baseline Requirements](https://cabforum.org/baseline-requirements/)
- [Let's Encrypt Certificate Lifespans](https://letsencrypt.org/2023/07/10/reducing-lifetime.html)
- [RFC 5280: X.509 PKI](https://tools.ietf.org/html/rfc5280)
- [Certificate Transparency](https://certificate.transparency.dev/)

### Tools & Services
- **Let's Encrypt** - Free automated certificates
- **Certbot** - ACME client for Let's Encrypt
- **cert-manager** - Kubernetes certificate management
- **SSL Labs** - Online certificate testing

### Learning
- [Anthropic API Documentation](https://docs.anthropic.com)
- [OpenSSL Cookbook](https://www.feistyduck.com/books/openssl-cookbook/)
- [Bash Scripting Guide](https://www.gnu.org/software/bash/manual/)

## 🤝 Support & Feedback

### Getting Help
1. Review this README and documentation
2. Check troubleshooting section
3. Verify all dependencies are installed
4. Test with known-good domains first

### Reporting Issues
When reporting issues, include:
- Domain being analyzed (if not sensitive)
- Complete error message
- Script version (`grep "Version:" cert-analyser-claude.sh`)
- Operating system and version
- OpenSSL version (`openssl version`)

### Feature Requests
Consider these enhancements:
- Additional output formats (JSON, HTML)
- Integration with certificate managers
- Bulk domain scanning
- Historical trend analysis
- Custom policy rules

## 📝 Version History

### v2.0 (Current) - November 2024
- ✨ Added Claude AI integration for intelligent analysis
- ✨ Created distributable skill package
- 🎯 Enhanced recommendations with priority/timeline
- 📊 Improved risk categorization
- 🔧 Better error handling and user feedback

### v1.0 - Original Release
- 📜 Basic certificate parsing and analysis
- ✅ CA/Browser Forum policy checking
- 📄 Markdown report generation
- 🔍 Cryptographic strength assessment

## 📄 License

MIT License - Feel free to use, modify, and distribute.

## 🙏 Credits

**Created for**: Security teams, operations engineers, compliance auditors  
**Purpose**: Simplify certificate analysis and prepare for upcoming policy changes  
**Powered by**: OpenSSL, Claude AI, and open source tools

---

## 🎯 Get Started Now

1. **Import the skill**: Upload `certificate-analyser-skill.skill` to Claude
2. **Or run standalone**: `./cert-analyser-claude.sh example.com`
3. **Review results**: Check generated markdown report
4. **Take action**: Follow prioritized recommendations

**Questions?** Review the documentation or check troubleshooting section.

**Ready to automate?** The Claude AI analysis will guide you toward the best automation strategy for your infrastructure.

---

*Certificate security is critical. This tool helps you stay ahead of policy changes and maintain strong security posture.*
