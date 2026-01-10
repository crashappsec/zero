// Package rules provides Semgrep rule generation and management
package rules

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/crashappsec/zero/pkg/core/feeds"
	"github.com/crashappsec/zero/pkg/core/rag"
)

// Manager handles rule generation and updates
type Manager struct {
	zeroHome  string
	ragLoader *rag.RAGLoader
	config    feeds.RuleConfig
}

// NewManager creates a new rule manager
func NewManager(zeroHome string) *Manager {
	ragPath := filepath.Join(zeroHome, "..", "rag")
	if envPath := os.Getenv("ZERO_RAG_PATH"); envPath != "" {
		ragPath = envPath
	}

	return &Manager{
		zeroHome:  zeroHome,
		ragLoader: rag.NewLoader(ragPath),
		config:    feeds.DefaultRuleConfig(),
	}
}

// NewManagerWithConfig creates a manager with custom config
func NewManagerWithConfig(zeroHome string, config feeds.RuleConfig) *Manager {
	m := NewManager(zeroHome)
	m.config = config
	return m
}

// RuleStatus tracks the status of generated rules
type RuleStatus struct {
	Type         string    `json:"type"` // "generated" or "community"
	LastGenerate time.Time `json:"last_generate"`
	RuleCount    int       `json:"rule_count"`
	SourceHash   string    `json:"source_hash"` // Hash of source files
	OutputFiles  []string  `json:"output_files"`
	Error        string    `json:"error,omitempty"`
}

// GenerateResult holds the result of rule generation
type GenerateResult struct {
	Type      string        `json:"type"`
	Success   bool          `json:"success"`
	Duration  time.Duration `json:"duration"`
	RuleCount int           `json:"rule_count"`
	Files     []string      `json:"files"`
	Error     string        `json:"error,omitempty"`
	Skipped   bool          `json:"skipped,omitempty"`
	Reason    string        `json:"reason,omitempty"`
}

// RefreshRules refreshes rules if needed based on configuration
func (m *Manager) RefreshRules(force bool) ([]GenerateResult, error) {
	var results []GenerateResult

	// Check community rules (Semgrep SAST rules)
	if m.config.CommunityRules.Enabled {
		result := m.refreshCommunityRules(force)
		results = append(results, result)
	}

	return results, nil
}

func (m *Manager) refreshCommunityRules(force bool) GenerateResult {
	start := time.Now()
	result := GenerateResult{Type: "community"}

	// Load previous status
	status, _ := m.loadStatus("community")

	// Check if we should sync based on frequency
	if !force && status != nil {
		if !m.config.CommunityRules.Frequency.ShouldSync(status.LastGenerate) {
			result.Skipped = true
			result.Reason = "not due for sync"
			return result
		}
	}

	// TODO: Actually download community rules from Semgrep registry
	// For now, just mark as successful if we have cached rules
	cacheDir := filepath.Join(m.zeroHome, "cache", "semgrep", "community")
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		// Create placeholder
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			result.Error = fmt.Sprintf("creating cache dir: %v", err)
			return result
		}
	}

	result.Success = true
	result.Duration = time.Since(start)
	result.RuleCount = 0 // Would be populated from download

	// Save status
	_ = m.saveStatus("community", &RuleStatus{
		Type:         "community",
		LastGenerate: time.Now(),
		RuleCount:    result.RuleCount,
	})

	return result
}

func (m *Manager) statusPath(ruleType string) string {
	return filepath.Join(m.zeroHome, "cache", "rules", ruleType+"-status.json")
}

func (m *Manager) loadStatus(ruleType string) (*RuleStatus, error) {
	data, err := os.ReadFile(m.statusPath(ruleType))
	if err != nil {
		return nil, err
	}

	var status RuleStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, err
	}

	return &status, nil
}

func (m *Manager) saveStatus(ruleType string, status *RuleStatus) error {
	path := m.statusPath(ruleType)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// GetStatus returns the status of all rule sources
func (m *Manager) GetStatus() map[string]*RuleStatus {
	statuses := make(map[string]*RuleStatus)

	if status, err := m.loadStatus("community"); err == nil {
		statuses["community"] = status
	}

	return statuses
}

// GetConfig returns the current configuration
func (m *Manager) GetConfig() feeds.RuleConfig {
	return m.config
}

// SetConfig updates the configuration
func (m *Manager) SetConfig(config feeds.RuleConfig) {
	m.config = config
}
