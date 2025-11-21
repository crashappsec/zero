<!--
Copyright (c) 2024 Crash Override Inc
101 Fulton St, 416, New York 10038
SPDX-License-Identifier: GPL-3.0
-->

# Skills and Prompts for Crash Override

A curated collection of skills, prompts, and tools to enhance your experience with the Crash Override platform. This repository provides reusable components, best practices, and a growing knowledge base powered by RAG (Retrieval-Augmented Generation) to help you get the most out of Crash Override.

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Repository Structure](#repository-structure)
- [Getting Started](#getting-started)
- [Available Skills](#available-skills)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [RAG Knowledge Base](#rag-knowledge-base)
- [Community](#community)
- [License](#license)

## Overview

This repository serves as a central hub for:

- **Skills**: Pre-built `.skill` files that extend Crash Override's capabilities
- **Prompts**: Tested and optimized prompt templates for common use cases
- **Tools**: Utilities and scripts to enhance your workflow
- **Documentation**: Guides, references, and best practices
- **RAG Knowledge Base**: A growing corpus of domain knowledge to improve responses

Whether you're analyzing security certificates, building software, or engineering better prompts, this repository provides battle-tested components to accelerate your work.

## Features

- **Production-Ready Skills**: Fully documented skills with examples and changelogs
- **Organized Prompts**: Categorized by domain (security, development, analysis)
- **Community-Driven**: Open source contributions welcome
- **RAG-Powered**: Enhanced context and knowledge retrieval
- **Best Practices**: Learn from real-world examples and conversations

## Repository Structure

```
skills-and-prompts/
├── skills/                          # Claude skills and documentation only
│   ├── supply-chain/                # Supply chain security skill
│   ├── dora-metrics/                # DORA metrics skill
│   ├── code-ownership/              # Code ownership skill
│   ├── certificate-analyzer/        # Certificate analysis skill
│   └── chalk-build-analyzer/        # Chalk build analyzer skill
│
├── utils/                           # Executable scripts and utilities
│   ├── supply-chain/                # Supply chain analysis (🚀 Beta)
│   │   ├── vulnerability-analysis/  # Vulnerability scanning module
│   │   ├── provenance-analysis/     # SLSA provenance verification module
│   │   ├── config.example.json      # Configuration template
│   │   ├── README.md                # Complete documentation
│   │   ├── CHANGELOG.md             # Version history
│   │   └── supply-chain-scanner.sh  # Central orchestrator
│   ├── dora-metrics/                # DORA metrics scripts (🔬 Experimental)
│   ├── code-ownership/              # Code ownership scripts (🔬 Experimental)
│   ├── certificate-analyzer/        # Certificate analyzer scripts (🔬 Experimental)
│   ├── chalk-build-analyzer/        # Chalk analyzer scripts (🔬 Experimental)
│   ├── validation/                  # Validation and testing utilities
│   ├── lib/                         # Shared libraries (config-loader, etc.)
│   ├── config.example.json          # Global configuration template
│   └── CONFIG.md                    # Configuration system documentation
│
├── prompts/                         # Prompt templates & examples
│   ├── supply-chain/                # Supply chain prompts
│   ├── dora/                        # DORA metrics prompts
│   └── code-ownership/              # Code ownership prompts
│
├── docs/                            # Documentation
│   ├── guides/                      # How-to guides and tutorials
│   └── references/                  # Reference documentation
│
├── CHANGELOG.md                     # Central changelog for all components
└── .github/                         # GitHub workflows and templates
```

## Getting Started

### Prerequisites

- Access to Crash Override platform
- Basic familiarity with Claude skills and prompts

### Installation

1. Clone this repository:
   ```bash
   git clone https://github.com/crashappsec/skills-and-prompts-and-rag.git
   cd skills-and-prompts-and-rag
   ```

2. Make all scripts executable:
   ```bash
   chmod +x bootstrap.sh
   ./bootstrap.sh
   ```

3. Set up your environment:
   ```bash
   # Copy the environment template
   cp .env.example .env

   # Edit .env and add your Anthropic API key
   # ANTHROPIC_API_KEY=sk-ant-xxx...
   ```

   Get your Anthropic API key from [https://console.anthropic.com/](https://console.anthropic.com/)

4. Browse the available skills in the `skills/` directory

5. Each skill includes:
   - `.skill` file - The skill implementation
   - `README.md` - Comprehensive documentation
   - `CHANGELOG.md` - Version history
   - `examples/` - Usage examples

### Using a Skill

1. Navigate to the skill directory (e.g., `skills/certificate-analyzer/`)
2. Read the README.md for usage instructions
3. Load the `.skill` file into Crash Override
4. Follow the examples to get started

## Available Skills

### Supply Chain Analyzer 🚀 Beta
Comprehensive supply chain security analysis including SBOM/BOM management, vulnerability analysis, taint analysis, format conversion (CycloneDX ↔ SPDX), version upgrades, SLSA compliance assessment, and provenance verification.

**Status**: Beta - Feature-complete and tested, ready for broader use with active development. See [complete documentation](utils/supply-chain/README.md).

**Capabilities:**
- Vulnerability detection (OSV.dev, deps.dev, CISA KEV)
- Intelligent prioritization (data-driven CVSS + KEV scoring)
- Taint/reachability analysis to identify exploitable vulnerabilities
- Automatic SBOM generation with syft
- Dependency analysis and optimization
- License compliance checking
- Format conversion between CycloneDX and SPDX
- SBOM version upgrades (CycloneDX 1.7, SPDX 2.3)
- SLSA framework assessment (Levels 0-4)
- Supply chain security and provenance verification
- CI/CD integration with automation scripts

[View Skill Documentation](skills/supply-chain/README.md) | [View Utils Documentation](utils/supply-chain/)

### DORA Metrics 🔬 Experimental
Comprehensive DORA (DevOps Research and Assessment) metrics analysis for measuring and improving software delivery performance using the four key metrics.

**Status**: Experimental - Basic functionality working, under active development. Not yet ready for production use. See [roadmap](utils/dora-metrics/README.md#roadmap-to-production).

**Capabilities:**
- Calculate all four DORA metrics (Deployment Frequency, Lead Time, Change Failure Rate, MTTR)
- Performance classification (Elite, High, Medium, Low)
- Benchmark comparison against DORA research
- Root cause analysis and trend detection
- Team comparison and best practice identification
- Improvement roadmaps with prioritized recommendations
- Executive reporting and stakeholder communication
- CI/CD integration with automation scripts

[View Skill Documentation](skills/dora-metrics/README.md) | [View Utils Documentation](utils/dora-metrics/)

### Code Ownership Analyzer 🔬 Experimental
Comprehensive code ownership analysis for understanding who owns what code, validating CODEOWNERS files, identifying risks, and optimizing code review processes.

**Status**: Experimental - Single repository analysis working, needs multi-repo support and additional features. See [roadmap](utils/code-ownership/README.md#roadmap-to-production).

**Capabilities:**
- Analyze ownership patterns from git history with weighted scoring algorithms
- Validate and generate CODEOWNERS files (GitHub, GitLab, Bitbucket formats)
- Calculate ownership metrics (coverage, distribution, health scores)
- Identify single points of failure and bus factor risks
- Plan knowledge transfers for departing team members
- Recommend optimal PR reviewers based on ownership
- Track owner activity and detect staleness
- Generate actionable improvement recommendations

[View Skill Documentation](skills/code-ownership/README.md) | [View Utils Documentation](utils/code-ownership/)

### Certificate Analyzer 🔬 Experimental
Comprehensive TLS/SSL certificate analysis including validation, expiration checks, and security assessments.

**Status**: Experimental - Basic certificate validation working, needs bulk scanning and monitoring capabilities. See [roadmap](utils/certificate-analyzer/README.md#roadmap-to-production).

[View Skill Documentation](skills/certificate-analyzer/README.md) | [View Utils Documentation](utils/certificate-analyzer/)

### Chalk Build Analyzer 🔬 Experimental
Analyze and interpret Chalk build artifacts, providing insights into software supply chain metadata and build performance.

**Status**: Experimental - Chalk metadata extraction working, needs policy enforcement and multi-artifact support. See [roadmap](utils/chalk-build-analyzer/README.md#roadmap-to-production).

[View Skill Documentation](skills/chalk-build-analyzer/README.md) | [View Utils Documentation](utils/chalk-build-analyzer/)

### Better Prompts 🚀 Beta
Tools and techniques for crafting effective prompts, with before/after examples and conversation patterns.

**Status**: Beta - Comprehensive guide with proven techniques, ready for broad use.

[View Documentation](skills/better-prompts/README.md)

### COCOMO Estimator 📋 Planned
Software cost, effort, and schedule estimation using COCOMO II models with automated repository analysis.

**Status**: Planned - Specification complete, implementation pending. See [planned features](skills/cocomo/README.md).

[View Specification](skills/cocomo/README.md) | [View Utils Plan](utils/cocomo/)

## Roadmap

We're continuously expanding and improving this repository. Check out our [Roadmap](ROADMAP.md) to see:

- **Planned Features**: Upcoming skills and enhancements
- **Current Priorities**: What we're working on now
- **Community Requests**: Ideas from the community
- **How to Contribute**: Pick up a roadmap item and help build it

### Upcoming Skills

**Bus Factor Analysis** - Calculate project risk from knowledge concentration, identify critical dependencies, and recommend mitigation strategies

**Security Posture Assessment** - Comprehensive security analysis combining vulnerability management, compliance frameworks, and risk scoring

[View Full Roadmap](ROADMAP.md) | [Suggest a Feature](https://github.com/crashappsec/skills-and-prompts-and-rag/issues/new?template=roadmap_item.md)

## Contributing

We welcome contributions from the community! Whether you're:

- Creating new skills
- Improving existing prompts
- Adding documentation
- Fixing bugs
- Sharing examples

Please see our [Contributing Guide](CONTRIBUTING.md) for details on how to get started.

## RAG Knowledge Base

This repository is designed to serve as a knowledge base for RAG systems, providing:

- **Structured Documentation**: Consistent formatting for easy parsing
- **Domain Expertise**: Deep knowledge in security, development, and analysis
- **Real-World Examples**: Tested patterns and solutions
- **Living Documentation**: Continuously updated with community contributions

The organized structure makes it easy to:
- Index content for semantic search
- Retrieve relevant context for specific tasks
- Build specialized knowledge domains
- Enhance LLM responses with curated information

## Community

### Support

- **Issues**: Report bugs or request features via [GitHub Issues](https://github.com/crashappsec/skills-and-prompts-and-rag/issues)
- **Discussions**: Join conversations in [GitHub Discussions](https://github.com/crashappsec/skills-and-prompts-and-rag/discussions)
- **Code of Conduct**: Please read our [Code of Conduct](CODE_OF_CONDUCT.md)

### Recognition

Contributors are recognized in our documentation and release notes. Thank you to everyone who helps make this project better!

## License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.

This means you are free to:
- Use the software for any purpose
- Change the software to suit your needs
- Share the software with your friends and neighbors
- Share the changes you make

Under the conditions that:
- You must share your modifications under GPL-3.0
- You must include the original copyright and license
- You must state significant changes made to the software

---

**Made with ❤️ by the Crash Override community**
