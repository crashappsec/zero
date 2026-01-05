// Package credentials manages API keys and tokens for Zero
package credentials

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Credentials holds API keys and tokens
type Credentials struct {
	GitHubToken    string `json:"github_token,omitempty"`
	AnthropicKey   string `json:"anthropic_api_key,omitempty"`
}

// Source indicates where a credential came from
type Source string

const (
	SourceNone       Source = "not found"
	SourceEnvVar     Source = "environment variable"
	SourceConfigFile Source = "~/.zero/credentials.json"
	SourceGHCLI      Source = "gh auth token"
)

// CredentialInfo contains a credential value and its source
type CredentialInfo struct {
	Value  string
	Source Source
	Valid  bool
}

// credentialsPath returns the path to the credentials file
func credentialsPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".zero", "credentials.json")
}

// Load reads credentials from file
func Load() (*Credentials, error) {
	path := credentialsPath()
	if path == "" {
		return &Credentials{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Credentials{}, nil
		}
		return nil, fmt.Errorf("reading credentials: %w", err)
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("parsing credentials: %w", err)
	}

	return &creds, nil
}

// Save writes credentials to file with restricted permissions
func Save(creds *Credentials) error {
	path := credentialsPath()
	if path == "" {
		return fmt.Errorf("could not determine home directory")
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding credentials: %w", err)
	}

	// Write with restricted permissions (owner read/write only)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing credentials: %w", err)
	}

	return nil
}

// GetGitHubToken returns the GitHub token from the best available source
// Priority: 1. GITHUB_TOKEN env var, 2. credentials.json, 3. gh auth token
func GetGitHubToken() CredentialInfo {
	// 1. Check environment variable
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return CredentialInfo{
			Value:  token,
			Source: SourceEnvVar,
			Valid:  true,
		}
	}

	// 2. Check credentials file
	if creds, err := Load(); err == nil && creds.GitHubToken != "" {
		return CredentialInfo{
			Value:  creds.GitHubToken,
			Source: SourceConfigFile,
			Valid:  true,
		}
	}

	// 3. Try gh auth token
	if out, err := exec.Command("gh", "auth", "token").Output(); err == nil {
		token := strings.TrimSpace(string(out))
		if token != "" {
			return CredentialInfo{
				Value:  token,
				Source: SourceGHCLI,
				Valid:  true,
			}
		}
	}

	return CredentialInfo{
		Source: SourceNone,
		Valid:  false,
	}
}

// GetAnthropicKey returns the Anthropic API key from the best available source
// Priority: 1. ANTHROPIC_API_KEY env var, 2. credentials.json
func GetAnthropicKey() CredentialInfo {
	// 1. Check environment variable
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		return CredentialInfo{
			Value:  key,
			Source: SourceEnvVar,
			Valid:  true,
		}
	}

	// 2. Check credentials file
	if creds, err := Load(); err == nil && creds.AnthropicKey != "" {
		return CredentialInfo{
			Value:  creds.AnthropicKey,
			Source: SourceConfigFile,
			Valid:  true,
		}
	}

	return CredentialInfo{
		Source: SourceNone,
		Valid:  false,
	}
}

// SetGitHubToken saves the GitHub token to the credentials file
func SetGitHubToken(token string) error {
	creds, err := Load()
	if err != nil {
		creds = &Credentials{}
	}
	creds.GitHubToken = token
	return Save(creds)
}

// SetAnthropicKey saves the Anthropic API key to the credentials file
func SetAnthropicKey(key string) error {
	creds, err := Load()
	if err != nil {
		creds = &Credentials{}
	}
	creds.AnthropicKey = key
	return Save(creds)
}

// ClearCredentials removes all stored credentials
func ClearCredentials() error {
	path := credentialsPath()
	if path == "" {
		return nil
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing credentials: %w", err)
	}
	return nil
}

// MaskToken returns a masked version of a token for display
func MaskToken(token string) string {
	if len(token) <= 8 {
		return "***"
	}
	return token[:4] + "..." + token[len(token)-4:]
}
