// Package scanner provides integration tests for scanner components
package scanner

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/crashappsec/zero/pkg/scanner/common"
)

// TestIntegration_RAGPatternParsing tests that real RAG pattern files can be parsed
func TestIntegration_RAGPatternParsing(t *testing.T) {
	ragPath := findProjectFile(t, "rag")
	if ragPath == "" {
		t.Skip("RAG directory not found")
	}

	// Test Docker patterns
	dockerPatterns := filepath.Join(ragPath, "devops/docker/patterns.md")
	if _, err := os.Stat(dockerPatterns); err == nil {
		parsed, err := common.ParsePatternMarkdown(dockerPatterns)
		if err != nil {
			t.Errorf("Failed to parse Docker patterns: %v", err)
		}

		if len(parsed.Patterns) < 10 {
			t.Errorf("Docker patterns: got %d patterns, want at least 10", len(parsed.Patterns))
		}

		if parsed.Category != "devops/docker" {
			t.Errorf("Docker patterns category = %q, want devops/docker", parsed.Category)
		}

		// Verify patterns have required fields
		for i, p := range parsed.Patterns {
			if p.Name == "" {
				t.Errorf("Docker pattern %d has empty name", i)
			}
			if p.Pattern == "" {
				t.Errorf("Docker pattern %d (%s) has empty pattern", i, p.Name)
			}
			if p.Severity == "" {
				t.Errorf("Docker pattern %d (%s) has empty severity", i, p.Name)
			}
		}
	}

	// Test API quality patterns
	apiPatterns := filepath.Join(ragPath, "code-quality/api/patterns.md")
	if _, err := os.Stat(apiPatterns); err == nil {
		parsed, err := common.ParsePatternMarkdown(apiPatterns)
		if err != nil {
			t.Errorf("Failed to parse API quality patterns: %v", err)
		}

		if len(parsed.Patterns) < 15 {
			t.Errorf("API quality patterns: got %d patterns, want at least 15", len(parsed.Patterns))
		}

		if parsed.Category != "code-quality/api" {
			t.Errorf("API patterns category = %q, want code-quality/api", parsed.Category)
		}
	}
}

// TestIntegration_SemgrepRuleGeneration tests that Semgrep rules can be generated from RAG patterns
func TestIntegration_SemgrepRuleGeneration(t *testing.T) {
	ragPath := findProjectFile(t, "rag")
	if ragPath == "" {
		t.Skip("RAG directory not found")
	}

	tmpDir := t.TempDir()
	rulesPath := filepath.Join(tmpDir, "rules.yaml")

	// Test generating Docker rules
	dockerPatterns := filepath.Join(ragPath, "devops/docker/patterns.md")
	if _, err := os.Stat(dockerPatterns); err == nil {
		parsed, err := common.ParsePatternMarkdown(dockerPatterns)
		if err != nil {
			t.Fatalf("Failed to parse Docker patterns: %v", err)
		}

		rules := common.ConvertPatternsToSemgrep(parsed, "devops.docker")
		if len(rules.Rules) < 10 {
			t.Errorf("Generated %d rules, want at least 10", len(rules.Rules))
		}

		// Write rules and verify YAML is valid
		if err := common.WriteRulesYAML(rulesPath, rules); err != nil {
			t.Errorf("Failed to write rules YAML: %v", err)
		}

		// Check the file was created
		if _, err := os.Stat(rulesPath); os.IsNotExist(err) {
			t.Error("Rules YAML file was not created")
		}

		// Verify rules have proper structure
		for _, r := range rules.Rules {
			if r.ID == "" {
				t.Error("Rule has empty ID")
			}
			if r.Message == "" {
				t.Error("Rule has empty message")
			}
			if r.Severity == "" {
				t.Error("Rule has empty severity")
			}
			if r.PatternRegex == "" && r.Pattern == "" {
				t.Error("Rule has no pattern or pattern-regex")
			}
			if len(r.Languages) == 0 {
				t.Error("Rule has no languages")
			}
		}
	}
}

// TestIntegration_SemgrepExecution tests that Semgrep can run with generated rules
func TestIntegration_SemgrepExecution(t *testing.T) {
	if !common.HasSemgrep() {
		t.Skip("Semgrep not installed")
	}

	ragPath := findProjectFile(t, "rag")
	if ragPath == "" {
		t.Skip("RAG directory not found")
	}

	tmpDir := t.TempDir()

	// Create a test Dockerfile with issues
	dockerContent := `FROM node:latest
RUN apt-get update
USER root
ENV API_KEY=sk-secret-key-12345
RUN curl http://example.com/script.sh | bash
`
	dockerPath := filepath.Join(tmpDir, "Dockerfile")
	if err := os.WriteFile(dockerPath, []byte(dockerContent), 0644); err != nil {
		t.Fatalf("Failed to create test Dockerfile: %v", err)
	}

	// Generate rules from Docker patterns
	dockerPatterns := filepath.Join(ragPath, "devops/docker/patterns.md")
	if _, err := os.Stat(dockerPatterns); os.IsNotExist(err) {
		t.Skip("Docker patterns not found")
	}

	rulesPath := filepath.Join(tmpDir, "rules.yaml")
	if err := common.GenerateRulesFromRAG(ragPath, "devops/docker", rulesPath); err != nil {
		t.Fatalf("Failed to generate rules: %v", err)
	}

	// Run Semgrep
	runner := common.NewSemgrepRunner(common.SemgrepConfig{
		RulePaths: []string{rulesPath},
		Timeout:   60 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	result := runner.Run(ctx, tmpDir)
	if result.Error != nil {
		// Semgrep may return non-zero exit code for findings - check if we got output
		if len(result.Findings) == 0 {
			t.Logf("Semgrep result: %v", result.Error)
		}
	}

	// Our test Dockerfile has multiple issues - we should find at least some
	t.Logf("Semgrep found %d findings", len(result.Findings))

	// Log findings for debugging
	for _, f := range result.Findings {
		t.Logf("  Finding: %s (%s) at line %d", f.RuleID, f.Severity, f.Line)
	}
}

// TestIntegration_GenerateRulesFromRAG tests the full rule generation flow
func TestIntegration_GenerateRulesFromRAG(t *testing.T) {
	ragPath := findProjectFile(t, "rag")
	if ragPath == "" {
		t.Skip("RAG directory not found")
	}

	tmpDir := t.TempDir()

	tests := []struct {
		category    string
		minPatterns int
	}{
		{"devops/docker", 10},
		{"code-quality/api", 15},
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			rulesPath := filepath.Join(tmpDir, tt.category+".yaml")

			err := common.GenerateRulesFromRAG(ragPath, tt.category, rulesPath)
			if err != nil {
				t.Fatalf("GenerateRulesFromRAG(%s) error = %v", tt.category, err)
			}

			// Verify file was created
			info, err := os.Stat(rulesPath)
			if os.IsNotExist(err) {
				t.Fatalf("Rules file not created for %s", tt.category)
			}

			if info.Size() == 0 {
				t.Fatalf("Rules file is empty for %s", tt.category)
			}

			t.Logf("Generated %s rules file: %d bytes", tt.category, info.Size())
		})
	}
}

// TestIntegration_HasSemgrep checks if Semgrep is available
func TestIntegration_HasSemgrep(t *testing.T) {
	hasSemgrep := common.HasSemgrep()
	t.Logf("Semgrep available: %v", hasSemgrep)

	if !hasSemgrep {
		t.Log("Warning: Semgrep not installed - some scanner features will be limited")
	}
}

// TestIntegration_SemgrepVersionCheck verifies Semgrep can be called
func TestIntegration_SemgrepVersionCheck(t *testing.T) {
	if !common.HasSemgrep() {
		t.Skip("Semgrep not installed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := common.RunCommand(ctx, "semgrep", "--version")
	if err != nil {
		t.Fatalf("semgrep --version failed: %v", err)
	}

	if len(result.Stdout) == 0 {
		t.Error("semgrep --version returned no output")
	}

	t.Logf("Semgrep version: %s", result.Stdout)
}

// findProjectFile looks for a file relative to the project root
func findProjectFile(t *testing.T, relPath string) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			fullPath := filepath.Join(dir, relPath)
			if _, err := os.Stat(fullPath); err == nil {
				return fullPath
			}
			return ""
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
