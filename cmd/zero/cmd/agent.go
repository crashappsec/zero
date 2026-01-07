package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/crashappsec/zero/pkg/agent"
	"github.com/spf13/cobra"
)

// Terminal color constants
const (
	colorReset   = "\033[0m"
	colorDim     = "\033[2m"
	colorRed     = "\033[0;31m"
	colorGreen   = "\033[0;32m"
	colorYellow  = "\033[1;33m"
	colorBlue    = "\033[0;34m"
	colorCyan    = "\033[0;36m"
	colorWhite   = "\033[0;37m"
	colorBold    = "\033[1m"
	colorMagenta = "\033[0;35m"
)

// colorize wraps text in ANSI color codes
func colorize(color, text string) string {
	if os.Getenv("NO_COLOR") != "" {
		return text
	}
	return color + text + colorReset
}

// printDivider prints a horizontal line
func printDivider() {
	fmt.Println(strings.Repeat("━", 78))
}

var (
	agentIDFlag   string
	voiceModeFlag string
	queryFlag     string
	noStreamFlag  bool
	jsonOutputFlag bool
)

// agentCmd represents the agent command
var agentCmd = &cobra.Command{
	Use:   "agent [project]",
	Short: "Start interactive agent chat",
	Long: `Start an interactive chat session with Zero agents.

Zero agents are AI-powered specialists that can analyze your code,
investigate security issues, and provide expert guidance.

Examples:
  zero agent                           # Chat with Zero (orchestrator)
  zero agent expressjs/express         # Chat about a specific project
  zero agent --agent cereal            # Chat with Cereal (supply chain)
  zero agent -q "list vulnerabilities" # One-shot query mode
  zero agent --voice minimal           # Use minimal voice mode

Available Agents:
  zero    - Master orchestrator (default)
  cereal  - Supply chain security
  razor   - Code security, SAST
  blade   - Compliance, auditing
  phreak  - Legal, licenses
  acid    - Frontend, React
  dade    - Backend, APIs
  nikon   - Architecture
  joey    - Build, CI/CD
  plague  - DevOps, infrastructure
  gibson  - DORA metrics
  gill    - Cryptography
  hal     - AI/ML security`,
	RunE: runAgentCmd,
}

func init() {
	rootCmd.AddCommand(agentCmd)

	agentCmd.Flags().StringVarP(&agentIDFlag, "agent", "a", "zero", "Agent to chat with")
	agentCmd.Flags().StringVar(&voiceModeFlag, "voice", "full", "Voice mode: full, minimal, neutral")
	agentCmd.Flags().StringVarP(&queryFlag, "query", "q", "", "One-shot query (non-interactive)")
	agentCmd.Flags().BoolVar(&noStreamFlag, "no-stream", false, "Disable streaming output")
	agentCmd.Flags().BoolVar(&jsonOutputFlag, "json", false, "Output in JSON format")
}

func runAgentCmd(cmd *cobra.Command, args []string) error {
	// Get project from args
	var projectID string
	if len(args) > 0 {
		projectID = args[0]
	}

	// Create runtime
	runtime, err := agent.NewRuntime(&agent.RuntimeOptions{})
	if err != nil {
		return fmt.Errorf("failed to create agent runtime: %w", err)
	}

	// Check if API key is configured
	if !runtime.IsAvailable() {
		fmt.Printf("  %s %s\n", colorize(colorRed, "✗"), "ANTHROPIC_API_KEY not configured")
		fmt.Println()
		fmt.Println("Set your API key:")
		fmt.Println("  export ANTHROPIC_API_KEY=sk-ant-...")
		fmt.Println()
		return fmt.Errorf("API key required")
	}

	// One-shot mode
	if queryFlag != "" {
		return runOneShotQuery(runtime, projectID)
	}

	// Interactive mode
	return runInteractiveMode(runtime, projectID)
}

func runOneShotQuery(runtime *agent.Runtime, projectID string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	req := &agent.ChatRequest{
		AgentID:   agentIDFlag,
		ProjectID: projectID,
		VoiceMode: voiceModeFlag,
		Message:   queryFlag,
	}

	var response strings.Builder
	var toolCalls []map[string]interface{}

	err := runtime.Chat(ctx, req, func(event agent.ChatEvent) {
		switch event.Type {
		case "text":
			if !jsonOutputFlag {
				fmt.Print(event.Text)
			}
			response.WriteString(event.Text)
		case "tool_call":
			if jsonOutputFlag {
				toolCalls = append(toolCalls, map[string]interface{}{
					"name":  event.ToolCall.Name,
					"input": json.RawMessage(event.ToolCall.Input),
				})
			} else if !noStreamFlag {
				fmt.Printf("\n%s %s\n", colorize(colorCyan, "▸"), event.ToolCall.Name)
			}
		case "error":
			if !jsonOutputFlag {
				fmt.Fprintf(os.Stderr, "\n%s\n", colorize(colorRed, event.Error))
			}
		}
	})

	if !jsonOutputFlag {
		fmt.Println()
	}

	if jsonOutputFlag {
		output := map[string]interface{}{
			"response":   response.String(),
			"tool_calls": toolCalls,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(output)
	}

	return err
}

func runInteractiveMode(runtime *agent.Runtime, projectID string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\n\nExiting...")
		cancel()
		os.Exit(0)
	}()

	// Print header
	printAgentHeader(runtime, projectID)

	// Print greeting
	greeting, err := runtime.GetGreeting(agentIDFlag, projectID)
	if err == nil && greeting != "" {
		printAgentMessage(agentIDFlag, greeting)
	}

	// Create session
	var sessionID string

	// Start REPL
	reader := bufio.NewReader(os.Stdin)

	for {
		// Print prompt
		fmt.Print(colorize(colorGreen, "\n❯ "))

		// Read input
		input, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Handle special commands
		if strings.HasPrefix(input, "/") {
			if handleAgentCommand(runtime, input, &agentIDFlag, &projectID) {
				continue
			}
		}

		// Check for exit
		if input == "exit" || input == "quit" || input == "q" {
			fmt.Println("Hack the planet.")
			break
		}

		// Send message
		req := &agent.ChatRequest{
			SessionID: sessionID,
			AgentID:   agentIDFlag,
			ProjectID: projectID,
			VoiceMode: voiceModeFlag,
			Message:   input,
		}

		fmt.Println()

		err = runtime.Chat(ctx, req, func(event agent.ChatEvent) {
			handleChatEvent(event, agentIDFlag)
		})

		if err != nil {
			fmt.Printf("  %s Chat error: %v\n", colorize(colorRed, "✗"), err)
		}

		fmt.Println()
	}

	return nil
}

func handleChatEvent(event agent.ChatEvent, currentAgent string) {
	switch event.Type {
	case "text":
		fmt.Print(event.Text)

	case "tool_call":
		// Show tool usage
		name := event.ToolCall.Name
		var inputMap map[string]interface{}
		json.Unmarshal(event.ToolCall.Input, &inputMap)

		// Format tool call display
		fmt.Printf("\n%s %s", colorize(colorCyan, "▸"), colorize(colorCyan, name))

		// Show key parameters
		switch name {
		case "Read":
			if path, ok := inputMap["file_path"].(string); ok {
				fmt.Printf(" %s", colorize(colorDim, path))
			}
		case "Grep":
			if pattern, ok := inputMap["pattern"].(string); ok {
				fmt.Printf(" %s", colorize(colorDim, pattern))
			}
		case "GetAnalysis":
			if scanner, ok := inputMap["scanner"].(string); ok {
				fmt.Printf(" %s", colorize(colorDim, scanner))
			}
		case "DelegateAgent":
			if agentID, ok := inputMap["agent_id"].(string); ok {
				fmt.Printf(" → %s", colorize(colorYellow, agentID))
			}
		}
		fmt.Println()

	case "tool_result":
		// Show brief result indicator
		if event.ToolResult.IsError {
			fmt.Printf("  %s\n", colorize(colorRed, "✗ Error"))
		} else {
			// Show truncated result
			content := event.ToolResult.Content
			lines := strings.Split(content, "\n")
			if len(lines) > 3 {
				fmt.Printf("  %s\n", colorize(colorDim, fmt.Sprintf("(%d lines)", len(lines))))
			}
		}

	case "error":
		fmt.Printf("\n%s %s\n", colorize(colorRed, "Error:"), event.Error)

	case "done":
		// Optionally show token usage
	}
}

func handleAgentCommand(runtime *agent.Runtime, input string, agentID, projectID *string) bool {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return true
	}

	cmd := parts[0]

	switch cmd {
	case "/help", "/h", "/?":
		printAgentHelp()
		return true

	case "/agents", "/list":
		printAgentList(runtime)
		return true

	case "/switch", "/s":
		if len(parts) < 2 {
			fmt.Println("Usage: /switch <agent_id>")
			return true
		}
		newAgent := parts[1]
		if _, _, _, ok := runtime.GetAgentInfo(newAgent); !ok {
			fmt.Printf("Unknown agent: %s\n", newAgent)
			return true
		}
		*agentID = newAgent
		greeting, _ := runtime.GetGreeting(newAgent, *projectID)
		printAgentMessage(newAgent, greeting)
		return true

	case "/project", "/p":
		if len(parts) < 2 {
			if *projectID != "" {
				fmt.Printf("Current project: %s\n", *projectID)
			} else {
				fmt.Println("No project set. Usage: /project <owner/repo>")
			}
			return true
		}
		*projectID = parts[1]
		fmt.Printf("Project set to: %s\n", *projectID)
		return true

	case "/voice", "/v":
		if len(parts) < 2 {
			fmt.Printf("Current voice mode: %s\n", voiceModeFlag)
			return true
		}
		newMode := parts[1]
		if newMode != "full" && newMode != "minimal" && newMode != "neutral" {
			fmt.Println("Voice modes: full, minimal, neutral")
			return true
		}
		voiceModeFlag = newMode
		fmt.Printf("Voice mode set to: %s\n", voiceModeFlag)
		return true

	case "/clear", "/c":
		fmt.Print("\033[H\033[2J") // Clear screen
		printAgentHeader(runtime, *projectID)
		return true

	case "/exit", "/quit", "/q":
		fmt.Println("Hack the planet.")
		os.Exit(0)
		return true

	default:
		fmt.Printf("Unknown command: %s (try /help)\n", cmd)
		return true
	}
}

func printAgentHeader(runtime *agent.Runtime, projectID string) {
	printDivider()

	name, persona, _, _ := runtime.GetAgentInfo(agentIDFlag)
	fmt.Printf("%s %s (%s)\n",
		colorize(colorGreen, "▸"),
		colorize(colorWhite+colorBold, name),
		colorize(colorDim, persona))

	if projectID != "" {
		fmt.Printf("  Project: %s\n", colorize(colorCyan, projectID))
	}

	fmt.Printf("  Voice: %s\n", colorize(colorDim, voiceModeFlag))
	fmt.Println()
	fmt.Printf("  Type %s for commands, %s to exit\n",
		colorize(colorDim, "/help"),
		colorize(colorDim, "/exit"))

	printDivider()
	fmt.Println()
}

func printAgentMessage(agentID, message string) {
	// Get agent color based on ID
	color := getAgentColorCode(agentID)

	name, _, _, _ := (&agent.AgentLoader{}).GetAgentInfo(agentID)
	if name == "" {
		name = agentID
	}

	fmt.Printf("%s: %s\n", colorize(color, name), message)
}

func getAgentColorCode(agentID string) string {
	colors := map[string]string{
		"zero":   colorGreen,
		"cereal": colorYellow,
		"razor":  colorRed,
		"blade":  colorBlue,
		"phreak": colorMagenta,
		"acid":   colorCyan,
		"dade":   colorWhite,
		"nikon":  colorBlue,
		"joey":   colorYellow,
		"plague": colorRed,
		"gibson": colorGreen,
		"gill":   colorCyan,
		"hal": colorMagenta,
	}
	if c, ok := colors[agentID]; ok {
		return c
	}
	return colorWhite
}

func printAgentHelp() {
	fmt.Println()
	fmt.Println(colorize(colorWhite+colorBold, "Commands:"))
	fmt.Println("  /help, /h      Show this help")
	fmt.Println("  /agents        List available agents")
	fmt.Println("  /switch <id>   Switch to different agent")
	fmt.Println("  /project <id>  Set project context (owner/repo)")
	fmt.Println("  /voice <mode>  Set voice mode (full/minimal/neutral)")
	fmt.Println("  /clear         Clear screen")
	fmt.Println("  /exit, /quit   Exit the agent")
	fmt.Println()
	fmt.Println(colorize(colorWhite+colorBold, "Tips:"))
	fmt.Println("  - Ask questions naturally")
	fmt.Println("  - Agents can use tools to investigate")
	fmt.Println("  - Use /switch to change specialists")
	fmt.Println()
}

func printAgentList(runtime *agent.Runtime) {
	fmt.Println()
	fmt.Println(colorize(colorWhite+colorBold, "Available Agents:"))
	fmt.Println()

	agents := []struct {
		ID   string
		Desc string
	}{
		{"zero", "Master orchestrator"},
		{"cereal", "Supply chain security"},
		{"razor", "Code security, SAST, secrets"},
		{"blade", "Compliance, SOC 2, ISO 27001"},
		{"phreak", "Legal, licenses, privacy"},
		{"acid", "Frontend, React, TypeScript"},
		{"dade", "Backend, APIs, databases"},
		{"nikon", "Architecture, system design"},
		{"joey", "Build, CI/CD, pipelines"},
		{"plague", "DevOps, infrastructure, K8s"},
		{"gibson", "DORA metrics, team health"},
		{"gill", "Cryptography specialist"},
		{"hal", "AI/ML security"},
	}

	for _, a := range agents {
		color := getAgentColorCode(a.ID)
		marker := " "
		if a.ID == agentIDFlag {
			marker = "▸"
		}
		fmt.Printf("  %s %s  %s\n",
			marker,
			colorize(color, fmt.Sprintf("%-8s", a.ID)),
			colorize(colorDim, a.Desc))
	}
	fmt.Println()
}
