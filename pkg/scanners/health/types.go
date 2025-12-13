package health

// Result holds all feature results
type Result struct {
	FeaturesRun []string `json:"features_run"`
	Summary     Summary  `json:"summary"`
	Findings    Findings `json:"findings"`
}

// Summary holds summaries from all features
type Summary struct {
	Technology    *TechnologySummary    `json:"technology,omitempty"`
	Documentation *DocumentationSummary `json:"documentation,omitempty"`
	Tests         *TestsSummary         `json:"tests,omitempty"`
	Ownership     *OwnershipSummary     `json:"ownership,omitempty"`
	Errors        []string              `json:"errors,omitempty"`
}

// Findings holds findings from all features
type Findings struct {
	Technology    *TechnologyFindings    `json:"technology,omitempty"`
	Documentation *DocumentationFindings `json:"documentation,omitempty"`
	Tests         *TestsFindings         `json:"tests,omitempty"`
	Ownership     *OwnershipFindings     `json:"ownership,omitempty"`
}

// Feature summaries

// TechnologySummary contains technology discovery summary
type TechnologySummary struct {
	TotalTechnologies int            `json:"total_technologies"`
	ByCategory        map[string]int `json:"by_category"`
	PrimaryLanguages  []string       `json:"primary_languages"`
	Frameworks        []string       `json:"frameworks"`
	Databases         []string       `json:"databases"`
	CloudServices     []string       `json:"cloud_services"`
	Error             string         `json:"error,omitempty"`
}

// DocumentationSummary contains documentation analysis summary
type DocumentationSummary struct {
	OverallScore       float64 `json:"overall_score"`
	HasReadme          bool    `json:"has_readme"`
	HasChangelog       bool    `json:"has_changelog"`
	HasContributing    bool    `json:"has_contributing"`
	HasLicense         bool    `json:"has_license"`
	HasAPIDocs         bool    `json:"has_api_docs"`
	DocumentedFiles    int     `json:"documented_files"`
	TotalSourceFiles   int     `json:"total_source_files"`
	DocumentationRatio float64 `json:"documentation_ratio"`
	Error              string  `json:"error,omitempty"`
}

// TestsSummary contains test coverage summary
type TestsSummary struct {
	OverallCoverage   float64 `json:"overall_coverage"`
	LineCoverage      float64 `json:"line_coverage"`
	BranchCoverage    float64 `json:"branch_coverage"`
	TotalFiles        int     `json:"total_files"`
	CoveredFiles      int     `json:"covered_files"`
	UncoveredFiles    int     `json:"uncovered_files"`
	TestFramework     string  `json:"test_framework"`
	TotalTests        int     `json:"total_tests"`
	HasCoverageConfig bool    `json:"has_coverage_config"`
	CoverageThreshold float64 `json:"coverage_threshold,omitempty"`
	Error             string  `json:"error,omitempty"`
}

// OwnershipSummary contains code ownership summary
type OwnershipSummary struct {
	TotalContributors int    `json:"total_contributors"`
	FilesAnalyzed     int    `json:"files_analyzed"`
	HasCodeowners     bool   `json:"has_codeowners"`
	CodeownersRules   int    `json:"codeowners_rules"`
	OrphanedFiles     int    `json:"orphaned_files"`
	Error             string `json:"error,omitempty"`
}

// Finding types

// TechnologyFindings contains detected technologies
type TechnologyFindings struct {
	Technologies []Technology `json:"technologies"`
}

// Technology represents a detected technology
type Technology struct {
	Name       string `json:"name"`
	Category   string `json:"category"`
	Version    string `json:"version,omitempty"`
	Confidence int    `json:"confidence"` // 0-100
	Source     string `json:"source"`     // config, sbom, extension
}

// DocumentationFindings contains documentation analysis findings
type DocumentationFindings struct {
	ProjectDocs ProjectDocumentation `json:"project_docs"`
	CodeDocs    CodeDocumentation    `json:"code_docs"`
	Issues      []DocIssue           `json:"issues,omitempty"`
}

// ProjectDocumentation holds project-level documentation info
type ProjectDocumentation struct {
	HasReadme           bool           `json:"has_readme"`
	ReadmeQuality       ReadmeAnalysis `json:"readme_quality,omitempty"`
	HasChangelog        bool           `json:"has_changelog"`
	HasContributing     bool           `json:"has_contributing"`
	HasLicense          bool           `json:"has_license"`
	HasAPIDocs          bool           `json:"has_api_docs"`
	HasArchitectureDocs bool           `json:"has_architecture_docs"`
	DocumentationFiles  []DocFile      `json:"documentation_files"`
}

// ReadmeAnalysis contains README quality analysis
type ReadmeAnalysis struct {
	WordCount          int      `json:"word_count"`
	HasInstallation    bool     `json:"has_installation"`
	HasUsage           bool     `json:"has_usage"`
	HasExamples        bool     `json:"has_examples"`
	HasBadges          bool     `json:"has_badges"`
	HasTableOfContents bool     `json:"has_table_of_contents"`
	MissingSections    []string `json:"missing_sections,omitempty"`
}

// DocFile represents a documentation file
type DocFile struct {
	Path      string `json:"path"`
	Type      string `json:"type"`
	WordCount int    `json:"word_count"`
}

// CodeDocumentation contains code documentation analysis
type CodeDocumentation struct {
	TotalFiles          int                 `json:"total_files"`
	DocumentedFiles     int                 `json:"documented_files"`
	DocumentationRatio  float64             `json:"documentation_ratio"`
	TotalFunctions      int                 `json:"total_functions"`
	DocumentedFunctions int                 `json:"documented_functions"`
	FunctionDocRatio    float64             `json:"function_doc_ratio"`
	ByLanguage          map[string]LangDocs `json:"by_language"`
	UndocumentedPublic  []string            `json:"undocumented_public,omitempty"`
}

// LangDocs contains documentation stats for a language
type LangDocs struct {
	TotalFiles      int     `json:"total_files"`
	DocumentedFiles int     `json:"documented_files"`
	Ratio           float64 `json:"ratio"`
}

// DocIssue represents a documentation issue
type DocIssue struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	File        string `json:"file,omitempty"`
	Suggestion  string `json:"suggestion,omitempty"`
}

// TestsFindings contains test coverage findings
type TestsFindings struct {
	Coverage       CoverageData       `json:"coverage"`
	Infrastructure TestInfrastructure `json:"infrastructure"`
	Issues         []TestIssue        `json:"issues,omitempty"`
}

// CoverageData contains coverage metrics
type CoverageData struct {
	OverallCoverage float64            `json:"overall_coverage"`
	LineCoverage    float64            `json:"line_coverage"`
	BranchCoverage  float64            `json:"branch_coverage"`
	TotalFiles      int                `json:"total_files"`
	CoveredFiles    int                `json:"covered_files"`
	UncoveredFiles  []string           `json:"uncovered_files,omitempty"`
	ByDirectory     map[string]float64 `json:"by_directory,omitempty"`
	LowCoverage     []FileCoverage     `json:"low_coverage_files,omitempty"`
}

// FileCoverage contains coverage for a single file
type FileCoverage struct {
	File           string  `json:"file"`
	LineCoverage   float64 `json:"line_coverage"`
	BranchCoverage float64 `json:"branch_coverage,omitempty"`
	UncoveredLines []int   `json:"uncovered_lines,omitempty"`
}

// TestInfrastructure contains test infrastructure info
type TestInfrastructure struct {
	Framework         string   `json:"framework"`
	TotalTests        int      `json:"total_tests"`
	TestFiles         int      `json:"test_files"`
	HasCoverageConfig bool     `json:"has_coverage_config"`
	CoverageThreshold float64  `json:"coverage_threshold,omitempty"`
	CIIntegration     []string `json:"ci_integration,omitempty"`
}

// TestIssue represents a test coverage issue
type TestIssue struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	File        string `json:"file,omitempty"`
	Suggestion  string `json:"suggestion,omitempty"`
}

// OwnershipFindings contains code ownership findings
type OwnershipFindings struct {
	Contributors  []Contributor   `json:"contributors"`
	Codeowners    []CodeownerRule `json:"codeowners,omitempty"`
	OrphanedFiles []string        `json:"orphaned_files,omitempty"`
}

// Contributor represents a code contributor
type Contributor struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Commits      int    `json:"commits"`
	FilesTouched int    `json:"files_touched"`
	LinesAdded   int    `json:"lines_added"`
	LinesRemoved int    `json:"lines_removed"`
}

// CodeownerRule represents a CODEOWNERS rule
type CodeownerRule struct {
	Pattern string   `json:"pattern"`
	Owners  []string `json:"owners"`
}
