// Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
// SPDX-License-Identifier: GPL-3.0

package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/crashappsec/zero/pkg/config"
	"github.com/crashappsec/zero/pkg/diff"
	"github.com/crashappsec/zero/pkg/terminal"
	"github.com/spf13/cobra"
)

var historyLimit int
var historyJSON bool

var historyCmd = &cobra.Command{
	Use:   "history [owner/repo]",
	Short: "Show scan history for a project",
	Long: `Display scan history showing previous scans with commit info and profiles.

Examples:
  zero history owner/repo             Show last 10 scans
  zero history owner/repo --limit 20  Show last 20 scans
  zero history owner/repo --json      Output as JSON`,
	Args: cobra.ExactArgs(1),
	RunE: runHistory,
}

func init() {
	rootCmd.AddCommand(historyCmd)

	historyCmd.Flags().IntVar(&historyLimit, "limit", 10, "Number of scans to show")
	historyCmd.Flags().BoolVar(&historyJSON, "json", false, "Output as JSON")
}

func runHistory(cmd *cobra.Command, args []string) error {
	repo := args[0]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	zeroHome := cfg.ZeroHome()
	if zeroHome == "" {
		zeroHome = ".zero"
	}

	term := terminal.New()

	// Use the diff package's history manager
	historyConfig := diff.DefaultHistoryConfig()
	historyMgr := diff.NewHistoryManager(zeroHome, historyConfig)

	// Load history
	history, err := historyMgr.LoadHistory(repo)
	if err != nil {
		return fmt.Errorf("failed to load history: %w", err)
	}

	if len(history.Scans) == 0 {
		term.Info("No scan history for %s", repo)
		term.Info("Run: zero hydrate %s", repo)
		return nil
	}

	// JSON output
	if historyJSON {
		// Limit scans
		if historyLimit > 0 && len(history.Scans) > historyLimit {
			history.Scans = history.Scans[:historyLimit]
		}

		output, err := json.MarshalIndent(history, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(output))
		return nil
	}

	// Text output
	term.Divider()
	term.Info("%s %s",
		term.Color(terminal.Bold, "Scan History:"),
		term.Color(terminal.Cyan, repo),
	)
	term.Info("Total scans: %d", history.TotalScans)
	term.Divider()

	// Show scans (up to limit)
	count := len(history.Scans)
	if historyLimit > 0 && count > historyLimit {
		count = historyLimit
	}

	for i := 0; i < count; i++ {
		scan := history.Scans[i]

		// Parse timestamp
		var timeStr string
		if t, err := time.Parse(time.RFC3339, scan.CompletedAt); err == nil {
			timeStr = t.Format("Jan 02, 2006 15:04")
		} else {
			timeStr = scan.CompletedAt
		}

		// Status indicator
		var statusIcon string
		switch scan.Status {
		case "complete":
			statusIcon = term.Color(terminal.Green, "✓")
		case "failed":
			statusIcon = term.Color(terminal.Red, "✗")
		default:
			statusIcon = term.Color(terminal.Yellow, "○")
		}

		// Branch info
		branchStr := ""
		if scan.Branch != "" {
			branchStr = fmt.Sprintf(" (%s)", scan.Branch)
		}

		// Findings summary
		findingsStr := ""
		if scan.FindingsSummary.Total > 0 {
			findingsStr = fmt.Sprintf("  Findings: %d (%d crit, %d high)",
				scan.FindingsSummary.Total,
				scan.FindingsSummary.Critical,
				scan.FindingsSummary.High,
			)
		}

		term.Info("")
		term.Info("%s %s %s%s",
			statusIcon,
			term.Color(terminal.Cyan, scan.CommitShort),
			timeStr,
			branchStr,
		)
		term.Info("  Profile: %s  Duration: %ds  Scanners: %d%s",
			scan.Profile,
			scan.DurationSeconds,
			len(scan.ScannersRun),
			findingsStr,
		)
	}

	if len(history.Scans) > count {
		term.Info("")
		term.Info("... and %d more (use --limit to show more)", len(history.Scans)-count)
	}

	// Tip about diff command
	if len(history.Scans) >= 2 {
		term.Info("")
		term.Info("Tip: Use 'zero diff %s' to compare scans", repo)
	}

	return nil
}
