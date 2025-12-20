package report

import "time"

// ReportType identifies the type of report
type ReportType string

const (
	TypeTechnical  ReportType = "technical"
	TypeExecutive  ReportType = "executive"
	TypeSummary    ReportType = "summary"
	TypeCompliance ReportType = "compliance"
)

// ReportMeta contains common report metadata
type ReportMeta struct {
	Repository   string
	Timestamp    time.Time
	ScannerName  string
	ScannerDesc  string
	Version      string
	Profile      string
}

// ReportConfig configures report generation
type ReportConfig struct {
	Types     []ReportType // Which report types to generate
	OutputDir string       // Where to write reports
	Format    string       // markdown, html, json
}

// DefaultConfig returns default report configuration
func DefaultConfig() ReportConfig {
	return ReportConfig{
		Types:  []ReportType{TypeTechnical, TypeExecutive},
		Format: "markdown",
	}
}

// ReportOutput represents a generated report
type ReportOutput struct {
	Type     ReportType
	Scanner  string
	Filename string
	Content  []byte
}

// ReportGenerator is the interface scanners implement for report generation
type ReportGenerator interface {
	// GenerateTechnicalReport creates a detailed technical report
	GenerateTechnicalReport() (string, error)

	// GenerateExecutiveReport creates a high-level executive summary
	GenerateExecutiveReport() (string, error)

	// ScannerName returns the name of the scanner
	ScannerName() string
}
