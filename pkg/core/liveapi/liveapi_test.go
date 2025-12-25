package liveapi

import (
	"context"
	"testing"
	"time"
)

// ============================================================================
// Cache Tests
// ============================================================================

func TestCache_SetAndGet(t *testing.T) {
	cache := NewCache(1 * time.Hour)
	key := "test-key"
	data := []byte("test-data")

	cache.Set(key, data)
	got, ok := cache.Get(key)

	if !ok {
		t.Fatal("Get should return true for existing key")
	}
	if string(got) != string(data) {
		t.Errorf("Get() = %q, want %q", got, data)
	}
}

func TestCache_GetMissing(t *testing.T) {
	cache := NewCache(1 * time.Hour)

	_, ok := cache.Get("nonexistent")
	if ok {
		t.Error("Get should return false for missing key")
	}
}

func TestCache_Expiration(t *testing.T) {
	cache := NewCache(50 * time.Millisecond)
	key := "test-key"
	data := []byte("test-data")

	cache.Set(key, data)

	// Should exist immediately
	if _, ok := cache.Get(key); !ok {
		t.Error("Key should exist immediately after set")
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should be expired
	if _, ok := cache.Get(key); ok {
		t.Error("Key should be expired after TTL")
	}
}

func TestCache_Delete(t *testing.T) {
	cache := NewCache(1 * time.Hour)
	key := "test-key"
	data := []byte("test-data")

	cache.Set(key, data)
	cache.Delete(key)

	if _, ok := cache.Get(key); ok {
		t.Error("Key should not exist after delete")
	}
}

func TestCache_Clear(t *testing.T) {
	cache := NewCache(1 * time.Hour)

	cache.Set("key1", []byte("data1"))
	cache.Set("key2", []byte("data2"))

	if cache.Size() != 2 {
		t.Errorf("Size() = %d, want 2", cache.Size())
	}

	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Size() after Clear() = %d, want 0", cache.Size())
	}
}

func TestCache_Size(t *testing.T) {
	cache := NewCache(1 * time.Hour)

	if cache.Size() != 0 {
		t.Errorf("Empty cache Size() = %d, want 0", cache.Size())
	}

	cache.Set("key1", []byte("data1"))
	cache.Set("key2", []byte("data2"))

	if cache.Size() != 2 {
		t.Errorf("Size() = %d, want 2", cache.Size())
	}
}

// ============================================================================
// RateLimiter Tests
// ============================================================================

func TestRateLimiter_TryAcquire(t *testing.T) {
	limiter := NewRateLimiter(2, 1*time.Second)

	// Should succeed twice
	if !limiter.TryAcquire() {
		t.Error("First TryAcquire should succeed")
	}
	if !limiter.TryAcquire() {
		t.Error("Second TryAcquire should succeed")
	}

	// Third should fail
	if limiter.TryAcquire() {
		t.Error("Third TryAcquire should fail (rate limited)")
	}
}

func TestRateLimiter_Available(t *testing.T) {
	limiter := NewRateLimiter(5, 1*time.Second)

	if limiter.Available() != 5 {
		t.Errorf("Available() = %d, want 5", limiter.Available())
	}

	limiter.TryAcquire()
	limiter.TryAcquire()

	if limiter.Available() != 3 {
		t.Errorf("Available() = %d, want 3", limiter.Available())
	}
}

func TestRateLimiter_Reset(t *testing.T) {
	limiter := NewRateLimiter(3, 1*time.Second)

	limiter.TryAcquire()
	limiter.TryAcquire()
	limiter.TryAcquire()

	if limiter.Available() != 0 {
		t.Errorf("Available() before reset = %d, want 0", limiter.Available())
	}

	limiter.Reset()

	if limiter.Available() != 3 {
		t.Errorf("Available() after reset = %d, want 3", limiter.Available())
	}
}

func TestRateLimiter_Wait_Success(t *testing.T) {
	limiter := NewRateLimiter(1, 1*time.Second)

	ctx := context.Background()
	err := limiter.Wait(ctx)
	if err != nil {
		t.Errorf("Wait() returned error: %v", err)
	}
}

func TestRateLimiter_Wait_ContextCancelled(t *testing.T) {
	limiter := NewRateLimiter(1, 1*time.Second)

	// Exhaust the limiter
	limiter.TryAcquire()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := limiter.Wait(ctx)
	if err == nil {
		t.Error("Wait() should return error when context is cancelled")
	}
}

func TestRateLimiter_Refill(t *testing.T) {
	limiter := NewRateLimiter(2, 100*time.Millisecond)

	// Exhaust tokens
	limiter.TryAcquire()
	limiter.TryAcquire()

	if limiter.Available() != 0 {
		t.Errorf("Available() = %d, want 0", limiter.Available())
	}

	// Wait for refill
	time.Sleep(150 * time.Millisecond)

	// Should have tokens again
	if limiter.Available() < 1 {
		t.Errorf("Available() after refill = %d, want >= 1", limiter.Available())
	}
}

// ============================================================================
// OSV Client Tests
// ============================================================================

func TestNewOSVClient(t *testing.T) {
	client := NewOSVClient()
	if client == nil {
		t.Fatal("NewOSVClient returned nil")
	}
	if client.Client == nil {
		t.Error("Client.Client should not be nil")
	}
}

func TestNewOSVClientWithTimeout(t *testing.T) {
	timeout := 60 * time.Second
	client := NewOSVClientWithTimeout(timeout)
	if client == nil {
		t.Fatal("NewOSVClientWithTimeout returned nil")
	}
}

func TestVulnerability_GetCVEs(t *testing.T) {
	v := &Vulnerability{
		ID:      "GHSA-1234",
		Aliases: []string{"CVE-2023-1234", "GHSA-5678", "CVE-2023-5678"},
	}

	cves := v.GetCVEs()
	if len(cves) != 2 {
		t.Errorf("GetCVEs() returned %d CVEs, want 2", len(cves))
	}
	if cves[0] != "CVE-2023-1234" {
		t.Errorf("First CVE = %q, want %q", cves[0], "CVE-2023-1234")
	}
}

func TestVulnerability_GetFixedVersion(t *testing.T) {
	v := &Vulnerability{
		Affected: []Affected{
			{
				Package: Package{Name: "lodash", Ecosystem: "npm"},
				Ranges: []Range{
					{
						Type: "SEMVER",
						Events: []Event{
							{Introduced: "0"},
							{Fixed: "4.17.21"},
						},
					},
				},
			},
		},
	}

	fixed := v.GetFixedVersion("npm", "lodash")
	if fixed != "4.17.21" {
		t.Errorf("GetFixedVersion() = %q, want %q", fixed, "4.17.21")
	}

	// Non-matching package should return empty
	fixed = v.GetFixedVersion("npm", "express")
	if fixed != "" {
		t.Errorf("GetFixedVersion() for non-matching package = %q, want empty", fixed)
	}
}

func TestVulnerability_GetHighestSeverity(t *testing.T) {
	tests := []struct {
		name     string
		severity []OSVSeverity
		want     string
	}{
		{
			name:     "no severity",
			severity: nil,
			want:     "unknown",
		},
		{
			name: "with CVSS_V3",
			severity: []OSVSeverity{
				{Type: "CVSS_V3", Score: "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H"},
			},
			want: "medium", // parseCVSSSeverity returns "medium" as default
		},
		{
			name: "without CVSS_V3",
			severity: []OSVSeverity{
				{Type: "CVSS_V2", Score: "6.5"},
			},
			want: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Vulnerability{Severity: tt.severity}
			got := v.GetHighestSeverity()
			if got != tt.want {
				t.Errorf("GetHighestSeverity() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestQueryRequest(t *testing.T) {
	req := QueryRequest{
		Package: &PackageQuery{
			Name:      "lodash",
			Ecosystem: "npm",
		},
		Version: "4.17.20",
	}

	if req.Package.Name != "lodash" {
		t.Error("Package name not set correctly")
	}
	if req.Version != "4.17.20" {
		t.Error("Version not set correctly")
	}
}

func TestBatchQueryRequest(t *testing.T) {
	req := BatchQueryRequest{
		Queries: []QueryRequest{
			{Package: &PackageQuery{Name: "lodash", Ecosystem: "npm"}, Version: "4.17.20"},
			{Package: &PackageQuery{Name: "express", Ecosystem: "npm"}, Version: "4.18.0"},
		},
	}

	if len(req.Queries) != 2 {
		t.Errorf("Queries length = %d, want 2", len(req.Queries))
	}
}
