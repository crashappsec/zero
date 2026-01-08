package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/crashappsec/zero/pkg/core/config"
	"github.com/crashappsec/zero/pkg/storage"
	"github.com/crashappsec/zero/pkg/storage/sqlite"
	"github.com/spf13/cobra"
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database management commands",
	Long:  `Commands for managing the SQLite database cache.`,
}

var dbSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync JSON analysis files to SQLite database",
	Long: `Synchronize all project analysis data from JSON files to SQLite.

This populates the database with:
- Project metadata
- Vulnerability counts
- Secret findings
- Package statistics

After syncing, API queries will be significantly faster.`,
	RunE: runDBSync,
}

var dbStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show database statistics",
	RunE:  runDBStats,
}

var dbResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset the database (delete and recreate)",
	RunE:  runDBReset,
}

func init() {
	rootCmd.AddCommand(dbCmd)
	dbCmd.AddCommand(dbSyncCmd)
	dbCmd.AddCommand(dbStatsCmd)
	dbCmd.AddCommand(dbResetCmd)
}

func runDBSync(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	zeroHome := cfg.ZeroHome()
	dbPath := filepath.Join(zeroHome, "zero.db")

	fmt.Printf("Syncing analysis data to %s...\n", dbPath)

	store, err := sqlite.New(dbPath)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Scan repos directory
	reposDir := filepath.Join(zeroHome, "repos")
	orgs, err := os.ReadDir(reposDir)
	if err != nil {
		return fmt.Errorf("reading repos directory: %w", err)
	}

	var synced, failed int
	for _, org := range orgs {
		if !org.IsDir() {
			continue
		}

		orgPath := filepath.Join(reposDir, org.Name())
		repos, err := os.ReadDir(orgPath)
		if err != nil {
			continue
		}

		for _, repo := range repos {
			if !repo.IsDir() {
				continue
			}

			projectID := fmt.Sprintf("%s/%s", org.Name(), repo.Name())
			repoPath := filepath.Join(orgPath, repo.Name())
			analysisPath := filepath.Join(repoPath, "analysis")

			// Check if analysis exists
			if _, err := os.Stat(analysisPath); os.IsNotExist(err) {
				continue
			}

			fmt.Printf("  Syncing %s...", projectID)

			// Create/update project (freshness will be calculated later)
			project := &storage.Project{
				ID:             projectID,
				Owner:          org.Name(),
				Name:           repo.Name(),
				RepoPath:       filepath.Join(repoPath, "repo"),
				AnalysisPath:   analysisPath,
				FreshnessLevel: "unknown",
				FreshnessAge:   0,
			}

			// Try to read freshness from freshness.json
			freshnessPath := filepath.Join(repoPath, "freshness.json")
			if data, err := os.ReadFile(freshnessPath); err == nil {
				var meta struct {
					LastScan time.Time `json:"last_scan"`
				}
				if json.Unmarshal(data, &meta) == nil && !meta.LastScan.IsZero() {
					project.LastScan = meta.LastScan
					age := time.Since(meta.LastScan)
					project.FreshnessAge = int(age.Hours())
					// Determine freshness level
					switch {
					case age < 24*time.Hour:
						project.FreshnessLevel = "fresh"
					case age < 7*24*time.Hour:
						project.FreshnessLevel = "stale"
					case age < 30*24*time.Hour:
						project.FreshnessLevel = "very-stale"
					default:
						project.FreshnessLevel = "expired"
					}
				}
			}

			if err := store.UpsertProject(ctx, project); err != nil {
				fmt.Printf(" FAILED: %v\n", err)
				failed++
				continue
			}

			// Sync analysis data
			if err := store.SyncProjectFromJSON(ctx, projectID, analysisPath); err != nil {
				fmt.Printf(" FAILED: %v\n", err)
				failed++
				continue
			}

			fmt.Println(" OK")
			synced++
		}
	}

	fmt.Printf("\nSync complete: %d projects synced, %d failed\n", synced, failed)

	// Show stats
	stats, err := store.GetAggregateStats(ctx)
	if err == nil {
		fmt.Printf("\nDatabase stats:\n")
		fmt.Printf("  Projects: %d\n", stats.TotalProjects)
		fmt.Printf("  Vulnerabilities: %d (critical: %d, high: %d, medium: %d, low: %d)\n",
			stats.TotalVulns, stats.VulnsBySeverity["critical"], stats.VulnsBySeverity["high"],
			stats.VulnsBySeverity["medium"], stats.VulnsBySeverity["low"])
		fmt.Printf("  Secrets: %d\n", stats.TotalSecrets)
		fmt.Printf("  Packages: %d\n", stats.TotalPackages)
	}

	return nil
}

func runDBStats(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	dbPath := filepath.Join(cfg.ZeroHome(), "zero.db")

	// Check if DB exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Println("Database not found. Run 'zero db sync' to create it.")
		return nil
	}

	store, err := sqlite.New(dbPath)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Get file info
	fileInfo, _ := os.Stat(dbPath)

	fmt.Printf("Database: %s\n", dbPath)
	fmt.Printf("Size: %.2f MB\n", float64(fileInfo.Size())/(1024*1024))
	fmt.Printf("Modified: %s\n", fileInfo.ModTime().Format(time.RFC3339))

	stats, err := store.GetAggregateStats(ctx)
	if err != nil {
		return fmt.Errorf("getting stats: %w", err)
	}

	fmt.Printf("\nContent:\n")
	fmt.Printf("  Projects: %d\n", stats.TotalProjects)
	fmt.Printf("  Vulnerabilities: %d\n", stats.TotalVulns)
	fmt.Printf("    Critical: %d\n", stats.VulnsBySeverity["critical"])
	fmt.Printf("    High: %d\n", stats.VulnsBySeverity["high"])
	fmt.Printf("    Medium: %d\n", stats.VulnsBySeverity["medium"])
	fmt.Printf("    Low: %d\n", stats.VulnsBySeverity["low"])
	fmt.Printf("  Secrets: %d\n", stats.TotalSecrets)
	fmt.Printf("  Packages: %d\n", stats.TotalPackages)
	fmt.Printf("  Technologies: %d\n", stats.TotalTechnologies)

	fmt.Printf("\nFreshness:\n")
	for level, count := range stats.FreshnessCounts {
		fmt.Printf("  %s: %d\n", level, count)
	}

	return nil
}

func runDBReset(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	dbPath := filepath.Join(cfg.ZeroHome(), "zero.db")

	// Check if DB exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Println("Database does not exist.")
		return nil
	}

	fmt.Printf("This will delete the database at %s\n", dbPath)
	fmt.Print("Are you sure? (y/N): ")

	var confirm string
	fmt.Scanln(&confirm)
	if confirm != "y" && confirm != "Y" {
		fmt.Println("Cancelled.")
		return nil
	}

	// Remove database files
	os.Remove(dbPath)
	os.Remove(dbPath + "-wal")
	os.Remove(dbPath + "-shm")

	fmt.Println("Database reset. Run 'zero db sync' to recreate.")
	return nil
}
