// Package evidence provides Evidence.dev report generation
package evidence

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/crashappsec/zero/pkg/terminal"
)

// Generator handles Evidence report generation
type Generator struct {
	zeroHome     string
	templatePath string
	term         *terminal.Terminal
}

// Options configures report generation
type Options struct {
	Repository  string
	OutputDir   string
	OpenBrowser bool
	DevServer   bool
	Force       bool
}

// NewGenerator creates a new Evidence generator
func NewGenerator(zeroHome string) *Generator {
	// Check for environment variable first (Docker deployment)
	templatePath := os.Getenv("ZERO_TEMPLATE_PATH")

	if templatePath == "" {
		// Template is bundled with Zero (development)
		templatePath = filepath.Join(zeroHome, "..", "reports", "template")

		// Check if running from source or installed
		if _, err := os.Stat(templatePath); os.IsNotExist(err) {
			// Try relative to executable
			exe, _ := os.Executable()
			templatePath = filepath.Join(filepath.Dir(exe), "reports", "template")
		}
	}

	return &Generator{
		zeroHome:     zeroHome,
		templatePath: templatePath,
		term:         terminal.New(),
	}
}

// ReportPath returns the path to the generated report
func (g *Generator) ReportPath(repo string) string {
	return filepath.Join(g.zeroHome, "repos", repo, "report", "index.html")
}

// AnalysisPath returns the path to analysis JSON files
func (g *Generator) AnalysisPath(repo string) string {
	return filepath.Join(g.zeroHome, "repos", repo, "analysis")
}

// Generate creates an Evidence report for a repository
func (g *Generator) Generate(opts Options) (string, error) {
	analysisPath := g.AnalysisPath(opts.Repository)

	// Check analysis data exists
	if _, err := os.Stat(analysisPath); os.IsNotExist(err) {
		return "", fmt.Errorf("no analysis data found for %s", opts.Repository)
	}

	// Determine output directory
	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = filepath.Join(g.zeroHome, "repos", opts.Repository, "report")
	}

	// Create working directory for Evidence build
	workDir := filepath.Join(g.zeroHome, "repos", opts.Repository, ".evidence-build")

	// Copy template to working directory
	if err := g.setupWorkDir(workDir, analysisPath); err != nil {
		return "", fmt.Errorf("setting up Evidence workspace: %w", err)
	}

	// Check for Node.js
	if !g.hasNode() {
		return "", fmt.Errorf("Node.js 18+ required for report generation. Install from https://nodejs.org")
	}

	// Install dependencies if needed
	if err := g.ensureDependencies(workDir); err != nil {
		return "", fmt.Errorf("installing dependencies: %w", err)
	}

	// Build or serve
	if opts.DevServer {
		return "", g.serve(workDir)
	}

	if err := g.build(workDir, outputDir); err != nil {
		return "", fmt.Errorf("building report: %w", err)
	}

	reportPath := filepath.Join(outputDir, "index.html")

	// Open browser if requested
	if opts.OpenBrowser {
		g.OpenBrowser(reportPath)
	}

	return reportPath, nil
}

// setupWorkDir prepares the Evidence working directory
func (g *Generator) setupWorkDir(workDir, analysisPath string) error {
	// Remove old work dir
	os.RemoveAll(workDir)

	// Create work dir
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return err
	}

	// Copy template files
	if err := copyDir(g.templatePath, workDir); err != nil {
		return fmt.Errorf("copying template: %w", err)
	}

	// Create symlink to analysis data
	dataLink := filepath.Join(workDir, "sources", "zero", "data")
	os.Remove(dataLink) // Remove if exists

	// Use relative path for symlink
	relPath, err := filepath.Rel(filepath.Join(workDir, "sources", "zero"), analysisPath)
	if err != nil {
		relPath = analysisPath // Fall back to absolute
	}

	if err := os.Symlink(relPath, dataLink); err != nil {
		// Fall back to copy if symlink fails (Windows)
		return copyDir(analysisPath, dataLink)
	}

	return nil
}

// hasNode checks if Node.js is available
func (g *Generator) hasNode() bool {
	cmd := exec.Command("node", "--version")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	// Check version >= 18
	var major int
	fmt.Sscanf(string(output), "v%d", &major)
	return major >= 18
}

// ensureDependencies installs npm dependencies
func (g *Generator) ensureDependencies(workDir string) error {
	nodeModules := filepath.Join(workDir, "node_modules")
	if _, err := os.Stat(nodeModules); err == nil {
		return nil // Already installed
	}

	g.term.Info("Installing report dependencies...")

	cmd := exec.Command("npm", "install", "--silent")
	cmd.Dir = workDir
	cmd.Stdout = nil
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// build runs Evidence build
func (g *Generator) build(workDir, outputDir string) error {
	// First run sources to process JavaScript data files
	g.term.Info("Processing data sources...")

	sourcesCmd := exec.Command("npm", "run", "sources")
	sourcesCmd.Dir = workDir
	sourcesCmd.Stdout = nil
	sourcesCmd.Stderr = os.Stderr

	if err := sourcesCmd.Run(); err != nil {
		// Sources might fail if no data, continue anyway
		g.term.Warn("Sources processing had warnings (continuing)")
	}

	g.term.Info("Building report...")

	cmd := exec.Command("npx", "evidence", "build")
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(), "NODE_ENV=production")
	cmd.Stdout = nil
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	// Copy build output to final location
	buildDir := filepath.Join(workDir, "build")
	return copyDir(buildDir, outputDir)
}

// serve starts Evidence dev server
func (g *Generator) serve(workDir string) error {
	g.term.Info("Starting report dev server...")
	g.term.Info("Press Ctrl+C to stop")

	cmd := exec.Command("npx", "evidence", "dev")
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// ServeAndOpen starts an HTTP server and opens the browser
func (g *Generator) ServeAndOpen(reportDir string) error {
	// Find an available port
	port, err := findAvailablePort()
	if err != nil {
		return fmt.Errorf("finding available port: %w", err)
	}

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	url := fmt.Sprintf("http://%s", addr)

	// Create file server
	fs := http.FileServer(http.Dir(reportDir))
	server := &http.Server{
		Addr:    addr,
		Handler: fs,
	}

	// Start server in background
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			g.term.Error("Server error: %v", err)
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Open browser
	g.openURL(url)

	g.term.Info("Report server running at %s", url)
	g.term.Info("Press Ctrl+C to stop")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)

	g.term.Info("Server stopped")
	return nil
}

// OpenBrowser starts HTTP server and opens the report (blocking)
func (g *Generator) OpenBrowser(reportPath string) {
	reportDir := filepath.Dir(reportPath)
	if err := g.ServeAndOpen(reportDir); err != nil {
		g.term.Error("Failed to serve report: %v", err)
	}
}

// openURL opens a URL in the default browser
func (g *Generator) openURL(url string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return
	}

	cmd.Start()
}

// findAvailablePort finds an available TCP port
func findAvailablePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

// copyDir recursively copies a directory, skipping node_modules
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// Skip node_modules - will be installed separately via npm
		if info.IsDir() && info.Name() == "node_modules" {
			return filepath.SkipDir
		}

		dstPath := filepath.Join(dst, relPath)

		// Handle symlinks
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(target, dstPath)
		}

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy file
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dstPath, data, info.Mode())
	})
}
