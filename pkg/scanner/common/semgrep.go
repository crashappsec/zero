// Package common provides shared utilities for scanners
// This file provides semgrep execution and pattern conversion utilities
package common

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// SemgrepRunner executes semgrep with rule files
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
	Category    string
	File        string
	Line        int
	Column      int
	Message     string
	Severity    string
	Match       string
	Remediation string
	Metadata    map[string]interface{}
}

// SemgrepResult contains results from running semgrep
type SemgrepResult struct {
	Findings []SemgrepFinding
	Error    error
	Duration time.Duration
}

// Run executes semgrep against the target path
func (sr *SemgrepRunner) Run(ctx context.Context, targetPath string) *SemgrepResult {
	result := &SemgrepResult{}

	if !ToolExists("semgrep") {
		result.Error = fmt.Errorf("semgrep not installed")
		return result
	}

	if len(sr.rulePaths) == 0 {
		result.Error = fmt.Errorf("no rule files configured")
		return result
	}

	// Validate all rule files exist
	for _, rulePath := range sr.rulePaths {
		if _, err := os.Stat(rulePath); os.IsNotExist(err) {
			// Skip missing files
			continue
		}
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
	hasRules := false
	for _, rulePath := range sr.rulePaths {
		if _, err := os.Stat(rulePath); err == nil {
			args = append(args, "--config", rulePath)
			hasRules = true
		}
	}

	if !hasRules {
		result.Error = fmt.Errorf("no valid rule files found")
		return result
	}

	args = append(args, targetPath)

	sr.onStatus("Running semgrep analysis...")

	ctx, cancel := context.WithTimeout(ctx, sr.timeout)
	defer cancel()

	cmdResult, err := RunCommand(ctx, "semgrep", args...)
	if err != nil {
		// Semgrep returns non-zero exit codes for findings too
		// Only treat as error if we got no output
		if cmdResult == nil || len(cmdResult.Stdout) == 0 {
			result.Error = fmt.Errorf("semgrep execution failed: %w", err)
			return result
		}
	}

	result.Duration = time.Since(start)

	// Parse the output
	findings := parseSemgrepOutput(cmdResult.Stdout, targetPath)
	result.Findings = findings

	return result
}

// RunOnFiles executes semgrep only on specific files
func (sr *SemgrepRunner) RunOnFiles(ctx context.Context, files []string, basePath string) *SemgrepResult {
	result := &SemgrepResult{}

	if !ToolExists("semgrep") {
		result.Error = fmt.Errorf("semgrep not installed")
		return result
	}

	if len(sr.rulePaths) == 0 {
		result.Error = fmt.Errorf("no rule files configured")
		return result
	}

	if len(files) == 0 {
		return result
	}

	start := time.Now()

	// Build semgrep args
	args := []string{
		"--json",
		"--metrics=off",
		"--timeout", "60",
		"--max-memory", "4096",
	}

	// Add all rule files
	hasRules := false
	for _, rulePath := range sr.rulePaths {
		if _, err := os.Stat(rulePath); err == nil {
			args = append(args, "--config", rulePath)
			hasRules = true
		}
	}

	if !hasRules {
		result.Error = fmt.Errorf("no valid rule files found")
		return result
	}

	// Add specific files to scan
	args = append(args, files...)

	sr.onStatus(fmt.Sprintf("Running semgrep on %d files...", len(files)))

	ctx, cancel := context.WithTimeout(ctx, sr.timeout)
	defer cancel()

	cmdResult, err := RunCommand(ctx, "semgrep", args...)
	if err != nil {
		if cmdResult == nil || len(cmdResult.Stdout) == 0 {
			result.Error = fmt.Errorf("semgrep execution failed: %w", err)
			return result
		}
	}

	result.Duration = time.Since(start)
	result.Findings = parseSemgrepOutput(cmdResult.Stdout, basePath)

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
func parseSemgrepOutput(data []byte, basePath string) []SemgrepFinding {
	var findings []SemgrepFinding

	var output semgrepOutput
	if err := json.Unmarshal(data, &output); err != nil {
		return findings
	}

	for _, r := range output.Results {
		// Make path relative
		file := r.Path
		if strings.HasPrefix(file, basePath) {
			file = strings.TrimPrefix(file, basePath+"/")
		}

		// Extract metadata
		category := getStringFromMetadata(r.Extra.Metadata, "category")
		remediation := getStringFromMetadata(r.Extra.Metadata, "remediation")

		// Map semgrep severity to our severity
		severity := mapSemgrepSeverity(r.Extra.Severity)

		finding := SemgrepFinding{
			RuleID:      r.CheckID,
			Category:    category,
			File:        file,
			Line:        r.Start.Line,
			Column:      r.Start.Col,
			Message:     r.Extra.Message,
			Severity:    severity,
			Match:       strings.TrimSpace(r.Extra.Lines),
			Remediation: remediation,
			Metadata:    r.Extra.Metadata,
		}

		findings = append(findings, finding)
	}

	return findings
}

func getStringFromMetadata(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

func mapSemgrepSeverity(sev string) string {
	switch strings.ToUpper(sev) {
	case "ERROR":
		return "high"
	case "WARNING":
		return "medium"
	case "INFO":
		return "low"
	default:
		return "info"
	}
}

// HasSemgrep checks if semgrep is installed
func HasSemgrep() bool {
	return ToolExists("semgrep")
}

// =========================================================================
// RAG Pattern Converter - Converts markdown patterns to Semgrep rules
// =========================================================================

// PatternRule represents a parsed pattern from RAG markdown
type PatternRule struct {
	Name        string
	Type        string // regex, semgrep
	Severity    string
	Pattern     string
	Description string
	Example     string
	Remediation string
	Language    string
	Category    string
	CWE         string
}

// ParsedPatternFile contains all patterns parsed from a file
type ParsedPatternFile struct {
	Category    string
	Description string
	CWE         string
	Patterns    []PatternRule
}

// ParsePatternMarkdown parses a RAG pattern markdown file
// Format:
//   ## Section Name (group heading, not a pattern)
//   ### Pattern Name
//   **Type**: regex
//   **Severity**: high
//   **Pattern**: `regex-here`
//   - Description text
//   - Example: `example code`
//   - Remediation: How to fix
func ParsePatternMarkdown(path string) (*ParsedPatternFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	result := &ParsedPatternFile{
		Patterns: []PatternRule{},
	}

	scanner := bufio.NewScanner(file)

	// Regex patterns for parsing
	categoryRe := regexp.MustCompile(`\*\*Category\*\*:\s*(.+)`)
	descRe := regexp.MustCompile(`\*\*Description\*\*:\s*(.+)`)
	cweRe := regexp.MustCompile(`\*\*CWE\*\*:\s*(.+)`)
	typeRe := regexp.MustCompile(`\*\*Type\*\*:\s*(\w+)`)
	severityRe := regexp.MustCompile(`\*\*Severity\*\*:\s*(\w+)`)
	patternRe := regexp.MustCompile(`\*\*Pattern\*\*:\s*` + "`" + `([^` + "`" + `]+)` + "`")

	var currentPattern *PatternRule
	var collectingDescription bool
	var headerParsed bool // Track if we're past the file header

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Parse file-level metadata (before first ##)
		if !headerParsed {
			if m := categoryRe.FindStringSubmatch(trimmed); m != nil {
				result.Category = m[1]
				continue
			}
			if m := descRe.FindStringSubmatch(trimmed); m != nil {
				result.Description = m[1]
				continue
			}
			if m := cweRe.FindStringSubmatch(trimmed); m != nil {
				result.CWE = m[1]
				continue
			}
		}

		// ## is a section heading (group of patterns), not a pattern itself
		if strings.HasPrefix(trimmed, "## ") {
			headerParsed = true
			// Save any in-progress pattern
			if currentPattern != nil && currentPattern.Pattern != "" {
				result.Patterns = append(result.Patterns, *currentPattern)
				currentPattern = nil
			}
			collectingDescription = false
			continue
		}

		// ### is the actual pattern name
		if strings.HasPrefix(trimmed, "### ") {
			// Save previous pattern if it has a valid pattern regex
			if currentPattern != nil && currentPattern.Pattern != "" {
				result.Patterns = append(result.Patterns, *currentPattern)
			}

			// Start new pattern
			name := strings.TrimPrefix(trimmed, "### ")
			currentPattern = &PatternRule{
				Name:     name,
				Category: result.Category,
				CWE:      result.CWE,
			}
			collectingDescription = false
			continue
		}

		// Skip --- dividers
		if trimmed == "---" {
			continue
		}

		// Parse pattern metadata
		if currentPattern != nil {
			if m := typeRe.FindStringSubmatch(trimmed); m != nil {
				currentPattern.Type = strings.ToLower(m[1])
				continue
			}
			if m := severityRe.FindStringSubmatch(trimmed); m != nil {
				currentPattern.Severity = strings.ToLower(m[1])
				continue
			}
			if m := patternRe.FindStringSubmatch(trimmed); m != nil {
				currentPattern.Pattern = m[1]
				collectingDescription = true
				continue
			}

			// Collect bullet points after pattern
			if collectingDescription && strings.HasPrefix(trimmed, "- ") {
				content := strings.TrimPrefix(trimmed, "- ")

				if strings.HasPrefix(content, "Example:") {
					currentPattern.Example = strings.TrimPrefix(content, "Example:")
					currentPattern.Example = strings.Trim(currentPattern.Example, " `")
				} else if strings.HasPrefix(content, "Remediation:") {
					currentPattern.Remediation = strings.TrimPrefix(content, "Remediation:")
					currentPattern.Remediation = strings.TrimSpace(currentPattern.Remediation)
				} else if currentPattern.Description == "" {
					currentPattern.Description = content
				}
			}
		}
	}

	// Don't forget the last pattern
	if currentPattern != nil && currentPattern.Pattern != "" {
		result.Patterns = append(result.Patterns, *currentPattern)
	}

	return result, nil
}

// SemgrepRule represents a semgrep rule for YAML output
type SemgrepRule struct {
	ID           string                 `yaml:"id"`
	Message      string                 `yaml:"message"`
	Severity     string                 `yaml:"severity"`
	Languages    []string               `yaml:"languages"`
	Metadata     map[string]interface{} `yaml:"metadata,omitempty"`
	Pattern      string                 `yaml:"pattern,omitempty"`
	PatternRegex string                 `yaml:"pattern-regex,omitempty"`
}

// SemgrepRuleFile represents a semgrep rules file
type SemgrepRuleFile struct {
	Rules []SemgrepRule `yaml:"rules"`
}

// ConvertPatternsToSemgrep converts parsed patterns to semgrep rules
func ConvertPatternsToSemgrep(parsed *ParsedPatternFile, rulePrefix string) *SemgrepRuleFile {
	rules := &SemgrepRuleFile{
		Rules: []SemgrepRule{},
	}

	for _, p := range parsed.Patterns {
		if p.Pattern == "" {
			continue
		}

		// Skip structural patterns - these are handled by file-level analysis, not Semgrep
		if p.Type == "structural" {
			continue
		}

		// Generate rule ID
		ruleID := fmt.Sprintf("zero.%s.%s", rulePrefix, sanitizeRuleID(p.Name))

		// Determine language
		languages := []string{"generic"}
		if p.Language != "" {
			lang := mapToSemgrepLanguage(p.Language)
			if lang != "" {
				languages = []string{lang}
			}
		}

		// Use pattern-level values with file-level fallbacks
		category := p.Category
		if category == "" {
			category = parsed.Category
		}
		cwe := p.CWE
		if cwe == "" {
			cwe = parsed.CWE
		}

		rule := SemgrepRule{
			ID:        ruleID,
			Message:   p.Description,
			Severity:  mapSeverityToSemgrep(p.Severity),
			Languages: languages,
			Metadata: map[string]interface{}{
				"category":    category,
				"remediation": p.Remediation,
			},
		}

		if cwe != "" {
			rule.Metadata["cwe"] = cwe
		}

		// Use pattern-regex for regex type patterns
		if p.Type == "regex" || p.Type == "" {
			rule.PatternRegex = p.Pattern
		} else {
			rule.Pattern = p.Pattern
		}

		rules.Rules = append(rules.Rules, rule)
	}

	return rules
}

func sanitizeRuleID(name string) string {
	id := strings.ToLower(name)
	id = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(id, "-")
	id = strings.Trim(id, "-")
	return id
}

func mapToSemgrepLanguage(lang string) string {
	mapping := map[string]string{
		"python":     "python",
		"javascript": "javascript",
		"typescript": "typescript",
		"go":         "go",
		"ruby":       "ruby",
		"java":       "java",
		"dockerfile": "dockerfile",
		"yaml":       "yaml",
		"json":       "json",
		"generic":    "generic",
	}
	return mapping[strings.ToLower(lang)]
}

func mapSeverityToSemgrep(sev string) string {
	switch strings.ToLower(sev) {
	case "critical", "high":
		return "ERROR"
	case "medium":
		return "WARNING"
	case "low", "info":
		return "INFO"
	default:
		return "INFO"
	}
}

// WriteRulesYAML writes semgrep rules to a YAML file
func WriteRulesYAML(path string, rules *SemgrepRuleFile) error {
	data, err := yaml.Marshal(rules)
	if err != nil {
		return fmt.Errorf("failed to marshal rules: %w", err)
	}
	return os.WriteFile(path, data, 0600)
}

// GenerateRulesFromRAG converts all pattern files in a RAG category to semgrep rules
func GenerateRulesFromRAG(ragPath, category, outputPath string) error {
	categoryDir := filepath.Join(ragPath, category)

	if _, err := os.Stat(categoryDir); os.IsNotExist(err) {
		return fmt.Errorf("category not found: %s", category)
	}

	allRules := &SemgrepRuleFile{
		Rules: []SemgrepRule{},
	}

	// Walk the category directory
	err := filepath.Walk(categoryDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Accept any .md file (patterns.md, docker.md, api.md, etc.)
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}

		// Parse the pattern file
		parsed, err := ParsePatternMarkdown(path)
		if err != nil {
			return nil // Skip files that fail to parse
		}

		// Generate rule prefix from path
		relPath, _ := filepath.Rel(ragPath, filepath.Dir(path))
		rulePrefix := strings.ReplaceAll(relPath, "/", ".")
		rulePrefix = strings.ReplaceAll(rulePrefix, "\\", ".")

		// Convert to semgrep rules
		rules := ConvertPatternsToSemgrep(parsed, rulePrefix)
		allRules.Rules = append(allRules.Rules, rules.Rules...)

		return nil
	})

	if err != nil {
		return fmt.Errorf("walking category dir: %w", err)
	}

	if len(allRules.Rules) == 0 {
		return nil // No rules to write
	}

	// Write output
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}

	return WriteRulesYAML(outputPath, allRules)
}
