package banter

import (
	"strings"
	"testing"
)

func TestGenerateID(t *testing.T) {
	// Test that ID generation works
	id := generateID()
	
	if len(id) != 8 {
		t.Errorf("Expected ID length of 8, got %d", len(id))
	}
	
	// Verify ID contains only expected characters
	validChars := "abcdefghijklmnopqrstuvwxyz0123456789"
	for _, char := range id {
		if !strings.ContainsRune(validChars, char) {
			t.Errorf("ID contains invalid character: %c", char)
		}
	}
}

func TestGenerateIDUniqueness(t *testing.T) {
	// Generate multiple IDs and verify they're different
	// This tests that crypto/rand is properly used
	ids := make(map[string]bool)
	iterations := 100
	
	for i := 0; i < iterations; i++ {
		id := generateID()
		if ids[id] {
			t.Errorf("Duplicate ID generated: %s", id)
		}
		ids[id] = true
	}
	
	if len(ids) != iterations {
		t.Errorf("Expected %d unique IDs, got %d", iterations, len(ids))
	}
}

func TestNewGenerator(t *testing.T) {
	gen, err := NewGenerator()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}
	
	if gen == nil {
		t.Fatal("Generator is nil")
	}
	
	if gen.personalities == nil {
		t.Error("Personalities not loaded")
	}
	
	if gen.enabled {
		t.Error("Generator should be disabled by default")
	}
}

func TestGeneratorSetEnabled(t *testing.T) {
	gen, err := NewGenerator()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}
	
	// Test enabling
	gen.SetEnabled(true)
	if !gen.IsEnabled() {
		t.Error("Generator should be enabled")
	}
	
	// Test disabling
	gen.SetEnabled(false)
	if gen.IsEnabled() {
		t.Error("Generator should be disabled")
	}
}

func TestGeneratePun(t *testing.T) {
	gen, err := NewGenerator()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}
	
	gen.SetEnabled(true)
	
	// Get first available agent
	agents := gen.ListAgents()
	if len(agents) == 0 {
		t.Skip("No agents available")
	}
	
	// Try to generate a pun
	msg := gen.GeneratePun(agents[0])
	if msg != nil {
		if msg.Type != "pun" {
			t.Errorf("Expected type 'pun', got '%s'", msg.Type)
		}
		if msg.ID == "" {
			t.Error("Message ID is empty")
		}
		if len(msg.ID) != 8 {
			t.Errorf("Expected ID length of 8, got %d", len(msg.ID))
		}
	}
}
