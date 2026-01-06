package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/crashappsec/zero/pkg/core/config"
	"github.com/crashappsec/zero/pkg/core/feeds"
	"github.com/crashappsec/zero/pkg/core/rag"
	"github.com/crashappsec/zero/pkg/core/terminal"
	techid "github.com/crashappsec/zero/pkg/scanner/technology-identification"
	"github.com/spf13/cobra"
)

var (
	feedsForce bool
	feedsType  string
)

var feedsCmd = &cobra.Command{
	Use:   "feeds",
	Short: "Manage security rules and data feeds",
	Long: `Manage security scanning rules and external data feeds.

Zero uses two sources of Semgrep rules:
  - RAG patterns: Custom rules generated from the rag/ knowledge base
  - Semgrep community: Official rules from semgrep.dev registry

Examples:
  zero feeds rag               Generate rules from RAG knowledge base
  zero feeds semgrep           Sync Semgrep community rules (SAST)
  zero feeds semgrep --force   Force sync even if fresh
  zero feeds status            Show feed status`,
}

var feedsSemgrepCmd = &cobra.Command{
	Use:   "semgrep",
	Short: "Sync Semgrep community rules for SAST scanning",
	Long: `Sync Semgrep community rules from the official registry.

Downloads the latest community rules for static analysis security testing (SAST).
These rules cover common vulnerabilities like SQL injection, XSS, and more.

By default, only syncs if rules are older than 1 week.`,
	RunE: runFeedsSemgrep,
}

var feedsStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show feed status",
	Long:  `Display the sync status of all configured feeds.`,
	RunE:  runFeedsStatus,
}

var feedsRagCmd = &cobra.Command{
	Use:   "rag",
	Short: "Generate rules from RAG knowledge base",
	Long: `Generate Semgrep rules from the RAG (Retrieval-Augmented Generation) knowledge base.

This command converts patterns from the rag/ directory into Semgrep-compatible
rules for use by Zero's custom scanners (technology detection, secrets, etc.).

The RAG knowledge base contains human-readable patterns in markdown format
that are converted to executable Semgrep YAML rules.`,
	RunE: runFeedsRag,
}

func init() {
	rootCmd.AddCommand(feedsCmd)
	feedsCmd.AddCommand(feedsSemgrepCmd)
	feedsCmd.AddCommand(feedsRagCmd)
	feedsCmd.AddCommand(feedsStatusCmd)

	feedsSemgrepCmd.Flags().BoolVar(&feedsForce, "force", false, "Force sync even if rules are fresh")

	feedsRagCmd.Flags().BoolVar(&feedsForce, "force", false, "Force regenerate even if RAG unchanged")
}

func runFeedsSemgrep(cmd *cobra.Command, args []string) error {
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
	term.Info("%s", term.Color(terminal.Bold, "Syncing Semgrep Community Rules"))
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
			term.Success("  %s %s (%s, %d rules)",
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

func runFeedsRag(cmd *cobra.Command, args []string) error {
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
	term.Info("%s", term.Color(terminal.Bold, "Generating Rules from RAG Knowledge Base"))
	term.Divider()

	// Find RAG directory
	ragPath := rag.FindRAGPath()
	if ragPath == "" {
		return fmt.Errorf("could not find rag/ directory")
	}

	// Output directory for generated rules
	outputDir := filepath.Join(zeroHome, "rules", "generated")

	term.Info("RAG source: %s", term.Color(terminal.Cyan, ragPath))
	term.Info("Output dir: %s", term.Color(terminal.Cyan, outputDir))
	term.Info("")

	start := time.Now()

	// Convert RAG markdown patterns to Semgrep YAML rules
	result, err := techid.ConvertRAGToSemgrep(ragPath, outputDir)
	if err != nil {
		return fmt.Errorf("rule generation failed: %w", err)
	}

	duration := time.Since(start)

	// Display results
	if len(result.TechDiscovery.Rules) > 0 {
		term.Success("  %s tech-discovery.yaml (%d rules)",
			term.Color(terminal.Green, "✓"),
			len(result.TechDiscovery.Rules))
	}
	if len(result.Secrets.Rules) > 0 {
		term.Success("  %s secrets.yaml (%d rules)",
			term.Color(terminal.Green, "✓"),
			len(result.Secrets.Rules))
	}
	if len(result.AIML.Rules) > 0 {
		term.Success("  %s ai-ml.yaml (%d rules)",
			term.Color(terminal.Green, "✓"),
			len(result.AIML.Rules))
	}
	if len(result.ConfigFiles.Rules) > 0 {
		term.Success("  %s config-files.yaml (%d rules)",
			term.Color(terminal.Green, "✓"),
			len(result.ConfigFiles.Rules))
	}
	if len(result.Cryptography.Rules) > 0 {
		term.Success("  %s cryptography.yaml (%d rules)",
			term.Color(terminal.Green, "✓"),
			len(result.Cryptography.Rules))
	}
	if len(result.DevOps.Rules) > 0 {
		term.Success("  %s devops.yaml (%d rules)",
			term.Color(terminal.Green, "✓"),
			len(result.DevOps.Rules))
	}
	if len(result.CodeSecurity.Rules) > 0 {
		term.Success("  %s code-security.yaml (%d rules)",
			term.Color(terminal.Green, "✓"),
			len(result.CodeSecurity.Rules))
	}
	if len(result.SupplyChain.Rules) > 0 {
		term.Success("  %s supply-chain.yaml (%d rules)",
			term.Color(terminal.Green, "✓"),
			len(result.SupplyChain.Rules))
	}

	term.Divider()
	term.Success("Generated %d rules in %s", result.TotalRules, formatDuration(duration))

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
