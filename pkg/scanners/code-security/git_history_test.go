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
	if len(scanner.patterns) == 0 {
		t.Error("NewGitHistoryScanner() has no patterns initialized")
	}
}

func TestGitHistoryScanner_Patterns(t *testing.T) {
	scanner := NewGitHistoryScanner(GitHistoryConfig{})

	// Verify key patterns exist
	patternNames := make(map[string]bool)
	for _, p := range scanner.patterns {
		patternNames[p.name] = true
	}

	expectedPatterns := []string{
		"aws_access_key",
		"github_token",
		"stripe_secret_key",
		"slack_token",
		"openai_api_key",
		"private_key",
		"jwt_token",
		"database_url",
	}

	for _, name := range expectedPatterns {
		if !patternNames[name] {
			t.Errorf("Expected pattern %q not found", name)
		}
	}
}

func TestGitHistoryScanner_PatternMatching(t *testing.T) {
	scanner := NewGitHistoryScanner(GitHistoryConfig{})

	tests := []struct {
		name        string
		input       string
		shouldMatch bool
		secretType  string
	}{
		{
			name:        "AWS access key",
			input:       "aws_access_key_id = AKIAIOSFODNN7REALKEY",
			shouldMatch: true,
			secretType:  "aws_access_key",
		},
		{
			name:        "GitHub PAT",
			input:       "token = ghp_1234567890abcdefghijklmnopqrstuvwxyz",
			shouldMatch: true,
			secretType:  "github_token",
		},
		{
			name:        "Stripe live key",
			input:       "stripe_key = sk_live_1234567890abcdefghijklmn",
			shouldMatch: true,
			secretType:  "stripe_secret_key",
		},
		{
			name:        "Slack bot token",
			input:       "slack = xoxb-1234567890-abcdefghij",
			shouldMatch: true,
			secretType:  "slack_token",
		},
		{
			name:        "OpenAI key",
			input:       "openai_key = sk-1234567890abcdefghijklmnopqrstuvwxyzabcdefghijklmnop",
			shouldMatch: true,
			secretType:  "openai_api_key",
		},
		{
			name:        "Private key header",
			input:       "-----BEGIN RSA PRIVATE KEY-----",
			shouldMatch: true,
			secretType:  "private_key",
		},
		{
			name:        "JWT token",
			input:       "token = eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
			shouldMatch: true,
			secretType:  "jwt_token",
		},
		{
			name:        "Database URL with creds",
			input:       "database_url = postgres://admin:secretpass@localhost:5432/db",
			shouldMatch: true,
			secretType:  "database_url",
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

			if matched != tt.shouldMatch {
				t.Errorf("Pattern matching for %q: got matched=%v, want matched=%v", tt.input, matched, tt.shouldMatch)
			}

			if tt.shouldMatch && matchedType != tt.secretType {
				t.Errorf("Pattern type for %q: got %q, want %q", tt.input, matchedType, tt.secretType)
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
