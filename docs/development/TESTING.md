# Testing Strategy

> Test strategy and coverage goals for Zero.

## Coverage Goals

| Package | Current | Target | Priority |
|---------|---------|--------|----------|
| `pkg/core/sarif` | 85% | 90% | Low |
| `pkg/core/errors` | 80% | 85% | Low |
| `pkg/core/feeds` | 75% | 80% | Low |
| `pkg/core/rag` | 60% | 80% | High |
| `pkg/scanner/code-security` | 28% | 70% | High |
| `pkg/scanner/code-packages` | 8% | 70% | High |
| `pkg/workflow/hydrate` | 17% | 70% | Medium |
| `pkg/api/handlers` | 0% | 70% | High |
| `pkg/core/scoring` | 0% | 70% | Medium |

**Overall Target:** 70% coverage on critical paths.

---

## Test Types

### Unit Tests

Fast, isolated tests for individual functions.

```go
// pkg/scanner/code-security/security_test.go
func TestDetectSecrets(t *testing.T) {
    tests := []struct {
        name     string
        content  string
        expected int
    }{
        {
            name:     "AWS key",
            content:  `aws_key = "AKIAIOSFODNN7EXAMPLE"`,
            expected: 1,
        },
        {
            name:     "No secrets",
            content:  `config = "safe_value"`,
            expected: 0,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            findings := detectSecrets(tt.content)
            if len(findings) != tt.expected {
                t.Errorf("got %d findings, want %d", len(findings), tt.expected)
            }
        })
    }
}
```

### Integration Tests

Test component interactions with real dependencies.

```go
// pkg/scanner/code-security/integration_test.go
//go:build integration

func TestSemgrepIntegration(t *testing.T) {
    if _, err := exec.LookPath("semgrep"); err != nil {
        t.Skip("semgrep not installed")
    }

    scanner := NewCodeSecurityScanner()
    result, err := scanner.Run(context.Background(), &ScanOptions{
        Target: "testdata/vulnerable",
    })

    require.NoError(t, err)
    assert.NotEmpty(t, result.Findings)
}
```

### End-to-End Tests

Full workflow tests.

```go
// pkg/workflow/hydrate/e2e_test.go
//go:build e2e

func TestHydrateWorkflow(t *testing.T) {
    // Create temp directory
    tmpDir := t.TempDir()

    // Run hydrate
    err := Hydrate(context.Background(), &HydrateOptions{
        Target:   "testdata/sample-repo",
        ZeroHome: tmpDir,
        Profile:  "quick",
    })

    require.NoError(t, err)

    // Verify outputs
    assert.FileExists(t, filepath.Join(tmpDir, "repos/sample-repo/analysis/code-packages.json"))
    assert.FileExists(t, filepath.Join(tmpDir, "repos/sample-repo/analysis/code-security.json"))
}
```

---

## Test Organization

### Directory Structure

```
pkg/scanner/code-security/
├── security.go
├── security_test.go          # Unit tests
├── security_integration_test.go  # Integration tests (build tag)
└── testdata/
    ├── vulnerable/           # Known vulnerable code
    │   ├── sql_injection.py
    │   ├── xss.js
    │   └── hardcoded_secret.go
    └── safe/                 # Known safe code
        └── sanitized.py
```

### Build Tags

```go
// Unit tests (default)
// No build tag needed

// Integration tests
//go:build integration

// E2E tests
//go:build e2e

// Run specific tags:
// go test -tags=integration ./...
// go test -tags=e2e ./...
```

---

## Test Fixtures

### Scanner Test Data

Create fixtures for each scanner type:

```
testdata/
├── sbom/
│   ├── npm-project/          # package.json + package-lock.json
│   ├── go-project/           # go.mod + go.sum
│   └── python-project/       # requirements.txt
├── vulnerabilities/
│   ├── cve-2021-44228/       # Log4j vulnerable code
│   └── cve-2022-22965/       # Spring4Shell
├── secrets/
│   ├── aws-keys.txt
│   ├── api-tokens.json
│   └── .env.example
└── iac/
    ├── terraform/
    ├── kubernetes/
    └── dockerfile/
```

### Golden Files

For complex output validation:

```go
func TestScannerOutput(t *testing.T) {
    result, _ := scanner.Run(ctx, opts)

    golden := filepath.Join("testdata", "golden", t.Name()+".json")

    if *update {
        // Update golden file
        os.WriteFile(golden, result.JSON(), 0644)
        return
    }

    expected, _ := os.ReadFile(golden)
    assert.JSONEq(t, string(expected), string(result.JSON()))
}
```

---

## Mocking

### External Tool Mocks

```go
// Mock semgrep execution
type mockSemgrepRunner struct {
    output string
    err    error
}

func (m *mockSemgrepRunner) Run(ctx context.Context, args []string) ([]byte, error) {
    return []byte(m.output), m.err
}

func TestSemgrepScanner_WithMock(t *testing.T) {
    scanner := &SemgrepScanner{
        runner: &mockSemgrepRunner{
            output: `{"results": []}`,
        },
    }
    // ...
}
```

### HTTP Mocks

```go
func TestOSVClient(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"vulns": []}`))
    }))
    defer server.Close()

    client := NewOSVClient(WithBaseURL(server.URL))
    vulns, err := client.Query(context.Background(), "lodash", "4.17.20")

    require.NoError(t, err)
    assert.Empty(t, vulns)
}
```

---

## Web UI Testing

### Component Tests

```tsx
// web/src/components/ScanResults.test.tsx
import { render, screen, waitFor } from '@testing-library/react';
import { ScanResults } from './ScanResults';

describe('ScanResults', () => {
  it('displays loading state', () => {
    render(<ScanResults jobId="123" />);
    expect(screen.getByText(/loading/i)).toBeInTheDocument();
  });

  it('displays findings', async () => {
    render(<ScanResults jobId="123" />);
    await waitFor(() => {
      expect(screen.getByText(/vulnerabilities/i)).toBeInTheDocument();
    });
  });
});
```

### Hook Tests

```tsx
// web/src/hooks/useApi.test.ts
import { renderHook, waitFor } from '@testing-library/react';
import { useChat } from './useApi';

describe('useChat', () => {
  it('sends message and receives response', async () => {
    const { result } = renderHook(() => useChat('zero'));

    await result.current.sendMessage('Hello');

    await waitFor(() => {
      expect(result.current.messages).toHaveLength(2);
      expect(result.current.messages[1].role).toBe('assistant');
    });
  });
});
```

---

## CI Integration

### GitHub Actions Test Job

```yaml
test:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with:
        go-version: '1.22'

    # Unit tests
    - run: go test -v -race -coverprofile=coverage.out ./...

    # Integration tests (with dependencies)
    - name: Install dependencies
      run: |
        pip install semgrep
        npm install -g @cyclonedx/cdxgen

    - run: go test -v -tags=integration ./...

    # Upload coverage
    - uses: codecov/codecov-action@v4
      with:
        files: coverage.out
```

### Coverage Requirements

```yaml
# codecov.yml
coverage:
  status:
    project:
      default:
        target: 70%
        threshold: 2%
    patch:
      default:
        target: 80%
```

---

## Running Tests

### Quick Reference

```bash
# Run all unit tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package
go test -v ./pkg/scanner/code-security/...

# Run integration tests
go test -tags=integration ./...

# Run E2E tests
go test -tags=e2e ./...

# Run with race detection
go test -race ./...

# Web UI tests
cd web && npm test

# Web UI coverage
cd web && npm run test:coverage
```

### Test Debugging

```bash
# Verbose output
go test -v ./pkg/scanner/... 2>&1 | tee test.log

# Run single test
go test -v -run TestDetectSecrets ./pkg/scanner/code-security/

# With timeout
go test -timeout 5m ./...

# Keep test cache
go test -count=1 ./...  # Disable cache
```

---

## Writing Good Tests

### Do's

- Test behavior, not implementation
- Use table-driven tests for multiple cases
- Test error conditions
- Use meaningful test names
- Clean up resources (use `t.Cleanup()`)

### Don'ts

- Don't test private functions directly
- Don't depend on test execution order
- Don't use sleep for synchronization
- Don't ignore flaky tests
- Don't commit skipped tests without issue link
