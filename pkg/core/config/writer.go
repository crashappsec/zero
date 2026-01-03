// Package config handles Zero configuration loading and management
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var configMutex sync.Mutex

// Save writes the configuration to the config file
func (c *Config) Save() error {
	configMutex.Lock()
	defer configMutex.Unlock()

	configPath := findConfigFile()
	if configPath == "" {
		// Default to config/zero.config.json
		configPath = "config/zero.config.json"
	}

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	// Marshal with indentation for readability
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	// Write atomically using temp file
	tmpPath := configPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("writing temp config: %w", err)
	}

	if err := os.Rename(tmpPath, configPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("replacing config: %w", err)
	}

	return nil
}

// UpdateSettings updates global settings and saves
func (c *Config) UpdateSettings(settings Settings) error {
	c.Settings = settings
	return c.Save()
}

// CreateProfile creates a new profile
func (c *Config) CreateProfile(name string, profile Profile) error {
	if c.Profiles == nil {
		c.Profiles = make(map[string]Profile)
	}

	if _, exists := c.Profiles[name]; exists {
		return fmt.Errorf("profile already exists: %s", name)
	}

	c.Profiles[name] = profile
	return c.Save()
}

// UpdateProfile updates an existing profile
func (c *Config) UpdateProfile(name string, profile Profile) error {
	if c.Profiles == nil {
		return fmt.Errorf("no profiles configured")
	}

	if _, exists := c.Profiles[name]; !exists {
		return fmt.Errorf("profile not found: %s", name)
	}

	c.Profiles[name] = profile
	return c.Save()
}

// DeleteProfile removes a profile
func (c *Config) DeleteProfile(name string) error {
	if c.Profiles == nil {
		return fmt.Errorf("no profiles configured")
	}

	// Protect built-in profiles
	builtIn := map[string]bool{
		"all-quick":    true,
		"all-complete": true,
	}
	if builtIn[name] {
		return fmt.Errorf("cannot delete built-in profile: %s", name)
	}

	if _, exists := c.Profiles[name]; !exists {
		return fmt.Errorf("profile not found: %s", name)
	}

	delete(c.Profiles, name)
	return c.Save()
}

// UpdateScanner updates scanner configuration
func (c *Config) UpdateScanner(name string, scanner Scanner) error {
	if c.Scanners == nil {
		c.Scanners = make(map[string]Scanner)
	}

	c.Scanners[name] = scanner
	return c.Save()
}

// GetConfigPath returns the current config file path
func GetConfigPath() string {
	return findConfigFile()
}

// Export returns the full config as JSON bytes
func (c *Config) Export() ([]byte, error) {
	return json.MarshalIndent(c, "", "  ")
}

// Import loads config from JSON bytes and saves
func Import(data []byte) (*Config, error) {
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing imported config: %w", err)
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if err := cfg.Save(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Check settings
	if c.Settings.ParallelRepos < 0 {
		return fmt.Errorf("parallel_repos must be non-negative")
	}
	if c.Settings.ParallelScanners < 0 {
		return fmt.Errorf("parallel_scanners must be non-negative")
	}
	if c.Settings.ScannerTimeoutSeconds < 0 {
		return fmt.Errorf("scanner_timeout_seconds must be non-negative")
	}

	// Check profiles reference valid scanners
	validScanners := map[string]bool{
		"sbom":                  true,
		"packages":              true,
		"code-crypto":           true,
		"code-security":         true,
		"code-quality":          true,
		"devops":                true,
		"tech-id":               true,
		"code-ownership":        true,
		"developer-experience":  true,
	}

	for profileName, profile := range c.Profiles {
		for _, scanner := range profile.Scanners {
			if !validScanners[scanner] {
				return fmt.Errorf("profile %s references unknown scanner: %s", profileName, scanner)
			}
		}
	}

	return nil
}

// Reload re-reads the config from disk
func Reload() (*Config, error) {
	configMutex.Lock()
	defer configMutex.Unlock()
	return Load()
}
