# RAG Test Fixtures

Test fixtures for validating RAG pattern detection accuracy.

## Directory Structure

```
testdata/rag/
├── known-secrets/       # Files that SHOULD trigger detection
│   ├── api_keys.py      # Various API key patterns
│   └── weak_crypto.py   # Weak cryptography patterns
├── false-positives/     # Files that should NOT trigger detection
│   └── not_secrets.py   # Patterns that look like secrets but aren't
└── tech-samples/        # Technology detection samples
    ├── react/           # React/TypeScript samples
    ├── python/          # Python/Flask samples
    └── go/              # Go/Gin samples
```

## Usage

### Testing Secret Detection

```go
// Load patterns from rag/devops/secrets/
// Run against testdata/rag/known-secrets/
// Verify all expected patterns are detected

// Run against testdata/rag/false-positives/
// Verify no false positives are triggered
```

### Testing Technology Detection

```go
// Load patterns from rag/technology-identification/
// Run against testdata/rag/tech-samples/
// Verify correct technologies are detected
```

## Adding New Test Cases

When adding new test fixtures:

1. Add files to appropriate subdirectory
2. Include comments explaining what SHOULD or should NOT be detected
3. Use realistic but obviously fake credentials for secrets
4. Update this README if adding new categories
