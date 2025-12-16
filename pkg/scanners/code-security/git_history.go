package codesecurity

import (
	"regexp"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// GitHistoryScanner scans git history for secrets
type GitHistoryScanner struct {
	config   GitHistoryConfig
	patterns []*secretPattern
}

// secretPattern defines a pattern for detecting a specific secret type
type secretPattern struct {
	name        string
	pattern     *regexp.Regexp
	severity    string
	description string
	technology  string
}

// NewGitHistoryScanner creates a new git history scanner
func NewGitHistoryScanner(config GitHistoryConfig) *GitHistoryScanner {
	scanner := &GitHistoryScanner{config: config}
	scanner.loadPatterns()
	return scanner
}

// loadPatterns loads secret patterns from RAG files
func (s *GitHistoryScanner) loadPatterns() {
	ragPatterns, err := LoadRAGSecretPatterns()
	if err != nil {
		// Log error but continue with empty patterns
		s.patterns = []*secretPattern{}
		return
	}

	s.patterns = make([]*secretPattern, 0, len(ragPatterns))
	for _, rp := range ragPatterns {
		s.patterns = append(s.patterns, &secretPattern{
			name:        rp.Name,
			pattern:     rp.Pattern,
			severity:    rp.Severity,
			description: rp.Description,
			technology:  rp.Technology,
		})
	}
}

// GitHistoryResult holds results from git history scanning
type GitHistoryResult struct {
	Findings       []SecretFinding
	CommitsScanned int
	SecretsFound   int
	SecretsRemoved int
}

// ScanRepository scans git history for secrets
func (s *GitHistoryScanner) ScanRepository(repoPath string) (*GitHistoryResult, error) {
	result := &GitHistoryResult{}

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

	// Calculate since date based on MaxAge
	since := s.parseSinceDate()

	// Get commit iterator
	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return result, err
	}

	// Track secrets by file+line to check if removed
	secretLocations := make(map[string]*SecretFinding)
	commitCount := 0

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

		// Get changes in this commit
		var parentTree *object.Tree
		if commit.NumParents() > 0 {
			parent, err := commit.Parent(0)
			if err == nil {
				parentTree, _ = parent.Tree()
			}
		}

		commitTree, err := commit.Tree()
		if err != nil {
			continue
		}

		// If no parent (initial commit), scan all files in tree
		if parentTree == nil {
			commitTree.Files().ForEach(func(f *object.File) error {
				s.scanFileContent(f, commitInfo, &secretLocations, result)
				return nil
			})
			continue
		}

		// Get diff between parent and current commit
		changes, err := parentTree.Diff(commitTree)
		if err != nil {
			continue
		}

		for _, change := range changes {
			// Handle file additions and modifications
			if change.To.Name != "" {
				file, err := commitTree.File(change.To.Name)
				if err != nil {
					continue
				}
				s.scanFileContent(file, commitInfo, &secretLocations, result)
			}

			// Handle deletions - mark secrets as removed
			if change.From.Name != "" && change.To.Name == "" {
				s.markSecretsRemoved(change.From.Name, &secretLocations, result)
			}
		}
	}

	result.CommitsScanned = commitCount

	// If ScanRemoved is enabled, check which secrets still exist
	if s.config.ScanRemoved {
		s.checkRemovedSecrets(&secretLocations, result)
	}

	return result, nil
}

// scanFileContent scans a file's content for secrets
func (s *GitHistoryScanner) scanFileContent(file *object.File, commitInfo *CommitInfo, locations *map[string]*SecretFinding, result *GitHistoryResult) {
	// Skip binary files
	isBinary, _ := file.IsBinary()
	if isBinary {
		return
	}

	// Skip large files
	if file.Size > 1024*1024 { // 1MB limit
		return
	}

	content, err := file.Contents()
	if err != nil {
		return
	}

	lines := strings.Split(content, "\n")
	for lineNum, line := range lines {
		for _, pattern := range s.patterns {
			if pattern.pattern.MatchString(line) {
				// Check for false positives
				if s.isFalsePositive(line, file.Name) {
					continue
				}

				locationKey := file.Name + ":" + string(rune(lineNum+1))

				// If we haven't seen this location before, add it
				if _, exists := (*locations)[locationKey]; !exists {
					match := pattern.pattern.FindString(line)
					finding := SecretFinding{
						RuleID:          "git-history-" + pattern.name,
						Type:            pattern.name,
						Severity:        pattern.severity,
						Message:         "Secret found in git history",
						File:            file.Name,
						Line:            lineNum + 1,
						Snippet:         redactHistorySecret(match),
						DetectionSource: "git_history",
						CommitInfo:      commitInfo,
					}
					(*locations)[locationKey] = &finding
					result.Findings = append(result.Findings, finding)
					result.SecretsFound++
				}
			}
		}
	}
}

// markSecretsRemoved marks secrets in a deleted file as removed
func (s *GitHistoryScanner) markSecretsRemoved(filename string, locations *map[string]*SecretFinding, result *GitHistoryResult) {
	for key, finding := range *locations {
		if strings.HasPrefix(key, filename+":") && finding.CommitInfo != nil && !finding.CommitInfo.IsRemoved {
			finding.CommitInfo.IsRemoved = true
			result.SecretsRemoved++
		}
	}
}

// checkRemovedSecrets checks if secrets were later removed from the repo
func (s *GitHistoryScanner) checkRemovedSecrets(locations *map[string]*SecretFinding, result *GitHistoryResult) {
	// Update the findings with IsRemoved status
	for i := range result.Findings {
		key := result.Findings[i].File + ":" + string(rune(result.Findings[i].Line))
		if loc, exists := (*locations)[key]; exists && loc.CommitInfo != nil {
			if result.Findings[i].CommitInfo != nil {
				result.Findings[i].CommitInfo.IsRemoved = loc.CommitInfo.IsRemoved
			}
		}
	}
}

// parseSinceDate parses the MaxAge config into a time.Time
func (s *GitHistoryScanner) parseSinceDate() time.Time {
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

// isFalsePositive checks if a match is a known false positive
func (s *GitHistoryScanner) isFalsePositive(line, filename string) bool {
	lineLower := strings.ToLower(line)
	filenameLower := strings.ToLower(filename)

	// Test file indicators
	testIndicators := []string{
		"_test.", "test_", "/test/", "/tests/",
		"mock", "fixture", "spec.", "/fixtures/",
		"example", "sample", "demo",
	}
	for _, indicator := range testIndicators {
		if strings.Contains(filenameLower, indicator) {
			return true
		}
	}

	// Placeholder patterns
	placeholders := []string{
		"example", "test", "sample", "demo", "placeholder",
		"your_", "your-", "xxx", "changeme", "replace",
		"insert", "enter_", "put_your", "dummy",
		"akiaiosfodnn7example", // AWS example key
		"sk_test_", "pk_test_", // Stripe test keys
	}
	for _, p := range placeholders {
		if strings.Contains(lineLower, p) {
			return true
		}
	}

	// Comment lines (rough heuristic)
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") ||
		strings.HasPrefix(trimmed, "*") || strings.HasPrefix(trimmed, "<!--") {
		// Check if it's documenting a secret format rather than containing one
		if strings.Contains(lineLower, "format") || strings.Contains(lineLower, "example") ||
			strings.Contains(lineLower, "like") || strings.Contains(lineLower, "e.g.") {
			return true
		}
	}

	// Documentation files
	docExtensions := []string{".md", ".rst", ".txt", ".adoc"}
	for _, ext := range docExtensions {
		if strings.HasSuffix(filenameLower, ext) {
			return true
		}
	}

	return false
}

// redactHistorySecret redacts the middle portion of a secret in git history
func redactHistorySecret(value string) string {
	if len(value) <= 8 {
		return "****"
	}
	return value[:4] + "****" + value[len(value)-4:]
}

// firstLine returns the first line of a multi-line string
func firstLine(s string) string {
	if idx := strings.Index(s, "\n"); idx != -1 {
		return strings.TrimSpace(s[:idx])
	}
	return strings.TrimSpace(s)
}
