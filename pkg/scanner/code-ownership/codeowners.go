// Package codeownership provides code ownership and CODEOWNERS analysis
package codeownership

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// CODEOWNERSAnalyzer validates and analyzes CODEOWNERS files
type CODEOWNERSAnalyzer struct {
	config            CODEOWNERSConfig
	sensitivePatterns []string
}

// NewCODEOWNERSAnalyzer creates a new analyzer
func NewCODEOWNERSAnalyzer(config CODEOWNERSConfig) *CODEOWNERSAnalyzer {
	return &CODEOWNERSAnalyzer{
		config:            config,
		sensitivePatterns: config.SensitivePatterns,
	}
}

// ValidationRuleID defines CODEOWNERS validation rule IDs
type ValidationRuleID string

const (
	// Syntax rules
	RuleInvalidPattern ValidationRuleID = "CO001"
	RuleInvalidOwner   ValidationRuleID = "CO002"
	RuleEmptyFile      ValidationRuleID = "CO003"

	// Permission rules
	RuleOwnerNotExist     ValidationRuleID = "CO010"
	RuleOwnerNoAccess     ValidationRuleID = "CO011"
	RuleEmptyTeam         ValidationRuleID = "CO012"
	RuleUserNotCollaborator ValidationRuleID = "CO013"

	// Coverage rules
	RuleNoDefaultOwner ValidationRuleID = "CO020"
	RuleLowCoverage    ValidationRuleID = "CO021"

	// Staleness rules
	RuleOwnerInactive ValidationRuleID = "CO030"
	RuleOwnerAbandoned ValidationRuleID = "CO031"

	// Best practice rules
	RuleIndividualNotTeam      ValidationRuleID = "CO040"
	RuleSingleOwner           ValidationRuleID = "CO041"
	RuleTooManyOwners         ValidationRuleID = "CO042"
	RuleSensitiveUnprotected  ValidationRuleID = "CO043"
	RuleOverlappingPatterns   ValidationRuleID = "CO044"
)

// Analyze performs full CODEOWNERS analysis
func (a *CODEOWNERSAnalyzer) Analyze(repoPath string, contributors []Contributor) (*CODEOWNERSAnalysis, error) {
	// Find CODEOWNERS file
	codeownersPath := a.findCodeownersFile(repoPath)
	if codeownersPath == "" {
		return &CODEOWNERSAnalysis{
			FilePath:   "",
			RulesCount: 0,
			Coverage:   0,
			ValidationIssues: []CODEOWNERSIssue{{
				ID:          string(RuleEmptyFile),
				Category:    "coverage",
				Severity:    "medium",
				Message:     "No CODEOWNERS file found",
				Remediation: "Create a CODEOWNERS file in .github/CODEOWNERS or the repository root",
			}},
			Recommendations: []CODEOWNERSRecommendation{{
				ID:       "REC001",
				Priority: "high",
				Type:     "add_codeowners",
				Message:  "Create a CODEOWNERS file to establish code ownership",
			}},
		}, nil
	}

	// Parse the CODEOWNERS file
	rules, parseIssues, err := a.parseCodeownersFile(codeownersPath)
	if err != nil {
		return nil, err
	}

	analysis := &CODEOWNERSAnalysis{
		FilePath:         codeownersPath,
		RulesCount:       len(rules),
		ValidationIssues: parseIssues,
	}

	// Run validation checks
	if a.config.Validate {
		issues := a.validateRules(rules, repoPath)
		analysis.ValidationIssues = append(analysis.ValidationIssues, issues...)

		// Check for best practices
		bestPracticeIssues := a.checkBestPractices(rules, repoPath)
		analysis.ValidationIssues = append(analysis.ValidationIssues, bestPracticeIssues...)
	}

	// Generate recommendations
	analysis.Recommendations = a.generateRecommendations(rules, analysis.ValidationIssues, contributors)

	// Calculate coverage
	analysis.Coverage = a.calculateCoverage(rules, repoPath)

	// Detect drift if enabled
	if a.config.DetectDrift {
		analysis.DriftAnalysis = a.detectDrift(rules, contributors)
	}

	return analysis, nil
}

// findCodeownersFile locates the CODEOWNERS file in standard locations
func (a *CODEOWNERSAnalyzer) findCodeownersFile(repoPath string) string {
	locations := []string{
		filepath.Join(repoPath, ".github", "CODEOWNERS"),
		filepath.Join(repoPath, "CODEOWNERS"),
		filepath.Join(repoPath, "docs", "CODEOWNERS"),
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc
		}
	}
	return ""
}

// parseCodeownersFile parses the CODEOWNERS file and returns rules and any parse issues
func (a *CODEOWNERSAnalyzer) parseCodeownersFile(path string) ([]CodeownerRule, []CODEOWNERSIssue, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	var rules []CodeownerRule
	var issues []CODEOWNERSIssue

	scanner := bufio.NewScanner(file)
	lineNum := 0

	ownerRegex := regexp.MustCompile(`^@[\w\-]+(/[\w\-]+)?$|^[\w\.\-]+@[\w\.\-]+$`)

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse the line
		parts := strings.Fields(line)
		if len(parts) < 2 {
			issues = append(issues, CODEOWNERSIssue{
				ID:          string(RuleInvalidPattern),
				Category:    "syntax",
				Severity:    "critical",
				Line:        lineNum,
				Pattern:     line,
				Message:     "Invalid CODEOWNERS line: pattern must have at least one owner",
				Remediation: "Add at least one owner after the pattern (e.g., '*.go @team/backend')",
			})
			continue
		}

		pattern := parts[0]
		owners := parts[1:]

		// Validate owners format
		validOwners := make([]string, 0, len(owners))
		for _, owner := range owners {
			if !ownerRegex.MatchString(owner) {
				issues = append(issues, CODEOWNERSIssue{
					ID:          string(RuleInvalidOwner),
					Category:    "syntax",
					Severity:    "high",
					Line:        lineNum,
					Pattern:     pattern,
					Owner:       owner,
					Message:     "Invalid owner format: must be @user, @org/team, or email",
					Remediation: "Use format @username, @org/team-name, or user@example.com",
				})
			} else {
				validOwners = append(validOwners, owner)
			}
		}

		if len(validOwners) > 0 {
			rules = append(rules, CodeownerRule{
				Pattern: pattern,
				Owners:  validOwners,
			})
		}
	}

	return rules, issues, scanner.Err()
}

// validateRules performs validation checks on the parsed rules
func (a *CODEOWNERSAnalyzer) validateRules(rules []CodeownerRule, _ string) []CODEOWNERSIssue {
	var issues []CODEOWNERSIssue

	// Check for default owner (pattern *)
	hasDefault := false
	for _, rule := range rules {
		if rule.Pattern == "*" {
			hasDefault = true
			break
		}
	}

	if !hasDefault {
		issues = append(issues, CODEOWNERSIssue{
			ID:          string(RuleNoDefaultOwner),
			Category:    "coverage",
			Severity:    "medium",
			Message:     "No default owner pattern (*) found",
			Remediation: "Add a default owner rule: * @org/default-team",
		})
	}

	// Check for overlapping patterns (later patterns override earlier ones)
	patternOrder := make(map[string]int)
	for i, rule := range rules {
		// Simple overlap check - more complex would need gitignore pattern matching
		if prevIdx, exists := patternOrder[rule.Pattern]; exists {
			issues = append(issues, CODEOWNERSIssue{
				ID:          string(RuleOverlappingPatterns),
				Category:    "best_practice",
				Severity:    "low",
				Line:        i + 1,
				Pattern:     rule.Pattern,
				Message:     "Duplicate pattern found (overrides line " + string(rune(prevIdx+1)) + ")",
				Remediation: "Remove duplicate pattern or consolidate owners",
			})
		}
		patternOrder[rule.Pattern] = i
	}

	return issues
}

// checkBestPractices checks for best practice violations
func (a *CODEOWNERSAnalyzer) checkBestPractices(rules []CodeownerRule, repoPath string) []CODEOWNERSIssue {
	var issues []CODEOWNERSIssue

	for _, rule := range rules {
		// Check for individual instead of team
		for _, owner := range rule.Owners {
			if strings.HasPrefix(owner, "@") && !strings.Contains(owner, "/") {
				// Individual user, not a team
				issues = append(issues, CODEOWNERSIssue{
					ID:          string(RuleIndividualNotTeam),
					Category:    "best_practice",
					Severity:    "low",
					Pattern:     rule.Pattern,
					Owner:       owner,
					Message:     "Individual user instead of team - reduces bus factor",
					Remediation: "Consider using a team (@org/team-name) instead of individual",
				})
			}
		}

		// Check for single owner (no backup)
		if len(rule.Owners) == 1 {
			issues = append(issues, CODEOWNERSIssue{
				ID:          string(RuleSingleOwner),
				Category:    "best_practice",
				Severity:    "medium",
				Pattern:     rule.Pattern,
				Message:     "Single owner with no backup - bus factor of 1",
				Remediation: "Add a backup owner or team for redundancy",
			})
		}

		// Check for too many owners (becomes unclear who's responsible)
		if len(rule.Owners) > 5 {
			issues = append(issues, CODEOWNERSIssue{
				ID:          string(RuleTooManyOwners),
				Category:    "best_practice",
				Severity:    "low",
				Pattern:     rule.Pattern,
				Message:     "Too many owners (" + string(rune(len(rule.Owners))) + ") - ownership becomes unclear",
				Remediation: "Consolidate owners into teams or reduce to primary + backup",
			})
		}
	}

	// Check for unprotected sensitive files
	sensitiveIssues := a.checkSensitiveFiles(rules, repoPath)
	issues = append(issues, sensitiveIssues...)

	return issues
}

// checkSensitiveFiles ensures sensitive files are protected
func (a *CODEOWNERSAnalyzer) checkSensitiveFiles(rules []CodeownerRule, repoPath string) []CODEOWNERSIssue {
	var issues []CODEOWNERSIssue

	for _, pattern := range a.sensitivePatterns {
		isProtected := false
		for _, rule := range rules {
			if a.patternMatches(rule.Pattern, pattern) {
				isProtected = true
				break
			}
		}

		if !isProtected {
			// Check if the file actually exists
			matches, _ := filepath.Glob(filepath.Join(repoPath, pattern))
			if len(matches) > 0 {
				issues = append(issues, CODEOWNERSIssue{
					ID:          string(RuleSensitiveUnprotected),
					Category:    "best_practice",
					Severity:    "medium",
					Pattern:     pattern,
					Message:     "Sensitive file pattern not protected in CODEOWNERS",
					Remediation: "Add a CODEOWNERS rule for " + pattern,
				})
			}
		}
	}

	return issues
}

// patternMatches checks if a CODEOWNERS pattern would match a file path
func (a *CODEOWNERSAnalyzer) patternMatches(codeownersPattern, filePath string) bool {
	// Simplified matching - real implementation would use gitignore-style matching
	if codeownersPattern == "*" {
		return true
	}

	// Direct match
	if codeownersPattern == filePath {
		return true
	}

	// Directory pattern
	if strings.HasSuffix(codeownersPattern, "/*") {
		dir := strings.TrimSuffix(codeownersPattern, "/*")
		return strings.HasPrefix(filePath, dir+"/") || strings.HasPrefix(filePath, dir)
	}

	// Glob pattern
	if strings.Contains(codeownersPattern, "*") {
		// Convert to regex-ish
		regexPattern := strings.ReplaceAll(codeownersPattern, "*", ".*")
		re, err := regexp.Compile("^" + regexPattern + "$")
		if err == nil && re.MatchString(filePath) {
			return true
		}
	}

	return false
}

// generateRecommendations creates actionable recommendations
func (a *CODEOWNERSAnalyzer) generateRecommendations(_ []CodeownerRule, issues []CODEOWNERSIssue, contributors []Contributor) []CODEOWNERSRecommendation {
	var recs []CODEOWNERSRecommendation

	// Count issue types
	issueTypes := make(map[string]int)
	for _, issue := range issues {
		issueTypes[issue.Category]++
	}

	// Recommend teams over individuals
	individualCount := 0
	for _, issue := range issues {
		if issue.ID == string(RuleIndividualNotTeam) {
			individualCount++
		}
	}
	if individualCount > 0 {
		recs = append(recs, CODEOWNERSRecommendation{
			ID:       "REC010",
			Priority: "medium",
			Type:     "add_team",
			Message:  "Replace individual owners with teams to improve bus factor",
		})
	}

	// Recommend backup owners
	singleOwnerCount := 0
	var affectedPaths []string
	for _, issue := range issues {
		if issue.ID == string(RuleSingleOwner) {
			singleOwnerCount++
			if issue.Pattern != "" {
				affectedPaths = append(affectedPaths, issue.Pattern)
			}
		}
	}
	if singleOwnerCount > 0 {
		// Suggest backup owners based on contributor data
		var suggestedOwners []string
		if len(contributors) >= 2 {
			for i := 0; i < 2 && i < len(contributors); i++ {
				suggestedOwners = append(suggestedOwners, "@"+contributors[i].Name)
			}
		}

		recs = append(recs, CODEOWNERSRecommendation{
			ID:              "REC020",
			Priority:        "high",
			Type:            "add_backup",
			Message:         "Add backup owners to patterns with single owner",
			AffectedPaths:   affectedPaths,
			SuggestedOwners: suggestedOwners,
		})
	}

	// Recommend protecting sensitive files
	for _, issue := range issues {
		if issue.ID == string(RuleSensitiveUnprotected) {
			recs = append(recs, CODEOWNERSRecommendation{
				ID:            "REC030",
				Priority:      "high",
				Type:          "protect_sensitive",
				Message:       "Add CODEOWNERS protection for sensitive file: " + issue.Pattern,
				AffectedPaths: []string{issue.Pattern},
			})
		}
	}

	return recs
}

// calculateCoverage determines what percentage of files are covered by CODEOWNERS
func (a *CODEOWNERSAnalyzer) calculateCoverage(rules []CodeownerRule, _ string) float64 {
	// Check for wildcard rule
	for _, rule := range rules {
		if rule.Pattern == "*" {
			return 1.0 // Full coverage with default owner
		}
	}

	// Without a default rule, calculate actual coverage
	// This is simplified - real implementation would walk the repo
	if len(rules) == 0 {
		return 0.0
	}

	// Estimate based on rule count (rough heuristic)
	// More sophisticated would match rules against actual files
	return 0.5 // Partial coverage estimate
}

// detectDrift compares declared owners with actual contributors
func (a *CODEOWNERSAnalyzer) detectDrift(rules []CodeownerRule, contributors []Contributor) *DriftAnalysis {
	if len(rules) == 0 || len(contributors) == 0 {
		return &DriftAnalysis{HasDrift: false, DriftScore: 0}
	}

	var driftItems []DriftItem
	var totalOverlap float64

	// For each rule, check if declared owners match actual top contributors
	for _, rule := range rules {
		// Get actual top contributors for this pattern
		// (simplified - real implementation would filter by path)
		topContributors := make([]string, 0)
		for i := 0; i < 3 && i < len(contributors); i++ {
			topContributors = append(topContributors, "@"+contributors[i].Name)
		}

		// Calculate overlap
		overlap := calculateOverlap(rule.Owners, topContributors)
		totalOverlap += overlap

		if overlap < 0.5 { // Less than 50% overlap indicates drift
			driftItems = append(driftItems, DriftItem{
				Path:            rule.Pattern,
				DeclaredOwners:  rule.Owners,
				ActualTopOwners: topContributors,
				OverlapScore:    overlap,
			})
		}
	}

	avgOverlap := totalOverlap / float64(len(rules))
	driftScore := (1 - avgOverlap) * 100 // Convert to 0-100 scale

	return &DriftAnalysis{
		HasDrift:     len(driftItems) > 0,
		DriftScore:   driftScore,
		DriftDetails: driftItems,
	}
}

// calculateOverlap determines how much two owner lists overlap (0-1)
func calculateOverlap(declared, actual []string) float64 {
	if len(declared) == 0 || len(actual) == 0 {
		return 0
	}

	// Normalize owners (lowercase, remove @)
	normalize := func(s string) string {
		return strings.ToLower(strings.TrimPrefix(s, "@"))
	}

	declaredSet := make(map[string]bool)
	for _, o := range declared {
		declaredSet[normalize(o)] = true
	}

	matches := 0
	for _, o := range actual {
		if declaredSet[normalize(o)] {
			matches++
		}
	}

	// Jaccard-like similarity
	union := len(declared) + len(actual) - matches
	if union == 0 {
		return 1.0
	}

	return float64(matches) / float64(union)
}
