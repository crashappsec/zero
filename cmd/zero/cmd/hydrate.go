package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"sort"
	"strings"
	"syscall"

	"github.com/crashappsec/zero/pkg/core/config"
	"github.com/crashappsec/zero/pkg/workflow/hydrate"
	"github.com/spf13/cobra"
)

// validTargetPattern matches valid GitHub owner/repo or org names
var validTargetPattern = regexp.MustCompile(`^[a-zA-Z0-9][-a-zA-Z0-9_.]*(/[a-zA-Z0-9][-a-zA-Z0-9_.]*)?$`)

var hydrateOpts hydrate.Options

var hydrateCmd = &cobra.Command{
	Use:     "hydrate <target> [profile]",
	Aliases: []string{"onboard", "cache"},
	Short:   "Clone and analyze repositories",
	Long: `Clone a repository or organization and run analysis.

Target can be:
  - owner/repo    Single repository (e.g., strapi/strapi)
  - org-name      GitHub organization (e.g., zero-test-org)

The profile argument specifies which analyzers to run. Profiles are defined
in the config file (config/zero.config.json).

Aliases: onboard, cache

Examples:
  zero hydrate strapi/strapi              Clone and analyze a single repo
  zero onboard strapi/strapi              Same as hydrate (alias)
  zero hydrate strapi/strapi all-quick    Clone with all-quick profile
  zero hydrate zero-test-org              Clone and analyze all org repos
  zero hydrate zero-test-org --limit 10   Limit to first 10 repos
  zero hydrate zero-test-org --demo       Demo mode: skip repos > 50MB`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runHydrate,
}

func init() {
	rootCmd.AddCommand(hydrateCmd)

	// Org options
	hydrateCmd.Flags().IntVar(&hydrateOpts.Limit, "limit", 25, "Maximum repos to process (org mode)")
	hydrateCmd.Flags().BoolVar(&hydrateOpts.Demo, "demo", false, "Demo mode: skip repos > 50MB, fetch replacements")

	// Clone options
	hydrateCmd.Flags().StringVar(&hydrateOpts.Branch, "branch", "", "Clone specific branch")
	hydrateCmd.Flags().IntVar(&hydrateOpts.Depth, "depth", 1, "Shallow clone depth")
	hydrateCmd.Flags().BoolVar(&hydrateOpts.CloneOnly, "clone-only", false, "Clone without scanning")

	// Scan options
	hydrateCmd.Flags().BoolVar(&hydrateOpts.Force, "force", false, "Re-scan even if results exist")
	hydrateCmd.Flags().BoolVar(&hydrateOpts.SkipSlow, "skip-slow", false, "Skip slow scanners")
	hydrateCmd.Flags().BoolVarP(&hydrateOpts.Yes, "yes", "y", false, "Auto-accept prompts")
	hydrateCmd.Flags().IntVar(&hydrateOpts.ParallelScanners, "parallel", 4, "Parallel scanners per repo")
}

func runHydrate(cmd *cobra.Command, args []string) error {
	// Load config to get available profiles
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Parse target: owner/repo (single repo) or org-name (organization)
	target := args[0]

	// Validate target format
	if !validTargetPattern.MatchString(target) || len(target) > 150 {
		return fmt.Errorf("invalid target format: %q\nTarget must be a valid GitHub owner/repo or org name", target)
	}

	profile := cfg.Settings.DefaultProfile
	if profile == "" {
		profile = "all-quick"
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

	_, err = h.Run(ctx)
	return err
}
