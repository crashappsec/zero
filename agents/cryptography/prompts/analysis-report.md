# Cipher Analysis Report Template

## Executive Summary

Provide a brief overview:
- Overall cryptographic security posture (critical/high/medium/low risk)
- Key findings count by severity
- Most urgent issues requiring immediate attention

## Critical Findings

List any critical severity findings that require immediate action:

### [Finding Title]
- **Type**: (e.g., Broken Cipher, Exposed Private Key, Disabled Cert Verification)
- **CWE**: CWE-XXX
- **Location**: `file:line`
- **Impact**: What could an attacker do with this vulnerability?
- **Remediation**: Specific steps to fix, with code examples if helpful

## High Severity Findings

### [Finding Title]
- **Type**: (e.g., Deprecated Hash, Weak Key Length, Hardcoded Key)
- **CWE**: CWE-XXX
- **Location**: `file:line`
- **Impact**: Description of the risk
- **Remediation**: How to fix

## Medium/Low Findings

Summarize less critical findings in a table:

| Type | Count | Files Affected | Recommendation |
|------|-------|----------------|----------------|
| MD5 hash usage | 3 | 2 | Replace with SHA-256 |
| Insecure random | 5 | 3 | Review if security-sensitive |

## Recommendations

### Immediate Actions (This Sprint)
1. Action item with specific guidance
2. Action item with specific guidance

### Short-term Improvements (This Quarter)
1. Action item with specific guidance
2. Action item with specific guidance

### Long-term Hardening
1. Consider implementing...
2. Plan migration to...

## Compliance Notes

If relevant, note compliance implications:
- **PCI DSS**: Requires TLS 1.2+, no weak ciphers
- **SOC 2**: Encryption at rest and in transit requirements
- **HIPAA**: PHI must be encrypted with strong algorithms

## References

- NIST Cryptographic Standards
- OWASP Cryptographic Failures
- Relevant CVEs or advisories
