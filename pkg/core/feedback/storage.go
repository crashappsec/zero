package feedback

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Storage manages feedback persistence
type Storage struct {
	zeroHome string
	mu       sync.RWMutex
	store    *FeedbackStore
}

// NewStorage creates a new feedback storage instance
func NewStorage(zeroHome string) *Storage {
	return &Storage{
		zeroHome: zeroHome,
	}
}

// feedbackPath returns the path to the feedback store file
func (s *Storage) feedbackPath() string {
	return filepath.Join(s.zeroHome, "feedback", "feedback.json")
}

// exportPath returns the path for feedback exports
func (s *Storage) exportPath(format string) string {
	return filepath.Join(s.zeroHome, "feedback", fmt.Sprintf("export.%s", format))
}

// Load loads the feedback store from disk
func (s *Storage) Load() (*FeedbackStore, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.feedbackPath()

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating feedback directory: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Return empty store
		s.store = NewFeedbackStore()
		return s.store, nil
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading feedback file: %w", err)
	}

	// Parse JSON
	store := &FeedbackStore{}
	if err := json.Unmarshal(data, store); err != nil {
		return nil, fmt.Errorf("parsing feedback file: %w", err)
	}

	// Initialize maps if nil (backwards compatibility)
	if store.Entries == nil {
		store.Entries = make(map[string]*Feedback)
	}
	if store.Stats == nil {
		store.Stats = &FeedbackStats{ByRule: make(map[string]int), ByRepo: make(map[string]int)}
	}

	s.store = store
	return s.store, nil
}

// Save persists the feedback store to disk
func (s *Storage) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.store == nil {
		return nil // Nothing to save
	}

	path := s.feedbackPath()

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating feedback directory: %w", err)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(s.store, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling feedback: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing feedback file: %w", err)
	}

	return nil
}

// AddFeedback adds feedback and saves to disk
func (s *Storage) AddFeedback(fb *Feedback) error {
	// Load current store
	store, err := s.Load()
	if err != nil {
		return err
	}

	// Add feedback
	store.Add(fb)

	// Save
	return s.Save()
}

// GetFeedback retrieves feedback by fingerprint
func (s *Storage) GetFeedback(fingerprint string) (*Feedback, error) {
	store, err := s.Load()
	if err != nil {
		return nil, err
	}

	fb, ok := store.Get(fingerprint)
	if !ok {
		return nil, nil
	}
	return fb, nil
}

// QueryFeedback searches for feedback matching criteria
func (s *Storage) QueryFeedback(q FeedbackQuery) ([]*Feedback, error) {
	store, err := s.Load()
	if err != nil {
		return nil, err
	}

	return store.Query(q), nil
}

// GetStats returns aggregate statistics
func (s *Storage) GetStats() (*FeedbackStats, error) {
	store, err := s.Load()
	if err != nil {
		return nil, err
	}

	return store.Stats, nil
}

// ExportCSV exports feedback to CSV format
func (s *Storage) ExportCSV() (string, error) {
	store, err := s.Load()
	if err != nil {
		return "", err
	}

	// Build CSV content
	csv := "fingerprint,verdict,rule_id,github_org,github_repo,file_path,line_start,reason,created_at\n"
	for _, fb := range store.Entries {
		ruleID := ""
		org := ""
		repo := ""
		file := ""
		line := 0
		if fb.Evidence != nil {
			ruleID = fb.Evidence.RuleID
			org = fb.Evidence.GitHubOrg
			repo = fb.Evidence.GitHubRepo
			file = fb.Evidence.FilePath
			line = fb.Evidence.LineStart
		}
		csv += fmt.Sprintf("%s,%s,%s,%s,%s,%s,%d,\"%s\",%s\n",
			fb.Fingerprint,
			fb.Verdict,
			ruleID,
			org,
			repo,
			file,
			line,
			escapeCSV(fb.Reason),
			fb.CreatedAt.Format("2006-01-02T15:04:05Z"),
		)
	}

	// Write to export file
	path := s.exportPath("csv")
	if err := os.WriteFile(path, []byte(csv), 0644); err != nil {
		return "", fmt.Errorf("writing CSV export: %w", err)
	}

	return path, nil
}

// ExportJSON exports feedback to JSON format (for rule training)
func (s *Storage) ExportJSON() (string, error) {
	store, err := s.Load()
	if err != nil {
		return "", err
	}

	// Convert to export format - define ExportEvidence first since ExportEntry references it
	type ExportEvidence struct {
		RuleID      string `json:"rule_id"`
		RuleSource  string `json:"rule_source"`
		GitHubOrg   string `json:"github_org"`
		GitHubRepo  string `json:"github_repo"`
		FilePath    string `json:"file_path"`
		LineStart   int    `json:"line_start"`
		MatchedText string `json:"matched_text"`
	}
	type ExportEntry struct {
		Fingerprint     string           `json:"fingerprint"`
		Verdict         Verdict          `json:"verdict"`
		Reason          string           `json:"reason"`
		Evidence        *ExportEvidence  `json:"evidence"`
		RuleImprovement *RuleImprovement `json:"rule_improvement,omitempty"`
	}

	export := struct {
		Version string        `json:"version"`
		Stats   *FeedbackStats `json:"stats"`
		Entries []ExportEntry `json:"entries"`
	}{
		Version: "1.0",
		Stats:   store.Stats,
		Entries: make([]ExportEntry, 0, len(store.Entries)),
	}

	for _, fb := range store.Entries {
		entry := ExportEntry{
			Fingerprint:     fb.Fingerprint,
			Verdict:         fb.Verdict,
			Reason:          fb.Reason,
			RuleImprovement: fb.RuleImprovement,
		}
		if fb.Evidence != nil {
			entry.Evidence = &ExportEvidence{
				RuleID:      fb.Evidence.RuleID,
				RuleSource:  fb.Evidence.RuleSource,
				GitHubOrg:   fb.Evidence.GitHubOrg,
				GitHubRepo:  fb.Evidence.GitHubRepo,
				FilePath:    fb.Evidence.FilePath,
				LineStart:   fb.Evidence.LineStart,
				MatchedText: fb.Evidence.MatchedText,
			}
		}
		export.Entries = append(export.Entries, entry)
	}

	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling export: %w", err)
	}

	path := s.exportPath("json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("writing JSON export: %w", err)
	}

	return path, nil
}

// GetFalsePositiveRules returns rules with high false positive rates
func (s *Storage) GetFalsePositiveRules(threshold float64) ([]RuleFPStats, error) {
	store, err := s.Load()
	if err != nil {
		return nil, err
	}

	// Aggregate by rule
	ruleStats := make(map[string]*RuleFPStats)
	for _, fb := range store.Entries {
		if fb.Evidence == nil || fb.Evidence.RuleID == "" {
			continue
		}

		ruleID := fb.Evidence.RuleID
		if _, ok := ruleStats[ruleID]; !ok {
			ruleStats[ruleID] = &RuleFPStats{RuleID: ruleID}
		}

		switch fb.Verdict {
		case VerdictFalsePositive:
			ruleStats[ruleID].FalsePositives++
		case VerdictTruePositive:
			ruleStats[ruleID].TruePositives++
		}
	}

	// Calculate rates and filter
	var results []RuleFPStats
	for _, stats := range ruleStats {
		total := stats.FalsePositives + stats.TruePositives
		if total > 0 {
			stats.Total = total
			stats.FPRate = float64(stats.FalsePositives) / float64(total)
			if stats.FPRate >= threshold {
				results = append(results, *stats)
			}
		}
	}

	return results, nil
}

// RuleFPStats contains false positive statistics for a rule
type RuleFPStats struct {
	RuleID         string  `json:"rule_id"`
	FalsePositives int     `json:"false_positives"`
	TruePositives  int     `json:"true_positives"`
	Total          int     `json:"total"`
	FPRate         float64 `json:"fp_rate"`
}

// escapeCSV escapes a string for CSV output
func escapeCSV(s string) string {
	// Replace double quotes with two double quotes
	result := ""
	for _, c := range s {
		if c == '"' {
			result += "\"\""
		} else if c == '\n' {
			result += " "
		} else {
			result += string(c)
		}
	}
	return result
}
