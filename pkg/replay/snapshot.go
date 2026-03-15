package replay

import (
	"context"
	"encoding/json"
	"time"
)

// RunSnapshot captures a complete execution for replay.
type RunSnapshot struct {
	ID        string    `json:"id"`
	RunID     string    `json:"run_id"`
	Version   string    `json:"version"` // Snapshot format version
	CreatedAt time.Time `json:"created_at"`

	// Agent configuration
	AgentName   string                 `json:"agent_name"`
	Goal        string                 `json:"goal"`
	AgentConfig map[string]interface{} `json:"agent_config"`

	// Model interactions
	ModelCalls []ModelCall `json:"model_calls"`

	// Tool executions
	ToolCalls []ToolExecution `json:"tool_calls"`

	// Environment state
	Environment map[string]string `json:"environment"`
	StartTime   time.Time         `json:"start_time"`
	EndTime     time.Time         `json:"end_time"`
	FinalState  string            `json:"final_state"` // completed, failed, cancelled

	// Verification
	Checksums  map[string]string `json:"checksums"`
	Compressed bool              `json:"compressed"`
	Encrypted  bool              `json:"encrypted"`
	SizeBytes  int64             `json:"size_bytes"`
}

// ModelCall represents a single LLM interaction.
type ModelCall struct {
	Sequence     int                    `json:"sequence"`
	Timestamp    time.Time              `json:"timestamp"`
	Model        string                 `json:"model"`
	Provider     string                 `json:"provider"`
	Prompt       string                 `json:"prompt"`
	Response     string                 `json:"response"`
	TokensUsed   int                    `json:"tokens_used"`
	Temperature  float64                `json:"temperature"`
	MaxTokens    int                    `json:"max_tokens,omitempty"`
	Seed         *int64                 `json:"seed,omitempty"` // For deterministic generation
	FinishReason string                 `json:"finish_reason"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ToolExecution represents a single tool call.
type ToolExecution struct {
	Sequence       int             `json:"sequence"`
	Timestamp      time.Time       `json:"timestamp"`
	ToolName       string          `json:"tool_name"`
	Input          json.RawMessage `json:"input"`
	Output         json.RawMessage `json:"output"`
	Error          string          `json:"error,omitempty"`
	Duration       time.Duration   `json:"duration"`
	RetryCount     int             `json:"retry_count,omitempty"`
	SideEffects    []SideEffect    `json:"side_effects,omitempty"`
	PolicyDecision string          `json:"policy_decision,omitempty"` // allow, deny
}

// SideEffect captures state changes for potential rollback.
type SideEffect struct {
	Type        string          `json:"type"`   // file_write, http_call, db_write, etc.
	Target      string          `json:"target"` // file path, URL, table name
	Description string          `json:"description"`
	Reversible  bool            `json:"reversible"`
	Replayable  bool            `json:"replayable"` // when true, executeSideEffect may re-execute (e.g. idempotent GET)
	RevertData  json.RawMessage `json:"revert_data,omitempty"`
	Timestamp   time.Time       `json:"timestamp"`
}

// ReplayMode defines how a run should be replayed.
type ReplayMode string

const (
	ReplayModeExact      ReplayMode = "exact"      // Use recorded responses
	ReplayModeLive       ReplayMode = "live"       // Re-execute with live APIs
	ReplayModeMixed      ReplayMode = "mixed"      // Recorded models, live tools
	ReplayModeDebug      ReplayMode = "debug"      // Step-through debugging
	ReplayModeValidation ReplayMode = "validation" // Verify consistency
)

// LiveToolExecutor executes tools during live/mixed replay. When nil, live/mixed use recorded output.
type LiveToolExecutor interface {
	Execute(ctx context.Context, toolName, version string, input json.RawMessage) (json.RawMessage, error)
}

// ReplayOptions configures replay behavior.
type ReplayOptions struct {
	Mode               ReplayMode             `json:"mode"`
	SnapshotID         string                 `json:"snapshot_id"`
	StartFromSequence  int                    `json:"start_from_sequence,omitempty"`
	StopAtSequence     int                    `json:"stop_at_sequence,omitempty"`
	Breakpoints        []int                  `json:"breakpoints,omitempty"`
	OverrideConfig     map[string]interface{} `json:"override_config,omitempty"`
	ExecuteSideEffects bool                   `json:"execute_side_effects"`
	VerifyOutputs      bool                   `json:"verify_outputs"`
	// LiveToolExecutor runs tools for live/mixed modes. nil = use recorded output (stub).
	LiveToolExecutor LiveToolExecutor `json:"-"`
}

// ReplayStepTrace records what happened at each step for developer-friendly output.
type ReplayStepTrace struct {
	Seq        int    `json:"seq"`
	Tool       string `json:"tool"`
	Source     string `json:"source"` // "recorded" | "live" | "breakpoint"
	InputSum   string `json:"input_summary"`
	OutputSum  string `json:"output_summary"`
	Result     string `json:"result"` // "ok" | "match" | "diverged" | ""
	Duration   string `json:"duration,omitempty"`
	Divergence string `json:"divergence,omitempty"` // when Result=diverged
}

// ReplayResult contains the outcome of a replay.
type ReplayResult struct {
	Success     bool          `json:"success"`
	SnapshotID  string        `json:"snapshot_id"`
	Mode        ReplayMode    `json:"mode"`
	StartedAt   time.Time     `json:"started_at"`
	CompletedAt time.Time     `json:"completed_at"`
	Duration    time.Duration `json:"duration"`

	// Step-by-step trace for developer output
	Trace []ReplayStepTrace `json:"trace,omitempty"`

	// Execution results
	ActionsRerun int          `json:"actions_rerun"`
	Matches      int          `json:"matches"`
	Divergences  []Divergence `json:"divergences"`

	// Final state
	FinalState map[string]interface{} `json:"final_state"`
	Output     json.RawMessage        `json:"output,omitempty"`
	Error      string                 `json:"error,omitempty"`

	// Metrics
	Metrics ReplayMetrics `json:"metrics"`
}

// Divergence represents a difference from the original execution.
type Divergence struct {
	Sequence    int         `json:"sequence"`
	Type        string      `json:"type"`      // output, decision, error, timing
	Component   string      `json:"component"` // model, tool, policy
	Expected    interface{} `json:"expected"`
	Actual      interface{} `json:"actual"`
	Impact      string      `json:"impact"` // minor, major, critical
	Description string      `json:"description"`
	Timestamp   time.Time   `json:"timestamp"`
}

// ReplayMetrics contains replay performance metrics.
type ReplayMetrics struct {
	TotalActions     int           `json:"total_actions"`
	ActionsMatched   int           `json:"actions_matched"`
	ActionsDiverged  int           `json:"actions_diverged"`
	Accuracy         float64       `json:"accuracy"` // matches / total
	AverageLatency   time.Duration `json:"average_latency"`
	OriginalDuration time.Duration `json:"original_duration"`
	ReplayDuration   time.Duration `json:"replay_duration"`
	SpeedupFactor    float64       `json:"speedup_factor"` // original / replay
}

// RecordingConfig configures what gets recorded.
type RecordingConfig struct {
	Enabled            bool   `json:"enabled"`
	CaptureModel       bool   `json:"capture_model"`
	CaptureTools       bool   `json:"capture_tools"`
	CaptureSideEffects bool   `json:"capture_side_effects"`
	CaptureEnvironment bool   `json:"capture_environment"`
	CompressData       bool   `json:"compress_data"`
	EncryptPII         bool   `json:"encrypt_pii"`
	MaxSnapshotSize    int64  `json:"max_snapshot_size"`
	StoragePath        string `json:"storage_path"`
}

// SnapshotMetadata holds minimal snapshot info for listings.
type SnapshotMetadata struct {
	ID         string        `json:"id"`
	RunID      string        `json:"run_id"`
	AgentName  string        `json:"agent_name"`
	Goal       string        `json:"goal"`
	CreatedAt  time.Time     `json:"created_at"`
	Duration   time.Duration `json:"duration"`
	FinalState string        `json:"final_state"`
	ModelCalls int           `json:"model_calls"`
	ToolCalls  int           `json:"tool_calls"`
	SizeBytes  int64         `json:"size_bytes"`
	Compressed bool          `json:"compressed"`
}

// FreezePoint represents a point where execution was paused.
type FreezePoint struct {
	ID            string                 `json:"id"`
	RunID         string                 `json:"run_id"`
	SnapshotID    string                 `json:"snapshot_id,omitempty"`
	Sequence      int                    `json:"sequence"`
	Timestamp     time.Time              `json:"timestamp"`
	Reason        string                 `json:"reason"`
	State         map[string]interface{} `json:"state"`
	PendingAction ToolExecution          `json:"pending_action"`
	Options       []FreezeOption         `json:"options"`
	Decision      *FreezeDecision        `json:"decision,omitempty"`
	Resolved      bool                   `json:"resolved"`
	ResolvedAt    *time.Time             `json:"resolved_at,omitempty"`
}

// FreezeOption represents a choice at a freeze point.
type FreezeOption struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
	RiskImpact  string `json:"risk_impact"` // low, medium, high
	Recommended bool   `json:"recommended"`
}

// FreezeDecision captures the resolution of a freeze point.
type FreezeDecision struct {
	Option     string                 `json:"option"`
	ApprovedBy string                 `json:"approved_by"`
	ApprovedAt time.Time              `json:"approved_at"`
	Reason     string                 `json:"reason"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// SnapshotVersion defines the snapshot format version.
const SnapshotVersion = "1.0.0"

// DivergenceImpact levels
const (
	ImpactMinor    = "minor"    // Negligible difference (timing, whitespace)
	ImpactMajor    = "major"    // Significant difference (output changed)
	ImpactCritical = "critical" // Breaking difference (error, crash)
)

// SideEffectType constants
const (
	SideEffectFileWrite   = "file_write"
	SideEffectFileDelete  = "file_delete"
	SideEffectHTTPCall    = "http_call"
	SideEffectDBWrite     = "db_write"
	SideEffectDBDelete    = "db_delete"
	SideEffectExternalAPI = "external_api"
)
