package scannertest

import (
	"testing"

	"github.com/crashappsec/zero/pkg/scanner"
	// Import scanners to register them (v3.6 super scanners)
	_ "github.com/crashappsec/zero/pkg/scanner/code-ownership"
	_ "github.com/crashappsec/zero/pkg/scanner/code-quality"
	_ "github.com/crashappsec/zero/pkg/scanner/code-security"
	_ "github.com/crashappsec/zero/pkg/scanner/code-crypto"
	_ "github.com/crashappsec/zero/pkg/scanner/developer-experience"
	_ "github.com/crashappsec/zero/pkg/scanner/devops"
	_ "github.com/crashappsec/zero/pkg/scanner/packages"
	_ "github.com/crashappsec/zero/pkg/scanner/sbom"
	_ "github.com/crashappsec/zero/pkg/scanner/tech-id"
)

func TestRegisteredScanners(t *testing.T) {
	// List all registered scanners
	scanners := scanner.List()

	// v3.6 super scanners
	expected := []string{
		"sbom",
		"packages",
		"code-crypto",
		"code-security",
		"code-quality",
		"devops",
		"developer-experience",
		"tech-id",
		"code-ownership",
	}

	t.Logf("Registered scanners: %v", scanners)

	if len(scanners) < len(expected) {
		t.Errorf("Expected at least %d scanners, got %d", len(expected), len(scanners))
	}

	for _, name := range expected {
		s, ok := scanner.Get(name)
		if !ok {
			t.Errorf("Scanner %s not registered", name)
			continue
		}
		t.Logf("Scanner %s: %s", name, s.Description())
	}
}

func TestTopologicalSort(t *testing.T) {
	// Get all scanners
	scanners := scanner.GetAll()

	sorted, err := scanner.TopologicalSort(scanners)
	if err != nil {
		t.Fatalf("TopologicalSort failed: %v", err)
	}

	t.Logf("Sorted order:")
	for i, s := range sorted {
		t.Logf("  %d. %s (deps: %v)", i+1, s.Name(), s.Dependencies())
	}

	// Verify SBOM comes before packages (since packages depends on sbom)
	sbomIdx := -1
	pkgIdx := -1

	for i, s := range sorted {
		switch s.Name() {
		case "sbom":
			sbomIdx = i
		case "packages":
			pkgIdx = i
		}
	}

	if sbomIdx >= 0 && pkgIdx >= 0 && sbomIdx > pkgIdx {
		t.Error("sbom should come before packages")
	}
}

func TestGroupByDependencies(t *testing.T) {
	scanners := scanner.GetAll()

	groups, err := scanner.GroupByDependencies(scanners)
	if err != nil {
		t.Fatalf("GroupByDependencies failed: %v", err)
	}

	t.Logf("Dependency groups:")
	for i, group := range groups {
		names := make([]string, len(group))
		for j, s := range group {
			names[j] = s.Name()
		}
		t.Logf("  Level %d: %v", i+1, names)
	}

	// First level should contain scanners with no deps
	if len(groups) > 0 {
		firstLevel := groups[0]
		for _, s := range firstLevel {
			deps := s.Dependencies()
			if len(deps) > 0 {
				t.Errorf("Scanner %s in first level but has deps: %v", s.Name(), deps)
			}
		}
	}
}

func TestScannerDescriptions(t *testing.T) {
	// Test that each scanner has a description defined
	scanners := scanner.GetAll()

	for _, s := range scanners {
		desc := s.Description()
		t.Logf("Scanner %s: %s", s.Name(), desc)

		if desc == "" {
			t.Errorf("Scanner %s has no description defined", s.Name())
		}
	}
}
