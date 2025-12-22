// Package codeownership provides code ownership and CODEOWNERS analysis
package codeownership

import (
	"sort"
	"strings"
)

// ContactGenerator generates incident contacts from ownership data
type ContactGenerator struct {
	config ContactsConfig
}

// NewContactGenerator creates a new contact generator
func NewContactGenerator(config ContactsConfig) *ContactGenerator {
	return &ContactGenerator{config: config}
}

// GenerateContacts creates incident contact lists for specified paths
func (g *ContactGenerator) GenerateContacts(
	paths []string,
	owners []EnhancedOwnership,
	codeownersRules []CodeownerRule,
) []IncidentContact {
	if !g.config.Enabled || len(paths) == 0 {
		return nil
	}

	contacts := make([]IncidentContact, 0, len(paths))

	for _, path := range paths {
		contact := g.generateForPath(path, owners, codeownersRules)
		contacts = append(contacts, contact)
	}

	return contacts
}

// generateForPath creates contacts for a single path
func (g *ContactGenerator) generateForPath(
	path string,
	owners []EnhancedOwnership,
	codeownersRules []CodeownerRule,
) IncidentContact {
	contact := IncidentContact{
		Path:    path,
		Primary: make([]ContactInfo, 0),
		Backup:  make([]ContactInfo, 0),
	}

	// Check CODEOWNERS for this path
	for _, rule := range codeownersRules {
		if matchesPattern(rule.Pattern, path) {
			contact.CodeownersMatch = &rule
			break
		}
	}

	// Rank owners by suitability for this path
	rankedOwners := g.rankOwnersForPath(path, owners)

	// Select primary contacts
	for i := 0; i < g.config.MinPrimary && i < len(rankedOwners); i++ {
		contact.Primary = append(contact.Primary, rankedOwners[i])
	}

	// Select backup contacts (next best after primary)
	for i := g.config.MinPrimary; i < g.config.MinPrimary+g.config.MinBackup && i < len(rankedOwners); i++ {
		contact.Backup = append(contact.Backup, rankedOwners[i])
	}

	return contact
}

// rankOwnersForPath ranks owners by their suitability for a specific path
func (g *ContactGenerator) rankOwnersForPath(path string, owners []EnhancedOwnership) []ContactInfo {
	type rankedContact struct {
		contact ContactInfo
		score   float64
	}

	ranked := make([]rankedContact, 0, len(owners))

	for _, owner := range owners {
		// Calculate expertise score (based on ownership score)
		expertiseScore := owner.OwnershipScore / 100

		// Calculate availability score (based on activity status)
		availabilityScore := g.calculateAvailability(owner.ActivityStatus)

		// Combine scores
		totalScore := (expertiseScore * 0.6) + (availabilityScore * 0.4)

		reason := g.determineReason(owner, path)

		ranked = append(ranked, rankedContact{
			contact: ContactInfo{
				Name:              owner.Name,
				Email:             owner.Email,
				ExpertiseScore:    expertiseScore,
				AvailabilityScore: availabilityScore,
				ReasonForContact:  reason,
			},
			score: totalScore,
		})
	}

	// Sort by combined score
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].score > ranked[j].score
	})

	// Extract contacts
	contacts := make([]ContactInfo, 0, len(ranked))
	for _, r := range ranked {
		contacts = append(contacts, r.contact)
	}

	return contacts
}

// calculateAvailability converts activity status to availability score
func (g *ContactGenerator) calculateAvailability(status string) float64 {
	switch status {
	case "active":
		return 1.0
	case "recent":
		return 0.8
	case "stale":
		return 0.5
	case "inactive":
		return 0.2
	case "abandoned":
		return 0.0
	default:
		return 0.5
	}
}

// determineReason explains why this person is recommended
func (g *ContactGenerator) determineReason(owner EnhancedOwnership, _ string) string {
	reasons := []string{}

	if owner.OwnershipScore >= 50 {
		reasons = append(reasons, "high ownership score")
	}

	if owner.ActivityStatus == "active" {
		reasons = append(reasons, "recently active")
	}

	if owner.PRReviewsGiven > 10 {
		reasons = append(reasons, "frequent reviewer")
	}

	if len(reasons) == 0 {
		return "contributor to this codebase"
	}

	return strings.Join(reasons, ", ")
}

// matchesPattern checks if a path matches a CODEOWNERS pattern
func matchesPattern(pattern, path string) bool {
	// Simple matching - real implementation would use gitignore-style
	if pattern == "*" {
		return true
	}

	// Exact match
	if pattern == path {
		return true
	}

	// Directory pattern (e.g., "src/*")
	if strings.HasSuffix(pattern, "/*") {
		dir := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(path, dir+"/") || path == dir
	}

	// Starts with pattern (e.g., "*.go" matches "file.go")
	if strings.HasPrefix(pattern, "*") {
		suffix := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(path, suffix)
	}

	// Ends with pattern
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}

	return false
}

// GenerateKeyPathContacts generates contacts for key paths in the repo
func (g *ContactGenerator) GenerateKeyPathContacts(
	repoPath string,
	owners []EnhancedOwnership,
	codeownersRules []CodeownerRule,
) []IncidentContact {
	// Key paths that are commonly needed for incident response
	keyPaths := []string{
		"src/",
		"pkg/",
		"lib/",
		"api/",
		"auth/",
		"security/",
		".github/workflows/",
		"deploy/",
		"infrastructure/",
		"database/",
	}

	return g.GenerateContacts(keyPaths, owners, codeownersRules)
}
