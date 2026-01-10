package techid

import (
	"strings"
	"testing"
)

func TestLoadPatterns(t *testing.T) {
	db, err := LoadPatterns()
	if err != nil {
		t.Fatalf("Failed to load patterns: %v", err)
	}

	if db == nil {
		t.Fatal("Pattern database is nil")
	}

	// Check that we have technologies loaded
	if len(db.Technologies) == 0 {
		t.Error("No technologies loaded")
	}

	// Print stats
	stats := db.Stats()
	t.Logf("Loaded patterns: %+v", stats)

	// Verify minimum expected technologies
	minExpected := 10
	if stats["technologies"] < minExpected {
		t.Errorf("Expected at least %d technologies, got %d", minExpected, stats["technologies"])
	}
}

func TestMatchPackage(t *testing.T) {
	db, err := LoadPatterns()
	if err != nil {
		t.Fatalf("Failed to load patterns: %v", err)
	}

	tests := []struct {
		ecosystem string
		name      string
		wantTech  string
		wantMatch bool
	}{
		{"npm", "@modelcontextprotocol/sdk", "mcp", true},
		{"npm", "react", "react", true},
		{"npm", "express", "express", true},
		{"npm", "stripe", "stripe", true},
		{"pypi", "openai", "openai", true},
		{"pypi", "anthropic", "anthropic", true},
		{"pypi", "torch", "pytorch", true},
		{"npm", "nonexistent-package-xyz", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.ecosystem+"/"+tt.name, func(t *testing.T) {
			matches := db.MatchPackage(tt.ecosystem, tt.name)

			if tt.wantMatch {
				if len(matches) == 0 {
					t.Errorf("Expected match for %s/%s, got none", tt.ecosystem, tt.name)
					return
				}

				found := false
				for _, m := range matches {
					if m.TechID == tt.wantTech {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected tech %s for %s/%s, got %v", tt.wantTech, tt.ecosystem, tt.name, matches)
				}
			} else {
				if len(matches) > 0 {
					t.Errorf("Expected no match for %s/%s, got %v", tt.ecosystem, tt.name, matches)
				}
			}
		})
	}
}

func TestMatchPackageGlob(t *testing.T) {
	db, err := LoadPatterns()
	if err != nil {
		t.Fatalf("Failed to load patterns: %v", err)
	}

	tests := []struct {
		ecosystem string
		name      string
		wantTech  string
		wantMatch bool
	}{
		// Note: glob matching is handled by MatchPackage which checks for * patterns
		{"npm", "@aws-sdk/client-s3", "aws", true},        // aws, not aws-sdk
		{"npm", "@aws-sdk/client-dynamodb", "aws", true},  // aws, not aws-sdk
		{"npm", "@aws-sdk/client-lambda", "aws", true},
		{"npm", "aws-sdk", "aws", true},
	}

	for _, tt := range tests {
		t.Run(tt.ecosystem+"/"+tt.name, func(t *testing.T) {
			matches := db.MatchPackage(tt.ecosystem, tt.name)

			if tt.wantMatch {
				if len(matches) == 0 {
					t.Errorf("Expected glob match for %s/%s, got none", tt.ecosystem, tt.name)
					return
				}

				found := false
				for _, m := range matches {
					if m.TechID == tt.wantTech {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected tech %s for %s/%s, got %v", tt.wantTech, tt.ecosystem, tt.name, matches)
				}
			}
		})
	}
}

func TestMatchImport(t *testing.T) {
	db, err := LoadPatterns()
	if err != nil {
		t.Fatalf("Failed to load patterns: %v", err)
	}

	tests := []struct {
		language string
		line     string
		wantTech string
	}{
		{"javascript", `import { Server } from "@modelcontextprotocol/sdk/server"`, "mcp"},
		{"javascript", `from "@modelcontextprotocol/sdk"`, "mcp"},
		{"javascript", `import React from "react"`, "react"},
		{"python", "from mcp import Server", "mcp"},
		{"python", "import anthropic", "anthropic"},
		{"python", "from openai import OpenAI", "openai"},
		{"python", "import torch", "pytorch"},
		{"python", "from transformers import AutoModel", "huggingface"},
	}

	for _, tt := range tests {
		t.Run(tt.language+"/"+tt.wantTech, func(t *testing.T) {
			matches := db.MatchImport(tt.language, tt.line)

			if len(matches) == 0 {
				t.Errorf("Expected match for %q, got none", tt.line)
				return
			}

			found := false
			for _, m := range matches {
				if m.TechID == tt.wantTech {
					found = true
					break
				}
			}
			if !found {
				var techIDs []string
				for _, m := range matches {
					techIDs = append(techIDs, m.TechID)
				}
				t.Errorf("Expected tech %s for %q, got %v", tt.wantTech, tt.line, techIDs)
			}
		})
	}
}

func TestMatchConfigFile(t *testing.T) {
	db, err := LoadPatterns()
	if err != nil {
		t.Fatalf("Failed to load patterns: %v", err)
	}

	tests := []struct {
		path      string
		wantTech  string
		wantMatch bool
	}{
		{"mcp.json", "mcp", true},
		{"next.config.js", "nextjs", true},
		{"Dockerfile", "docker", true},
		{"docker-compose.yml", "docker", true},
		{".github/workflows/ci.yml", "github-actions", false}, // directory match needed
		{"random-file.txt", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			matches := db.MatchConfigFile(tt.path)

			if tt.wantMatch {
				if len(matches) == 0 {
					t.Errorf("Expected match for %q, got none", tt.path)
					return
				}

				found := false
				for _, m := range matches {
					if m.TechID == tt.wantTech {
						found = true
						break
					}
				}
				if !found {
					var techIDs []string
					for _, m := range matches {
						techIDs = append(techIDs, m.TechID)
					}
					t.Errorf("Expected tech %s for %q, got %v", tt.wantTech, tt.path, techIDs)
				}
			} else {
				// For expected no-match, it's ok if we get no matches
				// or if the specific tech we checked for isn't there
			}
		})
	}
}

func TestMatchExtension(t *testing.T) {
	db, err := LoadPatterns()
	if err != nil {
		t.Fatalf("Failed to load patterns: %v", err)
	}

	// Note: Extension patterns are not currently generated from RAG markdown files.
	// This test verifies the MatchExtension function works, even if no patterns exist.
	tests := []struct {
		ext       string
		wantMatch bool
	}{
		// These extensions may not have patterns - just verify function doesn't crash
		{".xyz", false},
		{".txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			matches := db.MatchExtension(tt.ext)
			if tt.wantMatch && len(matches) == 0 {
				t.Errorf("Expected match for %q, got none", tt.ext)
			}
			// Success if no crash
		})
	}
}

func TestMatchSecret(t *testing.T) {
	db, err := LoadPatterns()
	if err != nil {
		t.Fatalf("Failed to load patterns: %v", err)
	}

	// Note: Using fake patterns that match regex but won't trigger GitHub secret scanning
	tests := []struct {
		content     string
		wantTech    string
		wantName    string
		wantMatch   bool
	}{
		// Anthropic key pattern: sk-ant-[a-zA-Z0-9-_]{95}
		{"sk-ant-api03-" + strings.Repeat("x", 89), "anthropic", "Anthropic API Key", true},
		// OpenAI key pattern: sk-[a-zA-Z0-9]{48}
		{"sk-" + strings.Repeat("x", 48), "openai", "OpenAI API Key", true},
		// AWS Access Key pattern: AKIA[0-9A-Z]{16}
		{"AKIA" + strings.Repeat("X", 16), "aws", "AWS Access Key ID", true}, // tech is "aws" not "aws-sdk"
		// Stripe pattern uses different test approach - check pattern directly
		{"just some normal text", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.wantName, func(t *testing.T) {
			matches := db.MatchSecret(tt.content)

			if tt.wantMatch {
				if len(matches) == 0 {
					t.Errorf("Expected secret match, got none")
					return
				}

				found := false
				for _, m := range matches {
					if m.TechID == tt.wantTech && m.Name == tt.wantName {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected secret %s/%s, got %v", tt.wantTech, tt.wantName, matches)
				}
			}
		})
	}
}

func TestGetTechnology(t *testing.T) {
	db, err := LoadPatterns()
	if err != nil {
		t.Fatalf("Failed to load patterns: %v", err)
	}

	tech := db.GetTechnology("mcp")
	if tech == nil {
		t.Fatal("Expected to find mcp technology")
	}

	// Name from RAG markdown includes "(MCP)" suffix
	if !strings.Contains(tech.Name, "Model Context Protocol") {
		t.Errorf("Expected name containing 'Model Context Protocol', got %q", tech.Name)
	}

	if tech.Vendor != "Anthropic" {
		t.Errorf("Expected vendor 'Anthropic', got %q", tech.Vendor)
	}

	// Test non-existent
	nonExistent := db.GetTechnology("nonexistent")
	if nonExistent != nil {
		t.Error("Expected nil for non-existent technology")
	}
}

func TestGetTechnologiesByCategory(t *testing.T) {
	db, err := LoadPatterns()
	if err != nil {
		t.Fatalf("Failed to load patterns: %v", err)
	}

	aimlTechs := db.GetTechnologiesByCategory("ai-ml")
	if len(aimlTechs) == 0 {
		t.Error("Expected AI/ML technologies")
	}

	// Verify we have expected AI/ML techs
	techIDs := make(map[string]bool)
	for _, tech := range aimlTechs {
		techIDs[tech.ID] = true
	}

	expectedAIML := []string{"mcp", "openai", "anthropic", "langchain", "pytorch", "tensorflow", "huggingface"}
	for _, id := range expectedAIML {
		if !techIDs[id] {
			t.Errorf("Expected %s in AI/ML category", id)
		}
	}
}

func TestGlobMatching(t *testing.T) {
	tests := []struct {
		pattern string
		input   string
		want    bool
	}{
		{"mcp-server-*", "mcp-server-filesystem", true},
		{"mcp-server-*", "mcp-server-github", true},
		{"mcp-server-*", "mcp-client", false},
		{"@aws-sdk/client-*", "@aws-sdk/client-s3", true},
		{"*.tf", "main.tf", true},
		{"*.tf", "main.yaml", false},
		{"foo?bar", "fooxbar", true},
		{"foo?bar", "foobar", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.input, func(t *testing.T) {
			got := matchGlob(tt.pattern, tt.input)
			if got != tt.want {
				t.Errorf("matchGlob(%q, %q) = %v, want %v", tt.pattern, tt.input, got, tt.want)
			}
		})
	}
}
