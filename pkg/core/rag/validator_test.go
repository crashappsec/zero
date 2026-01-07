package rag

import (
	"testing"
)

func TestNewValidator(t *testing.T) {
	v := NewValidator()
	if v == nil {
		t.Fatal("NewValidator returned nil")
	}
	if v.StrictMode {
		t.Error("StrictMode should be false by default")
	}
}

func TestValidateRegex(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{"valid simple", `foo`, false},
		{"valid with groups", `(foo|bar)`, false},
		{"valid with quantifiers", `foo+bar*`, false},
		{"valid import pattern", `import\s+.*\s+from\s+['"]react['"]`, false},
		{"valid short pattern", `api`, false},
		{"empty pattern", ``, true},
		{"invalid unclosed group", `(foo`, true},
		{"invalid bad quantifier", `*foo`, true},
		{"invalid unclosed bracket", `[abc`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateRegex(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRegex(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}

func TestValidateSeverity(t *testing.T) {
	v := NewValidator()

	validSeverities := []string{"critical", "high", "medium", "low", "info", "CRITICAL", "High", "MEDIUM"}
	for _, sev := range validSeverities {
		if !v.ValidateSeverity(sev) {
			t.Errorf("ValidateSeverity(%q) = false, want true", sev)
		}
	}

	invalidSeverities := []string{"severe", "warning", "error", ""}
	for _, sev := range invalidSeverities {
		if v.ValidateSeverity(sev) {
			t.Errorf("ValidateSeverity(%q) = true, want false", sev)
		}
	}
}

func TestValidatePattern(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name       string
		pattern    Pattern
		wantErrors int
		errorField string
	}{
		{
			name: "valid pattern",
			pattern: Pattern{
				ID:         "test-1",
				Pattern:    `import\s+React`,
				Type:       "regex",
				Severity:   "high",
				Message:    "React import detected",
				Languages:  []string{"javascript"},
				Confidence: 90,
			},
			wantErrors: 0,
		},
		{
			name: "empty pattern field",
			pattern: Pattern{
				ID:       "test-2",
				Pattern:  "",
				Type:     "regex",
				Severity: "high",
				Message:  "Test",
			},
			wantErrors: 2, // empty pattern + invalid regex
			errorField: "pattern",
		},
		{
			name: "invalid type",
			pattern: Pattern{
				ID:       "test-3",
				Pattern:  "test",
				Type:     "invalid",
				Severity: "high",
				Message:  "Test",
			},
			wantErrors: 1,
			errorField: "type",
		},
		{
			name: "invalid severity",
			pattern: Pattern{
				ID:       "test-4",
				Pattern:  "test",
				Type:     "regex",
				Severity: "severe",
				Message:  "Test",
			},
			wantErrors: 1,
			errorField: "severity",
		},
		{
			name: "invalid regex",
			pattern: Pattern{
				ID:       "test-5",
				Pattern:  "(unclosed",
				Type:     "regex",
				Severity: "high",
				Message:  "Test",
			},
			wantErrors: 1,
			errorField: "pattern",
		},
		{
			name: "unsupported language",
			pattern: Pattern{
				ID:        "test-6",
				Pattern:   "test",
				Type:      "regex",
				Severity:  "high",
				Message:   "Test",
				Languages: []string{"brainfuck"},
			},
			wantErrors: 1,
			errorField: "languages",
		},
		{
			name: "confidence out of range",
			pattern: Pattern{
				ID:         "test-7",
				Pattern:    "test",
				Type:       "regex",
				Severity:   "high",
				Message:    "Test",
				Confidence: 150,
			},
			wantErrors: 1,
			errorField: "confidence",
		},
		{
			name: "missing message",
			pattern: Pattern{
				ID:       "test-8",
				Pattern:  "test",
				Type:     "regex",
				Severity: "high",
				Message:  "",
			},
			wantErrors: 1,
			errorField: "message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := v.ValidatePattern(tt.pattern, "test.md")
			if len(errors) != tt.wantErrors {
				t.Errorf("ValidatePattern() got %d errors, want %d: %+v", len(errors), tt.wantErrors, errors)
			}
			if tt.errorField != "" && len(errors) > 0 {
				found := false
				for _, e := range errors {
					if e.Field == tt.errorField {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("ValidatePattern() expected error on field %q, got %+v", tt.errorField, errors)
				}
			}
		})
	}
}

func TestValidatePatternSet(t *testing.T) {
	v := NewValidator()

	ps := PatternSet{
		Source: "test-patterns.md",
		Patterns: []Pattern{
			{ID: "valid-1", Pattern: "test1", Type: "regex", Severity: "high", Message: "Test 1"},
			{ID: "valid-2", Pattern: "test2", Type: "regex", Severity: "medium", Message: "Test 2"},
			{ID: "invalid-1", Pattern: "(unclosed", Type: "regex", Severity: "high", Message: "Test 3"},
		},
	}

	result := v.ValidatePatternSet(ps)

	if result.TotalPatterns != 3 {
		t.Errorf("TotalPatterns = %d, want 3", result.TotalPatterns)
	}
	if result.ValidPatterns != 2 {
		t.Errorf("ValidPatterns = %d, want 2", result.ValidPatterns)
	}
	if result.InvalidPatterns != 1 {
		t.Errorf("InvalidPatterns = %d, want 1", result.InvalidPatterns)
	}
	if len(result.Errors) != 1 {
		t.Errorf("Errors count = %d, want 1", len(result.Errors))
	}
	if result.PatternsByType["regex"] != 3 {
		t.Errorf("PatternsByType[regex] = %d, want 3", result.PatternsByType["regex"])
	}
}

func TestValidationResultMethods(t *testing.T) {
	t.Run("IsValid", func(t *testing.T) {
		valid := &ValidationResult{Errors: nil}
		if !valid.IsValid() {
			t.Error("IsValid() = false for result with no errors")
		}

		invalid := &ValidationResult{Errors: []ValidationError{{Message: "test"}}}
		if invalid.IsValid() {
			t.Error("IsValid() = true for result with errors")
		}
	})

	t.Run("ValidationRate", func(t *testing.T) {
		r := &ValidationResult{
			TotalPatterns: 100,
			ValidPatterns: 90,
		}
		rate := r.ValidationRate()
		if rate != 90.0 {
			t.Errorf("ValidationRate() = %f, want 90.0", rate)
		}

		empty := &ValidationResult{TotalPatterns: 0}
		if empty.ValidationRate() != 100.0 {
			t.Errorf("ValidationRate() for empty = %f, want 100.0", empty.ValidationRate())
		}
	})

	t.Run("Summary", func(t *testing.T) {
		r := &ValidationResult{
			TotalPatterns:      10,
			ValidPatterns:      8,
			InvalidPatterns:    2,
			Errors:             []ValidationError{{Message: "err1"}, {Message: "err2"}},
			Warnings:           []ValidationError{{Message: "warn1"}},
			PatternsByType:     map[string]int{"regex": 8, "semgrep": 2},
			PatternsBySeverity: map[string]int{"high": 5, "medium": 5},
		}
		summary := r.Summary()
		if summary == "" {
			t.Error("Summary() returned empty string")
		}
		// Verify key content is present
		if !containsAll(summary, "Total Patterns", "Valid:", "Invalid:", "Errors:", "Warnings:", "By Type:", "By Severity:") {
			t.Errorf("Summary() missing expected sections: %s", summary)
		}
	})
}

func TestStrictMode(t *testing.T) {
	v := NewValidator()
	v.StrictMode = true

	// Pattern with warning-level issue (invalid severity becomes error in strict mode)
	p := Pattern{
		ID:       "test-strict",
		Pattern:  "test",
		Type:     "regex",
		Severity: "invalid-severity",
		Message:  "Test",
	}

	ps := PatternSet{Source: "test.md", Patterns: []Pattern{p}}
	result := v.ValidatePatternSet(ps)

	// In strict mode, the invalid severity warning should become an error
	if result.InvalidPatterns != 1 {
		t.Errorf("StrictMode: InvalidPatterns = %d, want 1", result.InvalidPatterns)
	}
	if len(result.Errors) == 0 {
		t.Error("StrictMode: expected errors but got none")
	}
}

func TestSupportedLanguages(t *testing.T) {
	expected := []string{
		"python", "javascript", "typescript", "go", "ruby",
		"java", "kotlin", "scala", "csharp", "php", "rust",
		"c", "cpp", "swift", "bash", "shell", "powershell", "generic",
	}

	for _, lang := range expected {
		if !SupportedLanguages[lang] {
			t.Errorf("Language %q should be supported", lang)
		}
	}
}

func TestValidPatternTypes(t *testing.T) {
	expected := []string{"regex", "semgrep", "glob"}
	for _, ptype := range expected {
		if !ValidPatternTypes[ptype] {
			t.Errorf("Pattern type %q should be valid", ptype)
		}
	}
}

// Helper function
func containsAll(s string, substrs ...string) bool {
	for _, sub := range substrs {
		found := false
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
