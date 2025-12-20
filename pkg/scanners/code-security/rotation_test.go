package codesecurity

import (
	"testing"
)

func TestNewRotationDatabase(t *testing.T) {
	db := NewRotationDatabase()
	if db == nil {
		t.Fatal("NewRotationDatabase() returned nil")
	}
	if len(db.guides) == 0 {
		t.Error("NewRotationDatabase() returned empty guides")
	}
}

func TestRotationDatabase_GetGuide(t *testing.T) {
	db := NewRotationDatabase()

	tests := []struct {
		secretType   string
		wantPriority string
		wantSteps    bool
	}{
		{"aws_access_key", "immediate", true},
		{"aws_secret_key", "immediate", true},
		{"github_token", "immediate", true},
		{"stripe_secret_key", "immediate", true},
		{"slack_token", "immediate", true},
		{"openai_api_key", "immediate", true},
		{"anthropic_api_key", "immediate", true},
		{"database_credential", "immediate", true},
		{"private_key", "immediate", true},
		{"jwt_secret", "high", true},
		{"google_api_key", "immediate", true},
		{"npm_token", "immediate", true},
		{"unknown_type", "high", true}, // Should return generic
	}

	for _, tt := range tests {
		t.Run(tt.secretType, func(t *testing.T) {
			guide := db.GetGuide(tt.secretType)
			if guide == nil {
				t.Fatalf("GetGuide(%q) returned nil", tt.secretType)
			}
			if guide.Priority != tt.wantPriority {
				t.Errorf("GetGuide(%q).Priority = %q, want %q", tt.secretType, guide.Priority, tt.wantPriority)
			}
			if tt.wantSteps && len(guide.Steps) == 0 {
				t.Errorf("GetGuide(%q).Steps is empty", tt.secretType)
			}
		})
	}
}

func TestGetServiceProvider(t *testing.T) {
	tests := []struct {
		secretType string
		expected   string
	}{
		{"aws_access_key", "aws"},
		{"aws_secret_key", "aws"},
		{"github_token", "github"},
		{"stripe_secret_key", "stripe"},
		{"slack_token", "slack"},
		{"openai_api_key", "openai"},
		{"anthropic_api_key", "anthropic"},
		{"database_credential", "database"},
		{"mysql_password", "database"},
		{"postgres_password", "database"},
		{"private_key", "crypto"},
		{"ssh_private_key", "crypto"},
		{"jwt_secret", "jwt"},
		{"google_api_key", "google"},
		{"gcp_service_account", "gcp"},
		{"azure_secret", "azure"},
		{"npm_token", "npm"},
		{"pypi_token", "pypi"},
		{"heroku_api_key", "heroku"},
		{"vercel_token", "vercel"},
		{"datadog_api_key", "datadog"},
		{"unknown_type", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.secretType, func(t *testing.T) {
			got := GetServiceProvider(tt.secretType)
			if got != tt.expected {
				t.Errorf("GetServiceProvider(%q) = %q, want %q", tt.secretType, got, tt.expected)
			}
		})
	}
}

func TestEnrichWithRotation(t *testing.T) {
	db := NewRotationDatabase()

	findings := []SecretFinding{
		{Type: "aws_access_key", File: "config.go", Line: 10},
		{Type: "github_token", File: "env.go", Line: 20},
		{Type: "unknown_secret", File: "secret.go", Line: 30},
	}

	enriched := EnrichWithRotation(findings, db)

	if len(enriched) != 3 {
		t.Fatalf("EnrichWithRotation returned %d findings, want 3", len(enriched))
	}

	// Check AWS key got rotation guidance
	if enriched[0].Rotation == nil {
		t.Error("AWS access key finding should have rotation guidance")
	}
	if enriched[0].ServiceProvider != "aws" {
		t.Errorf("AWS access key ServiceProvider = %q, want %q", enriched[0].ServiceProvider, "aws")
	}

	// Check GitHub token got rotation guidance
	if enriched[1].Rotation == nil {
		t.Error("GitHub token finding should have rotation guidance")
	}
	if enriched[1].ServiceProvider != "github" {
		t.Errorf("GitHub token ServiceProvider = %q, want %q", enriched[1].ServiceProvider, "github")
	}

	// Check unknown got generic guidance
	if enriched[2].Rotation == nil {
		t.Error("Unknown secret finding should have generic rotation guidance")
	}
}

func TestRotationGuide_HasRequiredFields(t *testing.T) {
	db := NewRotationDatabase()

	// Check that all guides have required fields
	knownTypes := []string{
		"aws_access_key", "github_token", "stripe_secret_key",
		"slack_token", "openai_api_key", "database_credential",
		"private_key", "jwt_secret", "generic_secret",
	}

	for _, secretType := range knownTypes {
		guide := db.GetGuide(secretType)

		if guide.Priority == "" {
			t.Errorf("Guide for %q has empty Priority", secretType)
		}

		validPriorities := map[string]bool{
			"immediate": true,
			"high":      true,
			"medium":    true,
			"low":       true,
		}
		if !validPriorities[guide.Priority] {
			t.Errorf("Guide for %q has invalid Priority: %q", secretType, guide.Priority)
		}

		if len(guide.Steps) == 0 {
			t.Errorf("Guide for %q has no Steps", secretType)
		}
	}
}

func TestRotationGuide_HasRotationURL(t *testing.T) {
	db := NewRotationDatabase()

	// These types should have rotation URLs
	typesWithURLs := []string{
		"aws_access_key", "github_token", "stripe_secret_key",
		"slack_token", "openai_api_key", "anthropic_api_key",
		"google_api_key", "npm_token",
	}

	for _, secretType := range typesWithURLs {
		guide := db.GetGuide(secretType)
		if guide.RotationURL == "" {
			t.Errorf("Guide for %q should have RotationURL", secretType)
		}
	}
}
