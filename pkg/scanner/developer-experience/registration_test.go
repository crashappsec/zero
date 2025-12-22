package developerexperience

import (
	"testing"

	"github.com/crashappsec/zero/pkg/scanner"
)

func TestDevXRegistered(t *testing.T) {
	s, ok := scanner.Get("developer-experience")
	if !ok {
		t.Fatal("developer-experience scanner not registered in registry")
	}
	if s.Name() != "developer-experience" {
		t.Errorf("expected name 'developer-experience', got %q", s.Name())
	}
	t.Logf("Developer Experience scanner registered: %s", s.Description())
}
