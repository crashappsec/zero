# Coding Standards

> Style guidelines for Go and TypeScript in the Zero codebase.

## Go Standards

### Formatting

- Use `gofmt` and `goimports` (enforced by CI)
- Maximum line length: 120 characters
- Use tabs for indentation

### Naming Conventions

| Type | Convention | Example |
|------|------------|---------|
| Package | lowercase, short | `scanner`, `config` |
| Exported function | PascalCase | `RunScanner()` |
| Unexported function | camelCase | `parseConfig()` |
| Constants | PascalCase or SCREAMING_SNAKE | `MaxRetries`, `DEFAULT_TIMEOUT` |
| Interfaces | PascalCase, noun or -er suffix | `Scanner`, `ConfigLoader` |
| Structs | PascalCase | `ScanResult`, `Finding` |

### Error Handling

```go
// GOOD: Always check errors
result, err := doSomething()
if err != nil {
    return fmt.Errorf("doSomething failed: %w", err)
}

// BAD: Ignoring errors
result, _ := doSomething()  // Never do this

// GOOD: Wrap errors with context
if err := scanner.Run(ctx); err != nil {
    return fmt.Errorf("scanner %s failed: %w", scanner.Name(), err)
}
```

### Resource Management

```go
// GOOD: Always defer Close() immediately after opening
file, err := os.Open(path)
if err != nil {
    return err
}
defer file.Close()

// GOOD: Check Close() errors for writes
f, err := os.Create(path)
if err != nil {
    return err
}
defer func() {
    if cerr := f.Close(); cerr != nil && err == nil {
        err = cerr
    }
}()
```

### Context Usage

```go
// GOOD: Pass context as first parameter
func (s *Scanner) Run(ctx context.Context, target string) error {
    // Check for cancellation
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    // ... do work
}

// GOOD: Use context for HTTP requests
req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
```

### Struct Organization

```go
// Order: exported fields first, then unexported, grouped logically
type Scanner struct {
    // Configuration (exported)
    Name        string
    Description string
    Features    []string

    // Runtime state (unexported)
    client  *http.Client
    cache   map[string]interface{}
    mu      sync.RWMutex
}
```

### Comments

```go
// Package scanner provides security analysis capabilities.
package scanner

// Scanner defines the interface for all security scanners.
// Implementations must be safe for concurrent use.
type Scanner interface {
    // Run executes the scanner against the target directory.
    // Returns findings or an error if the scan fails.
    Run(ctx context.Context, target string) ([]Finding, error)
}

// RunAll executes all enabled scanners concurrently.
// It returns early if ctx is cancelled.
func RunAll(ctx context.Context, scanners []Scanner) error {
    // ... implementation
}
```

---

## TypeScript/React Standards

### Formatting

- Use Prettier with project config (enforced by CI)
- Maximum line length: 100 characters
- Use 2-space indentation
- Single quotes for strings
- Semicolons required

### Naming Conventions

| Type | Convention | Example |
|------|------------|---------|
| Components | PascalCase | `ScanResults`, `ChatMessage` |
| Hooks | camelCase with `use` prefix | `useChat`, `useScanProgress` |
| Functions | camelCase | `fetchData`, `handleSubmit` |
| Constants | SCREAMING_SNAKE | `API_BASE`, `MAX_RETRIES` |
| Types/Interfaces | PascalCase | `ScanJob`, `StreamChunk` |
| Props interfaces | PascalCase with Props suffix | `ButtonProps`, `ChatProps` |

### Component Structure

```tsx
// 1. Imports (external, then internal, then styles)
import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import type { ScanJob } from '@/lib/types';

// 2. Types
interface ScanResultsProps {
  jobId: string;
  onComplete?: () => void;
}

// 3. Component
export function ScanResults({ jobId, onComplete }: ScanResultsProps) {
  // 3a. Hooks (state, effects, custom hooks)
  const [data, setData] = useState<ScanJob | null>(null);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    // ... effect logic
  }, [jobId]);

  // 3b. Event handlers
  const handleRefresh = () => {
    // ...
  };

  // 3c. Render helpers (if needed)
  const renderStatus = () => {
    // ...
  };

  // 3d. Early returns for loading/error states
  if (error) return <ErrorDisplay error={error} />;
  if (!data) return <Loading />;

  // 3e. Main render
  return (
    <div>
      {/* ... */}
    </div>
  );
}
```

### Hooks Best Practices

```tsx
// GOOD: Proper dependency arrays
useEffect(() => {
  const controller = new AbortController();
  fetchData(controller.signal);
  return () => controller.abort();
}, [dependency]);  // List all dependencies

// GOOD: Stable callbacks with useCallback
const handleClick = useCallback((id: string) => {
  setSelected(id);
}, []);  // Empty if no dependencies

// BAD: Missing cleanup
useEffect(() => {
  const interval = setInterval(fetchData, 5000);
  // Missing: return () => clearInterval(interval);
}, []);
```

### Type Safety

```tsx
// GOOD: Explicit types for API responses
interface ApiResponse<T> {
  data: T;
  error?: string;
}

// GOOD: Use discriminated unions for state
type State =
  | { status: 'idle' }
  | { status: 'loading' }
  | { status: 'success'; data: ScanJob }
  | { status: 'error'; error: Error };

// BAD: Using 'any'
const data: any = response.json();  // Never do this

// GOOD: Use 'unknown' and type guards
const data: unknown = await response.json();
if (isScanJob(data)) {
  // data is now typed as ScanJob
}
```

### Error Handling

```tsx
// GOOD: Try-catch with proper error typing
try {
  await api.scans.start(target);
} catch (err) {
  const message = err instanceof Error ? err.message : 'Unknown error';
  setError(message);
}

// GOOD: Error boundaries for components
<ErrorBoundary fallback={<ErrorFallback />}>
  <ScanResults jobId={id} />
</ErrorBoundary>
```

---

## General Guidelines

### File Organization

```
pkg/scanner/code-security/
├── security.go          # Main scanner implementation
├── security_test.go     # Tests
├── rag_secrets.go       # RAG secret detection
├── ai_analysis.go       # AI-assisted analysis
└── testdata/            # Test fixtures
    ├── vulnerable/
    └── safe/
```

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

Examples:
```
feat(scanner): Add reachability analysis for vulnerabilities
fix(web): Resolve chat streaming freeze issue
docs(api): Update endpoint documentation
refactor(agents): Extract common delegation logic
test(scanner): Add integration tests for code-packages
chore(deps): Update Go dependencies
```

### Code Review Checklist

- [ ] Error handling is complete (no ignored errors)
- [ ] Resources are properly closed (defer patterns)
- [ ] Context is passed and checked for cancellation
- [ ] Types are explicit (no `any` in TypeScript)
- [ ] Tests are included for new functionality
- [ ] No hardcoded secrets or sensitive data
- [ ] Comments explain "why", not "what"
- [ ] Public APIs are documented
