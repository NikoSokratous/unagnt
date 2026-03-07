package orchestrate

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/NikoSokratous/unagnt/internal/store"
)

func TestHandleReplayDeadLetterQueuesNewRun(t *testing.T) {
	db := t.TempDir() + "/server-deadletter.db"
	st, err := store.NewSQLite(db)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()

	s := NewServer("localhost:0", st, nil)
	s.runner = NewRunner(s, fakeStepExecutor{}, 1, 8)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.runner.Start(ctx)

	sourceRunID := "dead-src-1"
	if err := st.SaveDeadLetter(context.Background(), &store.DeadLetter{
		RunID:      sourceRunID,
		AgentName:  "agent-replay",
		Goal:       "recover this run",
		Source:     "api",
		Error:      "boom",
		Payload:    `{"x":1}`,
		Attempt:    2,
		MaxRetries: 1,
		FailedAt:   time.Now(),
	}); err != nil {
		t.Fatalf("save dead letter: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/runs/dead-letters/"+sourceRunID+"/replay", nil)
	req.SetPathValue("id", sourceRunID)
	rr := httptest.NewRecorder()
	s.handleReplayDeadLetter(rr, req)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d body=%s", rr.Code, rr.Body.String())
	}

	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	newRunID := resp["replayed_run_id"]
	if newRunID == "" {
		t.Fatalf("missing replayed_run_id in response: %v", resp)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		meta, err := st.GetRun(context.Background(), newRunID)
		if err == nil && meta != nil && meta.State == "completed" {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("replayed run %s did not complete", newRunID)
}

func TestHandleReplayDeadLetterAppliesOverrides(t *testing.T) {
	db := t.TempDir() + "/server-deadletter-overrides.db"
	st, err := store.NewSQLite(db)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()

	s := NewServer("localhost:0", st, nil)
	s.runner = NewRunner(s, fakeStepExecutor{}, 1, 8)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.runner.Start(ctx)

	sourceRunID := "dead-src-override-1"
	if err := st.SaveDeadLetter(context.Background(), &store.DeadLetter{
		RunID:      sourceRunID,
		AgentName:  "agent-replay",
		Goal:       "original goal",
		Source:     "api",
		Error:      "boom",
		Payload:    `{"x":1}`,
		Attempt:    1,
		MaxRetries: 0,
		FailedAt:   time.Now(),
	}); err != nil {
		t.Fatalf("save dead letter: %v", err)
	}

	body := map[string]any{
		"goal":            "override goal",
		"max_retries":     2,
		"retry_backoff_ms": 5,
		"timeout_ms":      200,
		"payload":         map[string]any{"x": 2},
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/v1/runs/dead-letters/"+sourceRunID+"/replay", bytes.NewReader(b))
	req.SetPathValue("id", sourceRunID)
	rr := httptest.NewRecorder()
	s.handleReplayDeadLetter(rr, req)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d body=%s", rr.Code, rr.Body.String())
	}

	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	newRunID := resp["replayed_run_id"]
	if newRunID == "" {
		t.Fatalf("missing replayed_run_id in response: %v", resp)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		meta, err := st.GetRun(context.Background(), newRunID)
		if err == nil && meta != nil {
			if meta.Goal != "override goal" {
				t.Fatalf("expected overridden goal, got %q", meta.Goal)
			}
			if meta.State == "completed" {
				return
			}
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("replayed run %s did not complete", newRunID)
}

func TestHandleReplayDeadLetterRejectsNegativeOverrides(t *testing.T) {
	db := t.TempDir() + "/server-deadletter-negative.db"
	st, err := store.NewSQLite(db)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()

	s := NewServer("localhost:0", st, nil)
	s.runner = NewRunner(s, fakeStepExecutor{}, 1, 8)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.runner.Start(ctx)

	sourceRunID := "dead-src-negative-1"
	if err := st.SaveDeadLetter(context.Background(), &store.DeadLetter{
		RunID:      sourceRunID,
		AgentName:  "agent-replay",
		Goal:       "original goal",
		Source:     "api",
		Error:      "boom",
		Payload:    `{"x":1}`,
		Attempt:    1,
		MaxRetries: 0,
		FailedAt:   time.Now(),
	}); err != nil {
		t.Fatalf("save dead letter: %v", err)
	}

	body := map[string]any{"timeout_ms": -1}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/v1/runs/dead-letters/"+sourceRunID+"/replay", bytes.NewReader(b))
	req.SetPathValue("id", sourceRunID)
	rr := httptest.NewRecorder()
	s.handleReplayDeadLetter(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for negative override, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestHandleListDeadLettersSupportsFilters(t *testing.T) {
	db := t.TempDir() + "/server-deadletter-filters.db"
	st, err := store.NewSQLite(db)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()

	now := time.Now()
	_ = st.SaveDeadLetter(context.Background(), &store.DeadLetter{
		RunID:      "a1",
		AgentName:  "agent-a",
		Goal:       "g1",
		Source:     "api",
		Error:      "e1",
		Attempt:    1,
		MaxRetries: 0,
		FailedAt:   now,
	})
	_ = st.SaveDeadLetter(context.Background(), &store.DeadLetter{
		RunID:      "b1",
		AgentName:  "agent-b",
		Goal:       "g2",
		Source:     "webhook",
		Error:      "e2",
		Attempt:    1,
		MaxRetries: 0,
		FailedAt:   now.Add(time.Second),
	})

	s := NewServer("localhost:0", st, nil)
	req := httptest.NewRequest(http.MethodGet, "/v1/runs/dead-letters?limit=1&source=webhook", nil)
	rr := httptest.NewRecorder()
	s.handleListDeadLetters(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	var body struct {
		DeadLetters []store.DeadLetter `json:"dead_letters"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(body.DeadLetters) != 1 || body.DeadLetters[0].Source != "webhook" {
		t.Fatalf("unexpected filtered result: %+v", body.DeadLetters)
	}
}
