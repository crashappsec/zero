// Package techid provides the consolidated technology identification super scanner
// This file manages semgrep rule generation, caching, and refresh
package techid

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/crashappsec/zero/pkg/core/rag"
)

// RuleManager handles semgrep rule lifecycle
type RuleManager struct {
	ragPath   string
	cachePath string
	ttl       time.Duration
	onStatus  func(message string)
}

// RuleManagerConfig configures the rule manager
type RuleManagerConfig struct {
	RAGPath   string        // Path to RAG directory
	CachePath string        // Path to cache directory
	TTL       time.Duration // Cache TTL (default 24h)
	OnStatus  func(message string) // Callback for status updates
}

// NewRuleManager creates a new rule manager
func NewRuleManager(cfg RuleManagerConfig) *RuleManager {
	if cfg.RAGPath == "" {
		cfg.RAGPath = rag.FindRAGPath()
	}
	if cfg.CachePath == "" {
		// Default to .zero/cache/technology-identification/rules
		homeDir, _ := os.UserHomeDir()
		cfg.CachePath = filepath.Join(homeDir, ".zero", "cache", "technology-identification", "rules")
	}
	if cfg.TTL == 0 {
		cfg.TTL = 24 * time.Hour
	}
	if cfg.OnStatus == nil {
		cfg.OnStatus = func(string) {} // No-op
	}
	return &RuleManager{
		ragPath:   cfg.RAGPath,
		cachePath: cfg.CachePath,
		ttl:       cfg.TTL,
		onStatus:  cfg.OnStatus,
	}
}

// RuleRefreshResult contains the result of a rule refresh
type RuleRefreshResult struct {
	Refreshed    bool
	TotalRules   int
	TechRules    int
	SecretRules  int
	AIMLRules    int
	CachePath    string
	FromCache    bool
	Error        error
}

// RefreshRules checks if rules need refreshing and regenerates them if needed
func (rm *RuleManager) RefreshRules(ctx context.Context, force bool) *RuleRefreshResult {
	result := &RuleRefreshResult{
		CachePath: rm.cachePath,
	}

	// Check if we need to refresh
	needsRefresh := force || rm.needsRefresh()

	if !needsRefresh {
		rm.onStatus("Using cached technology detection patterns")
		result.FromCache = true
		result.TotalRules = rm.countCachedRules()
		return result
	}

	rm.onStatus("Refreshing technology detection patterns...")

	// Ensure cache directory exists
	if err := os.MkdirAll(rm.cachePath, 0755); err != nil {
		result.Error = fmt.Errorf("failed to create cache directory: %w", err)
		return result
	}

	// Convert RAG patterns to semgrep rules
	rm.onStatus("Converting RAG patterns to semgrep rules...")
	convResult, err := ConvertRAGToSemgrep(rm.ragPath, rm.cachePath)
	if err != nil {
		result.Error = fmt.Errorf("failed to convert RAG patterns: %w", err)
		return result
	}

	result.Refreshed = true
	result.TotalRules = convResult.TotalRules
	result.TechRules = len(convResult.TechDiscovery.Rules)
	result.SecretRules = len(convResult.Secrets.Rules)
	result.AIMLRules = len(convResult.AIML.Rules)

	// Write timestamp file
	rm.writeTimestamp()

	rm.onStatus(fmt.Sprintf("Rules refreshed (%d patterns loaded)", result.TotalRules))
	return result
}

// needsRefresh checks if the cache needs to be refreshed
func (rm *RuleManager) needsRefresh() bool {
	// Check if cache exists
	timestampFile := filepath.Join(rm.cachePath, ".timestamp")
	info, err := os.Stat(timestampFile)
	if err != nil {
		return true // No timestamp = needs refresh
	}

	// Check TTL
	if time.Since(info.ModTime()) > rm.ttl {
		return true
	}

	// Check if any RAG files are newer than the cache
	ragModTime := rm.getLatestRAGModTime()
	if ragModTime.After(info.ModTime()) {
		return true
	}

	return false
}

// getLatestRAGModTime returns the latest modification time of RAG pattern files
func (rm *RuleManager) getLatestRAGModTime() time.Time {
	var latest time.Time

	techIDPath := filepath.Join(rm.ragPath, "technology-identification")
	filepath.Walk(techIDPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.Name() == "patterns.md" {
			if info.ModTime().After(latest) {
				latest = info.ModTime()
			}
		}
		return nil
	})

	return latest
}

// writeTimestamp writes a timestamp file to track cache freshness
func (rm *RuleManager) writeTimestamp() error {
	timestampFile := filepath.Join(rm.cachePath, ".timestamp")
	return os.WriteFile(timestampFile, []byte(time.Now().Format(time.RFC3339)), 0644)
}

// countCachedRules counts the rules in the cache
func (rm *RuleManager) countCachedRules() int {
	// Quick estimate by checking if files exist
	count := 0
	files := []string{"tech-discovery.yaml", "secrets.yaml", "ai-ml.yaml"}
	for _, f := range files {
		if _, err := os.Stat(filepath.Join(rm.cachePath, f)); err == nil {
			count++ // Just count files for quick estimate
		}
	}
	return count * 50 // Rough estimate of rules per file
}

// GetRulePaths returns paths to all generated rule files
func (rm *RuleManager) GetRulePaths() []string {
	var paths []string
	files := []string{"tech-discovery.yaml", "secrets.yaml", "ai-ml.yaml"}
	for _, f := range files {
		path := filepath.Join(rm.cachePath, f)
		if _, err := os.Stat(path); err == nil {
			paths = append(paths, path)
		}
	}
	return paths
}

// GetCachePath returns the cache directory path
func (rm *RuleManager) GetCachePath() string {
	return rm.cachePath
}

// ClearCache removes all cached rules
func (rm *RuleManager) ClearCache() error {
	return os.RemoveAll(rm.cachePath)
}

// DefaultRuleManager is the default rule manager instance
var DefaultRuleManager *RuleManager

// InitDefaultRuleManager initializes the default rule manager
func InitDefaultRuleManager(onStatus func(string)) {
	DefaultRuleManager = NewRuleManager(RuleManagerConfig{
		OnStatus: onStatus,
	})
}

// RefreshDefaultRules refreshes rules using the default manager
func RefreshDefaultRules(ctx context.Context, force bool, onStatus func(string)) *RuleRefreshResult {
	if DefaultRuleManager == nil {
		InitDefaultRuleManager(onStatus)
	}
	return DefaultRuleManager.RefreshRules(ctx, force)
}
