// Package codeownership provides code ownership and CODEOWNERS analysis
package codeownership

// FeatureConfig holds configuration for code ownership analysis
type FeatureConfig struct {
	Enabled             bool `json:"enabled"`
	AnalyzeContributors bool `json:"analyze_contributors"` // Analyze git contributors
	CheckCodeowners     bool `json:"check_codeowners"`     // Validate CODEOWNERS file
	DetectOrphans       bool `json:"detect_orphans"`       // Find files with no recent commits
	PeriodDays          int  `json:"period_days"`          // Analysis period (default 90)
}

// DefaultConfig returns default feature configuration
func DefaultConfig() FeatureConfig {
	return FeatureConfig{
		Enabled:             true,
		AnalyzeContributors: true,
		CheckCodeowners:     true,
		DetectOrphans:       true,
		PeriodDays:          90,
	}
}

// QuickConfig returns minimal config for fast scans
func QuickConfig() FeatureConfig {
	cfg := DefaultConfig()
	cfg.AnalyzeContributors = false // Skip contributor analysis (slow)
	cfg.DetectOrphans = false       // Skip orphan detection (slow)
	return cfg
}

// FullConfig returns config with all features enabled
func FullConfig() FeatureConfig {
	return FeatureConfig{
		Enabled:             true,
		AnalyzeContributors: true,
		CheckCodeowners:     true,
		DetectOrphans:       true,
		PeriodDays:          180, // Longer period for thorough analysis
	}
}
