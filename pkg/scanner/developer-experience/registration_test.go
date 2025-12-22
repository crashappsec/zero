package developerexperience

import (
	"testing"

	"github.com/crashappsec/zero/pkg/scanner"
)

func TestDevXRegistered(t *testing.T) {
	s, ok := scanner.Get("devx")
	if !ok {
		t.Fatal("devx scanner not registered in registry")
	}
	if s.Name() != "devx" {
		t.Errorf("expected name 'devx', got %q", s.Name())
	}
	t.Logf("DevX scanner registered: %s", s.Description())
}
