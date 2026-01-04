// Package sqlite provides a SQLite implementation of the storage.Store interface.
package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite" // Pure Go SQLite driver

	"github.com/crashappsec/zero/pkg/storage"
)

// Store implements storage.Store using SQLite.
type Store struct {
	db     *sql.DB
	dbPath string
}

// New creates a new SQLite store at the given path.
func New(dbPath string) (*Store, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating db directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(1) // SQLite only supports one writer
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	store := &Store{db: db, dbPath: dbPath}

	// Run migrations
	if err := store.Migrate(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return store, nil
}

// Ping checks if the database is accessible.
func (s *Store) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// ListProjects returns all projects with optional filtering.
func (s *Store) ListProjects(ctx context.Context, opts storage.ListOptions) ([]*storage.Project, error) {
	query := `SELECT id, owner, name, repo_path, analysis_path, file_count, disk_size,
		last_scan, freshness_level, freshness_age, created_at, updated_at
		FROM projects`

	var args []interface{}
	var conditions []string

	if opts.Owner != "" {
		conditions = append(conditions, "owner = ?")
		args = append(args, opts.Owner)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Sort
	sortBy := "last_scan"
	if opts.SortBy != "" {
		sortBy = opts.SortBy
	}
	sortOrder := "DESC"
	if !opts.SortDesc {
		sortOrder = "ASC"
	}
	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	// Pagination
	if opts.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, opts.Limit)
	}
	if opts.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, opts.Offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying projects: %w", err)
	}
	defer rows.Close()

	var projects []*storage.Project
	for rows.Next() {
		p := &storage.Project{}
		var lastScan, createdAt, updatedAt sql.NullTime
		err := rows.Scan(&p.ID, &p.Owner, &p.Name, &p.RepoPath, &p.AnalysisPath,
			&p.FileCount, &p.DiskSize, &lastScan, &p.FreshnessLevel, &p.FreshnessAge,
			&createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("scanning project row: %w", err)
		}
		if lastScan.Valid {
			p.LastScan = lastScan.Time
		}
		if createdAt.Valid {
			p.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			p.UpdatedAt = updatedAt.Time
		}
		projects = append(projects, p)
	}

	return projects, rows.Err()
}

// GetProject returns a single project by ID.
func (s *Store) GetProject(ctx context.Context, id string) (*storage.Project, error) {
	query := `SELECT id, owner, name, repo_path, analysis_path, file_count, disk_size,
		last_scan, freshness_level, freshness_age, created_at, updated_at
		FROM projects WHERE id = ?`

	p := &storage.Project{}
	var lastScan, createdAt, updatedAt sql.NullTime
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Owner, &p.Name, &p.RepoPath, &p.AnalysisPath,
		&p.FileCount, &p.DiskSize, &lastScan, &p.FreshnessLevel, &p.FreshnessAge,
		&createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying project: %w", err)
	}
	if lastScan.Valid {
		p.LastScan = lastScan.Time
	}
	if createdAt.Valid {
		p.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		p.UpdatedAt = updatedAt.Time
	}
	return p, nil
}

// UpsertProject creates or updates a project.
func (s *Store) UpsertProject(ctx context.Context, p *storage.Project) error {
	query := `INSERT INTO projects (id, owner, name, repo_path, analysis_path, file_count, disk_size,
		last_scan, freshness_level, freshness_age, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			owner = excluded.owner,
			name = excluded.name,
			repo_path = excluded.repo_path,
			analysis_path = excluded.analysis_path,
			file_count = excluded.file_count,
			disk_size = excluded.disk_size,
			last_scan = excluded.last_scan,
			freshness_level = excluded.freshness_level,
			freshness_age = excluded.freshness_age,
			updated_at = excluded.updated_at`

	now := time.Now()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	p.UpdatedAt = now

	_, err := s.db.ExecContext(ctx, query,
		p.ID, p.Owner, p.Name, p.RepoPath, p.AnalysisPath, p.FileCount, p.DiskSize,
		p.LastScan, p.FreshnessLevel, p.FreshnessAge, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("upserting project: %w", err)
	}
	return nil
}

// DeleteProject removes a project and all related data.
func (s *Store) DeleteProject(ctx context.Context, id string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete in order due to foreign keys
	tables := []string{"secrets", "vulnerabilities", "scanner_results", "scans", "findings_summary", "projects"}
	for _, table := range tables {
		var query string
		if table == "projects" {
			query = fmt.Sprintf("DELETE FROM %s WHERE id = ?", table)
		} else {
			query = fmt.Sprintf("DELETE FROM %s WHERE project_id = ?", table)
		}
		if _, err := tx.ExecContext(ctx, query, id); err != nil {
			return fmt.Errorf("deleting from %s: %w", table, err)
		}
	}

	return tx.Commit()
}

// CreateScan creates a new scan record.
func (s *Store) CreateScan(ctx context.Context, scan *storage.Scan) error {
	query := `INSERT INTO scans (id, project_id, profile, status, commit_sha, started_at, finished_at, duration_seconds, error)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.ExecContext(ctx, query,
		scan.ID, scan.ProjectID, scan.Profile, scan.Status, scan.CommitSHA,
		scan.StartedAt, nullTime(scan.FinishedAt), scan.DurationSeconds, nullString(scan.Error))
	if err != nil {
		return fmt.Errorf("creating scan: %w", err)
	}
	return nil
}

// UpdateScan updates an existing scan record.
func (s *Store) UpdateScan(ctx context.Context, scan *storage.Scan) error {
	query := `UPDATE scans SET status = ?, commit_sha = ?, finished_at = ?, duration_seconds = ?, error = ?
		WHERE id = ?`

	_, err := s.db.ExecContext(ctx, query,
		scan.Status, scan.CommitSHA, nullTime(scan.FinishedAt), scan.DurationSeconds,
		nullString(scan.Error), scan.ID)
	if err != nil {
		return fmt.Errorf("updating scan: %w", err)
	}
	return nil
}

// GetScan returns a scan by ID.
func (s *Store) GetScan(ctx context.Context, id string) (*storage.Scan, error) {
	query := `SELECT id, project_id, profile, status, commit_sha, started_at, finished_at, duration_seconds, error
		FROM scans WHERE id = ?`

	scan := &storage.Scan{}
	var finishedAt sql.NullTime
	var scanError sql.NullString
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&scan.ID, &scan.ProjectID, &scan.Profile, &scan.Status, &scan.CommitSHA,
		&scan.StartedAt, &finishedAt, &scan.DurationSeconds, &scanError)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying scan: %w", err)
	}
	if finishedAt.Valid {
		scan.FinishedAt = finishedAt.Time
	}
	if scanError.Valid {
		scan.Error = scanError.String
	}
	return scan, nil
}

// ListScans returns scans for a project.
func (s *Store) ListScans(ctx context.Context, projectID string, opts storage.ListOptions) ([]*storage.Scan, error) {
	query := `SELECT id, project_id, profile, status, commit_sha, started_at, finished_at, duration_seconds, error
		FROM scans WHERE project_id = ? ORDER BY started_at DESC`

	args := []interface{}{projectID}
	if opts.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, opts.Limit)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying scans: %w", err)
	}
	defer rows.Close()

	var scans []*storage.Scan
	for rows.Next() {
		scan := &storage.Scan{}
		var finishedAt sql.NullTime
		var scanError sql.NullString
		err := rows.Scan(&scan.ID, &scan.ProjectID, &scan.Profile, &scan.Status, &scan.CommitSHA,
			&scan.StartedAt, &finishedAt, &scan.DurationSeconds, &scanError)
		if err != nil {
			return nil, fmt.Errorf("scanning scan row: %w", err)
		}
		if finishedAt.Valid {
			scan.FinishedAt = finishedAt.Time
		}
		if scanError.Valid {
			scan.Error = scanError.String
		}
		scans = append(scans, scan)
	}

	return scans, rows.Err()
}

// GetLatestScan returns the most recent scan for a project.
func (s *Store) GetLatestScan(ctx context.Context, projectID string) (*storage.Scan, error) {
	scans, err := s.ListScans(ctx, projectID, storage.ListOptions{Limit: 1})
	if err != nil {
		return nil, err
	}
	if len(scans) == 0 {
		return nil, nil
	}
	return scans[0], nil
}

// UpsertFindingsSummary creates or updates a findings summary.
func (s *Store) UpsertFindingsSummary(ctx context.Context, summary *storage.FindingsSummary) error {
	query := `INSERT INTO findings_summary (project_id, vulns_critical, vulns_high, vulns_medium, vulns_low,
		vulns_total, secrets_total, packages_total, technologies_total, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(project_id) DO UPDATE SET
			vulns_critical = excluded.vulns_critical,
			vulns_high = excluded.vulns_high,
			vulns_medium = excluded.vulns_medium,
			vulns_low = excluded.vulns_low,
			vulns_total = excluded.vulns_total,
			secrets_total = excluded.secrets_total,
			packages_total = excluded.packages_total,
			technologies_total = excluded.technologies_total,
			updated_at = excluded.updated_at`

	if summary.UpdatedAt.IsZero() {
		summary.UpdatedAt = time.Now()
	}

	_, err := s.db.ExecContext(ctx, query,
		summary.ProjectID, summary.VulnsCritical, summary.VulnsHigh, summary.VulnsMedium, summary.VulnsLow,
		summary.VulnsTotal, summary.SecretsTotal, summary.PackagesTotal, summary.TechnologiesTotal, summary.UpdatedAt)
	if err != nil {
		return fmt.Errorf("upserting findings summary: %w", err)
	}
	return nil
}

// GetFindingsSummary returns the findings summary for a project.
func (s *Store) GetFindingsSummary(ctx context.Context, projectID string) (*storage.FindingsSummary, error) {
	query := `SELECT project_id, vulns_critical, vulns_high, vulns_medium, vulns_low,
		vulns_total, secrets_total, packages_total, technologies_total, updated_at
		FROM findings_summary WHERE project_id = ?`

	summary := &storage.FindingsSummary{}
	err := s.db.QueryRowContext(ctx, query, projectID).Scan(
		&summary.ProjectID, &summary.VulnsCritical, &summary.VulnsHigh, &summary.VulnsMedium, &summary.VulnsLow,
		&summary.VulnsTotal, &summary.SecretsTotal, &summary.PackagesTotal, &summary.TechnologiesTotal, &summary.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying findings summary: %w", err)
	}
	return summary, nil
}

// GetAggregateStats returns global statistics across all projects.
func (s *Store) GetAggregateStats(ctx context.Context) (*storage.AggregateStats, error) {
	stats := &storage.AggregateStats{
		VulnsBySeverity: make(map[string]int),
		FreshnessCounts: make(map[string]int),
	}

	// Count projects by freshness level
	freshnessQuery := `SELECT freshness_level, COUNT(*) FROM projects GROUP BY freshness_level`
	rows, err := s.db.QueryContext(ctx, freshnessQuery)
	if err != nil {
		return nil, fmt.Errorf("querying freshness counts: %w", err)
	}
	for rows.Next() {
		var level string
		var count int
		if err := rows.Scan(&level, &count); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scanning freshness row: %w", err)
		}
		stats.FreshnessCounts[level] = count
		stats.TotalProjects += count
	}
	rows.Close()

	// Aggregate findings
	aggQuery := `SELECT
		COALESCE(SUM(vulns_critical), 0),
		COALESCE(SUM(vulns_high), 0),
		COALESCE(SUM(vulns_medium), 0),
		COALESCE(SUM(vulns_low), 0),
		COALESCE(SUM(vulns_total), 0),
		COALESCE(SUM(secrets_total), 0),
		COALESCE(SUM(packages_total), 0),
		COALESCE(SUM(technologies_total), 0)
		FROM findings_summary`

	var critical, high, medium, low int
	err = s.db.QueryRowContext(ctx, aggQuery).Scan(
		&critical, &high, &medium, &low, &stats.TotalVulns,
		&stats.TotalSecrets, &stats.TotalPackages, &stats.TotalTechnologies)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("querying aggregate stats: %w", err)
	}

	stats.VulnsBySeverity["critical"] = critical
	stats.VulnsBySeverity["high"] = high
	stats.VulnsBySeverity["medium"] = medium
	stats.VulnsBySeverity["low"] = low

	return stats, nil
}

// UpsertVulnerabilities replaces all vulnerabilities for a project.
func (s *Store) UpsertVulnerabilities(ctx context.Context, projectID string, vulns []*storage.Vulnerability) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing vulnerabilities for this project
	if _, err := tx.ExecContext(ctx, "DELETE FROM vulnerabilities WHERE project_id = ?", projectID); err != nil {
		return fmt.Errorf("deleting old vulnerabilities: %w", err)
	}

	// Insert new vulnerabilities
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO vulnerabilities
		(project_id, vuln_id, package, version, severity, title, description, fix_version, source, scanner)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	for _, v := range vulns {
		_, err := stmt.ExecContext(ctx, projectID, v.VulnID, v.Package, v.Version,
			v.Severity, v.Title, v.Description, v.FixVersion, v.Source, v.Scanner)
		if err != nil {
			return fmt.Errorf("inserting vulnerability: %w", err)
		}
	}

	return tx.Commit()
}

// GetVulnerabilities returns vulnerabilities with filtering.
func (s *Store) GetVulnerabilities(ctx context.Context, opts storage.VulnOptions) ([]*storage.Vulnerability, int, error) {
	// Build query
	baseQuery := `FROM vulnerabilities WHERE 1=1`
	var args []interface{}

	if opts.ProjectID != "" {
		baseQuery += " AND project_id = ?"
		args = append(args, opts.ProjectID)
	}
	if len(opts.Severities) > 0 {
		placeholders := make([]string, len(opts.Severities))
		for i, sev := range opts.Severities {
			placeholders[i] = "?"
			args = append(args, sev)
		}
		baseQuery += fmt.Sprintf(" AND severity IN (%s)", strings.Join(placeholders, ","))
	}
	if opts.Package != "" {
		baseQuery += " AND package LIKE ?"
		args = append(args, "%"+opts.Package+"%")
	}

	// Get total count
	var total int
	countQuery := "SELECT COUNT(*) " + baseQuery
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting vulnerabilities: %w", err)
	}

	// Get results
	selectQuery := `SELECT id, project_id, vuln_id, package, version, severity, title, description, fix_version, source, scanner ` + baseQuery
	selectQuery += " ORDER BY CASE severity WHEN 'critical' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 WHEN 'low' THEN 4 ELSE 5 END"

	if opts.Limit > 0 {
		selectQuery += " LIMIT ?"
		args = append(args, opts.Limit)
	}
	if opts.Offset > 0 {
		selectQuery += " OFFSET ?"
		args = append(args, opts.Offset)
	}

	rows, err := s.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("querying vulnerabilities: %w", err)
	}
	defer rows.Close()

	var vulns []*storage.Vulnerability
	for rows.Next() {
		v := &storage.Vulnerability{}
		err := rows.Scan(&v.ID, &v.ProjectID, &v.VulnID, &v.Package, &v.Version,
			&v.Severity, &v.Title, &v.Description, &v.FixVersion, &v.Source, &v.Scanner)
		if err != nil {
			return nil, 0, fmt.Errorf("scanning vulnerability row: %w", err)
		}
		vulns = append(vulns, v)
	}

	return vulns, total, rows.Err()
}

// DeleteVulnerabilities removes all vulnerabilities for a project.
func (s *Store) DeleteVulnerabilities(ctx context.Context, projectID string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM vulnerabilities WHERE project_id = ?", projectID)
	return err
}

// UpsertSecrets replaces all secrets for a project.
func (s *Store) UpsertSecrets(ctx context.Context, projectID string, secrets []*storage.Secret) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing secrets for this project
	if _, err := tx.ExecContext(ctx, "DELETE FROM secrets WHERE project_id = ?", projectID); err != nil {
		return fmt.Errorf("deleting old secrets: %w", err)
	}

	// Insert new secrets
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO secrets
		(project_id, file, line, type, severity, description, redacted_match)
		VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	for _, s := range secrets {
		_, err := stmt.ExecContext(ctx, projectID, s.File, s.Line, s.Type, s.Severity, s.Description, s.RedactedMatch)
		if err != nil {
			return fmt.Errorf("inserting secret: %w", err)
		}
	}

	return tx.Commit()
}

// GetSecrets returns secrets with filtering.
func (s *Store) GetSecrets(ctx context.Context, opts storage.SecretOptions) ([]*storage.Secret, int, error) {
	// Build query
	baseQuery := `FROM secrets WHERE 1=1`
	var args []interface{}

	if opts.ProjectID != "" {
		baseQuery += " AND project_id = ?"
		args = append(args, opts.ProjectID)
	}
	if len(opts.Severities) > 0 {
		placeholders := make([]string, len(opts.Severities))
		for i, sev := range opts.Severities {
			placeholders[i] = "?"
			args = append(args, sev)
		}
		baseQuery += fmt.Sprintf(" AND severity IN (%s)", strings.Join(placeholders, ","))
	}
	if opts.Type != "" {
		baseQuery += " AND type = ?"
		args = append(args, opts.Type)
	}

	// Get total count
	var total int
	countQuery := "SELECT COUNT(*) " + baseQuery
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting secrets: %w", err)
	}

	// Get results
	selectQuery := `SELECT id, project_id, file, line, type, severity, description, redacted_match ` + baseQuery
	selectQuery += " ORDER BY CASE severity WHEN 'critical' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 WHEN 'low' THEN 4 ELSE 5 END"

	if opts.Limit > 0 {
		selectQuery += " LIMIT ?"
		args = append(args, opts.Limit)
	}
	if opts.Offset > 0 {
		selectQuery += " OFFSET ?"
		args = append(args, opts.Offset)
	}

	rows, err := s.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("querying secrets: %w", err)
	}
	defer rows.Close()

	var secrets []*storage.Secret
	for rows.Next() {
		sec := &storage.Secret{}
		err := rows.Scan(&sec.ID, &sec.ProjectID, &sec.File, &sec.Line, &sec.Type,
			&sec.Severity, &sec.Description, &sec.RedactedMatch)
		if err != nil {
			return nil, 0, fmt.Errorf("scanning secret row: %w", err)
		}
		secrets = append(secrets, sec)
	}

	return secrets, total, rows.Err()
}

// DeleteSecrets removes all secrets for a project.
func (s *Store) DeleteSecrets(ctx context.Context, projectID string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM secrets WHERE project_id = ?", projectID)
	return err
}

// SyncProjectFromJSON populates the database from JSON analysis files.
func (s *Store) SyncProjectFromJSON(ctx context.Context, projectID string, analysisDir string) error {
	// Extract summary stats from JSON files
	summary := &storage.FindingsSummary{ProjectID: projectID}
	var vulns []*storage.Vulnerability
	var secrets []*storage.Secret

	// Read code-packages.json for package vulns
	packagesPath := filepath.Join(analysisDir, "code-packages.json")
	if data, err := os.ReadFile(packagesPath); err == nil {
		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err == nil {
			vulns, summary.VulnsCritical, summary.VulnsHigh, summary.VulnsMedium, summary.VulnsLow =
				extractVulnsFromPackages(projectID, result)
			summary.VulnsTotal = summary.VulnsCritical + summary.VulnsHigh + summary.VulnsMedium + summary.VulnsLow

			// Extract package count from summary
			if summaryData, ok := result["summary"].(map[string]interface{}); ok {
				if sbom, ok := summaryData["sbom"].(map[string]interface{}); ok {
					if count, ok := sbom["total_components"].(float64); ok {
						summary.PackagesTotal = int(count)
					}
				}
			}
		}
	}

	// Read code-security.json for secrets
	securityPath := filepath.Join(analysisDir, "code-security.json")
	if data, err := os.ReadFile(securityPath); err == nil {
		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err == nil {
			secrets = extractSecrets(projectID, result)
			summary.SecretsTotal = len(secrets)

			// Add code vulns to count
			codeVulns, critical, high, medium, low := extractCodeVulns(projectID, result)
			vulns = append(vulns, codeVulns...)
			summary.VulnsCritical += critical
			summary.VulnsHigh += high
			summary.VulnsMedium += medium
			summary.VulnsLow += low
			summary.VulnsTotal += critical + high + medium + low
		}
	}

	// Read technology-identification.json for tech count
	techPath := filepath.Join(analysisDir, "technology-identification.json")
	if data, err := os.ReadFile(techPath); err == nil {
		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err == nil {
			if summaryData, ok := result["summary"].(map[string]interface{}); ok {
				if count, ok := summaryData["total_technologies"].(float64); ok {
					summary.TechnologiesTotal = int(count)
				}
			}
		}
	}

	// Update database
	if err := s.UpsertFindingsSummary(ctx, summary); err != nil {
		return fmt.Errorf("upserting findings summary: %w", err)
	}
	if err := s.UpsertVulnerabilities(ctx, projectID, vulns); err != nil {
		return fmt.Errorf("upserting vulnerabilities: %w", err)
	}
	if err := s.UpsertSecrets(ctx, projectID, secrets); err != nil {
		return fmt.Errorf("upserting secrets: %w", err)
	}

	return nil
}

// Helper functions

func nullTime(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: t, Valid: true}
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func extractVulnsFromPackages(projectID string, data map[string]interface{}) ([]*storage.Vulnerability, int, int, int, int) {
	var vulns []*storage.Vulnerability
	var critical, high, medium, low int

	findings, ok := data["findings"].(map[string]interface{})
	if !ok {
		return vulns, 0, 0, 0, 0
	}

	vulnFindings, ok := findings["vulns"].([]interface{})
	if !ok {
		return vulns, 0, 0, 0, 0
	}

	for _, vf := range vulnFindings {
		v, ok := vf.(map[string]interface{})
		if !ok {
			continue
		}

		vuln := &storage.Vulnerability{
			ProjectID: projectID,
			Source:    "package",
			Scanner:   "code-packages",
		}

		if id, ok := v["id"].(string); ok {
			vuln.VulnID = id
		}
		if pkg, ok := v["package"].(string); ok {
			vuln.Package = pkg
		}
		if ver, ok := v["version"].(string); ok {
			vuln.Version = ver
		}
		if sev, ok := v["severity"].(string); ok {
			vuln.Severity = strings.ToLower(sev)
			switch vuln.Severity {
			case "critical":
				critical++
			case "high":
				high++
			case "medium":
				medium++
			case "low":
				low++
			}
		}
		if title, ok := v["title"].(string); ok {
			vuln.Title = title
		}
		if desc, ok := v["description"].(string); ok {
			vuln.Description = desc
		}
		if fix, ok := v["fix_version"].(string); ok {
			vuln.FixVersion = fix
		}

		vulns = append(vulns, vuln)
	}

	return vulns, critical, high, medium, low
}

func extractCodeVulns(projectID string, data map[string]interface{}) ([]*storage.Vulnerability, int, int, int, int) {
	var vulns []*storage.Vulnerability
	var critical, high, medium, low int

	findings, ok := data["findings"].(map[string]interface{})
	if !ok {
		return vulns, 0, 0, 0, 0
	}

	vulnFindings, ok := findings["vulns"].([]interface{})
	if !ok {
		return vulns, 0, 0, 0, 0
	}

	for _, vf := range vulnFindings {
		v, ok := vf.(map[string]interface{})
		if !ok {
			continue
		}

		vuln := &storage.Vulnerability{
			ProjectID: projectID,
			Source:    "code",
			Scanner:   "code-security",
		}

		if id, ok := v["rule_id"].(string); ok {
			vuln.VulnID = id
		}
		if sev, ok := v["severity"].(string); ok {
			vuln.Severity = strings.ToLower(sev)
			switch vuln.Severity {
			case "critical":
				critical++
			case "high":
				high++
			case "medium":
				medium++
			case "low":
				low++
			}
		}
		if title, ok := v["message"].(string); ok {
			vuln.Title = title
		}
		if loc, ok := v["location"].(map[string]interface{}); ok {
			if file, ok := loc["file"].(string); ok {
				vuln.Package = file // Use file as "package" for code vulns
			}
		}

		vulns = append(vulns, vuln)
	}

	return vulns, critical, high, medium, low
}

func extractSecrets(projectID string, data map[string]interface{}) []*storage.Secret {
	var secrets []*storage.Secret

	findings, ok := data["findings"].(map[string]interface{})
	if !ok {
		return secrets
	}

	secretFindings, ok := findings["secrets"].([]interface{})
	if !ok {
		return secrets
	}

	for _, sf := range secretFindings {
		s, ok := sf.(map[string]interface{})
		if !ok {
			continue
		}

		secret := &storage.Secret{
			ProjectID: projectID,
		}

		if file, ok := s["file"].(string); ok {
			secret.File = file
		}
		if line, ok := s["line"].(float64); ok {
			secret.Line = int(line)
		}
		if t, ok := s["type"].(string); ok {
			secret.Type = t
		}
		if sev, ok := s["severity"].(string); ok {
			secret.Severity = strings.ToLower(sev)
		}
		if desc, ok := s["description"].(string); ok {
			secret.Description = desc
		}
		if match, ok := s["redacted_match"].(string); ok {
			secret.RedactedMatch = match
		}

		secrets = append(secrets, secret)
	}

	return secrets
}
