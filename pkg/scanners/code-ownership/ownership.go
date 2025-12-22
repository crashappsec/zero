// Package codeownership provides the code ownership and CODEOWNERS analysis super scanner
package codeownership

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	gitobj "github.com/go-git/go-git/v5/plumbing/object"

	"github.com/crashappsec/zero/pkg/external/github"
	"github.com/crashappsec/zero/pkg/analysis/languages"
	"github.com/crashappsec/zero/pkg/scanner"
)

const (
	Name        = "code-ownership"
	Description = "Code ownership and CODEOWNERS analysis"
)

// OwnershipScanner implements the code ownership super scanner
type OwnershipScanner struct {
	config FeatureConfig
}

// init registers the scanner
func init() {
	scanner.Register(&OwnershipScanner{
		config: DefaultConfig(),
	})
}

// Name returns the scanner name
func (s *OwnershipScanner) Name() string {
	return Name
}

// Description returns the scanner description
func (s *OwnershipScanner) Description() string {
	return Description
}

// Dependencies returns scanner dependencies (none for ownership scanner)
func (s *OwnershipScanner) Dependencies() []string {
	return []string{}
}

// EstimateDuration returns estimated scan duration based on repo size
func (s *OwnershipScanner) EstimateDuration(fileCount int) time.Duration {
	// Git log analysis can be slow for large repos
	base := 10 * time.Second
	perFile := 5 * time.Millisecond
	return base + time.Duration(fileCount)*perFile
}

// Run executes the ownership analysis
func (s *OwnershipScanner) Run(ctx context.Context, opts *scanner.ScanOptions) (*scanner.ScanResult, error) {
	startTime := time.Now()

	// Use default config
	cfg := s.config

	result := &Result{
		FeaturesRun: []string{},
		Summary:     Summary{},
		Findings:    Findings{},
	}

	if !cfg.Enabled {
		scanResult := scanner.NewScanResult(Name, "1.0.0", startTime)
		if err := scanResult.SetSummary(result.Summary); err != nil {
			return nil, fmt.Errorf("failed to set summary: %w", err)
		}
		if err := scanResult.SetFindings(result.Findings); err != nil {
			return nil, fmt.Errorf("failed to set findings: %w", err)
		}
		return scanResult, nil
	}

	// Use enhanced v2.0 analysis if enabled
	if cfg.EnhancedMode {
		return s.runEnhancedMode(ctx, opts, startTime)
	}

	result.FeaturesRun = append(result.FeaturesRun, "ownership")

	// Get languages (from cache or detect)
	var langStats *languages.DirectoryStats
	if cfg.DetectLanguages {
		result.FeaturesRun = append(result.FeaturesRun, "languages")
		langStats = s.getLanguageStats(opts)
	}

	// Open git repository
	repo, err := git.PlainOpen(opts.RepoPath)
	if err != nil {
		result.Summary.Errors = append(result.Summary.Errors, fmt.Sprintf("opening repository: %v", err))
		scanResult := scanner.NewScanResult(Name, "1.0.0", startTime)
		scanResult.SetSummary(result.Summary)
		scanResult.SetFindings(result.Findings)
		return scanResult, nil
	}

	// Check for shallow clone
	if isShallow := s.isShallowClone(opts.RepoPath); isShallow {
		result.Summary.IsShallowClone = true
		result.Summary.Warnings = append(result.Summary.Warnings,
			"Repository is a shallow clone. Contributor and competency analysis will be limited.")
	}

	periodDays := cfg.PeriodDays
	if periodDays <= 0 {
		periodDays = 90
	}

	now := time.Now()
	since := now.AddDate(0, 0, -periodDays)

	var fileOwners map[string][]string
	var contributors map[string]Contributor
	var orphanedFiles []string
	var devProfiles map[string]*DeveloperProfile

	// Analyze contributors, file ownership, and competency
	if cfg.AnalyzeContributors || cfg.DetectOrphans || cfg.AnalyzeCompetency {
		if cfg.AnalyzeCompetency {
			result.FeaturesRun = append(result.FeaturesRun, "competency")
			fileOwners, contributors, devProfiles = s.analyzeOwnershipWithCompetency(repo, since)
		} else {
			fileOwners, contributors = s.analyzeOwnership(repo, since)
		}
	}

	// Detect orphaned files
	if cfg.DetectOrphans && fileOwners != nil {
		for file, owners := range fileOwners {
			if len(owners) == 0 {
				orphanedFiles = append(orphanedFiles, file)
			}
		}
	}

	// Parse CODEOWNERS
	var codeowners []CodeownerRule
	if cfg.CheckCodeowners {
		codeowners = s.parseCodeowners(opts.RepoPath)
	}

	// Build contributor list
	var contribList []Contributor
	for email, c := range contributors {
		c.Email = email
		contribList = append(contribList, c)
	}
	sort.Slice(contribList, func(i, j int) bool {
		return contribList[i].Commits > contribList[j].Commits
	})
	if len(contribList) > 20 {
		contribList = contribList[:20]
	}

	// Build file ownership list
	var fileOwnershipList []FileOwnership
	for file, owners := range fileOwners {
		if len(owners) > 0 {
			topOwners := owners
			if len(topOwners) > 3 {
				topOwners = topOwners[:3]
			}
			fileOwnershipList = append(fileOwnershipList, FileOwnership{
				Path:            file,
				TopContributors: topOwners,
				CommitCount:     len(owners),
			})
		}
	}

	// Build competency list
	var competencyList []DeveloperProfile
	if devProfiles != nil {
		for _, profile := range devProfiles {
			// Calculate competency score and finalize profile
			s.finalizeProfile(profile)
			competencyList = append(competencyList, *profile)
		}
		// Sort by competency score descending
		sort.Slice(competencyList, func(i, j int) bool {
			return competencyList[i].CompetencyScore > competencyList[j].CompetencyScore
		})
		// Limit to top 30 developers
		if len(competencyList) > 30 {
			competencyList = competencyList[:30]
		}
	}

	// Update summary (preserve IsShallowClone and Warnings set earlier)
	result.Summary.TotalContributors = len(contributors)
	result.Summary.FilesAnalyzed = len(fileOwners)
	result.Summary.HasCodeowners = len(codeowners) > 0
	result.Summary.CodeownersRules = len(codeowners)
	result.Summary.OrphanedFiles = len(orphanedFiles)
	result.Summary.PeriodDays = periodDays

	// Add language stats to summary
	if langStats != nil {
		result.Summary.LanguagesDetected = langStats.LanguageCount
		topLangs := languages.TopLanguages(langStats, 5)
		for _, ls := range topLangs {
			result.Summary.TopLanguages = append(result.Summary.TopLanguages, LanguageInfo{
				Name:       ls.Language,
				FileCount:  ls.FileCount,
				Percentage: ls.Percentage,
			})
		}
	}

	result.Findings = Findings{
		Contributors:  contribList,
		Codeowners:    codeowners,
		OrphanedFiles: orphanedFiles,
		FileOwners:    fileOwnershipList,
		Competencies:  competencyList,
	}

	// Create scan result
	scanResult := scanner.NewScanResult(Name, "1.0.0", startTime)
	scanResult.Repository = opts.RepoPath
	if err := scanResult.SetSummary(result.Summary); err != nil {
		return nil, fmt.Errorf("failed to set summary: %w", err)
	}
	if err := scanResult.SetFindings(result.Findings); err != nil {
		return nil, fmt.Errorf("failed to set findings: %w", err)
	}

	// Add metadata
	metadata := map[string]any{
		"features_run": result.FeaturesRun,
		"period_days":  periodDays,
	}
	if err := scanResult.SetMetadata(metadata); err != nil {
		return nil, fmt.Errorf("failed to set metadata: %w", err)
	}

	// Write output
	if opts.OutputDir != "" {
		if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
			return nil, fmt.Errorf("creating output dir: %w", err)
		}
		resultFile := filepath.Join(opts.OutputDir, Name+".json")
		if err := scanResult.WriteJSON(resultFile); err != nil {
			return nil, fmt.Errorf("writing result: %w", err)
		}

	}

	return scanResult, nil
}

// analyzeOwnership analyzes git history to determine file ownership
func (s *OwnershipScanner) analyzeOwnership(repo *git.Repository, since time.Time) (map[string][]string, map[string]Contributor) {
	fileOwners := make(map[string][]string)
	contributors := make(map[string]Contributor)
	fileContribs := make(map[string]map[string]bool)

	ref, err := repo.Head()
	if err != nil {
		return fileOwners, contributors
	}

	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return fileOwners, contributors
	}

	var commits []*gitobj.Commit
	commitIter.ForEach(func(c *gitobj.Commit) error {
		if c.Author.When.After(since) {
			commits = append(commits, c)
		}
		return nil
	})

	for i, c := range commits {
		email := c.Author.Email
		contrib := contributors[email]
		contrib.Name = c.Author.Name
		contrib.Commits++

		if i < len(commits)-1 {
			parent := commits[i+1]
			files := s.getChangedFiles(c, parent)
			for _, f := range files {
				if fileContribs[f] == nil {
					fileContribs[f] = make(map[string]bool)
				}
				fileContribs[f][email] = true
			}
			contrib.FilesTouched += len(files)
		}

		contributors[email] = contrib
	}

	for file, contribs := range fileContribs {
		for email := range contribs {
			fileOwners[file] = append(fileOwners[file], email)
		}
	}

	return fileOwners, contributors
}

// getChangedFiles returns files changed between two commits
func (s *OwnershipScanner) getChangedFiles(commit, parent *gitobj.Commit) []string {
	var files []string

	commitTree, err := commit.Tree()
	if err != nil {
		return files
	}
	parentTree, err := parent.Tree()
	if err != nil {
		return files
	}

	changes, err := parentTree.Diff(commitTree)
	if err != nil {
		return files
	}

	for _, change := range changes {
		name := change.To.Name
		if name == "" {
			name = change.From.Name
		}
		if name != "" {
			files = append(files, name)
		}
	}

	return files
}

// analyzeOwnershipWithCompetency analyzes git history with per-language competency tracking
func (s *OwnershipScanner) analyzeOwnershipWithCompetency(repo *git.Repository, since time.Time) (map[string][]string, map[string]Contributor, map[string]*DeveloperProfile) {
	fileOwners := make(map[string][]string)
	contributors := make(map[string]Contributor)
	fileContribs := make(map[string]map[string]bool)
	devProfiles := make(map[string]*DeveloperProfile)
	// Track unique files per developer per language to avoid counting the same file multiple times
	devLangFiles := make(map[string]map[string]map[string]bool) // email -> language -> file -> seen

	ref, err := repo.Head()
	if err != nil {
		return fileOwners, contributors, devProfiles
	}

	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return fileOwners, contributors, devProfiles
	}

	var commits []*gitobj.Commit
	commitIter.ForEach(func(c *gitobj.Commit) error {
		if c.Author.When.After(since) {
			commits = append(commits, c)
		}
		return nil
	})

	for i, c := range commits {
		email := c.Author.Email
		contrib := contributors[email]
		contrib.Name = c.Author.Name
		contrib.Commits++

		// Track commit dates for recency and consistency scoring
		commitTime := c.Author.When
		contrib.CommitDates = append(contrib.CommitDates, commitTime)
		if contrib.LastCommit.IsZero() || commitTime.After(contrib.LastCommit) {
			contrib.LastCommit = commitTime
		}

		// Get or create developer profile
		profile := devProfiles[email]
		if profile == nil {
			profile = &DeveloperProfile{
				Name:      c.Author.Name,
				Email:     email,
				Languages: []LanguageStats{},
			}
			devProfiles[email] = profile
		}
		profile.TotalCommits++

		// Classify commit type based on message
		commitType := classifyCommitType(c.Message)
		switch commitType {
		case "feature":
			profile.FeatureCommits++
		case "bugfix":
			profile.BugFixCommits++
		case "refactor":
			profile.RefactorCommits++
		default:
			profile.OtherCommits++
		}

		if i < len(commits)-1 {
			parent := commits[i+1]
			files := s.getChangedFiles(c, parent)
			for _, f := range files {
				if fileContribs[f] == nil {
					fileContribs[f] = make(map[string]bool)
				}
				fileContribs[f][email] = true

				// Track language stats for developer
				lang := languages.DetectFromPath(f)
				if lang != "" && languages.IsProgrammingLanguage(lang) {
					// Initialize tracking maps if needed
					if devLangFiles[email] == nil {
						devLangFiles[email] = make(map[string]map[string]bool)
					}
					if devLangFiles[email][lang] == nil {
						devLangFiles[email][lang] = make(map[string]bool)
					}
					// Check if this is a unique file for this developer+language
					isNewFile := !devLangFiles[email][lang][f]
					if isNewFile {
						devLangFiles[email][lang][f] = true
					}
					s.updateLanguageStats(profile, lang, commitType, isNewFile)
				}
			}
			contrib.FilesTouched += len(files)
		}

		contributors[email] = contrib
	}

	for file, contribs := range fileContribs {
		for email := range contribs {
			fileOwners[file] = append(fileOwners[file], email)
		}
	}

	return fileOwners, contributors, devProfiles
}

// updateLanguageStats updates a developer's per-language statistics
// isNewFile should be true only if this is the first time the developer touched this file in this language
func (s *OwnershipScanner) updateLanguageStats(profile *DeveloperProfile, lang string, commitType string, isNewFile bool) {
	// Find existing language stats or create new
	var langStats *LanguageStats
	for i := range profile.Languages {
		if profile.Languages[i].Language == lang {
			langStats = &profile.Languages[i]
			break
		}
	}

	if langStats == nil {
		profile.Languages = append(profile.Languages, LanguageStats{Language: lang})
		langStats = &profile.Languages[len(profile.Languages)-1]
	}

	langStats.Commits++
	// Only increment FileCount for unique files (not per-commit)
	if isNewFile {
		langStats.FileCount++
	}

	switch commitType {
	case "feature":
		langStats.FeatureCommits++
	case "bugfix":
		langStats.BugFixCommits++
	}
}

// finalizeProfile calculates final metrics for a developer profile
func (s *OwnershipScanner) finalizeProfile(profile *DeveloperProfile) {
	if profile.TotalCommits == 0 {
		return
	}

	// Sort languages by commit count
	sort.Slice(profile.Languages, func(i, j int) bool {
		return profile.Languages[i].Commits > profile.Languages[j].Commits
	})

	// Set top language
	if len(profile.Languages) > 0 {
		profile.TopLanguage = profile.Languages[0].Language
	}

	// Calculate percentages for each language
	for i := range profile.Languages {
		profile.Languages[i].Percentage = float64(profile.Languages[i].Commits) / float64(profile.TotalCommits) * 100
	}

	// Limit to top 10 languages
	if len(profile.Languages) > 10 {
		profile.Languages = profile.Languages[:10]
	}

	// Calculate competency score
	// Factors: total commits, language breadth, bug fix ratio
	bugFixRatio := float64(profile.BugFixCommits) / float64(profile.TotalCommits)
	languageBreadth := float64(len(profile.Languages))
	commitVolume := float64(profile.TotalCommits)

	// Score formula: commits * (1 + bug_fix_bonus) * language_bonus
	// Bug fix work is weighted higher (fixing bugs shows deeper understanding)
	bugFixBonus := bugFixRatio * 0.5               // Up to 50% bonus for bug fixes
	languageBonus := 1.0 + (languageBreadth-1)*0.1 // 10% bonus per additional language

	profile.CompetencyScore = commitVolume * (1 + bugFixBonus) * languageBonus
}

// classifyCommitType analyzes commit message to determine type
func classifyCommitType(message string) string {
	msg := strings.ToLower(message)

	// Bug fix patterns
	bugPatterns := regexp.MustCompile(`(?i)(fix|bug|issue|patch|hotfix|resolve|closes? #|fixes? #)`)
	if bugPatterns.MatchString(msg) {
		return "bugfix"
	}

	// Refactor patterns
	refactorPatterns := regexp.MustCompile(`(?i)(refactor|cleanup|clean up|reorganize|restructure|simplify|optimize)`)
	if refactorPatterns.MatchString(msg) {
		return "refactor"
	}

	// Feature patterns (default for add/implement/create)
	featurePatterns := regexp.MustCompile(`(?i)(feat|feature|add|implement|create|new|introduce|support)`)
	if featurePatterns.MatchString(msg) {
		return "feature"
	}

	return "other"
}

// parseCodeowners parses the CODEOWNERS file
func (s *OwnershipScanner) parseCodeowners(repoPath string) []CodeownerRule {
	var rules []CodeownerRule

	paths := []string{
		filepath.Join(repoPath, "CODEOWNERS"),
		filepath.Join(repoPath, ".github", "CODEOWNERS"),
		filepath.Join(repoPath, "docs", "CODEOWNERS"),
	}

	var content []byte
	for _, p := range paths {
		if data, err := os.ReadFile(p); err == nil {
			content = data
			break
		}
	}

	if content == nil {
		return rules
	}

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 {
			rules = append(rules, CodeownerRule{
				Pattern: parts[0],
				Owners:  parts[1:],
			})
		}
	}

	return rules
}

// isShallowClone checks if the repository is a shallow clone
// by looking for the .git/shallow file
func (s *OwnershipScanner) isShallowClone(repoPath string) bool {
	shallowFile := filepath.Join(repoPath, ".git", "shallow")
	if _, err := os.Stat(shallowFile); err == nil {
		return true
	}
	return false
}

// getLanguageStats reads cached language data from the analysis directory,
// falling back to scanning if not available
func (s *OwnershipScanner) getLanguageStats(opts *scanner.ScanOptions) *languages.DirectoryStats {
	// Try to read cached languages.json from output directory
	if opts.OutputDir != "" {
		langFile := filepath.Join(opts.OutputDir, "languages.json")
		if data, err := os.ReadFile(langFile); err == nil {
			var stats languages.DirectoryStats
			if err := json.Unmarshal(data, &stats); err == nil {
				return &stats
			}
		}
	}

	// Fallback: run language detection ourselves
	langOpts := languages.DefaultScanOptions()
	langOpts.OnlyProgramming = true
	stats, err := languages.ScanDirectory(opts.RepoPath, langOpts)
	if err != nil {
		return nil
	}
	return stats
}

// ============================================================================
// Historical Stats Collection
// ============================================================================

// historicalStats holds full-history repository statistics
type historicalStats struct {
	TotalCommits         int
	UniqueContributors   int
	LastCommitDate       string
	LastCommitTime       time.Time
	DaysSinceLastCommit  int
	ActivityStatus       string
}

// collectHistoricalStats gathers full-history repository statistics
func (s *OwnershipScanner) collectHistoricalStats(repo *git.Repository) historicalStats {
	stats := historicalStats{}

	ref, err := repo.Head()
	if err != nil {
		return stats
	}

	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return stats
	}

	contributors := make(map[string]bool)
	var lastCommitTime time.Time
	commitCount := 0

	commitIter.ForEach(func(c *gitobj.Commit) error {
		commitCount++
		contributors[c.Author.Email] = true

		// Track most recent commit
		if lastCommitTime.IsZero() || c.Author.When.After(lastCommitTime) {
			lastCommitTime = c.Author.When
		}
		return nil
	})

	stats.TotalCommits = commitCount
	stats.UniqueContributors = len(contributors)

	if !lastCommitTime.IsZero() {
		stats.LastCommitTime = lastCommitTime
		stats.LastCommitDate = lastCommitTime.Format("2006-01-02")
		stats.DaysSinceLastCommit = int(time.Since(lastCommitTime).Hours() / 24)

		// Determine activity status
		switch {
		case stats.DaysSinceLastCommit <= ActivityThresholds.Active:
			stats.ActivityStatus = "active"
		case stats.DaysSinceLastCommit <= ActivityThresholds.Recent:
			stats.ActivityStatus = "recent"
		case stats.DaysSinceLastCommit <= ActivityThresholds.Stale:
			stats.ActivityStatus = "stale"
		case stats.DaysSinceLastCommit <= ActivityThresholds.Inactive:
			stats.ActivityStatus = "inactive"
		default:
			stats.ActivityStatus = "abandoned"
		}
	}

	return stats
}

// ============================================================================
// Enhanced Ownership Analysis (v2.0)
// ============================================================================

// RunEnhancedAnalysis performs enhanced ownership analysis with all v2.0 features
func (s *OwnershipScanner) RunEnhancedAnalysis(
	ctx context.Context,
	opts *scanner.ScanOptions,
	enhancedCfg EnhancedOwnershipConfig,
) (*Findings, *Summary, error) {
	findings := &Findings{}
	summary := &Summary{}

	// Check for GitHub token
	ghClient := github.NewOwnershipClient(enhancedCfg.GitHub.MaxPRs)
	summary.GitHubTokenPresent = ghClient.HasToken()

	if !summary.GitHubTokenPresent && enhancedCfg.GitHub.Enabled {
		summary.Warnings = append(summary.Warnings, GitHubTokenMessage)
	}

	// Get basic contributor data first
	repo, err := git.PlainOpen(opts.RepoPath)
	if err != nil {
		return nil, nil, fmt.Errorf("opening repository: %w", err)
	}

	// Collect historical repo stats first (full history)
	histStats := s.collectHistoricalStats(repo)
	summary.TotalCommits = histStats.TotalCommits
	summary.AllTimeContributors = histStats.UniqueContributors
	summary.LastCommitDate = histStats.LastCommitDate
	summary.DaysSinceLastCommit = histStats.DaysSinceLastCommit
	summary.RepoActivityStatus = histStats.ActivityStatus

	// Use adaptive period detection: if no recent activity, extend the analysis window
	periodDays := 90
	now := time.Now()

	// Check if there are commits within the default 90-day window
	if histStats.DaysSinceLastCommit > periodDays {
		// No recent commits - extend to cover the last active period
		// Use a graduated approach: extend to cover at least some commits
		if histStats.DaysSinceLastCommit <= 180 {
			periodDays = 180 // Extend to 6 months
		} else if histStats.DaysSinceLastCommit <= 365 {
			periodDays = 365 // Extend to 1 year
		} else {
			periodDays = histStats.DaysSinceLastCommit + 30 // Extend to cover all recent activity + buffer
		}
		summary.AnalysisPeriodAdjusted = true
		summary.Warnings = append(summary.Warnings,
			fmt.Sprintf("No commits in last 90 days. Analysis period extended to %d days to capture recent activity.", periodDays))
	}

	since := now.AddDate(0, 0, -periodDays)
	summary.PeriodDays = periodDays

	// Analyze with competency using adaptive period
	fileOwners, contributors, _ := s.analyzeOwnershipWithCompetency(repo, since)

	// Set period-specific contributor count
	summary.TotalContributors = len(contributors)
	summary.FilesAnalyzed = len(fileOwners)

	// Convert to contributor data for scoring
	var contribData []ContributorData
	for email, contrib := range contributors {
		cd := ContributorData{
			Name:         contrib.Name,
			Email:        email,
			Commits:      contrib.Commits,
			LinesAdded:   contrib.LinesAdded,
			LinesRemoved: contrib.LinesRemoved,
			LastCommit:   contrib.LastCommit,
			CommitDates:  contrib.CommitDates,
		}
		contribData = append(contribData, cd)
	}

	// Calculate enhanced ownership scores
	scorer := NewOwnershipScorer(enhancedCfg.Weights)
	findings.EnhancedOwnership = scorer.CalculateEnhancedOwnership(contribData, now)

	// Calculate bus factor
	summary.BusFactor, summary.BusFactorRisk = CalculateBusFactor(findings.EnhancedOwnership, 0.5)

	// Calculate ownership coverage
	var fileOwnerships []FileOwnership
	for file, owners := range fileOwners {
		fileOwnerships = append(fileOwnerships, FileOwnership{
			Path:            file,
			TopContributors: owners,
			CommitCount:     len(owners),
		})
	}
	summary.OwnershipCoverage = CalculateOwnershipCoverage(fileOwnerships, 1)

	// Analyze CODEOWNERS
	if enhancedCfg.CODEOWNERS.Validate {
		analyzer := NewCODEOWNERSAnalyzer(enhancedCfg.CODEOWNERS)
		var contribList []Contributor
		for email, c := range contributors {
			c.Email = email
			contribList = append(contribList, c)
		}
		codeownersAnalysis, err := analyzer.Analyze(opts.RepoPath, contribList)
		if err == nil {
			findings.CodeownersAnalysis = codeownersAnalysis
			summary.CodeownersIssues = len(codeownersAnalysis.ValidationIssues)
		}
	}

	// Detect monorepo
	if enhancedCfg.Monorepo.Enabled {
		detector := NewMonorepoDetector(enhancedCfg.Monorepo)
		monorepoAnalysis, err := detector.Detect(opts.RepoPath)
		if err == nil && monorepoAnalysis != nil {
			findings.Monorepo = monorepoAnalysis
			summary.IsMonorepo = monorepoAnalysis.IsMonorepo
			summary.WorkspaceCount = len(monorepoAnalysis.Workspaces)
		}
	}

	// Fetch PR review data if GitHub token available
	if summary.GitHubTokenPresent && enhancedCfg.GitHub.FetchPRReviews {
		// Extract owner/repo from path or git remote
		owner, repoName := extractRepoInfo(opts.RepoPath)
		if owner != "" && repoName != "" {
			prs, totalPRs, err := ghClient.FetchPRReviews(owner, repoName)
			if err != nil {
				summary.Warnings = append(summary.Warnings, fmt.Sprintf("PR analysis error: %v", err))
			} else if prs == nil && totalPRs > enhancedCfg.GitHub.MaxPRs {
				// Too many PRs, skipped
				summary.PRAnalysisSkipped = true
				findings.PRAnalysis = &PRAnalysis{
					Skipped:    true,
					SkipReason: "pr_count_exceeded",
					TotalPRs:   totalPRs,
					Threshold:  enhancedCfg.GitHub.MaxPRs,
				}
				summary.Warnings = append(summary.Warnings,
					fmt.Sprintf("PR analysis skipped: repository has %d PRs (threshold: %d). Increase max_prs in config to analyze.",
						totalPRs, enhancedCfg.GitHub.MaxPRs))
			} else if prs != nil {
				// Aggregate reviewer stats
				reviewerStats := github.AggregateReviewerStats(prs)
				findings.PRAnalysis = &PRAnalysis{
					PRsAnalyzed: len(prs),
					TotalPRs:    totalPRs,
				}

				// Convert to PRReviewer type
				for login, stats := range reviewerStats {
					findings.PRAnalysis.Reviewers = append(findings.PRAnalysis.Reviewers, PRReviewer{
						Name:           login,
						ReviewsGiven:   stats.ReviewsGiven,
						ApprovalsGiven: stats.Approvals,
						CommentsGiven:  stats.Comments,
					})
				}

				// Update enhanced ownership with PR review data
				for i := range findings.EnhancedOwnership {
					owner := &findings.EnhancedOwnership[i]
					// Try to match by email or name
					for login, stats := range reviewerStats {
						if strings.EqualFold(login, owner.Name) || strings.Contains(strings.ToLower(owner.Email), strings.ToLower(login)) {
							owner.PRReviewsGiven = stats.ReviewsGiven
							break
						}
					}
				}
			}
		}
	}

	// Generate incident contacts
	if enhancedCfg.Contacts.Enabled {
		contactGen := NewContactGenerator(enhancedCfg.Contacts)
		codeownersRules := s.parseCodeowners(opts.RepoPath)
		findings.IncidentContacts = contactGen.GenerateKeyPathContacts(
			opts.RepoPath,
			findings.EnhancedOwnership,
			codeownersRules,
		)
	}

	return findings, summary, nil
}

// extractRepoInfo attempts to get owner/repo from git remote
func extractRepoInfo(repoPath string) (owner, repo string) {
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		return "", ""
	}

	remotes, err := r.Remotes()
	if err != nil || len(remotes) == 0 {
		return "", ""
	}

	// Get origin remote URL
	for _, remote := range remotes {
		if remote.Config().Name == "origin" {
			urls := remote.Config().URLs
			if len(urls) > 0 {
				return parseGitURL(urls[0])
			}
		}
	}

	return "", ""
}

// parseGitURL extracts owner/repo from a git URL
func parseGitURL(url string) (owner, repo string) {
	// Handle SSH format: git@github.com:owner/repo.git
	if strings.HasPrefix(url, "git@") {
		parts := strings.Split(url, ":")
		if len(parts) == 2 {
			path := strings.TrimSuffix(parts[1], ".git")
			pathParts := strings.Split(path, "/")
			if len(pathParts) >= 2 {
				return pathParts[0], pathParts[1]
			}
		}
	}

	// Handle HTTPS format: https://github.com/owner/repo.git
	if strings.Contains(url, "github.com") {
		url = strings.TrimSuffix(url, ".git")
		parts := strings.Split(url, "/")
		if len(parts) >= 2 {
			return parts[len(parts)-2], parts[len(parts)-1]
		}
	}

	return "", ""
}

// runEnhancedMode runs the v2.0 enhanced ownership analysis
func (s *OwnershipScanner) runEnhancedMode(ctx context.Context, opts *scanner.ScanOptions, startTime time.Time) (*scanner.ScanResult, error) {
	// Get enhanced config
	enhancedCfg := DefaultEnhancedConfig()

	// Run enhanced analysis
	findings, summary, err := s.RunEnhancedAnalysis(ctx, opts, enhancedCfg)
	if err != nil {
		return nil, fmt.Errorf("enhanced analysis: %w", err)
	}

	// Get language stats (from cache or detect)
	langStats := s.getLanguageStats(opts)

	// Add language stats to summary
	if langStats != nil {
		summary.LanguagesDetected = langStats.LanguageCount
		topLangs := languages.TopLanguages(langStats, 5)
		for _, ls := range topLangs {
			summary.TopLanguages = append(summary.TopLanguages, LanguageInfo{
				Name:       ls.Language,
				FileCount:  ls.FileCount,
				Percentage: ls.Percentage,
			})
		}
	}

	// Check for shallow clone
	if isShallow := s.isShallowClone(opts.RepoPath); isShallow {
		summary.IsShallowClone = true
		// Only add warning if not already present
		hasWarning := false
		for _, w := range summary.Warnings {
			if strings.Contains(w, "shallow clone") {
				hasWarning = true
				break
			}
		}
		if !hasWarning {
			summary.Warnings = append(summary.Warnings,
				"Repository is a shallow clone. Contributor and competency analysis will be limited.")
		}
	}

	// Parse CODEOWNERS for summary
	codeowners := s.parseCodeowners(opts.RepoPath)
	summary.HasCodeowners = len(codeowners) > 0
	summary.CodeownersRules = len(codeowners)
	// Note: summary.PeriodDays is already set in RunEnhancedAnalysis (may be adaptive)

	// Store codeowners in findings if not already there
	if len(findings.Codeowners) == 0 {
		findings.Codeowners = codeowners
	}

	// Build features list
	featuresRun := []string{"ownership", "languages", "competency", "enhanced_scoring"}
	if enhancedCfg.CODEOWNERS.Validate {
		featuresRun = append(featuresRun, "codeowners_validation")
	}
	if enhancedCfg.Monorepo.Enabled {
		featuresRun = append(featuresRun, "monorepo_detection")
	}
	if enhancedCfg.Contacts.Enabled {
		featuresRun = append(featuresRun, "incident_contacts")
	}
	if enhancedCfg.GitHub.Enabled && summary.GitHubTokenPresent {
		featuresRun = append(featuresRun, "github_integration")
	}

	// Create scan result
	scanResult := scanner.NewScanResult(Name, "2.0.0", startTime)
	scanResult.Repository = opts.RepoPath
	if err := scanResult.SetSummary(summary); err != nil {
		return nil, fmt.Errorf("failed to set summary: %w", err)
	}
	if err := scanResult.SetFindings(findings); err != nil {
		return nil, fmt.Errorf("failed to set findings: %w", err)
	}

	// Add metadata
	metadata := map[string]any{
		"features_run":           featuresRun,
		"period_days":            summary.PeriodDays,
		"period_adjusted":        summary.AnalysisPeriodAdjusted,
		"enhanced_mode":          true,
		"total_commits":          summary.TotalCommits,
		"all_time_contributors":  summary.AllTimeContributors,
		"repo_activity_status":   summary.RepoActivityStatus,
		"days_since_last_commit": summary.DaysSinceLastCommit,
	}
	if err := scanResult.SetMetadata(metadata); err != nil {
		return nil, fmt.Errorf("failed to set metadata: %w", err)
	}

	// Write output
	if opts.OutputDir != "" {
		if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
			return nil, fmt.Errorf("creating output dir: %w", err)
		}
		resultFile := filepath.Join(opts.OutputDir, Name+".json")
		if err := scanResult.WriteJSON(resultFile); err != nil {
			return nil, fmt.Errorf("writing result: %w", err)
		}
	}

	return scanResult, nil
}
