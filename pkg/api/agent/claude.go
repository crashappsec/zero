package agent

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	claudeAPIURL     = "https://api.anthropic.com/v1/messages"
	claudeModel      = "claude-sonnet-4-20250514"
	claudeAPIVersion = "2023-06-01"
	maxTokens        = 4096
)

// ClaudeClient handles communication with the Claude API
type ClaudeClient struct {
	apiKey     string
	httpClient *http.Client
	zeroHome   string
}

// NewClaudeClient creates a new Claude API client
func NewClaudeClient(zeroHome string) *ClaudeClient {
	return &ClaudeClient{
		apiKey:     os.Getenv("ANTHROPIC_API_KEY"),
		httpClient: &http.Client{Timeout: 120 * time.Second},
		zeroHome:   zeroHome,
	}
}

// IsAvailable checks if the API key is configured
func (c *ClaudeClient) IsAvailable() bool {
	return c.apiKey != ""
}

// ClaudeMessage represents a message in the Claude API format
type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ClaudeRequest represents a request to the Claude API
type ClaudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	System    string          `json:"system,omitempty"`
	Messages  []ClaudeMessage `json:"messages"`
	Stream    bool            `json:"stream,omitempty"`
}

// ClaudeResponse represents a non-streaming response
type ClaudeResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// StreamEvent represents a streaming event from Claude
type StreamEvent struct {
	Type  string `json:"type"`
	Index int    `json:"index,omitempty"`
	Delta struct {
		Type string `json:"type,omitempty"`
		Text string `json:"text,omitempty"`
	} `json:"delta,omitempty"`
	Message *ClaudeResponse `json:"message,omitempty"`
}

// Chat sends a message and returns the response (non-streaming)
func (c *ClaudeClient) Chat(ctx context.Context, session *Session, userMessage string) (string, error) {
	if !c.IsAvailable() {
		return "", fmt.Errorf("ANTHROPIC_API_KEY not configured")
	}

	// Build messages from session history
	messages := c.buildMessages(session)
	messages = append(messages, ClaudeMessage{
		Role:    "user",
		Content: userMessage,
	})

	// Build system prompt
	systemPrompt := c.buildSystemPrompt(session)

	req := ClaudeRequest{
		Model:     claudeModel,
		MaxTokens: maxTokens,
		System:    systemPrompt,
		Messages:  messages,
		Stream:    false,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", claudeAPIURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", claudeAPIVersion)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var claudeResp ClaudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return "", fmt.Errorf("empty response from Claude")
	}

	return claudeResp.Content[0].Text, nil
}

// StreamCallback is called for each chunk of streamed content
type StreamCallback func(chunk StreamChunk)

// ChatStream sends a message and streams the response
func (c *ClaudeClient) ChatStream(ctx context.Context, session *Session, userMessage string, callback StreamCallback) error {
	if !c.IsAvailable() {
		return fmt.Errorf("ANTHROPIC_API_KEY not configured")
	}

	// Build messages from session history
	messages := c.buildMessages(session)
	messages = append(messages, ClaudeMessage{
		Role:    "user",
		Content: userMessage,
	})

	// Build system prompt
	systemPrompt := c.buildSystemPrompt(session)

	req := ClaudeRequest{
		Model:     claudeModel,
		MaxTokens: maxTokens,
		System:    systemPrompt,
		Messages:  messages,
		Stream:    true,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", claudeAPIURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", claudeAPIVersion)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Send start event
	callback(StreamChunk{
		Type:      "start",
		SessionID: session.ID,
		AgentID:   session.AgentID,
	})

	// Parse SSE stream
	scanner := bufio.NewScanner(resp.Body)
	var fullResponse strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		// Parse SSE data
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			// Skip [DONE] marker
			if data == "[DONE]" {
				break
			}

			var event StreamEvent
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				continue // Skip malformed events
			}

			switch event.Type {
			case "content_block_delta":
				if event.Delta.Text != "" {
					fullResponse.WriteString(event.Delta.Text)
					callback(StreamChunk{
						Type:      "delta",
						SessionID: session.ID,
						AgentID:   session.AgentID,
						Content:   event.Delta.Text,
					})
				}
			case "message_stop":
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading stream: %w", err)
	}

	// Send done event
	callback(StreamChunk{
		Type:      "done",
		SessionID: session.ID,
		AgentID:   session.AgentID,
		Content:   fullResponse.String(),
	})

	return nil
}

// buildMessages converts session messages to Claude format
func (c *ClaudeClient) buildMessages(session *Session) []ClaudeMessage {
	msgs := session.GetMessages()
	claudeMsgs := make([]ClaudeMessage, 0, len(msgs))

	for _, msg := range msgs {
		if msg.Role == RoleSystem {
			continue // System messages go in system prompt
		}
		claudeMsgs = append(claudeMsgs, ClaudeMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		})
	}

	return claudeMsgs
}

// buildSystemPrompt creates the system prompt for the agent
func (c *ClaudeClient) buildSystemPrompt(session *Session) string {
	agent := GetAgentInfo(session.AgentID)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("You are %s (%s), a specialist agent in the Zero security analysis system.\n\n", agent.Name, agent.Persona))
	sb.WriteString(fmt.Sprintf("Your expertise: %s\n\n", agent.Description))

	// Add project context if available
	if session.ProjectID != "" {
		sb.WriteString(fmt.Sprintf("Current project context: %s\n", session.ProjectID))
		sb.WriteString(fmt.Sprintf("Analysis data location: %s/repos/%s/analysis/\n\n", c.zeroHome, session.ProjectID))
	}

	// Add agent-specific guidance
	switch session.AgentID {
	case "zero":
		sb.WriteString(`You are the master orchestrator. You coordinate security analysis and can delegate to specialist agents when needed.

Available specialists you can recommend:
- Cereal: Supply chain security, dependency vulnerabilities, malcontent analysis
- Razor: Code security, SAST findings, secrets detection
- Gill: Cryptography analysis, cipher review, key management
- Plague: DevOps security, IaC, container scanning
- Blade: Compliance auditing, SOC 2, ISO 27001

When users ask about specific domains, suggest the appropriate specialist.
`)
	case "cereal":
		sb.WriteString(`Focus on:
- Analyzing dependency vulnerabilities (CVEs, severity, exploitability)
- Package health and maintenance status
- Malcontent findings (suspicious code patterns in dependencies)
- License compliance issues
- Typosquatting and dependency confusion risks
`)
	case "razor":
		sb.WriteString(`Focus on:
- Static analysis findings (code vulnerabilities)
- Secrets and credential detection
- Security anti-patterns in code
- Remediation recommendations
`)
	case "gill":
		sb.WriteString(`Focus on:
- Cryptographic implementations and weaknesses
- Cipher usage and recommendations
- Key management practices
- TLS/SSL configuration
- Random number generation security
`)
	}

	sb.WriteString("\nBe helpful, concise, and technically accurate. Use the Hackers (1995) persona appropriately but prioritize useful security insights.")

	return sb.String()
}
