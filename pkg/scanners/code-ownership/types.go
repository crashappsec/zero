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
	TotalContributors int            `json:"total_contributors"`
	FilesAnalyzed     int            `json:"files_analyzed"`
	HasCodeowners     bool           `json:"has_codeowners"`
	CodeownersRules   int            `json:"codeowners_rules"`
	OrphanedFiles     int            `json:"orphaned_files"`
	PeriodDays        int            `json:"period_days"`
	LanguagesDetected int            `json:"languages_detected,omitempty"`
	TopLanguages      []LanguageInfo `json:"top_languages,omitempty"`
	IsShallowClone    bool           `json:"is_shallow_clone,omitempty"`
	Warnings          []string       `json:"warnings,omitempty"`
	Errors            []string       `json:"errors,omitempty"`
}

// LanguageInfo holds summary info about a programming language in the repo
type LanguageInfo struct {
	Name       string  `json:"name"`
	FileCount  int     `json:"file_count"`
	Percentage float64 `json:"percentage"`
}

// Findings holds code ownership findings
type Findings struct {
	Contributors  []Contributor      `json:"contributors"`
	Codeowners    []CodeownerRule    `json:"codeowners,omitempty"`
	OrphanedFiles []string           `json:"orphaned_files,omitempty"`
	FileOwners    []FileOwnership    `json:"file_owners,omitempty"`
	Competencies  []DeveloperProfile `json:"competencies,omitempty"`
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

// DeveloperProfile represents a developer's competency profile across languages
type DeveloperProfile struct {
	Name            string          `json:"name"`
	Email           string          `json:"email"`
	TotalCommits    int             `json:"total_commits"`
	FeatureCommits  int             `json:"feature_commits"`
	BugFixCommits   int             `json:"bug_fix_commits"`
	RefactorCommits int             `json:"refactor_commits"`
	OtherCommits    int             `json:"other_commits"`
	Languages       []LanguageStats `json:"languages"`
	TopLanguage     string          `json:"top_language,omitempty"`
	CompetencyScore float64         `json:"competency_score"`
}

// LanguageStats tracks a developer's contributions in a specific language
type LanguageStats struct {
	Language       string  `json:"language"`
	FileCount      int     `json:"file_count"`      // Unique files touched in this language
	Commits        int     `json:"commits"`         // Total commits touching this language
	FeatureCommits int     `json:"feature_commits"` // Feature commits in this language
	BugFixCommits  int     `json:"bug_fix_commits"` // Bug fix commits in this language
	Percentage     float64 `json:"percentage"`      // Percentage of developer's total work
}

// CodeownerRule represents a CODEOWNERS rule
type CodeownerRule struct {
	Pattern string   `json:"pattern"`
	Owners  []string `json:"owners"`
}

// FileOwnership represents ownership information for a file
type FileOwnership struct {
	Path            string   `json:"path"`
	TopContributors []string `json:"top_contributors"`
	LastModified    string   `json:"last_modified,omitempty"`
	CommitCount     int      `json:"commit_count"`
}
