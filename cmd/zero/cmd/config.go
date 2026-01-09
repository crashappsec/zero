package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/crashappsec/zero/pkg/core/credentials"
	"github.com/crashappsec/zero/pkg/core/terminal"
	"github.com/spf13/cobra"
	xterm "golang.org/x/term"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Zero configuration and credentials",
	Long: `Manage Zero configuration and API credentials.

Credentials are stored in ~/.zero/credentials.json with restricted permissions.

Examples:
  zero config                     Show current configuration
  zero config set github_token    Set GitHub token (prompts for value)
  zero config set anthropic_key   Set Anthropic API key (prompts for value)
  zero config get github_token    Show GitHub token source
  zero config clear               Remove all stored credentials`,
	RunE: runConfigShow,
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> [value]",
	Short: "Set a configuration value",
	Long: `Set a configuration value.

Available keys:
  github_token    GitHub personal access token
  anthropic_key   Anthropic API key

Examples:
  zero config set github_token              # Prompts for token securely
  zero config set github_token <token>      # Set directly (use with caution)
  zero config set anthropic_key             # Prompts for key securely`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runConfigSet,
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Long: `Get a configuration value and show its source.

Available keys:
  github_token    GitHub personal access token
  anthropic_key   Anthropic API key

Examples:
  zero config get github_token
  zero config get anthropic_key`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigGet,
}

var configClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all stored credentials",
	Long:  `Remove all credentials from ~/.zero/credentials.json`,
	RunE:  runConfigClear,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configClearCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	term := terminal.New()

	fmt.Println()
	fmt.Println("\033[1mZero Configuration\033[0m")
	fmt.Println(strings.Repeat("─", 60))

	// GitHub Token
	ghInfo := credentials.GetGitHubToken()
	fmt.Print("  GitHub Token:    ")
	if ghInfo.Valid {
		fmt.Printf("\033[0;32m✓\033[0m %s \033[2m(%s)\033[0m\n",
			credentials.MaskToken(ghInfo.Value), ghInfo.Source)
	} else {
		fmt.Printf("\033[0;31m✗\033[0m \033[2mnot configured\033[0m\n")
	}

	// Anthropic Key
	akInfo := credentials.GetAnthropicKey()
	fmt.Print("  Anthropic Key:   ")
	if akInfo.Valid {
		fmt.Printf("\033[0;32m✓\033[0m %s \033[2m(%s)\033[0m\n",
			credentials.MaskToken(akInfo.Value), akInfo.Source)
	} else {
		fmt.Printf("\033[0;31m✗\033[0m \033[2mnot configured\033[0m\n")
	}

	fmt.Println()
	fmt.Println("\033[2mCredentials file: ~/.zero/credentials.json\033[0m")
	fmt.Println("\033[2mPriority: environment variable > config file > gh CLI\033[0m")
	fmt.Println()

	if !ghInfo.Valid || !akInfo.Valid {
		term.Info("Run %s to configure missing credentials",
			term.Color(terminal.Cyan, "zero config set <key>"))
	}

	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := strings.ToLower(args[0])
	term := terminal.New()

	var prompt string
	var setter func(string) error

	switch key {
	case "github_token", "github", "gh":
		prompt = "GitHub Token"
		setter = credentials.SetGitHubToken
	case "anthropic_key", "anthropic", "ak":
		prompt = "Anthropic API Key"
		setter = credentials.SetAnthropicKey
	default:
		return fmt.Errorf("unknown key: %s (use 'github_token' or 'anthropic_key')", key)
	}

	var value string
	var err error

	// Check if value was provided as argument
	if len(args) > 1 {
		value = args[1]
		term.Warn("Token provided on command line - consider using interactive mode for security")
	} else {
		// Read value securely (hidden input)
		fmt.Printf("Enter %s: ", prompt)

		value, err = readSecureInput()
		if err != nil {
			return fmt.Errorf("reading input: %w", err)
		}
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("value cannot be empty")
	}

	if err := setter(value); err != nil {
		return fmt.Errorf("saving credential: %w", err)
	}

	term.Success("%s saved to ~/.zero/credentials.json", prompt)
	return nil
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	key := strings.ToLower(args[0])

	var info credentials.CredentialInfo
	var name string

	switch key {
	case "github_token", "github", "gh":
		info = credentials.GetGitHubToken()
		name = "GitHub Token"
	case "anthropic_key", "anthropic", "ak":
		info = credentials.GetAnthropicKey()
		name = "Anthropic API Key"
	default:
		return fmt.Errorf("unknown key: %s (use 'github_token' or 'anthropic_key')", key)
	}

	fmt.Printf("%s: ", name)
	if info.Valid {
		fmt.Printf("\033[0;32m%s\033[0m \033[2m(from %s)\033[0m\n",
			credentials.MaskToken(info.Value), info.Source)
	} else {
		fmt.Printf("\033[0;31mnot configured\033[0m\n")
	}

	return nil
}

func runConfigClear(cmd *cobra.Command, args []string) error {
	term := terminal.New()

	// Confirm
	fmt.Print("Clear all stored credentials? [y/N]: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input != "y" && input != "yes" {
		fmt.Println("Cancelled")
		return nil
	}

	if err := credentials.ClearCredentials(); err != nil {
		return err
	}

	term.Success("Credentials cleared")
	return nil
}

// readSecureInput reads input without echoing (for passwords/tokens)
func readSecureInput() (string, error) {
	// Try to read securely (hidden)
	if xterm.IsTerminal(syscall.Stdin) {
		bytes, err := xterm.ReadPassword(syscall.Stdin)
		fmt.Println() // Add newline after hidden input
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	}

	// Fallback to regular input if not a terminal
	reader := bufio.NewReader(os.Stdin)
	return reader.ReadString('\n')
}
