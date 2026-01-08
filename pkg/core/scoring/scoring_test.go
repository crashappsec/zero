package scoring

import (
	"testing"

	"github.com/crashappsec/zero/pkg/core/findings"
)

func TestValueToGrade(t *testing.T) {
	tests := []struct {
		value    int
		expected string
	}{
		{100, "A"},
		{95, "A"},
		{90, "A"},
		{89, "B"},
		{85, "B"},
		{80, "B"},
		{79, "C"},
		{75, "C"},
		{70, "C"},
		{69, "D"},
		{65, "D"},
		{60, "D"},
		{59, "F"},
		{50, "F"},
		{0, "F"},
	}

	for _, tt := range tests {
		got := ValueToGrade(tt.value)
		if got != tt.expected {
			t.Errorf("ValueToGrade(%d) = %s, want %s", tt.value, got, tt.expected)
		}
	}
}

func TestValueToLevel(t *testing.T) {
	tests := []struct {
		value    int
		expected string
	}{
		{100, "excellent"},
		{90, "excellent"},
		{89, "good"},
		{75, "good"},
		{74, "fair"},
		{50, "fair"},
		{49, "poor"},
		{25, "poor"},
		{24, "critical"},
		{0, "critical"},
	}

	for _, tt := range tests {
		got := ValueToLevel(tt.value)
		if got != tt.expected {
			t.Errorf("ValueToLevel(%d) = %s, want %s", tt.value, got, tt.expected)
		}
	}
}

func TestNewScore(t *testing.T) {
	tests := []struct {
		value    int
		expected Score
	}{
		{100, Score{Value: 100, Grade: "A", Level: "excellent"}},
		{85, Score{Value: 85, Grade: "B", Level: "good"}},
		{65, Score{Value: 65, Grade: "D", Level: "fair"}},
		{30, Score{Value: 30, Grade: "F", Level: "poor"}},
		{0, Score{Value: 0, Grade: "F", Level: "critical"}},
		{-10, Score{Value: 0, Grade: "F", Level: "critical"}}, // Clamps to 0
		{150, Score{Value: 100, Grade: "A", Level: "excellent"}}, // Clamps to 100
	}

	for _, tt := range tests {
		got := NewScore(tt.value)
		if got.Value != tt.expected.Value || got.Grade != tt.expected.Grade || got.Level != tt.expected.Level {
			t.Errorf("NewScore(%d) = %+v, want %+v", tt.value, got, tt.expected)
		}
	}
}

func TestScore_IsExcellent(t *testing.T) {
	tests := []struct {
		value    int
		expected bool
	}{
		{100, true},
		{90, true},
		{89, false},
		{50, false},
	}

	for _, tt := range tests {
		s := NewScore(tt.value)
		if got := s.IsExcellent(); got != tt.expected {
			t.Errorf("Score{%d}.IsExcellent() = %v, want %v", tt.value, got, tt.expected)
		}
	}
}

func TestScore_IsGood(t *testing.T) {
	tests := []struct {
		value    int
		expected bool
	}{
		{100, true},
		{80, true},
		{79, false},
		{50, false},
	}

	for _, tt := range tests {
		s := NewScore(tt.value)
		if got := s.IsGood(); got != tt.expected {
			t.Errorf("Score{%d}.IsGood() = %v, want %v", tt.value, got, tt.expected)
		}
	}
}

func TestScore_IsPoor(t *testing.T) {
	tests := []struct {
		value    int
		expected bool
	}{
		{100, false},
		{70, false},
		{69, true},
		{50, true},
	}

	for _, tt := range tests {
		s := NewScore(tt.value)
		if got := s.IsPoor(); got != tt.expected {
			t.Errorf("Score{%d}.IsPoor() = %v, want %v", tt.value, got, tt.expected)
		}
	}
}

func TestScore_IsCritical(t *testing.T) {
	tests := []struct {
		value    int
		expected bool
	}{
		{100, false},
		{60, false},
		{59, true},
		{0, true},
	}

	for _, tt := range tests {
		s := NewScore(tt.value)
		if got := s.IsCritical(); got != tt.expected {
			t.Errorf("Score{%d}.IsCritical() = %v, want %v", tt.value, got, tt.expected)
		}
	}
}

func TestCalculate(t *testing.T) {
	tests := []struct {
		name       string
		components []ComponentScore
		expected   int
	}{
		{
			name:       "empty components",
			components: []ComponentScore{},
			expected:   0,
		},
		{
			name: "single component",
			components: []ComponentScore{
				{Name: "security", Value: 80, Weight: 1.0},
			},
			expected: 80,
		},
		{
			name: "equal weight components",
			components: []ComponentScore{
				{Name: "security", Value: 80, Weight: 1.0},
				{Name: "quality", Value: 60, Weight: 1.0},
			},
			expected: 70,
		},
		{
			name: "weighted components",
			components: []ComponentScore{
				{Name: "security", Value: 100, Weight: 2.0},
				{Name: "quality", Value: 50, Weight: 1.0},
			},
			expected: 83, // (100*2 + 50*1) / 3 = 250/3 = 83.33 -> 83
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Calculate(tt.components)
			if got.Value != tt.expected {
				t.Errorf("Calculate() = %d, want %d", got.Value, tt.expected)
			}
		})
	}
}

func TestNewComponentScore(t *testing.T) {
	cs := NewComponentScore("test", 75, 1.5)
	if cs.Name != "test" || cs.Value != 75 || cs.Weight != 1.5 {
		t.Errorf("NewComponentScore() = %+v, want Name=test, Value=75, Weight=1.5", cs)
	}
}

func TestComponentScore_WithMax(t *testing.T) {
	cs := NewComponentScore("test", 75, 1.0).WithMax(100)
	if cs.Max != 100 {
		t.Errorf("WithMax(100) = %d, want 100", cs.Max)
	}
}

func TestRiskFromCounts(t *testing.T) {
	tests := []struct {
		critical, high, medium int
		expected               RiskLevel
	}{
		{1, 0, 0, RiskCritical},
		{0, 3, 0, RiskHigh},
		{0, 1, 0, RiskMedium},
		{0, 0, 6, RiskMedium},
		{0, 0, 1, RiskLow},
		{0, 0, 0, RiskNone},
	}

	for _, tt := range tests {
		got := RiskFromCounts(tt.critical, tt.high, tt.medium)
		if got != tt.expected {
			t.Errorf("RiskFromCounts(%d, %d, %d) = %s, want %s",
				tt.critical, tt.high, tt.medium, got, tt.expected)
		}
	}
}

func TestRiskFromFindings(t *testing.T) {
	tests := []struct {
		name     string
		findings []findings.Finding
		expected RiskLevel
	}{
		{
			name:     "no findings",
			findings: []findings.Finding{},
			expected: RiskNone,
		},
		{
			name: "critical finding",
			findings: []findings.Finding{
				{Severity: findings.SeverityCritical},
			},
			expected: RiskCritical,
		},
		{
			name: "multiple high findings",
			findings: []findings.Finding{
				{Severity: findings.SeverityHigh},
				{Severity: findings.SeverityHigh},
				{Severity: findings.SeverityHigh},
			},
			expected: RiskHigh,
		},
		{
			name: "medium findings",
			findings: []findings.Finding{
				{Severity: findings.SeverityMedium},
			},
			expected: RiskLow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RiskFromFindings(tt.findings)
			if got != tt.expected {
				t.Errorf("RiskFromFindings() = %s, want %s", got, tt.expected)
			}
		})
	}
}

func TestSecurityScore(t *testing.T) {
	tests := []struct {
		name     string
		findings []findings.Finding
		expected int
	}{
		{
			name:     "no findings",
			findings: []findings.Finding{},
			expected: 100,
		},
		{
			name: "one critical",
			findings: []findings.Finding{
				{Severity: findings.SeverityCritical},
			},
			expected: 75,
		},
		{
			name: "one high",
			findings: []findings.Finding{
				{Severity: findings.SeverityHigh},
			},
			expected: 85,
		},
		{
			name: "one medium",
			findings: []findings.Finding{
				{Severity: findings.SeverityMedium},
			},
			expected: 95,
		},
		{
			name: "one low",
			findings: []findings.Finding{
				{Severity: findings.SeverityLow},
			},
			expected: 99,
		},
		{
			name: "mixed severities",
			findings: []findings.Finding{
				{Severity: findings.SeverityCritical},
				{Severity: findings.SeverityHigh},
				{Severity: findings.SeverityMedium},
			},
			expected: 55, // 100 - 25 - 15 - 5
		},
		{
			name: "many findings floor to 0",
			findings: []findings.Finding{
				{Severity: findings.SeverityCritical},
				{Severity: findings.SeverityCritical},
				{Severity: findings.SeverityCritical},
				{Severity: findings.SeverityCritical},
				{Severity: findings.SeverityCritical},
			},
			expected: 0, // 100 - 125 -> 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SecurityScore(tt.findings)
			if got.Value != tt.expected {
				t.Errorf("SecurityScore() = %d, want %d", got.Value, tt.expected)
			}
		})
	}
}

func TestSecurityScoreWithWeights(t *testing.T) {
	weights := map[findings.Severity]int{
		findings.SeverityCritical: 50,
		findings.SeverityHigh:     30,
		findings.SeverityMedium:   10,
		findings.SeverityLow:      2,
	}

	tests := []struct {
		name     string
		findings []findings.Finding
		expected int
	}{
		{
			name:     "no findings",
			findings: []findings.Finding{},
			expected: 100,
		},
		{
			name: "one critical",
			findings: []findings.Finding{
				{Severity: findings.SeverityCritical},
			},
			expected: 50,
		},
		{
			name: "mixed",
			findings: []findings.Finding{
				{Severity: findings.SeverityHigh},
				{Severity: findings.SeverityMedium},
			},
			expected: 60, // 100 - 30 - 10
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SecurityScoreWithWeights(tt.findings, weights)
			if got.Value != tt.expected {
				t.Errorf("SecurityScoreWithWeights() = %d, want %d", got.Value, tt.expected)
			}
		})
	}
}

func TestHealthScore(t *testing.T) {
	weights := map[string]int{
		"has_tests":   30,
		"has_docs":    20,
		"has_ci":      25,
		"has_license": 25,
	}

	tests := []struct {
		name       string
		indicators map[string]bool
		expected   int
	}{
		{
			name:       "all true",
			indicators: map[string]bool{"has_tests": true, "has_docs": true, "has_ci": true, "has_license": true},
			expected:   100,
		},
		{
			name:       "all false",
			indicators: map[string]bool{"has_tests": false, "has_docs": false, "has_ci": false, "has_license": false},
			expected:   0,
		},
		{
			name:       "half",
			indicators: map[string]bool{"has_tests": true, "has_docs": true, "has_ci": false, "has_license": false},
			expected:   50, // (30+20) / 100
		},
		{
			name:       "empty weights",
			indicators: map[string]bool{},
			expected:   100, // maxPossible = 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := weights
			if tt.name == "empty weights" {
				w = map[string]int{}
			}
			got := HealthScore(tt.indicators, w)
			if got.Value != tt.expected {
				t.Errorf("HealthScore() = %d, want %d", got.Value, tt.expected)
			}
		})
	}
}

func TestCombineScores(t *testing.T) {
	tests := []struct {
		name     string
		scores   []Score
		expected int
	}{
		{
			name:     "empty",
			scores:   []Score{},
			expected: 0,
		},
		{
			name:     "single",
			scores:   []Score{NewScore(80)},
			expected: 80,
		},
		{
			name:     "two equal",
			scores:   []Score{NewScore(80), NewScore(60)},
			expected: 70,
		},
		{
			name:     "three",
			scores:   []Score{NewScore(90), NewScore(80), NewScore(70)},
			expected: 80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CombineScores(tt.scores...)
			if got.Value != tt.expected {
				t.Errorf("CombineScores() = %d, want %d", got.Value, tt.expected)
			}
		})
	}
}

func TestWeightedCombine(t *testing.T) {
	tests := []struct {
		name     string
		scores   []Score
		weights  []float64
		expected int
	}{
		{
			name:     "empty",
			scores:   []Score{},
			weights:  []float64{},
			expected: 0,
		},
		{
			name:     "mismatched lengths",
			scores:   []Score{NewScore(80)},
			weights:  []float64{1.0, 2.0},
			expected: 0,
		},
		{
			name:     "equal weights",
			scores:   []Score{NewScore(80), NewScore(60)},
			weights:  []float64{1.0, 1.0},
			expected: 70,
		},
		{
			name:     "weighted",
			scores:   []Score{NewScore(100), NewScore(50)},
			weights:  []float64{2.0, 1.0},
			expected: 83, // (100*2 + 50*1) / 3
		},
		{
			name:     "zero weights",
			scores:   []Score{NewScore(80)},
			weights:  []float64{0.0},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WeightedCombine(tt.scores, tt.weights)
			if got.Value != tt.expected {
				t.Errorf("WeightedCombine() = %d, want %d", got.Value, tt.expected)
			}
		})
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		value, min, max int
		expected        int
	}{
		{50, 0, 100, 50},
		{0, 0, 100, 0},
		{100, 0, 100, 100},
		{150, 0, 100, 100}, // Above max
		{-10, 0, 100, 0},   // Below min
		{50, 50, 50, 0},    // max <= min
		{75, 50, 100, 50},  // 75 is 50% between 50 and 100
	}

	for _, tt := range tests {
		got := Normalize(tt.value, tt.min, tt.max)
		if got != tt.expected {
			t.Errorf("Normalize(%d, %d, %d) = %d, want %d",
				tt.value, tt.min, tt.max, got, tt.expected)
		}
	}
}

func TestInverseNormalize(t *testing.T) {
	tests := []struct {
		value, min, max int
		expected        int
	}{
		{0, 0, 100, 100},   // 0 issues = 100 score
		{100, 0, 100, 0},   // max issues = 0 score
		{50, 0, 100, 50},   // half issues = 50 score
	}

	for _, tt := range tests {
		got := InverseNormalize(tt.value, tt.min, tt.max)
		if got != tt.expected {
			t.Errorf("InverseNormalize(%d, %d, %d) = %d, want %d",
				tt.value, tt.min, tt.max, got, tt.expected)
		}
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		value    int
		expected int
	}{
		{50, 50},
		{0, 0},
		{100, 100},
		{-10, 0},
		{150, 100},
	}

	for _, tt := range tests {
		got := Clamp(tt.value)
		if got != tt.expected {
			t.Errorf("Clamp(%d) = %d, want %d", tt.value, got, tt.expected)
		}
	}
}

func TestRiskLevel_String(t *testing.T) {
	tests := []struct {
		level    RiskLevel
		expected string
	}{
		{RiskCritical, "critical"},
		{RiskHigh, "high"},
		{RiskMedium, "medium"},
		{RiskLow, "low"},
		{RiskNone, "none"},
	}

	for _, tt := range tests {
		got := tt.level.String()
		if got != tt.expected {
			t.Errorf("RiskLevel(%s).String() = %s, want %s", tt.level, got, tt.expected)
		}
	}
}

func TestRiskLevel_Score(t *testing.T) {
	tests := []struct {
		level    RiskLevel
		expected int
	}{
		{RiskCritical, 5},
		{RiskHigh, 4},
		{RiskMedium, 3},
		{RiskLow, 2},
		{RiskNone, 1},
		{RiskLevel("unknown"), 0},
	}

	for _, tt := range tests {
		got := tt.level.Score()
		if got != tt.expected {
			t.Errorf("RiskLevel(%s).Score() = %d, want %d", tt.level, got, tt.expected)
		}
	}
}

func TestRiskLevel_IsHighRisk(t *testing.T) {
	tests := []struct {
		level    RiskLevel
		expected bool
	}{
		{RiskCritical, true},
		{RiskHigh, true},
		{RiskMedium, false},
		{RiskLow, false},
		{RiskNone, false},
	}

	for _, tt := range tests {
		got := tt.level.IsHighRisk()
		if got != tt.expected {
			t.Errorf("RiskLevel(%s).IsHighRisk() = %v, want %v", tt.level, got, tt.expected)
		}
	}
}
