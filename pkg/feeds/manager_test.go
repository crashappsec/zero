package feeds

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

	if m.httpClient == nil {
		t.Error("Expected httpClient to be initialized")
	}

	if m.status == nil {
		t.Error("Expected status map to be initialized")
	}
}

func TestManagerStatusPath(t *testing.T) {
	m := NewManager("/tmp/test-zero")
	expected := "/tmp/test-zero/feed-status.json"

	if got := m.statusPath(); got != expected {
		t.Errorf("statusPath() = %s, want %s", got, expected)
	}
}

func TestManagerCacheDir(t *testing.T) {
	m := NewManager("/tmp/test-zero")
	expected := "/tmp/test-zero/cache/feeds"

	if got := m.cacheDir(); got != expected {
		t.Errorf("cacheDir() = %s, want %s", got, expected)
	}

	// Test with custom cache dir
	m.config.CacheDir = "/custom/cache"
	if got := m.cacheDir(); got != "/custom/cache" {
		t.Errorf("cacheDir() with custom = %s, want /custom/cache", got)
	}
}

func TestManagerSaveAndLoadStatus(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	// Add some status
	m.status[FeedSemgrepRules] = &FeedStatus{
		Type:        FeedSemgrepRules,
		Name:        "test",
		LastSync:    time.Now(),
		LastSuccess: time.Now(),
		ItemCount:   100,
	}

	// Save
	if err := m.SaveStatus(); err != nil {
		t.Fatalf("SaveStatus() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(m.statusPath()); os.IsNotExist(err) {
		t.Error("Expected status file to exist")
	}

	// Create new manager and load
	m2 := NewManager(tmpDir)
	if err := m2.LoadStatus(); err != nil {
		t.Fatalf("LoadStatus() error = %v", err)
	}

	status := m2.GetStatus(FeedSemgrepRules)
	if status == nil {
		t.Fatal("Expected status to be loaded")
	}

	if status.ItemCount != 100 {
		t.Errorf("ItemCount = %d, want 100", status.ItemCount)
	}
}

func TestManagerLoadStatusNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	// Should not error when file doesn't exist
	if err := m.LoadStatus(); err != nil {
		t.Errorf("LoadStatus() error = %v, want nil", err)
	}
}

func TestManagerShouldSync(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	// Never synced - should sync
	if !m.ShouldSync(FeedSemgrepRules) {
		t.Error("Expected ShouldSync to return true for never-synced feed")
	}

	// Add fresh status
	m.status[FeedSemgrepRules] = &FeedStatus{
		Type:     FeedSemgrepRules,
		LastSync: time.Now(),
	}

	// Weekly feed synced just now - should not sync
	if m.ShouldSync(FeedSemgrepRules) {
		t.Error("Expected ShouldSync to return false for fresh feed")
	}

	// Unknown feed - should not sync
	if m.ShouldSync(FeedType("unknown")) {
		t.Error("Expected ShouldSync to return false for unknown feed")
	}
}

func TestManagerSyncSemgrepRules(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rules := map[string]interface{}{
			"rules": []map[string]interface{}{
				{"id": "rule1", "pattern": "test"},
				{"id": "rule2", "pattern": "test2"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("ETag", "abc123")
		json.NewEncoder(w).Encode(rules)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	// Override feed URL to use mock server
	m.config.Feeds = []FeedConfig{
		{
			Type:      FeedSemgrepRules,
			Name:      "test",
			URL:       server.URL,
			Frequency: FreqAlways,
			Enabled:   true,
		},
	}

	ctx := context.Background()
	result := m.SyncFeed(ctx, FeedSemgrepRules)

	if !result.Success {
		t.Errorf("SyncFeed() success = false, error = %s", result.Error)
	}

	if result.ItemCount != 2 {
		t.Errorf("ItemCount = %d, want 2", result.ItemCount)
	}

	// Verify cache file exists
	cachePath := filepath.Join(tmpDir, "cache", "feeds", "semgrep", "rules.json")
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Error("Expected cached rules file to exist")
	}

	// Verify metadata file exists
	metaPath := filepath.Join(tmpDir, "cache", "feeds", "semgrep", "metadata.json")
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Error("Expected metadata file to exist")
	}
}

func TestManagerSyncFeedError(t *testing.T) {
	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	m.config.Feeds = []FeedConfig{
		{
			Type:      FeedSemgrepRules,
			Name:      "test",
			URL:       server.URL,
			Frequency: FreqAlways,
			Enabled:   true,
		},
	}

	ctx := context.Background()
	result := m.SyncFeed(ctx, FeedSemgrepRules)

	if result.Success {
		t.Error("Expected SyncFeed() to fail")
	}

	if result.Error == "" {
		t.Error("Expected error message")
	}
}

func TestManagerSyncIfNeeded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{"rules": []interface{}{}})
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	m.config.Feeds = []FeedConfig{
		{
			Type:      FeedSemgrepRules,
			Name:      "test",
			URL:       server.URL,
			Frequency: FreqWeekly,
			Enabled:   true,
		},
	}

	ctx := context.Background()

	// First sync should happen
	results, err := m.SyncIfNeeded(ctx)
	if err != nil {
		t.Fatalf("SyncIfNeeded() error = %v", err)
	}

	if len(results) == 0 {
		t.Fatal("Expected at least one result")
	}

	if results[0].Skipped {
		t.Error("Expected first sync to not be skipped")
	}

	// Second sync should be skipped (weekly frequency)
	results, err = m.SyncIfNeeded(ctx)
	if err != nil {
		t.Fatalf("SyncIfNeeded() error = %v", err)
	}

	if len(results) == 0 {
		t.Fatal("Expected at least one result")
	}

	if !results[0].Skipped {
		t.Error("Expected second sync to be skipped")
	}
}

func TestManagerHasCachedRules(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	// No cached rules initially
	if m.HasCachedRules() {
		t.Error("Expected HasCachedRules() to return false initially")
	}

	// Create cache file
	cacheDir := filepath.Join(tmpDir, "cache", "feeds", "semgrep")
	os.MkdirAll(cacheDir, 0755)
	os.WriteFile(filepath.Join(cacheDir, "rules.json"), []byte("{}"), 0644)

	// Now should have cached rules
	if !m.HasCachedRules() {
		t.Error("Expected HasCachedRules() to return true after creating cache")
	}
}

func TestManagerGetAllStatus(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	m.status[FeedSemgrepRules] = &FeedStatus{
		Type:      FeedSemgrepRules,
		Name:      "test",
		ItemCount: 50,
	}

	statuses := m.GetAllStatus()
	if len(statuses) != 1 {
		t.Errorf("GetAllStatus() returned %d statuses, want 1", len(statuses))
	}
}
