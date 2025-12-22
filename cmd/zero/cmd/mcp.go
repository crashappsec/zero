package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/crashappsec/zero/pkg/core/config"
	"github.com/crashappsec/zero/pkg/mcp"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start the MCP server for Claude Desktop integration",
	Long: `Start the Model Context Protocol (MCP) server.

The MCP server provides analysis data to Claude Desktop and other MCP clients.
It exposes tools for querying vulnerabilities, packages, licenses, and other
scan results from hydrated repositories.

Add to your Claude Desktop config:
{
  "mcpServers": {
    "zero": {
      "command": "/path/to/zero",
      "args": ["mcp"]
    }
  }
}`,
	RunE: runMCP,
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}

func runMCP(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	server := mcp.NewServer(cfg.ZeroHome())
	return server.Run(ctx)
}
