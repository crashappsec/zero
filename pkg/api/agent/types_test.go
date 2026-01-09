package agent

import (
	"sync"
	"testing"
	"time"
)

func TestNewSession(t *testing.T) {
	session := NewSession("test-id", "zero")

	if session.ID != "test-id" {
		t.Errorf("ID = %q, want test-id", session.ID)
	}
	if session.AgentID != "zero" {
		t.Errorf("AgentID = %q, want zero", session.AgentID)
	}
	if session.ProjectID != "" {
		t.Errorf("ProjectID = %q, want empty", session.ProjectID)
	}
	if len(session.Messages) != 0 {
		t.Errorf("Messages length = %d, want 0", len(session.Messages))
	}
	if session.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if session.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
}

func TestSession_AddMessage(t *testing.T) {
	session := NewSession("test", "zero")
	beforeUpdate := session.UpdatedAt

	// Small delay to ensure timestamp differs
	time.Sleep(time.Millisecond)

	session.AddMessage(RoleUser, "hello")
	session.AddMessage(RoleAssistant, "hi there")
	session.AddMessage(RoleSystem, "system message")

	msgs := session.GetMessages()
	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs))
	}

	if msgs[0].Role != RoleUser || msgs[0].Content != "hello" {
		t.Errorf("message 0 = %+v, want user/hello", msgs[0])
	}
	if msgs[1].Role != RoleAssistant || msgs[1].Content != "hi there" {
		t.Errorf("message 1 = %+v, want assistant/hi there", msgs[1])
	}
	if msgs[2].Role != RoleSystem || msgs[2].Content != "system message" {
		t.Errorf("message 2 = %+v, want system/system message", msgs[2])
	}

	// Check timestamps
	for i, msg := range msgs {
		if msg.Timestamp.IsZero() {
			t.Errorf("message %d timestamp is zero", i)
		}
	}

	// Check UpdatedAt was updated
	if !session.UpdatedAt.After(beforeUpdate) {
		t.Error("UpdatedAt should be after original timestamp")
	}
}

func TestSession_GetMessages_ReturnsCopy(t *testing.T) {
	session := NewSession("test", "zero")
	session.AddMessage(RoleUser, "original")

	msgs := session.GetMessages()
	msgs[0].Content = "modified"

	// Original should be unchanged
	original := session.GetMessages()
	if original[0].Content != "original" {
		t.Error("GetMessages should return a copy")
	}
}

func TestSession_SetProject(t *testing.T) {
	session := NewSession("test", "zero")
	beforeUpdate := session.UpdatedAt

	time.Sleep(time.Millisecond)

	session.SetProject("my-project")

	if session.ProjectID != "my-project" {
		t.Errorf("ProjectID = %q, want my-project", session.ProjectID)
	}
	if !session.UpdatedAt.After(beforeUpdate) {
		t.Error("UpdatedAt should be updated")
	}
}

func TestSession_ConcurrentAccess(t *testing.T) {
	session := NewSession("test", "zero")

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			session.AddMessage(RoleUser, "message")
			session.GetMessages()
			session.SetProject("project")
		}(i)
	}

	wg.Wait()

	msgs := session.GetMessages()
	if len(msgs) != 100 {
		t.Errorf("expected 100 messages, got %d", len(msgs))
	}
}

func TestSessionManager_Create(t *testing.T) {
	sm := NewSessionManager()

	session := sm.Create("sess-1", "cereal")

	if session.ID != "sess-1" {
		t.Errorf("ID = %q, want sess-1", session.ID)
	}
	if session.AgentID != "cereal" {
		t.Errorf("AgentID = %q, want cereal", session.AgentID)
	}

	// Should be retrievable
	retrieved, ok := sm.Get("sess-1")
	if !ok {
		t.Error("session should exist")
	}
	if retrieved != session {
		t.Error("retrieved session should be same instance")
	}
}

func TestSessionManager_Get(t *testing.T) {
	sm := NewSessionManager()

	// Non-existent
	_, ok := sm.Get("nonexistent")
	if ok {
		t.Error("should return false for non-existent session")
	}

	// After creation
	sm.Create("exists", "zero")
	session, ok := sm.Get("exists")
	if !ok {
		t.Error("should return true for existing session")
	}
	if session == nil {
		t.Error("session should not be nil")
	}
}

func TestSessionManager_GetOrCreate(t *testing.T) {
	sm := NewSessionManager()

	// First call creates
	session1 := sm.GetOrCreate("sess", "zero")
	if session1 == nil {
		t.Fatal("session should not be nil")
	}

	// Second call returns existing
	session2 := sm.GetOrCreate("sess", "different-agent")
	if session2 != session1 {
		t.Error("should return same session instance")
	}
	if session2.AgentID != "zero" {
		t.Error("agent ID should not change for existing session")
	}
}

func TestSessionManager_Delete(t *testing.T) {
	sm := NewSessionManager()
	sm.Create("to-delete", "zero")

	// Verify exists
	if _, ok := sm.Get("to-delete"); !ok {
		t.Fatal("session should exist before delete")
	}

	sm.Delete("to-delete")

	if _, ok := sm.Get("to-delete"); ok {
		t.Error("session should not exist after delete")
	}

	// Delete non-existent should not panic
	sm.Delete("nonexistent")
}

func TestSessionManager_List(t *testing.T) {
	sm := NewSessionManager()

	// Empty list
	sessions := sm.List()
	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(sessions))
	}

	// Create some sessions
	sm.Create("sess-1", "zero")
	sm.Create("sess-2", "cereal")
	sm.Create("sess-3", "razor")

	sessions = sm.List()
	if len(sessions) != 3 {
		t.Errorf("expected 3 sessions, got %d", len(sessions))
	}

	// Verify all sessions are present
	ids := make(map[string]bool)
	for _, s := range sessions {
		ids[s.ID] = true
	}
	for _, expected := range []string{"sess-1", "sess-2", "sess-3"} {
		if !ids[expected] {
			t.Errorf("session %q not found in list", expected)
		}
	}
}

func TestSessionManager_Cleanup(t *testing.T) {
	sm := NewSessionManager()

	// Create sessions with different ages
	old := sm.Create("old", "zero")
	old.UpdatedAt = time.Now().Add(-2 * time.Hour)

	recent := sm.Create("recent", "zero")
	recent.UpdatedAt = time.Now().Add(-30 * time.Minute)

	fresh := sm.Create("fresh", "zero")
	fresh.UpdatedAt = time.Now()

	// Cleanup sessions older than 1 hour
	removed := sm.Cleanup(time.Hour)

	if removed != 1 {
		t.Errorf("removed = %d, want 1", removed)
	}

	// Old should be gone
	if _, ok := sm.Get("old"); ok {
		t.Error("old session should have been removed")
	}

	// Recent and fresh should remain
	if _, ok := sm.Get("recent"); !ok {
		t.Error("recent session should remain")
	}
	if _, ok := sm.Get("fresh"); !ok {
		t.Error("fresh session should remain")
	}
}

func TestSessionManager_ConcurrentAccess(t *testing.T) {
	sm := NewSessionManager()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			id := "session"
			sm.GetOrCreate(id, "zero")
			sm.Get(id)
			sm.List()
		}(i)
	}

	wg.Wait()
}

func TestGetAgentInfo(t *testing.T) {
	tests := []struct {
		agentID     string
		expectName  string
		expectID    string
		expectEmpty bool
	}{
		{"zero", "Zero", "zero", false},
		{"cereal", "Cereal", "cereal", false},
		{"razor", "Razor", "razor", false},
		{"blade", "Blade", "blade", false},
		{"phreak", "Phreak", "phreak", false},
		{"acid", "Acid", "acid", false},
		{"dade", "Dade", "dade", false},
		{"nikon", "Nikon", "nikon", false},
		{"joey", "Joey", "joey", false},
		{"plague", "Plague", "plague", false},
		{"gibson", "Gibson", "gibson", false},
		{"gill", "Gill", "gill", false},
		{"hal", "Hal", "hal", false},
		// Unknown agent should default to zero
		{"unknown", "Zero", "zero", false},
		{"", "Zero", "zero", false},
	}

	for _, tt := range tests {
		t.Run(tt.agentID, func(t *testing.T) {
			info := GetAgentInfo(tt.agentID)

			if info.Name != tt.expectName {
				t.Errorf("Name = %q, want %q", info.Name, tt.expectName)
			}
			if info.ID != tt.expectID {
				t.Errorf("ID = %q, want %q", info.ID, tt.expectID)
			}
		})
	}
}

func TestAgentInfo_AllAgentsHaveRequiredFields(t *testing.T) {
	agents := []string{
		"zero", "cereal", "razor", "blade", "phreak",
		"acid", "dade", "nikon", "joey", "plague",
		"gibson", "gill", "hal",
	}

	for _, agentID := range agents {
		t.Run(agentID, func(t *testing.T) {
			info := GetAgentInfo(agentID)

			if info.ID == "" {
				t.Error("ID should not be empty")
			}
			if info.Name == "" {
				t.Error("Name should not be empty")
			}
			if info.Persona == "" {
				t.Error("Persona should not be empty")
			}
			if info.Description == "" {
				t.Error("Description should not be empty")
			}
			if info.Scanner == "" {
				t.Error("Scanner should not be empty")
			}
		})
	}
}

func TestRoleConstants(t *testing.T) {
	if RoleUser != "user" {
		t.Errorf("RoleUser = %q, want user", RoleUser)
	}
	if RoleAssistant != "assistant" {
		t.Errorf("RoleAssistant = %q, want assistant", RoleAssistant)
	}
	if RoleSystem != "system" {
		t.Errorf("RoleSystem = %q, want system", RoleSystem)
	}
}

func TestChatRequest_Fields(t *testing.T) {
	req := ChatRequest{
		SessionID: "sess-123",
		AgentID:   "cereal",
		ProjectID: "proj-456",
		VoiceMode: "minimal",
		Message:   "hello",
	}

	if req.SessionID != "sess-123" {
		t.Errorf("SessionID = %q, want sess-123", req.SessionID)
	}
	if req.AgentID != "cereal" {
		t.Errorf("AgentID = %q, want cereal", req.AgentID)
	}
	if req.ProjectID != "proj-456" {
		t.Errorf("ProjectID = %q, want proj-456", req.ProjectID)
	}
	if req.VoiceMode != "minimal" {
		t.Errorf("VoiceMode = %q, want minimal", req.VoiceMode)
	}
	if req.Message != "hello" {
		t.Errorf("Message = %q, want hello", req.Message)
	}
}

func TestStreamChunk_Fields(t *testing.T) {
	chunk := StreamChunk{
		Type:      "delta",
		SessionID: "sess",
		AgentID:   "zero",
		Content:   "response text",
		Error:     "",
	}

	if chunk.Type != "delta" {
		t.Errorf("Type = %q, want delta", chunk.Type)
	}
	if chunk.Content != "response text" {
		t.Errorf("Content = %q, want 'response text'", chunk.Content)
	}

	// Error chunk
	errChunk := StreamChunk{
		Type:  "error",
		Error: "something failed",
	}
	if errChunk.Error != "something failed" {
		t.Errorf("Error = %q, want 'something failed'", errChunk.Error)
	}
}

// Benchmarks

func BenchmarkNewSession(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewSession("test-id", "zero")
	}
}

func BenchmarkSession_AddMessage(b *testing.B) {
	session := NewSession("test", "zero")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		session.AddMessage(RoleUser, "test message")
	}
}

func BenchmarkSession_GetMessages(b *testing.B) {
	session := NewSession("test", "zero")
	for i := 0; i < 100; i++ {
		session.AddMessage(RoleUser, "message")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		session.GetMessages()
	}
}

func BenchmarkGetAgentInfo(b *testing.B) {
	agents := []string{"zero", "cereal", "razor", "unknown"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetAgentInfo(agents[i%len(agents)])
	}
}
