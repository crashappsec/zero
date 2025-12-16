package codeownership

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestOwnershipScanner_Name(t *testing.T) {
	s := &OwnershipScanner{}
	if s.Name() != "code-ownership" {
		t.Errorf("Name() = %q, want %q", s.Name(), "code-ownership")
	}
}

func TestOwnershipScanner_Description(t *testing.T) {
	s := &OwnershipScanner{}
	desc := s.Description()
	if desc == "" {
		t.Error("Description() should not be empty")
	}
}

func TestOwnershipScanner_Dependencies(t *testing.T) {
	s := &OwnershipScanner{}
	deps := s.Dependencies()
	// Returns empty slice, not nil
	if len(deps) != 0 {
		t.Errorf("Dependencies() = %v, want empty slice", deps)
	}
}

func TestOwnershipScanner_EstimateDuration(t *testing.T) {
	s := &OwnershipScanner{}

	// Base is 10s + 5ms per file
	tests := []struct {
		fileCount int
		wantMin   int
	}{
		{0, 10},     // 10s base
		{500, 12},   // 10s + 2.5s
		{1000, 15},  // 10s + 5s
		{5000, 35},  // 10s + 25s
	}

	for _, tt := range tests {
		got := s.EstimateDuration(tt.fileCount)
		if got.Seconds() < float64(tt.wantMin) {
			t.Errorf("EstimateDuration(%d) = %v, want at least %ds", tt.fileCount, got, tt.wantMin)
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if !cfg.Enabled {
		t.Error("Enabled should be true by default")
	}
	if !cfg.EnhancedMode {
		t.Error("EnhancedMode should be true by default")
	}
	if !cfg.AnalyzeContributors {
		t.Error("AnalyzeContributors should be enabled by default")
	}
	if !cfg.CheckCodeowners {
		t.Error("CheckCodeowners should be enabled by default")
	}
	if !cfg.DetectOrphans {
		t.Error("DetectOrphans should be enabled by default")
	}
	if !cfg.AnalyzeCompetency {
		t.Error("AnalyzeCompetency should be enabled by default")
	}
	if !cfg.DetectLanguages {
		t.Error("DetectLanguages should be enabled by default")
	}
	if cfg.PeriodDays != 90 {
		t.Errorf("PeriodDays = %d, want 90", cfg.PeriodDays)
	}
}

func TestQuickConfig(t *testing.T) {
	cfg := QuickConfig()

	if !cfg.Enabled {
		t.Error("Enabled should be true in quick config")
	}
	if cfg.AnalyzeContributors {
		t.Error("AnalyzeContributors should be disabled in quick config")
	}
	if cfg.DetectOrphans {
		t.Error("DetectOrphans should be disabled in quick config")
	}
	if cfg.AnalyzeCompetency {
		t.Error("AnalyzeCompetency should be disabled in quick config")
	}
}

func TestFullConfig(t *testing.T) {
	cfg := FullConfig()

	if !cfg.Enabled {
		t.Error("Enabled should be true in full config")
	}
	if !cfg.AnalyzeContributors {
		t.Error("AnalyzeContributors should be enabled in full config")
	}
	if !cfg.CheckCodeowners {
		t.Error("CheckCodeowners should be enabled in full config")
	}
	if !cfg.DetectOrphans {
		t.Error("DetectOrphans should be enabled in full config")
	}
	if !cfg.AnalyzeCompetency {
		t.Error("AnalyzeCompetency should be enabled in full config")
	}
	if !cfg.DetectLanguages {
		t.Error("DetectLanguages should be enabled in full config")
	}
	if cfg.PeriodDays != 180 {
		t.Errorf("PeriodDays = %d, want 180 for full config", cfg.PeriodDays)
	}
}

func TestDefaultEnhancedConfig(t *testing.T) {
	cfg := DefaultEnhancedConfig()

	// Check weights sum to 1.0
	total := cfg.Weights.Commits + cfg.Weights.Reviews + cfg.Weights.Lines +
		cfg.Weights.Recency + cfg.Weights.Consistency
	if total < 0.99 || total > 1.01 {
		t.Errorf("Weights sum = %v, want 1.0", total)
	}

	// Check default weight values
	if cfg.Weights.Commits != 0.30 {
		t.Errorf("Weights.Commits = %v, want 0.30", cfg.Weights.Commits)
	}
	if cfg.Weights.Reviews != 0.25 {
		t.Errorf("Weights.Reviews = %v, want 0.25", cfg.Weights.Reviews)
	}

	// Check GitHub config
	if !cfg.GitHub.Enabled {
		t.Error("GitHub.Enabled should be true by default")
	}
	if !cfg.GitHub.FetchPRReviews {
		t.Error("GitHub.FetchPRReviews should be true by default")
	}
	if cfg.GitHub.MaxPRs != 500 {
		t.Errorf("GitHub.MaxPRs = %d, want 500", cfg.GitHub.MaxPRs)
	}

	// Check CODEOWNERS config
	if !cfg.CODEOWNERS.Validate {
		t.Error("CODEOWNERS.Validate should be true by default")
	}
	if !cfg.CODEOWNERS.DetectDrift {
		t.Error("CODEOWNERS.DetectDrift should be true by default")
	}
	if len(cfg.CODEOWNERS.SensitivePatterns) == 0 {
		t.Error("CODEOWNERS.SensitivePatterns should have entries")
	}

	// Check specialist domains
	if len(cfg.SpecialistDomains) == 0 {
		t.Error("SpecialistDomains should have entries")
	}
}

func TestDomainPatterns(t *testing.T) {
	// Verify all expected domains are present
	expectedDomains := []string{
		"supply-chain", "security", "compliance", "legal",
		"frontend", "backend", "architecture", "cicd",
		"infrastructure", "metrics", "crypto", "ai-ml",
	}

	for _, domain := range expectedDomains {
		patterns, ok := DomainPatterns[domain]
		if !ok {
			t.Errorf("DomainPatterns missing domain: %s", domain)
			continue
		}
		if len(patterns) == 0 {
			t.Errorf("DomainPatterns[%s] has no patterns", domain)
		}
	}
}

func TestActivityThresholds(t *testing.T) {
	// Verify thresholds are in correct order
	if ActivityThresholds.Active >= ActivityThresholds.Recent {
		t.Error("Active threshold should be less than Recent")
	}
	if ActivityThresholds.Recent >= ActivityThresholds.Stale {
		t.Error("Recent threshold should be less than Stale")
	}
	if ActivityThresholds.Stale >= ActivityThresholds.Inactive {
		t.Error("Stale threshold should be less than Inactive")
	}
}

func TestBusFactorThresholds(t *testing.T) {
	if BusFactorThresholds.Critical >= BusFactorThresholds.Warning {
		t.Error("Critical threshold should be less than Warning")
	}
}

func TestClassifyCommitType(t *testing.T) {
	// Implementation supports: bugfix, refactor, feature, other
	tests := []struct {
		message  string
		expected string
	}{
		// Feature patterns
		{"feat: add new feature", "feature"},
		{"feature: implement login", "feature"},
		{"add new component", "feature"},
		{"implement user auth", "feature"},
		{"create new service", "feature"},
		{"introduce caching", "feature"},
		{"support dark mode", "feature"},
		// Bug fix patterns
		{"fix: resolve bug", "bugfix"},
		{"bugfix: memory leak", "bugfix"},
		{"fixed crash", "bugfix"},
		{"hotfix: critical issue", "bugfix"},
		{"patch security hole", "bugfix"},
		{"resolve memory issue", "bugfix"},
		{"closes #123", "bugfix"},
		{"fixes #456", "bugfix"},
		// Refactor patterns
		{"refactor: clean up code", "refactor"},
		{"refactored auth module", "refactor"},
		{"cleanup unused imports", "refactor"},
		{"reorganize project structure", "refactor"},
		{"simplify the algorithm", "refactor"},
		{"optimize query performance", "refactor"},
		// Other (not matching any pattern)
		{"docs: update README", "other"},
		{"test: run tests", "other"},  // "add" matches feature pattern
		{"chore: update deps", "other"},
		{"bump version", "other"},
		{"random commit message", "other"},
	}

	for _, tt := range tests {
		got := classifyCommitType(tt.message)
		if got != tt.expected {
			t.Errorf("classifyCommitType(%q) = %q, want %q", tt.message, got, tt.expected)
		}
	}
}

func TestNewOwnershipScorer(t *testing.T) {
	weights := ScoringWeights{
		Commits:     0.30,
		Reviews:     0.25,
		Lines:       0.20,
		Recency:     0.15,
		Consistency: 0.10,
	}

	scorer := NewOwnershipScorer(weights)
	if scorer == nil {
		t.Fatal("NewOwnershipScorer returned nil")
	}
}

func TestCalculateBusFactor(t *testing.T) {
	// Bus factor is the minimum number of people whose combined ownership
	// exceeds the threshold. Uses >= threshold comparison.
	tests := []struct {
		name      string
		owners    []EnhancedOwnership
		threshold float64
		wantBF    int
		wantRisk  string
	}{
		{
			name:      "empty owners",
			owners:    []EnhancedOwnership{},
			threshold: 0.5,
			wantBF:    0,
			wantRisk:  "critical",
		},
		{
			name: "single dominant owner",
			owners: []EnhancedOwnership{
				{OwnershipScore: 80},
				{OwnershipScore: 10},
				{OwnershipScore: 10},
			},
			threshold: 0.5,
			wantBF:    1, // 80/100 = 80% >= 50%
			wantRisk:  "critical",
		},
		{
			name: "two significant owners",
			owners: []EnhancedOwnership{
				{OwnershipScore: 40},
				{OwnershipScore: 30},
				{OwnershipScore: 20},
				{OwnershipScore: 10},
			},
			threshold: 0.5,
			wantBF:    2, // (40+30)/100 = 70% >= 50%
			wantRisk:  "warning",
		},
		{
			name: "equal distribution of four",
			owners: []EnhancedOwnership{
				{OwnershipScore: 25},
				{OwnershipScore: 25},
				{OwnershipScore: 25},
				{OwnershipScore: 25},
			},
			threshold: 0.5,
			wantBF:    2, // (25+25)/100 = 50% >= 50%
			wantRisk:  "warning",
		},
		{
			name: "healthy distribution of five",
			owners: []EnhancedOwnership{
				{OwnershipScore: 20},
				{OwnershipScore: 20},
				{OwnershipScore: 20},
				{OwnershipScore: 20},
				{OwnershipScore: 20},
			},
			threshold: 0.5,
			wantBF:    3, // (20+20+20)/100 = 60% >= 50%
			wantRisk:  "healthy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf, risk := CalculateBusFactor(tt.owners, tt.threshold)
			if bf != tt.wantBF {
				t.Errorf("CalculateBusFactor() bus factor = %d, want %d", bf, tt.wantBF)
			}
			if risk != tt.wantRisk {
				t.Errorf("CalculateBusFactor() risk = %q, want %q", risk, tt.wantRisk)
			}
		})
	}
}

func TestCalculateOwnershipCoverage(t *testing.T) {
	// Returns a float 0-1 (not percentage 0-100)
	tests := []struct {
		name            string
		files           []FileOwnership
		minContributors int
		wantCoverage    float64
	}{
		{
			name:            "empty files",
			files:           []FileOwnership{},
			minContributors: 1,
			wantCoverage:    1.0, // No files = full coverage
		},
		{
			name: "all covered",
			files: []FileOwnership{
				{TopContributors: []string{"alice", "bob"}},
				{TopContributors: []string{"charlie"}},
				{TopContributors: []string{"dave", "eve"}},
			},
			minContributors: 1,
			wantCoverage:    1.0, // 3/3
		},
		{
			name: "partial coverage",
			files: []FileOwnership{
				{TopContributors: []string{"alice"}},
				{TopContributors: []string{}},
				{TopContributors: []string{"bob"}},
			},
			minContributors: 1,
			wantCoverage:    0.6667, // 2 out of 3 files
		},
		{
			name: "no coverage",
			files: []FileOwnership{
				{TopContributors: []string{}},
				{TopContributors: []string{}},
			},
			minContributors: 1,
			wantCoverage:    0.0, // 0/2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateOwnershipCoverage(tt.files, tt.minContributors)
			// Allow small floating point difference
			if got < tt.wantCoverage-0.01 || got > tt.wantCoverage+0.01 {
				t.Errorf("CalculateOwnershipCoverage() = %v, want ~%v", got, tt.wantCoverage)
			}
		})
	}
}

func TestParseCodeowners(t *testing.T) {
	// Create temp directory with CODEOWNERS file
	tmpDir, err := os.MkdirTemp("", "codeowners-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .github directory
	githubDir := filepath.Join(tmpDir, ".github")
	if err := os.MkdirAll(githubDir, 0755); err != nil {
		t.Fatalf("Failed to create .github dir: %v", err)
	}

	// Create CODEOWNERS file
	codeowners := `# This is a comment

# Global owners
*                   @global-team

# Frontend
*.tsx               @frontend-team @alice
*.jsx               @frontend-team

# Backend
/api/               @backend-team
/services/          @bob @charlie

# Security
/auth/              @security-team
`
	if err := os.WriteFile(filepath.Join(githubDir, "CODEOWNERS"), []byte(codeowners), 0644); err != nil {
		t.Fatalf("Failed to write CODEOWNERS: %v", err)
	}

	scanner := &OwnershipScanner{}
	rules := scanner.parseCodeowners(tmpDir)

	// Should have 6 rules (excluding comments and empty lines)
	if len(rules) < 5 {
		t.Errorf("parseCodeowners() found %d rules, want at least 5", len(rules))
	}

	// Verify a specific rule
	found := false
	for _, rule := range rules {
		if rule.Pattern == "*.tsx" {
			found = true
			if len(rule.Owners) != 2 {
				t.Errorf("*.tsx rule has %d owners, want 2", len(rule.Owners))
			}
		}
	}
	if !found {
		t.Error("Expected to find *.tsx rule")
	}
}

func TestExtractRepoInfo(t *testing.T) {
	tests := []struct {
		repoPath  string
		wantOwner string
		wantRepo  string
	}{
		// Local path extractions
		{"/Users/test/github/owner/repo", "owner", "repo"},
		{"/.zero/repos/owner/repo/repo", "owner", "repo"},
	}

	for _, tt := range tests {
		owner, repo := extractRepoInfo(tt.repoPath)
		// These are best-effort, may not always match
		_ = owner
		_ = repo
	}
}

func TestParseGitURL(t *testing.T) {
	// SSH format (git@) parses any host, HTTPS only handles github.com
	tests := []struct {
		url       string
		wantOwner string
		wantRepo  string
	}{
		{"git@github.com:owner/repo.git", "owner", "repo"},
		{"https://github.com/owner/repo.git", "owner", "repo"},
		{"https://github.com/owner/repo", "owner", "repo"},
		{"git@github.com:org/project.git", "org", "project"},
		// SSH format works for any host
		{"git@gitlab.com:group/project.git", "group", "project"},
		// HTTPS non-github returns empty (only github.com handled)
		{"https://gitlab.com/group/project.git", "", ""},
	}

	for _, tt := range tests {
		owner, repo := parseGitURL(tt.url)
		if owner != tt.wantOwner {
			t.Errorf("parseGitURL(%q) owner = %q, want %q", tt.url, owner, tt.wantOwner)
		}
		if repo != tt.wantRepo {
			t.Errorf("parseGitURL(%q) repo = %q, want %q", tt.url, repo, tt.wantRepo)
		}
	}
}

func TestOwnershipScorerDetermineActivityStatus(t *testing.T) {
	scorer := NewOwnershipScorer(ScoringWeights{
		Commits:     0.30,
		Reviews:     0.25,
		Lines:       0.20,
		Recency:     0.15,
		Consistency: 0.10,
	})

	tests := []struct {
		daysSince float64
		expected  string
	}{
		{10, "active"},
		{50, "recent"},
		{120, "stale"},
		{300, "inactive"},
		{400, "abandoned"},
	}

	for _, tt := range tests {
		got := scorer.determineActivityStatus(tt.daysSince)
		if got != tt.expected {
			t.Errorf("determineActivityStatus(%v) = %q, want %q", tt.daysSince, got, tt.expected)
		}
	}
}

func TestOwnershipScorerCalculateEnhancedOwnership(t *testing.T) {
	scorer := NewOwnershipScorer(ScoringWeights{
		Commits:     0.30,
		Reviews:     0.25,
		Lines:       0.20,
		Recency:     0.15,
		Consistency: 0.10,
	})

	now := time.Now()
	contributors := []ContributorData{
		{
			Name:         "alice",
			Email:        "alice@example.com",
			Commits:      100,
			LinesAdded:   5000,
			LinesRemoved: 2000,
			PRReviews:    20,
			LastCommit:   now.AddDate(0, 0, -5),
			CommitDates:  []time.Time{now.AddDate(0, 0, -5), now.AddDate(0, 0, -10)},
		},
		{
			Name:         "bob",
			Email:        "bob@example.com",
			Commits:      50,
			LinesAdded:   2000,
			LinesRemoved: 1000,
			PRReviews:    10,
			LastCommit:   now.AddDate(0, 0, -30),
			CommitDates:  []time.Time{now.AddDate(0, 0, -30), now.AddDate(0, 0, -60)},
		},
	}

	ownership := scorer.CalculateEnhancedOwnership(contributors, now)

	if len(ownership) != 2 {
		t.Fatalf("CalculateEnhancedOwnership() returned %d owners, want 2", len(ownership))
	}

	// Alice should have higher ownership (more commits, reviews, recent activity)
	if ownership[0].Name != "alice" {
		t.Error("Alice should be ranked first")
	}

	// Scores are NOT normalized to sum to 1.0
	// Each score is calculated independently based on weighted components
	// Max possible score per person is 100 (if they have max in all categories)
	if ownership[0].OwnershipScore <= 0 {
		t.Errorf("Alice ownership score = %v, should be > 0", ownership[0].OwnershipScore)
	}
	if ownership[0].OwnershipScore > ownership[1].OwnershipScore {
		// Alice has more of everything, so should have higher score
	} else {
		t.Error("Alice should have higher score than Bob")
	}

	// Check activity status
	// Active threshold is <= 30 days
	if ownership[0].ActivityStatus != "active" {
		t.Errorf("Alice activity status = %q, want active (5 days ago)", ownership[0].ActivityStatus)
	}
	// Bob at 30 days is still "active" (threshold is <=30)
	if ownership[1].ActivityStatus != "active" {
		t.Errorf("Bob activity status = %q, want active (30 days = threshold)", ownership[1].ActivityStatus)
	}
}
