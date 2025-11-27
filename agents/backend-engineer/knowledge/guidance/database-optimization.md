# Database Optimization Guide

## Query Optimization

### Understanding Query Plans

Always analyze slow queries with `EXPLAIN ANALYZE`:

```sql
EXPLAIN ANALYZE SELECT * FROM orders WHERE user_id = 123;

-- Output interpretation:
-- Seq Scan = Full table scan (usually bad for large tables)
-- Index Scan = Using index (good)
-- Index Only Scan = Covered by index (best)
-- Bitmap Heap Scan = Combining multiple indexes
```

### Indexing Strategies

#### When to Create Indexes

1. **Primary keys** - Automatic
2. **Foreign keys** - Always index
3. **WHERE clause columns** - Frequently filtered
4. **JOIN columns** - Used in joins
5. **ORDER BY columns** - If sorting large results

#### Index Types

```sql
-- B-tree (default, most common)
CREATE INDEX idx_users_email ON users(email);

-- Partial index (filtered rows only)
CREATE INDEX idx_active_users ON users(email) WHERE active = true;

-- Covering index (includes all needed columns)
CREATE INDEX idx_users_search ON users(email) INCLUDE (name, created_at);

-- GIN for arrays/JSONB
CREATE INDEX idx_user_tags ON users USING GIN(tags);

-- Expression index
CREATE INDEX idx_users_lower_email ON users(LOWER(email));
```

#### Index Anti-Patterns

```sql
-- Don't: Function on indexed column (won't use index)
SELECT * FROM users WHERE LOWER(email) = 'john@example.com';

-- Do: Use expression index or store lowercase
CREATE INDEX idx_users_lower_email ON users(LOWER(email));

-- Don't: Leading wildcard (can't use index)
SELECT * FROM users WHERE email LIKE '%@gmail.com';

-- Do: Full-text search or reverse index
CREATE INDEX idx_users_email_reverse ON users(REVERSE(email));
```

### Query Patterns

#### N+1 Problem

```sql
-- Bad: N+1 queries
SELECT * FROM orders WHERE user_id = 1;  -- 1 query
SELECT * FROM users WHERE id = 1;        -- N queries in loop

-- Good: Single JOIN
SELECT o.*, u.*
FROM orders o
JOIN users u ON u.id = o.user_id
WHERE o.user_id = 1;

-- Good: Batch query
SELECT * FROM users WHERE id IN (1, 2, 3, ...);
```

#### Pagination

```sql
-- Offset pagination (slow on large offsets)
SELECT * FROM orders ORDER BY created_at DESC LIMIT 20 OFFSET 10000;

-- Keyset pagination (fast, consistent)
SELECT * FROM orders
WHERE created_at < '2024-01-01'
ORDER BY created_at DESC
LIMIT 20;
```

#### SELECT Only Needed Columns

```sql
-- Bad: Fetching unnecessary data
SELECT * FROM users;

-- Good: Explicit columns
SELECT id, name, email FROM users;
```

### Common Performance Issues

| Issue | Symptom | Solution |
|-------|---------|----------|
| Missing index | Seq Scan on large table | Add appropriate index |
| Wrong index | Index Scan but slow | Review index columns |
| Outdated stats | Bad row estimates | Run ANALYZE |
| Lock contention | Queries waiting | Optimize transactions |
| Connection exhaustion | Connection errors | Increase pool size |

## Schema Design

### Normalization vs Denormalization

**Normalize when:**
- Data integrity is critical
- Write-heavy workload
- Data changes frequently

**Denormalize when:**
- Read performance is critical
- Data rarely changes
- Joins are expensive

### Choosing Data Types

```sql
-- IDs
id UUID PRIMARY KEY DEFAULT gen_random_uuid()  -- Distributed
id BIGSERIAL PRIMARY KEY                       -- Single database

-- Timestamps
created_at TIMESTAMPTZ DEFAULT NOW()           -- Always use timezone
updated_at TIMESTAMPTZ

-- Money
price NUMERIC(10, 2)                           -- Exact precision
-- NOT: FLOAT or DOUBLE (precision loss)

-- JSON
metadata JSONB                                 -- Indexable, efficient
-- NOT: JSON (slower, no indexing)

-- Enums
status VARCHAR(20) CHECK (status IN ('pending', 'active', 'closed'))
-- Or: CREATE TYPE status AS ENUM (...)
```

### Table Partitioning

For very large tables (millions of rows):

```sql
-- Range partitioning (time-based)
CREATE TABLE orders (
  id UUID,
  created_at TIMESTAMPTZ,
  amount NUMERIC
) PARTITION BY RANGE (created_at);

CREATE TABLE orders_2024_q1 PARTITION OF orders
  FOR VALUES FROM ('2024-01-01') TO ('2024-04-01');

-- Hash partitioning (even distribution)
CREATE TABLE users PARTITION BY HASH (id);

CREATE TABLE users_0 PARTITION OF users
  FOR VALUES WITH (MODULUS 4, REMAINDER 0);
```

## Connection Management

### Connection Pooling

```javascript
// PgBouncer or application-level pooling
const pool = new Pool({
  host: 'localhost',
  database: 'myapp',
  max: 20,                      // Max connections
  idleTimeoutMillis: 30000,     // Close idle after 30s
  connectionTimeoutMillis: 2000 // Timeout waiting for connection
});

// Always release connections
const client = await pool.connect();
try {
  const result = await client.query('SELECT ...');
  return result.rows;
} finally {
  client.release(); // Always release!
}
```

### Connection Pool Sizing

Rule of thumb: `connections = (core_count * 2) + effective_spindle_count`

For cloud databases:
- Small: 10-20 connections
- Medium: 20-50 connections
- Large: 50-100 connections

## Caching

### Query Result Caching

```javascript
// Cache expensive queries
const cacheKey = `user:${userId}`;
let user = await cache.get(cacheKey);

if (!user) {
  user = await db.query('SELECT * FROM users WHERE id = $1', [userId]);
  await cache.set(cacheKey, user, { ttl: 300 }); // 5 min TTL
}
```

### Cache Invalidation

```javascript
// Invalidate on write
async function updateUser(userId, data) {
  await db.query('UPDATE users SET ...', [data, userId]);
  await cache.del(`user:${userId}`);
}

// Time-based expiration
await cache.set(key, value, { ttl: 60 }); // 1 minute

// Version-based invalidation
const version = await cache.get('users:version');
const cacheKey = `users:${userId}:${version}`;
```

## Transactions

### Isolation Levels

```sql
-- Read Committed (default in PostgreSQL)
-- Sees committed data at statement start
BEGIN;
SELECT * FROM accounts WHERE id = 1; -- May see different value
SELECT * FROM accounts WHERE id = 1; -- if committed between
COMMIT;

-- Repeatable Read
-- Sees snapshot at transaction start
BEGIN ISOLATION LEVEL REPEATABLE READ;
SELECT * FROM accounts WHERE id = 1; -- Same value
SELECT * FROM accounts WHERE id = 1; -- throughout transaction
COMMIT;

-- Serializable
-- Full isolation, may fail and retry
BEGIN ISOLATION LEVEL SERIALIZABLE;
-- ... critical operations ...
COMMIT;
```

### Deadlock Prevention

1. **Consistent ordering** - Always lock tables/rows in same order
2. **Short transactions** - Minimize lock duration
3. **Row-level locks** - Avoid table locks when possible

```sql
-- Use SELECT FOR UPDATE for row-level locks
SELECT * FROM accounts WHERE id = 1 FOR UPDATE;
UPDATE accounts SET balance = balance - 100 WHERE id = 1;
```

## Monitoring

### Key Metrics

- **Query latency** - p50, p95, p99 response times
- **Connections** - Active, idle, waiting
- **Cache hit ratio** - Should be > 99%
- **Lock waits** - Contention indicator
- **Replication lag** - For read replicas

### Slow Query Logging

```sql
-- PostgreSQL
ALTER SYSTEM SET log_min_duration_statement = '100ms';
SELECT pg_reload_conf();
```

### Table Statistics

```sql
-- Update statistics
ANALYZE users;

-- View table statistics
SELECT relname, n_live_tup, n_dead_tup, last_vacuum, last_analyze
FROM pg_stat_user_tables;

-- Find unused indexes
SELECT indexrelname, idx_scan
FROM pg_stat_user_indexes
WHERE idx_scan = 0;
```
