package orchestrate

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/NikoSokratous/unagnt/internal/store"
	"github.com/NikoSokratous/unagnt/pkg/observe"
)

type fakeStepExecutor struct{}

func (fakeStepExecutor) ExecuteStep(ctx context.Context, agentName, goal string, outputs map[string]interface{}) (*StepResult, error) {
	now := time.Now()
	return &StepResult{
		Name:        agentName,
		Agent:       agentName,
		Status:      "completed",
		RunID:       "fake-run",
		Output:      map[string]interface{}{"goal": goal},
		StartedAt:   now,
		CompletedAt: now,
	}, nil
}

type failingStepExecutor struct{}

func (failingStepExecutor) ExecuteStep(ctx context.Context, agentName, goal string, outputs map[string]interface{}) (*StepResult, error) {
	return nil, errors.New("boom")
}

type flakyStepExecutor struct {
	mu         sync.Mutex
	failures   int
	attempts   int
	resultGoal string
}

type slowStepExecutor struct {
	delay time.Duration
}

func (s slowStepExecutor) ExecuteStep(ctx context.Context, agentName, goal string, outputs map[string]interface{}) (*StepResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(s.delay):
	}
	now := time.Now()
	return &StepResult{
		Name:        agentName,
		Agent:       agentName,
		Status:      "completed",
		RunID:       "slow-run",
		Output:      map[string]interface{}{"goal": goal},
		StartedAt:   now,
		CompletedAt: now,
	}, nil
}

func (f *flakyStepExecutor) ExecuteStep(ctx context.Context, agentName, goal string, outputs map[string]interface{}) (*StepResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.attempts++
	if f.attempts <= f.failures {
		return nil, errors.New("transient failure")
	}
	now := time.Now()
	f.resultGoal = goal
	return &StepResult{
		Name:        agentName,
		Agent:       agentName,
		Status:      "completed",
		RunID:       "flaky-run",
		Output:      map[string]interface{}{"goal": goal},
		StartedAt:   now,
		CompletedAt: now,
	}, nil
}

func TestRunnerExecutesQueuedRun(t *testing.T) {
	db := t.TempDir() + "/runner.db"
	st, err := store.NewSQLite(db)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()

	s := NewServer("localhost:0", st, nil)
	r := NewRunner(s, fakeStepExecutor{}, 1, NewMemoryQueue(4))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r.Start(ctx)

	runID := "run-test-1"
	if err := r.Submit(RunRequest{
		RunID:     runID,
		AgentName: "agent-a",
		Goal:      "hello",
		Source:    "test",
	}); err != nil {
		t.Fatalf("submit: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		meta, err := st.GetRun(context.Background(), runID)
		if err == nil && meta != nil && meta.State == "completed" {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}

	t.Fatalf("run %s did not complete in time", runID)
}

func TestRunnerEmitsLifecycleEvents(t *testing.T) {
	db := t.TempDir() + "/runner-events.db"
	st, err := store.NewSQLite(db)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()

	s := NewServer("localhost:0", st, nil)
	r := NewRunner(s, fakeStepExecutor{}, 1, NewMemoryQueue(4))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r.Start(ctx)

	runID := "run-events-1"
	ch := s.eventHub.Subscribe(runID)
	defer s.eventHub.Unsubscribe(runID, ch)

	if err := r.Submit(RunRequest{
		RunID:     runID,
		AgentName: "agent-b",
		Goal:      "emit events",
		Source:    "test",
	}); err != nil {
		t.Fatalf("submit: %v", err)
	}

	deadline := time.After(2 * time.Second)
	seen := map[observe.EventType]bool{}
	for len(seen) < 2 {
		select {
		case evt := <-ch:
			seen[evt.Type] = true
		case <-deadline:
			t.Fatalf("expected init+completed events, got: %#v", seen)
		}
	}

	if !seen[observe.EventInit] || !seen[observe.EventCompleted] {
		t.Fatalf("missing lifecycle events: %#v", seen)
	}

	persistDeadline := time.Now().Add(2 * time.Second)
	for {
		evts, err := st.GetEvents(context.Background(), runID)
		if err == nil && len(evts) >= 2 {
			break
		}
		if err != nil && !strings.Contains(err.Error(), "SQLITE_BUSY") {
			t.Fatalf("get events: %v", err)
		}
		if time.Now().After(persistDeadline) {
			if err != nil {
				t.Fatalf("timed out waiting for persisted events: %v", err)
			}
			t.Fatal("timed out waiting for persisted lifecycle events")
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func TestRunnerEmitsErrorEvent(t *testing.T) {
	db := t.TempDir() + "/runner-error-events.db"
	st, err := store.NewSQLite(db)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()

	s := NewServer("localhost:0", st, nil)
	r := NewRunner(s, failingStepExecutor{}, 1, NewMemoryQueue(4))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r.Start(ctx)

	runID := "run-events-err-1"
	ch := s.eventHub.Subscribe(runID)
	defer s.eventHub.Unsubscribe(runID, ch)

	if err := r.Submit(RunRequest{
		RunID:     runID,
		AgentName: "agent-c",
		Goal:      "fail",
		Source:    "test",
	}); err != nil {
		t.Fatalf("submit: %v", err)
	}

	deadline := time.After(2 * time.Second)
	seenError := false
	for !seenError {
		select {
		case evt := <-ch:
			if evt.Type == observe.EventError {
				seenError = true
			}
		case <-deadline:
			t.Fatal("expected error event")
		}
	}
}

func TestRunnerRetriesThenSucceeds(t *testing.T) {
	db := t.TempDir() + "/runner-retries.db"
	st, err := store.NewSQLite(db)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()

	exec := &flakyStepExecutor{failures: 2}
	s := NewServer("localhost:0", st, nil)
	r := NewRunner(s, exec, 1, NewMemoryQueue(4))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r.Start(ctx)

	runID := "run-retry-1"
	if err := r.Submit(RunRequest{
		RunID:        runID,
		AgentName:    "agent-retry",
		Goal:         "retry me",
		Source:       "test",
		MaxRetries:   3,
		RetryBackoff: 10 * time.Millisecond,
	}); err != nil {
		t.Fatalf("submit: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		meta, err := st.GetRun(context.Background(), runID)
		if err == nil && meta != nil && meta.State == "completed" {
			exec.mu.Lock()
			attempts := exec.attempts
			exec.mu.Unlock()
			if attempts != 3 {
				t.Fatalf("expected 3 attempts, got %d", attempts)
			}
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("run %s did not complete in time", runID)
}

func TestRunnerPersistsDeadLetterOnTerminalFailure(t *testing.T) {
	db := t.TempDir() + "/runner-deadletters.db"
	st, err := store.NewSQLite(db)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()

	s := NewServer("localhost:0", st, nil)
	r := NewRunner(s, failingStepExecutor{}, 1, NewMemoryQueue(4))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r.Start(ctx)

	runID := "run-deadletter-1"
	if err := r.Submit(RunRequest{
		RunID:        runID,
		AgentName:    "agent-fail",
		Goal:         "always fail",
		Source:       "test",
		MaxRetries:   1,
		RetryBackoff: 10 * time.Millisecond,
		Outputs:      map[string]interface{}{"x": 1},
	}); err != nil {
		t.Fatalf("submit: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		entries, err := st.ListDeadLetters(context.Background(), 10)
		if err != nil && !strings.Contains(err.Error(), "SQLITE_BUSY") {
			t.Fatalf("list dead letters: %v", err)
		}
		if len(entries) > 0 && entries[0].RunID == runID {
			if entries[0].MaxRetries != 1 {
				t.Fatalf("expected max retries 1, got %d", entries[0].MaxRetries)
			}
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatal("expected dead-letter entry for terminal failure")
}

func TestClassifyFailureReason(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "deadline", err: context.DeadlineExceeded, want: "timeout"},
		{name: "canceled", err: context.Canceled, want: "cancelled"},
		{name: "generic", err: errors.New("boom"), want: "execution_error"},
		{name: "nil", err: nil, want: "unknown"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := classifyFailureReason(tc.err); got != tc.want {
				t.Fatalf("classifyFailureReason() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRunnerCancelDuringRetryBackoffInterruptsWithoutDeadLetter(t *testing.T) {
	db := t.TempDir() + "/runner-cancel-backoff.db"
	st, err := store.NewSQLite(db)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()

	s := NewServer("localhost:0", st, nil)
	r := NewRunner(s, failingStepExecutor{}, 1, NewMemoryQueue(4))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r.Start(ctx)

	runID := "run-cancel-backoff-1"
	ch := s.eventHub.Subscribe(runID)
	defer s.eventHub.Unsubscribe(runID, ch)

	if err := r.Submit(RunRequest{
		RunID:        runID,
		AgentName:    "agent-cancel",
		Goal:         "cancel me",
		Source:       "api",
		MaxRetries:   3,
		RetryBackoff: 500 * time.Millisecond,
	}); err != nil {
		t.Fatalf("submit: %v", err)
	}

	// Wait until the first retry event to ensure run is in backoff path.
	retrySeen := false
	retryDeadline := time.After(2 * time.Second)
	for !retrySeen {
		select {
		case evt := <-ch:
			if evt.Type == observe.EventReasoning {
				if typ, _ := evt.Data["type"].(string); typ == "retry" {
					retrySeen = true
				}
			}
		case <-retryDeadline:
			t.Fatal("did not observe retry event before cancellation")
		}
	}

	s.mu.Lock()
	cancelFn := s.runs[runID]
	s.mu.Unlock()
	if cancelFn == nil {
		t.Fatal("missing cancel function for active run")
	}
	cancelFn()

	waitDeadline := time.Now().Add(4 * time.Second)
	lastState := ""
	lastErr := ""
	for time.Now().Before(waitDeadline) {
		meta, err := st.GetRun(context.Background(), runID)
		if err != nil {
			lastErr = err.Error()
			time.Sleep(20 * time.Millisecond)
			continue
		}
		if meta != nil {
			lastState = meta.State
		}
		if meta != nil && meta.State == "interrupted" {
			dls, err := st.ListDeadLetters(context.Background(), 20)
			if err == nil {
				for _, dl := range dls {
					if dl.RunID == runID {
						t.Fatalf("did not expect dead-letter for interrupted run %s", runID)
					}
				}
			}
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("run %s did not transition to interrupted (last_state=%q last_err=%q)", runID, lastState, lastErr)
}

func TestRunnerQueueBackpressureUnderLoad(t *testing.T) {
	db := t.TempDir() + "/runner-backpressure.db"
	st, err := store.NewSQLite(db)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()

	s := NewServer("localhost:0", st, nil)
	r := NewRunner(s, slowStepExecutor{delay: 150 * time.Millisecond}, 1, NewMemoryQueue(2))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r.Start(ctx)

	accepted := 0
	rejected := 0
	for i := 0; i < 20; i++ {
		err := r.Submit(RunRequest{
			RunID:     fmt.Sprintf("bp-%d", i),
			AgentName: "agent-bp",
			Goal:      "load",
			Source:    "test",
		})
		if err != nil {
			rejected++
		} else {
			accepted++
		}
	}
	if accepted == 0 || rejected == 0 {
		t.Fatalf("expected both accepted and rejected submissions; accepted=%d rejected=%d", accepted, rejected)
	}
}
