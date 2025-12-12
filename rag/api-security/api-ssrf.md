# API Server-Side Request Forgery (SSRF)

Detection patterns for SSRF vulnerabilities in API endpoints.

## OWASP API Security Top 10

- **API7:2023** - Server Side Request Forgery

## Patterns

### Direct URL from Request

CATEGORY: api-ssrf
SEVERITY: critical
CONFIDENCE: 95
CWE: CWE-918
OWASP: API7:2023

Fetch/axios with user-provided URL:
```
PATTERN: (fetch|axios|got|request|http\.get|https\.get)\s*\(\s*req\.(body|params|query)\.(url|uri|href|link|target|redirect)
LANGUAGES: javascript, typescript
```

Python requests with user URL:
```
PATTERN: requests\.(get|post|put|delete|head|patch)\s*\(\s*request\.(args|form|json)\.(url|uri|href|link|target)
LANGUAGES: python
```

### URL Construction from User Input

CATEGORY: api-ssrf
SEVERITY: critical
CONFIDENCE: 90
CWE: CWE-918
OWASP: API7:2023

URL concatenation with user input:
```
PATTERN: (http|https):\/\/.*(\+|`\$\{).*req\.(body|params|query)
LANGUAGES: javascript, typescript
```

Template literal URL:
```
PATTERN: `(http|https):\/\/\$\{.*req\.(body|params|query)
LANGUAGES: javascript, typescript
```

Python f-string URL:
```
PATTERN: f['"](http|https):\/\/.*\{.*request\.(args|form|json)
LANGUAGES: python
```

### Image/File URL Fetch

CATEGORY: api-ssrf
SEVERITY: high
CONFIDENCE: 90
CWE: CWE-918
OWASP: API7:2023

Image download from user URL:
```
PATTERN: (imageUrl|image_url|avatarUrl|avatar_url|fileUrl|file_url|pictureUrl)\s*=\s*req\.(body|params|query)
LANGUAGES: javascript, typescript, python
```

Fetch for file download:
```
PATTERN: (download|fetch|get).*\(\s*req\.(body|params|query)\.(url|path|file|image)
LANGUAGES: javascript, typescript
```

### Webhook URL

CATEGORY: api-ssrf
SEVERITY: high
CONFIDENCE: 85
CWE: CWE-918
OWASP: API7:2023

Webhook endpoint from user input:
```
PATTERN: (webhook|callback|notify|endpoint).*=\s*req\.(body|params|query)
LANGUAGES: javascript, typescript, python
```

### XML External Entity (XXE)

CATEGORY: api-ssrf
SEVERITY: critical
CONFIDENCE: 90
CWE: CWE-611
OWASP: API7:2023

XML parsing without disabling external entities:
```
PATTERN: (DOMParser|xml2js|xmldom|libxmljs|etree\.parse|lxml\.etree)(?!.*resolveExternals\s*[=:]\s*false|noent\s*[=:]\s*false)
LANGUAGES: javascript, typescript, python
```

Dangerous XML parser options:
```
PATTERN: (parseXML|parse).*\{[^}]*(resolveExternals|expandEntities|loadExternalDTD)\s*[=:]\s*true
LANGUAGES: javascript, typescript
```

### PDF Generation SSRF

CATEGORY: api-ssrf
SEVERITY: high
CONFIDENCE: 85
CWE: CWE-918
OWASP: API7:2023

PDF from URL with user input:
```
PATTERN: (puppeteer|playwright|pdf|wkhtmltopdf|phantom).*(goto|navigate|url|create)\s*\([^)]*req\.(body|params|query)
LANGUAGES: javascript, typescript
```

### HTML Rendering SSRF

CATEGORY: api-ssrf
SEVERITY: high
CONFIDENCE: 85
CWE: CWE-918
OWASP: API7:2023

Browser automation with user URL:
```
PATTERN: (page\.goto|page\.navigate|browser\.newPage).*req\.(body|params|query)
LANGUAGES: javascript, typescript
```

### Cloud Metadata Access

CATEGORY: api-ssrf
SEVERITY: critical
CONFIDENCE: 95
CWE: CWE-918
OWASP: API7:2023

AWS metadata endpoint access pattern:
```
PATTERN: 169\.254\.169\.254
LANGUAGES: javascript, typescript, python, java, go
```

GCP metadata endpoint:
```
PATTERN: metadata\.google\.internal
LANGUAGES: javascript, typescript, python, java, go
```

Azure metadata endpoint:
```
PATTERN: 169\.254\.169\.254.*metadata
LANGUAGES: javascript, typescript, python, java, go
```

### DNS Rebinding Vulnerability

CATEGORY: api-ssrf
SEVERITY: high
CONFIDENCE: 80
CWE: CWE-918
OWASP: API7:2023

URL validation then fetch (TOCTOU):
```
PATTERN: (isValid|validate|check).*url.*\n.*fetch\s*\(\s*\1
LANGUAGES: javascript, typescript
```

### Internal Network Access

CATEGORY: api-ssrf
SEVERITY: high
CONFIDENCE: 85
CWE: CWE-918
OWASP: API7:2023

Localhost/internal IP patterns:
```
PATTERN: (127\.|10\.|172\.(1[6-9]|2[0-9]|3[01])\.|192\.168\.|localhost|0\.0\.0\.0)
LANGUAGES: javascript, typescript, python
```

### Redirect Following

CATEGORY: api-ssrf
SEVERITY: medium
CONFIDENCE: 80
CWE: CWE-918
OWASP: API7:2023

Following redirects without validation:
```
PATTERN: (followRedirects?|maxRedirects|redirect)\s*[=:]\s*(true|[1-9])
LANGUAGES: javascript, typescript
```

### GraphQL SSRF

CATEGORY: api-ssrf
SEVERITY: high
CONFIDENCE: 85
CWE: CWE-918
OWASP: API7:2023

GraphQL with URL fields:
```
PATTERN: (imageUrl|profileUrl|webhookUrl|callbackUrl)\s*:\s*String
LANGUAGES: graphql, javascript, typescript
```

### Import from URL

CATEGORY: api-ssrf
SEVERITY: critical
CONFIDENCE: 90
CWE: CWE-918
OWASP: API7:2023

Dynamic import from user URL:
```
PATTERN: import\s*\(\s*req\.(body|params|query)
LANGUAGES: javascript, typescript
```

Require from user input:
```
PATTERN: require\s*\(\s*req\.(body|params|query)
LANGUAGES: javascript, typescript
```

### Protocol Handler SSRF

CATEGORY: api-ssrf
SEVERITY: high
CONFIDENCE: 85
CWE: CWE-918
OWASP: API7:2023

File protocol in URL:
```
PATTERN: file:\/\/
LANGUAGES: javascript, typescript, python
```

Gopher protocol:
```
PATTERN: gopher:\/\/
LANGUAGES: javascript, typescript, python
```

Dict protocol:
```
PATTERN: dict:\/\/
LANGUAGES: javascript, typescript, python
```

## Remediation Examples

### Safe Patterns

URL allowlist validation:
```javascript
// SAFE: Validate against allowlist
const ALLOWED_DOMAINS = ['api.example.com', 'cdn.example.com'];
const url = new URL(req.body.url);
if (!ALLOWED_DOMAINS.includes(url.hostname)) {
  throw new Error('Domain not allowed');
}
```

Disable redirects:
```javascript
// SAFE: Don't follow redirects
const response = await fetch(url, { redirect: 'error' });
```

## References

- [OWASP API7:2023 - Server Side Request Forgery](https://owasp.org/API-Security/editions/2023/en/0xa7-server-side-request-forgery/)
- [CWE-918: Server-Side Request Forgery](https://cwe.mitre.org/data/definitions/918.html)
- [PortSwigger SSRF](https://portswigger.net/web-security/ssrf)
