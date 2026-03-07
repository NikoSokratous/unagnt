package orchestrate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/NikoSokratous/unagnt/internal/config"
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
	runner      *Runner
	triggerBus  *EventTriggerBus
	scheduler   *Scheduler
	webhooks    *WebhookHandler
}

// ServerConfig holds server configuration.
type ServerConfig struct {
	Addr      string
	Store     *store.SQLite
	APIKeys   []string
	RateLimit RateLimitConfig
	Webhooks  []config.WebhookConfig
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
		addr:       config.Addr,
		store:      config.Store,
		runs:       make(map[string]context.CancelFunc),
		apiKeys:    config.APIKeys,
		eventHub:   observe.NewEventHub(100),
		triggerBus: NewEventTriggerBus(256),
	}

	s.runner = NewRunner(s, &RuntimeStepExecutor{
		AllowSimulatedFallback: true,
		StorePath:              "agent.db",
	}, 2, 256)
	s.scheduler = NewScheduler(func(ctx context.Context, agent, goal string) error {
		if s.runner == nil {
			return nil
		}
		return s.runner.Submit(RunRequest{
			RunID:     uuid.New().String(),
			AgentName: agent,
			Goal:      goal,
			Source:    "schedule",
		})
	})

	if len(config.Webhooks) > 0 {
		webhooks, err := NewWebhookHandler(s, config.Webhooks)
		if err == nil {
			s.webhooks = webhooks
		}
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
	mux.HandleFunc("GET /v1/runs/{id}/events", s.handleGetRunEvents)
	mux.HandleFunc("GET /v1/runs", s.handleListRuns)
	mux.HandleFunc("GET /v1/runs/dead-letters", s.handleListDeadLetters)
	mux.HandleFunc("POST /v1/runs/dead-letters/{id}/replay", s.handleReplayDeadLetter)
	mux.HandleFunc("POST /v1/runs/{id}/cancel", s.handleCancelRun)
	mux.HandleFunc("GET /v1/runs/{id}/stream", s.handleStream)
	mux.HandleFunc("POST /v1/triggers/events", s.handlePublishEventTrigger)
	if s.webhooks != nil {
		for path := range s.webhooks.webhooks {
			mux.HandleFunc("POST "+path, s.webhooks.HandleWebhook)
		}
	}

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

	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if s.runner != nil {
		s.runner.Start(rootCtx)
	}
	if s.triggerBus != nil && s.runner != nil {
		s.triggerBus.Start(rootCtx, func(ctx context.Context, evt TriggerEvent) error {
			return s.runner.Submit(RunRequest{
				RunID:     uuid.New().String(),
				AgentName: evt.AgentName,
				Goal:      evt.Goal,
				Source:    "event",
				Outputs:   evt.Payload,
			})
		})
	}
	if s.scheduler != nil {
		go func() {
			_ = s.scheduler.Run(rootCtx)
		}()
	}

	return http.ListenAndServe(s.addr, handler)
}

func (s *Server) handleCreateRun(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AgentName      string `json:"agent_name"`
		Goal           string `json:"goal"`
		MaxRetries     int    `json:"max_retries,omitempty"`
		RetryBackoffMs int    `json:"retry_backoff_ms,omitempty"`
		TimeoutMs      int    `json:"timeout_ms,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.MaxRetries < 0 || req.RetryBackoffMs < 0 || req.TimeoutMs < 0 {
		http.Error(w, "max_retries, retry_backoff_ms, and timeout_ms must be >= 0", http.StatusBadRequest)
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

	if s.runner != nil {
		if err := s.runner.Submit(RunRequest{
			RunID:        runID,
			AgentName:    req.AgentName,
			Goal:         req.Goal,
			Source:       "api",
			MaxRetries:   req.MaxRetries,
			RetryBackoff: time.Duration(req.RetryBackoffMs) * time.Millisecond,
			Timeout:      time.Duration(req.TimeoutMs) * time.Millisecond,
		}); err != nil {
			run.State = "failed"
			run.UpdatedAt = time.Now()
			_ = s.store.SaveRun(r.Context(), run)
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
	}

	// Update metrics
	RunsCreated.Inc()

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

func (s *Server) handleGetRunEvents(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "run_id required", http.StatusBadRequest)
		return
	}
	evts, err := s.store.GetEvents(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"run_id": id,
		"events": evts,
	})
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

func (s *Server) handleListDeadLetters(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if raw := r.URL.Query().Get("limit"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v <= 0 {
			http.Error(w, "limit must be a positive integer", http.StatusBadRequest)
			return
		}
		limit = v
	}
	source := r.URL.Query().Get("source")

	deadLetters, err := s.store.ListDeadLetters(r.Context(), limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if source != "" {
		filtered := make([]store.DeadLetter, 0, len(deadLetters))
		for _, dl := range deadLetters {
			if dl.Source == source {
				filtered = append(filtered, dl)
			}
		}
		deadLetters = filtered
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"dead_letters": deadLetters,
	})
}

func (s *Server) handleReplayDeadLetter(w http.ResponseWriter, r *http.Request) {
	if s.runner == nil {
		http.Error(w, "runner unavailable", http.StatusServiceUnavailable)
		return
	}
	sourceRunID := r.PathValue("id")
	if sourceRunID == "" {
		http.Error(w, "run_id required", http.StatusBadRequest)
		return
	}

	dl, err := s.store.GetLatestDeadLetterByRunID(r.Context(), sourceRunID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if dl == nil {
		http.Error(w, fmt.Sprintf("dead-letter for run %s not found", sourceRunID), http.StatusNotFound)
		return
	}

	var replayReq struct {
		Goal           string                 `json:"goal,omitempty"`
		MaxRetries     *int                   `json:"max_retries,omitempty"`
		RetryBackoffMs *int                   `json:"retry_backoff_ms,omitempty"`
		TimeoutMs      *int                   `json:"timeout_ms,omitempty"`
		Payload        map[string]interface{} `json:"payload,omitempty"`
	}
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&replayReq); err != nil && !errors.Is(err, io.EOF) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	replayGoal := dl.Goal
	if replayReq.Goal != "" {
		replayGoal = replayReq.Goal
	}

	outputs := replayReq.Payload
	if outputs == nil {
		outputs = map[string]interface{}{}
		if dl.Payload != "" {
			_ = json.Unmarshal([]byte(dl.Payload), &outputs)
		}
	}

	maxRetries := dl.MaxRetries
	if replayReq.MaxRetries != nil {
		if *replayReq.MaxRetries < 0 {
			http.Error(w, "max_retries must be >= 0", http.StatusBadRequest)
			return
		}
		maxRetries = *replayReq.MaxRetries
	}
	retryBackoff := time.Duration(0)
	if replayReq.RetryBackoffMs != nil {
		if *replayReq.RetryBackoffMs < 0 {
			http.Error(w, "retry_backoff_ms must be >= 0", http.StatusBadRequest)
			return
		}
		retryBackoff = time.Duration(*replayReq.RetryBackoffMs) * time.Millisecond
	}
	timeout := time.Duration(0)
	if replayReq.TimeoutMs != nil {
		if *replayReq.TimeoutMs < 0 {
			http.Error(w, "timeout_ms must be >= 0", http.StatusBadRequest)
			return
		}
		timeout = time.Duration(*replayReq.TimeoutMs) * time.Millisecond
	}

	newRunID := uuid.New().String()
	now := time.Now()
	run := &store.RunMeta{
		RunID:     newRunID,
		AgentName: dl.AgentName,
		Goal:      replayGoal,
		State:     "pending",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.store.SaveRun(r.Context(), run); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.runner.Submit(RunRequest{
		RunID:        newRunID,
		AgentName:    dl.AgentName,
		Goal:         replayGoal,
		Source:       "dead-letter-replay",
		Outputs:      outputs,
		MaxRetries:   maxRetries,
		RetryBackoff: retryBackoff,
		Timeout:      timeout,
	}); err != nil {
		run.State = "failed"
		run.UpdatedAt = time.Now()
		_ = s.store.SaveRun(r.Context(), run)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":          "accepted",
		"source_run_id":   sourceRunID,
		"replayed_run_id": newRunID,
	})
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

func (s *Server) setRunCancel(runID string, cancel context.CancelFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.runs[runID] = cancel
}

func (s *Server) clearRunCancel(runID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.runs, runID)
}

// AddScheduledJob registers a cron-based scheduled execution routed through runner.
func (s *Server) AddScheduledJob(job Job) {
	if s.scheduler == nil {
		return
	}
	s.scheduler.Add(job)
}
