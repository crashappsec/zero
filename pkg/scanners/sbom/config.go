// Package sbom provides the SBOM (Software Bill of Materials) super scanner
package sbom

// FeatureConfig holds configuration for SBOM features
type FeatureConfig struct {
	Generation GenerationConfig `json:"generation"`
	Integrity  IntegrityConfig  `json:"integrity"`
}

// GenerationConfig configures SBOM generation
type GenerationConfig struct {
	Enabled        bool   `json:"enabled"`
	Tool           string `json:"tool"`             // cdxgen, syft, auto
	SpecVersion    string `json:"spec_version"`     // CycloneDX version (1.4, 1.5, 1.6)
	Format         string `json:"format"`           // json, xml
	FallbackToSyft bool   `json:"fallback_to_syft"` // Use syft if cdxgen fails
	IncludeDev     bool   `json:"include_dev"`      // Include dev dependencies
	Deep           bool   `json:"deep"`             // Deep analysis mode
}

// IntegrityConfig configures SBOM integrity verification
type IntegrityConfig struct {
	Enabled           bool `json:"enabled"`
	VerifyLockfiles   bool `json:"verify_lockfiles"`   // Compare SBOM against lockfiles
	DetectDrift       bool `json:"detect_drift"`       // Detect drift from previous SBOM
	CheckCompleteness bool `json:"check_completeness"` // Verify all deps are captured
}

// DefaultConfig returns default feature configuration
func DefaultConfig() FeatureConfig {
	return FeatureConfig{
		Generation: GenerationConfig{
			Enabled:        true,
			Tool:           "auto",
			SpecVersion:    "1.5",
			Format:         "json",
			FallbackToSyft: true,
			IncludeDev:     false,
			Deep:           false,
		},
		Integrity: IntegrityConfig{
			Enabled:           true,
			VerifyLockfiles:   true,
			DetectDrift:       false,
			CheckCompleteness: true,
		},
	}
}

// QuickConfig returns minimal config for fast scans
func QuickConfig() FeatureConfig {
	cfg := DefaultConfig()
	cfg.Integrity.Enabled = false
	cfg.Generation.Deep = false
	return cfg
}

// FullConfig returns config with all features enabled
func FullConfig() FeatureConfig {
	return FeatureConfig{
		Generation: GenerationConfig{
			Enabled:        true,
			Tool:           "auto",
			SpecVersion:    "1.5",
			Format:         "json",
			FallbackToSyft: true,
			IncludeDev:     false,
			Deep:           true,
		},
		Integrity: IntegrityConfig{
			Enabled:           true,
			VerifyLockfiles:   true,
			DetectDrift:       true,
			CheckCompleteness: true,
		},
	}
}
