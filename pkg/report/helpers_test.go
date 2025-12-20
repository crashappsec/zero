package report

import (
	"testing"
)

func TestSeverityBadge(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"critical", "CRITICAL"},
		{"CRITICAL", "CRITICAL"},
		{"high", "HIGH"},
		{"High", "HIGH"},
		{"medium", "MEDIUM"},
		{"low", "LOW"},
		{"info", "INFO"},
		{"informational", "INFO"},
		{"unknown", "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := SeverityBadge(tt.input); got != tt.expected {
				t.Errorf("SeverityBadge(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestScoreGrade(t *testing.T) {
	tests := []struct {
		score    int
		expected string
	}{
		{100, "A"},
		{95, "A"},
		{90, "A"},
		{89, "B"},
		{80, "B"},
		{79, "C"},
		{70, "C"},
		{69, "D"},
		{60, "D"},
		{59, "F"},
		{0, "F"},
		{-10, "F"},
	}

	for _, tt := range tests {
		t.Run(string(rune('0'+tt.score)), func(t *testing.T) {
			if got := ScoreGrade(tt.score); got != tt.expected {
				t.Errorf("ScoreGrade(%d) = %s, want %s", tt.score, got, tt.expected)
			}
		})
	}
}

func TestScoreStatus(t *testing.T) {
	tests := []struct {
		score    int
		expected string
	}{
		{100, "Excellent"},
		{90, "Excellent"},
		{89, "Good"},
		{75, "Good"},
		{74, "Fair"},
		{50, "Fair"},
		{49, "Poor"},
		{25, "Poor"},
		{24, "Critical"},
		{0, "Critical"},
	}

	for _, tt := range tests {
		t.Run(string(rune('0'+tt.score)), func(t *testing.T) {
			if got := ScoreStatus(tt.score); got != tt.expected {
				t.Errorf("ScoreStatus(%d) = %s, want %s", tt.score, got, tt.expected)
			}
		})
	}
}

func TestBoolToCheckmark(t *testing.T) {
	if got := BoolToCheckmark(true); got != "Yes" {
		t.Errorf("BoolToCheckmark(true) = %s, want Yes", got)
	}
	if got := BoolToCheckmark(false); got != "No" {
		t.Errorf("BoolToCheckmark(false) = %s, want No", got)
	}
}

func TestBoolToYesNo(t *testing.T) {
	if got := BoolToYesNo(true); got != "Yes" {
		t.Errorf("BoolToYesNo(true) = %s, want Yes", got)
	}
	if got := BoolToYesNo(false); got != "No" {
		t.Errorf("BoolToYesNo(false) = %s, want No", got)
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is a long string", 10, "this is..."},
		{"abc", 3, "abc"},
		{"abcd", 3, "abc"},
		{"abcdef", 5, "ab..."},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := Truncate(tt.input, tt.maxLen); got != tt.expected {
				t.Errorf("Truncate(%s, %d) = %s, want %s", tt.input, tt.maxLen, got, tt.expected)
			}
		})
	}
}

func TestFormatCount(t *testing.T) {
	tests := []struct {
		count    int
		singular string
		plural   string
		expected string
	}{
		{0, "item", "items", "0 items"},
		{1, "item", "items", "1 item"},
		{2, "item", "items", "2 items"},
		{100, "vulnerability", "vulnerabilities", "100 vulnerabilities"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := FormatCount(tt.count, tt.singular, tt.plural); got != tt.expected {
				t.Errorf("FormatCount(%d, %s, %s) = %s, want %s",
					tt.count, tt.singular, tt.plural, got, tt.expected)
			}
		})
	}
}

func TestFormatPercent(t *testing.T) {
	tests := []struct {
		value    float64
		expected string
	}{
		{0.0, "0.0%"},
		{50.0, "50.0%"},
		{99.9, "99.9%"},
		{100.0, "100.0%"},
		{33.333, "33.3%"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := FormatPercent(tt.value); got != tt.expected {
				t.Errorf("FormatPercent(%f) = %s, want %s", tt.value, got, tt.expected)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		seconds  int
		expected string
	}{
		{0, "0s"},
		{30, "30s"},
		{59, "59s"},
		{60, "1m 0s"},
		{90, "1m 30s"},
		{3600, "1h 0m"},
		{3661, "1h 1m"},
		{7200, "2h 0m"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := FormatDuration(tt.seconds); got != tt.expected {
				t.Errorf("FormatDuration(%d) = %s, want %s", tt.seconds, got, tt.expected)
			}
		})
	}
}

func TestRiskLevel(t *testing.T) {
	tests := []struct {
		critical int
		high     int
		medium   int
		expected string
	}{
		{1, 0, 0, "Critical"},
		{0, 3, 0, "High"},
		{0, 1, 0, "Medium"},
		{0, 0, 6, "Medium"},
		{0, 0, 3, "Low"},
		{0, 0, 0, "None"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := RiskLevel(tt.critical, tt.high, tt.medium); got != tt.expected {
				t.Errorf("RiskLevel(%d, %d, %d) = %s, want %s",
					tt.critical, tt.high, tt.medium, got, tt.expected)
			}
		})
	}
}

func TestSeverityOrder(t *testing.T) {
	tests := []struct {
		severity string
		expected int
	}{
		{"critical", 5},
		{"CRITICAL", 5},
		{"high", 4},
		{"medium", 3},
		{"moderate", 3},
		{"low", 2},
		{"info", 1},
		{"informational", 1},
		{"unknown", 0},
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			if got := SeverityOrder(tt.severity); got != tt.expected {
				t.Errorf("SeverityOrder(%s) = %d, want %d", tt.severity, got, tt.expected)
			}
		})
	}
}

func TestNormalizeSeverity(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"critical", "critical"},
		{"CRITICAL", "critical"},
		{"crit", "critical"},
		{"high", "high"},
		{"HIGH", "high"},
		{"medium", "medium"},
		{"moderate", "medium"},
		{"med", "medium"},
		{"low", "low"},
		{"info", "info"},
		{"informational", "info"},
		{"note", "info"},
		{"unknown", "info"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := NormalizeSeverity(tt.input); got != tt.expected {
				t.Errorf("NormalizeSeverity(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestEscapeMarkdown(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"plain text", "plain text"},
		{"*bold*", "\\*bold\\*"},
		{"_italic_", "\\_italic\\_"},
		{"`code`", "\\`code\\`"},
		{"[link](url)", "\\[link\\]\\(url\\)"},
		{"# heading", "\\# heading"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := EscapeMarkdown(tt.input); got != tt.expected {
				t.Errorf("EscapeMarkdown(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestWrapInCodeBlock(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"plain", "plain"},
		{"with*star", "`with*star`"},
		{"with`backtick", "`with`backtick`"},
		{"path/to/file", "path/to/file"},
		{"array[0]", "`array[0]`"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := WrapInCodeBlock(tt.input); got != tt.expected {
				t.Errorf("WrapInCodeBlock(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}
