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

	// Enhanced ownership fields (v2.0)
	BusFactor           int     `json:"bus_factor,omitempty"`             // Number of people who need to leave before knowledge is lost
	BusFactorRisk       string  `json:"bus_factor_risk,omitempty"`        // critical, warning, healthy
	OwnershipCoverage   float64 `json:"ownership_coverage,omitempty"`     // 0-1, percentage of files with clear owners
	CodeownersIssues    int     `json:"codeowners_issues,omitempty"`      // Number of validation issues
	IsMonorepo          bool    `json:"is_monorepo,omitempty"`
	WorkspaceCount      int     `json:"workspace_count,omitempty"`
	GitHubTokenPresent  bool    `json:"github_token_present,omitempty"`   // Whether GitHub API was available
	PRAnalysisSkipped   bool    `json:"pr_analysis_skipped,omitempty"`    // Whether PR analysis was skipped due to volume
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

	// Enhanced ownership findings (v2.0)
	EnhancedOwnership  []EnhancedOwnership   `json:"enhanced_ownership,omitempty"`
	CodeownersAnalysis *CODEOWNERSAnalysis   `json:"codeowners_analysis,omitempty"`
	Monorepo           *MonorepoAnalysis     `json:"monorepo,omitempty"`
	IncidentContacts   []IncidentContact     `json:"incident_contacts,omitempty"`
	OrgSpecialists     []OrgSpecialist       `json:"org_specialists,omitempty"`
	PRAnalysis         *PRAnalysis           `json:"pr_analysis,omitempty"`
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

// ============================================================================
// Enhanced Ownership Types (v2.0)
// ============================================================================

// EnhancedOwnership represents multi-factor ownership with score breakdown
type EnhancedOwnership struct {
	Name            string         `json:"name"`
	Email           string         `json:"email"`
	OwnershipScore  float64        `json:"ownership_score"`  // 0-100 weighted score
	ScoreBreakdown  ScoreBreakdown `json:"score_breakdown"`  // Component scores
	ActivityStatus  string         `json:"activity_status"`  // active, recent, stale, inactive, abandoned
	LastActive      string         `json:"last_active"`      // ISO 8601 date
	Confidence      float64        `json:"confidence"`       // 0-1 data quality indicator
	PRReviewsGiven  int            `json:"pr_reviews_given"` // Number of PR reviews
	PRReviewsOnCode int            `json:"pr_reviews_on_code"` // Reviews on their code
}

// ScoreBreakdown shows individual scoring components
type ScoreBreakdown struct {
	CommitScore      float64 `json:"commit_score"`      // 0-30 (30% weight)
	ReviewScore      float64 `json:"review_score"`      // 0-25 (25% weight)
	LinesScore       float64 `json:"lines_score"`       // 0-20 (20% weight)
	RecencyScore     float64 `json:"recency_score"`     // 0-15 (15% weight)
	ConsistencyScore float64 `json:"consistency_score"` // 0-10 (10% weight)
}

// ============================================================================
// CODEOWNERS Analysis Types
// ============================================================================

// CODEOWNERSAnalysis contains validation results and recommendations
type CODEOWNERSAnalysis struct {
	FilePath         string                   `json:"file_path"`
	RulesCount       int                      `json:"rules_count"`
	ValidationIssues []CODEOWNERSIssue        `json:"validation_issues,omitempty"`
	Recommendations  []CODEOWNERSRecommendation `json:"recommendations,omitempty"`
	DriftAnalysis    *DriftAnalysis           `json:"drift_analysis,omitempty"`
	Coverage         float64                  `json:"coverage"` // Percentage of files covered
}

// CODEOWNERSIssue represents a validation issue
type CODEOWNERSIssue struct {
	ID          string `json:"id"`          // e.g., CO001, CO010
	Category    string `json:"category"`    // syntax, permission, coverage, staleness, best_practice
	Severity    string `json:"severity"`    // critical, high, medium, low
	Line        int    `json:"line"`        // Line number in CODEOWNERS file
	Pattern     string `json:"pattern"`     // The pattern with the issue
	Owner       string `json:"owner"`       // The owner with the issue (if applicable)
	Message     string `json:"message"`     // Human-readable description
	Remediation string `json:"remediation"` // How to fix
}

// CODEOWNERSRecommendation suggests improvements
type CODEOWNERSRecommendation struct {
	ID          string   `json:"id"`
	Priority    string   `json:"priority"`    // high, medium, low
	Type        string   `json:"type"`        // add_team, add_backup, protect_sensitive, update_stale, fix_drift
	Message     string   `json:"message"`
	AffectedPaths []string `json:"affected_paths,omitempty"`
	SuggestedOwners []string `json:"suggested_owners,omitempty"`
}

// DriftAnalysis compares declared vs actual ownership
type DriftAnalysis struct {
	HasDrift     bool        `json:"has_drift"`
	DriftScore   float64     `json:"drift_score"` // 0-100, higher = more drift
	DriftDetails []DriftItem `json:"drift_details,omitempty"`
}

// DriftItem represents a single case of ownership drift
type DriftItem struct {
	Path            string   `json:"path"`
	DeclaredOwners  []string `json:"declared_owners"`
	ActualTopOwners []string `json:"actual_top_owners"`
	OverlapScore    float64  `json:"overlap_score"` // 0-1, 1 = perfect match
}

// ============================================================================
// Monorepo Types
// ============================================================================

// MonorepoAnalysis contains workspace-level ownership
type MonorepoAnalysis struct {
	IsMonorepo    bool                 `json:"is_monorepo"`
	Type          string               `json:"type,omitempty"` // turborepo, lerna, nx, pnpm, cargo, go
	ConfigFile    string               `json:"config_file,omitempty"`
	Workspaces    []WorkspaceOwnership `json:"workspaces,omitempty"`
	CrossWorkspaceOwners []string      `json:"cross_workspace_owners,omitempty"` // People who work across multiple workspaces
}

// WorkspaceOwnership represents ownership for a single workspace
type WorkspaceOwnership struct {
	Name             string              `json:"name"`
	Path             string              `json:"path"`
	TopContributors  []EnhancedOwnership `json:"top_contributors"`
	BusFactor        int                 `json:"bus_factor"`
	BusFactorRisk    string              `json:"bus_factor_risk"` // critical, warning, healthy
	LanguagesUsed    []string            `json:"languages_used,omitempty"`
	HasOwnCodeowners bool                `json:"has_own_codeowners"`
}

// ============================================================================
// Incident Contact Types
// ============================================================================

// IncidentContact provides "who to contact" for a path
type IncidentContact struct {
	Path         string          `json:"path"`
	Primary      []ContactInfo   `json:"primary"`
	Backup       []ContactInfo   `json:"backup"`
	CodeownersMatch *CodeownerRule `json:"codeowners_match,omitempty"` // If path matches CODEOWNERS
}

// ContactInfo represents a single contact with scoring
type ContactInfo struct {
	Name             string  `json:"name"`
	Email            string  `json:"email"`
	ExpertiseScore   float64 `json:"expertise_score"`   // 0-1 how well they know this code
	AvailabilityScore float64 `json:"availability_score"` // 0-1 based on recent activity
	ReasonForContact string  `json:"reason_for_contact"` // Why they're recommended
}

// ============================================================================
// Cross-Org Specialist Types
// ============================================================================

// OrgSpecialist represents a developer with cross-repo expertise
type OrgSpecialist struct {
	Name        string            `json:"name"`
	Email       string            `json:"email"`
	Domains     []DomainExpertise `json:"domains"`
	TopDomain   string            `json:"top_domain"`
	TotalScore  float64           `json:"total_score"`
	ReposActive int               `json:"repos_active"` // Number of repos with contributions
}

// DomainExpertise tracks expertise in a specific domain
type DomainExpertise struct {
	Domain     string  `json:"domain"`     // supply-chain, security, frontend, etc.
	Score      float64 `json:"score"`      // 0-100
	FileCount  int     `json:"file_count"` // Files touched in this domain
	RepoCount  int     `json:"repo_count"` // Repos with domain work
	Confidence float64 `json:"confidence"` // 0-1 data quality indicator
}

// ============================================================================
// PR Analysis Types
// ============================================================================

// PRAnalysis contains PR review data
type PRAnalysis struct {
	Skipped     bool   `json:"skipped,omitempty"`
	SkipReason  string `json:"reason,omitempty"`
	TotalPRs    int    `json:"total_prs,omitempty"`
	Threshold   int    `json:"threshold,omitempty"`
	PRsAnalyzed int    `json:"prs_analyzed,omitempty"`
	Reviewers   []PRReviewer `json:"reviewers,omitempty"`
}

// PRReviewer tracks PR review activity
type PRReviewer struct {
	Name           string `json:"name"`
	Email          string `json:"email"`
	ReviewsGiven   int    `json:"reviews_given"`
	ApprovalsGiven int    `json:"approvals_given"`
	CommentsGiven  int    `json:"comments_given"`
	FilesReviewed  int    `json:"files_reviewed"`
}
