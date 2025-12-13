// Package code provides the consolidated code security super scanner
package code

// FeatureConfig holds configuration for all code analysis features
type FeatureConfig struct {
	Vulns    VulnsConfig    `json:"vulns"`
	Secrets  SecretsConfig  `json:"secrets"`
	API      APIConfig      `json:"api"`
	TechDebt TechDebtConfig `json:"tech_debt"`
}

// VulnsConfig configures code vulnerability scanning
type VulnsConfig struct {
	Enabled         bool     `json:"enabled"`
	IncludeOWASP    bool     `json:"include_owasp"`     // Include OWASP Top 10 rules
	IncludeCWE      bool     `json:"include_cwe"`       // Include CWE-mapped rules
	SeverityMinimum string   `json:"severity_minimum"`  // low, medium, high, critical
	ExcludeRules    []string `json:"exclude_rules"`     // Rule IDs to skip
}

// SecretsConfig configures secret detection
type SecretsConfig struct {
	Enabled       bool `json:"enabled"`
	RedactSecrets bool `json:"redact_secrets"` // Redact secret values in output
}

// APIConfig configures API security scanning
type APIConfig struct {
	Enabled       bool `json:"enabled"`
	CheckAuth     bool `json:"check_auth"`     // Check authentication issues
	CheckInjection bool `json:"check_injection"` // Check injection vulnerabilities
	CheckSSRF     bool `json:"check_ssrf"`     // Check SSRF issues
	CheckCORS     bool `json:"check_cors"`     // Check CORS misconfig
}

// TechDebtConfig configures technical debt detection
type TechDebtConfig struct {
	Enabled           bool `json:"enabled"`
	IncludeMarkers    bool `json:"include_markers"`    // TODOs, FIXMEs, etc.
	IncludeIssues     bool `json:"include_issues"`     // Code smells
	IncludeComplexity bool `json:"include_complexity"` // Complexity metrics (requires Semgrep)
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
		},
		API: APIConfig{
			Enabled:       true,
			CheckAuth:     true,
			CheckInjection: true,
			CheckSSRF:     true,
			CheckCORS:     true,
		},
		TechDebt: TechDebtConfig{
			Enabled:           true,
			IncludeMarkers:    true,
			IncludeIssues:     true,
			IncludeComplexity: true,
		},
	}
}

// QuickConfig returns minimal config for fast scans
func QuickConfig() FeatureConfig {
	cfg := DefaultConfig()
	cfg.TechDebt.IncludeComplexity = false // Skip semgrep complexity analysis
	cfg.API.Enabled = false                // Skip separate API scan
	return cfg
}

// SecurityConfig returns security-focused config
func SecurityConfig() FeatureConfig {
	cfg := DefaultConfig()
	cfg.TechDebt.Enabled = false // Skip tech debt
	return cfg
}

// FullConfig returns config with all features enabled
func FullConfig() FeatureConfig {
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
		},
		API: APIConfig{
			Enabled:       true,
			CheckAuth:     true,
			CheckInjection: true,
			CheckSSRF:     true,
			CheckCORS:     true,
		},
		TechDebt: TechDebtConfig{
			Enabled:           true,
			IncludeMarkers:    true,
			IncludeIssues:     true,
			IncludeComplexity: true,
		},
	}
}
