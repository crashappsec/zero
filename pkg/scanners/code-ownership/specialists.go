// Package codeownership provides code ownership and CODEOWNERS analysis
package codeownership

import (
	"path/filepath"
	"sort"
	"strings"
)

// SpecialistAnalyzer identifies domain specialists across repositories
type SpecialistAnalyzer struct {
	domains []string
}

// NewSpecialistAnalyzer creates a new specialist analyzer
func NewSpecialistAnalyzer(domains []string) *SpecialistAnalyzer {
	if len(domains) == 0 {
		domains = DefaultEnhancedConfig().SpecialistDomains
	}
	return &SpecialistAnalyzer{domains: domains}
}

// DeveloperContribution tracks a developer's work in a repo
type DeveloperContribution struct {
	Name        string
	Email       string
	FilesMap    map[string]int // file path -> commit count
	TotalCommits int
}

// AnalyzeSpecialists identifies domain specialists from contribution data
func (a *SpecialistAnalyzer) AnalyzeSpecialists(contributions []DeveloperContribution) []OrgSpecialist {
	// Build domain expertise for each contributor
	specialistMap := make(map[string]*OrgSpecialist)

	for _, contrib := range contributions {
		key := contrib.Email
		if key == "" {
			key = contrib.Name
		}

		if _, exists := specialistMap[key]; !exists {
			specialistMap[key] = &OrgSpecialist{
				Name:    contrib.Name,
				Email:   contrib.Email,
				Domains: make([]DomainExpertise, 0),
			}
		}

		specialist := specialistMap[key]
		specialist.ReposActive++

		// Analyze files for domain patterns
		domainCounts := a.classifyFiles(contrib.FilesMap)

		// Update domain expertise
		for domain, fileCount := range domainCounts {
			found := false
			for i := range specialist.Domains {
				if specialist.Domains[i].Domain == domain {
					specialist.Domains[i].FileCount += fileCount
					specialist.Domains[i].RepoCount++
					found = true
					break
				}
			}
			if !found {
				specialist.Domains = append(specialist.Domains, DomainExpertise{
					Domain:    domain,
					FileCount: fileCount,
					RepoCount: 1,
				})
			}
		}
	}

	// Calculate scores and determine top domains
	specialists := make([]OrgSpecialist, 0, len(specialistMap))
	for _, s := range specialistMap {
		a.calculateScores(s)
		specialists = append(specialists, *s)
	}

	// Sort by total score
	sort.Slice(specialists, func(i, j int) bool {
		return specialists[i].TotalScore > specialists[j].TotalScore
	})

	return specialists
}

// classifyFiles determines which domains files belong to
func (a *SpecialistAnalyzer) classifyFiles(filesMap map[string]int) map[string]int {
	domainCounts := make(map[string]int)

	for filePath, commits := range filesMap {
		for _, domain := range a.domains {
			if a.matchesDomain(filePath, domain) {
				domainCounts[domain] += commits
			}
		}
	}

	return domainCounts
}

// matchesDomain checks if a file path matches a domain's patterns
func (a *SpecialistAnalyzer) matchesDomain(filePath, domain string) bool {
	patterns, ok := DomainPatterns[domain]
	if !ok {
		return false
	}

	// Normalize path
	filePath = strings.ToLower(filePath)

	for _, pattern := range patterns {
		if a.matchPattern(filePath, strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}

// matchPattern matches a file path against a glob-like pattern
func (a *SpecialistAnalyzer) matchPattern(filePath, pattern string) bool {
	// Direct extension match (e.g., "*.tsx")
	if strings.HasPrefix(pattern, "*") && !strings.Contains(pattern, "/") {
		ext := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(filePath, ext)
	}

	// Directory pattern (e.g., "auth/*" or "**/auth/**")
	if strings.Contains(pattern, "**") {
		// Remove ** and check if any part matches
		parts := strings.Split(pattern, "**")
		for _, part := range parts {
			part = strings.Trim(part, "/")
			if part != "" && strings.Contains(filePath, part) {
				return true
			}
		}
		return false
	}

	// Simple directory match (e.g., "auth/")
	if strings.HasSuffix(pattern, "/") || strings.HasSuffix(pattern, "/*") {
		dir := strings.TrimSuffix(strings.TrimSuffix(pattern, "/*"), "/")
		return strings.HasPrefix(filePath, dir+"/") || filePath == dir
	}

	// Exact or partial match
	matched, _ := filepath.Match(pattern, filePath)
	if matched {
		return true
	}

	// Check if the pattern matches any part of the path
	return strings.Contains(filePath, strings.Trim(pattern, "*"))
}

// calculateScores calculates domain scores and total score for a specialist
func (a *SpecialistAnalyzer) calculateScores(s *OrgSpecialist) {
	if len(s.Domains) == 0 {
		return
	}

	// Calculate total files across all domains
	var totalFiles int
	for _, d := range s.Domains {
		totalFiles += d.FileCount
	}

	if totalFiles == 0 {
		return
	}

	// Calculate scores for each domain
	var maxScore float64
	for i := range s.Domains {
		d := &s.Domains[i]

		// Score based on file count and repo spread
		fileScore := float64(d.FileCount) / float64(totalFiles) * 50
		repoScore := float64(d.RepoCount) / float64(s.ReposActive) * 50

		d.Score = fileScore + repoScore

		// Confidence based on data volume
		if d.FileCount >= 50 {
			d.Confidence = 1.0
		} else if d.FileCount >= 20 {
			d.Confidence = 0.8
		} else if d.FileCount >= 10 {
			d.Confidence = 0.6
		} else {
			d.Confidence = 0.4
		}

		if d.Score > maxScore {
			maxScore = d.Score
			s.TopDomain = d.Domain
		}

		s.TotalScore += d.Score
	}

	// Normalize total score to 0-100
	if len(s.Domains) > 0 {
		s.TotalScore = s.TotalScore / float64(len(s.Domains))
	}
}

// GetTopSpecialistsForDomain returns the top N specialists for a specific domain
func (a *SpecialistAnalyzer) GetTopSpecialistsForDomain(specialists []OrgSpecialist, domain string, limit int) []OrgSpecialist {
	var domainSpecialists []OrgSpecialist

	for _, s := range specialists {
		for _, d := range s.Domains {
			if d.Domain == domain && d.Score > 0 {
				// Create a copy with only this domain
				specialist := OrgSpecialist{
					Name:        s.Name,
					Email:       s.Email,
					TopDomain:   domain,
					TotalScore:  d.Score,
					ReposActive: d.RepoCount,
					Domains:     []DomainExpertise{d},
				}
				domainSpecialists = append(domainSpecialists, specialist)
				break
			}
		}
	}

	// Sort by score for this domain
	sort.Slice(domainSpecialists, func(i, j int) bool {
		return domainSpecialists[i].TotalScore > domainSpecialists[j].TotalScore
	})

	if limit > 0 && len(domainSpecialists) > limit {
		return domainSpecialists[:limit]
	}

	return domainSpecialists
}

// SuggestOwnersForPath suggests potential owners for a new file/path
func (a *SpecialistAnalyzer) SuggestOwnersForPath(filePath string, specialists []OrgSpecialist, limit int) []OrgSpecialist {
	// Determine which domain this path belongs to
	var matchedDomains []string
	for _, domain := range a.domains {
		if a.matchesDomain(filePath, domain) {
			matchedDomains = append(matchedDomains, domain)
		}
	}

	if len(matchedDomains) == 0 {
		// No specific domain, return top overall specialists
		if limit > 0 && len(specialists) > limit {
			return specialists[:limit]
		}
		return specialists
	}

	// Score specialists based on domain match
	type scoredSpecialist struct {
		specialist OrgSpecialist
		score      float64
	}

	var scored []scoredSpecialist
	for _, s := range specialists {
		var totalScore float64
		for _, d := range s.Domains {
			for _, matchedDomain := range matchedDomains {
				if d.Domain == matchedDomain {
					totalScore += d.Score
				}
			}
		}
		if totalScore > 0 {
			scored = append(scored, scoredSpecialist{
				specialist: s,
				score:      totalScore,
			})
		}
	}

	// Sort by score
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Extract specialists
	result := make([]OrgSpecialist, 0, limit)
	for i, s := range scored {
		if limit > 0 && i >= limit {
			break
		}
		result = append(result, s.specialist)
	}

	return result
}
