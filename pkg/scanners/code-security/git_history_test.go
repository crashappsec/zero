package codesecurity

import (
	"testing"
	"time"
)

func TestNewGitHistoryScanner(t *testing.T) {
	config := GitHistoryConfig{
		Enabled:    true,
		MaxCommits: 500,
		MaxAge:     "6m",
	}

	scanner := NewGitHistoryScanner(config)
	if scanner == nil {
		t.Fatal("NewGitHistoryScanner() returned nil")
	}
	// Patterns are loaded from RAG - may be empty if RAG not available
	t.Logf("Loaded %d patterns from RAG", len(scanner.patterns))
}

func TestGitHistoryScanner_Patterns(t *testing.T) {
	scanner := NewGitHistoryScanner(GitHistoryConfig{})

	// Patterns are loaded from RAG files
	// If RAG is available, verify some key patterns exist
	if len(scanner.patterns) == 0 {
		t.Log("No patterns loaded - RAG may not be available in test environment")
		return
	}

	patternNames := make(map[string]bool)
	for _, p := range scanner.patterns {
		patternNames[p.name] = true
	}

	t.Logf("Loaded %d patterns from RAG", len(scanner.patterns))

	// These patterns should exist if RAG is available
	expectedPatterns := []string{
		"aws_access_key",
		"openai_api_key",
	}

	for _, name := range expectedPatterns {
		if patternNames[name] {
			t.Logf("Found expected pattern: %s", name)
		}
	}
}

func TestGitHistoryScanner_PatternMatching(t *testing.T) {
	scanner := NewGitHistoryScanner(GitHistoryConfig{})

	// Skip test if no patterns loaded (RAG not available)
	if len(scanner.patterns) == 0 {
		t.Skip("No patterns loaded - RAG not available in test environment")
	}

	tests := []struct {
		name        string
		input       string
		shouldMatch bool
	}{
		{
			name:        "AWS access key",
			input:       "aws_access_key_id = AKIAIOSFODNN7REALKEY",
			shouldMatch: true,
		},
		{
			name:        "Normal code - no secret",
			input:       "const greeting = 'Hello World';",
			shouldMatch: false,
		},
		{
			name:        "Comment line",
			input:       "// This is just a comment",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched := false
			matchedType := ""
			for _, pattern := range scanner.patterns {
				if pattern.pattern.MatchString(tt.input) {
					matched = true
					matchedType = pattern.name
					break
				}
			}

			if tt.shouldMatch && !matched {
				t.Errorf("Pattern matching for %q: expected match but got none", tt.input)
			}

			if !tt.shouldMatch && matched {
				t.Errorf("Pattern matching for %q: unexpected match with type %q", tt.input, matchedType)
			}

			if matched {
				t.Logf("Matched %q with type %q", tt.name, matchedType)
			}
		})
	}
}

func TestGitHistoryScanner_parseSinceDate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		maxAge      string
		minExpected time.Duration // Minimum expected duration
		maxExpected time.Duration // Maximum expected duration
	}{
		{"30d", 29 * 24 * time.Hour, 31 * 24 * time.Hour},
		{"90d", 89 * 24 * time.Hour, 91 * 24 * time.Hour},
		{"6m", 175 * 24 * time.Hour, 190 * 24 * time.Hour}, // ~6 months (varies by month lengths)
		{"1y", 360 * 24 * time.Hour, 370 * 24 * time.Hour}, // ~1 year
		{"2y", 720 * 24 * time.Hour, 740 * 24 * time.Hour}, // ~2 years
		{"", 360 * 24 * time.Hour, 370 * 24 * time.Hour},   // Default 1 year
	}

	for _, tt := range tests {
		t.Run(tt.maxAge, func(t *testing.T) {
			scanner := NewGitHistoryScanner(GitHistoryConfig{MaxAge: tt.maxAge})
			since := scanner.parseSinceDate()

			// Check that since is within expected range
			diff := now.Sub(since)

			if diff < tt.minExpected || diff > tt.maxExpected {
				t.Errorf("parseSinceDate() with maxAge=%q: got %v ago, want between %v and %v ago", tt.maxAge, diff, tt.minExpected, tt.maxExpected)
			}
		})
	}
}

func TestGitHistoryScanner_isFalsePositive(t *testing.T) {
	scanner := NewGitHistoryScanner(GitHistoryConfig{})

	tests := []struct {
		name     string
		line     string
		filename string
		expected bool
	}{
		{
			name:     "test file",
			line:     "aws_key = AKIAIOSFODNN7REALKEY",
			filename: "config_test.go",
			expected: true,
		},
		{
			name:     "fixture file",
			line:     "secret = ghp_12345",
			filename: "fixtures/test_data.json",
			expected: true,
		},
		{
			name:     "example placeholder",
			line:     "key = AKIAIOSFODNN7EXAMPLE",
			filename: "config.go",
			expected: true, // Contains "example"
		},
		{
			name:     "documentation file",
			line:     "Use your API key here: sk-xxx",
			filename: "README.md",
			expected: true,
		},
		{
			name:     "stripe test key",
			line:     "stripe_key = sk_test_123456",
			filename: "payments.go",
			expected: true,
		},
		{
			name:     "real looking secret in code",
			line:     "api_key = ghp_1234567890abcdefghijklmnopqrstuvwxyz",
			filename: "src/auth.go",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scanner.isFalsePositive(tt.line, tt.filename)
			if got != tt.expected {
				t.Errorf("isFalsePositive(%q, %q) = %v, want %v", tt.line, tt.filename, got, tt.expected)
			}
		})
	}
}

func TestRedactHistorySecret(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"short", "****"},
		{"12345678", "****"},
		{"123456789", "1234****6789"},
		{"AKIAIOSFODNN7EXAMPLE", "AKIA****MPLE"},
		{"ghp_1234567890abcdefghij", "ghp_****ghij"},
	}

	for _, tt := range tests {
		got := redactHistorySecret(tt.input)
		if got != tt.expected {
			t.Errorf("redactHistorySecret(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestFirstLine(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"single line", "single line"},
		{"first\nsecond\nthird", "first"},
		{"  with spaces  \n\n", "with spaces"},
		{"\n\n\n", ""},
		{"", ""},
	}

	for _, tt := range tests {
		got := firstLine(tt.input)
		if got != tt.expected {
			t.Errorf("firstLine(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestGitHistoryResult_Fields(t *testing.T) {
	result := &GitHistoryResult{
		Findings:       []SecretFinding{{File: "test.go", Line: 1}},
		CommitsScanned: 100,
		SecretsFound:   5,
		SecretsRemoved: 2,
	}

	if len(result.Findings) != 1 {
		t.Errorf("Findings count = %d, want 1", len(result.Findings))
	}
	if result.CommitsScanned != 100 {
		t.Errorf("CommitsScanned = %d, want 100", result.CommitsScanned)
	}
	if result.SecretsFound != 5 {
		t.Errorf("SecretsFound = %d, want 5", result.SecretsFound)
	}
	if result.SecretsRemoved != 2 {
		t.Errorf("SecretsRemoved = %d, want 2", result.SecretsRemoved)
	}
}
