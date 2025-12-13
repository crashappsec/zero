// Package health provides the consolidated project health super scanner
package health

// FeatureConfig holds configuration for all project health features
type FeatureConfig struct {
	Technology    TechnologyConfig    `json:"technology"`
	Documentation DocumentationConfig `json:"documentation"`
	Tests         TestsConfig         `json:"tests"`
	Ownership     OwnershipConfig     `json:"ownership"`
}

// TechnologyConfig configures technology discovery
type TechnologyConfig struct {
	Enabled         bool `json:"enabled"`
	ScanExtensions  bool `json:"scan_extensions"`   // Detect from file extensions
	ScanConfig      bool `json:"scan_config"`       // Detect from config files
	ScanSBOM        bool `json:"scan_sbom"`         // Detect from SBOM
}

// DocumentationConfig configures documentation analysis
type DocumentationConfig struct {
	Enabled          bool `json:"enabled"`
	CheckReadme      bool `json:"check_readme"`      // Check for README quality
	CheckCodeDocs    bool `json:"check_code_docs"`   // Check code documentation
	CheckChangelog   bool `json:"check_changelog"`   // Check for CHANGELOG
	CheckAPIDocs     bool `json:"check_api_docs"`    // Check for API documentation
}

// TestsConfig configures test coverage analysis
type TestsConfig struct {
	Enabled            bool    `json:"enabled"`
	ParseReports       bool    `json:"parse_reports"`       // Parse existing coverage reports
	AnalyzeInfra       bool    `json:"analyze_infra"`       // Analyze test infrastructure
	CoverageThreshold  float64 `json:"coverage_threshold"`  // Minimum coverage threshold (default 80)
}

// OwnershipConfig configures code ownership analysis
type OwnershipConfig struct {
	Enabled            bool `json:"enabled"`
	AnalyzeContributors bool `json:"analyze_contributors"` // Analyze git contributors
	CheckCodeowners    bool `json:"check_codeowners"`     // Validate CODEOWNERS file
	DetectOrphans      bool `json:"detect_orphans"`       // Find files with no recent commits
	PeriodDays         int  `json:"period_days"`          // Analysis period (default 90)
}

// DefaultConfig returns default feature configuration
func DefaultConfig() FeatureConfig {
	return FeatureConfig{
		Technology: TechnologyConfig{
			Enabled:        true,
			ScanExtensions: true,
			ScanConfig:     true,
			ScanSBOM:       true,
		},
		Documentation: DocumentationConfig{
			Enabled:       true,
			CheckReadme:   true,
			CheckCodeDocs: true,
			CheckChangelog: true,
			CheckAPIDocs:  true,
		},
		Tests: TestsConfig{
			Enabled:           true,
			ParseReports:      true,
			AnalyzeInfra:      true,
			CoverageThreshold: 80,
		},
		Ownership: OwnershipConfig{
			Enabled:             true,
			AnalyzeContributors: true,
			CheckCodeowners:     true,
			DetectOrphans:       true,
			PeriodDays:          90,
		},
	}
}

// QuickConfig returns minimal config for fast scans
func QuickConfig() FeatureConfig {
	cfg := DefaultConfig()
	cfg.Documentation.CheckCodeDocs = false  // Skip code doc analysis (slow)
	cfg.Ownership.AnalyzeContributors = false // Skip contributor analysis (slow)
	cfg.Ownership.DetectOrphans = false       // Skip orphan detection (slow)
	return cfg
}

// SecurityConfig returns security-focused config (minimal for health scanner)
func SecurityConfig() FeatureConfig {
	cfg := DefaultConfig()
	cfg.Technology.Enabled = true       // Useful for understanding attack surface
	cfg.Documentation.Enabled = false   // Not security-focused
	cfg.Tests.Enabled = false           // Not security-focused
	cfg.Ownership.Enabled = false       // Not security-focused
	return cfg
}

// FullConfig returns config with all features enabled
func FullConfig() FeatureConfig {
	return FeatureConfig{
		Technology: TechnologyConfig{
			Enabled:        true,
			ScanExtensions: true,
			ScanConfig:     true,
			ScanSBOM:       true,
		},
		Documentation: DocumentationConfig{
			Enabled:       true,
			CheckReadme:   true,
			CheckCodeDocs: true,
			CheckChangelog: true,
			CheckAPIDocs:  true,
		},
		Tests: TestsConfig{
			Enabled:           true,
			ParseReports:      true,
			AnalyzeInfra:      true,
			CoverageThreshold: 80,
		},
		Ownership: OwnershipConfig{
			Enabled:             true,
			AnalyzeContributors: true,
			CheckCodeowners:     true,
			DetectOrphans:       true,
			PeriodDays:          90,
		},
	}
}
