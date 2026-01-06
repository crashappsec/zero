package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/crashappsec/zero/pkg/agent"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// WebSocket configuration
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 8192
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in dev; restrict in production
	},
}

// Handler manages agent chat functionality using the new runtime
type Handler struct {
	runtime  *agent.Runtime
	sessions *SessionManager
}

// NewHandler creates a new agent handler with the runtime
func NewHandler(zeroHome string) *Handler {
	runtime, err := agent.NewRuntime(&agent.RuntimeOptions{
		ZeroHome: zeroHome,
	})
	if err != nil {
		log.Printf("Warning: Failed to create agent runtime: %v", err)
		// Still create handler - will return errors when used
	}

	return &Handler{
		runtime:  runtime,
		sessions: NewSessionManager(),
	}
}

// HandleWebSocket handles WebSocket connections for agent chat
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get session ID from query param or generate new one
	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	// Get agent ID (default to "zero")
	agentID := r.URL.Query().Get("agent")
	if agentID == "" {
		agentID = "zero"
	}

	// Get voice mode
	voiceMode := r.URL.Query().Get("voice")
	if voiceMode == "" {
		voiceMode = "full"
	}

	// Get or create session
	session := h.sessions.GetOrCreate(sessionID, agentID)

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Get agent info
	agentName, _, _, _ := h.runtime.GetAgentInfo(agentID)
	if agentName == "" {
		agentName = agentID
	}

	// Create client
	client := &wsClient{
		handler:   h,
		session:   session,
		conn:      conn,
		send:      make(chan []byte, 256),
		voiceMode: voiceMode,
	}

	// Send connection confirmation
	client.sendJSON(map[string]interface{}{
		"type":       "connected",
		"session_id": session.ID,
		"agent_id":   session.AgentID,
		"agent_name": agentName,
	})

	// Start read/write pumps
	var wg sync.WaitGroup
	wg.Add(2)

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	go func() {
		defer wg.Done()
		client.writePump(ctx)
	}()

	go func() {
		defer wg.Done()
		client.readPump(ctx, cancel)
	}()

	wg.Wait()
}

// HandleChat handles HTTP POST requests for chat (non-WebSocket)
func (h *Handler) HandleChat(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request", err)
		return
	}

	if req.Message == "" {
		writeError(w, http.StatusBadRequest, "message is required", nil)
		return
	}

	// Get or create session
	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	agentID := req.AgentID
	if agentID == "" {
		agentID = "zero"
	}

	session := h.sessions.GetOrCreate(sessionID, agentID)

	// Set project context if provided
	if req.ProjectID != "" {
		session.SetProject(req.ProjectID)
	}

	// Check if runtime is available
	if h.runtime == nil || !h.runtime.IsAvailable() {
		writeError(w, http.StatusServiceUnavailable, "ANTHROPIC_API_KEY not configured", nil)
		return
	}

	// Add user message to session
	session.AddMessage(RoleUser, req.Message)

	// Get response from runtime with tool use
	var fullResponse string
	var toolCalls []map[string]interface{}

	chatReq := &agent.ChatRequest{
		AgentID:   agentID,
		ProjectID: session.ProjectID,
		VoiceMode: req.VoiceMode,
		Message:   req.Message,
	}

	err := h.runtime.Chat(r.Context(), chatReq, func(event agent.ChatEvent) {
		switch event.Type {
		case "text":
			fullResponse += event.Text
		case "tool_call":
			toolCalls = append(toolCalls, map[string]interface{}{
				"name":  event.ToolCall.Name,
				"input": json.RawMessage(event.ToolCall.Input),
			})
		}
	})

	if err != nil {
		writeError(w, http.StatusInternalServerError, "chat failed", err)
		return
	}

	// Add assistant response to session
	session.AddMessage(RoleAssistant, fullResponse)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"session_id": sessionID,
		"agent_id":   agentID,
		"response":   fullResponse,
		"tool_calls": toolCalls,
		"done":       true,
	})
}

// HandleChatStream handles SSE streaming chat with tool use
func (h *Handler) HandleChatStream(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request", err)
		return
	}

	if req.Message == "" {
		writeError(w, http.StatusBadRequest, "message is required", nil)
		return
	}

	// Get or create session
	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	agentID := req.AgentID
	if agentID == "" {
		agentID = "zero"
	}

	session := h.sessions.GetOrCreate(sessionID, agentID)

	if req.ProjectID != "" {
		session.SetProject(req.ProjectID)
	}

	if h.runtime == nil || !h.runtime.IsAvailable() {
		writeError(w, http.StatusServiceUnavailable, "ANTHROPIC_API_KEY not configured", nil)
		return
	}

	// Check Flusher support before setting SSE headers
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming not supported", nil)
		return
	}

	// Set up SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Add user message
	session.AddMessage(RoleUser, req.Message)

	// Send start event
	sendSSE(w, flusher, StreamChunk{
		Type:      "start",
		SessionID: sessionID,
		AgentID:   agentID,
	})

	// Stream response with tool use events
	var fullResponse string
	chatReq := &agent.ChatRequest{
		AgentID:   agentID,
		ProjectID: session.ProjectID,
		VoiceMode: req.VoiceMode,
		Message:   req.Message,
	}

	err := h.runtime.Chat(r.Context(), chatReq, func(event agent.ChatEvent) {
		switch event.Type {
		case "text":
			fullResponse += event.Text
			sendSSE(w, flusher, StreamChunk{
				Type:      "delta",
				SessionID: sessionID,
				AgentID:   agentID,
				Content:   event.Text,
			})

		case "tool_call":
			sendSSE(w, flusher, map[string]interface{}{
				"type":       "tool_call",
				"session_id": sessionID,
				"agent_id":   agentID,
				"tool_name":  event.ToolCall.Name,
				"tool_input": json.RawMessage(event.ToolCall.Input),
			})

		case "tool_result":
			sendSSE(w, flusher, map[string]interface{}{
				"type":       "tool_result",
				"session_id": sessionID,
				"agent_id":   agentID,
				"is_error":   event.ToolResult.IsError,
			})

		case "error":
			sendSSE(w, flusher, StreamChunk{
				Type:      "error",
				SessionID: sessionID,
				AgentID:   agentID,
				Error:     event.Error,
			})

		case "done":
			// Token usage could be included here
		}
	})

	if err != nil {
		sendSSE(w, flusher, StreamChunk{
			Type:      "error",
			SessionID: sessionID,
			AgentID:   agentID,
			Error:     err.Error(),
		})
		return
	}

	// Add assistant response to session
	if fullResponse != "" {
		session.AddMessage(RoleAssistant, fullResponse)
	}

	// Send done event
	sendSSE(w, flusher, StreamChunk{
		Type:      "done",
		SessionID: sessionID,
		AgentID:   agentID,
		Content:   fullResponse,
	})
}

// HandleGetSession returns session details
func (h *Handler) HandleGetSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	session, ok := h.sessions.Get(sessionID)
	if !ok {
		writeError(w, http.StatusNotFound, "session not found", nil)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"session_id": session.ID,
		"agent_id":   session.AgentID,
		"project_id": session.ProjectID,
		"messages":   session.GetMessages(),
		"created_at": session.CreatedAt,
		"updated_at": session.UpdatedAt,
	})
}

// HandleListSessions returns all active sessions
func (h *Handler) HandleListSessions(w http.ResponseWriter, r *http.Request) {
	sessions := h.sessions.List()
	items := make([]map[string]interface{}, len(sessions))

	for i, s := range sessions {
		items[i] = map[string]interface{}{
			"session_id":    s.ID,
			"agent_id":      s.AgentID,
			"project_id":    s.ProjectID,
			"message_count": len(s.Messages),
			"created_at":    s.CreatedAt,
			"updated_at":    s.UpdatedAt,
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":  items,
		"total": len(items),
	})
}

// HandleDeleteSession deletes a session
func (h *Handler) HandleDeleteSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	h.sessions.Delete(sessionID)
	w.WriteHeader(http.StatusNoContent)
}

// HandleListAgents returns available agents
func (h *Handler) HandleListAgents(w http.ResponseWriter, r *http.Request) {
	agents := []map[string]interface{}{
		{"id": "zero", "name": "Zero", "persona": "Zero Cool", "description": "Master orchestrator"},
		{"id": "cereal", "name": "Cereal", "persona": "Cereal Killer", "description": "Supply chain security"},
		{"id": "razor", "name": "Razor", "persona": "Razor", "description": "Code security, SAST, secrets"},
		{"id": "blade", "name": "Blade", "persona": "Blade", "description": "Compliance, SOC 2, ISO 27001"},
		{"id": "phreak", "name": "Phreak", "persona": "Phantom Phreak", "description": "Legal, licenses, privacy"},
		{"id": "acid", "name": "Acid", "persona": "Acid Burn", "description": "Frontend, React, TypeScript"},
		{"id": "dade", "name": "Dade", "persona": "Dade Murphy", "description": "Backend, APIs, databases"},
		{"id": "nikon", "name": "Nikon", "persona": "Lord Nikon", "description": "Architecture, system design"},
		{"id": "joey", "name": "Joey", "persona": "Joey", "description": "CI/CD, build optimization"},
		{"id": "plague", "name": "Plague", "persona": "The Plague", "description": "DevOps, IaC, Kubernetes"},
		{"id": "gibson", "name": "Gibson", "persona": "The Gibson", "description": "DORA metrics, team health"},
		{"id": "gill", "name": "Gill", "persona": "Gill Bates", "description": "Cryptography specialist"},
		{"id": "turing", "name": "Turing", "persona": "Alan Turing", "description": "AI/ML security"},
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":  agents,
		"total": len(agents),
	})
}

// wsClient represents a WebSocket client for agent chat
type wsClient struct {
	handler   *Handler
	session   *Session
	conn      *websocket.Conn
	send      chan []byte
	voiceMode string
}

func (c *wsClient) sendJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	select {
	case c.send <- data:
		return nil
	default:
		return fmt.Errorf("send buffer full")
	}
}

func (c *wsClient) readPump(ctx context.Context, cancel context.CancelFunc) {
	defer cancel()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			return
		}

		// Parse message
		var req ChatRequest
		if err := json.Unmarshal(message, &req); err != nil {
			c.sendJSON(StreamChunk{
				Type:      "error",
				SessionID: c.session.ID,
				AgentID:   c.session.AgentID,
				Error:     "invalid message format",
			})
			continue
		}

		// Handle message
		c.handleMessage(ctx, req)
	}
}

func (c *wsClient) handleMessage(ctx context.Context, req ChatRequest) {
	if req.Message == "" {
		c.sendJSON(StreamChunk{
			Type:      "error",
			SessionID: c.session.ID,
			AgentID:   c.session.AgentID,
			Error:     "message is required",
		})
		return
	}

	// Update project context if provided
	if req.ProjectID != "" {
		c.session.SetProject(req.ProjectID)
	}

	// Check runtime availability
	if c.handler.runtime == nil || !c.handler.runtime.IsAvailable() {
		c.sendJSON(StreamChunk{
			Type:      "error",
			SessionID: c.session.ID,
			AgentID:   c.session.AgentID,
			Error:     "ANTHROPIC_API_KEY not configured",
		})
		return
	}

	// Add user message
	c.session.AddMessage(RoleUser, req.Message)

	// Send start event
	c.sendJSON(StreamChunk{
		Type:      "start",
		SessionID: c.session.ID,
		AgentID:   c.session.AgentID,
	})

	// Stream response with tool use
	var fullResponse string
	chatReq := &agent.ChatRequest{
		AgentID:   c.session.AgentID,
		ProjectID: c.session.ProjectID,
		VoiceMode: c.voiceMode,
		Message:   req.Message,
	}

	err := c.handler.runtime.Chat(ctx, chatReq, func(event agent.ChatEvent) {
		switch event.Type {
		case "text":
			fullResponse += event.Text
			c.sendJSON(StreamChunk{
				Type:      "delta",
				SessionID: c.session.ID,
				AgentID:   c.session.AgentID,
				Content:   event.Text,
			})

		case "tool_call":
			c.sendJSON(map[string]interface{}{
				"type":       "tool_call",
				"session_id": c.session.ID,
				"agent_id":   c.session.AgentID,
				"tool_name":  event.ToolCall.Name,
				"tool_input": json.RawMessage(event.ToolCall.Input),
			})

		case "tool_result":
			c.sendJSON(map[string]interface{}{
				"type":       "tool_result",
				"session_id": c.session.ID,
				"agent_id":   c.session.AgentID,
				"is_error":   event.ToolResult.IsError,
			})

		case "error":
			c.sendJSON(StreamChunk{
				Type:      "error",
				SessionID: c.session.ID,
				AgentID:   c.session.AgentID,
				Error:     event.Error,
			})
		}
	})

	if err != nil {
		c.sendJSON(StreamChunk{
			Type:      "error",
			SessionID: c.session.ID,
			AgentID:   c.session.AgentID,
			Error:     err.Error(),
		})
		return
	}

	// Add assistant response
	if fullResponse != "" {
		c.session.AddMessage(RoleAssistant, fullResponse)
	}

	// Send done event
	c.sendJSON(StreamChunk{
		Type:      "done",
		SessionID: c.session.ID,
		AgentID:   c.session.AgentID,
		Content:   fullResponse,
	})
}

func (c *wsClient) writePump(ctx context.Context) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return

		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Helper functions

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string, err error) {
	resp := map[string]string{"error": message}
	if err != nil {
		resp["details"] = err.Error()
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

func sendSSE(w http.ResponseWriter, flusher http.Flusher, data interface{}) {
	jsonData, _ := json.Marshal(data)
	fmt.Fprintf(w, "data: %s\n\n", jsonData)
	flusher.Flush()
}
