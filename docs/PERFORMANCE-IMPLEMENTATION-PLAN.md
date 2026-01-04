# Zero Web Performance Implementation Plan

> **Status**: In Progress
> **Created**: 2026-01-04
> **Last Updated**: 2026-01-04

## Executive Summary

This document outlines the implementation plan to improve Zero's web UI performance while maintaining CLI independence. The solution uses **SQLite as a local query index** with **in-memory caching** on the backend and **SWR request deduplication** on the frontend.

## Problem Statement

### Current Performance Issues

**Backend (API Layer)**:
| Issue | Impact | Location |
|-------|--------|----------|
| Every API request reads full JSON files from disk | 500-2000ms response times | `pkg/api/handlers/analysis.go` |
| Directory walking to count files/sizes | 200-500ms per project | `pkg/api/handlers/projects.go` |
| `/api/analysis/stats` reads ALL repos' JSON files | 2-5 seconds | `GetAggregateStats()` |
| `CacheTTLHours` config exists but is never used | No caching benefit | `pkg/core/config/` |

**Frontend (Web UI)**:
| Issue | Impact | Location |
|-------|--------|----------|
| Dashboard calls `api.analysis.stats()` 3 times | Duplicate network requests | `web/src/app/page.tsx` |
| Vulnerabilities page: N projects = N parallel API calls | Request explosion | `web/src/app/vulnerabilities/page.tsx` |
| Polling every 2-5 seconds | Constant server load | `web/src/hooks/useApi.ts` |
| No SWR/React Query | No request deduplication | Custom `useFetch` hook |

### Performance Targets

| Metric | Current | Target |
|--------|---------|--------|
| `GET /api/analysis/stats` | 500-2000ms | <50ms |
| `GET /api/repos` | 200-500ms | <30ms |
| Dashboard full load | 3-5 seconds | <500ms |
| API calls per page load | 5-50 | 1-3 |

## Architecture Overview

```
CLI (independent)              Web (optimized)
     |                              |
     v                              v
JSON Files  <-- sync -->  SQLite Index  -->  In-Memory Cache  -->  API
(source of truth)         (fast queries)     (30-60s TTL)
                                                    |
                                                    v
                                            SWR (dedup + cache)
                                                    |
                                                    v
                                                React UI
```

### Key Principles

1. **CLI Independence**: JSON files remain source of truth, CLI works fully offline
2. **SQLite as Index**: Fast indexed queries without changing file storage
3. **Layered Caching**: Cache at API level (in-memory) and client level (SWR)
4. **Incremental Adoption**: Each phase provides immediate benefits

---

## Phase 1: SQLite Storage Layer

### Overview

Create a SQLite database as a fast query index. The database mirrors summary data from JSON files but enables indexed queries.

### Completed Work

**Files Created**:
- `pkg/storage/interface.go` - Store interface definition
- `pkg/storage/sqlite/store.go` - SQLite implementation
- `pkg/storage/sqlite/migrations.go` - Schema migrations

### Database Schema

```sql
-- Projects (replaces directory walking)
CREATE TABLE projects (
    id TEXT PRIMARY KEY,           -- "owner/repo"
    owner TEXT NOT NULL,
    name TEXT NOT NULL,
    repo_path TEXT NOT NULL,
    analysis_path TEXT NOT NULL,
    file_count INTEGER DEFAULT 0,
    disk_size INTEGER DEFAULT 0,
    last_scan TIMESTAMP,
    freshness_level TEXT,          -- fresh, stale, very-stale, expired
    freshness_age INTEGER DEFAULT 0,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

-- Aggregated findings (replaces reading all JSON files)
CREATE TABLE findings_summary (
    project_id TEXT PRIMARY KEY,
    vulns_critical INTEGER DEFAULT 0,
    vulns_high INTEGER DEFAULT 0,
    vulns_medium INTEGER DEFAULT 0,
    vulns_low INTEGER DEFAULT 0,
    vulns_total INTEGER DEFAULT 0,
    secrets_total INTEGER DEFAULT 0,
    packages_total INTEGER DEFAULT 0,
    technologies_total INTEGER DEFAULT 0,
    updated_at TIMESTAMP
);

-- Vulnerabilities (for cross-project queries)
CREATE TABLE vulnerabilities (
    id INTEGER PRIMARY KEY,
    project_id TEXT NOT NULL,
    vuln_id TEXT NOT NULL,         -- CVE/GHSA ID
    package TEXT NOT NULL,
    version TEXT,
    severity TEXT NOT NULL,
    title TEXT,
    description TEXT,
    fix_version TEXT,
    source TEXT,                   -- package, code
    scanner TEXT
);

-- Secrets (for aggregated view)
CREATE TABLE secrets (
    id INTEGER PRIMARY KEY,
    project_id TEXT NOT NULL,
    file TEXT NOT NULL,
    line INTEGER,
    type TEXT NOT NULL,
    severity TEXT,
    description TEXT,
    redacted_match TEXT
);
```

### Storage Interface

```go
type Store interface {
    // Projects
    ListProjects(ctx context.Context, opts ListOptions) ([]*Project, error)
    GetProject(ctx context.Context, id string) (*Project, error)
    UpsertProject(ctx context.Context, project *Project) error

    // Aggregations (fast indexed queries)
    GetAggregateStats(ctx context.Context) (*AggregateStats, error)

    // Vulnerabilities (cross-project)
    GetVulnerabilities(ctx context.Context, opts VulnOptions) ([]*Vulnerability, int, error)

    // Sync from JSON
    SyncProjectFromJSON(ctx context.Context, projectID, analysisDir string) error
}
```

### Remaining Work

1. **Integrate with hydrate workflow** (`pkg/workflow/hydrate/hydrate.go`)
   - Call `store.UpsertProject()` after clone
   - Call `store.SyncProjectFromJSON()` after scan completion
   - Update freshness level on freshness checks

2. **Add CLI command for initial sync**
   ```bash
   zero db sync  # Populate SQLite from existing JSON files
   ```

---

## Phase 2: API Layer Optimization

### Overview

Update API handlers to use SQLite store instead of file walking/reading. Add in-memory cache for frequently accessed data.

### In-Memory Cache Design

```go
// pkg/cache/cache.go
type Cache struct {
    store  sync.Map
    ttl    time.Duration
}

type entry struct {
    value     interface{}
    expiresAt time.Time
}

func (c *Cache) Get(key string) (interface{}, bool)
func (c *Cache) Set(key string, value interface{})
func (c *Cache) Delete(key string)
func (c *Cache) DeletePrefix(prefix string)  // For invalidation
```

### Cache Strategy

| Endpoint | TTL | Invalidation Trigger |
|----------|-----|---------------------|
| `GET /api/analysis/stats` | 30s | Scan completion |
| `GET /api/repos` | 10s | Scan completion |
| `GET /api/repos/{id}/analysis/*` | 60s | Scan completion for that project |
| `GET /api/vulnerabilities` | 30s | Scan completion |

### Handler Changes

**Before** (`pkg/api/handlers/analysis.go`):
```go
func (h *AnalysisHandler) GetAggregateStats(w http.ResponseWriter, r *http.Request) {
    // Walk all directories, read all JSON files - SLOW
    stats := walkAndAggregate(h.zeroHome)
    json.NewEncoder(w).Encode(stats)
}
```

**After**:
```go
func (h *AnalysisHandler) GetAggregateStats(w http.ResponseWriter, r *http.Request) {
    // Check cache first
    if cached, ok := h.cache.Get("aggregate_stats"); ok {
        json.NewEncoder(w).Encode(cached)
        return
    }

    // Single indexed query - FAST
    stats, err := h.store.GetAggregateStats(r.Context())
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    h.cache.Set("aggregate_stats", stats)
    json.NewEncoder(w).Encode(stats)
}
```

### New Endpoints

```
GET /api/vulnerabilities
    ?severity=critical,high    # Filter by severity
    ?project=owner/repo        # Filter by project
    &limit=100                 # Pagination
    &offset=0

GET /api/secrets
    ?severity=critical,high
    ?type=api_key
    &limit=100
    &offset=0
```

### Files to Modify

| File | Changes |
|------|---------|
| `pkg/api/server.go` | Initialize store and cache |
| `pkg/api/handlers/projects.go` | Use `store.ListProjects()` |
| `pkg/api/handlers/analysis.go` | Use `store.GetAggregateStats()`, add cache |
| `pkg/api/handlers/vulnerabilities.go` | New file for paginated endpoint |

---

## Phase 3: Web UI Optimization

### Overview

Add SWR for request deduplication and client-side caching. Reduce polling frequency and add visibility-aware polling.

### SWR Configuration

```typescript
// web/src/lib/swr-config.ts
import { SWRConfig } from 'swr';

export const swrConfig = {
  fetcher: (url: string) => fetch(url).then(r => r.json()),
  dedupingInterval: 5000,     // Dedupe requests within 5s
  revalidateOnFocus: true,    // Refresh when tab focused
  revalidateOnReconnect: true,
  errorRetryCount: 3,
};
```

### Hook Refactoring

**Before** (`web/src/hooks/useApi.ts`):
```typescript
// Custom hook - no deduplication
export function useFetch<T>(fetcher: () => Promise<T>) {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetcher().then(setData).finally(() => setLoading(false));
  }, []);

  return { data, loading };
}
```

**After** (`web/src/hooks/useApi.ts`):
```typescript
import useSWR from 'swr';

// SWR hook - automatic deduplication and caching
export function useAggregateStats() {
  return useSWR('/api/analysis/stats', fetcher, {
    refreshInterval: 30000,  // Refresh every 30s
  });
}

export function useProjects() {
  return useSWR('/api/repos', fetcher, {
    refreshInterval: 10000,
  });
}

export function useVulnerabilities(opts?: VulnOptions) {
  const params = new URLSearchParams(opts);
  return useSWR(`/api/vulnerabilities?${params}`, fetcher);
}
```

### Dashboard Optimization

**Before** (`web/src/app/page.tsx`):
```tsx
// 3 duplicate calls to same endpoint
function Dashboard() {
  const stats1 = useFetch(() => api.analysis.stats());
  const stats2 = useFetch(() => api.analysis.stats());
  const stats3 = useFetch(() => api.analysis.stats());

  return (
    <>
      <StatsCards stats={stats1.data} />
      <SeverityChart stats={stats2.data} />
      <TopIssues stats={stats3.data} />
    </>
  );
}
```

**After**:
```tsx
// Single SWR hook - automatically deduplicated
function Dashboard() {
  const { data: stats } = useAggregateStats();

  return (
    <>
      <StatsCards stats={stats} />
      <SeverityChart stats={stats} />
      <TopIssues stats={stats} />
    </>
  );
}
```

### Polling Optimization

**Before**:
```typescript
// Aggressive polling
useInterval(() => refetch(), 2000);  // Every 2s
useInterval(() => fetchQueue(), 5000);  // Every 5s
```

**After**:
```typescript
// Visibility-aware polling
import { usePageVisibility } from '@/hooks/usePageVisibility';

function usePolling(callback: () => void, interval: number) {
  const isVisible = usePageVisibility();

  useEffect(() => {
    if (!isVisible) return;  // Don't poll when tab hidden

    const id = setInterval(callback, interval);
    return () => clearInterval(id);
  }, [isVisible, callback, interval]);
}

// Increased intervals
usePolling(refetch, 10000);   // 10s instead of 2s
usePolling(fetchQueue, 30000); // 30s instead of 5s
```

### Files to Modify

| File | Changes |
|------|---------|
| `web/package.json` | Add `swr` dependency |
| `web/src/lib/swr-config.ts` | New: SWR configuration |
| `web/src/hooks/useApi.ts` | Replace custom hooks with SWR |
| `web/src/hooks/usePageVisibility.ts` | New: Visibility detection |
| `web/src/app/page.tsx` | Use shared `useAggregateStats()` |
| `web/src/app/vulnerabilities/page.tsx` | Use paginated endpoint |
| `web/src/app/secrets/page.tsx` | Use paginated endpoint |

---

## Phase 4: Optional Supabase Sync

### Overview

Add optional cloud sync for multi-user scenarios. SQLite remains the local source, Supabase provides cloud backup and realtime.

### Architecture

```
Local SQLite  ──[Background Sync]──>  Supabase
     |                                    |
     v                                    v
 Local API                         Supabase Realtime
                                         |
                                         v
                                   Other Users
```

### Implementation

```go
// pkg/storage/supabase/sync.go
type SupabaseSync struct {
    client    *supabase.Client
    sqliteDB  *sqlite.Store
    interval  time.Duration
}

func (s *SupabaseSync) Start(ctx context.Context) {
    ticker := time.NewTicker(s.interval)
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            s.syncProjects(ctx)
            s.syncFindings(ctx)
        }
    }
}
```

### Configuration

```bash
# Enable Supabase sync (optional)
export SUPABASE_URL="https://xxx.supabase.co"
export SUPABASE_KEY="your-anon-key"
```

---

## Implementation Order

| Phase | Priority | Effort | Impact |
|-------|----------|--------|--------|
| 1. SQLite Storage | P0 | Medium | High - Foundation for fast queries |
| 2. API Caching | P0 | Low | High - Immediate response time improvement |
| 3. Web SWR | P1 | Medium | High - Eliminate duplicate requests |
| 4. Supabase Sync | P2 | Medium | Medium - Cloud features when needed |

### Current Status

- [x] Phase 1: SQLite storage layer (interface, types, store, migrations)
- [ ] Phase 1: Integrate SQLite sync into hydrate workflow
- [ ] Phase 2: Add in-memory cache layer
- [ ] Phase 2: Update API handlers
- [ ] Phase 3: Add SWR to web UI
- [ ] Phase 3: Optimize polling
- [ ] Phase 4: Optional Supabase sync

---

## Testing Strategy

### Unit Tests

```go
// pkg/storage/sqlite/store_test.go
func TestStore_UpsertProject(t *testing.T) {
    store := setupTestDB(t)
    defer store.Close()

    project := &storage.Project{
        ID: "test/repo",
        Owner: "test",
        Name: "repo",
    }

    err := store.UpsertProject(context.Background(), project)
    require.NoError(t, err)

    got, err := store.GetProject(context.Background(), "test/repo")
    require.NoError(t, err)
    assert.Equal(t, project.ID, got.ID)
}
```

### Integration Tests

```bash
# Sync existing projects and verify counts
zero db sync
zero db verify  # Compare SQLite counts vs JSON file counts
```

### Performance Benchmarks

```go
func BenchmarkGetAggregateStats_SQLite(b *testing.B) {
    store := setupBenchDB(b)
    ctx := context.Background()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        store.GetAggregateStats(ctx)
    }
}
```

---

## Rollback Plan

If issues arise, rollback is straightforward:

1. **API handlers**: Revert to reading JSON files directly
2. **Web UI**: Revert to custom `useFetch` hooks
3. **SQLite**: Delete `.zero/zero.db` - no data loss since JSON is source of truth

---

## Monitoring

### Metrics to Track

| Metric | Target | Alert Threshold |
|--------|--------|-----------------|
| API response time (p95) | <100ms | >500ms |
| SQLite query time (p95) | <10ms | >50ms |
| Cache hit rate | >80% | <50% |
| Web UI initial load | <1s | >3s |

### Logging

```go
// Add timing logs to handlers
start := time.Now()
stats, err := h.store.GetAggregateStats(ctx)
log.Printf("GetAggregateStats took %v", time.Since(start))
```

---

## References

- [SQLite Write-Ahead Logging](https://www.sqlite.org/wal.html)
- [SWR Documentation](https://swr.vercel.app/)
- [Supabase Realtime](https://supabase.com/docs/guides/realtime)
