// Package scanners imports all scanner implementations to register them
package cmd

// Import all scanner packages to trigger their init() functions
// which register the scanners with the scanner.Registry
//
// Scanner Architecture (v4.0):
// - supply-chain: SBOM generation and package analysis (14 features)
// - code-security: Security-focused code analysis (vulns, secrets, api, crypto - 8 features)
// - code-quality: Code quality analysis (tech-debt, complexity, coverage, docs)
// - devops: DevOps and CI/CD security (includes GitHub Actions, DORA metrics)
// - tech-id: Technology identification and AI/ML security (ML-BOM generation)
// - code-ownership: Code ownership and CODEOWNERS analysis
// - developer-experience: Developer experience analysis (onboarding, tooling, workflow)
import (
	// Super scanners (v4.0)
	_ "github.com/crashappsec/zero/pkg/scanner/code-ownership"       // Code ownership analysis
	_ "github.com/crashappsec/zero/pkg/scanner/code-quality"         // Code quality analysis
	_ "github.com/crashappsec/zero/pkg/scanner/code-security"        // Security-focused code analysis + crypto
	_ "github.com/crashappsec/zero/pkg/scanner/developer-experience" // Developer experience analysis
	_ "github.com/crashappsec/zero/pkg/scanner/devops"               // DevOps and CI/CD security
	_ "github.com/crashappsec/zero/pkg/scanner/supply-chain"         // SBOM + package analysis
	_ "github.com/crashappsec/zero/pkg/scanner/tech-id"              // Technology and AI/ML security
)
