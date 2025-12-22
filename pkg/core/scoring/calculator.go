package scoring

import "github.com/crashappsec/zero/pkg/core/findings"

// Calculate computes weighted score from components
func Calculate(components []ComponentScore) Score {
	if len(components) == 0 {
		return Score{Value: 0, Grade: "F", Level: "critical"}
	}

	var totalWeight float64
	var weightedSum float64

	for _, c := range components {
		totalWeight += c.Weight
		weightedSum += float64(c.Value) * c.Weight
	}

	value := 0
	if totalWeight > 0 {
		value = int(weightedSum / totalWeight)
	}

	return Score{
		Value:      value,
		Grade:      ValueToGrade(value),
		Level:      ValueToLevel(value),
		Components: components,
	}
}

// ValueToGrade converts a numeric score to letter grade
func ValueToGrade(value int) string {
	switch {
	case value >= 90:
		return "A"
	case value >= 80:
		return "B"
	case value >= 70:
		return "C"
	case value >= 60:
		return "D"
	default:
		return "F"
	}
}

// ValueToLevel converts a numeric score to a level string
func ValueToLevel(value int) string {
	switch {
	case value >= 90:
		return "excellent"
	case value >= 75:
		return "good"
	case value >= 50:
		return "fair"
	case value >= 25:
		return "poor"
	default:
		return "critical"
	}
}

// RiskFromFindings calculates risk level from findings
func RiskFromFindings(fs []findings.Finding) RiskLevel {
	var critical, high, medium int

	for _, f := range fs {
		switch f.Severity {
		case findings.SeverityCritical:
			critical++
		case findings.SeverityHigh:
			high++
		case findings.SeverityMedium:
			medium++
		}
	}

	return RiskFromCounts(critical, high, medium)
}

// RiskFromCounts calculates risk level from severity counts
func RiskFromCounts(critical, high, medium int) RiskLevel {
	switch {
	case critical > 0:
		return RiskCritical
	case high > 2:
		return RiskHigh
	case high > 0 || medium > 5:
		return RiskMedium
	case medium > 0:
		return RiskLow
	default:
		return RiskNone
	}
}

// SecurityScore calculates security score from findings
// Starts at 100 and deducts based on severity
func SecurityScore(fs []findings.Finding) Score {
	score := 100

	for _, f := range fs {
		switch f.Severity {
		case findings.SeverityCritical:
			score -= 25
		case findings.SeverityHigh:
			score -= 15
		case findings.SeverityMedium:
			score -= 5
		case findings.SeverityLow:
			score -= 1
		}
	}

	if score < 0 {
		score = 0
	}

	return NewScore(score)
}

// SecurityScoreWithWeights calculates security score with custom weights
func SecurityScoreWithWeights(fs []findings.Finding, weights map[findings.Severity]int) Score {
	score := 100

	for _, f := range fs {
		if deduct, ok := weights[f.Severity]; ok {
			score -= deduct
		}
	}

	if score < 0 {
		score = 0
	}

	return NewScore(score)
}

// HealthScore calculates a health score from positive indicators
// Each indicator adds to the score up to 100
func HealthScore(indicators map[string]bool, weights map[string]int) Score {
	var total int
	var maxPossible int

	for name, weight := range weights {
		maxPossible += weight
		if indicators[name] {
			total += weight
		}
	}

	if maxPossible == 0 {
		return NewScore(100)
	}

	value := (total * 100) / maxPossible
	return NewScore(value)
}

// CombineScores combines multiple scores with equal weight
func CombineScores(scores ...Score) Score {
	if len(scores) == 0 {
		return NewScore(0)
	}

	total := 0
	for _, s := range scores {
		total += s.Value
	}

	return NewScore(total / len(scores))
}

// WeightedCombine combines scores with custom weights
func WeightedCombine(scores []Score, weights []float64) Score {
	if len(scores) == 0 || len(scores) != len(weights) {
		return NewScore(0)
	}

	var totalWeight float64
	var weightedSum float64

	for i, s := range scores {
		totalWeight += weights[i]
		weightedSum += float64(s.Value) * weights[i]
	}

	if totalWeight == 0 {
		return NewScore(0)
	}

	return NewScore(int(weightedSum / totalWeight))
}

// Normalize normalizes a value to 0-100 range
func Normalize(value, min, max int) int {
	if max <= min {
		return 0
	}
	if value <= min {
		return 0
	}
	if value >= max {
		return 100
	}
	return ((value - min) * 100) / (max - min)
}

// InverseNormalize normalizes a value where lower is better
func InverseNormalize(value, min, max int) int {
	return 100 - Normalize(value, min, max)
}

// Clamp ensures a value is within 0-100 range
func Clamp(value int) int {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}
