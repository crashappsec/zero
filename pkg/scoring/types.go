// Package scoring provides standardized score calculation utilities
package scoring

// Score represents a calculated score with metadata
type Score struct {
	Value      int              `json:"value"`      // 0-100
	Grade      string           `json:"grade"`      // A, B, C, D, F
	Level      string           `json:"level"`      // excellent, good, fair, poor, critical
	Components []ComponentScore `json:"components,omitempty"`
}

// ComponentScore represents a single component of a score
type ComponentScore struct {
	Name   string  `json:"name"`
	Value  int     `json:"value"`
	Weight float64 `json:"weight"`
	Max    int     `json:"max,omitempty"`
}

// RiskLevel represents overall risk assessment
type RiskLevel string

const (
	RiskCritical RiskLevel = "critical"
	RiskHigh     RiskLevel = "high"
	RiskMedium   RiskLevel = "medium"
	RiskLow      RiskLevel = "low"
	RiskNone     RiskLevel = "none"
)

// ScoreResult holds a complete scoring result with context
type ScoreResult struct {
	Score       Score             `json:"score"`
	RiskLevel   RiskLevel         `json:"risk_level"`
	Findings    int               `json:"findings"`
	Categories  map[string]int    `json:"categories,omitempty"`
	Metadata    map[string]any    `json:"metadata,omitempty"`
}

// NewScore creates a score with computed grade and level
func NewScore(value int) Score {
	if value < 0 {
		value = 0
	}
	if value > 100 {
		value = 100
	}
	return Score{
		Value: value,
		Grade: ValueToGrade(value),
		Level: ValueToLevel(value),
	}
}

// NewComponentScore creates a component score
func NewComponentScore(name string, value int, weight float64) ComponentScore {
	return ComponentScore{
		Name:   name,
		Value:  value,
		Weight: weight,
	}
}

// WithMax sets the maximum value for context
func (c ComponentScore) WithMax(max int) ComponentScore {
	c.Max = max
	return c
}

// IsExcellent returns true if score is A grade
func (s Score) IsExcellent() bool {
	return s.Value >= 90
}

// IsGood returns true if score is B grade or better
func (s Score) IsGood() bool {
	return s.Value >= 80
}

// IsPoor returns true if score is D or F grade
func (s Score) IsPoor() bool {
	return s.Value < 70
}

// IsCritical returns true if score is F grade
func (s Score) IsCritical() bool {
	return s.Value < 60
}

// String returns the risk level as a string
func (r RiskLevel) String() string {
	return string(r)
}

// Score returns a numeric value for the risk level
func (r RiskLevel) Score() int {
	switch r {
	case RiskCritical:
		return 5
	case RiskHigh:
		return 4
	case RiskMedium:
		return 3
	case RiskLow:
		return 2
	case RiskNone:
		return 1
	default:
		return 0
	}
}

// IsHighRisk returns true for critical or high risk
func (r RiskLevel) IsHighRisk() bool {
	return r == RiskCritical || r == RiskHigh
}
