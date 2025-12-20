// Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
// SPDX-License-Identifier: GPL-3.0

package diff

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// Formatter handles output formatting for diff results
type Formatter struct {
	writer io.Writer
	color  bool
}

// NewFormatter creates a new formatter
func NewFormatter(w io.Writer, color bool) *Formatter {
	return &Formatter{writer: w, color: color}
}

// FormatDelta formats a scan delta according to the specified format
func (f *Formatter) FormatDelta(delta *ScanDelta, format string) error {
	switch format {
	case "json":
		return f.formatJSON(delta)
	case "summary":
		return f.formatSummary(delta)
	default:
		return f.formatTable(delta)
	}
}

// formatJSON outputs delta as JSON
func (f *Formatter) formatJSON(delta *ScanDelta) error {
	data, err := json.MarshalIndent(delta, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(f.writer, string(data))
	return nil
}

// formatSummary outputs a one-line summary
func (f *Formatter) formatSummary(delta *ScanDelta) error {
	trend := ""
	switch delta.Summary.RiskTrend {
	case "improving":
		trend = f.colorize("improving", colorGreen) + " " + f.colorize("↓", colorGreen)
	case "degrading":
		trend = f.colorize("degrading", colorRed) + " " + f.colorize("↑", colorRed)
	default:
		trend = f.colorize("stable", colorYellow) + " ↔"
	}

	fmt.Fprintf(f.writer, "%s → %s: %s new, %s fixed, net %s (%s)\n",
		f.colorize(delta.BaselineCommit, colorCyan),
		f.colorize(delta.CompareCommit, colorCyan),
		f.formatCount(delta.Summary.TotalNew, true),
		f.formatCount(delta.Summary.TotalFixed, false),
		f.formatNetChange(delta.Summary.NetChange),
		trend,
	)
	return nil
}

// formatTable outputs delta as a formatted table
func (f *Formatter) formatTable(delta *ScanDelta) error {
	// Header
	f.printDivider()
	fmt.Fprintf(f.writer, "  %s\n", f.colorize("SCAN DIFF", colorBold))
	fmt.Fprintf(f.writer, "  Baseline: %s  →  Compare: %s\n",
		f.colorize(delta.BaselineCommit, colorCyan),
		f.colorize(delta.CompareCommit, colorCyan),
	)
	f.printDivider()

	// Summary
	fmt.Fprintln(f.writer)
	fmt.Fprintf(f.writer, "  %s\n", f.colorize("SUMMARY", colorBold))
	fmt.Fprintf(f.writer, "  ├─ New findings:     %s", f.formatNewCount(delta.Summary))
	fmt.Fprintln(f.writer)
	fmt.Fprintf(f.writer, "  ├─ Fixed findings:   %s", f.formatFixedCount(delta.Summary))
	fmt.Fprintln(f.writer)
	fmt.Fprintf(f.writer, "  ├─ Unchanged:        %d\n", delta.Summary.TotalUnchanged)
	if delta.Summary.TotalMoved > 0 {
		fmt.Fprintf(f.writer, "  ├─ Moved:            %d\n", delta.Summary.TotalMoved)
	}
	fmt.Fprintf(f.writer, "  └─ Net change:       %s %s\n",
		f.formatNetChange(delta.Summary.NetChange),
		f.formatRiskTrend(delta.Summary.RiskTrend),
	)

	// By scanner breakdown
	if len(delta.ScannerDeltas) > 0 {
		fmt.Fprintln(f.writer)
		fmt.Fprintf(f.writer, "  %s\n", f.colorize("BY SCANNER", colorBold))
		scanners := f.sortedScanners(delta.ScannerDeltas)
		for i, scanner := range scanners {
			sd := delta.ScannerDeltas[scanner]
			prefix := "├─"
			if i == len(scanners)-1 {
				prefix = "└─"
			}
			fmt.Fprintf(f.writer, "  %s %-20s +%d new, -%d fixed\n",
				prefix, scanner, len(sd.New), len(sd.Fixed))
		}
	}

	// New findings
	newFindings := f.collectNewFindings(delta)
	if len(newFindings) > 0 {
		fmt.Fprintln(f.writer)
		f.printDivider()
		fmt.Fprintf(f.writer, "  %s (%d)\n", f.colorize("NEW FINDINGS", colorRed), len(newFindings))
		f.printDivider()
		for _, finding := range newFindings {
			f.printFinding(finding, false)
		}
	}

	// Fixed findings
	fixedFindings := f.collectFixedFindings(delta)
	if len(fixedFindings) > 0 {
		fmt.Fprintln(f.writer)
		f.printDivider()
		fmt.Fprintf(f.writer, "  %s (%d)\n", f.colorize("FIXED FINDINGS", colorGreen), len(fixedFindings))
		f.printDivider()
		for _, finding := range fixedFindings {
			f.printFinding(finding, true)
		}
	}

	return nil
}

// printFinding prints a single finding
func (f *Formatter) printFinding(finding DeltaFinding, isFixed bool) {
	// Severity indicator
	var severityIcon string
	switch strings.ToLower(finding.Severity) {
	case "critical":
		severityIcon = f.colorize("●", colorRed) + " " + f.colorize("CRITICAL", colorRed)
	case "high":
		severityIcon = f.colorize("●", colorYellow) + " " + f.colorize("HIGH", colorYellow)
	case "medium":
		severityIcon = f.colorize("●", colorBlue) + " MEDIUM"
	case "low":
		severityIcon = "○ LOW"
	default:
		severityIcon = "○ INFO"
	}

	if isFixed {
		severityIcon = f.colorize("✓", colorGreen) + " " + severityIcon + f.colorize(" (FIXED)", colorGreen)
	}

	fmt.Fprintln(f.writer)
	fmt.Fprintf(f.writer, "  %s  %s\n", severityIcon, f.colorize(finding.Fingerprint.Scanner, colorCyan))

	// Location
	if finding.File != "" {
		loc := finding.File
		if finding.Line > 0 {
			loc = fmt.Sprintf("%s:%d", finding.File, finding.Line)
		}
		fmt.Fprintf(f.writer, "    File: %s\n", loc)
	}

	// Message
	if finding.Message != "" {
		msg := finding.Message
		if len(msg) > 80 {
			msg = msg[:77] + "..."
		}
		fmt.Fprintf(f.writer, "    %s\n", msg)
	}
}

// Helper functions

func (f *Formatter) printDivider() {
	fmt.Fprintln(f.writer, strings.Repeat("=", 80))
}

func (f *Formatter) formatNewCount(s DeltaSummary) string {
	if s.TotalNew == 0 {
		return "0"
	}
	parts := []string{f.colorize(fmt.Sprintf("+%d", s.TotalNew), colorRed)}
	if s.NewCritical > 0 || s.NewHigh > 0 {
		severities := []string{}
		if s.NewCritical > 0 {
			severities = append(severities, fmt.Sprintf("%d critical", s.NewCritical))
		}
		if s.NewHigh > 0 {
			severities = append(severities, fmt.Sprintf("%d high", s.NewHigh))
		}
		parts = append(parts, fmt.Sprintf("(%s)", strings.Join(severities, ", ")))
	}
	return strings.Join(parts, " ")
}

func (f *Formatter) formatFixedCount(s DeltaSummary) string {
	if s.TotalFixed == 0 {
		return "0"
	}
	parts := []string{f.colorize(fmt.Sprintf("-%d", s.TotalFixed), colorGreen)}
	if s.FixedCritical > 0 || s.FixedHigh > 0 {
		severities := []string{}
		if s.FixedCritical > 0 {
			severities = append(severities, fmt.Sprintf("%d critical", s.FixedCritical))
		}
		if s.FixedHigh > 0 {
			severities = append(severities, fmt.Sprintf("%d high", s.FixedHigh))
		}
		parts = append(parts, fmt.Sprintf("(%s)", strings.Join(severities, ", ")))
	}
	return strings.Join(parts, " ")
}

func (f *Formatter) formatCount(count int, isNew bool) string {
	if count == 0 {
		return "0"
	}
	if isNew {
		return f.colorize(fmt.Sprintf("+%d", count), colorRed)
	}
	return f.colorize(fmt.Sprintf("-%d", count), colorGreen)
}

func (f *Formatter) formatNetChange(net int) string {
	if net > 0 {
		return f.colorize(fmt.Sprintf("+%d", net), colorRed)
	} else if net < 0 {
		return f.colorize(fmt.Sprintf("%d", net), colorGreen)
	}
	return "0"
}

func (f *Formatter) formatRiskTrend(trend string) string {
	switch trend {
	case "improving":
		return f.colorize("(improving ↓)", colorGreen)
	case "degrading":
		return f.colorize("(degrading ↑)", colorRed)
	default:
		return f.colorize("(stable ↔)", colorYellow)
	}
}

func (f *Formatter) sortedScanners(deltas map[string]ScannerDelta) []string {
	var scanners []string
	for s := range deltas {
		scanners = append(scanners, s)
	}
	sort.Strings(scanners)
	return scanners
}

func (f *Formatter) collectNewFindings(delta *ScanDelta) []DeltaFinding {
	var findings []DeltaFinding
	for _, sd := range delta.ScannerDeltas {
		findings = append(findings, sd.New...)
	}
	// Sort by severity
	sort.Slice(findings, func(i, j int) bool {
		return severityRank(findings[i].Severity) < severityRank(findings[j].Severity)
	})
	return findings
}

func (f *Formatter) collectFixedFindings(delta *ScanDelta) []DeltaFinding {
	var findings []DeltaFinding
	for _, sd := range delta.ScannerDeltas {
		findings = append(findings, sd.Fixed...)
	}
	// Sort by severity
	sort.Slice(findings, func(i, j int) bool {
		return severityRank(findings[i].Severity) < severityRank(findings[j].Severity)
	})
	return findings
}

func severityRank(sev string) int {
	switch strings.ToLower(sev) {
	case "critical":
		return 0
	case "high":
		return 1
	case "medium":
		return 2
	case "low":
		return 3
	default:
		return 4
	}
}

// Color constants
const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
)

func (f *Formatter) colorize(s string, color string) string {
	if !f.color {
		return s
	}
	return color + s + colorReset
}
