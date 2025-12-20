// Package feeds provides external feed synchronization for security data
package feeds

import (
	"time"
)

// Frequency defines how often to sync a feed
type Frequency string

const (
	FreqNever   Frequency = "never"   // Never sync automatically
	FreqAlways  Frequency = "always"  // Sync on every hydrate
	FreqHourly  Frequency = "hourly"  // Sync if older than 1 hour
	FreqDaily   Frequency = "daily"   // Sync if older than 1 day
	FreqWeekly  Frequency = "weekly"  // Sync if older than 1 week
	FreqMonthly Frequency = "monthly" // Sync if older than 1 month
)

// Duration returns the duration for this frequency
func (f Frequency) Duration() time.Duration {
	switch f {
	case FreqAlways:
		return 0
	case FreqHourly:
		return time.Hour
	case FreqDaily:
		return 24 * time.Hour
	case FreqWeekly:
		return 7 * 24 * time.Hour
	case FreqMonthly:
		return 30 * 24 * time.Hour
	default:
		return 0
	}
}

// ShouldSync returns true if the feed should be synced based on last sync time
func (f Frequency) ShouldSync(lastSync time.Time) bool {
	if f == FreqNever {
		return false
	}
	if f == FreqAlways {
		return true
	}
	if lastSync.IsZero() {
		return true
	}
	return time.Since(lastSync) > f.Duration()
}

// FeedType identifies the type of feed
type FeedType string

const (
	FeedSemgrepRules FeedType = "semgrep-rules"
	FeedRAGPatterns  FeedType = "rag-patterns"
	// Note: Vulnerability data is queried LIVE via OSV.dev API during scans
	// Pre-approved URL: https://api.osv.dev/v1/query
)

// FeedConfig configures a single feed
type FeedConfig struct {
	Type      FeedType  `json:"type"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Frequency Frequency `json:"frequency"`
	Enabled   bool      `json:"enabled"`
}

// FeedStatus tracks the status of a feed
type FeedStatus struct {
	Type        FeedType  `json:"type"`
	Name        string    `json:"name"`
	LastSync    time.Time `json:"last_sync"`
	LastSuccess time.Time `json:"last_success"`
	LastError   string    `json:"last_error,omitempty"`
	Version     string    `json:"version,omitempty"`
	ItemCount   int       `json:"item_count,omitempty"`
}

// SyncResult holds the result of a feed sync operation
type SyncResult struct {
	Feed      FeedType      `json:"feed"`
	Success   bool          `json:"success"`
	Duration  time.Duration `json:"duration"`
	ItemCount int           `json:"item_count,omitempty"`
	Error     string        `json:"error,omitempty"`
	Skipped   bool          `json:"skipped,omitempty"`
	Reason    string        `json:"reason,omitempty"`
}

// Config holds all feed configurations
type Config struct {
	Feeds          []FeedConfig `json:"feeds"`
	DefaultFreq    Frequency    `json:"default_frequency"`
	CacheDir       string       `json:"cache_dir"`
	PreApprovedURLs []string    `json:"pre_approved_urls"`
}

// DefaultConfig returns default feed configuration
func DefaultConfig() Config {
	return Config{
		DefaultFreq: FreqDaily,
		Feeds: []FeedConfig{
			{
				Type:      FeedSemgrepRules,
				Name:      "semgrep-community",
				URL:       "https://semgrep.dev/c/p/default",
				Frequency: FreqWeekly,
				Enabled:   true,
			},
		},
		// Pre-approved URLs for security data
		// Note: OSV.dev is queried LIVE during scans, not cached here
		PreApprovedURLs: []string{
			"https://api.osv.dev",
			"https://semgrep.dev",
		},
	}
}

// RuleConfig configures rule generation
type RuleConfig struct {
	GeneratedRules RuleSourceConfig `json:"generated_rules"`
	CommunityRules RuleSourceConfig `json:"community_rules"`
}

// RuleSourceConfig configures a rule source
type RuleSourceConfig struct {
	Enabled   bool      `json:"enabled"`
	Frequency Frequency `json:"frequency"`
	OutputDir string    `json:"output_dir"`
}

// DefaultRuleConfig returns default rule configuration
func DefaultRuleConfig() RuleConfig {
	return RuleConfig{
		GeneratedRules: RuleSourceConfig{
			Enabled:   true,
			Frequency: FreqAlways, // Always check if RAG changed (fast, local)
			OutputDir: "rules/generated",
		},
		CommunityRules: RuleSourceConfig{
			Enabled:   true,
			Frequency: FreqWeekly, // Community rules don't change often
			OutputDir: "rules/community",
		},
	}
}

// IsURLPreApproved checks if a URL is in the pre-approved list
func (c *Config) IsURLPreApproved(url string) bool {
	for _, approved := range c.PreApprovedURLs {
		if len(url) >= len(approved) && url[:len(approved)] == approved {
			return true
		}
	}
	return false
}

// GetFeedConfig returns the config for a specific feed type
func (c *Config) GetFeedConfig(feedType FeedType) *FeedConfig {
	for i := range c.Feeds {
		if c.Feeds[i].Type == feedType {
			return &c.Feeds[i]
		}
	}
	return nil
}
