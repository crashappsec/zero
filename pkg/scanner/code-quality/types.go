package codequality

// Result holds all feature results
type Result struct {
	FeaturesRun []string `json:"features_run"`
	Summary     Summary  `json:"summary"`
	Findings    Findings `json:"findings"`
}

// Summary holds summaries from all features
type Summary struct {
	TechDebt     *TechDebtSummary     `json:"tech_debt,omitempty"`
	Complexity   *ComplexitySummary   `json:"complexity,omitempty"`
	TestCoverage *TestCoverageSummary `json:"test_coverage,omitempty"`
	CodeDocs     *CodeDocsSummary     `json:"code_docs,omitempty"`
	Errors       []string             `json:"errors,omitempty"`
}

// Findings holds findings from all features
type Findings struct {
	TechDebt   *TechDebtResult   `json:"tech_debt,omitempty"`
	Complexity *ComplexityResult `json:"complexity,omitempty"`
}

// Feature summaries

// TechDebtSummary contains technical debt summary
type TechDebtSummary struct {
	TotalMarkers  int            `json:"total_markers"`
	TotalIssues   int            `json:"total_issues"`
	ByType        map[string]int `json:"by_type"`
	ByPriority    map[string]int `json:"by_priority"`
	FilesAffected int            `json:"files_affected"`
	Error         string         `json:"error,omitempty"`
}

// ComplexitySummary contains complexity analysis summary
type ComplexitySummary struct {
	TotalIssues   int            `json:"total_issues"`
	High          int            `json:"high"`
	Medium        int            `json:"medium"`
	Low           int            `json:"low"`
	ByType        map[string]int `json:"by_type"`
	FilesAffected int            `json:"files_affected"`
	Error         string         `json:"error,omitempty"`
}

// TestCoverageSummary contains test coverage summary
type TestCoverageSummary struct {
	HasTestFiles    bool     `json:"has_test_files"`
	TestFrameworks  []string `json:"test_frameworks"`
	CoverageReports []string `json:"coverage_reports,omitempty"`
	LineCoverage    float64  `json:"line_coverage,omitempty"`
	MeetsThreshold  bool     `json:"meets_threshold"`
	Error           string   `json:"error,omitempty"`
}

// CodeDocsSummary contains documentation analysis summary
type CodeDocsSummary struct {
	HasReadme    bool   `json:"has_readme"`
	ReadmeFile   string `json:"readme_file,omitempty"`
	HasChangelog bool   `json:"has_changelog"`
	HasApiDocs   bool   `json:"has_api_docs"`
	Score        int    `json:"score"` // 0-100 documentation score
	Error        string `json:"error,omitempty"`
}

// Result types

// TechDebtResult contains technical debt findings
type TechDebtResult struct {
	Markers  []DebtMarker `json:"markers"`
	Issues   []DebtIssue  `json:"issues,omitempty"`
	Hotspots []FileDebt   `json:"hotspots"`
}

// ComplexityResult contains complexity findings
type ComplexityResult struct {
	Issues []ComplexityIssue `json:"issues"`
}

// Finding types

// DebtMarker represents a TODO/FIXME marker
type DebtMarker struct {
	Type     string `json:"type"`
	Priority string `json:"priority"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Text     string `json:"text"`
	Author   string `json:"author,omitempty"`
}

// DebtIssue represents a code smell or issue
type DebtIssue struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	File        string `json:"file"`
	Line        int    `json:"line,omitempty"`
	Description string `json:"description"`
	Suggestion  string `json:"suggestion,omitempty"`
	Source      string `json:"source,omitempty"`
}

// FileDebt represents debt statistics for a file
type FileDebt struct {
	File         string         `json:"file"`
	TotalMarkers int            `json:"total_markers"`
	ByType       map[string]int `json:"by_type"`
}

// ComplexityIssue represents a complexity finding
type ComplexityIssue struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	File        string `json:"file"`
	Line        int    `json:"line,omitempty"`
	Description string `json:"description"`
	Suggestion  string `json:"suggestion,omitempty"`
	Source      string `json:"source,omitempty"`
}
