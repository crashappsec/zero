package feeds

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const statusFilename = "feed-status.json"

// Manager handles feed synchronization
type Manager struct {
	zeroHome   string
	config     Config
	status     map[FeedType]*FeedStatus
	statusMu   sync.RWMutex
	httpClient *http.Client
}

// NewManager creates a new feed manager
func NewManager(zeroHome string) *Manager {
	return &Manager{
		zeroHome: zeroHome,
		config:   DefaultConfig(),
		status:   make(map[FeedType]*FeedStatus),
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// NewManagerWithConfig creates a manager with custom config
func NewManagerWithConfig(zeroHome string, config Config) *Manager {
	m := NewManager(zeroHome)
	m.config = config
	return m
}

// statusPath returns the path to the feed status file
func (m *Manager) statusPath() string {
	return filepath.Join(m.zeroHome, statusFilename)
}

// cacheDir returns the cache directory for feeds
func (m *Manager) cacheDir() string {
	if m.config.CacheDir != "" {
		return m.config.CacheDir
	}
	return filepath.Join(m.zeroHome, "cache", "feeds")
}

// LoadStatus loads the feed status from disk
func (m *Manager) LoadStatus() error {
	m.statusMu.Lock()
	defer m.statusMu.Unlock()

	data, err := os.ReadFile(m.statusPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("reading feed status: %w", err)
	}

	var statuses []FeedStatus
	if err := json.Unmarshal(data, &statuses); err != nil {
		return fmt.Errorf("parsing feed status: %w", err)
	}

	for _, s := range statuses {
		m.status[s.Type] = &FeedStatus{
			Type:        s.Type,
			Name:        s.Name,
			LastSync:    s.LastSync,
			LastSuccess: s.LastSuccess,
			LastError:   s.LastError,
			Version:     s.Version,
			ItemCount:   s.ItemCount,
		}
	}

	return nil
}

// SaveStatus saves the feed status to disk
func (m *Manager) SaveStatus() error {
	m.statusMu.RLock()
	defer m.statusMu.RUnlock()

	var statuses []FeedStatus
	for _, s := range m.status {
		statuses = append(statuses, *s)
	}

	data, err := json.MarshalIndent(statuses, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling feed status: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(m.statusPath()), 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	if err := os.WriteFile(m.statusPath(), data, 0644); err != nil {
		return fmt.Errorf("writing feed status: %w", err)
	}

	return nil
}

// GetStatus returns the status of a feed
func (m *Manager) GetStatus(feedType FeedType) *FeedStatus {
	m.statusMu.RLock()
	defer m.statusMu.RUnlock()
	return m.status[feedType]
}

// ShouldSync returns true if the feed should be synced
func (m *Manager) ShouldSync(feedType FeedType) bool {
	cfg := m.config.GetFeedConfig(feedType)
	if cfg == nil || !cfg.Enabled {
		return false
	}

	status := m.GetStatus(feedType)
	if status == nil {
		return true // Never synced
	}

	return cfg.Frequency.ShouldSync(status.LastSync)
}

// SyncAll syncs all enabled feeds
func (m *Manager) SyncAll(ctx context.Context) ([]SyncResult, error) {
	if err := m.LoadStatus(); err != nil {
		return nil, err
	}

	var results []SyncResult

	for _, cfg := range m.config.Feeds {
		if !cfg.Enabled {
			continue
		}

		result := m.SyncFeed(ctx, cfg.Type)
		results = append(results, result)
	}

	if err := m.SaveStatus(); err != nil {
		return results, err
	}

	return results, nil
}

// SyncIfNeeded syncs feeds that need syncing based on frequency
func (m *Manager) SyncIfNeeded(ctx context.Context) ([]SyncResult, error) {
	if err := m.LoadStatus(); err != nil {
		return nil, err
	}

	var results []SyncResult

	for _, cfg := range m.config.Feeds {
		if !cfg.Enabled {
			continue
		}

		if !m.ShouldSync(cfg.Type) {
			results = append(results, SyncResult{
				Feed:    cfg.Type,
				Skipped: true,
				Reason:  "not due for sync",
			})
			continue
		}

		result := m.SyncFeed(ctx, cfg.Type)
		results = append(results, result)
	}

	if err := m.SaveStatus(); err != nil {
		return results, err
	}

	return results, nil
}

// SyncFeed syncs a specific feed
func (m *Manager) SyncFeed(ctx context.Context, feedType FeedType) SyncResult {
	start := time.Now()

	cfg := m.config.GetFeedConfig(feedType)
	if cfg == nil {
		return SyncResult{
			Feed:    feedType,
			Success: false,
			Error:   "feed not configured",
		}
	}

	result := SyncResult{
		Feed: feedType,
	}

	var err error
	var itemCount int

	switch feedType {
	case FeedSemgrepRules:
		itemCount, err = m.syncSemgrepRules(ctx, cfg)
	default:
		err = fmt.Errorf("unknown feed type: %s", feedType)
	}

	result.Duration = time.Since(start)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		m.updateStatus(feedType, cfg.Name, false, err.Error(), 0)
	} else {
		result.Success = true
		result.ItemCount = itemCount
		m.updateStatus(feedType, cfg.Name, true, "", itemCount)
	}

	return result
}

func (m *Manager) updateStatus(feedType FeedType, name string, success bool, errMsg string, itemCount int) {
	m.statusMu.Lock()
	defer m.statusMu.Unlock()

	status, ok := m.status[feedType]
	if !ok {
		status = &FeedStatus{
			Type: feedType,
			Name: name,
		}
		m.status[feedType] = status
	}

	status.LastSync = time.Now()
	if success {
		status.LastSuccess = time.Now()
		status.LastError = ""
		status.ItemCount = itemCount
	} else {
		status.LastError = errMsg
	}
}

func (m *Manager) syncSemgrepRules(ctx context.Context, cfg *FeedConfig) (int, error) {
	// Create cache directory
	cacheDir := filepath.Join(m.cacheDir(), "semgrep")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return 0, fmt.Errorf("creating cache dir: %w", err)
	}

	// Download rules from Semgrep registry
	req, err := http.NewRequestWithContext(ctx, "GET", cfg.URL, nil)
	if err != nil {
		return 0, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("downloading rules: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("semgrep registry returned status %d: %s", resp.StatusCode, string(body))
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("reading response: %w", err)
	}

	// Parse to count rules
	var rulesData struct {
		Rules []json.RawMessage `json:"rules"`
	}
	if err := json.Unmarshal(body, &rulesData); err != nil {
		// If it's not JSON, it might be YAML - just save it
		rulesData.Rules = nil
	}

	// Save to cache file
	cachePath := filepath.Join(cacheDir, "rules.json")
	if err := os.WriteFile(cachePath, body, 0644); err != nil {
		return 0, fmt.Errorf("writing cache file: %w", err)
	}

	// Save metadata
	meta := map[string]interface{}{
		"url":         cfg.URL,
		"downloaded":  time.Now().Format(time.RFC3339),
		"rule_count":  len(rulesData.Rules),
		"size_bytes":  len(body),
		"etag":        resp.Header.Get("ETag"),
		"last_modified": resp.Header.Get("Last-Modified"),
	}
	metaData, _ := json.MarshalIndent(meta, "", "  ")
	metaPath := filepath.Join(cacheDir, "metadata.json")
	os.WriteFile(metaPath, metaData, 0644)

	return len(rulesData.Rules), nil
}

// GetCachedRulesPath returns the path to cached Semgrep rules
func (m *Manager) GetCachedRulesPath() string {
	return filepath.Join(m.cacheDir(), "semgrep", "rules.json")
}

// HasCachedRules returns true if cached Semgrep rules exist
func (m *Manager) HasCachedRules() bool {
	_, err := os.Stat(m.GetCachedRulesPath())
	return err == nil
}

// GetAllStatus returns status of all feeds
func (m *Manager) GetAllStatus() []FeedStatus {
	m.statusMu.RLock()
	defer m.statusMu.RUnlock()

	var statuses []FeedStatus
	for _, s := range m.status {
		statuses = append(statuses, *s)
	}
	return statuses
}

// GetConfig returns the current configuration
func (m *Manager) GetConfig() Config {
	return m.config
}

// SetConfig updates the configuration
func (m *Manager) SetConfig(config Config) {
	m.config = config
}
