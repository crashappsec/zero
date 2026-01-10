package suppression

import (
	"testing"

	"github.com/crashappsec/zero/pkg/core/findings"
)

func TestShouldSuppressBySeverity(t *testing.T) {
	tests := []struct {
		name            string
		findingSeverity string
		ruleSeverity    string
		expected        bool
	}{
		{"low finding, medium rule", "low", "medium", true},
		{"medium finding, medium rule", "medium", "medium", true},
		{"high finding, medium rule", "high", "medium", false},
		{"critical finding, high rule", "critical", "high", false},
		{"info finding, low rule", "info", "low", true},
		{"unknown finding", "unknown", "medium", false},
		{"unknown rule", "medium", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldSuppressBySeverity(tt.findingSeverity, tt.ruleSeverity)
			if result != tt.expected {
				t.Errorf("shouldSuppressBySeverity(%q, %q) = %v, want %v",
					tt.findingSeverity, tt.ruleSeverity, result, tt.expected)
			}
		})
	}
}

func TestServiceFilterGenericFindings(t *testing.T) {
	svc := NewService(nil) // No feedback storage

	findingsList := []map[string]interface{}{
		{
			"fingerprint": "abc123",
			"file_path":   "src/main.go",
			"severity":    "high",
		},
		{
			"fingerprint": "def456",
			"file_path":   "pkg/util_test.go", // _test. pattern matches
			"severity":    "low",              // Low in test should be suppressed
		},
		{
			"fingerprint": "ghi789",
			"file_path":   "pkg/util_test.go", // _test. pattern matches
			"severity":    "critical",         // Critical in test should NOT be suppressed
		},
		{
			"fingerprint": "jkl012",
			"file_path":   "/examples/demo.go", // /examples/ pattern matches
			"severity":    "medium",            // Medium in example should be suppressed
		},
	}

	filtered, result := svc.FilterGenericFindings(findingsList, "test/repo")

	if result.Original != 4 {
		t.Errorf("Original = %d, want 4", result.Original)
	}

	if result.Suppressed != 2 {
		t.Errorf("Suppressed = %d, want 2", result.Suppressed)
	}

	if result.Remaining != 2 {
		t.Errorf("Remaining = %d, want 2", result.Remaining)
	}

	if len(filtered) != 2 {
		t.Errorf("len(filtered) = %d, want 2", len(filtered))
	}

	// Verify the right ones were kept
	kept := make(map[string]bool)
	for _, f := range filtered {
		kept[f["fingerprint"].(string)] = true
	}

	if !kept["abc123"] {
		t.Error("abc123 (high in src) should not be suppressed")
	}
	if !kept["ghi789"] {
		t.Error("ghi789 (critical in test) should not be suppressed")
	}
}

func TestServiceIsSuppressed(t *testing.T) {
	svc := NewService(nil)

	tests := []struct {
		name        string
		fingerprint string
		filePath    string
		severity    string
		lineContent string
		expectSupp  bool
	}{
		{
			name:       "high in source",
			filePath:   "src/app.go",
			severity:   "high",
			expectSupp: false,
		},
		{
			name:       "low in test",
			filePath:   "pkg/util_test.go",
			severity:   "low",
			expectSupp: true,
		},
		{
			name:       "critical in test",
			filePath:   "pkg/util_test.go",
			severity:   "critical",
			expectSupp: false,
		},
		{
			name:        "medium in comment",
			filePath:    "src/config.go",
			severity:    "medium",
			lineContent: "// password = 'test123'",
			expectSupp:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suppressed, _ := svc.IsSuppressed(tt.fingerprint, tt.filePath, tt.severity, tt.lineContent)
			if suppressed != tt.expectSupp {
				t.Errorf("IsSuppressed() = %v, want %v", suppressed, tt.expectSupp)
			}
		})
	}
}

func TestDefaultContextRules(t *testing.T) {
	rules := DefaultContextRules()

	if len(rules) != 4 {
		t.Errorf("Expected 4 default rules, got %d", len(rules))
	}

	// Verify each rule has required fields
	for _, rule := range rules {
		if rule.Name == "" {
			t.Error("Rule missing name")
		}
		if rule.Condition == nil {
			t.Errorf("Rule %q missing condition", rule.Name)
		}
		if rule.Description == "" {
			t.Errorf("Rule %q missing description", rule.Name)
		}
		if rule.Severity == "" {
			t.Errorf("Rule %q missing severity", rule.Name)
		}
	}
}

func TestServiceWithContextRules(t *testing.T) {
	svc := NewService(nil)

	// Custom rules - only suppress critical in tests
	customRules := []ContextRule{
		{
			Name:        "test_critical",
			Condition:   func(c findings.Context) bool { return c.InTest },
			Severity:    "critical",
			Description: "Allow everything in tests",
		},
	}

	svc.WithContextRules(customRules)

	// Critical in test should now be suppressed
	suppressed, _ := svc.IsSuppressed("", "test_file.go", "critical", "")
	if suppressed {
		t.Error("Critical should NOT be suppressed with custom rule (severity threshold is max allowed)")
	}
}
