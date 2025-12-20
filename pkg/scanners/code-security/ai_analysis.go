package codesecurity

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// AIAnalyzer uses Claude to analyze findings for false positives
type AIAnalyzer struct {
	config    AIAnalysisConfig
	apiKey    string
	client    *http.Client
	repoPath  string
}

// NewAIAnalyzer creates a new AI analyzer
func NewAIAnalyzer(config AIAnalysisConfig, repoPath string) *AIAnalyzer {
	return &AIAnalyzer{
		config:   config,
		apiKey:   os.Getenv("ANTHROPIC_API_KEY"),
		client:   &http.Client{Timeout: 60 * time.Second},
		repoPath: repoPath,
	}
}

// IsAvailable checks if the AI analyzer is available (has API key)
func (a *AIAnalyzer) IsAvailable() bool {
	return a.apiKey != ""
}

// AnalyzeFindings analyzes findings for false positives
func (a *AIAnalyzer) AnalyzeFindings(ctx context.Context, findings []SecretFinding) ([]SecretFinding, error) {
	if !a.IsAvailable() {
		return findings, nil // Graceful degradation - return unchanged
	}

	// Limit number of findings to analyze
	toAnalyze := findings
	if len(toAnalyze) > a.config.MaxFindings {
		toAnalyze = toAnalyze[:a.config.MaxFindings]
	}

	// Filter to only analyze medium+ severity
	var filtered []SecretFinding
	for _, f := range toAnalyze {
		if f.Severity == "critical" || f.Severity == "high" || f.Severity == "medium" {
			filtered = append(filtered, f)
		}
	}

	if len(filtered) == 0 {
		return findings, nil
	}

	// Analyze each finding
	for i, f := range filtered {
		// Get file context
		context, err := a.getFileContext(f.File, f.Line)
		if err != nil {
			continue
		}

		// Analyze with Claude
		confidence, reasoning, isFP, err := a.analyzeWithClaude(ctx, f, context)
		if err != nil {
			continue
		}

		// Update the finding in the original slice
		for j := range findings {
			if findings[j].File == f.File && findings[j].Line == f.Line {
				findings[j].AIConfidence = confidence
				findings[j].AIReasoning = reasoning
				if confidence >= a.config.ConfidenceThreshold {
					findings[j].IsFalsePositive = &isFP
				}
				break
			}
		}

		// Rate limit: be kind to the API
		if i < len(filtered)-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	return findings, nil
}

// getFileContext reads surrounding lines for context
func (a *AIAnalyzer) getFileContext(file string, line int) (string, error) {
	fullPath := file
	if !strings.HasPrefix(file, "/") {
		fullPath = a.repoPath + "/" + file
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")

	// Get 5 lines before and after
	start := line - 6
	if start < 0 {
		start = 0
	}
	end := line + 5
	if end > len(lines) {
		end = len(lines)
	}

	var contextLines []string
	for i := start; i < end; i++ {
		prefix := "  "
		if i+1 == line {
			prefix = "> "
		}
		contextLines = append(contextLines, fmt.Sprintf("%s%d: %s", prefix, i+1, lines[i]))
	}

	return strings.Join(contextLines, "\n"), nil
}

// analyzeWithClaude calls the Claude API to analyze a finding
func (a *AIAnalyzer) analyzeWithClaude(ctx context.Context, finding SecretFinding, fileContext string) (float64, string, bool, error) {
	prompt := fmt.Sprintf(`Analyze this potential secret detection finding and determine if it's a FALSE POSITIVE or a REAL SECRET.

Finding Details:
- Type: %s
- File: %s
- Line: %d
- Severity: %s
- Detection Rule: %s
- Snippet: %s

File Context:
%s

Analyze whether this is:
1. A FALSE POSITIVE - Not a real secret (examples: test/mock data, placeholder values, example keys, documentation, already-rotated credentials)
2. A REAL SECRET - Actual credentials that need rotation

Consider:
- Is this in a test file, example, or documentation?
- Does the value look like a placeholder (contains "example", "test", "xxx", etc.)?
- Is this a well-known example key (like AWS's AKIAIOSFODNN7EXAMPLE)?
- Could this be dead code or commented out?
- Does the surrounding context suggest this is configuration for a test environment?

Respond in JSON format:
{
  "is_false_positive": true/false,
  "confidence": 0.0-1.0,
  "reasoning": "Brief explanation of your determination"
}`,
		finding.Type,
		finding.File,
		finding.Line,
		finding.Severity,
		finding.RuleID,
		finding.Snippet,
		fileContext,
	)

	response, err := a.callClaudeAPI(ctx, prompt)
	if err != nil {
		return 0, "", false, err
	}

	// Parse response
	var result struct {
		IsFalsePositive bool    `json:"is_false_positive"`
		Confidence      float64 `json:"confidence"`
		Reasoning       string  `json:"reasoning"`
	}

	// Try to extract JSON from response
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")
	if jsonStart != -1 && jsonEnd != -1 && jsonEnd > jsonStart {
		jsonStr := response[jsonStart : jsonEnd+1]
		if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
			return 0, "", false, err
		}
	} else {
		return 0, "", false, fmt.Errorf("no JSON in response")
	}

	return result.Confidence, result.Reasoning, result.IsFalsePositive, nil
}

// claudeRequest represents a request to the Claude API
type claudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []claudeMessage `json:"messages"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// callClaudeAPI makes a request to the Anthropic Claude API
func (a *AIAnalyzer) callClaudeAPI(ctx context.Context, prompt string) (string, error) {
	reqBody := claudeRequest{
		Model:     "claude-sonnet-4-20250514",
		MaxTokens: 500,
		Messages: []claudeMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error: %s", string(body))
	}

	var claudeResp claudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return "", err
	}

	if claudeResp.Error != nil {
		return "", fmt.Errorf("API error: %s", claudeResp.Error.Message)
	}

	if len(claudeResp.Content) == 0 {
		return "", fmt.Errorf("empty response")
	}

	return claudeResp.Content[0].Text, nil
}

// CountFalsePositives counts findings marked as false positives
func CountFalsePositives(findings []SecretFinding) (falsePositives, confirmed int) {
	for _, f := range findings {
		if f.IsFalsePositive != nil {
			if *f.IsFalsePositive {
				falsePositives++
			} else {
				confirmed++
			}
		}
	}
	return
}
