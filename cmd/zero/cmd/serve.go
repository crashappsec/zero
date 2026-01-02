package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/crashappsec/zero/pkg/api"
	"github.com/crashappsec/zero/pkg/core/terminal"
	"github.com/spf13/cobra"
)

var (
	servePort int
	serveDev  bool
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Zero web UI server",
	Long: `Start the Zero API server with an optional web UI.

The server provides:
  - REST API for projects, scans, and analysis data
  - WebSocket endpoints for real-time scan progress
  - Agent chat interface (coming soon)

Examples:
  zero serve                    # Start server on port 3001
  zero serve --port 8080        # Start on custom port
  zero serve --dev              # Enable CORS for frontend dev server`,
	RunE: runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().IntVarP(&servePort, "port", "p", 3001, "Port to listen on")
	serveCmd.Flags().BoolVar(&serveDev, "dev", false, "Enable development mode (CORS: *)")
}

func runServe(cmd *cobra.Command, args []string) error {
	// Create context that cancels on interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nShutting down server...")
		cancel()
	}()

	// Create and start server
	server, err := api.NewServer(&api.Options{
		Port:    servePort,
		DevMode: serveDev,
	})
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	// Print startup info
	term.Divider()
	term.Info("%s %s", term.Color(terminal.Green, "Zero API Server"), "v0.1.0-experimental")
	fmt.Println()
	term.Info("  API:       http://localhost:%d/api", servePort)
	term.Info("  Health:    http://localhost:%d/api/health", servePort)
	term.Info("  Projects:  http://localhost:%d/api/projects", servePort)
	fmt.Println()
	if serveDev {
		term.Info("  Mode: %s (CORS enabled for all origins)", term.Color(terminal.Yellow, "development"))
	} else {
		term.Info("  Mode: %s", term.Color(terminal.Green, "production"))
	}
	term.Divider()
	fmt.Println()
	term.Info("Press Ctrl+C to stop the server")
	fmt.Println()

	return server.Run(ctx)
}
