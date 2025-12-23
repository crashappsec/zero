<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Zero Pattern Architecture

This document describes the canonical architecture for pattern-based detection and data enrichment in Zero. **All scanners MUST follow this architecture.**

## Principle: Separation of Detection and Enrichment

Zero uses TWO distinct data sources for analysis:

| Type | Purpose | Source | Examples |
|------|---------|--------|----------|
| **Detection Patterns** | Signatures to find things in code | RAG → Semgrep | Tech-id, crypto, Dockerfile lint, API quality |
| **Enrichment Data** | External data to enhance findings | Live APIs | OSV.dev (vulns), deps.dev (health, deprecation) |

---

## 1. Detection Patterns (RAG → Semgrep)

For finding things in code - signatures, patterns, rules.

### Data Flow

```
rag/<category>/<subcategory>/patterns.md
        ↓
pkg/core/rules/manager.go (generateFromRAG)
        ↓
.zero/rules/generated/<category>.json
        ↓
Semgrep execution
        ↓
Scanner findings
```

### RAG Categories

| Category | Description | Scanner |
|----------|-------------|---------|
| `technology-identification` | Detect frameworks, languages, tools | tech-id |
| `secrets` | API keys, credentials, tokens | code-security |
| `devops-security` | Dockerfile, IaC, CI/CD patterns | devops |
| `code-security` | SAST patterns, API quality | code-security |
| `crypto` | Weak ciphers, hardcoded keys | crypto |

### Pattern File Format

```markdown
# Category Title

**Category**: category/subcategory
**Description**: What this category detects
**CWE**: CWE-XXX (Description)

---

## Pattern Name

### Language (e.g., Python, Javascript, Dockerfile)
**Type**: regex | semgrep
**Severity**: critical | high | medium | low | info
**Pattern**: `pattern-here`
- Description of what this detects
- Example: `code example`
- Remediation: How to fix
```

### Adding New Detection Patterns

1. **Create pattern file**: `rag/<category>/<subcategory>/patterns.md`
2. **Add category constant** (if new category): `pkg/core/rag/types.go`
   ```go
   const CategoryNewName RAGCategory = "category-name"
   ```
3. **Register category**: `pkg/core/rules/manager.go` in `generateFromRAG()`
   ```go
   categories := []rag.RAGCategory{
       rag.CategoryTechID,
       rag.CategorySecrets,
       rag.CategoryNewName,  // Add here
   }
   ```
4. **Use in scanner**: Call `runSemgrepWithRules(category)` instead of hardcoded regex

---

## 2. Enrichment Data (Live APIs)

For external data to enhance findings - NOT for detection.

### Enrichment Clients

| Client | API | Purpose | Cache TTL |
|--------|-----|---------|-----------|
| `liveapi/osv.go` | api.osv.dev | Vulnerability data | 15 min |
| `liveapi/deps.go` | api.deps.dev | Package health, deprecation, SLSA | 24 hours |

### Client Architecture

All enrichment clients extend the base `Client` type which provides:
- HTTP client with configurable timeout
- In-memory caching with TTL
- Rate limiting (token bucket)
- User-Agent identification

```go
type DepsDevClient struct {
    *Client
}

func NewDepsDevClient() *DepsDevClient {
    return &DepsDevClient{
        Client: NewClient("https://api.deps.dev/v3alpha",
            WithTimeout(30*time.Second),
            WithCache(24*time.Hour),
            WithRateLimit(10),
            WithUserAgent("Zero-Scanner/1.0"),
        ),
    }
}
```

### Adding New Enrichment Source

1. **Create client file**: `pkg/core/liveapi/<source>.go`
2. **Extend base Client**: Include caching and rate limiting
3. **Add to approved URLs**: `pkg/core/feeds/types.go`
   ```go
   PreApprovedURLs: []string{
       "https://api.osv.dev",
       "https://api.deps.dev",
       "https://api.newsource.com",  // Add here
   }
   ```
4. **Use in scanner**:
   ```go
   client := liveapi.NewSourceClient()
   data, err := client.GetData(ctx, params)
   ```

---

## NEVER Do These Things

### In Scanner Code

```go
// NEVER: Hardcoded regex patterns
var patterns = []struct{Pattern *regexp.Regexp}{...}  // ❌

// INSTEAD: Load from RAG/Semgrep
findings := s.runSemgrepWithRules("category")  // ✓
```

```go
// NEVER: Direct HTTP calls
resp, err := http.Get("https://api.example.com/...")  // ❌

// INSTEAD: Use liveapi client
client := liveapi.NewExampleClient()
data, err := client.Query(ctx, params)  // ✓
```

### Mixing Responsibilities

```go
// NEVER: Detection in enrichment code or vice versa
// Detection patterns belong in RAG
// Enrichment queries belong in liveapi
```

---

## Examples

### Good: Detection Pattern in RAG

```markdown
# rag/devops-security/dockerfile/patterns.md

## Running as Root

### Dockerfile
**Type**: regex
**Severity**: high
**Pattern**: `(?i)^USER\s+root\s*$`
- Running container as root is a security risk
- Remediation: Add `USER nonroot` before CMD
```

### Good: Enrichment in Scanner

```go
// pkg/scanner/packages/packages.go
func (s *Scanner) enrichWithHealth(ctx context.Context, pkg *Package) {
    client := liveapi.NewDepsDevClient()
    health, err := client.GetHealthScore(ctx, pkg.Ecosystem, pkg.Name, pkg.Version)
    if err == nil {
        pkg.HealthScore = health.Score
        pkg.IsDeprecated = health.IsDeprecated
    }
}
```

### Bad: Hardcoded Pattern in Scanner

```go
// ❌ DON'T DO THIS
var dockerfilePatterns = []struct{
    Pattern *regexp.Regexp
    Name    string
}{
    {regexp.MustCompile(`(?i)^USER\s+root`), "Running as root"},
}
```

---

## Migration Checklist

When refactoring existing scanners:

- [ ] Identify hardcoded regex patterns
- [ ] Create corresponding RAG pattern files
- [ ] Add category to `pkg/core/rag/types.go` (if new)
- [ ] Register category in `pkg/core/rules/manager.go`
- [ ] Update scanner to use Semgrep rules
- [ ] Remove hardcoded patterns
- [ ] Test findings match previous behavior

---

*Last Updated: 2025-12-23*
