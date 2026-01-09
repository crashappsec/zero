// Package rag provides validation for RAG patterns
package rag

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationError represents a pattern validation error
type ValidationError struct {
	File     string `json:"file"`
	Pattern  string `json:"pattern"`
	Field    string `json:"field"`
	Message  string `json:"message"`
	Severity string `json:"severity"` // "error" or "warning"
}

// ValidationResult represents the result of validating patterns
type ValidationResult struct {
	TotalPatterns    int               `json:"total_patterns"`
	ValidPatterns    int               `json:"valid_patterns"`
	InvalidPatterns  int               `json:"invalid_patterns"`
	Errors           []ValidationError `json:"errors,omitempty"`
	Warnings         []ValidationError `json:"warnings,omitempty"`
	PatternsByType   map[string]int    `json:"patterns_by_type"`
	PatternsBySeverity map[string]int  `json:"patterns_by_severity"`
}

// Validator validates RAG patterns
type Validator struct {
	StrictMode bool // If true, warnings become errors
}

// NewValidator creates a new pattern validator
func NewValidator() *Validator {
	return &Validator{
		StrictMode: false,
	}
}

// ValidSeverities defines valid severity values
var ValidSeverities = map[string]bool{
	"critical": true,
	"high":     true,
	"medium":   true,
	"low":      true,
	"info":     true,
}

// ValidPatternTypes defines valid pattern types
var ValidPatternTypes = map[string]bool{
	"regex":   true,
	"semgrep": true,
	"glob":    true,
}

// SupportedLanguages defines languages that can be validated
var SupportedLanguages = map[string]bool{
	"python":     true,
	"javascript": true,
	"typescript": true,
	"go":         true,
	"ruby":       true,
	"java":       true,
	"kotlin":     true,
	"scala":      true,
	"csharp":     true,
	"php":        true,
	"rust":       true,
	"c":          true,
	"cpp":        true,
	"swift":      true,
	"bash":       true,
	"shell":      true,
	"powershell": true,
	"generic":    true,
}

// ValidatePattern validates a single pattern
func (v *Validator) ValidatePattern(p Pattern, source string) []ValidationError {
	var errors []ValidationError

	// Validate pattern field is not empty
	if strings.TrimSpace(p.Pattern) == "" {
		errors = append(errors, ValidationError{
			File:     source,
			Pattern:  p.ID,
			Field:    "pattern",
			Message:  "pattern field is empty",
			Severity: "error",
		})
	}

	// Validate pattern type
	if p.Type != "" && !ValidPatternTypes[p.Type] {
		errors = append(errors, ValidationError{
			File:     source,
			Pattern:  p.ID,
			Field:    "type",
			Message:  fmt.Sprintf("invalid pattern type: %s (valid: regex, semgrep, glob)", p.Type),
			Severity: "error",
		})
	}

	// Validate severity
	if p.Severity != "" && !ValidSeverities[strings.ToLower(p.Severity)] {
		errors = append(errors, ValidationError{
			File:     source,
			Pattern:  p.ID,
			Field:    "severity",
			Message:  fmt.Sprintf("invalid severity: %s (valid: critical, high, medium, low, info)", p.Severity),
			Severity: "warning",
		})
	}

	// Validate regex patterns compile
	if p.Type == "regex" || p.Type == "" {
		if err := v.ValidateRegex(p.Pattern); err != nil {
			errors = append(errors, ValidationError{
				File:     source,
				Pattern:  p.ID,
				Field:    "pattern",
				Message:  fmt.Sprintf("invalid regex: %v", err),
				Severity: "error",
			})
		}
	}

	// Validate languages if specified
	for _, lang := range p.Languages {
		if !SupportedLanguages[strings.ToLower(lang)] {
			errors = append(errors, ValidationError{
				File:     source,
				Pattern:  p.ID,
				Field:    "languages",
				Message:  fmt.Sprintf("unsupported language: %s", lang),
				Severity: "warning",
			})
		}
	}

	// Validate confidence is in range
	if p.Confidence < 0 || p.Confidence > 100 {
		errors = append(errors, ValidationError{
			File:     source,
			Pattern:  p.ID,
			Field:    "confidence",
			Message:  fmt.Sprintf("confidence must be 0-100, got: %d", p.Confidence),
			Severity: "warning",
		})
	}

	// Warn about missing message
	if strings.TrimSpace(p.Message) == "" {
		errors = append(errors, ValidationError{
			File:     source,
			Pattern:  p.ID,
			Field:    "message",
			Message:  "pattern has no message/description",
			Severity: "warning",
		})
	}

	return errors
}

// ValidateRegex checks if a regex pattern compiles
func (v *Validator) ValidateRegex(pattern string) error {
	if pattern == "" {
		return fmt.Errorf("empty pattern")
	}

	_, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("regex compilation failed: %w", err)
	}

	return nil
}

// ValidateSeverity checks if a severity value is valid
func (v *Validator) ValidateSeverity(severity string) bool {
	return ValidSeverities[strings.ToLower(severity)]
}

// ValidatePatternSet validates a complete pattern set
func (v *Validator) ValidatePatternSet(ps PatternSet) *ValidationResult {
	result := &ValidationResult{
		PatternsByType:     make(map[string]int),
		PatternsBySeverity: make(map[string]int),
	}

	for _, p := range ps.Patterns {
		result.TotalPatterns++

		// Track type and severity stats
		ptype := p.Type
		if ptype == "" {
			ptype = "regex"
		}
		result.PatternsByType[ptype]++

		sev := strings.ToLower(p.Severity)
		if sev == "" {
			sev = "info"
		}
		result.PatternsBySeverity[sev]++

		// Validate pattern
		errors := v.ValidatePattern(p, ps.Source)

		hasError := false
		for _, err := range errors {
			if err.Severity == "error" {
				result.Errors = append(result.Errors, err)
				hasError = true
			} else {
				if v.StrictMode {
					result.Errors = append(result.Errors, err)
					hasError = true
				} else {
					result.Warnings = append(result.Warnings, err)
				}
			}
		}

		if hasError {
			result.InvalidPatterns++
		} else {
			result.ValidPatterns++
		}
	}

	return result
}

// ValidateCategory validates all patterns in a RAG category
func (v *Validator) ValidateCategory(loader *RAGLoader, category string) (*ValidationResult, error) {
	loadResult, err := loader.LoadCategory(category)
	if err != nil {
		return nil, fmt.Errorf("loading category %s: %w", category, err)
	}

	combined := &ValidationResult{
		PatternsByType:     make(map[string]int),
		PatternsBySeverity: make(map[string]int),
	}

	for _, ps := range loadResult.PatternSets {
		result := v.ValidatePatternSet(ps)

		// Merge results
		combined.TotalPatterns += result.TotalPatterns
		combined.ValidPatterns += result.ValidPatterns
		combined.InvalidPatterns += result.InvalidPatterns
		combined.Errors = append(combined.Errors, result.Errors...)
		combined.Warnings = append(combined.Warnings, result.Warnings...)

		for k, v := range result.PatternsByType {
			combined.PatternsByType[k] += v
		}
		for k, v := range result.PatternsBySeverity {
			combined.PatternsBySeverity[k] += v
		}
	}

	return combined, nil
}

// ValidateAll validates all RAG categories
func (v *Validator) ValidateAll(loader *RAGLoader) (*ValidationResult, error) {
	categories, err := loader.ListCategories()
	if err != nil {
		return nil, fmt.Errorf("listing categories: %w", err)
	}

	combined := &ValidationResult{
		PatternsByType:     make(map[string]int),
		PatternsBySeverity: make(map[string]int),
	}

	for _, category := range categories {
		result, err := v.ValidateCategory(loader, category)
		if err != nil {
			// Log error but continue with other categories
			combined.Errors = append(combined.Errors, ValidationError{
				File:     category,
				Pattern:  "",
				Field:    "category",
				Message:  fmt.Sprintf("failed to validate category: %v", err),
				Severity: "error",
			})
			continue
		}

		// Merge results
		combined.TotalPatterns += result.TotalPatterns
		combined.ValidPatterns += result.ValidPatterns
		combined.InvalidPatterns += result.InvalidPatterns
		combined.Errors = append(combined.Errors, result.Errors...)
		combined.Warnings = append(combined.Warnings, result.Warnings...)

		for k, v := range result.PatternsByType {
			combined.PatternsByType[k] += v
		}
		for k, v := range result.PatternsBySeverity {
			combined.PatternsBySeverity[k] += v
		}
	}

	return combined, nil
}

// Summary returns a human-readable summary of validation results
func (r *ValidationResult) Summary() string {
	var sb strings.Builder

	sb.WriteString("Validation Summary:\n")
	sb.WriteString(fmt.Sprintf("  Total Patterns:   %d\n", r.TotalPatterns))
	sb.WriteString(fmt.Sprintf("  Valid:            %d\n", r.ValidPatterns))
	sb.WriteString(fmt.Sprintf("  Invalid:          %d\n", r.InvalidPatterns))
	sb.WriteString(fmt.Sprintf("  Errors:           %d\n", len(r.Errors)))
	sb.WriteString(fmt.Sprintf("  Warnings:         %d\n", len(r.Warnings)))

	if len(r.PatternsByType) > 0 {
		sb.WriteString("\n  By Type:\n")
		for k, v := range r.PatternsByType {
			sb.WriteString(fmt.Sprintf("    %s: %d\n", k, v))
		}
	}

	if len(r.PatternsBySeverity) > 0 {
		sb.WriteString("\n  By Severity:\n")
		for k, v := range r.PatternsBySeverity {
			sb.WriteString(fmt.Sprintf("    %s: %d\n", k, v))
		}
	}

	return sb.String()
}

// IsValid returns true if there are no errors
func (r *ValidationResult) IsValid() bool {
	return len(r.Errors) == 0
}

// ValidationRate returns the percentage of valid patterns
func (r *ValidationResult) ValidationRate() float64 {
	if r.TotalPatterns == 0 {
		return 100.0
	}
	return float64(r.ValidPatterns) / float64(r.TotalPatterns) * 100.0
}
