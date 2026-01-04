package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/crashappsec/zero/pkg/core/config"
	"github.com/crashappsec/zero/pkg/core/feeds"
	"github.com/crashappsec/zero/pkg/core/rules"
	"github.com/crashappsec/zero/pkg/core/terminal"
	"github.com/spf13/cobra"
)

var (
	feedsForce bool
	feedsType  string
)

var feedsCmd = &cobra.Command{
	Use:   "feeds",
	Short: "Manage external data feeds",
	Long: `Synchronize and manage external data feeds.

Zero can sync with various data sources to enhance scanning:
  - Semgrep community rules
  - GitHub advisories
  - OSV vulnerability database

Examples:
  zero feeds sync              Sync all enabled feeds
  zero feeds sync --force      Force sync even if fresh
  zero feeds status            Show feed sync status
  zero feeds rules             Generate rules from RAG patterns`,
}

var feedsSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync external feeds",
	Long: `Synchronize external data feeds.

Downloads and caches data from configured feed sources.
By default, only syncs feeds that are due based on their frequency settings.`,
	RunE: runFeedsSync,
}

var feedsStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show feed status",
	Long:  `Display the sync status of all configured feeds.`,
	RunE:  runFeedsStatus,
}

var feedsRulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "Generate rules from RAG patterns",
	Long: `Generate Semgrep rules from RAG pattern definitions.

This command reads patterns from the RAG directory and generates
Semgrep-compatible rules for use in code security scanning.`,
	RunE: runFeedsRules,
}

func init() {
	rootCmd.AddCommand(feedsCmd)
	feedsCmd.AddCommand(feedsSyncCmd)
	feedsCmd.AddCommand(feedsStatusCmd)
	feedsCmd.AddCommand(feedsRulesCmd)

	feedsSyncCmd.Flags().BoolVar(&feedsForce, "force", false, "Force sync even if data is fresh")
	feedsSyncCmd.Flags().StringVar(&feedsType, "type", "", "Sync specific feed type only")

	feedsRulesCmd.Flags().BoolVar(&feedsForce, "force", false, "Force regenerate even if unchanged")
}

func runFeedsSync(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	zeroHome := cfg.ZeroHome()
	if zeroHome == "" {
		zeroHome = ".zero"
	}

	term := terminal.New()

	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		term.Info("\nInterrupted...")
		cancel()
	}()

	term.Divider()
	term.Info("%s", term.Color(terminal.Bold, "Syncing Security Feeds"))
	term.Divider()

	mgr := feeds.NewManager(zeroHome)

	var results []feeds.SyncResult
	if feedsForce {
		results, err = mgr.SyncAll(ctx)
	} else {
		results, err = mgr.SyncIfNeeded(ctx)
	}

	if err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	// Display results
	synced := 0
	skipped := 0
	failed := 0

	for _, r := range results {
		if r.Skipped {
			term.Info("  %s %s - skipped (%s)",
				term.Color(terminal.Dim, "○"),
				r.Feed,
				r.Reason,
			)
			skipped++
		} else if r.Success {
			term.Success("  %s %s (%s, %d items)",
				term.Color(terminal.Green, "✓"),
				r.Feed,
				formatDuration(r.Duration),
				r.ItemCount,
			)
			synced++
		} else {
			term.Error("  %s %s - %s",
				term.Color(terminal.Red, "✗"),
				r.Feed,
				r.Error,
			)
			failed++
		}
	}

	term.Divider()
	if failed > 0 {
		term.Info("Complete: %d synced, %d skipped, %d failed", synced, skipped, failed)
	} else {
		term.Success("Complete: %d synced, %d skipped", synced, skipped)
	}

	return nil
}

func runFeedsStatus(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	zeroHome := cfg.ZeroHome()
	if zeroHome == "" {
		zeroHome = ".zero"
	}

	term := terminal.New()

	mgr := feeds.NewManager(zeroHome)
	if err := mgr.LoadStatus(); err != nil {
		return fmt.Errorf("loading status: %w", err)
	}

	term.Divider()
	term.Info("%s", term.Color(terminal.Bold, "Feed Status"))
	term.Divider()

	feedConfig := mgr.GetConfig()
	statuses := mgr.GetAllStatus()

	// Build status map for quick lookup
	statusMap := make(map[feeds.FeedType]feeds.FeedStatus)
	for _, s := range statuses {
		statusMap[s.Type] = s
	}

	// Show each configured feed
	for _, fc := range feedConfig.Feeds {
		enabledStr := term.Color(terminal.Green, "enabled")
		if !fc.Enabled {
			enabledStr = term.Color(terminal.Dim, "disabled")
		}

		term.Info("\n%s %s",
			term.Color(terminal.Cyan, "▸"),
			term.Color(terminal.Bold, string(fc.Type)),
		)
		term.Info("  Status: %s", enabledStr)
		term.Info("  Frequency: %s", fc.Frequency)

		if status, ok := statusMap[fc.Type]; ok {
			if !status.LastSync.IsZero() {
				age := time.Since(status.LastSync)
				term.Info("  Last Sync: %s (%s ago)",
					status.LastSync.Format("2006-01-02 15:04"),
					formatAge(age),
				)
			}
			if status.ItemCount > 0 {
				term.Info("  Items: %d", status.ItemCount)
			}
			if status.LastError != "" {
				term.Error("  Last Error: %s", status.LastError)
			}
		} else {
			term.Info("  Last Sync: never")
		}
	}

	term.Divider()
	return nil
}

func runFeedsRules(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	zeroHome := cfg.ZeroHome()
	if zeroHome == "" {
		zeroHome = ".zero"
	}

	term := terminal.New()

	term.Divider()
	term.Info("%s", term.Color(terminal.Bold, "Generating Rules from RAG Patterns"))
	term.Divider()

	mgr := rules.NewManager(zeroHome)
	results, err := mgr.RefreshRules(feedsForce)
	if err != nil {
		return fmt.Errorf("rule generation failed: %w", err)
	}

	// Display results
	for _, r := range results {
		if r.Skipped {
			term.Info("  %s %s - skipped (%s)",
				term.Color(terminal.Dim, "○"),
				r.Type,
				r.Reason,
			)
		} else if r.Success {
			term.Success("  %s %s (%s, %d rules)",
				term.Color(terminal.Green, "✓"),
				r.Type,
				formatDuration(r.Duration),
				r.RuleCount,
			)
			for _, f := range r.Files {
				term.Info("    → %s", f)
			}
		} else {
			term.Error("  %s %s - %s",
				term.Color(terminal.Red, "✗"),
				r.Type,
				r.Error,
			)
		}
	}

	term.Divider()
	return nil
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

func formatAge(d time.Duration) string {
	hours := int(d.Hours())
	if hours < 1 {
		return "less than an hour"
	}
	if hours < 24 {
		return fmt.Sprintf("%d hours", hours)
	}
	days := hours / 24
	if days == 1 {
		return "1 day"
	}
	return fmt.Sprintf("%d days", days)
}
