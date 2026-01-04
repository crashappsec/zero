// Package agent provides agent chat functionality for the Zero API
package agent

import (
	"sync"
	"time"
)

// Role represents a message role
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
)

// Message represents a chat message
type Message struct {
	Role      Role      `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// Session represents a chat session with an agent
type Session struct {
	ID        string    `json:"id"`
	AgentID   string    `json:"agent_id"`   // e.g., "zero", "cereal", "razor"
	ProjectID string    `json:"project_id"` // optional - current project context
	Messages  []Message `json:"messages"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	mu        sync.RWMutex
}

// NewSession creates a new chat session
func NewSession(id, agentID string) *Session {
	return &Session{
		ID:        id,
		AgentID:   agentID,
		Messages:  []Message{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// AddMessage adds a message to the session
func (s *Session) AddMessage(role Role, content string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Messages = append(s.Messages, Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
	s.UpdatedAt = time.Now()
}

// GetMessages returns a copy of all messages
func (s *Session) GetMessages() []Message {
	s.mu.RLock()
	defer s.mu.RUnlock()
	msgs := make([]Message, len(s.Messages))
	copy(msgs, s.Messages)
	return msgs
}

// SetProject sets the project context for the session
func (s *Session) SetProject(projectID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ProjectID = projectID
	s.UpdatedAt = time.Now()
}

// SessionManager manages chat sessions
type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewSessionManager creates a new session manager
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

// Create creates a new session
func (m *SessionManager) Create(id, agentID string) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()
	session := NewSession(id, agentID)
	m.sessions[id] = session
	return session
}

// Get returns a session by ID
func (m *SessionManager) Get(id string) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	session, ok := m.sessions[id]
	return session, ok
}

// GetOrCreate gets an existing session or creates a new one
func (m *SessionManager) GetOrCreate(id, agentID string) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()
	if session, ok := m.sessions[id]; ok {
		return session
	}
	session := NewSession(id, agentID)
	m.sessions[id] = session
	return session
}

// Delete removes a session
func (m *SessionManager) Delete(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, id)
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
			removed++
		}
	}
	return removed
}

// AgentInfo contains agent metadata for prompts
type AgentInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Persona     string `json:"persona"`
	Description string `json:"description"`
	Scanner     string `json:"scanner"`
}

// GetAgentInfo returns info for a given agent ID
func GetAgentInfo(agentID string) AgentInfo {
	agents := map[string]AgentInfo{
		"zero":   {ID: "zero", Name: "Zero", Persona: "Zero Cool", Description: "Master orchestrator", Scanner: "all"},
		"cereal": {ID: "cereal", Name: "Cereal", Persona: "Cereal Killer", Description: "Supply chain security", Scanner: "code-packages"},
		"razor":  {ID: "razor", Name: "Razor", Persona: "Razor", Description: "Code security, SAST, secrets", Scanner: "code-security"},
		"blade":  {ID: "blade", Name: "Blade", Persona: "Blade", Description: "Compliance, SOC 2, ISO 27001", Scanner: "multiple"},
		"phreak": {ID: "phreak", Name: "Phreak", Persona: "Phantom Phreak", Description: "Legal, licenses, privacy", Scanner: "code-packages"},
		"acid":   {ID: "acid", Name: "Acid", Persona: "Acid Burn", Description: "Frontend, React, TypeScript", Scanner: "code-security"},
		"dade":   {ID: "dade", Name: "Dade", Persona: "Dade Murphy", Description: "Backend, APIs, databases", Scanner: "code-security"},
		"nikon":  {ID: "nikon", Name: "Nikon", Persona: "Lord Nikon", Description: "Architecture, system design", Scanner: "technology-identification"},
		"joey":   {ID: "joey", Name: "Joey", Persona: "Joey", Description: "CI/CD, build optimization", Scanner: "devops"},
		"plague": {ID: "plague", Name: "Plague", Persona: "The Plague", Description: "DevOps, IaC, Kubernetes", Scanner: "devops"},
		"gibson": {ID: "gibson", Name: "Gibson", Persona: "The Gibson", Description: "DORA metrics, team health", Scanner: "devops"},
		"gill":   {ID: "gill", Name: "Gill", Persona: "Gill Bates", Description: "Cryptography specialist", Scanner: "code-security"},
		"turing": {ID: "turing", Name: "Turing", Persona: "Alan Turing", Description: "AI/ML security", Scanner: "technology-identification"},
	}

	if info, ok := agents[agentID]; ok {
		return info
	}
	return agents["zero"] // Default to Zero
}

// ChatRequest represents an incoming chat request
type ChatRequest struct {
	SessionID string `json:"session_id,omitempty"` // Optional - creates new if empty
	AgentID   string `json:"agent_id,omitempty"`   // Agent to chat with (default: zero)
	ProjectID string `json:"project_id,omitempty"` // Optional project context
	Message   string `json:"message"`              // User message
}

// ChatResponse represents a chat response
type ChatResponse struct {
	SessionID string `json:"session_id"`
	AgentID   string `json:"agent_id"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
}

// StreamChunk represents a streaming response chunk
type StreamChunk struct {
	Type      string `json:"type"`                 // "start", "delta", "done", "error"
	SessionID string `json:"session_id"`
	AgentID   string `json:"agent_id"`
	Content   string `json:"content,omitempty"`
	Error     string `json:"error,omitempty"`
}
