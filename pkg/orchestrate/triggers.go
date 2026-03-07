package orchestrate

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/NikoSokratous/unagnt/pkg/observe"
)

// TriggerEvent is an event-driven execution request.
type TriggerEvent struct {
	AgentName string                 `json:"agent_name"`
	Goal      string                 `json:"goal"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
}

// EventTriggerBus is an in-process event-driven trigger source.
type EventTriggerBus struct {
	ch chan TriggerEvent
}

func NewEventTriggerBus(buffer int) *EventTriggerBus {
	if buffer <= 0 {
		buffer = 128
	}
	return &EventTriggerBus{ch: make(chan TriggerEvent, buffer)}
}

// RegisterEventHubTrigger subscribes to an EventHub run stream and forwards
// trigger metadata into the event-driven pipeline.
func (s *Server) RegisterEventHubTrigger(runID string) {
	if s.eventHub == nil || s.triggerBus == nil {
		return
	}
	ch := s.eventHub.Subscribe(runID)
	go func() {
		defer s.eventHub.Unsubscribe(runID, ch)
		for evt := range ch {
			if evt.Type != observe.EventCompleted {
				continue
			}
			agent, _ := evt.Data["trigger_agent"].(string)
			goal, _ := evt.Data["trigger_goal"].(string)
			if agent == "" || goal == "" {
				continue
			}
			payload := map[string]interface{}{}
			for k, v := range evt.Data {
				payload[k] = v
			}
			_ = s.triggerBus.Publish(TriggerEvent{
				AgentName: agent,
				Goal:      goal,
				Payload:   payload,
			})
		}
	}()
}

func (b *EventTriggerBus) Publish(evt TriggerEvent) bool {
	select {
	case b.ch <- evt:
		return true
	default:
		return false
	}
}

func (b *EventTriggerBus) Start(ctx context.Context, handler func(context.Context, TriggerEvent) error) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case evt := <-b.ch:
				_ = handler(ctx, evt)
			}
		}
	}()
}

func (s *Server) handlePublishEventTrigger(w http.ResponseWriter, r *http.Request) {
	if s.triggerBus == nil {
		http.Error(w, "trigger bus unavailable", http.StatusServiceUnavailable)
		return
	}

	var evt TriggerEvent
	if err := json.NewDecoder(r.Body).Decode(&evt); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if evt.AgentName == "" || evt.Goal == "" {
		http.Error(w, "agent_name and goal are required", http.StatusBadRequest)
		return
	}

	if ok := s.triggerBus.Publish(evt); !ok {
		http.Error(w, "trigger queue is full", http.StatusTooManyRequests)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "accepted",
	})
}
