package rules

import (
	"os"
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
	if !cfg.GeneratedRules.Enabled {
		t.Error("Expected generated rules to be enabled by default")
	}
}

func TestNewManagerWithConfig(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := feeds.RuleConfig{
		GeneratedRules: feeds.RuleSourceConfig{
			Enabled:   false,
			Frequency: feeds.FreqNever,
		},
		CommunityRules: feeds.RuleSourceConfig{
			Enabled: false,
		},
	}

	m := NewManagerWithConfig(tmpDir, cfg)

	got := m.GetConfig()
	if got.GeneratedRules.Enabled {
		t.Error("Expected generated rules to be disabled")
	}
}

func TestManagerStatusPath(t *testing.T) {
	m := NewManager("/tmp/zero")

	genPath := m.statusPath("generated")
	if !filepath.IsAbs(genPath) {
		t.Error("Expected absolute path")
	}
	if !contains(genPath, "generated-status.json") {
		t.Errorf("statusPath(generated) = %s, expected to contain 'generated-status.json'", genPath)
	}

	commPath := m.statusPath("community")
	if !contains(commPath, "community-status.json") {
		t.Errorf("statusPath(community) = %s, expected to contain 'community-status.json'", commPath)
	}
}

func TestManagerSaveAndLoadStatus(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	status := &RuleStatus{
		Type:       "generated",
		RuleCount:  42,
		SourceHash: "abc123",
	}

	// Save
	if err := m.saveStatus("generated", status); err != nil {
		t.Fatalf("saveStatus() error = %v", err)
	}

	// Load
	loaded, err := m.loadStatus("generated")
	if err != nil {
		t.Fatalf("loadStatus() error = %v", err)
	}

	if loaded.Type != "generated" {
		t.Errorf("Type = %s, want generated", loaded.Type)
	}
	if loaded.RuleCount != 42 {
		t.Errorf("RuleCount = %d, want 42", loaded.RuleCount)
	}
	if loaded.SourceHash != "abc123" {
		t.Errorf("SourceHash = %s, want abc123", loaded.SourceHash)
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
	m.saveStatus("generated", &RuleStatus{Type: "generated", RuleCount: 10})
	m.saveStatus("community", &RuleStatus{Type: "community", RuleCount: 100})

	statuses := m.GetStatus()

	if len(statuses) != 2 {
		t.Errorf("GetStatus() returned %d items, want 2", len(statuses))
	}

	if s, ok := statuses["generated"]; !ok || s.RuleCount != 10 {
		t.Error("Expected generated status with 10 rules")
	}

	if s, ok := statuses["community"]; !ok || s.RuleCount != 100 {
		t.Error("Expected community status with 100 rules")
	}
}

func TestManagerSetConfig(t *testing.T) {
	m := NewManager("/tmp/test")

	newCfg := feeds.RuleConfig{
		GeneratedRules: feeds.RuleSourceConfig{
			Enabled:   false,
			OutputDir: "custom/output",
		},
	}

	m.SetConfig(newCfg)

	got := m.GetConfig()
	if got.GeneratedRules.Enabled {
		t.Error("Expected generated rules to be disabled")
	}
	if got.GeneratedRules.OutputDir != "custom/output" {
		t.Errorf("OutputDir = %s, want custom/output", got.GeneratedRules.OutputDir)
	}
}

func TestManagerRefreshRulesDisabled(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := feeds.RuleConfig{
		GeneratedRules: feeds.RuleSourceConfig{Enabled: false},
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

func TestManagerRefreshGeneratedRulesSkipped(t *testing.T) {
	tmpDir := t.TempDir()

	// Create RAG directory with a pattern file
	ragDir := filepath.Join(tmpDir, "..", "rag", "technology-identification")
	os.MkdirAll(ragDir, 0755)
	os.WriteFile(filepath.Join(ragDir, "test.md"), []byte("# Test Pattern"), 0644)

	m := NewManager(tmpDir)

	// First refresh
	results, _ := m.RefreshRules(false)

	// Find generated result
	var genResult *GenerateResult
	for i := range results {
		if results[i].Type == "generated" {
			genResult = &results[i]
			break
		}
	}

	if genResult == nil {
		t.Skip("No generated rules result (RAG loader may not find patterns)")
		return
	}

	// Second refresh should skip (hash unchanged)
	results2, _ := m.RefreshRules(false)
	for _, r := range results2 {
		if r.Type == "generated" && !r.Skipped {
			// Only fail if we actually had rules the first time
			if genResult.RuleCount > 0 {
				t.Error("Expected second refresh to be skipped")
			}
		}
	}
}

func TestNormalizeSeverity(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"critical", "ERROR"},
		{"high", "ERROR"},
		{"medium", "WARNING"},
		{"low", "INFO"},
		{"info", "INFO"},
		{"unknown", "INFO"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := normalizeSeverity(tt.input); got != tt.expected {
				t.Errorf("normalizeSeverity(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
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
