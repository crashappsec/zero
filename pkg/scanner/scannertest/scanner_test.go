package scannertest

import (
	"testing"

	"github.com/crashappsec/zero/pkg/scanner"
	// Import scanners to register them
	_ "github.com/crashappsec/zero/pkg/scanners/licenses"
	_ "github.com/crashappsec/zero/pkg/scanners/package-health"
	_ "github.com/crashappsec/zero/pkg/scanners/package-malcontent"
	_ "github.com/crashappsec/zero/pkg/scanners/package-sbom"
	_ "github.com/crashappsec/zero/pkg/scanners/package-vulns"
)

func TestRegisteredScanners(t *testing.T) {
	// List all registered scanners
	scanners := scanner.List()

	expected := []string{
		"licenses",
		"package-health",
		"package-malcontent",
		"package-sbom",
		"package-vulns",
	}

	t.Logf("Registered scanners: %v", scanners)

	if len(scanners) != len(expected) {
		t.Errorf("Expected %d scanners, got %d", len(expected), len(scanners))
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

	// Verify SBOM comes before vulns and licenses
	sbomIdx := -1
	vulnsIdx := -1
	licensesIdx := -1

	for i, s := range sorted {
		switch s.Name() {
		case "package-sbom":
			sbomIdx = i
		case "package-vulns":
			vulnsIdx = i
		case "licenses":
			licensesIdx = i
		}
	}

	if sbomIdx >= 0 && vulnsIdx >= 0 && sbomIdx > vulnsIdx {
		t.Error("package-sbom should come before package-vulns")
	}

	if sbomIdx >= 0 && licensesIdx >= 0 && sbomIdx > licensesIdx {
		t.Error("package-sbom should come before licenses")
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

	// First level should contain scanners with no deps (sbom, malcontent)
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
