<!--
Copyright (c) 2024 Crash Override Inc
101 Fulton St, 416, New York 10038
SPDX-License-Identifier: GPL-3.0
-->

# Certificate Analyzer with Claude AI - Deliverables

## ⭐ RECOMMENDED: Start Here

### **certificate-analyzer.skill** (16 KB)
Import this .skill file into Claude for the easiest experience.

**How to use:**
1. Open Claude → Settings → Skills
2. Click "Import Skill"  
3. Upload certificate-analyzer.skill
4. Ask Claude: "Analyze the certificate for example.com"

**What you get:** Full AI-powered analysis with recommendations, all within Claude's interface.

---

## 🚀 Alternative Options

### **cert-analyzer-claude.sh** (23 KB)
Standalone script with Claude API integration.

**Use when:** You want command-line control or CI/CD integration

```bash
export ANTHROPIC_API_KEY="sk-ant-..."
./cert-analyzer-claude.sh example.com
```

### **cert-analyzer.sh** (24 KB)
Basic version without Claude AI (no API key required).

**Use when:** You need quick checks without AI enhancement

```bash
./cert-analyzer.sh example.com
```

---

## 📖 Documentation

- **README.md** - Complete guide (setup, usage, troubleshooting)
- **cert-analyzer-README.md** - Original documentation
- **PROJECT-OVERVIEW.md** - High-level project summary
- **certificate-analysis-prompt.md** - Technical prompt specification
- **sample-certificate-analysis-report.md** - Example output

---

## 🎯 Quick Decision Guide

**"I use Claude regularly"** → Import certificate-analyzer.skill  
**"I need CLI automation"** → Use cert-analyzer-claude.sh  
**"I don't have an API key"** → Use cert-analyzer.sh  

---

## ✅ What You'll Get

Every analysis provides:
- Days until expiration with urgency indicators
- CA/Browser Forum compliance status
- Cryptographic strength assessment  
- Certificate Transparency verification
- Risk categorization (Critical/Warning/Info)
- AI-powered recommendations (with Claude version)
- Complete technical details in markdown format

---

## 🔑 API Key Setup

1. Visit https://console.anthropic.com
2. Create API key
3. Set environment variable:
   ```bash
   export ANTHROPIC_API_KEY="sk-ant-..."
   ```

Required for: cert-analyzer-claude.sh  
Not required for: certificate-analyzer.skill (handled by Claude)

---

## 💡 Perfect For

- Pre-audit security assessments
- Certificate expiration monitoring
- Compliance reporting
- Incident response
- Automation planning (preparing for 90-180 day certs)

---

**Start with certificate-analyzer.skill for the best experience!**
