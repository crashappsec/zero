package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/crashappsec/zero/pkg/workflow/freshness"
)

func TestRefreshCommandFlagDefaults(t *testing.T) {
	// Verify default flag values
	if refreshForce {
		t.Error("Expected refreshForce default to be false")
	}
	if refreshAll {
		t.Error("Expected refreshAll default to be false")
	}
	if refreshParallel != 4 {
		t.Errorf("Expected refreshParallel default to be 4, got %d", refreshParallel)
	}
}

func TestFreshnessIntegration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a mock repo structure
	repoDir := filepath.Join(tmpDir, "repos", "test", "repo")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatalf("Failed to create repo dir: %v", err)
	}

	// Create freshness manager and test basic operations
	mgr := freshness.NewManager(tmpDir)

	// Record a scan
	scanResults := []freshness.ScanResult{
		{Name: "sbom", Success: true, Duration: 5 * time.Second},
		{Name: "code-security", Success: true, Duration: 10 * time.Second},
	}

	if err := mgr.RecordScan("test/repo", scanResults); err != nil {
		t.Fatalf("RecordScan failed: %v", err)
	}

	// Check freshness
	result, err := mgr.Check("test/repo")
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	if result.Level != freshness.LevelFresh {
		t.Errorf("Expected fresh level, got %s", result.Level)
	}

	// Should not need scan
	shouldScan, _, _ := mgr.ShouldScan("test/repo", false)
	if shouldScan {
		t.Error("Expected ShouldScan to return false for fresh repo")
	}
}

func TestStaleRepoDetection(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a stale metadata file (2 days old)
	repoDir := filepath.Join(tmpDir, "repos", "stale", "repo")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatalf("Failed to create repo dir: %v", err)
	}

	mgr := freshness.NewManager(tmpDir)

	// Save old metadata manually
	meta := &freshness.Metadata{
		Repository: "stale/repo",
		LastScan:   time.Now().Add(-2 * 24 * time.Hour), // 2 days ago
		ScannerStatus: map[string]freshness.Status{
			"test": {Success: true},
		},
	}

	if err := mgr.Save(meta); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Check freshness
	result, err := mgr.Check("stale/repo")
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	if result.Level != freshness.LevelStale {
		t.Errorf("Expected stale level, got %s", result.Level)
	}

	if !result.NeedsRefresh {
		t.Error("Expected NeedsRefresh to be true for stale repo")
	}
}

func TestFreshRepoShouldNotScan(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := freshness.NewManager(tmpDir)

	// Create a fresh repo
	repoDir := filepath.Join(tmpDir, "repos", "owner", "fresh-repo")
	os.MkdirAll(repoDir, 0755)

	meta := &freshness.Metadata{
		Repository:    "owner/fresh-repo",
		LastScan:      time.Now().Add(-1 * time.Hour), // 1 hour ago = fresh
		ScannerStatus: map[string]freshness.Status{},
	}
	mgr.Save(meta)

	// Fresh repo should not need scan
	shouldScan, reason, err := mgr.ShouldScan("owner/fresh-repo", false)
	if err != nil {
		t.Fatalf("ShouldScan failed: %v", err)
	}

	if shouldScan {
		t.Errorf("Expected fresh repo to not need scan, reason: %s", reason)
	}

	// Check level directly
	result, err := mgr.Check("owner/fresh-repo")
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	if result.Level != freshness.LevelFresh {
		t.Errorf("Expected LevelFresh, got %s", result.Level)
	}
}
