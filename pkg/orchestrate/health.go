package orchestrate

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/NikoSokratous/unagnt/internal/logging"
)

// HealthHandler handles /health endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// ReadyHandler handles /ready endpoint
func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	log := logging.Logger()

	// Check if store is accessible
	ctx := r.Context()
	_, err := s.store.ListRuns(ctx, 1)
	if err != nil {
		log.Warn().Err(err).Msg("readiness check failed: store unavailable")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]any{
			"status": "not ready",
			"error":  "store unavailable",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"status":    "ready",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
