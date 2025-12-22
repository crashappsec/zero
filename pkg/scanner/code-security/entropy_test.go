package codesecurity

import (
	"math"
	"testing"
)

func TestCalculateEntropy(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
		delta    float64 // Acceptable error margin
	}{
		{
			name:     "empty string",
			input:    "",
			expected: 0,
			delta:    0.001,
		},
		{
			name:     "single character",
			input:    "a",
			expected: 0,
			delta:    0.001,
		},
		{
			name:     "repeated characters",
			input:    "aaaaaaaaaa",
			expected: 0,
			delta:    0.001,
		},
		{
			name:     "two different chars equally distributed",
			input:    "ababababab",
			expected: 1.0, // log2(2) = 1
			delta:    0.001,
		},
		{
			name:     "high entropy random-looking string",
			input:    "aB3$xY9@mN2!pQ7^",
			expected: 4.0, // High entropy
			delta:    0.5,
		},
		{
			name:     "AWS access key pattern",
			input:    "AKIAIOSFODNN7EXAMPLE",
			expected: 3.5, // Moderate-high entropy
			delta:    0.5,
		},
		{
			name:     "base64-like string",
			input:    "c2VjcmV0X3Bhc3N3b3Jk",
			expected: 3.5,
			delta:    0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateEntropy(tt.input)
			if math.Abs(got-tt.expected) > tt.delta {
				t.Errorf("CalculateEntropy(%q) = %v, want ~%v (±%v)", tt.input, got, tt.expected, tt.delta)
			}
		})
	}
}

func TestEntropyAnalyzer_GetEntropyLevel(t *testing.T) {
	analyzer := NewEntropyAnalyzer(EntropyConfig{
		HighThreshold: 4.5,
		MedThreshold:  3.5,
	})

	tests := []struct {
		entropy  float64
		expected string
	}{
		{5.0, "high"},
		{4.5, "high"},
		{4.0, "medium"},
		{3.5, "medium"},
		{3.0, "low"},
		{2.0, "low"},
		{0, "low"},
	}

	for _, tt := range tests {
		got := analyzer.GetEntropyLevel(tt.entropy)
		if got != tt.expected {
			t.Errorf("GetEntropyLevel(%v) = %q, want %q", tt.entropy, got, tt.expected)
		}
	}
}

func TestEntropyAnalyzer_isFalsePositive(t *testing.T) {
	analyzer := NewEntropyAnalyzer(EntropyConfig{
		MinLength:     16,
		HighThreshold: 4.5,
		MedThreshold:  3.5,
	})

	tests := []struct {
		name     string
		value    string
		context  string
		expected bool
	}{
		{
			name:     "example placeholder",
			value:    "example_secret_key_here",
			context:  "secret = example_secret_key_here",
			expected: true,
		},
		{
			name:     "test placeholder",
			value:    "test_api_key_12345678",
			context:  "apikey = test_api_key_12345678",
			expected: true,
		},
		{
			name:     "AWS example key",
			value:    "AKIAIOSFODNN7EXAMPLE",
			context:  "aws_key = AKIAIOSFODNN7EXAMPLE",
			expected: true,
		},
		{
			name:     "UUID pattern",
			value:    "550e8400-e29b-41d4-a716-446655440000",
			context:  "id = 550e8400-e29b-41d4-a716-446655440000",
			expected: true,
		},
		{
			name:     "git SHA with context",
			value:    "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
			context:  "commit_sha = a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
			expected: true, // sha in context
		},
		{
			name:     "hash context indicator",
			value:    "a1b2c3d4e5f6789012345678901234567890abcd",
			context:  "sha256_hash = a1b2c3d4e5f6789012345678901234567890abcd",
			expected: true, // sha256 in context
		},
		{
			name:     "all same character",
			value:    "aaaaaaaaaaaaaaaa",
			context:  "key = aaaaaaaaaaaaaaaa",
			expected: true,
		},
		{
			name:     "test file indicator",
			value:    "someRandomHighEntropyString123",
			context:  "// in test_secrets.go file",
			expected: true,
		},
		{
			name:     "stripe test key",
			value:    "sk_test_abcdefghijklmnop",
			context:  "stripe_key = sk_test_abcdefghijklmnop",
			expected: true,
		},
		{
			name:     "real looking secret",
			value:    "ghp_1234567890abcdefGHIJK",
			context:  "token = ghp_1234567890abcdefGHIJK",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzer.isFalsePositive(tt.value, tt.context)
			if got != tt.expected {
				t.Errorf("isFalsePositive(%q, %q) = %v, want %v", tt.value, tt.context, got, tt.expected)
			}
		})
	}
}

func TestEntropyAnalyzer_inferSecretType(t *testing.T) {
	analyzer := NewEntropyAnalyzer(EntropyConfig{})

	// Test that inferSecretType returns a string (either from RAG or default)
	// The exact result depends on whether RAG patterns are available
	tests := []struct {
		value   string
		context string
	}{
		{"AKIAIOSFODNN7EXAMPLE", "aws_access_key_id"},
		{"ghp_1234567890abcdefghijklmnopqrstuvwxyz", "github_token"},
		{"sk_live_abcdefghijklmnopqrstuvwx", "stripe_key"},
		{"unknownFormatKey", "unknown context"},
	}

	for _, tt := range tests {
		t.Run(tt.value[:10], func(t *testing.T) {
			got := analyzer.inferSecretType(tt.value, tt.context)
			// Should return a non-empty string
			if got == "" {
				t.Errorf("inferSecretType(%q, %q) returned empty string", tt.value, tt.context)
			}
			// If RAG patterns are not available, should return "high_entropy_string"
			// If RAG patterns are available, should return a specific type
			t.Logf("inferSecretType(%q) = %q", tt.value[:10], got)
		})
	}
}

func TestEntropyAnalyzer_getSeverity(t *testing.T) {
	analyzer := NewEntropyAnalyzer(EntropyConfig{})

	tests := []struct {
		level    string
		expected string
	}{
		{"high", "medium"},
		{"medium", "low"},
		{"low", "info"},
	}

	for _, tt := range tests {
		got := analyzer.getSeverity(tt.level)
		if got != tt.expected {
			t.Errorf("getSeverity(%q) = %q, want %q", tt.level, got, tt.expected)
		}
	}
}

func TestEntropyAnalyzer_redactSnippet(t *testing.T) {
	analyzer := NewEntropyAnalyzer(EntropyConfig{})

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
		got := analyzer.redactSnippet(tt.input)
		if got != tt.expected {
			t.Errorf("redactSnippet(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestIsAllSameChar(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"", true},
		{"a", true},
		{"aaaaaaa", true},
		{"ab", false},
		{"aaaab", false},
	}

	for _, tt := range tests {
		got := isAllSameChar(tt.input)
		if got != tt.expected {
			t.Errorf("isAllSameChar(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestIsUniformAlphanumeric(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"short", false},           // Too short
		{"abcdefghijklmnop", true}, // All letters
		{"1234567890123456", true}, // All digits
		{"abc123def456ghi7", false}, // Mixed - not uniform
	}

	for _, tt := range tests {
		got := isUniformAlphanumeric(tt.input)
		if got != tt.expected {
			t.Errorf("isUniformAlphanumeric(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}
