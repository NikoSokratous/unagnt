package orchestrate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/NikoSokratous/unagnt/pkg/observe"
)

// handleStream streams events for a run using Server-Sent Events (SSE).
func (s *Server) handleStream(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("id")
	if runID == "" {
		http.Error(w, "run ID required", http.StatusBadRequest)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Flush headers immediately
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	// Subscribe to events
	eventChan := s.eventHub.Subscribe(runID)
	defer s.eventHub.Unsubscribe(runID, eventChan)

	// Heartbeat ticker (every 30s)
	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	ctx := r.Context()

	for {
		select {
		case <-ctx.Done():
			// Client disconnected
			return

		case event, ok := <-eventChan:
			if !ok {
				// Channel closed, run completed
				return
			}

			// Marshal event to JSON
			data, err := json.Marshal(event)
			if err != nil {
				continue
			}

			// Send SSE message
			fmt.Fprintf(w, "data: %s\n\n", data)

			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}

			// If completed/failed/interrupted, close after sending
			if event.Type == observe.EventCompleted ||
				event.Type == observe.EventError ||
				event.Type == observe.EventInterrupted {
				return
			}

		case <-heartbeat.C:
			// Send heartbeat comment to keep connection alive
			fmt.Fprintf(w, ": heartbeat\n\n")
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}
	}
}

// StreamEvent wraps an event with SSE metadata.
type StreamEvent struct {
	ID    string        `json:"id,omitempty"`
	Event string        `json:"event,omitempty"`
	Data  observe.Event `json:"data"`
}
