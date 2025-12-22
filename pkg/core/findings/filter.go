package findings

import "sort"

// FilterOptions configures finding filtering
type FilterOptions struct {
	MinSeverity  Severity
	MaxSeverity  Severity
	Categories   []string
	Scanners     []string
	IncludeInfo  bool
	Limit        int
	HasLocation  *bool // nil = don't filter, true = must have, false = must not have
}

// Filter returns findings matching the options
func Filter(findings []Finding, opts FilterOptions) []Finding {
	var result []Finding

	for _, f := range findings {
		// Severity filter
		if opts.MinSeverity != "" && f.Severity.Score() < opts.MinSeverity.Score() {
			continue
		}
		if opts.MaxSeverity != "" && f.Severity.Score() > opts.MaxSeverity.Score() {
			continue
		}

		// Skip info unless explicitly included
		if !opts.IncludeInfo && f.Severity == SeverityInfo {
			continue
		}

		// Category filter
		if len(opts.Categories) > 0 && !contains(opts.Categories, f.Category) {
			continue
		}

		// Scanner filter
		if len(opts.Scanners) > 0 && !contains(opts.Scanners, f.Scanner) {
			continue
		}

		// Location filter
		if opts.HasLocation != nil {
			hasLoc := f.Location != nil
			if *opts.HasLocation != hasLoc {
				continue
			}
		}

		result = append(result, f)

		if opts.Limit > 0 && len(result) >= opts.Limit {
			break
		}
	}

	return result
}

// SortBySeverity sorts findings by severity (critical first)
func SortBySeverity(findings []Finding) []Finding {
	sorted := make([]Finding, len(findings))
	copy(sorted, findings)
	sort.Slice(sorted, func(i, j int) bool {
		// Higher score = more severe = should come first
		if sorted[i].Severity.Score() != sorted[j].Severity.Score() {
			return sorted[i].Severity.Score() > sorted[j].Severity.Score()
		}
		// Secondary sort by confidence
		return sorted[i].Confidence.Score() > sorted[j].Confidence.Score()
	})
	return sorted
}

// SortByCategory sorts findings by category alphabetically
func SortByCategory(findings []Finding) []Finding {
	sorted := make([]Finding, len(findings))
	copy(sorted, findings)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Category != sorted[j].Category {
			return sorted[i].Category < sorted[j].Category
		}
		return sorted[i].Severity.Score() > sorted[j].Severity.Score()
	})
	return sorted
}

// SortByFile sorts findings by file location
func SortByFile(findings []Finding) []Finding {
	sorted := make([]Finding, len(findings))
	copy(sorted, findings)
	sort.Slice(sorted, func(i, j int) bool {
		// Findings without location come last
		if sorted[i].Location == nil && sorted[j].Location == nil {
			return false
		}
		if sorted[i].Location == nil {
			return false
		}
		if sorted[j].Location == nil {
			return true
		}
		// Sort by file, then line
		if sorted[i].Location.File != sorted[j].Location.File {
			return sorted[i].Location.File < sorted[j].Location.File
		}
		return sorted[i].Location.Line < sorted[j].Location.Line
	})
	return sorted
}

// GroupByCategory groups findings by category
func GroupByCategory(findings []Finding) map[string][]Finding {
	groups := make(map[string][]Finding)
	for _, f := range findings {
		groups[f.Category] = append(groups[f.Category], f)
	}
	return groups
}

// GroupBySeverity groups findings by severity
func GroupBySeverity(findings []Finding) map[Severity][]Finding {
	groups := make(map[Severity][]Finding)
	for _, f := range findings {
		groups[f.Severity] = append(groups[f.Severity], f)
	}
	return groups
}

// GroupByScanner groups findings by scanner
func GroupByScanner(findings []Finding) map[string][]Finding {
	groups := make(map[string][]Finding)
	for _, f := range findings {
		groups[f.Scanner] = append(groups[f.Scanner], f)
	}
	return groups
}

// GroupByFile groups findings by file location
func GroupByFile(findings []Finding) map[string][]Finding {
	groups := make(map[string][]Finding)
	for _, f := range findings {
		file := ""
		if f.Location != nil {
			file = f.Location.File
		}
		groups[file] = append(groups[file], f)
	}
	return groups
}

// Deduplicate removes duplicate findings based on ID
func Deduplicate(findings []Finding) []Finding {
	seen := make(map[string]bool)
	var result []Finding
	for _, f := range findings {
		if !seen[f.ID] {
			seen[f.ID] = true
			result = append(result, f)
		}
	}
	return result
}

// DeduplicateByContent removes findings with same title, category, and location
func DeduplicateByContent(findings []Finding) []Finding {
	seen := make(map[string]bool)
	var result []Finding
	for _, f := range findings {
		key := f.Title + "|" + f.Category
		if f.Location != nil {
			key += "|" + f.Location.File
		}
		if !seen[key] {
			seen[key] = true
			result = append(result, f)
		}
	}
	return result
}

// CountBySeverity returns counts by severity level
func CountBySeverity(findings []Finding) map[Severity]int {
	counts := make(map[Severity]int)
	for _, f := range findings {
		counts[f.Severity]++
	}
	return counts
}

// TopN returns the top N findings by severity
func TopN(findings []Finding, n int) []Finding {
	sorted := SortBySeverity(findings)
	if len(sorted) <= n {
		return sorted
	}
	return sorted[:n]
}

// CriticalAndHigh returns only critical and high severity findings
func CriticalAndHigh(findings []Finding) []Finding {
	return Filter(findings, FilterOptions{
		MinSeverity: SeverityHigh,
		IncludeInfo: false,
	})
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
