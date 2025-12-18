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
	"github.com/crashappsec/zero/pkg/diff"
	"github.com/crashappsec/zero/pkg/github"
	"github.com/crashappsec/zero/pkg/languages"
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
	Repo      github.Repository
	RepoPath  string
	FileCount int
	CloneOK   bool
	ScanOK    bool
	Progress  *scanner.Progress
	Duration  time.Duration

	// SBOM stats (populated after scan)
	SBOMPackages int    // Number of packages in SBOM
	SBOMSize     int64  // Size of SBOM file in bytes
	SBOMPath     string // Path to SBOM file

	// Code ownership stats (populated after scan)
	OwnershipContributors    int    // Number of contributors in analysis period
	OwnershipAllTime         int    // All-time contributors
	OwnershipLanguages       int    // Number of languages detected
	OwnershipTopLanguage     string // Top language by file count
	OwnershipBusFactor       int    // Bus factor (0 if not calculated)
	OwnershipBusFactorRisk   string // Bus factor risk level
	OwnershipTotalCommits    int    // Total commits in repo
	OwnershipLastCommitDays  int    // Days since last commit
	OwnershipPeriodAdjusted  bool   // Whether analysis period was extended
	OwnershipActivityStatus  string // Repo activity status

	// Tech-id stats (populated after scan)
	TechIDTotalTech     int      // Total technologies detected
	TechIDTopTechs      []string // Top 3 technologies
	TechIDTotalModels   int      // ML models detected
	TechIDSecurityCount int      // Security findings

	// Scanners that were run
	ScannersRun []string
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

	// Aggregate findings from all repos (only for scanners that were run, excluding skipped)
	runningScanners := filterOutSkipped(scanners, skipScanners)
	findings := h.aggregateFindings(repoStatuses, runningScanners)

	// Preserve scan history for diff/delta tracking
	h.preserveHistory(repoStatuses, runningScanners, scanID, start)

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
		// Already exists - check if we need to deepen history for ownership analysis
		if h.needsDeepHistory() && h.isShallowRepo(repoPath) {
			h.deepenHistory(ctx, repoPath)
		}

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

	// Clone with configurable depth (prefer HTTPS for broader compatibility)
	cloneURL := repo.CloneURL
	if cloneURL == "" && repo.NameWithOwner != "" {
		// Build HTTPS URL from nameWithOwner
		cloneURL = fmt.Sprintf("https://github.com/%s.git", repo.NameWithOwner)
	}
	if cloneURL == "" {
		// Fallback to SSH URL if HTTPS is unavailable
		cloneURL = repo.SSHURL
	}

	// Determine clone depth - use configured depth, profile default, or 1
	cloneDepth := 1
	if h.opts.Depth > 0 {
		cloneDepth = h.opts.Depth
	} else if h.needsDeepHistory() {
		// Profiles that need git history for ownership analysis
		cloneDepth = 100 // Enough commits for 90-day analysis
	}

	cloneArgs := []string{"clone", "--depth", fmt.Sprintf("%d", cloneDepth), cloneURL, repoPath}
	cmd := exec.CommandContext(ctx, "git", cloneArgs...)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		h.term.Error("%s clone failed: %v", repo.Name, err)
		return
	}

	status.CloneOK = true
	status.FileCount = h.countFiles(repoPath)

	// Detect and cache languages (available to all scanners)
	h.detectAndCacheLanguages(repo, repoPath)

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
	// Use parallel scanning if configured and multiple repos
	if h.opts.ParallelRepos > 1 && len(statuses) > 1 {
		return h.scanReposParallel(ctx, statuses, scanners, skipScanners)
	}

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
			msg := formatScanCompleteMessage(status)
			h.term.Success("%s", msg)
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

// scanReposParallel runs scans on multiple repos concurrently
func (h *Hydrate) scanReposParallel(ctx context.Context, statuses []*RepoStatus, scanners, skipScanners []string) (success, failed int) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := make(chan struct{}, h.opts.ParallelRepos)

	// Print header for parallel mode
	fmt.Printf("\n  Scanning %d repos in parallel (%d concurrent)...\n\n",
		len(statuses), h.opts.ParallelRepos)

	for _, status := range statuses {
		if !status.CloneOK {
			mu.Lock()
			failed++
			mu.Unlock()
			continue
		}

		wg.Add(1)
		go func(s *RepoStatus) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				return
			}

			// Initialize progress
			s.Progress = scanner.NewProgress(scanners)

			// Run scan (simplified - no line updates in parallel mode)
			h.scanRepoSimple(ctx, s, scanners, skipScanners)

			// Print result with scanner-appropriate stats
			mu.Lock()
			if s.ScanOK {
				msg := formatScanCompleteMessage(s)
				h.term.Success("%s", msg)
				success++
			} else {
				h.term.Error("%s failed", s.Repo.Name)
				failed++
			}
			mu.Unlock()
		}(status)
	}

	wg.Wait()
	fmt.Println()
	return success, failed
}

// scanRepoSimple runs scanners without real-time line updates (for parallel mode)
func (h *Hydrate) scanRepoSimple(ctx context.Context, status *RepoStatus, scanners, skipScanners []string) {
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

	// Run scanners
	opts := scanner.RunOptions{
		RepoPath:     status.RepoPath,
		OutputDir:    outputDir,
		Scanners:     scannerList,
		SkipScanners: skipScanners,
		Parallel:     h.opts.ParallelScanners,
		Timeout:      time.Duration(h.cfg.Settings.ScannerTimeoutSeconds) * time.Second,
	}

	result, err := h.runner.RunScanners(ctx, opts)
	if err != nil {
		status.ScanOK = false
		status.Duration = time.Since(start)
		return
	}

	// Check results
	allSuccess := result.Success
	for name, r := range result.Results {
		if r.Status == scanner.StatusFailed {
			allSuccess = false
		}
		// Update progress tracking (copy status and summary)
		if status.Progress.Results[name] != nil {
			status.Progress.Results[name].Status = r.Status
			status.Progress.Results[name].Summary = r.Summary
			status.Progress.Results[name].Duration = r.Duration
		}
	}

	status.ScanOK = allSuccess
	status.Duration = time.Since(start)
	status.ScannersRun = scanners

	// Extract stats from scanner outputs
	h.extractSBOMStats(status, outputDir)
	h.extractOwnershipStats(status, outputDir)
	h.extractTechIDStats(status, outputDir)
}

// extractSBOMStats reads SBOM file to get package count and file size
func (h *Hydrate) extractSBOMStats(status *RepoStatus, outputDir string) {
	sbomPath := filepath.Join(outputDir, "sbom.cdx.json")
	status.SBOMPath = sbomPath

	// Get file size
	info, err := os.Stat(sbomPath)
	if err != nil {
		return
	}
	status.SBOMSize = info.Size()

	// Read SBOM to get package count
	data, err := os.ReadFile(sbomPath)
	if err != nil {
		return
	}

	var sbom struct {
		Components []json.RawMessage `json:"components"`
	}
	if err := json.Unmarshal(data, &sbom); err != nil {
		return
	}
	status.SBOMPackages = len(sbom.Components)
}

// extractTechIDStats reads tech-id.json to get technology detection stats
func (h *Hydrate) extractTechIDStats(status *RepoStatus, outputDir string) {
	techIDPath := filepath.Join(outputDir, "tech-id.json")

	data, err := os.ReadFile(techIDPath)
	if err != nil {
		return
	}

	var techID struct {
		Summary struct {
			Technology *struct {
				TotalTechnologies int      `json:"total_technologies"`
				TopTechnologies   []string `json:"top_technologies"`
			} `json:"technology"`
			Models *struct {
				TotalModels int `json:"total_models"`
			} `json:"models"`
			Security *struct {
				TotalFindings int `json:"total_findings"`
			} `json:"security"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(data, &techID); err != nil {
		return
	}

	if techID.Summary.Technology != nil {
		status.TechIDTotalTech = techID.Summary.Technology.TotalTechnologies
		status.TechIDTopTechs = techID.Summary.Technology.TopTechnologies
	}
	if techID.Summary.Models != nil {
		status.TechIDTotalModels = techID.Summary.Models.TotalModels
	}
	if techID.Summary.Security != nil {
		status.TechIDSecurityCount = techID.Summary.Security.TotalFindings
	}
}

// extractOwnershipStats reads code-ownership.json to get contributor and language stats
func (h *Hydrate) extractOwnershipStats(status *RepoStatus, outputDir string) {
	ownershipPath := filepath.Join(outputDir, "code-ownership.json")

	data, err := os.ReadFile(ownershipPath)
	if err != nil {
		return
	}

	var ownership struct {
		Summary struct {
			TotalContributors      int    `json:"total_contributors"`
			AllTimeContributors    int    `json:"all_time_contributors"`
			LanguagesDetected      int    `json:"languages_detected"`
			BusFactor              int    `json:"bus_factor"`
			BusFactorRisk          string `json:"bus_factor_risk"`
			TotalCommits           int    `json:"total_commits"`
			DaysSinceLastCommit    int    `json:"days_since_last_commit"`
			AnalysisPeriodAdjusted bool   `json:"analysis_period_adjusted"`
			RepoActivityStatus     string `json:"repo_activity_status"`
			TopLanguages           []struct {
				Name       string  `json:"name"`
				FileCount  int     `json:"file_count"`
				Percentage float64 `json:"percentage"`
			} `json:"top_languages"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(data, &ownership); err != nil {
		return
	}

	status.OwnershipContributors = ownership.Summary.TotalContributors
	status.OwnershipAllTime = ownership.Summary.AllTimeContributors
	status.OwnershipLanguages = ownership.Summary.LanguagesDetected
	status.OwnershipBusFactor = ownership.Summary.BusFactor
	status.OwnershipBusFactorRisk = ownership.Summary.BusFactorRisk
	status.OwnershipTotalCommits = ownership.Summary.TotalCommits
	status.OwnershipLastCommitDays = ownership.Summary.DaysSinceLastCommit
	status.OwnershipPeriodAdjusted = ownership.Summary.AnalysisPeriodAdjusted
	status.OwnershipActivityStatus = ownership.Summary.RepoActivityStatus

	// Get top language name
	if len(ownership.Summary.TopLanguages) > 0 {
		status.OwnershipTopLanguage = ownership.Summary.TopLanguages[0].Name
	}
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
			// Only set running state if not already running (avoids resetting start time on status updates)
			startTimesMu.Lock()
			_, hasStartTime := scannerStartTimes[name]
			if !hasStartTime {
				scannerStartTimes[name] = time.Now()
				status.Progress.SetRunning(name)
			}
			startTimesMu.Unlock()

			// Show real-time status message if provided, otherwise show "running"
			displayStatus := "running"
			if summary != "" {
				displayStatus = summary
			}
			h.term.UpdateScannerStatus(linesUp, name, displayStatus, terminal.IconArrow, terminal.Cyan, "")

		case scanner.StatusComplete:
			startTimesMu.Lock()
			startTime, hasStartTime := scannerStartTimes[name]
			startTimesMu.Unlock()
			var duration time.Duration
			if hasStartTime {
				duration = time.Since(startTime)
			}
			status.Progress.SetComplete(name, summary, duration)
			h.term.UpdateScannerStatus(linesUp, name, summary, terminal.IconSuccess, terminal.Green, fmt.Sprintf("%ds", int(duration.Seconds())))

		case scanner.StatusFailed:
			startTimesMu.Lock()
			startTime, hasStartTime := scannerStartTimes[name]
			startTimesMu.Unlock()
			var duration time.Duration
			if hasStartTime {
				duration = time.Since(startTime)
			}
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
	status.ScannersRun = scanners

	// Copy results to progress tracker
	for name, res := range result.Results {
		if r, ok := status.Progress.Results[name]; ok {
			r.Status = res.Status
			r.Summary = res.Summary
			r.Duration = res.Duration
			r.Error = res.Error
		}
	}

	// Extract stats from scanner outputs
	h.extractSBOMStats(status, outputDir)
	h.extractOwnershipStats(status, outputDir)
	h.extractTechIDStats(status, outputDir)
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

// needsDeepHistory returns true if the current profile requires git history for analysis
func (h *Hydrate) needsDeepHistory() bool {
	profile := strings.ToLower(h.opts.Profile)
	// Profiles that include code-ownership scanner need git history
	deepHistoryProfiles := []string{
		"code-ownership-only",
		"ownership",
		"full",
		"health",
	}
	for _, p := range deepHistoryProfiles {
		if profile == p {
			return true
		}
	}
	return false
}

// isShallowRepo checks if the repository is a shallow clone
func (h *Hydrate) isShallowRepo(repoPath string) bool {
	shallowFile := filepath.Join(repoPath, ".git", "shallow")
	_, err := os.Stat(shallowFile)
	return err == nil
}

// detectAndCacheLanguages runs language detection and caches results for all scanners
func (h *Hydrate) detectAndCacheLanguages(repo github.Repository, repoPath string) {
	// Build analysis directory path
	projectID := github.ProjectID(repo.NameWithOwner)
	analysisDir := filepath.Join(h.zeroHome, "repos", projectID, "analysis")

	// Create analysis directory if needed
	if err := os.MkdirAll(analysisDir, 0755); err != nil {
		return // Best effort - scanners can detect languages themselves if this fails
	}

	// Run language detection
	opts := languages.DefaultScanOptions()
	opts.OnlyProgramming = true
	stats, err := languages.ScanDirectory(repoPath, opts)
	if err != nil {
		return
	}

	// Write to languages.json
	langFile := filepath.Join(analysisDir, "languages.json")
	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(langFile, data, 0644)
}

// deepenHistory fetches more commit history for a shallow clone
func (h *Hydrate) deepenHistory(ctx context.Context, repoPath string) {
	// Fetch more commits (100 should cover 90 days of activity for most repos)
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "fetch", "--deepen=100")
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Run() // Ignore errors - best effort
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

// formatBytes formats a byte size in human-readable form
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1fGB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1fMB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1fKB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}

// formatScanCompleteMessage creates an appropriate completion message based on which scanners were run
func formatScanCompleteMessage(s *RepoStatus) string {
	base := fmt.Sprintf("%s complete (%ds)", s.Repo.Name, int(s.Duration.Seconds()))

	// Check which scanners were run to show appropriate stats
	hasOwnershipOnly := len(s.ScannersRun) == 1 && s.ScannersRun[0] == "code-ownership"
	hasSBOM := containsScanner(s.ScannersRun, "sbom")

	// Code ownership only - show ownership stats (bus factor, contributors)
	if hasOwnershipOnly {
		// Build a comprehensive ownership summary
		parts := []string{}

		// Bus factor risk is the key metric
		if s.OwnershipBusFactorRisk != "" {
			parts = append(parts, fmt.Sprintf("bus factor: %s", s.OwnershipBusFactorRisk))
		}

		// Show contributor info (prefer period-specific, fallback to all-time)
		if s.OwnershipContributors > 0 {
			if s.OwnershipAllTime > s.OwnershipContributors {
				parts = append(parts, fmt.Sprintf("%d contributors (%d all-time)", s.OwnershipContributors, s.OwnershipAllTime))
			} else {
				parts = append(parts, fmt.Sprintf("%d contributors", s.OwnershipContributors))
			}
		} else if s.OwnershipAllTime > 0 {
			parts = append(parts, fmt.Sprintf("%d all-time contributors", s.OwnershipAllTime))
		}

		// Show activity status if repo is stale
		if s.OwnershipLastCommitDays > 90 && s.OwnershipTotalCommits > 0 {
			parts = append(parts, fmt.Sprintf("last commit %dd ago", s.OwnershipLastCommitDays))
		}

		// Note if period was adjusted
		if s.OwnershipPeriodAdjusted {
			parts = append(parts, "period extended")
		}

		if len(parts) > 0 {
			return fmt.Sprintf("%s - %s", base, strings.Join(parts, ", "))
		}
		// Truly no data
		return fmt.Sprintf("%s - no ownership data", base)
	}

	// SBOM was run - show package stats
	if hasSBOM && (s.SBOMPackages > 0 || s.SBOMSize > 0) {
		return fmt.Sprintf("%s - %d packages, %s", base, s.SBOMPackages, formatBytes(s.SBOMSize))
	}

	// Tech-id was run - show technology stats (if SBOM didn't already return above)
	hasTechID := containsScanner(s.ScannersRun, "tech-id")
	if hasTechID && s.TechIDTotalTech > 0 {
		result := fmt.Sprintf("%s - %d tech", base, s.TechIDTotalTech)
		if len(s.TechIDTopTechs) > 0 {
			result += ": " + strings.Join(s.TechIDTopTechs, ", ")
		}
		if s.TechIDTotalModels > 0 {
			result += fmt.Sprintf(", %d models", s.TechIDTotalModels)
		}
		return result
	}

	// Fallback to basic message
	return base
}

// containsScanner checks if a scanner name is in the list
func containsScanner(scanners []string, name string) bool {
	for _, s := range scanners {
		if s == name {
			return true
		}
	}
	return false
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// filterOutSkipped returns scanners excluding those in skipScanners
func filterOutSkipped(scanners, skipScanners []string) []string {
	if len(skipScanners) == 0 {
		return scanners
	}
	skipSet := make(map[string]bool)
	for _, s := range skipScanners {
		skipSet[s] = true
	}
	var result []string
	for _, s := range scanners {
		if !skipSet[s] {
			result = append(result, s)
		}
	}
	return result
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
// Only aggregates data from scanners that were actually run in this session
func (h *Hydrate) aggregateFindings(statuses []*RepoStatus, runningScanners []string) *terminal.ScanFindings {
	// Build set of running scanners for quick lookup
	scannerSet := make(map[string]bool)
	for _, s := range runningScanners {
		scannerSet[s] = true
	}

	findings := &terminal.ScanFindings{
		ScannersRun:   scannerSet,
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

		// Aggregate SBOM data (from sbom scanner)
		if scannerSet["sbom"] {
			h.aggregateSBOM(analysisDir, findings)
			// Add SBOM path and size to findings
			sbomPath := filepath.Join(analysisDir, "sbom.cdx.json")
			if _, err := os.Stat(sbomPath); err == nil {
				findings.SBOMPaths = append(findings.SBOMPaths, sbomPath)
			}
			// Add SBOM size from status (already extracted in scan phase)
			findings.SBOMSizeTotal += status.SBOMSize
		}

		// Aggregate vulnerability data (from package-analysis scanner)
		if scannerSet["package-analysis"] {
			h.aggregateVulns(analysisDir, findings)
			h.aggregateLicenses(analysisDir, findings)
			h.aggregateMalcontent(analysisDir, findings)
		}

		// Aggregate secrets data (from code-security scanner)
		if scannerSet["code-security"] {
			h.aggregateSecrets(analysisDir, findings)
		}

		// Aggregate tech-id data (from tech-id scanner)
		if scannerSet["tech-id"] {
			h.aggregateTechID(analysisDir, findings)
		}

		// Note: aggregateHealth was removed as the health scanner no longer exists.
		// The code-quality scanner provides different metrics (tech debt, complexity, coverage).
	}

	// Deduplicate top technologies across repos
	if len(findings.TechTopList) > 0 {
		findings.TechTopList = deduplicateAndLimit(findings.TechTopList, 5)
	}

	return findings
}

// deduplicateAndLimit removes duplicates and limits to n items
func deduplicateAndLimit(items []string, limit int) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, limit)
	for _, item := range items {
		if !seen[item] && len(result) < limit {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}

func (h *Hydrate) aggregateSBOM(dir string, findings *terminal.ScanFindings) {
	path := filepath.Join(dir, "sbom.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	// v3.5 SBOM scanner output format
	var result struct {
		Summary struct {
			Generation *struct {
				TotalComponents int            `json:"total_components"`
				ByEcosystem     map[string]int `json:"by_ecosystem"`
			} `json:"generation"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return
	}

	if result.Summary.Generation != nil {
		findings.TotalPackages += result.Summary.Generation.TotalComponents
		for eco, count := range result.Summary.Generation.ByEcosystem {
			findings.PackagesByEco[eco] += count
		}
	}
}

func (h *Hydrate) aggregateVulns(dir string, findings *terminal.ScanFindings) {
	path := filepath.Join(dir, "package-analysis.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	// v3.5 package-analysis scanner output format
	var result struct {
		Summary struct {
			Vulns *struct {
				Critical int `json:"critical"`
				High     int `json:"high"`
				Medium   int `json:"medium"`
				Low      int `json:"low"`
			} `json:"vulns"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return
	}

	if result.Summary.Vulns != nil {
		findings.VulnCritical += result.Summary.Vulns.Critical
		findings.VulnHigh += result.Summary.Vulns.High
		findings.VulnMedium += result.Summary.Vulns.Medium
		findings.VulnLow += result.Summary.Vulns.Low
	}
}

func (h *Hydrate) aggregateLicenses(dir string, findings *terminal.ScanFindings) {
	path := filepath.Join(dir, "package-analysis.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	// v3.5 package-analysis scanner output format
	var result struct {
		Summary struct {
			Licenses *struct {
				UniqueLicenses int            `json:"unique_licenses"`
				LicenseCounts  map[string]int `json:"license_counts"`
			} `json:"licenses"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return
	}

	// Track unique license types
	if result.Summary.Licenses != nil {
		if result.Summary.Licenses.UniqueLicenses > findings.LicenseTypes {
			findings.LicenseTypes = result.Summary.Licenses.UniqueLicenses
		}
		for lic, count := range result.Summary.Licenses.LicenseCounts {
			findings.LicenseCounts[lic] += count
		}
	}
}

func (h *Hydrate) aggregateSecrets(dir string, findings *terminal.ScanFindings) {
	path := filepath.Join(dir, "code-security.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	// v3.5 code-security scanner output format
	var result struct {
		Summary struct {
			Secrets *struct {
				Critical      int `json:"critical"`
				High          int `json:"high"`
				Medium        int `json:"medium"`
				TotalFindings int `json:"total_findings"`
			} `json:"secrets"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return
	}

	if result.Summary.Secrets != nil {
		findings.SecretsCritical += result.Summary.Secrets.Critical
		findings.SecretsHigh += result.Summary.Secrets.High
		findings.SecretsMedium += result.Summary.Secrets.Medium
		findings.SecretsTotal += result.Summary.Secrets.TotalFindings
	}
}

func (h *Hydrate) aggregateMalcontent(dir string, findings *terminal.ScanFindings) {
	path := filepath.Join(dir, "package-analysis.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	// v3.5 package-analysis scanner output format
	var result struct {
		Summary struct {
			Malcontent *struct {
				Critical int `json:"critical"`
				High     int `json:"high"`
			} `json:"malcontent"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return
	}

	if result.Summary.Malcontent != nil {
		findings.MalcontentCrit += result.Summary.Malcontent.Critical
		findings.MalcontentHigh += result.Summary.Malcontent.High
	}
}

// Note: aggregateHealth was removed as the health scanner no longer exists.
// The code-quality scanner provides different metrics (tech debt, complexity, coverage).

func (h *Hydrate) aggregateTechID(dir string, findings *terminal.ScanFindings) {
	path := filepath.Join(dir, "tech-id.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	// tech-id scanner output format
	var result struct {
		Summary struct {
			Technology *struct {
				TotalTechnologies int            `json:"total_technologies"`
				ByCategory        map[string]int `json:"by_category"`
				TopTechnologies   []string       `json:"top_technologies"`
			} `json:"technology"`
			Models *struct {
				TotalModels int `json:"total_models"`
			} `json:"models"`
			Frameworks *struct {
				TotalFrameworks int `json:"total_frameworks"`
			} `json:"frameworks"`
			Security *struct {
				TotalFindings int `json:"total_findings"`
			} `json:"security"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return
	}

	if result.Summary.Technology != nil {
		findings.TechTotalTechs += result.Summary.Technology.TotalTechnologies
		// Merge category counts
		if findings.TechByCategory == nil {
			findings.TechByCategory = make(map[string]int)
		}
		for cat, count := range result.Summary.Technology.ByCategory {
			findings.TechByCategory[cat] += count
		}
		// Collect top technologies (deduplicate later)
		findings.TechTopList = append(findings.TechTopList, result.Summary.Technology.TopTechnologies...)
	}

	if result.Summary.Models != nil {
		findings.TechMLModels += result.Summary.Models.TotalModels
	}

	if result.Summary.Frameworks != nil {
		findings.TechMLFrameworks += result.Summary.Frameworks.TotalFrameworks
	}

	if result.Summary.Security != nil {
		findings.TechSecurityCount += result.Summary.Security.TotalFindings
	}
}

// preserveHistory saves the scan results to history for diff/delta tracking
func (h *Hydrate) preserveHistory(statuses []*RepoStatus, scannersRun []string, scanID string, startTime time.Time) {
	historyConfig := diff.DefaultHistoryConfig()
	historyMgr := diff.NewHistoryManager(h.zeroHome, historyConfig)

	for _, status := range statuses {
		if !status.ScanOK {
			continue
		}

		projectID := github.ProjectID(status.Repo.NameWithOwner)

		// Get commit info
		commitHash := h.getFullCommitHash(status.RepoPath)
		commitShort := h.getCommitHash(status.RepoPath)
		branch := h.getCurrentBranch(status.RepoPath)

		// Build findings summary from aggregated data
		findingsSummary := h.buildFindingsSummary(projectID)

		// Create scan record
		record := diff.ScanRecord{
			ScanID:          scanID,
			CommitHash:      commitHash,
			CommitShort:     commitShort,
			Branch:          branch,
			StartedAt:       startTime.Format(time.RFC3339),
			CompletedAt:     time.Now().Format(time.RFC3339),
			DurationSeconds: int(time.Since(startTime).Seconds()),
			Profile:         h.opts.Profile,
			ScannersRun:     scannersRun,
			Status:          "complete",
			FindingsSummary: findingsSummary,
		}

		// Preserve scan to history
		if err := historyMgr.PreserveScan(projectID, record); err != nil {
			// Log error but don't fail the scan
			fmt.Fprintf(os.Stderr, "Warning: failed to preserve history for %s: %v\n", projectID, err)
		}
	}
}

// getFullCommitHash returns the full commit hash of the repo
func (h *Hydrate) getFullCommitHash(path string) string {
	cmd := exec.Command("git", "-C", path, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

// getCurrentBranch returns the current branch name
func (h *Hydrate) getCurrentBranch(path string) string {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// buildFindingsSummary reads scan results and builds a findings summary
func (h *Hydrate) buildFindingsSummary(projectID string) diff.FindingsSummary {
	summary := diff.FindingsSummary{}
	analysisDir := filepath.Join(h.zeroHome, "repos", projectID, "analysis")

	// Read code-security results
	codeSecPath := filepath.Join(analysisDir, "code-security.json")
	if data, err := os.ReadFile(codeSecPath); err == nil {
		var result struct {
			Summary struct {
				Vulns   *struct{ Critical, High, Medium, Low int } `json:"vulns"`
				Secrets *struct{ Critical, High, Medium, Low int } `json:"secrets"`
				API     *struct{ Critical, High, Medium, Low int } `json:"api"`
			} `json:"summary"`
		}
		if json.Unmarshal(data, &result) == nil {
			if result.Summary.Vulns != nil {
				summary.Critical += result.Summary.Vulns.Critical
				summary.High += result.Summary.Vulns.High
				summary.Medium += result.Summary.Vulns.Medium
				summary.Low += result.Summary.Vulns.Low
			}
			if result.Summary.Secrets != nil {
				summary.Critical += result.Summary.Secrets.Critical
				summary.High += result.Summary.Secrets.High
				summary.Medium += result.Summary.Secrets.Medium
				summary.Low += result.Summary.Secrets.Low
			}
			if result.Summary.API != nil {
				summary.Critical += result.Summary.API.Critical
				summary.High += result.Summary.API.High
				summary.Medium += result.Summary.API.Medium
				summary.Low += result.Summary.API.Low
			}
		}
	}

	// Read package-analysis results
	pkgPath := filepath.Join(analysisDir, "package-analysis.json")
	if data, err := os.ReadFile(pkgPath); err == nil {
		var result struct {
			Summary struct {
				Vulns *struct{ Critical, High, Medium, Low int } `json:"vulns"`
			} `json:"summary"`
		}
		if json.Unmarshal(data, &result) == nil && result.Summary.Vulns != nil {
			summary.Critical += result.Summary.Vulns.Critical
			summary.High += result.Summary.Vulns.High
			summary.Medium += result.Summary.Vulns.Medium
			summary.Low += result.Summary.Vulns.Low
		}
	}

	// Read crypto results
	cryptoPath := filepath.Join(analysisDir, "crypto.json")
	if data, err := os.ReadFile(cryptoPath); err == nil {
		var result struct {
			Summary struct {
				Ciphers *struct{ BySeverity map[string]int } `json:"ciphers"`
				Keys    *struct{ BySeverity map[string]int } `json:"keys"`
			} `json:"summary"`
		}
		if json.Unmarshal(data, &result) == nil {
			if result.Summary.Ciphers != nil {
				summary.Critical += result.Summary.Ciphers.BySeverity["critical"]
				summary.High += result.Summary.Ciphers.BySeverity["high"]
				summary.Medium += result.Summary.Ciphers.BySeverity["medium"]
				summary.Low += result.Summary.Ciphers.BySeverity["low"]
			}
			if result.Summary.Keys != nil {
				summary.Critical += result.Summary.Keys.BySeverity["critical"]
				summary.High += result.Summary.Keys.BySeverity["high"]
				summary.Medium += result.Summary.Keys.BySeverity["medium"]
				summary.Low += result.Summary.Keys.BySeverity["low"]
			}
		}
	}

	summary.Total = summary.Critical + summary.High + summary.Medium + summary.Low
	return summary
}
