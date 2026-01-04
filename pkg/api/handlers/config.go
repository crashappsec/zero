package handlers

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/go-chi/chi/v5"

	"github.com/crashappsec/zero/pkg/api/types"
	"github.com/crashappsec/zero/pkg/core/config"
)

// ConfigHandler handles configuration-related API requests
type ConfigHandler struct {
	cfg *config.Config
}

// NewConfigHandler creates a new config handler
func NewConfigHandler(cfg *config.Config) *ConfigHandler {
	return &ConfigHandler{
		cfg: cfg,
	}
}

// GetSettings returns current settings
func (h *ConfigHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	settings := map[string]interface{}{
		"default_profile":         h.cfg.Settings.DefaultProfile,
		"storage_path":            h.cfg.Settings.StoragePath,
		"parallel_repos":          h.cfg.Settings.ParallelRepos,
		"parallel_scanners":       h.cfg.Settings.ParallelScanners,
		"scanner_timeout_seconds": h.cfg.Settings.ScannerTimeoutSeconds,
		"cache_ttl_hours":         h.cfg.Settings.CacheTTLHours,
	}
	writeJSON(w, http.StatusOK, settings)
}

// UpdateSettings updates global settings
func (h *ConfigHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DefaultProfile        string `json:"default_profile,omitempty"`
		ParallelRepos         int    `json:"parallel_repos,omitempty"`
		ParallelScanners      int    `json:"parallel_scanners,omitempty"`
		ScannerTimeoutSeconds int    `json:"scanner_timeout_seconds,omitempty"`
		CacheTTLHours         int    `json:"cache_ttl_hours,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Update only provided fields
	if req.DefaultProfile != "" {
		h.cfg.Settings.DefaultProfile = req.DefaultProfile
	}
	if req.ParallelRepos > 0 {
		h.cfg.Settings.ParallelRepos = req.ParallelRepos
	}
	if req.ParallelScanners > 0 {
		h.cfg.Settings.ParallelScanners = req.ParallelScanners
	}
	if req.ScannerTimeoutSeconds > 0 {
		h.cfg.Settings.ScannerTimeoutSeconds = req.ScannerTimeoutSeconds
	}
	if req.CacheTTLHours > 0 {
		h.cfg.Settings.CacheTTLHours = req.CacheTTLHours
	}

	if err := h.cfg.Save(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save settings", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// ListProfiles returns all profiles with full details
func (h *ConfigHandler) ListProfiles(w http.ResponseWriter, r *http.Request) {
	var profiles []types.ProfileInfo

	for name, p := range h.cfg.Profiles {
		profiles = append(profiles, types.ProfileInfo{
			Name:          name,
			Description:   p.Description,
			EstimatedTime: p.EstimatedTime,
			Scanners:      p.Scanners,
		})
	}

	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Name < profiles[j].Name
	})

	writeJSON(w, http.StatusOK, types.ListResponse[types.ProfileInfo]{
		Data:  profiles,
		Total: len(profiles),
	})
}

// GetProfile returns a single profile
func (h *ConfigHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	profile, ok := h.cfg.GetProfile(name)
	if !ok {
		writeError(w, http.StatusNotFound, "profile not found", nil)
		return
	}

	writeJSON(w, http.StatusOK, types.ProfileInfo{
		Name:          name,
		Description:   profile.Description,
		EstimatedTime: profile.EstimatedTime,
		Scanners:      profile.Scanners,
	})
}

// CreateProfile creates a new profile
func (h *ConfigHandler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name          string   `json:"name"`
		Description   string   `json:"description"`
		EstimatedTime string   `json:"estimated_time"`
		Scanners      []string `json:"scanners"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "profile name is required", nil)
		return
	}

	if len(req.Scanners) == 0 {
		writeError(w, http.StatusBadRequest, "at least one scanner is required", nil)
		return
	}

	profile := config.Profile{
		Name:          req.Name,
		Description:   req.Description,
		EstimatedTime: req.EstimatedTime,
		Scanners:      req.Scanners,
	}

	if err := h.cfg.CreateProfile(req.Name, profile); err != nil {
		writeError(w, http.StatusConflict, err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusCreated, types.ProfileInfo{
		Name:          req.Name,
		Description:   req.Description,
		EstimatedTime: req.EstimatedTime,
		Scanners:      req.Scanners,
	})
}

// UpdateProfile updates an existing profile
func (h *ConfigHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var req struct {
		Description   string   `json:"description,omitempty"`
		EstimatedTime string   `json:"estimated_time,omitempty"`
		Scanners      []string `json:"scanners,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	existing, ok := h.cfg.GetProfile(name)
	if !ok {
		writeError(w, http.StatusNotFound, "profile not found", nil)
		return
	}

	// Update only provided fields
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.EstimatedTime != "" {
		existing.EstimatedTime = req.EstimatedTime
	}
	if len(req.Scanners) > 0 {
		existing.Scanners = req.Scanners
	}

	if err := h.cfg.UpdateProfile(name, *existing); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusOK, types.ProfileInfo{
		Name:          name,
		Description:   existing.Description,
		EstimatedTime: existing.EstimatedTime,
		Scanners:      existing.Scanners,
	})
}

// DeleteProfile removes a profile
func (h *ConfigHandler) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	if err := h.cfg.DeleteProfile(name); err != nil {
		if err.Error() == "profile not found: "+name {
			writeError(w, http.StatusNotFound, err.Error(), nil)
		} else {
			writeError(w, http.StatusBadRequest, err.Error(), nil)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetScanner returns scanner configuration
func (h *ConfigHandler) GetScanner(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	scanner, ok := h.cfg.GetScanner(name)
	if !ok {
		writeError(w, http.StatusNotFound, "scanner not found", nil)
		return
	}

	// Get features from raw config
	features := h.cfg.GetScannerFeatures(name)

	resp := map[string]interface{}{
		"name":           scanner.Name,
		"description":    scanner.Description,
		"estimated_time": scanner.EstimatedTime,
		"output_file":    scanner.OutputFile,
		"features":       features,
	}

	writeJSON(w, http.StatusOK, resp)
}

// UpdateScanner updates scanner configuration
func (h *ConfigHandler) UpdateScanner(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var req struct {
		Description   string                 `json:"description,omitempty"`
		EstimatedTime string                 `json:"estimated_time,omitempty"`
		Features      map[string]interface{} `json:"features,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	existing, ok := h.cfg.GetScanner(name)
	if !ok {
		writeError(w, http.StatusNotFound, "scanner not found", nil)
		return
	}

	// Update only provided fields
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.EstimatedTime != "" {
		existing.EstimatedTime = req.EstimatedTime
	}

	if err := h.cfg.UpdateScanner(name, *existing); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// ExportConfig returns full config as JSON
func (h *ConfigHandler) ExportConfig(w http.ResponseWriter, r *http.Request) {
	data, err := h.cfg.Export()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to export config", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=zero.config.json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// ImportConfig imports config from JSON
func (h *ConfigHandler) ImportConfig(w http.ResponseWriter, r *http.Request) {
	var raw json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON", err)
		return
	}

	newCfg, err := config.Import(raw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid config", err)
		return
	}

	// Update in-memory config
	*h.cfg = *newCfg

	writeJSON(w, http.StatusOK, map[string]string{"status": "imported"})
}
