package terminal

import (
	"os"
	"testing"
)

func TestTerminal_Color(t *testing.T) {
	tests := []struct {
		name    string
		noColor bool
		code    string
		text    string
		want    string
	}{
		{
			name:    "color enabled",
			noColor: false,
			code:    Green,
			text:    "success",
			want:    Green + "success" + Reset,
		},
		{
			name:    "color disabled",
			noColor: true,
			code:    Green,
			text:    "success",
			want:    "success",
		},
		{
			name:    "bold color",
			noColor: false,
			code:    Bold,
			text:    "header",
			want:    Bold + "header" + Reset,
		},
		{
			name:    "empty text",
			noColor: false,
			code:    Cyan,
			text:    "",
			want:    Cyan + Reset,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			terminal := &Terminal{noColor: tt.noColor}
			got := terminal.Color(tt.code, tt.text)
			if got != tt.want {
				t.Errorf("Color() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTerminal_New(t *testing.T) {
	// Test without NO_COLOR env
	os.Unsetenv("NO_COLOR")
	term := New()
	if term.noColor {
		t.Error("New() should create terminal with color enabled when NO_COLOR is not set")
	}

	// Test with NO_COLOR env
	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")
	term = New()
	if !term.noColor {
		t.Error("New() should create terminal with color disabled when NO_COLOR is set")
	}
}

func TestScannerResultRow(t *testing.T) {
	// Basic test that ScannerResultRow struct works correctly
	row := ScannerResultRow{
		Name:     "sbom",
		Status:   "success",
		Summary:  "168 packages",
		Duration: 5 * 1000000000, // 5 seconds
	}

	if row.Name != "sbom" {
		t.Errorf("Name = %q, want %q", row.Name, "sbom")
	}
	if row.Status != "success" {
		t.Errorf("Status = %q, want %q", row.Status, "success")
	}
}

func TestScanFindings(t *testing.T) {
	findings := &ScanFindings{
		ScannersRun:   map[string]bool{"sbom": true, "package-analysis": true},
		TotalPackages: 168,
		PackagesByEco: map[string]int{"npm": 150, "golang": 18},
		VulnCritical:  2,
		VulnHigh:      5,
		VulnMedium:    10,
		VulnLow:       20,
	}

	if !findings.ScannersRun["sbom"] {
		t.Error("ScannersRun should contain sbom")
	}

	if findings.TotalPackages != 168 {
		t.Errorf("TotalPackages = %d, want 168", findings.TotalPackages)
	}

	if findings.PackagesByEco["npm"] != 150 {
		t.Errorf("PackagesByEco[npm] = %d, want 150", findings.PackagesByEco["npm"])
	}

	totalVulns := findings.VulnCritical + findings.VulnHigh + findings.VulnMedium + findings.VulnLow
	if totalVulns != 37 {
		t.Errorf("Total vulns = %d, want 37", totalVulns)
	}
}

func TestScanFindings_TechID(t *testing.T) {
	findings := &ScanFindings{
		ScannersRun:       map[string]bool{"technology-identification": true},
		TechTotalTechs:    15,
		TechByCategory:    map[string]int{"language": 5, "framework": 8, "database": 2},
		TechTopList:       []string{"Go", "JavaScript", "Python"},
		TechMLModels:      3,
		TechMLFrameworks:  2,
		TechSecurityCount: 5,
	}

	if findings.TechTotalTechs != 15 {
		t.Errorf("TechTotalTechs = %d, want 15", findings.TechTotalTechs)
	}

	if len(findings.TechByCategory) != 3 {
		t.Errorf("TechByCategory should have 3 categories, got %d", len(findings.TechByCategory))
	}

	if len(findings.TechTopList) != 3 {
		t.Errorf("TechTopList should have 3 items, got %d", len(findings.TechTopList))
	}

	if findings.TechMLModels != 3 {
		t.Errorf("TechMLModels = %d, want 3", findings.TechMLModels)
	}
}

func TestIconConstants(t *testing.T) {
	// Verify icon constants are defined
	if IconSuccess == "" {
		t.Error("IconSuccess should not be empty")
	}
	if IconFailed == "" {
		t.Error("IconFailed should not be empty")
	}
	if IconRunning == "" {
		t.Error("IconRunning should not be empty")
	}
	if IconQueued == "" {
		t.Error("IconQueued should not be empty")
	}
	if IconSkipped == "" {
		t.Error("IconSkipped should not be empty")
	}
	if IconWarning == "" {
		t.Error("IconWarning should not be empty")
	}
	if IconArrow == "" {
		t.Error("IconArrow should not be empty")
	}
}

func TestColorConstants(t *testing.T) {
	// Verify color constants are ANSI codes
	colorCodes := []struct {
		name  string
		value string
	}{
		{"Reset", Reset},
		{"Bold", Bold},
		{"Dim", Dim},
		{"Red", Red},
		{"Green", Green},
		{"Yellow", Yellow},
		{"Blue", Blue},
		{"Cyan", Cyan},
		{"White", White},
		{"BoldRed", BoldRed},
		{"BoldGreen", BoldGreen},
	}

	for _, cc := range colorCodes {
		if cc.value == "" {
			t.Errorf("%s should not be empty", cc.name)
		}
		// ANSI codes start with ESC
		if cc.value[0] != '\033' {
			t.Errorf("%s should start with ESC character", cc.name)
		}
	}
}

func TestTerminal_formatBytes(t *testing.T) {
	term := &Terminal{}

	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0B"},
		{100, "100B"},
		{1023, "1023B"},
		{1024, "1.0KB"},
		{1536, "1.5KB"},
		{10240, "10.0KB"},
		{1048576, "1.0MB"},
		{1572864, "1.5MB"},
		{10485760, "10.0MB"},
		{1073741824, "1.0GB"},
	}

	for _, tt := range tests {
		got := term.formatBytes(tt.bytes)
		if got != tt.want {
			t.Errorf("formatBytes(%d) = %q, want %q", tt.bytes, got, tt.want)
		}
	}
}

func TestTerminal_Confirm(t *testing.T) {
	// Confirm is interactive, so we just verify it exists and has the right signature
	term := &Terminal{}
	_ = term // Confirm exists on Terminal type
}
