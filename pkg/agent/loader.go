// Package agent provides the agent runtime for Zero
package agent

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// AgentDefinition represents a fully loaded agent from agent.md
type AgentDefinition struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Persona      string            `json:"persona"`
	Domain       string            `json:"domain"`
	Character    string            `json:"character"`
	Role         string            `json:"role"`
	Capabilities []string          `json:"capabilities"`
	Process      string            `json:"process"`
	Autonomy     string            `json:"autonomy"`
	VoiceFull    string            `json:"voice_full"`
	VoiceMinimal string            `json:"voice_minimal"`
	VoiceNeutral string            `json:"voice_neutral"`
	DataSources  []string          `json:"data_sources"`
	Delegation   []DelegationRule  `json:"delegation"`
	Tools        []string          `json:"tools"`
	Limitations  string            `json:"limitations"`
	RawContent   string            `json:"-"` // Full markdown for reference
	Metadata     map[string]string `json:"metadata"`
}

// DelegationRule defines when an agent can delegate to another
type DelegationRule struct {
	Scenario   string `json:"scenario"`
	DelegateTo string `json:"delegate_to"`
	AgentID    string `json:"agent_id"`
	Example    string `json:"example"`
}

// AgentLoader loads agent definitions from the agents/ directory
type AgentLoader struct {
	agentsDir string
	cache     map[string]*AgentDefinition
	mu        sync.RWMutex
}

// NewAgentLoader creates a new agent loader
func NewAgentLoader(agentsDir string) *AgentLoader {
	return &AgentLoader{
		agentsDir: agentsDir,
		cache:     make(map[string]*AgentDefinition),
	}
}

// agentDirMapping maps agent IDs to their directory names
var agentDirMapping = map[string]string{
	"zero":   "orchestrator",
	"cereal": "supply-chain",
	"razor":  "code-security",
	"blade":  "compliance",
	"phreak": "legal",
	"acid":   "frontend",
	"dade":   "backend",
	"nikon":  "architecture",
	"joey":   "build",
	"plague": "devops",
	"gibson": "engineering-leader",
	"gill":   "cryptography",
	"turing": "ai-security",
}

// agentMetadata provides static metadata for agents
var agentMetadata = map[string]struct {
	Name      string
	Persona   string
	Character string
}{
	"zero":   {"Zero", "Zero Cool", "Dade Murphy"},
	"cereal": {"Cereal", "Cereal Killer", "Emmanuel Goldstein"},
	"razor":  {"Razor", "Razor", "Razor"},
	"blade":  {"Blade", "Blade", "Blade"},
	"phreak": {"Phreak", "Phantom Phreak", "Ramon Sanchez"},
	"acid":   {"Acid", "Acid Burn", "Kate Libby"},
	"dade":   {"Dade", "Dade Murphy", "Zero Cool"},
	"nikon":  {"Nikon", "Lord Nikon", "Paul Cook"},
	"joey":   {"Joey", "Joey", "Joey Pardella"},
	"plague": {"Plague", "The Plague", "Eugene Belford"},
	"gibson": {"Gibson", "The Gibson", "The Gibson"},
	"gill":   {"Gill", "Gill Bates", "Gill Bates"},
	"turing": {"Turing", "Alan Turing", "Alan Turing"},
}

// Load loads an agent definition by ID
func (l *AgentLoader) Load(agentID string) (*AgentDefinition, error) {
	// Check cache first
	l.mu.RLock()
	if agent, ok := l.cache[agentID]; ok {
		l.mu.RUnlock()
		return agent, nil
	}
	l.mu.RUnlock()

	// Find the agent directory
	dirName, ok := agentDirMapping[agentID]
	if !ok {
		return nil, fmt.Errorf("unknown agent: %s", agentID)
	}

	// Load the agent.md file
	agentPath := filepath.Join(l.agentsDir, dirName, "agent.md")
	content, err := os.ReadFile(agentPath)
	if err != nil {
		return nil, fmt.Errorf("reading agent file %s: %w", agentPath, err)
	}

	// Parse the markdown
	agent, err := l.ParseMarkdown(agentID, string(content))
	if err != nil {
		return nil, fmt.Errorf("parsing agent %s: %w", agentID, err)
	}

	// Cache the result
	l.mu.Lock()
	l.cache[agentID] = agent
	l.mu.Unlock()

	return agent, nil
}

// LoadAll loads all available agents
func (l *AgentLoader) LoadAll() ([]*AgentDefinition, error) {
	agents := make([]*AgentDefinition, 0, len(agentDirMapping))

	for agentID := range agentDirMapping {
		agent, err := l.Load(agentID)
		if err != nil {
			// Log but don't fail - some agents might not have definitions yet
			continue
		}
		agents = append(agents, agent)
	}

	return agents, nil
}

// ListAvailable returns a list of available agent IDs
func (l *AgentLoader) ListAvailable() []string {
	ids := make([]string, 0, len(agentDirMapping))
	for id := range agentDirMapping {
		ids = append(ids, id)
	}
	return ids
}

// GetAgentInfo returns basic info without loading full definition
func (l *AgentLoader) GetAgentInfo(agentID string) (name, persona, character string, ok bool) {
	if meta, exists := agentMetadata[agentID]; exists {
		return meta.Name, meta.Persona, meta.Character, true
	}
	return "", "", "", false
}

// ParseMarkdown parses an agent.md file into AgentDefinition
func (l *AgentLoader) ParseMarkdown(agentID, content string) (*AgentDefinition, error) {
	agent := &AgentDefinition{
		ID:          agentID,
		RawContent:  content,
		Metadata:    make(map[string]string),
		DataSources: []string{},
		Delegation:  []DelegationRule{},
		Tools:       []string{},
	}

	// Set metadata from static mapping
	if meta, ok := agentMetadata[agentID]; ok {
		agent.Name = meta.Name
		agent.Persona = meta.Persona
		agent.Character = meta.Character
	}

	// Parse identity section for domain
	agent.Domain = extractSection(content, "## Identity", "## ")
	if agent.Domain != "" {
		agent.Domain = extractBulletValue(agent.Domain, "Domain")
	}

	// Parse role section
	agent.Role = extractSection(content, "## Role", "## ")

	// Parse capabilities
	capSection := extractSection(content, "## Capabilities", "## ")
	agent.Capabilities = extractCapabilities(capSection)

	// Parse process
	agent.Process = extractSection(content, "## Process", "## ")

	// Parse autonomy
	agent.Autonomy = extractSection(content, "## Autonomy", "## ")

	// Parse limitations
	agent.Limitations = extractSection(content, "## Limitations", "## ")

	// Parse data sources
	dataSection := extractSection(content, "## Data Sources", "## ")
	agent.DataSources = extractDataSources(dataSection)

	// Parse delegation rules
	agent.Delegation = extractDelegationRules(content)

	// Parse tools
	agent.Tools = extractTools(content)

	// Parse voice modes from HTML comments
	agent.VoiceFull = extractVoiceMode(content, "full")
	agent.VoiceMinimal = extractVoiceMode(content, "minimal")
	agent.VoiceNeutral = extractVoiceMode(content, "neutral")

	return agent, nil
}

// ClearCache clears the agent cache
func (l *AgentLoader) ClearCache() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cache = make(map[string]*AgentDefinition)
}

// extractSection extracts content between a start header and the next header
func extractSection(content, startHeader, endPattern string) string {
	startIdx := strings.Index(content, startHeader)
	if startIdx == -1 {
		return ""
	}

	// Move past the header line
	startIdx += len(startHeader)
	for startIdx < len(content) && content[startIdx] != '\n' {
		startIdx++
	}
	if startIdx < len(content) {
		startIdx++ // Skip the newline
	}

	// Find the next section header
	remaining := content[startIdx:]
	endIdx := -1

	scanner := bufio.NewScanner(strings.NewReader(remaining))
	pos := 0
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, endPattern) && !strings.HasPrefix(line, startHeader) {
			endIdx = pos
			break
		}
		pos += len(line) + 1 // +1 for newline
	}

	if endIdx == -1 {
		// Check for HTML comment end (voice sections)
		if idx := strings.Index(remaining, "<!--"); idx != -1 {
			endIdx = idx
		} else {
			return strings.TrimSpace(remaining)
		}
	}

	return strings.TrimSpace(remaining[:endIdx])
}

// extractVoiceMode extracts content from <!-- VOICE:mode --> comments
func extractVoiceMode(content, mode string) string {
	startTag := fmt.Sprintf("<!-- VOICE:%s -->", mode)
	endTag := fmt.Sprintf("<!-- /VOICE:%s -->", mode)

	startIdx := strings.Index(content, startTag)
	if startIdx == -1 {
		return ""
	}
	startIdx += len(startTag)

	endIdx := strings.Index(content[startIdx:], endTag)
	if endIdx == -1 {
		return ""
	}

	return strings.TrimSpace(content[startIdx : startIdx+endIdx])
}

// extractBulletValue extracts a value from a bullet point like "- **Key:** Value"
func extractBulletValue(content, key string) string {
	pattern := fmt.Sprintf(`-\s*\*\*%s:\*\*\s*(.+)`, regexp.QuoteMeta(key))
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// extractCapabilities extracts capability subsections
func extractCapabilities(content string) []string {
	var capabilities []string
	re := regexp.MustCompile(`###\s+(.+)`)
	matches := re.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			capabilities = append(capabilities, strings.TrimSpace(match[1]))
		}
	}
	return capabilities
}

// extractDataSources extracts data source file references
func extractDataSources(content string) []string {
	var sources []string
	// Match patterns like `packages.json`, `code-security.json`
	re := regexp.MustCompile("`([a-z-]+\\.json)`")
	matches := re.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			sources = append(sources, match[1])
		}
	}
	return sources
}

// extractDelegationRules extracts delegation rules from the Autonomy section
func extractDelegationRules(content string) []DelegationRule {
	var rules []DelegationRule

	// Find the delegation table
	tableStart := strings.Index(content, "| Scenario |")
	if tableStart == -1 {
		return rules
	}

	// Parse table rows
	scanner := bufio.NewScanner(strings.NewReader(content[tableStart:]))
	lineNum := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		// Skip header and separator rows
		if lineNum <= 2 {
			continue
		}

		// Stop at end of table
		if !strings.HasPrefix(line, "|") {
			break
		}

		// Parse table row: | Scenario | Delegate To | Example |
		parts := strings.Split(line, "|")
		if len(parts) >= 4 {
			scenario := strings.TrimSpace(parts[1])
			delegateTo := strings.TrimSpace(parts[2])
			example := strings.TrimSpace(parts[3])

			// Extract agent ID from delegate (e.g., "**Phreak** (Legal)" -> "phreak")
			agentID := extractAgentIDFromDelegate(delegateTo)

			if scenario != "" && agentID != "" {
				rules = append(rules, DelegationRule{
					Scenario:   scenario,
					DelegateTo: delegateTo,
					AgentID:    agentID,
					Example:    example,
				})
			}
		}
	}

	return rules
}

// extractAgentIDFromDelegate extracts agent ID from delegate string
func extractAgentIDFromDelegate(delegate string) string {
	// Match **AgentName** pattern
	re := regexp.MustCompile(`\*\*(\w+)\*\*`)
	matches := re.FindStringSubmatch(delegate)
	if len(matches) > 1 {
		name := strings.ToLower(matches[1])
		// Map name to ID
		for id, meta := range agentMetadata {
			if strings.ToLower(meta.Name) == name {
				return id
			}
		}
	}
	return ""
}

// extractTools extracts available tools from the Autonomy section
func extractTools(content string) []string {
	var tools []string

	// Find the tools table
	tableStart := strings.Index(content, "| Tool |")
	if tableStart == -1 {
		return tools
	}

	// Parse table rows
	scanner := bufio.NewScanner(strings.NewReader(content[tableStart:]))
	lineNum := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		// Skip header and separator rows
		if lineNum <= 2 {
			continue
		}

		// Stop at end of table
		if !strings.HasPrefix(line, "|") {
			break
		}

		// Parse table row: | **Tool** | Purpose | When to Use |
		parts := strings.Split(line, "|")
		if len(parts) >= 2 {
			tool := strings.TrimSpace(parts[1])
			// Remove ** markers
			tool = strings.ReplaceAll(tool, "**", "")
			if tool != "" && tool != "Tool" {
				tools = append(tools, tool)
			}
		}
	}

	return tools
}
