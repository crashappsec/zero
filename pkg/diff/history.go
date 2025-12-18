// Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
// SPDX-License-Identifier: GPL-3.0

package diff

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// HistoryManager handles scan history operations
type HistoryManager struct {
	zeroHome string
	config   HistoryConfig
}

// NewHistoryManager creates a new history manager
func NewHistoryManager(zeroHome string, config HistoryConfig) *HistoryManager {
	return &HistoryManager{
		zeroHome: zeroHome,
		config:   config,
	}
}

// GetHistoryDir returns the history directory for a project
func (m *HistoryManager) GetHistoryDir(projectID string) string {
	return filepath.Join(m.zeroHome, "repos", projectID, "history")
}

// GetScansDir returns the scans directory for a project
func (m *HistoryManager) GetScansDir(projectID string) string {
	return filepath.Join(m.GetHistoryDir(projectID), "scans")
}

// GetHistoryFile returns the history.json path for a project
func (m *HistoryManager) GetHistoryFile(projectID string) string {
	return filepath.Join(m.GetHistoryDir(projectID), "history.json")
}

// GetAnalysisDir returns the current analysis directory for a project
func (m *HistoryManager) GetAnalysisDir(projectID string) string {
	return filepath.Join(m.zeroHome, "repos", projectID, "analysis")
}

// LoadHistory loads the history for a project
func (m *HistoryManager) LoadHistory(projectID string) (*History, error) {
	historyFile := m.GetHistoryFile(projectID)

	data, err := os.ReadFile(historyFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty history
			return &History{
				ProjectID:     projectID,
				RetentionDays: m.config.RetentionDays,
				Scans:         []ScanRecord{},
				ByCommit:      make(map[string][]string),
			}, nil
		}
		return nil, fmt.Errorf("failed to read history: %w", err)
	}

	var history History
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, fmt.Errorf("failed to parse history: %w", err)
	}

	return &history, nil
}

// SaveHistory saves the history for a project
func (m *HistoryManager) SaveHistory(projectID string, history *History) error {
	historyDir := m.GetHistoryDir(projectID)
	if err := os.MkdirAll(historyDir, 0755); err != nil {
		return fmt.Errorf("failed to create history directory: %w", err)
	}

	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	historyFile := m.GetHistoryFile(projectID)
	if err := os.WriteFile(historyFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write history: %w", err)
	}

	return nil
}

// GenerateScanID creates a unique scan ID from timestamp and commit
func GenerateScanID(commitShort string) string {
	timestamp := time.Now().Format("20060102-150405")
	if commitShort == "" {
		commitShort = "unknown"
	}
	return fmt.Sprintf("%s-%s", timestamp, commitShort)
}

// PreserveScan copies the current analysis to history and updates the index
func (m *HistoryManager) PreserveScan(projectID string, record ScanRecord) error {
	if !m.config.Enabled {
		return nil
	}

	// Create scan directory
	scanDir := filepath.Join(m.GetScansDir(projectID), record.ScanID)
	if err := os.MkdirAll(scanDir, 0755); err != nil {
		return fmt.Errorf("failed to create scan directory: %w", err)
	}

	// Copy analysis files to scan directory
	analysisDir := m.GetAnalysisDir(projectID)
	entries, err := os.ReadDir(analysisDir)
	if err != nil {
		return fmt.Errorf("failed to read analysis directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		// Copy JSON files (skip history.json if it exists there)
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		if entry.Name() == "history.json" {
			continue
		}

		srcPath := filepath.Join(analysisDir, entry.Name())
		dstPath := filepath.Join(scanDir, entry.Name())

		if err := copyFile(srcPath, dstPath); err != nil {
			return fmt.Errorf("failed to copy %s: %w", entry.Name(), err)
		}
	}

	// Load and update history
	history, err := m.LoadHistory(projectID)
	if err != nil {
		return fmt.Errorf("failed to load history: %w", err)
	}

	// Add new record at the beginning (most recent first)
	history.Scans = append([]ScanRecord{record}, history.Scans...)
	history.TotalScans = len(history.Scans)
	history.LastScanAt = record.CompletedAt

	if history.FirstScanAt == "" {
		history.FirstScanAt = record.CompletedAt
	}

	// Update commit index
	if history.ByCommit == nil {
		history.ByCommit = make(map[string][]string)
	}
	history.ByCommit[record.CommitHash] = append(history.ByCommit[record.CommitHash], record.ScanID)

	// Prune old scans
	if err := m.pruneHistory(projectID, history); err != nil {
		// Log but don't fail
		fmt.Fprintf(os.Stderr, "Warning: failed to prune history: %v\n", err)
	}

	// Save updated history
	if err := m.SaveHistory(projectID, history); err != nil {
		return fmt.Errorf("failed to save history: %w", err)
	}

	return nil
}

// pruneHistory removes old scans based on retention config
func (m *HistoryManager) pruneHistory(projectID string, history *History) error {
	if len(history.Scans) <= m.config.MaxScans {
		return nil
	}

	// Keep only MaxScans
	scansToRemove := history.Scans[m.config.MaxScans:]
	history.Scans = history.Scans[:m.config.MaxScans]

	// Remove scan directories
	for _, scan := range scansToRemove {
		scanDir := filepath.Join(m.GetScansDir(projectID), scan.ScanID)
		if err := os.RemoveAll(scanDir); err != nil {
			// Log but continue
			fmt.Fprintf(os.Stderr, "Warning: failed to remove scan %s: %v\n", scan.ScanID, err)
		}

		// Update commit index
		if commits, ok := history.ByCommit[scan.CommitHash]; ok {
			var newCommits []string
			for _, id := range commits {
				if id != scan.ScanID {
					newCommits = append(newCommits, id)
				}
			}
			if len(newCommits) == 0 {
				delete(history.ByCommit, scan.CommitHash)
			} else {
				history.ByCommit[scan.CommitHash] = newCommits
			}
		}
	}

	history.TotalScans = len(history.Scans)
	return nil
}

// GetScan retrieves a specific scan by ID
func (m *HistoryManager) GetScan(projectID, scanID string) (*ScanRecord, error) {
	history, err := m.LoadHistory(projectID)
	if err != nil {
		return nil, err
	}

	for _, scan := range history.Scans {
		if scan.ScanID == scanID {
			return &scan, nil
		}
	}

	return nil, fmt.Errorf("scan not found: %s", scanID)
}

// GetLatestScan returns the most recent scan
func (m *HistoryManager) GetLatestScan(projectID string) (*ScanRecord, error) {
	history, err := m.LoadHistory(projectID)
	if err != nil {
		return nil, err
	}

	if len(history.Scans) == 0 {
		return nil, fmt.Errorf("no scans found for %s", projectID)
	}

	return &history.Scans[0], nil
}

// GetScanByOffset returns a scan by offset from latest (0 = latest, 1 = previous, etc.)
func (m *HistoryManager) GetScanByOffset(projectID string, offset int) (*ScanRecord, error) {
	history, err := m.LoadHistory(projectID)
	if err != nil {
		return nil, err
	}

	if offset < 0 || offset >= len(history.Scans) {
		return nil, fmt.Errorf("scan offset %d out of range (have %d scans)", offset, len(history.Scans))
	}

	return &history.Scans[offset], nil
}

// GetScanByCommit returns scans for a specific commit
func (m *HistoryManager) GetScanByCommit(projectID, commitPrefix string) (*ScanRecord, error) {
	history, err := m.LoadHistory(projectID)
	if err != nil {
		return nil, err
	}

	// Search by prefix
	for _, scan := range history.Scans {
		if strings.HasPrefix(scan.CommitHash, commitPrefix) ||
			strings.HasPrefix(scan.CommitShort, commitPrefix) {
			return &scan, nil
		}
	}

	return nil, fmt.Errorf("no scan found for commit %s", commitPrefix)
}

// LoadScanResults loads the scanner results for a specific scan
func (m *HistoryManager) LoadScanResults(projectID, scanID string) (map[string]json.RawMessage, error) {
	scanDir := filepath.Join(m.GetScansDir(projectID), scanID)

	entries, err := os.ReadDir(scanDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("scan not found: %s", scanID)
		}
		return nil, fmt.Errorf("failed to read scan directory: %w", err)
	}

	results := make(map[string]json.RawMessage)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(scanDir, entry.Name())
		data, err := readPossiblyCompressed(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", entry.Name(), err)
		}

		// Use filename without extension as key
		key := strings.TrimSuffix(entry.Name(), ".json")
		results[key] = data
	}

	return results, nil
}

// ListScans returns all scans for a project
func (m *HistoryManager) ListScans(projectID string, limit int) ([]ScanRecord, error) {
	history, err := m.LoadHistory(projectID)
	if err != nil {
		return nil, err
	}

	if limit <= 0 || limit > len(history.Scans) {
		return history.Scans, nil
	}

	return history.Scans[:limit], nil
}

// ResolveScanRef resolves a scan reference to a scan ID
// Supports: scan ID, commit hash/prefix, "latest", "latest~N"
func (m *HistoryManager) ResolveScanRef(projectID, ref string) (*ScanRecord, error) {
	if ref == "" || ref == "latest" {
		return m.GetLatestScan(projectID)
	}

	// Handle "latest~N" syntax
	if strings.HasPrefix(ref, "latest~") || strings.HasPrefix(ref, "HEAD~") {
		offsetStr := strings.TrimPrefix(strings.TrimPrefix(ref, "latest~"), "HEAD~")
		offset := 0
		if _, err := fmt.Sscanf(offsetStr, "%d", &offset); err != nil {
			return nil, fmt.Errorf("invalid offset: %s", ref)
		}
		return m.GetScanByOffset(projectID, offset)
	}

	// Try as scan ID first
	if scan, err := m.GetScan(projectID, ref); err == nil {
		return scan, nil
	}

	// Try as commit hash
	return m.GetScanByCommit(projectID, ref)
}

// GetScanFiles returns the list of scanner output files in a scan
func (m *HistoryManager) GetScanFiles(projectID, scanID string) ([]string, error) {
	scanDir := filepath.Join(m.GetScansDir(projectID), scanID)

	entries, err := os.ReadDir(scanDir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			files = append(files, strings.TrimSuffix(entry.Name(), ".json"))
		}
	}

	sort.Strings(files)
	return files, nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// readPossiblyCompressed reads a file that may be gzip compressed
func readPossiblyCompressed(path string) ([]byte, error) {
	// Try .gz version first
	gzPath := path + ".gz"
	if _, err := os.Stat(gzPath); err == nil {
		f, err := os.Open(gzPath)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		gz, err := gzip.NewReader(f)
		if err != nil {
			return nil, err
		}
		defer gz.Close()

		return io.ReadAll(gz)
	}

	// Fall back to uncompressed
	return os.ReadFile(path)
}
