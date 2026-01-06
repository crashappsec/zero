# Zero Agent Architecture & Implementation Plan

## Executive Summary

Build a fully interactive agent system that:
1. Runs from CLI with rich REPL experience
2. Exposes via HTTP/WebSocket for web UI
3. Uses MCP for tool execution (Read, Grep, WebSearch, etc.)
4. Loads agent definitions from `agents/*.md` files
5. Supports autonomous investigation with tool use

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              ZERO AGENT SYSTEM                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐                 │
│  │   CLI REPL   │     │  HTTP/WS API │     │   Web UI     │                 │
│  │  (terminal)  │     │   (server)   │     │  (future)    │                 │
│  └──────┬───────┘     └──────┬───────┘     └──────┬───────┘                 │
│         │                    │                    │                          │
│         └────────────────────┼────────────────────┘                          │
│                              ▼                                               │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │                         AGENT RUNTIME                                  │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐   │  │
│  │  │   Session   │  │   Agent     │  │   Prompt    │  │   Tool      │   │  │
│  │  │   Manager   │  │   Loader    │  │   Builder   │  │   Router    │   │  │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └──────┬──────┘   │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
│                                                                     │        │
│                              ┌──────────────────────────────────────┘        │
│                              ▼                                               │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │                         MCP CLIENT                                     │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐   │  │
│  │  │    Read     │  │    Grep     │  │  WebSearch  │  │   Custom    │   │  │
│  │  │    Tool     │  │    Tool     │  │    Tool     │  │   Tools     │   │  │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘   │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
│                              │                                               │
│                              ▼                                               │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │                      LLM PROVIDER                                      │  │
│  │  ┌─────────────────────────────────────────────────────────────────┐  │  │
│  │  │  Claude API (with tool_use support)                             │  │  │
│  │  │  - Streaming responses                                          │  │  │
│  │  │  - Tool call/response loop                                      │  │  │
│  │  │  - Multi-turn conversations                                     │  │  │
│  │  └─────────────────────────────────────────────────────────────────┘  │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Component Design

### 1. Agent Runtime (`pkg/agent/`)

The core runtime that powers both CLI and web experiences.

```
pkg/agent/
├── runtime.go          # Main runtime orchestrator
├── session.go          # Session management (enhanced from api/agent/types.go)
├── loader.go           # Load agent.md files, parse voice modes
├── prompt.go           # Build system prompts from agent definitions
├── conversation.go     # Multi-turn conversation management
└── delegation.go       # Agent-to-agent delegation
```

#### runtime.go - Core Interface

```go
package agent

// Runtime is the main agent execution engine
type Runtime struct {
    loader      *AgentLoader
    sessions    *SessionManager
    mcpClient   *mcp.Client
    llmClient   *LLMClient
    zeroHome    string
}

// NewRuntime creates a new agent runtime
func NewRuntime(opts *RuntimeOptions) (*Runtime, error)

// Chat sends a message and returns a streaming response
func (r *Runtime) Chat(ctx context.Context, req *ChatRequest) (<-chan *StreamEvent, error)

// ExecuteTool runs a tool and returns the result
func (r *Runtime) ExecuteTool(ctx context.Context, call *ToolCall) (*ToolResult, error)

// SwitchAgent changes the active agent in a session
func (r *Runtime) SwitchAgent(sessionID, agentID string) error
```

### 2. Agent Loader (`pkg/agent/loader.go`)

Loads and parses agent definitions from markdown files.

```go
package agent

// AgentDefinition represents a fully loaded agent
type AgentDefinition struct {
    ID          string
    Name        string
    Persona     string
    Domain      string
    Role        string            // From ## Role section
    Capabilities []string         // From ## Capabilities
    Process     string            // From ## Process
    VoiceFull   string            // From <!-- VOICE:full -->
    VoiceMin    string            // From <!-- VOICE:minimal -->
    VoiceNeutral string           // From <!-- VOICE:neutral -->
    Knowledge   []string          // Paths to knowledge files
    DataSources []string          // Scanner outputs this agent uses
    Delegation  []DelegationRule  // Who can delegate to whom
    Tools       []ToolSpec        // Available tools
}

// AgentLoader loads agents from the agents/ directory
type AgentLoader struct {
    agentsDir string
    cache     map[string]*AgentDefinition
}

// Load loads an agent definition by ID
func (l *AgentLoader) Load(agentID string) (*AgentDefinition, error)

// LoadAll loads all available agents
func (l *AgentLoader) LoadAll() ([]*AgentDefinition, error)

// ParseMarkdown parses an agent.md file into AgentDefinition
func (l *AgentLoader) ParseMarkdown(content []byte) (*AgentDefinition, error)
```

### 3. MCP Integration (`pkg/agent/mcp.go`)

Tools executed via MCP protocol for consistency with Claude Code.

```go
package agent

// MCPToolProvider provides tools via MCP
type MCPToolProvider struct {
    client   *mcp.Client
    zeroHome string
}

// Available tools exposed via MCP
var BuiltinTools = []ToolSpec{
    {Name: "Read", Description: "Read file contents"},
    {Name: "Grep", Description: "Search for patterns in files"},
    {Name: "Glob", Description: "Find files matching pattern"},
    {Name: "WebSearch", Description: "Search the web"},
    {Name: "WebFetch", Description: "Fetch URL contents"},
    {Name: "Bash", Description: "Execute shell commands"},
    {Name: "ListProjects", Description: "List hydrated projects"},
    {Name: "GetAnalysis", Description: "Get scanner results"},
}

// ExecuteTool executes a tool via MCP
func (p *MCPToolProvider) ExecuteTool(ctx context.Context, call *ToolCall) (*ToolResult, error)
```

### 4. LLM Client with Tool Use (`pkg/agent/llm.go`)

Enhanced Claude client that handles the tool use loop.

```go
package agent

// LLMClient handles Claude API with tool use
type LLMClient struct {
    apiKey     string
    model      string
    httpClient *http.Client
}

// ChatWithTools sends a message and handles tool calls automatically
func (c *LLMClient) ChatWithTools(
    ctx context.Context,
    systemPrompt string,
    messages []Message,
    tools []ToolSpec,
    toolExecutor ToolExecutor,
    callback StreamCallback,
) error

// The tool use loop:
// 1. Send message to Claude with tools
// 2. If Claude returns tool_use, execute tool
// 3. Send tool_result back to Claude
// 4. Repeat until Claude returns text (end_turn)
```

### 5. CLI Command (`cmd/zero/cmd/agent.go`)

Interactive REPL for terminal use.

```go
package cmd

var agentCmd = &cobra.Command{
    Use:   "agent [project]",
    Short: "Start interactive agent chat",
    Long: `Start an interactive chat session with Zero agents.

Examples:
  zero agent                      # Chat with Zero (orchestrator)
  zero agent expressjs/express    # Chat about a specific project
  zero agent --agent cereal       # Chat with Cereal (supply chain)
  zero agent --query "list vulns" # One-shot query mode
  zero agent --voice minimal      # Use minimal voice mode`,
    RunE: runAgent,
}

func init() {
    rootCmd.AddCommand(agentCmd)

    agentCmd.Flags().StringVarP(&agentID, "agent", "a", "zero", "Agent to chat with")
    agentCmd.Flags().StringVarP(&voiceMode, "voice", "v", "full", "Voice mode: full, minimal, neutral")
    agentCmd.Flags().StringVarP(&query, "query", "q", "", "One-shot query (non-interactive)")
    agentCmd.Flags().BoolVar(&noStream, "no-stream", false, "Disable streaming output")
}
```

---

## Data Flow

### Interactive Chat Flow

```
User Input (CLI/Web)
        │
        ▼
┌───────────────────┐
│   Agent Runtime   │
│   - Get session   │
│   - Load agent    │
│   - Build prompt  │
└─────────┬─────────┘
          │
          ▼
┌───────────────────┐
│    LLM Client     │
│   - Send message  │
│   - Stream resp   │
└─────────┬─────────┘
          │
          ▼
    ┌─────────────┐
    │ Tool Call?  │──No──▶ Stream text to user
    └──────┬──────┘
           │ Yes
           ▼
┌───────────────────┐
│   MCP Client      │
│   - Execute tool  │
│   - Return result │
└─────────┬─────────┘
          │
          ▼
┌───────────────────┐
│   LLM Client      │
│   - Send result   │
│   - Continue...   │
└─────────┬─────────┘
          │
          ▼
    (Loop until end_turn)
```

### Tool Execution via MCP

```
┌─────────────────────────────────────────────────────────────┐
│                     MCP Tool Execution                       │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Claude returns:                                             │
│  {                                                           │
│    "type": "tool_use",                                       │
│    "name": "Read",                                           │
│    "input": {"file_path": "/path/to/file.go"}               │
│  }                                                           │
│                                                              │
│              │                                               │
│              ▼                                               │
│  ┌─────────────────────┐                                    │
│  │   MCP Tool Router   │                                    │
│  │   - Map tool name   │                                    │
│  │   - Validate input  │                                    │
│  │   - Execute         │                                    │
│  └──────────┬──────────┘                                    │
│             │                                                │
│             ▼                                                │
│  ┌─────────────────────┐     ┌─────────────────────┐       │
│  │   Built-in Tools    │     │   Zero MCP Server   │       │
│  │   - Read            │     │   - ListProjects    │       │
│  │   - Grep            │     │   - GetAnalysis     │       │
│  │   - Glob            │     │   - GetVulns        │       │
│  │   - WebSearch       │     │   - GetSecrets      │       │
│  │   - Bash            │     │   - GetMalcontent   │       │
│  └─────────────────────┘     └─────────────────────┘       │
│                                                              │
│  Result returned to Claude:                                  │
│  {                                                           │
│    "type": "tool_result",                                    │
│    "content": "... file contents ..."                        │
│  }                                                           │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## MCP Server Enhancement

Extend the existing MCP server to support agent tools.

### Current MCP Tools (pkg/mcp/)

```
list_projects      - List hydrated projects
get_project_summary - Get project analysis summary
get_vulnerabilities - Query vulnerabilities
get_malcontent     - Get malware findings
get_secrets        - Get detected secrets
```

### New MCP Tools for Agents

```
# File System Tools
read_file          - Read file contents (with line limits)
search_files       - Grep for patterns
find_files         - Glob for file patterns

# Web Tools
web_search         - Search the web (via API)
web_fetch          - Fetch URL contents

# Execution Tools
run_command        - Execute shell command (sandboxed)

# Agent Tools
delegate_agent     - Invoke another agent
get_agent_info     - Get agent capabilities
```

### MCP Server Architecture

```
pkg/mcp/
├── server.go           # Existing MCP server
├── tools/
│   ├── projects.go     # Project/analysis tools (existing)
│   ├── filesystem.go   # Read, Grep, Glob tools (new)
│   ├── web.go          # WebSearch, WebFetch (new)
│   ├── execution.go    # Bash tool (new, sandboxed)
│   └── agents.go       # Agent delegation tools (new)
└── client.go           # MCP client for agent runtime (new)
```

---

## Implementation Plan

### Phase 1: Core Runtime (Week 1)

#### 1.1 Agent Loader
- [ ] Create `pkg/agent/loader.go`
- [ ] Parse agent.md markdown format
- [ ] Extract voice modes from HTML comments
- [ ] Cache loaded agents
- [ ] Unit tests

#### 1.2 Session Manager
- [ ] Enhance `pkg/agent/session.go` from existing types.go
- [ ] Add project context binding
- [ ] Add conversation history with tool calls
- [ ] Add persistence (optional, file-based)

#### 1.3 Prompt Builder
- [ ] Create `pkg/agent/prompt.go`
- [ ] Build system prompt from AgentDefinition
- [ ] Include project context and data locations
- [ ] Support voice mode selection

### Phase 2: LLM Integration (Week 1-2)

#### 2.1 Tool Use Support
- [ ] Create `pkg/agent/llm.go`
- [ ] Implement Claude tool_use protocol
- [ ] Handle streaming with tool calls
- [ ] Implement tool execution loop

#### 2.2 Tool Definitions
- [ ] Define tool schemas in Claude format
- [ ] Map tools to MCP calls
- [ ] Handle tool errors gracefully

### Phase 3: MCP Enhancement (Week 2)

#### 3.1 File System Tools
- [ ] Implement `read_file` tool
- [ ] Implement `search_files` (grep) tool
- [ ] Implement `find_files` (glob) tool
- [ ] Add security boundaries

#### 3.2 Web Tools
- [ ] Implement `web_search` tool (using existing WebSearch)
- [ ] Implement `web_fetch` tool
- [ ] Rate limiting and caching

#### 3.3 MCP Client
- [ ] Create `pkg/mcp/client.go`
- [ ] Connect to MCP server
- [ ] Execute tools via MCP protocol
- [ ] Handle responses

### Phase 4: CLI Experience (Week 2-3)

#### 4.1 Agent Command
- [ ] Create `cmd/zero/cmd/agent.go`
- [ ] Implement REPL loop
- [ ] Streaming output with formatting
- [ ] Handle Ctrl+C, exit commands

#### 4.2 Terminal UI
- [ ] Colored output (agent personas)
- [ ] Markdown rendering
- [ ] Code syntax highlighting
- [ ] Progress indicators for tool calls

#### 4.3 One-Shot Mode
- [ ] `--query` flag for single questions
- [ ] Pipe-friendly output
- [ ] JSON output option

### Phase 5: HTTP/WebSocket API (Week 3)

#### 5.1 Enhanced API Endpoints
- [ ] Update `/api/chat` to use new runtime
- [ ] Add tool call events to streaming
- [ ] Add agent switching endpoint

#### 5.2 WebSocket Enhancements
- [ ] Stream tool calls to client
- [ ] Stream tool results
- [ ] Handle reconnection

### Phase 6: Agent Delegation (Week 3-4)

#### 6.1 Delegation System
- [ ] Create `pkg/agent/delegation.go`
- [ ] Parse delegation rules from agent.md
- [ ] Implement agent-to-agent calls
- [ ] Share context between agents

#### 6.2 Orchestrator Intelligence
- [ ] Zero routes to specialists
- [ ] Maintain conversation context
- [ ] Synthesize multi-agent responses

---

## File Structure

```
zero/
├── cmd/zero/cmd/
│   └── agent.go                    # NEW: CLI command
│
├── pkg/agent/                       # NEW: Agent runtime package
│   ├── runtime.go                  # Main runtime
│   ├── session.go                  # Session management
│   ├── loader.go                   # Agent.md loader
│   ├── prompt.go                   # Prompt builder
│   ├── llm.go                      # LLM client with tools
│   ├── conversation.go             # Conversation management
│   ├── delegation.go               # Agent delegation
│   ├── tools.go                    # Tool definitions
│   └── runtime_test.go             # Tests
│
├── pkg/mcp/
│   ├── server.go                   # Existing MCP server
│   ├── client.go                   # NEW: MCP client
│   └── tools/
│       ├── projects.go             # Existing tools
│       ├── filesystem.go           # NEW: Read/Grep/Glob
│       ├── web.go                  # NEW: WebSearch/Fetch
│       └── execution.go            # NEW: Bash (sandboxed)
│
├── pkg/api/agent/
│   ├── handler.go                  # Updated to use runtime
│   └── claude.go                   # Deprecated, use pkg/agent/llm.go
│
└── agents/                          # Agent definitions (existing)
    ├── orchestrator/agent.md
    ├── supply-chain/agent.md
    └── ...
```

---

## API Design

### CLI Usage

```bash
# Interactive mode
zero agent                           # Chat with Zero
zero agent expressjs/express         # With project context
zero agent --agent cereal            # Specific agent
zero agent --voice minimal           # Voice mode

# One-shot mode
zero agent -q "What vulnerabilities are in express?"
zero agent expressjs/express -q "Show me the critical CVEs"

# Pipe mode
echo "List all high severity findings" | zero agent expressjs/express
zero agent expressjs/express -q "..." --json | jq '.findings'
```

### HTTP API

```
POST /api/v2/chat
{
  "session_id": "uuid",           # Optional, creates new if empty
  "agent_id": "cereal",           # Optional, defaults to "zero"
  "project_id": "owner/repo",     # Optional project context
  "message": "Analyze the vulns",
  "voice_mode": "full",           # full, minimal, neutral
  "stream": true                  # Enable streaming
}

Response (streaming SSE):
data: {"type":"start","session_id":"..."}
data: {"type":"text","content":"Alright, check this out..."}
data: {"type":"tool_call","name":"Read","input":{...}}
data: {"type":"tool_result","content":"..."}
data: {"type":"text","content":"I found 3 critical CVEs..."}
data: {"type":"done","usage":{"input":1234,"output":567}}
```

### WebSocket API

```javascript
// Connect
ws = new WebSocket('ws://localhost:3001/ws/agent?session=xxx&agent=cereal')

// Send message
ws.send(JSON.stringify({
  message: "Analyze the dependencies",
  project_id: "expressjs/express"
}))

// Receive events
ws.onmessage = (event) => {
  const data = JSON.parse(event.data)
  switch(data.type) {
    case 'text': // Streaming text
    case 'tool_call': // Agent is using a tool
    case 'tool_result': // Tool returned result
    case 'agent_switch': // Delegated to another agent
    case 'done': // Response complete
    case 'error': // Error occurred
  }
}
```

---

## Tool Schema (Claude Format)

```json
{
  "name": "Read",
  "description": "Read the contents of a file",
  "input_schema": {
    "type": "object",
    "properties": {
      "file_path": {
        "type": "string",
        "description": "Absolute path to the file"
      },
      "offset": {
        "type": "integer",
        "description": "Line number to start from (1-indexed)"
      },
      "limit": {
        "type": "integer",
        "description": "Maximum lines to read"
      }
    },
    "required": ["file_path"]
  }
}
```

```json
{
  "name": "Grep",
  "description": "Search for patterns in files",
  "input_schema": {
    "type": "object",
    "properties": {
      "pattern": {
        "type": "string",
        "description": "Regex pattern to search for"
      },
      "path": {
        "type": "string",
        "description": "Directory or file to search"
      },
      "glob": {
        "type": "string",
        "description": "File pattern filter (e.g., *.go)"
      }
    },
    "required": ["pattern"]
  }
}
```

```json
{
  "name": "GetAnalysis",
  "description": "Get scanner analysis results for a project",
  "input_schema": {
    "type": "object",
    "properties": {
      "project_id": {
        "type": "string",
        "description": "Project ID (owner/repo)"
      },
      "scanner": {
        "type": "string",
        "enum": ["code-packages", "code-security", "devops", "technology-identification"],
        "description": "Scanner type"
      },
      "section": {
        "type": "string",
        "description": "Specific section (e.g., 'summary', 'findings.vulnerabilities')"
      }
    },
    "required": ["project_id", "scanner"]
  }
}
```

---

## Security Considerations

### Tool Execution Boundaries

1. **File System Access**
   - Restrict to `$ZERO_HOME` and project directories
   - No access to system files, credentials
   - Read-only by default

2. **Command Execution**
   - Whitelist allowed commands
   - No shell expansion
   - Timeout enforcement
   - Sandboxed environment

3. **Web Access**
   - Rate limiting
   - Domain restrictions (optional)
   - No access to internal networks

### API Security

1. **Authentication** (future)
   - API key for HTTP endpoints
   - Token-based sessions

2. **Rate Limiting**
   - Per-session limits
   - Global rate limits

---

## Success Metrics

1. **CLI Experience**
   - Sub-second response time for first token
   - Smooth streaming output
   - Clean tool call visualization

2. **Agent Quality**
   - Agents use correct tools for tasks
   - Delegation works correctly
   - Responses match agent personas

3. **MCP Integration**
   - All tools execute via MCP
   - Tool results properly formatted
   - Error handling robust

---

## Next Steps

1. **Review this plan** - Get feedback on architecture
2. **Start Phase 1** - Build agent loader and session manager
3. **Prototype CLI** - Basic REPL with hardcoded tools
4. **Add MCP** - Integrate tool execution
5. **Polish** - UI, streaming, edge cases
