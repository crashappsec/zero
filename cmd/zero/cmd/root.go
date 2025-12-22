// Package cmd implements the zero CLI commands
package cmd

import (
	"fmt"
	"os"

	"github.com/crashappsec/zero/pkg/core/terminal"
	"github.com/spf13/cobra"

	// Import scanners to register them
	_ "github.com/crashappsec/zero/pkg/scanner"
)

var (
	// Global flags
	verbose bool
	noColor bool

	// Terminal instance
	term *terminal.Terminal
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "zero",
	Short: "Zero - Security analysis for repositories",
	Long: `Zero provides security analysis tools and specialist AI agents for repository assessment.
Named after characters from the movie Hackers (1995) - "Hack the planet!"

Quick Start:
  zero hydrate owner/repo           Clone and scan a repository
  zero hydrate --org myorg          Clone and scan all org repos
  zero status                       Show analyzed projects
  zero report owner/repo            Generate security report`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if noColor {
			os.Setenv("NO_COLOR", "1")
		}
		term = terminal.New()
	},
	Run: func(cmd *cobra.Command, args []string) {
		printBanner()
		fmt.Println()
		cmd.Help()
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")
}

func printBanner() {
	banner := `
  ███████╗███████╗██████╗  ██████╗
  ╚══███╔╝██╔════╝██╔══██╗██╔═══██╗
    ███╔╝ █████╗  ██████╔╝██║   ██║
   ███╔╝  ██╔══╝  ██╔══██╗██║   ██║
  ███████╗███████╗██║  ██║╚██████╔╝
  ╚══════╝╚══════╝╚═╝  ╚═╝ ╚═════╝`

	fmt.Print("\033[0;32m")
	fmt.Println(banner)
	fmt.Print("\033[2m  crashoverride.com\033[0m\n")
}
