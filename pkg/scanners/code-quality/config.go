// Package codequality provides the consolidated code quality super scanner
package codequality

// FeatureConfig holds configuration for all code quality features
type FeatureConfig struct {
	TechDebt     TechDebtConfig     `json:"tech_debt"`
	Complexity   ComplexityConfig   `json:"complexity"`
	TestCoverage TestCoverageConfig `json:"test_coverage"`
	CodeDocs     CodeDocsConfig     `json:"code_docs"`
}

// TechDebtConfig configures technical debt detection
type TechDebtConfig struct {
	Enabled        bool     `json:"enabled"`
	IncludeMarkers bool     `json:"include_markers"` // TODOs, FIXMEs, etc.
	IncludeIssues  bool     `json:"include_issues"`  // Code smells
	MarkerTypes    []string `json:"marker_types"`    // Types to detect: TODO, FIXME, HACK, etc.
}

// ComplexityConfig configures complexity analysis
type ComplexityConfig struct {
	Enabled          bool `json:"enabled"`
	CheckCyclomatic  bool `json:"check_cyclomatic"`  // Cyclomatic complexity
	CheckCognitive   bool `json:"check_cognitive"`   // Cognitive complexity
	CheckNesting     bool `json:"check_nesting"`     // Nesting depth
	MaxFunctionLines int  `json:"max_function_lines"` // Max lines per function
	MaxCyclomatic    int  `json:"max_cyclomatic"`    // Max cyclomatic complexity
}

// TestCoverageConfig configures test coverage analysis
type TestCoverageConfig struct {
	Enabled          bool `json:"enabled"`
	ParseReports     bool `json:"parse_reports"`     // Parse existing coverage reports
	MinimumThreshold int  `json:"minimum_threshold"` // Minimum coverage percentage
}

// CodeDocsConfig configures documentation analysis
type CodeDocsConfig struct {
	Enabled        bool `json:"enabled"`
	CheckPublicAPI bool `json:"check_public_api"` // Check public API documentation
	CheckReadme    bool `json:"check_readme"`     // Check README quality
	CheckChangelog bool `json:"check_changelog"`  // Check for CHANGELOG
	CheckApiDocs   bool `json:"check_api_docs"`   // Check for API documentation
}

// DefaultConfig returns default feature configuration
func DefaultConfig() FeatureConfig {
	return FeatureConfig{
		TechDebt: TechDebtConfig{
			Enabled:        true,
			IncludeMarkers: true,
			IncludeIssues:  true,
			MarkerTypes:    []string{"TODO", "FIXME", "HACK", "XXX", "BUG", "WORKAROUND"},
		},
		Complexity: ComplexityConfig{
			Enabled:          true,
			CheckCyclomatic:  true,
			CheckCognitive:   true,
			CheckNesting:     true,
			MaxFunctionLines: 50,
			MaxCyclomatic:    10,
		},
		TestCoverage: TestCoverageConfig{
			Enabled:          true,
			ParseReports:     true,
			MinimumThreshold: 80,
		},
		CodeDocs: CodeDocsConfig{
			Enabled:        true,
			CheckPublicAPI: true,
			CheckReadme:    true,
			CheckChangelog: true,
			CheckApiDocs:   true,
		},
	}
}

// QuickConfig returns minimal config for fast scans
func QuickConfig() FeatureConfig {
	cfg := DefaultConfig()
	cfg.Complexity.Enabled = false // Skip semgrep complexity analysis
	cfg.TestCoverage.ParseReports = false
	return cfg
}

// FullConfig returns config with all features enabled
func FullConfig() FeatureConfig {
	return DefaultConfig()
}
