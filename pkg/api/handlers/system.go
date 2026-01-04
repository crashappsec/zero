package handlers

import (
	"net/http"
	"sort"
	"time"

	"github.com/crashappsec/zero/pkg/api/types"
	"github.com/crashappsec/zero/pkg/core/config"
)

// SystemHandler handles system-related API requests
type SystemHandler struct {
	cfg *config.Config
}

// NewSystemHandler creates a new system handler
func NewSystemHandler(cfg *config.Config) *SystemHandler {
	return &SystemHandler{
		cfg: cfg,
	}
}

// Health returns the health status
func (h *SystemHandler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, types.HealthResponse{
		Status:    "ok",
		Version:   "0.1.0-experimental",
		Timestamp: time.Now(),
	})
}

// GetConfig returns current configuration (non-sensitive)
func (h *SystemHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	cfg := map[string]interface{}{
		"default_profile":    h.cfg.Settings.DefaultProfile,
		"parallel_repos":     h.cfg.Settings.ParallelRepos,
		"parallel_scanners":  h.cfg.Settings.ParallelScanners,
		"scanner_timeout":    h.cfg.Settings.ScannerTimeoutSeconds,
		"available_profiles": h.cfg.GetProfileNames(),
	}
	writeJSON(w, http.StatusOK, cfg)
}

// ListProfiles returns available scan profiles
func (h *SystemHandler) ListProfiles(w http.ResponseWriter, r *http.Request) {
	var profiles []types.ProfileInfo

	for name, p := range h.cfg.Profiles {
		profiles = append(profiles, types.ProfileInfo{
			Name:          name,
			Description:   p.Description,
			EstimatedTime: p.EstimatedTime,
			Scanners:      p.Scanners,
		})
	}

	// Sort by name for consistent ordering
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Name < profiles[j].Name
	})

	writeJSON(w, http.StatusOK, types.ListResponse[types.ProfileInfo]{
		Data:  profiles,
		Total: len(profiles),
	})
}

// ListScanners returns available scanners
func (h *SystemHandler) ListScanners(w http.ResponseWriter, r *http.Request) {
	// Define the 7 v4.0 super scanners
	scanners := []types.ScannerInfo{
		{
			Name:        "code-packages",
			Description: "SBOM generation and package analysis (vulns, health, licenses, malcontent, provenance)",
			Features:    []string{"generation", "integrity", "vulns", "health", "licenses", "malcontent", "confusion", "typosquats", "deprecations", "duplicates", "reachability", "provenance", "bundle", "recommendations"},
		},
		{
			Name:        "code-security",
			Description: "Security-focused code analysis (SAST, secrets, API security, cryptography)",
			Features:    []string{"vulns", "secrets", "api", "ciphers", "keys", "random", "tls", "certificates"},
		},
		{
			Name:        "code-quality",
			Description: "Code quality metrics",
			Features:    []string{"tech_debt", "complexity", "test_coverage", "documentation"},
		},
		{
			Name:        "devops",
			Description: "DevOps and CI/CD security",
			Features:    []string{"iac", "containers", "github_actions", "dora", "git"},
		},
		{
			Name:        "technology-identification",
			Description: "Technology detection and ML-BOM generation",
			Features:    []string{"detection", "models", "frameworks", "datasets", "ai_security", "ai_governance", "infrastructure"},
		},
		{
			Name:        "code-ownership",
			Description: "Code ownership analysis",
			Features:    []string{"contributors", "bus_factor", "codeowners", "orphans", "churn", "patterns"},
		},
		{
			Name:        "developer-experience",
			Description: "Developer experience analysis",
			Features:    []string{"onboarding", "sprawl", "workflow"},
		},
	}

	writeJSON(w, http.StatusOK, types.ListResponse[types.ScannerInfo]{
		Data:  scanners,
		Total: len(scanners),
	})
}

// ListAgents returns available specialist agents
func (h *SystemHandler) ListAgents(w http.ResponseWriter, r *http.Request) {
	agents := []types.AgentInfo{
		{ID: "zero", Name: "Zero", Persona: "Zero Cool", Description: "Master orchestrator", Scanner: "all"},
		{ID: "cereal", Name: "Cereal", Persona: "Cereal Killer", Description: "Supply chain security", Scanner: "code-packages"},
		{ID: "razor", Name: "Razor", Persona: "Razor", Description: "Code security, SAST, secrets", Scanner: "code-security"},
		{ID: "blade", Name: "Blade", Persona: "Blade", Description: "Compliance, SOC 2, ISO 27001", Scanner: "multiple"},
		{ID: "phreak", Name: "Phreak", Persona: "Phantom Phreak", Description: "Legal, licenses, privacy", Scanner: "code-packages"},
		{ID: "acid", Name: "Acid", Persona: "Acid Burn", Description: "Frontend, React, TypeScript", Scanner: "code-security"},
		{ID: "dade", Name: "Dade", Persona: "Dade Murphy", Description: "Backend, APIs, databases", Scanner: "code-security"},
		{ID: "nikon", Name: "Nikon", Persona: "Lord Nikon", Description: "Architecture, system design", Scanner: "technology-identification"},
		{ID: "joey", Name: "Joey", Persona: "Joey", Description: "CI/CD, build optimization", Scanner: "devops"},
		{ID: "plague", Name: "Plague", Persona: "The Plague", Description: "DevOps, IaC, Kubernetes", Scanner: "devops"},
		{ID: "gibson", Name: "Gibson", Persona: "The Gibson", Description: "DORA metrics, team health", Scanner: "devops"},
		{ID: "gill", Name: "Gill", Persona: "Gill Bates", Description: "Cryptography specialist", Scanner: "code-security"},
		{ID: "turing", Name: "Turing", Persona: "Alan Turing", Description: "AI/ML security", Scanner: "technology-identification"},
	}

	writeJSON(w, http.StatusOK, types.ListResponse[types.AgentInfo]{
		Data:  agents,
		Total: len(agents),
	})
}
