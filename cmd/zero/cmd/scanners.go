// Package scanners imports all scanner implementations to register them
package cmd

// Import all scanner packages to trigger their init() functions
// which register the scanners with the scanner.Registry
//
// Scanner Architecture (v3.6):
// - sbom: SBOM generation and integrity (source of truth for package data)
// - packages: Package analysis features (depends on sbom output)
// - code-crypto: Cryptographic security analysis
// - code-security: Security-focused code analysis (vulns, secrets, api)
// - code-quality: Code quality analysis (tech-debt, complexity, coverage, docs)
// - devops: DevOps and CI/CD security (includes GitHub Actions, DORA metrics)
// - tech-id: Technology identification and AI/ML security (ML-BOM generation)
// - code-ownership: Code ownership and CODEOWNERS analysis
// - developer-experience: Developer experience analysis (onboarding, tooling, workflow)
import (
	// Super scanners (v3.6)
	_ "github.com/crashappsec/zero/pkg/scanner/code-ownership"        // Code ownership analysis
	_ "github.com/crashappsec/zero/pkg/scanner/code-quality"          // Code quality analysis
	_ "github.com/crashappsec/zero/pkg/scanner/code-security"         // Security-focused code analysis
	_ "github.com/crashappsec/zero/pkg/scanner/code-crypto"           // Cryptographic security
	_ "github.com/crashappsec/zero/pkg/scanner/developer-experience"  // Developer experience analysis
	_ "github.com/crashappsec/zero/pkg/scanner/devops"                // DevOps and CI/CD security
	_ "github.com/crashappsec/zero/pkg/scanner/packages"              // Package analysis (depends on sbom)
	_ "github.com/crashappsec/zero/pkg/scanner/sbom"                  // SBOM generation (source of truth)
	_ "github.com/crashappsec/zero/pkg/scanner/tech-id"               // Technology and AI/ML security
)
