// Package codeownership provides the code ownership and CODEOWNERS analysis super scanner
package codeownership

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	gitobj "github.com/go-git/go-git/v5/plumbing/object"

	"github.com/crashappsec/zero/pkg/languages"
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

	result.FeaturesRun = append(result.FeaturesRun, "ownership")

	// Detect languages in repository
	var langStats *languages.DirectoryStats
	if cfg.DetectLanguages {
		result.FeaturesRun = append(result.FeaturesRun, "languages")
		langOpts := languages.DefaultScanOptions()
		langOpts.OnlyProgramming = true
		var err error
		langStats, err = languages.ScanDirectory(opts.RepoPath, langOpts)
		if err != nil {
			result.Summary.Errors = append(result.Summary.Errors, fmt.Sprintf("scanning languages: %v", err))
		}
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

	result.Summary = Summary{
		TotalContributors: len(contributors),
		FilesAnalyzed:     len(fileOwners),
		HasCodeowners:     len(codeowners) > 0,
		CodeownersRules:   len(codeowners),
		OrphanedFiles:     len(orphanedFiles),
		PeriodDays:        periodDays,
	}

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
