package context

import (
	"context"
	"testing"
	"time"

	"github.com/NikoSokratous/unagnt/pkg/llm"
	"github.com/NikoSokratous/unagnt/pkg/runtime"
)

func TestNewEngine(t *testing.T) {
	assembler := NewDefaultAssembler()
	config := EngineConfig{
		MaxTokens:     8000,
		EnableCache:   true,
		CacheDuration: 30 * time.Second,
	}

	engine := NewEngine(config, assembler)

	if engine == nil {
		t.Fatal("Expected engine to be created")
	}

	if engine.Config.MaxTokens != 8000 {
		t.Errorf("Expected MaxTokens to be 8000, got %d", engine.Config.MaxTokens)
	}
}

func TestEngineAddProvider(t *testing.T) {
	engine := NewEngine(EngineConfig{}, NewDefaultAssembler())

	provider := NewWorkflowProvider(2)
	engine.AddProvider(provider)

	if len(engine.Providers) != 1 {
		t.Errorf("Expected 1 provider, got %d", len(engine.Providers))
	}
}

func TestDefaultAssembler(t *testing.T) {
	assembler := NewDefaultAssembler()

	fragments := []ContextFragment{
		{
			ProviderName: "test",
			Type:         FragmentTypeSystemPrompt,
			Content:      "Test system prompt",
			Priority:     1,
			TokenCount:   50,
		},
		{
			ProviderName: "test",
			Type:         FragmentTypeUserGoal,
			Content:      "Test goal",
			Priority:     2,
			TokenCount:   20,
		},
	}

	config := AssemblyConfig{
		MaxTokens: 8000,
	}

	messages, err := assembler.Assemble(fragments, config)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}

	if len(messages) == 0 {
		t.Error("Expected messages to be generated")
	}
}

func TestWorkflowProvider(t *testing.T) {
	provider := NewWorkflowProvider(2)

	if provider.Name() != "workflow_state" {
		t.Errorf("Expected name 'workflow_state', got '%s'", provider.Name())
	}

	if provider.Priority() != 2 {
		t.Errorf("Expected priority 2, got %d", provider.Priority())
	}

	ctx := context.Background()
	input := ContextInput{
		StepInput: runtime.StepInput{
			AgentName: "test-agent",
			Goal:      "Test goal",
			StepNum:   5,
			History:   []runtime.StepRecord{},
		},
	}

	fragment, err := provider.Fetch(ctx, input)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if fragment == nil {
		t.Fatal("Expected fragment to be returned")
	}

	if fragment.Type != FragmentTypeWorkflowState {
		t.Errorf("Expected type FragmentTypeWorkflowState, got %s", fragment.Type)
	}
}

func TestToolOutputProvider(t *testing.T) {
	provider := NewToolOutputProvider(4, 10)

	if provider.Name() != "tool_outputs" {
		t.Errorf("Expected name 'tool_outputs', got '%s'", provider.Name())
	}

	ctx := context.Background()
	history := []runtime.StepRecord{
		{
			Action: &runtime.ToolCall{
				Tool:    "test_tool",
				Version: "1",
				Input:   map[string]any{"key": "value"},
			},
			Result: &runtime.ToolResult{
				Output: map[string]any{"result": "success"},
			},
		},
	}

	input := ContextInput{
		StepInput: runtime.StepInput{
			History: history,
		},
	}

	fragment, err := provider.Fetch(ctx, input)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if fragment == nil {
		t.Fatal("Expected fragment to be returned")
	}

	if fragment.Type != FragmentTypeToolOutput {
		t.Errorf("Expected type FragmentTypeToolOutput, got %s", fragment.Type)
	}

	if fragment.Content == "" {
		t.Error("Expected content to be non-empty")
	}
}

func TestPolicyProvider(t *testing.T) {
	provider := NewPolicyProvider(nil, 1)

	if provider.Name() != "policy" {
		t.Errorf("Expected name 'policy', got '%s'", provider.Name())
	}

	ctx := context.Background()
	input := ContextInput{
		Tools: []llm.ToolInfo{
			{Name: "file_read", Version: "1", Description: "Read files"},
			{Name: "file_write", Version: "1", Description: "Write files"},
		},
	}

	fragment, err := provider.Fetch(ctx, input)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	// When engine is nil, it still returns a fragment with generic policies
	if fragment == nil {
		t.Fatal("Expected fragment to be returned even with nil engine")
	}

	if fragment.Type != FragmentTypePolicy {
		t.Errorf("Expected type FragmentTypePolicy, got %s", fragment.Type)
	}

	if fragment.Content == "" {
		t.Error("Expected policy content to be non-empty")
	}
}

func TestKnowledgeProvider(t *testing.T) {
	provider := NewKnowledgeProvider(5, []string{"docs", "kb"})

	if provider.Name() != "knowledge" {
		t.Errorf("Expected name 'knowledge', got '%s'", provider.Name())
	}

	// Should be disabled by default
	if provider.Enabled {
		t.Error("Expected provider to be disabled by default")
	}

	ctx := context.Background()
	input := ContextInput{}

	fragment, err := provider.Fetch(ctx, input)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	// Should return nil when disabled
	if fragment != nil && fragment.Content != "" {
		t.Error("Expected no content when provider is disabled")
	}
}

func TestTokenCounter(t *testing.T) {
	counter := &SimpleTokenCounter{}

	text := "This is a test sentence with multiple words."
	tokens := counter.Count(text)

	// Simple heuristic: 1 token ≈ 4 characters
	expected := len(text) / 4
	if tokens != expected {
		t.Errorf("Expected ~%d tokens, got %d", expected, tokens)
	}
}

func TestEngineFetchAllSequential(t *testing.T) {
	config := EngineConfig{
		MaxTokens: 8000,
		Parallel:  false,
	}
	engine := NewEngine(config, NewDefaultAssembler())

	engine.AddProvider(NewWorkflowProvider(1))
	engine.AddProvider(NewToolOutputProvider(2, 10))

	ctx := context.Background()
	input := ContextInput{
		StepInput: runtime.StepInput{
			AgentName: "test",
			Goal:      "Test goal",
		},
	}

	fragments, err := engine.FetchAll(ctx, input)
	if err != nil {
		t.Fatalf("FetchAll failed: %v", err)
	}

	if len(fragments) == 0 {
		t.Error("Expected fragments to be returned")
	}
}

func TestMetricsCollector(t *testing.T) {
	collector := NewMetricsCollector()

	collector.RecordProviderDuration("memory", 32*time.Millisecond)
	collector.RecordFragment("memory", 2800)
	collector.RecordTruncation()
	collector.RecordAssemblyDuration(45 * time.Millisecond)

	metrics := collector.GetMetrics()

	if metrics.ProviderDurations["memory"] != 32.0 {
		t.Errorf("Expected memory duration 32ms, got %.2f", metrics.ProviderDurations["memory"])
	}

	if metrics.TokensUsed["memory"] != 2800 {
		t.Errorf("Expected 2800 tokens, got %d", metrics.TokensUsed["memory"])
	}

	if metrics.TruncationEvents != 1 {
		t.Errorf("Expected 1 truncation event, got %d", metrics.TruncationEvents)
	}

	if metrics.AssemblyDurationMs != 45.0 {
		t.Errorf("Expected 45ms assembly duration, got %.2f", metrics.AssemblyDurationMs)
	}
}
