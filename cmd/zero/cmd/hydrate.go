package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"

	"github.com/crashappsec/zero/pkg/config"
	"github.com/crashappsec/zero/pkg/hydrate"
	"github.com/crashappsec/zero/pkg/report"
	"github.com/spf13/cobra"
)

var hydrateOpts hydrate.Options
var showReport bool

var hydrateCmd = &cobra.Command{
	Use:   "hydrate <target> [profile]",
	Short: "Clone and scan repositories",
	Long: `Clone a repository or organization and run security scanners.

Target can be:
  - owner/repo    Single repository (e.g., expressjs/express)
  - org-name      GitHub organization (e.g., phantom-tests)

The profile argument specifies which scanners to run. Profiles are defined
in the config file (config/zero.config.json).

Examples:
  zero hydrate expressjs/express          Clone and scan a single repo
  zero hydrate expressjs/express security Clone with security profile
  zero hydrate phantom-tests              Clone and scan all org repos
  zero hydrate phantom-tests quick        Clone org with quick profile
  zero hydrate phantom-tests --limit 10   Limit to first 10 repos`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runHydrate,
}

func init() {
	rootCmd.AddCommand(hydrateCmd)

	// Org options
	hydrateCmd.Flags().IntVar(&hydrateOpts.Limit, "limit", 100, "Maximum repos to process (org mode)")

	// Clone options
	hydrateCmd.Flags().StringVar(&hydrateOpts.Branch, "branch", "", "Clone specific branch")
	hydrateCmd.Flags().IntVar(&hydrateOpts.Depth, "depth", 1, "Shallow clone depth")
	hydrateCmd.Flags().BoolVar(&hydrateOpts.CloneOnly, "clone-only", false, "Clone without scanning")

	// Scan options
	hydrateCmd.Flags().BoolVar(&hydrateOpts.Force, "force", false, "Re-scan even if results exist")
	hydrateCmd.Flags().BoolVar(&hydrateOpts.SkipSlow, "skip-slow", false, "Skip slow scanners")
	hydrateCmd.Flags().BoolVarP(&hydrateOpts.Yes, "yes", "y", false, "Auto-accept prompts")
	hydrateCmd.Flags().IntVar(&hydrateOpts.ParallelScanners, "parallel", 4, "Parallel scanners per repo")

	// Report options
	hydrateCmd.Flags().BoolVar(&showReport, "report", false, "Show detailed findings report after scan")
}

func runHydrate(cmd *cobra.Command, args []string) error {
	// Load config to get available profiles
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Parse target: owner/repo (single repo) or org-name (organization)
	target := args[0]
	profile := cfg.Settings.DefaultProfile
	if profile == "" {
		profile = "standard"
	}

	// Check if target is org or repo based on presence of "/"
	if strings.Contains(target, "/") {
		// Single repo mode: owner/repo
		hydrateOpts.Repo = target
	} else {
		// Org mode: just the org name
		hydrateOpts.Org = target
	}

	// Second arg is profile if provided
	if len(args) > 1 {
		profile = args[1]
	}

	// Validate profile exists in config
	if _, ok := cfg.Profiles[profile]; !ok {
		availableProfiles := cfg.GetProfileNames()
		sort.Strings(availableProfiles)
		return fmt.Errorf("unknown profile: %s\n\nAvailable profiles:\n  %s\n\nProfiles are defined in config/zero.config.json",
			profile, strings.Join(availableProfiles, "\n  "))
	}

	hydrateOpts.Profile = profile

	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		term.Info("\nInterrupted, cleaning up...")
		cancel()
	}()

	h, err := hydrate.New(&hydrateOpts)
	if err != nil {
		return err
	}

	projectIDs, err := h.Run(ctx)
	if err != nil {
		return err
	}

	// Generate detailed report if requested
	if showReport && len(projectIDs) > 0 {
		reporter := report.NewReporter(cfg.ZeroHome())
		reporter.GenerateReport(projectIDs)
	}

	return nil
}
