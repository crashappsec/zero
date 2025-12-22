// Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
// SPDX-License-Identifier: GPL-3.0

package diff

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFingerprintGenerationCodeSecurity(t *testing.T) {
	gen := NewFingerprintGenerator()

	// Test code-security vuln findings
	vulnData := json.RawMessage(`{
		"findings": {
			"vulns": [
				{
					"rule_id": "G401",
					"title": "Use of weak cryptographic primitive",
					"severity": "high",
					"file": "crypto/cipher.go",
					"line": 42,
					"column": 10
				}
			],
			"secrets": [
				{
					"rule_id": "secret-aws",
					"type": "aws_access_key",
					"severity": "critical",
					"message": "AWS access key detected",
					"file": "config/prod.json",
					"line": 15,
					"column": 5,
					"snippet": "AKIAIOSFODNN7EXAMPLE"
				}
			],
			"api": []
		}
	}`)

	findings, err := gen.FingerprintFindings("code-security", vulnData)
	if err != nil {
		t.Fatalf("FingerprintFindings failed: %v", err)
	}

	if len(findings) != 2 {
		t.Fatalf("Expected 2 findings, got %d", len(findings))
	}

	// Check vuln finding
	vuln := findings[0]
	if vuln.Scanner != "code-security" {
		t.Errorf("Scanner = %q, want code-security", vuln.Scanner)
	}
	if vuln.Feature != "vulns" {
		t.Errorf("Feature = %q, want vulns", vuln.Feature)
	}
	if vuln.Severity != "high" {
		t.Errorf("Severity = %q, want high", vuln.Severity)
	}

	// Check secret finding
	secret := findings[1]
	if secret.Feature != "secrets" {
		t.Errorf("Feature = %q, want secrets", secret.Feature)
	}
	if secret.Severity != "critical" {
		t.Errorf("Severity = %q, want critical", secret.Severity)
	}
}

func TestFingerprintGenerationPackageAnalysis(t *testing.T) {
	gen := NewFingerprintGenerator()

	vulnData := json.RawMessage(`{
		"findings": {
			"vulns": [
				{
					"id": "CVE-2021-44228",
					"package": "log4j",
					"version": "2.14.0",
					"ecosystem": "maven",
					"severity": "critical",
					"title": "Log4Shell RCE"
				}
			]
		}
	}`)

	findings, err := gen.FingerprintFindings("package-analysis", vulnData)
	if err != nil {
		t.Fatalf("FingerprintFindings failed: %v", err)
	}

	if len(findings) != 1 {
		t.Fatalf("Expected 1 finding, got %d", len(findings))
	}

	vuln := findings[0]
	if vuln.Fingerprint.Scanner != "package-analysis/vulns" {
		t.Errorf("Fingerprint.Scanner = %q, want package-analysis/vulns", vuln.Fingerprint.Scanner)
	}

	expectedPrimaryKey := "CVE-2021-44228:log4j:maven"
	if vuln.Fingerprint.PrimaryKey != expectedPrimaryKey {
		t.Errorf("PrimaryKey = %q, want %q", vuln.Fingerprint.PrimaryKey, expectedPrimaryKey)
	}
}

func TestMatcherExactMatch(t *testing.T) {
	opts := DiffOptions{
		FuzzyMatch:    false,
		LineTolerance: 5,
	}
	matcher := NewMatcher(opts)

	fp := FindingFingerprint{
		Scanner:     "code-security/vulns",
		PrimaryKey:  "G401:crypto/cipher.go",
		LocationKey: "crypto/cipher.go:42:10",
		ContentHash: "abc123",
	}

	baseline := []FingerprintedFinding{
		{
			Fingerprint: fp,
			Severity:    "high",
			Scanner:     "code-security",
			Feature:     "vulns",
			File:        "crypto/cipher.go",
			Line:        42,
			Message:     "Use of weak cryptographic primitive",
		},
	}

	compare := []FingerprintedFinding{
		{
			Fingerprint: fp,
			Severity:    "high",
			Scanner:     "code-security",
			Feature:     "vulns",
			File:        "crypto/cipher.go",
			Line:        42,
			Message:     "Use of weak cryptographic primitive",
		},
	}

	results := matcher.MatchFindings(baseline, compare)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if results[0].Status != MatchExact {
		t.Errorf("Expected MatchExact, got %v", results[0].Status)
	}
}

func TestMatcherFuzzyMatch(t *testing.T) {
	opts := DiffOptions{
		FuzzyMatch:    true,
		LineTolerance: 5,
	}
	matcher := NewMatcher(opts)

	baseFp := FindingFingerprint{
		Scanner:     "code-security/vulns",
		PrimaryKey:  "G401:crypto/cipher.go",
		LocationKey: "crypto/cipher.go:42:10",
		ContentHash: "abc123",
	}

	compareFp := FindingFingerprint{
		Scanner:     "code-security/vulns",
		PrimaryKey:  "G401:crypto/cipher.go",
		LocationKey: "crypto/cipher.go:45:10", // Within tolerance
		ContentHash: "def456",                 // Different content hash to trigger fuzzy match
	}

	baseline := []FingerprintedFinding{
		{
			Fingerprint: baseFp,
			Severity:    "high",
			Scanner:     "code-security",
			File:        "crypto/cipher.go",
			Line:        42,
		},
	}

	compare := []FingerprintedFinding{
		{
			Fingerprint: compareFp,
			Severity:    "high",
			Scanner:     "code-security",
			File:        "crypto/cipher.go",
			Line:        45,
		},
	}

	results := matcher.MatchFindings(baseline, compare)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if results[0].Status != MatchSimilar {
		t.Errorf("Expected MatchSimilar, got %v", results[0].Status)
	}
}

func TestMatcherNewFinding(t *testing.T) {
	opts := DiffOptions{
		FuzzyMatch:    true,
		LineTolerance: 5,
	}
	matcher := NewMatcher(opts)

	baseline := []FingerprintedFinding{}

	compare := []FingerprintedFinding{
		{
			Fingerprint: FindingFingerprint{
				Scanner:     "code-security/vulns",
				PrimaryKey:  "G401:crypto/cipher.go",
				LocationKey: "crypto/cipher.go:42:10",
			},
			Severity: "high",
			Scanner:  "code-security",
			File:     "crypto/cipher.go",
			Line:     42,
		},
	}

	results := matcher.MatchFindings(baseline, compare)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if results[0].Status != MatchNew {
		t.Errorf("Expected MatchNew, got %v", results[0].Status)
	}
}

func TestMatcherFixedFinding(t *testing.T) {
	opts := DiffOptions{
		FuzzyMatch:    true,
		LineTolerance: 5,
	}
	matcher := NewMatcher(opts)

	baseline := []FingerprintedFinding{
		{
			Fingerprint: FindingFingerprint{
				Scanner:     "code-security/vulns",
				PrimaryKey:  "G401:crypto/cipher.go",
				LocationKey: "crypto/cipher.go:42:10",
			},
			Severity: "high",
			Scanner:  "code-security",
			File:     "crypto/cipher.go",
			Line:     42,
		},
	}

	compare := []FingerprintedFinding{}

	results := matcher.MatchFindings(baseline, compare)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if results[0].Status != MatchFixed {
		t.Errorf("Expected MatchFixed, got %v", results[0].Status)
	}
}

func TestMatcherMovedFinding(t *testing.T) {
	opts := DiffOptions{
		FuzzyMatch:    true,
		LineTolerance: 5,
	}
	matcher := NewMatcher(opts)

	contentHash := "deadbeef12345678"

	baseline := []FingerprintedFinding{
		{
			Fingerprint: FindingFingerprint{
				Scanner:     "code-security/vulns",
				PrimaryKey:  "G401:crypto/cipher.go",
				LocationKey: "crypto/cipher.go:42:10",
				ContentHash: contentHash,
			},
			Severity: "high",
			Scanner:  "code-security",
			File:     "crypto/cipher.go",
			Line:     42,
		},
	}

	compare := []FingerprintedFinding{
		{
			Fingerprint: FindingFingerprint{
				Scanner:     "code-security/vulns",
				PrimaryKey:  "G401:crypto/cipher.go",
				LocationKey: "crypto/cipher.go:100:10", // Moved far beyond tolerance
				ContentHash: contentHash,              // Same content hash
			},
			Severity: "high",
			Scanner:  "code-security",
			File:     "crypto/cipher.go",
			Line:     100,
		},
	}

	results := matcher.MatchFindings(baseline, compare)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if results[0].Status != MatchMoved {
		t.Errorf("Expected MatchMoved, got %v", results[0].Status)
	}
}

func TestDeltaSummaryComputation(t *testing.T) {
	summary := computeSummaryFromCounts(3, 2, 10, 1, 0, 0, 1)

	if summary.TotalNew != 3 {
		t.Errorf("TotalNew = %d, want 3", summary.TotalNew)
	}
	if summary.TotalFixed != 2 {
		t.Errorf("TotalFixed = %d, want 2", summary.TotalFixed)
	}
	if summary.TotalUnchanged != 10 {
		t.Errorf("TotalUnchanged = %d, want 10", summary.TotalUnchanged)
	}
	if summary.NetChange != 1 {
		t.Errorf("NetChange = %d, want 1", summary.NetChange)
	}
	if summary.RiskTrend != "degrading" {
		t.Errorf("RiskTrend = %q, want degrading", summary.RiskTrend)
	}
}

func TestDeltaSummaryImproving(t *testing.T) {
	summary := computeSummaryFromCounts(1, 5, 10, 0, 0, 2, 0)

	if summary.RiskTrend != "improving" {
		t.Errorf("RiskTrend = %q, want improving", summary.RiskTrend)
	}
}

func TestDeltaSummaryStable(t *testing.T) {
	summary := computeSummaryFromCounts(2, 2, 10, 0, 0, 0, 0)

	if summary.RiskTrend != "stable" {
		t.Errorf("RiskTrend = %q, want stable", summary.RiskTrend)
	}
}

// Helper function for testing summary computation
func computeSummaryFromCounts(newCount, fixedCount, unchanged, newCrit, newHigh, fixedCrit, fixedHigh int) DeltaSummary {
	summary := DeltaSummary{
		TotalNew:       newCount,
		TotalFixed:     fixedCount,
		TotalUnchanged: unchanged,
		NewCritical:    newCrit,
		NewHigh:        newHigh,
		FixedCritical:  fixedCrit,
		FixedHigh:      fixedHigh,
		NetChange:      newCount - fixedCount,
	}

	// Determine risk trend
	criticalDelta := newCrit - fixedCrit
	highDelta := newHigh - fixedHigh

	if criticalDelta > 0 || (criticalDelta == 0 && highDelta > 0) || summary.NetChange > 2 {
		summary.RiskTrend = "degrading"
	} else if fixedCrit > newCrit || (fixedCrit == newCrit && fixedHigh > newHigh) || summary.NetChange < -2 {
		summary.RiskTrend = "improving"
	} else {
		summary.RiskTrend = "stable"
	}

	return summary
}

func TestFormatterJSON(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, false)

	delta := &ScanDelta{
		BaselineScanID: "scan-1",
		CompareScanID:  "scan-2",
		BaselineCommit: "abc1234",
		CompareCommit:  "def5678",
		Summary: DeltaSummary{
			TotalNew:       2,
			TotalFixed:     1,
			TotalUnchanged: 5,
			NetChange:      1,
			RiskTrend:      "stable",
		},
		ScannerDeltas: map[string]ScannerDelta{},
	}

	err := f.FormatDelta(delta, "json")
	if err != nil {
		t.Fatalf("FormatDelta failed: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("Expected non-empty JSON output")
	}
	if !bytes.Contains([]byte(output), []byte(`"baseline_scan_id"`)) {
		t.Error("Expected JSON to contain baseline_scan_id")
	}
}

func TestFormatterSummary(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, false)

	delta := &ScanDelta{
		BaselineCommit: "abc1234",
		CompareCommit:  "def5678",
		Summary: DeltaSummary{
			TotalNew:   3,
			TotalFixed: 2,
			NetChange:  1,
			RiskTrend:  "degrading",
		},
	}

	err := f.FormatDelta(delta, "summary")
	if err != nil {
		t.Fatalf("FormatDelta failed: %v", err)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("abc1234")) {
		t.Error("Expected summary to contain baseline commit")
	}
	if !bytes.Contains([]byte(output), []byte("def5678")) {
		t.Error("Expected summary to contain compare commit")
	}
}

func TestHistoryManagerBasicOperations(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "zero-diff-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := DefaultHistoryConfig()
	mgr := NewHistoryManager(tmpDir, cfg)

	projectID := "test/repo"

	// Load empty history
	history, err := mgr.LoadHistory(projectID)
	if err != nil {
		t.Fatalf("LoadHistory failed: %v", err)
	}

	if history.ProjectID != projectID {
		t.Errorf("ProjectID = %q, want %q", history.ProjectID, projectID)
	}
	if len(history.Scans) != 0 {
		t.Errorf("Expected empty scans, got %d", len(history.Scans))
	}

	// Add a scan record
	scan := ScanRecord{
		ScanID:          "20241216-120000-abc1234",
		CommitHash:      "abc1234567890",
		CommitShort:     "abc1234",
		Branch:          "main",
		StartedAt:       time.Now().Add(-time.Minute).Format(time.RFC3339),
		CompletedAt:     time.Now().Format(time.RFC3339),
		DurationSeconds: 60,
		Profile:         "standard",
		ScannersRun:     []string{"code-security", "package-analysis"},
		Status:          "complete",
	}
	history.Scans = append(history.Scans, scan)
	history.TotalScans = 1

	// Save history
	err = mgr.SaveHistory(projectID, history)
	if err != nil {
		t.Fatalf("SaveHistory failed: %v", err)
	}

	// Reload and verify
	history2, err := mgr.LoadHistory(projectID)
	if err != nil {
		t.Fatalf("LoadHistory after save failed: %v", err)
	}

	if history2.TotalScans != 1 {
		t.Errorf("TotalScans = %d, want 1", history2.TotalScans)
	}
	if len(history2.Scans) != 1 {
		t.Errorf("Expected 1 scan, got %d", len(history2.Scans))
	}
	if history2.Scans[0].ScanID != scan.ScanID {
		t.Errorf("ScanID = %q, want %q", history2.Scans[0].ScanID, scan.ScanID)
	}
}

func TestHistoryManagerResolveScanRef(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "zero-diff-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := DefaultHistoryConfig()
	mgr := NewHistoryManager(tmpDir, cfg)

	projectID := "test/repo"

	// Create history with multiple scans
	history := &History{
		ProjectID:  projectID,
		TotalScans: 3,
		Scans: []ScanRecord{
			{ScanID: "scan-3", CommitHash: "ccc3333", CommitShort: "ccc3333", Status: "complete"},
			{ScanID: "scan-2", CommitHash: "bbb2222", CommitShort: "bbb2222", Status: "complete"},
			{ScanID: "scan-1", CommitHash: "aaa1111", CommitShort: "aaa1111", Status: "complete"},
		},
	}

	err = mgr.SaveHistory(projectID, history)
	if err != nil {
		t.Fatalf("SaveHistory failed: %v", err)
	}

	// Test "latest"
	scan, err := mgr.ResolveScanRef(projectID, "latest")
	if err != nil {
		t.Fatalf("ResolveScanRef(latest) failed: %v", err)
	}
	if scan.ScanID != "scan-3" {
		t.Errorf("latest resolved to %q, want scan-3", scan.ScanID)
	}

	// Test "latest~1"
	scan, err = mgr.ResolveScanRef(projectID, "latest~1")
	if err != nil {
		t.Fatalf("ResolveScanRef(latest~1) failed: %v", err)
	}
	if scan.ScanID != "scan-2" {
		t.Errorf("latest~1 resolved to %q, want scan-2", scan.ScanID)
	}

	// Test commit hash
	scan, err = mgr.ResolveScanRef(projectID, "aaa1111")
	if err != nil {
		t.Fatalf("ResolveScanRef(aaa1111) failed: %v", err)
	}
	if scan.ScanID != "scan-1" {
		t.Errorf("aaa1111 resolved to %q, want scan-1", scan.ScanID)
	}

	// Test scan ID directly
	scan, err = mgr.ResolveScanRef(projectID, "scan-2")
	if err != nil {
		t.Fatalf("ResolveScanRef(scan-2) failed: %v", err)
	}
	if scan.ScanID != "scan-2" {
		t.Errorf("scan-2 resolved to %q, want scan-2", scan.ScanID)
	}
}

func TestSeverityRank(t *testing.T) {
	tests := []struct {
		severity string
		want     int
	}{
		{"critical", 0},
		{"CRITICAL", 0},
		{"high", 1},
		{"HIGH", 1},
		{"medium", 2},
		{"low", 3},
		{"info", 4},
		{"unknown", 4},
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			got := severityRank(tt.severity)
			if got != tt.want {
				t.Errorf("severityRank(%q) = %d, want %d", tt.severity, got, tt.want)
			}
		})
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"./src/main.go", "src/main.go"},
		{"src/main.go", "src/main.go"},
		{"/src/main.go", "src/main.go"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizePath(tt.input)
			if got != tt.want {
				t.Errorf("normalizePath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestHashContent(t *testing.T) {
	// Same content should produce same hash
	hash1 := hashContent("test", "content")
	hash2 := hashContent("test", "content")

	if hash1 != hash2 {
		t.Errorf("Same content produced different hashes: %q vs %q", hash1, hash2)
	}

	// Different content should produce different hash
	hash3 := hashContent("different", "content")
	if hash1 == hash3 {
		t.Error("Different content produced same hash")
	}

	// Hash should be 16 characters
	if len(hash1) != 16 {
		t.Errorf("Hash length = %d, want 16", len(hash1))
	}
}

func TestMaskSecret(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"AKIAIOSFODNN7EXAMPLE", "AKIA****MPLE"},
		{"short", "****"},
		{"12345678901234567890", "1234****7890"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := maskSecret(tt.input)
			if got != tt.want {
				t.Errorf("maskSecret(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestPreserveScan(t *testing.T) {
	// Create temp directory structure
	tmpDir, err := os.MkdirTemp("", "zero-diff-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	projectID := "test/repo"

	// Create analysis directory with sample file
	analysisDir := filepath.Join(tmpDir, "repos", "test", "repo", "analysis")
	err = os.MkdirAll(analysisDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create analysis dir: %v", err)
	}

	// Write sample analysis file
	sampleContent := []byte(`{"findings": []}`)
	err = os.WriteFile(filepath.Join(analysisDir, "code-security.json"), sampleContent, 0644)
	if err != nil {
		t.Fatalf("Failed to write sample file: %v", err)
	}

	cfg := DefaultHistoryConfig()
	mgr := NewHistoryManager(tmpDir, cfg)

	// Preserve scan
	scan := ScanRecord{
		ScanID:      "20241216-120000-abc1234",
		CommitHash:  "abc1234567890",
		CommitShort: "abc1234",
		Status:      "complete",
	}

	err = mgr.PreserveScan(projectID, scan)
	if err != nil {
		t.Fatalf("PreserveScan failed: %v", err)
	}

	// Verify history was updated
	history, err := mgr.LoadHistory(projectID)
	if err != nil {
		t.Fatalf("LoadHistory after preserve failed: %v", err)
	}

	if len(history.Scans) != 1 {
		t.Errorf("Expected 1 scan, got %d", len(history.Scans))
	}

	// Verify scan directory was created
	scanDir := filepath.Join(tmpDir, "repos", "test", "repo", "history", "scans", scan.ScanID)
	if _, err := os.Stat(scanDir); os.IsNotExist(err) {
		t.Errorf("Scan directory was not created: %s", scanDir)
	}

	// Verify file was copied
	copiedFile := filepath.Join(scanDir, "code-security.json")
	if _, err := os.Stat(copiedFile); os.IsNotExist(err) {
		t.Errorf("Analysis file was not copied: %s", copiedFile)
	}
}

func TestFilterMatches(t *testing.T) {
	opts := DiffOptions{
		FuzzyMatch:    true,
		LineTolerance: 5,
		ShowNewOnly:   true,
	}
	matcher := NewMatcher(opts)

	results := []MatchResult{
		{Status: MatchNew, NewFinding: &DeltaFinding{Severity: "high"}},
		{Status: MatchFixed, OldFinding: &DeltaFinding{Severity: "medium"}},
		{Status: MatchExact, OldFinding: &DeltaFinding{Severity: "low"}},
	}

	filtered := matcher.FilterMatches(results)

	if len(filtered) != 1 {
		t.Fatalf("Expected 1 filtered result, got %d", len(filtered))
	}
	if filtered[0].Status != MatchNew {
		t.Errorf("Expected MatchNew, got %v", filtered[0].Status)
	}
}

func TestFilterMatchesBySeverity(t *testing.T) {
	opts := DiffOptions{
		FuzzyMatch: true,
		Severities: []string{"critical", "high"},
	}
	matcher := NewMatcher(opts)

	results := []MatchResult{
		{Status: MatchNew, NewFinding: &DeltaFinding{Severity: "critical"}},
		{Status: MatchNew, NewFinding: &DeltaFinding{Severity: "high"}},
		{Status: MatchNew, NewFinding: &DeltaFinding{Severity: "medium"}},
		{Status: MatchNew, NewFinding: &DeltaFinding{Severity: "low"}},
	}

	filtered := matcher.FilterMatches(results)

	if len(filtered) != 2 {
		t.Fatalf("Expected 2 filtered results, got %d", len(filtered))
	}
}

func TestParseLocation(t *testing.T) {
	tests := []struct {
		loc      string
		wantFile string
		wantLine int
	}{
		{"file.go:42", "file.go", 42},
		{"file.go:42:10", "file.go", 42},
		{"path/to/file.go:100", "path/to/file.go", 100},
		{"file.go", "file.go", 0},
	}

	for _, tt := range tests {
		t.Run(tt.loc, func(t *testing.T) {
			file, line := parseLocation(tt.loc)
			if file != tt.wantFile {
				t.Errorf("file = %q, want %q", file, tt.wantFile)
			}
			if line != tt.wantLine {
				t.Errorf("line = %d, want %d", line, tt.wantLine)
			}
		})
	}
}
