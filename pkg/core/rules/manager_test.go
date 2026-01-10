package rules

import (
	"path/filepath"
	"testing"

	"github.com/crashappsec/zero/pkg/core/feeds"
)

func TestNewManager(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	if m == nil {
		t.Fatal("Expected non-nil manager")
	}

	if m.zeroHome != tmpDir {
		t.Errorf("zeroHome = %s, want %s", m.zeroHome, tmpDir)
	}

	// Check default config
	cfg := m.GetConfig()
	if !cfg.CommunityRules.Enabled {
		t.Error("Expected community rules to be enabled by default")
	}
}

func TestNewManagerWithConfig(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := feeds.RuleConfig{
		CommunityRules: feeds.RuleSourceConfig{
			Enabled: false,
		},
	}

	m := NewManagerWithConfig(tmpDir, cfg)

	got := m.GetConfig()
	if got.CommunityRules.Enabled {
		t.Error("Expected community rules to be disabled")
	}
}

func TestManagerStatusPath(t *testing.T) {
	m := NewManager("/tmp/zero")

	commPath := m.statusPath("community")
	if !filepath.IsAbs(commPath) {
		t.Error("Expected absolute path")
	}
	if !contains(commPath, "community-status.json") {
		t.Errorf("statusPath(community) = %s, expected to contain 'community-status.json'", commPath)
	}
}

func TestManagerSaveAndLoadStatus(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	status := &RuleStatus{
		Type:      "community",
		RuleCount: 42,
	}

	// Save
	if err := m.saveStatus("community", status); err != nil {
		t.Fatalf("saveStatus() error = %v", err)
	}

	// Load
	loaded, err := m.loadStatus("community")
	if err != nil {
		t.Fatalf("loadStatus() error = %v", err)
	}

	if loaded.Type != "community" {
		t.Errorf("Type = %s, want community", loaded.Type)
	}
	if loaded.RuleCount != 42 {
		t.Errorf("RuleCount = %d, want 42", loaded.RuleCount)
	}
}

func TestManagerLoadStatusNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	_, err := m.loadStatus("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent status")
	}
}

func TestManagerGetStatus(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	// Save some status
	m.saveStatus("community", &RuleStatus{Type: "community", RuleCount: 100})

	statuses := m.GetStatus()

	if len(statuses) != 1 {
		t.Errorf("GetStatus() returned %d items, want 1", len(statuses))
	}

	if s, ok := statuses["community"]; !ok || s.RuleCount != 100 {
		t.Error("Expected community status with 100 rules")
	}
}

func TestManagerSetConfig(t *testing.T) {
	m := NewManager("/tmp/test")

	newCfg := feeds.RuleConfig{
		CommunityRules: feeds.RuleSourceConfig{
			Enabled:   false,
			OutputDir: "custom/output",
		},
	}

	m.SetConfig(newCfg)

	got := m.GetConfig()
	if got.CommunityRules.Enabled {
		t.Error("Expected community rules to be disabled")
	}
	if got.CommunityRules.OutputDir != "custom/output" {
		t.Errorf("OutputDir = %s, want custom/output", got.CommunityRules.OutputDir)
	}
}

func TestManagerRefreshRulesDisabled(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := feeds.RuleConfig{
		CommunityRules: feeds.RuleSourceConfig{Enabled: false},
	}
	m := NewManagerWithConfig(tmpDir, cfg)

	results, err := m.RefreshRules(false)
	if err != nil {
		t.Fatalf("RefreshRules() error = %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results when all disabled, got %d", len(results))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
