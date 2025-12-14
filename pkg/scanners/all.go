// Package scanners imports all scanner implementations to register them
package scanners

// Import all scanner packages to trigger their init() functions
// which register the scanners with the scanner.Registry
//
// Scanner Architecture (v3.4):
// - sbom: SBOM generation and integrity (source of truth for package data)
// - packages: Package analysis features (depends on sbom output)
// - crypto: Cryptographic security analysis
// - code-security: Security-focused code analysis (vulns, secrets, api)
// - code-quality: Code quality analysis (tech-debt, complexity, coverage, docs)
// - devops: DevOps and CI/CD security (includes GitHub Actions)
// - technology-identification: Technology identification and AI/ML security (ML-BOM generation)
// - code-ownership: Code ownership and CODEOWNERS analysis
// - code-health: Repository health aggregator
import (
	// Super scanners (v3.4)
	_ "github.com/crashappsec/zero/pkg/scanners/code-health"               // Project health metrics
	_ "github.com/crashappsec/zero/pkg/scanners/code-ownership"            // Code ownership analysis
	_ "github.com/crashappsec/zero/pkg/scanners/code-quality"              // Code quality analysis
	_ "github.com/crashappsec/zero/pkg/scanners/code-security"             // Security-focused code analysis
	_ "github.com/crashappsec/zero/pkg/scanners/crypto"                    // Cryptographic security
	_ "github.com/crashappsec/zero/pkg/scanners/devops"                    // DevOps and CI/CD security
	_ "github.com/crashappsec/zero/pkg/scanners/packages"                  // Package analysis (depends on sbom)
	_ "github.com/crashappsec/zero/pkg/scanners/sbom"                      // SBOM generation (source of truth)
	_ "github.com/crashappsec/zero/pkg/scanners/technology-identification" // Technology and AI/ML security
)
