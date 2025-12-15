// Package github provides GitHub API interactions
package github

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// PRReviewData contains PR review information for ownership analysis
type PRReviewData struct {
	PRNumber    int       `json:"number"`
	Title       string    `json:"title"`
	Author      string    `json:"author"`
	MergedAt    time.Time `json:"merged_at"`
	Reviews     []Review  `json:"reviews"`
	FilesChanged []string `json:"files_changed"`
}

// Review represents a single PR review
type Review struct {
	Author    string    `json:"author"`
	State     string    `json:"state"` // APPROVED, CHANGES_REQUESTED, COMMENTED
	SubmittedAt time.Time `json:"submitted_at"`
}

// TeamMember represents a member of a GitHub team
type TeamMember struct {
	Login string `json:"login"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Collaborator represents a repository collaborator
type Collaborator struct {
	Login      string `json:"login"`
	Permission string `json:"permission"` // admin, push, pull
}

// OwnershipClient provides methods for ownership-related GitHub API calls
type OwnershipClient struct {
	*Client
	maxPRs int
}

// NewOwnershipClient creates a client for ownership analysis
func NewOwnershipClient(maxPRs int) *OwnershipClient {
	return &OwnershipClient{
		Client: NewClient(),
		maxPRs: maxPRs,
	}
}

// HasToken returns whether a GitHub token is available
func (c *OwnershipClient) HasToken() bool {
	return c.token != ""
}

// FetchPRReviews fetches PR review data for a repository
func (c *OwnershipClient) FetchPRReviews(owner, repo string) ([]PRReviewData, int, error) {
	if !c.HasToken() {
		return nil, 0, fmt.Errorf("no GitHub token available")
	}

	// First, get the count of merged PRs
	countArgs := []string{
		"pr", "list",
		"--repo", fmt.Sprintf("%s/%s", owner, repo),
		"--state", "merged",
		"--json", "number",
		"--limit", "10000", // Get count
	}

	countCmd := exec.Command("gh", countArgs...)
	countOut, err := countCmd.Output()
	if err != nil {
		return nil, 0, fmt.Errorf("counting PRs: %w", err)
	}

	var countResult []struct{ Number int }
	if err := json.Unmarshal(countOut, &countResult); err != nil {
		return nil, 0, fmt.Errorf("parsing PR count: %w", err)
	}

	totalPRs := len(countResult)

	// If too many PRs, return early with warning
	if totalPRs > c.maxPRs {
		return nil, totalPRs, nil // Caller should check if result is nil but totalPRs > maxPRs
	}

	// Fetch PR details with reviews
	args := []string{
		"pr", "list",
		"--repo", fmt.Sprintf("%s/%s", owner, repo),
		"--state", "merged",
		"--json", "number,title,author,mergedAt,reviews",
		"--limit", fmt.Sprintf("%d", c.maxPRs),
	}

	cmd := exec.Command("gh", args...)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, 0, fmt.Errorf("fetching PRs: %s", string(exitErr.Stderr))
		}
		return nil, 0, fmt.Errorf("fetching PRs: %w", err)
	}

	var ghPRs []struct {
		Number   int    `json:"number"`
		Title    string `json:"title"`
		Author   struct {
			Login string `json:"login"`
		} `json:"author"`
		MergedAt string `json:"mergedAt"`
		Reviews  []struct {
			Author struct {
				Login string `json:"login"`
			} `json:"author"`
			State       string `json:"state"`
			SubmittedAt string `json:"submittedAt"`
		} `json:"reviews"`
	}

	if err := json.Unmarshal(out, &ghPRs); err != nil {
		return nil, 0, fmt.Errorf("parsing PRs: %w", err)
	}

	// Convert to our format
	result := make([]PRReviewData, 0, len(ghPRs))
	for _, pr := range ghPRs {
		mergedAt, _ := time.Parse(time.RFC3339, pr.MergedAt)

		reviews := make([]Review, 0, len(pr.Reviews))
		for _, r := range pr.Reviews {
			submittedAt, _ := time.Parse(time.RFC3339, r.SubmittedAt)
			reviews = append(reviews, Review{
				Author:      r.Author.Login,
				State:       r.State,
				SubmittedAt: submittedAt,
			})
		}

		result = append(result, PRReviewData{
			PRNumber: pr.Number,
			Title:    pr.Title,
			Author:   pr.Author.Login,
			MergedAt: mergedAt,
			Reviews:  reviews,
		})
	}

	return result, totalPRs, nil
}

// ResolveTeam returns the members of a GitHub team
func (c *OwnershipClient) ResolveTeam(org, teamSlug string) ([]TeamMember, error) {
	if !c.HasToken() {
		return nil, fmt.Errorf("no GitHub token available")
	}

	// Use gh api to fetch team members
	args := []string{
		"api",
		fmt.Sprintf("/orgs/%s/teams/%s/members", org, teamSlug),
		"--jq", ".[].login",
	}

	cmd := exec.Command("gh", args...)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("fetching team members: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("fetching team members: %w", err)
	}

	// Parse line-by-line output
	logins := strings.Split(strings.TrimSpace(string(out)), "\n")
	members := make([]TeamMember, 0, len(logins))
	for _, login := range logins {
		if login != "" {
			members = append(members, TeamMember{Login: login})
		}
	}

	return members, nil
}

// GetCollaborators returns collaborators for a repository
func (c *OwnershipClient) GetCollaborators(owner, repo string) ([]Collaborator, error) {
	if !c.HasToken() {
		return nil, fmt.Errorf("no GitHub token available")
	}

	args := []string{
		"api",
		fmt.Sprintf("/repos/%s/%s/collaborators", owner, repo),
		"--jq", ".[] | {login: .login, permission: .role_name}",
	}

	cmd := exec.Command("gh", args...)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("fetching collaborators: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("fetching collaborators: %w", err)
	}

	// Parse JSON lines
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	collaborators := make([]Collaborator, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		var collab Collaborator
		if err := json.Unmarshal([]byte(line), &collab); err != nil {
			continue // Skip malformed lines
		}
		collaborators = append(collaborators, collab)
	}

	return collaborators, nil
}

// CheckUserExists verifies if a GitHub user exists
func (c *OwnershipClient) CheckUserExists(username string) (bool, error) {
	if !c.HasToken() {
		return false, fmt.Errorf("no GitHub token available")
	}

	// Remove @ prefix if present
	username = strings.TrimPrefix(username, "@")

	// Handle team references
	if strings.Contains(username, "/") {
		parts := strings.Split(username, "/")
		if len(parts) == 2 {
			// This is a team reference, check if team exists
			args := []string{
				"api",
				fmt.Sprintf("/orgs/%s/teams/%s", parts[0], parts[1]),
				"--silent",
			}
			cmd := exec.Command("gh", args...)
			err := cmd.Run()
			return err == nil, nil
		}
	}

	// Check user exists
	args := []string{
		"api",
		fmt.Sprintf("/users/%s", username),
		"--silent",
	}
	cmd := exec.Command("gh", args...)
	err := cmd.Run()
	return err == nil, nil
}

// AggregateReviewerStats aggregates review statistics from PR data
func AggregateReviewerStats(prs []PRReviewData) map[string]*ReviewerStats {
	stats := make(map[string]*ReviewerStats)

	for _, pr := range prs {
		for _, review := range pr.Reviews {
			if review.Author == "" {
				continue
			}

			if _, exists := stats[review.Author]; !exists {
				stats[review.Author] = &ReviewerStats{
					Login: review.Author,
				}
			}

			s := stats[review.Author]
			s.ReviewsGiven++

			switch review.State {
			case "APPROVED":
				s.Approvals++
			case "CHANGES_REQUESTED":
				s.ChangesRequested++
			case "COMMENTED":
				s.Comments++
			}
		}
	}

	return stats
}

// ReviewerStats holds aggregated review statistics for a user
type ReviewerStats struct {
	Login            string `json:"login"`
	ReviewsGiven     int    `json:"reviews_given"`
	Approvals        int    `json:"approvals"`
	ChangesRequested int    `json:"changes_requested"`
	Comments         int    `json:"comments"`
}
