package agent

// BuiltinTools returns the tool definitions for Claude
func BuiltinTools() []ToolDefinition {
	return []ToolDefinition{
		ReadTool(),
		GrepTool(),
		GlobTool(),
		BashTool(),
		ListProjectsTool(),
		GetAnalysisTool(),
		HydrateProjectTool(),
		WebSearchTool(),
		WebFetchTool(),
		DelegateAgentTool(),
		GetSystemInfoTool(),
	}
}

// ReadTool defines the Read file tool
func ReadTool() ToolDefinition {
	return ToolDefinition{
		Name:        "Read",
		Description: "Read the contents of a file. Use this to examine source code, configuration files, or analysis results.",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"file_path": {
					Type:        "string",
					Description: "The absolute path to the file to read",
				},
				"offset": {
					Type:        "integer",
					Description: "Line number to start reading from (1-indexed). Optional.",
				},
				"limit": {
					Type:        "integer",
					Description: "Maximum number of lines to read. Optional, defaults to 500.",
				},
			},
			Required: []string{"file_path"},
		},
	}
}

// GrepTool defines the Grep search tool
func GrepTool() ToolDefinition {
	return ToolDefinition{
		Name:        "Grep",
		Description: "Search for patterns in files using regex. Returns matching lines with file paths and line numbers.",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"pattern": {
					Type:        "string",
					Description: "The regex pattern to search for",
				},
				"path": {
					Type:        "string",
					Description: "Directory or file to search in. Defaults to current project.",
				},
				"glob": {
					Type:        "string",
					Description: "File pattern filter (e.g., '*.go', '*.js'). Optional.",
				},
				"ignore_case": {
					Type:        "boolean",
					Description: "Case insensitive search. Optional, defaults to false.",
				},
				"max_results": {
					Type:        "integer",
					Description: "Maximum number of results to return. Optional, defaults to 50.",
				},
			},
			Required: []string{"pattern"},
		},
	}
}

// GlobTool defines the Glob file finder tool
func GlobTool() ToolDefinition {
	return ToolDefinition{
		Name:        "Glob",
		Description: "Find files matching a glob pattern. Returns list of matching file paths.",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"pattern": {
					Type:        "string",
					Description: "The glob pattern to match (e.g., '**/*.go', 'src/**/*.ts')",
				},
				"path": {
					Type:        "string",
					Description: "Base directory to search from. Optional.",
				},
			},
			Required: []string{"pattern"},
		},
	}
}

// BashTool defines the Bash command execution tool
func BashTool() ToolDefinition {
	return ToolDefinition{
		Name:        "Bash",
		Description: "Execute a shell command. Use for git operations, running scripts, or system commands. Commands are sandboxed.",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"command": {
					Type:        "string",
					Description: "The command to execute",
				},
				"working_dir": {
					Type:        "string",
					Description: "Working directory for the command. Optional.",
				},
				"timeout": {
					Type:        "integer",
					Description: "Timeout in seconds. Optional, defaults to 30.",
				},
			},
			Required: []string{"command"},
		},
	}
}

// ListProjectsTool defines the tool to list hydrated projects
func ListProjectsTool() ToolDefinition {
	return ToolDefinition{
		Name:        "ListProjects",
		Description: "List all hydrated projects with their scan status and freshness.",
		InputSchema: InputSchema{
			Type:       "object",
			Properties: map[string]Property{},
		},
	}
}

// GetAnalysisTool defines the tool to get scanner results
func GetAnalysisTool() ToolDefinition {
	return ToolDefinition{
		Name:        "GetAnalysis",
		Description: "Get scanner analysis results for a project. Returns structured JSON data from the specified scanner.",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"project_id": {
					Type:        "string",
					Description: "Project ID in format 'owner/repo' (e.g., 'expressjs/express')",
				},
				"scanner": {
					Type:        "string",
					Description: "Scanner type to get results for",
					Enum: []string{
						"code-packages",
						"code-security",
						"code-quality",
						"devops",
						"technology-identification",
						"code-ownership",
						"developer-experience",
					},
				},
				"section": {
					Type:        "string",
					Description: "Specific section to extract (e.g., 'summary', 'findings.vulnerabilities'). Optional, returns full results if not specified.",
				},
			},
			Required: []string{"project_id", "scanner"},
		},
	}
}

// HydrateProjectTool defines the tool to hydrate (clone and scan) repositories
func HydrateProjectTool() ToolDefinition {
	return ToolDefinition{
		Name:        "HydrateProject",
		Description: "Clone and scan a repository or organization. This runs security scanners and generates analysis data. Use for 'owner/repo' format for single repos or 'orgname' for all repos in an org.",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"target": {
					Type:        "string",
					Description: "Repository (owner/repo) or organization name to hydrate",
				},
				"profile": {
					Type:        "string",
					Description: "Scan profile to use. Options: all-quick (default), all-complete, code-packages, code-security, code-quality, devops",
				},
				"limit": {
					Type:        "integer",
					Description: "Maximum number of repos to hydrate when targeting an org (default: 25)",
				},
			},
			Required: []string{"target"},
		},
	}
}

// WebSearchTool defines the web search tool
func WebSearchTool() ToolDefinition {
	return ToolDefinition{
		Name:        "WebSearch",
		Description: "Search the web for information. Use for researching CVEs, security advisories, or documentation.",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"query": {
					Type:        "string",
					Description: "The search query",
				},
				"max_results": {
					Type:        "integer",
					Description: "Maximum number of results. Optional, defaults to 5.",
				},
			},
			Required: []string{"query"},
		},
	}
}

// WebFetchTool defines the web fetch tool
func WebFetchTool() ToolDefinition {
	return ToolDefinition{
		Name:        "WebFetch",
		Description: "Fetch content from a URL. Use for retrieving documentation, security bulletins, or API responses.",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"url": {
					Type:        "string",
					Description: "The URL to fetch",
				},
				"extract_text": {
					Type:        "boolean",
					Description: "Extract plain text from HTML. Optional, defaults to true.",
				},
			},
			Required: []string{"url"},
		},
	}
}

// DelegateAgentTool defines the agent delegation tool
func DelegateAgentTool() ToolDefinition {
	return ToolDefinition{
		Name:        "DelegateAgent",
		Description: "Delegate a task to another specialist agent. Use when the task requires expertise outside your domain.",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"agent_id": {
					Type:        "string",
					Description: "The ID of the agent to delegate to",
					Enum: []string{
						"zero",    // Orchestrator
						"cereal",  // Supply chain
						"razor",   // Code security
						"blade",   // Compliance
						"phreak",  // Legal
						"acid",    // Frontend
						"dade",    // Backend
						"nikon",   // Architecture
						"joey",    // Build
						"plague",  // DevOps
						"gibson",  // Engineering metrics
						"gill",    // Cryptography
						"hal",  // AI/ML security
					},
				},
				"task": {
					Type:        "string",
					Description: "Clear description of the task to delegate",
				},
				"context": {
					Type:        "string",
					Description: "Additional context for the delegated agent. Optional.",
				},
			},
			Required: []string{"agent_id", "task"},
		},
	}
}

// GetSystemInfoTool defines the tool to query Zero system metadata
func GetSystemInfoTool() ToolDefinition {
	return ToolDefinition{
		Name:        "GetSystemInfo",
		Description: "Get information about the Zero system itself - RAG patterns, rules, feeds, scanner capabilities, agents, and configuration. Use this to answer questions about Zero's detection capabilities.",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"category": {
					Type:        "string",
					Description: "Category of system info to retrieve",
					Enum: []string{
						"rag-stats",     // Pattern counts by RAG category
						"rag-patterns",  // List patterns in a specific RAG category
						"rules-status",  // Generated/community rule status
						"feeds-status",  // Feed sync status
						"scanners",      // Scanner inventory with features
						"profiles",      // Available scan profiles
						"config",        // Active configuration summary
						"agents",        // Available specialist agents
						"versions",      // Zero and scanner versions
						"help",          // Example questions and capabilities
					},
				},
				"filter": {
					Type:        "string",
					Description: "Optional filter (e.g., 'secrets' for rag-patterns, 'code-security' for scanners)",
				},
			},
			Required: []string{"category"},
		},
	}
}

// GetToolsForAgent returns the tools available for a specific agent
func GetToolsForAgent(agentID string) []ToolDefinition {
	// All agents get base tools
	tools := []ToolDefinition{
		ReadTool(),
		GrepTool(),
		GlobTool(),
		ListProjectsTool(),
		GetAnalysisTool(),
		GetSystemInfoTool(), // All agents can query system info
	}

	// Add web tools for investigation
	tools = append(tools, WebSearchTool(), WebFetchTool())

	// Add bash and hydrate for certain agents
	switch agentID {
	case "zero", "plague", "joey", "dade":
		tools = append(tools, BashTool(), HydrateProjectTool())
	}

	// Add delegation for orchestrator and specialists
	switch agentID {
	case "zero", "cereal", "razor", "blade", "nikon":
		tools = append(tools, DelegateAgentTool())
	}

	return tools
}
