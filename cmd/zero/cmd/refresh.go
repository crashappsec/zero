package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/crashappsec/zero/pkg/automation"
	"github.com/crashappsec/zero/pkg/config"
	"github.com/crashappsec/zero/pkg/freshness"
	"github.com/crashappsec/zero/pkg/scanner"
	"github.com/crashappsec/zero/pkg/terminal"
	"github.com/spf13/cobra"
)

var (
	refreshForce    bool
	refreshAll      bool
	refreshProfile  string
	refreshParallel int
)

var refreshCmd = &cobra.Command{
	Use:   "refresh [repo]",
	Short: "Refresh stale scan data",
	Long: `Refresh repositories with stale or outdated scan data.

By default, only refreshes repositories that need updating based on
freshness thresholds. Use --force to refresh regardless of staleness.

Examples:
  zero refresh                      Refresh all stale repos
  zero refresh owner/repo           Refresh specific repo
  zero refresh --force              Force refresh all repos
  zero refresh --all                Refresh all repos (not just stale)
  zero refresh --profile security   Use specific scan profile`,
	Args: cobra.MaximumNArgs(1),
	RunE: runRefresh,
}

func init() {
	rootCmd.AddCommand(refreshCmd)

	refreshCmd.Flags().BoolVar(&refreshForce, "force", false, "Force refresh even if data is fresh")
	refreshCmd.Flags().BoolVar(&refreshAll, "all", false, "Refresh all repos, not just stale ones")
	refreshCmd.Flags().StringVar(&refreshProfile, "profile", "", "Scan profile to use")
	refreshCmd.Flags().IntVar(&refreshParallel, "parallel", 4, "Number of parallel scans")
}

func runRefresh(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	zeroHome := cfg.ZeroHome()
	if zeroHome == "" {
		zeroHome = ".zero"
	}

	term := terminal.New()
	freshMgr := freshness.NewManager(zeroHome)

	// Determine profile
	profile := refreshProfile
	if profile == "" {
		profile = cfg.Settings.DefaultProfile
		if profile == "" {
			profile = "standard"
		}
	}

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

	// If specific repo provided
	if len(args) > 0 {
		return refreshRepo(ctx, term, cfg, freshMgr, args[0], profile)
	}

	// Get repos to refresh
	var repos []freshness.CheckResult
	var err2 error

	if refreshAll || refreshForce {
		repos, err2 = freshMgr.ListAll()
	} else {
		repos, err2 = freshMgr.ListStale()
	}

	if err2 != nil {
		return fmt.Errorf("listing repos: %w", err2)
	}

	if len(repos) == 0 {
		term.Success("All repositories are fresh!")
		return nil
	}

	term.Divider()
	term.Info("%s %d repositories",
		term.Color(terminal.Bold, "Refreshing"),
		len(repos),
	)
	term.Divider()

	// Show what will be refreshed
	for _, r := range repos {
		levelColor := terminal.Green
		switch r.Level {
		case freshness.LevelStale:
			levelColor = terminal.Yellow
		case freshness.LevelVeryStale:
			levelColor = terminal.Red
		case freshness.LevelExpired:
			levelColor = terminal.Red
		}
		term.Info("  %s %s (%s)",
			term.Color(levelColor, "●"),
			r.Repository,
			r.AgeString,
		)
	}
	term.Divider()

	// Run refreshes
	runner := scanner.NewRunner(zeroHome)
	success := 0
	failed := 0
	skipped := 0

	for _, r := range repos {
		select {
		case <-ctx.Done():
			term.Info("Cancelled")
			return nil
		default:
		}

		// Check freshness unless forcing
		if !refreshForce {
			shouldScan, reason, _ := freshMgr.ShouldScan(r.Repository, true)
			if !shouldScan {
				term.Info("  %s %s - skipped (%s)",
					term.Color(terminal.Dim, "○"),
					r.Repository,
					reason,
				)
				skipped++
				continue
			}
		}

		term.Info("  %s %s",
			term.Color(terminal.Cyan, "▸"),
			r.Repository,
		)

		scanners, _ := cfg.GetProfileScanners(profile)
		progress := scanner.NewProgress(scanners)
		start := time.Now()

		result, err := runner.Run(ctx, r.Repository, profile, progress, nil)
		duration := time.Since(start)

		if err != nil {
			term.Error("    Failed: %v", err)
			failed++
			continue
		}

		if result.Success {
			// Record successful scan
			scanResults := make([]freshness.ScanResult, 0)
			for name, sr := range result.Results {
				errStr := ""
				if sr.Error != nil {
					errStr = sr.Error.Error()
				}
				scanResults = append(scanResults, freshness.ScanResult{
					Name:     name,
					Success:  sr.Status == scanner.StatusComplete,
					Duration: sr.Duration,
					Error:    errStr,
				})
			}
			freshMgr.RecordScan(r.Repository, scanResults)

			term.Success("    Complete (%ds)", int(duration.Seconds()))
			success++
		} else {
			term.Error("    Failed")
			failed++
		}
	}

	term.Divider()
	if failed > 0 {
		term.Info("Complete: %d refreshed, %d failed, %d skipped", success, failed, skipped)
	} else {
		term.Success("Complete: %d refreshed, %d skipped", success, skipped)
	}

	return nil
}

func refreshRepo(ctx context.Context, term *terminal.Terminal, cfg *config.Config, freshMgr *freshness.Manager, repo, profile string) error {
	zeroHome := cfg.ZeroHome()
	if zeroHome == "" {
		zeroHome = ".zero"
	}

	// Check if repo exists
	repoPath := filepath.Join(zeroHome, "repos", repo)
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return fmt.Errorf("repo not found: %s (run hydrate first)", repo)
	}

	// Check freshness
	if !refreshForce {
		shouldScan, reason, _ := freshMgr.ShouldScan(repo, true)
		if !shouldScan {
			term.Success("%s is %s", repo, reason)
			return nil
		}
		term.Info("%s: %s", repo, reason)
	}

	term.Divider()
	term.Info("%s %s",
		term.Color(terminal.Bold, "Refreshing"),
		repo,
	)
	term.Divider()

	runner := scanner.NewRunner(zeroHome)
	scanners, _ := cfg.GetProfileScanners(profile)
	progress := scanner.NewProgress(scanners)
	start := time.Now()

	result, err := runner.Run(ctx, repo, profile, progress, nil)
	duration := time.Since(start)

	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	if result.Success {
		// Record successful scan
		scanResults := make([]freshness.ScanResult, 0)
		for name, sr := range result.Results {
			errStr := ""
			if sr.Error != nil {
				errStr = sr.Error.Error()
			}
			scanResults = append(scanResults, freshness.ScanResult{
				Name:     name,
				Success:  sr.Status == scanner.StatusComplete,
				Duration: sr.Duration,
				Error:    errStr,
			})
		}
		freshMgr.RecordScan(repo, scanResults)

		term.Divider()
		term.Success("Refreshed %s (%ds)", repo, int(duration.Seconds()))
	} else {
		term.Divider()
		term.Error("Failed to refresh %s", repo)
	}

	return nil
}

// RefreshScannerFunc adapts the scanner runner for the automation package
func RefreshScannerFunc(zeroHome string, cfg *config.Config, profile string) automation.ScannerFunc {
	return func(ctx context.Context, repo, scannerName string) (*automation.ScannerRun, error) {
		runner := scanner.NewRunner(zeroHome)
		progress := scanner.NewProgress([]string{scannerName})
		start := time.Now()

		result, err := runner.Run(ctx, repo, profile, progress, nil)
		if err != nil {
			return &automation.ScannerRun{
				Name:    scannerName,
				Success: false,
				Error:   err.Error(),
			}, err
		}

		// Find the scanner result
		if sr, ok := result.Results[scannerName]; ok {
			errStr := ""
			if sr.Error != nil {
				errStr = sr.Error.Error()
			}
			return &automation.ScannerRun{
				Name:     scannerName,
				Success:  sr.Status == scanner.StatusComplete,
				Duration: time.Since(start),
				Error:    errStr,
			}, nil
		}

		return &automation.ScannerRun{
			Name:    scannerName,
			Success: false,
			Error:   "scanner not found in results",
		}, nil
	}
}
