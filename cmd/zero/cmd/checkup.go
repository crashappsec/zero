package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"

	"github.com/crashappsec/zero/pkg/core/config"
	"github.com/crashappsec/zero/pkg/core/credentials"
	"github.com/crashappsec/zero/pkg/core/github"
	"github.com/crashappsec/zero/pkg/scanner"
	"github.com/crashappsec/zero/pkg/core/terminal"
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

// Prerequisites are foundational tools needed to run Zero or install other tools
var prerequisites = []ToolInstaller{
	{
		Name:        "go",
		Description: "Go runtime (required for some tool installations)",
		CheckCmd:    "go version",
		InstallCmds: []string{
			"brew install go",
			"curl -fsSL https://go.dev/dl/go1.23.4.linux-amd64.tar.gz | sudo tar -C /usr/local -xzf -",
		},
	},
	{
		Name:        "node",
		Description: "Node.js runtime (required for Evidence reports)",
		CheckCmd:    "node --version",
		InstallCmds: []string{
			"brew install node",
			"curl -fsSL https://fnm.vercel.app/install | bash && fnm install --lts",
		},
	},
	{
		Name:        "npm",
		Description: "Node package manager (required for cdxgen)",
		CheckCmd:    "npm --version",
		InstallCmds: []string{
			"brew install node", // npm comes with node
		},
	},
	{
		Name:        "python3",
		Description: "Python runtime (required for semgrep, checkov)",
		CheckCmd:    "python3 --version",
		InstallCmds: []string{
			"brew install python3",
			"sudo apt install python3",
		},
	},
}

var toolInstallers = map[string]ToolInstaller{
	"cdxgen": {
		Name:        "cdxgen",
		Description: "SBOM generation (CycloneDX)",
		CheckCmd:    "cdxgen --version",
		InstallCmds: []string{
			"npm install -g @cyclonedx/cdxgen@latest",
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
	"docker": {
		Name:        "docker",
		Description: "Container runtime (for Docker-based Zero)",
		CheckCmd:    "docker --version",
		InstallCmds: []string{
			"brew install --cask docker",
			"curl -fsSL https://get.docker.com | sh",
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

	missingPrereqs := printPrerequisites(term)
	printTokenStatus(result, gh, term)
	missingTools := printToolsStatus(result, term)
	missingTools = append(missingPrereqs, missingTools...)
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

func printPrerequisites(term *terminal.Terminal) []string {
	var missing []string

	fmt.Println("\033[1mPrerequisites\033[0m")
	fmt.Println(strings.Repeat("─", 60))

	for _, prereq := range prerequisites {
		// Check if installed
		parts := strings.Fields(prereq.CheckCmd)
		_, err := exec.LookPath(parts[0])
		if err == nil {
			// Get version
			cmd := exec.Command(parts[0], parts[1:]...)
			output, _ := cmd.Output()
			version := strings.TrimSpace(string(output))
			// Extract just the version number if possible
			if idx := strings.Index(version, "\n"); idx > 0 {
				version = version[:idx]
			}
			fmt.Printf("  \033[0;32m✓\033[0m %s \033[2m%s\033[0m\n", prereq.Name, version)
		} else {
			fmt.Printf("  \033[0;31m✗\033[0m %s \033[2m(not installed - %s)\033[0m\n", prereq.Name, prereq.Description)
			missing = append(missing, prereq.Name)
		}
	}
	fmt.Println()

	return missing
}

func printTokenStatus(result *github.RoadmapResult, gh *github.Client, term *terminal.Terminal) {
	fmt.Println("\033[1mCredentials\033[0m")
	fmt.Println(strings.Repeat("─", 60))

	// GitHub Token - show source
	ghInfo := credentials.GetGitHubToken()
	fmt.Print("  GitHub Token:    ")
	if ghInfo.Valid {
		fmt.Printf("\033[0;32m✓\033[0m %s \033[2m(%s)\033[0m\n",
			credentials.MaskToken(ghInfo.Value), ghInfo.Source)
		if result.TokenInfo.Valid {
			fmt.Printf("    User: %s, Type: %s\n", result.TokenInfo.Username, formatTokenType(result.TokenInfo.Type))
			if result.TokenInfo.RateLimit > 0 {
				fmt.Printf("    Rate Limit: %d/%d remaining\n",
					result.TokenInfo.RateRemaining, result.TokenInfo.RateLimit)
			}

			// Warn about classic tokens
			if result.TokenInfo.Type == "classic" {
				fmt.Println()
				fmt.Println("  \033[0;33m⚠ Security Warning: Classic PAT detected\033[0m")
				fmt.Println("    Classic tokens often have broader access than needed.")
				fmt.Println("    \033[2mRecommendation: Use a fine-grained PAT scoped to specific repos.\033[0m")
				fmt.Println("    \033[2mCreate one at: https://github.com/settings/tokens?type=beta\033[0m")
			}
		}
	} else {
		fmt.Printf("\033[0;31m✗\033[0m \033[2mnot configured\033[0m\n")
	}

	// Anthropic Key - show source
	akInfo := credentials.GetAnthropicKey()
	fmt.Print("  Anthropic Key:   ")
	if akInfo.Valid {
		fmt.Printf("\033[0;32m✓\033[0m %s \033[2m(%s)\033[0m\n",
			credentials.MaskToken(akInfo.Value), akInfo.Source)
	} else {
		fmt.Printf("\033[0;31m✗\033[0m \033[2mnot configured\033[0m\n")
	}

	if !ghInfo.Valid || !akInfo.Valid {
		fmt.Println()
		fmt.Println("  \033[2mTo configure:\033[0m")
		fmt.Println("    • Run: zero config set github_token")
		fmt.Println("    • Run: zero config set anthropic_key")
		fmt.Println("    • Or set environment variables")
	}
	fmt.Println()

	// Show accessible repos if token is valid
	if ghInfo.Valid && result.TokenInfo.Valid {
		printAccessibleRepos(gh, result.TokenInfo.Type, term)
	}
}

func printAccessibleRepos(gh *github.Client, tokenType string, term *terminal.Terminal) {
	fmt.Println("\033[1mAccessible Repositories\033[0m")
	fmt.Println(strings.Repeat("─", 60))

	if tokenType == "classic" {
		fmt.Println("  \033[0;33mThis token has access to the following repos:\033[0m")
		fmt.Println()
	}

	summary, err := gh.ListAccessibleRepos()
	if err != nil {
		fmt.Printf("  \033[2mUnable to list repos: %v\033[0m\n", err)
		fmt.Println()
		return
	}

	// Personal repos - list actual names
	if len(summary.PersonalRepos) > 0 {
		fmt.Printf("  \033[1mPersonal (%s)\033[0m\n", summary.User)
		printRepoList(summary.PersonalRepos)
	}

	// Organization repos - list actual names
	for _, org := range summary.Orgs {
		if len(org.Repos) > 0 {
			fmt.Printf("\n  \033[1m%s\033[0m", org.Login)
			if org.Description != "" {
				desc := org.Description
				if len(desc) > 40 {
					desc = desc[:37] + "..."
				}
				fmt.Printf(" \033[2m- %s\033[0m", desc)
			}
			fmt.Println()
			printRepoList(org.Repos)
		}
	}

	if len(summary.Orgs) == 0 && len(summary.PersonalRepos) == 0 {
		fmt.Println("  \033[2mNo repositories accessible\033[0m")
	}

	fmt.Printf("\n  Total: \033[0;32m%d repos\033[0m accessible to this token\n", summary.TotalRepos)
	fmt.Println()
}

func printRepoList(repos []github.AccessibleRepo) {
	for _, repo := range repos {
		visibility := "\033[2mpublic\033[0m"
		if repo.Private {
			visibility = "\033[0;33mprivate\033[0m"
		}
		// Show just repo name without owner for cleaner display
		name := repo.FullName
		if idx := strings.Index(name, "/"); idx > 0 {
			name = name[idx+1:]
		}
		fmt.Printf("    • %s (%s)\n", name, visibility)
	}
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
		// Look up in toolInstallers first, then prerequisites
		installer, ok := toolInstallers[tool]
		if !ok {
			// Check prerequisites
			for _, prereq := range prerequisites {
				if prereq.Name == tool {
					installer = prereq
					ok = true
					break
				}
			}
		}
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
