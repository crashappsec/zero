package report

import (
	"fmt"
	"strings"
)

// SeverityBadge returns a formatted severity badge
func SeverityBadge(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "CRITICAL"
	case "high":
		return "HIGH"
	case "medium":
		return "MEDIUM"
	case "low":
		return "LOW"
	case "info", "informational":
		return "INFO"
	default:
		return strings.ToUpper(severity)
	}
}

// ScoreGrade converts a numeric score (0-100) to a letter grade
func ScoreGrade(score int) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	case score >= 60:
		return "D"
	default:
		return "F"
	}
}

// ScoreStatus converts a numeric score to a status string
func ScoreStatus(score int) string {
	switch {
	case score >= 90:
		return "Excellent"
	case score >= 75:
		return "Good"
	case score >= 50:
		return "Fair"
	case score >= 25:
		return "Poor"
	default:
		return "Critical"
	}
}

// BoolToCheckmark converts a bool to a checkmark or X
func BoolToCheckmark(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

// BoolToYesNo converts a bool to Yes/No (plain text)
func BoolToYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

// Truncate truncates a string to maxLen with ellipsis
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// FormatCount formats a count with proper pluralization
func FormatCount(count int, singular, plural string) string {
	if count == 1 {
		return fmt.Sprintf("%d %s", count, singular)
	}
	return fmt.Sprintf("%d %s", count, plural)
}

// FormatPercent formats a percentage
func FormatPercent(value float64) string {
	return fmt.Sprintf("%.1f%%", value)
}

// FormatDuration formats a duration in human-readable form
func FormatDuration(seconds int) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	if seconds < 3600 {
		return fmt.Sprintf("%dm %ds", seconds/60, seconds%60)
	}
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

// RiskLevel returns a risk level string based on counts
func RiskLevel(critical, high, medium int) string {
	switch {
	case critical > 0:
		return "Critical"
	case high > 2:
		return "High"
	case high > 0 || medium > 5:
		return "Medium"
	case medium > 0:
		return "Low"
	default:
		return "None"
	}
}

// SeverityOrder returns the sort order for a severity (higher = more severe)
func SeverityOrder(severity string) int {
	switch strings.ToLower(severity) {
	case "critical":
		return 5
	case "high":
		return 4
	case "medium", "moderate":
		return 3
	case "low":
		return 2
	case "info", "informational":
		return 1
	default:
		return 0
	}
}

// NormalizeSeverity normalizes severity strings to standard values
func NormalizeSeverity(s string) string {
	switch strings.ToLower(s) {
	case "critical", "crit":
		return "critical"
	case "high":
		return "high"
	case "medium", "moderate", "med":
		return "medium"
	case "low":
		return "low"
	case "info", "informational", "note":
		return "info"
	default:
		return "info"
	}
}

// EscapeMarkdown escapes special markdown characters
func EscapeMarkdown(s string) string {
	replacer := strings.NewReplacer(
		`\`, `\\`,
		"`", "\\`",
		"*", "\\*",
		"_", "\\_",
		"{", "\\{",
		"}", "\\}",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		".", "\\.",
		"!", "\\!",
		"|", "\\|",
	)
	return replacer.Replace(s)
}

// WrapInCodeBlock wraps content in a code block if it contains special chars
func WrapInCodeBlock(s string) string {
	if strings.ContainsAny(s, "`*_[]()#") {
		return fmt.Sprintf("`%s`", s)
	}
	return s
}
