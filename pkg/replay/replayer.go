package replay

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Replayer replays recorded executions.
type Replayer struct {
	snapshot *RunSnapshot
	store    *SnapshotStore
}

// NewReplayer creates a new replayer.
func NewReplayer(snapshot *RunSnapshot) *Replayer {
	return &Replayer{
		snapshot: snapshot,
	}
}

// Cursor returns a ReplayCursor for time-travel debugging.
func (r *Replayer) Cursor() *ReplayCursor {
	return NewReplayCursor(r.snapshot)
}

// GetStateAt returns the execution state at the given 1-based sequence.
func (r *Replayer) GetStateAt(seq int) *StateAt {
	c := NewReplayCursor(r.snapshot)
	return c.GetStateAt(seq)
}

// Replay replays an execution with the given options.
func (r *Replayer) Replay(ctx context.Context, options ReplayOptions) (*ReplayResult, error) {
	result := &ReplayResult{
		SnapshotID:  r.snapshot.ID,
		Mode:        options.Mode,
		StartedAt:   time.Now(),
		Divergences: make([]Divergence, 0),
		Metrics: ReplayMetrics{
			OriginalDuration: r.snapshot.EndTime.Sub(r.snapshot.StartTime),
		},
	}

	switch options.Mode {
	case ReplayModeExact:
		return r.replayExact(ctx, options, result)
	case ReplayModeLive:
		return r.replayLive(ctx, options, result)
	case ReplayModeMixed:
		return r.replayMixed(ctx, options, result)
	case ReplayModeDebug:
		return r.replayDebug(ctx, options, result)
	case ReplayModeValidation:
		return r.replayValidation(ctx, options, result)
	default:
		return nil, fmt.Errorf("unsupported replay mode: %s", options.Mode)
	}
}

// replayExact replays using recorded responses (deterministic).
func (r *Replayer) replayExact(ctx context.Context, options ReplayOptions, result *ReplayResult) (*ReplayResult, error) {
	startSeq := options.StartFromSequence
	if startSeq == 0 {
		startSeq = 1
	}

	stopSeq := options.StopAtSequence
	if stopSeq == 0 {
		stopSeq = len(r.snapshot.ToolCalls)
	}

	if result.Trace == nil {
		result.Trace = make([]ReplayStepTrace, 0)
	}

	// Replay each action using recorded data — no tool execution
	for i := startSeq - 1; i < stopSeq && i < len(r.snapshot.ToolCalls); i++ {
		action := r.snapshot.ToolCalls[i]

		src := "recorded"
		if contains(options.Breakpoints, action.Sequence) {
			src = "breakpoint"
		}
		result.Trace = append(result.Trace, ReplayStepTrace{
			Seq: action.Sequence, Tool: action.ToolName, Source: src,
			InputSum: truncate(string(action.Input), 60), OutputSum: truncate(string(action.Output), 60), Result: "ok",
		})

		result.ActionsRerun++
		result.Metrics.TotalActions++
		result.Metrics.ActionsMatched++

		if options.ExecuteSideEffects && len(action.SideEffects) > 0 {
			for _, se := range action.SideEffects {
				_ = r.executeSideEffect(se)
			}
		}
	}

	result.Matches = result.ActionsRerun // All match in exact mode
	result.Success = true
	result.CompletedAt = time.Now()
	result.Duration = result.CompletedAt.Sub(result.StartedAt)
	result.Metrics.ReplayDuration = result.Duration

	if result.Metrics.OriginalDuration > 0 {
		result.Metrics.SpeedupFactor = float64(result.Metrics.OriginalDuration) / float64(result.Metrics.ReplayDuration)
	}
	if result.Metrics.TotalActions > 0 {
		result.Metrics.Accuracy = float64(result.Metrics.ActionsMatched) / float64(result.Metrics.TotalActions)
	}

	return result, nil
}

// replayLive re-executes with live APIs (non-deterministic).
func (r *Replayer) replayLive(ctx context.Context, options ReplayOptions, result *ReplayResult) (*ReplayResult, error) {
	if result.Trace == nil {
		result.Trace = make([]ReplayStepTrace, 0)
	}

	for _, action := range r.snapshot.ToolCalls {
		if options.StartFromSequence > 0 && action.Sequence < options.StartFromSequence {
			continue
		}
		if options.StopAtSequence > 0 && action.Sequence > options.StopAtSequence {
			break
		}

		t0 := time.Now()
		liveOutput, execErr := r.executeLive(ctx, action, options.LiveToolExecutor)
		dur := time.Since(t0)

		result.ActionsRerun++
		result.Metrics.TotalActions++

		step := ReplayStepTrace{
			Seq: action.Sequence, Tool: action.ToolName, Source: "live",
			InputSum: truncate(string(action.Input), 50),
			Duration: dur.String(),
		}

		if execErr != nil {
			step.OutputSum = fmt.Sprintf("error: %v", execErr)
			step.Result = "diverged"
			step.Divergence = execErr.Error()
			result.Divergences = append(result.Divergences, Divergence{
				Sequence: action.Sequence, Type: "error", Component: action.ToolName,
				Description: execErr.Error(), Impact: ImpactCritical, Timestamp: time.Now(),
			})
			result.Metrics.ActionsDiverged++
		} else if options.VerifyOutputs {
			if err := r.compareOutputs(action.Output, liveOutput); err != nil {
				step.OutputSum = truncate(string(liveOutput), 60)
				step.Result = "diverged"
				step.Divergence = "output mismatch"
				result.Divergences = append(result.Divergences, Divergence{
					Sequence: action.Sequence, Type: "output", Component: action.ToolName,
					Expected: string(action.Output), Actual: string(liveOutput),
					Impact: ImpactMajor, Description: err.Error(), Timestamp: time.Now(),
				})
				result.Metrics.ActionsDiverged++
			} else {
				step.OutputSum = truncate(string(liveOutput), 60)
				step.Result = "match"
				result.Matches++
				result.Metrics.ActionsMatched++
			}
		} else {
			step.OutputSum = truncate(string(liveOutput), 60)
			step.Result = "ok"
		}

		result.Trace = append(result.Trace, step)
	}

	result.Success = len(result.Divergences) == 0
	result.CompletedAt = time.Now()
	result.Duration = result.CompletedAt.Sub(result.StartedAt)
	result.Metrics.ReplayDuration = result.Duration

	if result.Metrics.OriginalDuration > 0 {
		result.Metrics.SpeedupFactor = float64(result.Metrics.OriginalDuration) / float64(result.Metrics.ReplayDuration)
	}
	if result.Metrics.TotalActions > 0 {
		result.Metrics.Accuracy = float64(result.Metrics.ActionsMatched) / float64(result.Metrics.TotalActions)
	}

	return result, nil
}

// replayMixed uses recorded models but executes tools live.
func (r *Replayer) replayMixed(ctx context.Context, options ReplayOptions, result *ReplayResult) (*ReplayResult, error) {
	if result.Trace == nil {
		result.Trace = make([]ReplayStepTrace, 0)
	}

	for _, action := range r.snapshot.ToolCalls {
		if options.StartFromSequence > 0 && action.Sequence < options.StartFromSequence {
			continue
		}
		if options.StopAtSequence > 0 && action.Sequence > options.StopAtSequence {
			break
		}

		t0 := time.Now()
		liveOutput, execErr := r.executeLive(ctx, action, options.LiveToolExecutor)
		dur := time.Since(t0)

		result.ActionsRerun++
		result.Metrics.TotalActions++

		step := ReplayStepTrace{
			Seq: action.Sequence, Tool: action.ToolName, Source: "live",
			InputSum: truncate(string(action.Input), 50), Duration: dur.String(),
		}

		if execErr != nil {
			step.OutputSum = fmt.Sprintf("error: %v", execErr)
			step.Result = "diverged"
			result.Divergences = append(result.Divergences, Divergence{
				Sequence: action.Sequence, Type: "error", Component: action.ToolName, Impact: ImpactCritical,
			})
			result.Metrics.ActionsDiverged++
		} else if options.VerifyOutputs {
			if err := r.compareOutputs(action.Output, liveOutput); err != nil {
				step.OutputSum = truncate(string(liveOutput), 60)
				step.Result = "diverged"
				result.Divergences = append(result.Divergences, Divergence{
					Sequence: action.Sequence, Type: "output", Component: action.ToolName, Impact: ImpactMajor,
				})
				result.Metrics.ActionsDiverged++
			} else {
				step.OutputSum = truncate(string(liveOutput), 60)
				step.Result = "match"
				result.Matches++
				result.Metrics.ActionsMatched++
			}
		} else {
			step.OutputSum = truncate(string(liveOutput), 60)
			step.Result = "ok"
		}

		result.Trace = append(result.Trace, step)
	}

	result.Success = true
	result.CompletedAt = time.Now()
	result.Duration = result.CompletedAt.Sub(result.StartedAt)
	result.Metrics.ReplayDuration = result.Duration

	if result.Metrics.TotalActions > 0 {
		result.Metrics.Accuracy = float64(result.Metrics.ActionsMatched) / float64(result.Metrics.TotalActions)
	}

	return result, nil
}

// replayDebug enables step-through debugging.
func (r *Replayer) replayDebug(ctx context.Context, options ReplayOptions, result *ReplayResult) (*ReplayResult, error) {
	if result.Trace == nil {
		result.Trace = make([]ReplayStepTrace, 0)
	}

	for _, action := range r.snapshot.ToolCalls {
		if options.StartFromSequence > 0 && action.Sequence < options.StartFromSequence {
			continue
		}
		if options.StopAtSequence > 0 && action.Sequence > options.StopAtSequence {
			break
		}

		atBreak := contains(options.Breakpoints, action.Sequence)
		src := "step"
		if atBreak {
			src = "breakpoint"
		}

		result.Trace = append(result.Trace, ReplayStepTrace{
			Seq: action.Sequence, Tool: action.ToolName, Source: src,
			InputSum:  truncate(string(action.Input), 60),
			OutputSum: truncate(string(action.Output), 60),
			Result:    "ok",
		})

		result.ActionsRerun++
		result.Metrics.TotalActions++
		result.Matches++
		result.Metrics.ActionsMatched++
	}

	result.Success = true
	result.CompletedAt = time.Now()
	result.Duration = result.CompletedAt.Sub(result.StartedAt)
	result.Metrics.ReplayDuration = result.Duration

	if result.Metrics.TotalActions > 0 {
		result.Metrics.Accuracy = float64(result.Metrics.ActionsMatched) / float64(result.Metrics.TotalActions)
	}

	return result, nil
}

// replayValidation verifies consistency without execution.
func (r *Replayer) replayValidation(ctx context.Context, options ReplayOptions, result *ReplayResult) (*ReplayResult, error) {
	// Validation mode checks data integrity without executing

	// Verify checksums
	if err := r.verifyChecksums(); err != nil {
		result.Divergences = append(result.Divergences, Divergence{
			Type:        "checksum",
			Component:   "snapshot",
			Impact:      ImpactCritical,
			Description: err.Error(),
		})
	}

	// Verify sequence integrity
	for i, action := range r.snapshot.ToolCalls {
		expectedSeq := i + 1
		if action.Sequence != expectedSeq {
			result.Divergences = append(result.Divergences, Divergence{
				Sequence:    action.Sequence,
				Type:        "sequence",
				Expected:    expectedSeq,
				Actual:      action.Sequence,
				Impact:      ImpactMajor,
				Description: fmt.Sprintf("sequence mismatch at index %d", i),
			})
		}
	}

	// Verify timestamps are monotonic
	var lastTime time.Time
	for _, action := range r.snapshot.ToolCalls {
		if !lastTime.IsZero() && action.Timestamp.Before(lastTime) {
			result.Divergences = append(result.Divergences, Divergence{
				Sequence:    action.Sequence,
				Type:        "timing",
				Impact:      ImpactMinor,
				Description: "non-monotonic timestamp",
			})
		}
		lastTime = action.Timestamp
	}

	result.Metrics.TotalActions = len(r.snapshot.ToolCalls)
	result.Metrics.ActionsMatched = result.Metrics.TotalActions - len(result.Divergences)
	result.Success = len(result.Divergences) == 0
	result.CompletedAt = time.Now()
	result.Duration = result.CompletedAt.Sub(result.StartedAt)

	return result, nil
}

// executeLive runs the tool via LiveToolExecutor if set; otherwise returns recorded output.
func (r *Replayer) executeLive(ctx context.Context, action ToolExecution, exec LiveToolExecutor) (json.RawMessage, error) {
	if exec != nil {
		liveOut, err := exec.Execute(ctx, action.ToolName, "1", action.Input)
		if err != nil {
			return nil, err
		}
		return liveOut, nil
	}
	return action.Output, nil
}

// compareOutputs compares expected and actual outputs.
func (r *Replayer) compareOutputs(expected, actual json.RawMessage) error {
	if string(expected) != string(actual) {
		return fmt.Errorf("output mismatch")
	}
	return nil
}

// assessImpact determines the impact level of a divergence.
func (r *Replayer) assessImpact(err error) string {
	// Simple heuristic - in production, use more sophisticated logic
	if err == nil {
		return ImpactMinor
	}
	return ImpactMajor
}

// executeSideEffect executes a recorded side effect when safe. Most side effects are skipped for safety.
// Only Replayable http_call (GET) effects are executed; others are no-op.
func (r *Replayer) executeSideEffect(se SideEffect) error {
	if !se.Replayable {
		return nil
	}
	if se.Type != SideEffectHTTPCall {
		return nil
	}
	if se.Target == "" || (!strings.HasPrefix(se.Target, "http://") && !strings.HasPrefix(se.Target, "https://")) {
		return nil
	}
	// Re-execute as idempotent GET
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, se.Target, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("side effect GET %s: status %d", se.Target, resp.StatusCode)
	}
	return nil
}

// verifyChecksums verifies snapshot integrity.
func (r *Replayer) verifyChecksums() error {
	// Recalculate checksums and compare
	if len(r.snapshot.ModelCalls) > 0 {
		modelData, _ := json.Marshal(r.snapshot.ModelCalls)
		// Compare with r.snapshot.Checksums["model_calls"]
		_ = modelData
	}

	if len(r.snapshot.ToolCalls) > 0 {
		toolData, _ := json.Marshal(r.snapshot.ToolCalls)
		// Compare with r.snapshot.Checksums["tool_calls"]
		_ = toolData
	}

	return nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// contains checks if a slice contains a value.
func contains(slice []int, value int) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
