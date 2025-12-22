package automation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/crashappsec/zero/pkg/workflow/freshness"
)

// RefreshConfig configures the refresh behavior
type RefreshConfig struct {
	// Scanners to run during refresh
	Scanners []string `json:"scanners"`

	// Whether to force refresh even if data is fresh
	Force bool `json:"force"`

	// Whether to check for new commits before refreshing
	CheckCommits bool `json:"check_commits"`

	// Parallel execution limit (0 = no limit)
	Parallelism int `json:"parallelism"`

	// Timeout for the entire refresh operation
	Timeout time.Duration `json:"timeout"`
}

// DefaultRefreshConfig returns default refresh configuration
func DefaultRefreshConfig() RefreshConfig {
	return RefreshConfig{
		Scanners:     []string{"sbom", "package-analysis", "code-security"},
		Force:        false,
		CheckCommits: true,
		Parallelism:  4,
		Timeout:      30 * time.Minute,
	}
}

// RefreshResult holds the result of a refresh operation
type RefreshResult struct {
	Repository string        `json:"repository"`
	Success    bool          `json:"success"`
	Skipped    bool          `json:"skipped"`
	Reason     string        `json:"reason,omitempty"`
	Scanners   []ScannerRun  `json:"scanners"`
	Duration   time.Duration `json:"duration"`
	Error      string        `json:"error,omitempty"`
}

// ScannerRun holds the result of running a single scanner
type ScannerRun struct {
	Name         string        `json:"name"`
	Success      bool          `json:"success"`
	Duration     time.Duration `json:"duration"`
	FindingCount int           `json:"finding_count"`
	Error        string        `json:"error,omitempty"`
}

// RefreshManager handles refresh operations
type RefreshManager struct {
	zeroHome         string
	config           RefreshConfig
	freshnessManager *freshness.Manager
	scannerFunc      ScannerFunc
}

// ScannerFunc is the function signature for running a scanner
type ScannerFunc func(ctx context.Context, repo, scanner string) (*ScannerRun, error)

// NewRefreshManager creates a new refresh manager
func NewRefreshManager(zeroHome string, scannerFunc ScannerFunc) *RefreshManager {
	return &RefreshManager{
		zeroHome:         zeroHome,
		config:           DefaultRefreshConfig(),
		freshnessManager: freshness.NewManager(zeroHome),
		scannerFunc:      scannerFunc,
	}
}

// NewRefreshManagerWithConfig creates a manager with custom config
func NewRefreshManagerWithConfig(zeroHome string, config RefreshConfig, scannerFunc ScannerFunc) *RefreshManager {
	m := NewRefreshManager(zeroHome, scannerFunc)
	m.config = config
	return m
}

// RefreshRepo refreshes a single repository
func (m *RefreshManager) RefreshRepo(ctx context.Context, repo string) (*RefreshResult, error) {
	start := time.Now()
	result := &RefreshResult{
		Repository: repo,
		Scanners:   make([]ScannerRun, 0),
	}

	// Check if refresh is needed
	if !m.config.Force {
		shouldScan, reason, err := m.freshnessManager.ShouldScan(repo, m.config.CheckCommits)
		if err != nil {
			result.Error = err.Error()
			return result, err
		}

		if !shouldScan {
			result.Skipped = true
			result.Reason = reason
			result.Success = true
			result.Duration = time.Since(start)
			return result, nil
		}

		result.Reason = reason
	}

	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, m.config.Timeout)
	defer cancel()

	// Run scanners
	var wg sync.WaitGroup
	var mu sync.Mutex
	semaphore := make(chan struct{}, m.config.Parallelism)
	if m.config.Parallelism == 0 {
		semaphore = make(chan struct{}, len(m.config.Scanners))
	}

	for _, scanner := range m.config.Scanners {
		wg.Add(1)
		go func(scannerName string) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			run, err := m.scannerFunc(ctx, repo, scannerName)
			if err != nil {
				run = &ScannerRun{
					Name:    scannerName,
					Success: false,
					Error:   err.Error(),
				}
			}

			mu.Lock()
			result.Scanners = append(result.Scanners, *run)
			mu.Unlock()
		}(scanner)
	}

	wg.Wait()

	// Check overall success
	result.Success = true
	for _, run := range result.Scanners {
		if !run.Success {
			result.Success = false
			break
		}
	}

	result.Duration = time.Since(start)

	// Record the scan in freshness metadata
	if result.Success {
		scanResults := make([]freshness.ScanResult, len(result.Scanners))
		for i, run := range result.Scanners {
			scanResults[i] = freshness.ScanResult{
				Name:         run.Name,
				Success:      run.Success,
				Duration:     run.Duration,
				FindingCount: run.FindingCount,
				Error:        run.Error,
			}
		}
		m.freshnessManager.RecordScan(repo, scanResults)
	}

	return result, nil
}

// RefreshAll refreshes all repositories that need it
func (m *RefreshManager) RefreshAll(ctx context.Context) ([]RefreshResult, error) {
	// Get list of stale repositories
	stale, err := m.freshnessManager.ListStale()
	if err != nil {
		return nil, fmt.Errorf("listing stale repos: %w", err)
	}

	var results []RefreshResult

	for _, check := range stale {
		result, err := m.RefreshRepo(ctx, check.Repository)
		if err != nil {
			results = append(results, RefreshResult{
				Repository: check.Repository,
				Success:    false,
				Error:      err.Error(),
			})
			continue
		}
		results = append(results, *result)
	}

	return results, nil
}

// RefreshStale refreshes only stale repositories
func (m *RefreshManager) RefreshStale(ctx context.Context) ([]RefreshResult, error) {
	return m.RefreshAll(ctx)
}

// GetStatus returns the refresh status for all repositories
func (m *RefreshManager) GetStatus() ([]RefreshStatus, error) {
	all, err := m.freshnessManager.ListAll()
	if err != nil {
		return nil, err
	}

	var statuses []RefreshStatus
	for _, check := range all {
		statuses = append(statuses, RefreshStatus{
			Repository:   check.Repository,
			Level:        string(check.Level),
			LastScan:     check.LastScan,
			Age:          check.AgeString,
			NeedsRefresh: check.NeedsRefresh,
			Scanners:     check.Scanners,
			Successful:   check.Successful,
		})
	}

	return statuses, nil
}

// RefreshStatus holds the refresh status of a repository
type RefreshStatus struct {
	Repository   string    `json:"repository"`
	Level        string    `json:"level"`
	LastScan     time.Time `json:"last_scan"`
	Age          string    `json:"age"`
	NeedsRefresh bool      `json:"needs_refresh"`
	Scanners     int       `json:"scanners"`
	Successful   int       `json:"successful"`
}

// GetConfig returns the current configuration
func (m *RefreshManager) GetConfig() RefreshConfig {
	return m.config
}

// SetConfig updates the configuration
func (m *RefreshManager) SetConfig(config RefreshConfig) {
	m.config = config
}
