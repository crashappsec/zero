<!--
SPDX-License-Identifier: GPL-3.0
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
-->

# Gibson Powers

> **Experimental Preview** - A collection of developer productivity and security engineering utilities powered by AI

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Status: Experimental](https://img.shields.io/badge/Status-Experimental-orange.svg)](https://github.com/crashappsec/gibson-powers)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)
[![GitHub Discussions](https://img.shields.io/badge/GitHub-Discussions-181717?logo=github)](https://github.com/crashappsec/gibson-powers/discussions)
[![Code of Conduct](https://img.shields.io/badge/Code%20of-Conduct-blue.svg)](CODE_OF_CONDUCT.md)

## What is Gibson Powers?

Gibson Powers is a suite of practical utilities for developers and security engineers, inspired by capabilities found in modern Developer Productivity Insights platforms (formerly known as Software Engineering Intelligence platforms). The name pays homage to the Gibson supercomputer from the film *Hackers* and adds a playful nod to Austin Powers.

### The Three-Tier Approach

Gibson Powers provides **three progressively powerful tiers** of capabilities:

```
Tier 1: Standalone Scripts     Tier 2: AI-Enhanced         Tier 3: Platform-Powered
─────────────────────────────  ─────────────────────────  ──────────────────────────
• Shell scripts                • Claude integration       • Crash Override platform
• Local analysis               • LLM-powered insights     • Enterprise features
• No dependencies              • Advanced reasoning       • Team collaboration
• Fast, simple                 • Comprehensive reports    • Historical analytics
```

**Tier 1** provides immediate value - just run the scripts on your code
**Tier 2** enhances analysis with Claude AI for deeper insights and recommendations
**Tier 3** (future) will integrate with the Crash Override platform for enterprise-scale deployments

This repository focuses on **Tiers 1 & 2**, giving you powerful standalone tools that get even better with AI.

## Features

- 🔒 **Supply Chain Security**: SBOM analysis, vulnerability scanning, provenance verification
- 🔍 **Technology Identification**: Automated technology stack detection and risk assessment
- ⚖️ **Legal Review**: License compliance, secret scanning, content safety analysis
- 📊 **DORA Metrics**: DevOps performance measurement (deployment frequency, lead time, etc.)
- 👥 **Code Ownership**: Bus factor analysis, knowledge transfer planning, CODEOWNERS generation
- 🔐 **Certificate Analysis**: X.509/TLS security review, expiration monitoring
- 📦 **Build Attestation**: Chalk build provenance verification, SLSA compliance
- 📈 **COCOMO Estimation**: Software development effort and cost estimation

All tools provide:
- ✅ **Standalone operation** (Tier 1) - works without any external services
- ✅ **AI enhancement** (Tier 2) - optional Claude integration for richer insights
- ✅ **Portable templates** - use prompts in Claude Desktop, Web, or API
- ✅ **Comprehensive documentation** - examples, guides, and best practices

## Quick Start

### Prerequisites

**For Tier 1 (Standalone)**:
- Bash 3.2+ (macOS/Linux)
- Git
- Standard Unix tools (jq, curl, etc.)

**For Tier 2 (AI-Enhanced)**:
- All Tier 1 prerequisites
- Claude API access (Anthropic API key)

### Installation

```bash
# Clone the repository
git clone https://github.com/crashappsec/gibson-powers.git
cd gibson-powers

# Make scripts executable
chmod +x utils/**/*.sh

# Try a standalone analysis on this repository
./utils/code-ownership/ownership-analyser-v2.sh .

# Try an AI-enhanced analysis (requires ANTHROPIC_API_KEY)
export ANTHROPIC_API_KEY="your-key"
./utils/code-ownership/ownership-analyser-claude.sh .
```

### Test Organization

🧪 **Want to try Gibson Powers on safe test repositories?**

We've created the [Gibson Powers Test Organization](https://github.com/Gibson-Powers-Test-Org) with sample repositories you can safely analyze without affecting real projects.

```bash
# Test Code Ownership Analysis
./utils/code-ownership/ownership-analyser.sh \
  https://github.com/Gibson-Powers-Test-Org/sample-repo

# Test DORA Metrics (if you have test data)
./utils/dora-metrics/dora-analyser.sh \
  --repo https://github.com/Gibson-Powers-Test-Org/sample-repo

# Test Supply Chain Analysis
./utils/supply-chain/supply-chain-scanner.sh \
  https://github.com/Gibson-Powers-Test-Org/sample-repo
```

**Perfect for:**
- Learning how the tools work
- Testing configurations
- Experimenting with features
- Contributing examples
- Creating tutorials

## Repository Structure

```
gibson-powers/
├── skills/                          # Claude Code skills (.skill files)
│   ├── supply-chain/                # Supply chain security skill
│   ├── technology-identification/   # Technology stack detection skill
│   ├── legal-review/                # Legal compliance skill
│   ├── dora-metrics/                # DORA metrics skill
│   ├── code-ownership/              # Code ownership skill
│   ├── certificate-analyser/        # Certificate analysis skill
│   ├── chalk-build-analyser/        # Chalk build analyser skill
│   └── better-prompts/              # Prompt engineering skill
│
├── utils/                           # Executable utilities (Tiers 1 & 2)
│   ├── supply-chain/                # Supply chain analysis tools
│   │   ├── supply-chain-scanner.sh      # Tier 1: Standalone scanner
│   │   ├── vulnerability-analysis/      # CVE scanning module
│   │   ├── provenance-analysis/         # SLSA/Sigstore verification
│   │   └── package-health-analysis/     # Dependency health checks
│   ├── technology-identification/   # Technology stack detection
│   │   └── technology-identification-analyser.sh
│   ├── legal-review/                # Legal compliance analysis
│   │   └── legal-analyser.sh
│   ├── dora-metrics/                # DORA metrics calculation
│   ├── code-ownership/              # Code ownership analysis
│   ├── certificate-analyser/        # X.509/TLS security analysis
│   ├── chalk-build-analyser/        # Build attestation verification
│   └── cocomo/                      # Software estimation tools
│
├── prompts/                         # Reusable prompt templates
│   ├── supply-chain/                # Supply chain prompts
│   ├── technology-identification/   # Technology detection prompts
│   ├── legal-review/                # Legal review prompts
│   ├── dora/                        # DORA metrics prompts
│   └── code-ownership/              # Code ownership prompts
│
├── rag/                             # RAG knowledge base
│   ├── supply-chain/                # Supply chain references
│   ├── technology-identification/   # Technology patterns
│   ├── legal-review/                # Legal compliance references
│   ├── dora-metrics/                # DORA best practices
│   └── code-ownership/              # Ownership patterns
│
└── docs/                            # Documentation
    ├── guides/                      # How-to guides
    └── references/                  # Reference docs
```

## Available Tools

### Supply Chain Security (🚀 Production-Ready)

Comprehensive software supply chain analysis with SBOM generation, vulnerability scanning, and provenance verification.

```bash
# Tier 1: Standalone analysis
./utils/supply-chain/supply-chain-scanner.sh --repo /path/to/repo

# Tier 2: AI-enhanced with recommendations
./utils/supply-chain/supply-chain-scanner-claude.sh --repo /path/to/repo
```

**Features**:
- SBOM generation (CycloneDX/SPDX)
- Vulnerability scanning (OSV.dev integration)
- SLSA provenance verification
- License compliance checking
- Dependency health assessment

[📖 Full Documentation](./utils/supply-chain/README.md)

### Technology Identification (🚀 Beta)

Automated detection and analysis of technology stacks across repositories.

```bash
# Standalone technology stack analysis
cd utils/technology-identification
./technology-identification-analyser.sh --repo owner/repo

# AI-enhanced with risk assessment and recommendations
export ANTHROPIC_API_KEY="your-key"
./technology-identification-analyser.sh --claude --repo owner/repo
```

**Features**:
- Multi-layered detection (6 layers with confidence scoring)
- Technology categorization (business tools, dev tools, languages, crypto, cloud)
- Version tracking and EOL detection
- Risk assessment (Critical → High → Medium → Low)
- Compliance implications (export control, licenses, data privacy)
- Executive and audit-focused reporting

[📖 Full Documentation](./utils/technology-identification/README.md)

### Legal Review (🚀 Production-Ready)

Comprehensive legal compliance analysis including licenses, secrets, and content safety.

```bash
# Standalone legal compliance scan
./utils/legal-review/legal-analyser.sh --repo owner/repo

# AI-enhanced with compliance recommendations
export ANTHROPIC_API_KEY="your-key"
./utils/legal-review/legal-analyser.sh --claude --repo owner/repo
```

**Features**:
- License compliance checking (SPDX, GPL, MIT, etc.)
- Secret scanning (API keys, credentials, tokens)
- Content safety analysis (inappropriate content detection)
- Export control compliance (ITAR/EAR)
- SBOM license extraction
- Audit-ready reporting

[📖 Full Documentation](./utils/legal-review/README.md)

### DORA Metrics (🔬 Experimental)

Measure software delivery performance using the four key DORA metrics.

```bash
# Tier 1: Calculate metrics from Git history
./utils/dora-metrics/dora-analyser.sh /path/to/repo

# Tier 2: AI-enhanced with insights and recommendations
./utils/dora-metrics/dora-analyser-claude.sh /path/to/repo
```

**Metrics**:
- Deployment Frequency
- Lead Time for Changes
- Change Failure Rate
- Mean Time to Recovery

[📖 Full Documentation](./utils/dora-metrics/README.md)

### Code Ownership Analysis (🚀 Production-Ready v3.0)

Analyze code ownership, identify knowledge risks, and plan succession.

```bash
# Tier 1: Detailed ownership analysis
./utils/code-ownership/ownership-analyser-v2.sh .

# Tier 2: AI-enhanced with strategic recommendations
./utils/code-ownership/ownership-analyser-claude.sh .
```

**Features**:
- Bus factor calculation
- Single points of failure (SPOF) detection
- CODEOWNERS file validation and generation
- Succession planning recommendations
- Historical trend tracking
- Markdown and CSV reporting

[📖 Full Documentation](./utils/code-ownership/README.md)

### Certificate Analysis (🔬 Experimental)

X.509 certificate and TLS configuration security review.

```bash
# Analyze a domain's certificate
./utils/certificate-analyser/cert-analyser.sh api.example.com

# AI-enhanced analysis with compliance insights
./utils/certificate-analyser/cert-analyser-claude.sh api.example.com
```

[📖 Full Documentation](./utils/certificate-analyser/README.md)

### Chalk Build Analyser (🔬 Experimental)

Verify build provenance and SLSA compliance using Chalk attestations.

```bash
# Extract and analyze Chalk marks
./utils/chalk-build-analyser/chalk-analyser.sh my-binary

# AI-enhanced compliance assessment
./utils/chalk-build-analyser/chalk-analyser-claude.sh my-binary
```

[📖 Full Documentation](./utils/chalk-build-analyser/README.md)

### COCOMO Estimation (🔬 Experimental)

Software development effort and cost estimation using COCOMO models.

```bash
# Estimate project effort
./utils/cocomo/cocomo-estimate.sh /path/to/repo
```

[📖 Full Documentation](./utils/cocomo/README.md)

## Portable Templates

All skills are available as **portable templates** that work in any Claude interface:

```bash
# Generate templates for use in Claude Desktop, Web, or API
./batch-create-templates.sh

# Templates created in ~/claude-templates/
# - supply-chain/supply-chain-comprehensive.md
# - dora-metrics/dora-analysis.md
# - code-ownership/ownership-analysis.md
# - security/certificate-analysis.md
# - And more...
```

Templates enable you to:
- ✅ Use Gibson Powers capabilities in Claude Desktop or Web
- ✅ Share analysis prompts with team members
- ✅ Integrate with CI/CD pipelines
- ✅ Customize for your specific needs

[📖 Template Documentation](~/claude-templates/CATALOG.md)

## Use Cases

### For Developers
- 📊 Measure and improve delivery performance (DORA metrics)
- 👥 Identify code owners and knowledge gaps
- 🔒 Audit dependencies for security vulnerabilities
- 📦 Verify build provenance and supply chain integrity
- 📈 Estimate project effort and timelines

### For Security Engineers
- 🔐 Supply chain security analysis (SBOM, vulnerabilities, provenance)
- 🔒 Certificate and TLS configuration review
- 📋 License compliance auditing
- 🛡️ Build attestation verification (SLSA, Sigstore)
- 🔍 Dependency health and risk assessment

### For Engineering Leaders
- 📈 Track team performance with DORA metrics
- 🎯 Identify single points of failure in codebases
- 📊 Plan knowledge transfer and succession
- 💰 Estimate project costs and timelines
- 🔄 Benchmark against industry standards

## Configuration

Gibson Powers uses a hierarchical configuration system:

1. Command-line arguments (highest priority)
2. Environment variables (`GIBSON_*`)
3. Local config (`.gibson.conf`)
4. Global config (`~/.config/gibson/config`)
5. Built-in defaults

```bash
# Example: Configure API keys
export GIBSON_ANTHROPIC_API_KEY="sk-ant-..."

# Example: Set analysis preferences
export GIBSON_ANALYSIS_DAYS=90
export GIBSON_OUTPUT_FORMAT=json

# Or use configuration files
cat > .gibson.conf << EOF
analysis_days=90
output_format=json
anthropic_api_key=sk-ant-...
EOF
```

[📖 Configuration Guide](./utils/CONFIG.md)

## Contributing

Gibson Powers is an **experimental preview** and we welcome contributions!

### How to Contribute

1. **Report bugs**: [Open an issue](https://github.com/crashappsec/gibson-powers/issues)
2. **Suggest features**: [Start a discussion](https://github.com/crashappsec/gibson-powers/discussions)
3. **Improve documentation**: Submit PRs for docs
4. **Add capabilities**: Create new analysers or enhance existing ones
5. **Share templates**: Contribute useful prompt templates

### Development

```bash
# Run tests
./utils/code-ownership/tests/run-all-tests.sh

# Validate scripts
./utils/validation/check-copyright.sh

# Format code
./utils/validation/format-scripts.sh
```

See [CONTRIBUTING.md](./CONTRIBUTING.md) for detailed guidelines.

## Roadmap

### Current (Experimental Preview)
- [x] Tier 1: Standalone scripts for all analysers
- [x] Tier 2: Claude AI integration
- [x] Portable template system
- [x] Comprehensive documentation

### Near-Term (Q1 2025)
- [ ] Enhanced CI/CD integrations (GitHub Actions, GitLab CI)
- [ ] Web dashboard for report visualization
- [ ] Additional analysers (test coverage, complexity metrics)
- [ ] Multi-repository batch analysis
- [ ] Team collaboration features

### Future (Tier 3)
- [ ] Crash Override platform integration
- [ ] Enterprise SSO and access control
- [ ] Historical trend analysis and alerts
- [ ] Custom metrics and KPI tracking
- [ ] API for programmatic access

See [ROADMAP.md](./ROADMAP.md) for details.

## Project Philosophy

Gibson Powers is built on these principles:

1. **Immediate Value**: Tier 1 tools provide instant insights without dependencies
2. **Progressive Enhancement**: Add AI (Tier 2) or platform features (Tier 3) when ready
3. **Open and Transparent**: GPL-3.0 licensed, open source, community-driven
4. **Practical Over Perfect**: Experimental features that solve real problems
5. **Learn in Public**: Share knowledge, examples, and best practices

## Community

- 💬 [Discussions](https://github.com/crashappsec/gibson-powers/discussions) - Ask questions, share ideas
- 🐛 [Issues](https://github.com/crashappsec/gibson-powers/issues) - Report bugs, request features
- 📖 [Wiki](https://github.com/crashappsec/gibson-powers/wiki) - Community knowledge base
- 🔐 [Security](./SECURITY.md) - Report security vulnerabilities

## License

Gibson Powers is licensed under the [GNU General Public License v3.0](./LICENSE).

```
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
```

## Acknowledgments

- Inspired by the Gibson supercomputer from the film *Hackers*
- Named with a playful nod to Austin Powers
- Built on research from DORA, SLSA, OpenSSF, and the open source community
- Powered by Anthropic's Claude AI (Tier 2)
- Part of the Crash Override ecosystem

## About

Gibson Powers is maintained by the open source community and sponsored by [Crash Override](https://crashoverride.com), a Developer Productivity Insights platform.

---

**Status**: Experimental Preview
**Version**: 3.0.0
**Last Updated**: 2024-11-22

🎮 **Hack the Planet!**
