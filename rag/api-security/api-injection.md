# API Injection Vulnerabilities

Detection patterns for injection vulnerabilities in API endpoints.

## OWASP API Security Top 10

- **API8:2023** - Security Misconfiguration (includes injection via misconfigured parsers)
- **API10:2023** - Unsafe Consumption of APIs

## Patterns

### SQL Injection in APIs

CATEGORY: api-injection
SEVERITY: critical
CONFIDENCE: 95
CWE: CWE-89
OWASP: A03:2021

String concatenation in SQL with request params:
```
PATTERN: (execute|query|raw)\s*\([^)]*(\+|`\$\{|\.format\(|%s|%d).*req\.(body|params|query)
LANGUAGES: javascript, typescript, python
```

Template literals in SQL:
```
PATTERN: (execute|query|raw)\s*\(`[^`]*\$\{.*req\.(body|params|query)
LANGUAGES: javascript, typescript
```

Python f-string SQL:
```
PATTERN: (execute|cursor\.execute)\s*\(\s*f['"][^'"]*\{.*request\.(args|form|json)
LANGUAGES: python
```

### NoSQL Injection

CATEGORY: api-injection
SEVERITY: critical
CONFIDENCE: 90
CWE: CWE-943
OWASP: A03:2021

MongoDB query with unsanitized input:
```
PATTERN: \.(find|findOne|findOneAndUpdate|updateOne|deleteOne)\s*\(\s*\{[^}]*req\.(body|params|query)
LANGUAGES: javascript, typescript
```

MongoDB $where with user input:
```
PATTERN: \$where\s*[=:]\s*.*req\.(body|params|query)
LANGUAGES: javascript, typescript
```

PyMongo with user input:
```
PATTERN: (find|find_one|update_one|delete_one)\s*\(\s*\{.*request\.(args|form|json)
LANGUAGES: python
```

### Command Injection in APIs

CATEGORY: api-injection
SEVERITY: critical
CONFIDENCE: 95
CWE: CWE-78
OWASP: A03:2021

Shell execution with request data:
```
PATTERN: (exec|spawn|execSync|execFile|system|popen|subprocess)\s*\([^)]*req\.(body|params|query)
LANGUAGES: javascript, typescript, python
```

Backtick command with user input:
```
PATTERN: `[^`]*\$\{.*req\.(body|params|query).*\}[^`]*`
LANGUAGES: javascript, typescript
```

### LDAP Injection

CATEGORY: api-injection
SEVERITY: high
CONFIDENCE: 85
CWE: CWE-90
OWASP: A03:2021

LDAP filter with user input:
```
PATTERN: (search|bind)\s*\([^)]*(\+|`\$\{|\.format\(|%s).*req\.(body|params|query)
LANGUAGES: javascript, typescript, python
```

### XPath Injection

CATEGORY: api-injection
SEVERITY: high
CONFIDENCE: 85
CWE: CWE-643
OWASP: A03:2021

XPath query with concatenation:
```
PATTERN: (xpath|evaluate|selectNodes)\s*\([^)]*(\+|`\$\{|\.format\().*req\.(body|params|query)
LANGUAGES: javascript, typescript, python
```

### Header Injection

CATEGORY: api-injection
SEVERITY: high
CONFIDENCE: 90
CWE: CWE-113
OWASP: A03:2021

Response header with user input:
```
PATTERN: (setHeader|set|header)\s*\([^,]+,\s*req\.(body|params|query)
LANGUAGES: javascript, typescript
```

### Log Injection

CATEGORY: api-injection
SEVERITY: medium
CONFIDENCE: 80
CWE: CWE-117
OWASP: A09:2021

Unsanitized user input in logs:
```
PATTERN: (logger\.|console\.|log\.)(info|warn|error|debug)\s*\([^)]*req\.(body|params|query)\.[^)]*\)
LANGUAGES: javascript, typescript, python
```

### Template Injection (SSTI)

CATEGORY: api-injection
SEVERITY: critical
CONFIDENCE: 90
CWE: CWE-1336
OWASP: A03:2021

Server-side template with user input:
```
PATTERN: (render_template_string|Template|render_string)\s*\([^)]*req\.(body|params|query)
LANGUAGES: python
```

Jinja2 with autoescape disabled:
```
PATTERN: Environment\s*\([^)]*autoescape\s*=\s*False
LANGUAGES: python
```

### GraphQL Injection

CATEGORY: api-injection
SEVERITY: high
CONFIDENCE: 85
CWE: CWE-89
OWASP: A03:2021

GraphQL query string concatenation:
```
PATTERN: (graphql|query)\s*[=:]\s*[`'"].*\$\{.*req\.(body|params|query)
LANGUAGES: javascript, typescript
```

### ORM Injection

CATEGORY: api-injection
SEVERITY: high
CONFIDENCE: 85
CWE: CWE-89
OWASP: A03:2021

Sequelize literal with user input:
```
PATTERN: Sequelize\.literal\s*\([^)]*req\.(body|params|query)
LANGUAGES: javascript, typescript
```

SQLAlchemy text with user input:
```
PATTERN: (text|literal_column)\s*\([^)]*request\.(args|form|json)
LANGUAGES: python
```

## References

- [OWASP Injection Prevention Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Injection_Prevention_Cheat_Sheet.html)
- [CWE-89: SQL Injection](https://cwe.mitre.org/data/definitions/89.html)
- [CWE-78: OS Command Injection](https://cwe.mitre.org/data/definitions/78.html)
