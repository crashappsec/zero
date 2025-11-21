<!--
Copyright (c) 2024 Crash Override Inc
101 Fulton St, 416, New York 10038
SPDX-License-Identifier: GPL-3.0
-->

# Validation Tools

This directory contains validation scripts to ensure repository standards are maintained.

## Available Scripts

### check-copyright.sh

Validates that all source files have proper copyright headers.

**Usage:**
```bash
# Check for missing copyright headers
./check-copyright.sh

# Automatically add missing copyright headers
./check-copyright.sh --fix
```

**Features:**
- Checks all `.md` and `.sh` files
- Excludes binary files and dependencies
- Can automatically add missing headers
- Handles shell script shebangs correctly

### check-commit-message.sh

Validates commit messages follow conventional commit format.

**Usage:**
```bash
# Check a commit message file
./check-commit-message.sh .git/COMMIT_EDITMSG

# Check the last commit
./check-commit-message.sh --check-last
```

**Validates:**
- Conventional commit format: `type(scope): description`
- Valid types: feat, fix, docs, style, refactor, test, chore, revert
- Message length (warns if > 100 characters)
- No period at end
- Lowercase description

## Commit Message Format

All commits should follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Types

- **feat**: A new feature
- **fix**: A bug fix
- **docs**: Documentation only changes
- **style**: Changes that don't affect code meaning (formatting, etc.)
- **refactor**: Code change that neither fixes a bug nor adds a feature
- **test**: Adding missing tests or correcting existing tests
- **chore**: Changes to build process or auxiliary tools

### Examples

```
feat: add certificate chain validation
fix(cert-analyzer): correct expiration date parsing
docs: update installation instructions
style: format code with prettier
refactor(skills): extract common validation logic
test: add unit tests for certificate parser
chore: update dependencies
```

## Copyright Header Format

All source files must include a copyright header:

**For Markdown files (.md):**
```markdown
<!--
Copyright (c) 2024 Crash Override Inc
101 Fulton St, 416, New York 10038
SPDX-License-Identifier: GPL-3.0
-->
```

**For Shell scripts (.sh):**
```bash
#!/bin/bash
# Copyright (c) 2024 Crash Override Inc
# 101 Fulton St, 416, New York 10038
# SPDX-License-Identifier: GPL-3.0
```

## Setting Up Git Hooks

To automatically validate commits, you can set up git hooks:

### Commit Message Hook

Create `.git/hooks/commit-msg`:
```bash
#!/bin/bash
./tools/validation/check-commit-message.sh "$1"
```

Make it executable:
```bash
chmod +x .git/hooks/commit-msg
```

### Pre-Commit Hook

Create `.git/hooks/pre-commit`:
```bash
#!/bin/bash
./tools/validation/check-copyright.sh
```

Make it executable:
```bash
chmod +x .git/hooks/pre-commit
```

## CI Integration

These scripts can be integrated into CI/CD pipelines:

```yaml
# Example GitHub Actions workflow
name: Validate Standards
on: [push, pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Check Copyright Headers
        run: ./tools/validation/check-copyright.sh
      - name: Check Commit Messages
        run: |
          for commit in $(git rev-list origin/main..HEAD); do
            git log -1 --pretty=%B $commit | ./tools/validation/check-commit-message.sh /dev/stdin
          done
```

## Troubleshooting

### "Permission denied" errors

Make scripts executable:
```bash
chmod +x tools/validation/*.sh
```

### False positives

If certain files shouldn't have copyright headers, add them to the `EXCLUDE_PATTERNS` array in `check-copyright.sh`.

## Contributing

When adding new validation rules:
1. Update the relevant script
2. Document the rule in this README
3. Add examples of valid/invalid cases
4. Test thoroughly before committing
