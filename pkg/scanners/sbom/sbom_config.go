// Package sbom provides configuration loading from sbom.config.json
package sbom

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// SBOMConfigFile represents the structure of sbom.config.json
type SBOMConfigFile struct {
	Schema   string `json:"$schema"`
	Version  string `json:"_version"`
	Output   OutputConfig `json:"output"`
	Dependencies DependenciesConfig `json:"dependencies"`
	Metadata MetadataConfig `json:"metadata"`
	ToolPreferences ToolPreferencesConfig `json:"tool_preferences"`
	PackageManagers PackageManagersConfig `json:"package_managers"`
	Integrity IntegrityFileConfig `json:"integrity"`
	Filtering FilteringConfig `json:"filtering"`
	Compliance ComplianceConfig `json:"compliance"`
}

// OutputConfig configures SBOM output format
type OutputConfig struct {
	Format             string `json:"format"`
	SpecVersion        string `json:"spec_version"`
	OutputFormat       string `json:"output_format"`
	Filename           string `json:"filename"`
	IncludeSerialNumber bool `json:"include_serial_number"`
	IncludeTimestamp   bool `json:"include_timestamp"`
}

// DependenciesConfig configures which dependencies to include
type DependenciesConfig struct {
	IncludeDev        bool `json:"include_dev"`
	IncludeTest       bool `json:"include_test"`
	IncludeOptional   bool `json:"include_optional"`
	IncludePeer       bool `json:"include_peer"`
	IncludeTransitive bool `json:"include_transitive"`
	IncludeBuild      bool `json:"include_build"`
}

// MetadataConfig configures SBOM metadata
type MetadataConfig struct {
	IncludeLicenses     bool `json:"include_licenses"`
	IncludeHashes       bool `json:"include_hashes"`
	IncludeURLs         bool `json:"include_urls"`
	IncludeDescriptions bool `json:"include_descriptions"`
	IncludeAuthors      bool `json:"include_authors"`
	IncludePurl         bool `json:"include_purl"`
	IncludeCPE          bool `json:"include_cpe"`
}

// ToolPreferencesConfig configures tool selection per ecosystem
type ToolPreferencesConfig struct {
	Default         string `json:"default"`
	FallbackEnabled bool   `json:"fallback_enabled"`
	JavaScript      *EcosystemToolConfig `json:"javascript,omitempty"`
	Python          *EcosystemToolConfig `json:"python,omitempty"`
	Rust            *EcosystemToolConfig `json:"rust,omitempty"`
	Go              *EcosystemToolConfig `json:"go,omitempty"`
	Java            *EcosystemToolConfig `json:"java,omitempty"`
	Ruby            *EcosystemToolConfig `json:"ruby,omitempty"`
	PHP             *EcosystemToolConfig `json:"php,omitempty"`
	DotNet          *EcosystemToolConfig `json:"dotnet,omitempty"`
}

// EcosystemToolConfig configures tools for a specific ecosystem
type EcosystemToolConfig struct {
	Tool     string                 `json:"tool"`
	Fallback string                 `json:"fallback"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

// PackageManagersConfig contains per-package-manager settings
type PackageManagersConfig struct {
	NPM      *NPMConfig      `json:"npm,omitempty"`
	Yarn     *YarnConfig     `json:"yarn,omitempty"`
	PNPM     *PNPMConfig     `json:"pnpm,omitempty"`
	Pip      *PipConfig      `json:"pip,omitempty"`
	Poetry   *PoetryConfig   `json:"poetry,omitempty"`
	UV       *UVConfig       `json:"uv,omitempty"`
	Cargo    *CargoConfig    `json:"cargo,omitempty"`
	GoMod    *GoConfig       `json:"go,omitempty"`
	Maven    *MavenConfig    `json:"maven,omitempty"`
	Gradle   *GradleConfig   `json:"gradle,omitempty"`
	Bundler  *BundlerConfig  `json:"bundler,omitempty"`
	Composer *ComposerConfig `json:"composer,omitempty"`
	NuGet    *NuGetConfig    `json:"nuget,omitempty"`
}

// NPMConfig contains npm-specific settings
type NPMConfig struct {
	UseNativeSBOM   bool     `json:"use_native_sbom"`
	Omit            []string `json:"omit"`
	PackageLockOnly bool     `json:"package_lock_only"`
	Workspaces      bool     `json:"workspaces"`
}

// YarnConfig contains yarn-specific settings
type YarnConfig struct {
	BerryMode      bool `json:"berry_mode"`
	ProductionOnly bool `json:"production_only"`
	Workspaces     bool `json:"workspaces"`
}

// PNPMConfig contains pnpm-specific settings
type PNPMConfig struct {
	ProductionOnly bool   `json:"production_only"`
	Workspaces     bool   `json:"workspaces"`
	Filter         string `json:"filter,omitempty"`
}

// PipConfig contains pip-specific settings
type PipConfig struct {
	RequirementsFile   string `json:"requirements_file"`
	UseEnvironment     bool   `json:"use_environment"`
	FetchPyPIMetadata  bool   `json:"fetch_pypi_metadata"`
}

// PoetryConfig contains poetry-specific settings
type PoetryConfig struct {
	Groups       []string `json:"groups"`
	Extras       []string `json:"extras"`
	FromLockFile bool     `json:"from_lock_file"`
}

// UVConfig contains uv-specific settings
type UVConfig struct {
	IncludeDev   bool `json:"include_dev"`
	FromLockFile bool `json:"from_lock_file"`
}

// CargoConfig contains cargo-specific settings
type CargoConfig struct {
	IncludeDev   bool     `json:"include_dev"`
	IncludeBuild bool     `json:"include_build"`
	AllFeatures  bool     `json:"all_features"`
	Features     []string `json:"features"`
}

// GoConfig contains go-specific settings
type GoConfig struct {
	IncludeTest bool   `json:"include_test"`
	IncludeStd  bool   `json:"include_std"`
	FromBinary  bool   `json:"from_binary"`
	BinaryPath  string `json:"binary_path,omitempty"`
}

// MavenConfig contains maven-specific settings
type MavenConfig struct {
	IncludeCompile   bool `json:"include_compile"`
	IncludeRuntime   bool `json:"include_runtime"`
	IncludeTest      bool `json:"include_test"`
	IncludeProvided  bool `json:"include_provided"`
	AggregateModules bool `json:"aggregate_modules"`
}

// GradleConfig contains gradle-specific settings
type GradleConfig struct {
	Configurations     []string `json:"configurations"`
	SkipConfigurations []string `json:"skip_configurations"`
	UseLockFile        bool     `json:"use_lock_file"`
}

// BundlerConfig contains bundler-specific settings
type BundlerConfig struct {
	ExcludeGroups []string `json:"exclude_groups"`
	IncludeGroups []string `json:"include_groups"`
}

// ComposerConfig contains composer-specific settings
type ComposerConfig struct {
	IncludeDev            bool `json:"include_dev"`
	FetchPackagistMetadata bool `json:"fetch_packagist_metadata"`
}

// NuGetConfig contains nuget-specific settings
type NuGetConfig struct {
	IncludeTransitive bool     `json:"include_transitive"`
	ExcludeDev        bool     `json:"exclude_dev"`
	UseLockFile       bool     `json:"use_lock_file"`
	TargetFrameworks  []string `json:"target_frameworks"`
}

// IntegrityFileConfig configures integrity verification from config file
type IntegrityFileConfig struct {
	VerifyLockfiles  bool `json:"verify_lockfiles"`
	RequireLockfile  bool `json:"require_lockfile"`
	VerifyChecksums  bool `json:"verify_checksums"`
	DetectDrift      bool `json:"detect_drift"`
}

// FilteringConfig configures package filtering
type FilteringConfig struct {
	ExcludePatterns []string `json:"exclude_patterns"`
	IncludePatterns []string `json:"include_patterns"`
	ExcludeScopes   []string `json:"exclude_scopes"`
}

// ComplianceConfig configures compliance settings
type ComplianceConfig struct {
	NTIAMinimum    bool     `json:"ntia_minimum"`
	CISA2025       bool     `json:"cisa_2025"`
	RequiredFields []string `json:"required_fields"`
}

// LoadSBOMConfig loads configuration from sbom.config.json
func LoadSBOMConfig(configPath string) (*SBOMConfigFile, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg SBOMConfigFile
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// FindSBOMConfig searches for sbom.config.json in standard locations
func FindSBOMConfig(repoPath string) string {
	// Check locations in order of priority
	locations := []string{
		filepath.Join(repoPath, "sbom.config.json"),
		filepath.Join(repoPath, ".zero", "sbom.config.json"),
		filepath.Join(repoPath, "config", "sbom.config.json"),
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc
		}
	}

	return ""
}

// ApplySBOMConfig applies sbom.config.json settings to FeatureConfig
func ApplySBOMConfig(cfg *FeatureConfig, sbomCfg *SBOMConfigFile) {
	// Apply output settings
	if sbomCfg.Output.SpecVersion != "" {
		cfg.Generation.SpecVersion = sbomCfg.Output.SpecVersion
	}
	if sbomCfg.Output.OutputFormat != "" {
		cfg.Generation.Format = sbomCfg.Output.OutputFormat
	}

	// Apply dependency settings
	cfg.Generation.IncludeDev = sbomCfg.Dependencies.IncludeDev

	// Apply tool preferences
	if sbomCfg.ToolPreferences.Default != "" {
		cfg.Generation.Tool = sbomCfg.ToolPreferences.Default
	}
	cfg.Generation.FallbackToSyft = sbomCfg.ToolPreferences.FallbackEnabled

	// Apply integrity settings
	cfg.Integrity.VerifyLockfiles = sbomCfg.Integrity.VerifyLockfiles
	cfg.Integrity.DetectDrift = sbomCfg.Integrity.DetectDrift
}
