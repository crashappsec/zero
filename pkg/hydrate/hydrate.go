// Package hydrate implements the hydrate command for cloning and scanning repos
package hydrate

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/crashappsec/zero/pkg/config"
	"github.com/crashappsec/zero/pkg/github"
	"github.com/crashappsec/zero/pkg/scanner"
	"github.com/crashappsec/zero/pkg/terminal"
)

// Options configures the hydrate command
type Options struct {
	Org          string
	Limit        int
	Profile      string
	Force        bool
	SkipSlow     bool
	Yes          bool
	Parallel     int
	SkipScanners []string
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
	runner   *scanner.Runner
	opts     *Options
	zeroHome string
}

// New creates a new Hydrate instance
func New(opts *Options) (*Hydrate, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	zeroHome := cfg.Settings.ZeroHome
	if zeroHome == "" {
		zeroHome = ".zero"
	}

	if opts.Parallel == 0 {
		opts.Parallel = cfg.Settings.ParallelJobs
	}

	return &Hydrate{
		cfg:      cfg,
		term:     terminal.New(),
		gh:       github.NewClient(),
		runner:   scanner.NewRunner(zeroHome),
		opts:     opts,
		zeroHome: zeroHome,
	}, nil
}

// Run executes the hydrate process
func (h *Hydrate) Run(ctx context.Context) error {
	start := time.Now()
	scanID := fmt.Sprintf("scan-%s", time.Now().Format("20060102-150405"))

	// Fetch repositories
	h.term.Info("Fetching repositories for %s...", h.term.Color(terminal.Cyan, h.opts.Org))

	repos, err := h.gh.ListOrgRepos(h.opts.Org, h.opts.Limit)
	if err != nil {
		return fmt.Errorf("listing repos: %w", err)
	}

	if len(repos) == 0 {
		return fmt.Errorf("no repositories found for org: %s", h.opts.Org)
	}

	// Get scanners for profile
	scanners, err := h.cfg.GetProfileScanners(h.opts.Profile)
	if err != nil {
		return fmt.Errorf("getting scanners: %w", err)
	}

	// Print header
	h.term.Divider()
	h.term.Info("%s %s", h.term.Color(terminal.Bold, "Hydrate Organization:"), h.term.Color(terminal.Cyan, h.opts.Org))
	h.term.Info("Scan ID:      %s", h.term.Color(terminal.Dim, scanID))
	h.term.Info("Repositories: %s", h.term.Color(terminal.Cyan, strconv.Itoa(len(repos))))
	h.term.Info("Profile:      %s", h.term.Color(terminal.Cyan, h.opts.Profile))
	h.term.Info("Parallel:     %s", h.term.Color(terminal.Cyan, fmt.Sprintf("%d jobs", h.opts.Parallel)))

	// Phase 1: Clone
	h.term.Header("CLONING")
	repoStatuses, err := h.cloneRepos(ctx, repos)
	if err != nil {
		return err
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

	// Print summary
	duration := int(time.Since(start).Seconds())
	diskUsage := h.getDiskUsage()
	totalFiles := h.getTotalFiles(repoStatuses)

	h.term.Divider()
	h.term.Summary(h.opts.Org, duration, successCount, failedCount, diskUsage, formatNumber(totalFiles))

	return nil
}

// cloneRepos clones all repositories in parallel
func (h *Hydrate) cloneRepos(ctx context.Context, repos []github.Repository) ([]*RepoStatus, error) {
	statuses := make([]*RepoStatus, len(repos))
	var wg sync.WaitGroup
	sem := make(chan struct{}, h.opts.Parallel)

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

	// Check if already cloned
	if _, err := os.Stat(repoPath); err == nil {
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

	// Create directory
	if err := os.MkdirAll(filepath.Dir(repoPath), 0755); err != nil {
		h.term.Error("%s clone failed: %v", repo.Name, err)
		return
	}

	// Clone with depth=1
	cmd := exec.CommandContext(ctx, "git", "clone",
		"--depth", "1",
		repo.SSHURL,
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

// scanRepos scans all repositories
func (h *Hydrate) scanRepos(ctx context.Context, statuses []*RepoStatus, scanners, skipScanners []string) (success, failed int) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, h.opts.Parallel)

	// Track overall progress
	var totalComplete atomic.Int32
	totalScanners := 0
	for _, s := range statuses {
		if s.CloneOK {
			totalScanners += len(scanners)
		}
	}

	// Progress update goroutine
	progressDone := make(chan struct{})
	go func() {
		ticker := time.NewTicker(300 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-progressDone:
				return
			case <-ticker.C:
				completed := int(totalComplete.Load())
				if completed > 0 && completed < totalScanners {
					active := h.getActiveScannersString(statuses)
					h.term.Progress(completed, totalScanners, active)
				}
			}
		}
	}()

	// Print initial status for all repos
	for _, status := range statuses {
		if !status.CloneOK {
			continue
		}

		status.Progress = scanner.NewProgress(scanners)
		estimate := scanner.TotalEstimate(scanners, status.FileCount)
		h.term.RepoScanning(status.Repo.Name, estimate)

		// Print scanner list
		firstActive := true
		for _, s := range scanners {
			est := scanner.EstimateTime(s, status.FileCount)
			if contains(skipScanners, s) {
				h.term.ScannerSkipped(s)
				status.Progress.SetSkipped(s)
			} else if firstActive {
				h.term.ScannerRunning(s, est)
				firstActive = false
			} else {
				h.term.ScannerQueued(s, est)
			}
		}
	}

	// Run scans
	results := make(chan *RepoStatus, len(statuses))

	for _, status := range statuses {
		if !status.CloneOK {
			failed++
			continue
		}

		wg.Add(1)
		go func(s *RepoStatus) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			h.scanRepo(ctx, s, scanners, skipScanners, &totalComplete)
			results <- s
		}(status)
	}

	// Collect results
	go func() {
		wg.Wait()
		close(results)
		close(progressDone)
	}()

	for status := range results {
		h.term.ClearLine()
		h.printRepoResult(status, scanners, skipScanners)
		if status.ScanOK {
			success++
		} else {
			failed++
		}
	}

	return success, failed
}

// scanRepo runs all scanners on a single repo
func (h *Hydrate) scanRepo(ctx context.Context, status *RepoStatus, scanners, skipScanners []string, totalComplete *atomic.Int32) {
	start := time.Now()

	result, err := h.runner.Run(ctx, status.Repo.NameWithOwner, h.opts.Profile, status.Progress, skipScanners)
	status.Duration = time.Since(start)

	if err != nil {
		status.ScanOK = false
		return
	}

	status.ScanOK = result.Success

	// Update total progress
	totalComplete.Add(int32(len(scanners) - len(skipScanners)))
}

// printRepoResult prints the results for a completed repo
func (h *Hydrate) printRepoResult(status *RepoStatus, scanners, skipScanners []string) {
	h.term.RepoComplete(status.Repo.Name, status.ScanOK)

	if !status.ScanOK {
		return
	}

	// Print each scanner result
	for _, s := range scanners {
		if contains(skipScanners, s) {
			h.term.ScannerSkipped(s)
			continue
		}

		if r, ok := status.Progress.Results[s]; ok {
			duration := int(r.Duration.Seconds())
			h.term.ScannerComplete(s, r.Summary, duration)
		}
	}
	fmt.Println()
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

	h.term.Warning("Slow scanner warning")
	h.term.Info("    Largest repo: %s (%s files)", h.term.Color(terminal.Cyan, largest.Repo.Name), formatNumber(largest.FileCount))
	fmt.Println()

	// Show slow scanners
	slowScanners := []string{}
	for _, s := range scanners {
		est := scanner.EstimateTime(s, largest.FileCount)
		if est > 10 {
			h.term.Info("    • %s: ~%ds on %s", s, est, largest.Repo.Name)
			slowScanners = append(slowScanners, s)
		}
	}

	// For now, return the slow scanners to skip (in full implementation, would prompt user)
	if h.opts.SkipSlow {
		return slowScanners
	}

	return nil
}

// getActiveScannersString returns a string of currently active scanners
func (h *Hydrate) getActiveScannersString(statuses []*RepoStatus) string {
	var active []string
	for _, s := range statuses {
		if s.Progress != nil {
			_, _, current := s.Progress.GetProgress()
			if current != "" {
				active = append(active, fmt.Sprintf("%s:%s", s.Repo.Name, current))
			}
		}
	}
	result := strings.Join(active, ", ")
	if len(result) > 60 {
		result = result[:60] + "..."
	}
	return result
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
