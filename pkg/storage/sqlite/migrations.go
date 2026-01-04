package sqlite

import (
	"context"
	"fmt"
)

// Migrate runs all database migrations.
func (s *Store) Migrate(ctx context.Context) error {
	// Create migrations table if not exists
	_, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS migrations (
			id INTEGER PRIMARY KEY,
			version INTEGER NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("creating migrations table: %w", err)
	}

	// Get current version
	var currentVersion int
	err = s.db.QueryRowContext(ctx, "SELECT COALESCE(MAX(version), 0) FROM migrations").Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("getting current migration version: %w", err)
	}

	// Run pending migrations
	for version, migration := range migrations {
		if version <= currentVersion {
			continue
		}

		if _, err := s.db.ExecContext(ctx, migration); err != nil {
			return fmt.Errorf("running migration %d: %w", version, err)
		}

		if _, err := s.db.ExecContext(ctx, "INSERT INTO migrations (version) VALUES (?)", version); err != nil {
			return fmt.Errorf("recording migration %d: %w", version, err)
		}
	}

	return nil
}

// migrations is an ordered map of version -> SQL.
var migrations = map[int]string{
	1: migration001,
	2: migration002,
}

const migration001 = `
-- Projects table (replaces directory walking)
CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    owner TEXT NOT NULL,
    name TEXT NOT NULL,
    repo_path TEXT NOT NULL,
    analysis_path TEXT NOT NULL,
    file_count INTEGER DEFAULT 0,
    disk_size INTEGER DEFAULT 0,
    last_scan TIMESTAMP,
    freshness_level TEXT DEFAULT 'unknown',
    freshness_age INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_projects_owner ON projects(owner);
CREATE INDEX IF NOT EXISTS idx_projects_last_scan ON projects(last_scan DESC);
CREATE INDEX IF NOT EXISTS idx_projects_freshness ON projects(freshness_level);

-- Scans table
CREATE TABLE IF NOT EXISTS scans (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    profile TEXT NOT NULL,
    status TEXT NOT NULL,
    commit_sha TEXT,
    started_at TIMESTAMP NOT NULL,
    finished_at TIMESTAMP,
    duration_seconds INTEGER DEFAULT 0,
    error TEXT,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_scans_project ON scans(project_id);
CREATE INDEX IF NOT EXISTS idx_scans_status ON scans(status);
CREATE INDEX IF NOT EXISTS idx_scans_started ON scans(started_at DESC);

-- Scanner results per scan
CREATE TABLE IF NOT EXISTS scanner_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    scan_id TEXT NOT NULL,
    scanner TEXT NOT NULL,
    status TEXT NOT NULL,
    duration_seconds INTEGER DEFAULT 0,
    finding_count INTEGER DEFAULT 0,
    output_file TEXT,
    error TEXT,
    FOREIGN KEY (scan_id) REFERENCES scans(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_scanner_results_scan ON scanner_results(scan_id);

-- Aggregated findings summary (denormalized for fast queries)
CREATE TABLE IF NOT EXISTS findings_summary (
    project_id TEXT PRIMARY KEY,
    vulns_critical INTEGER DEFAULT 0,
    vulns_high INTEGER DEFAULT 0,
    vulns_medium INTEGER DEFAULT 0,
    vulns_low INTEGER DEFAULT 0,
    vulns_total INTEGER DEFAULT 0,
    secrets_total INTEGER DEFAULT 0,
    packages_total INTEGER DEFAULT 0,
    technologies_total INTEGER DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);
`

const migration002 = `
-- Individual vulnerabilities (for cross-project queries)
CREATE TABLE IF NOT EXISTS vulnerabilities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id TEXT NOT NULL,
    vuln_id TEXT NOT NULL,
    package TEXT NOT NULL,
    version TEXT,
    severity TEXT NOT NULL,
    title TEXT,
    description TEXT,
    fix_version TEXT,
    source TEXT,
    scanner TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_vulns_project ON vulnerabilities(project_id);
CREATE INDEX IF NOT EXISTS idx_vulns_severity ON vulnerabilities(severity);
CREATE INDEX IF NOT EXISTS idx_vulns_package ON vulnerabilities(package);

-- Secrets (for aggregated view)
CREATE TABLE IF NOT EXISTS secrets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id TEXT NOT NULL,
    file TEXT NOT NULL,
    line INTEGER,
    type TEXT NOT NULL,
    severity TEXT,
    description TEXT,
    redacted_match TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_secrets_project ON secrets(project_id);
CREATE INDEX IF NOT EXISTS idx_secrets_severity ON secrets(severity);
`
