package code

// Result holds all feature results
type Result struct {
	FeaturesRun []string `json:"features_run"`
	Summary     Summary  `json:"summary"`
	Findings    Findings `json:"findings"`
}

// Summary holds summaries from all features
type Summary struct {
	Vulns    *VulnsSummary    `json:"vulns,omitempty"`
	Secrets  *SecretsSummary  `json:"secrets,omitempty"`
	API      *APISummary      `json:"api,omitempty"`
	TechDebt *TechDebtSummary `json:"tech_debt,omitempty"`
	Errors   []string         `json:"errors,omitempty"`
}

// Findings holds findings from all features
type Findings struct {
	Vulns    []VulnFinding    `json:"vulns,omitempty"`
	Secrets  []SecretFinding  `json:"secrets,omitempty"`
	API      []APIFinding     `json:"api,omitempty"`
	TechDebt *TechDebtResult  `json:"tech_debt,omitempty"`
}

// Feature summaries

// VulnsSummary contains code vulnerability summary
type VulnsSummary struct {
	TotalFindings int            `json:"total_findings"`
	Critical      int            `json:"critical"`
	High          int            `json:"high"`
	Medium        int            `json:"medium"`
	Low           int            `json:"low"`
	ByCWE         map[string]int `json:"by_cwe,omitempty"`
	ByCategory    map[string]int `json:"by_category,omitempty"`
	Error         string         `json:"error,omitempty"`
}

// SecretsSummary contains secret detection summary
type SecretsSummary struct {
	TotalFindings int            `json:"total_findings"`
	Critical      int            `json:"critical"`
	High          int            `json:"high"`
	Medium        int            `json:"medium"`
	Low           int            `json:"low"`
	RiskScore     int            `json:"risk_score"`
	RiskLevel     string         `json:"risk_level"`
	ByType        map[string]int `json:"by_type"`
	FilesAffected int            `json:"files_affected"`
	Error         string         `json:"error,omitempty"`
}

// APISummary contains API security summary
type APISummary struct {
	TotalFindings int            `json:"total_findings"`
	Critical      int            `json:"critical"`
	High          int            `json:"high"`
	Medium        int            `json:"medium"`
	Low           int            `json:"low"`
	ByCategory    map[string]int `json:"by_category"`
	Error         string         `json:"error,omitempty"`
}

// TechDebtSummary contains technical debt summary
type TechDebtSummary struct {
	TotalMarkers     int            `json:"total_markers"`
	TotalIssues      int            `json:"total_issues"`
	ComplexityIssues int            `json:"complexity_issues"`
	ByType           map[string]int `json:"by_type"`
	ByPriority       map[string]int `json:"by_priority"`
	FilesAffected    int            `json:"files_affected"`
	Error            string         `json:"error,omitempty"`
}

// Finding types

// VulnFinding represents a code vulnerability finding
type VulnFinding struct {
	RuleID      string   `json:"rule_id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Severity    string   `json:"severity"`
	File        string   `json:"file"`
	Line        int      `json:"line"`
	Column      int      `json:"column"`
	Category    string   `json:"category,omitempty"`
	CWE         []string `json:"cwe,omitempty"`
	OWASP       []string `json:"owasp,omitempty"`
	Fix         string   `json:"fix,omitempty"`
}

// SecretFinding represents a detected secret
type SecretFinding struct {
	RuleID   string `json:"rule_id"`
	Type     string `json:"type"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Snippet  string `json:"snippet"`
}

// APIFinding represents an API security finding
type APIFinding struct {
	RuleID      string `json:"rule_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Category    string `json:"category"`
	OWASPApi    string `json:"owasp_api,omitempty"`
}

// TechDebtResult contains technical debt findings
type TechDebtResult struct {
	Markers  []DebtMarker `json:"markers"`
	Issues   []DebtIssue  `json:"issues,omitempty"`
	Hotspots []FileDebt   `json:"hotspots"`
}

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
