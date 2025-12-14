# Go Modules Package Manager

**Ecosystem**: Go
**Package Registry**: https://proxy.golang.org (default proxy)
**Documentation**: https://go.dev/ref/mod

---

## TIER 1: Manifest Detection

### Manifest Files

| File | Required | Description |
|------|----------|-------------|
| `go.mod` | Yes | Module definition and dependencies |
| `go.work` | No | Workspace definition (Go 1.18+) |

### go.mod Detection

**Pattern**: `go\.mod$`
**Confidence**: 98% (HIGH)

### Required Sections for SBOM

```go
module github.com/myorg/myproject

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/go-sql-driver/mysql v1.7.1
)

require (
    // indirect dependencies
    golang.org/x/crypto v0.14.0 // indirect
    golang.org/x/sys v0.13.0 // indirect
)
```

### Dependency Types

| Section | Included in SBOM | Notes |
|---------|------------------|-------|
| `require` (direct) | Yes (always) | Direct dependencies |
| `require` (indirect) | Yes (always) | Transitive dependencies |
| `replace` | Yes | Local or version replacements |
| `exclude` | Metadata | Excluded versions |
| `retract` | Metadata | Retracted versions of this module |

---

## TIER 2: Lock File Detection

### Checksum File

| File | Format | Purpose |
|------|--------|---------|
| `go.sum` | Text | Cryptographic checksums |

**Pattern**: `go\.sum$`
**Confidence**: 98% (HIGH)

### go.sum Structure

```
github.com/gin-gonic/gin v1.9.1 h1:4idEAncQnU5cB7BeOkPtxjfCSye0AAm1R0RVIqJ+Jmg=
github.com/gin-gonic/gin v1.9.1/go.mod h1:hPrL4YkXt7akaWXRj/jnMGJ6YSzMT6hH8mC0Y4Q4mM=
golang.org/x/crypto v0.14.0 h1:wBqGXzWJW6m1XrIKlAH0Hs1JJ7+9KBwnIO8v66Q9cHc=
golang.org/x/crypto v0.14.0/go.mod h1:MVFd36DqK4CsrnJYDkBA3VC4m2GkXAM0PvzMCn4JQf4=
```

### Key go.sum Fields

| Field | SBOM Use |
|-------|----------|
| Module path | Package identifier |
| Version | Exact version (semantic or pseudo) |
| `h1:` hash | SHA-256 of module zip |
| `/go.mod` hash | SHA-256 of go.mod file |

---

## TIER 3: Configuration Extraction

### Proxy Configuration

**Environment Variables**:

| Variable | Purpose | Default |
|----------|---------|---------|
| `GOPROXY` | Module proxy URL(s) | `https://proxy.golang.org,direct` |
| `GOPRIVATE` | Private module patterns | (none) |
| `GONOPROXY` | Patterns to skip proxy | (none) |
| `GONOSUMDB` | Patterns to skip checksum DB | (none) |
| `GOSUMDB` | Checksum database | `sum.golang.org` |

### Common Configuration

```bash
# Use private registry
export GOPROXY="https://goproxy.mycompany.com,https://proxy.golang.org,direct"

# Private modules (skip proxy and sumdb)
export GOPRIVATE="github.com/mycompany/*,gitlab.mycompany.com/*"

# Corporate proxy with fallback
export GOPROXY="https://goproxy.mycompany.com,https://proxy.golang.org,direct"
export GONOSUMDB="github.com/mycompany/*"
```

### .netrc for Private Repos

```
machine github.com
login USERNAME
password TOKEN

machine gitlab.mycompany.com
login USERNAME
password TOKEN
```

---

## SBOM Generation

### Using cdxgen

```bash
# Install cdxgen
npm install -g @cyclonedx/cdxgen

# Generate SBOM
cdxgen -o sbom.json

# From go.mod only
cdxgen --project-type go -o sbom.json
```

### Using cyclonedx-gomod

```bash
# Install
go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest

# Generate SBOM
cyclonedx-gomod mod -json -output sbom.json

# Include test dependencies
cyclonedx-gomod mod -test -json -output sbom.json

# From specific module
cyclonedx-gomod mod -json -output sbom.json ./path/to/module
```

### Using syft

```bash
# Generate from directory
syft . -o cyclonedx-json > sbom.json

# Specify cataloger
syft . --select-catalogers go-module-file -o cyclonedx-json

# From compiled binary
syft ./mybinary -o cyclonedx-json > sbom.json
```

### Using govulncheck for Reachability

```bash
# Install govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest

# Check for reachable vulnerabilities
govulncheck ./...

# JSON output
govulncheck -json ./...
```

### Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `-test` | Include test dependencies | Exclude |
| `-std` | Include standard library | Exclude |
| `-licenses` | Detect licenses | Exclude |
| `-noserial` | Omit serial number | Include |

---

## Cache Locations

| Location | Path |
|----------|------|
| Module Cache | `$GOPATH/pkg/mod/` or `~/go/pkg/mod/` |
| Build Cache | `~/.cache/go-build/` (Linux/macOS) |
| Download Cache | `$GOPATH/pkg/mod/cache/download/` |

```bash
# Find module cache
go env GOMODCACHE

# Clean module cache
go clean -modcache

# Download dependencies
go mod download
```

---

## Best Practices

1. **Always commit go.sum** for reproducible builds
2. **Use `go mod tidy`** to clean up unused dependencies
3. **Vendor dependencies** for air-gapped builds with `go mod vendor`
4. **Use `go mod verify`** to verify checksums
5. **Pin versions explicitly** in go.mod
6. **Use Go workspaces** for multi-module projects

### Vendoring for Air-Gapped Builds

```bash
# Create vendor directory
go mod vendor

# Build with vendor
go build -mod=vendor ./...

# Verify vendor matches go.sum
go mod verify
```

---

## Troubleshooting

### Missing go.sum Entries
```bash
# Update go.sum
go mod tidy

# Download and update checksums
go mod download
```

### Module Not Found
```bash
# Clear module cache
go clean -modcache

# Re-download
go mod download
```

### Private Module Issues
```bash
# Configure GOPRIVATE
export GOPRIVATE="github.com/mycompany/*"

# Or use .netrc for authentication
```

### Version Conflicts
```bash
# See why a version was selected
go mod why -m <module>

# See module graph
go mod graph | grep <module>

# Use replace directive
# In go.mod:
# replace github.com/old/module => github.com/new/module v1.0.0
```

### Checksum Mismatch
```bash
# Verify checksums
go mod verify

# If mismatch, clear and re-download
go clean -modcache
go mod download
```

---

## Go Workspaces (1.18+)

For multi-module projects:

```go
// go.work
go 1.21

use (
    ./module-a
    ./module-b
    ./module-c
)
```

**SBOM Considerations**:
- Generate SBOM per module or for workspace
- Use `cyclonedx-gomod app` for applications
- Consider using workspace-aware tooling

---

## Binary Analysis

Go binaries embed dependency information:

```bash
# Extract dependencies from binary
go version -m ./mybinary

# Generate SBOM from binary with syft
syft ./mybinary -o cyclonedx-json > sbom.json
```

---

## References

- [Go Modules Reference](https://go.dev/ref/mod)
- [Go Module Proxy Protocol](https://go.dev/ref/mod#module-proxy)
- [cyclonedx-gomod](https://github.com/CycloneDX/cyclonedx-gomod)
- [govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck)
- [Go Checksum Database](https://sum.golang.org/)
