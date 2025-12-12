# API Versioning

Detection patterns for API versioning practices and issues.

## Overview

API versioning is essential for maintaining backward compatibility while evolving APIs.
This scanner identifies versioning patterns and potential issues.

**Note:** This scanner is informational and NOT included in the security profile.

## Patterns

### URL Path Versioning

CATEGORY: api-versioning
SEVERITY: info
CONFIDENCE: 90
CWE: none
OWASP: none

Version in URL path (common pattern):
```
PATTERN: \/v[0-9]+\/
LANGUAGES: javascript, typescript, python, java, go
```

Major.minor versioning:
```
PATTERN: \/v[0-9]+\.[0-9]+\/
LANGUAGES: javascript, typescript, python, java, go
```

### Header Versioning

CATEGORY: api-versioning
SEVERITY: info
CONFIDENCE: 85
CWE: none
OWASP: none

Custom version header:
```
PATTERN: (X-API-Version|Accept-Version|Api-Version)\s*[=:]
LANGUAGES: javascript, typescript, python
```

Accept header versioning:
```
PATTERN: Accept.*application\/vnd\.[^+]+\+json;\s*version=
LANGUAGES: javascript, typescript, python
```

### Query Parameter Versioning

CATEGORY: api-versioning
SEVERITY: info
CONFIDENCE: 85
CWE: none
OWASP: none

Version in query string:
```
PATTERN: [\?&]version=[0-9]
LANGUAGES: javascript, typescript, python
```

API version query param:
```
PATTERN: [\?&]api[-_]?version=[0-9]
LANGUAGES: javascript, typescript, python
```

### Deprecated API Endpoints

CATEGORY: api-versioning
SEVERITY: low
CONFIDENCE: 80
CWE: none
OWASP: none

Deprecated annotation/comment:
```
PATTERN: (@deprecated|@Deprecated|# deprecated|// deprecated|DEPRECATED)
LANGUAGES: javascript, typescript, python, java
```

Sunset header:
```
PATTERN: Sunset\s*[=:]
LANGUAGES: javascript, typescript, python
```

Deprecation header:
```
PATTERN: Deprecation\s*[=:]
LANGUAGES: javascript, typescript, python
```

### Version Mismatch

CATEGORY: api-versioning
SEVERITY: low
CONFIDENCE: 70
CWE: none
OWASP: none

Multiple version definitions:
```
PATTERN: (\/v1\/.*\/v2\/|\/v2\/.*\/v1\/)
LANGUAGES: javascript, typescript, python
```

### Missing Version

CATEGORY: api-versioning
SEVERITY: info
CONFIDENCE: 60
CWE: none
OWASP: none

API routes without version:
```
PATTERN: app\.(get|post|put|delete)\s*\(\s*['"]\/api\/(?!v[0-9])
LANGUAGES: javascript, typescript
```

### Legacy API Versions

CATEGORY: api-versioning
SEVERITY: info
CONFIDENCE: 75
CWE: none
OWASP: none

Very old API versions still active:
```
PATTERN: \/v[0-1]\/
LANGUAGES: javascript, typescript, python
```

### GraphQL Versioning

CATEGORY: api-versioning
SEVERITY: info
CONFIDENCE: 80
CWE: none
OWASP: none

GraphQL schema version:
```
PATTERN: (schemaVersion|schema_version|apiVersion)\s*[=:]\s*['"][0-9]
LANGUAGES: javascript, typescript, python
```

### OpenAPI Version Spec

CATEGORY: api-versioning
SEVERITY: info
CONFIDENCE: 90
CWE: none
OWASP: none

OpenAPI version field:
```
PATTERN: (openapi|swagger)\s*[=:]\s*['"][0-9]+\.[0-9]+
LANGUAGES: yaml, json
```

API info version:
```
PATTERN: info\s*:[\s\S]*?version\s*[=:]\s*['"][0-9]+\.[0-9]+
LANGUAGES: yaml, json
```

## Best Practices

### Recommended Patterns

1. **URL Path Versioning** (Most Common)
   - `/api/v1/users`
   - Clear, easy to understand
   - Works well with caching

2. **Header Versioning**
   - `Accept: application/vnd.myapi.v2+json`
   - Cleaner URLs
   - More RESTful

3. **Sunset Headers**
   - `Sunset: Sat, 31 Dec 2024 23:59:59 GMT`
   - Communicate deprecation timeline

### Anti-Patterns

1. **Query Parameter Versioning**
   - `?version=2`
   - Caching issues
   - Less RESTful

2. **No Versioning**
   - Breaking changes affect all clients
   - No migration path

## References

- [API Versioning Best Practices](https://www.postman.com/api-platform/api-versioning/)
- [REST API Versioning](https://restfulapi.net/versioning/)
- [Sunset HTTP Header](https://datatracker.ietf.org/doc/html/rfc8594)
