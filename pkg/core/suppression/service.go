// Package suppression provides filtering of findings based on user feedback and context
package suppression

import (
	"github.com/crashappsec/zero/pkg/core/feedback"
	"github.com/crashappsec/zero/pkg/core/findings"
)

// Service filters findings based on user feedback and context
type Service struct {
	storage      *feedback.Storage
	contextRules []ContextRule
}

// ContextRule defines a rule for suppressing findings by context
type ContextRule struct {
	Name        string                         // Rule name for logging
	Condition   func(findings.Context) bool    // When to apply suppression
	Severity    string                         // Only apply to findings at or below this severity
	Description string                         // Human-readable explanation
}

// Result contains the result of filtering findings
type Result struct {
	Original   int            // Total findings before filtering
	Suppressed int            // Number suppressed
	Remaining  int            // Findings after filtering
	Reasons    map[string]int // Count by suppression reason
	Details    []Detail       // Details for each suppressed finding
}

// Detail captures why a finding was suppressed
type Detail struct {
	Fingerprint string // Finding fingerprint
	Reason      string // Why it was suppressed
	Source      string // "feedback" or "context_rule"
	RuleName    string // Context rule name if applicable
}

// NewService creates a new suppression service
func NewService(storage *feedback.Storage) *Service {
	return &Service{
		storage:      storage,
		contextRules: DefaultContextRules(),
	}
}

// DefaultContextRules returns the default context-based suppression rules
func DefaultContextRules() []ContextRule {
	return []ContextRule{
		{
			Name:        "test_file",
			Condition:   func(c findings.Context) bool { return c.InTest },
			Severity:    "medium", // Only suppress medium and below in tests
			Description: "Finding is in a test file",
		},
		{
			Name:        "example_file",
			Condition:   func(c findings.Context) bool { return c.InExample },
			Severity:    "high", // Suppress up to high in examples
			Description: "Finding is in an example file",
		},
		{
			Name:        "documentation",
			Condition:   func(c findings.Context) bool { return c.InDocs },
			Severity:    "high", // Suppress up to high in docs
			Description: "Finding is in documentation",
		},
		{
			Name:        "comment",
			Condition:   func(c findings.Context) bool { return c.InComment },
			Severity:    "medium", // Only suppress medium and below in comments
			Description: "Finding is in a comment",
		},
	}
}

// WithContextRules sets custom context rules
func (s *Service) WithContextRules(rules []ContextRule) *Service {
	s.contextRules = rules
	return s
}

// FilterFindings filters a slice of findings, removing known false positives
// Returns the filtered slice and suppression statistics
func (s *Service) FilterFindings(findingsList []findings.Finding, project string) ([]findings.Finding, *Result) {
	result := &Result{
		Original: len(findingsList),
		Reasons:  make(map[string]int),
		Details:  make([]Detail, 0),
	}

	// Get false positives from feedback store
	fpFingerprints := make(map[string]bool)
	if s.storage != nil {
		fps, err := s.storage.QueryFeedback(feedback.FeedbackQuery{
			Verdict: feedback.VerdictFalsePositive,
		})
		if err == nil {
			for _, fp := range fps {
				fpFingerprints[fp.Fingerprint] = true
			}
		}
	}

	filtered := make([]findings.Finding, 0, len(findingsList))

	for _, f := range findingsList {
		// Get fingerprint from evidence or ID
		fingerprint := f.ID
		if f.Evidence != nil && f.Evidence.Fingerprint != "" {
			fingerprint = f.Evidence.Fingerprint
		}

		// Check feedback store first
		if fpFingerprints[fingerprint] {
			result.Suppressed++
			result.Reasons["feedback_false_positive"]++
			result.Details = append(result.Details, Detail{
				Fingerprint: fingerprint,
				Reason:      "marked as false positive",
				Source:      "feedback",
			})
			continue
		}

		// Check context rules
		suppressed := false
		if f.Evidence != nil {
			ctx := findings.DetectContext(f.Evidence.FilePath, f.Evidence.MatchedText)
			for _, rule := range s.contextRules {
				if rule.Condition(ctx) && shouldSuppressBySeverity(f.Severity.String(), rule.Severity) {
					result.Suppressed++
					result.Reasons["context_"+rule.Name]++
					result.Details = append(result.Details, Detail{
						Fingerprint: fingerprint,
						Reason:      rule.Description,
						Source:      "context_rule",
						RuleName:    rule.Name,
					})
					suppressed = true
					break
				}
			}
		}

		if !suppressed {
			filtered = append(filtered, f)
		}
	}

	result.Remaining = len(filtered)
	return filtered, result
}

// FilterGenericFindings filters findings using a generic interface for scanner integration
func (s *Service) FilterGenericFindings(findingsList []map[string]interface{}, project string) ([]map[string]interface{}, *Result) {
	result := &Result{
		Original: len(findingsList),
		Reasons:  make(map[string]int),
		Details:  make([]Detail, 0),
	}

	// Get false positives from feedback store
	fpFingerprints := make(map[string]bool)
	if s.storage != nil {
		fps, err := s.storage.QueryFeedback(feedback.FeedbackQuery{
			Verdict: feedback.VerdictFalsePositive,
		})
		if err == nil {
			for _, fp := range fps {
				fpFingerprints[fp.Fingerprint] = true
			}
		}
	}

	filtered := make([]map[string]interface{}, 0, len(findingsList))

	for _, f := range findingsList {
		fingerprint := getStringField(f, "fingerprint")

		// Check feedback store first
		if fingerprint != "" && fpFingerprints[fingerprint] {
			result.Suppressed++
			result.Reasons["feedback_false_positive"]++
			result.Details = append(result.Details, Detail{
				Fingerprint: fingerprint,
				Reason:      "marked as false positive",
				Source:      "feedback",
			})
			continue
		}

		// Check context rules
		suppressed := false
		filePath := getStringField(f, "file_path")
		if filePath == "" {
			// Try nested evidence structure
			if evidence, ok := f["evidence"].(map[string]interface{}); ok {
				filePath = getStringField(evidence, "file_path")
			}
		}

		matchedText := getStringField(f, "matched_text")
		if matchedText == "" {
			if evidence, ok := f["evidence"].(map[string]interface{}); ok {
				matchedText = getStringField(evidence, "matched_text")
			}
		}

		if filePath != "" {
			ctx := findings.DetectContext(filePath, matchedText)
			severity := getStringField(f, "severity")

			for _, rule := range s.contextRules {
				if rule.Condition(ctx) && shouldSuppressBySeverity(severity, rule.Severity) {
					result.Suppressed++
					result.Reasons["context_"+rule.Name]++
					result.Details = append(result.Details, Detail{
						Fingerprint: fingerprint,
						Reason:      rule.Description,
						Source:      "context_rule",
						RuleName:    rule.Name,
					})
					suppressed = true
					break
				}
			}
		}

		if !suppressed {
			filtered = append(filtered, f)
		}
	}

	result.Remaining = len(filtered)
	return filtered, result
}

// IsSuppressed checks if a single finding should be suppressed
func (s *Service) IsSuppressed(fingerprint string, filePath string, severity string, lineContent string) (bool, string) {
	// Check feedback store
	if s.storage != nil {
		fb, err := s.storage.GetFeedback(fingerprint)
		if err == nil && fb != nil && fb.Verdict == feedback.VerdictFalsePositive {
			return true, "marked as false positive in feedback"
		}
	}

	// Check context rules
	ctx := findings.DetectContext(filePath, lineContent)
	for _, rule := range s.contextRules {
		if rule.Condition(ctx) && shouldSuppressBySeverity(severity, rule.Severity) {
			return true, rule.Description
		}
	}

	return false, ""
}

// shouldSuppressBySeverity checks if finding severity should be suppressed given rule threshold
func shouldSuppressBySeverity(findingSeverity, ruleMaxSeverity string) bool {
	severityOrder := map[string]int{
		"info":     0,
		"low":      1,
		"medium":   2,
		"high":     3,
		"critical": 4,
	}

	findingLevel, ok := severityOrder[findingSeverity]
	if !ok {
		return false // Unknown severity, don't suppress
	}

	ruleLevel, ok := severityOrder[ruleMaxSeverity]
	if !ok {
		return false // Unknown threshold, don't suppress
	}

	return findingLevel <= ruleLevel
}

// getStringField safely extracts a string from a map
func getStringField(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
