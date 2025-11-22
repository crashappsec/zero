<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# 📦 Certificate Analyser with Claude AI - Complete Package Manifest

## 🎯 START HERE

### **🌟 certificate-analyser.skill** (16 KB) - RECOMMENDED
**The complete skill package for Claude**

Import this file into Claude for the easiest, most powerful experience. Just upload the .skill file and ask Claude to analyze any domain - that's it!

**Installation:**
```
Claude Settings → Skills → Import Skill → Upload certificate-analyser.skill
```

**Usage:**
```
"Please analyze the certificate for example.com"
```

---

## 📋 Complete File Listing

### Core Files

#### **certificate-analyser.skill** ⭐ (16 KB)
- **Type**: Distributable skill package (ZIP format)
- **Contains**: SKILL.md + cert-analyser script + analysis guidelines
- **Use**: Import into Claude for AI-powered certificate analysis
- **Best for**: Claude users who want seamless integration

#### **cert-analyser-claude.sh** 🚀 (23 KB)
- **Type**: Bash script with Claude API integration
- **Requires**: ANTHROPIC_API_KEY environment variable
- **Use**: Standalone command-line certificate analyser
- **Best for**: CI/CD pipelines, automation, scheduled monitoring

#### **cert-analyser.sh** 📊 (24 KB)
- **Type**: Basic bash script (no Claude integration)
- **Requires**: Only openssl, curl, standard utilities
- **Use**: Quick certificate checks without API key
- **Best for**: Fast checks, systems without API access

---

### Documentation Files

#### **README.md** 📖 (14 KB) - PRIMARY DOCUMENTATION
**Comprehensive guide covering everything:**
- Quick start for all three methods
- API key setup instructions
- Complete usage examples
- Troubleshooting guide
- Integration patterns (CI/CD, cron, Kubernetes)
- Best practices for ops/security/dev teams
- Advanced features and bulk operations

**Read this first for detailed instructions!**

#### **QUICK-START.md** ⚡ (2.4 KB)
**Super quick reference:**
- Which file to use when
- 30-second setup for each method
- Decision tree (skill vs script)

#### **PROJECT-OVERVIEW.md** 🔍 (12 KB)
**High-level project summary:**
- What the system does
- Key features and capabilities
- Use cases and scenarios
- Technical architecture
- Integration patterns
- Roadmap

#### **cert-analyser-README.md** 📄 (10 KB)
**Original documentation:**
- Detailed script usage
- Network configuration
- File system structure
- Example workflows

#### **certificate-analysis-prompt.md** 📝 (8 KB)
**Technical prompt specification:**
- Analysis methodology
- Report structure
- Command reference
- Useful for customization

#### **sample-certificate-analysis-report.md** 📊 (11 KB)
**Example report output:**
- Complete certificate analysis
- All report sections
- Status indicators
- Recommendations format

---

## 🎬 Quick Start by Preference

### "I Use Claude" → certificate-analyser.skill
```
1. Import into Claude (Settings → Skills)
2. Ask: "Analyze the certificate for example.com"
3. Get comprehensive AI-powered report
```

### "I Want CLI Control" → cert-analyser-claude.sh
```bash
export ANTHROPIC_API_KEY="sk-ant-..."
chmod +x cert-analyser-claude.sh
./cert-analyser-claude.sh example.com
```

### "I Don't Have API Key" → cert-analyser.sh
```bash
chmod +x cert-analyser.sh
./cert-analyser.sh example.com
```

---

## 🔑 API Key Requirements

| File | API Key Required? | Where Set |
|------|-------------------|-----------|
| certificate-analyser.skill | ❌ No (handled by Claude) | N/A |
| cert-analyser-claude.sh | ✅ Yes | Environment variable |
| cert-analyser.sh | ❌ No | N/A |

**Getting API Key:**
1. Visit https://console.anthropic.com
2. Sign up/login
3. API Keys → Create new key
4. Copy key (starts with `sk-ant-`)

---

## 📊 Feature Comparison

| Feature | Skill | Claude Script | Basic Script |
|---------|-------|---------------|--------------|
| Import into Claude | ✅ | ❌ | ❌ |
| Natural language usage | ✅ | ❌ | ❌ |
| CLI automation | ❌ | ✅ | ✅ |
| Claude AI analysis | ✅ | ✅ | ❌ |
| API key required | ❌ | ✅ | ❌ |
| CI/CD integration | ❌ | ✅ | ✅ |
| Offline capable | ❌ | ❌ | ✅* |

*After certificate retrieval

---

## 🎯 Use Case → File Mapping

### Security Audits
→ Use **skill** or **cert-analyser-claude.sh** for comprehensive AI analysis

### Expiration Monitoring
→ Use **cert-analyser-claude.sh** in cron jobs

### CI/CD Pipeline
→ Use **cert-analyser-claude.sh** or **cert-analyser.sh**

### Quick Manual Check
→ Use **skill** (easiest) or **cert-analyser.sh** (fastest)

### Compliance Reporting
→ Use **skill** or **cert-analyser-claude.sh** for detailed recommendations

---

## 💡 What Each Analysis Provides

### All Versions Include:
- ✅ Certificate chain details (leaf/intermediate/root)
- ✅ Days until expiration with urgency levels
- ✅ CA/Browser Forum compliance checking
- ✅ Cryptographic strength assessment
- ✅ Certificate Transparency verification
- ✅ OCSP and CRL availability
- ✅ SANs coverage analysis
- ✅ Markdown report output

### Claude AI Versions Add:
- 🤖 Intelligent security analysis
- 🤖 Context-aware recommendations
- 🤖 Prioritized action items by timeline
- 🤖 Risk categorization (Critical/Warning/Info)
- 🤖 Automation strategy guidance
- 🤖 Future-proofing advice (90-180 day certs)

---

## 🛠️ System Requirements

### Required Software
```bash
# All versions need:
- bash 4.0+
- openssl
- curl

# Claude-enhanced versions also need:
- jq

# Installation:
apt-get install openssl curl jq        # Ubuntu/Debian
brew install openssl curl jq           # macOS
yum install openssl curl jq            # RHEL/CentOS
```

### Network Requirements
- Outbound HTTPS (port 443) for certificate retrieval
- Internet access to anthropic.com for Claude API (if using AI features)

---

## 📈 Analysis Capabilities

### Certificate Analysis
- Validity period compliance (CA/B Forum 398-day policy)
- Cryptographic algorithms (SHA-256+, RSA 2048+, ECC P-256+)
- Certificate Transparency log verification
- Revocation checking mechanisms
- Chain validation and trust path
- Subject Alternative Names coverage
- Key usage and extended key usage
- Expiration tracking

### Recommendations Include
- Immediate actions (0-7 days)
- Short-term planning (7-30 days)
- Medium-term implementation (30-90 days)
- Long-term strategy (90+ days)
- Automation tool recommendations
- Migration strategies

---

## 🚀 Example Workflows

### Daily Monitoring
```bash
#!/bin/bash
export ANTHROPIC_API_KEY="sk-ant-..."
domains=(api.example.com www.example.com)
for d in "${domains[@]}"; do
    ./cert-analyser-claude.sh "$d"
done
```

### CI/CD Integration
```yaml
certificate-check:
  script:
    - ./cert-analyser-claude.sh $PROD_DOMAIN
    - if grep -q "❌" *.md; then exit 1; fi
```

### Bulk Analysis
```bash
cat domains.txt | while read domain; do
    ./cert-analyser-claude.sh "$domain"
    sleep 2  # Rate limiting
done
```

---

## 📚 Documentation Reading Order

1. **QUICK-START.md** - Get running in 2 minutes
2. **README.md** - Comprehensive guide when you need details
3. **sample-certificate-analysis-report.md** - See what you'll get
4. **PROJECT-OVERVIEW.md** - Understand the bigger picture
5. **certificate-analysis-prompt.md** - Technical deep dive

---

## ✅ Success Checklist

Before you start:
- [ ] Choose your method (skill/script)
- [ ] Install dependencies (openssl, curl, jq)
- [ ] Get API key (if using Claude features)
- [ ] Test with a known domain
- [ ] Review example report
- [ ] Set up monitoring schedule

---

## 🎓 Key Concepts

### CA/Browser Forum Policies
- **Current**: Max 398 days validity (since Sept 2020)
- **Future**: Expected 90-180 days (2024-2027)
- **Impact**: Manual renewals won't scale → automation mandatory

### Status Indicators
- ✓ **Compliant** - Meets current standards
- ⚠️ **Warning** - Acceptable but needs attention soon
- ❌ **Critical** - Urgent issue requiring immediate action
- ℹ️ **Info** - Best practices and optimization opportunities

### Risk Levels
- **Critical (❌)**: Expires ≤7 days, weak crypto, missing CT logs
- **Warning (⚠️)**: Expires ≤30 days, RSA 2048 (minimum)
- **Info (ℹ️)**: Industry trends, optimization opportunities

---

## 🔒 Security & Privacy

### What We Access
- ✅ Public certificate metadata only
- ✅ No private keys ever touched
- ✅ Read-only port 443 connections
- ✅ No data stored permanently

### API Usage
- ✅ Certificate data sent to Claude API (when using AI)
- ✅ Encrypted HTTPS transmission
- ✅ No logging of sensitive data
- ✅ Follow Anthropic's privacy policy

---

## 💼 For Different Teams

### Operations Teams
- Use **skill** for ad-hoc checks
- Use **cert-analyser-claude.sh** for scheduled monitoring
- Set up alerts for 30/14/7 day expiration
- Document renewal procedures

### Security Teams
- Include in audit procedures
- Monitor for weak cryptography
- Track CT log compliance
- Use AI recommendations for policy updates

### Development Teams
- Integrate **cert-analyser-claude.sh** in CI/CD
- Test certificate renewals in staging
- Plan ACME protocol adoption
- Infrastructure as Code for cert management

---

## 🎯 File Size Summary

```
Total Package: ~120 KB

Core Files:
  certificate-analyser.skill      16 KB  ⭐
  cert-analyser-claude.sh         23 KB  🚀
  cert-analyser.sh                24 KB  📊

Documentation:
  README.md                       14 KB  📖
  PROJECT-OVERVIEW.md             12 KB  🔍
  sample-certificate-report.md    11 KB  📊
  cert-analyser-README.md         10 KB  📄
  certificate-analysis-prompt.md   8 KB  📝
  QUICK-START.md                   2 KB  ⚡
```

---

## 🎉 You're All Set!

Everything you need is in this package:

✅ Three ways to analyze certificates (skill, enhanced script, basic script)  
✅ Comprehensive documentation  
✅ Example reports  
✅ Integration guides  
✅ Troubleshooting help  

**Start with certificate-analyser.skill for the easiest experience!**

Questions? Check README.md for detailed guidance.

---

*Built for security teams to stay ahead of certificate policy changes and maintain strong security posture.*

**Version**: 2.0 (Claude-Enhanced)  
**Release**: November 2024  
**Claude Model**: claude-sonnet-4-20250514
