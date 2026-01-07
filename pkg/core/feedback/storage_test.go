package feedback

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/crashappsec/zero/pkg/core/findings"
)

func TestNewStorage(t *testing.T) {
	storage := NewStorage("/tmp/test-zero")

	if storage.zeroHome != "/tmp/test-zero" {
		t.Errorf("zeroHome = %q, want %q", storage.zeroHome, "/tmp/test-zero")
	}
}

func TestStorage_FeedbackPath(t *testing.T) {
	storage := NewStorage("/tmp/test-zero")
	path := storage.feedbackPath()

	expected := "/tmp/test-zero/feedback/feedback.json"
	if path != expected {
		t.Errorf("feedbackPath() = %q, want %q", path, expected)
	}
}

func TestStorage_ExportPath(t *testing.T) {
	storage := NewStorage("/tmp/test-zero")

	csvPath := storage.exportPath("csv")
	if csvPath != "/tmp/test-zero/feedback/export.csv" {
		t.Errorf("exportPath(csv) = %q, want %q", csvPath, "/tmp/test-zero/feedback/export.csv")
	}

	jsonPath := storage.exportPath("json")
	if jsonPath != "/tmp/test-zero/feedback/export.json" {
		t.Errorf("exportPath(json) = %q, want %q", jsonPath, "/tmp/test-zero/feedback/export.json")
	}
}

func TestStorage_LoadEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewStorage(tmpDir)

	store, err := storage.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if store == nil {
		t.Fatal("Load() returned nil store")
	}
	if store.Version != "1.0" {
		t.Errorf("Version = %q, want %q", store.Version, "1.0")
	}
	if len(store.Entries) != 0 {
		t.Errorf("Entries should be empty, got %d", len(store.Entries))
	}
}

func TestStorage_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewStorage(tmpDir)

	// Load creates empty store
	store, err := storage.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Add feedback
	evidence := &findings.Evidence{
		Fingerprint: "test-fp",
		RuleID:      "test-rule",
		GitHubOrg:   "testorg",
		GitHubRepo:  "testrepo",
		FilePath:    "main.go",
		LineStart:   42,
	}
	fb := NewFeedback(evidence, VerdictFalsePositive, "Test reason")
	store.Add(fb)

	// Save
	err = storage.Save()
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file exists
	feedbackPath := storage.feedbackPath()
	if _, err := os.Stat(feedbackPath); os.IsNotExist(err) {
		t.Error("Feedback file should exist after Save()")
	}

	// Create new storage and load
	storage2 := NewStorage(tmpDir)
	store2, err := storage2.Load()
	if err != nil {
		t.Fatalf("Load() after save error = %v", err)
	}

	if len(store2.Entries) != 1 {
		t.Errorf("Entries count = %d, want 1", len(store2.Entries))
	}

	loaded, ok := store2.Get("test-fp")
	if !ok {
		t.Fatal("Loaded store missing feedback entry")
	}
	if loaded.Verdict != VerdictFalsePositive {
		t.Errorf("Verdict = %q, want %q", loaded.Verdict, VerdictFalsePositive)
	}
	if loaded.Reason != "Test reason" {
		t.Errorf("Reason = %q, want %q", loaded.Reason, "Test reason")
	}
}

func TestStorage_AddFeedback(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewStorage(tmpDir)

	evidence := &findings.Evidence{
		Fingerprint: "add-test-fp",
		RuleID:      "rule1",
	}
	fb := NewFeedback(evidence, VerdictTruePositive, "Real issue")

	err := storage.AddFeedback(fb)
	if err != nil {
		t.Fatalf("AddFeedback() error = %v", err)
	}

	// Verify it was saved
	storage2 := NewStorage(tmpDir)
	loaded, err := storage2.GetFeedback("add-test-fp")
	if err != nil {
		t.Fatalf("GetFeedback() error = %v", err)
	}
	if loaded == nil {
		t.Fatal("GetFeedback() returned nil")
	}
	if loaded.Verdict != VerdictTruePositive {
		t.Errorf("Verdict = %q, want %q", loaded.Verdict, VerdictTruePositive)
	}
}

func TestStorage_GetFeedback(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewStorage(tmpDir)

	// Get nonexistent
	fb, err := storage.GetFeedback("nonexistent")
	if err != nil {
		t.Fatalf("GetFeedback() error = %v", err)
	}
	if fb != nil {
		t.Error("GetFeedback() should return nil for nonexistent fingerprint")
	}

	// Add and get
	evidence := &findings.Evidence{Fingerprint: "get-test"}
	storage.AddFeedback(NewFeedback(evidence, VerdictIgnored, "Ignored"))

	fb, err = storage.GetFeedback("get-test")
	if err != nil {
		t.Fatalf("GetFeedback() error = %v", err)
	}
	if fb == nil {
		t.Fatal("GetFeedback() returned nil for existing fingerprint")
	}
}

func TestStorage_QueryFeedback(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewStorage(tmpDir)

	// Add multiple entries
	for i, v := range []Verdict{VerdictFalsePositive, VerdictTruePositive, VerdictFalsePositive} {
		evidence := &findings.Evidence{
			Fingerprint: string(rune('a' + i)),
			RuleID:      "rule1",
		}
		storage.AddFeedback(NewFeedback(evidence, v, "test"))
	}

	// Query all
	results, err := storage.QueryFeedback(FeedbackQuery{})
	if err != nil {
		t.Fatalf("QueryFeedback() error = %v", err)
	}
	if len(results) != 3 {
		t.Errorf("QueryFeedback() returned %d results, want 3", len(results))
	}

	// Query by verdict
	results, err = storage.QueryFeedback(FeedbackQuery{Verdict: VerdictFalsePositive})
	if err != nil {
		t.Fatalf("QueryFeedback() error = %v", err)
	}
	if len(results) != 2 {
		t.Errorf("QueryFeedback(FP) returned %d results, want 2", len(results))
	}
}

func TestStorage_GetStats(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewStorage(tmpDir)

	// Add entries
	verdicts := []Verdict{
		VerdictFalsePositive,
		VerdictFalsePositive,
		VerdictTruePositive,
		VerdictNeedsReview,
	}
	for i, v := range verdicts {
		evidence := &findings.Evidence{
			Fingerprint: string(rune('a' + i)),
			RuleID:      "rule1",
		}
		storage.AddFeedback(NewFeedback(evidence, v, "test"))
	}

	stats, err := storage.GetStats()
	if err != nil {
		t.Fatalf("GetStats() error = %v", err)
	}

	if stats.TotalFeedback != 4 {
		t.Errorf("TotalFeedback = %d, want 4", stats.TotalFeedback)
	}
	if stats.FalsePositives != 2 {
		t.Errorf("FalsePositives = %d, want 2", stats.FalsePositives)
	}
	if stats.TruePositives != 1 {
		t.Errorf("TruePositives = %d, want 1", stats.TruePositives)
	}
}

func TestStorage_ExportCSV(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewStorage(tmpDir)

	// Add feedback
	evidence := &findings.Evidence{
		Fingerprint: "csv-test",
		RuleID:      "rule1",
		GitHubOrg:   "org1",
		GitHubRepo:  "repo1",
		FilePath:    "main.go",
		LineStart:   10,
	}
	storage.AddFeedback(NewFeedback(evidence, VerdictFalsePositive, "Test reason"))

	path, err := storage.ExportCSV()
	if err != nil {
		t.Fatalf("ExportCSV() error = %v", err)
	}

	// Read and verify
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Reading CSV export error = %v", err)
	}

	csv := string(data)
	if !strings.Contains(csv, "fingerprint,verdict,rule_id") {
		t.Error("CSV should contain header row")
	}
	if !strings.Contains(csv, "csv-test") {
		t.Error("CSV should contain fingerprint")
	}
	if !strings.Contains(csv, "false_positive") {
		t.Error("CSV should contain verdict")
	}
	if !strings.Contains(csv, "rule1") {
		t.Error("CSV should contain rule_id")
	}
}

func TestStorage_ExportJSON(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewStorage(tmpDir)

	// Add feedback
	evidence := &findings.Evidence{
		Fingerprint: "json-test",
		RuleID:      "rule1",
		RuleSource:  "rag",
		GitHubOrg:   "org1",
		GitHubRepo:  "repo1",
		FilePath:    "main.go",
		LineStart:   20,
		MatchedText: "secret = 'abc123'",
	}
	storage.AddFeedback(NewFeedback(evidence, VerdictTruePositive, "Real secret"))

	path, err := storage.ExportJSON()
	if err != nil {
		t.Fatalf("ExportJSON() error = %v", err)
	}

	// Read and verify
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Reading JSON export error = %v", err)
	}

	var export struct {
		Version string `json:"version"`
		Stats   struct {
			TotalFeedback int `json:"total_feedback"`
		} `json:"stats"`
		Entries []struct {
			Fingerprint string `json:"fingerprint"`
			Verdict     string `json:"verdict"`
			Evidence    struct {
				RuleID   string `json:"rule_id"`
				FilePath string `json:"file_path"`
			} `json:"evidence"`
		} `json:"entries"`
	}

	if err := json.Unmarshal(data, &export); err != nil {
		t.Fatalf("Parsing JSON export error = %v", err)
	}

	if export.Version != "1.0" {
		t.Errorf("Version = %q, want %q", export.Version, "1.0")
	}
	if export.Stats.TotalFeedback != 1 {
		t.Errorf("TotalFeedback = %d, want 1", export.Stats.TotalFeedback)
	}
	if len(export.Entries) != 1 {
		t.Fatalf("Entries count = %d, want 1", len(export.Entries))
	}
	if export.Entries[0].Fingerprint != "json-test" {
		t.Errorf("Fingerprint = %q, want %q", export.Entries[0].Fingerprint, "json-test")
	}
	if export.Entries[0].Verdict != "true_positive" {
		t.Errorf("Verdict = %q, want %q", export.Entries[0].Verdict, "true_positive")
	}
}

func TestStorage_GetFalsePositiveRules(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewStorage(tmpDir)

	// Add feedback for rule1: 3 FP, 1 TP (75% FP rate)
	for i := 0; i < 3; i++ {
		evidence := &findings.Evidence{
			Fingerprint: string(rune('a' + i)),
			RuleID:      "rule1",
		}
		storage.AddFeedback(NewFeedback(evidence, VerdictFalsePositive, "FP"))
	}
	evidence := &findings.Evidence{Fingerprint: "d", RuleID: "rule1"}
	storage.AddFeedback(NewFeedback(evidence, VerdictTruePositive, "TP"))

	// Add feedback for rule2: 1 FP, 3 TP (25% FP rate)
	evidence = &findings.Evidence{Fingerprint: "e", RuleID: "rule2"}
	storage.AddFeedback(NewFeedback(evidence, VerdictFalsePositive, "FP"))
	for i := 0; i < 3; i++ {
		evidence := &findings.Evidence{
			Fingerprint: string(rune('f' + i)),
			RuleID:      "rule2",
		}
		storage.AddFeedback(NewFeedback(evidence, VerdictTruePositive, "TP"))
	}

	// Get rules with FP rate >= 50%
	rules, err := storage.GetFalsePositiveRules(0.5)
	if err != nil {
		t.Fatalf("GetFalsePositiveRules() error = %v", err)
	}

	if len(rules) != 1 {
		t.Fatalf("GetFalsePositiveRules(0.5) returned %d rules, want 1", len(rules))
	}
	if rules[0].RuleID != "rule1" {
		t.Errorf("RuleID = %q, want %q", rules[0].RuleID, "rule1")
	}
	if rules[0].FPRate != 0.75 {
		t.Errorf("FPRate = %f, want 0.75", rules[0].FPRate)
	}

	// Get rules with FP rate >= 20% (should include both)
	rules, err = storage.GetFalsePositiveRules(0.2)
	if err != nil {
		t.Fatalf("GetFalsePositiveRules() error = %v", err)
	}
	if len(rules) != 2 {
		t.Errorf("GetFalsePositiveRules(0.2) returned %d rules, want 2", len(rules))
	}
}

func TestStorage_LoadExistingFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create feedback directory and file manually
	feedbackDir := filepath.Join(tmpDir, "feedback")
	os.MkdirAll(feedbackDir, 0755)

	existingData := `{
		"version": "1.0",
		"entries": {
			"existing-fp": {
				"id": "test-id",
				"fingerprint": "existing-fp",
				"verdict": "false_positive",
				"reason": "Pre-existing",
				"created_at": "2024-01-01T00:00:00Z"
			}
		},
		"stats": {
			"total_feedback": 1,
			"false_positives": 1
		},
		"last_updated": "2024-01-01T00:00:00Z"
	}`

	feedbackPath := filepath.Join(feedbackDir, "feedback.json")
	os.WriteFile(feedbackPath, []byte(existingData), 0644)

	// Load
	storage := NewStorage(tmpDir)
	store, err := storage.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(store.Entries) != 1 {
		t.Errorf("Entries count = %d, want 1", len(store.Entries))
	}

	fb, ok := store.Get("existing-fp")
	if !ok {
		t.Fatal("Pre-existing feedback not loaded")
	}
	if fb.Reason != "Pre-existing" {
		t.Errorf("Reason = %q, want %q", fb.Reason, "Pre-existing")
	}
}

func TestEscapeCSV(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with \"quotes\"", "with \"\"quotes\"\""},
		{"with\nnewline", "with newline"},
		{"mixed \"quote\"\nand newline", "mixed \"\"quote\"\" and newline"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := escapeCSV(tt.input)
			if result != tt.expected {
				t.Errorf("escapeCSV(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
