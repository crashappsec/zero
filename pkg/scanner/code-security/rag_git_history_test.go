package codesecurity

import (
	"regexp"
	"testing"
)

func TestLoadRAGGitHistoryPatterns(t *testing.T) {
	// Clear cache before testing
	ClearRAGGitHistoryCache()

	gitignore, sensitive, err := LoadRAGGitHistoryPatterns()
	if err != nil {
		t.Logf("LoadRAGGitHistoryPatterns returned error: %v (this is OK if RAG not available)", err)
	}

	t.Logf("Loaded %d gitignore patterns and %d sensitive patterns", len(gitignore), len(sensitive))

	// If patterns were loaded, verify they have the expected structure
	if len(gitignore) > 0 {
		for _, p := range gitignore[:min(3, len(gitignore))] {
			t.Logf("Gitignore pattern: %s (category: %s, severity: %s)", p.Name, p.Category, p.Severity)
			if p.Pattern == nil {
				t.Errorf("Pattern %s has nil compiled regex", p.Name)
			}
		}
	}

	if len(sensitive) > 0 {
		for _, p := range sensitive[:min(3, len(sensitive))] {
			t.Logf("Sensitive pattern: %s (category: %s, severity: %s)", p.Name, p.Category, p.Severity)
			if p.Pattern == nil {
				t.Errorf("Pattern %s has nil compiled regex", p.Name)
			}
		}
	}
}

func TestFilepathToRegex(t *testing.T) {
	tests := []struct {
		pattern  string
		input    string
		expected bool
	}{
		{`\.env$`, "project/.env", true},      // Matches file named ".env"
		{`\.env$`, "config/.env", true},       // Matches .env in subdirectory
		{`\.env$`, ".environment", false},     // Does not match .environment
		{`node_modules/`, "node_modules/", true},
		{`node_modules/`, "path/node_modules/package.json", true},
		{`.*\.pem$`, "certs/server.pem", true},  // Pattern with .* matches any .pem file
		{`.*\.pem$`, "server.key", false},       // Does not match non-.pem files
	}

	for _, tt := range tests {
		regex := filepathToRegex(tt.pattern)
		fullRegex := "(?i)" + regex
		compiled, err := compilePatternForTest(fullRegex)
		if err != nil {
			t.Errorf("Failed to compile pattern %q: %v", tt.pattern, err)
			continue
		}

		matched := compiled.MatchString(tt.input)
		t.Logf("Pattern: %q -> Regex: %q, Input: %q, Match: %v", tt.pattern, fullRegex, tt.input, matched)
		if matched != tt.expected {
			t.Errorf("filepathToRegex(%q).MatchString(%q) = %v, want %v", tt.pattern, tt.input, matched, tt.expected)
		}
	}
}

func TestNormalizePatternName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Environment Files", "environment_files"},
		{"SSH Private Keys", "ssh_private_keys"},
		{"AWS-Credentials", "aws_credentials"},
		{"Test-Name_With.Special!Chars", "test_name_withspecialchars"},
	}

	for _, tt := range tests {
		got := normalizePatternName(tt.input)
		if got != tt.expected {
			t.Errorf("normalizePatternName(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestNormalizeGitHistorySeverity(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"critical", "critical"},
		{"CRITICAL", "critical"},
		{"high", "high"},
		{"HIGH", "high"},
		{"medium", "medium"},
		{"low", "low"},
		{"info", "info"},
		{"unknown", "medium"}, // Default
		{"", "medium"},        // Default
	}

	for _, tt := range tests {
		got := normalizeGitHistorySeverity(tt.input)
		if got != tt.expected {
			t.Errorf("normalizeGitHistorySeverity(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestConvertToSensitiveFilePatterns(t *testing.T) {
	ragPatterns := []*RAGGitHistoryPattern{
		{
			Name:        "env_file",
			RawPattern:  `\.env$`,
			Severity:    "critical",
			Category:    "credentials",
			Description: "Environment file",
		},
		{
			Name:        "private_key",
			RawPattern:  `id_rsa$`,
			Severity:    "critical",
			Category:    "keys",
			Description: "SSH private key",
		},
	}

	// Compile patterns
	for _, p := range ragPatterns {
		compiled, err := compilePatternForTest("(?i)" + p.RawPattern)
		if err != nil {
			t.Fatalf("Failed to compile pattern: %v", err)
		}
		p.Pattern = compiled
	}

	converted := ConvertToSensitiveFilePatterns(ragPatterns)

	if len(converted) != len(ragPatterns) {
		t.Errorf("ConvertToSensitiveFilePatterns returned %d patterns, want %d", len(converted), len(ragPatterns))
	}

	for i, c := range converted {
		if c.Category != ragPatterns[i].Category {
			t.Errorf("Pattern %d category = %q, want %q", i, c.Category, ragPatterns[i].Category)
		}
		if c.Severity != ragPatterns[i].Severity {
			t.Errorf("Pattern %d severity = %q, want %q", i, c.Severity, ragPatterns[i].Severity)
		}
	}
}

func TestGetRAGGitHistoryPatternCounts(t *testing.T) {
	// This just tests the function runs without error
	gitignore, sensitive := GetRAGGitHistoryPatternCounts()
	t.Logf("Pattern counts: gitignore=%d, sensitive=%d", gitignore, sensitive)
}

// Helper function for tests
func compilePatternForTest(pattern string) (*regexp.Regexp, error) {
	return regexp.Compile(pattern)
}

// min helper for older Go versions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
