package findings

import (
	"testing"
)

func TestParseSeverity(t *testing.T) {
	tests := []struct {
		input    string
		expected Severity
	}{
		{"critical", SeverityCritical},
		{"CRITICAL", SeverityCritical},
		{"crit", SeverityCritical},
		{"high", SeverityHigh},
		{"HIGH", SeverityHigh},
		{"medium", SeverityMedium},
		{"MEDIUM", SeverityMedium},
		{"moderate", SeverityMedium},
		{"med", SeverityMedium},
		{"low", SeverityLow},
		{"LOW", SeverityLow},
		{"info", SeverityInfo},
		{"INFO", SeverityInfo},
		{"informational", SeverityInfo},
		{"note", SeverityInfo},
		{"none", SeverityInfo},
		{"unknown", SeverityInfo},
		{"", SeverityInfo},
		{"  critical  ", SeverityCritical},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseSeverity(tt.input)
			if got != tt.expected {
				t.Errorf("ParseSeverity(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseConfidence(t *testing.T) {
	tests := []struct {
		input    string
		expected Confidence
	}{
		{"high", ConfidenceHigh},
		{"HIGH", ConfidenceHigh},
		{"certain", ConfidenceHigh},
		{"confirmed", ConfidenceHigh},
		{"medium", ConfidenceMedium},
		{"probable", ConfidenceMedium},
		{"likely", ConfidenceMedium},
		{"low", ConfidenceLow},
		{"possible", ConfidenceLow},
		{"tentative", ConfidenceLow},
		{"unknown", ConfidenceMedium},
		{"", ConfidenceMedium},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseConfidence(tt.input)
			if got != tt.expected {
				t.Errorf("ParseConfidence(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSeverity_Score(t *testing.T) {
	tests := []struct {
		severity Severity
		expected int
	}{
		{SeverityCritical, 5},
		{SeverityHigh, 4},
		{SeverityMedium, 3},
		{SeverityLow, 2},
		{SeverityInfo, 1},
		{Severity("unknown"), 0},
	}

	for _, tt := range tests {
		t.Run(string(tt.severity), func(t *testing.T) {
			if got := tt.severity.Score(); got != tt.expected {
				t.Errorf("%s.Score() = %d, want %d", tt.severity, got, tt.expected)
			}
		})
	}
}

func TestSeverity_Comparisons(t *testing.T) {
	if !SeverityCritical.IsHigherThan(SeverityHigh) {
		t.Error("Critical should be higher than High")
	}
	if !SeverityLow.IsLowerThan(SeverityMedium) {
		t.Error("Low should be lower than Medium")
	}
	if !SeverityHigh.IsAtLeast(SeverityHigh) {
		t.Error("High should be at least High")
	}
	if !SeverityCritical.IsAtLeast(SeverityMedium) {
		t.Error("Critical should be at least Medium")
	}
	if SeverityLow.IsAtLeast(SeverityHigh) {
		t.Error("Low should not be at least High")
	}
}

func TestSeverity_IsCriticalOrHigh(t *testing.T) {
	if !SeverityCritical.IsCriticalOrHigh() {
		t.Error("Critical should be CriticalOrHigh")
	}
	if !SeverityHigh.IsCriticalOrHigh() {
		t.Error("High should be CriticalOrHigh")
	}
	if SeverityMedium.IsCriticalOrHigh() {
		t.Error("Medium should not be CriticalOrHigh")
	}
}

func TestSeverity_String(t *testing.T) {
	if SeverityCritical.String() != "critical" {
		t.Errorf("String() = %q, want %q", SeverityCritical.String(), "critical")
	}
	if SeverityCritical.Upper() != "CRITICAL" {
		t.Errorf("Upper() = %q, want %q", SeverityCritical.Upper(), "CRITICAL")
	}
	if SeverityCritical.Title() != "Critical" {
		t.Errorf("Title() = %q, want %q", SeverityCritical.Title(), "Critical")
	}
}

func TestAllSeverities(t *testing.T) {
	severities := AllSeverities()
	if len(severities) != 5 {
		t.Errorf("AllSeverities() length = %d, want 5", len(severities))
	}
	if severities[0] != SeverityCritical {
		t.Error("First severity should be Critical")
	}
	if severities[4] != SeverityInfo {
		t.Error("Last severity should be Info")
	}
}

func TestConfidence_Score(t *testing.T) {
	if ConfidenceHigh.Score() != 3 {
		t.Errorf("ConfidenceHigh.Score() = %d, want 3", ConfidenceHigh.Score())
	}
	if ConfidenceMedium.Score() != 2 {
		t.Errorf("ConfidenceMedium.Score() = %d, want 2", ConfidenceMedium.Score())
	}
	if ConfidenceLow.Score() != 1 {
		t.Errorf("ConfidenceLow.Score() = %d, want 1", ConfidenceLow.Score())
	}
}

func TestFilter(t *testing.T) {
	findings := []Finding{
		{ID: "1", Severity: SeverityCritical, Category: "security", Scanner: "scanner1"},
		{ID: "2", Severity: SeverityHigh, Category: "security", Scanner: "scanner1"},
		{ID: "3", Severity: SeverityMedium, Category: "quality", Scanner: "scanner2"},
		{ID: "4", Severity: SeverityLow, Category: "quality", Scanner: "scanner2"},
		{ID: "5", Severity: SeverityInfo, Category: "info", Scanner: "scanner1"},
	}

	t.Run("MinSeverity", func(t *testing.T) {
		result := Filter(findings, FilterOptions{MinSeverity: SeverityHigh})
		if len(result) != 2 {
			t.Errorf("got %d findings, want 2", len(result))
		}
	})

	t.Run("Categories", func(t *testing.T) {
		result := Filter(findings, FilterOptions{Categories: []string{"security"}, IncludeInfo: true})
		if len(result) != 2 {
			t.Errorf("got %d findings, want 2", len(result))
		}
	})

	t.Run("Scanners", func(t *testing.T) {
		result := Filter(findings, FilterOptions{Scanners: []string{"scanner2"}, IncludeInfo: true})
		if len(result) != 2 {
			t.Errorf("got %d findings, want 2", len(result))
		}
	})

	t.Run("Limit", func(t *testing.T) {
		result := Filter(findings, FilterOptions{Limit: 2, IncludeInfo: true})
		if len(result) != 2 {
			t.Errorf("got %d findings, want 2", len(result))
		}
	})

	t.Run("ExcludeInfo", func(t *testing.T) {
		result := Filter(findings, FilterOptions{})
		if len(result) != 4 {
			t.Errorf("got %d findings, want 4 (excluding info)", len(result))
		}
	})
}

func TestSortBySeverity(t *testing.T) {
	findings := []Finding{
		{ID: "1", Severity: SeverityLow, Confidence: ConfidenceHigh},
		{ID: "2", Severity: SeverityCritical, Confidence: ConfidenceMedium},
		{ID: "3", Severity: SeverityHigh, Confidence: ConfidenceHigh},
		{ID: "4", Severity: SeverityCritical, Confidence: ConfidenceHigh},
	}

	sorted := SortBySeverity(findings)

	// First should be critical with high confidence
	if sorted[0].ID != "4" {
		t.Errorf("First finding should be ID 4 (critical+high), got %s", sorted[0].ID)
	}
	// Second should be critical with medium confidence
	if sorted[1].ID != "2" {
		t.Errorf("Second finding should be ID 2 (critical+medium), got %s", sorted[1].ID)
	}
	// Last should be low
	if sorted[3].ID != "1" {
		t.Errorf("Last finding should be ID 1 (low), got %s", sorted[3].ID)
	}
}

func TestGroupByCategory(t *testing.T) {
	findings := []Finding{
		{ID: "1", Category: "security"},
		{ID: "2", Category: "security"},
		{ID: "3", Category: "quality"},
	}

	groups := GroupByCategory(findings)
	if len(groups["security"]) != 2 {
		t.Errorf("security group should have 2 findings, got %d", len(groups["security"]))
	}
	if len(groups["quality"]) != 1 {
		t.Errorf("quality group should have 1 finding, got %d", len(groups["quality"]))
	}
}

func TestDeduplicate(t *testing.T) {
	findings := []Finding{
		{ID: "1", Title: "Finding 1"},
		{ID: "2", Title: "Finding 2"},
		{ID: "1", Title: "Finding 1 duplicate"},
	}

	result := Deduplicate(findings)
	if len(result) != 2 {
		t.Errorf("got %d findings, want 2", len(result))
	}
}

func TestDeduplicateByContent(t *testing.T) {
	loc := &Location{File: "test.go"}
	findings := []Finding{
		{ID: "1", Title: "SQL Injection", Category: "security", Location: loc},
		{ID: "2", Title: "SQL Injection", Category: "security", Location: loc},
		{ID: "3", Title: "XSS", Category: "security", Location: loc},
	}

	result := DeduplicateByContent(findings)
	if len(result) != 2 {
		t.Errorf("got %d findings, want 2", len(result))
	}
}

func TestCountBySeverity(t *testing.T) {
	findings := []Finding{
		{Severity: SeverityCritical},
		{Severity: SeverityCritical},
		{Severity: SeverityHigh},
		{Severity: SeverityMedium},
	}

	counts := CountBySeverity(findings)
	if counts[SeverityCritical] != 2 {
		t.Errorf("critical count = %d, want 2", counts[SeverityCritical])
	}
	if counts[SeverityHigh] != 1 {
		t.Errorf("high count = %d, want 1", counts[SeverityHigh])
	}
}

func TestTopN(t *testing.T) {
	findings := []Finding{
		{ID: "1", Severity: SeverityLow},
		{ID: "2", Severity: SeverityCritical},
		{ID: "3", Severity: SeverityHigh},
	}

	result := TopN(findings, 2)
	if len(result) != 2 {
		t.Errorf("got %d findings, want 2", len(result))
	}
	if result[0].ID != "2" {
		t.Error("first finding should be critical")
	}
}

func TestCriticalAndHigh(t *testing.T) {
	findings := []Finding{
		{Severity: SeverityCritical},
		{Severity: SeverityHigh},
		{Severity: SeverityMedium},
		{Severity: SeverityLow},
	}

	result := CriticalAndHigh(findings)
	if len(result) != 2 {
		t.Errorf("got %d findings, want 2", len(result))
	}
}

func TestNewFinding(t *testing.T) {
	f := NewFinding("test-1", "Test Finding", "Description", SeverityHigh)

	if f.ID != "test-1" {
		t.Errorf("ID = %q, want %q", f.ID, "test-1")
	}
	if f.Severity != SeverityHigh {
		t.Errorf("Severity = %q, want %q", f.Severity, SeverityHigh)
	}
	if f.Confidence != ConfidenceHigh {
		t.Errorf("Confidence = %q, want %q (default)", f.Confidence, ConfidenceHigh)
	}
}

func TestFinding_Builders(t *testing.T) {
	f := NewFinding("1", "Title", "Desc", SeverityHigh).
		WithLocation("file.go", 10).
		WithCategory("security").
		WithScanner("test-scanner").
		WithRemediation("Fix it").
		WithReferences("https://example.com").
		WithMetadata("key", "value")

	if f.Location == nil || f.Location.File != "file.go" {
		t.Error("Location not set correctly")
	}
	if f.Category != "security" {
		t.Error("Category not set correctly")
	}
	if f.Scanner != "test-scanner" {
		t.Error("Scanner not set correctly")
	}
	if f.Remediation != "Fix it" {
		t.Error("Remediation not set correctly")
	}
	if len(f.References) != 1 {
		t.Error("References not set correctly")
	}
	if f.Metadata["key"] != "value" {
		t.Error("Metadata not set correctly")
	}
}

func TestFindingSet(t *testing.T) {
	findings := []Finding{
		{ID: "1", Severity: SeverityCritical, Category: "security"},
		{ID: "2", Severity: SeverityHigh, Category: "security"},
		{ID: "3", Severity: SeverityMedium, Category: "quality"},
	}

	fs := NewFindingSet("test-scanner", findings)

	if fs.Scanner != "test-scanner" {
		t.Errorf("Scanner = %q, want %q", fs.Scanner, "test-scanner")
	}
	if fs.Summary.Total != 3 {
		t.Errorf("Total = %d, want 3", fs.Summary.Total)
	}
	if fs.CriticalCount() != 1 {
		t.Errorf("CriticalCount = %d, want 1", fs.CriticalCount())
	}
	if !fs.HasCritical() {
		t.Error("HasCritical should be true")
	}
	if !fs.HasHigh() {
		t.Error("HasHigh should be true")
	}
}
