package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Role represents a message role in the conversation
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
)

// Message represents a chat message in the conversation
type Message struct {
	ID        string    `json:"id"`
	Role      Role      `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`

	// Tool-related fields (for assistant messages with tool use)
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`

	// For tool_result messages
	ToolCallID string `json:"tool_call_id,omitempty"`
	ToolName   string `json:"tool_name,omitempty"`
	IsError    bool   `json:"is_error,omitempty"`
}

// ToolCall represents a tool invocation by the assistant
type ToolCall struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	ToolCallID string `json:"tool_call_id"`
	Content    string `json:"content"`
	IsError    bool   `json:"is_error,omitempty"`
}

// Session represents a chat session with an agent
type Session struct {
	ID        string    `json:"id"`
	AgentID   string    `json:"agent_id"`
	ProjectID string    `json:"project_id,omitempty"`
	VoiceMode string    `json:"voice_mode"`
	Messages  []Message `json:"messages"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Metadata for tracking
	TokensUsed   int               `json:"tokens_used,omitempty"`
	ToolsUsed    []string          `json:"tools_used,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`

	mu sync.RWMutex
}

// NewSession creates a new chat session
func NewSession(agentID string) *Session {
	return &Session{
		ID:        uuid.New().String(),
		AgentID:   agentID,
		VoiceMode: "full",
		Messages:  []Message{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  make(map[string]string),
		ToolsUsed: []string{},
	}
}

// AddUserMessage adds a user message to the session
func (s *Session) AddUserMessage(content string) *Message {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg := Message{
		ID:        uuid.New().String(),
		Role:      RoleUser,
		Content:   content,
		Timestamp: time.Now(),
	}
	s.Messages = append(s.Messages, msg)
	s.UpdatedAt = time.Now()
	return &msg
}

// AddAssistantMessage adds an assistant message to the session
func (s *Session) AddAssistantMessage(content string, toolCalls []ToolCall) *Message {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg := Message{
		ID:        uuid.New().String(),
		Role:      RoleAssistant,
		Content:   content,
		ToolCalls: toolCalls,
		Timestamp: time.Now(),
	}
	s.Messages = append(s.Messages, msg)
	s.UpdatedAt = time.Now()
	return &msg
}

// AddToolResult adds a tool result message to the session
func (s *Session) AddToolResult(toolCallID, toolName, content string, isError bool) *Message {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg := Message{
		ID:         uuid.New().String(),
		Role:       RoleUser, // Tool results are sent as user messages in Claude API
		Content:    content,
		ToolCallID: toolCallID,
		ToolName:   toolName,
		IsError:    isError,
		Timestamp:  time.Now(),
	}
	s.Messages = append(s.Messages, msg)
	s.UpdatedAt = time.Now()

	// Track tool usage
	s.trackToolUsage(toolName)

	return &msg
}

// trackToolUsage adds a tool to the used tools list if not already present
func (s *Session) trackToolUsage(toolName string) {
	for _, t := range s.ToolsUsed {
		if t == toolName {
			return
		}
	}
	s.ToolsUsed = append(s.ToolsUsed, toolName)
}

// GetMessages returns a copy of all messages
func (s *Session) GetMessages() []Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	msgs := make([]Message, len(s.Messages))
	copy(msgs, s.Messages)
	return msgs
}

// GetLastMessage returns the last message in the session
func (s *Session) GetLastMessage() *Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.Messages) == 0 {
		return nil
	}
	msg := s.Messages[len(s.Messages)-1]
	return &msg
}

// SetProject sets the project context for the session
func (s *Session) SetProject(projectID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ProjectID = projectID
	s.UpdatedAt = time.Now()
}

// SetVoiceMode sets the voice mode for the session
func (s *Session) SetVoiceMode(mode string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.VoiceMode = mode
	s.UpdatedAt = time.Now()
}

// SwitchAgent changes the active agent for this session
func (s *Session) SwitchAgent(agentID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.AgentID = agentID
	s.UpdatedAt = time.Now()
}

// AddTokens adds to the token count
func (s *Session) AddTokens(count int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TokensUsed += count
	s.UpdatedAt = time.Now()
}

// SetMetadata sets a metadata key-value pair
func (s *Session) SetMetadata(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Metadata == nil {
		s.Metadata = make(map[string]string)
	}
	s.Metadata[key] = value
	s.UpdatedAt = time.Now()
}

// MessageCount returns the number of messages
func (s *Session) MessageCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.Messages)
}

// SessionManager manages chat sessions
type SessionManager struct {
	sessions    map[string]*Session
	persistDir  string // Optional directory for persistence
	mu          sync.RWMutex
}

// NewSessionManager creates a new session manager
func NewSessionManager(persistDir string) *SessionManager {
	return &SessionManager{
		sessions:   make(map[string]*Session),
		persistDir: persistDir,
	}
}

// Create creates a new session
func (m *SessionManager) Create(agentID string) *Session {
	session := NewSession(agentID)

	m.mu.Lock()
	m.sessions[session.ID] = session
	m.mu.Unlock()

	return session
}

// Get returns a session by ID
func (m *SessionManager) Get(id string) (*Session, bool) {
	m.mu.RLock()
	session, ok := m.sessions[id]
	m.mu.RUnlock()

	if ok {
		return session, true
	}

	// Try to load from disk if persistence is enabled
	if m.persistDir != "" {
		session, err := m.loadFromDisk(id)
		if err == nil {
			m.mu.Lock()
			m.sessions[id] = session
			m.mu.Unlock()
			return session, true
		}
	}

	return nil, false
}

// GetOrCreate gets an existing session or creates a new one
func (m *SessionManager) GetOrCreate(id, agentID string) *Session {
	if session, ok := m.Get(id); ok {
		return session
	}

	session := NewSession(agentID)
	session.ID = id // Use provided ID

	m.mu.Lock()
	m.sessions[id] = session
	m.mu.Unlock()

	return session
}

// Delete removes a session
func (m *SessionManager) Delete(id string) {
	m.mu.Lock()
	delete(m.sessions, id)
	m.mu.Unlock()

	// Delete from disk if persistence is enabled
	if m.persistDir != "" {
		os.Remove(filepath.Join(m.persistDir, id+".json"))
	}
}

// List returns all sessions
func (m *SessionManager) List() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessions := make([]*Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		sessions = append(sessions, s)
	}
	return sessions
}

// Cleanup removes sessions older than maxAge
func (m *SessionManager) Cleanup(maxAge time.Duration) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	removed := 0

	for id, session := range m.sessions {
		if session.UpdatedAt.Before(cutoff) {
			delete(m.sessions, id)
			if m.persistDir != "" {
				os.Remove(filepath.Join(m.persistDir, id+".json"))
			}
			removed++
		}
	}

	return removed
}

// Save persists a session to disk
func (m *SessionManager) Save(session *Session) error {
	if m.persistDir == "" {
		return nil // Persistence not enabled
	}

	if err := os.MkdirAll(m.persistDir, 0755); err != nil {
		return fmt.Errorf("creating persist dir: %w", err)
	}

	session.mu.RLock()
	data, err := json.MarshalIndent(session, "", "  ")
	session.mu.RUnlock()

	if err != nil {
		return fmt.Errorf("marshaling session: %w", err)
	}

	path := filepath.Join(m.persistDir, session.ID+".json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing session file: %w", err)
	}

	return nil
}

// loadFromDisk loads a session from disk
func (m *SessionManager) loadFromDisk(id string) (*Session, error) {
	if m.persistDir == "" {
		return nil, fmt.Errorf("persistence not enabled")
	}

	path := filepath.Join(m.persistDir, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading session file: %w", err)
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("unmarshaling session: %w", err)
	}

	return &session, nil
}

// SaveAll persists all sessions to disk
func (m *SessionManager) SaveAll() error {
	m.mu.RLock()
	sessions := make([]*Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		sessions = append(sessions, s)
	}
	m.mu.RUnlock()

	for _, session := range sessions {
		if err := m.Save(session); err != nil {
			return err
		}
	}

	return nil
}

// Count returns the number of active sessions
func (m *SessionManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}
