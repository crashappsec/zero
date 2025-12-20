package codesecurity

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRAGSecretPatterns(t *testing.T) {
	// Clear cache before testing
	ClearRAGSecretsCache()

	patterns, err := LoadRAGSecretPatterns()
	if err != nil {
		t.Logf("LoadRAGSecretPatterns() returned error (may be expected if RAG not found): %v", err)
	}

	// In test environment, RAG may or may not be available
	// If available, we should have some patterns
	if patterns != nil && len(patterns) > 0 {
		t.Logf("Loaded %d secret patterns from RAG", len(patterns))

		// Verify patterns have required fields
		for i, p := range patterns {
			if i > 10 {
				break // Only check first 10 patterns
			}
			if p.Name == "" {
				t.Errorf("Pattern %d has empty Name", i)
			}
			if p.Pattern == nil {
				t.Errorf("Pattern %d (%s) has nil Pattern", i, p.Name)
			}
			if p.Severity == "" {
				t.Errorf("Pattern %d (%s) has empty Severity", i, p.Name)
			}
		}

		// Check for some expected patterns
		patternNames := make(map[string]bool)
		for _, p := range patterns {
			patternNames[p.Name] = true
		}

		// These should exist if AWS, GitHub, etc. RAG files are present
		expectedPatterns := []string{"aws_access_key", "openai_api_key"}
		for _, name := range expectedPatterns {
			if patternNames[name] {
				t.Logf("Found expected pattern: %s", name)
			}
		}
	} else {
		t.Log("No RAG patterns loaded (RAG directory may not be available)")
	}
}

func TestNormalizeSecretName(t *testing.T) {
	tests := []struct {
		name       string
		technology string
		expected   string
	}{
		{"AWS Access Key ID", "aws", "aws_access_key"},
		{"AWS Secret Access Key", "aws", "aws_secret_key"},
		{"API Key", "openai", "openai_api_key"},
		{"API Key", "anthropic", "anthropic_api_key"},
		{"Personal Access Token", "github", "github_token"},
		{"Secret Key", "stripe", "stripe_secret_key"},
		{"Bot Token", "slack", "slack_token"},
		{"RSA Private Key", "crypto", "private_key"},
		{"JWT Token", "jwt", "jwt_token"},
		{"Database Connection String", "postgres", "database_credential"},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_"+tt.technology, func(t *testing.T) {
			got := normalizeSecretName(tt.name, tt.technology)
			if got != tt.expected {
				t.Errorf("normalizeSecretName(%q, %q) = %q, want %q", tt.name, tt.technology, got, tt.expected)
			}
		})
	}
}

func TestNormalizeSeverity(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"critical", "critical"},
		{"CRITICAL", "critical"},
		{"high", "high"},
		{"HIGH", "high"},
		{"medium", "medium"},
		{"MEDIUM", "medium"},
		{"low", "low"},
		{"LOW", "low"},
		{"", "medium"},
		{"unknown", "medium"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeSeverity(tt.input)
			if got != tt.expected {
				t.Errorf("normalizeSeverity(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseSecretsFromFile(t *testing.T) {
	// Create a temporary pattern file for testing
	tmpDir, err := os.MkdirTemp("", "rag_secrets_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test patterns.md file
	content := `# Test Technology

**Category**: test-category
**Description**: Test technology for unit tests

---

## Secrets Detection

### API Keys and Credentials

#### Test API Key
**Pattern**: ` + "`" + `test_[A-Za-z0-9]{32}` + "`" + `
**Severity**: high
**Description**: Test API key pattern for unit testing

#### Test Token
**Pattern**: ` + "`" + `tok_test_[A-Za-z0-9]{24}` + "`" + `
**Severity**: critical
**Description**: Test token pattern
`

	testFile := filepath.Join(tmpDir, "patterns.md")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Parse the test file
	patterns, err := parseSecretsFromFile(testFile, tmpDir)
	if err != nil {
		t.Fatalf("parseSecretsFromFile() error: %v", err)
	}

	if len(patterns) != 2 {
		t.Errorf("Expected 2 patterns, got %d", len(patterns))
	}

	// Verify first pattern
	if len(patterns) >= 1 {
		p := patterns[0]
		if p.Severity != "high" {
			t.Errorf("First pattern severity = %q, want %q", p.Severity, "high")
		}
		if p.Pattern == nil {
			t.Error("First pattern has nil Pattern")
		} else {
			// Test that pattern matches expected format (test_ followed by exactly 32 alphanumeric chars)
			// Pattern: test_[A-Za-z0-9]{32}
			testString := "test_abcdefghijklmnopqrstuvwxyz012345" // test_ + 32 chars = valid
			if !p.Pattern.MatchString(testString) {
				t.Errorf("First pattern did not match expected string %q (raw pattern: %s)", testString, p.RawPattern)
			}
		}
	}

	// Verify second pattern
	if len(patterns) >= 2 {
		p := patterns[1]
		if p.Severity != "critical" {
			t.Errorf("Second pattern severity = %q, want %q", p.Severity, "critical")
		}
	}
}

func TestGetRAGPatternCount(t *testing.T) {
	// Clear cache
	ClearRAGSecretsCache()

	// Before loading, count should be 0
	count := GetRAGPatternCount()
	if count != 0 {
		t.Errorf("GetRAGPatternCount() before loading = %d, want 0", count)
	}

	// Load patterns
	LoadRAGSecretPatterns()

	// After loading, count should match cache
	count = GetRAGPatternCount()
	t.Logf("Loaded %d patterns", count)
}

func TestGetRAGPatternSummary(t *testing.T) {
	// Clear cache
	ClearRAGSecretsCache()

	summary := GetRAGPatternSummary()
	if summary == nil {
		t.Error("GetRAGPatternSummary() returned nil")
	}

	t.Logf("Pattern summary by category: %v", summary)
}

func TestClearRAGSecretsCache(t *testing.T) {
	// Load some patterns
	LoadRAGSecretPatterns()

	// Clear cache
	ClearRAGSecretsCache()

	// Verify cache is cleared
	ragSecretsCache.RLock()
	loaded := ragSecretsCache.loaded
	patterns := ragSecretsCache.patterns
	ragSecretsCache.RUnlock()

	if loaded {
		t.Error("Cache should not be marked as loaded after clear")
	}
	if len(patterns) != 0 {
		t.Errorf("Cache should be empty after clear, got %d patterns", len(patterns))
	}
}
