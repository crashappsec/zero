package codesecurity

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// GitHistorySecurityScanner scans git history for files that should have been purged
type GitHistorySecurityScanner struct {
	config           GitHistorySecurityConfig
	gitignoreRules   []gitignoreRule
	sensitivePatterns []sensitiveFilePattern
}

// gitignoreRule represents a parsed gitignore pattern
type gitignoreRule struct {
	Pattern    string
	Regex      *regexp.Regexp
	IsNegation bool
	IsDir      bool
	SourceLine int
}

// sensitiveFilePattern defines patterns for sensitive files
type sensitiveFilePattern struct {
	Pattern     string
	Regex       *regexp.Regexp
	Category    string
	Severity    string
	Description string
}

// GitHistorySecurityResult holds results from git history security scanning
type GitHistorySecurityResult struct {
	// Files matching gitignore patterns found in history
	GitignoreViolations []GitignoreViolation `json:"gitignore_violations"`

	// Sensitive files found in history
	SensitiveFiles []SensitiveFileFinding `json:"sensitive_files"`

	// Purge recommendations
	PurgeRecommendations []PurgeRecommendation `json:"purge_recommendations"`

	// Timeline of when sensitive files were added
	Timeline []HistoricalEvent `json:"timeline"`

	// Summary statistics
	Summary GitHistorySecuritySummary `json:"summary"`
}

// GitignoreViolation represents a file in history that matches gitignore
type GitignoreViolation struct {
	File           string      `json:"file"`
	GitignoreRule  string      `json:"gitignore_rule"`
	FirstCommit    *CommitInfo `json:"first_commit"`
	LastCommit     *CommitInfo `json:"last_commit,omitempty"`
	StillExists    bool        `json:"still_exists"`
	WasRemoved     bool        `json:"was_removed"`
	GitignoreAdded string      `json:"gitignore_added,omitempty"` // When the gitignore rule was added
}

// SensitiveFileFinding represents a sensitive file found in history
type SensitiveFileFinding struct {
	File        string      `json:"file"`
	Category    string      `json:"category"`
	Severity    string      `json:"severity"`
	Description string      `json:"description"`
	FirstCommit *CommitInfo `json:"first_commit"`
	LastCommit  *CommitInfo `json:"last_commit,omitempty"`
	StillExists bool        `json:"still_exists"`
	WasRemoved  bool        `json:"was_removed"`
	SizeBytes   int64       `json:"size_bytes,omitempty"`
}

// PurgeRecommendation recommends files to purge from history
type PurgeRecommendation struct {
	File        string   `json:"file"`
	Reason      string   `json:"reason"`
	Severity    string   `json:"severity"`
	Priority    int      `json:"priority"` // 1 = highest priority
	Command     string   `json:"command"`  // BFG or git-filter-repo command
	Alternative string   `json:"alternative,omitempty"`
	AffectedCommits int  `json:"affected_commits"`
}

// HistoricalEvent represents a timeline event
type HistoricalEvent struct {
	Date        string `json:"date"`
	EventType   string `json:"event_type"` // "committed", "gitignored", "removed"
	File        string `json:"file"`
	CommitHash  string `json:"commit_hash"`
	Author      string `json:"author"`
	Description string `json:"description"`
}

// GitHistorySecuritySummary contains summary statistics
type GitHistorySecuritySummary struct {
	TotalViolations        int            `json:"total_violations"`
	GitignoreViolations    int            `json:"gitignore_violations"`
	SensitiveFilesFound    int            `json:"sensitive_files_found"`
	FilesToPurge           int            `json:"files_to_purge"`
	CommitsScanned         int            `json:"commits_scanned"`
	ByCategory             map[string]int `json:"by_category"`
	BySeverity             map[string]int `json:"by_severity"`
	RiskScore              int            `json:"risk_score"`
	RiskLevel              string         `json:"risk_level"`
}

// NewGitHistorySecurityScanner creates a new git history security scanner
func NewGitHistorySecurityScanner(config GitHistorySecurityConfig) *GitHistorySecurityScanner {
	scanner := &GitHistorySecurityScanner{config: config}
	scanner.loadSensitivePatterns()
	return scanner
}

// loadSensitivePatterns loads patterns for sensitive file detection from RAG or falls back to hardcoded
func (s *GitHistorySecurityScanner) loadSensitivePatterns() {
	// Try to load from RAG first
	_, sensitiveRAG, err := LoadRAGGitHistoryPatterns()
	if err == nil && len(sensitiveRAG) > 0 {
		s.sensitivePatterns = ConvertToSensitiveFilePatterns(sensitiveRAG)
		return
	}

	// Fall back to hardcoded patterns if RAG not available
	patterns := []struct {
		pattern     string
		category    string
		severity    string
		description string
	}{
		// Environment and config files
		{`\.env$`, "credentials", "critical", "Environment file with potential secrets"},
		{`\.env\.(local|dev|development|prod|production|staging|test)$`, "credentials", "critical", "Environment file with potential secrets"},
		{`\.env\..*$`, "credentials", "high", "Environment file variant"},
		{`config\.json$`, "configuration", "medium", "Configuration file (may contain secrets)"},
		{`config\.ya?ml$`, "configuration", "medium", "Configuration file (may contain secrets)"},
		{`settings\.json$`, "configuration", "medium", "Settings file (may contain secrets)"},
		{`secrets\.json$`, "credentials", "critical", "Secrets configuration file"},
		{`secrets\.ya?ml$`, "credentials", "critical", "Secrets configuration file"},

		// Cloud credentials
		{`credentials\.json$`, "credentials", "critical", "Credentials file"},
		{`service[-_]?account.*\.json$`, "credentials", "critical", "GCP service account key"},
		{`\.aws/credentials$`, "credentials", "critical", "AWS credentials file"},
		{`\.aws/config$`, "credentials", "high", "AWS config file"},
		{`gcloud.*\.json$`, "credentials", "critical", "GCP credentials file"},
		{`\.gcp[-_]?credentials.*\.json$`, "credentials", "critical", "GCP credentials file"},

		// SSH and certificates
		{`id_rsa$`, "keys", "critical", "SSH private key"},
		{`id_dsa$`, "keys", "critical", "SSH private key"},
		{`id_ecdsa$`, "keys", "critical", "SSH private key"},
		{`id_ed25519$`, "keys", "critical", "SSH private key"},
		{`\.pem$`, "keys", "critical", "PEM key/certificate file"},
		{`\.key$`, "keys", "critical", "Private key file"},
		{`\.p12$`, "keys", "critical", "PKCS12 certificate file"},
		{`\.pfx$`, "keys", "critical", "PFX certificate file"},
		{`\.keystore$`, "keys", "critical", "Java keystore file"},
		{`\.jks$`, "keys", "critical", "Java keystore file"},

		// Database files
		{`\.sqlite3?$`, "database", "high", "SQLite database file"},
		{`\.db$`, "database", "high", "Database file"},
		{`\.sql$`, "database", "medium", "SQL dump file (may contain data)"},
		{`dump\.sql$`, "database", "high", "Database dump file"},
		{`backup.*\.sql$`, "database", "high", "Database backup file"},

		// Build artifacts
		{`node_modules/`, "build_artifact", "low", "Node.js dependencies directory"},
		{`vendor/`, "build_artifact", "low", "Vendor dependencies directory"},
		{`\.pyc$`, "build_artifact", "info", "Python bytecode"},
		{`__pycache__/`, "build_artifact", "info", "Python cache directory"},
		{`\.class$`, "build_artifact", "info", "Java bytecode"},
		{`\.jar$`, "build_artifact", "low", "Java archive"},
		{`\.war$`, "build_artifact", "low", "Java web archive"},
		{`dist/`, "build_artifact", "low", "Build output directory"},
		{`build/`, "build_artifact", "low", "Build output directory"},
		{`\.o$`, "build_artifact", "info", "Object file"},
		{`\.a$`, "build_artifact", "info", "Static library"},
		{`\.so$`, "build_artifact", "low", "Shared library"},
		{`\.dylib$`, "build_artifact", "low", "macOS dynamic library"},
		{`\.dll$`, "build_artifact", "low", "Windows DLL"},
		{`\.exe$`, "build_artifact", "low", "Windows executable"},

		// IDE and editor files
		{`\.idea/`, "ide", "info", "JetBrains IDE directory"},
		{`\.vscode/`, "ide", "info", "VS Code directory"},
		{`\.sublime-.*`, "ide", "info", "Sublime Text file"},
		{`.*\.swp$`, "ide", "info", "Vim swap file"},
		{`.*\.swo$`, "ide", "info", "Vim swap file"},
		{`\.DS_Store$`, "ide", "info", "macOS metadata file"},
		{`Thumbs\.db$`, "ide", "info", "Windows thumbnail cache"},

		// Logs and temporary files
		{`\.log$`, "logs", "low", "Log file (may contain sensitive data)"},
		{`npm-debug\.log.*`, "logs", "low", "npm debug log"},
		{`yarn-error\.log.*`, "logs", "low", "Yarn error log"},
		{`debug\.log$`, "logs", "low", "Debug log file"},
		{`\.tmp$`, "temporary", "info", "Temporary file"},
		{`\.temp$`, "temporary", "info", "Temporary file"},
		{`\.cache$`, "temporary", "info", "Cache file"},

		// Docker and Kubernetes secrets
		{`docker-compose\.override\.ya?ml$`, "configuration", "medium", "Docker compose override (may contain secrets)"},
		{`\.docker/config\.json$`, "credentials", "high", "Docker config with potential registry credentials"},
		{`kubeconfig$`, "credentials", "critical", "Kubernetes config with cluster credentials"},
		{`\.kube/config$`, "credentials", "critical", "Kubernetes config file"},

		// Password and secret files
		{`password.*`, "credentials", "critical", "Password file"},
		{`.*password.*\.txt$`, "credentials", "critical", "Password file"},
		{`secret.*\.txt$`, "credentials", "critical", "Secrets file"},
		{`token.*\.txt$`, "credentials", "high", "Token file"},
		{`api[-_]?key.*\.txt$`, "credentials", "critical", "API key file"},

		// Terraform state
		{`\.tfstate$`, "infrastructure", "critical", "Terraform state file (contains secrets)"},
		{`\.tfstate\.backup$`, "infrastructure", "critical", "Terraform state backup"},
		{`terraform\.tfvars$`, "infrastructure", "high", "Terraform variables (may contain secrets)"},

		// Ansible
		{`vault\.ya?ml$`, "infrastructure", "high", "Ansible vault file"},
		{`.*vault.*\.ya?ml$`, "infrastructure", "medium", "Potential Ansible vault file"},

		// Backup files
		{`.*\.bak$`, "backup", "medium", "Backup file"},
		{`.*\.backup$`, "backup", "medium", "Backup file"},
		{`.*\.old$`, "backup", "low", "Old file backup"},
		{`.*~$`, "backup", "info", "Editor backup file"},
	}

	s.sensitivePatterns = make([]sensitiveFilePattern, 0, len(patterns))
	for _, p := range patterns {
		regex, err := regexp.Compile("(?i)" + p.pattern)
		if err != nil {
			continue
		}
		s.sensitivePatterns = append(s.sensitivePatterns, sensitiveFilePattern{
			Pattern:     p.pattern,
			Regex:       regex,
			Category:    p.category,
			Severity:    p.severity,
			Description: p.description,
		})
	}
}

// parseGitignore parses a .gitignore file and returns rules
func (s *GitHistorySecurityScanner) parseGitignore(repoPath string) error {
	gitignorePath := filepath.Join(repoPath, ".gitignore")

	file, err := os.Open(gitignorePath)
	if err != nil {
		return err // No .gitignore file
	}
	defer file.Close()

	s.gitignoreRules = []gitignoreRule{}
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		rule := gitignoreRule{
			Pattern:    line,
			SourceLine: lineNum,
		}

		// Handle negation
		if strings.HasPrefix(line, "!") {
			rule.IsNegation = true
			line = line[1:]
		}

		// Handle directory-only patterns
		if strings.HasSuffix(line, "/") {
			rule.IsDir = true
			line = strings.TrimSuffix(line, "/")
		}

		// Convert gitignore glob to regex
		regex := gitignoreToRegex(line)
		compiled, err := regexp.Compile(regex)
		if err != nil {
			continue
		}
		rule.Regex = compiled

		s.gitignoreRules = append(s.gitignoreRules, rule)
	}

	return scanner.Err()
}

// gitignoreToRegex converts a gitignore pattern to a regex
func gitignoreToRegex(pattern string) string {
	// Escape special regex characters first
	escaped := regexp.QuoteMeta(pattern)

	// Convert gitignore wildcards to regex
	// ** matches any path
	escaped = strings.ReplaceAll(escaped, `\*\*`, ".*")
	// * matches anything except /
	escaped = strings.ReplaceAll(escaped, `\*`, "[^/]*")
	// ? matches single character except /
	escaped = strings.ReplaceAll(escaped, `\?`, "[^/]")

	// If pattern doesn't start with /, it can match anywhere
	if !strings.HasPrefix(pattern, "/") {
		escaped = "(^|/)" + escaped
	} else {
		escaped = "^" + strings.TrimPrefix(escaped, `\/`)
	}

	return escaped + "($|/)"
}

// matchesGitignore checks if a file path matches any gitignore rule
func (s *GitHistorySecurityScanner) matchesGitignore(path string) (bool, string) {
	matched := false
	matchedRule := ""

	for _, rule := range s.gitignoreRules {
		if rule.Regex.MatchString(path) {
			if rule.IsNegation {
				matched = false
				matchedRule = ""
			} else {
				matched = true
				matchedRule = rule.Pattern
			}
		}
	}

	return matched, matchedRule
}

// matchesSensitivePattern checks if a file path matches any sensitive pattern
func (s *GitHistorySecurityScanner) matchesSensitivePattern(path string) *sensitiveFilePattern {
	for _, pattern := range s.sensitivePatterns {
		if pattern.Regex.MatchString(path) {
			return &pattern
		}
	}
	return nil
}

// ScanRepository scans git history for security issues
func (s *GitHistorySecurityScanner) ScanRepository(repoPath string) (*GitHistorySecurityResult, error) {
	result := &GitHistorySecurityResult{
		GitignoreViolations:  []GitignoreViolation{},
		SensitiveFiles:       []SensitiveFileFinding{},
		PurgeRecommendations: []PurgeRecommendation{},
		Timeline:             []HistoricalEvent{},
		Summary: GitHistorySecuritySummary{
			ByCategory: make(map[string]int),
			BySeverity: make(map[string]int),
			RiskScore:  100,
			RiskLevel:  "excellent",
		},
	}

	// Parse gitignore
	if err := s.parseGitignore(repoPath); err != nil {
		// Continue without gitignore rules if file doesn't exist
	}

	// Open repository
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return result, err
	}

	// Get HEAD reference
	ref, err := repo.Head()
	if err != nil {
		return result, err
	}

	// Track files we've seen
	gitignoreViolations := make(map[string]*GitignoreViolation)
	sensitiveFiles := make(map[string]*SensitiveFileFinding)
	currentFiles := make(map[string]bool)
	commitCount := 0

	// Get current HEAD tree to check if files still exist
	headCommit, err := repo.CommitObject(ref.Hash())
	if err == nil {
		headTree, err := headCommit.Tree()
		if err == nil {
			headTree.Files().ForEach(func(f *object.File) error {
				currentFiles[f.Name] = true
				return nil
			})
		}
	}

	// Calculate since date
	since := s.parseSinceDate()

	// Get commit iterator
	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return result, err
	}

	// Collect commits within time range and limit
	var commits []*object.Commit
	err = commitIter.ForEach(func(c *object.Commit) error {
		if len(commits) >= s.config.MaxCommits {
			return nil
		}
		if c.Author.When.After(since) {
			commits = append(commits, c)
		}
		return nil
	})
	if err != nil {
		return result, err
	}

	// Process commits in chronological order (oldest first)
	for i := len(commits) - 1; i >= 0; i-- {
		commit := commits[i]
		commitCount++

		commitInfo := &CommitInfo{
			Hash:      commit.Hash.String(),
			ShortHash: commit.Hash.String()[:8],
			Author:    commit.Author.Name,
			Email:     commit.Author.Email,
			Date:      commit.Author.When.Format(time.RFC3339),
			Message:   firstLine(commit.Message),
		}

		// Get tree for this commit
		tree, err := commit.Tree()
		if err != nil {
			continue
		}

		// Scan all files in this commit
		tree.Files().ForEach(func(f *object.File) error {
			path := f.Name

			// Check gitignore violations
			if matched, rule := s.matchesGitignore(path); matched {
				if _, exists := gitignoreViolations[path]; !exists {
					gitignoreViolations[path] = &GitignoreViolation{
						File:          path,
						GitignoreRule: rule,
						FirstCommit:   commitInfo,
						StillExists:   currentFiles[path],
					}

					// Add timeline event
					result.Timeline = append(result.Timeline, HistoricalEvent{
						Date:        commitInfo.Date,
						EventType:   "committed",
						File:        path,
						CommitHash:  commitInfo.ShortHash,
						Author:      commitInfo.Author,
						Description: "File matching gitignore pattern was committed",
					})
				}
				gitignoreViolations[path].LastCommit = commitInfo
			}

			// Check sensitive file patterns
			if pattern := s.matchesSensitivePattern(path); pattern != nil {
				if _, exists := sensitiveFiles[path]; !exists {
					sensitiveFiles[path] = &SensitiveFileFinding{
						File:        path,
						Category:    pattern.Category,
						Severity:    pattern.Severity,
						Description: pattern.Description,
						FirstCommit: commitInfo,
						StillExists: currentFiles[path],
						SizeBytes:   f.Size,
					}

					// Add timeline event
					result.Timeline = append(result.Timeline, HistoricalEvent{
						Date:        commitInfo.Date,
						EventType:   "committed",
						File:        path,
						CommitHash:  commitInfo.ShortHash,
						Author:      commitInfo.Author,
						Description: pattern.Description,
					})
				}
				sensitiveFiles[path].LastCommit = commitInfo
			}

			return nil
		})
	}

	// Check which files were removed
	for path, violation := range gitignoreViolations {
		if !currentFiles[path] {
			violation.WasRemoved = true
		}
	}
	for path, finding := range sensitiveFiles {
		if !currentFiles[path] {
			finding.WasRemoved = true
		}
	}

	// Convert maps to slices
	for _, v := range gitignoreViolations {
		result.GitignoreViolations = append(result.GitignoreViolations, *v)
	}
	for _, f := range sensitiveFiles {
		result.SensitiveFiles = append(result.SensitiveFiles, *f)
	}

	// Generate purge recommendations
	result.PurgeRecommendations = s.generatePurgeRecommendations(result)

	// Sort timeline by date
	sort.Slice(result.Timeline, func(i, j int) bool {
		return result.Timeline[i].Date < result.Timeline[j].Date
	})

	// Calculate summary
	result.Summary.CommitsScanned = commitCount
	result.Summary.GitignoreViolations = len(result.GitignoreViolations)
	result.Summary.SensitiveFilesFound = len(result.SensitiveFiles)
	result.Summary.TotalViolations = result.Summary.GitignoreViolations + result.Summary.SensitiveFilesFound
	result.Summary.FilesToPurge = len(result.PurgeRecommendations)

	for _, f := range result.SensitiveFiles {
		result.Summary.ByCategory[f.Category]++
		result.Summary.BySeverity[f.Severity]++
	}

	// Calculate risk score
	penalty := result.Summary.BySeverity["critical"]*25 +
		result.Summary.BySeverity["high"]*15 +
		result.Summary.BySeverity["medium"]*5 +
		result.Summary.BySeverity["low"]*2

	result.Summary.RiskScore = 100 - penalty
	if result.Summary.RiskScore < 0 {
		result.Summary.RiskScore = 0
	}

	switch {
	case result.Summary.RiskScore < 40:
		result.Summary.RiskLevel = "critical"
	case result.Summary.RiskScore < 60:
		result.Summary.RiskLevel = "high"
	case result.Summary.RiskScore < 80:
		result.Summary.RiskLevel = "medium"
	case result.Summary.RiskScore < 95:
		result.Summary.RiskLevel = "low"
	default:
		result.Summary.RiskLevel = "excellent"
	}

	return result, nil
}

// generatePurgeRecommendations creates purge recommendations based on findings
func (s *GitHistorySecurityScanner) generatePurgeRecommendations(result *GitHistorySecurityResult) []PurgeRecommendation {
	var recommendations []PurgeRecommendation
	seen := make(map[string]bool)

	// Prioritize by severity
	severityPriority := map[string]int{
		"critical": 1,
		"high":     2,
		"medium":   3,
		"low":      4,
		"info":     5,
	}

	// Add recommendations for sensitive files
	for _, f := range result.SensitiveFiles {
		if seen[f.File] {
			continue
		}
		seen[f.File] = true

		priority := severityPriority[f.Severity]
		if priority == 0 {
			priority = 5
		}

		// Only recommend purge for critical/high severity or files that have been removed
		// (if they're still in the repo, just removing from history might not make sense)
		if f.Severity != "critical" && f.Severity != "high" && !f.WasRemoved {
			continue
		}

		recommendations = append(recommendations, PurgeRecommendation{
			File:            f.File,
			Reason:          f.Description,
			Severity:        f.Severity,
			Priority:        priority,
			Command:         generateBFGCommand(f.File),
			Alternative:     generateFilterRepoCommand(f.File),
			AffectedCommits: 1, // Would need more work to count actual affected commits
		})
	}

	// Add recommendations for gitignore violations (critical files only)
	for _, v := range result.GitignoreViolations {
		if seen[v.File] {
			continue
		}
		seen[v.File] = true

		// Check if this file pattern matches any critical sensitive patterns
		pattern := s.matchesSensitivePattern(v.File)
		if pattern == nil || (pattern.Severity != "critical" && pattern.Severity != "high") {
			continue
		}

		recommendations = append(recommendations, PurgeRecommendation{
			File:            v.File,
			Reason:          "File matches gitignore rule: " + v.GitignoreRule,
			Severity:        pattern.Severity,
			Priority:        severityPriority[pattern.Severity],
			Command:         generateBFGCommand(v.File),
			Alternative:     generateFilterRepoCommand(v.File),
			AffectedCommits: 1,
		})
	}

	// Sort by priority
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Priority < recommendations[j].Priority
	})

	return recommendations
}

// generateBFGCommand generates a BFG Repo-Cleaner command
func generateBFGCommand(file string) string {
	// Escape the file path for shell
	escaped := strings.ReplaceAll(file, "'", "'\\''")
	return "bfg --delete-files '" + escaped + "'"
}

// generateFilterRepoCommand generates a git-filter-repo command
func generateFilterRepoCommand(file string) string {
	// Escape the file path for shell
	escaped := strings.ReplaceAll(file, "'", "'\\''")
	return "git filter-repo --path '" + escaped + "' --invert-paths"
}

// parseSinceDate parses the MaxAge config into a time.Time
func (s *GitHistorySecurityScanner) parseSinceDate() time.Time {
	now := time.Now()
	maxAge := s.config.MaxAge
	if maxAge == "" {
		maxAge = "1y"
	}

	// Parse duration like "90d", "1y", "6m"
	if len(maxAge) < 2 {
		return now.AddDate(-1, 0, 0) // Default 1 year
	}

	unit := maxAge[len(maxAge)-1]
	value := 0
	for _, c := range maxAge[:len(maxAge)-1] {
		if c >= '0' && c <= '9' {
			value = value*10 + int(c-'0')
		}
	}

	switch unit {
	case 'd':
		return now.AddDate(0, 0, -value)
	case 'w':
		return now.AddDate(0, 0, -value*7)
	case 'm':
		return now.AddDate(0, -value, 0)
	case 'y':
		return now.AddDate(-value, 0, 0)
	default:
		return now.AddDate(-1, 0, 0) // Default 1 year
	}
}
