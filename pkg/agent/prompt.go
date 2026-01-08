package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PromptBuilder builds system prompts for agents
type PromptBuilder struct {
	loader   *AgentLoader
	zeroHome string
}

// NewPromptBuilder creates a new prompt builder
func NewPromptBuilder(loader *AgentLoader, zeroHome string) *PromptBuilder {
	return &PromptBuilder{
		loader:   loader,
		zeroHome: zeroHome,
	}
}

// BuildSystemPrompt builds the complete system prompt for an agent session
func (b *PromptBuilder) BuildSystemPrompt(session *Session) (string, error) {
	agent, err := b.loader.Load(session.AgentID)
	if err != nil {
		return "", fmt.Errorf("loading agent %s: %w", session.AgentID, err)
	}

	var sb strings.Builder

	// 1. Agent Identity
	sb.WriteString(b.buildIdentitySection(agent))

	// 2. Role and Expertise
	sb.WriteString(b.buildRoleSection(agent))

	// 3. Capabilities
	if len(agent.Capabilities) > 0 {
		sb.WriteString(b.buildCapabilitiesSection(agent))
	}

	// 4. Process
	if agent.Process != "" {
		sb.WriteString(b.buildProcessSection(agent))
	}

	// 5. Project Context (if set) or Available Projects
	if session.ProjectID != "" {
		sb.WriteString(b.buildProjectContext(session))
	} else {
		sb.WriteString(b.buildNoProjectContext())
	}

	// 6. Available Tools
	sb.WriteString(b.buildToolsSection(agent))

	// 7. Delegation (if applicable)
	if len(agent.Delegation) > 0 {
		sb.WriteString(b.buildDelegationSection(agent))
	}

	// 8. Voice/Personality
	sb.WriteString(b.buildVoiceSection(agent, session.VoiceMode))

	// 9. Guidelines
	sb.WriteString(b.buildGuidelinesSection())

	return sb.String(), nil
}

// buildIdentitySection builds the agent identity section
func (b *PromptBuilder) buildIdentitySection(agent *AgentDefinition) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Agent: %s\n\n", agent.Name))
	sb.WriteString(fmt.Sprintf("You are **%s** (%s), a specialist agent in the Zero security analysis system.\n\n", agent.Name, agent.Persona))

	if agent.Domain != "" {
		sb.WriteString(fmt.Sprintf("**Domain:** %s\n\n", agent.Domain))
	}

	return sb.String()
}

// buildRoleSection builds the role section
func (b *PromptBuilder) buildRoleSection(agent *AgentDefinition) string {
	if agent.Role == "" {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("## Your Role\n\n")
	sb.WriteString(agent.Role)
	sb.WriteString("\n\n")
	return sb.String()
}

// buildCapabilitiesSection builds the capabilities section
func (b *PromptBuilder) buildCapabilitiesSection(agent *AgentDefinition) string {
	var sb strings.Builder
	sb.WriteString("## Your Capabilities\n\n")

	for _, cap := range agent.Capabilities {
		sb.WriteString(fmt.Sprintf("- %s\n", cap))
	}
	sb.WriteString("\n")

	return sb.String()
}

// buildProcessSection builds the process section
func (b *PromptBuilder) buildProcessSection(agent *AgentDefinition) string {
	var sb strings.Builder
	sb.WriteString("## Your Process\n\n")
	sb.WriteString(agent.Process)
	sb.WriteString("\n\n")
	return sb.String()
}

// buildNoProjectContext builds guidance when no project is selected
func (b *PromptBuilder) buildNoProjectContext() string {
	var sb strings.Builder

	sb.WriteString("## No Project Selected\n\n")
	sb.WriteString("**IMPORTANT**: No project/repository is currently selected for analysis.\n\n")
	sb.WriteString("If the user asks about security issues, vulnerabilities, dependencies, or anything that requires analyzing a specific codebase:\n\n")
	sb.WriteString("1. Use the `ListProjects` tool to see available hydrated projects\n")
	sb.WriteString("2. **Analyze the user's question** to understand what they're looking for:\n")
	sb.WriteString("   - If they mention a language (Go, Python, JavaScript), filter to projects using that language\n")
	sb.WriteString("   - If they mention a framework (React, Express, Django), filter to relevant projects\n")
	sb.WriteString("   - If they mention a category (frontend, backend, ML), filter appropriately\n")
	sb.WriteString("3. Present a **smart, filtered list** based on their question context\n")
	sb.WriteString("4. ASK which one they want to analyze - do NOT assume\n\n")
	sb.WriteString("**Example**: If user asks \"Check Go best practices\", you should:\n")
	sb.WriteString("- List projects, identify which are Go-based\n")
	sb.WriteString("- Say: \"I found 3 Go projects: [list]. Which one should I analyze for Go best practices?\"\n\n")

	return sb.String()
}

// buildProjectContext builds the project context section
func (b *PromptBuilder) buildProjectContext(session *Session) string {
	var sb strings.Builder

	sb.WriteString("## Current Project Context\n\n")
	sb.WriteString(fmt.Sprintf("**Project:** %s\n\n", session.ProjectID))

	// Build paths to analysis data
	projectPath := filepath.Join(b.zeroHome, "repos", session.ProjectID)
	analysisPath := filepath.Join(projectPath, "analysis")
	repoPath := filepath.Join(projectPath, "repo")

	sb.WriteString("**Data Locations:**\n")
	sb.WriteString(fmt.Sprintf("- Repository: `%s`\n", repoPath))
	sb.WriteString(fmt.Sprintf("- Analysis: `%s`\n", analysisPath))
	sb.WriteString("\n")

	sb.WriteString("**Available Scanner Results:**\n")
	sb.WriteString("- `code-packages.json` - SBOM, vulnerabilities, package health, licenses, malcontent\n")
	sb.WriteString("- `code-security.json` - SAST findings, secrets, API security, cryptography\n")
	sb.WriteString("- `code-quality.json` - Tech debt, complexity, test coverage\n")
	sb.WriteString("- `devops.json` - IaC, containers, GitHub Actions, DORA metrics\n")
	sb.WriteString("- `technology-identification.json` - Technologies, ML models, frameworks\n")
	sb.WriteString("- `code-ownership.json` - Contributors, bus factor, ownership\n")
	sb.WriteString("- `developer-experience.json` - Onboarding, sprawl, workflow\n")
	sb.WriteString("\n")

	return sb.String()
}

// buildToolsSection builds the available tools section
func (b *PromptBuilder) buildToolsSection(agent *AgentDefinition) string {
	var sb strings.Builder

	sb.WriteString("## Available Tools\n\n")
	sb.WriteString("You have access to the following tools to complete your tasks:\n\n")

	// Core tools available to all agents
	tools := []struct {
		Name        string
		Description string
	}{
		{"Read", "Read file contents from the repository or analysis data"},
		{"Grep", "Search for patterns in files using regex"},
		{"Glob", "Find files matching a pattern"},
		{"Bash", "Execute shell commands (sandboxed)"},
		{"ListProjects", "List all hydrated projects"},
		{"GetAnalysis", "Get scanner results for a project"},
		{"GetSystemInfo", "Get information about Zero itself (patterns, scanners, rules, agents, config)"},
		{"WebSearch", "Search the web for information"},
		{"WebFetch", "Fetch content from a URL"},
	}

	// Add agent-specific tools
	if agent.ID == "zero" {
		tools = append(tools, struct {
			Name        string
			Description string
		}{"DelegateAgent", "Delegate to a specialist agent"})
	}

	for _, tool := range tools {
		sb.WriteString(fmt.Sprintf("- **%s**: %s\n", tool.Name, tool.Description))
	}
	sb.WriteString("\n")

	sb.WriteString("Use tools proactively to investigate and gather information. Don't guess - verify with tools.\n\n")

	return sb.String()
}

// buildDelegationSection builds the delegation section
func (b *PromptBuilder) buildDelegationSection(agent *AgentDefinition) string {
	var sb strings.Builder

	sb.WriteString("## Agent Delegation\n\n")
	sb.WriteString("You can delegate to other specialist agents when their expertise is needed:\n\n")

	sb.WriteString("| Scenario | Delegate To | Agent ID |\n")
	sb.WriteString("|----------|-------------|----------|\n")

	for _, rule := range agent.Delegation {
		sb.WriteString(fmt.Sprintf("| %s | %s | `%s` |\n", rule.Scenario, rule.DelegateTo, rule.AgentID))
	}
	sb.WriteString("\n")

	sb.WriteString("To delegate, use the DelegateAgent tool with the agent ID and a clear prompt.\n\n")

	return sb.String()
}

// buildVoiceSection builds the voice/personality section
func (b *PromptBuilder) buildVoiceSection(agent *AgentDefinition, voiceMode string) string {
	var sb strings.Builder

	sb.WriteString("## Communication Style\n\n")

	switch voiceMode {
	case "full":
		if agent.VoiceFull != "" {
			sb.WriteString(agent.VoiceFull)
		} else {
			sb.WriteString(b.defaultVoiceFull(agent))
		}
	case "minimal":
		if agent.VoiceMinimal != "" {
			sb.WriteString(agent.VoiceMinimal)
		} else {
			sb.WriteString(b.defaultVoiceMinimal(agent))
		}
	case "neutral":
		if agent.VoiceNeutral != "" {
			sb.WriteString(agent.VoiceNeutral)
		} else {
			sb.WriteString(b.defaultVoiceNeutral(agent))
		}
	default:
		// Default to full voice
		if agent.VoiceFull != "" {
			sb.WriteString(agent.VoiceFull)
		} else {
			sb.WriteString(b.defaultVoiceFull(agent))
		}
	}

	sb.WriteString("\n\n")
	return sb.String()
}

// defaultVoiceFull provides a default full voice if none specified
func (b *PromptBuilder) defaultVoiceFull(agent *AgentDefinition) string {
	return fmt.Sprintf(`You are %s, a security specialist. You can occasionally reference your expertise but focus on the analysis.

IMPORTANT: Do NOT start responses with "[Name] here" or announce yourself. Get straight to the point like a professional would.
Be helpful and informative. Prioritize accurate technical information over character roleplay.`, agent.Persona)
}

// defaultVoiceMinimal provides a default minimal voice
func (b *PromptBuilder) defaultVoiceMinimal(agent *AgentDefinition) string {
	return fmt.Sprintf(`You are %s, providing professional security analysis.

Do NOT announce yourself at the start of responses. Be direct, efficient, and focus on actionable findings.`, agent.Name)
}

// defaultVoiceNeutral provides a default neutral voice
func (b *PromptBuilder) defaultVoiceNeutral(agent *AgentDefinition) string {
	return fmt.Sprintf(`You are the %s module. Provide objective technical analysis.

Use formal, precise language. Focus on facts, findings, and recommendations.
No character roleplay. No self-announcement. Pure technical output.`, agent.Domain)
}

// buildGuidelinesSection builds general guidelines
func (b *PromptBuilder) buildGuidelinesSection() string {
	return `## Guidelines

1. **Be Accurate**: Verify information before stating it. Use tools to confirm.
2. **Be Concise**: Provide clear, actionable insights. Avoid unnecessary verbosity.
3. **Prioritize Risk**: Focus on high-impact findings first.
4. **Show Evidence**: Cite file paths, line numbers, and specific findings.
5. **Be Proactive**: Use tools without being asked when investigation is needed.
6. **Delegate When Needed**: If a question is outside your expertise, suggest the right specialist.
7. **Ask for Clarification**: If the user asks about a specific project/repository but no project context is set, use ListProjects to see available projects and ASK the user which one they want to analyze. Do NOT assume or guess. List the available options and ask them to choose.

`
}

// BuildDelegationPrompt builds a prompt for delegating to another agent
func (b *PromptBuilder) BuildDelegationPrompt(fromAgent, toAgent, task string) (string, error) {
	to, err := b.loader.Load(toAgent)
	if err != nil {
		return "", fmt.Errorf("loading target agent %s: %w", toAgent, err)
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Delegation from %s\n\n", fromAgent))
	sb.WriteString(fmt.Sprintf("You have been asked to assist with a task by %s.\n\n", fromAgent))
	sb.WriteString(fmt.Sprintf("**Task:** %s\n\n", task))
	sb.WriteString(fmt.Sprintf("Please apply your expertise in %s to help with this request.\n\n", to.Domain))

	return sb.String(), nil
}

// GetAgentGreeting returns a greeting message for an agent
func (b *PromptBuilder) GetAgentGreeting(agentID string, projectID string) (string, error) {
	agent, err := b.loader.Load(agentID)
	if err != nil {
		return "", fmt.Errorf("loading agent %s: %w", agentID, err)
	}

	var greeting string

	// Get project count for context
	projectCount := b.countHydratedProjects()
	projectsInfo := ""
	if projectCount > 0 {
		projectsInfo = fmt.Sprintf("\n\n**%d projects hydrated** - use `ListProjects` to see them.", projectCount)
	}

	switch agentID {
	case "zero":
		if projectID != "" {
			greeting = fmt.Sprintf("Zero here. I've got %s loaded. What do you want to dig into?", projectID)
		} else if projectCount > 0 {
			greeting = fmt.Sprintf("Zero here. Ready to investigate.%s\n\n**Try:**\n- \"Analyze security of [project]\"\n- \"What vulnerabilities are in [project]?\"\n- \"How many secrets detection patterns do we have?\"", projectsInfo)
		} else {
			greeting = "Zero here. No projects hydrated yet.\n\n**Get started:**\n- Run `zero hydrate owner/repo` to add a project\n- Ask \"What scanners are available?\" to learn about capabilities"
		}
	case "cereal":
		greeting = "FYI man, Cereal Killer here. Ready to check your supply chain."
		if projectID != "" {
			greeting = fmt.Sprintf("Alright, checking out %s. Let's see what's hiding in those dependencies.", projectID)
		}
	case "razor":
		greeting = "Razor here. Show me the code."
		if projectID != "" {
			greeting = fmt.Sprintf("Razor here. I've got %s ready for analysis. What vulnerabilities should I cut into?", projectID)
		}
	case "gill":
		greeting = "Gill here. Let's review your cryptographic implementations."
		if projectID != "" {
			greeting = fmt.Sprintf("Gill here. I'm ready to analyze the cryptography in %s.", projectID)
		}
	case "plague":
		greeting = "Plague here. Time to inspect your infrastructure."
		if projectID != "" {
			greeting = fmt.Sprintf("Plague here. Let's see what's running in %s.", projectID)
		}
	default:
		greeting = fmt.Sprintf("%s here. How can I help?", agent.Name)
		if projectID != "" {
			greeting = fmt.Sprintf("%s here. I've got %s ready. What would you like to know?", agent.Name, projectID)
		}
	}

	return greeting, nil
}

// countHydratedProjects counts how many projects are hydrated
func (b *PromptBuilder) countHydratedProjects() int {
	reposDir := filepath.Join(b.zeroHome, "repos")

	entries, err := os.ReadDir(reposDir)
	if err != nil {
		return 0
	}

	count := 0
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
				count++
			}
		}
	}

	return count
}
