// Package api provides the HTTP API layer for Zero
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/crashappsec/zero/pkg/api/handlers"
	"github.com/crashappsec/zero/pkg/api/ws"
	"github.com/crashappsec/zero/pkg/core/config"
)

// Server is the HTTP API server
type Server struct {
	cfg      *config.Config
	zeroHome string
	router   chi.Router
	hub      *ws.Hub
	port     int
	devMode  bool
}

// Options configures the server
type Options struct {
	Port    int
	DevMode bool
}

// NewServer creates a new API server
func NewServer(opts *Options) (*Server, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	zeroHome := cfg.ZeroHome()

	s := &Server{
		cfg:      cfg,
		zeroHome: zeroHome,
		port:     opts.Port,
		devMode:  opts.DevMode,
		hub:      ws.NewHub(),
	}

	s.setupRoutes()
	return s, nil
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	r := chi.NewRouter()

	// Base middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS configuration
	corsOpts := cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}
	if s.devMode {
		corsOpts.AllowedOrigins = []string{"*"}
	}
	r.Use(cors.Handler(corsOpts))

	// Create handlers
	projectHandler := handlers.NewProjectHandler(s.zeroHome, s.cfg)
	analysisHandler := handlers.NewAnalysisHandler(s.zeroHome)
	systemHandler := handlers.NewSystemHandler(s.cfg)

	// API routes
	r.Route("/api", func(r chi.Router) {
		// System endpoints
		r.Get("/health", systemHandler.Health)
		r.Get("/config", systemHandler.GetConfig)
		r.Get("/profiles", systemHandler.ListProfiles)
		r.Get("/scanners", systemHandler.ListScanners)
		r.Get("/agents", systemHandler.ListAgents)

		// Project endpoints
		r.Get("/projects", projectHandler.List)
		r.Get("/projects/{projectID}", projectHandler.Get)
		r.Delete("/projects/{projectID}", projectHandler.Delete)
		r.Get("/projects/{projectID}/freshness", projectHandler.GetFreshness)
		r.Get("/projects/{projectID}/analysis/{analysisType}", analysisHandler.GetAnalysis)

		// Analysis aggregation endpoints
		r.Get("/analysis/{projectID}/summary", analysisHandler.GetSummary)
		r.Get("/analysis/{projectID}/vulnerabilities", analysisHandler.GetVulnerabilities)
		r.Get("/analysis/{projectID}/secrets", analysisHandler.GetSecrets)
		r.Get("/analysis/{projectID}/dependencies", analysisHandler.GetDependencies)

		// Scan endpoints (Phase 2)
		// r.Post("/scans", scanHandler.Start)
		// r.Get("/scans/{jobID}", scanHandler.Get)
		// r.Delete("/scans/{jobID}", scanHandler.Cancel)
		// r.Get("/scans/active", scanHandler.ListActive)
	})

	// WebSocket endpoints (Phase 3)
	r.Get("/ws/scan/{jobID}", func(w http.ResponseWriter, r *http.Request) {
		// Placeholder for scan progress WebSocket
		json.NewEncoder(w).Encode(map[string]string{"error": "not implemented"})
	})
	r.Get("/ws/agent", func(w http.ResponseWriter, r *http.Request) {
		// Placeholder for agent chat WebSocket
		json.NewEncoder(w).Encode(map[string]string{"error": "not implemented"})
	})

	s.router = r
}

// Run starts the HTTP server
func (s *Server) Run(ctx context.Context) error {
	// Start WebSocket hub
	go s.hub.Run(ctx)

	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Starting Zero API server on %s", addr)
	if s.devMode {
		log.Printf("Development mode enabled (CORS: *)")
	}

	srv := &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

// Router returns the chi router for testing
func (s *Server) Router() chi.Router {
	return s.router
}
