package liveapi

import (
	"context"
	"sync"
	"time"
)

// RateLimiter implements a simple token bucket rate limiter
type RateLimiter struct {
	mu       sync.Mutex
	tokens   int
	maxTokens int
	interval time.Duration
	lastAdd  time.Time
}

// NewRateLimiter creates a new rate limiter
// maxRequests is the maximum number of requests allowed in the interval
func NewRateLimiter(maxRequests int, interval time.Duration) *RateLimiter {
	return &RateLimiter{
		tokens:    maxRequests,
		maxTokens: maxRequests,
		interval:  interval,
		lastAdd:   time.Now(),
	}
}

// Wait blocks until a token is available or context is cancelled
func (r *RateLimiter) Wait(ctx context.Context) error {
	for {
		r.mu.Lock()
		r.refill()

		if r.tokens > 0 {
			r.tokens--
			r.mu.Unlock()
			return nil
		}
		r.mu.Unlock()

		// Wait a bit before retrying
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(r.interval / time.Duration(r.maxTokens)):
			// Continue to retry
		}
	}
}

// TryAcquire attempts to acquire a token without blocking
// Returns true if successful
func (r *RateLimiter) TryAcquire() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.refill()

	if r.tokens > 0 {
		r.tokens--
		return true
	}
	return false
}

// refill adds tokens based on elapsed time (must be called with lock held)
func (r *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(r.lastAdd)

	// Calculate tokens to add based on elapsed time
	tokensToAdd := int(elapsed / (r.interval / time.Duration(r.maxTokens)))
	if tokensToAdd > 0 {
		r.tokens += tokensToAdd
		if r.tokens > r.maxTokens {
			r.tokens = r.maxTokens
		}
		r.lastAdd = now
	}
}

// Available returns the number of available tokens
func (r *RateLimiter) Available() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.refill()
	return r.tokens
}

// Reset resets the rate limiter to full capacity
func (r *RateLimiter) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tokens = r.maxTokens
	r.lastAdd = time.Now()
}
