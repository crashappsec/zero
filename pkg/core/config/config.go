// Package config handles Zero configuration loading and management
package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// Config represents the main Zero configuration
type Config struct {
	Version  string             `json:"_version"`
	Settings Settings           `json:"settings"`
	Profiles map[string]Profile `json:"profiles"`
	Scanners map[string]Scanner `json:"scanners,omitempty"` // Loaded from defaults
}

// Settings contains global settings
type Settings struct {
	DefaultProfile        string `json:"default_profile"`
	StoragePath           string `json:"storage_path"`
	ParallelRepos         int    `json:"parallel_repos"`
	ParallelScanners      int    `json:"parallel_scanners"`
	ScannerTimeoutSeconds int    `json:"scanner_timeout_seconds"`
	CacheTTLHours         int    `json:"cache_ttl_hours"`
	PersonalityMode       string `json:"personality_mode,omitempty"` // "full", "minimal", "neutral" (default: "minimal")
}

// Scanner defines a scanner configuration with features
type Scanner struct {
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	OutputFile    string                 `json:"output_file"`
	EstimatedTime string                 `json:"estimated_time,omitempty"`
	Dependencies  []string               `json:"dependencies,omitempty"`
	Features      map[string]interface{} `json:"features,omitempty"`
}

// Profile defines a scanning profile
type Profile struct {
	Name             string                            `json:"name"`
	Description      string                            `json:"description"`
	EstimatedTime    string                            `json:"estimated_time,omitempty"`
	Scanners         []string                          `json:"scanners"`
	FeatureOverrides map[string]map[string]interface{} `json:"feature_overrides,omitempty"`
}

// DefaultConfig returns a Config with default values (no file loading)
// Useful for testing and when config files are not available
func DefaultConfig() *Config {
	cfg := &Config{
		Version: "1.0",
		Settings: Settings{
			DefaultProfile:        "standard",
			StoragePath:           ".zero",
			ParallelRepos:         1,
			ParallelScanners:      4,
			ScannerTimeoutSeconds: 300,
			CacheTTLHours:         24,
		},
		Profiles: map[string]Profile{
			"packages": {
				Name:        "Packages",
				Description: "Package analysis profile",
				Scanners:    []string{"code-packages", "code-security", "code-quality"},
			},
			"standard": {
				Name:        "Standard",
				Description: "Standard scan profile",
				Scanners:    []string{"code-packages", "code-security", "code-quality", "devops"},
			},
		},
		Scanners: make(map[string]Scanner),
	}
	return cfg
}

// Load reads configuration from multiple sources and merges them
// Priority: defaults < main config < user config (~/.zero/config.json)
func Load() (*Config, error) {
	cfg := &Config{
		Settings: Settings{
			DefaultProfile:        "all-quick",
			StoragePath:           ".zero",
			ParallelRepos:         8,
			ParallelScanners:      4,
			ScannerTimeoutSeconds: 300,
			CacheTTLHours:         24,
		},
		Profiles: make(map[string]Profile),
		Scanners: make(map[string]Scanner),
	}

	// 1. Load scanner defaults
	if err := loadScannerDefaults(cfg); err != nil {
		// Non-fatal: continue without defaults but log warning
		log.Printf("warning: failed to load scanner defaults: %v", err)
	}

	// 2. Load main config
	mainConfigPath := findMainConfig()
	if mainConfigPath != "" {
		if err := loadAndMerge(cfg, mainConfigPath); err != nil {
			return nil, fmt.Errorf("loading main config: %w", err)
		}
	}

	// 3. Load user config overrides
	userConfigPath := findUserConfig()
	if userConfigPath != "" {
		if err := loadAndMerge(cfg, userConfigPath); err != nil {
			// Non-fatal: continue without user overrides but log warning
			log.Printf("warning: failed to load user config overrides from %s: %v", userConfigPath, err)
		}
	}

	return cfg, nil
}

// loadScannerDefaults loads scanner feature defaults from config/defaults/scanners.json
func loadScannerDefaults(cfg *Config) error {
	candidates := []string{
		"config/defaults/scanners.json",
		filepath.Join(execDir(), "config/defaults/scanners.json"),
	}

	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		// First unmarshal into raw map to filter out non-scanner entries (like _docs)
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("parsing scanner defaults: %w", err)
		}

		// Remove docs field before parsing scanners
		delete(raw, "_docs")

		// Now parse each scanner
		scanners := make(map[string]Scanner)
		for name, rawScanner := range raw {
			var scanner Scanner
			if err := json.Unmarshal(rawScanner, &scanner); err != nil {
				log.Printf("warning: skipping invalid scanner %q: %v", name, err)
				continue
			}
			scanners[name] = scanner
		}

		cfg.Scanners = scanners
		return nil
	}

	return fmt.Errorf("scanner defaults not found")
}

// findMainConfig looks for zero.config.json
func findMainConfig() string {
	candidates := []string{
		"config/zero.config.json",
		filepath.Join(execDir(), "config/zero.config.json"),
		"zero.config.json",
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

// findUserConfig looks for user config in ~/.zero/config.json
func findUserConfig() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	path := filepath.Join(home, ".zero", "config.json")
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return ""
}

// execDir returns the directory containing the executable
func execDir() string {
	exePath, _ := os.Executable()
	return filepath.Dir(exePath)
}

// loadAndMerge loads a config file and merges it into the existing config
func loadAndMerge(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var overlay Config
	if err := json.Unmarshal(data, &overlay); err != nil {
		return fmt.Errorf("parsing config %s: %w", path, err)
	}

	// Merge settings (overlay wins)
	if overlay.Settings.DefaultProfile != "" {
		cfg.Settings.DefaultProfile = overlay.Settings.DefaultProfile
	}
	if overlay.Settings.StoragePath != "" {
		cfg.Settings.StoragePath = overlay.Settings.StoragePath
	}
	if overlay.Settings.ParallelRepos != 0 {
		cfg.Settings.ParallelRepos = overlay.Settings.ParallelRepos
	}
	if overlay.Settings.ParallelScanners != 0 {
		cfg.Settings.ParallelScanners = overlay.Settings.ParallelScanners
	}
	if overlay.Settings.ScannerTimeoutSeconds != 0 {
		cfg.Settings.ScannerTimeoutSeconds = overlay.Settings.ScannerTimeoutSeconds
	}
	if overlay.Settings.CacheTTLHours != 0 {
		cfg.Settings.CacheTTLHours = overlay.Settings.CacheTTLHours
	}

	// Merge profiles (overlay wins for each profile)
	for name, profile := range overlay.Profiles {
		cfg.Profiles[name] = profile
	}

	// Merge scanners (overlay wins for each scanner)
	for name, scanner := range overlay.Scanners {
		cfg.Scanners[name] = scanner
	}

	return nil
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

// GetProfile returns profile configuration by name
func (c *Config) GetProfile(name string) (*Profile, bool) {
	p, ok := c.Profiles[name]
	if !ok {
		return nil, false
	}
	return &p, true
}

// GetScanner returns scanner configuration by name
func (c *Config) GetScanner(name string) (*Scanner, bool) {
	s, ok := c.Scanners[name]
	if !ok {
		return nil, false
	}
	return &s, true
}

// GetScannerFeatures returns features for a scanner, with profile overrides applied
func (c *Config) GetScannerFeatures(scanner, profile string) map[string]interface{} {
	s, ok := c.Scanners[scanner]
	if !ok {
		return nil
	}

	// Start with default features
	features := make(map[string]interface{})
	for k, v := range s.Features {
		features[k] = v
	}

	// Apply profile overrides if any
	if profile != "" {
		p, ok := c.Profiles[profile]
		if ok && p.FeatureOverrides != nil {
			if overrides, ok := p.FeatureOverrides[scanner]; ok {
				mergeFeatures(features, overrides)
			}
		}
	}

	return features
}

// mergeFeatures merges override features into base features
func mergeFeatures(base, override map[string]interface{}) {
	for k, v := range override {
		if baseMap, ok := base[k].(map[string]interface{}); ok {
			if overrideMap, ok := v.(map[string]interface{}); ok {
				mergeFeatures(baseMap, overrideMap)
				continue
			}
		}
		base[k] = v
	}
}

// GetScannerTimeout returns the timeout for a scanner (in seconds)
func (c *Config) GetScannerTimeout(scanner string) int {
	return c.Settings.ScannerTimeoutSeconds
}

// ZeroHome returns the storage path
func (c *Config) ZeroHome() string {
	return c.Settings.StoragePath
}

// SlowScanners returns scanners that are known to be slow
func (c *Config) SlowScanners() []string {
	return []string{
		"code-packages",
		"code-security",
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
