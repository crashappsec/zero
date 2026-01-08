package agent

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestCheckOrigin(t *testing.T) {
	tests := []struct {
		name     string
		origin   string
		host     string
		expected bool
	}{
		{
			name:     "no origin header - allowed",
			origin:   "",
			host:     "example.com",
			expected: true,
		},
		{
			name:     "localhost origin - allowed",
			origin:   "http://localhost:3000",
			host:     "example.com",
			expected: true,
		},
		{
			name:     "127.0.0.1 origin - allowed",
			origin:   "http://127.0.0.1:8080",
			host:     "example.com",
			expected: true,
		},
		{
			name:     "same origin - allowed",
			origin:   "http://example.com",
			host:     "example.com",
			expected: true,
		},
		{
			name:     "same origin with port - allowed",
			origin:   "http://example.com:8080",
			host:     "example.com:8080",
			expected: true,
		},
		{
			name:     "different origin - rejected",
			origin:   "http://evil.com",
			host:     "example.com",
			expected: false,
		},
		{
			name:     "invalid origin URL - rejected",
			origin:   "://invalid",
			host:     "example.com",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/ws", nil)
			req.Host = tt.host
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			result := checkOrigin(req)
			if result != tt.expected {
				t.Errorf("checkOrigin() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNewHandler(t *testing.T) {
	// Test with non-existent path (should still create handler, but runtime may be nil)
	h := NewHandler("/nonexistent/path")
	if h == nil {
		t.Fatal("NewHandler returned nil")
	}
	if h.sessions == nil {
		t.Error("sessions manager is nil")
	}
}

func TestHandler_HandleListAgents(t *testing.T) {
	h := &Handler{
		sessions: NewSessionManager(),
	}

	req := httptest.NewRequest("GET", "/api/agent/agents", nil)
	w := httptest.NewRecorder()

	h.HandleListAgents(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	data, ok := resp["data"].([]interface{})
	if !ok {
		t.Fatal("response missing 'data' array")
	}

	// Should have at least the core agents
	if len(data) < 10 {
		t.Errorf("expected at least 10 agents, got %d", len(data))
	}

	// Check first agent structure
	agent0 := data[0].(map[string]interface{})
	requiredFields := []string{"id", "name", "persona", "description"}
	for _, field := range requiredFields {
		if _, ok := agent0[field]; !ok {
			t.Errorf("agent missing field %q", field)
		}
	}

	// Verify "zero" is in the list
	foundZero := false
	for _, a := range data {
		agent := a.(map[string]interface{})
		if agent["id"] == "zero" {
			foundZero = true
			if agent["name"] != "Zero" {
				t.Errorf("zero agent name = %v, want Zero", agent["name"])
			}
			break
		}
	}
	if !foundZero {
		t.Error("zero agent not found in list")
	}
}

func TestHandler_HandleListSessions(t *testing.T) {
	h := &Handler{
		sessions: NewSessionManager(),
	}

	// Create some sessions
	h.sessions.Create("session1", "zero")
	h.sessions.Create("session2", "cereal")
	s3 := h.sessions.Create("session3", "razor")
	s3.AddMessage(RoleUser, "test message")
	s3.AddMessage(RoleAssistant, "test response")

	req := httptest.NewRequest("GET", "/api/agent/sessions", nil)
	w := httptest.NewRecorder()

	h.HandleListSessions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	total, ok := resp["total"].(float64)
	if !ok {
		t.Fatal("response missing 'total'")
	}
	if int(total) != 3 {
		t.Errorf("total = %v, want 3", total)
	}

	data := resp["data"].([]interface{})
	if len(data) != 3 {
		t.Errorf("data length = %d, want 3", len(data))
	}

	// Find session3 and verify message count
	for _, item := range data {
		s := item.(map[string]interface{})
		if s["session_id"] == "session3" {
			msgCount := s["message_count"].(float64)
			if int(msgCount) != 2 {
				t.Errorf("session3 message_count = %v, want 2", msgCount)
			}
		}
	}
}

func TestHandler_HandleGetSession(t *testing.T) {
	h := &Handler{
		sessions: NewSessionManager(),
	}

	// Create a session with messages
	session := h.sessions.Create("test-session", "zero")
	session.SetProject("test-project")
	session.AddMessage(RoleUser, "hello")
	session.AddMessage(RoleAssistant, "hi there")

	t.Run("existing session", func(t *testing.T) {
		// Need to set up chi context for URL params
		r := chi.NewRouter()
		r.Get("/api/agent/sessions/{sessionID}", h.HandleGetSession)

		req := httptest.NewRequest("GET", "/api/agent/sessions/test-session", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp["session_id"] != "test-session" {
			t.Errorf("session_id = %v, want test-session", resp["session_id"])
		}
		if resp["agent_id"] != "zero" {
			t.Errorf("agent_id = %v, want zero", resp["agent_id"])
		}
		if resp["project_id"] != "test-project" {
			t.Errorf("project_id = %v, want test-project", resp["project_id"])
		}

		messages := resp["messages"].([]interface{})
		if len(messages) != 2 {
			t.Errorf("messages length = %d, want 2", len(messages))
		}
	})

	t.Run("non-existent session", func(t *testing.T) {
		r := chi.NewRouter()
		r.Get("/api/agent/sessions/{sessionID}", h.HandleGetSession)

		req := httptest.NewRequest("GET", "/api/agent/sessions/nonexistent", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}

func TestHandler_HandleDeleteSession(t *testing.T) {
	h := &Handler{
		sessions: NewSessionManager(),
	}

	h.sessions.Create("to-delete", "zero")

	r := chi.NewRouter()
	r.Delete("/api/agent/sessions/{sessionID}", h.HandleDeleteSession)

	req := httptest.NewRequest("DELETE", "/api/agent/sessions/to-delete", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNoContent)
	}

	// Verify session is deleted
	if _, ok := h.sessions.Get("to-delete"); ok {
		t.Error("session should have been deleted")
	}
}

func TestHandler_HandleChat_ValidationErrors(t *testing.T) {
	h := &Handler{
		sessions: NewSessionManager(),
		runtime:  nil, // No runtime - will fail on API key check
	}

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/agent/chat", bytes.NewBufferString("invalid json"))
		w := httptest.NewRecorder()
		h.HandleChat(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("empty message", func(t *testing.T) {
		body := `{"message": ""}`
		req := httptest.NewRequest("POST", "/api/agent/chat", bytes.NewBufferString(body))
		w := httptest.NewRecorder()
		h.HandleChat(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("no runtime available", func(t *testing.T) {
		body := `{"message": "hello"}`
		req := httptest.NewRequest("POST", "/api/agent/chat", bytes.NewBufferString(body))
		w := httptest.NewRecorder()
		h.HandleChat(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
		}
	})
}

func TestHandler_HandleChatStream_ValidationErrors(t *testing.T) {
	h := &Handler{
		sessions: NewSessionManager(),
		runtime:  nil,
	}

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/agent/chat/stream", bytes.NewBufferString("bad"))
		w := httptest.NewRecorder()
		h.HandleChatStream(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("empty message", func(t *testing.T) {
		body := `{"message": ""}`
		req := httptest.NewRequest("POST", "/api/agent/chat/stream", bytes.NewBufferString(body))
		w := httptest.NewRecorder()
		h.HandleChatStream(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("no runtime available", func(t *testing.T) {
		body := `{"message": "hello"}`
		req := httptest.NewRequest("POST", "/api/agent/chat/stream", bytes.NewBufferString(body))
		w := httptest.NewRecorder()
		h.HandleChatStream(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
		}
	})
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"key": "value"}

	writeJSON(w, http.StatusCreated, data)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", contentType)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp["key"] != "value" {
		t.Errorf("response key = %q, want value", resp["key"])
	}
}

func TestWriteError(t *testing.T) {
	t.Run("without error details", func(t *testing.T) {
		w := httptest.NewRecorder()
		writeError(w, http.StatusBadRequest, "something went wrong", nil)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}

		var resp map[string]string
		json.NewDecoder(w.Body).Decode(&resp)
		if resp["error"] != "something went wrong" {
			t.Errorf("error = %q, want 'something went wrong'", resp["error"])
		}
		if _, ok := resp["details"]; ok {
			t.Error("details should not be present")
		}
	})

	t.Run("with error details", func(t *testing.T) {
		w := httptest.NewRecorder()
		writeError(w, http.StatusInternalServerError, "failed", &testError{msg: "underlying cause"})

		var resp map[string]string
		json.NewDecoder(w.Body).Decode(&resp)
		if resp["details"] != "underlying cause" {
			t.Errorf("details = %q, want 'underlying cause'", resp["details"])
		}
	})
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestSendSSE(t *testing.T) {
	w := httptest.NewRecorder()
	flusher := w // httptest.ResponseRecorder implements Flusher

	data := StreamChunk{
		Type:      "delta",
		SessionID: "sess1",
		AgentID:   "zero",
		Content:   "Hello",
	}

	sendSSE(w, flusher, data)

	body := w.Body.String()
	if !bytes.Contains([]byte(body), []byte("data: ")) {
		t.Error("SSE response should start with 'data: '")
	}
	if !bytes.Contains([]byte(body), []byte(`"type":"delta"`)) {
		t.Error("SSE response should contain type")
	}
	if !bytes.Contains([]byte(body), []byte(`"content":"Hello"`)) {
		t.Error("SSE response should contain content")
	}
}

// Integration test for session creation via chat request
func TestHandler_ChatCreatesSession(t *testing.T) {
	h := &Handler{
		sessions: NewSessionManager(),
		runtime:  nil, // Will fail at runtime check, but session should be created first
	}

	body := `{"message": "hello", "session_id": "new-session", "agent_id": "cereal", "project_id": "my-project"}`
	req := httptest.NewRequest("POST", "/api/agent/chat", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleChat(w, req)

	// Request will fail due to no runtime, but session should exist
	session, ok := h.sessions.Get("new-session")
	if !ok {
		t.Fatal("session should have been created")
	}
	if session.AgentID != "cereal" {
		t.Errorf("agent_id = %q, want cereal", session.AgentID)
	}
	if session.ProjectID != "my-project" {
		t.Errorf("project_id = %q, want my-project", session.ProjectID)
	}
}

// Benchmark session operations
func BenchmarkSessionManager_GetOrCreate(b *testing.B) {
	sm := NewSessionManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.GetOrCreate("session-1", "zero")
	}
}

func BenchmarkSessionManager_ConcurrentAccess(b *testing.B) {
	sm := NewSessionManager()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			i++
			id := "session-" + string(rune(i%100))
			sm.GetOrCreate(id, "zero")
			if s, ok := sm.Get(id); ok {
				s.AddMessage(RoleUser, "test")
			}
		}
	})
}
