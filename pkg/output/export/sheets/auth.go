// Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
// SPDX-License-Identifier: GPL-3.0

package sheets

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

// AuthConfig holds OAuth configuration
type AuthConfig struct {
	ClientID     string
	ClientSecret string
	TokenPath    string
}

// DefaultTokenPath returns the default path for caching OAuth tokens
func DefaultTokenPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".zero/google-token.json"
	}
	return filepath.Join(home, ".zero", "google-token.json")
}

// Authenticator handles Google OAuth authentication
type Authenticator struct {
	config    *oauth2.Config
	tokenPath string
}

// NewAuthenticator creates a new authenticator with the given configuration
func NewAuthenticator(cfg AuthConfig) *Authenticator {
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Scopes: []string{
			sheets.SpreadsheetsScope,
			sheets.DriveFileScope,
		},
		Endpoint: google.Endpoint,
	}

	tokenPath := cfg.TokenPath
	if tokenPath == "" {
		tokenPath = DefaultTokenPath()
	}

	return &Authenticator{
		config:    oauthConfig,
		tokenPath: tokenPath,
	}
}

// GetClient returns an authenticated HTTP client
func (a *Authenticator) GetClient(ctx context.Context) (*http.Client, error) {
	// Try to load cached token first
	token, err := a.loadToken()
	if err == nil && token.Valid() {
		return a.config.Client(ctx, token), nil
	}

	// If token exists but expired, try to refresh
	if token != nil && !token.Valid() && token.RefreshToken != "" {
		newToken, err := a.config.TokenSource(ctx, token).Token()
		if err == nil {
			if saveErr := a.saveToken(newToken); saveErr != nil {
				fmt.Printf("Warning: Failed to cache refreshed token: %v\n", saveErr)
			}
			return a.config.Client(ctx, newToken), nil
		}
	}

	// Need to authenticate via browser
	token, err = a.browserAuth(ctx)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Cache the token
	if err := a.saveToken(token); err != nil {
		fmt.Printf("Warning: Failed to cache token: %v\n", err)
	}

	return a.config.Client(ctx, token), nil
}

// browserAuth performs OAuth authentication via browser
func (a *Authenticator) browserAuth(ctx context.Context) (*oauth2.Token, error) {
	// Find an available port for the callback server
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, fmt.Errorf("failed to start callback server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// Set redirect URL with the found port
	a.config.RedirectURL = fmt.Sprintf("http://localhost:%d/callback", port)

	// Generate state for CSRF protection
	state := fmt.Sprintf("%d", time.Now().UnixNano())

	// Channel to receive the auth code
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	// Start callback server
	server := &http.Server{Addr: fmt.Sprintf("localhost:%d", port)}
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			errChan <- fmt.Errorf("invalid state parameter")
			http.Error(w, "Invalid state", http.StatusBadRequest)
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			errChan <- fmt.Errorf("no code in callback")
			http.Error(w, "No code received", http.StatusBadRequest)
			return
		}

		// Success page
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `
			<!DOCTYPE html>
			<html>
			<head><title>Zero - Authentication Successful</title></head>
			<body style="font-family: system-ui; text-align: center; padding: 50px;">
				<h1>Authentication Successful</h1>
				<p>You can close this window and return to the terminal.</p>
				<script>window.close();</script>
			</body>
			</html>
		`)
		codeChan <- code
	})

	// Start server in background
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Generate auth URL with explicit response_type=code
	authURL := a.config.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.ApprovalForce,
		oauth2.SetAuthURLParam("response_type", "code"),
	)

	fmt.Println()
	fmt.Println("Opening browser for Google authentication...")
	fmt.Println()
	fmt.Println("If browser doesn't open, visit this URL:")
	fmt.Println(authURL)
	fmt.Println()

	// Open browser
	if err := openBrowser(authURL); err != nil {
		fmt.Printf("Warning: Could not open browser: %v\n", err)
	}

	fmt.Println("Waiting for authentication...")

	// Wait for callback or timeout
	var code string
	select {
	case code = <-codeChan:
		// Success
	case err := <-errChan:
		server.Shutdown(ctx)
		return nil, err
	case <-time.After(5 * time.Minute):
		server.Shutdown(ctx)
		return nil, fmt.Errorf("authentication timed out")
	case <-ctx.Done():
		server.Shutdown(ctx)
		return nil, ctx.Err()
	}

	// Shutdown server
	server.Shutdown(ctx)

	// Exchange code for token
	token, err := a.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	return token, nil
}

// loadToken loads a cached token from disk
func (a *Authenticator) loadToken() (*oauth2.Token, error) {
	data, err := os.ReadFile(a.tokenPath)
	if err != nil {
		return nil, err
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}

	return &token, nil
}

// saveToken saves a token to disk
func (a *Authenticator) saveToken(token *oauth2.Token) error {
	// Ensure directory exists
	dir := filepath.Dir(a.tokenPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(a.tokenPath, data, 0600)
}

// ClearToken removes the cached token
func (a *Authenticator) ClearToken() error {
	if err := os.Remove(a.tokenPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// openBrowser opens a URL in the default browser
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}
