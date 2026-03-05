package observe

import (
	"encoding/json"

	"github.com/NikoSokratous/unagnt/pkg/runtime"
)

// Tracer records tool calls and reasoning for traces.
type Tracer struct {
	events []Event
}

// NewTracer creates a tracer.
func NewTracer() *Tracer {
	return &Tracer{events: nil}
}

// RecordStep appends a step to the trace with reasoning summary.
func (t *Tracer) RecordStep(runID string, step runtime.StepRecord, model ModelMeta) {
	if step.Action != nil {
		t.events = append(t.events, Event{
			RunID:     runID,
			StepID:    step.StepID,
			Timestamp: step.Timestamp,
			Type:      EventToolCall,
			Data: map[string]any{
				"tool":    step.Action.Tool,
				"version": step.Action.Version,
				"input":   step.Action.Input,
			},
			Model: model,
		})
	}
	if step.Result != nil {
		duration := ""
		if step.Result.Duration > 0 {
			duration = step.Result.Duration.String()
		}
		t.events = append(t.events, Event{
			RunID:     runID,
			StepID:    step.StepID,
			Timestamp: step.Timestamp,
			Type:      EventToolResult,
			Data: map[string]any{
				"tool":     step.Result.ToolID,
				"output":   step.Result.Output,
				"error":    step.Result.Error,
				"duration": duration,
			},
		})
	}
	if step.Reasoning != "" {
		t.events = append(t.events, Event{
			RunID:     runID,
			StepID:    step.StepID,
			Timestamp: step.Timestamp,
			Type:      EventReasoning,
			Data:      map[string]any{"reasoning": step.Reasoning},
			Model:     model,
		})
	}
}

// Events returns the recorded events.
func (t *Tracer) Events() []Event {
	return t.events
}

// ToJSON serializes the trace to JSON.
func (t *Tracer) ToJSON() ([]byte, error) {
	return json.MarshalIndent(t.events, "", "  ")
}
