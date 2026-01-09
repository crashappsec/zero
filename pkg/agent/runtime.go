package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/crashappsec/zero/pkg/core/github"
)

// Runtime is the main agent execution engine
type Runtime struct {
	loader    *AgentLoader
	sessions  *SessionManager
	prompt    *PromptBuilder
	llm       *LLMClient
	zeroHome  string
	agentsDir string
}

// RuntimeOptions configures the runtime
type RuntimeOptions struct {
	ZeroHome   string
	AgentsDir  string
	PersistDir string // For session persistence
	LLMOptions *LLMClientOptions
}

// NewRuntime creates a new agent runtime
func NewRuntime(opts *RuntimeOptions) (*Runtime, error) {
	if opts == nil {
		opts = &RuntimeOptions{}
	}

	// Default zero home - check multiple locations
	zeroHome := opts.ZeroHome
	if zeroHome == "" {
		zeroHome = os.Getenv("ZERO_HOME")
	}
	if zeroHome == "" {
		// Auto-detect: check for local .zero directory first
		zeroHome = findZeroHome()
	}

	// Default agents directory
	agentsDir := opts.AgentsDir
	if agentsDir == "" {
		// Try to find agents directory relative to executable or working directory
		if dir, err := findAgentsDir(); err == nil {
			agentsDir = dir
		} else {
			return nil, fmt.Errorf("agents directory not found: %w", err)
		}
	}

	loader := NewAgentLoader(agentsDir)
	sessions := NewSessionManager(opts.PersistDir)
	prompt := NewPromptBuilder(loader, zeroHome)
	llm := NewLLMClient(opts.LLMOptions)

	return &Runtime{
		loader:    loader,
		sessions:  sessions,
		prompt:    prompt,
		llm:       llm,
		zeroHome:  zeroHome,
		agentsDir: agentsDir,
	}, nil
}

// findZeroHome tries to locate the .zero directory
// Priority: ./zero -> ../.zero -> ~/.zero
func findZeroHome() string {
	// Check current directory
	if info, err := os.Stat(".zero"); err == nil && info.IsDir() {
		if absPath, err := filepath.Abs(".zero"); err == nil {
			return absPath
		}
		return ".zero"
	}

	// Check parent directory
	if info, err := os.Stat("../.zero"); err == nil && info.IsDir() {
		if absPath, err := filepath.Abs("../.zero"); err == nil {
			return absPath
		}
	}

	// Check relative to executable
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		localZero := filepath.Join(exeDir, ".zero")
		if info, err := os.Stat(localZero); err == nil && info.IsDir() {
			return localZero
		}
	}

	// Fall back to home directory
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".zero")
	}

	return ".zero"
}

// findAgentsDir tries to locate the agents directory
func findAgentsDir() (string, error) {
	// Try common locations
	candidates := []string{
		"agents",                    // Current directory
		"../agents",                 // Parent directory
		"../../agents",              // Two levels up
		filepath.Join(os.Getenv("ZERO_HOME"), "agents"),
	}

	// Also try relative to executable
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "agents"),
			filepath.Join(exeDir, "..", "agents"),
		)
	}

	for _, dir := range candidates {
		if dir == "" {
			continue
		}
		absDir, err := filepath.Abs(dir)
		if err != nil {
			continue
		}
		if info, err := os.Stat(absDir); err == nil && info.IsDir() {
			// Verify it has agent definitions
			if _, err := os.Stat(filepath.Join(absDir, "orchestrator", "agent.md")); err == nil {
				return absDir, nil
			}
		}
	}

	return "", fmt.Errorf("could not find agents directory")
}

// IsAvailable checks if the runtime is ready
func (r *Runtime) IsAvailable() bool {
	return r.llm.IsAvailable()
}

// ChatRequest represents a request to chat with an agent
type ChatRequest struct {
	SessionID string `json:"session_id,omitempty"` // Optional, creates new if empty
	AgentID   string `json:"agent_id,omitempty"`   // Default: "zero"
	ProjectID string `json:"project_id,omitempty"` // Optional project context
	VoiceMode string `json:"voice_mode,omitempty"` // Default: "full"
	Message   string `json:"message"`
}

// Chat sends a message and streams the response
func (r *Runtime) Chat(ctx context.Context, req *ChatRequest, callback func(ChatEvent)) error {
	// Get or create session
	agentID := req.AgentID
	if agentID == "" {
		agentID = "zero"
	}

	var session *Session
	if req.SessionID != "" {
		session = r.sessions.GetOrCreate(req.SessionID, agentID)
	} else {
		session = r.sessions.Create(agentID)
	}

	// Update session context
	if req.ProjectID != "" {
		session.SetProject(req.ProjectID)
	}
	if req.VoiceMode != "" {
		session.SetVoiceMode(req.VoiceMode)
	}

	// Add user message
	session.AddUserMessage(req.Message)

	// Build system prompt
	systemPrompt, err := r.prompt.BuildSystemPrompt(session)
	if err != nil {
		return fmt.Errorf("building system prompt: %w", err)
	}

	// Get tools for this agent
	tools := GetToolsForAgent(session.AgentID)

	// Create streaming tool executor (supports DelegateAgent streaming)
	toolExecutor := r.createStreamingToolExecutor(session)

	// Get conversation history
	messages := session.GetMessages()

	// Track the assistant's response and tool calls for session history
	var responseText strings.Builder
	var toolCalls []ToolCall
	var pendingToolResults []struct {
		callID   string
		toolName string
		content  string
		isError  bool
	}

	// Chat with Claude using streaming tool executor
	err = r.llm.ChatWithStreamingTools(ctx, systemPrompt, messages, tools, toolExecutor, func(event ChatEvent) {
		// Track token usage
		if event.Type == "done" && event.Usage != nil {
			session.AddTokens(event.Usage.InputTokens + event.Usage.OutputTokens)
		}

		// Collect response for session history
		switch event.Type {
		case "text":
			responseText.WriteString(event.Text)
		case "tool_call":
			if event.ToolCall != nil {
				toolCalls = append(toolCalls, *event.ToolCall)
			}
		case "tool_result":
			if event.ToolResult != nil {
				pendingToolResults = append(pendingToolResults, struct {
					callID   string
					toolName string
					content  string
					isError  bool
				}{
					callID:   event.ToolResult.ToolCallID,
					toolName: findToolName(toolCalls, event.ToolResult.ToolCallID),
					content:  event.ToolResult.Content,
					isError:  event.ToolResult.IsError,
				})
			}
		}

		// Forward event to callback
		callback(event)
	})

	if err != nil {
		return fmt.Errorf("chat error: %w", err)
	}

	// Save the conversation history to session
	// Add tool calls and results in order, then the final response
	if len(toolCalls) > 0 {
		// Add assistant message with tool calls (may include partial text before tools)
		session.AddAssistantMessage("", toolCalls)

		// Add tool results
		for _, tr := range pendingToolResults {
			session.AddToolResult(tr.callID, tr.toolName, tr.content, tr.isError)
		}
	}

	// Add the final assistant response text
	if responseText.Len() > 0 {
		session.AddAssistantMessage(responseText.String(), nil)
	}

	return nil
}

// findToolName finds the tool name for a given tool call ID
func findToolName(calls []ToolCall, callID string) string {
	for _, c := range calls {
		if c.ID == callID {
			return c.Name
		}
	}
	return "unknown"
}

// createStreamingToolExecutor creates a streaming tool executor for the session
// The callback parameter allows tools like DelegateAgent to stream progress
func (r *Runtime) createStreamingToolExecutor(session *Session) StreamingToolExecutor {
	return func(ctx context.Context, call *ToolCall, callback func(ChatEvent)) (*ToolResult, error) {
		switch call.Name {
		case "Read":
			return r.executeRead(ctx, session, call)
		case "Grep":
			return r.executeGrep(ctx, session, call)
		case "Glob":
			return r.executeGlob(ctx, session, call)
		case "Bash":
			return r.executeBash(ctx, session, call)
		case "ListProjects":
			return r.executeListProjects(ctx, session, call)
		case "GetAnalysis":
			return r.executeGetAnalysis(ctx, session, call)
		case "HydrateProject":
			return r.executeHydrateProject(ctx, session, call)
		case "WebSearch":
			return r.executeWebSearch(ctx, session, call)
		case "WebFetch":
			return r.executeWebFetch(ctx, session, call)
		case "DelegateAgent":
			return r.executeDelegateAgent(ctx, session, call, callback)
		case "GetSystemInfo":
			return r.executeGetSystemInfo(ctx, session, call)
		case "GetBillingData":
			return r.executeGetBillingData(ctx, session, call)
		default:
			return &ToolResult{
				ToolCallID: call.ID,
				Content:    fmt.Sprintf("Unknown tool: %s", call.Name),
				IsError:    true,
			}, nil
		}
	}
}

// executeRead executes the Read tool
func (r *Runtime) executeRead(ctx context.Context, session *Session, call *ToolCall) (*ToolResult, error) {
	var input struct {
		FilePath string `json:"file_path"`
		Offset   int    `json:"offset"`
		Limit    int    `json:"limit"`
	}
	if err := json.Unmarshal(call.Input, &input); err != nil {
		return &ToolResult{ToolCallID: call.ID, Content: fmt.Sprintf("Invalid input: %s", err), IsError: true}, nil
	}

	// Resolve path
	filePath := r.resolvePath(session, input.FilePath)

	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return &ToolResult{ToolCallID: call.ID, Content: fmt.Sprintf("Error reading file: %s", err), IsError: true}, nil
	}

	// Apply offset and limit
	lines := strings.Split(string(content), "\n")
	offset := input.Offset
	if offset > 0 {
		offset-- // Convert to 0-indexed
	}
	limit := input.Limit
	if limit == 0 {
		limit = 500
	}

	if offset >= len(lines) {
		return &ToolResult{ToolCallID: call.ID, Content: "Offset exceeds file length", IsError: true}, nil
	}

	end := offset + limit
	if end > len(lines) {
		end = len(lines)
	}

	// Format output with line numbers
	var result strings.Builder
	for i := offset; i < end; i++ {
		result.WriteString(fmt.Sprintf("%5d | %s\n", i+1, lines[i]))
	}

	return &ToolResult{ToolCallID: call.ID, Content: result.String()}, nil
}

// executeGrep executes the Grep tool
func (r *Runtime) executeGrep(ctx context.Context, session *Session, call *ToolCall) (*ToolResult, error) {
	var input struct {
		Pattern    string `json:"pattern"`
		Path       string `json:"path"`
		Glob       string `json:"glob"`
		IgnoreCase bool   `json:"ignore_case"`
		MaxResults int    `json:"max_results"`
	}
	if err := json.Unmarshal(call.Input, &input); err != nil {
		return &ToolResult{ToolCallID: call.ID, Content: fmt.Sprintf("Invalid input: %s", err), IsError: true}, nil
	}

	// Build ripgrep command
	args := []string{"--line-number", "--no-heading", "--color=never"}

	if input.IgnoreCase {
		args = append(args, "-i")
	}
	if input.Glob != "" {
		args = append(args, "--glob", input.Glob)
	}
	if input.MaxResults > 0 {
		args = append(args, "--max-count", fmt.Sprintf("%d", input.MaxResults))
	} else {
		args = append(args, "--max-count", "50")
	}

	args = append(args, input.Pattern)

	path := input.Path
	if path == "" && session.ProjectID != "" {
		path = filepath.Join(r.zeroHome, "repos", session.ProjectID, "repo")
	}
	if path != "" {
		args = append(args, path)
	}

	cmd := exec.CommandContext(ctx, "rg", args...)
	output, err := cmd.Output()
	if err != nil {
		// ripgrep returns exit code 1 if no matches found
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return &ToolResult{ToolCallID: call.ID, Content: "No matches found"}, nil
		}
		return &ToolResult{ToolCallID: call.ID, Content: fmt.Sprintf("Grep error: %s", err), IsError: true}, nil
	}

	return &ToolResult{ToolCallID: call.ID, Content: string(output)}, nil
}

// executeGlob executes the Glob tool
func (r *Runtime) executeGlob(ctx context.Context, session *Session, call *ToolCall) (*ToolResult, error) {
	var input struct {
		Pattern string `json:"pattern"`
		Path    string `json:"path"`
	}
	if err := json.Unmarshal(call.Input, &input); err != nil {
		return &ToolResult{ToolCallID: call.ID, Content: fmt.Sprintf("Invalid input: %s", err), IsError: true}, nil
	}

	basePath := input.Path
	if basePath == "" && session.ProjectID != "" {
		basePath = filepath.Join(r.zeroHome, "repos", session.ProjectID, "repo")
	}
	if basePath == "" {
		basePath = "."
	}

	pattern := filepath.Join(basePath, input.Pattern)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return &ToolResult{ToolCallID: call.ID, Content: fmt.Sprintf("Glob error: %s", err), IsError: true}, nil
	}

	if len(matches) == 0 {
		return &ToolResult{ToolCallID: call.ID, Content: "No files found matching pattern"}, nil
	}

	return &ToolResult{ToolCallID: call.ID, Content: strings.Join(matches, "\n")}, nil
}

// executeBash executes the Bash tool
func (r *Runtime) executeBash(ctx context.Context, session *Session, call *ToolCall) (*ToolResult, error) {
	var input struct {
		Command    string `json:"command"`
		WorkingDir string `json:"working_dir"`
		Timeout    int    `json:"timeout"`
	}
	if err := json.Unmarshal(call.Input, &input); err != nil {
		return &ToolResult{ToolCallID: call.ID, Content: fmt.Sprintf("Invalid input: %s", err), IsError: true}, nil
	}

	// Set timeout
	timeout := time.Duration(input.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// #nosec G204 -- Command execution is intentional agent functionality; input is from LLM agent
	cmd := exec.CommandContext(ctx, "bash", "-c", input.Command)

	// Set working directory
	if input.WorkingDir != "" {
		cmd.Dir = input.WorkingDir
	} else if session.ProjectID != "" {
		cmd.Dir = filepath.Join(r.zeroHome, "repos", session.ProjectID, "repo")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return &ToolResult{
			ToolCallID: call.ID,
			Content:    fmt.Sprintf("Command failed: %s\nOutput: %s", err, string(output)),
			IsError:    true,
		}, nil
	}

	return &ToolResult{ToolCallID: call.ID, Content: string(output)}, nil
}

// executeListProjects executes the ListProjects tool
func (r *Runtime) executeListProjects(ctx context.Context, session *Session, call *ToolCall) (*ToolResult, error) {
	reposDir := filepath.Join(r.zeroHome, "repos")

	// List organizations/owners
	entries, err := os.ReadDir(reposDir)
	if err != nil {
		if os.IsNotExist(err) {
			return &ToolResult{
				ToolCallID: call.ID,
				Content:    "No projects directory found. Run `./zero hydrate owner/repo` to add your first project.",
			}, nil
		}
		return &ToolResult{ToolCallID: call.ID, Content: fmt.Sprintf("Error reading repos: %s\n\nTry: Check that ZERO_HOME is set correctly or run from the zero directory.", err), IsError: true}, nil
	}

	var projects []string
	for _, owner := range entries {
		if !owner.IsDir() {
			continue
		}
		ownerPath := filepath.Join(reposDir, owner.Name())
		repos, err := os.ReadDir(ownerPath)
		if err != nil {
			continue
		}
		for _, repo := range repos {
			if repo.IsDir() {
				projects = append(projects, fmt.Sprintf("%s/%s", owner.Name(), repo.Name()))
			}
		}
	}

	if len(projects) == 0 {
		return &ToolResult{
			ToolCallID: call.ID,
			Content:    "No projects found.\n\n**To add a project:**\n```\n./zero hydrate owner/repo\n```\n\nExample: `./zero hydrate expressjs/express`",
		}, nil
	}

	return &ToolResult{ToolCallID: call.ID, Content: fmt.Sprintf("**%d projects available:**\n%s", len(projects), strings.Join(projects, "\n"))}, nil
}

// executeGetAnalysis executes the GetAnalysis tool
func (r *Runtime) executeGetAnalysis(ctx context.Context, session *Session, call *ToolCall) (*ToolResult, error) {
	var input struct {
		ProjectID string `json:"project_id"`
		Scanner   string `json:"scanner"`
		Section   string `json:"section"`
	}
	if err := json.Unmarshal(call.Input, &input); err != nil {
		return &ToolResult{ToolCallID: call.ID, Content: fmt.Sprintf("Invalid input: %s", err), IsError: true}, nil
	}

	// Build path to analysis file
	analysisPath := filepath.Join(r.zeroHome, "repos", input.ProjectID, "analysis", input.Scanner+".json")

	content, err := os.ReadFile(analysisPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Check if project exists at all
			projectPath := filepath.Join(r.zeroHome, "repos", input.ProjectID)
			if _, err := os.Stat(projectPath); os.IsNotExist(err) {
				return &ToolResult{
					ToolCallID: call.ID,
					Content:    fmt.Sprintf("Project '%s' not found.\n\n**To add this project:**\n```\n./zero hydrate %s\n```", input.ProjectID, input.ProjectID),
					IsError:    true,
				}, nil
			}
			// Project exists but scanner data missing
			return &ToolResult{
				ToolCallID: call.ID,
				Content:    fmt.Sprintf("Scanner '%s' data not found for %s.\n\n**To run this scanner:**\n```\n./zero hydrate %s %s\n```\n\n**Available scanners:** code-packages, code-security, code-quality, devops, technology-identification, code-ownership, developer-experience", input.Scanner, input.ProjectID, input.ProjectID, input.Scanner),
				IsError:    true,
			}, nil
		}
		return &ToolResult{ToolCallID: call.ID, Content: fmt.Sprintf("Error reading analysis: %s", err), IsError: true}, nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		return &ToolResult{ToolCallID: call.ID, Content: fmt.Sprintf("Error parsing JSON: %s", err), IsError: true}, nil
	}

	// If section specified, extract just that section
	if input.Section != "" {
		// Navigate to section (supports dot notation like "findings.vulnerabilities")
		parts := strings.Split(input.Section, ".")
		var current interface{} = data
		for _, part := range parts {
			if m, ok := current.(map[string]interface{}); ok {
				current = m[part]
			} else {
				return &ToolResult{ToolCallID: call.ID, Content: fmt.Sprintf("Section not found: %s", input.Section), IsError: true}, nil
			}
		}

		sectionJSON, err := json.MarshalIndent(current, "", "  ")
		if err != nil {
			return &ToolResult{ToolCallID: call.ID, Content: fmt.Sprintf("Error formatting section: %s", err), IsError: true}, nil
		}

		// Truncate section output if too large
		result := string(sectionJSON)
		if len(result) > 30000 {
			result = result[:30000] + "\n\n[Section truncated - try a more specific section path]"
		}
		return &ToolResult{ToolCallID: call.ID, Content: result}, nil
	}

	// No section specified - return summary with available sections
	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("Analysis: %s for %s\n", input.Scanner, input.ProjectID))
	summary.WriteString(fmt.Sprintf("File size: %d bytes\n\n", len(content)))
	summary.WriteString("Available sections:\n")

	for key, value := range data {
		switch v := value.(type) {
		case map[string]interface{}:
			summary.WriteString(fmt.Sprintf("  - %s (object with %d keys)\n", key, len(v)))
			// Show sub-keys for important sections
			if key == "summary" || key == "findings" || key == "results" {
				for subKey := range v {
					summary.WriteString(fmt.Sprintf("      .%s\n", subKey))
				}
			}
		case []interface{}:
			summary.WriteString(fmt.Sprintf("  - %s (array with %d items)\n", key, len(v)))
		default:
			summary.WriteString(fmt.Sprintf("  - %s\n", key))
		}
	}

	summary.WriteString("\nUse 'section' parameter to get specific data, e.g.:\n")
	summary.WriteString("  section: \"summary\"\n")
	summary.WriteString("  section: \"findings.secrets\"\n")

	// Also include summary section if it exists and is small enough
	if summaryData, ok := data["summary"]; ok {
		summaryJSON, err := json.MarshalIndent(summaryData, "", "  ")
		if err == nil && len(summaryJSON) < 5000 {
			summary.WriteString("\n--- Summary Section ---\n")
			summary.WriteString(string(summaryJSON))
		}
	}

	return &ToolResult{ToolCallID: call.ID, Content: summary.String()}, nil
}

// executeHydrateProject executes the HydrateProject tool
func (r *Runtime) executeHydrateProject(ctx context.Context, session *Session, call *ToolCall) (*ToolResult, error) {
	var input struct {
		Target  string `json:"target"`
		Profile string `json:"profile"`
		Limit   int    `json:"limit"`
	}
	if err := json.Unmarshal(call.Input, &input); err != nil {
		return &ToolResult{ToolCallID: call.ID, Content: fmt.Sprintf("Invalid input: %s", err), IsError: true}, nil
	}

	if input.Target == "" {
		return &ToolResult{ToolCallID: call.ID, Content: "Target is required (owner/repo or org name)", IsError: true}, nil
	}

	// Build the hydrate command
	args := []string{"hydrate", input.Target}

	if input.Profile != "" {
		args = append(args, input.Profile)
	}

	if input.Limit > 0 {
		args = append(args, "--limit", fmt.Sprintf("%d", input.Limit))
	}

	// Find the zero binary - try current directory first, then PATH
	zeroBin := "./zero"
	if _, err := os.Stat(zeroBin); os.IsNotExist(err) {
		// Try to find in PATH
		if path, err := exec.LookPath("zero"); err == nil {
			zeroBin = path
		} else {
			return &ToolResult{
				ToolCallID: call.ID,
				Content:    "Could not find zero binary. Make sure ./zero exists or zero is in PATH.",
				IsError:    true,
			}, nil
		}
	}

	// Set a longer timeout for hydration (can take several minutes)
	cmdCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, zeroBin, args...)

	// Set ZERO_HOME if we have it
	cmd.Env = os.Environ()
	if r.zeroHome != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("ZERO_HOME=%s", r.zeroHome))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if it was a timeout
		if cmdCtx.Err() == context.DeadlineExceeded {
			return &ToolResult{
				ToolCallID: call.ID,
				Content:    fmt.Sprintf("Hydration timed out after 30 minutes. Partial output:\n%s", string(output)),
				IsError:    true,
			}, nil
		}
		return &ToolResult{
			ToolCallID: call.ID,
			Content:    fmt.Sprintf("Hydration failed: %s\nOutput: %s", err, string(output)),
			IsError:    true,
		}, nil
	}

	// Truncate output if too long
	result := string(output)
	if len(result) > 10000 {
		result = result[:10000] + "\n\n[Output truncated...]"
	}

	return &ToolResult{ToolCallID: call.ID, Content: result}, nil
}

// executeWebSearch executes the WebSearch tool
func (r *Runtime) executeWebSearch(ctx context.Context, session *Session, call *ToolCall) (*ToolResult, error) {
	var input struct {
		Query      string `json:"query"`
		MaxResults int    `json:"max_results"`
	}
	if err := json.Unmarshal(call.Input, &input); err != nil {
		return &ToolResult{ToolCallID: call.ID, Content: fmt.Sprintf("Invalid input: %s", err), IsError: true}, nil
	}

	// TODO: Implement actual web search via API
	// For now, return a placeholder
	return &ToolResult{
		ToolCallID: call.ID,
		Content:    fmt.Sprintf("Web search for '%s' - This feature requires integration with a search API.", input.Query),
	}, nil
}

// executeWebFetch executes the WebFetch tool
func (r *Runtime) executeWebFetch(ctx context.Context, session *Session, call *ToolCall) (*ToolResult, error) {
	var input struct {
		URL         string `json:"url"`
		ExtractText bool   `json:"extract_text"`
	}
	if err := json.Unmarshal(call.Input, &input); err != nil {
		return &ToolResult{ToolCallID: call.ID, Content: fmt.Sprintf("Invalid input: %s", err), IsError: true}, nil
	}

	// TODO: Implement actual web fetch
	// For now, return a placeholder
	return &ToolResult{
		ToolCallID: call.ID,
		Content:    fmt.Sprintf("Fetching URL '%s' - This feature requires HTTP client implementation.", input.URL),
	}, nil
}

// executeDelegateAgent executes the DelegateAgent tool with streaming support
func (r *Runtime) executeDelegateAgent(ctx context.Context, session *Session, call *ToolCall, parentCallback func(ChatEvent)) (*ToolResult, error) {
	var input struct {
		AgentID string `json:"agent_id"`
		Task    string `json:"task"`
		Context string `json:"context"`
	}
	if err := json.Unmarshal(call.Input, &input); err != nil {
		return &ToolResult{ToolCallID: call.ID, Content: fmt.Sprintf("Invalid input: %s", err), IsError: true}, nil
	}

	// Get agent info for status updates
	agentName, _, _, _ := r.loader.GetAgentInfo(input.AgentID)
	if agentName == "" {
		agentName = input.AgentID
	}

	// Send delegation start event to parent
	parentCallback(ChatEvent{
		Type:           "delegation",
		DelegatedAgent: agentName,
		DelegatedEvent: "start",
	})

	// Create a new session for the delegated agent
	delegateSession := r.sessions.Create(input.AgentID)
	delegateSession.SetProject(session.ProjectID)
	delegateSession.SetVoiceMode(session.VoiceMode)

	// Build the delegation task
	task := input.Task
	if input.Context != "" {
		task = fmt.Sprintf("%s\n\nContext: %s", task, input.Context)
	}

	// Build system prompt for delegate
	systemPrompt, err := r.prompt.BuildSystemPrompt(delegateSession)
	if err != nil {
		return &ToolResult{ToolCallID: call.ID, Content: fmt.Sprintf("Error building delegate prompt: %s", err), IsError: true}, nil
	}

	// Add delegation context
	delegationPrompt, _ := r.prompt.BuildDelegationPrompt(session.AgentID, input.AgentID, task)
	systemPrompt = systemPrompt + "\n\n" + delegationPrompt

	// Get tools for delegate
	tools := GetToolsForAgent(input.AgentID)

	// Create streaming tool executor for delegate (supports nested delegation)
	toolExecutor := r.createStreamingToolExecutor(delegateSession)

	// Add task as user message
	delegateSession.AddUserMessage(task)
	messages := delegateSession.GetMessages()

	// Collect response and stream events to parent
	var response strings.Builder
	err = r.llm.ChatWithStreamingTools(ctx, systemPrompt, messages, tools, toolExecutor, func(event ChatEvent) {
		switch event.Type {
		case "text":
			response.WriteString(event.Text)
			// Stream text from sub-agent to parent
			parentCallback(ChatEvent{
				Type:           "delegation",
				Text:           event.Text,
				DelegatedAgent: agentName,
				DelegatedEvent: "text",
			})
		case "tool_call":
			// Notify parent about sub-agent's tool calls
			parentCallback(ChatEvent{
				Type:           "delegation",
				ToolCall:       event.ToolCall,
				DelegatedAgent: agentName,
				DelegatedEvent: "tool_call",
			})
		case "tool_result":
			// Notify parent about sub-agent's tool results
			parentCallback(ChatEvent{
				Type:           "delegation",
				ToolResult:     event.ToolResult,
				DelegatedAgent: agentName,
				DelegatedEvent: "tool_result",
			})
		}
	})

	// Send delegation end event
	parentCallback(ChatEvent{
		Type:           "delegation",
		DelegatedAgent: agentName,
		DelegatedEvent: "done",
	})

	if err != nil {
		return &ToolResult{ToolCallID: call.ID, Content: fmt.Sprintf("Delegation error: %s", err), IsError: true}, nil
	}

	// Format response with agent attribution
	return &ToolResult{
		ToolCallID: call.ID,
		Content:    fmt.Sprintf("**Response from %s:**\n\n%s", agentName, response.String()),
	}, nil
}

// resolvePath resolves a path relative to the project
func (r *Runtime) resolvePath(session *Session, path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	if session.ProjectID != "" {
		// Check if it's an analysis file
		if strings.HasSuffix(path, ".json") && !strings.Contains(path, "/") {
			analysisPath := filepath.Join(r.zeroHome, "repos", session.ProjectID, "analysis", path)
			if _, err := os.Stat(analysisPath); err == nil {
				return analysisPath
			}
		}
		// Otherwise, resolve relative to repo
		return filepath.Join(r.zeroHome, "repos", session.ProjectID, "repo", path)
	}

	return path
}

// GetSession returns a session by ID
func (r *Runtime) GetSession(id string) (*Session, bool) {
	return r.sessions.Get(id)
}

// ListSessions returns all sessions
func (r *Runtime) ListSessions() []*Session {
	return r.sessions.List()
}

// DeleteSession deletes a session
func (r *Runtime) DeleteSession(id string) {
	r.sessions.Delete(id)
}

// GetAgentInfo returns info about an agent
func (r *Runtime) GetAgentInfo(agentID string) (name, persona, character string, ok bool) {
	return r.loader.GetAgentInfo(agentID)
}

// ListAgents returns available agent IDs
func (r *Runtime) ListAgents() []string {
	return r.loader.ListAvailable()
}

// GetGreeting returns a greeting for an agent
func (r *Runtime) GetGreeting(agentID, projectID string) (string, error) {
	return r.prompt.GetAgentGreeting(agentID, projectID)
}

// executeGetSystemInfo executes the GetSystemInfo tool
func (r *Runtime) executeGetSystemInfo(ctx context.Context, session *Session, call *ToolCall) (*ToolResult, error) {
	var input struct {
		Category string `json:"category"`
		Filter   string `json:"filter"`
	}
	if err := json.Unmarshal(call.Input, &input); err != nil {
		return &ToolResult{ToolCallID: call.ID, Content: fmt.Sprintf("Invalid input: %s", err), IsError: true}, nil
	}

	if input.Category == "" {
		return &ToolResult{
			ToolCallID: call.ID,
			Content:    "Category is required. Valid categories: rag-stats, rag-patterns, rules-status, feeds-status, scanners, profiles, config, agents, versions, help",
			IsError:    true,
		}, nil
	}

	sysInfo := NewSystemInfo(r.zeroHome)
	result, err := sysInfo.GetSystemInfo(input.Category, input.Filter)
	if err != nil {
		return &ToolResult{ToolCallID: call.ID, Content: fmt.Sprintf("Error: %s", err), IsError: true}, nil
	}

	return &ToolResult{ToolCallID: call.ID, Content: result}, nil
}

// executeGetBillingData executes the GetBillingData tool
func (r *Runtime) executeGetBillingData(ctx context.Context, session *Session, call *ToolCall) (*ToolResult, error) {
	var input struct {
		Owner string `json:"owner"`
		Type  string `json:"type"`
	}
	if err := json.Unmarshal(call.Input, &input); err != nil {
		return &ToolResult{ToolCallID: call.ID, Content: fmt.Sprintf("Invalid input: %s", err), IsError: true}, nil
	}

	if input.Owner == "" {
		return &ToolResult{
			ToolCallID: call.ID,
			Content:    "Owner is required (organization name or username)",
			IsError:    true,
		}, nil
	}

	if input.Type == "" {
		return &ToolResult{
			ToolCallID: call.ID,
			Content:    "Type is required. Valid types: actions, packages, storage, summary",
			IsError:    true,
		}, nil
	}

	client := github.NewClient()
	if !client.HasToken() {
		return &ToolResult{
			ToolCallID: call.ID,
			Content:    "GitHub token not configured. Set GITHUB_TOKEN with read:org or admin:org scope to access billing data.",
			IsError:    true,
		}, nil
	}

	var result interface{}
	var err error

	switch input.Type {
	case "actions":
		result, err = client.GetOrgActionsBilling(input.Owner)
	case "packages":
		result, err = client.GetOrgPackagesBilling(input.Owner)
	case "storage":
		result, err = client.GetOrgStorageBilling(input.Owner)
	case "summary":
		result, err = client.GetOrgBillingSummary(input.Owner)
	default:
		return &ToolResult{
			ToolCallID: call.ID,
			Content:    fmt.Sprintf("Unknown billing type: %s. Valid types: actions, packages, storage, summary", input.Type),
			IsError:    true,
		}, nil
	}

	if err != nil {
		return &ToolResult{ToolCallID: call.ID, Content: fmt.Sprintf("Error fetching billing data: %s", err), IsError: true}, nil
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return &ToolResult{ToolCallID: call.ID, Content: fmt.Sprintf("Error formatting response: %s", err), IsError: true}, nil
	}

	return &ToolResult{ToolCallID: call.ID, Content: string(jsonData)}, nil
}
