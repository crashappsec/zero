// Package codecrypto provides the consolidated cryptographic security super scanner
package codecrypto

// FeatureConfig holds configuration for all crypto analysis features
type FeatureConfig struct {
	Ciphers      CiphersConfig      `json:"ciphers"`
	Keys         KeysConfig         `json:"keys"`
	Random       RandomConfig       `json:"random"`
	TLS          TLSConfig          `json:"tls"`
	Certificates CertificatesConfig `json:"certificates"`
}

// CiphersConfig configures weak cipher detection
type CiphersConfig struct {
	Enabled     bool `json:"enabled"`
	UseSemgrep  bool `json:"use_semgrep"`   // Use Semgrep for AST-based detection
	UsePatterns bool `json:"use_patterns"`  // Use regex pattern matching
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
	Enabled             bool `json:"enabled"`
	CheckProtocols      bool `json:"check_protocols"`       // Check for deprecated SSL/TLS versions
	CheckVerification   bool `json:"check_verification"`    // Check for disabled cert verification
	CheckCipherSuites   bool `json:"check_cipher_suites"`   // Check for weak cipher suites
	CheckInsecureURLs   bool `json:"check_insecure_urls"`   // Check for HTTP URLs
}

// CertificatesConfig configures X.509 certificate analysis
type CertificatesConfig struct {
	Enabled             bool `json:"enabled"`
	ExpiryWarningDays   int  `json:"expiry_warning_days"`   // Warn if expiring within N days
	CheckKeyStrength    bool `json:"check_key_strength"`
	CheckSignatureAlgo  bool `json:"check_signature_algo"`
	CheckSelfSigned     bool `json:"check_self_signed"`
	CheckValidityPeriod bool `json:"check_validity_period"`
}

// DefaultConfig returns default feature configuration
func DefaultConfig() FeatureConfig {
	return FeatureConfig{
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
	cfg.Ciphers.UseSemgrep = false // Faster without semgrep
	cfg.Certificates.Enabled = false
	cfg.TLS.CheckInsecureURLs = false
	return cfg
}

// SecurityConfig returns security-focused config
func SecurityConfig() FeatureConfig {
	cfg := DefaultConfig()
	cfg.TLS.CheckInsecureURLs = true // Enable all checks
	return cfg
}

// FullConfig returns config with all features enabled
func FullConfig() FeatureConfig {
	return FeatureConfig{
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
			CheckInsecureURLs: true,
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
