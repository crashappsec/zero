// Package feedback provides types and storage for analyst feedback on findings
package feedback

import (
	"time"

	"github.com/crashappsec/zero/pkg/core/findings"
)

// Verdict represents the analyst's determination about a finding
type Verdict string

const (
	// VerdictFalsePositive indicates the finding is a false positive
	VerdictFalsePositive Verdict = "false_positive"
	// VerdictTruePositive indicates the finding is a real issue
	VerdictTruePositive Verdict = "true_positive"
	// VerdictNeedsReview indicates the finding needs further review
	VerdictNeedsReview Verdict = "needs_review"
	// VerdictIgnored indicates the finding should be ignored (e.g., test file)
	VerdictIgnored Verdict = "ignored"
)

// Feedback represents analyst feedback on a specific finding
type Feedback struct {
	// ID is a unique identifier for this feedback entry
	ID string `json:"id"`

	// Fingerprint links to the finding's unique identifier
	Fingerprint string `json:"fingerprint"`

	// Evidence captures the original finding details for context
	Evidence *findings.Evidence `json:"evidence"`

	// Verdict is the analyst's determination
	Verdict Verdict `json:"verdict"`

	// Confidence is the analyst's confidence in their verdict (0.0-1.0)
	Confidence float64 `json:"confidence,omitempty"`

	// Reason explains why this verdict was given
	Reason string `json:"reason,omitempty"`

	// Category classifies the feedback (e.g., "test_code", "example", "documentation")
	Category string `json:"category,omitempty"`

	// RuleImprovement suggests changes to the detection rule
	RuleImprovement *RuleImprovement `json:"rule_improvement,omitempty"`

	// Analyst information
	AnalystID    string `json:"analyst_id,omitempty"`
	AnalystEmail string `json:"analyst_email,omitempty"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`

	// Source tracking
	Source string `json:"source,omitempty"` // "cli", "web", "api"
}

// RuleImprovement captures suggestions for improving detection rules
type RuleImprovement struct {
	// Action suggests what to do with the rule
	Action string `json:"action"` // "modify_pattern", "add_exclusion", "change_severity", "disable"

	// Details contains action-specific information
	SuggestedPattern   string   `json:"suggested_pattern,omitempty"`
	SuggestedExclusion string   `json:"suggested_exclusion,omitempty"`
	SuggestedSeverity  string   `json:"suggested_severity,omitempty"`
	AffectedFiles      []string `json:"affected_files,omitempty"`

	// Notes for rule maintainers
	Notes string `json:"notes,omitempty"`
}

// FeedbackStore represents a collection of feedback entries
type FeedbackStore struct {
	// Version of the feedback store format
	Version string `json:"version"`

	// Entries contains all feedback, indexed by fingerprint
	Entries map[string]*Feedback `json:"entries"`

	// Stats tracks aggregate statistics
	Stats *FeedbackStats `json:"stats"`

	// LastUpdated is when the store was last modified
	LastUpdated time.Time `json:"last_updated"`
}

// FeedbackStats tracks aggregate feedback statistics
type FeedbackStats struct {
	TotalFeedback    int            `json:"total_feedback"`
	FalsePositives   int            `json:"false_positives"`
	TruePositives    int            `json:"true_positives"`
	NeedsReview      int            `json:"needs_review"`
	Ignored          int            `json:"ignored"`
	ByRule           map[string]int `json:"by_rule"`            // Count by rule ID
	ByRepo           map[string]int `json:"by_repo"`            // Count by repo
	FalsePositiveRate float64       `json:"false_positive_rate"` // FP / (FP + TP)
}

// FeedbackQuery represents filters for querying feedback
type FeedbackQuery struct {
	// Filter by fingerprint
	Fingerprint string

	// Filter by verdict
	Verdict Verdict

	// Filter by rule ID
	RuleID string

	// Filter by repository
	GitHubOrg  string
	GitHubRepo string

	// Filter by time range
	Since *time.Time
	Until *time.Time

	// Pagination
	Limit  int
	Offset int
}

// NewFeedback creates a new Feedback entry from evidence
func NewFeedback(evidence *findings.Evidence, verdict Verdict, reason string) *Feedback {
	return &Feedback{
		ID:          generateFeedbackID(),
		Fingerprint: evidence.Fingerprint,
		Evidence:    evidence,
		Verdict:     verdict,
		Reason:      reason,
		CreatedAt:   time.Now(),
		Source:      "cli",
	}
}

// NewFeedbackStore creates an empty feedback store
func NewFeedbackStore() *FeedbackStore {
	return &FeedbackStore{
		Version:     "1.0",
		Entries:     make(map[string]*Feedback),
		Stats:       &FeedbackStats{ByRule: make(map[string]int), ByRepo: make(map[string]int)},
		LastUpdated: time.Now(),
	}
}

// Add adds or updates feedback in the store
func (s *FeedbackStore) Add(fb *Feedback) {
	// If updating existing entry, preserve creation time
	if existing, ok := s.Entries[fb.Fingerprint]; ok {
		fb.CreatedAt = existing.CreatedAt
		now := time.Now()
		fb.UpdatedAt = &now
	}

	s.Entries[fb.Fingerprint] = fb
	s.LastUpdated = time.Now()
	s.recalculateStats()
}

// Get retrieves feedback by fingerprint
func (s *FeedbackStore) Get(fingerprint string) (*Feedback, bool) {
	fb, ok := s.Entries[fingerprint]
	return fb, ok
}

// Query returns feedback matching the given criteria
func (s *FeedbackStore) Query(q FeedbackQuery) []*Feedback {
	var results []*Feedback

	for _, fb := range s.Entries {
		if !matchesQuery(fb, q) {
			continue
		}
		results = append(results, fb)
	}

	// Apply pagination
	if q.Offset > 0 && q.Offset < len(results) {
		results = results[q.Offset:]
	}
	if q.Limit > 0 && q.Limit < len(results) {
		results = results[:q.Limit]
	}

	return results
}

// recalculateStats updates aggregate statistics
func (s *FeedbackStore) recalculateStats() {
	stats := &FeedbackStats{
		ByRule: make(map[string]int),
		ByRepo: make(map[string]int),
	}

	for _, fb := range s.Entries {
		stats.TotalFeedback++

		switch fb.Verdict {
		case VerdictFalsePositive:
			stats.FalsePositives++
		case VerdictTruePositive:
			stats.TruePositives++
		case VerdictNeedsReview:
			stats.NeedsReview++
		case VerdictIgnored:
			stats.Ignored++
		}

		if fb.Evidence != nil {
			stats.ByRule[fb.Evidence.RuleID]++
			if fb.Evidence.GitHubOrg != "" && fb.Evidence.GitHubRepo != "" {
				repoKey := fb.Evidence.GitHubOrg + "/" + fb.Evidence.GitHubRepo
				stats.ByRepo[repoKey]++
			}
		}
	}

	// Calculate false positive rate
	total := stats.FalsePositives + stats.TruePositives
	if total > 0 {
		stats.FalsePositiveRate = float64(stats.FalsePositives) / float64(total)
	}

	s.Stats = stats
}

// matchesQuery checks if feedback matches query criteria
func matchesQuery(fb *Feedback, q FeedbackQuery) bool {
	if q.Fingerprint != "" && fb.Fingerprint != q.Fingerprint {
		return false
	}
	if q.Verdict != "" && fb.Verdict != q.Verdict {
		return false
	}
	if q.RuleID != "" && fb.Evidence != nil && fb.Evidence.RuleID != q.RuleID {
		return false
	}
	if q.GitHubOrg != "" && fb.Evidence != nil && fb.Evidence.GitHubOrg != q.GitHubOrg {
		return false
	}
	if q.GitHubRepo != "" && fb.Evidence != nil && fb.Evidence.GitHubRepo != q.GitHubRepo {
		return false
	}
	if q.Since != nil && fb.CreatedAt.Before(*q.Since) {
		return false
	}
	if q.Until != nil && fb.CreatedAt.After(*q.Until) {
		return false
	}
	return true
}

// generateFeedbackID creates a unique ID for feedback
func generateFeedbackID() string {
	return time.Now().Format("20060102-150405") + "-" + randomHex(4)
}

// randomHex generates a random hex string of n bytes
func randomHex(n int) string {
	const chars = "0123456789abcdef"
	result := make([]byte, n*2)
	for i := range result {
		result[i] = chars[time.Now().UnixNano()%16]
	}
	return string(result)
}
