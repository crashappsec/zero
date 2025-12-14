// Package scanners imports all scanner implementations to register them
package scanners

// Import all scanner packages to trigger their init() functions
// which register the scanners with the scanner.Registry
//
// Scanner Architecture (v3.3):
// - sbom: SBOM generation and integrity (source of truth for package data)
// - packages: Package analysis features (depends on sbom output)
// - crypto: Cryptographic security analysis
// - code-security: Security-focused code analysis (vulns, secrets, api)
// - quality: Code quality analysis (tech-debt, complexity, coverage, docs)
// - devops: DevOps and CI/CD security (includes GitHub Actions)
// - technology: Technology identification and AI/ML security (ML-BOM generation)
// - ownership: Code ownership and CODEOWNERS analysis
// - health: Repository health aggregator
import (
	// Super scanners (v3.3)
	_ "github.com/crashappsec/zero/pkg/scanners/code-security" // Security-focused code analysis
	_ "github.com/crashappsec/zero/pkg/scanners/crypto"
	_ "github.com/crashappsec/zero/pkg/scanners/devops"
	_ "github.com/crashappsec/zero/pkg/scanners/health"
	_ "github.com/crashappsec/zero/pkg/scanners/ownership"   // Code ownership analysis
	_ "github.com/crashappsec/zero/pkg/scanners/packages"    // Depends on sbom
	_ "github.com/crashappsec/zero/pkg/scanners/quality"     // Code quality analysis
	_ "github.com/crashappsec/zero/pkg/scanners/sbom"        // Must run first - source of truth
	_ "github.com/crashappsec/zero/pkg/scanners/technology"  // Technology identification and AI/ML
)
