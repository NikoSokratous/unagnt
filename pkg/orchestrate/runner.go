package orchestrate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/NikoSokratous/unagnt/internal/store"
	"github.com/NikoSokratous/unagnt/pkg/observe"
)

// RunRequest defines one asynchronous execution request handled by Runner.
type RunRequest struct {
	RunID        string
	AgentName    string
	Goal         string
	Source       string
	Outputs      map[string]interface{}
	MaxRetries   int
	RetryBackoff time.Duration
	Timeout      time.Duration
	CallbackFn   func(runID string, status string, output interface{}, err error)
}

// Runner executes queued runs with a step executor.
type Runner struct {
	server   *Server
	exec     StepExecutor
	workers  int
	queue    chan RunRequest
	stopOnce sync.Once
}

func NewRunner(server *Server, exec StepExecutor, workers int, queueSize int) *Runner {
	if workers <= 0 {
		workers = 1
	}
	if queueSize <= 0 {
		queueSize = 256
	}
	if exec == nil {
		exec = SimulatedExecutor{}
	}
	return &Runner{
		server:  server,
		exec:    exec,
		workers: workers,
		queue:   make(chan RunRequest, queueSize),
	}
}

func (r *Runner) Start(ctx context.Context) {
	for i := 0; i < r.workers; i++ {
		go r.worker(ctx)
	}
}

func (r *Runner) Stop() {
	r.stopOnce.Do(func() {
		close(r.queue)
	})
}

func (r *Runner) Submit(req RunRequest) error {
	select {
	case r.queue <- req:
		RunQueueDepth.Set(float64(len(r.queue)))
		return nil
	default:
		RunQueueRejected.Inc()
		return fmt.Errorf("runner queue is full")
	}
}

func (r *Runner) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case req, ok := <-r.queue:
			if !ok {
				return
			}
			RunQueueDepth.Set(float64(len(r.queue)))
			r.execute(ctx, req)
		}
	}
}

func (r *Runner) execute(parent context.Context, req RunRequest) {
	runCtx, cancel := context.WithCancel(parent)
	r.server.setRunCancel(req.RunID, cancel)
	defer func() {
		cancel()
		r.server.clearRunCancel(req.RunID)
	}()
	RunsActive.Inc()
	defer RunsActive.Dec()
	persistCtx, persistCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer persistCancel()
	startedAt := time.Now()

	now := time.Now()
	meta := &store.RunMeta{
		RunID:     req.RunID,
		AgentName: req.AgentName,
		Goal:      req.Goal,
		State:     "running",
		CreatedAt: now,
		UpdatedAt: now,
	}
	_ = r.saveRunWithRetry(runCtx, meta)
	r.emitEvent(runCtx, req.RunID, req.AgentName, observe.EventInit, map[string]any{
		"source": req.Source,
		"goal":   req.Goal,
	})
	maxAttempts := req.MaxRetries + 1
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	backoff := req.RetryBackoff
	if backoff <= 0 {
		backoff = 500 * time.Millisecond
	}

	var finalErr error
	var finalOutput interface{}
	finalAttempt := 0
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		finalAttempt = attempt
		attemptCtx := runCtx
		attemptCancel := func() {}
		if req.Timeout > 0 {
			attemptCtx, attemptCancel = context.WithTimeout(runCtx, req.Timeout)
		}
		stepResult, err := r.exec.ExecuteStep(attemptCtx, req.AgentName, req.Goal, req.Outputs)
		attemptCancel()
		if err == nil {
			finalOutput = stepResult.Output
			meta.State = "completed"
			meta.UpdatedAt = time.Now()
			_ = r.saveRunWithRetry(persistCtx, meta)
			RunDuration.Observe(time.Since(startedAt).Seconds())
			r.emitEvent(persistCtx, req.RunID, req.AgentName, observe.EventCompleted, map[string]any{
				"status":      "completed",
				"durationMs":  stepResult.Duration.Milliseconds(),
				"attempt":     attempt,
				"maxAttempts": maxAttempts,
			})
			if req.CallbackFn != nil {
				req.CallbackFn(req.RunID, meta.State, finalOutput, nil)
			}
			return
		}

		finalErr = err
		interrupted := errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
		retryable := !interrupted
		if attempt < maxAttempts && retryable {
			RunRetries.Inc()
			r.emitEvent(runCtx, req.RunID, req.AgentName, observe.EventReasoning, map[string]any{
				"type":        "retry",
				"attempt":     attempt,
				"maxAttempts": maxAttempts,
				"error":       err.Error(),
			})
			wait := time.Duration(attempt) * backoff
			timer := time.NewTimer(wait)
			aborted := false
			select {
			case <-runCtx.Done():
				timer.Stop()
				finalErr = runCtx.Err()
				aborted = true
			case <-timer.C:
			}
			if aborted {
				break
			}
			continue
		}
		break
	}

	eventType := observe.EventError
	meta.State = "failed"
	if runErr := runCtx.Err(); runErr != nil {
		finalErr = runErr
	}
	if errors.Is(finalErr, context.Canceled) || errors.Is(finalErr, context.DeadlineExceeded) {
		eventType = observe.EventInterrupted
		meta.State = "interrupted"
	}
	RunFailures.WithLabelValues(classifyFailureReason(finalErr), req.Source).Inc()
	meta.UpdatedAt = time.Now()
	_ = r.saveRunWithRetry(persistCtx, meta)
	r.emitEvent(persistCtx, req.RunID, req.AgentName, eventType, map[string]any{
		"error":       finalErr.Error(),
		"attempt":     finalAttempt,
		"maxAttempts": maxAttempts,
	})
	if meta.State == "failed" {
		r.persistDeadLetter(persistCtx, req, finalErr, finalAttempt, maxAttempts)
	}
	if req.CallbackFn != nil {
		req.CallbackFn(req.RunID, meta.State, nil, finalErr)
	}
}

func classifyFailureReason(err error) string {
	switch {
	case err == nil:
		return "unknown"
	case errors.Is(err, context.DeadlineExceeded):
		return "timeout"
	case errors.Is(err, context.Canceled):
		return "cancelled"
	default:
		return "execution_error"
	}
}

func (r *Runner) emitEvent(ctx context.Context, runID, agent string, eventType observe.EventType, data map[string]any) {
	if r == nil || r.server == nil {
		return
	}
	evt := observe.Event{
		RunID:     runID,
		Timestamp: time.Now(),
		Type:      eventType,
		Agent:     agent,
		Data:      data,
	}
	if r.server.eventHub != nil {
		r.server.eventHub.Publish(runID, evt)
	}
	if r.server.store != nil {
		_ = r.saveEventWithRetry(ctx, runID, &evt)
	}
}

func (r *Runner) persistDeadLetter(ctx context.Context, req RunRequest, err error, attempt, maxAttempts int) {
	if r == nil || r.server == nil || r.server.store == nil || err == nil {
		return
	}
	payload := ""
	if len(req.Outputs) > 0 {
		if b, mErr := json.Marshal(req.Outputs); mErr == nil {
			payload = string(b)
		}
	}
	_ = r.saveDeadLetterWithRetry(ctx, &store.DeadLetter{
		RunID:      req.RunID,
		AgentName:  req.AgentName,
		Goal:       req.Goal,
		Source:     req.Source,
		Error:      err.Error(),
		Payload:    payload,
		Attempt:    attempt,
		MaxRetries: maxAttempts - 1,
		FailedAt:   time.Now(),
	})
	RunDeadLetters.Inc()
}

func (r *Runner) saveRunWithRetry(ctx context.Context, meta *store.RunMeta) error {
	if r == nil || r.server == nil || r.server.store == nil {
		return nil
	}
	return withStoreRetry(func() error { return r.server.store.SaveRun(ctx, meta) })
}

func (r *Runner) saveEventWithRetry(ctx context.Context, runID string, evt *observe.Event) error {
	if r == nil || r.server == nil || r.server.store == nil {
		return nil
	}
	return withStoreRetry(func() error { return r.server.store.SaveEvent(ctx, runID, evt) })
}

func (r *Runner) saveDeadLetterWithRetry(ctx context.Context, dl *store.DeadLetter) error {
	if r == nil || r.server == nil || r.server.store == nil {
		return nil
	}
	return withStoreRetry(func() error { return r.server.store.SaveDeadLetter(ctx, dl) })
}

func withStoreRetry(fn func() error) error {
	const maxAttempts = 4
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := fn(); err != nil {
			lastErr = err
			msg := strings.ToLower(err.Error())
			if strings.Contains(msg, "sqlite_busy") || strings.Contains(msg, "database is locked") {
				time.Sleep(time.Duration(attempt) * 20 * time.Millisecond)
				continue
			}
			return err
		}
		return nil
	}
	return lastErr
}
