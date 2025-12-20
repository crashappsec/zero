// Package liveapi provides clients for live API queries (e.g., OSV)
package liveapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is a generic API client with caching and rate limiting
type Client struct {
	BaseURL     string
	HTTPClient  *http.Client
	Cache       *Cache
	RateLimiter *RateLimiter
	UserAgent   string
}

// ClientOption configures a Client
type ClientOption func(*Client)

// WithTimeout sets the HTTP client timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.HTTPClient.Timeout = timeout
	}
}

// WithCache enables caching with the specified TTL
func WithCache(ttl time.Duration) ClientOption {
	return func(c *Client) {
		c.Cache = NewCache(ttl)
	}
}

// WithRateLimit sets rate limiting
func WithRateLimit(requestsPerSecond int) ClientOption {
	return func(c *Client) {
		c.RateLimiter = NewRateLimiter(requestsPerSecond, time.Second)
	}
}

// WithUserAgent sets the User-Agent header
func WithUserAgent(ua string) ClientOption {
	return func(c *Client) {
		c.UserAgent = ua
	}
}

// NewClient creates a new API client
func NewClient(baseURL string, opts ...ClientOption) *Client {
	c := &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Cache:       NewCache(15 * time.Minute),
		RateLimiter: NewRateLimiter(10, time.Second),
		UserAgent:   "Zero-Scanner/1.0",
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Get performs a GET request
func (c *Client) Get(ctx context.Context, path string, result any) error {
	return c.doRequest(ctx, "GET", path, nil, result)
}

// Post performs a POST request with JSON body
func (c *Client) Post(ctx context.Context, path string, body any, result any) error {
	return c.doRequest(ctx, "POST", path, body, result)
}

// Query performs a cached, rate-limited API query (POST)
func (c *Client) Query(ctx context.Context, path string, body any, result any) error {
	// Check cache first
	cacheKey := c.cacheKey(path, body)
	if c.Cache != nil {
		if cached, ok := c.Cache.Get(cacheKey); ok {
			return json.Unmarshal(cached, result)
		}
	}

	// Wait for rate limiter
	if c.RateLimiter != nil {
		if err := c.RateLimiter.Wait(ctx); err != nil {
			return fmt.Errorf("rate limit: %w", err)
		}
	}

	// Make request
	respBody, err := c.doRequestRaw(ctx, "POST", path, body)
	if err != nil {
		return err
	}

	// Cache the response
	if c.Cache != nil {
		c.Cache.Set(cacheKey, respBody)
	}

	return json.Unmarshal(respBody, result)
}

func (c *Client) doRequest(ctx context.Context, method, path string, body any, result any) error {
	respBody, err := c.doRequestRaw(ctx, method, path, body)
	if err != nil {
		return err
	}
	if result != nil {
		return json.Unmarshal(respBody, result)
	}
	return nil
}

func (c *Client) doRequestRaw(ctx context.Context, method, path string, body any) ([]byte, error) {
	url := c.BaseURL + path

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       string(respBody),
		}
	}

	return respBody, nil
}

func (c *Client) cacheKey(path string, body any) string {
	data, _ := json.Marshal(body)
	return path + ":" + string(data)
}

// APIError represents an API error response
type APIError struct {
	StatusCode int
	Status     string
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %s: %s", e.Status, e.Body)
}

// IsNotFound returns true if the error is a 404
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == 404
}

// IsRateLimited returns true if the error is a 429
func (e *APIError) IsRateLimited() bool {
	return e.StatusCode == 429
}
