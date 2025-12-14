// Package codeownership provides code ownership and CODEOWNERS analysis
package codeownership

// Result holds all code ownership analysis results
type Result struct {
	FeaturesRun []string `json:"features_run"`
	Summary     Summary  `json:"summary"`
	Findings    Findings `json:"findings"`
}

// Summary holds ownership analysis summary
type Summary struct {
	TotalContributors int    `json:"total_contributors"`
	FilesAnalyzed     int    `json:"files_analyzed"`
	HasCodeowners     bool   `json:"has_codeowners"`
	CodeownersRules   int    `json:"codeowners_rules"`
	OrphanedFiles     int    `json:"orphaned_files"`
	PeriodDays        int    `json:"period_days"`
	Errors            []string `json:"errors,omitempty"`
}

// Findings holds code ownership findings
type Findings struct {
	Contributors  []Contributor   `json:"contributors"`
	Codeowners    []CodeownerRule `json:"codeowners,omitempty"`
	OrphanedFiles []string        `json:"orphaned_files,omitempty"`
	FileOwners    []FileOwnership `json:"file_owners,omitempty"`
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

// FileOwnership represents ownership information for a file
type FileOwnership struct {
	Path          string   `json:"path"`
	TopContributors []string `json:"top_contributors"`
	LastModified  string   `json:"last_modified,omitempty"`
	CommitCount   int      `json:"commit_count"`
}
