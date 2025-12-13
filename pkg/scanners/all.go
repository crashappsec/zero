// Package scanners imports all scanner implementations to register them
package scanners

// Import all scanner packages to trigger their init() functions
// which register the scanners with the scanner.Registry
//
// Scanner Architecture (v3.0):
// - sbom: SBOM generation and integrity (source of truth for package data)
// - packages: Package analysis features (depends on sbom output)
// - crypto: Cryptographic security analysis
// - code: Code security analysis
// - devops: DevOps and CI/CD security (formerly infra, includes GitHub Actions)
// - health: Repository health metrics
import (
	// Super scanners (v3.0)
	_ "github.com/crashappsec/zero/pkg/scanners/sbom"     // Must run first - source of truth
	_ "github.com/crashappsec/zero/pkg/scanners/packages" // Depends on sbom
	_ "github.com/crashappsec/zero/pkg/scanners/crypto"
	_ "github.com/crashappsec/zero/pkg/scanners/code"
	_ "github.com/crashappsec/zero/pkg/scanners/devops" // Renamed from infra, absorbed github-actions-security
	_ "github.com/crashappsec/zero/pkg/scanners/health"
)
