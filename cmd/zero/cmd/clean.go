package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/crashappsec/zero/pkg/core/config"
	"github.com/crashappsec/zero/pkg/core/terminal"
	"github.com/spf13/cobra"
)

var cleanOrg string
var cleanDryRun bool
var cleanYes bool

var cleanCmd = &cobra.Command{
	Use:   "clean [owner/repo]",
	Short: "Remove analysis data",
	Long: `Delete hydrated projects and analysis data.

Examples:
  zero clean                      Remove all (with confirmation)
  zero clean owner/repo           Remove specific project
  zero clean --org myorg          Remove all org projects
  zero clean --dry-run            Preview deletion`,
	Args: cobra.MaximumNArgs(1),
	RunE: runClean,
}

func init() {
	rootCmd.AddCommand(cleanCmd)

	cleanCmd.Flags().StringVar(&cleanOrg, "org", "", "Clean all repos in organization")
	cleanCmd.Flags().BoolVar(&cleanDryRun, "dry-run", false, "Preview what would be deleted")
	cleanCmd.Flags().BoolVarP(&cleanYes, "yes", "y", false, "Skip confirmation")
}

func runClean(cmd *cobra.Command, args []string) error {
	var repo string
	if len(args) > 0 {
		repo = args[0]
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	zeroHome := cfg.ZeroHome()
	if zeroHome == "" {
		zeroHome = ".zero"
	}

	term := terminal.New()
	reposPath := filepath.Join(zeroHome, "repos")

	// Determine what to clean
	var targets []string
	var totalSize int64

	if repo != "" {
		// Single repo
		repoPath := filepath.Join(reposPath, repo)
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			return fmt.Errorf("project not found: %s", repo)
		}
		targets = []string{repoPath}
		totalSize = getDirSize(repoPath)
	} else if cleanOrg != "" {
		// All repos in org
		orgPath := filepath.Join(reposPath, cleanOrg)
		if _, err := os.Stat(orgPath); os.IsNotExist(err) {
			return fmt.Errorf("organization not found: %s", cleanOrg)
		}
		targets = []string{orgPath}
		totalSize = getDirSize(orgPath)
	} else {
		// Everything
		if _, err := os.Stat(reposPath); os.IsNotExist(err) {
			term.Info("No data to clean")
			return nil
		}
		targets = []string{reposPath}
		totalSize = getDirSize(reposPath)
	}

	// Show what will be deleted
	term.Divider()
	if cleanDryRun {
		term.Info("%s (dry-run)", term.Color(terminal.Bold, "Would delete:"))
	} else {
		term.Info("%s", term.Color(terminal.Bold, "Will delete:"))
	}

	for _, t := range targets {
		term.Info("  %s", t)
	}
	term.Info("\nTotal size: %s", formatBytes(totalSize))

	if cleanDryRun {
		return nil
	}

	// Confirm
	if !cleanYes {
		if !term.Confirm("\nDelete these files?", false) {
			term.Info("Cancelled")
			return nil
		}
	}

	// Delete
	for _, t := range targets {
		if err := os.RemoveAll(t); err != nil {
			return fmt.Errorf("failed to delete %s: %w", t, err)
		}
	}

	term.Success("Deleted %s", formatBytes(totalSize))
	return nil
}

func getDirSize(path string) int64 {
	var size int64
	filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1fG", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.0fM", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.0fK", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}
