# RAG-to-Semgrep Pipeline

## Overview

Zero uses a **Retrieval-Augmented Generation (RAG)** knowledge base as the source of truth for detection patterns. These patterns are stored as human-readable markdown files and converted to Semgrep YAML rules for code scanning.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        RAG Pattern Files (Markdown)                         │
│                                                                              │
│  rag/technology-identification/                                             │
│  ├── cloud-providers/aws/patterns.md                                        │
│  ├── business-tools/stripe/patterns.md                                      │
│  ├── ai-ml/apis/openai/patterns.md                                          │
│  └── ... (106 technology patterns)                                          │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         rag-to-semgrep.py Converter                         │
│                                                                              │
│  • Parses markdown structure                                                │
│  • Extracts patterns by type (secrets, imports, packages)                  │
│  • Converts regex to Semgrep format                                         │
│  • Adds metadata (severity, CWE, technology)                               │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         Semgrep YAML Rules                                  │
│                                                                              │
│  utils/scanners/semgrep/rules/                                              │
│  ├── secrets.yaml        (242 rules)                                        │
│  ├── tech-discovery.yaml (import/package detection)                        │
│  ├── tech-debt.yaml      (TODO, FIXME markers)                             │
│  └── crypto-security.yaml (cryptography-specific)                          │
└─────────────────────────────────────────────────────────────────────────────┘
```

## RAG Pattern Format

Each technology pattern is defined in a markdown file with standardized sections:

```markdown
# Technology Name

**Category**: category/subcategory
**Description**: What this technology is

---

## Package Detection

### NPM
- `package-name`
- `@scope/package-name`

### PYPI
- `package-name`

### Go
- `github.com/org/package`

---

## Import Detection

### Python
**Pattern**: `^import openai`
- Official Python client
- Example: `import openai`

**Pattern**: `^from openai import`
- Module-level import
- Example: `from openai import OpenAI`

### Javascript
**Pattern**: `from ['"]openai['"]`
- ES6 import pattern
- Example: `import OpenAI from 'openai'`

**Pattern**: `require\(['"]openai['"]\)`
- CommonJS require
- Example: `const OpenAI = require('openai')`

---

## Secrets Detection

#### API Key
**Pattern**: `sk-[A-Za-z0-9]{48}`
**Severity**: critical
**Description**: OpenAI API key - grants access to OpenAI services

#### Organization ID
**Pattern**: `org-[A-Za-z0-9]{24}`
**Severity**: medium
**Description**: OpenAI organization identifier

---

## Environment Variables

- `OPENAI_API_KEY`
- `OPENAI_ORG_ID`

---

## Detection Confidence

**Import Detection**: 95%
**Package Detection**: 99%
**Secret Detection**: 90%
```

## Conversion Process

The `rag-to-semgrep.py` script performs the following steps:

### 1. Parse Markdown Structure

```python
class PatternParser:
    def __init__(self, file_path: str):
        self.content = Path(file_path).read_text()
        self.data = {
            'name': '',
            'category': '',
            'description': '',
            'packages': {'npm': [], 'pypi': [], 'go': []},
            'imports': {'python': [], 'javascript': [], 'go': []},
            'secrets': [],
            'env_vars': [],
            'confidence': {}
        }
        self._parse()
```

### 2. Extract Pattern Types

**Secrets Detection:**
```python
def _parse_secrets(self):
    # Find each secret pattern block
    pattern_blocks = re.findall(
        r'####\s*(.+?)\n.*?\*\*Pattern\*\*:\s*`([^`]+)`.*?\*\*Severity\*\*:\s*(\w+)',
        section, re.DOTALL
    )

    for name, pattern, severity in pattern_blocks:
        self.data['secrets'].append({
            'name': name.strip(),
            'pattern': pattern.strip(),
            'severity': severity.strip().upper()
        })
```

**Import Detection:**
```python
def _parse_imports(self):
    # Extract Python patterns
    patterns = re.findall(r'\*\*Pattern\*\*:\s*`([^`]+)`', section)
    self.data['imports']['python'] = patterns
```

### 3. Convert Regex to Semgrep

Regex patterns are converted to Semgrep's pattern syntax:

```python
def regex_to_semgrep(regex_pattern: str, language: str) -> str:
    # Python: `^import openai` -> `import openai`
    if pattern.startswith('^import '):
        return f'import {module}'

    # Python: `^from openai import` -> `from openai import $X`
    if pattern.startswith('^from ') and 'import' in pattern:
        return f'from {module} import $X'

    # JavaScript: `from ['"]openai['"]` -> `import $X from "openai"`
    if "from ['\"]" in pattern:
        return f'import $X from "{module}"'

    # Constructor: `new OpenAI\(` -> `new OpenAI(...)`
    if pattern.startswith('new ') and pattern.endswith('\\('):
        return f'new {class_name}(...)'
```

### 4. Generate Semgrep Rules

**Secrets Rule Format:**
```yaml
- id: zero.cloud-providers.aws.secret.access-key
  message: "Potential AWS Access Key exposed"
  severity: ERROR
  languages: [generic]
  metadata:
    technology: AWS
    category: secrets
    secret_type: Access Key
    confidence: 95
  pattern-regex: "AKIA[0-9A-Z]{16}"
```

**Import Detection Rule Format:**
```yaml
- id: zero.ai-ml.apis.openai.import.python
  message: "OpenAI library import detected"
  severity: INFO
  languages: [python]
  metadata:
    technology: OpenAI
    category: ai-ml/apis
    detection_type: import
    confidence: 95
  pattern-either:
    - pattern: import openai
    - pattern: from openai import $X
```

### 5. Write Output Files

Rules are grouped by category:
- `secrets.yaml` - All secret detection patterns
- `tech-discovery.yaml` - Import and package detection
- `tech-debt.yaml` - TODO/FIXME markers
- `crypto-security.yaml` - Cryptography-specific rules

## Running the Converter

```bash
# Generate all rules from RAG patterns
python3 utils/scanners/semgrep/rag-to-semgrep.py \
    rag/technology-identification \
    utils/scanners/semgrep/rules

# Output:
# Found 106 pattern files
#   Converted: aws -> 15 rules
#   Converted: stripe -> 8 rules
#   Converted: openai -> 6 rules
#   ...
# ================================================
# Total: 242 Semgrep rules generated
#   - tech-discovery: 89 rules
#   - secrets: 153 rules
```

## Technology Coverage

The RAG patterns cover 106 technologies across these categories:

| Category | Count | Examples |
|----------|-------|----------|
| Cloud Providers | 12 | AWS, Azure, GCP, DigitalOcean |
| AI/ML APIs | 8 | OpenAI, Anthropic, Cohere, Hugging Face |
| Payment Processing | 6 | Stripe, PayPal, Square, Braintree |
| Communication | 8 | Twilio, SendGrid, Mailgun, Slack |
| Databases | 10 | PostgreSQL, MongoDB, Redis, Supabase |
| Authentication | 6 | Auth0, Okta, Firebase Auth |
| Developer Tools | 15 | GitHub, GitLab, CircleCI, Docker |
| Security | 8 | HashiCorp Vault, AWS Secrets Manager |
| Analytics | 7 | Segment, Mixpanel, Amplitude |
| Other | 26 | Various APIs and services |

## Adding New Patterns

### 1. Create RAG Pattern File

```bash
mkdir -p rag/technology-identification/category/technology
```

Create `patterns.md`:

```markdown
# New Technology

**Category**: category/technology
**Description**: What this technology does

## Secrets Detection

#### API Key
**Pattern**: `NEWTECH_[A-Za-z0-9]{32}`
**Severity**: high
**Description**: New Technology API key

## Import Detection

### Python
**Pattern**: `^import newtech`
- Official Python client

### Javascript
**Pattern**: `from ['"]newtech['"]`
- ES6 import
```

### 2. Regenerate Rules

```bash
python3 utils/scanners/semgrep/rag-to-semgrep.py \
    rag/technology-identification \
    utils/scanners/semgrep/rules
```

### 3. Verify Rules

```bash
# Test the new rules
semgrep --config utils/scanners/semgrep/rules/secrets.yaml \
    --json /path/to/test/repo

# Check rule count
grep -c "^- id:" utils/scanners/semgrep/rules/secrets.yaml
```

## Severity Mapping

RAG severity levels map to Semgrep severity:

| RAG Severity | Semgrep Severity | Use Case |
|--------------|------------------|----------|
| critical | ERROR | Active credentials, private keys |
| high | WARNING | API keys, service tokens |
| medium | WARNING | Organization IDs, webhook URLs |
| low | INFO | Informational patterns |

## Quality Assurance

### Pattern Validation

The converter performs validation:

1. **Regex Syntax** - Validates regex patterns compile correctly
2. **Semgrep Syntax** - Ensures generated patterns are valid Semgrep
3. **Deduplication** - Removes duplicate patterns
4. **Metadata Completeness** - Verifies required fields present

### Testing New Patterns

```bash
# Test individual rule file
semgrep --config utils/scanners/semgrep/rules/secrets.yaml \
    --validate

# Test against sample repository
./utils/scanners/code-secrets/code-secrets.sh /path/to/test/repo
```

## See Also

- [Scanner Architecture](scanners.md) - How scanners use these rules
- [Creating Scanners](../scanners/creating-scanners.md) - Building custom scanners
- [RAG Technology Patterns](../../rag/technology-identification/) - Pattern source files
