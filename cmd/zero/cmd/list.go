package cmd

import (
	"fmt"

	"github.com/crashappsec/zero/pkg/scanner"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available scanners",
	Long:  `List all registered scanners and their descriptions.`,
	Run: func(cmd *cobra.Command, args []string) {
		term.Info("Available scanners:")
		fmt.Println()
		for _, name := range scanner.List() {
			s, ok := scanner.Get(name)
			if !ok {
				continue
			}
			fmt.Printf("  %-20s %s\n", name, s.Description())
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
