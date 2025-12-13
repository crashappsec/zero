# Legacy Shell Implementation

This folder contains the original shell-based Zero implementation that has been superseded by the Go rewrite.

## Contents

- `zero.sh` - Original shell-based CLI (47KB)
- `test-scan-debug.sh` - Debug script for testing
- `utils/` - All shell-based utilities and scanners
  - `utils/scanners/` - Shell scanner implementations
  - `utils/lib/` - Shared shell libraries
  - `utils/zero/` - Zero orchestrator scripts and config

## Migration

Zero has been completely rewritten in Go. Use the new CLI:

```bash
# Build
go build -o main ./cmd/zero

# Run
./main hydrate owner/repo
./main hydrate phantom-tests
./main checkup
./main list
```

The Go implementation is faster, type-safe, and easier to maintain.

See [MIGRATION.md](../docs/MIGRATION.md) for details on the migration from shell to Go.
