package cmd

import (
	"github.com/crashappsec/zero/pkg/status"
	"github.com/spf13/cobra"
)

var statusOpts status.Options

var statusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"list", "ls"},
	Short:   "Show hydrated projects",
	Long: `List all repositories that have been cloned and analyzed.

Examples:
  zero status                Show all projects
  zero status --org myorg    Show projects for specific org
  zero status --json         Output as JSON`,
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)

	statusCmd.Flags().StringVar(&statusOpts.Org, "org", "", "Filter by organization")
	statusCmd.Flags().BoolVarP(&statusOpts.Verbose, "verbose", "v", false, "Show detailed output")
	statusCmd.Flags().BoolVar(&statusOpts.JSON, "json", false, "Output as JSON")
}

func runStatus(cmd *cobra.Command, args []string) error {
	s, err := status.New(&statusOpts)
	if err != nil {
		return err
	}
	return s.Run()
}
