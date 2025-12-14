// Package codeownership provides the code ownership and CODEOWNERS analysis super scanner
package codeownership

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	gitobj "github.com/go-git/go-git/v5/plumbing/object"

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

	// Analyze contributors and file ownership
	if cfg.AnalyzeContributors || cfg.DetectOrphans {
		fileOwners, contributors = s.analyzeOwnership(repo, since)
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

	result.Summary = Summary{
		TotalContributors: len(contributors),
		FilesAnalyzed:     len(fileOwners),
		HasCodeowners:     len(codeowners) > 0,
		CodeownersRules:   len(codeowners),
		OrphanedFiles:     len(orphanedFiles),
		PeriodDays:        periodDays,
	}

	result.Findings = Findings{
		Contributors:  contribList,
		Codeowners:    codeowners,
		OrphanedFiles: orphanedFiles,
		FileOwners:    fileOwnershipList,
	}

	// Create scan result
	scanResult := scanner.NewScanResult(Name, "1.0.0", startTime)
	if err := scanResult.SetSummary(result.Summary); err != nil {
		return nil, fmt.Errorf("failed to set summary: %w", err)
	}
	if err := scanResult.SetFindings(result.Findings); err != nil {
		return nil, fmt.Errorf("failed to set findings: %w", err)
	}

	// Add metadata
	metadata := map[string]interface{}{
		"features_run": result.FeaturesRun,
		"period_days":  periodDays,
	}
	if err := scanResult.SetMetadata(metadata); err != nil {
		return nil, fmt.Errorf("failed to set metadata: %w", err)
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
