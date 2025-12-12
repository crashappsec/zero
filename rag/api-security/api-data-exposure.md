# API Data Exposure

Detection patterns for excessive data exposure and sensitive data leakage in APIs.

## OWASP API Security Top 10

- **API3:2023** - Broken Object Property Level Authorization
- **API6:2023** - Unrestricted Access to Sensitive Business Flows

## Patterns

### Excessive Data Exposure

CATEGORY: api-data-exposure
SEVERITY: high
CONFIDENCE: 85
CWE: CWE-200
OWASP: API3:2023

Returning full user object without filtering:
```
PATTERN: res\.(json|send)\s*\(\s*(user|account|profile|customer)\s*\)
LANGUAGES: javascript, typescript
```

Select * in API response:
```
PATTERN: (SELECT|select)\s+\*\s+FROM.*(res\.|return|response)
LANGUAGES: javascript, typescript, python
```

Mongoose/Sequelize find without select/attributes:
```
PATTERN: \.(find|findOne|findAll|findById)\s*\([^)]*\)(?!.*\.(select|lean|attributes))
LANGUAGES: javascript, typescript
```

### Password in Response

CATEGORY: api-data-exposure
SEVERITY: critical
CONFIDENCE: 95
CWE: CWE-200
OWASP: API3:2023

Password field not excluded:
```
PATTERN: (toJSON|toObject)\s*\([^)]*\)(?!.*password).*res\.(json|send)
LANGUAGES: javascript, typescript
```

Password hash in response:
```
PATTERN: (password|password_hash|passwordHash|hashed_password)\s*[=:][^,}]*[,}][^}]*(res\.|return|response)
LANGUAGES: javascript, typescript, python
```

### Sensitive Fields Exposure

CATEGORY: api-data-exposure
SEVERITY: high
CONFIDENCE: 90
CWE: CWE-200
OWASP: API3:2023

SSN, credit card, or PII in response:
```
PATTERN: (ssn|social_security|credit_card|creditCard|cardNumber|card_number|tax_id|taxId)\s*[=:][^}]*(res\.|return|response)
LANGUAGES: javascript, typescript, python
```

Internal IDs exposed:
```
PATTERN: (_id|internal_id|internalId|system_id|db_id)\s*[=:][^}]*(res\.|return|response)
LANGUAGES: javascript, typescript, python
```

### Debug Information Leakage

CATEGORY: api-data-exposure
SEVERITY: medium
CONFIDENCE: 85
CWE: CWE-209
OWASP: API3:2023

Stack trace in error response:
```
PATTERN: (res\.(json|send)|return)\s*\([^)]*\b(stack|stackTrace|trace)\b
LANGUAGES: javascript, typescript
```

SQL error details exposed:
```
PATTERN: catch\s*\([^)]*\)\s*\{[^}]*(res\.(json|send)|return)[^}]*\berr(or)?\.(message|stack)
LANGUAGES: javascript, typescript
```

### Verbose Error Messages

CATEGORY: api-data-exposure
SEVERITY: medium
CONFIDENCE: 80
CWE: CWE-209
OWASP: API3:2023

Database error details in response:
```
PATTERN: (SequelizeError|MongoError|PrismaClientKnownRequestError|DatabaseError)[^}]*(res\.|return)
LANGUAGES: javascript, typescript
```

### API Enumeration

CATEGORY: api-data-exposure
SEVERITY: medium
CONFIDENCE: 75
CWE: CWE-204
OWASP: API3:2023

Different error messages for existing vs non-existing:
```
PATTERN: (user|email|account)\s+(not found|doesn't exist|does not exist|invalid)
LANGUAGES: javascript, typescript, python
```

### GraphQL Introspection Enabled

CATEGORY: api-data-exposure
SEVERITY: medium
CONFIDENCE: 90
CWE: CWE-200
OWASP: API3:2023

GraphQL introspection not disabled in production:
```
PATTERN: introspection\s*[=:]\s*true
LANGUAGES: javascript, typescript
```

Missing introspection disable:
```
PATTERN: (ApolloServer|GraphQLServer|graphqlHTTP)\s*\(\s*\{(?!.*introspection\s*[=:]\s*false)
LANGUAGES: javascript, typescript
```

### Sensitive Data in Logs

CATEGORY: api-data-exposure
SEVERITY: high
CONFIDENCE: 85
CWE: CWE-532
OWASP: API3:2023

Logging passwords or tokens:
```
PATTERN: (log|logger|console)\.(info|debug|warn|error)\s*\([^)]*\b(password|token|secret|apiKey|api_key|authorization)\b
LANGUAGES: javascript, typescript, python
```

### Mass Assignment Exposure

CATEGORY: api-data-exposure
SEVERITY: high
CONFIDENCE: 85
CWE: CWE-915
OWASP: API3:2023

Direct request body to model:
```
PATTERN: (create|update|save)\s*\(\s*req\.body\s*\)
LANGUAGES: javascript, typescript
```

Spread operator with request body:
```
PATTERN: \{\s*\.\.\.req\.body\s*\}
LANGUAGES: javascript, typescript
```

### Internal Endpoint Exposure

CATEGORY: api-data-exposure
SEVERITY: high
CONFIDENCE: 80
CWE: CWE-200
OWASP: API6:2023

Internal/admin endpoints without protection:
```
PATTERN: (\/internal|\/admin|\/debug|\/metrics|\/health|\/status).*app\.(get|post|use)(?!.*auth)
LANGUAGES: javascript, typescript
```

Actuator endpoints exposed:
```
PATTERN: (\/actuator|\/env|\/heapdump|\/threaddump)
LANGUAGES: javascript, typescript, python, java
```

### Response Header Leakage

CATEGORY: api-data-exposure
SEVERITY: low
CONFIDENCE: 85
CWE: CWE-200
OWASP: API3:2023

Server version in headers:
```
PATTERN: (X-Powered-By|Server)\s*[=:]\s*['"][^'"]+['"]
LANGUAGES: javascript, typescript
```

## References

- [OWASP API3:2023 - Broken Object Property Level Authorization](https://owasp.org/API-Security/editions/2023/en/0xa3-broken-object-property-level-authorization/)
- [CWE-200: Exposure of Sensitive Information](https://cwe.mitre.org/data/definitions/200.html)
- [CWE-209: Error Message Information Leak](https://cwe.mitre.org/data/definitions/209.html)
