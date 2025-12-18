// Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
// SPDX-License-Identifier: GPL-3.0

package diff

import (
	"encoding/json"
	"time"
)

// History represents the scan history index for a project
type History struct {
	ProjectID     string              `json:"project_id"`
	TotalScans    int                 `json:"total_scans"`
	FirstScanAt   string              `json:"first_scan_at"`
	LastScanAt    string              `json:"last_scan_at"`
	RetentionDays int                 `json:"retention_days"`
	Scans         []ScanRecord        `json:"scans"`
	ByCommit      map[string][]string `json:"by_commit"` // commit -> []scan_id
}

// ScanRecord represents a single scan in history
type ScanRecord struct {
	ScanID          string          `json:"scan_id"`
	CommitHash      string          `json:"commit_hash"`
	CommitShort     string          `json:"commit_short"`
	Branch          string          `json:"branch,omitempty"`
	StartedAt       string          `json:"started_at"`
	CompletedAt     string          `json:"completed_at"`
	DurationSeconds int             `json:"duration_seconds"`
	Profile         string          `json:"profile"`
	ScannersRun     []string        `json:"scanners_run"`
	Status          string          `json:"status"` // complete, failed, partial
	FindingsSummary FindingsSummary `json:"findings_summary"`
}

// FindingsSummary provides a quick overview of findings in a scan
type FindingsSummary struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
	Info     int `json:"info"`
	Total    int `json:"total"`
}

// FindingFingerprint uniquely identifies a finding for comparison
type FindingFingerprint struct {
	Scanner     string `json:"scanner"`      // e.g., "code-security/secrets"
	PrimaryKey  string `json:"primary_key"`  // Stable identifier (e.g., "CVE:package:ecosystem")
	LocationKey string `json:"location_key"` // File:Line:Column (for code findings)
	ContentHash string `json:"content_hash"` // Hash of relevant content for fuzzy matching
}

// ScanDelta represents the diff between two scans
type ScanDelta struct {
	BaselineScanID string    `json:"baseline_scan_id"`
	CompareScanID  string    `json:"compare_scan_id"`
	BaselineCommit string    `json:"baseline_commit"`
	CompareCommit  string    `json:"compare_commit"`
	GeneratedAt    time.Time `json:"generated_at"`

	Summary       DeltaSummary               `json:"summary"`
	ScannerDeltas map[string]ScannerDelta    `json:"scanner_deltas"`
}

// DeltaSummary provides an overview of changes between scans
type DeltaSummary struct {
	TotalNew       int    `json:"total_new"`
	TotalFixed     int    `json:"total_fixed"`
	TotalUnchanged int    `json:"total_unchanged"`
	TotalMoved     int    `json:"total_moved"`

	// By severity
	NewCritical   int `json:"new_critical"`
	NewHigh       int `json:"new_high"`
	NewMedium     int `json:"new_medium"`
	NewLow        int `json:"new_low"`
	FixedCritical int `json:"fixed_critical"`
	FixedHigh     int `json:"fixed_high"`
	FixedMedium   int `json:"fixed_medium"`
	FixedLow      int `json:"fixed_low"`

	// Net change
	NetChange int    `json:"net_change"` // new - fixed
	RiskTrend string `json:"risk_trend"` // improving, stable, degrading
}

// ScannerDelta represents changes for a specific scanner
type ScannerDelta struct {
	Scanner   string         `json:"scanner"`
	Feature   string         `json:"feature,omitempty"` // e.g., "secrets" for code-security
	New       []DeltaFinding `json:"new,omitempty"`
	Fixed     []DeltaFinding `json:"fixed,omitempty"`
	Moved     []MovedFinding `json:"moved,omitempty"`
	Unchanged int            `json:"unchanged"`
}

// DeltaFinding represents a finding that changed between scans
type DeltaFinding struct {
	Fingerprint FindingFingerprint `json:"fingerprint"`
	Finding     json.RawMessage    `json:"finding"` // Original finding data
	Severity    string             `json:"severity"`
	Scanner     string             `json:"scanner"`
	File        string             `json:"file,omitempty"`
	Line        int                `json:"line,omitempty"`
	Message     string             `json:"message,omitempty"`
}

// MovedFinding represents a finding that moved location between scans
type MovedFinding struct {
	Fingerprint FindingFingerprint `json:"fingerprint"`
	OldLocation string             `json:"old_location"`
	NewLocation string             `json:"new_location"`
	Finding     json.RawMessage    `json:"finding"`
	Severity    string             `json:"severity"`
}

// MatchStatus represents the result of comparing a finding across scans
type MatchStatus string

const (
	MatchExact   MatchStatus = "exact"   // Same location, same content
	MatchMoved   MatchStatus = "moved"   // Different location, same content
	MatchSimilar MatchStatus = "similar" // Same location, slightly different
	MatchNew     MatchStatus = "new"     // Only in new scan
	MatchFixed   MatchStatus = "fixed"   // Only in old scan
)

// MatchResult represents the result of matching a finding across scans
type MatchResult struct {
	Status      MatchStatus     `json:"status"`
	OldFinding  *DeltaFinding   `json:"old_finding,omitempty"`
	NewFinding  *DeltaFinding   `json:"new_finding,omitempty"`
	Confidence  float64         `json:"confidence"` // 0.0 - 1.0
	Explanation string          `json:"explanation,omitempty"`
}

// DiffOptions configures how diff comparison is performed
type DiffOptions struct {
	Scanner        string   // Filter to specific scanner (empty = all)
	Scanners       []string // Filter to multiple scanners
	Severities     []string // Filter by severity
	FuzzyMatch     bool     // Enable fuzzy matching (default: true)
	LineTolerance  int      // Line tolerance for fuzzy matching (default: 5)
	ShowNewOnly    bool     // Only show new findings
	ShowFixedOnly  bool     // Only show fixed findings
	IncludeMoved   bool     // Include moved findings in output
	OutputFormat   string   // table, json, summary
}

// DefaultDiffOptions returns sensible defaults
func DefaultDiffOptions() DiffOptions {
	return DiffOptions{
		FuzzyMatch:    true,
		LineTolerance: 5,
		IncludeMoved:  true,
		OutputFormat:  "table",
	}
}

// HistoryConfig configures scan history retention
type HistoryConfig struct {
	Enabled       bool `json:"enabled"`
	RetentionDays int  `json:"retention_days"`
	MaxScans      int  `json:"max_scans"`
	CompressOld   bool `json:"compress_old"` // Gzip scans older than 7 days
}

// DefaultHistoryConfig returns default history configuration
func DefaultHistoryConfig() HistoryConfig {
	return HistoryConfig{
		Enabled:       true,
		RetentionDays: 90,
		MaxScans:      50,
		CompressOld:   true,
	}
}
