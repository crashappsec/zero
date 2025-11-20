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
├── skills/                    # Claude skills (.skill files)
│   ├── chalk-build-analyzer/  # Analyze Chalk build artifacts
│   ├── certificate-analyzer/  # TLS/SSL certificate analysis
│   └── prompt-engineering/    # Prompt crafting and optimization
│
├── prompts/                   # Prompt templates & examples
│   ├── security/              # Security-focused prompts
│   ├── development/           # Development and coding prompts
│   └── analysis/              # Analysis and investigation prompts
│
├── tools/                     # Scripts and utilities
│   ├── git-sync/              # Repository synchronization tools
│   └── validation/            # Validation and testing utilities
│
├── docs/                      # Documentation
│   ├── guides/                # How-to guides and tutorials
│   └── references/            # Reference documentation
│
└── .github/                   # GitHub workflows and templates
```

## Getting Started

### Prerequisites

- Access to Crash Override platform
- Basic familiarity with Claude skills and prompts

### Installation

1. Clone this repository:
   ```bash
   git clone https://github.com/yourusername/skills-and-prompts.git
   cd skills-and-prompts
   ```

2. Browse the available skills in the `skills/` directory

3. Each skill includes:
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

### Chalk Build Analyzer
Analyze and interpret Chalk build artifacts, providing insights into software supply chain metadata.

[View Documentation](skills/chalk-build-analyzer/README.md)

### Certificate Analyzer
Comprehensive TLS/SSL certificate analysis including validation, expiration checks, and security assessments.

[View Documentation](skills/certificate-analyzer/README.md)

### Prompt Engineering
Tools and techniques for crafting effective prompts, with before/after examples and conversation patterns.

[View Documentation](skills/prompt-engineering/README.md)

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

- **Issues**: Report bugs or request features via [GitHub Issues](https://github.com/yourusername/skills-and-prompts/issues)
- **Discussions**: Join conversations in [GitHub Discussions](https://github.com/yourusername/skills-and-prompts/discussions)
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
