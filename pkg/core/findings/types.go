// Package findings provides standardized finding types for all scanners
package findings

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// Finding represents a single security or quality finding
type Finding struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Severity    Severity       `json:"severity"`
	Confidence  Confidence     `json:"confidence"`
	Category    string         `json:"category"`
	Scanner     string         `json:"scanner"`
	Location    *Location      `json:"location,omitempty"`
	Evidence    *Evidence      `json:"evidence,omitempty"`
	Remediation string         `json:"remediation,omitempty"`
	References  []string       `json:"references,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

// Location represents where a finding was detected
type Location struct {
	File      string `json:"file"`
	Line      int    `json:"line,omitempty"`
	EndLine   int    `json:"end_line,omitempty"`
	Column    int    `json:"column,omitempty"`
	EndColumn int    `json:"end_column,omitempty"`
	Snippet   string `json:"snippet,omitempty"`
}

// Evidence captures comprehensive data for analyst review and rule improvement
type Evidence struct {
	// Source identification - where the finding was detected
	GitHubOrg  string `json:"github_org,omitempty"`  // e.g., "expressjs"
	GitHubRepo string `json:"github_repo,omitempty"` // e.g., "express"
	RepoURL    string `json:"repo_url,omitempty"`    // Full GitHub URL
	CommitSHA  string `json:"commit_sha,omitempty"`  // Exact commit scanned
	Branch     string `json:"branch,omitempty"`      // Branch name

	// Location within the repository
	FilePath    string `json:"file_path"`              // Relative path in repo
	LineStart   int    `json:"line_start,omitempty"`   // Starting line number
	LineEnd     int    `json:"line_end,omitempty"`     // Ending line number
	ColumnStart int    `json:"column_start,omitempty"` // Starting column
	ColumnEnd   int    `json:"column_end,omitempty"`   // Ending column

	// Match details - what triggered the detection
	MatchedText   string   `json:"matched_text,omitempty"`   // Exact text that triggered
	ContextBefore []string `json:"context_before,omitempty"` // Lines before match
	ContextAfter  []string `json:"context_after,omitempty"`  // Lines after match

	// Rule information - which rule matched and how
	RuleID      string `json:"rule_id"`                 // Unique rule identifier
	RuleVersion string `json:"rule_version,omitempty"`  // Semver of rule
	RuleSource  string `json:"rule_source,omitempty"`   // "rag", "semgrep-community", "custom"
	PatternType string `json:"pattern_type,omitempty"`  // "regex", "semgrep-semantic", "import"
	RawPattern  string `json:"raw_pattern,omitempty"`   // The actual pattern/regex that matched
	RAGFile     string `json:"rag_file,omitempty"`      // Source RAG markdown file path
	RAGCategory string `json:"rag_category,omitempty"`  // RAG category (e.g., "code-security")

	// Confidence signals - why we think this is a true positive
	ConfidenceScore   float64            `json:"confidence_score,omitempty"`   // 0.0-1.0 overall score
	ConfidenceSignals map[string]float64 `json:"confidence_signals,omitempty"` // Individual factors
	ConfidenceReason  string             `json:"confidence_reason,omitempty"`  // Human-readable explanation

	// Scan metadata - context about when/how it was found
	ScannerName    string    `json:"scanner_name,omitempty"`    // e.g., "code-security"
	ScannerVersion string    `json:"scanner_version,omitempty"` // Zero version
	ScanTimestamp  time.Time `json:"scan_timestamp,omitempty"`  // When scan ran
	ScanProfile    string    `json:"scan_profile,omitempty"`    // "all-quick", "all-complete", etc.

	// Fingerprint for deduplication and tracking
	Fingerprint string `json:"fingerprint,omitempty"` // Hash of key fields for dedup
}

// ComputeFingerprint generates a stable fingerprint for the evidence
// This allows tracking the same finding across scans
func (e *Evidence) ComputeFingerprint() string {
	// Fingerprint based on: repo + file + rule + matched text
	// This identifies the same finding even if line numbers change
	data := fmt.Sprintf("%s/%s:%s:%s:%s",
		e.GitHubOrg,
		e.GitHubRepo,
		e.FilePath,
		e.RuleID,
		e.MatchedText,
	)
	hash := sha256.Sum256([]byte(data))
	e.Fingerprint = hex.EncodeToString(hash[:16]) // First 16 bytes = 32 hex chars
	return e.Fingerprint
}

// NewEvidence creates a new Evidence with required fields
func NewEvidence(org, repo, filePath, ruleID string) *Evidence {
	e := &Evidence{
		GitHubOrg:     org,
		GitHubRepo:    repo,
		FilePath:      filePath,
		RuleID:        ruleID,
		ScanTimestamp: time.Now(),
	}
	return e
}

// WithMatch sets the matched text and computes fingerprint
func (e *Evidence) WithMatch(text string, lineStart, lineEnd int) *Evidence {
	e.MatchedText = text
	e.LineStart = lineStart
	e.LineEnd = lineEnd
	e.ComputeFingerprint()
	return e
}

// WithContext sets the surrounding context lines
func (e *Evidence) WithContext(before, after []string) *Evidence {
	e.ContextBefore = before
	e.ContextAfter = after
	return e
}

// WithRule sets the rule information
func (e *Evidence) WithRule(ruleID, ruleVersion, ruleSource, patternType, rawPattern string) *Evidence {
	e.RuleID = ruleID
	e.RuleVersion = ruleVersion
	e.RuleSource = ruleSource
	e.PatternType = patternType
	e.RawPattern = rawPattern
	return e
}

// WithRAGSource sets the RAG source information
func (e *Evidence) WithRAGSource(ragFile, ragCategory string) *Evidence {
	e.RAGFile = ragFile
	e.RAGCategory = ragCategory
	return e
}

// WithConfidence sets confidence signals
func (e *Evidence) WithConfidence(score float64, reason string, signals map[string]float64) *Evidence {
	e.ConfidenceScore = score
	e.ConfidenceReason = reason
	e.ConfidenceSignals = signals
	return e
}

// WithScanContext sets scan metadata
func (e *Evidence) WithScanContext(scannerName, scannerVersion, scanProfile string) *Evidence {
	e.ScannerName = scannerName
	e.ScannerVersion = scannerVersion
	e.ScanProfile = scanProfile
	return e
}

// WithGitContext sets git repository context
func (e *Evidence) WithGitContext(repoURL, commitSHA, branch string) *Evidence {
	e.RepoURL = repoURL
	e.CommitSHA = commitSHA
	e.Branch = branch
	return e
}

// FindingSet represents a collection of findings with metadata
type FindingSet struct {
	Scanner   string    `json:"scanner"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	Findings  []Finding `json:"findings"`
	Summary   Summary   `json:"summary"`
}

// Summary provides aggregated counts
type Summary struct {
	Total      int            `json:"total"`
	BySeverity map[string]int `json:"by_severity"`
	ByCategory map[string]int `json:"by_category"`
}

// NewFindingSet creates a new finding set with computed summary
func NewFindingSet(scanner string, findings []Finding) *FindingSet {
	fs := &FindingSet{
		Scanner:   scanner,
		Timestamp: time.Now(),
		Findings:  findings,
	}
	fs.ComputeSummary()
	return fs
}

// ComputeSummary calculates summary statistics
func (fs *FindingSet) ComputeSummary() {
	fs.Summary = Summary{
		Total:      len(fs.Findings),
		BySeverity: make(map[string]int),
		ByCategory: make(map[string]int),
	}
	for _, f := range fs.Findings {
		fs.Summary.BySeverity[string(f.Severity)]++
		fs.Summary.ByCategory[f.Category]++
	}
}

// NewFinding creates a new finding with defaults
func NewFinding(id, title, description string, severity Severity) Finding {
	return Finding{
		ID:          id,
		Title:       title,
		Description: description,
		Severity:    severity,
		Confidence:  ConfidenceHigh,
		CreatedAt:   time.Now(),
	}
}

// WithLocation adds location information
func (f Finding) WithLocation(file string, line int) Finding {
	f.Location = &Location{
		File: file,
		Line: line,
	}
	return f
}

// WithCategory sets the category
func (f Finding) WithCategory(category string) Finding {
	f.Category = category
	return f
}

// WithScanner sets the scanner name
func (f Finding) WithScanner(scanner string) Finding {
	f.Scanner = scanner
	return f
}

// WithRemediation sets the remediation advice
func (f Finding) WithRemediation(remediation string) Finding {
	f.Remediation = remediation
	return f
}

// WithReferences adds reference URLs
func (f Finding) WithReferences(refs ...string) Finding {
	f.References = append(f.References, refs...)
	return f
}

// WithMetadata adds a metadata key-value pair
func (f Finding) WithMetadata(key string, value any) Finding {
	if f.Metadata == nil {
		f.Metadata = make(map[string]any)
	}
	f.Metadata[key] = value
	return f
}

// WithEvidence attaches evidence for analyst review
func (f Finding) WithEvidence(evidence *Evidence) Finding {
	f.Evidence = evidence
	return f
}

// CriticalCount returns count of critical findings
func (fs *FindingSet) CriticalCount() int {
	return fs.Summary.BySeverity[string(SeverityCritical)]
}

// HighCount returns count of high findings
func (fs *FindingSet) HighCount() int {
	return fs.Summary.BySeverity[string(SeverityHigh)]
}

// MediumCount returns count of medium findings
func (fs *FindingSet) MediumCount() int {
	return fs.Summary.BySeverity[string(SeverityMedium)]
}

// LowCount returns count of low findings
func (fs *FindingSet) LowCount() int {
	return fs.Summary.BySeverity[string(SeverityLow)]
}

// InfoCount returns count of info findings
func (fs *FindingSet) InfoCount() int {
	return fs.Summary.BySeverity[string(SeverityInfo)]
}

// HasCritical returns true if there are critical findings
func (fs *FindingSet) HasCritical() bool {
	return fs.CriticalCount() > 0
}

// HasHigh returns true if there are high or critical findings
func (fs *FindingSet) HasHigh() bool {
	return fs.CriticalCount() > 0 || fs.HighCount() > 0
}
