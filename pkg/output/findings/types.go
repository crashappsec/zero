// Package findings provides standardized finding types for all scanners
package findings

import "time"

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
