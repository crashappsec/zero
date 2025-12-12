// Package github provides GitHub API interactions
package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

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
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		// Try gh auth token
		if out, err := exec.Command("gh", "auth", "token").Output(); err == nil {
			token = strings.TrimSpace(string(out))
		}
	}

	return &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ListOrgRepos returns repositories for an organization
func (c *Client) ListOrgRepos(org string, limit int) ([]Repository, error) {
	// Use gh CLI for simplicity and auth handling
	args := []string{
		"repo", "list", org,
		"--json", "name,nameWithOwner,sshUrl,defaultBranchRef,isPrivate,isArchived,isFork",
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
