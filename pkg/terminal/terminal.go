// Package terminal provides colored output and progress display
package terminal

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/term"
)

// ANSI color codes
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Red       = "\033[0;31m"
	Green     = "\033[0;32m"
	Yellow    = "\033[1;33m"
	Blue      = "\033[0;34m"
	Cyan      = "\033[0;36m"
	White     = "\033[0;37m"
	BoldRed   = "\033[1;31m"
	BoldGreen = "\033[1;32m"
)

// Icons for status display
const (
	IconSuccess  = "✓"
	IconFailed   = "✗"
	IconRunning  = "◐"
	IconQueued   = "○"
	IconSkipped  = "⊘"
	IconWarning  = "⚠"
	IconArrow    = "▸"
)

// Terminal provides thread-safe terminal output
type Terminal struct {
	mu       sync.Mutex
	noColor  bool
	width    int
}

// New creates a new Terminal instance
func New() *Terminal {
	width := 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		width = w
	}

	return &Terminal{
		noColor: os.Getenv("NO_COLOR") != "",
		width:   width,
	}
}

// Color wraps text in color codes if colors are enabled
func (t *Terminal) Color(code, text string) string {
	if t.noColor {
		return text
	}
	return code + text + Reset
}

// Success prints a success message
func (t *Terminal) Success(format string, args ...interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("  %s %s\n", t.Color(Green, IconSuccess), msg)
}

// Error prints an error message
func (t *Terminal) Error(format string, args ...interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("  %s %s\n", t.Color(Red, IconFailed), msg)
}

// Warning prints a warning message
func (t *Terminal) Warning(format string, args ...interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s  %s\n", t.Color(Yellow, IconWarning), t.Color(Bold, msg))
}

// Warn is an alias for Warning
func (t *Terminal) Warn(format string, args ...interface{}) {
	t.Warning(format, args...)
}

// Info prints an info message
func (t *Terminal) Info(format string, args ...interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Printf(format+"\n", args...)
}

// Header prints a section header
func (t *Terminal) Header(text string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Printf("\n%s\n\n", t.Color(Bold, text))
}

// SubHeader prints a sub-section header
func (t *Terminal) SubHeader(text string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Printf("%s %s\n", t.Color(Bold, text), t.Color(Dim, "(depth=1)"))
}

// Divider prints a horizontal line
func (t *Terminal) Divider() {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Println(strings.Repeat("━", min(t.width, 78)))
}

// Box prints text in a decorative box
func (t *Terminal) Box(text string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	width := len(text) + 4
	if width > t.width {
		width = t.width
	}
	border := strings.Repeat("─", width-2)
	fmt.Printf("  ╭%s╮\n", border)
	fmt.Printf("  │ %s │\n", text)
	fmt.Printf("  ╰%s╯\n", border)
}

// RepoCloned prints a cloned repo result
func (t *Terminal) RepoCloned(name, size, files, commit, status string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Printf("  %s %s %s %s\n",
		t.Color(Green, IconSuccess),
		name,
		t.Color(Dim, fmt.Sprintf("%s, %s files", size, files)),
		t.Color(Dim, fmt.Sprintf("(%s %s)", commit, status)),
	)
}

// RepoScanning prints a repo that's being scanned
func (t *Terminal) RepoScanning(name string, estimate int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if estimate > 10 {
		fmt.Printf("  %s %s %s\n",
			t.Color(Yellow, IconRunning),
			t.Color(Bold, name),
			t.Color(Dim, fmt.Sprintf("scanning (~%ds)...", estimate)),
		)
	} else {
		fmt.Printf("  %s %s %s\n",
			t.Color(Yellow, IconRunning),
			t.Color(Bold, name),
			t.Color(Dim, "scanning..."),
		)
	}
}

// ScannerRunning prints a running scanner line
func (t *Terminal) ScannerRunning(name string, estimate int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Printf("      %s %-20s %s %s\n",
		t.Color(Cyan, IconArrow),
		name,
		t.Color(Cyan, "running"),
		t.Color(Dim, fmt.Sprintf("~%ds", estimate)),
	)
}

// ScannerQueued prints a queued scanner line
func (t *Terminal) ScannerQueued(name string, estimate int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Printf("      %s %-20s %s\n",
		t.Color(Dim, IconQueued),
		t.Color(Dim, name),
		t.Color(Dim, fmt.Sprintf("queued  ~%ds", estimate)),
	)
}

// ScannerSkipped prints a skipped scanner line
func (t *Terminal) ScannerSkipped(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Printf("      %s %-20s %s\n",
		t.Color(Dim, IconSkipped),
		t.Color(Dim, name),
		t.Color(Dim, "skipped"),
	)
}

// ScannerComplete prints a completed scanner result
func (t *Terminal) ScannerComplete(name, summary string, duration int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Printf("      %s %-20s %s %s\n",
		t.Color(Green, IconSuccess),
		name,
		summary,
		t.Color(Dim, fmt.Sprintf("%ds", duration)),
	)
}

// ScannerFailed prints a failed scanner result
func (t *Terminal) ScannerFailed(name, errMsg string, duration int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Printf("      %s %-20s %s %s\n",
		t.Color(Red, IconFailed),
		name,
		t.Color(Red, errMsg),
		t.Color(Dim, fmt.Sprintf("%ds", duration)),
	)
}

// UpdateScannerStatus updates a scanner line in place (moves cursor up and rewrites)
func (t *Terminal) UpdateScannerStatus(linesUp int, name string, status string, icon string, color string, extra string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	// Move cursor up N lines, clear line, print, move back down
	fmt.Printf("\033[%dA\033[K      %s %-20s %s %s\033[%dB\r",
		linesUp,
		t.Color(color, icon),
		name,
		t.Color(color, status),
		t.Color(Dim, extra),
		linesUp,
	)
}

// LogScannerStatus logs a scanner status message on a new line (doesn't update in place)
func (t *Terminal) LogScannerStatus(name string, status string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Printf("          %s %s: %s\n",
		t.Color(Dim, IconArrow),
		t.Color(Dim, name),
		t.Color(Dim, status),
	)
}

// Progress prints an in-place progress line (overwrites current line)
func (t *Terminal) Progress(completed, total int, active string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Printf("\r\033[K  %s",
		t.Color(Dim, fmt.Sprintf("[%d/%d scanners] %s", completed, total, active)),
	)
}

// ClearLine clears the current line
func (t *Terminal) ClearLine() {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Print("\r\033[K")
}

// RepoComplete prints a completed repo header
func (t *Terminal) RepoComplete(name string, success bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if success {
		fmt.Printf("  %s %s\n", t.Color(Green, IconSuccess), t.Color(Bold, name))
	} else {
		fmt.Printf("  %s %s %s\n", t.Color(Red, IconFailed), t.Color(Bold, name), t.Color(Red, "(scan failed)"))
	}
}

// ScanComplete prints the final scanning complete message
func (t *Terminal) ScanComplete() {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Printf("%s %s\n", t.Color(Green, IconSuccess), t.Color(Bold, "Scanning complete"))
}

// ScanFindings holds aggregated scan findings
type ScanFindings struct {
	// Scanners that were run (used to conditionally show sections)
	ScannersRun     map[string]bool

	// SBOM
	SBOMPath        string   // Path to generated SBOM file(s)
	SBOMPaths       []string // Multiple SBOM paths for multi-repo scans
	SBOMSizeTotal   int64    // Total size of all SBOM files in bytes

	// Packages
	TotalPackages   int
	PackagesByEco   map[string]int

	// Vulnerabilities
	VulnCritical    int
	VulnHigh        int
	VulnMedium      int
	VulnLow         int
	VulnsByEco      map[string]int

	// Licenses
	LicenseTypes    int
	LicenseCounts   map[string]int

	// Secrets
	SecretsCritical int
	SecretsHigh     int
	SecretsMedium   int
	SecretsTotal    int

	// Malcontent
	MalcontentCrit  int
	MalcontentHigh  int

	// Health
	HealthCritical  int
	HealthWarnings  int

	// Tech-ID (Technology Detection)
	TechTotalTechs    int              // Total unique technologies detected
	TechByCategory    map[string]int   // Technologies by category (language, framework, etc.)
	TechTopList       []string         // Top technologies across all repos
	TechMLModels      int              // ML models detected
	TechMLFrameworks  int              // AI/ML frameworks detected
	TechSecurityCount int              // Security findings from AI/ML analysis
}

// Summary prints the hydrate summary
func (t *Terminal) Summary(org string, duration int, success, failed int, diskUsage, files string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	fmt.Printf("\n%s\n\n", t.Color(Bold+Green, "✓ Hydrate Complete"))

	fmt.Printf("%s\n", t.Color(Bold, "Summary"))
	fmt.Printf("  Organization:    %s\n", t.Color(Cyan, org))
	fmt.Printf("  Duration:        %ds\n", duration)
	if failed > 0 {
		fmt.Printf("  Repos scanned:   %s, %s\n",
			t.Color(Green, fmt.Sprintf("%d success", success)),
			t.Color(Red, fmt.Sprintf("%d failed", failed)),
		)
	} else {
		fmt.Printf("  Repos scanned:   %s\n", t.Color(Green, fmt.Sprintf("%d success", success)))
	}
	fmt.Printf("  Disk usage:      %s\n", diskUsage)
	fmt.Printf("  Total files:     %s\n", files)
}

// SummaryWithFindings prints the hydrate summary with aggregated findings
func (t *Terminal) SummaryWithFindings(org string, duration int, success, failed int, diskUsage, files string, findings *ScanFindings) {
	t.mu.Lock()
	defer t.mu.Unlock()

	fmt.Printf("\n%s\n\n", t.Color(Bold+Green, "✓ Hydrate Complete"))

	fmt.Printf("%s\n", t.Color(Bold, "Summary"))
	fmt.Printf("  Organization:    %s\n", t.Color(Cyan, org))
	fmt.Printf("  Duration:        %ds\n", duration)
	if failed > 0 {
		fmt.Printf("  Repos scanned:   %s, %s\n",
			t.Color(Green, fmt.Sprintf("%d success", success)),
			t.Color(Red, fmt.Sprintf("%d failed", failed)),
		)
	} else {
		fmt.Printf("  Repos scanned:   %s\n", t.Color(Green, fmt.Sprintf("%d success", success)))
	}
	fmt.Printf("  Disk usage:      %s\n", diskUsage)
	fmt.Printf("  Total files:     %s\n", files)

	if findings == nil {
		return
	}

	// Print findings section
	fmt.Printf("\n%s\n", t.Color(Bold, "Findings"))

	// SBOM path(s) and size
	if findings.SBOMPath != "" {
		sizeStr := ""
		if findings.SBOMSizeTotal > 0 {
			sizeStr = " " + t.Color(Dim, fmt.Sprintf("(%s)", t.formatBytes(findings.SBOMSizeTotal)))
		}
		fmt.Printf("  SBOM:            %s%s\n", t.Color(Cyan, findings.SBOMPath), sizeStr)
	} else if len(findings.SBOMPaths) > 0 {
		sizeStr := ""
		if findings.SBOMSizeTotal > 0 {
			sizeStr = " " + t.Color(Dim, fmt.Sprintf("(%s total)", t.formatBytes(findings.SBOMSizeTotal)))
		}
		if len(findings.SBOMPaths) == 1 {
			fmt.Printf("  SBOM:            %s%s\n", t.Color(Cyan, findings.SBOMPaths[0]), sizeStr)
		} else {
			fmt.Printf("  SBOMs:           %s%s\n", t.Color(Cyan, fmt.Sprintf("%d files generated", len(findings.SBOMPaths))), sizeStr)
			for _, p := range findings.SBOMPaths {
				fmt.Printf("                   %s\n", t.Color(Dim, p))
			}
		}
	}

	// Packages by ecosystem
	if findings.TotalPackages > 0 {
		ecoStr := t.formatEcosystemCounts(findings.PackagesByEco)
		fmt.Printf("  Packages:        %s %s\n",
			t.Color(Cyan, fmt.Sprintf("%d total", findings.TotalPackages)),
			t.Color(Dim, ecoStr))
	}

	// Vulnerabilities (only show if package-analysis scanner was run)
	if findings.ScannersRun["package-analysis"] {
		totalVulns := findings.VulnCritical + findings.VulnHigh + findings.VulnMedium + findings.VulnLow
		if totalVulns > 0 {
			vulnStr := ""
			if findings.VulnCritical > 0 {
				vulnStr += t.Color(BoldRed, fmt.Sprintf("%d critical", findings.VulnCritical)) + ", "
			}
			if findings.VulnHigh > 0 {
				vulnStr += t.Color(Red, fmt.Sprintf("%d high", findings.VulnHigh)) + ", "
			}
			if findings.VulnMedium > 0 {
				vulnStr += t.Color(Yellow, fmt.Sprintf("%d medium", findings.VulnMedium)) + ", "
			}
			if findings.VulnLow > 0 {
				vulnStr += fmt.Sprintf("%d low", findings.VulnLow)
			}
			vulnStr = strings.TrimSuffix(vulnStr, ", ")
			fmt.Printf("  Vulnerabilities: %s\n", vulnStr)
		} else {
			fmt.Printf("  Vulnerabilities: %s\n", t.Color(Green, "none found"))
		}
	}

	// Secrets (only show if code-security scanner was run)
	if findings.ScannersRun["code-security"] && findings.SecretsTotal > 0 {
		secretStr := ""
		if findings.SecretsCritical > 0 {
			secretStr += t.Color(BoldRed, fmt.Sprintf("%d critical", findings.SecretsCritical)) + ", "
		}
		if findings.SecretsHigh > 0 {
			secretStr += t.Color(Red, fmt.Sprintf("%d high", findings.SecretsHigh)) + ", "
		}
		if findings.SecretsMedium > 0 {
			secretStr += t.Color(Yellow, fmt.Sprintf("%d medium", findings.SecretsMedium))
		}
		secretStr = strings.TrimSuffix(secretStr, ", ")
		fmt.Printf("  Secrets:         %s\n", secretStr)
	}

	// Licenses (only show if package-analysis scanner was run)
	if findings.ScannersRun["package-analysis"] && findings.LicenseTypes > 0 {
		licStr := t.formatLicenseCounts(findings.LicenseCounts)
		fmt.Printf("  Licenses:        %s %s\n",
			t.Color(Cyan, fmt.Sprintf("%d types", findings.LicenseTypes)),
			t.Color(Dim, licStr))
	}

	// Malcontent (only show if package-analysis scanner was run)
	if findings.ScannersRun["package-analysis"] && (findings.MalcontentCrit > 0 || findings.MalcontentHigh > 0) {
		malStr := ""
		if findings.MalcontentCrit > 0 {
			malStr += t.Color(BoldRed, fmt.Sprintf("%d critical", findings.MalcontentCrit)) + ", "
		}
		if findings.MalcontentHigh > 0 {
			malStr += t.Color(Red, fmt.Sprintf("%d high", findings.MalcontentHigh))
		}
		malStr = strings.TrimSuffix(malStr, ", ")
		fmt.Printf("  Malcontent:      %s\n", malStr)
	}

	// Health (only show if code-quality scanner was run)
	if findings.ScannersRun["code-quality"] && (findings.HealthCritical > 0 || findings.HealthWarnings > 0) {
		healthStr := ""
		if findings.HealthCritical > 0 {
			healthStr += t.Color(Red, fmt.Sprintf("%d critical", findings.HealthCritical)) + ", "
		}
		if findings.HealthWarnings > 0 {
			healthStr += t.Color(Yellow, fmt.Sprintf("%d warnings", findings.HealthWarnings))
		}
		healthStr = strings.TrimSuffix(healthStr, ", ")
		fmt.Printf("  Package health:  %s\n", healthStr)
	}

	// Tech-ID (only show if tech-id scanner was run)
	if findings.ScannersRun["tech-id"] && findings.TechTotalTechs > 0 {
		// Show top technologies
		if len(findings.TechTopList) > 0 {
			techStr := strings.Join(findings.TechTopList, ", ")
			fmt.Printf("  Technologies:    %s %s\n",
				t.Color(Cyan, fmt.Sprintf("%d detected", findings.TechTotalTechs)),
				t.Color(Dim, "("+techStr+")"))
		} else {
			fmt.Printf("  Technologies:    %s\n", t.Color(Cyan, fmt.Sprintf("%d detected", findings.TechTotalTechs)))
		}

		// Show category breakdown if available
		if len(findings.TechByCategory) > 0 {
			catStr := t.formatCategoryCounts(findings.TechByCategory)
			fmt.Printf("                   %s\n", t.Color(Dim, catStr))
		}

		// Show ML/AI info if detected
		if findings.TechMLModels > 0 || findings.TechMLFrameworks > 0 {
			mlStr := ""
			if findings.TechMLModels > 0 {
				mlStr += fmt.Sprintf("%d models", findings.TechMLModels)
			}
			if findings.TechMLFrameworks > 0 {
				if mlStr != "" {
					mlStr += ", "
				}
				mlStr += fmt.Sprintf("%d frameworks", findings.TechMLFrameworks)
			}
			fmt.Printf("  AI/ML:           %s\n", t.Color(Cyan, mlStr))
		}

		// Show security findings if any
		if findings.TechSecurityCount > 0 {
			fmt.Printf("  AI Security:     %s\n", t.Color(Yellow, fmt.Sprintf("%d findings", findings.TechSecurityCount)))
		}
	}
}

func (t *Terminal) formatEcosystemCounts(counts map[string]int) string {
	if len(counts) == 0 {
		return ""
	}
	parts := make([]string, 0, len(counts))
	for eco, count := range counts {
		parts = append(parts, fmt.Sprintf("%s: %d", eco, count))
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

func (t *Terminal) formatLicenseCounts(counts map[string]int) string {
	if len(counts) == 0 {
		return ""
	}
	parts := make([]string, 0, len(counts))
	for lic, count := range counts {
		parts = append(parts, fmt.Sprintf("%s: %d", lic, count))
	}
	// Limit to top 5
	if len(parts) > 5 {
		parts = parts[:5]
		return "(" + strings.Join(parts, ", ") + ", ...)"
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

func (t *Terminal) formatCategoryCounts(counts map[string]int) string {
	if len(counts) == 0 {
		return ""
	}
	parts := make([]string, 0, len(counts))
	for cat, count := range counts {
		parts = append(parts, fmt.Sprintf("%s: %d", cat, count))
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

// formatBytes formats a byte size in human-readable form
func (t *Terminal) formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1fGB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1fMB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1fKB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}

// Confirm asks a yes/no question and returns the answer
func (t *Terminal) Confirm(prompt string, defaultYes bool) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	suffix := "[y/N]"
	if defaultYes {
		suffix = "[Y/n]"
	}

	fmt.Printf("%s %s: ", prompt, suffix)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return defaultYes
	}

	input = strings.TrimSpace(strings.ToLower(input))

	if input == "" {
		return defaultYes
	}

	return input == "y" || input == "yes"
}

// PromptChoice asks user to select from options
func (t *Terminal) PromptChoice(prompt string, options []string, defaultOption int) int {
	t.mu.Lock()
	defer t.mu.Unlock()

	fmt.Printf("\n%s\n", prompt)
	for i, opt := range options {
		marker := " "
		if i == defaultOption {
			marker = t.Color(Cyan, ">")
		}
		fmt.Printf("  %s %d) %s\n", marker, i+1, opt)
	}
	fmt.Printf("Choice [%d]: ", defaultOption+1)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return defaultOption
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultOption
	}

	var choice int
	if _, err := fmt.Sscanf(input, "%d", &choice); err != nil {
		return defaultOption
	}

	if choice < 1 || choice > len(options) {
		return defaultOption
	}

	return choice - 1
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ScannerResultRow represents a row in the scanner results table
type ScannerResultRow struct {
	Name     string
	Status   string        // "success", "failed", "skipped"
	Summary  string
	Duration time.Duration
}

// ScannerResultsTable prints a table of scanner results
func (t *Terminal) ScannerResultsTable(rows []ScannerResultRow) {
	t.mu.Lock()
	defer t.mu.Unlock()

	fmt.Printf("\n%s\n", t.Color(Bold, "Scanner Results"))

	// Print header
	fmt.Printf("  %-25s %-10s %-35s %s\n",
		t.Color(Dim, "Scanner"),
		t.Color(Dim, "Status"),
		t.Color(Dim, "Summary"),
		t.Color(Dim, "Duration"),
	)
	fmt.Printf("  %s\n", strings.Repeat("─", 78))

	// Print rows
	for _, row := range rows {
		statusIcon := IconSuccess
		statusColor := Green

		switch row.Status {
		case "failed":
			statusIcon = IconFailed
			statusColor = Red
		case "skipped":
			statusIcon = IconSkipped
			statusColor = Dim
		}

		// Truncate summary if too long
		summary := row.Summary
		if len(summary) > 35 {
			summary = summary[:32] + "..."
		}

		durationStr := ""
		if row.Duration > 0 {
			durationStr = fmt.Sprintf("%ds", int(row.Duration.Seconds()))
		}

		fmt.Printf("  %-25s %s %-8s %-35s %s\n",
			row.Name,
			t.Color(statusColor, statusIcon),
			t.Color(statusColor, row.Status),
			summary,
			t.Color(Dim, durationStr),
		)
	}
}
