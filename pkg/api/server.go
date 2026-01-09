// Package api provides the HTTP API layer for Zero
package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/crashappsec/zero/pkg/api/agent"
	"github.com/crashappsec/zero/pkg/api/handlers"
	"github.com/crashappsec/zero/pkg/api/jobs"
	"github.com/crashappsec/zero/pkg/api/ws"
	"github.com/crashappsec/zero/pkg/core/config"
	"github.com/crashappsec/zero/pkg/storage"
	"github.com/crashappsec/zero/pkg/storage/sqlite"
)

// Server is the HTTP API server
type Server struct {
	cfg          *config.Config
	zeroHome     string
	router       chi.Router
	hub          *ws.Hub
	queue        *jobs.Queue
	workerPool   *jobs.WorkerPool
	agentHandler *agent.Handler
	store        storage.Store
	port         int
	devMode      bool
}

// Options configures the server
type Options struct {
	Port       int
	DevMode    bool
	NumWorkers int // Number of scan workers (default: 1)
}

// NewServer creates a new API server
func NewServer(opts *Options) (*Server, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	zeroHome := cfg.ZeroHome()

	// Default workers
	if opts.NumWorkers <= 0 {
		opts.NumWorkers = 1
	}

	// Initialize SQLite store for fast queries
	dbPath := filepath.Join(zeroHome, "zero.db")
	store, err := sqlite.New(dbPath)
	if err != nil {
		log.Printf("Warning: Failed to initialize SQLite store: %v (falling back to JSON)", err)
		// Continue without store - handlers will fall back to JSON
	}

	hub := ws.NewHub()
	queue := jobs.NewQueue(100) // Max 100 queued jobs

	s := &Server{
		cfg:          cfg,
		zeroHome:     zeroHome,
		port:         opts.Port,
		devMode:      opts.DevMode,
		hub:          hub,
		queue:        queue,
		workerPool:   jobs.NewWorkerPool(queue, hub, opts.NumWorkers),
		agentHandler: agent.NewHandler(zeroHome),
		store:        store,
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
	// Note: Timeout middleware is applied per-route group to allow longer timeouts for chat

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
	repoHandler := handlers.NewProjectHandler(s.zeroHome, s.cfg) // Renamed: projects → repos
	analysisHandler := handlers.NewAnalysisHandler(s.zeroHome)
	systemHandler := handlers.NewSystemHandler(s.cfg)
	scanHandler := handlers.NewScanHandler(s.queue)
	configHandler := handlers.NewConfigHandler(s.cfg)

	// Standard API routes (with 60s timeout)
	r.Route("/api", func(r chi.Router) {
		// Group for standard routes with 60s timeout
		r.Group(func(r chi.Router) {
			r.Use(middleware.Timeout(60 * time.Second))

			// System endpoints
			r.Get("/health", systemHandler.Health)
			r.Get("/config", systemHandler.GetConfig)
			r.Get("/scanners", systemHandler.ListScanners)
			r.Get("/agents", systemHandler.ListAgents)

			// Repos endpoints (renamed from projects)
			r.Get("/repos", repoHandler.List)
			r.Get("/repos/{projectID}", repoHandler.Get)
			r.Delete("/repos/{projectID}", repoHandler.Delete)
			r.Get("/repos/{projectID}/freshness", repoHandler.GetFreshness)
			r.Get("/repos/{projectID}/analysis/{analysisType}", analysisHandler.GetAnalysis)

			// Backwards compatibility: /projects routes still work
			r.Get("/projects", repoHandler.List)
			r.Get("/projects/{projectID}", repoHandler.Get)
			r.Delete("/projects/{projectID}", repoHandler.Delete)
			r.Get("/projects/{projectID}/freshness", repoHandler.GetFreshness)
			r.Get("/projects/{projectID}/analysis/{analysisType}", analysisHandler.GetAnalysis)

			// Analysis aggregation endpoints
			r.Get("/analysis/stats", analysisHandler.GetAggregateStats)
			r.Get("/analysis/{projectID}/summary", analysisHandler.GetSummary)
			r.Get("/analysis/{projectID}/vulnerabilities", analysisHandler.GetVulnerabilities)
			r.Get("/analysis/{projectID}/secrets", analysisHandler.GetSecrets)
			r.Get("/analysis/{projectID}/dependencies", analysisHandler.GetDependencies)

			// Profile management
			r.Get("/profiles", configHandler.ListProfiles)
			r.Get("/profiles/{name}", configHandler.GetProfile)
			r.Post("/profiles", configHandler.CreateProfile)
			r.Put("/profiles/{name}", configHandler.UpdateProfile)
			r.Delete("/profiles/{name}", configHandler.DeleteProfile)

			// Settings management
			r.Get("/settings", configHandler.GetSettings)
			r.Put("/settings", configHandler.UpdateSettings)

			// Scanner configuration
			r.Get("/scanners/{name}", configHandler.GetScanner)
			r.Put("/scanners/{name}", configHandler.UpdateScanner)

			// Config export/import
			r.Get("/config/export", configHandler.ExportConfig)
			r.Post("/config/import", configHandler.ImportConfig)

			// Scan endpoints
			r.Post("/scans", scanHandler.Start)
			r.Get("/scans/active", scanHandler.ListActive)
			r.Get("/scans/history", scanHandler.ListHistory)
			r.Get("/scans/stats", scanHandler.Stats)
			r.Get("/scans/{jobID}", scanHandler.Get)
			r.Delete("/scans/{jobID}", scanHandler.Cancel)
		})

		// Agent chat endpoints (separate group with 5 min timeout for tool use)
		r.Group(func(r chi.Router) {
			r.Use(middleware.Timeout(300 * time.Second))
			r.Post("/chat", s.agentHandler.HandleChat)
			r.Post("/chat/stream", s.agentHandler.HandleChatStream)
			r.Get("/chat/sessions", s.agentHandler.HandleListSessions)
			r.Get("/chat/sessions/{sessionID}", s.agentHandler.HandleGetSession)
			r.Delete("/chat/sessions/{sessionID}", s.agentHandler.HandleDeleteSession)
		})
	})

	// WebSocket endpoints for real-time updates
	r.Get("/ws/scan/{jobID}", s.hub.HandleScanWS)
	r.Get("/ws/agent", s.agentHandler.HandleWebSocket)

	s.router = r
}

// Run starts the HTTP server
func (s *Server) Run(ctx context.Context) error {
	// Start WebSocket hub
	go s.hub.Run(ctx)

	// Start worker pool for scan jobs
	s.workerPool.Start(ctx)

	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Starting Zero API server on %s", addr)
	if s.devMode {
		log.Printf("Development mode enabled (CORS: *)")
	}

	srv := &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 300 * time.Second, // 5 min for chat with tool use
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	go func() {
		<-ctx.Done()
		log.Println("Shutting down server...")
		s.workerPool.Stop()
		if s.store != nil {
			s.store.Close()
		}
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
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
