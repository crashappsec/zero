// Package codeownership provides code ownership and CODEOWNERS analysis
package codeownership

import (
	"math"
	"sort"
	"time"
)

// OwnershipScorer calculates weighted ownership scores
type OwnershipScorer struct {
	weights ScoringWeights
}

// NewOwnershipScorer creates a new scorer with the given weights
func NewOwnershipScorer(weights ScoringWeights) *OwnershipScorer {
	return &OwnershipScorer{weights: weights}
}

// ContributorData holds raw data for scoring a contributor
type ContributorData struct {
	Name         string
	Email        string
	Commits      int
	LinesAdded   int
	LinesRemoved int
	LastCommit   time.Time
	CommitDates  []time.Time // All commit dates for consistency calculation
	PRReviews    int         // Number of PR reviews given (from GitHub API)
}

// CalculateEnhancedOwnership calculates ownership scores for all contributors
func (s *OwnershipScorer) CalculateEnhancedOwnership(contributors []ContributorData, now time.Time) []EnhancedOwnership {
	if len(contributors) == 0 {
		return nil
	}

	// Calculate maximums for normalization
	var maxCommits, maxLines, maxReviews int
	for _, c := range contributors {
		if c.Commits > maxCommits {
			maxCommits = c.Commits
		}
		lines := c.LinesAdded + c.LinesRemoved
		if lines > maxLines {
			maxLines = lines
		}
		if c.PRReviews > maxReviews {
			maxReviews = c.PRReviews
		}
	}

	// Avoid division by zero
	if maxCommits == 0 {
		maxCommits = 1
	}
	if maxLines == 0 {
		maxLines = 1
	}
	if maxReviews == 0 {
		maxReviews = 1
	}

	results := make([]EnhancedOwnership, 0, len(contributors))
	for _, c := range contributors {
		enhanced := s.calculateForContributor(c, maxCommits, maxLines, maxReviews, now)
		results = append(results, enhanced)
	}

	// Sort by ownership score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].OwnershipScore > results[j].OwnershipScore
	})

	return results
}

// calculateForContributor calculates the enhanced ownership for a single contributor
func (s *OwnershipScorer) calculateForContributor(c ContributorData, maxCommits, maxLines, maxReviews int, now time.Time) EnhancedOwnership {
	breakdown := ScoreBreakdown{}

	// 1. Commit score (0-30 by default)
	commitRatio := float64(c.Commits) / float64(maxCommits)
	breakdown.CommitScore = commitRatio * s.weights.Commits * 100

	// 2. Review score (0-25 by default)
	reviewRatio := float64(c.PRReviews) / float64(maxReviews)
	breakdown.ReviewScore = reviewRatio * s.weights.Reviews * 100

	// 3. Lines score (0-20 by default)
	lines := c.LinesAdded + c.LinesRemoved
	linesRatio := float64(lines) / float64(maxLines)
	breakdown.LinesScore = linesRatio * s.weights.Lines * 100

	// 4. Recency score (0-15 by default) - exponential decay with 90-day half-life
	daysSinceLastCommit := now.Sub(c.LastCommit).Hours() / 24
	halfLife := 90.0 // days
	recencyDecay := math.Pow(0.5, daysSinceLastCommit/halfLife)
	breakdown.RecencyScore = recencyDecay * s.weights.Recency * 100

	// 5. Consistency score (0-10 by default)
	consistency := s.calculateConsistency(c.CommitDates, now)
	breakdown.ConsistencyScore = consistency * s.weights.Consistency * 100

	// Total score
	totalScore := breakdown.CommitScore + breakdown.ReviewScore +
		breakdown.LinesScore + breakdown.RecencyScore + breakdown.ConsistencyScore

	// Activity status based on days since last commit
	activityStatus := s.determineActivityStatus(daysSinceLastCommit)

	// Confidence based on data completeness
	confidence := s.calculateConfidence(c, maxReviews)

	return EnhancedOwnership{
		Name:            c.Name,
		Email:           c.Email,
		OwnershipScore:  math.Round(totalScore*100) / 100, // Round to 2 decimal places
		ScoreBreakdown:  breakdown,
		ActivityStatus:  activityStatus,
		LastActive:      c.LastCommit.Format(time.RFC3339),
		Confidence:      confidence,
		PRReviewsGiven:  c.PRReviews,
		PRReviewsOnCode: 0, // This would be populated from GitHub API data
	}
}

// calculateConsistency measures how regularly someone contributes
// Returns 0-1 where 1 means very consistent (commits spread evenly)
func (s *OwnershipScorer) calculateConsistency(commitDates []time.Time, _ time.Time) float64 {
	if len(commitDates) < 2 {
		return 0.5 // Not enough data, neutral score
	}

	// Sort dates
	sorted := make([]time.Time, len(commitDates))
	copy(sorted, commitDates)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Before(sorted[j])
	})

	// Calculate coefficient of variation of gaps between commits
	// Lower CV = more consistent
	var gaps []float64
	for i := 1; i < len(sorted); i++ {
		gap := sorted[i].Sub(sorted[i-1]).Hours() / 24 // Gap in days
		gaps = append(gaps, gap)
	}

	if len(gaps) == 0 {
		return 0.5
	}

	// Calculate mean and standard deviation
	var sum float64
	for _, g := range gaps {
		sum += g
	}
	mean := sum / float64(len(gaps))

	if mean == 0 {
		return 1.0 // All commits on same day = consistent
	}

	var varianceSum float64
	for _, g := range gaps {
		varianceSum += (g - mean) * (g - mean)
	}
	stdDev := math.Sqrt(varianceSum / float64(len(gaps)))

	// Coefficient of variation (CV)
	cv := stdDev / mean

	// Convert CV to 0-1 score (lower CV = higher score)
	// CV of 0 = perfect consistency (score 1)
	// CV of 2 or more = very inconsistent (score 0)
	consistency := math.Max(0, 1-(cv/2))

	return consistency
}

// determineActivityStatus returns the activity status based on days since last commit
func (s *OwnershipScorer) determineActivityStatus(daysSinceLastCommit float64) string {
	switch {
	case daysSinceLastCommit <= float64(ActivityThresholds.Active):
		return "active"
	case daysSinceLastCommit <= float64(ActivityThresholds.Recent):
		return "recent"
	case daysSinceLastCommit <= float64(ActivityThresholds.Stale):
		return "stale"
	case daysSinceLastCommit <= float64(ActivityThresholds.Inactive):
		return "inactive"
	default:
		return "abandoned"
	}
}

// calculateConfidence returns 0-1 indicating data quality
func (s *OwnershipScorer) calculateConfidence(c ContributorData, maxReviews int) float64 {
	confidence := 0.5 // Base confidence

	// More commits = more confident
	if c.Commits >= 10 {
		confidence += 0.2
	} else if c.Commits >= 5 {
		confidence += 0.1
	}

	// Have PR review data = more confident
	if maxReviews > 0 {
		confidence += 0.2
	}

	// Have multiple commit dates for consistency calc
	if len(c.CommitDates) >= 5 {
		confidence += 0.1
	}

	return math.Min(1.0, confidence)
}

// CalculateBusFactor determines the number of people who need to leave
// before critical knowledge is lost
func CalculateBusFactor(owners []EnhancedOwnership, threshold float64) (int, string) {
	if len(owners) == 0 {
		return 0, "critical"
	}

	// Calculate total ownership
	var totalScore float64
	for _, o := range owners {
		totalScore += o.OwnershipScore
	}

	if totalScore == 0 {
		return 0, "critical"
	}

	// Bus factor is the minimum number of people whose combined
	// ownership exceeds the threshold (default 50%)
	var cumulative float64
	for i, o := range owners {
		cumulative += o.OwnershipScore
		if cumulative/totalScore >= threshold {
			busFactor := i + 1

			// Determine risk level
			var risk string
			switch {
			case busFactor <= BusFactorThresholds.Critical:
				risk = "critical"
			case busFactor <= BusFactorThresholds.Warning:
				risk = "warning"
			default:
				risk = "healthy"
			}

			return busFactor, risk
		}
	}

	return len(owners), "healthy"
}

// CalculateOwnershipCoverage determines what percentage of files have clear owners
func CalculateOwnershipCoverage(files []FileOwnership, minContributors int) float64 {
	if len(files) == 0 {
		return 1.0 // No files = full coverage
	}

	covered := 0
	for _, f := range files {
		if len(f.TopContributors) >= minContributors {
			covered++
		}
	}

	return float64(covered) / float64(len(files))
}
