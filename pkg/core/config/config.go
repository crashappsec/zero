// Package config handles Zero configuration loading and management
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the main Zero configuration from zero.config.json
type Config struct {
	Version  string              `json:"_version"`
	Settings Settings            `json:"settings"`
	Scanners map[string]Scanner  `json:"scanners"`
	Profiles map[string]Profile  `json:"profiles"`
}

// Settings contains global settings
type Settings struct {
	DefaultProfile        string      `json:"default_profile"`
	StoragePath           string      `json:"storage_path"`
	ParallelRepos         int         `json:"parallel_repos"`          // Parallel repo processing (default: 1 for sequential)
	ParallelScanners      int         `json:"parallel_scanners"`       // Parallel scanner execution per repo (default: 4)
	ScannerTimeoutSeconds int         `json:"scanner_timeout_seconds"`
	CacheTTLHours         int         `json:"cache_ttl_hours"`
	Environment           Environment `json:"environment"`
}

// Environment contains environment variable names
type Environment struct {
	GitHubTokenEnv   string `json:"github_token_env"`
	AnthropicKeyEnv  string `json:"anthropic_key_env"`
}

// Scanner defines a scanner configuration
type Scanner struct {
	Name           string         `json:"name"`
	Description    string         `json:"description"`
	Script         string         `json:"script"`
	ClaudeMode     string         `json:"claude_mode"`
	ClaudeFeatures []string       `json:"claude_features"`
	EstimatedTime  string         `json:"estimated_time"`
	OutputFile     string         `json:"output_file"`
	Options        ScannerOptions `json:"options,omitempty"`
}

// ScannerOptions contains scanner-specific configuration
type ScannerOptions struct {
	SBOM    *SBOMOptions    `json:"sbom,omitempty"`
	Secrets *SecretsOptions `json:"secrets,omitempty"`
	CodeSec *CodeSecOptions `json:"code_security,omitempty"`
}

// SBOMOptions configures SBOM generation (cdxgen/syft)
type SBOMOptions struct {
	// Tool preference: "cdxgen", "syft", or "auto" (default: auto - prefers cdxgen)
	Tool string `json:"tool"`
	// SpecVersion for CycloneDX (default: 1.5)
	SpecVersion string `json:"spec_version"`
	// Format output format: "json", "xml" (default: json)
	Format string `json:"format"`
	// Recurse into subdirectories for mono-repos (default: true)
	Recurse bool `json:"recurse"`
	// InstallDeps install dependencies before scanning (default: false for speed)
	InstallDeps bool `json:"install_deps"`
	// BabelAnalysis run babel for JS/TS usage analysis (default: false for speed)
	BabelAnalysis bool `json:"babel_analysis"`
	// Deep perform deep searches for C/C++, live OS, OCI images (default: false)
	Deep bool `json:"deep"`
	// Evidence generate SBOM with evidence (default: false)
	Evidence bool `json:"evidence"`
	// Profile cdxgen profile: "generic", "research", "appsec", "operational" (default: generic)
	Profile string `json:"profile"`
	// ExcludePatterns glob patterns to exclude
	ExcludePatterns []string `json:"exclude_patterns,omitempty"`
	// FallbackToSyft use syft if cdxgen fails (default: true)
	FallbackToSyft bool `json:"fallback_to_syft"`
	// TimeoutSeconds override default timeout for SBOM generation
	TimeoutSeconds int `json:"timeout_seconds,omitempty"`
}

// SecretsOptions configures secrets detection
type SecretsOptions struct {
	// Tool preference: "gitleaks", "trufflehog", "detect-secrets" (default: gitleaks)
	Tool string `json:"tool"`
	// IncludePatterns additional regex patterns to detect
	IncludePatterns []string `json:"include_patterns,omitempty"`
	// ExcludePatterns patterns to ignore
	ExcludePatterns []string `json:"exclude_patterns,omitempty"`
	// ExcludePaths paths to skip
	ExcludePaths []string `json:"exclude_paths,omitempty"`
	// Entropy enable entropy-based detection (default: true)
	Entropy bool `json:"entropy"`
}

// CodeSecOptions configures code security scanning
type CodeSecOptions struct {
	// Tool preference: "semgrep", "codeql" (default: semgrep)
	Tool string `json:"tool"`
	// Rulesets semgrep rulesets to use (default: ["auto"])
	Rulesets []string `json:"rulesets,omitempty"`
	// ExcludePaths paths to skip
	ExcludePaths []string `json:"exclude_paths,omitempty"`
	// Severity minimum severity: "INFO", "WARNING", "ERROR" (default: WARNING)
	Severity string `json:"severity"`
}

// Profile defines a scanning profile with its scanners
type Profile struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	EstimatedTime  string   `json:"estimated_time"`
	ClaudeMode     string   `json:"claude_mode"`
	SBOMGenerator  string   `json:"sbom_generator"`
	ClaudeScanners []string `json:"claude_scanners"`
	Scanners       []string `json:"scanners"`
}

// DefaultConfig returns sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Version: "1.0",
		Settings: Settings{
			DefaultProfile:        "standard",
			StoragePath:           ".zero",
			ParallelRepos:         1, // Sequential repo processing for cleaner output
			ParallelScanners:      4, // Parallel scanner execution for speed
			ScannerTimeoutSeconds: 300,
			CacheTTLHours:         24,
		},
		Profiles: map[string]Profile{
			"packages": {
				Name:        "Packages",
				Description: "Package-focused analysis",
				Scanners: []string{
					"package-sbom",
					"package-vulns",
					"package-health",
					"package-provenance",
					"package-malcontent",
					"package-bundle-optimization",
					"package-recommendations",
				},
			},
		},
	}
}

// Load reads configuration from the default location
func Load() (*Config, error) {
	// Find config file
	configPath := findConfigFile()
	if configPath == "" {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	// Apply defaults for missing values
	if cfg.Settings.ParallelRepos == 0 {
		cfg.Settings.ParallelRepos = 1 // Sequential repo processing by default
	}
	if cfg.Settings.ParallelScanners == 0 {
		cfg.Settings.ParallelScanners = 4 // Parallel scanner execution by default
	}
	if cfg.Settings.ScannerTimeoutSeconds == 0 {
		cfg.Settings.ScannerTimeoutSeconds = 300
	}
	if cfg.Settings.StoragePath == "" {
		cfg.Settings.StoragePath = ".zero"
	}
	if cfg.Settings.DefaultProfile == "" {
		cfg.Settings.DefaultProfile = "standard"
	}

	return &cfg, nil
}

// findConfigFile looks for zero.config.json in standard locations
func findConfigFile() string {
	// Get executable directory for relative path resolution
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)

	candidates := []string{
		"config/zero.config.json",
		filepath.Join(exeDir, "..", "config/zero.config.json"),
		"zero.config.json",
		filepath.Join(os.Getenv("HOME"), ".zero", "config.json"),
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

// GetProfileScanners returns the scanner list for a profile
func (c *Config) GetProfileScanners(profile string) ([]string, error) {
	p, ok := c.Profiles[profile]
	if !ok {
		return nil, fmt.Errorf("unknown profile: %s (available: %v)", profile, c.GetProfileNames())
	}
	return p.Scanners, nil
}

// GetProfileNames returns all available profile names
func (c *Config) GetProfileNames() []string {
	names := make([]string, 0, len(c.Profiles))
	for name := range c.Profiles {
		names = append(names, name)
	}
	return names
}

// GetScannerTimeout returns the timeout for a scanner (in seconds)
func (c *Config) GetScannerTimeout(scanner string) int {
	return c.Settings.ScannerTimeoutSeconds
}

// GetScanner returns scanner configuration by name
func (c *Config) GetScanner(name string) (*Scanner, bool) {
	s, ok := c.Scanners[name]
	if !ok {
		return nil, false
	}
	return &s, true
}

// GetScannerFeatures returns the features configuration for a scanner as a map
// This is used to pass scanner-specific feature configuration to scanners
func (c *Config) GetScannerFeatures(name string) map[string]interface{} {
	s, ok := c.Scanners[name]
	if !ok {
		return nil
	}

	// The scanner struct has a Features field that contains the raw JSON
	// We need to extract it from the raw config
	configPath := findConfigFile()
	if configPath == "" {
		return nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil
	}

	// Parse into a generic map to access features
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil
	}

	scanners, ok := raw["scanners"].(map[string]interface{})
	if !ok {
		return nil
	}

	scanner, ok := scanners[name].(map[string]interface{})
	if !ok {
		return nil
	}

	features, ok := scanner["features"].(map[string]interface{})
	if !ok {
		return nil
	}

	// For consistency with scanner expectations, we use the description to detect misconfig
	_ = s.Description

	return features
}

// GetProfile returns profile configuration by name
func (c *Config) GetProfile(name string) (*Profile, bool) {
	p, ok := c.Profiles[name]
	if !ok {
		return nil, false
	}
	return &p, true
}

// ZeroHome returns the storage path (resolves to absolute if needed)
func (c *Config) ZeroHome() string {
	return c.Settings.StoragePath
}

// SlowScanners returns scanners that are known to be slow on large repos
func (c *Config) SlowScanners() []string {
	return []string{
		"package-malcontent",
		"code-vulns",
		"code-secrets",
		"bundle-analysis",
		"container-security",
	}
}

// IsSlowScanner returns true if the scanner is known to be slow
func (c *Config) IsSlowScanner(name string) bool {
	for _, s := range c.SlowScanners() {
		if s == name {
			return true
		}
	}
	return false
}

// GetSBOMOptions returns SBOM configuration with sensible defaults
func (c *Config) GetSBOMOptions() *SBOMOptions {
	// Check if scanner has custom options
	if scanner, ok := c.Scanners["package-sbom"]; ok {
		if scanner.Options.SBOM != nil {
			opts := scanner.Options.SBOM
			// Apply defaults for unset values
			if opts.Tool == "" {
				opts.Tool = "auto"
			}
			if opts.SpecVersion == "" {
				opts.SpecVersion = "1.5"
			}
			if opts.Format == "" {
				opts.Format = "json"
			}
			if opts.Profile == "" {
				opts.Profile = "generic"
			}
			return opts
		}
	}

	// Return defaults optimized for speed
	return &SBOMOptions{
		Tool:           "auto",      // Prefer cdxgen, fall back to syft
		SpecVersion:    "1.5",
		Format:         "json",
		Recurse:        true,        // Support mono-repos
		InstallDeps:    false,       // Speed optimization
		BabelAnalysis:  false,       // Speed optimization
		Deep:           false,       // Speed optimization
		Evidence:       false,       // Speed optimization
		Profile:        "generic",
		FallbackToSyft: true,        // Resilience
		TimeoutSeconds: 300,         // 5 minutes default
	}
}

// GetSecretsOptions returns secrets scanner configuration with sensible defaults
func (c *Config) GetSecretsOptions() *SecretsOptions {
	if scanner, ok := c.Scanners["code-secrets"]; ok {
		if scanner.Options.Secrets != nil {
			return scanner.Options.Secrets
		}
	}
	return &SecretsOptions{
		Tool:    "gitleaks",
		Entropy: true,
	}
}

// GetCodeSecOptions returns code security scanner configuration with sensible defaults
func (c *Config) GetCodeSecOptions() *CodeSecOptions {
	if scanner, ok := c.Scanners["code-vulns"]; ok {
		if scanner.Options.CodeSec != nil {
			return scanner.Options.CodeSec
		}
	}
	return &CodeSecOptions{
		Tool:     "semgrep",
		Rulesets: []string{"auto"},
		Severity: "WARNING",
	}
}

// SheetsConfig holds configuration for Google Sheets export
type SheetsConfig struct {
	ClientID     string `json:"oauth_client_id"`
	ClientSecret string `json:"oauth_client_secret"`
	TokenPath    string `json:"token_cache_path"`
	Enabled      bool   `json:"enabled"`
}

// GetSheetsConfig returns Google Sheets export configuration
func (c *Config) GetSheetsConfig() *SheetsConfig {
	// Check environment variables first (override)
	clientID := os.Getenv("ZERO_SHEETS_CLIENT_ID")
	clientSecret := os.Getenv("ZERO_SHEETS_CLIENT_SECRET")
	tokenPath := os.Getenv("ZERO_SHEETS_TOKEN_PATH")

	// Default token path
	if tokenPath == "" {
		home, _ := os.UserHomeDir()
		tokenPath = filepath.Join(home, ".zero", "google-token.json")
	}

	return &SheetsConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenPath:    tokenPath,
		Enabled:      clientID != "" && clientSecret != "",
	}
}
