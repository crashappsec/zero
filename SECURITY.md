<!--
Copyright (c) 2024 Gibson Powers Contributors

SPDX-License-Identifier: GPL-3.0
-->

# Security Policy

## Supported Versions

We actively maintain and provide security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| main    | :white_check_mark: |

As this is an early-stage project, we currently only support the main branch. Once we establish versioned releases, this table will be updated accordingly.

## Reporting a Vulnerability

We take the security of Gibson Powers seriously. If you believe you have found a security vulnerability, please report it to us as described below.

### Where to Report

**Please do NOT report security vulnerabilities through public GitHub issues.**

Instead, please report them via:
- Email: mark@crashoverride.com
- GitHub Security Advisories: [Use the "Report a vulnerability" button](../../security/advisories/new)

### What to Include

Please include the following information in your report:

- **Description**: A clear description of the vulnerability
- **Impact**: What could an attacker accomplish by exploiting this?
- **Reproduction Steps**: Detailed steps to reproduce the vulnerability
- **Affected Components**: Which skills, prompts, or tools are affected?
- **Suggested Fix**: If you have ideas for how to fix it (optional)
- **Your Contact Information**: So we can follow up with questions

### What to Expect

When you report a vulnerability, you can expect:

1. **Acknowledgment**: We'll acknowledge receipt within 48 hours
2. **Assessment**: We'll assess the vulnerability and its impact
3. **Communication**: We'll keep you informed of our progress
4. **Resolution**: We'll work to fix verified vulnerabilities promptly
5. **Credit**: We'll credit you in the security advisory (unless you prefer to remain anonymous)

### Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Fix Timeline**: Depends on severity
  - Critical: Within 7 days
  - High: Within 14 days
  - Medium: Within 30 days
  - Low: Within 90 days

## Security Best Practices

### For Contributors

When contributing to this repository:

1. **Review Code**: Look for common security issues
   - Injection vulnerabilities
   - Insecure dependencies
   - Hardcoded secrets or credentials
   - Improper input validation

2. **Dependencies**: Keep dependencies updated and review for known vulnerabilities

3. **Secrets**: Never commit:
   - API keys
   - Passwords
   - Private keys
   - Access tokens
   - Any sensitive configuration

4. **Prompts**: When creating prompts, consider:
   - Prompt injection attacks
   - Data leakage through examples
   - Unintended command execution

### For Users

When using skills and prompts from this repository:

1. **Review First**: Always review skills and prompts before using them
2. **Test Safely**: Test in a safe environment before production use
3. **Understand Scope**: Know what permissions and access each skill requires
4. **Stay Updated**: Keep your local copy updated with security fixes
5. **Report Issues**: If you find something concerning, report it

## Known Security Considerations

### Skill Execution

Skills in this repository may:
- Execute commands on your system
- Access local files and directories
- Make network requests
- Process sensitive data

**Always review skills before executing them and understand what they do.**

### Prompt Injection

Prompts that accept user input may be vulnerable to prompt injection attacks. When using or creating prompts:

- Validate and sanitize user input
- Use clear delimiters between instructions and data
- Be cautious with dynamic prompt generation
- Test with malicious inputs

### Data Privacy

When using skills and prompts:

- Be aware of what data you're sharing
- Don't include sensitive information in prompts unless necessary
- Understand where your data is being processed
- Review privacy implications of each skill

## Security Updates

Security updates will be:
- Published as GitHub Security Advisories
- Documented in CHANGELOG.md
- Announced in release notes
- Tagged with `[SECURITY]` prefix in commits

## Disclosure Policy

We follow coordinated disclosure:

1. Security issues are reported privately
2. We work with reporters to understand and fix issues
3. Fixes are developed and tested
4. Public disclosure occurs after fix is released
5. Credit is given to reporters (with permission)

## Scope

### In Scope

- Skills in the `skills/` directory
- Prompts in the `prompts/` directory
- Tools in the `tools/` directory
- Documentation that could lead to security issues
- Dependencies with known vulnerabilities

### Out of Scope

- Issues in the Crash Override platform itself (report to Crash Override team)
- Social engineering attacks
- Physical security issues
- Issues requiring physical access to systems

## Questions?

If you have questions about this security policy, please:
- Open a discussion in GitHub Discussions (for general questions)
- Email mark@crashoverride.com (for security-related questions)

## Recognition

We appreciate security researchers who help keep Gibson Powers safe. With your permission, we will:
- Credit you in the security advisory
- List you in our security acknowledgments
- Provide a reference for your responsible disclosure

Thank you for helping keep Gibson Powers and our community safe!
