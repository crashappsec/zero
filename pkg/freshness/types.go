// Package freshness provides staleness detection and freshness metadata tracking
package freshness

import (
	"time"
)

// Level represents the freshness level of scan data
type Level string

const (
	LevelFresh     Level = "fresh"
	LevelStale     Level = "stale"
	LevelVeryStale Level = "very-stale"
	LevelExpired   Level = "expired"
	LevelUnknown   Level = "unknown"
)

// Metadata holds freshness information for a scanned repository
type Metadata struct {
	Repository    string            `json:"repository"`
	LastScan      time.Time         `json:"last_scan"`
	LastCommit    string            `json:"last_commit,omitempty"`
	LastCommitAt  time.Time         `json:"last_commit_at,omitempty"`
	ScannerStatus map[string]Status `json:"scanner_status"`
	Profile       string            `json:"profile,omitempty"`
	Version       string            `json:"version,omitempty"`
}

// Status holds freshness status for a single scanner
type Status struct {
	Scanner     string    `json:"scanner"`
	LastRun     time.Time `json:"last_run"`
	Duration    string    `json:"duration,omitempty"`
	Success     bool      `json:"success"`
	OutputFile  string    `json:"output_file,omitempty"`
	FindingCount int      `json:"finding_count,omitempty"`
	Error       string    `json:"error,omitempty"`
}

// Thresholds defines when data is considered stale
type Thresholds struct {
	FreshMaxHours    int `json:"fresh_max_hours"`    // Fresh if scanned within this many hours
	StaleMaxDays     int `json:"stale_max_days"`     // Stale after this many days
	VeryStaleMaxDays int `json:"very_stale_max_days"` // Very stale after this many days
	ExpiredMaxDays   int `json:"expired_max_days"`   // Expired after this many days
}

// DefaultThresholds returns default freshness thresholds
func DefaultThresholds() Thresholds {
	return Thresholds{
		FreshMaxHours:    24,  // Fresh for 24 hours
		StaleMaxDays:     7,   // Stale after 7 days
		VeryStaleMaxDays: 30,  // Very stale after 30 days
		ExpiredMaxDays:   90,  // Expired after 90 days
	}
}

// Check determines the freshness level based on last scan time
func (t Thresholds) Check(lastScan time.Time) Level {
	if lastScan.IsZero() {
		return LevelUnknown
	}

	age := time.Since(lastScan)

	switch {
	case age < time.Duration(t.FreshMaxHours)*time.Hour:
		return LevelFresh
	case age < time.Duration(t.StaleMaxDays)*24*time.Hour:
		return LevelStale
	case age < time.Duration(t.VeryStaleMaxDays)*24*time.Hour:
		return LevelVeryStale
	default:
		return LevelExpired
	}
}

// String returns a human-readable description of the level
func (l Level) String() string {
	switch l {
	case LevelFresh:
		return "Fresh"
	case LevelStale:
		return "Stale"
	case LevelVeryStale:
		return "Very Stale"
	case LevelExpired:
		return "Expired"
	default:
		return "Unknown"
	}
}

// NeedsRefresh returns true if the data should be refreshed
func (l Level) NeedsRefresh() bool {
	return l != LevelFresh
}

// IsStale returns true if the data is stale or worse
func (l Level) IsStale() bool {
	return l == LevelStale || l == LevelVeryStale || l == LevelExpired
}

// IsVeryStale returns true if the data is very stale or expired
func (l Level) IsVeryStale() bool {
	return l == LevelVeryStale || l == LevelExpired
}

// GetLevel returns the freshness level for the metadata
func (m *Metadata) GetLevel(thresholds Thresholds) Level {
	return thresholds.Check(m.LastScan)
}

// Age returns the age of the scan data
func (m *Metadata) Age() time.Duration {
	if m.LastScan.IsZero() {
		return 0
	}
	return time.Since(m.LastScan)
}

// AgeString returns a human-readable age string
func (m *Metadata) AgeString() string {
	age := m.Age()
	if age == 0 {
		return "never"
	}

	hours := int(age.Hours())
	if hours < 1 {
		return "less than an hour ago"
	}
	if hours < 24 {
		return pluralize(hours, "hour") + " ago"
	}

	days := hours / 24
	if days < 7 {
		return pluralize(days, "day") + " ago"
	}

	weeks := days / 7
	if weeks < 4 {
		return pluralize(weeks, "week") + " ago"
	}

	months := days / 30
	return pluralize(months, "month") + " ago"
}

func pluralize(count int, singular string) string {
	if count == 1 {
		return "1 " + singular
	}
	return string(rune('0'+count/10)) + string(rune('0'+count%10)) + " " + singular + "s"
}

// HasCommitChanged returns true if the repository has new commits since last scan
func (m *Metadata) HasCommitChanged(currentCommit string) bool {
	if m.LastCommit == "" || currentCommit == "" {
		return false // Can't determine
	}
	return m.LastCommit != currentCommit
}

// ScannerCount returns the number of scanners that have run
func (m *Metadata) ScannerCount() int {
	return len(m.ScannerStatus)
}

// SuccessfulScanners returns the number of successful scanner runs
func (m *Metadata) SuccessfulScanners() int {
	count := 0
	for _, s := range m.ScannerStatus {
		if s.Success {
			count++
		}
	}
	return count
}

// FailedScanners returns the names of failed scanners
func (m *Metadata) FailedScanners() []string {
	var failed []string
	for name, s := range m.ScannerStatus {
		if !s.Success {
			failed = append(failed, name)
		}
	}
	return failed
}

// CheckResult holds the result of a freshness check
type CheckResult struct {
	Repository   string        `json:"repository"`
	Level        Level         `json:"level"`
	LastScan     time.Time     `json:"last_scan"`
	Age          time.Duration `json:"age"`
	AgeString    string        `json:"age_string"`
	NeedsRefresh bool          `json:"needs_refresh"`
	CommitChanged bool         `json:"commit_changed,omitempty"`
	Scanners     int           `json:"scanners"`
	Successful   int           `json:"successful"`
	Failed       []string      `json:"failed,omitempty"`
}

// NewCheckResult creates a check result from metadata
func NewCheckResult(m *Metadata, thresholds Thresholds, currentCommit string) CheckResult {
	level := m.GetLevel(thresholds)
	return CheckResult{
		Repository:    m.Repository,
		Level:         level,
		LastScan:      m.LastScan,
		Age:           m.Age(),
		AgeString:     m.AgeString(),
		NeedsRefresh:  level.NeedsRefresh(),
		CommitChanged: m.HasCommitChanged(currentCommit),
		Scanners:      m.ScannerCount(),
		Successful:    m.SuccessfulScanners(),
		Failed:        m.FailedScanners(),
	}
}
