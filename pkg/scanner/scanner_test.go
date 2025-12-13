// Package scanner manages running security scanners
package scanner

import (
	"testing"
	"time"
)

func TestNewProgress(t *testing.T) {
	scanners := []string{"scanner1", "scanner2", "scanner3"}
	progress := NewProgress(scanners)

	if progress.TotalCount != 3 {
		t.Errorf("expected TotalCount 3, got %d", progress.TotalCount)
	}

	if progress.CompletedCount != 0 {
		t.Errorf("expected CompletedCount 0, got %d", progress.CompletedCount)
	}

	if len(progress.Results) != 3 {
		t.Errorf("expected 3 results, got %d", len(progress.Results))
	}

	for _, scanner := range scanners {
		if r, ok := progress.Results[scanner]; !ok {
			t.Errorf("expected result for scanner %s", scanner)
		} else if r.Status != StatusQueued {
			t.Errorf("expected status Queued for %s, got %s", scanner, r.Status)
		}
	}
}

func TestProgressSetRunning(t *testing.T) {
	progress := NewProgress([]string{"scanner1"})

	progress.SetRunning("scanner1")

	if progress.Current != "scanner1" {
		t.Errorf("expected Current 'scanner1', got %s", progress.Current)
	}

	if progress.Results["scanner1"].Status != StatusRunning {
		t.Errorf("expected status Running, got %s", progress.Results["scanner1"].Status)
	}
}

func TestProgressSetComplete(t *testing.T) {
	progress := NewProgress([]string{"scanner1"})
	progress.SetRunning("scanner1")

	progress.SetComplete("scanner1", "100 items", 5*time.Second)

	if progress.CompletedCount != 1 {
		t.Errorf("expected CompletedCount 1, got %d", progress.CompletedCount)
	}

	r := progress.Results["scanner1"]
	if r.Status != StatusComplete {
		t.Errorf("expected status Complete, got %s", r.Status)
	}
	if r.Summary != "100 items" {
		t.Errorf("expected summary '100 items', got %s", r.Summary)
	}
	if r.Duration != 5*time.Second {
		t.Errorf("expected duration 5s, got %v", r.Duration)
	}

	if progress.Current != "" {
		t.Errorf("expected Current to be cleared, got %s", progress.Current)
	}
}

func TestProgressSetFailed(t *testing.T) {
	progress := NewProgress([]string{"scanner1"})

	progress.SetFailed("scanner1", nil, 2*time.Second)

	if progress.CompletedCount != 1 {
		t.Errorf("expected CompletedCount 1, got %d", progress.CompletedCount)
	}

	if progress.Results["scanner1"].Status != StatusFailed {
		t.Errorf("expected status Failed, got %s", progress.Results["scanner1"].Status)
	}
}

func TestProgressSetSkipped(t *testing.T) {
	progress := NewProgress([]string{"scanner1"})

	progress.SetSkipped("scanner1")

	if progress.CompletedCount != 1 {
		t.Errorf("expected CompletedCount 1, got %d", progress.CompletedCount)
	}

	if progress.Results["scanner1"].Status != StatusSkipped {
		t.Errorf("expected status Skipped, got %s", progress.Results["scanner1"].Status)
	}
}

func TestProgressGetProgress(t *testing.T) {
	progress := NewProgress([]string{"scanner1", "scanner2", "scanner3"})

	progress.SetRunning("scanner1")
	progress.SetComplete("scanner1", "done", time.Second)
	progress.SetRunning("scanner2")

	completed, total, current := progress.GetProgress()

	if completed != 1 {
		t.Errorf("expected completed 1, got %d", completed)
	}
	if total != 3 {
		t.Errorf("expected total 3, got %d", total)
	}
	if current != "scanner2" {
		t.Errorf("expected current 'scanner2', got %s", current)
	}
}

func TestEstimateTime(t *testing.T) {
	tests := []struct {
		scanner   string
		fileCount int
		minTime   int
		maxTime   int
	}{
		{"package-sbom", 100, 3, 3},
		{"package-vulns", 100, 1, 1},
		{"package-health", 100, 2, 2},
		{"package-provenance", 100, 1, 1},
		{"package-malcontent", 100, 2, 2},      // min is 2
		{"package-malcontent", 10000, 5, 5},    // 10000/2000 = 5
		{"package-malcontent", 100000, 50, 50}, // 100000/2000 = 50
		{"code-vulns", 100, 5, 5},              // min is 5
		{"code-vulns", 10000, 10, 10},          // 10000/1000 = 10
		{"code-secrets", 100, 5, 5},            // min is 5
		{"unknown-scanner", 100, 2, 2},         // default is 2
	}

	for _, tt := range tests {
		t.Run(tt.scanner, func(t *testing.T) {
			est := EstimateTime(tt.scanner, tt.fileCount)
			if est < tt.minTime || est > tt.maxTime {
				t.Errorf("EstimateTime(%s, %d) = %d, want between %d and %d",
					tt.scanner, tt.fileCount, est, tt.minTime, tt.maxTime)
			}
		})
	}
}

func TestTotalEstimate(t *testing.T) {
	scanners := []string{"package-sbom", "package-vulns", "package-health"}
	// sbom=3, vulns=1, health=2 = 6

	total := TotalEstimate(scanners, 100)

	if total != 6 {
		t.Errorf("expected total 6, got %d", total)
	}
}

func TestNewRunner(t *testing.T) {
	runner := NewRunner(".zero")

	if runner.ZeroHome != ".zero" {
		t.Errorf("expected ZeroHome '.zero', got %s", runner.ZeroHome)
	}

	if runner.Timeout != 10*time.Minute {
		t.Errorf("expected Timeout 10m, got %v", runner.Timeout)
	}

	if runner.Parallel != 4 {
		t.Errorf("expected Parallel 4, got %d", runner.Parallel)
	}
}

func TestParseSummary(t *testing.T) {
	tests := []struct {
		name     string
		scanner  string
		json     string
		expected string
	}{
		{
			name:     "package-vulns with findings",
			scanner:  "package-vulns",
			json:     `{"summary": {"critical": 1, "high": 2, "medium": 3, "low": 4}}`,
			expected: "1 critical, 2 high, 3 medium, 4 low",
		},
		{
			name:     "package-vulns no findings",
			scanner:  "package-vulns",
			json:     `{"summary": {"critical": 0, "high": 0, "medium": 0, "low": 0}}`,
			expected: "no findings",
		},
		{
			name:     "package-sbom with packages",
			scanner:  "package-sbom",
			json:     `{"summary": {"total_packages": 150}}`,
			expected: "150 packages",
		},
		{
			name:     "package-sbom with components (fallback)",
			scanner:  "package-sbom",
			json:     `{"summary": {}, "components": [{}, {}, {}]}`,
			expected: "3 packages",
		},
		{
			name:     "invalid json",
			scanner:  "package-vulns",
			json:     `not json`,
			expected: "complete",
		},
		{
			name:     "unknown scanner",
			scanner:  "unknown",
			json:     `{"summary": {}}`,
			expected: "complete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseSummary(tt.scanner, []byte(tt.json))
			if result != tt.expected {
				t.Errorf("parseSummary(%s, ...) = %q, want %q", tt.scanner, result, tt.expected)
			}
		})
	}
}
