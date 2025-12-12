# API Rate Limiting & Resource Exhaustion

Detection patterns for missing rate limiting and resource exhaustion vulnerabilities.

## OWASP API Security Top 10

- **API4:2023** - Unrestricted Resource Consumption

## Patterns

### Missing Rate Limiting

CATEGORY: api-rate-limiting
SEVERITY: medium
CONFIDENCE: 75
CWE: CWE-770
OWASP: API4:2023

Express app without rate limiter:
```
PATTERN: const\s+app\s*=\s*express\s*\(\)(?![\s\S]*rate-limit|rateLimit|rateLimiter)
LANGUAGES: javascript, typescript
```

Login endpoint without rate limiting:
```
PATTERN: (\/login|\/signin|\/auth|\/authenticate).*app\.(post|get)(?!.*rateLimit|rateLimiter)
LANGUAGES: javascript, typescript
```

Password reset without rate limiting:
```
PATTERN: (\/reset-password|\/forgot-password|\/password-reset).*app\.(post|get)(?!.*rateLimit)
LANGUAGES: javascript, typescript
```

### Missing Request Size Limits

CATEGORY: api-rate-limiting
SEVERITY: medium
CONFIDENCE: 80
CWE: CWE-770
OWASP: API4:2023

Body parser without size limit:
```
PATTERN: (bodyParser|express)\.(json|urlencoded)\s*\(\s*\)
LANGUAGES: javascript, typescript
```

Multer without file size limit:
```
PATTERN: multer\s*\(\s*\{(?!.*limits)
LANGUAGES: javascript, typescript
```

### Missing Pagination

CATEGORY: api-rate-limiting
SEVERITY: medium
CONFIDENCE: 70
CWE: CWE-770
OWASP: API4:2023

Find all without limit:
```
PATTERN: \.(find|findAll|findMany)\s*\(\s*\{(?!.*limit|take|first)
LANGUAGES: javascript, typescript
```

Database query without pagination:
```
PATTERN: (SELECT|select)(?!.*LIMIT|limit).*FROM.*res\.(json|send)
LANGUAGES: javascript, typescript, python
```

### GraphQL Complexity Limits

CATEGORY: api-rate-limiting
SEVERITY: medium
CONFIDENCE: 80
CWE: CWE-770
OWASP: API4:2023

GraphQL without depth limiting:
```
PATTERN: (ApolloServer|GraphQLServer)\s*\(\s*\{(?!.*depthLimit|queryComplexity|maxDepth)
LANGUAGES: javascript, typescript
```

GraphQL without query complexity:
```
PATTERN: (ApolloServer|GraphQLServer)\s*\(\s*\{(?!.*complexity|costAnalysis)
LANGUAGES: javascript, typescript
```

### Unrestricted File Upload

CATEGORY: api-rate-limiting
SEVERITY: high
CONFIDENCE: 85
CWE: CWE-400
OWASP: API4:2023

File upload without size validation:
```
PATTERN: (upload|multer|formidable|busboy)(?!.*fileSize|maxFileSize|limits)
LANGUAGES: javascript, typescript
```

No file count limit:
```
PATTERN: (\.array|\.fields)\s*\([^)]*\)(?!.*maxCount|limits)
LANGUAGES: javascript, typescript
```

### Regex DoS (ReDoS)

CATEGORY: api-rate-limiting
SEVERITY: high
CONFIDENCE: 85
CWE: CWE-1333
OWASP: API4:2023

Dangerous regex patterns:
```
PATTERN: new\s+RegExp\s*\(\s*req\.(body|params|query)
LANGUAGES: javascript, typescript
```

Nested quantifiers (potential ReDoS):
```
PATTERN: \/\([^)]*[\*\+]\)[^\/]*[\*\+]\/
LANGUAGES: javascript, typescript
```

### Missing Timeout

CATEGORY: api-rate-limiting
SEVERITY: medium
CONFIDENCE: 75
CWE: CWE-400
OWASP: API4:2023

HTTP request without timeout:
```
PATTERN: (axios|fetch|request|got)\s*\(\s*[^)]*\)(?!.*timeout)
LANGUAGES: javascript, typescript
```

Database query without timeout:
```
PATTERN: (query|execute)\s*\(\s*[^)]*\)(?!.*timeout|maxTimeMS)
LANGUAGES: javascript, typescript, python
```

### Memory Exhaustion

CATEGORY: api-rate-limiting
SEVERITY: high
CONFIDENCE: 80
CWE: CWE-400
OWASP: API4:2023

Unbounded array growth:
```
PATTERN: \[\s*\]\.push\s*\([^)]*req\.(body|params|query)
LANGUAGES: javascript, typescript
```

Reading entire file into memory:
```
PATTERN: (readFileSync|readFile)\s*\([^)]*req\.(body|params|query)
LANGUAGES: javascript, typescript
```

### Connection Pool Exhaustion

CATEGORY: api-rate-limiting
SEVERITY: medium
CONFIDENCE: 75
CWE: CWE-400
OWASP: API4:2023

Database connection without pooling:
```
PATTERN: (createConnection|connect)\s*\([^)]*\)(?!.*pool|poolSize|connectionLimit)
LANGUAGES: javascript, typescript
```

### Batch Operation Limits

CATEGORY: api-rate-limiting
SEVERITY: medium
CONFIDENCE: 70
CWE: CWE-770
OWASP: API4:2023

Bulk operations without limits:
```
PATTERN: (bulkCreate|insertMany|bulkWrite)\s*\(\s*req\.body(?!.*limit)
LANGUAGES: javascript, typescript
```

### CPU-Intensive Operations

CATEGORY: api-rate-limiting
SEVERITY: medium
CONFIDENCE: 75
CWE: CWE-400
OWASP: API4:2023

Synchronous crypto in request handler:
```
PATTERN: (pbkdf2Sync|scryptSync|hashSync)\s*\([^)]*req\.(body|params|query)
LANGUAGES: javascript, typescript
```

JSON parsing of large input:
```
PATTERN: JSON\.parse\s*\(\s*req\.(body|params|query)
LANGUAGES: javascript, typescript
```

## References

- [OWASP API4:2023 - Unrestricted Resource Consumption](https://owasp.org/API-Security/editions/2023/en/0xa4-unrestricted-resource-consumption/)
- [CWE-770: Allocation of Resources Without Limits](https://cwe.mitre.org/data/definitions/770.html)
- [CWE-400: Uncontrolled Resource Consumption](https://cwe.mitre.org/data/definitions/400.html)
