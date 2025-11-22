<!--
Copyright (c) 2024 Gibson Powers Contributors

SPDX-License-Identifier: GPL-3.0
-->

# Contributing to Gibson Powers

Thank you for your interest in contributing to the Gibson Powers repository! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
- [Development Process](#development-process)
- [Skill Guidelines](#skill-guidelines)
- [Prompt Guidelines](#prompt-guidelines)
- [Documentation Standards](#documentation-standards)
- [Pull Request Process](#pull-request-process)
- [Community](#community)

## Code of Conduct

This project adheres to a Code of Conduct that all contributors are expected to follow. Please read [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) before contributing.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check existing issues to avoid duplicates. When creating a bug report, include:

- **Clear title and description**
- **Steps to reproduce** the issue
- **Expected behavior** vs actual behavior
- **Screenshots** if applicable
- **Environment details** (OS, Gibson Powers version, etc.)

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, include:

- **Clear use case** - Why is this enhancement needed?
- **Proposed solution** - How should it work?
- **Alternatives considered** - What other approaches did you think about?
- **Impact** - Who benefits from this enhancement?

### Contributing Code

1. **Skills** - New skills or improvements to existing skills
2. **Prompts** - Tested prompt templates for common use cases
3. **Tools** - Utilities that enhance the workflow
4. **Documentation** - Guides, examples, and references
5. **Tests** - Validation scripts and test cases

## Development Process

### Getting Started

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/yourusername/gibson-powers.git
   cd gibson-powers
   ```
3. Create a feature branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```

### Making Changes

1. Make your changes following our guidelines (see below)
2. Test your changes thoroughly
3. Update documentation as needed
4. Commit with clear, descriptive messages

### Commit Message Format

Use conventional commit format:

```
type(scope): brief description

Detailed explanation of changes if needed.

Closes #issue-number
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Formatting changes
- `refactor`: Code restructuring
- `test`: Adding tests
- `chore`: Maintenance tasks

**Examples:**
```
feat(skills): add DNS analyzer skill

Implements comprehensive DNS record analysis including:
- A/AAAA record validation
- MX record checking
- DNSSEC validation

Closes #42
```

## Skill Guidelines

### Skill Structure

Each skill must follow this structure:

```
skills/skill-name/
├── skill-name.skill          # The skill implementation
├── README.md                 # Documentation
├── CHANGELOG.md              # Version history
└── examples/                 # Usage examples
    └── example-1.md
```

### Skill Requirements

1. **Clear Purpose**: The skill should solve a specific, well-defined problem
2. **Documentation**: Comprehensive README with:
   - Purpose and use cases
   - Prerequisites
   - Usage instructions
   - Examples
   - Troubleshooting
3. **Examples**: At least one working example in the `examples/` directory
4. **CHANGELOG**: Track all versions and changes
5. **Testing**: Verify the skill works as expected

### Skill README Template

```markdown
# Skill Name

Brief description of what the skill does.

## Purpose

Detailed explanation of the problem this skill solves.

## Prerequisites

- Required tools or access
- Knowledge requirements

## Usage

Step-by-step instructions for using the skill.

## Examples

Link to examples in the examples/ directory.

## Troubleshooting

Common issues and solutions.

## Contributing

How to contribute improvements to this skill.

## License

Reference to repository license.
```

## Prompt Guidelines

### Prompt Organization

Place prompts in the appropriate category:
- `prompts/security/` - Security analysis and testing
- `prompts/development/` - Coding and development
- `prompts/analysis/` - Data and system analysis

### Prompt Requirements

1. **Clear Context**: Include necessary background information
2. **Specific Instructions**: Be precise about what you want
3. **Expected Output**: Describe the desired format and content
4. **Tested**: Verify the prompt produces good results
5. **Documented**: Include metadata:
   - Purpose
   - When to use
   - Expected output
   - Variations

### Prompt Template

```markdown
# Prompt: [Name]

## Purpose
What this prompt is designed to do.

## When to Use
Scenarios where this prompt is most effective.

## Prompt

\`\`\`
[The actual prompt text]
\`\`\`

## Expected Output
Description of what the response should look like.

## Variations
Alternative versions for different use cases.

## Examples
Real-world usage examples.
```

## Documentation Standards

### General Guidelines

- Use clear, concise language
- Include code examples where helpful
- Keep formatting consistent
- Link to related resources
- Update the table of contents

### Markdown Style

- Use ATX-style headers (`#`, `##`, etc.)
- Use fenced code blocks with language tags
- Use relative links for internal references
- Include alt text for images
- Keep line length reasonable (80-100 chars when practical)

## Pull Request Process

### Before Submitting

1. Test your changes thoroughly
2. Update relevant documentation
3. Add or update examples if needed
4. Ensure your code follows existing style
5. Rebase on the latest main branch

### Submitting a Pull Request

1. Push your changes to your fork
2. Create a pull request with:
   - **Clear title** describing the change
   - **Description** explaining what and why
   - **Related issues** referenced with #issue-number
   - **Testing performed** to verify changes
   - **Screenshots** if UI-related

### Pull Request Template

```markdown
## Description
Brief description of changes.

## Motivation and Context
Why is this change needed? What problem does it solve?

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Documentation update
- [ ] Performance improvement
- [ ] Other (please describe)

## How Has This Been Tested?
Describe testing performed.

## Checklist
- [ ] My code follows the project style
- [ ] I have updated documentation
- [ ] I have added tests (if applicable)
- [ ] All tests pass
- [ ] I have updated CHANGELOG.md

## Related Issues
Closes #issue-number
```

### Review Process

1. Maintainers will review your PR
2. Address any requested changes
3. Once approved, a maintainer will merge

### After Merge

- Your contribution will be included in the next release
- You'll be added to the contributors list
- Thank you for making this project better!

## Community

### Getting Help

- **GitHub Discussions**: Ask questions and share ideas
- **Issues**: Report bugs and request features
- **Documentation**: Check the docs/ directory

### Recognition

Contributors are recognized in:
- Release notes
- Contributors list
- Project documentation

## Questions?

If you have questions about contributing, please:
1. Check existing documentation
2. Search closed issues
3. Open a new discussion
4. Reach out to maintainers

Thank you for contributing to Gibson Powers!
