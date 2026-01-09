package freshness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const metadataFilename = "freshness.json"

// Manager handles freshness metadata operations
type Manager struct {
	zeroHome   string
	thresholds Thresholds
}

// NewManager creates a new freshness manager
func NewManager(zeroHome string) *Manager {
	return &Manager{
		zeroHome:   zeroHome,
		thresholds: DefaultThresholds(),
	}
}

// NewManagerWithThresholds creates a manager with custom thresholds
func NewManagerWithThresholds(zeroHome string, thresholds Thresholds) *Manager {
	return &Manager{
		zeroHome:   zeroHome,
		thresholds: thresholds,
	}
}

// metadataPath returns the path to the freshness metadata file for a repo
func (m *Manager) metadataPath(repo string) string {
	return filepath.Join(m.zeroHome, "repos", repo, metadataFilename)
}

// Load loads freshness metadata for a repository
func (m *Manager) Load(repo string) (*Metadata, error) {
	path := m.metadataPath(repo)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty metadata for repos that haven't been scanned
			return &Metadata{
				Repository:    repo,
				ScannerStatus: make(map[string]Status),
			}, nil
		}
		return nil, fmt.Errorf("reading freshness metadata: %w", err)
	}

	var meta Metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("parsing freshness metadata: %w", err)
	}

	return &meta, nil
}

// Save saves freshness metadata for a repository
func (m *Manager) Save(meta *Metadata) error {
	path := m.metadataPath(meta.Repository)

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling metadata: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing metadata: %w", err)
	}

	return nil
}

// Check checks the freshness of a repository
func (m *Manager) Check(repo string) (*CheckResult, error) {
	meta, err := m.Load(repo)
	if err != nil {
		return nil, err
	}

	// Get current commit if possible
	currentCommit := m.getCurrentCommit(repo)

	result := NewCheckResult(meta, m.thresholds, currentCommit)
	return &result, nil
}

// getCurrentCommit gets the current HEAD commit of a repo
func (m *Manager) getCurrentCommit(repo string) string {
	headPath := filepath.Join(m.zeroHome, "repos", repo, "repo", ".git", "HEAD")
	data, err := os.ReadFile(headPath)
	if err != nil {
		return ""
	}
	// This is simplified - in reality we'd need to resolve refs
	return string(data)
}

// RecordScan records a successful scan
func (m *Manager) RecordScan(repo string, scanners []ScanResult) error {
	meta, err := m.Load(repo)
	if err != nil {
		return err
	}

	meta.LastScan = time.Now()

	if meta.ScannerStatus == nil {
		meta.ScannerStatus = make(map[string]Status)
	}

	for _, s := range scanners {
		meta.ScannerStatus[s.Name] = Status{
			Scanner:      s.Name,
			LastRun:      time.Now(),
			Duration:     s.Duration.String(),
			Success:      s.Success,
			OutputFile:   s.OutputFile,
			FindingCount: s.FindingCount,
			Error:        s.Error,
		}
	}

	return m.Save(meta)
}

// ScanResult holds the result of a single scanner run
type ScanResult struct {
	Name         string
	Success      bool
	Duration     time.Duration
	OutputFile   string
	FindingCount int
	Error        string
}

// ListAll lists all repositories with their freshness status
func (m *Manager) ListAll() ([]CheckResult, error) {
	reposDir := filepath.Join(m.zeroHome, "repos")

	entries, err := os.ReadDir(reposDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading repos directory: %w", err)
	}

	var results []CheckResult
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		repo := entry.Name()
		result, err := m.Check(repo)
		if err != nil {
			continue // Skip repos with errors
		}

		results = append(results, *result)
	}

	// Sort by age (most recent first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].LastScan.After(results[j].LastScan)
	})

	return results, nil
}

// ListStale lists repositories that need refreshing
func (m *Manager) ListStale() ([]CheckResult, error) {
	all, err := m.ListAll()
	if err != nil {
		return nil, err
	}

	var stale []CheckResult
	for _, r := range all {
		if r.NeedsRefresh {
			stale = append(stale, r)
		}
	}

	return stale, nil
}

// ListByLevel lists repositories at a specific freshness level
func (m *Manager) ListByLevel(level Level) ([]CheckResult, error) {
	all, err := m.ListAll()
	if err != nil {
		return nil, err
	}

	var filtered []CheckResult
	for _, r := range all {
		if r.Level == level {
			filtered = append(filtered, r)
		}
	}

	return filtered, nil
}

// ShouldScan returns true if the repository should be scanned
// based on freshness and optional commit check
func (m *Manager) ShouldScan(repo string, checkCommit bool) (bool, string, error) {
	result, err := m.Check(repo)
	if err != nil {
		return true, "error checking freshness", err
	}

	// Never scanned
	if result.LastScan.IsZero() {
		return true, "never scanned", nil
	}

	// Check if commits have changed
	if checkCommit && result.CommitChanged {
		return true, "new commits detected", nil
	}

	// Check freshness level
	if result.NeedsRefresh {
		return true, fmt.Sprintf("data is %s", result.Level), nil
	}

	return false, "data is fresh", nil
}

// Delete removes freshness metadata for a repository
func (m *Manager) Delete(repo string) error {
	path := m.metadataPath(repo)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing metadata: %w", err)
	}
	return nil
}

// GetThresholds returns the current thresholds
func (m *Manager) GetThresholds() Thresholds {
	return m.thresholds
}

// SetThresholds updates the thresholds
func (m *Manager) SetThresholds(t Thresholds) {
	m.thresholds = t
}

// Summary returns a summary of all repositories by freshness level
type Summary struct {
	Total     int            `json:"total"`
	ByLevel   map[Level]int  `json:"by_level"`
	NeedScan  int            `json:"need_scan"`
	LastCheck time.Time      `json:"last_check"`
}

// GetSummary returns a summary of repository freshness
func (m *Manager) GetSummary() (*Summary, error) {
	all, err := m.ListAll()
	if err != nil {
		return nil, err
	}

	summary := &Summary{
		Total:     len(all),
		ByLevel:   make(map[Level]int),
		LastCheck: time.Now(),
	}

	for _, r := range all {
		summary.ByLevel[r.Level]++
		if r.NeedsRefresh {
			summary.NeedScan++
		}
	}

	return summary, nil
}
