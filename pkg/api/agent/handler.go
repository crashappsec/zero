package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

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

// Handler manages agent chat functionality
type Handler struct {
	sessions *SessionManager
	claude   *ClaudeClient
	zeroHome string
}

// NewHandler creates a new agent handler
func NewHandler(zeroHome string) *Handler {
	return &Handler{
		sessions: NewSessionManager(),
		claude:   NewClaudeClient(zeroHome),
		zeroHome: zeroHome,
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

	// Get or create session
	session := h.sessions.GetOrCreate(sessionID, agentID)

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Create client
	client := &wsClient{
		handler: h,
		session: session,
		conn:    conn,
		send:    make(chan []byte, 256),
	}

	// Send connection confirmation (use session's agentID in case it already existed)
	client.sendJSON(map[string]interface{}{
		"type":       "connected",
		"session_id": session.ID,
		"agent_id":   session.AgentID,
		"agent_name": GetAgentInfo(session.AgentID).Name,
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

	// Check if Claude is available
	if !h.claude.IsAvailable() {
		writeError(w, http.StatusServiceUnavailable, "ANTHROPIC_API_KEY not configured", nil)
		return
	}

	// Add user message to session
	session.AddMessage(RoleUser, req.Message)

	// Get response from Claude
	response, err := h.claude.Chat(r.Context(), session, req.Message)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "chat failed", err)
		return
	}

	// Add assistant response to session
	session.AddMessage(RoleAssistant, response)

	writeJSON(w, http.StatusOK, ChatResponse{
		SessionID: sessionID,
		AgentID:   agentID,
		Response:  response,
		Done:      true,
	})
}

// HandleChatStream handles SSE streaming chat
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

	if !h.claude.IsAvailable() {
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

	// Stream response
	var fullResponse string
	err := h.claude.ChatStream(r.Context(), session, req.Message, func(chunk StreamChunk) {
		data, _ := json.Marshal(chunk)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()

		if chunk.Type == "done" {
			fullResponse = chunk.Content
		}
	})

	if err != nil {
		errChunk := StreamChunk{
			Type:      "error",
			SessionID: sessionID,
			AgentID:   agentID,
			Error:     err.Error(),
		}
		data, _ := json.Marshal(errChunk)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
		return
	}

	// Add assistant response to session
	if fullResponse != "" {
		session.AddMessage(RoleAssistant, fullResponse)
	}
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

// wsClient represents a WebSocket client for agent chat
type wsClient struct {
	handler *Handler
	session *Session
	conn    *websocket.Conn
	send    chan []byte
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

	// Check Claude availability
	if !c.handler.claude.IsAvailable() {
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

	// Stream response
	var fullResponse string
	err := c.handler.claude.ChatStream(ctx, c.session, req.Message, func(chunk StreamChunk) {
		c.sendJSON(chunk)
		if chunk.Type == "done" {
			fullResponse = chunk.Content
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
