package sbom

// Result holds all SBOM feature results
type Result struct {
	FeaturesRun []string `json:"features_run"`
	Summary     Summary  `json:"summary"`
	Findings    Findings `json:"findings"`
}

// Summary holds summaries from all features
type Summary struct {
	Generation *GenerationSummary `json:"generation,omitempty"`
	Integrity  *IntegritySummary  `json:"integrity,omitempty"`
	Errors     []string           `json:"errors,omitempty"`
}

// Findings holds findings from all features
type Findings struct {
	Generation *GenerationFindings `json:"generation,omitempty"`
	Integrity  *IntegrityFindings  `json:"integrity,omitempty"`
}

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
	Type       string   `json:"type"`
	Name       string   `json:"name"`
	Version    string   `json:"version"`
	Purl       string   `json:"purl,omitempty"`
	Ecosystem  string   `json:"ecosystem,omitempty"`
	Licenses   []string `json:"licenses,omitempty"`
	Hashes     []Hash   `json:"hashes,omitempty"`
	Scope      string   `json:"scope,omitempty"` // required, optional, dev
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
	BomFormat   string `json:"bomFormat"`
	SpecVersion string `json:"specVersion"`
	Version     int    `json:"version"`
	SerialNumber string `json:"serialNumber,omitempty"`
	Timestamp   string `json:"timestamp,omitempty"`
	Tool        string `json:"tool,omitempty"`
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
	Lockfile     string `json:"lockfile"`
	Ecosystem    string `json:"ecosystem"`
	InSBOM       int    `json:"in_sbom"`
	InLockfile   int    `json:"in_lockfile"`
	Matched      int    `json:"matched"`
	Missing      int    `json:"missing"`
	Extra        int    `json:"extra"`
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
	Name        string `json:"name"`
	Ecosystem   string `json:"ecosystem,omitempty"`
	OldVersion  string `json:"old_version"`
	NewVersion  string `json:"new_version"`
}

// CycloneDXBOM represents the CycloneDX SBOM structure for parsing
type CycloneDXBOM struct {
	BomFormat    string `json:"bomFormat"`
	SpecVersion  string `json:"specVersion"`
	Version      int    `json:"version"`
	SerialNumber string `json:"serialNumber,omitempty"`
	Metadata     struct {
		Timestamp string `json:"timestamp,omitempty"`
		Tools     []struct {
			Name    string `json:"name,omitempty"`
			Version string `json:"version,omitempty"`
		} `json:"tools,omitempty"`
	} `json:"metadata,omitempty"`
	Components []struct {
		Type       string `json:"type"`
		Name       string `json:"name"`
		Version    string `json:"version"`
		Purl       string `json:"purl,omitempty"`
		BomRef     string `json:"bom-ref,omitempty"`
		Scope      string `json:"scope,omitempty"`
		Licenses   []struct {
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
