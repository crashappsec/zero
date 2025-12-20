package feeds

import (
	"testing"
	"time"
)

func TestFrequencyDuration(t *testing.T) {
	tests := []struct {
		freq     Frequency
		expected time.Duration
	}{
		{FreqAlways, 0},
		{FreqHourly, time.Hour},
		{FreqDaily, 24 * time.Hour},
		{FreqWeekly, 7 * 24 * time.Hour},
		{FreqMonthly, 30 * 24 * time.Hour},
		{FreqNever, 0},
	}

	for _, tt := range tests {
		t.Run(string(tt.freq), func(t *testing.T) {
			if got := tt.freq.Duration(); got != tt.expected {
				t.Errorf("Frequency(%s).Duration() = %v, want %v", tt.freq, got, tt.expected)
			}
		})
	}
}

func TestFrequencyShouldSync(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		freq     Frequency
		lastSync time.Time
		expected bool
	}{
		{"never returns false", FreqNever, now, false},
		{"always returns true", FreqAlways, now, true},
		{"zero time returns true", FreqDaily, time.Time{}, true},
		{"hourly fresh", FreqHourly, now.Add(-30 * time.Minute), false},
		{"hourly stale", FreqHourly, now.Add(-2 * time.Hour), true},
		{"daily fresh", FreqDaily, now.Add(-12 * time.Hour), false},
		{"daily stale", FreqDaily, now.Add(-25 * time.Hour), true},
		{"weekly fresh", FreqWeekly, now.Add(-3 * 24 * time.Hour), false},
		{"weekly stale", FreqWeekly, now.Add(-8 * 24 * time.Hour), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.freq.ShouldSync(tt.lastSync); got != tt.expected {
				t.Errorf("ShouldSync() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.DefaultFreq != FreqDaily {
		t.Errorf("DefaultFreq = %s, want %s", cfg.DefaultFreq, FreqDaily)
	}

	if len(cfg.Feeds) == 0 {
		t.Error("Expected at least one feed configured")
	}

	// Check Semgrep rules feed is configured
	semgrepFeed := cfg.GetFeedConfig(FeedSemgrepRules)
	if semgrepFeed == nil {
		t.Error("Expected Semgrep rules feed to be configured")
	} else {
		if !semgrepFeed.Enabled {
			t.Error("Expected Semgrep rules feed to be enabled")
		}
		if semgrepFeed.Frequency != FreqWeekly {
			t.Errorf("Semgrep feed frequency = %s, want %s", semgrepFeed.Frequency, FreqWeekly)
		}
	}

	// Check pre-approved URLs
	if len(cfg.PreApprovedURLs) == 0 {
		t.Error("Expected pre-approved URLs to be configured")
	}
}

func TestConfigIsURLPreApproved(t *testing.T) {
	cfg := DefaultConfig()

	tests := []struct {
		url      string
		expected bool
	}{
		{"https://api.osv.dev/v1/query", true},
		{"https://semgrep.dev/c/p/default", true},
		{"https://evil.com/malware", false},
		{"https://api.osv.dev", true},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			if got := cfg.IsURLPreApproved(tt.url); got != tt.expected {
				t.Errorf("IsURLPreApproved(%s) = %v, want %v", tt.url, got, tt.expected)
			}
		})
	}
}

func TestConfigGetFeedConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Should find Semgrep feed
	if fc := cfg.GetFeedConfig(FeedSemgrepRules); fc == nil {
		t.Error("Expected to find Semgrep rules feed")
	}

	// Should not find unknown feed
	if fc := cfg.GetFeedConfig(FeedType("unknown")); fc != nil {
		t.Error("Expected nil for unknown feed type")
	}
}

func TestDefaultRuleConfig(t *testing.T) {
	cfg := DefaultRuleConfig()

	if !cfg.GeneratedRules.Enabled {
		t.Error("Expected generated rules to be enabled")
	}

	if cfg.GeneratedRules.Frequency != FreqAlways {
		t.Errorf("Generated rules frequency = %s, want %s", cfg.GeneratedRules.Frequency, FreqAlways)
	}

	if !cfg.CommunityRules.Enabled {
		t.Error("Expected community rules to be enabled")
	}

	if cfg.CommunityRules.Frequency != FreqWeekly {
		t.Errorf("Community rules frequency = %s, want %s", cfg.CommunityRules.Frequency, FreqWeekly)
	}
}
