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

	"github.com/crashappsec/zero/pkg/core/credentials"
)

const (
	defaultClaudeAPIURL     = "https://api.anthropic.com/v1/messages"
	defaultClaudeModel      = "claude-sonnet-4-20250514"
	defaultClaudeAPIVersion = "2023-06-01"
	defaultMaxTokens        = 8192
	defaultTimeout          = 300 * time.Second
)

// LLMClient handles communication with Claude API
type LLMClient struct {
	apiKey     string
	apiURL     string
	model      string
	apiVersion string
	maxTokens  int
	httpClient *http.Client
}

// LLMClientOptions configures the LLM client
type LLMClientOptions struct {
	APIKey     string
	APIURL     string
	Model      string
	APIVersion string
	MaxTokens  int
	Timeout    time.Duration
}

// NewLLMClient creates a new LLM client
func NewLLMClient(opts *LLMClientOptions) *LLMClient {
	if opts == nil {
		opts = &LLMClientOptions{}
	}

	apiKey := opts.APIKey
	if apiKey == "" {
		// Use credentials package to get API key from env var or config file
		credInfo := credentials.GetAnthropicKey()
		if credInfo.Valid {
			apiKey = credInfo.Value
		}
	}

	apiURL := opts.APIURL
	if apiURL == "" {
		apiURL = defaultClaudeAPIURL
	}

	model := opts.Model
	if model == "" {
		model = os.Getenv("CLAUDE_MODEL")
		if model == "" {
			model = defaultClaudeModel
		}
	}

	apiVersion := opts.APIVersion
	if apiVersion == "" {
		apiVersion = defaultClaudeAPIVersion
	}

	maxTokens := opts.MaxTokens
	if maxTokens == 0 {
		maxTokens = defaultMaxTokens
	}

	timeout := opts.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}

	return &LLMClient{
		apiKey:     apiKey,
		apiURL:     apiURL,
		model:      model,
		apiVersion: apiVersion,
		maxTokens:  maxTokens,
		httpClient: &http.Client{Timeout: timeout},
	}
}

// IsAvailable checks if the API key is configured
func (c *LLMClient) IsAvailable() bool {
	return c.apiKey != ""
}

// ClaudeMessage represents a message in Claude API format
type ClaudeMessage struct {
	Role    string        `json:"role"`
	Content []ContentBlock `json:"content"`
}

// ContentBlock represents a content block in a message
type ContentBlock struct {
	Type string `json:"type"`

	// For text blocks
	Text string `json:"text,omitempty"`

	// For tool_use blocks
	ID    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`

	// For tool_result blocks
	ToolUseID string `json:"tool_use_id,omitempty"`
	Content   string `json:"content,omitempty"`
	IsError   bool   `json:"is_error,omitempty"`
}

// ToolDefinition defines a tool for Claude
type ToolDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"input_schema"`
}

// InputSchema defines the input schema for a tool
type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required,omitempty"`
}

// Property defines a property in the input schema
type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

// ClaudeRequest represents a request to Claude API
type ClaudeRequest struct {
	Model     string           `json:"model"`
	MaxTokens int              `json:"max_tokens"`
	System    string           `json:"system,omitempty"`
	Messages  []ClaudeMessage  `json:"messages"`
	Tools     []ToolDefinition `json:"tools,omitempty"`
	Stream    bool             `json:"stream,omitempty"`
}

// ClaudeResponse represents a response from Claude API
type ClaudeResponse struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []ContentBlock `json:"content"`
	Model        string         `json:"model"`
	StopReason   string         `json:"stop_reason"`
	StopSequence string         `json:"stop_sequence,omitempty"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// StreamEvent represents a streaming event from Claude
type StreamEvent struct {
	Type         string          `json:"type"`
	Index        int             `json:"index,omitempty"`
	ContentBlock *ContentBlock   `json:"content_block,omitempty"`
	Delta        *StreamDelta    `json:"delta,omitempty"`
	Message      *ClaudeResponse `json:"message,omitempty"`
	Usage        *struct {
		OutputTokens int `json:"output_tokens"`
	} `json:"usage,omitempty"`
}

// StreamDelta represents delta content in streaming
type StreamDelta struct {
	Type        string          `json:"type,omitempty"`
	Text        string          `json:"text,omitempty"`
	PartialJSON string          `json:"partial_json,omitempty"`
	StopReason  string          `json:"stop_reason,omitempty"`
}

// ChatEvent represents an event in the chat stream
type ChatEvent struct {
	Type string `json:"type"` // "text", "tool_call", "tool_result", "error", "done", "delegation"

	// For text events
	Text string `json:"text,omitempty"`

	// For tool_call events
	ToolCall *ToolCall `json:"tool_call,omitempty"`

	// For tool_result events
	ToolResult *ToolResult `json:"tool_result,omitempty"`

	// For error events
	Error string `json:"error,omitempty"`

	// For done events
	Usage *TokenUsage `json:"usage,omitempty"`

	// For delegation events (sub-agent progress)
	DelegatedAgent string `json:"delegated_agent,omitempty"` // Name of delegated agent
	DelegatedEvent string `json:"delegated_event,omitempty"` // Type of sub-event (tool_call, text, etc.)
}

// TokenUsage tracks token usage
type TokenUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// ToolExecutor is a function that executes a tool
type ToolExecutor func(ctx context.Context, call *ToolCall) (*ToolResult, error)

// StreamingToolExecutor is a function that executes a tool with event streaming support
// The callback parameter allows tools (like DelegateAgent) to stream progress events
type StreamingToolExecutor func(ctx context.Context, call *ToolCall, callback func(ChatEvent)) (*ToolResult, error)

// ChatWithTools sends a message and handles tool calls with streaming
func (c *LLMClient) ChatWithTools(
	ctx context.Context,
	systemPrompt string,
	messages []Message,
	tools []ToolDefinition,
	toolExecutor ToolExecutor,
	callback func(ChatEvent),
) error {
	if !c.IsAvailable() {
		return fmt.Errorf("ANTHROPIC_API_KEY not configured")
	}

	// Convert messages to Claude format
	claudeMessages := c.convertMessages(messages)

	// Main tool use loop
	for {
		// Send request to Claude
		response, err := c.sendStreamingRequest(ctx, systemPrompt, claudeMessages, tools, callback)
		if err != nil {
			callback(ChatEvent{Type: "error", Error: err.Error()})
			return err
		}

		// Check if we need to execute tools
		toolCalls := extractToolCalls(response)
		if len(toolCalls) == 0 {
			// No tool calls, we're done
			callback(ChatEvent{
				Type: "done",
				Usage: &TokenUsage{
					InputTokens:  response.Usage.InputTokens,
					OutputTokens: response.Usage.OutputTokens,
				},
			})
			return nil
		}

		// Add assistant response to messages
		claudeMessages = append(claudeMessages, ClaudeMessage{
			Role:    "assistant",
			Content: response.Content,
		})

		// Execute each tool call
		var toolResults []ContentBlock
		for _, call := range toolCalls {
			// Notify about tool call
			callback(ChatEvent{Type: "tool_call", ToolCall: &call})

			// Execute the tool
			result, err := toolExecutor(ctx, &call)
			if err != nil {
				result = &ToolResult{
					ToolCallID: call.ID,
					Content:    fmt.Sprintf("Error: %s", err.Error()),
					IsError:    true,
				}
			}

			// Notify about tool result
			callback(ChatEvent{Type: "tool_result", ToolResult: result})

			toolResults = append(toolResults, ContentBlock{
				Type:      "tool_result",
				ToolUseID: call.ID,
				Content:   result.Content,
				IsError:   result.IsError,
			})
		}

		// Add tool results to messages
		claudeMessages = append(claudeMessages, ClaudeMessage{
			Role:    "user",
			Content: toolResults,
		})

		// Continue the loop to get Claude's response to tool results
	}
}

// ChatWithStreamingTools sends a message and handles tool calls with streaming tool support
// This version allows tools to stream events (e.g., DelegateAgent streaming sub-agent progress)
func (c *LLMClient) ChatWithStreamingTools(
	ctx context.Context,
	systemPrompt string,
	messages []Message,
	tools []ToolDefinition,
	toolExecutor StreamingToolExecutor,
	callback func(ChatEvent),
) error {
	if !c.IsAvailable() {
		return fmt.Errorf("ANTHROPIC_API_KEY not configured")
	}

	// Convert messages to Claude format
	claudeMessages := c.convertMessages(messages)

	// Main tool use loop
	for {
		// Send request to Claude
		response, err := c.sendStreamingRequest(ctx, systemPrompt, claudeMessages, tools, callback)
		if err != nil {
			callback(ChatEvent{Type: "error", Error: err.Error()})
			return err
		}

		// Check if we need to execute tools
		toolCalls := extractToolCalls(response)
		if len(toolCalls) == 0 {
			// No tool calls, we're done
			callback(ChatEvent{
				Type: "done",
				Usage: &TokenUsage{
					InputTokens:  response.Usage.InputTokens,
					OutputTokens: response.Usage.OutputTokens,
				},
			})
			return nil
		}

		// Add assistant response to messages
		claudeMessages = append(claudeMessages, ClaudeMessage{
			Role:    "assistant",
			Content: response.Content,
		})

		// Execute each tool call
		var toolResults []ContentBlock
		for _, call := range toolCalls {
			// Notify about tool call
			callback(ChatEvent{Type: "tool_call", ToolCall: &call})

			// Execute the tool with callback for streaming support
			result, err := toolExecutor(ctx, &call, callback)
			if err != nil {
				result = &ToolResult{
					ToolCallID: call.ID,
					Content:    fmt.Sprintf("Error: %s", err.Error()),
					IsError:    true,
				}
			}

			// Notify about tool result
			callback(ChatEvent{Type: "tool_result", ToolResult: result})

			toolResults = append(toolResults, ContentBlock{
				Type:      "tool_result",
				ToolUseID: call.ID,
				Content:   result.Content,
				IsError:   result.IsError,
			})
		}

		// Add tool results to messages
		claudeMessages = append(claudeMessages, ClaudeMessage{
			Role:    "user",
			Content: toolResults,
		})

		// Continue the loop to get Claude's response to tool results
	}
}

// sendStreamingRequest sends a streaming request to Claude
func (c *LLMClient) sendStreamingRequest(
	ctx context.Context,
	systemPrompt string,
	messages []ClaudeMessage,
	tools []ToolDefinition,
	callback func(ChatEvent),
) (*ClaudeResponse, error) {
	req := ClaudeRequest{
		Model:     c.model,
		MaxTokens: c.maxTokens,
		System:    systemPrompt,
		Messages:  messages,
		Tools:     tools,
		Stream:    true,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", c.apiVersion)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse SSE stream
	return c.parseSSEStream(resp.Body, callback)
}

// parseSSEStream parses the SSE stream and returns the final response
func (c *LLMClient) parseSSEStream(body io.Reader, callback func(ChatEvent)) (*ClaudeResponse, error) {
	scanner := bufio.NewScanner(body)

	var response ClaudeResponse
	var currentContentBlocks []ContentBlock
	var currentText strings.Builder
	var currentToolInputs = make(map[int]strings.Builder) // index -> partial JSON

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		// Parse SSE data
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var event StreamEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue // Skip malformed events
		}

		switch event.Type {
		case "message_start":
			if event.Message != nil {
				response = *event.Message
			}

		case "content_block_start":
			if event.ContentBlock != nil {
				// Add new content block
				for len(currentContentBlocks) <= event.Index {
					currentContentBlocks = append(currentContentBlocks, ContentBlock{})
				}
				currentContentBlocks[event.Index] = *event.ContentBlock
			}

		case "content_block_delta":
			if event.Delta != nil {
				switch event.Delta.Type {
				case "text_delta":
					currentText.WriteString(event.Delta.Text)
					callback(ChatEvent{Type: "text", Text: event.Delta.Text})

				case "input_json_delta":
					if _, ok := currentToolInputs[event.Index]; !ok {
						currentToolInputs[event.Index] = strings.Builder{}
					}
					builder := currentToolInputs[event.Index]
					builder.WriteString(event.Delta.PartialJSON)
					currentToolInputs[event.Index] = builder
				}
			}

		case "content_block_stop":
			// Finalize the content block
			if event.Index < len(currentContentBlocks) {
				block := &currentContentBlocks[event.Index]
				if block.Type == "text" {
					block.Text = currentText.String()
				} else if block.Type == "tool_use" {
					if builder, ok := currentToolInputs[event.Index]; ok {
						inputStr := builder.String()
						if inputStr == "" {
							inputStr = "{}"
						}
						block.Input = json.RawMessage(inputStr)
					} else {
						// No input received, default to empty object
						block.Input = json.RawMessage("{}")
					}
				}
			}

		case "message_delta":
			if event.Delta != nil && event.Delta.StopReason != "" {
				response.StopReason = event.Delta.StopReason
			}
			if event.Usage != nil {
				response.Usage.OutputTokens = event.Usage.OutputTokens
			}

		case "message_stop":
			// Message complete
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading stream: %w", err)
	}

	// Build final response
	response.Content = currentContentBlocks

	return &response, nil
}

// convertMessages converts internal messages to Claude format
func (c *LLMClient) convertMessages(messages []Message) []ClaudeMessage {
	var claudeMessages []ClaudeMessage

	for _, msg := range messages {
		if msg.Role == RoleSystem {
			continue // System messages go in system prompt
		}

		var content []ContentBlock

		// Handle tool results
		if msg.ToolCallID != "" {
			content = append(content, ContentBlock{
				Type:      "tool_result",
				ToolUseID: msg.ToolCallID,
				Content:   msg.Content,
				IsError:   msg.IsError,
			})
		} else if len(msg.ToolCalls) > 0 {
			// Assistant message with tool calls
			if msg.Content != "" {
				content = append(content, ContentBlock{
					Type: "text",
					Text: msg.Content,
				})
			}
			for _, tc := range msg.ToolCalls {
				content = append(content, ContentBlock{
					Type:  "tool_use",
					ID:    tc.ID,
					Name:  tc.Name,
					Input: tc.Input,
				})
			}
		} else {
			// Regular text message
			content = append(content, ContentBlock{
				Type: "text",
				Text: msg.Content,
			})
		}

		claudeMessages = append(claudeMessages, ClaudeMessage{
			Role:    string(msg.Role),
			Content: content,
		})
	}

	return claudeMessages
}

// extractToolCalls extracts tool calls from a response
func extractToolCalls(response *ClaudeResponse) []ToolCall {
	var calls []ToolCall

	for _, block := range response.Content {
		if block.Type == "tool_use" {
			calls = append(calls, ToolCall{
				ID:    block.ID,
				Name:  block.Name,
				Input: block.Input,
			})
		}
	}

	return calls
}

// Chat sends a simple message without tools (non-streaming)
func (c *LLMClient) Chat(ctx context.Context, systemPrompt string, messages []Message) (string, error) {
	if !c.IsAvailable() {
		return "", fmt.Errorf("ANTHROPIC_API_KEY not configured")
	}

	claudeMessages := c.convertMessages(messages)

	req := ClaudeRequest{
		Model:     c.model,
		MaxTokens: c.maxTokens,
		System:    systemPrompt,
		Messages:  claudeMessages,
		Stream:    false,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.apiURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", c.apiVersion)

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

	// Extract text from response
	var text strings.Builder
	for _, block := range claudeResp.Content {
		if block.Type == "text" {
			text.WriteString(block.Text)
		}
	}

	return text.String(), nil
}
