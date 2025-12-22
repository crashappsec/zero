// Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
// SPDX-License-Identifier: GPL-3.0

package cmd

import (
	"fmt"
	"os"

	"github.com/crashappsec/zero/pkg/core/config"
	"github.com/crashappsec/zero/pkg/workflow/diff"
	"github.com/crashappsec/zero/pkg/core/terminal"
	"github.com/spf13/cobra"
	goterm "golang.org/x/term"
)

var (
	diffScanner   string
	diffSeverity  []string
	diffFormat    string
	diffFuzzy     bool
	diffTolerance int
	diffNewOnly   bool
	diffFixedOnly bool
	diffNoColor   bool
)

var diffCmd = &cobra.Command{
	Use:   "diff <owner/repo> [baseline] [compare]",
	Short: "Compare scan results between two points in time",
	Long: `Compare security scan findings between two scans to identify:
- New findings introduced
- Fixed findings resolved
- Findings that moved (code refactoring)
- Overall security posture trend

Arguments:
  owner/repo    Repository to analyze (required)
  baseline      Baseline scan (commit, scan-id, or "latest~N") [default: latest~1]
  compare       Compare scan (commit, scan-id, or "latest") [default: latest]

Examples:
  zero diff owner/repo                        Compare latest scan to previous
  zero diff owner/repo latest~5               Compare latest to 5 scans ago
  zero diff owner/repo abc123 def456          Compare specific commits
  zero diff owner/repo --scanner code-security  Compare specific scanner only
  zero diff owner/repo --severity critical,high Show only critical/high changes
  zero diff owner/repo --json                   Output as JSON
  zero diff owner/repo --new-only               Show only new findings`,
	Args: cobra.RangeArgs(1, 3),
	RunE: runDiff,
}

func init() {
	rootCmd.AddCommand(diffCmd)

	diffCmd.Flags().StringVar(&diffScanner, "scanner", "", "Scanner to compare (e.g., code-security, package-analysis)")
	diffCmd.Flags().StringSliceVar(&diffSeverity, "severity", nil, "Filter by severity (critical,high,medium,low)")
	diffCmd.Flags().StringVar(&diffFormat, "format", "table", "Output format: table, json, summary")
	diffCmd.Flags().BoolVar(&diffFuzzy, "fuzzy", true, "Enable fuzzy matching for moved code")
	diffCmd.Flags().IntVar(&diffTolerance, "tolerance", 5, "Line tolerance for fuzzy matching")
	diffCmd.Flags().BoolVar(&diffNewOnly, "new-only", false, "Show only new findings")
	diffCmd.Flags().BoolVar(&diffFixedOnly, "fixed-only", false, "Show only fixed findings")
	diffCmd.Flags().BoolVar(&diffNoColor, "no-color", false, "Disable color output")
}

func runDiff(cmd *cobra.Command, args []string) error {
	term := terminal.New()
	repo := args[0]

	// Parse baseline and compare refs
	baselineRef := "latest~1"
	compareRef := "latest"

	if len(args) >= 2 {
		baselineRef = args[1]
	}
	if len(args) >= 3 {
		compareRef = args[2]
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	zeroHome := cfg.ZeroHome()
	if zeroHome == "" {
		zeroHome = ".zero"
	}

	// Create history manager
	historyConfig := diff.DefaultHistoryConfig()
	historyMgr := diff.NewHistoryManager(zeroHome, historyConfig)

	// Check if project exists
	history, err := historyMgr.LoadHistory(repo)
	if err != nil {
		return fmt.Errorf("failed to load history for %s: %w", repo, err)
	}

	if len(history.Scans) == 0 {
		term.Warning("No scan history for %s", repo)
		term.Info("Run: zero hydrate %s", repo)
		term.Info("Then run another scan to create history for comparison")
		return nil
	}

	if len(history.Scans) < 2 {
		term.Warning("Need at least 2 scans for comparison (have %d)", len(history.Scans))
		term.Info("Run another scan: zero hydrate %s", repo)
		return nil
	}

	// Resolve scan references
	baselineScan, err := historyMgr.ResolveScanRef(repo, baselineRef)
	if err != nil {
		return fmt.Errorf("failed to resolve baseline '%s': %w", baselineRef, err)
	}

	compareScan, err := historyMgr.ResolveScanRef(repo, compareRef)
	if err != nil {
		return fmt.Errorf("failed to resolve compare '%s': %w", compareRef, err)
	}

	// Validate that we're comparing different scans
	if baselineScan.ScanID == compareScan.ScanID {
		term.Warning("Baseline and compare are the same scan: %s", baselineScan.ScanID)
		return nil
	}

	// Create diff options
	options := diff.DiffOptions{
		Scanner:       diffScanner,
		Severities:    diffSeverity,
		FuzzyMatch:    diffFuzzy,
		LineTolerance: diffTolerance,
		ShowNewOnly:   diffNewOnly,
		ShowFixedOnly: diffFixedOnly,
		IncludeMoved:  true,
		OutputFormat:  diffFormat,
	}

	// Compute delta
	computer := diff.NewDeltaComputer(historyMgr, options)
	delta, err := computer.ComputeDelta(repo, baselineScan.ScanID, compareScan.ScanID)
	if err != nil {
		return fmt.Errorf("failed to compute diff: %w", err)
	}

	// Format output
	useColor := !diffNoColor && goterm.IsTerminal(int(os.Stdout.Fd()))
	formatter := diff.NewFormatter(os.Stdout, useColor)

	return formatter.FormatDelta(delta, diffFormat)
}
