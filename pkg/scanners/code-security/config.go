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
	cfg.API.Enabled = false // Skip separate API scan for speed
	return cfg
}

// FullConfig returns config with all features enabled
func FullConfig() FeatureConfig {
	return DefaultConfig()
}
