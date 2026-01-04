// Package codesecurity provides the consolidated code security super scanner
package codesecurity

// FeatureConfig holds configuration for all code security features
type FeatureConfig struct {
	Vulns   VulnsConfig   `json:"vulns"`
	Secrets SecretsConfig `json:"secrets"`
	API     APIConfig     `json:"api"`
	// Crypto features (merged from code-crypto)
	Ciphers      CiphersConfig      `json:"ciphers"`
	Keys         KeysConfig         `json:"keys"`
	Random       RandomConfig       `json:"random"`
	TLS          TLSConfig          `json:"tls"`
	Certificates CertificatesConfig `json:"certificates"`
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
	EntropyAnalysis     EntropyConfig            `json:"entropy_analysis"`      // Entropy-based detection
	GitHistoryScan      GitHistoryConfig         `json:"git_history_scan"`      // Git history scanning
	GitHistorySecurity  GitHistorySecurityConfig `json:"git_history_security"`  // Git history security (gitignore violations, sensitive files)
	AIAnalysis          AIAnalysisConfig         `json:"ai_analysis"`           // Claude-powered FP reduction
	RotationGuidance    bool                     `json:"rotation_guidance"`     // Add rotation recommendations
	IaCSecrets          IaCSecretsConfig         `json:"iac_secrets"`           // IaC-specific secrets detection
}

// IaCSecretsConfig configures IaC-specific secrets detection
type IaCSecretsConfig struct {
	Enabled bool `json:"enabled"` // Scan IaC files for hardcoded secrets
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

// GitHistorySecurityConfig configures git history security scanning
type GitHistorySecurityConfig struct {
	Enabled              bool   `json:"enabled"`
	MaxCommits           int    `json:"max_commits"`             // Maximum commits to scan (default: 1000)
	MaxAge               string `json:"max_age"`                 // Maximum age to scan, e.g., "90d", "1y" (default: "1y")
	ScanGitignoreHistory bool   `json:"scan_gitignore_history"`  // Scan for gitignore violations in history
	ScanSensitiveFiles   bool   `json:"scan_sensitive_files"`    // Scan for sensitive file patterns
	GeneratePurgeReport  bool   `json:"generate_purge_report"`   // Generate purge recommendations
}

// AIAnalysisConfig configures Claude-powered false positive reduction
type AIAnalysisConfig struct {
	Enabled             bool    `json:"enabled"`
	MaxFindings         int     `json:"max_findings"`         // Maximum findings to analyze (default: 50)
	ConfidenceThreshold float64 `json:"confidence_threshold"` // Threshold to mark as FP (default: 0.8)
}

// APIConfig configures API scanning (security and quality)
type APIConfig struct {
	Enabled        bool `json:"enabled"`
	CheckAuth      bool `json:"check_auth"`      // Check authentication issues
	CheckInjection bool `json:"check_injection"` // Check injection vulnerabilities
	CheckSSRF      bool `json:"check_ssrf"`      // Check SSRF issues
	CheckCORS      bool `json:"check_cors"`      // Check CORS misconfig
	CheckOpenAPI   bool `json:"check_openapi"`   // Validate OpenAPI specs
	CheckGraphQL   bool `json:"check_graphql"`   // Check GraphQL security
	CheckOWASPAPI  bool `json:"check_owasp_api"` // Map to OWASP API Top 10

	// Non-security API quality checks
	CheckDesign       bool `json:"check_design"`       // REST design patterns, naming conventions
	CheckPerformance  bool `json:"check_performance"`  // N+1 queries, pagination, caching
	CheckObservability bool `json:"check_observability"` // Logging, error handling, metrics
	CheckDocumentation bool `json:"check_documentation"` // API documentation completeness
}

// CiphersConfig configures weak cipher detection
type CiphersConfig struct {
	Enabled     bool `json:"enabled"`
	UseSemgrep  bool `json:"use_semgrep"`  // Use Semgrep for AST-based detection
	UsePatterns bool `json:"use_patterns"` // Use regex pattern matching
}

// KeysConfig configures hardcoded key detection
type KeysConfig struct {
	Enabled       bool `json:"enabled"`
	CheckAPIKeys  bool `json:"check_api_keys"`
	CheckPrivate  bool `json:"check_private_keys"`
	CheckAWS      bool `json:"check_aws_keys"`
	CheckSigning  bool `json:"check_signing_keys"`
	RedactMatches bool `json:"redact_matches"` // Redact sensitive values in output
}

// RandomConfig configures insecure random detection
type RandomConfig struct {
	Enabled bool `json:"enabled"`
}

// TLSConfig configures TLS misconfiguration detection
type TLSConfig struct {
	Enabled           bool `json:"enabled"`
	CheckProtocols    bool `json:"check_protocols"`    // Check for deprecated SSL/TLS versions
	CheckVerification bool `json:"check_verification"` // Check for disabled cert verification
	CheckCipherSuites bool `json:"check_cipher_suites"` // Check for weak cipher suites
	CheckInsecureURLs bool `json:"check_insecure_urls"` // Check for HTTP URLs
}

// CertificatesConfig configures X.509 certificate analysis
type CertificatesConfig struct {
	Enabled             bool `json:"enabled"`
	ExpiryWarningDays   int  `json:"expiry_warning_days"` // Warn if expiring within N days
	CheckKeyStrength    bool `json:"check_key_strength"`
	CheckSignatureAlgo  bool `json:"check_signature_algo"`
	CheckSelfSigned     bool `json:"check_self_signed"`
	CheckValidityPeriod bool `json:"check_validity_period"`
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
			GitHistorySecurity: GitHistorySecurityConfig{
				Enabled:              false, // Disabled by default - can be slow on large repos
				MaxCommits:           1000,
				MaxAge:               "1y",
				ScanGitignoreHistory: true,
				ScanSensitiveFiles:   true,
				GeneratePurgeReport:  true,
			},
			AIAnalysis: AIAnalysisConfig{
				Enabled:             false, // Disabled by default - requires ANTHROPIC_API_KEY
				MaxFindings:         50,
				ConfidenceThreshold: 0.8,
			},
			RotationGuidance: true, // Enabled by default - no dependencies, high value
			IaCSecrets: IaCSecretsConfig{
				Enabled: true, // Enabled by default - catches secrets in Terraform, K8s, etc.
			},
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
			// Non-security quality checks
			CheckDesign:        true,
			CheckPerformance:   true,
			CheckObservability: true,
			CheckDocumentation: true,
		},
		// Crypto features
		Ciphers: CiphersConfig{
			Enabled:     true,
			UseSemgrep:  true,
			UsePatterns: true,
		},
		Keys: KeysConfig{
			Enabled:       true,
			CheckAPIKeys:  true,
			CheckPrivate:  true,
			CheckAWS:      true,
			CheckSigning:  true,
			RedactMatches: true,
		},
		Random: RandomConfig{
			Enabled: true,
		},
		TLS: TLSConfig{
			Enabled:           true,
			CheckProtocols:    true,
			CheckVerification: true,
			CheckCipherSuites: true,
			CheckInsecureURLs: false, // Noisy, off by default
		},
		Certificates: CertificatesConfig{
			Enabled:             true,
			ExpiryWarningDays:   90,
			CheckKeyStrength:    true,
			CheckSignatureAlgo:  true,
			CheckSelfSigned:     true,
			CheckValidityPeriod: true,
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
	// Crypto: faster without semgrep, skip certs
	cfg.Ciphers.UseSemgrep = false
	cfg.Certificates.Enabled = false
	cfg.TLS.CheckInsecureURLs = false
	return cfg
}

// SecurityConfig returns config optimized for security-focused scans
func SecurityConfig() FeatureConfig {
	cfg := DefaultConfig()
	cfg.Secrets.EntropyAnalysis.Enabled = true
	cfg.Secrets.GitHistoryScan.Enabled = true // Enable history scanning for security
	cfg.Secrets.GitHistoryScan.MaxCommits = 2000
	cfg.Secrets.GitHistoryScan.MaxAge = "2y"
	cfg.Secrets.GitHistorySecurity.Enabled = true // Enable git history security scanning
	cfg.Secrets.GitHistorySecurity.MaxCommits = 2000
	cfg.Secrets.GitHistorySecurity.MaxAge = "2y"
	cfg.Secrets.AIAnalysis.Enabled = true // Enable AI analysis if API key available
	cfg.Secrets.RotationGuidance = true
	// Crypto: enable all checks for security
	cfg.TLS.CheckInsecureURLs = true
	return cfg
}

// FullConfig returns config with all features enabled
func FullConfig() FeatureConfig {
	cfg := DefaultConfig()
	cfg.Secrets.EntropyAnalysis.Enabled = true
	cfg.Secrets.GitHistoryScan.Enabled = true
	cfg.Secrets.GitHistoryScan.MaxCommits = 5000
	cfg.Secrets.GitHistoryScan.MaxAge = "5y"
	cfg.Secrets.GitHistorySecurity.Enabled = true
	cfg.Secrets.GitHistorySecurity.MaxCommits = 5000
	cfg.Secrets.GitHistorySecurity.MaxAge = "5y"
	cfg.Secrets.AIAnalysis.Enabled = true
	cfg.Secrets.RotationGuidance = true
	// Crypto: all features enabled
	cfg.Ciphers.Enabled = true
	cfg.Ciphers.UseSemgrep = true
	cfg.Ciphers.UsePatterns = true
	cfg.Keys.Enabled = true
	cfg.Random.Enabled = true
	cfg.TLS.Enabled = true
	cfg.TLS.CheckInsecureURLs = true
	cfg.Certificates.Enabled = true
	return cfg
}
