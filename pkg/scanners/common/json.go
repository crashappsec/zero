package common

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ReadJSON reads and unmarshals a JSON file
func ReadJSON(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("parsing %s: %w", path, err)
	}
	return nil
}

// WriteJSON marshals and writes a JSON file
func WriteJSON(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}

// ParseJSON parses JSON bytes into a value
func ParseJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// GetString safely extracts a string from a map
func GetString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// GetInt safely extracts an int from a map
func GetInt(m map[string]interface{}, key string) int {
	switch v := m[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	}
	return 0
}

// GetFloat safely extracts a float from a map
func GetFloat(m map[string]interface{}, key string) float64 {
	switch v := m[key].(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	}
	return 0
}

// GetBool safely extracts a bool from a map
func GetBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

// GetArray safely extracts an array from a map
func GetArray(m map[string]interface{}, key string) []interface{} {
	if v, ok := m[key].([]interface{}); ok {
		return v
	}
	return nil
}

// GetMap safely extracts a map from a map
func GetMap(m map[string]interface{}, key string) map[string]interface{} {
	if v, ok := m[key].(map[string]interface{}); ok {
		return v
	}
	return nil
}

// MergeMaps merges src into dst (src values override dst)
func MergeMaps(dst, src map[string]interface{}) {
	for k, v := range src {
		dst[k] = v
	}
}

// CountBySeverity counts findings by severity level
func CountBySeverity(findings []map[string]interface{}) map[string]int {
	counts := map[string]int{
		"critical": 0,
		"high":     0,
		"medium":   0,
		"low":      0,
		"info":     0,
	}

	for _, f := range findings {
		sev := NormalizeSeverity(GetString(f, "severity"))
		counts[sev]++
	}

	return counts
}

// NormalizeSeverity normalizes severity strings to standard values
func NormalizeSeverity(s string) string {
	switch s {
	case "CRITICAL", "Critical", "critical":
		return "critical"
	case "HIGH", "High", "high":
		return "high"
	case "MEDIUM", "Medium", "medium", "MODERATE", "Moderate", "moderate":
		return "medium"
	case "LOW", "Low", "low":
		return "low"
	case "INFO", "Info", "info", "INFORMATIONAL", "Informational", "informational", "NOTE", "note":
		return "info"
	default:
		return "info"
	}
}

// SeverityScore returns a numeric score for sorting (higher = more severe)
func SeverityScore(s string) int {
	switch NormalizeSeverity(s) {
	case "critical":
		return 5
	case "high":
		return 4
	case "medium":
		return 3
	case "low":
		return 2
	case "info":
		return 1
	default:
		return 0
	}
}
