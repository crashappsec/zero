<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Prompts

This directory contains prompt templates and examples organized by category.

## Structure

```
prompts/
├── certificate-analysis/     # X.509 certificate analysis prompts
│   ├── security/            # Security audits, chain validation
│   ├── compliance/          # CA/B Forum, expiry monitoring
│   ├── operations/          # Format conversion, comparison
│   └── troubleshooting/     # TLS issue diagnosis
├── code-ownership/          # Code ownership analysis prompts
├── dora/                    # DORA metrics prompts
├── legal-review/            # License compliance prompts
├── supply-chain/            # SBOM and supply chain prompts
│   ├── security/            # Vulnerability scanning
│   ├── operations/          # SBOM management
│   └── compliance/          # License compliance
└── technology-identification/ # Technology detection prompts
```

## Categories

### Certificate Analysis
X.509 certificate security, compliance, and troubleshooting.
- [certificate-analysis/README.md](certificate-analysis/README.md)

### Code Ownership
CODEOWNERS analysis and team coverage validation.
- [code-ownership/README.md](code-ownership/README.md)

### DORA Metrics
DevOps Research and Assessment metrics analysis.
- [dora/README.md](dora/README.md)

### Legal Review
Software license compliance and legal analysis.
- [legal-review/README.md](legal-review/README.md)

### Supply Chain
SBOM management and supply chain security.
- [supply-chain/README.md](supply-chain/README.md)

### Technology Identification
Technology stack detection and analysis.
- [technology-identification/](technology-identification/)

## Usage

1. Browse the subdirectories to find prompt templates for your specific use case
2. Select a prompt template from the appropriate category
3. Customize the template with your specific inputs
4. Execute with Claude Code or the relevant skill

## Related Resources

- [RAG Knowledge Base](../rag/README.md) - Reference documentation
- [Skills](../skills/README.md) - Tool integrations
- [Utils](../utils/README.md) - Command-line tools
