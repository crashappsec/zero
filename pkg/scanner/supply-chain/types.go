// Package supplychain provides the consolidated supply chain security scanner
// This scanner generates SBOMs and performs comprehensive package analysis.
package supplychain

import "encoding/json"

// Result holds all feature results
type Result struct {
	FeaturesRun []string `json:"features_run"`
	Summary     Summary  `json:"summary"`
	Findings    Findings `json:"findings"`
}

// Summary holds summaries from all features
type Summary struct {
	// SBOM features
	Generation *GenerationSummary `json:"generation,omitempty"`
	Integrity  *IntegritySummary  `json:"integrity,omitempty"`
	// Package analysis features
	Vulns           *VulnsSummary           `json:"vulns,omitempty"`
	Health          *HealthSummary          `json:"health,omitempty"`
	Licenses        *LicensesSummary        `json:"licenses,omitempty"`
	Malcontent      *MalcontentSummary      `json:"malcontent,omitempty"`
	Confusion       *ConfusionSummary       `json:"confusion,omitempty"`
	Reachability    *ReachabilitySummary    `json:"reachability,omitempty"`
	Provenance      *ProvenanceSummary      `json:"provenance,omitempty"`
	Bundle          *BundleSummary          `json:"bundle,omitempty"`
	Recommendations *RecommendationsSummary `json:"recommendations,omitempty"`
	Typosquats      *TyposquatsSummary      `json:"typosquats,omitempty"`
	Deprecations    *DeprecationsSummary    `json:"deprecations,omitempty"`
	Duplicates      *DuplicatesSummary      `json:"duplicates,omitempty"`
	Errors          []string                `json:"errors,omitempty"`
}

// Findings holds findings from all features
type Findings struct {
	// SBOM features
	Generation *GenerationFindings `json:"generation,omitempty"`
	Integrity  *IntegrityFindings  `json:"integrity,omitempty"`
	// Package analysis features
	Vulns           interface{} `json:"vulns,omitempty"`
	Health          interface{} `json:"health,omitempty"`
	Licenses        interface{} `json:"licenses,omitempty"`
	Malcontent      interface{} `json:"malcontent,omitempty"`
	Confusion       interface{} `json:"confusion,omitempty"`
	Reachability    interface{} `json:"reachability,omitempty"`
	Provenance      interface{} `json:"provenance,omitempty"`
	Bundle          interface{} `json:"bundle,omitempty"`
	Recommendations interface{} `json:"recommendations,omitempty"`
	Typosquats      interface{} `json:"typosquats,omitempty"`
	Deprecations    interface{} `json:"deprecations,omitempty"`
	Duplicates      interface{} `json:"duplicates,omitempty"`
}

// ComponentData is a simplified view of SBOM component for package analysis
// This is populated from the sbom scanner output
type ComponentData struct {
	Name      string   `json:"name"`
	Version   string   `json:"version"`
	Purl      string   `json:"purl"`
	Ecosystem string   `json:"ecosystem"`
	Licenses  []string `json:"licenses,omitempty"`
	Scope     string   `json:"scope,omitempty"`
}

// Feature summaries

// VulnsSummary contains vulnerability scanning summary
type VulnsSummary struct {
	TotalVulnerabilities int    `json:"total_vulnerabilities"`
	Critical             int    `json:"critical"`
	High                 int    `json:"high"`
	Medium               int    `json:"medium"`
	Low                  int    `json:"low"`
	KEVCount             int    `json:"kev_count"`
	Error                string `json:"error,omitempty"`
}

// HealthSummary contains package health summary
type HealthSummary struct {
	TotalPackages   int     `json:"total_packages"`
	AnalyzedCount   int     `json:"analyzed_count"`
	HealthyCount    int     `json:"healthy_count"`
	WarningCount    int     `json:"warning_count"`
	CriticalCount   int     `json:"critical_count"`
	AverageScore    float64 `json:"average_score"`
	DeprecatedCount int     `json:"deprecated_count"`
	OutdatedCount   int     `json:"outdated_count"`
	Error           string  `json:"error,omitempty"`
}

// LicensesSummary contains license analysis summary
type LicensesSummary struct {
	TotalPackages    int            `json:"total_packages"`
	UniqueLicenses   int            `json:"unique_licenses"`
	Allowed          int            `json:"allowed"`
	Denied           int            `json:"denied"`
	NeedsReview      int            `json:"needs_review"`
	Unknown          int            `json:"unknown"`
	PolicyViolations int            `json:"policy_violations"`
	LicenseCounts    map[string]int `json:"license_counts,omitempty"`
	Error            string         `json:"error,omitempty"`
}

// MalcontentSummary contains malware detection summary
type MalcontentSummary struct {
	TotalFiles    int    `json:"total_files"`
	TotalFindings int    `json:"total_findings"`
	Critical      int    `json:"critical"`
	High          int    `json:"high"`
	Medium        int    `json:"medium"`
	Low           int    `json:"low"`
	FilesWithRisk int    `json:"files_with_risk"`
	Error         string `json:"error,omitempty"`
}

// ConfusionSummary contains dependency confusion summary
type ConfusionSummary struct {
	TotalFindings int            `json:"total_findings"`
	Critical      int            `json:"critical"`
	High          int            `json:"high"`
	Medium        int            `json:"medium"`
	Low           int            `json:"low"`
	ByEcosystem   map[string]int `json:"by_ecosystem,omitempty"`
	Error         string         `json:"error,omitempty"`
}

// ReachabilitySummary contains vulnerability reachability summary
type ReachabilitySummary struct {
	Supported           bool    `json:"supported"`
	TotalVulns          int     `json:"total_vulns"`
	ReachableVulns      int     `json:"reachable_vulns"`
	UnreachableVulns    int     `json:"unreachable_vulns"`
	UnknownReachability int     `json:"unknown_reachability"`
	ReductionPercent    float64 `json:"reduction_percent"`
	Error               string  `json:"error,omitempty"`
}

// ProvenanceSummary contains provenance verification summary
type ProvenanceSummary struct {
	TotalPackages    int     `json:"total_packages"`
	VerifiedCount    int     `json:"verified_count"`
	UnverifiedCount  int     `json:"unverified_count"`
	SuspiciousCount  int     `json:"suspicious_count"`
	VerificationRate float64 `json:"verification_rate"`
	Error            string  `json:"error,omitempty"`
}

// BundleSummary contains bundle analysis summary
type BundleSummary struct {
	TotalPackages       int    `json:"total_packages"`
	HeavyPackages       int    `json:"heavy_packages"`
	DuplicatePackages   int    `json:"duplicate_packages"`
	TreeshakeCandidates int    `json:"treeshake_candidates"`
	TotalSizeKB         int    `json:"total_size_kb"`
	Error               string `json:"error,omitempty"`
}

// RecommendationsSummary contains package recommendations summary
type RecommendationsSummary struct {
	TotalRecommendations    int    `json:"total_recommendations"`
	SecurityRecommendations int    `json:"security_recommendations"`
	HealthRecommendations   int    `json:"health_recommendations"`
	Error                   string `json:"error,omitempty"`
}

// TyposquatsSummary contains typosquatting detection summary
type TyposquatsSummary struct {
	TotalChecked     int    `json:"total_checked"`
	SuspiciousCount  int    `json:"suspicious_count"`
	NewPackagesCount int    `json:"new_packages_count"` // Packages < 30 days old
	Error            string `json:"error,omitempty"`
}

// DeprecationsSummary contains deprecated package summary
type DeprecationsSummary struct {
	TotalPackages    int            `json:"total_packages"`
	DeprecatedCount  int            `json:"deprecated_count"`
	ByEcosystem      map[string]int `json:"by_ecosystem,omitempty"`
	Error            string         `json:"error,omitempty"`
}

// DuplicatesSummary contains duplicate dependency summary
type DuplicatesSummary struct {
	TotalPackages        int    `json:"total_packages"`
	DuplicateVersions    int    `json:"duplicate_versions"`    // Same package, different versions
	DuplicateFunctionality int  `json:"duplicate_functionality"` // Different packages, same purpose
	Error                string `json:"error,omitempty"`
}

// Finding types

// VulnFinding represents a vulnerability finding
type VulnFinding struct {
	ID        string   `json:"id"`
	Aliases   []string `json:"aliases,omitempty"`
	Package   string   `json:"package"`
	Version   string   `json:"version"`
	Ecosystem string   `json:"ecosystem"`
	Severity  string   `json:"severity"`
	Title     string   `json:"title,omitempty"`
	FixedIn   string   `json:"fixed_in,omitempty"`
	InKEV     bool     `json:"in_kev"`
}

// HealthFinding represents a package health finding
type HealthFinding struct {
	Package       string  `json:"package"`
	Version       string  `json:"version"`
	Ecosystem     string  `json:"ecosystem"`
	Purl          string  `json:"purl"`
	HealthScore   float64 `json:"health_score"`
	Status        string  `json:"status"`
	IsDeprecated  bool    `json:"is_deprecated"`
	IsOutdated    bool    `json:"is_outdated"`
	LatestVersion string  `json:"latest_version,omitempty"`
}

// LicenseFinding represents a license finding
type LicenseFinding struct {
	Package   string   `json:"package"`
	Version   string   `json:"version"`
	Ecosystem string   `json:"ecosystem"`
	Licenses  []string `json:"licenses"`
	Status    string   `json:"status"`
}

// MalcontentFinding represents a malware detection finding
type MalcontentFinding struct {
	File      string   `json:"file"`
	Risk      string   `json:"risk"`
	RiskScore int      `json:"risk_score"`
	Behaviors []string `json:"behaviors"`
}

// ConfusionFinding represents a dependency confusion finding
type ConfusionFinding struct {
	Package       string `json:"package"`
	Ecosystem     string `json:"ecosystem"`
	RiskLevel     string `json:"risk_level"`
	RiskType      string `json:"risk_type"`
	Description   string `json:"description"`
	PublicExists  bool   `json:"public_exists"`
	PublicVersion string `json:"public_version,omitempty"`
	LocalVersion  string `json:"local_version,omitempty"`
	File          string `json:"file"`
}

// ReachabilityFinding represents a reachability analysis finding
type ReachabilityFinding struct {
	ID                 string `json:"id"`
	Package            string `json:"package"`
	Version            string `json:"version"`
	Summary            string `json:"summary"`
	ReachabilityStatus string `json:"reachability_status"`
	Reachable          bool   `json:"reachable"`
}

// ProvenanceFinding represents a provenance verification finding
type ProvenanceFinding struct {
	Package     string `json:"package"`
	Version     string `json:"version"`
	Verified    bool   `json:"verified"`
	Source      string `json:"source,omitempty"`
	Attestation string `json:"attestation,omitempty"`
}

// BundleFinding represents a bundle analysis finding
type BundleFinding struct {
	Package    string `json:"package"`
	Version    string `json:"version"`
	SizeKB     int    `json:"size_kb"`
	Issue      string `json:"issue"`
	Suggestion string `json:"suggestion,omitempty"`
}

// RecommendationFinding represents a package recommendation
type RecommendationFinding struct {
	Package        string `json:"package"`
	CurrentVersion string `json:"current_version"`
	Alternative    string `json:"alternative,omitempty"`
	Reason         string `json:"reason"`
	Priority       string `json:"priority"`
}

// TyposquatFinding represents a typosquatting finding
type TyposquatFinding struct {
	Package        string `json:"package"`
	Ecosystem      string `json:"ecosystem"`
	SimilarTo      string `json:"similar_to,omitempty"`
	Reason         string `json:"reason"`
	AgeInDays      int    `json:"age_in_days,omitempty"`
	RiskLevel      string `json:"risk_level"`
}

// DeprecationFinding represents a deprecated package finding
type DeprecationFinding struct {
	Package       string `json:"package"`
	Version       string `json:"version"`
	Ecosystem     string `json:"ecosystem"`
	Message       string `json:"message,omitempty"`
	Alternative   string `json:"alternative,omitempty"`
}

// DuplicateFinding represents a duplicate dependency finding
type DuplicateFinding struct {
	Package   string   `json:"package"`
	Versions  []string `json:"versions,omitempty"`
	IssueType string   `json:"issue_type"` // "version" or "functionality"
	Message   string   `json:"message"`
}

// =============================================================================
// SBOM Generation Types (merged from sbom scanner)
// =============================================================================

// GenerationSummary contains SBOM generation summary
type GenerationSummary struct {
	Tool            string         `json:"tool"`
	SpecVersion     string         `json:"spec_version"`
	TotalComponents int            `json:"total_components"`
	ByType          map[string]int `json:"by_type"`
	ByEcosystem     map[string]int `json:"by_ecosystem"`
	HasDependencies bool           `json:"has_dependencies"`
	SBOMPath        string         `json:"sbom_path"`
	Error           string         `json:"error,omitempty"`
}

// IntegritySummary contains SBOM integrity summary
type IntegritySummary struct {
	IsComplete      bool   `json:"is_complete"`
	DriftDetected   bool   `json:"drift_detected"`
	MissingPackages int    `json:"missing_packages"`
	ExtraPackages   int    `json:"extra_packages"`
	LockfilesFound  int    `json:"lockfiles_found"`
	Error           string `json:"error,omitempty"`
}

// GenerationFindings contains SBOM data
type GenerationFindings struct {
	Components   []Component   `json:"components"`
	Dependencies []Dependency  `json:"dependencies,omitempty"`
	Metadata     *SBOMMetadata `json:"metadata,omitempty"`
}

// Component represents a software component in the SBOM
type Component struct {
	Type       string     `json:"type"`
	Name       string     `json:"name"`
	Version    string     `json:"version"`
	Purl       string     `json:"purl,omitempty"`
	Ecosystem  string     `json:"ecosystem,omitempty"`
	Licenses   []string   `json:"licenses,omitempty"`
	Hashes     []Hash     `json:"hashes,omitempty"`
	Scope      string     `json:"scope,omitempty"` // required, optional, dev
	Properties []Property `json:"properties,omitempty"`
}

// Hash represents a component hash
type Hash struct {
	Algorithm string `json:"alg"`
	Content   string `json:"content"`
}

// Property represents a component property
type Property struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Dependency represents a dependency relationship
type Dependency struct {
	Ref       string   `json:"ref"`
	DependsOn []string `json:"dependsOn,omitempty"`
}

// SBOMMetadata contains SBOM document metadata
type SBOMMetadata struct {
	BomFormat    string `json:"bomFormat"`
	SpecVersion  string `json:"specVersion"`
	Version      int    `json:"version"`
	SerialNumber string `json:"serialNumber,omitempty"`
	Timestamp    string `json:"timestamp,omitempty"`
	Tool         string `json:"tool,omitempty"`
}

// IntegrityFindings contains integrity verification findings
type IntegrityFindings struct {
	LockfileComparisons []LockfileComparison `json:"lockfile_comparisons,omitempty"`
	MissingPackages     []MissingPackage     `json:"missing_packages,omitempty"`
	ExtraPackages       []ExtraPackage       `json:"extra_packages,omitempty"`
	DriftDetails        *DriftDetails        `json:"drift_details,omitempty"`
}

// LockfileComparison contains comparison between SBOM and lockfile
type LockfileComparison struct {
	Lockfile   string `json:"lockfile"`
	Ecosystem  string `json:"ecosystem"`
	InSBOM     int    `json:"in_sbom"`
	InLockfile int    `json:"in_lockfile"`
	Matched    int    `json:"matched"`
	Missing    int    `json:"missing"`
	Extra      int    `json:"extra"`
}

// MissingPackage represents a package in lockfile but not in SBOM
type MissingPackage struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Ecosystem string `json:"ecosystem"`
	Lockfile  string `json:"lockfile"`
}

// ExtraPackage represents a package in SBOM but not in lockfile
type ExtraPackage struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Ecosystem string `json:"ecosystem"`
}

// DriftDetails contains SBOM drift information
type DriftDetails struct {
	PreviousSBOMPath string          `json:"previous_sbom_path,omitempty"`
	Added            []ComponentDiff `json:"added,omitempty"`
	Removed          []ComponentDiff `json:"removed,omitempty"`
	VersionChanged   []VersionChange `json:"version_changed,omitempty"`
	TotalAdded       int             `json:"total_added"`
	TotalRemoved     int             `json:"total_removed"`
	TotalChanged     int             `json:"total_changed"`
}

// ComponentDiff represents a component that was added or removed
type ComponentDiff struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Ecosystem string `json:"ecosystem,omitempty"`
}

// VersionChange represents a version change between SBOMs
type VersionChange struct {
	Name       string `json:"name"`
	Ecosystem  string `json:"ecosystem,omitempty"`
	OldVersion string `json:"old_version"`
	NewVersion string `json:"new_version"`
}

// =============================================================================
// CycloneDX Parsing Types
// =============================================================================

// CycloneDXTool represents a tool in CycloneDX format
type CycloneDXTool struct {
	Type    string `json:"type,omitempty"`
	Author  string `json:"author,omitempty"`
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

// CycloneDXTools handles both CycloneDX 1.4 (array) and 1.5 (object with components) formats
type CycloneDXTools struct {
	Tools []CycloneDXTool // Parsed tools regardless of format
}

// UnmarshalJSON handles both array (1.4) and object (1.5) formats for tools
func (t *CycloneDXTools) UnmarshalJSON(data []byte) error {
	// Try CycloneDX 1.5 format first: {"components": [...]}
	var v15 struct {
		Components []CycloneDXTool `json:"components"`
	}
	if err := json.Unmarshal(data, &v15); err == nil && len(v15.Components) > 0 {
		t.Tools = v15.Components
		return nil
	}

	// Try CycloneDX 1.4 format: [...]
	var v14 []CycloneDXTool
	if err := json.Unmarshal(data, &v14); err == nil {
		t.Tools = v14
		return nil
	}

	// If neither works, just leave tools empty (don't fail)
	t.Tools = nil
	return nil
}

// CycloneDXBOM represents the CycloneDX SBOM structure for parsing
type CycloneDXBOM struct {
	BomFormat    string `json:"bomFormat"`
	SpecVersion  string `json:"specVersion"`
	Version      int    `json:"version"`
	SerialNumber string `json:"serialNumber,omitempty"`
	Metadata     struct {
		Timestamp string          `json:"timestamp,omitempty"`
		Tools     *CycloneDXTools `json:"tools,omitempty"`
		Component struct {
			BomRef string `json:"bom-ref,omitempty"`
			Type   string `json:"type,omitempty"`
			Name   string `json:"name,omitempty"`
		} `json:"component,omitempty"`
	} `json:"metadata,omitempty"`
	Components []struct {
		Type    string `json:"type"`
		Name    string `json:"name"`
		Version string `json:"version"`
		Purl    string `json:"purl,omitempty"`
		BomRef  string `json:"bom-ref,omitempty"`
		Scope   string `json:"scope,omitempty"`
		Licenses []struct {
			License struct {
				ID   string `json:"id,omitempty"`
				Name string `json:"name,omitempty"`
			} `json:"license,omitempty"`
			Expression string `json:"expression,omitempty"`
		} `json:"licenses,omitempty"`
		Hashes []struct {
			Alg     string `json:"alg"`
			Content string `json:"content"`
		} `json:"hashes,omitempty"`
		Properties []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"properties,omitempty"`
	} `json:"components,omitempty"`
	Dependencies []struct {
		Ref       string   `json:"ref"`
		DependsOn []string `json:"dependsOn,omitempty"`
	} `json:"dependencies,omitempty"`
}
