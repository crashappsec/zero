package freshness

import (
	"testing"
	"time"
)

func TestDefaultThresholds(t *testing.T) {
	th := DefaultThresholds()

	if th.FreshMaxHours != 24 {
		t.Errorf("FreshMaxHours = %d, want 24", th.FreshMaxHours)
	}
	if th.StaleMaxDays != 7 {
		t.Errorf("StaleMaxDays = %d, want 7", th.StaleMaxDays)
	}
	if th.VeryStaleMaxDays != 30 {
		t.Errorf("VeryStaleMaxDays = %d, want 30", th.VeryStaleMaxDays)
	}
	if th.ExpiredMaxDays != 90 {
		t.Errorf("ExpiredMaxDays = %d, want 90", th.ExpiredMaxDays)
	}
}

func TestThresholdsCheck(t *testing.T) {
	th := DefaultThresholds()
	now := time.Now()

	// Default thresholds:
	// - Fresh: < 24 hours
	// - Stale: 24 hours to < 7 days
	// - VeryStale: 7 days to < 30 days
	// - Expired: >= 30 days

	tests := []struct {
		name     string
		lastScan time.Time
		expected Level
	}{
		{"zero time", time.Time{}, LevelUnknown},
		{"just now", now, LevelFresh},
		{"12 hours ago", now.Add(-12 * time.Hour), LevelFresh},
		{"2 days ago", now.Add(-2 * 24 * time.Hour), LevelStale},
		{"6 days ago", now.Add(-6 * 24 * time.Hour), LevelStale},
		{"7 days ago", now.Add(-7 * 24 * time.Hour), LevelVeryStale},  // >= 7 days is VeryStale
		{"15 days ago", now.Add(-15 * 24 * time.Hour), LevelVeryStale},
		{"30 days ago", now.Add(-30 * 24 * time.Hour), LevelExpired}, // >= 30 days is Expired
		{"100 days ago", now.Add(-100 * 24 * time.Hour), LevelExpired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := th.Check(tt.lastScan); got != tt.expected {
				t.Errorf("Check(%s) = %s, want %s", tt.name, got, tt.expected)
			}
		})
	}
}

func TestLevelString(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{LevelFresh, "Fresh"},
		{LevelStale, "Stale"},
		{LevelVeryStale, "Very Stale"},
		{LevelExpired, "Expired"},
		{LevelUnknown, "Unknown"},
		{Level("invalid"), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.level), func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("Level(%s).String() = %s, want %s", tt.level, got, tt.expected)
			}
		})
	}
}

func TestLevelNeedsRefresh(t *testing.T) {
	tests := []struct {
		level    Level
		expected bool
	}{
		{LevelFresh, false},
		{LevelStale, true},
		{LevelVeryStale, true},
		{LevelExpired, true},
		{LevelUnknown, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.level), func(t *testing.T) {
			if got := tt.level.NeedsRefresh(); got != tt.expected {
				t.Errorf("Level(%s).NeedsRefresh() = %v, want %v", tt.level, got, tt.expected)
			}
		})
	}
}

func TestLevelIsStale(t *testing.T) {
	tests := []struct {
		level    Level
		expected bool
	}{
		{LevelFresh, false},
		{LevelStale, true},
		{LevelVeryStale, true},
		{LevelExpired, true},
		{LevelUnknown, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.level), func(t *testing.T) {
			if got := tt.level.IsStale(); got != tt.expected {
				t.Errorf("Level(%s).IsStale() = %v, want %v", tt.level, got, tt.expected)
			}
		})
	}
}

func TestLevelIsVeryStale(t *testing.T) {
	tests := []struct {
		level    Level
		expected bool
	}{
		{LevelFresh, false},
		{LevelStale, false},
		{LevelVeryStale, true},
		{LevelExpired, true},
		{LevelUnknown, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.level), func(t *testing.T) {
			if got := tt.level.IsVeryStale(); got != tt.expected {
				t.Errorf("Level(%s).IsVeryStale() = %v, want %v", tt.level, got, tt.expected)
			}
		})
	}
}

func TestMetadataGetLevel(t *testing.T) {
	th := DefaultThresholds()
	now := time.Now()

	m := &Metadata{
		Repository: "test/repo",
		LastScan:   now.Add(-2 * 24 * time.Hour),
	}

	if got := m.GetLevel(th); got != LevelStale {
		t.Errorf("GetLevel() = %s, want %s", got, LevelStale)
	}
}

func TestMetadataAge(t *testing.T) {
	now := time.Now()

	// Zero time
	m := &Metadata{}
	if m.Age() != 0 {
		t.Errorf("Age() for zero time = %v, want 0", m.Age())
	}

	// Recent scan
	m.LastScan = now.Add(-1 * time.Hour)
	age := m.Age()
	if age < 59*time.Minute || age > 61*time.Minute {
		t.Errorf("Age() = %v, want ~1 hour", age)
	}
}

func TestMetadataAgeString(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		lastScan time.Time
		contains string
	}{
		{"never scanned", time.Time{}, "never"},
		{"30 minutes ago", now.Add(-30 * time.Minute), "less than an hour"},
		{"2 hours ago", now.Add(-2 * time.Hour), "hour"},
		{"3 days ago", now.Add(-3 * 24 * time.Hour), "day"},
		{"2 weeks ago", now.Add(-14 * 24 * time.Hour), "week"},
		{"2 months ago", now.Add(-60 * 24 * time.Hour), "month"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metadata{LastScan: tt.lastScan}
			got := m.AgeString()
			if !containsString(got, tt.contains) {
				t.Errorf("AgeString() = %s, want to contain %s", got, tt.contains)
			}
		})
	}
}

func TestMetadataHasCommitChanged(t *testing.T) {
	m := &Metadata{LastCommit: "abc123"}

	tests := []struct {
		name    string
		current string
		want    bool
	}{
		{"same commit", "abc123", false},
		{"different commit", "def456", true},
		{"empty current", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := m.HasCommitChanged(tt.current); got != tt.want {
				t.Errorf("HasCommitChanged(%s) = %v, want %v", tt.current, got, tt.want)
			}
		})
	}

	// Empty last commit
	m.LastCommit = ""
	if m.HasCommitChanged("abc123") {
		t.Error("HasCommitChanged() with empty LastCommit should return false")
	}
}

func TestMetadataScannerCounts(t *testing.T) {
	m := &Metadata{
		ScannerStatus: map[string]Status{
			"scanner1": {Success: true},
			"scanner2": {Success: true},
			"scanner3": {Success: false, Error: "failed"},
		},
	}

	if got := m.ScannerCount(); got != 3 {
		t.Errorf("ScannerCount() = %d, want 3", got)
	}

	if got := m.SuccessfulScanners(); got != 2 {
		t.Errorf("SuccessfulScanners() = %d, want 2", got)
	}

	failed := m.FailedScanners()
	if len(failed) != 1 {
		t.Errorf("FailedScanners() = %d items, want 1", len(failed))
	}
}

func TestNewCheckResult(t *testing.T) {
	th := DefaultThresholds()
	now := time.Now()

	m := &Metadata{
		Repository: "test/repo",
		LastScan:   now.Add(-2 * 24 * time.Hour),
		LastCommit: "abc123",
		ScannerStatus: map[string]Status{
			"scanner1": {Success: true},
			"scanner2": {Success: false, Error: "failed"},
		},
	}

	result := NewCheckResult(m, th, "def456")

	if result.Repository != "test/repo" {
		t.Errorf("Repository = %s, want test/repo", result.Repository)
	}

	if result.Level != LevelStale {
		t.Errorf("Level = %s, want %s", result.Level, LevelStale)
	}

	if !result.NeedsRefresh {
		t.Error("NeedsRefresh should be true for stale data")
	}

	if !result.CommitChanged {
		t.Error("CommitChanged should be true when commits differ")
	}

	if result.Scanners != 2 {
		t.Errorf("Scanners = %d, want 2", result.Scanners)
	}

	if result.Successful != 1 {
		t.Errorf("Successful = %d, want 1", result.Successful)
	}

	if len(result.Failed) != 1 {
		t.Errorf("Failed = %d items, want 1", len(result.Failed))
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
