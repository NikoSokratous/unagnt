package orchestrate

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/NikoSokratous/unagnt/internal/store"
	"github.com/NikoSokratous/unagnt/pkg/api"
	"github.com/NikoSokratous/unagnt/pkg/observe"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server is a simple HTTP API server for run management.
type Server struct {
	addr        string
	store       *store.SQLite
	mu          sync.Mutex
	runs        map[string]context.CancelFunc
	auth        *AuthMiddleware
	apiKeys     []string
	eventHub    *observe.EventHub
	rateLimiter *RateLimiter
	rateLimitMw *RateLimitMiddleware
}

// ServerConfig holds server configuration.
type ServerConfig struct {
	Addr      string
	Store     *store.SQLite
	APIKeys   []string
	RateLimit RateLimitConfig
}

// NewServer creates an API server.
func NewServer(addr string, st *store.SQLite, apiKeys []string) *Server {
	return NewServerWithConfig(ServerConfig{
		Addr:    addr,
		Store:   st,
		APIKeys: apiKeys,
		RateLimit: RateLimitConfig{
			Enabled: false, // Disabled by default
		},
	})
}

// NewServerWithConfig creates an API server with full configuration.
func NewServerWithConfig(config ServerConfig) *Server {
	s := &Server{
		addr:     config.Addr,
		store:    config.Store,
		runs:     make(map[string]context.CancelFunc),
		apiKeys:  config.APIKeys,
		eventHub: observe.NewEventHub(100),
	}

	// Setup auth middleware if API keys provided
	if len(config.APIKeys) > 0 {
		s.auth = NewAuthMiddleware(config.APIKeys)
	}

	// Setup rate limiting if enabled
	if config.RateLimit.Enabled {
		rateLimiter, err := NewRateLimiter(config.RateLimit)
		if err == nil {
			s.rateLimiter = rateLimiter
			s.rateLimitMw = NewRateLimitMiddleware(rateLimiter, config.RateLimit)
		}
		// If rate limiter setup fails, continue without rate limiting
	}

	return s
}

// Run starts the HTTP server.
func (s *Server) Run() error {
	mux := http.NewServeMux()

	// Setup Web UI routes (served on root paths)
	s.SetupWebUIRoutes(mux)

	// Health and readiness (no auth required)
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /ready", s.handleReady)

	// Metrics endpoint (no auth required)
	mux.Handle("GET /metrics", promhttp.Handler())

	// Run management (auth required if configured)
	mux.HandleFunc("POST /v1/runs", s.handleCreateRun)
	mux.HandleFunc("GET /v1/runs/{id}", s.handleGetRun)
	mux.HandleFunc("GET /v1/runs", s.handleListRuns)
	mux.HandleFunc("POST /v1/runs/{id}/cancel", s.handleCancelRun)
	mux.HandleFunc("GET /v1/runs/{id}/stream", s.handleStream)

	// Policy Playground
	mux.HandleFunc("POST /v1/policy/check", api.HandlePolicyCheck)

	// Apply middlewares
	var handler http.Handler = mux

	// Apply rate limiting first (outermost)
	if s.rateLimitMw != nil {
		handler = s.rateLimitMw.Middleware(handler)
	}

	// Apply auth middleware (innermost, closer to handlers)
	if s.auth != nil {
		handler = s.auth.Middleware(handler)
	}

	return http.ListenAndServe(s.addr, handler)
}

func (s *Server) handleCreateRun(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AgentName string `json:"agent_name"`
		Goal      string `json:"goal"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	runID := uuid.New().String()
	now := time.Now()
	run := &store.RunMeta{
		RunID:     runID,
		AgentName: req.AgentName,
		Goal:      req.Goal,
		State:     "pending",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.store.SaveRun(r.Context(), run); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update metrics
	RunsCreated.Inc()
	RunsActive.Inc()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"run_id": runID})
}

func (s *Server) handleGetRun(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "run_id required", http.StatusBadRequest)
		return
	}
	run, err := s.store.GetRun(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if run == nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(run)
}

func (s *Server) handleListRuns(w http.ResponseWriter, r *http.Request) {
	ids, err := s.store.ListRuns(r.Context(), 100)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{"run_ids": ids})
}

func (s *Server) handleCancelRun(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "run_id required", http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	cancel, ok := s.runs[id]
	s.mu.Unlock()
	if ok && cancel != nil {
		cancel()
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"cancelled": ok})
}
