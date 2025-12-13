// Package hydrate implements the hydrate command for cloning and scanning repos
package hydrate

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/crashappsec/zero/pkg/config"
	"github.com/crashappsec/zero/pkg/github"
	"github.com/crashappsec/zero/pkg/scanner"
	"github.com/crashappsec/zero/pkg/terminal"
)

// Options configures the hydrate command
type Options struct {
	// Target
	Org   string // GitHub organization
	Repo  string // Single repo (owner/repo)
	Limit int    // Max repos in org mode

	// Clone options
	Branch    string // Clone specific branch
	Depth     int    // Shallow clone depth
	CloneOnly bool   // Clone without scanning

	// Scan options
	Profile          string   // Scan profile
	Force            bool     // Re-scan even if exists
	SkipSlow         bool     // Skip slow scanners
	Yes              bool     // Auto-accept prompts
	ParallelRepos    int      // Parallel repo processing (default: 1)
	ParallelScanners int      // Parallel scanner execution (default: 4)
	SkipScanners     []string // Scanners to skip
}

// RepoStatus tracks the status of a repo being processed
type RepoStatus struct {
	Repo       github.Repository
	RepoPath   string
	FileCount  int
	CloneOK    bool
	ScanOK     bool
	Progress   *scanner.Progress
	Duration   time.Duration
}

// Hydrate orchestrates the clone and scan process
type Hydrate struct {
	cfg      *config.Config
	term     *terminal.Terminal
	gh       *github.Client
	runner   *scanner.NativeRunner
	opts     *Options
	zeroHome string
}

// New creates a new Hydrate instance
func New(opts *Options) (*Hydrate, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	zeroHome := cfg.ZeroHome()
	if zeroHome == "" {
		zeroHome = ".zero"
	}

	if opts.ParallelRepos == 0 {
		opts.ParallelRepos = cfg.Settings.ParallelRepos
	}
	if opts.ParallelScanners == 0 {
		opts.ParallelScanners = cfg.Settings.ParallelScanners
	}

	return &Hydrate{
		cfg:      cfg,
		term:     terminal.New(),
		gh:       github.NewClient(),
		runner:   scanner.NewNativeRunner(zeroHome),
		opts:     opts,
		zeroHome: zeroHome,
	}, nil
}

// Run executes the hydrate process and returns scanned project IDs
func (h *Hydrate) Run(ctx context.Context) ([]string, error) {
	start := time.Now()
	scanID := fmt.Sprintf("scan-%s", time.Now().Format("20060102-150405"))

	var repos []github.Repository
	var targetName string

	// Single repo or org mode
	if h.opts.Repo != "" {
		// Single repo mode
		targetName = h.opts.Repo
		h.term.Info("Hydrating %s...", h.term.Color(terminal.Cyan, h.opts.Repo))

		// Parse owner/repo
		parts := strings.Split(h.opts.Repo, "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid repo format: use owner/repo")
		}

		repos = []github.Repository{{
			Name:          parts[1],
			NameWithOwner: h.opts.Repo,
			Owner:         parts[0],
			DefaultBranch: h.opts.Branch,
			SSHURL:        fmt.Sprintf("git@github.com:%s.git", h.opts.Repo),
			CloneURL:      fmt.Sprintf("https://github.com/%s.git", h.opts.Repo),
		}}
	} else {
		// Org mode
		targetName = h.opts.Org
		h.term.Info("Fetching repositories for %s...", h.term.Color(terminal.Cyan, h.opts.Org))

		var err error
		repos, err = h.gh.ListOrgRepos(h.opts.Org, h.opts.Limit)
		if err != nil {
			return nil, fmt.Errorf("listing repos: %w", err)
		}

		if len(repos) == 0 {
			return nil, fmt.Errorf("no repositories found for org: %s", h.opts.Org)
		}
	}

	// Get scanners for profile
	scanners, err := h.cfg.GetProfileScanners(h.opts.Profile)
	if err != nil {
		return nil, fmt.Errorf("getting scanners: %w", err)
	}

	// Print header
	h.term.Divider()
	if h.opts.Org != "" {
		h.term.Info("%s %s", h.term.Color(terminal.Bold, "Hydrate Organization:"), h.term.Color(terminal.Cyan, h.opts.Org))
	} else {
		h.term.Info("%s %s", h.term.Color(terminal.Bold, "Hydrate Repository:"), h.term.Color(terminal.Cyan, h.opts.Repo))
	}
	h.term.Info("Scan ID:      %s", h.term.Color(terminal.Dim, scanID))
	h.term.Info("Repositories: %s", h.term.Color(terminal.Cyan, strconv.Itoa(len(repos))))
	h.term.Info("Profile:      %s", h.term.Color(terminal.Cyan, h.opts.Profile))
	h.term.Info("Scanners:     %s", h.term.Color(terminal.Cyan, fmt.Sprintf("%d parallel", h.opts.ParallelScanners)))

	// Phase 1: Clone
	h.term.Header("CLONING")
	repoStatuses, err := h.cloneRepos(ctx, repos)
	if err != nil {
		return nil, err
	}

	// Collect project IDs for return
	var projectIDs []string
	for _, status := range repoStatuses {
		if status.CloneOK {
			projectIDs = append(projectIDs, github.ProjectID(status.Repo.NameWithOwner))
		}
	}

	// Stop here if clone-only
	if h.opts.CloneOnly {
		h.term.Divider()
		h.term.Success("Clone complete (--clone-only)")
		return projectIDs, nil
	}

	// Check for slow scanners
	skipScanners := h.opts.SkipScanners
	if h.shouldWarnSlowScanners(repoStatuses, scanners) {
		skipScanners = h.handleSlowScannerWarning(repoStatuses, scanners)
	}

	// Phase 2: Scan
	h.term.Header("SCANNING")
	successCount, failedCount := h.scanRepos(ctx, repoStatuses, scanners, skipScanners)

	h.term.ClearLine()
	h.term.ScanComplete()

	// Print per-scanner results table
	h.printScannerResultsTable(repoStatuses, scanners, skipScanners)

	// Aggregate findings from all repos
	findings := h.aggregateFindings(repoStatuses)

	// Print summary
	duration := int(time.Since(start).Seconds())
	diskUsage := h.getDiskUsage()
	totalFiles := h.getTotalFiles(repoStatuses)

	h.term.Divider()
	h.term.SummaryWithFindings(targetName, duration, successCount, failedCount, diskUsage, formatNumber(totalFiles), findings)

	return projectIDs, nil
}

// cloneRepos clones all repositories (sequential by default for clean output)
func (h *Hydrate) cloneRepos(ctx context.Context, repos []github.Repository) ([]*RepoStatus, error) {
	statuses := make([]*RepoStatus, len(repos))
	var wg sync.WaitGroup
	sem := make(chan struct{}, h.opts.ParallelRepos)

	for i, repo := range repos {
		statuses[i] = &RepoStatus{Repo: repo}
	}

	for i := range repos {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			status := statuses[idx]
			h.cloneRepo(ctx, status)
		}(i)
	}

	wg.Wait()
	return statuses, nil
}

// cloneRepo clones a single repository
func (h *Hydrate) cloneRepo(ctx context.Context, status *RepoStatus) {
	repo := status.Repo
	projectID := github.ProjectID(repo.NameWithOwner)
	repoPath := filepath.Join(h.zeroHome, "repos", projectID, "repo")
	status.RepoPath = repoPath

	// Check if already cloned (must have .git directory to be valid)
	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		// Already exists, just update stats
		status.CloneOK = true
		status.FileCount = h.countFiles(repoPath)

		// Get commit hash
		commit := h.getCommitHash(repoPath)
		size := h.getRepoSize(repoPath)

		h.term.RepoCloned(
			repo.Name,
			size,
			formatNumber(status.FileCount),
			commit,
			"up to date",
		)
		return
	}

	// Remove empty/invalid repo directory if it exists
	if _, err := os.Stat(repoPath); err == nil {
		os.RemoveAll(repoPath)
	}

	// Create directory
	if err := os.MkdirAll(filepath.Dir(repoPath), 0755); err != nil {
		h.term.Error("%s clone failed: %v", repo.Name, err)
		return
	}

	// Clone with depth=1 (prefer HTTPS for broader compatibility)
	cloneURL := repo.CloneURL
	if cloneURL == "" {
		// Build HTTPS URL from nameWithOwner
		cloneURL = fmt.Sprintf("https://github.com/%s.git", repo.NameWithOwner)
	}
	if cloneURL == "" {
		cloneURL = repo.SSHURL
	}
	cmd := exec.CommandContext(ctx, "git", "clone",
		"--depth", "1",
		cloneURL,
		repoPath,
	)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		h.term.Error("%s clone failed: %v", repo.Name, err)
		return
	}

	status.CloneOK = true
	status.FileCount = h.countFiles(repoPath)

	commit := h.getCommitHash(repoPath)
	size := h.getRepoSize(repoPath)

	h.term.RepoCloned(
		repo.Name,
		size,
		formatNumber(status.FileCount),
		commit,
		"cloned",
	)
}

// scanRepos scans all repositories sequentially with live progress
func (h *Hydrate) scanRepos(ctx context.Context, statuses []*RepoStatus, scanners, skipScanners []string) (success, failed int) {
	// Process repos sequentially for clear progress display
	for _, status := range statuses {
		// Check for cancellation
		select {
		case <-ctx.Done():
			h.term.Warning("Scan interrupted")
			return success, failed
		default:
		}

		if !status.CloneOK {
			failed++
			continue
		}

		// Initialize progress for this repo
		status.Progress = scanner.NewProgress(scanners)
		estimate := scanner.TotalEstimate(scanners, status.FileCount)
		h.term.RepoScanning(status.Repo.Name, estimate)

		// Build scanner line positions (from bottom: 0 = last scanner, N-1 = first scanner)
		scannerLinePos := make(map[string]int)
		lineNum := 0
		for _, s := range scanners {
			if contains(skipScanners, s) {
				h.term.ScannerSkipped(s)
				status.Progress.SetSkipped(s)
			} else {
				h.term.ScannerQueued(s, scanner.EstimateTime(s, status.FileCount))
				scannerLinePos[s] = lineNum
				lineNum++
			}
		}
		totalLines := lineNum

		// Run scan with real-time line updates
		h.scanRepoWithProgress(ctx, status, scanners, skipScanners, scannerLinePos, totalLines)

		// Move cursor past all scanner lines and print completion
		fmt.Println() // Ensure we're on a new line
		if status.ScanOK {
			h.term.Success("%s complete (%ds)", status.Repo.Name, int(status.Duration.Seconds()))
		} else {
			h.term.Error("%s failed", status.Repo.Name)
		}
		fmt.Println()

		if status.ScanOK {
			success++
		} else {
			failed++
		}
	}

	return success, failed
}

// scanRepoWithProgress runs all scanners with real-time line updates
func (h *Hydrate) scanRepoWithProgress(ctx context.Context, status *RepoStatus, scanners, skipScanners []string, linePos map[string]int, totalLines int) {
	start := time.Now()

	// Get scanners from registry
	scannerList, err := scanner.GetByNames(scanners)
	if err != nil {
		status.ScanOK = false
		status.Duration = time.Since(start)
		return
	}

	// Build output directory
	projectID := github.ProjectID(status.Repo.NameWithOwner)
	outputDir := filepath.Join(h.zeroHome, "repos", projectID, "analysis")

	// Track scanner start times for duration calculation (protected by mutex)
	var startTimesMu sync.Mutex
	scannerStartTimes := make(map[string]time.Time)

	// Set up progress callback with real-time line updates
	h.runner.OnProgress = func(name string, st scanner.Status, summary string) {
		pos, hasPos := linePos[name]
		if !hasPos {
			return // Skip scanners not in our list
		}

		// Calculate lines up from current position (bottom of scanner list)
		linesUp := totalLines - pos

		switch st {
		case scanner.StatusRunning:
			status.Progress.SetRunning(name)
			startTimesMu.Lock()
			scannerStartTimes[name] = time.Now()
			startTimesMu.Unlock()
			h.term.UpdateScannerStatus(linesUp, name, "running", terminal.IconArrow, terminal.Cyan, "")

		case scanner.StatusComplete:
			startTimesMu.Lock()
			startTime := scannerStartTimes[name]
			startTimesMu.Unlock()
			duration := time.Since(startTime)
			status.Progress.SetComplete(name, summary, duration)
			h.term.UpdateScannerStatus(linesUp, name, summary, terminal.IconSuccess, terminal.Green, fmt.Sprintf("%ds", int(duration.Seconds())))

		case scanner.StatusFailed:
			startTimesMu.Lock()
			startTime := scannerStartTimes[name]
			startTimesMu.Unlock()
			duration := time.Since(startTime)
			status.Progress.SetFailed(name, nil, duration)
			errMsg := "failed"
			if summary != "" {
				errMsg = summary
			}
			h.term.UpdateScannerStatus(linesUp, name, errMsg, terminal.IconFailed, terminal.Red, fmt.Sprintf("%ds", int(duration.Seconds())))
		}
	}

	// Run scanners (parallel execution within repo)
	result, err := h.runner.RunScanners(ctx, scanner.RunOptions{
		RepoPath:     status.RepoPath,
		OutputDir:    outputDir,
		Scanners:     scannerList,
		SkipScanners: skipScanners,
		Parallel:     h.opts.ParallelScanners,
	})
	status.Duration = time.Since(start)

	if err != nil {
		status.ScanOK = false
		return
	}

	status.ScanOK = result.Success

	// Copy results to progress tracker
	for name, res := range result.Results {
		if r, ok := status.Progress.Results[name]; ok {
			r.Status = res.Status
			r.Summary = res.Summary
			r.Duration = res.Duration
			r.Error = res.Error
		}
	}
}

// shouldWarnSlowScanners checks if we should warn about slow scanners
func (h *Hydrate) shouldWarnSlowScanners(statuses []*RepoStatus, scanners []string) bool {
	if h.opts.SkipSlow || h.opts.Yes {
		return false
	}

	// Check if any repo is large (>20k files)
	for _, s := range statuses {
		if s.FileCount > 20000 {
			// Check if profile has slow scanners
			for _, scanner := range scanners {
				if scanner == "package-malcontent" || scanner == "code-vulns" {
					return true
				}
			}
		}
	}
	return false
}

// handleSlowScannerWarning displays warning and returns scanners to skip
func (h *Hydrate) handleSlowScannerWarning(statuses []*RepoStatus, scanners []string) []string {
	// Find largest repo
	var largest *RepoStatus
	for _, s := range statuses {
		if largest == nil || s.FileCount > largest.FileCount {
			largest = s
		}
	}

	fmt.Println()
	h.term.Warning("Slow scanner warning")
	h.term.Info("    Largest repo: %s (%s files)", h.term.Color(terminal.Cyan, largest.Repo.Name), formatNumber(largest.FileCount))
	fmt.Println()

	// Show slow scanners and estimate total time
	slowScanners := []string{}
	totalEstimate := 0
	for _, s := range scanners {
		est := scanner.EstimateTime(s, largest.FileCount)
		if est > 10 {
			h.term.Info("    %s %s: ~%ds on large repos",
				h.term.Color(terminal.Yellow, "•"),
				s,
				est,
			)
			slowScanners = append(slowScanners, s)
			totalEstimate += est
		}
	}

	if len(slowScanners) == 0 {
		return nil
	}

	fmt.Println()
	h.term.Info("    Total estimated time for slow scanners: ~%ds", totalEstimate)
	fmt.Println()

	// If --skip-slow was specified, skip without prompting
	if h.opts.SkipSlow {
		h.term.Info("    %s Skipping slow scanners (--skip-slow)", h.term.Color(terminal.Yellow, terminal.IconSkipped))
		return slowScanners
	}

	// Interactive prompt
	skip := h.term.Confirm("    Skip slow scanners?", true)
	if skip {
		h.term.Info("    %s Skipping: %s", h.term.Color(terminal.Yellow, terminal.IconSkipped), strings.Join(slowScanners, ", "))
		return slowScanners
	}

	h.term.Info("    %s Running all scanners (this may take a while)", h.term.Color(terminal.Green, terminal.IconSuccess))
	return nil
}

// Helper methods

func (h *Hydrate) countFiles(path string) int {
	count := 0
	filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && !strings.Contains(p, ".git") {
			count++
		}
		return nil
	})
	return count
}

func (h *Hydrate) getCommitHash(path string) string {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--short", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func (h *Hydrate) getRepoSize(path string) string {
	cmd := exec.Command("du", "-sh", path)
	out, err := cmd.Output()
	if err != nil {
		return "0"
	}
	fields := strings.Fields(string(out))
	if len(fields) > 0 {
		return strings.ToLower(fields[0])
	}
	return "0"
}

func (h *Hydrate) getDiskUsage() string {
	reposPath := filepath.Join(h.zeroHome, "repos")
	cmd := exec.Command("du", "-sh", reposPath)
	out, err := cmd.Output()
	if err != nil {
		return "0"
	}
	fields := strings.Fields(string(out))
	if len(fields) > 0 {
		return fields[0]
	}
	return "0"
}

func (h *Hydrate) getTotalFiles(statuses []*RepoStatus) int {
	total := 0
	for _, s := range statuses {
		total += s.FileCount
	}
	return total
}


func formatNumber(n int) string {
	str := strconv.Itoa(n)
	if n < 1000 {
		return str
	}

	// Add thousand separators
	result := ""
	for i, c := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result += ","
		}
		result += string(c)
	}
	return result
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// printScannerResultsTable prints a summary table of all scanner results
func (h *Hydrate) printScannerResultsTable(statuses []*RepoStatus, scanners, skipScanners []string) {
	// Aggregate scanner results across all repos
	scannerResults := make(map[string]*aggregatedScannerResult)

	for _, name := range scanners {
		scannerResults[name] = &aggregatedScannerResult{
			Name:    name,
			Skipped: contains(skipScanners, name),
		}
	}

	// Aggregate results from all repos
	for _, status := range statuses {
		if status.Progress == nil {
			continue
		}
		for name, result := range status.Progress.Results {
			agg, ok := scannerResults[name]
			if !ok {
				continue
			}

			switch result.Status {
			case scanner.StatusComplete:
				agg.SuccessCount++
				agg.TotalDuration += result.Duration
				if result.Summary != "" {
					agg.LastSummary = result.Summary
				}
			case scanner.StatusFailed:
				agg.FailedCount++
				agg.TotalDuration += result.Duration
			case scanner.StatusSkipped:
				agg.Skipped = true
			}
		}
	}

	// Build rows for display (preserve scanner order)
	rows := make([]terminal.ScannerResultRow, 0, len(scanners))
	for _, name := range scanners {
		agg := scannerResults[name]

		status := "success"
		summary := agg.LastSummary
		if agg.Skipped {
			status = "skipped"
			summary = ""
		} else if agg.FailedCount > 0 {
			if agg.SuccessCount > 0 {
				status = "partial"
				summary = fmt.Sprintf("%d ok, %d failed", agg.SuccessCount, agg.FailedCount)
			} else {
				status = "failed"
				summary = "all failed"
			}
		} else if agg.SuccessCount == 0 {
			status = "skipped"
			summary = ""
		}

		if summary == "" && status == "success" {
			summary = "complete"
		}

		rows = append(rows, terminal.ScannerResultRow{
			Name:     name,
			Status:   status,
			Summary:  summary,
			Duration: agg.TotalDuration,
		})
	}

	h.term.ScannerResultsTable(rows)
}

type aggregatedScannerResult struct {
	Name          string
	SuccessCount  int
	FailedCount   int
	Skipped       bool
	TotalDuration time.Duration
	LastSummary   string
}

// aggregateFindings reads scan results and aggregates findings across all repos
func (h *Hydrate) aggregateFindings(statuses []*RepoStatus) *terminal.ScanFindings {
	findings := &terminal.ScanFindings{
		PackagesByEco: make(map[string]int),
		VulnsByEco:    make(map[string]int),
		LicenseCounts: make(map[string]int),
	}

	for _, status := range statuses {
		if !status.ScanOK {
			continue
		}

		projectID := github.ProjectID(status.Repo.NameWithOwner)
		analysisDir := filepath.Join(h.zeroHome, "repos", projectID, "analysis")

		// Aggregate SBOM data
		h.aggregateSBOM(analysisDir, findings)

		// Aggregate vulnerability data
		h.aggregateVulns(analysisDir, findings)

		// Aggregate license data
		h.aggregateLicenses(analysisDir, findings)

		// Aggregate secrets data
		h.aggregateSecrets(analysisDir, findings)

		// Aggregate malcontent data
		h.aggregateMalcontent(analysisDir, findings)

		// Aggregate health data
		h.aggregateHealth(analysisDir, findings)
	}

	return findings
}

func (h *Hydrate) aggregateSBOM(dir string, findings *terminal.ScanFindings) {
	path := filepath.Join(dir, "package-sbom.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var result struct {
		Summary struct {
			TotalPackages int            `json:"total_packages"`
			ByEcosystem   map[string]int `json:"by_ecosystem"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return
	}

	findings.TotalPackages += result.Summary.TotalPackages
	for eco, count := range result.Summary.ByEcosystem {
		findings.PackagesByEco[eco] += count
	}
}

func (h *Hydrate) aggregateVulns(dir string, findings *terminal.ScanFindings) {
	path := filepath.Join(dir, "package-vulns.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var result struct {
		Summary struct {
			Critical int `json:"critical"`
			High     int `json:"high"`
			Medium   int `json:"medium"`
			Low      int `json:"low"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return
	}

	findings.VulnCritical += result.Summary.Critical
	findings.VulnHigh += result.Summary.High
	findings.VulnMedium += result.Summary.Medium
	findings.VulnLow += result.Summary.Low
}

func (h *Hydrate) aggregateLicenses(dir string, findings *terminal.ScanFindings) {
	path := filepath.Join(dir, "licenses.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var result struct {
		Summary struct {
			UniqueLicenses int            `json:"unique_licenses"`
			LicenseCounts  map[string]int `json:"license_counts"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return
	}

	// Track unique license types
	if result.Summary.UniqueLicenses > findings.LicenseTypes {
		findings.LicenseTypes = result.Summary.UniqueLicenses
	}
	for lic, count := range result.Summary.LicenseCounts {
		findings.LicenseCounts[lic] += count
	}
}

func (h *Hydrate) aggregateSecrets(dir string, findings *terminal.ScanFindings) {
	path := filepath.Join(dir, "code-secrets.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var result struct {
		Summary struct {
			Critical      int `json:"critical"`
			High          int `json:"high"`
			Medium        int `json:"medium"`
			TotalFindings int `json:"total_findings"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return
	}

	findings.SecretsCritical += result.Summary.Critical
	findings.SecretsHigh += result.Summary.High
	findings.SecretsMedium += result.Summary.Medium
	findings.SecretsTotal += result.Summary.TotalFindings
}

func (h *Hydrate) aggregateMalcontent(dir string, findings *terminal.ScanFindings) {
	path := filepath.Join(dir, "package-malcontent.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var result struct {
		Summary struct {
			Critical int `json:"critical"`
			High     int `json:"high"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return
	}

	findings.MalcontentCrit += result.Summary.Critical
	findings.MalcontentHigh += result.Summary.High
}

func (h *Hydrate) aggregateHealth(dir string, findings *terminal.ScanFindings) {
	path := filepath.Join(dir, "package-health.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var result struct {
		Summary struct {
			CriticalCount int `json:"critical_count"`
			WarningCount  int `json:"warning_count"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return
	}

	findings.HealthCritical += result.Summary.CriticalCount
	findings.HealthWarnings += result.Summary.WarningCount
}
