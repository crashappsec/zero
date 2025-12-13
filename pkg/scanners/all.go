// Package scanners imports all scanner implementations to register them
package scanners

// Import all scanner packages to trigger their init() functions
// which register the scanners with the scanner.Registry
//
// Scanner Architecture (v3.2):
// - sbom: SBOM generation and integrity (source of truth for package data)
// - packages: Package analysis features (depends on sbom output)
// - crypto: Cryptographic security analysis
// - code-security: Security-focused code analysis (vulns, secrets, api)
// - quality: Code quality analysis (tech-debt, complexity, coverage, docs)
// - devops: DevOps and CI/CD security (includes GitHub Actions)
// - health: Repository health metrics
// - ai: AI/ML security and ML-BOM generation
import (
	// Super scanners (v3.2)
	_ "github.com/crashappsec/zero/pkg/scanners/ai"            // AI/ML security
	_ "github.com/crashappsec/zero/pkg/scanners/code-security" // Security-focused code analysis
	_ "github.com/crashappsec/zero/pkg/scanners/crypto"
	_ "github.com/crashappsec/zero/pkg/scanners/devops"
	_ "github.com/crashappsec/zero/pkg/scanners/health"
	_ "github.com/crashappsec/zero/pkg/scanners/packages" // Depends on sbom
	_ "github.com/crashappsec/zero/pkg/scanners/quality"  // Code quality analysis
	_ "github.com/crashappsec/zero/pkg/scanners/sbom"     // Must run first - source of truth
)
