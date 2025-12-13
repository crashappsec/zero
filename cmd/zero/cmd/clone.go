package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/crashappsec/zero/pkg/hydrate"
	"github.com/spf13/cobra"
)

var cloneOpts hydrate.Options

var cloneCmd = &cobra.Command{
	Use:   "clone [owner/repo]",
	Short: "Clone repositories without scanning",
	Long: `Clone repositories from GitHub without running scanners.

Examples:
  zero clone owner/repo              Clone single repo
  zero clone --org myorg             Clone all org repos
  zero clone --org myorg --limit 5   Clone first 5 repos
  zero clone owner/repo --depth 1    Shallow clone`,
	Args: cobra.MaximumNArgs(1),
	RunE: runClone,
}

func init() {
	rootCmd.AddCommand(cloneCmd)

	cloneCmd.Flags().StringVar(&cloneOpts.Org, "org", "", "GitHub organization")
	cloneCmd.Flags().IntVar(&cloneOpts.Limit, "limit", 100, "Maximum repos (org mode)")
	cloneCmd.Flags().StringVar(&cloneOpts.Branch, "branch", "", "Clone specific branch")
	cloneCmd.Flags().IntVar(&cloneOpts.Depth, "depth", 1, "Shallow clone depth")
	cloneCmd.Flags().BoolVar(&cloneOpts.Force, "force", false, "Re-clone even if exists")
	cloneCmd.Flags().IntVar(&cloneOpts.ParallelRepos, "parallel", 4, "Parallel repo cloning")
}

func runClone(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		cloneOpts.Repo = args[0]
	}

	if cloneOpts.Org == "" && cloneOpts.Repo == "" {
		return cmd.Help()
	}

	// Force clone-only mode
	cloneOpts.CloneOnly = true

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		term.Info("\nInterrupted, cleaning up...")
		cancel()
	}()

	h, err := hydrate.New(&cloneOpts)
	if err != nil {
		return err
	}

	_, err = h.Run(ctx)
	return err
}
