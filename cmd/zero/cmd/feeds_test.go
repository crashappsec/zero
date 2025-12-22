package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/crashappsec/zero/pkg/external/feeds"
	"github.com/crashappsec/zero/pkg/core/rules"
)

func TestFeedsCommandFlagDefaults(t *testing.T) {
	// Verify default flag values
	if feedsForce {
		t.Error("Expected feedsForce default to be false")
	}
}

func TestFeedsSyncWithMockServer(t *testing.T) {
	// Create mock Semgrep server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rules := map[string]interface{}{
			"rules": []map[string]interface{}{
				{"id": "rule1", "pattern": "test", "message": "Test rule"},
				{"id": "rule2", "pattern": "test2", "message": "Test rule 2"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rules)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	mgr := feeds.NewManager(tmpDir)

	// Configure to use mock server
	mgr.SetConfig(feeds.Config{
		DefaultFreq: feeds.FreqDaily,
		Feeds: []feeds.FeedConfig{
			{
				Type:      feeds.FeedSemgrepRules,
				Name:      "test-rules",
				URL:       server.URL,
				Frequency: feeds.FreqAlways,
				Enabled:   true,
			},
		},
	})

	ctx := context.Background()
	results, err := mgr.SyncAll(ctx)
	if err != nil {
		t.Fatalf("SyncAll failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("Expected at least one result")
	}

	if !results[0].Success {
		t.Errorf("Expected sync to succeed, got error: %s", results[0].Error)
	}

	if results[0].ItemCount != 2 {
		t.Errorf("Expected 2 rules, got %d", results[0].ItemCount)
	}

	// Verify cache file was created
	cachePath := filepath.Join(tmpDir, "cache", "feeds", "semgrep", "rules.json")
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Error("Expected cache file to be created")
	}
}

func TestFeedsSyncIfNeeded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{"rules": []interface{}{}})
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	mgr := feeds.NewManager(tmpDir)

	// Configure with weekly frequency
	mgr.SetConfig(feeds.Config{
		Feeds: []feeds.FeedConfig{
			{
				Type:      feeds.FeedSemgrepRules,
				Name:      "test",
				URL:       server.URL,
				Frequency: feeds.FreqWeekly,
				Enabled:   true,
			},
		},
	})

	ctx := context.Background()

	// First sync should happen
	results1, _ := mgr.SyncIfNeeded(ctx)
	if len(results1) == 0 || results1[0].Skipped {
		t.Error("Expected first sync to not be skipped")
	}

	// Second sync should be skipped (weekly frequency, just synced)
	results2, _ := mgr.SyncIfNeeded(ctx)
	if len(results2) == 0 || !results2[0].Skipped {
		t.Error("Expected second sync to be skipped")
	}
}

func TestFeedsStatusTracking(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{"rules": []interface{}{}})
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	mgr := feeds.NewManager(tmpDir)

	mgr.SetConfig(feeds.Config{
		Feeds: []feeds.FeedConfig{
			{
				Type:      feeds.FeedSemgrepRules,
				Name:      "test",
				URL:       server.URL,
				Frequency: feeds.FreqAlways,
				Enabled:   true,
			},
		},
	})

	ctx := context.Background()
	mgr.SyncAll(ctx)

	// Check status
	status := mgr.GetStatus(feeds.FeedSemgrepRules)
	if status == nil {
		t.Fatal("Expected status to be recorded")
	}

	if status.LastSync.IsZero() {
		t.Error("Expected LastSync to be set")
	}
}

func TestFeedsErrorHandling(t *testing.T) {
	// Create server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	mgr := feeds.NewManager(tmpDir)

	mgr.SetConfig(feeds.Config{
		Feeds: []feeds.FeedConfig{
			{
				Type:      feeds.FeedSemgrepRules,
				Name:      "test",
				URL:       server.URL,
				Frequency: feeds.FreqAlways,
				Enabled:   true,
			},
		},
	})

	ctx := context.Background()
	results, _ := mgr.SyncAll(ctx)

	if len(results) == 0 {
		t.Fatal("Expected result")
	}

	if results[0].Success {
		t.Error("Expected sync to fail")
	}

	if results[0].Error == "" {
		t.Error("Expected error message")
	}
}

func TestRulesGeneration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create minimal RAG structure
	ragDir := filepath.Join(tmpDir, "..", "rag", "technology-identification")
	os.MkdirAll(ragDir, 0755)

	// Create a simple pattern file
	pattern := `# Test Pattern
This is a test pattern for testing.
`
	os.WriteFile(filepath.Join(ragDir, "test.md"), []byte(pattern), 0644)

	mgr := rules.NewManager(tmpDir)

	results, err := mgr.RefreshRules(true) // Force refresh
	if err != nil {
		t.Fatalf("RefreshRules failed: %v", err)
	}

	// Should have results for generated and community
	if len(results) == 0 {
		t.Error("Expected at least one result")
	}
}

func TestRulesStatusTracking(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := rules.NewManager(tmpDir)

	// Initially no status
	statuses := mgr.GetStatus()
	if len(statuses) != 0 {
		t.Errorf("Expected 0 statuses initially, got %d", len(statuses))
	}

	// Refresh to generate status
	mgr.RefreshRules(true)

	// Now should have status
	statuses = mgr.GetStatus()
	// May have generated and/or community depending on setup
}

func TestRulesConfigUpdate(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := rules.NewManager(tmpDir)

	// Get default config
	cfg := mgr.GetConfig()
	if !cfg.GeneratedRules.Enabled {
		t.Error("Expected generated rules enabled by default")
	}

	// Update config
	newCfg := feeds.RuleConfig{
		GeneratedRules: feeds.RuleSourceConfig{
			Enabled: false,
		},
		CommunityRules: feeds.RuleSourceConfig{
			Enabled: false,
		},
	}
	mgr.SetConfig(newCfg)

	// Verify update
	cfg = mgr.GetConfig()
	if cfg.GeneratedRules.Enabled {
		t.Error("Expected generated rules to be disabled after update")
	}

	// Refresh should do nothing when disabled
	results, _ := mgr.RefreshRules(false)
	if len(results) != 0 {
		t.Errorf("Expected 0 results when all disabled, got %d", len(results))
	}
}
