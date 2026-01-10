package devops

import "time"

// Result holds all feature results
type Result struct {
	FeaturesRun []string `json:"features_run"`
	Summary     Summary  `json:"summary"`
	Findings    Findings `json:"findings"`
}

// Summary holds summaries from all features
type Summary struct {
	IaC           *IaCSummary           `json:"iac,omitempty"`
	Containers    *ContainersSummary    `json:"containers,omitempty"`
	GitHubActions *GitHubActionsSummary `json:"github_actions,omitempty"`
	DORA          *DORASummary          `json:"dora,omitempty"`
	Git           *GitSummary           `json:"git,omitempty"`
	Errors        []string              `json:"errors,omitempty"`
}

// Findings holds findings from all features
type Findings struct {
	IaC           []IaCFinding           `json:"iac,omitempty"`
	Containers    []ContainerFinding     `json:"containers,omitempty"`
	GitHubActions []GitHubActionsFinding `json:"github_actions,omitempty"`
	DORA          *DORAMetrics           `json:"dora,omitempty"`
	Git           *GitFindings           `json:"git,omitempty"`
}

// Feature summaries

// IaCSummary contains IaC security scan summary
type IaCSummary struct {
	TotalFindings    int                `json:"total_findings"`
	Critical         int                `json:"critical"`
	High             int                `json:"high"`
	Medium           int                `json:"medium"`
	Low              int                `json:"low"`
	ByType           map[string]int     `json:"by_type"`
	ByCategory       map[string]int     `json:"by_category,omitempty"`      // security vs best-practice
	FilesScanned     int                `json:"files_scanned"`
	Tool             string             `json:"tool"`
	Error            string             `json:"error,omitempty"`
	SecretsSummary   *IaCSecretsSummary `json:"secrets_summary,omitempty"`   // IaC secrets findings summary
	BestPractices    int                `json:"best_practices,omitempty"`    // count of best practice findings
}

// ContainersSummary contains container security summary
type ContainersSummary struct {
	TotalFindings      int            `json:"total_findings"`
	Critical           int            `json:"critical"`
	High               int            `json:"high"`
	Medium             int            `json:"medium"`
	Low                int            `json:"low"`
	DockerfilesScanned int            `json:"dockerfiles_scanned"`
	ImagesScanned      int            `json:"images_scanned"`
	ByImage            map[string]int `json:"by_image"`
	BySeverity         map[string]int `json:"by_severity"`
	Error              string         `json:"error,omitempty"`
}

// GitHubActionsSummary contains GitHub Actions security summary
type GitHubActionsSummary struct {
	TotalFindings    int            `json:"total_findings"`
	Critical         int            `json:"critical"`
	High             int            `json:"high"`
	Medium           int            `json:"medium"`
	Low              int            `json:"low"`
	ByCategory       map[string]int `json:"by_category"`
	WorkflowsScanned int            `json:"workflows_scanned"`
	Error            string         `json:"error,omitempty"`
}

// DORASummary contains DORA metrics summary
type DORASummary struct {
	DeploymentFrequency      float64 `json:"deployment_frequency"`
	DeploymentFrequencyClass string  `json:"deployment_frequency_class"`
	LeadTimeHours            float64 `json:"lead_time_hours"`
	LeadTimeClass            string  `json:"lead_time_class"`
	ChangeFailureRate        float64 `json:"change_failure_rate"`
	ChangeFailureClass       string  `json:"change_failure_class"`
	MTTRHours                float64 `json:"mttr_hours"`
	MTTRClass                string  `json:"mttr_class"`
	OverallClass             string  `json:"overall_class"`
	PeriodDays               int     `json:"period_days"`
	Error                    string  `json:"error,omitempty"`
	// PR-level cycle time metrics (LinearB alignment)
	AvgPickupHours float64 `json:"avg_pickup_hours,omitempty"`
	PickupClass    string  `json:"pickup_class,omitempty"`
	AvgReviewHours float64 `json:"avg_review_hours,omitempty"`
	ReviewClass    string  `json:"review_class,omitempty"`
	AvgMergeHours  float64 `json:"avg_merge_hours,omitempty"`
	MergeClass     string  `json:"merge_class,omitempty"`
	AvgPRSize      int     `json:"avg_pr_size,omitempty"`
	PRSizeClass    string  `json:"pr_size_class,omitempty"`
	TotalPRs       int     `json:"total_prs,omitempty"`
	// Rework rate (DORA 2025)
	ReworkRate      float64 `json:"rework_rate,omitempty"`
	ReworkClass     string  `json:"rework_class,omitempty"`
	RefactorRate    float64 `json:"refactor_rate,omitempty"`
	RefactorClass   string  `json:"refactor_class,omitempty"`
}

// GitSummary contains git insights summary
type GitSummary struct {
	TotalCommits          int    `json:"total_commits"`
	TotalContributors     int    `json:"total_contributors"`
	ActiveContributors30d int    `json:"active_contributors_30d"`
	ActiveContributors90d int    `json:"active_contributors_90d"`
	Commits90d            int    `json:"commits_90d"`
	BusFactor             int    `json:"bus_factor"`
	ActivityLevel         string `json:"activity_level"`
	Error                 string `json:"error,omitempty"`
}

// Finding types

// IaCFinding represents an IaC security or best-practice finding
type IaCFinding struct {
	RuleID      string `json:"rule_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	File        string `json:"file"`
	Line        int    `json:"line,omitempty"`
	Resource    string `json:"resource,omitempty"`
	Type        string `json:"type"`                  // terraform, kubernetes, dockerfile, cloudformation
	Category    string `json:"category,omitempty"`    // security, best-practice
	Resolution  string `json:"resolution,omitempty"`
	CheckType   string `json:"check_type,omitempty"`
	// Secret-related fields (for IaC secrets findings)
	SecretType  string `json:"secret_type,omitempty"`  // aws_key, password, token, etc.
	Snippet     string `json:"snippet,omitempty"`      // redacted code snippet
	IsSecret    bool   `json:"is_secret,omitempty"`    // true if this is a secrets finding
}

// IaCSecretsSummary contains IaC secrets scan summary
type IaCSecretsSummary struct {
	TotalFindings int            `json:"total_findings"`
	ByType        map[string]int `json:"by_type"`        // by IaC type
	BySecretType  map[string]int `json:"by_secret_type"` // by secret type
	BySeverity    map[string]int `json:"by_severity"`
	FilesScanned  int            `json:"files_scanned"`
}

// ContainerFinding represents a container vulnerability or lint finding
type ContainerFinding struct {
	VulnID       string   `json:"vuln_id"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Severity     string   `json:"severity"`
	Image        string   `json:"image"`
	Dockerfile   string   `json:"dockerfile"`
	Package      string   `json:"package"`
	Version      string   `json:"version"`
	FixedVersion string   `json:"fixed_version,omitempty"`
	CVSS         float64  `json:"cvss,omitempty"`
	References   []string `json:"references,omitempty"`
	Type         string   `json:"type,omitempty"`        // vulnerability, lint
	Line         int      `json:"line,omitempty"`        // line number for lint findings
	Remediation  string   `json:"remediation,omitempty"` // fix recommendation
}

// GitHubActionsFinding represents a GitHub Actions security finding
type GitHubActionsFinding struct {
	RuleID      string `json:"rule_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	File        string `json:"file"`
	Line        int    `json:"line,omitempty"`
	Category    string `json:"category"`
	Suggestion  string `json:"suggestion,omitempty"`
}

// DORAMetrics contains detailed DORA metrics
type DORAMetrics struct {
	DeploymentFrequency float64      `json:"deployment_frequency"`
	LeadTimeHours       float64      `json:"lead_time_hours"`
	ChangeFailureRate   float64      `json:"change_failure_rate"`
	MTTRHours           float64      `json:"mttr_hours"`
	TotalDeployments    int          `json:"total_deployments"`
	TotalCommits        int          `json:"total_commits"`
	Deployments         []Deployment `json:"deployments,omitempty"`
	// PR-level cycle time metrics (LinearB alignment)
	PRMetrics *PRMetrics `json:"pr_metrics,omitempty"`
	// Rework rate (DORA 2025)
	ReworkRate   float64 `json:"rework_rate,omitempty"`
	RefactorRate float64 `json:"refactor_rate,omitempty"`
}

// PRMetrics contains PR-level cycle time breakdowns (LinearB alignment)
type PRMetrics struct {
	TotalPRs       int     `json:"total_prs"`
	AvgPickupHours float64 `json:"avg_pickup_hours"` // PR opened → first review
	AvgReviewHours float64 `json:"avg_review_hours"` // First review → approval
	AvgMergeHours  float64 `json:"avg_merge_hours"`  // Approval → merge
	AvgCycleHours  float64 `json:"avg_cycle_hours"`  // PR opened → merge
	AvgPRSize      int     `json:"avg_pr_size"`      // additions + deletions
	// Tier classifications (LinearB benchmarks)
	PickupClass string `json:"pickup_class,omitempty"` // elite, good, fair, needs_focus
	ReviewClass string `json:"review_class,omitempty"`
	MergeClass  string `json:"merge_class,omitempty"`
	PRSizeClass string `json:"pr_size_class,omitempty"`
	// Individual PR data
	PRs []PRCycleTime `json:"prs,omitempty"`
}

// PRCycleTime contains cycle time data for a single PR
type PRCycleTime struct {
	Number      int       `json:"number"`
	Title       string    `json:"title"`
	Author      string    `json:"author"`
	CreatedAt   time.Time `json:"created_at"`
	MergedAt    time.Time `json:"merged_at"`
	PickupHours float64   `json:"pickup_hours"` // Time to first review
	ReviewHours float64   `json:"review_hours"` // First review to approval
	MergeHours  float64   `json:"merge_hours"`  // Approval to merge
	CycleHours  float64   `json:"cycle_hours"`  // Total cycle time
	Additions   int       `json:"additions"`
	Deletions   int       `json:"deletions"`
	Size        int       `json:"size"` // additions + deletions
}

// Deployment represents a release/deployment
type Deployment struct {
	Tag     string    `json:"tag"`
	Date    time.Time `json:"date"`
	Commits int       `json:"commits"`
	IsFix   bool      `json:"is_fix"`
}

// GitFindings contains git analysis findings
type GitFindings struct {
	Contributors   []Contributor   `json:"contributors"`
	HighChurnFiles []ChurnFile     `json:"high_churn_files,omitempty"`
	CodeAge        *CodeAgeStats   `json:"code_age,omitempty"`
	Patterns       *CommitPatterns `json:"patterns,omitempty"`
	Branches       *BranchInfo     `json:"branches,omitempty"`
}

// Contributor represents a git contributor
type Contributor struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	TotalCommits    int    `json:"total_commits"`
	Commits30d      int    `json:"commits_30d"`
	Commits90d      int    `json:"commits_90d"`
	Commits365d     int    `json:"commits_365d"`
	LinesAdded90d   int    `json:"lines_added_90d"`
	LinesRemoved90d int    `json:"lines_removed_90d"`
}

// ChurnFile represents a frequently modified file
type ChurnFile struct {
	File         string `json:"file"`
	Changes90d   int    `json:"changes_90d"`
	Contributors int    `json:"contributors"`
}

// CodeAgeStats represents code age distribution
type CodeAgeStats struct {
	SampledFiles int       `json:"sampled_files"`
	Age0to30     AgeBucket `json:"0_30_days"`
	Age31to90    AgeBucket `json:"31_90_days"`
	Age91to365   AgeBucket `json:"91_365_days"`
	Age365Plus   AgeBucket `json:"365_plus_days"`
}

// AgeBucket represents a code age bucket
type AgeBucket struct {
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

// CommitPatterns contains commit pattern analysis
type CommitPatterns struct {
	MostActiveDay      string `json:"most_active_day"`
	MostActiveHour     int    `json:"most_active_hour"`
	AvgCommitSizeLines int    `json:"avg_commit_size_lines"`
	FirstCommit        string `json:"first_commit"`
	LastCommit         string `json:"last_commit"`
	AvgCommitsPerWeek  int    `json:"avg_commits_per_week"`
}

// BranchInfo contains branch information
type BranchInfo struct {
	Current     string `json:"current"`
	Default     string `json:"default"`
	TotalCount  int    `json:"total_count"`
	RemoteCount int    `json:"remote_count"`
}
