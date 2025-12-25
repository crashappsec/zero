// Package rules provides Semgrep rule generation and management
package rules

import (
	"crypto/sha256"
	"encoding/hex"
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

	// Check generated rules (from RAG)
	if m.config.GeneratedRules.Enabled {
		result := m.refreshGeneratedRules(force)
		results = append(results, result)
	}

	// Check community rules
	if m.config.CommunityRules.Enabled {
		result := m.refreshCommunityRules(force)
		results = append(results, result)
	}

	return results, nil
}

func (m *Manager) refreshGeneratedRules(force bool) GenerateResult {
	start := time.Now()
	result := GenerateResult{Type: "generated"}

	// Check if RAG files have changed
	currentHash, err := m.computeRAGHash()
	if err != nil {
		result.Error = fmt.Sprintf("computing RAG hash: %v", err)
		return result
	}

	// Load previous status
	status, _ := m.loadStatus("generated")

	// Skip if not forced and hash hasn't changed
	if !force && status != nil && status.SourceHash == currentHash {
		result.Skipped = true
		result.Reason = "RAG files unchanged"
		return result
	}

	// Generate rules from RAG
	files, ruleCount, err := m.generateFromRAG()
	if err != nil {
		result.Error = err.Error()
		return result
	}

	result.Success = true
	result.Duration = time.Since(start)
	result.RuleCount = ruleCount
	result.Files = files

	// Save status
	m.saveStatus("generated", &RuleStatus{
		Type:         "generated",
		LastGenerate: time.Now(),
		RuleCount:    ruleCount,
		SourceHash:   currentHash,
		OutputFiles:  files,
	})

	return result
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
	m.saveStatus("community", &RuleStatus{
		Type:         "community",
		LastGenerate: time.Now(),
		RuleCount:    result.RuleCount,
	})

	return result
}

func (m *Manager) computeRAGHash() (string, error) {
	h := sha256.New()

	// Walk relevant RAG directories and hash file contents
	ragDirs := []string{
		"technology-identification",
		"devops",
		"code-quality",
	}

	for _, dir := range ragDirs {
		fullPath := filepath.Join(m.ragLoader.RAGPath(), dir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			continue
		}

		err := filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip errors
			}
			if info.IsDir() {
				return nil
			}
			if filepath.Ext(path) != ".md" && filepath.Ext(path) != ".json" {
				return nil
			}

			data, err := os.ReadFile(path)
			if err != nil {
				return nil // Skip unreadable files
			}

			h.Write([]byte(path))
			h.Write(data)
			return nil
		})

		if err != nil {
			return "", err
		}
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func (m *Manager) generateFromRAG() ([]string, int, error) {
	outputDir := filepath.Join(m.zeroHome, m.config.GeneratedRules.OutputDir)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, 0, fmt.Errorf("creating output dir: %w", err)
	}

	var files []string
	totalRules := 0

	// Load patterns from RAG categories
	categories := []string{
		"technology-identification",
		"secrets",
		"devops",
		"devops-security",
		"code-security",
		"code-quality",
		"architecture",
	}

	for _, category := range categories {
		if !m.ragLoader.HasCategory(category) {
			continue
		}

		result, err := m.ragLoader.LoadCategory(category)
		if err != nil {
			continue // Skip categories with errors
		}

		// Convert patterns to Semgrep rules
		rules := m.patternsToSemgrepRules(result.PatternSets)
		if len(rules) == 0 {
			continue
		}

		// Write rules file
		filename := fmt.Sprintf("%s.yaml", category)
		outputPath := filepath.Join(outputDir, filename)

		if err := m.writeRulesFile(outputPath, rules); err != nil {
			continue
		}

		files = append(files, outputPath)
		totalRules += len(rules)
	}

	return files, totalRules, nil
}

func (m *Manager) patternsToSemgrepRules(patternSets []rag.PatternSet) []rag.SemgrepRule {
	var rules []rag.SemgrepRule

	for _, ps := range patternSets {
		for _, p := range ps.Patterns {
			if p.Type != "semgrep" && p.Type != "regex" {
				continue
			}

			rule := rag.SemgrepRule{
				ID:        p.ID,
				Message:   p.Message,
				Severity:  normalizeSeverity(p.Severity),
				Languages: p.Languages,
			}

			if len(rule.Languages) == 0 {
				rule.Languages = []string{"generic"}
			}

			if p.Type == "regex" {
				rule.Pattern = "" // Use pattern-regex instead
				rule.Patterns = []rag.SemgrepPattern{
					{PatternRegex: p.Pattern},
				}
			} else {
				rule.Pattern = p.Pattern
			}

			rules = append(rules, rule)
		}
	}

	return rules
}

func normalizeSeverity(sev string) string {
	switch sev {
	case "critical", "high":
		return "ERROR"
	case "medium":
		return "WARNING"
	case "low", "info":
		return "INFO"
	default:
		return "INFO"
	}
}

func (m *Manager) writeRulesFile(path string, rules []rag.SemgrepRule) error {
	config := rag.SemgrepConfig{Rules: rules}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling rules: %w", err)
	}

	return os.WriteFile(path, data, 0644)
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

	return os.WriteFile(path, data, 0644)
}

// GetStatus returns the status of all rule sources
func (m *Manager) GetStatus() map[string]*RuleStatus {
	statuses := make(map[string]*RuleStatus)

	if status, err := m.loadStatus("generated"); err == nil {
		statuses["generated"] = status
	}

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
