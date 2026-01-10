// Package findings provides validation and confidence scoring for finding management
package findings

import (
	"fmt"
	"strings"
)

// Signal represents a confidence signal from detection
type Signal struct {
	Type        string  `json:"type"`        // e.g., "package", "import", "config", "api_call"
	Weight      float64 `json:"weight"`      // 0.0-1.0 contribution to confidence
	Source      string  `json:"source"`      // e.g., "package.json", "src/app.ts:5"
	Description string  `json:"description"` // Human-readable explanation
}

// ConfidenceScore represents a computed confidence with signals
type ConfidenceScore struct {
	Score   int      `json:"score"`   // 0-100 overall confidence
	Level   string   `json:"level"`   // "high", "medium", "low"
	Signals []Signal `json:"signals"` // Contributing signals
	Reason  string   `json:"reason"`  // Human-readable summary
}

// ComputeConfidence calculates confidence from signals
func ComputeConfidence(signals []Signal) ConfidenceScore {
	if len(signals) == 0 {
		return ConfidenceScore{
			Score:  50,
			Level:  "medium",
			Reason: "No confidence signals available",
		}
	}

	var totalWeight float64
	for _, s := range signals {
		totalWeight += s.Weight
	}

	// Normalize to 0-100
	score := int(totalWeight * 100)
	if score > 100 {
		score = 100
	}

	level := "low"
	if score >= 80 {
		level = "high"
	} else if score >= 50 {
		level = "medium"
	}

	reason := fmt.Sprintf("%d signals contributing to %d%% confidence", len(signals), score)

	return ConfidenceScore{
		Score:   score,
		Level:   level,
		Signals: signals,
		Reason:  reason,
	}
}

// Context represents context about where a finding occurred
type Context struct {
	InTest    bool   `json:"in_test"`    // Is this in a test file?
	InComment bool   `json:"in_comment"` // Is this in a comment?
	InDocs    bool   `json:"in_docs"`    // Is this in documentation?
	InExample bool   `json:"in_example"` // Is this in an example file?
	FileType  string `json:"file_type"`  // "source", "test", "config", "docs"
}

// ShouldFilter returns true if the context suggests this should be filtered
func (c Context) ShouldFilter() bool {
	return c.InTest || c.InComment || c.InDocs || c.InExample
}

// DetectContext analyzes a file path and content to determine context
func DetectContext(filePath string, lineContent string) Context {
	ctx := Context{FileType: "source"}

	// Detect test files
	testPatterns := []string{"_test.", ".test.", ".spec.", "/test/", "/tests/", "/__tests__/"}
	for _, p := range testPatterns {
		if strings.Contains(filePath, p) {
			ctx.InTest = true
			ctx.FileType = "test"
			break
		}
	}

	// Detect documentation
	docPatterns := []string{"/docs/", "/doc/", ".md", ".rst", ".txt"}
	for _, p := range docPatterns {
		if strings.Contains(filePath, p) {
			ctx.InDocs = true
			ctx.FileType = "docs"
			break
		}
	}

	// Detect examples
	examplePatterns := []string{"/example/", "/examples/", "/sample/", "/samples/", "/demo/"}
	for _, p := range examplePatterns {
		if strings.Contains(filePath, p) {
			ctx.InExample = true
			ctx.FileType = "example"
			break
		}
	}

	// Detect comments in line content (simple heuristic)
	trimmed := strings.TrimSpace(lineContent)
	if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
		ctx.InComment = true
	}

	return ctx
}
