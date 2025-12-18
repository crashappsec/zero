// Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
// SPDX-License-Identifier: GPL-3.0

package diff

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// DeltaComputer computes the delta between two scans
type DeltaComputer struct {
	historyMgr *HistoryManager
	fpGen      *FingerprintGenerator
	options    DiffOptions
}

// NewDeltaComputer creates a new delta computer
func NewDeltaComputer(historyMgr *HistoryManager, options DiffOptions) *DeltaComputer {
	return &DeltaComputer{
		historyMgr: historyMgr,
		fpGen:      NewFingerprintGenerator(),
		options:    options,
	}
}

// ComputeDelta computes the difference between two scans
func (c *DeltaComputer) ComputeDelta(projectID, baselineScanID, compareScanID string) (*ScanDelta, error) {
	// Load baseline scan results
	baselineResults, err := c.historyMgr.LoadScanResults(projectID, baselineScanID)
	if err != nil {
		return nil, fmt.Errorf("failed to load baseline scan: %w", err)
	}

	// Load compare scan results
	compareResults, err := c.historyMgr.LoadScanResults(projectID, compareScanID)
	if err != nil {
		return nil, fmt.Errorf("failed to load compare scan: %w", err)
	}

	// Get scan records for metadata
	baselineRecord, err := c.historyMgr.GetScan(projectID, baselineScanID)
	if err != nil {
		return nil, fmt.Errorf("failed to get baseline record: %w", err)
	}
	compareRecord, err := c.historyMgr.GetScan(projectID, compareScanID)
	if err != nil {
		return nil, fmt.Errorf("failed to get compare record: %w", err)
	}

	// Initialize delta
	delta := &ScanDelta{
		BaselineScanID: baselineScanID,
		CompareScanID:  compareScanID,
		BaselineCommit: baselineRecord.CommitShort,
		CompareCommit:  compareRecord.CommitShort,
		GeneratedAt:    time.Now(),
		ScannerDeltas:  make(map[string]ScannerDelta),
	}

	// Get list of scanners to compare
	scanners := c.getScannersToCompare(baselineResults, compareResults)

	// Process each scanner
	for _, scanner := range scanners {
		scannerDelta, err := c.computeScannerDelta(scanner, baselineResults[scanner], compareResults[scanner])
		if err != nil {
			continue // Skip scanner on error
		}

		delta.ScannerDeltas[scanner] = scannerDelta
	}

	// Compute summary
	delta.Summary = c.computeSummary(delta.ScannerDeltas)

	return delta, nil
}

// getScannersToCompare returns the list of scanners to compare
func (c *DeltaComputer) getScannersToCompare(baseline, compare map[string]json.RawMessage) []string {
	scannerSet := make(map[string]bool)

	for scanner := range baseline {
		scannerSet[scanner] = true
	}
	for scanner := range compare {
		scannerSet[scanner] = true
	}

	var scanners []string
	for scanner := range scannerSet {
		// Skip non-finding files
		if scanner == "languages" || scanner == "sbom.cdx" {
			continue
		}

		// Apply scanner filter if specified
		if c.options.Scanner != "" && scanner != c.options.Scanner {
			continue
		}
		if len(c.options.Scanners) > 0 && !containsIgnoreCase(c.options.Scanners, scanner) {
			continue
		}

		scanners = append(scanners, scanner)
	}

	return scanners
}

// computeScannerDelta computes the delta for a single scanner
func (c *DeltaComputer) computeScannerDelta(scanner string, baselineData, compareData json.RawMessage) (ScannerDelta, error) {
	delta := ScannerDelta{
		Scanner: scanner,
	}

	// Fingerprint findings
	var baselineFindings, compareFindings []FingerprintedFinding
	var err error

	if len(baselineData) > 0 {
		baselineFindings, err = c.fpGen.FingerprintFindings(scanner, baselineData)
		if err != nil {
			return delta, err
		}
	}

	if len(compareData) > 0 {
		compareFindings, err = c.fpGen.FingerprintFindings(scanner, compareData)
		if err != nil {
			return delta, err
		}
	}

	// Match findings
	matcher := NewMatcher(c.options)
	results := matcher.MatchFindings(baselineFindings, compareFindings)
	results = matcher.FilterMatches(results)

	// Categorize results
	for _, r := range results {
		switch r.Status {
		case MatchNew:
			if r.NewFinding != nil {
				delta.New = append(delta.New, *r.NewFinding)
			}
		case MatchFixed:
			if r.OldFinding != nil {
				delta.Fixed = append(delta.Fixed, *r.OldFinding)
			}
		case MatchMoved:
			if r.OldFinding != nil && r.NewFinding != nil {
				delta.Moved = append(delta.Moved, MovedFinding{
					Fingerprint: r.NewFinding.Fingerprint,
					OldLocation: r.OldFinding.Fingerprint.LocationKey,
					NewLocation: r.NewFinding.Fingerprint.LocationKey,
					Finding:     r.NewFinding.Finding,
					Severity:    r.NewFinding.Severity,
				})
			}
		case MatchExact, MatchSimilar:
			delta.Unchanged++
		}
	}

	return delta, nil
}

// computeSummary computes the overall summary from scanner deltas
func (c *DeltaComputer) computeSummary(scannerDeltas map[string]ScannerDelta) DeltaSummary {
	summary := DeltaSummary{}

	for _, sd := range scannerDeltas {
		// Count new findings
		for _, f := range sd.New {
			summary.TotalNew++
			c.countBySeverity(f.Severity, &summary.NewCritical, &summary.NewHigh, &summary.NewMedium, &summary.NewLow)
		}

		// Count fixed findings
		for _, f := range sd.Fixed {
			summary.TotalFixed++
			c.countBySeverity(f.Severity, &summary.FixedCritical, &summary.FixedHigh, &summary.FixedMedium, &summary.FixedLow)
		}

		// Count moved and unchanged
		summary.TotalMoved += len(sd.Moved)
		summary.TotalUnchanged += sd.Unchanged
	}

	// Calculate net change
	summary.NetChange = summary.TotalNew - summary.TotalFixed

	// Determine risk trend
	summary.RiskTrend = c.determineRiskTrend(summary)

	return summary
}

// countBySeverity increments the appropriate severity counter
func (c *DeltaComputer) countBySeverity(severity string, critical, high, medium, low *int) {
	switch strings.ToLower(severity) {
	case "critical":
		*critical++
	case "high":
		*high++
	case "medium":
		*medium++
	case "low":
		*low++
	}
}

// determineRiskTrend determines the overall risk trend
func (c *DeltaComputer) determineRiskTrend(summary DeltaSummary) string {
	// Weight critical/high more heavily
	newRisk := summary.NewCritical*10 + summary.NewHigh*5 + summary.NewMedium*2 + summary.NewLow
	fixedRisk := summary.FixedCritical*10 + summary.FixedHigh*5 + summary.FixedMedium*2 + summary.FixedLow

	if newRisk > fixedRisk {
		return "degrading"
	} else if fixedRisk > newRisk {
		return "improving"
	}
	return "stable"
}

// ComputeDeltaFromCurrent computes delta comparing current analysis to a historical scan
func (c *DeltaComputer) ComputeDeltaFromCurrent(projectID, baselineScanID string) (*ScanDelta, error) {
	// Load baseline scan results from history
	baselineResults, err := c.historyMgr.LoadScanResults(projectID, baselineScanID)
	if err != nil {
		return nil, fmt.Errorf("failed to load baseline scan: %w", err)
	}

	// Load current analysis results
	analysisDir := c.historyMgr.GetAnalysisDir(projectID)
	compareResults := make(map[string]json.RawMessage)

	// Read scanner files from analysis directory
	scannerFiles, err := c.historyMgr.GetScanFiles(projectID, baselineScanID)
	if err == nil {
		for _, scanner := range scannerFiles {
			data, err := readPossiblyCompressed(analysisDir + "/" + scanner + ".json")
			if err == nil {
				compareResults[scanner] = data
			}
		}
	}

	// Get baseline record for metadata
	baselineRecord, err := c.historyMgr.GetScan(projectID, baselineScanID)
	if err != nil {
		return nil, fmt.Errorf("failed to get baseline record: %w", err)
	}

	// Initialize delta
	delta := &ScanDelta{
		BaselineScanID: baselineScanID,
		CompareScanID:  "current",
		BaselineCommit: baselineRecord.CommitShort,
		CompareCommit:  "HEAD",
		GeneratedAt:    time.Now(),
		ScannerDeltas:  make(map[string]ScannerDelta),
	}

	// Get list of scanners to compare
	scanners := c.getScannersToCompare(baselineResults, compareResults)

	// Process each scanner
	for _, scanner := range scanners {
		scannerDelta, err := c.computeScannerDelta(scanner, baselineResults[scanner], compareResults[scanner])
		if err != nil {
			continue
		}
		delta.ScannerDeltas[scanner] = scannerDelta
	}

	// Compute summary
	delta.Summary = c.computeSummary(delta.ScannerDeltas)

	return delta, nil
}
