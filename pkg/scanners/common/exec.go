// Package common provides shared utilities for scanner implementations
package common

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// CommandResult holds the result of running an external command
type CommandResult struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
	Duration time.Duration
}

// RunCommand executes an external command with timeout support
func RunCommand(ctx context.Context, name string, args ...string) (*CommandResult, error) {
	start := time.Now()

	cmd := exec.CommandContext(ctx, name, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Disable color output for consistent parsing
	cmd.Env = append(os.Environ(), "NO_COLOR=1", "TERM=dumb")

	err := cmd.Run()
	duration := time.Since(start)

	result := &CommandResult{
		Stdout:   stdout.Bytes(),
		Stderr:   stderr.Bytes(),
		Duration: duration,
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitErr.ExitCode()
		// Some tools exit non-zero when they find issues (e.g., osv-scanner)
		// This is not necessarily an error for us
		return result, nil
	}

	if err != nil {
		return result, fmt.Errorf("running %s: %w", name, err)
	}

	return result, nil
}

// RunCommandWithInput executes a command with stdin input
func RunCommandWithInput(ctx context.Context, input []byte, name string, args ...string) (*CommandResult, error) {
	start := time.Now()

	cmd := exec.CommandContext(ctx, name, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = append(os.Environ(), "NO_COLOR=1", "TERM=dumb")

	err := cmd.Run()
	duration := time.Since(start)

	result := &CommandResult{
		Stdout:   stdout.Bytes(),
		Stderr:   stderr.Bytes(),
		Duration: duration,
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitErr.ExitCode()
		return result, nil
	}

	if err != nil {
		return result, fmt.Errorf("running %s: %w", name, err)
	}

	return result, nil
}

// ToolExists checks if a tool is available in PATH
func ToolExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// ToolVersion gets the version of a tool
func ToolVersion(name string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := RunCommand(ctx, name, "--version")
	if err != nil {
		return "", err
	}

	// Extract first line, first version-like string
	output := strings.TrimSpace(string(result.Stdout))
	lines := strings.Split(output, "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("no version output")
	}

	// Try to find version number
	for _, word := range strings.Fields(lines[0]) {
		// Check if word looks like a version number
		if len(word) > 0 && (word[0] >= '0' && word[0] <= '9') {
			return strings.TrimRight(word, ",;:"), nil
		}
		// Check for vX.Y.Z format
		if strings.HasPrefix(word, "v") && len(word) > 1 && (word[1] >= '0' && word[1] <= '9') {
			return word, nil
		}
	}

	return lines[0], nil
}

// RequireTool returns an error if the tool is not available
func RequireTool(name string) error {
	if !ToolExists(name) {
		return fmt.Errorf("required tool not found: %s", name)
	}
	return nil
}

// RequireTools returns an error if any tool is not available
func RequireTools(names ...string) error {
	var missing []string
	for _, name := range names {
		if !ToolExists(name) {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("required tools not found: %s", strings.Join(missing, ", "))
	}
	return nil
}

// PreferTool returns the first available tool from the list
func PreferTool(names ...string) (string, bool) {
	for _, name := range names {
		if ToolExists(name) {
			return name, true
		}
	}
	return "", false
}

// RunInDir executes a command in a specific directory
func RunInDir(ctx context.Context, dir, name string, args ...string) (*CommandResult, error) {
	start := time.Now()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = append(os.Environ(), "NO_COLOR=1", "TERM=dumb")

	err := cmd.Run()
	duration := time.Since(start)

	result := &CommandResult{
		Stdout:   stdout.Bytes(),
		Stderr:   stderr.Bytes(),
		Duration: duration,
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitErr.ExitCode()
		return result, nil
	}

	if err != nil {
		return result, fmt.Errorf("running %s in %s: %w", name, dir, err)
	}

	return result, nil
}
