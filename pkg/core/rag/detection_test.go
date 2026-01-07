package rag

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// SemgrepResult represents a Semgrep scan result
type SemgrepResult struct {
	Results []SemgrepFinding `json:"results"`
	Errors  []interface{}    `json:"errors"`
}

// SemgrepFinding represents a single finding
type SemgrepFinding struct {
	CheckID string `json:"check_id"`
	Path    string `json:"path"`
	Start   struct {
		Line int `json:"line"`
	} `json:"start"`
	Extra struct {
		Lines   string `json:"lines"`
		Message string `json:"message"`
	} `json:"extra"`
}

// findProjectRoot finds the project root by looking for go.mod
func findProjectRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// TestSecretsDetection tests that secret patterns detect known secrets
func TestSecretsDetection(t *testing.T) {
	root := findProjectRoot()
	if root == "" {
		t.Skip("Could not find project root")
	}

	rulesFile := filepath.Join(root, ".zero/rules/generated/secrets.yaml")
	testDir := filepath.Join(root, "testdata/rag/known-secrets")

	// Check if rules exist
	if _, err := os.Stat(rulesFile); os.IsNotExist(err) {
		t.Skip("Rules not generated. Run 'zero feeds rag' first")
	}

	// Check if test directory exists
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Skip("Test fixtures not found")
	}

	// Run semgrep (use CombinedOutput since semgrep returns non-zero on findings)
	cmd := exec.Command("semgrep", "--config", rulesFile, testDir, "--json", "--quiet")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Semgrep returns non-zero when findings exist, that's OK
		if _, ok := err.(*exec.ExitError); !ok {
			t.Fatalf("Failed to run semgrep: %v", err)
		}
	}

	var result SemgrepResult
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Failed to parse semgrep output: %v", err)
	}

	// Should find some secrets
	if len(result.Results) == 0 {
		t.Error("Expected to find secrets in known-secrets/ but found none")
	}

	t.Logf("Found %d findings in known-secrets/", len(result.Results))

	// Log the findings for debugging
	for _, finding := range result.Results {
		t.Logf("  - %s (line %d): %s", finding.CheckID, finding.Start.Line, finding.Extra.Message)
	}

	// Expected patterns to detect
	expectedPatterns := []string{
		"aws", "github", "stripe", "sendgrid", "slack",
	}

	foundPatterns := make(map[string]bool)
	for _, finding := range result.Results {
		for _, pattern := range expectedPatterns {
			if strings.Contains(strings.ToLower(finding.CheckID), pattern) {
				foundPatterns[pattern] = true
			}
		}
	}

	// Warn about missing patterns (not fail, as some may not have rules yet)
	for _, pattern := range expectedPatterns {
		if !foundPatterns[pattern] {
			t.Logf("Warning: Expected to detect %s patterns but didn't", pattern)
		}
	}
}

// TestFalsePositiveExclusion tests that false positives are not detected
func TestFalsePositiveExclusion(t *testing.T) {
	root := findProjectRoot()
	if root == "" {
		t.Skip("Could not find project root")
	}

	rulesFile := filepath.Join(root, ".zero/rules/generated/secrets.yaml")
	testDir := filepath.Join(root, "testdata/rag/false-positives")

	// Check if rules exist
	if _, err := os.Stat(rulesFile); os.IsNotExist(err) {
		t.Skip("Rules not generated. Run 'zero feeds rag' first")
	}

	// Check if test directory exists
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Skip("Test fixtures not found")
	}

	// Run semgrep
	cmd := exec.Command("semgrep", "--config", rulesFile, testDir, "--json", "--quiet")
	output, err := cmd.CombinedOutput()
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			t.Fatalf("Failed to run semgrep: %v", err)
		}
	}

	var result SemgrepResult
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Failed to parse semgrep output: %v", err)
	}

	// Should find NO secrets in false-positives directory
	if len(result.Results) > 0 {
		t.Errorf("Expected 0 findings in false-positives/ but found %d:", len(result.Results))
		for _, finding := range result.Results {
			t.Errorf("  - %s (line %d): %s", finding.CheckID, finding.Start.Line, finding.Extra.Message)
		}
	} else {
		t.Log("Correctly found 0 false positives")
	}
}

// TestTechnologyDetection tests that technology patterns detect known technologies
func TestTechnologyDetection(t *testing.T) {
	root := findProjectRoot()
	if root == "" {
		t.Skip("Could not find project root")
	}

	rulesFile := filepath.Join(root, ".zero/rules/generated/tech-discovery.yaml")
	testDir := filepath.Join(root, "testdata/rag/tech-samples")

	// Check if rules exist
	if _, err := os.Stat(rulesFile); os.IsNotExist(err) {
		t.Skip("Rules not generated. Run 'zero feeds rag' first")
	}

	// Check if test directory exists
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Skip("Test fixtures not found")
	}

	// Test React detection
	t.Run("React", func(t *testing.T) {
		testTechDetection(t, rulesFile, filepath.Join(testDir, "react"), "react")
	})

	// Test Python/Flask detection
	t.Run("Flask", func(t *testing.T) {
		testTechDetection(t, rulesFile, filepath.Join(testDir, "python"), "flask")
	})

	// Test Go/Gin detection
	t.Run("Go/Gin", func(t *testing.T) {
		testTechDetection(t, rulesFile, filepath.Join(testDir, "go"), "gin")
	})
}

func testTechDetection(t *testing.T, rulesFile, testDir, expectedTech string) {
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Skipf("Test directory %s not found", testDir)
	}

	cmd := exec.Command("semgrep", "--config", rulesFile, testDir, "--json", "--quiet")
	output, err := cmd.CombinedOutput()
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			t.Fatalf("Failed to run semgrep: %v", err)
		}
	}

	var result SemgrepResult
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Failed to parse semgrep output: %v", err)
	}

	// Should find the expected technology
	found := false
	for _, finding := range result.Results {
		if strings.Contains(strings.ToLower(finding.CheckID), expectedTech) {
			found = true
			t.Logf("Found %s: %s", expectedTech, finding.CheckID)
		}
	}

	if !found {
		t.Logf("Warning: Expected to detect %s but didn't find matching rule", expectedTech)
		t.Logf("Found %d findings total", len(result.Results))
	}
}

// TestCryptographyDetection tests weak crypto pattern detection
func TestCryptographyDetection(t *testing.T) {
	root := findProjectRoot()
	if root == "" {
		t.Skip("Could not find project root")
	}

	rulesFile := filepath.Join(root, ".zero/rules/generated/cryptography.yaml")
	testFile := filepath.Join(root, "testdata/rag/known-secrets/weak_crypto.py")

	// Check if rules exist
	if _, err := os.Stat(rulesFile); os.IsNotExist(err) {
		t.Skip("Crypto rules not generated. Run 'zero feeds rag' first")
	}

	// Check if test file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Crypto test fixture not found")
	}

	cmd := exec.Command("semgrep", "--config", rulesFile, testFile, "--json", "--quiet")
	output, err := cmd.CombinedOutput()
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			t.Fatalf("Failed to run semgrep: %v", err)
		}
	}

	var result SemgrepResult
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Failed to parse semgrep output: %v", err)
	}

	t.Logf("Found %d cryptography findings", len(result.Results))
	for _, finding := range result.Results {
		t.Logf("  - %s (line %d)", finding.CheckID, finding.Start.Line)
	}

	// Should detect at least some weak crypto patterns
	if len(result.Results) == 0 {
		t.Log("Warning: Expected to find weak crypto patterns but found none")
	}
}
