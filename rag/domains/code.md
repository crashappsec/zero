# Code Security Domain Knowledge

This document consolidates RAG knowledge for the **code** super scanner.

## Features Covered
- **vulns**: Code vulnerability detection (SAST)
- **secrets**: Secret/credential detection
- **api**: API security analysis
- **tech_debt**: Technical debt and code quality

## Related RAG Directories

### Code Security
- `rag/code-security/` - Core code security knowledge
  - OWASP Top 10 vulnerabilities
  - CWE mappings
  - Language-specific security patterns

### API Security
- `rag/api-security/` - API security knowledge
  - OWASP API Security Top 10
  - Authentication patterns
  - Input validation

### Secrets
- `rag/secrets-scanner/` - Secret detection knowledge
  - API key patterns
  - Credential detection
  - Token identification

### Technical Debt
- `rag/tech-debt/` - Technical debt knowledge
  - TODO/FIXME markers
  - Code complexity metrics
  - Maintainability issues

### Semgrep
- `rag/semgrep/` - Semgrep rule knowledge
  - Rule syntax
  - Language support
  - Custom rule development

## Key Concepts

### OWASP Top 10 (2021)
1. **A01 Broken Access Control** - Authorization flaws
2. **A02 Cryptographic Failures** - See crypto scanner
3. **A03 Injection** - SQL, NoSQL, OS, LDAP injection
4. **A04 Insecure Design** - Missing security controls
5. **A05 Security Misconfiguration** - Default configs, verbose errors
6. **A06 Vulnerable Components** - See packages scanner
7. **A07 Auth Failures** - Session management, credentials
8. **A08 Data Integrity Failures** - Insecure deserialization
9. **A09 Logging Failures** - Insufficient logging
10. **A10 SSRF** - Server-side request forgery

### OWASP API Security Top 10 (2023)
1. **API1 Broken Object Level Auth** - IDOR vulnerabilities
2. **API2 Broken Authentication** - Weak auth mechanisms
3. **API3 Broken Object Property Auth** - Mass assignment
4. **API4 Unrestricted Resource Consumption** - Rate limiting
5. **API5 Broken Function Level Auth** - Admin function access
6. **API6 SSRF** - Server-side request forgery
7. **API7 Security Misconfiguration** - CORS, headers
8. **API8 Lack of Protection from Automated Threats** - Bot protection
9. **API9 Improper Inventory Management** - Shadow APIs
10. **API10 Unsafe Consumption of APIs** - Third-party API trust

### Common CWE Categories
| CWE | Name | Severity |
|-----|------|----------|
| CWE-89 | SQL Injection | Critical |
| CWE-79 | XSS | High |
| CWE-78 | OS Command Injection | Critical |
| CWE-22 | Path Traversal | High |
| CWE-502 | Deserialization | Critical |
| CWE-918 | SSRF | High |
| CWE-352 | CSRF | Medium |
| CWE-798 | Hardcoded Credentials | High |

### Technical Debt Markers
- **TODO**: Future work items
- **FIXME**: Known bugs/issues
- **HACK**: Workarounds
- **XXX**: Dangerous code
- **BUG**: Known defects
- **DEPRECATED**: Obsolete code

## Agent Expertise

### Razor Agent
The **Razor** agent (code security specialist) should be consulted for:
- SAST finding analysis
- Vulnerability triage
- Remediation guidance
- Security code review

### Dade Agent
The **Dade** agent (backend engineer) may assist with:
- Backend vulnerability context
- API security patterns
- Database security

### Acid Agent
The **Acid** agent (frontend engineer) may assist with:
- XSS vulnerability analysis
- Frontend security patterns
- Client-side validation

## Output Schema

The code scanner produces a single `code.json` file with:
```json
{
  "features_run": ["vulns", "secrets", "api", "tech_debt"],
  "summary": {
    "vulns": { "total_findings": N, "critical": N, "high": N, ... },
    "secrets": { "total_findings": N, "risk_score": N, ... },
    "api": { "total_findings": N, ... },
    "tech_debt": { "total_markers": N, "by_type": {...} }
  },
  "findings": {
    "vulns": [...],
    "secrets": [...],
    "api": [...],
    "tech_debt": { "markers": [...], "hotspots": [...] }
  }
}
```

## Severity Classification

| Finding Type | Critical | High | Medium | Low |
|--------------|----------|------|--------|-----|
| Vulnerability | RCE, SQLi, Deser | XSS, Path Traversal | CSRF, Info Disclosure | Code Quality |
| Secrets | Private keys | API keys, tokens | Test credentials | Generic passwords |
| API | Auth bypass | IDOR, SSRF | Rate limit missing | CORS config |
| Tech Debt | - | Security TODOs | Bug markers | Info TODOs |

## Detection Tools

### Semgrep Rulesets
- `p/security-audit` - General security rules
- `p/owasp-top-ten` - OWASP coverage
- `p/secrets` - Secret detection
- Custom rules for API security

### Languages Supported
- JavaScript/TypeScript
- Python
- Go
- Java
- Ruby
- PHP
- C/C++
- Rust
