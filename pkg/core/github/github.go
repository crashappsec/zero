// Package github provides GitHub API interactions
package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/crashappsec/zero/pkg/core/credentials"
)

// ============================================================================
// Core Types and Client
// ============================================================================

// Repository represents a GitHub repository
type Repository struct {
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	NameWithOwner string `json:"nameWithOwner"`
	Owner         string `json:"owner"`
	CloneURL      string `json:"clone_url"`
	SSHURL        string `json:"ssh_url"`
	Size          int    `json:"size"` // in KB
	DefaultBranch string `json:"default_branch"`
	Private       bool   `json:"private"`
	Archived      bool   `json:"archived"`
	Fork          bool   `json:"fork"`
}

// Client provides GitHub API access
type Client struct {
	token      string
	httpClient *http.Client
}

// NewClient creates a new GitHub client
func NewClient() *Client {
	// Use credentials package to get token from best available source
	tokenInfo := credentials.GetGitHubToken()

	return &Client{
		token: tokenInfo.Value,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetToken returns the current token (for external use)
func (c *Client) GetToken() string {
	return c.token
}

// HasToken returns true if a token is configured
func (c *Client) HasToken() bool {
	return c.token != ""
}

// ============================================================================
// Repository Operations
// ============================================================================

// ListOrgRepos returns repositories for an organization
func (c *Client) ListOrgRepos(org string, limit int) ([]Repository, error) {
	// Use gh CLI for simplicity and auth handling
	args := []string{
		"repo", "list", org,
		"--json", "name,nameWithOwner,sshUrl,defaultBranchRef,isPrivate,isArchived,isFork,diskUsage",
		"--limit", fmt.Sprintf("%d", limit),
	}

	cmd := exec.Command("gh", args...)
	out, err := cmd.Output()
	if err != nil {
		// Try to get stderr for more info
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("listing repos: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("listing repos: %w", err)
	}

	var ghRepos []struct {
		Name             string `json:"name"`
		NameWithOwner    string `json:"nameWithOwner"`
		SSHURL           string `json:"sshUrl"`
		DiskUsage        int    `json:"diskUsage"` // Size in KB
		DefaultBranchRef struct {
			Name string `json:"name"`
		} `json:"defaultBranchRef"`
		IsPrivate  bool `json:"isPrivate"`
		IsArchived bool `json:"isArchived"`
		IsFork     bool `json:"isFork"`
	}

	if err := json.Unmarshal(out, &ghRepos); err != nil {
		return nil, fmt.Errorf("parsing repos: %w", err)
	}

	repos := make([]Repository, len(ghRepos))
	for i, r := range ghRepos {
		parts := strings.Split(r.NameWithOwner, "/")
		owner := ""
		if len(parts) >= 2 {
			owner = parts[0]
		}

		defaultBranch := r.DefaultBranchRef.Name
		if defaultBranch == "" {
			defaultBranch = "main"
		}

		repos[i] = Repository{
			Name:          r.Name,
			FullName:      r.NameWithOwner,
			NameWithOwner: r.NameWithOwner,
			Owner:         owner,
			SSHURL:        r.SSHURL,
			Size:          r.DiskUsage,
			DefaultBranch: defaultBranch,
			Private:       r.IsPrivate,
			Archived:      r.IsArchived,
			Fork:          r.IsFork,
		}
	}

	return repos, nil
}

// ProjectID returns the project identifier for a repo
func ProjectID(nameWithOwner string) string {
	return strings.ToLower(nameWithOwner)
}

// ShortName returns just the repo name without owner
func ShortName(nameWithOwner string) string {
	parts := strings.Split(nameWithOwner, "/")
	if len(parts) >= 2 {
		return parts[1]
	}
	return nameWithOwner
}

// ============================================================================
// Ownership Analysis (PR Reviews, Teams, Collaborators)
// ============================================================================

// PRReviewData contains PR review information for ownership analysis
type PRReviewData struct {
	PRNumber     int       `json:"number"`
	Title        string    `json:"title"`
	Author       string    `json:"author"`
	MergedAt     time.Time `json:"merged_at"`
	Reviews      []Review  `json:"reviews"`
	FilesChanged []string  `json:"files_changed"`
}

// Review represents a single PR review
type Review struct {
	Author      string    `json:"author"`
	State       string    `json:"state"` // APPROVED, CHANGES_REQUESTED, COMMENTED
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

// ReviewerStats holds aggregated review statistics for a user
type ReviewerStats struct {
	Login            string `json:"login"`
	ReviewsGiven     int    `json:"reviews_given"`
	Approvals        int    `json:"approvals"`
	ChangesRequested int    `json:"changes_requested"`
	Comments         int    `json:"comments"`
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
		Number int    `json:"number"`
		Title  string `json:"title"`
		Author struct {
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

// ============================================================================
// Token Permissions and Scanner Compatibility
// ============================================================================

// TokenInfo contains information about the GitHub token
type TokenInfo struct {
	// Token type: "classic" (PAT), "fine-grained", "github-app", "oauth", or "unknown"
	Type string `json:"type"`
	// Scopes for classic PATs (e.g., "repo", "read:org", "admin:org")
	Scopes []string `json:"scopes,omitempty"`
	// Permissions for fine-grained PATs
	Permissions map[string]string `json:"permissions,omitempty"`
	// Username associated with the token
	Username string `json:"username,omitempty"`
	// Whether the token is valid
	Valid bool `json:"valid"`
	// Error message if invalid
	Error string `json:"error,omitempty"`
	// Rate limit info
	RateLimit     int `json:"rate_limit,omitempty"`
	RateRemaining int `json:"rate_remaining,omitempty"`
}

// ScannerRequirement defines what a scanner needs to function
type ScannerRequirement struct {
	Scanner     string   `json:"scanner"`
	Description string   `json:"description"`
	// Required scopes for classic PATs
	RequiredScopes []string `json:"required_scopes,omitempty"`
	// Required permissions for fine-grained PATs (permission -> access level)
	RequiredPermissions map[string]string `json:"required_permissions,omitempty"`
	// Whether the scanner needs GitHub API access at all
	NeedsGitHubAPI bool `json:"needs_github_api"`
	// Whether the scanner needs org-level access
	NeedsOrgAccess bool `json:"needs_org_access"`
	// External tools required
	RequiredTools []string `json:"required_tools,omitempty"`
}

// ScannerCompatibility shows whether a scanner will work
type ScannerCompatibility struct {
	Scanner       string   `json:"scanner"`
	Status        string   `json:"status"` // "ready", "limited", "unavailable"
	Reason        string   `json:"reason,omitempty"`
	MissingScopes []string `json:"missing_scopes,omitempty"`
	MissingTools  []string `json:"missing_tools,omitempty"`
}

// RoadmapResult contains the full compatibility analysis
type RoadmapResult struct {
	TokenInfo   TokenInfo              `json:"token_info"`
	Scanners    []ScannerCompatibility `json:"scanners"`
	Summary     RoadmapSummary         `json:"summary"`
	ToolsStatus map[string]bool        `json:"tools_status"`
}

// RoadmapSummary provides a quick overview
type RoadmapSummary struct {
	Ready       int `json:"ready"`
	Limited     int `json:"limited"`
	Unavailable int `json:"unavailable"`
	Total       int `json:"total"`
}

// ScannerRequirements defines what each scanner needs
var ScannerRequirements = map[string]ScannerRequirement{
	// Package scanners - mostly local, some need API for health checks
	"package-sbom": {
		Scanner:        "package-sbom",
		Description:    "Generate Software Bill of Materials",
		NeedsGitHubAPI: false,
		RequiredTools:  []string{"cdxgen", "syft"},
	},
	"package-vulns": {
		Scanner:        "package-vulns",
		Description:    "Scan for known vulnerabilities",
		NeedsGitHubAPI: false,
		RequiredTools:  []string{"osv-scanner", "grype"},
	},
	"package-health": {
		Scanner:        "package-health",
		Description:    "Check package health scores",
		NeedsGitHubAPI: true,
		RequiredScopes: []string{"public_repo"},
		RequiredPermissions: map[string]string{
			"metadata": "read",
		},
	},
	"package-provenance": {
		Scanner:        "package-provenance",
		Description:    "Verify package provenance and signatures",
		NeedsGitHubAPI: true,
		RequiredScopes: []string{"public_repo"},
		RequiredPermissions: map[string]string{
			"metadata": "read",
		},
	},
	"package-malcontent": {
		Scanner:        "package-malcontent",
		Description:    "Detect malicious package behavior",
		NeedsGitHubAPI: false,
		RequiredTools:  []string{"malcontent"},
	},
	"licenses": {
		Scanner:        "licenses",
		Description:    "Analyze license compliance",
		NeedsGitHubAPI: false,
	},

	// Code security scanners - mostly local
	"code-vulns": {
		Scanner:        "code-vulns",
		Description:    "Static code analysis for vulnerabilities",
		NeedsGitHubAPI: false,
		RequiredTools:  []string{"semgrep"},
	},
	"code-secrets": {
		Scanner:        "code-secrets",
		Description:    "Detect hardcoded secrets (native RAG-based detection)",
		NeedsGitHubAPI: false,
		// Native implementation uses Semgrep p/secrets + RAG patterns + entropy analysis
	},
	"api-security": {
		Scanner:        "api-security",
		Description:    "Analyze API security",
		NeedsGitHubAPI: false,
	},

	// Crypto scanners - all local
	"crypto-ciphers": {
		Scanner:        "crypto-ciphers",
		Description:    "Detect weak cryptographic algorithms",
		NeedsGitHubAPI: false,
	},
	"crypto-keys": {
		Scanner:        "crypto-keys",
		Description:    "Find hardcoded keys",
		NeedsGitHubAPI: false,
	},
	"crypto-random": {
		Scanner:        "crypto-random",
		Description:    "Detect insecure random generation",
		NeedsGitHubAPI: false,
	},
	"crypto-tls": {
		Scanner:        "crypto-tls",
		Description:    "Analyze TLS configuration",
		NeedsGitHubAPI: false,
	},

	// Git/repo analysis - needs repo access
	"git": {
		Scanner:        "git",
		Description:    "Git history and contributor analysis",
		NeedsGitHubAPI: true,
		RequiredScopes: []string{"repo"},
		RequiredPermissions: map[string]string{
			"contents": "read",
		},
	},
	"code-ownership": {
		Scanner:        "code-ownership",
		Description:    "Analyze code ownership patterns",
		NeedsGitHubAPI: true,
		RequiredScopes: []string{"repo"},
		RequiredPermissions: map[string]string{
			"contents": "read",
		},
	},
	"dora-metrics": {
		Scanner:        "dora-metrics",
		Description:    "Calculate DORA metrics",
		NeedsGitHubAPI: true,
		NeedsOrgAccess: true,
		RequiredScopes: []string{"repo", "read:org"},
		RequiredPermissions: map[string]string{
			"contents":                    "read",
			"pull_requests":               "read",
			"organization_administration": "read",
		},
	},

	// Infrastructure scanners
	"iac-security": {
		Scanner:        "iac-security",
		Description:    "Infrastructure as Code security",
		NeedsGitHubAPI: false,
		RequiredTools:  []string{"trivy", "checkov"},
	},
	"container-security": {
		Scanner:        "container-security",
		Description:    "Container image analysis",
		NeedsGitHubAPI: false,
		RequiredTools:  []string{"trivy"},
	},
	"containers": {
		Scanner:        "containers",
		Description:    "Dockerfile best practices",
		NeedsGitHubAPI: false,
	},

	// Tech analysis - local
	"tech-discovery": {
		Scanner:        "tech-discovery",
		Description:    "Identify technologies used",
		NeedsGitHubAPI: false,
	},
	"tech-debt": {
		Scanner:        "tech-debt",
		Description:    "Analyze technical debt",
		NeedsGitHubAPI: false,
	},

	// GitHub-specific scanners that need API access
	"github-actions": {
		Scanner:        "github-actions",
		Description:    "Analyze GitHub Actions security",
		NeedsGitHubAPI: true,
		RequiredScopes: []string{"repo", "workflow"},
		RequiredPermissions: map[string]string{
			"actions":  "read",
			"contents": "read",
		},
	},
	"dependabot": {
		Scanner:        "dependabot",
		Description:    "Check Dependabot alerts",
		NeedsGitHubAPI: true,
		RequiredScopes: []string{"repo", "security_events"},
		RequiredPermissions: map[string]string{
			"vulnerability_alerts": "read",
		},
	},
	"code-scanning": {
		Scanner:        "code-scanning",
		Description:    "GitHub Code Scanning alerts",
		NeedsGitHubAPI: true,
		RequiredScopes: []string{"repo", "security_events"},
		RequiredPermissions: map[string]string{
			"security_events": "read",
		},
	},
	"secret-scanning": {
		Scanner:        "secret-scanning",
		Description:    "GitHub Secret Scanning alerts",
		NeedsGitHubAPI: true,
		RequiredScopes: []string{"repo", "security_events"},
		RequiredPermissions: map[string]string{
			"secret_scanning_alerts": "read",
		},
	},

	// Super scanners (v2.0) - consolidated scanners with multiple features
	"packages": {
		Scanner:        "packages",
		Description:    "Consolidated package scanner (SBOM, vulnerabilities, health, malcontent, licenses)",
		NeedsGitHubAPI: true,
		RequiredScopes: []string{"public_repo"},
		RequiredPermissions: map[string]string{
			"metadata": "read",
		},
		RequiredTools: []string{"cdxgen", "osv-scanner", "malcontent"},
	},
	"code-packages": {
		Scanner:        "code-packages",
		Description:    "SBOM generation + package/dependency analysis (vulns, health, licenses, malcontent)",
		NeedsGitHubAPI: true,
		RequiredScopes: []string{"public_repo"},
		RequiredPermissions: map[string]string{
			"metadata": "read",
		},
		RequiredTools: []string{"cdxgen"},
	},
	"crypto": {
		Scanner:        "crypto",
		Description:    "Consolidated cryptographic security scanner",
		NeedsGitHubAPI: false,
	},
	"code": {
		Scanner:        "code",
		Description:    "Consolidated code security scanner (SAST, secrets, API security)",
		NeedsGitHubAPI: false,
		RequiredTools:  []string{"semgrep"},
	},
	"infra": {
		Scanner:        "infra",
		Description:    "Consolidated infrastructure scanner (IaC, containers, GitHub Actions, DORA)",
		NeedsGitHubAPI: true,
		RequiredScopes: []string{"repo", "workflow", "read:org"},
		RequiredPermissions: map[string]string{
			"actions":  "read",
			"contents": "read",
		},
		RequiredTools: []string{"trivy"},
	},
	"devops": {
		Scanner:        "devops",
		Description:    "Consolidated DevOps scanner (renamed from infra)",
		NeedsGitHubAPI: true,
		RequiredScopes: []string{"repo", "workflow", "read:org"},
		RequiredPermissions: map[string]string{
			"actions":  "read",
			"contents": "read",
		},
		RequiredTools: []string{"trivy", "checkov"},
	},
	"health": {
		Scanner:        "health",
		Description:    "Consolidated project health scanner (technology, docs, tests, ownership)",
		NeedsGitHubAPI: true,
		RequiredScopes: []string{"repo"},
		RequiredPermissions: map[string]string{
			"contents": "read",
		},
	},
	"sbom": {
		Scanner:        "sbom",
		Description:    "SBOM generation and integrity super scanner",
		NeedsGitHubAPI: false,
		RequiredTools:  []string{"cdxgen", "syft"},
	},
	"ai": {
		Scanner:        "ai",
		Description:    "AI/ML security analysis and ML-BOM generation",
		NeedsGitHubAPI: false,
	},

	// Super scanners v3.6 - consolidated scanners with updated names
	"package-analysis": {
		Scanner:        "package-analysis",
		Description:    "Package analysis (vulnerabilities, health, licenses, malcontent)",
		NeedsGitHubAPI: true,
		RequiredScopes: []string{"public_repo"},
		RequiredPermissions: map[string]string{
			"metadata": "read",
		},
		RequiredTools: []string{"cdxgen", "osv-scanner", "malcontent"},
	},
	"code-security": {
		Scanner:        "code-security",
		Description:    "Code security scanning (SAST, secrets, API security)",
		NeedsGitHubAPI: false,
		RequiredTools:  []string{"semgrep"},
	},
	"code-crypto": {
		Scanner:        "code-crypto",
		Description:    "Cryptographic security analysis",
		NeedsGitHubAPI: false,
	},
	"code-quality": {
		Scanner:        "code-quality",
		Description:    "Code quality metrics (complexity, tech debt, test coverage)",
		NeedsGitHubAPI: false,
		RequiredTools:  []string{"semgrep"},
	},
	"tech-id": {
		Scanner:        "tech-id",
		Description:    "Technology detection and ML-BOM generation",
		NeedsGitHubAPI: false,
	},
	"technology-identification": {
		Scanner:        "technology-identification",
		Description:    "Technology detection, ML-BOM generation, AI/ML security analysis",
		NeedsGitHubAPI: false,
		RequiredTools:  []string{"semgrep"},
	},
	"devx": {
		Scanner:        "devx",
		Description:    "Developer experience analysis",
		NeedsGitHubAPI: false,
	},
	"developer-experience": {
		Scanner:        "developer-experience",
		Description:    "Developer experience analysis",
		NeedsGitHubAPI: false,
	},
	// Billing data access (for GetBillingData agent tool)
	"billing": {
		Scanner:        "billing",
		Description:    "GitHub billing and usage data (Actions minutes, Packages storage)",
		NeedsGitHubAPI: true,
		NeedsOrgAccess: true,
		RequiredScopes: []string{"admin:org"}, // read:org is NOT sufficient
		RequiredPermissions: map[string]string{
			"organization_administration": "write", // Fine-grained needs write, not just read
		},
	},
}

// CheckTokenPermissions analyzes a GitHub token and returns its capabilities
func (c *Client) CheckTokenPermissions() (*TokenInfo, error) {
	info := &TokenInfo{
		Type:  "unknown",
		Valid: false,
	}

	if c.token == "" {
		info.Error = "No GitHub token found (set GITHUB_TOKEN or use 'gh auth login')"
		return info, nil
	}

	// Make a request to check the token
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		info.Error = fmt.Sprintf("Failed to create request: %v", err)
		return info, nil
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		info.Error = fmt.Sprintf("Failed to connect to GitHub: %v", err)
		return info, nil
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode == 401 {
		info.Error = "Token is invalid or expired"
		return info, nil
	}

	if resp.StatusCode != 200 {
		info.Error = fmt.Sprintf("GitHub API returned status %d", resp.StatusCode)
		return info, nil
	}

	info.Valid = true

	// Parse user info
	var user struct {
		Login string `json:"login"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err == nil {
		info.Username = user.Login
	}

	// Check scopes from header (classic PATs)
	scopeHeader := resp.Header.Get("X-OAuth-Scopes")
	if scopeHeader != "" {
		info.Type = "classic"
		scopes := strings.Split(scopeHeader, ", ")
		for _, s := range scopes {
			s = strings.TrimSpace(s)
			if s != "" {
				info.Scopes = append(info.Scopes, s)
			}
		}
	}

	// Check for fine-grained token indicators
	// Fine-grained tokens don't return X-OAuth-Scopes
	if scopeHeader == "" && info.Valid {
		// Could be fine-grained or GitHub App token
		// Try to detect by checking a permission-specific endpoint
		info.Type = "fine-grained"
		info.Permissions = c.detectFineGrainedPermissions()
	}

	// Get rate limit info
	if limit := resp.Header.Get("X-RateLimit-Limit"); limit != "" {
		_, _ = fmt.Sscanf(limit, "%d", &info.RateLimit)
	}
	if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining != "" {
		_, _ = fmt.Sscanf(remaining, "%d", &info.RateRemaining)
	}

	return info, nil
}

// detectFineGrainedPermissions tries to determine permissions for fine-grained tokens
func (c *Client) detectFineGrainedPermissions() map[string]string {
	perms := make(map[string]string)

	// Test various endpoints to detect permissions
	endpoints := map[string]struct {
		url        string
		permission string
		level      string
	}{
		"repos": {
			url:        "https://api.github.com/user/repos?per_page=1",
			permission: "contents",
			level:      "read",
		},
		"orgs": {
			url:        "https://api.github.com/user/orgs?per_page=1",
			permission: "organization",
			level:      "read",
		},
	}

	for _, ep := range endpoints {
		req, _ := http.NewRequest("GET", ep.url, nil)
		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("Accept", "application/vnd.github+json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == 200 {
			perms[ep.permission] = ep.level
		}
	}

	return perms
}

// CheckToolAvailability checks if required external tools are installed
func CheckToolAvailability(tools []string) map[string]bool {
	status := make(map[string]bool)
	for _, tool := range tools {
		_, err := exec.LookPath(tool)
		status[tool] = err == nil
	}
	return status
}

// GenerateRoadmap analyzes token permissions and tool availability
func (c *Client) GenerateRoadmap(scanners []string) (*RoadmapResult, error) {
	result := &RoadmapResult{
		Scanners:    make([]ScannerCompatibility, 0),
		ToolsStatus: make(map[string]bool),
	}

	// Check token permissions
	tokenInfo, err := c.CheckTokenPermissions()
	if err != nil {
		return nil, err
	}
	result.TokenInfo = *tokenInfo

	// Collect all required tools
	allTools := make(map[string]bool)
	for _, name := range scanners {
		if req, ok := ScannerRequirements[name]; ok {
			for _, tool := range req.RequiredTools {
				allTools[tool] = true
			}
		}
	}

	// Check tool availability
	toolList := make([]string, 0, len(allTools))
	for tool := range allTools {
		toolList = append(toolList, tool)
	}
	result.ToolsStatus = CheckToolAvailability(toolList)

	// Analyze each scanner
	for _, name := range scanners {
		compat := ScannerCompatibility{
			Scanner: name,
			Status:  "ready",
		}

		req, ok := ScannerRequirements[name]
		if !ok {
			// Unknown scanner, assume it works
			compat.Reason = "No specific requirements defined"
			result.Scanners = append(result.Scanners, compat)
			continue
		}

		// Check required tools
		for _, tool := range req.RequiredTools {
			if !result.ToolsStatus[tool] {
				compat.MissingTools = append(compat.MissingTools, tool)
			}
		}

		// Check GitHub API requirements
		if req.NeedsGitHubAPI {
			if !tokenInfo.Valid {
				compat.Status = "unavailable"
				compat.Reason = "Requires valid GitHub token"
			} else if tokenInfo.Type == "classic" {
				// Check scopes for classic PAT
				for _, scope := range req.RequiredScopes {
					if !hasScope(tokenInfo.Scopes, scope) {
						compat.MissingScopes = append(compat.MissingScopes, scope)
					}
				}
			} else if tokenInfo.Type == "fine-grained" {
				// Check permissions for fine-grained PAT
				for perm, level := range req.RequiredPermissions {
					if tokenLevel, ok := tokenInfo.Permissions[perm]; !ok || !hasAccessLevel(tokenLevel, level) {
						compat.MissingScopes = append(compat.MissingScopes, fmt.Sprintf("%s:%s", perm, level))
					}
				}
			}
		}

		// Determine final status
		if len(compat.MissingTools) > 0 && len(compat.MissingScopes) > 0 {
			compat.Status = "unavailable"
			compat.Reason = fmt.Sprintf("Missing tools (%s) and permissions (%s)",
				strings.Join(compat.MissingTools, ", "),
				strings.Join(compat.MissingScopes, ", "))
		} else if len(compat.MissingTools) > 0 {
			// Check if any alternative tool is available
			hasAlternative := false
			for _, tool := range req.RequiredTools {
				if result.ToolsStatus[tool] {
					hasAlternative = true
					break
				}
			}
			if hasAlternative {
				compat.Status = "ready"
				compat.Reason = fmt.Sprintf("Using available tool (missing: %s)",
					strings.Join(compat.MissingTools, ", "))
			} else {
				compat.Status = "unavailable"
				compat.Reason = fmt.Sprintf("Missing required tools: %s",
					strings.Join(compat.MissingTools, ", "))
			}
		} else if len(compat.MissingScopes) > 0 {
			compat.Status = "limited"
			compat.Reason = fmt.Sprintf("Missing permissions: %s",
				strings.Join(compat.MissingScopes, ", "))
		}

		result.Scanners = append(result.Scanners, compat)
	}

	// Calculate summary
	for _, s := range result.Scanners {
		result.Summary.Total++
		switch s.Status {
		case "ready":
			result.Summary.Ready++
		case "limited":
			result.Summary.Limited++
		case "unavailable":
			result.Summary.Unavailable++
		}
	}

	return result, nil
}

// hasScope checks if a scope is present in the list
func hasScope(scopes []string, target string) bool {
	for _, s := range scopes {
		// Handle scope hierarchies (e.g., "repo" includes "public_repo")
		if s == target {
			return true
		}
		// "repo" scope includes many sub-scopes
		if s == "repo" && (target == "public_repo" || strings.HasPrefix(target, "repo:")) {
			return true
		}
		// "admin:org" includes "read:org" and "write:org"
		if s == "admin:org" && (target == "read:org" || target == "write:org") {
			return true
		}
		if s == "write:org" && target == "read:org" {
			return true
		}
	}
	return false
}

// hasAccessLevel checks if the token has sufficient access level
func hasAccessLevel(have, need string) bool {
	levels := map[string]int{
		"none":  0,
		"read":  1,
		"write": 2,
		"admin": 3,
	}
	return levels[have] >= levels[need]
}

// ============================================================================
// Accessible Repos and Organizations
// ============================================================================

// AccessibleRepo represents a repo the token can access
type AccessibleRepo struct {
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
	Owner    string `json:"owner"`
}

// AccessibleOrg represents an organization the user has access to
type AccessibleOrg struct {
	Login       string           `json:"login"`
	Description string           `json:"description"`
	Repos       []AccessibleRepo `json:"repos"`
}

// AccessibleRepoSummary provides a summary of accessible repos
type AccessibleRepoSummary struct {
	User          string           `json:"user"`
	PersonalRepos []AccessibleRepo `json:"personal_repos"`
	Orgs          []AccessibleOrg  `json:"orgs"`
	TotalRepos    int              `json:"total_repos"`
}

// ListAccessibleRepos returns a detailed list of repos the token can access
func (c *Client) ListAccessibleRepos() (*AccessibleRepoSummary, error) {
	if c.token == "" {
		return nil, fmt.Errorf("no GitHub token available")
	}

	summary := &AccessibleRepoSummary{}

	// Get authenticated user
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get user info: status %d", resp.StatusCode)
	}

	var user struct {
		Login string `json:"login"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	summary.User = user.Login

	// Get user's own repos
	summary.PersonalRepos = c.listUserRepos(user.Login)
	summary.TotalRepos = len(summary.PersonalRepos)

	// Get organizations
	orgsReq, err := http.NewRequest("GET", "https://api.github.com/user/orgs", nil)
	if err != nil {
		return summary, nil // Return partial result
	}
	orgsReq.Header.Set("Authorization", "Bearer "+c.token)
	orgsReq.Header.Set("Accept", "application/vnd.github+json")

	orgsResp, err := c.httpClient.Do(orgsReq)
	if err != nil {
		return summary, nil
	}
	defer orgsResp.Body.Close()

	if orgsResp.StatusCode == 200 {
		var orgs []struct {
			Login       string `json:"login"`
			Description string `json:"description"`
		}
		if err := json.NewDecoder(orgsResp.Body).Decode(&orgs); err == nil {
			for _, org := range orgs {
				repos := c.listOrgRepos(org.Login)
				summary.Orgs = append(summary.Orgs, AccessibleOrg{
					Login:       org.Login,
					Description: org.Description,
					Repos:       repos,
				})
				summary.TotalRepos += len(repos)
			}
		}
	}

	return summary, nil
}

// listUserRepos returns repos owned by the user
func (c *Client) listUserRepos(username string) []AccessibleRepo {
	// Use gh CLI for consistency with auth
	cmd := exec.Command("gh", "repo", "list", username,
		"--json", "nameWithOwner,isPrivate",
		"--limit", "100",
		"--no-archived")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var ghRepos []struct {
		NameWithOwner string `json:"nameWithOwner"`
		IsPrivate     bool   `json:"isPrivate"`
	}
	if err := json.Unmarshal(out, &ghRepos); err != nil {
		return nil
	}

	repos := make([]AccessibleRepo, len(ghRepos))
	for i, r := range ghRepos {
		repos[i] = AccessibleRepo{
			FullName: r.NameWithOwner,
			Private:  r.IsPrivate,
			Owner:    username,
		}
	}
	return repos
}

// listOrgRepos returns repos in an organization
func (c *Client) listOrgRepos(org string) []AccessibleRepo {
	cmd := exec.Command("gh", "repo", "list", org,
		"--json", "nameWithOwner,isPrivate",
		"--limit", "100",
		"--no-archived")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var ghRepos []struct {
		NameWithOwner string `json:"nameWithOwner"`
		IsPrivate     bool   `json:"isPrivate"`
	}
	if err := json.Unmarshal(out, &ghRepos); err != nil {
		return nil
	}

	repos := make([]AccessibleRepo, len(ghRepos))
	for i, r := range ghRepos {
		repos[i] = AccessibleRepo{
			FullName: r.NameWithOwner,
			Private:  r.IsPrivate,
			Owner:    org,
		}
	}
	return repos
}
