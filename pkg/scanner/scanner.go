// Package scanner manages running security scanners
package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
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

// Runner executes scanners
type Runner struct {
	BootstrapPath string
	ZeroHome      string
	Timeout       time.Duration
	Parallel      int
}

// NewRunner creates a new scanner runner
func NewRunner(zeroHome string) *Runner {
	// Find bootstrap.sh
	bootstrapPath := "utils/zero/scripts/bootstrap.sh"
	if _, err := os.Stat(bootstrapPath); err != nil {
		// Try relative to executable
		if exe, err := os.Executable(); err == nil {
			bootstrapPath = filepath.Join(filepath.Dir(exe), "..", "utils/zero/scripts/bootstrap.sh")
		}
	}

	return &Runner{
		BootstrapPath: bootstrapPath,
		ZeroHome:      zeroHome,
		Timeout:       10 * time.Minute,
		Parallel:      4,
	}
}

// RunResult holds the result of running all scanners on a repo
type RunResult struct {
	Success  bool
	Results  map[string]*Result
	Duration time.Duration
}

// Run executes all scanners for a repository
func (r *Runner) Run(ctx context.Context, repo, profile string, progress *Progress, skipScanners []string) (*RunResult, error) {
	start := time.Now()

	// Build skip scanners string
	skipStr := ""
	for _, s := range skipScanners {
		if skipStr != "" {
			skipStr += " "
		}
		skipStr += s
	}

	// Run bootstrap.sh with --scan-only
	args := []string{
		"--scan-only",
		"--" + profile,
		repo,
	}

	cmd := exec.CommandContext(ctx, r.BootstrapPath, args...)
	cmd.Env = append(os.Environ(),
		"SKIP_SCANNERS="+skipStr,
	)

	// Capture output
	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	if err != nil {
		return &RunResult{
			Success:  false,
			Duration: duration,
		}, fmt.Errorf("running scanners: %w\nOutput: %s", err, string(output))
	}

	// Parse results from analysis directory
	results := r.parseResults(repo, progress)

	return &RunResult{
		Success:  true,
		Results:  results,
		Duration: duration,
	}, nil
}

// parseResults reads scanner results from the analysis directory
func (r *Runner) parseResults(repo string, progress *Progress) map[string]*Result {
	projectID := filepath.Join(r.ZeroHome, "repos", repo, "analysis")

	progress.mu.RLock()
	results := make(map[string]*Result)
	for name, res := range progress.Results {
		results[name] = res

		// Try to read the JSON output
		jsonPath := filepath.Join(projectID, name+".json")
		if data, err := os.ReadFile(jsonPath); err == nil {
			res.Output = data
			res.Summary = parseSummary(name, data)
		}
	}
	progress.mu.RUnlock()

	return results
}

// parseSummary extracts a summary string from scanner JSON output
func parseSummary(scanner string, data []byte) string {
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return "complete"
	}

	// Try to get summary from common fields
	if summary, ok := result["summary"].(map[string]interface{}); ok {
		switch scanner {
		case "package-vulns":
			c := getInt(summary, "critical")
			h := getInt(summary, "high")
			m := getInt(summary, "medium")
			l := getInt(summary, "low")
			if c+h+m+l == 0 {
				return "no findings"
			}
			return fmt.Sprintf("%d critical, %d high, %d medium, %d low", c, h, m, l)

		case "package-sbom":
			if total, ok := summary["total_packages"].(float64); ok {
				return fmt.Sprintf("%.0f packages", total)
			}
			if components, ok := result["components"].([]interface{}); ok {
				return fmt.Sprintf("%d packages", len(components))
			}
		}
	}

	return "complete"
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return 0
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
