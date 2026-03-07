package replay

import "context"

// StateAt represents the execution state at a given sequence position.
type StateAt struct {
	Position      int             // Current sequence (1-based)
	ToolCalls     []ToolExecution // Tool calls up to and including Position
	ModelCalls    []ModelCall     // Model calls up to and including Position
	CurrentAction *ToolExecution  // The action at Position (if any)
	CanStepForward bool
	CanStepBack   bool
}

// ReplayCursor allows stepping forward and backward through a snapshot.
type ReplayCursor struct {
	snapshot *RunSnapshot
	position int // 1-based; 0 = before first action
}

// NewReplayCursor creates a cursor for time-travel debugging.
func NewReplayCursor(snapshot *RunSnapshot) *ReplayCursor {
	return &ReplayCursor{
		snapshot: snapshot,
		position: 0,
	}
}

// Position returns the current 1-based sequence position (0 = before first).
func (c *ReplayCursor) Position() int {
	return c.position
}

// CanStepForward returns true if there is a next action.
func (c *ReplayCursor) CanStepForward() bool {
	return c.position < len(c.snapshot.ToolCalls)
}

// CanStepBack returns true if we can step backward.
func (c *ReplayCursor) CanStepBack() bool {
	return c.position > 0
}

// StepForward advances the cursor by one action.
func (c *ReplayCursor) StepForward() bool {
	if !c.CanStepForward() {
		return false
	}
	c.position++

	return true
}

// StepBack moves the cursor back by one action.
func (c *ReplayCursor) StepBack() bool {
	if !c.CanStepBack() {
		return false
	}
	c.position--

	return true
}

// SeekToSequence sets the cursor to the given 1-based sequence.
func (c *ReplayCursor) SeekToSequence(seq int) {
	if seq < 0 {
		seq = 0
	}
	if seq > len(c.snapshot.ToolCalls) {
		seq = len(c.snapshot.ToolCalls)
	}
	c.position = seq
}

// GetStateAt returns the execution state at the current position.
func (c *ReplayCursor) GetStateAt(seq int) *StateAt {
	if seq < 0 {
		seq = 0
	}
	if seq > len(c.snapshot.ToolCalls) {
		seq = len(c.snapshot.ToolCalls)
	}

	toolCalls := make([]ToolExecution, 0, seq)
	for i := 0; i < seq && i < len(c.snapshot.ToolCalls); i++ {
		toolCalls = append(toolCalls, c.snapshot.ToolCalls[i])
	}

	modelCalls := make([]ModelCall, 0)
	for _, mc := range c.snapshot.ModelCalls {
		if mc.Sequence <= seq {
			modelCalls = append(modelCalls, mc)
		}
	}

	var currentAction *ToolExecution
	if seq >= 1 && seq <= len(c.snapshot.ToolCalls) {
		ac := c.snapshot.ToolCalls[seq-1]
		currentAction = &ac
	}

	return &StateAt{
		Position:       seq,
		ToolCalls:      toolCalls,
		ModelCalls:     modelCalls,
		CurrentAction:  currentAction,
		CanStepForward: seq < len(c.snapshot.ToolCalls),
		CanStepBack:    seq > 0,
	}
}

// ReplayRange replays only the slice of actions from `from` to `to` (1-based, inclusive).
// It returns a ReplayResult for that range.
func (r *Replayer) ReplayRange(ctx context.Context, from, to int, options ReplayOptions) (*ReplayResult, error) {
	opts := options
	opts.StartFromSequence = from
	opts.StopAtSequence = to

	return r.Replay(ctx, opts)
}
