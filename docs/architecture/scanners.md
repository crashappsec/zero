# Scanner Architecture

## Overview

Zero's scanners follow a layered architecture:

1. **Semgrep Engine** - The core pattern matching engine for all code analysis
2. **RAG Patterns** - Source of truth for detection rules (markdown → YAML)
3. **Scanner Wrappers** - Shell scripts that add validation, scoring, and formatting

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            Scanner Wrapper (Bash)                            │
│                                                                              │
│  • Invokes Semgrep with appropriate rules                                   │
│  • Post-processes results (validation, false positive reduction)            │
│  • Calculates risk scores                                                   │
│  • Formats JSON output                                                      │
│  • Generates recommendations                                                │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Semgrep Engine                                  │
│                                                                              │
│  • AST-aware pattern matching (not just regex)                              │
│  • Multi-language support (Python, JS, Go, Java, Ruby, PHP, C, etc.)       │
│  • High performance on large codebases                                      │
│  • Consistent JSON output format                                            │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         RAG-Generated Semgrep Rules                          │
│                                                                              │
│  utils/scanners/semgrep/rules/                                              │
│  ├── secrets.yaml        (242+ rules from Technology detection patterns)          │
│  ├── tech-discovery.yaml (technology identification rules)                  │
│  ├── tech-debt.yaml      (TODO, FIXME, complexity markers)                  │
│  └── crypto-security.yaml (cryptography-specific rules)                     │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         RAG Pattern Definitions                              │
│                                                                              │
│  rag/technology-identification/  (Technology detection patterns)                  │
│  ├── cloud-providers/aws/patterns.md                                        │
│  ├── cloud-providers/azure/patterns.md                                      │
│  ├── cloud-providers/gcp/patterns.md                                        │
│  ├── business-tools/stripe/patterns.md                                      │
│  ├── ai-ml/apis/openai/patterns.md                                          │
│  └── ... (100+ more)                                                        │
│                                                                              │
│  rag/cryptography/  (crypto-specific patterns)                              │
│  ├── weak-ciphers/patterns.md                                               │
│  ├── hardcoded-keys/patterns.md                                             │
│  ├── insecure-random/patterns.md                                            │
│  └── tls-misconfig/patterns.md                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Why This Architecture?

### Semgrep as the Engine

We use Semgrep for ALL code scanning because:

| Advantage | Description |
|-----------|-------------|
| **AST-aware** | Understands code structure, not just text patterns |
| **Multi-language** | One engine for Python, JS, Go, Java, Ruby, PHP, C, etc. |
| **High performance** | Optimized for large codebases |
| **Consistent format** | Same JSON output regardless of language |
| **Rule ecosystem** | Access to community rules when needed |

### RAG as Source of Truth

Pattern definitions live in markdown because:

| Advantage | Description |
|-----------|-------------|
| **Human-readable** | Easy to review and maintain |
| **Version controlled** | Track changes over time |
| **Documentation included** | Patterns include context and examples |
| **Single source** | One place defines detection + documentation |

### Wrappers Add Intelligence

Scanner scripts wrap Semgrep because raw findings need:

| Feature | Description |
|---------|-------------|
| **Validation** | API checks to confirm secrets are real |
| **Entropy analysis** | Detect high-entropy strings |
| **False positive reduction** | Filter test files, examples, etc. |
| **Risk scoring** | Calculate severity based on context |
| **Recommendations** | Generate actionable remediation steps |

## Scanner Types

### Code Scanners (Semgrep-based)

These scanners use Semgrep with our RAG-generated rules:

| Scanner | Purpose | Primary Rules |
|---------|---------|---------------|
| `code-vulns` | Security vulnerabilities | p/security-audit, p/owasp-top-ten |
| `code-secrets` | API keys, credentials | secrets.yaml (242+ RAG rules) |
| `code-crypto (ciphers)` | Weak ciphers | crypto-security.yaml + p/security-audit |
| `code-crypto (keys)` | Hardcoded keys | secrets.yaml + p/secrets |
| `code-crypto (random)` | Insecure RNG | crypto-security.yaml |
| `code-crypto (tls)` | TLS misconfig | p/insecure-transport |
| `tech-discovery` | Tech stack detection | tech-discovery.yaml |
| `tech-debt` | TODOs, complexity | tech-debt.yaml |

### Package Scanners (External Tools)

These scanners use specialized tools:

| Scanner | Tool | Purpose |
|---------|------|---------|
| `package-sbom` | syft/cdxgen | SBOM generation |
| `package-vulns` | osv-scanner | CVE detection |
| `package-health` | npm/PyPI APIs | Dependency health |
| `package-malcontent` | malcontent | Supply chain malware |
| `package-provenance` | sigstore | SLSA verification |

### Infrastructure Scanners (External Tools)

| Scanner | Tool | Purpose |
|---------|------|---------|
| `iac-security` | checkov | Terraform, K8s, CloudFormation |
| `container-security` | trivy, hadolint | Dockerfile, images |

## Rule Priority

When multiple rule sources are available, scanners load them in priority order:

```bash
# Priority 1: RAG-generated custom rules (most comprehensive)
if [[ -f "$RULES_DIR/secrets.yaml" ]]; then
    config_args+=("--config" "$RULES_DIR/secrets.yaml")
fi

# Priority 2: Domain-specific custom rules
if [[ -f "$RULES_DIR/crypto-security.yaml" ]]; then
    config_args+=("--config" "$RULES_DIR/crypto-security.yaml")
fi

# Priority 3: Semgrep registry rules (supplement)
config_args+=("--config" "p/secrets")
```

## Scanner Wrapper Structure

All scanner wrappers follow this pattern:

```bash
#!/bin/bash

# 1. Setup
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RULES_DIR="$SCRIPT_DIR/../semgrep/rules"

# 2. Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --local-path) REPO_PATH="$2"; shift 2 ;;
        --output) OUTPUT_FILE="$2"; shift 2 ;;
        *) REPO_PATH="$1"; shift ;;
    esac
done

# 3. Load rules (priority order)
config_args=()
if [[ -f "$RULES_DIR/secrets.yaml" ]]; then
    config_args+=("--config" "$RULES_DIR/secrets.yaml")
fi
config_args+=("--config" "p/secrets")  # Supplement

# 4. Run Semgrep
raw_output=$(semgrep "${config_args[@]}" --json "$REPO_PATH")

# 5. Post-process (validation, filtering)
validated=$(validate_findings "$raw_output")

# 6. Build output JSON
output=$(build_output "$validated")

# 7. Write results
echo "$output"
```

## The code-secrets Scanner

The `code-secrets` scanner is a good example of the full architecture:

### What it does:

1. **Loads RAG rules** - 242+ patterns from 106 technology definitions
2. **Runs Semgrep** - Finds potential secrets in code
3. **Validates findings** - Checks entropy, format, context
4. **Reduces false positives** - Filters test files, examples
5. **Scores risk** - Calculates severity based on secret type
6. **Generates recommendations** - Actionable remediation steps

### Rule sources:

```
secrets.yaml (242 rules from RAG)
├── AWS credentials (access keys, secret keys)
├── Azure credentials (storage keys, client secrets)
├── GCP credentials (service account keys, API keys)
├── Stripe API keys (live and test)
├── Twilio credentials (auth tokens, API keys)
├── SendGrid API keys
├── OpenAI API keys
├── Anthropic API keys
├── Database connection strings
├── JWT secrets
├── Private keys (RSA, EC, DSA, PGP)
└── ... 100+ more technology-specific patterns
```

### Validation logic:

```python
# Pseudo-code for validation
def validate_secret(finding):
    # Check entropy (random-looking strings)
    if entropy(finding.value) < 3.5:
        return False  # Too low entropy, likely not a secret

    # Check format matches expected pattern
    if not matches_expected_format(finding):
        return False

    # Check if in test/example file
    if is_test_file(finding.file):
        finding.severity = "low"

    # API validation (if enabled)
    if api_validation_enabled:
        if not is_valid_credential(finding):
            return False

    return True
```

## RAG Pattern Format

Patterns are defined in markdown with this structure:

```markdown
# Technology Name

**Category**: category/subcategory
**Description**: What this technology is

---

## Package Detection

### NPM
- `package-name`

### PYPI
- `package-name`

---

## Import Detection

### Python
**Pattern**: `import pattern`
- Description
- Example: `import example`

### Javascript
**Pattern**: `require pattern`
- Description

---

## Secrets Detection

#### Secret Type Name
**Pattern**: `regex_pattern`
**Severity**: critical|high|medium|low
**Description**: What this secret is

---

## Environment Variables
- `VAR_NAME`
```

## Generating Rules

The `rag-to-semgrep.py` script converts RAG patterns to Semgrep rules:

```bash
# Generate all rules
python3 utils/scanners/semgrep/rag-to-semgrep.py \
    rag/technology-identification \
    utils/scanners/semgrep/rules

# Output:
# - secrets.yaml (242 rules)
# - tech-discovery.yaml
# - tech-debt.yaml
```

### Conversion process:

1. Parse markdown patterns.md files
2. Extract patterns by type (secrets, imports, packages)
3. Convert regex patterns to Semgrep format
4. Add metadata (severity, CWE, technology)
5. Write YAML rule files

## Adding New Patterns

To add detection for a new technology:

### 1. Create RAG pattern file

```bash
mkdir -p rag/technology-identification/category/technology
```

Create `patterns.md`:

```markdown
# New Technology

**Category**: category/technology
**Description**: What it does

## Secrets Detection

#### API Key
**Pattern**: `NEWTECH_[A-Za-z0-9]{32}`
**Severity**: high
**Description**: New Technology API key
```

### 2. Regenerate rules

```bash
python3 utils/scanners/semgrep/rag-to-semgrep.py \
    rag/technology-identification \
    utils/scanners/semgrep/rules
```

### 3. Test

```bash
./utils/scanners/code-secrets/code-secrets.sh /path/to/test/repo
```

## See Also

- [RAG Pipeline](rag-pipeline.md) - Detailed RAG-to-Semgrep conversion
- [Scanner Reference](../scanners/reference.md) - All available scanners
- [Creating Scanners](../scanners/creating-scanners.md) - How to add new scanners
