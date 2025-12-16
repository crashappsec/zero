// Package techid provides the consolidated technology identification super scanner
// This file provides semgrep execution wrapper for technology detection
package techid

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/crashappsec/zero/pkg/scanners/common"
)

// SemgrepRunner executes semgrep with generated rules
type SemgrepRunner struct {
	rulePaths []string
	timeout   time.Duration
	onStatus  func(string)
}

// SemgrepConfig configures semgrep execution
type SemgrepConfig struct {
	RulePaths []string      // Paths to rule YAML files
	Timeout   time.Duration // Execution timeout
	OnStatus  func(string)  // Status callback
}

// NewSemgrepRunner creates a new semgrep runner
func NewSemgrepRunner(cfg SemgrepConfig) *SemgrepRunner {
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Minute
	}
	if cfg.OnStatus == nil {
		cfg.OnStatus = func(string) {}
	}
	return &SemgrepRunner{
		rulePaths: cfg.RulePaths,
		timeout:   cfg.Timeout,
		onStatus:  cfg.OnStatus,
	}
}

// SemgrepFinding represents a finding from semgrep
type SemgrepFinding struct {
	RuleID      string
	Technology  string
	Category    string
	File        string
	Line        int
	Column      int
	Message     string
	Severity    string
	Confidence  int
	Match       string
	SecretType  string // For secrets
}

// SemgrepResult contains results from running semgrep
type SemgrepResult struct {
	Findings       []SemgrepFinding
	Technologies   map[string]int // Tech name -> count
	Secrets        []SemgrepFinding
	Error          error
	Duration       time.Duration
	SemgrepVersion string
}

// Run executes semgrep against the target path
func (sr *SemgrepRunner) Run(ctx context.Context, targetPath string) *SemgrepResult {
	result := &SemgrepResult{
		Technologies: make(map[string]int),
	}

	if !common.ToolExists("semgrep") {
		result.Error = fmt.Errorf("semgrep not installed")
		return result
	}

	if len(sr.rulePaths) == 0 {
		result.Error = fmt.Errorf("no rule files configured")
		return result
	}

	start := time.Now()

	// Build semgrep args
	args := []string{
		"--json",
		"--metrics=off",
		"--timeout", "60",
		"--max-memory", "4096",
		"--exclude", "node_modules",
		"--exclude", "vendor",
		"--exclude", ".git",
		"--exclude", "dist",
		"--exclude", "build",
		"--exclude", "__pycache__",
		"--exclude", "*.min.js",
	}

	// Add all rule files
	for _, rulePath := range sr.rulePaths {
		args = append(args, "--config", rulePath)
	}

	args = append(args, targetPath)

	sr.onStatus("Scanning for technologies with semgrep...")

	ctx, cancel := context.WithTimeout(ctx, sr.timeout)
	defer cancel()

	cmdResult, err := common.RunCommand(ctx, "semgrep", args...)
	if err != nil {
		result.Error = fmt.Errorf("semgrep execution failed: %w", err)
		return result
	}

	result.Duration = time.Since(start)

	// Parse the output
	findings, secrets, techCounts := parseSemgrepOutput(cmdResult.Stdout, targetPath)
	result.Findings = findings
	result.Secrets = secrets
	result.Technologies = techCounts

	return result
}

// semgrepOutput represents the JSON output from semgrep
type semgrepOutput struct {
	Version string `json:"version"`
	Results []struct {
		CheckID string `json:"check_id"`
		Path    string `json:"path"`
		Start   struct {
			Line   int `json:"line"`
			Col    int `json:"col"`
			Offset int `json:"offset"`
		} `json:"start"`
		End struct {
			Line   int `json:"line"`
			Col    int `json:"col"`
			Offset int `json:"offset"`
		} `json:"end"`
		Extra struct {
			Lines    string                 `json:"lines"`
			Message  string                 `json:"message"`
			Severity string                 `json:"severity"`
			Metadata map[string]interface{} `json:"metadata"`
		} `json:"extra"`
	} `json:"results"`
	Errors []struct {
		Level   string `json:"level"`
		Message string `json:"message"`
	} `json:"errors"`
}

// parseSemgrepOutput parses semgrep JSON output into findings
func parseSemgrepOutput(data []byte, basePath string) ([]SemgrepFinding, []SemgrepFinding, map[string]int) {
	var findings []SemgrepFinding
	var secrets []SemgrepFinding
	techCounts := make(map[string]int)

	var output semgrepOutput
	if err := json.Unmarshal(data, &output); err != nil {
		return findings, secrets, techCounts
	}

	for _, r := range output.Results {
		// Make path relative
		file := r.Path
		if strings.HasPrefix(file, basePath) {
			file = strings.TrimPrefix(file, basePath+"/")
		}

		// Extract metadata
		technology := getStringFromMetadata(r.Extra.Metadata, "technology")
		category := getStringFromMetadata(r.Extra.Metadata, "category")
		confidence := getIntFromMetadata(r.Extra.Metadata, "confidence")
		secretType := getStringFromMetadata(r.Extra.Metadata, "secret_type")

		finding := SemgrepFinding{
			RuleID:     r.CheckID,
			Technology: technology,
			Category:   category,
			File:       file,
			Line:       r.Start.Line,
			Column:     r.Start.Col,
			Message:    r.Extra.Message,
			Severity:   r.Extra.Severity,
			Confidence: confidence,
			Match:      strings.TrimSpace(r.Extra.Lines),
			SecretType: secretType,
		}

		// Categorize as secret or technology finding
		if strings.Contains(r.CheckID, ".secret.") || category == "secrets" {
			secrets = append(secrets, finding)
		} else {
			findings = append(findings, finding)
			if technology != "" {
				techCounts[technology]++
			}
		}
	}

	return findings, secrets, techCounts
}

// getStringFromMetadata extracts a string value from metadata
func getStringFromMetadata(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

// getIntFromMetadata extracts an int value from metadata
func getIntFromMetadata(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case int64:
			return int(v)
		}
	}
	return 0
}

// HasSemgrep checks if semgrep is installed
func HasSemgrep() bool {
	return common.ToolExists("semgrep")
}

// GetSemgrepVersion returns the installed semgrep version
func GetSemgrepVersion() string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := common.RunCommand(ctx, "semgrep", "--version")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(result.Stdout))
}

// RunSemgrepWithRules is a convenience function to run semgrep with rules from RuleManager
func RunSemgrepWithRules(ctx context.Context, targetPath string, ruleManager *RuleManager, onStatus func(string)) *SemgrepResult {
	rulePaths := ruleManager.GetRulePaths()

	// Add absolute paths
	var absPaths []string
	for _, p := range rulePaths {
		absPath, err := filepath.Abs(p)
		if err == nil {
			absPaths = append(absPaths, absPath)
		} else {
			absPaths = append(absPaths, p)
		}
	}

	runner := NewSemgrepRunner(SemgrepConfig{
		RulePaths: absPaths,
		OnStatus:  onStatus,
	})

	return runner.Run(ctx, targetPath)
}
