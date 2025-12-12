# API Authentication & Authorization

Detection patterns for API authentication and authorization vulnerabilities.

## OWASP API Security Top 10

- **API1:2023** - Broken Object Level Authorization (BOLA)
- **API2:2023** - Broken Authentication
- **API5:2023** - Broken Function Level Authorization

## Patterns

### Missing Authentication Middleware

CATEGORY: api-auth
SEVERITY: critical
CONFIDENCE: 90
CWE: CWE-306
OWASP: API2:2023

Express.js routes without authentication:
```
PATTERN: app\.(get|post|put|delete|patch)\s*\(\s*['"][^'"]+['"]\s*,\s*(?!.*(?:auth|authenticate|isAuthenticated|requireAuth|verifyToken|passport|jwt|session)).*\(\s*req\s*,\s*res
LANGUAGES: javascript, typescript
```

FastAPI endpoints without dependencies:
```
PATTERN: @app\.(get|post|put|delete|patch)\s*\([^)]*\)\s*\n(?:async\s+)?def\s+\w+\s*\([^)]*\)(?!.*Depends)
LANGUAGES: python
```

Flask routes without login_required:
```
PATTERN: @app\.route\s*\([^)]+\)\s*\ndef\s+\w+\s*\((?!.*login_required|auth_required)
LANGUAGES: python
```

### Broken Object Level Authorization (BOLA)

CATEGORY: api-auth
SEVERITY: critical
CONFIDENCE: 85
CWE: CWE-639
OWASP: API1:2023

Direct object reference without ownership check:
```
PATTERN: \.(findById|findOne|findByPk|get)\s*\(\s*(req\.params\.|req\.query\.|request\.args)
LANGUAGES: javascript, typescript, python
```

User ID from request used directly in query:
```
PATTERN: (user_id|userId|user\.id)\s*=\s*(req\.params|req\.query|request\.args|request\.form)
LANGUAGES: javascript, typescript, python
```

### Missing JWT Validation

CATEGORY: api-auth
SEVERITY: high
CONFIDENCE: 90
CWE: CWE-287
OWASP: API2:2023

JWT decode without verification:
```
PATTERN: jwt\.decode\s*\([^)]*verify\s*=\s*False
LANGUAGES: python
```

JWT without algorithm specification:
```
PATTERN: jwt\.(sign|verify)\s*\([^)]*\)(?!.*algorithm)
LANGUAGES: javascript, typescript
```

Accepting 'none' algorithm:
```
PATTERN: algorithms?\s*[=:]\s*\[.*['"]none['"]
LANGUAGES: javascript, typescript, python
```

### Hardcoded JWT Secrets

CATEGORY: api-auth
SEVERITY: critical
CONFIDENCE: 95
CWE: CWE-798
OWASP: API2:2023

JWT secret in code:
```
PATTERN: (jwt_secret|JWT_SECRET|secret_key|SECRET_KEY)\s*[=:]\s*['"][^'"]{8,}['"]
LANGUAGES: javascript, typescript, python
```

### Session Fixation

CATEGORY: api-auth
SEVERITY: high
CONFIDENCE: 85
CWE: CWE-384
OWASP: API2:2023

Session not regenerated after login:
```
PATTERN: (login|authenticate|signin).*\{[^}]*(?!regenerate|destroy|create).*session
LANGUAGES: javascript, typescript
```

### Missing CORS Authentication

CATEGORY: api-auth
SEVERITY: medium
CONFIDENCE: 80
CWE: CWE-942
OWASP: API2:2023

CORS with credentials but wildcard origin:
```
PATTERN: (credentials|withCredentials)\s*[=:]\s*true.*origin\s*[=:]\s*['"]\*['"]
LANGUAGES: javascript, typescript, python
```

### Broken Function Level Authorization

CATEGORY: api-auth
SEVERITY: high
CONFIDENCE: 85
CWE: CWE-285
OWASP: API5:2023

Admin endpoints without role check:
```
PATTERN: (\/admin|\/manage|\/internal|\/system).*(?!role|permission|isAdmin|authorize)
LANGUAGES: javascript, typescript, python
```

### API Key in URL

CATEGORY: api-auth
SEVERITY: medium
CONFIDENCE: 90
CWE: CWE-598
OWASP: API2:2023

API key passed in query string:
```
PATTERN: (api_key|apikey|api-key|access_token)\s*=\s*(req\.query|request\.args|params\[)
LANGUAGES: javascript, typescript, python
```

## References

- [OWASP API Security Top 10 2023](https://owasp.org/API-Security/editions/2023/en/0x11-t10/)
- [CWE-306: Missing Authentication](https://cwe.mitre.org/data/definitions/306.html)
- [CWE-639: Insecure Direct Object Reference](https://cwe.mitre.org/data/definitions/639.html)
