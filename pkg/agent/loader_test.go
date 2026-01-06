package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAgentLoader_Load(t *testing.T) {
	// Find the agents directory relative to this test
	// In tests, we're in pkg/agent, so agents is at ../../agents
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	agentsDir := filepath.Join(wd, "..", "..", "agents")
	if _, err := os.Stat(agentsDir); os.IsNotExist(err) {
		t.Skipf("agents directory not found at %s, skipping test", agentsDir)
	}

	loader := NewAgentLoader(agentsDir)

	tests := []struct {
		agentID      string
		wantName     string
		wantPersona  string
		wantHasRole  bool
		wantHasVoice bool
	}{
		{
			agentID:      "zero",
			wantName:     "Zero",
			wantPersona:  "Zero Cool",
			wantHasRole:  true,
			wantHasVoice: true,
		},
		{
			agentID:      "cereal",
			wantName:     "Cereal",
			wantPersona:  "Cereal Killer",
			wantHasRole:  true,
			wantHasVoice: true,
		},
		{
			agentID:      "razor",
			wantName:     "Razor",
			wantPersona:  "Razor",
			wantHasRole:  true,
			wantHasVoice: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.agentID, func(t *testing.T) {
			agent, err := loader.Load(tt.agentID)
			if err != nil {
				t.Fatalf("Load(%s) error: %v", tt.agentID, err)
			}

			if agent.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", agent.Name, tt.wantName)
			}

			if agent.Persona != tt.wantPersona {
				t.Errorf("Persona = %q, want %q", agent.Persona, tt.wantPersona)
			}

			if tt.wantHasRole && agent.Role == "" {
				t.Error("Role is empty, expected content")
			}

			if tt.wantHasVoice && agent.VoiceFull == "" {
				t.Error("VoiceFull is empty, expected content")
			}

			// Verify caching works
			agent2, err := loader.Load(tt.agentID)
			if err != nil {
				t.Fatalf("Load(%s) second call error: %v", tt.agentID, err)
			}
			if agent != agent2 {
				t.Error("Expected cached agent to be returned")
			}
		})
	}
}

func TestAgentLoader_LoadUnknown(t *testing.T) {
	loader := NewAgentLoader("/tmp/nonexistent")

	_, err := loader.Load("unknown-agent")
	if err == nil {
		t.Error("Expected error for unknown agent")
	}
}

func TestAgentLoader_ParseMarkdown(t *testing.T) {
	content := `# Agent: Test Agent

## Identity

- **Name:** TestAgent
- **Domain:** Testing
- **Character Reference:** Test Character

## Role

You are a test agent for unit testing purposes.

## Capabilities

### Testing
- Run unit tests
- Verify behavior

### Validation
- Check inputs
- Validate outputs

## Process

1. Load test data
2. Execute tests
3. Report results

## Limitations

- Cannot access production systems
- Limited to test environments

## Autonomy

### Agent Delegation

| Scenario | Delegate To | Example |
|----------|-------------|---------|
| Security issues | **Razor** (Security) | "Check for vulns" |
| Legal questions | **Phreak** (Legal) | "License check" |

### Tools Available

| Tool | Purpose | When to Use |
|------|---------|-------------|
| **Read** | Read files | When examining code |
| **Grep** | Search | Finding patterns |

---

<!-- VOICE:full -->
## Voice & Personality

You are the test agent with full personality.

### Speech Patterns
- Use technical language
- Be precise

<!-- /VOICE:full -->

<!-- VOICE:minimal -->
## Communication Style

Professional and concise.
<!-- /VOICE:minimal -->

<!-- VOICE:neutral -->
## Communication Style

Objective technical output.
<!-- /VOICE:neutral -->
`

	loader := NewAgentLoader("/tmp")
	agent, err := loader.ParseMarkdown("test", content)
	if err != nil {
		t.Fatalf("ParseMarkdown error: %v", err)
	}

	// Check basic fields
	if agent.ID != "test" {
		t.Errorf("ID = %q, want %q", agent.ID, "test")
	}

	// Check role was extracted
	if agent.Role == "" {
		t.Error("Role is empty")
	}
	if !contains(agent.Role, "test agent") {
		t.Errorf("Role doesn't contain expected content: %s", agent.Role)
	}

	// Check capabilities
	if len(agent.Capabilities) != 2 {
		t.Errorf("Capabilities count = %d, want 2", len(agent.Capabilities))
	}

	// Check delegation rules
	if len(agent.Delegation) != 2 {
		t.Errorf("Delegation count = %d, want 2", len(agent.Delegation))
	} else {
		if agent.Delegation[0].AgentID != "razor" {
			t.Errorf("First delegation agent = %q, want %q", agent.Delegation[0].AgentID, "razor")
		}
	}

	// Check tools
	if len(agent.Tools) != 2 {
		t.Errorf("Tools count = %d, want 2", len(agent.Tools))
	}

	// Check voice modes
	if agent.VoiceFull == "" {
		t.Error("VoiceFull is empty")
	}
	if agent.VoiceMinimal == "" {
		t.Error("VoiceMinimal is empty")
	}
	if agent.VoiceNeutral == "" {
		t.Error("VoiceNeutral is empty")
	}
}

func TestAgentLoader_ListAvailable(t *testing.T) {
	loader := NewAgentLoader("/tmp")
	available := loader.ListAvailable()

	if len(available) == 0 {
		t.Error("Expected at least one available agent")
	}

	// Check that key agents are in the list
	foundZero := false
	foundCereal := false
	for _, id := range available {
		if id == "zero" {
			foundZero = true
		}
		if id == "cereal" {
			foundCereal = true
		}
	}

	if !foundZero {
		t.Error("Expected 'zero' in available agents")
	}
	if !foundCereal {
		t.Error("Expected 'cereal' in available agents")
	}
}

func TestAgentLoader_GetAgentInfo(t *testing.T) {
	loader := NewAgentLoader("/tmp")

	name, persona, character, ok := loader.GetAgentInfo("cereal")
	if !ok {
		t.Fatal("Expected ok=true for cereal")
	}
	if name != "Cereal" {
		t.Errorf("name = %q, want %q", name, "Cereal")
	}
	if persona != "Cereal Killer" {
		t.Errorf("persona = %q, want %q", persona, "Cereal Killer")
	}
	if character != "Emmanuel Goldstein" {
		t.Errorf("character = %q, want %q", character, "Emmanuel Goldstein")
	}

	_, _, _, ok = loader.GetAgentInfo("nonexistent")
	if ok {
		t.Error("Expected ok=false for nonexistent agent")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsLower(s, substr))
}

func containsLower(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return strings.Contains(s, substr)
}
