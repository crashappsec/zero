package findings

import "strings"

// Severity represents finding severity level
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

// Confidence represents finding confidence level
type Confidence string

const (
	ConfidenceHigh   Confidence = "high"
	ConfidenceMedium Confidence = "medium"
	ConfidenceLow    Confidence = "low"
)

// ParseSeverity normalizes severity strings to standard Severity
func ParseSeverity(s string) Severity {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "critical", "crit":
		return SeverityCritical
	case "high":
		return SeverityHigh
	case "medium", "moderate", "med":
		return SeverityMedium
	case "low":
		return SeverityLow
	case "info", "informational", "note", "none":
		return SeverityInfo
	default:
		return SeverityInfo
	}
}

// ParseConfidence normalizes confidence strings
func ParseConfidence(s string) Confidence {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "high", "certain", "confirmed":
		return ConfidenceHigh
	case "medium", "probable", "likely":
		return ConfidenceMedium
	case "low", "possible", "tentative":
		return ConfidenceLow
	default:
		return ConfidenceMedium
	}
}

// Score returns numeric score for sorting (higher = more severe)
func (s Severity) Score() int {
	switch s {
	case SeverityCritical:
		return 5
	case SeverityHigh:
		return 4
	case SeverityMedium:
		return 3
	case SeverityLow:
		return 2
	case SeverityInfo:
		return 1
	default:
		return 0
	}
}

// String returns the string representation
func (s Severity) String() string {
	return string(s)
}

// Upper returns uppercase string representation
func (s Severity) Upper() string {
	return strings.ToUpper(string(s))
}

// Title returns title case string representation
func (s Severity) Title() string {
	str := string(s)
	if len(str) == 0 {
		return str
	}
	return strings.ToUpper(str[:1]) + str[1:]
}

// IsHigherThan returns true if this severity is higher than other
func (s Severity) IsHigherThan(other Severity) bool {
	return s.Score() > other.Score()
}

// IsLowerThan returns true if this severity is lower than other
func (s Severity) IsLowerThan(other Severity) bool {
	return s.Score() < other.Score()
}

// IsAtLeast returns true if this severity is at least as high as other
func (s Severity) IsAtLeast(other Severity) bool {
	return s.Score() >= other.Score()
}

// IsCriticalOrHigh returns true for critical or high severity
func (s Severity) IsCriticalOrHigh() bool {
	return s == SeverityCritical || s == SeverityHigh
}

// Score returns numeric score for confidence (higher = more confident)
func (c Confidence) Score() int {
	switch c {
	case ConfidenceHigh:
		return 3
	case ConfidenceMedium:
		return 2
	case ConfidenceLow:
		return 1
	default:
		return 0
	}
}

// String returns the string representation
func (c Confidence) String() string {
	return string(c)
}

// AllSeverities returns all severity levels in order (most to least severe)
func AllSeverities() []Severity {
	return []Severity{
		SeverityCritical,
		SeverityHigh,
		SeverityMedium,
		SeverityLow,
		SeverityInfo,
	}
}

// AllConfidences returns all confidence levels in order (most to least confident)
func AllConfidences() []Confidence {
	return []Confidence{
		ConfidenceHigh,
		ConfidenceMedium,
		ConfidenceLow,
	}
}
