package feedback

import (
	"testing"
	"time"

	"github.com/crashappsec/zero/pkg/core/findings"
)

func TestNewFeedback(t *testing.T) {
	evidence := &findings.Evidence{
		Fingerprint: "abc123",
		RuleID:      "test-rule",
		GitHubOrg:   "testorg",
		GitHubRepo:  "testrepo",
		FilePath:    "src/main.go",
		LineStart:   10,
	}

	fb := NewFeedback(evidence, VerdictFalsePositive, "Test code")

	if fb.ID == "" {
		t.Error("NewFeedback() should generate an ID")
	}
	if fb.Fingerprint != "abc123" {
		t.Errorf("Fingerprint = %q, want %q", fb.Fingerprint, "abc123")
	}
	if fb.Verdict != VerdictFalsePositive {
		t.Errorf("Verdict = %q, want %q", fb.Verdict, VerdictFalsePositive)
	}
	if fb.Reason != "Test code" {
		t.Errorf("Reason = %q, want %q", fb.Reason, "Test code")
	}
	if fb.Source != "cli" {
		t.Errorf("Source = %q, want %q", fb.Source, "cli")
	}
	if fb.Evidence != evidence {
		t.Error("Evidence should be set")
	}
	if fb.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestNewFeedbackStore(t *testing.T) {
	store := NewFeedbackStore()

	if store.Version != "1.0" {
		t.Errorf("Version = %q, want %q", store.Version, "1.0")
	}
	if store.Entries == nil {
		t.Error("Entries should be initialized")
	}
	if len(store.Entries) != 0 {
		t.Errorf("Entries should be empty, got %d", len(store.Entries))
	}
	if store.Stats == nil {
		t.Error("Stats should be initialized")
	}
	if store.Stats.ByRule == nil {
		t.Error("Stats.ByRule should be initialized")
	}
	if store.Stats.ByRepo == nil {
		t.Error("Stats.ByRepo should be initialized")
	}
	if store.LastUpdated.IsZero() {
		t.Error("LastUpdated should be set")
	}
}

func TestFeedbackStore_Add(t *testing.T) {
	store := NewFeedbackStore()

	evidence := &findings.Evidence{
		Fingerprint: "fp1",
		RuleID:      "rule1",
		GitHubOrg:   "org1",
		GitHubRepo:  "repo1",
	}
	fb := NewFeedback(evidence, VerdictFalsePositive, "Test")

	store.Add(fb)

	if len(store.Entries) != 1 {
		t.Errorf("Entries count = %d, want 1", len(store.Entries))
	}
	if store.Stats.TotalFeedback != 1 {
		t.Errorf("TotalFeedback = %d, want 1", store.Stats.TotalFeedback)
	}
	if store.Stats.FalsePositives != 1 {
		t.Errorf("FalsePositives = %d, want 1", store.Stats.FalsePositives)
	}
}

func TestFeedbackStore_AddUpdate(t *testing.T) {
	store := NewFeedbackStore()

	evidence := &findings.Evidence{
		Fingerprint: "fp1",
		RuleID:      "rule1",
	}

	// Add initial feedback
	fb1 := NewFeedback(evidence, VerdictFalsePositive, "Initial")
	store.Add(fb1)
	originalCreatedAt := fb1.CreatedAt

	// Wait a bit to ensure time difference
	time.Sleep(10 * time.Millisecond)

	// Update with same fingerprint
	fb2 := NewFeedback(evidence, VerdictTruePositive, "Updated")
	store.Add(fb2)

	// Should still have 1 entry
	if len(store.Entries) != 1 {
		t.Errorf("Entries count = %d, want 1", len(store.Entries))
	}

	// Check the stored feedback
	stored, ok := store.Get("fp1")
	if !ok {
		t.Fatal("Feedback not found")
	}
	if stored.Verdict != VerdictTruePositive {
		t.Errorf("Verdict = %q, want %q", stored.Verdict, VerdictTruePositive)
	}
	if stored.Reason != "Updated" {
		t.Errorf("Reason = %q, want %q", stored.Reason, "Updated")
	}
	// CreatedAt should be preserved
	if !stored.CreatedAt.Equal(originalCreatedAt) {
		t.Error("CreatedAt should be preserved on update")
	}
	// UpdatedAt should be set
	if stored.UpdatedAt == nil {
		t.Error("UpdatedAt should be set on update")
	}

	// Stats should reflect the update
	if store.Stats.FalsePositives != 0 {
		t.Errorf("FalsePositives = %d, want 0", store.Stats.FalsePositives)
	}
	if store.Stats.TruePositives != 1 {
		t.Errorf("TruePositives = %d, want 1", store.Stats.TruePositives)
	}
}

func TestFeedbackStore_Get(t *testing.T) {
	store := NewFeedbackStore()

	// Get from empty store
	fb, ok := store.Get("nonexistent")
	if ok {
		t.Error("Get() should return false for nonexistent fingerprint")
	}
	if fb != nil {
		t.Error("Get() should return nil for nonexistent fingerprint")
	}

	// Add and get
	evidence := &findings.Evidence{Fingerprint: "fp1"}
	store.Add(NewFeedback(evidence, VerdictTruePositive, "Test"))

	fb, ok = store.Get("fp1")
	if !ok {
		t.Error("Get() should return true for existing fingerprint")
	}
	if fb == nil {
		t.Error("Get() should return feedback for existing fingerprint")
	}
	if fb.Fingerprint != "fp1" {
		t.Errorf("Fingerprint = %q, want %q", fb.Fingerprint, "fp1")
	}
}

func TestFeedbackStore_Query(t *testing.T) {
	store := NewFeedbackStore()

	// Add various feedback entries
	entries := []struct {
		fingerprint string
		verdict     Verdict
		ruleID      string
		org         string
		repo        string
	}{
		{"fp1", VerdictFalsePositive, "rule1", "org1", "repo1"},
		{"fp2", VerdictTruePositive, "rule1", "org1", "repo2"},
		{"fp3", VerdictFalsePositive, "rule2", "org2", "repo1"},
		{"fp4", VerdictNeedsReview, "rule2", "org2", "repo2"},
		{"fp5", VerdictIgnored, "rule3", "org1", "repo1"},
	}

	for _, e := range entries {
		evidence := &findings.Evidence{
			Fingerprint: e.fingerprint,
			RuleID:      e.ruleID,
			GitHubOrg:   e.org,
			GitHubRepo:  e.repo,
		}
		store.Add(NewFeedback(evidence, e.verdict, "test"))
	}

	tests := []struct {
		name     string
		query    FeedbackQuery
		wantLen  int
	}{
		{
			name:    "empty query returns all",
			query:   FeedbackQuery{},
			wantLen: 5,
		},
		{
			name:    "filter by verdict",
			query:   FeedbackQuery{Verdict: VerdictFalsePositive},
			wantLen: 2,
		},
		{
			name:    "filter by rule ID",
			query:   FeedbackQuery{RuleID: "rule1"},
			wantLen: 2,
		},
		{
			name:    "filter by org",
			query:   FeedbackQuery{GitHubOrg: "org1"},
			wantLen: 3,
		},
		{
			name:    "filter by repo",
			query:   FeedbackQuery{GitHubRepo: "repo1"},
			wantLen: 3,
		},
		{
			name:    "filter by fingerprint",
			query:   FeedbackQuery{Fingerprint: "fp1"},
			wantLen: 1,
		},
		{
			name:    "combined filters",
			query:   FeedbackQuery{Verdict: VerdictFalsePositive, GitHubOrg: "org1"},
			wantLen: 1,
		},
		{
			name:    "no matches",
			query:   FeedbackQuery{RuleID: "nonexistent"},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := store.Query(tt.query)
			if len(results) != tt.wantLen {
				t.Errorf("Query() returned %d results, want %d", len(results), tt.wantLen)
			}
		})
	}
}

func TestFeedbackStore_QueryPagination(t *testing.T) {
	store := NewFeedbackStore()

	// Add 10 entries
	for i := 0; i < 10; i++ {
		evidence := &findings.Evidence{Fingerprint: string(rune('a' + i))}
		store.Add(NewFeedback(evidence, VerdictFalsePositive, "test"))
	}

	// Test limit
	results := store.Query(FeedbackQuery{Limit: 3})
	if len(results) != 3 {
		t.Errorf("Query with Limit=3 returned %d results, want 3", len(results))
	}

	// Test offset
	results = store.Query(FeedbackQuery{Offset: 5})
	if len(results) != 5 {
		t.Errorf("Query with Offset=5 returned %d results, want 5", len(results))
	}

	// Test limit and offset
	results = store.Query(FeedbackQuery{Offset: 3, Limit: 2})
	if len(results) != 2 {
		t.Errorf("Query with Offset=3, Limit=2 returned %d results, want 2", len(results))
	}
}

func TestFeedbackStore_Stats(t *testing.T) {
	store := NewFeedbackStore()

	// Add feedback entries
	entries := []struct {
		fingerprint string
		verdict     Verdict
		ruleID      string
		org         string
		repo        string
	}{
		{"fp1", VerdictFalsePositive, "rule1", "org1", "repo1"},
		{"fp2", VerdictFalsePositive, "rule1", "org1", "repo1"},
		{"fp3", VerdictTruePositive, "rule1", "org1", "repo2"},
		{"fp4", VerdictTruePositive, "rule2", "org2", "repo1"},
		{"fp5", VerdictNeedsReview, "rule2", "org2", "repo2"},
		{"fp6", VerdictIgnored, "rule3", "org1", "repo1"},
	}

	for _, e := range entries {
		evidence := &findings.Evidence{
			Fingerprint: e.fingerprint,
			RuleID:      e.ruleID,
			GitHubOrg:   e.org,
			GitHubRepo:  e.repo,
		}
		store.Add(NewFeedback(evidence, e.verdict, "test"))
	}

	stats := store.Stats

	if stats.TotalFeedback != 6 {
		t.Errorf("TotalFeedback = %d, want 6", stats.TotalFeedback)
	}
	if stats.FalsePositives != 2 {
		t.Errorf("FalsePositives = %d, want 2", stats.FalsePositives)
	}
	if stats.TruePositives != 2 {
		t.Errorf("TruePositives = %d, want 2", stats.TruePositives)
	}
	if stats.NeedsReview != 1 {
		t.Errorf("NeedsReview = %d, want 1", stats.NeedsReview)
	}
	if stats.Ignored != 1 {
		t.Errorf("Ignored = %d, want 1", stats.Ignored)
	}

	// FP rate = 2 / (2 + 2) = 0.5
	expectedFPRate := 0.5
	if stats.FalsePositiveRate != expectedFPRate {
		t.Errorf("FalsePositiveRate = %f, want %f", stats.FalsePositiveRate, expectedFPRate)
	}

	// Check by rule
	if stats.ByRule["rule1"] != 3 {
		t.Errorf("ByRule[rule1] = %d, want 3", stats.ByRule["rule1"])
	}
	if stats.ByRule["rule2"] != 2 {
		t.Errorf("ByRule[rule2] = %d, want 2", stats.ByRule["rule2"])
	}

	// Check by repo
	if stats.ByRepo["org1/repo1"] != 3 {
		t.Errorf("ByRepo[org1/repo1] = %d, want 3", stats.ByRepo["org1/repo1"])
	}
}

func TestMatchesQuery_TimeFilters(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	tomorrow := now.Add(24 * time.Hour)
	lastWeek := now.Add(-7 * 24 * time.Hour)

	fb := &Feedback{
		Fingerprint: "fp1",
		CreatedAt:   now,
	}

	// Since filter
	if !matchesQuery(fb, FeedbackQuery{Since: &yesterday}) {
		t.Error("Should match when created after Since")
	}
	if matchesQuery(fb, FeedbackQuery{Since: &tomorrow}) {
		t.Error("Should not match when created before Since")
	}

	// Until filter
	if !matchesQuery(fb, FeedbackQuery{Until: &tomorrow}) {
		t.Error("Should match when created before Until")
	}
	if matchesQuery(fb, FeedbackQuery{Until: &lastWeek}) {
		t.Error("Should not match when created after Until")
	}

	// Combined Since and Until
	if !matchesQuery(fb, FeedbackQuery{Since: &yesterday, Until: &tomorrow}) {
		t.Error("Should match when within time range")
	}
}

func TestVerdictConstants(t *testing.T) {
	// Ensure verdict constants have expected values
	if VerdictFalsePositive != "false_positive" {
		t.Errorf("VerdictFalsePositive = %q, want %q", VerdictFalsePositive, "false_positive")
	}
	if VerdictTruePositive != "true_positive" {
		t.Errorf("VerdictTruePositive = %q, want %q", VerdictTruePositive, "true_positive")
	}
	if VerdictNeedsReview != "needs_review" {
		t.Errorf("VerdictNeedsReview = %q, want %q", VerdictNeedsReview, "needs_review")
	}
	if VerdictIgnored != "ignored" {
		t.Errorf("VerdictIgnored = %q, want %q", VerdictIgnored, "ignored")
	}
}
