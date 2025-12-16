// Package scanner provides the scanner framework for security analysis
package scanner

import (
	"context"
	"encoding/json"
	"os"
	"time"
)

// Scanner defines the interface all scanners must implement
type Scanner interface {
	// Name returns the scanner identifier (e.g., "package-vulns")
	Name() string

	// Description returns a human-readable description
	Description() string

	// Run executes the scanner and returns results
	Run(ctx context.Context, opts *ScanOptions) (*ScanResult, error)

	// Dependencies returns scanners that must run first (e.g., "package-sbom")
	Dependencies() []string

	// EstimateDuration returns estimated duration based on file count
	EstimateDuration(fileCount int) time.Duration
}

// ScanOptions contains inputs for a scanner run
type ScanOptions struct {
	// RepoPath is the path to the repository to scan
	RepoPath string

	// OutputDir is where to write results (e.g., .zero/repos/owner/repo/analysis)
	OutputDir string

	// SBOMPath is path to pre-generated SBOM (optional, for scanners that need it)
	SBOMPath string

	// Timeout is the maximum duration for this scanner
	Timeout time.Duration

	// Verbose enables verbose logging
	Verbose bool

	// OnStatus is called with progress messages during scanning
	// This allows scanners to report what they're doing in real-time
	OnStatus func(message string)

	// ExtraArgs contains scanner-specific options
	ExtraArgs map[string]string

	// FeatureConfig contains feature-specific configuration for super scanners
	FeatureConfig map[string]interface{}
}

// Feature describes a feature within a super scanner
type Feature struct {
	Name        string
	Description string
	Default     bool // Enabled by default?
}

// ScanResult represents scanner output
type ScanResult struct {
	// Analyzer is the scanner name
	Analyzer string `json:"analyzer"`

	// Version is the scanner version
	Version string `json:"version"`

	// Timestamp is when the scan completed
	Timestamp string `json:"timestamp"`

	// DurationSeconds is how long the scan took
	DurationSeconds int `json:"duration_seconds"`

	// Repository is the repo that was scanned
	Repository string `json:"repository,omitempty"`

	// Summary contains aggregated findings info
	Summary json.RawMessage `json:"summary"`

	// Findings contains detailed findings (optional for some scanners)
	Findings json.RawMessage `json:"findings,omitempty"`

	// Metadata contains scanner-specific metadata
	Metadata json.RawMessage `json:"metadata,omitempty"`

	// Error contains error message if scan failed
	Error string `json:"error,omitempty"`
}

// ScanSummary is a common summary structure used by many scanners
type ScanSummary struct {
	TotalFindings int            `json:"total_findings,omitempty"`
	Critical      int            `json:"critical,omitempty"`
	High          int            `json:"high,omitempty"`
	Medium        int            `json:"medium,omitempty"`
	Low           int            `json:"low,omitempty"`
	Info          int            `json:"info,omitempty"`
	TotalPackages int            `json:"total_packages,omitempty"`
	ByType        map[string]int `json:"by_type,omitempty"`
	Status        string         `json:"status,omitempty"`
}

// Finding is a common finding structure
type Finding struct {
	ID          string          `json:"id,omitempty"`
	Title       string          `json:"title,omitempty"`
	Description string          `json:"description,omitempty"`
	Severity    string          `json:"severity"`
	Category    string          `json:"category,omitempty"`
	Package     string          `json:"package,omitempty"`
	Version     string          `json:"version,omitempty"`
	File        string          `json:"file,omitempty"`
	Line        int             `json:"line,omitempty"`
	Confidence  string          `json:"confidence,omitempty"`
	References  []string        `json:"references,omitempty"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
}

// NewScanResult creates a new scan result with common fields populated
func NewScanResult(analyzer, version string, start time.Time) *ScanResult {
	return &ScanResult{
		Analyzer:        analyzer,
		Version:         version,
		Timestamp:       time.Now().UTC().Format(time.RFC3339),
		DurationSeconds: int(time.Since(start).Seconds()),
	}
}

// SetSummary marshals and sets the summary
func (r *ScanResult) SetSummary(summary interface{}) error {
	data, err := json.Marshal(summary)
	if err != nil {
		return err
	}
	r.Summary = data
	return nil
}

// SetFindings marshals and sets the findings
func (r *ScanResult) SetFindings(findings interface{}) error {
	data, err := json.Marshal(findings)
	if err != nil {
		return err
	}
	r.Findings = data
	return nil
}

// SetMetadata marshals and sets the metadata
func (r *ScanResult) SetMetadata(metadata interface{}) error {
	data, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	r.Metadata = data
	return nil
}

// WriteJSON writes the result to a JSON file
func (r *ScanResult) WriteJSON(path string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
