package banter

import (
	"encoding/json"
	"net/http"
)

// Handler handles banter-related API requests
type Handler struct {
	service *Service
}

// NewHandler creates a new banter handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetStatus returns banter service status
func (h *Handler) GetStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"enabled":    h.service.IsEnabled(),
		"agents":     h.service.Generator().ListAgents(),
		"interval":   h.service.interval.String(),
	}
	writeJSON(w, http.StatusOK, status)
}

// SetEnabled enables or disables banter
func (h *Handler) SetEnabled(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Enabled bool `json:"enabled"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	h.service.SetEnabled(req.Enabled)
	writeJSON(w, http.StatusOK, map[string]bool{"enabled": req.Enabled})
}

// GenerateBanter generates a single banter message on demand
func (h *Handler) GenerateBanter(w http.ResponseWriter, r *http.Request) {
	var ctx Context
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&ctx); err != nil {
			// Ignore decode errors for optional body, use empty context
			ctx = Context{}
		}
	}

	// Temporarily enable generator for on-demand generation
	wasEnabled := h.service.IsEnabled()
	h.service.generator.SetEnabled(true)
	defer h.service.generator.SetEnabled(wasEnabled)

	msg := h.service.Generator().GenerateIdleBanter(&ctx)
	if msg == nil {
		writeError(w, http.StatusInternalServerError, "failed to generate banter", nil)
		return
	}

	writeJSON(w, http.StatusOK, msg)
}

// GenerateExchange generates a multi-agent exchange
func (h *Handler) GenerateExchange(w http.ResponseWriter, r *http.Request) {
	var ctx Context
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&ctx); err != nil {
			// Ignore decode errors for optional body, use empty context
			ctx = Context{}
		}
	}

	// Temporarily enable generator for on-demand generation
	wasEnabled := h.service.IsEnabled()
	h.service.generator.SetEnabled(true)
	defer h.service.generator.SetEnabled(wasEnabled)

	messages := h.service.Generator().GenerateExchange(&ctx)
	if len(messages) == 0 {
		writeError(w, http.StatusInternalServerError, "failed to generate exchange", nil)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"messages": messages,
		"count":    len(messages),
	})
}

// GetAgent returns a specific agent's personality
func (h *Handler) GetAgent(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "agent name required", nil)
		return
	}

	agent, ok := h.service.Generator().GetAgent(name)
	if !ok {
		writeError(w, http.StatusNotFound, "agent not found", nil)
		return
	}

	writeJSON(w, http.StatusOK, agent)
}

// ListAgents returns all agents with basic info
func (h *Handler) ListAgents(w http.ResponseWriter, r *http.Request) {
	gen := h.service.Generator()
	agents := gen.ListAgents()

	var agentList []map[string]string
	for _, name := range agents {
		if agent, ok := gen.GetAgent(name); ok {
			agentList = append(agentList, map[string]string{
				"id":          name,
				"name":        agent.Name,
				"full_name":   agent.FullName,
				"character":   agent.Character,
				"domain":      agent.Domain,
				"personality": agent.Personality,
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"agents": agentList,
		"count":  len(agentList),
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log encoding error but can't change status code after WriteHeader
		_ = err
	}
}

func writeError(w http.ResponseWriter, status int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := map[string]string{"error": message}
	if err != nil {
		resp["details"] = err.Error()
	}
	if encErr := json.NewEncoder(w).Encode(resp); encErr != nil {
		// Log encoding error but can't change status code after WriteHeader
		_ = encErr
	}
}
