// Package codepackages provides the consolidated code packages scanner
// This scanner generates SBOMs and performs comprehensive package analysis.
package codepackages

// FeatureConfig holds configuration for all supply chain features
type FeatureConfig struct {
	// SBOM features (from sbom scanner)
	Generation GenerationConfig `json:"generation"`
	Integrity  IntegrityConfig  `json:"integrity"`
	// Package analysis features
	Vulns           VulnsConfig           `json:"vulns"`
	Health          HealthConfig          `json:"health"`
	Malcontent      MalcontentConfig      `json:"malcontent"`
	Provenance      ProvenanceConfig      `json:"provenance"`
	Bundle          BundleConfig          `json:"bundle"`
	Recommendations RecommendationsConfig `json:"recommendations"`
	Confusion       ConfusionConfig       `json:"confusion"`
	Reachability    ReachabilityConfig    `json:"reachability"`
	Licenses        LicensesConfig        `json:"licenses"`
	Typosquats      TyposquatsConfig      `json:"typosquats"`
	Deprecations    DeprecationsConfig    `json:"deprecations"`
	Duplicates      DuplicatesConfig      `json:"duplicates"`
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
	Enabled           bool   `json:"enabled"`
	VerifyLockfiles   bool   `json:"verify_lockfiles"`   // Compare SBOM against lockfiles
	DetectDrift       bool   `json:"detect_drift"`       // Detect drift from previous SBOM
	CheckCompleteness bool   `json:"check_completeness"` // Verify all deps are captured
	PreviousSBOMPath  string `json:"previous_sbom_path"` // Path to previous SBOM for diff
}

// VulnsConfig configures vulnerability scanning
type VulnsConfig struct {
	Enabled           bool     `json:"enabled"`
	SeverityThreshold string   `json:"severity_threshold"` // critical, high, medium, low
	IncludeKEV        bool     `json:"include_kev"`        // Check CISA KEV catalog
	IgnoreIDs         []string `json:"ignore_ids"`         // CVE IDs to ignore
}

// HealthConfig configures package health checking
type HealthConfig struct {
	Enabled        bool    `json:"enabled"`
	MinScore       float64 `json:"min_score"`       // Minimum health score (0-10)
	MaxPackages    int     `json:"max_packages"`    // Limit API calls
	CheckScorecard bool    `json:"check_scorecard"` // Include OpenSSF Scorecard
}

// MalcontentConfig configures supply chain compromise detection
type MalcontentConfig struct {
	Enabled         bool   `json:"enabled"`
	MinRiskLevel    string `json:"min_risk_level"` // critical, high, medium, low
	ScanNodeModules bool   `json:"scan_node_modules"`
}

// ProvenanceConfig configures provenance verification
type ProvenanceConfig struct {
	Enabled           bool `json:"enabled"`
	RequireSignatures bool `json:"require_signatures"`
	RequireSLSA       bool `json:"require_slsa"`
}

// BundleConfig configures bundle size optimization
type BundleConfig struct {
	Enabled         bool `json:"enabled"`
	SizeThresholdKB int  `json:"size_threshold_kb"` // Alert if package > this size
	CheckDuplicates bool `json:"check_duplicates"`
	CheckTreeshake  bool `json:"check_treeshake"`
}

// RecommendationsConfig configures package alternatives
type RecommendationsConfig struct {
	Enabled            bool `json:"enabled"`
	IncludeSecurity    bool `json:"include_security"`
	IncludeHealth      bool `json:"include_health"`
	IncludePerformance bool `json:"include_performance"`
}

// ConfusionConfig configures dependency confusion detection
type ConfusionConfig struct {
	Enabled   bool `json:"enabled"`
	CheckNPM  bool `json:"check_npm"`
	CheckPyPI bool `json:"check_pypi"`
	CheckGo   bool `json:"check_go"`
}

// ReachabilityConfig configures vulnerability reachability analysis
type ReachabilityConfig struct {
	Enabled      bool `json:"enabled"`
	Experimental bool `json:"experimental"` // Warning: uses experimental osv-scanner feature
}

// LicensesConfig configures license compliance
type LicensesConfig struct {
	Enabled       bool     `json:"enabled"`
	AllowedList   []string `json:"allowed"`   // Explicitly allowed licenses
	BlockedList   []string `json:"blocked"`   // Explicitly blocked licenses
	FailOnUnknown bool     `json:"fail_on_unknown"`
}

// TyposquatsConfig configures typosquatting detection
type TyposquatsConfig struct {
	Enabled           bool `json:"enabled"`
	CheckSimilarNames bool `json:"check_similar_names"`
	CheckNewPackages  bool `json:"check_new_packages"` // Flag packages < 30 days old
}

// DeprecationsConfig configures deprecated package detection
type DeprecationsConfig struct {
	Enabled       bool `json:"enabled"`
	CheckNPM      bool `json:"check_npm"`
	CheckPyPI     bool `json:"check_pypi"`
	CheckGo       bool `json:"check_go"`
}

// DuplicatesConfig configures duplicate dependency detection
type DuplicatesConfig struct {
	Enabled            bool `json:"enabled"`
	CheckVersions      bool `json:"check_versions"`       // Multiple versions of same package
	CheckFunctionality bool `json:"check_functionality"`  // Different packages with same purpose
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
		Vulns: VulnsConfig{
			Enabled:           true,
			SeverityThreshold: "low",
			IncludeKEV:        true,
		},
		Health: HealthConfig{
			Enabled:        true,
			MinScore:       0,
			MaxPackages:    50,
			CheckScorecard: true,
		},
		Malcontent: MalcontentConfig{
			Enabled:         true,
			MinRiskLevel:    "medium",
			ScanNodeModules: false,
		},
		Provenance: ProvenanceConfig{
			Enabled:           false, // Off by default - slow
			RequireSignatures: false,
			RequireSLSA:       false,
		},
		Bundle: BundleConfig{
			Enabled:         false, // Off by default - npm specific
			SizeThresholdKB: 500,
			CheckDuplicates: true,
			CheckTreeshake:  true,
		},
		Recommendations: RecommendationsConfig{
			Enabled:            false, // Off by default
			IncludeSecurity:    true,
			IncludeHealth:      true,
			IncludePerformance: false,
		},
		Confusion: ConfusionConfig{
			Enabled:  true,
			CheckNPM: true,
			CheckPyPI: true,
			CheckGo:  false,
		},
		Reachability: ReachabilityConfig{
			Enabled:      false, // Off by default - experimental
			Experimental: true,
		},
		Licenses: LicensesConfig{
			Enabled:       true,
			AllowedList:   []string{},
			BlockedList:   []string{"GPL-3.0", "AGPL-3.0"},
			FailOnUnknown: false,
		},
		Typosquats: TyposquatsConfig{
			Enabled:           true,
			CheckSimilarNames: true,
			CheckNewPackages:  true,
		},
		Deprecations: DeprecationsConfig{
			Enabled:  true,
			CheckNPM: true,
			CheckPyPI: true,
			CheckGo:  true,
		},
		Duplicates: DuplicatesConfig{
			Enabled:            true,
			CheckVersions:      true,
			CheckFunctionality: false,
		},
	}
}

// QuickConfig returns minimal config for fast scans
func QuickConfig() FeatureConfig {
	cfg := DefaultConfig()
	cfg.Vulns.Enabled = true
	cfg.Licenses.Enabled = true
	cfg.Health.Enabled = false
	cfg.Malcontent.Enabled = false
	cfg.Provenance.Enabled = false
	cfg.Bundle.Enabled = false
	cfg.Recommendations.Enabled = false
	cfg.Confusion.Enabled = false
	cfg.Reachability.Enabled = false
	cfg.Typosquats.Enabled = false
	cfg.Deprecations.Enabled = false
	cfg.Duplicates.Enabled = false
	return cfg
}

// SecurityConfig returns security-focused config
func SecurityConfig() FeatureConfig {
	cfg := DefaultConfig()
	cfg.Vulns.Enabled = true
	cfg.Vulns.IncludeKEV = true
	cfg.Malcontent.Enabled = true
	cfg.Confusion.Enabled = true
	cfg.Reachability.Enabled = true
	cfg.Typosquats.Enabled = true
	cfg.Bundle.Enabled = false
	cfg.Recommendations.Enabled = false
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
			IncludeDev:     true, // Full config includes dev deps
			Deep:           true, // Full config uses deep analysis
		},
		Integrity: IntegrityConfig{
			Enabled:           true,
			VerifyLockfiles:   true,
			DetectDrift:       true,
			CheckCompleteness: true,
		},
		Vulns: VulnsConfig{
			Enabled:           true,
			SeverityThreshold: "low",
			IncludeKEV:        true,
		},
		Health: HealthConfig{
			Enabled:        true,
			MinScore:       0,
			MaxPackages:    100,
			CheckScorecard: true,
		},
		Malcontent: MalcontentConfig{
			Enabled:         true,
			MinRiskLevel:    "low",
			ScanNodeModules: true,
		},
		Provenance: ProvenanceConfig{
			Enabled:           true,
			RequireSignatures: false,
			RequireSLSA:       false,
		},
		Bundle: BundleConfig{
			Enabled:         true,
			SizeThresholdKB: 500,
			CheckDuplicates: true,
			CheckTreeshake:  true,
		},
		Recommendations: RecommendationsConfig{
			Enabled:            true,
			IncludeSecurity:    true,
			IncludeHealth:      true,
			IncludePerformance: true,
		},
		Confusion: ConfusionConfig{
			Enabled:  true,
			CheckNPM: true,
			CheckPyPI: true,
			CheckGo:  true,
		},
		Reachability: ReachabilityConfig{
			Enabled:      true,
			Experimental: true,
		},
		Licenses: LicensesConfig{
			Enabled:       true,
			AllowedList:   []string{},
			BlockedList:   []string{"GPL-3.0", "AGPL-3.0", "SSPL-1.0"},
			FailOnUnknown: false,
		},
		Typosquats: TyposquatsConfig{
			Enabled:           true,
			CheckSimilarNames: true,
			CheckNewPackages:  true,
		},
		Deprecations: DeprecationsConfig{
			Enabled:  true,
			CheckNPM: true,
			CheckPyPI: true,
			CheckGo:  true,
		},
		Duplicates: DuplicatesConfig{
			Enabled:            true,
			CheckVersions:      true,
			CheckFunctionality: true,
		},
	}
}
