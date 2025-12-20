package freshness

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	m := NewManager("/tmp/test-zero")

	if m.zeroHome != "/tmp/test-zero" {
		t.Errorf("zeroHome = %s, want /tmp/test-zero", m.zeroHome)
	}

	// Check default thresholds
	th := m.GetThresholds()
	if th.FreshMaxHours != 24 {
		t.Errorf("FreshMaxHours = %d, want 24", th.FreshMaxHours)
	}
}

func TestNewManagerWithThresholds(t *testing.T) {
	th := Thresholds{
		FreshMaxHours:    12,
		StaleMaxDays:     3,
		VeryStaleMaxDays: 14,
		ExpiredMaxDays:   30,
	}

	m := NewManagerWithThresholds("/tmp/test-zero", th)

	got := m.GetThresholds()
	if got.FreshMaxHours != 12 {
		t.Errorf("FreshMaxHours = %d, want 12", got.FreshMaxHours)
	}
}

func TestManagerMetadataPath(t *testing.T) {
	m := NewManager("/tmp/test-zero")
	expected := "/tmp/test-zero/repos/owner/repo/freshness.json"

	if got := m.metadataPath("owner/repo"); got != expected {
		t.Errorf("metadataPath() = %s, want %s", got, expected)
	}
}

func TestManagerSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	// Create test metadata
	meta := &Metadata{
		Repository: "test/repo",
		LastScan:   time.Now(),
		LastCommit: "abc123",
		ScannerStatus: map[string]Status{
			"scanner1": {Scanner: "scanner1", Success: true},
		},
	}

	// Save
	if err := m.Save(meta); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file exists
	expectedPath := filepath.Join(tmpDir, "repos", "test/repo", "freshness.json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Error("Expected metadata file to exist")
	}

	// Load
	loaded, err := m.Load("test/repo")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.Repository != "test/repo" {
		t.Errorf("Repository = %s, want test/repo", loaded.Repository)
	}

	if loaded.LastCommit != "abc123" {
		t.Errorf("LastCommit = %s, want abc123", loaded.LastCommit)
	}

	if len(loaded.ScannerStatus) != 1 {
		t.Errorf("ScannerStatus has %d items, want 1", len(loaded.ScannerStatus))
	}
}

func TestManagerLoadNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	// Load non-existent repo - should return empty metadata, not error
	loaded, err := m.Load("nonexistent/repo")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.Repository != "nonexistent/repo" {
		t.Errorf("Repository = %s, want nonexistent/repo", loaded.Repository)
	}

	if loaded.ScannerStatus == nil {
		t.Error("Expected ScannerStatus to be initialized")
	}
}

func TestManagerCheck(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	// Check non-existent repo
	result, err := m.Check("nonexistent/repo")
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	if result.Level != LevelUnknown {
		t.Errorf("Level = %s, want %s for never-scanned repo", result.Level, LevelUnknown)
	}

	// Save some metadata and check again
	meta := &Metadata{
		Repository: "test/repo",
		LastScan:   time.Now().Add(-2 * 24 * time.Hour),
		LastCommit: "abc123",
	}
	m.Save(meta)

	result, err = m.Check("test/repo")
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	if result.Level != LevelStale {
		t.Errorf("Level = %s, want %s", result.Level, LevelStale)
	}
}

func TestManagerRecordScan(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	scanResults := []ScanResult{
		{Name: "scanner1", Success: true, Duration: 5 * time.Second},
		{Name: "scanner2", Success: false, Error: "failed", Duration: 2 * time.Second},
	}

	if err := m.RecordScan("test/repo", scanResults); err != nil {
		t.Fatalf("RecordScan() error = %v", err)
	}

	// Load and verify
	meta, err := m.Load("test/repo")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if meta.Repository != "test/repo" {
		t.Errorf("Repository = %s, want test/repo", meta.Repository)
	}

	if len(meta.ScannerStatus) != 2 {
		t.Errorf("ScannerStatus has %d items, want 2", len(meta.ScannerStatus))
	}

	if s, ok := meta.ScannerStatus["scanner1"]; !ok || !s.Success {
		t.Error("Expected scanner1 to be successful")
	}

	if s, ok := meta.ScannerStatus["scanner2"]; !ok || s.Success {
		t.Error("Expected scanner2 to be failed")
	}
}

func TestManagerShouldScan(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	// Never scanned - should scan
	shouldScan, reason, err := m.ShouldScan("new/repo", false)
	if err != nil {
		t.Fatalf("ShouldScan() error = %v", err)
	}
	if !shouldScan {
		t.Error("Expected ShouldScan to return true for never-scanned repo")
	}
	if reason != "never scanned" {
		t.Errorf("reason = %s, want 'never scanned'", reason)
	}

	// Scan the repo
	m.RecordScan("new/repo", []ScanResult{{Name: "test", Success: true}})

	// Just scanned - should not scan
	shouldScan, reason, err = m.ShouldScan("new/repo", false)
	if err != nil {
		t.Fatalf("ShouldScan() error = %v", err)
	}
	if shouldScan {
		t.Error("Expected ShouldScan to return false for just-scanned repo")
	}
	if reason != "data is fresh" {
		t.Errorf("reason = %s, want 'data is fresh'", reason)
	}
}

func TestManagerDelete(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	// Save metadata
	meta := &Metadata{Repository: "test/repo", LastScan: time.Now()}
	m.Save(meta)

	// Verify exists
	expectedPath := filepath.Join(tmpDir, "repos", "test/repo", "freshness.json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatal("Expected metadata file to exist before delete")
	}

	// Delete
	if err := m.Delete("test/repo"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify deleted
	if _, err := os.Stat(expectedPath); !os.IsNotExist(err) {
		t.Error("Expected metadata file to be deleted")
	}

	// Delete non-existent should not error
	if err := m.Delete("nonexistent/repo"); err != nil {
		t.Errorf("Delete() non-existent error = %v, want nil", err)
	}
}

func TestManagerSetThresholds(t *testing.T) {
	m := NewManager("/tmp/test")

	newTh := Thresholds{
		FreshMaxHours:    6,
		StaleMaxDays:     2,
		VeryStaleMaxDays: 7,
		ExpiredMaxDays:   14,
	}

	m.SetThresholds(newTh)

	got := m.GetThresholds()
	if got.FreshMaxHours != 6 {
		t.Errorf("FreshMaxHours = %d, want 6", got.FreshMaxHours)
	}
}

func TestManagerGetSummary(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	// Create repos directory structure
	reposDir := filepath.Join(tmpDir, "repos")

	// Create a couple of repos with different ages
	repos := []struct {
		name string
		age  time.Duration
	}{
		{"fresh/repo", 1 * time.Hour},
		{"stale/repo", 3 * 24 * time.Hour},
	}

	for _, r := range repos {
		repoDir := filepath.Join(reposDir, r.name)
		os.MkdirAll(repoDir, 0755)

		meta := &Metadata{
			Repository: r.name,
			LastScan:   time.Now().Add(-r.age),
		}
		m.Save(meta)
	}

	summary, err := m.GetSummary()
	if err != nil {
		t.Fatalf("GetSummary() error = %v", err)
	}

	if summary.Total != 2 {
		t.Errorf("Total = %d, want 2", summary.Total)
	}

	// At least one should need scan (the stale one)
	if summary.NeedScan < 1 {
		t.Errorf("NeedScan = %d, want >= 1", summary.NeedScan)
	}
}
