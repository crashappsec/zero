// Package codesecurity provides the consolidated code security super scanner
package codesecurity

// FeatureConfig holds configuration for all code security features
type FeatureConfig struct {
	Vulns   VulnsConfig   `json:"vulns"`
	Secrets SecretsConfig `json:"secrets"`
	API     APIConfig     `json:"api"`
}

// VulnsConfig configures code vulnerability scanning
type VulnsConfig struct {
	Enabled         bool     `json:"enabled"`
	IncludeOWASP    bool     `json:"include_owasp"`    // Include OWASP Top 10 rules
	IncludeCWE      bool     `json:"include_cwe"`      // Include CWE-mapped rules
	SeverityMinimum string   `json:"severity_minimum"` // low, medium, high, critical
	ExcludeRules    []string `json:"exclude_rules"`    // Rule IDs to skip
}

// SecretsConfig configures secret detection
type SecretsConfig struct {
	Enabled       bool `json:"enabled"`
	RedactSecrets bool `json:"redact_secrets"` // Redact secret values in output

	// Enhanced detection features
	EntropyAnalysis  EntropyConfig     `json:"entropy_analysis"`  // Entropy-based detection
	GitHistoryScan   GitHistoryConfig  `json:"git_history_scan"`  // Git history scanning
	AIAnalysis       AIAnalysisConfig  `json:"ai_analysis"`       // Claude-powered FP reduction
	RotationGuidance bool              `json:"rotation_guidance"` // Add rotation recommendations
}

// EntropyConfig configures entropy-based secret detection
type EntropyConfig struct {
	Enabled       bool    `json:"enabled"`
	MinLength     int     `json:"min_length"`     // Minimum string length to check (default: 16)
	HighThreshold float64 `json:"high_threshold"` // Entropy threshold for high confidence (default: 4.5)
	MedThreshold  float64 `json:"med_threshold"`  // Entropy threshold for medium confidence (default: 3.5)
}

// GitHistoryConfig configures git history secret scanning
type GitHistoryConfig struct {
	Enabled     bool   `json:"enabled"`
	MaxCommits  int    `json:"max_commits"`  // Maximum commits to scan (default: 1000)
	MaxAge      string `json:"max_age"`      // Maximum age to scan, e.g., "90d", "1y" (default: "1y")
	ScanRemoved bool   `json:"scan_removed"` // Track if secrets were later removed
}

// AIAnalysisConfig configures Claude-powered false positive reduction
type AIAnalysisConfig struct {
	Enabled             bool    `json:"enabled"`
	MaxFindings         int     `json:"max_findings"`         // Maximum findings to analyze (default: 50)
	ConfidenceThreshold float64 `json:"confidence_threshold"` // Threshold to mark as FP (default: 0.8)
}

// APIConfig configures API security scanning
type APIConfig struct {
	Enabled        bool `json:"enabled"`
	CheckAuth      bool `json:"check_auth"`      // Check authentication issues
	CheckInjection bool `json:"check_injection"` // Check injection vulnerabilities
	CheckSSRF      bool `json:"check_ssrf"`      // Check SSRF issues
	CheckCORS      bool `json:"check_cors"`      // Check CORS misconfig
	CheckOpenAPI   bool `json:"check_openapi"`   // Validate OpenAPI specs
	CheckGraphQL   bool `json:"check_graphql"`   // Check GraphQL security
	CheckOWASPAPI  bool `json:"check_owasp_api"` // Map to OWASP API Top 10
}

// DefaultConfig returns default feature configuration
func DefaultConfig() FeatureConfig {
	return FeatureConfig{
		Vulns: VulnsConfig{
			Enabled:         true,
			IncludeOWASP:    true,
			IncludeCWE:      true,
			SeverityMinimum: "low",
		},
		Secrets: SecretsConfig{
			Enabled:       true,
			RedactSecrets: true,
			EntropyAnalysis: EntropyConfig{
				Enabled:       true, // Enabled by default - fast and no dependencies
				MinLength:     16,
				HighThreshold: 4.5,
				MedThreshold:  3.5,
			},
			GitHistoryScan: GitHistoryConfig{
				Enabled:     false, // Disabled by default - can be slow on large repos
				MaxCommits:  1000,
				MaxAge:      "1y",
				ScanRemoved: true,
			},
			AIAnalysis: AIAnalysisConfig{
				Enabled:             false, // Disabled by default - requires ANTHROPIC_API_KEY
				MaxFindings:         50,
				ConfidenceThreshold: 0.8,
			},
			RotationGuidance: true, // Enabled by default - no dependencies, high value
		},
		API: APIConfig{
			Enabled:        true,
			CheckAuth:      true,
			CheckInjection: true,
			CheckSSRF:      true,
			CheckCORS:      true,
			CheckOpenAPI:   true,
			CheckGraphQL:   true,
			CheckOWASPAPI:  true,
		},
	}
}

// QuickConfig returns minimal config for fast scans
func QuickConfig() FeatureConfig {
	cfg := DefaultConfig()
	cfg.API.Enabled = false                    // Skip separate API scan for speed
	cfg.Secrets.EntropyAnalysis.Enabled = true // Keep entropy - it's fast
	cfg.Secrets.GitHistoryScan.Enabled = false
	cfg.Secrets.AIAnalysis.Enabled = false
	cfg.Secrets.RotationGuidance = false // Skip for speed
	return cfg
}

// SecurityConfig returns config optimized for security-focused scans
func SecurityConfig() FeatureConfig {
	cfg := DefaultConfig()
	cfg.Secrets.EntropyAnalysis.Enabled = true
	cfg.Secrets.GitHistoryScan.Enabled = true // Enable history scanning for security
	cfg.Secrets.GitHistoryScan.MaxCommits = 2000
	cfg.Secrets.GitHistoryScan.MaxAge = "2y"
	cfg.Secrets.AIAnalysis.Enabled = true // Enable AI analysis if API key available
	cfg.Secrets.RotationGuidance = true
	return cfg
}

// FullConfig returns config with all features enabled
func FullConfig() FeatureConfig {
	cfg := DefaultConfig()
	cfg.Secrets.EntropyAnalysis.Enabled = true
	cfg.Secrets.GitHistoryScan.Enabled = true
	cfg.Secrets.GitHistoryScan.MaxCommits = 5000
	cfg.Secrets.GitHistoryScan.MaxAge = "5y"
	cfg.Secrets.AIAnalysis.Enabled = true
	cfg.Secrets.RotationGuidance = true
	return cfg
}
