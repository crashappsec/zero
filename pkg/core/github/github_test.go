package github

import (
	"testing"
)

func TestProjectID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase", "owner/repo", "owner/repo"},
		{"uppercase", "Owner/Repo", "owner/repo"},
		{"mixed", "MyOrg/MyRepo", "myorg/myrepo"},
		{"already lowercase", "expressjs/express", "expressjs/express"},
		{"single part", "repo", "repo"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProjectID(tt.input)
			if got != tt.expected {
				t.Errorf("ProjectID(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestShortName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"with owner", "owner/repo", "repo"},
		{"no owner", "repo", "repo"},
		{"empty", "", ""},
		{"multiple slashes", "org/sub/repo", "sub"},
		{"expressjs", "expressjs/express", "express"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShortName(tt.input)
			if got != tt.expected {
				t.Errorf("ShortName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestHasScope(t *testing.T) {
	tests := []struct {
		name   string
		scopes []string
		target string
		want   bool
	}{
		{"exact match", []string{"repo", "read:org"}, "repo", true},
		{"no match", []string{"public_repo"}, "repo", false},
		{"repo includes public_repo", []string{"repo"}, "public_repo", true},
		{"admin:org includes read:org", []string{"admin:org"}, "read:org", true},
		{"admin:org includes write:org", []string{"admin:org"}, "write:org", true},
		{"write:org includes read:org", []string{"write:org"}, "read:org", true},
		{"empty scopes", []string{}, "repo", false},
		{"nil target", []string{"repo"}, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasScope(tt.scopes, tt.target)
			if got != tt.want {
				t.Errorf("hasScope(%v, %q) = %v, want %v", tt.scopes, tt.target, got, tt.want)
			}
		})
	}
}

func TestHasAccessLevel(t *testing.T) {
	tests := []struct {
		name string
		have string
		need string
		want bool
	}{
		{"admin has admin", "admin", "admin", true},
		{"admin has write", "admin", "write", true},
		{"admin has read", "admin", "read", true},
		{"write has write", "write", "write", true},
		{"write has read", "write", "read", true},
		{"write lacks admin", "write", "admin", false},
		{"read has read", "read", "read", true},
		{"read lacks write", "read", "write", false},
		{"none has none", "none", "none", true},
		{"none lacks read", "none", "read", false},
		{"unknown level", "unknown", "read", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasAccessLevel(tt.have, tt.need)
			if got != tt.want {
				t.Errorf("hasAccessLevel(%q, %q) = %v, want %v", tt.have, tt.need, got, tt.want)
			}
		})
	}
}

func TestCheckToolAvailability(t *testing.T) {
	// Test with known tools that should exist on most systems
	tools := []string{"git", "nonexistent-tool-xyz"}
	status := CheckToolAvailability(tools)

	// git should exist
	if !status["git"] {
		t.Log("git not found - this is unexpected on most systems")
	}

	// nonexistent tool should not exist
	if status["nonexistent-tool-xyz"] {
		t.Error("nonexistent-tool-xyz should not exist")
	}
}

func TestAggregateReviewerStats(t *testing.T) {
	prs := []PRReviewData{
		{
			PRNumber: 1,
			Author:   "author1",
			Reviews: []Review{
				{Author: "reviewer1", State: "APPROVED"},
				{Author: "reviewer2", State: "CHANGES_REQUESTED"},
			},
		},
		{
			PRNumber: 2,
			Author:   "author2",
			Reviews: []Review{
				{Author: "reviewer1", State: "APPROVED"},
				{Author: "reviewer1", State: "COMMENTED"},
			},
		},
		{
			PRNumber: 3,
			Author:   "author1",
			Reviews: []Review{
				{Author: "", State: "APPROVED"}, // Empty author should be skipped
			},
		},
	}

	stats := AggregateReviewerStats(prs)

	// Check reviewer1 stats
	r1 := stats["reviewer1"]
	if r1 == nil {
		t.Fatal("reviewer1 stats should exist")
	}
	if r1.ReviewsGiven != 3 {
		t.Errorf("reviewer1 ReviewsGiven = %d, want 3", r1.ReviewsGiven)
	}
	if r1.Approvals != 2 {
		t.Errorf("reviewer1 Approvals = %d, want 2", r1.Approvals)
	}
	if r1.Comments != 1 {
		t.Errorf("reviewer1 Comments = %d, want 1", r1.Comments)
	}

	// Check reviewer2 stats
	r2 := stats["reviewer2"]
	if r2 == nil {
		t.Fatal("reviewer2 stats should exist")
	}
	if r2.ChangesRequested != 1 {
		t.Errorf("reviewer2 ChangesRequested = %d, want 1", r2.ChangesRequested)
	}

	// Empty author should not be in stats
	if _, exists := stats[""]; exists {
		t.Error("empty author should not be in stats")
	}
}

func TestTokenInfo(t *testing.T) {
	info := &TokenInfo{
		Type:     "classic",
		Valid:    true,
		Username: "testuser",
		Scopes:   []string{"repo", "read:org"},
	}

	if info.Type != "classic" {
		t.Errorf("Type = %q, want %q", info.Type, "classic")
	}
	if !info.Valid {
		t.Error("Valid should be true")
	}
	if len(info.Scopes) != 2 {
		t.Errorf("Scopes length = %d, want 2", len(info.Scopes))
	}
}

func TestScannerRequirementsExist(t *testing.T) {
	// Verify that scanner requirements are defined for common scanners
	expectedScanners := []string{
		"package-sbom",
		"package-vulns",
		"code-vulns",
		"crypto",
		"sbom",
	}

	for _, scanner := range expectedScanners {
		if _, exists := ScannerRequirements[scanner]; !exists {
			t.Errorf("ScannerRequirements missing entry for %q", scanner)
		}
	}
}

func TestNewClient(t *testing.T) {
	// NewClient should not panic
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
}

func TestClient_HasToken(t *testing.T) {
	// Create client without token
	client := &Client{token: ""}
	if client.HasToken() {
		t.Error("HasToken should return false for empty token")
	}

	// Create client with token
	client.token = "test-token"
	if !client.HasToken() {
		t.Error("HasToken should return true for non-empty token")
	}
}

func TestClient_GetToken(t *testing.T) {
	token := "test-token-123"
	client := &Client{token: token}
	if got := client.GetToken(); got != token {
		t.Errorf("GetToken() = %q, want %q", got, token)
	}
}
