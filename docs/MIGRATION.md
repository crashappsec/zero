# Migration to Go CLI

This document tracks the migration from shell-based scanners to the Go CLI implementation.

## Summary

Zero has been rewritten in Go. The Go implementation provides:
- 25 scanners with self-registration
- Proper dependency ordering via topological sort
- Configurable scanner options via JSON
- Token permission checking via `roadmap` command
- Better error handling and fallback support

## Scanner Migration Status

### Fully Migrated (Go Implementation Complete)

| Shell Script | Go Scanner | Status |
|--------------|------------|--------|
| `utils/scanners/package-sbom/package-sbom.sh` | `pkg/scanners/sbom/sbom.go` | Complete - Supports cdxgen/syft with config |
| `utils/scanners/package-vulns/package-vulns.sh` | `pkg/scanners/vulns/vulns.go` | Complete - Uses osv-scanner/grype |
| `utils/scanners/licenses/licenses.sh` | `pkg/scanners/licenses/licenses.go` | Complete |
| `utils/scanners/package-health/package-health.sh` | `pkg/scanners/health/health.go` | Complete |
| `utils/scanners/package-malcontent/package-malcontent.sh` | `pkg/scanners/malcontent/malcontent.go` | Complete |
| `utils/scanners/package-provenance/package-provenance.sh` | `pkg/scanners/provenance/provenance.go` | Complete |
| `utils/scanners/package-bundle-optimization/package-bundle-optimization.sh` | `pkg/scanners/bundle/bundle.go` | Complete |
| `utils/scanners/package-recommendations/package-recommendations.sh` | `pkg/scanners/recommendations/recommendations.go` | Complete |
| `utils/scanners/code-vulns/code-vulns.sh` | `pkg/scanners/codevulns/codevulns.go` | Complete |
| `utils/scanners/code-secrets/code-secrets.sh` | `pkg/scanners/secrets/secrets.go` | Complete |
| `utils/scanners/api-security/api-security.sh` | `pkg/scanners/api/api.go` | Complete |
| `utils/scanners/iac-security/iac-security.sh` | `pkg/scanners/iac/iac.go` | Complete |
| `utils/scanners/container-security/container-security.sh` | `pkg/scanners/container/container.go` | Complete |
| `utils/scanners/containers/containers.sh` | `pkg/scanners/containers/containers.go` | Complete |
| `utils/scanners/code-crypto (ciphers)/code-crypto (ciphers).sh` | `pkg/scanners/code-crypto (ciphers)/code-crypto (ciphers).go` | Complete |
| `utils/scanners/code-crypto (keys)/code-crypto (keys).sh` | `pkg/scanners/code-crypto (keys)/code-crypto (keys).go` | Complete |
| `utils/scanners/code-crypto (random)/code-crypto (random).sh` | `pkg/scanners/code-crypto (random)/code-crypto (random).go` | Complete |
| `utils/scanners/code-crypto (tls)/code-crypto (tls).sh` | `pkg/scanners/code-crypto (tls)/code-crypto (tls).go` | Complete |
| `utils/scanners/digital-certificates/digital-certificates.sh` | `pkg/scanners/certs/certs.go` | Complete |
| `utils/scanners/tech-discovery/tech-discovery.sh` | `pkg/scanners/tech/tech.go` | Complete |
| `utils/scanners/code-ownership/code-ownership.sh` | `pkg/scanners/ownership/ownership.go` | Complete |
| `utils/scanners/dora/dora.sh` | `pkg/scanners/dora/dora.go` | Complete |
| `utils/scanners/git/git.sh` | `pkg/scanners/git/git.go` | Complete |
| `utils/scanners/documentation/documentation.sh` | `pkg/scanners/docs/docs.go` | Complete |
| `utils/scanners/test-coverage/test-coverage.sh` | `pkg/scanners/testcoverage/testcoverage.go` | Complete |
| `utils/scanners/tech-debt/tech-debt.sh` | `pkg/scanners/techdebt/techdebt.go` | Complete |

### Shell Scripts That Can Be Retired

These shell scripts have Go replacements and can be safely removed:

```
utils/scanners/package-sbom/package-sbom.sh
utils/scanners/package-vulns/package-vulns.sh
utils/scanners/licenses/licenses.sh
utils/scanners/package-health/package-health.sh
utils/scanners/package-malcontent/package-malcontent.sh
utils/scanners/package-provenance/package-provenance.sh
utils/scanners/package-bundle-optimization/package-bundle-optimization.sh
utils/scanners/package-recommendations/package-recommendations.sh
utils/scanners/code-vulns/code-vulns.sh
utils/scanners/code-secrets/code-secrets.sh
utils/scanners/api-security/api-security.sh
utils/scanners/iac-security/iac-security.sh
utils/scanners/container-security/container-security.sh
utils/scanners/containers/containers.sh
utils/scanners/code-crypto (ciphers)/code-crypto (ciphers).sh
utils/scanners/code-crypto (keys)/code-crypto (keys).sh
utils/scanners/code-crypto (random)/code-crypto (random).sh
utils/scanners/code-crypto (tls)/code-crypto (tls).sh
utils/scanners/digital-certificates/digital-certificates.sh
utils/scanners/tech-discovery/tech-discovery.sh
utils/scanners/code-ownership/code-ownership.sh
utils/scanners/dora/dora.sh
utils/scanners/git/git.sh
utils/scanners/documentation/documentation.sh
utils/scanners/test-coverage/test-coverage.sh
utils/scanners/tech-debt/tech-debt.sh
```

### Shell Scripts to Keep (Agent System / Claude Integration)

These are used by the Claude Code agent system and should be kept:

```
utils/zero/lib/agent-loader.sh      # Agent context loading
utils/zero/scripts/agent.sh         # Agent mode entry point
utils/lib/agent-personality.sh      # Agent personalities
```

### Shell Scripts to Keep (Utilities)

These provide useful utilities that aren't part of the scanner system:

```
utils/lib/github.sh                 # GitHub API helpers (may migrate later)
utils/lib/sbom.sh                   # SBOM parsing utilities (may migrate later)
utils/lib/markdown.sh               # Markdown generation (may migrate later)
```

### Shell Scripts to Keep (Testing)

Test scripts should be kept until Go equivalents exist:

```
utils/scanners/*/run-tests.sh       # Scanner test harnesses
utils/scanners/*/run-all-tests.sh   # Full test suites
```

## Python Scripts

Only 4 Python scripts exist in the repo (excluding cloned repos):

| Script | Purpose | Status |
|--------|---------|--------|
| `utils/scanners/semgrep/rag-to-semgrep.py` | Convert RAG to semgrep rules | Keep - unique tool |
| `utils/scanners/semgrep/rag-security-to-semgrep.py` | Convert security RAG | Keep - unique tool |
| `utils/scanners/code-security/tests/test-samples/sql-injection.py` | Test fixture | Keep - test data |
| `utils/scanners/code-security/tests/test-samples/hardcoded-secrets.py` | Test fixture | Keep - test data |

## Commands Migration

| Shell Command | Go Command | Status |
|---------------|------------|--------|
| `./zero hydrate` | `./zero hydrate` | Complete |
| `./zero scan` | `./zero scan` | Complete |
| `./zero status` | `./zero status` | Complete |
| `./zero report` | `./zero report` | Complete |
| `./zero check` | `./zero check` | Complete |
| N/A | `./zero roadmap` | New - Token analysis |
| N/A | `./zero list` | New - Scanner listing |
| N/A | `./zero history` | New - Scan history |
| N/A | `./zero clean` | New - Data cleanup |

## Configuration Migration

The Go implementation uses `utils/zero/config/zero.config.json` for configuration:

```json
{
  "settings": {
    "default_profile": "standard",
    "scanner_timeout_seconds": 300,
    "parallel_jobs": 4
  },
  "scanners": {
    "package-sbom": {
      "options": {
        "sbom": {
          "tool": "auto",
          "spec_version": "1.5",
          "recurse": true,
          "install_deps": false,
          "fallback_to_syft": true
        }
      }
    }
  }
}
```

## Building and Running

```bash
# Build the Go CLI
go build -o zero ./cmd/zero

# Run
./zero hydrate owner/repo

# Check what scanners will work
./zero roadmap

# List all scanners
./zero list
```

## Architecture

```
cmd/zero/
├── main.go           # Entry point
└── cmd/
    ├── root.go       # Cobra root command
    ├── hydrate.go    # Clone + scan
    ├── scan.go       # Scan existing repo
    ├── status.go     # Show projects
    ├── report.go     # Generate reports
    ├── roadmap.go    # Token analysis (NEW)
    ├── list.go       # List scanners (NEW)
    └── ...

pkg/
├── scanner/          # Scanner framework
│   ├── scanner.go    # Interface + types
│   ├── registry.go   # Scanner registration
│   └── runner.go     # Execution engine
├── scanners/         # Scanner implementations
│   ├── all.go        # Import all scanners
│   ├── sbom/
│   ├── vulns/
│   └── ...
├── config/           # Configuration
├── github/           # GitHub API + permissions
├── hydrate/          # Orchestration
└── terminal/         # UI helpers
```

## Next Steps

1. **Remove retired shell scanners** - Delete the scripts listed in "Shell Scripts That Can Be Retired"
2. **Update zero.sh** - Point to Go binary instead of shell scripts
3. **Migrate remaining utilities** - github.sh, sbom.sh, markdown.sh
4. **Add more scanner tests** - Expand Go test coverage
