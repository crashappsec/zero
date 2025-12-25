# Helm Chart Best Practices Patterns

**Category**: devops/iac-best-practices
**Description**: Helm chart organizational and operational best practices
**Type**: best-practice

---

## Chart.yaml Best Practices

### Missing Chart Description
**Type**: regex
**Severity**: low
**Category**: best-practice
**Pattern**: `name:\s*[^\n]+\n(?:(?!description:).)*version:`
- Charts should have a description in Chart.yaml
- Remediation: Add `description: "What this chart deploys"`

### Missing Chart Keywords
**Type**: regex
**Severity**: low
**Category**: best-practice
**Pattern**: `version:\s*[^\n]+\n(?:(?!keywords:).)*$`
- Consider adding keywords for discoverability
- Remediation: Add `keywords: [keyword1, keyword2]`

### Missing Maintainers
**Type**: regex
**Severity**: low
**Category**: best-practice
**Pattern**: `version:\s*[^\n]+\n(?:(?!maintainers:).)*$`
- Charts should list maintainers for support
- Remediation: Add `maintainers: [{ name: "Name", email: "email" }]`

---

## Values.yaml Best Practices

### Uncommented Values
**Type**: structural
**Severity**: low
**Category**: best-practice
- Values should have comments explaining purpose and options
- Remediation: Add comments above each configurable value

### Missing Image Tag Configuration
**Type**: regex
**Severity**: medium
**Category**: best-practice
**Pattern**: `image:\s*\n\s+repository:[^\n]+\n(?:(?!tag:).)*\n`
- Image configuration should include separate tag field
- Remediation: Add `tag: "latest"` for versioning flexibility

### Missing Resource Defaults
**Type**: regex
**Severity**: medium
**Category**: best-practice
**Pattern**: `^(?:(?!resources:).)*$`
- values.yaml should define default resource requests/limits
- Remediation: Add `resources: { requests: {...}, limits: {...} }`

---

## Template Best Practices

### Missing NOTES.txt
**Type**: structural
**Severity**: low
**Category**: best-practice
- Charts should include templates/NOTES.txt for post-install guidance
- Remediation: Create NOTES.txt with usage instructions

### Missing Helper Functions
**Type**: structural
**Severity**: low
**Category**: best-practice
- Charts should use _helpers.tpl for reusable template functions
- Remediation: Define common labels and name functions in _helpers.tpl

### Hardcoded Values in Templates
**Type**: regex
**Severity**: medium
**Category**: best-practice
**Pattern**: `replicas:\s*\d+(?!\s*\||\s*\})`
- Avoid hardcoded values; use .Values references
- Example: `replicas: 3` (bad)
- Remediation: Use `replicas: {{ .Values.replicaCount }}` (good)

---

## Security Best Practices

### Missing Service Account Configuration
**Type**: regex
**Severity**: medium
**Category**: best-practice
**Pattern**: `^values\.yaml$(?:(?!serviceAccount:).)*$`
- Charts should allow service account configuration
- Remediation: Add `serviceAccount: { create: true, name: "" }`

### Missing Pod Security Context
**Type**: regex
**Severity**: medium
**Category**: best-practice
**Pattern**: `^values\.yaml$(?:(?!podSecurityContext:).)*$`
- Charts should define pod security context options
- Remediation: Add `podSecurityContext: { runAsNonRoot: true }`
