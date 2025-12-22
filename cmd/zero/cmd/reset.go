package cmd

import (
	"fmt"
	"os"

	"github.com/crashappsec/zero/pkg/core/config"
	"github.com/crashappsec/zero/pkg/core/terminal"
	"github.com/spf13/cobra"
)

var resetDryRun bool
var resetYes bool

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset Zero to a clean state",
	Long: `Delete all Zero data for a fresh start.

This removes:
  - All cloned repositories (.zero/repos/)
  - All analysis results (.zero/repos/*/analysis/)
  - Downloaded Semgrep rules (rag/semgrep/community-rules/)
  - Generated rules and caches
  - Scan history and index files

Use this to test Zero end-to-end from a clean state.

Examples:
  zero reset              Reset everything (with confirmation)
  zero reset --dry-run    Preview what would be deleted
  zero reset -y           Skip confirmation`,
	RunE: runReset,
}

func init() {
	rootCmd.AddCommand(resetCmd)

	resetCmd.Flags().BoolVar(&resetDryRun, "dry-run", false, "Preview what would be deleted")
	resetCmd.Flags().BoolVarP(&resetYes, "yes", "y", false, "Skip confirmation")
}

func runReset(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		// Config might not exist, that's ok
		cfg = &config.Config{}
	}

	zeroHome := cfg.ZeroHome()
	if zeroHome == "" {
		zeroHome = ".zero"
	}

	term := terminal.New()

	// Define all paths to clean
	type cleanTarget struct {
		path        string
		description string
	}

	targets := []cleanTarget{
		{zeroHome, "Zero data directory (repos, analysis, index)"},
		{"rag/semgrep/community-rules", "Downloaded Semgrep community rules"},
	}

	// Calculate sizes and filter to existing paths
	var existingTargets []cleanTarget
	var totalSize int64

	for _, t := range targets {
		if info, err := os.Stat(t.path); err == nil {
			size := int64(0)
			if info.IsDir() {
				size = getDirSize(t.path)
			} else {
				size = info.Size()
			}
			existingTargets = append(existingTargets, t)
			totalSize += size
		}
	}

	if len(existingTargets) == 0 {
		term.Info("Nothing to reset - Zero is already clean")
		return nil
	}

	// Show what will be deleted
	term.Divider()
	if resetDryRun {
		term.Info("%s (dry-run)", term.Color(terminal.Bold, "Would delete:"))
	} else {
		term.Info("%s", term.Color(terminal.Bold, "Will delete:"))
	}
	term.Info("")

	for _, t := range existingTargets {
		size := getDirSize(t.path)
		term.Info("  %s", term.Color(terminal.Cyan, t.path))
		term.Info("    %s (%s)", t.description, formatBytes(size))
	}

	term.Info("")
	term.Info("Total size: %s", term.Color(terminal.Bold, formatBytes(totalSize)))

	if resetDryRun {
		return nil
	}

	// Confirm
	if !resetYes {
		term.Info("")
		term.Warning("This will delete all Zero data and cannot be undone!")
		if !term.Confirm("Continue with reset?", false) {
			term.Info("Cancelled")
			return nil
		}
	}

	// Delete
	term.Info("")
	for _, t := range existingTargets {
		term.Info("Deleting %s...", t.path)
		if err := os.RemoveAll(t.path); err != nil {
			return fmt.Errorf("failed to delete %s: %w", t.path, err)
		}
	}

	term.Info("")
	term.Success("Reset complete! Deleted %s", formatBytes(totalSize))
	term.Info("")
	term.Info("Run %s to start fresh", term.Color(terminal.Cyan, "zero hydrate <owner/repo>"))

	return nil
}
