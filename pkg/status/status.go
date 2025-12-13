// Package status implements the status command for showing hydrated projects
package status

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/crashappsec/zero/pkg/config"
	"github.com/crashappsec/zero/pkg/terminal"
)

// Project represents a hydrated project
type Project struct {
	ID           string
	Owner        string
	Name         string
	RepoPath     string
	AnalysisPath string
	LastScan     time.Time
	ScanCount    int
	FileCount    int
	DiskSize     int64
	Scanners     []ScannerResult
}

// ScannerResult represents a scanner's results for a project
type ScannerResult struct {
	Name      string
	Status    string
	Timestamp time.Time
	Summary   string
}

// Options configures the status command
type Options struct {
	Org     string
	Verbose bool
	JSON    bool
}

// Status handles the status command
type Status struct {
	cfg      *config.Config
	term     *terminal.Terminal
	opts     *Options
	zeroHome string
}

// New creates a new Status instance
func New(opts *Options) (*Status, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	return &Status{
		cfg:      cfg,
		term:     terminal.New(),
		opts:     opts,
		zeroHome: cfg.ZeroHome(),
	}, nil
}

// Run executes the status command
func (s *Status) Run() error {
	projects, err := s.listProjects()
	if err != nil {
		return err
	}

	if len(projects) == 0 {
		s.term.Info("No hydrated projects found.")
		s.term.Info("Run: zero hydrate --org <org> to get started")
		return nil
	}

	// Filter by org if specified
	if s.opts.Org != "" {
		filtered := make([]*Project, 0)
		for _, p := range projects {
			if p.Owner == s.opts.Org {
				filtered = append(filtered, p)
			}
		}
		projects = filtered
	}

	if s.opts.JSON {
		return s.outputJSON(projects)
	}

	return s.outputTable(projects)
}

// listProjects finds all hydrated projects
func (s *Status) listProjects() ([]*Project, error) {
	reposPath := filepath.Join(s.zeroHome, "repos")

	if _, err := os.Stat(reposPath); os.IsNotExist(err) {
		return nil, nil
	}

	var projects []*Project

	// Walk the repos directory
	orgs, err := os.ReadDir(reposPath)
	if err != nil {
		return nil, fmt.Errorf("reading repos directory: %w", err)
	}

	for _, org := range orgs {
		if !org.IsDir() {
			continue
		}

		orgPath := filepath.Join(reposPath, org.Name())
		repos, err := os.ReadDir(orgPath)
		if err != nil {
			continue
		}

		for _, repo := range repos {
			if !repo.IsDir() {
				continue
			}

			project := s.loadProject(org.Name(), repo.Name())
			if project != nil {
				projects = append(projects, project)
			}
		}
	}

	// Sort by last scan time (most recent first)
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].LastScan.After(projects[j].LastScan)
	})

	return projects, nil
}

// loadProject loads project information from disk
func (s *Status) loadProject(owner, name string) *Project {
	projectID := fmt.Sprintf("%s/%s", owner, name)
	basePath := filepath.Join(s.zeroHome, "repos", owner, name)
	repoPath := filepath.Join(basePath, "repo")
	analysisPath := filepath.Join(basePath, "analysis")

	// Check if repo exists
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return nil
	}

	project := &Project{
		ID:           projectID,
		Owner:        owner,
		Name:         name,
		RepoPath:     repoPath,
		AnalysisPath: analysisPath,
	}

	// Get file count
	project.FileCount = countFiles(repoPath)

	// Get disk size
	project.DiskSize = getDirSize(basePath)

	// Load manifest for scan info
	manifestPath := filepath.Join(analysisPath, "manifest.json")
	if data, err := os.ReadFile(manifestPath); err == nil {
		var manifest struct {
			ScanID string `json:"scan_id"`
			Scan   struct {
				CompletedAt string   `json:"completed_at"`
				Completed   []string `json:"scanners_completed"`
				Failed      []string `json:"scanners_failed"`
			} `json:"scan"`
			Analyses map[string]struct {
				Status     string `json:"status"`
				DurationMS int    `json:"duration_ms"`
			} `json:"analyses"`
		}
		if err := json.Unmarshal(data, &manifest); err == nil {
			if t, err := time.Parse(time.RFC3339, manifest.Scan.CompletedAt); err == nil {
				project.LastScan = t
			}

			// Load scanner results from analyses
			for scannerName, analysis := range manifest.Analyses {
				result := ScannerResult{
					Name:   scannerName,
					Status: analysis.Status,
				}

				// Try to get summary from scanner output
				outputPath := filepath.Join(analysisPath, scannerName+".json")
				if summary := s.getScannerSummary(scannerName, outputPath); summary != "" {
					result.Summary = summary
				}

				project.Scanners = append(project.Scanners, result)
			}
		}
	}

	return project
}

// getScannerSummary extracts a summary from scanner output
func (s *Status) getScannerSummary(scanner, path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return ""
	}

	summary, ok := result["summary"].(map[string]interface{})
	if !ok {
		return ""
	}

	switch scanner {
	case "package-vulns":
		c := getInt(summary, "critical")
		h := getInt(summary, "high")
		m := getInt(summary, "medium")
		l := getInt(summary, "low")
		total := c + h + m + l
		if total == 0 {
			return "no vulnerabilities"
		}
		return fmt.Sprintf("%d vulns (%d critical, %d high)", total, c, h)

	case "package-sbom":
		if total, ok := summary["total_packages"].(float64); ok {
			return fmt.Sprintf("%.0f packages", total)
		}

	case "code-secrets":
		if total, ok := summary["total"].(float64); ok {
			if total == 0 {
				return "no secrets"
			}
			return fmt.Sprintf("%.0f secrets found", total)
		}

	case "code-vulns":
		if total, ok := summary["total"].(float64); ok {
			if total == 0 {
				return "no issues"
			}
			return fmt.Sprintf("%.0f issues", total)
		}
	}

	return ""
}

// outputTable prints projects in table format
func (s *Status) outputTable(projects []*Project) error {
	s.term.Divider()
	s.term.Info("%s", s.term.Color(terminal.Bold, "Hydrated Projects"))
	fmt.Println()

	// Group by org if no org filter
	if s.opts.Org == "" {
		byOrg := make(map[string][]*Project)
		for _, p := range projects {
			byOrg[p.Owner] = append(byOrg[p.Owner], p)
		}

		// Get sorted org names
		orgs := make([]string, 0, len(byOrg))
		for org := range byOrg {
			orgs = append(orgs, org)
		}
		sort.Strings(orgs)

		for _, org := range orgs {
			orgProjects := byOrg[org]
			s.term.Info("%s %s (%d repos)",
				s.term.Color(terminal.Cyan, "▸"),
				s.term.Color(terminal.Bold, org),
				len(orgProjects),
			)

			for _, p := range orgProjects {
				s.printProjectLine(p)
			}
			fmt.Println()
		}
	} else {
		for _, p := range projects {
			s.printProjectLine(p)
		}
	}

	// Summary
	s.term.Divider()
	totalSize := int64(0)
	totalFiles := 0
	for _, p := range projects {
		totalSize += p.DiskSize
		totalFiles += p.FileCount
	}

	s.term.Info("Total: %d projects, %s, %s files",
		len(projects),
		formatBytes(totalSize),
		formatNumber(totalFiles),
	)

	return nil
}

// printProjectLine prints a single project line
func (s *Status) printProjectLine(p *Project) {
	age := time.Since(p.LastScan)
	ageStr := formatAge(age)

	// Get key findings
	findings := s.getKeyFindings(p)

	if s.opts.Verbose {
		s.term.Info("    %s %s",
			s.term.Color(terminal.Green, "✓"),
			p.Name,
		)
		s.term.Info("      %s", s.term.Color(terminal.Dim, fmt.Sprintf("Last scan: %s ago | %s | %s files",
			ageStr, formatBytes(p.DiskSize), formatNumber(p.FileCount))))
		if findings != "" {
			s.term.Info("      %s", findings)
		}
	} else {
		s.term.Info("    %s %-25s %s %s",
			s.term.Color(terminal.Green, "✓"),
			p.Name,
			s.term.Color(terminal.Dim, fmt.Sprintf("(%s ago)", ageStr)),
			findings,
		)
	}
}

// getKeyFindings returns a summary of key findings
func (s *Status) getKeyFindings(p *Project) string {
	var parts []string

	for _, scanner := range p.Scanners {
		if scanner.Summary != "" && !strings.Contains(scanner.Summary, "no ") {
			parts = append(parts, scanner.Summary)
		}
	}

	if len(parts) == 0 {
		return s.term.Color(terminal.Green, "clean")
	}

	return strings.Join(parts, ", ")
}

// outputJSON prints projects as JSON
func (s *Status) outputJSON(projects []*Project) error {
	data, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

// Helper functions

func countFiles(path string) int {
	count := 0
	filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && !strings.Contains(p, ".git") {
			count++
		}
		return nil
	})
	return count
}

func getDirSize(path string) int64 {
	var size int64
	filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return 0
}

func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1fG", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.0fM", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.0fK", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}

func formatAge(d time.Duration) string {
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

func formatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	return fmt.Sprintf("%d,%03d", n/1000, n%1000)
}
