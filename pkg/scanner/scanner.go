// Package scanner manages running security scanners
package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"
	"time"
)

// Status represents scanner execution status
type Status string

const (
	StatusQueued   Status = "queued"
	StatusRunning  Status = "running"
	StatusComplete Status = "complete"
	StatusFailed   Status = "failed"
	StatusSkipped  Status = "skipped"
	StatusTimeout  Status = "timeout"
)

// Result holds the result of a scanner run
type Result struct {
	Scanner   string
	Status    Status
	Summary   string
	Duration  time.Duration
	Error     error
	Output    json.RawMessage
}

// Progress tracks scanner progress for a repo
type Progress struct {
	mu             sync.RWMutex
	Current        string
	CompletedCount int
	TotalCount     int
	Results        map[string]*Result
}

// NewProgress creates a new progress tracker
func NewProgress(scanners []string) *Progress {
	results := make(map[string]*Result)
	for _, s := range scanners {
		results[s] = &Result{
			Scanner: s,
			Status:  StatusQueued,
		}
	}
	return &Progress{
		TotalCount: len(scanners),
		Results:    results,
	}
}

// SetRunning marks a scanner as running
func (p *Progress) SetRunning(scanner string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Current = scanner
	if r, ok := p.Results[scanner]; ok {
		r.Status = StatusRunning
	}
}

// SetComplete marks a scanner as complete
func (p *Progress) SetComplete(scanner string, summary string, duration time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if r, ok := p.Results[scanner]; ok {
		r.Status = StatusComplete
		r.Summary = summary
		r.Duration = duration
	}
	p.CompletedCount++
	if p.Current == scanner {
		p.Current = ""
	}
}

// SetFailed marks a scanner as failed
func (p *Progress) SetFailed(scanner string, err error, duration time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if r, ok := p.Results[scanner]; ok {
		r.Status = StatusFailed
		r.Error = err
		r.Duration = duration
	}
	p.CompletedCount++
}

// SetSkipped marks a scanner as skipped
func (p *Progress) SetSkipped(scanner string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if r, ok := p.Results[scanner]; ok {
		r.Status = StatusSkipped
	}
	p.CompletedCount++
}

// GetProgress returns current progress info
func (p *Progress) GetProgress() (completed, total int, current string) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.CompletedCount, p.TotalCount, p.Current
}

// RunResult holds the result of running all scanners on a repo
type RunResult struct {
	Success  bool
	Results  map[string]*Result
	Duration time.Duration
}

// Runner executes scanners (wraps NativeRunner for backward compatibility)
type Runner struct {
	native *NativeRunner
}

// NewRunner creates a new scanner runner
func NewRunner(zeroHome string) *Runner {
	return &Runner{
		native: NewNativeRunner(zeroHome),
	}
}

// Run executes all scanners for a repository using native Go scanners
func (r *Runner) Run(ctx context.Context, repo, profile string, progress *Progress, skipScanners []string) (*RunResult, error) {
	// Get scanners for the profile from registry
	var scannersToRun []Scanner
	for name := range progress.Results {
		s, ok := Get(name)
		if !ok {
			progress.SetFailed(name, fmt.Errorf("scanner not found: %s", name), 0)
			continue
		}
		scannersToRun = append(scannersToRun, s)
	}

	// Set up paths
	repoPath := filepath.Join(r.native.ZeroHome, "repos", repo, "repo")
	outputDir := filepath.Join(r.native.ZeroHome, "repos", repo, "analysis")

	// Set up progress callback
	r.native.OnProgress = func(scanner string, status Status, summary string) {
		switch status {
		case StatusRunning:
			progress.SetRunning(scanner)
		case StatusComplete:
			progress.SetComplete(scanner, summary, 0)
		case StatusFailed:
			progress.SetFailed(scanner, fmt.Errorf("%s", summary), 0)
		case StatusSkipped:
			progress.SetSkipped(scanner)
		}
	}

	// Run using native runner
	return r.native.RunScanners(ctx, RunOptions{
		RepoPath:     repoPath,
		OutputDir:    outputDir,
		Scanners:     scannersToRun,
		SkipScanners: skipScanners,
	})
}

// EstimateTime returns estimated scan time in seconds based on file count
func EstimateTime(scanner string, fileCount int) int {
	switch scanner {
	case "package-malcontent":
		est := fileCount / 2000
		if est < 2 {
			return 2
		}
		return est
	case "package-sbom":
		return 3
	case "package-vulns":
		return 1
	case "package-health":
		return 2
	case "package-provenance":
		return 1
	case "code-vulns", "code-secrets":
		est := fileCount / 1000
		if est < 5 {
			return 5
		}
		return est
	default:
		return 2
	}
}

// TotalEstimate returns total estimated time for all scanners
func TotalEstimate(scanners []string, fileCount int) int {
	total := 0
	for _, s := range scanners {
		total += EstimateTime(s, fileCount)
	}
	return total
}
