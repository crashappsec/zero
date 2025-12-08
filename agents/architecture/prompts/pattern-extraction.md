# Pattern Extraction Prompt

**Purpose**: Extract technology identification patterns from official documentation, SDKs, and API references.

## Prompt

```
You are a technology pattern extraction expert. Your task is to analyze documentation for {TECHNOLOGY_NAME} and extract identification patterns that can be used to detect this technology in source code repositories.

## Technology Information

**Technology**: {TECHNOLOGY_NAME}
**Category**: {CATEGORY} (e.g., business-tools/payment, developer-tools/infrastructure)
**Version**: {VERSION} (if applicable)
**Documentation Source**: {DOC_URL}

## Documentation Content

{DOCUMENTATION_TEXT}

## Extraction Tasks

### 1. API Endpoint Patterns

Extract all API endpoint patterns:
- Base URLs (production and test/sandbox)
- API version paths
- Common endpoint patterns
- Regional endpoints (if applicable)

**Format**:
```markdown
## API Endpoints

### Production
- `https://api.example.com/v1/*` - Confidence: 85%
- `https://api.example.com/v2/*` - Confidence: 85%

### Test/Sandbox
- `https://sandbox.example.com/v1/*` - Confidence: 70%

### Regional
- `https://api.us-east-1.example.com/*` - Confidence: 80%
- `https://api.eu-west-1.example.com/*` - Confidence: 80%
```

### 2. SDK Import Patterns

Extract import/require patterns for all supported languages:

**Format**:
```markdown
## SDK Patterns

### JavaScript/Node.js
```javascript
// High confidence patterns
const example = require('example-sdk');
import Example from 'example-sdk';
const client = new Example({ apiKey: '...' });
```
Confidence:
- With instantiation: 95%
- Import only: 75%

### Python
```python
import example_sdk
from example_sdk import Client
client = Client(api_key='...')
```
Confidence: 85%

### Go
```go
import "github.com/example/example-go"
```
Confidence: 90%

### Ruby
```ruby
require 'example'
Example.api_key = '...'
```
Confidence: 85%
```

### 3. Configuration File Patterns

Extract configuration file patterns:

**Format**:
```markdown
## Configuration Files

### File Names
- `example.config.js` - Confidence: 90%
- `.examplerc` - Confidence: 85%
- `example.yml` - Confidence: 85%

### Configuration Content Patterns
```yaml
# example.yml
example:
  api_key: ${EXAMPLE_API_KEY}
  environment: production
```
Confidence: 80%
```

### 4. Environment Variable Patterns

Extract environment variable naming conventions:

**Format**:
```markdown
## Environment Variables

- `EXAMPLE_API_KEY` - Confidence: 70%
- `EXAMPLE_SECRET_KEY` - Confidence: 70%
- `EXAMPLE_ENVIRONMENT` - Confidence: 65%
- `EXAMPLE_WEBHOOK_SECRET` - Confidence: 65%
```

### 5. Package Dependencies

Extract package manager identifiers:

**Format**:
```markdown
## Package Dependencies

### npm
- `example-sdk` - Confidence: 95%
- `@example/client` - Confidence: 95%

### Python (PyPI)
- `example-sdk` - Confidence: 95%
- `example_client` - Confidence: 95%

### Go
- `github.com/example/example-go` - Confidence: 95%

### Ruby (RubyGems)
- `example` - Confidence: 95%

### Java (Maven)
- `com.example:example-sdk` - Confidence: 95%
```

### 6. Webhook & Callback Patterns

Extract webhook signature verification and callback patterns:

**Format**:
```markdown
## Webhook Patterns

```javascript
// Webhook signature verification
const signature = example.webhooks.verify(body, headers, secret);
```
Confidence: 90% (strong indicator of integration)

```python
# Python webhook handler
is_valid = example.Webhook.verify(payload, signature, secret)
```
Confidence: 90%
```

### 7. Version Detection Methods

Extract methods to detect the technology version:

**Format**:
```markdown
## Version Detection

### From Package Manifest
- package.json: `"example-sdk": "^3.5.0"` → Version 3.5.x
- requirements.txt: `example-sdk==3.5.0` → Version 3.5.0

### From SDK
```javascript
console.log(example.VERSION);  // "3.5.0"
```

### From API Headers
```http
Example-API-Version: 2023-11-01
User-Agent: example-sdk-node/3.5.0
```

### From Binary
```bash
example --version  # example-cli 3.5.0
```
```

### 8. Authentication Patterns

Extract authentication mechanisms:

**Format**:
```markdown
## Authentication Patterns

### API Key (Header)
```http
Authorization: Bearer sk_live_xxxxx
X-API-Key: xxxxx
```
Confidence: 85%

### API Key (Query Parameter)
```
https://api.example.com/v1/resource?api_key=xxxxx
```
Confidence: 75%

### OAuth 2.0
```javascript
const token = await example.oauth.authorize({
  client_id: '...',
  client_secret: '...'
});
```
Confidence: 90%
```

### 9. Detection Rules

Create detection rules with confidence levels:

**Format**:
```markdown
## Detection Rules

### Definitive (95-100%)
- Package dependency declared in manifest
- SDK import with instantiation
- Webhook signature verification pattern

### High Confidence (80-94%)
- API endpoint with authentication
- Environment variables + import statement
- Configuration file with API keys

### Medium Confidence (60-79%)
- Import without usage verification
- Environment variables only
- API endpoint without authentication

### Low Confidence (0-59%)
- Documentation mentions
- Comments referencing technology
```

## Output Format

Generate the following markdown files:

1. **api-patterns.md** - API endpoint patterns
2. **import-patterns.md** - SDK import patterns (all languages)
3. **config-patterns.md** - Configuration file patterns
4. **env-variables.md** - Environment variable patterns
5. **versions.md** - Version detection and history

## Quality Criteria

- **Accuracy**: Patterns must be verifiable in official documentation
- **Completeness**: Cover all major programming languages and use cases
- **Confidence**: Assign realistic confidence scores based on signal strength
- **Versioning**: Note when patterns changed across versions
- **Examples**: Include real code examples from documentation

Generate comprehensive, accurate pattern files that can be used for automated technology detection.
```

## Usage

```bash
# Extract patterns from Stripe documentation
./rag-updater/extract-patterns.sh \
  --technology stripe \
  --category business-tools/payment \
  --docs-url https://stripe.com/docs/api \
  --output rag/technology-identification/business-tools/payment/stripe/

# Extract patterns from Terraform documentation
./rag-updater/extract-patterns.sh \
  --technology terraform \
  --category developer-tools/infrastructure \
  --docs-url https://developer.hashicorp.com/terraform/docs \
  --output rag/technology-identification/developer-tools/infrastructure/terraform/
```
