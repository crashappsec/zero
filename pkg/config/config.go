// Package config handles Zero configuration loading and management
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the main Zero configuration
type Config struct {
	Version  string            `json:"version"`
	Profiles map[string]Profile `json:"profiles"`
	Scanners map[string]Scanner `json:"scanners"`
	Settings Settings           `json:"settings"`
}

// Profile defines a scanning profile with its scanners
type Profile struct {
	Description string   `json:"description"`
	Scanners    []string `json:"scanners"`
}

// Scanner defines a scanner configuration
type Scanner struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Command     string `json:"command"`
	Timeout     int    `json:"timeout"` // seconds
}

// Settings contains global settings
type Settings struct {
	ParallelJobs int    `json:"parallel_jobs"`
	ZeroHome     string `json:"zero_home"`
	CloneDepth   int    `json:"clone_depth"`
}

// DefaultConfig returns sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Version: "1.0",
		Settings: Settings{
			ParallelJobs: 4,
			ZeroHome:     ".zero",
			CloneDepth:   1,
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
	if cfg.Settings.ParallelJobs == 0 {
		cfg.Settings.ParallelJobs = 4
	}
	if cfg.Settings.CloneDepth == 0 {
		cfg.Settings.CloneDepth = 1
	}
	if cfg.Settings.ZeroHome == "" {
		cfg.Settings.ZeroHome = ".zero"
	}

	return &cfg, nil
}

// findConfigFile looks for zero.config.json in standard locations
func findConfigFile() string {
	candidates := []string{
		"utils/zero/config/zero.config.json",
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
		return nil, fmt.Errorf("unknown profile: %s", profile)
	}
	return p.Scanners, nil
}

// GetScannerTimeout returns the timeout for a scanner (or default)
func (c *Config) GetScannerTimeout(scanner string) int {
	if s, ok := c.Scanners[scanner]; ok && s.Timeout > 0 {
		return s.Timeout
	}
	return 120 // default 2 minutes
}
