package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"

	"github.com/crashappsec/zero/pkg/config"
	"github.com/crashappsec/zero/pkg/github"
	"github.com/crashappsec/zero/pkg/scanner"
	"github.com/crashappsec/zero/pkg/terminal"
	"github.com/spf13/cobra"
)

var checkupProfile string
var checkupFix bool

var checkupCmd = &cobra.Command{
	Use:   "checkup",
	Short: "Check your setup and install missing tools",
	Long: `Analyze your GitHub token permissions and installed tools to determine
which scanners will work fully, partially, or not at all.

This helps you understand:
- What GitHub API access your token provides
- Which scanners need specific permissions
- What external tools are missing
- Offers to install missing tools automatically

Examples:
  zero checkup                    Check all registered scanners
  zero checkup --fix              Offer to install missing tools
  zero checkup --profile security Check scanners in security profile`,
	RunE: runCheckup,
}

func init() {
	rootCmd.AddCommand(checkupCmd)
	checkupCmd.Flags().StringVar(&checkupProfile, "profile", "", "Check specific profile scanners")
	checkupCmd.Flags().BoolVar(&checkupFix, "fix", false, "Offer to install missing tools")
}

// ToolInstaller contains install commands for each tool
type ToolInstaller struct {
	Name        string
	Description string
	CheckCmd    string   // Command to check if installed
	InstallCmds []string // Install commands by preference (brew, go, npm, pip)
}

var toolInstallers = map[string]ToolInstaller{
	"cdxgen": {
		Name:        "cdxgen",
		Description: "SBOM generation (CycloneDX)",
		CheckCmd:    "cdxgen --version",
		InstallCmds: []string{
			"npm install -g @cyclonedx/cdxgen",
		},
	},
	"syft": {
		Name:        "syft",
		Description: "SBOM generation (fallback)",
		CheckCmd:    "syft version",
		InstallCmds: []string{
			"brew install syft",
			"curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b /usr/local/bin",
		},
	},
	"grype": {
		Name:        "grype",
		Description: "Vulnerability scanning",
		CheckCmd:    "grype version",
		InstallCmds: []string{
			"brew install grype",
			"curl -sSfL https://raw.githubusercontent.com/anchore/grype/main/install.sh | sh -s -- -b /usr/local/bin",
		},
	},
	"osv-scanner": {
		Name:        "osv-scanner",
		Description: "Vulnerability scanning (OSV database)",
		CheckCmd:    "osv-scanner --version",
		InstallCmds: []string{
			"go install github.com/google/osv-scanner/cmd/osv-scanner@latest",
			"brew install osv-scanner",
		},
	},
	"semgrep": {
		Name:        "semgrep",
		Description: "Code security scanning (SAST)",
		CheckCmd:    "semgrep --version",
		InstallCmds: []string{
			"brew install semgrep",
			"pip install semgrep",
			"pip3 install semgrep",
		},
	},
	"gitleaks": {
		Name:        "gitleaks",
		Description: "Secrets detection",
		CheckCmd:    "gitleaks version",
		InstallCmds: []string{
			"brew install gitleaks",
			"go install github.com/gitleaks/gitleaks/v8@latest",
		},
	},
	"trivy": {
		Name:        "trivy",
		Description: "Container and IaC scanning",
		CheckCmd:    "trivy version",
		InstallCmds: []string{
			"brew install trivy",
			"curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin",
		},
	},
	"checkov": {
		Name:        "checkov",
		Description: "Infrastructure as Code scanning",
		CheckCmd:    "checkov --version",
		InstallCmds: []string{
			"pip install checkov",
			"pip3 install checkov",
			"brew install checkov",
		},
	},
	"malcontent": {
		Name:        "malcontent",
		Description: "Supply chain malware detection",
		CheckCmd:    "mal --version",
		InstallCmds: []string{
			"go install github.com/chainguard-dev/malcontent/cmd/mal@latest",
		},
	},
	"gh": {
		Name:        "gh",
		Description: "GitHub CLI (authentication)",
		CheckCmd:    "gh --version",
		InstallCmds: []string{
			"brew install gh",
			"curl -sS https://webi.sh/gh | sh",
		},
	},
}

func runCheckup(cmd *cobra.Command, args []string) error {
	term := terminal.New()

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Determine which scanners to check
	var scanners []string
	if checkupProfile != "" {
		scanners, err = cfg.GetProfileScanners(checkupProfile)
		if err != nil {
			return err
		}
	} else {
		scanners = scanner.List()
	}

	// Create GitHub client and generate checkup
	gh := github.NewClient()
	result, err := gh.GenerateRoadmap(scanners)
	if err != nil {
		return fmt.Errorf("running checkup: %w", err)
	}

	// Print results
	fmt.Println()
	printBanner()
	fmt.Println()

	printTokenStatus(result, term)
	missingTools := printToolsStatus(result, term)
	printScannerCompatibility(result, checkupProfile, term)
	printSummary(result, term)

	// Offer to install missing tools
	if checkupFix && len(missingTools) > 0 {
		fmt.Println()
		offerToolInstallation(missingTools, term)
	} else if len(missingTools) > 0 {
		fmt.Println()
		term.Info("Run %s to install missing tools", term.Color(terminal.Cyan, "zero checkup --fix"))
	}

	fmt.Println()
	return nil
}

func printTokenStatus(result *github.RoadmapResult, term *terminal.Terminal) {
	fmt.Println("\033[1mGitHub Token Status\033[0m")
	fmt.Println(strings.Repeat("─", 60))

	if !result.TokenInfo.Valid {
		fmt.Printf("  \033[0;31m✗\033[0m Status: \033[0;31mInvalid or Missing\033[0m\n")
		if result.TokenInfo.Error != "" {
			fmt.Printf("    Error: %s\n", result.TokenInfo.Error)
		}
		fmt.Println()
		fmt.Println("  \033[2mTo authenticate:\033[0m")
		fmt.Println("    • Run: gh auth login")
		fmt.Println("    • Or set: export GITHUB_TOKEN=ghp_...")
	} else {
		fmt.Printf("  \033[0;32m✓\033[0m Status: \033[0;32mValid\033[0m\n")
		fmt.Printf("    User: %s\n", result.TokenInfo.Username)
		fmt.Printf("    Type: %s\n", formatTokenType(result.TokenInfo.Type))

		if result.TokenInfo.RateLimit > 0 {
			fmt.Printf("    Rate Limit: %d/%d remaining\n",
				result.TokenInfo.RateRemaining, result.TokenInfo.RateLimit)
		}

		if len(result.TokenInfo.Scopes) > 0 {
			fmt.Printf("    Scopes: %s\n", strings.Join(result.TokenInfo.Scopes, ", "))
		}

		if len(result.TokenInfo.Permissions) > 0 {
			perms := make([]string, 0)
			for k, v := range result.TokenInfo.Permissions {
				perms = append(perms, fmt.Sprintf("%s:%s", k, v))
			}
			sort.Strings(perms)
			fmt.Printf("    Permissions: %s\n", strings.Join(perms, ", "))
		}
	}
	fmt.Println()
}

func printToolsStatus(result *github.RoadmapResult, term *terminal.Terminal) []string {
	var missingTools []string

	if len(result.ToolsStatus) > 0 {
		fmt.Println("\033[1mExternal Tools\033[0m")
		fmt.Println(strings.Repeat("─", 60))

		// Sort tools
		tools := make([]string, 0, len(result.ToolsStatus))
		for tool := range result.ToolsStatus {
			tools = append(tools, tool)
		}
		sort.Strings(tools)

		for _, tool := range tools {
			available := result.ToolsStatus[tool]
			if available {
				fmt.Printf("  \033[0;32m✓\033[0m %s\n", tool)
			} else {
				installer, ok := toolInstallers[tool]
				desc := ""
				if ok {
					desc = fmt.Sprintf(" - %s", installer.Description)
				}
				fmt.Printf("  \033[0;31m✗\033[0m %s \033[2m(not installed%s)\033[0m\n", tool, desc)
				missingTools = append(missingTools, tool)
			}
		}
		fmt.Println()
	}

	return missingTools
}

func printScannerCompatibility(result *github.RoadmapResult, profile string, term *terminal.Terminal) {
	title := "Scanner Compatibility"
	if profile != "" {
		title = fmt.Sprintf("Scanner Compatibility (%s profile)", profile)
	}
	fmt.Printf("\033[1m%s\033[0m\n", title)
	fmt.Println(strings.Repeat("─", 60))

	// Group scanners by status
	ready := make([]github.ScannerCompatibility, 0)
	limited := make([]github.ScannerCompatibility, 0)
	unavailable := make([]github.ScannerCompatibility, 0)

	for _, s := range result.Scanners {
		switch s.Status {
		case "ready":
			ready = append(ready, s)
		case "limited":
			limited = append(limited, s)
		case "unavailable":
			unavailable = append(unavailable, s)
		}
	}

	if len(ready) > 0 {
		fmt.Println("\n  \033[0;32mReady\033[0m")
		for _, s := range ready {
			fmt.Printf("    \033[0;32m✓\033[0m %s", s.Scanner)
			if s.Reason != "" {
				fmt.Printf(" \033[2m(%s)\033[0m", s.Reason)
			}
			fmt.Println()
		}
	}

	if len(limited) > 0 {
		fmt.Println("\n  \033[0;33mLimited\033[0m")
		for _, s := range limited {
			fmt.Printf("    \033[0;33m⚠\033[0m %s", s.Scanner)
			if s.Reason != "" {
				fmt.Printf("\n      \033[2m%s\033[0m", s.Reason)
			}
			fmt.Println()
		}
	}

	if len(unavailable) > 0 {
		fmt.Println("\n  \033[0;31mUnavailable\033[0m")
		for _, s := range unavailable {
			fmt.Printf("    \033[0;31m✗\033[0m %s", s.Scanner)
			if s.Reason != "" {
				fmt.Printf("\n      \033[2m%s\033[0m", s.Reason)
			}
			fmt.Println()
		}
	}
}

func printSummary(result *github.RoadmapResult, term *terminal.Terminal) {
	fmt.Println()
	fmt.Println("\033[1mSummary\033[0m")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("  Ready:       \033[0;32m%d\033[0m\n", result.Summary.Ready)
	fmt.Printf("  Limited:     \033[0;33m%d\033[0m\n", result.Summary.Limited)
	fmt.Printf("  Unavailable: \033[0;31m%d\033[0m\n", result.Summary.Unavailable)
	fmt.Printf("  Total:       %d\n", result.Summary.Total)
}

func offerToolInstallation(missingTools []string, term *terminal.Terminal) {
	fmt.Println("\033[1mInstall Missing Tools\033[0m")
	fmt.Println(strings.Repeat("─", 60))

	reader := bufio.NewReader(os.Stdin)

	for _, tool := range missingTools {
		installer, ok := toolInstallers[tool]
		if !ok {
			continue
		}

		// Find best install command for this platform
		installCmd := findBestInstallCmd(installer.InstallCmds)
		if installCmd == "" {
			fmt.Printf("  \033[2mNo installer available for %s on this platform\033[0m\n", tool)
			continue
		}

		fmt.Printf("\n  Install %s? (%s)\n", term.Color(terminal.Cyan, tool), installer.Description)
		fmt.Printf("    Command: %s\n", term.Color(terminal.Dim, installCmd))
		fmt.Printf("    [y/N]: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "y" || input == "yes" {
			fmt.Printf("    Installing %s...\n", tool)
			if err := runInstallCmd(installCmd); err != nil {
				term.Error("    Failed: %v", err)
			} else {
				term.Success("    Installed %s", tool)
			}
		} else {
			fmt.Printf("    Skipped\n")
		}
	}
}

func findBestInstallCmd(cmds []string) string {
	// Check if brew is available (macOS)
	if runtime.GOOS == "darwin" {
		if _, err := exec.LookPath("brew"); err == nil {
			for _, cmd := range cmds {
				if strings.HasPrefix(cmd, "brew ") {
					return cmd
				}
			}
		}
	}

	// Check if go is available
	if _, err := exec.LookPath("go"); err == nil {
		for _, cmd := range cmds {
			if strings.HasPrefix(cmd, "go install") {
				return cmd
			}
		}
	}

	// Check if npm is available
	if _, err := exec.LookPath("npm"); err == nil {
		for _, cmd := range cmds {
			if strings.HasPrefix(cmd, "npm ") {
				return cmd
			}
		}
	}

	// Check if pip is available
	if _, err := exec.LookPath("pip3"); err == nil {
		for _, cmd := range cmds {
			if strings.HasPrefix(cmd, "pip3 ") || strings.HasPrefix(cmd, "pip ") {
				return strings.Replace(cmd, "pip ", "pip3 ", 1)
			}
		}
	}
	if _, err := exec.LookPath("pip"); err == nil {
		for _, cmd := range cmds {
			if strings.HasPrefix(cmd, "pip ") {
				return cmd
			}
		}
	}

	// Fall back to curl if available
	if _, err := exec.LookPath("curl"); err == nil {
		for _, cmd := range cmds {
			if strings.HasPrefix(cmd, "curl ") {
				return cmd
			}
		}
	}

	// Return first command as fallback
	if len(cmds) > 0 {
		return cmds[0]
	}
	return ""
}

func runInstallCmd(cmdStr string) error {
	// Split command into parts
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	// Handle piped commands
	if strings.Contains(cmdStr, "|") {
		cmd := exec.Command("sh", "-c", cmdStr)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func formatTokenType(t string) string {
	switch t {
	case "classic":
		return "Classic PAT"
	case "fine-grained":
		return "Fine-grained PAT"
	case "github-app":
		return "GitHub App"
	case "oauth":
		return "OAuth Token"
	default:
		return t
	}
}
